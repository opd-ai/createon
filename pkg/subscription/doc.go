// Package subscription provides subscription lifecycle management and payment processing.
//
// The subscription package handles the creation and verification of user subscriptions
// to creator tiers, including cryptocurrency payment processing through the paywall
// middleware.
//
// # Features
//
//   - Multi-currency support (Bitcoin and Monero)
//   - Tier-based access control with hierarchical permissions
//   - Subscription expiration tracking
//   - Payment status tracking and verification
//
// # Tier Hierarchy
//
// Higher tier indices grant access to lower tiers. For example, a user subscribed to
// tier3 (index 2) can access content from tier1 (index 0) and tier2 (index 1).
//
// # Usage
//
// Create a Manager with file storage and paywall configuration:
//
//	sm := subscription.NewManager(fm, pw, "data", cfg)
//
// Create a new subscription:
//
//	payment, err := sm.CreateSubscription(ctx, "user@example.com", "creator", "tier1")
//
// Verify access to content:
//
//	hasAccess := sm.VerifyAccess(ctx, "user@example.com", "creator", "tier1")
package subscription
