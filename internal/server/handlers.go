// Copyright (c) 2026 Michael Lechner. All rights reserved.
package server

import (
	"context"
	"log/slog"

	mlcartifact "github.com/hmsoft0815/mlcartifact/client"
	"github.com/hmsoft0815/wollmilchsau/internal/bundler"
	"github.com/hmsoft0815/wollmilchsau/internal/parser"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func (s *WollmilchsauServer) handlePromptUsage(_ context.Context, _ *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	return &mcp.GetPromptResult{
		Description: "Instructions on when to offload thinking to wollmilchsau",
		Messages: []*mcp.PromptMessage{
			{
				Role:    mcp.Role("user"),
				Content: &mcp.TextContent{Text: GetPromptUsageText(s.EnableArtifacts)},
			},
		},
	}, nil
}

func (s *WollmilchsauServer) handleCheckSyntax(_ context.Context, _ *mcp.CallToolRequest, in CheckSyntaxInput) (*mcp.CallToolResult, *CheckSyntaxResult, error) {
	plan := &parser.ExecutionPlan{
		Files: []parser.VirtualFile{
			{Name: "check.ts", Content: in.Code},
		},
		EntryPoint: "check.ts",
	}

	_, err := bundler.Bundle(plan)

	meta := &CheckSyntaxResult{
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
			Content:           []mcp.Content{&mcp.TextContent{Text: "### Syntax Check Failed\n" + mustJSON(meta)}},
			StructuredContent: meta,
			IsError:           true,
		}, meta, nil
	}

	meta.Summary = "Syntax is valid"
	return &mcp.CallToolResult{
		Content:           []mcp.Content{&mcp.TextContent{Text: "### Syntax Check Passed\n" + mustJSON(meta)}},
		StructuredContent: meta,
	}, meta, nil
}

func (s *WollmilchsauServer) handleExecuteScript(ctx context.Context, _ *mcp.CallToolRequest, in ExecuteScriptInput) (*mcp.CallToolResult, *ExecutionResult, error) {
	plan := &parser.ExecutionPlan{
		Files: []parser.VirtualFile{
			{Name: "script.ts", Content: in.Code},
		},
		EntryPoint: "script.ts",
		TimeoutMs:  in.TimeoutMs,
	}

	return s.runExecution(ctx, plan, ToolExecuteScript)
}

func (s *WollmilchsauServer) handleExecuteProject(ctx context.Context, _ *mcp.CallToolRequest, in ExecuteProjectInput) (*mcp.CallToolResult, *ExecutionResult, error) {
	plan := &parser.ExecutionPlan{
		EntryPoint: in.EntryPoint,
		TimeoutMs:  in.TimeoutMs,
	}

	for _, f := range in.Files {
		plan.Files = append(plan.Files, parser.VirtualFile{Name: f.Name, Content: f.Content})
	}

	return s.runExecution(ctx, plan, ToolExecuteProject)
}

func (s *WollmilchsauServer) handleExecuteArtifact(ctx context.Context, _ *mcp.CallToolRequest, in ExecuteArtifactInput) (*mcp.CallToolResult, *ExecutionResult, error) {
	// 1. Fetch artifact from service
	var cli *mlcartifact.Client
	var err error
	if s.ArtifactAddr != "" {
		cli, err = mlcartifact.NewClientWithAddr(s.ArtifactAddr)
	} else {
		cli, err = mlcartifact.NewClient()
	}
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: "Failed to connect to artifact service: " + err.Error()}},
			IsError: true,
		}, nil, nil
	}
	defer func() {
		if closeErr := cli.Close(); closeErr != nil {
			slog.Error("Failed to close artifact client", "error", closeErr)
		}
	}()

	opts := []mlcartifact.ReadOption{}
	if in.UserID != "" {
		opts = append(opts, mlcartifact.WithReadUserID(in.UserID))
	}

	res, err := cli.Read(ctx, in.ArtifactID, opts...)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: "Failed to read artifact: " + err.Error()}},
			IsError: true,
		}, nil, nil
	}

	plan := &parser.ExecutionPlan{
		Files: []parser.VirtualFile{
			{Name: res.Filename, Content: string(res.Content)},
		},
		EntryPoint: res.Filename,
		TimeoutMs:  in.TimeoutMs,
	}

	return s.runExecution(ctx, plan, ToolExecuteArtifact)
}
