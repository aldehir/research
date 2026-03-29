package lua

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEval_BasicPrint(t *testing.T) {
	e := NewEvaluator(5 * time.Second)
	result, err := e.Eval(`print("hello world")`)
	require.NoError(t, err)
	assert.Equal(t, "hello world\n", result)
}

func TestEval_MultiplePrints(t *testing.T) {
	e := NewEvaluator(5 * time.Second)
	result, err := e.Eval(`print("line 1")
print("line 2")`)
	require.NoError(t, err)
	assert.Equal(t, "line 1\nline 2\n", result)
}

func TestEval_PrintMultipleArgs(t *testing.T) {
	e := NewEvaluator(5 * time.Second)
	result, err := e.Eval(`print(1, 2, 3)`)
	require.NoError(t, err)
	assert.Equal(t, "1\t2\t3\n", result)
}

func TestEval_ReturnValue(t *testing.T) {
	e := NewEvaluator(5 * time.Second)
	result, err := e.Eval(`return 42`)
	require.NoError(t, err)
	assert.Equal(t, "42\n", result)
}

func TestEval_MathOperations(t *testing.T) {
	e := NewEvaluator(5 * time.Second)
	result, err := e.Eval(`print(2 + 3 * 4)`)
	require.NoError(t, err)
	assert.Equal(t, "14\n", result)
}

func TestEval_SyntaxError(t *testing.T) {
	e := NewEvaluator(5 * time.Second)
	_, err := e.Eval(`if then end`)
	require.Error(t, err)
}

func TestEval_RuntimeError(t *testing.T) {
	e := NewEvaluator(5 * time.Second)
	_, err := e.Eval(`error("something broke")`)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "something broke")
}

func TestEval_Timeout(t *testing.T) {
	e := NewEvaluator(100 * time.Millisecond)
	_, err := e.Eval(`while true do end`)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "timeout")
}

func TestEval_SandboxNoOS(t *testing.T) {
	e := NewEvaluator(5 * time.Second)
	_, err := e.Eval(`os.execute("echo hi")`)
	require.Error(t, err)
}

func TestEval_SandboxNoIO(t *testing.T) {
	e := NewEvaluator(5 * time.Second)
	_, err := e.Eval(`io.open("/etc/passwd")`)
	require.Error(t, err)
}

func TestEval_SandboxNoLoadfile(t *testing.T) {
	e := NewEvaluator(5 * time.Second)
	_, err := e.Eval(`loadfile("/etc/passwd")`)
	require.Error(t, err)
}

func TestEval_SandboxNoDofile(t *testing.T) {
	e := NewEvaluator(5 * time.Second)
	_, err := e.Eval(`dofile("/etc/passwd")`)
	require.Error(t, err)
}

func TestEval_SafeLibsAvailable(t *testing.T) {
	e := NewEvaluator(5 * time.Second)

	// string lib
	result, err := e.Eval(`print(string.upper("hello"))`)
	require.NoError(t, err)
	assert.Equal(t, "HELLO\n", result)

	// table lib
	result, err = e.Eval(`local t = {3,1,2}; table.sort(t); print(t[1], t[2], t[3])`)
	require.NoError(t, err)
	assert.Equal(t, "1\t2\t3\n", result)

	// math lib
	result, err = e.Eval(`print(math.floor(3.7))`)
	require.NoError(t, err)
	assert.Equal(t, "3\n", result)
}

func TestEval_EmptyCode(t *testing.T) {
	e := NewEvaluator(5 * time.Second)
	result, err := e.Eval(``)
	require.NoError(t, err)
	assert.Equal(t, "", result)
}
