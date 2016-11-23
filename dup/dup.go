package dup

import (
	"fmt"
	"time"

	ledgertools "github.com/ginabythebay/ledger-tools"
)

type key struct {
	account string
	amount  string
	date    string
}

func newKey(account, amount string, t time.Time) key {
	return key{account, amount, t.Format("2006/01/02")}
}

type Pair struct {
	First  *ledgertools.Posting
	Second *ledgertools.Posting
}

func (p Pair) CompilerText() string {
	return fmt.Sprintf("Apparent duplicate %s %s\n\tat %s %s(%s:%d)\n\tat %s %s(%s:%d)", p.First.AmountText(), p.First.Account, p.Second.Xact.DateText(), p.Second.Xact.Payee, p.Second.Xact.SrcFile, p.Second.BegLine, p.First.Xact.DateText(), p.First.Xact.Payee, p.First.Xact.SrcFile, p.First.BegLine)

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
		f.AllPairs = append(f.AllPairs, Pair{p, m})
	}

	f.m[k] = append(f.m[k], p)
}
