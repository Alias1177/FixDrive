
-- Создание таблицы водителей
CREATE TABLE drivers (
                         id SERIAL PRIMARY KEY,
                         email VARCHAR(255) UNIQUE NOT NULL,
                         password_hash VARCHAR(255) NOT NULL,
                         phone_number VARCHAR(20),
                         first_name VARCHAR(100),
                         last_name VARCHAR(100),
                         license_number VARCHAR(50) UNIQUE NOT NULL,
                         license_expiry_date DATE NOT NULL,
                         vehicle_brand VARCHAR(100),
                         vehicle_model VARCHAR(100),
                         vehicle_number VARCHAR(20) UNIQUE NOT NULL,
                         vehicle_year INTEGER,
                         status VARCHAR(20) DEFAULT 'pending',
                         rating DECIMAL(3,2) DEFAULT 0.00,
                         created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                         updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Индексы для быстрого поиска
CREATE INDEX idx_drivers_email ON drivers(email);
CREATE INDEX idx_drivers_license_number ON drivers(license_number);
CREATE INDEX idx_drivers_vehicle_number ON drivers(vehicle_number);
CREATE INDEX idx_drivers_status ON drivers(status);