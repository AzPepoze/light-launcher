package discordrpc

import (
	"fmt"
	"io"
	"os"
)

// socketCandidates returns the named pipes Discord may listen on.
func socketCandidates() []string {
	candidates := []string{}
	for i := 0; i < 10; i++ {
		candidates = append(candidates, fmt.Sprintf(`\\.\pipe\discord-ipc-%d`, i))
	}
	return candidates
}

// dialSocket opens the Discord named pipe like a file. Pipes have no dial timeout.
func dialSocket(pipePath string) (io.ReadWriteCloser, error) {
	pipe, err := os.OpenFile(pipePath, os.O_RDWR, 0)
	if err != nil {
		return nil, err
	}
	return pipe, nil
}
