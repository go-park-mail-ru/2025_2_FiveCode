CREATE TABLE IF NOT EXISTS "user"
(
    id             INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    email          TEXT        NOT NULL UNIQUE CHECK (LENGTH(email) <= 40 AND
                                                      email ~ '^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$'),
    password_hash  TEXT        NOT NULL,
    username       TEXT        NOT NULL CHECK (LENGTH(username) >= 1 AND LENGTH(username) <= 40 AND
                                                      username ~ '^[a-zA-Z0-9_]+$'),
    
    avatar_file_id INTEGER,     
    
    created_at     TIMESTAMPTZ NOT NULL,
    updated_at     TIMESTAMPTZ NOT NULL
);
