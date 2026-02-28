# Future Workflows: MCP Orchestration
Copyright (c) 2026 Michael Lechner. All rights reserved.

This document outlines the vision for **wollmilchsau** as a programmable middleware for AI agents.

## The Concept: "MCP-Fetch"

The core idea is to transform **wollmilchsau** from an isolated sandbox into an orchestrator that can communicate with other MCP servers. This enables complex data processing workflows directly in TypeScript/JavaScript.

### 1. Server Registration
At startup, **wollmilchsau** loads a configuration (e.g., `mcp_registry.json`) defining known MCP servers:

```json
{
  "postgres": { "type": "sse", "url": "http://db-server:8080/sse" },
  "weather": { "type": "stdio", "command": "weather-mcp-server" }
}
```

### 2. The `mcp` Object in the Sandbox
A global `mcp` object is injected into the V8 sandbox, providing a fetch-like interface:

```javascript
// Example: Statistical analysis of database records
const rawData = await mcp.call("postgres", "query", { 
    sql: "SELECT age FROM users WHERE active = true" 
});

// Processing directly inside wollmilchsau
const ages = rawData.map(r => r.age);
const avgAge = ages.reduce((a, b) => a + b, 0) / ages.length;

console.log(`Average age of active users: ${avgAge}`);
```

## Use Cases

### Statistical Analysis
Instead of having the LLM read hundreds of rows of raw data (exhausting the context window), it writes a short script for **wollmilchsau**. The script fetches data from a Database MCP, calculates mean, variance, or trends, and returns only the compact result.

### Complex Data Transformations
Fetch data from an API MCP server, merge it with data from a File MCP server, and format the result into a structured report.

### Programmable Integration Testing
A script that calls multiple tools from different servers sequentially and verifies the consistency of the results.

## Graphics Generation & Visualization

Another exciting expansion is the combination of **wollmilchsau** with specialized graphics MCP servers (e.g., for D2, Mermaid, Gnuplot, or SVG generation).

**Workflow:**
1.  A script in **wollmilchsau** calculates complex data (e.g., Fibonacci sequences or statistical distributions).
2.  The script uses `mcp.call` to invoke a graphics server, passing the calculated data.
3.  The graphics server returns an image (SVG/PNG), which **wollmilchsau** presents to the user.

This enables "programmable graphics," where the logic resides in TypeScript and the rendering is handled by specialized tools.

## Technical Roadmap
1.  **Bridge Logic:** Go-side implementation of an MCP client supporting both SSE and Stdio.
2.  **V8 Injection:** Mapping Go client functions to JavaScript callbacks using `v8go.FunctionTemplate`.
3.  **Async/Await Support:** Proper Promise handling in the sandbox via V8 Taskrunner integration.

---
*This vision positions **wollmilchsau** as the central hub for intelligent, data-driven agent workflows.*
