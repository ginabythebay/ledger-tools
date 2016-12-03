package dup

import (
	"bytes"
	"fmt"
	"html/template"
	"strings"

	ledgertools "github.com/ginabythebay/ledger-tools"
)

type codePair struct {
	// the code that we matched on.  If one of the transactions
	// prefixed the code with a #, this Code will contain that prefix.
	Code string
	One  *ledgertools.Transaction
	Two  *ledgertools.Transaction
}

func newCodePair(one, two *ledgertools.Transaction) codePair {
	code := one.Code
	if strings.HasPrefix(two.Code, "#") {
		code = two.Code
	}
	return codePair{code, one, two}
}

var cpCompilerTemplate = template.Must(template.New("CompilerOutput").Parse(strings.TrimSpace(`
Code duplicate ({{.Code}})
	at {{.One.DateText}} {{.One.Payee}} ({{.One.SrcFile}}:{{.One.BegLine}})
	at {{.Two.DateText}} {{.Two.Payee}} ({{.Two.SrcFile}}:{{.Two.BegLine}})
`)))

func (p codePair) compilerText() (string, error) {
	var buf bytes.Buffer
	err := cpCompilerTemplate.Execute(&buf, p)
	return buf.String(), err
}

func (p codePair) isSuppressed() bool {
	return isDateSuppressed(
		p.One.DateText(),
		suppressCodeDuplicates,
		p.Two.Notes,
	) || isDateSuppressed(
		p.Two.DateText(),
		suppressCodeDuplicates,
		p.One.Notes,
	)
}

func (p codePair) accumXMLErrors(accum map[string]*file) {
	addCodeXMLError(accum, p.Code, p.One, p.Two)
	addCodeXMLError(accum, p.Code, p.Two, p.One)
}

func addCodeXMLError(accum map[string]*file, code string, p, other *ledgertools.Transaction) {
	f, ok := accum[p.SrcFile]
	if !ok {
		f = &file{
			Name: p.SrcFile,
		}
		accum[p.SrcFile] = f
	}
	msg := fmt.Sprintf("Possible duplicate of %s (%s) at %s:%d", other.DateText(), other.Code, other.SrcFile, other.BegLine)
	f.Errors = append(f.Errors, xmlError{
		Line:     p.BegLine,
		Severity: severity,
		Message:  msg,
		Source:   source,
	})
}
