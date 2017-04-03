package ops

import (
	"fmt"
	"regexp"
	"strings"
)

// negate returns the negative version of the input.  We avoid any
// conversion to a number her and just treat it as a raw string.
func negate(s string) string {
	if strings.HasPrefix(s, "-") {
		return s[1:]
	}
	return "-" + s
}

// Mutator is a function that knows how to mutate a Line
type Mutator func(l *Line) error

// Line represents a line we read from a csv file
type Line struct {
	LineNo int
	Record []string
}

// NewLine creates a Line structure
func NewLine(lineNo int, record []string) *Line {
	return &Line{lineNo, record}
}

// ReplaceHeader changes the first line
func ReplaceHeader(header []string) Mutator {
	return func(l *Line) error {
		if l.LineNo == 1 {
			l.Record = header
		}
		return nil
	}
}

func MoveAndNegateIfPresent(from int, to int) Mutator {
	return func(l *Line) error {
		if l.LineNo != 1 {
			value := l.Record[from]
			if value != "" {
				l.Record[to] = negate(value)
				l.Record[from] = ""
			}
		}
		return nil
	}
}

func Negate(i int) Mutator {
	return func(l *Line) error {
		if l.LineNo != 1 {
			value := l.Record[i]
			l.Record[i] = negate(value)
		}
		return nil
	}
}

// CheckWithdrawal converts an entry like 'Check Withdrawal: #999999'
// to '(#999999)'
func CheckWithdrawal(i int) Mutator {
	r := regexp.MustCompile(`^Check Withdrawal: (#\d+).*$`)
	return func(l *Line) error {
		if l.LineNo != 1 {
			value := l.Record[i]
			if matches := r.FindStringSubmatch(value); matches != nil {
				l.Record[i] = fmt.Sprintf("(%s)", matches[1])
			}
		}
		return nil
	}
}

// RemoveText removes the substring rem, if it is present, up to one time.
func RemoveText(i int, rem string) Mutator {
	return func(l *Line) error {
		if l.LineNo != 1 {
			value := l.Record[i]
			l.Record[i] = strings.Replace(value, rem, "", 1)
		}
		return nil
	}
}

func DeparenNegatives(i int) Mutator {
	return func(l *Line) error {
		if l.LineNo != 1 {
			value := l.Record[i]
			if strings.HasPrefix(value, "(") && strings.HasSuffix(value, ")") {
				value = strings.TrimPrefix(value, "(")
				value = strings.TrimSuffix(value, ")")
				value = negate(value)
				l.Record[i] = value
			}
		}
		return nil
	}
}

func EnsureDollars(i int) Mutator {
	return func(l *Line) error {
		if l.LineNo != 1 {
			value := l.Record[i]
			if value != "" && (!strings.HasPrefix(value, "$")) {
				l.Record[i] = "$" + value
			}
		}
		return nil
	}
}

func StripCommas(i int) Mutator {
	return func(l *Line) error {
		if l.LineNo != 1 {
			value := l.Record[i]
			if strings.Contains(value, ",") {
				l.Record[i] = strings.Replace(value, ",", "", -1)
			}
		}
		return nil
	}
}

func StripSuffix(i int, suffix string) Mutator {
	return func(l *Line) error {
		if l.LineNo != 1 {
			l.Record[i] = strings.TrimSuffix(l.Record[i], suffix)
		}
		return nil
	}
}

// StripNewlines removes any newlines in a record
func StripNewlines(l *Line) error {
	for i, c := range l.Record {
		if strings.Contains(c, "\n") {
			l.Record[i] = strings.Replace(c, "\n", "", -1)
		}
	}
	return nil
}
