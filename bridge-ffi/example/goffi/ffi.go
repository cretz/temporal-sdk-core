package goffi

/*
#cgo CFLAGS:-I${SRCDIR}/../../include
#cgo !windows LDFLAGS:-ltemporal_sdk_core_bridge_ffi -lm -ldl -pthread
#cgo windows LDFLAGS:-ltemporal_sdk_core_bridge_ffi -luserenv -lole32 -lntdll -lws2_32 -lkernel32 -lsecur32 -lcrypt32 -lbcrypt -lncrypt
#cgo linux,amd64 LDFLAGS:-L${SRCDIR}/lib/linux-x86_64
#cgo linux,arm64 LDFLAGS:-L${SRCDIR}/lib/linux-aarch64
#cgo darwin,amd64 LDFLAGS:-L${SRCDIR}/lib/macos-x86_64
#cgo windows,amd64 LDFLAGS:-L${SRCDIR}/lib/windows-x86_64
#include <sdk-core-bridge.h>

typedef void (*core_init_callback_fn)(void* user_data, struct tmprl_core_t* core, struct tmprl_bytes_t* bytes);

extern void callback_core_init(void* user_data, struct tmprl_core_t* core, struct tmprl_bytes_t* bytes);
*/
import "C"
import (
	"fmt"
	"unsafe"

	bridgepb "github.com/temporalio/sdk-core/bridge-ffi/example/goffi/corepb/bridgepb"
)

//export go_callback_core_init
func go_callback_core_init(user_data unsafe.Pointer, core *C.tmprl_core_t, bytes *C.tmprl_bytes_t) {
	fmt.Printf("YAY, CALLED BACK!\n")
}

func bytesPtrAndLen(b []byte) (*C.uint8_t, C.size_t) {
	return (*C.uint8_t)(C.CBytes(b)), C.size_t(len(b))
}

type Runtime struct {
	runtime *C.tmprl_runtime_t
}

func NewRuntime() *Runtime {
	return &Runtime{runtime: C.tmprl_runtime_new()}
}

func (r *Runtime) Close() {
	if r.runtime != nil {
		C.tmprl_runtime_free(r.runtime)
		r.runtime = nil
	}
}

type Core struct {
	core *C.tmprl_core_t
}

func NewCore(runtime *Runtime, config *bridgepb.InitRequest) (*Core, error) {
	b, err := config.Marshal()
	if err != nil {
		return nil, err
	}
	bPtr, bLen := bytesPtrAndLen(b)
	C.tmprl_core_init(runtime.runtime, bPtr, bLen, nil, C.core_init_callback_fn(C.callback_core_init))
	return nil, nil
}
