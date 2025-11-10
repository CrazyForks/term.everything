// Assuming surface.Role has a Data field that can be set to nil
package wayland

import (
	"fmt"

	"github.com/mmulet/term.everything/wayland/protocols"
)

type XwaylandShellV1 struct {
}

func (x *XwaylandShellV1) XwaylandShellV1_destroy(
	s protocols.ClientState,
	object_id protocols.ObjectID[protocols.XwaylandShellV1],
) bool {
	return true
}

func (x *XwaylandShellV1) XwaylandShellV1_get_xwayland_surface(
	s protocols.ClientState,
	object_id protocols.ObjectID[protocols.XwaylandShellV1],
	id protocols.ObjectID[protocols.XwaylandSurfaceV1],
	surface_id protocols.ObjectID[protocols.WlSurface],
) {
	surface := GetWlSurfaceObject(s, surface_id)
	if surface == nil {
		fmt.Printf("xwayland_shell_v1_get_xwayland_surface: surface not found\n")
		SendError(
			s,
			object_id,
			protocols.XwaylandShellV1Error_enum_role,
			"surface not found",
		)
		return
	}

	if surface.Role == nil {
		surface.Role = &SurfaceRoleXWaylandSurface{}
	}

	surfaceRole, ok := surface.Role.(*SurfaceRoleXWaylandSurface)
	if !ok {
		fmt.Printf("xwayland_shell_v1_get_xwayland_surface: surface has wrong role type\n")
		SendError(
			s,
			object_id,
			protocols.XwaylandShellV1Error_enum_role,
			"surface already has a role",
		)
		return
	}

	xwayland_surface := surfaceRole.Data
	if xwayland_surface != nil {
		SendError(
			s,
			object_id,
			protocols.XwaylandShellV1Error_enum_role,
			"surface already has a role",
		)
		return
	}
	AddObject(s, id, MakeXwaylandSurfaceV1())
}

func (x *XwaylandShellV1) OnBind(
	_s protocols.ClientState,
	_name protocols.AnyObjectID,
	_interface_ string,
	_new_id protocols.AnyObjectID,
	_version_number uint32,
) {
}

func MakeXwaylandShellV1() *protocols.XwaylandShellV1 {
	return &protocols.XwaylandShellV1{
		Delegate: &XwaylandShellV1{},
	}
}
