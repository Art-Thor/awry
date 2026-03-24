package main

import "testing"

func TestWhoamiProfile(t *testing.T) {
	tests := []struct {
		name        string
		profileFlag string
		current     string
		want        string
		wantErr     bool
	}{
		{name: "uses explicit flag", profileFlag: "prod-admin", current: "sandbox-admin", want: "prod-admin"},
		{name: "falls back to current", current: "sandbox-admin", want: "sandbox-admin"},
		{name: "errors when no profile", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := whoamiProfile(tt.profileFlag, tt.current)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("whoamiProfile() = %q, want %q", got, tt.want)
			}
		})
	}
}
