// Copyright (c) 2026 Michael Lechner. All rights reserved.
package executor

import (
	"context"
	"encoding/json"
	"log/slog"
	"strings"
	"time"

	mlcartifact "github.com/hmsoft0815/mlcartifact/client"
	v8 "rogchap.com/v8go"
)

// InjectArtifactService adds the global 'artifact' object to the V8 context using a default client.
func InjectArtifactService(iso *v8.Isolate, v8ctx *v8.Context) error {
	cli, err := mlcartifact.NewClient()
	if err != nil {
		return err
	}
	return InjectArtifactServiceWithClient(iso, v8ctx, cli)
}

// InjectArtifactServiceWithClient adds the global 'artifact' object to the V8 context
// using the provided client. Useful for testing.
func InjectArtifactServiceWithClient(iso *v8.Isolate, v8ctx *v8.Context, cli *mlcartifact.Client) error {
	global := v8ctx.Global()
	artObj := v8.NewObjectTemplate(iso)

	_ = artObj.Set("write", v8.NewFunctionTemplate(iso, artifactWriteCallback(iso, v8ctx, cli)))
	_ = artObj.Set("read", v8.NewFunctionTemplate(iso, artifactReadCallback(iso, v8ctx, cli)))
	_ = artObj.Set("list", v8.NewFunctionTemplate(iso, artifactListCallback(iso, v8ctx, cli)))
	_ = artObj.Set("delete", v8.NewFunctionTemplate(iso, artifactDeleteCallback(iso, v8ctx, cli)))

	inst, _ := artObj.NewInstance(v8ctx)
	_ = global.Set("artifact", inst)

	return nil
}

// InjectOpenArtifact adds wollmilchsau.openArtifact(name, mimeType) to the V8 context.
//
// Usage from JS:
//
//	const fh = wollmilchsau.openArtifact("results.csv", "text/csv");
//	fh.write(csvData);
//	const meta = fh.close(); // → { id, uri, name, mimeType, fileSize }
//	console.log(`Saved ${meta.fileSize} bytes → ${meta.uri}`);
//
// When close() is called the buffer is uploaded via gRPC and the ArtifactRef
// is appended to res.CreatedArtifacts so the MCP handler can automatically
// add a resource_link content item to the tool response.
func InjectOpenArtifact(iso *v8.Isolate, v8ctx *v8.Context, cli *mlcartifact.Client, res *Result) error {
	global := v8ctx.Global()

	// Retrieve or create the `wollmilchsau` namespace object.
	wmVal, err := global.Get("wollmilchsau")
	var wmInst *v8.Object
	if err != nil || wmVal.IsUndefined() || wmVal.IsNull() {
		tmpl := v8.NewObjectTemplate(iso)
		wmInst, _ = tmpl.NewInstance(v8ctx)
		_ = global.Set("wollmilchsau", wmInst)
	} else {
		wmInst = wmVal.Object()
	}

	openFn := v8.NewFunctionTemplate(iso, func(info *v8.FunctionCallbackInfo) *v8.Value {
		args := info.Args()
		if len(args) < 1 {
			return wrapError(iso, v8ctx, "wollmilchsau.openArtifact requires (name, optional mimeType)")
		}
		name := args[0].String()
		mimeType := "application/octet-stream"
		if len(args) >= 2 {
			mimeType = args[1].String()
		}

		// Buffer that accumulates write() calls — allocated once per openArtifact call.
		var buf strings.Builder

		handle := v8.NewObjectTemplate(iso)

		// fh.write(data string) — appends to the in-memory buffer.
		_ = handle.Set("write", v8.NewFunctionTemplate(iso, func(info *v8.FunctionCallbackInfo) *v8.Value {
			if len(info.Args()) > 0 {
				buf.WriteString(info.Args()[0].String())
			}
			return v8.Undefined(iso)
		}))

		// fh.close() — uploads buffer, registers ArtifactRef, returns metadata to JS.
		_ = handle.Set("close", v8.NewFunctionTemplate(iso, func(info *v8.FunctionCallbackInfo) *v8.Value {
			content := []byte(buf.String())

			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			resp, err := cli.Write(ctx, name, content,
				mlcartifact.WithMimeType(mimeType),
				mlcartifact.WithSource("wollmilchsau"),
			)
			if err != nil {
				slog.Error("wollmilchsau.openArtifact close() failed", "error", err, "filename", name)
				return wrapError(iso, v8ctx, "wollmilchsau.openArtifact close() failed: "+err.Error())
			}

			ref := ArtifactRef{
				ID:       resp.Id,
				URI:      resp.Uri,
				Name:     resp.Filename,
				MimeType: mimeType,
				FileSize: int64(len(content)),
			}
			// Register in Result so the MCP handler can add resource_link items.
			res.CreatedArtifacts = append(res.CreatedArtifacts, ref)

			return wrapResult(iso, v8ctx, ref)
		}))

		inst, _ := handle.NewInstance(v8ctx)
		return inst.Value
	})

	_ = wmInst.Set("openArtifact", openFn.GetFunction(v8ctx))
	return nil
}

func artifactWriteCallback(iso *v8.Isolate, v8ctx *v8.Context, cli *mlcartifact.Client) v8.FunctionCallback {
	return func(info *v8.FunctionCallbackInfo) *v8.Value {
		args := info.Args()
		if len(args) < 2 {
			return v8.Undefined(iso)
		}

		filename := args[0].String()
		content := []byte(args[1].String())

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		opts := []mlcartifact.WriteOption{}
		if len(args) >= 3 && !args[2].IsUndefined() {
			opts = append(opts, mlcartifact.WithMimeType(args[2].String()))
		}
		if len(args) >= 4 && !args[3].IsUndefined() {
			opts = append(opts, mlcartifact.WithExpiresHours(int32(args[3].Integer())))
		}
		if len(args) >= 5 && !args[4].IsUndefined() {
			opts = append(opts, mlcartifact.WithDescription(args[4].String()))
		}
		if len(args) >= 6 && !args[5].IsUndefined() {
			opts = append(opts, mlcartifact.WithUserID(args[5].String()))
		}

		res, err := cli.Write(ctx, filename, content, opts...)
		if err != nil {
			return wrapError(iso, v8ctx, "artifact.write failed: "+err.Error())
		}

		return wrapResult(iso, v8ctx, res)
	}
}

func artifactReadCallback(iso *v8.Isolate, v8ctx *v8.Context, cli *mlcartifact.Client) v8.FunctionCallback {
	return func(info *v8.FunctionCallbackInfo) *v8.Value {
		args := info.Args()
		if len(args) < 1 {
			return v8.Undefined(iso)
		}

		id := args[0].String()
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		opts := []mlcartifact.ReadOption{}
		if len(args) >= 2 && !args[1].IsUndefined() {
			opts = append(opts, mlcartifact.WithReadUserID(args[1].String()))
		}

		res, err := cli.Read(ctx, id, opts...)
		if err != nil {
			return wrapError(iso, v8ctx, "artifact.read failed: "+err.Error())
		}

		return wrapResult(iso, v8ctx, res)
	}
}

func artifactListCallback(iso *v8.Isolate, v8ctx *v8.Context, cli *mlcartifact.Client) v8.FunctionCallback {
	return func(info *v8.FunctionCallbackInfo) *v8.Value {
		args := info.Args()
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		userID := ""
		if len(args) >= 1 && !args[0].IsUndefined() {
			userID = args[0].String()
		}

		res, err := cli.List(ctx, userID)
		if err != nil {
			return wrapError(iso, v8ctx, "artifact.list failed: "+err.Error())
		}

		return wrapResult(iso, v8ctx, res.Items)
	}
}

func artifactDeleteCallback(iso *v8.Isolate, v8ctx *v8.Context, cli *mlcartifact.Client) v8.FunctionCallback {
	return func(info *v8.FunctionCallbackInfo) *v8.Value {
		args := info.Args()
		if len(args) < 1 {
			return v8.Undefined(iso)
		}

		id := args[0].String()
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		opts := []mlcartifact.DeleteOption{}
		if len(args) >= 2 && !args[1].IsUndefined() {
			opts = append(opts, mlcartifact.WithDeleteUserID(args[1].String()))
		}

		res, err := cli.Delete(ctx, id, opts...)
		if err != nil {
			return wrapError(iso, v8ctx, "artifact.delete failed: "+err.Error())
		}

		return wrapResult(iso, v8ctx, res)
	}
}

func wrapResult(iso *v8.Isolate, ctx *v8.Context, val any) *v8.Value {
	b, _ := json.Marshal(val)
	v, _ := v8.JSONParse(ctx, string(b))
	return v
}

func wrapError(iso *v8.Isolate, ctx *v8.Context, msg string) *v8.Value {
	// For simplicity, we return an object with 'error' property
	obj, _ := v8.NewObjectTemplate(iso).NewInstance(ctx)
	errVal, _ := v8.NewValue(iso, msg)
	_ = obj.Set("error", errVal)
	return obj.Value
}
