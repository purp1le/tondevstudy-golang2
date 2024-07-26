package main

import (
	"ton-lessons2/internal/app"
	"ton-lessons2/internal/scanner"
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

	scanner, err := scanner.NewScanner()
	if err != nil {
		return err
	}

	scanner.Listen()

	return nil
}