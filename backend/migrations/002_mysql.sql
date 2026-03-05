-- MySQL Migration Script
-- For production deployment with MySQL/MariaDB

CREATE TABLE IF NOT EXISTS users (
    id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    created_at DATETIME(3) DEFAULT CURRENT_TIMESTAMP(3),
    updated_at DATETIME(3) DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    deleted_at DATETIME(3),
    username VARCHAR(255) NOT NULL UNIQUE,
    password VARCHAR(255) NOT NULL,
    failed_login_attempts INT DEFAULT 0,
    locked_until DATETIME(3),
    last_login_at DATETIME(3),
    INDEX idx_users_username (username),
    INDEX idx_users_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS external_models (
    id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    created_at DATETIME(3) DEFAULT CURRENT_TIMESTAMP(3),
    updated_at DATETIME(3) DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    deleted_at DATETIME(3),
    name VARCHAR(255) NOT NULL,
    model VARCHAR(255) NOT NULL,
    strategy VARCHAR(50) DEFAULT 'round_robin',
    INDEX idx_external_models_model (model),
    INDEX idx_external_models_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS request_sources (
    id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    created_at DATETIME(3) DEFAULT CURRENT_TIMESTAMP(3),
    updated_at DATETIME(3) DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    deleted_at DATETIME(3),
    external_model_id INT UNSIGNED NOT NULL,
    name VARCHAR(255) NOT NULL,
    api_url VARCHAR(2048) NOT NULL,
    api_key VARCHAR(1024) NOT NULL,
    model_name VARCHAR(255) NOT NULL,
    billing_mode VARCHAR(50) DEFAULT 'count',
    limit_count BIGINT DEFAULT 0,
    limit_tokens BIGINT DEFAULT 0,
    limit_reset_interval BIGINT DEFAULT 0,
    limit_reset_time VARCHAR(50) DEFAULT '',
    last_reset_at DATETIME(3),
    used_count BIGINT DEFAULT 0,
    used_tokens BIGINT DEFAULT 0,
    is_active TINYINT(1) DEFAULT 1,
    INDEX idx_request_sources_external_model_id (external_model_id),
    INDEX idx_request_sources_is_active (is_active),
    INDEX idx_request_sources_deleted_at (deleted_at),
    FOREIGN KEY (external_model_id) REFERENCES external_models(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS api_keys (
    id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    created_at DATETIME(3) DEFAULT CURRENT_TIMESTAMP(3),
    updated_at DATETIME(3) DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    deleted_at DATETIME(3),
    key VARCHAR(512) NOT NULL UNIQUE,
    note TEXT,
    external_model_id INT UNSIGNED DEFAULT 0,
    used_count BIGINT DEFAULT 0,
    used_tokens BIGINT DEFAULT 0,
    input_tokens BIGINT DEFAULT 0,
    output_tokens BIGINT DEFAULT 0,
    INDEX idx_api_keys_key (key),
    INDEX idx_api_keys_external_model_id (external_model_id),
    INDEX idx_api_keys_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS usage_records (
    id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    created_at DATETIME(3) DEFAULT CURRENT_TIMESTAMP(3),
    api_key_id INT UNSIGNED NOT NULL,
    external_model_id INT UNSIGNED NOT NULL,
    source_id INT UNSIGNED NOT NULL,
    model VARCHAR(255) NOT NULL,
    input_tokens BIGINT DEFAULT 0,
    output_tokens BIGINT DEFAULT 0,
    success TINYINT(1) DEFAULT 1,
    INDEX idx_usage_records_api_key_id (api_key_id),
    INDEX idx_usage_records_external_model_id (external_model_id),
    INDEX idx_usage_records_source_id (source_id),
    INDEX idx_usage_records_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Insert default admin user (password: admin123)
INSERT INTO users (username, password)
SELECT 'admin', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy'
WHERE NOT EXISTS (SELECT 1 FROM users WHERE username = 'admin');
