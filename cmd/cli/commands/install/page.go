package install

import (
	"github.com/ethpandaops/contributoor-installer/internal/display"
	"github.com/rivo/tview"
)

type page = display.Page

func newPage(parent *page, id string, title string, help string, content tview.Primitive) *page {
	return display.NewPage(parent, id, title, help, content)
}
