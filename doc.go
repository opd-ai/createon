// Package createon provides core types and interfaces for the Createon platform.
//
// Createon is a self-hosted alternative to Patreon that enables creators to monetize
// their content using cryptocurrency payments. It supports both Bitcoin and Monero,
// providing privacy-focused payment options while maintaining a simple flat-file
// architecture for easy deployment.
//
// # Core Types
//
// The package defines the core domain types:
//   - Creator: Represents a content creator with subscription tiers
//   - Tier: Defines subscription levels with BTC and XMR pricing
//   - Post: Represents markdown content posts
//   - Subscription: Tracks user subscriptions to creator tiers
//   - Payment: Represents cryptocurrency payment transactions
//   - User: Represents registered platform users
//   - Session: Manages user authentication sessions
//
// # Interfaces
//
// The package defines interfaces for the main managers:
//   - FileManager: Thread-safe file operations
//   - ContentManager: Post CRUD operations
//   - SubscriptionManager: Subscription and payment handling
//   - CreatorManager: Creator CRUD operations
//
// # Configuration
//
// Use LoadConfig to load application settings from a YAML configuration file.
package createon
