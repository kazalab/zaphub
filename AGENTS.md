# AGENTS.md — zaphub

## What it is

Go API: French TV EPG (from xmltvfr.fr), weekly cinema releases (from AlloCiné), YouTube RSS feeds. Data cached in memory, served as JSON. A responsive, dependency-free web UI is embedded in the binary and served at `/`.

## Project metadata

Repository metadata is centralized in `.meta.yml` at the repository root:

- `name`, `description`, `homepage`, `topics`, `license`

The description is intentionally factual and does not claim native Home Assistant integration. Run `mise run repo-bootstrap` to apply this metadata to the GitHub repository (description, homepage, topics).

## Prerequisites

[mise](https://mise.jdx.dev) manages tools (Go, Dagger, yq) and tasks. Go is required to run the daemon locally; Docker is optional and only needed if you want to run the published OCI image.

## Commands

All common commands run through `mise` tasks declared in `mise.toml`. The heavy lifting (tests, lint, build) is executed inside Dagger pipelines to keep the host environment clean and reproducible.

```sh
mise run bootstrap        # install tools and download Go dependencies
mise run build            # build artifacts with GoReleaser (output in ./dist)
mise run lint             # run Markdown, JSON and Go linters
mise run fmt              # format Go source files
mise run verify           # full validation before a commit or PR
mise run repo-bootstrap   # configure GitHub repository metadata from .meta.yml
```

Run from repo root. For the complete task list, see `mise.toml`.

To start the server locally during development:

```sh
go run ./cmd/zaphub
```

The published production image is available at `ghcr.io/kazalab/zaphub:latest`.

## Web UI

`GET /` serves a responsive web UI (vanilla HTML/CSS/JS, no framework) embedded into the binary. Source lives at the repository root in `web/` (`embed.go`, `index.html`, `style.css`, `app.js`, `status.html`), embedded via `//go:embed index.html app.js style.css status.html` in `web/embed.go`. The files must stay at `web/` root (the embed paths are relative to the package directory). `api.New` takes an `fs.FS` (nil disables the UI); the embedded FS is passed as-is (no `fs.Sub`, the files are at the embed root). There is no dev-served version — edit files in `web/` then rebuild/re-run. The UI is a 3-tab SPA (TV, YouTube, Cinéma) that calls the existing JSON API directly; the service status lives on a separate `status.html` page reached by clicking the status pill in the top bar. When changing the UI, keep it dependency-free.

## Project structure

```
cmd/zaphub/main.go       # thin entry point: Execute()
internal/
  api/                   # HTTP handlers, router, server
  cinema/                # cinema provider abstraction (AlloCiné): fetcher, parser, store, refresher
  cli/                   # cobra root command (daemon by default + version subcommand)
    root.go              # NewCommand / Execute
    daemon.go            # runDaemon: slog, config, errgroup (refresh runtime + server), graceful shutdown
    version.go           # appVersion/appCommit/appBuildDate (juli3nk/go-utils/version)
  config/                # env-based config (PORT, LOG_LEVEL, EPG_URL, etc.)
  epg/                   # XMLTV fetcher, parser, store, evening logic
  model/                 # shared types
  refresh/               # runtime orchestrating periodic refreshes
  youtube/               # RSS feed parser, store, refresher
web/                     # embedded web UI (html/css/js) + embed.go
```

CLI: running the binary starts the server by default; `zaphub version` prints the version. Flags override env config (e.g. `--log-level`). No submodules, no code generation, no migrations. Single `go.mod` at root.

## Testing quirks

- Tests run with `-race` always. No way to skip it via the task definitions.
- `internal/api/server_test.go` is an integration test hitting real HTTP endpoints; `newTestServer` returns `(*Server, *epg.Store, *cinema.Store, *youtube.Store)` and must be updated if `api.New`'s signature changes.
- `TestWebUI` verifies static file serving (including `status.html`) with an `fstest.MapFS` and that API routes aren't shadowed by the static handler.
- No test fixtures directory; test data is inline in test files.

## Linting & conventions

- Go 1.26, formatted with `gofumpt` through `mise run fmt`.
- Linters enabled in `.golangci.yml`: errcheck, govet, ineffassign, staticcheck, unused, gosec, revive, misspell.
- All discarded return values must be handled explicitly. For `Close()` calls, use `defer func() { _ = x.Close() }()`.
- `gosec G118` on the shutdown goroutine in `internal/api/router.go` is suppressed with `//nolint:gosec` because the goroutine intentionally uses `context.Background()` for graceful server shutdown.
- Dependencies: `spf13/cobra` (CLI), `juli3nk/go-utils` (version), `golang.org/x/sync` (errgroup), `golang.org/x/net` (HTML parsing for the AlloCiné provider).
- Structured logging via stdlib `log/slog` (JSON handler, level from `LOG_LEVEL` or `--log-level`).
- All data fetched over HTTP; no database.
- Environment variables for config (see `internal/config/config.go`).
- The published production image is built by GoReleaser/Ko on top of `ghcr.io/juli3nk/static:20260901` and pushed to `ghcr.io/kazalab/zaphub`.
- `.dockerignore` excludes `.git`, `data/`, and `*.md`.
