package filters

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Matrix86/driplane/data"

	"github.com/xuri/excelize/v2"
)

// createTestXLSX creates a simple .xlsx file with one sheet and returns its path.
func createTestXLSX(t *testing.T, sheetName string, rows [][]string) string {
	t.Helper()
	f := excelize.NewFile()
	idx, err := f.NewSheet(sheetName)
	if err != nil {
		t.Fatalf("failed to create sheet: %s", err)
	}
	f.SetActiveSheet(idx)
	// Remove default "Sheet1" if it differs
	if sheetName != "Sheet1" {
		_ = f.DeleteSheet("Sheet1")
	}

	for i, row := range rows {
		for j, cell := range row {
			cellRef, _ := excelize.CoordinatesToCellName(j+1, i+1)
			if err := f.SetCellValue(sheetName, cellRef, cell); err != nil {
				t.Fatalf("failed to set cell value: %s", err)
			}
		}
	}

	tmp := filepath.Join(t.TempDir(), "test.xlsx")
	if err := f.SaveAs(tmp); err != nil {
		t.Fatalf("failed to save test xlsx: %s", err)
	}
	return tmp
}

func TestNewXLSFilter(t *testing.T) {
	filter, err := NewXLSFilter(map[string]string{
		"filename": "/tmp/{{.main}}.xlsx",
	})
	if err != nil {
		t.Fatalf("constructor returned error: %s", err)
	}

	f, ok := filter.(*XLS)
	if !ok {
		t.Fatal("cannot cast to *XLS")
	}
	if f.filename == nil {
		t.Error("expected filename template to be set")
	}
}

func TestNewXLSFilterTarget(t *testing.T) {
	filter, err := NewXLSFilter(map[string]string{
		"target": "custom",
	})
	if err != nil {
		t.Fatalf("constructor returned error: %s", err)
	}
	f := filter.(*XLS)
	if f.target != "custom" {
		t.Errorf("expected target 'custom', got '%s'", f.target)
	}
}

func TestNewXLSFilterDefaults(t *testing.T) {
	filter, err := NewXLSFilter(map[string]string{})
	if err != nil {
		t.Fatalf("constructor returned error: %s", err)
	}
	f := filter.(*XLS)
	if f.target != "main" {
		t.Errorf("default target should be 'main', got '%s'", f.target)
	}
	if f.filename != nil {
		t.Error("filename should be nil when not set")
	}
}

func TestNewXLSFilterInvalidTemplate(t *testing.T) {
	_, err := NewXLSFilter(map[string]string{
		"filename": "{{.invalid",
	})
	if err == nil {
		t.Error("expected error for invalid filename template")
	}
}

func TestXLSDoFilterWithFilename(t *testing.T) {
	rows := [][]string{
		{"A1", "B1", "C1"},
		{"A2", "B2", "C2"},
	}
	xlsPath := createTestXLSX(t, "Sheet1", rows)

	filter, _ := NewXLSFilter(map[string]string{
		"filename": xlsPath,
	})
	f := filter.(*XLS)
	bus := &FakeBus{}
	f.bus = bus

	msg := data.NewMessage("ignored")
	ok, err := f.DoFilter(msg)
	if err != nil {
		t.Fatalf("DoFilter returned error: %s", err)
	}
	if !ok {
		t.Error("DoFilter should return true")
	}

	if len(bus.Collected) != 2 {
		t.Fatalf("expected 2 propagated messages, got %d", len(bus.Collected))
	}

	expected := []string{"A1,B1,C1", "A2,B2,C2"}
	for i, m := range bus.Collected {
		if m.GetMessage() != expected[i] {
			t.Errorf("row %d: expected '%s', got '%s'", i, expected[i], m.GetMessage())
		}
		extra := m.GetExtra()
		if extra["xls_sheet"] != "Sheet1" {
			t.Errorf("row %d: expected sheet 'Sheet1', got '%v'", i, extra["xls_sheet"])
		}
		if extra["xls_filename"] != xlsPath {
			t.Errorf("row %d: expected filename '%s', got '%v'", i, xlsPath, extra["xls_filename"])
		}
	}
}

func TestXLSDoFilterNonexistentFile(t *testing.T) {
	filter, _ := NewXLSFilter(map[string]string{
		"filename": "/tmp/definitely_does_not_exist_xls_test.xlsx",
	})
	f := filter.(*XLS)

	msg := data.NewMessage("test")
	ok, err := f.DoFilter(msg)
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
	if ok {
		t.Error("DoFilter should return false on error")
	}
}

func TestXLSDoFilterWithBytes(t *testing.T) {
	rows := [][]string{
		{"X1", "Y1"},
		{"X2", "Y2"},
	}
	xlsPath := createTestXLSX(t, "DataSheet", rows)

	content, err := os.ReadFile(xlsPath)
	if err != nil {
		t.Fatalf("failed to read test xlsx: %s", err)
	}

	filter, _ := NewXLSFilter(map[string]string{})
	f := filter.(*XLS)
	bus := &FakeBus{}
	f.bus = bus

	msg := data.NewMessageWithExtra("test", map[string]interface{}{})
	msg.SetTarget("main", content)

	ok, err := f.DoFilter(msg)
	if err != nil {
		t.Fatalf("DoFilter returned error: %s", err)
	}
	if !ok {
		t.Error("DoFilter should return true")
	}

	if len(bus.Collected) != 2 {
		t.Fatalf("expected 2 propagated messages, got %d", len(bus.Collected))
	}

	expected := []string{"X1,Y1", "X2,Y2"}
	for i, m := range bus.Collected {
		if m.GetMessage() != expected[i] {
			t.Errorf("row %d: expected '%s', got '%s'", i, expected[i], m.GetMessage())
		}
		extra := m.GetExtra()
		if extra["xls_sheet"] != "DataSheet" {
			t.Errorf("row %d: expected sheet 'DataSheet', got '%v'", i, extra["xls_sheet"])
		}
	}
}

func TestXLSDoFilterUnsupportedType(t *testing.T) {
	filter, _ := NewXLSFilter(map[string]string{})
	f := filter.(*XLS)

	msg := data.NewMessage("not_bytes")
	ok, err := f.DoFilter(msg)
	if err == nil {
		t.Error("expected error for unsupported data type")
	}
	if ok {
		t.Error("DoFilter should return false on error")
	}
}

func TestXLSDoFilterInvalidBytes(t *testing.T) {
	filter, _ := NewXLSFilter(map[string]string{})
	f := filter.(*XLS)

	msg := data.NewMessageWithExtra("test", map[string]interface{}{})
	msg.SetTarget("main", []byte("this is not a valid xlsx"))

	ok, err := f.DoFilter(msg)
	if err == nil {
		t.Error("expected error for invalid xlsx bytes")
	}
	if ok {
		t.Error("DoFilter should return false on error")
	}
}

func TestXLSOnEvent(t *testing.T) {
	filter, _ := NewXLSFilter(map[string]string{})
	f := filter.(*XLS)
	f.OnEvent(&data.Event{})
}

func TestXLSFilterRegistered(t *testing.T) {
	if _, ok := filterFactories["xlsfilter"]; !ok {
		t.Error("xls filter should be registered as 'xlsfilter'")
	}
}
