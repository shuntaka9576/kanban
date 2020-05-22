package ui

import (
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"

	b64 "encoding/base64"

	"github.com/gdamore/tcell"
	"github.com/google/go-github/github"
	"github.com/rivo/tview"
	"github.com/shuntaka9576/kanban/api"
	"github.com/shuntaka9576/kanban/pkg/markdown"
	"gopkg.in/kyokomi/emoji.v1"
)

type Columns struct {
	*tview.Grid
	columns []*Column
}

type Card struct {
	Id   int64
	Card api.Card
}

type Column struct {
	*tview.Table
	cards []*Card
	Id    int64
}

func newColumns() *Columns {
	columnsGrid := tview.NewGrid()

	return &Columns{
		Grid:    columnsGrid,
		columns: []*Column{},
	}
}

func newColumn(apiColumn api.Column, tui *Tui) *Column {
	columnIdRE := regexp.MustCompile(`.+:ProjectColumn(\d+)$`)
	sDec, _ := b64.StdEncoding.DecodeString(apiColumn.Id)
	match := columnIdRE.FindStringSubmatch(string(sDec))
	colId, _ := strconv.ParseInt(match[1], 10, 64)

	columnTable := tview.NewTable()
	columnTable.SetBorder(true)
	columnTable.SetBackgroundColor(tcell.ColorDefault)
	columnTable.SetTitle(" " + apiColumn.Name + " ")
	columnTable.SetSelectedStyle(tcell.Color207, tcell.ColorDefault, tcell.AttrBold)
	columnTable.SetSelectable(false, false).Select(0, 0).SetFixed(0, 1)

	cards := []*Card{}
	for _, card := range apiColumn.Cards {
		cardIdRE := regexp.MustCompile(`.+:ProjectCard(\d+)$`)
		sDec, _ := b64.StdEncoding.DecodeString(card.Id)
		match := cardIdRE.FindStringSubmatch(string(sDec))
		cardId, _ := strconv.ParseInt(match[1], 10, 64)

		cards = append(cards, &Card{Id: cardId, Card: card})
	}

	column := &Column{
		Table: columnTable,
		cards: cards,
		Id:    colId,
	}
	column.setCards(apiColumn.Cards)
	column.setKeyBindings(tui)

	return column
}

func (c *Column) setKeyBindings(tui *Tui) {
	c.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyCtrlN:
			row, _ := tui.view.content.GetScrollOffset()
			tui.view.content.ScrollTo(row+2, 0)
		case tcell.KeyCtrlP:
			row, _ := tui.view.content.GetScrollOffset()
			tui.view.content.ScrollTo(row-2, 0)
		}

		switch event.Rune() {
		case 'h':
			tui.view.columns.columns[tui.pos.focusCol].unForcus(tui)

			if tui.pos.focusCol-1 < 0 {
				tui.pos.focusCol = len(tui.view.columns.columns) - 1
			} else {
				tui.pos.focusCol--
			}

			tui.view.columns.columns[tui.pos.focusCol].forcus(tui)
		case 'l':
			tui.view.columns.columns[tui.pos.focusCol].unForcus(tui)

			if tui.pos.focusCol+1 > len(tui.view.columns.columns)-1 {
				tui.pos.focusCol = 0
			} else {
				tui.pos.focusCol++
			}

			tui.view.columns.columns[tui.pos.focusCol].forcus(tui)
		case 'R':
			ctx := context.Background()
			go api.ProjectWithContext(ctx, tui.ghpjSettings.Client, tui.ghpjSettings.Repository, tui.ghpjSettings.SearchString, tui.notice.ghpjChan)
		case 'P':
			cmd := exec.Command("open", tui.ghpjSettings.ProjectUrl)
			cmd.Run() // TODO error handling
		case 'p':
			row, _ := tui.view.columns.columns[tui.pos.focusCol].GetSelection()
			url := tui.view.columns.columns[tui.pos.focusCol].cards[row].Card.Url
			cmd := exec.Command("open", url)
			cmd.Run() // TODO error handling
		case 'n':
			row, _ := tui.view.columns.columns[tui.pos.focusCol].GetSelection()
			cardId := tui.view.columns.columns[tui.pos.focusCol].cards[row].Id
			ctx := context.Background()
			nextCardPos := tui.pos.focusCol
			if tui.pos.focusCol+1 > len(tui.view.columns.columns)-1 {
				nextCardPos = 0
			} else {
				nextCardPos++
			}
			tui.ghpjSettings.V3Client.Projects.MoveProjectCard(ctx, cardId, &github.ProjectCardMoveOptions{Position: "top", ColumnID: tui.view.columns.columns[nextCardPos].Id})
			go api.ProjectWithContext(ctx, tui.ghpjSettings.Client, tui.ghpjSettings.Repository, tui.ghpjSettings.SearchString, tui.notice.ghpjChan)
		case 'b':
			row, _ := tui.view.columns.columns[tui.pos.focusCol].GetSelection()
			cardId := tui.view.columns.columns[tui.pos.focusCol].cards[row].Id
			nextCardPos := tui.pos.focusCol
			if tui.pos.focusCol-1 < 0 {
				nextCardPos = len(tui.view.columns.columns) - 1
			} else {
				nextCardPos--
			}
			ctx := context.Background()
			tui.ghpjSettings.V3Client.Projects.MoveProjectCard(ctx, cardId, &github.ProjectCardMoveOptions{Position: "top", ColumnID: tui.view.columns.columns[nextCardPos].Id})
			go api.ProjectWithContext(ctx, tui.ghpjSettings.Client, tui.ghpjSettings.Repository, tui.ghpjSettings.SearchString, tui.notice.ghpjChan)
		}

		return event
	})

	c.SetSelectionChangedFunc(func(row, col int) {
		if len(tui.view.columns.columns[tui.pos.focusCol].cards) > 1 {
			c.setContent(tui, row, col)
		}
	})
}

func (c *Column) setContent(tui *Tui, row, col int) {
	if len(tui.view.columns.columns[tui.pos.focusCol].cards) > 0 && row > -1 {
		title := tui.view.columns.columns[tui.pos.focusCol].cards[row].Card.Title
		url := tui.view.columns.columns[tui.pos.focusCol].cards[row].Card.Url
		body := tui.view.columns.columns[tui.pos.focusCol].cards[row].Card.Body

		tui.view.content.Clear()
		tui.view.content.SetText(tview.TranslateANSI(string(markdown.ConvertShellString(fmt.Sprintf("%s\n%s\n\n%s", title, url, body)))))
		tui.view.content.ScrollTo(0, 0)
	}
}

func (c *Column) forcus(tui *Tui) {
	c.SetSelectable(true, true)
	c.SetBorderColor(FOCUS_COLUMN_COLOR)
	c.SetTitleColor(FOCUS_COLUMN_COLOR)
	row, col := tui.view.columns.columns[tui.pos.focusCol].GetSelection()
	c.setContent(tui, row, col)

	tui.App.SetFocus(c)
}

func (c *Column) unForcus(tui *Tui) {
	c.SetSelectable(false, false)
	c.SetBorderColor(tcell.ColorWhite)
	c.SetTitleColor(tcell.ColorWhite)
}

func (c *Column) setCards(cards []api.Card) {
	cellNo := 0
	for r := 0; r < len(cards); r++ {

		if !cards[r].IsArchived {
			if cards[r].Title != "" {
				cell := tview.NewTableCell(cards[r].Title)
				c.SetCell(cellNo, 0, cell)
				cellNo++
			} else if cards[r].Note != "" {
				cell := tview.NewTableCell(emoji.Sprintf(":memo:") + cards[r].Note)
				c.SetCell(cellNo, 0, cell)
				cellNo++
			}
		}
	}
}
