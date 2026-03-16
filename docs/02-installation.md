# Installation & Setup

> **Get a Nimbus project running in under 2 minutes** — from installing the CLI to seeing your first page in the browser.

---

## Prerequisites

Before installing Nimbus, ensure you have the following:

| Requirement | Minimum Version | Check Command |
|-------------|----------------|---------------|
| **Go** | 1.22+ (recommended 1.26) | `go version` |
| **Git** | Any recent version | `git --version` |
| **Database** (optional) | PostgreSQL 14+, MySQL 8+, or SQLite 3 | `psql --version` / `sqlite3 --version` |

> **Note:** SQLite is the default database and requires no external setup — perfect for development and small projects.

---

## Installing the Nimbus CLI

The Nimbus CLI is your primary tool for creating projects, generating code, running migrations, and managing your application.

```bash
# Install the latest Nimbus CLI
go install github.com/CodeSyncr/nimbus/cmd/nimbus@latest

# Verify installation
nimbus --version
# Nimbus CLI v0.1.4
```

> **Tip:** Make sure `$GOPATH/bin` (usually `~/go/bin`) is in your `PATH`.

---

## Creating a New Project

### Using the CLI (Recommended)

```bash
# Create a new project with interactive prompts
nimbus new my-app

# You'll be asked:
# ? Project name: my-app
# ? Database driver: sqlite / postgres / mysql
# ? Include AI SDK? Yes / No
# ? Include Inertia.js (SPA)? Yes / No
```

This scaffolds a complete project with the recommended directory structure, pre-configured settings, and example code.

### Manual Setup (From Starter Kit)

If you prefer to clone the starter directly:

```bash
# Clone the starter kit
git clone https://github.com/CodeSyncr/nimbus-starter.git my-app
cd my-app

# Remove the git history and start fresh
rm -rf .git
git init

# Install dependencies
go mod tidy

# Copy environment file
cp .env.example .env
```

---

## Project Structure After Scaffolding

```
my-app/
├── main.go                 # Application entry point
├── go.mod                  # Go module definition
├── .env                    # Environment variables (never commit!)
├── .air.toml               # Hot reload configuration
│
├── app/                    # Application code
│   ├── controllers/        # HTTP controllers
│   │   └── hello_world.go
│   ├── models/             # Database models (GORM)
│   ├── middleware/          # Custom middleware
│   ├── validators/          # Request validators
│   ├── jobs/               # Background jobs (queue)
│   └── plugins/            # Custom plugins
│
├── bin/                    # Application bootstrap
│   └── server.go           # Boot() — wires everything together
│
├── config/                 # Configuration files (19 files scaffolded)
│   ├── config.go           # Master loader — calls all loadXxx() functions
│   ├── env.go              # env(), envInt(), envBool() helpers
│   ├── app.go              # App name, port, env, host, key
│   ├── database.go         # DB driver, DSN, connection fields
│   ├── bodyparser.go       # JSON/form/multipart size limits
│   ├── cache.go            # Cache driver + TTL
│   ├── cors.go             # CORS origins, methods, credentials
│   ├── hash.go             # Bcrypt driver/cost
│   ├── limiter.go          # Rate limiting rules
│   ├── logger.go           # Log level + format
│   ├── mail.go             # SMTP driver settings
│   ├── queue.go            # Queue driver (sync/redis/sqs/kafka)
│   ├── session.go          # Session driver, cookie settings
│   ├── shield.go           # Security headers + CSRF
│   ├── static.go           # Static file serving
│   └── storage.go          # File storage driver
│
├── database/
│   ├── migrations/         # Schema migrations
│   │   └── registry.go     # Migration registry
│   └── seeders/            # Database seeders
│
├── start/                  # Application bootstrap hooks
│   ├── routes.go           # Route definitions
│   ├── kernel.go           # Middleware registration
│   ├── jobs.go             # Queue job registration
│   └── schedule.go         # Scheduled task registration
│
├── resources/
│   ├── views/              # .nimbus templates
│   │   ├── home.nimbus     # Home page template
│   │   └── layout.nimbus   # Main layout
│   ├── css/                # Stylesheets
│   └── js/                 # JavaScript files
│
├── public/                 # Static files (served directly)
│   ├── css/
│   └── js/
│
├── storage/                # App-generated files (logs, uploads)
├── Dockerfile              # Docker build
├── deploy.yaml             # Deployment config
└── render.yaml             # Render.com config
```

---

## Environment Configuration

Create a `.env` file in your project root:

```env
# Application
APP_NAME=my-app
APP_ENV=development
APP_PORT=3333
APP_KEY=your-secret-key-here-min-32-chars

# Database
DB_DRIVER=sqlite
DB_DSN=database.sqlite

# For PostgreSQL:
# DB_DRIVER=postgres
# DB_HOST=localhost
# DB_PORT=5432
# DB_USER=postgres
# DB_PASSWORD=secret
# DB_DATABASE=my_app

# For MySQL:
# DB_DRIVER=mysql
# DB_HOST=localhost
# DB_PORT=3306
# DB_USER=root
# DB_PASSWORD=secret
# DB_DATABASE=my_app

# Cache
CACHE_DRIVER=memory
CACHE_TTL_MINUTES=60

# Session
SESSION_DRIVER=cookie
SESSION_COOKIE=nimbus_session

# Queue
QUEUE_DRIVER=memory

# AI (optional)
# AI_PROVIDER=openai
# OPENAI_API_KEY=sk-...

# Mail (optional)
# MAIL_DRIVER=smtp
# SMTP_HOST=smtp.mailtrap.io
# SMTP_PORT=587
# SMTP_USERNAME=your-username
# SMTP_PASSWORD=your-password
```

---

## Running the Application

### Development Mode with Hot Reload

```bash
# Start with hot reload (recommended for development)
nimbus serve

# Or manually with air:
air

# Or without hot reload:
go run main.go
```

Your application starts at **http://localhost:3333** by default.

### What Happens on Startup

When you run `bin.Boot()`, the following sequence executes:

```
1. config.Load()           → Load .env, parse config files
2. nimbus.New()             → Create App with router, container, events, scheduler
3. bootMail()              → Configure mail driver (SMTP)
4. bootCache()             → Initialize cache backend (memory/Redis/...)
5. bootDatabase(app)       → Connect to DB, register in container
6. bootQueue()             → Initialize queue system, register jobs
7. registerPlugins(app)    → Register all plugins (Horizon, Shield, AI, Telescope, MCP, ...)
8. registerMiddleware(app) → Apply middleware stack (Logger, Recover, Shield, CSRF, ...)
9. registerRoutes(app)     → Define all route handlers
```

Then `app.Run()`:
```
1. app.Boot()              → Run provider/plugin lifecycle
2. Listen on :3333         → Start HTTP server
3. Start scheduler         → Run cron-like tasks
4. Signal handler          → SIGINT/SIGTERM → graceful shutdown
```

---

## Running Database Migrations

```bash
# Run all pending migrations
go run main.go migrate
# Or: nimbus db:migrate

# Seed the database with test data
go run main.go seed
# Or: nimbus db:seed
```

---

## Real-Life Example: Setting Up a Blog API

Here's a complete walkthrough of creating a new Nimbus project for a blog API:

```bash
# 1. Create the project
nimbus new blog-api
cd blog-api

# 2. Configure PostgreSQL
cat > .env << 'EOF'
APP_NAME=blog-api
APP_ENV=development
APP_PORT=3333
APP_KEY=my-super-secret-key-for-blog-api-32ch
DB_DRIVER=postgres
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=secret
DB_DATABASE=blog_api
CACHE_DRIVER=memory
QUEUE_DRIVER=memory
EOF

# 3. Generate a model and migration
nimbus make:model Post --migration
nimbus make:model Comment --migration

# 4. Generate controllers
nimbus make:controller PostController --resource
nimbus make:controller CommentController --resource

# 5. Run migrations
go run main.go migrate

# 6. Start the server
nimbus serve
```

Then add routes in `start/routes.go`:

```go
func RegisterRoutes(app *nimbus.App) {
    api := app.Router.Group("/api/v1")
    api.Resource("posts",    &controllers.PostController{DB: app.Container.MustMake("db").(*nimbus.DB)})
    api.Resource("comments", &controllers.CommentController{DB: app.Container.MustMake("db").(*nimbus.DB)})
}
```

You now have a full RESTful API with:
- `GET /api/v1/posts` — List all posts
- `POST /api/v1/posts` — Create a post
- `GET /api/v1/posts/:id` — Get a single post
- `PUT /api/v1/posts/:id` — Update a post
- `DELETE /api/v1/posts/:id` — Delete a post

---

## Verifying Your Setup

After starting the server, verify everything works:

```bash
# Health check
curl http://localhost:3333/health
# {"status":"ok","checks":[{"name":"database","status":"ok","duration":"2ms"}]}

# Home page
curl http://localhost:3333/
# Returns the welcome page HTML

# API endpoint
curl http://localhost:3333/api/v1/posts
# []
```

---

## IDE Setup

### VS Code (Recommended)

Install the **Nimbus for VS Code** extension for:
- `.nimbus` template syntax highlighting
- Go to definition for routes
- Code snippets for controllers, models, middleware
- Integrated terminal commands

### GoLand / IntelliJ

Standard Go plugin works. Add `.nimbus` file association to HTML for syntax highlighting.

---

## Troubleshooting

| Problem | Solution |
|---------|----------|
| `nimbus: command not found` | Ensure `~/go/bin` is in your `PATH` |
| `database connection failed` | Check `.env` credentials and that the database server is running |
| Port 3333 in use | Set `APP_PORT=3334` in `.env` or stop the conflicting process |
| `air: command not found` | Install air: `go install github.com/cosmtrek/air@latest` |
| Templates not updating | Air may not watch `.nimbus` files — add `include_ext = ["go", "nimbus"]` to `.air.toml` |

**Next:** [Folder Structure](03-folder-structure.md) →
