package amazon

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

var stdEmail = strings.TrimSpace(`
Amazon.com Shipping Confirmation
http://www.amazon.com/someconfirmationlink

--------------------------------------------------------------------
Hello Gina White,

"Some Item..." and one other item have shipped.

Details
Order #123-1234567-1234567

Arriving:
    Monday, September 26

Track your package at:
	https://www.amazon.com/sometrackinglink

Shipped to:
    Gina White
    Some address...


====================================================================

    Total Before Tax: $26.38
    Tax Collected: $1.64
    Shipment Total: $28.02

====================================================================

View or manage your order in Your Orders:
https://www.amazon.com/someorderlink

We hope to see you again soon.<br/>
Amazon.com

--------------------------------------------------------------------
Unless otherwise noted, items sold by Amazon.com LLC are subject to sales tax in select states in accordance with the applicable laws of that state. If your order contains one or more items from a seller other than Amazon.com LLC, it may be subject to state and local sales tax, depending upon the sellers business policies and the location of their operations. For more tax and seller information, visit: http://www.amazon.com/sellerinfo

Items in this shipment may be subject to California's Electronic Waste Recycling Act. For any items not sold by Amazon.com LLC or Amazon Digital Services, Inc. that are subject to that Act, the seller of that item is responsible for submitting the California Electronic Waste Recycling fees on your behalf.

Your invoice can be accessed here:
https://www.amazon.com/invoicelink

This email was sent from a notification-only address that cannot accept incoming email. Please do not reply to this message.
`)

var stdMsg = ledgertools.NewMessage(
	"Sat, 24 Sep 2016 19:23:57 +0000",
	"client@somehost.com",
	fromMatcher,
	"Your Amazon.com order has shipped (#123-1234567-1234567)",
	stdEmail)

func TestStdImport(t *testing.T) {
	parsed, err := importMessage(stdMsg)
	ok(t, err)

	year, month, day := parsed.Date.Date()
	equals(t, 2016, year)
	equals(t, time.September, month)
	equals(t, 24, day)

	equals(t, "123-1234567-1234567", parsed.CheckNumber)
	equals(t,
		[]string{
			`"Some Item..." and one other item have shipped.`,
			"Order #123-1234567-1234567",
			"https://www.amazon.com/sometrackinglink",
			"https://www.amazon.com/someorderlink",
			"https://www.amazon.com/invoicelink",
		},
		parsed.Comments)
	equals(t, "$28.02", parsed.Amount)
	equals(t, defaultPayment, parsed.PaymentInstrument)
}

func BenchmarkStdImport(b *testing.B) {
	for i := 0; i < b.N; i++ {
		importMessage(stdMsg)
	}
}

var smileEmail = strings.TrimSpace(`
AmazonSmile Shipping Confirmation
http://smile.amazon.com/ref=TE_SIMP_g

--------------------------------------------------------------------
Hello Gina White,

"Some item..." has shipped.

Details
Order #987-9876543-9876543

Arriving:
    Friday, October 28 - Wednesday, November 2

Track your package at:
	https://smile.amazon.com/trackinglink

Shipped to:
    Gina White
    Some address...

		
====================================================================

    Total Before Tax: $22.99 
    Shipment Total: $22.99

====================================================================

View or manage your order in Your Orders:
https://smile.amazon.com/orderlink

We hope to see you again soon.<br/>
AmazonSmile

--------------------------------------------------------------------
Unless otherwise noted, items sold by Amazon.com LLC are subject to sales tax in select states in accordance with the applicable laws of that state. If your order contains one or more items from a seller other than Amazon.com LLC, it may be subject to state and local sales tax, depending upon the sellers business policies and the location of their operations. For more tax and seller information, visit: http://smile.amazon.com/sellerlink

Items in this shipment may be subject to California's Electronic Waste Recycling Act. For any items not sold by Amazon.com LLC or Amazon Digital Services, Inc. that are subject to that Act, the seller of that item is responsible for submitting the California Electronic Waste Recycling fees on your behalf.

Your invoice can be accessed here:
https://smile.amazon.com/invoicelink

This email was sent from a notification-only address that cannot accept incoming email. Please do not reply to this message.
`)

var smileMsg = ledgertools.NewMessage(
	"Mon, 24 Oct 2016 21:11:17 +0000",
	"client@somehost.com",
	fromMatcher,
	"Your AmazonSmile order has shipped (#987-9876543-9876543)",
	smileEmail)

func TestSmileImport(t *testing.T) {
	parsed, err := importMessage(smileMsg)
	ok(t, err)

	year, month, day := parsed.Date.Date()
	equals(t, 2016, year)
	equals(t, time.October, month)
	equals(t, 24, day)

	equals(t, "987-9876543-9876543", parsed.CheckNumber)
	equals(t,
		[]string{
			`"Some item..." has shipped.`,
			"Order #987-9876543-9876543",
			"https://smile.amazon.com/trackinglink",
			"https://smile.amazon.com/orderlink",
			"https://smile.amazon.com/invoicelink",
		},
		parsed.Comments)
	equals(t, "$22.99", parsed.Amount)
	equals(t, defaultPayment, parsed.PaymentInstrument)
}

func BenchmarkSmileImport(b *testing.B) {
	for i := 0; i < b.N; i++ {
		importMessage(stdMsg)
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
