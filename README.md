# awry

A terminal-based AWS profile manager with a clean TUI interface.

Browse, inspect, and emit shell commands for AWS profiles without leaving your terminal.

![Go](https://img.shields.io/badge/Go-1.23+-00ADD8?logo=go&logoColor=white)
![License](https://img.shields.io/badge/license-MIT-blue)
[![Release](https://img.shields.io/github/v/release/Art-Thor/awry)](https://github.com/Art-Thor/awry/releases)

## Features

- **Dual-panel TUI** — profile list + detail view side by side
- **SSO v2 support** — resolves `sso_session` references automatically
- **Profile type detection** — identifies SSO, Role, and Static credential profiles
- **Fuzzy search** — press `/` to filter profiles by name
- **Active profile highlight** — shows which profile is currently set
- **Shell integration** — select a profile and emit an `export` command ready for `eval`

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
# Interactive selection and apply in one shot
eval "$(awry)"

# Launch the TUI without applying the result
awry

# List all profiles
awry list

# Show current profile
awry current

# Print shell code for a profile
awry use my-profile

# Apply a specific profile in one shot
eval "$(awry use my-profile)"
```

## TUI Keybindings

| Key | Action |
|-----|--------|
| `↑` `↓` / `j` `k` | Navigate profiles |
| `Enter` | Emit shell command |
| `/` | Fuzzy search |
| `Esc` | Clear search |
| `q` | Quit |

## How It Works

awry reads your AWS configuration from `~/.aws/config` and `~/.aws/credentials`, merges them, and presents a unified view. When you select a profile, it outputs:

```bash
export AWS_PROFILE='your-profile'
```

Wrap it with `eval` to apply the switch to your current shell session.

Important: `awry` cannot modify the parent shell by itself. Running `awry` directly lets you browse and prints shell code, but your shell only changes when you use `eval "$(awry)"` or `eval "$(awry use <profile>)"`.

For role profiles, `awry` only sets `AWS_PROFILE`. The actual assume-role or SSO resolution still happens later in the AWS CLI or SDK.

## Configuration

awry respects these environment variables:

| Variable | Description |
|----------|-------------|
| `AWS_CONFIG_FILE` | Custom path to AWS config file |
| `AWS_SHARED_CREDENTIALS_FILE` | Custom path to credentials file |
| `AWS_PROFILE` | Currently active profile |
| `AWS_DEFAULT_PROFILE` | Fallback active profile |

## Shell Alias

Add to your `~/.bashrc` or `~/.zshrc` for quick profile switching:

```bash
alias ap='eval "$(awry)"'
```

## License

MIT
