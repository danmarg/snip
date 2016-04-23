package main

import (
	"bufio"
	"fmt"
	"github.com/codegangsta/cli"
	"io"
	"io/ioutil"
	"os"
	"regexp"
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
func getInput(ctx *cli.Context) (io.Reader, error) {
	if len(ctx.Args()) > 1 {
		return os.Open(ctx.Args()[1])
	}
	return os.Stdin, nil
}

func doMatch(exp *regexp.Regexp, buf []byte, invert, onlymatching bool, w io.Writer) error {
	if onlymatching {
		for _, s := range exp.FindAll(buf, -1) {
			for _, b := range [][]byte{s, []byte("\n")} {
				if _, err := w.Write(b); err != nil {
					return err
				}
			}
		}
	} else {
		if onlymatching != exp.Match(buf) {
			for _, b := range [][]byte{buf, []byte("\n")} {
				if _, err := w.Write(b); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func match(exp *regexp.Regexp, invert, multiline, onlymatching bool, r io.Reader, w io.Writer) error {
	if invert && onlymatching {
		return fmt.Errorf("incompatible flags: --invert and --onlymatching")
	}
	if !multiline {
		// Line-by-line match.
		scn := bufio.NewScanner(r)
		for scn.Scan() {
			if err := doMatch(exp, scn.Bytes(), invert, onlymatching, w); err != nil {
				return err
			}
		}
	} else {
		// XXX: We could do a lot better job here.
		buf, err := ioutil.ReadAll(r)
		if err != nil {
			return err
		}
		return doMatch(exp, buf, invert, onlymatching, w)
	}

	return nil
}
