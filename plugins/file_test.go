package plugins

import (
	"os"
	"path"
	"testing"
)

func TestFilePackage_Copy(t *testing.T) {
	h := GetFile()

	type Test struct {
		Name           string
		Filename       string
		CreateFile     bool
		ExpectedStatus bool
		ExpectedError  string
	}
	tests := []Test{
		{"FileNotFound", path.Join(os.TempDir(), "notexistentfile"), false, false, "stat /tmp/notexistentfile: no such file or directory"},
		{"NotRegularFile", path.Join(os.TempDir()), false, false, "/tmp is not a regular file"},
		{"CopyOK", path.Join(os.TempDir(), "test1"), true, true, ""},
	}

	for _, v := range tests {
		if v.CreateFile {
			file, err := os.Create(v.Filename)
			if err != nil {
				t.Errorf("%s: cannot create a temporary file", v.Name)
			}
			defer os.Remove(v.Filename)

			if _, err = file.Write([]byte("test")); err != nil {
				t.Errorf("%s: can't write on file", v.Name)
			}
		}

		dst := path.Join(os.TempDir(), "destination")
		had := h.Copy(v.Filename, dst)
		if v.ExpectedStatus != had.Status {
			t.Errorf("%s: wrong status: expected=%#v had=%#v", v.Name, v.ExpectedStatus, had.Status)
		}
		if v.ExpectedStatus == false && v.ExpectedError != had.Error.Error() {
			t.Errorf("%s: wrong result: expected=%#v had=%#v", v.Name, v.ExpectedError, had.Error.Error())
		}

		if v.CreateFile {
			os.Remove(dst)
		}
	}
}

func TestFilePackage_Move(t *testing.T) {
	h := GetFile()

	type Test struct {
		Name           string
		Filename       string
		Destination    string
		CreateFile     bool
		ExpectedStatus bool
		ExpectedError  string
	}
	tests := []Test{
		{"FileNotFound", path.Join(os.TempDir(), "notexistentfile"), "", false, false, "stat /tmp/notexistentfile: no such file or directory"},
		{"NotRegularSrcFile", path.Join(os.TempDir()), "", false, false, "/tmp is not a regular file"},
		{"NotRegularDstFile", path.Join(os.TempDir(), "test1"), os.TempDir(), true, false, "rename /tmp/test1 /tmp: file exists"},
		{"MoveOK", path.Join(os.TempDir(), "test2"), path.Join(os.TempDir(), "newfile"), true, true, ""},
	}

	for _, v := range tests {
		if v.CreateFile {
			file, err := os.Create(v.Filename)
			if err != nil {
				t.Errorf("%s: cannot create a temporary file", v.Name)
			}
			defer os.Remove(v.Filename)

			if _, err = file.Write([]byte("test")); err != nil {
				t.Errorf("%s: can't write on file", v.Name)
			}
		}

		had := h.Move(v.Filename, v.Destination)
		if v.ExpectedStatus != had.Status {
			t.Errorf("%s: wrong status: expected=%#v had=%#v", v.Name, v.ExpectedStatus, had.Status)
		}
		if v.ExpectedStatus == false && v.ExpectedError != had.Error.Error() {
			t.Errorf("%s: wrong result: expected=%#v had=%#v", v.Name, v.ExpectedError, had.Error.Error())
		}

		if v.CreateFile {
			os.Remove(v.Destination)
		}
	}
}

func TestFilePackage_Truncate(t *testing.T) {
	h := GetFile()

	type Test struct {
		Name           string
		Filename       string
		CreateFile     bool
		ExpectedStatus bool
		ExpectedError  string
		ExpectedSize   int64
	}
	tests := []Test{
		{"FileNotFound", path.Join(os.TempDir(), "notexistentfile"), false, false, "stat /tmp/notexistentfile: no such file or directory", 0},
		{"NotRegularFile", path.Join(os.TempDir()), false, false, "/tmp is not a regular file", 0},
		{"TruncateZero", path.Join(os.TempDir(), "test1"), true, true, "", 0},
		{"TruncateTwo", path.Join(os.TempDir(), "test1"), true, true, "", 2},
	}

	for _, v := range tests {
		if v.CreateFile {
			file, err := os.Create(v.Filename)
			if err != nil {
				t.Errorf("%s: cannot create a temporary file", v.Name)
			}
			defer os.Remove(v.Filename)

			if _, err = file.Write([]byte("test")); err != nil {
				t.Errorf("%s: can't write on file", v.Name)
			}
		}

		had := h.Truncate(v.Filename, v.ExpectedSize)
		if v.ExpectedStatus != had.Status {
			t.Errorf("%s: wrong status: expected=%#v had=%#v", v.Name, v.ExpectedStatus, had.Status)
		}
		if v.ExpectedStatus == false && v.ExpectedError != had.Error.Error() {
			t.Errorf("%s: wrong result: expected=%#v had=%#v", v.Name, v.ExpectedError, had.Error.Error())
		}
		if v.ExpectedStatus != false {
			info, err := os.Stat(v.Filename)
			if os.IsNotExist(err) {
				t.Errorf("%s: wrong result: file not found", v.Name)
			}

			if info.Size() != v.ExpectedSize {
				t.Errorf("%s: wrong size: expected=%d had=%d", v.Name, v.ExpectedSize, info.Size())
			}
		}
	}
}

func TestFilePackage_Delete(t *testing.T) {
	h := GetFile()

	type Test struct {
		Name           string
		Filename       string
		CreateFile     bool
		ExpectedStatus bool
		ExpectedError  string
		ExpectedExist  bool
	}
	tests := []Test{
		{"FileNotFound", path.Join(os.TempDir(), "notexistentfile"), false, false, "stat /tmp/notexistentfile: no such file or directory", false},
		{"NotRegularFile", path.Join(os.TempDir()), false, false, "/tmp is not a regular file", false},
		{"FileRemoved", path.Join(os.TempDir(), "test1"), true, true, "", true},
	}

	for _, v := range tests {
		if v.CreateFile {
			file, err := os.Create(v.Filename)
			if err != nil {
				t.Errorf("%s: cannot create a temporary file", v.Name)
			}
			defer os.Remove(v.Filename)

			if _, err = file.Write([]byte("test")); err != nil {
				t.Errorf("%s: can't write on file", v.Name)
			}
		}

		had := h.Delete(v.Filename)
		if v.ExpectedStatus != had.Status {
			t.Errorf("%s: wrong status: expected=%#v had=%#v", v.Name, v.ExpectedStatus, had.Status)
		}
		if v.ExpectedStatus == false && v.ExpectedError != had.Error.Error() {
			t.Errorf("%s: wrong result: expected=%#v had=%#v", v.Name, v.ExpectedError, had.Error.Error())
		}
		if v.ExpectedStatus != false {
			_, err := os.Stat(v.Filename)
			if os.IsExist(err) == v.ExpectedExist {
				t.Errorf("%s: wrong result: file has not been not deleted correctly", v.Name)
			}
		}
	}
}

func TestFilePackage_Exists(t *testing.T) {
	h := GetFile()

	type Test struct {
		Name           string
		Filename       string
		CreateFile     bool
		ExpectedStatus bool
		ExpectedError  string
	}
	tests := []Test{
		{"FileNotFound", path.Join(os.TempDir(), "notexistentfile"), false, false, "stat /tmp/notexistentfile: no such file or directory"},
		{"NotRegularFile", path.Join(os.TempDir()), false, false, "/tmp is not a regular file"},
		{"FileExist", path.Join(os.TempDir(), "test1"), true, true, ""},
	}

	for _, v := range tests {
		if v.CreateFile {
			file, err := os.Create(v.Filename)
			if err != nil {
				t.Errorf("%s: cannot create a temporary file", v.Name)
			}
			defer os.Remove(v.Filename)

			if _, err = file.Write([]byte("test")); err != nil {
				t.Errorf("%s: can't write on file", v.Name)
			}
		}

		had := h.Exists(v.Filename)
		if v.ExpectedStatus != had.Status {
			t.Errorf("%s: wrong status: expected=%#v had=%#v", v.Name, v.ExpectedStatus, had.Status)
		}
		if v.ExpectedStatus == false && v.ExpectedError != had.Error.Error() {
			t.Errorf("%s: wrong result: expected=%#v had=%#v", v.Name, v.ExpectedError, had.Error.Error())
		}
	}
}

func TestFilePackage_AppendString(t *testing.T) {
	h := GetFile()

	type Test struct {
		Name           string
		Filename       string
		CreateFile     bool
		InitFile       bool
		ExpectedStatus bool
		ExpectedError  string
		ExpectedString string
	}
	tests := []Test{
		{"CannotOpenFile", os.TempDir(), false, false, false, "open /tmp: is a directory", ""},
		{"CreateFile", path.Join(os.TempDir(), "create"), false, false, true, "", "append ok"},
		{"AppendFile", path.Join(os.TempDir(), "append"), true, true, true, "", "file ok append ok"},
	}

	for _, v := range tests {
		if v.CreateFile {
			file, err := os.Create(v.Filename)
			if err != nil {
				t.Errorf("%s: cannot create a temporary file", v.Name)
			}
			defer os.Remove(v.Filename)

			if v.InitFile {
				if _, err = file.Write([]byte("file ok ")); err != nil {
					t.Errorf("%s: can't write on file", v.Name)
				}
			}
		}

		had := h.AppendString(v.Filename, "append ok")
		defer os.Remove(v.Filename)
		if v.ExpectedStatus != had.Status {
			t.Errorf("%s: wrong status: expected=%#v had=%#v", v.Name, v.ExpectedStatus, had.Status)
		}
		if v.ExpectedStatus == false && v.ExpectedError != had.Error.Error() {
			t.Errorf("%s: wrong result: expected=%#v had=%#v", v.Name, v.ExpectedError, had.Error.Error())
		}
		if v.ExpectedStatus {
			content, err := os.ReadFile(v.Filename)
			if err != nil {
				t.Errorf("%s: cannot read the file '%s'", v.Name, v.Filename)
			}
			if string(content) != v.ExpectedString {
				t.Errorf("%s: wrong write: expected=%#v had=%#v", v.Name, v.ExpectedString, string(content))
			}
		}

	}
}
