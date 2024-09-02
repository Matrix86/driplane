package utils

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"
)

// IsFlagPassed returns true if the flag is present on the command line
func IsFlagPassed(name string) bool {
	found := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == name {
			found = true
		}
	})
	return found
}

// FileExists returns true if the file is present on disk
func FileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

// DirExists returns true if the directory is present on disk
func DirExists(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return info.IsDir()
}

// MD5Sum calculate the MD5 Hash of a string or []byte
func MD5Sum(content interface{}) string {
	var hash [16]byte
	if v, ok := content.([]byte); ok {
		hash = md5.Sum(v)
	} else if v, ok := content.(string); ok {
		hash = md5.Sum([]byte(v))
	}
	return hex.EncodeToString(hash[:])
}

// Md5File calculate the MD5 Hash of a file
func Md5File(filename string) (string, error) {
	f, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := md5.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)[:]), nil
}

// Sha1File calculate the SHA1 Hash of a file
func Sha1File(filename string) (string, error) {
	f, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha1.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)[:]), nil
}

// Sha256File calculate the SHA256 Hash of a file
func Sha256File(filename string) (string, error) {
	f, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)[:]), nil
}

// Sha512File calculate the SHA512 Hash of a file
func Sha512File(filename string) (string, error) {
	f, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha512.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)[:]), nil
}

// FlatStruct tries to flat a struct object in a map[string]string
func FlatStruct(s interface{}) map[string]string {
	flatten := make(map[string]string)
	if s != nil {
		flatType(s, "", flatten)
	}
	return flatten
}

func flatType(s interface{}, prefix string, flatten map[string]string) {
	v := reflect.ValueOf(s)
	kind := v.Kind()
	if kind == reflect.Ptr || kind == reflect.Interface {
		v = reflect.Indirect(v)
		kind = v.Kind()
		if kind == reflect.Invalid {
			return
		}
	}
	t := v.Type()

	switch kind {
	case reflect.Map:
		if t.Key().Kind() != reflect.String {
			break
		}
		for _, childKey := range v.MapKeys() {
			childValue := v.MapIndex(childKey)
			key := childKey.String()
			if prefix != "" {
				key = fmt.Sprintf("%s_%s", prefix, key)
			}
			flatType(childValue.Interface(), key, flatten)
		}
	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			if v.Field(i).CanInterface() {
				childValue := v.Field(i)
				childKey := t.Field(i).Name
				key := childKey
				if prefix != "" {
					key = fmt.Sprintf("%s_%s", prefix, key)
				}
				flatType(childValue.Interface(), key, flatten)
			}
		}
	case reflect.Array, reflect.Slice:
		slice := reflect.ValueOf(s)
		res := make([]string, 0)
		for i := 0; i < slice.Len(); i++ {
			///res = fmt.Sprintf("%s,%s", res, slice.Index(i))
			res = append(res, fmt.Sprintf("%v", slice.Index(i)))
		}
		flatten[strings.ToLower(prefix)] = strings.Join(res, ",")
	case reflect.String:
		flatten[strings.ToLower(prefix)] = s.(string)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		flatten[strings.ToLower(prefix)] = fmt.Sprintf("%d", s)
	case reflect.Bool:
		flatten[strings.ToLower(prefix)] = fmt.Sprintf("%t", s)
	case reflect.Float32, reflect.Float64:
		flatten[strings.ToLower(prefix)] = fmt.Sprintf("%f", s)
	default:
		flatten[strings.ToLower(prefix)] = fmt.Sprintf("%v", s)
	}
}
