/*
|--------------------------------------------------------------------------
| Routes
|--------------------------------------------------------------------------
|
| This file defines all HTTP routes for the application. Register
| your controllers, resource routes, and page handlers here.
|
| See: /docs/routing
|
*/

package start

import (
	"encoding/json"
	"strings"

	"github.com/CodeSyncr/nimbus"
	"github.com/CodeSyncr/nimbus/database"
	"github.com/CodeSyncr/nimbus/health"
	"github.com/CodeSyncr/nimbus/http"
	"github.com/CodeSyncr/nimbus/plugins/ai"
	"github.com/CodeSyncr/nimbus/plugins/unpoly"
	"github.com/CodeSyncr/nimbus/queue"

	"nimbus-starter/app/controllers"
	"nimbus-starter/app/docs"
	"nimbus-starter/app/jobs"
)

// RegisterRoutes wires every route onto the application router.
func RegisterRoutes(app *nimbus.App) {

	// ── Pages ──────────────────────────────────────────────
	app.Router.Get("/", homeHandler)
	app.Router.Get("/health", healthHandler)

	// ── Resources ──────────────────────────────────────────
	app.Router.Resource("hello", &controllers.HelloWorld{})

	// ── Demos (sample apps) ────────────────────────────────
	demos := app.Router.Group("/demos")
	demos.Get("/", demosIndexHandler)
	todoCtrl := &controllers.Todo{
		DB: app.Container.MustMake("db").(*nimbus.DB),
	}
	demos.Resource("todo", todoCtrl)
	demos.Get("/todo/confirm", todoCtrl.Confirm)
	demos.Post("/todo/:id/update", func(c *http.Context) error { return todoCtrl.Update(c) })
	demos.Post("/todo/:id/delete", func(c *http.Context) error { return todoCtrl.Destroy(c) })
	demos.Post("/todo/:id/toggle", func(c *http.Context) error { return todoCtrl.Toggle(c) })
	demos.Get("/counter", func(c *http.Context) error { return (&controllers.Counter{}).Index(c) })
	demos.Post("/counter/increment", func(c *http.Context) error { return (&controllers.Counter{}).Increment(c) })
	demos.Post("/counter/decrement", func(c *http.Context) error { return (&controllers.Counter{}).Decrement(c) })
	demos.Post("/counter/set", func(c *http.Context) error { return (&controllers.Counter{}).Set(c) })
	aiCtrl := &controllers.AI{}
	demos.Get("/ai", func(c *http.Context) error { return aiCtrl.Index(c) })
	demos.Post("/ai/generate", func(c *http.Context) error { return aiCtrl.Generate(c) })
	demos.Get("/mcp", mcpDemoHandler)
	demos.Get("/livewire/fixtures", livewireFixturesHandler)

	// Queue demo: dispatch a welcome email job to the "default" queue.
	demos.Post("/queue/demo", queueDemoHandler)

	// ── Livewire plugin docs (separate from /docs tree) ───
	app.Router.Get("/livewire/docs", livewireDocsIndexHandler)
	app.Router.Get("/livewire/docs/:page", livewireDocsPageHandler)

	// ── Documentation ──────────────────────────────────────
	app.Router.Get("/docs", docsIndexHandler)
	app.Router.Get("/docs/*", docsPageHandler)
	app.Router.Get("/api/docs/index", docsIndexAPIHandler)
	app.Router.Post("/api/docs/chat", docsChatAPIHandler)
}

// ── Handlers ─────────────────────────────────────────────────

func homeHandler(c *http.Context) error {
	if unpoly.IsUnpoly(c) {
		unpoly.SetTitle(c, "Welcome · Nimbus")
	}
	return c.View("home", map[string]any{
		"title":   "Welcome",
		"appName": "Nimbus",
		"tagline": "Laravel style framework for Go",
		"version": "0.1.4",
	})
}

func healthHandler(c *http.Context) error {
	checker := health.New()
	if database.DB != nil {
		checker.DB(database.DB)
	}
	result := checker.Run(c.Request.Context())
	code := http.StatusOK
	if result.Status != "ok" {
		code = http.StatusServiceUnavailable
	}
	if unpoly.IsUnpoly(c) {
		unpoly.RenderNothing(c)
		unpoly.EmitEvent(c, "health:checked", map[string]any{
			"status": result.Status,
			"checks": result.Checks,
		})
	}
	return c.JSON(code, result)
}

func demosIndexHandler(c *http.Context) error {
	return c.View("apps/index", map[string]any{"title": "Demos"})
}

func mcpDemoHandler(c *http.Context) error {
	return c.View("apps/mcp/index", map[string]any{"title": "MCP Demo"})
}

func livewireFixturesHandler(c *http.Context) error {
	return c.View("apps/livewire/fixtures", map[string]any{
		"title":       "Livewire fixtures",
		"activeDemos": true,
	})
}

func queueDemoHandler(c *http.Context) error {
	job := &jobs.SendWelcomeEmail{
		UserID: 1,
		Email:  "queue-demo@example.com",
	}
	if err := queue.Dispatch(job).Dispatch(c.Request.Context()); err != nil {
		// Keep the demo usable even if a durable queue backend (e.g. Redis) isn't running.
		// We still surface the queue error so it's obvious what to fix.
		if inlineErr := job.Handle(c.Request.Context()); inlineErr == nil {
			return c.JSON(http.StatusAccepted, map[string]any{
				"status":      "processed_inline",
				"queued":      false,
				"queue_error": err.Error(),
			})
		}
		return c.JSON(http.StatusInternalServerError, map[string]any{
			"error":       "failed to queue job",
			"queue_error": err.Error(),
		})
	}
	return c.JSON(http.StatusAccepted, map[string]string{"status": "queued"})
}

// ── Documentation ────────────────────────────────────────────

var docsTitles = map[string]string{
	// Start
	"getting-started":      "Getting Started",
	"introduction":         "Introduction",
	"installation":         "Installation",
	"folder-structure":     "Folder Structure",
	"configuration":        "Configuration",
	"deployment":           "Deployment",
	"production-readiness": "Production Readiness",
	"versioning-policy":    "Versioning & Release Policy",
	"release-checklist":    "Release Checklist",
	"faqs":                 "FAQs",
	// Basics
	"routing":            "Routing",
	"controllers":        "Controllers",
	"http-context":       "HTTP Context",
	"middleware":         "Middleware",
	"request":            "Request",
	"response":           "Response",
	"body-parser":        "Body Parser",
	"validation":         "Validation",
	"file-uploads":       "File Uploads",
	"locale":             "Localization",
	"api-resources":      "API Resources",
	"session":            "Session",
	"exception-handling": "Exception Handling",
	"static-files":       "Static Files",
	// Nimbus Template
	"nimbus-template": "Nimbus Template",
	// Inertia
	"inertia":       "Inertia.js",
	"inertia-setup": "Inertia Setup",
	"inertia-hmr":   "Inertia HMR & Vite",
	// Data Layer
	"database":                        "Database & ORM",
	"database-query-select":           "Select Query Builder",
	"database-query-insert":           "Insert Query Builder",
	"database-query-raw":              "Raw Query Builder",
	"database-migrations-intro":       "Migrations Introduction",
	"database-migrations-schema":      "Schema Builder",
	"database-migrations-table":       "Table Builder",
	"database-models-intro":           "Models Introduction",
	"database-models-schema-classes":  "Schema Classes",
	"database-models-crud":            "CRUD Operations",
	"database-models-hooks":           "Hooks",
	"database-models-query-builder":   "Query Builder",
	"database-models-naming-strategy": "Naming Strategy",
	"database-models-query-scopes":    "Query Scopes",
	"database-models-serializing":     "Serializing Models",
	"database-models-relationships":   "Relationships",
	"database-models-factories":       "Model Factories",
	"migrations":                      "Migrations",
	"seeders":                         "Seeders",
	"nosql":                           "NoSQL / MongoDB",
	"nosql-query-builder":             "NoSQL Query Builder",
	"multi-db":                        "Multiple DB Connections",
	// Auth
	"auth":              "Auth & Guards",
	"auth-introduction": "Auth Introduction",
	"hash":              "Hashing",
	"session-guard":     "Session Guard",
	"access-tokens":     "Access Tokens",
	"stateless-guard":   "Stateless Guard",
	"authorization":     "Authorization",
	// Security
	"shield":        "Shield",
	"cors":          "CORS",
	"csrf":          "CSRF",
	"rate-limiting": "Rate Limiting",
	"health":        "Health Checks",
	// Core Concepts
	"application-lifecycle": "Application Lifecycle",
	"dependency-injection":  "Dependency Injection",
	"service-providers":     "Service Providers",
	"plugins":               "Plugins",
	"container-services":    "Container Services",
	// Helpers
	"helpers-string":     "Fluent String",
	"helpers-collection": "Collections",
	"helpers-time":       "Date & Time",
	"helpers-pipeline":   "Async Pipelines",
	// Digging Deeper
	"cache":              "Cache",
	"cache-remember":     "Remember & Type-Safe API",
	"cache-backends":     "Cache Backends",
	"cache-invalidation": "Cache Invalidation & ORM Hooks",
	"storage":            "Storage",
	"drive":              "Drive",
	"transmit":           "Transmit",
	"reverb":             "Reverb",
	"events":             "Events",
	"logger":             "Logger",
	"mail":               "Mail",
	"notification":       "Notifications",
	"queue":              "Queue",
	"scheduler":          "Scheduler",
	"websockets":         "WebSockets",
	// Command Line
	"cli":               "Introduction",
	"creating-commands": "Creating Commands",
	"command-arguments": "Command Arguments",
	"command-flags":     "Command Flags",
	"prompts":           "Prompts",
	"terminal-ui":       "Terminal UI",
	"repl":              "Repl",
	"hot-reload":        "Hot Reload",
	// AI
	"ai":       "AI SDK",
	"ai-video": "AI Video Pipeline",
	"mcp":      "MCP",
	// Testing
	"testing-introduction": "Testing Introduction",
	"http-tests":           "HTTP Tests",
	// Advanced Features
	"workflow":       "Workflow Engine",
	"feature-flags":  "Feature Flags",
	"multi-tenancy":  "Multi-Tenancy",
	"presence":       "Realtime Presence",
	"openapi":        "OpenAPI Generation",
	"studio":         "Studio Admin Panel",
	"edge-functions": "Edge Functions",
	"metrics":        "Runtime Metrics",
	// Plugins
	"telescope": "Telescope",
	"horizon":   "Horizon",
	"pulse":     "Pulse",
	"socialite": "Socialite",
	"unpoly":    "Unpoly",
}

// livewireDirectiveDetailSlugs are per-directive doc pages (registered in init).
var livewireDirectiveDetailSlugs = []string{
	"wire-bind", "wire-click", "wire-submit", "wire-model", "wire-navigate", "wire-current",
	"wire-cloak", "wire-dirty", "wire-confirm", "wire-loading", "wire-transition",
	"wire-init", "wire-intersect", "wire-poll", "wire-offline", "wire-ignore", "wire-ref",
	"wire-show", "wire-text", "wire-replace", "wire-sort", "wire-stream",
}

var livewireDirectivePageTOC = []map[string]any{
	{"id": "intro", "label": "Introduction"},
	{"id": "basic-usage", "label": "Basic usage"},
	{"id": "common-use-cases", "label": "Common use cases"},
	{"id": "reference", "label": "Reference"},
}

func livewireDirectiveAttrTitle(slug string) string {
	return "wire:" + strings.TrimPrefix(slug, "wire-")
}

func init() {
	for _, slug := range livewireDirectiveDetailSlugs {
		attr := livewireDirectiveAttrTitle(slug)
		livewireDocRegistry[slug] = struct {
			View  string
			Title string
			TOC   []map[string]any
		}{
			View:  "livewire/docs/" + slug,
			Title: attr + " · Nimbus Livewire",
			TOC:   livewireDirectivePageTOC,
		}
	}
}

// livewireDocsOrder is sidebar / prev-next order.
var livewireDocsOrder = append(append([]string{
	"quickstart", "installation", "upgrade-guide", "parity", "components", "nesting", "pages", "properties",
	"actions", "forms", "validation", "pagination", "url-query-parameters", "computed-properties", "redirecting", "file-downloads", "teleport", "events", "lifecycle-hooks", "navigate", "alpine", "styles", "islands", "lazy-loading", "loading-states", "directives",
}, livewireDirectiveDetailSlugs...), "testing")

// livewireDocRegistry maps URL slug → view name, browser title, and optional TOC (in-page anchors).
var livewireDocRegistry = map[string]struct {
	View  string
	Title string
	TOC   []map[string]any
}{
	"quickstart": {
		View:  "livewire/docs/quickstart",
		Title: "Quickstart · Nimbus Livewire",
		TOC: []map[string]any{
			{"id": "install", "label": "Install"},
			{"id": "register", "label": "Register components"},
			{"id": "mount-route", "label": "Full-page route"},
			{"id": "layout-script", "label": "Layout script"},
		},
	},
	"installation": {
		View:  "livewire/docs/installation",
		Title: "Installation · Nimbus Livewire",
		TOC: []map[string]any{
			{"id": "overview", "label": "Overview"},
			{"id": "endpoints", "label": "Script & update endpoint"},
			{"id": "payload", "label": "Update payload"},
		},
	},
	"upgrade-guide": {
		View:  "livewire/docs/upgrade-guide",
		Title: "Upgrade guide · Nimbus Livewire",
		TOC:   []map[string]any{{"id": "notes", "label": "Notes"}},
	},
	"parity": {
		View:  "livewire/docs/parity",
		Title: "Parity matrix · Nimbus Livewire",
		TOC: []map[string]any{
			{"id": "overview", "label": "Overview"},
			{"id": "matrix", "label": "Parity matrix"},
			{"id": "fixtures", "label": "Fixtures page"},
			{"id": "notes", "label": "Notes / known gaps"},
		},
	},
	"components": {
		View:  "livewire/docs/components",
		Title: "Components · Nimbus Livewire",
		TOC: []map[string]any{
			{"id": "contract", "label": "Component contract"},
			{"id": "register", "label": "Registering"},
			{"id": "namespaced", "label": "Namespaced names"},
			{"id": "embed-template", "label": "Embed in views"},
			{"id": "embed-controller", "label": "Embed from Go"},
			{"id": "props-mount", "label": "PropsMount vs SetState"},
			{"id": "pages-vs-embed", "label": "Pages vs embedded"},
			{"id": "state", "label": "State & serialization"},
			{"id": "scaffold", "label": "Scaffolding"},
			{"id": "troubleshooting", "label": "Troubleshooting"},
			{"id": "see-also", "label": "See also"},
		},
	},
	"nesting": {
		View:  "livewire/docs/nesting",
		Title: "Nesting components · Nimbus Livewire",
		TOC: []map[string]any{
			{"id": "overview", "label": "Overview"},
			{"id": "render-nested", "label": "RenderNested & keys"},
			{"id": "wire-id-render", "label": "RenderForWireID"},
			{"id": "props", "label": "Passing props"},
			{"id": "loops", "label": "Lists & slot keys"},
			{"id": "parent-child", "label": "Parent ↔ child"},
			{"id": "limits", "label": "Laravel parity limits"},
		},
	},
	"pages": {
		View:  "livewire/docs/pages-doc",
		Title: "Pages · Nimbus Livewire",
		TOC: []map[string]any{
			{"id": "routing", "label": "Routing to components"},
			{"id": "layouts", "label": "Layouts"},
			{"id": "layout-scaffold", "label": "Creating the layout"},
			{"id": "default-layout", "label": "Default layout options"},
			{"id": "layout-picker", "label": "Component layouts"},
			{"id": "title", "label": "Page title"},
			{"id": "named-slots", "label": "Named slots"},
			{"id": "route-params", "label": "Route parameters"},
			{"id": "model-binding", "label": "Model binding"},
			{"id": "see-also", "label": "See also"},
		},
	},
	"properties": {
		View:  "livewire/docs/properties",
		Title: "Properties · Nimbus Livewire",
		TOC: []map[string]any{
			{"id": "initializing", "label": "Initializing"},
			{"id": "fill", "label": "Bulk assignment"},
			{"id": "only", "label": "PickKeys (only)"},
			{"id": "binding", "label": "Data binding"},
			{"id": "reset", "label": "Reset"},
			{"id": "pull", "label": "Pull"},
			{"id": "types", "label": "Supported types"},
			{"id": "wireables", "label": "Custom types"},
			{"id": "javascript", "label": "JavaScript API"},
			{"id": "security", "label": "Security"},
			{"id": "helpers", "label": "Package helpers"},
			{"id": "see-also", "label": "See also"},
		},
	},
	"actions": {
		View:  "livewire/docs/actions",
		Title: "Actions · Nimbus Livewire",
		TOC: []map[string]any{
			{"id": "basics", "label": "Basics"},
			{"id": "parameters", "label": "Parameters"},
			{"id": "dependency-injection", "label": "Dependency injection"},
			{"id": "directives", "label": "Event directives"},
			{"id": "modifiers", "label": "Modifiers"},
			{"id": "magic", "label": "Magic actions"},
			{"id": "skip-render", "label": "Skip re-render"},
			{"id": "javascript-api", "label": "JavaScript API"},
			{"id": "js-actions", "label": "JS-only actions"},
			{"id": "loading", "label": "Loading & forms"},
			{"id": "confirm", "label": "Confirm"},
			{"id": "security", "label": "Security"},
			{"id": "see-also", "label": "See also"},
		},
	},
	"forms": {
		View:  "livewire/docs/forms",
		Title: "Forms · Nimbus Livewire",
		TOC: []map[string]any{
			{"id": "submit", "label": "Submitting"},
			{"id": "default-vs-live", "label": "Default vs .live"},
			{"id": "validation", "label": "Validation"},
			{"id": "form-objects", "label": "Form objects"},
			{"id": "reset-pull", "label": "Reset & pull"},
			{"id": "loading", "label": "Loading"},
			{"id": "live-blur", "label": "Live, blur, debounce"},
			{"id": "dirty", "label": "Dirty"},
			{"id": "autosave", "label": "Real-time save"},
			{"id": "components-ui", "label": "Partials & custom UI"},
			{"id": "files", "label": "File uploads"},
			{"id": "see-also", "label": "See also"},
		},
	},
	"validation": {
		View:  "livewire/docs/validation",
		Title: "Validation · Nimbus Livewire",
		TOC: []map[string]any{
			{"id": "overview", "label": "Overview"},
			{"id": "basic", "label": "Basic validation"},
			{"id": "realtime", "label": "Real-time validation"},
			{"id": "js", "label": "JavaScript errors"},
			{"id": "notes", "label": "Notes / limitations"},
		},
	},
	"pagination": {
		View:  "livewire/docs/pagination",
		Title: "Pagination · Nimbus Livewire",
		TOC: []map[string]any{
			{"id": "overview", "label": "Overview"},
			{"id": "basic", "label": "Basic usage"},
			{"id": "url", "label": "URL query string tracking"},
			{"id": "reset", "label": "Resetting the page"},
			{"id": "methods", "label": "Page navigation methods"},
			{"id": "multiple", "label": "Multiple paginators"},
			{"id": "hooks", "label": "Hooking into page updates"},
			{"id": "themes", "label": "Themes & custom views"},
			{"id": "notes", "label": "Notes / limitations"},
		},
	},
	"url-query-parameters": {
		View:  "livewire/docs/url-query-parameters",
		Title: "URL query parameters · Nimbus Livewire",
		TOC: []map[string]any{
			{"id": "overview", "label": "Overview"},
			{"id": "basic", "label": "Basic usage"},
			{"id": "init", "label": "Initializing from URL"},
			{"id": "nullable", "label": "Nullable values"},
			{"id": "alias", "label": "Aliases"},
			{"id": "except", "label": "Excluding values"},
			{"id": "keep", "label": "Display on page load (keep)"},
			{"id": "history", "label": "Storing in history"},
			{"id": "method", "label": "QueryString() method"},
			{"id": "notes", "label": "Notes / limitations"},
		},
	},
	"computed-properties": {
		View:  "livewire/docs/computed-properties",
		Title: "Computed properties · Nimbus Livewire",
		TOC: []map[string]any{
			{"id": "overview", "label": "Overview"},
			{"id": "basic", "label": "Basic usage"},
			{"id": "performance", "label": "Performance (memoization)"},
			{"id": "clearing", "label": "Clearing the memo"},
			{"id": "persist", "label": "Caching between requests (persist)"},
			{"id": "cache", "label": "Caching across components (cache)"},
			{"id": "notes", "label": "Notes / limitations"},
		},
	},
	"redirecting": {
		View:  "livewire/docs/redirecting",
		Title: "Redirecting · Nimbus Livewire",
		TOC: []map[string]any{
			{"id": "overview", "label": "Overview"},
			{"id": "basic", "label": "Basic usage"},
			{"id": "navigate", "label": "Redirect using wire:navigate"},
			{"id": "intended", "label": "Redirect intended"},
			{"id": "flash", "label": "Flash messages"},
			{"id": "notes", "label": "Notes / limitations"},
		},
	},
	"file-downloads": {
		View:  "livewire/docs/file-downloads",
		Title: "File downloads · Nimbus Livewire",
		TOC: []map[string]any{
			{"id": "overview", "label": "Overview"},
			{"id": "basic", "label": "Basic usage"},
			{"id": "streaming", "label": "Streaming downloads"},
			{"id": "testing", "label": "Testing file downloads"},
			{"id": "notes", "label": "Notes / limitations"},
		},
	},
	"teleport": {
		View:  "livewire/docs/teleport",
		Title: "Teleport · Nimbus Livewire",
		TOC: []map[string]any{
			{"id": "overview", "label": "Overview"},
			{"id": "basic", "label": "Basic usage"},
			{"id": "rules", "label": "Rules"},
			{"id": "notes", "label": "Notes / limitations"},
		},
	},
	"events": {
		View:  "livewire/docs/events",
		Title: "Events · Nimbus Livewire",
		TOC: []map[string]any{
			{"id": "overview", "label": "Overview"},
			{"id": "dispatch-server", "label": "Dispatching from Go"},
			{"id": "listen-server", "label": "Listening in Go"},
			{"id": "dynamic-names", "label": "Dynamic event names"},
			{"id": "client-dispatch", "label": "Client: find & globals"},
			{"id": "wire-click-dispatch", "label": "wire:click $dispatch"},
			{"id": "alpine", "label": "Alpine & vanilla JS"},
			{"id": "navigate-events", "label": "Navigate events"},
			{"id": "not-yet", "label": "Not in Nimbus yet"},
		},
	},
	"lifecycle-hooks": {
		View:  "livewire/docs/lifecycle-hooks",
		Title: "Lifecycle hooks · Nimbus Livewire",
		TOC: []map[string]any{
			{"id": "overview", "label": "Hook map"},
			{"id": "request-order", "label": "Request order"},
			{"id": "first-render", "label": "First render"},
			{"id": "post-update", "label": "POST update"},
			{"id": "mount-boot", "label": "Mount vs boot"},
			{"id": "hydrate-dehydrate", "label": "Hydrate & dehydrate"},
			{"id": "property-hooks", "label": "Property hooks"},
			{"id": "exception-hook", "label": "Exception handler"},
			{"id": "traits-forms", "label": "Traits & forms"},
			{"id": "directives", "label": "Client directives"},
		},
	},
	"navigate": {
		View:  "livewire/docs/navigate",
		Title: "Navigate · Nimbus Livewire",
		TOC: []map[string]any{
			{"id": "usage", "label": "Usage"},
			{"id": "behavior", "label": "Behavior"},
		},
	},
	"alpine": {
		View:  "livewire/docs/alpine",
		Title: "Alpine · Nimbus Livewire",
		TOC: []map[string]any{
			{"id": "script-order", "label": "Script order"},
			{"id": "basic", "label": "Alpine inside components"},
			{"id": "wire-magic", "label": "Magic $wire"},
			{"id": "reactivity", "label": "Reactivity note"},
			{"id": "entangle", "label": "Entangle"},
			{"id": "find", "label": "Livewire.find"},
			{"id": "bundle", "label": "Bundling"},
		},
	},
	"styles": {
		View:  "livewire/docs/styles",
		Title: "Styles · Nimbus Livewire",
		TOC: []map[string]any{
			{"id": "overview", "label": "Overview"},
			{"id": "scoped", "label": "Scoped styles"},
			{"id": "root", "label": "Targeting the root"},
			{"id": "global", "label": "Global styles"},
			{"id": "dedupe", "label": "Deduplication"},
			{"id": "browser", "label": "Browser support"},
		},
	},
	"islands": {
		View:  "livewire/docs/islands",
		Title: "Islands · Nimbus Livewire",
		TOC: []map[string]any{
			{"id": "overview", "label": "Overview"},
			{"id": "basic", "label": "Basic usage"},
			{"id": "lazy", "label": "Lazy / defer"},
			{"id": "placeholder", "label": "Placeholders"},
			{"id": "trigger", "label": "Triggering islands"},
			{"id": "modes", "label": "Append / prepend"},
			{"id": "js", "label": "JavaScript / Alpine"},
			{"id": "limits", "label": "Limitations"},
		},
	},
	"lazy-loading": {
		View:  "livewire/docs/lazy-loading",
		Title: "Lazy loading · Nimbus Livewire",
		TOC: []map[string]any{
			{"id": "overview", "label": "Overview"},
			{"id": "lazy-vs-defer", "label": "Lazy vs defer"},
			{"id": "basic", "label": "Basic example"},
			{"id": "placeholder", "label": "Placeholder HTML"},
			{"id": "props", "label": "Passing props"},
			{"id": "notes", "label": "Notes"},
		},
	},
	"loading-states": {
		View:  "livewire/docs/loading-states",
		Title: "Loading states · Nimbus Livewire",
		TOC: []map[string]any{
			{"id": "overview", "label": "Overview"},
			{"id": "basic", "label": "Basic usage"},
			{"id": "how", "label": "How it works"},
			{"id": "tailwind", "label": "Tailwind patterns"},
			{"id": "css", "label": "Plain CSS"},
			{"id": "wire-loading", "label": "wire:loading vs data-loading"},
		},
	},
	"directives": {
		View:  "livewire/docs/directives",
		Title: "HTML directives · Nimbus Livewire",
		TOC: []map[string]any{
			{"id": "intro", "label": "Introduction"},
			{"id": "overview", "label": "All directives"},
		},
	},
	"testing": {
		View:  "livewire/docs/testing",
		Title: "Testing · Nimbus Livewire",
		TOC: []map[string]any{
			{"id": "overview", "label": "Overview"},
			{"id": "quickstart", "label": "Quickstart"},
			{"id": "api", "label": "API reference"},
			{"id": "laravel-map", "label": "Laravel → Nimbus map"},
			{"id": "http", "label": "HTTP / JSON tests"},
			{"id": "e2e", "label": "Browser / E2E"},
			{"id": "run", "label": "Plugin tests"},
		},
	},
}

func livewireDocsShortTitle(full string) string {
	if i := strings.Index(full, " · "); i > 0 {
		return full[:i]
	}
	return full
}

func livewireDocsIndexHandler(c *http.Context) error {
	c.Redirect(302, "/livewire/docs/quickstart")
	return nil
}

func livewireDocsPageHandler(c *http.Context) error {
	page := c.Param("page")
	meta, ok := livewireDocRegistry[page]
	if !ok {
		return c.NotFound()
	}
	data := map[string]any{
		"title":       meta.Title,
		"docNavTitle": livewireDocsShortTitle(meta.Title),
		"docPage":     page,
		"toc":         meta.TOC,
	}
	idx := -1
	for ii, s := range livewireDocsOrder {
		if s == page {
			idx = ii
			break
		}
	}
	if idx > 0 {
		prev := livewireDocsOrder[idx-1]
		pm := livewireDocRegistry[prev]
		data["docPrev"] = "/livewire/docs/" + prev
		data["docPrevTitle"] = livewireDocsShortTitle(pm.Title)
	}
	if idx >= 0 && idx < len(livewireDocsOrder)-1 {
		next := livewireDocsOrder[idx+1]
		nm := livewireDocRegistry[next]
		data["docNext"] = "/livewire/docs/" + next
		data["docNextTitle"] = livewireDocsShortTitle(nm.Title)
	}
	return c.View(meta.View, data)
}

func docsIndexHandler(c *http.Context) error {
	if unpoly.IsUnpoly(c) {
		unpoly.SetTitle(c, "Documentation · Nimbus")
	}
	return c.View("docs/index", map[string]any{"title": "Documentation"})
}

// docsOrder is the reading order for prev/next navigation.
var docsOrder = []string{
	"getting-started", "introduction", "installation", "folder-structure", "configuration", "deployment", "production-readiness", "versioning-policy", "release-checklist", "faqs",
	"routing", "controllers", "http-context", "middleware", "request", "response", "body-parser",
	"validation", "file-uploads", "locale", "api-resources", "session", "exception-handling", "static-files",
	"nimbus-template",
	"inertia", "inertia-setup", "inertia-hmr",
	"database", "database-query-select", "database-query-insert", "database-query-raw",
	"database-migrations-intro", "database-migrations-schema", "database-migrations-table",
	"database-models-intro", "database-models-schema-classes", "database-models-crud",
	"database-models-hooks", "database-models-query-builder", "database-models-naming-strategy",
	"database-models-query-scopes", "database-models-serializing", "database-models-relationships",
	"database-models-factories",
	"migrations", "seeders",
	"nosql", "nosql-query-builder", "multi-db",
	"auth", "auth-introduction", "hash", "session-guard", "access-tokens", "stateless-guard", "authorization",
	"shield", "cors", "csrf", "rate-limiting", "health",
	"application-lifecycle", "dependency-injection", "service-providers", "plugins", "container-services",
	"helpers-string", "helpers-collection", "helpers-time", "helpers-pipeline",
	"cache", "cache-remember", "cache-backends", "cache-invalidation",
	"storage", "drive", "transmit", "events", "logger", "mail", "notification", "queue", "scheduler", "websockets",
	"workflow", "feature-flags", "multi-tenancy", "presence", "openapi", "studio", "edge-functions", "metrics",
	"telescope", "horizon", "pulse", "reverb", "socialite", "unpoly",
	"ai", "ai-video", "mcp",
	"cli", "creating-commands", "command-arguments", "command-flags", "prompts", "terminal-ui", "repl", "hot-reload",
	"testing-introduction", "http-tests",
}

// docsRedirect maps old doc slugs to new ones (for backward compatibility).
var docsRedirect = map[string]string{
	"nimbus-templates": "nimbus-template",
	"layouts":          "nimbus-template",
	"views":            "nimbus-template",
	"gettingstarted":   "getting-started",
}

func docsPageHandler(c *http.Context) error {
	page := c.Param("*")
	if redirect, ok := docsRedirect[page]; ok {
		page = redirect
	}
	title, ok := docsTitles[page]
	if !ok {
		if unpoly.IsUnpoly(c) {
			unpoly.EmitEvent(c, "docs:notfound", map[string]any{"page": page})
		}
		return c.JSON(http.StatusNotFound, map[string]string{"error": "not found"})
	}

	if unpoly.IsUnpoly(c) {
		unpoly.SetTitle(c, title+" · Nimbus")
		unpoly.EmitEvent(c, "docs:navigate", map[string]any{
			"page":  page,
			"title": title,
		})
	}

	data := map[string]any{"title": title, "currentSlug": page}
	for i, slug := range docsOrder {
		if slug != page {
			continue
		}
		if i > 0 {
			prev := docsOrder[i-1]
			data["prevSlug"] = prev
			data["prevTitle"] = docsTitles[prev]
		}
		if i < len(docsOrder)-1 {
			next := docsOrder[i+1]
			data["nextSlug"] = next
			data["nextTitle"] = docsTitles[next]
		}
		break
	}

	return c.View("docs/"+page, data)
}

// docsIndexAPIHandler returns the docs index for search (slug + title).
func docsIndexAPIHandler(c *http.Context) error {
	items := make([]map[string]string, 0, len(docsOrder))
	for _, slug := range docsOrder {
		if title, ok := docsTitles[slug]; ok {
			items = append(items, map[string]string{"slug": slug, "title": title})
		}
	}
	return c.JSON(http.StatusOK, items)
}

// docsChatAPIHandler handles AI chat with doc context.
func docsChatAPIHandler(c *http.Context) error {
	var body struct {
		Message     string `json:"message"`
		CurrentPage string `json:"currentPage"`
	}
	if err := json.NewDecoder(c.Request.Body).Decode(&body); err != nil || body.Message == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "message required"})
	}

	system := "You are a Nimbus documentation assistant. Answer ONLY based on the documentation below. Do not make up APIs or features—if it's not in the docs, say so.\n\n" +
		"RULES:\n" +
		"- Use only information from the documentation provided.\n" +
		"- Format with markdown: **bold**, `code`, ```go ... ``` for code blocks.\n" +
		"- For EVERY code block, put the file path on the first line: // File: path/to/file.go (Go) or # File: path (shell). Example:\n" +
		"  ```go\n  // File: start/routes.go\n  package main\n  ...\n  ```\n" +
		"- If the question is not about Nimbus, respond: \"I can only help with Nimbus documentation.\"\n" +
		"- Be concise. Cite the relevant doc section when possible."
	if body.CurrentPage != "" {
		if title, ok := docsTitles[body.CurrentPage]; ok {
			system += "\n\nThe user is viewing: " + title + ". Prioritize this topic."
		}
	}
	system += "\n\n--- NIMBUS DOCUMENTATION ---\n" + docs.GetDocsContext()

	response, err := ai.Generate(c.Request.Context(), body.Message, ai.WithSystem(system))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]string{"text": response.Text})
}
