-- golang-migrateでは<version>_<name>.<direction>.sqlで命名する。
CREATE TABLE IF NOT EXISTS targets (
    id SERIAL PRIMARY KEY,
    name VARCHAR NOT NULL,
    description TEXT,
    notes TEXT,
    updated_at TIMESTAMP NOT NULL,
    update_user INT REFERENCES users(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    create_user INT REFERENCES users(id)
);
