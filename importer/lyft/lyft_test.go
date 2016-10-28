package lyft

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
Hi Gina, thanks for riding with Jane D!

Receipt #999999999999999999
Ride completed on October 20 at 10:38 PM
Your Driver was Jane
Pickup: 450 California St, San Francisco, CA 94104
Dropoff: 700 4th St, San Francisco, CA 94107

Lyft fare (3.74mi, 20m 2s): $10.95
Prime Time  + 50%*: $5.48
Service fee: $1.75
Tip: $2.00

Total charged to Visa ***1234: $20.18

*50% Prime Time was included in your total.
Prime Time encourages more people to drive when Lyft gets really busy.
Learn More at http://email.lyftmail.com/someurl
--
Lose something, go to http://email.lyftmail.com/someotherurl
To learn more about our Zero Tolerance Policies, go to http://email.lyftmai=
l.com/somethirdurl
`)

var happyMsg = ledgertools.NewMessage(
	"Fri, 21 Oct 2016 14:23:49 +0000",
	"client@somehost.com",
	fromMatcher,
	"Your ride with Jane",
	happyEmail)

func TestHappyImport(t *testing.T) {
	parsed, err := ImportMessage(happyMsg)
	ok(t, err)

	year, month, day := parsed.Date.Date()
	equals(t, 2016, year)
	equals(t, time.October, month)
	equals(t, 21, day)

	equals(t, "999999999999999999", parsed.CheckNumber)
	equals(t,
		[]string{
			"Ride completed on October 20 at 10:38 PM",
			"Your Driver was Jane",
			"Pickup: 450 California St, San Francisco, CA 94104",
			"Dropoff: 700 4th St, San Francisco, CA 94107",
		},
		parsed.Comments)
	equals(t, "$20.18", parsed.Amount)
	equals(t, "Visa ***1234", parsed.PaymentInstrument)
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
