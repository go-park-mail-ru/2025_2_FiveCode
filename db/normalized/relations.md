# Нормализованные отношения и функциональные зависимости

## Краткие описания таблиц

* **USER** — аккаунты пользователей: учётные записи, логины, имя и привязка аватара.
* **FILE** — загруженные файлы с метаданными (тип, размеры, вес, URL).
* **NOTE** — заметки (дерево родитель→дети), владельцы, заголовки, статус архива, иконки.
* **BLOCK** — атомарные блоки контента внутри заметок (тип, порядок).
* **BLOCK_TEXT** — текстовое содержимое текстовых блоков.
* **BLOCK_TEXT_FORMAT** — диапазоны форматирования текста (жирный, курсив, ссылки, шрифты).
* **BLOCK_CODE** — содержимое кода для код‑блоков (язык, текст кода).
* **BLOCK_ATTACHMENT** — файловые вложения блоков (файл и подпись).
* **NOTE_PERMISSION** — выданные права на заметки (кому, кем, какая роль, шаринг).
* **FAVORITE** — отметки «в избранном» (кто какую заметку добавил).
* **TAG** — теги и автор их создания.
* **NOTE_TAG** — связь заметок с тегами.

> Обозначения: `PK` — первичный ключ; `FK` — внешний ключ; ФЗ — функциональные зависимости в виде `X → Y`.

---

## USER

**PK:** `{id}`  
**FK:** `avatar_file_id → FILE(id)`

**ФЗ (минимальные)**

* `id → email, password_hash, username, avatar_file_id, created_at, updated_at`
* `email → id` *(следовательно, `email → password_hash, username, avatar_file_id, created_at, updated_at`)*

**Нормальные формы**

* **1НФ:** да (атомарность).
* **2НФ:** да (PK не составной).
* **3НФ:** да (нет транзитивных зависимостей от ключа).
* **НФБК:** да (все детерминанты — суперклавиши `id`, `email`).

```mermaid
erDiagram
    USERS {
        INTEGER id PK
        TEXT email
        TEXT password_hash
        TEXT username
        INTEGER avatar_file_id FK
        TIMESTAMPTZ created_at
        TIMESTAMPTZ updated_at
    }
```

---

## FILE

**PK:** `{id}`

**ФЗ**

* `id → url, mime_type, size_bytes, width, height, created_at, updated_at`

**Нормальные формы:** 1НФ, 2НФ, 3НФ, **НФБК** — да.

```mermaid
erDiagram
    FILES {
        INTEGER id PK
        TEXT url
        TEXT mime_type
        INTEGER size_bytes
        INTEGER width
        INTEGER height
        TIMESTAMPTZ created_at
        TIMESTAMPTZ updated_at
    }
```

---

## NOTE

**PK:** `{id}`  
**FK:** `owner_id → USER(id)`; `parent_note_id → NOTE(id)`; `icon_file_id → FILE(id)`

**ФЗ**

* `id → owner_id, parent_note_id, title, icon_file_id, is_archived, is_shared, created_at, updated_at, deleted_at`

**Нормальные формы:** 1НФ, 2НФ, 3НФ, **НФБК** — да.

```mermaid
erDiagram
    NOTES {
        INTEGER id PK
        INTEGER owner_id FK
        INTEGER parent_note_id FK
        TEXT title
        INTEGER icon_file_id FK
        BOOLEAN is_archived
        BOOLEAN is_shared
        TIMESTAMPTZ created_at
        TIMESTAMPTZ updated_at
        TIMESTAMPTZ deleted_at
    }
```

---

## BLOCK

**PK:** `{id}`  
**FK:** `note_id → NOTE(id)`; `last_edited_by → USER(id)`

**ФЗ**

* `id → note_id, type, position, created_at, updated_at, last_edited_by`
* `(note_id, position) → id, type, created_at, updated_at, last_edited_by`

**Нормальные формы:** 1НФ, 2НФ, 3НФ, **НФБК** — да.

```mermaid
erDiagram
    BLOCKS {
        INTEGER id PK
        INTEGER note_id FK
        TEXT type
        NUMERIC position
        TIMESTAMPTZ created_at
        TIMESTAMPTZ updated_at
        INTEGER last_edited_by FK
    }
```

---

## BLOCK_TEXT

**PK:** `{id}`  
**FK:** `block_id → BLOCK(id)`

**ФЗ**

* `id → block_id, text, created_at, updated_at`
* `block_id → id, text, created_at, updated_at` *(уникальность `block_id`)*

**Нормальные формы**

* **1НФ:** да (атомарность).
* **2НФ:** да (PK не составной).
* **3НФ:** да (нет транзитивных зависимостей).
* **НФБК:** да (детерминанты — суперклавиши `id`, `block_id`).

```mermaid
erDiagram
    BLOCK_TEXT {
        INTEGER id PK
        INTEGER block_id FK
        TEXT text
        TIMESTAMPTZ created_at
        TIMESTAMPTZ updated_at
    }
```

---

## BLOCK_TEXT_FORMAT

**PK:** `{id}`  
**FK:** `block_text_id → BLOCK_TEXT(id)`

**ФЗ**

*
`id → block_text_id, start_offset, end_offset, bold, italic, underline, strikethrough, link, font, size, created_at, updated_at`

**Нормальные формы**

* **1НФ:** да (атомарность).
* **2НФ:** да (PK не составной).
* **3НФ:** да (нет транзитивных зависимостей от ключа).
* **НФБК:** да (единственный детерминант — ключ `id`).

**Примечание:** Диапазоны форматирования могут накладываться друг на друга, что позволяет применять множественное
форматирование к одному участку текста.

```mermaid
erDiagram
    BLOCK_TEXT_FORMAT {
        INTEGER id PK
        INTEGER block_text_id FK
        INTEGER start_offset
        INTEGER end_offset
        BOOLEAN bold
        BOOLEAN italic
        BOOLEAN underline
        BOOLEAN strikethrough
        TEXT link
        TEXT font
        INTEGER size
        TIMESTAMPTZ created_at
        TIMESTAMPTZ updated_at
    }
```

---

## BLOCK_CODE

**PK:** `{block_id}`  
**FK:** `block_id → BLOCK(id)`

**ФЗ**

* `block_id → language, code_text, created_at, updated_at`

**Нормальные формы:** 1НФ, 2НФ, 3НФ, **НФБК** — да.

```mermaid
erDiagram
    BLOCK_CODE {
        INTEGER block_id PK, FK
        TEXT language
        TEXT code_text
        TIMESTAMPTZ created_at
        TIMESTAMPTZ updated_at
    }
```

---

## BLOCK_ATTACHMENT

**PK:** `{id}`  
**FK:** `block_id → BLOCK(id)`; `file_id → FILE(id)`

**ФЗ**

* `id → block_id, file_id, caption, created_at, updated_at`

**Нормальные формы**

* **1НФ:** да.
* **2НФ:** да (PK не составной).
* **3НФ:** да (нет транзитивных зависимостей от `id`).
* **НФБК:** да (детерминант — только ключ `id`).

```mermaid
erDiagram
    BLOCK_ATTACHMENTS {
        INTEGER id PK
        INTEGER block_id FK
        INTEGER file_id FK
        TEXT caption
        TIMESTAMPTZ created_at
        TIMESTAMPTZ updated_at
    }
```

---

## NOTE_PERMISSION

**PK:** `{note_permission_id}`  
**FK:** `note_id → NOTE(id)`; `granted_by → USER(id)`; `granted_to → USER(id)`

**ФЗ**

* `note_permission_id → note_id, granted_by, granted_to, role, can_share, created_at, updated_at`

**Нормальные формы**

* **1НФ:** да.
* **2НФ:** да (все неключевые зависят от всего ключа).
* **3НФ:** да.
* **НФБК:** да.

```mermaid
erDiagram
    NOTE_PERMISSIONS {
        INTEGER note_permission_id PK
        INTEGER note_id FK
        INTEGER granted_by FK
        INTEGER granted_to FK
        TEXT role
        BOOLEAN can_share
        TIMESTAMPTZ created_at
        TIMESTAMPTZ updated_at
    }
```

---

## FAVORITE

**PK:** `{user_id, note_id}`  
**FK:** `user_id → USER(id)`; `note_id → NOTE(id)`

**ФЗ**

* `(user_id, note_id) → created_at, updated_at`

**Нормальные формы:** 1НФ, 2НФ, 3НФ, **НФБК** — да.

```mermaid
erDiagram
    FAVORITES {
        INTEGER user_id PK, FK
        INTEGER note_id PK, FK
        TIMESTAMPTZ created_at
        TIMESTAMPTZ updated_at
    }
```

---

## TAG

**PK:** `{id}`  
**FK:** `created_by → USER(id)`

**ФЗ**

* `id → name, created_by, created_at, updated_at`
* `name → id` *(следовательно, `name → created_by, created_at, updated_at`)*

**Нормальные формы:** 1НФ, 2НФ, 3НФ, **НФБК** — да (детерминанты — ключи `id`/`name`).

```mermaid
erDiagram
    TAGS {
        INTEGER id PK
        TEXT name
        INTEGER created_by FK
        TIMESTAMPTZ created_at
        TIMESTAMPTZ updated_at
    }
```

---

## NOTE_TAG

**PK:** `{note_id, tag_id}`  
**FK:** `note_id → NOTE(id)`; `tag_id → TAG(id)`

**ФЗ**

* `(note_id, tag_id) → created_at, updated_at`

**Нормальные формы:** 1НФ, 2НФ, 3НФ, **НФБК** — да.

```mermaid
erDiagram
    NOTE_TAGS {
        INTEGER note_id PK, FK
        INTEGER tag_id PK, FK
        TIMESTAMPTZ created_at
        TIMESTAMPTZ updated_at
    }
```

---