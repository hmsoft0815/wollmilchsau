// Copyright (c) 2026 Michael Lechner. All rights reserved.
package executor

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"connectrpc.com/connect"
	"github.com/hmsoft0815/mlcartifact"
	pb "github.com/hmsoft0815/mlcartifact/proto"
	v8 "rogchap.com/v8go"
)

type mockArtifactService struct {
	lastWrite *pb.WriteRequest
	readData  []byte
	listItems []*pb.ArtifactInfo
}

func (m *mockArtifactService) Write(ctx context.Context, req *connect.Request[pb.WriteRequest]) (*connect.Response[pb.WriteResponse], error) {
	m.lastWrite = req.Msg
	return connect.NewResponse(&pb.WriteResponse{
		Id:       "test-id",
		Filename: req.Msg.Filename,
		Uri:      "mcp:///test-id",
	}), nil
}

func (m *mockArtifactService) Read(ctx context.Context, req *connect.Request[pb.ReadRequest]) (*connect.Response[pb.ReadResponse], error) {
	return connect.NewResponse(&pb.ReadResponse{
		Content:  m.readData,
		MimeType: "text/plain",
		Filename: "test.txt",
	}), nil
}

func (m *mockArtifactService) List(ctx context.Context, req *connect.Request[pb.ListRequest]) (*connect.Response[pb.ListResponse], error) {
	return connect.NewResponse(&pb.ListResponse{
		Items: m.listItems,
	}), nil
}

func (m *mockArtifactService) Delete(ctx context.Context, req *connect.Request[pb.DeleteRequest]) (*connect.Response[pb.DeleteResponse], error) {
	return connect.NewResponse(&pb.DeleteResponse{
		Deleted: true,
	}), nil
}

func TestArtifactBridge(t *testing.T) {
	mockSvc := &mockArtifactService{
		readData: []byte("hello artifact"),
		listItems: []*pb.ArtifactInfo{
			{Id: "1", Filename: "f1.txt"},
		},
	}

	cli := mlcartifact.NewClientWithService(mockSvc)

	iso := v8.NewIsolate()
	defer iso.Dispose()

	v8ctx := v8.NewContext(iso)
	defer v8ctx.Close()

	if err := InjectArtifactServiceWithClient(iso, v8ctx, cli); err != nil {
		t.Fatalf("Failed to inject artifact service: %v", err)
	}

	// Test Write
	t.Run("write", func(t *testing.T) {
		js := `
			(() => {
				const res = artifact.write("test.txt", "content", "text/plain", 24, "This is a test description");
				return JSON.stringify(res);
			})()
		`
		val, err := v8ctx.RunScript(js, "test_write.js")
		if err != nil {
			t.Fatalf("Script failed: %v", err)
		}
		// ... parser logic ...
		resStr := val.String()
		var res map[string]any
		if err := json.Unmarshal([]byte(resStr), &res); err != nil {
			t.Fatalf("Failed to parse response: %v", err)
		}

		if res["id"] != "test-id" {
			t.Errorf("Expected id test-id, got %v", res["id"])
		}
		if mockSvc.lastWrite.Filename != "test.txt" {
			t.Errorf("Expected filename test.txt, got %s", mockSvc.lastWrite.Filename)
		}
		if mockSvc.lastWrite.Description != "This is a test description" {
			t.Errorf("Expected description 'This is a test description', got '%s'", mockSvc.lastWrite.Description)
		}
	})

	// Test List
	t.Run("list", func(t *testing.T) {
		js := `
			(() => {
				const items = artifact.list();
				return JSON.stringify(items);
			})()
		`
		val, err := v8ctx.RunScript(js, "test_list.js")
		if err != nil {
			t.Fatalf("Script failed: %v", err)
		}

		resStr := val.String()
		if !strings.Contains(resStr, "f1.txt") {
			t.Errorf("Expected list to contain f1.txt, got %s", resStr)
		}
	})
}
