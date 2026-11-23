package win32

import (
	"golang.org/x/sys/windows"
	"sync/atomic"
)

var (
	libAdvapi32 = windows.NewLazySystemDLL("advapi32.dll")
	libComctl32 = windows.NewLazySystemDLL("comctl32.dll")
	libComdlg32 = windows.NewLazySystemDLL("comdlg32.dll")
	libGdi32    = windows.NewLazySystemDLL("gdi32.dll")
	libGdiplus  = windows.NewLazySystemDLL("gdiplus.dll")
	libKernel32 = windows.NewLazySystemDLL("kernel32.dll")
	libMsimg32  = windows.NewLazySystemDLL("msimg32.dll")
	libOle32    = windows.NewLazySystemDLL("ole32.dll")
	libOleaut32 = windows.NewLazySystemDLL("oleaut32.dll")
	libPdh      = windows.NewLazySystemDLL("pdh.dll")
	libShell32  = windows.NewLazySystemDLL("shell32.dll")
	libShlwapi  = windows.NewLazySystemDLL("shlwapi.dll")
	libUser32   = windows.NewLazySystemDLL("user32.dll")
	libUserenv  = windows.NewLazySystemDLL("userenv.dll")
	libUxtheme  = windows.NewLazySystemDLL("uxtheme.dll")
	libVersion  = windows.NewLazySystemDLL("version.dll")

	//
	libApi_ms_win_core_winrt_string_l1_1_0 = windows.NewLazySystemDLL(
		"api-ms-win-core-winrt-string-l1-1-0.dll")
	libApi_ms_win_core_winrt_l1_1_0 = windows.NewLazySystemDLL(
		"api-ms-win-core-winrt-l1-1-0.dll")
)

func lazyAddr(pAddr *uintptr, lib *windows.LazyDLL, procName string) uintptr {
	addr := atomic.LoadUintptr(pAddr)
	if addr == 0 {
		addr = lib.NewProc(procName).Addr()
		atomic.StoreUintptr(pAddr, addr)
	}
	return addr
}
