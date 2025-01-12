// pkg/subscription/manager.go
package subscription

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"github.com/opd-ai/createon/pkg/files"
	"github.com/opd-ai/paywall/wallet"

	"github.com/google/uuid"
	. "github.com/opd-ai/createon"
	"github.com/opd-ai/paywall"
)

type Manager struct {
	files   *files.Manager
	paywall *paywall.Paywall
	dataDir string
}

func NewManager(files *files.Manager, pw *paywall.Paywall, dataDir string) *Manager {
	return &Manager{
		files:   files,
		paywall: pw,
		dataDir: dataDir,
	}
}

// CreateSubscription initiates a new subscription with multi-currency support
func (m *Manager) CreateSubscription(ctx context.Context, email, creatorUser, tierID string) (*paywall.Payment, error) {
	// Load creator config to get tier details
	var creator Creator
	creatorPath := filepath.Join("creators", creatorUser, "config.yaml")
	if err := m.files.ReadYAML(creatorPath, &creator); err != nil {
		return nil, fmt.Errorf("failed to load creator: %w", err)
	}

	// Find requested tier
	var selectedTier *Tier
	for _, tier := range creator.Tiers {
		if tier.ID == tierID {
			selectedTier = &tier
			break
		}
	}
	if selectedTier == nil {
		return nil, fmt.Errorf("tier not found: %s", tierID)
	}

	// Configure paywall for this subscription with both BTC and XMR
	var err error
	m.paywall, err = paywall.NewPaywall(paywall.Config{
		PriceInBTC:       selectedTier.PriceBTC,
		PriceInXMR:       selectedTier.PriceXMR,
		TestNet:          true, // TODO: Make configurable
		Store:            paywall.NewFileStore(),
		PaymentTimeout:   24 * time.Hour,
		MinConfirmations: 1,
		// XMR configuration
		XMRUser:     m.getXMRConfig().RPCUser,
		XMRPassword: m.getXMRConfig().RPCPassword,
		XMRRPC:      m.getXMRConfig().RPCURL,
	})
	if err != nil {
		return nil, fmt.Errorf(" : %w", err)
	}

	// Generate payment request with both BTC and XMR addresses
	payment, err := m.paywall.CreatePayment()
	if err != nil {
		return nil, fmt.Errorf("failed to create payment: %w", err)
	}

	// Generate subscription ID
	subID := uuid.New().String()

	// Create initial subscription record with multi-currency support
	sub := Subscription{
		ID:          subID,
		Email:       email,
		CreatorUser: creatorUser,
		TierID:      tierID,
		ExpiresAt:   time.Now().AddDate(0, 1, 0), // 1 month subscription
		Payments: []Payment{{
			ID:        payment.ID,
			Amounts:   payment.Amounts,
			Addresses: payment.Addresses,
			Status:    "pending",
			TxID:      make(map[wallet.WalletType]string),
			CreatedAt: time.Now(),
		}},
	}

	// Save subscription
	subPath := filepath.Join("subscriptions", subID+".yaml")
	if err := m.files.WriteYAML(subPath, sub); err != nil {
		return nil, fmt.Errorf("failed to save subscription: %w", err)
	}

	return payment, nil
}

func (m *Manager) getXMRConfig() wallet.MoneroConfig {
	return wallet.MoneroConfig{
		RPCURL:      "http://127.0.0.1:18081", // Default local node
		RPCUser:     "",                       // Default no auth
		RPCPassword: "",                       // Default no auth
	}
}

// ProcessPayment handles payment confirmation using the middleware's store
func (m *Manager) ProcessPayment(ctx context.Context, paymentID string) error {
	// Get payment from paywall's store
	payment, err := m.paywall.Store.GetPayment(paymentID)
	if err != nil {
		return fmt.Errorf("failed to get payment: %w", err)
	}

	if payment == nil {
		return fmt.Errorf("payment not found: %s", paymentID)
	}

	// Check if either BTC or XMR payment is confirmed
	if payment.Status != paywall.StatusConfirmed {
		return fmt.Errorf("payment not confirmed: %s", paymentID)
	}

	// Find and update subscription
	subs, err := m.files.ListFiles("subscriptions")
	if err != nil {
		return fmt.Errorf("failed to list subscriptions: %w", err)
	}

	for _, subFile := range subs {
		var sub Subscription
		if err := m.files.ReadYAML(filepath.Join("subscriptions", subFile), &sub); err != nil {
			continue
		}

		// Find matching payment
		for i, p := range sub.Payments {
			if p.ID == paymentID {
				// Update payment status
				sub.Payments[i].Status = "confirmed"

				// Store transaction IDs for both currencies if available
				if payment.TransactionID != "" {
					// Determine which currency was used
					if _, ok := payment.Amounts[wallet.Bitcoin]; ok {
						sub.Payments[i].TxID[wallet.Bitcoin] = payment.TransactionID
					}
					if _, ok := payment.Amounts[wallet.Monero]; ok {
						sub.Payments[i].TxID[wallet.Monero] = payment.TransactionID
					}
				}

				// Save updated subscription
				if err := m.files.WriteYAML(filepath.Join("subscriptions", sub.ID+".yaml"), sub); err != nil {
					return fmt.Errorf("failed to save subscription: %w", err)
				}
				return nil
			}
		}
	}

	return fmt.Errorf("subscription not found for payment: %s", paymentID)
}

// VerifyAccess checks if a subscriber has access to content
func (m *Manager) VerifyAccess(ctx context.Context, email, creatorUser, requiredTier string) bool {
	// Implementation remains the same as it's payment-method agnostic
	// The payment method doesn't affect access verification
	return m.verifyAccessImpl(email, creatorUser, requiredTier)
}

// GetActiveSubscriptions returns all active subscriptions for a creator
func (m *Manager) GetActiveSubscriptions(ctx context.Context, creatorUser string) ([]Subscription, error) {
	subs, err := m.files.ListFiles("subscriptions")
	if err != nil {
		return nil, fmt.Errorf("failed to list subscriptions: %w", err)
	}

	var active []Subscription
	now := time.Now()

	for _, subFile := range subs {
		var sub Subscription
		if err := m.files.ReadYAML(filepath.Join("subscriptions", subFile), &sub); err != nil {
			continue
		}

		if sub.CreatorUser == creatorUser && now.Before(sub.ExpiresAt) {
			active = append(active, sub)
		}
	}

	return active, nil
}
