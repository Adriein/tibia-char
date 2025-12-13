CREATE TABLE IF NOT EXISTS tc_world (
    tw_id SERIAL PRIMARY KEY,
    tw_name VARCHAR UNIQUE NOT NULL
);

CREATE TABLE IF NOT EXISTS tc_vocation (
    tv_id SERIAL PRIMARY KEY,
    tv_name VARCHAR UNIQUE NOT NULL
);

CREATE TABLE IF NOT EXISTS tc_gender (
    tg_id SMALLSERIAL PRIMARY KEY,
    tg_name VARCHAR UNIQUE NOT NULL
);


CREATE TABLE IF NOT EXISTS tc_auction (
    ta_id INT PRIMARY KEY,
    ta_tibia_auction_id INT UNIQUE NOT NULL,
    ta_img VARCHAR NOT NULL,
    ta_char_name VARCHAR NOT NULL,
    ta_char_level INT NOT NULL,
    ta_char_vocation INT NOT NULL,
    ta_char_gender SMALLINT NOT NULL,
    ta_char_world INT NOT NULL,
    ta_current_bid INT NOT NULL,
    ta_auction_start TIMESTAMP WITHOUT TIME ZONE NOT NULL,
    ta_auction_end TIMESTAMP WITHOUT TIME ZONE NOT NULL,
    ta_is_active BOOLEAN NOT NULL,
    ta_date_add TIMESTAMP WITHOUT TIME ZONE NOT NULL,
    ta_date_upd TIMESTAMP WITHOUT TIME ZONE NOT NULL,

    CONSTRAINT fk_vocation_auction
        FOREIGN KEY (ta_char_vocation)
        REFERENCES tc_vocation (tv_id),

    CONSTRAINT fk_gender_auction
        FOREIGN KEY (ta_char_gender)
        REFERENCES tc_gender (tg_id),

    CONSTRAINT fk_world_auction
        FOREIGN KEY (ta_char_world)
        REFERENCES tc_world (tw_id)
);

CREATE INDEX idx_ta_active_vocation_world ON tc_auction (
    ta_is_active,
    ta_char_vocation,
    ta_char_world
);

CREATE INDEX idx_ta_fk_vocation ON tc_auction (ta_char_vocation);
CREATE INDEX idx_ta_fk_gender ON tc_auction (ta_char_gender);
CREATE INDEX idx_ta_fk_world ON tc_auction (ta_char_world);

CREATE TABLE IF NOT EXISTS tc_bid_history (
    tbh_id BIGSERIAL PRIMARY KEY,
    tbh_auction_id INT NOT NULL,
    tbh_bid INT NOT NULL,
    tbh_date_add TIMESTAMP WITHOUT TIME ZONE NOT NULL,

    CONSTRAINT fk_bid_auction
        FOREIGN KEY (tbh_auction_id)
        REFERENCES tc_auction (ta_id)
        ON DELETE CASCADE
);

CREATE INDEX idx_tbh_auction_id_time ON tc_bid_history (tbh_auction_id, tbh_date_add DESC);