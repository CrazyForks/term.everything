package wayland

import (
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"slices"
	"time"

	"github.com/mmulet/term.everything/wayland/protocols"
)

const WAIT_TIME = time.Millisecond / 2

type Client struct {
	drawableSurfaces map[protocols.ObjectID[protocols.WlSurface]]bool
	topLevelSurfaces map[protocols.ObjectID[protocols.XdgToplevel]]bool

	UnixConnection *net.UnixConn

	CompositorVersion uint32

	DisplayID protocols.ObjectID[protocols.WlDisplay]

	MessageBuffer []byte

	OutgoingChannel chan protocols.OutgoingEvent

	Decoder MessageDecoder

	UnclaimedFDs []protocols.FileDescriptor

	Objects map[protocols.AnyObjectID]any

	RolesToSurfaces map[protocols.AnyObjectID]protocols.ObjectID[protocols.WlSurface]

	FrameDrawRequests chan protocols.ObjectID[protocols.WlCallback]

	GlobalBinds map[protocols.GlobalID]map[protocols.AnyObjectID]protocols.Version

	LastGetMessageTime time.Time
}

func (c *Client) AddFrameDrawRequest(cb protocols.ObjectID[protocols.WlCallback]) {
	c.FrameDrawRequests <- cb
}

func (c *Client) GetSurfaceIDFromRole(roleObjectID protocols.AnyObjectID) *protocols.ObjectID[protocols.WlSurface] {
	if sid, ok := c.RolesToSurfaces[roleObjectID]; ok {
		return &sid
	}
	return nil
}

func (c *Client) GetSurfaceFromRole(roleObjectID protocols.AnyObjectID) any {
	sidAny := c.GetSurfaceIDFromRole(roleObjectID)
	if sidAny == nil {
		return nil
	}
	surface := GetWlSurfaceObject(c, *sidAny)
	return surface
}

func (c *Client) UnregisterRoleToSurface(roleID protocols.AnyObjectID) {
	delete(c.RolesToSurfaces, roleID)
}

func (c *Client) RegisterRoleToSurface(roleID protocols.AnyObjectID, surfaceID protocols.ObjectID[protocols.WlSurface]) {
	c.RolesToSurfaces[roleID] = surfaceID
}

/**
 * Seed if maybe_desceneding_id is a descendant of surface_id
 * @param s
 * @param surface_id
 * @param maybe_descendant_id
 */
func (c *Client) FindDescendantSurface(surfaceID protocols.ObjectID[protocols.WlSurface], maybeDescendantID protocols.ObjectID[protocols.WlSurface]) bool {

	surface := GetWlSurfaceObject(c, surfaceID)
	if surface == nil {
		return false
	}

	for _, childID := range surface.ChildrenInDrawOrder {
		if childID == nil {
			continue
		}
		if *childID == maybeDescendantID {
			return true
		}
	}

	for _, childID := range surface.ChildrenInDrawOrder {
		if childID == nil {
			continue
		}
		if c.FindDescendantSurface(*childID, maybeDescendantID) {
			return true
		}
	}

	return false
}

func (c *Client) SendError(objectID protocols.AnyObjectID, code uint32, message string) {
	protocols.WlDisplay_error(c,
		protocols.ObjectID[protocols.WlDisplay](protocols.GlobalID_WlDisplay),
		objectID,
		code,
		message,
	)
}

func (c *Client) GetGlobalBinds(globalID protocols.GlobalID) any {
	return c.GlobalBinds[globalID]
}

func (c *Client) RemoveGlobalBind(globalID protocols.GlobalID, id protocols.AnyObjectID) {
	binds, ok := c.GlobalBinds[globalID]
	if !ok {
		return
	}
	delete(binds, id)
}

/**
 * Add a bound object_id to a list
 * of global_ids. SO that you can
 * ask, What are all the objects bound
 * to this global for this client?
 * @param global_id
 * @param object_id
 */
func (c *Client) AddGlobalBind(globalID protocols.GlobalID, objectID protocols.AnyObjectID, version protocols.Version) {
	binds, ok := c.GlobalBinds[globalID]
	if !ok {
		binds = make(map[protocols.AnyObjectID]protocols.Version)
		c.GlobalBinds[globalID] = binds
	}
	binds[objectID] = version
}

func (c *Client) AddObject(id protocols.AnyObjectID, v any) {
	if v == nil {
		log.Printf("AddObject: object is nil for id %d", uint32(id))
	}
	if _, already_have := c.Objects[id]; already_have {
		log.Printf("AddObject: object already exists for id %d", uint32(id))
	}
	c.Objects[id] = v
}

func (c *Client) RemoveObject(id protocols.AnyObjectID) {
	delete(c.Objects, id)
}

func (c *Client) GetObject(id protocols.AnyObjectID) any {
	object, ok := c.Objects[id]
	if !ok {
		return nil
	}
	return object
}

func MakeClient(conn *net.UnixConn) *Client {
	return &Client{
		UnixConnection:    conn,
		CompositorVersion: 1,
		DisplayID:         protocols.ObjectID[protocols.WlDisplay](1),

		MessageBuffer: make([]byte, 64*1024),

		OutgoingChannel: make(chan protocols.OutgoingEvent, 1024),

		UnclaimedFDs:    make([]protocols.FileDescriptor, 0, 8),
		Objects:         make(map[protocols.AnyObjectID]any),
		RolesToSurfaces: make(map[protocols.AnyObjectID]protocols.ObjectID[protocols.WlSurface]),

		drawableSurfaces: make(map[protocols.ObjectID[protocols.WlSurface]]bool),
		topLevelSurfaces: make(map[protocols.ObjectID[protocols.XdgToplevel]]bool),

		GlobalBinds:        make(map[protocols.GlobalID]map[protocols.AnyObjectID]protocols.Version),
		LastGetMessageTime: time.Now().Add(-WAIT_TIME - time.Millisecond),
		FrameDrawRequests:  make(chan protocols.ObjectID[protocols.WlCallback], 1024),
	}
}

func (c *Client) MainLoop() {
	for {

		for {
			select {
			case ev := <-c.OutgoingChannel:
				if err := c.SendPendingMessage(ev); err != nil {
					log.Printf("Send error: %v", err)
					return
				}
			default:
				elapsed := time.Since(c.LastGetMessageTime)
				if elapsed < WAIT_TIME {
					time.Sleep(WAIT_TIME - elapsed)
				}
				c.LastGetMessageTime = time.Now()
				goto drained
			}
		}
	drained:

		// Receive once with short deadline; parse and dispatch.
		n, fds, err := GetMessageAndFileDescriptors(c.UnixConnection, c.MessageBuffer)
		if err != nil {
			// treat unexpected read errors as fatal
			log.Printf("Recv error: %v", err)
			return
		}
		if err := c.ParseMessages(n, fds); err != nil {
			log.Printf("Parse error: %v", err)
			return
		}
	}
}

func (c *Client) Send(ev protocols.OutgoingEvent) {
	// Allow backpressure to naturally block the sender goroutine.
	c.OutgoingChannel <- ev
}

/**
 *
 * @param message
 * @returns Returns if we should continue listening or sending on this socket any more
 * returns falsy mostly if the client has disconnected
 */
func (c *Client) SendPendingMessage(ev protocols.OutgoingEvent) error {
	// if WaylandDebugTimeOnly() {

	// 	log.Printf("client -> eid=%d opcode=%d len=%d fd=%v",
	// 		uint32(ev.ObjectID), ev.Opcode, len(ev.Data), ev.FileDescriptor)
	// }
	/**
	 * 8 bytes is the header length + the length of the message
	 * #### Header is
	 * - 4 bytes for object_id
	 * - 2 bytes for opcode
	 * - 2 bytes for size
	 */
	size := 8 + len(ev.Data)
	buf := make([]byte, size)

	// Wayland header: object_id (u32), size (u16), opcode (u16)
	binary.LittleEndian.PutUint32(buf[0:4], uint32(ev.ObjectID))
	binary.LittleEndian.PutUint16(buf[4:6], uint16(size))
	binary.LittleEndian.PutUint16(buf[6:8], uint16(ev.Opcode))
	copy(buf[8:], ev.Data)

	var fds []int
	if ev.FileDescriptor != nil {
		fds = []int{int(*ev.FileDescriptor)}
	}
	return SendMessageAndFileDescriptors(c.UnixConnection, buf, fds)
	// re

	// if err != nil {
	// 	// if EPIPE or similar => disconnect
	// 	if errors.Is(err, syscall.EPIPE) || errors.Is(err, syscall.ECONNRESET) {
	// 		return false
	// 	}
	// 	log.Printf("Send error: %v (n=%d ok=%v)", err, n, ok)
	// 	return false
	// }
	// return true
}

func (c *Client) ParseMessages(n int, fds []int) error {
	// if len(fds) > 0 && WaylandDebugTimeOnly() {
	// 	log.Printf("client: received %d file descriptors", len(fds))
	// }
	for _, fd := range fds {
		c.UnclaimedFDs = append(c.UnclaimedFDs, protocols.FileDescriptor(fd))
	}

	if n < 0 {
		return fmt.Errorf("negative byte count received: %d", n)
	}

	if n == 0 {
		/**
		* Time out
		 */
		return nil
	}

	msgs := c.Decoder.Consume(c.MessageBuffer[:n])
	for i := range msgs {
		m := msgs[i]
		obj := c.GetObject(m.ObjectID)
		if obj == nil {
			// if WaylandDebugTimeOnly() {
			// 	log.Printf("client: request for unknown object %d", uint32(m.ObjectID))
			// }
			continue
		}

		theType, ok := obj.(protocols.OnRequestable)
		if !ok {
			log.Printf("client: object %d has unknown type; cannot dispatch", uint32(m.ObjectID))
			continue
		}
		theType.OnRequest(c, m)
	}
	return nil
}

func (c *Client) ClaimFileDescriptor() *protocols.FileDescriptor {
	if len(c.UnclaimedFDs) == 0 {
		return nil
	}
	fd := c.UnclaimedFDs[0]
	c.UnclaimedFDs = slices.Delete(c.UnclaimedFDs, 0, 1)
	return &fd
}

func (c *Client) SetCompositorVersion(v uint32) { c.CompositorVersion = v }
func (c *Client) GetCompositorVersion() uint32  { return c.CompositorVersion }

func (c *Client) DrawableSurfaces() map[protocols.ObjectID[protocols.WlSurface]]bool {
	return c.drawableSurfaces
}
func (c *Client) TopLevelSurfaces() map[protocols.ObjectID[protocols.XdgToplevel]]bool {
	return c.topLevelSurfaces
}
