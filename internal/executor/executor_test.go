// Copyright (c) 2026 Michael Lechner. All rights reserved.
package executor

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestExecute_Polyfills(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tests := []struct {
		name     string
		code     string
		contains string
	}{
		{
			name: "performance.now",
			code: `
				const t1 = performance.now();
				for(let i=0; i<1000; i++);
				const t2 = performance.now();
				console.log('perf_ok:', t2 > t1 && typeof t1 === 'number');
			`,
			contains: "perf_ok: true",
		},
		{
			name: "atob and btoa",
			code: `
				const encoded = btoa("Hello World");
				const decoded = atob(encoded);
				console.log('b64_ok:', encoded === "SGVsbG8gV29ybGQ=" && decoded === "Hello World");
			`,
			contains: "b64_ok: true",
		},
		{
			name: "crypto.getRandomValues",
			code: `
				const arr = new Uint8Array(16);
				crypto.getRandomValues(arr);
				const sum = arr.reduce((a, b) => a + b, 0);
				// It's statistically very unlikely (1 in 2^128) that 16 random bytes sum to 0.
				console.log('crypto_ok:', arr.length === 16 && sum > 0);
			`,
			contains: "crypto_ok: true",
		},
		{
			name: "TextEncoder and TextDecoder",
			code: `
				const encoder = new TextEncoder();
				const decoder = new TextDecoder();
				const str = "Wollmilchsau 🚀";
				const buf = encoder.encode(str);
				const out = decoder.decode(buf);
				console.log('text_ok:', out === str && buf instanceof Uint8Array);
			`,
			contains: "text_ok: true",
		},
		{
			name: "Buffer minimal API",
			code: `
				const b1 = Buffer.from("SGVsbG8=", "base64");
				const b2 = Buffer.from("World");
				const b3 = Buffer.alloc(5);
				console.log('buffer_ok:', b1 instanceof Uint8Array && b1.length === 5 && b2.length === 5 && b3.length === 5);
			`,
			contains: "buffer_ok: true",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := Execute(ctx, tt.code, "test.js", nil)
			if !res.Success {
				t.Errorf("Execution failed: %s\nStderr: %s", res.Summary, res.Stderr)
				return
			}
			if !strings.Contains(res.Stdout, tt.contains) {
				t.Errorf("Expected output to contain %q, got %q", tt.contains, res.Stdout)
			}
		})
	}
}
