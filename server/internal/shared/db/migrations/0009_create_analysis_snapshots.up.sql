CREATE TABLE IF NOT EXISTS analysis_snapshots (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL,
    query_text TEXT NOT NULL DEFAULT '',
    severity VARCHAR(16) NOT NULL,
    score INT NOT NULL,
    delta_avg_tag DOUBLE PRECISION NOT NULL,
    delta_var_tag DOUBLE PRECISION NOT NULL,
    delta_prototype_rate DOUBLE PRECISION NOT NULL,
    p_value DOUBLE PRECISION,
    significant BOOLEAN,
    current_count INT NOT NULL,
    current_avg_tag DOUBLE PRECISION NOT NULL,
    current_var_tag DOUBLE PRECISION NOT NULL,
    current_prototype_rate DOUBLE PRECISION NOT NULL,
    baseline_count INT,
    baseline_avg_tag DOUBLE PRECISION,
    baseline_var_tag DOUBLE PRECISION,
    baseline_prototype_rate DOUBLE PRECISION,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    create_user VARCHAR(32) NOT NULL,
    deleted_at TIMESTAMP DEFAULT NULL,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_analysis_snapshots_user_created
ON analysis_snapshots(user_id, created_at DESC)
WHERE deleted_at IS NULL;
