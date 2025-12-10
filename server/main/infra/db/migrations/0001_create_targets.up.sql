-- golang-migrateでは<version>_<name>.<direction>.sqlで命名する。

CREATE TABLE IF NOT EXISTS targets (
    id SERIAL PRIMARY KEY,
    name VARCHAR(64) NOT NULL,
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    create_user VARCHAR(32) NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    update_user VARCHAR(32) NOT NULL,
    deleted_at TIMESTAMP DEFAULT NULL
);

-- 論理削除されていない観察対象のみユニーク制約を適用
CREATE UNIQUE INDEX IF NOT EXISTS unique_active_target_name ON targets(name) WHERE deleted_at IS NULL;
