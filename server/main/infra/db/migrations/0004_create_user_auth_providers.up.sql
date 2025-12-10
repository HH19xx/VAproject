-- ユーザーの認証プロバイダー情報を管理するテーブル
-- 1ユーザーが複数の認証方法（local, google, github等）を持てる
CREATE TABLE IF NOT EXISTS user_auth_providers (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    provider VARCHAR(20) NOT NULL, -- 'local', 'google', 'github' 等
    provider_user_id VARCHAR(255), -- OAuth側のユーザーID（localの場合はNULL）
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    create_user VARCHAR(32) NOT NULL,
    deleted_at TIMESTAMP DEFAULT NULL
);

-- 論理削除されていない認証プロバイダーのみユニーク制約を適用
CREATE UNIQUE INDEX IF NOT EXISTS unique_active_user_provider ON user_auth_providers(user_id, provider) WHERE deleted_at IS NULL;

-- 検索性能向上のためのインデックス
CREATE INDEX IF NOT EXISTS idx_user_auth_providers_user_id ON user_auth_providers(user_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_user_auth_providers_provider ON user_auth_providers(provider) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_user_auth_providers_provider_user_id ON user_auth_providers(provider, provider_user_id) WHERE deleted_at IS NULL;
