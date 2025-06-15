#!/bin/bash

# This script deploys the microservices to Kubernetes.
# It should be run from the root of the project.

echo "Deploying to Kubernetes..."

echo "Applying Kubernetes manifests..."
kubectl apply -f k8s/manifests/namespace.yaml
echo "Waiting for namespace to be created..."
sleep 5

# Apply all other manifests
kubectl apply -f k8s/manifests/traefik-rbac.yaml
kubectl apply -f k8s/manifests/traefik.yaml
kubectl apply -f k8s/manifests/orders-db.yaml
kubectl apply -f k8s/manifests/payments-db.yaml
kubectl apply -f k8s/manifests/kafka.yaml
kubectl apply -f k8s/manifests/orders-service-config.yaml
kubectl apply -f k8s/manifests/payments-service-config.yaml
kubectl apply -f k8s/manifests/orders-service.yaml
kubectl apply -f k8s/manifests/payments-service.yaml
kubectl apply -f k8s/manifests/ingress.yaml

# Apply the client last
kubectl apply -f k8s/manifests/orders-client.yaml

kubectl apply -f k8s/manifests/deployments.yaml

echo "Deployment complete."
MINIKUBE_IP=$(minikube ip)
echo "Run 'minikube tunnel' in a separate terminal."
echo "The application at http://localhost (or http://$MINIKUBE_IP if tunnel doesn't work)"
echo "Traefik dashboard at http://localhost:8080" 