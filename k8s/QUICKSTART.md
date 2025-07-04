# Быстрый старт FixDrive в Kubernetes

## 1. Подготовка

```bash
# Собери образ
docker build -t fixdrive:latest .

# Или через Docker Hub
docker tag fixdrive:latest your-username/fixdrive:latest
docker push your-username/fixdrive:latest
```

## 2. Настройка секретов

✅ **Секреты уже настроены** с твоими значениями из .env файла:
- DB_PASSWORD: your_db_password
- JWT_SECRET: your-super-secret-jwt-key-here
- TWILIO_ACCOUNT_SID: ACxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
- TWILIO_AUTH_TOKEN: your_twilio_auth_token_here
- TWILIO_FROM_PHONE: +1234567890

Если нужно изменить, отредактируй `secret.yaml`

## 3. Обнови образ

В `app.yaml` замени `fixdrive:latest` на свой образ

## 4. Деплой

```bash
cd k8s
./deploy.sh
```

## 5. Проверка

```bash
kubectl get pods -n fixdrive
kubectl get svc -n fixdrive
kubectl logs -f deployment/fixdrive-app -n fixdrive
```

## 6. Доступ

Добавь в `/etc/hosts`:
```
<IP-кластера> fixdrive.local
```

Открой: `http://fixdrive.local`

## Альтернативы

### Через kustomize
```bash
kubectl apply -k k8s/
```

### Через Helm (если есть)
```bash
helm install fixdrive ./helm-chart
```

### Удаление
```bash
kubectl delete namespace fixdrive
```

## Полезные команды

```bash
# Статус
kubectl get all -n fixdrive

# Логи
kubectl logs -f -l app=fixdrive-app -n fixdrive

# Порт-форвардинг для тестирования
kubectl port-forward svc/fixdrive-app 8080:8080 -n fixdrive

# Подключиться к поду
kubectl exec -it deployment/fixdrive-app -n fixdrive -- sh
``` 