# 🛠 Tech Stack & Constraints: wollmilchsau

## Kern-Versionen
- **Sprache:** Go 1.24.2 (Backend), TypeScript/JavaScript (Sandbox)
- **Framework:** mark3labs/mcp-go (MCP)
- **Datenbank:** Keine lokale DB; optionaler `mlcartifact` (gRPC) Support

## Bibliotheken (Erlaubt/Fixiert)
- **Bundling:** evanw/esbuild (In-process Transpilation)
- **Execution:** rogchap/v8go (CGO Bindings zu V8)
- **Artifacts:** connectrpc.com/connect (gRPC Client)
- **Logging:** Structured Logging (slog)

## Einschränkungen (Constraints)
- **Kein Node.js:** Keine Node.js APIs (`fs`, `os`, `process`, `net`) verfügbar.
- **Kein Netzwerk:** `fetch` und `XMLHttpRequest` sind in der Sandbox deaktiviert.
- **Keine Timer:** `setTimeout` und `setInterval` sind deaktiviert (Sandboxed V8).
- **CGO Erforderlich:** Da `v8go` CGO benötigt, müssen alle Builds `CGO_ENABLED=1` gesetzt haben.

## Styling-Regeln
- **Variablen:** Go (PascalCase/camelCase), TypeScript (camelCase)
- **Kommentare:** Dokumentation wichtiger interner Logik (Bundler/Executor) auf Englisch.
- **Error-Handling:** Explizite Go Error Checks; Source Map Auflösung bei Sandbox-Fehlern Pflicht.

---

## 📋 Meta

- **Zuletzt aktualisiert:** 2026-04-15
- **Aktualisiert von:** Gemini-CLI
- **Status:** Aktuell
