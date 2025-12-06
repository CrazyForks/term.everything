package termeverything

// /**
//   - According to sleuthing here are the mod makes
//   - 1 << 0 Shift
//   - 1 << 1 Lock
//   - 1 << 2 Control
//   - 1 << 3 Alt
//   - 1 << 4 Mod2
//   - 1 << 5 Mod3
//   - 1 << 6 Mod4
//   - 1 << 7 Mod5
//   - 1 << 8 Button1
//   - 1 << 9 Button2
//   - 1 << 10 Button3
//   - 1 << 11 Button4
//   - 1 << 12 Button5
const (
	ModShift   = 1 << 0
	ModLock    = 1 << 1
	ModControl = 1 << 2
	ModAlt     = 1 << 3
)

// numericKeys and alphaKeys exported here for KeycodeSingleCodes.go
var numericKeys = []Linux_Event_Codes{
	KEY_0,
	KEY_1,
	KEY_2,
	KEY_3,
	KEY_4,
	KEY_5,
	KEY_6,
	KEY_7,
	KEY_8,
	KEY_9,
}

var alphaKeys = []Linux_Event_Codes{
	KEY_A,
	KEY_B,
	KEY_C,
	KEY_D,
	KEY_E,
	KEY_F,
	KEY_G,
	KEY_H,
	KEY_I,
	KEY_J,
	KEY_K,
	KEY_L,
	KEY_M,
	KEY_N,
	KEY_O,
	KEY_P,
	KEY_Q,
	KEY_R,
	KEY_S,
	KEY_T,
	KEY_U,
	KEY_V,
	KEY_W,
	KEY_X,
	KEY_Y,
	KEY_Z,
}

type XkbdCode interface {
	isXkbdCode()
	OrModifiers(int)
	GetModifiers() int
}

func (*KeyCode) isXkbdCode() {}

type KeyCode struct {
	KeyCode   Linux_Event_Codes
	Modifiers int
}

func (k *KeyCode) OrModifiers(modifiers int) {
	k.Modifiers |= modifiers
}

func (k *KeyCode) GetModifiers() int {
	return k.Modifiers
}

func ConvertKeycodeToXbdCode(data []byte) []XkbdCode {
	if len(data) == 1 {
		if out := KeycodeSingleCodes(int(data[0])); out != nil {
			return []XkbdCode{out}
		}
		return nil
	}
	if len(data) == 2 {
		return parse_length_2(data)
	}
	if len(data) == 3 {
		return parse_length_3(data)
	}
	if data[0] == 27 && data[1] == 91 && data[2] == 60 {
		return ParseSGRMouseSequences(data)
	}

	if len(data) == 4 {
		return parse_length_4(data)
	}
	if len(data) == 5 {
		if data[0] == 27 && data[1] == 91 && data[4] == 126 {
			if data[2] == 49 {
				if keyCode, ok := f5_through_8_codes[data[3]]; ok {
					return []XkbdCode{&KeyCode{KeyCode: keyCode, Modifiers: 0}}
				}
			}
			if data[2] == 50 {
				if keyCode, ok := f9_through_12_codes[data[3]]; ok {
					return []XkbdCode{&KeyCode{KeyCode: keyCode, Modifiers: 0}}
				}
			}
		}
	}

	if len(data)%6 == 0 {
		out := make([]XkbdCode, 0)
		numEvents := len(data) / 6
		for i := range numEvents {
			slice := data[i*6 : (i+1)*6]
			if slice[0] == 27 && slice[1] == 91 && slice[3] == 59 {
				switch slice[2] {
				case 49: // '1'
					if modifiers := modifiers_for_arrow_and_page_up_etc(slice[4]); modifiers > 0 {
						eventOut := parse_length_3([]byte{27, 91, slice[5]})
						if len(eventOut) > 0 {
							for _, e := range eventOut {
								e.OrModifiers(modifiers)
							}
							out = append(out, eventOut...)
							continue
						}
					}
				case 50, 51, 52, 53, 54: // '2'..'6'
					if modifiers := modifiers_for_arrow_and_page_up_etc(slice[4]); slice[5] == 126 && modifiers > 0 {
						eventOut := parse_length_4([]byte{27, 91, slice[2], 126})
						if len(eventOut) > 0 {
							for _, e := range eventOut {
								e.OrModifiers(modifiers)
							}
							out = append(out, eventOut...)
							continue
						}
					}
				}
			}

			// if value := PointerCode(slice); value != nil {
			// 	out = append(out, value)
			// }
		}
		return out
	}

	if len(data) == 7 {
		if data[0] == 27 && data[1] == 91 && data[4] == 59 && data[6] == 126 {
			if modifiers := modifiers_for_arrow_and_page_up_etc(data[5]); modifiers > 0 {
				if data[2] == 49 {
					if keyCode, ok := f5_through_8_codes[data[3]]; ok {
						return []XkbdCode{&KeyCode{KeyCode: keyCode, Modifiers: modifiers}}
					}
				}
				if data[2] == 50 {
					if keyCode, ok := f9_through_12_codes[data[3]]; ok {
						return []XkbdCode{&KeyCode{KeyCode: keyCode, Modifiers: modifiers}}
					}
				}
			}
		}
	}

	// Unrecognized
	// TODO maybe return error here?
	return nil
}

func parse_length_2(data []byte) []XkbdCode {
	if len(data) < 2 {
		return nil
	}
	if data[0] == 27 {
		out := KeycodeSingleCodes(int(data[1]))
		if out == nil {
			return nil
		}
		out.Modifiers |= ModAlt
		return []XkbdCode{out}
	}
	out := make([]XkbdCode, 0, 2)
	if out1 := KeycodeSingleCodes(int(data[0])); out1 != nil {
		out = append(out, out1)
	}
	if out2 := KeycodeSingleCodes(int(data[1])); out2 != nil {
		out = append(out, out2)
	}
	return out
}

func parse_length_3(data []byte) []XkbdCode {
	if len(data) < 3 {
		return nil
	}
	if data[0] != 27 {
		a := KeycodeSingleCodes(int(data[0]))
		b := parse_length_2(data[1:])
		if a != nil {
			return append([]XkbdCode{a}, b...)
		}
		return b
	}
	if data[1] == 79 { // 'O'
		switch data[2] {
		case 80: // 'P'
			return []XkbdCode{&KeyCode{KeyCode: KEY_F1, Modifiers: 0}}
		case 81: // 'Q'
			return []XkbdCode{&KeyCode{KeyCode: KEY_F2, Modifiers: 0}}
		case 82: // 'R'
			return []XkbdCode{&KeyCode{KeyCode: KEY_F3, Modifiers: 0}}
		case 83: // 'S'
			return []XkbdCode{&KeyCode{KeyCode: KEY_F4, Modifiers: 0}}
		}
	}
	if data[1] == 91 { // '['
		switch data[2] {
		case 65: // 'A'
			return []XkbdCode{&KeyCode{KeyCode: KEY_UP, Modifiers: 0}}
		case 66: // 'B'
			return []XkbdCode{&KeyCode{KeyCode: KEY_DOWN, Modifiers: 0}}
		case 67: // 'C'
			return []XkbdCode{&KeyCode{KeyCode: KEY_RIGHT, Modifiers: 0}}
		case 68: // 'D'
			return []XkbdCode{&KeyCode{KeyCode: KEY_LEFT, Modifiers: 0}}
		case 70: // 'F'
			return []XkbdCode{&KeyCode{KeyCode: KEY_END, Modifiers: 0}}
		case 72: // 'H'
			return []XkbdCode{&KeyCode{KeyCode: KEY_HOME, Modifiers: 0}}
		case 90: // 'Z' => Shift+Tab
			return []XkbdCode{&KeyCode{KeyCode: KEY_TAB, Modifiers: ModShift}}
		// These work for alt+F1, shift+F2, etc in some terminals
		case 80: // 'P'
			return []XkbdCode{&KeyCode{KeyCode: KEY_F1, Modifiers: 0}}
		case 81: // 'Q'
			return []XkbdCode{&KeyCode{KeyCode: KEY_F2, Modifiers: 0}}
		case 82: // 'R'
			return []XkbdCode{&KeyCode{KeyCode: KEY_F3, Modifiers: 0}}
		case 83: // 'S'
			return []XkbdCode{&KeyCode{KeyCode: KEY_F4, Modifiers: 0}}
		}
	}
	return nil
}

func modifiers_for_arrow_and_page_up_etc(slice4 byte) int {
	switch slice4 {
	case 50: // '2'
		return ModShift
	case 51: // '3'
		return ModAlt
	case 52: // '4'
		return ModShift | ModAlt
	case 53: // '5'
		return ModControl
	case 54: // '6'
		return ModControl | ModShift
	default:
		return -1
	}
}

func parse_length_4(data []byte) []XkbdCode {
	if len(data) < 4 {
		return nil
	}
	if data[0] != 27 {
		a := KeycodeSingleCodes(int(data[0]))
		b := parse_length_3(data[1:])
		if a != nil {
			return append([]XkbdCode{a}, b...)
		}
		return b
	}
	if data[1] == 91 { // '['
		if data[2] == 50 && data[3] == 126 { // "2~"
			return []XkbdCode{&KeyCode{KeyCode: KEY_INSERT, Modifiers: 0}}
		}
		if data[2] == 51 && data[3] == 126 { // "3~"
			return []XkbdCode{&KeyCode{KeyCode: KEY_DELETE, Modifiers: 0}}
		}
		if data[2] == 53 && data[3] == 126 { // "5~"
			return []XkbdCode{&KeyCode{KeyCode: KEY_PAGEUP, Modifiers: 0}}
		}
		if data[2] == 54 && data[3] == 126 { // "6~"
			return []XkbdCode{&KeyCode{KeyCode: KEY_PAGEDOWN, Modifiers: 0}}
		}
	}
	return nil
}

var f5_through_8_codes = map[byte]Linux_Event_Codes{
	53: KEY_F5, // '5'
	55: KEY_F6, // '7'
	56: KEY_F7, // '8'
	57: KEY_F8, // '9'
}

var f9_through_12_codes = map[byte]Linux_Event_Codes{
	48: KEY_F9,  // '0'
	49: KEY_F10, // '1'
	51: KEY_F11, // '3'
	52: KEY_F12, // '4'
}
