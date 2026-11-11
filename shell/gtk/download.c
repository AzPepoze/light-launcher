// download.c captures the first assist download. WebKit reports the request before the response
// headers exist, so the capture waits for decide-destination, reads the real filename and size,
// cancels the download, then asks the profile cookie manager for the Cookie header.
#include "_cgo_export.h"
#include "shell_internal.h"

#include <libsoup/soup.h>

typedef struct {
	ShellApp *app;
	WebKitDownload *download;
	char *uri, *filename, *referer, *user_agent;
	long long size;
	int started, emitted;
	guint capture_timeout, cookie_timeout;
} Capture;

static void free_capture(Capture *capture) {
	if (capture->download != NULL)
		g_object_unref(capture->download);
	if (capture->capture_timeout != 0)
		g_source_remove(capture->capture_timeout);
	if (capture->cookie_timeout != 0)
		g_source_remove(capture->cookie_timeout);
	g_free(capture->uri);
	g_free(capture->filename);
	g_free(capture->referer);
	g_free(capture->user_agent);
	g_free(capture);
}

static void emit_download(Capture *capture, const char *cookie) {
	if (capture->emitted)
		return;
	capture->emitted = 1;
	llGoDownload(capture->app->handle, capture->uri, capture->filename, capture->size,
		     (char *)(cookie != NULL ? cookie : ""), capture->referer, capture->user_agent);
}

static gboolean on_cookie_timeout(gpointer data) {
	Capture *capture = data;
	capture->cookie_timeout = 0;
	emit_download(capture, "");
	return G_SOURCE_REMOVE;
}

static void on_cookies(GObject *source, GAsyncResult *result, gpointer data) {
	Capture *capture = data;
	if (capture->cookie_timeout != 0) {
		g_source_remove(capture->cookie_timeout);
		capture->cookie_timeout = 0;
	}
	GError *error = NULL;
	GList *cookies = webkit_cookie_manager_get_cookies_finish(
		WEBKIT_COOKIE_MANAGER(source), result, &error);
	GString *header = g_string_new(NULL);
	for (GList *item = cookies; item != NULL; item = item->next) {
		SoupCookie *cookie = item->data;
		const char *name = soup_cookie_get_name(cookie);
		if (name == NULL || *name == '\0')
			continue;
		if (header->len > 0)
			g_string_append(header, "; ");
		g_string_append_printf(header, "%s=%s", name, soup_cookie_get_value(cookie));
	}
	if (error != NULL) {
		llGoWarn(capture->app->handle, (char *)"cannot read the profile cookies");
		g_error_free(error);
	}
	g_list_free_full(cookies, (GDestroyNotify)soup_cookie_free);
	emit_download(capture, header->str);
	g_string_free(header, TRUE);
	free_capture(capture);
}

// start_capture reads the response headers, cancels the real download and asks for its cookies.
// The download is not given a destination, so WebKit writes nothing before this point.
static void start_capture(Capture *capture, const char *suggested) {
	if (capture->started)
		return;
	capture->started = 1;
	if (capture->capture_timeout != 0) {
		g_source_remove(capture->capture_timeout);
		capture->capture_timeout = 0;
	}
	WebKitURIResponse *response = capture->download != NULL
		? webkit_download_get_response(capture->download) : NULL;
	if (suggested != NULL && *suggested != '\0') {
		g_free(capture->filename);
		capture->filename = g_strdup(suggested);
	}
	if (response != NULL) {
		guint64 length = webkit_uri_response_get_content_length(response);
		if (length > 0)
			capture->size = (long long)length;
	}
	if (capture->download != NULL) {
		webkit_download_cancel(capture->download);
		g_object_unref(capture->download);
		capture->download = NULL;
	}
	if (ll_current == NULL || ll_current->session == NULL) {
		emit_download(capture, "");
		free_capture(capture);
		return;
	}
	WebKitCookieManager *manager = webkit_network_session_get_cookie_manager(ll_current->session);
	webkit_cookie_manager_get_cookies(manager, capture->uri, NULL, on_cookies, capture);
	capture->cookie_timeout = g_timeout_add_seconds(3, on_cookie_timeout, capture);
}

static gboolean on_capture_timeout(gpointer data) {
	Capture *capture = data;
	capture->capture_timeout = 0;
	start_capture(capture, NULL);
	return G_SOURCE_REMOVE;
}

static gboolean on_decide_destination(WebKitDownload *download, const gchar *suggested, gpointer data) {
	start_capture(data, suggested);
	return FALSE;
}

static void on_download_started(WebKitNetworkSession *session, WebKitDownload *download, gpointer data) {
	ShellApp *app = data;
	if (app->captured) {
		webkit_download_cancel(download);
		return;
	}
	WebKitURIRequest *request = webkit_download_get_request(download);
	const char *uri = request != NULL ? webkit_uri_request_get_uri(request) : NULL;
	if (uri == NULL || !llGoCaptureDownload(app->handle, (char *)uri)) {
		webkit_download_cancel(download);
		ll_app_warn(app, "ignored a download on another scheme");
		return;
	}
	app->captured = 1;
	const char *page = webkit_web_view_get_uri(app->view);
	const char *user_agent = webkit_settings_get_user_agent(webkit_web_view_get_settings(app->view));
	Capture *capture = g_new0(Capture, 1);
	capture->app = app;
	capture->uri = g_strdup(uri);
	capture->filename = g_strdup("");
	capture->referer = g_strdup(page != NULL ? page : "");
	capture->user_agent = g_strdup(user_agent != NULL ? user_agent : "");
	capture->size = -1;
	capture->download = g_object_ref(download);
	g_signal_connect(download, "decide-destination", G_CALLBACK(on_decide_destination), capture);
	capture->capture_timeout = g_timeout_add_seconds(3, on_capture_timeout, capture);
}

void ll_download_connect(ShellApp *app) {
	g_signal_connect(app->session, "download-started", G_CALLBACK(on_download_started), app);
}
