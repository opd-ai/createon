// types.go
package createon

import (
	"time"

	"github.com/opd-ai/paywall/wallet"
)

// Creator represents a content creator on the platform
type Creator struct {
	Username    string    `yaml:"username" validate:"required,alphanum"`
	DisplayName string    `yaml:"display_name" validate:"required"`
	Bio         string    `yaml:"bio"`
	AvatarPath  string    `yaml:"avatar_path"`
	BTCAddress  string    `yaml:"btc_address" validate:"required"`
	XMRAddress  string    `yaml:"xmr_address" validate:"required"`
	SocialLinks []string  `yaml:"social_links"`
	Tiers       []Tier    `yaml:"tiers" validate:"required,dive"`
	CreatedAt   time.Time `yaml:"created_at"`
}

// Tier represents a subscription tier with pricing options for both Bitcoin and Monero.
// Each tier defines a level of access to creator content with associated pricing.
type Tier struct {
	ID          string   `yaml:"id" validate:"required"`
	Name        string   `yaml:"name" validate:"required"`
	Description string   `yaml:"description"`
	PriceBTC    float64  `yaml:"price_btc" validate:"required,gt=0"`
	PriceXMR    float64  `yaml:"price_xmr" validate:"required,gt=0"`
	Features    []string `yaml:"features"`
}

// Post represents a content post created by a creator.
// Posts contain markdown content and are associated with a specific tier for access control.
type Post struct {
	ID        string    `yaml:"id" validate:"required"`
	Title     string    `yaml:"title" validate:"required"`
	TierID    string    `yaml:"tier_id" validate:"required"`
	Content   string    `yaml:"content" validate:"required"`
	Tags      []string  `yaml:"tags"`
	CreatedAt time.Time `yaml:"created_at"`
	UpdatedAt time.Time `yaml:"updated_at"`
	Published bool      `yaml:"published"`
}

// Subscription represents a user's active subscription to a creator's tier.
// It tracks payment history and expiration for access control.
type Subscription struct {
	ID          string    `yaml:"id"`
	Email       string    `yaml:"email"`
	CreatorUser string    `yaml:"creator_user"`
	TierID      string    `yaml:"tier_id"`
	ExpiresAt   time.Time `yaml:"expires_at"`
	Payments    []Payment `yaml:"payments"`
}

// Payment represents a cryptocurrency payment for a subscription.
// It supports multiple wallet types (Bitcoin and Monero) with addresses and amounts.
type Payment struct {
	ID        string                        `yaml:"id"`
	Amounts   map[wallet.WalletType]float64 `yaml:"amounts"`
	Addresses map[wallet.WalletType]string  `yaml:"addresses"`
	Status    string                        `yaml:"status"`
	TxID      map[wallet.WalletType]string  `yaml:"tx_id,omitempty"`
	CreatedAt time.Time                     `yaml:"created_at"`
}

// PostFilter represents filtering criteria for listing posts.
// It allows filtering by tier, tags, publication status, and date range.
type PostFilter struct {
	TierID    string
	Tags      []string
	Published *bool
	Since     time.Time
	Until     time.Time
}

// User represents a registered user on the platform.
// Users can subscribe to creators and access tier-restricted content.
type User struct {
	Email        string    `yaml:"email" validate:"required,email"`
	PasswordHash string    `yaml:"password_hash" validate:"required"`
	CreatedAt    time.Time `yaml:"created_at"`
}

// Session represents an active user session for authentication.
// Sessions are token-based and have configurable expiration times.
type Session struct {
	ID        string    `yaml:"id"`
	Email     string    `yaml:"email"`
	Token     string    `yaml:"token"`
	ExpiresAt time.Time `yaml:"expires_at"`
	CreatedAt time.Time `yaml:"created_at"`
}
