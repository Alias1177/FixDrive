# 🚀 Установка Kubernetes на сервер

## 📋 Подготовка

**Твой сервер:** `31.97.76.106`  
**Текущее состояние:** Docker + docker-compose  
**Цель:** Добавить Kubernetes для гибкого деплоя

## ⚠️ Важные моменты

### 1. **Порты и конфликты**
Kubernetes и docker-compose могут работать параллельно, НО:
- **Docker-compose** использует host порты (80, 443, 5432, 6379)
- **Kubernetes** использует NodePort или LoadBalancer

### 2. **Ресурсы сервера**
- **Минимум:** 2 CPU, 4GB RAM
- **Рекомендуется:** 4 CPU, 8GB RAM

## 🔧 Установка

### Шаг 1: Подключаемся к серверу

```bash
ssh root@31.97.76.106
```

### Шаг 2: Загружаем и запускаем скрипт

```bash
# Загружаем скрипт
curl -O https://raw.githubusercontent.com/твой-username/FixDrive/main/k8s/server-setup.sh

# Или копируем вручную из репозитория
# nano server-setup.sh  # вставляем содержимое

# Запускаем установку
chmod +x server-setup.sh
./server-setup.sh
```

### Шаг 3: Проверяем установку

```bash
# Статус кластера
kubectl get nodes

# Статус подов
kubectl get pods -A

# Ingress controller
kubectl get svc -n ingress-nginx
```

## 🔑 Настройка CI/CD

### Шаг 4: Получаем kubeconfig

```bash
# На сервере выполняем
cat ~/.kube/config | base64 -w 0
```

### Шаг 5: Добавляем в GitHub Secrets

1. Идем в GitHub → Settings → Secrets and variables → Actions
2. Создаем `KUBE_CONFIG` с полученной base64 строкой

## 🌐 Настройка портов

### Вариант 1: Разные порты (рекомендуется)

```yaml
# docker-compose.yml (текущий)
ports:
  - "80:8080"    # HTTP
  - "443:8080"   # HTTPS  

# Kubernetes ingress
# NodePort: 30000-32767
# Пример: 30080 для HTTP, 30443 для HTTPS
```

### Вариант 2: Один порт (требует остановки docker-compose)

```bash
# Останавливаем docker-compose
docker-compose down

# Kubernetes займет порты 80/443 через ingress
```

## 📊 Мониторинг и управление

### Полезные команды:

```bash
# Статус кластера
kubectl get nodes
kubectl get pods -A

# Логи
kubectl logs -f deployment/fixdrive-app -n fixdrive

# Масштабирование
kubectl scale deployment fixdrive-app --replicas=3 -n fixdrive

# Обновление
kubectl rollout restart deployment/fixdrive-app -n fixdrive
```

## 🎯 Результат

После установки у тебя будет:

✅ **Kubernetes кластер** на сервере 31.97.76.106  
✅ **Ingress controller** для внешнего доступа  
✅ **CI/CD готов** для деплоя  
✅ **Автоскалинг** (HPA) настроен  
✅ **Мониторинг** через kubectl  

## 🔄 Миграция с docker-compose

1. **Тестируем** Kubernetes деплой на других портах
2. **Убеждаемся** что все работает  
3. **Переключаемся** постепенно
4. **Останавливаем** docker-compose когда готовы

## 📞 Если что-то пошло не так

```bash
# Полная переустановка
kubeadm reset
rm -rf ~/.kube
./server-setup.sh

# Логи установки
journalctl -u kubelet
``` 