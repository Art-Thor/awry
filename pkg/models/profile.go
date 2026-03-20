package models

// ProfileType represents the authentication method of an AWS profile.
type ProfileType string

const (
	ProfileTypeStatic ProfileType = "STATIC"
	ProfileTypeSSO    ProfileType = "SSO"
	ProfileTypeRole   ProfileType = "ROLE"
	ProfileTypeUnknown ProfileType = "UNKNOWN"
)

// Profile represents a normalized AWS profile from config/credentials files.
type Profile struct {
	Name          string
	Region        string
	Output        string
	Type          ProfileType
	SourceProfile string
	RoleARN       string
	SSOStartURL   string
	SSOAccountID  string
	SSORegion     string
	SSORoleName   string
	HasCredentials bool // true if static keys exist in credentials file
}

// DisplayRegion returns the region or a fallback if empty.
func (p Profile) DisplayRegion() string {
	if p.Region != "" {
		return p.Region
	}
	return "—"
}
