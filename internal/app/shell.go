package app

import "fmt"

// ExportCommand returns a POSIX-shell-safe command that sets AWS_PROFILE.
func ExportCommand(profile string) string {
	return fmt.Sprintf("export AWS_PROFILE=%s", shellQuote(profile))
}

func shellQuote(s string) string {
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
