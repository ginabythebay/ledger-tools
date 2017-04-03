package csv

import (
	"encoding/csv"
	"io"
	"log"

	"github.com/ginabythebay/ledger-tools/csv/ops"
)

// Process runs a batch of mutators against the csv file that is in
// reader and sends it to writer.
func Process(mutators []ops.Mutator, reader io.Reader, writer io.Writer, header []string) (cnt int, err error) {

	r := csv.NewReader(reader)
	w := csv.NewWriter(writer)

	extraLine := 0
	if header != nil {
		if err = w.Write(header); err != nil {
			return 0, err
		}
		extraLine = 1
	}

	lineNo := 0
	for {
		var record []string
		record, err = r.Read()
		if err == io.EOF {
			break
		}
		lineNo++
		if err != nil {
			parseErr, ok := err.(*csv.ParseError)
			if ok && parseErr.Err == csv.ErrFieldCount && len(record) == 1 {
				log.Printf("Skipping blank line %d", parseErr.Line)
				continue
			} else {
				return 0, err
			}
		}
		line := ops.NewLine(lineNo, record)

		for _, m := range mutators {
			err = m(line)
			if err != nil {
				return 0, err
			}
		}
		if err = w.Write(line.Record); err != nil {
			return 0, err
		}

	}
	w.Flush()
	if err = w.Error(); err != nil {
		return 0, err
	}
	return lineNo + extraLine, nil
}
