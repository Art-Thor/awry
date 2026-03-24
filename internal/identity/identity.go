package identity

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

// Identity describes the resolved AWS caller identity for a profile.
type Identity struct {
	Profile   string
	AccountID string
	ARN       string
	Principal string
}

var loadCallerIdentity = func(ctx context.Context, profile string) (*callerIdentityOutput, error) {
	cmd := exec.CommandContext(ctx, "aws", "sts", "get-caller-identity", "--profile", profile, "--output", "json")
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		message := strings.TrimSpace(stderr.String())
		if message == "" {
			message = err.Error()
		}
		return nil, fmt.Errorf("%s", message)
	}

	var output callerIdentityOutput
	if err := json.Unmarshal(stdout.Bytes(), &output); err != nil {
		return nil, fmt.Errorf("parsing aws sts output: %w", err)
	}

	return &output, nil
}

type callerIdentityOutput struct {
	Account *string `json:"Account"`
	Arn     *string `json:"Arn"`
}

// Lookup resolves STS caller identity for the given profile.
func Lookup(ctx context.Context, profile string) (Identity, error) {
	if profile == "" {
		return Identity{}, fmt.Errorf("profile name is required")
	}

	output, err := loadCallerIdentity(ctx, profile)
	if err != nil {
		return Identity{}, classifyLookupError(ctx, profile, err)
	}

	identity := Identity{Profile: profile}
	if output.Account != nil {
		identity.AccountID = *output.Account
	}
	if output.Arn != nil {
		identity.ARN = *output.Arn
		identity.Principal = principalFromARN(*output.Arn)
	}

	return identity, nil
}

func classifyLookupError(ctx context.Context, profile string, err error) error {
	message := strings.ToLower(err.Error())

	switch {
	case strings.Contains(message, "no valid credential") || strings.Contains(message, "failed to refresh cached credentials"):
		return fmt.Errorf("no valid AWS credentials available for profile %q", profile)
	case strings.Contains(message, "sso session") && strings.Contains(message, "expired"):
		return fmt.Errorf("AWS SSO session expired for profile %q; run `aws sso login --profile %s`", profile, profile)
	case strings.Contains(message, "token has expired"):
		return fmt.Errorf("AWS session expired for profile %q", profile)
	default:
		if ctx.Err() == context.DeadlineExceeded {
			return fmt.Errorf("timed out looking up identity for profile %q", profile)
		}
		return fmt.Errorf("looking up identity for profile %q: %w", profile, err)
	}
}

func principalFromARN(arn string) string {
	parts := strings.Split(arn, "/")
	if len(parts) == 0 {
		return arn
	}
	return parts[len(parts)-1]
}
