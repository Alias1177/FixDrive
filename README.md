# FixDrive - Сервис авторизации и OTP

Микросервис для авторизации клиентов и водителей с поддержкой OTP верификации.

## 🚀 Возможности

- JWT авторизация для клиентов и водителей
- Refresh токены с возможностью отзыва
- OTP верификация через SMS (Twilio)
- Хранение сессий в Redis
- PostgreSQL для основных данных
- Docker поддержка
- Kubernetes готовность

## 📋 API Endpoints

### 🔐 Client Auth API (`/auth/client/`)

#### POST `/auth/client/register`
Регистрация нового клиента

**Запрос:**
```json
{
  "email": "user@example.com",
  "password": "password123",
  "phone_number": "+1234567890",
  "first_name": "Иван",
  "last_name": "Иванов"
}
```

**Ответ:**
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "refresh_token_value",
  "token_type": "Bearer",
  "expires_in": 900,
  "expires_at": "2024-01-01T15:30:00Z",
  "user_info": {
    "id": 1,
    "email": "user@example.com",
    "phone_number": "+1234567890",
    "first_name": "Иван",
    "last_name": "Иванов",
    "status": "active"
  }
}
```

#### POST `/auth/client/login`
Авторизация клиента

**Запрос:**
```json
{
  "email": "user@example.com",
  "password": "password123"
}
```

**Ответ:** Аналогично `/register`

#### POST `/auth/client/refresh`
Обновление access токена

**Запрос:**
```json
{
  "refresh_token": "refresh_token_value"
}
```

**Ответ:** Новая пара токенов

#### POST `/auth/client/logout`
Выход из аккаунта

**Запрос:**
```json
{
  "refresh_token": "refresh_token_value"
}
```

**Ответ:**
```json
{
  "message": "Logged out successfully"
}
```

#### POST `/auth/client/logout-all`
Выход из всех устройств

**Заголовки:**
```
Authorization: Bearer <access_token>
```

**Ответ:**
```json
{
  "message": "Logged out from all devices"
}
```

#### GET `/auth/client/me`
Получение профиля пользователя

**Заголовки:**
```
Authorization: Bearer <access_token>
```

**Ответ:**
```json
{
  "id": 1,
  "email": "user@example.com",
  "phone_number": "+1234567890",
  "first_name": "Иван",
  "last_name": "Иванов",
  "status": "active"
}
```

#### GET `/auth/client/list`
Получение списка всех клиентов

**Ответ:** Массив объектов UserInfo

### 🚗 Driver Auth API (`/auth/driver/`)

#### POST `/auth/driver/register`
Регистрация нового водителя

**Запрос:**
```json
{
  "email": "driver@example.com",
  "password": "password123",
  "phone_number": "+1234567890",
  "first_name": "Петр",
  "last_name": "Петров",
  "license_number": "ABC123456",
  "license_expiry_date": "2025-12-31",
  "vehicle_brand": "Toyota",
  "vehicle_model": "Camry",
  "vehicle_number": "А123БВ777",
  "vehicle_year": 2020
}
```

**Ответ:**
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "refresh_token_value",
  "token_type": "Bearer",
  "expires_in": 900,
  "expires_at": "2024-01-01T15:30:00Z",
  "driver_info": {
    "id": 1,
    "email": "driver@example.com",
    "phone_number": "+1234567890",
    "first_name": "Петр",
    "last_name": "Петров",
    "license_number": "ABC123456",
    "license_expiry_date": "2025-12-31",
    "vehicle_brand": "Toyota",
    "vehicle_model": "Camry",
    "vehicle_number": "А123БВ777",
    "vehicle_year": 2020,
    "status": "active",
    "rating": 0.0
  }
}
```

#### POST `/auth/driver/login`
Авторизация водителя

**Запрос:**
```json
{
  "email": "driver@example.com",
  "password": "password123"
}
```

**Ответ:** Аналогично `/register`

#### POST `/auth/driver/refresh`
Обновление access токена

**Запрос:**
```json
{
  "refresh_token": "refresh_token_value"
}
```

#### POST `/auth/driver/logout`
Выход из аккаунта

**Запрос:**
```json
{
  "refresh_token": "refresh_token_value"
}
```

#### POST `/auth/driver/logout-all`
Выход из всех устройств

**Заголовки:**
```
Authorization: Bearer <access_token>
```

#### GET `/auth/driver/me`
Получение профиля водителя

**Заголовки:**
```
Authorization: Bearer <access_token>
```

#### GET `/auth/driver/list`
Получение списка всех водителей

**Ответ:** Массив объектов DriverInfo

### 📱 OTP API (`/otp/`)

#### POST `/otp/send`
Отправка OTP кода на телефон

**Запрос:**
```json
{
  "phone": "+994516995513"
}
```

**Ответ:**
```json
{
  "success": true,
  "message": "OTP код отправлен"
}
```

#### POST `/otp/verify`
Проверка OTP кода

**Запрос:**
```json
{
  "phone": "+994516995513",
  "code": "123456"
}
```

**Ответ при успехе:**
```json
{
  "success": true,
  "message": "Код успешно верифицирован"
}
```

**Ответ при ошибке:**
```json
{
  "success": false,
  "message": "Неверный или истекший код"
}
```

### 🏥 Health Check

#### GET `/health`
Проверка состояния сервиса

**Ответ:**
```
OK
```

## 🔧 Конфигурация

### Environment Variables

```bash
# Сервер
SERVER_PORT=8080

# База данных PostgreSQL
DB_HOST=localhost
DB_PORT=5442
DB_USER=user
DB_PASSWORD=password
DB_NAME=db

# Redis
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0

# JWT
JWT_SECRET=your_super_secret_key

# Twilio для SMS
TWILIO_ACCOUNT_SID=your_twilio_account_sid
TWILIO_AUTH_TOKEN=your_auth_token
TWILIO_FROM_PHONE=+1234567890
```

## 🚀 Запуск

### С Docker Compose

```bash
# Запуск всех сервисов
docker-compose up -d

# Просмотр логов
docker-compose logs -f

# Остановка
docker-compose down
```

### Локально

```bash
# Установка зависимостей
go mod tidy

# Миграции БД
make migrate-up

# Запуск приложения
go run cmd/main.go
```

### Kubernetes

```bash
# Применение манифестов
kubectl apply -f k8s/

# Проверка статуса
kubectl get pods -n fixdrive
```

## 📁 Структура проекта

```
FixDrive/
├── cmd/                    # Точки входа
│   └── main.go
├── config/                 # Конфигурация
│   └── config.go
├── internal/               # Внутренняя логика
│   ├── client/auth/        # Авторизация клиентов
│   ├── driver/auth/        # Авторизация водителей
│   └── otp/               # OTP сервис
├── models/                 # Модели данных
├── repo/                   # Репозитории
├── migrations/             # Миграции БД
├── k8s/                    # Kubernetes манифесты
└── scripts/               # Скрипты
```

## 🔐 Безопасность

- JWT токены с коротким временем жизни (15 минут)
- Refresh токены с возможностью отзыва
- Пароли хешируются с помощью bcrypt
- OTP коды хранятся в Redis с TTL 5 минут
- Валидация входных данных
- CORS и другие security middleware

## 🧪 Тестирование

```bash
# Запуск всех тестов
go test ./...

# Тесты с покрытием
go test -cover ./...

# Тесты для конкретного модуля
go test ./internal/client/auth/...
```

## 📊 Мониторинг

- Health check endpoint: `/health`
- Логирование запросов
- Metrics готовность (можно добавить Prometheus)

## 🤝 Разработка

1. Форкнуть репозиторий
2. Создать feature branch
3. Сделать изменения
4. Добавить тесты
5. Создать Pull Request

## 📝 Лицензия

MIT License 