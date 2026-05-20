//nolint:mnd // Config defaults and file modes are conventional constants.
package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"gopkg.in/yaml.v3"
)

const (
	FileOverrideEnv = "GITLAB_TUI_CONFIG_FILE"
	DefaultTokenEnv = "GITLAB_TOKEN"
)

type Config struct {
	DefaultAccount       string          `yaml:"default_account"`
	Accounts             []Account       `yaml:"accounts"`
	RecentProjectsLimit  int             `yaml:"recent_projects_limit"`
	RecentProjectHistory []RecentProject `yaml:"recent_projects,omitempty"`
	Emoji                EmojiConfig     `yaml:"emoji"`
	Layout               LayoutConfig    `yaml:"layout"`
}

// LayoutConfig holds UI layout settings.
type LayoutConfig struct {
	LeftPanelWidth LeftPanelWidthConfig `yaml:"left_panel_width"`
}

// LeftPanelWidthConfig sets the left panel width as a percentage of the terminal width
// for each navigation level. Valid range: 10–90. Zero values fall back to the default (35).
type LeftPanelWidthConfig struct {
	Sections   int `yaml:"sections"`
	EntityList int `yaml:"entity_list"`
	Detail     int `yaml:"detail"`
	FileDiff   int `yaml:"file_diff"`
}

func DefaultLayoutConfig() LayoutConfig {
	return LayoutConfig{
		LeftPanelWidth: LeftPanelWidthConfig{
			Sections:   35,
			EntityList: 35,
			Detail:     35,
			FileDiff:   35,
		},
	}
}

type Account struct {
	ID       string `yaml:"id"`
	Host     string `yaml:"host"`
	TokenEnv string `yaml:"token_env"`
}

type RecentProject struct {
	Account    string    `yaml:"account"`
	Path       string    `yaml:"path"`
	LastUsedAt time.Time `yaml:"last_used_at"`
}

type Paths struct {
	Env []string
}

func Default() Config {
	return Config{
		DefaultAccount:      "default",
		RecentProjectsLimit: 10,
		Accounts: []Account{{
			ID:       "default",
			Host:     "https://gitlab.com",
			TokenEnv: DefaultTokenEnv,
		}},
		Emoji:  DefaultEmojiConfig(),
		Layout: DefaultLayoutConfig(),
	}
}

func (p Paths) Path() (string, error) {
	if override := getenv(p.Env, FileOverrideEnv); override != "" {
		return override, nil
	}

	if xdg := getenv(p.Env, "XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "gitlab-tui", "config.yaml"), nil
	}

	home := getenv(p.Env, "HOME")
	if home == "" {
		var err error

		home, err = os.UserHomeDir()
		if err != nil || home == "" {
			return "", errors.New("cannot resolve config path: HOME is not set")
		}
	}

	return filepath.Join(home, ".config", "gitlab-tui", "config.yaml"), nil
}

func Load(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, err
	}

	cfg := Config{RecentProjectsLimit: 10}
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return Config{}, err
	}

	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func Save(path string, cfg Config) error {
	if err := cfg.Validate(); err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0o600)
}

func Init(path string, cfg Config) error {
	if _, err := os.Stat(path); err == nil {
		return fmt.Errorf("config already exists: %s", path)
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}

	return Save(path, cfg)
}

func (c Config) Validate() error {
	if c.DefaultAccount == "" {
		return errors.New("default_account is required")
	}

	if len(c.Accounts) == 0 {
		return errors.New("at least one account is required")
	}

	seen := map[string]bool{}
	defaultFound := false

	for _, account := range c.Accounts {
		if account.ID == "" {
			return errors.New("account id is required")
		}

		if seen[account.ID] {
			return fmt.Errorf("duplicate account id: %s", account.ID)
		}

		seen[account.ID] = true

		if account.ID == c.DefaultAccount {
			defaultFound = true
		}

		if account.Host == "" {
			return fmt.Errorf("account %s host is required", account.ID)
		}

		if account.TokenEnv == "" {
			return fmt.Errorf("account %s token_env is required", account.ID)
		}
	}

	if !defaultFound {
		return fmt.Errorf("default account %q does not exist", c.DefaultAccount)
	}

	w := c.Layout.LeftPanelWidth
	for name, v := range map[string]int{
		"sections": w.Sections, "entity_list": w.EntityList,
		"detail": w.Detail, "file_diff": w.FileDiff,
	} {
		if v != 0 && (v < 10 || v > 90) {
			return fmt.Errorf("layout.left_panel_width.%s must be between 10 and 90 (got %d)", name, v)
		}
	}

	return nil
}

func (c Config) RecentProjects() []RecentProject {
	if c.RecentProjectsLimit == 0 {
		return nil
	}

	projects := append([]RecentProject(nil), c.RecentProjectHistory...)
	sortRecentProjects(projects)

	if c.RecentProjectsLimit < len(projects) {
		projects = projects[:c.RecentProjectsLimit]
	}

	return projects
}

func (c Config) RecentProjectsForAccount(accountID string) []RecentProject {
	projects := make([]RecentProject, 0, len(c.RecentProjectHistory))

	for _, project := range c.RecentProjectHistory {
		if project.Account == accountID {
			projects = append(projects, project)
		}
	}

	sortRecentProjects(projects)

	return projects
}

func (c *Config) RememberProject(project RecentProject) {
	for i, existing := range c.RecentProjectHistory {
		if existing.Account == project.Account && existing.Path == project.Path {
			c.RecentProjectHistory[i].LastUsedAt = project.LastUsedAt
			return
		}
	}

	c.RecentProjectHistory = append(c.RecentProjectHistory, project)
}

func sortRecentProjects(projects []RecentProject) {
	sort.SliceStable(projects, func(i int, j int) bool {
		return projects[i].LastUsedAt.After(projects[j].LastUsedAt)
	})
}

func (c Config) Account(id string) (Account, bool) {
	for _, account := range c.Accounts {
		if account.ID == id {
			return account, true
		}
	}

	return Account{}, false
}

func (a Account) Token(env []string) (string, error) {
	token := getenv(env, a.TokenEnv)
	if token == "" {
		return "", fmt.Errorf("token env %s is not set", a.TokenEnv)
	}

	return token, nil
}

func getenv(env []string, key string) string {
	if env == nil {
		return os.Getenv(key)
	}

	prefix := key + "="
	for _, entry := range env {
		if len(entry) >= len(prefix) && entry[:len(prefix)] == prefix {
			return entry[len(prefix):]
		}
	}

	return ""
}
