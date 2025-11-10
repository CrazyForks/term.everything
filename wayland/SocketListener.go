package wayland

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
)

// SocketListener listens for new Wayland client connections on a Unix socket.
// It provides a channel to receive new connections. Whoevere reads from
// the channel is responsible for closing the connections when done.
type SocketListener struct {
	WaylandDisplayName string
	SocketPath         string
	Listener           *net.UnixListener

	OnConnection chan *net.UnixConn
}

type HasDisplayName interface {
	WaylandDisplayName() string
}

func MakeSocketListener(args HasDisplayName) (*SocketListener, error) {
	displayName := GetWaylandDisplayName(args)
	socketPath := GetSocketPathFromName(displayName)

	ln, fd, err := ListenToWaylandSocket(displayName, socketPath)
	if err != nil {
		return nil, fmt.Errorf("failed to listen on wayland socket: %w", err)
	}
	_ = fd

	w := &SocketListener{
		WaylandDisplayName: displayName,
		SocketPath:         socketPath,
		Listener:           ln,
		OnConnection:       make(chan *net.UnixConn, 32),
	}

	return w, nil
}

func (w *SocketListener) MainLoop() error {
	for {
		conn, err := w.Listener.AcceptUnix()
		if err != nil {
			return fmt.Errorf("failed to accept connection: %w", err)
		}
		w.OnConnection <- conn
	}
}

func (w *SocketListener) MainLoopThenClose() error {
	defer w.Close()
	return w.MainLoop()
}

func (w *SocketListener) Close() error {
	if w.Listener != nil {
		w.Listener.Close()
	}
	return removeFileIfExists(w.SocketPath)
}

func GetWaylandDisplayName(args HasDisplayName) string {
	if args.WaylandDisplayName() != "" {
		return args.WaylandDisplayName()
	}
	if v := os.Getenv("WAYLAND_DISPLAY_NAME"); v != "" {
		return v
	}

	for i := 2; i < 1000; i++ {
		name := fmt.Sprintf("wayland-%d", i)
		path := GetSocketPathFromName(name)
		if _, err := os.Stat(path); err == nil {
			continue
		} else if os.IsNotExist(err) {
			return name
		} else {
			continue
		}
	}
	fmt.Fprintf(os.Stderr, "Failed to find an open wayland socket name. Pass one with --wayland-display-name <name>\n")
	os.Exit(1)
	return ""
}

func GetSocketPathFromName(socketName string) string {
	runtimeDir := os.Getenv("XDG_RUNTIME_DIR")
	if runtimeDir == "" {
		runtimeDir = "/tmp"
	}
	return filepath.Join(runtimeDir, socketName)
}

func removeFileIfExists(p string) error {
	if p == "" {
		return fmt.Errorf("empty path")
	}
	if _, err := os.Lstat(p); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	return os.Remove(p)
}
