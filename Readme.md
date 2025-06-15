# Production-Go-API-Template

A starter kit for building production-ready REST APIs in Go.
It follows clean-architecture principles, so you can ship fast without sacrificing structure.

## Table of Contents

- [What and Why?](#what-and-why)
- [Quick Start](#quick-start)
- [Project Structure](#project-structure)
  - [`/cmd` - Application Entry Point](#cmd---application-entry-point)
  - [`/config` - Configuration Management](#config---configuration-management)
  - [`/api` - HTTP Layer](#api---http-layer)
  - [`/pkg` - Shared Utilities](#pkg---shared-utilities)
- [The Default Item API](#the-default-item-api)
- [Security Features](#security-features)
- [Observability](#observability)
- [Configuration](#configuration)
- [Possible Enhancements](#possible-enhancements)
  - [Database & Persistence](#database--persistence)
  - [Security Enhancements](#security-enhancements)
- [Learnings](#learnings)
- [Contributing](#contributing)

## What and Why?

A few months back I wanted to turn everything I’d learned about Go into a real-world app.

**The challenge:** Design a scalable project layout while relying on as few third-party libraries as possible. (That ruled out frameworks like Chi)
This template is the result—and it’s the codebase I use in the [blog post](https://trakora.de/blog/weg-zu-minimalistischen-http‑apis)

The item management is just an example - you can easily adapt this foundation for any domain-specific API you need to build.


## Quick Start

1. **Generate your API credentials**:
   ```bash
   python generate_tokens.py
   ```

2. **Set up environment**:
   ```bash
   cp .env.sample .env
   # Paste the tokens you just generated into .env
   ```

3. **Run the server**:
   ```bash
   go run cmd/main.go
   ```
   or if you prefer live reloads start it with [air](https://github.com/air-verse/air)
   ```bash
   air
   ```

The API starts on port 8080 with a SQLite database and full request logging.

## Project Structure
How the Code Is Organized:


### `/cmd` - Application Entry Point

The main application lives here. `main.go` ties everything together - loads configuration, sets up the database, configures middleware, and starts the HTTP server. It handles graceful shutdown and wires up all the components.

### `/config` - Configuration Management  

Environment-based configuration that supports development and production settings. Handles server ports, timeouts, CORS settings, authentication tokens, security parameters, and database configuration. Uses struct tags for easy environment variable mapping.

### `/api` - HTTP Layer

The web layer of the application:

- **`/api/router`** - HTTP routing setup. Sets up all the routes and connects them to handlers
- **`/api/router/middleware`** - Request processing pipeline:
  - `authentication.go` - HMAC + Bearer token security with IP blocking and rate limiting
  - `requestlog.go` - Comprehensive request/response logging for debugging
  - `cors.go` - Cross-origin request handling
  - `request_id.go` - Unique ID tracking for each request
  - `inject_deps.go` - Dependency injection for handlers
- **`/api/resource`** - Domain-specific handlers and logic:
  - `health/` - Health check endpoints for monitoring
  - `item/` - Sample CRUD operations for items

### `/pkg` - Shared Utilities

Reusable packages:

- **`/pkg/logger`** - Structured logging with request ID correlation using zerolog
- **`/pkg/router`** - HTTP response utilities and route mounting helpers  
- **`/pkg/validator`** - JSON validation and context value extraction utilities
- **`/pkg/constants`** - Application-wide constants
- **`/pkg/contextkeys`** - Type-safe context keys for request scoped data

## The Default Item API

The core functionality demonstrates a complete REST API pattern:

**Endpoints:**
- `POST /api/v1/items` - Create new items
- `GET /api/v1/items` - List all items  
- `GET /api/v1/items/{id}` - Get specific item
- `PUT /api/v1/items/{id}` - Update item
- `DELETE /api/v1/items/{id}` - Delete item

**Architecture Pattern:**
Each resource follows handler → service → repository pattern for clean separation of concerns.

## Security Features

This isn't just a simple CRUD API - it has enterprise-grade simple security:

**Two-Layer Authentication:**
1. **Bearer Token** - Every request needs `Authorization: Bearer <token>` header
2. **HMAC Signature** - Additional `X-Timestamp` and `X-Signature` headers prevent replay attacks

**Rate Limiting:**
- Tracks failed authentication attempts per IP address
- Progressive slowdown - response time increases with each failed attempt  
- Automatic IP blocking after too many failures
- Subnet-level blocking for persistent attackers
- Automatic cleanup of expired blocks

**Request Security:**
- Timestamp validation (±5 minutes) prevents replay attacks
- All requests need current timestamp and valid HMAC signature
- Client IP extraction handles load balancers and proxies correctly

## Observability 

**Request Tracing:**
Every request gets a unique ID that flows through all logs, making debugging much easier.

**Comprehensive Logging:**
- Full request/response logging with sanitized headers
- Structured JSON logs with request ID correlation
- Performance metrics (response time, status codes)
- Security events (failed auth attempts, IP blocks)

**Health Monitoring:**
- `/healthz` - Basic health check
- `/livez` - Liveness probe with uptime and system info

## Configuration

Environment variables control everything:

```bash
# Server settings
SERVER_PORT=8080
SERVER_DEBUG=true
SERVER_CORS_ORIGINS=*

# Security settings  
SECURITY_MAX_FAILURES=5
SECURITY_FAIL_WINDOW=1m
SECURITY_BLOCK_DURATION=10m
SECURITY_SLOWDOWN_STEP=200ms

# Authentication (generated by generate_tokens.py)
API_TOKEN=your-secure-token
SECRET=your-hmac-secret
```


## Possible Enhancements

While this API is already quite solid, here are some enhancements you might consider for even better production readiness:

### Database & Persistence

**PostgreSQL Migration**
- Replace SQLite with e.g. PostgreSQL
- Add connection pooling and database health checks
- Implement proper database migrations with versioning (like golang-migrate)
- Add read replicas for scaling read operations

**Caching Layer**
- Redis for session storage and rate limiting data (currently in-memory)
- Cache frequently accessed items to reduce database load
- Implement cache invalidation strategies

### Security Enhancements

**Token Management**
- Automatic token rotation (currently tokens are just generated once)
- JWT tokens with proper expiration and refresh mechanisms
- API key management with different permission levels per client

**Advanced Protection**
- Request body hashing in HMAC signature (prevents tampering)
- Google Cloud Armor or Cloudflare for DDoS protection
- Rate limiting based on authenticated user, not just IP

## Learnings
For any small, publicly exposed API, this template is a rock-solid starting point.
I deployed the service on Google Kubernetes Engine (GKE) and injected the API keys into environment variables using the Google Secret Provider Class.

- The HMAC timestamp is very sensitive to clock drift.
- The API always operates in UTC, regardless of the server’s local timezone.
- Clients must also send timestamps in UTC; otherwise, the hash check will fail.

### Minimal Python example:
```py

now_utc = datetime.now()
timestamp = int(now_utc.timestamp())

message = f"{api_token}|{timestamp}|{method}|{path}"
 
signature = hmac.new(
    hmac_secret.encode('utf-8'),
    message.encode('utf-8'),
    hashlib.sha256
).hexdigest()

headers = {
    "Authorization": f"Bearer {api_token}",
    "X-Timestamp": str(timestamp),
    "X-Signature": signature,
    "Content-Type": "application/json"
}

```

## Contributing

Contributions are welcome! This template is designed to be a solid foundation that can be enhanced and adapted for various use cases.

**How to Contribute:**

1. Fork the repository and create a feature branch from `main`
2. Implement your enhancement
3. Update documentation - including this README if needed
4. Submit a pull request with a clear description of your changes

**Code Style:**
- Follow standard Go conventions
- **Keep the minimal dependency philosophy**
- Maintain clear separation

Feel free to open an issue first to discuss changes.