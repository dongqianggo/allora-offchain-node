package usecase

import (
	"allora_offchain_node/lib"
	"testing"

	alloraMath "github.com/allora-network/allora-chain/math"
	emissionstypes "github.com/allora-network/allora-chain/x/emissions/types"
	"github.com/stretchr/testify/assert"
)

func (suite *UseCaseSuite) SetupTest() {
	// Any setup needed for each test
}

func TestComputeWorkerBundle(t *testing.T) {
	workerOptions := map[string]string{
		"InferenceEndpoint": "http://source:8000/inference/{Token}",
		"Token":             "ETH",
	}

	tests := []struct {
		name             string
		workerConfig     lib.WorkerResponse
		mockSetup        func(*MockAlloraAdapter)
		expectedResponse emissionstypes.InferenceForecastBundle
		expectError      bool
		errorContains    string
		address          string
	}{
		{
			name: "Happy path - valid prediction",
			workerConfig: lib.WorkerResponse{
				WorkerConfig: lib.WorkerConfig{
					TopicId:                 emissionstypes.TopicId(1),
					InferenceEntrypointName: "api-worker-reputer",
					ForecastEntrypointName:  "api-worker-reputer",
					InferenceEntrypoint:     nil, // Will be set in the test
					ForecastEntrypoint:      nil, // Will be set in the test
					Parameters:              workerOptions,
					LoopSeconds:             10,
				},
				InfererValue: "9.5",
				ForecasterValues: []lib.NodeValue{
					{Value: "9.7", Worker: "worker1"},
				},
			},
			mockSetup: func(m *MockAlloraAdapter) {
			},
			expectedResponse: emissionstypes.InferenceForecastBundle{
				Inference: &emissionstypes.Inference{
					TopicId:     uint64(1),
					BlockHeight: 1,
					Inferer:     "worker1",
					Value:       alloraMath.MustNewDecFromString("9.5"),
				},
				Forecast: &emissionstypes.Forecast{
					TopicId:     uint64(1),
					BlockHeight: 1,
					Forecaster:  "worker1",
					ForecastElements: []*emissionstypes.ForecastElement{
						{
							Inferer: "worker1",
							Value:   alloraMath.MustNewDecFromString("9.7"),
						},
					},
				},
			},
			expectError: false,
			address:     "worker1",
		},
		// Add more test cases here
		{
			name: "Invalid inference value",
			workerConfig: lib.WorkerResponse{
				WorkerConfig: lib.WorkerConfig{
					TopicId:                 emissionstypes.TopicId(1),
					InferenceEntrypointName: "api-worker-reputer",
					ForecastEntrypointName:  "api-worker-reputer",
					InferenceEntrypoint:     nil,
					ForecastEntrypoint:      nil,
					Parameters:              workerOptions,
					LoopSeconds:             10,
				},
				InfererValue: "invalid",
				ForecasterValues: []lib.NodeValue{
					{Value: "9.7", Worker: "worker1"},
				},
			},
			mockSetup:     func(m *MockAlloraAdapter) {},
			expectError:   true,
			errorContains: "invalid decimal string",
			address:       "worker1",
		},
		{
			name: "Invalid forecast value",
			workerConfig: lib.WorkerResponse{
				WorkerConfig: lib.WorkerConfig{
					TopicId:                 emissionstypes.TopicId(1),
					InferenceEntrypointName: "api-worker-reputer",
					ForecastEntrypointName:  "api-worker-reputer",
					InferenceEntrypoint:     nil,
					ForecastEntrypoint:      nil,
					Parameters:              workerOptions,
					LoopSeconds:             10,
				},
				InfererValue: "9.5",
				ForecasterValues: []lib.NodeValue{
					{Value: "invalid", Worker: "worker1"},
				},
			},
			mockSetup:     func(m *MockAlloraAdapter) {},
			expectError:   true,
			errorContains: "invalid decimal string",
			address:       "worker1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockAdapter := NewMockAlloraAdapter()
			tt.mockSetup(mockAdapter)
			tt.workerConfig.InferenceEntrypoint = mockAdapter
			tt.workerConfig.ForecastEntrypoint = mockAdapter

			suite := &UseCaseSuite{}
			suite.Node.Wallet.Address = tt.address
			response, err := suite.BuildWorkerPayload(tt.workerConfig, 1)
			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorContains)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResponse.Inference.BlockHeight, response.Inference.BlockHeight)
				assert.Equal(t, tt.expectedResponse.Inference.Inferer, response.Inference.Inferer)
				assert.Equal(t, tt.expectedResponse.Inference.TopicId, response.Inference.TopicId)
				assert.Equal(t, tt.expectedResponse.Inference.Value, response.Inference.Value)
				assert.Equal(t, tt.expectedResponse.Forecast.BlockHeight, response.Forecast.BlockHeight)
				assert.Equal(t, tt.expectedResponse.Forecast.Forecaster, response.Forecast.Forecaster)
				assert.Equal(t, tt.expectedResponse.Forecast.TopicId, response.Forecast.TopicId)
				assert.Equal(t, tt.expectedResponse.Forecast.ForecastElements, response.Forecast.ForecastElements)
				assert.Equal(t, len(tt.expectedResponse.Forecast.ForecastElements), len(response.Forecast.ForecastElements))
				for _, expectedElement := range tt.expectedResponse.Forecast.ForecastElements {
					found := false
					for _, actualElement := range response.Forecast.ForecastElements {
						if expectedElement.Inferer == actualElement.Inferer && expectedElement.Value.Equal(actualElement.Value) {
							found = true
							break
						}
					}
					assert.True(t, found, "Expected forecast element not found: %v", expectedElement)
				}
			}

			mockAdapter.AssertExpectations(t)
		})
	}
}

// Add more test functions as needed
