CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE users (
    id            uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    login         text NOT NULL,
    password_hash text NOT NULL,
    created_at    timestamptz NOT NULL DEFAULT now(),
    updated_at    timestamptz NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX users_login_key ON users (lower(login));

CREATE TABLE sessions (
    id         uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    uuid NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    token_hash text NOT NULL UNIQUE,
    expires_at timestamptz NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX sessions_user_id_idx ON sessions (user_id);
CREATE INDEX sessions_expires_at_idx ON sessions (expires_at);

CREATE TABLE documents (
    id         uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    title      text NOT NULL,
    categories text[] NOT NULL DEFAULT '{}',
    body       text NOT NULL DEFAULT '',
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX documents_categories_idx ON documents USING gin (categories);
CREATE INDEX documents_updated_at_idx ON documents (updated_at DESC);

-- Снимок склеенной базы знаний, привязанный к чатам.
CREATE TABLE kb_snapshots (
    id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    content         text NOT NULL,
    content_hash    text NOT NULL UNIQUE,
    documents_count integer NOT NULL DEFAULT 0,
    chars_count     integer NOT NULL DEFAULT 0,
    is_active       boolean NOT NULL DEFAULT false,
    created_at      timestamptz NOT NULL DEFAULT now()
);

-- Активный снимок всегда ровно один.
CREATE UNIQUE INDEX kb_snapshots_single_active_idx ON kb_snapshots (is_active) WHERE is_active;

CREATE TABLE chats (
    id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         uuid NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    kb_snapshot_id  uuid REFERENCES kb_snapshots (id) ON DELETE SET NULL,
    title           text NOT NULL DEFAULT 'Новый чат',
    model           text NOT NULL DEFAULT '',
    created_at      timestamptz NOT NULL DEFAULT now(),
    updated_at      timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX chats_user_id_updated_at_idx ON chats (user_id, updated_at DESC);

CREATE TABLE messages (
    id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    chat_id     uuid NOT NULL REFERENCES chats (id) ON DELETE CASCADE,
    role        text NOT NULL CHECK (role IN ('user', 'assistant', 'system')),
    content     text NOT NULL DEFAULT '',
    attachments jsonb NOT NULL DEFAULT '[]'::jsonb,
    usage       jsonb,
    error       text,
    created_at  timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX messages_chat_id_created_at_idx ON messages (chat_id, created_at);
