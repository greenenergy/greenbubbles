package itemeditor

import (
	"fmt"
	"log"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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

func TestItemCollectionEditor(t *testing.T) {
	editor := NewEditor()

	serverDefs := [][2]string{
		{"dev", "localhost"},
		{"staging", "localhost"},
		{"prod", "localhost"},
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
		editor.Tree.AddChildren(item)
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

	editor.Update(tea.WindowSizeMsg{
		Width:  50,
		Height: 50,
	})
	fmt.Println(editor.View())
	IterateStructFields(editor.Tree)
	fmt.Println(editor.View())
}
