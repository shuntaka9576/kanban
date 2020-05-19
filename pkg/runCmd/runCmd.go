package runCmd

import (
	"bytes"
	"os/exec"
)

type CmdOutputWithStderr struct {
	*exec.Cmd
}

type CmdError struct {
	Stderr *bytes.Buffer
	Args   []string
	Err    error
}

func (c *CmdError) Error() string {
	return c.Err.Error()
}

func (c *CmdOutputWithStderr) Output() ([]byte, error) {
	stderrBuf := &bytes.Buffer{}
	c.Cmd.Stderr = stderrBuf

	output, err := c.Cmd.Output()
	if err != nil {
		return nil, &CmdError{
			Stderr: stderrBuf,
			Args:   c.Args,
			Err:    err,
		}
	}

	return output, nil
}
