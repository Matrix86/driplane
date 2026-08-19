package web

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/Matrix86/driplane/core"
)

// Kind identifies one of the editable file collections
type Kind string

// The editable file collections
const (
	KindRules     Kind = "rules"
	KindTemplates Kind = "templates"
	KindJS        Kind = "js"
)

// Store errors
var (
	ErrInvalidKind = errors.New("unknown file kind")
	ErrInvalidPath = errors.New("invalid file name")
	ErrInvalidExt  = errors.New("file extension not allowed")
	ErrNotFound    = errors.New("file not found")
	ErrConflict    = errors.New("file has been modified on disk")
)

// FileInfo describes an editable file
type FileInfo struct {
	Name    string    `json:"name"`
	Size    int64     `json:"size"`
	ModTime time.Time `json:"mtime"`
}

// Store gives safe access to the rules, templates and js directories
type Store struct {
	roots map[Kind]string
	exts  map[Kind][]string
}

// NewStore builds a Store from the driplane configuration. Only the kinds
// whose directory is configured are exposed.
func NewStore(cfg *core.Configuration) (*Store, error) {
	s := &Store{
		roots: make(map[Kind]string),
		exts: map[Kind][]string{
			KindRules:     {".rule"},
			KindTemplates: {".txt", ".fmt", ".tpl"},
			KindJS:        {".js"},
		},
	}

	paths := map[Kind]string{
		KindRules:     cfg.Get("general.rules_path"),
		KindTemplates: cfg.Get("general.templates_path"),
		KindJS:        cfg.Get("general.js_path"),
	}

	for kind, path := range paths {
		if path == "" {
			continue
		}
		abs, err := filepath.Abs(path)
		if err != nil {
			return nil, fmt.Errorf("resolving %s path: %s", kind, err)
		}
		s.roots[kind] = abs
	}

	if _, ok := s.roots[KindRules]; !ok {
		return nil, fmt.Errorf("general.rules_path is required")
	}
	return s, nil
}

// Root returns the absolute directory backing a kind
func (s *Store) Root(kind Kind) (string, error) {
	root, ok := s.roots[kind]
	if !ok {
		return "", ErrInvalidKind
	}
	return root, nil
}

// Kinds returns the configured kinds, sorted
func (s *Store) Kinds() []Kind {
	out := make([]Kind, 0, len(s.roots))
	for k := range s.roots {
		out = append(out, k)
	}
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}

// resolve validates a flat file name and returns its absolute path inside the
// kind root. Sub-directories, absolute paths, traversal and symlinks pointing
// outside the root are all rejected.
func (s *Store) resolve(kind Kind, name string) (string, error) {
	root, err := s.Root(kind)
	if err != nil {
		return "", err
	}

	if name == "" || filepath.IsAbs(name) || strings.ContainsRune(name, os.PathSeparator) {
		return "", ErrInvalidPath
	}
	if strings.ContainsRune(name, '/') || strings.ContainsRune(name, '\\') {
		return "", ErrInvalidPath
	}

	clean := filepath.Clean(name)
	if clean != name || clean == "." || clean == ".." || strings.HasPrefix(clean, ".") {
		return "", ErrInvalidPath
	}

	// A leading (or otherwise embedded-but-boundary) space survives
	// filepath.Clean unchanged, so " a.rule" and "a.rule" would both exist as
	// distinct, visually indistinguishable files. Reject it outright; interior
	// spaces such as "a b.rule" are unaffected and remain legitimate.
	if strings.TrimSpace(clean) != clean {
		return "", ErrInvalidPath
	}

	// Windows reserves these device names regardless of extension: opening
	// "CON.rule" or "con.rule" for writing addresses the console device, not
	// a file, so on Windows the write silently fails or hangs instead of
	// creating a rule. They are refused unconditionally (not just on
	// runtime.GOOS == "windows") because driplane releases for multiple
	// platforms and a rule store that behaves differently depending on where
	// it happens to run is worse than one that is uniformly strict; nobody
	// legitimately needs a rule called AUX.rule.
	base := strings.TrimSuffix(clean, filepath.Ext(clean))
	switch strings.ToUpper(base) {
	case "CON", "PRN", "AUX", "NUL",
		"COM1", "COM2", "COM3", "COM4", "COM5", "COM6", "COM7", "COM8", "COM9",
		"LPT1", "LPT2", "LPT3", "LPT4", "LPT5", "LPT6", "LPT7", "LPT8", "LPT9":
		return "", ErrInvalidPath
	}

	allowed := false
	for _, ext := range s.exts[kind] {
		if strings.EqualFold(filepath.Ext(clean), ext) {
			allowed = true
			break
		}
	}
	if !allowed {
		return "", ErrInvalidExt
	}

	full := filepath.Join(root, clean)

	// if the file already exists, make sure it does not escape the root
	// through a symlink
	if _, err := os.Lstat(full); err == nil {
		realPath, err := filepath.EvalSymlinks(full)
		if err != nil {
			return "", ErrInvalidPath
		}
		realRoot, err := filepath.EvalSymlinks(root)
		if err != nil {
			return "", fmt.Errorf("resolving root: %s", err)
		}
		if realPath != filepath.Join(realRoot, clean) {
			return "", ErrInvalidPath
		}
	}

	return full, nil
}

// List returns the files of a kind, sorted by name
func (s *Store) List(kind Kind) ([]FileInfo, error) {
	root, err := s.Root(kind)
	if err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(root)
	if err != nil {
		return nil, err
	}

	out := make([]FileInfo, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if _, err := s.resolve(kind, e.Name()); err != nil {
			continue // hidden, backup or foreign extension
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		out = append(out, FileInfo{Name: e.Name(), Size: info.Size(), ModTime: info.ModTime().UTC()})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out, nil
}

// Read returns the content of a file and its modification time
func (s *Store) Read(kind Kind, name string) ([]byte, time.Time, error) {
	full, err := s.resolve(kind, name)
	if err != nil {
		return nil, time.Time{}, err
	}

	info, err := os.Stat(full)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, time.Time{}, ErrNotFound
		}
		return nil, time.Time{}, err
	}

	content, err := os.ReadFile(full)
	if err != nil {
		return nil, time.Time{}, err
	}
	return content, info.ModTime().UTC(), nil
}

// Write saves a file atomically, keeping the previous content in a .bak file.
// When ifUnmodifiedSince is not zero and the file on disk is newer, the write
// is refused with ErrConflict.
func (s *Store) Write(kind Kind, name string, content []byte, ifUnmodifiedSince time.Time) error {
	full, err := s.resolve(kind, name)
	if err != nil {
		return err
	}

	if info, err := os.Stat(full); err == nil {
		if !ifUnmodifiedSince.IsZero() && info.ModTime().UTC().After(ifUnmodifiedSince.UTC()) {
			return ErrConflict
		}
		previous, err := os.ReadFile(full)
		if err == nil {
			if err := os.WriteFile(full+".bak", previous, 0644); err != nil {
				return fmt.Errorf("writing backup: %s", err)
			}
		}
	} else if !os.IsNotExist(err) {
		return err
	}

	tmp, err := os.CreateTemp(filepath.Dir(full), ".driplane-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)

	if _, err := tmp.Write(content); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Chmod(tmpName, 0644); err != nil {
		return err
	}
	return os.Rename(tmpName, full)
}

// Create creates a new file, failing with ErrConflict if a file by that name
// already exists. The existence check and the creation are a single
// filesystem operation via O_EXCL, so unlike a preliminary Read followed by
// Write, concurrent callers racing to create the same name cannot both
// succeed: exactly one open wins, every other one observes os.ErrExist.
func (s *Store) Create(kind Kind, name string, content []byte) error {
	full, err := s.resolve(kind, name)
	if err != nil {
		return err
	}

	f, err := os.OpenFile(full, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
	if err != nil {
		if errors.Is(err, os.ErrExist) {
			return ErrConflict
		}
		return err
	}

	if _, err := f.Write(content); err != nil {
		f.Close()
		os.Remove(full) // don't leave a half-written file behind
		return err
	}
	if err := f.Close(); err != nil {
		os.Remove(full)
		return err
	}
	return nil
}

// Delete removes a file
func (s *Store) Delete(kind Kind, name string) error {
	full, err := s.resolve(kind, name)
	if err != nil {
		return err
	}
	if err := os.Remove(full); err != nil {
		if os.IsNotExist(err) {
			return ErrNotFound
		}
		return err
	}
	return nil
}
