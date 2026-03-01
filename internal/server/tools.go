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
	)
}
