package termeverything

import (
	"strconv"
	"strings"
)

type PointerEvent interface {
	isPointerEvent()
	isXkbdCode()
	OrModifiers(int)
	GetModifiers() int
}

type PointerMove struct {
	Row       int
	Col       int
	Modifiers int
}

func (*PointerMove) isPointerEvent() {}
func (*PointerMove) isXkbdCode()     {}
func (p *PointerMove) OrModifiers(modifiers int) {
	p.Modifiers |= modifiers
}

func (p *PointerMove) GetModifiers() int {
	return p.Modifiers
}

type PointerButtonPress struct {
	Modifiers                 int
	NeedToReleaseOtherButtons bool
	Button                    LINUX_BUTTON_CODES
}

func (*PointerButtonPress) isPointerEvent() {}
func (*PointerButtonPress) isXkbdCode()     {}
func (p *PointerButtonPress) OrModifiers(modifiers int) {
	p.Modifiers |= modifiers
}

func (p *PointerButtonPress) GetModifiers() int {
	return p.Modifiers
}

/**
 * Pointer button release is special
 * because we can't be sure of which
 * button is being released
 */
type PointerButtonRelease struct {
	Button              LINUX_BUTTON_CODES
	NeedsButtonGuessing bool
	Modifiers           int
}

func (*PointerButtonRelease) isPointerEvent() {}
func (*PointerButtonRelease) isXkbdCode()     {}
func (p *PointerButtonRelease) OrModifiers(modifiers int) {
	p.Modifiers |= modifiers
}

func (p *PointerButtonRelease) GetModifiers() int {
	return p.Modifiers
}

type PointerWheel struct {
	Up        bool
	Modifiers int
}

func (*PointerWheel) isPointerEvent() {}
func (*PointerWheel) isXkbdCode()     {}
func (p *PointerWheel) OrModifiers(modifiers int) {
	p.Modifiers |= modifiers
}

func (p *PointerWheel) GetModifiers() int {
	return p.Modifiers
}

func MouseModifiers(code, base int) int {
	modeType := code - base
	modifiers := 0
	if (modeType & 0b1000) != 0 {
		modifiers |= ModControl
	}
	if (modeType & 0b1_0000) != 0 {
		modifiers |= ModAlt
	}
	return modifiers
}

func ParseMouseCode(code string) XkbdCode {
	parts := strings.Split(code, ";")
	if len(parts) != 3 {
		return nil
	}
	buttonPart := parts[0]
	colPart := parts[1]
	rowAndTermPart := parts[2]
	pressRelease := rowAndTermPart[len(rowAndTermPart)-1]
	rowPart := rowAndTermPart[:len(rowAndTermPart)-1]

	button, err := strconv.Atoi(buttonPart)
	if err != nil {
		return nil
	}
	col, err := strconv.Atoi(colPart)
	if err != nil {
		return nil
	}
	col = col - 1
	row, err := strconv.Atoi(rowPart)
	if err != nil {
		return nil
	}
	row = row - 1
	press := false
	switch pressRelease {
	case 'M':
		press = true
	case 'm':
		press = false
	default:
		return nil
	}

	d := button + 32
	/**
	* Mouse time!
	 */
	switch d {

	case 67, 75, 83, 91:
		modifiers := MouseModifiers(d, 67)
		return &PointerMove{
			Row:       row,
			Col:       col,
			Modifiers: modifiers,
		}
	case 64, 72, 80, 88:
		/**
		* This is pointer moving while
		* holding left mouse button down
		*
		* so far it has always followed
		* a button down event,
		* so I'm just sending a pointer move
		* rather than a button followed by a move
		 */
		modifiers := MouseModifiers(d, 64)
		return &PointerMove{
			Row:       row,
			Col:       col,
			Modifiers: modifiers,
		}
	case 65, 73, 81, 89:
		/**
		 * Move while holding middle mouse button down
		 */
		modifiers := MouseModifiers(d, 65)
		return &PointerMove{
			Row:       row,
			Col:       col,
			Modifiers: modifiers,
		}
	case 66, 74, 82, 90:
		/**
		 * Move while holding right mouse button down
		 */
		modifiers := MouseModifiers(d, 66)
		return &PointerMove{
			Row:       row,
			Col:       col,
			Modifiers: modifiers,
		}

	// Mouse button left down
	case 32, 40, 48, 56:
		if press {
			return &PointerButtonPress{
				Button:                    BTN_LEFT,
				NeedToReleaseOtherButtons: false,
				Modifiers:                 MouseModifiers(d, 32),
			}
		}
		return &PointerButtonRelease{
			Button:    BTN_LEFT,
			Modifiers: MouseModifiers(d, 32),
		}
	// Mouse button middle down
	case 33, 41, 49, 57:
		if press {
			return &PointerButtonPress{
				Button:                    BTN_MIDDLE,
				NeedToReleaseOtherButtons: false,
				Modifiers:                 MouseModifiers(d, 33),
			}
		}
		return &PointerButtonRelease{
			Button:    BTN_MIDDLE,
			Modifiers: MouseModifiers(d, 33),
		}
	// Mouse button right down
	case 34, 42, 50, 58:
		if press {
			return &PointerButtonPress{
				Button:                    BTN_RIGHT,
				NeedToReleaseOtherButtons: false,
				Modifiers:                 MouseModifiers(d, 34),
			}
		}
		return &PointerButtonRelease{
			Button:    BTN_RIGHT,
			Modifiers: MouseModifiers(d, 34),
		}
	// Mouse wheel up
	case 96, 104, 112, 120:
		return &PointerWheel{
			Up:        true,
			Modifiers: MouseModifiers(d, 96),
		}
	// Mouse wheel down
	case 97, 105, 113, 121:
		return &PointerWheel{
			Up:        false,
			Modifiers: MouseModifiers(d, 97),
		}
	}
	return nil
}

func ParseSGRMouseSequences(data []byte) []XkbdCode {
	codes := strings.Split(string(data), "\x1b[<")
	if len(codes) < 2 {
		return nil
	}
	codes = codes[1:] // First split is empty string before first ESC[<
	out := make([]XkbdCode, 0)
	for _, code := range codes {

		buttonCode := ParseMouseCode(code)
		if buttonCode == nil {
			continue
		}
		out = append(out, buttonCode)
	}
	return out
}

func PointerCode(data []byte) PointerEvent {
	if !(len(data) >= 3 && data[0] == 27 && data[1] == 91 && data[2] == 77) {
		return nil
	}

	d := int(data[3])

	/**
	 * Mouse time!
	 */
	switch d {

	case 67, 75, 83, 91:
		// @TODO why 33
		if len(data) < 6 {
			return nil
		}
		col := int(data[4]) - 33
		row := int(data[5]) - 33
		modifiers := MouseModifiers(d, 67)
		return &PointerMove{
			Row:       row,
			Col:       col,
			Modifiers: modifiers,
		}
	case 64, 72, 80, 88:
		// @again why 33
		if len(data) < 6 {
			return nil
		}
		col := int(data[4]) - 33
		row := int(data[5]) - 33
		/**
		 * This is pointer moving while
		 * holding a button down
		 *
		 * so far it has always followed
		 * a button down event,
		 * so I'm just sending a pointer move
		 * rather than a button followed by a move
		 */
		modifiers := MouseModifiers(d, 64)
		return &PointerMove{
			Row:       row,
			Col:       col,
			Modifiers: modifiers,
		}

	// Mouse button left down
	case 32, 40, 48, 56:
		return &PointerButtonPress{
			Button:                    BTN_LEFT,
			NeedToReleaseOtherButtons: true,
			Modifiers:                 MouseModifiers(d, 32),
		}
	// Mouse button middle down
	case 33, 41, 49, 57:
		return &PointerButtonPress{
			Button:                    BTN_MIDDLE,
			NeedToReleaseOtherButtons: true,
			Modifiers:                 MouseModifiers(d, 33),
		}
	// Mouse button right down
	case 34, 42, 50, 58:
		return &PointerButtonPress{
			Button:                    BTN_RIGHT,
			NeedToReleaseOtherButtons: true,
			Modifiers:                 MouseModifiers(d, 34),
		}
	// Mouse button up (cannot be sure which button)
	case 35, 43, 51, 59:
		return &PointerButtonRelease{
			NeedsButtonGuessing: true,
			Modifiers:           MouseModifiers(d, 35),
		}
	// Mouse wheel up
	case 96, 104, 112, 120:
		return &PointerWheel{
			Up:        true,
			Modifiers: MouseModifiers(d, 96),
		}
	// Mouse wheel down
	case 97, 105, 113, 121:
		return &PointerWheel{
			Up:        false,
			Modifiers: MouseModifiers(d, 97),
		}
	}

	return nil
}
