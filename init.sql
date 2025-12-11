CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    firstname VARCHAR(255) NOT NULL,
    lastname VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

ALTER TABLE users ADD COLUMN IF NOT EXISTS two_fa_enabled BOOLEAN DEFAULT false;

CREATE TABLE IF NOT EXISTS two_fa_codes (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    code VARCHAR(6) NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    attempts INTEGER DEFAULT 0,
    is_used BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT valid_code CHECK (code ~ '^[0-9]{6}$')
);

CREATE TABLE IF NOT EXISTS diplomas (
    id SERIAL PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    description TEXT
);

CREATE TABLE IF NOT EXISTS refresh_token (
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token TEXT NOT NULL,
    expires_at TIMESTAMP NOT NULL
);

CREATE TABLE IF NOT EXISTS login_attempts (
    email VARCHAR(255) NOT NULL,
    result BOOLEAN NOT NULL,
    attempt_time TIMESTAMP NOT NULL,
    blocked_until TIMESTAMP
);

DELETE FROM two_fa_codes WHERE expires_at < NOW() - INTERVAL '1 hour';
DELETE FROM login_attempts WHERE attempt_time < NOW() - INTERVAL '24 hours';
DELETE FROM refresh_token WHERE expires_at < NOW();

INSERT INTO users (firstname, lastname, email, password_hash, two_fa_enabled) VALUES
('Иван', 'Иванов', 'ivan@example.com', '$2a$10$WoBnb8Ao2ah5somIbd4a5ukKglisIpp1QQ/g7oByqbQBFwGSECS36', true),
('Петр', 'Петров', 'petr@example.com', '$2a$10$WoBnb8Ao2ah5somIbd4a5ukKglisIpp1QQ/g7oByqbQBFwGSECS36', false),
('Мария', 'Сидорова', 'maria@example.com', '$2a$10$WoBnb8Ao2ah5somIbd4a5ukKglisIpp1QQ/g7oByqbQBFwGSECS36', false)
ON CONFLICT (email) DO UPDATE SET
    two_fa_enabled = EXCLUDED.two_fa_enabled;

INSERT INTO diplomas (title, description) VALUES
('Диплом по веб-разработке', 'Исследование современных фреймворков для веб-разработки'),
('Диплом по машинному обучению', 'Применение нейронных сетей для анализа изображений'),
('Диплом по базам данных', 'Оптимизация запросов в распределенных системах'),
('Диплом по кибербезопасности', 'Методы защиты от SQL-инъекций'),
('Диплом по мобильной разработке', 'Сравнение кроссплатформенных решений')
ON CONFLICT DO NOTHING;

CREATE INDEX IF NOT EXISTS idx_two_fa_codes_user_id ON two_fa_codes(user_id);
CREATE INDEX IF NOT EXISTS idx_two_fa_codes_expires ON two_fa_codes(expires_at);
CREATE INDEX IF NOT EXISTS idx_two_fa_codes_code ON two_fa_codes(code);
CREATE INDEX IF NOT EXISTS idx_two_fa_codes_used ON two_fa_codes(is_used) WHERE is_used = false;

CREATE INDEX IF NOT EXISTS idx_two_fa_codes_user_created ON two_fa_codes(user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_two_fa_codes_created_at ON two_fa_codes(created_at);
CREATE INDEX IF NOT EXISTS idx_users_two_fa_enabled ON users(two_fa_enabled) WHERE two_fa_enabled = true;

SELECT setval('users_id_seq', (SELECT COALESCE(MAX(id), 1) FROM users));
SELECT setval('diplomas_id_seq', (SELECT COALESCE(MAX(id), 1) FROM diplomas));
SELECT setval('two_fa_codes_id_seq', (SELECT COALESCE(MAX(id), 1) FROM two_fa_codes));

DO $$ BEGIN
    RAISE NOTICE 'Init.sql executed successfully!';
    RAISE NOTICE 'Two-factor authentication support added';
    RAISE NOTICE 'Users with 2FA enabled: Ivan, Maria';
    RAISE NOTICE 'User with 2FA disabled: Petr';
END $$;