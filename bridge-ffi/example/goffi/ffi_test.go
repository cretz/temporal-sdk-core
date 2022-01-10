package goffi_test

import (
	"log"
	"time"

	"github.com/temporalio/sdk-core/bridge-ffi/example/goffi"
	bridgepb "github.com/temporalio/sdk-core/bridge-ffi/example/goffi/corepb/bridgepb"
)

func Example() {
	rt := goffi.NewRuntime()
	defer rt.Close()
	log.Printf("pre-core")
	_, err := goffi.NewCore(rt, &bridgepb.InitRequest{})
	log.Printf("post-core: %v", err)
	time.Sleep(2 * time.Second)
	log.Printf("done waiting")
	// Output:
}
