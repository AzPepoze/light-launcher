package discordrpc

import (
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"time"
)

// socketCandidates returns the Unix sockets Discord may listen on.
func socketCandidates() []string {
	bases := []string{}
	if runtimeDir := os.Getenv("XDG_RUNTIME_DIR"); runtimeDir != "" {
		bases = append(bases, runtimeDir)
	}
	if tmpDir := os.Getenv("TMPDIR"); tmpDir != "" {
		bases = append(bases, tmpDir)
	}
	bases = append(bases, "/tmp")

	seen := map[string]bool{}
	candidates := []string{}
	for _, base := range bases {
		for i := 0; i < 10; i++ {
			candidate := filepath.Join(base, fmt.Sprintf("discord-ipc-%d", i))
			if !seen[candidate] {
				seen[candidate] = true
				candidates = append(candidates, candidate)
			}
		}
	}
	return candidates
}

func dialSocket(socketPath string) (io.ReadWriteCloser, error) {
	conn, err := net.DialTimeout("unix", socketPath, 250*time.Millisecond)
	if err != nil {
		return nil, err
	}
	return conn, nil
}
