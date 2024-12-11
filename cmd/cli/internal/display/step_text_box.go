package display

import (
	"fmt"

	"github.com/rivo/tview"
)

type page struct {
	parent      *page
	id          string
	title       string
	description string
	content     tview.Primitive
}

type TextBoxStep struct {
	Wizard      Wizard
	Modal       *TextBoxModalLayout
	Step, Total int
	showImpl    func(*TextBoxModalLayout)
}

type TextBoxStepOptions struct {
	Step       int
	Total      int
	Title      string
	HelperText string
	Width      int
	Labels     []string
	MaxLengths []int
	Regexes    []string
	OnDone     func(map[string]string)
	OnBack     func()
	PageID     string
}

func NewTextBoxStep(w Wizard, opts TextBoxStepOptions) *TextBoxStep {
	step := &TextBoxStep{
		Wizard:   w,
		Step:     opts.Step,
		Total:    opts.Total,
		showImpl: nil,
	}

	title := fmt.Sprintf("[%d/%d] %s", opts.Step, opts.Total, opts.Title)

	modal := NewTextBoxModal(w.GetApp(), TextBoxModalOptions{
		Title:      title,
		Width:      opts.Width,
		Text:       opts.HelperText,
		Labels:     opts.Labels,
		MaxLengths: opts.MaxLengths,
		Regexes:    opts.Regexes,
		OnDone:     opts.OnDone,
		OnBack:     opts.OnBack,
	})

	step.Modal = modal

	page := newPage(nil, opts.PageID, "Config Wizard > "+title, "", modal.BorderGrid)
	w.GetPages().AddPage(page.id, page.content, true, false)
	modal.Page = page

	return step
}

func (s *TextBoxStep) Show() error {
	s.showImpl(s.Modal)
	return nil
}

func (s *TextBoxStep) Next() (WizardStep, error) {
	return s.Wizard.GetCurrentStep(), nil
}

func (s *TextBoxStep) Previous() (WizardStep, error) {
	return s.Wizard.GetCurrentStep(), nil
}

func (s *TextBoxStep) GetTitle() string {
	return s.Modal.Title
}

func (s *TextBoxStep) GetProgress() (int, int) {
	return s.Step, s.Total
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
