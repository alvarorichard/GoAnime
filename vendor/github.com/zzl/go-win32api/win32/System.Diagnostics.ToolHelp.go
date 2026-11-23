package win32

import (
	"syscall"
	"unsafe"
)

const (
	MAX_MODULE_NAME32 uint32 = 0xff
	HF32_DEFAULT      uint32 = 0x1
	HF32_SHARED       uint32 = 0x2
)

// enums

// enum
// flags
type CREATE_TOOLHELP_SNAPSHOT_FLAGS uint32

const (
	TH32CS_INHERIT      CREATE_TOOLHELP_SNAPSHOT_FLAGS = 2147483648
	TH32CS_SNAPALL      CREATE_TOOLHELP_SNAPSHOT_FLAGS = 15
	TH32CS_SNAPHEAPLIST CREATE_TOOLHELP_SNAPSHOT_FLAGS = 1
	TH32CS_SNAPMODULE   CREATE_TOOLHELP_SNAPSHOT_FLAGS = 8
	TH32CS_SNAPMODULE32 CREATE_TOOLHELP_SNAPSHOT_FLAGS = 16
	TH32CS_SNAPPROCESS  CREATE_TOOLHELP_SNAPSHOT_FLAGS = 2
	TH32CS_SNAPTHREAD   CREATE_TOOLHELP_SNAPSHOT_FLAGS = 4
)

// enum
type HEAPENTRY32_FLAGS uint32

const (
	LF32_FIXED    HEAPENTRY32_FLAGS = 1
	LF32_FREE     HEAPENTRY32_FLAGS = 2
	LF32_MOVEABLE HEAPENTRY32_FLAGS = 4
)

// structs

type HEAPLIST32 struct {
	DwSize        uintptr
	Th32ProcessID uint32
	Th32HeapID    uintptr
	DwFlags       uint32
}

type HEAPENTRY32 struct {
	DwSize        uintptr
	HHandle       HANDLE
	DwAddress     uintptr
	DwBlockSize   uintptr
	DwFlags       HEAPENTRY32_FLAGS
	DwLockCount   uint32
	DwResvd       uint32
	Th32ProcessID uint32
	Th32HeapID    uintptr
}

type PROCESSENTRY32W struct {
	DwSize              uint32
	CntUsage            uint32
	Th32ProcessID       uint32
	Th32DefaultHeapID   uintptr
	Th32ModuleID        uint32
	CntThreads          uint32
	Th32ParentProcessID uint32
	PcPriClassBase      int32
	DwFlags             uint32
	SzExeFile           [260]uint16
}

type PROCESSENTRY32 struct {
	DwSize              uint32
	CntUsage            uint32
	Th32ProcessID       uint32
	Th32DefaultHeapID   uintptr
	Th32ModuleID        uint32
	CntThreads          uint32
	Th32ParentProcessID uint32
	PcPriClassBase      int32
	DwFlags             uint32
	SzExeFile           [260]CHAR
}

type THREADENTRY32 struct {
	DwSize             uint32
	CntUsage           uint32
	Th32ThreadID       uint32
	Th32OwnerProcessID uint32
	TpBasePri          int32
	TpDeltaPri         int32
	DwFlags            uint32
}

type MODULEENTRY32W struct {
	DwSize        uint32
	Th32ModuleID  uint32
	Th32ProcessID uint32
	GlblcntUsage  uint32
	ProccntUsage  uint32
	ModBaseAddr   *byte
	ModBaseSize   uint32
	HModule       HINSTANCE
	SzModule      [256]uint16
	SzExePath     [260]uint16
}

type MODULEENTRY32 struct {
	DwSize        uint32
	Th32ModuleID  uint32
	Th32ProcessID uint32
	GlblcntUsage  uint32
	ProccntUsage  uint32
	ModBaseAddr   *byte
	ModBaseSize   uint32
	HModule       HINSTANCE
	SzModule      [256]CHAR
	SzExePath     [260]CHAR
}

var (
	pCreateToolhelp32Snapshot    uintptr
	pHeap32ListFirst             uintptr
	pHeap32ListNext              uintptr
	pHeap32First                 uintptr
	pHeap32Next                  uintptr
	pToolhelp32ReadProcessMemory uintptr
	pProcess32FirstW             uintptr
	pProcess32NextW              uintptr
	pProcess32First              uintptr
	pProcess32Next               uintptr
	pThread32First               uintptr
	pThread32Next                uintptr
	pModule32FirstW              uintptr
	pModule32NextW               uintptr
	pModule32First               uintptr
	pModule32Next                uintptr
)

func CreateToolhelp32Snapshot(dwFlags CREATE_TOOLHELP_SNAPSHOT_FLAGS, th32ProcessID uint32) (HANDLE, WIN32_ERROR) {
	addr := lazyAddr(&pCreateToolhelp32Snapshot, libKernel32, "CreateToolhelp32Snapshot")
	ret, _, err := syscall.SyscallN(addr, uintptr(dwFlags), uintptr(th32ProcessID))
	return ret, WIN32_ERROR(err)
}

func Heap32ListFirst(hSnapshot HANDLE, lphl *HEAPLIST32) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pHeap32ListFirst, libKernel32, "Heap32ListFirst")
	ret, _, err := syscall.SyscallN(addr, hSnapshot, uintptr(unsafe.Pointer(lphl)))
	return BOOL(ret), WIN32_ERROR(err)
}

func Heap32ListNext(hSnapshot HANDLE, lphl *HEAPLIST32) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pHeap32ListNext, libKernel32, "Heap32ListNext")
	ret, _, err := syscall.SyscallN(addr, hSnapshot, uintptr(unsafe.Pointer(lphl)))
	return BOOL(ret), WIN32_ERROR(err)
}

func Heap32First(lphe *HEAPENTRY32, th32ProcessID uint32, th32HeapID uintptr) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pHeap32First, libKernel32, "Heap32First")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lphe)), uintptr(th32ProcessID), th32HeapID)
	return BOOL(ret), WIN32_ERROR(err)
}

func Heap32Next(lphe *HEAPENTRY32) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pHeap32Next, libKernel32, "Heap32Next")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lphe)))
	return BOOL(ret), WIN32_ERROR(err)
}

func Toolhelp32ReadProcessMemory(th32ProcessID uint32, lpBaseAddress unsafe.Pointer, lpBuffer unsafe.Pointer, cbRead uintptr, lpNumberOfBytesRead *uintptr) BOOL {
	addr := lazyAddr(&pToolhelp32ReadProcessMemory, libKernel32, "Toolhelp32ReadProcessMemory")
	ret, _, _ := syscall.SyscallN(addr, uintptr(th32ProcessID), uintptr(lpBaseAddress), uintptr(lpBuffer), cbRead, uintptr(unsafe.Pointer(lpNumberOfBytesRead)))
	return BOOL(ret)
}

func Process32FirstW(hSnapshot HANDLE, lppe *PROCESSENTRY32W) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pProcess32FirstW, libKernel32, "Process32FirstW")
	ret, _, err := syscall.SyscallN(addr, hSnapshot, uintptr(unsafe.Pointer(lppe)))
	return BOOL(ret), WIN32_ERROR(err)
}

func Process32NextW(hSnapshot HANDLE, lppe *PROCESSENTRY32W) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pProcess32NextW, libKernel32, "Process32NextW")
	ret, _, err := syscall.SyscallN(addr, hSnapshot, uintptr(unsafe.Pointer(lppe)))
	return BOOL(ret), WIN32_ERROR(err)
}

func Process32First(hSnapshot HANDLE, lppe *PROCESSENTRY32) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pProcess32First, libKernel32, "Process32First")
	ret, _, err := syscall.SyscallN(addr, hSnapshot, uintptr(unsafe.Pointer(lppe)))
	return BOOL(ret), WIN32_ERROR(err)
}

func Process32Next(hSnapshot HANDLE, lppe *PROCESSENTRY32) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pProcess32Next, libKernel32, "Process32Next")
	ret, _, err := syscall.SyscallN(addr, hSnapshot, uintptr(unsafe.Pointer(lppe)))
	return BOOL(ret), WIN32_ERROR(err)
}

func Thread32First(hSnapshot HANDLE, lpte *THREADENTRY32) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pThread32First, libKernel32, "Thread32First")
	ret, _, err := syscall.SyscallN(addr, hSnapshot, uintptr(unsafe.Pointer(lpte)))
	return BOOL(ret), WIN32_ERROR(err)
}

func Thread32Next(hSnapshot HANDLE, lpte *THREADENTRY32) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pThread32Next, libKernel32, "Thread32Next")
	ret, _, err := syscall.SyscallN(addr, hSnapshot, uintptr(unsafe.Pointer(lpte)))
	return BOOL(ret), WIN32_ERROR(err)
}

func Module32FirstW(hSnapshot HANDLE, lpme *MODULEENTRY32W) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pModule32FirstW, libKernel32, "Module32FirstW")
	ret, _, err := syscall.SyscallN(addr, hSnapshot, uintptr(unsafe.Pointer(lpme)))
	return BOOL(ret), WIN32_ERROR(err)
}

func Module32NextW(hSnapshot HANDLE, lpme *MODULEENTRY32W) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pModule32NextW, libKernel32, "Module32NextW")
	ret, _, err := syscall.SyscallN(addr, hSnapshot, uintptr(unsafe.Pointer(lpme)))
	return BOOL(ret), WIN32_ERROR(err)
}

func Module32First(hSnapshot HANDLE, lpme *MODULEENTRY32) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pModule32First, libKernel32, "Module32First")
	ret, _, err := syscall.SyscallN(addr, hSnapshot, uintptr(unsafe.Pointer(lpme)))
	return BOOL(ret), WIN32_ERROR(err)
}

func Module32Next(hSnapshot HANDLE, lpme *MODULEENTRY32) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pModule32Next, libKernel32, "Module32Next")
	ret, _, err := syscall.SyscallN(addr, hSnapshot, uintptr(unsafe.Pointer(lpme)))
	return BOOL(ret), WIN32_ERROR(err)
}
