#ifndef LIGHT_LAUNCHER_SHELL_H
#define LIGHT_LAUNCHER_SHELL_H

#include <stdint.h>

typedef struct {
	int x, y, width, height, positioned;
} LlGeometry;

int ll_shell_run(uintptr_t handle,
                 const char *url,
                 const char *bridge_script,
                 const char *profile_dir,
                 const char *title,
                 const char *banner,
                 int transparent,
                 int assist,
                 int devtools,
                 int disable_gpu,
                 LlGeometry geometry);
int ll_shell_screen_count(void);
int ll_shell_screen_area(int index, int *x, int *y, int *width, int *height);
void ll_shell_quit(void);
void ll_shell_eval(const char *script);
void ll_shell_pick(int id, int folder);
int ll_open_uri(const char *uri);

#endif
