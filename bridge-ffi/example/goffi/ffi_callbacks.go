package goffi

/*
#cgo CFLAGS:-I${SRCDIR}/../../include
#include <sdk-core-bridge.h>

extern void go_callback_core_init(void*, tmprl_core_t*, tmprl_bytes_t*);

void callback_core_init(void* user_data, struct tmprl_core_t* core, struct tmprl_bytes_t* bytes) {
	go_callback_core_init(user_data, core, bytes);
}
*/
import "C"
