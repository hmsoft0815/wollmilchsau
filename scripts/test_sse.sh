#!/bin/bash
# Copyright (c) 2026 Michael Lechner. All rights reserved.

PORT=8081
ADDR=":$PORT"
URL="http://localhost:$PORT"
LOG_DIR="./test_logs"

rm -rf "$LOG_DIR"
mkdir -p "$LOG_DIR"

echo "🔨 Building server..."
make build > /dev/null

echo "🚀 Starting server in SSE mode with logging enabled..."
./build/wollmilchsau -addr "$ADDR" -log-dir "$LOG_DIR" &
SERVER_PID=$!

trap "kill $SERVER_PID 2>/dev/null" EXIT

sleep 3

echo "📡 Connecting to SSE stream..."
SSE_LOG=$(mktemp)
curl -s -N "$URL/sse" > "$SSE_LOG" &
CURL_PID=$!

sleep 3
ENDPOINT_URL=$(grep "data: " "$SSE_LOG" | head -n 1 | sed 's/data: //' | tr -d '\r')

if [ -z "$ENDPOINT_URL" ]; then
    echo "❌ Failed to get SSE endpoint."
    kill $CURL_PID 2>/dev/null
    exit 1
fi

echo "🔑 Initializing..."
curl -s -X POST "$ENDPOINT_URL" -H "Content-Type: application/json" \
     -d '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test","version":"1"}}}' > /dev/null

echo "🧪 Calling 'execute_script'..."
curl -s -X POST "$ENDPOINT_URL" -H "Content-Type: application/json" \
     -d '{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"execute_script","arguments":{"code":"console.log(\"Log Test\");"}}}' > /dev/null

echo "🧪 Calling 'execute_project' (Multi-File)..."
curl -s -X POST "$ENDPOINT_URL" -H "Content-Type: application/json" \
     -d '{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"execute_project","arguments":{"files":[{"name":"main.ts","content":"import { greet } from \"./utils\";\nconsole.log(greet(\"World\"));"},{"name":"utils.ts","content":"export function greet(n: string) { return `Hello ${n}!`; }"}],"entryPoint":"main.ts"}}}' > /dev/null

echo "📥 Waiting for result..."
sleep 3

echo "📁 Checking Log Directory..."
ls -l "$LOG_DIR"

ZIP_COUNT=$(ls "$LOG_DIR"/*.zip 2>/dev/null | wc -l)

if [ "$ZIP_COUNT" -gt 0 ]; then
    echo "🎉 SUCCESS: $ZIP_COUNT ZIP log file(s) created!"
    
    ZIP_FILE=$(ls "$LOG_DIR"/*.zip | head -n 1)
    echo "📦 Contents of $ZIP_FILE:"
    unzip -l "$ZIP_FILE"
else
    echo "❌ FAILURE: No ZIP log files found."
    exit 1
fi

kill $CURL_PID 2>/dev/null
# rm "$SSE_LOG"
