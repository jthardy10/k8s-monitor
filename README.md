# Kubernetes Cluster Monitor

A real-time monitoring solution for Kubernetes clusters with web dashboard.

## Features

- Multi-cluster monitoring
- Real-time metrics collection
- Resource usage tracking (CPU/Memory)
- Pod and node status monitoring
- Web dashboard with charts
- REST API with authentication

## Quick Start

### Prerequisites

- Go 1.21 or later
- Access to Kubernetes cluster(s)
- kubectl configured with cluster access

### Installation

1. Clone the repository:

   git clone <your-repo-url>
   cd k8s-monitor

2. Build the application:

   go build -o k8s-monitor cmd/server/main.go

3. Run the monitor:

   ./k8s-monitor

4. Access the dashboard at `http://localhost:8080/dashboard`

### API Authentication

All API endpoints (except /health) require an API key header:

   X-API-Key: dev-api-key-123

### API Endpoints

- `GET /health` - Health check endpoint
- `GET /clusters` - List registered clusters
- `POST /clusters` - Register new cluster
- `GET /clusters/:name` - Get cluster metrics
- `GET /clusters/:name/history` - Get historical metrics
- `DELETE /clusters/:name` - Remove cluster registration

## Development

### Project Structure

   .
   ├── cmd/
   │   └── server/
   │       └── main.go
   ├── internal/
   │   ├── cluster/
   │   ├── middleware/
   │   ├── monitoring/
   │   └── storage/
   └── static/
       └── index.html

### Building from Source

   go mod download
   go build -o k8s-monitor cmd/server/main.go

## License

MIT License
