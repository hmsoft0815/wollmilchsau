# --- STAGE 1: Build ---
FROM golang:1.24-bookworm AS builder

# Build-Abhängigkeiten installieren (für CGO / V8 erforderlich)
RUN apt-get update && apt-get install -y \
    gcc \
    g++ \
    make \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Abhängigkeiten zuerst kopieren für besseres Caching
COPY go.mod go.sum ./
RUN go mod download

# Quellcode kopieren
COPY . .

# Binary bauen (CGO muss an sein für v8go)
RUN CGO_ENABLED=1 go build -trimpath -ldflags="-s -w" -o wollmilchsau ./cmd/main.go

# --- STAGE 2: Run ---
FROM debian:bookworm-slim

# Wichtige Laufzeit-Bibliotheken für V8 und CA-Zertifikate installieren
RUN apt-get update && apt-get install -y \
    ca-certificates \
    libc6 \
    libstdc++6 \
    && rm -rf /var/lib/apt/lists/*

# User für Sicherheit (kein root)
RUN useradd -m appuser
USER appuser
WORKDIR /home/appuser

# Nur das fertige Binary aus dem Builder kopieren
COPY --from=builder /app/wollmilchsau /usr/local/bin/wollmilchsau

# Umgebungsvariable für den Port (Standard: :8000)
ENV ADDR=:8000
EXPOSE 8000

# Einstiegspunkt
ENTRYPOINT ["wollmilchsau"]
CMD ["--addr", ":8000"]
