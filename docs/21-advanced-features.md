# Advanced Features

> **Framework differentiators** — OpenAPI generation, Studio admin panel, Workflow engine, Feature flags, Multi-tenancy, Realtime presence, Edge functions, Telescope, and more.

---

## OpenAPI 3.0 Generation

Automatically generate OpenAPI spec from your routes:

```go
import "github.com/CodeSyncr/nimbus/openapi"

app.Use(openapi.New())
// Spec available at /openapi.json
// Swagger UI at /docs
```

### Route Metadata

Add metadata to routes for richer API documentation:

```go
app.Get("/api/products", handler, router.RouteMeta{
    Summary:     "List products",
    Description: "Returns a paginated list of products",
    Tags:        []string{"Products"},
    Params: []router.ParamMeta{
        {Name: "page", In: "query", Type: "integer", Description: "Page number"},
        {Name: "per_page", In: "query", Type: "integer", Description: "Items per page"},
    },
    Responses: map[int]router.ResponseMeta{
        200: {Description: "Product list", Schema: []Product{}},
        401: {Description: "Unauthorized"},
    },
})
```

The OpenAPI plugin scans all registered routes and generates a complete spec including:
- Path parameters, query parameters
- Request/response bodies
- Authentication requirements
- Tag grouping

---

## Studio (Admin Panel)

Auto-generated admin panel for your models:

```go
import "github.com/CodeSyncr/nimbus/studio"

app.Use(studio.New())
// Admin panel at /studio
```

Studio automatically:
- Discovers all GORM models
- Generates CRUD interfaces
- Shows database statistics
- Provides search and filtering
- Displays model relationships
- Handles pagination

### Configuration

```go
app.Use(studio.New(studio.Config{
    Prefix:     "/admin",     // URL prefix (default: /studio)
    Models:     []any{&User{}, &Product{}, &Order{}},
    ReadOnly:   false,        // Disable writes
}))
```

---

## Workflow Engine

Durable, step-based workflows with persistence and error recovery:

```go
import "github.com/CodeSyncr/nimbus/workflow"

app.Use(workflow.NewPlugin())

// Define a workflow
wf := workflow.Define("onboarding", func(ctx *workflow.Context) error {
    // Step 1: Create account
    err := ctx.Step("create_account", func() error {
        return createAccount(ctx.Input)
    })
    if err != nil {
        return err
    }

    // Step 2: Send welcome email
    err = ctx.Step("send_email", func() error {
        return sendWelcomeEmail(ctx.Input)
    })
    if err != nil {
        return err
    }

    // Step 3: Setup defaults
    err = ctx.Step("setup_defaults", func() error {
        return setupUserDefaults(ctx.Input)
    })

    return err
})

// Start a workflow instance
instance, err := workflow.Start("onboarding", map[string]any{
    "email": "user@example.com",
    "plan":  "pro",
})
```

### Workflow Features

- **Step persistence** — Each step is checkpointed; restart resumes from last completed step
- **Error handling** — Steps can fail and be retried
- **Dashboard** — View workflow status at `/workflows`
- **Parallel steps** — Run independent steps concurrently
- **Timeouts** — Set step-level timeouts

---

## Feature Flags

Runtime feature toggling without deployments:

```go
import "github.com/CodeSyncr/nimbus/flags"

app.Use(flags.NewPlugin())

// Define flags
flags.Define("dark_mode", false)         // Boolean flag with default
flags.Define("max_upload_mb", 10)        // Integer flag
flags.Define("beta_features", false)

// Check flags
if flags.Enabled("dark_mode") {
    // Show dark theme
}

// Get typed values
maxUpload := flags.Get[int]("max_upload_mb")

// Toggle at runtime (via API or admin panel)
flags.Set("dark_mode", true)
```

### Use Cases

```go
// A/B testing
func (ctrl *ProductController) Show(c *http.Context) error {
    product := getProduct(c.Param("id"))

    if flags.Enabled("new_product_page") {
        return c.View("products/show_v2", product)
    }
    return c.View("products/show", product)
}

// Gradual rollout
func (ctrl *CheckoutController) Store(c *http.Context) error {
    if flags.Enabled("new_payment_flow") {
        return processPaymentV2(c)
    }
    return processPayment(c)
}
```

---

## Multi-Tenancy

Serve multiple tenants from a single application:

```go
import "github.com/CodeSyncr/nimbus/tenancy"

app.Use(tenancy.New(tenancy.Config{
    Strategy: "subdomain",  // subdomain | header | path
    Header:   "X-Tenant-ID",
}))
```

### Tenant Resolution

```go
// In a controller
func (ctrl *DashboardController) Index(c *http.Context) error {
    tenant := tenancy.FromContext(c.Request.Context())
    
    // Query scoped to tenant
    var users []User
    db.Where("tenant_id = ?", tenant.ID).Find(&users)

    return c.JSON(200, users)
}
```

### Strategies

| Strategy | Resolution | Example |
|----------|-----------|---------|
| `subdomain` | `acme.myapp.com` | Extract from Host header |
| `header` | `X-Tenant-ID: acme` | Custom header |
| `path` | `/tenant/acme/...` | URL path prefix |

---

## Realtime Presence

Track who's online in real-time:

```go
import "github.com/CodeSyncr/nimbus/presence"

app.Use(presence.New(presence.Config{
    TTL:            30 * time.Second,  // Heartbeat timeout
    CleanupInterval: 10 * time.Second,
}))
```

### API

```go
// Join a channel
presence.Join("chat:lobby", userID, metadata)

// Leave a channel
presence.Leave("chat:lobby", userID)

// Track heartbeat
presence.Heartbeat("chat:lobby", userID)

// Get online members
members := presence.Members("chat:lobby")

// Count online users
count := presence.Count("chat:lobby")
```

### Use Cases

- **Chat applications** — Show who's in a chat room
- **Collaboration tools** — Show who's editing a document
- **Gaming** — Track players in a game lobby
- **Live dashboards** — Show active viewers

---

## Telescope (Debug Dashboard)

Monitor requests, queries, jobs, and more:

```go
import "github.com/CodeSyncr/nimbus/plugins/telescope"

tel := telescope.New()
app.Use(tel)
// Dashboard at /telescope
```

### Watchers

```go
// Track HTTP requests
app.UseMiddleware(telescope.RequestWatcher(tel))

// Track database queries (via GORM plugin)
telescope.DatabaseWatcher(tel, db)

// Track queue jobs
telescope.QueueWatcher(tel)

// Track cache operations
telescope.CacheWatcher(tel)
```

### Dashboard Views

| Tab | Shows |
|-----|-------|
| Requests | All HTTP requests with timing, status, headers |
| Queries | SQL queries with execution time and bindings |
| Jobs | Queue job status, timing, failures |
| Cache | Cache hits, misses, and performance |
| Exceptions | Error traces with context |
| Logs | Application log entries |

---

## Edge Functions

Serverless-style request handlers at the edge:

```go
import "github.com/CodeSyncr/nimbus/edge"

edge.Register("hello", func(ctx *edge.Context) *edge.Response {
    return edge.JSON(200, map[string]string{
        "message": "Hello from the edge!",
        "region":  ctx.Region,
    })
})
```

Edge functions are lightweight, isolated handlers optimized for CDN/edge deployment. They receive a simplified context and return a response — no middleware chain, no database, just fast computation at the edge.

---

## WebSocket Support

Real-time bidirectional communication:

```go
import "github.com/CodeSyncr/nimbus/websocket"

hub := websocket.NewHub()
go hub.Run()

app.Get("/ws", func(c *http.Context) error {
    conn, err := hub.Upgrade(c.Writer, c.Request)
    if err != nil {
        return err
    }
    // conn is now a WebSocket, auto-registered with the hub
    return nil
})

// Broadcast to all connected clients
hub.Broadcast([]byte(`{"event": "update", "data": {...}}`))
```

---

## Transmit (Server-Sent Events)

One-way real-time streaming from server to client:

```go
import "github.com/CodeSyncr/nimbus/packages/transmit"

t := transmit.New()

// Register SSE endpoint
app.Get("/events", t.Handler())

// Send events from anywhere
t.Broadcast("notifications", map[string]any{
    "title":   "New Order",
    "orderID": 42,
})
```

Client-side:

```javascript
const es = new EventSource('/events?channel=notifications');
es.onmessage = (e) => {
    const data = JSON.parse(e.data);
    showNotification(data.title);
};
```

---

## Events System

Application-wide pub/sub for decoupled architecture:

```go
import "github.com/CodeSyncr/nimbus/events"

// Listen for events
events.Listen("order.created", func(payload any) error {
    order := payload.(*Order)
    // Send confirmation email
    // Update inventory
    // Notify warehouse
    return nil
})

// Dispatch events
events.Dispatch("order.created", order)

// Async dispatch (non-blocking)
events.DispatchAsync("order.created", order)
```

### Built-in Framework Events

| Event | Payload | When |
|-------|---------|------|
| `events.AppBooted` | nil | Application boot complete |
| `events.AppStarted` | port string | Server listening |
| `events.AppShutdown` | os.Signal | Graceful shutdown started |
| `events.RouteRegistered` | nil | Plugin routes mounted |
| `events.DatabaseQuery` | — | Database query executed |

---

## File Storage

Unified file operations across local filesystem and cloud providers:

```go
import "github.com/CodeSyncr/nimbus/storage"

// Store a file
err := storage.Put("uploads/photo.jpg", reader)

// Read a file
reader, err := storage.Get("uploads/photo.jpg")

// Delete a file
err := storage.Delete("uploads/photo.jpg")

// Check existence
exists, err := storage.Exists("uploads/photo.jpg")
```

### File Uploads

```go
func (ctrl *ProfileController) Upload(c *http.Context) error {
    file := storage.NewUploadedFile(c.Request.FormFile("avatar"))

    // Validate
    if !storage.AllowedExtensions(file, ".jpg", ".png", ".webp") {
        return c.JSON(422, map[string]string{"error": "Invalid file type"})
    }
    if !storage.MaxFileSize(file, 5*1024*1024) { // 5MB
        return c.JSON(422, map[string]string{"error": "File too large"})
    }

    // Store with random name
    path, err := file.StoreRandomName(localDriver, "avatars")
    if err != nil {
        return err
    }

    return c.JSON(200, map[string]string{"path": path})
}
```

### Signed URLs

Generate temporary, secure download URLs:

```go
gen := storage.NewSignedURLGenerator(secretKey, "https://myapp.com")

// Generate URL valid for 1 hour
url := gen.TemporaryURL("documents/invoice.pdf", time.Hour)

// Serve signed files
app.Get("/files/*", storage.ServeSignedFiles(localDriver, gen, "/files"))
```

---

## Session Management

Server-side session storage with multiple backends:

```go
import "github.com/CodeSyncr/nimbus/session"

app.UseMiddleware(session.Middleware(session.Config{
    Store:      session.NewRedisStore(redisClient),
    CookieName: "nimbus_session",
    MaxAge:     7 * 24 * time.Hour,
    HttpOnly:   true,
    Secure:     true,
    SameSite:   http.SameSiteLaxMode,
}))
```

### Session Drivers

| Driver | Constructor | Use Case |
|--------|-------------|----------|
| Memory | `NewMemoryStore()` | Development |
| Cookie | `NewCookieStore(key)` | Simple apps (AES-256-GCM encrypted) |
| Redis | `NewRedisStore(client)` | Production |
| Database | `NewDatabaseStore(db, table)` | Simple production |

### Session API

```go
func (ctrl *CartController) Add(c *http.Context) error {
    sess := session.FromContext(c.Request.Context())

    // Get cart items
    cart := sess.Get("cart")

    // Update cart
    sess.Set("cart", updatedCart)

    // Delete a key
    sess.Delete("flash_message")

    // Regenerate ID (after login for security)
    sess.Regenerate()

    return c.JSON(200, cart)
}
```

---

## Logging

Structured logging built on Zap:

```go
import "github.com/CodeSyncr/nimbus/logger"

logger.Info("server started", "port", 3000)
logger.Error("database connection failed", "error", err, "host", dbHost)
logger.Debug("processing request", "path", path, "method", method)

// Scoped logger with persistent fields
reqLogger := logger.With("request_id", requestID, "user_id", userID)
reqLogger.Info("processing order")
reqLogger.Error("payment failed", "error", err)
```

### Request-Scoped Logger

Create a logger that automatically includes request metadata (request ID, method, path):

```go
func (ctrl *OrderController) Create(c *http.Context) error {
    log := logger.ForRequest(c)
    log.Info("creating order")
    // Output: {"msg":"creating order","request_id":"abc-123","method":"POST","path":"/orders"}

    // Store the logger in context for downstream use
    logger.WithContext(c, log)
    return nil
}
```

### Log Rotation

Automatic file rotation by size with backup retention:

```go
writer := logger.NewRotatingWriter(logger.RotationConfig{
    Path:       "logs/app.log",
    MaxSizeMB:  100,  // Rotate after 100MB
    MaxBackups: 5,    // Keep 5 old files
})
```

---

## Health Checks

Monitor application and dependency health:

```go
import "github.com/CodeSyncr/nimbus/health"

checker := health.New()
checker.DB(db)
checker.Redis(redisClient)
checker.Add("external_api", func(ctx context.Context) error {
    resp, err := http.Get("https://api.example.com/health")
    if err != nil || resp.StatusCode != 200 {
        return fmt.Errorf("external API unhealthy")
    }
    return nil
})

app.Get("/health", checker.Handler())
// Returns: {"status": "ok", "checks": {"database": "ok", "redis": "ok", "external_api": "ok"}}
```

---

## Metrics

Prometheus-compatible metrics with counters, gauges, and histograms:

```go
import "github.com/CodeSyncr/nimbus/metrics"

// Counter — monotonically increasing value
requestCount := metrics.Counter("http_requests_total", "Total HTTP requests", metrics.Labels{"method": "GET"})
requestCount.Inc()

// Gauge — value that can go up and down
activeConns := metrics.Gauge("active_connections", "Current active connections")
activeConns.Set(42)
activeConns.Inc()
activeConns.Dec()

// Histogram — distribution of values across buckets
latency := metrics.Histogram("http_request_duration_seconds", "Request latency", metrics.DefaultBuckets)
latency.Observe(0.250)
```

### Prometheus Endpoint

Expose metrics in Prometheus text format:

```go
// Using the handler directly
app.Get("/metrics", metrics.Handler(metrics.DefaultRegistry))

// Or use the Metrics middleware (auto-tracks request count + latency)
app.UseMiddleware(middleware.Metrics())
```

### Runtime Stats

Built-in Go runtime metrics:

```go
stats := metrics.ReadRuntimeStats()
// stats.Goroutines  — Active goroutine count
// stats.HeapAlloc   — Current heap allocation
// stats.HeapSys     — Total heap from OS
// stats.NumGC       — Completed GC cycles
// stats.HeapObjects — Allocated heap objects
```

---

**Next:** [README & Index](00-README.md) →
