package usecase

import (
	"allora_offchain_node/lib"
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"

	errorsmod "cosmossdk.io/errors"
	alloraMath "github.com/allora-network/allora-chain/math"
	emissionstypes "github.com/allora-network/allora-chain/x/emissions/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	"github.com/rs/zerolog/log"
)

// Get the reputer's values at the block from the chain
// Compute loss bundle with the reputer provided Loss function and ground truth
// sign and commit to chain
func (suite *UseCaseSuite) BuildCommitReputerPayload(reputer lib.ReputerConfig, nonce lib.BlockHeight) error {
	ctx := context.Background()

	valueBundle, err := suite.Node.GetReputerValuesAtBlock(reputer.TopicId, nonce)
	if err != nil {
		return errorsmod.Wrapf(err, "error getting reputer values, topic: %d, blockHeight: %d", reputer.TopicId, nonce)
	}
	valueBundle.ReputerRequestNonce = &emissionstypes.ReputerRequestNonce{
		ReputerNonce: &emissionstypes.Nonce{BlockHeight: nonce},
	}
	valueBundle.Reputer = suite.Node.Wallet.Address

	sourceTruth, err := reputer.GroundTruthEntrypoint.GroundTruth(reputer, nonce)
	if err != nil {
		return errorsmod.Wrapf(err, "error getting source truth from reputer, topicId: %d, blockHeight: %d", reputer.TopicId, nonce)
	}
	suite.Metrics.IncrementMetricsCounter(lib.TruthRequestCount, suite.Node.Chain.Address, reputer.TopicId)

	lossBundle, err := suite.ComputeLossBundle(sourceTruth, valueBundle, reputer)
	if err != nil {
		return errorsmod.Wrapf(err, "error computing loss bundle, topic: %d, blockHeight: %d", reputer.TopicId, nonce)
	}
	suite.Metrics.IncrementMetricsCounter(lib.ReputerDataBuildCount, suite.Node.Chain.Address, reputer.TopicId)

	signedValueBundle, err := suite.SignReputerValueBundle(&lossBundle)
	if err != nil {
		return errorsmod.Wrapf(err, "error signing reputer value bundle, topic: %d, blockHeight: %d", reputer.TopicId, nonce)
	}

	if err := signedValueBundle.Validate(); err != nil {
		return errorsmod.Wrapf(err, "error validating reputer value bundle, topic: %d, blockHeight: %d", reputer.TopicId, nonce)
	}

	req := &emissionstypes.InsertReputerPayloadRequest{
		Sender:             suite.Node.Wallet.Address,
		ReputerValueBundle: signedValueBundle,
	}
	reqJSON, err := json.Marshal(req)
	if err != nil {
		log.Error().Err(err).Uint64("topicId", reputer.TopicId).Msgf("Error marshaling MsgInserReputerPayload to print Msg as JSON")
	} else {
		log.Info().Uint64("topicId", reputer.TopicId).Msgf("Sending InsertReputerPayload to chain %s", string(reqJSON))
	}
	if suite.Node.Wallet.SubmitTx {
		_, err = suite.Node.SendDataWithRetry(ctx, req, "Send Reputer Data to chain")
		if err != nil {
			return errorsmod.Wrapf(err, "error sending Reputer Data to chain, topic: %d, blockHeight: %d", reputer.TopicId, nonce)
		}
		suite.Metrics.IncrementMetricsCounter(lib.ReputerChainSubmissionCount, suite.Node.Chain.Address, reputer.TopicId)
	} else {
		log.Info().Uint64("topicId", reputer.TopicId).Msg("SubmitTx=false; Skipping sending Reputer Data to chain")
	}

	return nil
}

func (suite *UseCaseSuite) ComputeLossBundle(sourceTruth string, vb *emissionstypes.ValueBundle, reputer lib.ReputerConfig) (emissionstypes.ValueBundle, error) {
	if vb == nil {
		return emissionstypes.ValueBundle{}, errors.New("nil ValueBundle")
	}
	// Check if vb is empty
	if IsEmpty(*vb) {
		return emissionstypes.ValueBundle{}, errors.New("empty ValueBundle")
	}
	if err := emissionstypes.ValidateDec(vb.CombinedValue); err != nil {
		return emissionstypes.ValueBundle{}, errors.New("ValueBundle - invalid CombinedValue")
	}
	if err := emissionstypes.ValidateDec(vb.NaiveValue); err != nil {
		return emissionstypes.ValueBundle{}, errors.New("ValueBundle - invalid NaiveValue")
	}

	lossMethodOptions := reputer.LossFunctionParameters.LossMethodOptions
	// Use the cached IsNeverNegative value
	is_never_negative := false
	if reputer.LossFunctionParameters.IsNeverNegative != nil {
		is_never_negative = *reputer.LossFunctionParameters.IsNeverNegative
	} else {
		var err error
		is_never_negative, err = reputer.LossFunctionEntrypoint.IsLossFunctionNeverNegative(reputer, lossMethodOptions)
		if err != nil {
			return emissionstypes.ValueBundle{}, errorsmod.Wrapf(err, "failed to determine if loss function is never negative")
		}
		// cache the result
		reputer.LossFunctionParameters.IsNeverNegative = &is_never_negative
	}

	losses := emissionstypes.ValueBundle{
		TopicId:             vb.TopicId,
		ReputerRequestNonce: vb.ReputerRequestNonce,
		Reputer:             vb.Reputer,
		ExtraData:           vb.ExtraData,
	}

	computeLoss := func(value alloraMath.Dec, description string) (alloraMath.Dec, error) {
		lossStr, err := reputer.LossFunctionEntrypoint.LossFunction(reputer, sourceTruth, value.String(), lossMethodOptions)
		if err != nil {
			return alloraMath.Dec{}, errorsmod.Wrapf(err, "error computing loss for %s", description)
		}

		loss, err := alloraMath.NewDecFromString(lossStr)
		if err != nil {
			return alloraMath.Dec{}, errorsmod.Wrapf(err, "error parsing loss value for %s", description)
		}

		if is_never_negative {
			loss, err = alloraMath.Log10(loss)
			if err != nil {
				return alloraMath.Dec{}, errorsmod.Wrapf(err, "error Log10 for %s", description)
			}
		}

		if err := emissionstypes.ValidateDec(loss); err != nil {
			return alloraMath.Dec{}, errorsmod.Wrapf(err, "invalid loss value for %s", description)
		}

		return loss, nil
	}

	// Combined Value
	if combinedLoss, err := computeLoss(vb.CombinedValue, "combined value"); err != nil {
		return emissionstypes.ValueBundle{}, errorsmod.Wrapf(err, "error computing loss for combined value")
	} else {
		losses.CombinedValue = combinedLoss
	}

	// Naive Value
	if naiveLoss, err := computeLoss(vb.NaiveValue, "naive value"); err != nil {
		return emissionstypes.ValueBundle{}, errorsmod.Wrapf(err, "error computing loss for naive value")
	} else {
		losses.NaiveValue = naiveLoss
	}

	// Inferer Values
	losses.InfererValues = make([]*emissionstypes.WorkerAttributedValue, len(vb.InfererValues))
	for i, val := range vb.InfererValues {
		if loss, err := computeLoss(val.Value, fmt.Sprintf("inferer value %d", i)); err != nil {
			return emissionstypes.ValueBundle{}, errorsmod.Wrapf(err, "error computing loss for inferer value")
		} else {
			losses.InfererValues[i] = &emissionstypes.WorkerAttributedValue{Worker: val.Worker, Value: loss}
		}
	}

	// Forecaster Values
	losses.ForecasterValues = make([]*emissionstypes.WorkerAttributedValue, len(vb.ForecasterValues))
	for i, val := range vb.ForecasterValues {
		if loss, err := computeLoss(val.Value, fmt.Sprintf("forecaster value %d", i)); err != nil {
			return emissionstypes.ValueBundle{}, errorsmod.Wrapf(err, "error computing loss for forecaster value")
		} else {
			losses.ForecasterValues[i] = &emissionstypes.WorkerAttributedValue{Worker: val.Worker, Value: loss}
		}
	}

	// One Out Inferer Values
	losses.OneOutInfererValues = make([]*emissionstypes.WithheldWorkerAttributedValue, len(vb.OneOutInfererValues))
	for i, val := range vb.OneOutInfererValues {
		if loss, err := computeLoss(val.Value, fmt.Sprintf("one out inferer value %d", i)); err != nil {
			return emissionstypes.ValueBundle{}, errorsmod.Wrapf(err, "error computing loss for one-out inferer value")
		} else {
			losses.OneOutInfererValues[i] = &emissionstypes.WithheldWorkerAttributedValue{Worker: val.Worker, Value: loss}
		}
	}

	// One Out Forecaster Values
	losses.OneOutForecasterValues = make([]*emissionstypes.WithheldWorkerAttributedValue, len(vb.OneOutForecasterValues))
	for i, val := range vb.OneOutForecasterValues {
		if loss, err := computeLoss(val.Value, fmt.Sprintf("one out forecaster value %d", i)); err != nil {
			return emissionstypes.ValueBundle{}, errorsmod.Wrapf(err, "error computing loss for one-out forecaster value")
		} else {
			losses.OneOutForecasterValues[i] = &emissionstypes.WithheldWorkerAttributedValue{Worker: val.Worker, Value: loss}
		}
	}

	// One In Forecaster Values
	losses.OneInForecasterValues = make([]*emissionstypes.WorkerAttributedValue, len(vb.OneInForecasterValues))
	for i, val := range vb.OneInForecasterValues {
		if loss, err := computeLoss(val.Value, fmt.Sprintf("one in forecaster value %d", i)); err != nil {
			return emissionstypes.ValueBundle{}, errorsmod.Wrapf(err, "error computing loss for one-in forecaster value")
		} else {
			losses.OneInForecasterValues[i] = &emissionstypes.WorkerAttributedValue{Worker: val.Worker, Value: loss}
		}
	}
	return losses, nil
}

func (suite *UseCaseSuite) SignReputerValueBundle(valueBundle *emissionstypes.ValueBundle) (*emissionstypes.ReputerValueBundle, error) {
	// Marshall and sign the bundle
	protoBytesIn := make([]byte, 0) // Create a byte slice with initial length 0 and capacity greater than 0
	protoBytesIn, err := valueBundle.XXX_Marshal(protoBytesIn, true)
	if err != nil {
		return &emissionstypes.ReputerValueBundle{}, errorsmod.Wrapf(err, "error marshalling valueBundle")
	}
	sig, pk, err := suite.Node.Chain.Client.Context().Keyring.Sign(suite.Node.Chain.Account.Name, protoBytesIn, signing.SignMode_SIGN_MODE_DIRECT)
	pkStr := hex.EncodeToString(pk.Bytes())
	if err != nil {
		return &emissionstypes.ReputerValueBundle{}, errorsmod.Wrapf(err, "error signing valueBundle")
	}

	reputerValueBundle := &emissionstypes.ReputerValueBundle{
		ValueBundle: valueBundle,
		Signature:   sig,
		Pubkey:      pkStr,
	}

	return reputerValueBundle, nil
}
