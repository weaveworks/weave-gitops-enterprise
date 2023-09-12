package version

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/weaveworks/weave-gitops/cmd/gitops/version"
)

// TestVersionCommand serves as regression point for gitops version command
// which comes directly from weave gitops oss. It ensures that the command
// is available and generating the expeted output.
func TestVersionCommand(t *testing.T) {
	cmd := version.Cmd
	cmd.SetArgs([]string{})

	version.Version = "v1.0.0"
	version.GitCommit = "abcd"
	version.Branch = "main"
	version.BuildTime = "2023-09-12_16:25:01"

	// capture stdout to be able to assert version content
	originalStdout := os.Stdout
	reader, writer, _ := os.Pipe()
	os.Stdout = writer

	err := cmd.Execute()
	assert.NoError(t, err)

	//read command output from buffer
	writer.Close()
	var capturedOutput string
	var buf bytes.Buffer
	_, err = io.Copy(&buf, reader)
	assert.NoError(t, err)
	capturedOutput = buf.String()
	fmt.Println(capturedOutput)

	os.Stdout = originalStdout

	//assert version has been properly generated
	assert.Contains(t, capturedOutput, fmt.Sprintf("Current Version: %s", version.Version))
	assert.Contains(t, capturedOutput, fmt.Sprintf("GitCommit: %s", version.GitCommit))
	assert.Contains(t, capturedOutput, fmt.Sprintf("BuildTime: %s", version.BuildTime))
	assert.Contains(t, capturedOutput, fmt.Sprintf("Branch: %s", version.Branch))

}
