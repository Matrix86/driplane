package plugin

import (
	"os"
	"path/filepath"
	"sort"
	"testing"
)

func TestParse(t *testing.T) {
	code := `function hello(name) { return "Hello, " + name; }`
	p, err := Parse(code)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	if p == nil {
		t.Fatal("Parse returned nil")
	}
	if p.Code != code {
		t.Errorf("Code mismatch: expected=%q had=%q", code, p.Code)
	}
}

func TestParseError(t *testing.T) {
	code := `function hello( { invalid syntax`
	_, err := Parse(code)
	if err == nil {
		t.Error("expected error for invalid JS but got nil")
	}
}

func TestLoad(t *testing.T) {
	code := `function greet(name) { return "Hi " + name; }`
	tmpFile := filepath.Join(os.TempDir(), "test_plugin.js")
	if err := os.WriteFile(tmpFile, []byte(code), 0644); err != nil {
		t.Fatalf("cannot create temp file: %v", err)
	}
	defer os.Remove(tmpFile)

	p, err := Load(tmpFile)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if p.Name != "test_plugin" {
		t.Errorf("wrong name: expected=test_plugin had=%s", p.Name)
	}
	if p.Path != tmpFile {
		t.Errorf("wrong path: expected=%s had=%s", tmpFile, p.Path)
	}
}

func TestLoadNonExistent(t *testing.T) {
	_, err := Load("/nonexistent/path/plugin.js")
	if err == nil {
		t.Error("expected error for nonexistent file but got nil")
	}
}

func TestLoadInvalidJS(t *testing.T) {
	code := `function broken( { syntax error`
	tmpFile := filepath.Join(os.TempDir(), "bad_plugin.js")
	if err := os.WriteFile(tmpFile, []byte(code), 0644); err != nil {
		t.Fatalf("cannot create temp file: %v", err)
	}
	defer os.Remove(tmpFile)

	_, err := Load(tmpFile)
	if err == nil {
		t.Error("expected error for invalid JS but got nil")
	}
}

func TestHasFunc(t *testing.T) {
	code := `function myFunc() { return 1; }
var myVar = "hello";`
	p, err := Parse(code)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if !p.HasFunc("myFunc") {
		t.Error("expected HasFunc('myFunc') to be true")
	}
	if p.HasFunc("nonExistent") {
		t.Error("expected HasFunc('nonExistent') to be false")
	}
	if p.HasFunc("myVar") {
		t.Error("expected HasFunc('myVar') to be false, it's a variable")
	}
}

func TestCall(t *testing.T) {
	code := `function add(a, b) { return a + b; }
function noReturn() { }`
	p, err := Parse(code)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Call with valid function
	result, err := p.Call("add", 3, 4)
	if err != nil {
		t.Fatalf("Call failed: %v", err)
	}
	if result.(int64) != 7 {
		t.Errorf("wrong result: expected=7 had=%v", result)
	}

	// Call function that returns undefined
	result, err = p.Call("noReturn")
	if err != nil {
		t.Fatalf("Call noReturn failed: %v", err)
	}
	if result != nil {
		t.Errorf("expected nil result from noReturn, got %v", result)
	}

	// Call non-existent function
	_, err = p.Call("nonExistent")
	if err == nil {
		t.Error("expected error for non-existent function")
	}
}

func TestCallWithStrings(t *testing.T) {
	code := `function concat(a, b) { return a + b; }`
	p, err := Parse(code)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	result, err := p.Call("concat", "hello ", "world")
	if err != nil {
		t.Fatalf("Call failed: %v", err)
	}
	if result.(string) != "hello world" {
		t.Errorf("wrong result: expected='hello world' had=%v", result)
	}
}

func TestSet(t *testing.T) {
	code := `function getVal() { return myVal; }`
	p, err := Parse(code)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	err = p.Set("myVal", 42)
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	result, err := p.Call("getVal")
	if err != nil {
		t.Fatalf("Call failed: %v", err)
	}
	if result.(int64) != 42 {
		t.Errorf("wrong result: expected=42 had=%v", result)
	}
}

func TestMethods(t *testing.T) {
	code := `function alpha() {}
function beta() {}
var gamma = "not a function";`
	p, err := Parse(code)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	methods := p.Methods()
	sort.Strings(methods)

	if len(methods) != 2 {
		t.Fatalf("expected 2 methods, got %d: %v", len(methods), methods)
	}
	if methods[0] != "alpha" || methods[1] != "beta" {
		t.Errorf("wrong methods: %v", methods)
	}
}

func TestObjects(t *testing.T) {
	code := `var str = "hello";
var num = 42;
function fn() {}`
	p, err := Parse(code)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	objs := p.Objects()
	sort.Strings(objs)

	found := map[string]bool{}
	for _, o := range objs {
		found[o] = true
	}
	if !found["str"] {
		t.Error("expected 'str' in objects")
	}
	if !found["num"] {
		t.Error("expected 'num' in objects")
	}
	if found["fn"] {
		t.Error("'fn' should not be in objects, it's a function")
	}
}

func TestGetObject(t *testing.T) {
	code := `var myStr = "test value";`
	p, err := Parse(code)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	val, err := p.GetObject("myStr")
	if err != nil {
		t.Fatalf("GetObject failed: %v", err)
	}
	if val.(string) != "test value" {
		t.Errorf("wrong value: expected='test value' had=%v", val)
	}

	_, err = p.GetObject("nonExistent")
	if err == nil {
		t.Error("expected error for non-existent object")
	}
}

func TestGetTypeObject(t *testing.T) {
	code := `
var myString = "hello";
var myNumber = 42;
var myBool = true;
var myArray = [1, 2, 3];
var myNull = null;
`
	p, err := Parse(code)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	tests := []struct {
		name     string
		expected string
	}{
		{"myString", "StringPrimitive"},
		{"myNumber", "NumberPrimitive"},
		{"myBool", "BooleanPrimitive"},
		{"myArray", "ArrayObject"},
		{"nonExistent", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := p.GetTypeObject(tt.name)
			if got != tt.expected {
				t.Errorf("GetTypeObject(%s): expected=%s had=%s", tt.name, tt.expected, got)
			}
		})
	}
}

func TestIsStringPrimitive(t *testing.T) {
	code := `var s = "hello"; var n = 42;`
	p, err := Parse(code)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	if !p.IsStringPrimitive("s") {
		t.Error("expected IsStringPrimitive('s') to be true")
	}
	if p.IsStringPrimitive("n") {
		t.Error("expected IsStringPrimitive('n') to be false")
	}
}

func TestIsBooleanPrimitive(t *testing.T) {
	code := `var b = true; var n = 42;`
	p, err := Parse(code)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	if !p.IsBooleanPrimitive("b") {
		t.Error("expected IsBooleanPrimitive('b') to be true")
	}
	if p.IsBooleanPrimitive("n") {
		t.Error("expected IsBooleanPrimitive('n') to be false")
	}
}

func TestIsNumberPrimitive(t *testing.T) {
	code := `var n = 42; var s = "str";`
	p, err := Parse(code)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	if !p.IsNumberPrimitive("n") {
		t.Error("expected IsNumberPrimitive('n') to be true")
	}
	if p.IsNumberPrimitive("s") {
		t.Error("expected IsNumberPrimitive('s') to be false")
	}
}

func TestIsArrayObject(t *testing.T) {
	code := `var a = [1,2,3]; var n = 42;`
	p, err := Parse(code)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	if !p.IsArrayObject("a") {
		t.Error("expected IsArrayObject('a') to be true")
	}
	if p.IsArrayObject("n") {
		t.Error("expected IsArrayObject('n') to be false")
	}
}

func TestIsDateObject(t *testing.T) {
	code := `var d = new Date(); var n = 42;`
	p, err := Parse(code)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	if !p.IsDateObject("d") {
		t.Error("expected IsDateObject('d') to be true")
	}
	if p.IsDateObject("n") {
		t.Error("expected IsDateObject('n') to be false")
	}
}

func TestIsRegExpObject(t *testing.T) {
	code := `var r = /test/; var n = 42;`
	p, err := Parse(code)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	if !p.IsRegExpObject("r") {
		t.Error("expected IsRegExpObject('r') to be true")
	}
	if p.IsRegExpObject("n") {
		t.Error("expected IsRegExpObject('n') to be false")
	}
}

func TestIsErrorObject(t *testing.T) {
	code := `var e = new Error("test"); var n = 42;`
	p, err := Parse(code)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	if !p.IsErrorObject("e") {
		t.Error("expected IsErrorObject('e') to be true")
	}
	if p.IsErrorObject("n") {
		t.Error("expected IsErrorObject('n') to be false")
	}
}

func TestCloneFromCode(t *testing.T) {
	code := `function hello() { return "world"; }`
	p, err := Parse(code)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	clone := p.Clone()
	if clone == p {
		t.Error("clone should be a different instance")
	}
	if !clone.HasFunc("hello") {
		t.Error("clone should have the hello function")
	}

	result, err := clone.Call("hello")
	if err != nil {
		t.Fatalf("Call on clone failed: %v", err)
	}
	if result.(string) != "world" {
		t.Errorf("wrong result from clone: expected=world had=%v", result)
	}
}

func TestCloneFromFile(t *testing.T) {
	code := `function greet() { return "hi"; }`
	tmpFile := filepath.Join(os.TempDir(), "clone_test.js")
	if err := os.WriteFile(tmpFile, []byte(code), 0644); err != nil {
		t.Fatalf("cannot create temp file: %v", err)
	}
	defer os.Remove(tmpFile)

	p, err := Load(tmpFile)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	clone := p.Clone()
	if clone.Path != tmpFile {
		t.Errorf("clone path mismatch: expected=%s had=%s", tmpFile, clone.Path)
	}
	if !clone.HasFunc("greet") {
		t.Error("clone should have the greet function")
	}
}

func TestDefines(t *testing.T) {
	// Save and restore original Defines
	origDefines := Defines
	defer func() { Defines = origDefines }()

	Defines = map[string]interface{}{
		"magicNumber": 42,
	}

	code := `function getMagic() { return magicNumber; }`
	p, err := Parse(code)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	result, err := p.Call("getMagic")
	if err != nil {
		t.Fatalf("Call failed: %v", err)
	}
	if result.(int64) != 42 {
		t.Errorf("wrong result: expected=42 had=%v", result)
	}
}

func TestIsStringObject(t *testing.T) {
	code := `var s = new String("hello"); var n = 42;`
	p, err := Parse(code)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	if !p.IsStringObject("s") {
		t.Error("expected IsStringObject('s') to be true")
	}
	if p.IsStringObject("n") {
		t.Error("expected IsStringObject('n') to be false")
	}
}

func TestIsBooleanObject(t *testing.T) {
	// Note: In goja, `new Boolean(true)` exports as a primitive bool,
	// so the BooleanObject className branch is not reachable through
	// standard JS constructors. We verify it returns false for primitives.
	code := `var b = true; var n = 42;`
	p, err := Parse(code)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	if p.IsBooleanObject("b") {
		t.Error("expected IsBooleanObject('b') to be false for primitive bool")
	}
	if p.IsBooleanObject("n") {
		t.Error("expected IsBooleanObject('n') to be false")
	}
}

func TestIsNumberObject(t *testing.T) {
	// Note: In goja, `new Number(42)` exports as a primitive number,
	// so the NumberObject className branch is not reachable through
	// standard JS constructors. We verify it returns false for primitives.
	code := `var n = 42; var s = "str";`
	p, err := Parse(code)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	if p.IsNumberObject("n") {
		t.Error("expected IsNumberObject('n') to be false for primitive number")
	}
	if p.IsNumberObject("s") {
		t.Error("expected IsNumberObject('s') to be false")
	}
}
