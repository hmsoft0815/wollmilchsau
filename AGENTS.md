<!-- mlc-dochub:begin — auto-managed, do not edit between these markers -->
## MLC Doc Hub — wollmilchsau

Structured documentation lives in `.mlcai/`, maintained through the
`mlc-dochub` MCP server.

- **Project ID:** `wollmilchsau` — wollmilchsau — LLM Script Execution Sandbox
- **Source directory:** `/mnt/data2tb/wollmilchsau` (read source files with native Read —
  the MCP tools only touch `.mlcai/`)

### Existing docs

Read any of these directly (native Read is fine) to gather context before you work:

- `.mlcai/INTEGRATION.md` — How this project fits into the larger system
- `.mlcai/TECH_STACK.md` — Stack & dependencies
- `.mlcai/API_CONTRACT.md` — API contract / endpoints
- `.mlcai/DECISION_LOG.md` — Architecture decisions
- `.mlcai/DETAIL_DOCS.md` — Project detail docs (links to docs/)
- `.mlcai/USER_DOCS.md` — End-user perspective
- `.mlcai/BACKLOG.md` — Bugs / ideas / tasks
- `.mlcai/PRODUCT.md` — Product marketing page pointer (submodule meta)
- `.mlcai/INFRASTRUCTURE.md` — Infrastructure & deployment

### Product marketing page (separate concern)

Customer-facing marketing copy is **not** project documentation. It lives in
`./mlcprodweb/`, backed by `mlc@nas.local:/volume1/homes/mlc/repositories/product/wollmilchsau.git`. Rendered live at https://mlcgo.eu/products/wollmilchsau/.

### Working with `.mlcai/`

**Never write a `.mlcai/` file with a native editor.** Every create / update /
delete goes through the `mlc-dochub` MCP tools — they stamp the `## 📋 Meta`
footer, append to the activity log and guard against concurrent edits. Reading
with a native Read is fine and usually cheaper.

**If the tools are not available to you, read but do not write** — and point the
user at `task install-all` in the mlcintegration checkout (https://github.com/mlc911/mlcintegration).

The server states its full operating rules on connect (`author=`, `base_modified`,
which doc serves which purpose). Clients that drop server-level instructions —
Antigravity does, verified 29.08.2026 — get the same rules from the global
`~/.gemini/GEMINI.md`, section 6.
<!-- mlc-dochub:end -->

## 🛠️ Development Commands

### Build & Run

```bash
# Quick build (Taskfile - preferred)
task build                          # → build/wollmilchsau
task tidy                           # go mod tidy

# Or via Makefile
make build                          # → build/wollmilchsau
make deps                           # download + tidy
make check                          # fmt + vet + lint + test
```

Binary lands in `build/wollmilchsau`. CGO is mandatory (v8go dependency) — set `CGO_ENABLED=1` explicitly when cross-compiling.

### Test

```bash
# All unit tests
go test ./... -v -count=1           # via Makefile: make test
make test-race                      # with Go race detector

# MCP script integration tests (requires CLI tool)
make install-tester                 # go install mcp-tester
make test-mcp                       # runs basic.mcp + extended.mcp
```

MCP test scripts live in `tests/` (`basic.mcp`, `extended.mcp`).

### Lint

```bash
make lint                           # requires golangci-lint
golangci-lint run --config .golangci.yml ./...
```

`.golangci.yml` — enables errcheck, govet (all except fieldalignment), ineffassign, staticcheck, misspell. `errcheck` is silenced for defer/remove paths and in testclient/artifacts_test.

### Docker

```bash
make docker-build                   # via Taskfile
docker build -t wollmilchsau .       # via Dockerfile (multi-stage)
# Run:  docker run -p 8000:8000 wollmilchsau   (SSE mode on :8000)
```

Dockerfile uses `golang:1.24-bookworm` builder + `debian:bookworm-slim` runtime. Requires `gcc/g++/make` in build stage, `ca-certificates/libc6/libstdc++6` at runtime for v8go's native libraries.

## 📐 Architecture Overview

**wollmilchsau** is an MCP server that lets LLM agents execute TypeScript/JavaScript in a sandboxed V8 isolate. The agent writes code → wollmilchsau compiles it → runs it → returns structured results or source-mapped errors.

### Entry Point & Modes

`cmd/main.go:18` sets up CLI flags, then starts either **stdio mode** (default, for Claude Desktop etc.) or **SSE/HTTP mode** (`-addr :8080`). No external DB or service dependency — everything is in-process.

### Key Internals (`internal/`)

| Package | Responsibility |
|---------|---------------|
| `server/` | MCP tool registration, handler functions. `tools.go` defines tool schemas (MCP input/output). `server.go:27` `New()` wires all tools. |
| `parser/` | Parses and validates MCP request parameters (`code`, `files`, `entryPoint`, `timeoutMs`). `types.go` has shared structs. |
| `bundler/` | Bundles TypeScript → JavaScript via **in-process** esbuild (`github.com/evanw/esbuild`). Handles multi-file projects, generates source maps. |
| `executor/` | The core: creates a V8 isolate per request, injects polyfills and console hooks, runs code with 128MB heap / configurable CPU limits via watchdog goroutine. |
| `sourcemap/` | Custom VLQ decoder that resolves generated JS positions back to original TS line/column for error reporting. |
| `requestlog/` | Optional ZIP archive of every request/response (enabled via `-log-dir`). |

### Data Flow (single-file execution)

```
MCP callTool → server.handleExecuteScript → parser.ParseParameters
  → bundler.Bundler.Bundle() → executor.Execute() → structured Result
```

For multi-file: `handleExecuteProject` adds a bundling step before executor.

Error stack traces from V8 (`v8.JSError`) are parsed by location string, then resolved through the source map in `executor.go:160` for developer-friendly line numbers.

### Bundled JS Packages (npm-Pakete für die Sandbox)

wollmilchsau kann Node.js-Pakete in die V8-Sandbox injizieren, damit LLM-Agenten sie via `require()` oder ES-Module-Imports nutzen können.

**Installation nur beim Serverstart — keine runtime-Installation möglich:**
Die Pakete werden einmalig beim Serverstart per `npm install` heruntergeladen und dann als VirtualFiles (via esbuild) in jedes Ausführungskontext injiziert. **Während der Sandbox-Ausführung ist kein Internetzugang, keine Paketverwaltung und keine nachträgliche Installation möglich.**

- **Verfügbare Pakete abfragen:** Agenten können das **list_js_packages** MCP-Tool im Sandbox-Kontext nutzen, um eine Liste aller verfügbaren Pakete (Name, Version, Typ) zu erhalten.
- **Default-Pakete:** `crypto-js`, `lodash` (+ `@types/lodash`), `mathjs`, `zod`
- **Eigene Pakete hinzufügen:** `-bundled-js-deps=crypto-js,lodash,mathjs,zod`
- **Achtung bei Scoped-Packages** (z.B. `@types/*`): Werden korrekt als Ordnerstruktur in `node_modules/` installiert und von esbuild gefunden.
- **CJS vs ESM:** Default-Ausgabe ist CommonJS. ES-Module-Imports benötigen ggf. esbuild-Transpilation für `require()` → import Konvertierung.

### Artifact Integration (optional)

When `-enable-artifacts -artifact-addr localhost:50051`:
- `mlcartifact.Client` connects via gRPC to the `mlcartifact` service
- `wollmilchsau.openArtifact()` and low-level `artifact.*` API are injected into the JS sandbox
- Tool responses include MCP resource links for downloaded artifacts

## ⚠️ Gotchas & Constraints

1. **CGO is mandatory.** `v8go` needs CGO_ENABLED=1. Bare Go cross-compilation (GOOS/X without CGO) will fail. Use Docker or build natively per platform. macOS universal binaries require `lipo`. Windows needs `mingw-w64`.
2. **Sandbox is strict-nothing.** No network (fetch/XMLHttpRequest), no filesystem, no timers (setTimeout/setInterval disabled), no Node.js globals. If the agent's code uses any of these, it will silently fail or throw.
3. **Heap limit is 128MB.** The watchdog in `executor.go:108` monitors and calls `iso.TerminateExecution()` when exceeded.
4. **Binary name "wollmilchsau" means "holy cow" (German).** Not a coding standard quirk — just an inside joke.
5. **Artifacts test is broken.** `internal/executor/artifacts_test.go` has 4 compiler errors — `pb.FindRequest`, `pb.PatchRequest`, `pb.PatchResponse` are undefined. Update against the current protobuf definitions when fixing.
6. **Module path** is `github.com/hmsoft0815/wollmilchsau`. Internal imports use this base path (e.g., `github.com/hmsoft0815/wollmilchsau/internal/server`).
7. **npm packages are pre-installed only.** No `npm install` or internet access in the sandbox. If a package isn't listed by `list_js_packages`, it will not be available at runtime. Default bundled packages: `crypto-js` (cryptographic functions), `lodash` (utilities), `@types/lodash` (TS types), `mathjs` (math engine), `zod` (schema validation).

## 📝 Code Style

- Package comments at the top of each file (`// Package server wires the MCP tools...`)
- Copyright header on every source file: `// Copyright (c) 2026 Michael Lechner. All rights reserved.`
- Error handling via return values, not panics — V8 errors are wrapped in `*v8.JSError` type assertion
- Uses Go's `log/slog` for logging (not zap/zerolog)
