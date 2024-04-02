START TRANSACTION;

CREATE TABLE IF NOT EXISTS urls
(
    id        INT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    short_url VARCHAR(20) UNIQUE  NOT NULL,
    url       VARCHAR(200) UNIQUE NOT NULL
);

COMMIT