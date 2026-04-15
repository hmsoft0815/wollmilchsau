# 🤖 AI & Integration Context: wollmilchsau

## 1. Identität & Zweck
- **Kernaufgabe:** Sandboxed TypeScript/JavaScript Ausführungsumgebung für KI-Agenten (MCP Server). Ermöglicht deterministische Berechnungen und Datentransformationen ohne "Reasoning-Overhead".
- **Technischer Stack:** Go (Backend), V8 (Execution), esbuild (Bundling).
- **Hoster/Infrastruktur:** Docker, Binary (Linux amd64), MCP-Client Integration (stdio/SSE).

## 2. Die "Nachbarschaft" (System-Kontext)
- **Upstream (Wovon hänge ich ab?):**
  - **mlcartifact** -> [https://github.com/hmsoft0815/mlcartifact](https://github.com/hmsoft0815/mlcartifact) -> Optionaler Dienst zum Speichern großer Ausgaben als persistente Artefakte.
- **Downstream (Wer nutzt mich?):**
  - **KI-Agenten (z.B. Claude Desktop, Custom Agents)** -> Nutzen die `execute_script` und `execute_project` Tools für Code-Offloading.
- **Shared Resources:**
  - Nutzt den `mlcartifact` gRPC-Endpunkt (standardmäßig `localhost:50051`) falls aktiviert.

## 3. Schnittstellen-Vertrag
- **Primäre API:** MCP (Model Context Protocol) über stdio oder SSE (HTTP).
- **Auth-Mechanismus:** Keine interne Auth für stdio; SSE ist für interne/gesicherte Netzwerke vorgesehen.
- **Wichtige Datenmodelle:**
  - `ExecutionResult`: { success: boolean, stdout: string, stderr: string, result: any, diagnostics: Diagnostic[] }
- **API-Doku-Link:** MCP Tool Schema kann via `./wollmilchsau -dump` extrahiert werden.

## 4. Leitplanken & Regeln
- **Naming:** Go-Konventionen (PascalCase für exportierte Typen), TypeScript (camelCase) für die Sandbox-Skripte.
- **Testing:** Unit-Tests für Bundler und Executor; Integrationstests via `mcp-tester`.
- **Sicherheit:** Strikte Sandbox-Isolierung (kein Netzwerk, kein Dateisystem, kein Zugriff auf Node.js APIs).

## 5. Aktueller Fokus (Status)
- **Bekannte Probleme:** CGO-Abhängigkeit erschwert Cross-Compilation für macOS/Windows (manuelle Builds erforderlich).
- **Nächste Schritte:** Optimierung der Source Map Auflösung für komplexe Multi-File Projekte.

---

## 📋 Meta

- **Zuletzt aktualisiert:** 2026-04-15
- **Aktualisiert von:** Gemini-CLI
- **Status:** Aktuell
