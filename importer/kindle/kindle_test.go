package kindle

import (
	"fmt"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"
	"time"

	ledgertools "github.com/ginabythebay/ledger-tools"
)

var happyEmail = strings.TrimSpace(`
Hello Gina White,

Thank you for shopping with us. All Kindle content, including books and Kindle active content, that you've purchased from the Kindle Store is stored in your Kindle library https://www.amazon.com/liblink

............................................................................

Order Information:

E-mail Address:
        somebody@gmail.com

Order Grand Total:
        $5.98
............................................................................

Order Summary:

Details:

Order #: D12-1234567-1234567

Items Subtotal:                               	$5.98
Tax Collected:                                  $0.00
                                              ........................
Grand Total:                                    $5.98
............................................................................

First Title
Kindle Edition
Sold by Amazon  Digital Services  LLC

Second Title
Kindle Edition
Sold by Amazon  Digital Services  LLC

............................................................................

You can view your receipt or invoice by visiting the Order details page:

http://www.amazon.com/orderdetailslink

The charge for this order will appear on your credit card statement from
the merchant 'AMZN Payment Services'.

You can review your orders in Your Account.

If you've explored the links on that page but still have a question, please
visit our online Help Department:

http://www.amazon.com/somehelplink
............................................................................

Please note: This e-mail was sent from a notification-only address that
cannot accept incoming e-mail. Please do not reply to this message.

Thanks again for shopping with us.
`)

var happyMsg = ledgertools.NewMessage(
	"Thu, 20 Oct 2016 22:09:58 +0000",
	"client@somehost.com",
	fromMatcher,
	"Amazon.com order of First Title and 1 more item",
	happyEmail,
	"")

func TestHappyImport(t *testing.T) {
	parsed, err := importMessage(happyMsg)
	ok(t, err)

	year, month, day := parsed.Date.Date()
	equals(t, 2016, year)
	equals(t, time.October, month)
	equals(t, 20, day)

	equals(t, "D12-1234567-1234567", parsed.CheckNumber)
	equals(t,
		[]string{
			"Order #: D12-1234567-1234567",
			"First Title",
			"Second Title",
			"http://www.amazon.com/orderdetailslink",
		},
		parsed.Comments)
	equals(t, "$5.98", parsed.Amount)
	equals(t, defaultPayment, parsed.PaymentInstrument)
}

func BenchmarkHappyImport(b *testing.B) {
	for i := 0; i < b.N; i++ {
		importMessage(happyMsg)
	}
}

// assert fails the test if the condition is false.
func assert(tb testing.TB, condition bool, msg string, v ...interface{}) {
	if !condition {
		_, file, line, _ := runtime.Caller(1)
		fmt.Printf("\033[31m%s:%d: "+msg+"\033[39m\n\n", append([]interface{}{filepath.Base(file), line}, v...)...)
		tb.FailNow()
	}
}

// ok fails the test if an err is not nil.
func ok(tb testing.TB, err error) {
	if err != nil {
		_, file, line, _ := runtime.Caller(1)
		fmt.Printf("\033[31m%s:%d: unexpected error: %s\033[39m\n\n", filepath.Base(file), line, err.Error())
		tb.FailNow()
	}
}

// equals fails the test if exp is not equal to act.
func equals(tb testing.TB, exp, act interface{}) {
	if !reflect.DeepEqual(exp, act) {
		_, file, line, _ := runtime.Caller(1)
		fmt.Printf("\033[31m%s:%d:\n\n\texp: %#v\n\n\tgot: %#v\033[39m\n\n", filepath.Base(file), line, exp, act)
		tb.FailNow()
	}
}
