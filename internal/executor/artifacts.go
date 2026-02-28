// Copyright (c) 2026 Michael Lechner. All rights reserved.
package executor

import (
	"context"
	"encoding/json"
	"time"

	"github.com/hmsoft0815/mlcartifact"
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

		res, err := cli.Read(ctx, id)
		if err != nil {
			return wrapError(iso, v8ctx, "artifact.read failed: "+err.Error())
		}

		return wrapResult(iso, v8ctx, res)
	}
}

func artifactListCallback(iso *v8.Isolate, v8ctx *v8.Context, cli *mlcartifact.Client) v8.FunctionCallback {
	return func(info *v8.FunctionCallbackInfo) *v8.Value {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		res, err := cli.List(ctx, "")
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

		res, err := cli.Delete(ctx, id)
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
