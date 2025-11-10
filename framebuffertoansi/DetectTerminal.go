package framebuffertoansi

// #cgo pkg-config: chafa glib-2.0
// #include "chafa.h"
// #include <glib.h>
//
// static ChafaTermInfo* detect_term_info_from_env() {
//     gchar **envp = g_get_environ();
//     ChafaTermInfo *term_info = chafa_term_db_detect(chafa_term_db_get_default(), envp);
//     g_strfreev(envp);
//     return term_info;
// }
import "C"
import "os"

func getDefaultPixelMode(termInfo *C.ChafaTermInfo) C.ChafaPixelMode {
	if C.chafa_term_info_have_seq(termInfo, C.CHAFA_TERM_SEQ_BEGIN_ITERM2_IMAGE) != 0 {
		return C.CHAFA_PIXEL_MODE_ITERM2
	} else if C.chafa_term_info_have_seq(termInfo, C.CHAFA_TERM_SEQ_BEGIN_KITTY_IMMEDIATE_IMAGE_V1) != 0 {
		return C.CHAFA_PIXEL_MODE_KITTY
	} else if C.chafa_term_info_have_seq(termInfo, C.CHAFA_TERM_SEQ_BEGIN_SIXELS) != 0 {
		return C.CHAFA_PIXEL_MODE_SIXELS
	} else {
		return C.CHAFA_PIXEL_MODE_SYMBOLS
	}
}

func getPixelMode(termInfo *C.ChafaTermInfo) C.ChafaPixelMode {
	override := os.Getenv("TERM_EVERYTHING_PIXEL_MODE")
	if override == "" {
		return getDefaultPixelMode(termInfo)
	}
	switch override {
	case "SYMBOLS":
		return C.CHAFA_PIXEL_MODE_SYMBOLS
	case "SIXELS":
		return C.CHAFA_PIXEL_MODE_SIXELS
	case "KITTY":
		return C.CHAFA_PIXEL_MODE_KITTY
	case "ITERM2":
		return C.CHAFA_PIXEL_MODE_ITERM2
	default:
		return getDefaultPixelMode(termInfo)
	}
}

func getDefaultCanvasMode(termInfo *C.ChafaTermInfo, pixelMode C.ChafaPixelMode) C.ChafaCanvasMode {
	switch pixelMode {
	case C.CHAFA_PIXEL_MODE_ITERM2, C.CHAFA_PIXEL_MODE_SIXELS, C.CHAFA_PIXEL_MODE_KITTY:
		return C.CHAFA_CANVAS_MODE_TRUECOLOR
	case C.CHAFA_PIXEL_MODE_SYMBOLS:
		fallthrough
	default:
		if C.chafa_term_info_have_seq(termInfo, C.CHAFA_TERM_SEQ_SET_COLOR_FGBG_DIRECT) != 0 &&
			C.chafa_term_info_have_seq(termInfo, C.CHAFA_TERM_SEQ_SET_COLOR_FG_DIRECT) != 0 &&
			C.chafa_term_info_have_seq(termInfo, C.CHAFA_TERM_SEQ_SET_COLOR_BG_DIRECT) != 0 {
			return C.CHAFA_CANVAS_MODE_TRUECOLOR
		} else if C.chafa_term_info_have_seq(termInfo, C.CHAFA_TERM_SEQ_SET_COLOR_FGBG_256) != 0 &&
			C.chafa_term_info_have_seq(termInfo, C.CHAFA_TERM_SEQ_SET_COLOR_FG_256) != 0 &&
			C.chafa_term_info_have_seq(termInfo, C.CHAFA_TERM_SEQ_SET_COLOR_BG_256) != 0 {
			return C.CHAFA_CANVAS_MODE_INDEXED_240
		} else if C.chafa_term_info_have_seq(termInfo, C.CHAFA_TERM_SEQ_SET_COLOR_FGBG_16) != 0 &&
			C.chafa_term_info_have_seq(termInfo, C.CHAFA_TERM_SEQ_SET_COLOR_FG_16) != 0 &&
			C.chafa_term_info_have_seq(termInfo, C.CHAFA_TERM_SEQ_SET_COLOR_BG_16) != 0 {
			return C.CHAFA_CANVAS_MODE_INDEXED_16
		} else if C.chafa_term_info_have_seq(termInfo, C.CHAFA_TERM_SEQ_INVERT_COLORS) != 0 &&
			C.chafa_term_info_have_seq(termInfo, C.CHAFA_TERM_SEQ_RESET_ATTRIBUTES) != 0 {
			return C.CHAFA_CANVAS_MODE_FGBG_BGFG
		} else {
			return C.CHAFA_CANVAS_MODE_FGBG
		}
	}
}

func getCanvasMode(termInfo *C.ChafaTermInfo, pixelMode C.ChafaPixelMode) C.ChafaCanvasMode {
	override := os.Getenv("TERM_EVERYTHING_CANVAS_MODE")
	if override == "" {
		return getDefaultCanvasMode(termInfo, pixelMode)
	}
	switch override {
	case "TRUECOLOR":
		return C.CHAFA_CANVAS_MODE_TRUECOLOR
	case "INDEXED_256":
		return C.CHAFA_CANVAS_MODE_INDEXED_256
	case "INDEXED_240":
		return C.CHAFA_CANVAS_MODE_INDEXED_240
	case "INDEXED_16":
		return C.CHAFA_CANVAS_MODE_INDEXED_16
	case "FGBG_BGFG":
		return C.CHAFA_CANVAS_MODE_FGBG_BGFG
	case "FGBG":
		return C.CHAFA_CANVAS_MODE_FGBG
	case "INDEXED_8":
		return C.CHAFA_CANVAS_MODE_INDEXED_8
	case "INDEXED_16_8":
		return C.CHAFA_CANVAS_MODE_INDEXED_16_8
	default:
		return getDefaultCanvasMode(termInfo, pixelMode)
	}
}

func DetectTerminal() (termInfo *C.ChafaTermInfo, mode C.ChafaCanvasMode, pixelMode C.ChafaPixelMode) {
	termInfo = C.detect_term_info_from_env()
	pixelMode = getPixelMode(termInfo)
	mode = getCanvasMode(termInfo, pixelMode)
	return
}
