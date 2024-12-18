package display

import "github.com/rivo/tview"

// Page represents a single page in a wizard or config interface.
type Page struct {
	ID          string
	Title       string
	Help        string
	Description string
	Content     tview.Primitive
	Parent      *Page
}

// PageInterface defines the interface that all pages must implement.
type PageInterface interface {
	GetPage() *Page
}

// NewPage creates a new page with the given parameters.
func NewPage(parent *Page, id string, title string, help string, content tview.Primitive) *Page {
	return &Page{
		ID:      id,
		Title:   title,
		Help:    help,
		Content: content,
		Parent:  parent,
	}
}

// GetID returns the page's ID.
func (p *Page) GetID() string {
	return p.ID
}

// GetTitle returns the page's title.
func (p *Page) GetTitle() string {
	return p.Title
}

// GetHelp returns the page's help text.
func (p *Page) GetHelp() string {
	return p.Help
}

// GetContent returns the page's content.
func (p *Page) GetContent() tview.Primitive {
	return p.Content
}

// GetParent returns the page's parent page.
func (p *Page) GetParent() *Page {
	return p.Parent
}
