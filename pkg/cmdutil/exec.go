package cmdutil

import (
	"fmt"
	"os"
	"os/exec"
)

// Run a command and return an error saying what happened, perhaps stderr
func Run(cmd *exec.Cmd) error {
	_, err := cmd.Output()
	return ExecError(err)
}

// Output runs a command and return an error saying what happened, perhaps stderr
func Output(cmd *exec.Cmd) ([]byte, error) {
	out, err := cmd.Output()
	return out, ExecError(err)
}

// ExitError wraps exec.ExitError with more useful message
type ExitError struct {
	*exec.ExitError
}

func (e ExitError) Error() string {
	return e.ExitError.Error() + ": " + string(e.Stderr)
}

// Unwrap is for Go 1.13 errors package
func (e ExitError) Unwrap() error {
	return e.ExitError
}

// ExecError will wrap err, if err is of type exec.ExitError
func ExecError(err error) error {
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			return &ExitError{ee}
		}
	}
	return err
}

// ErrorExit will print a message to stderr and exit
func ErrorExit(msg string, err interface{}) {
	fmt.Fprintf(os.Stderr, "%s: %v\n", msg, err)
	os.Exit(1)
}
