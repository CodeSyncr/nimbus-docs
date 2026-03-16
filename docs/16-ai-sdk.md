# AI SDK

> **Multi-provider AI integration** — generate text, stream responses, and build AI-powered features with OpenAI, Anthropic, Google Gemini, Cohere, Mistral, xAI, and Ollama.

---

## Introduction

Nimbus includes a first-class AI SDK that lets you integrate generative AI into your application with a unified API. Switch between providers by changing configuration — your application code stays the same. Inspired by Vercel's AI SDK, adapted for Go.

Features:

- **Multi-provider** — OpenAI, Anthropic, Google Gemini, Cohere, Mistral, xAI (Grok), Ollama (local)
- **Unified API** — `ai.Generate()` and `ai.Stream()` work identically across all providers
- **Configurable** — Choose models, temperature, system prompts per request
- **Plugin architecture** — Register the AI plugin to enable SDK globally
- **Streaming support** — Real-time token-by-token response streaming
- **Local development** — Use Ollama for free, offline AI development

---

## Configuration

```env
AI_PROVIDER=openai           # openai | anthropic | gemini | cohere | mistral | xai | ollama
AI_API_KEY=sk-...            # Provider API key
AI_MODEL=gpt-4o              # Default model
AI_BASE_URL=                 # Custom API base URL (optional, required for Ollama)
```

### Plugin Registration

```go
// bin/server.go
import "github.com/CodeSyncr/nimbus/plugins/ai"

app.Use(ai.New())
```

This registers the AI plugin globally, making `ai.Generate()` and `ai.Stream()` available throughout your application.

---

## Quick Start

### Generate Text

```go
import "github.com/CodeSyncr/nimbus/packages/ai"

result, err := ai.Generate(ctx, "Explain quantum computing in simple terms")
if err != nil {
    return err
}
fmt.Println(result) // "Quantum computing uses quantum bits (qubits)..."
```

### With Options

```go
result, err := ai.Generate(ctx, "Write a haiku about Go programming",
    ai.WithModel("gpt-4o"),
    ai.WithTemperature(0.9),
    ai.WithMaxTokens(100),
    ai.WithSystemPrompt("You are a creative poet."),
)
```

---

## Providers

### OpenAI

```env
AI_PROVIDER=openai
AI_API_KEY=sk-...
AI_MODEL=gpt-4o
```

Available models: `gpt-4o`, `gpt-4o-mini`, `gpt-4-turbo`, `gpt-3.5-turbo`

### Anthropic (Claude)

```env
AI_PROVIDER=anthropic
AI_API_KEY=sk-ant-...
AI_MODEL=claude-sonnet-4-20250514
```

Available models: `claude-sonnet-4-20250514`, `claude-3-5-sonnet-20241022`, `claude-3-haiku-20240307`

### Google Gemini

```env
AI_PROVIDER=gemini
AI_API_KEY=AIza...
AI_MODEL=gemini-1.5-pro
```

Available models: `gemini-1.5-pro`, `gemini-1.5-flash`, `gemini-pro`

### Cohere

```env
AI_PROVIDER=cohere
AI_API_KEY=...
AI_MODEL=command-r-plus
```

### Mistral

```env
AI_PROVIDER=mistral
AI_API_KEY=...
AI_MODEL=mistral-large-latest
```

### xAI (Grok)

```env
AI_PROVIDER=xai
AI_API_KEY=xai-...
AI_MODEL=grok-2
```

### Ollama (Local)

Free, offline AI development on your machine:

```env
AI_PROVIDER=ollama
AI_BASE_URL=http://localhost:11434
AI_MODEL=llama3.1
```

```bash
# Install Ollama
brew install ollama
ollama pull llama3.1
ollama serve
```

---

## Generation Options

| Option | Description | Default |
|--------|-------------|---------|
| `WithModel(name)` | Override the default model | From env |
| `WithTemperature(t)` | Creativity (0.0 = deterministic, 1.0 = creative) | 0.7 |
| `WithMaxTokens(n)` | Maximum response length | Provider default |
| `WithSystemPrompt(s)` | System instruction for the AI | None |

```go
result, err := ai.Generate(ctx, prompt,
    ai.WithModel("gpt-4o"),
    ai.WithTemperature(0.3),
    ai.WithMaxTokens(500),
    ai.WithSystemPrompt("You are a helpful coding assistant."),
)
```

---

## Using in Controllers

The nimbus-starter includes an AI demo controller:

```go
// app/controllers/ai.go
package controllers

import (
    "github.com/CodeSyncr/nimbus/http"
    "github.com/CodeSyncr/nimbus/packages/ai"
)

type AIController struct{}

func (ctrl *AIController) Generate(c *http.Context) error {
    var body struct {
        Prompt string
    }

    if err := json.NewDecoder(c.Request.Body).Decode(&body); err != nil {
        return c.JSON(400, map[string]string{"error": "invalid request"})
    }

    result, err := ai.Generate(c.Request.Context(), body.Prompt)
    if err != nil {
        return c.JSON(500, map[string]string{"error": err.Error()})
    }

    return c.JSON(200, map[string]string{"result": result})
}
```

---

## Real-Life Examples

### AI-Powered Search

```go
func (ctrl *SearchController) SmartSearch(c *http.Context) error {
    query := c.Query("q")

    // Use AI to understand intent and generate SQL
    sqlQuery, err := ai.Generate(c.Request.Context(),
        fmt.Sprintf("Convert this natural language query to a SQL WHERE clause for a products table (columns: name, description, category, price): %s", query),
        ai.WithModel("gpt-4o"),
        ai.WithTemperature(0.1),
        ai.WithSystemPrompt("Return ONLY the WHERE clause, no explanation. Use ILIKE for text matching."),
    )
    if err != nil {
        // Fallback to simple LIKE search
        var products []Product
        db.Where("name ILIKE ?", "%"+query+"%").Find(&products)
        return c.JSON(200, products)
    }

    var products []Product
    db.Where(sqlQuery).Find(&products)
    return c.JSON(200, products)
}
```

### Content Generation

```go
func (ctrl *BlogController) GenerateDraft(c *http.Context) error {
    var req struct {
        Topic    string
        Keywords []string
        Tone     string // professional, casual, technical
    }
    json.NewDecoder(c.Request.Body).Decode(&req)

    prompt := fmt.Sprintf(
        "Write a blog post about '%s'. Include these keywords: %s. Tone: %s. Format in Markdown.",
        req.Topic, strings.Join(req.Keywords, ", "), req.Tone,
    )

    draft, err := ai.Generate(c.Request.Context(), prompt,
        ai.WithModel("gpt-4o"),
        ai.WithMaxTokens(2000),
        ai.WithTemperature(0.7),
    )
    if err != nil {
        return c.JSON(500, map[string]string{"error": "generation failed"})
    }

    return c.JSON(200, map[string]string{"draft": draft})
}
```

### AI Documentation Chat

The nimbus-starter includes AI-powered documentation chat:

```go
// start/routes.go
docs.Post("/docs/ai/chat", func(c *http.Context) error {
    var body struct {
        Question string
    }
    json.NewDecoder(c.Request.Body).Decode(&body)

    result, err := ai.Generate(
        c.Request.Context(),
        body.Question,
        ai.WithSystemPrompt("You are a Nimbus framework expert. Answer questions about the Nimbus Go web framework."),
    )
    if err != nil {
        return c.JSON(500, map[string]string{"error": err.Error()})
    }

    return c.JSON(200, map[string]string{"answer": result})
})
```

### Customer Support Bot

```go
func (ctrl *SupportController) Chat(c *http.Context) error {
    var req struct {
        Message   string
        SessionID string
    }
    json.NewDecoder(c.Request.Body).Decode(&req)

    // Load conversation history from cache
    key := "chat:" + req.SessionID
    history, _ := cache.RememberT[[]string](key, time.Hour, func() ([]string, error) {
        return []string{}, nil
    })

    // Build context from FAQ database
    var faqs []FAQ
    db.Where("category = ?", "general").Find(&faqs)
    faqContext := formatFAQs(faqs)

    systemPrompt := fmt.Sprintf(`You are a helpful customer support agent.
Use this FAQ to answer questions:
%s

Previous messages: %s`, faqContext, strings.Join(history, "\n"))

    response, err := ai.Generate(c.Request.Context(), req.Message,
        ai.WithSystemPrompt(systemPrompt),
        ai.WithTemperature(0.3),
    )
    if err != nil {
        return c.JSON(500, map[string]string{"error": "AI unavailable"})
    }

    // Save to history
    history = append(history, "User: "+req.Message, "Bot: "+response)
    cache.Set(key, history, time.Hour)

    return c.JSON(200, map[string]string{"response": response})
}
```

### Code Review Assistant

```go
func reviewCode(ctx context.Context, code string, language string) (string, error) {
    return ai.Generate(ctx, code,
        ai.WithSystemPrompt(fmt.Sprintf(`You are a senior %s developer doing code review.
Review the following code and provide:
1. Potential bugs or issues
2. Performance improvements
3. Security concerns
4. Code style suggestions
Be concise and actionable.`, language)),
        ai.WithTemperature(0.2),
        ai.WithMaxTokens(1000),
    )
}
```

---

## Error Handling

```go
result, err := ai.Generate(ctx, prompt)
if err != nil {
    switch {
    case strings.Contains(err.Error(), "rate_limit"):
        // Wait and retry
        time.Sleep(time.Second)
        result, err = ai.Generate(ctx, prompt)
    case strings.Contains(err.Error(), "context_length"):
        // Prompt too long, truncate
        result, err = ai.Generate(ctx, prompt[:2000])
    default:
        logger.Error("AI generation failed", "error", err)
        // Fallback to non-AI response
    }
}
```

---

## Best Practices

1. **Use low temperature for factual tasks** — 0.1-0.3 for code, SQL, classifications
2. **Use higher temperature for creative tasks** — 0.7-0.9 for writing, brainstorming
3. **Write clear system prompts** — Define the AI's role and constraints
4. **Limit max tokens** — Prevent runaway costs with `WithMaxTokens()`
5. **Handle errors gracefully** — Always have a non-AI fallback
6. **Use Ollama for development** — Free, fast, no API key needed
7. **Cache AI responses** — Use `cache.Remember` for repeated queries
8. **Sanitize user input** — Don't pass untrusted input directly to prompts
9. **Monitor usage** — Track costs per provider via Telescope

**Next:** [MCP (Model Context Protocol)](17-mcp.md) →
