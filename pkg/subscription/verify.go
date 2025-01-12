package subscription

import (
	"fmt"
	"path/filepath"
	"time"

	. "github.com/opd-ai/createon"
)

// pkg/subscription/manager.go
// Add this method to the existing file

// verifyAccessImpl checks if a subscriber has valid access to the required tier
func (m *Manager) verifyAccessImpl(email, creatorUser, requiredTier string) bool {
	// List all subscriptions
	subs, err := m.files.ListFiles("subscriptions")
	if err != nil {
		return false
	}

	now := time.Now()

	// Find active subscription
	for _, subFile := range subs {
		var sub Subscription
		if err := m.files.ReadYAML(filepath.Join("subscriptions", subFile), &sub); err != nil {
			continue
		}

		// Check if subscription matches user and creator
		if sub.Email == email && sub.CreatorUser == creatorUser {
			// Check if subscription is expired
			if now.After(sub.ExpiresAt) {
				continue
			}

			// Verify at least one payment is confirmed
			hasConfirmedPayment := false
			for _, payment := range sub.Payments {
				if payment.Status == "confirmed" {
					hasConfirmedPayment = true
					break
				}
			}

			if !hasConfirmedPayment {
				continue
			}

			// Load creator config to check tier hierarchy
			var creator Creator
			creatorPath := filepath.Join("creators", creatorUser, "config.yaml")
			if err := m.files.ReadYAML(creatorPath, &creator); err != nil {
				continue
			}

			// Check if subscription tier includes access to required tier
			return m.tierIncludesAccess(creator.Tiers, sub.TierID, requiredTier)
		}
	}

	return false
}

// tierIncludesAccess checks if subscribedTier includes access to requiredTier
// Higher tier indices indicate more access (e.g., tier3 > tier2 > tier1)
func (m *Manager) tierIncludesAccess(tiers []Tier, subscribedTier, requiredTier string) bool {
	var subTierIndex, reqTierIndex int
	found := 0

	// Find tier indices
	for i, t := range tiers {
		if t.ID == subscribedTier {
			subTierIndex = i
			found++
		}
		if t.ID == requiredTier {
			reqTierIndex = i
			found++
		}
		if found == 2 {
			break
		}
	}

	// If either tier wasn't found, deny access
	if found != 2 {
		return false
	}

	// Higher tier index means more access
	// Example: If user has tier3 (index 2) they can access tier1 (index 0) and tier2 (index 1)
	return subTierIndex >= reqTierIndex
}

// Helper method to check if a subscription is currently valid
func (m *Manager) isSubscriptionValid(sub *Subscription) bool {
	if sub == nil {
		return false
	}

	// Check expiration
	if time.Now().After(sub.ExpiresAt) {
		return false
	}

	// Check for at least one confirmed payment
	for _, payment := range sub.Payments {
		if payment.Status == "confirmed" {
			return true
		}
	}

	return false
}

// Helper method to get subscription by ID
func (m *Manager) getSubscription(subID string) (*Subscription, error) {
	path := filepath.Join("subscriptions", subID+".yaml")
	var sub Subscription
	if err := m.files.ReadYAML(path, &sub); err != nil {
		return nil, fmt.Errorf("failed to load subscription: %w", err)
	}
	return &sub, nil
}
