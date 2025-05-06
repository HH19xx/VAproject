-- 初期ユーザーを削除
DELETE FROM users WHERE name = 'admin' AND password = 'admin123';

-- ユーザーテーブルを削除
DROP TABLE IF EXISTS users;
