package internal

import (
	"database/sql"
	"log"
	"os"

	"github.com/adriein/tibia-char/database"
	"github.com/adriein/tibia-char/pkg/constants"
	"github.com/adriein/tibia-char/pkg/helper"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
)

type App struct {
	Databse   *sql.DB
	Router    *gin.Engine
	Validator *validator.Validate
}

func NewApp() *App {
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

	return &App{
		Databse: database.New(),
	}
}

func (a *App) SetRouter(r *gin.Engine) {
	a.Router = r
}

func (a *App) SetValidator(v *validator.Validate) {
	a.Validator = v
}
