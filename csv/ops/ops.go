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

func ReplaceHeader(header []string) Mutator {
	return func(l *Line) error {
		if l.LineNo == 1 {
			l.Record = header
		}
		return nil
	}
}

func MoveAndNegateIfPresent(from int, to int) Mutator {
	return func(l *Line) error {
		if l.LineNo != 1 {
			value := l.Record[from]
			if value != "" {
				l.Record[to] = negate(value)
				l.Record[from] = ""
			}
		}
		return nil
	}
}

func Negate(i int) Mutator {
	return func(l *Line) error {
		if l.LineNo != 1 {
			value := l.Record[i]
			l.Record[i] = negate(value)
		}
		return nil
	}
}

func EnsureDollars(i int) Mutator {
	return func(l *Line) error {
		if l.LineNo != 1 {
			value := l.Record[i]
			if value != "" && (!strings.HasPrefix(value, "$")) {
				l.Record[i] = "$" + value
			}
		}
		return nil
	}
}

func StripCommas(i int) Mutator {
	return func(l *Line) error {
		if l.LineNo != 1 {
			value := l.Record[i]
			if strings.Contains(value, ",") {
				l.Record[i] = strings.Replace(value, ",", "", -1)
			}
		}
		return nil
	}
}

func StripNewlines(l *Line) error {
	for i, c := range l.Record {
		if strings.Contains(c, "\n") {
			l.Record[i] = strings.Replace(c, "\n", "", -1)
		}
	}
	return nil
}
