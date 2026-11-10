package systemsvc

import (
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows/registry"
)

// The pinned x/sys version has no bindings for these two calls, so they are loaded directly.
var (
	kernel32                 = syscall.NewLazyDLL("kernel32.dll")
	procGetSystemTimes       = kernel32.NewProc("GetSystemTimes")
	procGlobalMemoryStatusEx = kernel32.NewProc("GlobalMemoryStatusEx")
)

// memoryStatusEx mirrors MEMORYSTATUSEX from the Windows SDK.
type memoryStatusEx struct {
	length               uint32
	memoryLoad           uint32
	totalPhys            uint64
	availPhys            uint64
	totalPageFile        uint64
	availPageFile        uint64
	totalVirtual         uint64
	availVirtual         uint64
	availExtendedVirtual uint64
}

// systemTimes returns idle, kernel and user time in 100 ns units. Kernel time already includes idle time.
func systemTimes() (idle, kernel, user uint64, ok bool) {
	var idleTime, kernelTime, userTime syscall.Filetime
	result, _, _ := procGetSystemTimes.Call(
		uintptr(unsafe.Pointer(&idleTime)),
		uintptr(unsafe.Pointer(&kernelTime)),
		uintptr(unsafe.Pointer(&userTime)),
	)
	if result == 0 {
		return 0, 0, 0, false
	}
	return filetimeValue(idleTime), filetimeValue(kernelTime), filetimeValue(userTime), true
}

func filetimeValue(ft syscall.Filetime) uint64 {
	return uint64(ft.HighDateTime)<<32 | uint64(ft.LowDateTime)
}

// globalMemory returns total and available physical memory in bytes.
func globalMemory() (total, avail uint64, ok bool) {
	var status memoryStatusEx
	status.length = uint32(unsafe.Sizeof(status))
	result, _, _ := procGlobalMemoryStatusEx.Call(uintptr(unsafe.Pointer(&status)))
	if result == 0 {
		return 0, 0, false
	}
	return status.totalPhys, status.availPhys, true
}

// readRegistryString returns a string value under HKEY_LOCAL_MACHINE, or "" when it is missing.
func readRegistryString(path, name string) string {
	key, err := registry.OpenKey(registry.LOCAL_MACHINE, path, registry.QUERY_VALUE)
	if err != nil {
		return ""
	}
	defer key.Close()
	value, _, err := key.GetStringValue(name)
	if err != nil {
		return ""
	}
	return value
}
