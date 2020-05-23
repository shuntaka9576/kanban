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
	nodeId, _ := b64.StdEncoding.DecodeString(apiColumn.Id)
	match := columnIdRE.FindStringSubmatch(string(nodeId))
	colId, _ := strconv.ParseInt(match[1], 10, 64)

	columnTable := tview.NewTable()
	columnTable.SetBorder(true)
	columnTable.SetBackgroundColor(tcell.ColorDefault)
	columnTable.SetTitle(" " + apiColumn.Name + " ")
	columnTable.SetSelectedStyle(tcell.Color207, tcell.ColorDefault, tcell.AttrBold)
	columnTable.SetSelectable(false, false).Select(0, 0).SetFixed(0, 1)

	column := &Column{
		Table: columnTable,
		cards: []*Card{},
		Id:    colId,
	}
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
			tui.pos.focusCol = tui.getLeftPos()
			tui.view.columns.columns[tui.pos.focusCol].forcus(tui)
		case 'l':
			tui.view.columns.columns[tui.pos.focusCol].unForcus(tui)
			tui.pos.focusCol = tui.getRightPos()
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
			if len(tui.view.columns.columns[tui.pos.focusCol].cards) > 0 {
				row, _ := tui.view.columns.columns[tui.pos.focusCol].GetSelection()
				cardId := tui.view.columns.columns[tui.pos.focusCol].cards[row].Id
				ctx := context.Background()
				tui.view.columns.columns[tui.pos.focusCol].cards[row].SetTextColor(tcell.ColorOrange)
				go func() {
					tui.ghpjSettings.V3Client.Projects.MoveProjectCard(ctx, cardId, &github.ProjectCardMoveOptions{Position: "top", ColumnID: tui.view.columns.columns[tui.getRightPos()].Id})
					tui.notice.columUpdateDoneChan <- struct{}{}
				}()
			}
		case 'b':
			if len(tui.view.columns.columns[tui.pos.focusCol].cards) > 0 {
				row, _ := tui.view.columns.columns[tui.pos.focusCol].GetSelection()
				cardId := tui.view.columns.columns[tui.pos.focusCol].cards[row].Id
				ctx := context.Background()
				tui.view.columns.columns[tui.pos.focusCol].cards[row].SetTextColor(tcell.ColorOrange)
				go func() {
					tui.ghpjSettings.V3Client.Projects.MoveProjectCard(ctx, cardId, &github.ProjectCardMoveOptions{Position: "top", ColumnID: tui.view.columns.columns[tui.getLeftPos()].Id})
					tui.notice.columUpdateDoneChan <- struct{}{}
				}()
			}
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

func (c *Column) SetCards(cards []api.Card) {
	var colNum = 0
	for _, apiCard := range cards {
		var text string

		if apiCard.IsArchived {
			continue
		} else {
			if apiCard.Title != "" {
				text = apiCard.Title
			} else if apiCard.Note != "" {
				text = emoji.Sprintf(":memo:") + apiCard.Note
			}
		}
		card := &Card{
			TableCell: tview.NewTableCell(text),
		}
		cardIdRE := regexp.MustCompile(`.+:ProjectCard(\d+)$`)
		nodeId, _ := b64.StdEncoding.DecodeString(apiCard.Id)
		match := cardIdRE.FindStringSubmatch(string(nodeId))
		cardId, _ := strconv.ParseInt(match[1], 10, 64)

		card.Card = apiCard
		card.Id = cardId

		c.SetCell(colNum, 0, card.TableCell)
		c.cards = append(c.cards, card)
		colNum++
	}
}
