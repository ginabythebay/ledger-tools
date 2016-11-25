package dup

import (
	"encoding/xml"
	"fmt"
	"io"

	ledgertools "github.com/ginabythebay/ledger-tools"
)

const (
	version  = "7.2"
	severity = "warning"
	source   = "dupdetector"
)

type checkstyle struct {
	XMLName xml.Name `xml:"checkstyle"`
	Version string   `xml:"version,attr"`
	Files   []*file
}

type file struct {
	XMLName xml.Name `xml:"file"`
	Name    string   `xml:"name,attr"`
	Errors  []xmlError
}

type xmlError struct {
	XMLName  xml.Name `xml:"error"`
	Line     int      `xml:"line,attr"`
	Severity string   `xml:"severity,attr"`
	Message  string   `xml:"message,attr"`
	Source   string   `xml:"source,attr"`
}

func CheckStyleWriter(allPairs []Pair, w io.Writer) Writer {
	return func() error {
		accum := map[string]*file{}
		for _, p := range allPairs {
			add(accum, p.One, p.Two)
			add(accum, p.Two, p.One)
		}

		var allFiles []*file
		for _, f := range accum {
			allFiles = append(allFiles, f)
		}
		cs := checkstyle{
			Version: version,
			Files:   allFiles,
		}
		enc := xml.NewEncoder(w)
		enc.Indent("", "  ")
		return enc.Encode(&cs)
	}
}

func add(accum map[string]*file, p, other *ledgertools.Posting) {
	f, ok := accum[p.Xact.SrcFile]
	if !ok {
		f = &file{
			Name: p.Xact.SrcFile,
		}
		accum[p.Xact.SrcFile] = f
	}
	msg := fmt.Sprintf("Possible duplicate of %s %s %s at %s:%d", other.Xact.DateText(), other.AmountText(), other.Account, other.Xact.SrcFile, other.BegLine)
	f.Errors = append(f.Errors, xmlError{
		Line:     p.BegLine,
		Severity: severity,
		Message:  msg,
		Source:   source,
	})
}
