package win32

import (
	"syscall"
	"unsafe"
)

type (
	HeapHandle = uintptr
)

const (
	FILE_CACHE_MAX_HARD_ENABLE   uint32 = 0x1
	FILE_CACHE_MAX_HARD_DISABLE  uint32 = 0x2
	FILE_CACHE_MIN_HARD_ENABLE   uint32 = 0x4
	FILE_CACHE_MIN_HARD_DISABLE  uint32 = 0x8
	MEHC_PATROL_SCRUBBER_PRESENT uint32 = 0x1
)

// enums

// enum
// flags
type FILE_MAP uint32

const (
	FILE_MAP_WRITE           FILE_MAP = 2
	FILE_MAP_READ            FILE_MAP = 4
	FILE_MAP_ALL_ACCESS      FILE_MAP = 983071
	FILE_MAP_EXECUTE         FILE_MAP = 32
	FILE_MAP_COPY            FILE_MAP = 1
	FILE_MAP_RESERVE         FILE_MAP = 2147483648
	FILE_MAP_TARGETS_INVALID FILE_MAP = 1073741824
	FILE_MAP_LARGE_PAGES     FILE_MAP = 536870912
)

// enum
// flags
type HEAP_FLAGS uint32

const (
	HEAP_NONE                     HEAP_FLAGS = 0
	HEAP_NO_SERIALIZE             HEAP_FLAGS = 1
	HEAP_GROWABLE                 HEAP_FLAGS = 2
	HEAP_GENERATE_EXCEPTIONS      HEAP_FLAGS = 4
	HEAP_ZERO_MEMORY              HEAP_FLAGS = 8
	HEAP_REALLOC_IN_PLACE_ONLY    HEAP_FLAGS = 16
	HEAP_TAIL_CHECKING_ENABLED    HEAP_FLAGS = 32
	HEAP_FREE_CHECKING_ENABLED    HEAP_FLAGS = 64
	HEAP_DISABLE_COALESCE_ON_FREE HEAP_FLAGS = 128
	HEAP_CREATE_ALIGN_16          HEAP_FLAGS = 65536
	HEAP_CREATE_ENABLE_TRACING    HEAP_FLAGS = 131072
	HEAP_CREATE_ENABLE_EXECUTE    HEAP_FLAGS = 262144
	HEAP_MAXIMUM_TAG              HEAP_FLAGS = 4095
	HEAP_PSEUDO_TAG_FLAG          HEAP_FLAGS = 32768
	HEAP_TAG_SHIFT                HEAP_FLAGS = 18
	HEAP_CREATE_SEGMENT_HEAP      HEAP_FLAGS = 256
	HEAP_CREATE_HARDENED          HEAP_FLAGS = 512
)

// enum
// flags
type PAGE_PROTECTION_FLAGS uint32

const (
	PAGE_NOACCESS                   PAGE_PROTECTION_FLAGS = 1
	PAGE_READONLY                   PAGE_PROTECTION_FLAGS = 2
	PAGE_READWRITE                  PAGE_PROTECTION_FLAGS = 4
	PAGE_WRITECOPY                  PAGE_PROTECTION_FLAGS = 8
	PAGE_EXECUTE                    PAGE_PROTECTION_FLAGS = 16
	PAGE_EXECUTE_READ               PAGE_PROTECTION_FLAGS = 32
	PAGE_EXECUTE_READWRITE          PAGE_PROTECTION_FLAGS = 64
	PAGE_EXECUTE_WRITECOPY          PAGE_PROTECTION_FLAGS = 128
	PAGE_GUARD                      PAGE_PROTECTION_FLAGS = 256
	PAGE_NOCACHE                    PAGE_PROTECTION_FLAGS = 512
	PAGE_WRITECOMBINE               PAGE_PROTECTION_FLAGS = 1024
	PAGE_GRAPHICS_NOACCESS          PAGE_PROTECTION_FLAGS = 2048
	PAGE_GRAPHICS_READONLY          PAGE_PROTECTION_FLAGS = 4096
	PAGE_GRAPHICS_READWRITE         PAGE_PROTECTION_FLAGS = 8192
	PAGE_GRAPHICS_EXECUTE           PAGE_PROTECTION_FLAGS = 16384
	PAGE_GRAPHICS_EXECUTE_READ      PAGE_PROTECTION_FLAGS = 32768
	PAGE_GRAPHICS_EXECUTE_READWRITE PAGE_PROTECTION_FLAGS = 65536
	PAGE_GRAPHICS_COHERENT          PAGE_PROTECTION_FLAGS = 131072
	PAGE_GRAPHICS_NOCACHE           PAGE_PROTECTION_FLAGS = 262144
	PAGE_ENCLAVE_THREAD_CONTROL     PAGE_PROTECTION_FLAGS = 2147483648
	PAGE_REVERT_TO_FILE_MAP         PAGE_PROTECTION_FLAGS = 2147483648
	PAGE_TARGETS_NO_UPDATE          PAGE_PROTECTION_FLAGS = 1073741824
	PAGE_TARGETS_INVALID            PAGE_PROTECTION_FLAGS = 1073741824
	PAGE_ENCLAVE_UNVALIDATED        PAGE_PROTECTION_FLAGS = 536870912
	PAGE_ENCLAVE_MASK               PAGE_PROTECTION_FLAGS = 268435456
	PAGE_ENCLAVE_DECOMMIT           PAGE_PROTECTION_FLAGS = 268435456
	PAGE_ENCLAVE_SS_FIRST           PAGE_PROTECTION_FLAGS = 268435457
	PAGE_ENCLAVE_SS_REST            PAGE_PROTECTION_FLAGS = 268435458
	SEC_PARTITION_OWNER_HANDLE      PAGE_PROTECTION_FLAGS = 262144
	SEC_64K_PAGES                   PAGE_PROTECTION_FLAGS = 524288
	SEC_FILE                        PAGE_PROTECTION_FLAGS = 8388608
	SEC_IMAGE                       PAGE_PROTECTION_FLAGS = 16777216
	SEC_PROTECTED_IMAGE             PAGE_PROTECTION_FLAGS = 33554432
	SEC_RESERVE                     PAGE_PROTECTION_FLAGS = 67108864
	SEC_COMMIT                      PAGE_PROTECTION_FLAGS = 134217728
	SEC_NOCACHE                     PAGE_PROTECTION_FLAGS = 268435456
	SEC_WRITECOMBINE                PAGE_PROTECTION_FLAGS = 1073741824
	SEC_LARGE_PAGES                 PAGE_PROTECTION_FLAGS = 2147483648
	SEC_IMAGE_NO_EXECUTE            PAGE_PROTECTION_FLAGS = 285212672
)

// enum
type UNMAP_VIEW_OF_FILE_FLAGS uint32

const (
	MEM_UNMAP_NONE                 UNMAP_VIEW_OF_FILE_FLAGS = 0
	MEM_UNMAP_WITH_TRANSIENT_BOOST UNMAP_VIEW_OF_FILE_FLAGS = 1
	MEM_PRESERVE_PLACEHOLDER       UNMAP_VIEW_OF_FILE_FLAGS = 2
)

// enum
type VIRTUAL_FREE_TYPE uint32

const (
	MEM_DECOMMIT VIRTUAL_FREE_TYPE = 16384
	MEM_RELEASE  VIRTUAL_FREE_TYPE = 32768
)

// enum
// flags
type VIRTUAL_ALLOCATION_TYPE uint32

const (
	MEM_COMMIT              VIRTUAL_ALLOCATION_TYPE = 4096
	MEM_RESERVE             VIRTUAL_ALLOCATION_TYPE = 8192
	MEM_RESET               VIRTUAL_ALLOCATION_TYPE = 524288
	MEM_RESET_UNDO          VIRTUAL_ALLOCATION_TYPE = 16777216
	MEM_REPLACE_PLACEHOLDER VIRTUAL_ALLOCATION_TYPE = 16384
	MEM_LARGE_PAGES         VIRTUAL_ALLOCATION_TYPE = 536870912
	MEM_RESERVE_PLACEHOLDER VIRTUAL_ALLOCATION_TYPE = 262144
	MEM_FREE                VIRTUAL_ALLOCATION_TYPE = 65536
)

// enum
// flags
type LOCAL_ALLOC_FLAGS uint32

const (
	LHND          LOCAL_ALLOC_FLAGS = 66
	LMEM_FIXED    LOCAL_ALLOC_FLAGS = 0
	LMEM_MOVEABLE LOCAL_ALLOC_FLAGS = 2
	LMEM_ZEROINIT LOCAL_ALLOC_FLAGS = 64
	LPTR          LOCAL_ALLOC_FLAGS = 64
	NONZEROLHND   LOCAL_ALLOC_FLAGS = 2
	NONZEROLPTR   LOCAL_ALLOC_FLAGS = 0
)

// enum
// flags
type GLOBAL_ALLOC_FLAGS uint32

const (
	GHND          GLOBAL_ALLOC_FLAGS = 66
	GMEM_FIXED    GLOBAL_ALLOC_FLAGS = 0
	GMEM_MOVEABLE GLOBAL_ALLOC_FLAGS = 2
	GMEM_ZEROINIT GLOBAL_ALLOC_FLAGS = 64
	GPTR          GLOBAL_ALLOC_FLAGS = 64
)

// enum
// flags
type PAGE_TYPE uint32

const (
	MEM_PRIVATE PAGE_TYPE = 131072
	MEM_MAPPED  PAGE_TYPE = 262144
	MEM_IMAGE   PAGE_TYPE = 16777216
)

// enum
type MEMORY_RESOURCE_NOTIFICATION_TYPE int32

const (
	LowMemoryResourceNotification  MEMORY_RESOURCE_NOTIFICATION_TYPE = 0
	HighMemoryResourceNotification MEMORY_RESOURCE_NOTIFICATION_TYPE = 1
)

// enum
type OFFER_PRIORITY int32

const (
	VmOfferPriorityVeryLow     OFFER_PRIORITY = 1
	VmOfferPriorityLow         OFFER_PRIORITY = 2
	VmOfferPriorityBelowNormal OFFER_PRIORITY = 3
	VmOfferPriorityNormal      OFFER_PRIORITY = 4
)

// enum
type WIN32_MEMORY_INFORMATION_CLASS int32

const (
	MemoryRegionInfo WIN32_MEMORY_INFORMATION_CLASS = 0
)

// enum
type WIN32_MEMORY_PARTITION_INFORMATION_CLASS int32

const (
	MemoryPartitionInfo                WIN32_MEMORY_PARTITION_INFORMATION_CLASS = 0
	MemoryPartitionDedicatedMemoryInfo WIN32_MEMORY_PARTITION_INFORMATION_CLASS = 1
)

// enum
type MEM_EXTENDED_PARAMETER_TYPE int32

const (
	MemExtendedParameterInvalidType         MEM_EXTENDED_PARAMETER_TYPE = 0
	MemExtendedParameterAddressRequirements MEM_EXTENDED_PARAMETER_TYPE = 1
	MemExtendedParameterNumaNode            MEM_EXTENDED_PARAMETER_TYPE = 2
	MemExtendedParameterPartitionHandle     MEM_EXTENDED_PARAMETER_TYPE = 3
	MemExtendedParameterUserPhysicalHandle  MEM_EXTENDED_PARAMETER_TYPE = 4
	MemExtendedParameterAttributeFlags      MEM_EXTENDED_PARAMETER_TYPE = 5
	MemExtendedParameterImageMachine        MEM_EXTENDED_PARAMETER_TYPE = 6
	MemExtendedParameterMax                 MEM_EXTENDED_PARAMETER_TYPE = 7
)

// enum
type HEAP_INFORMATION_CLASS int32

const (
	HeapCompatibilityInformation      HEAP_INFORMATION_CLASS = 0
	HeapEnableTerminationOnCorruption HEAP_INFORMATION_CLASS = 1
	HeapOptimizeResources             HEAP_INFORMATION_CLASS = 3
	HeapTag                           HEAP_INFORMATION_CLASS = 7
)

// structs

type PROCESS_HEAP_ENTRY_Anonymous_Block struct {
	HMem       HANDLE
	DwReserved [3]uint32
}

type PROCESS_HEAP_ENTRY_Anonymous_Region struct {
	DwCommittedSize   uint32
	DwUnCommittedSize uint32
	LpFirstBlock      unsafe.Pointer
	LpLastBlock       unsafe.Pointer
}

type PROCESS_HEAP_ENTRY_Anonymous struct {
	Data [3]uint64
}

func (this *PROCESS_HEAP_ENTRY_Anonymous) Block() *PROCESS_HEAP_ENTRY_Anonymous_Block {
	return (*PROCESS_HEAP_ENTRY_Anonymous_Block)(unsafe.Pointer(this))
}

func (this *PROCESS_HEAP_ENTRY_Anonymous) BlockVal() PROCESS_HEAP_ENTRY_Anonymous_Block {
	return *(*PROCESS_HEAP_ENTRY_Anonymous_Block)(unsafe.Pointer(this))
}

func (this *PROCESS_HEAP_ENTRY_Anonymous) Region() *PROCESS_HEAP_ENTRY_Anonymous_Region {
	return (*PROCESS_HEAP_ENTRY_Anonymous_Region)(unsafe.Pointer(this))
}

func (this *PROCESS_HEAP_ENTRY_Anonymous) RegionVal() PROCESS_HEAP_ENTRY_Anonymous_Region {
	return *(*PROCESS_HEAP_ENTRY_Anonymous_Region)(unsafe.Pointer(this))
}

type PROCESS_HEAP_ENTRY struct {
	LpData       unsafe.Pointer
	CbData       uint32
	CbOverhead   byte
	IRegionIndex byte
	WFlags       uint16
	PROCESS_HEAP_ENTRY_Anonymous
}

type HEAP_SUMMARY struct {
	Cb           uint32
	CbAllocated  uintptr
	CbCommitted  uintptr
	CbReserved   uintptr
	CbMaxReserve uintptr
}

type WIN32_MEMORY_RANGE_ENTRY struct {
	VirtualAddress unsafe.Pointer
	NumberOfBytes  uintptr
}

type WIN32_MEMORY_REGION_INFORMATION_Anonymous_Anonymous struct {
	Bitfield_ uint32
}

type WIN32_MEMORY_REGION_INFORMATION_Anonymous struct {
	WIN32_MEMORY_REGION_INFORMATION_Anonymous_Anonymous
}

func (this *WIN32_MEMORY_REGION_INFORMATION_Anonymous) Flags() *uint32 {
	return (*uint32)(unsafe.Pointer(this))
}

func (this *WIN32_MEMORY_REGION_INFORMATION_Anonymous) FlagsVal() uint32 {
	return *(*uint32)(unsafe.Pointer(this))
}

func (this *WIN32_MEMORY_REGION_INFORMATION_Anonymous) Anonymous() *WIN32_MEMORY_REGION_INFORMATION_Anonymous_Anonymous {
	return (*WIN32_MEMORY_REGION_INFORMATION_Anonymous_Anonymous)(unsafe.Pointer(this))
}

func (this *WIN32_MEMORY_REGION_INFORMATION_Anonymous) AnonymousVal() WIN32_MEMORY_REGION_INFORMATION_Anonymous_Anonymous {
	return *(*WIN32_MEMORY_REGION_INFORMATION_Anonymous_Anonymous)(unsafe.Pointer(this))
}

type WIN32_MEMORY_REGION_INFORMATION struct {
	AllocationBase    unsafe.Pointer
	AllocationProtect uint32
	WIN32_MEMORY_REGION_INFORMATION_Anonymous
	RegionSize uintptr
	CommitSize uintptr
}

type WIN32_MEMORY_PARTITION_INFORMATION struct {
	Flags                  uint32
	NumaNode               uint32
	Channel                uint32
	NumberOfNumaNodes      uint32
	ResidentAvailablePages uint64
	CommittedPages         uint64
	CommitLimit            uint64
	PeakCommitment         uint64
	TotalNumberOfPages     uint64
	AvailablePages         uint64
	ZeroPages              uint64
	FreePages              uint64
	StandbyPages           uint64
	Reserved               [16]uint64
	MaximumCommitLimit     uint64
	Reserved2              uint64
	PartitionId            uint32
}

type MEMORY_BASIC_INFORMATION struct {
	BaseAddress       unsafe.Pointer
	AllocationBase    unsafe.Pointer
	AllocationProtect PAGE_PROTECTION_FLAGS
	PartitionId       uint16
	RegionSize        uintptr
	State             VIRTUAL_ALLOCATION_TYPE
	Protect           PAGE_PROTECTION_FLAGS
	Type              PAGE_TYPE
}

type MEMORY_BASIC_INFORMATION32 struct {
	BaseAddress       uint32
	AllocationBase    uint32
	AllocationProtect PAGE_PROTECTION_FLAGS
	RegionSize        uint32
	State             VIRTUAL_ALLOCATION_TYPE
	Protect           PAGE_PROTECTION_FLAGS
	Type              PAGE_TYPE
}

type MEMORY_BASIC_INFORMATION64 struct {
	BaseAddress       uint64
	AllocationBase    uint64
	AllocationProtect PAGE_PROTECTION_FLAGS
	Alignment1__      uint32
	RegionSize        uint64
	State             VIRTUAL_ALLOCATION_TYPE
	Protect           PAGE_PROTECTION_FLAGS
	Type              PAGE_TYPE
	Alignment2__      uint32
}

type CFG_CALL_TARGET_INFO struct {
	Offset uintptr
	Flags  uintptr
}

type MEM_ADDRESS_REQUIREMENTS struct {
	LowestStartingAddress unsafe.Pointer
	HighestEndingAddress  unsafe.Pointer
	Alignment             uintptr
}

type MEM_EXTENDED_PARAMETER_Anonymous1 struct {
	Bitfield_ uint64
}

type MEM_EXTENDED_PARAMETER_Anonymous2 struct {
	Data [1]uint64
}

func (this *MEM_EXTENDED_PARAMETER_Anonymous2) ULong64() *uint64 {
	return (*uint64)(unsafe.Pointer(this))
}

func (this *MEM_EXTENDED_PARAMETER_Anonymous2) ULong64Val() uint64 {
	return *(*uint64)(unsafe.Pointer(this))
}

func (this *MEM_EXTENDED_PARAMETER_Anonymous2) Pointer() *unsafe.Pointer {
	return (*unsafe.Pointer)(unsafe.Pointer(this))
}

func (this *MEM_EXTENDED_PARAMETER_Anonymous2) PointerVal() unsafe.Pointer {
	return *(*unsafe.Pointer)(unsafe.Pointer(this))
}

func (this *MEM_EXTENDED_PARAMETER_Anonymous2) Size() *uintptr {
	return (*uintptr)(unsafe.Pointer(this))
}

func (this *MEM_EXTENDED_PARAMETER_Anonymous2) SizeVal() uintptr {
	return *(*uintptr)(unsafe.Pointer(this))
}

func (this *MEM_EXTENDED_PARAMETER_Anonymous2) Handle() *HANDLE {
	return (*HANDLE)(unsafe.Pointer(this))
}

func (this *MEM_EXTENDED_PARAMETER_Anonymous2) HandleVal() HANDLE {
	return *(*HANDLE)(unsafe.Pointer(this))
}

func (this *MEM_EXTENDED_PARAMETER_Anonymous2) ULong() *uint32 {
	return (*uint32)(unsafe.Pointer(this))
}

func (this *MEM_EXTENDED_PARAMETER_Anonymous2) ULongVal() uint32 {
	return *(*uint32)(unsafe.Pointer(this))
}

type MEM_EXTENDED_PARAMETER struct {
	MEM_EXTENDED_PARAMETER_Anonymous1
	MEM_EXTENDED_PARAMETER_Anonymous2
}

// func types

type PBAD_MEMORY_CALLBACK_ROUTINE = uintptr
type PBAD_MEMORY_CALLBACK_ROUTINE_func = func()

type PSECURE_MEMORY_CACHE_CALLBACK = uintptr
type PSECURE_MEMORY_CACHE_CALLBACK_func = func(Addr unsafe.Pointer, Range uintptr) BOOLEAN

var (
	pHeapCreate                         uintptr
	pHeapDestroy                        uintptr
	pHeapAlloc                          uintptr
	pHeapReAlloc                        uintptr
	pHeapFree                           uintptr
	pHeapSize                           uintptr
	pGetProcessHeap                     uintptr
	pHeapCompact                        uintptr
	pHeapSetInformation                 uintptr
	pHeapValidate                       uintptr
	pHeapSummary                        uintptr
	pGetProcessHeaps                    uintptr
	pHeapLock                           uintptr
	pHeapUnlock                         uintptr
	pHeapWalk                           uintptr
	pHeapQueryInformation               uintptr
	pVirtualAlloc                       uintptr
	pVirtualProtect                     uintptr
	pVirtualFree                        uintptr
	pVirtualQuery                       uintptr
	pVirtualAllocEx                     uintptr
	pVirtualProtectEx                   uintptr
	pVirtualQueryEx                     uintptr
	pCreateFileMappingW                 uintptr
	pOpenFileMappingW                   uintptr
	pMapViewOfFile                      uintptr
	pMapViewOfFileEx                    uintptr
	pVirtualFreeEx                      uintptr
	pFlushViewOfFile                    uintptr
	pUnmapViewOfFile                    uintptr
	pGetLargePageMinimum                uintptr
	pGetProcessWorkingSetSizeEx         uintptr
	pSetProcessWorkingSetSizeEx         uintptr
	pVirtualLock                        uintptr
	pVirtualUnlock                      uintptr
	pGetWriteWatch                      uintptr
	pResetWriteWatch                    uintptr
	pCreateMemoryResourceNotification   uintptr
	pQueryMemoryResourceNotification    uintptr
	pGetSystemFileCacheSize             uintptr
	pSetSystemFileCacheSize             uintptr
	pCreateFileMappingNumaW             uintptr
	pPrefetchVirtualMemory              uintptr
	pCreateFileMappingFromApp           uintptr
	pMapViewOfFileFromApp               uintptr
	pUnmapViewOfFileEx                  uintptr
	pAllocateUserPhysicalPages          uintptr
	pFreeUserPhysicalPages              uintptr
	pMapUserPhysicalPages               uintptr
	pAllocateUserPhysicalPagesNuma      uintptr
	pVirtualAllocExNuma                 uintptr
	pGetMemoryErrorHandlingCapabilities uintptr
	pRegisterBadMemoryNotification      uintptr
	pUnregisterBadMemoryNotification    uintptr
	pOfferVirtualMemory                 uintptr
	pReclaimVirtualMemory               uintptr
	pDiscardVirtualMemory               uintptr
	pRtlCompareMemory                   uintptr
	pGlobalAlloc                        uintptr
	pGlobalReAlloc                      uintptr
	pGlobalSize                         uintptr
	pGlobalUnlock                       uintptr
	pGlobalLock                         uintptr
	pGlobalFlags                        uintptr
	pGlobalHandle                       uintptr
	pGlobalFree                         uintptr
	pLocalAlloc                         uintptr
	pLocalReAlloc                       uintptr
	pLocalLock                          uintptr
	pLocalHandle                        uintptr
	pLocalUnlock                        uintptr
	pLocalSize                          uintptr
	pLocalFlags                         uintptr
	pLocalFree                          uintptr
	pCreateFileMappingA                 uintptr
	pCreateFileMappingNumaA             uintptr
	pOpenFileMappingA                   uintptr
	pMapViewOfFileExNuma                uintptr
	pIsBadReadPtr                       uintptr
	pIsBadWritePtr                      uintptr
	pIsBadCodePtr                       uintptr
	pIsBadStringPtrA                    uintptr
	pIsBadStringPtrW                    uintptr
	pMapUserPhysicalPagesScatter        uintptr
	pAddSecureMemoryCacheCallback       uintptr
	pRemoveSecureMemoryCacheCallback    uintptr
)

func HeapCreate(flOptions HEAP_FLAGS, dwInitialSize uintptr, dwMaximumSize uintptr) (HeapHandle, WIN32_ERROR) {
	addr := lazyAddr(&pHeapCreate, libKernel32, "HeapCreate")
	ret, _, err := syscall.SyscallN(addr, uintptr(flOptions), dwInitialSize, dwMaximumSize)
	return ret, WIN32_ERROR(err)
}

func HeapDestroy(hHeap HeapHandle) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pHeapDestroy, libKernel32, "HeapDestroy")
	ret, _, err := syscall.SyscallN(addr, hHeap)
	return BOOL(ret), WIN32_ERROR(err)
}

func HeapAlloc(hHeap HeapHandle, dwFlags HEAP_FLAGS, dwBytes uintptr) unsafe.Pointer {
	addr := lazyAddr(&pHeapAlloc, libKernel32, "HeapAlloc")
	ret, _, _ := syscall.SyscallN(addr, hHeap, uintptr(dwFlags), dwBytes)
	return (unsafe.Pointer)(ret)
}

func HeapReAlloc(hHeap HeapHandle, dwFlags HEAP_FLAGS, lpMem unsafe.Pointer, dwBytes uintptr) unsafe.Pointer {
	addr := lazyAddr(&pHeapReAlloc, libKernel32, "HeapReAlloc")
	ret, _, _ := syscall.SyscallN(addr, hHeap, uintptr(dwFlags), uintptr(lpMem), dwBytes)
	return (unsafe.Pointer)(ret)
}

func HeapFree(hHeap HeapHandle, dwFlags HEAP_FLAGS, lpMem unsafe.Pointer) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pHeapFree, libKernel32, "HeapFree")
	ret, _, err := syscall.SyscallN(addr, hHeap, uintptr(dwFlags), uintptr(lpMem))
	return BOOL(ret), WIN32_ERROR(err)
}

func HeapSize(hHeap HeapHandle, dwFlags HEAP_FLAGS, lpMem unsafe.Pointer) uintptr {
	addr := lazyAddr(&pHeapSize, libKernel32, "HeapSize")
	ret, _, _ := syscall.SyscallN(addr, hHeap, uintptr(dwFlags), uintptr(lpMem))
	return ret
}

func GetProcessHeap() (HeapHandle, WIN32_ERROR) {
	addr := lazyAddr(&pGetProcessHeap, libKernel32, "GetProcessHeap")
	ret, _, err := syscall.SyscallN(addr)
	return ret, WIN32_ERROR(err)
}

func HeapCompact(hHeap HeapHandle, dwFlags HEAP_FLAGS) (uintptr, WIN32_ERROR) {
	addr := lazyAddr(&pHeapCompact, libKernel32, "HeapCompact")
	ret, _, err := syscall.SyscallN(addr, hHeap, uintptr(dwFlags))
	return ret, WIN32_ERROR(err)
}

func HeapSetInformation(HeapHandle HeapHandle, HeapInformationClass HEAP_INFORMATION_CLASS, HeapInformation unsafe.Pointer, HeapInformationLength uintptr) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pHeapSetInformation, libKernel32, "HeapSetInformation")
	ret, _, err := syscall.SyscallN(addr, HeapHandle, uintptr(HeapInformationClass), uintptr(HeapInformation), HeapInformationLength)
	return BOOL(ret), WIN32_ERROR(err)
}

func HeapValidate(hHeap HeapHandle, dwFlags HEAP_FLAGS, lpMem unsafe.Pointer) BOOL {
	addr := lazyAddr(&pHeapValidate, libKernel32, "HeapValidate")
	ret, _, _ := syscall.SyscallN(addr, hHeap, uintptr(dwFlags), uintptr(lpMem))
	return BOOL(ret)
}

func HeapSummary(hHeap HANDLE, dwFlags uint32, lpSummary *HEAP_SUMMARY) BOOL {
	addr := lazyAddr(&pHeapSummary, libKernel32, "HeapSummary")
	ret, _, _ := syscall.SyscallN(addr, hHeap, uintptr(dwFlags), uintptr(unsafe.Pointer(lpSummary)))
	return BOOL(ret)
}

func GetProcessHeaps(NumberOfHeaps uint32, ProcessHeaps *HeapHandle) (uint32, WIN32_ERROR) {
	addr := lazyAddr(&pGetProcessHeaps, libKernel32, "GetProcessHeaps")
	ret, _, err := syscall.SyscallN(addr, uintptr(NumberOfHeaps), uintptr(unsafe.Pointer(ProcessHeaps)))
	return uint32(ret), WIN32_ERROR(err)
}

func HeapLock(hHeap HeapHandle) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pHeapLock, libKernel32, "HeapLock")
	ret, _, err := syscall.SyscallN(addr, hHeap)
	return BOOL(ret), WIN32_ERROR(err)
}

func HeapUnlock(hHeap HeapHandle) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pHeapUnlock, libKernel32, "HeapUnlock")
	ret, _, err := syscall.SyscallN(addr, hHeap)
	return BOOL(ret), WIN32_ERROR(err)
}

func HeapWalk(hHeap HeapHandle, lpEntry *PROCESS_HEAP_ENTRY) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pHeapWalk, libKernel32, "HeapWalk")
	ret, _, err := syscall.SyscallN(addr, hHeap, uintptr(unsafe.Pointer(lpEntry)))
	return BOOL(ret), WIN32_ERROR(err)
}

func HeapQueryInformation(HeapHandle HeapHandle, HeapInformationClass HEAP_INFORMATION_CLASS, HeapInformation unsafe.Pointer, HeapInformationLength uintptr, ReturnLength *uintptr) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pHeapQueryInformation, libKernel32, "HeapQueryInformation")
	ret, _, err := syscall.SyscallN(addr, HeapHandle, uintptr(HeapInformationClass), uintptr(HeapInformation), HeapInformationLength, uintptr(unsafe.Pointer(ReturnLength)))
	return BOOL(ret), WIN32_ERROR(err)
}

func VirtualAlloc(lpAddress unsafe.Pointer, dwSize uintptr, flAllocationType VIRTUAL_ALLOCATION_TYPE, flProtect PAGE_PROTECTION_FLAGS) (unsafe.Pointer, WIN32_ERROR) {
	addr := lazyAddr(&pVirtualAlloc, libKernel32, "VirtualAlloc")
	ret, _, err := syscall.SyscallN(addr, uintptr(lpAddress), dwSize, uintptr(flAllocationType), uintptr(flProtect))
	return (unsafe.Pointer)(ret), WIN32_ERROR(err)
}

func VirtualProtect(lpAddress unsafe.Pointer, dwSize uintptr, flNewProtect PAGE_PROTECTION_FLAGS, lpflOldProtect *PAGE_PROTECTION_FLAGS) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pVirtualProtect, libKernel32, "VirtualProtect")
	ret, _, err := syscall.SyscallN(addr, uintptr(lpAddress), dwSize, uintptr(flNewProtect), uintptr(unsafe.Pointer(lpflOldProtect)))
	return BOOL(ret), WIN32_ERROR(err)
}

func VirtualFree(lpAddress unsafe.Pointer, dwSize uintptr, dwFreeType VIRTUAL_FREE_TYPE) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pVirtualFree, libKernel32, "VirtualFree")
	ret, _, err := syscall.SyscallN(addr, uintptr(lpAddress), dwSize, uintptr(dwFreeType))
	return BOOL(ret), WIN32_ERROR(err)
}

func VirtualQuery(lpAddress unsafe.Pointer, lpBuffer *MEMORY_BASIC_INFORMATION, dwLength uintptr) (uintptr, WIN32_ERROR) {
	addr := lazyAddr(&pVirtualQuery, libKernel32, "VirtualQuery")
	ret, _, err := syscall.SyscallN(addr, uintptr(lpAddress), uintptr(unsafe.Pointer(lpBuffer)), dwLength)
	return ret, WIN32_ERROR(err)
}

func VirtualAllocEx(hProcess HANDLE, lpAddress unsafe.Pointer, dwSize uintptr, flAllocationType VIRTUAL_ALLOCATION_TYPE, flProtect PAGE_PROTECTION_FLAGS) (unsafe.Pointer, WIN32_ERROR) {
	addr := lazyAddr(&pVirtualAllocEx, libKernel32, "VirtualAllocEx")
	ret, _, err := syscall.SyscallN(addr, hProcess, uintptr(lpAddress), dwSize, uintptr(flAllocationType), uintptr(flProtect))
	return (unsafe.Pointer)(ret), WIN32_ERROR(err)
}

func VirtualProtectEx(hProcess HANDLE, lpAddress unsafe.Pointer, dwSize uintptr, flNewProtect PAGE_PROTECTION_FLAGS, lpflOldProtect *PAGE_PROTECTION_FLAGS) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pVirtualProtectEx, libKernel32, "VirtualProtectEx")
	ret, _, err := syscall.SyscallN(addr, hProcess, uintptr(lpAddress), dwSize, uintptr(flNewProtect), uintptr(unsafe.Pointer(lpflOldProtect)))
	return BOOL(ret), WIN32_ERROR(err)
}

func VirtualQueryEx(hProcess HANDLE, lpAddress unsafe.Pointer, lpBuffer *MEMORY_BASIC_INFORMATION, dwLength uintptr) (uintptr, WIN32_ERROR) {
	addr := lazyAddr(&pVirtualQueryEx, libKernel32, "VirtualQueryEx")
	ret, _, err := syscall.SyscallN(addr, hProcess, uintptr(lpAddress), uintptr(unsafe.Pointer(lpBuffer)), dwLength)
	return ret, WIN32_ERROR(err)
}

var CreateFileMapping = CreateFileMappingW

func CreateFileMappingW(hFile HANDLE, lpFileMappingAttributes *SECURITY_ATTRIBUTES, flProtect PAGE_PROTECTION_FLAGS, dwMaximumSizeHigh uint32, dwMaximumSizeLow uint32, lpName PWSTR) (HANDLE, WIN32_ERROR) {
	addr := lazyAddr(&pCreateFileMappingW, libKernel32, "CreateFileMappingW")
	ret, _, err := syscall.SyscallN(addr, hFile, uintptr(unsafe.Pointer(lpFileMappingAttributes)), uintptr(flProtect), uintptr(dwMaximumSizeHigh), uintptr(dwMaximumSizeLow), uintptr(unsafe.Pointer(lpName)))
	return ret, WIN32_ERROR(err)
}

var OpenFileMapping = OpenFileMappingW

func OpenFileMappingW(dwDesiredAccess uint32, bInheritHandle BOOL, lpName PWSTR) (HANDLE, WIN32_ERROR) {
	addr := lazyAddr(&pOpenFileMappingW, libKernel32, "OpenFileMappingW")
	ret, _, err := syscall.SyscallN(addr, uintptr(dwDesiredAccess), uintptr(bInheritHandle), uintptr(unsafe.Pointer(lpName)))
	return ret, WIN32_ERROR(err)
}

func MapViewOfFile(hFileMappingObject HANDLE, dwDesiredAccess FILE_MAP, dwFileOffsetHigh uint32, dwFileOffsetLow uint32, dwNumberOfBytesToMap uintptr) (unsafe.Pointer, WIN32_ERROR) {
	addr := lazyAddr(&pMapViewOfFile, libKernel32, "MapViewOfFile")
	ret, _, err := syscall.SyscallN(addr, hFileMappingObject, uintptr(dwDesiredAccess), uintptr(dwFileOffsetHigh), uintptr(dwFileOffsetLow), dwNumberOfBytesToMap)
	return (unsafe.Pointer)(ret), WIN32_ERROR(err)
}

func MapViewOfFileEx(hFileMappingObject HANDLE, dwDesiredAccess FILE_MAP, dwFileOffsetHigh uint32, dwFileOffsetLow uint32, dwNumberOfBytesToMap uintptr, lpBaseAddress unsafe.Pointer) (unsafe.Pointer, WIN32_ERROR) {
	addr := lazyAddr(&pMapViewOfFileEx, libKernel32, "MapViewOfFileEx")
	ret, _, err := syscall.SyscallN(addr, hFileMappingObject, uintptr(dwDesiredAccess), uintptr(dwFileOffsetHigh), uintptr(dwFileOffsetLow), dwNumberOfBytesToMap, uintptr(lpBaseAddress))
	return (unsafe.Pointer)(ret), WIN32_ERROR(err)
}

func VirtualFreeEx(hProcess HANDLE, lpAddress unsafe.Pointer, dwSize uintptr, dwFreeType VIRTUAL_FREE_TYPE) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pVirtualFreeEx, libKernel32, "VirtualFreeEx")
	ret, _, err := syscall.SyscallN(addr, hProcess, uintptr(lpAddress), dwSize, uintptr(dwFreeType))
	return BOOL(ret), WIN32_ERROR(err)
}

func FlushViewOfFile(lpBaseAddress unsafe.Pointer, dwNumberOfBytesToFlush uintptr) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pFlushViewOfFile, libKernel32, "FlushViewOfFile")
	ret, _, err := syscall.SyscallN(addr, uintptr(lpBaseAddress), dwNumberOfBytesToFlush)
	return BOOL(ret), WIN32_ERROR(err)
}

func UnmapViewOfFile(lpBaseAddress unsafe.Pointer) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pUnmapViewOfFile, libKernel32, "UnmapViewOfFile")
	ret, _, err := syscall.SyscallN(addr, uintptr(lpBaseAddress))
	return BOOL(ret), WIN32_ERROR(err)
}

func GetLargePageMinimum() uintptr {
	addr := lazyAddr(&pGetLargePageMinimum, libKernel32, "GetLargePageMinimum")
	ret, _, _ := syscall.SyscallN(addr)
	return ret
}

func GetProcessWorkingSetSizeEx(hProcess HANDLE, lpMinimumWorkingSetSize *uintptr, lpMaximumWorkingSetSize *uintptr, Flags *uint32) BOOL {
	addr := lazyAddr(&pGetProcessWorkingSetSizeEx, libKernel32, "GetProcessWorkingSetSizeEx")
	ret, _, _ := syscall.SyscallN(addr, hProcess, uintptr(unsafe.Pointer(lpMinimumWorkingSetSize)), uintptr(unsafe.Pointer(lpMaximumWorkingSetSize)), uintptr(unsafe.Pointer(Flags)))
	return BOOL(ret)
}

func SetProcessWorkingSetSizeEx(hProcess HANDLE, dwMinimumWorkingSetSize uintptr, dwMaximumWorkingSetSize uintptr, Flags uint32) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pSetProcessWorkingSetSizeEx, libKernel32, "SetProcessWorkingSetSizeEx")
	ret, _, err := syscall.SyscallN(addr, hProcess, dwMinimumWorkingSetSize, dwMaximumWorkingSetSize, uintptr(Flags))
	return BOOL(ret), WIN32_ERROR(err)
}

func VirtualLock(lpAddress unsafe.Pointer, dwSize uintptr) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pVirtualLock, libKernel32, "VirtualLock")
	ret, _, err := syscall.SyscallN(addr, uintptr(lpAddress), dwSize)
	return BOOL(ret), WIN32_ERROR(err)
}

func VirtualUnlock(lpAddress unsafe.Pointer, dwSize uintptr) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pVirtualUnlock, libKernel32, "VirtualUnlock")
	ret, _, err := syscall.SyscallN(addr, uintptr(lpAddress), dwSize)
	return BOOL(ret), WIN32_ERROR(err)
}

func GetWriteWatch(dwFlags uint32, lpBaseAddress unsafe.Pointer, dwRegionSize uintptr, lpAddresses unsafe.Pointer, lpdwCount *uintptr, lpdwGranularity *uint32) uint32 {
	addr := lazyAddr(&pGetWriteWatch, libKernel32, "GetWriteWatch")
	ret, _, _ := syscall.SyscallN(addr, uintptr(dwFlags), uintptr(lpBaseAddress), dwRegionSize, uintptr(lpAddresses), uintptr(unsafe.Pointer(lpdwCount)), uintptr(unsafe.Pointer(lpdwGranularity)))
	return uint32(ret)
}

func ResetWriteWatch(lpBaseAddress unsafe.Pointer, dwRegionSize uintptr) uint32 {
	addr := lazyAddr(&pResetWriteWatch, libKernel32, "ResetWriteWatch")
	ret, _, _ := syscall.SyscallN(addr, uintptr(lpBaseAddress), dwRegionSize)
	return uint32(ret)
}

func CreateMemoryResourceNotification(NotificationType MEMORY_RESOURCE_NOTIFICATION_TYPE) (HANDLE, WIN32_ERROR) {
	addr := lazyAddr(&pCreateMemoryResourceNotification, libKernel32, "CreateMemoryResourceNotification")
	ret, _, err := syscall.SyscallN(addr, uintptr(NotificationType))
	return ret, WIN32_ERROR(err)
}

func QueryMemoryResourceNotification(ResourceNotificationHandle HANDLE, ResourceState *BOOL) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pQueryMemoryResourceNotification, libKernel32, "QueryMemoryResourceNotification")
	ret, _, err := syscall.SyscallN(addr, ResourceNotificationHandle, uintptr(unsafe.Pointer(ResourceState)))
	return BOOL(ret), WIN32_ERROR(err)
}

func GetSystemFileCacheSize(lpMinimumFileCacheSize *uintptr, lpMaximumFileCacheSize *uintptr, lpFlags *uint32) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pGetSystemFileCacheSize, libKernel32, "GetSystemFileCacheSize")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpMinimumFileCacheSize)), uintptr(unsafe.Pointer(lpMaximumFileCacheSize)), uintptr(unsafe.Pointer(lpFlags)))
	return BOOL(ret), WIN32_ERROR(err)
}

func SetSystemFileCacheSize(MinimumFileCacheSize uintptr, MaximumFileCacheSize uintptr, Flags uint32) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pSetSystemFileCacheSize, libKernel32, "SetSystemFileCacheSize")
	ret, _, err := syscall.SyscallN(addr, MinimumFileCacheSize, MaximumFileCacheSize, uintptr(Flags))
	return BOOL(ret), WIN32_ERROR(err)
}

var CreateFileMappingNuma = CreateFileMappingNumaW

func CreateFileMappingNumaW(hFile HANDLE, lpFileMappingAttributes *SECURITY_ATTRIBUTES, flProtect PAGE_PROTECTION_FLAGS, dwMaximumSizeHigh uint32, dwMaximumSizeLow uint32, lpName PWSTR, nndPreferred uint32) (HANDLE, WIN32_ERROR) {
	addr := lazyAddr(&pCreateFileMappingNumaW, libKernel32, "CreateFileMappingNumaW")
	ret, _, err := syscall.SyscallN(addr, hFile, uintptr(unsafe.Pointer(lpFileMappingAttributes)), uintptr(flProtect), uintptr(dwMaximumSizeHigh), uintptr(dwMaximumSizeLow), uintptr(unsafe.Pointer(lpName)), uintptr(nndPreferred))
	return ret, WIN32_ERROR(err)
}

func PrefetchVirtualMemory(hProcess HANDLE, NumberOfEntries uintptr, VirtualAddresses *WIN32_MEMORY_RANGE_ENTRY, Flags uint32) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pPrefetchVirtualMemory, libKernel32, "PrefetchVirtualMemory")
	ret, _, err := syscall.SyscallN(addr, hProcess, NumberOfEntries, uintptr(unsafe.Pointer(VirtualAddresses)), uintptr(Flags))
	return BOOL(ret), WIN32_ERROR(err)
}

func CreateFileMappingFromApp(hFile HANDLE, SecurityAttributes *SECURITY_ATTRIBUTES, PageProtection PAGE_PROTECTION_FLAGS, MaximumSize uint64, Name PWSTR) (HANDLE, WIN32_ERROR) {
	addr := lazyAddr(&pCreateFileMappingFromApp, libKernel32, "CreateFileMappingFromApp")
	ret, _, err := syscall.SyscallN(addr, hFile, uintptr(unsafe.Pointer(SecurityAttributes)), uintptr(PageProtection), uintptr(MaximumSize), uintptr(unsafe.Pointer(Name)))
	return ret, WIN32_ERROR(err)
}

func MapViewOfFileFromApp(hFileMappingObject HANDLE, DesiredAccess FILE_MAP, FileOffset uint64, NumberOfBytesToMap uintptr) (unsafe.Pointer, WIN32_ERROR) {
	addr := lazyAddr(&pMapViewOfFileFromApp, libKernel32, "MapViewOfFileFromApp")
	ret, _, err := syscall.SyscallN(addr, hFileMappingObject, uintptr(DesiredAccess), uintptr(FileOffset), NumberOfBytesToMap)
	return (unsafe.Pointer)(ret), WIN32_ERROR(err)
}

func UnmapViewOfFileEx(BaseAddress unsafe.Pointer, UnmapFlags UNMAP_VIEW_OF_FILE_FLAGS) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pUnmapViewOfFileEx, libKernel32, "UnmapViewOfFileEx")
	ret, _, err := syscall.SyscallN(addr, uintptr(BaseAddress), uintptr(UnmapFlags))
	return BOOL(ret), WIN32_ERROR(err)
}

func AllocateUserPhysicalPages(hProcess HANDLE, NumberOfPages *uintptr, PageArray *uintptr) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pAllocateUserPhysicalPages, libKernel32, "AllocateUserPhysicalPages")
	ret, _, err := syscall.SyscallN(addr, hProcess, uintptr(unsafe.Pointer(NumberOfPages)), uintptr(unsafe.Pointer(PageArray)))
	return BOOL(ret), WIN32_ERROR(err)
}

func FreeUserPhysicalPages(hProcess HANDLE, NumberOfPages *uintptr, PageArray *uintptr) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pFreeUserPhysicalPages, libKernel32, "FreeUserPhysicalPages")
	ret, _, err := syscall.SyscallN(addr, hProcess, uintptr(unsafe.Pointer(NumberOfPages)), uintptr(unsafe.Pointer(PageArray)))
	return BOOL(ret), WIN32_ERROR(err)
}

func MapUserPhysicalPages(VirtualAddress unsafe.Pointer, NumberOfPages uintptr, PageArray *uintptr) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pMapUserPhysicalPages, libKernel32, "MapUserPhysicalPages")
	ret, _, err := syscall.SyscallN(addr, uintptr(VirtualAddress), NumberOfPages, uintptr(unsafe.Pointer(PageArray)))
	return BOOL(ret), WIN32_ERROR(err)
}

func AllocateUserPhysicalPagesNuma(hProcess HANDLE, NumberOfPages *uintptr, PageArray *uintptr, nndPreferred uint32) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pAllocateUserPhysicalPagesNuma, libKernel32, "AllocateUserPhysicalPagesNuma")
	ret, _, err := syscall.SyscallN(addr, hProcess, uintptr(unsafe.Pointer(NumberOfPages)), uintptr(unsafe.Pointer(PageArray)), uintptr(nndPreferred))
	return BOOL(ret), WIN32_ERROR(err)
}

func VirtualAllocExNuma(hProcess HANDLE, lpAddress unsafe.Pointer, dwSize uintptr, flAllocationType VIRTUAL_ALLOCATION_TYPE, flProtect uint32, nndPreferred uint32) (unsafe.Pointer, WIN32_ERROR) {
	addr := lazyAddr(&pVirtualAllocExNuma, libKernel32, "VirtualAllocExNuma")
	ret, _, err := syscall.SyscallN(addr, hProcess, uintptr(lpAddress), dwSize, uintptr(flAllocationType), uintptr(flProtect), uintptr(nndPreferred))
	return (unsafe.Pointer)(ret), WIN32_ERROR(err)
}

func GetMemoryErrorHandlingCapabilities(Capabilities *uint32) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pGetMemoryErrorHandlingCapabilities, libKernel32, "GetMemoryErrorHandlingCapabilities")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(Capabilities)))
	return BOOL(ret), WIN32_ERROR(err)
}

func RegisterBadMemoryNotification(Callback PBAD_MEMORY_CALLBACK_ROUTINE) unsafe.Pointer {
	addr := lazyAddr(&pRegisterBadMemoryNotification, libKernel32, "RegisterBadMemoryNotification")
	ret, _, _ := syscall.SyscallN(addr, Callback)
	return (unsafe.Pointer)(ret)
}

func UnregisterBadMemoryNotification(RegistrationHandle unsafe.Pointer) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pUnregisterBadMemoryNotification, libKernel32, "UnregisterBadMemoryNotification")
	ret, _, err := syscall.SyscallN(addr, uintptr(RegistrationHandle))
	return BOOL(ret), WIN32_ERROR(err)
}

func OfferVirtualMemory(VirtualAddress unsafe.Pointer, Size uintptr, Priority OFFER_PRIORITY) uint32 {
	addr := lazyAddr(&pOfferVirtualMemory, libKernel32, "OfferVirtualMemory")
	ret, _, _ := syscall.SyscallN(addr, uintptr(VirtualAddress), Size, uintptr(Priority))
	return uint32(ret)
}

func ReclaimVirtualMemory(VirtualAddress unsafe.Pointer, Size uintptr) uint32 {
	addr := lazyAddr(&pReclaimVirtualMemory, libKernel32, "ReclaimVirtualMemory")
	ret, _, _ := syscall.SyscallN(addr, uintptr(VirtualAddress), Size)
	return uint32(ret)
}

func DiscardVirtualMemory(VirtualAddress unsafe.Pointer, Size uintptr) uint32 {
	addr := lazyAddr(&pDiscardVirtualMemory, libKernel32, "DiscardVirtualMemory")
	ret, _, _ := syscall.SyscallN(addr, uintptr(VirtualAddress), Size)
	return uint32(ret)
}

func RtlCompareMemory(Source1 unsafe.Pointer, Source2 unsafe.Pointer, Length uintptr) uintptr {
	addr := lazyAddr(&pRtlCompareMemory, libKernel32, "RtlCompareMemory")
	ret, _, _ := syscall.SyscallN(addr, uintptr(Source1), uintptr(Source2), Length)
	return ret
}

func GlobalAlloc(uFlags GLOBAL_ALLOC_FLAGS, dwBytes uintptr) (uintptr, WIN32_ERROR) {
	addr := lazyAddr(&pGlobalAlloc, libKernel32, "GlobalAlloc")
	ret, _, err := syscall.SyscallN(addr, uintptr(uFlags), dwBytes)
	return ret, WIN32_ERROR(err)
}

func GlobalReAlloc(hMem uintptr, dwBytes uintptr, uFlags uint32) (uintptr, WIN32_ERROR) {
	addr := lazyAddr(&pGlobalReAlloc, libKernel32, "GlobalReAlloc")
	ret, _, err := syscall.SyscallN(addr, hMem, dwBytes, uintptr(uFlags))
	return ret, WIN32_ERROR(err)
}

func GlobalSize(hMem uintptr) (uintptr, WIN32_ERROR) {
	addr := lazyAddr(&pGlobalSize, libKernel32, "GlobalSize")
	ret, _, err := syscall.SyscallN(addr, hMem)
	return ret, WIN32_ERROR(err)
}

func GlobalUnlock(hMem uintptr) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pGlobalUnlock, libKernel32, "GlobalUnlock")
	ret, _, err := syscall.SyscallN(addr, hMem)
	return BOOL(ret), WIN32_ERROR(err)
}

func GlobalLock(hMem uintptr) (unsafe.Pointer, WIN32_ERROR) {
	addr := lazyAddr(&pGlobalLock, libKernel32, "GlobalLock")
	ret, _, err := syscall.SyscallN(addr, hMem)
	return (unsafe.Pointer)(ret), WIN32_ERROR(err)
}

func GlobalFlags(hMem uintptr) (uint32, WIN32_ERROR) {
	addr := lazyAddr(&pGlobalFlags, libKernel32, "GlobalFlags")
	ret, _, err := syscall.SyscallN(addr, hMem)
	return uint32(ret), WIN32_ERROR(err)
}

func GlobalHandle(pMem unsafe.Pointer) (uintptr, WIN32_ERROR) {
	addr := lazyAddr(&pGlobalHandle, libKernel32, "GlobalHandle")
	ret, _, err := syscall.SyscallN(addr, uintptr(pMem))
	return ret, WIN32_ERROR(err)
}

func GlobalFree(hMem uintptr) (uintptr, WIN32_ERROR) {
	addr := lazyAddr(&pGlobalFree, libKernel32, "GlobalFree")
	ret, _, err := syscall.SyscallN(addr, hMem)
	return ret, WIN32_ERROR(err)
}

func LocalAlloc(uFlags LOCAL_ALLOC_FLAGS, uBytes uintptr) (uintptr, WIN32_ERROR) {
	addr := lazyAddr(&pLocalAlloc, libKernel32, "LocalAlloc")
	ret, _, err := syscall.SyscallN(addr, uintptr(uFlags), uBytes)
	return ret, WIN32_ERROR(err)
}

func LocalReAlloc(hMem uintptr, uBytes uintptr, uFlags uint32) (uintptr, WIN32_ERROR) {
	addr := lazyAddr(&pLocalReAlloc, libKernel32, "LocalReAlloc")
	ret, _, err := syscall.SyscallN(addr, hMem, uBytes, uintptr(uFlags))
	return ret, WIN32_ERROR(err)
}

func LocalLock(hMem uintptr) (unsafe.Pointer, WIN32_ERROR) {
	addr := lazyAddr(&pLocalLock, libKernel32, "LocalLock")
	ret, _, err := syscall.SyscallN(addr, hMem)
	return (unsafe.Pointer)(ret), WIN32_ERROR(err)
}

func LocalHandle(pMem unsafe.Pointer) (uintptr, WIN32_ERROR) {
	addr := lazyAddr(&pLocalHandle, libKernel32, "LocalHandle")
	ret, _, err := syscall.SyscallN(addr, uintptr(pMem))
	return ret, WIN32_ERROR(err)
}

func LocalUnlock(hMem uintptr) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pLocalUnlock, libKernel32, "LocalUnlock")
	ret, _, err := syscall.SyscallN(addr, hMem)
	return BOOL(ret), WIN32_ERROR(err)
}

func LocalSize(hMem uintptr) (uintptr, WIN32_ERROR) {
	addr := lazyAddr(&pLocalSize, libKernel32, "LocalSize")
	ret, _, err := syscall.SyscallN(addr, hMem)
	return ret, WIN32_ERROR(err)
}

func LocalFlags(hMem uintptr) (uint32, WIN32_ERROR) {
	addr := lazyAddr(&pLocalFlags, libKernel32, "LocalFlags")
	ret, _, err := syscall.SyscallN(addr, hMem)
	return uint32(ret), WIN32_ERROR(err)
}

func LocalFree(hMem uintptr) (uintptr, WIN32_ERROR) {
	addr := lazyAddr(&pLocalFree, libKernel32, "LocalFree")
	ret, _, err := syscall.SyscallN(addr, hMem)
	return ret, WIN32_ERROR(err)
}

func CreateFileMappingA(hFile HANDLE, lpFileMappingAttributes *SECURITY_ATTRIBUTES, flProtect PAGE_PROTECTION_FLAGS, dwMaximumSizeHigh uint32, dwMaximumSizeLow uint32, lpName PSTR) (HANDLE, WIN32_ERROR) {
	addr := lazyAddr(&pCreateFileMappingA, libKernel32, "CreateFileMappingA")
	ret, _, err := syscall.SyscallN(addr, hFile, uintptr(unsafe.Pointer(lpFileMappingAttributes)), uintptr(flProtect), uintptr(dwMaximumSizeHigh), uintptr(dwMaximumSizeLow), uintptr(unsafe.Pointer(lpName)))
	return ret, WIN32_ERROR(err)
}

func CreateFileMappingNumaA(hFile HANDLE, lpFileMappingAttributes *SECURITY_ATTRIBUTES, flProtect PAGE_PROTECTION_FLAGS, dwMaximumSizeHigh uint32, dwMaximumSizeLow uint32, lpName PSTR, nndPreferred uint32) (HANDLE, WIN32_ERROR) {
	addr := lazyAddr(&pCreateFileMappingNumaA, libKernel32, "CreateFileMappingNumaA")
	ret, _, err := syscall.SyscallN(addr, hFile, uintptr(unsafe.Pointer(lpFileMappingAttributes)), uintptr(flProtect), uintptr(dwMaximumSizeHigh), uintptr(dwMaximumSizeLow), uintptr(unsafe.Pointer(lpName)), uintptr(nndPreferred))
	return ret, WIN32_ERROR(err)
}

func OpenFileMappingA(dwDesiredAccess uint32, bInheritHandle BOOL, lpName PSTR) (HANDLE, WIN32_ERROR) {
	addr := lazyAddr(&pOpenFileMappingA, libKernel32, "OpenFileMappingA")
	ret, _, err := syscall.SyscallN(addr, uintptr(dwDesiredAccess), uintptr(bInheritHandle), uintptr(unsafe.Pointer(lpName)))
	return ret, WIN32_ERROR(err)
}

func MapViewOfFileExNuma(hFileMappingObject HANDLE, dwDesiredAccess FILE_MAP, dwFileOffsetHigh uint32, dwFileOffsetLow uint32, dwNumberOfBytesToMap uintptr, lpBaseAddress unsafe.Pointer, nndPreferred uint32) (unsafe.Pointer, WIN32_ERROR) {
	addr := lazyAddr(&pMapViewOfFileExNuma, libKernel32, "MapViewOfFileExNuma")
	ret, _, err := syscall.SyscallN(addr, hFileMappingObject, uintptr(dwDesiredAccess), uintptr(dwFileOffsetHigh), uintptr(dwFileOffsetLow), dwNumberOfBytesToMap, uintptr(lpBaseAddress), uintptr(nndPreferred))
	return (unsafe.Pointer)(ret), WIN32_ERROR(err)
}

func IsBadReadPtr(lp unsafe.Pointer, ucb uintptr) BOOL {
	addr := lazyAddr(&pIsBadReadPtr, libKernel32, "IsBadReadPtr")
	ret, _, _ := syscall.SyscallN(addr, uintptr(lp), ucb)
	return BOOL(ret)
}

func IsBadWritePtr(lp unsafe.Pointer, ucb uintptr) BOOL {
	addr := lazyAddr(&pIsBadWritePtr, libKernel32, "IsBadWritePtr")
	ret, _, _ := syscall.SyscallN(addr, uintptr(lp), ucb)
	return BOOL(ret)
}

func IsBadCodePtr(lpfn FARPROC) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pIsBadCodePtr, libKernel32, "IsBadCodePtr")
	ret, _, err := syscall.SyscallN(addr, lpfn)
	return BOOL(ret), WIN32_ERROR(err)
}

func IsBadStringPtrA(lpsz PSTR, ucchMax uintptr) BOOL {
	addr := lazyAddr(&pIsBadStringPtrA, libKernel32, "IsBadStringPtrA")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpsz)), ucchMax)
	return BOOL(ret)
}

var IsBadStringPtr = IsBadStringPtrW

func IsBadStringPtrW(lpsz PWSTR, ucchMax uintptr) BOOL {
	addr := lazyAddr(&pIsBadStringPtrW, libKernel32, "IsBadStringPtrW")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpsz)), ucchMax)
	return BOOL(ret)
}

func MapUserPhysicalPagesScatter(VirtualAddresses unsafe.Pointer, NumberOfPages uintptr, PageArray *uintptr) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pMapUserPhysicalPagesScatter, libKernel32, "MapUserPhysicalPagesScatter")
	ret, _, err := syscall.SyscallN(addr, uintptr(VirtualAddresses), NumberOfPages, uintptr(unsafe.Pointer(PageArray)))
	return BOOL(ret), WIN32_ERROR(err)
}

func AddSecureMemoryCacheCallback(pfnCallBack PSECURE_MEMORY_CACHE_CALLBACK) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pAddSecureMemoryCacheCallback, libKernel32, "AddSecureMemoryCacheCallback")
	ret, _, err := syscall.SyscallN(addr, pfnCallBack)
	return BOOL(ret), WIN32_ERROR(err)
}

func RemoveSecureMemoryCacheCallback(pfnCallBack PSECURE_MEMORY_CACHE_CALLBACK) BOOL {
	addr := lazyAddr(&pRemoveSecureMemoryCacheCallback, libKernel32, "RemoveSecureMemoryCacheCallback")
	ret, _, _ := syscall.SyscallN(addr, pfnCallBack)
	return BOOL(ret)
}
