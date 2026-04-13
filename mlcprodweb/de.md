# Wollmilchsau

Wollmilchsau ist eine Hochleistungs-Engine zur Ausführung von TypeScript, die speziell für Model Context Protocol (MCP) Server entwickelt wurde.

## Was es macht

Die Ausführung von TypeScript-Logik innerhalb eines MCP-Servers erfordert normalerweise eine vollständige Node.js- oder Deno-Runtime. Dies führt oft zu erheblichem Overhead und Komplexität bei der Bereitstellung. Wollmilchsau löst dieses Problem, indem es die leistungsstarke V8-JavaScript-Engine direkt in ein Go-Binary einbettet. Dadurch lassen sich TypeScript-Tools mit nahezu nativer Performance, ohne externe Abhängigkeiten und mit minimalem Speicherverbrauch ausführen.

## Wichtigste Funktionen

- **V8 Engine Performance**: Durch die Verwendung der gleichen Engine, die auch Chrome und Node.js antreibt, stellt Wollmilchsau sicher, dass Ihre MCP-Tools so schnell wie möglich laufen.
- **Eingebettetes esbuild**: Bündelt und transpiliert Ihre TypeScript-Dateien automatisch zur Laufzeit – Sie benötigen keine komplexe Build-Pipeline mehr.
- **Nahtlose Go-Integration**: Entwickelt als skriptfähiges Herzstück für Go-basierte MCP-Server, ermöglicht es Ihnen, Tool-Logik in TypeScript zu definieren, während Sie die Robustheit von Go für den Server-Transport beibehalten.
- **Artifact-Unterstützung**: Integrierte Anbindung an den `mlcartifact`-Dienst, wodurch Ihre Skripte persistente Ergebnisse einfach erstellen und verwalten können.
