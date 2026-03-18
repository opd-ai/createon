# Createon

Createon is a self-hosted alternative to Patreon that enables creators to monetize their content using cryptocurrency payments. It supports both Bitcoin and Monero, providing privacy-focused payment options while maintaining a simple flat-file architecture for easy deployment.

## Features

- **Multi-Currency Support**
  - Bitcoin (BTC) payments
  - Monero (XMR) payments
  - Configurable payment timeouts
  - Automatic payment verification

- **Creator Management**
  - Multiple subscription tiers
  - Markdown content support
  - Custom pricing per tier
  - Profile customization <!-- REVIEW: Fields exist but no CLI/upload support. See GAPS.md. -->

- **Content Management**
  - Markdown-based posts
  - Tier-restricted content
  - Content versioning <!-- REVIEW: Not currently implemented. See GAPS.md for implementation plan. -->
  - Tags and categories <!-- REVIEW: Tags field exists but filtering/display not implemented. See GAPS.md. -->

- **Subscription System**
  - Automated payment processing
  - Tier-based access control
  - Subscription expiration handling
  - Payment status tracking

- **File-Based Storage**
  - No database required
  - Simple backup/restore
  - Portable deployments
  - Thread-safe operations

## Installation

```bash
# Clone the repository
git clone https://github.com/opd-ai/createon
cd createon

# Install dependencies
go mod download

# Build the binary
go build -o createon cmd/createon/main.go
```

## Configuration

Create a `config/server.yaml` file:

```yaml
server:
  host: "localhost"
  port: 8080

data_dir: "./data"
template_dir: "./templates"
assets_dir: "./assets"

paywall:
  testnet: true
  default_btc: 0.0001
  default_xmr: 0.01
  timeout: "24h"
```

## Usage

### Managing Creators

```bash
# Add a new creator
# Note: Tier format is name:price_btc:price_xmr
createon creator add johndoe \
  -n "John Doe" \
  -b "Digital Artist" \
  -t "Basic:0.0001:0.01" \
  -t "Premium:0.0005:0.05"
<!-- REVIEW: Tier parsing has a bug - fmt.Sscanf with %s:%s:%s doesn't parse colons correctly. See pkg/cli/creator.go:62. -->

# List all creators
createon creator list
```

### Publishing Content

```bash
# Publish a new post
createon post publish johndoe content.md \
  -t "My First Post" \
  -r "tier1"
```

### Managing Subscriptions

```bash
# Check subscription status
createon sub verify user@example.com johndoe tier1

# List active subscriptions
createon sub list johndoe
```

### Backing Up Data

```bash
# Create backup
createon backup create backup.tar.gz

# Restore from backup
createon backup restore backup.tar.gz
```

## Directory Structure

```
/data
  /creators/
    /{username}/
      config.yaml      # Creator profile & tiers
      posts/
        {post-id}.md   # Markdown content
        metadata.yaml  # Post metadata
  /subscriptions/
    {sub-id}.yaml     # Subscription details
  /payments/
    {payment-id}.yaml # Payment tracking
```

## Development

### Requirements

- Go 1.21+
- Bitcoin node (optional, for BTC payments)
- Monero node (optional, for XMR payments)

### Building from Source

```bash
go build -o createon cmd/createon/main.go
```

### Running Tests

```bash
go test ./...
```

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/my-feature`)
3. Commit changes (`git commit -am 'Add new feature'`)
4. Push to branch (`git push origin feature/my-feature`)
5. Open a Pull Request

## Security Considerations

- Always run Bitcoin/Monero nodes in a secure environment
- Keep private keys safely stored
- Use HTTPS for production deployments
- Regularly backup your data directory
- Monitor payment verification logs

### Bitcoin Node Compatibility

This project uses btcd v0.24.2 as a dependency. If you are running your own Bitcoin node for payment processing:

- **Minimum btcd version**: v0.24.2 (fixes CVE-2024-38365)
- **Bitcoin Core**: Version 25.0 or later recommended
- Ensure your node is fully synced before processing payments
- Keep your Bitcoin node software updated to the latest stable version

### Monero Node Compatibility

- **Minimum Monero version**: v0.18.4.0 or later
- Use the official Monero daemon or compatible forks
- Configure RPC authentication for production deployments

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Acknowledgments

- [btcd](https://github.com/btcsuite/btcd) - Bitcoin node implementation
- [go-monero-rpc-client](https://github.com/monero-ecosystem/go-monero-rpc-client) - Monero RPC client
- [cobra](https://github.com/spf13/cobra) - CLI framework
- [goldmark](https://github.com/yuin/goldmark) - Markdown parser

## Project Status

This project is in active development. While it's functional, use in production environments should be done with caution and proper testing.