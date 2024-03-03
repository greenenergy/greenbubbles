package main

import (
	"fmt"
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/greenenergy/greenbubbles/itemeditor"
	"github.com/greenenergy/greenbubbles/teatree"
)

type ServerDefinition struct {
	Name        string        `json:"name"`
	Host        string        `json:"host"`
	SigningCert func() string `json:"signing_cert"`
	AuthPort    int           `json:"auth_port"`
	CmdPort     int           `json:"cmd_port"`
	certdata    string
}

func NewServerDefinition() *ServerDefinition {
	sd := ServerDefinition{}
	sd.SigningCert = sd.GetSigningCert
	return &sd
}

func (sd *ServerDefinition) GetSigningCert() string {
	if sd.certdata == "" {
		sd.certdata = "first set"
	} else {
		sd.certdata = "second set"
	}
	return sd.certdata
}

func New() *App {
	var app = App{
		ItemEditor: itemeditor.NewEditor(),
	}

	additem := teatree.NewItem("[Add Server]", false, nil, nil, nil, nil, nil, nil, nil)
	additem.SetSelectFunc(func(ti *teatree.TreeItem) {
		log.Println("--- at SelectFunc ---")
		app.ItemEditor.Tree.AddChildren(teatree.NewItem("<unnamed>", false, nil, nil, nil, nil, nil, nil, NewServerDefinition()))
	})
	app.ItemEditor.Tree.AddChildren(additem)

	serverDefs := [][2]string{
		{"dev", "localhost"},
		{"staging", "http://staging"},
		{"prod", "http://prod"},
	}

	var icon func(ti *teatree.TreeItem) string
	var iconStyle func(ti *teatree.TreeItem) lipgloss.Style
	var labelStyle func(ti *teatree.TreeItem) lipgloss.Style
	var openFunc func(*teatree.TreeItem)
	var closeFunc func(*teatree.TreeItem)
	var children []*teatree.TreeItem
	canHaveChildren := false

	addItem := func(name string, sd *ServerDefinition) error {
		item := teatree.NewItem(name, canHaveChildren, children, icon, labelStyle, iconStyle, openFunc, closeFunc, sd)
		app.ItemEditor.Tree.AddChildren(item)
		return nil
	}

	for _, def := range serverDefs {
		sd := NewServerDefinition()
		sd.Name = def[0]
		sd.Host = def[1]

		if err := addItem(def[0], sd); err != nil {
			log.Fatal(err)
		}
	}

	return &app
}

type App struct {
	ItemEditor *itemeditor.ItemCollectionEditor
	Width      int
	Height     int
	quitting   bool
}

func (a *App) Init() tea.Cmd {
	return nil
}

func (a *App) View() string {
	if a.quitting {
		return ""
	}
	return a.ItemEditor.View()
}

func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch tmsg := msg.(type) {
	case tea.KeyMsg:
		switch tmsg.String() {
		case "ctrl+c", "q":
			a.quitting = true
			return a, tea.Quit
		}
	}

	ie, cmd := a.ItemEditor.Update(msg)
	a.ItemEditor = ie.(*itemeditor.ItemCollectionEditor)
	return a, cmd
}

func main() {
	f, err := tea.LogToFile("debug.log", "debug")
	if err != nil {
		fmt.Println("problem opening log file:", err.Error())
		return
	}
	defer f.Close()

	m := New()
	p := tea.NewProgram(m)
	if _, err := p.Run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
