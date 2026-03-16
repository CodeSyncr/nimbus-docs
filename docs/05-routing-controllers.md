# Routing & Controllers

> **Expressive, Laravel-style routing with resource controllers** — define clean URLs, group routes with middleware, and scaffold RESTful APIs in minutes.

---

## Introduction

Routing is the backbone of any web application — it maps HTTP requests to the code that handles them. Nimbus provides a **Chi-based router** enhanced with:

- **Named routes** with URL generation
- **Resource controllers** (7 RESTful actions in one line)
- **Route groups** with shared prefixes and middleware
- **Route parameters** (`:id`, wildcards)
- **Route metadata** for OpenAPI documentation
- **Automatic error handling** (validation → 422, other errors → 500)

All routes are defined in `start/routes.go`, following the convention of keeping route definitions separate from handler logic.

---

## Defining Routes

### Basic Routes

```go
// start/routes.go
package start

import (
    "github.com/CodeSyncr/nimbus"
    "github.com/CodeSyncr/nimbus/http"
)

func RegisterRoutes(app *nimbus.App) {
    // Simple GET route
    app.Router.Get("/", func(c *http.Context) error {
        return c.String(http.StatusOK, "Hello, World!")
    })

    // POST route
    app.Router.Post("/submit", func(c *http.Context) error {
        return c.JSON(http.StatusOK, map[string]string{"status": "received"})
    })

    // All HTTP methods
    app.Router.Get("/users",     listUsers)
    app.Router.Post("/users",    createUser)
    app.Router.Put("/users/:id", updateUser)
    app.Router.Patch("/users/:id", patchUser)
    app.Router.Delete("/users/:id", deleteUser)

    // Match any HTTP method
    app.Router.Any("/webhook", webhookHandler)
}
```

### Available HTTP Methods

| Method | Router Call | Use Case |
|--------|-----------|----------|
| GET | `app.Router.Get(path, handler)` | Read/retrieve data |
| POST | `app.Router.Post(path, handler)` | Create new resources |
| PUT | `app.Router.Put(path, handler)` | Full update of a resource |
| PATCH | `app.Router.Patch(path, handler)` | Partial update |
| DELETE | `app.Router.Delete(path, handler)` | Remove a resource |
| ANY | `app.Router.Any(path, handler)` | Match all methods |

---

## Route Parameters

### Basic Parameters

Use `:param` syntax to capture URL segments:

```go
app.Router.Get("/users/:id", func(c *http.Context) error {
    userID := c.Param("id")  // "42" for /users/42
    return c.JSON(http.StatusOK, map[string]string{"user_id": userID})
})

app.Router.Get("/posts/:postId/comments/:commentId", func(c *http.Context) error {
    postID := c.Param("postId")
    commentID := c.Param("commentId")
    return c.JSON(http.StatusOK, map[string]string{
        "post":    postID,
        "comment": commentID,
    })
})
```

### Wildcard Parameters

Use `*` to capture the rest of the path:

```go
app.Router.Get("/docs/*", func(c *http.Context) error {
    page := c.Param("*")  // "getting-started/installation" for /docs/getting-started/installation
    return c.View("docs/"+page, nil)
})

app.Router.Get("/files/*", func(c *http.Context) error {
    filePath := c.Param("*")  // Full path after /files/
    return serveFile(c, filePath)
})
```

### Query Parameters

```go
app.Router.Get("/search", func(c *http.Context) error {
    query := c.Request.URL.Query().Get("q")        // /search?q=nimbus
    page := c.QueryInt("page", 1)                   // /search?page=3 → 3
    perPage := c.QueryInt("per_page", 20)            // Default 20

    results := searchProducts(query, page, perPage)
    return c.JSON(http.StatusOK, results)
})
```

---

## Route Groups

Groups share a common prefix and middleware, keeping your route definitions DRY:

```go
func RegisterRoutes(app *nimbus.App) {
    // Public routes
    app.Router.Get("/", homeHandler)
    app.Router.Get("/health", healthHandler)

    // API v1 group
    api := app.Router.Group("/api/v1")
    api.Get("/products", listProducts)
    api.Get("/products/:id", showProduct)

    // Authenticated API routes
    auth := api.Group("", authMiddleware)   // No extra prefix, just middleware
    auth.Post("/products", createProduct)
    auth.Put("/products/:id", updateProduct)
    auth.Delete("/products/:id", deleteProduct)

    // Admin routes with prefix AND middleware
    admin := app.Router.Group("/admin", requireAdmin)
    admin.Get("/dashboard", adminDashboard)
    admin.Get("/users", adminUsers)
    admin.Resource("settings", &controllers.Settings{})

    // Demo routes (from nimbus-starter)
    demos := app.Router.Group("/demos")
    demos.Get("/", demosIndexHandler)
    demos.Get("/counter", counterHandler)
    demos.Resource("todo", &controllers.Todo{DB: db})
}
```

### Nested Groups

```go
api := app.Router.Group("/api")

v1 := api.Group("/v1")
v1.Get("/users", v1ListUsers)

v2 := api.Group("/v2")
v2.Get("/users", v2ListUsers)  // Different handler for v2
```

---

## Controllers

Controllers organize related request handlers into a struct. This is the recommended approach for anything beyond trivial routes.

### Creating a Controller

```go
// app/controllers/product.go
package controllers

import (
    "github.com/CodeSyncr/nimbus/http"
    "github.com/CodeSyncr/nimbus"
)

type Product struct {
    DB *nimbus.DB
}

func (p *Product) List(ctx *http.Context) error {
    var products []models.Product
    p.DB.Find(&products)
    return ctx.JSON(http.StatusOK, products)
}

func (p *Product) Show(ctx *http.Context) error {
    id := ctx.Param("id")
    var product models.Product
    if p.DB.First(&product, id).Error != nil {
        return ctx.JSON(http.StatusNotFound, map[string]string{"error": "not found"})
    }
    return ctx.JSON(http.StatusOK, product)
}

func (p *Product) Create(ctx *http.Context) error {
    var input struct {
        Name  string
        Price float64
    }
    if err := json.NewDecoder(ctx.Request.Body).Decode(&input); err != nil {
        return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "invalid JSON"})
    }
    product := &models.Product{Name: input.Name, Price: input.Price}
    p.DB.Create(product)
    return ctx.JSON(http.StatusCreated, product)
}
```

### Registering Controller Routes

```go
func RegisterRoutes(app *nimbus.App) {
    productCtrl := &controllers.Product{
        DB: app.Container.MustMake("db").(*nimbus.DB),
    }

    app.Router.Get("/api/products",     func(c *http.Context) error { return productCtrl.List(c) })
    app.Router.Get("/api/products/:id", func(c *http.Context) error { return productCtrl.Show(c) })
    app.Router.Post("/api/products",    func(c *http.Context) error { return productCtrl.Create(c) })
}
```

---

## Resource Controllers

Resource controllers provide all 7 RESTful CRUD actions in a single line. Implement the `router.ResourceController` interface:

### The ResourceController Interface

```go
type ResourceController interface {
    Index(ctx *http.Context) error    // GET    /resource          — List all
    Create(ctx *http.Context) error   // GET    /resource/create   — Show create form
    Store(ctx *http.Context) error    // POST   /resource          — Create new
    Show(ctx *http.Context) error     // GET    /resource/:id      — Show one
    Edit(ctx *http.Context) error     // GET    /resource/:id/edit — Show edit form
    Update(ctx *http.Context) error   // PUT    /resource/:id      — Update
    Destroy(ctx *http.Context) error  // DELETE /resource/:id      — Delete
}
```

### Implementing a Resource Controller

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
        "title": "Todos",
        "items": items,
    })
}

func (todo *Todo) Create(ctx *http.Context) error {
    return ctx.View("apps/todo/form", map[string]any{
        "title": "New Todo",
        "item":  nil,
    })
}

func (todo *Todo) Store(ctx *http.Context) error {
    _ = ctx.Request.ParseForm()
    title := strings.TrimSpace(ctx.Request.FormValue("title"))
    item := &models.Todo{Title: title, Done: false}
    todo.DB.Create(item)
    ctx.Redirect(http.StatusFound, "/demos/todo")
    return nil
}

func (todo *Todo) Show(ctx *http.Context) error {
    id, _ := strconv.ParseUint(ctx.Param("id"), 10, 64)
    var item models.Todo
    if todo.DB.First(&item, id).Error != nil {
        return ctx.JSON(http.StatusNotFound, map[string]string{"error": "not found"})
    }
    return ctx.View("apps/todo/show", map[string]any{"item": item})
}

func (todo *Todo) Edit(ctx *http.Context) error {
    id, _ := strconv.ParseUint(ctx.Param("id"), 10, 64)
    var item models.Todo
    todo.DB.First(&item, id)
    return ctx.View("apps/todo/form", map[string]any{
        "title": "Edit Todo",
        "item":  item,
    })
}

func (todo *Todo) Update(ctx *http.Context) error {
    id, _ := strconv.ParseUint(ctx.Param("id"), 10, 64)
    var item models.Todo
    todo.DB.First(&item, id)
    _ = ctx.Request.ParseForm()
    title := strings.TrimSpace(ctx.Request.FormValue("title"))
    done := ctx.Request.FormValue("done") == "on"
    todo.DB.Model(&item).Updates(map[string]any{"title": title, "done": done})
    ctx.Redirect(http.StatusFound, "/demos/todo")
    return nil
}

func (todo *Todo) Destroy(ctx *http.Context) error {
    id := ctx.Param("id")
    todo.DB.Delete(&models.Todo{}, id)
    ctx.Redirect(http.StatusFound, "/demos/todo")
    return nil
}
```

### Registering a Resource

```go
// One line registers all 7 routes!
app.Router.Resource("todo", &controllers.Todo{DB: db})

// Generates:
// GET    /todo          → Index
// GET    /todo/create   → Create
// POST   /todo          → Store
// GET    /todo/:id      → Show
// GET    /todo/:id/edit → Edit
// PUT    /todo/:id      → Update
// PATCH  /todo/:id      → Update (also)
// DELETE /todo/:id      → Destroy
```

### Resource Options

```go
// API-only (no Create/Edit form routes)
app.Router.Resource("posts", &controllers.Post{}, router.ApiOnly())

// Only specific actions
app.Router.Resource("logs", &controllers.Log{}, router.Only("index", "show"))

// Exclude specific actions
app.Router.Resource("pages", &controllers.Page{}, router.Except("destroy"))
```

---

## Named Routes & URL Generation

```go
// Name a route
app.Router.Get("/users/:id/profile", showProfile).As("user.profile")

// Generate URLs
url := app.Router.URL("user.profile", "42")  // "/users/42/profile"
```

---

## Route Metadata (OpenAPI)

Annotate routes for automatic OpenAPI spec generation:

```go
app.Router.Post("/api/users", createUser).
    Describe("Create a new user").
    Tag("Users").
    Body(&CreateUserRequest{}).
    Returns(201, &User{}).
    Returns(422, &ValidationErrors{})

app.Router.Get("/api/users/:id", showUser).
    Describe("Get user by ID").
    Tag("Users").
    Returns(200, &User{}).
    Returns(404, nil)
```

---

## HTTP Context

The `*http.Context` provides everything you need for request handling:

### Request Data

```go
func handler(c *http.Context) error {
    // URL parameters
    id := c.Param("id")

    // Query parameters
    page := c.QueryInt("page", 1)
    search := c.Request.URL.Query().Get("q")

    // Request body (JSON)
    var body CreatePostRequest
    json.NewDecoder(c.Request.Body).Decode(&body)

    // Form data
    c.Request.ParseForm()
    name := c.Request.FormValue("name")

    // File uploads
    file, header, err := c.Request.FormFile("avatar")

    // Request headers
    token := c.Request.Header.Get("Authorization")

    // Request-scoped storage
    c.Set("user", currentUser)
    user := c.Get("user")

    return nil
}
```

### Response Methods

```go
func handler(c *http.Context) error {
    // JSON response
    return c.JSON(http.StatusOK, map[string]string{"message": "hello"})

    // String response
    return c.String(http.StatusOK, "plain text")

    // HTML template
    return c.View("home", map[string]any{"title": "Home"})

    // Redirect
    c.Redirect(http.StatusFound, "/dashboard")
    return nil

    // Set status code
    c.Status(http.StatusCreated)
    return c.JSON(http.StatusCreated, created)
}
```

---

## Real-Life Examples

### Example 1: E-Commerce API

```go
func RegisterRoutes(app *nimbus.App) {
    db := app.Container.MustMake("db").(*nimbus.DB)

    // Public storefront
    app.Router.Get("/", homeHandler)
    app.Router.Get("/products", listProductsHandler)
    app.Router.Get("/products/:slug", productDetailHandler)
    app.Router.Get("/categories/:id", categoryHandler)

    // Shopping cart (session-based)
    cart := app.Router.Group("/cart")
    cart.Get("/", cartHandler)
    cart.Post("/add", addToCartHandler)
    cart.Post("/remove", removeFromCartHandler)
    cart.Post("/checkout", checkoutHandler)

    // Customer API (token auth)
    api := app.Router.Group("/api/v1", tokenAuth)
    api.Get("/orders", listOrdersHandler)
    api.Get("/orders/:id", showOrderHandler)
    api.Post("/orders/:id/cancel", cancelOrderHandler)
    api.Get("/profile", profileHandler)
    api.Put("/profile", updateProfileHandler)

    // Admin panel
    admin := app.Router.Group("/admin", requireAdmin)
    admin.Resource("products", &controllers.AdminProduct{DB: db})
    admin.Resource("orders",   &controllers.AdminOrder{DB: db})
    admin.Resource("users",    &controllers.AdminUser{DB: db})
    admin.Get("/dashboard", adminDashboard)
    admin.Get("/reports/sales", salesReportHandler)
}
```

### Example 2: Multi-Tenant SaaS

```go
func RegisterRoutes(app *nimbus.App) {
    // Public marketing site
    app.Router.Get("/", landingPage)
    app.Router.Get("/pricing", pricingPage)
    app.Router.Post("/signup", signupHandler)
    app.Router.Post("/login", loginHandler)

    // Tenant-scoped API
    tenant := app.Router.Group("/api", authMiddleware, tenantMiddleware)
    tenant.Resource("projects",  &controllers.Project{})
    tenant.Resource("tasks",     &controllers.Task{})
    tenant.Resource("members",   &controllers.Member{})

    // Tenant admin
    settings := tenant.Group("/settings", requireOwner)
    settings.Get("/billing", billingHandler)
    settings.Put("/billing", updateBillingHandler)
    settings.Get("/team", teamSettingsHandler)
    settings.Post("/invite", inviteMemberHandler)

    // Webhooks (no auth, but signature verification)
    app.Router.Post("/webhooks/stripe", stripeWebhookHandler)
    app.Router.Post("/webhooks/github", githubWebhookHandler)
}
```

### Example 3: Blog with CMS

```go
func RegisterRoutes(app *nimbus.App) {
    // Public blog
    app.Router.Get("/", blogIndexHandler)
    app.Router.Get("/post/:slug", postHandler)
    app.Router.Get("/category/:slug", categoryHandler)
    app.Router.Get("/tag/:slug", tagHandler)
    app.Router.Get("/search", searchHandler)
    app.Router.Get("/feed.xml", rssFeedHandler)

    // Auth
    app.Router.Get("/login", loginFormHandler)
    app.Router.Post("/login", loginHandler)
    app.Router.Post("/logout", logoutHandler)

    // CMS (authenticated writers)
    cms := app.Router.Group("/cms", requireAuth)
    cms.Resource("posts",      &controllers.CMSPost{})
    cms.Resource("categories", &controllers.CMSCategory{})
    cms.Get("/media", mediaLibraryHandler)
    cms.Post("/media/upload", uploadHandler)
    cms.Get("/analytics", analyticsHandler)
}
```

---

## Mounting Static File Servers

```go
// Serve static files from the "public" directory
fs := http.FileServer(http.Dir("public"))
app.Router.Mount("/public", http.StripPrefix("/public", fs))

// Or configure via config
if config.Static.Enabled {
    fs := http.FileServer(http.Dir(config.Static.Root))
    app.Router.Mount(config.Static.Prefix, http.StripPrefix(config.Static.Prefix, fs))
}
```

---

## Health Check Route

Every Nimbus app should have a health check for load balancers and monitoring:

```go
app.Router.Get("/health", func(c *http.Context) error {
    checker := health.New()
    if database.DB != nil {
        checker.DB(database.DB)
    }
    result := checker.Run(c.Request.Context())

    code := http.StatusOK
    if result.Status != "ok" {
        code = http.StatusServiceUnavailable
    }
    return c.JSON(code, result)
})
// Response: {"status":"ok","checks":[{"name":"database","status":"ok","duration":"2ms"}]}
```

---

## Best Practices

1. **Keep routes.go focused** — Only define routes, not handler logic. Use controllers.
2. **Use resource controllers** for CRUD — One line replaces 7 route definitions.
3. **Group related routes** — Use prefixes and middleware to organize.
4. **Name important routes** — Enables URL generation without hardcoding paths.
5. **Use route metadata** — Add `.Describe()`, `.Tag()` for OpenAPI documentation.
6. **Inject dependencies** — Pass `DB` via struct fields, not global state.
7. **Version your API** — `/api/v1/`, `/api/v2/` groups make breaking changes manageable.
8. **Use route model binding** — Reduce boilerplate by resolving models from URL params automatically.

---

## Route Model Binding

Automatically resolve URL parameters to model instances. Implement the `Bindable` interface on your model:

```go
import "github.com/CodeSyncr/nimbus/router"

type User struct {
    ID    uint
    Name  string
    Email string
}

// RouteKey returns the URL parameter name to match
func (u *User) RouteKey() string { return "id" }

// FindForRoute resolves the model from the parameter value
func (u *User) FindForRoute(value string) (any, error) {
    var user User
    err := db.Where("id = ?", value).First(&user).Error
    return &user, err
}
```

Register the binding on the router:

```go
app.Router.Use(router.BindModel(router.ModelBinding{
    Param:      "id",
    ContextKey: "user",
    Model:      &User{},
}))
```

Now in your handlers the model is already resolved:

```go
app.Router.Get("/users/:id", func(c *http.Context) error {
    user, _ := c.Get("user")
    return c.JSON(200, user.(*User))
})
```

If the model is not found, a `404` is returned automatically.

---

## Typed Param Helpers

Convert route parameters to typed values without boilerplate:

```go
// int (returns 0 on parse error)
id := c.ParamInt("id")

// int64
id64 := c.ParamInt64("id")
```

---

## Fallback Routes

Register a fallback handler for when no route matches — ideal for SPAs or structured API 404s:

```go
// SPA fallback
app.Router.Fallback(func(c *http.Context) error {
    return c.View("index", nil)
})

// API fallback
api := app.Router.Group("/api")
api.Fallback(func(c *http.Context) error {
    return c.JSON(404, map[string]string{
        "error":   "not_found",
        "message": "The requested endpoint does not exist.",
    })
})
```

**Next:** [Database & ORM](06-database.md) →
