package ui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/Art-Thor/awry/internal/awsconfig"
	"github.com/Art-Thor/awry/internal/identity"
	"github.com/Art-Thor/awry/internal/session"
	"github.com/Art-Thor/awry/pkg/models"
)

var (
	lookupIdentity = identity.Lookup
	resolveSession = session.ResolveProfile
)

// Model is the top-level Bubble Tea model.
type Model struct {
	profiles       []models.Profile
	filtered       []models.Profile
	cursor         int
	currentProfile string
	searchQuery    string
	searching      bool
	selected       *models.Profile
	width          int
	height         int
	quitting       bool
	helpVisible    bool
	identity       *identity.Identity
	identityErr    error
	sessionInfo    *session.Info
	sessionErr     error
}

// SelectedProfile returns the profile the user chose (nil if none).
func (m Model) SelectedProfile() *models.Profile {
	return m.selected
}

// New creates the initial model.
func New() (Model, error) {
	profiles, err := awsconfig.LoadProfiles()
	if err != nil {
		return Model{}, fmt.Errorf("loading profiles: %w", err)
	}

	m := Model{
		profiles:       profiles,
		currentProfile: awsconfig.CurrentProfile(),
	}
	m.pinActiveToTop()
	m.filtered = m.profiles
	return m, nil
}

func (m Model) Init() tea.Cmd {
	return m.refreshActiveRuntimeCmd()
}
