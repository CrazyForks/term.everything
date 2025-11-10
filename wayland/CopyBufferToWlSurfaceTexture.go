package wayland

import (
	"fmt"

	"github.com/mmulet/term.everything/wayland/protocols"
)

func CopyBufferToWlSurfaceTexture(
	s protocols.ClientState,
	surfaceID protocols.ObjectID[protocols.WlSurface],
	zIndex int,
	maybebufferID *protocols.ObjectID[protocols.WlBuffer],
) {
	surface := GetWlSurfaceObject(s, surfaceID)
	if maybebufferID == nil {
		delete(s.DrawableSurfaces(), surfaceID)
		/**
		 * Time to remove the texture from the surface
		 */
		if surface == nil {
			return
		}
		surface.Texture = nil
		return
	}
	bufferId := *maybebufferID
	if surface == nil {
		return
	}

	pool := GetWlPoolObject_FromBuffer(s, bufferId)
	if pool == nil {
		fmt.Println("Could not get pool delegate; can't commit")
		return
	}

	if pool.MapState == MapStateDestroyed {
		fmt.Printf("Could not get pool.buffer_pointer; can't commit! pool %d buffer %d\n",
			pool.WlShmPoolObjectID, bufferId)
		return
	}

	bufferInfo, ok := pool.Buffers[bufferId]
	if !ok {
		fmt.Println("Could not get buffer_info; can't commit")
		return
	}

	x := surface.Offset.X
	y := surface.Offset.Y

	if surface.Role == nil {
		return
	}

	switch role := surface.Role.(type) {
	case *SurfaceRoleXdgPopup:
		return
	case *SurfaceRoleSubSurface:
		if role.Data != nil {
			sub_surface := GetWlSubsurfaceObject(s, *role.Data)
			if sub_surface != nil {
				x += sub_surface.Position.X
				y += sub_surface.Position.Y
			}
			/**
			 * @TODO should this be relative to the parent?
			 */
		}
	case *SurfaceRoleXWaylandSurface:
		/**
		 * @TODO
		 */
		fmt.Println("ON commit xwayland_surface_v1")
	case *SurfaceRoleXdgToplevel:
		if surface.XdgSurfaceState != nil {
			xdg_surface_state := GetXdgSurfaceObject(s, *surface.XdgSurfaceState)
			if xdg_surface_state != nil {
				// x = surface.xdg_surface_state.window_geometry.x;
				// y = surface.xdg_surface_state.window_geometry.y;
				// console.log(
				//   "reposition xdg_toplevel, ",
				//   x,
				//   y,
				//   "for surface",
				//   surface_id
				// );
			}
		}
	case *SurfaceRoleCursor:
		/**
		 * @TODO is this right?
		 */
		if !role.HasData() {
			/**
			 * From the docs:
			 * When the use as a cursor ends, the wl_surface is unmapped
			 *
			 * So I think that means if it isn't a cursor anymore,
			 * we should not draw it
			 */
			return
		}
		x += int32(Pointer.WindowX) + role.Data.Hotspot.X
		y += int32(Pointer.WindowY) + role.Data.Hotspot.Y

	}
	surface.Position.X = x
	surface.Position.Y = y
	surface.Position.Z = int32(zIndex)

	if surface.Texture != nil {
		if surface.Texture.Stride != uint32(bufferInfo.Stride) ||
			surface.Texture.Width != uint32(bufferInfo.Width) ||
			surface.Texture.Height != uint32(bufferInfo.Height) {
			surface.Texture = nil
		}
	}

	if surface.Texture == nil {
		size := int(bufferInfo.Stride) * int(bufferInfo.Height)
		if size < 0 {
			fmt.Println("Invalid buffer size; can't commit")
			return
		}
		surface.Texture = &Texture{
			Stride: uint32(bufferInfo.Stride),
			Width:  uint32(bufferInfo.Width),
			Height: uint32(bufferInfo.Height),
			Data:   make([]byte, size),
		}
	}

	memMap, ok := pool.MemMaps[pool.WlShmPoolObjectID]
	if !ok {
		fmt.Println("No memmap for pool; can't commit")
		return
	}

	total := int(bufferInfo.Stride) * int(bufferInfo.Height)
	if total < 0 || total > len(surface.Texture.Data) {
		fmt.Println("Computed copy size out of bounds; can't commit")
		return
	}

	offset := int(bufferInfo.Offset)

	src := memMap.Bytes
	if offset < 0 || offset+total > len(src) {
		fmt.Println("Pool memory bounds error during copy; can't commit")
		return
	}

	copy(surface.Texture.Data, src[offset:offset+total])

	s.DrawableSurfaces()[surfaceID] = true
}
