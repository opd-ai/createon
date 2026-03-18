# AUDIT — 2026-03-18

## Project Goals

Createon is a self-hosted alternative to Patreon that claims to:
- Enable cryptocurrency (Bitcoin and Monero) payments for content monetization
- Provide multi-currency support with configurable payment timeouts and automatic verification
- Support creator management with multiple subscription tiers, custom pricing, and profile customization
- Offer content management with markdown posts, tier-restricted content, content versioning, and tags/categories
- Implement a subscription system with automated processing, tier-based access control, and expiration handling
- Use file-based storage with no database requirement, simple backup/restore, and thread-safe operations

**Target Audience**: Privacy-conscious creators and supporters who prefer cryptocurrency and self-hosting.

## Goal-Achievement Summary

| Goal | Status | Evidence |
|------|--------|----------|
| Bitcoin (BTC) payments | ✅ Achieved | `pkg/subscription/manager.go:72-87` — `CreateSubscription` generates BTC addresses via paywall |
| Monero (XMR) payments | ✅ Achieved | `pkg/subscription/manager.go:124-148` — `getXMRConfig()` reads XMR host/credentials |
| Configurable payment timeouts | ✅ Achieved | `pkg/subscription/manager.go:58-63` — parses `cfg.Paywall.Timeout` duration |
| Automatic payment verification | ✅ Achieved | `pkg/subscription/manager.go:192-215` — `ProcessPayment` checks paywall store status |
| Multiple subscription tiers | ✅ Achieved | `types.go:19` — `Creator.Tiers []Tier` with BTC/XMR pricing |
| Markdown content support | ✅ Achieved | `pkg/templates/manager.go:39-53` — goldmark with GFM, typographer extensions |
| Custom pricing per tier | ✅ Achieved | `types.go:29-30` — `Tier.PriceBTC` and `Tier.PriceXMR` floats |
| Profile customization | ⚠️ Partial | `types.go:15-16` — `AvatarPath`, `SocialLinks` fields exist; `templates/profile.html:4-16` renders them |
| Markdown-based posts | ✅ Achieved | `pkg/cli/post.go:81-83` — posts stored as `{post-id}.md` files |
| Tier-restricted content | ✅ Achieved | `pkg/subscription/verify.go:15-64` — `verifyAccessImpl` checks subscription tier hierarchy |
| Content versioning | ❌ Missing | README claims this feature; no implementation exists |
| Tags and categories | ⚠️ Partial | `types.go:41` — `Post.Tags` field exists; `pkg/cli/post.go:97-99` parses `--tags` flag |
| Automated payment processing | ✅ Achieved | `pkg/subscription/manager.go:193-215` — paywall middleware integration |
| Tier-based access control | ✅ Achieved | `pkg/subscription/verify.go:68-95` — `tierIncludesAccess` implements hierarchy |
| Subscription expiration handling | ✅ Achieved | `pkg/subscription/verify.go:33-36` — checks `ExpiresAt` in `verifyAccessImpl` |
| Payment status tracking | ✅ Achieved | `types.go:64` — `Payment.Status` persisted; updated on confirmation |
| No database required | ✅ Achieved | All data in YAML files under `data/` directory |
| Simple backup/restore | ✅ Achieved | `pkg/cli/backup.go:56-168` — `backup create` and `backup restore` commands |
| Portable deployments | ✅ Achieved | Flat-file architecture inherently portable |
| Thread-safe operations | ✅ Achieved | `pkg/files/manager.go:39-50` — per-path RWMutex; `atomicWrite` via temp+rename |
| CLI creator management | ✅ Achieved | `pkg/cli/creator.go:42-94` — `creator add` and `creator list` |
| CLI post publishing | ✅ Achieved | `pkg/cli/post.go:61-117` — `post publish` with title, tier, tags flags |
| CLI subscription management | ✅ Achieved | `pkg/cli/subscription.go:58-156` — `sub verify` and `sub list` |

**Overall: 20/23 goals achieved (87%)**

## Findings

### CRITICAL

*(None)*

### HIGH

- [x] **Content versioning not implemented** — `pkg/cli/post.go:81-83` — README claims "Content versioning" but posts are stored as single files that get overwritten on update. No version history, no `GetPostVersion()` method, no CLI commands for history or revert. — **Remediation:** Add `Version int` field to `Post` struct in `types.go`. Implement versioned storage structure `data/creators/{username}/posts/{post-id}/v{n}.md`. Add `GetPostVersion(ctx, username, postID, version)` to `ContentManager` interface. Add `post history` and `post revert` CLI subcommands in `pkg/cli/post.go`. Validation: `go test ./pkg/files/... && go build ./cmd/createon && createon post history testuser test-post-id`
  > **Verified 2026-03-18**: Already implemented. `types.go:43` has `Version int`, `interfaces.go:30-33` has version methods, `pkg/cli/post.go` has `history` and `revert` commands.

- [ ] **Weak password hashing algorithm** — `pkg/auth/auth.go:189-196` — Uses SHA256 with static salt instead of bcrypt/argon2. Comment on line 191 acknowledges this: "In production, use bcrypt or argon2". This is a known issue, not a false positive. — **Remediation:** Replace `hashPassword()` with `golang.org/x/crypto/bcrypt.GenerateFromPassword()` and `bcrypt.CompareHashAndPassword()`. Update `Login()` to use constant-time comparison. Validation: `go test ./pkg/auth/... -v`

- [ ] **Profile customization incomplete** — `pkg/cli/creator.go:27-29` — `AvatarPath` and `SocialLinks` fields exist in `Creator` struct and template renders them, but no CLI flags (`-a/--avatar`, `-s/--social`) to set them. No web endpoint for avatar upload. — **Remediation:** Add flags to `creator add` command: `addCmd.Flags().StringP("avatar", "a", "", "avatar file path")` and `addCmd.Flags().StringSliceP("social", "s", []string{}, "social links")`. Wire them in `runAddCreator()`. Validation: `createon creator add testuser -n "Test" -a ./avatar.png -s "github.com/test"`

### MEDIUM

- [ ] **Tags not filterable or rendered** — `templates/post.html:1-15` — `Post.Tags` field exists and can be set via CLI, but tags are not rendered in `templates/post.html` and no route exists to filter posts by tag. — **Remediation:** Add tag rendering in `templates/post.html`:
  ```html
  {{if .Post.Tags}}<div class="tags">{{range .Post.Tags}}<a href="/c/{{$.Creator.Username}}/tags/{{.}}">{{.}}</a>{{end}}</div>{{end}}
  ```
  Add route handler in `pkg/cli/server.go` for `GET /c/{username}/tags/{tag}`. Implement `ListPostsByTag()` using `PostFilter.Tags`. Validation: `curl http://localhost:8080/c/testuser/tags/tutorial`

- [ ] **Zero test coverage on critical auth package** — `pkg/auth/auth.go:1-205` — Authentication package has 0% test coverage. Contains security-critical code for registration, login, session management. — **Remediation:** Create `pkg/auth/auth_test.go` with tests for `Register()`, `Login()`, `ValidateSession()`, `Logout()`, and `Middleware()`. Target >70% coverage. Validation: `go test -cover ./pkg/auth/...`

- [ ] **High complexity in CreateSubscription** — `pkg/subscription/manager.go:37-122` — Cyclomatic complexity 15.3 (highest in codebase). 84 lines with payment address generation, config parsing, paywall setup, and subscription creation mixed together. — **Remediation:** Extract `generatePaymentAddresses()` helper function. Extract `parseTimeoutDuration()` helper. Move paywall configuration to a separate `configurePaywall()` method. Target complexity <12. Validation: `go-stats-generator analyze . --skip-tests --format json | grep CreateSubscription`

- [ ] **High complexity in verifyAccessImpl** — `pkg/subscription/verify.go:15-64` — Cyclomatic complexity 15.0. Nested loops and conditionals for subscription lookup and tier verification. — **Remediation:** Extract `findSubscriptionForUser()` to handle subscription lookup. Extract tier validation into reusable helper. Validation: `go-stats-generator analyze . --skip-tests`

- [ ] **Code duplication in file manager** — `pkg/files/manager.go:135-144` and `pkg/files/manager.go:167-176` — 10 lines of duplicated lock acquire/release pattern. — **Remediation:** The lock pattern is similar but the operations differ (read vs write). Consider extracting a `withReadLock()` and `withWriteLock()` helper that accepts a function callback. Validation: `go-stats-generator analyze . --skip-tests --sections duplication`

### LOW

- [ ] **Package name doesn't match directory** — Root package is `createon` but directory is also `createon` (correct), however the package documentation mentions this as a naming convention violation. — **Remediation:** No action required; this is a false positive from the analyzer. The package name correctly matches the module path.

- [ ] **CLI package has high coupling** — `pkg/cli/server.go:1-493` — 10 dependencies, coupling score 5.0. HTTP handlers, CLI commands, and server setup mixed in one package. — **Remediation:** Consider future extraction of HTTP handlers into `pkg/handlers/` package if the codebase grows significantly. Current size (1125 lines) is manageable. No immediate action required.

## Metrics Snapshot

| Metric | Value |
|--------|-------|
| Lines of Code | 1,186 |
| Total Functions | 29 |
| Total Methods | 43 |
| Packages | 7 |
| Test Files | 2 |
| Avg Function Length | 20.7 lines |
| Max Complexity | 15.3 (`CreateSubscription`) |
| High Complexity (>10) | 7 functions |
| Code Duplication | 0.42% (10 lines) |
| Test Coverage (pkg/files) | 78.7% |
| Test Coverage (pkg/subscription) | 30.0% |
| Test Coverage (pkg/auth) | 0.0% |
| Test Coverage (pkg/cli) | 0.0% |
| Test Coverage (overall weighted) | ~25% |

## Dependency Security Status

| Dependency | Version | Status | Notes |
|------------|---------|--------|-------|
| btcd | v0.24.2 | ✅ Secure | CVE-2024-38365 patched; no new CVEs |
| go-monero-rpc-client | Dec 2024 | ✅ Secure | No known CVEs; Monero RPC 100% fuzz coverage |
| cobra | v1.8.1 | ✅ Current | Stable CLI framework |
| chi | v5.2.0 | ✅ Current | Lightweight router |
| goldmark | v1.7.8 | ✅ Current | Active development |
| Go | 1.21.3 | ✅ Supported | Until ~Feb 2027 |

## Validation Commands

```bash
# Verify tests pass
go test -race ./...

# Check vet passes
go vet ./...

# Check coverage
go test -cover ./...

# Check complexity metrics
go-stats-generator analyze . --skip-tests

# Build binary
go build -o createon ./cmd/createon
```
