//go:build wasip1

package main

// A minimal waPC guest ABI (https://wapc.io/docs/spec/) for the standard Go
// toolchain, in place of github.com/wapc/wapc-guest-tinygo.

import (
	"errors"
	"unsafe"
)

//go:wasmimport wapc __guest_request
func guestRequest(operation, payload unsafe.Pointer)

//go:wasmimport wapc __guest_response
func guestResponse(ptr unsafe.Pointer, n uint32)

//go:wasmimport wapc __guest_error
func guestError(ptr unsafe.Pointer, n uint32)

//go:wasmimport wapc __host_call
func wapcHostCall(bindingPtr unsafe.Pointer, bindingLen uint32,
	namespacePtr unsafe.Pointer, namespaceLen uint32,
	operationPtr unsafe.Pointer, operationLen uint32,
	payloadPtr unsafe.Pointer, payloadLen uint32) uint32

//go:wasmimport wapc __host_response_len
func hostResponseLen() uint32

//go:wasmimport wapc __host_response
func hostResponse(ptr unsafe.Pointer)

//go:wasmimport wapc __host_error_len
func hostErrorLen() uint32

//go:wasmimport wapc __host_error
func hostError(ptr unsafe.Pointer)

// bytesPtr returns a pointer the host may read or write len(b) bytes at. A
// zero-length slice gets a valid pointer too.
func bytesPtr(b []byte) unsafe.Pointer {
	if len(b) == 0 {
		return unsafe.Pointer(&zero)
	}
	return unsafe.Pointer(unsafe.SliceData(b))
}

func stringPtr(s string) unsafe.Pointer {
	if len(s) == 0 {
		return unsafe.Pointer(&zero)
	}
	return unsafe.Pointer(unsafe.StringData(s))
}

var zero byte

//go:wasmexport __guest_call
func guestCall(operationSize, payloadSize uint32) uint32 {
	operation := make([]byte, operationSize)
	payload := make([]byte, payloadSize)
	guestRequest(bytesPtr(operation), bytesPtr(payload))

	fn, ok := functions[string(operation)]
	if !ok {
		msg := `Could not find function "` + string(operation) + `"`
		guestError(stringPtr(msg), uint32(len(msg)))
		return 0
	}
	resp, err := fn(payload)
	if err != nil {
		msg := err.Error()
		guestError(stringPtr(msg), uint32(len(msg)))
		return 0
	}
	guestResponse(bytesPtr(resp), uint32(len(resp)))
	return 1
}

func hostCall(binding, namespace, operation string, payload []byte) ([]byte, error) {
	ok := wapcHostCall(
		stringPtr(binding), uint32(len(binding)),
		stringPtr(namespace), uint32(len(namespace)),
		stringPtr(operation), uint32(len(operation)),
		bytesPtr(payload), uint32(len(payload)))
	if ok == 0 {
		msg := make([]byte, hostErrorLen())
		hostError(bytesPtr(msg))
		return nil, errors.New("Host error: " + string(msg))
	}
	resp := make([]byte, hostResponseLen())
	hostResponse(bytesPtr(resp))
	return resp, nil
}
