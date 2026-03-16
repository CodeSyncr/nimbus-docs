# Scheduler

> **Cron-like task scheduling** — define recurring tasks in Go with named tasks, panic recovery, daily-at scheduling, and a single CLI command.

---

## Introduction

Instead of managing crontab entries on your server, Nimbus lets you define scheduled tasks right in your application code. Define tasks with fluent intervals like `EveryMinute()`, `Hourly()`, or `Daily("03:00")`, and run them all with a single command.

Features:

- **Named tasks** — Every task has a name for identification and listing
- **Fluent interval API** — `EveryMinute`, `EveryFiveMinutes`, `Hourly`, `Daily` (with time-of-day)
- **Panic recovery** — Tasks recover from panics automatically
- **Non-blocking** — `Start(ctx)` runs in background, `Stop()` gracefully shuts down
- **Context-aware** — Tasks receive `context.Context` for cancellation
- **CLI integration** — `schedule:run` to start, `schedule:list` to inspect

> **Note:** The `scheduler/` package is deprecated. Use `schedule/` instead. The old package is a thin compatibility wrapper that delegates to `schedule/`.

---

## Defining Tasks

Tasks are registered in `start/schedule.go`:

```go
// start/schedule.go
package start

import (
    "context"
    "fmt"
    "time"

    "github.com/CodeSyncr/nimbus/schedule"
)

func RegisterSchedule(s *schedule.Scheduler) {
    // Run every minute
    s.EveryMinute("health-check", func(ctx context.Context) error {
        fmt.Println("[schedule] health check running...")
        return pingHealthEndpoint()
    })

    // Run every 5 minutes
    s.EveryFiveMinutes("cache-cleanup", func(ctx context.Context) error {
        return cleanUpExpiredCache()
    })

    // Run every hour
    s.Hourly("session-cleanup", func(ctx context.Context) error {
        return cleanUpExpiredSessions()
    })

    // Run daily at 3:00 AM
    s.Daily("03:00", "daily-report", func(ctx context.Context) error {
        return generateDailyReport()
    })

    // Run at custom interval
    s.Every(30*time.Second, "heartbeat", func(ctx context.Context) error {
        return pingExternalService(ctx)
    })
}
```

---

## Scheduler API

### Creating a Scheduler

```go
import "github.com/CodeSyncr/nimbus/schedule"

s := schedule.New()
```

### Interval Methods

| Method | Interval | Use Case |
|--------|----------|----------|
| `Every(duration, name, fn)` | Custom interval | Fine-grained control |
| `EveryMinute(name, fn)` | 1 minute | Health checks, queue monitoring |
| `EveryFiveMinutes(name, fn)` | 5 minutes | Cache cleanup |
| `Hourly(name, fn)` | 1 hour | Stats aggregation, session cleanup |
| `Daily(at, name, fn)` | 24 hours at specific time | Reports, backups, billing |

Each method accepts a **name** (`string`) and a function `func(ctx context.Context) error`.

The `Daily` method also accepts an `at` parameter in `"HH:MM"` format (24-hour, local timezone).

### Running & Stopping

```go
// Start the scheduler (non-blocking, runs in background)
s.Start(ctx)

// Stop the scheduler gracefully
s.Stop()
```

### Counting Tasks

```go
count := s.Count()
fmt.Printf("Scheduler has %d registered tasks\n", count)
```

---

## Running the Scheduler

### Via CLI

```bash
# Start the scheduler (runs all registered tasks)
go run main.go schedule:run

# List all scheduled tasks
go run main.go schedule:list
```

### Via bin/server.go

The nimbus-starter integrates scheduling into the server:

```go
// bin/server.go
func RunSchedule(s *schedule.Scheduler) {
    fmt.Println("Scheduler started. Press Ctrl+C to stop.")
    ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
    defer cancel()
    s.Start(ctx)
    <-ctx.Done()
    s.Stop()
    fmt.Println("Scheduler stopped.")
}

func RunScheduleList(s *schedule.Scheduler) {
    count := s.Count()
    fmt.Printf("Registered tasks: %d\n", count)
}
```

---

## Real-Life Examples

### Database Cleanup

```go
func RegisterSchedule(s *schedule.Scheduler) {
    // Clean expired sessions every hour
    s.Hourly("clean-sessions", func(ctx context.Context) error {
        result := db.Where("expires_at < ?", time.Now()).Delete(&Session{})
        logger.Info("cleaned expired sessions", "deleted", result.RowsAffected)
        return result.Error
    })

    // Remove soft-deleted records older than 30 days — runs daily at 2:00 AM
    s.Daily("02:00", "prune-soft-deletes", func(ctx context.Context) error {
        cutoff := time.Now().AddDate(0, 0, -30)
        db.Unscoped().Where("deleted_at < ?", cutoff).Delete(&Order{})
        db.Unscoped().Where("deleted_at < ?", cutoff).Delete(&User{})
        return nil
    })
}
```

### Metrics & Reporting

```go
func RegisterSchedule(s *schedule.Scheduler) {
    // Aggregate hourly metrics
    s.Hourly("snapshot-metrics", func(ctx context.Context) error {
        stats := metrics.ReadRuntimeStats()
        db.Create(&MetricSnapshot{
            Goroutines: stats.Goroutines,
            HeapAlloc:  stats.HeapAlloc,
            Timestamp:  time.Now(),
        })
        return nil
    })

    // Daily revenue report at 6:00 AM
    s.Daily("06:00", "daily-revenue-report", func(ctx context.Context) error {
        var total float64
        db.Model(&Order{}).
            Where("created_at >= ?", time.Now().AddDate(0, 0, -1)).
            Select("COALESCE(SUM(total), 0)").Scan(&total)

        queue.Dispatch(&jobs.SendDailyReport{
            Date:       time.Now().AddDate(0, 0, -1),
            Revenue:    total,
            AdminEmail: "admin@company.com",
        }).Dispatch(ctx)
        return nil
    })
}
```

### Cache Warming

```go
func RegisterSchedule(s *schedule.Scheduler) {
    // Pre-warm popular product cache every 5 minutes
    s.EveryFiveMinutes("warm-product-cache", func(ctx context.Context) error {
        var products []Product
        db.Order("view_count DESC").Limit(50).Find(&products)
        for _, p := range products {
            cache.Set(fmt.Sprintf("product:%d", p.ID), p, 10*time.Minute)
        }
        return nil
    })
}
```

### External API Sync

```go
func RegisterSchedule(s *schedule.Scheduler) {
    // Sync exchange rates daily at midnight
    s.Daily("00:00", "sync-exchange-rates", func(ctx context.Context) error {
        rates, err := fetchExchangeRates()
        if err != nil {
            logger.Error("exchange rate sync failed", "error", err)
            return err
        }
        for currency, rate := range rates {
            cache.Set("rate:"+currency, rate, 25*time.Hour)
        }
        return nil
    })

    // Sync inventory from supplier every hour
    s.Hourly("sync-inventory", func(ctx context.Context) error {
        return queue.Dispatch(&jobs.SyncInventory{SupplierID: 1}).Dispatch(ctx)
    })
}
```

---

## Production Deployment

### Systemd Service

```ini
[Unit]
Description=Nimbus Scheduler
After=network.target

[Service]
Type=simple
User=www-data
WorkingDirectory=/opt/myapp
ExecStart=/opt/myapp/myapp schedule:run
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
```

### Docker

```dockerfile
# Run scheduler in a separate container
CMD ["./myapp", "schedule:run"]
```

### Supervisor

```ini
[program:scheduler]
command=/opt/myapp/myapp schedule:run
directory=/opt/myapp
autostart=true
autorestart=true
stderr_logfile=/var/log/myapp/scheduler.err.log
stdout_logfile=/var/log/myapp/scheduler.out.log
```

---

## Migration from `scheduler/` to `schedule/`

The old `scheduler/` package is deprecated. Here's how to migrate:

| Old (`scheduler/`) | New (`schedule/`) |
|---------------------|-------------------|
| `import "nimbus/scheduler"` | `import "nimbus/schedule"` |
| `*scheduler.Scheduler` | `*schedule.Scheduler` |
| `s.EveryMinute(fn)` | `s.EveryMinute("name", fn)` |
| `s.EveryHour(fn)` | `s.Hourly("name", fn)` |
| `s.Daily(fn)` | `s.Daily("00:00", "name", fn)` |
| `s.Weekly(fn)` | `s.Every(7*24*time.Hour, "name", fn)` |
| `s.Run(ctx)` (blocking) | `s.Start(ctx)` (non-blocking) + `<-ctx.Done()` |
| `s.Tasks()` → `[]Task` | `s.Count()` → `int` |

---

## Best Practices

1. **Give tasks descriptive names** — Makes `schedule:list` output useful
2. **Keep tasks idempotent** — Running twice should be safe
3. **Log task execution** — Use `logger.Info` for visibility
4. **Handle errors gracefully** — Return errors for monitoring, don't panic (panics are recovered automatically)
5. **Use queues for heavy work** — Schedule should dispatch jobs, not run them inline
6. **Run in a separate process** — Don't block your web server
7. **Monitor task health** — Check if scheduler is alive (health checks)
8. **Use `Daily("HH:MM")` for time-sensitive tasks** — Reports, billing, backups

**Next:** [AI SDK](16-ai-sdk.md) →
