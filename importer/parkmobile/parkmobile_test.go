package parkmobile

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
<html>     <head>         <title></title>     </head>  <body style="font-family: Verdana; font-size: x-small">         <table>         <tr>             <td width="5%"></td>             <td colspan="2">                 <img alt="Parkmobile Paying Made simple" src="https://content.parkmobile.us/Phonixx/img/PM_logo_email_small.png"  height="70px" width="70px" border="0" />                  <br/></td>             <td></td>         </tr>         <tr><td colspan="4" height="15px"></td></tr>          <tr>              <td width="5%"></td>              <td colspan="2"><b>CONFIRMATION - PARKING SESSION DEACTIVATED<br />                              </b><br/></td>              <td></td>          </tr>         <tr>             <td></td>             <td colspan="2">This confirmation indicates that your session has been deactivated.<br />                 <br /></td>             <td></td>         </tr>         <tr>             <td></td>             <td colspan="2"><b>Transaction Details:</b><br />                 <br /></td>             <td></td>         </tr>   <tr>             <td></td>             <td width="20%">Session Id:</td>             <td width="70%">SESSION_ID</td>             <td width="5%"></td>         </tr>         <tr>             <td></td>             <td width="20%">Activated:</td>             <td width="70%">11/2/2016 3:22 PM</td>             <td width="5%"></td>         </tr>          <tr>             <td></td>             <td width="20%">Deactivated:</td>             <td width="70%">11/2/2016 3:59 PM</td>             <td width="5%"></td>         </tr>         <tr>             <td>&nbsp;</td>             <td>Zone:</td>             <td>7207</td>             <td>&nbsp;</td>         </tr>         <tr>             <td></td>             <td>Location</td>             <td>Stanford University</td>             <td></td>         </tr>         <tr>             <td></td>             <td>Space:</td>             <td>20</td>             <td></td>         </tr>         <tr>             <td></td>             <td>License Plate Number:</td>             <td>SOME_PLATE</td>             <td></td>         </tr>            <tr>             <td></td>             <td>Parking fee:</td>             <td>$1.25</td>             <td></td>         </tr>         <tr>             <td></td>             <td>Transaction fee:</td>             <td>$0.35</td>             <td></td>         </tr> <!--                  <tr>              <td></td>              <td>Discounts:</td>              <td>N/A&nbsp;&nbsp;&nbsp;N/A</td>              <td>&nbsp;</td>          </tr> -->   <!--                  <tr>             <td></td>             <td>Taxes:</td>             <td>$0.00</td>             <td></td>         </tr> -->                  <tr>             <td></td>             <td><b>Total Cost:</b></td>             <td><b>$1.60</b></td>             <td></td>         </tr>      <tr>              <td>&nbsp;</td>              <td colspan="2">                  <br />                       For questions about your transaction, please visit <a href="http://phonixx.parkmobile.us">phonixx.parkmobile.us</a> to search our Knowledge Base,                  or email us at <a href="mailto:helpdesk@parkmobileglobal.com">helpdesk@parkmobileglobal.com</a> with the transaction details listed above.                                 <br />                  To stop receiving parking confirmation email messages,                    <a href="https://dlweb.ParkMobile.us/unsub">click this link</a>.<br />                  <br />                  <br />                  To stop receiving ALL emails sent by Parkmobile, <a href="https://dlweb.ParkMobile.us/stopall">click this link</a>.<br />              </td>              <td></td>          </tr>                     <tr>              <td></td>                <td colspan="2"><i><br/>  This message was sent by Parkmobile. If you no longer wish to receive these alerts,                     log in to your Personal Pages at <a href="http://phonixx.parkmobile.us">phonixx.parkmobile.us</a>                    and choose Alerts & Messages to adjust your settings.</i><br />                  <br /></td>              <td></td>          </tr>                 <tr>             <td></td>             <td colspan="2" style="font-family: Verdana; font-size: x-small; color: darkgray; text-align: center">                 <br />                 <br />                 Parkmobile USA | 1100 Spring Street NW, Suite 200, Atlanta, GA 30309 <br/>                 Member Services: 877-727-5457 | helpdesk@parkmobileglobal.com | phonixx.parkmobile.us             </td>             <td></td>         </tr>      </table> </body> </html>
`)

var happyMsg = ledgertools.NewMessage(
	"2 Nov 2016 19:37:29 -0400",
	"client@somehost.com",
	from,
	"Parking Session Deactivated",
	"",
	happyEmail)

func TestHappyImport(t *testing.T) {
	parsed, err := importMessage(happyMsg)
	ok(t, err)

	year, month, day := parsed.Date.Date()
	equals(t, 2016, year)
	equals(t, time.November, month)
	equals(t, 2, day)

	equals(t, "SESSION_ID", parsed.CheckNumber)
	equals(t,
		[]string{
			"Activated 11/2/2016 3:22 PM",
			"Deactivated 11/2/2016 3:59 PM",
			"Zone 7207",
			"Location Stanford University",
			"Space 20",
			"License Plate Number SOME_PLATE",
			"Parking fee $1.25",
			"Transaction fee $0.35",
		},
		parsed.Comments)
	equals(t, "$1.60", parsed.Amount)
	equals(t, "", parsed.PaymentInstrument)
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
		fmt.Printf("\033[31m%s:%d: unexpected error: %+v\033[39m\n\n", filepath.Base(file), line, err)
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
