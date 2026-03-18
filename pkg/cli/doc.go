// Package cli provides the command-line interface for the Createon platform.
//
// The cli package implements all Cobra-based commands for managing the Createon
// self-hosted creator platform. It provides commands for server operation, creator
// management, content publishing, subscription handling, and data backup.
//
// # Commands
//
// The following top-level commands are available:
//
//   - server: Start the HTTP server for web access
//   - creator: Manage content creators (add, list)
//   - post: Publish and manage creator posts
//   - sub: Verify and list subscriptions
//   - backup: Create and restore data backups
//
// # Server Command
//
// The server command starts the HTTP server with configurable host, port, and
// data directory:
//
//	createon server --host localhost --port 8080 --data ./data
//
// # Creator Commands
//
// Add a new creator with subscription tiers:
//
//	createon creator add johndoe -n "John Doe" -b "Digital Artist" -t "Basic:0.0001:0.01"
//
// List all creators:
//
//	createon creator list
//
// # Post Commands
//
// Publish a markdown post:
//
//	createon post publish johndoe content.md -t "My Post" -r tier1 --tags "tutorial,golang"
//
// # Subscription Commands
//
// Verify a user's access:
//
//	createon sub verify user@example.com johndoe tier1
//
// List active subscriptions:
//
//	createon sub list johndoe
//
// # Backup Commands
//
// Create a backup:
//
//	createon backup create backup.tar.gz
//
// Restore from backup:
//
//	createon backup restore backup.tar.gz --force
//
// # Configuration
//
// All commands support a --config flag to specify the configuration file path.
// The default is config/server.yaml.
package cli
