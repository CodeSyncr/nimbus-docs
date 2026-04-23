# Middleware

> **Layered request processing** — intercept, transform, and guard HTTP requests with composable middleware functions.

---

## Introduction

Middleware functions sit between the incoming request and your route handler. They can:

- **Inspect** the request (logging, monitoring)
- **Transform** the request (parsing, authentication)
- **Guard** the route (authorization, rate limiting)
- **Modify** the response (CORS headers, compression)
- **Short-circuit** the pipeline (return 401/403/429 without hitting the handler)

Nimbus uses a three-layer middleware architecture inspired by Laravel patterns:

```
Request → [Server Middleware] → [Router Middleware] → [Named Middleware] → Handler → Response
```

---

## Middleware Architecture

### Three Layers

| Layer | Scope | Defined In | Example |
|-------|-------|-----------|---------|
| **Server Middleware** | Every HTTP request, even 404s | `start/kernel.go` | Logger, Recover, Shield |
| **Router Middleware** | All registered routes | `start/kernel.go` | CORS, CSRF |
| **Named Middleware** | Specific routes/groups | `start/routes.go` | Auth, Admin, RateLimit |

### The Middleware Signature

```go
type Middleware = func(HandlerFunc) HandlerFunc
type HandlerFunc = func(*http.Context) error
```

A middleware receives the `next` handler and returns a new handler:

```go
func myMiddleware() router.Middleware {
    return func(next router.HandlerFunc) router.HandlerFunc {
        return func(c *http.Context) error {
            // Before handler
            fmt.Println("Request started")

            err := next(c)  // Call the next middleware/handler

            // After handler
            fmt.Println("Request finished")

            return err
        }
    }
}
```

---

## Registering Middleware

### HTTP Kernel (`start/kernel.go`)

The kernel defines the server-wide middleware stack:

```go
package start

import (
    "github.com/CodeSyncr/nimbus"
    "github.com/CodeSyncr/nimbus/errors"
    "github.com/CodeSyncr/nimbus/middleware"
    "github.com/CodeSyncr/nimbus/packages/shield"
    "github.com/CodeSyncr/nimbus/plugins/telescope"
    "github.com/CodeSyncr/nimbus/plugins/unpoly"
    "github.com/CodeSyncr/nimbus/router"
)

func RegisterMiddleware(app *nimbus.App) {
    // ── Server Middleware ──────────────────────────────────
    // Runs on EVERY request, even if no route matches
    shieldCfg := shield.DefaultConfig()
    shieldCfg.CSRF.ExceptPaths = append(shieldCfg.CSRF.ExceptPaths, "/api/docs/chat")

    app.Router.Use(
        middleware.Logger(),             // Log every request
        middleware.Recover(),            // Catch panics
        errors.Handler(),                // Smart error pages
        shield.Guard(shieldCfg),         // Security headers
        shield.CSRFGuard(shieldCfg.CSRF),// CSRF protection
        unpoly.ServerProtocol(),         // Unpoly integration
    )

    // Telescope (conditional)
    if te := app.Plugin("telescope"); te != nil {
        if t, ok := te.(*telescope.Plugin); ok {
            app.Router.Use(t.RequestWatcher())
        }
    }
}

// ── Named Middleware ────────────────────────────────────────
var Middleware = map[string]router.Middleware{
    // "auth":  middleware.RequireAuth(),
    // "guest": guestMiddleware(),
}
```

### Route-Level Middleware

```go
// Apply to a group
admin := app.Router.Group("/admin", requireAdmin)

// Apply to a single route
app.Router.Get("/dashboard", dashboardHandler, authMiddleware)

// Apply using named middleware from kernel
app.Router.Get("/profile", profileHandler, start.Middleware["auth"])
```

---

## Built-In Middleware

### Logger

Logs every request with method, path, status, and duration:

```go
app.Router.Use(middleware.Logger())
// Output: 2024/01/15 10:30:45 GET /api/products 200 3.2ms
```

### Recover

Catches panics and returns 500 JSON response instead of crashing:

```go
app.Router.Use(middleware.Recover())
// If handler panics:
// {"error":"internal server error"}
```

### CORS

Adds Cross-Origin Resource Sharing headers:

```go
app.Router.Use(middleware.CORS("https://myapp.com"))

// Or with multiple origins:
app.Router.Use(middleware.CORS("https://myapp.com, https://admin.myapp.com"))
```

### CSRF

Validates CSRF tokens on state-changing requests (POST, PUT, DELETE):

```go
store := middleware.NewMemoryCSRFStore()
app.Router.Use(middleware.CSRF(store))

// Generate token for forms:
token := middleware.GenerateCSRFToken()
```

### Rate Limiting

In-memory rate limiting per client:

```go
// 100 requests per minute per IP
app.Router.Use(middleware.RateLimit(100, time.Minute, middleware.DefaultKeyFn))
```

#### Redis Rate Limiting (Multi-Instance)

```go
import "github.com/CodeSyncr/nimbus/middleware"

// Redis-backed for distributed apps
app.Router.Use(middleware.RateLimitRedis(redisClient, 100, time.Minute, middleware.DefaultKeyFn))
```

#### Custom Key Function

```go
// Rate limit by user ID instead of IP
userRateLimit := middleware.RateLimit(50, time.Minute, func(r *http.Request) string {
    user := auth.UserFromContext(r.Context())
    if user != nil {
        return user.GetID()
    }
    return r.RemoteAddr
})
```

### Shield

Comprehensive security headers and protections:

```go
import "github.com/CodeSyncr/nimbus/packages/shield"

cfg := shield.DefaultConfig()
app.Router.Use(shield.Guard(cfg))

// Adds headers:
// X-Frame-Options: DENY
// X-Content-Type-Options: nosniff
// X-XSS-Protection: 1; mode=block
// Referrer-Policy: strict-origin-when-cross-origin
// Content-Security-Policy: ...
```

### Error Handler

Smart error pages — detailed in development, clean in production:

```go
import "github.com/CodeSyncr/nimbus/errors"

app.Router.Use(errors.Handler())
// Development: Full stack trace, code snippet, request details
// Production: Clean "Something went wrong" page
```

### Request ID

Generates a unique request ID for every request via `X-Request-Id` header and context store:

```go
app.Router.Use(middleware.RequestID())

// Access in handlers:
id, _ := c.Get("request_id")
```

If the incoming request already has `X-Request-Id`, it is reused (useful behind load balancers).

### Timeout

Wraps each request with a context deadline:

```go
app.Router.Use(middleware.Timeout(30 * time.Second))

// Check in long-running handlers:
if c.Ctx().Err() != nil {
    return c.Ctx().Err() // context cancelled or deadline exceeded
}
```

### Body Limit

Limits request body size. Returns 413 when exceeded:

```go
app.Router.Use(middleware.BodyLimit(10 * 1024 * 1024)) // 10 MB
```

### Gzip

Compresses response bodies for clients that accept it. Responses under 256 bytes are skipped:

```go
app.Router.Use(middleware.Gzip())
```

### Secure Headers

Adds production-ready security headers (HSTS, X-Frame-Options, XSS protection, etc.):

```go
app.Router.Use(middleware.SecureHeaders(middleware.SecureHeadersConfig{
    HSTS:               true,
    HSTSMaxAge:         31536000,
    FrameOptions:       "DENY",
    ContentTypeNoSniff: true,
    XSSProtection:      true,
}))
```

### Trusted Proxies

Strips forwarded-for headers when the request does not come from a trusted proxy IP:

```go
app.Router.Use(middleware.TrustedProxies("10.0.0.0/8", "172.16.0.0/12"))
```

### Metrics

Records Prometheus-compatible HTTP metrics (request count, duration, in-flight, response size):

```go
app.Router.Use(middleware.Metrics())

// Mount the /metrics endpoint for Prometheus scraping:
app.Router.Handle("/metrics", metrics.Handler())
```

---

## Creating Custom Middleware

### Example 1: Authentication Middleware

```go
// app/middleware/auth.go
package middleware

import (
    "github.com/CodeSyncr/nimbus/auth"
    "github.com/CodeSyncr/nimbus/http"
    "github.com/CodeSyncr/nimbus/router"
)

func RequireAuth() router.Middleware {
    return func(next router.HandlerFunc) router.HandlerFunc {
        return func(c *http.Context) error {
            user := auth.UserFromContext(c.Request.Context())
            if user == nil {
                return c.JSON(http.StatusUnauthorized, map[string]string{
                    "error": "authentication required",
                })
            }
            return next(c)
        }
    }
}
```

### Example 2: Role-Based Authorization

```go
// app/middleware/role.go
func RequireRole(roles ...string) router.Middleware {
    return func(next router.HandlerFunc) router.HandlerFunc {
        return func(c *http.Context) error {
            user := auth.UserFromContext(c.Request.Context())
            if user == nil {
                return c.JSON(http.StatusUnauthorized, map[string]string{
                    "error": "authentication required",
                })
            }

            // Check if user has any of the required roles
            userRole := user.(*models.User).Role
            for _, role := range roles {
                if userRole == role {
                    return next(c)
                }
            }

            return c.JSON(http.StatusForbidden, map[string]string{
                "error": "insufficient permissions",
            })
        }
    }
}

// Usage:
admin := app.Router.Group("/admin", RequireRole("admin", "super_admin"))
```

### Example 3: Request Timing

```go
func RequestTiming() router.Middleware {
    return func(next router.HandlerFunc) router.HandlerFunc {
        return func(c *http.Context) error {
            start := time.Now()

            err := next(c)

            duration := time.Since(start)
            c.Response.Header().Set("X-Response-Time", duration.String())

            if duration > 500*time.Millisecond {
                logger.Warn("Slow request",
                    "path", c.Request.URL.Path,
                    "duration", duration,
                )
            }

            return err
        }
    }
}
```

### Example 4: JSON-Only Middleware

```go
func RequireJSON() router.Middleware {
    return func(next router.HandlerFunc) router.HandlerFunc {
        return func(c *http.Context) error {
            contentType := c.Request.Header.Get("Content-Type")
            if c.Request.Method != "GET" && !strings.Contains(contentType, "application/json") {
                return c.JSON(http.StatusUnsupportedMediaType, map[string]string{
                    "error": "Content-Type must be application/json",
                })
            }
            return next(c)
        }
    }
}
```

### Example 5: API Key Middleware

```go
func RequireAPIKey(validKeys map[string]string) router.Middleware {
    return func(next router.HandlerFunc) router.HandlerFunc {
        return func(c *http.Context) error {
            key := c.Request.Header.Get("X-API-Key")
            if key == "" {
                key = c.Request.URL.Query().Get("api_key")
            }

            clientName, valid := validKeys[key]
            if !valid {
                return c.JSON(http.StatusUnauthorized, map[string]string{
                    "error": "invalid or missing API key",
                })
            }

            // Store client info for handlers
            c.Set("api_client", clientName)
            return next(c)
        }
    }
}

// Usage:
keys := map[string]string{
    "sk_live_abc123": "Mobile App",
    "sk_live_def456": "Partner API",
}
api := app.Router.Group("/api", RequireAPIKey(keys))
```

### Example 6: Tenant Resolution (Multi-Tenancy)

```go
func ResolveTenant() router.Middleware {
    return func(next router.HandlerFunc) router.HandlerFunc {
        return func(c *http.Context) error {
            // Resolve tenant from subdomain
            host := c.Request.Host
            parts := strings.Split(host, ".")
            if len(parts) < 3 {
                return c.JSON(http.StatusBadRequest, map[string]string{
                    "error": "tenant not found",
                })
            }
            tenantSlug := parts[0]

            var tenant models.Tenant
            if db.Where("slug = ?", tenantSlug).First(&tenant).Error != nil {
                return c.JSON(http.StatusNotFound, map[string]string{
                    "error": "tenant not found",
                })
            }

            // Store tenant in context
            c.Set("tenant", &tenant)
            c.Set("tenant_id", tenant.ID)

            return next(c)
        }
    }
}
```

---

## Real-Life Example: Full API Middleware Stack

```go
func RegisterMiddleware(app *nimbus.App) {
    // Server middleware (all requests)
    app.Router.Use(
        middleware.Logger(),
        middleware.Recover(),
        RequestTiming(),
        errors.Handler(),
        shield.Guard(shield.DefaultConfig()),
    )
}

func RegisterRoutes(app *nimbus.App) {
    // Public API — rate limited
    public := app.Router.Group("/api/v1",
        middleware.CORS("*"),
        middleware.RateLimit(60, time.Minute, middleware.DefaultKeyFn),
        RequireJSON(),
    )
    public.Get("/products", listProducts)
    public.Get("/categories", listCategories)

    // Authenticated API — stricter rate limiting
    auth := public.Group("",
        RequireAuth(),
        middleware.RateLimit(120, time.Minute, userKeyFn),
    )
    auth.Get("/profile", showProfile)
    auth.Put("/profile", updateProfile)
    auth.Resource("orders", &controllers.Order{})

    // Admin API
    admin := auth.Group("/admin",
        RequireRole("admin"),
        middleware.RateLimit(300, time.Minute, userKeyFn),
    )
    admin.Resource("users", &controllers.AdminUser{})
    admin.Get("/analytics", analyticsHandler)

    // Webhooks — no auth but signature verification
    webhooks := app.Router.Group("/webhooks")
    webhooks.Post("/stripe", VerifyStripeSignature(), stripeWebhookHandler)
    webhooks.Post("/github", VerifyGitHubSignature(), githubWebhookHandler)
}
```

---

## Middleware Execution Order

Middleware executes in the order it's registered. Later middleware wraps earlier middleware:

```
Request
  → Logger          (before)
    → Recover        (before)
      → Auth         (before — may short-circuit)
        → Handler    (execute)
      → Auth         (after)
    → Recover        (after — catches panics)
  → Logger           (after — logs duration)
Response
```

---

## Best Practices

1. **Order matters** — Logger and Recover should be first (outermost)
2. **Keep middleware focused** — Each middleware should do one thing well
3. **Short-circuit early** — Return errors immediately for unauthorized requests
4. **Use named middleware** — Map names to middleware for route-level application
5. **Avoid heavy computation** — Middleware runs on every request; keep it fast
6. **Log security events** — Failed auth attempts, rate limit hits, suspicious activity
7. **Test middleware independently** — Write unit tests for each middleware function

**Next:** [Validation](08-validation.md) →
