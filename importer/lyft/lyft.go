package lyft

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	ledgertools "github.com/ginabythebay/ledger-tools"
	"github.com/ginabythebay/ledger-tools/importer"
	"github.com/pkg/errors"
)

const (
	// From is the email address we expect to get ride summaries from
	From = "no-reply@lyftmail.com"

	// SubjectPrefix is the common prefix we see in ride summary subjects
	SubjectPrefix = "Your ride with"
)

var pacificTz *time.Location

var (
	receiptRe = regexp.MustCompile("^Receipt #(.+)$")
	chargeRe  = regexp.MustCompile("^Total charged to ([^:]+): (.+)$")
)

// We expect to find lines starting with these and will import those
// lines as comments.
var commentPrefixes = []string{
	"Ride completed on",
	"Your Driver was",
	"Pickup:",
	"Dropoff:",
}

const payee = "Lyft"

func init() {
	var err error
	if pacificTz, err = time.LoadLocation("America/Los_Angeles"); err != nil {
		panic(fmt.Sprintf("Loading America/Los_Angeles: %+v", err))
	}
}

// ImportMessage imports an email message.  Returns nil if msg does
// not appear to be a lyft ride summary.  Returns an error if it does
// appear// to be a lyft ride summary, but we have trouble parsing it.
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
	if msg.From != From {
		return nil, nil
	}
	if !strings.HasPrefix(msg.Subject, SubjectPrefix) {
		return nil, nil
	}

	date, err := time.Parse(time.RFC1123Z, msg.Date)
	if err != nil {
		return nil, errors.Wrapf(err, "Parsing date in %v", msg)
	}
	date = date.In(pacificTz)

	// build all these up by looking at the message text
	var checkNumber string
	var comments []string
	var amount string
	var instrument string

	for _, line := range strings.Split(msg.TextPlain, "\n") {
		for _, c := range commentPrefixes {
			if strings.HasPrefix(line, c) {
				comments = append(comments, line)
				continue
			}
		}

		if match := receiptRe.FindStringSubmatch(line); match != nil {
			checkNumber = match[1]
			continue
		}

		if match := chargeRe.FindStringSubmatch(line); match != nil {
			instrument = match[1]
			amount = match[2]
			continue
		}
	}

	// Verify that we found everything we were expecting
	if checkNumber == "" {
		return nil, errors.Errorf("Missing valid receipt line in %q", msg.TextPlain)
	}
	if len(comments) != len(commentPrefixes) {
		return nil, errors.Errorf("Missing comments.  Found %q in %q.  Expected to find lines starting with %q", comments, msg.TextPlain, commentPrefixes)
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
