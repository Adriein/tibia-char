package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/adriein/tibia-char/pkg/constants"
	"github.com/rotisserie/eris"
)

func New() *sql.DB {
	databaseDsn := fmt.Sprintf(
		"postgresql://%s:%s@%s:5432/%s?sslmode=disable",
		os.Getenv(constants.DatabaseUser),
		os.Getenv(constants.DatabasePassword),
		os.Getenv(constants.DatabaseHost),
		os.Getenv(constants.DatabaseName),
	)

	database, dbConnErr := sql.Open("postgres", databaseDsn)

	if dbConnErr != nil {
		log.Fatal(dbConnErr.Error())
	}

	return database
}

func CloseRowsSafely(rows *sql.Rows, err *error) {
	if rowsErr := rows.Close(); rowsErr != nil && *err == nil {
		*err = eris.Wrap(rowsErr, "Failed to close rows")
	}
	if streamErr := rows.Err(); streamErr != nil && *err == nil {
		*err = eris.Wrap(streamErr, "Database stream cut off")
	}
}
