CREATE TABLE IF NOT EXISTS settings (
    id SERIAL PRIMARY KEY,
    key VARCHAR(64) NOT NULL,
    value VARCHAR(256) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    create_user VARCHAR(32) NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    update_user VARCHAR(32) NOT NULL,
    deleted_at TIMESTAMP DEFAULT NULL
);

-- 論理削除されていない設定のみユニーク制約を適用
CREATE UNIQUE INDEX IF NOT EXISTS unique_active_setting_key ON settings(key) WHERE deleted_at IS NULL;
