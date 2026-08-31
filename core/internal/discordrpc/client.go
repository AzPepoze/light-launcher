package discordrpc

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

const (
	opHandshake = 0
	opFrame     = 1
)

type Client struct {
	conn net.Conn
}

type packetHeader struct {
	Opcode uint32
	Length uint32
}

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

func Connect(clientID string) (*Client, error) {
	if clientID == "" {
		return nil, errors.New("discord client id is empty")
	}

	var lastErr error
	for _, socketPath := range socketCandidates() {
		conn, err := net.DialTimeout("unix", socketPath, 250*time.Millisecond)
		if err != nil {
			lastErr = err
			continue
		}

		client := &Client{conn: conn}
		if err := client.writePacket(opHandshake, map[string]any{
			"v":         1,
			"client_id": clientID,
		}); err != nil {
			_ = conn.Close()
			lastErr = err
			continue
		}

		// Read READY reply before SET_ACTIVITY.
		_ = conn.SetReadDeadline(time.Now().Add(2 * time.Second))
		if _, _, err := client.readPacket(); err != nil {
			_ = conn.Close()
			lastErr = err
			continue
		}
		_ = conn.SetReadDeadline(time.Time{})
		return client, nil
	}

	if lastErr == nil {
		lastErr = errors.New("discord IPC socket not found")
	}
	return nil, lastErr
}

func (c *Client) SetActivity(gameName string, startedAt time.Time) error {
	if c == nil || c.conn == nil {
		return errors.New("discord rpc is not connected")
	}

	return c.writePacket(opFrame, map[string]any{
		"cmd": "SET_ACTIVITY",
		"args": map[string]any{
			"pid": os.Getpid(),
			"activity": map[string]any{
				"details": "Playing " + gameName,
				"state":   "Launched with LightLauncher",
				"timestamps": map[string]any{
					"start": startedAt.Unix(),
				},
			},
		},
		"nonce": strconv.FormatInt(time.Now().UnixNano(), 10),
	})
}

func (c *Client) ClearActivity() error {
	if c == nil || c.conn == nil {
		return nil
	}
	return c.writePacket(opFrame, map[string]any{
		"cmd": "SET_ACTIVITY",
		"args": map[string]any{
			"pid":      os.Getpid(),
			"activity": nil,
		},
		"nonce": strconv.FormatInt(time.Now().UnixNano(), 10),
	})
}

func (c *Client) Close() error {
	if c == nil || c.conn == nil {
		return nil
	}
	err := c.conn.Close()
	c.conn = nil
	return err
}

func (c *Client) writePacket(opcode uint32, payload any) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	header := make([]byte, 8)
	binary.LittleEndian.PutUint32(header[0:4], opcode)
	binary.LittleEndian.PutUint32(header[4:8], uint32(len(body)))
	if _, err := c.conn.Write(header); err != nil {
		return err
	}
	_, err = c.conn.Write(body)
	return err
}

func (c *Client) readPacket() (uint32, []byte, error) {
	headerBytes := make([]byte, 8)
	if _, err := io.ReadFull(c.conn, headerBytes); err != nil {
		return 0, nil, err
	}
	header := packetHeader{
		Opcode: binary.LittleEndian.Uint32(headerBytes[0:4]),
		Length: binary.LittleEndian.Uint32(headerBytes[4:8]),
	}
	if header.Length > 4*1024*1024 {
		return 0, nil, fmt.Errorf("discord rpc packet too large: %d", header.Length)
	}
	body := make([]byte, header.Length)
	if _, err := io.ReadFull(c.conn, body); err != nil {
		return 0, nil, err
	}
	return header.Opcode, body, nil
}
