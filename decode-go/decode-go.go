package decode

import (
	"context"
	_ "embed"
	"fmt"
	"os"

	"github.com/wapc/wapc-go"
	"github.com/wapc/wapc-go/engines/wazero"
)

//go:embed decode-wapc/decode-wapc.wasm
var guestDecodeWasm []byte

func Decode(ctx context.Context, getKey func(string) string, enc string) (string, error) {
	engine := wazero.Engine()

	module, err := engine.New(ctx, makeHost(getKey), guestDecodeWasm, &wapc.ModuleConfig{
		Logger: wapc.PrintlnLogger,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	})
	if err != nil {
		return "", err
	}
	defer module.Close(ctx)

	instance, err := module.Instantiate(ctx)
	if err != nil {
		return "", err
	}
	defer instance.Close(ctx)

	result, err := instance.Invoke(ctx, "decode", []byte(enc))
	if err != nil {
		return "", err
	}

	return string(result), nil
}

func makeHost(getKey func(string) string) wapc.HostCallHandler {
	return func(_ context.Context, binding, namespace, operation string, payload []byte) ([]byte, error) {
		switch namespace {
		case "libmxf":
			switch operation {
			case "get_key":
				return []byte(getKey(string(payload))), nil
			}
		}
		return nil, fmt.Errorf("unknown op %s:%s:%s", binding, namespace, operation)
	}
}
