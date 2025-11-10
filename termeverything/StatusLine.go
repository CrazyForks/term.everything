package termeverything

import (
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"strings"

	"github.com/mmulet/term.everything/escapecodes"
	"github.com/mmulet/term.everything/framebuffertoansi"
)

type LineButton struct {
	String   string
	Callback func()
	Keycode  *Linux_Event_Codes
}

type StatusLineTextOrButton interface {
	IsStatusLineTextOrButton()
}

type StatusLineText struct {
	String string
}

func (s *StatusLineText) IsStatusLineTextOrButton() {}

type StatusLineButton struct {
	Button LineButton
}

func (s *StatusLineButton) IsStatusLineTextOrButton() {}

type Status_Line struct {
	TextLoopTime float64

	ShowStatusLine bool

	TerminalMousePosition struct {
		x int
		y int
	}

	TerminalMouseButton struct {
		pressed         bool
		frame_held_time float64
	}

	b       map[string]*StatusLineButton
	Sponsor *StatusLineButton
	Bugs    *StatusLineButton
}

func (s *Status_Line) UpdateMousePosition(code *PointerMove) {
	if code == nil {
		return
	}
	s.TerminalMousePosition.x = code.Col
	s.TerminalMousePosition.y = code.Row
}

func (s *Status_Line) HandleTerminalMousePress(pressed bool) {
	if pressed {
		if s.TerminalMouseButton.pressed {
			/**
			 * Mouse state has not changed
			 * do nothing
			 */
			return
		}
		s.TerminalMouseButton.pressed = true
		s.TerminalMouseButton.frame_held_time = 0
		return
	}
	s.TerminalMouseButton.pressed = false
	s.TerminalMouseButton.frame_held_time = 0
}

func (s *Status_Line) PostFrame(delta_time float64) {
	if s.TerminalMouseButton.pressed {
		s.TerminalMouseButton.frame_held_time += delta_time
	}
}

func MakeStatusLine() *Status_Line {
	sl := &Status_Line{
		ShowStatusLine: true,
	}
	sl.TerminalMousePosition.x = -1
	sl.TerminalMousePosition.y = -1

	escape := KEY_ESC
	sl.b = map[string]*StatusLineButton{
		"escape": &StatusLineButton{
			Button: LineButton{
				Keycode: &escape,
				String:  "[ESC] to quit",
				Callback: func() {
					GlobalExitChan <- 0
				},
			},
		},
		"left": &StatusLineButton{
			Button: LineButton{
				String: "[]",
				Callback: func() {
					fmt.Println("left")
				},
			},
		},
	}

	sl.Sponsor = &StatusLineButton{
		Button: LineButton{
			String: "[Sponsor this project]",
			Callback: func() {
				_ = exec.Command("xdg-open", "https://github.com/sponsors/mmulet").Start()
			},
		},
	}

	sl.Bugs = &StatusLineButton{
		Button: LineButton{
			String: "[Report bugs here]",
			Callback: func() {
				title := url.QueryEscape("Bug Report")
				body := url.QueryEscape(sl.buildBugBody())
				_ = exec.Command("xdg-open",
					"https://github.com/mmulet/term.everything/issues/new?title="+title+"&body="+body).Start()
			},
		},
	}

	return sl
}

func (s *Status_Line) Draw(delta_time float64, app_title *string, keys_pressed_this_frame map[Linux_Event_Codes]bool) string {
	if !s.ShowStatusLine {
		return ""
	}

	text := s.Line(keys_pressed_this_frame,
		s.b["escape"], &StatusLineText{" "},
		s.Sponsor, &StatusLineText{" | "},
		s.ChooseAppTitle(app_title), &StatusLineText{" | "},
	)

	s.TextLoopTime += delta_time

	width := 0
	if winsize, err := framebuffertoansi.GetWinsize(os.Stdout.Fd()); err == nil {
		width = int(winsize.Col)
	}
	if width > 1 && len(text) >= width {
		return text[:width-1]
	}
	return text
}

func (s *Status_Line) buildBugBody() string {
	return fmt.Sprintf(`
Quick question before you fill this out:
  Is your app opening a new window instead of opening in the terminal?
    If so, do you have any other windows of the current app open?
    For example, firefox likes to open a new window (not in the terminal)
    if you already have at least one firefox window open.
    Close all other windows, and see if the problem still happens.

## Describe the bug
A clear and concise description of what the bug is.

## To Reproduce
Steps to reproduce the behavior:
1. Go to '...'
2. Click on '....'
3. Scroll down to '....'
4. See error

## Expected behavior
A clear and concise description of what you expected to happen.

## Screenshots
If applicable, add screenshots to help explain your problem.

## Additional context
Add any other context about the problem here.
        

## System Information
Generated from your system, please include this information in your report:
- Platform: %s
- Architecture: %s
- Terminal: %s
- OS: %s
- OS Details: %s
- XDG_SESSION_TYPE: %s
- Wayland Display: %s
- X11 Display: %s
- term.everything version: %s
        `,
		os.Getenv("GOOS"),
		os.Getenv("GOARCH"),
		getEnvOr("TERM", "N/A"),
		os.Getenv("GOOS"),
		s.GetOsDetails(),
		getEnvOr("XDG_SESSION_TYPE", "N/A"),
		getEnvOr("WAYLAND_DISPLAY", "N/A"),
		getEnvOr("DISPLAY", "N/A"),
		version,
	)
}

func getEnvOr(k, def string) string {
	v := os.Getenv(k)
	if v == "" {
		return def
	}
	return v
}

func (s *Status_Line) ChooseAppTitle(appTitle *string) StatusLineTextOrButton {
	if appTitle == nil || *appTitle == "" {
		return s.Bugs
	}
	return &StatusLineText{*appTitle}
}

func (s *Status_Line) KeyboardKeyHitButton(button LineButton, keys_pressed_this_frame map[Linux_Event_Codes]bool) LineButton {
	if button.Keycode == nil {
		return button
	}
	if _, ok := keys_pressed_this_frame[*button.Keycode]; ok {
		button.Callback()
		// Replace callback with no-op
		button.Callback = func() {}
		return button
	}
	return button
}

func (s *Status_Line) Line(keys_pressed_this_frame map[Linux_Event_Codes]bool, parts ...StatusLineTextOrButton) string {
	position := 0
	var out strings.Builder

	for _, v := range parts {
		switch it := v.(type) {
		case *StatusLineText:
			out.WriteString(it.String)
			position += len(it.String)
		case *StatusLineButton:
			btn := s.KeyboardKeyHitButton(it.Button, keys_pressed_this_frame)
			nextString := btn.String
			/**
			 * for the rare case where
			 * both click on button and
			 * hold key at the same time
			 */
			already_called_callback := false
			if s.TerminalMousePosition.y == 0 &&
				int(s.TerminalMousePosition.x) >= position &&
				int(s.TerminalMousePosition.x) < position+len(nextString) {
				out.WriteString(escapecodes.BgWhite + escapecodes.FgBlack + nextString + escapecodes.Reset)
				if s.TerminalMouseButton.pressed &&
					s.TerminalMouseButton.frame_held_time == 0 {
					if !already_called_callback {
						btn.Callback()
					}
				}
			} else {
				out.WriteString(nextString)
			}
			position += len(nextString)
		}
	}
	return out.String()
}

func (s *Status_Line) GetOsDetails() string {
	data, err := os.ReadFile("/etc/os-release")
	if err != nil {
		return "Unable to determine OS details"
	}
	lines := strings.Split(string(data), "\n")
	osInfo := make(map[string]string)
	for _, line := range lines {
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := parts[0]
		value := strings.Trim(parts[1], `"`)
		osInfo[key] = value
	}
	return fmt.Sprintf("%s (ID: %s, VERSION: %s)",
		firstOr(osInfo["PRETTY_NAME"], "Unknown"),
		firstOr(osInfo["ID"], "N/A"),
		firstOr(osInfo["VERSION"], "N/A"),
	)
}

func firstOr(a, b string) string {
	if a == "" {
		return b
	}
	return a
}
