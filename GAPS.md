# Implementation Gaps — 2026-03-18

## Content Versioning

- **Stated Goal**: README claims "Content versioning" under Content Management features. Users should be able to track edit history and revert content changes.
- **Current State**: Posts are stored as single markdown files (`{post-id}.md`) that are overwritten on each update. No version history is preserved. The `ContentManager` interface in `interfaces.go` defines `UpdatePost()` but has no version-related methods.
- **Impact**: Creators cannot recover from accidental edits, track content evolution, or audit changes over time. This is a significant gap for a content platform where editorial control is essential.
- **Closing the Gap**:
  1. Add `Version int` field to `Post` struct in `types.go`
  2. Change storage structure to `data/creators/{username}/posts/{post-id}/v{n}.md`
  3. Modify `pkg/files/manager.go` to auto-increment version and preserve previous files
  4. Add `GetPostVersion(ctx, username, postID, version)` and `ListPostVersions(ctx, username, postID)` to `ContentManager` interface
  5. Add `post history [username] [post-id]` CLI subcommand
  6. Add `post revert [username] [post-id] [version]` CLI subcommand
  7. Estimated effort: 4-6 hours

## Tag Filtering and Display

- **Stated Goal**: README claims "Tags and categories" under Content Management. Users should be able to organize and discover content by topic.
- **Current State**: `Post.Tags` field exists in `types.go:41`. CLI flag `--tags` parses comma-separated tags in `pkg/cli/post.go:97-99`. However, tags are not rendered in `templates/post.html`, no filtering route exists (`/c/{username}/tags/{tag}`), and `PostFilter.Tags` in `types.go:73` is unused.
- **Impact**: Tags exist only in metadata but serve no functional purpose. Users cannot discover related content or navigate by topic—a basic expectation for any content platform.
- **Closing the Gap**:
  1. Add tag rendering to `templates/post.html`:
     ```html
     {{if .Post.Tags}}<div class="tags">{{range .Post.Tags}}<a href="/c/{{$.Creator.Username}}/tags/{{.}}" class="tag">{{.}}</a>{{end}}</div>{{end}}
     ```
  2. Add route `GET /c/{username}/tags/{tag}` in `pkg/cli/server.go`
  3. Implement tag filtering in `ListPosts()` using the existing `PostFilter.Tags` field
  4. Add `post list --tag=<tag>` CLI flag to filter posts
  5. Add tag cloud section to `templates/profile.html`
  6. Estimated effort: 3-4 hours

## Profile Customization CLI

- **Stated Goal**: README claims "Profile customization" under Creator Management. Creators should be able to set avatars and social links.
- **Current State**: `Creator` struct has `AvatarPath string` and `SocialLinks []string` fields in `types.go:15-16`. Template `templates/profile.html:4-16` renders these fields when present. However, `creator add` command in `pkg/cli/creator.go:28-30` has no flags for avatar or social links. No `creator update` command exists. No web endpoint for avatar upload.
- **Impact**: Creators must manually edit YAML files to set profile images and social links, which defeats the purpose of a user-friendly CLI-driven platform.
- **Closing the Gap**:
  1. Add `-a/--avatar` flag to `creator add` command in `pkg/cli/creator.go`
  2. Add `-s/--social` flag (comma-separated URLs) to `creator add`
  3. Add `creator update` command for modifying existing profiles
  4. Add `POST /c/{username}/avatar` endpoint for web-based file uploads
  5. Store avatars to `data/creators/{username}/avatar.{ext}`
  6. Serve avatar files via static handler at `/assets/avatars/`
  7. Estimated effort: 3-4 hours

## Production-Grade Password Hashing

- **Stated Goal**: Security Considerations section advises "Keep private keys safely stored" and general security best practices. Implicit goal is secure user authentication.
- **Current State**: `pkg/auth/auth.go:191-196` uses SHA256 with a static salt for password hashing. A comment on line 193 explicitly acknowledges: "In production, use bcrypt or argon2".
- **Impact**: SHA256 with static salt is vulnerable to rainbow table attacks and lacks the adaptive cost factor needed for password security. This is a known issue documented in the code itself.
- **Closing the Gap**:
  1. Add dependency: `go get golang.org/x/crypto/bcrypt`
  2. Replace `hashPassword()` with `bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)`
  3. Update `Login()` to use `bcrypt.CompareHashAndPassword()`
  4. Add migration logic to re-hash existing passwords on first login
  5. Estimated effort: 2-3 hours

## Test Coverage for Auth Package

- **Stated Goal**: The project has CI that runs `go test -race` on push/PR (`.github/workflows/ci.yml`). Implicit goal is test coverage for critical paths.
- **Current State**: `pkg/auth/` has 0% test coverage. This package handles registration, login, session management, and HTTP middleware—all security-critical code paths.
- **Impact**: Changes to authentication logic carry high regression risk. Security bugs in session management or password verification could go undetected.
- **Closing the Gap**:
  1. Create `pkg/auth/auth_test.go` with tests for:
     - `TestRegister` (success, duplicate email)
     - `TestLogin` (success, wrong password, nonexistent user)
     - `TestValidateSession` (valid, expired, invalid token)
     - `TestLogout`
     - `TestMiddleware` (with and without session cookie)
  2. Target >70% coverage for `pkg/auth/`
  3. Estimated effort: 4-6 hours

## Complexity Reduction

- **Stated Goal**: Maintainable codebase (implicit in professional software engineering).
- **Current State**: Two functions exceed complexity threshold of 12:
  - `CreateSubscription` (15.3) in `pkg/subscription/manager.go:37-122`
  - `verifyAccessImpl` (15.0) in `pkg/subscription/verify.go:15-64`
- **Impact**: High-complexity functions are harder to test, review, and modify safely. They correlate with higher defect rates.
- **Closing the Gap**:
  1. Extract `generatePaymentAddresses()` helper from `CreateSubscription`
  2. Extract `parseTimeoutDuration()` helper from `CreateSubscription`
  3. Extract `findSubscriptionForUser()` helper from `verifyAccessImpl`
  4. Target: No function with complexity >12
  5. Estimated effort: 2-3 hours (after test coverage in place)

---

## Summary

| Gap | Priority | Effort | Impact |
|-----|----------|--------|--------|
| Content versioning | P1 | 4-6h | High — Feature claimed but missing |
| Tag filtering/display | P2 | 3-4h | Medium — Feature partially implemented |
| Profile customization CLI | P3 | 3-4h | Medium — Feature partially implemented |
| Password hashing | P2 | 2-3h | High — Security vulnerability |
| Auth test coverage | P3 | 4-6h | Medium — Regression risk |
| Complexity reduction | P4 | 2-3h | Low — Maintainability |

**Total estimated effort to close all gaps: 18-26 hours (~3-4 developer days)**

Completing P1-P3 would bring the project to 100% of its stated feature set. P4 items improve maintainability and security posture.
