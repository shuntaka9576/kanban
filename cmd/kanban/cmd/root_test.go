package cmd

import (
	"testing"
)

func TestDetermineBaseRepository_flag(t *testing.T) {
	rootCmd := RootCmd
	// TODO replace OWNER/REPO mock
	rootCmd.SetArgs([]string{"--repo", "shuntaka9576/memo"})

	if _, err := RootCmd.ExecuteC(); err != nil {
		t.Errorf("command exec error: %v", err)
	}
}
