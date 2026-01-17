# Sovereign Firm: Production Architecture

## Overview

Transform prototype into production-ready, multi-tenant SaaS with auth, database, caching, and object storage.

## Infrastructure Stack

| Component | Technology | Purpose |
|-----------|------------|---------|
| Database | PostgreSQL + pgx | Project metadata, users, tenants |
| Cache | Redis | State caching, rate limiting, sessions |
| Object Storage | MinIO | Generated artifacts, file storage |
| Auth | JWT (RS256) | Authentication & authorization |
| API Framework | Chi Router | OpenAPI-compliant routing |
| Workflow | Temporal | Already implemented |

## Directory Structure

```
pkg/
├── database/           # PostgreSQL layer
│   ├── postgres.go     # Connection pool, RLS context
│   ├── migrations/     # SQL migrations
│   └── queries/        # sqlc generated code
├── auth/               # Authentication
│   ├── jwt.go          # Token generation/validation
│   ├── middleware.go   # Chi middleware
│   ├── password.go     # Argon2id hashing
│   └── handlers.go     # Auth endpoints
├── cache/              # Redis layer
│   ├── redis.go        # Client wrapper
│   ├── project.go      # Project state cache
│   └── session.go      # Session management
├── storage/            # MinIO layer
│   ├── minio.go        # Client wrapper
│   └── artifacts.go    # Artifact storage
└── api/                # OpenAPI layer
    ├── openapi.yaml    # API specification
    ├── types.gen.go    # Generated types
    ├── server.gen.go   # Generated routes
    └── handlers.go     # Handler implementations
```

## Database Schema

### Tables

```sql
-- Tenants (organizations)
tenants: id, name, slug, plan, settings, created_at

-- Users
users: id, tenant_id, email, password_hash, role, created_at

-- Projects (links to Temporal workflows)
projects: id, tenant_id, workflow_id, name, description, phase, config, created_by, created_at, updated_at

-- Artifacts (links to MinIO objects)
artifacts: id, project_id, type, path, size, checksum, created_at
```

### Row-Level Security

All tables have RLS policies enforcing tenant isolation via `SET app.tenant_id`.

## API Endpoints

### Authentication
- `POST /api/auth/register` - Register new user + tenant
- `POST /api/auth/login` - Get JWT tokens
- `POST /api/auth/refresh` - Refresh access token
- `POST /api/auth/logout` - Invalidate refresh token

### Projects
- `GET /api/projects` - List user's projects
- `POST /api/projects` - Create new project (starts Temporal workflow)
- `GET /api/projects/{id}` - Get project details
- `GET /api/projects/{id}/state` - Get full Temporal state
- `POST /api/projects/{id}/message` - Send message to workflow
- `GET /api/projects/{id}/stream` - WebSocket for real-time updates
- `GET /api/projects/{id}/artifacts` - List project artifacts
- `GET /api/projects/{id}/artifacts/{artifactId}` - Download artifact

### Health
- `GET /health` - Basic health check
- `GET /health/ready` - Readiness (DB, Redis, MinIO, Temporal)

## Caching Strategy

| Key Pattern | TTL | Purpose |
|-------------|-----|---------|
| `project:{id}:state` | 5min | ConsultancyState from Temporal |
| `project:{id}:files` | 1min | File list cache |
| `user:{id}:session` | 24h | Session data |
| `tenant:{id}:rate` | 1min | Rate limit counter |
| `ws:connections:{id}` | - | Active WebSocket tracking |

## MinIO Bucket Structure

```
sovereign-firm/
├── {tenant_id}/
│   └── {project_id}/
│       ├── source/          # Generated source code
│       ├── docs/            # Generated documentation
│       ├── artifacts/       # Build artifacts
│       └── snapshots/       # Phase snapshots
```

## Environment Variables

```bash
# Database
DATABASE_URL=postgres://user:pass@localhost:5432/sovereign_firm
DATABASE_MAX_CONNS=25

# Redis
REDIS_URL=redis://localhost:6379/0
REDIS_PASSWORD=

# MinIO
MINIO_ENDPOINT=localhost:9000
MINIO_ACCESS_KEY=minioadmin
MINIO_SECRET_KEY=minioadmin
MINIO_BUCKET=sovereign-firm
MINIO_USE_SSL=false

# Auth
JWT_PRIVATE_KEY_PATH=/path/to/private.pem
JWT_PUBLIC_KEY_PATH=/path/to/public.pem
JWT_ACCESS_TOKEN_TTL=15m
JWT_REFRESH_TOKEN_TTL=168h

# Temporal (existing)
TEMPORAL_HOST=localhost:7233

# Server
PORT=8080
```

## Dependencies to Add

```go
// go.mod additions
github.com/go-chi/chi/v5
github.com/go-chi/cors
github.com/golang-jwt/jwt/v5
github.com/jackc/pgx/v5
github.com/redis/go-redis/v9
github.com/minio/minio-go/v7
github.com/golang-migrate/migrate/v4
golang.org/x/crypto  // argon2
```
