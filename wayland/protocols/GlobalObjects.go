package protocols

type GlobalID AnyObjectID
type Version uint32

const (
	GlobalID_WlDisplay                        GlobalID = 1
	GlobalID_WlCompositor                     GlobalID = 0xff00000
	GlobalID_WlSubcompositor                  GlobalID = 0xff00001
	GlobalID_WlOutput                         GlobalID = 0xff00002
	GlobalID_WlSeat                           GlobalID = 0xff00003
	GlobalID_WlShm                            GlobalID = 0xff00004
	GlobalID_XdgWmBase                        GlobalID = 0xff00005
	GlobalID_WlDataDeviceManager              GlobalID = 0xff00006
	GlobalID_WlKeyboard                       GlobalID = 0xff00007
	GlobalID_WlPointer                        GlobalID = 0xff00008
	GlobalID_ZwpXwaylandKeyboardGrabManagerV1 GlobalID = 0xff00009
	GlobalID_XwaylandShellV1                  GlobalID = 0xff00011
	GlobalID_WlDataDevice                     GlobalID = 0xff00012
	GlobalID_WlTouch                          GlobalID = 0xff00013
	GlobalID_ZxdgDecorationManagerV1          GlobalID = 0xff00014
)

type AdvertisedGlobalObjectName struct {
	Name    string
	Id      GlobalID
	Version uint32
}

var AdvertisedGlobalObjectNames = []AdvertisedGlobalObjectName{
	{"wl_compositor", GlobalID_WlCompositor, 6},
	/**
	 * Turning off the wl_subcompositor will turn off
	 * decorations. Any other side effects??? Looks like
	 * GameScope has it turned off, so maybe we could do that
	 * too.
	 *
	 * some programs will crash if wl_subcompositor is not
	 * advertised.
	 */
	{"wl_subcompositor", GlobalID_WlSubcompositor, 1},
	{"wl_output", GlobalID_WlOutput, 5},

	{"wl_seat", GlobalID_WlSeat, 10},
	{"wl_shm", GlobalID_WlShm, 2},
	{"xdg_wm_base", GlobalID_XdgWmBase, 6},
	{"wl_data_device_manager", GlobalID_WlDataDeviceManager, 3},
	{"zxdg_decoration_manager_v1", GlobalID_ZxdgDecorationManagerV1, 1},
	/**
	 * @TODO only advertise these to Xwayland clients
	 */
	{"zwp_xwayland_keyboard_grab_manager_v1", GlobalID_ZwpXwaylandKeyboardGrabManagerV1, 1},
	{"xwayland_shell_v1", GlobalID_XwaylandShellV1, 1},
}

func GetGlobalWlDisplayBinds(cs ClientState) map[ObjectID[WlDisplay]]Version {

	v := cs.GetGlobalBinds(GlobalID(GlobalID_WlDisplay))
	m := v.(map[ObjectID[WlDisplay]]Version)
	return m
}

func GetGlobalWlCompositorBinds(cs ClientState) map[ObjectID[WlCompositor]]Version {

	v := cs.GetGlobalBinds(GlobalID(GlobalID_WlCompositor))
	m := v.(map[ObjectID[WlCompositor]]Version)
	return m
}

func GetGlobalWlSubcompositorBinds(cs ClientState) map[ObjectID[WlSubcompositor]]Version {

	v := cs.GetGlobalBinds(GlobalID(GlobalID_WlSubcompositor))
	m := v.(map[ObjectID[WlSubcompositor]]Version)
	return m
}

func GetGlobalWlOutputBinds(cs ClientState) map[ObjectID[WlOutput]]Version {

	v := cs.GetGlobalBinds(GlobalID(GlobalID_WlOutput))
	m := v.(map[ObjectID[WlOutput]]Version)
	return m
}

func GetGlobalWlSeatBinds(cs ClientState) map[ObjectID[WlSeat]]Version {

	v := cs.GetGlobalBinds(GlobalID(GlobalID_WlSeat))
	m := v.(map[ObjectID[WlSeat]]Version)
	return m
}

func GetGlobalWlShmBinds(cs ClientState) map[ObjectID[WlShm]]Version {

	v := cs.GetGlobalBinds(GlobalID(GlobalID_WlShm))
	m := v.(map[ObjectID[WlShm]]Version)
	return m
}

func GetGlobalXdgWmBaseBinds(cs ClientState) map[ObjectID[XdgWmBase]]Version {

	v := cs.GetGlobalBinds(GlobalID(GlobalID_XdgWmBase))
	m := v.(map[ObjectID[XdgWmBase]]Version)
	return m
}

func GetGlobalWlDataDeviceManagerBinds(cs ClientState) map[ObjectID[WlDataDeviceManager]]Version {

	v := cs.GetGlobalBinds(GlobalID(GlobalID_WlDataDeviceManager))
	m := v.(map[ObjectID[WlDataDeviceManager]]Version)
	return m
}

func GetGlobalWlKeyboardBinds(cs ClientState) map[ObjectID[WlKeyboard]]Version {

	v := cs.GetGlobalBinds(GlobalID(GlobalID_WlKeyboard))
	m := v.(map[ObjectID[WlKeyboard]]Version)
	return m
}

func GetGlobalWlPointerBinds(cs ClientState) map[ObjectID[WlPointer]]Version {

	v := cs.GetGlobalBinds(GlobalID(GlobalID_WlPointer))
	m := v.(map[ObjectID[WlPointer]]Version)
	return m
}

func GetGlobalZwpXwaylandKeyboardGrabManagerV1Binds(cs ClientState) map[ObjectID[ZwpXwaylandKeyboardGrabManagerV1]]Version {

	v := cs.GetGlobalBinds(GlobalID(GlobalID_ZwpXwaylandKeyboardGrabManagerV1))
	m := v.(map[ObjectID[ZwpXwaylandKeyboardGrabManagerV1]]Version)
	return m
}

func GetGlobalXwaylandShellV1Binds(cs ClientState) map[ObjectID[XwaylandShellV1]]Version {

	v := cs.GetGlobalBinds(GlobalID(GlobalID_XwaylandShellV1))
	m := v.(map[ObjectID[XwaylandShellV1]]Version)
	return m
}

func GetGlobalWlDataDeviceBinds(cs ClientState) map[ObjectID[WlDataDevice]]Version {

	v := cs.GetGlobalBinds(GlobalID(GlobalID_WlDataDevice))
	m := v.(map[ObjectID[WlDataDevice]]Version)
	return m
}

func GetGlobalWlTouchBinds(cs ClientState) map[ObjectID[WlTouch]]Version {

	v := cs.GetGlobalBinds(GlobalID(GlobalID_WlTouch))
	m := v.(map[ObjectID[WlTouch]]Version)
	return m
}

func GetGlobalZxdgDecorationManagerV1Binds(cs ClientState) map[ObjectID[ZxdgDecorationManagerV1]]Version {

	v := cs.GetGlobalBinds(GlobalID(GlobalID_ZxdgDecorationManagerV1))
	m := v.(map[ObjectID[ZxdgDecorationManagerV1]]Version)
	return m
}
