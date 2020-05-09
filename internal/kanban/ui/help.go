package ui

import (
	"fmt"
	"time"

	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
)

type Help struct {
	*tview.TextView
}

const DEFAULT_HELP_MESSAGE = `
h        move left column
l        move right column
j        move down
k        move up
CtrN     content down
CtrP     content up
R        refresh
P        open browser(Project Board)
p        open browser(Issue)`

var HELP_MESSAGE = func() string {
	t := time.Now()
	dateTime := t.Format("2006/1/2 15:04:05")

	message := fmt.Sprintf("last updated %s\n %s", dateTime, DEFAULT_HELP_MESSAGE)

	return message
}

func newHelp() *Help {
	textView := tview.NewTextView()
	textView.SetBackgroundColor(tcell.ColorDefault)
	textView.SetBorder(true)
	textView.SetTitle(" Infomation ")
	textView.SetTitleAlign(tview.AlignLeft)
	textView.SetText("last updated loading...\n" + DEFAULT_HELP_MESSAGE)

	help := &Help{
		TextView: textView,
	}

	return help
}

func (h *Help) Update() {
	h.SetText(HELP_MESSAGE())
}
