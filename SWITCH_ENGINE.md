# 🔄 Projektplan: JavaScript Engine Migration

> ## ⛔ Verworfen am 2026-08-30
>
> **Dieser Plan wird nicht umgesetzt.** V8 läuft sauber, und der Umbau wäre
> ein Eingriff in die Ausführungsschicht — der zentrale und heikelste Teil
> dieses Servers — ohne einen Fehler, der ihn erzwingt.
>
> Die genannten Gründe stimmen weiterhin (Binärgröße, CGO, Sandbox). Sie
> wiegen nur den Aufwand und das Risiko nicht auf, solange die vorhandene
> Engine ihre Aufgabe erfüllt.
>
> **Umgesetzt wurde davon nichts.** Weder QuickJS noch eine andere Engine
> kommt im Code vor; `rogchap.com/v8go` steht unverändert in `go.mod` und wird
> an sechs Stellen benutzt. Es gibt also nichts zurückzunehmen.
>
> Das Dokument bleibt bewusst stehen, statt gelöscht zu werden: die Analyse
> ist brauchbar, und ohne den Vermerk wird derselbe Vorschlag in einem halben
> Jahr neu geschrieben. Wer ihn wieder aufgreifen will, sollte zuerst
> benennen, welches konkrete Problem V8 heute verursacht.
>
> Der Branch `feature/quickjs-engine` trägt seinen Namen von diesem Plan,
> enthält aber ausschließlich die davon unabhängige Arbeit an gebündelten
> npm-Abhängigkeiten.

## Übersicht

**Ziel:** V8 (via `rogchap.com/v8go`) durch eine alternative JavaScript-Engine ersetzen, um die Abhängigkeiten zu reduzieren, die Binärgröße zu verkleinern und die Sandbox-Sicherheit zu verbessern.

**Gründe für die Migration:**
- V8 bringt ~150 MB native Bibliotheken mit (libc++, libv8)
- CGO-Abhängigkeit erfordert gcc/g++ in Docker-Builds — erhöht Angreiffläche und Build-Zeit
- Die Sandbox-Funktionen (kein fs, kein network, keine Timer) lassen sich mit leichtgewichtigeren Engines einfacher erzwingen
- Reduzierte Binärgröße für Edge/Container-Deployment

---

## 1. Engine-Vergleich

### Option A: QuickJS (Empfehlung ⭐)
| Kriterium | Wert |
|-----------|------|
| **Go-Binding** | `github.com/NobleLi/go-quickjs` oder eigener FFI-Wrappper |
| **Binärgröße** | ~2–5 MB (QuickJS hat nur ~300 KB native Lib) |
| **Speicherbedarf** | ~1–5 MB pro Isolate |
| **Standardsupport** | ES2020 (gut, aber nicht ES2024+ wie V8) |
| **Performance** | Langsamer als V8 (~2–5x), aber ausreichend für MCP-Skripte |
| **Sandbox-Kontrolle** | Einfach — nur API-Whitebox injizieren, Standardmäßig leer |
| **Lizenz** | MIT (Bindings) / ISC (QuickJS selbst) |
| **Linux-Unterstützung** | Exzellent — QuickJS ist rein C, keine Platform-Specifics |
| **CGO** | Erforderlich für FFI-Binding, aber gcc nur während des Builds nötig (nicht runtime) |

### Option B: SpiderMonkey (Mozilla)
| Kriterium | Wert |
|-----------|------|
| **Go-Binding** | Kein stabiles Go-Binding — müsste neu geschrieben werden |
| **Binärgröße** | ~30–50 MB (libmozjs) |
| **Speicherbedarf** | ~10–30 MB pro Isolate |
| **Standardsupport** | Voller ES-standards (Mozilla-Firefox-Akte) |
| **Performance** | Nahe an V8 |
| **Sandbox-Kontrolle** | Mittel — JSContext kann restringen, aber komplexer |
| **Lizenz** | MPL 2.0 (copyleft-Like) |

### Option C: Deno Core (QuickJS-basiert)
| Kriterium | Wert |
|-----------|------|
| **Go-Binding** | `deno_core.rs` via CGO — noch experimentell |
| **Binärgröße** | ~20–30 MB |
| **Standardsupport** | Vollständig (Deno/Firefox-nah) |
| **Nachteil** | Sehr schwergewichtig, Rust-Dependency-chain, experimentelle Go-Bindings |

### Entscheidung: QuickJS ⭐

- Kleinste Angriffsfläche und Binärgröße
- Ausreichende ECMA-Standar-Konformität für MCP-Skripte
- Einfachste Sandbox-Implementierung (Standard = leer)
- Gut dokumentierte C-API

---

## 2. Architektur-Änderungen

### 2.1 Neue Paketstruktur

```
internal/
├── executor/          ← bestehend, enthält types.go
│   ├── quickjs_engine_qjs.go  ← NEW: QuickJS Isolate-Manager
│   ├── quickjs_inject.go      ← NEW: Polyfill-Injection für QJS
│   ├── polyfills.go           ← REFACTOR: Engine-agnostische Polyfills als JS-Src-Strings
│   ├── artifacts.go           ← UNVERÄNDERT
│   ├── executor.go            ← REFACTOR: Execute() delegiert an Engine-Interface
│   └── types.go               ← UNVERÄNDERT
```

### 2.2 Engine-Interface (Abstraktion)

Neues Interface in `internal/executor/engine.go`:

```go
package executor

// Engine abstracts the JavaScript runtime. Every engine must implement this.
type Engine interface {
    NewIsolate() Isolate
    Name() string
}

type Isolate interface {
    Close() error
    RunScript(src, filename string) (Value, error)
    GlobalObject() Object
    GetHeapStats() HeapStats
    TerminateExecution()
}
```

---

## 3. Implementierungsphasen

### Phase 1: Foundation (Tag 1–2) ⏱️ Prio: Hoch

**Ziel:** Interface-Abstraktion + QuickJS-PACKET erstellen. Noch kein Live-Umschalten. Der bestehende V8-Pfad bleibt vollfunfktionsfähig.

| # | Task | Aufwand | Datei(en) |
|---|------|---------|-----------|
| 1.1 | `internal/executor/engine.go` — Interface-Definition | 30 Min | engine.go |
| 1.2 | V8-Adapter: `v8_impl.go` — bestehenden Code auf Interface portieren | 2 Std | v8_impl.go, v8_inject.go |
| 1.3 | QuickJS-Paket erstellen: Bindings zu `libquickjs.so` | 4 Std | quickjs_engine_qjs.go |
| 1.4 | Polyfills als JS-Strings extrahieren (engine-agnostisch) | 1 Std | polyfills.go → js_src_packs/ |
| 1.5 | Unit-Tests für Engine-Interface schreiben | 2 Std | engine_test.go |

### Phase 2: QuickJS-Implementierung (Tag 3–4) ⏱️ Prio: Hoch

**Ziel:** QuickJS-basierter Executor funktionsfähig. Feature-parity mit V8-Pfad prüfen.

| # | Task | Aufwand | Datei(en) |
|---|------|---------|-----------|
| 2.1 | `NewContext()` in C via cgo wrappen | 2 Std | quickjs_bridge.c/h |
| 2.2 | `RunScript()` — Skript-Ausführung implementieren | 2 Std | quickjs_engine_qjs.go |
| 2.3 | Object/Function-Templates für console.log/warn/error/injecten | 2 Std | quickjs_inject_qjs.go |
| 2.4 | JS-Polyfills (performance, atob/btoa, crypto, TextEncoder, Buffer) auf QJS portieren | 4 Std | polyfills.go + quickjs_bridge.c/h |
| 2.5 | Error-Handling: QuickJS Exceptions parsen → Diagnostic | 2 Std | quickjs_error.go |
| 2.6 | `__encode_utf8`/`__decode_utf8` bridge für Unicode-Korrektur in QJS | 1 Std | quickjs_bridge.c/h |
| 2.7 | Memory Limit Watchdog (Heap-Statistiken aus QJS lesen) | 2 Std | watchdog.go |
| 2.8 | Artifact-Polyfill-Injektion (`artifact.*` API injizieren) | 2 Std | quickjs_artifact_inject.go |
| 2.9 | Integrationstest: `execute()` mit QuickJS → gleiche Resultate wie V8 | 3 Std | integration_test.go |

### Phase 3: Umschalte-Logik & Konfiguration (Tag 5) ⏱️ Prio: Mittel

**Ziel:** Engine über CLI-Flag umschaltbar. Default = QuickJS. Rollback auf V8 möglich.

| # | Task | Aufwand | Datei(en) |
|---|------|---------|-----------|
| 3.1 | CLI-Flag `-engine v8\|quickjs` in `cmd/main.go` | 30 Min | main.go, flags.go |
| 3.2 | Engine-Factory: NewEngine(ename string) Engine in server/server.go | 1 Std | engine_select.go |
| 3.3 | Config-Passing durch Handler-Kette (Execute → executor.Execute) | 1 Std | execution.go, handler |
| 3.4 | Feature-Matrix prüfen: Alle Polyfills funktionieren mit QJS | 2 Std | — |
| 3.5 | Fallback-Logik: Wenn QuickJS nicht verfügbar (.so fehlt), warn + auto-fallback auf V8 | 1 Std | engine_select.go |

### Phase 4: Docker & Build (Tag 6) ⏱️ Prio: Mittel

**Ziel:** Docker-Build mit libquickjs statt gcc/libv8. Runtime ohne gcc.

| # | Task | Aufwand | Datei(en) |
|---|------|---------|-----------|
| 4.1 | Dockerfile-Builder-Schritt anpassen (gcc/g++/make → nur für Build, nicht runtime) | 30 Min | Dockerfile |
| 4.2 | Runtime-Schritt: libquickjs installieren statt libc6+libstdc++6 | 30 Min | Dockerfile |
| 4.3 | Binary-Größe messen und dokumentieren | 15 Min | — |

### Phase 5: Tests & Qualität (Tag 7) ⏱️ Prio: Mittel

**Ziel:** Feature-parity sicherstellen, Regressionen finden.

| # | Task | Aufwand | Datei(en) |
|---|------|---------|-----------|
| 5.1 | Alle bestehenden Unit-Tests mit QuickJS durchlaufen | 2 Std | tests/ |
| 5.2 | MCP-Skript-Integrationstests (basic.mcp, extended.mcp) gegen neuen Engine-Pfad | 3 Std | test-mcp |
| 5.3 | Performance-Benchmark: V8 vs QuickJS auf gleicher Arbeitloads | 1 Std | benchmark_test.go |
| 5.4 | Sicherheitsreview der Sandbox: alle Node.js- APIs blockiert? | 2 Std | — |
| 5.5 | Error-mapping prüfen: SourceMap-Auflösung funktioniert mit QJS-Error-Format | 1 Std | sourcemap/sourcemap.go |

### Phase 6: Dokumentation & Cleanup (Tag 8) ⏱️ Prio: Niedrig

| # | Task | Aufwand | Datei(en) |
|---|------|---------|-----------|
| 6.1 | ARCHITECTURE.md aktualisieren: Engine-Abstraktion dokumentieren | 1 Std | .mlcai/ARCHITECTURE.md |
| 6.2 | TECH_STACK.md update: V8 → QuickJS, Deps anpassen | 30 Min | .mlcai/TECH_STACK.md |
| 6.3 | `go.mod` bereinigen: v8go entfernen (als deprecated markieren) | 30 Min | go.mod |
| 6.4 | Alte V8-Implementierung (`v8_impl.go`) als DEPRECATED kennzeichnen + Remove-Milestone setzen | 30 Min | — |

---

## 4. Migrationsschritte im Code (Detail-Look)

### Schritt 1: bestehende `Execute()` umstrukturieren

**Aktuell** (`executor.go:27`):
```go
func Execute(ctx, js, filename, artifactAddr string) *Result {
    iso := v8.NewIsolate()        // V8-spezifik — weg damit
    defer iso.Dispose()
    v8ctx := v8.NewContext(...)   // V8-spezifik — weg damit
    val, runErr := v8ctx.RunScript(js, filename)
    ...
}
```

**Neu**:
```go
func Execute(ctx context.Context, cfg ExecuteConfig) *Result {
    type ExecuteConfig struct {
        Script       string
        Filename     string
        SourceMap    *sourcemap.SourceMap
        ArtifactAddr string
        Engine       executor.Engine  // → "v8" oder "quickjs"
    }

    iso := cfg.Engine.NewIsolate()
    defer iso.Close()
    val, runErr := iso.RunScript(cfg.Script, cfg.Filename)
    ...
}
```

### Schritt 2: Polyfills engine-agnostisch machen

Polyfills als JS-Quelltext-Strings in `polyfills/engine_polyfills.go` (getrennt von der Bridge-Logik). QuickJS erhält eigene Inject-Funktion (`quickjs_inject_qjs.go`), V8 hat die bestehende (`v8_inject_v8go.go`).

### Schritt 3: Error-Format-Kompatibilität

**Problem:** V8 gibt Errors im Format `Error: message\n at file.js:123:45`. QuickJS verwendet `Error: message\n    at file.js:123 (eval)。

Lösung in `executor.go:handleExecute()`:
```go
func handleExecuteError(err error, sm *sourcemap.SourceMap, res *Result) {
    d := extractDiagnostic(err, sm)  // parsed mit engine-agnostischem Regex
    ...
}
```

Der Diagnostic-Extractor muss das neue Error-Format von QuickJS unterstützen — ein kleiner zusätzlicher Parser.

### Schritt 4: Docker-Konfiguration ändern

**Aktuell (Dockerfile):**
```dockerfile
# Builder stage
apt-get install -y gcc g++ make

# Runtime
apt-get install -y ca-certificates libc6 libstdc++6
```

**Neu:**
```dockerfile
# Builder stage — gcc nur für CGO-Compilation, nicht runtime
apt-get install -y gcc
COPY QuickJS/libquickjs.a /usr/lib/

# Runtime — keine gcc nötig, nur libquickjs.so + ca-certificates
apt-get install -y ca-certificates libquickjs1
```

---

## 5. Risikobewertung

| Risiko | Schwere | Wahrscheinlichkeit | Gegenmaßnahme |
|--------|---------|-------------------|---------------|
| QuickJS unterstützt kein gewünschtes ES-Feature des MCP-Skripts | Hoch | Mittel | Polyfill-JS-Strategie: feature-detect → polyfill. Fallback auf V8 via `-engine v8`. |
| QuickJS-Binding Memory-Leaks im cgo-Layer | Mittel | Mittel | Exhaustiver Integrationstest; `runtime.MemStats` Monitoring |
| performance_now-Polyfill mit QJS Timing nicht präzise genug | Niedrig | Hoch | QJS hat eigene Timer — polyfill muss mit Go `time.Now()` via cgo syncen |
| TextEncoder/TextDecoder-Encoding in QJS anders als V8 | Mittel | Mittel | Alle Encoding-Tests müssen explizit abgedeckt werden |
| QuickJS `.so` nicht auf alten Linux-Distros verfügbar | Mittel | Mittel | libquickjs.a statisch linken (CGO_LDFLAGS), keine Runtime-Abhängigkeit |

---

## 6. Feature-Matrix: V8 → QuickJS Parity Check

Je Polyfill / API muss geprüft werden, ob das Verhalten identisch ist:

| API | V8-Status | QuickJS-Ziel | Blocker? |
|-----|-----------|-------------|----------|
| `console.log/info/warn/error` | ✓ via cgo FunctionTemplate | ✓ ObjectTemplate + JS-Funktion | Nein |
| `setTimeout/setInterval` | ✗ deaktiviert | Deaktivieren (Standard) | Nein |
| `performance.now()` | ✓ Polyfill via time.Now | ✓ Polyfill via cgo bridge | Nein |
| `atob/btoa` | ✓ Polyfill | ✓ identisch (base64, JS-seitig) | Nein |
| `crypto.getRandomValues()` | ✓ Polyfill via rand.Read + `__get_random_b64` | ✓ Polyfill via rand.Read cbridge | Kein Blocker |
| `TextEncoder.encode()` | ✓ Polyfill via UTF8-Bridge | ✓ identisch JS-Seite | Kein Blocker |
| `TextDecoder.decode()` | ✓ Polyfill via UTF8-Bridge | ✓ identisch JS-Seite | Kein Blocker |
| `Buffer.from/alloc` | ✓ Minimal-Polyfill | ✓ identisch JS-Seite | Nein |
| `fetch/XMLHttpRequest` | ✗ deaktiviert | Nicht unterstützt von QJS (default) | Nein |
| `fs/require/module.paths` | ✗ nicht verfügbar in V8-Sandbox | Nicht unterstützt von QJS (default) | Nein |
| Quelltext-Error-Location-Parsing | ✓ via jsErr.Location | Muss neu: QuickJS Exception Stack Trace Format | **Ja** — kleiner Aufwand |

---

## 7. Zeitabschätzung

| Phase | Personen-Wochen | Vorbedingung |
|-------|-----------------|-------------|
| 1. Foundation | ~1 Personentag | keine |
| 2. QuickJS-Implementierung | ~3 Personentage | Phase 1 fertig |
| 3. Umschalte-Logik | ~1 Personentag | Phase 2 fertig |
| 4. Docker/Build | ~0.5 Personen-tag | Phase 3 fertig |
| 5. Tests & Qualität | ~2 Personen-tage | Phase 3 fertig (parallel zu 3 möglich) |
| 6. Cleanup | ~~1 Personentag | alle Phasen fertig |
| **Gesamt** | **~7–8 Personentage** | sequenziell ~10 Tage, parallelisierung möglich |

---

## 8. Rollback-Plan

Sollte die Migration scheitern oder kritische Regressionen gefunden werden:

1. **Immediate:** CLI-Fallback auf V8 via `-engine v8` — bleibt vollständig aktiviert
2. **Kurzfristig:** QuickJS als `--experimental` flag kennzeichnen (nicht default)
3. **Langfristig:** Falls nach Q4 2026 keine kritischen Bugs: QuickJS wird default, V8 als deprecated markiert

**Rollback-Code:**
```bash
# Sofortiger Rollback: Engine zurück auf V8 setzen
wollmilchsau --engine v8 --addr :8000
```

---

## 📋 Meta

- **Erstellt:** 2026-08-19
- **Autor:** CLAUDE-CLI
- **Status:** Entwurf
- **Abhängigkeiten:** Keine — eigenständiger Migrationsplan
- **Review-Erforderlich:** Architektur-Led (Performance-Messung vor finaler Bindings-Wahl)
