package csv

import (
	"bytes"
	"fmt"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"

	"github.com/ginabythebay/ledger-tools/csv/citi"
	"github.com/ginabythebay/ledger-tools/csv/ops"
)

//
// BEGIN HELPERS FROM https://github.com/benbjohnson/testing
//

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

//
// END HELPERS FROM https://github.com/benbjohnson/testing
//

var citiInput = strings.TrimSpace(`
"Status","Date","Description","Debit","Credit"
"Cleared","07/27/2016","GITHUB.COM  2IK4B        415-448-6673 CA
","1.01",""
"Cleared","08/09/2016","NORDSTROM #0427          SAN FRANCISCOCA
","","2.02"
"Cleared","08/25/2016","INTEREST CHARGED TO STANDARD PURCH      
","3.03",""
`)

const citiExpected = `cleared,date,payee,amount,credit
Cleared,07/27/2016,GITHUB.COM  2IK4B        415-448-6673 CA,$-1.01,
Cleared,08/09/2016,NORDSTROM #0427          SAN FRANCISCOCA,$2.02,
Cleared,08/25/2016,INTEREST CHARGED TO STANDARD PURCH      ,$-3.03,
`

type testCase struct {
	name     string
	mutators []ops.Mutator
	input    string
	expected string
}

func (c testCase) testFunc(t *testing.T) {
	var buf bytes.Buffer
	actLineCnt, err := Process(c.mutators, strings.NewReader(c.input), &buf)
	if err != nil {
		t.Error(err)
	}
	actual := buf.String()
	equals(t, c.expected, actual)
	equals(t, strings.Count(c.expected, "\n"), actLineCnt)
}

var cases = []testCase{
	testCase{"citi", citi.Mutators(), citiInput, citiExpected},
}

func TestProcess(t *testing.T) {
	for _, c := range cases {
		t.Run(c.name, c.testFunc)
	}
}
