```mermaid
erDiagram
  USER {
    UUID id
    TEXT email
    TEXT password
    TEXT username
    UUID avatar_file_id
    TIMESTAMPTZ created_at
    TIMESTAMPTZ updated_at
  }

  FILE {
    UUID id
    TEXT url
    TEXT mime_type
    INTEGER size_bytes
    INTEGER width
    INTEGER height
    TIMESTAMPTZ created_at
  }

  NOTE {
    UUID id
    UUID owner_id
    UUID parent_note_id
    TEXT title
    UUID icon_file_id
    BOOLEAN is_archived
    TIMESTAMPTZ created_at
    TIMESTAMPTZ updated_at
    TIMESTAMPTZ deleted_at
  }

  BLOCK {
    UUID id
    UUID note_id
    TEXT type
    NUMERIC position
    TIMESTAMPTZ created_at
    TIMESTAMPTZ updated_at
  }

  BLOCK_TEXT_SPAN {
    UUID block_id
    NUMERIC position
    TEXT text
    BOOLEAN bold
    BOOLEAN italic
    BOOLEAN underline
    BOOLEAN strikethrough
    TEXT font
    INTEGER size
    TIMESTAMPTZ created_at
    TIMESTAMPTZ updated_at
  }

  BLOCK_CODE {
    UUID block_id
    TEXT language
    TEXT code_text
    TIMESTAMPTZ created_at
    TIMESTAMPTZ updated_at
  }

  BLOCK_BLOCK_ATTACHMENT {
    UUID id
    UUID block_id
    UUID file_id
    TEXT caption
    TIMESTAMPTZ created_at
  }

  NOTE_PERMISSION {
    UUID note_permission_id
    UUID note_id
    UUID granted_by
    UUID granted_to
    TEXT role
    BOOLEAN can_share
    TIMESTAMPTZ granted_at
    TIMESTAMPTZ updated_at
  }

  FAVORITE {
    UUID user_id
    UUID note_id
    TIMESTAMPTZ created_at
  }

  TAG {
    UUID id
    TEXT name
    UUID created_by
    TIMESTAMPTZ updated_at
    TIMESTAMPTZ created_at
  }

  NOTE_TAG {
    UUID note_id
    UUID tag_id
    TIMESTAMPTZ created_at
  }

  USER ||--o{ NOTE : owns
  NOTE |o--o{ NOTE : parent_of
  NOTE ||--o{ BLOCK : contains
  BLOCK ||--o{ BLOCK_TEXT_SPAN : has
  BLOCK ||--o| BLOCK_CODE : opts
  FILE |o--|| BLOCK_ATTACHMENT : used_by
  BLOCK ||--o| BLOCK_ATTACHMENT : embeds
  USER ||--o{ NOTE_PERMISSION : granted_to
  USER ||--o{ NOTE_PERMISSION : granted_by
  NOTE ||--o{ NOTE_PERMISSION : shared_note
  USER ||--o{ FAVORITE : stars
  NOTE ||--o{ FAVORITE : starred
  TAG ||--o{ NOTE_TAG : link
  NOTE ||--o{ NOTE_TAG : link
  FILE |o--o| USER : avatar
  FILE |o--o| NOTE : icon
