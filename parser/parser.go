package parser

import (
	"bufio"
	"fmt"
	"io"
	"regexp"
	"strings"
	"time"
)

var headerRE = regexp.MustCompile("(\\S+)\\s+([^;]+)(;.*)?")

var postingRE = regexp.MustCompile("(.*?)\\s{2,}(.*)")

// Posting represents an account the movement of value into or out of it.
type Posting struct {
	Account string
	Amount  string
}

func (p Posting) String() string {
	return fmt.Sprintf("\t%s\t%s", p.Account, p.Amount)
}

// Transaction represents the movement of value between one or more accounts
type Transaction struct {
	Line     int
	Date     time.Time
	Payee    string
	Comment  string
	Postings []Posting
}

func (t Transaction) String() string {
	lines := make([]string, 0, len(t.Postings)+1)

	date := t.Date.Format("2006/01/02")
	header := fmt.Sprintf("%s\t%s%s", date, t.Payee, t.Comment)
	lines = append(lines, header)

	for _, p := range t.Postings {
		lines = append(lines, p.String())
	}

	return strings.Join(lines, "\n")
}

// Validate checks to see if we have what appears to be a full transaction
func (t *Transaction) Validate() (err error) {
	if len(t.Postings) < 2 {
		return fmt.Errorf("%d: Need at least 2 postings and only found %d postings for transaction that starts %q %q %q", t.Line, len(t.Postings), t.Date, t.Payee, t.Comment)
	}
	return nil
}

// ParseLedger parses a Register file.  It is expected to be in exactly the
// format that is output by reckon (https://github.com/cantino/reckon)
func ParseLedger(reader io.Reader) (ledger []Transaction, err error) {
	ledger = make([]Transaction, 0, 100)
	lineNo := 0

	var next *Transaction
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		lineNo++
		switch {
		case next == nil:
			if line == "" {
				continue
			}
			if next, err = parseHeader(lineNo, line); err != nil {
				return nil, err
			}
		default:
			if line == "" {
				if ledger, err = addIfValid(*next, ledger); err != nil {
					return nil, err
				}
				next = nil
			} else {
				p, err := parsePosting(lineNo, line)
				if err != nil {
					return nil, err
				}

				next.Postings = append(next.Postings, p)
			}
		}
	}
	if next == nil {
		// nothing to do, we either added the last transaction to the ledger or there were none
	} else {
		if ledger, err = addIfValid(*next, ledger); err != nil {
			return nil, err
		}
	}
	return ledger, nil
}

func addIfValid(next Transaction, in []Transaction) (out []Transaction, err error) {
	if err = next.Validate(); err != nil {
		return nil, err
	}
	return append(in, next), nil
}

func parseHeader(lineNo int, line string) (t *Transaction, err error) {
	matches := headerRE.FindStringSubmatch(line)
	if matches == nil || len(matches) < 3 {
		return nil, fmt.Errorf("%d: Unexpected Transaction Header %q", lineNo, line)
	}
	dateStr := matches[1]
	payee := matches[2]
	comment := ""
	if len(matches) > 3 {
		comment = matches[3]
	}

	date, err := time.Parse("2006/01/02", dateStr)
	if err != nil {
		return nil, fmt.Errorf("%d: %s", lineNo, err)
	}

	t = &Transaction{
		lineNo,
		date,
		payee,
		comment,
		[]Posting{},
	}
	return t, nil
}

func parsePosting(lineNo int, line string) (p Posting, err error) {
	matches := postingRE.FindStringSubmatch(line)
	if matches == nil || len(matches) != 3 {
		return Posting{}, fmt.Errorf("%d: Unexpected Posting %q", lineNo, line)
	}
	account := matches[1]
	value := matches[2]
	return Posting{account, value}, nil
}
