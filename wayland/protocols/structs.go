package protocols

type ObjectID[T any] uint32

type AnyObjectID = ObjectID[any]

type OnBindable interface {
	OnBind(s ClientState, name AnyObjectID, interface_ string, new_id AnyObjectID, version_number uint32)
}

type WaylandObject[T OnBindable] interface {
	Delegate() T
	OnRequest(s FileDescriptorClaimClientState, message Message)
}

type OnRequestable interface {
	OnRequest(s FileDescriptorClaimClientState, message Message)
}

// Don't use these functions directly; use the ones in wayland/types.go
type ClientState interface {
	RemoveObject(AnyObjectID)
	// RemoveGlobalBind(GlobalID, AnyObjectID)
	AddObject(AnyObjectID, any)
	SetCompositorVersion(uint32)
	GetCompositorVersion() uint32
	GetObject(AnyObjectID) any

	RegisterRoleToSurface(AnyObjectID, ObjectID[WlSurface])
	UnregisterRoleToSurface(AnyObjectID)
	Send(OutgoingEvent)
	SendError(AnyObjectID, uint32, string)

	DrawableSurfaces() map[ObjectID[WlSurface]]bool
	TopLevelSurfaces() map[ObjectID[XdgToplevel]]bool
	AddFrameDrawRequest(ObjectID[WlCallback])

	GetSurfaceIDFromRole(AnyObjectID) *ObjectID[WlSurface]

	GetSurfaceFromRole(AnyObjectID) any

	GetGlobalBinds(GlobalID) any
	// AddGlobalBind(GlobalID, AnyObjectID, Version)

	AddGlobalWlShmBind(ObjectID[WlShm], Version)
	AddGlobalWlSeatBind(ObjectID[WlSeat], Version)
	AddGlobalWlOutputBind(ObjectID[WlOutput], Version)
	AddGlobalWlKeyboardBind(ObjectID[WlKeyboard], Version)
	AddGlobalWlPointerBind(ObjectID[WlPointer], Version)
	AddGlobalWlTouchBind(ObjectID[WlTouch], Version)
	AddGlobalWlDataDeviceBind(ObjectID[WlDataDevice], Version)
	AddGlobalZwpXwaylandKeyboardGrabManagerV1Bind(ObjectID[ZwpXwaylandKeyboardGrabManagerV1], Version)

	RemoveGlobalWlShmBind(ObjectID[WlShm])
	RemoveGlobalWlSeatBind(ObjectID[WlSeat])
	RemoveGlobalWlOutputBind(ObjectID[WlOutput])
	RemoveGlobalWlKeyboardBind(ObjectID[WlKeyboard])
	RemoveGlobalWlPointerBind(ObjectID[WlPointer])
	RemoveGlobalWlTouchBind(ObjectID[WlTouch])
	RemoveGlobalWlDataDeviceBind(ObjectID[WlDataDevice])
	RemoveGlobalZwpXwaylandKeyboardGrabManagerV1Bind(ObjectID[ZwpXwaylandKeyboardGrabManagerV1])
}

type OutgoingEvent struct {
	ObjectID       AnyObjectID
	Opcode         uint16
	Data           []byte
	FileDescriptor *FileDescriptor
}

type FileDescriptorClaimClientState interface {
	ClientState
	ClaimFileDescriptor() *FileDescriptor
}
type FileDescriptor int

type Sender interface {
	Send(OutgoingEvent)
}

type DecodeStateType int

type Message struct {
	ObjectID AnyObjectID
	Opcode   uint16
	Size     uint16
	Data     []byte
}

type DecodeState struct {
	Phase    DecodeStateType
	I        uint // bit offset within current field (0,8,16,24)
	ObjectID AnyObjectID
	Opcode   uint16
	Size     uint16
	Data     []byte
}

type Fixed = float64
