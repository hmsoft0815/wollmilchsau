# 🏗 Infrastructure & Deployment: wollmilchsau — LLM Script Execution Sandbox

## Hosting & Umgebungen

| Umgebung | URL / Adresse | Anmerkung |
|----------|--------------|-----------|
| Production | Linux (amd64) | Empfohlene Plattform (Binary oder Docker) |
| Staging | — | — |
| Local Dev | localhost (stdio/SSE) | Standard für Claude Desktop (stdio) |

## Deployment

- **Methode:** Docker Build oder Manual Build via Taskfile. Release-Binaries via GitHub Actions.
- **Start-Befehl:** `./wollmilchsau -addr :8080` (SSE/HTTP Mode) oder via stdio (MCP Default).
- **Config-Pfad:** `-log-dir /path/to/logs` für Request-Audit-Archive (ZIP).
- **Logs:** Strukturierte JSON-Logs an `stdout` oder Archive in `-log-dir`.

## Abhängigkeiten (Laufzeit)

- **V8 Engine:** (Laufzeit-Isolation) — statisch gelinkt/in-process via `v8go` (CGO).
- **esbuild:** (TypeScript Transpilation) — in-process.
- **mlcartifact:** (Optional) — gRPC Server auf `localhost:50051` für persistente Artefakte.

## Secrets / Umgebungsvariablen

| Variable | Zweck | Wo gesetzt |
|----------|-------|------------|
| `WOLL_MILCHSAU_ADDR` | SSE Listen Address | OS Environment / Override Flag |
| `WOLL_MILCHSAU_LOG_DIR` | Directory for ZIP logs | OS Environment / Override Flag |
| `WOLL_MILCHSAU_ARTIFACTS_ENABLED` | Enable Artifact API | OS Environment / Override Flag |
| `WOLL_MILCHSAU_ARTIFACTS_ADDR` | Artifact Server (gRPC) | OS Environment / Override Flag |

## Monitoring & Alerts

- **Health-Check:** `GET /health` (Nur im SSE-Modus verfügbar).
- **Alerts:** Keine integrierte Alerting-Lösung. Empfehlung: External Uptime Monitoring für SSE-Endpoints.
- **Metrics:** Keine nativen Prometheus/OpenTelemetry Metriken (Lokal-Fokus).

---

## 📋 Meta

- **Zuletzt aktualisiert:** 2026-04-17
- **Aktualisiert von:** gemini-2.0-flash-exp
- **Status:** Aktuell
