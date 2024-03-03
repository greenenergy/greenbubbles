package itemeditor

import (
	"log"
	"reflect"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/greenenergy/greenbubbles/teatree"
)

func IterateStructFields(structInput interface{}) {
	val := reflect.ValueOf(structInput)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	// Check if the input is indeed a struct
	if val.Kind() != reflect.Struct {
		log.Println("Provided input is not a struct!")
		return
	}

	// Iterate over the struct fields
	for i := 0; i < val.NumField(); i++ {
		field := val.Type().Field(i)
		value := val.Field(i)

		// Check if the field is a function
		// Make sure it returns only one result, a string:
		if value.Kind() == reflect.Func && value.Type().NumOut() == 1 && value.Type().Out(0).Kind() == reflect.String {
			results := value.Call(nil) // Call the function without arguments
			if len(results) > 0 {
				// Capture and print the return value
				log.Printf("Field Name: %s, Field Type: %s, Function Return Value: %s\n", field.Name, field.Type, results[0].String())
			}
		} else {
			log.Printf("Field Name: %s, Field Type: %s, Field Value: %v\n", field.Name, field.Type, value)
		}
	}

	//// Iterate over the struct fields
	//for i := 0; i < val.NumField(); i++ {
	//	field := val.Type().Field(i)
	//	value := val.Field(i)

	//	// You can do something specific based on the field name, type, or value here
	//	fmt.Printf("Field Name: %s, Field Type: %s, Field Value: %v\n", field.Name, field.Type, value)
	//}
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
	Width       int
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

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		// The help box at the bottom spreads across both upper views
		log.Println("Got a size msg:", msg)
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
	return ice.Tree.View()
	//return ice.help.View(nil)
}
