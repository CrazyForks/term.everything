package wayland

import (
	"github.com/mmulet/term.everything/wayland/protocols"
)

type ZwpXwaylandKeyboardGrabManagerV1 struct {
}

func (m *ZwpXwaylandKeyboardGrabManagerV1) ZwpXwaylandKeyboardGrabManagerV1_destroy(
	s protocols.ClientState,
	object_id protocols.ObjectID[protocols.ZwpXwaylandKeyboardGrabManagerV1],
) bool {
	s.RemoveGlobalZwpXwaylandKeyboardGrabManagerV1Bind(object_id)
	return true
}

func (m *ZwpXwaylandKeyboardGrabManagerV1) ZwpXwaylandKeyboardGrabManagerV1_grab_keyboard(
	s protocols.ClientState,
	_object_id protocols.ObjectID[protocols.ZwpXwaylandKeyboardGrabManagerV1],
	id protocols.ObjectID[protocols.ZwpXwaylandKeyboardGrabV1],
	_surface protocols.ObjectID[protocols.WlSurface],
	_seat protocols.ObjectID[protocols.WlSeat],
) {
	AddObject(s, id, MakeZwpXwaylandKeyboardGrabV1())
	/**
	 * @TODO grab the keyboard
	 */
}

func (m *ZwpXwaylandKeyboardGrabManagerV1) OnBind(
	_s protocols.ClientState,
	_name protocols.AnyObjectID,
	_interface_ string,
	_new_id protocols.AnyObjectID,
	_version_number uint32,
) {
}

func MakeZwpXwaylandKeyboardGrabManagerV1() *protocols.ZwpXwaylandKeyboardGrabManagerV1 {
	return &protocols.ZwpXwaylandKeyboardGrabManagerV1{
		Delegate: &ZwpXwaylandKeyboardGrabManagerV1{},
	}
}
