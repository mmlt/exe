package exe

import (
	"bytes"
	"fmt"
	"github.com/go-logr/logr"
	"io"
	"os"
	"os/exec"
	"path/filepath"
)

type Runner interface {
	// Run executes 'cmd' with 'stdin', 'args' and (optional) 'options'.
	// Return stdout and stderr upon completion.
	Run(options *Opt, stdin string, cmd string, args ...string) (stdout, stderr string, err error)
}

// Run implements Runner for execution of processes.
type Run struct {
	Log logr.Logger
}

func (r *Run) Run(options *Opt, stdin string, cmd string, args ...string) (stdout, stderr string, err error) {
	c := exec.Command(cmd, args...)

	if options != nil {
		c.Env = options.Env
		c.Dir = options.Dir
	}

	if stdin != "" {
		sin, err := c.StdinPipe()
		if err != nil {
			r.Log.Error(err, "run stdin")
			return "", "", err
		}

		go func() {
			_, err := io.WriteString(sin, stdin)
			if err != nil {
				r.Log.Error(err, "run stdin")
			}

			sin.Close()
		}()
	}

	var sout, serr bytes.Buffer
	c.Stdout, c.Stderr = &sout, &serr
	err = c.Run()
	stdout, stderr = string(sout.Bytes()), string(serr.Bytes())
	r.Log.Info("Run-result", "stderr", stderr, "stdout", stdout)
	if err != nil {
		return stdout, stderr, fmt.Errorf("%s %v: %w - %s", cmd, args, err, stderr)
	}

	return
}

// RunRecord implements Runner for execution of processes and recording stdout/stderr.
type RunRecord struct {
	// Dir to store recordings.
	Dir string
	//
	Log logr.Logger
}

func (rr *RunRecord) Run(options *Opt, stdin string, cmd string, args ...string) (stdout, stderr string, err error) {
	r := &Run{
		Log: rr.Log,
	}
	stdout, stderr, err = r.Run(options, stdin, cmd, args...)

	rr.mustRecord(stdout, stderr, err, options, stdin, cmd, args...)

	return
}

func (rr *RunRecord) mustRecord(stdout string, stderr string, inerr error, options *Opt, stdin string, cmd string, args ...string) {
	p := filepath.Join(rr.Dir, filename(stdin, cmd, args...))

	f, err := os.Create(p)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	//mustWrite(f, stdout, stderr, err, options, stdin, cmd, args...)
	mustWrite(f, recording{
		cmd:    cmd,
		args:   args,
		stdin:  stdin,
		stdout: stdout,
		stderr: stderr,
		err:    inerr,
	})
}

// RunPlayback implements Runner for playback of process stdout/stderr output without running them.
type RunPlayback struct {
	// Dir to store recordings.
	Dir string
	//
	Log logr.Logger
}

func (rp *RunPlayback) Run(options *Opt, stdin string, cmd string, args ...string) (stdout, stderr string, errout error) {
	p := filepath.Join(rp.Dir, filename(stdin, cmd, args...))

	f, err := os.Open(p)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	rec := mustRead(f)
	stdout = rec.stdout
	stderr = rec.stderr
	errout = rec.err

	return
}
