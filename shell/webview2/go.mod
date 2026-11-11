module light-launcher/shell/webview2

go 1.25.6

replace light-launcher/shell/common => ../common

require (
	github.com/jchv/go-webview2 v0.0.0-20260205173254-56598839c808
	golang.org/x/sys v0.40.0
	light-launcher/shell/common v0.0.0-00010101000000-000000000000
)

require github.com/jchv/go-winloader v0.0.0-20250406163304-c1995be93bd1 // indirect
