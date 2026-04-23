# Nimbus Framework Documentation

> **The Go framework for humans** — Build full-stack web applications with Laravel-inspired elegance and the performance of Go.

---

## Table of Contents

### Getting Started

| # | Topic | Description |
|---|-------|-------------|
| 01 | [Introduction](01-introduction.md) | Framework overview, philosophy, design principles, comparison with other frameworks |
| 02 | [Installation](02-installation.md) | Prerequisites, CLI install, project creation, first app walkthrough |
| 03 | [Folder Structure](03-folder-structure.md) | Deep dive into every directory and file in a Nimbus project |
| 04 | [Configuration](04-configuration.md) | Environment variables, config loading, type-safe config API |

### Core Concepts

| # | Topic | Description |
|---|-------|-------------|
| 05 | [Routing & Controllers](05-routing-controllers.md) | Routes, params, groups, ResourceController, HTTP context |
| 06 | [Database & ORM](06-database.md) | GORM integration, models, migrations, seeders, factories, scopes, hooks |
| 07 | [Middleware](07-middleware.md) | Three-layer middleware, built-in middleware, custom middleware |
| 08 | [Validation](08-validation.md) | VineJS-style validation, schemas, rules, form requests, database rules |
| 09 | [Auth & Security](09-auth-security.md) | Session/token/stateless guards, policies, Shield, CSRF, CORS, hashing |
| 10 | [Views & Templates](10-views-templates.md) | Edge-inspired .nimbus templates, layouts, components, partials |

### Framework Features

| # | Topic | Description |
|---|-------|-------------|
| 11 | [Plugin System](11-plugins.md) | Plugin architecture, lifecycle, built-in plugins, creating plugins |
| 12 | [Queue & Jobs](12-queue-jobs.md) | Background jobs, queue drivers, dispatching, Horizon dashboard |
| 13 | [Cache](13-cache.md) | Multi-driver caching, Remember, namespaces, type-safe generics |
| 14 | [Mail](14-mail.md) | Email sending, SMTP drivers, provider support |
| 15 | [Scheduler](15-scheduler.md) | Cron-like task scheduling, intervals, CLI integration |

### AI & Modern Features

| # | Topic | Description |
|---|-------|-------------|
| 16 | [AI SDK](16-ai-sdk.md) | Multi-provider AI (OpenAI, Claude, Gemini, Ollama), text generation |
| 17 | [MCP](17-mcp.md) | Model Context Protocol servers, tools, resources |

### Tooling

| # | Topic | Description |
|---|-------|-------------|
| 18 | [CLI](18-cli.md) | Commands, generators, AI copilot, test generator |
| 19 | [Testing](19-testing.md) | HTTP tests, database testing, factories, AI test generation |
| 20 | [Deployment](20-deployment.md) | Docker, Fly.io, Railway, Render, AWS, GCP, production checklist |

### Advanced

| # | Topic | Description |
|---|-------|-------------|
| 21 | [Advanced Features](21-advanced-features.md) | OpenAPI, Studio, Workflows, Feature Flags, Multi-tenancy, WebSockets, Events, Storage, Sessions, Logging, Metrics |

---

## Quick Links

### Create a New Project

```bash
go install github.com/CodeSyncr/nimbus/cmd/nimbus@latest
nimbus new myapp
cd myapp
nimbus serve
```

### Essential Commands

```bash
nimbus serve                    # Start dev server with hot reload
nimbus make:model User          # Generate a model
nimbus make:controller Product  # Generate a controller
nimbus make:migration add_users # Generate a migration
nimbus db:migrate               # Run migrations
nimbus db:seed                  # Seed database
nimbus deploy                   # Deploy to production
nimbus ai "create a blog API"   # AI code generation
nimbus test:generate            # AI test generation
```

### Tech Stack

| Component | Library |
|-----------|---------|
| HTTP Router | [chi](https://github.com/go-chi/chi) |
| ORM | [GORM](https://gorm.io) |
| CLI | [Cobra](https://github.com/spf13/cobra) |
| Logger | [Zap](https://github.com/uber-go/zap) |
| WebSocket | [Gorilla WebSocket](https://github.com/gorilla/websocket) |
| Templates | Custom Edge-inspired engine |
| Validation | Custom VineJS-inspired rules |
| Cache | Memory, Redis, Memcached, DynamoDB, Cloudflare KV |
| Session | Memory, Cookie (AES-256), Redis, Database |
| Queue | Memory, Redis |
| Mail | SMTP (SES, Mailgun, SendGrid, Postmark) |

---

## Architecture Overview

```
Request → Router → Middleware Chain → Controller → Response
                                         ↓
                        Models ←→ Database (GORM)
                        Cache  ←→ Redis/Memory
                        Queue  ←→ Background Jobs
                        Views  ←→ .nimbus Templates
                        Events ←→ Listeners
```

---

## Contributing

1. Fork the repository
2. Create your feature branch: `git checkout -b feature/amazing`
3. Commit your changes: `git commit -m 'Add amazing feature'`
4. Push to the branch: `git push origin feature/amazing`
5. Open a Pull Request

---

## License

MIT License — see [LICENSE](../LICENSE) for details.
