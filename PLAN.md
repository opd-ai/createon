# Implementation Plan: Achieve Functional Minimum Viable Product

## Project Context
- **What it does**: Createon is a self-hosted Patreon alternative enabling creators to monetize content using Bitcoin and Monero cryptocurrency payments, with flat-file storage and tiered subscriptions.
- **Current goal**: Make the project buildable and achieve MVP status with core documented features working end-to-end.
- **Estimated Scope**: Medium (7 functions above complexity threshold, 2.6% duplication, 42% doc coverage)

## Goal-Achievement Status
| Stated Goal | Current Status | This Plan Addresses |
|-------------|---------------|---------------------|
| Project builds and runs | ❌ Empty main.go | Yes |
| Bitcoin (BTC) payments | ⚠️ Partial (hardcoded testnet) | Yes |
| Monero (XMR) payments | ⚠️ Partial (hardcoded config) | Yes |
| Configurable payment timeouts | ⚠️ Partial (ignored) | Yes |
| Automatic payment verification | ✅ Achieved | No |
| Multiple subscription tiers | ✅ Achieved | No |
| Markdown content support | ✅ Achieved | No |
| Profile customization | ⚠️ Partial (no upload/render) | No (lower priority) |
| Tier-restricted content | ✅ Achieved | No |
| Content versioning | ❌ Missing | No (future work) |
| Tags and categories | ⚠️ Partial (data only) | No (lower priority) |
| CLI post publishing | ❌ Missing | Yes |
| CLI subscription management | ❌ Missing | Yes |
| CLI backup/restore | ❌ Missing | Yes |
| User session management | ❌ Missing | Yes |
| Test coverage | ❌ 0% | Yes |

## Metrics Summary
- **Complexity hotspots on goal-critical paths**: 7 functions above threshold (>9.0)
  - `ProcessPayment` (19.4) — payment flow critical path
  - `verifyAccessImpl` (15.0) — subscription verification critical path
  - `CreateSubscription` (11.4) — payment flow critical path
  - `runListCreators` (11.4) — CLI operations
  - `WriteYAML`/`WriteFile` (10.9 each) — data persistence
  - `runServer` (10.9) — server initialization
- **Duplication ratio**: 2.6% (35 lines, 2 clone pairs in `pkg/files/manager.go`)
- **Doc coverage**: 42.1% overall (functions: 40%, types: 5.9%)
- **Package coupling**: `cli` has highest coupling (3.5), depends on 7 external packages
- **TODO annotations**: 3 (user session at server.go:249,274; config at manager.go:58)

---

## Implementation Steps

### Step 1: Fix Critical Build Blockers
- **Deliverable**: Working binary that can be built and executed
- **Dependencies**: None (first step)
- **Goal Impact**: Enables all other work; transforms project from non-functional to usable
- **Files to modify**:
  - `cmd/createon/main.go` — Create valid entry point calling `cli.Execute()`
  - `pkg/cli/creator.go:50,100` — Fix `fmt.Errorf` missing `err` argument
- **Acceptance**: `go build -o createon ./cmd/createon && ./createon --help` succeeds
- **Validation**: 
  ```bash
  go build ./cmd/createon && go vet ./...
  ```

### Step 2: Wire Payment Configuration Through
- **Deliverable**: Runtime-configurable cryptocurrency node connections and payment settings
- **Dependencies**: Step 1 (project must build)
- **Goal Impact**: Enables production use with real payments (BTC and XMR)
- **Files to modify**:
  - `config.go` — Add `XMRHost`, `XMRUser`, `XMRPassword`, `BTCHost` to `PaywallConfig` struct
  - `pkg/subscription/manager.go:58` — Use `cfg.Paywall.TestNet` instead of hardcoded `true`
  - `pkg/subscription/manager.go:106-112` — Use config values in `getXMRConfig()`
  - `pkg/cli/server.go:93` — Parse and use `cfg.Paywall.Timeout` duration
- **Acceptance**: 
  - Changing `config/server.yaml` `testnet: false` initializes mainnet mode
  - Setting `xmr_host: remote.node:18081` uses that host
  - Setting `timeout: "1h"` creates subscriptions with 1-hour timeout
- **Validation**:
  ```bash
  go-stats-generator analyze ./pkg/subscription ./config.go --format json | jq '.documentation.todo_comments | map(select(.description | contains("configurable"))) | length' # should be 0
  ```

### Step 3: Implement Post Publishing CLI
- **Deliverable**: `createon post publish [username] [file.md] -t [title] -r [tier]` command
- **Dependencies**: Step 1 (project must build)
- **Goal Impact**: Fulfills documented CLI interface; enables content workflow
- **Files to create**:
  - `pkg/cli/post.go` — New file with `post` subcommand and `publish` action
- **Logic**:
  1. Generate UUID for post ID
  2. Read markdown from input file
  3. Create `data/creators/{username}/posts/{post-id}.md`
  4. Create/update metadata YAML with title, tier, tags, timestamps
- **Acceptance**: 
  ```bash
  echo "# Test Post" > /tmp/test.md
  ./createon post publish testcreator /tmp/test.md -t "My Title" -r tier1
  ls data/creators/testcreator/posts/*.md  # shows new file
  ```
- **Validation**:
  ```bash
  go-stats-generator analyze ./pkg/cli/post.go --format json | jq '.functions | map(select(.complexity.overall > 15)) | length' # should be 0
  ```

### Step 4: Implement Subscription Management CLI
- **Deliverable**: `createon sub verify` and `createon sub list` commands
- **Dependencies**: Step 1 (project must build)
- **Goal Impact**: Fulfills documented CLI interface; enables subscription administration
- **Files to create**:
  - `pkg/cli/subscription.go` — New file with `sub` subcommand
- **Logic**:
  - `sub verify [email] [creator] [tier]` — Call `Manager.VerifyAccess()`, print result
  - `sub list [creator]` — Call `Manager.GetActiveSubscriptions()`, print formatted table
- **Acceptance**:
  ```bash
  ./createon sub list testcreator  # outputs subscription table or "No active subscriptions"
  ./createon sub verify user@test.com testcreator tier1  # outputs "Access: granted" or "Access: denied"
  ```
- **Validation**:
  ```bash
  go build ./cmd/createon && ./createon sub --help  # shows subcommands
  ```

### Step 5: Implement Backup/Restore CLI
- **Deliverable**: `createon backup create` and `createon backup restore` commands
- **Dependencies**: Step 1 (project must build)
- **Goal Impact**: Fulfills documented CLI interface; enables data portability
- **Files to create**:
  - `pkg/cli/backup.go` — New file with `backup` subcommand
- **Logic**:
  - `backup create [file.tar.gz]` — tar/gzip the data directory
  - `backup restore [file.tar.gz]` — extract archive to data directory (with confirmation)
- **Acceptance**:
  ```bash
  ./createon backup create /tmp/backup.tar.gz
  tar -tzf /tmp/backup.tar.gz | head  # shows data/ contents
  ```
- **Validation**:
  ```bash
  go build ./cmd/createon && ./createon backup --help  # shows subcommands
  ```

### Step 6: Implement User Session Management
- **Deliverable**: Cookie-based authentication with login/logout/dashboard routes
- **Dependencies**: Step 1 (project must build)
- **Goal Impact**: Enables multi-user system; removes hardcoded `user@example.com`
- **Files to modify/create**:
  - `types.go` — Add `User` struct with email, password hash, created_at
  - `pkg/cli/server.go` — Add `/login`, `/logout`, `/register`, `/dashboard` routes
  - `pkg/cli/server.go:249,274` — Replace hardcoded email with session user
  - `templates/login.html` — New login form template
  - `templates/dashboard.html` — New user dashboard template
- **Logic**:
  1. Users stored in `data/users/{email-hash}.yaml`
  2. Session cookies with secure random token
  3. Middleware extracts user from cookie, adds to request context
- **Acceptance**: 
  - Two different users can register and log in
  - Each user can only access their own subscribed content
  - `/dashboard` shows user's active subscriptions
- **Validation**:
  ```bash
  go-stats-generator analyze ./pkg/cli/server.go --format json | jq '.documentation.todo_comments | map(select(.description | contains("session"))) | length' # should be 0
  ```

### Step 7: Add Missing Templates
- **Deliverable**: `templates/home.html` and `templates/profile.html` templates
- **Dependencies**: Step 1 (project must build)
- **Goal Impact**: Enables homepage and creator profile pages to render
- **Files to create**:
  - `templates/home.html` — List all creators with links to profiles
  - `templates/profile.html` — Creator profile with tier listing
- **Files to modify**:
  - `pkg/cli/server.go:172-201` — Verify template data matches template expectations
- **Acceptance**:
  ```bash
  curl -s http://localhost:8080/ | grep -q "creators" && echo "Homepage works"
  ```
- **Validation**:
  ```bash
  ./createon server &
  curl -s http://localhost:8080/ | grep -i error || echo "No template errors"
  ```

### Step 8: Reduce ProcessPayment Complexity
- **Deliverable**: `ProcessPayment` function refactored to complexity < 15
- **Dependencies**: Step 1 (project must build)
- **Goal Impact**: Reduces bug risk in critical payment path; improves maintainability
- **Files to modify**:
  - `pkg/subscription/manager.go:115-170` — Extract helper functions:
    - `findSubscriptionByPaymentID(paymentID string) (*Subscription, int, error)`
    - `updateSubscriptionStatus(sub *Subscription, txID string) error`
- **Acceptance**: Function complexity drops from 19.4 to < 15
- **Validation**:
  ```bash
  go-stats-generator analyze ./pkg/subscription/manager.go --format json | jq '.functions[] | select(.name == "ProcessPayment") | .complexity.overall' # < 15
  ```

### Step 9: Eliminate Code Duplication in File Manager
- **Deliverable**: Single `atomicWrite` helper replacing duplicated logic
- **Dependencies**: Step 1 (project must build)
- **Goal Impact**: Reduces maintenance burden; prevents divergent behavior
- **Files to modify**:
  - `pkg/files/manager.go:76-100` and `145-167` — Extract to shared `atomicWrite` function
- **Acceptance**: Duplication ratio drops from 2.6% to < 2%
- **Validation**:
  ```bash
  go-stats-generator analyze ./pkg/files --format json | jq '.duplication.duplication_ratio' # < 0.02
  ```

### Step 10: Add Core Unit Tests
- **Deliverable**: Test files for critical paths with > 50% coverage on those paths
- **Dependencies**: Steps 1-6 (core functionality must exist)
- **Goal Impact**: Enables safe refactoring; validates correctness; establishes CI foundation
- **Files to create**:
  - `pkg/files/manager_test.go` — Test atomic writes, locking, read/write round-trips
  - `pkg/subscription/verify_test.go` — Test tier hierarchy logic, expiration handling
  - `pkg/subscription/manager_test.go` — Test payment flow (with mock paywall)
- **Acceptance**: `go test ./...` runs > 10 test functions and all pass
- **Validation**:
  ```bash
  go test ./... -v | grep -c "=== RUN" # > 10
  go test ./... 2>&1 | grep -q "PASS" && echo "Tests pass"
  ```

### Step 11: Improve Documentation Coverage
- **Deliverable**: GoDoc comments on all exported functions and types
- **Dependencies**: Steps 1-9 (code must be stable before documenting)
- **Goal Impact**: Enables contributors; improves maintainability
- **Files to modify**:
  - `config.go:26` — Add GoDoc for `LoadConfig`
  - `pkg/cli/root.go:15` — Add GoDoc for `Execute`
  - `pkg/subscription/manager.go:24` — Add GoDoc for `NewManager`
  - `types.go` — Add GoDoc for all exported types
- **Acceptance**: Documentation coverage > 60%
- **Validation**:
  ```bash
  go-stats-generator analyze . --format json | jq '.documentation.coverage.overall' # > 60
  ```

### Step 12: Create CI Pipeline
- **Deliverable**: `.github/workflows/ci.yml` running build, vet, and tests on push
- **Dependencies**: Step 10 (tests must exist)
- **Goal Impact**: Prevents regressions; enables confident contributions
- **Files to create**:
  - `.github/workflows/ci.yml`
- **Workflow jobs**:
  1. `go build ./...`
  2. `go vet ./...`
  3. `go test -race ./...`
- **Acceptance**: Push to repository triggers green CI run
- **Validation**: GitHub Actions shows passing workflow

---

## Scope Calibration

| Metric | Current | Target | Classification |
|--------|---------|--------|----------------|
| Functions above complexity 9.0 | 7 | <5 | Medium |
| Duplication ratio | 2.6% | <2% | Small |
| Doc coverage gap | 57.9% | <40% | Large |
| Build status | ❌ Fails | ✅ Passes | Critical |
| Test coverage | 0% | >30% | Large |

**Overall Scope**: Medium — The project has solid foundations but requires work across multiple areas to achieve MVP. Critical path is Steps 1-6 (buildable binary with documented CLI and authentication).

---

## Dependency Graph

```
Step 1 (Build Blockers) ─┬─> Step 2 (Config)
                         ├─> Step 3 (Post CLI)
                         ├─> Step 4 (Sub CLI)
                         ├─> Step 5 (Backup CLI)
                         ├─> Step 6 (Auth) ────────> Step 10 (Tests) ─> Step 12 (CI)
                         ├─> Step 7 (Templates)
                         ├─> Step 8 (Refactor ProcessPayment)
                         └─> Step 9 (Refactor Duplication) ─> Step 11 (Docs)
```

---

## Risk Notes

- **btcd v0.24.2**: The project uses the patched version that fixes CVE-2024-38365. Document minimum node version requirements in README for users running Bitcoin nodes.
- **go-monero-rpc-client**: Community-maintained with no direct CVEs. Pin version in go.mod for reproducibility.
- **opd-ai/paywall**: Internal dependency at unstable version (0.0.0-date). Consider tagging releases for production stability.

---

## Validation Commands Summary

```bash
# After Step 1: Build succeeds
go build -o createon ./cmd/createon && ./createon --help

# After Step 2: Config wired through
go-stats-generator analyze . --format json | jq '.documentation.todo_comments | length'  # 0 TODOs about config

# After Steps 3-5: CLI commands exist
./createon post --help && ./createon sub --help && ./createon backup --help

# After Step 6: Auth works
curl -c cookies.txt -d "email=test@test.com&password=test" http://localhost:8080/login
curl -b cookies.txt http://localhost:8080/dashboard | grep -q "subscriptions"

# After Step 8: Complexity reduced
go-stats-generator analyze ./pkg/subscription --format json | jq '[.functions[] | select(.complexity.overall > 15)] | length'  # 0

# After Step 9: Duplication reduced
go-stats-generator analyze . --format json | jq '.duplication.duplication_ratio'  # < 0.02

# After Step 10: Tests pass
go test ./... -v | grep -c "PASS"  # > 0

# After Step 11: Docs improved
go-stats-generator analyze . --format json | jq '.documentation.coverage.overall'  # > 60
```

---

*Generated: 2026-03-18 | Tool: go-stats-generator v1.0.0*
