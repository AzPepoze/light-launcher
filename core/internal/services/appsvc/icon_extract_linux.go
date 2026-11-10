package appsvc

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// iconPngSize is the largest side of an extracted icon PNG.
const iconPngSize = 128

var (
	extractorTool = newToolResolver([]toolCandidate{
		{name: "wrestool", probe: []string{"--version"}},
		{name: "icoextract", probe: []string{"--help"}},
	})
	converterTool = newToolResolver([]toolCandidate{
		{name: "magick", probe: []string{"-version"}},
		{name: "convert", probe: []string{"-version"}},
	})
)

// extractIcoFile writes the executable's icon to tempDir and returns the .ico path, or "" when there is none.
func extractIcoFile(ctx context.Context, exePath, tempDir string) string {
	tool := extractorTool.get(ctx)
	if tool == "" {
		return ""
	}

	var icoPath string
	var err error
	if tool == "wrestool" {
		icoPath, err = extractWithWrestool(ctx, exePath, tempDir)
	} else {
		icoPath, err = extractWithIcoextract(ctx, exePath, tempDir)
	}
	if err != nil {
		// The tool failed at runtime, so the next call probes for a tool again.
		extractorTool.reset()
		return ""
	}
	return icoPath
}

func extractWithWrestool(ctx context.Context, exePath, tempDir string) (string, error) {
	if err := runTool(ctx, "wrestool", "-x", "--output="+tempDir, exePath); err != nil {
		return "", err
	}
	entries, err := os.ReadDir(tempDir)
	if err != nil {
		return "", err
	}
	for _, entry := range entries {
		if strings.HasSuffix(strings.ToLower(entry.Name()), ".ico") {
			return filepath.Join(tempDir, entry.Name()), nil
		}
	}
	return "", nil
}

func extractWithIcoextract(ctx context.Context, exePath, tempDir string) (string, error) {
	outIco := filepath.Join(tempDir, "icon.ico")
	if err := runTool(ctx, "icoextract", exePath, outIco); err != nil {
		return "", err
	}
	if _, err := os.Stat(outIco); err != nil {
		return "", nil
	}
	return outIco, nil
}

// extractIconPNG converts the largest ICO frame to PNG. It returns nil when no PNG can be made.
func extractIconPNG(ctx context.Context, ico []byte, icoPath, tempDir string) []byte {
	frames := parseIcoFrames(ico)
	if len(frames) == 0 {
		return nil
	}
	largest := largestIcoFrame(frames)
	if largest.png {
		return append([]byte(nil), ico[largest.offset:largest.offset+largest.size]...)
	}

	tool := converterTool.get(ctx)
	if tool == "" {
		return nil
	}
	png, err := convertFrameToPNG(ctx, tool, icoPath, largest.index, tempDir)
	if err != nil {
		converterTool.reset()
		return nil
	}
	return png
}

func convertFrameToPNG(ctx context.Context, tool, icoPath string, frameIndex int, tempDir string) ([]byte, error) {
	outPng := filepath.Join(tempDir, "icon.png")
	source := fmt.Sprintf("%s[%d]", icoPath, frameIndex)
	resize := fmt.Sprintf("%dx%d>", iconPngSize, iconPngSize)
	if err := runTool(ctx, tool, source, "-resize", resize, "PNG32:"+outPng); err != nil {
		return nil, err
	}
	png, err := os.ReadFile(outPng)
	if err != nil {
		return nil, err
	}
	if len(png) == 0 {
		return nil, nil
	}
	return png, nil
}
