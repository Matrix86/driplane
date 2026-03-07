package utils

import (
	"os"
	"path"
	"testing"
)

func TestIsFlagPassed(t *testing.T) {
	if IsFlagPassed("test") {
		t.Errorf("it should return false")
	}

	if IsFlagPassed("test.timeout") == false {
		t.Errorf("it should return true")
	}
}

func TestFileExists(t *testing.T) {
	file, err := os.CreateTemp(os.TempDir(), "prefix")
	if err != nil {
		t.Errorf("cannot create a temporary file")
	}
	defer os.Remove(file.Name())

	if FileExists(file.Name()) == false {
		t.Errorf("it should return true")
	}

	if FileExists("/false/file") == true {
		t.Errorf("it should return false")
	}
}

func TestDirExists(t *testing.T) {
	tmpDir := "/tmp/test_directory"
	os.Mkdir(tmpDir, os.ModeDir)
	defer os.Remove(tmpDir)

	if DirExists(tmpDir) == false {
		t.Errorf("it should return true")
	}

	if DirExists("/false/directory") {
		t.Errorf("it should return false")
	}
}

func TestMD5Sum(t *testing.T) {
	type Test struct {
		Key      interface{}
		Expected string
	}
	tests := []Test{
		Test{"test1", "5a105e8b9d40e1329780d62ea2265d8a"},
		{"", "d41d8cd98f00b204e9800998ecf8427e"},
		{[]byte("test"), "098f6bcd4621d373cade4e832627b4f6"},
	}

	for _, v := range tests {
		had := MD5Sum(v.Key)
		if had != v.Expected {
			t.Errorf("wrong md5: expected=%s had=%s", v.Expected, had)
		}
	}
}

func TestMd5File(t *testing.T) {
	type Test struct {
		Filename       string
		CreateFile     bool
		ExpectedString string
		ExpectedError  string
	}
	tests := []Test{
		{path.Join(os.TempDir(), "testmd5"), true, "702edca0b2181c15d457eacac39de39b", ""},
		{path.Join(os.TempDir(), "notexist"), false, "", "open " + path.Join(os.TempDir(), "notexist") + ": no such file or directory"},
	}

	for _, v := range tests {
		if v.CreateFile {
			file, err := os.Create(v.Filename)
			if err != nil {
				t.Errorf("cannot create a temporary file")
			}
			defer os.Remove(v.Filename)

			text := []byte("This is a test!")
			if _, err = file.Write(text); err != nil {
				t.Error("can't write on file")
			}
		}

		had, err := Md5File(v.Filename)
		if v.ExpectedError == "" && err != nil {
			t.Errorf("wrong error: expected=nil had=%#v", err)
		} else if err != nil && err.Error() != v.ExpectedError {
			t.Errorf("wrong error: expected=%#v had=%#v", v.ExpectedError, err.Error())
		}
		if had != v.ExpectedString {
			t.Errorf("wrong hash: expected=%s had=%s", v.ExpectedString, had)
		}
	}
}

func TestSha1File(t *testing.T) {
	// Test with non-existent file
	_, err := Sha1File(path.Join(os.TempDir(), "notexist_sha1"))
	if err == nil {
		t.Errorf("expected error for non-existent file")
	}

	// Test with valid file
	filename := path.Join(os.TempDir(), "testsha1")
	if err := os.WriteFile(filename, []byte("This is a test!"), 0644); err != nil {
		t.Fatalf("cannot create test file: %v", err)
	}
	defer os.Remove(filename)

	had, err := Sha1File(filename)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	expected := "8b6ccb43dca2040c3cfbcd7bfff0b387d4538c33"
	if had != expected {
		t.Errorf("wrong hash: expected=%s had=%s", expected, had)
	}
}

func TestSha256File(t *testing.T) {
	// Test with non-existent file
	_, err := Sha256File(path.Join(os.TempDir(), "notexist_sha256"))
	if err == nil {
		t.Errorf("expected error for non-existent file")
	}

	// Test with valid file
	filename := path.Join(os.TempDir(), "testsha256")
	if err := os.WriteFile(filename, []byte("This is a test!"), 0644); err != nil {
		t.Fatalf("cannot create test file: %v", err)
	}
	defer os.Remove(filename)

	had, err := Sha256File(filename)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	expected := "54ba1fdce5a89e0d3eee6e4c587497833bc38c3586ff02057dd6451fd2d6b640"
	if had != expected {
		t.Errorf("wrong hash: expected=%s had=%s", expected, had)
	}
}

func TestSha512File(t *testing.T) {
	// Test with non-existent file
	_, err := Sha512File(path.Join(os.TempDir(), "notexist_sha512"))
	if err == nil {
		t.Errorf("expected error for non-existent file")
	}

	// Test with valid file
	filename := path.Join(os.TempDir(), "testsha512")
	if err := os.WriteFile(filename, []byte("This is a test!"), 0644); err != nil {
		t.Fatalf("cannot create test file: %v", err)
	}
	defer os.Remove(filename)

	had, err := Sha512File(filename)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	expected := "d4d6331e89ced845639272bc64ca3ef4e94a57c88431c61aef91f4399e30c6ada32c042f72cedad9cb1c7cfaf04d92e06ad044b557ca16f554f1c6d66b06d0e0"
	if had != expected {
		t.Errorf("wrong hash: expected=%s had=%s", expected, had)
	}
}

func TestFlatStruct(t *testing.T) {
	type SubTest struct {
		Map    map[string]string
		Slice  []string
		Int    int
		Float  float32
		String string
		Bool   bool
	}
	type Test struct {
		Map        map[string]string
		Slice      []string
		Int        int
		Float      float32
		String     string
		Bool       bool
		Struct     SubTest
		ignoreThis string
	}
	i := &Test{
		Map:    map[string]string{"key1": "value1", "key2": "value2"},
		Slice:  []string{"one", "two", "three"},
		Int:    10,
		Float:  1.55,
		String: "This is a test",
		Bool:   true,
		Struct: SubTest{
			Map:    map[string]string{"key1": "value1", "key2": "value2"},
			Slice:  []string{"one", "two", "three"},
			Int:    10,
			Float:  1.55,
			String: "This is a test",
			Bool:   true,
		},
		ignoreThis: "This field will be ignored because it is private",
	}

	res := map[string]string{
		"bool":            "true",
		"struct_int":      "10",
		"struct_float":    "1.550000",
		"struct_bool":     "true",
		"int":             "10",
		"float":           "1.550000",
		"string":          "This is a test",
		"map_key2":        "value2",
		"struct_slice":    "one,two,three",
		"struct_map_key2": "value2",
		"struct_string":   "This is a test",
		"map_key1":        "value1",
		"slice":           "one,two,three",
		"struct_map_key1": "value1",
	}

	flat := FlatStruct(i)
	for k, v := range res {
		if vv, ok := flat[k]; !ok {
			t.Errorf("key not found: %s", k)
		} else if vv != v {
			t.Errorf("bad value key '%s': expected=%#v had=%#v", k, v, vv)
		}
	}
}

func TestFlatStructNil(t *testing.T) {
	flat := FlatStruct(nil)
	if len(flat) != 0 {
		t.Errorf("expected empty map for nil input, got %d entries", len(flat))
	}
}

func TestFlatStructPointerToNil(t *testing.T) {
	var ptr *struct{ Name string }
	flat := FlatStruct(ptr)
	if len(flat) != 0 {
		t.Errorf("expected empty map for nil pointer, got %d entries", len(flat))
	}
}

func TestFlatStructMapWithIntKeys(t *testing.T) {
	// Map with non-string keys should not crash
	m := map[int]string{1: "one", 2: "two"}
	flat := FlatStruct(m)
	// Non-string key maps are skipped, so flat should be empty
	if len(flat) != 0 {
		t.Errorf("expected empty map for int-keyed map, got %d entries", len(flat))
	}
}
