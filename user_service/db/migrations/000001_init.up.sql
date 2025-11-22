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

CREATE TABLE IF NOT EXISTS "user"
(
    id             INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    email          TEXT        NOT NULL UNIQUE CHECK (LENGTH(email) <= 40 AND
                                                      email ~ '^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$'),
    password_hash  TEXT        NOT NULL,
    username       TEXT        NOT NULL UNIQUE CHECK (LENGTH(username) >= 1 AND LENGTH(username) <= 40 AND
                                                      username ~ '^[a-zA-Z0-9_]+$'),
    
    avatar_file_id INTEGER,     
    
    created_at     TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TRIGGER trigger_set_timestamps
    BEFORE INSERT OR UPDATE
    ON "user"
    FOR EACH ROW
EXECUTE FUNCTION set_timestamps();