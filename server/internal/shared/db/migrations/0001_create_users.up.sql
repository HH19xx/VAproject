-- golang-migrateでは<version>_<name>.<direction>.sqlで命名する。

CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(32) NOT NULL,
    email VARCHAR(128) NOT NULL,
    password_hash VARCHAR(128) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    create_user VARCHAR(32) NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    update_user VARCHAR(32) NOT NULL,
    deleted_at TIMESTAMP DEFAULT NULL
);

-- 論理削除されていないユーザーのみユニーク制約を適用（削除済みユーザーとの重複を許可）
CREATE UNIQUE INDEX IF NOT EXISTS unique_active_username ON users(username) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX IF NOT EXISTS unique_active_email ON users(email) WHERE deleted_at IS NULL;


-- 初期管理者ユーザー（パスワードはハッシュ化して登録することを推奨）
-- 既にデータが存在する場合はスキップする
INSERT INTO users (
    username,
    email,
    password_hash,
    created_at,
    create_user,
    updated_at,
    update_user
)
SELECT
    'admin',
    'admin@example.com',
    '$2a$10$dummyhashforadmin',
    CURRENT_TIMESTAMP,
    'system',
    CURRENT_TIMESTAMP,
    'system'
WHERE NOT EXISTS (
    SELECT 1 FROM users WHERE username = 'admin'
);
