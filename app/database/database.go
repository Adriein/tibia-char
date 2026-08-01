package database

import (
	"context"
	"database/sql"
	"log/slog"
	"os"
	"time"

	"github.com/adriein/tibia-char/pkg/constants"
	"github.com/lib/pq"
	"github.com/rotisserie/eris"
)

func New(logger *slog.Logger) *sql.DB {
	databaseDsn := os.Getenv(constants.DatabaseUrl)

	wrappedDriver := &slowQueryDriver{
		Driver:    &pq.Driver{},
		logger:    logger,
		threshold: 10 * time.Second,
	}

	connector := &slowQueryConnector{
		driver: wrappedDriver,
		dsn:    databaseDsn,
	}

	database := sql.OpenDB(connector)

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
