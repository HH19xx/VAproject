-- スキーマを作成
CREATE SCHEMA IF NOT EXISTS master;
CREATE SCHEMA IF NOT EXISTS base;
CREATE SCHEMA IF NOT EXISTS temp;
CREATE SCHEMA IF NOT EXISTS log;

-- baseスキーマにadminテーブルを作成
CREATE TABLE IF NOT EXISTS base.admin (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- baseスキーマにclientテーブルを作成
CREATE TABLE IF NOT EXISTS base.client (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- baseスキーマにtargetテーブルを作成
CREATE TABLE IF NOT EXISTS base.target (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- logスキーマにaction_logsテーブルを作成
CREATE TABLE IF NOT EXISTS log.action_logs (
    id SERIAL PRIMARY KEY,
    target_id INT,
    action_type INT,
    timestamp TIMESTAMP NOT NULL,
    notes TEXT
);

-- masterスキーマにaction_typesテーブルを作成
CREATE TABLE IF NOT EXISTS master.action_types (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT
);

-- masterスキーマにsettingsテーブルを作成
CREATE TABLE IF NOT EXISTS master.settings (
    id SERIAL PRIMARY KEY,
    key VARCHAR(255) NOT NULL,
    value TEXT
);
