// Command fjf is a standalone shortcut for "ghhelp findJobFailures": it
// downloads a GitHub Actions job log and prints a summary of the failing
// tests, grouped by failure reason.
//
//	fjf [-v] [-token tok] <jobUrl>
//
// The GitHub token is read from the GH_TOKEN or GITHUB_TOKEN environment
// variable (or pass -token). Use -v for full failure blocks (stack traces).
package main

import (
	"flag"
	"fmt"
	"os"

	"go-utilities/internal/ghhelp"
)

func main() {
	token := flag.String("token", "", "GitHub token (defaults to $GH_TOKEN or $GITHUB_TOKEN)")
	verbose := flag.Bool("v", false, "verbose: print full failure blocks (stack traces) instead of the grouped summary")
	flag.Usage = usage
	flag.Parse()

	args := flag.Args()
	if len(args) < 1 {
		usage()
		os.Exit(2)
	}

	tok := *token
	if tok == "" {
		tok = firstNonEmptyEnv("GH_TOKEN", "GITHUB_TOKEN")
	}
	if tok == "" {
		fmt.Fprintln(os.Stderr, "error: no GitHub token; set GH_TOKEN, GITHUB_TOKEN, or pass -token")
		os.Exit(1)
	}

	if err := ghhelp.FindJobFailures(tok, args[0], *verbose); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func firstNonEmptyEnv(names ...string) string {
	for _, n := range names {
		if v := os.Getenv(n); v != "" {
			return v
		}
	}
	return ""
}

func usage() {
	fmt.Fprintln(os.Stderr, "fjf - shortcut for ghhelp findJobFailures")
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "usage: fjf [-v] [-token tok] <jobUrl>")
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "flags:")
	flag.PrintDefaults()
}
