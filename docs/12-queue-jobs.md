# Queue & Jobs

> **Background job processing** — dispatch tasks to queues, process them asynchronously with retries, and monitor with Horizon.

---

## Introduction

Not every task should block the HTTP response. Sending emails, processing images, generating reports, syncing with external APIs — these are all perfect candidates for **background jobs**. Nimbus provides:

- **Job definitions** — Struct-based jobs with `Handle()` and `Failed()` methods
- **Queue dispatching** — `queue.Dispatch(job)` to push jobs to a queue
- **Multiple backends** — In-memory (dev), Redis (production), SQS (cloud)
- **Retry handling** — Automatic retries with configurable attempts
- **Failed job handling** — Custom `Failed()` callback for error recovery
- **Job registration** — Central registry in `start/jobs.go`
- **Horizon dashboard** — Real-time monitoring of queue throughput

---

## Creating a Job

Jobs are structs that implement the `queue.Job` interface:

```go
// app/jobs/send_welcome_email.go
package jobs

import (
    "context"
    "fmt"

    "github.com/CodeSyncr/nimbus/queue"
)

type SendWelcomeEmail struct {
    UserID uint
    Email  string
}

// Handle is called when the job is processed by a worker
func (j *SendWelcomeEmail) Handle(ctx context.Context) error {
    fmt.Printf("[queue] sending welcome email to %s (user_id=%d)\n", j.Email, j.UserID)
    // Send actual email via mail driver
    // mail.Send(j.Email, "Welcome!", "welcome-template", data)
    return nil
}

// Failed is called when the job exhausts all retry attempts (optional)
func (j *SendWelcomeEmail) Failed(ctx context.Context, err error) {
    fmt.Printf("[queue] FAILED welcome email to %s: %v\n", j.Email, err)
    // Log to error tracking service
    // sentry.CaptureException(err)
}

// Ensure interface compliance at compile time
var (
    _ queue.Job       = (*SendWelcomeEmail)(nil)
    _ queue.FailedJob = (*SendWelcomeEmail)(nil)
)
```

---

## Registering Jobs

All jobs must be registered so the queue system can deserialize them:

```go
// start/jobs.go
package start

import (
    "github.com/CodeSyncr/nimbus/queue"
    "nimbus-starter/app/jobs"
)

func RegisterQueueJobs() {
    queue.Register(&jobs.SendWelcomeEmail{})
    queue.Register(&jobs.ProcessPayment{})
    queue.Register(&jobs.GenerateReport{})
    queue.Register(&jobs.SyncInventory{})
    queue.Register(&jobs.ResizeImage{})
}
```

---

## Dispatching Jobs

### Basic Dispatch

```go
import "github.com/CodeSyncr/nimbus/queue"

// Dispatch a job to the default queue
job := &jobs.SendWelcomeEmail{
    UserID: user.ID,
    Email:  user.Email,
}
err := queue.Dispatch(job).Dispatch(ctx)
```

### From a Controller

```go
// app/controllers/auth.go
func (ctrl *AuthController) Register(ctx *http.Context) error {
    // ... validate and create user ...

    // Dispatch welcome email (non-blocking)
    job := &jobs.SendWelcomeEmail{
        UserID: user.ID,
        Email:  user.Email,
    }
    if err := queue.Dispatch(job).Dispatch(ctx.Request.Context()); err != nil {
        // Log but don't fail the registration
        logger.Error("failed to queue welcome email", "error", err)
    }

    return ctx.JSON(http.StatusCreated, map[string]any{
        "user":    user,
        "message": "Registration successful",
    })
}
```

### From a Route Handler

```go
// Queue demo from nimbus-starter
demos.Post("/queue/demo", func(c *http.Context) error {
    job := &jobs.SendWelcomeEmail{
        UserID: 1,
        Email:  "queue-demo@example.com",
    }
    if err := queue.Dispatch(job).Dispatch(c.Request.Context()); err != nil {
        return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to queue job"})
    }
    return c.JSON(http.StatusAccepted, map[string]string{"status": "queued"})
})
```

---

## Queue Boot

The queue system is initialized during application boot:

```go
// bin/server.go
func bootQueue() {
    queue.Boot(&queue.BootConfig{
        RegisterJobs: start.RegisterQueueJobs,
    })
}
```

---

## Configuration

```env
QUEUE_DRIVER=memory          # memory | redis | sqs
QUEUE_DEFAULT=default        # Default queue name
QUEUE_FAILED_TABLE=failed_jobs
```

### Memory Driver (Development)

Jobs are processed in-process. Simple, no external dependencies:

```env
QUEUE_DRIVER=memory
```

### Redis Driver (Production)

Distributed job processing across multiple workers:

```env
QUEUE_DRIVER=redis
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0
```

### SQS Driver (Cloud)

AWS SQS for serverless/cloud deployments:

```env
QUEUE_DRIVER=sqs
AWS_ACCESS_KEY_ID=AKIA...
AWS_SECRET_ACCESS_KEY=...
AWS_REGION=us-east-1
SQS_QUEUE_URL=https://sqs.us-east-1.amazonaws.com/123456/my-queue
```

---

## Real-Life Job Examples

### Order Processing Pipeline

```go
// app/jobs/process_order.go
type ProcessOrder struct {
    OrderID uint
}

func (j *ProcessOrder) Handle(ctx context.Context) error {
    var order models.Order
    if err := db.Preload("Items.Product").First(&order, j.OrderID).Error; err != nil {
        return fmt.Errorf("order not found: %w", err)
    }

    // 1. Verify payment
    if err := verifyPayment(order); err != nil {
        return fmt.Errorf("payment verification failed: %w", err)
    }

    // 2. Update stock
    for _, item := range order.Items {
        if err := db.Model(&models.Product{}).
            Where("id = ?", item.ProductID).
            Update("stock", gorm.Expr("stock - ?", item.Quantity)).Error; err != nil {
            return fmt.Errorf("stock update failed: %w", err)
        }
    }

    // 3. Update order status
    db.Model(&order).Update("status", "processing")

    // 4. Chain: dispatch shipping label generation
    queue.Dispatch(&GenerateShippingLabel{OrderID: order.ID}).Dispatch(ctx)

    // 5. Send confirmation email
    queue.Dispatch(&SendOrderConfirmation{
        OrderID: order.ID,
        Email:   order.User.Email,
    }).Dispatch(ctx)

    return nil
}

func (j *ProcessOrder) Failed(ctx context.Context, err error) {
    // Mark order as failed, notify admin
    db.Model(&models.Order{}).Where("id = ?", j.OrderID).Update("status", "failed")
    notifyAdmin("Order processing failed", j.OrderID, err)
}
```

### Image Processing

```go
// app/jobs/resize_image.go
type ResizeImage struct {
    ImageID  uint
    Sizes    []int  // [100, 300, 600, 1200]
}

func (j *ResizeImage) Handle(ctx context.Context) error {
    var image models.Image
    db.First(&image, j.ImageID)

    for _, size := range j.Sizes {
        resized, err := resize(image.Path, size)
        if err != nil {
            return fmt.Errorf("resize %dpx failed: %w", size, err)
        }

        // Upload to S3
        key := fmt.Sprintf("images/%d/%dpx.jpg", image.ID, size)
        if err := storage.Put(key, resized); err != nil {
            return fmt.Errorf("upload failed: %w", err)
        }
    }

    db.Model(&image).Update("processed", true)
    return nil
}
```

### Report Generation

```go
// app/jobs/generate_report.go
type GenerateMonthlyReport struct {
    Month    int
    Year     int
    TenantID uint
}

func (j *GenerateMonthlyReport) Handle(ctx context.Context) error {
    // Query data
    var orders []models.Order
    db.Where("tenant_id = ? AND EXTRACT(MONTH FROM created_at) = ? AND EXTRACT(YEAR FROM created_at) = ?",
        j.TenantID, j.Month, j.Year,
    ).Preload("Items").Find(&orders)

    // Generate PDF
    pdf := generatePDF(orders)

    // Store in S3
    key := fmt.Sprintf("reports/%d/%d-%02d.pdf", j.TenantID, j.Year, j.Month)
    storage.Put(key, pdf)

    // Email to tenant admin
    queue.Dispatch(&SendReportEmail{
        TenantID: j.TenantID,
        ReportURL: storage.URL(key),
    }).Dispatch(ctx)

    return nil
}
```

### External API Sync

```go
// app/jobs/sync_inventory.go
type SyncInventory struct {
    SupplierID uint
}

func (j *SyncInventory) Handle(ctx context.Context) error {
    supplier := getSupplier(j.SupplierID)

    // Fetch from external API
    products, err := supplier.FetchInventory(ctx)
    if err != nil {
        return fmt.Errorf("api call failed: %w", err)
    }

    // Update local inventory
    for _, p := range products {
        db.Model(&models.Product{}).
            Where("sku = ?", p.SKU).
            Updates(map[string]any{
                "stock": p.Quantity,
                "price": p.Price,
            })
    }

    logger.Info("inventory synced",
        "supplier", supplier.Name,
        "products", len(products),
    )
    return nil
}
```

---

## Horizon (Queue Dashboard)

Monitor your queues in real-time:

```go
import "github.com/CodeSyncr/nimbus/plugins/horizon"

app.Use(horizon.New())
// Dashboard at /horizon
```

Horizon shows:
- **Active jobs** — Currently processing
- **Completed jobs** — Successfully processed
- **Failed jobs** — Jobs that exhausted retries
- **Throughput** — Jobs per minute/hour
- **Wait times** — Time between dispatch and processing

---

## Best Practices

1. **Keep jobs small and focused** — One job, one responsibility
2. **Make jobs idempotent** — Running twice should be safe (in case of retries)
3. **Serialize minimal data** — Store IDs, not full objects
4. **Chain jobs for pipelines** — Dispatch follow-up jobs from `Handle()`
5. **Implement `Failed()`** — Always handle failure gracefully
6. **Use Redis for production** — In-memory is only for development
7. **Monitor with Horizon** — Watch for queue backlogs and failures
8. **Don't queue trivial work** — Only queue tasks that are slow or unreliable

**Next:** [Cache](13-cache.md) →
