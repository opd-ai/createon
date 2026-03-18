// interfaces.go
package createon

import (
	"context"

	"github.com/opd-ai/paywall"
)

// FileManager handles file system operations for the Createon platform.
// It provides thread-safe methods for reading and writing files and YAML data.
type FileManager interface {
	ReadYAML(path string, v interface{}) error
	WriteYAML(path string, v interface{}) error
	ReadFile(path string) ([]byte, error)
	WriteFile(path string, data []byte) error
	ListFiles(path string) ([]string, error)
	Exists(path string) bool
}

// ContentManager handles post operations for creator content.
// It provides methods for creating, reading, updating, and deleting posts.
// All post operations support versioning - updates create new versions.
type ContentManager interface {
	CreatePost(ctx context.Context, username, title string, content []byte, post Post) error
	GetPost(ctx context.Context, username, postID string) (*Post, []byte, error)
	ListPosts(ctx context.Context, username string, filter PostFilter) ([]Post, error)
	UpdatePost(ctx context.Context, username, postID string, content []byte, post Post) error
	DeletePost(ctx context.Context, username, postID string) error
	// GetPostVersion retrieves a specific version of a post
	GetPostVersion(ctx context.Context, username, postID string, version int) (*Post, []byte, error)
	// ListPostVersions returns all versions of a post with metadata
	ListPostVersions(ctx context.Context, username, postID string) ([]PostVersion, error)
}

// SubscriptionManager handles subscriptions and cryptocurrency payments.
// It manages the subscription lifecycle from creation through payment verification.
type SubscriptionManager interface {
	CreateSubscription(ctx context.Context, email, creatorUser, tierID string) (*paywall.Payment, error)
	VerifyAccess(ctx context.Context, email, creatorUser, tierID string) bool
	ProcessPayment(ctx context.Context, paymentID string) error
	GetActiveSubscriptions(ctx context.Context, creatorUser string) ([]Subscription, error)
}

// CreatorManager handles creator operations on the platform.
// It provides methods for creating, reading, updating, and listing creators.
type CreatorManager interface {
	CreateCreator(ctx context.Context, creator *Creator) error
	GetCreator(ctx context.Context, username string) (*Creator, error)
	UpdateCreator(ctx context.Context, username string, creator *Creator) error
	ListCreators(ctx context.Context) ([]*Creator, error)
}
