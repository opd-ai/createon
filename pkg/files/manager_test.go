package files

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
)

func TestNewManager(t *testing.T) {
	tmpDir := t.TempDir()

	m, err := NewManager(tmpDir)
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}

	if m.rootDir == "" {
		t.Error("rootDir should not be empty")
	}
}

func TestWriteAndReadYAML(t *testing.T) {
	tmpDir := t.TempDir()
	m, err := NewManager(tmpDir)
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}

	type TestData struct {
		Name  string `yaml:"name"`
		Value int    `yaml:"value"`
	}

	original := TestData{Name: "test", Value: 42}

	// Write YAML
	if err := m.WriteYAML("test.yaml", original); err != nil {
		t.Fatalf("WriteYAML failed: %v", err)
	}

	// Read YAML
	var result TestData
	if err := m.ReadYAML("test.yaml", &result); err != nil {
		t.Fatalf("ReadYAML failed: %v", err)
	}

	if result.Name != original.Name || result.Value != original.Value {
		t.Errorf("expected %+v, got %+v", original, result)
	}
}

func TestWriteAndReadFile(t *testing.T) {
	tmpDir := t.TempDir()
	m, err := NewManager(tmpDir)
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}

	content := []byte("Hello, World!")

	// Write file
	if err := m.WriteFile("test.txt", content); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	// Read file
	result, err := m.ReadFile("test.txt")
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}

	if string(result) != string(content) {
		t.Errorf("expected %q, got %q", content, result)
	}
}

func TestAtomicWrite(t *testing.T) {
	tmpDir := t.TempDir()
	m, err := NewManager(tmpDir)
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}

	// Write initial content
	if err := m.WriteFile("atomic.txt", []byte("original")); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	// Check for temp files (should be cleaned up)
	files, _ := os.ReadDir(tmpDir)
	for _, f := range files {
		if filepath.Ext(f.Name()) == "" && f.Name() != "atomic.txt" {
			t.Errorf("temp file not cleaned up: %s", f.Name())
		}
	}

	// Verify content is correct
	data, err := m.ReadFile("atomic.txt")
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}

	if string(data) != "original" {
		t.Errorf("expected 'original', got %q", data)
	}
}

func TestExists(t *testing.T) {
	tmpDir := t.TempDir()
	m, err := NewManager(tmpDir)
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}

	// File doesn't exist
	if m.Exists("nonexistent.txt") {
		t.Error("Exists should return false for nonexistent file")
	}

	// Create file
	if err := m.WriteFile("exists.txt", []byte("test")); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	// File exists
	if !m.Exists("exists.txt") {
		t.Error("Exists should return true for existing file")
	}
}

func TestListFiles(t *testing.T) {
	tmpDir := t.TempDir()
	m, err := NewManager(tmpDir)
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}

	// Create some files
	m.WriteFile("file1.txt", []byte("1"))
	m.WriteFile("file2.txt", []byte("2"))

	// Create a subdirectory (should not be listed)
	os.MkdirAll(filepath.Join(tmpDir, "subdir"), 0o755)

	files, err := m.ListFiles("")
	if err != nil {
		t.Fatalf("ListFiles failed: %v", err)
	}

	if len(files) != 2 {
		t.Errorf("expected 2 files, got %d", len(files))
	}
}

func TestConcurrentAccess(t *testing.T) {
	tmpDir := t.TempDir()
	m, err := NewManager(tmpDir)
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}

	type Counter struct {
		Value int `yaml:"value"`
	}

	// Initialize counter
	m.WriteYAML("counter.yaml", Counter{Value: 0})

	// Concurrent writes
	var wg sync.WaitGroup
	numGoroutines := 10

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(val int) {
			defer wg.Done()
			m.WriteYAML("counter.yaml", Counter{Value: val})
		}(i)
	}

	wg.Wait()

	// File should exist and be valid YAML
	var result Counter
	if err := m.ReadYAML("counter.yaml", &result); err != nil {
		t.Fatalf("ReadYAML failed after concurrent writes: %v", err)
	}

	// Value should be one of the written values
	if result.Value < 0 || result.Value >= numGoroutines {
		t.Errorf("unexpected value: %d", result.Value)
	}
}

func TestNestedDirectories(t *testing.T) {
	tmpDir := t.TempDir()
	m, err := NewManager(tmpDir)
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}

	// Write to nested path
	if err := m.WriteFile("a/b/c/deep.txt", []byte("deep")); err != nil {
		t.Fatalf("WriteFile to nested path failed: %v", err)
	}

	// Read back
	data, err := m.ReadFile("a/b/c/deep.txt")
	if err != nil {
		t.Fatalf("ReadFile from nested path failed: %v", err)
	}

	if string(data) != "deep" {
		t.Errorf("expected 'deep', got %q", data)
	}
}
