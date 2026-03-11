package analytics

import (
	"github.com/CodeSyncr/nimbus/http"
	"github.com/CodeSyncr/nimbus/router"
)

// RegisterRoutes mounts the plugin's HTTP routes onto the application router.
func (p *AnalyticsPlugin) RegisterRoutes(r *router.Router) {
	r.Get("/analytics/status", p.statusHandler)
}

func (p *AnalyticsPlugin) statusHandler(c *http.Context) error {
	return c.JSON(http.StatusOK, map[string]string{
		"plugin":  p.Name(),
		"version": p.Version(),
		"status":  "ok",
	})
}
