// Copyright (c) 2026 Michael Lechner. All rights reserved.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	mcpserver "github.com/hmsoft0815/wollmilchsau/internal/server"
	"github.com/mark3labs/mcp-go/server"
	v8 "rogchap.com/v8go"
)

func main() {
	versionFlag := flag.Bool("version", false, "Show version information")
	dumpFlag := flag.Bool("dump", false, "Dump MCP tool schema")
	addrFlag := flag.String("addr", "", "Listen address for SSE (e.g. ':8080'). If empty, uses stdio.")
	logDirFlag := flag.String("log-dir", "", "Directory to store complete request/response ZIP archives (optional)")
	enableArtifactsFlag := flag.Bool("enable-artifacts", false, "Enable the artifact service integration (artifact global object and execute_artifact tool)")
	flag.Parse()

	if *versionFlag {
		fmt.Printf("wollmilchsau version: %s\n", mcpserver.ServerVersion)
		fmt.Printf("V8 version:          %s\n", v8.Version())
		fmt.Printf("esbuild version:     %s\n", "v0.24.2")
		return
	}

	if *dumpFlag {
		tools := mcpserver.GetTools(*enableArtifactsFlag)
		b, _ := json.MarshalIndent(tools, "", "  ")
		fmt.Println(string(b))
		return
	}

	ws := mcpserver.New(*logDirFlag, *enableArtifactsFlag)

	if *addrFlag != "" {
		// SSE Mode
		sse := server.NewSSEServer(ws.MCPServer,
			server.WithBaseURL(fmt.Sprintf("http://localhost%s", *addrFlag)),
			server.WithSSEContextFunc(func(ctx context.Context, r *http.Request) context.Context {
				return mcpserver.WithRemoteIP(ctx, r.RemoteAddr)
			}),
		)

		slog.Info("SSE server started", "addr", *addrFlag, "name", mcpserver.ServerName, "log_dir", *logDirFlag)
		if err := http.ListenAndServe(*addrFlag, sse); err != nil {
			slog.Error("http server failed", "err", err)
			os.Exit(1)
		}
	} else {
		// Stdio Mode
		slog.Info("stdio server started", "name", mcpserver.ServerName, "version", mcpserver.ServerVersion, "log_dir", *logDirFlag)

		err := server.ServeStdio(ws.MCPServer, server.WithStdioContextFunc(func(ctx context.Context) context.Context {
			return mcpserver.WithRemoteIP(ctx, "stdio")
		}))

		if err != nil {
			slog.Error("fatal error", "err", err)
			os.Exit(1)
		}
	}
}
