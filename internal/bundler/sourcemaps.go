package bundler

import (
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/hmsoft0815/wollmilchsau/internal/sourcemap"
)

// extractSourceMap splits the inline source map comment from the JS bundle.
// esbuild appends: //# sourceMappingURL=data:application/json;base64,<data>
func extractSourceMap(js string) (cleanJS string, sm *sourcemap.SourceMap, err error) {
	const prefix = "//# sourceMappingURL=data:application/json;base64,"
	idx := strings.LastIndex(js, prefix)
	if idx < 0 {
		return js, nil, fmt.Errorf("no inline source map found")
	}
	cleanJS = strings.TrimRight(js[:idx], "\n\r ")
	encoded := strings.TrimSpace(js[idx+len(prefix):])
	mapJSON, decErr := base64.StdEncoding.DecodeString(encoded)
	if decErr != nil {
		return cleanJS, nil, fmt.Errorf("base64 decode: %w", decErr)
	}
	sm, parseErr := sourcemap.Parse(mapJSON)
	if parseErr != nil {
		return cleanJS, nil, parseErr
	}
	return cleanJS, sm, nil
}
