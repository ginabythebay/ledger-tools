package citi

import "github.com/ginabythebay/ledger-tools/csv/ops"

var headers = []string{"cleared", "date", "payee", "amount", "credit"}

const (
	cleared int = iota
	date
	payee
	amount
	credit
)

// Mutators returns the operations we want to perform on citibank csv files.
func Mutators() []ops.Mutator {
	return []ops.Mutator{
		ops.ReplaceHeader(headers),
		ops.StripNewlines,
		ops.MoveAndNegateIfPresent(credit, amount),
		ops.EnsureDollars(amount),
		ops.StripCommas(amount),
	}
}
