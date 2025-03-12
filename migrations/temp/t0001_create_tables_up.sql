CREATE TABLE temp_actions (
    id SERIAL PRIMARY KEY,
    target_id INT NOT NULL,
    action_type_id INT NOT NULL,
    timestamp TIMESTAMP NOT NULL,
    FOREIGN KEY (target_id) REFERENCES target(id),
    FOREIGN KEY (action_type_id) REFERENCES action_types(id)
);
