package awsconfig

import "os"

// CurrentProfile returns the currently active AWS profile name
// by checking environment variables.
func CurrentProfile() string {
	if v := os.Getenv("AWS_PROFILE"); v != "" {
		return v
	}
	return os.Getenv("AWS_DEFAULT_PROFILE")
}
