package importer

import (
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	ledgertools "github.com/ginabythebay/ledger-tools"
	"github.com/ginabythebay/ledger-tools/rules"
	"github.com/pkg/errors"
)

const amountAlignmentCol = 65
const indent = "    "

// Rule inputs
const (
	instrumentKey = "Instrument"
	payeeKey      = "Payee"
)

// Rule outputs
const (
	costAccountKey    = "CostAccount"
	paymentAccountKey = "PaymentAccount"
)

var (
	validInputs = []string{instrumentKey, payeeKey}
	validOuputs = []string{costAccountKey, paymentAccountKey}
)

type Parser func(msg ledgertools.Message) (*Parsed, error)

// MsgImporter knows how to import messages
type MsgImporter struct {
	rs         *rules.RuleSet
	allParsers []Parser
}

// NewMsgImporter creates a new MsgImporter
func NewMsgImporter(ruleConfig []byte, allParsers []Parser) (*MsgImporter, error) {
	rs, err := rules.From(ruleConfig, validInputs, validOuputs)
	if err != nil {
		return nil, errors.Wrap(err, "reading rule config")
	}
	return &MsgImporter{rs, allParsers}, nil
}

// ImportMessage imports an email message and produces a Transaction.
// nil will be returned if the email message is of a type we don't
// recognize
func (mi *MsgImporter) ImportMessage(msg ledgertools.Message) (*Transaction, error) {

	var parsed *Parsed
	var err error
	for _, parser := range mi.allParsers {
		parsed, err = parser(msg)
		if err != nil {
			return nil, errors.Wrapf(err, "%s", parser)
		}
		if parsed != nil {
			break
		}
	}
	if parsed == nil {
		return nil, nil
	}

	result, err := parsed.transaction(mi.rs)
	if err != nil {
		return nil, errors.Wrap(err, "transaction")
	}
	return result, nil
}

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
		rules.Input(instrumentKey, p.PaymentInstrument),
		rules.Input(payeeKey, p.Payee))

	if costAccount = mappings.Get(costAccountKey); costAccount == "" {
		return nil, errors.Errorf("Unable to determine %q for payee %q.  rs=%#v", costAccountKey, p.Payee, rs)
	}
	if paymentAccount = mappings.Get(paymentAccountKey); paymentAccount == "" {
		return nil, errors.Errorf("Unable to determine %q for instrument %q.  rs=%#v", paymentAccountKey, p.PaymentInstrument, rs)
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
