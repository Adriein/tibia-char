package main

import (
	"log"
	"os"

	"github.com/adriein/tibia-char/internal/scrap"
	"github.com/adriein/tibia-char/pkg/constants"
	"github.com/adriein/tibia-char/pkg/helper"
	"github.com/joho/godotenv"
)

func main() {
	if os.Getenv(constants.Env) != constants.Production {
		dotenvErr := godotenv.Load()

		if dotenvErr != nil {
			log.Fatal("Error loading .env file")
		}
	}

	checker := helper.NewEnvVarChecker(
		constants.DatabaseUser,
		constants.DatabasePassword,
		constants.DatabaseName,
		constants.ServerPort,
		constants.Env,
	)

	if envCheckerErr := checker.Check(); envCheckerErr != nil {
		log.Fatal(envCheckerErr.Error())
	}

	cron := scrap.NewService()

	err := cron.ScrapBazaar()

	if err != nil {
		log.Fatal(err.Error())
	}
}
