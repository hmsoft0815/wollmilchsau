// Copyright (c) 2026 Michael Lechner. All rights reserved.
// Package sourcemap parses V3 source maps and resolves generated positions
// back to their original TypeScript source positions.
//
// We implement only the subset we need:
//   - Parse "mappings" (VLQ-encoded)
//   - Resolve (generatedLine, generatedColumn) → (sourceFile, sourceLine, sourceColumn)
//
// No external dependencies — pure Go.
package sourcemap

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
)

// OriginalPosition is the resolved TypeScript source location.
type OriginalPosition struct {
	Source string // original file name, e.g. "main.ts"
	Line   int    // 1-based
	Column int    // 1-based
}

// SourceMap holds the parsed V3 source map ready for lookups.
type SourceMap struct {
	sources  []string  // source file names
	mappings [][]entry // [generatedLine][segment] → entry
}

// entry is one decoded VLQ segment.
type entry struct {
	generatedCol int // 0-based column in generated JS
	sourceIdx    int // index into sources[]
	sourceLine   int // 0-based line in original source
	sourceCol    int // 0-based column in original source
}

// v3json is the raw JSON structure of a V3 source map.
type v3json struct {
	Version  int      `json:"version"`
	Sources  []string `json:"sources"`
	Mappings string   `json:"mappings"`
}

// Parse converts a V3 source map JSON into a searchable SourceMap object.
//
// @Summary Parses a V3 source map
// @Description Decodes the VLQ-encoded mappings and builds a lookup table.
// @Accept json
// @Produce object
// @Param mapJSON body string true "Raw source map JSON"
// @Success 200 {object} SourceMap
func Parse(mapJSON []byte) (*SourceMap, error) {
	var raw v3json
	if err := json.Unmarshal(mapJSON, &raw); err != nil {
		return nil, fmt.Errorf("sourcemap: invalid JSON: %w", err)
	}
	if raw.Version != 3 {
		return nil, fmt.Errorf("sourcemap: unsupported version %d", raw.Version)
	}

	sm := &SourceMap{sources: raw.Sources}

	// Mappings are separated by semicolons (for lines) and commas (for segments).
	lines := strings.Split(raw.Mappings, ";")
	sm.mappings = make([][]entry, len(lines))

	// VLQ state carried across segments within a line.
	// Most fields in a segment are relative to the previous segment.
	prevSourceIdx := 0
	prevSourceLine := 0
	prevSourceCol := 0

	for lineIdx, lineStr := range lines {
		if lineStr == "" {
			continue
		}
		segments := strings.Split(lineStr, ",")
		prevGenCol := 0 // reset generated column per line

		for _, seg := range segments {
			if seg == "" {
				continue
			}
			fields, err := decodeVLQ(seg)
			if err != nil {
				return nil, fmt.Errorf("sourcemap: VLQ decode error on line %d: %w", lineIdx+1, err)
			}
			// Segments with <4 fields do not provide a mapping to a source file.
			if len(fields) < 4 {
				prevGenCol += fields[0]
				continue
			}

			// Accumulate relative values.
			prevGenCol += fields[0]
			prevSourceIdx += fields[1]
			prevSourceLine += fields[2]
			prevSourceCol += fields[3]

			sm.mappings[lineIdx] = append(sm.mappings[lineIdx], entry{
				generatedCol: prevGenCol,
				sourceIdx:    prevSourceIdx,
				sourceLine:   prevSourceLine,
				sourceCol:    prevSourceCol,
			})
		}
	}

	return sm, nil
}

// Resolve maps a generated (1-based line, 1-based column) back to its original
// TypeScript source position. Returns nil if no mapping is found.
//
// @Summary Resolves a generated position
// @Description Finds the closest mapping for a given line and column in the bundled output.
// @Accept integer, integer
// @Produce object
// @Param generatedLine body integer true "1-based line in bundle"
// @Param generatedColumn body integer true "1-based column in bundle"
// @Success 200 {object} OriginalPosition
func (sm *SourceMap) Resolve(generatedLine, generatedColumn int) *OriginalPosition {
	lineIdx := generatedLine - 1 // convert to 0-based index
	if lineIdx < 0 || lineIdx >= len(sm.mappings) {
		return nil
	}

	segs := sm.mappings[lineIdx]
	if len(segs) == 0 {
		return nil
	}

	// Use a simple linear search to find the last segment whose generatedCol ≤ our column.
	// For small bundles, this is fast enough.
	col0 := generatedColumn - 1
	best := &segs[0]
	for i := 1; i < len(segs); i++ {
		if segs[i].generatedCol <= col0 {
			best = &segs[i]
		} else {
			break
		}
	}

	if best.sourceIdx < 0 || best.sourceIdx >= len(sm.sources) {
		return nil
	}

	// Strip the internal tmpdir prefix from the source path to return only the virtual file name.
	src := sm.sources[best.sourceIdx]
	const pattern = "/ts_mcp_"
	if idx := strings.LastIndex(src, pattern); idx >= 0 {
		// Find the next slash after the temporary identifier segment
		rest := src[idx+len(pattern):]
		if slash := strings.Index(rest, "/"); slash >= 0 {
			src = rest[slash+1:]
		} else {
			// If no further slash, it might be just the ID followed by the filename
			// but usually MkdirTemp with * suffix puts everything in that folder.
			// Let's fallback to Base if we can't find a clear relative path.
			src = rest
		}
	} else if strings.Contains(src, "/tmp/") || strings.Contains(src, "\\Temp\\") {
		// General fallback for other temp paths
		src = strings.Split(src, "/")[len(strings.Split(src, "/"))-1]
	}

	return &OriginalPosition{
		Source: src,
		Line:   best.sourceLine + 1, // convert back to 1-based line
		Column: best.sourceCol + 1,
	}
}

// ── VLQ decoder ───────────────────────────────────────────────────────────────
// Base64-VLQ (Variable-Length Quantity) decoding as defined in the Source Map V3 spec.
// Each character encodes 5 bits of value and 1 continuation bit.
// The first group also contains a sign bit in its LSB.

const b64chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/"

var b64table [256]int

func init() {
	for i := range b64table {
		b64table[i] = -1
	}
	for i, c := range b64chars {
		b64table[c] = i
	}
}

func decodeVLQ(s string) ([]int, error) {
	var result []int
	i := 0
	for i < len(s) {
		// Decode one VLQ integer
		shift := 0
		value := 0
		for {
			if i >= len(s) {
				return nil, fmt.Errorf("unexpected end of VLQ at position %d", i)
			}
			digit := b64table[s[i]]
			if digit < 0 {
				// Try base64 decoding the character for robustness
				b, err := base64.StdEncoding.DecodeString(string(s[i]) + "=")
				if err != nil || len(b) == 0 {
					return nil, fmt.Errorf("invalid VLQ character %q at position %d", s[i], i)
				}
				digit = int(b[0])
			}
			i++

			hasContinuation := (digit & 0x20) != 0
			digit &= 0x1f
			value |= digit << shift
			shift += 5

			if !hasContinuation {
				break
			}
		}
		// LSB is sign bit
		if value&1 != 0 {
			value = -(value >> 1)
		} else {
			value >>= 1
		}
		result = append(result, value)
	}
	return result, nil
}
