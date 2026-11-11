#ifndef LIGHT_LAUNCHER_SHELL_INTERNAL_H
#define LIGHT_LAUNCHER_SHELL_INTERNAL_H

#include <gtk/gtk.h>
#include <webkit/webkit.h>
#include <stdint.h>

typedef struct {
	uintptr_t handle;
	GtkWidget *window;
	WebKitWebView *view;
	WebKitNetworkSession *session;
	GMainLoop *loop;
	int assist, closed, captured;
} ShellApp;

extern ShellApp *ll_current;
extern int ll_quit_requested;

void ll_app_warn(ShellApp *app, const char *message);
void ll_download_connect(ShellApp *app);

#endif
