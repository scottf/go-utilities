// Command ghhelp is a small GitHub Actions helper CLI.
//
//	ghhelp findJobFailures <jobUrl>   (or the shortcut: ghhelp fjf <jobUrl>)
//
// The GitHub token is read from the GH_TOKEN or GITHUB_TOKEN environment
// variable (or pass -token).
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

	if *token == "" {
		*token = firstNonEmptyEnv("GH_TOKEN", "GITHUB_TOKEN")
	}

	switch args[0] {
	case "findJobFailures", "fjf":
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "usage: ghhelp findJobFailures|fjf <jobUrl>")
			os.Exit(2)
		}
		if *token == "" {
			fmt.Fprintln(os.Stderr, "error: no GitHub token; set GH_TOKEN, GITHUB_TOKEN, or pass -token")
			os.Exit(1)
		}
		if err := ghhelp.FindJobFailures(*token, args[1], *verbose); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", args[0])
		usage()
		os.Exit(2)
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
	fmt.Fprintln(os.Stderr, "ghhelp - GitHub Actions helper")
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "commands:")
	fmt.Fprintln(os.Stderr, "  findJobFailures <jobUrl>   summarize failing tests from a job log (grouped by reason)")
	fmt.Fprintln(os.Stderr, "  fjf <jobUrl>               shortcut for findJobFailures (also a standalone 'fjf' command)")
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "flags:")
	flag.PrintDefaults()
}
