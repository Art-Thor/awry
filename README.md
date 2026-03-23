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
- **Fuzzy search** — press `/` to filter profiles by name
- **Active profile highlight** — shows which profile is currently set
- **Shell integration** — install a shell wrapper so `awry` updates your current shell

## Install

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

## Usage

```bash
# One-time setup for zsh
echo 'eval "$(command awry init zsh)"' >> ~/.zshrc
source ~/.zshrc

# One-time setup for bash
echo 'eval "$(command awry init bash)"' >> ~/.bashrc
source ~/.bashrc

# Interactive selection and apply in one shot after setup
awry

# List all profiles
awry list

# Show current profile
awry current

# Switch directly to a specific profile
awry use my-profile
```

## TUI Keybindings

| Key | Action |
|-----|--------|
| `↑` `↓` / `j` `k` | Navigate profiles |
| `Enter` | Select profile |
| `/` | Fuzzy search |
| `Esc` | Clear search |
| `q` | Quit |

## How It Works

awry reads your AWS configuration from `~/.aws/config` and `~/.aws/credentials`, merges them, and presents a unified view.

To let `awry` actually change your current shell, install the wrapper once:

```bash
eval "$(command awry init zsh)"
```

or:

```bash
eval "$(command awry init bash)"
```

That defines a shell function named `awry` which calls the real binary and automatically evaluates the emitted export command for `awry` and `awry use ...`.

Use `command awry ...` if you ever want to bypass the wrapper and call the underlying binary directly.

Under the hood, the binary still emits:

```bash
export AWS_PROFILE='your-profile'
```

Without the wrapper, a standalone binary cannot modify the parent shell, so use `eval "$(awry)"` or `eval "$(awry use <profile>)"` directly.

For role profiles, `awry` only sets `AWS_PROFILE`. The actual assume-role or SSO resolution still happens later in the AWS CLI or SDK.

## Configuration

awry respects these environment variables:

| Variable | Description |
|----------|-------------|
| `AWS_CONFIG_FILE` | Custom path to AWS config file |
| `AWS_SHARED_CREDENTIALS_FILE` | Custom path to credentials file |
| `AWS_PROFILE` | Currently active profile |
| `AWS_DEFAULT_PROFILE` | Fallback active profile |

## Shell Setup

Add one of these lines to your shell config for persistent setup:

```bash
eval "$(command awry init zsh)"
```

```bash
eval "$(command awry init bash)"
```

## License

MIT
