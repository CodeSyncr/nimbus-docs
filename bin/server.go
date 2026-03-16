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
	"github.com/CodeSyncr/nimbus/database/nosql"
	"github.com/CodeSyncr/nimbus/http"
	"github.com/CodeSyncr/nimbus/mail"
	"github.com/CodeSyncr/nimbus/packages/shield"
	"github.com/CodeSyncr/nimbus/plugins/ai"
	"github.com/CodeSyncr/nimbus/plugins/horizon"
	nimbusmcp "github.com/CodeSyncr/nimbus/plugins/mcp"
	"github.com/CodeSyncr/nimbus/plugins/telescope"
	"github.com/CodeSyncr/nimbus/plugins/transmit"
	"github.com/CodeSyncr/nimbus/plugins/unpoly"
	"github.com/CodeSyncr/nimbus/queue"
	"github.com/CodeSyncr/nimbus/schedule"

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
	bootNoSQL(app)
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

func bootNoSQL(app *nimbus.App) {
	if config.Database.MongoURI == "" {
		return // NoSQL not configured — skip silently
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	mongoDriver, err := nosql.ConnectMongo(ctx, nosql.MongoConfig{
		URI:      config.Database.MongoURI,
		Database: config.Database.MongoDatabase,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "MongoDB connection failed: %v\n", err)
		os.Exit(1)
	}

	// Register in the nosql connection manager
	nosql.Register("mongo", mongoDriver)

	// Make NoSQL globally available via the nimbus package and the IoC container.
	nimbus.SetNoSQL(mongoDriver)
	app.Container.Singleton("nosql", func() *nimbus.NoSQL {
		return nimbus.GetNoSQL()
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
	app.Use(horizon.NewWithOptions(horizon.Options{
		Config: &horizon.Config{
			Environments: toHorizonEnvs(config.Horizon.Environments),
			Defaults: horizon.SupervisorDefaults{
				Connection: config.Horizon.Defaults.Connection,
				Timeout:    config.Horizon.Defaults.Timeout,
				Tries:      config.Horizon.Defaults.Tries,
				Backoff:    config.Horizon.Defaults.Backoff,
			},
			Waits:    config.Horizon.Waits,
			Silenced: config.Horizon.Silenced,
		},
		RedisURL: config.Horizon.RedisURL,
	}))
	mcpPlugin := nimbusmcp.New()
	mcpPlugin.Web("/mcp/weather", appmcp.WeatherServer)

	// Build Transmit config from config package.
	transmitCfg := &transmit.Config{
		Path:         config.Transmit.Path,
		PingInterval: config.Transmit.PingInterval,
	}
	if config.Transmit.Transport == "redis" {
		if rt, err := transmit.NewRedisTransport(transmit.RedisTransportConfig{
			URL:     config.Transmit.Redis.URL,
			Channel: config.Transmit.Redis.Channel,
		}); err == nil {
			transmitCfg.Transport = rt
		}
	}

	// Build Shield config from config package.
	shieldCfg := toShieldConfig(config.Shield)

	app.Use(
		shield.NewPlugin(shieldCfg),
		unpoly.New(),
		ai.New(),
		telescope.New(),
		transmit.New(transmitCfg),
		mcpPlugin,
		analytics.New(),
	)
}

// toShieldConfig converts a config.ShieldConfig into a shield.Config.
func toShieldConfig(c config.ShieldConfig) shield.Config {
	return shield.Config{
		ContentTypeNosniff:           c.ContentTypeNosniff,
		XSSProtection:                c.XSSProtection,
		FrameGuard:                   c.FrameGuard,
		ReferrerPolicy:               c.ReferrerPolicy,
		DNSPrefetchControl:           c.DNSPrefetchControl,
		DownloadOptions:              c.DownloadOptions,
		PermittedCrossDomainPolicies: c.PermittedCrossDomainPolicies,
		CrossOriginOpenerPolicy:      c.CrossOriginOpenerPolicy,
		CrossOriginResourcePolicy:    c.CrossOriginResourcePolicy,
		CrossOriginEmbedderPolicy:    c.CrossOriginEmbedderPolicy,
		HSTS: shield.HSTSConfig{
			Enabled:           c.HSTS.Enabled,
			MaxAge:            c.HSTS.MaxAge,
			IncludeSubdomains: c.HSTS.IncludeSubdomains,
			Preload:           c.HSTS.Preload,
		},
		CSP: shield.CSPConfig{
			Enabled:    c.CSP.Enabled,
			ReportOnly: c.CSP.ReportOnly,
		},
		CSRF: shield.CSRFConfig{
			Enabled:     c.CSRF.Enabled,
			CookieName:  c.CSRF.CookieName,
			HeaderName:  c.CSRF.HeaderName,
			FieldName:   c.CSRF.FieldName,
			MaxAge:      c.CSRF.MaxAge,
			Secure:      c.CSRF.Secure,
			SameSite:    c.CSRF.SameSite,
			Path:        c.CSRF.Path,
			Domain:      c.CSRF.Domain,
			HttpOnly:    c.CSRF.HttpOnly,
			ExceptPaths: c.CSRF.ExceptPaths,
			RotateToken: c.CSRF.RotateToken,
		},
	}
}

// toHorizonEnvs converts config types to horizon types.
func toHorizonEnvs(envs map[string]config.HorizonEnvironmentConfig) map[string]horizon.EnvironmentConfig {
	out := make(map[string]horizon.EnvironmentConfig, len(envs))
	for name, ec := range envs {
		sups := make(map[string]horizon.SupervisorConfig, len(ec.Supervisors))
		for sn, sc := range ec.Supervisors {
			sups[sn] = horizon.SupervisorConfig{
				Connection:      sc.Connection,
				Queue:           sc.Queue,
				Balance:         sc.Balance,
				Processes:       sc.Processes,
				MinProcesses:    sc.MinProcesses,
				MaxProcesses:    sc.MaxProcesses,
				BalanceMaxShift: sc.BalanceMaxShift,
				BalanceCooldown: sc.BalanceCooldown,
				Tries:           sc.Tries,
				Timeout:         sc.Timeout,
				Backoff:         sc.Backoff,
				Force:           sc.Force,
			}
		}
		out[name] = horizon.EnvironmentConfig{Supervisors: sups}
	}
	return out
}

func registerMiddleware(app *nimbus.App) {
	start.RegisterMiddleware(app)
}

func registerRoutes(app *nimbus.App) {
	if config.Static.Enabled {
		fs := http.FileServer(http.Dir(config.Static.Root))
		// E.g. Mount "/public" -> route to "public" dir
		app.Router.Mount(config.Static.Prefix, http.StripPrefix(config.Static.Prefix, fs))
	}
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
// Uses the schedule package with named tasks, panic recovery, and daily-at scheduling.
func RunSchedule() {
	config.Load()
	ctx := context.Background()

	s := schedule.New()
	start.RegisterSchedule(s)
	s.Start(ctx)

	// Block until interrupted.
	<-ctx.Done()
	s.Stop()
}

// RunScheduleList lists scheduled tasks. Called when main is invoked with "schedule:list".
func RunScheduleList() {
	config.Load()
	s := schedule.New()
	start.RegisterSchedule(s)
	count := s.Count()
	if count == 0 {
		fmt.Println("No scheduled tasks registered.")
		return
	}
	fmt.Printf("Scheduled tasks: %d registered\n", count)
}
