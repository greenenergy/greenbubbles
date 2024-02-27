package teatree

import (
	"log"
	"strings"
	"sync"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// This is the material design icon in the nerdfont/material symbols set, as found in https://pictogrammers.com/library/mdi/
const NoChevron = " "
const ChevronRight = "\U000F0142"
const ChevronDown = "\U000F0140"

type ItemHolder interface {
	GetItems() []*TreeItem
	// GetPath - recursively search through the parent hierarchy and return the name of each item
	// They will be ordered from oldest ancestor to most recent descendant. The item itself will
	// be the last one in the list
	GetPath() []string
	AddChildren(...*TreeItem) ItemHolder
	GetParent() ItemHolder
	Refresh() // This tells the item holder to delete all of its children and re-read them.
}

// Styling

var (
	unfocusedStyle = lipgloss.NewStyle()
	// Border(lipgloss.HiddenBorder()).
	// BorderTop(false).
	// BorderBottom(false)
	focusedStyle = lipgloss.NewStyle().
		//Border(lipgloss.RoundedBorder()).
		//BorderTop(false).
		//BorderBottom(false).
		Background(lipgloss.Color("62")).
		BorderForeground(lipgloss.Color("62"))
	//Background(lipgloss.Color("#FFFFFF"))
)

type TreeItem struct {
	sync.Mutex
	ParentTree      *Tree
	Parent          ItemHolder
	Name            string
	Children        []*TreeItem
	CanHaveChildren bool // CanHaveChildren: By setting this to True, you say that this item can have children. This allows for the implementation of a lazy loader, when you supply an Open() function. This affects how the item is rendered.
	Open            bool
	Data            interface{}
	OpenFunc        func(*TreeItem)
	CloseFunc       func(*TreeItem)
	icon            func(*TreeItem) string         // Function returns what the icon should be.
	labelStyle      func(*TreeItem) lipgloss.Style // Function returns the style for the label, intended for color
	iconStyle       func(*TreeItem) lipgloss.Style // Function returns the style for the icon, intended for color
	indent          int
}

func (ti *TreeItem) Icon() string {
	if ti.icon != nil {
		return ti.icon(ti)
	}
	return ""
}
func (ti *TreeItem) IconStyle() lipgloss.Style {
	if ti.iconStyle != nil {
		return ti.iconStyle(ti)
	}
	return lipgloss.NewStyle()
}

func (ti *TreeItem) LabelStyle() lipgloss.Style {
	if ti.labelStyle != nil {
		return ti.labelStyle(ti)
	}
	return lipgloss.NewStyle()
}

func (ti *TreeItem) GetParent() ItemHolder {
	return ti.Parent
}

func (ti *TreeItem) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return ti, nil
}

func (ti *TreeItem) Init() tea.Cmd {
	return nil
}

func (ti *TreeItem) Refresh() {
	ti.Children = []*TreeItem{}
	ti.Open = false
}

func (ti *TreeItem) GetItems() []*TreeItem {
	return ti.Children
}

// SelectPrevious - this is being invoked on a TreeItem that is currently selected and the
// user wants to move up to the previous selection. This will involve recursively going up the
// tree until we find the one to select or we get to the top
//
// TODO: SelectPrevious() does not work in the following situation:
//
// 󰅀 Item 1
//    Sub 1
//      SubSub 1   **TOHERE
// 󰅀 Item 2	**FROMHERE
//    AA
//    BB
// 󰅂 Item 3

func (ti *TreeItem) SelectPrevious() {

	// User has pressed 'up'. If we're at the top of the "view", then we want to try and scroll up by one
	// We do this by checking the viewtop value. If it is > 0, then the view has scrolled down, so we
	// can decrement it by one.

	tree := ti.ParentTree
	atTop := false

	if tree.ActiveLine == 0 {
		atTop = true
	} else {
		tree.ActiveLine -= 1
	}

	siblingItems := ti.Parent.GetItems()

	for x, item := range siblingItems {
		if item == ti {
			// Check to see if there is a previous one in the current list
			if x-1 >= 0 {
				newItem := siblingItems[x-1]
				// We want the next one "up". This could be the previous sibling, or if that sibling is open, it could be one of the siblings's children
				for {
					// Can this item have children?
					if newItem.CanHaveChildren && newItem.Open && len(newItem.Children) > 0 {
						// If so, we want to check again with the last one in this list. This could
						// be a many level tree, and we always want to check the last unopened child
						lastKid := len(newItem.Children) - 1
						// Descend into this child and iterate through checking this one
						newItem = newItem.Children[lastKid]
					} else {
						if atTop {
							tree.ScrollUp(1)
						}
						item.ParentTree.SetActive(newItem)
						return
					}

					//item.ParentTree.ActiveItem = siblingItems[x-1]
				}
			} else {
				// Nope, we were at the top. So now we just activate our Parent
				// Just make sure the parent is an item and not the tree. If it's the tree,
				// then we stop
				par, ok := ti.Parent.(*TreeItem)
				if ok {
					if atTop {
						tree.ScrollUp(1)
					}
					item.ParentTree.SetActive(par)
				}
			}
			return
		}
	}
}

// We're being told to select the next item relative to our current position.
func (ti *TreeItem) SelectNext() {
	// See if we need to scroll first:

	tree := ti.ParentTree

	atBottom := false
	windowBottom := tree.Height

	if tree.ActiveLine == windowBottom-1 {
		atBottom = true
	} else {
		tree.ActiveLine += 1
	}
	descend := true
	for {
		parentItems := ti.Parent.GetItems()

		// Check to see if we are going down into a child. If so, select the first child
		if len(ti.Children) > 0 && ti.Open && descend {
			ti.ParentTree.SetActive(ti.Children[0])
			if atBottom {
				tree.ScrollDown(1)
			}
			return
		}

		// If we're not descending into one of our own children, then we are going to the next
		// sibling item to us
		for x, item := range parentItems {
			// Find the current item
			if item == ti {
				if x+1 < len(parentItems) {
					// Select the sibling if there is one
					ti.ParentTree.SetActive(parentItems[x+1])
					if atBottom {
						tree.ScrollDown(1)
					}
					// If there is no sibling, then we are at the end and just do nothing
					return
				}
			}
		}
		if ti.Parent == ti.ParentTree {
			// We've hit the top
			return
		}
		// To get here, the user has tried to go past the end of the current list of children.
		// So we now need to tell our parent to choose a sibling. We turn off descent because
		// we aren't here going into an item, because we've already gone beyond the end of a list
		descend = false
		ti = ti.Parent.(*TreeItem)
	}
}

func (ti *TreeItem) ViewScrolled(viewtop, curline, bottomline int) (int, string) {
	// Return the view string for myself plus my children if I am open
	var pre_s string
	for x := 0; x < ti.indent; x++ {
		pre_s += "  "
	}
	if ti.CanHaveChildren {
		if ti.Open {
			pre_s += ChevronDown
		} else {
			pre_s += ChevronRight
		}
	} else {
		pre_s += NoChevron
	}

	ai := ti.ParentTree.ActiveItem

	var s string
	render := true
	if curline < 0 {
		render = false
	}

	if render {
		var baseline lipgloss.Style
		if ai != nil && ai == ti {
			baseline = focusedStyle
		} else {
			baseline = unfocusedStyle
		}
		istyle := baseline.Inherit(ti.IconStyle())
		lstyle := baseline.Inherit(ti.LabelStyle())
		s = pre_s + istyle.Render(s+ti.Icon()) + baseline.Render(" ") + lstyle.Render(ti.Name)
	}

	curline += 1

	if curline > bottomline {
		return curline, ""
	}

	if len(ti.Children) > 0 && ti.Open {
		var kids []string
		for _, item := range ti.Children {
			item.indent = ti.indent + 1
			var tmps string
			curline, tmps = item.ViewScrolled(viewtop, curline, bottomline)
			if strings.TrimSpace(tmps) != "" {
				kids = append(kids, tmps)
			}
			if curline >= bottomline {
				break
			}
		}
		var composite []string
		if s != "" {
			composite = append(composite, s)
		}

		composite = append(composite, kids...)
		s = lipgloss.JoinVertical(
			lipgloss.Left,
			composite...,
		)
	}
	return curline, s
}

func (ti *TreeItem) View() string {
	// Return the view string for myself plus my children if I am open
	var s string
	for x := 0; x < ti.indent; x++ {
		s += "  "
	}
	if ti.CanHaveChildren {
		if ti.Open {
			s += ChevronDown
		} else {
			s += ChevronRight
		}
	} else {
		s += NoChevron
	}

	ai := ti.ParentTree.ActiveItem

	if ai != nil && ai == ti {
		// If this is the active item, then we should be highlit
		s = focusedStyle.Render(s + ti.Icon() + " " + ti.Name)
	} else {
		//s += ti.Icon + " " + ti.Name
		s = unfocusedStyle.Render(s + ti.Icon() + " " + ti.Name)
	}

	if len(ti.Children) > 0 && ti.Open {
		var kids []string
		for _, item := range ti.Children {
			item.indent = ti.indent + 1
			inners := ""
			inners += item.View()
			kids = append(kids, inners)
		}
		composite := []string{s}
		composite = append(composite, kids...)
		s = lipgloss.JoinVertical(
			lipgloss.Left,
			composite...,
		)
	}
	return s
}

func (ti *TreeItem) GetPath() []string {
	var path []string
	if ti.Parent != nil {
		path = ti.Parent.GetPath()
	}
	return append(path, ti.Name)
}

func (ti *TreeItem) OpenChildren() {
	ti.Open = true
}

func (ti *TreeItem) CloseChildren() {
	ti.Open = false
}

func (ti *TreeItem) ToggleChildren() {
	if ti.CanHaveChildren {
		ti.Open = !ti.Open
		if ti.Open {
			if ti.OpenFunc != nil {
				ti.OpenFunc(ti)
			}
		} else {
			if ti.CloseFunc != nil {
				ti.CloseFunc(ti)
			}
		}
	}
}

// AddChildren - adds a list of children to an item, and then returns the item
// AddChild - adds a child item to the item. Adding a child will result in the automatic inclusion of
// the collapse chevron
func (ti *TreeItem) AddChildren(children ...*TreeItem) ItemHolder {
	// TODO: Should this do any mutex
	ti.Lock()
	ti.Children = append(ti.Children, children...)
	ti.Unlock()
	ti.CanHaveChildren = true // If it wasn't set before, it will be now

	for _, child := range children {
		child.Parent = ti
		child.ParentTree = ti.ParentTree
	}

	return ti
}

func NewItem(name string, canHaveChildren bool, children []*TreeItem, icon func(*TreeItem) string, labelStyle, iconStyle func(*TreeItem) lipgloss.Style, openFunc, closeFunc func(*TreeItem), data interface{}) *TreeItem {
	return &TreeItem{
		Name:            name,
		icon:            icon,
		labelStyle:      labelStyle,
		iconStyle:       iconStyle,
		Children:        children,
		Open:            false,
		OpenFunc:        openFunc,
		CloseFunc:       closeFunc,
		Data:            data,
		CanHaveChildren: canHaveChildren,
	}
}

type KeyMap struct {
	Space    key.Binding
	GoToTop  key.Binding
	GoToLast key.Binding
	Down     key.Binding
	Up       key.Binding
	PageUp   key.Binding
	PageDown key.Binding
	Back     key.Binding
	Open     key.Binding
	Select   key.Binding
}

type Tree struct {
	sync.Mutex
	viewtop              int // for scrolling
	Width                int
	Height               int
	ClosedChildrenSymbol string
	OpenChildrenSymbol   string
	ActiveItem           *TreeItem
	ActiveLine           int // Which line, (from 0..Height) is the cursor on?
	Items                []*TreeItem
	initialized          bool
	Style                lipgloss.Style
	KeyMap               KeyMap
}

func DefaultKeyMap() KeyMap {
	return KeyMap{
		Space:    key.NewBinding(key.WithKeys(" "), key.WithHelp(" ", "space")),
		GoToTop:  key.NewBinding(key.WithKeys("g"), key.WithHelp("g", "first")),
		GoToLast: key.NewBinding(key.WithKeys("G"), key.WithHelp("G", "last")),
		Down:     key.NewBinding(key.WithKeys("j", "down", "ctrl+n"), key.WithHelp("j", "down")),
		Up:       key.NewBinding(key.WithKeys("k", "up", "ctrl+p"), key.WithHelp("k", "up")),
		PageUp:   key.NewBinding(key.WithKeys("K", "pgup"), key.WithHelp("pgup", "page up")),
		PageDown: key.NewBinding(key.WithKeys("J", "pgdown"), key.WithHelp("pgdown", "page down")),
		Back:     key.NewBinding(key.WithKeys("h", "backspace", "left", "esc"), key.WithHelp("h", "back")),
		Open:     key.NewBinding(key.WithKeys("l", "right", "enter"), key.WithHelp("l", "open")),
		Select:   key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "select")),
	}
}

func New() tea.Model {
	t := Tree{
		OpenChildrenSymbol:   ChevronDown,
		ClosedChildrenSymbol: ChevronRight,
		KeyMap:               DefaultKeyMap(),
	}
	t.setInitialValues()
	return &t
}

func (t *Tree) setInitialValues() {
	t.initialized = true
}

// Returning nil here means you can't go "up" outside of the tree widget, so if this widget is embedded with others,
// this will prevent getting out of the tree.
func (t *Tree) GetParent() ItemHolder {
	return nil
}

func (t *Tree) AddChildren(i ...*TreeItem) ItemHolder {
	if len(i) == 0 {
		return t
	}
	t.Lock()
	t.Items = append(t.Items, i...)
	t.Unlock()
	// After we add the items, if we didn't have an active item, let's make it the first
	// one in the list
	if t.ActiveItem == nil {
		t.ActiveItem = t.Items[0]
	}
	for _, item := range i {
		item.Parent = t
		item.ParentTree = t
	}
	return t
}

func (t *Tree) GetItems() []*TreeItem {
	return t.Items
}

func (t *Tree) GetPath() []string {
	return []string{}
}

func (t *Tree) Init() tea.Cmd {
	return nil
}

// SelectPrevious - selects the previous TreeItem. This involves first getting the parent and then telling the parent to select the previous item from the current selection. If we're already at the first child, then we go to the grandparent and select the previous parent item from us, and then we descend to the most open child and activate that.rune
func (t *Tree) SelectPrevious() {
	active := t.ActiveItem
	if active != nil {
		active.SelectPrevious()
		return
	}
}

// SelectNext is like SelectPrevious, but the other way
func (t *Tree) SelectNext() {
	active := t.ActiveItem
	if active != nil {
		active.SelectNext()
		return
	}
}

// ToggleChild will toggle the open/closed state of the current selection. This only has meaning if there
// are actually children
func (t *Tree) ToggleChild() {
	if t.ActiveItem != nil {
		t.ActiveItem.ToggleChildren()
	}
}
func (t *Tree) Refresh() {
	t.Items = []*TreeItem{}
}

func (t *Tree) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		// TODO: Do I take into account margin & border?
		t.Width = msg.Width
		t.Height = msg.Height
		t.initialized = true

	// TODO: Convert these simple strings to a configurable keymap
	case tea.KeyMsg:
		switch msg.String() {
		case "?":
			log.Println("info")
		case "up", "k":
			t.SelectPrevious()
		case "down", "j":
			t.SelectNext()
		case " ", ".":
			t.ToggleChild()
			return t, nil
		}
	}

	var cmd tea.Cmd
	if t.ActiveItem != nil {
		var i tea.Model
		i, cmd = t.ActiveItem.Update(msg)
		t.ActiveItem = i.(*TreeItem)
	}
	return t, cmd
}

func (t *Tree) SetActive(ti *TreeItem) {
	t.ActiveItem = ti
}

// ScrollDown moves the "display" area down the virtual list. This actually looks like scrolling up ((the items move up the screen) Not sure if this is counterintuitive or not
func (t *Tree) ScrollDown(n int) {
	t.viewtop += n
}

func (t *Tree) ScrollUp(n int) {
	t.viewtop -= n
}

func (t *Tree) View() string {
	if !t.initialized {
		return ""
	}
	var views []string

	// Iterate through the children, calling View() on each of them.
	curline := -t.viewtop
	bottom := t.Height
	var v string
	for _, item := range t.Items {
		item.indent = 0
		curline, v = item.ViewScrolled(t.viewtop, curline, bottom)
		if curline > bottom {
			break
		}
		if v != "" {
			views = append(views, v)
		}
	}

	s := lipgloss.JoinVertical(
		lipgloss.Left, views...,
	)
	return s
}
