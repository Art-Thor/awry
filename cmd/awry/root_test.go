package main

import "testing"

func TestCurrentProfileOutput(t *testing.T) {
	tests := []struct {
		name    string
		current string
		want    string
		wantErr bool
	}{
		{name: "active profile", current: "managed-services", want: "managed-services"},
		{name: "missing profile", current: "", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := currentProfileOutput(tt.current)
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
				t.Fatalf("currentProfileOutput(%q) = %q, want %q", tt.current, got, tt.want)
			}
		})
	}
}
