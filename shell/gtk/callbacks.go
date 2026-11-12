//go:build linux

package main

/*
#include <stdint.h>
*/
import "C"

import "runtime/cgo"

//export llGoReady
func llGoReady(handle C.uintptr_t) {
	if session := sessionFrom(handle); session != nil {
		session.onReady()
	}
}

//export llGoDownload
func llGoDownload(handle C.uintptr_t, rawURL, filename *C.char, size C.longlong,
	cookie, referer, userAgent *C.char) {
	if session := sessionFrom(handle); session != nil {
		session.onDownload(goString(rawURL), goString(filename), int64(size),
			goString(cookie), goString(referer), goString(userAgent))
	}
}

//export llGoClosed
func llGoClosed(handle C.uintptr_t) {
	if session := sessionFrom(handle); session != nil {
		session.onClosed()
	}
}

//export llGoGeometry
func llGoGeometry(handle C.uintptr_t, x, y, width, height C.int, positioned C.int) {
	if session := sessionFrom(handle); session != nil {
		session.onGeometry(int(x), int(y), int(width), int(height), positioned != 0)
	}
}

//export llGoBridgeMessage
func llGoBridgeMessage(handle C.uintptr_t, body, pageURI *C.char) {
	if session := sessionFrom(handle); session != nil {
		session.onBridgeMessage(goString(body), goString(pageURI))
	}
}

//export llGoPickerResult
func llGoPickerResult(handle C.uintptr_t, id C.int, path *C.char) {
	if session := sessionFrom(handle); session != nil {
		session.onPickerResult(int(id), goString(path))
	}
}

//export llGoLinkClicked
func llGoLinkClicked(handle C.uintptr_t, rawURL *C.char) C.int {
	session := sessionFrom(handle)
	if session != nil && session.linkClicked(goString(rawURL)) {
		return 1
	}
	return 0
}

//export llGoWebURL
func llGoWebURL(handle C.uintptr_t, rawURL *C.char) C.int {
	session := sessionFrom(handle)
	if session != nil && session.webURL(goString(rawURL)) {
		return 1
	}
	return 0
}

//export llGoCaptureDownload
func llGoCaptureDownload(handle C.uintptr_t, rawURL *C.char) C.int {
	session := sessionFrom(handle)
	if session != nil && session.captureDownload(goString(rawURL)) {
		return 1
	}
	return 0
}

//export llGoWarn
func llGoWarn(handle C.uintptr_t, message *C.char) {
	log.Warn("gtk", "%s", goString(message))
}

func sessionFrom(handle C.uintptr_t) *shellSession {
	session, _ := cgo.Handle(uintptr(handle)).Value().(*shellSession)
	return session
}

func goString(value *C.char) string {
	if value == nil {
		return ""
	}
	return C.GoString(value)
}
