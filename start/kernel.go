/*
|--------------------------------------------------------------------------
| HTTP Kernel
|--------------------------------------------------------------------------
|
| The HTTP kernel file is used to register the middleware with the
| server or the router.
|
*/

package start

import (
	"github.com/CodeSyncr/nimbus"
	"github.com/CodeSyncr/nimbus/errors"
	"github.com/CodeSyncr/nimbus/middleware"
	"github.com/CodeSyncr/nimbus/packages/shield"
	"github.com/CodeSyncr/nimbus/plugins/telescope"
	"github.com/CodeSyncr/nimbus/plugins/unpoly"
	"github.com/CodeSyncr/nimbus/router"
)

// RegisterMiddleware configures all middleware layers on the application.
func RegisterMiddleware(app *nimbus.App) {

	// ── Server Middleware ──────────────────────────────────
	// Runs on every HTTP request, even if there is no route
	// registered for the request URL.
	shieldCfg := shield.DefaultConfig()
	shieldCfg.CSRF.ExceptPaths = append(shieldCfg.CSRF.ExceptPaths, "/api/docs/chat")

	app.Router.Use(
		middleware.Logger(),
		middleware.Recover(),
		errors.Handler(),
		shield.Guard(shieldCfg),
		shield.CSRFGuard(shieldCfg.CSRF),
		unpoly.ServerProtocol(),
	)
	// Telescope request watcher (must use plugin instance for shared store)
	if te := app.Plugin("telescope"); te != nil {
		if t, ok := te.(*telescope.Plugin); ok {
			app.Router.Use(t.RequestWatcher())
		}
	}

	// ── Router Middleware ──────────────────────────────────
	// Runs on all HTTP requests with a registered route.
	// Uncomment as needed:
	//
	// app.Router.Use(
	//     middleware.CORS(),
	//     middleware.CSRF(),
	// )
}

// ── Named Middleware ────────────────────────────────────────
// Named middleware must be explicitly assigned to routes or
// route groups. Reference them in start/routes.go:
//
//   app.Router.Get("/dashboard", handler).Use(Middleware["auth"])
//   admin := app.Router.Group("/admin", Middleware["auth"])

var Middleware = map[string]router.Middleware{
	// "auth":  middleware.RequireAuth(),
	// "guest": guestMiddleware(),
}
