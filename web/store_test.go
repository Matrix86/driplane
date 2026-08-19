package web

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Matrix86/driplane/core"
)

func newTestStore(t *testing.T) (*Store, string) {
	t.Helper()
	dir := t.TempDir()
	for _, sub := range []string{"rules", "templates", "js"} {
		if err := os.MkdirAll(filepath.Join(dir, sub), 0755); err != nil {
			t.Fatalf("mkdir %s: %s", sub, err)
		}
	}

	cfg := &core.Configuration{}
	cfg.SetAll(map[string]string{
		"general.rules_path":     filepath.Join(dir, "rules"),
		"general.templates_path": filepath.Join(dir, "templates"),
		"general.js_path":        filepath.Join(dir, "js"),
	})

	s, err := NewStore(cfg)
	if err != nil {
		t.Fatalf("NewStore: %s", err)
	}
	return s, dir
}

func TestStoreRejectsTraversal(t *testing.T) {
	s, _ := newTestStore(t)

	cases := []struct {
		name string
		file string
		kind Kind // zero value means KindRules
	}{
		{"parent", "../evil.rule", ""},
		{"nested parent", "sub/../../evil.rule", ""},
		{"absolute", "/etc/passwd", ""},
		{"subdirectory", "sub/evil.rule", ""},
		{"empty", "", ""},
		{"dot", ".", ""},
		{"wrong extension", "evil.sh", ""},
		{"no extension", "evil", ""},
		{"windows reserved device name", "CON.rule", ""},
		{"windows reserved device name lower case", "con.rule", ""},
		{"windows reserved device name, different kind", "NUL.js", KindJS},
		{"windows reserved com port name", "COM1.rule", ""},
		{"leading space", " a.rule", ""},
	}

	for _, tc := range cases {
		kind := tc.kind
		if kind == "" {
			kind = KindRules
		}
		t.Run(tc.name, func(t *testing.T) {
			if _, _, err := s.Read(kind, tc.file); err == nil {
				t.Errorf("Read(%q) should fail", tc.file)
			}
			if err := s.Write(kind, tc.file, []byte("x"), time.Time{}); err == nil {
				t.Errorf("Write(%q) should fail", tc.file)
			}
			if err := s.Delete(kind, tc.file); err == nil {
				t.Errorf("Delete(%q) should fail", tc.file)
			}
		})
	}
}

// TestStoreAcceptsInteriorSpace makes sure the leading-space rejection does
// not over-reach: a space in the middle of a name is a legitimate file name
// and must keep working end-to-end.
func TestStoreAcceptsInteriorSpace(t *testing.T) {
	s, _ := newTestStore(t)

	if err := s.Write(KindRules, "a .rule", []byte("A => echo();\n"), time.Time{}); err != nil {
		t.Fatalf("Write(%q) should succeed, got %v", "a .rule", err)
	}
	content, _, err := s.Read(KindRules, "a .rule")
	if err != nil {
		t.Fatalf("Read(%q) should succeed, got %v", "a .rule", err)
	}
	if string(content) != "A => echo();\n" {
		t.Errorf("unexpected content: %q", content)
	}
	if err := s.Delete(KindRules, "a .rule"); err != nil {
		t.Fatalf("Delete(%q) should succeed, got %v", "a .rule", err)
	}
}

func TestStoreRejectsUnknownKind(t *testing.T) {
	s, _ := newTestStore(t)
	if _, err := s.List(Kind("config")); err == nil {
		t.Error("List on an unknown kind should fail")
	}
}

func TestStoreWriteReadListDelete(t *testing.T) {
	s, dir := newTestStore(t)

	if err := s.Write(KindRules, "a.rule", []byte("A => echo();\n"), time.Time{}); err != nil {
		t.Fatalf("Write: %s", err)
	}

	content, _, err := s.Read(KindRules, "a.rule")
	if err != nil {
		t.Fatalf("Read: %s", err)
	}
	if string(content) != "A => echo();\n" {
		t.Errorf("unexpected content: %q", content)
	}

	files, err := s.List(KindRules)
	if err != nil {
		t.Fatalf("List: %s", err)
	}
	if len(files) != 1 || files[0].Name != "a.rule" {
		t.Fatalf("expected [a.rule], got %+v", files)
	}

	// the second write should leave a backup
	if err := s.Write(KindRules, "a.rule", []byte("B => echo();\n"), time.Time{}); err != nil {
		t.Fatalf("second Write: %s", err)
	}
	bak, err := os.ReadFile(filepath.Join(dir, "rules", "a.rule.bak"))
	if err != nil {
		t.Fatalf("reading backup: %s", err)
	}
	if string(bak) != "A => echo();\n" {
		t.Errorf("backup should hold the previous content, got %q", bak)
	}

	if err := s.Delete(KindRules, "a.rule"); err != nil {
		t.Fatalf("Delete: %s", err)
	}
	if _, _, err := s.Read(KindRules, "a.rule"); err == nil {
		t.Error("Read after Delete should fail")
	}
}

func TestStoreWriteConflictOnStaleMtime(t *testing.T) {
	s, _ := newTestStore(t)

	if err := s.Write(KindRules, "a.rule", []byte("A => echo();\n"), time.Time{}); err != nil {
		t.Fatalf("Write: %s", err)
	}
	_, mtime, err := s.Read(KindRules, "a.rule")
	if err != nil {
		t.Fatalf("Read: %s", err)
	}

	// simulate a manual edit made after the file was opened in the editor
	stale := mtime.Add(-time.Minute)
	if err := s.Write(KindRules, "a.rule", []byte("C => echo();\n"), stale); err != ErrConflict {
		t.Errorf("expected ErrConflict, got %v", err)
	}

	// with the current mtime instead it should succeed
	if err := s.Write(KindRules, "a.rule", []byte("C => echo();\n"), mtime); err != nil {
		t.Errorf("write with the current mtime should succeed, got %v", err)
	}
}

// TestStoreKindsReportsOnlyConfiguredPaths proves general.templates_path and
// general.js_path stay genuinely optional: a store built with only
// general.rules_path configured must report only KindRules, not all three,
// since NewStore only registers the kinds whose directory is set.
func TestStoreKindsReportsOnlyConfiguredPaths(t *testing.T) {
	dir := t.TempDir()
	cfg := &core.Configuration{}
	cfg.SetAll(map[string]string{"general.rules_path": dir})

	s, err := NewStore(cfg)
	if err != nil {
		t.Fatalf("NewStore: %s", err)
	}

	kinds := s.Kinds()
	if len(kinds) != 1 || kinds[0] != KindRules {
		t.Errorf("expected [rules], got %v", kinds)
	}
}

func TestStoreRejectsSymlinkEscape(t *testing.T) {
	s, dir := newTestStore(t)

	outside := filepath.Join(dir, "outside.rule")
	if err := os.WriteFile(outside, []byte("X => echo();\n"), 0644); err != nil {
		t.Fatalf("writing outside file: %s", err)
	}
	link := filepath.Join(dir, "rules", "link.rule")
	if err := os.Symlink(outside, link); err != nil {
		t.Skipf("symlink not supported: %s", err)
	}

	if _, _, err := s.Read(KindRules, "link.rule"); err == nil {
		t.Error("reading a symlink pointing outside the root should fail")
	}
}
