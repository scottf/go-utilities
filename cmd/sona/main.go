// Command sona is a CLI front-end for the Sonatype checker. Run it with no
// project name to print the help, the config file location, and the list of
// configured projects; pass a project name to check that one project.
//
//	sona                       # help + config path + project list
//	sona FLINK                 # check the project named FLINK (case-insensitive)
//	sona -d FLINK              # same, but print each URL as it is checked
//	sona -c other.json FLINK   # use a different config file
//
// Command shape: sona [-d] [-c path-to-config] project-name
// The -d and -c flags may appear in any order; the project name is the lone
// non-flag argument.
package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"go-utilities/internal/sonatype"
)

// options holds the parsed command line.
type options struct {
	debug      bool
	configPath string
	project    string
	help       bool
}

func main() {
	opts, err := parseArgs(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		printUsageLine(os.Stderr)
		os.Exit(2)
	}

	sonatype.Debug = opts.debug

	if opts.help {
		printHelp(os.Stdout)
		return
	}

	absPath := absConfigPath(opts.configPath)

	params, err := sonatype.LoadParams(opts.configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	if len(params) == 0 {
		fmt.Fprintf(os.Stderr, "no projects found in %s\n", absPath)
		os.Exit(1)
	}

	// No project name: show the help, then the config location and projects.
	if opts.project == "" {
		printHelp(os.Stdout)
		fmt.Fprintln(os.Stdout)
		listProjects(absPath, params)
		return
	}

	// A name: run that project (case-insensitive match).
	for _, p := range params {
		if strings.EqualFold(p.Project, opts.project) {
			if err := sonatype.Check(p); err != nil {
				fmt.Fprintf(os.Stderr, "error checking %s: %v\n", p.Project, err)
				os.Exit(1)
			}
			return
		}
	}
	fmt.Fprintf(os.Stderr, "no project matching %q (run 'sona' to list projects)\n", opts.project)
	os.Exit(1)
}

// parseArgs reads the command line. -d and -c may appear in any order; -c
// consumes the following argument as its path. The single non-flag argument is
// the project name.
func parseArgs(args []string) (options, error) {
	opts := options{configPath: defaultConfigPath()}
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "-h" || arg == "-help" || arg == "--help":
			opts.help = true
		case arg == "-d":
			opts.debug = true
		case arg == "-c":
			if i+1 >= len(args) {
				return opts, fmt.Errorf("-c requires a path-to-config argument")
			}
			i++
			opts.configPath = args[i]
		case strings.HasPrefix(arg, "-"):
			return opts, fmt.Errorf("unknown flag %q", arg)
		default:
			if opts.project != "" {
				return opts, fmt.Errorf("unexpected argument %q; the project name must be a single token", arg)
			}
			opts.project = arg
		}
	}
	return opts, nil
}

func defaultConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ".sonatype-checker.json"
	}
	return filepath.Join(home, ".sonatype-checker.json")
}

// absConfigPath resolves the config path to an absolute path for display,
// falling back to the given path if resolution fails.
func absConfigPath(p string) string {
	if abs, err := filepath.Abs(p); err == nil {
		return abs
	}
	return p
}

// listProjects prints the config location followed by every configured project.
func listProjects(configPath string, params []sonatype.CheckerParams) {
	fmt.Printf("Config: %s\n\n", configPath)
	fmt.Println("Projects:")

	width := 0
	for _, p := range params {
		if len(p.Project) > width {
			width = len(p.Project)
		}
	}
	for _, p := range params {
		fmt.Printf("  %-*s  %s:%s\n", width, p.Project, p.GroupID, p.Component)
	}
}

func printUsageLine(out io.Writer) {
	fmt.Fprintln(out, "usage: sona [-d] [-c path-to-config] project-name")
}

// printHelp writes the thorough help message to out.
func printHelp(out io.Writer) {
	fmt.Fprintf(out, `sona — check when Maven artifacts were last published to Sonatype.

For each configured project, sona fetches release metadata from repo1.maven.org
and snapshot metadata from central.sonatype.com, expands each artifact across
its configured JDK qualifiers, and prints when each one was last updated
(timestamps are RFC3339 UTC).

Usage:
  sona [-d] [-c path-to-config] project-name

  Run with no project name to print this help followed by the config file
  location and the list of configured projects. The -d and -c flags may appear
  in any order; the project name is the lone non-flag argument.

Flags:
  -d           debug: print each URL before it is checked
  -c <path>    read projects from an alternate config file
               (default %s)

Config file:
  A JSON array of projects with snake_case keys: project, group_id, component,
  jdk_qualifiers, release_versions, snapshot_versions. A null entry in
  jdk_qualifiers means the bare artifact (no -jdkNN suffix); e.g.
  [null,"17","21","25"] checks component, component-jdk17, -jdk21, and -jdk25.
`, defaultConfigPath())
}
