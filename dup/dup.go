package dup

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"strings"
	"text/template"
	"time"

	ledgertools "github.com/ginabythebay/ledger-tools"
)

var compilerTemplate = template.Must(template.New("CompilerOutput").Parse(strings.TrimSpace(`
Possible duplicate {{.One.AmountText}} {{.One.Account}}
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
		split := strings.SplitAfterN(line, "SuppressAmountDuplicates:", 2)
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

func isDateSuppressed(date string, notes []string) bool {
	for _, d := range suppressedDates(notes) {
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

type Pair struct {
	One *ledgertools.Posting
	Two *ledgertools.Posting
}

func (p Pair) compilerText() (string, error) {
	var buf bytes.Buffer
	err := compilerTemplate.Execute(&buf, p)
	return buf.String(), err
}

func (p Pair) isSuppressed() bool {
	return isDateSuppressed(p.One.Xact.DateText(), p.Two.Notes) || isDateSuppressed(p.Two.Xact.DateText(), p.One.Notes)
}

func (p Pair) accumXMLErrors(accum map[string]*file) {
	addXMLError(accum, p.One, p.Two)
	addXMLError(accum, p.Two, p.One)
}

func addXMLError(accum map[string]*file, p, other *ledgertools.Posting) {
	f, ok := accum[p.Xact.SrcFile]
	if !ok {
		f = &file{
			Name: p.Xact.SrcFile,
		}
		accum[p.Xact.SrcFile] = f
	}
	msg := fmt.Sprintf("Possible duplicate of %s %s %s at %s:%d", other.Xact.DateText(), other.AmountText(), other.Account, other.Xact.SrcFile, other.BegLine)
	f.Errors = append(f.Errors, xmlError{
		Line:     p.BegLine,
		Severity: severity,
		Message:  msg,
		Source:   source,
	})
}

// Finder tracks postings and looks for potential duplicates, based on the amount and the date.
type Finder struct {
	// # days to use when looking for matches.  0 means only look for matches on exactly the same day
	Days int

	m map[key][]*ledgertools.Posting

	allDuplicates []duplicate
}

// NewFinder creates a new Finder.
func NewFinder(days int) *Finder {
	return &Finder{
		Days: days,
		m:    make(map[key][]*ledgertools.Posting),
	}
}

// Add adds t and its postings and tracks any existing postings that
// have the same amount and are within the configured number of days.
func (f *Finder) Add(t *ledgertools.Transaction) {
	for _, p := range t.Postings {
		f.addPosting(p)
	}
}

func (f *Finder) addPosting(p *ledgertools.Posting) {
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
		p := Pair{m, p}
		if !p.isSuppressed() {
			f.allDuplicates = append(f.allDuplicates, p)
		}
	}

	f.m[k] = append(f.m[k], p)
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
