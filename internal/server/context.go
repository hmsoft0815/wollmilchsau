// Copyright (c) 2026 Michael Lechner. All rights reserved.
package server

import (
	"context"
)

type contextKey string

const (
	ContextKeyRemoteIP contextKey = "remote_ip"
)

// WithRemoteIP adds the remote IP to the context. (SSE only)
func WithRemoteIP(ctx context.Context, ip string) context.Context {
	return context.WithValue(ctx, ContextKeyRemoteIP, ip)
}

// GetRemoteIP extracts the remote IP from the context.
func GetRemoteIP(ctx context.Context) string {
	if ip, ok := ctx.Value(ContextKeyRemoteIP).(string); ok {
		return ip
	}
	return "unknown"
}
