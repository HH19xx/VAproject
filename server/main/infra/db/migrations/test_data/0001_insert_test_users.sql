-- テストユーザーデータ挿入用DML
-- パスワードはすべて "password123" でbcryptハッシュ化済み
-- このファイルは開発/テスト環境でのみ実行され、本番環境では実行されない

-- テストユーザー1: 通常ユーザー
-- username: testuser1, email: testuser1@example.com, password: password123
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
    'testuser1',
    'testuser1@example.com',
    '$2a$10$AKtUau.FKIcBYm/YdwXEY.UvUDH.bMBx1ojEpntrOVJU8uhONxhMi',
    CURRENT_TIMESTAMP,
    'system',
    CURRENT_TIMESTAMP,
    'system'
WHERE NOT EXISTS (
    SELECT 1 FROM users WHERE username = 'testuser1'
);

-- テストユーザー2: 開発用ユーザー
-- username: developer, email: dev@example.com, password: password123
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
    'developer',
    'dev@example.com',
    '$2a$10$AKtUau.FKIcBYm/YdwXEY.UvUDH.bMBx1ojEpntrOVJU8uhONxhMi',
    CURRENT_TIMESTAMP,
    'system',
    CURRENT_TIMESTAMP,
    'system'
WHERE NOT EXISTS (
    SELECT 1 FROM users WHERE username = 'developer'
);

-- テストユーザー3: テスト専用ユーザー
-- username: tester, email: test@vaproject.local, password: password123
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
    'tester',
    'test@vaproject.local',
    '$2a$10$AKtUau.FKIcBYm/YdwXEY.UvUDH.bMBx1ojEpntrOVJU8uhONxhMi',
    CURRENT_TIMESTAMP,
    'system',
    CURRENT_TIMESTAMP,
    'system'
WHERE NOT EXISTS (
    SELECT 1 FROM users WHERE username = 'tester'
);
