# Validation

> **VineJS-inspired, chainable validation rules** — validate request data with expressive schemas, custom messages, and database-aware rules.

---

## Introduction

Nimbus provides a powerful validation system inspired by modern Laravel-style validation ergonomics. It offers:

- **Chainable rule builder** — `validation.String().Required().Min(1).Max(255).Email()`
- **Struct-based validation** — Define rules alongside your data structure
- **Form request pattern** — Combine validation + authorization in one struct
- **Database rules** — `Unique()` and `Exists()` check against your database
- **Custom messages** — Override default error messages per field/rule
- **Schema providers** — Implement `SchemaProvider` interface for auto-validation
- **Request binding** — Decode JSON body → validate → authorize in one call

---

## Quick Start

### Basic Validation

```go
// app/validators/todo.go
package validators

import "github.com/CodeSyncr/nimbus/validation"

type Todo struct {
    Title   string
    Content string
}

// Rules returns the validation schema
func (v *Todo) Rules() validation.Schema {
    return validation.Schema{
        "title": validation.String().Required().Min(1).Max(255).Trim(),
    }
}

// Validate runs the rules against the struct fields
func (v *Todo) Validate() error {
    return validation.ValidateStruct(v)
}
```

### Using in a Controller

```go
// app/controllers/todo.go
func (todo *Todo) Store(ctx *http.Context) error {
    _ = ctx.Request.ParseForm()
    v := &validators.Todo{
        Title: strings.TrimSpace(ctx.Request.FormValue("title")),
    }

    if err := v.Validate(); err != nil {
        // Validation failed — show form with errors
        return ctx.View("apps/todo/form", map[string]any{
            "title": "New Todo",
            "item":  nil,
            "error": "Title is required (1–255 chars)",
        })
    }

    // Validation passed — create the record
    item := &models.Todo{Title: v.Title, Done: false}
    todo.DB.Create(item)
    ctx.Redirect(http.StatusFound, "/demos/todo")
    return nil
}
```

---

## Validation Rules

### String Rules

```go
validation.String()                // Base string rule
    .Required()                    // Must not be empty
    .Min(n)                        // Minimum length
    .Max(n)                        // Maximum length
    .Email()                       // Must be valid email format
    .URL()                         // Must be valid URL
    .Alpha()                       // Letters only (a-zA-Z)
    .AlphaNum()                    // Letters and numbers
    .Trim()                        // Trim whitespace before validation
    .Regex(pattern)                // Must match regex pattern
    .In("opt1", "opt2", "opt3")    // Must be one of the listed values
    .Confirmed()                   // Must match "{field}_confirmation" field
    .Unique(opts)                  // Must be unique in database
    .Exists(opts)                  // Must exist in database
```

### Number Rules

```go
validation.Number()                // Base number rule
    .Required()                    // Must be provided
    .Min(n)                        // Minimum value
    .Max(n)                        // Maximum value
    .Positive()                    // Must be > 0
    .Between(min, max)             // Must be within range
```

### Boolean Rules

```go
validation.Bool()                  // Base boolean rule
    .Required()                    // Must be provided
```

### UInt Rules

```go
validation.UInt()                  // Unsigned integer
    .Required()                    // Must be provided
    .Min(n)                        // Minimum value
    .Max(n)                        // Maximum value
```

---

## Real-Life Validation Schemas

### User Registration

```go
// app/validators/register.go
package validators

import "github.com/CodeSyncr/nimbus/validation"

type Register struct {
    Name                 string
    Email                string
    Password             string
    PasswordConfirmation string
}

func (v *Register) Rules() validation.Schema {
    return validation.Schema{
        "name":     validation.String().Required().Min(2).Max(100).Trim(),
        "email":    validation.String().Required().Email().Unique(validation.UniqueOpts{
            Table:  "users",
            Column: "email",
        }),
        "password": validation.String().Required().Min(8).Max(128).Confirmed(),
    }
}

func (v *Register) Validate() error {
    return validation.ValidateStruct(v)
}

// Custom error messages (optional)
func (v *Register) Messages() map[string]string {
    return map[string]string{
        "name.required":     "Please enter your name",
        "email.required":    "Email address is required",
        "email.email":       "Please enter a valid email address",
        "email.unique":      "This email is already registered",
        "password.required": "Password is required",
        "password.min":      "Password must be at least 8 characters",
        "password.confirmed": "Passwords do not match",
    }
}
```

### Product Creation

```go
// app/validators/product.go
type CreateProduct struct {
    Name        string
    Description string
    Price       float64
    SKU         string
    CategoryID  uint
    Stock       int
}

func (v *CreateProduct) Rules() validation.Schema {
    return validation.Schema{
        "name":        validation.String().Required().Min(2).Max(255),
        "description": validation.String().Max(5000),
        "price":       validation.Number().Required().Positive(),
        "sku":         validation.String().Required().Regex(`^[A-Z]{2}-\d{3}$`).Unique(validation.UniqueOpts{
            Table:  "products",
            Column: "sku",
        }),
        "category_id": validation.UInt().Required().Exists(validation.ExistsOpts{
            Table:  "categories",
            Column: "id",
        }),
        "stock":       validation.Number().Required().Min(0),
    }
}
```

### Update Profile

```go
// app/validators/profile.go
type UpdateProfile struct {
    Name     string
    Email    string
    Bio      string
    Website  string
    Location string
    UserID   uint   // Set by controller, not from request
}

func (v *UpdateProfile) Rules() validation.Schema {
    return validation.Schema{
        "name":     validation.String().Required().Min(1).Max(100).Trim(),
        "email":    validation.String().Required().Email().Unique(validation.UniqueOpts{
            Table:     "users",
            Column:    "email",
            IgnoreID:  v.UserID,  // Exclude current user from unique check
        }),
        "bio":      validation.String().Max(500),
        "website":  validation.String().URL(),
        "location": validation.String().Max(100),
    }
}
```

### Order Placement

```go
// app/validators/order.go
type PlaceOrder struct {
    Items []OrderItemInput
    ShippingAddress string
    PaymentMethod   string
    Notes           string
}

type OrderItemInput struct {
    ProductID uint
    Quantity  int
}

func (v *PlaceOrder) Rules() validation.Schema {
    return validation.Schema{
        "shipping_address": validation.String().Required().Min(10).Max(500),
        "payment_method":   validation.String().Required().In("credit_card", "paypal", "bank_transfer"),
        "notes":            validation.String().Max(1000),
    }
}
```

---

## Form Requests

Form requests combine **validation**, **authorization**, and **data binding** into a single struct:

```go
import "github.com/CodeSyncr/nimbus/validation"

// Define a form request
type CreatePostRequest struct {
    validation.BaseFormRequest[CreatePostPayload]
}

type CreatePostPayload struct {
    Title      string
    Content    string
    CategoryID uint
    Tags       []string
}

func (r *CreatePostRequest) Rules() validation.Schema {
    return validation.Schema{
        "title":       validation.String().Required().Min(5).Max(200),
        "content":     validation.String().Required().Min(50),
        "category_id": validation.UInt().Required(),
    }
}

// Optional: authorization check
func (r *CreatePostRequest) Authorize(c *http.Context) bool {
    user := auth.UserFromContext(c.Request.Context())
    return user != nil && user.(*models.User).Role == "writer"
}
```

### Using Form Requests

```go
func (ctrl *PostController) Store(ctx *http.Context) error {
    req := &CreatePostRequest{}
    payload, validationErrors, err := validation.BindAndValidate[CreatePostPayload](ctx, req)

    if err != nil {
        return ctx.JSON(http.StatusForbidden, map[string]string{"error": "unauthorized"})
    }

    if validationErrors != nil {
        return ctx.JSON(http.StatusUnprocessableEntity, validationErrors.ToMap())
    }

    // payload is validated and authorized
    post := &models.Post{
        Title:      payload.Title,
        Content:    payload.Content,
        CategoryID: payload.CategoryID,
    }
    ctrl.DB.Create(post)

    return ctx.JSON(http.StatusCreated, post)
}
```

---

## Schema-Based Validation

Implement the `SchemaProvider` interface for automatic validation via `BindAndValidateSchema`:

```go
type SchemaProvider interface {
    Rules() Schema
}

// Optional interfaces:
type MessageProvider interface {
    Messages() map[string]string
}

type Authorizer interface {
    Authorize(*http.Context) bool
}

type Preparer interface {
    Prepare()  // Called before validation (e.g., normalize data)
}
```

### Usage

```go
func (ctrl *UserController) Update(ctx *http.Context) error {
    req := &validators.UpdateProfile{}
    if err := validation.BindAndValidateSchema(ctx, req); err != nil {
        if ve, ok := err.(*validation.ValidationErrors); ok {
            return ctx.JSON(http.StatusUnprocessableEntity, ve.ToMap())
        }
        return ctx.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
    }

    // req is validated and populated
    // ...
}
```

---

## Database Validation Rules

### Unique Rule

Check that a value doesn't already exist in the database:

```go
validation.String().Unique(validation.UniqueOpts{
    Table:    "users",      // Database table
    Column:   "email",      // Column to check
    IgnoreID: currentUser.ID, // Exclude this ID (for updates)
})
```

### Exists Rule

Check that a value exists in the database (e.g., foreign key validation):

```go
validation.UInt().Exists(validation.ExistsOpts{
    Table:  "categories", // Database table
    Column: "id",         // Column to check
})
```

### Setup

Database rules require setting the DB connection:

```go
// In bin/server.go, after database boot
validation.SetDB(db)
```

---

## Validation Errors

Validation errors are returned as a `ValidationErrors` map:

```go
type ValidationErrors map[string][]string

// Methods:
errors.Error()   // "Validation failed: title: required, min; email: email"
errors.ToMap()   // map[string][]string{"title": ["required", "min"], "email": ["email"]}
```

### JSON Response Format

```json
{
  "title": ["Title is required", "Title must be at least 5 characters"],
  "email": ["Must be a valid email address"],
  "price": ["Price must be positive"]
}
```

### Handling in Templates

```go
func (ctrl *PostController) Store(ctx *http.Context) error {
    v := &validators.CreatePost{
        Title: ctx.Request.FormValue("title"),
        Content: ctx.Request.FormValue("content"),
    }

    if err := v.Validate(); err != nil {
        return ctx.View("posts/create", map[string]any{
            "errors":  err.(*validation.ValidationErrors).ToMap(),
            "old":     v, // Repopulate form with old input
        })
    }

    // ...
}
```

---

## Real-Life Example: Complete API Validation

```go
// app/validators/auth.go
package validators

import "github.com/CodeSyncr/nimbus/validation"

// Login validator
type Login struct {
    Email    string
    Password string
}

func (v *Login) Rules() validation.Schema {
    return validation.Schema{
        "email":    validation.String().Required().Email(),
        "password": validation.String().Required().Min(1),
    }
}

func (v *Login) Validate() error {
    return validation.ValidateStruct(v)
}

// ResetPassword validator
type ResetPassword struct {
    Email string
}

func (v *ResetPassword) Rules() validation.Schema {
    return validation.Schema{
        "email": validation.String().Required().Email().Exists(validation.ExistsOpts{
            Table:  "users",
            Column: "email",
        }),
    }
}

func (v *ResetPassword) Messages() map[string]string {
    return map[string]string{
        "email.exists": "No account found with this email address",
    }
}

// ChangePassword validator
type ChangePassword struct {
    CurrentPassword      string
    NewPassword          string
    NewPasswordConfirm   string
}

func (v *ChangePassword) Rules() validation.Schema {
    return validation.Schema{
        "current_password": validation.String().Required(),
        "new_password":     validation.String().Required().Min(8).Max(128).Confirmed(),
    }
}
```

### Controller Using Validation

```go
// app/controllers/auth.go
func (ctrl *AuthController) Login(ctx *http.Context) error {
    v := &validators.Login{}
    if err := json.NewDecoder(ctx.Request.Body).Decode(v); err != nil {
        return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "invalid JSON"})
    }

    if err := v.Validate(); err != nil {
        return ctx.JSON(http.StatusUnprocessableEntity, err.(*validation.ValidationErrors).ToMap())
    }

    // Validation passed — authenticate
    user, err := authenticateUser(v.Email, v.Password)
    if err != nil {
        return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid credentials"})
    }

    token := generateToken(user)
    return ctx.JSON(http.StatusOK, map[string]any{
        "token": token,
        "user":  user,
    })
}
```

---

## Conditional Validation

Apply validation rules only when specific conditions are met.

### Field-Based Conditions

Use `When()` to apply a rule only when another field has a specific value:

```go
func (r *CreateUserRequest) Rules() validation.Schema {
    return validation.Schema{
        "role":       validation.String().Required().In("personal", "business", "creator"),
        // company_id is required only when role is "business"
        "company_id": validation.When("role", "business", validation.String().Required()),
        // bio has different rules based on role
        "bio": validation.When("role", "creator", validation.String().Required().Min(10)).
            Otherwise(validation.String().Max(500)),
    }
}
```

### Predicate-Based Conditions

Use `WhenFn()` for custom logic:

```go
func (r *OrderRequest) Rules() validation.Schema {
    return validation.Schema{
        "delivery_type": validation.String().Required().In("digital", "physical"),
        "shipping_address": validation.WhenFn(
            func(data map[string]any) bool {
                return data["delivery_type"] == "physical"
            },
            validation.String().Required().Min(10),
        ),
    }
}
```

### Otherwise

Chain `.Otherwise()` to provide a fallback rule when the condition is not met:

```go
"discount_code": validation.When("is_premium", true, 
    validation.String().Required(),
).Otherwise(
    validation.String().Max(0), // non-premium users can't use discount codes
)
```

---

## Best Practices

1. **Keep validators in `app/validators/`** — One file per resource
2. **Use form requests for APIs** — Combine validation + authorization + binding
3. **Always validate user input** — Never trust client data
4. **Use database rules** — `Unique()` and `Exists()` prevent invalid references
5. **Provide custom messages** — User-friendly error messages improve UX
6. **Validate early** — Check input before doing any database operations
7. **Return 422 for validation errors** — HTTP standard for "Unprocessable Entity"
8. **Repopulate forms on error** — Pass old input back to the view

**Next:** [Authentication & Security](09-auth-security.md) →
