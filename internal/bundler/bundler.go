// Copyright (c) 2026 Michael Lechner. All rights reserved.
// Package bundler transpiles and bundles TypeScript using esbuild in-process.
package bundler

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/evanw/esbuild/pkg/api"
	"github.com/hmsoft0815/wollmilchsau/internal/parser"
	"github.com/hmsoft0815/wollmilchsau/internal/sourcemap"
)

// BundleResult holds the output of a bundle operation.
type BundleResult struct {
	JS        string               // bundled JavaScript (source map comment stripped)
	SourceMap *sourcemap.SourceMap // parsed source map for error resolution; may be nil
	Warnings  []BundleMessage      // structured warnings
}

// BundleError is returned when esbuild reports compile errors.
// Positions already refer to the original TypeScript source.
type BundleError struct {
	Messages []BundleMessage
}

// BundleMessage is a single esbuild diagnostic with original TS position.
type BundleMessage struct {
	Text   string // error or warning message
	Source string // original .ts file name (e.g., "main.ts")
	Line   int    // 1-based line number in original source
	Column int    // 1-based column number in original source
}

func (e *BundleError) Error() string {
	if len(e.Messages) == 0 {
		return "bundle failed"
	}
	m := e.Messages[0]
	if m.Source != "" && m.Line > 0 {
		return fmt.Sprintf("%s:%d:%d: %s", m.Source, m.Line, m.Column, m.Text)
	}
	return fmt.Sprintf("bundle failed: %s", m.Text)
}

// Bundle transpiles and bundles a project from an ExecutionPlan.
// It uses a temporary directory to store virtual files for esbuild.
//
// @Summary Bundles a TypeScript project
// @Description Writes virtual files to a temporary directory and uses esbuild to bundle them.
// @Accept object
// @Produce object
// @Param plan body parser.ExecutionPlan true "Execution plan containing virtual files"
// @Success 200 {object} BundleResult
func Bundle(plan *parser.ExecutionPlan) (*BundleResult, error) {
	// 1. Setup temporary directory for esbuild.
	// Since esbuild works on files, we materialize our virtual project here.
	tmpDir, err := os.MkdirTemp("", "ts_mcp_*")
	if err != nil {
		return nil, fmt.Errorf("creating temp dir: %w", err)
	}
	// Always cleanup the temp files after bundling.
	defer os.RemoveAll(tmpDir)

	// 2. Write each virtual file to the temp directory.
	for _, vf := range plan.Files {
		dest := filepath.Join(tmpDir, filepath.FromSlash(vf.Name))
		if err := os.MkdirAll(filepath.Dir(dest), 0o700); err != nil {
			return nil, fmt.Errorf("creating dir for %q: %w", vf.Name, err)
		}
		if err := os.WriteFile(dest, []byte(vf.Content), 0o600); err != nil {
			return nil, fmt.Errorf("writing %q: %w", vf.Name, err)
		}
	}

	entryPath := filepath.Join(tmpDir, filepath.FromSlash(plan.EntryPoint))

	// 3. Invoke esbuild in-process to bundle the TypeScript project.
	result := api.Build(api.BuildOptions{
		EntryPoints:    []string{entryPath},
		Bundle:         true,
		Platform:       api.PlatformNode,
		Target:         api.ES2020,
		Format:         api.FormatIIFE,
		GlobalName:     "__entry__", // ensure the bundle is wrapped in an IIFE
		Write:          false,       // don't write to disk, keep result in memory
		LogLevel:       api.LogLevelSilent,
		Sourcemap:      api.SourceMapInline,       // append base64 source map to result JS
		SourcesContent: api.SourcesContentExclude, // don't embed original TS source in map
	})

	warnings := make([]BundleMessage, 0, len(result.Warnings))
	for _, w := range result.Warnings {
		warnings = append(warnings, toBundleMessage(w, tmpDir))
	}

	// 4. Handle compilation errors.
	if len(result.Errors) > 0 {
		msgs := make([]BundleMessage, 0, len(result.Errors))
		for _, e := range result.Errors {
			msgs = append(msgs, toBundleMessage(e, tmpDir))
		}
		return nil, &BundleError{Messages: msgs}
	}

	if len(result.OutputFiles) == 0 {
		return nil, &BundleError{Messages: []BundleMessage{{Text: "esbuild produced no output"}}}
	}

	// 5. Extract and parse the inline source map.
	rawJS := string(result.OutputFiles[0].Contents)
	js, sm, _ := extractSourceMap /* nolint:errcheck */ (rawJS) // failures here are non-fatal

	return &BundleResult{
		JS:        js,
		SourceMap: sm,
		Warnings:  warnings,
	}, nil
}

func toBundleMessage(msg api.Message, tmpDir string) BundleMessage {
	bm := BundleMessage{Text: msg.Text}
	if msg.Location != nil {
		bm.Source = stripTmpDir(msg.Location.File, tmpDir)
		bm.Line = msg.Location.Line
		bm.Column = msg.Location.Column + 1 // esbuild is 0-based
	}
	return bm
}

func stripTmpDir(path, tmpDir string) string {
	// 1. Try absolute path comparison
	absPath, err1 := filepath.Abs(path)
	absTmp, err2 := filepath.Abs(tmpDir)
	if err1 == nil && err2 == nil {
		if strings.HasPrefix(absPath, absTmp) {
			rel := strings.TrimPrefix(absPath, absTmp)
			return strings.TrimPrefix(rel, string(filepath.Separator))
		}
	}

	// 2. Fallback: search for our unique prefix pattern in the path
	const pattern = "ts_mcp_"
	if idx := strings.Index(path, pattern); idx >= 0 {
		// Find the next slash after the pattern segment
		rest := path[idx:]
		if firstSlash := strings.Index(rest, string(filepath.Separator)); firstSlash >= 0 {
			return rest[firstSlash+1:]
		}
	}

	// 3. Last resort: just return the base name if it's still look like a temp path
	return filepath.Base(path)
}
