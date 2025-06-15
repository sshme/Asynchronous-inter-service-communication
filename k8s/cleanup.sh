#!/bin/bash

# This script cleans up all Kubernetes resources created by the deploy.sh script.
# It should be run from the root of the project.

echo "Cleaning up Kubernetes resources..."

kubectl delete namespace microservices

echo "Cleanup complete." 