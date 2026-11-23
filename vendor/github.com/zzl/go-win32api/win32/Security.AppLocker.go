package win32

import (
	"syscall"
	"unsafe"
)

const (
	SAFER_SCOPEID_MACHINE                   uint32 = 0x1
	SAFER_SCOPEID_USER                      uint32 = 0x2
	SAFER_LEVELID_FULLYTRUSTED              uint32 = 0x40000
	SAFER_LEVELID_NORMALUSER                uint32 = 0x20000
	SAFER_LEVELID_CONSTRAINED               uint32 = 0x10000
	SAFER_LEVELID_UNTRUSTED                 uint32 = 0x1000
	SAFER_LEVELID_DISALLOWED                uint32 = 0x0
	SAFER_LEVEL_OPEN                        uint32 = 0x1
	SAFER_MAX_FRIENDLYNAME_SIZE             uint32 = 0x100
	SAFER_MAX_DESCRIPTION_SIZE              uint32 = 0x100
	SAFER_MAX_HASH_SIZE                     uint32 = 0x40
	SAFER_CRITERIA_IMAGEPATH                uint32 = 0x1
	SAFER_CRITERIA_NOSIGNEDHASH             uint32 = 0x2
	SAFER_CRITERIA_IMAGEHASH                uint32 = 0x4
	SAFER_CRITERIA_AUTHENTICODE             uint32 = 0x8
	SAFER_CRITERIA_URLZONE                  uint32 = 0x10
	SAFER_CRITERIA_APPX_PACKAGE             uint32 = 0x20
	SAFER_CRITERIA_IMAGEPATH_NT             uint32 = 0x1000
	SAFER_POLICY_JOBID_MASK                 uint32 = 0xff000000
	SAFER_POLICY_JOBID_CONSTRAINED          uint32 = 0x4000000
	SAFER_POLICY_JOBID_UNTRUSTED            uint32 = 0x3000000
	SAFER_POLICY_ONLY_EXES                  uint32 = 0x10000
	SAFER_POLICY_SANDBOX_INERT              uint32 = 0x20000
	SAFER_POLICY_HASH_DUPLICATE             uint32 = 0x40000
	SAFER_POLICY_ONLY_AUDIT                 uint32 = 0x1000
	SAFER_POLICY_BLOCK_CLIENT_UI            uint32 = 0x2000
	SAFER_POLICY_UIFLAGS_MASK               uint32 = 0xff
	SAFER_POLICY_UIFLAGS_INFORMATION_PROMPT uint32 = 0x1
	SAFER_POLICY_UIFLAGS_OPTION_PROMPT      uint32 = 0x2
	SAFER_POLICY_UIFLAGS_HIDDEN             uint32 = 0x4
	SRP_POLICY_EXE                          string = "EXE"
	SRP_POLICY_DLL                          string = "DLL"
	SRP_POLICY_MSI                          string = "MSI"
	SRP_POLICY_SCRIPT                       string = "SCRIPT"
	SRP_POLICY_SHELL                        string = "SHELL"
	SRP_POLICY_NOV2                         string = "IGNORESRPV2"
	SRP_POLICY_APPX                         string = "APPX"
	SRP_POLICY_WLDPMSI                      string = "WLDPMSI"
	SRP_POLICY_WLDPSCRIPT                   string = "WLDPSCRIPT"
	SRP_POLICY_WLDPCONFIGCI                 string = "WLDPCONFIGCI"
	SRP_POLICY_MANAGEDINSTALLER             string = "MANAGEDINSTALLER"
)

// enums

// enum
// flags
type SAFER_COMPUTE_TOKEN_FROM_LEVEL_FLAGS uint32

const (
	SAFER_TOKEN_NULL_IF_EQUAL SAFER_COMPUTE_TOKEN_FROM_LEVEL_FLAGS = 1
	SAFER_TOKEN_COMPARE_ONLY  SAFER_COMPUTE_TOKEN_FROM_LEVEL_FLAGS = 2
	SAFER_TOKEN_MAKE_INERT    SAFER_COMPUTE_TOKEN_FROM_LEVEL_FLAGS = 4
	SAFER_TOKEN_WANT_FLAGS    SAFER_COMPUTE_TOKEN_FROM_LEVEL_FLAGS = 8
)

// enum
type SAFER_POLICY_INFO_CLASS int32

const (
	SaferPolicyLevelList                    SAFER_POLICY_INFO_CLASS = 1
	SaferPolicyEnableTransparentEnforcement SAFER_POLICY_INFO_CLASS = 2
	SaferPolicyDefaultLevel                 SAFER_POLICY_INFO_CLASS = 3
	SaferPolicyEvaluateUserScope            SAFER_POLICY_INFO_CLASS = 4
	SaferPolicyScopeFlags                   SAFER_POLICY_INFO_CLASS = 5
	SaferPolicyDefaultLevelFlags            SAFER_POLICY_INFO_CLASS = 6
	SaferPolicyAuthenticodeEnabled          SAFER_POLICY_INFO_CLASS = 7
)

// enum
type SAFER_OBJECT_INFO_CLASS int32

const (
	SaferObjectLevelId                 SAFER_OBJECT_INFO_CLASS = 1
	SaferObjectScopeId                 SAFER_OBJECT_INFO_CLASS = 2
	SaferObjectFriendlyName            SAFER_OBJECT_INFO_CLASS = 3
	SaferObjectDescription             SAFER_OBJECT_INFO_CLASS = 4
	SaferObjectBuiltin                 SAFER_OBJECT_INFO_CLASS = 5
	SaferObjectDisallowed              SAFER_OBJECT_INFO_CLASS = 6
	SaferObjectDisableMaxPrivilege     SAFER_OBJECT_INFO_CLASS = 7
	SaferObjectInvertDeletedPrivileges SAFER_OBJECT_INFO_CLASS = 8
	SaferObjectDeletedPrivileges       SAFER_OBJECT_INFO_CLASS = 9
	SaferObjectDefaultOwner            SAFER_OBJECT_INFO_CLASS = 10
	SaferObjectSidsToDisable           SAFER_OBJECT_INFO_CLASS = 11
	SaferObjectRestrictedSidsInverted  SAFER_OBJECT_INFO_CLASS = 12
	SaferObjectRestrictedSidsAdded     SAFER_OBJECT_INFO_CLASS = 13
	SaferObjectAllIdentificationGuids  SAFER_OBJECT_INFO_CLASS = 14
	SaferObjectSingleIdentification    SAFER_OBJECT_INFO_CLASS = 15
	SaferObjectExtendedError           SAFER_OBJECT_INFO_CLASS = 16
)

// enum
type SAFER_IDENTIFICATION_TYPES int32

const (
	SaferIdentityDefault         SAFER_IDENTIFICATION_TYPES = 0
	SaferIdentityTypeImageName   SAFER_IDENTIFICATION_TYPES = 1
	SaferIdentityTypeImageHash   SAFER_IDENTIFICATION_TYPES = 2
	SaferIdentityTypeUrlZone     SAFER_IDENTIFICATION_TYPES = 3
	SaferIdentityTypeCertificate SAFER_IDENTIFICATION_TYPES = 4
)

// structs

type SAFER_CODE_PROPERTIES_V1 struct {
	CbSize           uint32
	DwCheckFlags     uint32
	ImagePath        PWSTR
	HImageFileHandle HANDLE
	UrlZoneId        uint32
	ImageHash        [64]byte
	DwImageHashSize  uint32
	ImageSize        int64
	HashAlgorithm    uint32
	PByteBlock       *byte
	HWndParent       HWND
	DwWVTUIChoice    uint32
}

type SAFER_CODE_PROPERTIES_V2 struct {
	CbSize             uint32
	DwCheckFlags       uint32
	ImagePath          PWSTR
	HImageFileHandle   HANDLE
	UrlZoneId          uint32
	ImageHash          [64]byte
	DwImageHashSize    uint32
	ImageSize          int64
	HashAlgorithm      uint32
	PByteBlock         *byte
	HWndParent         HWND
	DwWVTUIChoice      uint32
	PackageMoniker     PWSTR
	PackagePublisher   PWSTR
	PackageName        PWSTR
	PackageVersion     uint64
	PackageIsFramework BOOL
}

type SAFER_IDENTIFICATION_HEADER struct {
	DwIdentificationType SAFER_IDENTIFICATION_TYPES
	CbStructSize         uint32
	IdentificationGuid   syscall.GUID
	LastModified         FILETIME
}

type SAFER_PATHNAME_IDENTIFICATION struct {
	Header       SAFER_IDENTIFICATION_HEADER
	Description  [256]uint16
	ImageName    PWSTR
	DwSaferFlags uint32
}

type SAFER_HASH_IDENTIFICATION struct {
	Header        SAFER_IDENTIFICATION_HEADER
	Description   [256]uint16
	FriendlyName  [256]uint16
	HashSize      uint32
	ImageHash     [64]byte
	HashAlgorithm uint32
	ImageSize     int64
	DwSaferFlags  uint32
}

type SAFER_HASH_IDENTIFICATION2 struct {
	HashIdentification SAFER_HASH_IDENTIFICATION
	HashSize           uint32
	ImageHash          [64]byte
	HashAlgorithm      uint32
}

type SAFER_URLZONE_IDENTIFICATION struct {
	Header       SAFER_IDENTIFICATION_HEADER
	UrlZoneId    uint32
	DwSaferFlags uint32
}

var (
	pSaferGetPolicyInformation  uintptr
	pSaferSetPolicyInformation  uintptr
	pSaferCreateLevel           uintptr
	pSaferCloseLevel            uintptr
	pSaferIdentifyLevel         uintptr
	pSaferComputeTokenFromLevel uintptr
	pSaferGetLevelInformation   uintptr
	pSaferSetLevelInformation   uintptr
	pSaferRecordEventLogEntry   uintptr
	pSaferiIsExecutableFileType uintptr
)

func SaferGetPolicyInformation(dwScopeId uint32, SaferPolicyInfoClass SAFER_POLICY_INFO_CLASS, InfoBufferSize uint32, InfoBuffer unsafe.Pointer, InfoBufferRetSize *uint32, lpReserved unsafe.Pointer) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pSaferGetPolicyInformation, libAdvapi32, "SaferGetPolicyInformation")
	ret, _, err := syscall.SyscallN(addr, uintptr(dwScopeId), uintptr(SaferPolicyInfoClass), uintptr(InfoBufferSize), uintptr(InfoBuffer), uintptr(unsafe.Pointer(InfoBufferRetSize)), uintptr(lpReserved))
	return BOOL(ret), WIN32_ERROR(err)
}

func SaferSetPolicyInformation(dwScopeId uint32, SaferPolicyInfoClass SAFER_POLICY_INFO_CLASS, InfoBufferSize uint32, InfoBuffer unsafe.Pointer, lpReserved unsafe.Pointer) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pSaferSetPolicyInformation, libAdvapi32, "SaferSetPolicyInformation")
	ret, _, err := syscall.SyscallN(addr, uintptr(dwScopeId), uintptr(SaferPolicyInfoClass), uintptr(InfoBufferSize), uintptr(InfoBuffer), uintptr(lpReserved))
	return BOOL(ret), WIN32_ERROR(err)
}

func SaferCreateLevel(dwScopeId uint32, dwLevelId uint32, OpenFlags uint32, pLevelHandle *SAFER_LEVEL_HANDLE, lpReserved unsafe.Pointer) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pSaferCreateLevel, libAdvapi32, "SaferCreateLevel")
	ret, _, err := syscall.SyscallN(addr, uintptr(dwScopeId), uintptr(dwLevelId), uintptr(OpenFlags), uintptr(unsafe.Pointer(pLevelHandle)), uintptr(lpReserved))
	return BOOL(ret), WIN32_ERROR(err)
}

func SaferCloseLevel(hLevelHandle SAFER_LEVEL_HANDLE) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pSaferCloseLevel, libAdvapi32, "SaferCloseLevel")
	ret, _, err := syscall.SyscallN(addr, hLevelHandle)
	return BOOL(ret), WIN32_ERROR(err)
}

func SaferIdentifyLevel(dwNumProperties uint32, pCodeProperties *SAFER_CODE_PROPERTIES_V2, pLevelHandle *SAFER_LEVEL_HANDLE, lpReserved unsafe.Pointer) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pSaferIdentifyLevel, libAdvapi32, "SaferIdentifyLevel")
	ret, _, err := syscall.SyscallN(addr, uintptr(dwNumProperties), uintptr(unsafe.Pointer(pCodeProperties)), uintptr(unsafe.Pointer(pLevelHandle)), uintptr(lpReserved))
	return BOOL(ret), WIN32_ERROR(err)
}

func SaferComputeTokenFromLevel(LevelHandle SAFER_LEVEL_HANDLE, InAccessToken HANDLE, OutAccessToken *HANDLE, dwFlags SAFER_COMPUTE_TOKEN_FROM_LEVEL_FLAGS, lpReserved unsafe.Pointer) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pSaferComputeTokenFromLevel, libAdvapi32, "SaferComputeTokenFromLevel")
	ret, _, err := syscall.SyscallN(addr, LevelHandle, InAccessToken, uintptr(unsafe.Pointer(OutAccessToken)), uintptr(dwFlags), uintptr(lpReserved))
	return BOOL(ret), WIN32_ERROR(err)
}

func SaferGetLevelInformation(LevelHandle SAFER_LEVEL_HANDLE, dwInfoType SAFER_OBJECT_INFO_CLASS, lpQueryBuffer unsafe.Pointer, dwInBufferSize uint32, lpdwOutBufferSize *uint32) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pSaferGetLevelInformation, libAdvapi32, "SaferGetLevelInformation")
	ret, _, err := syscall.SyscallN(addr, LevelHandle, uintptr(dwInfoType), uintptr(lpQueryBuffer), uintptr(dwInBufferSize), uintptr(unsafe.Pointer(lpdwOutBufferSize)))
	return BOOL(ret), WIN32_ERROR(err)
}

func SaferSetLevelInformation(LevelHandle SAFER_LEVEL_HANDLE, dwInfoType SAFER_OBJECT_INFO_CLASS, lpQueryBuffer unsafe.Pointer, dwInBufferSize uint32) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pSaferSetLevelInformation, libAdvapi32, "SaferSetLevelInformation")
	ret, _, err := syscall.SyscallN(addr, LevelHandle, uintptr(dwInfoType), uintptr(lpQueryBuffer), uintptr(dwInBufferSize))
	return BOOL(ret), WIN32_ERROR(err)
}

func SaferRecordEventLogEntry(hLevel SAFER_LEVEL_HANDLE, szTargetPath PWSTR, lpReserved unsafe.Pointer) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pSaferRecordEventLogEntry, libAdvapi32, "SaferRecordEventLogEntry")
	ret, _, err := syscall.SyscallN(addr, hLevel, uintptr(unsafe.Pointer(szTargetPath)), uintptr(lpReserved))
	return BOOL(ret), WIN32_ERROR(err)
}

func SaferiIsExecutableFileType(szFullPathname PWSTR, bFromShellExecute BOOLEAN) BOOL {
	addr := lazyAddr(&pSaferiIsExecutableFileType, libAdvapi32, "SaferiIsExecutableFileType")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(szFullPathname)), uintptr(bFromShellExecute))
	return BOOL(ret)
}
