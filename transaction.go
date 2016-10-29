package ledgertools

import (
	"fmt"
	"strings"
	"time"
	"unicode/utf8"
)

const amountAlignmentCol = 65
const indent = "    "

// Transaction is a simple version of a transaction that is capable of
// represent movement of money between 2 accounts.
type Transaction struct {
	Date        time.Time
	CheckNumber string // may not be set
	Payee       string
	Comments    []string // These should not contain the leading ; character

	Amount         string // to apply to CostAccount
	CostAccount    string
	PaymentAccount string
}

// NewTransaction creates a new Transaction
func NewTransaction(date time.Time, checkNumber, payee string, comments []string, amount, costAccount, paymentAccount string) *Transaction {
	return &Transaction{
		date,
		checkNumber,
		payee,
		comments,
		amount,
		costAccount,
		paymentAccount,
	}
}

func (t Transaction) String() string {
	var lines []string
	dateText := t.Date.Format("2006/01/02")
	var tokens []string
	if t.CheckNumber == "" {
		tokens = []string{dateText, t.Payee}
	} else {
		tokens = []string{dateText, "(#" + t.CheckNumber + ")", t.Payee}
	}
	header := strings.Join(tokens, " ")
	lines = append(lines, header)
	for _, c := range t.Comments {
		lines = append(lines, fmt.Sprintf("%s; %s", indent, c))
	}
	lines = append(lines, align(t.CostAccount, t.Amount))
	lines = append(lines, indent+t.PaymentAccount)
	return strings.Join(lines, "\n")
}

func align(account string, amount string) string {
	prefix := indent + account
	suffix := "  " + amount
	delta := amountAlignmentCol - (utf8.RuneCountInString(prefix) + utf8.RuneCountInString(suffix))
	var middle string
	if delta > 0 {
		middle = strings.Repeat(" ", delta)
	}
	return prefix + middle + suffix
}
