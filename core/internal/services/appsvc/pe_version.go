package appsvc

import (
	"debug/pe"
	"encoding/binary"
	"io"
	"strings"
	"unicode/utf16"
)

const (
	imageDirectoryEntryResource = 2
	rtVersion                   = 16 // resource type of the VS_VERSIONINFO block
	resourceDirHeaderSize       = 16
	resourceEntrySize           = 8
	resourceDataEntrySize       = 16
	maxResourceEntries          = 256
	maxVersionBytes             = 64 << 10
	maxVersionDepth             = 8
	versionBlockHeaderSize      = 6
	resourceIsDirectory         = 0x80000000
	resourceIsNamed             = 0x80000000
)

// versionStrings holds the names an executable declares in its VERSIONINFO resource.
type versionStrings struct {
	productName     string
	fileDescription string
}

// readVersionStrings reads ProductName and FileDescription from the executable's version resource.
// It returns empty strings for files that are not PE files, have no version resource or are corrupt.
func readVersionStrings(r io.ReaderAt) (out versionStrings) {
	defer func() {
		// debug/pe is not hardened against every corrupt file, so a bad file is treated as having no names.
		if recover() != nil {
			out = versionStrings{}
		}
	}()

	file, err := pe.NewFile(r)
	if err != nil {
		return versionStrings{}
	}
	table, ok := newResourceTable(r, file)
	if !ok {
		return versionStrings{}
	}
	data, ok := table.versionData()
	if !ok {
		return versionStrings{}
	}
	collectVersionStrings(data, &out, 0)
	return out
}

// resourceTable reads the resource directory of a PE file. Offsets inside it are relative to root.
type resourceTable struct {
	r    io.ReaderAt
	file *pe.File
	root uint32
}

func newResourceTable(r io.ReaderAt, file *pe.File) (resourceTable, bool) {
	dirs, ok := dataDirectories(file)
	if !ok {
		return resourceTable{}, false
	}
	entry := dirs[imageDirectoryEntryResource]
	if entry.VirtualAddress == 0 || entry.Size == 0 {
		return resourceTable{}, false
	}
	return resourceTable{r: r, file: file, root: entry.VirtualAddress}, true
}

func dataDirectories(file *pe.File) ([16]pe.DataDirectory, bool) {
	switch header := file.OptionalHeader.(type) {
	case *pe.OptionalHeader32:
		return header.DataDirectory, true
	case *pe.OptionalHeader64:
		return header.DataDirectory, true
	default:
		return [16]pe.DataDirectory{}, false
	}
}

// resourceEntry is one row of a resource directory: a child directory or a data entry.
type resourceEntry struct {
	id     uint32
	offset uint32
	isDir  bool
	named  bool
}

// versionData returns the first VERSIONINFO block: the first language of the first version resource.
func (t resourceTable) versionData() ([]byte, bool) {
	typeEntry, ok := t.child(0, rtVersion, false)
	if !ok || !typeEntry.isDir {
		return nil, false
	}
	nameEntry, ok := t.child(typeEntry.offset, 0, true)
	if !ok || !nameEntry.isDir {
		return nil, false
	}
	langEntry, ok := t.child(nameEntry.offset, 0, true)
	if !ok || langEntry.isDir {
		return nil, false
	}
	return t.leafData(langEntry.offset)
}

// child finds an entry by id in the directory at rel. With anyID it takes the first unnamed entry.
func (t resourceTable) child(rel, id uint32, anyID bool) (resourceEntry, bool) {
	entries, ok := t.entries(rel)
	if !ok {
		return resourceEntry{}, false
	}
	for _, entry := range entries {
		if entry.named {
			continue
		}
		if anyID || entry.id == id {
			return entry, true
		}
	}
	return resourceEntry{}, false
}

func (t resourceTable) entries(rel uint32) ([]resourceEntry, bool) {
	header, ok := t.readRVA(t.root+rel, resourceDirHeaderSize)
	if !ok {
		return nil, false
	}
	count := int(binary.LittleEndian.Uint16(header[12:])) + int(binary.LittleEndian.Uint16(header[14:]))
	count = min(count, maxResourceEntries)
	raw, ok := t.readRVA(t.root+rel+resourceDirHeaderSize, uint32(count*resourceEntrySize))
	if !ok {
		return nil, false
	}
	entries := make([]resourceEntry, count)
	for i := range entries {
		row := raw[i*resourceEntrySize:]
		id := binary.LittleEndian.Uint32(row)
		offset := binary.LittleEndian.Uint32(row[4:])
		entries[i] = resourceEntry{
			id:     id &^ resourceIsNamed,
			offset: offset &^ resourceIsDirectory,
			isDir:  offset&resourceIsDirectory != 0,
			named:  id&resourceIsNamed != 0,
		}
	}
	return entries, true
}

// leafData reads the bytes of the data entry at rel, if it is a sane size.
func (t resourceTable) leafData(rel uint32) ([]byte, bool) {
	entry, ok := t.readRVA(t.root+rel, resourceDataEntrySize)
	if !ok {
		return nil, false
	}
	dataRVA := binary.LittleEndian.Uint32(entry)
	size := binary.LittleEndian.Uint32(entry[4:])
	if size == 0 || size > maxVersionBytes {
		return nil, false
	}
	return t.readRVA(dataRVA, size)
}

// readRVA reads n bytes at a relative virtual address, or reports false when they are not in the file.
func (t resourceTable) readRVA(rva, n uint32) ([]byte, bool) {
	offset, ok := rvaToFileOffset(t.file, rva)
	if !ok {
		return nil, false
	}
	buf := make([]byte, n)
	got, _ := t.r.ReadAt(buf, offset)
	if got < len(buf) {
		return nil, false
	}
	return buf, true
}

// rvaToFileOffset maps a relative virtual address to its place in the file through the section table.
func rvaToFileOffset(file *pe.File, rva uint32) (int64, bool) {
	for _, section := range file.Sections {
		size := max(section.VirtualSize, section.Size)
		if rva >= section.VirtualAddress && rva-section.VirtualAddress < size {
			return int64(section.Offset) + int64(rva-section.VirtualAddress), true
		}
	}
	return 0, false
}

// collectVersionStrings walks the VERSIONINFO tree and keeps the first non-empty ProductName and FileDescription.
func collectVersionStrings(data []byte, out *versionStrings, depth int) {
	if depth > maxVersionDepth {
		return
	}
	for pos := 0; pos < len(data); {
		block, ok := readVersionBlock(data, pos)
		if !ok {
			return
		}
		if block.isText {
			out.keep(block.key, block.text())
		}
		collectVersionStrings(block.children, out, depth+1)
		pos = block.next
	}
}

// keep stores a version string under its key when that field is still empty.
func (v *versionStrings) keep(key, text string) {
	text = strings.TrimSpace(text)
	if text == "" {
		return
	}
	switch key {
	case "ProductName":
		v.productName = firstNonEmpty(v.productName, text)
	case "FileDescription":
		v.fileDescription = firstNonEmpty(v.fileDescription, text)
	}
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}

// versionBlock is one node of the VERSIONINFO tree.
type versionBlock struct {
	key      string
	isText   bool
	value    []byte
	children []byte
	next     int
}

// text returns the value of a text block as a string.
func (b versionBlock) text() string {
	text, _, _ := readUTF16Z(b.value)
	return text
}

// readVersionBlock reads the block at start. Blocks are 4-byte aligned, so next is the start of the sibling.
func readVersionBlock(data []byte, start int) (versionBlock, bool) {
	if start+versionBlockHeaderSize > len(data) {
		return versionBlock{}, false
	}
	length := int(binary.LittleEndian.Uint16(data[start:]))
	valueLen := int(binary.LittleEndian.Uint16(data[start+2:]))
	isText := binary.LittleEndian.Uint16(data[start+4:]) == 1
	if length < versionBlockHeaderSize || start+length > len(data) {
		return versionBlock{}, false
	}
	block := data[start : start+length]

	key, end, ok := readUTF16Z(block[versionBlockHeaderSize:])
	if !ok {
		return versionBlock{}, false
	}
	valueStart := align4(versionBlockHeaderSize + end)
	valueBytes := valueLen
	if isText {
		valueBytes = valueLen * 2
	}
	if valueStart+valueBytes > len(block) {
		return versionBlock{}, false
	}
	childStart := min(align4(valueStart+valueBytes), len(block))
	return versionBlock{
		key:      key,
		isText:   isText,
		value:    block[valueStart : valueStart+valueBytes],
		children: block[childStart:],
		next:     align4(start + length),
	}, true
}

// readUTF16Z decodes a zero-terminated UTF-16 little-endian string. It returns the string and the bytes used, including the terminator.
func readUTF16Z(b []byte) (string, int, bool) {
	units := make([]uint16, 0, 32)
	for i := 0; i+1 < len(b); i += 2 {
		unit := binary.LittleEndian.Uint16(b[i:])
		if unit == 0 {
			return string(utf16.Decode(units)), i + 2, true
		}
		units = append(units, unit)
	}
	return "", 0, false
}

func align4(n int) int {
	return (n + 3) &^ 3
}
