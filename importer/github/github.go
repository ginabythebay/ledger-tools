package github

import (
	"strings"
	"time"

	ledgertools "github.com/ginabythebay/ledger-tools"
	"github.com/ginabythebay/ledger-tools/importer"
	"github.com/ginabythebay/ledger-tools/importer/mailimp"
	"github.com/pkg/errors"
)

const (
	// From is the email address we expect to get invoices from
	From        = "support@github.com"
	fromMatcher = "<" + From + ">"

	// SubjectPrefix is the common prefix we see in invoices
	SubjectPrefix = "[GitHub] Payment Receipt for"
)

var amountMatcher = mailimp.PrefixMatcher([]string{"Amount: USD "})
var chargeMatcher = mailimp.PrefixMatcher([]string{"Charged to:"})
var transactionMatcher = mailimp.PrefixMatcher([]string{"Transaction ID:"})

var commentMatchers = []mailimp.PrefixMatcher{
	{"GITHUB RECEIPT - PERSONAL SUBSCRIPTION"},
	{"For service through:"},
}

const payee = "Github"

// ImportMessage imports an email message.  Returns nil if msg does
// not appear to be a github invoice.  Returns an error if it does
// appear to be a github invoice, but we have trouble parsing it.
func ImportMessage(msg ledgertools.Message) (*importer.Parsed, error) {
	if !strings.Contains(msg.From, fromMatcher) {
		return nil, nil
	}
	if !strings.HasPrefix(msg.Subject, SubjectPrefix) {
		return nil, nil
	}

	date, err := time.Parse(time.RFC1123Z, msg.Date)
	if err != nil {
		return nil, errors.Wrapf(err, "Parsing date in %v", msg)
	}
	date = date.In(mailimp.PacificTz)

	// build all these up by looking at the message text
	var checkNumber string
	var comments = make([]string, 0, len(commentMatchers))
	var amount string
	var instrument string

	splitter := mailimp.NewLineSplitter(msg.TextPlain)
	for {
		line, ok := splitter.Next()
		if !ok {
			break
		}
		for _, m := range commentMatchers {
			if m.Match(line) != nil {
				comments = append(comments, line)
				continue
			}
		}

		if match := transactionMatcher.Match(line); match != nil {
			checkNumber = strings.TrimSpace(match())
			continue
		}

		if match := amountMatcher.Match(line); match != nil {
			// rest looks like "$7.00*"
			amount = strings.TrimRight(match(), "*")
			continue
		}

		if match := chargeMatcher.Match(line); match != nil {
			instrument = strings.TrimSpace(match())
			continue
		}
	}

	// Verify that we found everything we were expecting
	if checkNumber == "" {
		return nil, errors.Errorf("Missing transaction id line in %q", msg.TextPlain)
	}
	if len(comments) != len(commentMatchers) {
		return nil, errors.Errorf("Missing comments.  Found %q in %q.  Expected to find lines starting with %q", comments, msg.TextPlain, commentMatchers)
	}
	if amount == "" {
		return nil, errors.Errorf("Missing amount line in %q", msg.TextPlain)
	}
	if instrument == "" {
		return nil, errors.Errorf("Missing Charged to line in %q", msg.TextPlain)
	}

	return importer.NewParsed(
		date,
		checkNumber,
		payee,
		comments,
		amount,
		instrument), nil
}
