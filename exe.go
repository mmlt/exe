package exe

// Opt are the exec options, see https://godoc.org/os/exec#Cmd for details.
type Opt struct {
	// Dir is the working directory.
	Dir string
	// Env is the execution environment.
	Env []string
}
