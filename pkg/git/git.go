package git

import (
	"errors"
	"net/url"
	"os/exec"
	"regexp"
	"sort"
	"strings"

	"github.com/shuntaka9576/kanban/pkg/runCmd"
	"github.com/shuntaka9576/kanban/pkg/stringUtil"
)

type Repository interface {
	RepoName() string
	RepoOwner() string
}

type baseRepository struct {
	owner, name string
}

func (b *baseRepository) RepoOwner() string {
	return b.owner
}

func (b *baseRepository) RepoName() string {
	return b.name
}

func Repo(owner, name string) (*baseRepository, error) {
	return &baseRepository{
		owner: owner,
		name:  name,
	}, nil
}

type Remote struct {
	Name     string
	FetchUrl *url.URL
	PushUrl  *url.URL
}

type RemoteRepository struct {
	*Remote
	owner string
	name  string
}

func (r *RemoteRepository) RepoOwner() string {
	return r.owner
}

func (r *RemoteRepository) RepoName() string {
	return r.name
}

func BaseRepoFromRemote() (*RemoteRepository, error) {
	gitRemoteCmd := exec.Command("git", "remote", "-v")
	gitRemoteWithStderrCmd := &runCmd.CmdOutputWithStderr{Cmd: gitRemoteCmd}
	gitRemoteResult, err := gitRemoteWithStderrCmd.Output()
	if err != nil {
		return nil, err
	}

	remotes, err := ParseGitRemote(string(gitRemoteResult))
	if err != nil {
		return nil, err
	}

	remoteRepositories, err := RemoteRepositoriesFromRemote(remotes)
	if err != nil {
		return nil, err
	}
	sort.Sort(remoteRepositories)

	return remoteRepositories[0], nil
}

func RemoteRepositoriesFromRemote(remotes []*Remote) (remoteRepositories RemoteRepositories, err error) {
	for _, remote := range remotes {
		var paths []string
		if remote.FetchUrl != nil {
			paths = strings.Split(strings.TrimPrefix(remote.FetchUrl.Path, "/"), "/")
		}
		if remote.PushUrl != nil {
			paths = strings.Split(strings.TrimPrefix(remote.PushUrl.Path, "/"), "/")
		}
		if len(paths) > 2 {
			return nil, errors.New("invalid remote url")
		}
		remoteRepo := &RemoteRepository{Remote: remote, owner: paths[0], name: strings.TrimSuffix(paths[1], ".git")}

		remoteRepositories = append(remoteRepositories, remoteRepo)
	}
	return
}

type RemoteRepositories []*RemoteRepository

func remoteNameSortScore(name string) int {
	switch strings.ToLower(name) {
	case "upstream":
		return 3
	case "github":
		return 2
	case "origin":
		return 1
	default:
		return 0
	}
}

func (r RemoteRepositories) Len() int      { return len(r) }
func (r RemoteRepositories) Swap(i, j int) { r[i], r[j] = r[j], r[i] }
func (r RemoteRepositories) Less(i, j int) bool {
	return remoteNameSortScore(r[i].Remote.Name) > remoteNameSortScore(r[j].Remote.Name)
}

func ParseGitRemote(gitRemoteResult string) (remotes []*Remote, err error) {
	gitRemoteRegex := regexp.MustCompile(`(.+)\s+(.+)\s+\((fetch|push)\)`)
	protocolRegex := regexp.MustCompile("^[a-zA-Z_+-]+://")
	gitRemoteResultLines := strings.Split(stringUtil.TrimSuffixLine(gitRemoteResult), "\n")

	for _, gitRemoteLine := range gitRemoteResultLines {
		match := gitRemoteRegex.FindStringSubmatch(gitRemoteLine)
		if match == nil {
			continue
		}
		name := strings.TrimSpace(match[1])
		gitUrl := strings.TrimSpace(match[2])
		urlType := strings.TrimSpace(match[3])

		var remote *Remote
		if len(remotes) > 0 {
			if name == remotes[len(remotes)-1].Name {
				remote = remotes[len(remotes)-1]
			}
		}
		if remote == nil {
			remote = &Remote{
				Name: name,
			}
			remotes = append(remotes, remote)
		}

		if !protocolRegex.MatchString(gitUrl) {
			gitUrl = "ssh://" + strings.Replace(gitUrl, ":", "/", 1)
		}

		parsedGitUrl, err := url.Parse(gitUrl)
		if err != nil {
			return nil, err
		}

		switch urlType {
		case "push":
			remote.PushUrl = parsedGitUrl
		case "fetch":
			remote.FetchUrl = parsedGitUrl
		}
	}
	return
}
