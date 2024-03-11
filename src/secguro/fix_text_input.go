package main

import (
	"errors"
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/muesli/reflow/wordwrap"
)

func getTextInput(prompt string, defaultAnswer string) (string, error) {
	p := tea.NewProgram(initialModelTextInput(prompt, defaultAnswer), tea.WithAltScreen())

	// Run returns the model as a tea.Model.
	m, err := p.Run()
	if err != nil {
		return "", err
	}

	// Assert the final tea.Model to the local model and return the final state.
	if m, ok := m.(modelTextInput); ok {
		return m.textInput.Value(), nil
	}

	return "", errors.New("text input terminated with error due to failed type assertion")
}

type (
	errMsg error
)

type modelTextInput struct {
	windowWidth int
	prompt      string
	textInput   textinput.Model
	err         error
}

func initialModelTextInput(prompt string, defaultAnswer string) modelTextInput {
	ti := textinput.New()
	ti.Placeholder = ""
	ti.Focus()
	ti.CharLimit = 156
	ti.Width = -1
	ti.SetValue(defaultAnswer)

	return modelTextInput{
		windowWidth: 0,
		prompt:      prompt,
		textInput:   ti,
		err:         nil,
	}
}

func (m modelTextInput) Init() tea.Cmd {
	return textinput.Blink
}

func (m modelTextInput) Update(msg tea.Msg) (tea.Model, tea.Cmd) { //nolint: ireturn // must be like this
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.windowWidth = msg.Width
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

	return m, cmd
}

func (m modelTextInput) View() string {
	return wordwrap.String(fmt.Sprintf(
		m.prompt+"\n\n%s\n\n%s",
		m.textInput.View(),
		"(esc to quit)",
	)+"\n", m.windowWidth)
}
