package filebrowser

import (
	"encoding/json"
	"io/fs"
	"log"
	"path"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/greenenergy/greenbubbles/teatree"
)

const IconFolder = "\U000F024B"
const IconFile = "\U000F0214"

const GoGopherDev = "\ue626"
const GoGopher = "\ue724"
const GoTitle = "\U000F07D3"

type FileBrowserModel struct {
	dir      string
	result   *string
	Tree     *teatree.Tree
	info     func()
	quitting bool
	err      error
}

func (fbm *FileBrowserModel) Value(value *string) *FileBrowserModel {
	fbm.result = value
	return fbm
}

func (fbm *FileBrowserModel) Blur() tea.Cmd {
	return fbm.Tree.Blur()
}

func (fbm *FileBrowserModel) Focus() tea.Cmd {
	return fbm.Tree.Focus()
}

func (fbm *FileBrowserModel) Error() error {
	return fbm.err
}

func (fbm *FileBrowserModel) Run() error {
	return fbm.Tree.Run()
}

// Skip returns whether this input should be skipped or not.
func (fbm *FileBrowserModel) Skip() bool {
	return fbm.Tree.Skip()
}

func (fbm *FileBrowserModel) KeyBinds() []key.Binding {
	return fbm.Tree.KeyBinds()
}

// GetValue returns the field's value.
func (fbm *FileBrowserModel) GetValue() any {
	return fbm.Tree.GetValue()
}

// GetKey returns the field's key.
func (fbm *FileBrowserModel) GetKey() string {
	return fbm.Tree.GetKey()
}

// WithHeight sets the height of the input field.
func (fbm *FileBrowserModel) WithHeight(height int) huh.Field {
	fbm.Tree.WithHeight(height)
	return fbm
}

// WithWidth sets the width of the input field.
func (fbm *FileBrowserModel) WithWidth(width int) huh.Field {
	fbm.Tree.WithWidth(width)
	return fbm
}

// WithPosition sets the position of the input field.
func (fbm *FileBrowserModel) WithPosition(p huh.FieldPosition) huh.Field {
	//i.keymap.Prev.SetEnabled(!p.IsFirst())
	//i.keymap.Next.SetEnabled(!p.IsLast())
	//i.keymap.Submit.SetEnabled(p.IsLast())
	fbm.Tree.WithPosition(p)
	return fbm
}

// WithTheme sets the theme on a field.
func (fbm *FileBrowserModel) WithTheme(theme *huh.Theme) huh.Field {
	fbm.Tree.WithTheme(theme)
	return fbm
}

// WithAccessible sets the accessible mode of the input field.
func (fbm *FileBrowserModel) WithAccessible(accessible bool) huh.Field {
	fbm.Tree.WithAccessible(accessible)
	return fbm
}

// WithKeyMap sets the keymap on an input field.
func (fbm *FileBrowserModel) WithKeyMap(k *huh.KeyMap) huh.Field {
	fbm.Tree.WithKeyMap(k)
	//t.textinput.KeyMap.AcceptSuggestion = i.keymap.AcceptSuggestion
	return fbm
}

func (fm *FileBrowserModel) Init() tea.Cmd {
	return nil
}

func (fm *FileBrowserModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch tmsg := msg.(type) {
	/*
		// This cuts the display area in half to make debugging easier. We can see accidental scrolls, for example.
			case tea.WindowSizeMsg:
				tmsg.Height = tmsg.Height / 2
				log.Printf("WindowSizeMsg: Halving height: Height: %d", tmsg.Height)
				msg = tmsg
	*/

	case tea.KeyMsg:
		switch tmsg.String() {
		case "enter":
			// TODO: If you select something in your current directory ".", then the file will be named
			// ".whatever" instead of "./whatever". For some reason the slash is not imserted between
			// the value of fm.dir and the first actual path value.
			fullList := append([]string{fm.dir}, fm.Tree.ActiveItem.GetPath()...)
			log.Println("returning:", fullList)

			res := path.Join(fullList...)
			if fm.result != nil {
				*fm.result = res
			}

			log.Println("result:", fm.result)
			fm.quitting = true
			return fm, tea.Quit

		case "r": // Refresh - it will cause the parent of the currently selected item to delete all children and re-fetch them.
			parent := fm.Tree.ActiveItem.GetParent()
			parent.Refresh()
			if _, ok := parent.(*teatree.Tree); ok {
				// If we're already at the top level, it means a refresh of the root tree
				fm.Tree.ActiveItem = fm.Tree.Items[0]
			} else {
				// If we're not at the top, the simplest thing is to just activate the parent of the
				// current item and close it for re-opening
				fm.Tree.ActiveItem = parent.(*teatree.TreeItem)
			}

		case "ctrl+c", "q":
			fm.quitting = true
			return fm, tea.Quit

		case "?":
			if fm.info != nil {
				fm.info()
			}
			return fm, nil
		}
	}
	_, cmd := fm.Tree.Update(msg)
	return fm, cmd
}

func (fm *FileBrowserModel) View() string {
	if fm.quitting {
		return ""
	}
	dummy, err := json.MarshalIndent(fm, "", "    ")
	if err != nil {
		log.Println("error formatting filebrowser:", err.Error())
	} else {
		log.Println("at FileBrowserModel.View():", string(dummy))
	}
	return fm.Tree.View()
}

func TextColor(ti *teatree.TreeItem) lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFFFF")) // white
}

func FolderIcon(ti *teatree.TreeItem) string {
	return IconFolder
}

func FolderColor(ti *teatree.TreeItem) lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFCF00")) // yellow
}

func GoFileIcon(ti *teatree.TreeItem) string {
	return GoTitle
}

func GoFileColor(ti *teatree.TreeItem) lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00FFFF")) // cyan
}

func FileIcon(ti *teatree.TreeItem) string {
	return IconFile
}

func FileColor(ti *teatree.TreeItem) lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("#7FFF7F")) // palegreen
}

func (fm *FileBrowserModel) walk(p string, item teatree.ItemHolder) error {
	err := filepath.WalkDir(p, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			log.Printf("prevent panic by handling failure accessing a path %q: %v\n", path, err)
			return err
		}
		if path == p {
			// We don't want to render the folder we were sent. This is redundant and confusing for the user.
			return nil
		}

		var icon func(ti *teatree.TreeItem) string
		var iconStyle func(ti *teatree.TreeItem) lipgloss.Style
		var labelStyle func(ti *teatree.TreeItem) lipgloss.Style

		canHaveChildren := d.IsDir()

		labelStyle = TextColor
		// Default the icon and style to basic file style and icon
		// These will be overridden by the next if statement
		icon = FileIcon
		iconStyle = FileColor

		if d.IsDir() {
			icon = FolderIcon
			iconStyle = FolderColor
		} else {
			if strings.HasSuffix(d.Name(), ".go") {
				icon = GoFileIcon
				iconStyle = GoFileColor
			}
		}

		openFunc := func(ti *teatree.TreeItem) {
			// This function is called when the user toggles an item that can have children. For now that only means this is a folder and we are now supposed to walk the ti's path, adding items
			// If we have no children, then we should walk the directory.
			// If we DO have children, they will be cached. Added a "r" refresh handler to
			// cause a directory to be re-read, to take care of directory changes.
			if len(ti.Children) == 0 {
				err = fm.walk(path, ti)
			}
		}
		var children []*teatree.TreeItem
		newitem := teatree.NewItem(d.Name(), canHaveChildren, children, icon, labelStyle, iconStyle, openFunc, nil, nil)
		item.AddChildren(newitem)

		if d.IsDir() && path != p {
			// Do not descend into subdirectories
			return filepath.SkipDir
		}

		return nil
	})

	if err != nil {
		log.Printf("error walking the path %q: %v\n", p, err)
		return err
	}
	return nil
}

// func New(dir string) tea.Model {
func New(dir string) *FileBrowserModel {
	fm := &FileBrowserModel{
		dir:  dir,
		Tree: teatree.New().(*teatree.Tree),
	}
	fm.info = func() {
		log.Print("INFO")
		log.Println("testing")
	}
	if err := fm.walk(dir, fm.Tree); err != nil {
		log.Fatal(err)
	}
	return fm
}
