package discordrpc

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"time"
)

const (
	opHandshake = 0
	opFrame     = 1
)

type Client struct {
	conn io.ReadWriteCloser
}

type packetHeader struct {
	Opcode uint32
	Length uint32
}

func Connect(clientID string) (*Client, error) {
	if clientID == "" {
		return nil, errors.New("discord client id is empty")
	}

	var lastErr error
	for _, socketPath := range socketCandidates() {
		conn, err := dialSocket(socketPath)
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
		setReadDeadline(conn, time.Now().Add(2*time.Second))
		if _, _, err := client.readPacket(); err != nil {
			_ = conn.Close()
			lastErr = err
			continue
		}
		setReadDeadline(conn, time.Time{})
		return client, nil
	}

	if lastErr == nil {
		lastErr = errors.New("discord IPC socket not found")
	}
	return nil, lastErr
}

// setReadDeadline sets the read deadline when the transport supports it. Windows pipes do not.
func setReadDeadline(conn io.ReadWriteCloser, deadline time.Time) {
	if deadlineConn, ok := conn.(interface{ SetReadDeadline(time.Time) error }); ok {
		_ = deadlineConn.SetReadDeadline(deadline)
	}
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
