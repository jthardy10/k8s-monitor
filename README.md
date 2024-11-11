# Kubernetes Monitoring Tool

A multi-cloud Kubernetes monitoring solution that demonstrates cluster monitoring and management capabilities.

## Features
- Cluster registration and management
- Pod status monitoring
- Resource usage tracking
- Health status reporting

## Setup
1. Install dependencies: `go mod download`
2. Run locally: `go run cmd/server/main.go`
3. Build Docker image: `docker build -t k8s-monitor .`
4. Deploy to Kubernetes: `kubectl apply -f deployments/deployment.yaml`

## API Endpoints
- GET /health - Health check endpoint
- GET /clusters - List registered clusters

## Architecture
- Written in Go
- Uses client-go for Kubernetes interaction
- REST API using Gin framework
- Prometheus metrics integration
