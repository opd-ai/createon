// interfaces.go
package createon

import (
	"context"

	"github.com/opd-ai/paywall"
)

// FileManager handles file system operations
type FileManager interface {
	ReadYAML(path string, v interface{}) error
	WriteYAML(path string, v interface{}) error
	ReadFile(path string) ([]byte, error)
	WriteFile(path string, data []byte) error
	ListFiles(path string) ([]string, error)
	Exists(path string) bool
}

// ContentManager handles post operations
type ContentManager interface {
	CreatePost(ctx context.Context, username, title string, content []byte, post Post) error
	GetPost(ctx context.Context, username, postID string) (*Post, []byte, error)
	ListPosts(ctx context.Context, username string, filter PostFilter) ([]Post, error)
	UpdatePost(ctx context.Context, username, postID string, content []byte, post Post) error
	DeletePost(ctx context.Context, username, postID string) error
}

// SubscriptionManager handles subscriptions and payments
type SubscriptionManager interface {
	CreateSubscription(ctx context.Context, email, creatorUser, tierID string) (*paywall.Payment, error)
	VerifyAccess(ctx context.Context, email, creatorUser, tierID string) bool
	ProcessPayment(ctx context.Context, paymentID string) error
	GetActiveSubscriptions(ctx context.Context, creatorUser string) ([]Subscription, error)
}

// CreatorManager handles creator operations
type CreatorManager interface {
	CreateCreator(ctx context.Context, creator *Creator) error
	GetCreator(ctx context.Context, username string) (*Creator, error)
	UpdateCreator(ctx context.Context, username string, creator *Creator) error
	ListCreators(ctx context.Context) ([]*Creator, error)
}
