/*
================================================================================
TABLES
================================================================================
*/

CREATE TABLE IF NOT EXISTS tc_currency_rates (
    tcr_id SMALLINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    tcr_usd NUMERIC(4, 3) NOT NULL,
    tcr_eur NUMERIC(4, 3) NOT NULL,
    tcr_aud NUMERIC(4, 3) NOT NULL,
    tcr_gbp NUMERIC(4, 3) NOT NULL,
    tcr_pln NUMERIC(4, 3) NOT NULL,
    tcr_brl NUMERIC(4, 3) NOT NULL,
    tcr_date_upd TIMESTAMP(0) WITHOUT TIME ZONE NOT NULL
);

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
    ta_boss_points INT NOT NULL,
    ta_charm_expansion BOOLEAN NOT NULL,
    ta_charm_points INT NOT NULL,
    ta_task_expansion BOOLEAN NOT NULL,
    ta_current_bid INT NOT NULL,
    ta_current_bid_fiat INT NOT NULL,
    ta_current_bid_currency VARCHAR NOT NULL,
    ta_auction_stage VARCHAR NOT NULL,
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
        REFERENCES tc_auction_recording (tar_auction_id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS tc_featured_items (
    tfi_id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    tfi_auction_id INT NOT NULL,
    tfi_item_id INT,

    CONSTRAINT fk_featured_items_auction_recording
        FOREIGN KEY (tfi_auction_id)
        REFERENCES tc_auction_recording (tar_auction_id) ON DELETE CASCADE,

    CONSTRAINT unique_auction_item
        UNIQUE (tfi_auction_id, tfi_item_id)
);

CREATE TABLE IF NOT EXISTS tc_charms (
    tc_id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    tc_type VARCHAR NOT NULL,
    tc_name VARCHAR NOT NULL
);

CREATE TABLE IF NOT EXISTS tc_auction_charms (
    tac_auction_id INT NOT NULL,
    tac_charm_id INT NOT NULL,
    tac_grade SMALLINT NOT NULL,
    PRIMARY KEY (tac_auction_id, tac_charm_id),

    CONSTRAINT fk_auction_charms_auction_recording
        FOREIGN KEY (tac_auction_id)
        REFERENCES tc_auction_recording (tar_auction_id) ON DELETE CASCADE,

    CONSTRAINT fk_auction_charms_charms
        FOREIGN KEY (tac_charm_id)
        REFERENCES tc_charms (tc_id)
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
        REFERENCES tc_auction_recording (tar_auction_id) ON DELETE CASCADE,

    CONSTRAINT fk_auction_imbuements_imbuements
        FOREIGN KEY (tai_imbuement_id)
        REFERENCES tc_imbuements (ti_id)
);

CREATE TABLE IF NOT EXISTS tc_quests (
    tq_id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    tq_name VARCHAR NOT NULL
);

CREATE TABLE IF NOT EXISTS tc_auction_quests (
    taq_auction_id INT NOT NULL,
    taq_quest_id INT NOT NULL,
    PRIMARY KEY (taq_auction_id, taq_quest_id),

    CONSTRAINT fk_auction_quests_auction_recording
        FOREIGN KEY (taq_auction_id)
        REFERENCES tc_auction_recording (tar_auction_id) ON DELETE CASCADE,

    CONSTRAINT fk_auction_quests_quests
        FOREIGN KEY (taq_quest_id)
        REFERENCES tc_quests (tq_id)
);

CREATE TABLE IF NOT EXISTS tc_outfits (
    to_auction_id INT NOT NULL,
    to_name VARCHAR NOT NULL,
    to_addons INT NOT NULL,
    PRIMARY KEY(to_auction_id, to_name),

    CONSTRAINT fk_outfits_auction_recording
        FOREIGN KEY (to_auction_id)
        REFERENCES tc_auction_recording (tar_auction_id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS tc_mounts (
    tm_auction_id INT NOT NULL,
    tm_name VARCHAR NOT NULL,
    PRIMARY KEY(tm_auction_id, tm_name),

    CONSTRAINT fk_mounts_auction_recording
        FOREIGN KEY (tm_auction_id)
        REFERENCES tc_auction_recording (tar_auction_id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS tc_aggregated_auction_stats (
    taas_subset_key VARCHAR PRIMARY KEY,
    taas_median_price NUMERIC(20, 2) NOT NULL,
    taas_mean_price NUMERIC(20, 2) NOT NULL,
    taas_std_deviation NUMERIC(20, 2) NOT NULL,
    taas_min_price INT NOT NULL,
    taas_max_price INT NOT NULL,
    taas_mode_price INT NOT NULL,
    taas_sample_size INT NOT NULL,
    taas_date_upd TIMESTAMP(0) WITHOUT TIME ZONE NOT NULL
);

CREATE TABLE IF NOT EXISTS tc_flags (
    tf_id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    tf_name VARCHAR NOT NULL
);

CREATE TABLE IF NOT EXISTS tc_auction_flags (
    taf_auction_id INT NOT NULL,
    taf_flag_id INT NOT NULL,
    PRIMARY KEY (taf_auction_id, taf_flag_id),

    CONSTRAINT fk_auction_flags_auction_recording
        FOREIGN KEY (taf_auction_id)
        REFERENCES tc_auction_recording (tar_auction_id) ON DELETE CASCADE,

    CONSTRAINT fk_auction_flags_flags
        FOREIGN KEY (taf_flag_id)
        REFERENCES tc_flags (tf_id)
);

/*
================================================================================
INDEX
================================================================================
*/

CREATE INDEX idx_ta_fk_vocation ON tc_auction (ta_char_vocation);
CREATE INDEX idx_ta_fk_gender ON tc_auction (ta_char_gender);
CREATE INDEX idx_ta_fk_world ON tc_auction (ta_char_world);
CREATE INDEX idx_tar_auction_id_time ON tc_auction_recording (tar_auction_id, tar_date_add DESC);
CREATE INDEX idx_tai_imbuement_id ON tc_auction_imbuements (tai_imbuement_id);
CREATE INDEX idx_taq_quest_id ON tc_auction_quests (taq_quest_id);

/*
================================================================================
SEED DATA
================================================================================
*/

INSERT INTO tc_vocation (tv_name) VALUES
('Knight'),
('Paladin'),
('Sorcerer'),
('Druid'),
('Monk'),
('None');

INSERT INTO tc_gender (tg_name) VALUES
('Male'),
('Female');

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

INSERT INTO tc_quests (tq_name) VALUES
('The Postman Missions'),
('The Djinn War (blue)'),
('The Djinn War (green)'),
('The Travelling Trader (Rashid)'),
('The Thieves Guild'),
('Shadows of Yalahar'),
('The Pits of Inferno'),
('The Inquisition'),
('Barbarian Test'),
('Lion''s Rock'),
('The Shattered Isles'),
('The Ice Islands'),
('Twenty Miles Beneath the Sea'),
('The Explorer Society'),
('Blood Brothers'),
('The New Frontier'),
('Wrath of the Emperor'),
('The Ape City'),
('Rathleton (Citzen)'),
('Dark Trails'),
('Asura Palace'),
('The Dream Courts'),
('The Secret Library'),
('Soul War'),
('Primal Ordeal'),
('Rotten Blood'),
('Hero of Rathleton'),
('Cults of Tibia'),
('The Curse Spreads'),
('Grimvale'),
('Bigfoot''s Burden (Rank IV)'),
('Bigfoot''s Burden (Free boss access)'),
('Kilmaresh'),
('Heart of Destruction'),
('Feaster of Souls'),
('Dangerous Depths (Warzone 4)'),
('Dangerous Depths (Warzone 5)'),
('Dangerous Depths (Warzone 6)'),
('Ferumbras'' Ascendant'),
('The Order of the Cobra'),
('The Order of the Lion'),
('The Order of the Falcon');

INSERT INTO tc_charms (tc_type, tc_name) VALUES
('Minor', 'Adrenaline Burst'),
('Minor', 'Bless'),
('Major', 'Carnage'),
('Minor', 'Cleanse'),
('Minor', 'Cripple'),
('Major', 'Curse (Charm)'),
('Major', 'Divine Wrath'),
('Major', 'Dodge'),
('Major', 'Enflame'),
('Minor', 'Fatal Hold'),
('Major', 'Freeze'),
('Minor', 'Gut'),
('Major', 'Low Blow'),
('Minor', 'Numb'),
('Major', 'Overflux'),
('Major', 'Overpower'),
('Major', 'Parry'),
('Major', 'Poison'),
('Major', 'Savage Blow'),
('Minor', 'Scavenge'),
('Minor', 'Vampiric Embrace'),
('Minor', 'Void Inversion'),
('Minor', 'Voids Call'),
('Major', 'Wound'),
('Major', 'Zap'),
('Major', 'Curse');

INSERT INTO tc_flags (tf_name) VALUES
('none'),
('good_deal'),
('bad_deal'),
('hot'),
('featured'),
('rare_outfit'),
('rare_mount');