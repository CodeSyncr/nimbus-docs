package analytics

import "github.com/CodeSyncr/nimbus"

// Compile-time interface checks.
var (
	_ nimbus.Plugin        = (*AnalyticsPlugin)(nil)
	_ nimbus.HasRoutes     = (*AnalyticsPlugin)(nil)
	_ nimbus.HasMiddleware = (*AnalyticsPlugin)(nil)
	_ nimbus.HasConfig     = (*AnalyticsPlugin)(nil)
)

// AnalyticsPlugin is a Nimbus plugin.
type AnalyticsPlugin struct {
	nimbus.BasePlugin
}

// New creates a new AnalyticsPlugin instance.
func New() *AnalyticsPlugin {
	return &AnalyticsPlugin{
		BasePlugin: nimbus.BasePlugin{
			PluginName:    "analytics",
			PluginVersion: "0.1.0",
		},
	}
}
