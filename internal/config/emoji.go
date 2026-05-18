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

// Resolve returns the effective emoji map: defaults merged with overrides when enabled,
// or all-empty when disabled.
func (e EmojiConfig) Resolve() EmojiMap {
	if !e.Enabled {
		return EmojiMap{}
	}

	d := defaultEmojiMap()
	m := e.Icons

	if m.Author == "" {
		m.Author = d.Author
	}

	if m.Branch == "" {
		m.Branch = d.Branch
	}

	if m.State == "" {
		m.State = d.State
	}

	if m.Pipeline == "" {
		m.Pipeline = d.Pipeline
	}

	if m.Approvals == "" {
		m.Approvals = d.Approvals
	}

	if m.Labels == "" {
		m.Labels = d.Labels
	}

	if m.Reviewers == "" {
		m.Reviewers = d.Reviewers
	}

	if m.Draft == "" {
		m.Draft = d.Draft
	}

	return m
}

func DefaultEmojiConfig() EmojiConfig {
	return EmojiConfig{Enabled: true}
}
