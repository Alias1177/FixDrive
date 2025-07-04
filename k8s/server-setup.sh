#!/bin/bash

# Скрипт установки Kubernetes на сервер
# Запускать на сервере 31.97.76.106

set -e

echo "🚀 Устанавливаю Kubernetes на сервер..."

# Обновляем систему
apt-get update && apt-get upgrade -y

# Устанавливаем Docker (если еще не установлен)
if ! command -v docker &> /dev/null; then
    curl -fsSL https://get.docker.com -o get-docker.sh
    sh get-docker.sh
    systemctl enable docker
    systemctl start docker
fi

# Отключаем swap (обязательно для Kubernetes)
swapoff -a
sed -i '/ swap / s/^\(.*\)$/#\1/g' /etc/fstab

# Устанавливаем kubeadm, kubelet, kubectl
apt-get update && apt-get install -y apt-transport-https ca-certificates curl
curl -fsSL https://pkgs.k8s.io/core:/stable:/v1.29/deb/Release.key | gpg --dearmor -o /etc/apt/keyrings/kubernetes-apt-keyring.gpg
echo 'deb [signed-by=/etc/apt/keyrings/kubernetes-apt-keyring.gpg] https://pkgs.k8s.io/core:/stable:/v1.29/deb/ /' | tee /etc/apt/sources.list.d/kubernetes.list

apt-get update
apt-get install -y kubelet kubeadm kubectl
apt-mark hold kubelet kubeadm kubectl

# Инициализируем кластер
echo "🔧 Инициализирую Kubernetes кластер..."
kubeadm init --pod-network-cidr=10.244.0.0/16

# Настраиваем kubectl для root
mkdir -p $HOME/.kube
cp -i /etc/kubernetes/admin.conf $HOME/.kube/config
chown $(id -u):$(id -g) $HOME/.kube/config

# Разрешаем запуск подов на master ноде (для single-node)
kubectl taint nodes --all node-role.kubernetes.io/control-plane-

# Устанавливаем сетевой плагин (Flannel)
kubectl apply -f https://raw.githubusercontent.com/flannel-io/flannel/master/Documentation/kube-flannel.yml

# Ждем готовности
echo "⏳ Ожидаю готовности кластера..."
kubectl wait --for=condition=ready node --all --timeout=300s

# Устанавливаем NGINX Ingress Controller
kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/controller-v1.8.2/deploy/static/provider/baremetal/deploy.yaml

# Ждем готовности ingress
kubectl wait --namespace ingress-nginx --for=condition=ready pod --selector=app.kubernetes.io/component=controller --timeout=300s

echo "✅ Kubernetes успешно установлен!"
echo ""
echo "🔑 Для настройки CI/CD скопируй kubeconfig:"
echo "cat ~/.kube/config | base64 -w 0"
echo ""
echo "📋 Этот base64 строку добавь в GitHub Secrets как KUBE_CONFIG" 