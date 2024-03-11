package filebrowser

import (
	"io/fs"
	"log"
	"path"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
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
	result   string
	Tree     *teatree.Tree
	info     func()
	quitting bool
	err      error
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

func (fbm *FileBrowserModel) KeyBinds() []key.Binding {
	return fbm.Tree.KeyBinds()
}

/*
type Field interface {
	// WithTheme sets the theme on a field.
	WithTheme(*Theme) Field

	// WithAccessible sets whether the field should run in accessible mode.
	WithAccessible(bool) Field

	// WithKeyMap sets the keymap on a field.
	WithKeyMap(*KeyMap) Field

	// WithWidth sets the width of a field.
	WithWidth(int) Field

	// WithHeight sets the height of a field.
	WithHeight(int) Field

	// WithPosition tells the field the index of the group and position it is in.
	WithPosition(FieldPosition) Field

	// GetKey returns the field's key.
	GetKey() string

	// GetValue returns the field's value.
	GetValue() any
}

*/

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
			fm.result = path.Join(fullList...)
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

func New(dir string) tea.Model {
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

func (t *FileBrowserModel) Value() string {
	return t.result
}
