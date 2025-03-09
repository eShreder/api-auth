# User Server API

Simple Go API for user registration and authentication using JWT tokens and SQLite.

## Features

- User registration with invite tokens
- Authentication using JWT tokens signed with RSA keys
- Data storage in SQLite
- Password hashing using bcrypt
- Invite tokens creation only via command line
- Configuration via `.env` file or command line flags
- Public key endpoint for token verification in other services

## Requirements

- Go 1.16 or higher
- SQLite

## Installation

1. Clone the repository:
```
git clone https://github.com/user/user-server.git
cd user-server
```

2. Install dependencies:
```
go mod tidy
```

3. Generate RSA keys (if you haven't done so already):
```
mkdir -p keys
openssl genpkey -algorithm RSA -out keys/private.pem -pkeyopt rsa_keygen_bits:2048
openssl rsa -pubout -in keys/private.pem -out keys/public.pem
```

4. Create `.env` file based on the example:
```
cp .env.example .env
```

## Configuration

The application can be configured in two ways:

1. Via `.env` file:
```
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

```
go run cmd/server/main.go
```

### Creating an invite token

```
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
  "email": "user@example.com",
  "password": "password123",
  "invite_token": "your_invite_token"
}
```

Response:
```json
{
  "token": "jwt_token"
}
```

### User Login

```
POST /api/auth/login
```

Request body:
```json
{
  "email": "user@example.com",
  "password": "password123"
}
```

Response:
```json
{
  "token": "jwt_token"
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
  "email": "user@example.com"
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

```
go build -o user-server cmd/server/main.go
```

### Building the invite token creation utility

```
go build -o create-invite cmd/invite/main.go
``` 