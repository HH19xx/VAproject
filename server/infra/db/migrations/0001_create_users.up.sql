-- golang-migrateでは<version>_<name>.<direction>.sqlで命名する。
-- ユーザーテーブルの作成
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    name VARCHAR NOT NULL,
--   email VARCHAR UNIQUE NOT NULL,
--   仮の平文のパスワード
    password VARCHAR(20) NOT NULL,
--   本番用のパスワード
--   password_hash TEXT NOT NULL,
    notes TEXT,
    updated_at TIMESTAMP NOT NULL,
    update_user INT REFERENCES users(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    create_user INT REFERENCES users(id)
);

-- 開発用の初期ユーザー admin を追加（パスワードは仮に admin123）
INSERT INTO users (
    name,
    password,
    notes,
    updated_at,
    update_user,
    created_at,
    create_user
) VALUES (
    'admin',
    'admin123',
    '開発用の初期管理者ユーザー（平文パスワード）',
    CURRENT_TIMESTAMP,
    NULL,
    CURRENT_TIMESTAMP,
    NULL
);
