# AUDIT — 2026-03-18

## Project Goals

Createon is a **self-hosted alternative to Patreon** that enables creators to monetize content using cryptocurrency payments. Based on the README, the project claims to provide:

**Payment Features:**
- Bitcoin (BTC) payments
- Monero (XMR) payments
- Configurable payment timeouts
- Automatic payment verification

**Creator Management:**
- Multiple subscription tiers with custom pricing
- Markdown content support
- Profile customization (avatar, bio, social links)

**Content Management:**
- Markdown-based posts
- Tier-restricted content
- Content versioning
- Tags and categories

**Subscription System:**
- Automated payment processing
- Tier-based access control
- Subscription expiration handling
- Payment status tracking

**Architecture:**
- File-based storage (no database required)
- Simple backup/restore
- Portable deployments
- Thread-safe operations

**CLI Interface:**
- `createon server` — run the web server
- `createon creator add/list` — manage creators
- `createon post publish` — publish content
- `createon sub verify/list` — manage subscriptions
- `createon backup create/restore` — backup data

**Target Audience:** Privacy-conscious creators seeking independence from mainstream platforms, and supporters who prefer cryptocurrency payments.

---

## Goal-Achievement Summary

| Goal | Status | Evidence |
|------|--------|----------|
| Bitcoin (BTC) payments | ⚠️ Partial | `pkg/subscription/manager.go:55-66` configures paywall with BTC; `TestNet: true` hardcoded at line 58 |
| Monero (XMR) payments | ⚠️ Partial | `pkg/subscription/manager.go:63-65` uses hardcoded XMR config; no runtime configuration |
| Configurable payment timeouts | ⚠️ Partial | `config.go:22` defines `Timeout` field; hardcoded `24 * time.Hour` at `pkg/cli/server.go:93` |
| Automatic payment verification | ✅ Achieved | `pkg/subscription/manager.go:115-170` ProcessPayment checks paywall store status |
| Multiple subscription tiers | ✅ Achieved | `types.go:18` Creator.Tiers slice; `pkg/cli/creator.go:30` supports `-t` flag |
| Custom pricing per tier | ✅ Achieved | `types.go:26-27` Tier.PriceBTC and Tier.PriceXMR fields |
| Markdown content support | ✅ Achieved | `pkg/templates/manager.go:46-59` goldmark with GFM, typographer, auto heading IDs |
| Profile customization | ⚠️ Partial | `types.go:14-17` has AvatarPath, SocialLinks fields; no upload endpoint or rendering |
| Markdown-based posts | ✅ Achieved | `pkg/cli/server.go:231-236` reads `.md` files from posts directory |
| Tier-restricted content | ✅ Achieved | `pkg/subscription/verify.go:15-64` verifyAccessImpl checks tier hierarchy |
| Content versioning | ❌ Missing | No version history or diff capability found in codebase |
| Tags and categories | ⚠️ Partial | `types.go:36` Post.Tags field exists; no filtering, UI rendering, or CLI support |
| Automated payment processing | ✅ Achieved | `pkg/subscription/manager.go:33-104` CreateSubscription and ProcessPayment |
| Tier-based access control | ✅ Achieved | `pkg/subscription/verify.go:66-94` tierIncludesAccess implements hierarchy |
| Subscription expiration handling | ✅ Achieved | `pkg/subscription/verify.go:33-36` checks ExpiresAt |
| Payment status tracking | ✅ Achieved | `pkg/subscription/manager.go:147` updates payment status to "confirmed" |
| No database required | ✅ Achieved | All data stored as YAML files; `pkg/files/manager.go` implements file operations |
| Simple backup/restore | ❌ Missing | README documents `createon backup` commands; not implemented |
| Portable deployments | ✅ Achieved | Flat-file architecture is inherently portable |
| Thread-safe operations | ✅ Achieved | `pkg/files/manager.go:38-50` uses per-path RWMutex; atomic writes via temp+rename |
| CLI server command | ✅ Achieved | `pkg/cli/server.go:27-40` implements `server` subcommand |
| CLI creator management | ✅ Achieved | `pkg/cli/creator.go:16-40` implements `creator add/list` |
| CLI post publishing | ❌ Missing | README documents `createon post publish`; not implemented |
| CLI subscription management | ❌ Missing | README documents `createon sub verify/list`; not implemented |
| Web server functionality | ❌ CRITICAL | `cmd/createon/main.go:1` is empty; project cannot build |
| User session management | ❌ Missing | TODO at `pkg/cli/server.go:249,274`; hardcoded `user@example.com` |

**Overall: 13/26 goals fully achieved (50%), 6 partial (23%), 7 missing (27%)**

---

## Findings

### CRITICAL

- [ ] **Empty entry point prevents project build** — `cmd/createon/main.go:1` — The main.go file is empty (contains only a newline), making the entire project unbuildable. All documented CLI commands are unreachable. — **Remediation:** Create `cmd/createon/main.go` with:
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
  **Validation:** `go build -o createon ./cmd/createon && ./createon --help`

- [ ] **fmt.Errorf missing argument causes build failure** — `pkg/cli/creator.go:50` — Format string `%w` has no corresponding argument, causing compilation failure. — **Remediation:** Change line 50 from:
  ```go
  return fmt.Errorf("File manager creation error: %w")
  ```
  to:
  ```go
  return fmt.Errorf("File manager creation error: %w", err)
  ```
  **Validation:** `go vet ./pkg/cli/...`

- [ ] **fmt.Errorf missing argument causes build failure** — `pkg/cli/creator.go:100` — Same issue as above. — **Remediation:** Change line 100 from:
  ```go
  return fmt.Errorf("File manager creation error: %w")
  ```
  to:
  ```go
  return fmt.Errorf("File manager creation error: %w", err)
  ```
  **Validation:** `go vet ./pkg/cli/...`

### HIGH

- [ ] **Hardcoded TestNet mode prevents production use** — `pkg/subscription/manager.go:58` — `TestNet: true` is hardcoded, ignoring `Config.Paywall.TestNet`. Production deployments cannot process real payments. — **Remediation:** Pass config through Manager struct and use `cfg.Paywall.TestNet` instead of literal `true`. **Validation:** Change config to `testnet: false`, verify paywall initializes in mainnet mode.

- [ ] **Hardcoded XMR configuration** — `pkg/subscription/manager.go:106-112` — `getXMRConfig()` returns hardcoded `127.0.0.1:18081` with no authentication. Users cannot configure external Monero nodes. — **Remediation:** Add `XMRHost`, `XMRUser`, `XMRPassword` fields to `Config.Paywall` struct and wire through to `getXMRConfig()`. **Validation:** Set non-default values in config, verify they're used.

- [ ] **Payment timeout configuration ignored** — `pkg/cli/server.go:93` — Uses hardcoded `24 * time.Hour` instead of parsed `Config.Paywall.Timeout`. — **Remediation:** Parse timeout string with `time.ParseDuration(cfg.Paywall.Timeout)` and use result. **Validation:** Set `timeout: "1h"` in config, verify payment timeout is 1 hour.

- [ ] **User session management not implemented** — `pkg/cli/server.go:249,274` — Hardcoded `user@example.com` in `handleViewPost` and `handleSubscribe`. No login/logout system exists. — **Remediation:** Implement session-based authentication with cookies or JWT. Wire actual user email from session into access verification calls. **Validation:** Multiple users can subscribe and access only their own subscribed content.

- [ ] **ProcessPayment has excessive complexity** — `pkg/subscription/manager.go:115` — Cyclomatic complexity 13, nesting depth 5 (overall 19.4). High risk for bugs, difficult to test. — **Remediation:** Extract subscription-finding loop into `findSubscriptionByPaymentID(paymentID string) (*Subscription, int, error)` helper. Extract transaction ID update logic into separate method. **Validation:** `go-stats-generator analyze . | grep ProcessPayment` shows complexity < 10.

- [ ] **btcd v0.24.2 has known critical vulnerability** — `go.mod:17` — CVE-2024-38365 affects btcd consensus logic (FindAndDelete). While v0.24.2 is listed as the fix version, the project should document minimum node version requirements. — **Remediation:** Add security note to README documenting Bitcoin node compatibility requirements and upgrade recommendations. **Validation:** README contains btcd version guidance.

### MEDIUM

- [ ] **No test files exist** — entire codebase — Zero test coverage for any package. — **Remediation:** Add unit tests starting with `pkg/files/manager_test.go` (atomic writes), `pkg/subscription/verify_test.go` (tier hierarchy), and `pkg/subscription/manager_test.go` (payment flow). **Validation:** `go test ./...` reports > 0 test functions.

- [ ] **Code duplication in file operations** — `pkg/files/manager.go:76-100` and `pkg/files/manager.go:145-167` — 25 lines of atomic write logic duplicated between WriteYAML and WriteFile. — **Remediation:** Extract shared logic into `atomicWrite(fullPath string, writeFunc func(*os.File) error) error` helper. **Validation:** `go-stats-generator analyze . | grep "Duplication Ratio"` shows < 2%.

- [ ] **Smaller code duplication in file operations** — `pkg/files/manager.go:129-138` and `pkg/files/manager.go:193-202` — 10 lines of similar file reading/listing patterns. — **Remediation:** Consider extracting common lock acquisition pattern. **Validation:** Re-run duplication analysis.

- [ ] **Missing GoDoc comments on exported functions** — `pkg/subscription/manager.go:24`, `pkg/cli/root.go:15`, `config.go:26` — `NewManager`, `Execute`, and `LoadConfig` lack documentation. Overall doc coverage 42.1%. — **Remediation:** Add GoDoc comments explaining purpose, parameters, and return values. **Validation:** `go-stats-generator analyze . | grep "Function Coverage"` shows > 80%.

- [ ] **Empty error message in CreateSubscription** — `pkg/subscription/manager.go:68` — Error message is just ` : %w` which provides no context if paywall creation fails. — **Remediation:** Change to `fmt.Errorf("failed to configure paywall: %w", err)`. **Validation:** Manually trigger paywall error, verify meaningful error message.

- [ ] **Template references undefined variables** — `templates/post.html:7-8` — References `.Creator.Username` and `.Post.CreatedAt` but `handleViewPost` at `pkg/cli/server.go:261-264` doesn't pass Creator or Post in PageData. — **Remediation:** Pass creator and post metadata in PageData struct. **Validation:** Visit a post page, verify no template errors.

- [ ] **Payment template references .Content as Tier but receives Payment** — `templates/payment.html:6-10` — Template expects `.Content` to have `.Name`, `.Description`, `.PriceBTC` (Tier fields), but `handleSubscribe` at `pkg/cli/server.go:281` passes `*paywall.Payment`. — **Remediation:** Either pass tier object in Content, or update template to render payment addresses/amounts from Payment struct. **Validation:** Subscribe flow shows payment information correctly.

### LOW

- [ ] **Package name doesn't match directory** — `createon` package in repo root — `go-stats-generator` reports naming convention violation. — **Remediation:** Either rename directory to `createon` or accept this as intentional for import path aesthetics. **Validation:** Acknowledged as design choice.

- [ ] **types.go is a generic filename** — `types.go` — Naming convention suggests more specific naming. — **Remediation:** Consider renaming to `models.go` or splitting by domain (e.g., `creator.go`, `subscription.go`). **Validation:** Optional improvement.

- [ ] **Memory leak potential in file manager locks** — `pkg/files/manager.go:227-233` — `Cleanup()` method exists but is never called. Lock map grows unbounded over time. — **Remediation:** Call `Cleanup()` periodically or implement LRU eviction for lock map. **Validation:** Under load testing, memory usage remains stable.

- [ ] **Tier parsing uses fmt.Sscanf with potential issues** — `pkg/cli/creator.go:63` — `Sscanf` with `%s:%s:%s` may not correctly parse tier names containing spaces. — **Remediation:** Use `strings.Split(t, ":")` for more reliable parsing. **Validation:** Create tier with multi-word name, verify correct parsing.

---

## Metrics Snapshot

| Metric | Value |
|--------|-------|
| Lines of Code | 696 |
| Total Functions | 11 |
| Total Methods | 28 |
| Total Packages | 5 |
| Test Files | 0 |
| Avg Function Length | 22.7 lines |
| Longest Function | CreateSubscription (70 lines) |
| High Complexity (>10) | 3 functions |
| Highest Complexity | ProcessPayment (19.4) |
| Documentation Coverage | 42.1% |
| Function Doc Coverage | 40.0% |
| Type Doc Coverage | 5.9% |
| Code Duplication | 2.64% (35 lines) |
| TODOs in Code | 3 |
| Circular Dependencies | None |

---

## Dependency Analysis

| Dependency | Version | Status | Notes |
|------------|---------|--------|-------|
| btcsuite/btcd | v0.24.2 | ⚠️ Monitor | CVE-2024-38365 fixed in v0.24.2; ensure Bitcoin node compatibility |
| go-monero-rpc-client | Dec 2024 | ⚠️ Monitor | No direct CVEs; Monero v0.18.4.0+ recommended for node |
| go-chi/chi/v5 | v5.2.0 | ✅ Current | No known issues |
| spf13/cobra | v1.8.1 | ✅ Current | No known issues |
| spf13/viper | v1.19.0 | ✅ Current | No known issues |
| yuin/goldmark | v1.7.8 | ✅ Current | No known issues |
| google/uuid | v1.4.0 | ✅ Current | No known issues |
| opd-ai/paywall | Jan 2025 | ⚠️ Pin | Internal dependency; pin to specific version |

---

## Build Status

| Check | Result | Command |
|-------|--------|---------|
| `go build ./...` | ❌ FAIL | Empty main.go |
| `go vet ./pkg/...` | ❌ FAIL | fmt.Errorf format errors |
| `go test -race ./...` | ❌ FAIL | Build failure |
| Buildable packages | ⚠️ Partial | `pkg/files`, `pkg/templates`, `pkg/subscription` build individually |

---

## Report Metadata

- **Generated:** 2026-03-18
- **Tool Version:** go-stats-generator v1.0.0
- **Go Version:** 1.21.3
- **Files Analyzed:** 11
- **Analysis Time:** 51ms
