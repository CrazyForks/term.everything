package termeverything

import (
	"fmt"
	"os"
	"os/signal"
	"strconv"
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
}

func MakeTerminalWindow(
	socket_listener *wayland.SocketListener,
	desktop_size wayland.Size,
	args *CommandLineArgs,

) *TerminalWindow {
	tw := &TerminalWindow{
		SocketListener:           socket_listener,
		VirtualMonitorSize:       desktop_size,
		Mode:                     WindowMode_Passthrough,
		FrameEvents:              make(chan XkbdCode, 8192),
		Args:                     args,
		KeySerial:                0,
		PressedMouseButton:       nil,
		SharedRenderedScreenSize: &RenderedScreenSize{},
	}

	os.Stdout.WriteString(escapecodes.EnableAlternativeScreenBuffer)
	// TODO turn this on, I might be missing the mouse up events without it
	// os.Stdout.WriteString(escapecodes.EnableNormalMouseTracking)
	os.Stdout.WriteString(escapecodes.EnableMouseTracking)

	os.Stdout.WriteString(escapecodes.HideCursor)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh,
		syscall.SIGINT,
		syscall.SIGQUIT,
		syscall.SIGTERM,
		syscall.SIGUSR1,
		syscall.SIGUSR2,
	)
	go func() {
		for range sigCh {
			tw.OnExit()
			os.Exit(0)
		}
	}()

	return tw
}

func (tw *TerminalWindow) OnExit() {
	for _, s := range tw.Clients {
		for surface := range s.TopLevelSurfaces() {
			protocols.XdgToplevel_close(s, surface)
		}
	}

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

		// Very literal: TS line -> const codes = convert_keycode_to_xbd_code(chunk);
		codes := ConvertKeycodeToXbdCode(chunk)

		now := uint32(time.Now().UnixMilli())

		for {
			select {
			case client := <-tw.GetClients:
				//TODO removing clients
				tw.Clients = append(tw.Clients, client)
			default:
				goto DoneGettingClients
			}
		}
	DoneGettingClients:

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
							if release != nil {
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
				if tw.PressedMouseButton == nil {
					break
				}
				buttonToRelease := *tw.PressedMouseButton
				tw.PressedMouseButton = nil

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

type TerminalDrawLoop struct {
	VirtualMonitorSize wayland.Size

	Clients []*wayland.Client

	TimeOfLastTerminalDraw *float64

	HideStatusBar bool

	/**
	 * Don't draw until at least MinTerminalTimeSeconds has passed
	 * since the last frame has been drawn to the terminal. (Not drawn
	 * to the canvas, that is done as fast as possible)
	 *
	 * This is set from the --max-frame-rate argument.
	 */
	MinTerminalTimeSeconds *float64

	DrawState *framebuffertoansi.DrawState

	Desktop *Desktop

	SharedRenderedScreenSize *RenderedScreenSize

	FrameEvents chan XkbdCode

	TimeOfStartOfLastFrame *float64

	DesiredFrameTimeSeconds float64

	StatusLine *Status_Line

	GetClients chan *wayland.Client
}

func MakeTerminalDrawLoop(desktop_size wayland.Size,
	hide_status_bar bool,
	willShowAppRightAtStartup bool,
	sharedRenderedScreenSize *RenderedScreenSize,
	args *CommandLineArgs,

) *TerminalDrawLoop {
	tw := &TerminalDrawLoop{
		Clients:                  make([]*wayland.Client, 0),
		TimeOfLastTerminalDraw:   nil,
		MinTerminalTimeSeconds:   nil,
		SharedRenderedScreenSize: sharedRenderedScreenSize,
		HideStatusBar:            hide_status_bar,
		DrawState: framebuffertoansi.MakeDrawState(
			DisplayServerType() == DisplayServerTypeX11,
		),
		VirtualMonitorSize: desktop_size,

		Desktop: MakeDesktop(wayland.Size{
			Width:  desktop_size.Width,
			Height: desktop_size.Height,
		}, willShowAppRightAtStartup),

		TimeOfStartOfLastFrame:  nil,
		DesiredFrameTimeSeconds: 0.016, // ~60 FPS
		StatusLine:              MakeStatusLine(),
	}
	if args != nil && args.MaxFrameRate != "" {
		if fps, err := strconv.ParseFloat(args.MaxFrameRate, 64); err == nil && fps > 0 {
			v := 1.0 / fps
			tw.MinTerminalTimeSeconds = &v
		}
	}

	return tw
}

func (tw *TerminalDrawLoop) GetAppTitle() *string {
	for _, s := range tw.Clients {
		for topLevelID := range s.TopLevelSurfaces() {
			top_level := wayland.GetXdgToplevelObject(s, topLevelID)
			if top_level == nil {
				continue
			}
			return top_level.Title
		}
	}
	return nil
}

func (tw *TerminalDrawLoop) DrawToTerminal(start_of_frame float64, status_line string) {
	if tw.MinTerminalTimeSeconds != nil {
		last := 0.0
		if tw.TimeOfLastTerminalDraw != nil {
			last = *tw.TimeOfLastTerminalDraw
		}
		if start_of_frame-last < *tw.MinTerminalTimeSeconds {
			return
		}
		tw.TimeOfLastTerminalDraw = &start_of_frame
	}

	var statusLine *string
	if !tw.HideStatusBar {
		statusLine = &status_line
	}

	widthCells, heightCells := tw.DrawState.DrawDesktop(
		tw.Desktop.Buffer,
		tw.VirtualMonitorSize.Width,
		tw.VirtualMonitorSize.Height,
		statusLine,
	)
	tw.SharedRenderedScreenSize.WidthCells = &widthCells
	tw.SharedRenderedScreenSize.HeightCells = &heightCells

}

func (tw *TerminalDrawLoop) MainLoop() {
	keys_pressed_this_frame := map[Linux_Event_Codes]bool{}
	for {
		start_of_frame := float64(time.Now().UnixMilli()) / 1000.0
		var delta_time float64
		if tw.TimeOfStartOfLastFrame != nil {
			delta_time = start_of_frame - *tw.TimeOfStartOfLastFrame
		} else {
			delta_time = tw.DesiredFrameTimeSeconds
		}

		for _, s := range tw.Clients {
			for {
				select {
				case callback_id := <-s.FrameDrawRequests:
					protocols.WlCallback_done(s, callback_id, uint32(time.Now().UnixMilli()))
				default:
					goto DoneCallbacks
				}
			}
		DoneCallbacks:
		}

		for _, s := range tw.Clients {
			pointer_surface_id := wayland.Pointer.PointerSurfaceID[s]
			if pointer_surface_id == nil {
				continue
			}
			surface := wayland.GetWlSurfaceObject(s, *pointer_surface_id)
			if surface == nil {
				continue
			}
			surface.Position.X = int32(wayland.Pointer.WindowX)
			surface.Position.Y = int32(wayland.Pointer.WindowY)
			surface.Position.Z = 1000

		}

		tw.Desktop.DrawClients(tw.Clients)

		status_line := tw.StatusLine.Draw(delta_time, tw.GetAppTitle(), keys_pressed_this_frame)

		tw.DrawToTerminal(start_of_frame, status_line)

		// const draw_time = Date.now();

		// const time_until_next_frame = Math.max(
		//   0,
		//   this.desired_frame_time_seconds - (draw_time - start_of_frame)
		// );

		tw.TimeOfStartOfLastFrame = &start_of_frame

		tw.StatusLine.PostFrame(delta_time)

		clear(keys_pressed_this_frame)

		timeout := time.After(time.Duration(tw.DesiredFrameTimeSeconds * float64(time.Second)))

		for {
			select {
			case code := <-tw.FrameEvents:
				switch c := code.(type) {
				case *KeyCode:
					keys_pressed_this_frame[c.KeyCode] = true
				case *PointerMove:
					tw.StatusLine.UpdateMousePosition(c)
				case *PointerButtonPress:
					tw.StatusLine.HandleTerminalMousePress(true)
				case *PointerButtonRelease:
					tw.StatusLine.HandleTerminalMousePress(false)
				case *PointerWheel:
				}
			case client := <-tw.GetClients:
				//TODO removing clients
				tw.Clients = append(tw.Clients, client)
			case <-timeout:
				goto KeyReadLoop
			}
		}
	KeyReadLoop:
		// /**
		//  * I know sleep is bad for timing.
		//  * @TODO replace with polling later on.
		//  */
		// time.Sleep(time.Duration(tw.DesiredFrameTimeSeconds * float64(time.Second)))
	}
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
