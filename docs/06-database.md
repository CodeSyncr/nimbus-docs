# Database & ORM

> **GORM-powered database layer with migrations, seeders, factories, query scopes, and relationships** — everything you need to work with data in a structured, maintainable way.

---

## Introduction

Nimbus uses **GORM** — the most mature and feature-rich ORM for Go — enhanced with Nimbus-specific conventions:

- **Multi-driver support** — PostgreSQL, MySQL, and SQLite with zero config changes
- **Migration system** — Versioned schema changes with up/down support and batch tracking
- **Seeder system** — Populate databases with test or default data
- **Model factories** — Generate fake data for testing with a Faker helper
- **Query builder** — Fluent, chainable queries with GORM's full power
- **Query scopes** — Reusable query modifiers (soft delete, pagination, filtering)
- **Model hooks** — Before/After callbacks for create, update, save, delete
- **Relationships** — HasMany, BelongsTo, HasOne, ManyToMany via GORM tags
- **Serialization** — Control JSON output with Omit/Pick options
- **Transactions** — Managed and manual transaction support
- **Pagination** — Built-in paginator with URL generation

---

## Connecting to the Database

### Configuration

```go
// config/database.go
var Database DatabaseConfig

type DatabaseConfig struct {
    Driver string  // sqlite | postgres | mysql
    DSN    string  // Connection string
}
```

### Boot Sequence

```go
// bin/server.go
func bootDatabase(app *nimbus.App) {
    db, err := database.ConnectWithConfig(database.ConnectConfig{
        Driver: config.Database.Driver,
        DSN:    config.Database.DSN,
    })
    if err != nil {
        fmt.Fprintf(os.Stderr, "Database connection failed: %v\n", err)
        os.Exit(1)
    }

    // Make DB globally available
    nimbus.SetDB(db)
    app.Container.Singleton("db", func() *nimbus.DB {
        return nimbus.GetDB()
    })
}
```

### Environment Variables

```env
# SQLite (development)
DB_DRIVER=sqlite
DB_DSN=database.sqlite

# PostgreSQL (production)
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
```

---

## Models

### Defining Models

Models are Go structs that embed `database.Model` for standard fields:

```go
// app/models/todo.go
package models

import "github.com/CodeSyncr/nimbus/database"

type Todo struct {
    database.Model        // Provides: ID (uint), CreatedAt, UpdatedAt, DeletedAt
    Title string
    Done  bool
}
```

The `database.Model` provides:
| Field | Type | Description |
|-------|------|-------------|
| `ID` | `uint` | Auto-incrementing primary key |
| `CreatedAt` | `time.Time` | Set on creation |
| `UpdatedAt` | `time.Time` | Updated on save |
| `DeletedAt` | `gorm.DeletedAt` | Soft delete (null = not deleted) |

### Real-Life Example: E-Commerce Models

```go
// app/models/user.go
package models

import (
    "time"
    "github.com/CodeSyncr/nimbus/database"
)

type User struct {
    database.Model
    Email       string     `gorm:"uniqueIndex;not null"`
    Password    string     `json:"-"` // Never expose in JSON
    Name        string
    Role        string     `gorm:"default:'customer'"`
    LastLoginAt *time.Time
    Orders      []Order
    Addresses   []Address
}

func (u *User) GetID() string {
    return fmt.Sprintf("%d", u.ID)
}

func (u *User) IsAdmin() bool {
    return u.Role == "admin"
}

// app/models/product.go
type Product struct {
    database.Model
    Name        string   `gorm:"not null"`
    Slug        string   `gorm:"uniqueIndex;not null"`
    Description string
    Price       float64  `gorm:"not null"`
    SKU         string   `gorm:"uniqueIndex"`
    Stock       int      `gorm:"default:0"`
    CategoryID  uint
    Category    Category
    Images      []Image
    Active      bool     `gorm:"default:true"`
}

func (p *Product) InStock() bool {
    return p.Stock > 0
}

func (p *Product) FormattedPrice() string {
    return fmt.Sprintf("$%.2f", p.Price)
}

// app/models/order.go
type Order struct {
    database.Model
    UserID       uint
    User         User
    Items        []OrderItem
    Status       string     `gorm:"default:'pending'"` // pending, paid, shipped, delivered, cancelled
    Total        float64
    ShippingAddr string
    PaidAt       *time.Time
    ShippedAt    *time.Time
}

func (o *Order) IsPaid() bool {
    return o.PaidAt != nil
}

// app/models/order_item.go
type OrderItem struct {
    database.Model
    OrderID   uint
    ProductID uint
    Product   Product
    Quantity  int
    Price     float64 // Price at time of purchase
}
```

---

## CRUD Operations

### Create

```go
// Single record
product := &models.Product{
    Name:  "Wireless Headphones",
    Price: 79.99,
    SKU:   "WH-001",
    Stock: 150,
}
db.Create(product)
// product.ID is now set

// Bulk create
products := []models.Product{
    {Name: "Mouse", Price: 29.99, SKU: "MS-001"},
    {Name: "Keyboard", Price: 49.99, SKU: "KB-001"},
}
db.Create(&products)
```

### Read

```go
// Find by ID
var product models.Product
db.First(&product, 1)           // Find by primary key
db.First(&product, "id = ?", 1) // Explicit query

// Find all
var products []models.Product
db.Find(&products)

// Conditions
db.Where("price > ?", 50).Find(&products)
db.Where("category_id = ? AND active = ?", categoryID, true).Find(&products)

// First matching record
db.Where("sku = ?", "WH-001").First(&product)

// Or conditions
db.Where("name LIKE ?", "%wireless%").Or("name LIKE ?", "%bluetooth%").Find(&products)
```

### Update

```go
// Update specific fields
db.Model(&product).Updates(map[string]any{
    "price": 69.99,
    "stock": 200,
})

// Update single field
db.Model(&product).Update("price", 69.99)

// Update with struct (only non-zero fields)
db.Model(&product).Updates(models.Product{Price: 69.99, Stock: 200})

// Bulk update
db.Model(&models.Product{}).Where("category_id = ?", 5).Update("active", false)
```

### Delete

```go
// Soft delete (sets DeletedAt)
db.Delete(&product)

// Delete by ID
db.Delete(&models.Product{}, 1)

// Bulk delete
db.Where("stock = 0").Delete(&models.Product{})

// Hard delete (permanent)
db.Unscoped().Delete(&product)
```

---

## Query Builder

Nimbus wraps GORM with a fluent query builder:

```go
import "github.com/CodeSyncr/nimbus/database"

// Fluent queries
query := database.From(&models.Product{})
products, err := query.
    Where("active", true).
    Where("price >", 20).
    OrderBy("created_at", "desc").
    Limit(10).
    Get()

// Single record
product, err := database.QueryFor(&models.Product{}).
    Where("sku", "WH-001").
    First()

// Select specific columns
products, err := database.From(&models.Product{}).
    Select("id", "name", "price").
    Where("category_id", 5).
    Get()
```

---

## Query Scopes

Reusable query modifiers for common patterns:

```go
import "github.com/CodeSyncr/nimbus/database"

// Built-in scopes
db.Scopes(database.LatestScope()).Find(&products)      // ORDER BY created_at DESC
db.Scopes(database.OldestScope()).Find(&products)      // ORDER BY created_at ASC
db.Scopes(database.LimitScope(10)).Find(&products)     // LIMIT 10

// Conditional scope
db.Scopes(database.WhenScope(
    minPrice > 0,
    database.WhereScope("price >= ?", minPrice),
)).Find(&products)

// Soft delete scopes
db.Scopes(database.WithTrashed).Find(&products)       // Include soft-deleted
db.Scopes(database.OnlyTrashed).Find(&products)       // Only soft-deleted
database.Restore(db, &product)                          // Restore soft-deleted
database.ForceDelete(db, &product)                      // Permanent delete

// Custom scopes
func ActiveProducts(db *gorm.DB) *gorm.DB {
    return db.Where("active = ? AND stock > ?", true, 0)
}

func PriceRange(min, max float64) func(*gorm.DB) *gorm.DB {
    return func(db *gorm.DB) *gorm.DB {
        return db.Where("price BETWEEN ? AND ?", min, max)
    }
}

// Usage
db.Scopes(ActiveProducts, PriceRange(20, 100)).Find(&products)
```

### Advanced Query Helpers

```go
// Chunk large datasets (memory-efficient processing)
database.Chunk[models.Product](db, 100, func(products []models.Product) error {
    for _, p := range products {
        processProduct(p)
    }
    return nil
})

// Check existence
exists := database.Exists(db, &models.Product{}, "sku = ?", "WH-001")

// First or create
product, err := database.FirstOrCreate[models.Product](db, 
    models.Product{SKU: "WH-001"},                    // Search criteria
    models.Product{Name: "Headphones", Price: 79.99},  // Create with these if not found
)

// Update or create (upsert)
product, err := database.UpdateOrCreate[models.Product](db,
    models.Product{SKU: "WH-001"},                    // Match criteria
    models.Product{Price: 69.99, Stock: 200},          // Update/create values
)

// Pluck single column
names, err := database.Pluck[string](db, &models.Product{}, "name")

// Count with grouping
counts, err := database.CountBy(db, &models.Product{}, "category_id")
```

---

## Relationships

### Defining Relationships

```go
// One-to-Many: User has many Orders
type User struct {
    database.Model
    Name   string
    Orders []Order
}

// Belongs-To: Order belongs to User
type Order struct {
    database.Model
    UserID uint
    User   User
    Items  []OrderItem
}

// Many-to-Many: Product has many Tags (via join table)
type Product struct {
    database.Model
    Name string
    Tags []Tag
}

type Tag struct {
    database.Model
    Name     string
    Products []Product
}
```

### Eager Loading (Preloading)

```go
// Load user with orders
var user models.User
db.Preload("Orders").First(&user, 1)

// Nested preloading
db.Preload("Orders.Items.Product").First(&user, 1)

// Conditional preloading
db.Preload("Orders", "status = ?", "paid").First(&user, 1)

// Multiple preloads
db.Preload("Orders").Preload("Addresses").First(&user, 1)

// Using Nimbus helper
database.Preload(db, "Orders").First(&user, 1)
```

### Real-Life Example: Loading an Order with All Details

```go
func (ctrl *OrderController) Show(ctx *http.Context) error {
    id := ctx.Param("id")
    var order models.Order

    err := ctrl.DB.
        Preload("User").
        Preload("Items.Product").
        Preload("Items.Product.Images").
        First(&order, id).Error

    if err != nil {
        return ctx.JSON(http.StatusNotFound, map[string]string{"error": "order not found"})
    }

    return ctx.JSON(http.StatusOK, order)
}
// Returns:
// {
//   "id": 42,
//   "user": {"id": 1, "name": "John", "email": "john@example.com"},
//   "items": [
//     {
//       "product": {"name": "Wireless Headphones", "price": 79.99, "images": [...]},
//       "quantity": 2,
//       "price": 79.99
//     }
//   ],
//   "total": 159.98,
//   "status": "paid"
// }
```

---

## Migrations

### Creating Migrations

```go
// database/migrations/registry.go
package migrations

import (
    "gorm.io/gorm"
    "github.com/CodeSyncr/nimbus/database"
    "my-app/app/models"
)

func All() []database.Migration {
    return []database.Migration{
        {
            Name: "001_create_users_table",
            Up: func(db *gorm.DB) error {
                return db.AutoMigrate(&models.User{})
            },
            Down: func(db *gorm.DB) error {
                return db.Migrator().DropTable("users")
            },
        },
        {
            Name: "002_create_products_table",
            Up: func(db *gorm.DB) error {
                return db.AutoMigrate(&models.Product{})
            },
            Down: func(db *gorm.DB) error {
                return db.Migrator().DropTable("products")
            },
        },
        {
            Name: "003_create_orders_table",
            Up: func(db *gorm.DB) error {
                if err := db.AutoMigrate(&models.Order{}); err != nil {
                    return err
                }
                return db.AutoMigrate(&models.OrderItem{})
            },
            Down: func(db *gorm.DB) error {
                db.Migrator().DropTable("order_items")
                return db.Migrator().DropTable("orders")
            },
        },
        {
            Name: "004_add_index_to_products_sku",
            Up: func(db *gorm.DB) error {
                return db.Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_products_sku ON products(sku)").Error
            },
            Down: func(db *gorm.DB) error {
                return db.Exec("DROP INDEX IF EXISTS idx_products_sku").Error
            },
        },
    }
}
```

### Running Migrations

```bash
# Run all pending migrations
go run main.go migrate

# Or via CLI
nimbus db:migrate
```

The migrator tracks executed migrations in a `schema_migrations` table and runs them in batches, allowing rollback of the last batch.

---

## Seeders

### Creating Seeders

```go
// database/seeders/seeders.go
package seeders

import (
    "gorm.io/gorm"
    "github.com/CodeSyncr/nimbus/database"
    "my-app/app/models"
)

func All() []database.Seeder {
    return []database.Seeder{
        database.SeedFunc(func(db *gorm.DB) error {
            users := []models.User{
                {Name: "Admin", Email: "admin@example.com", Password: "$2a$10$...", Role: "admin"},
                {Name: "John Doe", Email: "john@example.com", Password: "$2a$10$...", Role: "customer"},
            }
            return db.Create(&users).Error
        }),
        database.SeedFunc(func(db *gorm.DB) error {
            categories := []models.Category{
                {Name: "Electronics", Slug: "electronics"},
                {Name: "Clothing", Slug: "clothing"},
                {Name: "Books", Slug: "books"},
            }
            return db.Create(&categories).Error
        }),
        database.SeedFunc(func(db *gorm.DB) error {
            products := []models.Product{
                {Name: "Wireless Headphones", Slug: "wireless-headphones", Price: 79.99, SKU: "WH-001", Stock: 150, CategoryID: 1},
                {Name: "Bluetooth Speaker", Slug: "bluetooth-speaker", Price: 49.99, SKU: "BS-001", Stock: 75, CategoryID: 1},
                {Name: "Go Programming", Slug: "go-programming", Price: 39.99, SKU: "BK-001", Stock: 500, CategoryID: 3},
            }
            return db.Create(&products).Error
        }),
    }
}
```

### Running Seeders

```bash
go run main.go seed
# Or: nimbus db:seed
```

---

## Model Factories

Generate fake data for testing:

```go
import "github.com/CodeSyncr/nimbus/database"

// Define a factory
factory := database.NewFactory()
factory.Define("users", func(f *database.Faker) map[string]any {
    return map[string]any{
        "name":     f.Sentence(2),
        "email":    f.Email(),
        "password": "$2a$10$hashedpassword",
        "role":     "customer",
    }
})

// Create a single record
user := factory.Create("users")

// Create multiple records
users := factory.CreateMany("users", 50)

// Override specific fields
admin := factory.Merge(map[string]any{"role": "admin"}).Create("users")
```

### Faker Helpers

```go
f.Sentence(wordCount)     // "Lorem ipsum dolor"
f.Paragraph(sentenceCount) // Multi-sentence paragraph
f.Email()                  // "random@example.com"
f.Word()                   // "nimbus"
f.Int(min, max)            // Random integer
```

---

## Model Hooks

Execute code before/after database operations:

```go
import "github.com/CodeSyncr/nimbus/database"

database.RegisterHooks(db, "products", database.Hooks{
    BeforeCreate: func(db *gorm.DB) error {
        // Auto-generate slug from name
        product := db.Statement.Dest.(*models.Product)
        if product.Slug == "" {
            product.Slug = slugify(product.Name)
        }
        return nil
    },
    AfterCreate: func(db *gorm.DB) error {
        // Send notification
        product := db.Statement.Dest.(*models.Product)
        notifyAdmins("New product created: " + product.Name)
        return nil
    },
    BeforeUpdate: func(db *gorm.DB) error {
        // Validate price changes
        return nil
    },
    AfterDelete: func(db *gorm.DB) error {
        // Clean up related files
        return nil
    },
})
```

---

## Transactions

### Managed Transactions

```go
import "github.com/CodeSyncr/nimbus/database"

err := database.Transaction(db, func(tx *gorm.DB) error {
    // Create order
    order := &models.Order{UserID: userID, Total: total}
    if err := tx.Create(order).Error; err != nil {
        return err  // Auto-rollback
    }

    // Create order items
    for _, item := range cartItems {
        orderItem := &models.OrderItem{
            OrderID:   order.ID,
            ProductID: item.ProductID,
            Quantity:  item.Quantity,
            Price:     item.Price,
        }
        if err := tx.Create(orderItem).Error; err != nil {
            return err  // Auto-rollback
        }

        // Decrease stock
        if err := tx.Model(&models.Product{}).
            Where("id = ? AND stock >= ?", item.ProductID, item.Quantity).
            Update("stock", gorm.Expr("stock - ?", item.Quantity)).Error; err != nil {
            return err  // Auto-rollback if insufficient stock
        }
    }

    return nil  // Auto-commit
})
```

### Manual Transactions

```go
tx := db.Begin()

if err := tx.Create(&order).Error; err != nil {
    tx.Rollback()
    return err
}

if err := tx.Create(&items).Error; err != nil {
    tx.Rollback()
    return err
}

tx.Commit()
```

---

## Pagination

```go
import "github.com/CodeSyncr/nimbus/database"

func (ctrl *ProductController) Index(ctx *http.Context) error {
    page := ctx.QueryInt("page", 1)
    perPage := ctx.QueryInt("per_page", 20)

    var products []models.Product
    paginator, err := database.Paginate(ctrl.DB, &products, page, perPage)
    if err != nil {
        return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
    }

    return ctx.JSON(http.StatusOK, map[string]any{
        "data":       products,
        "pagination": paginator,
    })
}
// Response:
// {
//   "data": [...products...],
//   "pagination": {
//     "current_page": 1,
//     "per_page": 20,
//     "total": 150,
//     "last_page": 8,
//     "has_more": true
//   }
// }
```

---

## Serialization

Control what fields are included in JSON output:

```go
import "github.com/CodeSyncr/nimbus/database"

// Omit sensitive fields
result := database.Serialize(user, database.SerializeOptions{
    Omit: []string{"Password", "CreatedAt", "UpdatedAt", "DeletedAt"},
})

// Pick only specific fields
result := database.Serialize(user, database.SerializeOptions{
    Pick: []string{"ID", "Name", "Email"},
})

// Check if a field has been modified
if database.IsDirty(product, "price") {
    notifyPriceChange(product)
}
```

---

## Auto-Create Database

Nimbus can automatically create the database if it doesn't exist:

```go
database.CreateDatabaseIfNotExists(database.ConnectConfig{
    Driver: "postgres",
    DSN:    "host=localhost port=5432 user=postgres password=secret dbname=my_new_app sslmode=disable",
})
```

Works with PostgreSQL and MySQL. Useful for development and CI/CD pipelines.

---

## Best Practices

1. **Use `database.Model` for all models** — Consistent ID, timestamps, and soft delete support
2. **Keep models in `app/models/`** — One file per model, named after the entity
3. **Use migrations for schema changes** — Never use `AutoMigrate` in production
4. **Use scopes for reusable queries** — Don't repeat `WHERE active = true` everywhere
5. **Preload relationships explicitly** — Avoid N+1 queries with `Preload()`
6. **Use transactions for multi-step operations** — Maintain data consistency
7. **Use factories for testing** — Generate realistic test data programmatically
8. **Never expose passwords in JSON** — Use `json:"-"` tag on sensitive fields

**Next:** [Middleware](07-middleware.md) →
