package main

import (
	"fmt"
	"github.com/codegangsta/cli"
	"os"
	"strconv"
	"strings"
)

func main() {
	app := cli.NewApp()
	app.Name = "snip"
	app.Usage = "snip, cut, trim, chop: a lovechild of grep and sed."
	app.Author = "Daniel Margolis"
	app.Version = "o_0"

	// Global options.
	app.Flags = []cli.Flag{
		// re2 flags
		cli.BoolFlag{
			Name:  "insensitive, i",
			Usage: "case insensitive",
		},
		cli.BoolFlag{
			Name:  "multiline, m",
			Usage: "multiline",
		},
		cli.BoolFlag{
			Name:  "dotall, s",
			Usage: "let . match \\n",
		},
		cli.BoolFlag{
			Name:  "ungreedy, U",
			Usage: "swap meaning of x* and x*?, x+ and x+?",
		},
		// input control
		cli.BoolFlag{
			Name:  "recursive, r",
			Usage: "recursive",
		},
	}

	// Commands.
	app.Commands = []cli.Command{
		{
			Name:      "match",
			ShortName: "m",
			Usage:     "[pattern] [file]? regular expression match",
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "invert, v",
					Usage: "invert matches",
				},
				cli.BoolFlag{
					Name:  "onlymatching, o",
					Usage: "output only matching",
				},
			},
			// Use After so we get error handling for free.
			Action: func(ctx *cli.Context) {},
			After: func(ctx *cli.Context) error {
				exp, err := getPattern(ctx)
				if err != nil {
					return err
				}
				in, err := getInput(ctx, 1, ctx.GlobalBool("r"))
				if err != nil {
					return err
				}
				return match(exp,
					ctx.Bool("v"), ctx.GlobalBool("m"), ctx.Bool("o"),
					in, os.Stdout)
			},
		},
		{
			Name:      "replace",
			ShortName: "s",
			Usage:     "[pattern] [pattern] [file]? regular expression replace",
			Action:    func(ctx *cli.Context) {},
			After: func(ctx *cli.Context) error {
				exp, err := getPattern(ctx)
				if err != nil {
					return err
				}
				if len(ctx.Args()) < 2 {
					return fmt.Errorf("missing required replacement pattern")
				}
				repl := ctx.Args().Tail()[0]
				in, err := getInput(ctx, 2, ctx.GlobalBool("r"))
				if err != nil {
					return err
				}
				return replace(exp, repl, ctx.GlobalBool("m"), in, os.Stdout)
			},
		},
		{
			Name:      "split",
			ShortName: "c",
			Usage:     "[pattern] [file]? split input lines",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "fields, f",
					Usage: "fields to output (1-indexed)",
				},
			},
			Action: func(ctx *cli.Context) {},
			After: func(ctx *cli.Context) error {
				exp, err := getPattern(ctx)
				if err != nil {
					return err
				}
				in, err := getInput(ctx, 1, ctx.GlobalBool("r"))
				if err != nil {
					return err
				}
				fs := ctx.String("f")
				fs_ := strings.Split(fs, ",")
				fields := make([]int, len(fs_))
				for i, f := range fs_ {
					if f, err := strconv.ParseInt(f, 10, 32); err != nil || f < 1 {
						return fmt.Errorf("invalid field value " + fs_[i])
					} else {
						fields[i] = int(f - 1)
					}
				}
				return split(exp, fields, ctx.GlobalBool("m"), in, os.Stdout)
			},
		},
	}

	app.RunAndExitOnError()
}
