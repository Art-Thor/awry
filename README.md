# awry

A terminal-based AWS profile manager with a clean TUI interface.

Browse, inspect, and switch AWS profiles without leaving your terminal.

![Go](https://img.shields.io/badge/Go-1.23+-00ADD8?logo=go&logoColor=white)
![License](https://img.shields.io/badge/license-MIT-blue)
[![Release](https://img.shields.io/github/v/release/Art-Thor/awry)](https://github.com/Art-Thor/awry/releases)

## Features

- **Dual-panel TUI** — profile list + detail view side by side
- **SSO v2 support** — resolves `sso_session` references automatically
- **Profile type detection** — identifies SSO, Role, and Static credential profiles
- **Validation-aware profile health** — surfaces invalid, expired, and missing-credential states
- **Fuzzy search** — press `/` to filter profiles by name
- **Active profile highlight** — shows which profile is currently set
- **Shell integration** — install a shell wrapper so `awry` updates your current shell
- **Identity and session details** — see who you are and how long the active session has left

## Install

Pick one installation method.

If you are not sure which one to use, use Homebrew on macOS/Linux or `go install` if you already have Go set up.

**Homebrew:**

```bash
brew install Art-Thor/tap/awry
```

**Go:**

```bash
go install github.com/Art-Thor/awry/cmd/awry@latest
```

**From source:**

```bash
git clone https://github.com/Art-Thor/awry.git
cd awry
make build
```

If you built from source, run the local binary with:

```bash
./awry
```

## Quick Start

If you want `awry` to work like this:

- run `awry`
- choose an AWS profile
- press `Enter`
- your current shell now uses that profile

do this once after installing:

```bash
awry setup-shell
```

Then run the command it prints, for example:

```bash
source ~/.zshrc
```

or:

```bash
source ~/.bashrc
```

After that, verify the shell integration is active:

```bash
type awry
```

You should see `awry is a function`.

Now the normal flow works:

```bash
awry
awry current
echo "$AWS_PROFILE"
```

## First-Time Setup

### 1. Make sure you already have AWS profiles

`awry` reads your existing AWS CLI configuration. It does not create AWS profiles for you.

If you already use `aws sso login`, `aws configure`, or named AWS profiles, you are probably ready.

To check, run:

```bash
aws configure list-profiles
```

If that prints profile names, `awry` should be able to see them.

### 2. Install shell integration once

Run:

```bash
awry setup-shell
```

This updates your shell startup file so `awry` can change the current shell session after you press `Enter`.

### 3. Reload your shell config

If `awry setup-shell` says it updated `~/.zshrc`, run:

```bash
source ~/.zshrc
```

If it updated `~/.bashrc`, run:

```bash
source ~/.bashrc
```

If it also mentions `~/.bash_profile`, that is expected on some bash setups.

### 4. Confirm the wrapper is active

Run:

```bash
type awry
```

Expected result:

```bash
awry is a function
```

If you only see a file path or binary path, the shell config has not been reloaded yet.

## Usage

```bash
# One-time setup for your current shell
awry setup-shell

# Interactive selection and apply in one shot after setup
awry

# List all profiles
awry list

# Show current profile
awry current

# Show AWS caller identity for the active profile
awry whoami

# Switch directly to a specific profile
awry use my-profile

# Bypass the shell wrapper and call the real binary directly
command awry list
```

## TUI Keybindings

| Key | Action |
|-----|--------|
| `↑` `↓` / `j` `k` | Navigate profiles |
| `Enter` | Select profile |
| `r` | Refresh session and identity |
| `?` | Open keyboard help |
| `/` | Fuzzy search |
| `Esc` | Clear search |
| `q` | Quit |

## How It Works

awry reads your AWS configuration from `~/.aws/config` and `~/.aws/credentials`, merges them, and presents a unified view.

Profiles also include lightweight health signals so broken or incomplete local AWS setup is easier to spot before switching. In the TUI, `awry` can surface states such as:

- `[EXPIRED]` for expired active sessions
- `[NO CREDS]` when the active profile has no usable credentials
- `[INVALID]` for obviously broken profile definitions such as unknown auth type or a role profile without `source_profile`

To let `awry` actually change your current shell, install the wrapper once:

```bash
awry setup-shell
```

That appends the right setup line to your shell config and tells you which file to `source`.

If you prefer to install the wrapper manually for the current session, use:

```bash
eval "$(command awry init bash)"
```

or:

```bash
eval "$(command awry init zsh)"
```

That defines a shell function named `awry` which calls the real binary and automatically evaluates the emitted export command for `awry` and `awry use ...`.

Use `command awry ...` if you ever want to bypass the wrapper and call the underlying binary directly.

Under the hood, the binary still emits:

```bash
export AWS_PROFILE='your-profile'
```

Without the wrapper, a standalone binary cannot modify the parent shell, so use `eval "$(awry)"` or `eval "$(awry use <profile>)"` directly.

For role profiles, `awry` only sets `AWS_PROFILE`. The actual assume-role or SSO resolution still happens later in the AWS CLI or SDK.

If the selected profile is currently active, the detail pane also shows live runtime information when available:

- session status and remaining lifetime
- AWS account ID
- full caller ARN
- principal or role/user name

For any highlighted profile, the detail pane also shows:

- normalized profile type
- health status
- region
- `source_profile` and `role_arn` when present
- the export command that `Enter` will emit

Press `r` in the TUI to refresh those runtime details after re-authenticating or re-running `aws sso login`.
Press `?` in the TUI to see the full keyboard reference without leaving the app.

## Screenshots

If you plan to share `awry` publicly, add a screenshot or short GIF here showing:

- the profile list and detail pane
- the active profile runtime badge
- session time left and identity details
- the `?` help overlay

## Troubleshooting

### `awry` shows profiles, but `awry current` still says no active profile set

That usually means shell integration is not loaded yet.

Run:

```bash
awry setup-shell
type awry
```

If `type awry` does not say `awry is a function`, reload your shell config with `source ~/.zshrc` or `source ~/.bashrc`.

### A profile shows `[INVALID]` in the TUI

That means `awry` found a profile entry but the local configuration looks incomplete.

Common causes:

- a role profile is missing `source_profile`
- a profile exists in config but has no usable auth settings
- your AWS configuration is partially defined across files

Check the profile in:

```bash
~/.aws/config
~/.aws/credentials
```

Then refresh the TUI with `r` if the affected profile is active.

### A profile shows `[NO CREDS]` or `[EXPIRED]`

- `[NO CREDS]` usually means the active profile cannot currently authenticate
- `[EXPIRED]` usually means your SSO or temporary session needs to be refreshed

Typical fix:

```bash
aws sso login --profile <profile>
```

Then return to `awry` and press `r` to refresh runtime state.

### `awry setup-shell` updated the wrong shell file

You can choose the shell explicitly:

```bash
awry setup-shell zsh
```

or:

```bash
awry setup-shell bash
```

### I only want to test once without changing shell config

Use:

```bash
eval "$(command awry)"
```

or:

```bash
eval "$(command awry use my-profile)"
```

### How do I know which profile is active right now?

Use:

```bash
awry current
echo "$AWS_PROFILE"
```

## Configuration

awry respects these environment variables:

| Variable | Description |
|----------|-------------|
| `AWS_CONFIG_FILE` | Custom path to AWS config file |
| `AWS_SHARED_CREDENTIALS_FILE` | Custom path to credentials file |
| `AWS_PROFILE` | Currently active profile |
| `AWS_DEFAULT_PROFILE` | Fallback active profile |

## Shell Setup

Recommended:

```bash
awry setup-shell
```

Manual persistent setup:

```bash
echo 'eval "$(command awry init zsh)"' >> ~/.zshrc
source ~/.zshrc
```

```bash
echo 'eval "$(command awry init bash)"' >> ~/.bashrc
source ~/.bashrc
```

## License

MIT
