-- AND検索に必要なタグ基盤を初回作成時に直接作る

CREATE TABLE IF NOT EXISTS tags (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL,
    name VARCHAR(128) NOT NULL,
    group_name VARCHAR(64),
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    create_user VARCHAR(32) NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    update_user VARCHAR(32) NOT NULL,
    deleted_at TIMESTAMP DEFAULT NULL,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE UNIQUE INDEX IF NOT EXISTS unique_active_tag_name_per_user
ON tags(user_id, name)
WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_tags_user_active
ON tags(user_id)
WHERE deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS target_tags (
    id SERIAL PRIMARY KEY,
    target_id INT NOT NULL,
    tag_id INT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    create_user VARCHAR(32) NOT NULL,
    FOREIGN KEY (target_id) REFERENCES targets(id) ON DELETE CASCADE,
    FOREIGN KEY (tag_id) REFERENCES tags(id) ON DELETE CASCADE
);

CREATE UNIQUE INDEX IF NOT EXISTS unique_target_tag_pair
ON target_tags(target_id, tag_id);

CREATE INDEX IF NOT EXISTS idx_target_tags_target_id
ON target_tags(target_id);

CREATE INDEX IF NOT EXISTS idx_target_tags_tag_id
ON target_tags(tag_id);
