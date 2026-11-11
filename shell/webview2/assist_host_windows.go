//go:build windows

package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"unsafe"

	"github.com/jchv/go-webview2/pkg/edge"
	"golang.org/x/sys/windows"

	"light-launcher/shell/common/assist"
	"light-launcher/shell/common/urlx"
)

const (
	assistWidth       = 1000
	assistHeight      = 720
	assistWindowClass = "LightLauncherAssistWindow"

	wsOverlappedWindow = 0x00CF0000
	swShow             = 5
	cwUseDefault       = 0x80000000

	wmDestroy = 0x0002
	wmSize    = 0x0005
	wmClose   = 0x0010

	colorWindow = 5
	idcArrow    = 32512
)

var (
	user32 = windows.NewLazySystemDLL("user32.dll")

	procRegisterClassExW = user32.NewProc("RegisterClassExW")
	procCreateWindowExW  = user32.NewProc("CreateWindowExW")
	procDefWindowProcW   = user32.NewProc("DefWindowProcW")
	procDestroyWindow    = user32.NewProc("DestroyWindow")
	procShowWindow       = user32.NewProc("ShowWindow")
	procUpdateWindow     = user32.NewProc("UpdateWindow")
	procGetMessageW      = user32.NewProc("GetMessageW")
	procTranslateMessage = user32.NewProc("TranslateMessage")
	procDispatchMessageW = user32.NewProc("DispatchMessageW")
	procPostQuitMessage  = user32.NewProc("PostQuitMessage")
	procPostMessageW     = user32.NewProc("PostMessageW")
	procLoadCursorW      = user32.NewProc("LoadCursorW")
)

type assistHost struct {
	hwnd     uintptr
	chromium *edge.Chromium
	webview  *webview
	emitter  *assist.Writer

	downloadSeen bool
	closing      bool
	finished     bool
	failed       bool

	// WebView2 keeps these COM objects; the fields keep them reachable.
	navigationHandler *eventHandler
	newWindowHandler  *eventHandler
	downloadHandler   *eventHandler
	cookiesHandler    *getCookiesHandler
}

// runAssist runs one human-driven assist session. The page is shown in a normal
// WebView2 window; the shell only observes and cancels the download the page
// starts, so the launcher can download the file itself.
func runAssist(args []string) int {
	opts, err := parseAssistOptions(args)
	if err != nil {
		log.Error("args", "%v", err)
		fmt.Fprintln(os.Stderr, assist.Usage)
		return 2
	}

	emitter := assist.NewWriter(os.Stdout)
	if !webView2Installed() {
		err := errors.New("WebView2 runtime is not installed")
		log.Error("window", "%v", err)
		emitter.Write(assist.Error(err.Error()))
		return 1
	}

	host, err := newAssistHost(opts, emitter)
	if err != nil {
		log.Error("window", "%v", err)
		emitter.Write(assist.Error(err.Error()))
		return 1
	}
	defer host.destroy()

	if err := emitter.Write(assist.Ready()); err != nil {
		log.Error("assist", "cannot write ready event: %v", err)
		return 1
	}
	host.run()

	if err := emitter.Err(); err != nil {
		log.Error("assist", "cannot write output: %v", err)
		return 1
	}
	if host.failed {
		return 1
	}
	if !host.downloadSeen {
		if err := emitter.Write(assist.Closed()); err != nil {
			log.Error("assist", "cannot write closed event: %v", err)
			return 1
		}
	}
	return 0
}

func newAssistHost(opts assistOptions, emitter *assist.Writer) (*assistHost, error) {
	profile, err := filepath.Abs(opts.profile)
	if err != nil {
		return nil, fmt.Errorf("profile path: %w", err)
	}
	if err := os.MkdirAll(profile, 0o700); err != nil {
		log.Error("assist", "cannot create profile folder %q: %v", profile, err)
		return nil, errors.New("cannot create the browser profile folder")
	}

	host := &assistHost{emitter: emitter}
	hwnd, err := host.createWindow(opts.title, assistWidth, assistHeight)
	if err != nil {
		return nil, err
	}
	host.hwnd = hwnd

	chromium := edge.NewChromium()
	chromium.DataPath = profile
	host.chromium = chromium
	if !chromium.Embed(hwnd) {
		host.destroy()
		return nil, errors.New("WebView2 runtime could not start")
	}

	base := chromium.GetICoreWebView2_3()
	if base == nil {
		host.destroy()
		return nil, errors.New("WebView2 runtime is too old for download capture")
	}
	downloads, err := (*webview)(unsafe.Pointer(base)).queryInterface(iidWebView4)
	if err != nil {
		host.destroy()
		return nil, fmt.Errorf("WebView2 download interface: %w", err)
	}
	host.webview = (*webview)(downloads)

	if err := host.bindEvents(); err != nil {
		host.destroy()
		return nil, err
	}
	if script := bannerScript(opts.message); script != "" {
		chromium.Init(script)
	}
	chromium.Navigate(opts.url.String())
	showWindow(hwnd)
	chromium.Resize()
	return host, nil
}

func (h *assistHost) bindEvents() error {
	navigation := newEventHandler(func(sender, args unsafe.Pointer) uintptr {
		h.onNavigationStarting((*navigationStartingArgs)(args))
		return 0
	})
	if err := h.webview.addNavigationStarting(navigation); err != nil {
		return fmt.Errorf("watch navigation: %w", err)
	}
	h.navigationHandler = navigation

	newWindow := newEventHandler(func(sender, args unsafe.Pointer) uintptr {
		h.onNewWindow((*newWindowArgs)(args))
		return 0
	})
	if err := h.webview.addNewWindowRequested(newWindow); err != nil {
		return fmt.Errorf("watch new windows: %w", err)
	}
	h.newWindowHandler = newWindow

	download := newEventHandler(func(sender, args unsafe.Pointer) uintptr {
		h.onDownload((*downloadStartingArgs)(args))
		return 0
	})
	if err := h.webview.addDownloadStarting(download); err != nil {
		return fmt.Errorf("watch downloads: %w", err)
	}
	h.downloadHandler = download
	return nil
}

func (h *assistHost) onNavigationStarting(args *navigationStartingArgs) {
	uri, err := args.uri()
	if err != nil {
		log.Warn("assist", "cannot read navigation address: %v", err)
		args.putCancel(true)
		return
	}
	if urlx.IsWebAddress(uri) {
		return
	}
	log.Info("assist", "blocked a navigation that is not http or https")
	if err := args.putCancel(true); err != nil {
		log.Error("assist", "cannot cancel navigation: %v", err)
	}
}

func (h *assistHost) onNewWindow(args *newWindowArgs) {
	if err := args.putHandled(true); err != nil {
		log.Error("assist", "cannot keep a popup in this window: %v", err)
	}
	uri, err := args.uri()
	if err != nil {
		log.Warn("assist", "cannot read popup address: %v", err)
		return
	}
	if !urlx.IsWebAddress(uri) {
		log.Info("assist", "blocked a popup that is not http or https")
		return
	}
	h.chromium.Navigate(uri)
}

func (h *assistHost) onDownload(args *downloadStartingArgs) {
	// The download cannot progress until this handler returns, so the metadata
	// is read first and the cancel runs on every path before returning.
	defer h.cancelDownload(args)

	operation, err := args.downloadOperation()
	if err != nil {
		h.fail(fmt.Sprintf("cannot read download: %v", err))
		return
	}
	defer operation.release()

	uri, err := operation.uri()
	if err != nil {
		h.fail(fmt.Sprintf("cannot read download address: %v", err))
		return
	}
	if !urlx.IsWebAddress(uri) {
		log.Info("assist", "ignored a download that is not http or https")
		return
	}
	if h.downloadSeen {
		return
	}
	h.downloadSeen = true

	filename := ""
	if path, err := args.resultFilePath(); err != nil {
		log.Warn("assist", "cannot read download file name: %v", err)
	} else {
		filename = assist.ResultFileName(path)
	}

	size := int64(-1)
	if total, err := operation.totalBytes(); err != nil {
		log.Warn("assist", "cannot read download size: %v", err)
	} else {
		size = total
	}

	referer, err := h.webview.source()
	if err != nil {
		log.Warn("assist", "cannot read page source: %v", err)
	}

	manager, err := h.webview.cookieManager()
	if err != nil {
		log.Warn("assist", "cannot read cookies: %v", err)
		h.finishDownload(uri, filename, size, referer, nil)
		return
	}
	handler := newGetCookiesHandler(func(result uintptr, list *cookieList) {
		var cookies []assist.CookiePair
		switch {
		case failed(result):
			log.Warn("assist", "cookie lookup failed: %v", hresultError(result))
		case list == nil:
			log.Warn("assist", "cookie lookup returned no list")
		default:
			pairs, err := list.cookies()
			if err != nil {
				log.Warn("assist", "cannot read cookies: %v", err)
			} else {
				cookies = pairs
			}
		}
		h.finishDownload(uri, filename, size, referer, cookies)
	})
	h.cookiesHandler = handler
	err = manager.getCookies(uri, handler)
	manager.release()
	if err != nil {
		log.Warn("assist", "cannot request cookies: %v", err)
		h.finishDownload(uri, filename, size, referer, nil)
	}
}

func (h *assistHost) cancelDownload(args *downloadStartingArgs) {
	if err := args.putCancel(true); err != nil {
		log.Error("assist", "cannot cancel download: %v", err)
	}
	if err := args.putHandled(true); err != nil {
		log.Error("assist", "cannot hide download dialog: %v", err)
	}
}

func (h *assistHost) finishDownload(uri, filename string, size int64, referer string, cookies []assist.CookiePair) {
	if h.finished || h.failed {
		return
	}
	h.finished = true
	userAgent, err := h.webview.userAgent()
	if err != nil {
		log.Warn("assist", "cannot read user agent: %v", err)
	}
	headers := assist.DownloadHeaders(assist.CookieHeader(cookies), referer, userAgent)
	if err := h.emitter.Write(assist.DownloadWithHeaders(uri, filename, size, headers)); err != nil {
		log.Error("assist", "cannot write download event: %v", err)
	}
	h.closeWindow()
}

// fail reports a fatal problem after the window exists and stops the session.
func (h *assistHost) fail(message string) {
	if h.finished || h.failed {
		return
	}
	h.failed = true
	log.Error("assist", "%s", message)
	h.emitter.Write(assist.Error(message))
	h.closeWindow()
}

func (h *assistHost) closeWindow() {
	h.closing = true
	if h.hwnd != 0 {
		procPostMessageW.Call(h.hwnd, wmClose, 0, 0)
	}
}

func (h *assistHost) run() {
	var message windowMessage
	for {
		result, _, _ := procGetMessageW.Call(uintptr(unsafe.Pointer(&message)), 0, 0, 0)
		if int32(result) <= 0 {
			return
		}
		procTranslateMessage.Call(uintptr(unsafe.Pointer(&message)))
		procDispatchMessageW.Call(uintptr(unsafe.Pointer(&message)))
	}
}

func (h *assistHost) resize() {
	if h.chromium != nil {
		h.chromium.Resize()
	}
}

func (h *assistHost) destroy() {
	if h.hwnd != 0 {
		procDestroyWindow.Call(h.hwnd)
		h.hwnd = 0
	}
}

type wndClassEx struct {
	cbSize        uint32
	style         uint32
	lpfnWndProc   uintptr
	cbClsExtra    int32
	cbWndExtra    int32
	hInstance     windows.Handle
	hIcon         windows.Handle
	hCursor       windows.Handle
	hbrBackground windows.Handle
	lpszMenuName  *uint16
	lpszClassName *uint16
	hIconSm       windows.Handle
}

type windowPoint struct {
	x int32
	y int32
}

type windowMessage struct {
	hwnd     uintptr
	message  uint32
	wParam   uintptr
	lParam   uintptr
	time     uint32
	point    windowPoint
	lPrivate uint32
}

func (h *assistHost) createWindow(title string, width, height int) (uintptr, error) {
	var instance windows.Handle
	if err := windows.GetModuleHandleEx(0, nil, &instance); err != nil {
		return 0, fmt.Errorf("module handle: %w", err)
	}
	className, err := windows.UTF16PtrFromString(assistWindowClass)
	if err != nil {
		return 0, err
	}
	windowTitle, err := windows.UTF16PtrFromString(title)
	if err != nil {
		return 0, err
	}
	cursor, _, _ := procLoadCursorW.Call(0, idcArrow)

	class := wndClassEx{
		cbSize:        uint32(unsafe.Sizeof(wndClassEx{})),
		style:         3,
		lpfnWndProc:   windows.NewCallback(h.windowProc),
		hInstance:     instance,
		hCursor:       windows.Handle(cursor),
		hbrBackground: colorWindow + 1,
		lpszClassName: className,
	}
	atom, _, callErr := procRegisterClassExW.Call(uintptr(unsafe.Pointer(&class)))
	if atom == 0 {
		return 0, fmt.Errorf("register window class: %w", callErr)
	}

	hwnd, _, callErr := procCreateWindowExW.Call(
		0,
		uintptr(unsafe.Pointer(className)),
		uintptr(unsafe.Pointer(windowTitle)),
		wsOverlappedWindow,
		cwUseDefault,
		cwUseDefault,
		uintptr(width),
		uintptr(height),
		0,
		0,
		uintptr(instance),
		0,
	)
	if hwnd == 0 {
		return 0, fmt.Errorf("create window: %w", callErr)
	}
	return hwnd, nil
}

func (h *assistHost) windowProc(hwnd, message, wParam, lParam uintptr) uintptr {
	switch message {
	case wmSize:
		h.resize()
		return 0
	case wmClose:
		if h.downloadSeen && !h.closing {
			return 0
		}
		procDestroyWindow.Call(hwnd)
		return 0
	case wmDestroy:
		h.hwnd = 0
		procPostQuitMessage.Call(0)
		return 0
	}
	result, _, _ := procDefWindowProcW.Call(hwnd, message, wParam, lParam)
	return result
}

func showWindow(hwnd uintptr) {
	procShowWindow.Call(hwnd, swShow)
	procUpdateWindow.Call(hwnd)
}
