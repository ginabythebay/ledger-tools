// Package mailimp provides helpers for email importers
package mailimp

import (
	"fmt"
	"strings"
	"time"
	"unicode"
)

// PacificTz is the time location for America/Los_Angeles
var PacificTz *time.Location

func init() {
	var err error
	if PacificTz, err = time.LoadLocation("America/Los_Angeles"); err != nil {
		panic(fmt.Sprintf("Loading America/Los_Angeles: %+v", err))
	}
}

// PrefixMatcher knows how to match line prefixes
type PrefixMatcher []string

// Match checks to see if a line matches the configured prefixes.  If
// a match is found, we return the rest of the line (the non-matching
// portion).  Otherwise return ""
func (m PrefixMatcher) Match(line string) string {
	for _, prefix := range m {
		if strings.HasPrefix(line, prefix) {
			return strings.TrimRightFunc(strings.TrimPrefix(line, prefix), unicode.IsSpace)
		}
	}
	return ""
}

// LineSplitter splits some text into lines.  Using this instead of
// strings.Split() doubled our speed in a lyft benchmark.
type LineSplitter struct {
	remainingText string
}

func NewLineSplitter(remainingText string) *LineSplitter {
	return &LineSplitter{remainingText}
}

// Next returns the next line and true if there was a line.
func (split *LineSplitter) Next() (string, bool) {
	if len(split.remainingText) == 0 {
		return "", false
	}

	i := strings.IndexRune(split.remainingText, '\n')
	if i == -1 {
		line := split.remainingText
		split.remainingText = ""
		return line, true
	}
	line := split.remainingText[:i]
	split.remainingText = split.remainingText[i+1:]
	return line, true
}
