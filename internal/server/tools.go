// Copyright (c) 2026 Michael Lechner. All rights reserved.
package server

import (
	"github.com/mark3labs/mcp-go/mcp"
)

// GetTools returns the definitions of all tools registered in this server.
func GetTools(enableArtifacts bool) []mcp.Tool {
	tools := []mcp.Tool{
		toolExecuteScript(enableArtifacts),
		toolExecuteProject(enableArtifacts),
		toolCheckSyntax(),
	}
	if enableArtifacts {
		tools = append(tools, toolExecuteArtifact(enableArtifacts))
	}
	return tools
}

func toolCheckSyntax() mcp.Tool {
	return mcp.NewTool(
		ToolCheckSyntax,
		mcp.WithDescription(ToolCheckSyntaxDescription),
		mcp.WithString(ParamCode,
			mcp.Required(),
			mcp.Description(ParamCodeDescription),
		),
		mcp.WithToolIcons(mcp.Icon{
			Src:      "data:image/svg+xml;base64,PHN2ZyB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciIHdpZHRoPSIyNCIgaGVpZ2h0PSIyNCIgdmlld0JveD0iMCAwIDI0IDI0IiBmaWxsPSJub25lIiBzdHJva2U9ImN1cnJlbnRDb2xvciIgc3Ryb2tlLXdpZHRoPSIyIiBzdHJva2UtbGluZWNhcD0icm91bmQiIHN0cm9rZS1saW5lam9pbj0icm91bmQiPjxwYXRoIGQ9Ik05IDExbDMgMyA4LTgtMi0yLTEwIDEwem0tMyAwbC00IDQgNCA0IDItMi00LTQtMi0yek0wIDI0aDI0Ii8+PC9zdmc+",
			MIMEType: mimeTypeSVG,
		}),
		mcp.WithOutputSchema[CheckSyntaxResult](),
	)
}

func toolExecuteScript(enableArtifacts bool) mcp.Tool {
	return mcp.NewTool(
		ToolExecuteScript,
		mcp.WithDescription(GetToolExecuteScriptDescription(enableArtifacts)),
		mcp.WithString(ParamCode,
			mcp.Required(),
			mcp.Description(ParamCodeDescription),
		),
		mcp.WithNumber(ParamTimeoutMs,
			mcp.Description(ParamTimeoutMsDescription),
		),
		mcp.WithToolIcons(mcp.Icon{
			Src:      "data:image/svg+xml;base64,PHN2ZyB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciIHdpZHRoPSIyNCIgaGVpZ2h0PSIyNCIgdmlld0JveD0iMCAwIDI0IDI0IiBmaWxsPSJub25lIiBzdHJva2U9ImN1cnJlbnRDb2xvciIgc3Ryb2tlLXdpZHRoPSIyIiBzdHJva2UtbGluZWNhcD0icm91bmQiIHN0cm9rZS1saW5lam9pbj0icm91bmQiPjxwYXRoIGQ9Ik0xNiAxOGwtMiAybC0yLTIybTQtOGw0IDRsLTQgNE0yMiAxOXYtMk0xNSA1aC0yTTUgNWgtMk01IDE1aC0yTTUgMTloLTJNMjIgNXYtMk0yMiAxOXYtMk05IDVoLTJNOSAxOWgtMk0xMyA1aC0yTTEzIDE5aC0yTTE3IDVoLTJNMjIgOXYtMiIvPjwvc3ZnPg==",
			MIMEType: mimeTypeSVG,
		}),
		mcp.WithOutputSchema[ExecutionResult](),
	)
}

func toolExecuteProject(enableArtifacts bool) mcp.Tool {
	tool := mcp.NewTool(
		ToolExecuteProject,
		mcp.WithDescription(GetToolExecuteProjectDescription(enableArtifacts)),
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

	mcp.WithToolIcons(mcp.Icon{
		Src:      "data:image/svg+xml;base64,PHN2ZyB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciIHdpZHRoPSIyNCIgaGVpZ2h0PSIyNCIgdmlld0JveD0iMCAwIDI0IDI0IiBmaWxsPSJub25lIiBzdHJva2U9ImN1cnJlbnRDb2xvciIgc3Ryb2tlLXdpZHRoPSIyIiBzdHJva2UtbGluZWNhcD0icm91bmQiIHN0cm9rZS1saW5lam9pbj0icm91bmQiPjxwYXRoIGQ9Ik0xMiAyTDQgNnYxMmwxIDguNWwtOC00VjZ6TTEyIDIybDgtNGwtOC00TC04IDR6TTQgNmw4IDRsOC00TTIgMTV2MkwxMiAyMmw4LTUtMnYtMiIvPjwvc3ZnPg==",
		MIMEType: mimeTypeSVG,
	})(&tool)

	mcp.WithOutputSchema[ExecutionResult]()(&tool)

	return tool
}

func toolExecuteArtifact(enableArtifacts bool) mcp.Tool {
	return mcp.NewTool(
		ToolExecuteArtifact,
		mcp.WithDescription(GetToolExecuteArtifactDescription(enableArtifacts)),
		mcp.WithString(ParamArtifactID,
			mcp.Required(),
			mcp.Description(ParamArtifactIDDescription),
		),
		mcp.WithNumber(ParamTimeoutMs,
			mcp.Description(ParamTimeoutMsDescription),
		),
		mcp.WithToolIcons(mcp.Icon{
			Src:      "data:image/svg+xml;base64,PHN2ZyB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciIHdpZHRoPSIyNCIgaGVpZ2h0PSIyNCIgdmlld0JveD0iMCAwIDI0IDI0IiBmaWxsPSJub25lIiBzdHJva2U9ImN1cnJlbnRDb2xvciIgc3Ryb2tlLXdpZHRoPSIyIiBzdHJva2UtbGluZWNhcD0icm91bmQiIHN0cm9rZS1saW5lam9pbj0icm91bmQiPjxwYXRoIGQ9Ik0xNCAydkg2YTIgMiAwIDAgMC0yIDJ2MTZhMiAyIDAgMCAwIDIgMmgxMmEyIDIgMCAwIDAgMi0yVjhsLTYtNnoiLz48cG9seWxpbmUgcG9pbnRzPSIxNCAyIDE0IDggMjAgOCIvPjwvc3ZnPg==",
			MIMEType: mimeTypeSVG,
		}),
		mcp.WithOutputSchema[ExecutionResult](),
	)
}
