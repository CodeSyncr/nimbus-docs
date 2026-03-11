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
	demos.Post("/todo/:id/update", func(c *http.Context) error { return todoCtrl.Update(c) })
	demos.Post("/todo/:id/delete", func(c *http.Context) error { return todoCtrl.Destroy(c) })
	demos.Get("/counter", func(c *http.Context) error { return (&controllers.Counter{}).Index(c) })
	demos.Post("/counter/increment", func(c *http.Context) error { return (&controllers.Counter{}).Increment(c) })
	demos.Post("/counter/decrement", func(c *http.Context) error { return (&controllers.Counter{}).Decrement(c) })
	demos.Post("/counter/set", func(c *http.Context) error { return (&controllers.Counter{}).Set(c) })
	aiCtrl := &controllers.AI{}
	demos.Get("/ai", func(c *http.Context) error { return aiCtrl.Index(c) })
	demos.Post("/ai/generate", func(c *http.Context) error { return aiCtrl.Generate(c) })
	demos.Get("/mcp", mcpDemoHandler)

	// Queue demo: dispatch a welcome email job to the "default" queue.
	demos.Post("/queue/demo", queueDemoHandler)

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

func queueDemoHandler(c *http.Context) error {
	job := &jobs.SendWelcomeEmail{
		UserID: 1,
		Email:  "queue-demo@example.com",
	}
	if err := queue.Dispatch(job).Dispatch(c.Request.Context()); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to queue job"})
	}
	return c.JSON(http.StatusAccepted, map[string]string{"status": "queued"})
}

// ── Documentation ────────────────────────────────────────────

var docsTitles = map[string]string{
	// Start
	"introduction":     "Introduction",
	"installation":     "Installation",
	"folder-structure": "Folder Structure",
	"configuration":    "Configuration",
	"deployment":       "Deployment",
	"faqs":             "FAQs",
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
	// Auth
	"auth":              "Auth & Guards",
	"auth-introduction": "Auth Introduction",
	"hash":              "Hashing",
	"session-guard":     "Session Guard",
	"access-tokens":     "Access Tokens",
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
	// Digging Deeper
	"cache":              "Cache",
	"cache-remember":     "Remember & Type-Safe API",
	"cache-backends":     "Cache Backends",
	"cache-invalidation": "Cache Invalidation & ORM Hooks",
	"storage":            "Storage",
	"drive":              "Drive",
	"transmit":           "Transmit",
	"events":             "Events",
	"logger":             "Logger",
	"mail":               "Mail",
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
	"ai":  "AI SDK",
	"mcp": "MCP",
	// Testing
	"testing-introduction": "Testing Introduction",
	"http-tests":           "HTTP Tests",
}

func docsIndexHandler(c *http.Context) error {
	if unpoly.IsUnpoly(c) {
		unpoly.SetTitle(c, "Documentation · Nimbus")
	}
	return c.View("docs/index", map[string]any{"title": "Documentation"})
}

// docsOrder is the reading order for prev/next navigation.
var docsOrder = []string{
	"introduction", "installation", "folder-structure", "configuration", "deployment", "faqs",
	"routing", "controllers", "http-context", "middleware", "request", "response", "body-parser",
	"validation", "file-uploads", "session", "exception-handling", "static-files",
	"nimbus-template",
	"inertia", "inertia-setup", "inertia-hmr",
	"database", "database-query-select", "database-query-insert", "database-query-raw",
	"database-migrations-intro", "database-migrations-schema", "database-migrations-table",
	"database-models-intro", "database-models-schema-classes", "database-models-crud",
	"database-models-hooks", "database-models-query-builder", "database-models-naming-strategy",
	"database-models-query-scopes", "database-models-serializing", "database-models-relationships",
	"database-models-factories",
	"migrations", "seeders",
	"auth", "auth-introduction", "hash", "session-guard", "access-tokens", "authorization",
	"shield", "cors", "csrf", "rate-limiting", "health",
	"application-lifecycle", "dependency-injection", "service-providers", "plugins", "container-services",
	"cache", "cache-remember", "cache-backends", "cache-invalidation",
	"storage", "drive", "transmit", "events", "logger", "mail", "queue", "scheduler", "websockets",
	"cli", "creating-commands", "command-arguments", "command-flags", "prompts", "terminal-ui", "repl", "hot-reload",
	"ai", "mcp",
	"testing-introduction", "http-tests",
}

// docsRedirect maps old doc slugs to new ones (for backward compatibility).
var docsRedirect = map[string]string{
	"nimbus-templates": "nimbus-template",
	"layouts":          "nimbus-template",
	"views":            "nimbus-template",
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
