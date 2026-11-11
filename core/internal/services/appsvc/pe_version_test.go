package appsvc

import (
	"bytes"
	"debug/pe"
	"encoding/binary"
	"testing"
	"unicode/utf16"
)

const (
	testSectionRVA        = 0x1000
	testSectionFileOffset = 0x200
	testFileAlignment     = 0x200
	testResourceVersionAt = 88
)

// buildTestPE returns a 32-bit PE file whose only resource is a VERSIONINFO block with the given strings.
// An empty string leaves that field out.
func buildTestPE(product, description string) []byte {
	resource := testResourceTree(testVersionInfo(product, description))

	var headers bytes.Buffer
	headers.WriteString("PE\x00\x00")
	mustBinaryWrite(&headers, pe.FileHeader{
		Machine:              0x14c,
		NumberOfSections:     1,
		SizeOfOptionalHeader: 224,
		Characteristics:      0x102,
	})
	mustBinaryWrite(&headers, pe.OptionalHeader32{
		Magic:               0x10b,
		SectionAlignment:    0x1000,
		FileAlignment:       testFileAlignment,
		SizeOfImage:         0x2000,
		SizeOfHeaders:       testSectionFileOffset,
		ImageBase:           0x400000,
		Subsystem:           3,
		NumberOfRvaAndSizes: 16,
		DataDirectory: [16]pe.DataDirectory{
			imageDirectoryEntryResource: {VirtualAddress: testSectionRVA, Size: uint32(len(resource))},
		},
	})
	mustBinaryWrite(&headers, pe.SectionHeader32{
		Name:             [8]byte{'.', 'r', 's', 'r', 'c'},
		VirtualSize:      uint32(len(resource)),
		VirtualAddress:   testSectionRVA,
		SizeOfRawData:    testFileAlignment,
		PointerToRawData: testSectionFileOffset,
		Characteristics:  0x40000040,
	})

	file := make([]byte, testSectionFileOffset)
	file[0], file[1] = 'M', 'Z'
	binary.LittleEndian.PutUint32(file[0x3c:], 0x40)
	copy(file[0x40:], headers.Bytes())
	return append(file, padTo(resource, testFileAlignment)...)
}

// testVersionInfo builds the VS_VERSION_INFO block with one string table.
func testVersionInfo(product, description string) []byte {
	var strs [][]byte
	if product != "" {
		strs = append(strs, versionTestBlock("ProductName", true, utf16Z(product)))
	}
	if description != "" {
		strs = append(strs, versionTestBlock("FileDescription", true, utf16Z(description)))
	}
	table := versionTestBlock("040904b0", false, nil, strs...)
	fileInfo := versionTestBlock("StringFileInfo", false, nil, table)
	fixed := make([]byte, 52)
	binary.LittleEndian.PutUint32(fixed, 0xfeef04bd)
	return versionTestBlock("VS_VERSION_INFO", false, fixed, fileInfo)
}

// versionTestBlock encodes one VERSIONINFO block. Its value and children stay 4-byte aligned.
func versionTestBlock(key string, text bool, value []byte, children ...[]byte) []byte {
	block := make([]byte, versionBlockHeaderSize)
	block = append(block, utf16Z(key)...)
	block = padTo(block, 4)
	block = append(block, value...)
	block = padTo(block, 4)
	for _, child := range children {
		block = append(block, child...)
	}
	valueLen := len(value)
	wType := uint16(0)
	if text {
		valueLen = len(value) / 2
		wType = 1
	}
	binary.LittleEndian.PutUint16(block, uint16(len(block)))
	binary.LittleEndian.PutUint16(block[2:], uint16(valueLen))
	binary.LittleEndian.PutUint16(block[4:], wType)
	return block
}

// testResourceTree lays out a resource directory with type, name and language levels around the version data.
func testResourceTree(version []byte) []byte {
	const (
		typeDirAt   = 0
		nameDirAt   = 24
		langDirAt   = 48
		dataEntryAt = 72
	)
	res := make([]byte, testResourceVersionAt)
	putSingleEntryDir := func(at int, id, offset uint32) {
		binary.LittleEndian.PutUint16(res[at+14:], 1)
		binary.LittleEndian.PutUint32(res[at+16:], id)
		binary.LittleEndian.PutUint32(res[at+20:], offset)
	}
	putSingleEntryDir(typeDirAt, rtVersion, resourceIsDirectory|nameDirAt)
	putSingleEntryDir(nameDirAt, 1, resourceIsDirectory|langDirAt)
	putSingleEntryDir(langDirAt, 0x409, dataEntryAt)
	binary.LittleEndian.PutUint32(res[dataEntryAt:], testSectionRVA+testResourceVersionAt)
	binary.LittleEndian.PutUint32(res[dataEntryAt+4:], uint32(len(version)))
	return append(res, version...)
}

func utf16Z(s string) []byte {
	units := utf16.Encode([]rune(s + "\x00"))
	out := make([]byte, 0, len(units)*2)
	for _, unit := range units {
		out = binary.LittleEndian.AppendUint16(out, unit)
	}
	return out
}

func padTo(b []byte, multiple int) []byte {
	for len(b)%multiple != 0 {
		b = append(b, 0)
	}
	return b
}

func mustBinaryWrite(buf *bytes.Buffer, data any) {
	if err := binary.Write(buf, binary.LittleEndian, data); err != nil {
		panic(err)
	}
}

func TestReadVersionStringsReadsProductName(t *testing.T) {
	got := readVersionStrings(bytes.NewReader(buildTestPE("Synthetic Game", "Synthetic Description")))
	want := versionStrings{productName: "Synthetic Game", fileDescription: "Synthetic Description"}
	if got != want {
		t.Fatalf("readVersionStrings = %+v, want %+v", got, want)
	}
}

func TestReadVersionStringsWithoutProductName(t *testing.T) {
	got := readVersionStrings(bytes.NewReader(buildTestPE("", "Only Description")))
	if got.productName != "" || got.fileDescription != "Only Description" {
		t.Fatalf("readVersionStrings = %+v", got)
	}
}

func TestReadVersionStringsReturnsEmptyForNonPE(t *testing.T) {
	inputs := map[string][]byte{
		"empty":     {},
		"text":      []byte("this is not an executable"),
		"mz only":   []byte("MZ"),
		"no header": make([]byte, 4096),
	}
	for name, data := range inputs {
		if got := readVersionStrings(bytes.NewReader(data)); got != (versionStrings{}) {
			t.Errorf("%s: readVersionStrings = %+v, want empty", name, got)
		}
	}
}

func TestReadVersionStringsSurvivesCorruptFiles(t *testing.T) {
	valid := buildTestPE("Synthetic Game", "")
	versionAt := testSectionFileOffset + testResourceVersionAt

	corrupt := map[string]func([]byte) []byte{
		"huge data size": func(b []byte) []byte {
			binary.LittleEndian.PutUint32(b[testSectionFileOffset+72+4:], 0xffffffff)
			return b
		},
		"data outside file": func(b []byte) []byte {
			binary.LittleEndian.PutUint32(b[testSectionFileOffset+72:], 0x7fff0000)
			return b
		},
		"block length too large": func(b []byte) []byte {
			binary.LittleEndian.PutUint16(b[versionAt:], 0xffff)
			return b
		},
		"block length too small": func(b []byte) []byte {
			binary.LittleEndian.PutUint16(b[versionAt:], 2)
			return b
		},
	}
	for name, mutate := range corrupt {
		data := mutate(append([]byte(nil), valid...))
		if got := readVersionStrings(bytes.NewReader(data)); got != (versionStrings{}) {
			t.Errorf("%s: readVersionStrings = %+v, want empty", name, got)
		}
	}

	for n := 0; n < len(valid); n += 3 {
		readVersionStrings(bytes.NewReader(valid[:n]))
	}
	for i := range valid[:testSectionFileOffset+testResourceVersionAt+64] {
		flipped := append([]byte(nil), valid...)
		flipped[i] ^= 0xff
		readVersionStrings(bytes.NewReader(flipped))
	}
}
