package dup

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"

	ledgertools "github.com/ginabythebay/ledger-tools"
	"github.com/ginabythebay/ledger-tools/register"
)

var fileText = strings.TrimSpace(`
commodity $

account Expenses:Grocery
account Liabilities:Another Credit Card
account Liabilities:Credit Card

tag SuppressDuplicates

2016/03/21 Local Grocery Store
    Expenses:Grocery                          $10.00
    Liabilities:Credit Card

2016/03/22 Another Local Grocery Store
    Expenses:Grocery                          $10.00
    Liabilities:Another Credit Card

2016/03/25 Another Local Grocery Store
    Expenses:Grocery                          $10.00
    ; SuppressDuplicates: 2016/03/22
    Liabilities:Credit Card

2016/04/21 Another Local Grocery Store
    Expenses:Grocery                          $10.00
    Liabilities:Another Credit Card
`)

func Test_JavacIntegration(t *testing.T) {
	allTrans, fileName := readRegister(t, fileText)

	finder := NewFinder(3)
	for _, t := range allTrans {
		finder.Add(t)
	}

	var b bytes.Buffer
	write := JavacWriter(finder.AllPairs, &b)
	ok(t, write())

	exp := fmt.Sprintf(strings.TrimSpace(`
Possible duplicate $10.00 Expenses:Grocery
	at 2016/03/21 Local Grocery Store (%s:10)
	at 2016/03/22 Another Local Grocery Store (%s:14)

 1 potential duplicates found`), fileName, fileName)
	equals(t, exp, strings.TrimSpace(b.String()))
}

func readRegister(t *testing.T, s string) ([]*ledgertools.Transaction, string) {
	f, err := ioutil.TempFile("", "testregister")
	tempName := f.Name()
	ok(t, err)
	defer func() {
		ok(t, os.Remove(tempName))
	}()

	b := []byte(s)
	for len(b) != 0 {
		var n int
		n, err = f.Write(b)
		ok(t, err)
		b = b[n:]
	}
	ok(t, f.Close())

	allTrans, err := register.Read(tempName)
	ok(t, err)
	return allTrans, tempName
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
