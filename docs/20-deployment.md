# Deployment

> **Deploy anywhere** — Docker, Fly.io, Railway, Render, AWS, GCP, or any platform that runs Go binaries.

---

## Introduction

Nimbus applications compile to a single static Go binary with zero runtime dependencies. This means you can deploy to virtually any platform. Nimbus also provides:

- **Nimbus Forge** — Built-in deployment CLI (`nimbus deploy`)
- **Dockerfile** — Production-ready Docker configuration included
- **Platform configs** — Pre-configured for Render, Railway, Fly.io
- **Zero-downtime deploys** — Graceful shutdown with connection draining

---

## Building for Production

### Compile the Binary

```bash
# Standard build
go build -o myapp .

# Optimized production build (smaller binary, stripped debug info)
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o myapp .
```

### Run in Production

```bash
APP_ENV=production PORT=8080 ./myapp
```

Required environment variables:

```env
APP_ENV=production
APP_KEY=your-secret-key-min-32-chars
PORT=8080
DB_DRIVER=postgres
DB_HOST=your-db-host
DB_PORT=5432
DB_DATABASE=myapp_production
DB_USERNAME=myapp
DB_PASSWORD=secure-password
```

---

## Docker

The nimbus-starter includes a production Dockerfile:

```dockerfile
# Build stage
FROM golang:1.22-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/server .

# Production stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata
WORKDIR /app

COPY --from=builder /app/server .
COPY --from=builder /app/resources ./resources
COPY --from=builder /app/public ./public

EXPOSE 8080

CMD ["./server"]
```

### Build & Run

```bash
# Build the image
docker build -t myapp .

# Run locally
docker run -p 8080:8080 \
  -e APP_ENV=production \
  -e DB_DRIVER=postgres \
  -e DB_HOST=host.docker.internal \
  -e DB_DATABASE=myapp \
  myapp

# Docker Compose
docker-compose up -d
```

### Docker Compose

```yaml
# docker-compose.yml
version: '3.8'

services:
  app:
    build: .
    ports:
      - "8080:8080"
    environment:
      - APP_ENV=production
      - DB_DRIVER=postgres
      - DB_HOST=db
      - DB_PORT=5432
      - DB_DATABASE=myapp
      - DB_USERNAME=postgres
      - DB_PASSWORD=secret
      - REDIS_URL=redis://redis:6379
    depends_on:
      - db
      - redis

  db:
    image: postgres:16-alpine
    environment:
      POSTGRES_DB: myapp
      POSTGRES_PASSWORD: secret
    volumes:
      - postgres_data:/var/lib/postgresql/data

  redis:
    image: redis:7-alpine
    volumes:
      - redis_data:/data

volumes:
  postgres_data:
  redis_data:
```

---

## Nimbus Forge (CLI Deploy)

Deploy directly from the command line:

### Initialize

```bash
nimbus deploy:init
```

This creates `deploy.yaml`:

```yaml
# deploy.yaml
app: myapp
region: us-east-1
platform: fly     # fly | railway | render | aws | gcp
instances: 1
memory: 256       # MB
env:
  APP_ENV: production
  PORT: "8080"
```

### Deploy

```bash
# Deploy to configured platform
nimbus deploy

# Or use the alias
nimbus forge
```

### Manage Deployments

```bash
# Check status
nimbus deploy:status

# View logs
nimbus deploy:logs

# Manage environment variables
nimbus deploy:env set DATABASE_URL=postgres://...
nimbus deploy:env list

# Rollback to previous version
nimbus deploy:rollback
```

---

## Platform-Specific Guides

### Fly.io

```bash
# Install flyctl
brew install flyctl

# Initialize
fly launch

# Deploy
fly deploy

# Set secrets
fly secrets set APP_KEY=your-secret-key
fly secrets set DATABASE_URL=postgres://...
```

**fly.toml:**

```toml
app = "myapp"
primary_region = "iad"

[build]
  dockerfile = "Dockerfile"

[env]
  PORT = "8080"
  APP_ENV = "production"

[http_service]
  internal_port = 8080
  force_https = true
  auto_stop_machines = true
  auto_start_machines = true

[[services]]
  internal_port = 8080
  protocol = "tcp"

  [[services.ports]]
    port = 80
    handlers = ["http"]

  [[services.ports]]
    port = 443
    handlers = ["tls", "http"]

  [services.concurrency]
    type = "connections"
    hard_limit = 25
    soft_limit = 20
```

### Railway

```bash
# Install Railway CLI
npm i -g @railway/cli

# Login & deploy
railway login
railway init
railway up
```

**railway.json:**

```json
{
  "$schema": "https://railway.app/railway.schema.json",
  "build": {
    "builder": "DOCKERFILE",
    "dockerfilePath": "Dockerfile"
  },
  "deploy": {
    "restartPolicyType": "ON_FAILURE",
    "restartPolicyMaxRetries": 10
  }
}
```

### Render

The nimbus-starter includes a `render.yaml` blueprint:

```yaml
# render.yaml
services:
  - type: web
    name: myapp
    env: docker
    plan: starter
    healthCheckPath: /health
    envVars:
      - key: APP_ENV
        value: production
      - key: PORT
        value: "8080"
      - key: DATABASE_URL
        fromDatabase:
          name: myapp-db
          property: connectionString

databases:
  - name: myapp-db
    plan: starter
```

Deploy: Connect your GitHub repo to Render, and it auto-deploys on push.

### AWS (ECS/Fargate)

```bash
# Build and push to ECR
aws ecr get-login-password | docker login --username AWS --password-stdin $ECR_URL
docker build -t myapp .
docker tag myapp:latest $ECR_URL/myapp:latest
docker push $ECR_URL/myapp:latest

# Deploy via ECS
aws ecs update-service --cluster myapp --service myapp --force-new-deployment
```

### Google Cloud Run

```bash
# Build and push to Artifact Registry
gcloud builds submit --tag gcr.io/$PROJECT_ID/myapp

# Deploy
gcloud run deploy myapp \
  --image gcr.io/$PROJECT_ID/myapp \
  --platform managed \
  --region us-central1 \
  --allow-unauthenticated \
  --set-env-vars "APP_ENV=production,PORT=8080"
```

### DigitalOcean App Platform

```yaml
# .do/app.yaml
name: myapp
services:
  - name: api
    dockerfile_path: Dockerfile
    instance_count: 1
    instance_size_slug: basic-xxs
    http_port: 8080
    health_check:
      http_path: /health
    envs:
      - key: APP_ENV
        value: production
      - key: DATABASE_URL
        scope: RUN_TIME
        type: SECRET
```

---

## Production Checklist

### Environment

- [ ] `APP_ENV=production`
- [ ] `APP_KEY` set to a strong, unique secret (32+ chars)
- [ ] All database credentials set
- [ ] Redis URL configured (if using cache/sessions/queues)

### Security

- [ ] HTTPS enabled (via reverse proxy or platform)
- [ ] CORS configured for your domain
- [ ] CSRF protection enabled
- [ ] Rate limiting configured
- [ ] Security headers via Shield middleware

### Performance

- [ ] Cache driver set to Redis (not memory)
- [ ] Session driver set to Redis or database
- [ ] Queue driver set to Redis
- [ ] Database connection pooling configured
- [ ] Static assets served via CDN

### Monitoring

- [ ] Health check endpoint at `/health`
- [ ] Logging configured (JSON output for log aggregation)
- [ ] Telescope enabled for debugging (restrict access in production)
- [ ] Error tracking (Sentry, Bugsnag) configured

### Database

- [ ] Migrations run: `./myapp migrate`
- [ ] Seeds run (if needed): `./myapp seed`
- [ ] Database backups configured
- [ ] Connection string uses SSL

### Process Management

- [ ] Scheduler running in separate process: `./myapp schedule:run`
- [ ] Queue worker running in separate process: `./myapp queue:work`
- [ ] Process manager (systemd, supervisor) for restart on crash

---

## Health Checks

```go
import "github.com/CodeSyncr/nimbus/health"

checker := health.New()
checker.DB(db)           // Database ping
checker.Redis(redisClient) // Redis ping

// Custom check
checker.Add("disk", func(ctx context.Context) error {
    var stat syscall.Statfs_t
    syscall.Statfs("/", &stat)
    availableGB := stat.Bavail * uint64(stat.Bsize) / 1e9
    if availableGB < 1 {
        return fmt.Errorf("low disk space: %dGB", availableGB)
    }
    return nil
})

// Mount handler
app.Get("/health", checker.Handler())
```

Response:

```json
{"status": "ok", "checks": {"database": "ok", "redis": "ok", "disk": "ok"}}
```

Or if degraded:

```json
{"status": "degraded", "checks": {"database": "ok", "redis": "error: connection refused"}}
```

---

## Best Practices

1. **Use multi-stage Docker builds** — Keep images small (<50MB)
2. **Set `CGO_ENABLED=0`** — Static binary, no system dependencies
3. **Use health checks** — Platforms need `/health` for load balancing
4. **Separate processes** — Web server, scheduler, queue worker in separate containers
5. **Use environment variables** — Never commit secrets to git
6. **Enable graceful shutdown** — Nimbus handles SIGTERM automatically
7. **Run migrations separately** — Don't auto-migrate in production
8. **Use managed databases** — AWS RDS, Cloud SQL, Render Postgres
9. **Set up CI/CD** — GitHub Actions → Docker build → Deploy

**Next:** [Advanced Features](21-advanced-features.md) →
