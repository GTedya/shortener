START TRANSACTION;

ALTER TABLE urls ADD COLUMN is_deleted bool DEFAULT false;

COMMIT