package importer

import (
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/ginabythebay/ledger-tools/rules"
	"github.com/pkg/errors"
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

func (t Transaction) String() string {
	var lines []string
	dateText := t.Date.Format("2006-01-02")
	var header string
	if t.CheckNumber == "" {
		header = fmt.Sprintf("%s %s", dateText, t.Payee)
	} else {
		header = fmt.Sprintf("%s (%s) %s", dateText, t.CheckNumber, t.Payee)
	}
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

// Parsed represents parsed data that we can convert to a Transaction with the help of a RuleSet.
type Parsed struct {
	Date        time.Time
	CheckNumber string // may not be set
	Payee       string
	Comments    []string // These should not contain the leading ; character

	Amount            string
	PaymentInstrument string
}

// NewParsed Creates a new Parsed entry
func NewParsed(date time.Time, checkNumber, payee string, comments []string, amount, paymentInstrument string) *Parsed {
	return &Parsed{date, checkNumber, payee, comments, amount, paymentInstrument}
}

func (p Parsed) transaction(rs *rules.RuleSet) (*Transaction, error) {
	var costAccount, paymentAccount string
	mappings := rs.Apply(
		rules.Input("Instrument", p.PaymentInstrument),
		rules.Input("Payee", p.Payee))

	if costAccount = mappings.Get("CostAccount"); costAccount == "" {
		return nil, errors.Errorf("Unable to determine cost account for payee %q", p.Payee)
	}
	if paymentAccount = mappings.Get("PaymentAccount"); paymentAccount == "" {
		return nil, errors.Errorf("Unable to determine payment account for instrument %q", p.PaymentInstrument)
	}

	return &Transaction{
		p.Date,
		p.CheckNumber,
		p.Payee,
		p.Comments,
		p.Amount,
		costAccount,
		paymentAccount,
	}, nil
}
