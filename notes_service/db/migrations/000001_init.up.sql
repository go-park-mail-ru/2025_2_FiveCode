-- ENUM TYPES
CREATE TYPE block_type AS ENUM ('text', 'code', 'attachment');
CREATE TYPE note_role AS ENUM ('editor', 'commenter', 'viewer');
CREATE TYPE text_font AS ENUM ('Inter', 'Roboto', 'Montserrat', 'Manrope');

-- NOTE
CREATE TABLE IF NOT EXISTS note
(
    id                   INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    owner_id             INTEGER,
    parent_note_id       INTEGER REFERENCES note (id) ON DELETE CASCADE,
    title                TEXT        NOT NULL CHECK (LENGTH(title) >= 1 AND LENGTH(title) <= 200),
    icon_file_id         INTEGER,
    is_archived          BOOLEAN     NOT NULL DEFAULT false,
    is_shared            BOOLEAN     NOT NULL DEFAULT false,
    public_access_level  note_role   DEFAULT NULL,
    share_uuid           UUID        DEFAULT NULL UNIQUE,
    created_at           TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at           TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at           TIMESTAMPTZ
);

-- BLOCK
CREATE TABLE IF NOT EXISTS block
(
    id             INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    note_id        INTEGER        NOT NULL REFERENCES note (id) ON DELETE CASCADE,
    type           block_type,
    position       NUMERIC(12, 6) NOT NULL CHECK (position >= 0),
    created_at     TIMESTAMPTZ    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at     TIMESTAMPTZ    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_edited_by INTEGER
);

CREATE TABLE block_text
(
    id         INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    block_id   INTEGER UNIQUE NOT NULL REFERENCES block (id) ON DELETE CASCADE,
    text       TEXT           NOT NULL,
    created_at TIMESTAMPTZ    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ    NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- BLOCK_TEXT_FORMAT
CREATE TABLE block_text_format
(
    id            INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    block_text_id INTEGER     NOT NULL REFERENCES block_text (id) ON DELETE CASCADE,
    start_offset  INTEGER     NOT NULL CHECK (start_offset >= 0),
    end_offset    INTEGER     NOT NULL CHECK (end_offset > start_offset),
    bold          BOOLEAN              DEFAULT false,
    italic        BOOLEAN              DEFAULT false,
    underline     BOOLEAN              DEFAULT false,
    strikethrough BOOLEAN              DEFAULT false,
    link          TEXT CHECK (link IS NULL OR link ~ '^https?:\/\/.+'),
    font          text_font            DEFAULT 'Inter',
    size          INTEGER              DEFAULT 12 CHECK (size > 0 AND size <= 72),
    created_at    TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- BLOCK_CODE
CREATE TABLE IF NOT EXISTS block_code
(
    block_id   INTEGER PRIMARY KEY REFERENCES block (id) ON DELETE CASCADE,
    language   TEXT CHECK (LENGTH(language) <= 50),
    code_text  TEXT        NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- BLOCK_ATTACHMENT
CREATE TABLE IF NOT EXISTS block_attachment
(
    id         INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    block_id   INTEGER REFERENCES block (id) ON DELETE CASCADE,
    file_id    INTEGER     NOT NULL,
    caption    TEXT CHECK (LENGTH(caption) <= 255),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- NOTE_PERMISSION
CREATE TABLE IF NOT EXISTS note_permission
(
    note_permission_id INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    note_id            INTEGER REFERENCES note (id) ON DELETE CASCADE,
    granted_by         INTEGER,
    granted_to         INTEGER,
    role               note_role,
    can_share          BOOLEAN     NOT NULL DEFAULT false,
    created_at         TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at         TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT unique_note_granted_to UNIQUE (note_id, granted_to)
);

-- FAVORITE
CREATE TABLE IF NOT EXISTS favorite
(
    user_id    INTEGER     NOT NULL,
    note_id    INTEGER     NOT NULL REFERENCES note (id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (user_id, note_id)
);

-- TAG
CREATE TABLE IF NOT EXISTS tag
(
    id         INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name       TEXT        NOT NULL UNIQUE CHECK (LENGTH(name) <= 50),
    created_by INTEGER,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- NOTE_TAG
CREATE TABLE IF NOT EXISTS note_tag
(
    note_id    INTEGER     NOT NULL REFERENCES note (id) ON DELETE CASCADE,
    tag_id     INTEGER     NOT NULL REFERENCES tag (id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (note_id, tag_id)
);
