package importer

import (
	"fmt"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestTransactionString(t *testing.T) {
	when, err := time.Parse("2006-01-02", "2016-10-28")
	ok(t, err)
	trans := Transaction{
		when,
		"#3030",
		"Giant Corporation",
		[]string{"first comment", "second comment"},
		"$30.00",
		"Expenses:Go",
		"Liabilities:CreditCard",
	}

	equals(t,
		strings.TrimSpace(`
2016-10-28 (#3030) Giant Corporation
    ; first comment
    ; second comment
    Expenses:Go                                            $30.00
    Liabilities:CreditCard
`),
		trans.String(),
	)

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
