package decode

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"os"

	"github.com/wapc/wapc-go"
	"github.com/wapc/wapc-go/engines/wazero"
)

//go:embed decode-wapc/decode-wapc.wasm
var guestDecodeWasm []byte

// The guest module is compiled once per process and shared by every Decode.
// Compiling the 8.5 MB guest takes seconds; instantiating a compiled module
// takes milliseconds. Each Decode gets its own instance, so concurrent calls
// share no guest memory.
var (
	// guestLock is a mutex that a waiter can leave when its context ends:
	// hold it by sending, release it by receiving.
	guestLock   = make(chan struct{}, 1)
	guestModule wapc.Module
)

// compiledGuest returns the shared compiled guest, compiling it on the first
// call. A failed compile is not cached, so a later call tries again. A caller
// whose ctx ends while another call compiles returns ctx.Err().
func compiledGuest(ctx context.Context) (wapc.Module, error) {
	select {
	case guestLock <- struct{}{}:
	case <-ctx.Done():
		return nil, ctx.Err()
	}
	defer func() { <-guestLock }()
	if guestModule != nil {
		return guestModule, nil
	}
	// The module outlives the call that compiles it, so it does not take
	// the caller's context.
	module, err := wazero.Engine().New(context.Background(), hostCall, guestDecodeWasm, &wapc.ModuleConfig{
		Logger: wapc.PrintlnLogger,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	})
	if err != nil {
		return nil, err
	}
	guestModule = module
	return guestModule, nil
}

// Decode decodes the MXF message enc, calling getKey for each data key the
// message needs. It is safe for concurrent use. A caller whose ctx has ended
// before the guest starts, or ends while it waits for the first compile,
// gets ctx.Err(). A guest that is already running is not stopped.
func Decode(ctx context.Context, getKey func(string) string, enc string) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	module, err := compiledGuest(ctx)
	if err != nil {
		return "", err
	}

	instance, err := module.Instantiate(ctx)
	if err != nil {
		return "", err
	}
	defer instance.Close(ctx)

	// The shared host handler reads this call's getKey from the context
	// the guest's host calls run under.
	result, err := instance.Invoke(withGetKey(ctx, getKey), "decode", []byte(enc))
	if err != nil {
		return "", err
	}

	return string(result), nil
}

type getKeyContextKey struct{}

func withGetKey(ctx context.Context, getKey func(string) string) context.Context {
	return context.WithValue(ctx, getKeyContextKey{}, getKey)
}

var errNoGetKey = errors.New("no getKey for this call")

// hostCall answers the guest's host calls. waPC runs it under the context
// passed to Invoke.
func hostCall(ctx context.Context, binding, namespace, operation string, payload []byte) ([]byte, error) {
	switch namespace {
	case "libmxf":
		switch operation {
		case "get_key":
			getKey, ok := ctx.Value(getKeyContextKey{}).(func(string) string)
			if !ok || getKey == nil {
				return nil, errNoGetKey
			}
			return []byte(getKey(string(payload))), nil
		}
	}
	return nil, fmt.Errorf("unknown op %s:%s:%s", binding, namespace, operation)
}
