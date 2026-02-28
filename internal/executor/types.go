// Copyright (c) 2026 Michael Lechner. All rights reserved.
package executor

type Severity string

const (
	SeverityError   Severity = "error"
	SeverityWarning Severity = "warning"
	SeverityInfo    Severity = "info"
)

// Diagnostic represents a single error or warning with position resolved to original TypeScript.
type Diagnostic struct {
	Severity Severity `json:"severity"`
	Message  string   `json:"message"`
	Source   string   `json:"source,omitempty"` // original .ts file, e.g. "main.ts"
	Line     int      `json:"line,omitempty"`   // 1-based in original .ts
	Column   int      `json:"column,omitempty"` // 1-based in original .ts

	// Internal debug fields (hidden from LLM/JSON)
	GeneratedLine   int `json:"-"`
	GeneratedColumn int `json:"-"`
}

// Result is the output of a single Execute call.
// It contains stdout, stderr, exit code and potential diagnostics.
type Result struct {
	Stdout      string       `json:"stdout"`      // Standard output captured from console.log
	Stderr      string       `json:"stderr"`      // Standard error captured from console.warn/error
	ExitCode    int          `json:"exitCode"`    // 0 for success, non-zero for error
	Success     bool         `json:"success"`     // true if execution finished without runtime errors
	DurationMs  int64        `json:"durationMs"`  // execution time in milliseconds
	Summary     string       `json:"summary"`     // High-level summary of the result
	Diagnostics []Diagnostic `json:"diagnostics"` // list of errors/warnings with mapped positions
}
