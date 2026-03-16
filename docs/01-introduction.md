# Introduction to Nimbus

> **The Laravel/AdonisJS-style web framework for Go** — bringing developer joy, convention-over-configuration, and batteries-included architecture to the Go ecosystem.

---

## What is Nimbus?

Nimbus is a **full-stack web framework for Go** that brings the developer experience of frameworks like Laravel (PHP), AdonisJS (Node.js), and Ruby on Rails to the Go ecosystem. It provides an opinionated, batteries-included foundation for building modern web applications, APIs, and microservices — without sacrificing Go's performance, type safety, or simplicity.

Unlike minimalist Go frameworks that leave you assembling middleware stacks, ORM wrappers, and CLI scaffolding from scratch, Nimbus gives you **everything out of the box**:

- A powerful **Chi-based router** with resource controllers and named routes
- A full **GORM-powered ORM** with migrations, seeders, factories, and query scopes
- A **VineJS-inspired validation system** with chainable rules
- A **plugin architecture** with 12 capability interfaces
- Built-in **authentication** (session + token guards), **authorization** (policies), and **security** (Shield, CSRF, CORS, rate limiting)
- **Template engine** with layouts, components, slots, and partials (`.nimbus` files)
- **Queue system**, **scheduler**, **cache** (memory/Redis/Memcached/DynamoDB), **mail**, **events**, **storage** (S3/GCS/R2)
- **CLI code generators** for models, controllers, migrations, middleware, and more
- **AI SDK** with multi-provider support (OpenAI, Anthropic, Gemini, Cohere, Mistral, Ollama)
- **MCP (Model Context Protocol)** server integration
- **WebSocket** and **SSE (Transmit)** support
- **Telescope** (request debugging), **Horizon** (queue dashboard), and **Studio** (admin panel)

---

## Why Nimbus?

### The Problem

Go is fantastic for building high-performance backend services. But when you want to build a **full web application** — with routing, ORM, authentication, validation, templating, queue workers, scheduled tasks, deployment tooling, and an admin panel — you end up:

1. **Gluing together 15+ libraries** (gorilla/mux + gorm + viper + jwt-go + ...)
2. **Writing the same boilerplate** in every project (config loading, middleware stacks, error handling)
3. **Reinventing conventions** that frameworks like Laravel solved years ago
4. **Missing developer tools** like code generators, interactive CLIs, and debug dashboards

### The Solution

Nimbus provides a **cohesive, convention-based framework** where all these pieces work together seamlessly:

```go
// main.go — That's it. Your entire application entry point.
package main

import "nimbus-starter/bin"

func main() {
    app := bin.Boot()
    app.Run()
}
```

### Design Philosophy

| Principle | Description |
|-----------|-------------|
| **Convention over Configuration** | Sensible defaults for everything. Override only what you need. |
| **Batteries Included** | Every common web feature included — no hunting for compatible packages. |
| **Developer Joy** | Beautiful CLI, hot reload, code generators, AI-assisted development. |
| **Go-Idiomatic** | Leverages Go's strengths — interfaces, goroutines, type safety — not fighting them. |
| **Plugin Architecture** | Everything is optional. Use what you need, ignore what you don't. |
| **Production Ready** | Health checks, graceful shutdown, structured logging, metrics, security hardening. |

---

## Real-World Use Cases

### 1. SaaS Application
Build a multi-tenant SaaS platform with user authentication, subscription management, background job processing, and real-time notifications:

```go
// Multi-tenant route with auth and rate limiting
admin := app.Router.Group("/api/v1", authMiddleware, tenantMiddleware)
admin.Resource("projects",  &controllers.Project{})
admin.Resource("invoices",  &controllers.Invoice{})
admin.Post("/webhooks/stripe", webhookHandler)

// Background job for invoice generation
queue.Dispatch(&jobs.GenerateMonthlyInvoices{TenantID: tenant.ID})

// Real-time notification via SSE
transmit.Broadcast("tenant."+tenantID, "invoice.created", invoiceData)
```

### 2. REST API for Mobile Apps
Build a JSON API with token authentication, validation, pagination, and caching:

```go
api := app.Router.Group("/api/v2")
api.Use(middleware.RateLimit(100, time.Minute, middleware.DefaultKeyFn))

api.Post("/auth/login",    authCtrl.Login)
api.Post("/auth/register", authCtrl.Register)

// Protected routes
protected := api.Group("", tokenAuthMiddleware)
protected.Resource("posts",    &controllers.PostAPI{})
protected.Resource("comments", &controllers.CommentAPI{})
protected.Get("/feed", feedCtrl.Index)  // Cached, paginated feed
```

### 3. Internal Admin Dashboard
Auto-generate an admin panel from your GORM models with Studio:

```go
app.Use(studio.New(studio.Config{
    Models: []any{&models.User{}, &models.Order{}, &models.Product{}},
    Auth:   adminAuthMiddleware,
}))
// Admin panel auto-generated at /studio with CRUD, search, filters, and charts
```

### 4. AI-Powered Application
Integrate multiple AI providers with streaming support:

```go
// AI text generation with multiple providers
response, err := ai.Generate(ctx, "Summarize this article: "+article,
    ai.WithProvider("anthropic"),
    ai.WithModel("claude-sonnet-4-20250514"),
    ai.WithMaxTokens(500),
)

// MCP server for AI tool integration
mcpServer := nimbusmcp.NewServer("MyApp", "1.0.0")
mcpServer.AddTool(mcp.NewTool("search_products", ...),  searchHandler)
mcpServer.AddTool(mcp.NewTool("create_order", ...),     orderHandler)
```

### 5. Real-Time Collaborative App
WebSocket channels with presence tracking:

```go
// Presence channel for collaborative editing
app.Router.Get("/ws", websocket.Handler(func(conn *websocket.Conn) {
    presence.Join("document."+docID, conn, userData)
    // Users can see who's editing in real-time
}))
```

---

## Framework Comparison

| Feature | Nimbus (Go) | Laravel (PHP) | AdonisJS (Node) | Gin (Go) | Fiber (Go) |
|---------|------------|---------------|-----------------|----------|------------|
| Full MVC | ✅ | ✅ | ✅ | ❌ | ❌ |
| ORM | ✅ (GORM) | ✅ (Eloquent) | ✅ (Lucid) | ❌ | ❌ |
| Migrations | ✅ | ✅ | ✅ | ❌ | ❌ |
| Validation | ✅ (VineJS-style) | ✅ | ✅ (VineJS) | ❌ | ❌ |
| Auth Guards | ✅ | ✅ | ✅ | ❌ | ❌ |
| Queue System | ✅ | ✅ | ❌ | ❌ | ❌ |
| Task Scheduler | ✅ | ✅ | ❌ | ❌ | ❌ |
| CLI Generator | ✅ | ✅ (Artisan) | ✅ (Ace) | ❌ | ❌ |
| Template Engine | ✅ (.nimbus) | ✅ (Blade) | ✅ (Edge) | ✅ | ✅ |
| Plugin System | ✅ | ✅ (Packages) | ✅ | ❌ | ❌ |
| AI SDK | ✅ | ❌ | ❌ | ❌ | ❌ |
| MCP Support | ✅ | ❌ | ❌ | ❌ | ❌ |
| Admin Panel | ✅ (Studio) | ✅ (Nova) | ❌ | ❌ | ❌ |
| Performance | 🔥 Go-speed | 🐢 | ⚡ | 🔥 | 🔥 |

---

## Under the Hood

Nimbus stands on the shoulders of well-tested Go libraries:

| Component | Library | Why |
|-----------|---------|-----|
| HTTP Router | [chi](https://github.com/go-chi/chi) | Fast, composable, stdlib-compatible |
| ORM | [GORM](https://gorm.io) | Most mature Go ORM, rich ecosystem |
| CLI | [Cobra](https://github.com/spf13/cobra) | Industry-standard CLI framework |
| Environment | [godotenv](https://github.com/joho/godotenv) | Simple .env loading |
| Hot Reload | [Air](https://github.com/cosmtrek/air) | Fast, reliable live reload |
| Interactive Prompts | [Survey](https://github.com/AlecAivazis/survey) | Beautiful terminal prompts |
| Terminal UI | [Lipgloss](https://github.com/charmbracelet/lipgloss) | Styled terminal output |
| WebSockets | [gorilla/websocket](https://github.com/gorilla/websocket) | Battle-tested WS implementation |
| MCP | [mcp-go](https://github.com/mark3labs/mcp-go) | Model Context Protocol for Go |

---

## Quick Start

```bash
# Install the Nimbus CLI
go install github.com/CodeSyncr/nimbus/cmd/nimbus@latest

# Create a new project
nimbus new my-app

# Navigate to project
cd my-app

# Start development server with hot reload
nimbus serve

# Your app is running at http://localhost:3333
```

**Next:** [Installation & Setup](02-installation.md) →
