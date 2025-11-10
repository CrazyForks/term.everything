package main

import (
	"fmt"
	"strings"
)

func argFlatmap(a Arg, postfix string) []string {
	switch v := a.(type) {
	case *ArgNewID:
		if v.Interface == nil {
			base := sanitizedArgName(a)
			return []string{
				base + "Interface" + postfix,
				base + "Version" + postfix,
				base + "ID" + postfix,
			}
		}
	}
	return []string{sanitizedArgName(a) + postfix}
}

func genRequestHandler(i Interface) string {
	if len(i.Requests) == 0 {
		return ""
	}

	var out strings.Builder

	for idx, req := range i.Requests {
		var argList []string
		for _, a := range req.Args {
			argList = append(argList, argFlatmap(a, "")...)
		}

		var debugPieces []string
		for _, a := range req.Args {
			for _, nm := range argFlatmap(a, "") {
				debugPieces = append(debugPieces, fmt.Sprintf("\"%s: \", %s", nm, nm))
			}
		}
		debugArgs := ""
		if len(debugPieces) > 0 {
			debugArgs = strings.Join(debugPieces, ", \", \", ")
		} else {
			debugArgs = "\")\""
		}

		isAutoRemove := req.Name == "release" || req.Name == "destroy"

		fmt.Fprintf(&out, "case %d: {\n\n", idx)

		for _, a := range req.Args {
			out.WriteString(genArgParseCode(a, i.Name))
			out.WriteString("\n")
		}

		out.WriteString("if DebugRequests {\n")
		fmt.Fprintf(&out, "  fmt.Print(\"%s@\", message.ObjectID, \".%s(\")\n", i.Name, req.Name)
		if len(debugPieces) > 0 {
			fmt.Fprintf(&out, "  fmt.Println(%s, \")\")\n", debugArgs)
		} else {
			out.WriteString("  fmt.Println(\")\")\n")
		}
		out.WriteString("}\n\n")

		call := fmt.Sprintf("d.%s_%s(s, ObjectID[%s](message.ObjectID), %s)", i.Name, req.Name, i.Name, strings.Join(argList, ", "))
		if isAutoRemove {
			fmt.Fprintf(&out, "autoRemove := %s\n", call)
		} else {
			fmt.Fprintf(&out, "%s\n", call)
		}

		if req.Name == "release" {
			out.WriteString("if autoRemove {\n")
			out.WriteString("  s.RemoveObject(message.ObjectID)\n")

			switch i.Name {
			case "WlShm", "WlSeat", "WlOutput", "WlKeyboard", "WlPointer", "WlTouch", "WlDataDevice", "ZwpXwaylandKeyboardGrabManagerV1":
				fmt.Fprintf(&out, "  s.RemoveGlobal%sBind(ObjectID[%s](message.ObjectID))\n", i.Name, i.Name)
			}

			out.WriteString("}\n")
		}
		if req.Name == "destroy" {
			out.WriteString("if autoRemove {\n")
			out.WriteString("  s.RemoveObject(message.ObjectID)\n")
			out.WriteString("}\n")
		}

		out.WriteString("break\n")
		out.WriteString("}\n\n")
	}

	return out.String()
}

func genArgParseCode(a Arg, interfaceName string) string {
	name := sanitizedArgName(a)

	switch v := a.(type) {
	case *ArgFixed:
		return fmt.Sprintf(`%sRaw := uint32(message.Data[_data_in_offset__+0]) | uint32(message.Data[_data_in_offset__+1])<<8 |
  uint32(message.Data[_data_in_offset__+2])<<16 | uint32(message.Data[_data_in_offset__+3])<<24
%s := float64(int32(%sRaw)) / 256.0
_data_in_offset__ += 4
`, name, name, name)

	case *ArgNewID:
		if v.Interface == nil {
			return genArgParseCode(&ArgString{ArgCommon: ArgCommon{ArgName: name + "Interface"}}, interfaceName) +
				genArgParseCode(&ArgUint{ArgCommon: ArgCommon{ArgName: name + "Version"}}, interfaceName) +
				genArgParseCode(&ArgObject{ArgCommon: ArgCommon{ArgName: name + "ID"}}, interfaceName)
		}
		return fmt.Sprintf(`%sVal := uint32(message.Data[_data_in_offset__+0]) | uint32(message.Data[_data_in_offset__+1])<<8 |
  uint32(message.Data[_data_in_offset__+2])<<16 | uint32(message.Data[_data_in_offset__+3])<<24
%s := ObjectID[%s](%sVal)
_data_in_offset__ += 4
`, name, name, *v.Interface, name)

	case *ArgUint:
		if v.Enum != nil {
			return fmt.Sprintf(`%s := %s(uint32(message.Data[_data_in_offset__+0]) | uint32(message.Data[_data_in_offset__+1])<<8 |
  uint32(message.Data[_data_in_offset__+2])<<16 | uint32(message.Data[_data_in_offset__+3])<<24)
_data_in_offset__ += 4
`, name, enumName(interfaceName, *v.Enum))
		}
		return fmt.Sprintf(`%s := uint32(message.Data[_data_in_offset__+0]) | uint32(message.Data[_data_in_offset__+1])<<8 |
  uint32(message.Data[_data_in_offset__+2])<<16 | uint32(message.Data[_data_in_offset__+3])<<24
_data_in_offset__ += 4
`, name)

	case *ArgInt:
		return fmt.Sprintf(`%s := int32(uint32(message.Data[_data_in_offset__+0]) | uint32(message.Data[_data_in_offset__+1])<<8 |
  uint32(message.Data[_data_in_offset__+2])<<16 | uint32(message.Data[_data_in_offset__+3])<<24)
_data_in_offset__ += 4
`, name)

	case *ArgObject:
		if v.Interface != nil {
			if v.AllowNull != nil && *v.AllowNull {
				return fmt.Sprintf(`%sTmp := uint32(message.Data[_data_in_offset__+0]) | uint32(message.Data[_data_in_offset__+1])<<8 |
  uint32(message.Data[_data_in_offset__+2])<<16 | uint32(message.Data[_data_in_offset__+3])<<24
_data_in_offset__ += 4
var %s *ObjectID[%s]
if %sTmp != 0 {
  tmp := ObjectID[%s](%sTmp)
  %s = &tmp
}
`, name, name, *v.Interface, name, *v.Interface, name, name)
			}
			return fmt.Sprintf(`%s := ObjectID[%s](uint32(message.Data[_data_in_offset__+0]) | uint32(message.Data[_data_in_offset__+1])<<8 |
  uint32(message.Data[_data_in_offset__+2])<<16 | uint32(message.Data[_data_in_offset__+3])<<24)
_data_in_offset__ += 4
`, name, *v.Interface)
		}
		return fmt.Sprintf(`%s := AnyObjectID(uint32(message.Data[_data_in_offset__+0]) | uint32(message.Data[_data_in_offset__+1])<<8 |
  uint32(message.Data[_data_in_offset__+2])<<16 | uint32(message.Data[_data_in_offset__+3])<<24)
_data_in_offset__ += 4
`, name)

	case *ArgString:
		return fmt.Sprintf(`%sLen := int(uint32(message.Data[_data_in_offset__+0]) | uint32(message.Data[_data_in_offset__+1])<<8 |
  uint32(message.Data[_data_in_offset__+2])<<16 | uint32(message.Data[_data_in_offset__+3])<<24)
_data_in_offset__ += 4
%s := string(message.Data[_data_in_offset__ : _data_in_offset__+%sLen-1]) // NUL-terminated
// 4-byte alignment
if %sLen%%4 != 0 {
  _data_in_offset__ += %sLen + (4 - (%sLen %% 4))
} else {
  _data_in_offset__ += %sLen
}
`, name, name, name, name, name, name, name)

	case *ArgArray:
		return fmt.Sprintf(`%sLen := int(uint32(message.Data[_data_in_offset__+0]) | uint32(message.Data[_data_in_offset__+1])<<8 |
  uint32(message.Data[_data_in_offset__+2])<<16 | uint32(message.Data[_data_in_offset__+3])<<24)
_data_in_offset__ += 4
%s := message.Data[_data_in_offset__ : _data_in_offset__+%sLen]
if %sLen%%4 != 0 {
  _data_in_offset__ += %sLen + (4 - (%sLen %% 4))
} else {
  _data_in_offset__ += %sLen
}
`, name, name, name, name, name, name, name)

	case *ArgFd:
		return fmt.Sprintf(`%s := s.ClaimFileDescriptor()
`, name)

	default:
		panic(fmt.Errorf("unknown arg kind: %T", a))
	}
}
