package register

import (
	"encoding/csv"
	"fmt"
	"io"
	"io/ioutil"
	"math/big"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"

	ledgertools "github.com/ginabythebay/ledger-tools"
	"github.com/pkg/errors"
)

var csvFormat = strings.Join(
	[]string{
		`%(quoted(filename)),`,
		`%(quoted(xact.beg_line)),`,
		`%(quoted(join(xact.note))),`,
		`%(quoted(date)),`,
		`%(quoted(code)),`,
		`%(quoted(payee)),`,
		`%(quoted(beg_line)),`,
		`%(quoted(display_account)),`,
		`%(quoted(commodity(scrub(display_amount)))),`,
		`%(quoted(quantity(scrub(display_amount)))),`,
		`%(quoted(cleared ? "*" : (pending ? "!" : ""))),`,
		`%(quoted(join(note)))`,
		`\n`,
	},
	"")

// offsets into csv records to extract data.  Must match csvFormat, above
const (
	colFilename = iota
	colTransBegLine
	colTransNote
	colDate
	colCode
	colPayee
	colBegLine
	colAccount
	colCurrency
	colAmount
	colState
	colPostingNote
)

const dateLayout = "2006/01/02"

// Read reads the default register file.  Depends on calling ledger.
func Read(filename string) ([]*ledgertools.Transaction, error) {
	ledger, err := exec.LookPath("ledger")
	if err != nil {
		return nil, errors.Wrap(err, "lookpath")
	}

	cmd := exec.Command(ledger, "csv", "--csv-format", csvFormat)
	if filename != "" {
		cmd.Args = append(cmd.Args, "-f", filename)
	}

	outPipe, err := cmd.StdoutPipe()
	if err != nil {
		return nil, errors.Wrap(err, "StdoutPipe")
	}
	var wg sync.WaitGroup
	wg.Add(1)

	var result []*ledgertools.Transaction
	var readErr error
	go func() {
		defer func() {
			// clear stdout so that cmd.Run will complete, even if we
			// had an error partway through
			_, _ = io.Copy(ioutil.Discard, outPipe)
			wg.Done()
		}()

		result, readErr = ReadLedgerCsv(outPipe)
	}()

	errPipe, err := cmd.StderrPipe()
	if err != nil {
		return nil, errors.Wrap(err, "StderrPipe")
	}
	go func() {
		_, _ = io.Copy(os.Stderr, errPipe)
	}()

	err = cmd.Run()
	if err != nil {
		return nil, errors.Wrap(err, "Run")
	}
	wg.Wait()
	if readErr != nil {
		return nil, errors.Wrap(readErr, "read")
	}

	return result, nil
}

// ReadLedgerCsv knows how to read ledger-style csv files (where
// things are escaped by backlslashes)
func ReadLedgerCsv(source io.ReadCloser) ([]*ledgertools.Transaction, error) {
	var result []*ledgertools.Transaction
	converted := newConverter(source)

	// TODO(gina) look into reworking NextTransaction api to be
	// more stream-friendly.  Currently we read everything into
	// memory as Flatteneds so we can convert that to another
	// entire in-memory version, of Transactions

	var allFlattened []ledgertools.Flattened
	r := csv.NewReader(converted)
	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, errors.Wrap(err, "csv read")
		}

		// TODO(gina) look into using string interning here.
		// Except for notes, we are going to see lots of the same
		// strings over and over and we don't need to put so much
		// memory pressure on the runtime for immutable data.

		var f ledgertools.Flattened
		f, err = parse(record)
		if err != nil {
			return nil, errors.Wrap(err, "parse")
		}

		allFlattened = append(allFlattened, f)
	}

	for len(allFlattened) != 0 {
		var t *ledgertools.Transaction
		var err error
		t, allFlattened, err = ledgertools.NextTransaction(allFlattened)
		if err != nil {
			return nil, errors.Wrap(err, "NextTransaction")
		}
		result = append(result, t)
	}

	return result, nil
}

func parse(record []string) (ledgertools.Flattened, error) {
	var f ledgertools.Flattened
	transBeginLine, err := strconv.Atoi(record[colTransBegLine])
	if err != nil {
		return f, errors.Wrapf(err, "convert %s to int", record[colTransBegLine])
	}
	postingBeginLine, err := strconv.Atoi(record[colBegLine])
	if err != nil {
		return f, errors.Wrapf(err, "convert %s to int", record[colBegLine])
	}

	var date time.Time
	date, err = time.Parse(dateLayout, record[colDate])
	if err != nil {
		return f, errors.Wrapf(err, "convert %s to date", record[colDate])
	}

	transNote := strings.Split(record[colTransNote], "\n")
	postingNote := strings.Split(record[colPostingNote], "\n")

	var amount big.Float
	if _, ok := amount.SetString(record[colAmount]); !ok {
		if err != nil {
			return f, fmt.Errorf("convert %q to big.Float", record[colAmount])
		}
	}

	state := ' '
	if len(record[colState]) > 0 {
		state = []rune(record[colState])[0]
	}

	f = ledgertools.NewFlattened(
		record[colFilename],
		transBeginLine,
		date,
		record[colCode],
		record[colPayee],
		transNote,
		postingBeginLine,
		record[colAccount],
		record[colCurrency],
		amount,
		state,
		postingNote,
	)
	return f, nil
}
