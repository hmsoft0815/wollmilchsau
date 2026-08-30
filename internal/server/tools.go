// Copyright (c) 2026 Michael Lechner. All rights reserved.
package server

import (
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func (s *WollmilchsauServer) registerTools() {
	toolExecScript := &mcp.Tool{
		Name:        ToolExecuteScript,
		Description: GetToolExecuteScriptDescription(s.EnableArtifacts),
	}
	mcp.AddTool(s.Server, toolExecScript, s.handleExecuteScript)
	s.tools = append(s.tools, toolExecScript)

	toolExecProj := &mcp.Tool{
		Name:        ToolExecuteProject,
		Description: GetToolExecuteProjectDescription(s.EnableArtifacts),
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"entryPoint": map[string]any{
					"type":        "string",
					"description": "The name of the file to start execution from (e.g. 'main.ts').",
				},
				"files": map[string]any{
					"description": "A list of virtual files {name, content} to include in the project.",
					"anyOf": []map[string]any{
						{
							"type": "array",
							"items": map[string]any{
								"type": "object",
								"properties": map[string]any{
									"name":    map[string]any{"type": "string", "description": "Filename (e.g. main.ts)"},
									"content": map[string]any{"type": "string", "description": "File content"},
								},
								"required": []string{"name", "content"},
							},
						},
						{
							"type":        "string",
							"description": "JSON-encoded array of {name, content} objects.",
						},
					},
				},
				"timeoutMs": map[string]any{
					"type":        "integer",
					"description": "Maximum execution time in milliseconds (100 - 30000).",
				},
			},
			"required": []string{"entryPoint", "files"},
		},
	}
	mcp.AddTool(s.Server, toolExecProj, s.handleExecuteProject)
	s.tools = append(s.tools, toolExecProj)

	if s.EnableArtifacts {
		toolExecArt := &mcp.Tool{
			Name:        ToolExecuteArtifact,
			Description: GetToolExecuteArtifactDescription(s.EnableArtifacts),
		}
		mcp.AddTool(s.Server, toolExecArt, s.handleExecuteArtifact)
		s.tools = append(s.tools, toolExecArt)
	}

	toolChkSyntax := &mcp.Tool{
		Name:        ToolCheckSyntax,
		Description: ToolCheckSyntaxDescription,
	}
	mcp.AddTool(s.Server, toolChkSyntax, s.handleCheckSyntax)
	s.tools = append(s.tools, toolChkSyntax)

	toolListDeps := &mcp.Tool{
		Name:        ToolListJSPackages,
		Description: listJSPKGDesc,
	}
	mcp.AddTool(s.Server, toolListDeps, s.handleListJSPackages)
	s.tools = append(s.tools, toolListDeps)
}

func (s *WollmilchsauServer) registerPrompts() {
	s.Server.AddPrompt(&mcp.Prompt{
		Name:        PromptUsage,
		Description: PromptUsageDescription,
	}, s.handlePromptUsage)
}

// GetTools returns tools for static inspection or dump.
func GetTools(enableArtifacts bool, bundledDeps []string) []*mcp.Tool {
	ws := New("", enableArtifacts, "", bundledDeps)
	return ws.GetTools()
}
