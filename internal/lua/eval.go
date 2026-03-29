package lua

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"time"

	lua "github.com/yuin/gopher-lua"
)

// Evaluator runs Lua code in a sandboxed environment.
type Evaluator struct {
	timeout time.Duration
}

// NewEvaluator creates a new Lua evaluator with the given execution timeout.
func NewEvaluator(timeout time.Duration) *Evaluator {
	return &Evaluator{timeout: timeout}
}

// Eval executes Lua code and returns captured output. Dangerous modules
// (os, io) and functions (loadfile, dofile) are removed. Execution is
// cancelled after the configured timeout.
func (e *Evaluator) Eval(code string) (string, error) {
	if code == "" {
		return "", nil
	}

	L := lua.NewState(lua.Options{SkipOpenLibs: true})
	defer L.Close()

	// Open only safe libs
	for _, lib := range []struct {
		name string
		fn   lua.LGFunction
	}{
		{lua.LoadLibName, lua.OpenPackage},
		{lua.BaseLibName, lua.OpenBase},
		{lua.TabLibName, lua.OpenTable},
		{lua.StringLibName, lua.OpenString},
		{lua.MathLibName, lua.OpenMath},
	} {
		L.Push(L.NewFunction(lib.fn))
		L.Push(lua.LString(lib.name))
		L.Call(1, 0)
	}

	// Remove dangerous globals
	L.SetGlobal("loadfile", lua.LNil)
	L.SetGlobal("dofile", lua.LNil)

	// Capture print output
	var buf bytes.Buffer
	L.SetGlobal("print", L.NewFunction(func(L *lua.LState) int {
		n := L.GetTop()
		for i := 1; i <= n; i++ {
			if i > 1 {
				buf.WriteByte('\t')
			}
			buf.WriteString(L.ToStringMeta(L.Get(i)).String())
		}
		buf.WriteByte('\n')
		return 0
	}))

	ctx, cancel := context.WithTimeout(context.Background(), e.timeout)
	defer cancel()
	L.SetContext(ctx)

	if err := L.DoString(code); err != nil {
		if ctx.Err() != nil {
			return "", fmt.Errorf("execution timeout after %s", e.timeout)
		}
		// Clean up gopher-lua error prefix
		msg := err.Error()
		if strings.HasPrefix(msg, "<string>:") {
			msg = "line " + msg[len("<string>:"):]
		}
		return "", fmt.Errorf("%s", msg)
	}

	// Capture return value if present
	ret := L.Get(-1)
	if ret != lua.LNil && ret.Type() != lua.LTFunction {
		// Only append if DoString pushed a value
		top := L.GetTop()
		if top > 0 {
			buf.WriteString(ret.String())
			buf.WriteByte('\n')
		}
	}

	return buf.String(), nil
}
