package register

import (
	"fmt"
	"io"
)

type state int

const (
	// most common state. Outside of quoted field.
	start state = iota
	// in quoted field
	quoted
	// in quoted field and that previous character was a backslash
	escape
)

type converter struct {
	delegate  io.Reader
	buf       []byte // place we read into
	remaining []byte // what is still left to be read from
	escaped   []byte // if non-empty, contains raw bytes ready to be copied to output, before remaining
	s         state
}

func newConverter(r io.Reader) *converter {
	return &converter{
		delegate: r,
		buf:      make([]byte, 4092),
	}
}

func (c *converter) Read(p []byte) (n int, err error) {
	fmt.Printf("entered read.  c=%+v\n", c) // output for debug

	if len(c.escaped) != 0 {
		n = copy(p, c.escaped)
		c.escaped = c.escaped[n:]
		fmt.Printf("### escaped exit %d, len(c.escaped)=%d, %+v\n", n, len(c.escaped), p) // output for debug
		return n, nil
	}

	if len(c.remaining) == 0 {
		n, err = c.delegate.Read(c.buf)
		if n == 0 {
			fmt.Printf("### read error %+v\n", err) // output for debug

			return n, err
		}
		fmt.Printf("read %d bytes from delegate\n", n) // output for debug
		c.remaining = c.buf[:n]
	}

	i := 0 // cursor to p
	for i < len(p) && len(c.remaining) != 0 {
		next := c.remaining[0]
		c.remaining = c.remaining[1:]
		switch c.s {
		case start:
			p[i] = next
			i++
			if next == '"' {
				c.s = quoted
			}
		case quoted:
			switch next {
			case '"':
				p[i] = next
				i++
				c.s = start
			case '\\':
				c.s = escape
			default:
				p[i] = next
				i++
			}
		case escape:
			if next == '"' {
				c.escaped = []byte{'"', '"'}
			} else {
				c.escaped = []byte{next}
			}
			c.s = quoted
			fmt.Printf("### escaping exit %d, %+v, %+v\n", i, p, err) // output for debug
			return i, err
		}
	}

	fmt.Printf("### normal exit %d, %+v, %+v\n", i, p, err) // output for debug
	return i, err
}
