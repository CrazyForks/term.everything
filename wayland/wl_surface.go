package wayland

import (
	"fmt"
	"image"

	"github.com/mmulet/term.everything/wayland/protocols"
)

type Texture struct {
	Stride uint32
	Width  uint32
	Height uint32
	Data   []byte
}

func (t *Texture) AsRGBA() *image.RGBA {
	if t == nil || t.Width == 0 || t.Height == 0 || len(t.Data) == 0 {
		return nil
	}
	return &image.RGBA{
		Pix:    t.Data,
		Stride: int(t.Stride),
		Rect:   image.Rect(0, 0, int(t.Width), int(t.Height)),
	}
}

type WlSurface struct {
	Position struct {
		X int32
		Y int32
		Z int32
	}

	Texture *Texture
	/**
	 * xdg_surface is not a role,
	 * but have to keep track anyway.
	 */
	XdgSurfaceState *protocols.ObjectID[protocols.XdgSurface]

	/**
	 * nil represents to draw the current surface.
	 * By index, 0 is the bottom, and the last index is the top.
	 */
	ChildrenInDrawOrder []*protocols.ObjectID[protocols.WlSurface]

	Role SurfaceRole

	BufferTransform protocols.WlOutputTransform_enum
	BufferScale     int32

	/**
	 * Null means infinite, (ie we can accept input from everywhere)
	 */
	InputRegion *protocols.ObjectID[protocols.WlRegion]
	/**
	 * Unlink opaque region, null means empty!
	 */
	OpaqueRegion *protocols.ObjectID[protocols.WlRegion]

	PendingUpdate SurfaceUpdate

	Offset Point

	/**
	 * Don't care about regions for now.
	 * Just need to clear this when the surface has
	 * been drawn.
	 */
	Damaged bool
}

func (w *WlSurface) ClearRoleData() {
	if w.Role == nil {
		return
	}
	w.Role.ClearData()
}

func (w *WlSurface) HasRoleData() bool {
	if w.Role == nil {
		return false
	}
	return w.Role.HasData()
}

// destroy_texture = (s: Wayland_Client, surface_id: Object_ID<w>) => {
//   // if (!this.texture) {
//   //   return;
//   // }

//   delete s.texture_from_surface_id[surface_id];
//   // cpp.destroy_texture_for_wl_surface(s.client_state, surface_id);
//   // this.texture = null;
// };

/**
 *
 * Below are the wl_surface_delegate methods
 */

func (w *WlSurface) WlSurface_destroy(
	s protocols.ClientState,
	object_id protocols.ObjectID[protocols.WlSurface],
) bool {

	// this.destroy_texture(s, object_id);

	if !w.HasRoleData() {
		return true
	}

	SendError(s, object_id, protocols.WlSurfaceError_enum_defunct_role_object, "Surface destroyed before role")
	fmt.Printf("wl_surface@%d Destroying surface before role is destroyed", object_id)

	return true
}

func (w *WlSurface) WlSurface_attach(
	s protocols.ClientState,
	object_id protocols.ObjectID[protocols.WlSurface],
	buffer *protocols.ObjectID[protocols.WlBuffer],
	x int32,
	y int32,
) {
	w.PendingUpdate.Buffer = buffer
	if buffer != nil && *buffer == 0 {
		w.PendingUpdate.Buffer = nil
	}

	if s.GetCompositorVersion() < 5 {
		w.Offset = Point{X: x, Y: y}
		return
	}
	if x == 0 && y == 0 {
		return
	}

	SendError(
		s,
		object_id,
		protocols.WlSurfaceError_enum_invalid_offset,
		"x and y must be 0 if version >= 5",
	)
}

func (w *WlSurface) WlSurface_damage(
	_ protocols.ClientState,
	_ protocols.ObjectID[protocols.WlSurface],
	x int32,
	y int32,
	width int32,
	height int32,
) {
	if w.PendingUpdate.Damage == nil {
		w.PendingUpdate.Damage = []Rect{}
	}
	w.PendingUpdate.Damage = append(w.PendingUpdate.Damage, Rect{X: x, Y: y, Width: width, Height: height})
}

func (w *WlSurface) WlSurface_frame(
	s protocols.ClientState,
	_ protocols.ObjectID[protocols.WlSurface],
	callback protocols.ObjectID[protocols.WlCallback],
) {
	s.AddFrameDrawRequest(callback)
}

func (w *WlSurface) WlSurface_set_opaque_region(
	_ protocols.ClientState,
	_ protocols.ObjectID[protocols.WlSurface],
	region *protocols.ObjectID[protocols.WlRegion],
) {
	w.PendingUpdate.OpaqueRegion = region
}

func (w *WlSurface) WlSurface_set_input_region(
	_ protocols.ClientState,
	_ protocols.ObjectID[protocols.WlSurface],
	region *protocols.ObjectID[protocols.WlRegion],
) {
	w.PendingUpdate.InputRegion = region
}

func (w *WlSurface) WlSurface_commit(
	s protocols.ClientState,
	object_id protocols.ObjectID[protocols.WlSurface],
) {
	pendingBufferTextureUpdates := []PendingBufferUpdates{}
	pendingBufferTextureUpdates = ApplyWlSurfaceDoubleBufferedState(s, object_id, false, pendingBufferTextureUpdates, 0)

	for _, upd := range pendingBufferTextureUpdates {
		CopyBufferToWlSurfaceTexture(s, upd.Surface, upd.ZIndex, upd.Buffer)
	}

	for _, upd := range pendingBufferTextureUpdates {
		/**
		 * @TODO Is there every an occasion where the buffer would
		 * be used more than once, ie can we always release it here?
		 */
		// After consumption, tell client its buffer can be re-used
		if upd.Buffer != nil {
			protocols.WlBuffer_release(s, *upd.Buffer)
		}
	}
}

func (w *WlSurface) WlSurface_set_buffer_transform(
	_ protocols.ClientState,
	_ protocols.ObjectID[protocols.WlSurface],
	transform int32,
) {
	t := protocols.WlOutputTransform_enum(transform)
	w.PendingUpdate.BufferTransform = &t
}

func (w *WlSurface) WlSurface_set_buffer_scale(
	_ protocols.ClientState,
	_ protocols.ObjectID[protocols.WlSurface],
	scale int32,
) {
	w.PendingUpdate.BufferScale = &scale
}

func (w *WlSurface) WlSurface_damage_buffer(
	_ protocols.ClientState,
	_ protocols.ObjectID[protocols.WlSurface],
	x int32,
	y int32,
	width int32,
	height int32,
) {
	if w.PendingUpdate.DamageBuffer == nil {
		w.PendingUpdate.DamageBuffer = []Rect{}
	}
	w.PendingUpdate.DamageBuffer = append(w.PendingUpdate.DamageBuffer, Rect{X: x, Y: y, Width: width, Height: height})
}

func (w *WlSurface) WlSurface_offset(
	_ protocols.ClientState,
	_ protocols.ObjectID[protocols.WlSurface],
	x int32,
	y int32,
) {
	w.PendingUpdate.Offset = &Point{X: x, Y: y}
}

func (w *WlSurface) OnBind(
	cs protocols.ClientState,
	_ protocols.AnyObjectID,
	_ string,
	newId_any protocols.AnyObjectID,
	version uint32,
) {
}

func (w *WlSurface) ResetPendingUpdate() {
	w.PendingUpdate = SurfaceUpdate{}
}

// Constructor
func MakeWlSurface() *protocols.WlSurface {
	ws := &WlSurface{
		BufferScale:         1,
		ChildrenInDrawOrder: []*protocols.ObjectID[protocols.WlSurface]{nil},
		Position:            struct{ X, Y, Z int32 }{X: 0, Y: 0, Z: 0},
		XdgSurfaceState:     nil,
		InputRegion:         nil,
		OpaqueRegion:        nil,
		BufferTransform:     protocols.WlOutputTransform_enum_normal, // default transform can be set by client; nil means "normal" until set
	}
	return &protocols.WlSurface{
		Delegate: ws,
	}
}
