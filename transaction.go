package ledgertools

import (
	"fmt"
	"math/big"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/pkg/errors"
)

const amountAlignmentCol = 65
const indent = "    "

// Posting represents a change to an account, along with associated metadata.
type Posting struct {
	// These fields are common to all postings in the same transaction.
	Date    time.Time
	CheckNo string // may not be set
	Payee   string

	Account  string
	Currency string
	Amount   big.Float
	State    rune
}

func (p *Posting) String() string {
	prefix := indent + p.Account
	suffix := "  " + p.Currency + p.Amount.Text('f', 2)
	delta := amountAlignmentCol - (utf8.RuneCountInString(prefix) + utf8.RuneCountInString(suffix))
	var middle string
	if delta > 0 {
		middle = strings.Repeat(" ", delta)
	}
	return prefix + middle + suffix
}

// Transaction is group of related Postings, with an optional shared
// comment.
type Transaction struct {
	Comments []string // These should not contain the leading ; character
	Postings []Posting
}

// NewTransaction creates a new Transaction
func NewTransaction(date time.Time, checkNo, payee string, comments []string, amountText, costAccount, paymentAccount string) (*Transaction, error) {
	currency, amount, err := parseAmount(amountText)
	if err != nil {
		return nil, errors.Wrap(err, "parseAmount")
	}
	var negatedAmount big.Float
	negatedAmount.Neg(&amount)
	cost := Posting{
		Date:     date,
		CheckNo:  checkNo,
		Payee:    payee,
		Account:  costAccount,
		Currency: currency,
		Amount:   amount,
	}
	payment := Posting{
		Date:     date,
		CheckNo:  checkNo,
		Payee:    payee,
		Account:  paymentAccount,
		Currency: currency,
		Amount:   negatedAmount,
	}
	return &Transaction{comments, []Posting{cost, payment}}, nil
}

func parseAmount(s string) (currency string, amount big.Float, err error) {
	if !strings.HasPrefix(s, "$") {
		return "", amount, errors.New("unable to parse %q as an amount.  Currently we only support amounts with a leading $")
	}
	currency = "$"
	_, _, err = amount.Parse(s[1:], 10)
	return currency, amount, err
}

func (t Transaction) String() string {
	var lines []string
	first := &t.Postings[0]
	dateText := first.Date.Format("2006/01/02")
	var tokens []string
	if first.CheckNo == "" {
		tokens = []string{dateText, first.Payee}
	} else {
		tokens = []string{dateText, "(#" + first.CheckNo + ")", first.Payee}
	}
	header := strings.Join(tokens, " ")
	lines = append(lines, header)
	for _, c := range t.Comments {
		lines = append(lines, fmt.Sprintf("%s; %s", indent, c))
	}
	for _, p := range t.Postings {
		lines = append(lines, p.String())
	}
	return strings.Join(lines, "\n")
}

func SortTransactions(transactions []*Transaction) {
	s := sorter(transactions)
	sort.Sort(s)
}

type sorter []*Transaction

func (s sorter) Len() int {
	return len(s)
}

func (s sorter) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s sorter) Less(i, j int) bool {
	return s[i].Postings[0].Date.Before(s[j].Postings[0].Date)
}
