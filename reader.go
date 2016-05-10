package main

import (
	"encoding/csv"
	"io"
	"log"
	"os"

	"github.com/codegangsta/cli"
)

// amountNames is the set of names of columns that contain amounts.
var amountNames = map[string]bool{
	"amount":  true,
	"balance": true,
}

// Format encapsulates what we know about the csv file.
type Format struct {
	amounts []int // indices to columns where amounts live
}

// NewFormat creates a Format based on a header record
func NewFormat(header []string) (format *Format) {
	amounts := []int{}
	for i, token := range header {
		if amountNames[token] {
			amounts = append(amounts, i)
		}
	}
	return &Format{amounts: amounts}
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

func cmdMain(c *cli.Context) (result error) {
	in := os.Stdin // default
	inputName := c.GlobalString("in")
	if inputName != "" {
		f, err := os.Open(inputName)
		if err != nil {
			log.Fatal(err)
		}
		in = f
		defer f.Close()
	}
	log.Printf("Reading from %q \n", in)
	csvIn := csv.NewReader(in)

	out := os.Stdout // default
	outputName := c.GlobalString("out")
	if outputName != "" {
		f, err := os.Create(outputName)
		if err != nil {
			log.Fatal(err)
		}
		out = f
		defer f.Close()
	}
	log.Printf("Writing to %q \n", out)
	csvOut := csv.NewWriter(out)

	var format *Format

	for {
		record, err := csvIn.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		if format == nil {
			format = NewFormat(record)
		} else {
			format.Apply(record)
		}
		csvOut.Write(record)
	}

	csvOut.Flush()

	return nil
}

func main() {
	app := cli.NewApp()
	app.Usage = "Augment ledger"
	app.Action = cmdMain
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "i, in",
			Usage: "Name of input file (default: stdin)",
		},
		cli.StringFlag{
			Name:  "o, out",
			Usage: "Name of output file (default: stdout)",
		},
	}
	app.Run(os.Args)
}
