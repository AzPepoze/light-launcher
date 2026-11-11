// shell.c owns the GTK4 window, the bridge and the link policy. Go starts one main loop on the
// thread that called gtk_init, so every GTK call in this file runs on that thread.
#include "_cgo_export.h"
#include "shell.h"
#include "shell_internal.h"

typedef struct {
	uintptr_t handle;
	int id, folder;
	GtkFileDialog *dialog;
} Picker;

ShellApp *ll_current;
int ll_quit_requested;

void ll_app_warn(ShellApp *app, const char *message) {
	llGoWarn(app->handle, (char *)message);
}

static GtkWidget *new_window(const char *title, int width, int height) {
	GtkWidget *window = gtk_window_new();
	g_object_ref_sink(window);
	gtk_window_set_title(GTK_WINDOW(window), title != NULL ? title : "LightLauncher");
	gtk_window_set_default_size(GTK_WINDOW(window), width, height);
	gtk_widget_set_size_request(window, 800, 600);
	return window;
}

static WebKitWebView *new_view(WebKitNetworkSession *session) {
	if (session != NULL)
		return WEBKIT_WEB_VIEW(g_object_new(WEBKIT_TYPE_WEB_VIEW, "network-session", session, NULL));
	return WEBKIT_WEB_VIEW(g_object_new(WEBKIT_TYPE_WEB_VIEW, NULL));
}

static void install_css(void) {
	GtkCssProvider *provider = gtk_css_provider_new();
	gtk_css_provider_load_from_string(provider,
		".assist-banner { background-color: #1f2937; color: #ffffff; }\n"
		"window.transparent { background-color: transparent; }\n");
	gtk_style_context_add_provider_for_display(gdk_display_get_default(),
						  GTK_STYLE_PROVIDER(provider),
						  GTK_STYLE_PROVIDER_PRIORITY_APPLICATION);
	g_object_unref(provider);
}

static void configure_view(ShellApp *app, int disable_gpu, int devtools, int transparent) {
	WebKitSettings *settings = webkit_web_view_get_settings(app->view);
	webkit_settings_set_enable_smooth_scrolling(settings, TRUE);
	webkit_settings_set_hardware_acceleration_policy(settings,
		disable_gpu ? WEBKIT_HARDWARE_ACCELERATION_POLICY_NEVER
			    : WEBKIT_HARDWARE_ACCELERATION_POLICY_ALWAYS);
	webkit_settings_set_enable_developer_extras(settings, devtools ? TRUE : FALSE);
	if (transparent) {
		GdkRGBA clear = {0, 0, 0, 0};
		webkit_web_view_set_background_color(app->view, &clear);
		gtk_widget_add_css_class(app->window, "transparent");
	}
}

static void on_script_message(WebKitUserContentManager *manager, JSCValue *value, gpointer data) {
	ShellApp *app = data;
	char *body = jsc_value_to_string(value);
	const char *page = webkit_web_view_get_uri(app->view);
	llGoBridgeMessage(app->handle, body != NULL ? body : (char *)"",
			  (char *)(page != NULL ? page : ""));
	g_free(body);
}

static void install_bridge(ShellApp *app, const char *source) {
	if (source == NULL || *source == '\0')
		return;
	WebKitUserContentManager *manager = webkit_web_view_get_user_content_manager(app->view);
	g_signal_connect(manager, "script-message-received::llshell",
			 G_CALLBACK(on_script_message), app);
	if (!webkit_user_content_manager_register_script_message_handler(manager, "llshell", NULL)) {
		ll_app_warn(app, "cannot register the bridge message handler");
		return;
	}
	WebKitUserScript *script = webkit_user_script_new(source,
		WEBKIT_USER_CONTENT_INJECT_TOP_FRAME,
		WEBKIT_USER_SCRIPT_INJECT_AT_DOCUMENT_START, NULL, NULL);
	webkit_user_content_manager_add_script(manager, script);
	webkit_user_script_unref(script);
}

static void on_picker_done(GObject *source, GAsyncResult *result, gpointer data) {
	Picker *picker = data;
	GError *error = NULL;
	GFile *file = picker->folder
		? gtk_file_dialog_select_folder_finish(picker->dialog, result, &error)
		: gtk_file_dialog_open_finish(picker->dialog, result, &error);
	char *path = file != NULL ? g_file_get_path(file) : NULL;
	if (file != NULL)
		g_object_unref(file);
	if (error != NULL) {
		if (!g_error_matches(error, GTK_DIALOG_ERROR, GTK_DIALOG_ERROR_DISMISSED))
			llGoWarn(picker->handle, (char *)"file dialog failed");
		g_error_free(error);
	}
	llGoPickerResult(picker->handle, picker->id, path);
	g_free(path);
	g_object_unref(picker->dialog);
	g_free(picker);
}

void ll_shell_pick(int id, int folder) {
	if (ll_current == NULL)
		return;
	GtkFileDialog *dialog = gtk_file_dialog_new();
	gtk_file_dialog_set_title(dialog, folder ? "Choose a folder" : "Choose a file");
	gtk_file_dialog_set_modal(dialog, TRUE);
	Picker *picker = g_new0(Picker, 1);
	picker->handle = ll_current->handle;
	picker->id = id;
	picker->folder = folder;
	picker->dialog = g_object_ref(dialog);
	if (folder)
		gtk_file_dialog_select_folder(dialog, GTK_WINDOW(ll_current->window), NULL,
					      on_picker_done, picker);
	else
		gtk_file_dialog_open(dialog, GTK_WINDOW(ll_current->window), NULL,
				     on_picker_done, picker);
	g_object_unref(dialog);
}

void ll_shell_eval(const char *script) {
	if (ll_current != NULL && ll_current->view != NULL)
		webkit_web_view_evaluate_javascript(ll_current->view, script, -1, NULL, NULL, NULL, NULL, NULL);
}

void ll_shell_quit(void) {
	ll_quit_requested = 1;
	if (ll_current != NULL && ll_current->loop != NULL)
		g_main_loop_quit(ll_current->loop);
}

int ll_open_uri(const char *uri) {
	return g_app_info_launch_default_for_uri(uri, NULL, NULL) ? 1 : 0;
}

static gboolean on_close_request(GtkWindow *window, gpointer data) {
	ShellApp *app = data;
	app->closed = 1;
	llGoClosed(app->handle);
	ll_quit_requested = 1;
	if (app->loop != NULL)
		g_main_loop_quit(app->loop);
	return FALSE;
}

static gboolean on_decide_policy(WebKitWebView *view, WebKitPolicyDecision *decision,
				 WebKitPolicyDecisionType type, gpointer data) {
	ShellApp *app = data;
	if (type != WEBKIT_POLICY_DECISION_TYPE_NAVIGATION_ACTION)
		return FALSE;
	WebKitNavigationAction *action =
		webkit_navigation_policy_decision_get_navigation_action(
			WEBKIT_NAVIGATION_POLICY_DECISION(decision));
	if (action == NULL)
		return FALSE;
	WebKitURIRequest *request = webkit_navigation_action_get_request(action);
	const char *uri = request != NULL ? webkit_uri_request_get_uri(request) : NULL;
	if (uri == NULL)
		return FALSE;
	if (app->assist) {
		if (!llGoWebURL(app->handle, (char *)uri)) {
			ll_app_warn(app, "blocked navigation on another scheme");
			webkit_policy_decision_ignore(decision);
			return TRUE;
		}
		return FALSE;
	}
	if (webkit_navigation_action_get_navigation_type(action) == WEBKIT_NAVIGATION_TYPE_LINK_CLICKED &&
	    llGoLinkClicked(app->handle, (char *)uri)) {
		webkit_policy_decision_ignore(decision);
		return TRUE;
	}
	return FALSE;
}

static GtkWidget *on_create(WebKitWebView *view, WebKitNavigationAction *action, gpointer data) {
	ShellApp *app = data;
	WebKitURIRequest *request = webkit_navigation_action_get_request(action);
	const char *uri = request != NULL ? webkit_uri_request_get_uri(request) : NULL;
	if (uri == NULL)
		return NULL;
	if (app->assist) {
		if (llGoWebURL(app->handle, (char *)uri))
			webkit_web_view_load_uri(app->view, uri);
		else
			ll_app_warn(app, "blocked a new window on another scheme");
		return NULL;
	}
	llGoLinkClicked(app->handle, (char *)uri);
	return NULL;
}

static void shell_app_free(ShellApp *app) {
	if (app->window != NULL) {
		if (!app->closed)
			gtk_window_destroy(GTK_WINDOW(app->window));
		g_object_unref(app->window);
	}
	if (app->session != NULL)
		g_object_unref(app->session);
	if (app->loop != NULL)
		g_main_loop_unref(app->loop);
	g_free(app);
}

int ll_shell_run(uintptr_t handle, const char *url, const char *bridge_script,
		 const char *profile_dir, const char *title, const char *banner,
		 int transparent, int assist, int devtools, int disable_gpu) {
	if (!gtk_init_check()) {
		llGoWarn(handle, (char *)"cannot start GTK");
		return 1;
	}
	install_css();
	ll_current = g_new0(ShellApp, 1);
	ll_current->handle = handle;
	ll_current->assist = assist;
	ll_current->loop = g_main_loop_new(NULL, FALSE);
	ll_quit_requested = 0;

	if (assist) {
		char *cache = g_build_filename(profile_dir, "cache", NULL);
		g_mkdir_with_parents(profile_dir, 0700);
		g_mkdir_with_parents(cache, 0700);
		ll_current->session = webkit_network_session_new(profile_dir, cache);
		g_free(cache);
		if (ll_current->session == NULL) {
			ll_app_warn(ll_current, "cannot create the browser profile");
			shell_app_free(ll_current);
			ll_current = NULL;
			return 1;
		}
		webkit_network_session_set_persistent_credential_storage_enabled(ll_current->session, TRUE);
		WebKitCookieManager *cookies = webkit_network_session_get_cookie_manager(ll_current->session);
		char *cookie_file = g_build_filename(profile_dir, "cookies.sqlite", NULL);
		webkit_cookie_manager_set_persistent_storage(cookies, cookie_file,
							     WEBKIT_COOKIE_PERSISTENT_STORAGE_SQLITE);
		g_free(cookie_file);
	}

	ll_current->window = new_window(title, assist ? 1000 : 1100, 720);
	ll_current->view = new_view(ll_current->session);
	configure_view(ll_current, disable_gpu, devtools, transparent && !assist);
	if (assist && banner != NULL && *banner != '\0') {
		GtkWidget *box = gtk_box_new(GTK_ORIENTATION_VERTICAL, 0);
		GtkWidget *label = gtk_label_new(banner);
		gtk_label_set_wrap(GTK_LABEL(label), TRUE);
		gtk_label_set_justify(GTK_LABEL(label), GTK_JUSTIFY_CENTER);
		gtk_widget_set_margin_top(label, 6);
		gtk_widget_set_margin_bottom(label, 6);
		gtk_widget_set_margin_start(label, 12);
		gtk_widget_set_margin_end(label, 12);
		gtk_widget_add_css_class(label, "assist-banner");
		gtk_box_append(GTK_BOX(box), label);
		gtk_box_append(GTK_BOX(box), GTK_WIDGET(ll_current->view));
		gtk_window_set_child(GTK_WINDOW(ll_current->window), box);
	} else {
		gtk_window_set_child(GTK_WINDOW(ll_current->window), GTK_WIDGET(ll_current->view));
	}

	g_signal_connect(ll_current->view, "decide-policy", G_CALLBACK(on_decide_policy), ll_current);
	g_signal_connect(ll_current->view, "create", G_CALLBACK(on_create), ll_current);
	g_signal_connect(ll_current->window, "close-request", G_CALLBACK(on_close_request), ll_current);
	if (ll_current->session != NULL)
		ll_download_connect(ll_current);
	if (!assist)
		install_bridge(ll_current, bridge_script);

	gtk_window_present(GTK_WINDOW(ll_current->window));
	llGoReady(handle);
	if (!ll_quit_requested) {
		webkit_web_view_load_uri(ll_current->view, url);
		g_main_loop_run(ll_current->loop);
	}
	shell_app_free(ll_current);
	ll_current = NULL;
	return 0;
}
