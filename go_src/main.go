package main

// #cgo CFLAGS: -I${SRCDIR}/../c_interop/include
// #cgo LDFLAGS: -L${SRCDIR}/../c_interop/build -l:libinterop.a -lstdc++
// #cgo LDFLAGS: -L${SRCDIR}/../deps/chafa/lib -l:libchafa.a -lm -lglib-2.0
// #cgo LDFLAGS: -pthread -lpcre2-8
// #include "init_draw_state_go.h"
// #include "draw_desktop_go.h"
import "C"

import (
	"fmt"
	"time"

	"go.wit.com/log"
	"go.wit.com/widget"
)

func main() {

	draw_state := C.init_draw_state_go(true)
	width := 256
	height := 256

	grid := make([]byte, width*height*4)
	// Fill a white square in the middle
	startX := 96
	startY := 96
	for y := startY; y < startY+64; y++ {
		for x := startX; x < startX+64; x++ {
			index := (y*width + x) * 4
			grid[index] = 255
			grid[index+1] = 255
			grid[index+2] = 255
			grid[index+3] = 255
		}
	}

	enable_alternative_screen_buffer := "\x1b[?1049h"
	disable_alternative_screen_buffer := "\x1b[?1049l"

	fmt.Print(enable_alternative_screen_buffer)
	defer fmt.Print(disable_alternative_screen_buffer)

	C.draw_desktop_go(draw_state, (*C.uchar)(C.CBytes(grid)), C.uint32_t(width), C.uint32_t(height), C.CString("Hello, world!"))

	time.Sleep(time.Second)

}

type TermEverything struct {
	initialized bool
	pluginChan  chan widget.Action
	frozenChan  chan widget.Action
	callback    chan widget.Action
}

func New() *TermEverything {
	me := new(TermEverything)
	me.pluginChan = make(chan widget.Action, 1)
	me.frozenChan = make(chan widget.Action, 1)

	go me.catchActionChannel()

	log.Log(TERM_EVERYTHING, "Init() start channel reciever")
	go me.catchActionChannel()
	log.Log(TERM_EVERYTHING, "Init() END")
	return me
}

func (me *TermEverything) catchActionChannel() {
	for {
		action := <-me.pluginChan
		log.Log(TERM_EVERYTHING, "Plugin got action: ", action)

		me.doAction(action)
	}
}

func (me *TermEverything) Callback(guiCallback chan widget.Action) {
	me.callback = guiCallback
}

// this is the function that receives things from the application
func (me *TermEverything) PluginChannel() chan widget.Action {
	return me.pluginChan
}

// this is the function that receives things from the application
func (me *TermEverything) FrozenChannel() chan widget.Action {
	return me.frozenChan
}
