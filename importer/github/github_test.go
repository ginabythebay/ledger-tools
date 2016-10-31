package github

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
We received payment for your GitHub.com subscription. Thanks for your business!

Questions? Please contact support@github.com.

------------------------------------
GITHUB RECEIPT - PERSONAL SUBSCRIPTION - ginabythebay

Personal plan

Amount: USD $7.00*

Charged to: Card (1*** **** **** 1234)
Transaction ID: ABC1ABCD
Date: 27 Oct 2016 04:11AM PDT
For service through: 2016-11-26


GitHub, Inc.
88 Colin P. Kelly Jr. Street
San Francisco, CA 94107
------------------------------------

* EU customers: Prices inclusive of VAT, where applicable
`)

var happyMsg = ledgertools.NewMessage(
	"Thu, 27 Oct 2016 04:11:14 -0700",
	"client@somehost.com",
	fromMatcher,
	"[GitHub] Payment Receipt for ginabythebay",
	happyEmail)

func TestHappyImport(t *testing.T) {
	parsed, err := importMessage(happyMsg)
	ok(t, err)

	year, month, day := parsed.Date.Date()
	equals(t, 2016, year)
	equals(t, time.October, month)
	equals(t, 27, day)

	equals(t, "ABC1ABCD", parsed.CheckNumber)
	equals(t,
		[]string{
			"GITHUB RECEIPT - PERSONAL SUBSCRIPTION - ginabythebay",
			"For service through: 2016-11-26",
		},
		parsed.Comments)
	equals(t, "$7.00", parsed.Amount)
	equals(t, "Card (1*** **** **** 1234)", parsed.PaymentInstrument)
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
