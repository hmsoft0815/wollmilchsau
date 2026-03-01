// Copyright (c) 2026 Michael Lechner. All rights reserved.
package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/hmsoft0815/wollmilchsau/internal/bundler"
	"github.com/hmsoft0815/wollmilchsau/internal/executor"
	"github.com/hmsoft0815/wollmilchsau/internal/parser"
	"github.com/hmsoft0815/wollmilchsau/internal/requestlog"
	"github.com/mark3labs/mcp-go/mcp"
)

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
