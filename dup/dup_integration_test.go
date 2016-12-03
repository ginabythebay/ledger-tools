package dup

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"

	ledgertools "github.com/ginabythebay/ledger-tools"
	"github.com/ginabythebay/ledger-tools/register"
)

//go:generate ./integration_src.sh

var csvText = strings.TrimSpace(`
"integration_src.ledger","9","","2016/03/21","","Local Grocery Store","10","Expenses:Grocery","$","10","",""
"integration_src.ledger","9","","2016/03/21","","Local Grocery Store","11","Liabilities:Credit Card","$","-10","",""
"integration_src.ledger","12","","2016/03/22","","Another Local Grocery Store","13","Expenses:Grocery","$","10","",""
"integration_src.ledger","12","","2016/03/22","","Another Local Grocery Store","14","Liabilities:Another Credit Card","$","-10","",""
"integration_src.ledger","15","","2016/03/25","","Another Local Grocery Store","16","Expenses:Grocery","$","10",""," SuppressDuplicates: 2016/03/22"
"integration_src.ledger","15","","2016/03/25","","Another Local Grocery Store","18","Liabilities:Credit Card","$","-10","",""
"integration_src.ledger","19","","2016/04/21","","Another Local Grocery Store","20","Expenses:Grocery","$","10","",""
"integration_src.ledger","19","","2016/04/21","","Another Local Grocery Store","21","Liabilities:Another Credit Card","$","-10","",""
`)

func Test_JavacIntegration(t *testing.T) {
	allTrans := readRegister(t, csvText)

	finder := NewFinder(3)
	for _, t := range allTrans {
		finder.Add(t)
	}

	var b bytes.Buffer
	write := JavacWriter(finder.AllPairs, &b)
	ok(t, write())

	exp := strings.TrimSpace(`
Possible duplicate $10.00 Expenses:Grocery
	at 2016/03/21 Local Grocery Store (integration_src.ledger:10)
	at 2016/03/22 Another Local Grocery Store (integration_src.ledger:13)

 1 potential duplicates found
`)
	equals(t, exp, strings.TrimSpace(b.String()))
}

func readRegister(t *testing.T, s string) []*ledgertools.Transaction {

	allTrans, err := register.ReadLedgerCsv(ioutil.NopCloser(strings.NewReader(s)))
	ok(t, err)
	return allTrans
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
