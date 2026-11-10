//go:build !linux

package runtimesvc

import "errors"

var errFileTimesUnavailable = errors.New("file birth time is not read on this platform")

// fileTimes is not implemented outside Linux, so prefix creation dates are reported as unknown.
func fileTimes(target string) (birthMs, changeMs, modifyMs float64, err error) {
	return 0, 0, 0, errFileTimesUnavailable
}
