package wayland

import (
	"github.com/mmulet/term.everything/wayland/protocols"
)

type WlRegistryDelegateImpl struct{}

func (w *WlRegistryDelegateImpl) WlRegistry_bind(s protocols.ClientState, object_id protocols.ObjectID[protocols.WlRegistry], name uint32, idInterface string, idVersion uint32, idID protocols.AnyObjectID) {
	object := s.GetObject(protocols.AnyObjectID(name))
	s.AddObject(idID, object)
	version := protocols.Version(idVersion)

	switch name {
	case uint32(protocols.GlobalID_WlShm):
		s.AddGlobalWlShmBind(protocols.ObjectID[protocols.WlShm](idID), version)
	case uint32(protocols.GlobalID_WlSeat):
		s.AddGlobalWlSeatBind(protocols.ObjectID[protocols.WlSeat](idID), version)
	case uint32(protocols.GlobalID_WlOutput):
		s.AddGlobalWlOutputBind(protocols.ObjectID[protocols.WlOutput](idID), version)
	case uint32(protocols.GlobalID_WlKeyboard):
		s.AddGlobalWlKeyboardBind(protocols.ObjectID[protocols.WlKeyboard](idID), version)
	case uint32(protocols.GlobalID_WlPointer):
		s.AddGlobalWlPointerBind(protocols.ObjectID[protocols.WlPointer](idID), version)
	case uint32(protocols.GlobalID_WlTouch):
		s.AddGlobalWlTouchBind(protocols.ObjectID[protocols.WlTouch](idID), version)
	case uint32(protocols.GlobalID_WlDataDevice):
		s.AddGlobalWlDataDeviceBind(protocols.ObjectID[protocols.WlDataDevice](idID), version)
	case uint32(protocols.GlobalID_ZwpXwaylandKeyboardGrabManagerV1):
		s.AddGlobalZwpXwaylandKeyboardGrabManagerV1Bind(protocols.ObjectID[protocols.ZwpXwaylandKeyboardGrabManagerV1](idID), version)
		// const set = s.global_binds.get(name) ?? new Set();
		// set.add(id_id);
		// s.global_binds.set(name, set);
	}
	// TODO turn this back on
	//  if (wayland_debug_time_only()) {
	//     console.log(
	//       `client#${s.client_socket} ${id_interface}_on_bind(name:`,
	//       name,
	//       ",interface:",
	//       id_interface,
	//       ",id:",
	//       id_id,
	//       ",version:",
	//       id_version,
	//       `)`
	//     );
	//   }

	if object != nil {

		delegate := object.(protocols.WaylandObject[protocols.OnBindable]).Delegate()
		delegate.OnBind(
			s,
			protocols.AnyObjectID(name),
			idInterface,
			idID,
			idVersion,
		)
	}
}

func (w *WlRegistryDelegateImpl) OnBind(
	s protocols.ClientState,
	_ protocols.AnyObjectID,
	_ string,
	newId_any protocols.AnyObjectID,
	version uint32,
) {
}

func MakeWlRegistry() *protocols.WlRegistry {
	return &protocols.WlRegistry{
		Delegate: &WlRegistryDelegateImpl{},
	}
}
