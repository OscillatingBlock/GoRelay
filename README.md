# GoRelay

A simple HTTP reverse proxy load balancer written in Go, designed to distribute incoming requests across multiple backend servers using a round-robin algorithm, with health checks to ensure reliability.

## Features
- **Load Balancing**: Distributes requests across backends using round-robin scheduling.
- **Health Checks**: Periodically checks backend health (every 10s by default) via HEAD requests.
- **Graceful Shutdown**: Supports clean server shutdown on SIGINT/SIGTERM.
- **Configurable**: Uses a YAML config file for port, backends, and health check interval.
- **Logging**: Detailed logs for errors, warnings, and server events.

## Project Structure
```
GoRelay/
├── cmd/api/main.go            # Entry point for the load balancer
├── configs/config.yaml        # Configuration file
├── internal/
│   ├── loadbalancer/
│   │   ├── delivery/          # HTTP handlers
│   │   ├── repository/        # Config and health check logic
│   │   ├── usecase/           # Load balancing and health check business logic
│   │   └── models/            # Backend and server pool models
│   └── server/                # HTTP server setup
├── pkg/
│   └── logger/                # Custom logger
```

## Prerequisites
- Go 1.18 or higher
- A terminal to run commands
- Optional: Backend servers for testing (e.g., simple HTTP servers)

## Installation
1. Clone the repository:
   ```bash
   git clone git@github.com:OscillatingBlock/GoRelay.git 
   cd GoRelay
   ```
2. Install dependencies:
   ```bash
   go mod tidy
   ```
3. update configurations
4. go run cmd/api/main.go


## Configuration
Edit `configs/config.yaml`:
```yaml
port: "8080"                <!-- mention the port for gorelay to listen on -->
backends:                   <!-- list the available backends url -->
  - "http://localhost:8081"
  - "http://localhost:8082"
healthInterval: 10s         <!-- interval to run health checks and update backend status -->
algorithm: round_robin      <!-- available algorithms: round_robin, leastconn -->
```

## Shutdown
Press `Ctrl+C` to trigger graceful shutdown, allowing in-flight requests to complete within 10 seconds.

## Troubleshooting
- **Health endpoint returns incorrect count**: Ensure `healthInterval` is set and backends respond to HEAD requests on `/`.
- **Connection refused errors**: Verify backend servers are running on the configured ports.
- Check logs in the console for detailed error messages.

## Contributing
- Fork the repository, make changes, and submit a pull request.
- Ensure tests pass and add new tests for new features.
