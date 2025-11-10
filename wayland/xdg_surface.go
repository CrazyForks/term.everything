package wayland

import (
	"fmt"
	"time"

	"github.com/mmulet/term.everything/wayland/protocols"
)

type XdgWindowGeometry struct {
	X      int32
	Y      int32
	Width  int32
	Height int32
}

type XdgSurface struct {
	Version        uint32
	XdgSurfaceID   protocols.ObjectID[protocols.XdgSurface]
	OnConfigure    map[uint32]chan uint32
	LatestSerial   uint32
	WindowGeometry XdgWindowGeometry
}

var GlobalEnterSerial uint32 = 0

// Sends a configure event and waits for the client to ack it. Blocks until ack received.
func (x *XdgSurface) configure(s protocols.ClientState) {
	serial := x.LatestSerial
	x.LatestSerial++

	configureChannel := make(chan uint32, 1)
	x.OnConfigure[serial] = configureChannel

	protocols.XdgSurface_configure(s, x.XdgSurfaceID, serial)

	<-configureChannel

}

/**
 * xdg_surface methods
 */

func (x *XdgSurface) XdgSurface_destroy(
	s protocols.ClientState,
	objectID protocols.ObjectID[protocols.XdgSurface],
) bool {

	if surface_id := GetSurfaceIDFromRole(s, objectID); surface_id != nil {

		if ws := GetWlSurfaceObject(s, *surface_id); ws != nil && ws.HasRoleData() {
			fmt.Printf("xdg_surface Destroying surface before role is destroyed")
			SendError(
				s,
				objectID,
				protocols.XdgSurfaceError_enum_defunct_role_object,
				"Surface destroyed before role",
			)
		}
	}

	UnregisterRoleToSurface(s, objectID)
	return true
}

func (x *XdgSurface) XdgSurface_get_toplevel(
	s protocols.ClientState,
	xdgSurfaceObjectID protocols.ObjectID[protocols.XdgSurface],
	id protocols.ObjectID[protocols.XdgToplevel],
) {
	surface_id := GetSurfaceIDFromRole(s, xdgSurfaceObjectID)
	if surface_id == nil {
		SendError(s,
			xdgSurfaceObjectID,
			protocols.XdgSurfaceError_enum_not_constructed,
			"surface not found",
		)
		return
	}

	surface := GetSurfaceFromRole(s, xdgSurfaceObjectID)
	if surface == nil {
		SendError(s,
			xdgSurfaceObjectID,
			protocols.XdgSurfaceError_enum_not_constructed,
			"surface not found",
		)
		return
	}

	if surface.Role == nil {
		surface.Role = &SurfaceRoleXdgToplevel{}
	}

	surfaceRole, surface_role_is_xdg_toplevel := surface.Role.(*SurfaceRoleXdgToplevel)
	if !surface_role_is_xdg_toplevel {
		SendError(s,
			xdgSurfaceObjectID,
			protocols.XdgSurfaceError_enum_defunct_role_object,
			"surface not found",
		)
		return
	}

	if surfaceRole.Data != nil {
		SendError(s,
			xdgSurfaceObjectID,
			protocols.XdgSurfaceError_enum_defunct_role_object,
			"surface already has a role",
		)
		return
	}

	surfaceRole.Data = &id
	AddObject(s, id, MakeXdgToplevel())

	RegisterRoleToSurface(s, id, *surface_id)
	s.TopLevelSurfaces()[id] = true

	protocols.XdgToplevel_configure(
		s,
		id,
		int32(VirtualMonitorSize.Width),
		int32(VirtualMonitorSize.Height),
		ToBytes([]protocols.XdgToplevelState_enum{
			protocols.XdgToplevelState_enum_maximized,
			protocols.XdgToplevelState_enum_fullscreen,
		}),
	)

	// TODO this is necessary, but when?
	protocols.XdgSurface_configure(s, xdgSurfaceObjectID, 0)

	// TODO should this be here

	if output_binds := protocols.GetGlobalWlOutputBinds(s); output_binds != nil {
		for output_id, _ := range output_binds {
			protocols.WlSurface_enter(s, *surface_id, output_id)
		}
	}
	serial := GlobalEnterSerial
	GlobalEnterSerial += 1

	if keyboard_binds := protocols.GetGlobalWlKeyboardBinds(s); keyboard_binds != nil {

		for keyboard_id, _ := range keyboard_binds {
			protocols.WlKeyboard_enter(s, keyboard_id, serial, *surface_id, []byte{})
		}
	}

	if pointer_binds := protocols.GetGlobalWlPointerBinds(s); pointer_binds != nil {
		for pointer_id, version := range pointer_binds {
			protocols.WlPointer_enter(s, pointer_id, serial, *surface_id, Pointer.WindowX, Pointer.WindowY)
			protocols.WlPointer_frame(s, uint32(version), pointer_id)
		}
	}

	/**
	 * commented because it
	 * causes firefox to crash with:
	 * "GraphicsCriticalError": "|[0][GFX1-]: (ubuntu:gnome) Wayland protocol error: listener function for opcode 2 of xdg_toplevel is NULL\n (t=0.503382) ",
	 */
	// xdg_toplevel_funcs.configure_bounds(
	//   s,
	//   this.version,
	//   id,
	//   virtual_monitor_size.width,
	//   virtual_monitor_size.height
	// );
	/**
	 * So with Xwayland clients, pointer enter doesn't stick?
	 *
	 * So let's do it again
	 * after a timeout
	 * @TODO where to actually put this?
	 */

	go func() {
		time.Sleep(100 * time.Millisecond)
		if pointer_binds := protocols.GetGlobalWlPointerBinds(s); pointer_binds != nil {
			for pointer_id, _ := range pointer_binds {
				pointer := GetWlPointerObject(s, pointer_id)
				if pointer == nil {
					continue
				}
				protocols.WlPointer_enter(s, pointer_id, 0, *surface_id, pointer.WindowX, pointer.WindowY)
			}
		}
	}()

}

func (x *XdgSurface) XdgSurface_get_popup(
	s protocols.ClientState,
	xdgSurfaceObjectID protocols.ObjectID[protocols.XdgSurface],
	id protocols.ObjectID[protocols.XdgPopup],
	parent *protocols.ObjectID[protocols.XdgSurface],
	positionerID protocols.ObjectID[protocols.XdgPositioner],
) {
	surface_id := GetSurfaceIDFromRole(s, xdgSurfaceObjectID)
	if surface_id == nil {
		SendError(
			s,
			xdgSurfaceObjectID,
			protocols.XdgSurfaceError_enum_not_constructed,
			"surface not found",
		)
		return
	}

	surface := GetSurfaceFromRole(s, xdgSurfaceObjectID)
	if surface == nil {
		SendError(
			s,
			xdgSurfaceObjectID,
			protocols.XdgSurfaceError_enum_not_constructed,
			"surface not found",
		)
		return
	}

	if surface.Role == nil {
		surface.Role = &SurfaceRoleXdgPopup{}
	}

	surfaceRole, surface_role_is_xdg_popup := surface.Role.(*SurfaceRoleXdgPopup)
	if !surface_role_is_xdg_popup {
		SendError(
			s,
			xdgSurfaceObjectID,
			protocols.XdgSurfaceError_enum_defunct_role_object,
			"surface role is not xdg_popup",
		)
		return
	}
	positioner := GetXdgPositionerObject(s, positionerID)
	if positioner == nil {
		SendError(
			s,
			xdgSurfaceObjectID,
			protocols.XdgSurfaceError_enum_not_constructed,
			"positioner not found",
		)
		return
	}
	surfaceRole.Data = &id

	AddObject(s, id, MakeXdgPopup(x.Version, parent, positioner.state))

	RegisterRoleToSurface(s, id, *surface_id)

	protocols.XdgPopup_configure(
		s,
		id,
		0, 0,
		int32(VirtualMonitorSize.Width),
		int32(VirtualMonitorSize.Height),
	)
}

func (x *XdgSurface) XdgSurface_set_window_geometry(
	s protocols.ClientState,
	objectID protocols.ObjectID[protocols.XdgSurface],
	wx int32,
	wy int32,
	ww int32,
	wh int32,
) {

	surface := GetSurfaceFromRole(s, objectID)
	if surface == nil {
		return
	}

	surface.PendingUpdate.XdgSurfaceWindowGeometry = &XdgWindowGeometry{
		X:      wx,
		Y:      wy,
		Width:  ww,
		Height: wh,
	}
}

func (x *XdgSurface) XdgSurface_ack_configure(
	_ protocols.ClientState,
	_ protocols.ObjectID[protocols.XdgSurface],
	serial uint32,
) {
	if x.OnConfigure == nil {
		return
	}
	if cb, ok := x.OnConfigure[serial]; ok && cb != nil {
		cb <- serial
	}
	for k := range x.OnConfigure {
		if k > serial {
			continue
		}
		//TODO do I need to close the channel?
		if cb, ok := x.OnConfigure[k]; ok && cb != nil {
			close(cb)
		}
		delete(x.OnConfigure, k)
	}
}

func (x *XdgSurface) OnBind(
	cs protocols.ClientState,
	_ protocols.AnyObjectID,
	_ string,
	newId_any protocols.AnyObjectID,
	version uint32,
) {
}

func MakeXdgSurface(version uint32, xdg_surface_id protocols.ObjectID[protocols.XdgSurface]) *protocols.XdgSurface {
	return &protocols.XdgSurface{
		Delegate: &XdgSurface{
			Version:      version,
			XdgSurfaceID: xdg_surface_id,
			OnConfigure:  make(map[uint32]chan uint32),
		},
	}
}
