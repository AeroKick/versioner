package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const listHeight = 10

var (
	titleStyle        = lipgloss.NewStyle().MarginLeft(2)
	itemStyle         = lipgloss.NewStyle().PaddingLeft(4)
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("170"))
	paginationStyle   = list.DefaultStyles().PaginationStyle.PaddingLeft(4)
	helpStyle         = list.DefaultStyles().HelpStyle.PaddingLeft(4).PaddingBottom(1)
	quitTextStyle     = lipgloss.NewStyle().Margin(1, 0, 2, 4)
)

type item struct {
	title, desc string
}

func (i item) FilterValue() string { return "" }

type itemDelegate struct{}

func (d itemDelegate) Height() int                             { return 1 }
func (d itemDelegate) Spacing() int                            { return 0 }
func (d itemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(item)
	if !ok {
		return
	}

	str := fmt.Sprintf("%d. %s", index+1, i.desc)

	fn := itemStyle.Render
	if index == m.Index() {
		fn = func(s ...string) string {
			return selectedItemStyle.Render("> " + strings.Join(s, " "))
		}
	}

	fmt.Fprint(w, fn(str))
}

type VersionFile struct {
	File  string `json:"file"`
	Field string `json:"field"`
}

type VersionUpdate struct {
	File        string
	Current     string
	Updated     string
	UpdateError error
}

type model struct {
	list   list.Model
	choice item
}

func (m *model) Init() tea.Cmd {
	return nil
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetWidth(msg.Width)
		return m, nil

	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "ctrl+c":
			return m, tea.Quit

		case "enter":
			i, ok := m.list.SelectedItem().(item)
			if ok {
				m.choice = i
			}
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m model) View() string {
	return "\n" + m.list.View()
}

// Semantic Version struct
type SemanticVersion struct {
	Major int
	Minor int
	Patch int
}

func (v *SemanticVersion) Bump(typ string) {
	switch typ {
	case "Major":
		v.Major++
		v.Minor = 0
		v.Patch = 0
	case "Minor":
		v.Minor++
		v.Patch = 0
	case "Patch":
		v.Patch++
	}
}

func (v SemanticVersion) String() string {
	return fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
}

func ParseSemanticVersion(s string) (SemanticVersion, error) {
	var v SemanticVersion
	parts := strings.Split(s, ".")
	if len(parts) != 3 {
		return v, errors.New("invalid semantic version")
	}

	fmt.Sscanf(s, "%d.%d.%d", &v.Major, &v.Minor, &v.Patch)
	return v, nil
}

func ReadVersionerFile() ([]VersionFile, error) {
	data, err := ioutil.ReadFile("versioner.json")
	if err != nil {
		return nil, err
	}

	var versionFiles []VersionFile
	err = json.Unmarshal(data, &versionFiles)
	return versionFiles, err
}

func BumpFileVersion(file VersionFile, bumpType string) VersionUpdate {
	data, err := os.ReadFile(file.File)
	if err != nil {
		return VersionUpdate{File: file.File, UpdateError: err}
	}

	var content map[string]interface{}
	err = json.Unmarshal(data, &content)
	if err != nil {
		return VersionUpdate{File: file.File, UpdateError: err}
	}

	versionStr, ok := content[file.Field].(string)
	if !ok {
		return VersionUpdate{File: file.File, UpdateError: errors.New("invalid version field")}
	}

	version, err := ParseSemanticVersion(versionStr)
	if err != nil {
		return VersionUpdate{File: file.File, UpdateError: err}
	}

	version.Bump(bumpType)
	content[file.Field] = version.String()

	updatedData, err := json.MarshalIndent(content, "", "  ")
	if err != nil {
		return VersionUpdate{File: file.File, UpdateError: err}
	}

	err = ioutil.WriteFile(file.File, updatedData, 0644)
	if err != nil {
		return VersionUpdate{File: file.File, UpdateError: err}
	}

	return VersionUpdate{File: file.File, Current: versionStr, Updated: version.String()}
}

func main() {
	versionFiles, err := ReadVersionerFile()
	if err != nil {
		fmt.Println("Error reading versioner.json:", err)
		os.Exit(1)
	}

	// Ask the user for the bump type using the provided list code
	items := []list.Item{
		item{title: "Major", desc: "Major version bump"},
		item{title: "Minor", desc: "Minor version bump"},
		item{title: "Patch", desc: "Patch version bump"},
	}

	l := list.New(items, itemDelegate{}, 20, listHeight)
	l.Title = "Select bump type"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Styles.Title = titleStyle
	l.Styles.PaginationStyle = paginationStyle
	l.Styles.HelpStyle = helpStyle
	m := model{list: l}

	p := tea.NewProgram(&m, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}

	selected := m.choice.title

	log.Println("Selected:", selected)

	var updates []VersionUpdate
	for _, vf := range versionFiles {
		updates = append(updates, BumpFileVersion(vf, selected))
	}

	for _, update := range updates {
		if update.UpdateError != nil {
			fmt.Printf("Error updating %s: %s\n", update.File, update.UpdateError)
			continue
		}
		fmt.Printf("Updated %s from %s to %s\n", update.File, update.Current, update.Updated)
	}
}
