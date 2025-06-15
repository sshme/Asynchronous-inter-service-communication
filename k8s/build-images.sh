#!/bin/bash

# This script builds the Docker images for the microservices directly inside minikube's docker daemon.
# It should be run from the root of the project.

echo "Switching to minikube's docker daemon..."
eval $(minikube -p minikube docker-env)

echo "Building Docker images..."

echo "Building orders-client for Kubernetes..."
docker build -t orders-client:k8s -f ./orders-client/Dockerfile.k8s ./orders-client

echo "Building orders-service..."
docker build -t orders-service:latest ./orders-service

echo "Building payments-service..."
docker build -t payments-service:latest ./payments-service

echo "Docker images built successfully inside minikube."

eval $(minikube docker-env -u)

echo "All images built successfully!"

docker images | grep -E "(orders-service|payments-service|orders-client)"
echo "Images are ready for deployment in Kubernetes!" 