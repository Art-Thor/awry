package app

import "github.com/Art-Thor/awry/pkg/shellenv"

const (
	ShellBash = shellenv.ShellBash
	ShellZsh  = shellenv.ShellZsh
	ShellFish = shellenv.ShellFish
)

// ExportCommand returns a POSIX-shell-safe command that sets AWS_PROFILE.
func ExportCommand(profile string) string {
	return shellenv.ExportCommand(profile)
}

// ExportCommandForShell returns shell code that sets AWS_PROFILE.
func ExportCommandForShell(profile, shell string) string {
	return shellenv.ExportCommandForShell(profile, shell)
}
