package main

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"time"

	ledgertools "github.com/ginabythebay/ledger-tools"
	"github.com/ginabythebay/ledger-tools/csv"
	"github.com/ginabythebay/ledger-tools/csv/citi"
	"github.com/ginabythebay/ledger-tools/csv/ops"
	"github.com/ginabythebay/ledger-tools/csv/sffire"
	"github.com/ginabythebay/ledger-tools/csv/techcu"
	"github.com/ginabythebay/ledger-tools/dup"
	"github.com/ginabythebay/ledger-tools/gmail"
	"github.com/ginabythebay/ledger-tools/importer"
	"github.com/ginabythebay/ledger-tools/importer/amazon"
	"github.com/ginabythebay/ledger-tools/importer/github"
	"github.com/ginabythebay/ledger-tools/importer/kindle"
	"github.com/ginabythebay/ledger-tools/importer/lyft"
	"github.com/ginabythebay/ledger-tools/importer/parkmobile"
	"github.com/ginabythebay/ledger-tools/parser"
	"github.com/ginabythebay/ledger-tools/register"
	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

var csvTypes = map[string][]ops.Mutator{
	"citi":   citi.Mutators(),
	"sffire": sffire.Mutators(),
	"techcu": techcu.Mutators(),
}
var typeNames []string

func init() {
	for name := range csvTypes {
		typeNames = append(typeNames, name)
	}
}

type openStreams struct {
	all []io.Closer
}

func (s openStreams) Close() error {
	var firstError error
	for _, i := range s.all {
		e := i.Close()
		if firstError == nil && e != nil {
			firstError = e
		}
	}
	return firstError
}

func (s openStreams) add(c io.Closer) {
	s.all = append(s.all, c)
}

// Format encapsulates what we know about the csv file.
type Format struct {
	amounts []int // indices to columns where amounts live
}

// Apply modifies the record according to the structure we have
func (format *Format) Apply(record []string) {
	for _, i := range format.amounts {
		token := []rune(record[i])
		if len(token) > 2 { // ()
			first := token[0]
			last := token[len(token)-1]
			if first == '(' && last == ')' {
				token[0] = '-'
				token = token[0 : len(token)-1]
				record[i] = string(token)
			}
		}
	}
}

func openInput(name string, d *os.File) (in *os.File, err error) {
	in = d
	if name != "" {
		if in, err = os.Open(name); err != nil {
			return nil, err
		}
	}
	return in, nil
}

func openOutput(name string, d *os.File) (out *os.File, err error) {
	out = d
	if name != "" {
		if out, err = os.Create(name); err != nil {
			return nil, err
		}
	}
	return out, nil
}

func cmdPrint(c *cli.Context) (result error) {
	streams := openStreams{}
	defer streams.Close()

	in, err := openInput(c.String("in"), os.Stdin)
	if err != nil {
		log.Fatal(err)
	}
	if in != os.Stdin {
		streams.add(in)
	}
	log.Printf("Reading from %q \n", in)

	ledger, err := parser.ParseLedger(in)
	if err != nil {
		log.Fatal(err)
	}

	o, err := openOutput(c.String("out"), os.Stdout)
	if err != nil {
		log.Fatal(err)
	}
	if o != os.Stdout {
		streams.add(o)
	}
	log.Printf("Writing to %q \n", o)
	out := bufio.NewWriter(o)

	for i, t := range ledger {
		if i != 0 {
			fmt.Fprintln(out)
		}
		fmt.Fprintf(out, "%s\n", t)
	}

	out.Flush()

	return nil
}

var allGmailImporters = []importer.GmailImporter{
	parkmobile.GmailImporter,
	amazon.GmailImporter,
	github.GmailImporter,
	kindle.GmailImporter,
	lyft.GmailImporter,
}

func cmdGmail(c *cli.Context) (result error) {
	var allParsers []importer.Parser
	var allQuerySets []gmail.QuerySet
	for _, imp := range allGmailImporters {
		allParsers = append(allParsers, imp.Parsers...)
		allQuerySets = append(allQuerySets, imp.Queries...)
	}

	imp, err := msgImporter(allParsers)
	if err != nil {
		log.Fatalf("Get msg importer %+v", err)
	}

	gm, err := gmail.GetService()
	if err != nil {
		log.Fatalf("Get Gamil Service %+v", err)
	}
	var allTransactions []*ledgertools.Transaction

	var options []gmail.QueryOption
	if after := c.String("after"); after == "" {
		options = append(options, gmail.QueryNewerThan(c.Int("days")))
	} else {
		options = append(options, gmail.QueryAfter(after))
	}
	if before := c.String("before"); before != "" {
		options = append(options, gmail.QueryBefore(before))
	}

	for _, qs := range allQuerySets {
		var query []gmail.QueryOption
		query = append(query, options...)
		query = append(query, qs...)
		msgs, err := gm.QueryMessages(query...)
		if err != nil {
			log.Fatalf("Get mail %+v", err)
		}
		for _, m := range msgs {
			xact, err := imp.ImportMessage(m)
			if err != nil {
				log.Fatalf("Unable to import %#v\n %+v", m, err)
			}
			if xact == nil {
				log.Fatalf("Unable to recognize %#v", m)
			}
			allTransactions = append(allTransactions, xact)
		}
	}

	ledgertools.SortTransactions(allTransactions)
	for i, xact := range allTransactions {
		if i != 0 {
			fmt.Println()
		}
		fmt.Println(xact.String())
	}

	return nil
}

func combine(options []gmail.QueryOption, more ...gmail.QueryOption) []gmail.QueryOption {
	result := make([]gmail.QueryOption, 0, len(options)+len(more))
	result = append(result, options...)
	result = append(result, more...)
	return result
}

func msgImporter(allParsers []importer.Parser) (*importer.MsgImporter, error) {
	config, err := readRuleConfig()
	if err != nil {
		return nil, errors.Wrap(err, "readRuleConfig")
	}
	return importer.NewMsgImporter(config, allParsers)
}

func readRuleConfig() ([]byte, error) {
	usr, err := user.Current()
	if err != nil {
		return nil, errors.Wrap(err, "get user")
	}
	ruleFileName := filepath.Join(usr.HomeDir, ".config", "ledger-tools", "rules.yaml")
	return ioutil.ReadFile(ruleFileName)
}

func cmdCsv(c *cli.Context) (result error) {
	if c.String("type") == "" {
		log.Fatalf("You must set the -type flag.  Valid values are [%s]", strings.Join(typeNames, ", "))
	}
	csvType := c.String("type")
	mutators := csvTypes[csvType]
	if mutators == nil {
		log.Fatalf("Unexpected csv type %q.  Valid types are [%s]", csvType, strings.Join(typeNames, ", "))
	}

	streams := openStreams{}
	defer streams.Close()

	in, err := openInput(c.String("in"), os.Stdin)
	if err != nil {
		log.Fatal(err)
	}
	if in != os.Stdin {
		streams.add(in)
	}

	o, err := openOutput(c.String("out"), os.Stdout)
	if err != nil {
		log.Fatal(err)
	}
	if o != os.Stdout {
		streams.add(o)
	}

	var headers []string
	if csvType == "citi" {
		headers = citi.Headers
	}

	cnt, err := csv.Process(mutators, in, o, headers)
	if err != nil {
		log.Fatal(err)
	}
	if o != os.Stdout {
		fmt.Printf("Wrote %d lines to %s\n", cnt, o.Name())
	}

	return nil
}

func cmdLint(c *cli.Context) (result error) {
	checkStyle := c.Bool("checkstyle")
	start := time.Now()
	allTrans, err := register.Read(c.String("file"))
	if err != nil {
		log.Fatal(err)
	}
	if !checkStyle {
		fmt.Printf("Read %d transactions in %s\n", len(allTrans), time.Since(start))
	}

	finder := dup.NewFinder(c.Int("dupdays"))
	for _, t := range allTrans {
		finder.Add(t)
	}

	if checkStyle {
		err = finder.WriteCheckStyle(os.Stdout)
	} else {
		err = finder.WriteJavacStyle(os.Stdout)
	}
	if err != nil {
		log.Fatal(err)
	}

	return nil
}

func main() {
	app := cli.NewApp()
	app.Usage = "Augment ledger"

	app.Commands = []cli.Command{
		{
			Name: "csv",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "i, in",
					Usage: "Name of input file (default: stdin)",
				},
				cli.StringFlag{
					Name:  "o, out",
					Usage: "Name of output file (default: stdout)",
				},
				cli.StringFlag{
					Name:  "t, type",
					Usage: fmt.Sprintf("Type of file we are processing.  Must be one of [%s]]", strings.Join(typeNames, ", ")),
				},
			},
			Usage:  "Process a csv file, making it ready for ledger convert",
			Action: cmdCsv,
		},
		{
			Name:   "lint",
			Usage:  "EXPERIMENTAL: Look for potentially duplicate postings",
			Action: cmdLint,
			Flags: []cli.Flag{
				cli.IntFlag{
					Name:  "d, dupdays",
					Value: 3,
					Usage: "Number of days to consider when looking for possible duplicate postings.  A value of 0 will consider only same-day postings.",
				},
				cli.BoolFlag{
					Name:  "c, checkstyle",
					Usage: "Uses checkstyle-compatible output",
				},
				cli.StringFlag{
					Name:  "f, file",
					Usage: "Name of file to lint.  If not specified, the default ledger file will be used.",
				},
			},
		},
		{
			Name: "gmail",
			Flags: []cli.Flag{
				cli.IntFlag{
					Name:  "d, days",
					Value: 30,
					Usage: "Query for emails newer than this many days.  Ignored if --after is set.",
				},
				cli.StringFlag{
					Name:  "a, after",
					Usage: "Query for emails after this date.  Example value \"2004/04/16\".  Setting this will cause --days to be ignored.",
				},
				cli.StringFlag{
					Name:  "b, before",
					Usage: "Query for emails before this date.  Example value \"2004/04/18\".",
				},
			},
			Usage:  "Process gmail",
			Action: cmdGmail,
		},
		{
			Name: "print",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "i, in",
					Usage: "Name of input file (default: stdin)",
				},
				cli.StringFlag{
					Name:  "o, out",
					Usage: "Name of output file (default: stdout)",
				},
			},
			Usage:  "Read a reckon file and print it",
			Action: cmdPrint,
		},
	}
	app.Run(os.Args)
}
