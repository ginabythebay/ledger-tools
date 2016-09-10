package citi

import "github.com/ginabythebay/ledger-tools/csv/ops"

var citiHeaders = []string{"cleared", "date", "payee", "amount", "credit"}

const (
	citiCleared int = iota
	citiDate
	citiPayee
	citiAmount
	citiCredit
)

// Mutators returns the operations we want to perform on citibank csv files.
func Mutators() []ops.Mutator {
	return []ops.Mutator{
		ops.ReplaceHeader(citiHeaders),
		ops.StripNewlines,
		ops.MoveAndNegateIfPresent(citiCredit, citiAmount),
		ops.EnsureDollars(citiAmount),
		ops.StripCommas(citiAmount),
	}
}
