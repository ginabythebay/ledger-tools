package techcu

import "github.com/ginabythebay/ledger-tools/csv/ops"

var headers = []string{"cleared", "date", "payee", "amount", "credit"}

const (
	account int = iota
	date
	amount
	balance
	category
	description
	memo
	notes
)

// Mutators returns the operations we want to perform on citibank csv files.
func Mutators() []ops.Mutator {
	return []ops.Mutator{
		ops.DeparenNegatives(amount),
		ops.CheckWithdrawal(description),
	}
}
