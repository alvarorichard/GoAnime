package win32

import (
	"syscall"
	"unsafe"
)

const (
	MAX_REASON_NAME_LEN                 uint32 = 0x40
	MAX_REASON_DESC_LEN                 uint32 = 0x100
	MAX_REASON_BUGID_LEN                uint32 = 0x20
	MAX_REASON_COMMENT_LEN              uint32 = 0x200
	SHUTDOWN_TYPE_LEN                   uint32 = 0x20
	POLICY_SHOWREASONUI_NEVER           uint32 = 0x0
	POLICY_SHOWREASONUI_ALWAYS          uint32 = 0x1
	POLICY_SHOWREASONUI_WORKSTATIONONLY uint32 = 0x2
	POLICY_SHOWREASONUI_SERVERONLY      uint32 = 0x3
	SNAPSHOT_POLICY_NEVER               uint32 = 0x0
	SNAPSHOT_POLICY_ALWAYS              uint32 = 0x1
	SNAPSHOT_POLICY_UNPLANNED           uint32 = 0x2
	MAX_NUM_REASONS                     uint32 = 0x100
)

// enums

// enum
// flags
type SHUTDOWN_REASON uint32

const (
	SHTDN_REASON_NONE                           SHUTDOWN_REASON = 0
	SHTDN_REASON_FLAG_COMMENT_REQUIRED          SHUTDOWN_REASON = 16777216
	SHTDN_REASON_FLAG_DIRTY_PROBLEM_ID_REQUIRED SHUTDOWN_REASON = 33554432
	SHTDN_REASON_FLAG_CLEAN_UI                  SHUTDOWN_REASON = 67108864
	SHTDN_REASON_FLAG_DIRTY_UI                  SHUTDOWN_REASON = 134217728
	SHTDN_REASON_FLAG_MOBILE_UI_RESERVED        SHUTDOWN_REASON = 268435456
	SHTDN_REASON_FLAG_USER_DEFINED              SHUTDOWN_REASON = 1073741824
	SHTDN_REASON_FLAG_PLANNED                   SHUTDOWN_REASON = 2147483648
	SHTDN_REASON_MAJOR_OTHER                    SHUTDOWN_REASON = 0
	SHTDN_REASON_MAJOR_NONE                     SHUTDOWN_REASON = 0
	SHTDN_REASON_MAJOR_HARDWARE                 SHUTDOWN_REASON = 65536
	SHTDN_REASON_MAJOR_OPERATINGSYSTEM          SHUTDOWN_REASON = 131072
	SHTDN_REASON_MAJOR_SOFTWARE                 SHUTDOWN_REASON = 196608
	SHTDN_REASON_MAJOR_APPLICATION              SHUTDOWN_REASON = 262144
	SHTDN_REASON_MAJOR_SYSTEM                   SHUTDOWN_REASON = 327680
	SHTDN_REASON_MAJOR_POWER                    SHUTDOWN_REASON = 393216
	SHTDN_REASON_MAJOR_LEGACY_API               SHUTDOWN_REASON = 458752
	SHTDN_REASON_MINOR_OTHER                    SHUTDOWN_REASON = 0
	SHTDN_REASON_MINOR_NONE                     SHUTDOWN_REASON = 255
	SHTDN_REASON_MINOR_MAINTENANCE              SHUTDOWN_REASON = 1
	SHTDN_REASON_MINOR_INSTALLATION             SHUTDOWN_REASON = 2
	SHTDN_REASON_MINOR_UPGRADE                  SHUTDOWN_REASON = 3
	SHTDN_REASON_MINOR_RECONFIG                 SHUTDOWN_REASON = 4
	SHTDN_REASON_MINOR_HUNG                     SHUTDOWN_REASON = 5
	SHTDN_REASON_MINOR_UNSTABLE                 SHUTDOWN_REASON = 6
	SHTDN_REASON_MINOR_DISK                     SHUTDOWN_REASON = 7
	SHTDN_REASON_MINOR_PROCESSOR                SHUTDOWN_REASON = 8
	SHTDN_REASON_MINOR_NETWORKCARD              SHUTDOWN_REASON = 9
	SHTDN_REASON_MINOR_POWER_SUPPLY             SHUTDOWN_REASON = 10
	SHTDN_REASON_MINOR_CORDUNPLUGGED            SHUTDOWN_REASON = 11
	SHTDN_REASON_MINOR_ENVIRONMENT              SHUTDOWN_REASON = 12
	SHTDN_REASON_MINOR_HARDWARE_DRIVER          SHUTDOWN_REASON = 13
	SHTDN_REASON_MINOR_OTHERDRIVER              SHUTDOWN_REASON = 14
	SHTDN_REASON_MINOR_BLUESCREEN               SHUTDOWN_REASON = 15
	SHTDN_REASON_MINOR_SERVICEPACK              SHUTDOWN_REASON = 16
	SHTDN_REASON_MINOR_HOTFIX                   SHUTDOWN_REASON = 17
	SHTDN_REASON_MINOR_SECURITYFIX              SHUTDOWN_REASON = 18
	SHTDN_REASON_MINOR_SECURITY                 SHUTDOWN_REASON = 19
	SHTDN_REASON_MINOR_NETWORK_CONNECTIVITY     SHUTDOWN_REASON = 20
	SHTDN_REASON_MINOR_WMI                      SHUTDOWN_REASON = 21
	SHTDN_REASON_MINOR_SERVICEPACK_UNINSTALL    SHUTDOWN_REASON = 22
	SHTDN_REASON_MINOR_HOTFIX_UNINSTALL         SHUTDOWN_REASON = 23
	SHTDN_REASON_MINOR_SECURITYFIX_UNINSTALL    SHUTDOWN_REASON = 24
	SHTDN_REASON_MINOR_MMC                      SHUTDOWN_REASON = 25
	SHTDN_REASON_MINOR_SYSTEMRESTORE            SHUTDOWN_REASON = 26
	SHTDN_REASON_MINOR_TERMSRV                  SHUTDOWN_REASON = 32
	SHTDN_REASON_MINOR_DC_PROMOTION             SHUTDOWN_REASON = 33
	SHTDN_REASON_MINOR_DC_DEMOTION              SHUTDOWN_REASON = 34
	SHTDN_REASON_UNKNOWN                        SHUTDOWN_REASON = 255
	SHTDN_REASON_LEGACY_API                     SHUTDOWN_REASON = 2147942400
	SHTDN_REASON_VALID_BIT_MASK                 SHUTDOWN_REASON = 3238002687
)

// enum
// flags
type SHUTDOWN_FLAGS uint32

const (
	SHUTDOWN_FORCE_OTHERS          SHUTDOWN_FLAGS = 1
	SHUTDOWN_FORCE_SELF            SHUTDOWN_FLAGS = 2
	SHUTDOWN_RESTART               SHUTDOWN_FLAGS = 4
	SHUTDOWN_POWEROFF              SHUTDOWN_FLAGS = 8
	SHUTDOWN_NOREBOOT              SHUTDOWN_FLAGS = 16
	SHUTDOWN_GRACE_OVERRIDE        SHUTDOWN_FLAGS = 32
	SHUTDOWN_INSTALL_UPDATES       SHUTDOWN_FLAGS = 64
	SHUTDOWN_RESTARTAPPS           SHUTDOWN_FLAGS = 128
	SHUTDOWN_SKIP_SVC_PRESHUTDOWN  SHUTDOWN_FLAGS = 256
	SHUTDOWN_HYBRID                SHUTDOWN_FLAGS = 512
	SHUTDOWN_RESTART_BOOTOPTIONS   SHUTDOWN_FLAGS = 1024
	SHUTDOWN_SOFT_REBOOT           SHUTDOWN_FLAGS = 2048
	SHUTDOWN_MOBILE_UI             SHUTDOWN_FLAGS = 4096
	SHUTDOWN_ARSO                  SHUTDOWN_FLAGS = 8192
	SHUTDOWN_CHECK_SAFE_FOR_SERVER SHUTDOWN_FLAGS = 16384
	SHUTDOWN_VAIL_CONTAINER        SHUTDOWN_FLAGS = 32768
	SHUTDOWN_SYSTEM_INITIATED      SHUTDOWN_FLAGS = 65536
)

// enum
type EXIT_WINDOWS_FLAGS uint32

const (
	EWX_HYBRID_SHUTDOWN EXIT_WINDOWS_FLAGS = 4194304
	EWX_LOGOFF          EXIT_WINDOWS_FLAGS = 0
	EWX_POWEROFF        EXIT_WINDOWS_FLAGS = 8
	EWX_REBOOT          EXIT_WINDOWS_FLAGS = 2
	EWX_RESTARTAPPS     EXIT_WINDOWS_FLAGS = 64
	EWX_SHUTDOWN        EXIT_WINDOWS_FLAGS = 1
)

var (
	pInitiateSystemShutdownA    uintptr
	pInitiateSystemShutdownW    uintptr
	pAbortSystemShutdownA       uintptr
	pAbortSystemShutdownW       uintptr
	pInitiateSystemShutdownExA  uintptr
	pInitiateSystemShutdownExW  uintptr
	pInitiateShutdownA          uintptr
	pInitiateShutdownW          uintptr
	pCheckForHiberboot          uintptr
	pExitWindowsEx              uintptr
	pLockWorkStation            uintptr
	pShutdownBlockReasonCreate  uintptr
	pShutdownBlockReasonQuery   uintptr
	pShutdownBlockReasonDestroy uintptr
)

func InitiateSystemShutdownA(lpMachineName PSTR, lpMessage PSTR, dwTimeout uint32, bForceAppsClosed BOOL, bRebootAfterShutdown BOOL) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pInitiateSystemShutdownA, libAdvapi32, "InitiateSystemShutdownA")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpMachineName)), uintptr(unsafe.Pointer(lpMessage)), uintptr(dwTimeout), uintptr(bForceAppsClosed), uintptr(bRebootAfterShutdown))
	return BOOL(ret), WIN32_ERROR(err)
}

var InitiateSystemShutdown = InitiateSystemShutdownW

func InitiateSystemShutdownW(lpMachineName PWSTR, lpMessage PWSTR, dwTimeout uint32, bForceAppsClosed BOOL, bRebootAfterShutdown BOOL) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pInitiateSystemShutdownW, libAdvapi32, "InitiateSystemShutdownW")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpMachineName)), uintptr(unsafe.Pointer(lpMessage)), uintptr(dwTimeout), uintptr(bForceAppsClosed), uintptr(bRebootAfterShutdown))
	return BOOL(ret), WIN32_ERROR(err)
}

func AbortSystemShutdownA(lpMachineName PSTR) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pAbortSystemShutdownA, libAdvapi32, "AbortSystemShutdownA")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpMachineName)))
	return BOOL(ret), WIN32_ERROR(err)
}

var AbortSystemShutdown = AbortSystemShutdownW

func AbortSystemShutdownW(lpMachineName PWSTR) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pAbortSystemShutdownW, libAdvapi32, "AbortSystemShutdownW")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpMachineName)))
	return BOOL(ret), WIN32_ERROR(err)
}

func InitiateSystemShutdownExA(lpMachineName PSTR, lpMessage PSTR, dwTimeout uint32, bForceAppsClosed BOOL, bRebootAfterShutdown BOOL, dwReason SHUTDOWN_REASON) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pInitiateSystemShutdownExA, libAdvapi32, "InitiateSystemShutdownExA")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpMachineName)), uintptr(unsafe.Pointer(lpMessage)), uintptr(dwTimeout), uintptr(bForceAppsClosed), uintptr(bRebootAfterShutdown), uintptr(dwReason))
	return BOOL(ret), WIN32_ERROR(err)
}

var InitiateSystemShutdownEx = InitiateSystemShutdownExW

func InitiateSystemShutdownExW(lpMachineName PWSTR, lpMessage PWSTR, dwTimeout uint32, bForceAppsClosed BOOL, bRebootAfterShutdown BOOL, dwReason SHUTDOWN_REASON) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pInitiateSystemShutdownExW, libAdvapi32, "InitiateSystemShutdownExW")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpMachineName)), uintptr(unsafe.Pointer(lpMessage)), uintptr(dwTimeout), uintptr(bForceAppsClosed), uintptr(bRebootAfterShutdown), uintptr(dwReason))
	return BOOL(ret), WIN32_ERROR(err)
}

func InitiateShutdownA(lpMachineName PSTR, lpMessage PSTR, dwGracePeriod uint32, dwShutdownFlags SHUTDOWN_FLAGS, dwReason SHUTDOWN_REASON) uint32 {
	addr := lazyAddr(&pInitiateShutdownA, libAdvapi32, "InitiateShutdownA")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpMachineName)), uintptr(unsafe.Pointer(lpMessage)), uintptr(dwGracePeriod), uintptr(dwShutdownFlags), uintptr(dwReason))
	return uint32(ret)
}

var InitiateShutdown = InitiateShutdownW

func InitiateShutdownW(lpMachineName PWSTR, lpMessage PWSTR, dwGracePeriod uint32, dwShutdownFlags SHUTDOWN_FLAGS, dwReason SHUTDOWN_REASON) uint32 {
	addr := lazyAddr(&pInitiateShutdownW, libAdvapi32, "InitiateShutdownW")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpMachineName)), uintptr(unsafe.Pointer(lpMessage)), uintptr(dwGracePeriod), uintptr(dwShutdownFlags), uintptr(dwReason))
	return uint32(ret)
}

func CheckForHiberboot(pHiberboot *BOOLEAN, bClearFlag BOOLEAN) uint32 {
	addr := lazyAddr(&pCheckForHiberboot, libAdvapi32, "CheckForHiberboot")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pHiberboot)), uintptr(bClearFlag))
	return uint32(ret)
}

func ExitWindowsEx(uFlags EXIT_WINDOWS_FLAGS, dwReason SHUTDOWN_REASON) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pExitWindowsEx, libUser32, "ExitWindowsEx")
	ret, _, err := syscall.SyscallN(addr, uintptr(uFlags), uintptr(dwReason))
	return BOOL(ret), WIN32_ERROR(err)
}

func LockWorkStation() (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pLockWorkStation, libUser32, "LockWorkStation")
	ret, _, err := syscall.SyscallN(addr)
	return BOOL(ret), WIN32_ERROR(err)
}

func ShutdownBlockReasonCreate(hWnd HWND, pwszReason PWSTR) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pShutdownBlockReasonCreate, libUser32, "ShutdownBlockReasonCreate")
	ret, _, err := syscall.SyscallN(addr, hWnd, uintptr(unsafe.Pointer(pwszReason)))
	return BOOL(ret), WIN32_ERROR(err)
}

func ShutdownBlockReasonQuery(hWnd HWND, pwszBuff PWSTR, pcchBuff *uint32) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pShutdownBlockReasonQuery, libUser32, "ShutdownBlockReasonQuery")
	ret, _, err := syscall.SyscallN(addr, hWnd, uintptr(unsafe.Pointer(pwszBuff)), uintptr(unsafe.Pointer(pcchBuff)))
	return BOOL(ret), WIN32_ERROR(err)
}

func ShutdownBlockReasonDestroy(hWnd HWND) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pShutdownBlockReasonDestroy, libUser32, "ShutdownBlockReasonDestroy")
	ret, _, err := syscall.SyscallN(addr, hWnd)
	return BOOL(ret), WIN32_ERROR(err)
}
