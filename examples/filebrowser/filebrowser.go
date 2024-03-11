package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/greenenergy/greenbubbles/filebrowser"
)

func main() {
	var debug = flag.Bool("d", false, "create debug log")
	flag.Parse()

	if len(flag.Args()) < 1 {
		fmt.Println("usage: filebrowser <foldername>")
		return
	}
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

	dir := flag.Arg(0)
	m := filebrowser.New(dir)
	p := tea.NewProgram(m)
	if _, err := p.Run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Println(m.Value())
}
