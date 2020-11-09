package exe

import (
	"github.com/go-logr/stdr"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"testing"
)

var tests = []struct {
	it         string
	options    *Opt
	cmd        string
	args       []string
	stdin      string
	wantErr    string
	wantStdout string
	wantStderr string
}{
	{
		it:         "should_echo_on_stdout",
		cmd:        "echo",
		args:       []string{"-n", "hello world"},
		wantStdout: "hello world",
	},
	{
		it:         "should_error",
		cmd:        "ls",
		args:       []string{"nonexisting"},
		wantErr:    "ls [nonexisting]: exit status 2 - ls: cannot access 'nonexisting': No such file or directory\n",
		wantStderr: "ls: cannot access 'nonexisting': No such file or directory\n",
	},
	{
		it:         "should_read_stdin_and_write_stdout",
		cmd:        "base64",
		args:       []string{"-d"},
		stdin:      "aGVsbG8gd29ybGQ=",
		wantStdout: "hello world",
	},
	{
		it: "should_use_the_specified_environment",
		options: &Opt{
			Env: []string{"SONG=HappyHappyJoyJoy"},
		},
		cmd:        "env",
		wantStdout: "SONG=HappyHappyJoyJoy\n",
	},
	{
		it: "should_execute_in_the_specified_dir",
		options: &Opt{
			Dir: "/tmp",
		},
		cmd:        "pwd",
		wantStdout: "/tmp\n",
	},
}

// TestRun runs a process and asserts the outputs.
func TestRun(t *testing.T) {
	for _, tst := range tests {
		t.Run(tst.it, func(t *testing.T) {
			r := &Run{
				Log: stdr.New(nil),
			}
			stdout, stderr, err := r.Run(tst.options, tst.stdin, tst.cmd, tst.args...)
			if tst.wantErr != "" {
				assert.EqualError(t, err, tst.wantErr)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tst.wantStdout, stdout)
			assert.Equal(t, tst.wantStderr, stderr)
		})
	}
}

// TestRunRecordAndPlayback runs a process and records its outputs. Next a playback run uses the recordings to fake
// the process run.
// There should be no difference in the cmd output from a normal run and a playback run.
// NB. Recordings are not deleted automatically, to delete; rm -r /tmp/TestRunRecord*
func TestRunRecordAndPlayback(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "TestRunRecord")
	assert.NoError(t, err)

	for _, tst := range tests {
		t.Run(tst.it, func(t *testing.T) {
			// record
			rr := &RunRecord{
				Dir: tempDir,
				Log: stdr.New(nil),
			}
			stdout, stderr, err := rr.Run(tst.options, tst.stdin, tst.cmd, tst.args...)
			if tst.wantErr != "" {
				assert.EqualError(t, err, tst.wantErr, "record")
			} else {
				assert.NoError(t, err, "record")
			}
			assert.Equal(t, tst.wantStdout, stdout, "record")
			assert.Equal(t, tst.wantStderr, stderr, "record")

			// playback
			rp := &RunPlayback{
				Dir: tempDir,
				Log: stdr.New(nil),
			}
			stdout, stderr, err = rp.Run(tst.options, tst.stdin, tst.cmd, tst.args...)
			if tst.wantErr != "" {
				assert.EqualError(t, err, tst.wantErr, "playback")
			} else {
				assert.NoError(t, err, "playback")
			}
			assert.Equal(t, tst.wantStdout, stdout, "playback")
			assert.Equal(t, tst.wantStderr, stderr, "playback")
		})
	}
}
