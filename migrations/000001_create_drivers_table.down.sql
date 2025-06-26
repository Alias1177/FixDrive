-- Удаление индексов
DROP INDEX IF EXISTS idx_client_status;
DROP INDEX IF EXISTS idx_client_email;

-- Удаление таблицы
DROP TABLE IF EXISTS client;