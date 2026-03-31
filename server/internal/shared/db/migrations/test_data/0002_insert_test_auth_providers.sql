-- テストユーザーの認証プロバイダー情報を挿入
-- このファイルは開発/テスト環境でのみ実行され、本番環境では実行されない

-- testuser1: local認証のみ
INSERT INTO user_auth_providers (user_id, provider, provider_user_id, created_at, create_user)
SELECT u.id, 'local', NULL, CURRENT_TIMESTAMP, 'system'
FROM users u
WHERE u.username = 'testuser1'
AND NOT EXISTS (
    SELECT 1 FROM user_auth_providers uap 
    WHERE uap.user_id = u.id AND uap.provider = 'local'
);

-- developer: local認証のみ
INSERT INTO user_auth_providers (user_id, provider, provider_user_id, created_at, create_user)
SELECT u.id, 'local', NULL, CURRENT_TIMESTAMP, 'system'
FROM users u
WHERE u.username = 'developer'
AND NOT EXISTS (
    SELECT 1 FROM user_auth_providers uap 
    WHERE uap.user_id = u.id AND uap.provider = 'local'
);

-- tester: local認証のみ
INSERT INTO user_auth_providers (user_id, provider, provider_user_id, created_at, create_user)
SELECT u.id, 'local', NULL, CURRENT_TIMESTAMP, 'system'
FROM users u
WHERE u.username = 'tester'
AND NOT EXISTS (
    SELECT 1 FROM user_auth_providers uap 
    WHERE uap.user_id = u.id AND uap.provider = 'local'
);
