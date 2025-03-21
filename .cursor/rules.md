# User Authentication Server

## Project Overview
A secure user authentication server built with Go, providing JWT-based authentication with RSA encryption for invite tokens. The server uses SQLite for data storage and follows best practices for security and performance.

## Project Structure
```
user-server/
├── cmd/
│   ├── server/         # Main server application
│   └── invite/         # Utility for creating invite tokens
├── pkg/
│   ├── auth/          # Authentication logic and JWT handling
│   ├── database/      # Database operations and models
│   └── handlers/      # HTTP request handlers
├── keys/              # RSA keys storage (not in git)
├── data/             # SQLite database storage
├── Dockerfile        # Multi-stage Docker build
├── .env.example      # Environment variables template
└── go.mod           # Go module definition
```

## Key Features
- JWT-based authentication
- RSA-encrypted invite tokens
- SQLite database for user storage
- Docker support with multi-stage builds
- Environment-based configuration
- Secure password hashing
- Rate limiting for login attempts

## Development Guidelines

### Code Style
- Follow standard Go formatting (gofmt)
- Use meaningful variable and function names
- Add comments for complex logic
- Keep functions focused and small
- Use error handling consistently

### Security Practices
- Never commit sensitive data (keys, passwords)
- Use environment variables for configuration
- Implement proper input validation
- Follow OWASP security guidelines
- Use secure password hashing (bcrypt)
- Implement rate limiting for sensitive endpoints

### Docker Guidelines
- Use multi-stage builds for minimal image size
- Run as non-root user
- Mount sensitive data as volumes
- Use environment variables for configuration
- Keep base images up to date
- Use golang:1.24 for build stage
- Use alpine:latest for final stage

### API Endpoints
```
POST /api/v1/auth/register - User registration
POST /api/v1/auth/login    - User authentication
GET  /api/v1/auth/key      - Get public RSA key
```

### Environment Variables
Required environment variables:
- `PORT` - Server port (default: 8080)
- `DB_PATH` - SQLite database path
- `JWT_SECRET` - JWT signing secret
- `RSA_PRIVATE_KEY_PATH` - Path to RSA private key
- `RSA_PUBLIC_KEY_PATH` - Path to RSA public key

### Dependencies
- Go 1.24+
- SQLite3
- Required Go packages (see go.mod)

## Build and Run

### Local Development
```bash
# Install dependencies
go mod download

# Build server
go build -o user-server cmd/server/main.go

# Build invite token utility
go build -o create-invite cmd/invite/main.go

# Run server
./user-server
```

### Docker
```bash
# Build image
docker build -t user-server .

# Run container
docker run -d \
  --name user-server \
  -p 8080:8080 \
  -v $(pwd)/keys:/app/keys \
  -v $(pwd)/data:/app/data \
  --env-file .env \
  user-server
```

## Testing
- Write unit tests for critical components
- Test error handling scenarios
- Validate security measures
- Test rate limiting
- Verify token encryption/decryption

## Deployment
- Use HTTPS in production
- Set up proper monitoring
- Implement logging
- Configure backup strategy
- Set up CI/CD pipeline

## Maintenance
- Regular dependency updates
- Security patches
- Performance monitoring
- Database maintenance
- Log rotation

## Contributing
1. Fork the repository
2. Create feature branch
3. Make changes
4. Run tests
5. Submit pull request

## License
[Specify License] 