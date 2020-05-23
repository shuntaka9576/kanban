package ui

import (
	"github.com/rivo/tview"
	"github.com/shuntaka9576/kanban/api"
)

type Card struct {
	*tview.TableCell
	Id int64
	api.Card
}
