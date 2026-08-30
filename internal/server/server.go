// Copyright (c) 2026 Michael Lechner. All rights reserved.
package server

import (
	"context"

	"github.com/hmsoft0815/wollmilchsau/internal/npminstall"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// WollmilchsauServer wraps the MCP server with additional configuration.
type WollmilchsauServer struct {
	MCPServer       *server.MCPServer
	LogDir          string
	EnableArtifacts bool
	ArtifactAddr    string
	pkgManager      *npminstall.Manager
}

// serverIcon is the default icon for the wollmilchsau server.
var serverIcon = mcp.Icon{
	Src:      "data:image/svg+xml;base64,PHN2ZyB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciIHdpZHRoPSIyNCIgaGVpZ2h0PSIyNCIgdmlld0JveD0iMCAwIDI0IDI0IiBmaWxsPSJub25lIiBzdHJva2U9ImN1cnJlbnRDb2xvciIgc3Ryb2tlLXdpZHRoPSIyIiBzdHJva2UtbGluZWNhcD0icm91bmQiIHN0cm9rZS1saW5lam9pbj0icm91bmQiPjxwb2x5bGluZSBwb2ludHM9IjQgMTcgMTAgMTEgNCAxIi8+PGxpbmUgeDE9IjEyIiB5MT0iMTkiIHgyPSIyMCIgeTI9IjE5Ii8+PC9zdmc+",
	MIMEType: mimeTypeSVG,
}

// New creates a new MCP server wrapper for TypeScript execution.
func New(logDir string, enableArtifacts bool, artifactAddr string, bundledDeps []string) *WollmilchsauServer {
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

	// If bundled deps were specified, install them and attach manager.
	if len(bundledDeps) > 0 {
		ws.setupBundledDeps(bundledDeps)
	}

	s.AddTool(toolExecuteScript(enableArtifacts), ws.handleExecuteScript)
	s.AddTool(toolExecuteProject(enableArtifacts), ws.handleExecuteProject)
	if enableArtifacts {
		s.AddTool(toolExecuteArtifact(enableArtifacts), ws.handleExecuteArtifact)
	}
	s.AddTool(toolCheckSyntax(), ws.handleCheckSyntax)

	// Always register list_js_packages (tool is always available,
	// even when no deps are configured — it will say so).
	s.AddTool(toolListJSPackages(), ws.handleListJSPackages)

	s.AddPrompt(mcp.NewPrompt(PromptUsage, mcp.WithPromptDescription(PromptUsageDescription)), ws.handlePromptUsage)

	return ws
}

// setupBundledDeps installs the requested npm packages and attaches the manager.
func (s *WollmilchsauServer) setupBundledDeps(packages []string) {
	mgr, err := npminstall.NewManager("")
	if err != nil {
		return // best-effort: no bundled deps attached
	}

	if err := mgr.Install(packages); err != nil {
		return // install failed silently; agent won't see bundled packages
	}

	s.pkgManager = mgr
}

// bundledPackageInfos returns the list of available JS packages, or nil if
// none are configured. Intended for JSON serialisation in tool responses.
func (s *WollmilchsauServer) bundledPackageInfos() []map[string]any {
	if s.pkgManager == nil {
		return nil
	}
	return s.pkgManager.PackageInfos()
}
