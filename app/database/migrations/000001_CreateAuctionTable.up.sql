CREATE TABLE IF NOT EXISTS tc_world (
    tw_id SMALLINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    tw_name VARCHAR UNIQUE NOT NULL,
    tw_location VARCHAR NOT NULL,
    tw_battle_eye VARCHAR NOT NULL,
    tw_pvp VARCHAR NOT NULL
);

CREATE TABLE IF NOT EXISTS tc_vocation (
    tv_id SMALLINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    tv_name VARCHAR UNIQUE NOT NULL
);

CREATE TABLE IF NOT EXISTS tc_gender (
    tg_id SMALLINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    tg_name VARCHAR UNIQUE NOT NULL
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
    ta_world_transfer BOOLEAN NOT NULL,
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
        REFERENCES tc_world (tw_id)
);

CREATE INDEX idx_ta_fk_vocation ON tc_auction (ta_char_vocation);
CREATE INDEX idx_ta_fk_gender ON tc_auction (ta_char_gender);
CREATE INDEX idx_ta_fk_world ON tc_auction (ta_char_world);

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
    ts_auction_id INT PRIMARY KEY,
    ts_axe INT NOT NULL,
    ts_club INT NOT NULL,
    ts_distance INT NOT NULL,
    ts_fishing INT NOT NULL,
    ts_fist INT NOT NULL,
    ts_magic_level INT NOT NULL,
    ts_shielding INT NOT NULL,
    ts_sword INT NOT NULL,

    CONSTRAINT fk_skills_auction_recording
        FOREIGN KEY (ts_auction_id)
        REFERENCES tc_auction_recording (tar_auction_id)
);

CREATE TABLE IF NOT EXISTS tc_featured_items (
    tfi_id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    tfi_auction_id INT NOT NULL,
    tfi_item_id INT,

    CONSTRAINT fk_featured_items_auction_recording
        FOREIGN KEY (tfi_auction_id)
        REFERENCES tc_auction_recording (tar_auction_id),

    CONSTRAINT unique_auction_item
        UNIQUE (tfi_auction_id, tfi_item_id)
);

CREATE TABLE IF NOT EXISTS tc_charm (
    tc_auction_id INT PRIMARY KEY,
    tc_expansion BOOLEAN,
    tc_points INT,

    CONSTRAINT fk_charm_auction_recording
        FOREIGN KEY (tc_auction_id)
        REFERENCES tc_auction_recording (tar_auction_id)
);

CREATE TABLE IF NOT EXISTS tc_imbuements (
    ti_id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    ti_name VARCHAR
);

CREATE TABLE IF NOT EXISTS tc_auction_imbuements (
    tai_auction_id INT NOT NULL,
    tai_imbuement_id INT NOT NULL,
    PRIMARY KEY (tai_auction_id, tai_imbuement_id),

    CONSTRAINT fk_auction_imbuements_auction_recording
        FOREIGN KEY (tai_auction_id)
        REFERENCES tc_auction_recording (tar_auction_id),

    CONSTRAINT fk_auction_imbuements_imbuements
        FOREIGN KEY (tai_imbuement_id)
        REFERENCES tc_imbuements (ti_id)
);

CREATE INDEX idx_tai_imbuement ON tc_auction_imbuements (tai_imbuement_id);

INSERT INTO tc_vocation (tv_name) VALUES ('Knight'), ('Paladin'), ('Sorcerer'), ('Druid'), ('Monk'), ('None');

INSERT INTO tc_gender (tg_name) VALUES ('Male'), ('Female');

INSERT INTO tc_imbuements (ti_name) VALUES
('Powerful Bash'),
('Powerful Blockade'),
('Powerful Chop'),
('Powerful Cloud Fabric'),
('Powerful Demon Presence'),
('Powerful Dragon Hide'),
('Powerful Electrify'),
('Powerful Epiphany'),
('Powerful Featherweight'),
('Powerful Frost'),
('Powerful Lich Shroud'),
('Powerful Precision'),
('Powerful Punch'),
('Powerful Quara Scale'),
('Powerful Reap'),
('Powerful Scorch'),
('Powerful Slash'),
('Powerful Snake Skin'),
('Powerful Strike'),
('Powerful Swiftness'),
('Powerful Vampirism'),
('Powerful Venom'),
('Powerful Vibrancy'),
('Powerful Void');