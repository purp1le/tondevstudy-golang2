package app

import (
	"encoding/json"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/xssnick/tonutils-go/liteclient"
)

type appConfig struct {
	Logger struct {
		LogLvl string // debug, info, error
	}

	MainnetConfig *liteclient.GlobalConfig

	Wallet struct {
		Seed []string
	}
}

var CFG *appConfig = &appConfig{}

func InitConfig() error {
	godotenv.Load(".env")

	CFG.Logger.LogLvl = os.Getenv("LOG_LVL")

	jsonConfig, err := os.Open("mainnet-config.json")
	if err != nil {
		return err
	}

	if err := json.NewDecoder(jsonConfig).Decode(&CFG.MainnetConfig); err != nil {
		return err
	}
	defer jsonConfig.Close()

	CFG.Wallet.Seed = strings.Split(os.Getenv("SEED"), " ")

	return nil
}
