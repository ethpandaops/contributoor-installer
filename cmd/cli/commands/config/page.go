package config

import "github.com/rivo/tview"

type page struct {
	parent      *page
	id          string
	title       string
	description string
	content     tview.Primitive
}

func newPage(parent *page, id string, title string, description string, content tview.Primitive) *page {
	return &page{
		parent:      parent,
		id:          id,
		title:       title,
		description: description,
		content:     content,
	}
}

func (p *page) getHeader() string {
	if p.parent == nil {
		return p.title
	}
	return p.parent.getHeader() + " > " + p.title
}

type settingsPage interface {
	getPage() *page
	handleLayoutChanged()
}
