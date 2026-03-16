# Testing

> **Comprehensive testing support** — HTTP tests, database helpers, and AI-generated test suites for confident deployments.

---

## Introduction

Nimbus embraces Go's standard `testing` package while adding framework-specific helpers for HTTP tests, database setup/teardown, and factory-based data generation. Combined with the AI test generator, you can achieve high test coverage with minimal effort.

Features:

- **HTTP test helpers** — Test routes without starting a server
- **Database test helpers** — Transaction-based isolation, auto-rollback
- **Model factories** — Generate test data with Faker
- **AI test generation** — `nimbus test:generate` creates tests from controllers
- **Standard Go testing** — No special test runner, uses `go test`

---

## Quick Start

### Basic HTTP Test

```go
// app/controllers/hello_test.go
package controllers_test

import (
    "net/http"
    "net/http/httptest"
    "testing"

    "nimbus-starter/start"
)

func TestHelloWorld(t *testing.T) {
    // Set up the app
    app := setupTestApp()

    // Create a test request
    req := httptest.NewRequest("GET", "/", nil)
    rec := httptest.NewRecorder()

    // Execute
    app.ServeHTTP(rec, req)

    // Assert
    if rec.Code != http.StatusOK {
        t.Errorf("expected 200, got %d", rec.Code)
    }

    if !strings.Contains(rec.Body.String(), "Hello") {
        t.Error("response should contain 'Hello'")
    }
}
```

---

## Test Setup

### Application Setup

Create a shared test helper that boots a minimal app:

```go
// testing/setup.go
package testing

import (
    "github.com/CodeSyncr/nimbus"
    "nimbus-starter/start"
    "nimbus-starter/config"
)

func SetupTestApp() *nimbus.App {
    // Load test config
    os.Setenv("APP_ENV", "testing")
    os.Setenv("DB_DRIVER", "sqlite")
    os.Setenv("DB_DATABASE", ":memory:")

    cfg := config.LoadConfig()
    app := nimbus.New(cfg)

    // Register routes
    start.RegisterRoutes(app)

    return app
}
```

### Database Setup with Transactions

Wrap each test in a transaction that rolls back automatically:

```go
func TestWithDatabase(t *testing.T) {
    db := setupTestDB()

    // Start a transaction
    tx := db.Begin()
    defer tx.Rollback() // Auto-rollback after test

    // Create test data within transaction
    tx.Create(&models.User{
        Name:  "Test User",
        Email: "test@example.com",
    })

    // Run assertions against tx
    var count int64
    tx.Model(&models.User{}).Count(&count)
    if count != 1 {
        t.Errorf("expected 1 user, got %d", count)
    }
    // Transaction rolls back — database stays clean
}
```

---

## HTTP Testing

### GET Request

```go
func TestListProducts(t *testing.T) {
    app := SetupTestApp()

    req := httptest.NewRequest("GET", "/api/products", nil)
    req.Header.Set("Accept", "application/json")
    rec := httptest.NewRecorder()

    app.ServeHTTP(rec, req)

    if rec.Code != 200 {
        t.Fatalf("expected 200, got %d", rec.Code)
    }

    var result []map[string]any
    json.NewDecoder(rec.Body).Decode(&result)

    if len(result) == 0 {
        t.Error("expected products in response")
    }
}
```

### POST Request with JSON Body

```go
func TestCreateProduct(t *testing.T) {
    app := SetupTestApp()

    body := `{"name": "Widget", "price": 29.99, "stock": 100}`
    req := httptest.NewRequest("POST", "/api/products", strings.NewReader(body))
    req.Header.Set("Content-Type", "application/json")
    rec := httptest.NewRecorder()

    app.ServeHTTP(rec, req)

    if rec.Code != 201 {
        t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
    }

    var product map[string]any
    json.NewDecoder(rec.Body).Decode(&product)

    if product["name"] != "Widget" {
        t.Errorf("expected name 'Widget', got '%s'", product["name"])
    }
}
```

### PUT/PATCH Request

```go
func TestUpdateProduct(t *testing.T) {
    app := SetupTestApp()

    body := `{"name": "Updated Widget", "price": 39.99}`
    req := httptest.NewRequest("PUT", "/api/products/1", strings.NewReader(body))
    req.Header.Set("Content-Type", "application/json")
    rec := httptest.NewRecorder()

    app.ServeHTTP(rec, req)

    if rec.Code != 200 {
        t.Fatalf("expected 200, got %d", rec.Code)
    }
}
```

### DELETE Request

```go
func TestDeleteProduct(t *testing.T) {
    app := SetupTestApp()

    req := httptest.NewRequest("DELETE", "/api/products/1", nil)
    rec := httptest.NewRecorder()

    app.ServeHTTP(rec, req)

    if rec.Code != 200 && rec.Code != 204 {
        t.Fatalf("expected 200 or 204, got %d", rec.Code)
    }
}
```

### Authenticated Requests

```go
func TestProtectedRoute(t *testing.T) {
    app := SetupTestApp()

    // Without auth — should be 401
    req := httptest.NewRequest("GET", "/api/profile", nil)
    rec := httptest.NewRecorder()
    app.ServeHTTP(rec, req)

    if rec.Code != 401 {
        t.Errorf("expected 401 without auth, got %d", rec.Code)
    }

    // With auth token
    req = httptest.NewRequest("GET", "/api/profile", nil)
    req.Header.Set("Authorization", "Bearer test-token-123")
    rec = httptest.NewRecorder()
    app.ServeHTTP(rec, req)

    if rec.Code != 200 {
        t.Errorf("expected 200 with auth, got %d", rec.Code)
    }
}
```

---

## Model Factories

Use the framework's factory system for generating test data:

```go
import "github.com/CodeSyncr/nimbus/database"

// Define a factory
var UserFactory = database.NewFactory(func(f *database.Faker) models.User {
    return models.User{
        Name:  f.Name(),
        Email: f.Email(),
        Age:   f.IntBetween(18, 65),
    }
})

// Use in tests
func TestUserList(t *testing.T) {
    db := setupTestDB()
    tx := db.Begin()
    defer tx.Rollback()

    // Create 10 users
    users := UserFactory.CreateMany(tx, 10)

    var count int64
    tx.Model(&models.User{}).Count(&count)
    if count != 10 {
        t.Errorf("expected 10 users, got %d", count)
    }
}
```

### Factory with Overrides

```go
// Create with specific attributes
admin := UserFactory.Create(tx, func(u *models.User) {
    u.Name = "Admin"
    u.Email = "admin@example.com"
    u.Role = "admin"
})
```

---

## Table-Driven Tests

Go's standard pattern for comprehensive test coverage:

```go
func TestValidation(t *testing.T) {
    tests := []struct {
        name       string
        input      string
        wantStatus int
        wantError  string
    }{
        {
            name:       "valid product",
            input:      `{"name": "Widget", "price": 29.99}`,
            wantStatus: 201,
        },
        {
            name:       "missing name",
            input:      `{"price": 29.99}`,
            wantStatus: 422,
            wantError:  "name is required",
        },
        {
            name:       "negative price",
            input:      `{"name": "Widget", "price": -5}`,
            wantStatus: 422,
            wantError:  "price must be positive",
        },
        {
            name:       "empty body",
            input:      `{}`,
            wantStatus: 422,
        },
    }

    app := SetupTestApp()

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            req := httptest.NewRequest("POST", "/api/products", strings.NewReader(tt.input))
            req.Header.Set("Content-Type", "application/json")
            rec := httptest.NewRecorder()

            app.ServeHTTP(rec, req)

            if rec.Code != tt.wantStatus {
                t.Errorf("status: got %d, want %d", rec.Code, tt.wantStatus)
            }

            if tt.wantError != "" && !strings.Contains(rec.Body.String(), tt.wantError) {
                t.Errorf("body should contain %q, got %s", tt.wantError, rec.Body.String())
            }
        })
    }
}
```

---

## AI Test Generation

Let AI generate tests automatically:

```bash
nimbus test:generate
# or
nimbus tg
```

This scans your controllers and generates comprehensive tests:

```go
// Generated by Nimbus AI Test Generator
func TestTodoController_Index(t *testing.T) {
    app := SetupTestApp()

    t.Run("returns list of todos", func(t *testing.T) {
        req := httptest.NewRequest("GET", "/todos", nil)
        rec := httptest.NewRecorder()
        app.ServeHTTP(rec, req)

        assert(t, rec.Code, 200)
    })
}

func TestTodoController_Store(t *testing.T) {
    app := SetupTestApp()

    t.Run("creates todo with valid data", func(t *testing.T) {
        body := `{"title": "Buy groceries"}`
        req := httptest.NewRequest("POST", "/todos", strings.NewReader(body))
        req.Header.Set("Content-Type", "application/json")
        rec := httptest.NewRecorder()
        app.ServeHTTP(rec, req)

        assert(t, rec.Code, 201)
    })

    t.Run("rejects empty title", func(t *testing.T) {
        body := `{"title": ""}`
        req := httptest.NewRequest("POST", "/todos", strings.NewReader(body))
        req.Header.Set("Content-Type", "application/json")
        rec := httptest.NewRecorder()
        app.ServeHTTP(rec, req)

        assert(t, rec.Code, 422)
    })
}
```

---

## Running Tests

```bash
# Run all tests
go test ./...

# Run tests with verbose output
go test -v ./...

# Run specific test
go test -v -run TestCreateProduct ./app/controllers/

# Run with race detection
go test -race ./...

# Run with coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

---

## Best Practices

1. **Use table-driven tests** — Cover multiple cases systematically
2. **Isolate with transactions** — Wrap each test in `tx.Begin()` / `tx.Rollback()`
3. **Use factories for test data** — Consistent, realistic data generation
4. **Test both success and failure** — Validate error responses too
5. **Use `t.Run()` for subtests** — Better output and selective execution
6. **Run with `-race`** — Catch concurrency bugs early
7. **Generate with AI, then customize** — Use `nimbus tg` as a starting point
8. **Test middleware separately** — Unit test middleware independent of routes
9. **Use environment variables** — `APP_ENV=testing` for test-specific config
10. **Keep tests fast** — Use SQLite `:memory:` for database tests

**Next:** [Deployment](20-deployment.md) →
