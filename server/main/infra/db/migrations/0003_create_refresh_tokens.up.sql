-- リフレッシュトークン管理テーブルを作成する。
-- セキュリティのため、プレーントークンは保存せず、SHA-256等でハッシュ化した値を保存する。
-- 監査カラム（created_at, create_user, updated_at, update_user）を必ず含める。

CREATE TABLE IF NOT EXISTS refresh_tokens (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash TEXT NOT NULL UNIQUE,
    expires_at TIMESTAMP NOT NULL,
    revoked BOOLEAN NOT NULL DEFAULT FALSE,

    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    create_user TEXT NOT NULL,
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    update_user TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_refresh_tokens_user_id ON refresh_tokens(user_id);
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_valid ON refresh_tokens(user_id, revoked, expires_at);

