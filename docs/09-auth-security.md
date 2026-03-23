# Authentication & Security

> **Multi-guard authentication with session and token support, authorization policies, and comprehensive security hardening** — protect your application at every layer.

---

## Introduction

Nimbus provides a complete authentication and security system:

- **Auth Guards** — Session-based (web) and Token-based (API) authentication
- **User Context** — Store and retrieve the authenticated user throughout request lifecycle
- **Policies** — Define authorization rules per resource
- **Shield** — Security headers (X-Frame-Options, CSP, HSTS, etc.)
- **CSRF Protection** — Automatic token validation for state-changing requests
- **CORS** — Cross-Origin Resource Sharing configuration
- **Rate Limiting** — Per-IP and per-user request throttling
- **Password Hashing** — bcrypt-based secure password storage

---

## Authentication

### Guards

Nimbus supports two authentication guards:

| Guard | Use Case | Storage | Header/Cookie |
|-------|----------|---------|--------------|
| **Session** | Web applications | Server-side session | Cookie |
| **Token** | API (Opaque/Stateful) | Database | `Authorization: Bearer <token>` |
| **Stateless** | API (JWT/PASETO) | Client-side token | `Authorization: Bearer <token>` |

### Configuration

```go
// config/auth.go
var Auth AuthConfig

type AuthConfig struct {
    DefaultGuard string           // "session", "token", or "stateless"
    Session      SessionGuardConfig
    Token        TokenGuardConfig
    Stateless    StatelessTokenConfig
}

func loadAuth() {
    Auth = AuthConfig{
        DefaultGuard: env("AUTH_GUARD", "session"),
        Session: SessionGuardConfig{
            CookieName: env("SESSION_COOKIE", "nimbus_session"),
            MaxAge:     envInt("SESSION_MAX_AGE", 604800), // 7 days
        },
        Token: TokenGuardConfig{
            HeaderName: "Authorization",
            Scheme:     "Bearer",
            ExpiresIn:  envInt("TOKEN_EXPIRES_IN", 86400), // 1 day
        },
    }
}
```

### Session Guard

For server-rendered web applications:

```go
import "github.com/CodeSyncr/nimbus/auth"

// Create a session guard with user loader
guard := auth.NewSessionGuard(func(id string) (auth.User, error) {
    var user models.User
    if err := db.First(&user, id).Error; err != nil {
        return nil, err
    }
    return &user, nil
})

// Login
func loginHandler(c *http.Context) error {
    email := c.Request.FormValue("email")
    password := c.Request.FormValue("password")

    var user models.User
    if err := db.Where("email = ?", email).First(&user).Error; err != nil {
        return c.View("login", map[string]any{"error": "Invalid credentials"})
    }

    if !hash.Check(password, user.Password) {
        return c.View("login", map[string]any{"error": "Invalid credentials"})
    }

    // Set user in context
    guard.Login(c, &user)
    c.Redirect(http.StatusFound, "/dashboard")
    return nil
}

// Logout
func logoutHandler(c *http.Context) error {
    guard.Logout(c)
    c.Redirect(http.StatusFound, "/")
    return nil
}

// Get current user
func dashboardHandler(c *http.Context) error {
    user := guard.User(c)
    if user == nil {
        c.Redirect(http.StatusFound, "/login")
        return nil
    }
    return c.View("dashboard", map[string]any{"user": user})
}
```

### Stateless Guard (JWT/PASETO API Authentication)

For modern APIs, mobile applications, and distributed services where you don't want to query the database for token verification on every request:

```go
// Login endpoint — returns a stateless token (JWT or PASETO)
func (ctrl *AuthController) Login(ctx *http.Context) error {
    // ... verification logic ...

    // Get the stateless guard from container
    guard := app.Container.MustMake("auth.stateless").(*auth.StatelessGuard)

    // Generate token
    token, err := guard.GenerateToken(user.GetID(), config.Auth.Stateless.ExpiresIn)
    if err != nil {
        return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "token generation failed"})
    }

    return ctx.JSON(http.StatusOK, map[string]any{
        "token":      token,
        "token_type": "Bearer",
        "user":       user,
    })
}

// Protected API routes using auth:api middleware
// In start/routes.go:
// protected := app.Router.Group("/api", start.Middleware["auth:api"])
```

#### JWT vs PASETO

Nimbus supports both common drivers. Switch them in your `.env`:

- **JWT**: Industry standard, widely supported, but prone to misconfiguration (e.g., `alg: none`).
- **PASETO**: Modern, secure-by-default alternative that avoids most JWT pitfalls. Recommended for new projects.

```env
AUTH_TOKEN_DRIVER=paseto
AUTH_TOKEN_SECRET=your-32-character-hex-key
```

### Token Guard (Opaque/Stateful API Authentication)

For APIs where you need full control over token revocation (e.g., "Logout from all devices"):

```go
// Generate stateful token (stored in database)
token, err := auth.GenerateToken(user.GetID(), config.Auth.Token.ExpiresIn)
```

### User Context

Store and retrieve the authenticated user in any handler:

```go
import "github.com/CodeSyncr/nimbus/auth"

// Middleware: Set user in context
func authMiddleware(next router.HandlerFunc) router.HandlerFunc {
    return func(c *http.Context) error {
        user := resolveUser(c)  // From session or token
        if user != nil {
            ctx := auth.WithUser(c.Request.Context(), user)
            c.Request = c.Request.WithContext(ctx)
        }
        return next(c)
    }
}

// Handler: Get user from context
user := auth.UserFromContext(c.Request.Context())
if user != nil {
    fmt.Println("Logged in as:", user.GetID())
}
```

### The User Interface

Your model must implement the `auth.User` interface:

```go
type User interface {
    GetID() string
}

// Implementation
type User struct {
    database.Model
    Email    string
    Password string `json:"-"`
    Name     string
    Role     string
}

func (u *User) GetID() string {
    return fmt.Sprintf("%d", u.ID)
}
```

---

## Authorization (Policies)

Policies define **who can do what** for each resource:

```go
import "github.com/CodeSyncr/nimbus/auth"

// Define a policy
type PostPolicy struct{}

func (p *PostPolicy) ViewAny(user auth.User) bool {
    return true  // Anyone can view posts
}

func (p *PostPolicy) View(user auth.User, post *models.Post) bool {
    return post.Published || user.GetID() == fmt.Sprintf("%d", post.AuthorID)
}

func (p *PostPolicy) Create(user auth.User) bool {
    return user.(*models.User).Role == "writer" || user.(*models.User).Role == "admin"
}

func (p *PostPolicy) Update(user auth.User, post *models.Post) bool {
    return user.GetID() == fmt.Sprintf("%d", post.AuthorID) || user.(*models.User).Role == "admin"
}

func (p *PostPolicy) Delete(user auth.User, post *models.Post) bool {
    return user.(*models.User).Role == "admin"
}

// Register the policy
auth.RegisterPolicy("posts", &PostPolicy{})

// Use in controllers
func (ctrl *PostController) Update(ctx *http.Context) error {
    user := auth.UserFromContext(ctx.Request.Context())
    post := getPost(ctx.Param("id"))

    if !auth.Can(user, "update", "posts", post) {
        return ctx.JSON(http.StatusForbidden, map[string]string{"error": "not authorized"})
    }

    // User is authorized — proceed with update
    // ...
}
```

---

## Shield (Security Headers)

Shield adds security headers to every response:

```go
import "github.com/CodeSyncr/nimbus/packages/shield"

cfg := shield.DefaultConfig()
app.Router.Use(shield.Guard(cfg))
```

### Default Headers

| Header | Value | Purpose |
|--------|-------|---------|
| `X-Frame-Options` | `DENY` | Prevent clickjacking |
| `X-Content-Type-Options` | `nosniff` | Prevent MIME sniffing |
| `X-XSS-Protection` | `1; mode=block` | XSS filter |
| `Referrer-Policy` | `strict-origin-when-cross-origin` | Control referrer info |
| `X-DNS-Prefetch-Control` | `off` | Prevent DNS prefetch |
| `X-Permitted-Cross-Domain-Policies` | `none` | Flash/PDF restrictions |

### Content Security Policy

```go
cfg := shield.DefaultConfig()
cfg.CSP = shield.CSPConfig{
    DefaultSrc: []string{"'self'"},
    ScriptSrc:  []string{"'self'", "https://cdn.example.com"},
    StyleSrc:   []string{"'self'", "'unsafe-inline'"},
    ImgSrc:     []string{"'self'", "data:", "https:"},
    FontSrc:    []string{"'self'", "https://fonts.gstatic.com"},
    ConnectSrc: []string{"'self'", "https://api.example.com"},
}
```

---

## CSRF Protection

Automatically validates CSRF tokens on POST, PUT, PATCH, and DELETE requests:

```go
// In kernel.go
shieldCfg := shield.DefaultConfig()
shieldCfg.CSRF.ExceptPaths = append(shieldCfg.CSRF.ExceptPaths, 
    "/api/docs/chat",     // Exempt API endpoints  
    "/webhooks/stripe",   // Exempt webhooks
)

app.Router.Use(shield.CSRFGuard(shieldCfg.CSRF))
```

### In Templates

Nimbus automatically injects a `csrfField` variable into all views:

```html
<form method="POST" action="/posts">
    {{ .csrfField }}
    <input type="text" name="title" />
    <button type="submit">Create</button>
</form>
```

### In JavaScript

```javascript
// Get CSRF token from meta tag
const token = document.querySelector('meta[name="csrf-token"]').content;

// Include in AJAX requests
fetch('/api/posts', {
    method: 'POST',
    headers: {
        'Content-Type': 'application/json',
        'X-CSRF-Token': token,
    },
    body: JSON.stringify({ title: 'My Post' }),
});
```

---

## CORS

Cross-Origin Resource Sharing for API access from browsers:

```go
import "github.com/CodeSyncr/nimbus/middleware"

// Allow specific origin
app.Router.Use(middleware.CORS("https://myapp.com"))

// Allow multiple origins
app.Router.Use(middleware.CORS("https://myapp.com, https://admin.myapp.com"))

// Allow all (development only!)
app.Router.Use(middleware.CORS("*"))
```

---

## Rate Limiting

Protect against abuse with per-client rate limiting:

```go
import "github.com/CodeSyncr/nimbus/middleware"

// In-memory rate limiting (single instance)
app.Router.Use(middleware.RateLimit(100, time.Minute, middleware.DefaultKeyFn))

// Redis rate limiting (distributed, multi-instance)
app.Router.Use(middleware.RateLimitRedis(redisClient, 100, time.Minute, middleware.DefaultKeyFn))

// Different rates for different routes
public := app.Router.Group("/api",
    middleware.RateLimit(60, time.Minute, middleware.DefaultKeyFn),  // 60 req/min
)

authenticated := app.Router.Group("/api",
    authMiddleware,
    middleware.RateLimit(300, time.Minute, userKeyFn),  // 300 req/min for logged-in users
)
```

### Custom Rate Limit Response

When rate limited, the response includes headers:

```
HTTP/1.1 429 Too Many Requests
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 0
X-RateLimit-Reset: 1705300000
Retry-After: 42
```

---

## Password Hashing

```go
import "github.com/CodeSyncr/nimbus/hash"

// Hash a password
hashed, err := hash.Make("user-password")
// $2a$10$...

// Verify a password
if hash.Check("user-password", hashed) {
    // Password matches
}

// In registration
func register(ctx *http.Context) error {
    password := ctx.Request.FormValue("password")
    hashed, _ := hash.Make(password)

    user := &models.User{
        Email:    ctx.Request.FormValue("email"),
        Password: hashed,
    }
    db.Create(user)
    // ...
}
```

---

## Real-Life Example: Complete Auth System

### Registration Flow

```go
func (ctrl *AuthController) Register(ctx *http.Context) error {
    // 1. Validate input
    v := &validators.Register{}
    json.NewDecoder(ctx.Request.Body).Decode(v)
    if err := v.Validate(); err != nil {
        return ctx.JSON(http.StatusUnprocessableEntity, err)
    }

    // 2. Hash password
    hashed, _ := hash.Make(v.Password)

    // 3. Create user
    user := &models.User{
        Name:     v.Name,
        Email:    v.Email,
        Password: hashed,
        Role:     "customer",
    }
    if err := db.Create(user).Error; err != nil {
        return ctx.JSON(http.StatusConflict, map[string]string{"error": "email already exists"})
    }

    // 4. Generate token
    token, _ := auth.GenerateToken(user.GetID(), config.Auth.Token.ExpiresIn)

    // 5. Dispatch welcome email
    queue.Dispatch(&jobs.SendWelcomeEmail{UserID: user.ID, Email: user.Email})

    return ctx.JSON(http.StatusCreated, map[string]any{
        "token": token,
        "user":  user,
    })
}
```

### Protected Dashboard

```go
func RegisterRoutes(app *nimbus.App) {
    // Public
    app.Router.Post("/api/auth/register", authCtrl.Register)
    app.Router.Post("/api/auth/login", authCtrl.Login)
    app.Router.Post("/api/auth/forgot-password", authCtrl.ForgotPassword)

    // Authenticated
    protected := app.Router.Group("/api", authMiddleware)
    protected.Get("/profile", profileCtrl.Show)
    protected.Put("/profile", profileCtrl.Update)
    protected.Post("/auth/logout", authCtrl.Logout)
    protected.Put("/auth/password", authCtrl.ChangePassword)

    // Admin only
    admin := protected.Group("/admin", RequireRole("admin"))
    admin.Resource("users", &controllers.AdminUser{})
    admin.Get("/dashboard", adminCtrl.Dashboard)
}
```

---

## AI Request Shield

Nimbus includes an AI-powered request protection system that detects malicious patterns:

```go
import "github.com/CodeSyncr/nimbus/shield"

app.Router.Use(shield.AIShield(shield.AIShieldConfig{
    BlockSQLInjection:  true,
    BlockXSS:           true,
    BlockPathTraversal: true,
    CustomPatterns:     []string{`(?i)eval\s*\(`},
    LogBlocked:         true,
}))
```

---

## Security Best Practices

1. **Always hash passwords** — Never store plaintext passwords
2. **Use CSRF protection** — Enabled by default for web forms
3. **Set secure cookies in production** — `Secure: true`, `HttpOnly: true`
4. **Rate limit authentication endpoints** — Prevent brute force attacks
5. **Validate and sanitize input** — Use validators on every endpoint
6. **Use Shield for security headers** — Defense-in-depth approach
7. **Keep secrets in environment variables** — Never in code or commits
8. **Use token auth for APIs** — Sessions are for browser-based apps
9. **Implement authorization policies** — Don't just authenticate, authorize
10. **Log security events** — Failed logins, rate limit hits, blocked requests

**Next:** [Views & Templates](10-views-templates.md) →
