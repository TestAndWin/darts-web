# Contributing to Darts Web

## Commit Message Guidelines

We use semantic commit messages for automatic versioning. These are **recommended but not enforced**.

### Format

```
<type>(<scope>): <subject>

<body>

<footer>
```

### Types

- **feat:** New feature (→ Minor version bump: 1.0.0 → 1.1.0)
- **fix:** Bug fix (→ Patch version bump: 1.0.0 → 1.0.1)
- **perf:** Performance improvement (→ Patch version bump)
- **refactor:** Code refactoring without functional changes (→ Patch version bump)
- **docs:** Documentation only (→ no release)
- **style:** Code formatting (→ no release)
- **test:** Adding or modifying tests (→ no release)
- **chore:** Build, dependencies, tooling (→ no release)
- **ci:** CI/CD configuration (→ no release)

### Examples

```bash
# Feature (minor bump)
git commit -m "feat: add CSV export for player statistics"

# Bug fix (patch bump)
git commit -m "fix: correct bust throw calculation"

# With scope
git commit -m "feat(ui): add dark mode toggle"

# Breaking change (major bump)
git commit -m "feat!: redesign API structure

BREAKING CHANGE: API endpoints moved from /api/* to /v2/api/*"

# No release
git commit -m "docs: update deployment instructions"
git commit -m "chore: update dependencies"
```

### What happens?

- Commits with `feat:`, `fix:`, `perf:`, `refactor:` trigger automatic releases
- Other commits (docs, chore, test) don't trigger releases
- Non-semantic commits still work, they just won't trigger releases

## Development Workflow

1. Create feature branch: `git checkout -b feature/my-feature`
2. Commit changes (semantic messages recommended)
3. Push and create pull request
4. After merge to main: semantic-release automatically creates version

## Questions?

See [CLAUDE.md](./CLAUDE.md) for detailed development documentation.
