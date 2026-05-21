package config

type EmojiMap struct {
	Author    string `yaml:"author"`
	Branch    string `yaml:"branch"`
	State     string `yaml:"state"`
	Pipeline  string `yaml:"pipeline"`
	Approvals string `yaml:"approvals"`
	Labels    string `yaml:"labels"`
	Reviewers string `yaml:"reviewers"`
	Assignees string `yaml:"assignees"`
	Draft     string `yaml:"draft"`
}

type EmojiConfig struct {
	Enabled bool     `yaml:"enabled"`
	Icons   EmojiMap `yaml:"icons"`
}

func defaultEmojiMap() EmojiMap {
	return EmojiMap{
		Author:    "👤",
		Branch:    "🌿",
		State:     "🟢",
		Pipeline:  "⚙️",
		Approvals: "✅",
		Labels:    "🏷️",
		Reviewers: "👥",
		Assignees: "",
		Draft:     "📝",
	}
}

func DefaultEmojiConfig() EmojiConfig {
	return EmojiConfig{Enabled: true, Icons: defaultEmojiMap()}
}


// Resolve returns the effective emoji map: defaults merged with overrides when enabled,
// or all-empty when disabled.
func (config EmojiConfig) Resolve() EmojiMap {
	if !config.Enabled {
		return EmojiMap{}
	}

	defaults := defaultEmojiMap()
	icons := config.Icons

	if icons.Author == "" {
		icons.Author = defaults.Author
	}

	if icons.Branch == "" {
		icons.Branch = defaults.Branch
	}

	if icons.State == "" {
		icons.State = defaults.State
	}

	if icons.Pipeline == "" {
		icons.Pipeline = defaults.Pipeline
	}

	if icons.Approvals == "" {
		icons.Approvals = defaults.Approvals
	}

	if icons.Labels == "" {
		icons.Labels = defaults.Labels
	}

	if icons.Reviewers == "" {
		icons.Reviewers = defaults.Reviewers
	}

	if icons.Draft == "" {
		icons.Draft = defaults.Draft
	}

	return icons
}

