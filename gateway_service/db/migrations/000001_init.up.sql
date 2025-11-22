CREATE OR REPLACE FUNCTION set_timestamps()
    RETURNS TRIGGER AS
$$
BEGIN
    IF (TG_OP = 'INSERT') THEN
        NEW.created_at := CURRENT_TIMESTAMP;
        NEW.updated_at := CURRENT_TIMESTAMP;
        RETURN NEW;
    END IF;

    IF (TG_OP = 'UPDATE') THEN
        NEW.updated_at := CURRENT_TIMESTAMP;
        RETURN NEW;
    END IF;
END;
$$ LANGUAGE plpgsql;

CREATE TABLE IF NOT EXISTS file
(
    id         INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    url        TEXT        NOT NULL UNIQUE CHECK (LENGTH(url) <= 255 AND
                                                  url ~ '^https?:\/\/([a-zA-Z0-9.-]+)(:[0-9]+)?(\/.*)?$'),
    mime_type  TEXT        NOT NULL CHECK (LENGTH(mime_type) <= 50),
    size_bytes INTEGER     NOT NULL CHECK (size_bytes >= 0 AND size_bytes <= 1024 * 1024 * 1024), -- 1 гб
    width      INTEGER CHECK (width >= 0),
    height     INTEGER CHECK (height >= 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TRIGGER trigger_set_timestamps
    BEFORE INSERT OR UPDATE
    ON file
    FOR EACH ROW
EXECUTE FUNCTION set_timestamps();