package ui

import (
	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
)

type Content struct {
	*tview.TextView
}

func newContent() *Content {
	textView := tview.NewTextView()
	textView.SetBackgroundColor(tcell.ColorDefault)
	textView.SetBorder(true)
	textView.SetDynamicColors(true)
	textView.SetScrollable(true)

	return &Content{
		TextView: textView,
	}
}
