package wayland

import (
	"github.com/mmulet/term.everything/wayland/protocols"
)

type ZxdgDecorationManagerV1 struct{}

func (z *ZxdgDecorationManagerV1) ZxdgDecorationManagerV1_destroy(
	s protocols.ClientState,
	object_id protocols.ObjectID[protocols.ZxdgDecorationManagerV1],
) bool {
	return true
}

func (z *ZxdgDecorationManagerV1) ZxdgDecorationManagerV1_get_toplevel_decoration(
	s protocols.ClientState,
	_ protocols.ObjectID[protocols.ZxdgDecorationManagerV1],
	decoration_id protocols.ObjectID[protocols.ZxdgToplevelDecorationV1],
	toplevel protocols.ObjectID[protocols.XdgToplevel],
) {
	AddObject(s, decoration_id, MakeZxdgToplevelDecorationV1(toplevel))
	protocols.ZxdgToplevelDecorationV1_configure(s, decoration_id, protocols.ZxdgToplevelDecorationV1Mode_enum_server_side)
}

func (z *ZxdgDecorationManagerV1) OnBind(
	_ protocols.ClientState,
	_ protocols.AnyObjectID,
	_ string,
	_ protocols.AnyObjectID,
	_ uint32,
) {
	// no-op
}

func MakeZxdgDecorationManagerV1() *protocols.ZxdgDecorationManagerV1 {
	return &protocols.ZxdgDecorationManagerV1{
		Delegate: &ZxdgDecorationManagerV1{},
	}
}
