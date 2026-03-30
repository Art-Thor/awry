package shellenv

import "fmt"

const (
	ShellBash = "bash"
	ShellZsh  = "zsh"
	ShellFish = "fish"
)

func ExportCommand(profile string) string {
	return ExportCommandForShell(profile, "")
}

func ExportCommandForShell(profile, shell string) string {
	switch shell {
	case ShellFish:
		return fmt.Sprintf("set -gx AWS_PROFILE %s", fishQuote(profile))
	case ShellBash, ShellZsh, "":
		return fmt.Sprintf("export AWS_PROFILE=%s", posixQuote(profile))
	default:
		return fmt.Sprintf("export AWS_PROFILE=%s", posixQuote(profile))
	}
}

func posixQuote(s string) string {
	if s == "" {
		return "''"
	}

	quoted := "'"
	for _, r := range s {
		if r == '\'' {
			quoted += `'"'"'`
			continue
		}
		quoted += string(r)
	}
	quoted += "'"

	return quoted
}

func fishQuote(s string) string {
	if s == "" {
		return "''"
	}

	quoted := "'"
	for _, r := range s {
		if r == '\'' {
			quoted += `\'`
			continue
		}
		quoted += string(r)
	}
	quoted += "'"

	return quoted
}
