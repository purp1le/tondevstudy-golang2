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

	Postgres struct {
		Host     string
		Port     string
		User     string
		Password string
		Name     string
		SslMode  string
		Timezone string
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

	CFG.Postgres.Host = os.Getenv("POSTGRES_HOST")
	CFG.Postgres.Port = os.Getenv("POSTGRES_PORT")
	CFG.Postgres.User = os.Getenv("POSTGRES_USER")
	CFG.Postgres.Password = os.Getenv("POSTGRES_PASSWORD")
	CFG.Postgres.Name = os.Getenv("POSTGRES_DB")
	CFG.Postgres.SslMode = os.Getenv("POSTGRES_SSLMODE")
	CFG.Postgres.Timezone = os.Getenv("POSTGRES_TIMEZONE")

	return nil
}
