// Copyright (c) 2026 Michael Lechner. All rights reserved.
package server

import (
	"context"
	"log/slog"

	"github.com/hmsoft0815/mlcartifact"
	"github.com/hmsoft0815/wollmilchsau/internal/bundler"
	"github.com/hmsoft0815/wollmilchsau/internal/executor"
	"github.com/hmsoft0815/wollmilchsau/internal/parser"
	"github.com/mark3labs/mcp-go/mcp"
)

func (s *WollmilchsauServer) handlePromptUsage(ctx context.Context, req mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	return &mcp.GetPromptResult{
		Description: "Instructions on when to offload thinking to wollmilchsau",
		Messages: []mcp.PromptMessage{
			{
				Role:    "system",
				Content: mcp.NewTextContent(PromptUsageText),
			},
		},
	}, nil
}

func (s *WollmilchsauServer) handleCheckSyntax(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args, _ := req.Params.Arguments.(map[string]any)
	code, _ := args[ParamCode].(string)

	plan := &parser.ExecutionPlan{
		Files: []parser.VirtualFile{
			{Name: "check.ts", Content: code},
		},
		EntryPoint: "check.ts",
	}

	// We use the bundler just to see if it compiles
	_, err := bundler.Bundle(plan)

	meta := struct {
		Success     bool                  `json:"success"`
		Summary     string                `json:"summary"`
		Diagnostics []executor.Diagnostic `json:"diagnostics,omitempty"`
	}{
		Success: err == nil,
	}

	if err != nil {
		if be, ok := err.(*bundler.BundleError); ok {
			result := buildFailResult(be)
			meta.Summary = result.Summary
			meta.Diagnostics = result.Diagnostics
		} else {
			meta.Summary = "Internal check error: " + err.Error()
		}
		return &mcp.CallToolResult{
			Content:           []mcp.Content{mcp.NewTextContent("### Syntax Check Failed\n" + mustJSON(meta))},
			StructuredContent: meta,
			IsError:           true,
		}, nil
	}

	meta.Summary = "Syntax is valid"
	return &mcp.CallToolResult{
		Content:           []mcp.Content{mcp.NewTextContent("### Syntax Check Passed\n" + mustJSON(meta))},
		StructuredContent: meta,
	}, nil
}

func (s *WollmilchsauServer) handleExecuteScript(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args, _ := req.Params.Arguments.(map[string]any)
	code, _ := args[ParamCode].(string)
	timeout, _ := args[ParamTimeoutMs].(float64)

	plan := &parser.ExecutionPlan{
		Files: []parser.VirtualFile{
			{Name: "script.ts", Content: code},
		},
		EntryPoint: "script.ts",
		TimeoutMs:  int(timeout),
	}

	return s.runExecution(ctx, plan, ToolExecuteScript)
}

func (s *WollmilchsauServer) handleExecuteProject(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args, _ := req.Params.Arguments.(map[string]any)
	filesRaw, _ := args[ParamFiles].([]any)
	entryPoint, _ := args[ParamEntryPoint].(string)
	timeout, _ := args[ParamTimeoutMs].(float64)

	plan := &parser.ExecutionPlan{
		EntryPoint: entryPoint,
		TimeoutMs:  int(timeout),
	}

	for _, f := range filesRaw {
		fm, ok := f.(map[string]any)
		if !ok {
			continue
		}
		name, _ := fm["name"].(string)
		content, _ := fm["content"].(string)
		plan.Files = append(plan.Files, parser.VirtualFile{Name: name, Content: content})
	}

	return s.runExecution(ctx, plan, ToolExecuteProject)
}

func (s *WollmilchsauServer) handleExecuteArtifact(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args, _ := req.Params.Arguments.(map[string]any)
	artifactID, _ := args[ParamArtifactID].(string)
	timeout, _ := args[ParamTimeoutMs].(float64)

	// 1. Fetch artifact from service
	cli, err := mlcartifact.NewClient()
	if err != nil {
		return mcp.NewToolResultErrorFromErr("Failed to connect to artifact service", err), nil
	}
	defer func() {
		if closeErr := cli.Close(); closeErr != nil {
			slog.Error("Failed to close artifact client", "error", closeErr)
		}
	}()

	res, err := cli.Read(ctx, artifactID)
	if err != nil {
		return mcp.NewToolResultErrorFromErr("Failed to read artifact", err), nil
	}

	plan := &parser.ExecutionPlan{
		Files: []parser.VirtualFile{
			{Name: res.Filename, Content: string(res.Content)},
		},
		EntryPoint: res.Filename,
		TimeoutMs:  int(timeout),
	}

	return s.runExecution(ctx, plan, ToolExecuteArtifact)
}
