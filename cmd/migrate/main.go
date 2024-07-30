package main

import (
	"ton-lessons2/internal/app"
	"ton-lessons2/internal/storage"
)

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}

func run() error {
	if err := app.InitApp(); err != nil {
		return err
	}

	dbtx := app.DB.Begin()

	if err := dbtx.AutoMigrate(
		&storage.Block{},
		&storage.DedustSwap{},
	); err != nil {
		dbtx.Rollback()
		return err
	}

	if err := dbtx.Commit().Error; err != nil {
		return err
	}
	return nil
}