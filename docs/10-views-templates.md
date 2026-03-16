# Views & Templates

> **Edge-inspired template engine** — layouts, components, slots, partials, and expressive syntax with `.nimbus` templates.

---

## Introduction

Nimbus includes a custom template engine inspired by AdonisJS's **Edge** template engine. It uses `.nimbus` files with a clean, expressive syntax that's easy to read and write:

- **Layouts** — Define page shells with `@layout('name')`
- **Components** — Reusable UI pieces with slots: `@card()...@end`
- **Partials** — Include fragments with `@include('path')`
- **Variable output** — `{{ .var }}` (escaped) and `{{{ .raw }}}` (unescaped)
- **Control flow** — `@if`, `@else`, `@each` for loops
- **Debugging** — `@dump(var)` for variable inspection
- **CSRF integration** — Auto-injected CSRF fields in forms

---

## Template Syntax

### Variable Output

```html
{{-- Escaped output (HTML-safe) --}}
<h1>{{ .title }}</h1>
<p>Welcome, {{ .user.Name }}</p>

{{-- Raw/unescaped output (for trusted HTML) --}}
{{{ .htmlContent }}}
{{{ .csrfField }}}

{{-- Comments (stripped from output) --}}
{{-- This will not appear in the rendered HTML --}}
```

### Conditionals

```html
@if(.user)
    <p>Welcome back, {{ .user.Name }}!</p>
@else
    <p>Please <a href="/login">log in</a>.</p>
@endif

@if(.items)
    @if(.empty)
        <p>No items found.</p>
    @else
        <p>Found {{ .count }} items.</p>
    @endif
@endif
```

### Loops

```html
@each(item in .items)
    <div class="todo-item">
        <h3>{{ .item.Title }}</h3>
        @if(.item.Done)
            <span class="badge">✓ Done</span>
        @else
            <span class="badge">Pending</span>
        @endif
    </div>
@endeach

{{-- Empty state --}}
@if(.empty)
    <div class="empty-state">
        <p>Nothing here yet. Create your first item!</p>
    </div>
@endif
```

### Debugging

```html
{{-- Dump a variable for debugging --}}
@dump(.user)
@dump(.items)
```

---

## Layouts

Layouts define the outer shell of your pages (HTML head, nav, footer):

### Defining a Layout

```html
{{-- resources/views/layout.nimbus --}}
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{ .title }} · My App</title>
    <link rel="stylesheet" href="/public/css/app.css">
</head>
<body>
    <nav class="navbar">
        <a href="/">Home</a>
        <a href="/about">About</a>
        @if(.user)
            <a href="/dashboard">Dashboard</a>
            <span>{{ .user.Name }}</span>
        @else
            <a href="/login">Login</a>
        @endif
    </nav>

    <main class="container">
        {{{ .Content }}}
    </main>

    <footer>
        <p>© 2024 My App. Built with Nimbus.</p>
    </footer>

    <script src="/public/js/app.js"></script>
</body>
</html>
```

### Using a Layout

```html
{{-- resources/views/home.nimbus --}}
@layout('layout')

<div class="hero">
    <h1>{{ .title }}</h1>
    <p>{{ .tagline }}</p>
    <a href="/get-started" class="btn">Get Started</a>
</div>

<section class="features">
    <div class="feature">
        <h3>Fast</h3>
        <p>Built on Go for blazing performance.</p>
    </div>
    <div class="feature">
        <h3>Simple</h3>
        <p>Convention over configuration.</p>
    </div>
</section>
```

### Nested Layouts

```html
{{-- resources/views/layout-docs.nimbus --}}
@layout('layout')

<div class="docs-layout">
    <aside class="docs-sidebar">
        @include('docs/sidebar')
    </aside>
    <article class="docs-content">
        {{{ .Content }}}
    </article>
</div>
```

```html
{{-- resources/views/docs/routing.nimbus --}}
@layout('layout-docs')

<h1>Routing</h1>
<p>Learn how to define routes in Nimbus...</p>
```

---

## Components

Components are reusable UI blocks with slots:

### Defining a Component

```html
{{-- resources/views/components/card.nimbus --}}
<div class="card">
    @if(.title)
        <div class="card-header">
            <h3>{{ .title }}</h3>
        </div>
    @endif
    <div class="card-body">
        {{{ .slots.main }}}
    </div>
    @if(.footer)
        <div class="card-footer">
            {{ .footer }}
        </div>
    @endif
</div>
```

### Using a Component

```html
@card(title="Recent Orders" footer="View all orders")
    <table>
        <tr><th>Order</th><th>Total</th><th>Status</th></tr>
        @each(order in .orders)
            <tr>
                <td>#{{ .order.ID }}</td>
                <td>{{ .order.FormattedTotal }}</td>
                <td>{{ .order.Status }}</td>
            </tr>
        @endeach
    </table>
@end
```

### Component Examples

#### Alert Component

```html
{{-- resources/views/components/alert.nimbus --}}
<div class="alert alert-{{ .type }}">
    @if(.dismissible)
        <button class="alert-close">&times;</button>
    @endif
    {{{ .slots.main }}}
</div>
```

```html
@alert(type="success" dismissible="true")
    Your order has been placed successfully!
@end

@alert(type="error")
    Payment failed. Please try again.
@end
```

#### Modal Component

```html
{{-- resources/views/components/modal.nimbus --}}
<div class="modal" id="{{ .id }}">
    <div class="modal-backdrop"></div>
    <div class="modal-content">
        <div class="modal-header">
            <h3>{{ .title }}</h3>
            <button class="modal-close">&times;</button>
        </div>
        <div class="modal-body">
            {{{ .slots.main }}}
        </div>
    </div>
</div>
```

---

## Partials

Include reusable template fragments:

```html
{{-- resources/views/partials/header.nimbus --}}
<header class="site-header">
    <div class="logo">
        <a href="/">{{ .appName }}</a>
    </div>
    <nav>
        <a href="/products">Products</a>
        <a href="/about">About</a>
        <a href="/contact">Contact</a>
    </nav>
</header>
```

```html
{{-- Include in any template --}}
@include('partials/header')
```

---

## Rendering Views from Controllers

### Basic View Rendering

```go
func homeHandler(c *http.Context) error {
    return c.View("home", map[string]any{
        "title":   "Welcome",
        "appName": "Nimbus",
        "tagline": "The Go framework for humans",
        "version": "0.1.4",
    })
}
```

### With Model Data

```go
func (todo *Todo) Index(ctx *http.Context) error {
    var items []models.Todo
    todo.DB.Find(&items)

    doneCount := 0
    for _, it := range items {
        if it.Done { doneCount++ }
    }

    return ctx.View("apps/todo/index", map[string]any{
        "title":        "Todo",
        "items":        items,
        "empty":        len(items) == 0,
        "doneCount":    doneCount,
        "pendingCount": len(items) - doneCount,
        "donePercent":  (doneCount * 100) / max(len(items), 1),
    })
}
```

### CSRF Token Auto-Injection

When Shield is enabled, Nimbus automatically adds a `csrfField` variable to every view:

```html
<form method="POST" action="/todos">
    {{{ .csrfField }}}
    <input type="text" name="title" placeholder="What needs to be done?" />
    <button type="submit">Add</button>
</form>
```

---

## Real-Life Example: Todo Application

### Layout

```html
{{-- resources/views/apps/layout.nimbus --}}
@layout('layout')

<div class="app-container">
    <nav class="app-nav">
        <a href="/demos/todo">Todos</a>
        <a href="/demos/counter">Counter</a>
        <a href="/demos/ai">AI</a>
    </nav>
    <div class="app-content">
        {{{ .Content }}}
    </div>
</div>
```

### Todo List

```html
{{-- resources/views/apps/todo/index.nimbus --}}
@layout('apps/layout')

<div class="todo-app">
    <h1>{{ .title }}</h1>

    {{-- Stats bar --}}
    <div class="stats">
        <span>{{ .pendingCount }} pending</span>
        <span>{{ .doneCount }} done</span>
        <div class="progress-bar">
            <div class="progress" style="width: {{ .donePercent }}%"></div>
        </div>
    </div>

    {{-- Add new todo --}}
    <form method="POST" action="/demos/todo">
        {{{ .csrfField }}}
        <input type="text" name="title" placeholder="What needs to be done?" required />
        <button type="submit">Add</button>
    </form>

    {{-- Todo items --}}
    @if(.empty)
        <div class="empty-state">
            <p>🎉 Nothing to do! Add your first task above.</p>
        </div>
    @else
        @each(item in .items)
            <div class="todo-item @if(.item.Done) done @endif">
                <form method="POST" action="/demos/todo/{{ .item.ID }}/toggle">
                    {{{ .csrfField }}}
                    <button type="submit" class="toggle">
                        @if(.item.Done) ✓ @else ○ @endif
                    </button>
                </form>
                <span class="title">{{ .item.Title }}</span>
                <div class="actions">
                    <a href="/demos/todo/{{ .item.ID }}/edit">Edit</a>
                    <form method="POST" action="/demos/todo/{{ .item.ID }}/delete">
                        {{{ .csrfField }}}
                        <button type="submit" class="delete">Delete</button>
                    </form>
                </div>
            </div>
        @endeach
    @endif
</div>
```

### Todo Form (Create/Edit)

```html
{{-- resources/views/apps/todo/form.nimbus --}}
@layout('apps/layout')

<h1>{{ .title }}</h1>

@if(.error)
    @alert(type="error")
        {{ .error }}
    @end
@endif

<form method="POST" action="@if(.item) /demos/todo/{{ .item.ID }}/update @else /demos/todo @endif">
    {{{ .csrfField }}}
    <div class="form-group">
        <label for="title">Title</label>
        <input type="text" name="title" id="title"
               value="@if(.item) {{ .item.Title }} @endif"
               required minlength="1" maxlength="255" />
    </div>
    @if(.item)
        <div class="form-group">
            <label>
                <input type="checkbox" name="done" @if(.item.Done) checked @endif />
                Mark as done
            </label>
        </div>
    @endif
    <button type="submit">@if(.item) Update @else Create @endif</button>
    <a href="/demos/todo">Cancel</a>
</form>
```

---

## Template Engine Internals

### Syntax Conversion

The Nimbus template engine converts `.nimbus` syntax to Go's `html/template`:

| Nimbus Syntax | Go Template | Purpose |
|--------------|-------------|---------|
| `{{ .var }}` | `{{ .var }}` | Escaped output |
| `{{{ .raw }}}` | `{{ raw .expr }}` | Unescaped output |
| `{{-- comment --}}` | *(stripped)* | Comments |
| `@if(.cond)` | `{{ if .cond }}` | Conditional |
| `@else` | `{{ else }}` | Else branch |
| `@endif` | `{{ end }}` | End conditional |
| `@each(x in .list)` | `{{ range .list }}` | Loop |
| `@endeach` | `{{ end }}` | End loop |
| `@layout('name')` | *(layout wrapping)* | Set parent layout |
| `@include('path')` | `{{ template "path" . }}` | Include partial |
| `@component()...@end` | *(component rendering)* | Component with slots |
| `@dump(var)` | `{{ dump .var }}` | Debug output |

### View Engine Configuration

```go
import "github.com/CodeSyncr/nimbus/view"

// The engine auto-detects views in resources/views/ or views/
engine := view.Default()

// Custom view root
engine.SetRoot("custom/views/path")

// Register plugin views (e.g., admin panel, docs)
view.RegisterPluginViews("telescope", telescopeFS)
```

---

## Plugin Views

Plugins can register their own view templates with a namespace prefix:

```go
import "embed"

//go:embed views/*
var pluginViews embed.FS

func (p *MyPlugin) Register(app *nimbus.App) {
    view.RegisterPluginViews("myplugin", pluginViews)
}

// Now accessible as "myplugin/dashboard" in controllers
func handler(c *http.Context) error {
    return c.View("myplugin/dashboard", data)
}
```

---

## Best Practices

1. **Use layouts for consistent structure** — Header, nav, footer in one place
2. **Create components for repeated UI** — Cards, alerts, modals, forms
3. **Use partials for shared fragments** — Sidebar, breadcrumbs, pagination
4. **Escaped output by default** — Use `{{ .var }}` unless you specifically need raw HTML
5. **Don't put logic in templates** — Compute values in controllers, pass simple data to views
6. **Name templates descriptively** — `apps/todo/index.nimbus`, not `page1.nimbus`
7. **Use `@dump`for debugging** — Quick way to inspect template data during development

**Next:** [Plugin System](11-plugins.md) →
