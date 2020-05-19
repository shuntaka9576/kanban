package git

import (
	"net/url"
	"reflect"
	"testing"
)

func TestParseGitRemote(t *testing.T) {
	tests := []struct {
		gitRemoteResult string
		expected        map[string]string
	}{
		{
			gitRemoteResult: "origin  https://github.com/shuntaka9576/kanban.git (fetch)\norigin  https://github.com/shuntaka9576/kanban.git (push)\n",
			expected: map[string]string{
				"Name": "origin",
				"Url":  "https://github.com/shuntaka9576/kanban.git",
			},
		},
		{
			gitRemoteResult: "origin  git@github.com:shuntaka9576/kanban.git (fetch)\norigin  git@github.com:shuntaka9576/kanban.git (push)\n",
			expected: map[string]string{
				"Name": "origin",
				"Url":  "ssh://git@github.com/shuntaka9576/kanban.git",
			},
		},
	}

	for _, test := range tests {
		remotes, err := ParseGitRemote(test.gitRemoteResult)
		if err != nil {
			t.Errorf("got err %v", err)
		}

		got := remotes[0]
		url, _ := url.Parse(test.expected["Url"])
		expected := &Remote{
			Name:     test.expected["Name"],
			PushUrl:  url,
			FetchUrl: url,
		}

		if !reflect.DeepEqual(got, expected) {
			t.Errorf("got: %v, expected: %v", got, expected)
		}
	}
}

// TODO
// func TestBaseRepoFromRemote(t *testing.T) {
// 	repo, err := BaseRepoFromRemote()
// 	if err != nil {
// 		t.Errorf("got err %v", err)
// 	}
// }
