package parser

import (
	"reflect"
	"strings"
	"testing"
	"time"
)

var fileText = `
2016/03/01	payee 1; comment 1
	Expenses:Unknown					$83.93
	Assets:Checking					-$83.93

2016/03/15	payee 2; comment 2
	Expenses:Unknown					$229.42
	Assets:Checking					-$229.42
`

// TestParseScenario only tests the simple good scenario for now
func TestParseScenario(t *testing.T) {
	ledger, err := ParseLedger(strings.NewReader(fileText))
	if err != nil {
		t.Fatal(err)
	}
	if len(ledger) != 2 {
		t.Fatalf("Expected 2 transactions but read %d transactions", len(ledger))
	}

	firstDate, err := time.Parse("1/2/2006", "3/1/2016")
	if err != nil {
		t.Fatal(err)
	}

	firstExpected := Transaction{
		Line:    2,
		Date:    firstDate,
		Payee:   "payee 1",
		Comment: "; comment 1",
		Postings: []Posting{
			{"Expenses:Unknown", "$83.93"},
			{"Assets:Checking", "-$83.93"},
		},
	}
	if !reflect.DeepEqual(firstExpected, ledger[0]) {
		t.Errorf("First entry mismatch.  Expected %q and got %q", firstExpected, ledger[0])
	}

	secondDate, err := time.Parse("1/2/2006", "3/15/2016")
	if err != nil {
		t.Fatal(err)
	}

	secondExpected := Transaction{
		Line:    6,
		Date:    secondDate,
		Payee:   "payee 2",
		Comment: "; comment 2",
		Postings: []Posting{
			{"Expenses:Unknown", "$229.42"},
			{"Assets:Checking", "-$229.42"},
		},
	}
	if !reflect.DeepEqual(secondExpected, ledger[1]) {
		t.Errorf("Second entry mismatch.  Expected %q and got %q", secondExpected, ledger[1])
	}
}
