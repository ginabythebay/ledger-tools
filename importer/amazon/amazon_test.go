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
	"Sat, 4 Sep 2016 19:23:57 +0000",
	"client@somehost.com",
	fromMatcher,
	"Your Amazon.com order has shipped (#123-1234567-1234567)",
	stdEmail,
	"")

func TestStdImport(t *testing.T) {
	parsed, err := importMessage(stdMsg)
	ok(t, err)

	year, month, day := parsed.Date.Date()
	equals(t, 2016, year)
	equals(t, time.September, month)
	equals(t, 4, day)

	equals(t, "123-1234567-1234567", parsed.CheckNumber)
	equals(t,
		[]string{
			`"Some Item..." and one other item have shipped.`,
			"Order #123-1234567-1234567",
			"https://www.amazon.com/sometrackinglink",
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

var stdEmail2 = strings.TrimSpace(`
Hello Gina White,

"Some item..." has shipped.

Details
Order #987-9876543-9876543

Arriving:
    Friday, October 21 - Thursday, November 3

Why tracking information may not be available?:
http://www.amazon.com/why

Shipped to:
    Gina White
    3333 SOME ST...

		
====================================================================

    Total Before Tax: $14.99 
    Shipment Total: $14.99

====================================================================

View or manage your order in Your Orders:
https://www.amazon.com/manage

We hope to see you again soon.<br/>
Amazon.com

--------------------------------------------------------------------
Unless otherwise noted, items sold by Amazon.com LLC are subject to sales tax in select states in accordance with the applicable laws of that state. If your order contains one or more items from a seller other than Amazon.com LLC, it may be subject to state and local sales tax, depending upon the sellers business policies and the location of their operations. For more tax and seller information, visit: http://www.amazon.com/gp/help/customer/display.html?ie=UTF8&nodeId=200962600 

Items in this shipment may be subject to California's Electronic Waste Recycling Act. For any items not sold by Amazon.com LLC or Amazon Digital Services, Inc. that are subject to that Act, the seller of that item is responsible for submitting the California Electronic Waste Recycling fees on your behalf.

Your invoice can be accessed here:
https://www.amazon.com/invoice

This email was sent from a notification-only address that cannot accept incoming email. Please do not reply to this message.  
`)

var stdMsg2 = ledgertools.NewMessage(
	"Sat, 27 Sep 2016 19:23:57 +0000",
	"client@somehost.com",
	fromMatcher,
	"Your Amazon.com order has shipped (#987-9876543-9876543)",
	stdEmail2,
	"")

func TestStdImport2(t *testing.T) {
	parsed, err := importMessage(stdMsg2)
	ok(t, err)

	year, month, day := parsed.Date.Date()
	equals(t, 2016, year)
	equals(t, time.September, month)
	equals(t, 27, day)

	equals(t, "987-9876543-9876543", parsed.CheckNumber)
	equals(t,
		[]string{
			`"Some item..." has shipped.`,
			"Order #987-9876543-9876543",
			"http://www.amazon.com/why",
		},
		parsed.Comments)
	equals(t, "$14.99", parsed.Amount)
	equals(t, defaultPayment, parsed.PaymentInstrument)
}

func BenchmarkStdImport2(b *testing.B) {
	for i := 0; i < b.N; i++ {
		importMessage(stdMsg2)
	}
}

var stdEmail3 = strings.TrimSpace(`
Amazon Shipping Confirmation
https://www.amazon.com?ie=UTF8&ref_=scr_home

____________________________________________________________________

Hi Gina, your package will arrive:
Tuesday, December 27

Track your package:
https://www.amazon.com/trackinglink

On the way:
some things...
Order #anordernumber

Ship to:
Gina White
an address...

Shipment total:
$11.73

Return or replace items in Your Orders
https://www.amazon.com/returnorreplacelink

____________________________________________________________________


Unless otherwise noted, items sold by Amazon.com LLC are subject to sales tax in select states in accordance with the applicable laws of that state. If your order contains one or more items from a seller other than Amazon.com LLC, it may be subject to state and local sales tax, depending upon the sellers business policies and the location of their operations. Learn more about tax and seller information:
http://www.amazon.com/infolink

Your invoice can be accessed here:
https://www.amazon.com/invoicelink

This email was sent from a notification-only address that cannot accept incoming email. Please do not reply to this message.
`)

var stdMsg3 = ledgertools.NewMessage(
	"Fri, 23 Dec 2016 22:07:42 +0000",
	"client@somehost.com",
	fromMatcher,
	"Your Amazon.com order has shipped (#987-9876543-9876543)",
	stdEmail3,
	"")

func TestStdImport3(t *testing.T) {
	parsed, err := importMessage(stdMsg3)
	ok(t, err)

	year, month, day := parsed.Date.Date()
	equals(t, 2016, year)
	equals(t, time.December, month)
	equals(t, 23, day)

	equals(t, "anordernumber", parsed.CheckNumber)
	equals(t,
		[]string{
			"https://www.amazon.com/trackinglink",
			"some things...",
			"Order #anordernumber",
		},
		parsed.Comments)
	equals(t, "$11.73", parsed.Amount)
	equals(t, defaultPayment, parsed.PaymentInstrument)
}

func BenchmarkStdImport3(b *testing.B) {
	for i := 0; i < b.N; i++ {
		importMessage(stdMsg3)
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
	smileEmail,
	"")

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

var bookEmail = strings.TrimSpace(`
Amazon.com Shipping Confirmation
http://www.amazon.com/ref=TE_SIMP_g

--------------------------------------------------------------------
Hello Gina White,

"Book Title..." has shipped.

Details
Order #123-1234567-1234567

Arriving:
    Tuesday, October 18

Track your package at:
	https://www.amazon.com/trackinglink

Shipped to:
    Gina White
    Some Address...

		
====================================================================

    Total Before Tax: $21.37 
    Tax Collected: $1.87
    Shipment Total: $23.24

====================================================================

View or manage your order in Your Orders:
https://www.amazon.com/orderlink

Return or replace your items in Your Orders(https://www.amazon.com/historylink)


We hope to see you again soon.<br/>
Amazon.com

--------------------------------------------------------------------
Unless otherwise noted, items sold by Amazon.com LLC are subject to sales tax in select states in accordance with the applicable laws of that state. If your order contains one or more items from a seller other than Amazon.com LLC, it may be subject to state and local sales tax, depending upon the sellers business policies and the location of their operations. For more tax and seller information, visit: http://www.amazon.com/sellerlink

Your invoice can be accessed here:
https://www.amazon.com/invoicelink

This email was sent from a notification-only address that cannot accept incoming email. Please do not reply to this message.     `)

var bookMsg = ledgertools.NewMessage(
	"Sun, 16 Oct 2016 18:39:50 +0000",
	"client@somehost.com",
	fromMatcher,
	"Your Amazon.com order of \"Book Title...\" has shipped!",
	bookEmail,
	"")

func TestBookImport(t *testing.T) {
	parsed, err := importMessage(bookMsg)
	ok(t, err)

	year, month, day := parsed.Date.Date()
	equals(t, 2016, year)
	equals(t, time.October, month)
	equals(t, 16, day)

	equals(t, "123-1234567-1234567", parsed.CheckNumber)
	equals(t,
		[]string{
			`"Book Title..." has shipped.`,
			"Order #123-1234567-1234567",
			"https://www.amazon.com/trackinglink",
		},
		parsed.Comments)
	equals(t, "$23.24", parsed.Amount)
	equals(t, defaultPayment, parsed.PaymentInstrument)
}

func BenchmarkBookImport(b *testing.B) {
	for i := 0; i < b.N; i++ {
		importMessage(stdMsg)
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
