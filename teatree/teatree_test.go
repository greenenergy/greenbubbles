package teatree

import (
	"fmt"
	"testing"

	"github.com/charmbracelet/lipgloss"
)

/*
NerdFont Cheatsheet: https://www.nerdfonts.com/cheat-sheet

Important: If the codepoint has 4 digits, it can be prefixed with "\u" and only be 4 characters long.
But if it has 5 digits, then you have to use "\U000" as a prefix
*/

const ServerPlus = "\U000F0490"
const ServerPlusOutline = "\U000F1C9B"
const ServerNetworkOutline = "\U000F1C99"
const ServerNetwork = "\U000F0480"
const ServerNetworkOff = "\U000F048E"
const ServerOff = "\U000F048F"
const Server = "\U000F048B"
const SecureServer = "\U000F0492"
const FileCloud = "\U000F0217"
const Penguin = "\uebc6"
const Database = "\U000f01bc"
const User = "\uf007"
const UserGroup = "\U000f0849"
const Organization = "\uf42b"

const Folder = "\U000F024B"
const File = "\U000F0214"
const GoGopherDev = "\ue626"
const GoGopher = "\ue724"
const GoTitle = "\U000F07D3"

func TestSymbols(t *testing.T) {
	m := New()

	m.Update(nil)
	redstyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0000"))
	greenstyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF00"))
	bluestyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#0000FF"))

	/*
		Backblue := lipgloss.NewStyle().
			Background(lipgloss.Color("#00007F")).
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("63")).
			Padding(1).
			Margin(1, 2, 3, 4).
			Inherit(bluestyle)
	*/

	//fmt.Println(m.View())
	fmt.Println(bluestyle.Render(Penguin), "Penguin")
	fmt.Println(greenstyle.Render(ServerPlus), "ServerPlus")
	fmt.Println(bluestyle.Render(Server), "Server")
	fmt.Println(redstyle.Render(SecureServer), "SecureServer")
	fmt.Println(FileCloud, "FileCloud")
	fmt.Println(Database, "Database")
	fmt.Println(User, "User")
	fmt.Println(UserGroup, "UserGroup")
	fmt.Println(GoGopherDev, "GoGopherDev")
	fmt.Println(GoGopher, "GoGopher")
	fmt.Println(GoTitle, "GoTitle")
	fmt.Println(redstyle.Render(Organization), "Organization")
}

/*
func TestTree(t *testing.T) {
	m := New()
	tr := m.(*Tree)

	root := NewItem("", Folder, true, nil, nil, nil, nil)
	home := NewItem("home", Folder, true, nil, nil, nil, nil)
	cfox := NewItem("cfox", Folder, true, nil, nil, nil, nil)
	work := NewItem("work", Folder, true, nil, nil, nil, nil)
	testproj := NewItem("testproj", Folder, true, nil, nil, nil, nil)
	mgrthing := NewItem("mgrthing", Folder, true, nil, nil, nil, nil)
	game1 := NewItem("game1", File, false, nil, nil, nil, nil)

	newitem := NewItem(d.Name(), canHaveChildren, children, icon, labelStyle, iconStyle, openFunc, nil, nil)

	tr.AddChildren(root)
	root.AddChildren(home)
	home.AddChildren(cfox)
	cfox.AddChildren(work)
	work.AddChildren(testproj, mgrthing)
	testproj.AddChildren(game1)

	fmt.Println("child path:", strings.Join(cfox.GetPath(), "/"))
	fmt.Println("child path:", strings.Join(game1.GetPath(), "/"))
}
*/

var _ = `
Hierarchy test
Here's the tree I am representing:


󰅀 Item 1	-7
	A		-6
	B		-5
	C		-4
󰅀 Item 2	-3
	AA		-2
	BB		-1
--------------------
	CC
󰅂 Item 3
󰅀 Item 4
	A4
	B4
	C4

So with a "topscroll" = -7, rendering starts with Item 1. It is open, as is its sibling Item2

Renders Item 1. The current render line is -7 so we don't actually render, but we do a "lipglosss.JoinVertical"

`
