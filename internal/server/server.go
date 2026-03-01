// Copyright (c) 2026 Michael Lechner. All rights reserved.
// Package server wires the MCP tools using mcp-go.
package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/hmsoft0815/mlcartifact"
	"github.com/hmsoft0815/wollmilchsau/internal/bundler"
	"github.com/hmsoft0815/wollmilchsau/internal/executor"
	"github.com/hmsoft0815/wollmilchsau/internal/parser"
	"github.com/hmsoft0815/wollmilchsau/internal/requestlog"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// WollmilchsauServer wraps the MCP server with additional configuration.
type WollmilchsauServer struct {
	MCPServer *server.MCPServer
	LogDir    string
}

// GetTools returns the definitions of all tools registered in this server.
func GetTools() []mcp.Tool {
	return []mcp.Tool{
		toolExecuteScript(),
		toolExecuteProject(),
		toolExecuteArtifact(),
		toolCheckSyntax(),
	}
}

// New creates a new MCP server wrapper for TypeScript execution.
func New(logDir string) *WollmilchsauServer {
	s := server.NewMCPServer(
		ServerName,
		ServerVersion,
		server.WithToolCapabilities(true),
		server.WithPromptCapabilities(true),
	)

	ws := &WollmilchsauServer{
		MCPServer: s,
		LogDir:    logDir,
	}

	s.AddTool(toolExecuteScript(), ws.handleExecuteScript)
	s.AddTool(toolExecuteProject(), ws.handleExecuteProject)
	s.AddTool(toolExecuteArtifact(), ws.handleExecuteArtifact)
	s.AddTool(toolCheckSyntax(), ws.handleCheckSyntax)

	s.AddPrompt(mcp.NewPrompt(PromptUsage, mcp.WithPromptDescription(PromptUsageDescription)), ws.handlePromptUsage)

	return ws
}

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

func toolCheckSyntax() mcp.Tool {
	return mcp.NewTool(
		ToolCheckSyntax,
		mcp.WithDescription(ToolCheckSyntaxDescription),
		mcp.WithString(ParamCode,
			mcp.Required(),
			mcp.Description(ParamCodeDescription),
		),
	)
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

func toolExecuteScript() mcp.Tool {
	return mcp.NewTool(
		ToolExecuteScript,
		mcp.WithDescription(ToolExecuteScriptDescription),
		mcp.WithString(ParamCode,
			mcp.Required(),
			mcp.Description(ParamCodeDescription),
		),
		mcp.WithNumber(ParamTimeoutMs,
			mcp.Description(ParamTimeoutMsDescription),
		),
	)
}

func toolExecuteProject() mcp.Tool {
	tool := mcp.NewTool(
		ToolExecuteProject,
		mcp.WithDescription(ToolExecuteProjectDescription),
	)

	// Manually add the complex 'files' property since helper functions are limited
	tool.InputSchema.Properties[ParamFiles] = map[string]any{
		"type": "array",
		"items": map[string]any{
			"type": "object",
			"properties": map[string]any{
				"name":    map[string]any{"type": "string", "description": "Filename (e.g. main.ts)"},
				"content": map[string]any{"type": "string", "description": "File content"},
			},
			"required": []string{"name", "content"},
		},
		"description": ParamFilesDescription,
	}
	tool.InputSchema.Required = append(tool.InputSchema.Required, ParamFiles)

	// Add simpler properties using helpers
	mcp.WithString(ParamEntryPoint,
		mcp.Required(),
		mcp.Description(ParamEntryPointDescription),
	)(&tool)

	mcp.WithNumber(ParamTimeoutMs,
		mcp.Description(ParamTimeoutMsDescription),
	)(&tool)

	return tool
}

func toolExecuteArtifact() mcp.Tool {
	return mcp.NewTool(
		ToolExecuteArtifact,
		mcp.WithDescription(ToolExecuteArtifactDescription),
		mcp.WithString(ParamArtifactID,
			mcp.Required(),
			mcp.Description(ParamArtifactIDDescription),
		),
		mcp.WithNumber(ParamTimeoutMs,
			mcp.Description(ParamTimeoutMsDescription),
		),
	)
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
	defer cli.Close()

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

func (s *WollmilchsauServer) runExecution(ctx context.Context, plan *parser.ExecutionPlan, toolName string) (*mcp.CallToolResult, error) {
	if plan.TimeoutMs == 0 {
		plan.TimeoutMs = 10_000
	}

	if err := parser.ValidatePlan(plan); err != nil {
		res := mcp.NewToolResultText("validation error: " + err.Error())
		res.IsError = true
		return res, nil
	}

	bundle, bundleErr := bundler.Bundle(plan)
	if bundleErr != nil {
		if be, ok := bundleErr.(*bundler.BundleError); ok {
			result := buildFailResult(be)
			meta := struct {
				Summary     string                `json:"summary"`
				Success     bool                  `json:"success"`
				ExitCode    int                   `json:"exitCode"`
				Diagnostics []executor.Diagnostic `json:"diagnostics,omitempty"`
			}{
				Summary:     result.Summary,
				Success:     result.Success,
				ExitCode:    result.ExitCode,
				Diagnostics: result.Diagnostics,
			}
			slog.Warn("build failed", "err", result.Summary)

			// Log to ZIP if enabled
			s.maybeLogRequest(ctx, toolName, plan, result)

			return &mcp.CallToolResult{
				Content:           []mcp.Content{mcp.NewTextContent("### Build Failure\n" + mustJSON(meta))},
				StructuredContent: meta,
				IsError:           true,
			}, nil
		}
		res := mcp.NewToolResultText("bundle error: " + bundleErr.Error())
		res.IsError = true
		return res, nil
	}

	execCtx, cancel := context.WithTimeout(ctx, time.Duration(plan.TimeoutMs)*time.Millisecond)
	defer cancel()

	result := executor.Execute(execCtx, bundle.JS, plan.EntryPoint, bundle.SourceMap)

	for _, w := range bundle.Warnings {
		result.Diagnostics = append(result.Diagnostics, executor.Diagnostic{
			Severity: executor.SeverityWarning,
			Message:  w.Text,
			Source:   w.Source,
			Line:     w.Line,
			Column:   w.Column,
		})
	}

	contents := []mcp.Content{}
	meta := struct {
		Summary     string                `json:"summary"`
		Success     bool                  `json:"success"`
		ExitCode    int                   `json:"exitCode"`
		DurationMs  int64                 `json:"durationMs"`
		Diagnostics []executor.Diagnostic `json:"diagnostics,omitempty"`
	}{
		Summary:     result.Summary,
		Success:     result.Success,
		ExitCode:    result.ExitCode,
		DurationMs:  result.DurationMs,
		Diagnostics: result.Diagnostics,
	}
	contents = append(contents, mcp.NewTextContent("### Status\n"+mustJSON(meta)))

	if strings.TrimSpace(result.Stdout) != "" {
		contents = append(contents, mcp.NewTextContent("### Standard Output\n```\n"+result.Stdout+"\n```"))
	}
	if strings.TrimSpace(result.Stderr) != "" {
		contents = append(contents, mcp.NewTextContent("### Standard Error\n```\n"+result.Stderr+"\n```"))
	}

	// Log to ZIP if enabled
	s.maybeLogRequest(ctx, toolName, plan, result)

	slog.Info("tool executed", "tool", toolName, "summary", result.Summary, "duration_ms", result.DurationMs, "success", result.Success)

	return &mcp.CallToolResult{
		Content:           contents,
		StructuredContent: meta,
		IsError:           !result.Success && result.ExitCode != 0,
	}, nil
}

func (s *WollmilchsauServer) maybeLogRequest(ctx context.Context, tool string, plan *parser.ExecutionPlan, result *executor.Result) {
	if s.LogDir == "" {
		return
	}

	remoteIP := GetRemoteIP(ctx)
	entry := requestlog.Entry{
		RemoteIP:  remoteIP,
		Tool:      tool,
		Plan:      plan,
		Result:    result,
		Timestamp: time.Now(),
	}

	zipPath, err := requestlog.LogRequest(s.LogDir, entry)
	if err != nil {
		slog.Error("failed to log request to zip", "err", err)
		return
	}

	slog.Info("request archived", "ip", remoteIP, "tool", tool, "zip", zipPath)
}

func buildFailResult(be *bundler.BundleError) *executor.Result {
	diags := make([]executor.Diagnostic, 0, len(be.Messages))
	for _, m := range be.Messages {
		diags = append(diags, executor.Diagnostic{
			Severity: executor.SeverityError,
			Message:  m.Text,
			Source:   m.Source,
			Line:     m.Line,
			Column:   m.Column,
		})
	}
	summary := "Build failed"
	if len(be.Messages) > 0 {
		m := be.Messages[0]
		summary = fmt.Sprintf("Build Error: %s in %s:%d", m.Text, m.Source, m.Line)
	}
	return &executor.Result{
		ExitCode:    1,
		Success:     false,
		Summary:     summary,
		Diagnostics: diags,
	}
}

func mustJSON(v any) string {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Sprintf(`{"error": %q}`, err.Error())
	}
	return string(b)
}
