package termeverything

import (
	"strconv"
	"time"

	"github.com/mmulet/term.everything/framebuffertoansi"
	"github.com/mmulet/term.everything/wayland"
	"github.com/mmulet/term.everything/wayland/protocols"
)

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
	frameEvents chan XkbdCode,
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
		FrameEvents:             frameEvents,
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
