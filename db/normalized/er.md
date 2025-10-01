```mermaid
erDiagram
  USER {
    INTEGER id
    TEXT email
    TEXT password_hash
    TEXT username
    INTEGER avatar_file_id
    TIMESTAMPTZ created_at
    TIMESTAMPTZ updated_at
  }

  FILE {
    INTEGER id
    TEXT url
    TEXT mime_type
    INTEGER size_bytes
    INTEGER width
    INTEGER height
    TIMESTAMPTZ created_at
  }

  NOTE {
    INTEGER id
    INTEGER owner_id
    INTEGER parent_note_id
    TEXT title
    INTEGER icon_file_id
    BOOLEAN is_archived
    TIMESTAMPTZ created_at
    TIMESTAMPTZ updated_at
    TIMESTAMPTZ deleted_at
  }

  BLOCK {
    INTEGER id
    INTEGER note_id
    TEXT type
    NUMERIC position
    TIMESTAMPTZ created_at
    TIMESTAMPTZ updated_at
    INTEGER last_edited_by
  }

  BLOCK_TEXT_SPAN {
    INTEGER block_id
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
    INTEGER block_id
    TEXT language
    TEXT code_text
    TIMESTAMPTZ created_at
    TIMESTAMPTZ updated_at
  }

  BLOCK_ATTACHMENT {
    INTEGER id
    INTEGER block_id
    INTEGER file_id
    TEXT caption
    TIMESTAMPTZ created_at
  }

  NOTE_PERMISSION {
    INTEGER note_permission_id
    INTEGER note_id
    INTEGER granted_by
    INTEGER granted_to
    TEXT role
    BOOLEAN can_share
    TIMESTAMPTZ granted_at
    TIMESTAMPTZ updated_at
  }

  FAVORITE {
    INTEGER user_id
    INTEGER note_id
    TIMESTAMPTZ created_at
  }

  TAG {
    INTEGER id
    TEXT name
    INTEGER created_by
    TIMESTAMPTZ updated_at
    TIMESTAMPTZ created_at
  }

  NOTE_TAG {
    INTEGER note_id
    INTEGER tag_id
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
  USER ||--o{ BLOCK : last_edited_by