-- golang-migrateでは<version>_<name>.<direction>.sqlで命名する。
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    name VARCHAR NOT NULL,
--   email VARCHAR UNIQUE NOT NULL,
--   仮のパスワード
    password VARCHAR(20) NOT NULL,
--   本番用のパスワード
--   password_hash TEXT NOT NULL,
    notes TEXT,
    updated_at TIMESTAMP NOT NULL,
    update_user INT REFERENCES users(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    create_user INT REFERENCES users(id)
);
