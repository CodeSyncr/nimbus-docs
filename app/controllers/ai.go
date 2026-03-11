package controllers

import (
	"github.com/CodeSyncr/nimbus/http"
	"github.com/CodeSyncr/nimbus/plugins/ai"
)

// AI controller for the AI SDK demo.
type AI struct{}

// Index shows the AI demo page with a prompt form.
func (a *AI) Index(c *http.Context) error {
	return c.View("apps/ai/index", map[string]any{
		"title": "AI SDK",
	})
}

// Generate handles the prompt submission and returns AI-generated text.
func (a *AI) Generate(c *http.Context) error {
	prompt := c.Request.FormValue("prompt")
	if prompt == "" {
		c.Redirect(http.StatusFound, "/demos/ai")
		return nil
	}

	response, err := ai.Generate(c.Request.Context(), prompt)
	if err != nil {
		return c.View("apps/ai/index", map[string]any{
			"title": "AI SDK",
			"error": err.Error(),
		})
	}

	return c.View("apps/ai/index", map[string]any{
		"title": "AI SDK",
		"text":  response.Text,
	})
}
