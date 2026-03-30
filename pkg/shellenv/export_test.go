package shellenv

import "testing"

func TestExportCommand(t *testing.T) {
	tests := []struct {
		name    string
		profile string
		want    string
	}{
		{name: "simple", profile: "managed-services", want: "export AWS_PROFILE='managed-services'"},
		{name: "space", profile: "dev sandbox", want: "export AWS_PROFILE='dev sandbox'"},
		{name: "single quote", profile: "team's-prod", want: "export AWS_PROFILE='team'\"'\"'s-prod'"},
		{name: "metacharacters", profile: "prod; rm -rf /", want: "export AWS_PROFILE='prod; rm -rf /'"},
		{name: "empty", profile: "", want: "export AWS_PROFILE=''"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ExportCommand(tt.profile); got != tt.want {
				t.Fatalf("ExportCommand(%q) = %q, want %q", tt.profile, got, tt.want)
			}
		})
	}
}

func TestExportCommandForShell(t *testing.T) {
	tests := []struct {
		name    string
		shell   string
		profile string
		want    string
	}{
		{name: "bash", shell: "bash", profile: "sandbox", want: "export AWS_PROFILE='sandbox'"},
		{name: "zsh", shell: "zsh", profile: "sandbox", want: "export AWS_PROFILE='sandbox'"},
		{name: "fish", shell: "fish", profile: "sandbox", want: "set -gx AWS_PROFILE 'sandbox'"},
		{name: "fish with space", shell: "fish", profile: "dev sandbox", want: "set -gx AWS_PROFILE 'dev sandbox'"},
		{name: "fish with quote", shell: "fish", profile: "team's-prod", want: "set -gx AWS_PROFILE 'team\\'s-prod'"},
		{name: "unknown shell falls back", shell: "nu", profile: "sandbox", want: "export AWS_PROFILE='sandbox'"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ExportCommandForShell(tt.profile, tt.shell); got != tt.want {
				t.Fatalf("ExportCommandForShell(%q, %q) = %q, want %q", tt.profile, tt.shell, got, tt.want)
			}
		})
	}
}
