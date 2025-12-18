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
    header_file_id       INTEGER,
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

CREATE MATERIALIZED VIEW note_search_index AS
SELECT
    n.id as note_id,
    n.owner_id,
    n.title,
    n.updated_at,
    n.deleted_at,
    COALESCE(string_agg(DISTINCT bt.text, ' '), '') as text_content,
    COALESCE(string_agg(DISTINCT bc.code_text, ' '), '') as code_content,
    COALESCE(string_agg(DISTINCT bt.text, ' '), '') || ' ' ||
    COALESCE(string_agg(DISTINCT bc.code_text, ' '), '') as content,

    setweight(to_tsvector('russian', COALESCE(n.title, '')), 'A') ||
    setweight(to_tsvector('english', COALESCE(n.title, '')), 'A') ||
    setweight(to_tsvector('russian', COALESCE(string_agg(DISTINCT bt.text, ' '), '')), 'A') ||
    setweight(to_tsvector('english', COALESCE(string_agg(DISTINCT bt.text, ' '), '')), 'A') ||
    setweight(to_tsvector('russian', COALESCE(string_agg(DISTINCT bc.code_text, ' '), '')), 'A') ||
    setweight(to_tsvector('english', COALESCE(string_agg(DISTINCT bc.code_text, ' '), '')), 'A') as search_vector
FROM note n
         LEFT JOIN block b ON b.note_id = n.id
         LEFT JOIN block_text bt ON bt.block_id = b.id AND b.type = 'text'
         LEFT JOIN block_code bc ON bc.block_id = b.id AND b.type = 'code'
GROUP BY n.id, n.owner_id, n.title, n.updated_at, n.deleted_at;

CREATE INDEX idx_note_search_vector ON note_search_index USING GIN(search_vector);

CREATE INDEX idx_note_search_filter ON note_search_index(owner_id, deleted_at, updated_at DESC);

CREATE UNIQUE INDEX idx_note_search_id ON note_search_index(note_id);

CREATE OR REPLACE FUNCTION refresh_search_index()
    RETURNS TRIGGER AS $$
BEGIN
    PERFORM pg_notify('refresh_search_index', '');
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER note_search_refresh_note
    AFTER INSERT OR UPDATE OR DELETE ON note
    FOR EACH STATEMENT
EXECUTE FUNCTION refresh_search_index();

CREATE TRIGGER note_search_refresh_block
    AFTER INSERT OR UPDATE OR DELETE ON block
    FOR EACH STATEMENT
EXECUTE FUNCTION refresh_search_index();

CREATE TRIGGER note_search_refresh_block_text
    AFTER INSERT OR UPDATE OR DELETE ON block_text
    FOR EACH STATEMENT
EXECUTE FUNCTION refresh_search_index();

CREATE TRIGGER note_search_refresh_block_code
    AFTER INSERT OR UPDATE OR DELETE ON block_code
    FOR EACH STATEMENT
EXECUTE FUNCTION refresh_search_index();

-- Триграммный поиск для поиска подстрок
CREATE EXTENSION IF NOT EXISTS pg_trgm;
CREATE INDEX idx_note_content_trgm ON note_search_index USING GIN(content gin_trgm_ops);
CREATE INDEX idx_note_title_trgm ON note_search_index USING GIN(title gin_trgm_ops);