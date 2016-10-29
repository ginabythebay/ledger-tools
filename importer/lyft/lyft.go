package lyft

import (
	"strings"
	"time"

	ledgertools "github.com/ginabythebay/ledger-tools"
	"github.com/ginabythebay/ledger-tools/importer"
	"github.com/ginabythebay/ledger-tools/importer/mailimp"
	"github.com/pkg/errors"
)

const (
	// From is the email address we expect to get ride summaries from
	From        = "no-reply@lyftmail.com"
	fromMatcher = "<" + From + ">"

	// SubjectPrefix is the common prefix we see in ride summary subjects
	SubjectPrefix = "Your ride with"
)

var receiptMatcher = mailimp.PrefixMatcher([]string{"Receipt #"})
var chargeMatcher = mailimp.PrefixMatcher([]string{"Total charged to "})

var commentMatchers = []mailimp.PrefixMatcher{
	{"Ride completed on ", "Line completed on "},
	{"Your Driver was "},
	{"Pickup: "},
	{"Dropoff: "},
}

const payee = "Lyft"

// ImportMessage imports an email message.  Returns nil if msg does
// not appear to be a lyft ride summary.  Returns an error if it does
// appear to be a lyft ride summary, but we have trouble parsing it.
// An example valid email would be:
//
//   Hi Gina, thanks for riding with Jane D!
//
//   Receipt #999999999999999999
//   Ride completed on October 20 at 10:38 PM
//   Your Driver was Jane
//   Pickup: 450 California St, San Francisco, CA 94104
//   Dropoff: 700 4th St, San Francisco, CA 94107
//
//   Lyft fare (3.74mi, 20m 2s): $10.95
//   Prime Time  + 50%*: $5.48
//   Service fee: $1.75
//   Tip: $2.00
//
//   Total charged to Visa ***1234: $20.18
//
//   *50% Prime Time was included in your total.
//   Prime Time encourages more people to drive when Lyft gets really busy.
//   Learn More at http://email.lyftmail.com/someurl
//   --
//   Lose something, go to http://email.lyftmail.com/someotherurl
//   To learn more about our Zero Tolerance Policies, go to http://email.lyftmai=
//   l.com/somethirdurl
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

		if match := receiptMatcher.Match(line); match != nil {
			checkNumber = match()
			continue
		}

		if match := chargeMatcher.Match(line); match != nil {
			// rest should like like 'Visa ***1234: $20.18'
			tokens := strings.SplitN(match(), ":", 2)
			if len(tokens) == 2 {
				instrument = tokens[0]
				amount = strings.TrimSpace(tokens[1])
				continue
			}
		}
	}

	// Verify that we found everything we were expecting
	if checkNumber == "" {
		return nil, errors.Errorf("Missing valid receipt line in %q", msg.TextPlain)
	}
	if len(comments) != len(commentMatchers) {
		return nil, errors.Errorf("Missing comments.  Found %q in %q.  Expected to find lines starting with %q", comments, msg.TextPlain, commentMatchers)
	}
	if amount == "" || instrument == "" {
		return nil, errors.Errorf("charge line in %q", msg.TextPlain)
	}

	return importer.NewParsed(
		date,
		checkNumber,
		payee,
		comments,
		amount,
		instrument), nil
}
