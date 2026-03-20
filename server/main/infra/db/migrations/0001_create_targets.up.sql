CREATE TABLE IF NOT EXISTS targets (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL,
    name VARCHAR(64) NOT NULL,
    description TEXT,
    match_mode VARCHAR(10) NOT NULL DEFAULT 'AND',
    query_text TEXT NOT NULL DEFAULT '',
    any_tag_ids INT[] NOT NULL DEFAULT '{}'::INT[],
    any_tag_groups TEXT NOT NULL DEFAULT '',
    exclude_tag_ids INT[] NOT NULL DEFAULT '{}'::INT[],
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    create_user VARCHAR(32) NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    update_user VARCHAR(32) NOT NULL,
    deleted_at TIMESTAMP DEFAULT NULL,
    CONSTRAINT targets_match_mode_check CHECK (match_mode IN ('AND')),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE UNIQUE INDEX IF NOT EXISTS unique_active_target_name_per_user
ON targets(user_id, name)
WHERE deleted_at IS NULL;
