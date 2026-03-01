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

// validateFilename rejects path traversal, absolute paths and invalid characters.
func validateFilename(name string) error {
	if name == "" {
		return fmt.Errorf("file name must not be empty")
	}

	// 1. Basic path traversal & absolute path checks
	if strings.HasPrefix(name, "/") || strings.HasPrefix(name, "\\") {
		return fmt.Errorf("file name must not be absolute: %q", name)
	}
	if strings.Contains(name, "..") {
		return fmt.Errorf("file name must not contain '..': %q", name)
	}

	// 2. Character whitelist: [a-zA-Z0-9._/-]
	// 3. Structural issues
	if err := checkSyntaxAndChars(name); err != nil {
		return err
	}

	// 4. Windows reserved names (case-insensitive, all segments)
	return checkReservedNames(name)
}

// now with check for allowed chars - which should never be a problem unless
// someone tries to be clever or using chinese chars or whatever
func checkSyntaxAndChars(name string) error {
	for _, r := range name {
		if !strings.ContainsRune(onlyAllowedChars, r) {
			return fmt.Errorf("invalid character %q in file name: %q (only alphanumeric, '.', '_', '-', '/' allowed)", r, name)
		}
	}

	if strings.Contains(name, "//") {
		return fmt.Errorf("file name must not contain consecutive slashes: %q", name)
	}
	if strings.HasPrefix(name, "./") || strings.HasSuffix(name, "/") || strings.HasSuffix(name, ".") {
		return fmt.Errorf("invalid file name structure: %q", name)
	}
	return nil
}

func checkReservedNames(name string) error {
	segments := strings.Split(name, "/")
	for _, seg := range segments {
		base := strings.ToUpper(seg)
		if dotIdx := strings.Index(base, "."); dotIdx != -1 {
			base = base[:dotIdx]
		}

		if isReserved(base) {
			return fmt.Errorf("file name contains reserved system name: %q", seg)
		}
	}
	return nil
}

func isReserved(base string) bool {
	switch base {
	case "CON", "PRN", "AUX", "NUL":
		return true
	case "COM1", "COM2", "COM3", "COM4", "COM5", "COM6", "COM7", "COM8", "COM9":
		return true
	case "LPT1", "LPT2", "LPT3", "LPT4", "LPT5", "LPT6", "LPT7", "LPT8", "LPT9":
		return true
	}
	return false
}
