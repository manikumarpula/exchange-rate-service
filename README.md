# Exchange Rate Service

A robust, scalable exchange rate service built with Go and Go-Kit architecture, providing real-time and historical currency exchange rates from multiple providers.

## 🚀 Features

- **Multi-Provider Support**: Aggregates data from exchangerate.host, open.er-api.com, and coinlayer
- **Fallback Strategy**: Automatic failover between providers for high availability
- **Caching**: Redis-based caching for improved performance and reduced API calls
- **Historical Data**: Access to historical exchange rates
- **Currency Conversion**: Convert amounts between different currencies
- **Time Series**: Get exchange rates for date ranges
- **Health Monitoring**: Built-in health checks and monitoring
- **RESTful API**: Clean, documented REST API endpoints

## 🏗️ Architecture

This service follows the **Go-Kit** layered architecture pattern:

```
┌─────────────────┐
│   HTTP Layer    │ ← API handlers, routing, middleware
├─────────────────┤
│  Service Layer  │ ← Business logic, validation
├─────────────────┤
│ Repository Layer│ ← Data access, caching, external APIs
├─────────────────┤
│   Models        │ ← Data structures, DTOs
└─────────────────┘
```

### Key Components

- **Transport Layer**: HTTP handlers with Gorilla Mux
- **Service Layer**: Business logic and use cases
- **Repository Layer**: Data access with provider fallback
- **Cache Layer**: Redis for performance optimization
- **Provider Clients**: Multiple exchange rate data sources

## 📋 Prerequisites

- Go 1.21 or later
- Redis (for caching)
- Docker & Docker Compose (optional)

## 🛠️ Installation

### Quick Start with Setup Script

```bash
# Clone the repository
git clone <your-repo-url>
cd exchange-rate-service

# Run the setup script
./scripts/setup.sh
```

### Manual Setup

1. **Install Go dependencies**:

   ```bash
   go mod tidy
   ```

2. **Set environment variables**:

   ```bash
   cp .env.example .env
   # Edit .env with your configuration
   ```

3. **Start Redis**:

   ```bash
   # Using Docker
   docker run -d -p 6379:6379 redis:7-alpine

   # Or install locally
   brew install redis  # macOS
   sudo apt-get install redis-server  # Ubuntu
   ```

## 🚀 Running the Service

### Local Development

```bash
# Run directly
go run ./cmd/server

# Or use Makefile
make run
```

### Docker

```bash
# Build and run with Docker Compose
make docker-run

# Or manually
docker-compose up --build
```

### Production

```bash
# Build the binary
make build

# Run the binary
./bin/server
```

## 📚 API Endpoints

### Health Check

- `GET /health` - Service health status

### Core API (v1)

- `GET /api/v1/currencies` - List supported currencies
- `GET /api/v1/rates?base=USD` - Get all rates for a base currency
- `GET /api/v1/rates/{base}/{target}` - Get latest rate between currencies
- `GET /api/v1/rates/{base}/{target}/{date}` - Get historical rate
- `POST /api/v1/convert` - Convert currency amounts
- `GET /api/v1/timeseries/{base}/{target}` - Get time series data

### Example Usage

```bash
# Get USD to EUR rate
curl "http://localhost:8080/api/v1/rates/USD/EUR"

# Convert 100 USD to EUR
curl -X POST "http://localhost:8080/api/v1/convert" \
  -H "Content-Type: application/json" \
  -d '{"from_currency": "USD", "to_currency": "EUR", "amount": 100}'

# Get historical rate
curl "http://localhost:8080/api/v1/rates/USD/EUR/2024-01-15"

# Get time series
curl "http://localhost:8080/api/v1/timeseries/USD/EUR?start_date=2024-01-01&end_date=2024-01-31"
```

## ⚙️ Configuration

### Environment Variables

| Variable           | Description               | Default          |
| ------------------ | ------------------------- | ---------------- |
| `PORT`             | Server port               | `8080`           |
| `SHUTDOWN_TIMEOUT` | Graceful shutdown timeout | `30s`            |
| `REDIS_ADDR`       | Redis server address      | `localhost:6379` |
| `REDIS_PASSWORD`   | Redis password            | ``               |
| `REDIS_DB`         | Redis database number     | `0`              |

### Provider Configuration

Each provider can be configured with:

- Base URL
- API Key (if required)
- Timeout settings
- Priority (fallback order)

## 🧪 Testing

```bash
# Run all tests
make test

# Run specific test
go test ./internal/service -v

# Run with coverage
go test -cover ./...
```

## 🔧 Development

### Available Make Commands

```bash
make help          # Show all available commands
make build         # Build the application
make run           # Run locally
make test          # Run tests
make clean         # Clean build artifacts
make fmt           # Format code
make lint          # Lint code
make docker-build  # Build Docker image
make docker-run    # Run with Docker Compose
make docker-stop   # Stop Docker services
```

### Project Structure

```
exchange-rate-service/
├── cmd/server/           # Application entry point
├── internal/             # Private application code
│   ├── api/             # HTTP handlers and routing
│   ├── service/         # Business logic layer
│   ├── repository/      # Data access layer
│   ├── models/          # Data structures
│   └── utils/           # Utilities and helpers
├── configs/              # Configuration management
├── scripts/              # Setup and utility scripts
├── test/                 # Test files
├── Dockerfile            # Docker configuration
├── docker-compose.yml    # Docker Compose setup
├── Makefile              # Build and development commands
└── README.md             # This file
```

## 🚀 Deployment

### Docker

```bash
# Build image
docker build -t exchange-rate-service .

# Run container
docker run -p 8080:8080 exchange-rate-service
```

### Kubernetes

```bash
# Apply manifests
kubectl apply -f k8s/

# Or use Helm
helm install exchange-rate-service ./helm/
```

### Environment Variables for Production

```bash
# Set production environment variables
export ENV=production
export LOG_LEVEL=info
export REDIS_ADDR=redis-cluster:6379
export REDIS_PASSWORD=your-secure-password
```

## 📊 Monitoring & Observability

### Health Checks

- **Endpoint**: `/health`
- **Response**: Service status, provider health, cache status
- **Use Case**: Load balancer health checks, monitoring dashboards

### Logging

- Structured logging with Go-Kit
- Log levels: debug, info, warn, error
- Request tracing and correlation IDs

### Metrics

- Request count and latency
- Cache hit/miss ratios
- Provider response times
- Error rates by endpoint

## 🔒 Security

### API Security

- Input validation and sanitization
- Rate limiting (configurable)
- CORS configuration
- Request logging and monitoring

### Data Security

- No sensitive data logging
- Secure Redis configuration
- Provider API key management
- HTTPS enforcement in production

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

### Development Guidelines

- Follow Go best practices and idioms
- Write tests for new functionality
- Update documentation for API changes
- Use conventional commit messages

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- [Go-Kit](https://github.com/go-kit/kit) - Microservice toolkit
- [Gorilla Mux](https://github.com/gorilla/mux) - HTTP router and URL matcher
- [Redis](https://redis.io/) - In-memory data structure store
- Exchange rate data providers for their APIs

## 📞 Support

- **Issues**: [GitHub Issues](https://github.com/your-username/exchange-rate-service/issues)
- **Documentation**: [API Docs](http://localhost:8080/) (when running)
- **Email**: your-email@example.com

---

**Happy coding! 🚀**
