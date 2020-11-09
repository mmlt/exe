package exe

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"hash/fnv"
	"io"
	"strings"
	"unicode"
)

// Recording is the data that is stored during record and used during playback.
type recording struct {
	cmd    string
	args   []string
	stdin  string
	stdout string
	stderr string
	err    error
	//TODO timing
}

// Recordings are stored in files that are easy to read, diff and edit.
//
// File format:
// 	First line of file contains JSON with meta data.
//  Other fields contain the starting line number of the section.
//  The file ends with a JSON line that contains timing data for replaying async commands.
//
// Filename format is <cmd>-<args>-<hash> were args is limited in length and only contains alphanumeric chars.

// Sep is the section separator.
// It should contain at least one LF, the rest is to improve readability.
const sep = "\n----\n"
const sepLen = 2

// RecodingHeader stores the line number offsets in the file.
type RecordingHeader struct {
	// Version of file format.
	Version int `json:"version"`
	// Cmd is the line number were the cmd is recorded.
	Cmd    int `json:"cmd"`
	Stdin  int `json:"stdin,omitempty"`
	Stdout int `json:"stdout,omitempty"`
	Stderr int `json:"stderr,omitempty"`
	Err    int `json:"err,omitempty"`
	Timing int `json:"timing,omitempty"`
}

// MustWrite writes recording rec to out.
func mustWrite(out io.Writer, rec recording) {
	var errStr string
	if rec.err != nil {
		errStr = rec.err.Error()
	}

	// write header
	h := RecordingHeader{
		Version: 1,
	}
	h.Cmd = 1 + sepLen
	h.Stdin = h.Cmd + sepLen
	h.Stdout = h.Stdin + strings.Count(rec.stdin, "\n") + sepLen
	h.Stderr = h.Stdout + strings.Count(rec.stdout, "\n") + sepLen
	h.Err = h.Stderr + strings.Count(rec.stderr, "\n") + sepLen
	h.Timing = h.Err + strings.Count(errStr, "\n") + sepLen

	b, err := json.Marshal(&h)
	if err != nil {
		panic(err)
	}
	mustWriteBytes(out, b)
	mustWriteString(out, sep)

	// write cmd
	mustWriteString(out, strings.Join(append([]string{rec.cmd}, rec.args...), " "))
	mustWriteString(out, sep)

	// write stdin, out, err
	mustWriteString(out, rec.stdin)
	mustWriteString(out, sep)
	mustWriteString(out, rec.stdout)
	mustWriteString(out, sep)
	mustWriteString(out, rec.stderr)
	mustWriteString(out, sep)
	mustWriteString(out, errStr)
	mustWriteString(out, sep)

	mustWriteString(out, "timing JSON TODO")
}

func mustRead(in io.Reader) recording {
	// read in
	var lines [][]byte
	scanner := bufio.NewScanner(in)
	for scanner.Scan() {
		lines = append(lines, scanner.Bytes())
	}
	if err := scanner.Err(); err != nil {
		panic(err)
	}

	// header
	var h RecordingHeader
	err := json.Unmarshal(lines[0], &h)
	if err != nil {
		panic(err)
	}

	// sections
	c := strings.Split(string(lines[h.Cmd-1]), " ")
	rec := recording{
		cmd:    c[0],
		args:   c[1:],
		stdin:  slice(lines, h.Stdin-1, h.Stdout-sepLen),
		stdout: slice(lines, h.Stdout-1, h.Stderr-sepLen),
		stderr: slice(lines, h.Stderr-1, h.Err-sepLen),
	}
	errStr := slice(lines, h.Err-1, h.Timing-sepLen)
	if errStr != "" {
		rec.err = errors.New(errStr)
	}

	return rec
}

func slice(lines [][]byte, begin int, end int) string {
	var b bytes.Buffer
	for i := begin; i < end; i++ {
		b.WriteString(string(lines[i]))
		if i < end-1 || len(lines[i+1]) == 0 {
			b.WriteString("\n")
		}
	}
	return b.String()
}

// Filename return an unique name that can be used as a filename.
// Format <cmd>-<args>-<hash>
func filename(stdin string, cmd string, args ...string) string {
	// hash
	h := fnv.New32a()
	h.Write([]byte(stdin))
	h.Write([]byte(cmd))
	for _, v := range args {
		h.Write([]byte(v))
	}

	// args without non alphanumeric chars and limited to maxArgsLen.
	const maxArgsLen = 20
	var a strings.Builder
	var i int
	for _, v := range strings.Join(args, "") {
		if unicode.IsLetter(v) || unicode.IsDigit(v) {
			a.WriteRune(v)
			i++
			if i > maxArgsLen {
				break
			}
		}
	}

	return fmt.Sprintf("%s-%s-%x", cmd, a.String(), h.Sum32())
}

func mustWriteString(out io.Writer, s string) {
	mustWriteBytes(out, []byte(s))
}

func mustWriteBytes(out io.Writer, b []byte) {
	_, err := out.Write(b)
	if err != nil {
		panic(err)
	}
}
