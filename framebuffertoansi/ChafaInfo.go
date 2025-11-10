package framebuffertoansi

// #cgo pkg-config: chafa glib-2.0
// #include "chafa.h"
// #include <glib.h>
//
// // Small helpers to safely access GString from Go.
// static const char* gstring_str(const GString* s) { return s->str; }
// static gsize gstring_len(const GString* s) { return s->len; }
//
// // Feature test for OCTANT availability to mimic the original #ifdef.
// #ifdef CHAFA_VERSION_1_16
// #define HAVE_CHAFA_OCTANT 1
// static ChafaSymbolTags octant_symbol() { return CHAFA_SYMBOL_TAG_OCTANT; }
// #else
// #define HAVE_CHAFA_OCTANT 0
// static ChafaSymbolTags octant_symbol() { return 0; }
// #endif
// static int chafa_have_octant() { return HAVE_CHAFA_OCTANT; }
//
import "C"
import (
	"os"
	"runtime"
	"unsafe"
)

type ChafaInfo struct {
	TermInfo  *C.ChafaTermInfo
	Mode      C.ChafaCanvasMode
	PixelMode C.ChafaPixelMode
	SymbolMap *C.ChafaSymbolMap
	Config    *C.ChafaCanvasConfig
	Canvas    *C.ChafaCanvas

	WidthCells            int
	HeightCells           int
	WidthOfACellInPixels  int
	HeightOfACellInPixels int
	SessionTypeIsX11      bool
	PixelTypeOverride     C.ChafaPixelType
}

func (ci *ChafaInfo) ConvertImage(texturePixels []byte, textureWidth, textureHeight, textureStride uint32) string {
	if len(texturePixels) == 0 {
		return ""
	}

	pixelsPtr := (*C.guint8)(unsafe.Pointer(&texturePixels[0]))

	C.chafa_canvas_draw_all_pixels(
		ci.Canvas,
		ci.getPixelType(),
		pixelsPtr,
		C.int(textureWidth),
		C.int(textureHeight),
		C.int(textureStride),
	)
	runtime.KeepAlive(texturePixels)

	gstr := C.chafa_canvas_print(ci.Canvas, ci.TermInfo)
	if gstr == nil {
		return ""
	}
	defer C.g_string_free(gstr, C.gboolean(1))

	ptrStr := C.gstring_str(gstr)
	length := C.gstring_len(gstr)
	if ptrStr == nil || length == 0 {
		return ""
	}
	return C.GoStringN((*C.char)(unsafe.Pointer(ptrStr)), C.int(length))
}

func MakeChafaInfo(widthCells, heightCells, widthOfACellInPixels, heightOfACellInPixels int, sessionTypeIsX11 bool) *ChafaInfo {
	termInfo, mode, pixelMode := DetectTerminal()

	ci := &ChafaInfo{
		TermInfo:              termInfo,
		Mode:                  mode,
		PixelMode:             pixelMode,
		WidthCells:            widthCells,
		HeightCells:           heightCells,
		WidthOfACellInPixels:  widthOfACellInPixels,
		HeightOfACellInPixels: heightOfACellInPixels,
		SessionTypeIsX11:      sessionTypeIsX11,
	}

	// Symbol map from env override (or default)
	ci.SymbolMap = C.chafa_symbol_map_new()
	C.chafa_symbol_map_add_by_tags(ci.SymbolMap, getChafaSymbolTags())

	// Canvas config
	ci.Config = C.chafa_canvas_config_new()
	C.chafa_canvas_config_set_canvas_mode(ci.Config, ci.Mode)
	C.chafa_canvas_config_set_pixel_mode(ci.Config, ci.PixelMode)
	C.chafa_canvas_config_set_geometry(ci.Config, C.int(widthCells), C.int(heightCells))
	C.chafa_canvas_config_set_symbol_map(ci.Config, ci.SymbolMap)
	// chafa_canvas_config_set_optimizations(config, TRUE);
	C.chafa_canvas_config_set_work_factor(ci.Config, C.gfloat(0.0))
	// C.chafa_canvas_config_set_preprocessing_enabled(ci.Config, C.gboolean(0))
	// C.chafa_canvas_config_set_dither_intensity(ci.Config, C.CHAFA_DITHER_MODE_DIFFUSION)

	if widthOfACellInPixels > 0 && heightOfACellInPixels > 0 {
		/* We know the pixel dimensions of each cell. Store it in the config. */

		C.chafa_canvas_config_set_cell_geometry(ci.Config, C.int(widthOfACellInPixels), C.int(heightOfACellInPixels))
	}

	ci.Canvas = C.chafa_canvas_new(ci.Config)

	ci.PixelTypeOverride = getChafaPixelType()

	return ci
}

func (ci *ChafaInfo) getPixelType() C.ChafaPixelType {
	if ci.PixelTypeOverride != C.CHAFA_PIXEL_MAX {
		return ci.PixelTypeOverride
	}
	if ci.PixelMode == C.CHAFA_PIXEL_MODE_KITTY && !ci.SessionTypeIsX11 {
		return C.CHAFA_PIXEL_RGBA8_UNASSOCIATED
	}
	return C.CHAFA_PIXEL_BGRA8_UNASSOCIATED
}

func getChafaPixelType() C.ChafaPixelType {
	switch os.Getenv("TERM_EVERYTHING_PIXEL_TYPE") {
	case "":
		return C.CHAFA_PIXEL_MAX // No override
	case "RGBA8":
		return C.CHAFA_PIXEL_RGBA8_UNASSOCIATED
	case "BGRA8":
		return C.CHAFA_PIXEL_BGRA8_UNASSOCIATED
	case "ARGB8":
		return C.CHAFA_PIXEL_ARGB8_UNASSOCIATED
	case "ABGR8":
		return C.CHAFA_PIXEL_ABGR8_UNASSOCIATED
	case "RGBA8_PREMULTIPLIED":
		return C.CHAFA_PIXEL_RGBA8_PREMULTIPLIED
	case "BGRA8_PREMULTIPLIED":
		return C.CHAFA_PIXEL_BGRA8_PREMULTIPLIED
	case "ARGB8_PREMULTIPLIED":
		return C.CHAFA_PIXEL_ARGB8_PREMULTIPLIED
	case "ABGR8_PREMULTIPLIED":
		return C.CHAFA_PIXEL_ABGR8_PREMULTIPLIED
	default:
		return C.CHAFA_PIXEL_MAX
	}
}

var defaultSymbolTags = C.ChafaSymbolTags(C.CHAFA_SYMBOL_TAG_ALL)

func getChafaSymbolTags() C.ChafaSymbolTags {
	switch os.Getenv("TERM_EVERYTHING_SYMBOLS") {
	case "":
		return defaultSymbolTags
	case "NONE":
		return C.CHAFA_SYMBOL_TAG_NONE
	case "SPACE":
		return C.CHAFA_SYMBOL_TAG_SPACE
	case "SOLID":
		return C.CHAFA_SYMBOL_TAG_SOLID
	case "STIPPLE":
		return C.CHAFA_SYMBOL_TAG_STIPPLE
	case "BLOCK":
		return C.CHAFA_SYMBOL_TAG_BLOCK
	case "BORDER":
		return C.CHAFA_SYMBOL_TAG_BORDER
	case "DIAGONAL":
		return C.CHAFA_SYMBOL_TAG_DIAGONAL
	case "DOT":
		return C.CHAFA_SYMBOL_TAG_DOT
	case "QUAD":
		return C.CHAFA_SYMBOL_TAG_QUAD
	case "HHALF":
		return C.CHAFA_SYMBOL_TAG_HHALF
	case "VHALF":
		return C.CHAFA_SYMBOL_TAG_VHALF
	case "HALF":
		return C.CHAFA_SYMBOL_TAG_HALF
	case "INVERTED":
		return C.CHAFA_SYMBOL_TAG_INVERTED
	case "BRAILLE":
		return C.CHAFA_SYMBOL_TAG_BRAILLE
	case "TECHNICAL":
		return C.CHAFA_SYMBOL_TAG_TECHNICAL
	case "GEOMETRIC":
		return C.CHAFA_SYMBOL_TAG_GEOMETRIC
	case "ASCII":
		return C.CHAFA_SYMBOL_TAG_ASCII
	case "ALPHA":
		return C.CHAFA_SYMBOL_TAG_ALPHA
	case "DIGIT":
		return C.CHAFA_SYMBOL_TAG_DIGIT
	case "ALNUM":
		return C.CHAFA_SYMBOL_TAG_ALNUM
	case "NARROW":
		return C.CHAFA_SYMBOL_TAG_NARROW
	case "WIDE":
		return C.CHAFA_SYMBOL_TAG_WIDE
	case "AMBIGUOUS":
		return C.CHAFA_SYMBOL_TAG_AMBIGUOUS
	case "UGLY":
		return C.CHAFA_SYMBOL_TAG_UGLY
	case "LEGACY":
		return C.CHAFA_SYMBOL_TAG_LEGACY
	case "SEXTANT":
		return C.CHAFA_SYMBOL_TAG_SEXTANT
	case "WEDGE":
		return C.CHAFA_SYMBOL_TAG_WEDGE
	case "LATIN":
		return C.CHAFA_SYMBOL_TAG_LATIN
	case "IMPORTED":
		return C.CHAFA_SYMBOL_TAG_IMPORTED
	case "OCTANT":
		if C.chafa_have_octant() != 0 {
			return C.ChafaSymbolTags(C.chafa_have_octant())
		}
		return defaultSymbolTags
	case "ALL":
		return C.CHAFA_SYMBOL_TAG_ALL
	default:
		return defaultSymbolTags
	}
}

func (ci *ChafaInfo) Destroy() {
	if ci == nil {
		return
	}
	if ci.Canvas != nil {
		C.chafa_canvas_unref(ci.Canvas)
		ci.Canvas = nil
	}
	if ci.Config != nil {
		C.chafa_canvas_config_unref(ci.Config)
		ci.Config = nil
	}
	if ci.SymbolMap != nil {
		C.chafa_symbol_map_unref(ci.SymbolMap)
		ci.SymbolMap = nil
	}
	if ci.TermInfo != nil {
		C.chafa_term_info_unref(ci.TermInfo)
		ci.TermInfo = nil
	}
}
