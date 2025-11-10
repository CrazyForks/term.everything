package termeverything

/*
#cgo CFLAGS: -Wall
#include <termios.h>
#include <unistd.h>

static int get_termios(int fd, struct termios *t) { return tcgetattr(fd, t); }
static int set_termios_now(int fd, const struct termios *t) { return tcsetattr(fd, TCSANOW, t); }

// Put TTY into "raw-ish" mode for input, but keep output processing so stdout isn't garbled.
static void make_raw(struct termios *t) {
    cfmakeraw(t);               // disable canonical, echo, signals, etc.
    t->c_cc[VMIN]  = 1;
    t->c_cc[VTIME] = 0;

    // Preserve output post-processing (NL -> CRNL), like the shell default.
    t->c_oflag |= OPOST;
#ifdef ONLCR
    t->c_oflag |= ONLCR;
#endif
}
*/
import "C"
import "fmt"

func EnableRawModeFD(fd int) (func() error, error) {
	if C.isatty(C.int(fd)) == 0 {
		return func() error { return nil }, nil
	}

	var orig C.struct_termios
	if C.get_termios(C.int(fd), &orig) != 0 {
		return nil, fmt.Errorf("tcgetattr failed")
	}

	raw := orig
	C.make_raw(&raw)

	if C.set_termios_now(C.int(fd), &raw) != 0 {
		return nil, fmt.Errorf("tcsetattr (raw) failed")
	}

	restored := false
	restore := func() error {
		if restored {
			return nil
		}
		restored = true
		if C.set_termios_now(C.int(fd), &orig) != 0 {
			return fmt.Errorf("tcsetattr (restore) failed")
		}
		return nil
	}

	return restore, nil
}
