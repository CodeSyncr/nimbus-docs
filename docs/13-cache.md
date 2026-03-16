# Cache

> **Type-safe, multi-driver caching** — keep your app fast with in-memory, Redis, Memcached, DynamoDB, or Cloudflare KV caching, with namespace isolation and lazy-load helpers.

---

## Introduction

Caching is how you turn a 200ms database query into a 1ms memory lookup. Nimbus gives you a unified caching API that works identically across backends — switch from development (memory) to production (Redis) by changing a single environment variable.

Key features:

- **Unified `Store` interface** — `Get`, `Set`, `Delete`, `Remember`
- **Five drivers** — Memory, Redis, Memcached, DynamoDB, Cloudflare Workers KV
- **Type-safe generics** — `RememberT[T]()` returns typed values, no casting
- **Namespace isolation** — `cache.Namespace("users")` scopes keys and enables bulk clear
- **Lazy-load with `Remember`** — Cache-or-compute in a single call
- **ORM hook integration** — Auto-invalidate cache on model changes

---

## Quick Start

```go
import (
    "time"
    "github.com/CodeSyncr/nimbus/cache"
)

// Store a value for 10 minutes
cache.Set("user:42", user, 10*time.Minute)

// Retrieve it
val, ok := cache.Get("user:42")

// Delete it
cache.Delete("user:42")

// Check existence
if cache.Has("user:42") { /* cached */ }
if cache.Missing("user:42") { /* not cached */ }
```

---

## Configuration

```env
CACHE_DRIVER=memory          # memory | redis | memcached | dynamodb | cloudflare
CACHE_DEFAULT_TTL=10m        # Default time-to-live

# Redis
REDIS_URL=redis://localhost:6379

# Memcached
MEMCACHED_SERVERS=localhost:11211

# DynamoDB
CACHE_DYNAMO_TABLE=nimbus-cache
AWS_REGION=us-east-1

# Cloudflare Workers KV
CF_ACCOUNT_ID=...
CF_NAMESPACE_ID=...
CF_API_TOKEN=...
```

### Boot Configuration

```go
// bin/server.go
func bootCache() {
    cache.Boot(&cache.BootConfig{
        Driver:     "redis",
        RedisURL:   "redis://localhost:6379",
        DefaultTTL: 10 * time.Minute,
    })
}
```

The boot function reads environment variables as fallbacks — you can configure via `.env` or programmatically.

---

## Store Interface

Every backend implements this interface:

```go
type Store interface {
    Set(key string, value any, ttl time.Duration) error
    Get(key string) (any, bool)
    Delete(key string) error
    Remember(key string, ttl time.Duration, fn func() (any, error)) (any, error)
}
```

All distributed stores (Redis, Memcached, DynamoDB) JSON-serialize values automatically.

---

## Core API

### Set & Get

```go
// Store with explicit TTL
cache.Set("dashboard:stats", stats, 5*time.Minute)

// Store forever (TTL = 0)
cache.SetForever("config:features", features)

// Retrieve
val, ok := cache.Get("dashboard:stats")
if ok {
    stats := val.(*DashboardStats)
}
```

### Delete

```go
cache.Delete("user:42")
```

### Has & Missing

```go
if cache.Has("session:abc123") {
    // Session exists in cache
}

if cache.Missing("user:42:profile") {
    // Need to load from database
}
```

### Pull (Get and Delete)

Retrieve a value and immediately remove it — perfect for one-time-use data like flash messages or verification codes:

```go
code, ok := cache.Pull("verification:user42")
if ok {
    // Use code, it's already been removed from cache
    verifyEmail(code.(string))
}
```

### Remember (Cache-or-Compute)

The most powerful cache method. Checks if the key exists; if not, calls the function, stores the result, and returns it:

```go
stats, err := cache.Remember("dashboard:stats", 5*time.Minute, func() (any, error) {
    // This only runs if not cached
    return computeExpensiveStats()
})
```

### RememberT (Type-Safe)

Generic version that returns a typed value — no casting needed:

```go
user, err := cache.RememberT[User]("user:42", 10*time.Minute, func() (User, error) {
    var u User
    err := db.First(&u, 42).Error
    return u, err
})
// user is already type User, no casting
```

---

## Drivers

### Memory (Development)

In-process, zero configuration. Not shared across instances:

```go
store := cache.NewMemoryStore()
```

- Zero TTL means no expiry
- Lazy expiration (checked on read)
- Thread-safe with `sync.RWMutex`
- Supports namespace `Clear()` via key prefix scanning

### Redis (Production)

Distributed caching with automatic key prefixing (`nimbus:cache:`):

```go
store := cache.NewRedisStore(redisClient)
// Or with custom prefix:
store := cache.NewRedisStoreWithPrefix(redisClient, "myapp:")
```

- Values are JSON-serialized
- TTL is delegated to Redis
- Supports namespace `Clear()` via `SCAN` + `DELETE`

### Memcached (High-Throughput)

Distributed cache with multi-server support:

```go
store := cache.NewMemcachedStore("localhost:11211")
// Multi-server:
store := cache.NewMemcachedStore("10.0.0.1:11211", "10.0.0.2:11211")
```

- Values are JSON-serialized
- Key limit is 250 bytes (auto-truncated)
- No native namespace `Clear()` support (flushes not scoped)

### DynamoDB (Serverless/Cloud)

AWS-native caching with auto-expiry via DynamoDB TTL:

```go
store := cache.NewDynamoDBStore(awsConfig, "nimbus-cache")
```

**Table schema:**
| Attribute | Type | Description |
|-----------|------|-------------|
| `pk` | String | Partition key (prefixed cache key) |
| `val` | String | JSON-serialized value |
| `ttl` | Number | Unix timestamp for auto-deletion |

Enable TTL on the `ttl` attribute in DynamoDB settings for automatic cleanup.

### Cloudflare Workers KV (Edge)

Global edge caching via Cloudflare:

```go
store := cache.NewCloudflareKVStore(accountID, namespaceID, apiToken)
```

---

## Namespaces

Group related cache entries and clear them together:

```go
usersCache := cache.Namespace("users")

// All operations scoped to "users:" prefix
usersCache.Set("42", userData, 10*time.Minute)  // key: "users:42"
usersCache.Set("99", otherUser, 10*time.Minute) // key: "users:99"

val, ok := usersCache.Get("42")  // reads "users:42"

// Clear ALL entries in the namespace
usersCache.Clear()  // removes all "users:*" keys
```

### Namespace Interface

```go
type NamespaceStore interface {
    Store
    Clear() error  // Remove all entries in this namespace
}
```

---

## Real-Life Examples

### API Response Caching

```go
func (ctrl *ProductController) Index(c *http.Context) error {
    page := c.QueryInt("page", 1)
    key := fmt.Sprintf("products:page:%d", page)

    products, err := cache.RememberT[[]Product](key, 5*time.Minute, func() ([]Product, error) {
        var products []Product
        err := db.Scopes(database.Paginate(page, 20)).
            Preload("Category").
            Find(&products).Error
        return products, err
    })
    if err != nil {
        return err
    }

    return c.JSON(200, products)
}
```

### User Session Data

```go
func getUserProfile(userID uint) (*Profile, error) {
    key := fmt.Sprintf("profile:%d", userID)
    return cache.RememberT[*Profile](key, 30*time.Minute, func() (*Profile, error) {
        var profile Profile
        err := db.Preload("Settings").Preload("Preferences").
            Where("user_id = ?", userID).First(&profile).Error
        return &profile, err
    })
}
```

### Cache Invalidation on Update

```go
func (ctrl *ProductController) Update(c *http.Context) error {
    id := c.Param("id")

    // ... validate and update product ...

    // Invalidate specific cache entry
    cache.Delete(fmt.Sprintf("product:%s", id))

    // Invalidate related listing caches
    productsCache := cache.Namespace("products")
    productsCache.Clear()  // Clear all product listing pages

    return c.JSON(200, product)
}
```

### Rate Limiting with Cache

```go
func checkRateLimit(ip string, limit int) bool {
    key := fmt.Sprintf("ratelimit:%s", ip)
    val, ok := cache.Get(key)
    if !ok {
        cache.Set(key, 1, time.Minute)
        return true  // First request
    }
    count := val.(int)
    if count >= limit {
        return false  // Exceeded
    }
    cache.Set(key, count+1, time.Minute)
    return true
}
```

### Configuration Caching

```go
func getFeatureFlags(tenantID uint) map[string]bool {
    key := fmt.Sprintf("flags:%d", tenantID)
    flags, _ := cache.RememberT[map[string]bool](key, time.Hour, func() (map[string]bool, error) {
        var results []FeatureFlag
        db.Where("tenant_id = ?", tenantID).Find(&results)
        flags := make(map[string]bool)
        for _, f := range results {
            flags[f.Name] = f.Enabled
        }
        return flags, nil
    })
    return flags
}
```

---

## Atomic Locks

Prevent cache stampede and coordinate distributed work with atomic locks:

```go
lock := cache.NewLock("rebuild:dashboard", 30*time.Second)

// Try to acquire the lock
if lock.Acquire() {
    defer lock.Release()
    // Only one process runs this at a time
    rebuildDashboard()
}
```

### Blocking Lock

Wait up to a timeout for the lock to become available:

```go
lock := cache.NewLock("generate:report", time.Minute)

// Block up to 10 seconds waiting for the lock
if lock.Block(10 * time.Second) {
    defer lock.Release()
    generateReport()
} else {
    // Lock not acquired within timeout
    return errors.New("could not acquire lock")
}
```

### Cache-Aware Lock (AtomicLock)

Combine caching with locking to prevent stampede on expensive computations:

```go
result, err := cache.AtomicLock("expensive:query", 5*time.Minute, func() (any, error) {
    // Only ONE goroutine/process computes this
    // Others wait and get the cached result
    return computeExpensiveQuery()
})
```

This is especially useful in horizontally-scaled deployments where multiple instances might try to rebuild the same cache entry simultaneously.

---

## Best Practices

1. **Use `Remember` / `RememberT`** — Eliminates cache stampede and simplifies code
2. **Choose TTL carefully** — Too short wastes compute; too long serves stale data
3. **Invalidate on writes** — Always clear cache when underlying data changes
4. **Use namespaces** — Group related keys for easy bulk invalidation
5. **Use Memory for dev, Redis for prod** — Switch via `CACHE_DRIVER` env var
6. **Don't cache user-specific data globally** — Include user ID in cache keys
7. **Monitor cache hit rates** — Use Telescope to track cache performance
8. **Serialize minimal data** — Don't cache entire database rows if you only need 2 fields
9. **Use atomic locks** — Prevent cache stampede on expensive computations

**Next:** [Mail](14-mail.md) →
