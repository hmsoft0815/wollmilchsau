# 🔌 API Contract: wollmilchsau

## 1. Bereitgestellte Endpunkte (Exposed)
Dieses Projekt ist primär ein MCP Server (Model Context Protocol).

### MCP Tools (v1)
- **Kommunikation:** stdio (Standard für lokale Agents) oder SSE (HTTP Port 8000).
- **Tools:**
  - `execute_script`: Führt ein einzelnes TypeScript/JavaScript Snippet aus.
  - `execute_project`: Führt ein Multi-File TypeScript Projekt aus (Dateien + Entry Point).
  - `check_syntax`: Validiert TypeScript Syntax ohne Ausführung.
  - `execute_artifact`: Lädt ein Code-Artefakt von `mlcartifact` und führt es aus.
- **Fehlerformat:**
  ```json
  { "success": false, "summary": "Error message", "diagnostics": [ { "file": "main.ts", "line": 5, "column": 10, "message": "Syntax error" } ] }
  ```

---

## 2. Konsumierte APIs (Consumed)
Von welchen externen oder internen Diensten hänge ich ab?

### Intern: mlcartifact (gRPC/Connect)
- **Zweck:** Speichern und Abrufen von Artefakten (CSV, Reports, Skripte).
- **Endpunkt:** `localhost:50051` (standardmäßig).
- **Abhängigkeit:** Optional (muss mit `-enable-artifacts` aktiviert werden).

---

## 3. Globale Konventionen
- **Datumsformat:** ISO 8601 (UTC) für Logging und Metadaten.
- **Limits:** 128MB heap memory, 10s default timeout für Skripte.
- **Logging:** ZIP-Archive für Request-Logs können via `-log-dir` konfiguriert werden.

---

## 📋 Meta

- **Zuletzt aktualisiert:** 2026-04-15
- **Aktualisiert von:** Gemini-CLI
- **Status:** Aktuell
