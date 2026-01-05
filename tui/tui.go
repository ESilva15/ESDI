package tui

import (
	"fmt"
	"os"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func Init() error {
	cm := getCommands()

	app := tview.NewApplication()

	output := tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true).
		SetWrap(true)

	output.
		SetBorder(true).
		SetTitle("OUTPUT:")

	input := tview.NewInputField().
		SetLabel("> ").
		SetFieldWidth(0)

	input.
		SetBorder(true).
		SetTitle("INPUT:")

	input.SetDoneFunc(func(key tcell.Key) {
		if key != tcell.KeyEnter {
			return
		}

		text := input.GetText()
		fmt.Fprintf(output, "> %s\n", text)

		args, err := parseArgs(text)
		if err != nil {
			fmt.Fprintf(output, "ERROR\n%s", err.Error())
			return
		}

		cmdOutput, err := cm.RunCommand(args)
		if err != nil {
			fmt.Fprintf(output, "ERROR\n%s", err.Error())
			return
		} else {
			fmt.Fprintf(output, "%s\n", cmdOutput)
		}

		input.SetText("")
		output.ScrollToEnd()
	})

	layout := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(output, 0, 1, false).
		AddItem(input, 3, 0, true)

	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyCtrlC {
			app.Stop()
			return nil
		}
		return event
	})

	if err := app.SetRoot(layout, true).Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}

	return nil
}
