// Copyright (c) 2026 Michael Lechner. All rights reserved.
package requestlog

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/hmsoft0815/wollmilchsau/internal/executor"
	"github.com/hmsoft0815/wollmilchsau/internal/parser"
)

// Entry captures all information about a single request for archiving.
type Entry struct {
	ID        string                `json:"id"`
	Timestamp time.Time             `json:"timestamp"`
	RemoteIP  string                `json:"remoteIp"`
	Tool      string                `json:"tool"`
	Plan      *parser.ExecutionPlan `json:"plan"`
	Result    *executor.Result      `json:"result"`
}

// LogRequest bundles the request and response into a ZIP file.
func LogRequest(dir string, entry Entry) (string, error) {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}

	if entry.ID == "" {
		entry.ID = uuid.New().String()
	}
	if entry.Timestamp.IsZero() {
		entry.Timestamp = time.Now()
	}

	filename := fmt.Sprintf("req_%s_%s.zip", entry.Timestamp.Format("20060102_150405"), entry.ID[:8])
	path := filepath.Join(dir, filename)

	f, err := os.Create(path)
	if err != nil {
		return "", err
	}
	defer f.Close() // nolint:errcheck

	zw := zip.NewWriter(f)
	defer zw.Close() // nolint:errcheck

	// 1. Add metadata.json
	metaJSON, _ := json.MarshalIndent(entry, "", "  ")
	if err := addFileToZip(zw, "info.json", metaJSON); err != nil {
		return path, err
	}

	// 2. Add source files
	for _, vf := range entry.Plan.Files {
		if err := addFileToZip(zw, filepath.Join("src", vf.Name), []byte(vf.Content)); err != nil {
			return path, err
		}
	}

	// 3. Add response.json
	respJSON, _ := json.MarshalIndent(entry.Result, "", "  ")
	if err := addFileToZip(zw, "response.json", respJSON); err != nil {
		return path, err
	}

	return path, nil
}

func addFileToZip(zw *zip.Writer, name string, content []byte) error {
	w, err := zw.Create(name)
	if err != nil {
		return err
	}
	_, err = w.Write(content)
	return err
}
