// Copyright (c) 2026 Michael Lechner. All rights reserved.
package parser

import (
	"fmt"
)

// VirtualFile represents a single TypeScript/JavaScript source file.
type VirtualFile struct {
	Name    string `json:"name"`
	Content string `json:"content"`
}

// ExecutionPlan is the result of parsing a command block.
// It contains all virtual files and metadata for bundling/execution.
type ExecutionPlan struct {
	Files      []VirtualFile `json:"files"`      // List of TypeScript/JavaScript source files
	EntryPoint string        `json:"entryPoint"` // Name of the entry point file
	TimeoutMs  int           `json:"timeoutMs"`  // Maximum execution time in milliseconds
}

// ParseError is returned for any malformed command block.
type ParseError struct {
	Line    int    `json:"line"`    // 1-based line number where the error occurred
	Message string `json:"message"` // Detailed error description
}

func (e *ParseError) Error() string {
	if e.Line > 0 {
		return fmt.Sprintf("parse error on line %d: %s", e.Line, e.Message)
	}
	return fmt.Sprintf("parse error: %s", e.Message)
}
