package main

import (
	"bufio"
	"fmt"
	"github.com/codegangsta/cli"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
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
func getInput(ctx *cli.Context, offset int, recursive bool) ([]io.Reader, error) {
	if len(ctx.Args()) > offset {
		n := ctx.Args()[offset]
		if s, err := os.Stat(n); err == nil {
			if s.IsDir() {
				if recursive {
					r := []io.Reader{}
					filepath.Walk(n, func(p string, i os.FileInfo, err error) error {
						if err != nil {
							return err
						}
						if !i.IsDir() {
							f, e := os.Open(p)
							if e != nil {
								return e
							}
							r = append(r, f)
						}
						return nil
					})
					return r, nil
				}
				return nil, fmt.Errorf("directory given but --recursive not specified")
			}
			f, e := os.Open(n)
			return []io.Reader{f}, e
		} else {
			return nil, err
		}
	}
	return []io.Reader{os.Stdin}, nil
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

func doScan(multiline bool, rs []io.Reader, w io.Writer, proc func([]byte) [][]byte) error {
	for _, r := range rs {
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
			if err := writeln(proc(buf), w, false); err != nil {
				return err
			}
		}
	}
	return nil
}

func match(exp *regexp.Regexp, invert, multiline, onlymatching bool, r []io.Reader, w io.Writer) error {
	if invert && onlymatching {
		return fmt.Errorf("incompatible flags: --invert and --onlymatching")
	}
	return doScan(multiline, r, w, func(buf []byte) [][]byte {
		if onlymatching {
			return exp.FindAll(buf, -1)
		} else {
			if onlymatching != exp.Match(buf) {
				return [][]byte{buf}
			}
		}
		return nil
	})
}

func replace(exp *regexp.Regexp, repl string, multiline bool, r []io.Reader, w io.Writer) error {
	return doScan(multiline, r, w, func(buf []byte) [][]byte {
		return [][]byte{exp.ReplaceAll(buf, []byte(repl))}
	})
}

func split(exp *regexp.Regexp, fields []int, multiline bool, r []io.Reader, w io.Writer) error {
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
