package repl

import (
	esdi "esdi/esdi"

	tea "github.com/charmbracelet/bubbletea"
)

type model struct {
	esdi *esdi.ESDI
	mov  string
}

func initialMode() model {
	return model{
		esdi: esdi.NewESDI(),
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "left", "h":
			// Go left
			m.mov = msg.String()
		case "down", "j":
			// Go down
			m.mov = msg.String()
		case "up", "k":
			// Go up
			m.mov = msg.String()
		case "right", "l":
			// Go right
			m.mov = msg.String()
		}
	}

	return m, nil
}

func (m model) View() string {
	s := "What should we buy at the market?\n\n"
	s += m.mov + "\n"

	return s
}

func Run() error {
	p := tea.NewProgram(initialMode())
	if _, err := p.Run(); err != nil {
		return err
	}

	return nil
}
