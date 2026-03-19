# Contributing to Onyx

All contributions are welcome — code, documentation, bug reports, or plugin ideas.

## Getting started

```bash
git clone https://github.com/Elchi-dev/onyx.git
cd onyx

# Install all dev tools (golangci-lint, air, nfpm, goimports, git hooks)
bash scripts/dev/dev-setup.sh

# Build
make build

# Test
make test
```

## Day-to-day workflow

Use the dev helper scripts (they live in `scripts/dev/` and are gitignored):

```bash
# Push with quality checks (fmt → vet → test → lint → commit → push)
bash scripts/dev/push.sh

# Cut a release
bash scripts/dev/release.sh v0.2.0
```

## Branching model

- `main` — always releasable
- `feat/<name>` — new features
- `fix/<name>` — bug fixes
- `docs/<name>` — documentation only

## Commit style

```
feat(dashboard): add route management UI
fix(proxy): handle nil event handler gracefully
refactor(middleware): extract statusRecorder to shared file
docs: update plugin API guide
chore: bump gorilla/websocket to v1.5.3
```

## Code conventions

- Every exported symbol must have a doc comment.
- Use `fmt.Errorf("context: %w", err)` for error wrapping — always.
- No global state — pass dependencies through constructors.
- Keep packages small. `internal/app` is the only package allowed to import
  multiple internal packages.
- New middleware belongs in `internal/middleware/`.
- New database queries belong in `internal/database/`.

## Writing a plugin

See [docs/plugins.md](docs/plugins.md) and copy `plugins/example/` as a starting point.

The `plugin.Plugin` interface is stable and will not change before v1.0.

## Testing

Tests live alongside the code in `_test.go` files.
Run: `make test` (includes race detection).

## Reporting issues

Use the GitHub issue templates. Include your Onyx version (`onyx --version`) and OS.
