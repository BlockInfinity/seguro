package main

import (
	"errors"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/muesli/reflow/wordwrap"
)

func getOptionChoice(prompt string, choices []string) (int, error) {
	if len(choices) == 0 {
		return 0, errors.New("empty array given for choices")
	}

	p := tea.NewProgram(initialModelChooseOption(prompt, choices), tea.WithAltScreen())

	// Run returns the model as a tea.Model.
	m, err := p.Run()
	if err != nil {
		return 0, err
	}

	// Assert the final tea.Model to the local model and return the final state.
	if m, ok := m.(modelChooseOption); ok && m.cursor >= 0 {
		return m.cursor, nil
	}

	return 0, errors.New("option chooser terminated with error due to failed " +
		"type assertion")
}

type modelChooseOption struct {
	windowWidth int
	prompt      string
	choices     []string
	cursor      int
}

func (m modelChooseOption) Init() tea.Cmd {
	return nil
}

func (m modelChooseOption) Update(msg tea.Msg) (tea.Model, tea.Cmd) { //nolint: ireturn // must be like this
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.windowWidth = msg.Width
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			m.cursor = -1
			return m, tea.Quit

		case "enter":
			return m, tea.Quit

		case "down", "j":
			m.cursor++
			if m.cursor >= len(m.choices) {
				m.cursor = 0
			}

		case "up", "k":
			m.cursor--
			if m.cursor < 0 {
				m.cursor = len(m.choices) - 1
			}
		}
	}

	return m, nil
}

func (m modelChooseOption) View() string {
	s := strings.Builder{}
	s.WriteString(m.prompt + "\n\n")

	for i := 0; i < len(m.choices); i++ {
		if m.cursor == i {
			s.WriteString("(•) ")
		} else {
			s.WriteString("( ) ")
		}
		s.WriteString(m.choices[i])
		s.WriteString("\n")
	}
	s.WriteString("\n(esc to go back)\n")

	return wordwrap.String(s.String(), m.windowWidth)
}

func initialModelChooseOption(prompt string, choices []string) modelChooseOption {
	return modelChooseOption{
		windowWidth: 0,
		prompt:      prompt,
		choices:     choices,
		cursor:      0,
	}
}
