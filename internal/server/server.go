// Copyright (c) 2026 Michael Lechner. All rights reserved.
// Package server wires the MCP tools using mcp-go.
package server

import (
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// WollmilchsauServer wraps the MCP server with additional configuration.
type WollmilchsauServer struct {
	MCPServer       *server.MCPServer
	LogDir          string
	EnableArtifacts bool
}

// New creates a new MCP server wrapper for TypeScript execution.
func New(logDir string, enableArtifacts bool) *WollmilchsauServer {
	s := server.NewMCPServer(
		ServerName,
		ServerVersion,
		server.WithToolCapabilities(true),
		server.WithPromptCapabilities(true),
	)

	ws := &WollmilchsauServer{
		MCPServer:       s,
		LogDir:          logDir,
		EnableArtifacts: enableArtifacts,
	}

	s.AddTool(toolExecuteScript(enableArtifacts), ws.handleExecuteScript)
	s.AddTool(toolExecuteProject(enableArtifacts), ws.handleExecuteProject)
	if enableArtifacts {
		s.AddTool(toolExecuteArtifact(enableArtifacts), ws.handleExecuteArtifact)
	}
	s.AddTool(toolCheckSyntax(), ws.handleCheckSyntax)

	s.AddPrompt(mcp.NewPrompt(PromptUsage, mcp.WithPromptDescription(PromptUsageDescription)), ws.handlePromptUsage)

	return ws
}
