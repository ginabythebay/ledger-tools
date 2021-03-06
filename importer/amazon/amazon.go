package amazon

import (
	"strings"
	"time"

	ledgertools "github.com/ginabythebay/ledger-tools"
	"github.com/ginabythebay/ledger-tools/gmail"
	"github.com/ginabythebay/ledger-tools/importer"
	"github.com/ginabythebay/ledger-tools/importer/mailimp"
	"github.com/pkg/errors"
)

const (
	from        = "ship-confirm@amazon.com"
	fromMatcher = "<" + from + ">"

	defaultPayment = "AmazonDefaultPayment"
	payee          = "Amazon"
)

var subjectPrefixes = []string{
	"Your Amazon.com order has shipped",  // 'standard' orders
	"Your AmazonSmile order has shipped", // smile orders
	"Your Amazon.com order of",           // book orders
}

func queries() []gmail.QuerySet {
	var result []gmail.QuerySet
	for _, sp := range subjectPrefixes {
		qs := []gmail.QueryOption{gmail.QueryFrom(from), gmail.QuerySubject(sp)}
		result = append(result, qs)
	}
	return result
}

// GmailImporter knows how to fetch and parse amazon emails.
var GmailImporter = importer.NewGmailImporter(
	queries(),
	[]importer.Parser{
		importMessage,
	},
)

var orderMatcher = mailimp.PrefixMatcher([]string{"Order #"})
var totalMatcher = mailimp.PrefixMatcher([]string{"    Shipment Total: "})
var totalPrefixMatcher = mailimp.PrefixMatcher([]string{"Shipment total:"})

var commentMatchers = []mailimp.LineMatcher{
	mailimp.SuffixMatcher([]string{" has shipped.", " have shipped."}),
	mailimp.PrefixMatcher([]string{"Order #"}),
}

// If we see any of these, then we want to capture the following line
// as a comment.
var commentPrefixes = []mailimp.LineMatcher{
	mailimp.PrefixMatcher([]string{
		"Track your package at:",
		"Track your package:",
		"Why tracking information may not be available?:",
	}),
	mailimp.PrefixMatcher([]string{"On the way:"}),
}

// importMessage imports an email message.  Returns nil if msg does
// not appear to be a amazon shipment summary.  Returns an error if it
// does appear to be a amazon shipment summary, but we have trouble
// parsing it.  An example valid email would be:
//
//   Amazon.com Shipping Confirmation
//   http://www.amazon.com/someconfirmationlink
//
//   --------------------------------------------------------------------
//   Hello Gina White,
//
//   "Some Item..." and one other item have shipped.
//
//   Details
//   Order #123-1234567-1234567
//
//   Arriving:
//       Monday, September 26
//
//   Track your package at:
//   	https://www.amazon.com/sometrackinglink
//
//   Shipped to:
//       Gina White
//       Some address...
//
//
//   ====================================================================
//
//       Total Before Tax: $26.38
//       Tax Collected: $1.64
//       Shipment Total: $28.02
//
//   ====================================================================
//
//   View or manage your order in Your Orders:
//   https://www.amazon.com/someorderlink
//
//   We hope to see you again soon.<br/>
//   Amazon.com
//
//   --------------------------------------------------------------------
func importMessage(msg ledgertools.Message) (*importer.Parsed, error) {
	if !strings.Contains(msg.From, fromMatcher) {
		return nil, nil
	}
	subjectMatch := false
	for _, sp := range subjectPrefixes {
		if strings.HasPrefix(msg.Subject, sp) {
			subjectMatch = true
			break
		}
	}
	if !subjectMatch {
		return nil, nil
	}

	date, err := time.Parse("Mon, 2 Jan 2006 15:04:05 -0700", msg.Date)
	if err != nil {
		return nil, errors.Wrapf(err, "Parsing date in %v", msg)
	}
	date = date.In(mailimp.PacificTz)

	// build all these up by looking at the message text
	var checkNumber string
	var comments = make([]string, 0, len(commentMatchers)+len(commentPrefixes))
	var amount string

	splitter := mailimp.NewLineSplitter(msg.TextPlain)
	var lastLine string
	for {
		line, ok := splitter.Next()
		if !ok {
			break
		}

		for _, m := range commentMatchers {
			if m.Match(line) != nil {
				comments = append(comments, line)
			}
		}
		if lastLine != "" {
			for _, pre := range commentPrefixes {
				if pre.Match(lastLine) != nil {
					comments = append(comments, strings.TrimSpace(line))
				}
			}

			if totalPrefixMatcher.Match(lastLine) != nil {
				amount = line
			}
		}

		if match := orderMatcher.Match(line); match != nil {
			checkNumber = match()
			continue
		}

		if match := totalMatcher.Match(line); match != nil {
			amount = match()
			continue
		}

		lastLine = line
	}

	// Verify that we found everything we were expecting
	if checkNumber == "" {
		return nil, errors.Errorf("Missing valid order line in %q", msg.TextPlain)
	}
	if len(comments) != len(commentMatchers)+len(commentPrefixes)-1 {
		return nil, errors.Errorf("Missing comments.  Found %q in %q", comments, msg.TextPlain)
	}
	if amount == "" {
		return nil, errors.Errorf("Missing total line in %q", msg.TextPlain)
	}

	return importer.NewParsed(
		date,
		checkNumber,
		payee,
		comments,
		amount,
		defaultPayment), nil

}
