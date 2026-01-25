CREATE TABLE IF NOT EXISTS target_action_types (
    id SERIAL PRIMARY KEY,
    target_id INT NOT NULL,
    action_type_id INT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    create_user VARCHAR(32) NOT NULL,
    FOREIGN KEY (target_id) REFERENCES targets(id) ON DELETE CASCADE,
    FOREIGN KEY (action_type_id) REFERENCES action_types(id) ON DELETE CASCADE,
    UNIQUE (target_id, action_type_id)
);
CREATE INDEX IF NOT EXISTS IDX_target_action_types_target ON target_action_types(target_id);
CREATE INDEX IF NOT EXISTS IDX_target_action_types_action ON target_action_types(action_type_id);
