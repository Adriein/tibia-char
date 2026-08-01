package database

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"log/slog"
	"os"
	"time"

	"github.com/adriein/tibia-char/pkg/constants"
	"github.com/lib/pq"
	"github.com/rotisserie/eris"
)

type slowQueryConnector struct {
	driver *slowQueryDriver
	dsn    string
}

func (c *slowQueryConnector) Connect(ctx context.Context) (driver.Conn, error) {
	return c.driver.Open(c.dsn)
}

func (c *slowQueryConnector) Driver() driver.Driver {
	return c.driver
}

type slowQueryDriver struct {
	driver.Driver
	logger    *slog.Logger
	threshold time.Duration
}

func (d *slowQueryDriver) Open(name string) (driver.Conn, error) {
	conn, err := d.Driver.Open(name)

	if err != nil {
		return nil, err
	}

	return &customConnection{Conn: conn, logger: d.logger, threshold: d.threshold}, nil
}

type customConnection struct {
	driver.Conn
	logger    *slog.Logger
	threshold time.Duration
}

func (c *customConnection) QueryContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	start := time.Now()

	var rows driver.Rows
	var err error

	if queryer, ok := c.Conn.(driver.QueryerContext); ok {
		rows, err = queryer.QueryContext(ctx, query, args)
	} else {
		dargs := make([]driver.Value, len(args))
		for i, nv := range args {
			dargs[i] = nv.Value
		}
		rows, err = c.Conn.(driver.Queryer).Query(query, dargs)
	}

	duration := time.Since(start)

	if duration >= c.threshold {
		c.logger.Warn("Slow query detected",
			slog.Duration("duration", duration),
			slog.String("query", query),
			slog.Any("error", err),
		)
	}

	return rows, err
}

func (c *customConnection) ExecContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	start := time.Now()

	var res driver.Result
	var err error

	if execer, ok := c.Conn.(driver.ExecerContext); ok {
		res, err = execer.ExecContext(ctx, query, args)
	} else {
		dargs := make([]driver.Value, len(args))
		for i, nv := range args {
			dargs[i] = nv.Value
		}
		res, err = c.Conn.(driver.Execer).Exec(query, dargs)
	}

	duration := time.Since(start)

	if duration >= c.threshold {
		c.logger.Warn("Slow query detected",
			slog.Duration("duration", duration),
			slog.String("query", query),
			slog.Any("error", err),
		)
	}

	return res, err
}

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
