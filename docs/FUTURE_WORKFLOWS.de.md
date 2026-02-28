# Zukünftige Workflows: MCP-Orchestrierung
Copyright (c) 2026 Michael Lechner. Alle Rechte vorbehalten.

Dieses Dokument beschreibt die Vision für **wollmilchsau** als programmierbare Middleware für KI-Agenten.

## Das Konzept: "MCP-Fetch"

Die Idee ist, **wollmilchsau** nicht nur als isolierte Sandbox zu nutzen, sondern als Orchestrator, der mit anderen MCP-Servern kommunizieren kann. Dies ermöglicht komplexe Datenverarbeitungs-Workflows direkt in TypeScript/JavaScript.

### 1. Registrierung von Servern
Beim Start von **wollmilchsau** wird eine Konfigurationsdatei (z. B. `mcp_registry.json`) geladen, die bekannte MCP-Server definiert:

```json
{
  "postgres": { "type": "sse", "url": "http://db-server:8080/sse" },
  "weather": { "type": "stdio", "command": "weather-mcp-server" }
}
```

### 2. Das `mcp`-Objekt in der Sandbox
In der V8-Sandbox wird ein globales Objekt `mcp` injiziert, das ein Fetch-ähnliches Interface bietet:

```javascript
// Beispiel: Statistische Auswertung von Datenbank-Daten
const rawData = await mcp.call("postgres", "query", { 
    sql: "SELECT age FROM users WHERE active = true" 
});

// Verarbeitung direkt in wollmilchsau
const ages = rawData.map(r => r.age);
const avgAge = ages.reduce((a, b) => a + b, 0) / ages.length;

console.log(`Durchschnittsalter der aktiven Nutzer: ${avgAge}`);
```

## Anwendungsfälle

### Statistische Auswertungen
Anstatt dass die LLM hunderte Zeilen Rohdaten lesen muss (was das Kontext-Fenster sprengt), schreibt sie ein kurzes Skript für **wollmilchsau**. Das Skript holt die Daten vom Datenbank-MCP, berechnet Durchschnitt, Varianz oder Trends und gibt nur das kompakte Ergebnis zurück.

### Komplexe Daten-Transformationen
Daten von einem API-MCP-Server abrufen, mit Daten aus einem Datei-MCP-Server mergen und das Ergebnis strukturiert aufbereiten.

### Programmierbare Tests
Ein Skript, das mehrere Tools anderer Server nacheinander aufruft und die Konsistenz der Rückgabewerte prüft.

## Grafik-Erstellung & Visualisierung

Eine weitere spannende Erweiterung ist die Kombination von **wollmilchsau** mit spezialisierten Grafik-MCP-Servern (z. B. für D2, Mermaid, Gnuplot oder SVG-Generierung).

**Workflow:**
1.  Ein Skript in **wollmilchsau** berechnet komplexe Daten (z. B. eine Fibonacci-Folge oder statistische Verteilungen).
2.  Das Skript ruft via `mcp.call` einen Grafik-Server auf und übergibt die berechneten Daten.
3.  Der Grafik-Server gibt ein Bild (SVG/PNG) zurück, welches **wollmilchsau** dem Nutzer präsentiert.

Dies ermöglicht "programmierbare Grafiken", bei denen die Logik in TypeScript liegt und die Darstellung von spezialisierten Tools übernommen wird.

## Technische Umsetzung (Roadmap)
1.  **Bridge-Logic:** Go-seitige Implementierung eines MCP-Clients, der SSE- und Stdio-Verbindungen hält.
2.  **V8-Injektion:** Mapping der Go-Client-Funktionen auf JavaScript-Callbacks (`v8go.FunctionTemplate`).
3.  **Async/Await Support:** Da MCP-Calls asynchron sind, muss die Sandbox korrekt mit Promises umgehen können (V8 Taskrunner-Integration).

---
*Diese Vision macht **wollmilchsau** zur Schaltzentrale für intelligente, datengetriebene Agenten-Workflows.*
