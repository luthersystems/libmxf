package libmxf

import (
	"encoding/hex"
	"fmt"
	"os"

	"github.com/luthersystems/elps/elpsutil"
	"github.com/luthersystems/elps/lisp"
	"github.com/luthersystems/elps/parser"
	"github.com/luthersystems/substrate/chaincodes/substrate01/lib/shiro/libmxf"
)

var env *lisp.LEnv

func init() {
	env = lisp.NewEnv(nil)
	env.Runtime.Reader = parser.NewReader()
	rc := lisp.InitializeUserEnv(env)
	if !rc.IsNil() {
		fmt.Fprintln(os.Stderr, rc)
		os.Exit(1)
	}

	loaders := libmxf.RequiredLoaders()
	loaders = append(loaders, func(env *lisp.LEnv) *lisp.LVal {
		env.AddBuiltins(true, elpsutil.Function("mxf-wasm-apply", lisp.Formals("get-key", "dsid"), builtinWasmApply))
		return lisp.Nil()
	})

	loader := elpsutil.LoadAll(loaders...)
	rc = loader(env)
	if !rc.IsNil() {
		fmt.Fprintln(os.Stderr, rc)
		os.Exit(1)
	}
	rc = env.InPackage(lisp.String(lisp.DefaultUserPackage))
	if !rc.IsNil() {
		fmt.Fprintln(os.Stderr, rc)
		os.Exit(1)
	}
}

// MxfDecode decodes an encoded message.
func MxfDecode(getKey func(string) string, enc string) *lisp.LVal {
	cells := make([]*lisp.LVal, 0, 4)
	cells = append(cells, lisp.Symbol("mxf:wasm-decode"))
	cells = append(cells, lisp.Symbol("mxf-wasm-apply"))
	cells = append(cells, lisp.Native(getKey))
	cells = append(cells, lisp.String(enc))
	expr := lisp.SExpr(cells)
	out := env.Eval(expr)
	err := lisp.GoError(out)
	if err != nil {
		return lisp.Error(fmt.Errorf("internal error: %w", err))
	}

	return out
}

// Log outputs to console.
func Log(msg string) {
	expr := lisp.SExpr([]*lisp.LVal{lisp.Symbol("debug-print"), lisp.String(msg)})
	result := env.Eval(expr)
	fmt.Println(result)
}

func builtinWasmApply(env *lisp.LEnv, args *lisp.LVal) *lisp.LVal {
	getKeyFn := args.Cells[0]
	if getKeyFn.Type != lisp.LNative {
		return env.Errorf("first argument is not a native type: %v", getKeyFn.Type)
	}
	fn, ok := getKeyFn.Native.(func(string) string)
	if !ok {
		return env.Errorf("first argument is not a wrapped function")
	}

	dsid := args.Cells[1]
	if dsid.Type != lisp.LString {
		return env.Errorf("second argument is not a string: %v", dsid.Type)
	}
	dsidStr := dsid.Str

	key := fn(dsidStr)

	b, err := hex.DecodeString(key)
	if err != nil {
		return env.Errorf("invalid hex [%s]: %s", key, err)
	}
	return lisp.Bytes(b)
}
