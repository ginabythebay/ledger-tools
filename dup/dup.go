package dup

import (
	"bytes"
	"strings"
	"text/template"
	"time"

	ledgertools "github.com/ginabythebay/ledger-tools"
)

var compilerTemplate = template.Must(template.New("CompilerOutput").Parse(strings.TrimSpace(`
Apparent duplicate {{.One.AmountText}} {{.One.Account}}
	at {{.One.Xact.DateText}} {{.One.Xact.Payee}} ({{.One.Xact.SrcFile}}:{{.One.BegLine}})
	at {{.Two.Xact.DateText}} {{.Two.Xact.Payee}} ({{.Two.Xact.SrcFile}}:{{.Two.BegLine}})
`)))

type key struct {
	account string
	amount  string
	date    string
}

func newKey(account, amount string, t time.Time) key {
	return key{account, amount, t.Format("2006/01/02")}
}

func suppressedDates(notes []string) []string {
	var dates []string
	for _, line := range notes {
		split := strings.SplitAfterN(line, "SuppressDuplicates:", 2)
		if len(split) != 2 {
			continue
		}
		someCandidates := strings.Split(split[1], ",")
		for _, c := range someCandidates {
			c = strings.TrimSpace(c)
			_, err := time.Parse("2006/01/02", c)
			if err == nil {
				dates = append(dates, c)
			}
		}
	}
	return dates
}

func isDateSuppressed(date string, notes []string) bool {
	for _, d := range suppressedDates(notes) {
		if date == d {
			return true
		}
	}
	return false
}

type Pair struct {
	One *ledgertools.Posting
	Two *ledgertools.Posting
}

func (p Pair) CompilerText() (string, error) {
	var buf bytes.Buffer
	err := compilerTemplate.Execute(&buf, p)
	return buf.String(), err
}

func (p Pair) IsSuppressed() bool {
	return isDateSuppressed(p.One.Xact.DateText(), p.Two.Notes) || isDateSuppressed(p.Two.Xact.DateText(), p.One.Notes)
}

// Finder tracks postings and looks for potential duplicates, based on the amount and the date.
type Finder struct {
	// # days to use when looking for matches.  0 means only look for matches on exactly the same day
	Days int

	m map[key][]*ledgertools.Posting

	AllPairs []Pair
}

// NewFinder creates a new Finder.
func NewFinder(days int) *Finder {
	return &Finder{
		Days: days,
		m:    make(map[key][]*ledgertools.Posting),
	}
}

// Add adds p and tracks any existing postings that have the same amount and
// are within the configured number of days.
func (f *Finder) Add(p *ledgertools.Posting) {
	var matches []*ledgertools.Posting
	t := p.Xact.Date
	amount := p.AmountText()
	k := newKey(p.Account, amount, t)

	if f.Days >= 0 {
		matches = append(matches, f.m[k]...)
	}
	for i := 1; i <= f.Days; i++ {
		before := t.AddDate(0, 0, -i)
		matches = append(matches, f.m[newKey(p.Account, amount, before)]...)

		after := t.AddDate(0, 0, i)
		matches = append(matches, f.m[newKey(p.Account, amount, after)]...)
	}

	for _, m := range matches {
		f.AllPairs = append(f.AllPairs, Pair{m, p})
	}

	f.m[k] = append(f.m[k], p)
}
