CREATE TABLE IF NOT EXISTS prototypes (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL,
    name VARCHAR(128) NOT NULL,
    description TEXT,
    parent_prototype_id INT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    create_user VARCHAR(32) NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    update_user VARCHAR(32) NOT NULL,
    deleted_at TIMESTAMP DEFAULT NULL,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (parent_prototype_id) REFERENCES prototypes(id) ON DELETE SET NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS unique_active_prototype_name_per_user
ON prototypes(user_id, name)
WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_prototypes_user_active
ON prototypes(user_id)
WHERE deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS prototype_tags (
    id SERIAL PRIMARY KEY,
    prototype_id INT NOT NULL,
    tag_id INT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    create_user VARCHAR(32) NOT NULL,
    FOREIGN KEY (prototype_id) REFERENCES prototypes(id) ON DELETE CASCADE,
    FOREIGN KEY (tag_id) REFERENCES tags(id) ON DELETE CASCADE
);

CREATE UNIQUE INDEX IF NOT EXISTS unique_prototype_tag_pair
ON prototype_tags(prototype_id, tag_id);

CREATE INDEX IF NOT EXISTS idx_prototype_tags_prototype_id
ON prototype_tags(prototype_id);

CREATE INDEX IF NOT EXISTS idx_prototype_tags_tag_id
ON prototype_tags(tag_id);
