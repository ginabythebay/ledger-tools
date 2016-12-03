package dup

import (
	"encoding/xml"
	"io"
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

func CheckStyleWriter(allDuplicates []Duplicate, w io.Writer) Writer {
	return func() error {
		accum := map[string]*file{}
		for _, d := range allDuplicates {
			d.accumXMLErrors(accum)
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
