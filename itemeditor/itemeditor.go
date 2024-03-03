package itemeditor

import (
	"fmt"
	"reflect"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/greenenergy/greenbubbles/teatree"
)

func IterateStructFields(structInput interface{}) []string {
	var results []string
	val := reflect.ValueOf(structInput)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	// Check if the input is indeed a struct
	if val.Kind() != reflect.Struct {
		return nil
	}

	// Iterate over the struct fields
	for i := 0; i < val.NumField(); i++ {
		field := val.Type().Field(i)
		value := val.Field(i)

		// Check if the field is a function
		// Make sure it returns only one result, a string:
		if value.Kind() == reflect.Func && value.Type().NumOut() == 1 && value.Type().Out(0).Kind() == reflect.String {
			funcResults := value.Call(nil) // Call the function without arguments
			if len(funcResults) > 0 {
				// Capture and print the return value
				//log.Printf("Field Name: %s, Field Type: %s, Function Return Value: %s\n", field.Name, field.Type, funcResults[0].String())
				results = append(results, fmt.Sprintf("%s:%s", field.Name, funcResults[0].String()))
			}
		} else {
			//log.Printf("Field Name: %s, Field Type: %s, Field Value: %v\n", field.Name, field.Type, value)
			results = append(results, fmt.Sprintf("%s:%s", field.Name, value))
		}
	}

	return results
}

type ItemEntry struct {
	Data interface{}
}

type ItemList struct {
	Items []*ItemEntry
}

func (il *ItemList) AddItem(ie *ItemEntry) error {
	il.Items = append(il.Items, ie)
	return nil
}

type ItemCollectionEditor struct {
	Width       int // The width is for the whole control, giving the tree view the left half (so half this value)
	Height      int
	initialized bool
	Tree        *teatree.Tree
	//help        *help.Model
	quitting bool
}

func NewEditor() *ItemCollectionEditor {
	return &ItemCollectionEditor{
		Tree: teatree.New().(*teatree.Tree),
	}
}

func (ice *ItemCollectionEditor) Init() tea.Cmd {
	return nil
}

func (ice *ItemCollectionEditor) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if !ice.initialized {
		ice.initialized = true
	}

	switch tmsg := msg.(type) {
	case tea.WindowSizeMsg:
		// The help box at the bottom spreads across both upper views
		ice.Width = tmsg.Width
		ice.Height = tmsg.Height
		// Get the full width and then pass half width onto the tree
		msg = tea.WindowSizeMsg{
			Width:  tmsg.Width / 2,
			Height: tmsg.Height,
		}
	}
	_, cmd := ice.Tree.Update(msg)
	return ice, cmd
}

func (ice *ItemCollectionEditor) View() string {
	if !ice.initialized {
		return "not initialized"
	}
	if ice.quitting {
		return "Bye!\n"
	}
	treeview := ice.Tree.View()
	itemDump := IterateStructFields(ice.Tree.ActiveItem.Data)
	activeView := strings.Join(itemDump, "\n")

	s := lipgloss.JoinHorizontal(
		lipgloss.Top, treeview, activeView,
	)
	return s
}
