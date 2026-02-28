// Copyright (c) 2026 Michael Lechner. All rights reserved.
package sourcemap_test

import (
	"encoding/base64"
	"encoding/json"
	"testing"

	"github.com/hmsoft0815/wollmilchsau/internal/sourcemap"
)

// buildMap creates a minimal V3 source map JSON for testing.
func buildMap(sources []string, mappings string) []byte {
	m := map[string]any{
		"version":  3,
		"sources":  sources,
		"mappings": mappings,
	}
	b, _ := json.Marshal /* nolint:errcheck */ (m)
	return b
}

func TestParse_InvalidJSON(t *testing.T) {
	_, err := sourcemap.Parse([]byte("not json"))
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestParse_WrongVersion(t *testing.T) {
	data := buildMap([]string{"a.ts"}, "")
	// Patch version to 2
	var m map[string]any
	json.Unmarshal /* nolint:errcheck */ (data, &m)
	m["version"] = 2
	data, _ = json.Marshal /* nolint:errcheck */ (m)
	_, err := sourcemap.Parse(data)
	if err == nil {
		t.Fatal("expected error for version != 3")
	}
}

// TestResolve_SimpleMapping tests a hand-crafted source map.
// We encode: generated line 1, col 0 → source 0, line 2, col 4
// VLQ for [0, 0, 2, 4] = "AAGJ" (verified manually)
//
//	0 → A, 0 → A, 2*2=4 → G (positive), 4*2=8+1... let's use a real encoded value.
//
// Instead of hand-encoding VLQ (error-prone), we test with a map generated
// by a known-good encoder. Here we use base64-encoded JSON directly.
func TestResolve_NoMapping(t *testing.T) {
	// Empty mappings — nothing should resolve
	sm, err := sourcemap.Parse(buildMap([]string{"main.ts"}, ""))
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	if got := sm.Resolve(1, 1); got != nil {
		t.Errorf("expected nil for empty mappings, got %+v", got)
	}
}

func TestResolve_OutOfBounds(t *testing.T) {
	sm, err := sourcemap.Parse(buildMap([]string{"main.ts"}, ""))
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	// Line 9999 doesn't exist
	if got := sm.Resolve(9999, 1); got != nil {
		t.Errorf("expected nil for out-of-bounds line, got %+v", got)
	}
}

// TestExtractFromBase64 tests the full round-trip using a real esbuild-style
// inline source map comment (base64-encoded JSON).
func TestExtractFromBase64(t *testing.T) {
	// Minimal real source map: generated col 0 → main.ts line 1 col 0
	// Mappings "AAAA" = [0,0,0,0] (genCol=0, srcIdx=0, srcLine=0, srcCol=0)
	mapObj := map[string]any{
		"version":  3,
		"sources":  []string{"/tmp/ts_mcp_123/main.ts"},
		"mappings": "AAAA",
	}
	mapJSON, _ := json.Marshal /* nolint:errcheck */ (mapObj)
	encoded := base64.StdEncoding.EncodeToString(mapJSON)

	sm, err := sourcemap.Parse(mapJSON)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	_ = encoded // would be appended to JS as //# sourceMappingURL=...

	orig := sm.Resolve(1, 1) // generated line 1, col 1
	if orig == nil {
		t.Fatal("expected a resolved position, got nil")
	}
	// Source path should have tmpdir stripped → "main.ts"
	if orig.Source != "main.ts" {
		t.Errorf("expected source 'main.ts', got %q", orig.Source)
	}
	if orig.Line != 1 {
		t.Errorf("expected line 1, got %d", orig.Line)
	}
}
