# Agent Instructions

## Development Workflow

When working on features or fixes, follow this workflow:

### 1. Build & Test
```bash
go build ./...
go test ./...
```

### 2. Lint
```bash
golangci-lint run
```

### 3. Commit, Push & Tag

After completing a feature and all tests/lint pass:

```bash
# Stage and commit with descriptive message
git add <files>
git commit -m "feat: description of changes

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>"

# Push to remote
git push

# Tag new version
git tag vX.Y.Z
git push origin vX.Y.Z
```

### Version Tagging Guidelines

- **Minor version bump (v0.X.0)**: New features, enhancements, non-breaking changes
- **Patch version bump (v0.0.X)**: Bug fixes, small improvements
- **Major version bump (vX.0.0)**: Breaking changes (only when explicitly requested)

Default to **minor** version bumps unless:
- The change is a small bug fix (use patch)
- User explicitly requests a major version bump

### Commit Message Format

Use conventional commits:
- `feat:` - New features
- `fix:` - Bug fixes
- `refactor:` - Code refactoring
- `docs:` - Documentation changes
- `test:` - Test additions/changes
- `chore:` - Maintenance tasks

Always include the co-author line for AI-assisted commits.

## Code Style

- Follow Go idioms and best practices
- Keep changes focused and minimal
- Avoid over-engineering
- Don't add unnecessary abstractions
- Prefer editing existing files over creating new ones
