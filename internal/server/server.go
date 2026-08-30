// Copyright (c) 2026 Michael Lechner. All rights reserved.
package server

import (
	"github.com/hmsoft0815/wollmilchsau/internal/npminstall"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// WollmilchsauServer wraps the MCP server with additional configuration.
type WollmilchsauServer struct {
	Server          *mcp.Server
	LogDir          string
	EnableArtifacts bool
	ArtifactAddr    string
	pkgManager      *npminstall.Manager
	tools           []*mcp.Tool
}

// serverIcon is the default icon for the wollmilchsau server.
var serverIcon = mcp.Icon{
	Source:   "data:image/svg+xml;base64,PHN2ZyB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciIHdpZHRoPSIyNCIgaGVpZ2h0PSIyNCIgdmlld0JveD0iMCAwIDI0IDI0IiBmaWxsPSJub25lIiBzdHJva2U9ImN1cnJlbnRDb2xvciIgc3Ryb2tlLXdpZHRoPSIyIiBzdHJva2UtbGluZWNhcD0icm91bmQiIHN0cm9rZS1saW5lam9pbj0icm91bmQiPjxwb2x5bGluZSBwb2ludHM9IjQgMTcgMTAgMTEgNCAxIi8+PGxpbmUgeDE9IjEyIiB5MT0iMTkiIHgyPSIyMCIgeTI9IjE5Ii8+PC9zdmc+",
	MIMEType: mimeTypeSVG,
}

// New creates a new MCP server wrapper for TypeScript execution.
func New(logDir string, enableArtifacts bool, artifactAddr string, bundledDeps []string) *WollmilchsauServer {
	impl := &mcp.Implementation{
		Name:    ServerName,
		Title:   ServerTitle,
		Version: ServerVersion,
		Icons:   []mcp.Icon{serverIcon},
	}

	s := mcp.NewServer(impl, nil)

	ws := &WollmilchsauServer{
		Server:          s,
		LogDir:          logDir,
		EnableArtifacts: enableArtifacts,
		ArtifactAddr:    artifactAddr,
	}

	// If bundled deps were specified, install them and attach manager.
	if len(bundledDeps) > 0 {
		ws.setupBundledDeps(bundledDeps)
	}

	ws.registerTools()
	ws.registerPrompts()

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
// none are configured.
func (s *WollmilchsauServer) bundledPackageInfos() []PackageInfo {
	if s.pkgManager == nil {
		return nil
	}
	rawInfos := s.pkgManager.PackageInfos()
	infos := make([]PackageInfo, 0, len(rawInfos))
	for _, raw := range rawInfos {
		name, _ := raw["name"].(string)
		ver, _ := raw["version"].(string)
		typ, _ := raw["type"].(string)
		main, _ := raw["main"].(string)
		desc, _ := raw["description"].(string)
		infos = append(infos, PackageInfo{
			Name:        name,
			Version:     ver,
			Type:        typ,
			Main:        main,
			Description: desc,
		})
	}
	return infos
}

// GetTools returns the definitions of all tools registered on this server.
func (s *WollmilchsauServer) GetTools() []*mcp.Tool {
	return s.tools
}
