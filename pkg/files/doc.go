// Package files provides thread-safe file system operations for the Createon platform.
//
// The files package implements atomic file writes and per-path locking to ensure
// safe concurrent access to the flat-file data storage. All operations use the
// Manager type which maintains a root directory and manages file locks.
//
// # Features
//
//   - Atomic writes using temporary files and rename operations
//   - Per-path read-write mutex locking for concurrent safety
//   - YAML serialization/deserialization support
//   - Directory listing and existence checking
//
// # Usage
//
// Create a new Manager with a root directory:
//
//	fm, err := files.NewManager("./data")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// Read and write YAML files:
//
//	var creator Creator
//	err := fm.ReadYAML("creators/johndoe/config.yaml", &creator)
//
//	err = fm.WriteYAML("creators/johndoe/config.yaml", creator)
package files
