# AGENTS.md - Development Guidelines for tibia-char

## Project Overview

This is a Go-based web application for Tibia character auction monitoring. The application scrapes auction data from Tibia's official website, provides currency conversion, and displays character information via a web interface.

## Build & Development Commands

### Essential Commands
```bash
# Start development database
make dev

# Stop all containers
make stop

# Create database migration
make create-migration name=<migration_name>

# Run database migrations
make migrate

# Rollback database migration
make rollback

# Start production environment
make run

# Clean all data and containers
make clean
```

### Go Development Commands
```bash
# Build main application
go build ./cmd/tibia-char

# Build cron jobs
go build ./cmd/cron/scrapper
go build ./cmd/cron/currency
go build ./cmd/cron/scheduler

# Generate templates (Templ)
templ generate

# Run from app directory
cd app && go run ./cmd/tibia-char/main.go
```

### Testing Commands
Currently no test files exist in the codebase. When adding tests:
```bash
# Run all tests (when available)
go test ./...

# Run specific test package
go test ./internal/auction

# Run with coverage
go test -cover ./...

# Run single test file
go test -run TestSpecificFunction ./internal/auction
```

## Project Structure

```
app/
├── cmd/                    # Entry points
│   ├── tibia-char/        # Main web server
│   └── cron/              # Scheduled jobs
│       ├── scrapper/      # Data scraping
│       ├── currency/      # Currency updates
│       └── scheduler/     # Auction refresh
├── internal/              # Application logic
│   ├── auction/           # Core auction logic
│   ├── currency/          # Currency conversion
│   ├── health/            # Health endpoints
│   └── server/            # HTTP server setup
├── pkg/                   # Shared utilities
│   ├── constants/         # Application constants
│   ├── enums/            # Type enumerations
│   ├── helper/           # Utility functions
│   ├── middleware/       # HTTP middleware
│   └── vendor/           # Third-party integrations
├── database/             # Database setup and migrations
└── assets/              # Static assets (CSS, etc.)
```

## Code Style Guidelines

### Imports
- Group imports: standard library, third-party, internal packages
- Use absolute imports with `github.com/adriein/tibia-char/`
- Order alphabetically within each group

```go
import (
    "context"
    "fmt"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/rotisserie/eris"

    "github.com/adriein/tibia-char/internal/auction"
    "github.com/adriein/tibia-char/pkg/constants"
)
```

### Naming Conventions
- **Constants**: `CamelCase` in constants package, `UPPER_SNAKE_CASE` for env vars
- **Variables**: `camelCase` for local variables, `PascalCase` for exported
- **Functions**: `PascalCase` for exported, `camelCase` for unexported
- **Structs**: `PascalCase`, use `NewStructName()` constructors
- **Interfaces**: `InterfaceName` suffix, or descriptive name ending in `er` (e.g., `Repository`)

### Error Handling
- Use `github.com/rotisserie/eris` for structured error wrapping
- Always wrap errors with context using `eris.Wrap(err, "context")`
- Use `eris.Errorf()` for creating new errors with formatting
- Log errors with trace IDs when available
- Return errors from functions, don't panic unless unrecoverable

```go
if err != nil {
    return nil, eris.Wrap(err, "Failed to save auction")
}

if err := repository.Save(auction); err != nil {
    return eris.Wrapf(err, "Error saving auction with ID %d", auction.ID)
}
```

### Logging
- Use `log.New(os.Stderr, "[Component] ", log.LstdFlags|log.LUTC)` for component loggers
- Include trace IDs in log messages for request tracking
- Log important operations and errors, not every step
- Use UTC timestamps for consistency

### Types & Interfaces
- Define DTOs for external data transfer
- Use interfaces for repositories and services
- Create enums in `pkg/enums/` for fixed value sets
- Use pointer types for optional fields in structs

### Database Patterns
- Use `sql.DB` for database connections
- Create repository pattern for data access
- Use transactions for multi-statement operations
- migrations in `database/migrations/` with up/down files

### HTTP/API Patterns
- Use Gin framework for HTTP routing
- Implement middleware for tracing, error handling, timezone
- Return structured responses with consistent format
- Use context for request-scoped values (trace ID, timezone)

### Concurrency
- Use `errgroup.Group` for coordinated goroutines
- Implement semaphores for concurrency limiting
- Use `sync.RWMutex` for thread-safe data structures
- Pass context through goroutine chains

### Configuration
- Environment variables defined in `pkg/constants/`
- Use `.env` files for development (not committed)
- Validate required environment variables on startup
- Separate production/development configurations

## Development Workflow

1. **Environment Setup**: Run `make dev` to start database
2. **Database Setup**: Run `make migrate` to apply schema
3. **Template Generation**: Run `templ generate` after `.templ` changes
4. **Testing**: Add tests as needed, use `go test ./...`
5. **Code Review**: Follow style guidelines above
6. **Build**: Ensure `go build` succeeds for all entry points

## Key Libraries & Dependencies

- **Web Framework**: `github.com/gin-gonic/gin`
- **Database**: `github.com/lib/pq` (PostgreSQL)
- **Templating**: `github.com/a-h/templ`
- **Error Handling**: `github.com/rotisserie/eris`
- **Web Scraping**: `github.com/gocolly/colly/v2`
- **Environment**: `github.com/joho/godotenv`

## Notes for Agents

- Application uses PostgreSQL database via Docker Compose
- No existing test suite - tests need to be added
- Templates use `templ` library requiring generation step
- Multiple entry points: web server and cron jobs
- Heavy use of external APIs and web scraping
- Currency conversion functionality integrated
- Error handling and tracing are important throughout