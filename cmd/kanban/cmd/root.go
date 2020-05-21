package cmd

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/shuntaka9576/kanban/api"
	"github.com/shuntaka9576/kanban/internal/kanban/config"
	"github.com/shuntaka9576/kanban/internal/kanban/ui"
	"github.com/shuntaka9576/kanban/pkg/git"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var RootCmd = &cobra.Command{
	Use:           "kanban",
	Short:         "GitHub Project TUI",
	Long:          "GitHub Project Viewer",
	SilenceErrors: true,
	SilenceUsage:  true,
	RunE:          runRootCmd,
}

type FlagError struct {
	Err error
}

func (fe FlagError) Error() string {
	return fe.Err.Error()
}

func (fe FlagError) Unwrap() error {
	return fe.Err
}

var Version = "DEV"
var BuildDate = ""
var versionOutput = ""

var versionCmd = &cobra.Command{
	Use:    "version",
	Hidden: true,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Print(versionOutput)
	},
}

func init() {
	RootCmd.AddCommand(versionCmd)
	RootCmd.PersistentFlags().StringP("repo", "R", "", "Select another repository using the `OWNER/REPO` format")
	RootCmd.PersistentFlags().Bool("help", false, "Show help for command")
	RootCmd.Flags().StringP("search", "S", "", "Search project name string(default first project)")

	if BuildDate == "" {
		RootCmd.Version = Version
	} else {
		RootCmd.Version = fmt.Sprintf("%s (%s)", Version, BuildDate)
	}
	versionOutput = fmt.Sprintf("kanban version %s\n", RootCmd.Version)

	RootCmd.SetFlagErrorFunc(func(cmd *cobra.Command, err error) error {
		if err == pflag.ErrHelp {
			return err
		}
		return &FlagError{Err: err}
	})
}

func runRootCmd(cmd *cobra.Command, args []string) error {
	config := config.New()
	token, err := config.AuthToken()
	if err != nil {
		return err
	}

	baseRepository, err := DetermineBaseRepository(cmd)
	if err != nil {
		return err
	}

	searchString, err := cmd.Flags().GetString("search")
	if err != nil {
		return err
	}

	if err != nil {
		return err
	}

	tui := ui.NewTui(&ui.GhpjSettings{
		Client:       api.NewGitHubClient(token),
		Repository:   baseRepository,
		SearchString: searchString,
	})

	ctx := context.Background()
	tui.Run(ctx)

	return nil
}

func DetermineBaseRepository(cmd *cobra.Command) (git.Repository, error) {
	ownerRepoString, err := cmd.Flags().GetString("repo")

	var repo git.Repository
	if err != nil || ownerRepoString == "" {
		repo, err = git.BaseRepoFromRemote()
		if err != nil {
			return nil, err
		}
	} else {
		ownerRepoList := strings.Split(ownerRepoString, "/")
		if len(ownerRepoList) != 2 {
			return nil, FlagError{errors.New("invalid --repo value: " + ownerRepoString + "\nPlease set OWNER/REPO format")}
		}
		repo, err = git.Repo(ownerRepoList[0], ownerRepoList[1])
	}

	return repo, err
}
