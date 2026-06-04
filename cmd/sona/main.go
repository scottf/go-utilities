// Command sona is a CLI front-end for the Sonatype checker. Run it with no
// arguments to pick a project from an interactive list, or pass a project name
// / number / "all" to run non-interactively.
//
//	sona              # show the menu and pick one
//	sona 3            # run the 3rd project in the list
//	sona FLINK        # run the project named FLINK (case-insensitive)
//	sona all          # run every project
//	sona -f path.json # use a different config file
package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"go-utilities/internal/sonatype"
)

func main() {
	configPath := flag.String("f", defaultConfigPath(), "path to the sonatype-checker JSON config")
	debug := flag.Bool("debug", false, "print each URL before it is checked")
	flag.Parse()

	sonatype.Debug = *debug

	params, err := sonatype.LoadParams(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	if len(params) == 0 {
		fmt.Fprintf(os.Stderr, "no projects found in %s\n", *configPath)
		os.Exit(1)
	}

	selection := strings.Join(flag.Args(), " ")
	if selection == "" {
		selection = promptForSelection(params)
	}

	chosen, err := resolveSelection(params, selection)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	for _, p := range chosen {
		if err := sonatype.Check(p); err != nil {
			fmt.Fprintf(os.Stderr, "error checking %s: %v\n", p.Project, err)
			os.Exit(1)
		}
	}
}

func defaultConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ".sonatype-checker.json"
	}
	return filepath.Join(home, ".sonatype-checker.json")
}

// promptForSelection prints the numbered menu and reads a choice from stdin.
func promptForSelection(params []sonatype.CheckerParams) string {
	fmt.Println("Projects:")
	for i, p := range params {
		fmt.Printf("  %2d) %-16s %s:%s\n", i+1, p.Project, p.GroupID, p.Component)
	}
	fmt.Printf("  %2s) %s\n", "a", "all")
	fmt.Print("Choose a project (number, name, or 'a' for all): ")

	reader := bufio.NewReader(os.Stdin)
	line, _ := reader.ReadString('\n')
	return strings.TrimSpace(line)
}

// resolveSelection turns a selection string into the list of projects to run.
func resolveSelection(params []sonatype.CheckerParams, selection string) ([]sonatype.CheckerParams, error) {
	sel := strings.TrimSpace(selection)
	if sel == "" {
		return nil, fmt.Errorf("no selection made")
	}
	if sel == "a" || strings.EqualFold(sel, "all") {
		return params, nil
	}

	// Numeric selection (1-based).
	if n, err := strconv.Atoi(sel); err == nil {
		if n < 1 || n > len(params) {
			return nil, fmt.Errorf("choice %d out of range (1-%d)", n, len(params))
		}
		return []sonatype.CheckerParams{params[n-1]}, nil
	}

	// Name selection (case-insensitive).
	for _, p := range params {
		if strings.EqualFold(p.Project, sel) {
			return []sonatype.CheckerParams{p}, nil
		}
	}
	return nil, fmt.Errorf("no project matching %q", sel)
}
