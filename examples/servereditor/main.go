package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/greenenergy/greenbubbles/filebrowser"
	"github.com/greenenergy/greenbubbles/itemeditor"
	"github.com/greenenergy/greenbubbles/teatree"
)

type ServerDefinition struct {
	Name        string        `json:"name"`
	Host        string        `json:"host"`
	SigningCert func() string `json:"-"`
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
		sd.certdata = "<unset>"
	}
	return sd.certdata
}

func New() *App {
	var app = App{
		ItemEditor: itemeditor.NewEditor(),
	}

	addServerLabelStyle := func(ti *teatree.TreeItem) lipgloss.Style {
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00FFFF")).
			Background(lipgloss.Color("#000030"))
	}

	additem := teatree.NewItem("[Add Server]", false, nil, nil, addServerLabelStyle, nil, nil, nil, nil)
	additem.SetSelectFunc(func(ti *teatree.TreeItem) {
		log.Println("at SelectFunc - should add a new child")
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
	ItemEditor  *itemeditor.ItemCollectionEditor
	Width       int
	Height      int
	quitting    bool
	popup       bool
	initialized bool
}

func (a *App) Init() tea.Cmd {
	return nil
}

// func (a *App) EditServerDefinition(sd *ServerDefinition) error {
func (a *App) EditServerDefinition(ti *teatree.TreeItem) error {
	sd := ti.Data.(*ServerDefinition)
	authPortStr := strconv.Itoa(sd.AuthPort)
	cmdPortStr := strconv.Itoa(sd.CmdPort)

	validPortFunc := func(str string) error {
		port, err := strconv.Atoi(str)
		if err != nil || port < 0 || port > 65535 {
			return errors.New("Sorry, must be a valid port number from 1-65535")
		}
		return nil
	}

	dummy, err := json.MarshalIndent(a, "", "    ")
	if err != nil {
		log.Println("Error formatting App:", err.Error())
	} else {
		log.Println("At EditServerDefinition:", string(dummy))
	}

	form := huh.NewForm(
		// Need to edit:
		huh.NewGroup(
			huh.NewInput().
				Value(&sd.Name).
				Title("Server name").
				Placeholder("dev").
				Description("Mnemonic for server"),
			huh.NewInput().
				Value(&sd.Host).
				Title("Host Addr").
				Placeholder("localhost").
				Description("host name"),
		),
		huh.NewGroup(
			huh.NewInput().
				Value(&authPortStr).
				Title("Auth Port").
				Placeholder("50001").
				Validate(validPortFunc).
				Description("Authentication port, only for login. 1-65535"),
			huh.NewInput().
				Value(&cmdPortStr).
				Title("Cmd Port").
				Placeholder("50000").
				Validate(validPortFunc).
				Description("Command port. All commands go here. 1-65535"),
		),
		huh.NewGroup(
			filebrowser.New("/").
				Value(&sd.certdata).
				WithWidth(a.Width).
				WithHeight(a.Height),
		),
	)
	err = form.Run()
	ti.Name = sd.Name
	sd.AuthPort, _ = strconv.Atoi(authPortStr)
	sd.CmdPort, _ = strconv.Atoi(cmdPortStr)

	return err
}

func (a *App) View() string {
	if a.quitting {
		return ""
	}
	v := a.ItemEditor.View()
	if !a.popup {
		return v
	}
	// Now do a form for the ServerDefinition

	if err := a.EditServerDefinition(a.ItemEditor.Tree.ActiveItem); err != nil {
		log.Println("*** Error editing server def:", err.Error())
	}
	a.popup = false

	return a.ItemEditor.View()
	/*
		// Draw a box 50% of the available space, in the center of the area
		left := lipgloss.Position(a.Width / 4
		top := a.Height / 4
		var t lipgloss.Position
		width := a.Width / 2
		height := a.Height / 2

		itemDump := itemeditor.IterateStructFields(a.ItemEditor.Tree.ActiveItem.Data)
		data := strings.Join(itemDump, "\n")
		var opts lipgloss.WhitespaceOption

		//newbox := lipgloss.NewStyle()

		box := lipgloss.Place(width, height, left, top, data, opts)
		return strings.Join([]string{v, box}, "\n")
	*/
}

func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch tmsg := msg.(type) {
	case tea.WindowSizeMsg:
		a.Width = tmsg.Width
		a.Height = tmsg.Height
		if !a.initialized {
			a.initialized = true
			return a, tea.ClearScreen
		}

	case tea.KeyMsg:
		log.Println("keymsg:", tmsg.String())
		switch tmsg.String() {
		case " ":
			a.popup = !a.popup
			log.Println("popup:", a.popup)
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
	var debug = flag.Bool("d", false, "create debug log")
	flag.Parse()

	// Since Bubbletea captures all console I/O, we can just write
	// everything to a logfile instead and tail that separately
	if debug != nil && *debug {
		f, err := tea.LogToFile("debug.log", "debug")
		if err != nil {
			fmt.Println("problem opening log file:", err.Error())
			return
		}
		defer f.Close()

	} else {
		// If there is no debug desired, then silence it
		log.SetOutput(io.Discard)
	}

	m := New()
	p := tea.NewProgram(m)
	if _, err := p.Run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
