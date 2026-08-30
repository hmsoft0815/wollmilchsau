// Copyright (c) 2026 Michael Lechner. All rights reserved.
package executor

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/hmsoft0815/wollmilchsau/internal/bundler"
	"github.com/hmsoft0815/wollmilchsau/internal/parser"
)

func TestSandboxIsolation_NoGlobals(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tests := []struct {
		name string
		code string
	}{
		{
			name: "No fetch",
			code: `if (typeof fetch !== "undefined") throw new Error("fetch should be undefined");`,
		},
		{
			name: "No XMLHttpRequest",
			code: `if (typeof XMLHttpRequest !== "undefined") throw new Error("XMLHttpRequest should be undefined");`,
		},
		{
			name: "No process",
			code: `if (typeof process !== "undefined") throw new Error("process should be undefined");`,
		},
		{
			name: "No setTimeout",
			code: `if (typeof setTimeout !== "undefined") throw new Error("setTimeout should be undefined");`,
		},
		{
			name: "No setInterval",
			code: `if (typeof setInterval !== "undefined") throw new Error("setInterval should be undefined");`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := Execute(ctx, tt.code, "test.js", nil, "")
			if !res.Success {
				t.Fatalf("isolation check failed: %s (stderr: %s)", res.Summary, res.Stderr)
			}
		})
	}
}

func TestExecute_SourceMapRuntimeErrorMapping(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	plan := &parser.ExecutionPlan{
		EntryPoint: "main.ts",
		Files: []parser.VirtualFile{
			{
				Name: "main.ts",
				Content: `// line 1
// line 2
function blowUp(): void {
    throw new Error("kaboom from typescript");
}
blowUp();`,
			},
		},
	}

	bundle, err := bundler.Bundle(plan)
	if err != nil {
		t.Fatalf("bundle failed: %v", err)
	}

	res := Execute(ctx, bundle.JS, plan.EntryPoint, bundle.SourceMap, "")
	if res.Success {
		t.Fatal("expected runtime error")
	}
	if !strings.Contains(res.Summary, "kaboom from typescript") {
		t.Errorf("expected error message in summary, got: %s", res.Summary)
	}
	if len(res.Diagnostics) == 0 {
		t.Fatal("expected at least one diagnostic")
	}

	diag := res.Diagnostics[0]
	if diag.Source != "main.ts" {
		t.Errorf("expected diagnostic source 'main.ts', got %q", diag.Source)
	}
	// Line 4 is throw new Error("kaboom from typescript")
	if diag.Line != 4 {
		t.Errorf("expected line 4, got line %d", diag.Line)
	}
}
