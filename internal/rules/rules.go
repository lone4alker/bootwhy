package rules

import "github.com/lone4alker/bootwhy/internal/parse"

type Evidence struct {
	Unit           string
	Entries        []parse.Entry
	PreviousUnit   string
	Overlapping    []string
	JournalMissing bool
	JournalPartial bool
}

type Check struct {
	Name   string                         `json:"name"`
	Weight int                            `json:"weight"`
	Stated bool                           `json:"stated"`
	Test   func(*Evidence) (bool, string) `json:"-"`
}

type Hypothesis struct {
	Name     string   `json:"name"`
	Explain  string   `json:"explain"`
	Evidence []Check  `json:"-"`
	Option   string   `json:"option,omitempty"`
	Undo     string   `json:"undo,omitempty"`
	Risk     string   `json:"risk"`
	Source   string   `json:"source"`
	Units    []string `json:"units,omitempty"`
	Distros  []string `json:"distros,omitempty"`
}
