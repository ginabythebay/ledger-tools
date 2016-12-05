package dup

import (
	"bytes"
	"fmt"
	"html/template"
	"strings"

	ledgertools "github.com/ginabythebay/ledger-tools"
)

var apCompilerTemplate = template.Must(template.New("CompilerOutput").Parse(strings.TrimSpace(`
Possible duplicate {{.One.AmountText}} {{.One.Account}}
	at {{.One.Xact.DateText}} {{.One.Xact.Payee}} ({{.One.Xact.SrcFile}}:{{.One.BegLine}})
	at {{.Two.Xact.DateText}} {{.Two.Xact.Payee}} ({{.Two.Xact.SrcFile}}:{{.Two.BegLine}})
`)))

type amountPair struct {
	One *ledgertools.Posting
	Two *ledgertools.Posting
}

func (p amountPair) compilerText() (string, error) {
	var buf bytes.Buffer
	err := apCompilerTemplate.Execute(&buf, p)
	return buf.String(), err
}

func (p amountPair) isSuppressed() bool {
	return isDateSuppressed(
		p.One.Xact.DateText(),
		suppressAmountDuplicates,
		p.Two.Notes,
	) || isDateSuppressed(
		p.Two.Xact.DateText(),
		suppressAmountDuplicates,
		p.One.Notes,
	)
}

func (p amountPair) accumXMLErrors(accum map[string]*file) {
	addAmountXMLError(accum, p.One, p.Two)
	addAmountXMLError(accum, p.Two, p.One)
}

func addAmountXMLError(accum map[string]*file, p, other *ledgertools.Posting) {
	f, ok := accum[p.Xact.SrcFile]
	if !ok {
		f = &file{
			Name: p.Xact.SrcFile,
		}
		accum[p.Xact.SrcFile] = f
	}
	msg := fmt.Sprintf("Possible duplicate of %s %s %s %s at %s:%d", other.Xact.DateText(), other.Xact.Payee, other.AmountText(), other.Account, other.Xact.SrcFile, other.BegLine)
	f.Errors = append(f.Errors, xmlError{
		Line:     p.BegLine,
		Severity: severity,
		Message:  msg,
		Source:   source,
	})
}
