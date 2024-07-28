package app

import (
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDatabase() error {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=%s",
		CFG.Postgres.Host,
		CFG.Postgres.User,
		CFG.Postgres.Password,
		CFG.Postgres.Name,
		CFG.Postgres.Port,
		CFG.Postgres.SslMode,
		CFG.Postgres.Timezone,
	)

	db, err := gorm.Open(postgres.Open(dsn))
	if err != nil {
		return err
	}

	DB = db

	return nil
}
