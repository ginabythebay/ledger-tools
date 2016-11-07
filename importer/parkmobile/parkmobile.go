package parkmobile

import (
	"fmt"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	ledgertools "github.com/ginabythebay/ledger-tools"
	"github.com/ginabythebay/ledger-tools/gmail"
	"github.com/ginabythebay/ledger-tools/importer"
	"github.com/ginabythebay/ledger-tools/importer/mailimp"
	"github.com/pkg/errors"
)

// GmailImporter knows how to fetch and parse lyft emails.
var GmailImporter = importer.NewGmailImporter(
	[]gmail.QuerySet{
		{gmail.QueryFrom(from), gmail.QuerySubject(subjectPrefix)},
	},
	[]importer.Parser{
		importMessage,
	},
)

const (
	from = "noreply@parkmobileglobal.com"

	subjectPrefix = "Parking Session Deactivated"
)

const (
	costMatcher    = "Total Cost:"
	sessionMatcher = "Session Id:"
)

var cmtMatch = map[string]bool{
	"Activated:":            true,
	"Deactivated:":          true,
	"Zone:":                 true,
	"Location":              true,
	"Space:":                true,
	"License Plate Number:": true,
	"Parking fee:":          true,
	"Transaction fee:":      true,
}

const payee = "ParkMobile"
const instrument = ""

// similar  time.RFC1123Z, but:
//  without the leading day,
//  yand allowing for single digit day of month
const dateFormat = "_2 Jan 2006 15:04:05 -0700"

type tr []string

// importMessage imports an email message.  Returns nil if msg does
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
func importMessage(msg ledgertools.Message) (*importer.Parsed, error) {
	if msg.From != from {
		return nil, nil
	}
	if !strings.HasPrefix(msg.Subject, subjectPrefix) {
		return nil, nil
	}
	date, err := time.Parse(dateFormat, msg.Date)
	if err != nil {
		return nil, errors.Wrapf(err, "Parsing date in %v", msg)
	}
	date = date.In(mailimp.PacificTz)

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(msg.TextHTML))
	if err != nil {
		return nil, errors.Wrap(err, "DocFromReader")
	}

	var rows []tr

	doc.Find("table").
		First().
		Find("tr").
		Each(func(i int, row *goquery.Selection) {
			var currentRow []string
			row.Find("td").
				Each(func(i int, cell *goquery.Selection) {
					currentRow = append(currentRow, cell.Text())
				})
			rows = append(rows, currentRow)
		})

	var amount string
	var checkNo string
	var comments []string
	for _, r := range rows {
		var lastCell string
		for _, cell := range r {
			if lastCell == costMatcher {
				amount = cell
			}
			if lastCell == sessionMatcher {
				checkNo = cell
			}
			if cmtMatch[lastCell] {
				s := fmt.Sprintf(
					"%s %s", strings.TrimSuffix(lastCell, ":"), cell)
				comments = append(comments, s)
			}

			lastCell = cell
		}
	}

	// Verify that we found everything we were expecting
	if checkNo == "" {
		return nil, errors.Errorf("missing valid session line in %q", msg.TextHTML)
	}
	if len(comments) != len(cmtMatch) {
		return nil, errors.Errorf("missing comments.  Found %q in %q.  Expected to find lines starting with %v", comments, msg.TextHTML, cmtMatch)
	}
	if amount == "" {
		return nil, errors.Errorf("missing total cost %q", msg.TextHTML)
	}

	return importer.NewParsed(
		date,
		checkNo,
		payee,
		comments,
		amount,
		instrument), nil
}
