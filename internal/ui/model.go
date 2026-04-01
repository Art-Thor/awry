package ui

import (
	"fmt"
	"sort"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/Art-Thor/awry/internal/awsconfig"
	"github.com/Art-Thor/awry/internal/config"
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
	favorites      map[string]struct{}
	configPath     string
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
		favorites:      map[string]struct{}{},
	}

	cfg, path, err := config.Load()
	if err != nil {
		return Model{}, fmt.Errorf("loading config: %w", err)
	}
	m.configPath = path
	for _, favorite := range cfg.Favorites {
		m.favorites[favorite] = struct{}{}
	}

	m.pinActiveToTop()
	m.pinFavoritesAfterActive()
	m.filtered = m.profiles
	return m, nil
}

func (m Model) isFavorite(name string) bool {
	_, ok := m.favorites[name]
	return ok
}

func (m *Model) toggleFavorite(profileName string) error {
	if m.favorites == nil {
		m.favorites = map[string]struct{}{}
	}

	if _, ok := m.favorites[profileName]; ok {
		delete(m.favorites, profileName)
	} else {
		m.favorites[profileName] = struct{}{}
	}

	if err := m.saveConfig(); err != nil {
		return err
	}

	m.pinActiveToTop()
	m.pinFavoritesAfterActive()
	m.applyFilter()
	m.restoreCursor(profileName)
	return nil
}

func (m Model) saveConfig() error {
	keys := make([]string, 0, len(m.favorites))
	for name := range m.favorites {
		keys = append(keys, name)
	}
	sort.Strings(keys)

	return config.Save(config.Config{Favorites: keys}, m.configPath)
}

func (m *Model) restoreCursor(profileName string) {
	for i, profile := range m.filtered {
		if profile.Name == profileName {
			m.cursor = i
			return
		}
	}
	if len(m.filtered) == 0 {
		m.cursor = 0
	}
}

func (m Model) Init() tea.Cmd {
	return m.refreshActiveRuntimeCmd()
}
