package win32

import (
	"syscall"
	"unsafe"
)

// structs

type OVERLAPPED_Anonymous_Anonymous struct {
	Offset     uint32
	OffsetHigh uint32
}

type OVERLAPPED_Anonymous struct {
	OVERLAPPED_Anonymous_Anonymous
}

func (this *OVERLAPPED_Anonymous) Anonymous() *OVERLAPPED_Anonymous_Anonymous {
	return (*OVERLAPPED_Anonymous_Anonymous)(unsafe.Pointer(this))
}

func (this *OVERLAPPED_Anonymous) AnonymousVal() OVERLAPPED_Anonymous_Anonymous {
	return *(*OVERLAPPED_Anonymous_Anonymous)(unsafe.Pointer(this))
}

func (this *OVERLAPPED_Anonymous) Pointer() *unsafe.Pointer {
	return (*unsafe.Pointer)(unsafe.Pointer(this))
}

func (this *OVERLAPPED_Anonymous) PointerVal() unsafe.Pointer {
	return *(*unsafe.Pointer)(unsafe.Pointer(this))
}

type OVERLAPPED struct {
	Internal     uintptr
	InternalHigh uintptr
	OVERLAPPED_Anonymous
	HEvent HANDLE
}

type OVERLAPPED_ENTRY struct {
	LpCompletionKey            uintptr
	LpOverlapped               *OVERLAPPED
	Internal                   uintptr
	DwNumberOfBytesTransferred uint32
}

// func types

type LPOVERLAPPED_COMPLETION_ROUTINE = uintptr
type LPOVERLAPPED_COMPLETION_ROUTINE_func = func(dwErrorCode uint32, dwNumberOfBytesTransfered uint32, lpOverlapped *OVERLAPPED)

var (
	pCreateIoCompletionPort      uintptr
	pGetQueuedCompletionStatus   uintptr
	pGetQueuedCompletionStatusEx uintptr
	pPostQueuedCompletionStatus  uintptr
	pDeviceIoControl             uintptr
	pGetOverlappedResult         uintptr
	pCancelIoEx                  uintptr
	pCancelIo                    uintptr
	pGetOverlappedResultEx       uintptr
	pCancelSynchronousIo         uintptr
	pBindIoCompletionCallback    uintptr
)

func CreateIoCompletionPort(FileHandle HANDLE, ExistingCompletionPort HANDLE, CompletionKey uintptr, NumberOfConcurrentThreads uint32) (HANDLE, WIN32_ERROR) {
	addr := lazyAddr(&pCreateIoCompletionPort, libKernel32, "CreateIoCompletionPort")
	ret, _, err := syscall.SyscallN(addr, FileHandle, ExistingCompletionPort, CompletionKey, uintptr(NumberOfConcurrentThreads))
	return ret, WIN32_ERROR(err)
}

func GetQueuedCompletionStatus(CompletionPort HANDLE, lpNumberOfBytesTransferred *uint32, lpCompletionKey *uintptr, lpOverlapped **OVERLAPPED, dwMilliseconds uint32) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pGetQueuedCompletionStatus, libKernel32, "GetQueuedCompletionStatus")
	ret, _, err := syscall.SyscallN(addr, CompletionPort, uintptr(unsafe.Pointer(lpNumberOfBytesTransferred)), uintptr(unsafe.Pointer(lpCompletionKey)), uintptr(unsafe.Pointer(lpOverlapped)), uintptr(dwMilliseconds))
	return BOOL(ret), WIN32_ERROR(err)
}

func GetQueuedCompletionStatusEx(CompletionPort HANDLE, lpCompletionPortEntries *OVERLAPPED_ENTRY, ulCount uint32, ulNumEntriesRemoved *uint32, dwMilliseconds uint32, fAlertable BOOL) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pGetQueuedCompletionStatusEx, libKernel32, "GetQueuedCompletionStatusEx")
	ret, _, err := syscall.SyscallN(addr, CompletionPort, uintptr(unsafe.Pointer(lpCompletionPortEntries)), uintptr(ulCount), uintptr(unsafe.Pointer(ulNumEntriesRemoved)), uintptr(dwMilliseconds), uintptr(fAlertable))
	return BOOL(ret), WIN32_ERROR(err)
}

func PostQueuedCompletionStatus(CompletionPort HANDLE, dwNumberOfBytesTransferred uint32, dwCompletionKey uintptr, lpOverlapped *OVERLAPPED) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pPostQueuedCompletionStatus, libKernel32, "PostQueuedCompletionStatus")
	ret, _, err := syscall.SyscallN(addr, CompletionPort, uintptr(dwNumberOfBytesTransferred), dwCompletionKey, uintptr(unsafe.Pointer(lpOverlapped)))
	return BOOL(ret), WIN32_ERROR(err)
}

func DeviceIoControl(hDevice HANDLE, dwIoControlCode uint32, lpInBuffer unsafe.Pointer, nInBufferSize uint32, lpOutBuffer unsafe.Pointer, nOutBufferSize uint32, lpBytesReturned *uint32, lpOverlapped *OVERLAPPED) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pDeviceIoControl, libKernel32, "DeviceIoControl")
	ret, _, err := syscall.SyscallN(addr, hDevice, uintptr(dwIoControlCode), uintptr(lpInBuffer), uintptr(nInBufferSize), uintptr(lpOutBuffer), uintptr(nOutBufferSize), uintptr(unsafe.Pointer(lpBytesReturned)), uintptr(unsafe.Pointer(lpOverlapped)))
	return BOOL(ret), WIN32_ERROR(err)
}

func GetOverlappedResult(hFile HANDLE, lpOverlapped *OVERLAPPED, lpNumberOfBytesTransferred *uint32, bWait BOOL) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pGetOverlappedResult, libKernel32, "GetOverlappedResult")
	ret, _, err := syscall.SyscallN(addr, hFile, uintptr(unsafe.Pointer(lpOverlapped)), uintptr(unsafe.Pointer(lpNumberOfBytesTransferred)), uintptr(bWait))
	return BOOL(ret), WIN32_ERROR(err)
}

func CancelIoEx(hFile HANDLE, lpOverlapped *OVERLAPPED) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pCancelIoEx, libKernel32, "CancelIoEx")
	ret, _, err := syscall.SyscallN(addr, hFile, uintptr(unsafe.Pointer(lpOverlapped)))
	return BOOL(ret), WIN32_ERROR(err)
}

func CancelIo(hFile HANDLE) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pCancelIo, libKernel32, "CancelIo")
	ret, _, err := syscall.SyscallN(addr, hFile)
	return BOOL(ret), WIN32_ERROR(err)
}

func GetOverlappedResultEx(hFile HANDLE, lpOverlapped *OVERLAPPED, lpNumberOfBytesTransferred *uint32, dwMilliseconds uint32, bAlertable BOOL) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pGetOverlappedResultEx, libKernel32, "GetOverlappedResultEx")
	ret, _, err := syscall.SyscallN(addr, hFile, uintptr(unsafe.Pointer(lpOverlapped)), uintptr(unsafe.Pointer(lpNumberOfBytesTransferred)), uintptr(dwMilliseconds), uintptr(bAlertable))
	return BOOL(ret), WIN32_ERROR(err)
}

func CancelSynchronousIo(hThread HANDLE) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pCancelSynchronousIo, libKernel32, "CancelSynchronousIo")
	ret, _, err := syscall.SyscallN(addr, hThread)
	return BOOL(ret), WIN32_ERROR(err)
}

func BindIoCompletionCallback(FileHandle HANDLE, Function LPOVERLAPPED_COMPLETION_ROUTINE, Flags uint32) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pBindIoCompletionCallback, libKernel32, "BindIoCompletionCallback")
	ret, _, err := syscall.SyscallN(addr, FileHandle, Function, uintptr(Flags))
	return BOOL(ret), WIN32_ERROR(err)
}
