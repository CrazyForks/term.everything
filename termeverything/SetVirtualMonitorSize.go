package termeverything

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/mmulet/term.everything/wayland"
)

func SetVirtualMonitorSize(newVirtualMonitorSize string) {
	if newVirtualMonitorSize == "" {
		return
	}
	parts := strings.Split(newVirtualMonitorSize, "x")
	if len(parts) != 2 {
		fmt.Fprintf(os.Stderr, "Invalid virtual monitor size %s, expected <width>x<height>\n", newVirtualMonitorSize)
		os.Exit(1)
	}
	width, err1 := strconv.Atoi(parts[0])
	height, err2 := strconv.Atoi(parts[1])
	if err1 != nil || err2 != nil {
		fmt.Fprintf(os.Stderr, "Invalid virtual monitor size %s, expected <width>x<height>\n", newVirtualMonitorSize)
		os.Exit(1)
	}
	if width <= 0 || height <= 0 {
		fmt.Fprintf(os.Stderr, "Invalid virtual monitor size %s, expected <width>x<height>\n", newVirtualMonitorSize)
		os.Exit(1)
	}
	wayland.VirtualMonitorSize.Width = wayland.Pixels(width)
	wayland.VirtualMonitorSize.Height = wayland.Pixels(height)
}
