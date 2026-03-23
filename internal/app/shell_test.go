package app

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
