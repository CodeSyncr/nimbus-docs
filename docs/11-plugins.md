# Plugin System

> **Modular, capability-based plugin architecture** — extend your application with reusable, self-contained modules that integrate naturally into the framework lifecycle.

---

## Introduction

Nimbus's plugin system is one of its most powerful features. Instead of scattering configuration across multiple files, plugins are self-contained modules that can:

- Register **routes** and **middleware**
- Provide **configuration** defaults
- Add **services** to the IoC container (bindings)
- Register **CLI commands**
- Define **scheduled tasks**
- Emit and listen for **events**
- Add **health checks**
- Run **database migrations**
- Provide **view templates**
- Execute **graceful shutdown** logic

All through a simple, opt-in interface system — implement only the capabilities you need.

---

## The Plugin Interface

Every plugin must implement the base `Plugin` interface:

```go
type Plugin interface {
    Name() string              // Unique plugin identifier
    Version() string           // Semantic version (e.g., "1.0.0")
    Register(app *App)         // Called during app registration phase
    Boot(app *App)             // Called after all plugins are registered
}
```

### Lifecycle

```
For each plugin:
  1. Register(app)        → Register services, bindings
  2. Boot(app)           → Access other plugins' services, finalize setup

After all plugins:
  3. Apply capabilities   → Routes, middleware, commands, schedule, events, health
  4. App starts           → HTTP server, scheduler, signal handlers
  5. Shutdown             → HasShutdown.Shutdown() called in reverse order
```

---

## Capability Interfaces

Plugins opt into capabilities by implementing additional interfaces:

| Interface | Method | Purpose |
|-----------|--------|---------|
| `HasRoutes` | `Routes(router)` | Register HTTP routes |
| `HasMiddleware` | `Middleware() []Middleware` | Add global middleware |
| `HasConfig` | `DefaultConfig() map[string]any` | Provide default configuration |
| `HasBindings` | `Bindings(container)` | Register services in IoC container |
| `HasCommands` | `Commands() []*cobra.Command` | Add CLI commands |
| `HasSchedule` | `Schedule(scheduler)` | Define scheduled tasks |
| `HasEvents` | `Events(events)` | Register event listeners |
| `HasHealthChecks` | `HealthChecks() []HealthCheck` | Add health check probes |
| `HasMigrations` | `Migrations() []Migration` | Provide database migrations |
| `HasViews` | `Views() fs.FS` | Provide template files |
| `HasShutdown` | `Shutdown(ctx)` | Graceful shutdown logic |

---

## Creating a Plugin

### Step 1: Define the Plugin Struct

```go
// app/plugins/analytics/analytics.go
package analytics

import (
    "fmt"
    "sync/atomic"
    "time"
    
    "github.com/CodeSyncr/nimbus"
    "github.com/CodeSyncr/nimbus/http"
    "github.com/CodeSyncr/nimbus/router"
)

type Plugin struct {
    nimbus.BasePlugin     // Provides no-op implementations
    totalRequests int64
    startTime     time.Time
}

func New() *Plugin {
    return &Plugin{
        startTime: time.Now(),
    }
}

func (p *Plugin) Name() string    { return "analytics" }
func (p *Plugin) Version() string { return "1.0.0" }
```

### Step 2: Implement Registration

```go
func (p *Plugin) Register(app *nimbus.App) {
    // Register services, prepare state
    fmt.Println("[analytics] Plugin registered")
}

func (p *Plugin) Boot(app *nimbus.App) {
    // Access other plugins, finalize setup
    fmt.Println("[analytics] Plugin booted")
}
```

### Step 3: Add Capabilities

```go
// HasRoutes — Add plugin-specific routes
func (p *Plugin) Routes(r *router.Router) {
    r.Get("/__analytics", func(c *http.Context) error {
        return c.JSON(http.StatusOK, map[string]any{
            "total_requests": atomic.LoadInt64(&p.totalRequests),
            "uptime":         time.Since(p.startTime).String(),
        })
    })
}

// HasMiddleware — Add request-counting middleware
func (p *Plugin) Middleware() []router.Middleware {
    return []router.Middleware{
        func(next router.HandlerFunc) router.HandlerFunc {
            return func(c *http.Context) error {
                atomic.AddInt64(&p.totalRequests, 1)
                return next(c)
            }
        },
    }
}

// HasHealthChecks — Report analytics health
func (p *Plugin) HealthChecks() []nimbus.HealthCheck {
    return []nimbus.HealthCheck{
        {
            Name: "analytics",
            Check: func() error {
                if atomic.LoadInt64(&p.totalRequests) < 0 {
                    return fmt.Errorf("negative request count")
                }
                return nil
            },
        },
    }
}
```

### Step 4: Register the Plugin

```go
// bin/server.go
func registerPlugins(app *nimbus.App) {
    app.Use(
        analytics.New(),
        // ... other plugins
    )
}
```

---

## Built-In Plugins

Nimbus ships with a rich set of built-in plugins:

### Installing Plugins via CLI

The recommended way to add a built-in plugin is via the CLI:

```bash
nimbus plugin:install telescope
```

This command automatically:

1. **Downloads the package** — runs `go get`
2. **Patches `bin/server.go`** — adds the import and `app.Use(...)` call
3. **Patches `start/kernel.go`** — adds middleware wiring (if needed)
4. **Updates `.env.example`** — adds required environment variables
5. **Scaffolds `config/<plugin>.go`** — creates a fully documented config file with typed structs and sensible defaults
6. **Patches `config/config.go`** — adds the `loadXxx()` call to the master loader

#### Plugins with Auto-Generated Config

| Plugin | Config File | Key Settings |
|--------|-------------|--------------|
| **Telescope** | `config/telescope.go` | `Enabled`, `Path`, `MaxEntries`, `Watchers` (per entry type) |
| **Horizon** | `config/horizon.go` | `Path`, `RedisURL`, `Defaults`, `Environments` (per-env supervisors) |
| **Transmit** | `config/transmit.go` | `Path`, `PingInterval`, `Transport`, `Redis` (for multi-instance) |
| **Socialite** | `config/socialite.go` | `SocialiteProviders()` — GitHub, Google, Discord, Apple |

```bash
# Example: Install Telescope
nimbus plugin:install telescope

# ✓ bin/server.go updated
# ✓ start/kernel.go updated
# ✓ .env.example updated
# ✓ config/telescope.go created
# ✓ config/config.go updated
# ✓ Plugin "telescope" installed successfully.
```

After installation, customize the generated config file to your needs — every field has sensible defaults that work out of the box.

### Telescope (Debug Dashboard)

Request monitoring, exception tracking, and query logging:

```go
import "github.com/CodeSyncr/nimbus/plugins/telescope"

app.Use(telescope.New())
// Dashboard available at /telescope
// Tracks: requests, exceptions, queries, cache, queue, mail, schedule
```

### Horizon (Queue Dashboard)

Monitor and manage background jobs:

```go
import "github.com/CodeSyncr/nimbus/plugins/horizon"

app.Use(horizon.New())
// Dashboard at /horizon
// Shows: jobs, failed jobs, throughput, wait times
```

### AI SDK

Multi-provider AI text generation:

```go
import "github.com/CodeSyncr/nimbus/plugins/ai"

app.Use(ai.New())

// Use in controllers
response, err := ai.Generate(ctx, "Summarize this article...",
    ai.WithProvider("openai"),
    ai.WithModel("gpt-4"),
)
```

### Shield

Security headers and protections (detailed in [Auth & Security](09-auth-security.md)):

```go
import "github.com/CodeSyncr/nimbus/packages/shield"

app.Use(shield.NewPlugin(shield.DefaultConfig()))
```

### Unpoly

Server-side integration with the Unpoly JavaScript framework:

```go
import "github.com/CodeSyncr/nimbus/plugins/unpoly"

app.Use(unpoly.New())

// In handlers
if unpoly.IsUnpoly(ctx) {
    unpoly.SetTitle(ctx, "Page Title")
    unpoly.EmitEvent(ctx, "item:created", data)
    unpoly.ExpireCache(ctx, "/items*")
}
```

### MCP (Model Context Protocol)

AI tool integration servers:

```go
import nimbusmcp "github.com/CodeSyncr/nimbus/plugins/mcp"

mcpPlugin := nimbusmcp.New()
mcpPlugin.Web("/mcp/weather", weatherServer)
app.Use(mcpPlugin)
```

### Studio (Admin Panel)

Auto-generated admin panel from GORM models:

```go
import "github.com/CodeSyncr/nimbus/studio"

app.Use(studio.New(studio.Config{
    Models: []any{&models.User{}, &models.Product{}, &models.Order{}},
}))
// Admin CRUD at /studio
```

---

## Real-Life Plugin Examples

### Example 1: Audit Log Plugin

Track all data changes for compliance:

```go
package audit

import (
    "encoding/json"
    "fmt"
    "time"
    
    "github.com/CodeSyncr/nimbus"
    "github.com/CodeSyncr/nimbus/auth"
    "github.com/CodeSyncr/nimbus/http"
    "github.com/CodeSyncr/nimbus/router"
)

type AuditEntry struct {
    ID        uint      `gorm:"primaryKey"`
    UserID    string
    Action    string    // create, update, delete
    Resource  string    // users, orders, etc.
    ResourceID string
    Changes   string    // JSON diff
    IP        string
    CreatedAt time.Time
}

type Plugin struct {
    nimbus.BasePlugin
    db *gorm.DB
}

func New() *Plugin { return &Plugin{} }
func (p *Plugin) Name() string    { return "audit" }
func (p *Plugin) Version() string { return "1.0.0" }

func (p *Plugin) Boot(app *nimbus.App) {
    p.db = app.Container.MustMake("db").(*nimbus.DB)
    p.db.AutoMigrate(&AuditEntry{})
}

func (p *Plugin) Routes(r *router.Router) {
    r.Get("/__audit", func(c *http.Context) error {
        var entries []AuditEntry
        p.db.Order("created_at DESC").Limit(100).Find(&entries)
        return c.JSON(http.StatusOK, entries)
    })
}

func (p *Plugin) Log(userID, action, resource, resourceID string, changes any) {
    changesJSON, _ := json.Marshal(changes)
    p.db.Create(&AuditEntry{
        UserID:     userID,
        Action:     action,
        Resource:   resource,
        ResourceID: resourceID,
        Changes:    string(changesJSON),
        CreatedAt:  time.Now(),
    })
}
```

### Example 2: Feature Flags Plugin

```go
package flags

import (
    "github.com/CodeSyncr/nimbus"
    "github.com/CodeSyncr/nimbus/http"
    "github.com/CodeSyncr/nimbus/router"
)

type Plugin struct {
    nimbus.BasePlugin
    flags map[string]bool
}

func New(flags map[string]bool) *Plugin {
    return &Plugin{flags: flags}
}

func (p *Plugin) Name() string    { return "feature-flags" }
func (p *Plugin) Version() string { return "1.0.0" }

func (p *Plugin) IsEnabled(flag string) bool {
    enabled, ok := p.flags[flag]
    return ok && enabled
}

func (p *Plugin) Bindings(c *nimbus.Container) {
    c.Instance("flags", p)
}

func (p *Plugin) Routes(r *router.Router) {
    r.Get("/__flags", func(c *http.Context) error {
        return c.JSON(http.StatusOK, p.flags)
    })
}

// Usage in controllers:
// flags := app.Container.MustMake("flags").(*flags.Plugin)
// if flags.IsEnabled("new_checkout") { ... }
```

### Example 3: Stripe Payments Plugin

```go
package payments

type Plugin struct {
    nimbus.BasePlugin
    stripeKey    string
    webhookSecret string
}

func New(key, webhookSecret string) *Plugin {
    return &Plugin{stripeKey: key, webhookSecret: webhookSecret}
}

func (p *Plugin) Name() string    { return "payments" }
func (p *Plugin) Version() string { return "1.0.0" }

func (p *Plugin) DefaultConfig() map[string]any {
    return map[string]any{
        "payments.currency": "usd",
        "payments.tax_rate": 0.0,
    }
}

func (p *Plugin) Bindings(c *nimbus.Container) {
    c.Singleton("payments", func() *Plugin { return p })
    c.Singleton("stripe", func() *stripe.Client {
        return stripe.New(p.stripeKey)
    })
}

func (p *Plugin) Routes(r *router.Router) {
    r.Post("/webhooks/stripe", p.HandleWebhook)
    r.Get("/checkout/:id", p.ShowCheckout)
    r.Post("/checkout/:id/pay", p.ProcessPayment)
}

func (p *Plugin) Schedule(s *scheduler.Scheduler) {
    s.EveryDay(func(ctx context.Context) error {
        return p.reconcilePayments(ctx)
    })
}

func (p *Plugin) HealthChecks() []nimbus.HealthCheck {
    return []nimbus.HealthCheck{
        {Name: "stripe", Check: p.pingStripe},
    }
}
```

---

## The BasePlugin

For convenience, `nimbus.BasePlugin` provides no-op implementations of all interface methods. Embed it and override only what you need:

```go
type MyPlugin struct {
    nimbus.BasePlugin
}

func (p *MyPlugin) Name() string    { return "my-plugin" }
func (p *MyPlugin) Version() string { return "1.0.0" }

// Only implement Routes — all other capabilities are no-ops
func (p *MyPlugin) Routes(r *router.Router) {
    r.Get("/my-endpoint", handler)
}
```

---

## Accessing Plugins

```go
// Get a plugin by name
plugin := app.Plugin("analytics")
if analytics, ok := plugin.(*analytics.Plugin); ok {
    count := analytics.TotalRequests()
}

// List all plugins
for _, p := range app.Plugins() {
    fmt.Printf("  %s v%s\n", p.Name(), p.Version())
}

// Access plugin config
config := app.PluginConfig("payments")
```

---

## Plugin Lifecycle Hooks

```go
// OnBoot — runs after all plugins are booted
app.OnBoot(func(app *nimbus.App) {
    fmt.Println("All plugins are ready!")
})

// OnStart — runs when the HTTP server starts
app.OnStart(func(app *nimbus.App) {
    fmt.Println("Server is listening!")
})

// OnShutdown — runs during graceful shutdown
app.OnShutdown(func(app *nimbus.App) {
    fmt.Println("Cleaning up...")
})
```

---

## Best Practices

1. **One responsibility per plugin** — Keep plugins focused and composable
2. **Use `BasePlugin`** — Embed it to avoid implementing unused interfaces
3. **Register services in `Bindings()`** — Make them available via the container
4. **Set defaults in `DefaultConfig()`** — Allow override via `.env`
5. **Add health checks** — Monitor critical dependencies
6. **Implement `HasShutdown`** — Clean up connections, flush buffers
7. **Namespace your routes** — Use a prefix like `/__myplugin` for admin routes
8. **Version your plugin** — Follow semver for breaking changes

**Next:** [Queue & Jobs](12-queue-jobs.md) →
