package ui

import (
	"context"
	"fmt"

	"github.com/gdamore/tcell"
	"github.com/google/go-github/github"
	"github.com/rivo/tview"
	"github.com/shuntaka9576/kanban/api"
	"github.com/shuntaka9576/kanban/pkg/git"
)

const FOCUS_COLUMN_COLOR = tcell.Color40

type Tui struct {
	App          *tview.Application
	notice       *Notciation
	view         *View
	ghpjSettings *GhpjSettings
	pos          *Pos
}

type GhpjSettings struct {
	Client       *api.Client
	Repository   git.Repository
	SearchString string
	ProjectUrl   string
	V3Client     *github.Client
}

type Notciation struct {
	ghpjChan chan *api.GithubProject
}

type View struct {
	columns *Columns
	content *Content
	help    *Help
}

type Pos struct {
	focusCol int
}

func NewTui(g *GhpjSettings) *Tui {
	app := tview.NewApplication()
	app.SetBeforeDrawFunc(func(s tcell.Screen) bool {
		s.Clear()
		return false
	})
	ghpjChan := make(chan *api.GithubProject)

	return &Tui{
		App:          app,
		ghpjSettings: g,
		notice: &Notciation{
			ghpjChan: ghpjChan,
		},
		pos:  &Pos{},
		view: &View{},
	}
}

func (tui *Tui) Run(ctx context.Context) {
	tui.initialize()
	go api.ProjectWithContext(ctx, tui.ghpjSettings.Client, tui.ghpjSettings.Repository, tui.ghpjSettings.SearchString, tui.notice.ghpjChan)

	go func() {
		for {
			select {
			case <-ctx.Done():
				fmt.Printf("err!\n")
			case ghpj := <-tui.notice.ghpjChan:
				tui.display(ghpj)
			}
		}
	}()

	if err := tui.App.Run(); err != nil {
		panic(err)
	}

}

func (tui *Tui) initialize() {
	tui.view.content = newContent()
	tui.view.help = newHelp()
	tui.view.columns = newColumns()

	contentAndHelpGrid := tview.NewGrid()
	contentAndHelpGrid.SetColumns(0, 0, 0)
	contentAndHelpGrid.AddItem(tui.view.content, 0, 0, 1, 2, 0, 0, false)
	contentAndHelpGrid.AddItem(tui.view.help, 0, 2, 1, 1, 0, 0, false)

	allGrid := tview.NewGrid().SetRows(0, 0)
	allGrid.AddItem(tui.view.columns, 0, 0, 1, 3, 0, 0, false)
	allGrid.AddItem(contentAndHelpGrid, 1, 0, 1, 3, 0, 0, false)

	loadingView := tview.NewTextView()
	loadingView.SetBorder(true)
	loadingView.SetText("loading...")
	loadingView.SetBackgroundColor(tcell.ColorDefault)
	tui.view.columns.AddItem(loadingView, 0, 0, 1, 1, 0, 0, false)

	tui.App.SetRoot(allGrid, true)
}

func (tui *Tui) display(ghpj *api.GithubProject) {
	posBackup := false
	var beforePos *Pos
	var beforeRow, beforeCol int
	if len(tui.view.columns.columns) > 0 {
		beforePos = tui.pos
		beforeRow, beforeCol = tui.view.columns.columns[tui.pos.focusCol].GetSelection()
		posBackup = true
	}

	if ghpj != nil {
		tui.ghpjSettings.ProjectUrl = ghpj.ProjectUrl
		tui.view.columns.Clear()
		tui.view.columns.columns = []*Column{}

		colSize := make([]int, len(ghpj.Columns)-1)
		tui.view.columns.SetColumns(colSize...)

		for i, column := range ghpj.Columns {
			col := newColumn(column, tui)
			tui.view.columns.AddItem(col, 0, i, 1, 1, 0, 0, false)
			tui.view.columns.columns = append(tui.view.columns.columns, col)
		}

		if posBackup {
			tui.view.columns.columns[beforePos.focusCol].forcus(tui)
			if cell := tui.view.columns.columns[beforePos.focusCol].GetCell(beforeRow, beforeCol); cell.Text != "" {
				tui.view.columns.columns[beforePos.focusCol].Select(beforeRow, beforeCol)
				tui.view.columns.columns[tui.pos.focusCol].setContent(tui, beforeRow, beforeCol)
			}
		}

		if len(tui.view.columns.columns) > 0 {
			tui.view.columns.columns[tui.pos.focusCol].forcus(tui)
			tui.view.columns.columns[tui.pos.focusCol].setContent(tui, 0, 0)
		}
	} else {
		errText := tview.NewTextView()
		errText.SetBackgroundColor(tcell.ColorDefault)
		errText.SetBorder(true)
		errText.SetText("API Request Error")
		tui.view.columns.Clear()
		tui.view.columns.AddItem(errText, 0, 0, 1, 1, 0, 0, false)
	}

	tui.view.help.Update()
	tui.App.Draw()
}
