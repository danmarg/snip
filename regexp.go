package main

import (
	"bufio"
	"fmt"
	"github.com/codegangsta/cli"
	"io"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
)

// getPattern returns the specified pattern or nil.
func getPattern(ctx *cli.Context) (*regexp.Regexp, error) {
	if !ctx.Args().Present() {
		return nil, fmt.Errorf("missing pattern")
	}
	exp := ctx.Args().First()
	// Assemble flags.
	fs := ""
	for _, f := range []string{"i", "m", "s", "U"} {
		if ctx.GlobalBool(f) {
			fs += f
		}
	}
	if fs != "" {
		exp = "(?" + fs + ")" + exp
	}

	return regexp.Compile(exp)
}

type namedReader struct {
	Name   *string
	Reader io.Reader
}

// getInput returns the input reader(s) or error.
func getInput(ctx *cli.Context, offset int) ([]namedReader, error) {
	if len(ctx.Args()) > offset {
		ms := ctx.Args()[offset:]
		ns := make([]namedReader, len(ms))
		for i, m := range ms {
			m := m
			f, err := os.Open(m)
			if err != nil {
				return nil, err
			}
			ns[i] = namedReader{&m, f}
		}
		return ns, nil
	} else {
		return []namedReader{{nil, os.Stdin}}, nil
	}
}

func writeln(ms [][]byte, w io.Writer, newline bool) error {
	for _, m := range ms {
		if _, err := w.Write(m); err != nil {
			return err
		}
		if newline {
			if _, err := w.Write([]byte("\n")); err != nil {
				return err
			}
		}
	}
	return nil
}

func doScan(multiline bool, r io.Reader, w io.Writer, proc func([]byte) [][]byte) error {
	if !multiline {
		// Line-by-line match.
		scn := bufio.NewScanner(r)
		for scn.Scan() {
			if err := writeln(proc(scn.Bytes()), w, true); err != nil {
				return err
			}
		}
	} else {
		// XXX: We could do a lot better job here.
		buf, err := ioutil.ReadAll(r)
		if err != nil {
			return err
		}
		return writeln(proc(buf), w, false)
	}
	return nil
}

func prefix(fname *string, line []byte) []byte {
	if fname == nil {
		return line
	}
	return append([]byte(*fname+": "), line...)
}

func match(exp *regexp.Regexp,
	invert, multiline, onlymatching, filenames bool,
	rs []namedReader, w io.Writer) error {
	if invert && onlymatching {
		return fmt.Errorf("incompatible flags: --invert and --onlymatching")
	}
	for _, r := range rs {
		if err := doScan(multiline, r.Reader, w, func(buf []byte) [][]byte {
			if onlymatching {
				ms := exp.FindAll(buf, -1)
				if filenames {
					for i, x := range ms {
						// Add filename.
						ms[i] = prefix(r.Name, x)
					}
				}
				return ms
			} else {
				if onlymatching != exp.Match(buf) {
					if filenames {
						return [][]byte{prefix(r.Name, buf)}
					} else {
						return [][]byte{buf}
					}
				}
			}
			return nil
		}); err != nil {
			return err
		}
	}
	return nil
}

func replace(exp *regexp.Regexp, repl string, multiline bool, rs []namedReader, w io.Writer) error {
	for _, r := range rs {
		if err := doScan(multiline, r.Reader, w, func(buf []byte) [][]byte {
			return [][]byte{exp.ReplaceAll(buf, []byte(repl))}
		}); err != nil {
			return err
		}
	}
	return nil
}

func split(exp *regexp.Regexp, fields []int, multiline bool, rs []namedReader, w io.Writer) error {
	for _, r := range rs {
		if err := doScan(multiline, r.Reader, w, func(buf []byte) [][]byte {
			parts := exp.Split(string(buf), -1)
			ms := make([]string, len(fields))
			for i, f := range fields {
				if f < len(parts) {
					ms[i] = parts[f]
				}
			}
			return [][]byte{[]byte(strings.Join(ms, ","))}
		}); err != nil {
			return err
		}
	}
	return nil
}
