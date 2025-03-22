# User Server API
(by claude)

A simple Go API for user registration and authentication using JWT tokens and SQLite.

## Features

- User registration with invite tokens
- Authentication using JWT tokens signed with RSA keys
- Data storage in SQLite
- Password hashing using bcrypt
- Invite tokens creation only via command line
- Configuration via `.env` file or command line flags
- Public key endpoint for token verification in other services
- Refresh token support for extended sessions

## Requirements

- Go 1.24 or higher
- SQLite

## Installation

1. Clone the repository:
```bash
git clone https://github.com/user/user-server.git
cd user-server
```

2. Install dependencies:
```bash
go mod tidy
```

3. Generate RSA keys (if you haven't done so already):
```bash
mkdir -p keys
openssl genpkey -algorithm RSA -out keys/private.pem -pkeyopt rsa_keygen_bits:2048
openssl rsa -pubout -in keys/private.pem -out keys/public.pem
```

4. Create `.env` file based on the example:
```bash
cp .env.example .env
```

## Configuration

The application can be configured in two ways:

1. Via `.env` file:
```env
# Path to SQLite database file
DB_PATH=user-server.db

# Paths to RSA keys
PRIVATE_KEY_PATH=keys/private.pem
PUBLIC_KEY_PATH=keys/public.pem

# HTTP listen address
ADDR=:8080

# JWT token lifetime in hours
JWT_TTL=24
```

2. Via command line flags:
```
-db           Path to SQLite database file (default: value from .env or "user-server.db")
-private-key  Path to RSA private key (default: value from .env or "keys/private.pem")
-public-key   Path to RSA public key (default: value from .env or "keys/public.pem")
-addr         HTTP listen address (default: value from .env or ":8080")
-jwt-ttl      JWT token lifetime in hours (default: value from .env or 24)
```

Command line flags take precedence over values from the `.env` file.

## Running

### Starting the server

```bash
go run cmd/server/main.go
```

### Creating an invite token

```bash
go run cmd/invite/main.go
```

## API Endpoints

### User Registration

```
POST /api/auth/register
```

Request body:
```json
{
  "username": "johndoe",
  "password": "password123",
  "invite_token": "your_invite_token"
}
```

Response:
```json
{
  "access_token": "jwt_token",
  "refresh_token": "refresh_token",
  "expires_in": 86400
}
```

### User Login

```
POST /api/auth/login
```

Request body:
```json
{
  "username": "johndoe",
  "password": "password123"
}
```

Response:
```json
{
  "access_token": "jwt_token",
  "refresh_token": "refresh_token",
  "expires_in": 86400
}
```

### Refresh Token

```
POST /api/auth/refresh
```

Request body:
```json
{
  "refresh_token": "your_refresh_token"
}
```

Response:
```json
{
  "access_token": "new_jwt_token",
  "refresh_token": "new_refresh_token",
  "expires_in": 86400
}
```

### Getting Current User Information

```
GET /api/me
```

Headers:
```
Authorization: Bearer jwt_token
```

Response:
```json
{
  "id": 1,
  "username": "johndoe"
}
```

### Getting Public Key for Token Verification

```
GET /api/auth/public-key
```

Response:
```json
{
  "key": "-----BEGIN PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA...\n-----END PUBLIC KEY-----\n"
}
```

This endpoint returns the RSA public key in PEM format, which can be used by other services to verify JWT tokens issued by this server.

## Building

### Building the server

```bash
go build -o user-server cmd/server/main.go
```

### Building the invite token creation utility

```bash
go build -o create-invite cmd/invite/main.go
```

## Docker

### Building the Image

```bash
docker build -t user-server .
```

### Running the Container

```bash
docker run -d \
  --name user-server \
  -p 8080:8080 \
  -v $(pwd)/keys:/app/keys \
  -v $(pwd)/data:/app/data \
  --env-file .env \
  user-server
```

### Environment Variables

Create a `.env` file based on `.env.example` and configure the following variables:

- `DB_PATH` - path to the database file (default: user-server.db)
- `PRIVATE_KEY_PATH` - path to the RSA private key
- `PUBLIC_KEY_PATH` - path to the RSA public key
- `ADDR` - HTTP listen address (default: :8080)
- `JWT_TTL` - JWT token lifetime in hours (default: 24)

### Volume Mounts

- `keys/` - directory for RSA keys
- `data/` - directory for SQLite database 