package termeverything

import "fmt"

func KeycodeSingleCodes(d int) *KeyCode {
	if d >= 1 && d <= 26 {
		/**
		 * @TODO not sure what to do about the
		 * ctrl+keys that are shadowed
		 * by these keys
		 */
		switch d {
		case 3, 9, 13:
			// skip (handled below)
		default:
			return &KeyCode{
				KeyCode:   alphaKeys[d-1],
				Modifiers: ModControl,
			}
		}
	}

	if d >= 48 && d <= 57 {
		return &KeyCode{
			KeyCode:   numericKeys[d-48],
			Modifiers: 0,
		}
	}

	if d >= 65 && d <= 90 {
		return &KeyCode{
			KeyCode:   alphaKeys[d-65],
			Modifiers: ModShift,
		}
	}

	if d >= 97 && d <= 122 {
		return &KeyCode{
			KeyCode:   alphaKeys[d-97],
			Modifiers: 0,
		}
	}

	switch d {
	case 33: // !
		return &KeyCode{KeyCode: KEY_1, Modifiers: ModShift}
	case 64: // @
		return &KeyCode{KeyCode: KEY_2, Modifiers: ModShift}
	case 35: // #
		return &KeyCode{KeyCode: KEY_3, Modifiers: ModShift}
	case 36: // $
		return &KeyCode{KeyCode: KEY_4, Modifiers: ModShift}
	case 37: // %
		return &KeyCode{KeyCode: KEY_5, Modifiers: ModShift}
	case 34: // "
		return &KeyCode{KeyCode: KEY_APOSTROPHE, Modifiers: ModShift}
	case 39: // '
		return &KeyCode{KeyCode: KEY_APOSTROPHE, Modifiers: 0}
	case 94: // ^
		return &KeyCode{KeyCode: KEY_6, Modifiers: ModShift}
	case 38: // &
		return &KeyCode{KeyCode: KEY_7, Modifiers: ModShift}
	case 42: // *
		return &KeyCode{KeyCode: KEY_8, Modifiers: ModShift}
	case 40: // (
		return &KeyCode{KeyCode: KEY_9, Modifiers: ModShift}
	case 41: // )
		return &KeyCode{KeyCode: KEY_0, Modifiers: ModShift}

	case 3: // escape (as per original TS)
		return &KeyCode{KeyCode: KEY_ESC, Modifiers: 0}
	case 27: // escape (shift)
		return &KeyCode{KeyCode: KEY_ESC, Modifiers: ModShift}

	case 96: // `
		return &KeyCode{KeyCode: KEY_GRAVE, Modifiers: 0}
	case 126: // ~
		return &KeyCode{KeyCode: KEY_GRAVE, Modifiers: ModShift}

	case 45: // -
		return &KeyCode{KeyCode: KEY_MINUS, Modifiers: 0}
	case 95: // _
		return &KeyCode{KeyCode: KEY_MINUS, Modifiers: ModShift}

	case 61: // =
		return &KeyCode{KeyCode: KEY_EQUAL, Modifiers: 0}
	case 43: // +
		return &KeyCode{KeyCode: KEY_EQUAL, Modifiers: ModShift}

	case 8: // CTRL+backspace (overshadowed by CTRL+H)
		return &KeyCode{KeyCode: KEY_BACKSPACE, Modifiers: ModControl}
	case 127: // backspace
		return &KeyCode{KeyCode: KEY_BACKSPACE, Modifiers: 0}

	case 9: // tab
		return &KeyCode{KeyCode: KEY_TAB, Modifiers: 0}
	case 13: // enter
		return &KeyCode{KeyCode: KEY_ENTER, Modifiers: 0}

	case 32: // space
		return &KeyCode{KeyCode: KEY_SPACE, Modifiers: 0}
	case 0: // ctrl + space
		return &KeyCode{KeyCode: KEY_SPACE, Modifiers: ModControl}

	case 59: // ;
		return &KeyCode{KeyCode: KEY_SEMICOLON, Modifiers: 0}
	case 58: // :
		return &KeyCode{KeyCode: KEY_SEMICOLON, Modifiers: ModShift}

	case 91: // [
		return &KeyCode{KeyCode: KEY_LEFTBRACE, Modifiers: 0}
	case 123: // {
		return &KeyCode{KeyCode: KEY_LEFTBRACE, Modifiers: ModShift}

	case 93: // ]
		return &KeyCode{KeyCode: KEY_RIGHTBRACE, Modifiers: 0}
	case 125: // }
		return &KeyCode{KeyCode: KEY_RIGHTBRACE, Modifiers: ModShift}
	case 129: // ctrl+]
		return &KeyCode{KeyCode: KEY_RIGHTBRACE, Modifiers: ModControl}

	case 92: // \
		return &KeyCode{KeyCode: KEY_BACKSLASH, Modifiers: 0}
	case 124: // |
		return &KeyCode{KeyCode: KEY_BACKSLASH, Modifiers: ModShift}

	case 44: // ,
		return &KeyCode{KeyCode: KEY_COMMA, Modifiers: 0}
	case 60: // <
		return &KeyCode{KeyCode: KEY_COMMA, Modifiers: ModShift}

	case 46: // .
		return &KeyCode{KeyCode: KEY_DOT, Modifiers: 0}
	case 62: // >
		return &KeyCode{KeyCode: KEY_DOT, Modifiers: ModShift}

	case 47: // /
		return &KeyCode{KeyCode: KEY_SLASH, Modifiers: 0}
	case 63: // ?
		return &KeyCode{KeyCode: KEY_SLASH, Modifiers: ModShift}
	case 31: // ctrl+/
		return &KeyCode{KeyCode: KEY_SLASH, Modifiers: ModControl}
	}

	fmt.Printf("Unrecognized key code: %d\n", d)
	return nil
}
