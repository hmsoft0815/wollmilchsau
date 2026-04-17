# 🏗️ Technical Architecture: wollmilchsau

Internal architecture and implementation details for wollmilchsau developers.

## Architecture Flow

```
┌─────────────┐     MCP JSON-RPC      ┌─────────────────┐
│  LLM Agent  │ ───────────────────►  │   wollmilchsau  │
└─────────────┘                       └────────┬────────┘
                                               │
                               ┌─────────────────▼─────────────────┐
                               │                                   │
                               │  ┌───────────┐  ┌─────────────┐  │
                               │  │   MCP     │  │    Server   │  │
                               │  │  Handler  │  │   (stdio)   │  │
                               │  └──────┬────┘  └─────────────┘  │
                               │         │                         │
                               │         ▼                         │
                               │  ┌───────────────┐                │
                               │  │  Request      │                │
                               │  │  Parser       │                │
                               │  └───────┬───────┘                │
                               │         │                         │
                    ┌──────────┘         ▼         ┌─────────────┐
                    │   ┌─────────────────┐         │  MCP        │
                    │   ┃  TypeScript/JS  │◄────────│  Tools      │
                    │   │  (Code/Files)   │         │  Registry   │
                    │   └────────┬────────┘         └─────────────┘
                    │            │
                    │            ▼
                    │   ┌─────────────────┐
                    │   │   Bundler       │
                    │   │  (esbuild)      │─────────►(In-Process)
                    │   └────────┬────────┘
                    │            │
                    │            ▼
                    │   ┌─────────────────┐
                    │   │   Executor      │
                    │   │  (v8go/V8)      │─────────►(Sandbox)
                    │   └────────┬────────┘
                    │            │
                    │  ┌─────────▼─────────┐
                    │  │  Source Map       │
                    │  │  Resolver         │
                    │  └───────────────────┘
                    │
                    ▼
            ┌───────────────┐
            │  Response     │
            │  (Structured) │
            └─────┬─────────┘
                  │
            ┌─────▼─────────┐
            │  LLM Agent    │
            └───────────────┘
```

## Request Flow

1. **Incoming Request**
   - MCP-Client sends `request/callTool` with `execute_script` or `execute_project`.
   - Payload includes `code`, `files`, `entryPoint`, `timeoutMs`.

2. **Parsing & Validation**
   - `internal/parser/` parses parameters.
   - Validates structure and calculates timeout (Default: 10s).

3. **Bundling**
   - `internal/bundler/` uses esbuild (in-process) to transpile TS -> JS.
   - Modules in `files` are bundled (for multi-file projects).
   - Generates Source Maps for error mapping.

4. **Execution**
   - `internal/executor/` creates a V8 Isolate with strict sandboxing:
     - Disk/Network/Timers: **Disabled**.
     - Node.js Globals: **Disabled**.

5. **Error Handling**
   - JS Errors are caught and their stack trace is parsed.
   - `internal/sourcemap/` maps V8 positions back to original TypeScript lines.

6. **Response**
   - Structured JSON with `result` or `error` (with source context).
   - Artifact links if applicable.

## Sandbox Constraints

Isolation is enforced via `v8go` (CGO bindings to V8):

- **Heap Limit:** 128MB.
- **CPU Limit:** Forced via V8 TaskRunner timeouts.
- **Strictly Logic:** No access to `require()`, `process`, `fs`, `os`, `fetch`.

## Source Maps Implementation

- **Format:** VLQ-encoded.
- **Resolver:** `internal/sourcemap/sourcemap.go`.
- **Functionality:** Parses V8 stack traces and maps positions to original TS snippets for the AI agent to self-correct.

## Artifact Integration

When enabled (`-enable-artifacts`):

- `wollmilchsau.openArtifact(filename, mimeType)` is injected into the sandbox.
- Communicates via gRPC with an `mlcartifact` server.
- Tool response automatically includes an MCP `resource_link`.

## Request Logging

Optional via `-log-dir`:

- Every request is archived as a ZIP file.
- Contains: Payload, Response, TS Code, JS Bundle, Stacktrace.
- Useful for audit trails and debugging failed agent interactions.
