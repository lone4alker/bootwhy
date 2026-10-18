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
	Name   string
	Weight int
	Stated bool
	Test   func(*Evidence) (bool, string)
}

type Hypothesis struct {
	Name, Explain  string
	Evidence       []Check
	Option, Undo   string
	Risk, Source   string
	Units, Distros []string
}
