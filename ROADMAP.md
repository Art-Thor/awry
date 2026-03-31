# Roadmap

The `awry` roadmap is organized by release milestones. This is not a week-by-week schedule, but a list of what we want to build.

## v0.2.0 — Profile Switching Foundation

- Status: shipped in `v0.2.x`
- `awry use <profile>` command that prints a shell export and works with `eval`
- Shell-safe interactive output so TUI/status rendering never leaks ANSI sequences into command substitution
- Explicit shell integration flow: selecting a profile updates the current shell through `eval $(awry)` or the `awry init` shell wrapper for bash, zsh, and fish
- Role profiles switch cleanly via `AWS_PROFILE`, with docs that clarify assume-role happens later in AWS CLI/SDK
- Active profile pinned to the top of the TUI list
- Active profile badge visible in the list
- Profile type badges visible in the list: `[SSO]`, `[ROLE]`, `[STATIC]`

Delivered in `v0.2.x`:

- shell-safe export generation shared across TUI and CLI commands
- `awry init [bash|zsh|fish]` shell wrapper setup for real in-shell profile switching
- updated README and release checks for shell-switching workflows

## v0.3.0 — STS Identity

- Status: shipped in `v0.3.x`
- `awry whoami` command that calls `sts:GetCallerIdentity`
- Detail panel shows Account ID, ARN, and User/Role on `i`
- Graceful handling for expired or missing credentials

Delivered in `v0.3.x`:

- `awry whoami` for the active or requested profile
- active-profile detail pane shows session lifetime, account ID, ARN, and principal
- TUI refresh support for runtime session and identity state via `r`
- friendly expired and missing-credential messaging in the detail pane

## v0.4.0 — Validation and Error States

- Status: shipped in `v0.4.0`
- Detect invalid or expired profiles more clearly in the TUI
- Status badges: `[EXPIRED]`, `[INVALID]`, `[NO CREDS]`
- Better matching reliability and ambiguity errors for `awry use` and `awry export`
- SSO session expiration checks by reading cached token state

Delivered in `v0.4.0`:

- clearer matching resolution order and ambiguity output
- ranked profile suggestions for ambiguous matches
- TUI health field and invalid-profile indicators
- improved empty-state and current-profile guidance

## v0.5.0 — Favorites and Recents

- Pin or unpin profiles with `p`, persisted in `~/.config/awry/config.yaml`
- Track recently used profiles and show them in a dedicated section
- Config layer via Viper

## v0.6.0 — Safe Mode

- Configurable production patterns such as `prod`, `production`, and `live`
- Visual warning banner on production profiles
- Optional confirmation before switching to production
- Color-coded danger levels

## v0.7.0 — SSO Login

- `awry login <profile>` runs the SSO OIDC flow
- `Ctrl+L` logs in the selected SSO profile from the TUI
- Session expiration countdown in the detail panel

## v0.8.0 — Role Chains and Architecture

- Visualize role assumption chains such as `default -> base -> prod-admin`
- Refactor into `/internal/aws`, `/internal/session`, and `/internal/shell`

## v0.9.0 — Doctor and Validation

- Add `awry doctor` to validate the full local AWS setup
- Explain broken config, missing credentials, and expired session causes
- Add fix-oriented guidance for common local AWS problems

## v0.10.0 — Help and UX Polish

- `?` opens a hotkey overlay
- Responsive layout that adapts to terminal size
- Color themes with `NO_COLOR` support
- Demo GIF and polished README with screenshots

## v0.11.0 — Tags and Filtering

- User-defined tags per profile in config
- Filter by tag in the TUI with `t`
- `awry list --tag prod`

## v1.0.0 — Stable Release

- Full test coverage on critical paths
- Shell completion for bash, zsh, and fish
- Polished error messages
- Launch materials for Reddit and Hacker News
