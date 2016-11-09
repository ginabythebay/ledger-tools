package ledgertools

import (
	"fmt"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"
)

type tc struct {
	name    string
	imports []Flattened
	amounts [][]string
}

func TestImportScenario(t *testing.T) {
	cases := []tc{
		// This test case has a transaction with 2 postings followed
		// by a transaction with 3 postings.
		tc{
			"first",
			[]Flattened{
				flat(t, "a.txt", 1, "$1.00"),
				flat(t, "a.txt", 1, "$-1.00"),

				flat(t, "a.txt", 5, "$2.50"),
				flat(t, "a.txt", 5, "$3.55"),
				flat(t, "a.txt", 5, "$-6.05"),
			},
			[][]string{
				[]string{"1.00", "-1.00"},
				[]string{"2.50", "3.55", "-6.05"},
			},
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			var allTrans []*Transaction
			imports := c.imports
			for len(imports) != 0 {
				var tr *Transaction
				var err error
				tr, imports, err = NextTransaction(imports)
				ok(t, err)
				allTrans = append(allTrans, tr)
			}
			check(t, c.amounts, allTrans)
		})
	}
}

func flat(t *testing.T, file string, line int, amountText string) Flattened {
	currency, amount, err := parseAmount(amountText)
	ok(t, err)
	return Flattened{
		SrcFile:  file,
		BegLine:  line,
		Currency: currency,
		Amount:   amount,
	}
}

func check(t *testing.T, expectedAmounts [][]string, found []*Transaction) {
	equals(t, len(expectedAmounts), len(found))
	for i := 0; i < len(expectedAmounts); i++ {
		exp := expectedAmounts[i]
		tr := found[i]
		equals(t, len(exp), len(tr.Postings))
		for j := 0; j < len(exp); j++ {
			equals(t, exp[j], tr.Postings[j].Amount.Text('f', 2))
		}
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
		fmt.Printf("\033[31m%s:%d: unexpected error: %+v\033[39m\n\n", filepath.Base(file), line, err.Error())
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
