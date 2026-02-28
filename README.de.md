# wollmilchsau (Go + V8 + esbuild)

MCP-Server in Go — Hochperformante TypeScript-Ausführung mit eingebettetem V8 und esbuild.

> 🇬🇧 [English Version](README.md)

---

## Warum Model Context Protocol (MCP)?

KI-Agenten müssen oft Code ausführen oder Daten verarbeiten, um komplexe Aufgaben zu erfüllen. Während LLMs gut darin sind, Code zu schreiben, können sie diesen nicht sicher selbst ausführen.

**wollmilchsau** fungiert als "isolierte Gehirnerweiterung":
- **Sicherheit**: Der Code läuft in einer isolierten V8-Umgebung ohne Netzwerk- oder Dateisystemzugriff.
- **Geschwindigkeit**: In-Process-Bundling (esbuild) und V8-Ausführung bedeuten null Node.js-Overhead.
- **Selbstkorrektur**: Strukturierte Fehler und Source-Maps ermöglichen es Agenten, ihre eigenen Bugs zu fixen.

---

## Features

- **In-Process Bundling:** Nutzt `esbuild` direkt in Go (kein Node.js-Subprozess erforderlich).
- **Isolierte Ausführung:** Führt Code in frischen V8-Isolates aus.
- **Source-Map-Unterstützung:** Fehler werden automatisch auf die ursprünglichen TS-Dateien zurückgeführt.
- **LLM-optimierte Ausgabe:** Strukturierte JSON-Metadaten und getrennte Inhaltsblöcke.
- **SSE & Stdio Support:** Betrieb als lokaler Prozess oder eigenständiger HTTP-Server.
- **Artefakt-Integration:** Automatische Anbindung an **mlcartifact**, um Ausführungsergebnisse, große Datenblöcke oder generierte Berichte persistent zu speichern.
- **Request-Archivierung (ZIP-Logging):** Optionale vollständige Archivierung jedes Requests (Quelldateien + Metadaten + Ergebnis) in kompakten ZIP-Dateien.

## Stack

| Komponente | Library | Zweck |
|---|---|---|
| MCP-Protokoll | `mark3labs/mcp-go` | JSON-RPC 2.0 Implementierung |
| TS-Bundling | `evanw/esbuild` | Schnelle, In-Process Transpilierung |
| JS-Ausführung | `rogchap/v8go` | CGo-Bindings zu V8 |
| Source-Maps | Custom | VLQ-Dekodierung und Positionsauflösung |
| Logging | `log/slog` | Strukturiertes Logging für den Produktivbetrieb |

## Erste Schritte

### Voraussetzungen

- **Go 1.23+**
- **C++ Compiler:** `build-essential` (Linux) oder `llvm` (macOS).

### Build

```bash
make build
# Ausgabe: build/wollmilchsau
```

### Betrieb

Der Server unterstützt zwei Transport-Modi:

1. **stdio (Standard):** Ideal für die lokale Nutzung mit Claude Desktop.
   ```bash
   ./build/wollmilchsau
   ```
2. **SSE (HTTP):** Eigenständiger Server für Remote-Zugriff.
   ```bash
   ./build/wollmilchsau -addr :8080
   ```

### CLI-Flags

- `-addr string`: Listen-Adresse für SSE (z.B. `:8080`). Falls leer, wird stdio genutzt.
- `-log-dir string`: Pfad zu einem Verzeichnis, in dem jeder Request als ZIP-Datei archiviert wird.
- `-version`: Zeigt Versionsinformationen (wollmilchsau, V8, esbuild).
- `-dump`: Gibt das komplette MCP-Tool-Schema als JSON aus.

## Erweitertes Request-Logging

Wenn `-log-dir` angegeben wird, erstellt wollmilchsau für jeden eingehenden Tool-Aufruf ein ZIP-Archiv. Dies ist ideal für die Überprüfung und das Debugging von LLM-Verhalten, ohne die primären Logdateien aufzublähen.

Jede ZIP-Datei enthält:
- `info.json`: Metadaten (Zeitstempel, Client-IP, Tool-Name, Ausführungsplan).
- `src/`: Alle vom LLM bereitgestellten virtuellen Quelldateien.
- `response.json`: Das vollständige JSON-Ergebnis des Executors.

## Tools

### `execute_script`
Führt einen einzelnen Code-Snippet aus. Ideal für schnelle Berechnungen.
- `code`: Der auszuführende TypeScript/JavaScript Code.
- `timeoutMs`: (Optional) Maximale Laufzeit.

### `execute_project`
Führt ein Projekt mit mehreren Dateien aus.
- `files`: Array aus `{name, content}` Objekten.
- `entryPoint`: Die Startdatei (z.B. `main.ts`).

## 🔒 Einschränkungen der Ausführungsumgebung

Um Sicherheit und Performance zu gewährleisten, ist die Umgebung streng isoliert (Sandboxed):

- **Ressourcen-Limits:**
  - **Arbeitsspeicher:** Aktives Heap-Monitoring. Skripte sind auf **128MB** genutzten Heap begrenzt. Das Überschreiten führt zum sofortigen Abbruch.
  - **CPU / Zeit:** Konfigurierbarer Timeout (Standard 10s). Skripte werden nach Ablauf hart via `iso.TerminateExecution()` gestoppt.
- **Nur ECMA-262:** Reines V8-Sandbox-Environment. Moderne JS/TS Features werden unterstützt, aber umgebungsspezifische APIs sind eingeschränkt.
- **Kein Netzwerk:** `fetch`, `XMLHttpRequest` oder jeglicher andere Netzwerkzugriff ist deaktiviert.
- **Keine Event-Loop-Timer:** `setTimeout`, `setInterval` und `setImmediate` stehen nicht zur Verfügung. Die Ausführung erfolgt rein synchron.
- **Keine Node.js / Web APIs:** Kein Zugriff auf `fs`, `os`, `process` oder DOM APIs.
- **Eingeschränktes i18n:** Das `Intl` Objekt ist verfügbar, aber auf die Locale `en-US` beschränkt.
- **Reine Logik:** Ideal für Algorithmen, Datentransformationen und mathematische Berechnungen.

## Claude Desktop Konfiguration

Ergänzen Sie Ihre `claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "wollmilchsau": {
      "command": "/absoluter/pfad/zu/wollmilchsau/build/wollmilchsau",
      "args": ["-log-dir", "/pfad/zu/logs"]
    }
  }
}
```

## 🚀 Ausblick: MCP-Orchestrierung

In Zukunft soll **wollmilchsau** als Orchestrator für andere MCP-Server fungieren. Durch ein Fetch-ähnliches Interface innerhalb der Sandbox können Skripte Daten von anderen Servern (z. B. Datenbanken) abrufen und verarbeiten.

Details zu dieser Vision finden Sie unter [FUTURE_WORKFLOWS.de.md](docs/FUTURE_WORKFLOWS.de.md).

## 📜 Lizenz & Ethische Nutzung

Dieses Projekt steht unter der **MIT-Lizenz**.

### 🕊️ Anmerkung des Autors (Nicht bindend)
Obwohl die Lizenz eine breite Nutzung erlaubt, bitte ich (der Autor) darum, diese Software **nicht** für folgende Zwecke zu verwenden:
* **Militärische Zwecke** oder die Produktion und Entwicklung von Waffen.
* Aktivitäten von Organisationen oder Personen, die die **militärische Aggression gegen die Ukraine** unterstützen.

Des Weiteren bitte ich aufgrund vergangener beruflicher Erfahrungen meinen ehemaligen Auftragsgeber **Isensix, Inc.** sowie dessen Käufer **Dwyer-Omega** ausdrücklich darum, diese Software in keiner Weise zu nutzen.

*Diese Bitte ist ein Appell an die Berufsethik und das persönliche Gewissen und stellt keine rechtliche Änderung der MIT-Lizenz dar.*
