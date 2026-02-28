// Copyright (c) 2026 Michael Lechner. All rights reserved.
// Package parser handles validation of execution plans and virtual files.
package parser

import (
	"fmt"
	"strings"
)

// ValidatePlan ensures the ExecutionPlan is logically sound.
func ValidatePlan(plan *ExecutionPlan) error {
	if len(plan.Files) == 0 {
		return &ParseError{Message: "at least one file required"}
	}

	if plan.EntryPoint == "" {
		return &ParseError{Message: "entry point must not be empty"}
	}

	found := false
	for _, f := range plan.Files {
		if f.Name == "" {
			return &ParseError{Message: "file name must not be empty"}
		}
		if err := validateFilename(f.Name); err != nil {
			return &ParseError{Message: err.Error()}
		}
		if f.Name == plan.EntryPoint {
			found = true
		}
	}

	if !found {
		return &ParseError{
			Message: fmt.Sprintf("entry point %q not found in provided files", plan.EntryPoint),
		}
	}

	// Clamp timeout
	if plan.TimeoutMs < 100 {
		plan.TimeoutMs = 100
	}
	if plan.TimeoutMs > 30_000 {
		plan.TimeoutMs = 30_000
	}

	return nil
}

// validateFilename rejects path traversal and absolute paths.
func validateFilename(name string) error {
	if strings.HasPrefix(name, "/") {
		return fmt.Errorf("file name must not be absolute: %q", name)
	}
	if strings.Contains(name, "..") {
		return fmt.Errorf("file name must not contain '..': %q", name)
	}
	if strings.ContainsAny(name, "\x00\\") {
		return fmt.Errorf("invalid character in file name: %q", name)
	}
	return nil
}
