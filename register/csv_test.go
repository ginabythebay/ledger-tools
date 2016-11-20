package register

import (
	"fmt"
	"io"
	"io/ioutil"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"
)

func Test_converter_ReadSimple(t *testing.T) {
	c := newConverter(strings.NewReader("abcdef"))
	c.buf = make([]byte, 3)

	found, err := readAll(c)
	ok(t, err)
	equals(t, "abcdef", found)
}

func Test_converter_ReadQuoted(t *testing.T) {
	c := newConverter(strings.NewReader(`a"b"\c`))
	c.buf = make([]byte, 3)

	found, err := readAll(c)
	ok(t, err)
	equals(t, `a"b"\c`, found)
}

func Test_converter_ReadEscaped(t *testing.T) {
	c := newConverter(strings.NewReader(`a"b\"\\"c`))
	c.buf = make([]byte, 3)

	found, err := readAll(c)
	ok(t, err)
	equals(t, `a"b""\"c`, found)
}

func readAll(r io.Reader) (string, error) {
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// assert fails the test if the condition is false.
func assert(tb testing.TB, condition bool, msg string, v ...interface{}) {
	if !condition {
		_, file, line, _ := runtime.Caller(1)
		fmt.Printf("\033[31m%s:%d: "+msg+"\033[39m\n\n", append([]interface{}{filepath.Base(file), line}, v...)...)
		tb.FailNow()
	}
}

// ok fails the test if an err is not nil.
func ok(tb testing.TB, err error) {
	if err != nil {
		_, file, line, _ := runtime.Caller(1)
		fmt.Printf("\033[31m%s:%d: unexpected error: %s\033[39m\n\n", filepath.Base(file), line, err.Error())
		tb.FailNow()
	}
}

// equals fails the test if exp is not equal to act.
func equals(tb testing.TB, exp, act interface{}) {
	if !reflect.DeepEqual(exp, act) {
		_, file, line, _ := runtime.Caller(1)
		fmt.Printf("\033[31m%s:%d:\n\n\texp: %#v\n\n\tgot: %#v\033[39m\n\n", filepath.Base(file), line, exp, act)
		tb.FailNow()
	}
}
