//go:build windows

package plugins

import (
	"context"
	"errors"
	"fmt"
	"unsafe"

	"golang.org/x/sys/windows"

	"light-launcher/core/internal/contract"
)

var (
	advapi32        = windows.NewLazySystemDLL("advapi32.dll")
	procCredWriteW  = advapi32.NewProc("CredWriteW")
	procCredReadW   = advapi32.NewProc("CredReadW")
	procCredDeleteW = advapi32.NewProc("CredDeleteW")
	procCredFree    = advapi32.NewProc("CredFree")
)

const (
	credTypeGeneric         = 1
	credPersistLocalMachine = 2
	credMaxBlobBytes        = 2560
)

// credential mirrors the CREDENTIALW struct from wincred.h.
type credential struct {
	Flags              uint32
	Type               uint32
	TargetName         *uint16
	Comment            *uint16
	LastWritten        windows.Filetime
	CredentialBlobSize uint32
	CredentialBlob     *byte
	Persist            uint32
	AttributeCount     uint32
	Attributes         uintptr
	TargetAlias        *uint16
	UserName           *uint16
}

// newPlatformSecrets uses the Windows Credential Manager.
func newPlatformSecrets(_ string) contract.Secrets {
	return windowsCredentials{}
}

// windowsCredentials stores secrets as generic credentials in the Credential Manager.
type windowsCredentials struct{}

func credentialTarget(plugin, key string) string {
	return "light-launcher/" + plugin + "/" + key
}

func (windowsCredentials) Get(_ context.Context, plugin, key string) (string, bool, error) {
	target, err := windows.UTF16PtrFromString(credentialTarget(plugin, key))
	if err != nil {
		return "", false, err
	}
	var found *credential
	result, _, callErr := procCredReadW.Call(
		uintptr(unsafe.Pointer(target)),
		credTypeGeneric,
		0,
		uintptr(unsafe.Pointer(&found)),
	)
	if result == 0 {
		if errors.Is(callErr, windows.ERROR_NOT_FOUND) {
			return "", false, nil
		}
		return "", false, fmt.Errorf("credential read: %w", callErr)
	}
	defer procCredFree.Call(uintptr(unsafe.Pointer(found)))
	blob := unsafe.Slice(found.CredentialBlob, found.CredentialBlobSize)
	return string(blob), true, nil
}

func (windowsCredentials) Set(_ context.Context, plugin, key, value string) error {
	if len(value) > credMaxBlobBytes {
		return fmt.Errorf("secret is larger than %d bytes", credMaxBlobBytes)
	}
	target, err := windows.UTF16PtrFromString(credentialTarget(plugin, key))
	if err != nil {
		return err
	}
	blob := []byte(value)
	entry := &credential{
		Type:               credTypeGeneric,
		TargetName:         target,
		CredentialBlobSize: uint32(len(blob)),
		Persist:            credPersistLocalMachine,
	}
	if len(blob) > 0 {
		entry.CredentialBlob = &blob[0]
	}
	result, _, callErr := procCredWriteW.Call(uintptr(unsafe.Pointer(entry)), 0)
	if result == 0 {
		return fmt.Errorf("credential write: %w", callErr)
	}
	return nil
}

func (windowsCredentials) Delete(_ context.Context, plugin, key string) error {
	target, err := windows.UTF16PtrFromString(credentialTarget(plugin, key))
	if err != nil {
		return err
	}
	result, _, callErr := procCredDeleteW.Call(uintptr(unsafe.Pointer(target)), credTypeGeneric, 0)
	if result == 0 && !errors.Is(callErr, windows.ERROR_NOT_FOUND) {
		return fmt.Errorf("credential delete: %w", callErr)
	}
	return nil
}
