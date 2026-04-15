# 📜 Decision Log (ADR)

Dieses Dokument listet fundamentale Entscheidungen auf, die im Projekt getroffen wurden.

## [2026-04-15]: Sandboxed V8 statt Node.js Subprozess
- **Kontext:** Wahl zwischen Node.js als Subprozess oder In-Process V8.
- **Entscheidung:** Wir nutzen `rogchap.com/v8go` (V8 Bindings).
- **Grund:** Minimierung des Overheads, strikte Sandbox-Isolierung ohne Node.js APIs, bessere Performance bei Mikrosekunden-Laufzeiten.
- **Konsequenz:** Erfordert CGO für Builds; Cross-Compilation ist schwieriger.

## [2026-04-15]: In-Process esbuild
- **Kontext:** Wahl zwischen einem externen Bundler (z.B. tsc) oder esbuild.
- **Entscheidung:** Wir nutzen `github.com/evanw/esbuild`.
- **Grund:** Extrem hohe Performance (Mikrosekunden-Bundling), keine externe Abhängigkeit erforderlich.
- **Konsequenz:** Bundling erfolgt In-Memory vor der V8-Ausführung.

---

## [2026-04-15]: Artefakt-Integration (mlcartifact)
- **Kontext:** Speicherung großer Skript-Ausgaben (Reports, CSVs).
- **Entscheidung:** Optionale Integration von `mlcartifact` via gRPC.
- **Grund:** Saubere Trennung von Ausführung (wollmilchsau) und Persistenz (mlcartifact).
- **Konsequenz:** Ermöglicht die Anzeige von Artefakten direkt im KI-Agenten-Interface.

---

## 📋 Meta

- **Zuletzt aktualisiert:** 2026-04-15
- **Aktualisiert von:** Gemini-CLI
- **Status:** Aktuell
