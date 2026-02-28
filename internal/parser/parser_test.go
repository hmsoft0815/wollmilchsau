// Copyright (c) 2026 Michael Lechner. All rights reserved.
package parser

import (
	"testing"
)

func TestValidatePlan_Success(t *testing.T) {
	plan := &ExecutionPlan{
		Files: []VirtualFile{
			{Name: "a.ts", Content: "export const a = 1;"},
			{Name: "main.ts", Content: "import { a } from './a'; console.log(a);"},
		},
		EntryPoint: "main.ts",
		TimeoutMs:  5000,
	}

	err := ValidatePlan(plan)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidatePlan_MissingEntryPoint(t *testing.T) {
	plan := &ExecutionPlan{
		Files: []VirtualFile{
			{Name: "a.ts", Content: "1"},
		},
		EntryPoint: "missing.ts",
	}
	err := ValidatePlan(plan)
	if err == nil {
		t.Fatal("expected error for missing entry point")
	}
}

func TestValidatePlan_EmptyFiles(t *testing.T) {
	plan := &ExecutionPlan{
		Files:      []VirtualFile{},
		EntryPoint: "main.ts",
	}
	err := ValidatePlan(plan)
	if err == nil {
		t.Fatal("expected error for empty files")
	}
}

func TestValidatePlan_InvalidFilename(t *testing.T) {
	plan := &ExecutionPlan{
		Files: []VirtualFile{
			{Name: "../secret.ts", Content: "1"},
		},
		EntryPoint: "../secret.ts",
	}
	err := ValidatePlan(plan)
	if err == nil {
		t.Fatal("expected error for path traversal in filename")
	}
}
