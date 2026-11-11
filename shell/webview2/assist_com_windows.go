//go:build windows

package main

import (
	"errors"
	"fmt"
	"unsafe"

	"github.com/jchv/go-webview2/pkg/edge"
	"golang.org/x/sys/windows"
)

// The go-webview2 package exposes the window and navigation calls but not the
// DownloadStarting event, the cookie manager or the user agent setting. These
// small wrappers add only those COM calls; the vtable layouts follow the
// WebView2 IDL.

var (
	iidWebView4  = edge.NewGUID("{20d02d59-6df2-42dc-bd06-f98a694b1302}")
	iidSettings2 = edge.NewGUID("{ee9a0f68-f46c-4e32-ac23-ef8cac224d2a}")
)

const eNoInterface = 0x80004002

type iunknownVtable struct {
	queryInterface edge.ComProc
	addRef         edge.ComProc
	release        edge.ComProc
}

type iunknownObject struct {
	vtbl *iunknownVtable
}

func (o *iunknownObject) queryInterface(iid *edge.GUID) (unsafe.Pointer, error) {
	if iid == nil {
		return nil, errors.New("missing interface id")
	}
	var out unsafe.Pointer
	hr, _, _ := o.vtbl.queryInterface.Call(
		uintptr(unsafe.Pointer(o)),
		uintptr(unsafe.Pointer(iid)),
		uintptr(unsafe.Pointer(&out)),
	)
	if failed(hr) {
		return nil, hresultError(hr)
	}
	if out == nil {
		return nil, errors.New("COM call returned no interface")
	}
	return out, nil
}

func (o *iunknownObject) release() {
	o.vtbl.release.Call(uintptr(unsafe.Pointer(o)))
}

type eventToken struct {
	value int64
}

// eventHandler is the COM object that receives a WebView2 event.
type eventHandler struct {
	vtbl   *eventHandlerVtable
	invoke func(sender, args unsafe.Pointer) uintptr
}

type eventHandlerVtable struct {
	iunknownVtable
	invoke edge.ComProc
}

func newEventHandler(invoke func(sender, args unsafe.Pointer) uintptr) *eventHandler {
	handler := &eventHandler{invoke: invoke}
	handler.vtbl = &eventHandlerVtable{
		iunknownVtable: iunknownVtable{
			queryInterface: edge.NewComProc(eventQueryInterface),
			addRef:         edge.NewComProc(eventAddRef),
			release:        edge.NewComProc(eventRelease),
		},
		invoke: edge.NewComProc(eventInvoke),
	}
	return handler
}

func eventQueryInterface(this, iid, out unsafe.Pointer) uintptr {
	return eNoInterface
}

func eventAddRef(this unsafe.Pointer) uintptr {
	return 1
}

func eventRelease(this unsafe.Pointer) uintptr {
	return 1
}

func eventInvoke(this, sender, args unsafe.Pointer) uintptr {
	handler := (*eventHandler)(this)
	return handler.invoke(sender, args)
}

type getCookiesHandler struct {
	vtbl   *getCookiesVtable
	invoke func(result uintptr, list *cookieList)
}

type getCookiesVtable struct {
	iunknownVtable
	invoke edge.ComProc
}

func newGetCookiesHandler(invoke func(result uintptr, list *cookieList)) *getCookiesHandler {
	handler := &getCookiesHandler{invoke: invoke}
	handler.vtbl = &getCookiesVtable{
		iunknownVtable: iunknownVtable{
			queryInterface: edge.NewComProc(eventQueryInterface),
			addRef:         edge.NewComProc(eventAddRef),
			release:        edge.NewComProc(eventRelease),
		},
		invoke: edge.NewComProc(getCookiesInvoke),
	}
	return handler
}

func getCookiesInvoke(this unsafe.Pointer, result uintptr, list unsafe.Pointer) uintptr {
	handler := (*getCookiesHandler)(this)
	handler.invoke(result, (*cookieList)(list))
	return 0
}

// webview is ICoreWebView2_4. The library stops at ICoreWebView2_3, so the
// inherited vtable entries are listed here in IDL order.
type webview struct {
	vtbl *webviewVtable
}

type webviewVtable struct {
	iunknownVtable
	getSettings                            edge.ComProc
	getSource                              edge.ComProc
	navigate                               edge.ComProc
	navigateToString                       edge.ComProc
	addNavigationStarting                  edge.ComProc
	removeNavigationStarting               edge.ComProc
	addContentLoading                      edge.ComProc
	removeContentLoading                   edge.ComProc
	addSourceChanged                       edge.ComProc
	removeSourceChanged                    edge.ComProc
	addHistoryChanged                      edge.ComProc
	removeHistoryChanged                   edge.ComProc
	addNavigationCompleted                 edge.ComProc
	removeNavigationCompleted              edge.ComProc
	addFrameNavigationStarting             edge.ComProc
	removeFrameNavigationStarting          edge.ComProc
	addFrameNavigationCompleted            edge.ComProc
	removeFrameNavigationCompleted         edge.ComProc
	addScriptDialogOpening                 edge.ComProc
	removeScriptDialogOpening              edge.ComProc
	addPermissionRequested                 edge.ComProc
	removePermissionRequested              edge.ComProc
	addProcessFailed                       edge.ComProc
	removeProcessFailed                    edge.ComProc
	addScriptToExecuteOnDocumentCreated    edge.ComProc
	removeScriptToExecuteOnDocumentCreated edge.ComProc
	executeScript                          edge.ComProc
	capturePreview                         edge.ComProc
	reload                                 edge.ComProc
	postWebMessageAsJSON                   edge.ComProc
	postWebMessageAsString                 edge.ComProc
	addWebMessageReceived                  edge.ComProc
	removeWebMessageReceived               edge.ComProc
	callDevToolsProtocolMethod             edge.ComProc
	getBrowserProcessID                    edge.ComProc
	getCanGoBack                           edge.ComProc
	getCanGoForward                        edge.ComProc
	goBack                                 edge.ComProc
	goForward                              edge.ComProc
	getDevToolsProtocolEventReceiver       edge.ComProc
	stop                                   edge.ComProc
	addNewWindowRequested                  edge.ComProc
	removeNewWindowRequested               edge.ComProc
	addDocumentTitleChanged                edge.ComProc
	removeDocumentTitleChanged             edge.ComProc
	getDocumentTitle                       edge.ComProc
	addHostObjectToScript                  edge.ComProc
	removeHostObjectFromScript             edge.ComProc
	openDevToolsWindow                     edge.ComProc
	addContainsFullScreenElementChanged    edge.ComProc
	removeContainsFullScreenElementChanged edge.ComProc
	getContainsFullScreenElement           edge.ComProc
	addWebResourceRequested                edge.ComProc
	removeWebResourceRequested             edge.ComProc
	addWebResourceRequestedFilter          edge.ComProc
	removeWebResourceRequestedFilter       edge.ComProc
	addWindowCloseRequested                edge.ComProc
	removeWindowCloseRequested             edge.ComProc

	// ICoreWebView2_2
	addWebResourceResponseReceived    edge.ComProc
	removeWebResourceResponseReceived edge.ComProc
	navigateWithWebResourceRequest    edge.ComProc
	addDOMContentLoaded               edge.ComProc
	removeDOMContentLoaded            edge.ComProc
	getCookieManager                  edge.ComProc
	getEnvironment                    edge.ComProc

	// ICoreWebView2_3
	trySuspend                          edge.ComProc
	resume                              edge.ComProc
	getIsSuspended                      edge.ComProc
	setVirtualHostNameToFolderMapping   edge.ComProc
	clearVirtualHostNameToFolderMapping edge.ComProc

	// ICoreWebView2_4
	addFrameCreated        edge.ComProc
	removeFrameCreated     edge.ComProc
	addDownloadStarting    edge.ComProc
	removeDownloadStarting edge.ComProc
}

func (w *webview) queryInterface(iid *edge.GUID) (unsafe.Pointer, error) {
	return (*iunknownObject)(unsafe.Pointer(w)).queryInterface(iid)
}

func (w *webview) addNavigationStarting(handler *eventHandler) error {
	var token eventToken
	hr, _, _ := w.vtbl.addNavigationStarting.Call(
		uintptr(unsafe.Pointer(w)),
		uintptr(unsafe.Pointer(handler)),
		uintptr(unsafe.Pointer(&token)),
	)
	if failed(hr) {
		return hresultError(hr)
	}
	return nil
}

func (w *webview) addNewWindowRequested(handler *eventHandler) error {
	var token eventToken
	hr, _, _ := w.vtbl.addNewWindowRequested.Call(
		uintptr(unsafe.Pointer(w)),
		uintptr(unsafe.Pointer(handler)),
		uintptr(unsafe.Pointer(&token)),
	)
	if failed(hr) {
		return hresultError(hr)
	}
	return nil
}

func (w *webview) addDownloadStarting(handler *eventHandler) error {
	var token eventToken
	hr, _, _ := w.vtbl.addDownloadStarting.Call(
		uintptr(unsafe.Pointer(w)),
		uintptr(unsafe.Pointer(handler)),
		uintptr(unsafe.Pointer(&token)),
	)
	if failed(hr) {
		return hresultError(hr)
	}
	return nil
}

func (w *webview) source() (string, error) {
	var raw *uint16
	hr, _, _ := w.vtbl.getSource.Call(
		uintptr(unsafe.Pointer(w)),
		uintptr(unsafe.Pointer(&raw)),
	)
	if failed(hr) {
		return "", hresultError(hr)
	}
	defer windows.CoTaskMemFree(unsafe.Pointer(raw))
	return windows.UTF16PtrToString(raw), nil
}

func (w *webview) cookieManager() (*cookieManager, error) {
	var manager *cookieManager
	hr, _, _ := w.vtbl.getCookieManager.Call(
		uintptr(unsafe.Pointer(w)),
		uintptr(unsafe.Pointer(&manager)),
	)
	if failed(hr) {
		return nil, hresultError(hr)
	}
	if manager == nil {
		return nil, errors.New("cookie manager is not available")
	}
	return manager, nil
}

func (w *webview) userAgent() (string, error) {
	var settings *iunknownObject
	hr, _, _ := w.vtbl.getSettings.Call(
		uintptr(unsafe.Pointer(w)),
		uintptr(unsafe.Pointer(&settings)),
	)
	if failed(hr) {
		return "", hresultError(hr)
	}
	if settings == nil {
		return "", errors.New("settings are not available")
	}
	defer settings.release()

	raw, err := settings.queryInterface(iidSettings2)
	if err != nil {
		return "", err
	}
	settings2 := (*webviewSettings2)(raw)
	defer settings2.release()

	var userAgent *uint16
	hr, _, _ = settings2.vtbl.getUserAgent.Call(
		uintptr(unsafe.Pointer(settings2)),
		uintptr(unsafe.Pointer(&userAgent)),
	)
	if failed(hr) {
		return "", hresultError(hr)
	}
	defer windows.CoTaskMemFree(unsafe.Pointer(userAgent))
	return windows.UTF16PtrToString(userAgent), nil
}

type webviewSettings2 struct {
	vtbl *webviewSettings2Vtable
}

type webviewSettings2Vtable struct {
	iunknownVtable
	getIsScriptEnabled                edge.ComProc
	putIsScriptEnabled                edge.ComProc
	getIsWebMessageEnabled            edge.ComProc
	putIsWebMessageEnabled            edge.ComProc
	getAreDefaultScriptDialogsEnabled edge.ComProc
	putAreDefaultScriptDialogsEnabled edge.ComProc
	getIsStatusBarEnabled             edge.ComProc
	putIsStatusBarEnabled             edge.ComProc
	getAreDevToolsEnabled             edge.ComProc
	putAreDevToolsEnabled             edge.ComProc
	getAreDefaultContextMenusEnabled  edge.ComProc
	putAreDefaultContextMenusEnabled  edge.ComProc
	getAreHostObjectsAllowed          edge.ComProc
	putAreHostObjectsAllowed          edge.ComProc
	getIsZoomControlEnabled           edge.ComProc
	putIsZoomControlEnabled           edge.ComProc
	getIsBuiltInErrorPageEnabled      edge.ComProc
	putIsBuiltInErrorPageEnabled      edge.ComProc
	getUserAgent                      edge.ComProc
	putUserAgent                      edge.ComProc
}

func (s *webviewSettings2) release() {
	s.vtbl.release.Call(uintptr(unsafe.Pointer(s)))
}

type cookieManager struct {
	vtbl *cookieManagerVtable
}

type cookieManagerVtable struct {
	iunknownVtable
	createCookie                   edge.ComProc
	copyCookie                     edge.ComProc
	getCookies                     edge.ComProc
	addOrUpdateCookie              edge.ComProc
	deleteCookie                   edge.ComProc
	deleteCookies                  edge.ComProc
	deleteCookiesWithDomainAndPath edge.ComProc
	deleteAllCookies               edge.ComProc
}

func (m *cookieManager) getCookies(rawURL string, handler *getCookiesHandler) error {
	target, err := windows.UTF16PtrFromString(rawURL)
	if err != nil {
		return err
	}
	hr, _, _ := m.vtbl.getCookies.Call(
		uintptr(unsafe.Pointer(m)),
		uintptr(unsafe.Pointer(target)),
		uintptr(unsafe.Pointer(handler)),
	)
	if failed(hr) {
		return hresultError(hr)
	}
	return nil
}

func (m *cookieManager) release() {
	m.vtbl.release.Call(uintptr(unsafe.Pointer(m)))
}

type cookieList struct {
	vtbl *cookieListVtable
}

type cookieListVtable struct {
	iunknownVtable
	getCount        edge.ComProc
	getValueAtIndex edge.ComProc
}

func (l *cookieList) cookies() ([]cookiePair, error) {
	var count uint32
	hr, _, _ := l.vtbl.getCount.Call(
		uintptr(unsafe.Pointer(l)),
		uintptr(unsafe.Pointer(&count)),
	)
	if failed(hr) {
		return nil, hresultError(hr)
	}
	cookies := make([]cookiePair, 0, count)
	for index := uint32(0); index < count; index++ {
		var cookie *cookieObject
		hr, _, _ := l.vtbl.getValueAtIndex.Call(
			uintptr(unsafe.Pointer(l)),
			uintptr(index),
			uintptr(unsafe.Pointer(&cookie)),
		)
		if failed(hr) {
			return nil, hresultError(hr)
		}
		if cookie == nil {
			return nil, errors.New("cookie list returned no cookie")
		}
		name, nameErr := cookie.name()
		value, valueErr := cookie.value()
		cookie.release()
		if nameErr != nil {
			return nil, nameErr
		}
		if valueErr != nil {
			return nil, valueErr
		}
		cookies = append(cookies, cookiePair{Name: name, Value: value})
	}
	return cookies, nil
}

type cookieObject struct {
	vtbl *cookieObjectVtable
}

type cookieObjectVtable struct {
	iunknownVtable
	getName       edge.ComProc
	getValue      edge.ComProc
	putValue      edge.ComProc
	getDomain     edge.ComProc
	getPath       edge.ComProc
	getExpires    edge.ComProc
	putExpires    edge.ComProc
	getIsHTTPOnly edge.ComProc
	putIsHTTPOnly edge.ComProc
	getSameSite   edge.ComProc
	putSameSite   edge.ComProc
	getIsSecure   edge.ComProc
	putIsSecure   edge.ComProc
	getIsSession  edge.ComProc
}

func (c *cookieObject) name() (string, error) {
	var raw *uint16
	hr, _, _ := c.vtbl.getName.Call(
		uintptr(unsafe.Pointer(c)),
		uintptr(unsafe.Pointer(&raw)),
	)
	if failed(hr) {
		return "", hresultError(hr)
	}
	defer windows.CoTaskMemFree(unsafe.Pointer(raw))
	return windows.UTF16PtrToString(raw), nil
}

func (c *cookieObject) value() (string, error) {
	var raw *uint16
	hr, _, _ := c.vtbl.getValue.Call(
		uintptr(unsafe.Pointer(c)),
		uintptr(unsafe.Pointer(&raw)),
	)
	if failed(hr) {
		return "", hresultError(hr)
	}
	defer windows.CoTaskMemFree(unsafe.Pointer(raw))
	return windows.UTF16PtrToString(raw), nil
}

func (c *cookieObject) release() {
	c.vtbl.release.Call(uintptr(unsafe.Pointer(c)))
}

type downloadOperation struct {
	vtbl *downloadOperationVtable
}

type downloadOperationVtable struct {
	iunknownVtable
	addBytesReceivedChanged       edge.ComProc
	removeBytesReceivedChanged    edge.ComProc
	addEstimatedEndTimeChanged    edge.ComProc
	removeEstimatedEndTimeChanged edge.ComProc
	addStateChanged               edge.ComProc
	removeStateChanged            edge.ComProc
	getURI                        edge.ComProc
	getContentDisposition         edge.ComProc
	getMIMEType                   edge.ComProc
	getTotalBytesToReceive        edge.ComProc
	getBytesReceived              edge.ComProc
	getEstimatedEndTime           edge.ComProc
	getResultFilePath             edge.ComProc
	getState                      edge.ComProc
	getInterruptReason            edge.ComProc
	cancel                        edge.ComProc
	pause                         edge.ComProc
	resume                        edge.ComProc
	getCanResume                  edge.ComProc
}

func (o *downloadOperation) uri() (string, error) {
	var raw *uint16
	hr, _, _ := o.vtbl.getURI.Call(
		uintptr(unsafe.Pointer(o)),
		uintptr(unsafe.Pointer(&raw)),
	)
	if failed(hr) {
		return "", hresultError(hr)
	}
	defer windows.CoTaskMemFree(unsafe.Pointer(raw))
	return windows.UTF16PtrToString(raw), nil
}

func (o *downloadOperation) totalBytes() (int64, error) {
	var total int64
	hr, _, _ := o.vtbl.getTotalBytesToReceive.Call(
		uintptr(unsafe.Pointer(o)),
		uintptr(unsafe.Pointer(&total)),
	)
	if failed(hr) {
		return -1, hresultError(hr)
	}
	return total, nil
}

func (o *downloadOperation) release() {
	o.vtbl.release.Call(uintptr(unsafe.Pointer(o)))
}

type downloadStartingArgs struct {
	vtbl *downloadStartingArgsVtable
}

type downloadStartingArgsVtable struct {
	iunknownVtable
	getDownloadOperation edge.ComProc
	getCancel            edge.ComProc
	putCancel            edge.ComProc
	getResultFilePath    edge.ComProc
	putResultFilePath    edge.ComProc
	getHandled           edge.ComProc
	putHandled           edge.ComProc
	getDeferral          edge.ComProc
}

func (a *downloadStartingArgs) downloadOperation() (*downloadOperation, error) {
	var operation *downloadOperation
	hr, _, _ := a.vtbl.getDownloadOperation.Call(
		uintptr(unsafe.Pointer(a)),
		uintptr(unsafe.Pointer(&operation)),
	)
	if failed(hr) {
		return nil, hresultError(hr)
	}
	if operation == nil {
		return nil, errors.New("download operation is not available")
	}
	return operation, nil
}

func (a *downloadStartingArgs) resultFilePath() (string, error) {
	var raw *uint16
	hr, _, _ := a.vtbl.getResultFilePath.Call(
		uintptr(unsafe.Pointer(a)),
		uintptr(unsafe.Pointer(&raw)),
	)
	if failed(hr) {
		return "", hresultError(hr)
	}
	defer windows.CoTaskMemFree(unsafe.Pointer(raw))
	return windows.UTF16PtrToString(raw), nil
}

func (a *downloadStartingArgs) putCancel(cancel bool) error {
	hr, _, _ := a.vtbl.putCancel.Call(
		uintptr(unsafe.Pointer(a)),
		boolArg(cancel),
	)
	if failed(hr) {
		return hresultError(hr)
	}
	return nil
}

func (a *downloadStartingArgs) putHandled(handled bool) error {
	hr, _, _ := a.vtbl.putHandled.Call(
		uintptr(unsafe.Pointer(a)),
		boolArg(handled),
	)
	if failed(hr) {
		return hresultError(hr)
	}
	return nil
}

type navigationStartingArgs struct {
	vtbl *navigationStartingArgsVtable
}

type navigationStartingArgsVtable struct {
	iunknownVtable
	getURI             edge.ComProc
	getIsUserInitiated edge.ComProc
	getIsRedirected    edge.ComProc
	getRequestHeaders  edge.ComProc
	getCancel          edge.ComProc
	putCancel          edge.ComProc
	getNavigationID    edge.ComProc
}

func (a *navigationStartingArgs) uri() (string, error) {
	var raw *uint16
	hr, _, _ := a.vtbl.getURI.Call(
		uintptr(unsafe.Pointer(a)),
		uintptr(unsafe.Pointer(&raw)),
	)
	if failed(hr) {
		return "", hresultError(hr)
	}
	defer windows.CoTaskMemFree(unsafe.Pointer(raw))
	return windows.UTF16PtrToString(raw), nil
}

func (a *navigationStartingArgs) putCancel(cancel bool) error {
	hr, _, _ := a.vtbl.putCancel.Call(
		uintptr(unsafe.Pointer(a)),
		boolArg(cancel),
	)
	if failed(hr) {
		return hresultError(hr)
	}
	return nil
}

type newWindowArgs struct {
	vtbl *newWindowArgsVtable
}

type newWindowArgsVtable struct {
	iunknownVtable
	getURI             edge.ComProc
	putNewWindow       edge.ComProc
	getNewWindow       edge.ComProc
	putHandled         edge.ComProc
	getHandled         edge.ComProc
	getIsUserInitiated edge.ComProc
	getDeferral        edge.ComProc
	getWindowFeatures  edge.ComProc
}

func (a *newWindowArgs) uri() (string, error) {
	var raw *uint16
	hr, _, _ := a.vtbl.getURI.Call(
		uintptr(unsafe.Pointer(a)),
		uintptr(unsafe.Pointer(&raw)),
	)
	if failed(hr) {
		return "", hresultError(hr)
	}
	defer windows.CoTaskMemFree(unsafe.Pointer(raw))
	return windows.UTF16PtrToString(raw), nil
}

func (a *newWindowArgs) putHandled(handled bool) error {
	hr, _, _ := a.vtbl.putHandled.Call(
		uintptr(unsafe.Pointer(a)),
		boolArg(handled),
	)
	if failed(hr) {
		return hresultError(hr)
	}
	return nil
}

func boolArg(value bool) uintptr {
	if value {
		return 1
	}
	return 0
}

func failed(hr uintptr) bool {
	return int32(hr) < 0
}

func hresultError(hr uintptr) error {
	return fmt.Errorf("COM call failed with HRESULT 0x%08X", uint32(hr))
}
