package browserassist

import (
	"bufio"
	"errors"
	"io"
)

const (
	readBuffer     = 32 << 10
	maxLineBytes   = 64 << 10
	maxStreamBytes = 1 << 20
)

var (
	errLineTooLong = errors.New("shell output line is too long")
	errStreamLimit = errors.New("shell output is too large")
)

// lineReader reads NDJSON lines with a per-line and whole-stream budget.
type lineReader struct {
	buf       *bufio.Reader
	remaining int64
}

func newLineReader(r io.Reader) *lineReader {
	return &lineReader{buf: bufio.NewReaderSize(r, readBuffer), remaining: maxStreamBytes}
}

// next returns one complete line with its newline. A partial last line is dropped.
func (r *lineReader) next() ([]byte, error) {
	var line []byte
	tooLong := false
	for {
		chunk, err := r.buf.ReadSlice('\n')
		if limitErr := r.spend(len(chunk)); limitErr != nil {
			return nil, limitErr
		}
		if err == nil {
			if tooLong || len(line)+len(chunk) > maxLineBytes {
				return nil, errLineTooLong
			}
			return append(line, chunk...), nil
		}
		if errors.Is(err, bufio.ErrBufferFull) {
			if len(line)+len(chunk) > maxLineBytes {
				tooLong = true
			} else if !tooLong {
				line = append(line, chunk...)
			}
			continue
		}
		if errors.Is(err, io.EOF) {
			return nil, io.EOF
		}
		return nil, err
	}
}

func (r *lineReader) spend(n int) error {
	r.remaining -= int64(n)
	if r.remaining < 0 {
		return errStreamLimit
	}
	return nil
}
