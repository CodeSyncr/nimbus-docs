# MCP (Model Context Protocol)

> **Give AI models access to your application** — expose tools, resources, and data via the Model Context Protocol for seamless AI integration.

---

## Introduction

The **Model Context Protocol (MCP)** is an open standard that lets AI models interact with external systems. Nimbus provides first-class MCP server support, allowing you to expose your application's functionality as **tools** and **resources** that AI models (Claude, GPT, etc.) can use.

Think of MCP as an API specifically designed for AI consumption:

- **Tools** — Functions the AI can call (like API endpoints, but for AI)
- **Resources** — Data the AI can read (like database records, files, configs)
- **Resource Templates** — Dynamic resources with parameters (like parameterized URLs)

---

## Quick Start

### Creating an MCP Server

```go
// app/mcp/weather_server.go
package mcp

import (
    "context"
    "encoding/json"
    "fmt"

    mcpserver "github.com/CodeSyncr/nimbus/packages/mcp"
)

func NewWeatherServer() *mcpserver.Server {
    s := mcpserver.NewServer("weather", "1.0.0")

    // Register a tool
    s.AddTool(mcpserver.Tool{
        Name:        "get_weather",
        Description: "Get current weather for a city",
        InputSchema: json.RawMessage(`{
            "type": "object",
            "properties": {
                "city": {"type": "string", "description": "City name"}
            },
            "required": ["city"]
        }`),
    }, func(ctx context.Context, args map[string]any) (*mcpserver.ToolResult, error) {
        city := args["city"].(string)
        weather := fetchWeather(city)
        return mcpserver.NewToolResult(weather), nil
    })

    return s
}
```

### Registering the Plugin

```go
// bin/server.go
import "github.com/CodeSyncr/nimbus/plugins/mcp"

mcpPlugin := mcp.New()
mcpPlugin.RegisterServer(weatherMCP.NewWeatherServer())
app.Use(mcpPlugin)
```

The MCP server is now available at `/mcp` — AI models can discover and use your tools.

---

## Tools

Tools are functions that AI models can call. Define them with a JSON Schema for the input and a handler function.

### Simple Tool

```go
s.AddTool(mcpserver.Tool{
    Name:        "hello",
    Description: "Say hello to someone",
    InputSchema: json.RawMessage(`{
        "type": "object",
        "properties": {
            "name": {"type": "string", "description": "Name of the person"}
        },
        "required": ["name"]
    }`),
}, func(ctx context.Context, args map[string]any) (*mcpserver.ToolResult, error) {
    name := args["name"].(string)
    return mcpserver.NewToolResult(fmt.Sprintf("Hello, %s!", name)), nil
})
```

### Database Query Tool

```go
s.AddTool(mcpserver.Tool{
    Name:        "search_products",
    Description: "Search products by name or category",
    InputSchema: json.RawMessage(`{
        "type": "object",
        "properties": {
            "query": {"type": "string", "description": "Search term"},
            "category": {"type": "string", "description": "Product category (optional)"},
            "limit": {"type": "integer", "description": "Max results (default: 10)"}
        },
        "required": ["query"]
    }`),
}, func(ctx context.Context, args map[string]any) (*mcpserver.ToolResult, error) {
    query := args["query"].(string)
    limit := 10
    if l, ok := args["limit"].(float64); ok {
        limit = int(l)
    }

    var products []Product
    q := db.Where("name ILIKE ?", "%"+query+"%")
    if cat, ok := args["category"].(string); ok && cat != "" {
        q = q.Where("category = ?", cat)
    }
    q.Limit(limit).Find(&products)

    return mcpserver.NewToolResult(products), nil
})
```

### Action Tool (Create/Update/Delete)

```go
s.AddTool(mcpserver.Tool{
    Name:        "create_todo",
    Description: "Create a new todo item",
    InputSchema: json.RawMessage(`{
        "type": "object",
        "properties": {
            "title": {"type": "string", "description": "Todo title"},
            "priority": {"type": "string", "enum": ["low", "medium", "high"]}
        },
        "required": ["title"]
    }`),
}, func(ctx context.Context, args map[string]any) (*mcpserver.ToolResult, error) {
    todo := models.Todo{
        Title:    args["title"].(string),
        Priority: args["priority"].(string),
    }
    if err := db.Create(&todo).Error; err != nil {
        return nil, fmt.Errorf("failed to create todo: %w", err)
    }
    return mcpserver.NewToolResult(todo), nil
})
```

---

## Resources

Resources expose read-only data to AI models. They're like GET endpoints but designed for AI consumption.

### Static Resource

```go
s.AddResource(mcpserver.Resource{
    URI:         "config://app",
    Name:        "App Configuration",
    Description: "Current application configuration",
    MimeType:    "application/json",
}, func(ctx context.Context) (*mcpserver.ResourceContent, error) {
    config := map[string]any{
        "name":        "MyApp",
        "version":     "1.0.0",
        "environment": os.Getenv("APP_ENV"),
    }
    return mcpserver.NewResourceContent(config), nil
})
```

### Resource Templates

Resource templates have dynamic parameters — the AI can request specific data:

```go
s.AddResourceTemplate(mcpserver.ResourceTemplate{
    URITemplate: "forecast://{city}/daily",
    Name:        "Daily Forecast",
    Description: "Get daily weather forecast for a city",
    MimeType:    "application/json",
}, func(ctx context.Context, params map[string]string) (*mcpserver.ResourceContent, error) {
    city := params["city"]
    forecast := getForecast(city)
    return mcpserver.NewResourceContent(forecast), nil
})
```

---

## Real-Life Examples

### E-Commerce MCP Server

```go
func NewEcommerceMCPServer(db *gorm.DB) *mcpserver.Server {
    s := mcpserver.NewServer("ecommerce", "1.0.0")

    // Tool: Search products
    s.AddTool(mcpserver.Tool{
        Name:        "search_products",
        Description: "Search products in the catalog",
        InputSchema: json.RawMessage(`{
            "type": "object",
            "properties": {
                "query": {"type": "string"},
                "min_price": {"type": "number"},
                "max_price": {"type": "number"},
                "in_stock": {"type": "boolean"}
            },
            "required": ["query"]
        }`),
    }, func(ctx context.Context, args map[string]any) (*mcpserver.ToolResult, error) {
        q := db.Model(&Product{})
        if query, ok := args["query"].(string); ok {
            q = q.Where("name ILIKE ? OR description ILIKE ?", "%"+query+"%", "%"+query+"%")
        }
        if min, ok := args["min_price"].(float64); ok {
            q = q.Where("price >= ?", min)
        }
        if max, ok := args["max_price"].(float64); ok {
            q = q.Where("price <= ?", max)
        }
        if inStock, ok := args["in_stock"].(bool); ok && inStock {
            q = q.Where("stock > 0")
        }

        var products []Product
        q.Limit(20).Find(&products)
        return mcpserver.NewToolResult(products), nil
    })

    // Tool: Get order status
    s.AddTool(mcpserver.Tool{
        Name:        "get_order_status",
        Description: "Check the status of an order by ID",
        InputSchema: json.RawMessage(`{
            "type": "object",
            "properties": {
                "order_id": {"type": "integer"}
            },
            "required": ["order_id"]
        }`),
    }, func(ctx context.Context, args map[string]any) (*mcpserver.ToolResult, error) {
        orderID := int(args["order_id"].(float64))
        var order Order
        err := db.Preload("Items").First(&order, orderID).Error
        if err != nil {
            return mcpserver.NewToolResult(map[string]string{"error": "Order not found"}), nil
        }
        return mcpserver.NewToolResult(order), nil
    })

    // Resource: Store statistics
    s.AddResource(mcpserver.Resource{
        URI:         "stats://store",
        Name:        "Store Statistics",
        Description: "Current store statistics",
    }, func(ctx context.Context) (*mcpserver.ResourceContent, error) {
        var stats struct {
            TotalProducts int64
            TotalOrders   int64
            Revenue       float64
        }
        db.Model(&Product{}).Count(&stats.TotalProducts)
        db.Model(&Order{}).Count(&stats.TotalOrders)
        db.Model(&Order{}).Select("COALESCE(SUM(total), 0)").Scan(&stats.Revenue)
        return mcpserver.NewResourceContent(stats), nil
    })

    return s
}
```

### DevOps MCP Server

```go
func NewDevOpsMCPServer() *mcpserver.Server {
    s := mcpserver.NewServer("devops", "1.0.0")

    // Tool: Health check
    s.AddTool(mcpserver.Tool{
        Name:        "health_check",
        Description: "Run application health checks",
        InputSchema: json.RawMessage(`{"type": "object"}`),
    }, func(ctx context.Context, args map[string]any) (*mcpserver.ToolResult, error) {
        result := healthChecker.Run(ctx)
        return mcpserver.NewToolResult(result), nil
    })

    // Tool: Runtime metrics
    s.AddTool(mcpserver.Tool{
        Name:        "get_metrics",
        Description: "Get application runtime metrics (goroutines, memory, GC)",
        InputSchema: json.RawMessage(`{"type": "object"}`),
    }, func(ctx context.Context, args map[string]any) (*mcpserver.ToolResult, error) {
        stats := metrics.ReadRuntimeStats()
        return mcpserver.NewToolResult(stats), nil
    })

    // Tool: Clear cache
    s.AddTool(mcpserver.Tool{
        Name:        "clear_cache",
        Description: "Clear a cache namespace",
        InputSchema: json.RawMessage(`{
            "type": "object",
            "properties": {
                "namespace": {"type": "string", "description": "Cache namespace to clear"}
            },
            "required": ["namespace"]
        }`),
    }, func(ctx context.Context, args map[string]any) (*mcpserver.ToolResult, error) {
        ns := args["namespace"].(string)
        cache.Namespace(ns).Clear()
        return mcpserver.NewToolResult("Cache cleared: " + ns), nil
    })

    return s
}
```

### CMS MCP Server

```go
func NewCMSMCPServer(db *gorm.DB) *mcpserver.Server {
    s := mcpserver.NewServer("cms", "1.0.0")

    // Tool: Create blog post
    s.AddTool(mcpserver.Tool{
        Name:        "create_post",
        Description: "Create a new blog post",
        InputSchema: json.RawMessage(`{
            "type": "object",
            "properties": {
                "title": {"type": "string"},
                "content": {"type": "string"},
                "tags": {"type": "array", "items": {"type": "string"}},
                "status": {"type": "string", "enum": ["draft", "published"]}
            },
            "required": ["title", "content"]
        }`),
    }, func(ctx context.Context, args map[string]any) (*mcpserver.ToolResult, error) {
        post := Post{
            Title:   args["title"].(string),
            Content: args["content"].(string),
            Status:  "draft",
        }
        if status, ok := args["status"].(string); ok {
            post.Status = status
        }
        db.Create(&post)
        return mcpserver.NewToolResult(post), nil
    })

    // Resource template: Get post by slug
    s.AddResourceTemplate(mcpserver.ResourceTemplate{
        URITemplate: "posts://{slug}",
        Name:        "Blog Post",
        Description: "Get a blog post by its slug",
    }, func(ctx context.Context, params map[string]string) (*mcpserver.ResourceContent, error) {
        var post Post
        err := db.Where("slug = ?", params["slug"]).First(&post).Error
        if err != nil {
            return nil, fmt.Errorf("post not found")
        }
        return mcpserver.NewResourceContent(post), nil
    })

    return s
}
```

---

## Connecting AI Clients

### Claude Desktop

Add to your Claude Desktop config (`claude_desktop_config.json`):

```json
{
  "mcpServers": {
    "myapp": {
      "command": "curl",
      "args": ["-N", "http://localhost:3000/mcp"]
    }
  }
}
```

### Any MCP Client

Your Nimbus MCP server is available at the `/mcp` endpoint. Any MCP-compatible client can connect and discover your tools and resources.

---

## Best Practices

1. **Write clear descriptions** — AI models rely on descriptions to understand tools
2. **Use JSON Schema validation** — Define required fields and types
3. **Return structured data** — JSON is better than plain text for AI processing
4. **Handle errors gracefully** — Return error messages in results, don't panic
5. **Limit scope** — Don't expose destructive operations without safeguards
6. **Add authentication** — Protect MCP endpoints in production
7. **Log tool usage** — Track what AI models are doing via Telescope

**Next:** [CLI](18-cli.md) →
