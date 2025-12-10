CREATE TABLE IF NOT EXISTS temp_actions (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL,
    action_type_id INT NOT NULL,
    target_id INT NOT NULL,
    notes TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    create_user VARCHAR(32) NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    update_user VARCHAR(32) NOT NULL,
    deleted_at TIMESTAMP DEFAULT NULL,
    FOREIGN KEY (user_id) REFERENCES users(id),
    FOREIGN KEY (action_type_id) REFERENCES action_types(id),
    FOREIGN KEY (target_id) REFERENCES targets(id)
);
CREATE INDEX IF NOT EXISTS IDX_temp_actions_user ON temp_actions(user_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS IDX_temp_actions_action_type ON temp_actions(action_type_id);
CREATE INDEX IF NOT EXISTS IDX_temp_actions_target ON temp_actions(target_id);
