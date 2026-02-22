package auction

import (
	"context"
	"database/sql"
	"time"

	"github.com/adriein/tibia-char/pkg/enums"
	"github.com/lib/pq"
	"github.com/rotisserie/eris"
)

type VocationRepository interface {
	Get(vocation string) (*Vocation, error)
}

type PgVocationRepsitory struct {
	connection *sql.DB
}

func NewPgVocationRepository(c *sql.DB) *PgVocationRepsitory {
	return &PgVocationRepsitory{connection: c}
}

func (vr *PgVocationRepsitory) Get(vocation string) (*Vocation, error) {
	query := `SELECT * FROM tc_vocation WHERE tv_name = $1;`

	var dto Vocation

	err := vr.connection.QueryRow(query, vocation).Scan(&dto.Id, &dto.Name)

	if err != nil {
		return nil, eris.New(err.Error())
	}

	return &dto, nil
}

type GenderRepository interface {
	Get(gender string) (*Gender, error)
}

type PgGenderRepository struct {
	connection *sql.DB
}

func NewPgGenderRepository(c *sql.DB) *PgGenderRepository {
	return &PgGenderRepository{connection: c}
}

func (gr *PgGenderRepository) Get(gender string) (*Gender, error) {
	query := `SELECT * FROM tc_gender WHERE tg_name = $1`

	var dto Gender

	err := gr.connection.QueryRow(query, gender).Scan(&dto.Id, &dto.Name)

	if err != nil {
		return nil, eris.New(err.Error())
	}

	return &dto, nil
}

type WorldRepository interface {
	GetOrCreate(world *World) (*World, error)
}

type PgWorldRepository struct {
	connection *sql.DB
}

func NewPgWorldRepository(c *sql.DB) *PgWorldRepository {
	return &PgWorldRepository{connection: c}
}

func (wr *PgWorldRepository) GetOrCreate(world *World) (*World, error) {
	query := `
		WITH existing AS (
    		SELECT * FROM tc_world WHERE tw_name = $1
		),
		inserted AS (
    		INSERT INTO tc_world (tw_name, tw_location, tw_battle_eye, tw_pvp)
    		SELECT $1, $2, $3, $4
    		WHERE NOT EXISTS (SELECT 1 FROM existing)
    		ON CONFLICT (tw_name) DO NOTHING
    		RETURNING *
		)
		SELECT * FROM inserted
		UNION ALL
		SELECT * FROM existing
		LIMIT 1;
	`

	var dto World

	var battleEye string

	err := wr.connection.QueryRow(
		query,
		world.Name,
		world.Location,
		world.BattleEye.String(),
		world.Pvp,
	).Scan(
		&dto.Id,
		&dto.Name,
		&dto.Location,
		&battleEye,
		&dto.Pvp,
	)

	if err != nil {
		return nil, eris.Wrap(err, world.Name)
	}

	dto.BattleEye, err = enums.GetBattleEyeFromString(battleEye)

	if err != nil {
		return nil, err
	}

	return &dto, nil
}

type AuctionRepository interface {
	Save(auction *Auction) error
	GetActiveAuctions(ctx context.Context) ([]*Auction, error)
	GetActiveAuctionsFinishingIn(ctx context.Context, duration time.Duration) ([]*Auction, error)
	GetAuctionsWithFilter(ctx context.Context, filter *AuctionFilter) ([]*Auction, error)
	GetAuctionsPendingToConsolidate(ctx context.Context) ([]*Auction, error)
	GetAllAuctionPrices(ctx context.Context) ([]*Auction, error)
}

type PgAuctionRepository struct {
	connection *sql.DB
}

func NewPgAuctionRepository(connection *sql.DB) *PgAuctionRepository {
	return &PgAuctionRepository{
		connection: connection,
	}
}

func (r *PgAuctionRepository) Save(auction *Auction) error {
	tx, err := r.connection.Begin()

	if err != nil {
		return eris.Wrap(err, "Error creating transaction")
	}

	auctionQuery := `
		INSERT INTO tc_auction (
			ta_auction_id,
			ta_tibia_auction_link,
			ta_img,
			ta_char_name,
			ta_char_level,
			ta_char_vocation,
			ta_char_gender,
			ta_char_world,
			ta_world_transfer,
			ta_boss_points,
			ta_charm_expansion,
			ta_charm_points,
			ta_task_expansion,
			ta_current_bid,
			ta_current_bid_fiat,
			ta_current_bid_currency,
			ta_auction_stage,
			ta_auction_start,
			ta_auction_end,
			ta_date_add,
			ta_date_upd
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21)
		RETURNING ta_id;
	`

	var generatedId int64

	err = tx.QueryRow(
		auctionQuery,
		auction.AuctionID,
		auction.TibiaAuctionLink,
		auction.Img,
		auction.CharName,
		auction.CharLevel,
		auction.CharVocation.Id,
		auction.CharGender.Id,
		auction.CharWorld.Id,
		auction.WorldTransfer,
		auction.BossPoints,
		auction.CharmExpansion,
		auction.CharmPoints,
		false,
		auction.Bid,
		auction.BidFiat,
		auction.BidCurrency,
		auction.Stage,
		auction.AuctionStart,
		auction.AuctionEnd,
		auction.DateAdd,
		auction.DateUpd,
	).Scan(&generatedId)

	if err != nil {
		tx.Rollback()

		return eris.Wrap(err, "Error on insert on tc_auction")
	}

	recordingQuery := `
		INSERT INTO tc_auction_recording (
			tar_auction_id,
			tar_recordable_id,
			tar_status,
			tar_date_add,
			tar_date_upd
		)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (tar_auction_id) DO UPDATE SET
			tar_recordable_id = EXCLUDED.tar_recordable_id,
			tar_status = EXCLUDED.tar_status,
			tar_date_upd = EXCLUDED.tar_date_upd;
	`

	_, err = tx.Exec(
		recordingQuery,
		auction.AuctionID,
		generatedId,
		auction.Status.String(),
		auction.DateAdd,
		auction.DateUpd,
	)

	if err != nil {
		tx.Rollback()

		return eris.Wrap(err, "Error in upsert on tc_auction_recording")
	}

	skillsQuery := `
		INSERT INTO tc_skills (
			ts_auction_id,
			ts_axe,
			ts_club,
			ts_distance,
			ts_fishing,
			ts_fist,
			ts_magic_level,
			ts_shielding,
			ts_sword
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (ts_auction_id) DO UPDATE SET
			ts_axe = EXCLUDED.ts_axe,
			ts_club = EXCLUDED.ts_club,
			ts_distance = EXCLUDED.ts_distance,
			ts_fishing = EXCLUDED.ts_fishing,
			ts_fist = EXCLUDED.ts_fist,
			ts_magic_level = EXCLUDED.ts_magic_level,
			ts_shielding = EXCLUDED.ts_shielding,
			ts_sword = EXCLUDED.ts_sword
		;
	`

	_, err = tx.Exec(
		skillsQuery,
		auction.Skills.AuctionID,
		auction.Skills.Axe,
		auction.Skills.Club,
		auction.Skills.Distance,
		auction.Skills.Fishing,
		auction.Skills.Fist,
		auction.Skills.MagicLevel,
		auction.Skills.Shielding,
		auction.Skills.Sword,
	)

	if err != nil {
		tx.Rollback()

		return eris.Wrap(err, "Error in upsert on tc_skills")
	}

	featuredItemsQuery := `
		INSERT INTO tc_featured_items (
			tfi_auction_id,
			tfi_item_id
		)
		SELECT $1, unnest($2::int[])
		ON CONFLICT DO NOTHING;
	`

	featuredItemIDS := make([]int, len(auction.FeaturedItems))

	for i, item := range auction.FeaturedItems {
		featuredItemIDS[i] = item.ItemID
	}

	_, err = tx.Exec(
		featuredItemsQuery,
		auction.AuctionID,
		pq.Array(featuredItemIDS),
	)

	if err != nil {
		tx.Rollback()

		return eris.Wrap(err, "Error in insert on tc_featured_items")
	}

	charmQuery := `
		INSERT INTO tc_auction_charms (
			tac_auction_id,
			tac_charm_id,
			tac_grade
		)
		SELECT $1, unnest($2::int[]), unnest($3::smallint[])
		ON CONFLICT DO NOTHING;
	`

	charmsIDS := make([]int, len(auction.Charms))
	grades := make([]int, len(auction.Charms))

	for i, charm := range auction.Charms {
		charmsIDS[i] = charm.ID
		grades[i] = charm.Grade
	}

	_, err = tx.Exec(
		charmQuery,
		auction.AuctionID,
		pq.Array(charmsIDS),
		pq.Array(grades),
	)

	if err != nil {
		tx.Rollback()

		return eris.Wrap(err, "Error in insert on tc_charm")
	}

	imbuementsQuery := `
		INSERT INTO tc_auction_imbuements (
			tai_auction_id,
			tai_imbuement_id
		)
		SELECT $1, unnest($2::int[])
		ON CONFLICT DO NOTHING;
	`

	imbuementIDS := make([]int, len(auction.Imbuements))

	for i, imbuement := range auction.Imbuements {
		imbuementIDS[i] = imbuement.ID
	}

	_, err = tx.Exec(
		imbuementsQuery,
		auction.AuctionID,
		pq.Array(imbuementIDS),
	)

	if err != nil {
		tx.Rollback()

		return eris.Wrap(err, "Error in insert on tc_auction_imbuements")
	}

	questsQuery := `
		INSERT INTO tc_auction_quests (
			taq_auction_id,
			taq_quest_id
		)
		SELECT $1, unnest($2::int[])
		ON CONFLICT DO NOTHING;
	`

	questIDS := make([]int, len(auction.Quests))

	for i, quest := range auction.Quests {
		questIDS[i] = quest.ID
	}

	_, err = tx.Exec(
		questsQuery,
		auction.AuctionID,
		pq.Array(questIDS),
	)

	if err != nil {
		tx.Rollback()

		return eris.Wrap(err, "Error in insert on tc_auction_quests")
	}

	return tx.Commit()
}

func (r *PgAuctionRepository) GetActiveAuctions(ctx context.Context) ([]*Auction, error) {
	query := `
		SELECT
			a.ta_id,
			a.ta_auction_id,
			a.ta_tibia_auction_link,
			a.ta_img,
			a.ta_char_name,
			a.ta_char_level,
			v.*,
			g.*,
			w.*,
			ts.*,
			a.ta_world_transfer,
			a.ta_boss_points,
			a.ta_charm_expansion,
			a.ta_charm_points,
			a.ta_task_expansion,
			a.ta_current_bid,
			a.ta_current_bid_fiat,
			a.ta_current_bid_currency,
			a.ta_auction_stage,
			a.ta_auction_start,
			a.ta_auction_end,
			tar.tar_status,
			tar.tar_date_add,
			tar.tar_date_upd
		FROM
			tc_auction a
		INNER JOIN
			tc_vocation v ON a.ta_char_vocation = v.tv_id
		INNER JOIN
			tc_gender g ON a.ta_char_gender = g.tg_id
		INNER JOIN
			tc_world w ON a.ta_char_world = w.tw_id
		INNER JOIN
			tc_auction_recording tar ON a.ta_id = tar.tar_recordable_id
		INNER JOIN
			tc_skills ts ON a.ta_auction_id = ts.ts_auction_id
		WHERE
			tar.tar_status = 'active'
		ORDER BY
			a.ta_auction_end ASC;
	`

	ctxTimeout, cancel := context.WithTimeout(ctx, time.Second*10)

	defer cancel()

	rows, err := r.connection.QueryContext(ctxTimeout, query)

	if err != nil {
		return nil, eris.Wrap(err, "Failed to query active auctions")
	}

	defer rows.Close()

	auctionMap := make(map[int]*Auction)

	var orderedIDs []int

	for rows.Next() {
		var (
			auction         Auction
			vocation        Vocation
			gender          Gender
			world           World
			skills          Skills
			battleEyeString string
			statusString    string
		)

		err := rows.Scan(
			&auction.ID,
			&auction.AuctionID,
			&auction.TibiaAuctionLink,
			&auction.Img,
			&auction.CharName,
			&auction.CharLevel,
			&vocation.Id,
			&vocation.Name,
			&gender.Id,
			&gender.Name,
			&world.Id,
			&world.Name,
			&world.Location,
			&battleEyeString,
			&world.Pvp,
			&skills.AuctionID,
			&skills.Axe,
			&skills.Club,
			&skills.Distance,
			&skills.Fishing,
			&skills.Fist,
			&skills.MagicLevel,
			&skills.Shielding,
			&skills.Sword,
			&auction.WorldTransfer,
			&auction.BossPoints,
			&auction.CharmExpansion,
			&auction.CharmPoints,
			&auction.TaskExpansion,
			&auction.Bid,
			&auction.BidFiat,
			&auction.BidCurrency,
			&auction.Stage,
			&auction.AuctionStart,
			&auction.AuctionEnd,
			&statusString,
			&auction.DateAdd,
			&auction.DateUpd,
		)

		if err != nil {
			return nil, eris.Wrap(err, "Failed to scan auction")
		}

		status, err := enums.GetAuctionRecordableStatusFromString(statusString)

		if err != nil {
			return nil, eris.Wrap(err, "Failed to parse auction status")
		}

		battleEyeEnum, err := enums.GetBattleEyeFromString(battleEyeString)

		if err != nil {
			return nil, eris.Wrap(err, "Failed parsing Battle Eye")
		}

		auction.Status = status
		auction.CharVocation = &vocation
		auction.CharGender = &gender
		world.BattleEye = battleEyeEnum
		auction.CharWorld = &world
		auction.Skills = &skills
		auction.Charms = make([]*Charm, 0)
		auction.FeaturedItems = make([]*FeaturedItem, 0)
		auction.Imbuements = make([]*Imbuement, 0)

		auctionMap[auction.AuctionID] = &auction

		orderedIDs = append(orderedIDs, auction.AuctionID)
	}

	if err = rows.Err(); err != nil {
		return nil, eris.Wrap(err, "Failed iterating auction rows")
	}

	if len(orderedIDs) == 0 {
		return []*Auction{}, nil
	}

	featuredItemsQuery := `
		SELECT
			tfi.tfi_id,
			tfi.tfi_auction_id,
			tfi.tfi_item_id
		FROM
			tc_featured_items tfi
		WHERE
			tfi.tfi_auction_id = ANY($1);
	`

	featuredItemsRows, err := r.connection.QueryContext(ctx, featuredItemsQuery, pq.Array(orderedIDs))

	if err != nil {
		return nil, eris.Wrap(err, "Failed to query featured items")
	}

	defer featuredItemsRows.Close()

	for featuredItemsRows.Next() {
		var item FeaturedItem

		if err := featuredItemsRows.Scan(&item.ID, &item.AuctionID, &item.ItemID); err != nil {
			return nil, eris.Wrap(err, "Failed to scan featured item")
		}

		if auction, ok := auctionMap[item.AuctionID]; ok {
			auction.FeaturedItems = append(auction.FeaturedItems, &item)
		}
	}

	if err = featuredItemsRows.Err(); err != nil {
		return nil, eris.Wrap(err, "Failed iterating featured item rows")
	}

	imbuementsQuery := `
		SELECT
			tai.tai_auction_id,
			ti.ti_id,
			ti.ti_name
		FROM
			tc_auction_imbuements tai
		INNER JOIN
			tc_imbuements ti ON tai.tai_imbuement_id = ti.ti_id
		WHERE
			tai.tai_auction_id = ANY($1);
	`
	imbuementRows, err := r.connection.QueryContext(ctx, imbuementsQuery, pq.Array(orderedIDs))

	if err != nil {
		return nil, eris.Wrap(err, "Failed to query imbuements")
	}

	defer imbuementRows.Close()

	for imbuementRows.Next() {
		var imbuement Imbuement
		var auctionID int

		if err := imbuementRows.Scan(&auctionID, &imbuement.ID, &imbuement.Name); err != nil {
			return nil, eris.Wrap(err, "Failed to scan imbuement")
		}

		if auction, ok := auctionMap[auctionID]; ok {
			auction.Imbuements = append(auction.Imbuements, &imbuement)
		}
	}

	if err = imbuementRows.Err(); err != nil {
		return nil, eris.Wrap(err, "Failed iterating imbuement rows")
	}

	charmsQuery := `
		SELECT
			tac.tac_auction_id,
			tac.tac_charm_id,
			tac.tac_grade,
			tc.tc_name,
			tc.tc_type
		FROM
			tc_auction_charms tac
		INNER JOIN
			tc_charms tc ON tac.tac_charm_id = tc.tc_id
		WHERE
			tac.tac_auction_id = ANY($1);
	`
	charmsRows, err := r.connection.QueryContext(ctx, charmsQuery, pq.Array(orderedIDs))

	if err != nil {
		return nil, eris.Wrap(err, "Failed to query charms")
	}

	defer charmsRows.Close()

	for charmsRows.Next() {
		var charm Charm
		var auctionID int

		if err := charmsRows.Scan(&auctionID, &charm.ID, &charm.Grade, &charm.Name, &charm.Type); err != nil {
			return nil, eris.Wrap(err, "Failed to scan charm")
		}

		if auction, ok := auctionMap[auctionID]; ok {
			auction.Charms = append(auction.Charms, &charm)
		}
	}

	if err = charmsRows.Err(); err != nil {
		return nil, eris.Wrap(err, "Failed iterating charms rows")
	}

	questsQuery := `
		SELECT
			taq.taq_auction_id,
			taq.taq_quest_id,
			tq.tq_name
		FROM
			tc_auction_quests taq
		INNER JOIN
			tc_quests tq ON taq.taq_quest_id = tq.tq_id
		WHERE
			taq.taq_auction_id = ANY($1);
	`
	questsRows, err := r.connection.QueryContext(ctx, questsQuery, pq.Array(orderedIDs))

	if err != nil {
		return nil, eris.Wrap(err, "Failed to query quests")
	}

	defer questsRows.Close()

	for questsRows.Next() {
		var quest Quest
		var auctionID int

		if err := questsRows.Scan(&auctionID, &quest.ID, &quest.Name); err != nil {
			return nil, eris.Wrap(err, "Failed to scan quest")
		}

		if auction, ok := auctionMap[auctionID]; ok {
			auction.Quests = append(auction.Quests, &quest)
		}
	}

	if err = questsRows.Err(); err != nil {
		return nil, eris.Wrap(err, "Failed iterating quests rows")
	}

	bidRegistryQuery := `
		SELECT
			ta.ta_auction_id,
			ta.ta_current_bid,
			ta.ta_date_add
		FROM
			tc_auction ta
		WHERE
			ta.ta_auction_id = ANY($1);
	`
	bidRegistryRows, err := r.connection.QueryContext(ctx, bidRegistryQuery, pq.Array(orderedIDs))

	if err != nil {
		return nil, eris.Wrap(err, "Failed to query bid registry")
	}

	defer bidRegistryRows.Close()

	for bidRegistryRows.Next() {
		var registry BidRegistry
		var auctionID int

		if err := bidRegistryRows.Scan(&auctionID, &registry.Amount, &registry.DateAdd); err != nil {
			return nil, eris.Wrap(err, "Failed to scan registry")
		}

		if auction, ok := auctionMap[auctionID]; ok {
			auction.BidRegistry = append(auction.BidRegistry, &registry)
		}
	}

	if err = bidRegistryRows.Err(); err != nil {
		return nil, eris.Wrap(err, "Failed iterating bid registry rows")
	}

	auctions := make([]*Auction, 0, len(orderedIDs))

	for _, auctionID := range orderedIDs {
		auctions = append(auctions, auctionMap[auctionID])
	}

	return auctions, nil
}

func (r *PgAuctionRepository) GetActiveAuctionsFinishingIn(ctx context.Context, duration time.Duration) ([]*Auction, error) {
	query := `
		SELECT
			a.ta_id,
			a.ta_auction_id,
			a.ta_tibia_auction_link,
			a.ta_auction_start,
			a.ta_auction_end,
			tar.tar_status,
			tar.tar_date_add,
			tar.tar_date_upd
		FROM
			tc_auction a
		INNER JOIN
			tc_auction_recording tar ON a.ta_id = tar.tar_recordable_id
		WHERE
			tar.tar_status = 'active'
		AND
			a.ta_auction_end <= NOW() + make_interval(secs => $1)
		;
	`

	ctxTimeout, cancel := context.WithTimeout(ctx, time.Second*10)

	defer cancel()

	rows, err := r.connection.QueryContext(ctxTimeout, query, int(duration.Seconds()))

	if err != nil {
		return nil, eris.Wrap(err, "Failed to query active auctions")
	}

	var result []*Auction

	for rows.Next() {
		var (
			auction      Auction
			statusString string
		)

		err := rows.Scan(
			&auction.ID,
			&auction.AuctionID,
			&auction.TibiaAuctionLink,
			&auction.AuctionStart,
			&auction.AuctionEnd,
			&statusString,
			&auction.DateAdd,
			&auction.DateUpd,
		)

		if err != nil {
			return nil, eris.Wrap(err, "Failed to scan auction")
		}

		status, err := enums.GetAuctionRecordableStatusFromString(statusString)

		if err != nil {
			return nil, eris.Wrap(err, "Failed to parse auction status")
		}

		auction.Status = status

		result = append(result, &auction)
	}

	return result, nil
}

func (r *PgAuctionRepository) GetAuctionsWithFilter(ctx context.Context, filter *AuctionFilter) ([]*Auction, error) {
	query := `
		SELECT
			a.ta_id,
			a.ta_auction_id,
			a.ta_tibia_auction_link,
			a.ta_img,
			a.ta_char_name,
			a.ta_char_level,
			v.*,
			g.*,
			w.*,
			ts.*,
			a.ta_world_transfer,
			a.ta_boss_points,
			a.ta_charm_expansion,
			a.ta_charm_points,
			a.ta_task_expansion,
			a.ta_current_bid,
			a.ta_current_bid_fiat,
			a.ta_current_bid_currency,
			a.ta_auction_stage,
			a.ta_auction_start,
			a.ta_auction_end,
			tar.tar_status,
			tar.tar_date_add,
			tar.tar_date_upd
		FROM
			tc_auction a
		INNER JOIN
			tc_vocation v ON a.ta_char_vocation = v.tv_id
		INNER JOIN
			tc_gender g ON a.ta_char_gender = g.tg_id
		INNER JOIN
			tc_world w ON a.ta_char_world = w.tw_id
		INNER JOIN
			tc_auction_recording tar ON a.ta_id = tar.tar_recordable_id
		INNER JOIN
			tc_skills ts ON a.ta_auction_id = ts.ts_auction_id
		WHERE
			tar.tar_status = 'active'
		ORDER BY
			a.ta_auction_end ASC
		LIMIT $1 OFFSET $2;
	`

	ctxTimeout, cancel := context.WithTimeout(ctx, time.Second*10)

	defer cancel()

	offset := filter.Page * filter.Limit

	rows, err := r.connection.QueryContext(ctxTimeout, query, filter.Limit, offset)

	if err != nil {
		return nil, eris.Wrap(err, "Failed to query auctions with filter")
	}

	defer rows.Close()

	auctionMap := make(map[int]*Auction)

	var orderedIDs []int

	for rows.Next() {
		var (
			auction         Auction
			vocation        Vocation
			gender          Gender
			world           World
			skills          Skills
			battleEyeString string
			statusString    string
		)

		err := rows.Scan(
			&auction.ID,
			&auction.AuctionID,
			&auction.TibiaAuctionLink,
			&auction.Img,
			&auction.CharName,
			&auction.CharLevel,
			&vocation.Id,
			&vocation.Name,
			&gender.Id,
			&gender.Name,
			&world.Id,
			&world.Name,
			&world.Location,
			&battleEyeString,
			&world.Pvp,
			&skills.AuctionID,
			&skills.Axe,
			&skills.Club,
			&skills.Distance,
			&skills.Fishing,
			&skills.Fist,
			&skills.MagicLevel,
			&skills.Shielding,
			&skills.Sword,
			&auction.WorldTransfer,
			&auction.BossPoints,
			&auction.CharmExpansion,
			&auction.CharmPoints,
			&auction.TaskExpansion,
			&auction.Bid,
			&auction.BidFiat,
			&auction.BidCurrency,
			&auction.Stage,
			&auction.AuctionStart,
			&auction.AuctionEnd,
			&statusString,
			&auction.DateAdd,
			&auction.DateUpd,
		)

		if err != nil {
			return nil, eris.Wrap(err, "Failed to scan auction")
		}

		status, err := enums.GetAuctionRecordableStatusFromString(statusString)

		if err != nil {
			return nil, eris.Wrap(err, "Failed to parse auction status")
		}

		battleEyeEnum, err := enums.GetBattleEyeFromString(battleEyeString)

		if err != nil {
			return nil, eris.Wrap(err, "Failed parsing Battle Eye")
		}

		auction.Status = status
		auction.CharVocation = &vocation
		auction.CharGender = &gender
		world.BattleEye = battleEyeEnum
		auction.CharWorld = &world
		auction.Skills = &skills
		auction.Charms = make([]*Charm, 0)
		auction.FeaturedItems = make([]*FeaturedItem, 0)
		auction.Imbuements = make([]*Imbuement, 0)

		auctionMap[auction.AuctionID] = &auction

		orderedIDs = append(orderedIDs, auction.AuctionID)
	}

	if err = rows.Err(); err != nil {
		return nil, eris.Wrap(err, "Failed iterating auction rows")
	}

	if len(orderedIDs) == 0 {
		return []*Auction{}, nil
	}

	featuredItemsQuery := `
		SELECT
			tfi.tfi_id,
			tfi.tfi_auction_id,
			tfi.tfi_item_id
		FROM
			tc_featured_items tfi
		WHERE
			tfi.tfi_auction_id = ANY($1);
	`

	featuredItemsRows, err := r.connection.QueryContext(ctx, featuredItemsQuery, pq.Array(orderedIDs))

	if err != nil {
		return nil, eris.Wrap(err, "Failed to query featured items")
	}

	defer featuredItemsRows.Close()

	for featuredItemsRows.Next() {
		var item FeaturedItem

		if err := featuredItemsRows.Scan(&item.ID, &item.AuctionID, &item.ItemID); err != nil {
			return nil, eris.Wrap(err, "Failed to scan featured item")
		}

		if auction, ok := auctionMap[item.AuctionID]; ok {
			auction.FeaturedItems = append(auction.FeaturedItems, &item)
		}
	}

	if err = featuredItemsRows.Err(); err != nil {
		return nil, eris.Wrap(err, "Failed iterating featured item rows")
	}

	imbuementsQuery := `
		SELECT
			tai.tai_auction_id,
			ti.ti_id,
			ti.ti_name
		FROM
			tc_auction_imbuements tai
		INNER JOIN
			tc_imbuements ti ON tai.tai_imbuement_id = ti.ti_id
		WHERE
			tai.tai_auction_id = ANY($1);
	`
	imbuementRows, err := r.connection.QueryContext(ctx, imbuementsQuery, pq.Array(orderedIDs))

	if err != nil {
		return nil, eris.Wrap(err, "Failed to query imbuements")
	}

	defer imbuementRows.Close()

	for imbuementRows.Next() {
		var imbuement Imbuement
		var auctionID int

		if err := imbuementRows.Scan(&auctionID, &imbuement.ID, &imbuement.Name); err != nil {
			return nil, eris.Wrap(err, "Failed to scan imbuement")
		}

		if auction, ok := auctionMap[auctionID]; ok {
			auction.Imbuements = append(auction.Imbuements, &imbuement)
		}
	}

	if err = imbuementRows.Err(); err != nil {
		return nil, eris.Wrap(err, "Failed iterating imbuement rows")
	}

	charmsQuery := `
		SELECT
			tac.tac_auction_id,
			tac.tac_charm_id,
			tac.tac_grade,
			tc.tc_name,
			tc.tc_type
		FROM
			tc_auction_charms tac
		INNER JOIN
			tc_charms tc ON tac.tac_charm_id = tc.tc_id
		WHERE
			tac.tac_auction_id = ANY($1);
	`
	charmsRows, err := r.connection.QueryContext(ctx, charmsQuery, pq.Array(orderedIDs))

	if err != nil {
		return nil, eris.Wrap(err, "Failed to query charms")
	}

	defer charmsRows.Close()

	for charmsRows.Next() {
		var charm Charm
		var auctionID int

		if err := charmsRows.Scan(&auctionID, &charm.ID, &charm.Grade, &charm.Name, &charm.Type); err != nil {
			return nil, eris.Wrap(err, "Failed to scan charm")
		}

		if auction, ok := auctionMap[auctionID]; ok {
			auction.Charms = append(auction.Charms, &charm)
		}
	}

	if err = charmsRows.Err(); err != nil {
		return nil, eris.Wrap(err, "Failed iterating charms rows")
	}

	questsQuery := `
		SELECT
			taq.taq_auction_id,
			taq.taq_quest_id,
			tq.tq_name
		FROM
			tc_auction_quests taq
		INNER JOIN
			tc_quests tq ON taq.taq_quest_id = tq.tq_id
		WHERE
			taq.taq_auction_id = ANY($1);
	`
	questsRows, err := r.connection.QueryContext(ctx, questsQuery, pq.Array(orderedIDs))

	if err != nil {
		return nil, eris.Wrap(err, "Failed to query quests")
	}

	defer questsRows.Close()

	for questsRows.Next() {
		var quest Quest
		var auctionID int

		if err := questsRows.Scan(&auctionID, &quest.ID, &quest.Name); err != nil {
			return nil, eris.Wrap(err, "Failed to scan quest")
		}

		if auction, ok := auctionMap[auctionID]; ok {
			auction.Quests = append(auction.Quests, &quest)
		}
	}

	if err = questsRows.Err(); err != nil {
		return nil, eris.Wrap(err, "Failed iterating quests rows")
	}

	bidRegistryQuery := `
		SELECT
			ta.ta_auction_id,
			ta.ta_current_bid,
			ta.ta_date_add
		FROM
			tc_auction ta
		WHERE
			ta.ta_auction_id = ANY($1);
	`
	bidRegistryRows, err := r.connection.QueryContext(ctx, bidRegistryQuery, pq.Array(orderedIDs))

	if err != nil {
		return nil, eris.Wrap(err, "Failed to query bid registry")
	}

	defer bidRegistryRows.Close()

	for bidRegistryRows.Next() {
		var registry BidRegistry
		var auctionID int

		if err := bidRegistryRows.Scan(&auctionID, &registry.Amount, &registry.DateAdd); err != nil {
			return nil, eris.Wrap(err, "Failed to scan registry")
		}

		if auction, ok := auctionMap[auctionID]; ok {
			auction.BidRegistry = append(auction.BidRegistry, &registry)
		}
	}

	if err = bidRegistryRows.Err(); err != nil {
		return nil, eris.Wrap(err, "Failed iterating bid registry rows")
	}

	auctions := make([]*Auction, 0, len(orderedIDs))

	for _, auctionID := range orderedIDs {
		auctions = append(auctions, auctionMap[auctionID])
	}

	return auctions, nil
}

func (r *PgAuctionRepository) GetAuctionsPendingToConsolidate(ctx context.Context) ([]*Auction, error) {
	finishedAucQuery := `
		SELECT DISTINCT ON (tar.tar_auction_id)
			a.ta_id,
			a.ta_auction_id,
			a.ta_tibia_auction_link
		FROM
    		tc_auction_recording tar
		INNER JOIN tc_auction a ON tar.tar_recordable_id = a.ta_id
		WHERE
    		tar.tar_status = 'active'
		AND
    		a.ta_auction_end < now()
		ORDER BY tar.tar_auction_id, a.ta_date_add DESC
	`

	ctxTimeout, cancel := context.WithTimeout(ctx, time.Second*10)

	defer cancel()

	rows, err := r.connection.QueryContext(ctxTimeout, finishedAucQuery)

	if err != nil {
		return nil, eris.Wrap(err, "Failed to query auctions")
	}

	defer rows.Close()

	var result []*Auction

	for rows.Next() {
		var (
			auction Auction
		)

		err := rows.Scan(
			&auction.ID,
			&auction.AuctionID,
			&auction.TibiaAuctionLink,
		)

		if err != nil {
			return nil, eris.Wrap(err, "Failed to scan auction")
		}

		result = append(result, &auction)
	}

	incorrectBidStageQuery := `
			SELECT DISTINCT ON (tar.tar_auction_id)
				a.ta_id,
				a.ta_auction_id,
				a.ta_tibia_auction_link
			FROM
	    		tc_auction_recording tar
			INNER JOIN tc_auction a ON tar.tar_recordable_id = a.ta_id
			WHERE
	    		tar.tar_status = 'archived'
			AND
	    		a.ta_auction_stage = 'current'
			AND
				a.ta_date_add <= NOW() - INTERVAL '24 hours'
			ORDER BY tar.tar_auction_id, a.ta_date_add DESC
		`

	rows, err = r.connection.Query(incorrectBidStageQuery)

	if err != nil {
		return nil, eris.Wrap(err, "Failed to query auctions")
	}

	defer rows.Close()

	for rows.Next() {
		var (
			auction Auction
		)

		err := rows.Scan(
			&auction.ID,
			&auction.AuctionID,
			&auction.TibiaAuctionLink,
		)

		if err != nil {
			return nil, eris.Wrap(err, "Failed to scan auction")
		}

		result = append(result, &auction)
	}

	return result, nil
}

func (r *PgAuctionRepository) GetAllAuctionPrices(ctx context.Context) ([]*Auction, error) {
	finishedAucQuery := `
		SELECT
			a.ta_id,
			a.ta_auction_id,
			a.ta_current_bid
		FROM
    		tc_auction_recording tar
		INNER JOIN tc_auction a ON tar.tar_recordable_id = a.ta_id
		WHERE
			a.ta_auction_stage = 'current'
		OR
			a.ta_auction_stage = 'winning';
	`

	ctxTimeout, cancel := context.WithTimeout(ctx, time.Second*10)

	defer cancel()

	rows, err := r.connection.QueryContext(ctxTimeout, finishedAucQuery)

	if err != nil {
		return nil, eris.Wrap(err, "Failed to query all auctions prices")
	}

	defer rows.Close()

	var result []*Auction

	for rows.Next() {
		var (
			auction Auction
		)

		err := rows.Scan(
			&auction.ID,
			&auction.AuctionID,
			&auction.Bid,
		)

		if err != nil {
			return nil, eris.Wrap(err, "Failed to scan auction")
		}

		result = append(result, &auction)
	}

	return result, nil
}
