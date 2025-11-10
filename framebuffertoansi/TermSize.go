package framebuffertoansi

import (
	"os"
	"syscall"
	"unsafe"
)

type TermSize struct {
	WidthCells  int
	HeightCells int

	WidthPixels  int
	HeightPixels int

	FontRatio float64

	WidthOfACellInPixels  int
	HeightOfACellInPixels int
}
type WinSize struct {
	Row    uint16
	Col    uint16
	Xpixel uint16
	Ypixel uint16
}

func MakeTermSize() TermSize {
	ts := TermSize{
		WidthCells:            -1,
		HeightCells:           -1,
		WidthPixels:           -1,
		HeightPixels:          -1,
		WidthOfACellInPixels:  -1,
		HeightOfACellInPixels: -1,
		FontRatio:             0.5,
	}

	tryFDs := []uintptr{os.Stdout.Fd(), os.Stderr.Fd(), os.Stdin.Fd()}
	for _, fd := range tryFDs {
		if ws, err := GetWinsize(fd); err == nil {
			ts.WidthCells = int(ws.Col)
			ts.HeightCells = int(ws.Row)
			ts.WidthPixels = int(ws.Xpixel)
			ts.HeightPixels = int(ws.Ypixel)
			break
		}
	}

	if ts.WidthCells <= 0 {
		ts.WidthCells = -1
	}
	if ts.HeightCells <= 2 {
		ts.HeightCells = -1
	}

	if ts.WidthPixels <= 0 || ts.HeightPixels <= 0 {
		ts.WidthPixels = -1
		ts.HeightPixels = -1
	}

	if ts.WidthCells > 0 && ts.HeightCells > 0 && ts.WidthPixels > 0 && ts.HeightPixels > 0 {
		ts.WidthOfACellInPixels = ts.WidthPixels / ts.WidthCells
		ts.HeightOfACellInPixels = ts.HeightPixels / ts.HeightCells
		if ts.HeightOfACellInPixels > 0 {
			ts.FontRatio = float64(ts.WidthOfACellInPixels) / float64(ts.HeightOfACellInPixels)
		}
	} else {
		ts.WidthOfACellInPixels = -1
		ts.HeightOfACellInPixels = -1
		ts.FontRatio = 0.5
	}

	return ts
}

func GetWinsize(fd uintptr) (WinSize, error) {
	var ws WinSize
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, fd, uintptr(syscall.TIOCGWINSZ), uintptr(unsafe.Pointer(&ws)))
	if errno != 0 {
		return ws, errno
	}
	return ws, nil
}
