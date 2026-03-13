#!/bin/bash
# scripts/test_artifacts_integration.sh
# Automatisiert den Integrationstest zwischen wollmilchsau und mlcartifact.

set -e

# Konfiguration
ARTIFACT_DIR="../mlcartifact"
ARTIFACT_BIN="$ARTIFACT_DIR/bin/artifact-server"
TEST_GRPC_PORT="19590"
TEST_SSE_PORT="18082"
TEST_GRPC_ADDR="localhost:$TEST_GRPC_PORT"

echo "🚀 Starte Vorbereitung für Integrationstest..."

# 1. Sicherstellen, dass der Artifact Server gebaut ist
if [ ! -f "$ARTIFACT_BIN" ]; then
    echo "📦 Baue mlcartifact Server..."
    (cd "$ARTIFACT_DIR" && make build-server)
fi

# 2. Artifact Server im Hintergrund starten
echo "📡 Starte mlcartifact auf gRPC Port $TEST_GRPC_PORT (SSE: $TEST_SSE_PORT)..."
$ARTIFACT_BIN -grpc-addr ":$TEST_GRPC_PORT" -addr ":$TEST_SSE_PORT" > artifact_test_server.log 2>&1 &
SERVER_PID=$!

# Trap um den Server am Ende IMMER zu beenden
trap "echo '🛑 Stoppe Artifact Server (PID: $SERVER_PID)...'; kill $SERVER_PID || true" EXIT

# 3. Warten bis der Server bereit ist (10s)
echo "⏳ Warte 10s auf Server-Initialisierung..."
sleep 10

# 4. Integrationstest ausführen
echo "🧪 Starte wollmilchsau Integrationstests..."
# Wir setzen ARTIFACT_GRPC_ADDR für den Test-Client
export ARTIFACT_GRPC_ADDR="$TEST_GRPC_ADDR"

# Wir führen den Test aus und fangen den Exit-Code ab
set +e
go test -v -run TestArtifactIntegration ./internal/executor
TEST_EXIT_CODE=$?
set -e

if [ $TEST_EXIT_CODE -eq 0 ]; then
    echo "✅ Integrationstest ERFOLGREICH"
else
    echo "❌ Integrationstest FEHLGESCHLAGEN"
    echo "--- Server Logs ---"
    tail -n 20 artifact_test_server.log
fi

exit $TEST_EXIT_CODE
