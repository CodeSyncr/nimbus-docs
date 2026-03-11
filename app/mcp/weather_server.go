package mcp

import (
	"context"
	"fmt"
	"strings"

	nimbusmcp "github.com/CodeSyncr/nimbus/plugins/mcp"
	"github.com/mark3labs/mcp-go/mcp"
)

// WeatherServer is the MCP demo server with weather tools and resources.
var WeatherServer = nimbusmcp.NewServer("Weather Demo", "1.0.0",
	nimbusmcp.WithInstructions("This server provides weather information and forecasts. Use get_weather for current conditions, or read the forecast resource."),
)

func init() {
	WeatherServer.AddTool(
		mcp.NewTool("get_weather",
			mcp.WithDescription("Get the current weather for a location"),
			mcp.WithString("location", mcp.Required(), mcp.Description("City or location name (e.g. London, Tokyo)")),
			mcp.WithString("units", mcp.Enum("celsius", "fahrenheit"), mcp.DefaultString("celsius")),
		),
		handleGetWeather,
	)

	WeatherServer.AddTool(
		mcp.NewTool("hello",
			mcp.WithDescription("Say hello to someone"),
			mcp.WithString("name", mcp.Required(), mcp.Description("Name to greet")),
		),
		handleHello,
	)

	WeatherServer.AddResource(
		mcp.NewResourceTemplate("weather://forecast/{location}", "Weather Forecast",
			mcp.WithTemplateDescription("Weather forecast for a location"),
			mcp.WithTemplateMIMEType("application/json"),
		),
		handleForecastResource,
	)
}

func handleGetWeather(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	location, err := req.RequireString("location")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	units := req.GetString("units", "celsius")
	temp := "22"
	if units == "fahrenheit" {
		temp = "72"
	}
	return mcp.NewToolResultText(fmt.Sprintf("Weather in %s: Sunny, %s°%s. Light breeze, low humidity.", location, temp, units)), nil
}

func handleHello(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	name, err := req.RequireString("name")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return mcp.NewToolResultText(fmt.Sprintf("Hello, %s! Welcome to the Nimbus MCP demo.", name)), nil
}

func handleForecastResource(ctx context.Context, req mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	uri := req.Params.URI
	// Extract location from URI (e.g. weather://forecast/london -> london)
	location := "unknown"
	if parts := strings.Split(uri, "/"); len(parts) >= 4 {
		location = parts[len(parts)-1]
	}
	content := fmt.Sprintf(`{"location":"%s","forecast":[{"day":"Today","high":22,"low":14,"conditions":"Sunny"},{"day":"Tomorrow","high":24,"low":15,"conditions":"Partly cloudy"}]}`, location)
	return []mcp.ResourceContents{
		mcp.TextResourceContents{URI: uri, MIMEType: "application/json", Text: content},
	}, nil
}
