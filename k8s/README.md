# FixDrive в Kubernetes

## Подготовка

### 1. Собери Docker образ

```bash
# Собери образ
docker build -t fixdrive:latest .

# Загрузи в registry (Docker Hub, Google Container Registry, etc.)
docker tag fixdrive:latest your-registry/fixdrive:latest
docker push your-registry/fixdrive:latest
```

### 2. Настрой секреты

Секреты уже настроены с твоими значениями из .env файла:

```bash
# Уже настроенные значения в base64:
DB_PASSWORD: your_db_password
JWT_SECRET: your-super-secret-jwt-key-here
TWILIO_ACCOUNT_SID: ACxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
TWILIO_AUTH_TOKEN: your_twilio_auth_token_here
TWILIO_FROM_PHONE: +1234567890
```

Если нужно изменить - отредактируй `secret.yaml` и закодируй новые значения:
```bash
echo -n "новое-значение" | base64
```

### 3. Обнови миграции

Скопируй содержимое всех .sql файлов из папки `migrations/` в ConfigMap в файле `migration-job.yaml`.

### 4. Измени образ

В файле `app.yaml` замени `fixdrive:latest` на свой образ.

## Деплой

### Автоматический деплой

```bash
cd k8s
chmod +x deploy.sh
./deploy.sh
```

### Ручной деплой

```bash
# Создай namespace
kubectl apply -f namespace.yaml

# Создай конфигурацию
kubectl apply -f configmap.yaml
kubectl apply -f secret.yaml

# Деплой БД
kubectl apply -f postgres.yaml
kubectl apply -f redis.yaml

# Жди готовности БД
kubectl wait --for=condition=available --timeout=300s deployment/postgres -n fixdrive

# Миграции
kubectl apply -f migration-job.yaml

# Приложение
kubectl apply -f app.yaml

# Ingress
kubectl apply -f ingress.yaml
```

## Проверка

```bash
# Статус подов
kubectl get pods -n fixdrive

# Статус сервисов
kubectl get svc -n fixdrive

# Логи приложения
kubectl logs -f deployment/fixdrive-app -n fixdrive

# Проверка ingress
kubectl get ingress -n fixdrive
```

## Доступ

Добавь в `/etc/hosts`:
```
<IP-адрес-кластера> fixdrive.local
```

Приложение будет доступно на `http://fixdrive.local`

## Эндпоинты

- `/health` - health check
- `/auth/client/*` - аутентификация клиентов
- `/auth/driver/*` - аутентификация водителей
- `/otp/*` - OTP сервис

## Удаление

```bash
kubectl delete namespace fixdrive
```

## Мониторинг

```bash
# Логи всех подов
kubectl logs -f -l app=fixdrive-app -n fixdrive

# Описание пода
kubectl describe pod <pod-name> -n fixdrive

# Подключение к поду
kubectl exec -it <pod-name> -n fixdrive -- sh
``` 