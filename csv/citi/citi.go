package citi

import "github.com/ginabythebay/ledger-tools/csv/ops"

// Headers are to be inserted
var Headers = []string{"date", "amount", "payee", "", ""}

const (
	date int = iota
	amount
	payee
	code
	name
)

// Mutators returns the operations we want to perform on citibank csv files.
func Mutators() []ops.Mutator {
	return []ops.Mutator{
		ops.StripNewlines,
	}
}
