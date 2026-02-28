// Copyright (c) 2026 Michael Lechner. All rights reserved.
// Package executor runs JavaScript in an isolated V8 context via v8go.
package executor

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/hmsoft0815/wollmilchsau/internal/sourcemap"
	v8 "rogchap.com/v8go"
)

// Execute runs the provided JavaScript inside a fresh and isolated V8 Isolate.
//
// @Summary Executes JavaScript in V8
// @Description Creates a new V8 context, injects console.log wrappers, and runs the code.
// @Accept string
// @Produce json
// @Param js body string true "Bundled JavaScript code"
// @Param filename body string true "Name of the entry file for stack traces"
// @Param sm body object false "Source map for position resolution"
// @Success 200 {object} Result
func Execute(ctx context.Context, js string, filename string, sm *sourcemap.SourceMap) *Result {
	start := time.Now()
	res := &Result{Diagnostics: []Diagnostic{}}

	iso := v8.NewIsolate()
	defer iso.Dispose()

	var stdout, stderr strings.Builder
	global := v8.NewObjectTemplate(iso)
	consoleTmpl := v8.NewObjectTemplate(iso)

	makeLogger := func(target *strings.Builder) *v8.FunctionTemplate {
		return v8.NewFunctionTemplate(iso, func(info *v8.FunctionCallbackInfo) *v8.Value {
			parts := make([]string, len(info.Args()))
			for i, arg := range info.Args() {
				parts[i] = arg.String()
			}
			target.WriteString(strings.Join(parts, " "))
			target.WriteByte('\n')
			return nil
		})
	}

	_ = consoleTmpl.Set("log", makeLogger(&stdout))
	_ = consoleTmpl.Set("info", makeLogger(&stdout))
	_ = consoleTmpl.Set("warn", makeLogger(&stderr))
	_ = consoleTmpl.Set("error", makeLogger(&stderr))

	v8ctx := v8.NewContext(iso, global)
	defer v8ctx.Close()

	if console, err := consoleTmpl.NewInstance(v8ctx); err == nil {
		_ = v8ctx.Global().Set("console", console)
	}

	if err := InjectPolyfills(iso, v8ctx); err != nil {
		slog.Error("failed to inject polyfills", "err", err)
	}

	if err := InjectArtifactService(iso, v8ctx); err != nil {
		slog.Error("failed to inject artifact service", "err", err)
	}

	watchdogDone := startWatchdog(ctx, iso, 128*1024*1024)
	val, runErr := v8ctx.RunScript(js, filename)
	watchdogDone <- struct{}{}

	res.Stdout = stdout.String()
	res.Stderr = stderr.String()
	res.DurationMs = time.Since(start).Milliseconds()

	if runErr != nil {
		handleExecuteError(ctx, iso, runErr, sm, res)
	} else {
		_ = val
		res.ExitCode = 0
		res.Success = true
		res.Summary = "Execution finished successfully"
	}

	return res
}

func startWatchdog(ctx context.Context, iso *v8.Isolate, maxMemoryBytes uint64) chan struct{} {
	done := make(chan struct{}, 1)
	go func() {
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				iso.TerminateExecution()
				return
			case <-ticker.C:
				if stats := iso.GetHeapStatistics(); stats.UsedHeapSize > maxMemoryBytes {
					iso.TerminateExecution()
					return
				}
			case <-done:
				return
			}
		}
	}()
	return done
}

func handleExecuteError(ctx context.Context, iso *v8.Isolate, err error, sm *sourcemap.SourceMap, res *Result) {
	res.Success = false
	if ctx.Err() != nil {
		res.Stderr += "execution terminated: timeout exceeded\n"
		res.ExitCode = 124
		res.Summary = "Execution timed out"
		return
	}

	diag := extractDiagnostic(err, sm)
	res.Diagnostics = append(res.Diagnostics, diag)
	res.ExitCode = 1

	if strings.Contains(err.Error(), "Execution terminated") {
		stats := iso.GetHeapStatistics()
		const maxMemoryBytes = 128 * 1024 * 1024
		if stats.UsedHeapSize > maxMemoryBytes {
			res.Stderr += fmt.Sprintf("execution terminated: memory limit exceeded (%d MB)\n", stats.UsedHeapSize/1024/1024)
			res.Summary = "Execution terminated: Memory limit exceeded"
		} else {
			res.Summary = "Execution terminated (internal error or forced stop)"
		}
	} else {
		res.Summary = fmt.Sprintf("Runtime Error: %s in %s:%d", diag.Message, diag.Source, diag.Line)
	}
}

// extractDiagnostic converts a V8 error into a Diagnostic, resolving positions
// using the provided source map if available.
func extractDiagnostic(err error, sm *sourcemap.SourceMap) Diagnostic {
	d := Diagnostic{Severity: SeverityError}
	var genLine, genCol int

	jsErr, ok := err.(*v8.JSError)
	if !ok {
		d.Message = err.Error()
		return d
	}
	d.Message = jsErr.Message

	if jsErr.Location == "" {
		return d
	}

	// Parse "filename:line:col" from V8 Location string.
	// We look for the last two colon-separated segments.
	parts := strings.Split(jsErr.Location, ":")
	if len(parts) >= 3 {
		fmt.Sscanf /* nolint:errcheck */ (parts[len(parts)-2], "%d", &genLine)
		fmt.Sscanf /* nolint:errcheck */ (parts[len(parts)-1], "%d", &genCol)
	} else {
		// Fallback if location format is unexpected.
		genLine = 1
		genCol = 1
	}

	d.GeneratedLine = genLine
	d.GeneratedColumn = genCol

	// Attempt to resolve generated line/col back to original TypeScript via source map.
	if sm != nil {
		if orig := sm.Resolve(genLine, genCol); orig != nil {
			d.Source = orig.Source
			d.Line = orig.Line
			d.Column = orig.Column
			return d
		}
	}

	// No source map or no mapping — fallback to generated position within the bundle.
	d.Source = "<bundle>"
	d.Line = genLine
	d.Column = genCol
	return d
}
