package dup

import (
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

// Finder tracks postings and looks for potential duplicates, based on the amount and the date.
type Finder struct {
	// # days to use when looking for matches.  0 means only look for matches on exactly the same day
	Days int

	m map[key][]*ledgertools.Posting
}

// NewFinder creates a new Finder.
func NewFinder(days int) *Finder {
	return &Finder{
		days,
		make(map[key][]*ledgertools.Posting),
	}
}

// Add adds p and returns any existing postings that have the same amount and
// are within the configured number of days.
func (f *Finder) Add(p *ledgertools.Posting) []*ledgertools.Posting {
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

	f.m[k] = append(f.m[k], p)

	return matches
}
