// Copyright (c) 2026 Michael Lechner. All rights reserved.
package server

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hmsoft0815/wollmilchsau/internal/npminstall"
	"github.com/mark3labs/mcp-go/mcp"
)

func TestGetTools(t *testing.T) {
	toolsNoArt := GetTools(false, nil)
	names := make(map[string]bool)
	for _, tool := range toolsNoArt {
		names[tool.Name] = true
	}

	expectedTools := []string{
		ToolExecuteScript,
		ToolExecuteProject,
		ToolCheckSyntax,
		ToolListJSPackages,
	}
	for _, expected := range expectedTools {
		if !names[expected] {
			t.Errorf("expected tool %s to be registered", expected)
		}
	}

	if names[ToolExecuteArtifact] {
		t.Error("execute_artifact should not be present when enableArtifacts=false")
	}

	toolsWithArt := GetTools(true, nil)
	hasArtifactTool := false
	for _, tool := range toolsWithArt {
		if tool.Name == ToolExecuteArtifact {
			hasArtifactTool = true
			break
		}
	}
	if !hasArtifactTool {
		t.Error("expected execute_artifact tool when enableArtifacts=true")
	}
}

func TestListJSPackages_Empty(t *testing.T) {
	s := New("", false, "", nil)
	var req mcp.CallToolRequest
	req.Params.Name = ToolListJSPackages

	res, err := s.handleListJSPackages(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	txtContent, ok := res.Content[0].(mcp.TextContent)
	if !ok || !strings.Contains(txtContent.Text, "No bundled JS packages are configured") {
		t.Errorf("unexpected response: %+v", res.Content)
	}
}

func TestListJSPackages_WithPackages(t *testing.T) {
	dir := t.TempDir()
	mgr, err := npminstall.NewManager(dir)
	if err != nil {
		t.Fatal(err)
	}

	nm := mgr.NodeModulesPath()
	pkgDir := filepath.Join(nm, "lodash")
	if err = os.MkdirAll(pkgDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(pkgDir, "package.json"), []byte(`{
		"name": "lodash",
		"version": "4.17.21",
		"main": "lodash.js"
	}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(pkgDir, "lodash.js"), []byte(`module.exports = { version: "4.17.21" };`), 0o644); err != nil {
		t.Fatal(err)
	}

	if err = mgr.Reload(); err != nil {
		t.Fatal(err)
	}

	s := New("", false, "", nil)
	s.pkgManager = mgr

	var req mcp.CallToolRequest
	req.Params.Name = ToolListJSPackages

	res, err := s.handleListJSPackages(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	txtContent, ok := res.Content[0].(mcp.TextContent)
	if !ok || !strings.Contains(txtContent.Text, "lodash") {
		t.Errorf("expected lodash in output: %+v", res.Content)
	}

	infos := s.bundledPackageInfos()
	if len(infos) != 1 || infos[0]["name"] != "lodash" {
		t.Errorf("unexpected bundledPackageInfos: %+v", infos)
	}
}

func TestExecuteScript_WithBundledDeps(t *testing.T) {
	dir := t.TempDir()
	mgr, err := npminstall.NewManager(dir)
	if err != nil {
		t.Fatal(err)
	}

	nm := mgr.NodeModulesPath()
	pkgDir := filepath.Join(nm, "mymath")
	if err = os.MkdirAll(pkgDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(pkgDir, "package.json"), []byte(`{
		"name": "mymath",
		"version": "1.0.0",
		"main": "index.js"
	}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(pkgDir, "index.js"), []byte(`
		exports.add = function(a, b) { return a + b; };
	`), 0o644); err != nil {
		t.Fatal(err)
	}

	if err = mgr.Reload(); err != nil {
		t.Fatal(err)
	}

	s := New("", false, "", nil)
	s.pkgManager = mgr

	var req mcp.CallToolRequest
	req.Params.Name = ToolExecuteScript
	req.Params.Arguments = map[string]any{
		ParamCode: `
			const math = require("mymath");
			console.log("RESULT=" + math.add(20, 22));
		`,
	}

	res, err := s.handleExecuteScript(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.IsError {
		t.Fatalf("tool execution reported error: %+v", res)
	}

	foundStdout := false
	for _, c := range res.Content {
		if tc, ok := c.(mcp.TextContent); ok {
			if strings.Contains(tc.Text, "RESULT=42") {
				foundStdout = true
			}
		}
	}
	if !foundStdout {
		t.Errorf("expected stdout RESULT=42, got content: %+v", res.Content)
	}
}

func TestExecuteProject_WithBundledDeps(t *testing.T) {
	dir := t.TempDir()
	mgr, err := npminstall.NewManager(dir)
	if err != nil {
		t.Fatal(err)
	}

	nm := mgr.NodeModulesPath()
	pkgDir := filepath.Join(nm, "helper")
	if err = os.MkdirAll(pkgDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(pkgDir, "package.json"), []byte(`{
		"name": "helper",
		"version": "1.0.0",
		"main": "index.js"
	}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(pkgDir, "index.js"), []byte(`
		exports.greet = function(name) { return "Hello " + name; };
	`), 0o644); err != nil {
		t.Fatal(err)
	}

	if err = mgr.Reload(); err != nil {
		t.Fatal(err)
	}

	s := New("", false, "", nil)
	s.pkgManager = mgr

	var req mcp.CallToolRequest
	req.Params.Name = ToolExecuteProject
	req.Params.Arguments = map[string]any{
		ParamEntryPoint: "main.ts",
		ParamFiles: []any{
			map[string]any{
				"name":    "main.ts",
				"content": `import { greet } from "helper"; console.log(greet("Antigravity"));`,
			},
		},
	}

	res, err := s.handleExecuteProject(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.IsError {
		t.Fatalf("tool execution reported error: %+v", res)
	}

	foundGreeting := false
	for _, c := range res.Content {
		if tc, ok := c.(mcp.TextContent); ok && strings.Contains(tc.Text, "Hello Antigravity") {
			foundGreeting = true
		}
	}
	if !foundGreeting {
		t.Errorf("expected greeting in output, got: %+v", res.Content)
	}
}
