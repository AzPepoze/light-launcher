#ifndef LIGHT_LAUNCHER_SHELL_H
#define LIGHT_LAUNCHER_SHELL_H

#include <stdint.h>

int ll_shell_run(uintptr_t handle,
                 const char *url,
                 const char *bridge_script,
                 const char *profile_dir,
                 const char *title,
                 const char *banner,
                 int transparent,
                 int assist,
                 int devtools,
                 int disable_gpu);
void ll_shell_quit(void);
void ll_shell_eval(const char *script);
void ll_shell_pick(int id, int folder);
int ll_open_uri(const char *uri);

#endif
