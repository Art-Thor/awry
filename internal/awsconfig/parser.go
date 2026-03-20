package awsconfig

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/go-ini/ini"

	"github.com/artthor/awry/pkg/models"
)

// DefaultConfigPath returns the path to ~/.aws/config.
func DefaultConfigPath() string {
	if v := os.Getenv("AWS_CONFIG_FILE"); v != "" {
		return v
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".aws", "config")
}

// DefaultCredentialsPath returns the path to ~/.aws/credentials.
func DefaultCredentialsPath() string {
	if v := os.Getenv("AWS_SHARED_CREDENTIALS_FILE"); v != "" {
		return v
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".aws", "credentials")
}

// LoadProfiles reads and merges profiles from config and credentials files.
func LoadProfiles() ([]models.Profile, error) {
	return LoadProfilesFrom(DefaultConfigPath(), DefaultCredentialsPath())
}

// LoadProfilesFrom reads profiles from the given config and credentials paths.
func LoadProfilesFrom(configPath, credentialsPath string) ([]models.Profile, error) {
	profiles := make(map[string]*models.Profile)

	if err := parseConfigFile(configPath, profiles); err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	if err := parseCredentialsFile(credentialsPath, profiles); err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	result := make([]models.Profile, 0, len(profiles))
	for _, p := range profiles {
		p.Type = detectType(p)
		result = append(result, *p)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})

	return result, nil
}

// ssoSession holds data from [sso-session <name>] sections.
type ssoSession struct {
	startURL string
	region   string
}

func parseConfigFile(path string, profiles map[string]*models.Profile) error {
	cfg, err := ini.LoadSources(ini.LoadOptions{
		InsensitiveKeys:  true,
		AllowBooleanKeys: true,
	}, path)
	if err != nil {
		return err
	}

	// First pass: collect sso-session sections.
	sessions := make(map[string]ssoSession)
	for _, section := range cfg.Sections() {
		name := section.Name()
		if strings.HasPrefix(name, "sso-session ") {
			sessName := strings.TrimPrefix(name, "sso-session ")
			sessions[sessName] = ssoSession{
				startURL: section.Key("sso_start_url").String(),
				region:   section.Key("sso_region").String(),
			}
		}
	}

	// Second pass: collect profiles.
	for _, section := range cfg.Sections() {
		name := section.Name()
		if name == "DEFAULT" || strings.HasPrefix(name, "sso-session ") {
			continue
		}

		// Config file uses "profile <name>" prefix, except for "default".
		name = strings.TrimPrefix(name, "profile ")

		p := getOrCreate(profiles, name)
		if v := section.Key("region").String(); v != "" {
			p.Region = v
		}
		if v := section.Key("output").String(); v != "" {
			p.Output = v
		}
		if v := section.Key("source_profile").String(); v != "" {
			p.SourceProfile = v
		}
		if v := section.Key("role_arn").String(); v != "" {
			p.RoleARN = v
		}
		if v := section.Key("sso_start_url").String(); v != "" {
			p.SSOStartURL = v
		}
		if v := section.Key("sso_account_id").String(); v != "" {
			p.SSOAccountID = v
		}
		if v := section.Key("sso_region").String(); v != "" {
			p.SSORegion = v
		}
		if v := section.Key("sso_role_name").String(); v != "" {
			p.SSORoleName = v
		}

		// Resolve SSO v2: sso_session references a [sso-session <name>] section.
		if sessName := section.Key("sso_session").String(); sessName != "" {
			if sess, ok := sessions[sessName]; ok {
				if p.SSOStartURL == "" {
					p.SSOStartURL = sess.startURL
				}
				if p.SSORegion == "" {
					p.SSORegion = sess.region
				}
			}
		}
	}

	return nil
}

func parseCredentialsFile(path string, profiles map[string]*models.Profile) error {
	cfg, err := ini.LoadSources(ini.LoadOptions{
		InsensitiveKeys:  true,
		AllowBooleanKeys: true,
	}, path)
	if err != nil {
		return err
	}

	for _, section := range cfg.Sections() {
		name := section.Name()
		if name == "DEFAULT" {
			continue
		}

		p := getOrCreate(profiles, name)

		if section.HasKey("aws_access_key_id") {
			p.HasCredentials = true
		}
		// Pick up region from credentials if not set in config.
		if v := section.Key("region").String(); v != "" && p.Region == "" {
			p.Region = v
		}
	}

	return nil
}

func getOrCreate(profiles map[string]*models.Profile, name string) *models.Profile {
	if p, ok := profiles[name]; ok {
		return p
	}
	p := &models.Profile{Name: name}
	profiles[name] = p
	return p
}

func detectType(p *models.Profile) models.ProfileType {
	switch {
	case p.SSOStartURL != "":
		return models.ProfileTypeSSO
	case p.RoleARN != "":
		return models.ProfileTypeRole
	case p.HasCredentials:
		return models.ProfileTypeStatic
	default:
		return models.ProfileTypeUnknown
	}
}
