# Goal-Achievement Assessment

## Project Context

- **What it claims to do**: Createon is a self-hosted Patreon alternative that enables creators to monetize content using cryptocurrency (Bitcoin and Monero), with flat-file storage, tiered subscriptions, markdown content support, and automated payment verification.

- **Target audience**: Privacy-conscious creators who want independence from mainstream platforms, and supporters who prefer cryptocurrency payments. Technical users comfortable with self-hosting.

- **Architecture**:
  | Package | Role |
  |---------|------|
  | `createon` (root) | Core types (Creator, Tier, Post, Subscription, Payment), config loading, interfaces |
  | `pkg/cli` | Cobra CLI commands (server, creator management) |
  | `pkg/files` | Thread-safe flat-file storage manager with atomic writes |
  | `pkg/subscription` | Subscription lifecycle and payment verification |
  | `pkg/templates` | HTML template rendering with goldmark Markdown support |

- **Existing CI/quality gates**: None. No GitHub Actions workflows, no `.gitlab-ci.yml`. Makefile only has `build`, `fmt`, and `doc` targets.

---

## Goal-Achievement Summary

| Stated Goal | Status | Evidence | Gap Description |
|-------------|--------|----------|-----------------|
| **Bitcoin (BTC) payments** | ⚠️ Partial | `CreateSubscription` generates BTC addresses via `opd-ai/paywall`; no BTC-specific RPC config exposed | No configurable BTC node connection; testnet hardcoded |
| **Monero (XMR) payments** | ⚠️ Partial | `getXMRConfig()` returns hardcoded localhost:18081; paywall dependency supports XMR | No runtime config for XMR RPC credentials |
| **Configurable payment timeouts** | ⚠️ Partial | `Config.Paywall.Timeout` exists but unused; hardcoded `24 * time.Hour` in `CreateSubscription` | Timeout config not wired through |
| **Automatic payment verification** | ✅ Achieved | `ProcessPayment` checks paywall store status, updates subscription on confirmation | Works via paywall middleware |
| **Multiple subscription tiers** | ✅ Achieved | `Creator.Tiers` slice with BTC/XMR pricing per tier; CLI supports `-t` flag | |
| **Markdown content support** | ✅ Achieved | `goldmark` with GFM, typographer, auto heading IDs; `RenderMarkdown()` | |
| **Custom pricing per tier** | ✅ Achieved | `Tier.PriceBTC`, `Tier.PriceXMR` floats per tier | |
| **Profile customization** | ⚠️ Partial | `Creator` struct has `DisplayName`, `Bio`, `AvatarPath`, `SocialLinks` | No upload endpoint for avatars; social links not rendered |
| **Markdown-based posts** | ✅ Achieved | Posts stored as `{post-id}.md` files; rendered via templates | |
| **Tier-restricted content** | ✅ Achieved | `VerifyAccess` checks subscription tier hierarchy before serving posts | |
| **Content versioning** | ❌ Missing | README claims it; no implementation found | No version history or diff capability |
| **Tags and categories** | ⚠️ Partial | `Post.Tags` field exists; `PostFilter.Tags` defined | No CLI/API to filter by tags; not rendered in UI |
| **Automated payment processing** | ✅ Achieved | Paywall middleware integration with file-based store | |
| **Tier-based access control** | ✅ Achieved | `tierIncludesAccess()` implements tier hierarchy (higher tier ≥ lower tiers) | |
| **Subscription expiration handling** | ✅ Achieved | `ExpiresAt` field checked in `verifyAccessImpl` | |
| **Payment status tracking** | ✅ Achieved | `Payment.Status` persisted in YAML; updated on confirmation | |
| **No database required** | ✅ Achieved | All data in YAML files under `data/` directory | |
| **Simple backup/restore** | ❌ Missing | README shows `createon backup create/restore` commands; CLI commands not implemented | |
| **Portable deployments** | ✅ Achieved | Flat-file architecture inherently portable | |
| **Thread-safe operations** | ✅ Achieved | `files.Manager` uses per-path RWMutex; atomic writes via temp file + rename | |
| **CLI creator management** | ✅ Achieved | `creator add`, `creator list` implemented in `pkg/cli/creator.go` | |
| **CLI post publishing** | ❌ Missing | README shows `createon post publish`; not implemented | |
| **CLI subscription management** | ❌ Missing | README shows `createon sub verify/list`; not implemented | |
| **Web server** | ⚠️ Partial | Server starts, routes defined; but `main.go` is empty (cannot build) | **Critical**: Entry point missing |
| **User session management** | ❌ Missing | TODO comments at lines 249, 274 in `server.go`; hardcoded `user@example.com` | No auth/session system |

**Overall: 12/23 goals fully achieved (52%)**

---

## Critical Finding: Project Cannot Build

The entry point `cmd/createon/main.go` is **empty** (contains only a newline). This means:
- `go build ./cmd/createon` fails
- The CLI documented in README cannot be run
- All CLI commands exist but are unreachable

---

## Roadmap

### Priority 1: Make the Project Buildable (Critical)

The project cannot be used at all without a working entry point.

- [ ] **Create `cmd/createon/main.go`** with package declaration and main function calling `cli.Execute()`
  ```go
  package main
  
  import (
      "log"
      "github.com/opd-ai/createon/pkg/cli"
  )
  
  func main() {
      if err := cli.Execute(); err != nil {
          log.Fatal(err)
      }
  }
  ```
- [ ] **Validation**: `go build -o createon ./cmd/createon/main.go` succeeds; `./createon --help` shows subcommands

---

### Priority 2: Implement Documented CLI Commands

README documents commands that don't exist, creating a trust gap with users.

- [ ] **Implement `post` subcommand** (`pkg/cli/post.go`)
  - `post publish [username] [file.md] -t [title] -r [tier]`
  - Should create `{post-id}.md` and `metadata.yaml` in `data/creators/{username}/posts/`
- [ ] **Implement `sub` subcommand** (`pkg/cli/subscription.go`)
  - `sub verify [email] [creator] [tier]` — calls `VerifyAccess`
  - `sub list [creator]` — calls `GetActiveSubscriptions`
- [ ] **Implement `backup` subcommand** (`pkg/cli/backup.go`)
  - `backup create [file.tar.gz]` — tar/gzip `data/` directory
  - `backup restore [file.tar.gz]` — extract to `data/`
- [ ] **Validation**: Each command runs without error; documented examples from README work

---

### Priority 3: User Authentication and Sessions

Content access checks currently use hardcoded `user@example.com`. No login/logout exists.

- [ ] **Design session system** (cookie-based or JWT)
- [ ] **Implement `/login` and `/logout` routes** referenced in `base.html`
- [ ] **Implement `/dashboard` route** referenced in `base.html`
- [ ] **Wire user email from session** into `handleViewPost` and `handleSubscribe`
- [ ] **Validation**: User can log in, subscribe, and access tier-restricted content

---

### Priority 4: Wire Configuration Through

Config fields exist but are ignored at runtime.

- [ ] **Use `Config.Paywall.Timeout`** instead of hardcoded `24 * time.Hour` in `CreateSubscription`
- [ ] **Make XMR RPC configurable** — add `XMRHost`, `XMRUser`, `XMRPassword` to `Config.Paywall`; use in `getXMRConfig()`
- [ ] **Make BTC RPC configurable** — similar pattern for Bitcoin node connection
- [ ] **Remove hardcoded `TestNet: true`** — read from config
- [ ] **Validation**: Changing `config/server.yaml` values affects runtime behavior

---

### Priority 5: Implement Content Versioning

README claims "content versioning" but no implementation exists.

- [ ] **Design versioning approach** (options: git-style history per post, or simple `{post-id}_v{n}.md` naming)
- [ ] **Implement version storage** in posts directory
- [ ] **Add `post history` CLI command** to list versions
- [ ] **Validation**: Editing a post creates a new version; old versions are retrievable

---

### Priority 6: Complete Tags/Categories Feature

Post model has tags but they're never used.

- [ ] **Render tags in `post.html` template**
- [ ] **Add tag filter to CLI** (`post list --tag=art`)
- [ ] **Add `/c/{username}/tags/{tag}` route** to filter posts by tag
- [ ] **Validation**: Creating post with tags → tags visible on page → filtering works

---

### Priority 7: Profile Customization Completion

Avatar and social links fields exist but aren't functional.

- [ ] **Add avatar upload** via CLI or web form
- [ ] **Render `Creator.SocialLinks`** in profile template
- [ ] **Serve avatar files** from `data/creators/{username}/avatar.*`
- [ ] **Validation**: Creator profile shows avatar and clickable social links

---

### Priority 8: Add Basic CI/Testing

No tests exist; no CI runs.

- [ ] **Add unit tests** for `pkg/files/Manager` (atomic writes, locking)
- [ ] **Add unit tests** for `pkg/subscription` (tier hierarchy, expiration)
- [ ] **Create `.github/workflows/ci.yml`** running `go test ./...` and `go vet ./...`
- [ ] **Validation**: CI passes on push; `go test ./...` has >0 test functions

---

### Priority 9: Code Quality Improvements (Lower Impact)

Based on `go-stats-generator` findings:

- [ ] **Extract duplicated atomic write logic** in `pkg/files/manager.go:76-100` and `145-167` into shared helper
- [ ] **Add GoDoc comments** to exported functions `LoadConfig`, `Execute`, `NewManager`
- [ ] **Reduce `ProcessPayment` complexity** (cyclomatic: 13, nesting: 5) by extracting subscription-finding loop
- [ ] **Validation**: Re-run `go-stats-generator`; no function >15 complexity; doc coverage >60%

---

## Metrics Snapshot

| Metric | Value |
|--------|-------|
| Lines of Code | 696 |
| Functions | 11 |
| Methods | 28 |
| Packages | 5 |
| Test Files | 0 |
| Avg Function Length | 22.7 lines |
| High Complexity (>10) | 1 function (`ProcessPayment`: 19.4) |
| Doc Coverage | 42.1% |
| Code Duplication | 2.64% (35 lines) |
| TODOs in Code | 3 |

---

## Dependency Notes

- **Go 1.21.3**: Supported until Go 1.25 release (~Feb 2027)
- **go-monero-rpc-client**: Community maintained; no known CVEs; consider pinning version
- **btcd v0.24.2**: Stable Bitcoin library; well-maintained

---

## Competitive Context

Createon's unique value proposition vs alternatives:
- **vs Ghost**: Native crypto support (Ghost requires manual integration)
- **vs NotOnlyFans**: Supports Monero (privacy coin) in addition to BTC
- **Differentiator**: Flat-file simplicity; no database setup required

To maintain competitive position, priorities 1-3 (buildable binary, documented CLI, user sessions) are essential for any real-world use.
