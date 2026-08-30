// Copyright (c) 2026 Michael Lechner. All rights reserved.
package executor

import (
	"context"
	"strings"
	"testing"
	"time"
)

// TestPolyfillWithoutArtifactService verifies that without a real artifact server,
// core polyfills still work and Execute succeeds.
func TestPolyfillWithoutArtifactService(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	res := Execute(ctx, `
const b64 = btoa("test") === "dGVzdA==";
var arr = new Uint8Array(16);
crypto.getRandomValues(arr);
var encOk = TextEncoder && !isNaN(TextEncoder.prototype.encode.call(new TextEncoder(), "x").length);
var decOk = TextDecoder;
console.log(b64 + "|" + arr.length + "|" + (encOk ? 1 : 0) + "|" + (decOk ? 1 : 0));
`, "test.js", nil, "")

	if !res.Success {
		t.Fatalf("execution should succeed without artifact service: %v; stderr=%q", res.Summary, res.Stderr)
	}
	if res.ExitCode != 0 {
		t.Errorf("expected exit code 0, got %d", res.ExitCode)
	}
	if !strings.Contains(res.Stdout, "true|16|1|1") {
		t.Errorf("core polyfills should still work; stdout=%q", res.Stdout)
	}
	if len(res.CreatedArtifacts) != 0 {
		t.Errorf("expected 0 artifacts without a server, got %d", len(res.CreatedArtifacts))
	}
}

// TestExecuteTimeoutViaContext verifies that context cancellation is detected.
func TestExecuteTimeoutViaContext(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	res := Execute(ctx, `while (true) {}`, "test.js", nil, "")

	if res.Success {
		t.Errorf("expected execution to fail (timeout)")
	}
	if res.ExitCode != 124 {
		t.Errorf("expected exit code 124 for timeout, got %d", res.ExitCode)
	}
	if !strings.Contains(res.Summary, "timed out") {
		t.Errorf("expected 'timed out' in summary, got %q", res.Summary)
	}
}

// TestExecuteRuntimeError verifies runtime errors are reported correctly.
func TestExecuteRuntimeError(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	res := Execute(ctx, `throw new Error("test error");`, "test.js", nil, "")

	if res.Success {
		t.Errorf("expected execution to fail")
	}
	if res.ExitCode != 1 {
		t.Errorf("expected exit code 1, got %d", res.ExitCode)
	}
	if !strings.Contains(res.Summary, "Runtime Error") {
		t.Errorf("expected 'Runtime Error' in summary, got %q", res.Summary)
	}
	if len(res.Diagnostics) == 0 {
		t.Error("expected diagnostics for runtime error")
	}
	if len(res.Diagnostics) > 0 && res.Diagnostics[0].Severity != "error" {
		t.Errorf("expected severity 'error', got %q", res.Diagnostics[0].Severity)
	}
}

// TestExecuteSuccess verifies successful code execution.
func TestExecuteSuccess(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	res := Execute(ctx, `
const a = 1 + 2;
console.log("result=" + a);
console.warn("warn msg");
console.error("err msg");
`, "test.js", nil, "")

	if !res.Success {
		t.Fatalf("unexpected failure: %v", res.Summary)
	}
	if res.ExitCode != 0 {
		t.Errorf("expected exit code 0, got %d", res.ExitCode)
	}
	if !strings.Contains(res.Stdout, "result=3") {
		t.Errorf("expected 'result=3' in stdout, got %q", res.Stdout)
	}
	if !strings.Contains(res.Stderr, "warn msg") {
		t.Errorf("expected 'warn msg' in stderr, got %q", res.Stderr)
	}
	if len(res.Diagnostics) != 0 {
		t.Errorf("expected no diagnostics, got %v", res.Diagnostics)
	}
}

// TestExecuteDurationMs verifies the duration metric is nonzero.
func TestExecuteDurationMs(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	res := Execute(ctx, `console.log("done");`, "test.js", nil, "")

	if !res.Success {
		t.Fatalf("unexpected failure: %v", res.Summary)
	}
	if res.DurationMs <= 0 {
		t.Errorf("expected positive DurationMs, got %d", res.DurationMs)
	}
}
