package csv

import (
	"encoding/csv"
	"io"

	"github.com/ginabythebay/ledger-tools/csv/ops"
)

// Process runs a batch of mutators against the csv file that is in
// reader and sends it to writer.
func Process(mutators []ops.Mutator, reader io.Reader, writer io.Writer) (cnt int, err error) {

	r := csv.NewReader(reader)
	w := csv.NewWriter(writer)

	lineNo := 0
	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		lineNo++
		if err != nil {
			return 0, err
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
	return lineNo, nil
}
