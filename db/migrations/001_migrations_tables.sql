-- 001_migrations_table.sql


-- ENUM TYPES
CREATE TYPE block_type AS ENUM ('text', 'code', 'attachment');
CREATE TYPE note_role AS ENUM ('editor', 'commenter', 'viewer');


-- FILE
CREATE TABLE IF NOT EXISTS file
(
    id         SERIAL PRIMARY KEY,
    url        TEXT        NOT NULL UNIQUE CHECK (LENGTH(url) <= 255),
    mime_type  TEXT        NOT NULL CHECK (LENGTH(mime_type) <= 50),
    size_bytes INTEGER     NOT NULL CHECK (size_bytes >= 0),
    width      INTEGER CHECK (width >= 0),
    height     INTEGER CHECK (height >= 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);


-- USER
CREATE TABLE IF NOT EXISTS user
(
    id             SERIAL PRIMARY KEY,
    email          TEXT        NOT NULL UNIQUE CHECK (LENGTH(email) <= 40),
    password_hash  TEXT        NOT NULL,
    username       TEXT        NOT NULL UNIQUE CHECK (LENGTH(username) >= 3 AND LENGTH(username) <= 40),
    avatar_file_id INTEGER     REFERENCES file (id) ON DELETE SET NULL,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);


-- NOTE
CREATE TABLE IF NOT EXISTS note
(
    id             SERIAL PRIMARY KEY,
    owner_id       INTEGER     NOT NULL REFERENCES user (id) ON DELETE SET NULL,
    parent_note_id INTEGER REFERENCES note (id) ON DELETE CASCADE,
    title          TEXT        NOT NULL CHECK (LENGTH(title) >= 1 AND LENGTH(title) <= 200),
    icon_file_id   INTEGER     REFERENCES file (id) ON DELETE SET NULL,
    is_archived    BOOLEAN     NOT NULL DEFAULT false,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at     TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);


-- BLOCK
CREATE TABLE IF NOT EXISTS block
(
    id         SERIAL PRIMARY KEY,
    note_id    INTEGER        NOT NULL REFERENCES note (id) ON DELETE CASCADE,
    type       block_type,
    position   NUMERIC(12, 6) NOT NULL CHECK (position >= 0),
    created_at TIMESTAMPTZ    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_edited_by TEXT REFERENCES user (id) ON DELETE SET NULL
);


-- BLOCK_TEXT_SPAN
CREATE TABLE IF NOT EXISTS block_text_span
(
    block_id      INTEGER PRIMARY KEY REFERENCES block (id) ON DELETE CASCADE,
    position      NUMERIC(12, 6) NOT NULL CHECK (position >= 0),
    text          TEXT           NOT NULL,
    bold          BOOLEAN                 DEFAULT false,
    italic        BOOLEAN                 DEFAULT false,
    underline     BOOLEAN                 DEFAULT false,
    strikethrough BOOLEAN                 DEFAULT false,
    font          TEXT                    DEFAULT 'Times New Roman' CHECK (LENGTH(font) <= 50),
    size          INTEGER                 DEFAULT 12 CHECK (size > 0 AND size <= 72),
    created_at    TIMESTAMPTZ    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at    TIMESTAMPTZ    NOT NULL DEFAULT CURRENT_TIMESTAMP
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
    id         SERIAL PRIMARY KEY,
    block_id   INTEGER REFERENCES block (id) ON DELETE CASCADE,
    file_id    INTEGER     NOT NULL REFERENCES file (id) ON DELETE CASCADE,
    caption    TEXT        NOT NULL CHECK (LENGTH(caption) <= 255),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);


-- NOTE_PERMISSION
CREATE TABLE IF NOT EXISTS note_permission
(
    note_permission_id SERIAL PRIMARY KEY,
    note_id            INTEGER REFERENCES note (id) ON DELETE CASCADE,
    granted_by         INTEGER REFERENCES user (id),
    granted_to         INTEGER REFERENCES user (id),
    role               note_role,
    can_share          BOOLEAN     NOT NULL DEFAULT false,
    granted_at         TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at         TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);


-- FAVORITE
CREATE TABLE IF NOT EXISTS favorite
(
    user_id    INTEGER     NOT NULL REFERENCES user (id) ON DELETE CASCADE,
    note_id    INTEGER     NOT NULL REFERENCES note (id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (user_id, note_id)
);


-- TAG
CREATE TABLE IF NOT EXISTS tag
(
    id         SERIAL PRIMARY KEY,
    name       TEXT        NOT NULL UNIQUE CHECK (LENGTH(name) <= 50),
    created_by INTEGER REFERENCES user (id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);


-- NOTE_TAG
CREATE TABLE IF NOT EXISTS note_tag
(
    note_id    INTEGER     NOT NULL REFERENCES note (id) ON DELETE CASCADE,
    tag_id     INTEGER     NOT NULL REFERENCES tag (id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (note_id, tag_id)
);
