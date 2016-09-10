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
		citiHeader,
		citiCreditToAmount,
		citiNegateAmount,
		citiDollarAmount,
		ops.StripNewlines,
	}
}

func citiHeader(l *ops.Line) error {
	l.ReplaceHeader(citiHeaders)
	return nil
}

func citiCreditToAmount(l *ops.Line) error {
	l.MoveAndNegateIfPresent(citiCredit, citiAmount)
	return nil
}

func citiNegateAmount(l *ops.Line) error {
	l.Negate(citiAmount)
	return nil
}

func citiDollarAmount(l *ops.Line) error {
	l.EnsureDollars(citiAmount)
	return nil
}
