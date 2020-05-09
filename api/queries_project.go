package api

import (
	"context"

	"github.com/shuntaka9576/kanban/pkg/git"
)

type GithubProject struct {
	ProjectUrl string
	Columns    []Column
}

type Column struct {
	Name  string
	Cards []Card
}

type Card struct {
	Title      string
	Number     int
	Url        string
	Body       string
	Labels     []Label
	Assignees  []Assignee
	Note       string
	IsArchived bool
}

type Label struct {
	Name string
}

type Assignee struct {
	Login     string
	AvatorUrl string
	Name      string
	Id        string
	Url       string
}

const issueFragments = `
	fragment issue on Issue {
		title
		number
		url
		body
		labels(first: 10) {
			edges {
				cursor
				node {
					color
					name
				}
			}
		}
		assignees(first: 10) {
			edges {
				cursor
				node {
					name
					url
					id
					avatarUrl(size: 10)
					login
				}
			}
		}
	}
`

const projectCardConnectionFragments = `
	fragment projectCardConnection on ProjectCardConnection {
			edges {
			cursor
			node {
				content {
					...issue
				}
				note
				isArchived
			}
		}
	}
`

const projectConnectionFragments = `
	fragment projectConnection on ProjectConnection {
			edges {
			cursor
			node {
				columns(first: 10) {
					edges {
						cursor
						node {
							name
							cards(first: 100) {
								...projectCardConnection
							}
						}
					}
				}
				name
				url
			}
		}
	}
`

func Project(client *Client, repo git.Repository, searchString string) (*GithubProject, error) {
	type LabelNode struct {
		Color string `json:"color"`
		Name  string `json:"name"`
	}
	type AssigneeNode struct {
		Login     string `json:"login"`
		AvatorUrl string `json:"avatorUrl"`
		Name      string `json:"name"`
		Id        string `json:"id"`
		Url       string `json:"url"`
	}
	type AssigneesEdge struct {
		Cursor string       `json:"cursor"`
		Node   AssigneeNode `json:"node"`
	}
	type LabelEdge struct {
		Cursor string    `json:"cursor"`
		Node   LabelNode `json:"node"`
	}
	type Issue struct {
		Title  string `json:"title"`
		Number int    `json:"number"`
		Url    string `json:"url"`
		Body   string `json:"body"`
		Labels struct {
			Edges []LabelEdge `json:"edges"`
		} `json:"labels"`
		Assignees struct {
			Edges []AssigneesEdge `json:"edges"`
		} `json:"assignees"`
	}
	type CardNode struct {
		Content    Issue  `json:"content"`
		Note       string `json:"note"`
		IsArchived bool   `json:"isArchived"`
	}
	type CardEdge struct {
		Cursor string   `json:"cursor"`
		Node   CardNode `json:"node"`
	}
	type ColumnNode struct {
		Name  string `json:"name"`
		Cards struct {
			Edges []CardEdge `json:"edges"`
		} `json:"cards"`
	}
	type ColumnEdge struct {
		Cursor string     `json:"cursor"`
		Node   ColumnNode `json:"node"`
	}
	type ProjectNode struct {
		Columns struct {
			Edges []ColumnEdge `json:"edges"`
		} `json:"columns"`
		Name string `json:"name"`
		Url  string `json:"url"`
	}
	type ProjectEdge struct {
		Cursor string      `json:"cursor"`
		Node   ProjectNode `json:"node"`
	}
	type Repository struct {
		Name     string `json:"name"`
		Projects struct {
			Edges []ProjectEdge `json:"edges"`
		} `json:"projects"`
	}
	type response struct {
		Repository Repository `json:"repository"`
	}

	const query = issueFragments + projectCardConnectionFragments + projectConnectionFragments + `
	query ($owner: String!, $name: String!, $searchString: String!) {
		repository(owner: $owner, name: $name) {
			name
			projects(first: 1, search: $searchString) {
				...projectConnection
			}
		}
	}
	`

	variables := map[string]interface{}{
		"owner":        repo.RepoOwner(),
		"name":         repo.RepoName(),
		"searchString": searchString,
	}

	var resp response
	err := client.GraphQL(query, variables, &resp)
	if err != nil {
		return nil, err
	}

	var githubProject GithubProject

	for _, projects := range resp.Repository.Projects.Edges {
		githubProject.ProjectUrl = projects.Node.Url

		for _, columsEdges := range projects.Node.Columns.Edges {
			column := Column{
				Name: columsEdges.Node.Name,
			}
			for _, cardEdge := range columsEdges.Node.Cards.Edges {
				card := Card{
					Title:      cardEdge.Node.Content.Title,
					Number:     cardEdge.Node.Content.Number,
					Url:        cardEdge.Node.Content.Url,
					Body:       cardEdge.Node.Content.Body,
					Note:       cardEdge.Node.Note,
					IsArchived: cardEdge.Node.IsArchived,
				}
				for _, assigneesEdge := range cardEdge.Node.Content.Assignees.Edges {
					assignee := Assignee{
						Login:     assigneesEdge.Node.Login,
						AvatorUrl: assigneesEdge.Node.AvatorUrl,
						Name:      assigneesEdge.Node.Name,
						Id:        assigneesEdge.Node.Id,
						Url:       assigneesEdge.Node.Url,
					}
					card.Assignees = append(card.Assignees, assignee)
				}
				for _, labelEdge := range cardEdge.Node.Content.Labels.Edges {
					label := Label{
						Name: labelEdge.Node.Name,
					}
					card.Labels = append(card.Labels, label)
				}
				column.Cards = append(column.Cards, card)
			}
			githubProject.Columns = append(githubProject.Columns, column)
		}
	}

	return &githubProject, nil
}

func ProjectWithContext(ctx context.Context, client *Client, repo git.Repository, searchString string, ghpjChan chan *GithubProject) {
	gp, _ := Project(client, repo, searchString) // TODO error handling
	ghpjChan <- gp
}
