package wayland

import (
	"fmt"

	"github.com/mmulet/term.everything/wayland/protocols"
)

type WlPointer struct {
	// Last cursor surface set via set_cursor
	PointerSurfaceID map[protocols.ClientState]*protocols.ObjectID[protocols.WlSurface]

	WindowX float32
	WindowY float32
}

func (p *WlPointer) WlPointer_set_cursor(
	s protocols.ClientState,
	object_id protocols.ObjectID[protocols.WlPointer],
	_ uint32, // serial - TODO: validate most recent serial if/when tracked
	surface_id *protocols.ObjectID[protocols.WlSurface],
	hotspot_x int32,
	hotspot_y int32,
) {
	/**
	 * @TODO look at the serial and see it if valid (you are only supposed
	 * to respond to the most recent serial)
	 */
	// if (serial <= this.last_pointer_enter_serial) {
	//   console.error("Ignoring set cursor for stale serial");
	//   return;
	// }

	pointerSurfaceID, ok := p.PointerSurfaceID[s]
	if ok && !AreSame(pointerSurfaceID, surface_id) {
		if oldPointerSurface := GetWlSurfaceObject(s, *pointerSurfaceID); oldPointerSurface != nil {
			oldPointerSurface.Texture = nil
			if oldPointerSurface.Role != nil {
				if _, isCursor := oldPointerSurface.Role.(*SurfaceRoleCursor); isCursor {
					oldPointerSurface.Role = nil
				}

			}
		}
	}

	p.PointerSurfaceID[s] = surface_id

	if surface_id == nil {
		return
	}

	surface := GetWlSurfaceObject(s, *surface_id)
	if surface == nil {
		fmt.Printf("Surface not found")
		return
	}

	_, isCursor := surface.Role.(*SurfaceRoleCursor)
	if surface.Role != nil && !isCursor {
		SendError(s,
			object_id,
			protocols.WlPointerError_enum_role,
			"Surface already has a role")
		fmt.Printf("Surface already has a role")
		return
	}

	surface.Role = &SurfaceRoleCursor{
		Data: &SurfaceRoleCursorData{
			Hotspot: CursorHotspot{
				X: hotspot_x,
				Y: hotspot_y,
			},
		},
	}

}

func (p *WlPointer) AfterGetPointer(_ protocols.ClientState, _ protocols.ObjectID[protocols.WlPointer]) {
	/** @TODO: Implement wl_pointer_set_cursor */
	/**
	 * @TODO probably pointer.enter
	 */
	// this.last_pointer_enter_serial += 1;
}

func (p *WlPointer) WlPointer_release(
	_ protocols.ClientState,
	_ protocols.ObjectID[protocols.WlPointer],
) bool {
	return true
}

func (p *WlPointer) OnBind(
	s protocols.ClientState,
	_ protocols.AnyObjectID,
	_ string,
	newId_any protocols.AnyObjectID,
	version uint32,
) {
	// No-op for now
}

func MakeWlPointer() *protocols.WlPointer {
	return &protocols.WlPointer{
		Delegate: &WlPointer{},
	}
}

var Pointer = WlPointer{}
