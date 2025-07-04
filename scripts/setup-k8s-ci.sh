#!/bin/bash

# Скрипт для настройки Kubernetes CI/CD
set -e

echo "🔧 Настройка Kubernetes CI/CD для FixDrive"
echo "============================================"

# Проверяем что kubectl доступен
if ! command -v kubectl &> /dev/null; then
    echo "❌ kubectl не найден. Установи kubectl сначала."
    exit 1
fi

# Проверяем подключение к кластеру
if ! kubectl cluster-info &> /dev/null; then
    echo "❌ Нет подключения к Kubernetes кластеру."
    echo "Запусти: minikube start"
    exit 1
fi

echo "✅ Подключение к кластеру активно"

# Получаем текущий контекст
CURRENT_CONTEXT=$(kubectl config current-context)
echo "📋 Текущий контекст: $CURRENT_CONTEXT"

# Создаем kubeconfig для CI/CD
echo "🔑 Создание kubeconfig для GitHub Actions..."
kubectl config view --raw --minify > kubeconfig-ci.yaml

# Кодируем в base64
echo "🔢 Кодирование в base64..."
KUBECONFIG_BASE64=$(cat kubeconfig-ci.yaml | base64 | tr -d '\n')

echo ""
echo "🎯 ДОБАВЬ В GITHUB SECRETS:"
echo "=========================="
echo "Имя секрета: KUBE_CONFIG"
echo "Значение:"
echo "$KUBECONFIG_BASE64"
echo ""

# Создаем .env файл с примером секретов
echo "📝 Создание примера секретов..."
cat > github-secrets-example.txt << 'EOF'
# Секреты для GitHub Actions

# Docker Hub
DOCKER_USERNAME=твой_docker_username
DOCKER_PASSWORD=твой_docker_password_или_токен

# Kubernetes (скопируй значение выше)
KUBE_CONFIG=сгенерированное_значение_base64

# Переменные приложения
DB_PASSWORD=your_db_password
JWT_SECRET=your-super-secret-jwt-key-here
TWILIO_ACCOUNT_SID=ACxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
TWILIO_AUTH_TOKEN=your_twilio_auth_token_here
TWILIO_FROM_PHONE=+1234567890
EOF

echo "✅ Создан файл: github-secrets-example.txt"
echo "✅ Создан файл: kubeconfig-ci.yaml"

# Проверяем minikube
if command -v minikube &> /dev/null && minikube status &> /dev/null; then
    MINIKUBE_IP=$(minikube ip)
    echo ""
    echo "🌐 Для доступа к приложению:"
    echo "Добавь в /etc/hosts:"
    echo "$MINIKUBE_IP fixdrive.local"
fi

echo ""
echo "🚀 Следующие шаги:"
echo "1. Открой GitHub репозиторий → Settings → Secrets and variables → Actions"
echo "2. Добавь все секреты из github-secrets-example.txt"
echo "3. Сделай push или запусти workflow вручную"
echo "4. Выбери 'kubernetes' в качестве среды деплоя"
echo ""
echo "📚 Подробная документация: CI-CD-SETUP.md"

# Очистка
rm -f kubeconfig-ci.yaml

echo "✅ Настройка завершена!" 