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

// Manager handles subscription lifecycle and payment processing
type Manager struct {
	files   *files.Manager
	paywall *paywall.Paywall
	dataDir string
	cfg     *Config
}

// NewManager creates a new subscription manager with the given configuration
func NewManager(files *files.Manager, pw *paywall.Paywall, dataDir string, cfg *Config) *Manager {
	return &Manager{
		files:   files,
		paywall: pw,
		dataDir: dataDir,
		cfg:     cfg,
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

	// Parse timeout duration from config, default to 24 hours
	timeout := 24 * time.Hour
	if m.cfg != nil && m.cfg.Paywall.Timeout != "" {
		if parsed, err := time.ParseDuration(m.cfg.Paywall.Timeout); err == nil {
			timeout = parsed
		}
	}

	// Determine testnet mode from config
	testNet := true // default to testnet for safety
	if m.cfg != nil {
		testNet = m.cfg.Paywall.TestNet
	}

	// Configure paywall for this subscription with both BTC and XMR
	var err error
	m.paywall, err = paywall.NewPaywall(paywall.Config{
		PriceInBTC:       selectedTier.PriceBTC,
		PriceInXMR:       selectedTier.PriceXMR,
		TestNet:          testNet,
		Store:            paywall.NewFileStore(),
		PaymentTimeout:   timeout,
		MinConfirmations: 1,
		// XMR configuration
		XMRUser:     m.getXMRConfig().RPCUser,
		XMRPassword: m.getXMRConfig().RPCPassword,
		XMRRPC:      m.getXMRConfig().RPCURL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to configure paywall: %w", err)
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
	// Default values
	host := "http://127.0.0.1:18081"
	user := ""
	password := ""

	// Use config values if available
	if m.cfg != nil {
		if m.cfg.Paywall.XMRHost != "" {
			host = m.cfg.Paywall.XMRHost
		}
		if m.cfg.Paywall.XMRUser != "" {
			user = m.cfg.Paywall.XMRUser
		}
		if m.cfg.Paywall.XMRPassword != "" {
			password = m.cfg.Paywall.XMRPassword
		}
	}

	return wallet.MoneroConfig{
		RPCURL:      host,
		RPCUser:     user,
		RPCPassword: password,
	}
}

// findSubscriptionByPaymentID searches for a subscription containing the given payment ID
func (m *Manager) findSubscriptionByPaymentID(paymentID string) (*Subscription, int, error) {
	subs, err := m.files.ListFiles("subscriptions")
	if err != nil {
		return nil, -1, fmt.Errorf("failed to list subscriptions: %w", err)
	}

	for _, subFile := range subs {
		var sub Subscription
		if err := m.files.ReadYAML(filepath.Join("subscriptions", subFile), &sub); err != nil {
			continue
		}

		for i, p := range sub.Payments {
			if p.ID == paymentID {
				return &sub, i, nil
			}
		}
	}

	return nil, -1, fmt.Errorf("subscription not found for payment: %s", paymentID)
}

// updatePaymentStatus updates a subscription's payment status with transaction info
func (m *Manager) updatePaymentStatus(sub *Subscription, paymentIndex int, transactionID string, amounts map[wallet.WalletType]float64) error {
	sub.Payments[paymentIndex].Status = "confirmed"

	if transactionID != "" {
		if sub.Payments[paymentIndex].TxID == nil {
			sub.Payments[paymentIndex].TxID = make(map[wallet.WalletType]string)
		}
		if _, ok := amounts[wallet.Bitcoin]; ok {
			sub.Payments[paymentIndex].TxID[wallet.Bitcoin] = transactionID
		}
		if _, ok := amounts[wallet.Monero]; ok {
			sub.Payments[paymentIndex].TxID[wallet.Monero] = transactionID
		}
	}

	return m.files.WriteYAML(filepath.Join("subscriptions", sub.ID+".yaml"), sub)
}

// ProcessPayment handles payment confirmation using the middleware's store
func (m *Manager) ProcessPayment(ctx context.Context, paymentID string) error {
	payment, err := m.paywall.Store.GetPayment(paymentID)
	if err != nil {
		return fmt.Errorf("failed to get payment: %w", err)
	}
	if payment == nil {
		return fmt.Errorf("payment not found: %s", paymentID)
	}
	if payment.Status != paywall.StatusConfirmed {
		return fmt.Errorf("payment not confirmed: %s", paymentID)
	}

	sub, paymentIndex, err := m.findSubscriptionByPaymentID(paymentID)
	if err != nil {
		return err
	}

	if err := m.updatePaymentStatus(sub, paymentIndex, payment.TransactionID, payment.Amounts); err != nil {
		return fmt.Errorf("failed to save subscription: %w", err)
	}

	return nil
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
