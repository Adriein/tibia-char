package database

import (
	"context"
	"database/sql"
	"log/slog"
	"os"
	"time"

	"github.com/adriein/tibia-char/pkg/constants"
	"github.com/rotisserie/eris"
)

func New(logger *slog.Logger) *sql.DB {
	databaseDsn := os.Getenv(constants.DatabaseUrl)

	database, err := sql.Open("postgres", databaseDsn)

	if err != nil {
		enhancedErr := eris.Wrap(err, "Failed db open conn on db init")

		logger.Error(eris.ToString(enhancedErr, true))
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	defer cancel()

	if err := database.PingContext(ctx); err != nil {
		enhancedErr := eris.Wrap(err, "Failed db ping on db init")

		logger.Error(eris.ToString(enhancedErr, true))
		os.Exit(1)
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
