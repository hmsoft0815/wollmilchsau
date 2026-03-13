// Copyright (c) 2026 Michael Lechner. All rights reserved.
package executor

import (
	"context"
	"encoding/json"
	"os"
	"testing"
	"time"

	mlcartifact "github.com/hmsoft0815/mlcartifact/client"
	v8 "rogchap.com/v8go"
)

func TestArtifactIntegration(t *testing.T) {
	addr := os.Getenv("ARTIFACT_GRPC_ADDR")
	if addr == "" {
		t.Skip("Skipping integration test: ARTIFACT_GRPC_ADDR not set")
	}

	cli, err := mlcartifact.NewClientWithAddr(addr)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer cli.Close()

	iso := v8.NewIsolate()
	defer iso.Dispose()
	v8ctx := v8.NewContext(iso)
	defer v8ctx.Close()

	if err := InjectArtifactServiceWithClient(iso, v8ctx, cli); err != nil {
		t.Fatalf("Failed to inject artifact service: %v", err)
	}

	t.Run("Write and Read Integration", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		filename := "integration_test.txt"
		content := "Real artifact content " + time.Now().Format(time.RFC3339)

		// 1. Write via JS
		jsWrite := `JSON.stringify(artifact.write("` + filename + `", "` + content + `", "text/plain"))`
		val, err := v8ctx.RunScript(jsWrite, "test_write.js")
		if err != nil {
			t.Fatalf("JS Write failed: %v", err)
		}
		t.Logf("JS Write result: %s", val.String())

		// Parse ID from result
		var writeRes struct {
			ID string `json:"id"`
		}
		if err := wrapResultToStruct(v8ctx, val, &writeRes); err != nil {
			t.Fatalf("Failed to parse write result: %v", err)
		}

		if writeRes.ID == "" {
			t.Fatal("Artifact ID is empty")
		}

		// 2. Read via Go Client to verify
		readRes, err := cli.Read(ctx, writeRes.ID)
		if err != nil {
			t.Fatalf("Go Read failed: %v", err)
		}

		if string(readRes.Content) != content {
			t.Errorf("Content mismatch. Expected %q, got %q", content, string(readRes.Content))
		}

		// 3. Delete
		_, err = cli.Delete(ctx, writeRes.ID)
		if err != nil {
			t.Errorf("Cleanup failed: %v", err)
		}
	})

	t.Run("Write and Read with UserID", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		filename := "user_test.txt"
		content := "User scoped content"
		userID := "test_user_123"

		// 1. Write via JS with UserID (6th argument)
		jsWrite := `JSON.stringify(artifact.write("` + filename + `", "` + content + `", "text/plain", 1, "test desc", "` + userID + `"))`
		val, err := v8ctx.RunScript(jsWrite, "test_user_write.js")
		if err != nil {
			t.Fatalf("JS Write failed: %v", err)
		}

		var writeRes struct {
			ID    string `json:"id"`
			Error string `json:"error"`
		}
		if err := wrapResultToStruct(v8ctx, val, &writeRes); err != nil {
			t.Fatalf("Failed to parse write result: %v", err)
		}

		if writeRes.Error != "" {
			t.Fatalf("Artifact write error: %s", writeRes.Error)
		}

		// 2. Read via JS with UserID (2nd argument)
		jsRead := `JSON.stringify(artifact.read("` + writeRes.ID + `", "` + userID + `"))`
		readVal, err := v8ctx.RunScript(jsRead, "test_user_read.js")
		if err != nil {
			t.Fatalf("JS Read failed: %v", err)
		}

		var readRes struct {
			Content string `json:"content"`
			Error   string `json:"error"`
		}
		// Note: content comes as base64 in JSON result if using our wrapResult (which uses json.Marshal on pb.ReadResponse)
		// Actually pb.ReadResponse.Content is []byte, which Marshals to base64.
		if err := wrapResultToStruct(v8ctx, readVal, &readRes); err != nil {
			t.Fatalf("Failed to parse read result: %v", err)
		}

		if readRes.Error != "" {
			t.Fatalf("Artifact read error: %s", readRes.Error)
		}

		// 3. Cleanup
		_, _ = cli.Delete(ctx, writeRes.ID, mlcartifact.WithDeleteUserID(userID))
	})
}

func wrapResultToStruct(ctx *v8.Context, v *v8.Value, target any) error {
	// Our JS callbacks in artifacts.go currently return JSON.stringify results
	// but wrapResult actually returns a parsed JSON object.
	// However, if the JS returns JSON.stringify(res), then v.String() IS the JSON.
	return json.Unmarshal([]byte(v.String()), target)
}
