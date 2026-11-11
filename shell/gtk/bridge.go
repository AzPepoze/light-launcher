//go:build linux

package main

import _ "embed"

// bridgeScript defines window.llShell. It is injected into every page the shell loads.
//
//go:embed bridge.js
var bridgeScript string
