package win32

import (
	"syscall"
	"unsafe"
)

const (
	OBJ_HANDLE_TAGBITS                     int32  = 3
	RTL_BALANCED_NODE_RESERVED_PARENT_MASK uint32 = 0x3
	OBJ_INHERIT                            int32  = 2
	OBJ_PERMANENT                          int32  = 16
	OBJ_EXCLUSIVE                          int32  = 32
	OBJ_CASE_INSENSITIVE                   int32  = 64
	OBJ_OPENIF                             int32  = 128
	OBJ_OPENLINK                           int32  = 256
	OBJ_KERNEL_HANDLE                      int32  = 512
	OBJ_FORCE_ACCESS_CHECK                 int32  = 1024
	OBJ_IGNORE_IMPERSONATED_DEVICEMAP      int32  = 2048
	OBJ_DONT_REPARSE                       int32  = 4096
	OBJ_VALID_ATTRIBUTES                   int32  = 8178
	NULL64                                 uint32 = 0x0
	MAXUCHAR                               uint32 = 0xff
	MAXUSHORT                              uint32 = 0xffff
	MAXULONG                               uint32 = 0xffffffff
)

// enums

// enum
type EXCEPTION_DISPOSITION int32

const (
	ExceptionContinueExecution EXCEPTION_DISPOSITION = 0
	ExceptionContinueSearch    EXCEPTION_DISPOSITION = 1
	ExceptionNestedException   EXCEPTION_DISPOSITION = 2
	ExceptionCollidedUnwind    EXCEPTION_DISPOSITION = 3
)

// enum
type EVENT_TYPE int32

const (
	NotificationEvent    EVENT_TYPE = 0
	SynchronizationEvent EVENT_TYPE = 1
)

// enum
type TIMER_TYPE int32

const (
	NotificationTimer    TIMER_TYPE = 0
	SynchronizationTimer TIMER_TYPE = 1
)

// enum
type WAIT_TYPE int32

const (
	WaitAll          WAIT_TYPE = 0
	WaitAny          WAIT_TYPE = 1
	WaitNotification WAIT_TYPE = 2
	WaitDequeue      WAIT_TYPE = 3
	WaitDpc          WAIT_TYPE = 4
)

// enum
type NT_PRODUCT_TYPE int32

const (
	NtProductWinNt    NT_PRODUCT_TYPE = 1
	NtProductLanManNt NT_PRODUCT_TYPE = 2
	NtProductServer   NT_PRODUCT_TYPE = 3
)

// enum
type SUITE_TYPE int32

const (
	SmallBusiness           SUITE_TYPE = 0
	Enterprise              SUITE_TYPE = 1
	BackOffice              SUITE_TYPE = 2
	CommunicationServer     SUITE_TYPE = 3
	TerminalServer          SUITE_TYPE = 4
	SmallBusinessRestricted SUITE_TYPE = 5
	EmbeddedNT              SUITE_TYPE = 6
	DataCenter              SUITE_TYPE = 7
	SingleUserTS            SUITE_TYPE = 8
	Personal                SUITE_TYPE = 9
	Blade                   SUITE_TYPE = 10
	EmbeddedRestricted      SUITE_TYPE = 11
	SecurityAppliance       SUITE_TYPE = 12
	StorageServer           SUITE_TYPE = 13
	ComputeServer           SUITE_TYPE = 14
	WHServer                SUITE_TYPE = 15
	PhoneNT                 SUITE_TYPE = 16
	MultiUserTS             SUITE_TYPE = 17
	MaxSuiteType            SUITE_TYPE = 18
)

// enum
type COMPARTMENT_ID int32

const (
	UNSPECIFIED_COMPARTMENT_ID COMPARTMENT_ID = 0
	DEFAULT_COMPARTMENT_ID     COMPARTMENT_ID = 1
)

// structs

type SLIST_ENTRY struct {
	Next *SLIST_ENTRY
}

type QUAD_Anonymous struct {
	Data [1]uint64
}

func (this *QUAD_Anonymous) UseThisFieldToCopy() *int64 {
	return (*int64)(unsafe.Pointer(this))
}

func (this *QUAD_Anonymous) UseThisFieldToCopyVal() int64 {
	return *(*int64)(unsafe.Pointer(this))
}

func (this *QUAD_Anonymous) DoNotUseThisField() *float64 {
	return (*float64)(unsafe.Pointer(this))
}

func (this *QUAD_Anonymous) DoNotUseThisFieldVal() float64 {
	return *(*float64)(unsafe.Pointer(this))
}

type QUAD struct {
	QUAD_Anonymous
}

type PROCESSOR_NUMBER struct {
	Group    uint16
	Number   byte
	Reserved byte
}

type STRING struct {
	Length        uint16
	MaximumLength uint16
	Buffer        PSTR
}

type CSTRING struct {
	Length        uint16
	MaximumLength uint16
	Buffer        PSTR
}

type LIST_ENTRY struct {
	Flink *LIST_ENTRY
	Blink *LIST_ENTRY
}

type SINGLE_LIST_ENTRY struct {
	Next *SINGLE_LIST_ENTRY
}

type RTL_BALANCED_NODE_Anonymous1_Anonymous struct {
	Left  *RTL_BALANCED_NODE
	Right *RTL_BALANCED_NODE
}

type RTL_BALANCED_NODE_Anonymous1 struct {
	RTL_BALANCED_NODE_Anonymous1_Anonymous
}

func (this *RTL_BALANCED_NODE_Anonymous1) Children() *[2]*RTL_BALANCED_NODE {
	return (*[2]*RTL_BALANCED_NODE)(unsafe.Pointer(this))
}

func (this *RTL_BALANCED_NODE_Anonymous1) ChildrenVal() [2]*RTL_BALANCED_NODE {
	return *(*[2]*RTL_BALANCED_NODE)(unsafe.Pointer(this))
}

func (this *RTL_BALANCED_NODE_Anonymous1) Anonymous() *RTL_BALANCED_NODE_Anonymous1_Anonymous {
	return (*RTL_BALANCED_NODE_Anonymous1_Anonymous)(unsafe.Pointer(this))
}

func (this *RTL_BALANCED_NODE_Anonymous1) AnonymousVal() RTL_BALANCED_NODE_Anonymous1_Anonymous {
	return *(*RTL_BALANCED_NODE_Anonymous1_Anonymous)(unsafe.Pointer(this))
}

type RTL_BALANCED_NODE_Anonymous2 struct {
	Data [1]uint64
}

func (this *RTL_BALANCED_NODE_Anonymous2) Bitfield_() *byte {
	return (*byte)(unsafe.Pointer(this))
}

func (this *RTL_BALANCED_NODE_Anonymous2) Bitfield_Val() byte {
	return *(*byte)(unsafe.Pointer(this))
}

func (this *RTL_BALANCED_NODE_Anonymous2) ParentValue() *uintptr {
	return (*uintptr)(unsafe.Pointer(this))
}

func (this *RTL_BALANCED_NODE_Anonymous2) ParentValueVal() uintptr {
	return *(*uintptr)(unsafe.Pointer(this))
}

type RTL_BALANCED_NODE struct {
	RTL_BALANCED_NODE_Anonymous1
	RTL_BALANCED_NODE_Anonymous2
}

type LIST_ENTRY32 struct {
	Flink uint32
	Blink uint32
}

type LIST_ENTRY64 struct {
	Flink uint64
	Blink uint64
}

type SINGLE_LIST_ENTRY32 struct {
	Next uint32
}

type WNF_STATE_NAME struct {
	Data [2]uint32
}

type STRING32 struct {
	Length        uint16
	MaximumLength uint16
	Buffer        uint32
}

type STRING64 struct {
	Length        uint16
	MaximumLength uint16
	Buffer        uint64
}

type OBJECT_ATTRIBUTES64 struct {
	Length                   uint32
	RootDirectory            uint64
	ObjectName               uint64
	Attributes               uint32
	SecurityDescriptor       uint64
	SecurityQualityOfService uint64
}

type OBJECT_ATTRIBUTES32 struct {
	Length                   uint32
	RootDirectory            uint32
	ObjectName               uint32
	Attributes               uint32
	SecurityDescriptor       uint32
	SecurityQualityOfService uint32
}

type OBJECTID struct {
	Lineage    syscall.GUID
	Uniquifier uint32
}

type EXCEPTION_REGISTRATION_RECORD struct {
	Next    *EXCEPTION_REGISTRATION_RECORD
	Handler EXCEPTION_ROUTINE
}

type NT_TIB_Anonymous struct {
	Data [1]uint64
}

func (this *NT_TIB_Anonymous) FiberData() *unsafe.Pointer {
	return (*unsafe.Pointer)(unsafe.Pointer(this))
}

func (this *NT_TIB_Anonymous) FiberDataVal() unsafe.Pointer {
	return *(*unsafe.Pointer)(unsafe.Pointer(this))
}

func (this *NT_TIB_Anonymous) Version() *uint32 {
	return (*uint32)(unsafe.Pointer(this))
}

func (this *NT_TIB_Anonymous) VersionVal() uint32 {
	return *(*uint32)(unsafe.Pointer(this))
}

type NT_TIB struct {
	ExceptionList *EXCEPTION_REGISTRATION_RECORD
	StackBase     unsafe.Pointer
	StackLimit    unsafe.Pointer
	SubSystemTib  unsafe.Pointer
	NT_TIB_Anonymous
	ArbitraryUserPointer unsafe.Pointer
	Self                 *NT_TIB
}

type SLIST_HEADER_Anonymous struct {
	Alignment uint64
	Region    uint64
}

type SLIST_HEADER_HeaderX64 struct {
	Bitfield1_ uint64
	Bitfield2_ uint64
}

type SLIST_HEADER struct {
	SLIST_HEADER_Anonymous
}

func (this *SLIST_HEADER) Anonymous() *SLIST_HEADER_Anonymous {
	return (*SLIST_HEADER_Anonymous)(unsafe.Pointer(this))
}

func (this *SLIST_HEADER) AnonymousVal() SLIST_HEADER_Anonymous {
	return *(*SLIST_HEADER_Anonymous)(unsafe.Pointer(this))
}

func (this *SLIST_HEADER) HeaderX64() *SLIST_HEADER_HeaderX64 {
	return (*SLIST_HEADER_HeaderX64)(unsafe.Pointer(this))
}

func (this *SLIST_HEADER) HeaderX64Val() SLIST_HEADER_HeaderX64 {
	return *(*SLIST_HEADER_HeaderX64)(unsafe.Pointer(this))
}

type FLOATING_SAVE_AREA struct {
	ControlWord   uint32
	StatusWord    uint32
	TagWord       uint32
	ErrorOffset   uint32
	ErrorSelector uint32
	DataOffset    uint32
	DataSelector  uint32
	RegisterArea  [80]byte
	Cr0NpxState   uint32
}

// func types

type EXCEPTION_ROUTINE = uintptr
type EXCEPTION_ROUTINE_func = func(ExceptionRecord *EXCEPTION_RECORD, EstablisherFrame unsafe.Pointer, ContextRecord *CONTEXT, DispatcherContext unsafe.Pointer) EXCEPTION_DISPOSITION
