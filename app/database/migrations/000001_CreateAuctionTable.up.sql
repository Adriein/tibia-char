CREATE TABLE IF NOT EXISTS tc_world (
    tw_id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    tw_name VARCHAR UNIQUE NOT NULL,
    tw_location VARCHAR NOT NULL,
    tw_battle_eye VARCHAR NOT NULL,
    tw_pvp VARCHAR NOT NULL
);

CREATE TABLE IF NOT EXISTS tc_vocation (
    tv_id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    tv_name VARCHAR UNIQUE NOT NULL
);

CREATE TABLE IF NOT EXISTS tc_gender (
    tg_id SMALLSERIAL PRIMARY KEY,
    tg_name VARCHAR UNIQUE NOT NULL
);

CREATE TABLE IF NOT EXISTS tc_world_transfer (
    twt_id SMALLSERIAL PRIMARY KEY,
    twt_name VARCHAR UNIQUE NOT NULL
);


CREATE TABLE IF NOT EXISTS tc_auction (
    ta_id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    ta_auction_id INT NOT NULL,
    ta_tibia_auction_link VARCHAR NOT NULL,
    ta_img VARCHAR NOT NULL,
    ta_char_name VARCHAR NOT NULL,
    ta_char_level INT NOT NULL,
    ta_char_vocation INT NOT NULL,
    ta_char_gender SMALLINT NOT NULL,
    ta_char_world INT NOT NULL,
    ta_world_transfer INT NOT NULL,
    ta_current_bid INT NOT NULL,
    ta_auction_start TIMESTAMP(0) WITHOUT TIME ZONE NOT NULL,
    ta_auction_end TIMESTAMP(0) WITHOUT TIME ZONE NOT NULL,
    ta_date_add TIMESTAMP(0) WITHOUT TIME ZONE NOT NULL,
    ta_date_upd TIMESTAMP(0) WITHOUT TIME ZONE NOT NULL,

    CONSTRAINT fk_vocation_auction
        FOREIGN KEY (ta_char_vocation)
        REFERENCES tc_vocation (tv_id),

    CONSTRAINT fk_gender_auction
        FOREIGN KEY (ta_char_gender)
        REFERENCES tc_gender (tg_id),

    CONSTRAINT fk_world_auction
        FOREIGN KEY (ta_char_world)
        REFERENCES tc_world (tw_id),

    CONSTRAINT fk_world_transfer_auction
        FOREIGN KEY (ta_world_transfer)
        REFERENCES tc_world_transfer (twt_id)
);

CREATE INDEX idx_ta_fk_vocation ON tc_auction (ta_char_vocation);
CREATE INDEX idx_ta_fk_gender ON tc_auction (ta_char_gender);
CREATE INDEX idx_ta_fk_world ON tc_auction (ta_char_world);
CREATE INDEX idx_ta_fk_world_transfer ON tc_auction (ta_world_transfer);

CREATE TABLE IF NOT EXISTS tc_auction_recording (
    tar_auction_id INT PRIMARY KEY,
    tar_recordable_id BIGINT NOT NULL,
    tar_status VARCHAR NOT NULL,
    tar_date_add TIMESTAMP(0) WITHOUT TIME ZONE NOT NULL,
    tar_date_upd TIMESTAMP(0) WITHOUT TIME ZONE NOT NULL,

    CONSTRAINT fk_recording_auction
        FOREIGN KEY (tar_recordable_id)
        REFERENCES tc_auction (ta_id)
);

CREATE INDEX idx_tar_auction_id_time ON tc_auction_recording (tar_auction_id, tar_date_add DESC);

CREATE TABLE IF NOT EXISTS tc_skills (
    tas_id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    tas_auction_id BIGINT NOT NULL,
    tas_axe INT NOT NULL,
    tas_club INT NOT NULL,
    tas_distance INT NOT NULL,
    tas_fishing INT NOT NULL,
    tas_fist INT NOT NULL,
    tas_magic_level INT NOT NULL,
    tas_shielding INT NOT NULL,
    tas_sword INT NOT NULL,

    CONSTRAINT fk_skills_auction_recording
        FOREIGN KEY (tas_auction_id)
        REFERENCES tc_auction_recording (tar_auction_id)
);

INSERT INTO tc_vocation (tv_name) VALUES ('Knight'), ('Paladin'), ('Sorcerer'), ('Druid'), ('Monk');

INSERT INTO tc_gender (tg_name) VALUES ('Male'), ('Female');

INSERT INTO tc_world_transfer (twt_name) VALUES ('immediately'), ('forbidden');