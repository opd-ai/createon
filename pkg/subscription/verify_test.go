package subscription

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	. "github.com/opd-ai/createon"
	"github.com/opd-ai/createon/pkg/files"
)

func setupTestManager(t *testing.T) (*Manager, string) {
	tmpDir := t.TempDir()
	fm, err := files.NewManager(tmpDir)
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}

	// Create directories
	os.MkdirAll(filepath.Join(tmpDir, "subscriptions"), 0o755)
	os.MkdirAll(filepath.Join(tmpDir, "creators", "testcreator"), 0o755)

	// Create test creator
	creator := Creator{
		Username:    "testcreator",
		DisplayName: "Test Creator",
		Tiers: []Tier{
			{ID: "tier1", Name: "Basic", PriceBTC: 0.001, PriceXMR: 0.01},
			{ID: "tier2", Name: "Premium", PriceBTC: 0.002, PriceXMR: 0.02},
			{ID: "tier3", Name: "VIP", PriceBTC: 0.005, PriceXMR: 0.05},
		},
	}
	fm.WriteYAML(filepath.Join("creators", "testcreator", "config.yaml"), creator)

	m := NewManager(fm, nil, tmpDir, nil)
	return m, tmpDir
}

func TestTierIncludesAccess_SameTier(t *testing.T) {
	m, _ := setupTestManager(t)

	tiers := []Tier{
		{ID: "tier1", Name: "Basic"},
		{ID: "tier2", Name: "Premium"},
		{ID: "tier3", Name: "VIP"},
	}

	// Same tier should have access
	if !m.tierIncludesAccess(tiers, "tier1", "tier1") {
		t.Error("tier1 should have access to tier1")
	}
	if !m.tierIncludesAccess(tiers, "tier2", "tier2") {
		t.Error("tier2 should have access to tier2")
	}
}

func TestTierIncludesAccess_HigherTier(t *testing.T) {
	m, _ := setupTestManager(t)

	tiers := []Tier{
		{ID: "tier1", Name: "Basic"},
		{ID: "tier2", Name: "Premium"},
		{ID: "tier3", Name: "VIP"},
	}

	// Higher tier should have access to lower tiers
	if !m.tierIncludesAccess(tiers, "tier3", "tier1") {
		t.Error("tier3 should have access to tier1")
	}
	if !m.tierIncludesAccess(tiers, "tier3", "tier2") {
		t.Error("tier3 should have access to tier2")
	}
	if !m.tierIncludesAccess(tiers, "tier2", "tier1") {
		t.Error("tier2 should have access to tier1")
	}
}

func TestTierIncludesAccess_LowerTier(t *testing.T) {
	m, _ := setupTestManager(t)

	tiers := []Tier{
		{ID: "tier1", Name: "Basic"},
		{ID: "tier2", Name: "Premium"},
		{ID: "tier3", Name: "VIP"},
	}

	// Lower tier should NOT have access to higher tiers
	if m.tierIncludesAccess(tiers, "tier1", "tier2") {
		t.Error("tier1 should NOT have access to tier2")
	}
	if m.tierIncludesAccess(tiers, "tier1", "tier3") {
		t.Error("tier1 should NOT have access to tier3")
	}
	if m.tierIncludesAccess(tiers, "tier2", "tier3") {
		t.Error("tier2 should NOT have access to tier3")
	}
}

func TestTierIncludesAccess_InvalidTier(t *testing.T) {
	m, _ := setupTestManager(t)

	tiers := []Tier{
		{ID: "tier1", Name: "Basic"},
		{ID: "tier2", Name: "Premium"},
	}

	// Non-existent tier should deny access
	if m.tierIncludesAccess(tiers, "nonexistent", "tier1") {
		t.Error("nonexistent tier should NOT have access")
	}
	if m.tierIncludesAccess(tiers, "tier1", "nonexistent") {
		t.Error("should NOT have access to nonexistent tier")
	}
}

func TestIsSubscriptionValid_Active(t *testing.T) {
	m, _ := setupTestManager(t)

	sub := &Subscription{
		ID:        "sub1",
		Email:     "test@example.com",
		ExpiresAt: time.Now().Add(24 * time.Hour),
		Payments: []Payment{
			{ID: "pay1", Status: "confirmed"},
		},
	}

	if !m.isSubscriptionValid(sub) {
		t.Error("subscription with confirmed payment and future expiry should be valid")
	}
}

func TestIsSubscriptionValid_Expired(t *testing.T) {
	m, _ := setupTestManager(t)

	sub := &Subscription{
		ID:        "sub1",
		Email:     "test@example.com",
		ExpiresAt: time.Now().Add(-24 * time.Hour), // Expired
		Payments: []Payment{
			{ID: "pay1", Status: "confirmed"},
		},
	}

	if m.isSubscriptionValid(sub) {
		t.Error("expired subscription should NOT be valid")
	}
}

func TestIsSubscriptionValid_NoConfirmedPayment(t *testing.T) {
	m, _ := setupTestManager(t)

	sub := &Subscription{
		ID:        "sub1",
		Email:     "test@example.com",
		ExpiresAt: time.Now().Add(24 * time.Hour),
		Payments: []Payment{
			{ID: "pay1", Status: "pending"},
		},
	}

	if m.isSubscriptionValid(sub) {
		t.Error("subscription without confirmed payment should NOT be valid")
	}
}

func TestIsSubscriptionValid_NilSubscription(t *testing.T) {
	m, _ := setupTestManager(t)

	if m.isSubscriptionValid(nil) {
		t.Error("nil subscription should NOT be valid")
	}
}

func TestVerifyAccessImpl_NoSubscription(t *testing.T) {
	m, _ := setupTestManager(t)

	if m.verifyAccessImpl("user@test.com", "testcreator", "tier1") {
		t.Error("user without subscription should NOT have access")
	}
}

func TestVerifyAccessImpl_WithValidSubscription(t *testing.T) {
	m, tmpDir := setupTestManager(t)
	fm, _ := files.NewManager(tmpDir)

	// Create a valid subscription
	sub := Subscription{
		ID:          "sub1",
		Email:       "user@test.com",
		CreatorUser: "testcreator",
		TierID:      "tier2",
		ExpiresAt:   time.Now().Add(24 * time.Hour),
		Payments: []Payment{
			{ID: "pay1", Status: "confirmed"},
		},
	}
	fm.WriteYAML(filepath.Join("subscriptions", "sub1.yaml"), sub)

	// Should have access to tier2 and tier1
	if !m.verifyAccessImpl("user@test.com", "testcreator", "tier2") {
		t.Error("user should have access to subscribed tier")
	}
	if !m.verifyAccessImpl("user@test.com", "testcreator", "tier1") {
		t.Error("user should have access to lower tier")
	}

	// Should NOT have access to tier3
	if m.verifyAccessImpl("user@test.com", "testcreator", "tier3") {
		t.Error("user should NOT have access to higher tier")
	}
}
