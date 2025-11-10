package termeverything

import "os"

type DisplayServerTypeEnum int

const (
	DisplayServerTypeUnknown DisplayServerTypeEnum = iota
	DisplayServerTypeX11
	DisplayServerTypeWayland
)

func DisplayServerType() DisplayServerTypeEnum {

	if displayType, ok := os.LookupEnv("XDG_SESSION_TYPE"); ok {

		switch displayType {
		case "x11":
			return DisplayServerTypeX11
		case "wayland":
			return DisplayServerTypeWayland
		}
	}
	return DisplayServerTypeUnknown
}
