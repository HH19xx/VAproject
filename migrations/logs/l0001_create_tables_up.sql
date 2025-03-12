CREATE TABLE action_logs (
    id SERIAL PRIMARY KEY,
    target_id INT NOT NULL,
    action_type_id INT NOT NULL,
    timestamp TIMESTAMP NOT NULL,
    details TEXT,
    FOREIGN KEY (target_id) REFERENCES target(id),
    FOREIGN KEY (action_type_id) REFERENCES action_types(id)
);
