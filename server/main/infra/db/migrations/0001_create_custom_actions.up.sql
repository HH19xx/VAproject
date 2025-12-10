-- golang-migrateでは<version>_<name>.<direction>.sqlで命名する。
CREATE TABLE IF NOT EXISTS custom_actions (
    id SERIAL PRIMARY KEY,
    user_id INT REFERENCES users(id),
    target_id INT REFERENCES targets(id),
    action_type_id INT REFERENCES action_types(id),
    notes TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    create_user INT REFERENCES users(id),
    updated_at TIMESTAMP NOT NULL,
    update_user INT REFERENCES users(id),
    deleted_at TIMESTAMP DEFAULT NULL
);
