// Package auth provides user authentication and session management for the Createon platform.
//
// The auth package handles user registration, login, logout, and session validation.
// It stores user credentials and session data using the file-based storage system.
//
// # Security
//
// Passwords are hashed before storage. Sessions use secure random tokens with
// configurable timeout periods. The package provides HTTP middleware for both
// optional authentication (populates user in context if present) and required
// authentication (redirects to login if not authenticated).
//
// # User Management
//
// Users are stored as YAML files indexed by email hash:
//
//	data/users/{email_hash}.yaml
//
// # Session Management
//
// Sessions are stored as YAML files with token-based lookup:
//
//	data/sessions/{session_id}.yaml
//
// # Usage
//
// Create a new auth manager:
//
//	am := auth.NewManager(fileManager, 24*time.Hour)
//
// Register a new user:
//
//	err := am.Register("user@example.com", "password123")
//
// Login and get a session:
//
//	session, err := am.Login("user@example.com", "password123")
//
// Use middleware in HTTP handlers:
//
//	router.Use(am.Middleware)          // Optional auth
//	router.With(am.RequireAuth).Get()  // Required auth
package auth
