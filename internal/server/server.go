// Copyright (c) 2026 Michael Lechner. All rights reserved.
// Package server wires the MCP tools using mcp-go.
package server

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// WollmilchsauServer wraps the MCP server with additional configuration.
type WollmilchsauServer struct {
	MCPServer       *server.MCPServer
	LogDir          string
	EnableArtifacts bool
	ArtifactAddr    string
}

// serverIcon is the default icon for the wollmilchsau server (a "terminal/code" glyph).
var serverIcon = mcp.Icon{
	Src:      "data:image/svg+xml;base64,PHN2ZyB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciIHdpZHRoPSIyNCIgaGVpZ2h0PSIyNCIgdmlld0JveD0iMCAwIDI0IDI0IiBmaWxsPSJub25lIiBzdHJva2U9ImN1cnJlbnRDb2xvciIgc3Ryb2tlLXdpZHRoPSIyIiBzdHJva2UtbGluZWNhcD0icm91bmQiIHN0cm9rZS1saW5lam9pbj0icm91bmQiPjxwb2x5bGluZSBwb2ludHM9IjQgMTcgMTAgMTEgNCAxIi8+PGxpbmUgeDE9IjEyIiB5MT0iMTkiIHgyPSIyMCIgeTI9IjE5Ii8+PC9zdmc+",
	MIMEType: mimeTypeSVG,
}

// New creates a new MCP server wrapper for TypeScript execution.
func New(logDir string, enableArtifacts bool, artifactAddr string) *WollmilchsauServer {
	hooks := &server.Hooks{}
	hooks.AddAfterInitialize(func(_ context.Context, _ any, _ *mcp.InitializeRequest, result *mcp.InitializeResult) {
		result.ServerInfo.Title = ServerTitle
		result.ServerInfo.Icons = []mcp.Icon{serverIcon}
	})

	s := server.NewMCPServer(
		ServerName,
		ServerVersion,
		server.WithToolCapabilities(true),
		server.WithPromptCapabilities(true),
		server.WithHooks(hooks),
	)

	ws := &WollmilchsauServer{
		MCPServer:       s,
		LogDir:          logDir,
		EnableArtifacts: enableArtifacts,
		ArtifactAddr:    artifactAddr,
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
