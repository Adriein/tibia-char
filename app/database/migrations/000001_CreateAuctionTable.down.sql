ALTER TABLE tc_auction DROP CONSTRAINT fk_vocation_auction;
ALTER TABLE tc_auction DROP CONSTRAINT fk_gender_auction;
ALTER TABLE tc_auction DROP CONSTRAINT fk_world_auction;

ALTER TABLE tc_auction_recording DROP CONSTRAINT fk_recording_auction;
ALTER TABLE tc_skills DROP CONSTRAINT fk_skills_auction_recording;
ALTER TABLE tc_charm DROP CONSTRAINT fk_charm_auction_recording;

DROP TABLE tc_charm;
DROP TABLE tc_skills;
DROP TABLE tc_featured_items;
DROP TABLE tc_auction_recording;

DROP TABLE tc_world;
DROP TABLE tc_vocation;
DROP TABLE tc_gender;

DROP TABLE tc_auction;

