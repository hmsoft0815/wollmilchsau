// Copyright (c) 2026 Michael Lechner. All rights reserved.
package requestlog

import (
	"archive/zip"
	"io"
	"os"
	"testing"
	"time"

	"github.com/hmsoft0815/wollmilchsau/internal/executor"
	"github.com/hmsoft0815/wollmilchsau/internal/parser"
)

func TestLogRequest_Success(t *testing.T) {
	tmpDir := t.TempDir()

	entry := Entry{
		ID:        "test-uuid-1234",
		Timestamp: time.Date(2026, 8, 30, 12, 0, 0, 0, time.UTC),
		RemoteIP:  "127.0.0.1",
		Tool:      "execute_script",
		Plan: &parser.ExecutionPlan{
			EntryPoint: "main.ts",
			Files: []parser.VirtualFile{
				{Name: "main.ts", Content: "console.log('hello');"},
			},
		},
		Result: &executor.Result{
			Success:  true,
			ExitCode: 0,
			Summary:  "Execution finished successfully",
			Stdout:   "hello\n",
		},
	}

	zipPath, err := LogRequest(tmpDir, entry)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, statErr := os.Stat(zipPath); statErr != nil {
		t.Fatalf("zip file does not exist: %v", statErr)
	}

	// Open and inspect ZIP contents
	zr, openErr := zip.OpenReader(zipPath)
	if openErr != nil {
		t.Fatalf("failed to open zip: %v", openErr)
	}
	defer func() { _ = zr.Close() }()

	filesInZip := make(map[string]string)
	for _, f := range zr.File {
		rc, readOpenErr := f.Open()
		if readOpenErr != nil {
			t.Fatalf("failed to open zip entry %s: %v", f.Name, readOpenErr)
		}
		data, readErr := io.ReadAll(rc)
		_ = rc.Close()
		if readErr != nil {
			t.Fatalf("failed to read zip entry %s: %v", f.Name, readErr)
		}
		filesInZip[f.Name] = string(data)
	}

	if _, ok := filesInZip["info.json"]; !ok {
		t.Error("expected info.json in zip")
	}
	if _, ok := filesInZip["response.json"]; !ok {
		t.Error("expected response.json in zip")
	}
	if content, ok := filesInZip["src/main.ts"]; !ok || content != "console.log('hello');" {
		t.Errorf("expected src/main.ts with content, got %q", content)
	}
}

func TestLogRequest_DefaultIDAndTimestamp(t *testing.T) {
	tmpDir := t.TempDir()

	entry := Entry{
		Tool: "execute_script",
		Plan: &parser.ExecutionPlan{
			EntryPoint: "main.ts",
			Files: []parser.VirtualFile{
				{Name: "main.ts", Content: "1+1"},
			},
		},
		Result: &executor.Result{
			Success: true,
		},
	}

	zipPath, err := LogRequest(tmpDir, entry)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := os.Stat(zipPath); err != nil {
		t.Fatalf("expected zip file to exist: %v", err)
	}
}
