package exe

import (
	"bytes"
	"errors"
	"github.com/stretchr/testify/assert"
	"testing"
)

// TestRecording takes a recording, serializes and deserializes it and asserts that the resulting recording matches
// the original.
func TestRecording(t *testing.T) {
	var tests = []struct {
		it  string
		rec recording
	}{
		{
			it: "should_handle_text_ending_with_LF",
			rec: recording{
				cmd:    "command",
				args:   []string{"1", "2", "3"},
				stdin:  "single line with LF\n",
				stdout: "double line\nwith LF\n",
				stderr: "\n",
			},
		},
		{
			it: "should_handle_text_with_no_LF_at_end",
			rec: recording{
				cmd:    "command",
				args:   []string{"1", "2", "3"},
				stdin:  "single line without LF",
				stdout: "double line\nwithout LF",
				stderr: "",
			},
		},
		{
			it: "should_handle_err",
			rec: recording{
				cmd:    "command",
				args:   []string{},
				stderr: "clstrfck",
				err:    errors.New("exit 1: clstrfck"),
			},
		},
	}

	for _, tst := range tests {
		t.Run(tst.it, func(t *testing.T) {
			var b bytes.Buffer
			mustWrite(&b, tst.rec)
			got := mustRead(&b)
			assert.Equal(t, tst.rec, got)
		})
	}
}
