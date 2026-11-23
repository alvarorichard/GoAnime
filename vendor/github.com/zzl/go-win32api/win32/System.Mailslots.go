package win32

import (
	"syscall"
	"unsafe"
)

var (
	pCreateMailslotA uintptr
	pCreateMailslotW uintptr
	pGetMailslotInfo uintptr
	pSetMailslotInfo uintptr
)

func CreateMailslotA(lpName PSTR, nMaxMessageSize uint32, lReadTimeout uint32, lpSecurityAttributes *SECURITY_ATTRIBUTES) (HANDLE, WIN32_ERROR) {
	addr := lazyAddr(&pCreateMailslotA, libKernel32, "CreateMailslotA")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpName)), uintptr(nMaxMessageSize), uintptr(lReadTimeout), uintptr(unsafe.Pointer(lpSecurityAttributes)))
	return ret, WIN32_ERROR(err)
}

var CreateMailslot = CreateMailslotW

func CreateMailslotW(lpName PWSTR, nMaxMessageSize uint32, lReadTimeout uint32, lpSecurityAttributes *SECURITY_ATTRIBUTES) (HANDLE, WIN32_ERROR) {
	addr := lazyAddr(&pCreateMailslotW, libKernel32, "CreateMailslotW")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpName)), uintptr(nMaxMessageSize), uintptr(lReadTimeout), uintptr(unsafe.Pointer(lpSecurityAttributes)))
	return ret, WIN32_ERROR(err)
}

func GetMailslotInfo(hMailslot HANDLE, lpMaxMessageSize *uint32, lpNextSize *uint32, lpMessageCount *uint32, lpReadTimeout *uint32) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pGetMailslotInfo, libKernel32, "GetMailslotInfo")
	ret, _, err := syscall.SyscallN(addr, hMailslot, uintptr(unsafe.Pointer(lpMaxMessageSize)), uintptr(unsafe.Pointer(lpNextSize)), uintptr(unsafe.Pointer(lpMessageCount)), uintptr(unsafe.Pointer(lpReadTimeout)))
	return BOOL(ret), WIN32_ERROR(err)
}

func SetMailslotInfo(hMailslot HANDLE, lReadTimeout uint32) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pSetMailslotInfo, libKernel32, "SetMailslotInfo")
	ret, _, err := syscall.SyscallN(addr, hMailslot, uintptr(lReadTimeout))
	return BOOL(ret), WIN32_ERROR(err)
}
