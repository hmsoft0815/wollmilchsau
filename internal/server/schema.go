// Copyright (c) 2026 Michael Lechner. All rights reserved.
package server

import (
	"encoding/json"

	"github.com/hmsoft0815/wollmilchsau/internal/executor"
)

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

// ExecuteScriptInput defines the input parameters for the execute_script tool.
type ExecuteScriptInput struct {
	Code      string `json:"code" jsonschema:"The TypeScript/JavaScript code to execute."`
	TimeoutMs int    `json:"timeoutMs,omitempty" jsonschema:"Maximum execution time in milliseconds (100 - 30000)."`
}

// ProjectFile represents a single virtual file in a multi-file project.
type ProjectFile struct {
	Name    string `json:"name" jsonschema:"Filename (e.g. main.ts)"`
	Content string `json:"content" jsonschema:"File content"`
}

// ProjectFiles is a slice of ProjectFile that supports unmarshaling from both
// a JSON array and a JSON string (e.g. from heredocs or CLI clients).
type ProjectFiles []ProjectFile

// UnmarshalJSON unmarshals either a JSON array or a JSON-encoded string containing an array.
func (pf *ProjectFiles) UnmarshalJSON(data []byte) error {
	if len(data) == 0 || string(data) == "null" {
		*pf = nil
		return nil
	}
	if data[0] == '"' {
		var s string
		if err := json.Unmarshal(data, &s); err != nil {
			return err
		}
		var list []ProjectFile
		if err := json.Unmarshal([]byte(s), &list); err != nil {
			return err
		}
		*pf = list
		return nil
	}
	var list []ProjectFile
	if err := json.Unmarshal(data, &list); err != nil {
		return err
	}
	*pf = list
	return nil
}

// ExecuteProjectInput defines the input parameters for the execute_project tool.
type ExecuteProjectInput struct {
	Files      ProjectFiles `json:"files" jsonschema:"A list of virtual files {name, content} to include in the project."`
	EntryPoint string       `json:"entryPoint" jsonschema:"The name of the file to start execution from (e.g. 'main.ts')."`
	TimeoutMs  int          `json:"timeoutMs,omitempty" jsonschema:"Maximum execution time in milliseconds (100 - 30000)."`
}

// CheckSyntaxInput defines the input parameters for the check_syntax tool.
type CheckSyntaxInput struct {
	Code string `json:"code" jsonschema:"The TypeScript/JavaScript code to validate."`
}

// ListJSPackagesInput defines empty input parameters for list_js_packages tool.
type ListJSPackagesInput struct{}

// PackageInfo describes a bundled JavaScript package.
type PackageInfo struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	Type        string `json:"type,omitempty"`
	Main        string `json:"main,omitempty"`
	Description string `json:"description,omitempty"`
}

// ListJSPackagesResult represents the structured output of list_js_packages.
type ListJSPackagesResult struct {
	Packages []PackageInfo `json:"packages"`
	Count    int           `json:"count"`
}

// ExecuteArtifactInput defines the input parameters for the execute_artifact tool.
type ExecuteArtifactInput struct {
	ArtifactID string `json:"artifactId" jsonschema:"The ID or filename of the artifact to execute."`
	UserID     string `json:"userId,omitempty" jsonschema:"Optional user ID to scope the artifact lookup."`
	TimeoutMs  int    `json:"timeoutMs,omitempty" jsonschema:"Maximum execution time in milliseconds (100 - 30000)."`
}
