CREATE TABLE driver_refresh_tokens (
                                       id SERIAL PRIMARY KEY,
                                       driver_id INTEGER NOT NULL,
                                       token VARCHAR(255) NOT NULL UNIQUE,
                                       expires_at TIMESTAMP NOT NULL,
                                       created_at TIMESTAMP DEFAULT NOW(),
                                       is_revoked BOOLEAN DEFAULT false,
                                       CONSTRAINT fk_driver_refresh_tokens_driver_id FOREIGN KEY (driver_id) REFERENCES drivers(id) ON DELETE CASCADE
);

CREATE INDEX idx_driver_refresh_tokens_token ON driver_refresh_tokens(token);
CREATE INDEX idx_driver_refresh_tokens_driver_id ON driver_refresh_tokens(driver_id);
CREATE INDEX idx_driver_refresh_tokens_expires_at ON driver_refresh_tokens(expires_at);