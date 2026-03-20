CREATE TABLE IF NOT EXISTS action_logs (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL,
    title VARCHAR(128) NOT NULL,
    prototype_id INT,
    occurred_at TIMESTAMP NOT NULL,
    "timestamp" TIMESTAMP NOT NULL,
    notes TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    create_user VARCHAR(32) NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    update_user VARCHAR(32) NOT NULL,
    deleted_at TIMESTAMP DEFAULT NULL,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (prototype_id) REFERENCES prototypes(id) ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_action_logs_user_occurred_at
ON action_logs(user_id, occurred_at DESC)
WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_action_logs_user_prototype
ON action_logs(user_id, prototype_id)
WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_action_logs_timestamp
ON action_logs("timestamp");

CREATE TABLE IF NOT EXISTS action_log_tags (
    id SERIAL PRIMARY KEY,
    action_log_id INT NOT NULL,
    tag_id INT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    create_user VARCHAR(32) NOT NULL,
    FOREIGN KEY (action_log_id) REFERENCES action_logs(id) ON DELETE CASCADE,
    FOREIGN KEY (tag_id) REFERENCES tags(id) ON DELETE CASCADE
);

CREATE UNIQUE INDEX IF NOT EXISTS unique_action_log_tag_pair
ON action_log_tags(action_log_id, tag_id);

CREATE INDEX IF NOT EXISTS idx_action_log_tags_action_log_id
ON action_log_tags(action_log_id);

CREATE INDEX IF NOT EXISTS idx_action_log_tags_tag_id
ON action_log_tags(tag_id);

CREATE INDEX IF NOT EXISTS idx_action_log_tags_tag_action_log
ON action_log_tags(tag_id, action_log_id);

CREATE TABLE IF NOT EXISTS action_log_attributes (
    id SERIAL PRIMARY KEY,
    action_log_id INT NOT NULL,
    attr_key VARCHAR(64) NOT NULL,
    value_number DOUBLE PRECISION,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    create_user VARCHAR(32) NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    update_user VARCHAR(32) NOT NULL,
    FOREIGN KEY (action_log_id) REFERENCES action_logs(id) ON DELETE CASCADE
);

CREATE UNIQUE INDEX IF NOT EXISTS unique_action_log_attribute_key
ON action_log_attributes(action_log_id, attr_key);

CREATE INDEX IF NOT EXISTS idx_action_log_attributes_action_log_id
ON action_log_attributes(action_log_id);

CREATE INDEX IF NOT EXISTS idx_action_log_attributes_key_value
ON action_log_attributes(attr_key, value_number);
