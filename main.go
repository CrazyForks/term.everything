package main

import (
	"github.com/mmulet/term.everything/termeverything"
)

//go:generate go generate ./wayland

func main() {
	termeverything.MainLoop()
}
