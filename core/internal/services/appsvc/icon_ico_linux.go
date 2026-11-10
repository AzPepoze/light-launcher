package appsvc

import (
	"bytes"
	"encoding/binary"
)

// icoFrame is one image inside an .ico file.
type icoFrame struct {
	index  int
	width  int
	height int
	size   int
	offset int
	png    bool
}

var pngSignature = []byte{0x89, 'P', 'N', 'G', 0x0d, 0x0a, 0x1a, 0x0a}

const (
	icoHeaderSize = 6
	icoEntrySize  = 16
)

// parseIcoFrames reads the frame table. Broken entries are skipped; a bad header gives no frames.
func parseIcoFrames(ico []byte) []icoFrame {
	if len(ico) < icoHeaderSize || binary.LittleEndian.Uint16(ico[0:]) != 0 || binary.LittleEndian.Uint16(ico[2:]) != 1 {
		return nil
	}
	count := int(binary.LittleEndian.Uint16(ico[4:]))
	frames := []icoFrame{}
	for i := 0; i < count; i++ {
		entry := icoHeaderSize + i*icoEntrySize
		if entry+icoEntrySize > len(ico) {
			break
		}
		size := uint64(binary.LittleEndian.Uint32(ico[entry+8:]))
		offset := uint64(binary.LittleEndian.Uint32(ico[entry+12:]))
		if size < 8 || offset+size > uint64(len(ico)) {
			continue
		}
		frames = append(frames, newIcoFrame(ico, i, entry, int(offset), int(size)))
	}
	return frames
}

func newIcoFrame(ico []byte, index, entry, offset, size int) icoFrame {
	return icoFrame{
		index:  index,
		width:  icoDimension(ico[entry]),
		height: icoDimension(ico[entry+1]),
		size:   size,
		offset: offset,
		png:    bytes.Equal(ico[offset:offset+8], pngSignature),
	}
}

// icoDimension maps the stored byte to pixels, where 0 means 256.
func icoDimension(value byte) int {
	if value == 0 {
		return 256
	}
	return int(value)
}

// largestIcoFrame returns the first frame with the largest area. frames must not be empty.
func largestIcoFrame(frames []icoFrame) icoFrame {
	best := frames[0]
	for _, frame := range frames[1:] {
		if frame.width*frame.height > best.width*best.height {
			best = frame
		}
	}
	return best
}
