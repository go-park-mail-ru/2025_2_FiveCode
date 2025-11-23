CREATE TABLE IF NOT EXISTS file
(
    id         INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    url        TEXT        NOT NULL UNIQUE CHECK (LENGTH(url) <= 255 AND
                                                  url ~ '^https?:\/\/([a-zA-Z0-9.-]+)(:[0-9]+)?(\/.*)?$'),
    mime_type  TEXT        NOT NULL CHECK (LENGTH(mime_type) <= 50),
    size_bytes INTEGER     NOT NULL CHECK (size_bytes >= 0 AND size_bytes <= 1024 * 1024 * 1024), -- 1 гб
    width      INTEGER CHECK (width >= 0),
    height     INTEGER CHECK (height >= 0),
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);
