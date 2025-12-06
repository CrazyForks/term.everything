package termeverything

import (
	"fmt"
	"os"
	"os/signal"
	"slices"
	"syscall"
	"time"

	"github.com/mmulet/term.everything/escapecodes"
	"github.com/mmulet/term.everything/framebuffertoansi"
	"github.com/mmulet/term.everything/wayland"
	"github.com/mmulet/term.everything/wayland/protocols"
)

type RenderedScreenSize struct {
	WidthCells  *int
	HeightCells *int
}

type WindowMode int

const (
	WindowMode_Passthrough WindowMode = iota
	WindowMode_Capture
)

var GlobalExitChan = make(chan int)

type TerminalWindow struct {
	SocketListener     *wayland.SocketListener
	VirtualMonitorSize wayland.Size

	Mode WindowMode

	FrameEvents chan XkbdCode

	Args *CommandLineArgs

	KeySerial uint32

	PressedMouseButton *LINUX_BUTTON_CODES

	Clients []*wayland.Client

	GetClients chan *wayland.Client

	SharedRenderedScreenSize *RenderedScreenSize

	RestoreTerminalMode func() error
}

func MakeTerminalWindow(
	socket_listener *wayland.SocketListener,
	desktop_size wayland.Size,
	args *CommandLineArgs,

) *TerminalWindow {

	restoreTerminalMode, err := EnableRawModeFD(int(os.Stdin.Fd()))
	if err != nil {
		panic(err)
	}

	tw := &TerminalWindow{
		SocketListener:           socket_listener,
		VirtualMonitorSize:       desktop_size,
		Mode:                     WindowMode_Passthrough,
		FrameEvents:              make(chan XkbdCode, 8192),
		Args:                     args,
		KeySerial:                0,
		PressedMouseButton:       nil,
		SharedRenderedScreenSize: &RenderedScreenSize{},
		Clients:                  make([]*wayland.Client, 0),
		// RestoreTerminalMode:      func() error { return nil },
		RestoreTerminalMode: restoreTerminalMode,
		GetClients:          make(chan *wayland.Client, 32),
	}

	if !protocols.DebugRequests {
		os.Stdout.WriteString(escapecodes.EnableAlternativeScreenBuffer)
		os.Stdout.WriteString(escapecodes.EnableMouseTracking)
		os.Stdout.WriteString(escapecodes.EnableSGR)

		os.Stdout.WriteString(escapecodes.HideCursor)
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh,
		syscall.SIGINT,
		syscall.SIGQUIT,
		syscall.SIGTERM,
		syscall.SIGUSR1,
		syscall.SIGUSR2,
	)
	go func() {
		exit_code := 0
		select {
		case exit_code = <-GlobalExitChan:
		case <-sigCh:
		}
		tw.OnExit()
		os.Exit(exit_code)
	}()

	return tw
}

func (tw *TerminalWindow) OnExit() {
	for _, s := range tw.Clients {
		for surface := range s.TopLevelSurfaces() {
			protocols.XdgToplevel_close(s, surface)
		}
	}
	tw.RestoreTerminalMode()

	os.Stdout.WriteString(escapecodes.DisableAlternativeScreenBuffer)
	os.Stdout.WriteString(escapecodes.ShowCursor)

	// TODO re-enable if enabled above
	// os.Stdout.WriteString(escapecodes.DisableNormalMouseTracking)
	os.Stdout.WriteString(escapecodes.DisableMouseTracking)

}

func (tw *TerminalWindow) InputLoop() {
	buf := make([]byte, 4096)
	for {

		n, err := os.Stdin.Read(buf)

		if err != nil || n == 0 {
			fmt.Printf("Error reading stdin: %v\n", err)
			return
		}
		chunk := buf[:n]
		for {
			select {
			case client := <-tw.GetClients:
				//TODO removing client
				tw.Clients = append(tw.Clients, client)
			default:
				goto GotData
			}
		}
	GotData:
		codes := ConvertKeycodeToXbdCode(chunk)
		tw.ProcessCodes(codes)
	}
}

func (tw *TerminalWindow) ProcessCodes(codes []XkbdCode) {
	clients_to_delete := make([]int, 0)
	for i, s := range tw.Clients {
		s.Access.Lock()
		if s.Status != wayland.ClientStatus_Connected {
			s.Access.Unlock()
			clients_to_delete = append(clients_to_delete, i)
			continue
		} else {
			defer s.Access.Unlock()
		}
	}
	for i := len(clients_to_delete) - 1; i >= 0; i-- {
		index := clients_to_delete[i]
		tw.Clients = slices.Delete(tw.Clients, index, index+1)
	}
	now := uint32(time.Now().UnixMilli())

	for _, code := range codes {
		tw.FrameEvents <- code
		new_key_serial := tw.KeySerial
		tw.KeySerial += 2

		for _, s := range tw.Clients {
			if keyboard_map := protocols.GetGlobalWlKeyboardBinds(s); keyboard_map != nil {
				modifiers := code.GetModifiers()
				for keyboardID := range keyboard_map {
					protocols.WlKeyboard_modifiers(
						s,
						keyboardID,
						new_key_serial,
						uint32(modifiers),
						0, 0, 0,
					)
				}
			}
		}
		switch c := code.(type) {
		case *KeyCode:
			for _, s := range tw.Clients {
				if keyboard_map := protocols.GetGlobalWlKeyboardBinds(s); keyboard_map != nil {
					for keyboardID := range keyboard_map {
						protocols.WlKeyboard_key(
							s,
							keyboardID,
							new_key_serial,
							now,
							uint32(c.KeyCode),
							protocols.WlKeyboardKeyState_enum_pressed,
						)
						/**
						 * There is no key up code in
						 * ANSI escape codes, so
						 * just say it is released
						 * instantly
						 */
						protocols.WlKeyboard_key(
							s,
							keyboardID,
							new_key_serial+1,
							now,
							uint32(c.KeyCode),
							protocols.WlKeyboardKeyState_enum_released,
						)
					}
				}
			}

		case *PointerMove:
			cols, rows := tw.CurrentTerminalSize()
			x := float32(c.Col) *
				(float32(tw.VirtualMonitorSize.Width) /
					float32(cols))
			y := float32(c.Row) *
				(float32(tw.VirtualMonitorSize.Height) /
					float32(rows))

			wayland.Pointer.WindowX = x
			wayland.Pointer.WindowY = y

			for _, s := range tw.Clients {
				if pointers_map := protocols.GetGlobalWlPointerBinds(s); pointers_map != nil {
					for pointerID, version := range pointers_map {
						protocols.WlPointer_motion(
							s,
							pointerID,
							uint32(time.Now().UnixMilli()),
							x,
							y,
						)
						protocols.WlPointer_frame(
							s,
							uint32(version),
							pointerID,
						)
					}
				}
			}

		case *PointerButtonPress:

			release := tw.GetButtonToReleaseAndUpdatePressedMouseButton(c.Button)
			for _, s := range tw.Clients {

				if pointer_map := protocols.GetGlobalWlPointerBinds(s); pointer_map != nil {
					for pointerID, version := range pointer_map {
						protocols.WlPointer_button(
							s,
							pointerID,
							uint32(time.Now().UnixMilli()),
							uint32(time.Now().UnixMilli()),
							uint32(c.Button),
							protocols.WlPointerButtonState_enum_pressed,
						)
						protocols.WlPointer_frame(
							s,
							uint32(version),
							pointerID,
						)
						if c.NeedToReleaseOtherButtons && release != nil {
							protocols.WlPointer_button(
								s,
								pointerID,
								uint32(time.Now().UnixMilli()),
								uint32(time.Now().UnixMilli()),
								uint32(*release),
								protocols.WlPointerButtonState_enum_released,
							)
							protocols.WlPointer_frame(
								s,
								uint32(version),
								pointerID,
							)
						}
					}
				}
			}

		case *PointerButtonRelease:
			buttonToRelease := c.Button
			if c.NeedsButtonGuessing {
				if tw.PressedMouseButton == nil {
					break
				}
				buttonToRelease = *tw.PressedMouseButton
				tw.PressedMouseButton = nil
			}

			for _, s := range tw.Clients {

				if pointer_map := protocols.GetGlobalWlPointerBinds(s); pointer_map != nil {
					for pointerID, version := range pointer_map {
						protocols.WlPointer_button(
							s,
							pointerID,
							uint32(time.Now().UnixMilli()),
							uint32(time.Now().UnixMilli()),
							uint32(buttonToRelease),
							protocols.WlPointerButtonState_enum_released,
						)
						protocols.WlPointer_frame(
							s,
							uint32(version),
							pointerID,
						)
					}
				}
			}

		case *PointerWheel:
			_, rows := tw.CurrentTerminalSize()

			var scale float32 = 0.5
			if (c.Modifiers & ModAlt) != 0 {
				scale = 1
			}
			amount := scale * float32(tw.ScrollDirection(c.Up)) * float32(tw.VirtualMonitorSize.Height) / float32(rows)
			for _, s := range tw.Clients {
				if pointer_id := protocols.GetGlobalWlPointerBinds(s); pointer_id != nil {
					for pointerID, version := range pointer_id {
						protocols.WlPointer_axis(
							s,
							pointerID,
							uint32(time.Now().UnixMilli()),
							protocols.WlPointerAxis_enum_vertical_scroll,
							amount,
						)
						protocols.WlPointer_frame(
							s,
							uint32(version),
							pointerID,
						)
					}
				}
			}
		default:
			// literal never_default(code) equivalent: do nothing
		}
	}
}

func (tw *TerminalWindow) ScrollDirection(code_up bool) float32 {
	var code float32 = 1.0
	if code_up {
		code = -1.0
	}
	var reverse float32 = 1.0
	if tw.Args != nil && tw.Args.ReverseScroll {
		reverse = -1.0
	}
	return code * reverse
}

/**
 * Because we only get release updates for one button at a time
 * assume that when you press another mouse button you will
 * release the one you already have pressed.
 */
func (tw *TerminalWindow) GetButtonToReleaseAndUpdatePressedMouseButton(new_pressed_button LINUX_BUTTON_CODES) *LINUX_BUTTON_CODES {
	old_pressed_mouse_button := tw.PressedMouseButton
	tw.PressedMouseButton = &new_pressed_button
	//TODO I think this a bug, but keeping it for now because I dont
	// want to make any behavior changes while porting
	if old_pressed_mouse_button == nil || *tw.PressedMouseButton == new_pressed_button {
		return nil
	}
	return old_pressed_mouse_button
}

func (tw *TerminalWindow) CurrentTerminalSize() (cols, rows int) {
	if tw.SharedRenderedScreenSize != nil && tw.SharedRenderedScreenSize.WidthCells != nil && tw.SharedRenderedScreenSize.HeightCells != nil {
		return *tw.SharedRenderedScreenSize.WidthCells, *tw.SharedRenderedScreenSize.HeightCells
	}
	ws, err := framebuffertoansi.GetWinsize(1)
	if err != nil || ws.Col <= 0 || ws.Row <= 0 {
		return 80, 24
	}
	return int(ws.Col), int(ws.Row)
}
