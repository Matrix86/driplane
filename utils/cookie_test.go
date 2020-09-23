package utils

import (
	"net/http"
	"os"
	"path"
	"reflect"
	"testing"
	"time"
)

func TestParseCookieFile(t *testing.T) {
	type Test struct {
		Name string
		Filename        string
		CreateFile      bool
		FileContent string
		ExpectedCookies []*http.Cookie
		ExpectedError   string
	}
	tests := []Test{
		{ "FileNotExist", path.Join(os.TempDir(), "notexist"), false, "", []*http.Cookie(nil), "open /tmp/notexist: no such file or directory"},
		{ "WrongJson",path.Join(os.TempDir(), "wrongjson"), true, "wrong", []*http.Cookie(nil), "invalid character 'w' looking for beginning of value"},
		{ "ExportedJson", path.Join(os.TempDir(), "cookie_test_file"), true, "[\n  {\n    \"domain\": \"test.com\",\n    \"expirationDate\": 1627486224,\n    \"hostOnly\": true,\n    \"httpOnly\": false,\n    \"name\": \"testCookie\",\n    \"path\": \"/\",\n    \"sameSite\": \"unspecified\",\n    \"secure\": false,\n    \"session\": false,\n    \"storeId\": \"0\",\n    \"value\": \"testValue\",\n    \"id\": 1\n  }\n]", []*http.Cookie{&http.Cookie{
			Name:       "testCookie",
			Value:      "testValue",
			Path:       "/",
			Domain:     "test.com",
			Expires:    time.Unix(1627486224, 0),
			RawExpires: "",
			MaxAge:     0,
			Secure:     false,
			HttpOnly:   false,
			SameSite:   0,
			Raw:        "",
			Unparsed:   nil,
		}},
		"open /tmp/notexist: no such file or directory"},
	}

	for _, v := range tests {
		if v.CreateFile {
			file, err := os.Create(v.Filename)
			if err != nil {
				t.Errorf("%s: cannot create a temporary file", v.Name)
			}
			defer os.Remove(v.Filename)

			if _, err = file.Write([]byte(v.FileContent)); err != nil {
				t.Errorf("%s: can't write on file", v.Name)
			}
		}

		had, err := ParseCookieFile(v.Filename)
		if v.ExpectedError == "" && err != nil {
			t.Errorf("%s: wrong error: expected=nil had=%#v", v.Name, err)
		} else if err != nil && err.Error() != v.ExpectedError {
			t.Errorf("%s: wrong error: expected=%#v had=%#v", v.Name, v.ExpectedError, err.Error())
		}
		if reflect.DeepEqual(had, v.ExpectedCookies) == false {
			t.Errorf("%s: wrong cookie: expected=%#v had=%#v", v.Name, v.ExpectedCookies, had)
		}
	}
}
