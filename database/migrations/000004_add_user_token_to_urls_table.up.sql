START TRANSACTION;

ALTER TABLE urls ADD COLUMN user_token text;

COMMIT