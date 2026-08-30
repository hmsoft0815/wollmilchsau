// Copyright (c) 2026 Michael Lechner. All rights reserved.
package server

import (
	"context"
	"log/slog"

	"github.com/hmsoft0815/wollmilchsau/internal/parser"
	"github.com/mark3labs/mcp-go/mcp"
)

// handleListJSPackages returns all bundled JS packages with metadata.
func (s *WollmilchsauServer) handleListJSPackages(_ context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if s.pkgManager == nil {
		return mcp.NewToolResultText("No bundled JS packages are configured."), nil
	}

	infos := s.pkgManager.PackageInfos()
	meta := struct {
		Packages []map[string]any `json:"packages"`
		Count    int              `json:"count"`
	}{
		Packages: infos,
		Count:    len(infos),
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.NewTextContent("### Bundled JS Packages\n" + mustJSON(meta)),
		},
		StructuredContent: meta,
	}, nil
}

// injectBundledDeps appends VirtualFiles from the manager's node_modules to
// plan.Files so that esbuild can resolve require/import against them.
func (s *WollmilchsauServer) injectBundledDeps(plan *parser.ExecutionPlan) {
	if s.pkgManager == nil {
		return
	}

	vfs, err := s.pkgManager.ToVirtualFiles()
	if err != nil {
		slog.Warn("failed to convert bundled deps to VirtualFiles", "err", err)
		return
	}
	plan.Files = append(plan.Files, vfs...)
}
