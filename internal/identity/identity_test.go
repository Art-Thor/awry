package identity

import (
	"context"
	"errors"
	"testing"
)

func TestLookup(t *testing.T) {
	original := loadCallerIdentity
	t.Cleanup(func() {
		loadCallerIdentity = original
	})

	t.Run("returns normalized identity", func(t *testing.T) {
		loadCallerIdentity = func(ctx context.Context, profile string) (*callerIdentityOutput, error) {
			return &callerIdentityOutput{
				Account: stringPtr("123456789012"),
				Arn:     stringPtr("arn:aws:sts::123456789012:assumed-role/Admin/sandbox-admin"),
			}, nil
		}

		identity, err := Lookup(context.Background(), "sandbox-admin")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if identity.Profile != "sandbox-admin" {
			t.Fatalf("unexpected profile: %q", identity.Profile)
		}
		if identity.AccountID != "123456789012" {
			t.Fatalf("unexpected account id: %q", identity.AccountID)
		}
		if identity.Principal != "sandbox-admin" {
			t.Fatalf("unexpected principal: %q", identity.Principal)
		}
	})

	t.Run("requires profile", func(t *testing.T) {
		_, err := Lookup(context.Background(), "")
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("maps expired sso errors", func(t *testing.T) {
		loadCallerIdentity = func(ctx context.Context, profile string) (*callerIdentityOutput, error) {
			return nil, errors.New("the SSO session associated with this profile has expired or is otherwise invalid")
		}

		_, err := Lookup(context.Background(), "sandbox-admin")
		if err == nil || err.Error() != "AWS SSO session expired for profile \"sandbox-admin\"; run `aws sso login --profile sandbox-admin`" {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("maps missing credentials errors", func(t *testing.T) {
		loadCallerIdentity = func(ctx context.Context, profile string) (*callerIdentityOutput, error) {
			return nil, errors.New("no valid credential sources for S3Backend found")
		}

		_, err := Lookup(context.Background(), "sandbox-admin")
		if err == nil || err.Error() != "no valid AWS credentials available for profile \"sandbox-admin\"" {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("wraps unknown errors", func(t *testing.T) {
		loadCallerIdentity = func(ctx context.Context, profile string) (*callerIdentityOutput, error) {
			return nil, errors.New("boom")
		}

		_, err := Lookup(context.Background(), "sandbox-admin")
		if err == nil || err.Error() != "looking up identity for profile \"sandbox-admin\": boom" {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestPrincipalFromARN(t *testing.T) {
	tests := []struct {
		name string
		arn  string
		want string
	}{
		{name: "assumed role", arn: "arn:aws:sts::123456789012:assumed-role/Admin/sandbox-admin", want: "sandbox-admin"},
		{name: "iam user", arn: "arn:aws:iam::123456789012:user/alice", want: "alice"},
		{name: "no slash", arn: "arn:aws:iam::123456789012:root", want: "arn:aws:iam::123456789012:root"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := principalFromARN(tt.arn); got != tt.want {
				t.Fatalf("principalFromARN(%q) = %q, want %q", tt.arn, got, tt.want)
			}
		})
	}
}

func stringPtr(value string) *string {
	return &value
}
