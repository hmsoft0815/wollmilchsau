// Copyright (c) 2026 Michael Lechner. All rights reserved.
package server

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestCheckSyntax_Valid(t *testing.T) {
	s := New("", false, "", nil)

	in := CheckSyntaxInput{
		Code: "const a: number = 10; console.log(a);",
	}

	res, out, err := s.handleCheckSyntax(context.Background(), nil, in)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.IsError || out == nil || !out.Success {
		t.Fatalf("expected syntax check to pass: %+v", res)
	}

	if out.Summary != "Syntax is valid" {
		t.Errorf("expected 'Syntax is valid', got %q", out.Summary)
	}
}

func TestCheckSyntax_Invalid(t *testing.T) {
	s := New("", false, "", nil)

	in := CheckSyntaxInput{
		Code: "const a = ;",
	}

	res, out, err := s.handleCheckSyntax(context.Background(), nil, in)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !res.IsError || out == nil || out.Success {
		t.Fatalf("expected syntax check to fail: %+v", res)
	}

	if len(out.Diagnostics) == 0 {
		t.Error("expected diagnostics for syntax error")
	}
}

func TestPromptUsage(t *testing.T) {
	s := New("", true, "", nil)

	res, err := s.handlePromptUsage(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(res.Messages) == 0 {
		t.Fatal("expected prompt messages")
	}

	txt, ok := res.Messages[0].Content.(*mcp.TextContent)
	if !ok || !strings.Contains(txt.Text, "wollmilchsau") {
		t.Errorf("expected prompt text to mention wollmilchsau, got: %+v", res.Messages[0].Content)
	}
}

func TestExecuteProject_ValidationError(t *testing.T) {
	s := New("", false, "", nil)

	in := ExecuteProjectInput{
		EntryPoint: "nonexistent.ts",
		Files: []ProjectFile{
			{
				Name:    "main.ts",
				Content: "console.log(1);",
			},
		},
	}

	res, out, err := s.handleExecuteProject(context.Background(), nil, in)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !res.IsError || out != nil {
		t.Errorf("expected validation error, got %+v (out=%+v)", res, out)
	}
}

func TestProjectFiles_UnmarshalJSONString(t *testing.T) {
	rawJSON := `"[{\"name\":\"main.ts\",\"content\":\"console.log('hi');\"}]"`
	var pf ProjectFiles
	if err := json.Unmarshal([]byte(rawJSON), &pf); err != nil {
		t.Fatalf("failed to unmarshal JSON string: %v", err)
	}

	if len(pf) != 1 || pf[0].Name != "main.ts" || pf[0].Content != "console.log('hi');" {
		t.Errorf("unexpected unmarshaled content: %+v", pf)
	}
}

func TestProjectFiles_UnmarshalJSONArray(t *testing.T) {
	rawJSON := `[{"name":"main.ts","content":"console.log('hi');"}]`
	var pf ProjectFiles
	if err := json.Unmarshal([]byte(rawJSON), &pf); err != nil {
		t.Fatalf("failed to unmarshal JSON array: %v", err)
	}

	if len(pf) != 1 || pf[0].Name != "main.ts" || pf[0].Content != "console.log('hi');" {
		t.Errorf("unexpected unmarshaled content: %+v", pf)
	}
}
