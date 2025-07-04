#!/bin/bash

# Деплой FixDrive в Kubernetes

set -e

echo "🚀 Деплой FixDrive в Kubernetes..."

# Создаем namespace
echo "📦 Создаю namespace..."
kubectl apply -f namespace.yaml

# Создаем configmap и secret
echo "⚙️ Создаю конфигурацию..."
kubectl apply -f configmap.yaml
kubectl apply -f secret.yaml

# Деплоим базу данных
echo "🗄️ Деплою PostgreSQL..."
kubectl apply -f postgres.yaml

# Деплоим Redis
echo "🔥 Деплою Redis..."
kubectl apply -f redis.yaml

# Ждем готовности БД
echo "⏳ Жду готовности PostgreSQL..."
kubectl wait --for=condition=available --timeout=300s deployment/postgres -n fixdrive

# Запускаем миграции
echo "🔄 Запускаю миграции..."
kubectl apply -f migration-job.yaml

# Деплоим приложение
echo "🏗️ Деплою приложение..."
kubectl apply -f app.yaml

# Создаем ingress
echo "🌐 Создаю ingress..."
kubectl apply -f ingress.yaml

# Создаем HPA
echo "📊 Создаю автоскейлинг..."
kubectl apply -f hpa.yaml

echo "✅ Деплой завершен!"
echo ""
echo "Для проверки статуса:"
echo "  kubectl get pods -n fixdrive"
echo "  kubectl get svc -n fixdrive"
echo "  kubectl get ingress -n fixdrive"
echo ""
echo "Логи приложения:"
echo "  kubectl logs -f deployment/fixdrive-app -n fixdrive"
echo ""
echo "Для доступа к приложению добавь в /etc/hosts:"
echo "  <IP-адрес-кластера> fixdrive.local" 