CREATE TABLE IF NOT EXISTS error_logs (
    id SERIAL PRIMARY KEY,
    error_message TEXT NOT NULL,
    error_code VARCHAR(32),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    create_user VARCHAR(32) NOT NULL,
    stack_trace TEXT,
    user_id INT,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    update_user VARCHAR(32) NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users(id)
);
CREATE INDEX IF NOT EXISTS IDX_error_logs_user ON error_logs(user_id);
