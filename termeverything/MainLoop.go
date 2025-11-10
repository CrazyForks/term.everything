package termeverything

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

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
		terminalWindow.FrameEvents,
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
			go client.MainLoop()
		}
	}()

	if len(args.Positionals) > 0 {
		cmdStr := strings.Join(args.Positionals, " ")
		shell := args.Shell
		cmd := exec.Command(shell, "-c", cmdStr)

		baseEnv := os.Environ()
		filtered := make([]string, 0, len(baseEnv))
		for _, e := range baseEnv {
			if strings.HasPrefix(e, "DISPLAY=") {
				continue
			}

			if !args.SupportOldApps && strings.HasPrefix(e, "XDG_SESSION_TYPE=") {
				continue
			}
			filtered = append(filtered, e)
		}
		filtered = append(filtered, fmt.Sprintf("WAYLAND_DISPLAY=%s", listener.WaylandDisplayName))
		if !args.SupportOldApps {
			filtered = append(filtered, "XDG_SESSION_TYPE=wayland")
		}

		cmd.Env = filtered
		// cmd.Stdout = os.Stdout
		// cmd.Stderr = os.Stderr
		// cmd.Stdin = os.Stdin

		if err := cmd.Start(); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to start command: %v\n", err)
		} else {
			go func() {
				_ = cmd.Wait()
			}()
		}
	}

	<-done

	//TODO start xwaylnd_if_neccessary

	// // Wait for SigInt, TODO something different
	// sig := make(chan os.Signal, 1)
	// signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	// <-sig
	// _ = listener.Close()
	// fmt.Println("Shutdown complete")
}
