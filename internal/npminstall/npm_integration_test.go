// Copyright (c) 2026 Michael Lechner. All rights reserved.
//go:build npm_integration
// +build npm_integration

package npminstall

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInstallAndVirtualFiles(t *testing.T) {
	dir := t.TempDir()

	// Create minimal package.json.
	if err := os.WriteFile(filepath.Join(dir, "package.json"), []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}

	mgr, err := NewManager(dir)
	if err != nil {
		t.Fatalf("NewManager: %v", err)
	}

	err = mgr.Install([]string{"lodash", "zod"})
	if err != nil {
		t.Fatalf("Install: %v", err)
	}

	pkgs := mgr.Packages()
	names := make(map[string]bool)
	for _, p := range pkgs {
		names[p.Name] = true
	}
	if !names["lodash"] {
		t.Errorf("expected lodash in packages, got %v", names)
	}
	if !names["zod"] {
		t.Errorf("expected zod in packages, got %v", names)
	}

	vfs, err := mgr.ToVirtualFiles()
	if err != nil {
		t.Fatalf("ToVirtualFiles: %v", err)
	}
	if len(vfs) == 0 {
		t.Fatal("expected non-empty VirtualFile slice")
	}

	// Check that lodash main file is present.
	hasLodashMain := false
	for _, vf := range vfs {
		if vf.Name == "node_modules/lodash/lodash.js" || vf.Name == "node_modules/lodash/core.js" {
			hasLodashMain = true
			break
		}
	}
	if !hasLodashMain {
		t.Error("expected lodash main file in VirtualFiles")
	}

	t.Logf("Installed %d files for %d packages", len(vfs), len(pkgs))
}
