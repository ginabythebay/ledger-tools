package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/codegangsta/cli"
	"github.com/ginabythebay/ledger-tools/parser"
)

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

func cmdRereckon(c *cli.Context) (result error) {
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
