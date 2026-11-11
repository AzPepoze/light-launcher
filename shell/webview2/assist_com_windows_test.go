//go:build windows

package main

import (
	"testing"
	"unsafe"

	"github.com/jchv/go-webview2/pkg/edge"
)

// TestCOMVtableOrder checks the slots that the wrapper calls. A wrong slot
// calls an unrelated COM method on the real runtime.
func TestCOMVtableOrder(t *testing.T) {
	size := unsafe.Sizeof(edge.ComProc(0))
	tests := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"webview source", unsafe.Offsetof((webviewVtable{}).getSource), 4},
		{"webview navigation starting", unsafe.Offsetof((webviewVtable{}).addNavigationStarting), 7},
		{"webview new window", unsafe.Offsetof((webviewVtable{}).addNewWindowRequested), 44},
		{"webview cookie manager", unsafe.Offsetof((webviewVtable{}).getCookieManager), 66},
		{"webview download starting", unsafe.Offsetof((webviewVtable{}).addDownloadStarting), 75},
		{"settings user agent", unsafe.Offsetof((webviewSettings2Vtable{}).getUserAgent), 21},
		{"cookie manager get cookies", unsafe.Offsetof((cookieManagerVtable{}).getCookies), 5},
		{"cookie list count", unsafe.Offsetof((cookieListVtable{}).getCount), 3},
		{"cookie name", unsafe.Offsetof((cookieObjectVtable{}).getName), 3},
		{"download uri", unsafe.Offsetof((downloadOperationVtable{}).getURI), 9},
		{"download total bytes", unsafe.Offsetof((downloadOperationVtable{}).getTotalBytesToReceive), 12},
		{"download args operation", unsafe.Offsetof((downloadStartingArgsVtable{}).getDownloadOperation), 3},
		{"download args cancel", unsafe.Offsetof((downloadStartingArgsVtable{}).putCancel), 5},
		{"download args file path", unsafe.Offsetof((downloadStartingArgsVtable{}).getResultFilePath), 6},
		{"download args handled", unsafe.Offsetof((downloadStartingArgsVtable{}).putHandled), 9},
		{"navigation uri", unsafe.Offsetof((navigationStartingArgsVtable{}).getURI), 3},
		{"navigation cancel", unsafe.Offsetof((navigationStartingArgsVtable{}).putCancel), 8},
		{"new window uri", unsafe.Offsetof((newWindowArgsVtable{}).getURI), 3},
		{"new window handled", unsafe.Offsetof((newWindowArgsVtable{}).putHandled), 6},
		{"get cookies invoke", unsafe.Offsetof((getCookiesVtable{}).invoke), 3},
		{"event invoke", unsafe.Offsetof((eventHandlerVtable{}).invoke), 3},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if index := test.got / size; index != test.want {
				t.Fatalf("vtable slot = %d, want %d", index, test.want)
			}
		})
	}
}
