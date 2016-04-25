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

// getInput returns the input file or err.
func getInput(ctx *cli.Context, offset int) (*string, io.Reader, error) {
	if len(ctx.Args()) > offset {
		n := ctx.Args()[offset]
		f, e := os.Open(n)
		return &n, f, e
	}
	return nil, os.Stdin, nil
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

func match(fname *string, exp *regexp.Regexp,
	invert, multiline, onlymatching, filenames bool,
	r io.Reader, w io.Writer) error {
	if invert && onlymatching {
		return fmt.Errorf("incompatible flags: --invert and --onlymatching")
	}
	return doScan(multiline, r, w, func(buf []byte) [][]byte {
		if onlymatching {
			r := exp.FindAll(buf, -1)
			if filenames {
				for i, x := range r {
					// Add filename.
					r[i] = prefix(fname, x)
				}
			}
			return r
		} else {
			if onlymatching != exp.Match(buf) {
				if filenames {
					return [][]byte{prefix(fname, buf)}
				} else {
					return [][]byte{buf}
				}
			}
		}
		return nil
	})
}

func replace(exp *regexp.Regexp, repl string, multiline bool, r io.Reader, w io.Writer) error {
	return doScan(multiline, r, w, func(buf []byte) [][]byte {
		return [][]byte{exp.ReplaceAll(buf, []byte(repl))}
	})
}

func split(exp *regexp.Regexp, fields []int, multiline bool, r io.Reader, w io.Writer) error {
	return doScan(multiline, r, w, func(buf []byte) [][]byte {
		parts := exp.Split(string(buf), -1)
		ms := make([]string, len(fields))
		for i, f := range fields {
			if f < len(parts) {
				ms[i] = parts[f]
			}
		}
		return [][]byte{[]byte(strings.Join(ms, ","))}
	})
}
