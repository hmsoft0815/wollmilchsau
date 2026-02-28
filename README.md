# wollmilchsau (Go + V8 + esbuild)

MCP Server in Go — High-performance TypeScript execution with embedded V8 and esbuild.

Copyright (c) 2026 Michael Lechner. All rights reserved.
Licensed under the MIT License.

> 🇩🇪 [Deutsche Version](README.de.md)

---

## Why Model Context Protocol (MCP)?

AI agents often need to execute code or process data to fulfill complex tasks. While LLMs are good at writing code, they cannot safely execute it themselves. 

**wollmilchsau** acts as a "sandboxed brain extension":
- **Safety**: Code runs in an isolated V8 environment without network or filesystem access.
- **Speed**: In-process bundling (esbuild) and V8 execution mean zero Node.js overhead.
- **Self-Correction**: Returns structured errors and source maps so agents can fix their own bugs.

---

## Features

- **In-Process Bundling:** Uses `esbuild` directly in Go (no Node.js subprocess required).
- **Isolated Execution:** Runs code in fresh V8 isolates for safety and performance.
- **Source Map Support:** Runtime errors and build warnings are automatically mapped back to original TypeScript files and lines.
- **LLM-Optimized Output:** Returns structured JSON metadata with human-readable summaries and separate content blocks for stdout/stderr.
- **SSE & Stdio Support:** Runs as a local process or a standalone HTTP server.
- **Artifact Integration:** Automatically connects to **mlcartifact** to persist execution results, large data blocks, or generated reports.
- **Request Archiving (ZIP Logging):** Optional full archiving of every request (source files + metadata + result) in compact ZIP files.

## Stack

| Component | Library | Purpose |
|---|---|---|
| MCP Protocol | `mark3labs/mcp-go` | JSON-RPC 2.0 Implementation |
| TS Bundling | `evanw/esbuild` | Fast, in-process transpilation |
| JS Execution | `rogchap/v8go` | CGo Bindings to V8 |
| Source Maps | Custom | VLQ decoding and position resolution |
| Logging | `log/slog` | Structured production logging |

## Getting Started

### Prerequisites

- **Go 1.23+**
- **C++ Compiler:** `build-essential` (Linux) or `llvm` (macOS) required for `v8go` CGo bindings.

### Pre-built Binaries (Linux)

**The easiest way:** Download the latest Linux binaries from the **[GitHub Releases](https://github.com/hmsoft0815/wollmilchsau/releases)** page. 

> [!NOTE]
> Due to the V8 dependency (CGo), we currently only provide automated binaries for Linux. For Windows and macOS, please follow the [Build](#build) section or use Go 1.24+.

### Build

```bash
make build
# Output: build/wollmilchsau
```

### Running

The server supports two transport modes:

1. **stdio (default):** Ideal for local use with Claude Desktop.
   ```bash
   ./build/wollmilchsau
   ```
2. **SSE (HTTP):** Standalone server for remote access.
   ```bash
   ./build/wollmilchsau -addr :8080
   ```

### CLI Flags

- `-addr string`: Listen address for SSE (e.g. `:8080`). If empty, uses stdio.
- `-log-dir string`: Path to a directory where every request will be archived as a ZIP file.
- `-version`: Show version information (wollmilchsau, V8, esbuild).
- `-dump`: Dump the full MCP tool schema as JSON.

## Advanced Request Logging

When `-log-dir` is specified, wollmilchsau creates a ZIP archive for every incoming tool call. This is ideal for auditing and debugging LLM behavior without bloating your primary log files.

Each ZIP file contains:
- `info.json`: Metadata (Timestamp, Client IP, Tool Name, Execution Plan).
- `src/`: All virtual source files provided by the LLM.
- `response.json`: The complete JSON result returned by the executor.

## MCP Best Practices

**wollmilchsau** is built as a reference implementation for high-performance, safe MCP servers:

- **Capability Signaling:** Explicitly declares tool support.
- **Structured Tool Results:** Instead of returning raw error strings, the server returns structured JSON metadata including `summary`, `exitCode`, and `diagnostics`. This allows LLMs to "understand" and fix their own code.
- **Diagnostic Source Mapping:** Uses V8 source maps to point errors back to the *original* TypeScript lines, making it easier for agents to debug multi-file projects.
- **Dual-Transport Support:**
  - **Stdio:** Optimized for local integration (e.g., Claude Desktop).
  - **SSE:** Optimized for remote or distributed AI workflows.
- **Privacy-First Logging:** Optional full-request archiving (`-log-dir`) for debugging without exposing secrets in regular application logs.

## Tools

### `execute_script`
Executes a single code snippet. Ideal for math and logic tests.
- `code`: The TypeScript/JavaScript code.
- `timeoutMs`: (Optional) Max execution time (default 10s).

### `execute_project`
Executes a multi-file project. Ideal for complex data processing.
- `files`: Array of `{name, content}` objects.
- `entryPoint`: The main file to run (e.g., `main.ts`).
- `timeoutMs`: (Optional) Max execution time.

### `check_syntax`
Validates TypeScript/JavaScript syntax without executing it. Returns a boolean success status and detailed diagnostics if it fails.
- `code`: The code snippet to check.
## 🔒 Execution Environment Constraints

To ensure safety and performance, the execution environment is strictly sandboxed:

- **Resource Limits:**
  - **Memory:** Active heap monitoring. Scripts are limited to **128MB** of used heap. Exceeding this triggers immediate termination.
  - **CPU / Time:** Configurable timeout (default 10s). Scripts are forcefully terminated using `iso.TerminateExecution()` if they exceed the limit.
- **ECMA-262 Only:** Pure V8 sandbox. Modern JavaScript/TypeScript features are supported, but environment-specific APIs are restricted.
- **No Network:** `fetch`, `XMLHttpRequest`, or any other form of network access is disabled.
- **No Event Loop Timers:** `setTimeout`, `setInterval`, and `setImmediate` are not available. Execution is strictly synchronous.
- **No Node.js / Web APIs:** No access to `fs`, `os`, `process`, or DOM APIs.
- **Limited i18n:** The `Intl` object is available but limited to the `en-US` locale.
- **Pure Logic:** Ideal for algorithms, data transformation, and mathematical computations.

### Installation via Script (Linux/macOS)

The fastest way to install **wollmilchsau**:

```bash
curl -sfL https://raw.githubusercontent.com/hmsoft0815/wollmilchsau/main/scripts/install.sh | sh
```

Download the latest version as a **ZIP/TAR**, or install via **.deb** or **.rpm** from the **[GitHub Releases](https://github.com/hmsoft0815/wollmilchsau/releases)** page.

### Docker Support

You can also run **wollmilchsau** as a Docker container. This is recommended if you want a fully isolated environment.

**Build the image:**
```bash
docker build -t wollmilchsau .
```

**Run the container:**
```bash
docker run -p 8000:8000 wollmilchsau
```

---

## Claude Desktop Integration

To use **wollmilchsau** as a tool in Claude Desktop, add it to your configuration file:

- **MacOS**: `~/Library/Application Support/Claude/claude_desktop_config.json`
- **Windows**: `%APPDATA%\Claude\claude_desktop_config.json`

```json
{
  "mcpServers": {
    "wollmilchsau": {
      "command": "wollmilchsau",
      "args": ["-log-dir", "/your/absolute/path/to/logs"]
    }
  }
}
```
*Note: If the binary is not in your PATH, provide the absolute path.*

---

---

## IMPORTANT Artifact Integration

**wollmilchsau** is deeply integrated with the [mlcartifact](https://github.com/hmsoft0815/mlcartifact) system. When configured, it can automatically save large execution results, charts, or complex data structures as persistent artifacts.

**How it works:**
1. Wollmilchsau executes your TypeScript/JavaScript code.
2. If the code generates an "Artifact" (via internal helpers), it is securely stored in the **artifact-server**.
3. Reached the LLM as an artifact ID, which can then be presented to the user.

> [!TIP]
> **Best Practice:** Run the `artifact-server` alongside `wollmilchsau` for the full experience. Start the server with `artifact-server -grpc-addr :9590`.

---
## 🚀 Future Roadmap: MCP Orchestration

We envision **wollmilchsau** as a central orchestrator for other MCP servers. By providing a fetch-like interface within the sandbox, scripts will be able to query other servers (e.g., databases) and process the data locally.

Read more about this vision in [FUTURE_WORKFLOWS.md](docs/FUTURE_WORKFLOWS.md).

## 📜 License & Ethical Use

This project is licensed under the **MIT License**. 

### 🕊️ Author's Note (Non-binding)
While the license allows broad usage, I (the author) kindly request that this software **not** be used for:
* **Military purposes** or the production and development of weapons.
* Activities by entities or individuals supporting the **military aggression against Ukraine**.

Furthermore, due to past professional experiences, I explicitly and kindly request that my former contractor, **Isensix, Inc.**, and its acquirer, **Dwyer-Omega**, do not use this software in any way.

*This request is an appeal to professional ethics and personal conscience and does not constitute a legal modification of the MIT License.*
