// pkg/files/manager.go
package files

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"gopkg.in/yaml.v3"
)

// Manager implements thread-safe file operations
type Manager struct {
	rootDir string
	mu      sync.RWMutex
	locks   map[string]*sync.RWMutex
	locksMu sync.RWMutex
}

// NewManager creates a new file manager instance
func NewManager(rootDir string) (*Manager, error) {
	absPath, err := filepath.Abs(rootDir)
	if err != nil {
		return nil, fmt.Errorf("invalid root directory: %w", err)
	}

	if err := os.MkdirAll(absPath, 0755); err != nil {
		return nil, fmt.Errorf("could not create root directory: %w", err)
	}

	return &Manager{
		rootDir: absPath,
		locks:   make(map[string]*sync.RWMutex),
	}, nil
}

// getLock returns a mutex for the given path
func (m *Manager) getLock(path string) *sync.RWMutex {
	m.locksMu.Lock()
	defer m.locksMu.Unlock()

	if lock, exists := m.locks[path]; exists {
		return lock
	}

	lock := &sync.RWMutex{}
	m.locks[path] = lock
	return lock
}

// ReadYAML reads and parses a YAML file
func (m *Manager) ReadYAML(path string, v interface{}) error {
	fullPath := filepath.Join(m.rootDir, path)
	lock := m.getLock(fullPath)

	lock.RLock()
	defer lock.RUnlock()

	f, err := os.Open(fullPath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer f.Close()

	decoder := yaml.NewDecoder(f)
	if err := decoder.Decode(v); err != nil {
		return fmt.Errorf("failed to decode YAML: %w", err)
	}

	return nil
}

// WriteYAML writes data to a YAML file atomically
func (m *Manager) WriteYAML(path string, v interface{}) error {
	fullPath := filepath.Join(m.rootDir, path)
	lock := m.getLock(fullPath)

	lock.Lock()
	defer lock.Unlock()

	// Create temporary file
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	tmpFile, err := os.CreateTemp(dir, ".tmp-*")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpPath := tmpFile.Name()

	// Clean up temp file on failure
	success := false
	defer func() {
		if !success {
			os.Remove(tmpPath)
		}
	}()

	// Write data
	encoder := yaml.NewEncoder(tmpFile)
	if err := encoder.Encode(v); err != nil {
		tmpFile.Close()
		return fmt.Errorf("failed to encode YAML: %w", err)
	}

	if err := tmpFile.Sync(); err != nil {
		tmpFile.Close()
		return fmt.Errorf("failed to sync file: %w", err)
	}

	if err := tmpFile.Close(); err != nil {
		return fmt.Errorf("failed to close temp file: %w", err)
	}

	// Atomic rename
	if err := os.Rename(tmpPath, fullPath); err != nil {
		return fmt.Errorf("failed to rename temp file: %w", err)
	}

	success = true
	return nil
}

// ReadFile reads a file's contents
func (m *Manager) ReadFile(path string) ([]byte, error) {
	fullPath := filepath.Join(m.rootDir, path)
	lock := m.getLock(fullPath)

	lock.RLock()
	defer lock.RUnlock()

	data, err := os.ReadFile(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	return data, nil
}

// WriteFile writes data to a file atomically
func (m *Manager) WriteFile(path string, data []byte) error {
	fullPath := filepath.Join(m.rootDir, path)
	lock := m.getLock(fullPath)

	lock.Lock()
	defer lock.Unlock()

	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	tmpFile, err := os.CreateTemp(dir, ".tmp-*")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpPath := tmpFile.Name()

	success := false
	defer func() {
		if !success {
			os.Remove(tmpPath)
		}
	}()

	if _, err := tmpFile.Write(data); err != nil {
		tmpFile.Close()
		return fmt.Errorf("failed to write data: %w", err)
	}

	if err := tmpFile.Sync(); err != nil {
		tmpFile.Close()
		return fmt.Errorf("failed to sync file: %w", err)
	}

	if err := tmpFile.Close(); err != nil {
		return fmt.Errorf("failed to close temp file: %w", err)
	}

	if err := os.Rename(tmpPath, fullPath); err != nil {
		return fmt.Errorf("failed to rename temp file: %w", err)
	}

	success = true
	return nil
}

// ListFiles returns a list of files in a directory
func (m *Manager) ListFiles(path string) ([]string, error) {
	fullPath := filepath.Join(m.rootDir, path)
	lock := m.getLock(fullPath)

	lock.RLock()
	defer lock.RUnlock()

	entries, err := os.ReadDir(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	var files []string
	for _, entry := range entries {
		if !entry.IsDir() {
			files = append(files, entry.Name())
		}
	}

	return files, nil
}

// Exists checks if a file or directory exists
func (m *Manager) Exists(path string) bool {
	fullPath := filepath.Join(m.rootDir, path)
	lock := m.getLock(fullPath)

	lock.RLock()
	defer lock.RUnlock()

	_, err := os.Stat(fullPath)
	return err == nil
}

// Cleanup removes expired locks
func (m *Manager) Cleanup() {
	m.locksMu.Lock()
	defer m.locksMu.Unlock()

	// In a production system, we might want to implement
	// cleanup of unused locks to prevent memory leaks
	m.locks = make(map[string]*sync.RWMutex)
}
