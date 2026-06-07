package auction

import (
	"database/sql"
	"errors"

	"github.com/rotisserie/eris"
)

//TODO: aggregation table is broken for sure i have more records than aggregations something is not taked into account

type AggAuctionStatsRepository interface {
	Save(stats *AggAuctionStats) error
	GetByKey(key string) (*AggAuctionStats, error)
}

type PgAggAuctionStatsRepsitory struct {
	connection *sql.DB
}

func NewPgAggAuctionRepository(c *sql.DB) *PgAggAuctionStatsRepsitory {
	return &PgAggAuctionStatsRepsitory{connection: c}
}

func (r *PgAggAuctionStatsRepsitory) Save(stats *AggAuctionStats) error {
	query := `
		INSERT INTO tc_aggregated_auction_stats (
			taas_subset_key,
			taas_median_price,
			taas_mean_price,
			taas_std_deviation,
			taas_min_price,
			taas_max_price,
			taas_mode_price,
			taas_sample_size,
			taas_date_upd
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, TIMEZONE('UTC', NOW()))
		ON CONFLICT (taas_subset_key) DO UPDATE SET
			taas_median_price = EXCLUDED.taas_median_price,
			taas_mean_price = EXCLUDED.taas_mean_price,
			taas_std_deviation = EXCLUDED.taas_std_deviation,
			taas_min_price = EXCLUDED.taas_min_price,
			taas_max_price = EXCLUDED.taas_max_price,
			taas_mode_price = EXCLUDED.taas_mode_price,
			taas_sample_size = EXCLUDED.taas_sample_size,
			taas_date_upd = EXCLUDED.taas_date_upd
		;
	`

	_, err := r.connection.Exec(
		query,
		stats.SubsetKey,
		stats.Median,
		stats.Mean,
		stats.StdDeviation,
		stats.MinPrice,
		stats.MaxPrice,
		stats.Mode,
		stats.SampleSize,
	)

	if err != nil {
		return eris.Wrap(err, "Error in upsert on tc_aggregated_auction_stats")
	}

	return nil
}

func (r *PgAggAuctionStatsRepsitory) GetByKey(key string) (*AggAuctionStats, error) {
	query := `
		SELECT
			taas_subset_key,
			taas_median_price,
			taas_mean_price,
			taas_std_deviation,
			taas_min_price,
			taas_max_price,
			taas_mode_price,
			taas_sample_size
		FROM
			tc_aggregated_auction_stats
		WHERE
			taas_subset_key = $1;
	`

	var stats AggAuctionStats

	err := r.connection.QueryRow(query, key).Scan(
		&stats.SubsetKey,
		&stats.Median,
		&stats.Mean,
		&stats.StdDeviation,
		&stats.MinPrice,
		&stats.MaxPrice,
		&stats.Mode,
		&stats.SampleSize,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, eris.Wrapf(ErrAggAuctionStatsNotFound, "No aggregated auction stats found for key: %s", key)
		}

		return nil, eris.Wrap(err, "Error querying aggregated auction stats by key")
	}

	return &stats, nil
}
