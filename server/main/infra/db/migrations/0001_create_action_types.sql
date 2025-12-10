
CREATE TABLE IF NOT EXISTS action_types (
    id SERIAL PRIMARY KEY,
    action_name VARCHAR(64) NOT NULL,
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    create_user VARCHAR(32) NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    update_user VARCHAR(32) NOT NULL,
    deleted_at TIMESTAMP DEFAULT NULL
);

-- 論理削除されていない行動種別のみユニーク制約を適用
CREATE UNIQUE INDEX IF NOT EXISTS unique_active_action_name ON action_types(action_name) WHERE deleted_at IS NULL;
