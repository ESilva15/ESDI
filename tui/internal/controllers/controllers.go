// Package controllers defines the controller stuff for our TUI
package controllers

import (
	"fmt"
	"log/slog"

	"github.com/rivo/tview"
)

type Controller struct {
	Logger *slog.Logger
	App    *tview.Application
}

func ListFormButtonLabels(form *tview.Form) []string {
	list := make([]string, form.GetButtonCount())
	for idx := range form.GetButtonCount() {
		btn := form.GetButton(idx)
		list = append(list, btn.GetLabel())
	}

	return list
}

func SetFormButtonCallback(form *tview.Form, btnLabel string, fn func()) error {
	btnIndex := form.GetButtonIndex(btnLabel)
	if btnIndex == -1 {
		availableButtons := ListFormButtonLabels(form)
		return fmt.Errorf("no button with label: `%s` : [%v]", availableButtons)
	}

	button := form.GetButton(btnIndex)
	if button == nil {
		return fmt.Errorf("not button with ID %d in form", btnIndex)
	}

	button.SetSelectedFunc(fn)

	return nil
}
