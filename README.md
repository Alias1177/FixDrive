# FixDrive - OTP Integration

## OTP API

### Отправка OTP кода

```bash
POST /otp/send
Content-Type: application/json

{
  "phone": "+994516995513"
}
```

Ответ:
```json
{
  "success": true,
  "message": "OTP код отправлен"
}
```

### Верификация OTP кода

```bash
POST /otp/verify
Content-Type: application/json

{
  "phone": "+994516995513",
  "code": "123456"
}
```

Ответ при успехе:
```json
{
  "success": true,
  "message": "Код успешно верифицирован"
}
```

Ответ при ошибке:
```json
{
  "success": false,
  "message": "Неверный или истекший код"
}
```

## Настройка

### Environment Variables

```bash
# Twilio
TWILIO_ACCOUNT_SID=your_twilio_account_sid
TWILIO_AUTH_TOKEN=your_auth_token

# Redis
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=

# Database
DB_HOST=localhost
DB_PORT=5442
DB_USER=user
DB_PASSWORD=password
DB_NAME=db
```

### Запуск

```bash
# Запуск с Docker Compose
docker-compose up -d

# Установка зависимостей
go mod tidy

# Запуск приложения
go run cmd/main.go
```

## Особенности

- OTP коды хранятся в Redis с TTL 5 минут
- 6-значные коды генерируются криптографически стойким ГПСЧ
- После успешной верификации код удаляется из Redis
- SMS отправляется через Twilio API 