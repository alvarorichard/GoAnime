package win32

import (
	"syscall"
	"unsafe"
)

type (
	HPSS     = uintptr
	HPSSWALK = uintptr
)

const (
	PSS_PERF_RESOLUTION uint32 = 0xf4240
)

// enums

// enum
// flags
type PSS_HANDLE_FLAGS uint32

const (
	PSS_HANDLE_NONE                           PSS_HANDLE_FLAGS = 0
	PSS_HANDLE_HAVE_TYPE                      PSS_HANDLE_FLAGS = 1
	PSS_HANDLE_HAVE_NAME                      PSS_HANDLE_FLAGS = 2
	PSS_HANDLE_HAVE_BASIC_INFORMATION         PSS_HANDLE_FLAGS = 4
	PSS_HANDLE_HAVE_TYPE_SPECIFIC_INFORMATION PSS_HANDLE_FLAGS = 8
)

// enum
type PSS_OBJECT_TYPE int32

const (
	PSS_OBJECT_TYPE_UNKNOWN   PSS_OBJECT_TYPE = 0
	PSS_OBJECT_TYPE_PROCESS   PSS_OBJECT_TYPE = 1
	PSS_OBJECT_TYPE_THREAD    PSS_OBJECT_TYPE = 2
	PSS_OBJECT_TYPE_MUTANT    PSS_OBJECT_TYPE = 3
	PSS_OBJECT_TYPE_EVENT     PSS_OBJECT_TYPE = 4
	PSS_OBJECT_TYPE_SECTION   PSS_OBJECT_TYPE = 5
	PSS_OBJECT_TYPE_SEMAPHORE PSS_OBJECT_TYPE = 6
)

// enum
// flags
type PSS_CAPTURE_FLAGS uint32

const (
	PSS_CAPTURE_NONE                             PSS_CAPTURE_FLAGS = 0
	PSS_CAPTURE_VA_CLONE                         PSS_CAPTURE_FLAGS = 1
	PSS_CAPTURE_RESERVED_00000002                PSS_CAPTURE_FLAGS = 2
	PSS_CAPTURE_HANDLES                          PSS_CAPTURE_FLAGS = 4
	PSS_CAPTURE_HANDLE_NAME_INFORMATION          PSS_CAPTURE_FLAGS = 8
	PSS_CAPTURE_HANDLE_BASIC_INFORMATION         PSS_CAPTURE_FLAGS = 16
	PSS_CAPTURE_HANDLE_TYPE_SPECIFIC_INFORMATION PSS_CAPTURE_FLAGS = 32
	PSS_CAPTURE_HANDLE_TRACE                     PSS_CAPTURE_FLAGS = 64
	PSS_CAPTURE_THREADS                          PSS_CAPTURE_FLAGS = 128
	PSS_CAPTURE_THREAD_CONTEXT                   PSS_CAPTURE_FLAGS = 256
	PSS_CAPTURE_THREAD_CONTEXT_EXTENDED          PSS_CAPTURE_FLAGS = 512
	PSS_CAPTURE_RESERVED_00000400                PSS_CAPTURE_FLAGS = 1024
	PSS_CAPTURE_VA_SPACE                         PSS_CAPTURE_FLAGS = 2048
	PSS_CAPTURE_VA_SPACE_SECTION_INFORMATION     PSS_CAPTURE_FLAGS = 4096
	PSS_CAPTURE_IPT_TRACE                        PSS_CAPTURE_FLAGS = 8192
	PSS_CAPTURE_RESERVED_00004000                PSS_CAPTURE_FLAGS = 16384
	PSS_CREATE_BREAKAWAY_OPTIONAL                PSS_CAPTURE_FLAGS = 67108864
	PSS_CREATE_BREAKAWAY                         PSS_CAPTURE_FLAGS = 134217728
	PSS_CREATE_FORCE_BREAKAWAY                   PSS_CAPTURE_FLAGS = 268435456
	PSS_CREATE_USE_VM_ALLOCATIONS                PSS_CAPTURE_FLAGS = 536870912
	PSS_CREATE_MEASURE_PERFORMANCE               PSS_CAPTURE_FLAGS = 1073741824
	PSS_CREATE_RELEASE_SECTION                   PSS_CAPTURE_FLAGS = 2147483648
)

// enum
type PSS_QUERY_INFORMATION_CLASS int32

const (
	PSS_QUERY_PROCESS_INFORMATION         PSS_QUERY_INFORMATION_CLASS = 0
	PSS_QUERY_VA_CLONE_INFORMATION        PSS_QUERY_INFORMATION_CLASS = 1
	PSS_QUERY_AUXILIARY_PAGES_INFORMATION PSS_QUERY_INFORMATION_CLASS = 2
	PSS_QUERY_VA_SPACE_INFORMATION        PSS_QUERY_INFORMATION_CLASS = 3
	PSS_QUERY_HANDLE_INFORMATION          PSS_QUERY_INFORMATION_CLASS = 4
	PSS_QUERY_THREAD_INFORMATION          PSS_QUERY_INFORMATION_CLASS = 5
	PSS_QUERY_HANDLE_TRACE_INFORMATION    PSS_QUERY_INFORMATION_CLASS = 6
	PSS_QUERY_PERFORMANCE_COUNTERS        PSS_QUERY_INFORMATION_CLASS = 7
)

// enum
type PSS_WALK_INFORMATION_CLASS int32

const (
	PSS_WALK_AUXILIARY_PAGES PSS_WALK_INFORMATION_CLASS = 0
	PSS_WALK_VA_SPACE        PSS_WALK_INFORMATION_CLASS = 1
	PSS_WALK_HANDLES         PSS_WALK_INFORMATION_CLASS = 2
	PSS_WALK_THREADS         PSS_WALK_INFORMATION_CLASS = 3
)

// enum
// flags
type PSS_DUPLICATE_FLAGS uint32

const (
	PSS_DUPLICATE_NONE         PSS_DUPLICATE_FLAGS = 0
	PSS_DUPLICATE_CLOSE_SOURCE PSS_DUPLICATE_FLAGS = 1
)

// enum
// flags
type PSS_PROCESS_FLAGS uint32

const (
	PSS_PROCESS_FLAGS_NONE        PSS_PROCESS_FLAGS = 0
	PSS_PROCESS_FLAGS_PROTECTED   PSS_PROCESS_FLAGS = 1
	PSS_PROCESS_FLAGS_WOW64       PSS_PROCESS_FLAGS = 2
	PSS_PROCESS_FLAGS_RESERVED_03 PSS_PROCESS_FLAGS = 4
	PSS_PROCESS_FLAGS_RESERVED_04 PSS_PROCESS_FLAGS = 8
	PSS_PROCESS_FLAGS_FROZEN      PSS_PROCESS_FLAGS = 16
)

// enum
// flags
type PSS_THREAD_FLAGS uint32

const (
	PSS_THREAD_FLAGS_NONE       PSS_THREAD_FLAGS = 0
	PSS_THREAD_FLAGS_TERMINATED PSS_THREAD_FLAGS = 1
)

// structs

type PSS_PROCESS_INFORMATION struct {
	ExitStatus                 uint32
	PebBaseAddress             unsafe.Pointer
	AffinityMask               uintptr
	BasePriority               int32
	ProcessId                  uint32
	ParentProcessId            uint32
	Flags                      PSS_PROCESS_FLAGS
	CreateTime                 FILETIME
	ExitTime                   FILETIME
	KernelTime                 FILETIME
	UserTime                   FILETIME
	PriorityClass              uint32
	PeakVirtualSize            uintptr
	VirtualSize                uintptr
	PageFaultCount             uint32
	PeakWorkingSetSize         uintptr
	WorkingSetSize             uintptr
	QuotaPeakPagedPoolUsage    uintptr
	QuotaPagedPoolUsage        uintptr
	QuotaPeakNonPagedPoolUsage uintptr
	QuotaNonPagedPoolUsage     uintptr
	PagefileUsage              uintptr
	PeakPagefileUsage          uintptr
	PrivateUsage               uintptr
	ExecuteFlags               uint32
	ImageFileName              [260]uint16
}

type PSS_VA_CLONE_INFORMATION struct {
	VaCloneHandle HANDLE
}

type PSS_AUXILIARY_PAGES_INFORMATION struct {
	AuxPagesCaptured uint32
}

type PSS_VA_SPACE_INFORMATION struct {
	RegionCount uint32
}

type PSS_HANDLE_INFORMATION struct {
	HandlesCaptured uint32
}

type PSS_THREAD_INFORMATION struct {
	ThreadsCaptured uint32
	ContextLength   uint32
}

type PSS_HANDLE_TRACE_INFORMATION struct {
	SectionHandle HANDLE
	Size          uint32
}

type PSS_PERFORMANCE_COUNTERS struct {
	TotalCycleCount         uint64
	TotalWallClockPeriod    uint64
	VaCloneCycleCount       uint64
	VaCloneWallClockPeriod  uint64
	VaSpaceCycleCount       uint64
	VaSpaceWallClockPeriod  uint64
	AuxPagesCycleCount      uint64
	AuxPagesWallClockPeriod uint64
	HandlesCycleCount       uint64
	HandlesWallClockPeriod  uint64
	ThreadsCycleCount       uint64
	ThreadsWallClockPeriod  uint64
}

type PSS_AUXILIARY_PAGE_ENTRY struct {
	Address          unsafe.Pointer
	BasicInformation MEMORY_BASIC_INFORMATION
	CaptureTime      FILETIME
	PageContents     unsafe.Pointer
	PageSize         uint32
}

type PSS_VA_SPACE_ENTRY struct {
	BaseAddress          unsafe.Pointer
	AllocationBase       unsafe.Pointer
	AllocationProtect    uint32
	RegionSize           uintptr
	State                uint32
	Protect              uint32
	Type                 uint32
	TimeDateStamp        uint32
	SizeOfImage          uint32
	ImageBase            unsafe.Pointer
	CheckSum             uint32
	MappedFileNameLength uint16
	MappedFileName       PWSTR
}

type PSS_HANDLE_ENTRY_TypeSpecificInformation_Process struct {
	ExitStatus      uint32
	PebBaseAddress  unsafe.Pointer
	AffinityMask    uintptr
	BasePriority    int32
	ProcessId       uint32
	ParentProcessId uint32
	Flags           uint32
}

type PSS_HANDLE_ENTRY_TypeSpecificInformation_Thread struct {
	ExitStatus        uint32
	TebBaseAddress    unsafe.Pointer
	ProcessId         uint32
	ThreadId          uint32
	AffinityMask      uintptr
	Priority          int32
	BasePriority      int32
	Win32StartAddress unsafe.Pointer
}

type PSS_HANDLE_ENTRY_TypeSpecificInformation_Mutant struct {
	CurrentCount   int32
	Abandoned      BOOL
	OwnerProcessId uint32
	OwnerThreadId  uint32
}

type PSS_HANDLE_ENTRY_TypeSpecificInformation_Event struct {
	ManualReset BOOL
	Signaled    BOOL
}

type PSS_HANDLE_ENTRY_TypeSpecificInformation_Section struct {
	BaseAddress          unsafe.Pointer
	AllocationAttributes uint32
	MaximumSize          int64
}

type PSS_HANDLE_ENTRY_TypeSpecificInformation_Semaphore struct {
	CurrentCount int32
	MaximumCount int32
}

type PSS_HANDLE_ENTRY_TypeSpecificInformation struct {
	Data [6]uint64
}

func (this *PSS_HANDLE_ENTRY_TypeSpecificInformation) Process() *PSS_HANDLE_ENTRY_TypeSpecificInformation_Process {
	return (*PSS_HANDLE_ENTRY_TypeSpecificInformation_Process)(unsafe.Pointer(this))
}

func (this *PSS_HANDLE_ENTRY_TypeSpecificInformation) ProcessVal() PSS_HANDLE_ENTRY_TypeSpecificInformation_Process {
	return *(*PSS_HANDLE_ENTRY_TypeSpecificInformation_Process)(unsafe.Pointer(this))
}

func (this *PSS_HANDLE_ENTRY_TypeSpecificInformation) Thread() *PSS_HANDLE_ENTRY_TypeSpecificInformation_Thread {
	return (*PSS_HANDLE_ENTRY_TypeSpecificInformation_Thread)(unsafe.Pointer(this))
}

func (this *PSS_HANDLE_ENTRY_TypeSpecificInformation) ThreadVal() PSS_HANDLE_ENTRY_TypeSpecificInformation_Thread {
	return *(*PSS_HANDLE_ENTRY_TypeSpecificInformation_Thread)(unsafe.Pointer(this))
}

func (this *PSS_HANDLE_ENTRY_TypeSpecificInformation) Mutant() *PSS_HANDLE_ENTRY_TypeSpecificInformation_Mutant {
	return (*PSS_HANDLE_ENTRY_TypeSpecificInformation_Mutant)(unsafe.Pointer(this))
}

func (this *PSS_HANDLE_ENTRY_TypeSpecificInformation) MutantVal() PSS_HANDLE_ENTRY_TypeSpecificInformation_Mutant {
	return *(*PSS_HANDLE_ENTRY_TypeSpecificInformation_Mutant)(unsafe.Pointer(this))
}

func (this *PSS_HANDLE_ENTRY_TypeSpecificInformation) Event() *PSS_HANDLE_ENTRY_TypeSpecificInformation_Event {
	return (*PSS_HANDLE_ENTRY_TypeSpecificInformation_Event)(unsafe.Pointer(this))
}

func (this *PSS_HANDLE_ENTRY_TypeSpecificInformation) EventVal() PSS_HANDLE_ENTRY_TypeSpecificInformation_Event {
	return *(*PSS_HANDLE_ENTRY_TypeSpecificInformation_Event)(unsafe.Pointer(this))
}

func (this *PSS_HANDLE_ENTRY_TypeSpecificInformation) Section() *PSS_HANDLE_ENTRY_TypeSpecificInformation_Section {
	return (*PSS_HANDLE_ENTRY_TypeSpecificInformation_Section)(unsafe.Pointer(this))
}

func (this *PSS_HANDLE_ENTRY_TypeSpecificInformation) SectionVal() PSS_HANDLE_ENTRY_TypeSpecificInformation_Section {
	return *(*PSS_HANDLE_ENTRY_TypeSpecificInformation_Section)(unsafe.Pointer(this))
}

func (this *PSS_HANDLE_ENTRY_TypeSpecificInformation) Semaphore() *PSS_HANDLE_ENTRY_TypeSpecificInformation_Semaphore {
	return (*PSS_HANDLE_ENTRY_TypeSpecificInformation_Semaphore)(unsafe.Pointer(this))
}

func (this *PSS_HANDLE_ENTRY_TypeSpecificInformation) SemaphoreVal() PSS_HANDLE_ENTRY_TypeSpecificInformation_Semaphore {
	return *(*PSS_HANDLE_ENTRY_TypeSpecificInformation_Semaphore)(unsafe.Pointer(this))
}

type PSS_HANDLE_ENTRY struct {
	Handle                  HANDLE
	Flags                   PSS_HANDLE_FLAGS
	ObjectType              PSS_OBJECT_TYPE
	CaptureTime             FILETIME
	Attributes              uint32
	GrantedAccess           uint32
	HandleCount             uint32
	PointerCount            uint32
	PagedPoolCharge         uint32
	NonPagedPoolCharge      uint32
	CreationTime            FILETIME
	TypeNameLength          uint16
	TypeName                PWSTR
	ObjectNameLength        uint16
	ObjectName              PWSTR
	TypeSpecificInformation PSS_HANDLE_ENTRY_TypeSpecificInformation
}

type PSS_THREAD_ENTRY struct {
	ExitStatus               uint32
	TebBaseAddress           unsafe.Pointer
	ProcessId                uint32
	ThreadId                 uint32
	AffinityMask             uintptr
	Priority                 int32
	BasePriority             int32
	LastSyscallFirstArgument unsafe.Pointer
	LastSyscallNumber        uint16
	CreateTime               FILETIME
	ExitTime                 FILETIME
	KernelTime               FILETIME
	UserTime                 FILETIME
	Win32StartAddress        unsafe.Pointer
	CaptureTime              FILETIME
	Flags                    PSS_THREAD_FLAGS
	SuspendCount             uint16
	SizeOfContextRecord      uint16
	ContextRecord            *CONTEXT
}

type PSS_ALLOCATOR struct {
	Context      unsafe.Pointer
	AllocRoutine uintptr
	FreeRoutine  uintptr
}

var (
	pPssCaptureSnapshot           uintptr
	pPssFreeSnapshot              uintptr
	pPssQuerySnapshot             uintptr
	pPssWalkSnapshot              uintptr
	pPssDuplicateSnapshot         uintptr
	pPssWalkMarkerCreate          uintptr
	pPssWalkMarkerFree            uintptr
	pPssWalkMarkerGetPosition     uintptr
	pPssWalkMarkerSetPosition     uintptr
	pPssWalkMarkerSeekToBeginning uintptr
)

func PssCaptureSnapshot(ProcessHandle HANDLE, CaptureFlags PSS_CAPTURE_FLAGS, ThreadContextFlags uint32, SnapshotHandle *HPSS) uint32 {
	addr := lazyAddr(&pPssCaptureSnapshot, libKernel32, "PssCaptureSnapshot")
	ret, _, _ := syscall.SyscallN(addr, ProcessHandle, uintptr(CaptureFlags), uintptr(ThreadContextFlags), uintptr(unsafe.Pointer(SnapshotHandle)))
	return uint32(ret)
}

func PssFreeSnapshot(ProcessHandle HANDLE, SnapshotHandle HPSS) uint32 {
	addr := lazyAddr(&pPssFreeSnapshot, libKernel32, "PssFreeSnapshot")
	ret, _, _ := syscall.SyscallN(addr, ProcessHandle, SnapshotHandle)
	return uint32(ret)
}

func PssQuerySnapshot(SnapshotHandle HPSS, InformationClass PSS_QUERY_INFORMATION_CLASS, Buffer unsafe.Pointer, BufferLength uint32) uint32 {
	addr := lazyAddr(&pPssQuerySnapshot, libKernel32, "PssQuerySnapshot")
	ret, _, _ := syscall.SyscallN(addr, SnapshotHandle, uintptr(InformationClass), uintptr(Buffer), uintptr(BufferLength))
	return uint32(ret)
}

func PssWalkSnapshot(SnapshotHandle HPSS, InformationClass PSS_WALK_INFORMATION_CLASS, WalkMarkerHandle HPSSWALK, Buffer unsafe.Pointer, BufferLength uint32) uint32 {
	addr := lazyAddr(&pPssWalkSnapshot, libKernel32, "PssWalkSnapshot")
	ret, _, _ := syscall.SyscallN(addr, SnapshotHandle, uintptr(InformationClass), WalkMarkerHandle, uintptr(Buffer), uintptr(BufferLength))
	return uint32(ret)
}

func PssDuplicateSnapshot(SourceProcessHandle HANDLE, SnapshotHandle HPSS, TargetProcessHandle HANDLE, TargetSnapshotHandle *HPSS, Flags PSS_DUPLICATE_FLAGS) uint32 {
	addr := lazyAddr(&pPssDuplicateSnapshot, libKernel32, "PssDuplicateSnapshot")
	ret, _, _ := syscall.SyscallN(addr, SourceProcessHandle, SnapshotHandle, TargetProcessHandle, uintptr(unsafe.Pointer(TargetSnapshotHandle)), uintptr(Flags))
	return uint32(ret)
}

func PssWalkMarkerCreate(Allocator *PSS_ALLOCATOR, WalkMarkerHandle *HPSSWALK) uint32 {
	addr := lazyAddr(&pPssWalkMarkerCreate, libKernel32, "PssWalkMarkerCreate")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(Allocator)), uintptr(unsafe.Pointer(WalkMarkerHandle)))
	return uint32(ret)
}

func PssWalkMarkerFree(WalkMarkerHandle HPSSWALK) uint32 {
	addr := lazyAddr(&pPssWalkMarkerFree, libKernel32, "PssWalkMarkerFree")
	ret, _, _ := syscall.SyscallN(addr, WalkMarkerHandle)
	return uint32(ret)
}

func PssWalkMarkerGetPosition(WalkMarkerHandle HPSSWALK, Position *uintptr) uint32 {
	addr := lazyAddr(&pPssWalkMarkerGetPosition, libKernel32, "PssWalkMarkerGetPosition")
	ret, _, _ := syscall.SyscallN(addr, WalkMarkerHandle, uintptr(unsafe.Pointer(Position)))
	return uint32(ret)
}

func PssWalkMarkerSetPosition(WalkMarkerHandle HPSSWALK, Position uintptr) uint32 {
	addr := lazyAddr(&pPssWalkMarkerSetPosition, libKernel32, "PssWalkMarkerSetPosition")
	ret, _, _ := syscall.SyscallN(addr, WalkMarkerHandle, Position)
	return uint32(ret)
}

func PssWalkMarkerSeekToBeginning(WalkMarkerHandle HPSSWALK) uint32 {
	addr := lazyAddr(&pPssWalkMarkerSeekToBeginning, libKernel32, "PssWalkMarkerSeekToBeginning")
	ret, _, _ := syscall.SyscallN(addr, WalkMarkerHandle)
	return uint32(ret)
}
