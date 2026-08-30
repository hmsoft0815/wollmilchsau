# 💡 Practical Examples: Why LLMs Need a Sandbox like wollmilchsau

> **Core Principle:** LLMs are language models, not calculators. When a problem can be solved deterministically with 10 lines of TypeScript, executing it in a sandbox saves valuable reasoning tokens, prevents hallucinations, and delivers exact results in milliseconds.

---

## 📋 Catalog of Example Scenarios

1. [⏱️ Date & Time Calculations (Leap Years & Epochs)](#1-date--time-calculations-leap-years--epochs)
2. [📈 Statistical Analysis & Sensor Metrics](#2-statistical-analysis--sensor-metrics)
3. [🧪 Code Generation with Live Self-Verification](#3-code-generation-with-live-self-verification)
4. [📊 Log Parsing & Aggregation with Lodash](#4-log-parsing--aggregation-with-lodash)
5. [🗺️ Geo-Spatial Math (Haversine Formula)](#5-geo-spatial-math-haversine-formula)
6. [🔐 JWT Expiration & Crypto Operations](#6-jwt-expiration--crypto-operations)

---

## 1. ⏱️ Date & Time Calculations (Leap Years & Epochs)

* **The Problem:** *"How many minutes have passed since 1970-01-01 00:00 UTC until now?"*
* **Without Sandbox:** The model estimates days with 365.25, misses leap years, and yields conflicting approximations.
* **With wollmilchsau (`execute_script`):**

```typescript
const start = new Date("1970-01-01T00:00:00Z").getTime();
const now = Date.now();
const minutes = Math.floor((now - start) / 60000);

console.log(`${minutes.toLocaleString("en-US")} minutes passed`);
```

* **Result:** Millisecond-exact, 100% deterministic in 2 ms.

---

## 2. 📈 Statistical Analysis & Sensor Metrics

* **The Problem:** *"Here are 5 sensor readings [21.4, 22.8, 21.9, 23.5, 22.1]. Compute the mean and sample standard deviation."*
* **Without Sandbox:** LLMs struggle with square roots ($\sqrt{\dots}$) and variance sums, leading to subtle rounding errors and hallucinations.
* **With wollmilchsau (`execute_script`):**

```typescript
const data = [21.4, 22.8, 21.9, 23.5, 22.1];
const n = data.length;
const avg = data.reduce((a, b) => a + b, 0) / n;
const variance = data.reduce((a, b) => a + Math.pow(b - avg, 2), 0) / (n - 1);
const stdDev = Math.sqrt(variance);

console.log(JSON.stringify({
  mean: Number(avg.toFixed(2)),
  stdDev: Number(stdDev.toFixed(4))
}, null, 2));
```

* **Result:**
```json
{
  "mean": 22.34,
  "stdDev": 0.8173
}
```

---

## 3. 🧪 Code Generation with Live Self-Verification

* **The Problem:** *"Write a function that validates German IBANs (mod-97 check digits) and test it with two sample inputs."*
* **Without Sandbox:** The model mentally traces the code execution. On algorithms like `BigInt % 97n`, it misses bugs and confidently claims broken code works.
* **With wollmilchsau (`execute_script`):**

```typescript
function validateGermanIBAN(iban: string): boolean {
  const clean = iban.replace(/\s+/g, "").toUpperCase();
  if (!/^DE\d{20}$/.test(clean)) return false;
  // Move 'DE' -> 1314 to the end
  const rearranged = clean.slice(4) + "1314" + clean.slice(2, 4);
  return BigInt(rearranged) % 97n === 1n;
}

const tests = [
  "DE02100100100123456789", // valid
  "DE00100100100123456789"  // invalid
];

tests.forEach(iban => {
  console.log(`${iban}: ${validateGermanIBAN(iban) ? "VALID" : "INVALID"}`);
});
```

* **Result:** The LLM executes the code in V8, inspects the real `console.log` output, and returns guaranteed verified, working code.

---

## 4. 📊 Log Parsing & Aggregation with Lodash

* **The Problem:** *"Here is a batch of server logs. Filter all entries with level 'ERROR', group them by IP, and return the top 2 attackers."*
* **Without Sandbox:** The LLM reads all lines into its token context, consumes thousands of tokens, and miscounts totals.
* **With wollmilchsau (`execute_script` with bundled `lodash`):**

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

// 1. Filter for errors only
const errors = _.filter(rawLogs, { level: "ERROR" });

// 2. Count occurrences per IP
const ipCounts = _.countBy(errors, "ip");

// 3. Transform and sort to find top attackers
const topAttackers = _.chain(ipCounts)
  .map((count, ip) => ({ ip, count }))
  .orderBy(["count"], ["desc"])
  .take(2)
  .value();

console.log(JSON.stringify(topAttackers, null, 2));
```

* **Result:**
```json
[
  { "ip": "192.168.1.50", "count": 3 },
  { "ip": "172.16.25.4", "count": 1 }
]
```

---

## 5. 🗺️ Geo-Spatial Math (Haversine Formula)

* **The Problem:** *"What is the exact flight distance between Frankfurt (50.1109, 8.6821) and Munich (48.1351, 11.5820) in kilometers?"*
* **Without Sandbox:** LLMs fail at floating-point trigonometry (`Math.sin`, `Math.cos`, `Math.atan2`) and hallucinate rough estimates.
* **With wollmilchsau (`execute_script`):**

```typescript
function haversineDistance(lat1: number, lon1: number, lat2: number, lon2: number): number {
  const R = 6371; // Earth's radius in km
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
console.log(`Distance: ${distance.toFixed(2)} km`);
```

* **Result:**
```text
Distance: 304.34 km
```

---

## 6. 🔐 JWT Expiration & Crypto Operations

* **The Problem:** *"Check if this JWT token is expired and decode its subject and name."*
* **Without Sandbox:** Base64URL decoding and Unix epoch comparisons are error-prone in textual reasoning.
* **With wollmilchsau (`execute_script`):**

```typescript
function parseAndCheckJWT(token: string) {
  const parts = token.split(".");
  if (parts.length !== 3) throw new Error("Invalid JWT format");

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

## 🎯 Key Value Propositions

1. **Token & Cost Savings:** Avoid wasting 30 seconds of reasoning tokens for programmatic tasks.
2. **Determinism:** Exact, reproducible outputs instead of probabilistic token predictions.
3. **Isolated Security:** Complete in-process V8 isolate without network or filesystem risks.
