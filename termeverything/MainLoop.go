package termeverything

import (
	"fmt"
	"os"

	"github.com/mmulet/term.everything/wayland"
)

func MainLoop() {
	args := ParseArgs()
	SetVirtualMonitorSize(args.VirtualMonitorSize)
	listener, err := wayland.MakeSocketListener(&args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create socket listener: %v\n", err)
		os.Exit(1)
	}

	displaySize := wayland.Size{
		Width:  uint32(wayland.VirtualMonitorSize.Width),
		Height: uint32(wayland.VirtualMonitorSize.Height),
	}

	terminalWindow := MakeTerminalWindow(listener,
		displaySize,
		&args,
	)

	terminanDrawLoop := MakeTerminalDrawLoop(
		displaySize,
		args.HideStatusBar,
		len(args.Positionals) > 0,
		terminalWindow.SharedRenderedScreenSize,
		&args,
	)

	go listener.MainLoopThenClose()
	go terminalWindow.InputLoop()
	go terminanDrawLoop.MainLoop()

	done := make(chan struct{})
	go func() {
		for {
			conn := <-listener.OnConnection
			client := wayland.MakeClient(conn)
			terminalWindow.GetClients <- client
			terminanDrawLoop.GetClients <- client
		}
	}()
	<-done

	//TODO start xwaylnd_if_neccessary

	// // Wait for SigInt, TODO something different
	// sig := make(chan os.Signal, 1)
	// signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	// <-sig
	// _ = listener.Close()
	// fmt.Println("Shutdown complete")
}
