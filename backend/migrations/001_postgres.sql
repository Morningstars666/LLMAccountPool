-- PostgreSQL Migration Script
-- For production deployment with PostgreSQL

-- Create tables
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    username VARCHAR(255) NOT NULL UNIQUE,
    password VARCHAR(255) NOT NULL,
    failed_login_attempts INTEGER DEFAULT 0,
    locked_until TIMESTAMP WITH TIME ZONE,
    last_login_at TIMESTAMP WITH TIME ZONE
);

CREATE TABLE IF NOT EXISTS external_models (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    name VARCHAR(255) NOT NULL,
    model VARCHAR(255) NOT NULL,
    strategy VARCHAR(50) DEFAULT 'round_robin'
);

CREATE TABLE IF NOT EXISTS request_sources (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    external_model_id INTEGER NOT NULL REFERENCES external_models(id),
    name VARCHAR(255) NOT NULL,
    api_url VARCHAR(2048) NOT NULL,
    api_key VARCHAR(1024) NOT NULL,
    model_name VARCHAR(255) NOT NULL,
    billing_mode VARCHAR(50) DEFAULT 'count',
    limit_count BIGINT DEFAULT 0,
    limit_tokens BIGINT DEFAULT 0,
    limit_reset_interval BIGINT DEFAULT 0,
    limit_reset_time VARCHAR(50) DEFAULT '',
    last_reset_at TIMESTAMP WITH TIME ZONE,
    used_count BIGINT DEFAULT 0,
    used_tokens BIGINT DEFAULT 0,
    is_active BOOLEAN DEFAULT true
);

CREATE TABLE IF NOT EXISTS api_keys (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    key VARCHAR(512) NOT NULL UNIQUE,
    note TEXT,
    external_model_id INTEGER DEFAULT 0,
    used_count BIGINT DEFAULT 0,
    used_tokens BIGINT DEFAULT 0,
    input_tokens BIGINT DEFAULT 0,
    output_tokens BIGINT DEFAULT 0
);

CREATE TABLE IF NOT EXISTS usage_records (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    api_key_id INTEGER NOT NULL,
    external_model_id INTEGER NOT NULL,
    source_id INTEGER NOT NULL,
    model VARCHAR(255) NOT NULL,
    input_tokens BIGINT DEFAULT 0,
    output_tokens BIGINT DEFAULT 0,
    success BOOLEAN DEFAULT true
);

-- Create indexes for better query performance
CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
CREATE INDEX IF NOT EXISTS idx_users_deleted_at ON users(deleted_at);
CREATE INDEX IF NOT EXISTS idx_external_models_model ON external_models(model);
CREATE INDEX IF NOT EXISTS idx_external_models_deleted_at ON external_models(deleted_at);
CREATE INDEX IF NOT EXISTS idx_request_sources_external_model_id ON request_sources(external_model_id);
CREATE INDEX IF NOT EXISTS idx_request_sources_is_active ON request_sources(is_active);
CREATE INDEX IF NOT EXISTS idx_request_sources_deleted_at ON request_sources(deleted_at);
CREATE INDEX IF NOT EXISTS idx_api_keys_key ON api_keys(key);
CREATE INDEX IF NOT EXISTS idx_api_keys_external_model_id ON api_keys(external_model_id);
CREATE INDEX IF NOT EXISTS idx_api_keys_deleted_at ON api_keys(deleted_at);
CREATE INDEX IF NOT EXISTS idx_usage_records_api_key_id ON usage_records(api_key_id);
CREATE INDEX IF NOT EXISTS idx_usage_records_external_model_id ON usage_records(external_model_id);
CREATE INDEX IF NOT EXISTS idx_usage_records_source_id ON usage_records(source_id);
CREATE INDEX IF NOT EXISTS idx_usage_records_created_at ON usage_records(created_at);

-- Insert default admin user (password: admin123)
-- Note: Generate the hash using bcrypt
INSERT INTO users (username, password, created_at, updated_at)
SELECT 'admin', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', NOW(), NOW()
WHERE NOT EXISTS (SELECT 1 FROM users WHERE username = 'admin');
