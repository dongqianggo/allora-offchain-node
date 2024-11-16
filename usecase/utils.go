package usecase

import (
	"fmt"
	"net/http"
	"time"
	"allora_offchain_node/lib"

	emissionstypes "github.com/allora-network/allora-chain/x/emissions/types"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

func (suite *UseCaseSuite) Wait(seconds int64) {
	time.Sleep(time.Duration(seconds) * time.Second)
}

func IsEmpty(vb emissionstypes.ValueBundle) bool {
	return vb.TopicId == 0 &&
		vb.ReputerRequestNonce == nil &&
		vb.Reputer == "" &&
		vb.CombinedValue.IsZero() &&
		vb.NaiveValue.IsZero() &&
		len(vb.InfererValues) == 0 &&
		len(vb.ForecasterValues) == 0 &&
		len(vb.OneOutInfererValues) == 0 &&
		len(vb.OneOutForecasterValues) == 0 &&
		len(vb.OneInForecasterValues) == 0 &&
		len(vb.OneOutInfererForecasterValues) == 0 &&
		len(vb.ExtraData) == 0
}

func GetLocalIP() {
	resp, err := http.Get("https://api.ipify.org")
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
	defer resp.Body.Close()

	var ip string
	_, err = fmt.Fscan(resp.Body, &ip)
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
	log.Info().Str("local ip--------------->:", ip).Msg("get local ip")

	lib.LOCALIP = ip
}

func ReadFile() {
	var configViperConfig = viper.New()
	configViperConfig.SetConfigName("server")
	configViperConfig.SetConfigType("yaml")
	configViperConfig.AddConfigPath("./")
	//读取配置文件内容
	if err := configViperConfig.ReadInConfig(); err != nil {
		panic(err)
	}
	var c lib.ConfigStruct
	if err := configViperConfig.Unmarshal(&c); err != nil {
		panic(err)
	}

	lib.CONFIG_STRUCT = c
	log.Info().Interface("--------config", lib.CONFIG_STRUCT).Msg("read config file")
}
