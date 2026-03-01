// Copyright (c) 2026 Michael Lechner. All rights reserved.
package parser

import (
	"testing"
)

const mainTS = "main.ts"

func TestValidatePlanSuccess(t *testing.T) {
	plan := &ExecutionPlan{
		Files: []VirtualFile{
			{Name: "a.ts", Content: "export const a = 1;"},
			{Name: mainTS, Content: "import { a } from './a'; console.log(a);"},
		},
		EntryPoint: mainTS,
		TimeoutMs:  5000,
	}

	err := ValidatePlan(plan)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidatePlanMissingEntryPoint(t *testing.T) {
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

func TestValidatePlanEmptyFiles(t *testing.T) {
	plan := &ExecutionPlan{
		Files:      []VirtualFile{},
		EntryPoint: mainTS,
	}
	err := ValidatePlan(plan)
	if err == nil {
		t.Fatal("expected error for empty files")
	}
}

func TestValidateFilename(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{"main.ts", false},
		{"src/utils.ts", false},
		{"_hidden.js", false},
		{"my-file_123.test.ts", false},
		{"", true},
		{"/abs/path.ts", true},
		{"\\win\\path.ts", true},
		{"../traversal.ts", true},
		{"src/../traversal.ts", true},
		{"./start.ts", true},
		{"end/", true},
		{"dots.", true},
		{"space in name.ts", true},
		{"special!@#.ts", true},
		{"double//slash.ts", true},
		{"nul", true},
		{"aux.ts", true},
		{"com1/file.ts", true},
		{"CON.js", true},
		{"LPT9", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateFilename(tt.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateFilename(%q) error = %v, wantErr %v", tt.name, err, tt.wantErr)
			}
		})
	}
}
