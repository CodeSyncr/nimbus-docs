# Configuration

> **Type-safe, environment-driven configuration** — inspired by Laravel's config system with Go's compile-time safety.

---

## Introduction

Nimbus uses a **layered configuration system** that combines environment variables (`.env`), type-safe config structs, and a central loader. This gives you:

- **Environment-specific settings** — different values for development, staging, and production
- **Type safety** — no stringly-typed config access; compile-time checked structs
- **Centralized loading** — one call to `config.Load()` initializes everything
- **Sensible defaults** — works out of the box, override only what you need

---

## How Configuration Works

### The Loading Flow

```
.env file → godotenv → config.LoadAuto() → individual loaders → typed structs
```

1. **`.env` file** is read by `godotenv` and populates `os.Getenv()`
2. **`config.LoadAuto()`** merges env vars into a dot-notation key store
3. **Individual loaders** (`loadApp()`, `loadDatabase()`, etc.) read from the store and populate typed structs
4. **Your code** accesses config via typed struct fields: `config.App.Port`, `config.Database.Driver`

### The Master Loader

```go
// config/config.go
package config

import nimbusconfig "github.com/CodeSyncr/nimbus/config"

func Load() {
    _ = nimbusconfig.LoadAuto()   // Load .env, populate key store

    loadApp()          // config.App
    loadDatabase()     // config.Database
    loadQueue()        // config.Queue
    loadAuth()         // config.Auth
    loadBodyParser()   // config.BodyParser
    loadCache()        // config.Cache
    loadCORS()         // config.CORS
    loadHash()         // config.Hash
    loadLimiter()      // config.Limiter
    loadLogger()       // config.Logger
    loadMail()         // config.Mail
    loadSession()      // config.Session
    loadStatic()       // config.Static
    loadStorage()      // config.Storage
}
```

---

## Config File Reference

### Application Config (`config/app.go`)

Controls core application settings.

```go
package config

var App AppConfig

type AppConfig struct {
    Name string    // Application name (shown in logs, emails)
    Env  string    // development | production | test
    Port int       // HTTP listen port
    Host string    // Bind address (0.0.0.0 for all interfaces)
    Key  string    // Encryption key (min 32 chars for AES-256)
    HTTP HTTPConfig
}

type HTTPConfig struct {
    AllowMethodSpoofing bool   // Allow _method field in forms for PUT/DELETE
    Cookie CookieConfig
}

type CookieConfig struct {
    Domain   string
    Path     string
    MaxAge   int       // Cookie lifetime in seconds
    HttpOnly bool      // Prevent JavaScript access
    Secure   bool      // HTTPS only (auto-enabled in production)
    SameSite string    // strict | lax | none
}

func loadApp() {
    App = AppConfig{
        Name: cfg("app.name", "nimbus-starter"),
        Env:  cfg("app.env", "development"),
        Port: cfgInt("app.port", 3333),
        Host: cfg("app.host", "0.0.0.0"),
        Key:  cfg("app.key", ""),
        HTTP: HTTPConfig{
            AllowMethodSpoofing: false,
            Cookie: CookieConfig{
                Domain:   "",
                Path:     "/",
                MaxAge:   7200,
                HttpOnly: true,
                Secure:   cfg("app.env", "development") == "production",
                SameSite: "lax",
            },
        },
    }
}
```

**Environment variables:**
```env
APP_NAME=my-app
APP_ENV=development
APP_PORT=3333
APP_HOST=0.0.0.0
APP_KEY=your-32-character-secret-key-here
```

---

### Database Config (`config/database.go`)

Supports PostgreSQL, MySQL, and SQLite with automatic DSN construction.

```go
var Database DatabaseConfig

type DatabaseConfig struct {
    Driver string    // sqlite | postgres | mysql
    DSN    string    // Connection string
}

func loadDatabase() {
    driver := cfg("database.driver", "sqlite")
    var dsn string

    switch driver {
    case "postgres", "pg":
        dsn = cfg("database.dsn", "")
        if dsn == "" {
            // Auto-construct from individual fields
            dsn = "host=" + cfg("database.host", "localhost") +
                " port=" + cfg("database.port", "5432") +
                " user=" + cfg("database.user", "postgres") +
                " password=" + cfg("database.password", "") +
                " dbname=" + cfg("database.database", "nimbus") +
                " sslmode=disable"
        }
    case "mysql":
        dsn = cfg("database.dsn", "")
        if dsn == "" {
            dsn = cfg("database.user", "root") + ":" +
                cfg("database.password", "") +
                "@tcp(" + cfg("database.host", "localhost") + ":" +
                cfg("database.port", "3306") + ")/" +
                cfg("database.database", "nimbus") +
                "?charset=utf8mb4&parseTime=True"
        }
    default: // sqlite
        dsn = cfg("database.dsn", "database.sqlite")
    }

    Database = DatabaseConfig{Driver: driver, DSN: dsn}
}
```

**Environment variables:**
```env
# SQLite (simplest — great for dev)
DB_DRIVER=sqlite
DB_DSN=database.sqlite

# PostgreSQL
DB_DRIVER=postgres
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=secret
DB_DATABASE=my_app

# MySQL
DB_DRIVER=mysql
DB_HOST=localhost
DB_PORT=3306
DB_USER=root
DB_PASSWORD=secret
DB_DATABASE=my_app

# Or provide a full DSN directly:
DB_DSN=postgres://user:pass@host:5432/dbname?sslmode=disable
```

---

### Authentication Config (`config/auth.go`)

Configure default guard (session vs token) and their settings.

```go
var Auth AuthConfig

type AuthConfig struct {
    DefaultGuard string           // session | token | stateless
    Session      SessionGuardConfig
    Token        TokenGuardConfig
    Stateless    StatelessTokenConfig
}

type SessionGuardConfig struct {
    CookieName string    // Session cookie name
    MaxAge     int       // Session lifetime (seconds)
}

type TokenGuardConfig struct {
    HeaderName string    // Header to read token from
    Scheme     string    // Token prefix (Bearer)
    ExpiresIn  int       // Token lifetime (seconds)
}

type StatelessTokenConfig struct {
    Driver    string         // jwt | paseto
    Secret    string         // Signing secret or PASETO key
    ExpiresIn time.Duration  // Token lifetime
}
```

**Environment variables:**
```env
AUTH_GUARD=session
SESSION_COOKIE=nimbus_session
SESSION_MAX_AGE=604800          # 7 days
TOKEN_EXPIRES_IN=86400          # 1 day

# Stateless (JWT/PASETO)
AUTH_TOKEN_DRIVER=jwt           # jwt | paseto
AUTH_TOKEN_SECRET=your-secret
AUTH_TOKEN_EXPIRES_IN=24h
```

**Real-life example — JWT API auth:**
```env
# Production API config (JWT)
AUTH_GUARD=stateless
AUTH_TOKEN_DRIVER=jwt
AUTH_TOKEN_EXPIRES_IN=1h
```

---

### Cache Config (`config/cache.go`)

```go
var Cache CacheConfig

type CacheConfig struct {
    Driver     string          // memory | redis | memcached | dynamodb
    DefaultTTL time.Duration   // Default TTL for cached items
}
```

**Environment variables:**
```env
CACHE_DRIVER=memory            # memory | redis | memcached | dynamodb
CACHE_TTL_MINUTES=60
```

**Real-life example — Redis cache for production:**
```env
CACHE_DRIVER=redis
CACHE_TTL_MINUTES=30
REDIS_HOST=redis.internal
REDIS_PORT=6379
REDIS_PASSWORD=your-redis-password
```

---

### Session Config (`config/session.go`)

```go
var Session SessionConfig

type SessionConfig struct {
    Driver     string    // cookie | memory
    CookieName string   // Session cookie name
    MaxAge     int       // Lifetime in seconds
    HttpOnly   bool      // Prevent JS access
    Secure     bool      // HTTPS only (auto in production)
    SameSite   string    // strict | lax | none
}
```

**Environment variables:**
```env
SESSION_DRIVER=cookie
SESSION_COOKIE=nimbus_session
SESSION_MAX_AGE=604800
```

---

### Mail Config (`config/mail.go`)

```go
var Mail MailConfig

type MailConfig struct {
    Driver string        // smtp | log | memory
    From   string        // Default sender
    SMTP   SMTPConfig
}

type SMTPConfig struct {
    Host     string
    Port     int
    Username string
    Password string
    From     string
}
```

**Environment variables:**
```env
MAIL_DRIVER=smtp
MAIL_FROM=noreply@myapp.com
SMTP_HOST=smtp.mailtrap.io
SMTP_PORT=587
SMTP_USERNAME=your-mailtrap-username
SMTP_PASSWORD=your-mailtrap-password
```

**Real-life example — Production email with SendGrid:**
```env
MAIL_DRIVER=smtp
MAIL_FROM=hello@myapp.com
SMTP_HOST=smtp.sendgrid.net
SMTP_PORT=587
SMTP_USERNAME=apikey
SMTP_PASSWORD=SG.xxxxx.xxxxx
```

---

### Queue Config (`config/queue.go`)

```go
var Queue QueueConfig

type QueueConfig struct {
    Driver       string    // sync | redis | database | sqs | kafka
    RedisURL     string
    SQSQueueURL  string
    KafkaBrokers string
    KafkaTopic   string
    KafkaGroupID string
}
```

**Environment variables:**
```env
QUEUE_DRIVER=sync
REDIS_URL=redis://localhost:6379
SQS_QUEUE_URL=
KAFKA_BROKERS=
KAFKA_TOPIC=nimbus-queue
KAFKA_GROUP_ID=nimbus-queue
```

---

### Rate Limiter Config (`config/limiter.go`)

```go
var Limiter LimiterConfig

type LimiterConfig struct {
    Enabled       bool
    Requests      int               // Max requests per window
    Window        time.Duration     // Window duration
    KeyFunc       string            // "ip" | "user" | "custom"
    Store         string            // "memory" | "redis"
    RedisURL      string
    Headers       bool              // Send X-RateLimit-* headers
    BlockDuration time.Duration     // How long to block after limit hit
}
```

**Environment variables:**
```env
RATE_LIMIT_ENABLED=true
RATE_LIMIT_REQUESTS=100
RATE_LIMIT_WINDOW_SECONDS=60
RATE_LIMIT_KEY=ip
RATE_LIMIT_STORE=memory
RATE_LIMIT_HEADERS=true
```

---

### Logger Config (`config/logger.go`)

```go
var Logger LoggerConfig

type LoggerConfig struct {
    Level  string    // debug | info | warn | error
    Format string    // json | text
}
```

**Environment variables:**
```env
LOG_LEVEL=info
LOG_FORMAT=json
```

---

### Static Files Config (`config/static.go`)

```go
var Static StaticConfig

type StaticConfig struct {
    Enabled bool    // Enable static file serving
    Root    string  // Directory to serve from
    Prefix  string  // URL prefix
    MaxAge  int     // Cache-Control max-age (seconds)
}
```

**Environment variables:**
```env
STATIC_ENABLED=true
STATIC_ROOT=public
STATIC_PREFIX=/public
STATIC_MAX_AGE=86400
```

---

### Storage Config (`config/storage.go`)

```go
var Storage StorageConfig

type StorageConfig struct {
    Driver string              // local | s3 | gcs | r2
    Local  LocalStorageConfig
}

type LocalStorageConfig struct {
    Root string    // Directory for local file storage
}
```

**Environment variables:**
```env
STORAGE_DRIVER=local
STORAGE_ROOT=storage

# S3
AWS_ACCESS_KEY_ID=AKIA...
AWS_SECRET_ACCESS_KEY=...
AWS_REGION=us-east-1
S3_BUCKET=my-app-uploads

# GCS
GCS_BUCKET=my-app-uploads
GCS_PROJECT_ID=my-project

# Cloudflare R2
R2_ACCOUNT_ID=...
R2_ACCESS_KEY=...
R2_SECRET_KEY=...
R2_BUCKET=my-app-uploads
```

---

### CORS Config (`config/cors.go`)

Cross-Origin Resource Sharing controls which external domains can make requests to your API.

```go
var CORS CORSConfig

type CORSConfig struct {
    Enabled          bool
    AllowOrigins     []string    // ["*"] or specific domains
    AllowMethods     []string    // GET, POST, PUT, etc.
    AllowHeaders     []string
    ExposeHeaders    []string
    AllowCredentials bool
    MaxAge           int         // Preflight cache (seconds)
}
```

**Environment variables:**
```env
CORS_ENABLED=true
CORS_ORIGIN=*
CORS_CREDENTIALS=false
CORS_MAX_AGE=86400
```

> When `AllowCredentials` is `true`, browsers reject the `"*"` origin. Nimbus automatically reflects the requesting origin instead.

---

### Body Parser Config (`config/bodyparser.go`)

Limits and allowed content types for incoming request bodies.

```go
var BodyParser BodyParserConfig

type BodyParserConfig struct {
    JSONLimit      string      // e.g. "1mb"
    FormLimit      string
    MultipartLimit string      // e.g. "10mb" for file uploads
    AllowedTypes   []string
}
```

**Environment variables:**
```env
BODY_JSON_LIMIT=1mb
BODY_FORM_LIMIT=1mb
BODY_MULTIPART_LIMIT=10mb
```

---

### Hash Config (`config/hash.go`)

```go
var Hash HashConfig

type HashConfig struct {
    Driver     string    // bcrypt
    BcryptCost int       // Cost factor (default: 10)
}
```

**Environment variables:**
```env
HASH_DRIVER=bcrypt
HASH_BCRYPT_COST=10
```

---

### Shield Config (`config/shield.go`)

Shield protects your app by setting security HTTP headers and providing CSRF protection.

```go
var Shield ShieldConfig

type ShieldConfig struct {
    ContentTypeNosniff bool
    XSSProtection      string
    FrameGuard         string        // SAMEORIGIN | DENY | ALLOW-FROM
    ReferrerPolicy     string
    CSRF               CSRFConfig
}

type CSRFConfig struct {
    Enabled     bool
    CookieName  string
    HeaderName  string          // X-CSRF-Token
    FieldName   string          // _csrf (form field)
    MaxAge      int
    Secure      bool            // HTTPS only (auto in production)
    SameSite    http.SameSite
    Path        string
    HttpOnly    bool
    ExceptPaths []string        // Routes to skip CSRF (e.g. webhooks)
}
```

**Environment variables:**
```env
CSRF_ENABLED=true
```

> Shield is scaffolded as a base config file on `nimbus new`. When you install the Shield plugin (`nimbus plugin:install shield`), the framework middleware uses these values automatically.

---

### Plugin Configs (scaffolded on install)

These config files are **not** included in the default project. They are scaffolded automatically when you run `nimbus plugin:install <name>`:

| Plugin | Config File | Key Struct | Environment Variables |
|--------|-------------|------------|-----------------------|
| Telescope | `config/telescope.go` | `TelescopeConfig` | `TELESCOPE_ENABLED`, `TELESCOPE_PATH`, `TELESCOPE_MAX_ENTRIES` |
| Horizon | `config/horizon.go` | `HorizonConfig` | `HORIZON_PATH`, `REDIS_URL` |
| Transmit | `config/transmit.go` | `TransmitConfig` | `TRANSMIT_PATH`, `TRANSMIT_PING_INTERVAL`, `TRANSMIT_TRANSPORT` |
| Socialite | `config/socialite.go` | `SocialiteProviders()` | `GITHUB_CLIENT_ID`, `GOOGLE_CLIENT_ID`, etc. |

The `loadXxx()` function is also automatically added to your `config/config.go` on install.

---

## Framework Config API

Beyond the starter's config structs, Nimbus provides a generic config API for reading values:

### Type-Safe Getters

```go
import "github.com/CodeSyncr/nimbus/config"

// Generic getters with type inference
port := config.Get[int]("app.port")               // Returns 0 if not set
name := config.Get[string]("app.name")             // Returns "" if not set
debug := config.Get[bool]("app.debug")             // Returns false if not set

// With defaults
port := config.GetOrDefault[int]("app.port", 3333)
name := config.GetOrDefault[string]("app.name", "nimbus")

// Must — panics if not set (use for required config)
key := config.Must[string]("app.key")  // Panics if APP_KEY not set
```

### Schema-Based Loading

Load config directly into a struct using tags:

```go
type RedisConfig struct {
    Host     string `config:"redis.host" env:"REDIS_HOST" default:"localhost"`
    Port     int    `config:"redis.port" env:"REDIS_PORT" default:"6379"`
    Password string `config:"redis.password" env:"REDIS_PASSWORD"`
    DB       int    `config:"redis.db" env:"REDIS_DB" default:"0"`
}

var redis RedisConfig
config.LoadInto(&redis)
// redis.Host = "localhost" (from default if not set)
```

---

## Real-Life Example: Multi-Environment Setup

### Development (`.env`)
```env
APP_ENV=development
APP_PORT=3333
DB_DRIVER=sqlite
DB_DSN=dev.sqlite
CACHE_DRIVER=memory
LOG_LEVEL=debug
```

### Staging (`.env.staging`)
```env
APP_ENV=staging
APP_PORT=8080
DB_DRIVER=postgres
DB_HOST=staging-db.internal
DB_DATABASE=myapp_staging
CACHE_DRIVER=redis
REDIS_HOST=staging-redis.internal
LOG_LEVEL=info
```

### Production (environment variables set by platform)
```env
APP_ENV=production
APP_PORT=8080
APP_KEY=production-secret-key-must-be-very-long
DB_DRIVER=postgres
DB_DSN=postgres://user:pass@prod-db.internal:5432/myapp?sslmode=require
CACHE_DRIVER=redis
REDIS_HOST=prod-redis.internal
REDIS_PASSWORD=super-secret
LOG_LEVEL=warn
MAIL_DRIVER=smtp
SMTP_HOST=smtp.sendgrid.net
```

### Checking Environment in Code

```go
// In a controller or service
if config.App.Env == "production" {
    // Use strict security settings
    // Don't expose debug info
}

if config.App.Env == "development" {
    // Show detailed error pages
    // Enable Telescope debugging
}
```

---

## Adding Custom Configuration

To add a new config section for your application:

### Step 1: Create the config file

```go
// config/payments.go
package config

var Payments PaymentsConfig

type PaymentsConfig struct {
    Provider       string  // stripe | paypal
    StripeKey      string
    StripeSecret   string
    WebhookSecret  string
    Currency       string
    TaxRate        float64
}

func loadPayments() {
    Payments = PaymentsConfig{
        Provider:      env("PAYMENT_PROVIDER", "stripe"),
        StripeKey:     env("STRIPE_KEY", ""),
        StripeSecret:  env("STRIPE_SECRET", ""),
        WebhookSecret: env("STRIPE_WEBHOOK_SECRET", ""),
        Currency:      env("PAYMENT_CURRENCY", "usd"),
        TaxRate:       envFloat("TAX_RATE", 0.0),
    }
}
```

### Step 2: Register in `config.go`

```go
func Load() {
    _ = nimbusconfig.LoadAuto()

    loadApp()
    loadDatabase()
    // ... existing loaders ...
    loadPayments()    // ← Add your loader
}
```

### Step 3: Use in your code

```go
// app/controllers/checkout.go
func (c *Checkout) ProcessPayment(ctx *http.Context) error {
    if config.Payments.Provider == "stripe" {
        // Use Stripe API with config.Payments.StripeSecret
    }
    // Apply tax
    total := subtotal * (1 + config.Payments.TaxRate)
    // ...
}
```

### Step 4: Add to `.env`

```env
PAYMENT_PROVIDER=stripe
STRIPE_KEY=pk_test_xxxxx
STRIPE_SECRET=sk_test_xxxxx
STRIPE_WEBHOOK_SECRET=whsec_xxxxx
PAYMENT_CURRENCY=usd
TAX_RATE=0.08
```

---

## Helper Functions

The `config/env.go` file provides convenience functions for reading environment variables:

```go
// Read string with default
func env(key, fallback string) string

// Read int with default
func envInt(key string, fallback int) int

// Read float with default  
func envFloat(key string, fallback float64) float64

// Read bool with default
func envBool(key string, fallback bool) bool

// Config store access (dot-notation keys)
func cfg(key, fallback string) string
func cfgInt(key string, fallback int) int
```

---

## Best Practices

1. **Never hardcode secrets** — Always use `.env` and environment variables
2. **Never commit `.env`** — Add it to `.gitignore`; provide `.env.example` with placeholder values
3. **Use typed config structs** — Avoid raw `os.Getenv()` calls scattered through your code
4. **Set sensible defaults** — Every config should work out of the box for local development
5. **Validate required config** — Use `config.ValidateEnv()` to fail fast on missing variables
6. **Group related config** — One file per concern (`mail.go`, `cache.go`, not one giant config)
7. **Document your config** — Add comments explaining what each setting does and valid values

---

## Environment Validation

Validate that required environment variables are set at boot time. If any are missing, the application panics with a clear error message:

```go
import "github.com/CodeSyncr/nimbus/config"

func init() {
    config.ValidateEnv(
        config.Required("APP_KEY"),
        config.Required("DB_DSN"),
        config.Required("REDIS_URL"),
    )
}
// If DB_DSN is missing:
// panic: "environment validation failed: DB_DSN is required"
```

### Custom Rules

Use `EnvRule` for more complex validation beyond presence checks:

```go
config.ValidateEnv(
    config.Required("PORT"),
    config.EnvRule{
        Key:     "APP_ENV",
        Message: "APP_ENV must be development, staging, or production",
        Validate: func(val string) bool {
            return val == "development" || val == "staging" || val == "production"
        },
    },
)
```

Call `ValidateEnv()` early in your boot sequence (e.g. in `config.Load()` or a package `init()` function) to catch missing variables before any service starts.

**Next:** [Routing & Controllers](05-routing-controllers.md) →
