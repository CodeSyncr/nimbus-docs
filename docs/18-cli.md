# CLI

> **Full-featured command-line interface** — project scaffolding, code generators, database commands, deployment, AI assistance, and more.

---

## Introduction

The Nimbus CLI provides everything you need to build, develop, and deploy your application. Built on [Cobra](https://github.com/spf13/cobra) with interactive prompts via [Survey](https://github.com/AlecAivazis/survey) and beautiful output via [Lipgloss](https://github.com/charmbracelet/lipgloss).

Features:

- **Project scaffolding** — `nimbus new` creates a complete project with interactive wizard
- **Code generators** — `make:model`, `make:controller`, `make:migration`, and 10+ generators
- **Database commands** — `db:migrate`, `db:seed`, `db:rollback`, `db:create`
- **Development server** — `nimbus serve` with hot reload via Air
- **AI Copilot** — `nimbus ai` generates code from natural language
- **AI Test Generator** — `nimbus test:generate` creates tests from controller code
- **Deployment** — `nimbus deploy` to Fly.io, Railway, Render, AWS, GCP
- **Plugin management** — `plugin:install`, `plugin list`
- **Queue & Scheduler** — `queue:work`, `schedule:run`, `schedule:list`

---

## Installation

```bash
go install github.com/CodeSyncr/nimbus/cmd/nimbus@latest
```

Verify:

```bash
nimbus --version
```

---

## Complete Command Reference

### Project Commands

| Command | Alias | Description |
|---------|-------|-------------|
| `nimbus new [name]` | `create` | Create a new Nimbus project with interactive wizard |
| `nimbus serve` | | Start dev server with hot reload (Air) |
| `nimbus build` | | Build frontend assets (CSS/JS) to public/ |
| `nimbus repl` | | Start interactive REPL session |

### Code Generators

All generators follow the pattern `nimbus make:<type> <Name>`:

| Command | Creates | Location |
|---------|---------|----------|
| `make:model User` | GORM model with `database.Model` | `app/models/user.go` |
| `make:controller User` | HTTP controller struct | `app/controllers/user.go` |
| `make:migration create_users` | Timestamped migration file | `database/migrations/` |
| `make:middleware Auth` | Middleware function | `app/middleware/auth.go` |
| `make:job SendEmail` | Queue job with Handle/Failed | `app/jobs/send_email.go` |
| `make:seeder UserSeeder` | Database seeder | `database/seeders/user_seeder.go` |
| `make:validator User` | Validation schema with rules | `app/validators/user.go` |
| `make:command Greet` | Custom CLI command | `app/commands/greet.go` |
| `make:plugin Analytics` | Full plugin skeleton (9 files) | `app/plugins/analytics/` |

### Auth Scaffolding

| Command | Description |
|---------|-------------|
| `make:auth` | Scaffold complete auth system (model, controller, views, routes) |
| `make:api-token` | Scaffold API token auth (migration + controller) |

### Database Commands

| Command | Description |
|---------|-------------|
| `db:create` | Create the database |
| `db:migrate` | Run pending migrations |
| `db:rollback` | Rollback the last migration batch |
| `db:seed` | Run database seeders |

### Queue & Scheduler

| Command | Description |
|---------|-------------|
| `queue:work` | Start processing queue jobs |
| `schedule:run` | Start the scheduler (blocks) |
| `schedule:list` | List all scheduled tasks |

### Deploy (Forge)

| Command | Description |
|---------|-------------|
| `deploy` / `forge` | Deploy to production |
| `deploy:init` | Initialize deployment configuration |
| `deploy:status` | Check deployment status |
| `deploy:logs` | View deployment logs |
| `deploy:env` | Manage deployment environment variables |
| `deploy:rollback` | Rollback to previous deployment |
| `make:deploy-config` | Generate deploy.yaml |

### Plugins

| Command | Description |
|---------|-------------|
| `plugin:install [name]` | Install a Nimbus plugin |
| `plugin list` | List available plugins |

Available plugins: `telescope`, `horizon`, `inertia`, `unpoly`, `ai`, `mcp`, `drive`, `transmit`, `scout`, `pulse`, `socialite`

When you install a plugin, the CLI automatically:

1. Runs `go get` to download the package
2. Patches `bin/server.go` to register the plugin
3. Patches `start/kernel.go` (if needed) for middleware
4. Adds environment variables to `.env.example`
5. **Scaffolds a config file** in `config/` (if the plugin has one)
6. **Patches `config/config.go`** to add the `loadXxx()` call

#### Plugin Config Files

Plugins that scaffold config files on install:

| Plugin | Config File | Loader Added |
|--------|-------------|--------------|
| `telescope` | `config/telescope.go` | `loadTelescope()` |
| `horizon` | `config/horizon.go` | `loadHorizon()` |
| `transmit` | `config/transmit.go` | `loadTransmit()` |
| `socialite` | `config/socialite.go` | (provider helper) |

```bash
# Example: Install Telescope
nimbus plugin:install telescope

# Output:
# ✓ bin/server.go updated
# ✓ start/kernel.go updated
# ✓ .env.example updated
# ✓ config/telescope.go created
# ✓ config/config.go updated
# ✓ Plugin "telescope" installed successfully.
```

### Horizon

| Command | Description |
|---------|-------------|
| `horizon:forget` | Forget completed/failed jobs |
| `horizon:clear` | Clear all jobs from queue |

### AI Commands

| Command | Alias | Description |
|---------|-------|-------------|
| `nimbus ai "<description>"` | | AI Copilot — generate code from natural language |
| `nimbus test:generate` | `test:gen`, `tg` | AI Test Generator — create tests from controllers |

---

## Project Scaffolding

### Interactive Wizard

```bash
nimbus new myapp
```

The wizard prompts for:

1. **Project name** — Used for module path and directory
2. **Database driver** — PostgreSQL, MySQL, or SQLite
3. **Features** — Select from: Auth, API, Queue, Scheduler, AI, MCP
4. **Frontend** — Tailwind CSS, plain CSS, or none

This generates a complete project with:

```
myapp/
├── main.go
├── go.mod
├── .env
├── .env.example
├── .air.toml
├── app/
│   ├── controllers/
│   ├── models/
│   ├── middleware/
│   ├── validators/
│   └── jobs/
├── config/                    # ← Full config directory (19 files)
│   ├── config.go              # Master loader — calls all loadXxx() functions
│   ├── env.go                 # env(), envInt(), envBool() helpers
│   ├── app.go                 # App name, env, port, host, key
│   ├── database.go            # DB driver, DSN, connection fields
│   ├── bodyparser.go          # JSON/form/multipart size limits
│   ├── cache.go               # Cache driver + TTL
│   ├── cors.go                # CORS origins, methods, credentials
│   ├── hash.go                # Bcrypt driver/cost
│   ├── limiter.go             # Rate limiting rules
│   ├── logger.go              # Log level + format
│   ├── mail.go                # SMTP driver settings
│   ├── queue.go               # Queue driver (sync/redis/sqs/kafka)
│   ├── session.go             # Session driver, cookie settings
│   ├── shield.go              # Security headers + CSRF
│   ├── static.go              # Static file serving
│   └── storage.go             # File storage driver
├── database/
│   ├── migrations/
│   └── seeders/
├── resources/views/
├── public/
├── start/
│   ├── routes.go
│   ├── kernel.go
│   ├── jobs.go
│   └── schedule.go
└── bin/
    └── server.go
```

Every config file follows the same pattern: a typed struct, sensible defaults, and environment variable overrides — aligned with Laravel-style conventions.

> **Note:** Plugin-specific configs (telescope, horizon, transmit, socialite) are **not** included by default. They are scaffolded automatically when you run `nimbus plugin:install <name>`.

---

## Code Generator Details

### make:model

```bash
nimbus make:model Product
```

Generates:

```go
// app/models/product.go
package models

import "github.com/CodeSyncr/nimbus/database"

type Product struct {
    database.Model
}
```

### make:controller

```bash
nimbus make:controller Product
```

Generates a controller with `ResourceController` interface stubs:

```go
// app/controllers/product.go
package controllers

import "github.com/CodeSyncr/nimbus/http"

type ProductController struct{}

func (c *ProductController) Index(ctx *http.Context) error { return nil }
func (c *ProductController) Create(ctx *http.Context) error { return nil }
func (c *ProductController) Store(ctx *http.Context) error { return nil }
func (c *ProductController) Show(ctx *http.Context) error { return nil }
func (c *ProductController) Edit(ctx *http.Context) error { return nil }
func (c *ProductController) Update(ctx *http.Context) error { return nil }
func (c *ProductController) Destroy(ctx *http.Context) error { return nil }
```

### make:migration

```bash
nimbus make:migration create_products_table
```

Generates a timestamped migration:

```go
// database/migrations/20250101120000_create_products_table.go
package migrations

import "github.com/CodeSyncr/nimbus/lucid"

func init() {
    Register("20250101120000_create_products_table", func(db *lucid.DB) error {
        // Up
        return db.AutoMigrate(&Product{})
    }, func(db *lucid.DB) error {
        // Down
        return db.Migrator().DropTable("products")
    })
}
```

### make:plugin

```bash
nimbus make:plugin Analytics
```

Generates a complete plugin skeleton with 9 files:

```
app/plugins/analytics/
├── plugin.go          # Main plugin with all interfaces
├── config.go          # Plugin configuration
├── middleware.go       # Plugin middleware
├── routes.go          # Plugin routes
├── handlers.go        # HTTP handlers
├── store.go           # Data store
├── models.go          # Database models
├── views/             # Plugin views
│   └── dashboard.nimbus
└── README.md          # Plugin documentation
```

---

## AI Copilot

Generate code from natural language descriptions:

```bash
# Generate a controller
nimbus ai "create a product controller with CRUD operations and image upload"

# Generate a model
nimbus ai "create a blog post model with title, content, slug, published_at, and author relationship"

# Generate middleware
nimbus ai "create rate limiting middleware that limits to 100 requests per minute per IP"

# Generate a complete feature
nimbus ai "add a comment system to the blog with nested replies and spam detection"
```

The AI Copilot:
1. Parses your natural language description
2. Determines what files to create
3. Generates idiomatic Nimbus code
4. Writes files to the correct locations
5. Provides setup instructions

---

## AI Test Generator

Automatically generate tests from your controller code:

```bash
nimbus test:generate
# or
nimbus tg
```

This:
1. Scans `app/controllers/` for controller files
2. Analyzes each controller's methods
3. Generates test files with:
   - HTTP test setup and teardown
   - Test cases for success and error paths
   - Request body examples
   - Response validation
   - Edge case coverage

---

## Development Server

```bash
nimbus serve
```

Features:
- **Hot reload** — Restarts on file changes using [Air](https://github.com/cosmtrek/air)
- **Colored output** — Pretty-printed logs with timestamps
- **Error display** — Compile errors shown with source context
- **Port selection** — Uses `PORT` from `.env` or defaults to `3000`

---

## Best Practices

1. **Use generators for consistency** — `make:*` commands follow project conventions
2. **Name things correctly** — Use PascalCase for models and controllers (e.g., `make:model UserProfile`)
3. **Run migrations after generating** — `nimbus db:migrate` applies new migrations
4. **Use AI copilot for boilerplate** — Let AI generate the scaffolding, then customize
5. **Check `plugin list` before installing** — See what's available
6. **Use `deploy:init` first** — Configure before deploying

**Next:** [Testing](19-testing.md) →
