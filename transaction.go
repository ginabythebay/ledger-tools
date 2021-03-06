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
	BegLine  int
	Account  string
	Currency string
	Amount   big.Float
	State    rune
	Notes    []string
	Xact     *Transaction
}

func (p *Posting) String() string {
	var prefix string
	if p.State == 0 {
		prefix = indent + p.Account
	} else {
		prefix = indent + " " + string(p.State) + " " + p.Account
	}

	suffix := "  " + p.AmountText()
	delta := amountAlignmentCol - (utf8.RuneCountInString(prefix) + utf8.RuneCountInString(suffix))
	var middle string
	if delta > 0 {
		middle = strings.Repeat(" ", delta)
	}
	return prefix + middle + suffix
}

// AmountText returns the currency and amount in a text format.
func (p *Posting) AmountText() string {
	return p.Currency + p.Amount.Text('f', 2)
}

// Transaction is group of related Postings, with an optional shared
// comment.
type Transaction struct {
	SrcFile  string // may not be set
	BegLine  int    // may not be set
	Date     time.Time
	Code     string // may not be set.  The thing in parentheses.  e.g. check #
	Payee    string
	Notes    []string // may not be set
	Postings []*Posting
}

// LinkPostings points all postings back to their parent transaction
// and returns a pointer to that transaction.
func (t *Transaction) LinkPostings() *Transaction {
	for _, p := range t.Postings {
		p.Xact = t
	}
	return t
}

// SyntheticTransaction creates a new Transaction
func SyntheticTransaction(date time.Time, code, payee string, notes []string, amountText, costAccount, paymentAccount string) (*Transaction, error) {
	currency, amount, err := parseAmount(amountText)
	if err != nil {
		return nil, errors.Wrap(err, "parseAmount")
	}
	var negatedAmount big.Float
	negatedAmount.Neg(&amount)
	cost := Posting{
		Account:  costAccount,
		Currency: currency,
		Amount:   amount,
	}
	payment := Posting{
		Account:  paymentAccount,
		Currency: currency,
		Amount:   negatedAmount,
	}

	t := &Transaction{
		Date:     date,
		Code:     code,
		Payee:    payee,
		Notes:    notes,
		Postings: []*Posting{&cost, &payment}}
	return t.LinkPostings(), nil
}

// NextTransaction creates the next transaction from a series of
// imports, returns that, as well as the remaining imports.
func NextTransaction(imports []Flattened) (*Transaction, []Flattened, error) {
	if len(imports) == 0 {
		return nil, imports, nil
	}
	var first Flattened
	for i, f := range imports {
		if i == 0 {
			first = f
			continue
		}
		if !first.sameTransaction(&f) {
			return consolidate(imports, i)
		}
	}
	return consolidate(imports, len(imports))
}

func consolidate(imports []Flattened, end int) (*Transaction, []Flattened, error) {
	use := imports[:end]
	if len(use) < 2 {
		return nil, nil, errors.Errorf("unable to import %#v, we need at least 2 entries and there were %d entries", use, len(use))
	}

	var accum big.Float
	var postings []*Posting
	for _, f := range use {
		p := Posting{f.PostingBegLine, f.Account, f.Currency, f.Amount, f.State, f.PostingNotes, nil}
		postings = append(postings, &p)
		accum.Add(&accum, &f.Amount)
	}
	if !isNearZero(accum) {
		return nil, nil, errors.Errorf("unable to import %#v, they should sum to 0.0 but instead they summed to %s", use, accum.Text('f', 5))
	}

	first := imports[0]
	t := &Transaction{
		first.SrcFile,
		first.BegLine,
		first.Date,
		first.Code,
		first.Payee,
		first.TransNotes,
		postings,
	}
	return t.LinkPostings(), imports[end:], nil
}

var nearZero big.Float

func init() {
	nearZero.Parse("0.001", 10)
}

func isNearZero(f big.Float) bool {
	f.Abs(&f)
	c := f.Cmp(&nearZero)
	return c == -1 || c == 0
}

func parseAmount(s string) (currency string, amount big.Float, err error) {
	if !strings.HasPrefix(s, "$") {
		return "", amount, errors.New("unable to parse %q as an amount.  Currently we only support amounts with a leading $")
	}
	currency = "$"
	_, _, err = amount.Parse(s[1:], 10)
	return currency, amount, err
}

func (t *Transaction) DateText() string {
	return t.Date.Format("2006/01/02")
}

func (t *Transaction) String() string {
	var lines []string
	dateText := t.DateText()
	var tokens []string
	if t.Code == "" {
		tokens = []string{dateText, t.Payee}
	} else {
		tokens = []string{dateText, "(#" + t.Code + ")", t.Payee}
	}
	header := strings.Join(tokens, " ")
	lines = append(lines, header)
	for _, n := range t.Notes {
		lines = append(lines, fmt.Sprintf("%s; %s", indent, n))
	}
	for _, p := range t.Postings {
		lines = append(lines, p.String())
	}
	return strings.Join(lines, "\n")
}

// Flattened is meant as an aid to importing transactions from csv,
// where postings are flattened together with a transaction and appear
// on the same line.  Contains all the fields from Transaction and
// from Posting.
type Flattened struct {
	SrcFile    string
	BegLine    int
	Date       time.Time
	Code       string // optional.  The thing in parentheses.  e.g. check #
	Payee      string
	TransNotes []string // optional.  notes that apply to the entire transaction.

	PostingBegLine int
	Account        string
	Currency       string // must be "$" for now
	Amount         big.Float
	State          rune     // optional
	PostingNotes   []string // optional
}

func NewFlattened(srcFile string, begLine int, date time.Time, code string, payee string, transNotes []string, postingBegLine int, account string, currency string, amount big.Float, state rune, postingNotes []string) Flattened {
	return Flattened{
		srcFile, begLine, date, code, payee, transNotes, postingBegLine, account, currency, amount, state, postingNotes}
}

// sameTransaction returns true if the two entries appear to be from
// the same transaction.
func (f *Flattened) sameTransaction(o *Flattened) bool {
	if f.SrcFile != o.SrcFile {
		return false
	}
	if f.BegLine != o.BegLine {
		return false
	}
	return true
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
	return s[i].Date.Before(s[j].Date)
}
