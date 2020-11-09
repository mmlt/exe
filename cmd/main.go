package main

import (
	"flag"
	"fmt"
	"github.com/go-logr/logr"
	"github.com/go-logr/stdr"
	"github.com/mmlt/exe"
	stdlog "log"
	"os"
	"path/filepath"
)

// Version of the recording tool.
var Version string

func main() {
	var directory string
	flag.StringVar(&directory, "d", "",
		`Directory to store recordings`)
	var verbosity int
	flag.IntVar(&verbosity, "v", 0,
		`Log verbosity, higher numbers produce more output`)
	var version bool
	flag.BoolVar(&version, "version", false,
		"Print version")
	var hlp bool
	flag.BoolVar(&hlp, "help", false,
		`Help page`)

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s -d <path> -- <command> [<args>]\n\n", os.Args[0])
		flag.PrintDefaults()
	}

	flag.Parse()

	if hlp {
		const help = `%[1]s records the stdout/stderr of a command into a given directory.
The recordings can be used in unit-tests and fakes.

Example:
  %[1]s -d /tmp/test -- ls -a
    Run ls -a and stores the recorded output in the /tmp/test directory.

See https://github.com/mmlt/exe
`
		fmt.Fprintf(os.Stderr, help, filepath.Base(os.Args[0]), Version)
		os.Exit(0)
	}

	if version {
		fmt.Println(filepath.Base(os.Args[0]), Version)
		os.Exit(0)
	}

	if directory == "" {
		fmt.Fprint(os.Stderr, "Error: -d must be set.\n\n")
		flag.Usage()
		os.Exit(0)
	}

	stdr.SetVerbosity(verbosity)
	log := stdr.New(stdlog.New(os.Stderr, "I ", stdlog.Ltime))
	err := run(log, directory, flag.Args())
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "E", err)
		os.Exit(1)
	}
}

// Run the args command and record output in dir.
func run(log logr.Logger, dir string, args []string) error {
	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		return err
	}

	opt := &exe.Opt{Env: os.Environ()}
	rr := &exe.RunRecord{
		Dir: dir,
		Log: log,
	}
	stdout, stderr, err := rr.Run(opt, "", args[0], args[1:]...)

	fmt.Fprint(os.Stdout, stdout)
	fmt.Fprint(os.Stderr, stderr)

	return err
}
