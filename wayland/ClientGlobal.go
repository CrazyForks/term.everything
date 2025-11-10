package wayland

import "github.com/mmulet/term.everything/wayland/protocols"

func (c *Client) AddGlobalWlShmBind(objectID protocols.ObjectID[protocols.WlShm], version protocols.Version) {
	binds, ok := c.GlobalBinds[protocols.GlobalID_WlShm]
	if !ok {
		binds = make(map[protocols.ObjectID[protocols.WlShm]]protocols.Version)
		c.GlobalBinds[protocols.GlobalID_WlShm] = binds
	}
	binds.(map[protocols.ObjectID[protocols.WlShm]]protocols.Version)[objectID] = version
}

func (c *Client) AddGlobalWlSeatBind(objectID protocols.ObjectID[protocols.WlSeat], version protocols.Version) {
	binds, ok := c.GlobalBinds[protocols.GlobalID_WlSeat]
	if !ok {
		binds = make(map[protocols.ObjectID[protocols.WlSeat]]protocols.Version)
		c.GlobalBinds[protocols.GlobalID_WlSeat] = binds
	}
	binds.(map[protocols.ObjectID[protocols.WlSeat]]protocols.Version)[objectID] = version
}

func (c *Client) AddGlobalWlOutputBind(objectID protocols.ObjectID[protocols.WlOutput], version protocols.Version) {
	binds, ok := c.GlobalBinds[protocols.GlobalID_WlOutput]
	if !ok {
		binds = make(map[protocols.ObjectID[protocols.WlOutput]]protocols.Version)
		c.GlobalBinds[protocols.GlobalID_WlOutput] = binds
	}
	binds.(map[protocols.ObjectID[protocols.WlOutput]]protocols.Version)[objectID] = version
}

func (c *Client) AddGlobalWlKeyboardBind(objectID protocols.ObjectID[protocols.WlKeyboard], version protocols.Version) {
	binds, ok := c.GlobalBinds[protocols.GlobalID_WlKeyboard]
	if !ok {
		binds = make(map[protocols.ObjectID[protocols.WlKeyboard]]protocols.Version)
		c.GlobalBinds[protocols.GlobalID_WlKeyboard] = binds
	}
	binds.(map[protocols.ObjectID[protocols.WlKeyboard]]protocols.Version)[objectID] = version
}

func (c *Client) AddGlobalWlPointerBind(objectID protocols.ObjectID[protocols.WlPointer], version protocols.Version) {
	binds, ok := c.GlobalBinds[protocols.GlobalID_WlPointer]
	if !ok {
		binds = make(map[protocols.ObjectID[protocols.WlPointer]]protocols.Version)
		c.GlobalBinds[protocols.GlobalID_WlPointer] = binds
	}
	binds.(map[protocols.ObjectID[protocols.WlPointer]]protocols.Version)[objectID] = version
}

func (c *Client) AddGlobalWlTouchBind(objectID protocols.ObjectID[protocols.WlTouch], version protocols.Version) {
	binds, ok := c.GlobalBinds[protocols.GlobalID_WlTouch]
	if !ok {
		binds = make(map[protocols.ObjectID[protocols.WlTouch]]protocols.Version)
		c.GlobalBinds[protocols.GlobalID_WlTouch] = binds
	}
	binds.(map[protocols.ObjectID[protocols.WlTouch]]protocols.Version)[objectID] = version
}

func (c *Client) AddGlobalWlDataDeviceBind(objectID protocols.ObjectID[protocols.WlDataDevice], version protocols.Version) {
	binds, ok := c.GlobalBinds[protocols.GlobalID_WlDataDevice]
	if !ok {
		binds = make(map[protocols.ObjectID[protocols.WlDataDevice]]protocols.Version)
		c.GlobalBinds[protocols.GlobalID_WlDataDevice] = binds
	}
	binds.(map[protocols.ObjectID[protocols.WlDataDevice]]protocols.Version)[objectID] = version
}

func (c *Client) AddGlobalZwpXwaylandKeyboardGrabManagerV1Bind(objectID protocols.ObjectID[protocols.ZwpXwaylandKeyboardGrabManagerV1], version protocols.Version) {
	binds, ok := c.GlobalBinds[protocols.GlobalID_ZwpXwaylandKeyboardGrabManagerV1]
	if !ok {
		binds = make(map[protocols.ObjectID[protocols.ZwpXwaylandKeyboardGrabManagerV1]]protocols.Version)
		c.GlobalBinds[protocols.GlobalID_ZwpXwaylandKeyboardGrabManagerV1] = binds
	}
	binds.(map[protocols.ObjectID[protocols.ZwpXwaylandKeyboardGrabManagerV1]]protocols.Version)[objectID] = version
}

func (c *Client) RemoveGlobalWlShmBind(objectID protocols.ObjectID[protocols.WlShm]) {
	binds, ok := c.GlobalBinds[protocols.GlobalID_WlShm]
	if !ok {
		return
	}
	delete(binds.(map[protocols.ObjectID[protocols.WlShm]]protocols.Version), objectID)
}

func (c *Client) RemoveGlobalWlSeatBind(objectID protocols.ObjectID[protocols.WlSeat]) {
	binds, ok := c.GlobalBinds[protocols.GlobalID_WlSeat]
	if !ok {
		return
	}
	delete(binds.(map[protocols.ObjectID[protocols.WlSeat]]protocols.Version), objectID)
}

func (c *Client) RemoveGlobalWlOutputBind(objectID protocols.ObjectID[protocols.WlOutput]) {
	binds, ok := c.GlobalBinds[protocols.GlobalID_WlOutput]
	if !ok {
		return
	}
	delete(binds.(map[protocols.ObjectID[protocols.WlOutput]]protocols.Version), objectID)
}
func (c *Client) RemoveGlobalWlKeyboardBind(objectID protocols.ObjectID[protocols.WlKeyboard]) {
	binds, ok := c.GlobalBinds[protocols.GlobalID_WlKeyboard]
	if !ok {
		return
	}
	delete(binds.(map[protocols.ObjectID[protocols.WlKeyboard]]protocols.Version), objectID)
}

func (c *Client) RemoveGlobalWlPointerBind(objectID protocols.ObjectID[protocols.WlPointer]) {
	binds, ok := c.GlobalBinds[protocols.GlobalID_WlPointer]
	if !ok {
		return
	}
	delete(binds.(map[protocols.ObjectID[protocols.WlPointer]]protocols.Version), objectID)
}

func (c *Client) RemoveGlobalWlTouchBind(objectID protocols.ObjectID[protocols.WlTouch]) {
	binds, ok := c.GlobalBinds[protocols.GlobalID_WlTouch]
	if !ok {
		return
	}
	delete(binds.(map[protocols.ObjectID[protocols.WlTouch]]protocols.Version), objectID)
}

func (c *Client) RemoveGlobalWlDataDeviceBind(objectID protocols.ObjectID[protocols.WlDataDevice]) {
	binds, ok := c.GlobalBinds[protocols.GlobalID_WlDataDevice]
	if !ok {
		return
	}
	delete(binds.(map[protocols.ObjectID[protocols.WlDataDevice]]protocols.Version), objectID)
}

func (c *Client) RemoveGlobalZwpXwaylandKeyboardGrabManagerV1Bind(objectID protocols.ObjectID[protocols.ZwpXwaylandKeyboardGrabManagerV1]) {
	binds, ok := c.GlobalBinds[protocols.GlobalID_ZwpXwaylandKeyboardGrabManagerV1]
	if !ok {
		return
	}
	delete(binds.(map[protocols.ObjectID[protocols.ZwpXwaylandKeyboardGrabManagerV1]]protocols.Version), objectID)
}
