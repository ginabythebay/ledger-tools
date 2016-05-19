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

func openInput(c *cli.Context, d *os.File) (in *os.File, err error) {
	in = d
	inputName := c.String("in")
	if inputName != "" {
		if in, err = os.Open(inputName); err != nil {
			return nil, err
		}
	}
	return in, nil
}

func openOutput(c *cli.Context, d *os.File) (out *os.File, err error) {
	out = d
	inputName := c.String("out")
	if inputName != "" {
		if out, err = os.Create(inputName); err != nil {
			return nil, err
		}
	}
	return out, nil
}

func cmdPrint(c *cli.Context) (result error) {
	in, err := openInput(c, os.Stdin)
	if err != nil {
		log.Fatal(err)
	}
	if in != os.Stdin {
		defer in.Close()
	}
	log.Printf("Reading from %q \n", in)
	csvIn := csv.NewReader(in)

	out, err := openOutput(c, os.Stdout)
	if err != nil {
		log.Fatal(err)
	}
	if out != os.Stdout {
		defer out.Close()
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

func cmdRereckon(c *cli.Context) (result error) {
	log.Println("TODO(gina)")
	return nil
}

func main() {
	app := cli.NewApp()
	app.Usage = "Augment ledger"

	app.Commands = []cli.Command{
		{
			Name: "rereckon",
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
					Name:  "c, config",
					Usage: "Name of yaml configuration file (required)",
				},
			},
			Usage:  "Post-process a reckon file",
			Action: cmdRereckon,
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
