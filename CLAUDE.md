# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Tech Stack

- **Backend**: Go 1.25+ with SQLite (go-sqlite3), standard library HTTP server
- **Frontend**: React 19 with Vite, Tailwind CSS
- **Deployment**: Docker with multi-stage builds, Kubernetes/Helm charts

## Development Commands

**IMPORTANT**: This project includes a Makefile for common tasks. Prefer using `make` commands when available.

### Quick Commands (Makefile - Preferred)

```bash
# Show all available commands
make help

# Install dependencies
make deps

# Run tests
make test
make test-verbose
make test-coverage

# Run backend server (http://localhost:8080)
make run-backend

# Run frontend dev server (http://localhost:5173)
make run-frontend

# Build for production
make build-backend
make build-frontend

# Build and export Docker image (uses VERSION file)
make build
make export

# Clean build artifacts
make clean
```

### Backend (Manual)

```bash
# Run the backend server (default: http://localhost:8080)
go run cmd/server/main.go

# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests for specific package
go test ./internal/game
go test ./internal/store

# Build for production
go build -o darts-server cmd/server/main.go
```

### Frontend (Manual)

```bash
# Install dependencies (required after clone)
cd frontend && npm install

# Start dev server (http://localhost:5173)
npm run dev

# Build for production (outputs to dist/)
npm run build

# Lint code
npm run lint
```

### Docker

```bash
# Build Docker image (via Makefile - recommended)
make build

# Build and export as tar
make export

# Or manually
docker build -t darts-web .

# Run container
docker run -p 8080:8080 -v $(pwd)/data:/data darts-web
```

### Kubernetes/Helm

```bash
# Deploy with Helm
helm install darts-web ./charts/darts-web

# Deploy with raw K8s manifests
kubectl apply -f k8s/darts-app.yaml
```

### Versioning

The application version is managed via Git tags. To release a new version:

```bash
# Create and push a tag
git tag v1.0.15 -m "Release 1.0.15"
git push origin v1.0.15

# On the server, deploy the latest tag
./deploy.sh

# Or deploy a specific tag
./deploy.sh v1.0.15
```

### Automated Versioning with semantic-release

The project uses semantic-release for automatic versioning:

**Workflow:**
1. Push commits to `main` branch
2. GitHub Actions analyzes commits
3. For release-worthy commits (feat:, fix:, etc.):
   - Determines next version number
   - Creates git tag (e.g., `v1.0.18`)
   - Generates CHANGELOG.md
   - Creates GitHub Release with release notes

**Commit Message Conventions (recommended):**

```bash
# Features (minor version bump)
git commit -m "feat: add player statistics export"

# Bug fixes (patch version bump)
git commit -m "fix: correct bust throw calculation"

# Breaking changes (major version bump)
git commit -m "feat!: redesign API endpoints

BREAKING CHANGE: API endpoints now use /v2/ prefix"

# No release
git commit -m "docs: update deployment instructions"
git commit -m "chore: update dependencies"
```

**Important:** Commits without semantic prefix still work, they just don't trigger automatic releases.

**View releases:**
- GitHub: https://github.com/TestAndWin/darts-web/releases
- CHANGELOG.md in repository root

**Deployment remains manual:**
1. Push commits to main → semantic-release creates tag automatically
2. On server: `./deploy.sh` (deploys latest tag)
3. Or specific version: `./deploy.sh v1.0.18`

**Note:** Manual tags still work for special cases.

## Architecture Overview

### Backend Layered Architecture

The backend follows a clean separation of concerns:

1. **cmd/server/main.go**: Entry point, sets up HTTP server with CORS middleware, handles graceful shutdown
2. **internal/handlers**: HTTP request handling layer with input validation and JSON responses
3. **internal/game/engine.go**: Core game logic engine - enforces darts rules, validates throws, manages turns, detects busts/checkouts
4. **internal/store**: Database layer with SQLite operations - uses transactions for consistency, parameterized queries for safety
5. **internal/models**: Shared data structures used across all layers

### Critical Backend Details

- **Thread Safety**: `Handler.getGameLock()` provides per-game mutex locking to prevent race conditions during concurrent throws
- **Game State Management**: The `Engine` is stateless; all state lives in the `Game` struct which is persisted in the store
- **Bust Handling**: When a throw goes negative or leaves exactly 1 point (impossible checkout), the entire turn is reverted to the turn-start score and the turn immediately ends
- **Set Wins**: Games use a best-of-sets system (1/3/5). Winning a set resets all players' current points but preserves sets won

### Environment Variables

| Variable | Default | Purpose |
|----------|---------|---------|
| `PORT` | `8080` | Server port |
| `DB_PATH` | `./darts.db` | SQLite database file path |
| `BASE_PATH` | `""` | URL prefix (e.g., `/darts` for reverse proxy) |
| `CORS_ORIGIN` | `*` | CORS allowed origin |
| `VITE_API_URL` | `http://localhost:8080/api` | Frontend API endpoint (frontend only) |

**Important**: When `BASE_PATH` is set, both API routes and static files are served under that prefix (e.g., `/darts/api/users`, `/darts/index.html`)

### Frontend Architecture

- **src/App.jsx**: Main component managing app state (users, current game) and routing between setup/game views
- **src/components/GameSetup.jsx**: Player selection and game configuration
- **src/components/ActiveGame.jsx**: Live game interface with throw input and score display
- **src/components/UserStats.jsx**: Player statistics display
- **src/components/ErrorBoundary.jsx**: React error boundary for graceful error handling
- **src/services/api.js**: API client functions for backend communication

**Frontend Build Config**: `vite.config.js` sets `base: '/darts/'` to match the backend's `BASE_PATH` for production deployment

### Game Rules Implementation

The `internal/game/engine.go` enforces:

- **Valid throws**: 0-20 points, or 25 (bull); multipliers 1x/2x/3x
- **Bull constraints**: Can be single (25) or double (50), but never triple
- **Bust conditions**: Score goes negative OR lands on exactly 1 point
- **Bust behavior**: Entire turn reverts to turn-start score, turn ends immediately
- **Checkout**: Player must reach exactly 0 to win the set
- **Turn limit**: 3 throws per turn (unless bust occurs earlier)

### Database Schema Key Points

- Users table: ID, Name, CreatedAt
- Games table: ID, Status (PENDING/ACTIVE/FINISHED), Settings (JSON), WinnerID, CreatedAt
- GamePlayers table: Links users to games with order, sets won, current points
- Throws table: Full throw history with validity flag and score after each throw

## Common Development Workflows

### Adding a New API Endpoint

1. Define handler method in `internal/handlers/handlers.go`
2. Register route in `cmd/server/main.go` with HTTP method prefix (e.g., `"GET /api/endpoint"`)
3. Add corresponding API client function in `frontend/src/services/api.js`
4. Call from React components

### Modifying Game Logic

1. Update rules in `internal/game/engine.go`
2. Add/modify tests in `internal/game/engine_test.go`
3. Run tests to verify: `go test ./internal/game`
4. If data model changes, update `internal/models/models.go` and database schema in `internal/store/db.go`

### Testing Changes

Always run backend tests before committing:
```bash
go test ./...
```

## Deployment Notes

- **Docker build**: Multi-stage build compiles Go binary with CGO enabled for SQLite, builds frontend, produces minimal Alpine image
- **Persistence**: Database should be mounted at `/data/darts.db` via volume (K8s PVC or Docker volume)
- **Production settings**: Set `CORS_ORIGIN` to specific domain (not `*`), ensure `BASE_PATH` matches ingress path
- **Helm deployment**: See `charts/darts-web/values.yaml` for configuration options including resource limits, ingress, and persistence settings
