CREATE TABLE IF NOT EXISTS action_logs (
    id SERIAL PRIMARY KEY,
    target_id INT NULL,
    action_type INT NOT NULL,
    timestamp TIMESTAMP NOT NULL,
    notes TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    create_user VARCHAR(32) NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    update_user VARCHAR(32) NOT NULL,
    deleted_at TIMESTAMP DEFAULT NULL,
    FOREIGN KEY (target_id) REFERENCES targets(id) ON DELETE SET NULL,
    FOREIGN KEY (action_type) REFERENCES action_types(id)
);
CREATE INDEX IF NOT EXISTS IDX_action_logs_target ON action_logs(target_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS IDX_action_logs_action_type ON action_logs(action_type) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS IDX_action_logs_timestamp ON action_logs(timestamp);
