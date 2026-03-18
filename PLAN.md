# Implementation Plan: Complete Content Management Features

## Project Context
- **What it does**: Createon is a self-hosted Patreon alternative enabling cryptocurrency (BTC/XMR) monetization with flat-file storage.
- **Current goal**: Implement content versioning—the highest-priority missing feature claimed in README.
- **Estimated Scope**: Medium (11 functions above complexity 9.0, 4 feature gaps remaining)

## Goal-Achievement Status
| Stated Goal | Current Status | This Plan Addresses |
|-------------|---------------|---------------------|
| Content versioning | ❌ Missing | **Yes** |
| Tags and categories | ⚠️ Partial (data exists, no UI/filtering) | Yes |
| Profile customization | ⚠️ Partial (fields exist, no upload/CLI) | Yes |
| Test coverage >50% | ⚠️ Partial (30% weighted average) | Yes |
| Bitcoin (BTC) payments | ✅ Achieved | No |
| Monero (XMR) payments | ✅ Achieved | No |
| Tier-based access control | ✅ Achieved | No |
| Subscription management | ✅ Achieved | No |
| Backup/restore | ✅ Achieved | No |
| Thread-safe operations | ✅ Achieved | No |

## Metrics Summary
- **Complexity hotspots on goal-critical paths**: 11 functions above threshold 9.0
  - `CreateSubscription` (15.3) — subscription creation
  - `verifyAccessImpl` (15.0) — access control
  - `runSubList` (14.0) — CLI
  - `runServer` (14.0) — HTTP server setup
  - `runBackupRestore` (12.7) — backup
  - `runListCreators` (11.4) — CLI
  - `atomicWrite` (10.9) — file operations
  - `handleSubscribe` (10.1) — HTTP handler
  - `runPostPublish` (9.6) — post publishing (goal-critical for versioning)
  - `handleViewPost` (9.6) — post viewing
  - `findSubscriptionByPaymentID` (9.3) — payment lookup
- **Duplication ratio**: 0.42% (10 duplicated lines in `pkg/files/manager.go:135-144` and `167-176`)
- **Doc coverage**: 88.2% overall (functions: 100%, methods: 82%, types: 90%)
- **Package coupling**: `cli` package (1125 lines) concentrates HTTP handlers, CLI commands, and business logic—potential future separation point

## Implementation Steps

### Step 1: Implement Content Versioning Core
- **Deliverable**: Add version storage mechanism for posts
  - Add `Version int` field to `Post` struct in `types.go`
  - Create versioned directory structure: `data/creators/{username}/posts/{post-id}/v{n}.md`
  - Modify `pkg/files/manager.go` to preserve previous versions on update
  - Add `GetPostVersion()` and `ListPostVersions()` methods to file manager
- **Dependencies**: None (foundational work)
- **Goal Impact**: Directly implements "Content versioning" feature claimed in README
- **Acceptance**: Post update creates new version file; previous version preserved; `go test ./pkg/files/...` passes
- **Validation**: 
  ```bash
  # Create post, update twice, verify 3 versions exist
  createon post publish testuser content.md -t "Test"
  createon post publish testuser content.md -t "Test v2"
  ls data/creators/testuser/posts/*/  # Should show v1.md, v2.md
  ```

### Step 2: Add Version CLI Commands
- **Deliverable**: CLI commands for version management
  - Add `post history [username] [post-id]` subcommand in `pkg/cli/post.go`
  - Add `post revert [username] [post-id] [version]` subcommand
  - Display version list with timestamps and sizes
- **Dependencies**: Step 1 (versioning storage)
- **Goal Impact**: Makes versioning user-accessible; completes the "Content versioning" feature
- **Acceptance**: `createon post history` lists versions; `createon post revert` restores content
- **Validation**:
  ```bash
  createon post history testuser test-post-id  # Lists versions
  createon post revert testuser test-post-id 1  # Restores v1
  ```

### Step 3: Complete Tag Filtering and Display
- **Deliverable**: Functional tags system
  - Update `templates/post.html` to render tags as clickable links
  - Add route `GET /c/{username}/tags/{tag}` in `pkg/cli/server.go`
  - Implement `ListPostsByTag()` in `pkg/files/manager.go` using `PostFilter.Tags`
  - Add `post list --tag=<tag>` CLI flag
  - Add tag cloud section to `templates/profile.html`
- **Dependencies**: None (independent feature)
- **Goal Impact**: Completes "Tags and categories" feature; enables content discovery
- **Acceptance**: Tags visible on posts; clicking tag filters posts; CLI filters work
- **Validation**:
  ```bash
  go-stats-generator analyze ./pkg/cli/server.go --skip-tests --format json | grep -c "handleTagFilter"
  createon post list testuser --tag=tutorial  # Filters by tag
  curl http://localhost:8080/c/testuser/tags/tutorial  # Returns filtered posts
  ```

### Step 4: Complete Profile Customization
- **Deliverable**: Avatar and social link management
  - Add `-a/--avatar` and `-s/--social` flags to `creator add` in `pkg/cli/creator.go`
  - Add `creator update` command for modifying existing profiles
  - Add `POST /c/{username}/avatar` endpoint for file uploads
  - Store avatars to `data/creators/{username}/avatar.{ext}`
  - Serve avatars via `/assets/avatars/` static route
  - Render social links in `templates/profile.html`
- **Dependencies**: None (independent feature)
- **Goal Impact**: Completes "Profile customization" feature
- **Acceptance**: `creator add --avatar=./photo.jpg --social="twitter.com/x"` works; profile displays both
- **Validation**:
  ```bash
  createon creator add testuser -n "Test" -a ./avatar.png -s "github.com/test"
  ls data/creators/testuser/avatar.*  # Avatar file exists
  curl http://localhost:8080/c/testuser | grep -c "github.com/test"  # Social link rendered
  ```

### Step 5: Expand Test Coverage to Critical Paths
- **Deliverable**: Unit tests for untested packages
  - Create `pkg/auth/auth_test.go`:
    - `TestRegisterUser` (success, duplicate email)
    - `TestLoginUser` (success, wrong password, no user)
    - `TestSessionManagement` (create, validate, expire)
  - Create `pkg/cli/cli_test.go`:
    - `TestCreatorAddCommand`
    - `TestPostPublishCommand`
    - `TestBackupRestoreRoundtrip`
  - Create `pkg/templates/templates_test.go`:
    - `TestRenderMarkdown` (GFM features)
    - `TestTemplateExecution` (all templates render)
- **Dependencies**: Steps 1-4 (test new features)
- **Goal Impact**: Reduces regression risk; enables confident refactoring
- **Acceptance**: `go test -cover ./...` reports >50% overall
- **Validation**:
  ```bash
  go test -cover ./... 2>&1 | grep "coverage"
  # Target: pkg/auth >70%, overall >50%
  ```

### Step 6: Reduce Complexity Hotspots
- **Deliverable**: Refactor highest-complexity functions
  - Extract `generatePaymentAddresses()` helper from `CreateSubscription` (15.3 → <12)
  - Extract `validateAndLoadSubscription()` from `verifyAccessImpl` (15.0 → <12)
  - Extract duplicated atomic write logic in `pkg/files/manager.go:135-144` and `167-176` into shared `writeAtomically()` helper
  - Consider extracting HTTP handlers from `pkg/cli/server.go` into `pkg/handlers/` if time permits
- **Dependencies**: Step 5 (tests protect refactoring)
- **Goal Impact**: Improves maintainability; reduces bug surface
- **Acceptance**: No function with complexity >12; duplication ratio <0.3%
- **Validation**:
  ```bash
  go-stats-generator analyze . --skip-tests --format json --sections functions,duplication 2>/dev/null | \
    python3 -c "import sys,json; d=json.load(sys.stdin); \
    high=[f['name'] for f in d['functions'] if f['complexity']['overall']>12]; \
    print('High complexity:', high or 'None'); \
    print('Duplication:', d['duplication']['duplication_ratio'])"
  ```

## Dependency Graph
```
Step 1 (Versioning Core)
    └── Step 2 (Version CLI)
            └── Step 5 (Tests) ──→ Step 6 (Refactoring)
                      ↑
Step 3 (Tags) ────────┘
Step 4 (Profiles) ────┘
```

## Risk Assessment

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| Versioning breaks existing posts | Low | High | Migrate existing posts to v1 on first access; add version detection |
| Test coverage slows development | Medium | Low | Prioritize critical paths (auth, subscription); defer CLI tests if needed |
| Refactoring introduces regressions | Medium | Medium | Complete Step 5 before Step 6; run tests continuously |

## Dependency Status (No Action Required)

| Dependency | Version | Security Status |
|------------|---------|-----------------|
| btcd | v0.24.2 | ✅ Patched (CVE-2024-38365 fixed) |
| go-monero-rpc-client | Dec 2024 | ✅ Maintained |
| cobra | v1.8.1 | ✅ Current |
| chi | v5.2.0 | ✅ Current |
| goldmark | v1.7.8 | ✅ Current |
| Go | 1.21.3 | ✅ Supported until ~Feb 2027 |

## Success Criteria

Completing Steps 1-4 achieves **100% of README-stated features**.

| Milestone | Steps | Verification |
|-----------|-------|--------------|
| Content versioning complete | 1-2 | `post history` and `post revert` work |
| Tags feature complete | 3 | Posts filterable by tag via URL and CLI |
| Profile customization complete | 4 | Avatar and social links settable and visible |
| Test coverage threshold | 5 | `go test -cover ./...` reports >50% |
| Complexity reduced | 6 | No function with complexity >12 |

## Estimated Effort

| Step | Effort | Notes |
|------|--------|-------|
| Step 1 | 4-6 hours | Core versioning logic; migration handling |
| Step 2 | 2-3 hours | CLI is straightforward with existing patterns |
| Step 3 | 3-4 hours | Route + template + filter logic |
| Step 4 | 3-4 hours | File upload adds complexity |
| Step 5 | 4-6 hours | Test writing is time-intensive |
| Step 6 | 2-3 hours | Mechanical refactoring with test safety net |
| **Total** | **18-26 hours** | ~3-4 developer days |
