```mermaid
erDiagram
  USER {
    id
    email
    password
    username
    avatar_file_id
    created_at
    updated_at
  }
  FILE {
    id
    url
    mime_type
    size_bytes
    width
    height
    created_at
  }
  NOTE {
    id
    owner_id
    parent_note_id
    title
    icon_file_id
    is_archived
    created_at
    updated_at
    deleted_at
  }
  BLOCK {
    id
    note_id
    type
    position
    created_at
    updated_at
  }
  BLOCK_TEXT_SPAN {
    block_id
    position
    text
    bold
    italic
    underline
    strikethrough
    font
    size
    created_at
    updated_at
  }
  BLOCK_CODE {
    block_id
    language
    code_text
    created_at
    updated_at
  }
  ATTACHMENT {
    id
    block_id
    file_id
    caption
    created_at
  }
  NOTE_PERMISSION {
    note_id
    granted_by
    granted_to
    role
    can_share
    granted_at
    updated_at
  }
  FAVORITE {
    user_id
    note_id
    created_at
  }
  TAG {
    id
    name
    created_by
    updated_at
    created_at
  }
  NOTE_TAG {
    note_id
    tag_id
    created_at
  }
  USER ||--o{ NOTE : owns
  NOTE |o--o{ NOTE : parent_of
  NOTE ||--o{ BLOCK : contains
  BLOCK ||--o{ BLOCK_TEXT_SPAN : has
  BLOCK ||--o| BLOCK_CODE : opts
  FILE |o--|| ATTACHMENT : used_by
  BLOCK ||--o| ATTACHMENT : embeds
  USER ||--o{ NOTE_PERMISSION : granted_to
  USER ||--o{ NOTE_PERMISSION : granted_by
  NOTE ||--o{ NOTE_PERMISSION : shared_note
  USER ||--o{ FAVORITE : stars
  NOTE ||--o{ FAVORITE : starred
  TAG ||--o{ NOTE_TAG : link
  NOTE ||--o{ NOTE_TAG : link
  FILE |o--o| USER : avatar
  FILE |o--o| NOTE : icon