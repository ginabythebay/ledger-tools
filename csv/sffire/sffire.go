package sffire

import "github.com/ginabythebay/ledger-tools/csv/ops"

var headers = []string{"cleared", "date", "payee", "amount", "credit"}

const (
	date int = iota
	amount
	category
	description
	memo
	notes
)

// Mutators returns the operations we want to perform on citibank csv files.
func Mutators() []ops.Mutator {
	return []ops.Mutator{
		ops.StripSuffix(date, " 12:00:00"),
		ops.EnsureDollars(amount),
	}
}
