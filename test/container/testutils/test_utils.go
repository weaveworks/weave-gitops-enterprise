package testutils

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/weaveworks/wks/pkg/apis/wksprovider/machine/scripts"
	"github.com/weaveworks/wks/pkg/plan"
)

type Operation struct {
	Kind   string
	Arg    string
	Output string // for operations that output something on stdouterr, we keep it there.
}

type TestRunner struct {
	T      *testing.T
	Runner plan.Runner

	ops []Operation
}

var _ plan.Runner = &TestRunner{}

// RunCommand implements plan.Runner.
func (r *TestRunner) RunCommand(cmd string) (stdouterr string, err error) {
	r.T.Log("RunCommand:", cmd)
	stdouterr, err = r.Runner.RunCommand(cmd)
	r.T.Logf("Output:\n%s", stdouterr)

	r.pushRunCommand(cmd, stdouterr)
	return
}

// RunScript implements Runner.
func (r *TestRunner) RunScript(path string, args interface{}) (stdouterr string, err error) {
	r.T.Log("RunScript:", path)
	stdouterr, err = scripts.Run(path, args, r)
	r.T.Logf("Output:\n%s", stdouterr)
	r.pushRunCommand(path, stdouterr)
	return
}

// WriteFile implements plan.Runner.
func (r *TestRunner) WriteFile(content []byte, path string, perm os.FileMode) error {
	r.pushWriteFile(path)

	r.T.Log("WriteFile:", path)
	return r.Runner.WriteFile(content, path, perm)
}

// Give tests visibility on the operations done by a applying a resource.

func (r *TestRunner) Operations() []Operation {
	return r.ops
}

func (r *TestRunner) ResetOperations() {
	r.ops = nil
}

func (r *TestRunner) pushRunCommand(cmd string, output string) {
	r.ops = append(r.ops, Operation{
		Kind:   "RunCommand",
		Arg:    cmd,
		Output: output,
	})
}

func (r *TestRunner) pushWriteFile(path string) {
	r.ops = append(r.ops, Operation{
		Kind: "WriteFile",
		Arg:  path,
	})
}

func (r *TestRunner) Operation(i int) Operation {
	if i >= 0 {
		return r.ops[i]
	}
	return r.ops[len(r.ops)+i]
}

// Other utilities
func AssertEmptyState(t *testing.T, s plan.Resource, r plan.Runner) {
	state, err := s.QueryState(r)
	assert.NoError(t, err)
	assert.Equal(t, plan.EmptyState, state)
}
