# Folder Structure

> **A guided tour of every directory and file in a Nimbus project** — understanding the conventions that make your code organized and maintainable.

---

## Overview

Nimbus follows a **convention-over-configuration** directory structure inspired by Laravel. Every file has a designated home, making it easy for teams to navigate projects they didn't write.

```
my-app/
├── main.go                      ← Entry point
├── go.mod                       ← Go module definition
├── .env                         ← Environment variables
├── .air.toml                    ← Hot reload config
├── Dockerfile                   ← Production container
├── deploy.yaml / render.yaml    ← Deployment configs
│
├── app/                         ← YOUR application code
│   ├── controllers/             ← HTTP request handlers
│   ├── models/                  ← Database models (GORM structs)
│   ├── middleware/              ← Custom middleware
│   ├── validators/              ← Request validation schemas
│   ├── jobs/                    ← Background queue jobs
│   ├── mcp/                     ← MCP server definitions
│   ├── docs/                    ← Documentation context
│   └── plugins/                 ← Custom plugins
│
├── bin/                         ← Bootstrap logic
│   └── server.go                ← Boot() — wires the app together
│
├── config/                      ← Configuration structs
│   ├── config.go                ← Master loader
│   ├── app.go                   ← App name, port, env
│   ├── database.go              ← DB connection
│   ├── auth.go                  ← Authentication
│   ├── cache.go                 ← Cache
│   ├── session.go               ← Session
│   ├── cors.go, hash.go, ...    ← Feature-specific configs
│   └── env.go                   ← Env helper functions
│
├── database/
│   ├── migrations/              ← Schema migrations
│   │   └── registry.go          ← Migration list
│   └── seeders/                 ← Test data seeders
│
├── start/                       ← Application hooks
│   ├── routes.go                ← All route definitions
│   ├── kernel.go                ← Middleware registration
│   ├── jobs.go                  ← Queue job registration
│   └── schedule.go              ← Scheduled tasks
│
├── resources/
│   ├── views/                   ← .nimbus templates
│   ├── css/                     ← Source stylesheets
│   └── js/                      ← Source JavaScript
│
├── public/                      ← Publicly served static files
│   ├── css/                     ← Compiled CSS
│   └── js/                      ← Compiled JS
│
└── storage/                     ← App-generated files
    ├── logs/                    ← Log files
    └── uploads/                 ← User uploads
```

---

## Directory Deep Dive

### `main.go` — The Entry Point

This is the simplest file in your project. It boots the application and starts the HTTP server, plus handles CLI subcommands:

```go
package main

import (
    "fmt"
    "os"
    "nimbus-starter/bin"
)

func main() {
    if len(os.Args) > 1 {
        switch os.Args[1] {
        case "migrate":
            bin.RunMigrations()
            return
        case "seed":
            bin.RunSeeders()
            return
        case "schedule:run":
            bin.RunSchedule()
            return
        case "schedule:list":
            bin.RunScheduleList()
            return
        }
    }

    app := bin.Boot()
    app.Run()
}
```

**Convention:** Never put business logic in `main.go`. It's purely a dispatcher.

---

### `app/` — Your Application Code

This is where all your business logic lives. Each subdirectory has a clear responsibility:

#### `app/controllers/` — HTTP Handlers

Controllers handle HTTP requests and return responses. They should be thin — delegate business logic to models or services.

```go
// app/controllers/todo.go
package controllers

type Todo struct {
    DB *nimbus.DB
}

func (todo *Todo) Index(ctx *http.Context) error {
    var items []models.Todo
    todo.DB.Find(&items)
    return ctx.View("apps/todo/index", map[string]any{
        "title": "Todo",
        "items": items,
    })
}
```

**Real-life example:** An e-commerce app might have:
```
app/controllers/
├── product.go          # Product CRUD
├── cart.go             # Shopping cart
├── checkout.go         # Payment flow
├── order.go            # Order management
├── user.go             # User profile
└── admin/
    ├── dashboard.go    # Admin dashboard
    └── reports.go      # Sales reports
```

#### `app/models/` — Database Models

GORM model structs with relationships, hooks, and business methods:

```go
// app/models/todo.go
package models

import "github.com/CodeSyncr/nimbus/database"

type Todo struct {
    database.Model        // ID, CreatedAt, UpdatedAt, DeletedAt
    Title string
    Done  bool
}
```

**Real-life example — E-commerce models:**
```go
// app/models/product.go
type Product struct {
    database.Model
    Name        string
    Price       float64
    SKU         string   `gorm:"uniqueIndex"`
    CategoryID  uint
    Category    Category
    InStock     bool
}

// app/models/order.go
type Order struct {
    database.Model
    UserID     uint
    User       User
    Items      []OrderItem
    Total      float64
    Status     string // pending, paid, shipped, delivered
    PaidAt     *time.Time
}
```

#### `app/validators/` — Request Validation

VineJS-style validation schemas:

```go
// app/validators/todo.go
package validators

import "github.com/CodeSyncr/nimbus/validation"

type Todo struct {
    Title   string
    Content string
}

func (v *Todo) Rules() validation.Schema {
    return validation.Schema{
        "title": validation.String().Required().Min(1).Max(255).Trim(),
    }
}

func (v *Todo) Validate() error {
    return validation.ValidateStruct(v)
}
```

#### `app/jobs/` — Background Jobs

Jobs are dispatched to queues and processed asynchronously:

```go
// app/jobs/send_welcome_email.go
package jobs

type SendWelcomeEmail struct {
    UserID uint
    Email  string
}

func (j *SendWelcomeEmail) Handle(ctx context.Context) error {
    fmt.Printf("[queue] sending welcome email to %s\n", j.Email)
    return nil
}

func (j *SendWelcomeEmail) Failed(ctx context.Context, err error) {
    fmt.Printf("[queue] FAILED welcome email to %s: %v\n", j.Email, err)
}
```

#### `app/middleware/` — Custom Middleware

Application-specific middleware:

```go
// app/middleware/admin.go
package middleware

func RequireAdmin() router.Middleware {
    return func(next router.HandlerFunc) router.HandlerFunc {
        return func(c *http.Context) error {
            user := auth.UserFromContext(c.Request.Context())
            if user == nil || !user.IsAdmin {
                return c.JSON(http.StatusForbidden, map[string]string{"error": "forbidden"})
            }
            return next(c)
        }
    }
}
```

#### `app/plugins/` — Custom Plugins

Your own plugins following the Nimbus plugin interface:

```go
// app/plugins/analytics/analytics.go
package analytics

type Plugin struct {
    nimbus.BasePlugin
    hits int64
}

func New() *Plugin { return &Plugin{} }
func (p *Plugin) Name() string    { return "analytics" }
func (p *Plugin) Version() string { return "1.0.0" }
```

#### `app/mcp/` — MCP Servers

Model Context Protocol servers for AI tool integration:

```go
// app/mcp/weather_server.go
var WeatherServer = nimbusmcp.NewServer("Weather Demo", "1.0.0")

func init() {
    WeatherServer.AddTool(
        mcp.NewTool("get_weather", mcp.WithDescription("Get weather for a location")),
        handleGetWeather,
    )
}
```

---

### `bin/server.go` — The Bootstrap

This is the **wiring file** — it connects all the pieces together. Think of it as your application's composition root:

```go
package bin

func Boot() *nimbus.App {
    loadConfig()           // 1. Load .env and config structs
    app := newApp()        // 2. Create the Nimbus app

    bootMail()             // 3. Configure mail
    bootCache()            // 4. Initialize cache
    bootDatabase(app)      // 5. Connect database
    bootQueue()            // 6. Initialize queue

    registerPlugins(app)   // 7. Register plugins (Horizon, Shield, AI, Telescope, MCP, ...)
    registerMiddleware(app)// 8. Apply middleware stack
    registerRoutes(app)    // 9. Define routes

    return app
}
```

**Convention:** Keep `Boot()` as a configuration orchestrator. Business logic belongs in `app/`.

---

### `config/` — Configuration

Each config file defines a struct and a loader function. The master `config.go` calls all loaders:

```go
// config/config.go
func Load() {
    _ = nimbusconfig.LoadAuto()  // Load .env file

    loadApp()         // App name, port, env
    loadDatabase()    // DB driver, DSN
    loadQueue()       // Queue driver
    loadAuth()        // Auth guard config
    loadBodyParser()  // Request body limits
    loadCache()       // Cache driver, TTL
    loadCORS()        // CORS origins
    loadHash()        // Bcrypt rounds
    loadLimiter()     // Rate limit settings
    loadLogger()      // Log level, format
    loadMail()        // SMTP config
    loadSession()     // Session driver, cookie
    loadStatic()      // Static file serving
    loadStorage()     // S3/local storage
}
```

**Real-life example — Adding Redis config:**
```go
// config/redis.go
package config

var Redis RedisConfig

type RedisConfig struct {
    Host     string
    Port     int
    Password string
    DB       int
}

func loadRedis() {
    Redis = RedisConfig{
        Host:     env("REDIS_HOST", "localhost"),
        Port:     envInt("REDIS_PORT", 6379),
        Password: env("REDIS_PASSWORD", ""),
        DB:       envInt("REDIS_DB", 0),
    }
}
```

---

### `start/` — Application Hooks

These files are called during boot and define the application's behavior:

| File | Purpose | Called By |
|------|---------|-----------|
| `routes.go` | Define all HTTP routes | `bin/server.go → registerRoutes()` |
| `kernel.go` | Register middleware stack | `bin/server.go → registerMiddleware()` |
| `jobs.go` | Register queue job types | `bin/server.go → bootQueue()` |
| `schedule.go` | Define cron-like scheduled tasks | `bin.RunSchedule()` |

---

### `database/` — Migrations & Seeders

#### Migrations

```go
// database/migrations/registry.go
package migrations

import "github.com/CodeSyncr/nimbus/database"

func All() []database.Migration {
    return []database.Migration{
        {
            Name: "001_create_todos_table",
            Up: func(db *gorm.DB) error {
                return db.AutoMigrate(&models.Todo{})
            },
            Down: func(db *gorm.DB) error {
                return db.Migrator().DropTable("todos")
            },
        },
    }
}
```

#### Seeders

```go
// database/seeders/seeders.go
package seeders

func All() []database.Seeder {
    return []database.Seeder{
        database.SeedFunc(func(db *gorm.DB) error {
            return db.Create(&models.Todo{Title: "Learn Nimbus", Done: false}).Error
        }),
    }
}
```

---

### `resources/views/` — Templates

Nimbus templates use the `.nimbus` extension with Edge-inspired syntax:

```html
{{-- resources/views/home.nimbus --}}
@layout('layout')

<div class="container">
    <h1>{{ .title }}</h1>
    <p>Welcome to {{ .appName }}</p>

    @if(.items)
        @each(item in .items)
            <div class="card">{{ .item.Title }}</div>
        @endeach
    @else
        <p>No items yet.</p>
    @endif
</div>
```

---

### `public/` — Static Assets

Files in `public/` are served directly by the HTTP server. No processing or compilation.

```
public/
├── css/
│   └── app.css         # Your compiled CSS
├── js/
│   └── app.js          # Your compiled JS
├── images/
│   └── logo.png
└── favicon.ico
```

Access: `http://localhost:3333/public/css/app.css`

---

### `storage/` — Generated Files

App-generated files that shouldn't be committed to version control:

```
storage/
├── logs/
│   └── app.log         # Application logs
├── uploads/            # User file uploads
├── cache/              # File-based cache (if configured)
└── temp/               # Temporary files
```

**Convention:** Add `storage/` to `.gitignore` (except `storage/.gitkeep`).

---

## File Naming Conventions

| Type | Convention | Example |
|------|-----------|---------|
| Controllers | PascalCase, singular noun | `todo.go`, `user_profile.go` |
| Models | PascalCase, singular noun | `todo.go`, `order_item.go` |
| Middleware | snake_case, descriptive | `require_auth.go`, `rate_limit.go` |
| Validators | Match model name | `todo.go`, `user.go` |
| Jobs | Descriptive action | `send_welcome_email.go`, `process_payment.go` |
| Migrations | Numbered prefix | `001_create_users.go`, `002_add_email_index.go` |
| Views | kebab-case | `user-profile.nimbus`, `order-details.nimbus` |
| Config | Feature name | `database.go`, `cache.go`, `mail.go` |

**Next:** [Configuration](04-configuration.md) →
