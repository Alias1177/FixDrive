-- Создание таблицы водителей
CREATE TABLE client (
                        id SERIAL PRIMARY KEY,
                        email VARCHAR(255) UNIQUE NOT NULL,
                        password_hash VARCHAR(255) NOT NULL,
                        phone_number VARCHAR(20),
                        first_name VARCHAR(100),
                        last_name VARCHAR(100),
                        status VARCHAR(20) DEFAULT 'active',
                        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                        updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Индексы для быстрого поиска
CREATE INDEX idx_client_email ON client(email);
CREATE INDEX idx_client_status ON client(status);