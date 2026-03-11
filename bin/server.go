/*
|--------------------------------------------------------------------------
| HTTP Server Entrypoint
|--------------------------------------------------------------------------
|
| The "server.go" file boots the Nimbus application: it loads
| environment variables, connects to the database, registers
| middleware and routes, then starts the HTTP server.
|
| This is imported by main.go. You should not need to modify
| main.go — edit this file, start/kernel.go, or start/routes.go
| instead.
|
*/

package bin

import (
	"context"
	"fmt"
	"net/smtp"
	"os"
	"time"

	"github.com/CodeSyncr/nimbus"
	"github.com/CodeSyncr/nimbus/cache"
	"github.com/CodeSyncr/nimbus/database"
	"github.com/CodeSyncr/nimbus/mail"
	"github.com/CodeSyncr/nimbus/packages/shield"
	"github.com/CodeSyncr/nimbus/plugins/ai"
	"github.com/CodeSyncr/nimbus/plugins/horizon"
	nimbusmcp "github.com/CodeSyncr/nimbus/plugins/mcp"
	"github.com/CodeSyncr/nimbus/plugins/telescope"
	"github.com/CodeSyncr/nimbus/plugins/unpoly"
	"github.com/CodeSyncr/nimbus/queue"
	"github.com/CodeSyncr/nimbus/scheduler"

	appmcp "nimbus-starter/app/mcp"
	"nimbus-starter/app/plugins/analytics"
	"nimbus-starter/config"
	"nimbus-starter/database/migrations"
	"nimbus-starter/database/seeders"
	"nimbus-starter/start"
)

// Boot creates, configures and returns the Nimbus application.
func Boot() *nimbus.App {
	loadConfig()

	app := newApp()

	bootMail()
	bootCache()
	bootDatabase(app)
	bootQueue()

	registerPlugins(app)
	registerMiddleware(app)
	registerRoutes(app)

	return app
}

// configureMail wires config.Mail into the global mail driver.
func bootMail() {
	if config.Mail.Driver != "smtp" {
		return
	}
	host := config.Mail.SMTP.Host
	port := config.Mail.SMTP.Port
	if host == "" || port == 0 {
		return
	}
	addr := fmt.Sprintf("%s:%d", host, port)
	var auth smtp.Auth
	if config.Mail.SMTP.Username != "" {
		auth = smtp.PlainAuth("", config.Mail.SMTP.Username, config.Mail.SMTP.Password, host)
	}
	mail.Default = mail.NewSMTPDriver(addr, auth, config.Mail.SMTP.From)
}

func loadConfig() {
	config.Load()
}

func newApp() *nimbus.App {
	return nimbus.New()
}

func bootCache() {
	cache.Boot(nil)
}

func bootDatabase(app *nimbus.App) {
	db, err := database.ConnectWithConfig(database.ConnectConfig{
		Driver: config.Database.Driver,
		DSN:    config.Database.DSN,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Database connection failed: %v\n", err)
		os.Exit(1)
	}

	// Make DB globally available via the nimbus package and the IoC container.
	nimbus.SetDB(db)
	app.Container.Singleton("db", func() *nimbus.DB {
		return nimbus.GetDB()
	})
}

func bootQueue() {
	queue.Boot(&queue.BootConfig{
		RegisterJobs: start.RegisterQueueJobs,
	})
}

// registerPlugins attaches all core plugins to the app.
func registerPlugins(app *nimbus.App) {
	// Plugins are registered before middleware and routes so
	// their services are available in the container during boot.
	// Plugin routes and middleware are applied automatically.
	app.Use(horizon.New())
	mcpPlugin := nimbusmcp.New()
	mcpPlugin.Web("/mcp/weather", appmcp.WeatherServer)

	app.Use(
		shield.NewPlugin(shield.DefaultConfig()),
		unpoly.New(),
		ai.New(),
		telescope.New(),
		mcpPlugin,
		analytics.New(),
	)
}

func registerMiddleware(app *nimbus.App) {
	start.RegisterMiddleware(app)
}

func registerRoutes(app *nimbus.App) {
	start.RegisterRoutes(app)
}

// RunMigrations runs database migrations. Called when main is invoked with "migrate" arg.
func RunMigrations() {
	config.Load()
	db, err := database.ConnectWithConfig(database.ConnectConfig{
		Driver: config.Database.Driver,
		DSN:    config.Database.DSN,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Database connection failed: %v\n", err)
		os.Exit(1)
	}
	migrator := database.NewMigrator(db, migrations.All())
	if err := migrator.Up(); err != nil {
		fmt.Fprintf(os.Stderr, "Migration failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Migrations completed.")
}

// RunSeeders runs database seeders. Called when main is invoked with "seed" arg.
func RunSeeders() {
	config.Load()
	db, err := database.ConnectWithConfig(database.ConnectConfig{
		Driver: config.Database.Driver,
		DSN:    config.Database.DSN,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Database connection failed: %v\n", err)
		os.Exit(1)
	}
	runner := database.NewSeedRunner(db, seeders.All())
	if err := runner.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Seeding failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Seeders completed.")
}

// RunSchedule runs scheduled tasks. Intended to be called from "nimbus schedule:run".
// A simple, timer-based scheduler is used; register tasks in start.RegisterSchedule.
func RunSchedule() {
	config.Load()
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	s := scheduler.New()
	start.RegisterSchedule(s)
	s.Run(ctx)
}

// RunScheduleList lists scheduled tasks. Called when main is invoked with "schedule:list".
func RunScheduleList() {
	config.Load()
	s := scheduler.New()
	start.RegisterSchedule(s)
	tasks := s.Tasks()
	if len(tasks) == 0 {
		fmt.Println("No scheduled tasks registered.")
		return
	}
	fmt.Println("Scheduled tasks:")
	for i, t := range tasks {
		fmt.Printf("  %d) Every %s\n", i+1, t.Interval)
	}
}
