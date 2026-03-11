package analytics

// DefaultConfig returns the default configuration for the Analytics plugin.
// These values can be overridden by the application's .env file.
func (p *AnalyticsPlugin) DefaultConfig() map[string]any {
	return map[string]any{
		"enabled": true,
	}
}
