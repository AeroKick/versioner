package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

const listHeight = 3

type FileDetail struct {
	File  string `json:"file"`
	Field string `json:"field"`
}

type Version struct {
	Major int
	Minor int
	Patch int
}

type model struct {
	list       list.Model
	choice     string
	quitting   bool
	versioners []FileDetail
}

func (m *model) bumpVersion(bumpType string) {
	for _, v := range m.versioners {
		content, err := ioutil.ReadFile(v.File)
		if err != nil {
			fmt.Println("Error reading file:", err)
			os.Exit(1)
		}

		var data map[string]interface{}
		if err := json.Unmarshal(content, &data); err != nil {
			fmt.Println("Error parsing JSON:", err)
			os.Exit(1)
		}

		versionStr, ok := data[v.Field].(string)
		if !ok {
			fmt.Printf("Invalid version format in %s\n", v.File)
			os.Exit(1)
		}

		versionParts := strings.Split(versionStr, ".")
		if len(versionParts) != 3 {
			fmt.Printf("Invalid version format in %s\n", v.File)
			os.Exit(1)
		}

		version := Version{}
		fmt.Sscanf(versionStr, "%d.%d.%d", &version.Major, &version.Minor, &version.Patch)

		switch bumpType {
		case "Major":
			version.Major++
			version.Minor = 0
			version.Patch = 0
		case "Minor":
			version.Minor++
			version.Patch = 0
		case "Patch":
			version.Patch++
		}

		newVersion := fmt.Sprintf("%d.%d.%d", version.Major, version.Minor, version.Patch)
		data[v.Field] = newVersion

		newContent, err := json.MarshalIndent(data, "", "  ")
		if err != nil {
			fmt.Println("Error marshalling JSON:", err)
			os.Exit(1)
		}

		if err := ioutil.WriteFile(v.File, newContent, 0644); err != nil {
			fmt.Println("Error writing file:", err)
			os.Exit(1)
		}

		fmt.Printf("Updated %s to version %s\n", v.File, newVersion)
	}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			m.quitting = true
			return m, tea.Quit

		case "enter":
			if m.choice == "" {
				i, ok := m.list.SelectedItem().(string)
				if ok {
					m.choice = i
					m.bumpVersion(m.choice)
				}
			}
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m model) View() string {
	if m.choice != "" {
		return fmt.Sprintf("Bumped version using %s", m.choice)
	}
	if m.quitting {
		return "Exiting without updating."
	}
	return m.list.View()
}

func main() {
	content, err := ioutil.ReadFile("versioner.json")
	if err != nil {
		fmt.Println("Error reading versioner.json:", err)
		os.Exit(1)
	}

	var versioners []FileDetail
	if err := json.Unmarshal(content, &versioners); err != nil {
		fmt.Println("Error decoding versioner.json:", err)
		os.Exit(1)
	}

	items := []string{"Major", "Minor", "Patch"}

	l := list.NewModel(items, nil, 0, listHeight)
	l.Title = "Choose version bump type: Major, Minor, Patch"
	l.SetShowStatusBar(false)

	m := model{list: l, versioners: versioners}

	p := tea.NewProgram(m)
	if err := p.Start(); err != nil {
		fmt.Println("Error starting program:", err)
		os.Exit(1)
	}
}
