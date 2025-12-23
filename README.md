# Darts Web

[![Latest Release](https://img.shields.io/github/v/release/TestAndWin/darts-web)](https://github.com/TestAndWin/darts-web/releases)

A modern web-based darts scoring system with real-time game tracking and player statistics.

## Features

- Real-time game tracking for 301 and 501 games
- Player management and performance statistics
- Best-of-1, 3, or 5 sets support
- Clean, responsive UI built with React
- REST API backend

**Tech Stack:** Go + SQLite backend, React + Vite frontend

## Development Mode

### Install Dependencies

```bash
make deps
```

### Run the Application

You need to run both backend and frontend simultaneously in separate terminals:

**Terminal 1 - Backend:**
```bash
make run-backend
```
Server starts on `http://localhost:8080`

**Terminal 2 - Frontend:**
```bash
make run-frontend
```
Frontend starts on `http://localhost:5173`

## Docker Export

### Build and Export Image

```bash
make export
```

This command:
- Builds a multi-architecture Docker image (linux/amd64, linux/arm64)
- Tags it with the version from the latest git tag
- Exports it as `darts-app-{VERSION}.tar`

### Versioning

Versions are managed automatically via [semantic-release](https://github.com/semantic-release/semantic-release).

**For Developers:**
- Push to `main` branch
- Use semantic commit messages (recommended): `feat:`, `fix:`, etc.
- Releases are created automatically

**For Deployment:**
```bash
# Deploy latest version
./deploy.sh

# Deploy specific version
./deploy.sh v1.0.18
```

See [CHANGELOG.md](./CHANGELOG.md) for release history and [CONTRIBUTING.md](./CONTRIBUTING.md) for commit conventions.