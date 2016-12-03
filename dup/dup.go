package dup

import (
	"encoding/xml"
	"fmt"
	"io"
	"strings"
	"time"

	ledgertools "github.com/ginabythebay/ledger-tools"
)

const (
	suppressAmountDuplicates = "SuppressAmountDuplicates:"
	suppressCodeDuplicates   = "SuppressCodeDuplicates:"
)

type amountKey struct {
	account string
	amount  string
	date    string
}

func newAmountKey(account, amount string, t time.Time) amountKey {
	return amountKey{account, amount, t.Format("2006/01/02")}
}

func suppressedDates(suppressPrefix string, notes []string) []string {
	var dates []string
	for _, line := range notes {
		split := strings.SplitAfterN(line, suppressPrefix, 2)
		if len(split) != 2 {
			continue
		}
		someCandidates := strings.Split(split[1], ",")
		for _, c := range someCandidates {
			c = strings.TrimSpace(c)
			if len(c) < 10 {
				// dates are 10 characters long.
				break
			}
			_, err := time.Parse("2006/01/02", c[:10])
			if err == nil {
				dates = append(dates, c[:10])
			}
			if len(c) > 10 {
				// sometimes there is glop after a valid date.  If we hit that case, just bail out until the next line
				break
			}
		}
	}

	return dates
}

func isDateSuppressed(date string, suppressPrefix string, notes []string) bool {
	for _, d := range suppressedDates(suppressPrefix, notes) {
		if date == d {
			return true
		}
	}
	return false
}

type duplicate interface {
	compilerText() (string, error)
	isSuppressed() bool
	accumXMLErrors(accum map[string]*file)
}

// Finder tracks postings and looks for potential duplicates, based on the amount and the date.
type Finder struct {
	// # days to use when looking for matches.  0 means only look for matches on exactly the same day
	Days int

	codeMap   map[string][]*ledgertools.Transaction
	amountMap map[amountKey][]*ledgertools.Posting

	allDuplicates []duplicate
}

// NewFinder creates a new Finder.
func NewFinder(days int) *Finder {
	return &Finder{
		Days:      days,
		codeMap:   make(map[string][]*ledgertools.Transaction),
		amountMap: make(map[amountKey][]*ledgertools.Posting),
	}
}

// Add adds t and its postings and tracks any existing postings that
// have the same amount and are within the configured number of days.
func (f *Finder) Add(t *ledgertools.Transaction) {
	f.addCodeXact(t)
	for _, p := range t.Postings {
		f.addAmountPosting(p)
	}
}

func (f *Finder) addCodeXact(t *ledgertools.Transaction) {
	if t.Code == "" {
		return
	}
	code := t.Code
	if !strings.HasPrefix(code, "#") {
		code = "#" + code
	}
	for _, m := range f.codeMap[code] {
		cp := newCodePair(m, t)
		if !cp.isSuppressed() {
			f.allDuplicates = append(f.allDuplicates, cp)
		}
	}
	f.codeMap[code] = append(f.codeMap[code], t)
}

func (f *Finder) addAmountPosting(p *ledgertools.Posting) {
	var matches []*ledgertools.Posting
	t := p.Xact.Date
	amount := p.AmountText()
	k := newAmountKey(p.Account, amount, t)

	if f.Days >= 0 {
		matches = append(matches, f.amountMap[k]...)
	}
	for i := 1; i <= f.Days; i++ {
		before := t.AddDate(0, 0, -i)
		matches = append(matches, f.amountMap[newAmountKey(p.Account, amount, before)]...)

		after := t.AddDate(0, 0, i)
		matches = append(matches, f.amountMap[newAmountKey(p.Account, amount, after)]...)
	}

	for _, m := range matches {
		p := amountPair{m, p}
		if !p.isSuppressed() {
			f.allDuplicates = append(f.allDuplicates, p)
		}
	}

	f.amountMap[k] = append(f.amountMap[k], p)
}

// WriteJavacStyle writes javac-style output for all duplicates found.
func (f *Finder) WriteJavacStyle(w io.Writer) error {
	matchCount := 0
	for _, d := range f.allDuplicates {
		s, err := d.compilerText()
		if err != nil {
			return err
		}
		matchCount++

		fmt.Fprintln(w, s)
	}

	fmt.Fprintf(w, "\n %d potential duplicates found\n", matchCount)
	return nil
}

// WriteCheckStyle writes checkstyle (xml) output for all duplicates found.
func (f *Finder) WriteCheckStyle(w io.Writer) error {
	accum := map[string]*file{}
	for _, d := range f.allDuplicates {
		d.accumXMLErrors(accum)
	}

	var allFiles []*file
	for _, f := range accum {
		allFiles = append(allFiles, f)
	}
	cs := checkstyle{
		Version: version,
		Files:   allFiles,
	}
	enc := xml.NewEncoder(w)
	enc.Indent("", "  ")
	return enc.Encode(&cs)
}

const (
	version  = "7.2"
	severity = "warning"
	source   = "dupdetector"
)

type checkstyle struct {
	XMLName xml.Name `xml:"checkstyle"`
	Version string   `xml:"version,attr"`
	Files   []*file
}

type file struct {
	XMLName xml.Name `xml:"file"`
	Name    string   `xml:"name,attr"`
	Errors  []xmlError
}

type xmlError struct {
	XMLName  xml.Name `xml:"error"`
	Line     int      `xml:"line,attr"`
	Severity string   `xml:"severity,attr"`
	Message  string   `xml:"message,attr"`
	Source   string   `xml:"source,attr"`
}
