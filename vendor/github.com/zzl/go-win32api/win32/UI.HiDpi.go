package win32

import (
	"syscall"
	"unsafe"
)

type (
	DPI_AWARENESS_CONTEXT = uintptr
)

const (
	DPI_AWARENESS_CONTEXT_UNAWARE              DPI_AWARENESS_CONTEXT = ^DPI_AWARENESS_CONTEXT(0x0)
	DPI_AWARENESS_CONTEXT_SYSTEM_AWARE         DPI_AWARENESS_CONTEXT = ^DPI_AWARENESS_CONTEXT(0x1)
	DPI_AWARENESS_CONTEXT_PER_MONITOR_AWARE    DPI_AWARENESS_CONTEXT = ^DPI_AWARENESS_CONTEXT(0x2)
	DPI_AWARENESS_CONTEXT_PER_MONITOR_AWARE_V2 DPI_AWARENESS_CONTEXT = ^DPI_AWARENESS_CONTEXT(0x3)
	DPI_AWARENESS_CONTEXT_UNAWARE_GDISCALED    DPI_AWARENESS_CONTEXT = ^DPI_AWARENESS_CONTEXT(0x4)
)

// enums

// enum
type DPI_AWARENESS int32

const (
	DPI_AWARENESS_INVALID           DPI_AWARENESS = -1
	DPI_AWARENESS_UNAWARE           DPI_AWARENESS = 0
	DPI_AWARENESS_SYSTEM_AWARE      DPI_AWARENESS = 1
	DPI_AWARENESS_PER_MONITOR_AWARE DPI_AWARENESS = 2
)

// enum
type DPI_HOSTING_BEHAVIOR int32

const (
	DPI_HOSTING_BEHAVIOR_INVALID DPI_HOSTING_BEHAVIOR = -1
	DPI_HOSTING_BEHAVIOR_DEFAULT DPI_HOSTING_BEHAVIOR = 0
	DPI_HOSTING_BEHAVIOR_MIXED   DPI_HOSTING_BEHAVIOR = 1
)

// enum
// flags
type DIALOG_CONTROL_DPI_CHANGE_BEHAVIORS uint32

const (
	DCDC_DEFAULT             DIALOG_CONTROL_DPI_CHANGE_BEHAVIORS = 0
	DCDC_DISABLE_FONT_UPDATE DIALOG_CONTROL_DPI_CHANGE_BEHAVIORS = 1
	DCDC_DISABLE_RELAYOUT    DIALOG_CONTROL_DPI_CHANGE_BEHAVIORS = 2
)

// enum
// flags
type DIALOG_DPI_CHANGE_BEHAVIORS uint32

const (
	DDC_DEFAULT                  DIALOG_DPI_CHANGE_BEHAVIORS = 0
	DDC_DISABLE_ALL              DIALOG_DPI_CHANGE_BEHAVIORS = 1
	DDC_DISABLE_RESIZE           DIALOG_DPI_CHANGE_BEHAVIORS = 2
	DDC_DISABLE_CONTROL_RELAYOUT DIALOG_DPI_CHANGE_BEHAVIORS = 4
)

// enum
type PROCESS_DPI_AWARENESS int32

const (
	PROCESS_DPI_UNAWARE           PROCESS_DPI_AWARENESS = 0
	PROCESS_SYSTEM_DPI_AWARE      PROCESS_DPI_AWARENESS = 1
	PROCESS_PER_MONITOR_DPI_AWARE PROCESS_DPI_AWARENESS = 2
)

// enum
type MONITOR_DPI_TYPE int32

const (
	MDT_EFFECTIVE_DPI MONITOR_DPI_TYPE = 0
	MDT_ANGULAR_DPI   MONITOR_DPI_TYPE = 1
	MDT_RAW_DPI       MONITOR_DPI_TYPE = 2
	MDT_DEFAULT       MONITOR_DPI_TYPE = 0
)

var (
	pOpenThemeDataForDpi                    uintptr
	pSetDialogControlDpiChangeBehavior      uintptr
	pGetDialogControlDpiChangeBehavior      uintptr
	pSetDialogDpiChangeBehavior             uintptr
	pGetDialogDpiChangeBehavior             uintptr
	pGetSystemMetricsForDpi                 uintptr
	pAdjustWindowRectExForDpi               uintptr
	pLogicalToPhysicalPointForPerMonitorDPI uintptr
	pPhysicalToLogicalPointForPerMonitorDPI uintptr
	pSystemParametersInfoForDpi             uintptr
	pSetThreadDpiAwarenessContext           uintptr
	pGetThreadDpiAwarenessContext           uintptr
	pGetWindowDpiAwarenessContext           uintptr
	pGetAwarenessFromDpiAwarenessContext    uintptr
	pGetDpiFromDpiAwarenessContext          uintptr
	pAreDpiAwarenessContextsEqual           uintptr
	pIsValidDpiAwarenessContext             uintptr
	pGetDpiForWindow                        uintptr
	pGetDpiForSystem                        uintptr
	pGetSystemDpiForProcess                 uintptr
	pEnableNonClientDpiScaling              uintptr
	pSetProcessDpiAwarenessContext          uintptr
	pGetDpiAwarenessContextForProcess       uintptr
	pSetThreadDpiHostingBehavior            uintptr
	pGetThreadDpiHostingBehavior            uintptr
	pGetWindowDpiHostingBehavior            uintptr
)

func OpenThemeDataForDpi(hwnd HWND, pszClassList PWSTR, dpi uint32) uintptr {
	addr := lazyAddr(&pOpenThemeDataForDpi, libUxtheme, "OpenThemeDataForDpi")
	ret, _, _ := syscall.SyscallN(addr, hwnd, uintptr(unsafe.Pointer(pszClassList)), uintptr(dpi))
	return ret
}

func SetDialogControlDpiChangeBehavior(hWnd HWND, mask DIALOG_CONTROL_DPI_CHANGE_BEHAVIORS, values DIALOG_CONTROL_DPI_CHANGE_BEHAVIORS) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pSetDialogControlDpiChangeBehavior, libUser32, "SetDialogControlDpiChangeBehavior")
	ret, _, err := syscall.SyscallN(addr, hWnd, uintptr(mask), uintptr(values))
	return BOOL(ret), WIN32_ERROR(err)
}

func GetDialogControlDpiChangeBehavior(hWnd HWND) (DIALOG_CONTROL_DPI_CHANGE_BEHAVIORS, WIN32_ERROR) {
	addr := lazyAddr(&pGetDialogControlDpiChangeBehavior, libUser32, "GetDialogControlDpiChangeBehavior")
	ret, _, err := syscall.SyscallN(addr, hWnd)
	return DIALOG_CONTROL_DPI_CHANGE_BEHAVIORS(ret), WIN32_ERROR(err)
}

func SetDialogDpiChangeBehavior(hDlg HWND, mask DIALOG_DPI_CHANGE_BEHAVIORS, values DIALOG_DPI_CHANGE_BEHAVIORS) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pSetDialogDpiChangeBehavior, libUser32, "SetDialogDpiChangeBehavior")
	ret, _, err := syscall.SyscallN(addr, hDlg, uintptr(mask), uintptr(values))
	return BOOL(ret), WIN32_ERROR(err)
}

func GetDialogDpiChangeBehavior(hDlg HWND) (DIALOG_DPI_CHANGE_BEHAVIORS, WIN32_ERROR) {
	addr := lazyAddr(&pGetDialogDpiChangeBehavior, libUser32, "GetDialogDpiChangeBehavior")
	ret, _, err := syscall.SyscallN(addr, hDlg)
	return DIALOG_DPI_CHANGE_BEHAVIORS(ret), WIN32_ERROR(err)
}

func GetSystemMetricsForDpi(nIndex int32, dpi uint32) (int32, WIN32_ERROR) {
	addr := lazyAddr(&pGetSystemMetricsForDpi, libUser32, "GetSystemMetricsForDpi")
	ret, _, err := syscall.SyscallN(addr, uintptr(nIndex), uintptr(dpi))
	return int32(ret), WIN32_ERROR(err)
}

func AdjustWindowRectExForDpi(lpRect *RECT, dwStyle WINDOW_STYLE, bMenu BOOL, dwExStyle WINDOW_EX_STYLE, dpi uint32) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pAdjustWindowRectExForDpi, libUser32, "AdjustWindowRectExForDpi")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpRect)), uintptr(dwStyle), uintptr(bMenu), uintptr(dwExStyle), uintptr(dpi))
	return BOOL(ret), WIN32_ERROR(err)
}

func LogicalToPhysicalPointForPerMonitorDPI(hWnd HWND, lpPoint *POINT) BOOL {
	addr := lazyAddr(&pLogicalToPhysicalPointForPerMonitorDPI, libUser32, "LogicalToPhysicalPointForPerMonitorDPI")
	ret, _, _ := syscall.SyscallN(addr, hWnd, uintptr(unsafe.Pointer(lpPoint)))
	return BOOL(ret)
}

func PhysicalToLogicalPointForPerMonitorDPI(hWnd HWND, lpPoint *POINT) BOOL {
	addr := lazyAddr(&pPhysicalToLogicalPointForPerMonitorDPI, libUser32, "PhysicalToLogicalPointForPerMonitorDPI")
	ret, _, _ := syscall.SyscallN(addr, hWnd, uintptr(unsafe.Pointer(lpPoint)))
	return BOOL(ret)
}

func SystemParametersInfoForDpi(uiAction uint32, uiParam uint32, pvParam unsafe.Pointer, fWinIni uint32, dpi uint32) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pSystemParametersInfoForDpi, libUser32, "SystemParametersInfoForDpi")
	ret, _, err := syscall.SyscallN(addr, uintptr(uiAction), uintptr(uiParam), uintptr(pvParam), uintptr(fWinIni), uintptr(dpi))
	return BOOL(ret), WIN32_ERROR(err)
}

func SetThreadDpiAwarenessContext(dpiContext DPI_AWARENESS_CONTEXT) DPI_AWARENESS_CONTEXT {
	addr := lazyAddr(&pSetThreadDpiAwarenessContext, libUser32, "SetThreadDpiAwarenessContext")
	ret, _, _ := syscall.SyscallN(addr, dpiContext)
	return ret
}

func GetThreadDpiAwarenessContext() DPI_AWARENESS_CONTEXT {
	addr := lazyAddr(&pGetThreadDpiAwarenessContext, libUser32, "GetThreadDpiAwarenessContext")
	ret, _, _ := syscall.SyscallN(addr)
	return ret
}

func GetWindowDpiAwarenessContext(hwnd HWND) DPI_AWARENESS_CONTEXT {
	addr := lazyAddr(&pGetWindowDpiAwarenessContext, libUser32, "GetWindowDpiAwarenessContext")
	ret, _, _ := syscall.SyscallN(addr, hwnd)
	return ret
}

func GetAwarenessFromDpiAwarenessContext(value DPI_AWARENESS_CONTEXT) DPI_AWARENESS {
	addr := lazyAddr(&pGetAwarenessFromDpiAwarenessContext, libUser32, "GetAwarenessFromDpiAwarenessContext")
	ret, _, _ := syscall.SyscallN(addr, value)
	return DPI_AWARENESS(ret)
}

func GetDpiFromDpiAwarenessContext(value DPI_AWARENESS_CONTEXT) uint32 {
	addr := lazyAddr(&pGetDpiFromDpiAwarenessContext, libUser32, "GetDpiFromDpiAwarenessContext")
	ret, _, _ := syscall.SyscallN(addr, value)
	return uint32(ret)
}

func AreDpiAwarenessContextsEqual(dpiContextA DPI_AWARENESS_CONTEXT, dpiContextB DPI_AWARENESS_CONTEXT) BOOL {
	addr := lazyAddr(&pAreDpiAwarenessContextsEqual, libUser32, "AreDpiAwarenessContextsEqual")
	ret, _, _ := syscall.SyscallN(addr, dpiContextA, dpiContextB)
	return BOOL(ret)
}

func IsValidDpiAwarenessContext(value DPI_AWARENESS_CONTEXT) BOOL {
	addr := lazyAddr(&pIsValidDpiAwarenessContext, libUser32, "IsValidDpiAwarenessContext")
	ret, _, _ := syscall.SyscallN(addr, value)
	return BOOL(ret)
}

func GetDpiForWindow(hwnd HWND) uint32 {
	addr := lazyAddr(&pGetDpiForWindow, libUser32, "GetDpiForWindow")
	ret, _, _ := syscall.SyscallN(addr, hwnd)
	return uint32(ret)
}

func GetDpiForSystem() uint32 {
	addr := lazyAddr(&pGetDpiForSystem, libUser32, "GetDpiForSystem")
	ret, _, _ := syscall.SyscallN(addr)
	return uint32(ret)
}

func GetSystemDpiForProcess(hProcess HANDLE) uint32 {
	addr := lazyAddr(&pGetSystemDpiForProcess, libUser32, "GetSystemDpiForProcess")
	ret, _, _ := syscall.SyscallN(addr, hProcess)
	return uint32(ret)
}

func EnableNonClientDpiScaling(hwnd HWND) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pEnableNonClientDpiScaling, libUser32, "EnableNonClientDpiScaling")
	ret, _, err := syscall.SyscallN(addr, hwnd)
	return BOOL(ret), WIN32_ERROR(err)
}

func SetProcessDpiAwarenessContext(value DPI_AWARENESS_CONTEXT) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pSetProcessDpiAwarenessContext, libUser32, "SetProcessDpiAwarenessContext")
	ret, _, err := syscall.SyscallN(addr, value)
	return BOOL(ret), WIN32_ERROR(err)
}

func GetDpiAwarenessContextForProcess(hProcess HANDLE) DPI_AWARENESS_CONTEXT {
	addr := lazyAddr(&pGetDpiAwarenessContextForProcess, libUser32, "GetDpiAwarenessContextForProcess")
	ret, _, _ := syscall.SyscallN(addr, hProcess)
	return ret
}

func SetThreadDpiHostingBehavior(value DPI_HOSTING_BEHAVIOR) DPI_HOSTING_BEHAVIOR {
	addr := lazyAddr(&pSetThreadDpiHostingBehavior, libUser32, "SetThreadDpiHostingBehavior")
	ret, _, _ := syscall.SyscallN(addr, uintptr(value))
	return DPI_HOSTING_BEHAVIOR(ret)
}

func GetThreadDpiHostingBehavior() DPI_HOSTING_BEHAVIOR {
	addr := lazyAddr(&pGetThreadDpiHostingBehavior, libUser32, "GetThreadDpiHostingBehavior")
	ret, _, _ := syscall.SyscallN(addr)
	return DPI_HOSTING_BEHAVIOR(ret)
}

func GetWindowDpiHostingBehavior(hwnd HWND) DPI_HOSTING_BEHAVIOR {
	addr := lazyAddr(&pGetWindowDpiHostingBehavior, libUser32, "GetWindowDpiHostingBehavior")
	ret, _, _ := syscall.SyscallN(addr, hwnd)
	return DPI_HOSTING_BEHAVIOR(ret)
}
