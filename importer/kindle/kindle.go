package kindle

import (
	"strings"
	"time"

	ledgertools "github.com/ginabythebay/ledger-tools"
	"github.com/ginabythebay/ledger-tools/importer"
	"github.com/ginabythebay/ledger-tools/importer/mailimp"
	"github.com/pkg/errors"
)

const (
	// From is the email address we expect to get amazon digital orders from
	From        = "digital-no-reply@amazon.com"
	fromMatcher = "<" + From + ">"

	// SubjectPrefix is the common prefix we see in shipment subjects
	SubjectPrefix = "Amazon.com order of "

	// DefaultPayment represents the payment instrument we always
	// return as Amazon doesn't send us any payment instrument
	// information.
	DefaultPayment = "KindleDefaultPayment"
)

const payee = "Amazon Kindle"

var orderMatcher = mailimp.PrefixMatcher([]string{"Order #:"})

const (
	grandTotalLine = "Order Grand Total:"
	kindleEdition  = "Kindle Edition"
)

var commentMatchers = []mailimp.LineMatcher{
	mailimp.PrefixMatcher([]string{"Order #:"}),
}

// If we see any of these, then we want to capture the following non-blank line
// as a comment.
var commentPrefixes = []string{
	"You can view your receipt or invoice by visiting the Order details page:",
}

// ImportMessage imports an email message.  Returns nil if msg does
// not appear to be a amazon digital order.  Returns an error if it
// does appear to be a amazon digital order, but we have trouble
// parsing it.  An example valid email would be:
//
//   Hello Gina White,
//
//   Thank you for shopping with us. All Kindle content, including books and Kindle active content, that you've purchased from the Kindle Store is stored in your Kindle library https://www.amazon.com/liblink
//
//   ............................................................................
//
//   Order Information:
//
//   E-mail Address:
//           somebody@gmail.com
//
//   Order Grand Total:
//           $5.98
//   ............................................................................
//
//   Order Summary:
//
//   Details:
//
//   Order #: D12-1234567-1234567
//
//   Items Subtotal:                               	$5.98
//   Tax Collected:                                  $0.00
//                                                 ........................
//   Grand Total:                                    $5.98
//   ............................................................................
//
//   First Title
//   Kindle Edition
//   Sold by Amazon  Digital Services  LLC
//
//   Second Title
//   Kindle Edition
//   Sold by Amazon  Digital Services  LLC
//
//   ............................................................................
//
//   You can view your receipt or invoice by visiting the Order details page:
//
//   http://www.amazon.com/orderdetailslink
//
//   The charge for this order will appear on your credit card statement from
//   the merchant 'AMZN Payment Services'.
//
//   You can review your orders in Your Account.
//
//   If you've explored the links on that page but still have a question, please
//   visit our online Help Department:
//
//   http://www.amazon.com/somehelplink
//   ............................................................................
//
//   Please note: This e-mail was sent from a notification-only address that
//   cannot accept incoming e-mail. Please do not reply to this message.
//
//   Thanks again for shopping with us.
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
	var comments = make([]string, 0, len(commentMatchers)+len(commentPrefixes)+5)
	var amount string

	splitter := mailimp.NewLineSplitter(msg.TextPlain)
	var line, lastLine string
	var ok bool
	for {
		if line != "" {
			lastLine = line
		}
		line, ok = splitter.Next()
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
				if lastLine == pre {
					candidate := strings.TrimSpace(line)
					if candidate != "" {
						comments = append(comments, candidate)
					}
				}
			}
		}

		if match := orderMatcher.Match(line); match != nil {
			checkNumber = strings.TrimSpace(match())
			continue
		}

		if lastLine == grandTotalLine {
			amount = strings.TrimSpace(line)
			continue
		}
		if line == kindleEdition {
			comments = append(comments, strings.TrimSpace(lastLine))
			continue
		}

	}

	// Verify that we found everything we were expecting
	if checkNumber == "" {
		return nil, errors.Errorf("Missing valid order line in %q", msg.TextPlain)
	}
	if len(comments) < len(commentMatchers)+len(commentPrefixes)+1 {
		return nil, errors.Errorf("Missing comments.  Found %q in %q", comments, msg.TextPlain)
	}
	if amount == "" {
		return nil, errors.Errorf("Total line in %q", msg.TextPlain)
	}

	return importer.NewParsed(
		date,
		checkNumber,
		payee,
		comments,
		amount,
		DefaultPayment), nil
}
