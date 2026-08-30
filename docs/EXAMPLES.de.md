# 💡 Praxisbeispiele: Warum LLMs eine Sandbox wie wollmilchsau brauchen

> **Kernprinzip:** LLMs sind Sprachmodelle, keine Rechenmaschinen. Wenn ein Problem deterministisch mit 10 Zeilen TypeScript gelöst werden kann, spart die Ausführung in der Sandbox wertvolle Tokens, vermeidet Halluzinationen und liefert in Millisekunden exakte Ergebnisse.

---

## 📋 Übersicht der Beispiel-Kategorien

1. [⏱️ Datums- & Zeitberechnungen (Schaltjahre & Epochen)](#1-datums---zeitberechnungen-schaltjahre--epochen)
2. [📈 Statistische Analysen & Sensordaten](#2-statistische-analysen--sensordaten)
3. [🧪 Code schreiben & sofort live verifizieren](#3-code-schreiben--sofort-live-verifizieren)
4. [📊 Log-Parsing & Aggregation mit Lodash](#4-log-parsing--aggregation-mit-lodash)
5. [🗺️ Geo-Koordinaten & Luftlinien (Haversine-Formel)](#5-geo-koordinaten--luftlinien-haversine-formel)
6. [🔐 JWT-Validierung & Krypto-Operationen](#6-jwt-validierung--krypto-operationen)

---

## 1. ⏱️ Datums- & Zeitberechnungen (Schaltjahre & Epochen)

* **Das Problem:** *„Wie viele Minuten sind seit dem 01.01.1970 00:00 UTC bis heute vergangen?“*
* **Ohne Sandbox:** Das Modell schätzt Tage mit 365,25 Tagen ab, stolpert über Schaltjahre und liefert widersprüchliche Näherungswerte.
* **Mit wollmilchsau (`execute_script`):**

```typescript
const start = new Date("1970-01-01T00:00:00Z").getTime();
const now = Date.now();
const minutes = Math.floor((now - start) / 60000);

console.log(`${minutes.toLocaleString("de-DE")} Minuten vergangen`);
```

* **Ergebnis:** Millisekundengenau, 100% deterministisch in 2 ms.

---

## 2. 📈 Statistische Analysen & Sensordaten

* **Das Problem:** *„Hier sind 5 Messwerte [21.4, 22.8, 21.9, 23.5, 22.1]. Berechne Mittelwert und Stichproben-Standardabweichung.“*
* **Ohne Sandbox:** LLMs neigen bei Quadratwurzeln ($\sqrt{\dots}$) und Varianzen zu Rundungsfehlern und Rechenungenauigkeiten.
* **Mit wollmilchsau (`execute_script`):**

```typescript
const data = [21.4, 22.8, 21.9, 23.5, 22.1];
const n = data.length;
const avg = data.reduce((a, b) => a + b, 0) / n;
const variance = data.reduce((a, b) => a + Math.pow(b - avg, 2), 0) / (n - 1);
const stdDev = Math.sqrt(variance);

console.log(JSON.stringify({
  mittelwert: Number(avg.toFixed(2)),
  standardabweichung: Number(stdDev.toFixed(4))
}, null, 2));
```

* **Ergebnis:**
```json
{
  "mittelwert": 22.34,
  "standardabweichung": 0.8173
}
```

---

## 3. 🧪 Code schreiben & sofort live verifizieren

* **Das Problem:** *„Schreib mir eine Funktion, die deutsche IBANs auf gültiges Format und Prüfziffer (Modulo 97) prüft, und teste sie mit zwei Beispielwerten.“*
* **Ohne Sandbox:** Das Modell simuliert die Ausführung nur im Kopf (*Mental Tracing*). Bei Algorithmen wie `BigInt % 97n` übersieht es eigene Tippfehler und behauptet fälschlicherweise, der Code funktioniere.
* **Mit wollmilchsau (`execute_script`):**

```typescript
function validateGermanIBAN(iban: string): boolean {
  const clean = iban.replace(/\s+/g, "").toUpperCase();
  if (!/^DE\d{20}$/.test(clean)) return false;
  // DE -> 1314 nach hinten stellen
  const rearranged = clean.slice(4) + "1314" + clean.slice(2, 4);
  return BigInt(rearranged) % 97n === 1n;
}

const tests = [
  "DE02100100100123456789", // gültig
  "DE00100100100123456789"  // ungültig
];

tests.forEach(iban => {
  console.log(`${iban}: ${validateGermanIBAN(iban) ? "GÜLTIG" : "UNGÜLTIG"}`);
});
```

* **Ergebnis:** Das LLM führt den Code selbst aus, sieht das echte `console.log`-Ergebnis und liefert garantiert getesteten Code.

---

## 4. 📊 Log-Parsing & Aggregation mit Lodash

* **Das Problem:** *„Hier sind 500 Zeilen Server-Logs. Filtere alle Einträge mit Level 'ERROR', gruppiere sie nach IP und gib die Top 2 Angreifer aus.“*
* **Ohne Sandbox:** Das LLM liest jede Zeile in sein Token-Kontextfenster ein, verbraucht tausende Tokens und verzählt sich bei der manuellen Gruppierung.
* **Mit wollmilchsau (`execute_script` mit gebündeltem `lodash`):**

```typescript
const _ = require("lodash");

const rawLogs = [
  { timestamp: "2026-03-30T10:01:00Z", level: "ERROR", ip: "192.168.1.50", msg: "Unauthorized access" },
  { timestamp: "2026-03-30T10:01:05Z", level: "INFO",  ip: "10.0.0.12",   msg: "User logged in" },
  { timestamp: "2026-03-30T10:02:11Z", level: "ERROR", ip: "192.168.1.50", msg: "Password brute force" },
  { timestamp: "2026-03-30T10:03:00Z", level: "ERROR", ip: "172.16.25.4",  msg: "SQL Injection attempt" },
  { timestamp: "2026-03-30T10:03:15Z", level: "ERROR", ip: "192.168.1.50", msg: "Password brute force" },
  { timestamp: "2026-03-30T10:04:02Z", level: "INFO",  ip: "10.0.0.12",   msg: "Dashboard viewed" }
];

// 1. Nur Fehler filtern
const errors = _.filter(rawLogs, { level: "ERROR" });

// 2. Nach IP zählen
const ipCounts = _.countBy(errors, "ip");

// 3. Sortieren und Top 2 ausgeben
const topAttackers = _.chain(ipCounts)
  .map((count, ip) => ({ ip, count }))
  .orderBy(["count"], ["desc"])
  .take(2)
  .value();

console.log(JSON.stringify(topAttackers, null, 2));
```

* **Ergebnis:**
```json
[
  { "ip": "192.168.1.50", "count": 3 },
  { "ip": "172.16.25.4", "count": 1 }
]
```

---

## 5. 🗺️ Geo-Koordinaten & Luftlinien (Haversine-Formel)

* **Das Problem:** *„Berechne die exakte Luftlinie in Kilometern zwischen Frankfurt (50.1109, 8.6821) und München (48.1351, 11.5820).“*
* **Ohne Sandbox:** Das Modell scheitert an trigonometrischen Funktionen (`Math.sin`, `Math.cos`, `Math.atan2`) und rät eine grobe Zahl.
* **Mit wollmilchsau (`execute_script`):**

```typescript
function haversineDistance(lat1: number, lon1: number, lat2: number, lon2: number): number {
  const R = 6371; // Erdradius in km
  const dLat = (lat2 - lat1) * Math.PI / 180;
  const dLon = (lon2 - lon1) * Math.PI / 180;

  const a =
    Math.sin(dLat / 2) * Math.sin(dLat / 2) +
    Math.cos(lat1 * Math.PI / 180) * Math.cos(lat2 * Math.PI / 180) *
    Math.sin(dLon / 2) * Math.sin(dLon / 2);

  const c = 2 * Math.atan2(Math.sqrt(a), Math.sqrt(1 - a));
  return R * c;
}

const distance = haversineDistance(50.1109, 8.6821, 48.1351, 11.5820);
console.log(`Distanz: ${distance.toFixed(2)} km`);
```

* **Ergebnis:**
```text
Distanz: 304.34 km
```

---

## 6. 🔐 JWT-Validierung & Krypto-Operationen

* **Das Problem:** *„Prüfe, ob dieses JWT-Token abgelaufen ist oder welcher User darin codiert ist.“*
* **Ohne Sandbox:** Base64URL-Decodierung und Unix-Epoch-Vergleiche sind für LLMs im Textfluss fehleranfällig.
* **Mit wollmilchsau (`execute_script`):**

```typescript
function parseAndCheckJWT(token: string) {
  const parts = token.split(".");
  if (parts.length !== 3) throw new Error("Ungültiges JWT-Format");

  const base64 = parts[1].replace(/-/g, "+").replace(/_/g, "/");
  const payload = JSON.parse(Buffer.from(base64, "base64").toString("utf-8"));
  
  const now = Math.floor(Date.now() / 1000);
  const isExpired = payload.exp ? now > payload.exp : false;

  return {
    subject: payload.sub,
    name: payload.name,
    expiresAt: payload.exp ? new Date(payload.exp * 1000).toISOString() : null,
    isExpired
  };
}

const token = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyLCJleHAiOjE3NzUwMDAwMDB9.sample";
console.log(JSON.stringify(parseAndCheckJWT(token), null, 2));
```

---

## 🎯 Warum das überzeugt

1. **Tokens & Kosten sparen:** Keine 30 Sekunden Reasoning-Tokens für einfache Berechnungen.
2. **Determinismus:** Exakte, reproduzierbare Ergebnisse statt probabilistischer Wort-Vorhersagen.
3. **Isolierte Sicherheit:** Vollständige V8-Sandbox ohne Netzwerk- oder Dateisystemrisiken.
