package filters

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Matrix86/driplane/data"
)

func TestNewJsFilterMissingPath(t *testing.T) {
	_, err := NewJsFilter(map[string]string{})
	// Without path and without rules_path/js_path, it should error
	if err == nil {
		t.Errorf("expected error when path is missing")
	}
}

func TestNewJsFilterMissingFunction(t *testing.T) {
	// Create a temp JS file with a DoFilter function
	tmpDir, err := os.MkdirTemp("", "jsfilter_test_*")
	if err != nil {
		t.Fatalf("cannot create temp dir: %s", err)
	}
	defer os.RemoveAll(tmpDir)

	jsContent := `function DoFilter(text, extra, params) { return { "filtered": true, "data": text }; }`
	jsFile := filepath.Join(tmpDir, "test.js")
	os.WriteFile(jsFile, []byte(jsContent), 0644)

	filter, err := NewJsFilter(map[string]string{
		"path": jsFile,
	})
	if err != nil {
		t.Fatalf("constructor returned error: %s", err)
	}

	f, ok := filter.(*Js)
	if !ok {
		t.Fatal("cannot cast to *Js")
	}
	// default function should be DoFilter
	if f.function != "DoFilter" {
		t.Errorf("default function should be 'DoFilter', got '%s'", f.function)
	}
}

func TestNewJsFilterCustomFunction(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "jsfilter_custom_*")
	if err != nil {
		t.Fatalf("cannot create temp dir: %s", err)
	}
	defer os.RemoveAll(tmpDir)

	jsContent := `function CustomFilter(text, extra, params) { return { "filtered": true, "data": text }; }`
	jsFile := filepath.Join(tmpDir, "custom.js")
	os.WriteFile(jsFile, []byte(jsContent), 0644)

	filter, err := NewJsFilter(map[string]string{
		"path":     jsFile,
		"function": "CustomFilter",
	})
	if err != nil {
		t.Fatalf("constructor returned error: %s", err)
	}

	f := filter.(*Js)
	if f.function != "CustomFilter" {
		t.Errorf("expected function 'CustomFilter', got '%s'", f.function)
	}
}

func TestNewJsFilterMissingFunctionInFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "jsfilter_nofunc_*")
	if err != nil {
		t.Fatalf("cannot create temp dir: %s", err)
	}
	defer os.RemoveAll(tmpDir)

	jsContent := `function SomeOtherFunction() { return true; }`
	jsFile := filepath.Join(tmpDir, "nofunc.js")
	os.WriteFile(jsFile, []byte(jsContent), 0644)

	_, err = NewJsFilter(map[string]string{
		"path": jsFile,
	})
	if err == nil {
		t.Errorf("expected error when DoFilter function is missing from JS file")
	}
}

func TestNewJsFilterInvalidFile(t *testing.T) {
	_, err := NewJsFilter(map[string]string{
		"path": "/tmp/nonexistent_js_file_99999.js",
	})
	if err == nil {
		t.Errorf("expected error for nonexistent JS file")
	}
}

func TestNewJsFilterRelativePathWithJsPath(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "jsfilter_relpath_*")
	if err != nil {
		t.Fatalf("cannot create temp dir: %s", err)
	}
	defer os.RemoveAll(tmpDir)

	jsContent := `function DoFilter(text, extra, params) { return { "filtered": true, "data": text }; }`
	jsFile := filepath.Join(tmpDir, "relative.js")
	os.WriteFile(jsFile, []byte(jsContent), 0644)

	filter, err := NewJsFilter(map[string]string{
		"path":            "relative.js",
		"general.js_path": tmpDir,
	})
	if err != nil {
		t.Fatalf("constructor returned error: %s", err)
	}

	f := filter.(*Js)
	if f.filepath != jsFile {
		t.Errorf("expected filepath '%s', got '%s'", jsFile, f.filepath)
	}
}

func TestNewJsFilterRelativePathWithRulesPath(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "jsfilter_rulespath_*")
	if err != nil {
		t.Fatalf("cannot create temp dir: %s", err)
	}
	defer os.RemoveAll(tmpDir)

	jsContent := `function DoFilter(text, extra, params) { return { "filtered": true, "data": text }; }`
	jsFile := filepath.Join(tmpDir, "fromrules.js")
	os.WriteFile(jsFile, []byte(jsContent), 0644)

	filter, err := NewJsFilter(map[string]string{
		"path":               "fromrules.js",
		"general.rules_path": tmpDir,
	})
	if err != nil {
		t.Fatalf("constructor returned error: %s", err)
	}

	f := filter.(*Js)
	expected := filepath.Join(tmpDir, "fromrules.js")
	if f.filepath != expected {
		t.Errorf("expected filepath '%s', got '%s'", expected, f.filepath)
	}
}

func TestJsDoFilterSimple(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "jsfilter_dofilter_*")
	if err != nil {
		t.Fatalf("cannot create temp dir: %s", err)
	}
	defer os.RemoveAll(tmpDir)

	jsContent := `function DoFilter(text, extra, params) {
		return { "filtered": true, "data": "processed: " + text };
	}`
	jsFile := filepath.Join(tmpDir, "dofilter.js")
	os.WriteFile(jsFile, []byte(jsContent), 0644)

	filter, err := NewJsFilter(map[string]string{
		"path": jsFile,
	})
	if err != nil {
		t.Fatalf("constructor returned error: %s", err)
	}

	f := filter.(*Js)
	msg := data.NewMessage("hello")
	ok, err := f.DoFilter(msg)
	if err != nil {
		t.Fatalf("DoFilter returned error: %s", err)
	}
	if !ok {
		t.Errorf("DoFilter should return true when filtered=true")
	}
	if msg.GetMessage() != "processed: hello" {
		t.Errorf("expected 'processed: hello', got '%v'", msg.GetMessage())
	}
}

func TestJsDoFilterReturnsFalse(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "jsfilter_false_*")
	if err != nil {
		t.Fatalf("cannot create temp dir: %s", err)
	}
	defer os.RemoveAll(tmpDir)

	jsContent := `function DoFilter(text, extra, params) {
		return { "filtered": false, "data": text };
	}`
	jsFile := filepath.Join(tmpDir, "filterfalse.js")
	os.WriteFile(jsFile, []byte(jsContent), 0644)

	filter, err := NewJsFilter(map[string]string{
		"path": jsFile,
	})
	if err != nil {
		t.Fatalf("constructor returned error: %s", err)
	}

	f := filter.(*Js)
	msg := data.NewMessage("hello")
	ok, err := f.DoFilter(msg)
	if err != nil {
		t.Fatalf("DoFilter returned error: %s", err)
	}
	if ok {
		t.Errorf("DoFilter should return false when filtered=false")
	}
}

func TestJsDoFilterWithMapData(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "jsfilter_map_*")
	if err != nil {
		t.Fatalf("cannot create temp dir: %s", err)
	}
	defer os.RemoveAll(tmpDir)

	jsContent := `function DoFilter(text, extra, params) {
		return { "filtered": true, "data": {"main": "new_main", "custom_key": "custom_val"} };
	}`
	jsFile := filepath.Join(tmpDir, "mapdata.js")
	os.WriteFile(jsFile, []byte(jsContent), 0644)

	filter, err := NewJsFilter(map[string]string{
		"path": jsFile,
	})
	if err != nil {
		t.Fatalf("constructor returned error: %s", err)
	}

	f := filter.(*Js)
	msg := data.NewMessage("original")
	ok, err := f.DoFilter(msg)
	if err != nil {
		t.Fatalf("DoFilter returned error: %s", err)
	}
	if !ok {
		t.Errorf("DoFilter should return true")
	}
	if msg.GetMessage() != "new_main" {
		t.Errorf("expected main 'new_main', got '%v'", msg.GetMessage())
	}
	extra := msg.GetExtra()
	if v, ok := extra["custom_key"]; !ok || v != "custom_val" {
		t.Errorf("expected custom_key 'custom_val', got '%v'", v)
	}
}

func TestJsOnEvent(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "jsfilter_event_*")
	defer os.RemoveAll(tmpDir)

	jsContent := `function DoFilter(text, extra, params) { return { "filtered": true, "data": text }; }`
	jsFile := filepath.Join(tmpDir, "event.js")
	os.WriteFile(jsFile, []byte(jsContent), 0644)

	filter, _ := NewJsFilter(map[string]string{"path": jsFile})
	f := filter.(*Js)
	f.OnEvent(&data.Event{})
}

func TestJsFilterRegistered(t *testing.T) {
	if _, ok := filterFactories["jsfilter"]; !ok {
		t.Errorf("js filter should be registered as 'jsfilter'")
	}
}
