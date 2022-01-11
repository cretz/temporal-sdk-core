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

typedef void (*core_callback_fn)(void* user_data, struct tmprl_bytes_t* bytes);

extern void callback_core(void* user_data, struct tmprl_bytes_t* bytes);
*/
import "C"
import (
	"fmt"
	"runtime/cgo"
	"unsafe"

	"github.com/gogo/protobuf/proto"
	bridgepb "github.com/temporalio/sdk-core/bridge-ffi/example/goffi/corepb/bridgepb"
)

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

func NewCore(runtime *Runtime, in *bridgepb.InitRequest) (*Core, error) {
	var core Core
	var resp bridgepb.InitResponse
	req, inPtr, inLen, reqHandle := core.newRequest(in, &resp)
	C.tmprl_core_init(runtime.runtime, inPtr, inLen, reqHandle, initCallback)
	<-req.ch
	// Ignore response unless core is nil
	if req.core.core == nil {
		return nil, fmt.Errorf("failed initializing: %v", resp.Error.Message)
	}
	return req.core, nil
}

func (c *Core) Shutdown() {
	req, inPtr, inLen, reqHandle := c.newRequest(nil, nil)
	C.tmprl_core_shutdown(c.core, inPtr, inLen, reqHandle, callback)
	<-req.ch
}

func (c *Core) newRequest(
	reqProto interface{ Marshal() ([]byte, error) },
	respProto proto.Message,
) (req *request, inputPtr *C.uint8_t, inputLen C.size_t, reqHandle unsafe.Pointer) {
	req = &request{core: c, resp: respProto, ch: make(chan *request, 1)}
	if reqProto != nil {
		var err error
		if req.inputBytes, err = reqProto.Marshal(); err != nil {
			panic(err)
		}
		inputPtr, inputLen = bytesPtrAndLen(req.inputBytes)
	}
	reqHandle = unsafe.Pointer(cgo.NewHandle(req))
	return
}

func bytesPtrAndLen(b []byte) (*C.uint8_t, C.size_t) {
	return (*C.uint8_t)(C.CBytes(b)), C.size_t(len(b))
}

func protoFromBytes(bytes *C.tmprl_bytes_t, p proto.Message) error {
	// TODO(cretz): Why is the pointer not valid on an empty byte set?
	if bytes.len == 0 {
		return nil
	}
	in := (*[1<<30 - 1]byte)(unsafe.Pointer(bytes.bytes))
	return proto.Unmarshal(in[:bytes.len], p)
}

type request struct {
	// Expected to be passed in except for core init requests where it is set
	core *Core
	// This is set to nil on unmarshal error
	resp proto.Message
	// Sent on complete, should have buffer of 1 since non-blocking send is used
	ch chan *request
	// Only used to maintain a reference so it's not GC'd
	inputBytes []byte
}

//export go_callback_core_init
func go_callback_core_init(user_data C.uintptr_t, core *C.tmprl_core_t, bytes *C.tmprl_bytes_t) {
	// Set core (which may be nil) then invoke normal callback
	cgo.Handle(user_data).Value().(*request).core = &Core{core}
	go_callback_core(user_data, bytes)
}

var initCallback = C.core_init_callback_fn(C.callback_core_init)
var callback = C.core_callback_fn(C.callback_core)

//export go_callback_core
func go_callback_core(user_data C.uintptr_t, bytes *C.tmprl_bytes_t) {
	h := cgo.Handle(user_data)
	defer h.Delete()
	req := h.Value().(*request)

	// Deserialize
	defer C.tmprl_bytes_free(req.core.core, bytes)
	if req.resp != nil {
		if err := protoFromBytes(bytes, req.resp); err != nil {
			// TODO(cretz): Warn?
			req.resp = nil
		}
	}

	// Non-blocking channel send
	select {
	case req.ch <- req:
	default:
	}
}