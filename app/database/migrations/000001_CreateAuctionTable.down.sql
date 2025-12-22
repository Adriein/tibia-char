ALTER TABLE tc_auction DROP CONSTRAINT fk_vocation_auction;
ALTER TABLE tc_auction DROP CONSTRAINT fk_gender_auction;
ALTER TABLE tc_auction DROP CONSTRAINT fk_world_auction;

ALTER TABLE tc_auction_recording DROP CONSTRAINT fk_recording_auction;

DROP TABLE tc_auction_recording;

DROP TABLE tc_world;
DROP TABLE tc_vocation;
DROP TABLE tc_gender;

DROP TABLE tc_auction;

