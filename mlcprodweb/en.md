# Wollmilchsau

Wollmilchsau is a high-performance TypeScript execution engine designed specifically for Model Context Protocol (MCP) servers.

## What it does

Executing TypeScript logic within an MCP server usually requires a full Node.js or Deno runtime, which adds significant overhead and complexity. Wollmilchsau solves this by embedding the powerful V8 JavaScript engine directly into a Go binary. This allows it to run TypeScript tools with near-native performance, zero external dependencies, and a minimal memory footprint.

## Key features

- **V8 Engine Performance**: By using the same engine that powers Chrome and Node.js, Wollmilchsau ensures your MCP tools run as fast as possible.
- **Embedded esbuild**: Automatically bundles and transpiles your TypeScript files on the fly, so you don't need a complex build pipeline.
- **Seamless Go Integration**: Designed to be the scriptable heart of Go-based MCP servers, allowing you to define tool logic in TypeScript while maintaining the robustness of Go for the server transport.
- **Artifact Support**: Built-in connection to the `mlcartifact` service, allowing your scripts to produce and manage persistent results easily.
