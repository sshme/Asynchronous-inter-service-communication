#!/bin/bash

# This script deploys the application to Kubernetes.

echo "Deploying to Kubernetes..."

echo "Applying namespace..."
kubectl apply -f k8s/manifests/namespace.yaml

echo "Waiting for namespace to be created..."
sleep 5 

# Apply CRDs, RBAC, and other fundamental configs first
kubectl apply -f k8s/manifests/traefik-rbac.yaml
kubectl apply -f k8s/manifests/traefik.yaml
kubectl apply -f k8s/manifests/orders-db.yaml
kubectl apply -f k8s/manifests/payments-db.yaml
kubectl apply -f k8s/manifests/kafka.yaml
kubectl apply -f k8s/manifests/redis.yaml

echo "Waiting for databases and Kafka to be ready..."
kubectl wait --for=condition=ready pod/orders-db-0 -n microservices --timeout=1200s
kubectl wait --for=condition=ready pod/payments-db-0 -n microservices --timeout=1200s
kubectl wait --for=condition=ready pod/kafka-0 -n microservices --timeout=1800s

# Run the database migration jobs
echo "Running database migration jobs..."
kubectl apply -f k8s/manifests/migration-job.yaml
kubectl apply -f k8s/manifests/payments-migration-job.yaml

echo "Waiting for migration jobs to complete..."
kubectl wait --for=condition=complete job/orders-service-migration -n microservices --timeout=1200s
kubectl wait --for=condition=complete job/payments-service-migration -n microservices --timeout=1200s

echo "Applying application services and ingress..."
kubectl apply -R -f k8s/manifests/

sleep 10
kubectl apply -f k8s/manifests/deployments.yaml

echo "Deployment complete."
echo "Run 'minikube tunnel' in a separate terminal."
echo "Then you can access the application at http://localhost (or the minikube IP)"
echo "You can access the Traefik dashboard at http://localhost:8080" 