// Copyright (c) 2026 Michael Lechner. All rights reserved.
package server

import "github.com/hmsoft0815/wollmilchsau/internal/executor"

// ExecutionResult represents the structured output of an execution tool.
type ExecutionResult struct {
	Summary     string                `json:"summary"`
	Success     bool                  `json:"success"`
	ExitCode    int                   `json:"exitCode"`
	DurationMs  int64                 `json:"durationMs,omitempty"`
	Diagnostics []executor.Diagnostic `json:"diagnostics,omitempty"`
}

// CheckSyntaxResult represents the structured output of the check_syntax tool.
type CheckSyntaxResult struct {
	Success     bool                  `json:"success"`
	Summary     string                `json:"summary"`
	Diagnostics []executor.Diagnostic `json:"diagnostics,omitempty"`
}
