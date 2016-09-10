package ops

import (
	"strings"
)

// negate returns the negative version of the input.  We avoid any
// conversion to a number her and just treat it as a raw string.
func negate(s string) string {
	if strings.HasPrefix(s, "-") {
		return s[1:]
	}
	return "-" + s
}

// Mutator is a function that knows how to mutate a Line
type Mutator func(l *Line) error

// Line represents a line we read from a csv file
type Line struct {
	LineNo int
	Record []string
}

// NewLine creates a Line structure
func NewLine(lineNo int, record []string) *Line {
	return &Line{lineNo, record}
}

func (l *Line) ReplaceHeader(header []string) {
	if l.LineNo == 1 {
		l.Record = header
	}
}

func (l *Line) MoveAndNegateIfPresent(from int, to int) {
	if l.LineNo != 1 {
		value := l.Record[from]
		if value != "" {
			l.Record[to] = negate(value)
			l.Record[from] = ""
		}
	}
}

func (l *Line) Negate(i int) {
	if l.LineNo != 1 {
		value := l.Record[i]
		l.Record[i] = negate(value)
	}
}

func (l *Line) EnsureDollars(i int) {
	if l.LineNo != 1 {
		value := l.Record[i]
		if value != "" && (!strings.HasPrefix(value, "$")) {
			l.Record[i] = "$" + value
		}
	}
}

func (l *Line) StripCommas(i int) error {
	if l.LineNo != 1 {
		value := l.Record[i]
		if strings.Contains(value, ",") {
			l.Record[i] = strings.Replace(value, ",", "", -1)
		}
	}
	return nil
}

func StripNewlines(l *Line) error {
	for i, c := range l.Record {
		if strings.Contains(c, "\n") {
			l.Record[i] = strings.Replace(c, "\n", "", -1)
		}
	}
	return nil
}
