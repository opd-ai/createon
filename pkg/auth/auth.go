// Package auth provides user authentication and session management
package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"path/filepath"
	"time"

	. "github.com/opd-ai/createon"
	"github.com/opd-ai/createon/pkg/files"
)

type contextKey string

const userContextKey contextKey = "user"

// Manager handles user authentication and sessions
type Manager struct {
	files          *files.Manager
	sessionTimeout time.Duration
}

// NewManager creates a new authentication manager
func NewManager(fm *files.Manager, sessionTimeout time.Duration) *Manager {
	if sessionTimeout == 0 {
		sessionTimeout = 24 * time.Hour
	}
	return &Manager{
		files:          fm,
		sessionTimeout: sessionTimeout,
	}
}

// Register creates a new user account
func (m *Manager) Register(email, password string) error {
	// Check if user already exists
	userPath := filepath.Join("users", hashEmail(email)+".yaml")
	if m.files.Exists(userPath) {
		return fmt.Errorf("user already exists")
	}

	// Hash password
	passwordHash := hashPassword(password)

	user := User{
		Email:        email,
		PasswordHash: passwordHash,
		CreatedAt:    time.Now(),
	}

	return m.files.WriteYAML(userPath, user)
}

// Login authenticates a user and creates a session
func (m *Manager) Login(email, password string) (*Session, error) {
	// Load user
	userPath := filepath.Join("users", hashEmail(email)+".yaml")
	var user User
	if err := m.files.ReadYAML(userPath, &user); err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	// Verify password
	if user.PasswordHash != hashPassword(password) {
		return nil, fmt.Errorf("invalid credentials")
	}

	// Create session
	token, err := generateToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate session token: %w", err)
	}

	session := Session{
		ID:        token[:16],
		Email:     email,
		Token:     token,
		ExpiresAt: time.Now().Add(m.sessionTimeout),
		CreatedAt: time.Now(),
	}

	// Save session
	sessionPath := filepath.Join("sessions", session.ID+".yaml")
	if err := m.files.WriteYAML(sessionPath, session); err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	return &session, nil
}

// Logout invalidates a session
func (m *Manager) Logout(token string) error {
	session, err := m.ValidateSession(token)
	if err != nil {
		return err
	}

	// Delete session file
	sessionPath := filepath.Join("sessions", session.ID+".yaml")
	// Mark session as expired instead of deleting
	session.ExpiresAt = time.Now().Add(-1 * time.Hour)
	return m.files.WriteYAML(sessionPath, session)
}

// ValidateSession checks if a session token is valid
func (m *Manager) ValidateSession(token string) (*Session, error) {
	if len(token) < 16 {
		return nil, fmt.Errorf("invalid session token")
	}

	sessionID := token[:16]
	sessionPath := filepath.Join("sessions", sessionID+".yaml")

	var session Session
	if err := m.files.ReadYAML(sessionPath, &session); err != nil {
		return nil, fmt.Errorf("session not found")
	}

	// Check token match
	if session.Token != token {
		return nil, fmt.Errorf("invalid session token")
	}

	// Check expiration
	if time.Now().After(session.ExpiresAt) {
		return nil, fmt.Errorf("session expired")
	}

	return &session, nil
}

// GetUserFromContext retrieves the user from request context
func GetUserFromContext(ctx context.Context) *User {
	user, ok := ctx.Value(userContextKey).(*User)
	if !ok {
		return nil
	}
	return user
}

// SetUserInContext stores the user in request context
func SetUserInContext(ctx context.Context, user *User) context.Context {
	return context.WithValue(ctx, userContextKey, user)
}

// Middleware returns HTTP middleware for authentication
func (m *Manager) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Try to get session from cookie
		cookie, err := r.Cookie("session")
		if err == nil && cookie.Value != "" {
			session, err := m.ValidateSession(cookie.Value)
			if err == nil {
				// Load user and add to context
				userPath := filepath.Join("users", hashEmail(session.Email)+".yaml")
				var user User
				if err := m.files.ReadYAML(userPath, &user); err == nil {
					ctx := SetUserInContext(r.Context(), &user)
					r = r.WithContext(ctx)
				}
			}
		}
		next.ServeHTTP(w, r)
	})
}

// RequireAuth returns middleware that requires authentication
func (m *Manager) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := GetUserFromContext(r.Context())
		if user == nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// hashEmail creates a filesystem-safe hash of an email
func hashEmail(email string) string {
	hash := sha256.Sum256([]byte(email))
	return hex.EncodeToString(hash[:])[:32]
}

// hashPassword creates a hash of a password
func hashPassword(password string) string {
	// Use SHA256 with a static salt for simplicity
	// In production, use bcrypt or argon2
	salted := "createon-salt-" + password
	hash := sha256.Sum256([]byte(salted))
	return hex.EncodeToString(hash[:])
}

// generateToken creates a secure random token
func generateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
