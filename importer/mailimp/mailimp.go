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

// Match represents a match.  When called, it returns the 'rest' of
// a string (e.g. the suffix when we matched on prefix)
type Match func() string

// LineMatcher knows how to match lines
type LineMatcher interface {
	Match(line string) Match
}

// PrefixMatcher knows how to match line prefixes
type PrefixMatcher []string

// Match checks to see if a line matches the configured prefixes.  If
// a match is found, we return the rest of the line (the non-matching
// portion).  Otherwise return nil
func (m PrefixMatcher) Match(line string) Match {
	for _, prefix := range m {
		if strings.HasPrefix(line, prefix) {
			return func() string {
				return strings.TrimRightFunc(
					strings.TrimPrefix(line, prefix), unicode.IsSpace)
			}
		}
	}
	return nil
}

// SuffixMatcher knows how to match line suffixes
type SuffixMatcher []string

// Match checks to see if a line matches the configured suffixes.  If
// a match is found, we return the rest of the line (the non-matching
// portion).  Otherwise return nil
func (m SuffixMatcher) Match(line string) Match {
	for _, suffix := range m {
		if strings.HasSuffix(line, suffix) {
			return func() string {
				return strings.TrimRightFunc(
					strings.TrimSuffix(line, suffix), unicode.IsSpace)
			}
		}
	}
	return nil
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
	var line string
	if i == -1 {
		line = split.remainingText
		split.remainingText = ""
	} else {
		line = split.remainingText[:i]
		split.remainingText = split.remainingText[i+1:]
	}
	for line != "" && line[len(line)-1] == '\r' {
		line = line[:len(line)-1]

	}
	return line, true
}
