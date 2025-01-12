// types.go
package createon

import (
	"time"

	"github.com/opd-ai/paywall/wallet"
)

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

type Tier struct {
	ID          string   `yaml:"id" validate:"required"`
	Name        string   `yaml:"name" validate:"required"`
	Description string   `yaml:"description"`
	PriceBTC    float64  `yaml:"price_btc" validate:"required,gt=0"`
	PriceXMR    float64  `yaml:"price_xmr" validate:"required,gt=0"`
	Features    []string `yaml:"features"`
}

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

type Subscription struct {
	ID          string    `yaml:"id"`
	Email       string    `yaml:"email"`
	CreatorUser string    `yaml:"creator_user"`
	TierID      string    `yaml:"tier_id"`
	ExpiresAt   time.Time `yaml:"expires_at"`
	Payments    []Payment `yaml:"payments"`
}

type Payment struct {
	ID        string                        `yaml:"id"`
	Amounts   map[wallet.WalletType]float64 `yaml:"amounts"`
	Addresses map[wallet.WalletType]string  `yaml:"addresses"`
	Status    string                        `yaml:"status"`
	TxID      map[wallet.WalletType]string  `yaml:"tx_id,omitempty"`
	CreatedAt time.Time                     `yaml:"created_at"`
}

type PostFilter struct {
	TierID    string
	Tags      []string
	Published *bool
	Since     time.Time
	Until     time.Time
}
