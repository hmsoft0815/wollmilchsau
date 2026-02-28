// Copyright (c) 2026 Michael Lechner. All rights reserved.
package bundler

import (
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"

	"github.com/hmsoft0815/wollmilchsau/internal/parser"
)

func TestExtractSourceMap(t *testing.T) {
	// Construct a mock JS bundle with an inline source map.
	content := map[string]any{
		"version":  3,
		"sources":  []string{"main.ts"},
		"mappings": "AAAA",
	}
	mapJSON, _ := json.Marshal(content)
	encoded := base64.StdEncoding.EncodeToString(mapJSON)
	js := "console.log('hello');\n//# sourceMappingURL=data:application/json;base64," + encoded

	cleanJS, sm, err := extractSourceMap(js)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cleanJS != "console.log('hello');" {
		t.Errorf("expected clean JS 'console.log(\\'hello\\');', got %q", cleanJS)
	}

	if sm == nil {
		t.Fatal("expected source map, got nil")
	}

	// Verify the source map content.
	pos := sm.Resolve(1, 1)
	if pos == nil {
		t.Fatal("expected resolved position, got nil")
	}
	if pos.Source != "main.ts" {
		t.Errorf("expected source 'main.ts', got %q", pos.Source)
	}
}

func TestExtractSourceMap_Missing(t *testing.T) {
	js := "console.log('no map here');"
	_, sm, err := extractSourceMap(js)
	if err == nil {
		t.Error("expected error for missing source map")
	}
	if sm != nil {
		t.Errorf("expected nil source map, got %+v", sm)
	}
}

func TestBundle_SourceMap(t *testing.T) {
	// Simple execution plan with one TS file.
	plan := &parser.ExecutionPlan{
		Files: []parser.VirtualFile{
			{
				Name:    "main.ts",
				Content: "const x: number = 42;\nconsole.log(x);",
			},
		},
		EntryPoint: "main.ts",
		TimeoutMs:  1000,
	}

	result, err := Bundle(plan)
	if err != nil {
		t.Fatalf("bundle failed: %v", err)
	}

	if result.SourceMap == nil {
		t.Fatal("bundle result missing source map")
	}

	// The source map should be able to resolve back to main.ts.
	// Since we use api.FormatIIFE and GlobalName "__entry__",
	// the bundled JS will have some wrapper code.

	// Let's check if we can resolve at least something in main.ts.
	found := false
	lines := strings.Split(result.JS, "\n")
	for i := range lines {
		// Try resolving multiple columns on each line to find any mapping to main.ts
		for col := 1; col <= 80; col += 4 {
			pos := result.SourceMap.Resolve(i+1, col)
			if pos != nil && pos.Source == "main.ts" {
				found = true
				break
			}
		}
		if found {
			break
		}
	}

	if !found {
		t.Error("could not resolve any position back to 'main.ts' in bundled JS")
	}
}
