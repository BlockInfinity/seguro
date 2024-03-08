package main

import (
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// TODO: get answer in cleaner way; like in fix_choose_option
var providedAnswer string

func getTextInput(prompt string, defaultAnswer string) (string, error) {
	p := tea.NewProgram(initialModelTextInput(prompt, defaultAnswer), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		return "", err
	}

	return providedAnswer, nil
}

type (
	errMsg error
)

type modelTextInput struct {
	prompt    string
	textInput textinput.Model
	err       error
}

func initialModelTextInput(prompt string, defaultAnswer string) modelTextInput {
	ti := textinput.New()
	ti.Placeholder = ""
	ti.Focus()
	ti.CharLimit = 156
	ti.Width = -1
	ti.SetValue(defaultAnswer)

	return modelTextInput{
		prompt:    prompt,
		textInput: ti,
		err:       nil,
	}
}

func (m modelTextInput) Init() tea.Cmd {
	return textinput.Blink
}

func (m modelTextInput) Update(msg tea.Msg) (tea.Model, tea.Cmd) { //nolint: ireturn // must be like this
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg: //nolint: exhaustive
		switch msg.Type {
		case tea.KeyEnter, tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		}

	// We handle errors just like any other message
	case errMsg:
		m.err = msg
		return m, nil
	}

	m.textInput, cmd = m.textInput.Update(msg)
	providedAnswer = m.textInput.Value()

	return m, cmd
}

func (m modelTextInput) View() string {
	return fmt.Sprintf(
		m.prompt+"\n\n%s\n\n%s",
		m.textInput.View(),
		"(esc to quit)",
	) + "\n"
}
