package win32

import (
	"syscall"
	"unsafe"
)

type (
	HWINWATCH                         = uintptr
	FEATURE_STATE_CHANGE_SUBSCRIPTION = uintptr
	FH_SERVICE_PIPE_HANDLE            = uintptr
)

const (
	WLDP_DLL                                                           string = "WLDP.DLL"
	WLDP_GETLOCKDOWNPOLICY_FN                                          string = "WldpGetLockdownPolicy"
	WLDP_ISCLASSINAPPROVEDLIST_FN                                      string = "WldpIsClassInApprovedList"
	WLDP_SETDYNAMICCODETRUST_FN                                        string = "WldpSetDynamicCodeTrust"
	WLDP_ISDYNAMICCODEPOLICYENABLED_FN                                 string = "WldpIsDynamicCodePolicyEnabled"
	WLDP_QUERYDANAMICCODETRUST_FN                                      string = "WldpQueryDynamicCodeTrust"
	WLDP_QUERYDYNAMICCODETRUST_FN                                      string = "WldpQueryDynamicCodeTrust"
	WLDP_QUERYWINDOWSLOCKDOWNMODE_FN                                   string = "WldpQueryWindowsLockdownMode"
	WLDP_SETWINDOWSLOCKDOWNRESTRICTION_FN                              string = "WldpSetWindowsLockdownRestriction"
	WLDP_QUERYDEVICESECURITYINFORMATION_FN                             string = "WldpQueryDeviceSecurityInformation"
	WLDP_QUERYWINDOWSLOCKDOWNRESTRICTION_FN                            string = "WldpQueryWindowsLockdownRestriction"
	WLDP_ISAPPAPPROVEDBYPOLICY_FN                                      string = "WldpIsAppApprovedByPolicy"
	WLDP_QUERYPOLICYSETTINGENABLED_FN                                  string = "WldpQueryPolicySettingEnabled"
	WLDP_QUERYPOLICYSETTINGENABLED2_FN                                 string = "WldpQueryPolicySettingEnabled2"
	WLDP_ISWCOSPRODUCTIONCONFIGURATION_FN                              string = "WldpIsWcosProductionConfiguration"
	WLDP_RESETWCOSPRODUCTIONCONFIGURATION_FN                           string = "WldpResetWcosProductionConfiguration"
	WLDP_ISPRODUCTIONCONFIGURATION_FN                                  string = "WldpIsProductionConfiguration"
	WLDP_RESETPRODUCTIONCONFIGURATION_FN                               string = "WldpResetProductionConfiguration"
	WLDP_LOCKDOWN_UNDEFINED                                            uint32 = 0x0
	WLDP_LOCKDOWN_DEFINED_FLAG                                         uint32 = 0x80000000
	WLDP_LOCKDOWN_CONFIG_CI_FLAG                                       uint32 = 0x1
	WLDP_LOCKDOWN_CONFIG_CI_AUDIT_FLAG                                 uint32 = 0x2
	WLDP_LOCKDOWN_UMCIENFORCE_FLAG                                     uint32 = 0x4
	WLDP_LOCKDOWN_AUDIT_FLAG                                           uint32 = 0x8
	WLDP_LOCKDOWN_EXCLUSION_FLAG                                       uint32 = 0x10
	WLDP_LOCKDOWN_OFF                                                  uint32 = 0x80000000
	WLDP_HOST_INFORMATION_REVISION                                     uint32 = 0x1
	WLDP_FLAGS_SKIPSIGNATUREVALIDATION                                 uint32 = 0x100
	MAX_TDI_ENTITIES                                                   uint32 = 0x1000
	INFO_CLASS_GENERIC                                                 uint32 = 0x100
	INFO_CLASS_PROTOCOL                                                uint32 = 0x200
	INFO_CLASS_IMPLEMENTATION                                          uint32 = 0x300
	INFO_TYPE_PROVIDER                                                 uint32 = 0x100
	INFO_TYPE_ADDRESS_OBJECT                                           uint32 = 0x200
	INFO_TYPE_CONNECTION                                               uint32 = 0x300
	ENTITY_LIST_ID                                                     uint32 = 0x0
	INVALID_ENTITY_INSTANCE                                            int32  = -1
	CONTEXT_SIZE                                                       uint32 = 0x10
	ENTITY_TYPE_ID                                                     uint32 = 0x1
	CO_TL_NBF                                                          uint32 = 0x400
	CO_TL_SPX                                                          uint32 = 0x402
	CO_TL_TCP                                                          uint32 = 0x404
	CO_TL_SPP                                                          uint32 = 0x406
	CL_TL_NBF                                                          uint32 = 0x401
	CL_TL_UDP                                                          uint32 = 0x403
	ER_ICMP                                                            uint32 = 0x380
	CL_NL_IPX                                                          uint32 = 0x301
	CL_NL_IP                                                           uint32 = 0x303
	AT_ARP                                                             uint32 = 0x280
	AT_NULL                                                            uint32 = 0x282
	IF_GENERIC                                                         uint32 = 0x200
	IF_MIB                                                             uint32 = 0x202
	IOCTL_TDI_TL_IO_CONTROL_ENDPOINT                                   uint32 = 0x210038
	DCI_VERSION                                                        uint32 = 0x100
	DCICREATEPRIMARYSURFACE                                            uint32 = 0x1
	DCICREATEOFFSCREENSURFACE                                          uint32 = 0x2
	DCICREATEOVERLAYSURFACE                                            uint32 = 0x3
	DCIENUMSURFACE                                                     uint32 = 0x4
	DCIESCAPE                                                          uint32 = 0x5
	DCI_OK                                                             uint32 = 0x0
	DCI_FAIL_GENERIC                                                   int32  = -1
	DCI_FAIL_UNSUPPORTEDVERSION                                        int32  = -2
	DCI_FAIL_INVALIDSURFACE                                            int32  = -3
	DCI_FAIL_UNSUPPORTED                                               int32  = -4
	DCI_ERR_CURRENTLYNOTAVAIL                                          int32  = -5
	DCI_ERR_INVALIDRECT                                                int32  = -6
	DCI_ERR_UNSUPPORTEDFORMAT                                          int32  = -7
	DCI_ERR_UNSUPPORTEDMASK                                            int32  = -8
	DCI_ERR_TOOBIGHEIGHT                                               int32  = -9
	DCI_ERR_TOOBIGWIDTH                                                int32  = -10
	DCI_ERR_TOOBIGSIZE                                                 int32  = -11
	DCI_ERR_OUTOFMEMORY                                                int32  = -12
	DCI_ERR_INVALIDPOSITION                                            int32  = -13
	DCI_ERR_INVALIDSTRETCH                                             int32  = -14
	DCI_ERR_INVALIDCLIPLIST                                            int32  = -15
	DCI_ERR_SURFACEISOBSCURED                                          int32  = -16
	DCI_ERR_XALIGN                                                     int32  = -17
	DCI_ERR_YALIGN                                                     int32  = -18
	DCI_ERR_XYALIGN                                                    int32  = -19
	DCI_ERR_WIDTHALIGN                                                 int32  = -20
	DCI_ERR_HEIGHTALIGN                                                int32  = -21
	DCI_STATUS_POINTERCHANGED                                          uint32 = 0x1
	DCI_STATUS_STRIDECHANGED                                           uint32 = 0x2
	DCI_STATUS_FORMATCHANGED                                           uint32 = 0x4
	DCI_STATUS_SURFACEINFOCHANGED                                      uint32 = 0x8
	DCI_STATUS_CHROMAKEYCHANGED                                        uint32 = 0x10
	DCI_STATUS_WASSTILLDRAWING                                         uint32 = 0x20
	DCI_SURFACE_TYPE                                                   uint32 = 0xf
	DCI_PRIMARY                                                        uint32 = 0x0
	DCI_OFFSCREEN                                                      uint32 = 0x1
	DCI_OVERLAY                                                        uint32 = 0x2
	DCI_VISIBLE                                                        uint32 = 0x10
	DCI_CHROMAKEY                                                      uint32 = 0x20
	DCI_1632_ACCESS                                                    uint32 = 0x40
	DCI_DWORDSIZE                                                      uint32 = 0x80
	DCI_DWORDALIGN                                                     uint32 = 0x100
	DCI_WRITEONLY                                                      uint32 = 0x200
	DCI_ASYNC                                                          uint32 = 0x400
	DCI_CAN_STRETCHX                                                   uint32 = 0x1000
	DCI_CAN_STRETCHY                                                   uint32 = 0x2000
	DCI_CAN_STRETCHXN                                                  uint32 = 0x4000
	DCI_CAN_STRETCHYN                                                  uint32 = 0x8000
	DCI_CANOVERLAY                                                     uint32 = 0x10000
	FILE_FLAG_OPEN_REQUIRING_OPLOCK                                    uint32 = 0x40000
	PROGRESS_CONTINUE                                                  uint32 = 0x0
	PROGRESS_CANCEL                                                    uint32 = 0x1
	PROGRESS_STOP                                                      uint32 = 0x2
	PROGRESS_QUIET                                                     uint32 = 0x3
	COPY_FILE_FAIL_IF_EXISTS                                           uint32 = 0x1
	COPY_FILE_RESTARTABLE                                              uint32 = 0x2
	COPY_FILE_OPEN_SOURCE_FOR_WRITE                                    uint32 = 0x4
	COPY_FILE_ALLOW_DECRYPTED_DESTINATION                              uint32 = 0x8
	COPY_FILE_COPY_SYMLINK                                             uint32 = 0x800
	COPY_FILE_NO_BUFFERING                                             uint32 = 0x1000
	COPY_FILE_REQUEST_SECURITY_PRIVILEGES                              uint32 = 0x2000
	COPY_FILE_RESUME_FROM_PAUSE                                        uint32 = 0x4000
	COPY_FILE_NO_OFFLOAD                                               uint32 = 0x40000
	COPY_FILE_IGNORE_EDP_BLOCK                                         uint32 = 0x400000
	COPY_FILE_IGNORE_SOURCE_ENCRYPTION                                 uint32 = 0x800000
	COPY_FILE_DONT_REQUEST_DEST_WRITE_DAC                              uint32 = 0x2000000
	COPY_FILE_REQUEST_COMPRESSED_TRAFFIC                               uint32 = 0x10000000
	COPY_FILE_OPEN_AND_COPY_REPARSE_POINT                              uint32 = 0x200000
	COPY_FILE_DIRECTORY                                                uint32 = 0x80
	COPY_FILE_SKIP_ALTERNATE_STREAMS                                   uint32 = 0x8000
	COPY_FILE_DISABLE_PRE_ALLOCATION                                   uint32 = 0x4000000
	COPY_FILE_ENABLE_LOW_FREE_SPACE_MODE                               uint32 = 0x8000000
	FAIL_FAST_GENERATE_EXCEPTION_ADDRESS                               uint32 = 0x1
	FAIL_FAST_NO_HARD_ERROR_DLG                                        uint32 = 0x2
	DTR_CONTROL_DISABLE                                                uint32 = 0x0
	DTR_CONTROL_ENABLE                                                 uint32 = 0x1
	DTR_CONTROL_HANDSHAKE                                              uint32 = 0x2
	RTS_CONTROL_DISABLE                                                uint32 = 0x0
	RTS_CONTROL_ENABLE                                                 uint32 = 0x1
	RTS_CONTROL_HANDSHAKE                                              uint32 = 0x2
	RTS_CONTROL_TOGGLE                                                 uint32 = 0x3
	GMEM_NOCOMPACT                                                     uint32 = 0x10
	GMEM_NODISCARD                                                     uint32 = 0x20
	GMEM_MODIFY                                                        uint32 = 0x80
	GMEM_DISCARDABLE                                                   uint32 = 0x100
	GMEM_NOT_BANKED                                                    uint32 = 0x1000
	GMEM_SHARE                                                         uint32 = 0x2000
	GMEM_DDESHARE                                                      uint32 = 0x2000
	GMEM_NOTIFY                                                        uint32 = 0x4000
	GMEM_LOWER                                                         uint32 = 0x1000
	GMEM_VALID_FLAGS                                                   uint32 = 0x7f72
	GMEM_INVALID_HANDLE                                                uint32 = 0x8000
	GMEM_DISCARDED                                                     uint32 = 0x4000
	GMEM_LOCKCOUNT                                                     uint32 = 0xff
	THREAD_PRIORITY_ERROR_RETURN                                       uint32 = 0x7fffffff
	VOLUME_NAME_DOS                                                    uint32 = 0x0
	VOLUME_NAME_GUID                                                   uint32 = 0x1
	VOLUME_NAME_NT                                                     uint32 = 0x2
	VOLUME_NAME_NONE                                                   uint32 = 0x4
	DRIVE_UNKNOWN                                                      uint32 = 0x0
	DRIVE_NO_ROOT_DIR                                                  uint32 = 0x1
	DRIVE_REMOVABLE                                                    uint32 = 0x2
	DRIVE_FIXED                                                        uint32 = 0x3
	DRIVE_REMOTE                                                       uint32 = 0x4
	DRIVE_CDROM                                                        uint32 = 0x5
	DRIVE_RAMDISK                                                      uint32 = 0x6
	FILE_TYPE_UNKNOWN                                                  uint32 = 0x0
	FILE_TYPE_DISK                                                     uint32 = 0x1
	FILE_TYPE_CHAR                                                     uint32 = 0x2
	FILE_TYPE_PIPE                                                     uint32 = 0x3
	FILE_TYPE_REMOTE                                                   uint32 = 0x8000
	IGNORE                                                             uint32 = 0x0
	INFINITE                                                           uint32 = 0xffffffff
	CBR_110                                                            uint32 = 0x6e
	CBR_300                                                            uint32 = 0x12c
	CBR_600                                                            uint32 = 0x258
	CBR_1200                                                           uint32 = 0x4b0
	CBR_2400                                                           uint32 = 0x960
	CBR_4800                                                           uint32 = 0x12c0
	CBR_9600                                                           uint32 = 0x2580
	CBR_14400                                                          uint32 = 0x3840
	CBR_19200                                                          uint32 = 0x4b00
	CBR_38400                                                          uint32 = 0x9600
	CBR_56000                                                          uint32 = 0xdac0
	CBR_57600                                                          uint32 = 0xe100
	CBR_115200                                                         uint32 = 0x1c200
	CBR_128000                                                         uint32 = 0x1f400
	CBR_256000                                                         uint32 = 0x3e800
	CE_TXFULL                                                          uint32 = 0x100
	CE_PTO                                                             uint32 = 0x200
	CE_IOE                                                             uint32 = 0x400
	CE_DNS                                                             uint32 = 0x800
	CE_OOP                                                             uint32 = 0x1000
	CE_MODE                                                            uint32 = 0x8000
	IE_BADID                                                           int32  = -1
	IE_OPEN                                                            int32  = -2
	IE_NOPEN                                                           int32  = -3
	IE_MEMORY                                                          int32  = -4
	IE_DEFAULT                                                         int32  = -5
	IE_HARDWARE                                                        int32  = -10
	IE_BYTESIZE                                                        int32  = -11
	IE_BAUDRATE                                                        int32  = -12
	RESETDEV                                                           uint32 = 0x7
	LPTx                                                               uint32 = 0x80
	S_QUEUEEMPTY                                                       uint32 = 0x0
	S_THRESHOLD                                                        uint32 = 0x1
	S_ALLTHRESHOLD                                                     uint32 = 0x2
	S_NORMAL                                                           uint32 = 0x0
	S_LEGATO                                                           uint32 = 0x1
	S_STACCATO                                                         uint32 = 0x2
	S_PERIOD512                                                        uint32 = 0x0
	S_PERIOD1024                                                       uint32 = 0x1
	S_PERIOD2048                                                       uint32 = 0x2
	S_PERIODVOICE                                                      uint32 = 0x3
	S_WHITE512                                                         uint32 = 0x4
	S_WHITE1024                                                        uint32 = 0x5
	S_WHITE2048                                                        uint32 = 0x6
	S_WHITEVOICE                                                       uint32 = 0x7
	S_SERDVNA                                                          int32  = -1
	S_SEROFM                                                           int32  = -2
	S_SERMACT                                                          int32  = -3
	S_SERQFUL                                                          int32  = -4
	S_SERBDNT                                                          int32  = -5
	S_SERDLN                                                           int32  = -6
	S_SERDCC                                                           int32  = -7
	S_SERDTP                                                           int32  = -8
	S_SERDVL                                                           int32  = -9
	S_SERDMD                                                           int32  = -10
	S_SERDSH                                                           int32  = -11
	S_SERDPT                                                           int32  = -12
	S_SERDFQ                                                           int32  = -13
	S_SERDDR                                                           int32  = -14
	S_SERDSR                                                           int32  = -15
	S_SERDST                                                           int32  = -16
	FS_CASE_IS_PRESERVED                                               uint32 = 0x2
	FS_CASE_SENSITIVE                                                  uint32 = 0x1
	FS_UNICODE_STORED_ON_DISK                                          uint32 = 0x4
	FS_PERSISTENT_ACLS                                                 uint32 = 0x8
	FS_VOL_IS_COMPRESSED                                               uint32 = 0x8000
	FS_FILE_COMPRESSION                                                uint32 = 0x10
	FS_FILE_ENCRYPTION                                                 uint32 = 0x20000
	OFS_MAXPATHNAME                                                    uint32 = 0x80
	MAXINTATOM                                                         uint32 = 0xc000
	SCS_32BIT_BINARY                                                   uint32 = 0x0
	SCS_DOS_BINARY                                                     uint32 = 0x1
	SCS_WOW_BINARY                                                     uint32 = 0x2
	SCS_PIF_BINARY                                                     uint32 = 0x3
	SCS_POSIX_BINARY                                                   uint32 = 0x4
	SCS_OS216_BINARY                                                   uint32 = 0x5
	SCS_64BIT_BINARY                                                   uint32 = 0x6
	SCS_THIS_PLATFORM_BINARY                                           uint32 = 0x6
	FIBER_FLAG_FLOAT_SWITCH                                            uint32 = 0x1
	UMS_VERSION                                                        uint32 = 0x100
	FILE_SKIP_COMPLETION_PORT_ON_SUCCESS                               uint32 = 0x1
	FILE_SKIP_SET_EVENT_ON_HANDLE                                      uint32 = 0x2
	CRITICAL_SECTION_NO_DEBUG_INFO                                     uint32 = 0x1000000
	HINSTANCE_ERROR                                                    uint32 = 0x20
	FORMAT_MESSAGE_MAX_WIDTH_MASK                                      uint32 = 0xff
	FILE_ENCRYPTABLE                                                   uint32 = 0x0
	FILE_IS_ENCRYPTED                                                  uint32 = 0x1
	FILE_SYSTEM_ATTR                                                   uint32 = 0x2
	FILE_ROOT_DIR                                                      uint32 = 0x3
	FILE_SYSTEM_DIR                                                    uint32 = 0x4
	FILE_UNKNOWN                                                       uint32 = 0x5
	FILE_SYSTEM_NOT_SUPPORT                                            uint32 = 0x6
	FILE_USER_DISALLOWED                                               uint32 = 0x7
	FILE_READ_ONLY                                                     uint32 = 0x8
	FILE_DIR_DISALLOWED                                                uint32 = 0x9
	EFS_USE_RECOVERY_KEYS                                              uint32 = 0x1
	CREATE_FOR_IMPORT                                                  uint32 = 0x1
	CREATE_FOR_DIR                                                     uint32 = 0x2
	OVERWRITE_HIDDEN                                                   uint32 = 0x4
	EFSRPC_SECURE_ONLY                                                 uint32 = 0x8
	EFS_DROP_ALTERNATE_STREAMS                                         uint32 = 0x10
	BACKUP_INVALID                                                     uint32 = 0x0
	BACKUP_GHOSTED_FILE_EXTENTS                                        uint32 = 0xb
	STREAM_NORMAL_ATTRIBUTE                                            uint32 = 0x0
	STREAM_MODIFIED_WHEN_READ                                          uint32 = 0x1
	STREAM_CONTAINS_SECURITY                                           uint32 = 0x2
	STREAM_CONTAINS_PROPERTIES                                         uint32 = 0x4
	STREAM_SPARSE_ATTRIBUTE                                            uint32 = 0x8
	STREAM_CONTAINS_GHOSTED_FILE_EXTENTS                               uint32 = 0x10
	STARTF_HOLOGRAPHIC                                                 uint32 = 0x40000
	SHUTDOWN_NORETRY                                                   uint32 = 0x1
	PROTECTION_LEVEL_SAME                                              uint32 = 0xffffffff
	PROC_THREAD_ATTRIBUTE_NUMBER                                       uint32 = 0xffff
	PROC_THREAD_ATTRIBUTE_THREAD                                       uint32 = 0x10000
	PROC_THREAD_ATTRIBUTE_INPUT                                        uint32 = 0x20000
	PROC_THREAD_ATTRIBUTE_ADDITIVE                                     uint32 = 0x40000
	PROCESS_CREATION_MITIGATION_POLICY_DEP_ENABLE                      uint32 = 0x1
	PROCESS_CREATION_MITIGATION_POLICY_DEP_ATL_THUNK_ENABLE            uint32 = 0x2
	PROCESS_CREATION_MITIGATION_POLICY_SEHOP_ENABLE                    uint32 = 0x4
	PROCESS_CREATION_CHILD_PROCESS_RESTRICTED                          uint32 = 0x1
	PROCESS_CREATION_CHILD_PROCESS_OVERRIDE                            uint32 = 0x2
	PROCESS_CREATION_CHILD_PROCESS_RESTRICTED_UNLESS_SECURE            uint32 = 0x4
	PROCESS_CREATION_ALL_APPLICATION_PACKAGES_OPT_OUT                  uint32 = 0x1
	PROCESS_CREATION_DESKTOP_APP_BREAKAWAY_ENABLE_PROCESS_TREE         uint32 = 0x1
	PROCESS_CREATION_DESKTOP_APP_BREAKAWAY_DISABLE_PROCESS_TREE        uint32 = 0x2
	PROCESS_CREATION_DESKTOP_APP_BREAKAWAY_OVERRIDE                    uint32 = 0x4
	ATOM_FLAG_GLOBAL                                                   uint32 = 0x2
	GET_SYSTEM_WOW64_DIRECTORY_NAME_A_A                                string = "GetSystemWow64DirectoryA"
	GET_SYSTEM_WOW64_DIRECTORY_NAME_A_W                                string = "GetSystemWow64DirectoryA"
	GET_SYSTEM_WOW64_DIRECTORY_NAME_A_T                                string = "GetSystemWow64DirectoryA"
	GET_SYSTEM_WOW64_DIRECTORY_NAME_W_A                                string = "GetSystemWow64DirectoryW"
	GET_SYSTEM_WOW64_DIRECTORY_NAME_W_W                                string = "GetSystemWow64DirectoryW"
	GET_SYSTEM_WOW64_DIRECTORY_NAME_W_T                                string = "GetSystemWow64DirectoryW"
	GET_SYSTEM_WOW64_DIRECTORY_NAME_T_A                                string = "GetSystemWow64DirectoryW"
	GET_SYSTEM_WOW64_DIRECTORY_NAME_T_W                                string = "GetSystemWow64DirectoryW"
	GET_SYSTEM_WOW64_DIRECTORY_NAME_T_T                                string = "GetSystemWow64DirectoryW"
	BASE_SEARCH_PATH_ENABLE_SAFE_SEARCHMODE                            uint32 = 0x1
	BASE_SEARCH_PATH_DISABLE_SAFE_SEARCHMODE                           uint32 = 0x10000
	BASE_SEARCH_PATH_PERMANENT                                         uint32 = 0x8000
	COPYFILE2_MESSAGE_COPY_OFFLOAD                                     int32  = 1
	COPYFILE2_IO_CYCLE_SIZE_MIN                                        uint32 = 0x1000
	COPYFILE2_IO_CYCLE_SIZE_MAX                                        uint32 = 0x40000000
	COPYFILE2_IO_RATE_MIN                                              uint32 = 0x200
	EVENTLOG_FULL_INFO                                                 uint32 = 0x0
	OPERATION_API_VERSION                                              uint32 = 0x1
	MAX_COMPUTERNAME_LENGTH                                            uint32 = 0xf
	LOGON32_PROVIDER_WINNT35                                           uint32 = 0x1
	LOGON32_PROVIDER_VIRTUAL                                           uint32 = 0x4
	LOGON_ZERO_PASSWORD_BUFFER                                         uint32 = 0x80000000
	HW_PROFILE_GUIDLEN                                                 uint32 = 0x27
	DOCKINFO_UNDOCKED                                                  uint32 = 0x1
	DOCKINFO_DOCKED                                                    uint32 = 0x2
	DOCKINFO_USER_SUPPLIED                                             uint32 = 0x4
	TC_NORMAL                                                          uint32 = 0x0
	TC_HARDERR                                                         uint32 = 0x1
	TC_GP_TRAP                                                         uint32 = 0x2
	TC_SIGNAL                                                          uint32 = 0x3
	AC_LINE_OFFLINE                                                    uint32 = 0x0
	AC_LINE_ONLINE                                                     uint32 = 0x1
	AC_LINE_BACKUP_POWER                                               uint32 = 0x2
	AC_LINE_UNKNOWN                                                    uint32 = 0xff
	BATTERY_FLAG_HIGH                                                  uint32 = 0x1
	BATTERY_FLAG_LOW                                                   uint32 = 0x2
	BATTERY_FLAG_CRITICAL                                              uint32 = 0x4
	BATTERY_FLAG_CHARGING                                              uint32 = 0x8
	BATTERY_FLAG_NO_BATTERY                                            uint32 = 0x80
	BATTERY_FLAG_UNKNOWN                                               uint32 = 0xff
	BATTERY_PERCENTAGE_UNKNOWN                                         uint32 = 0xff
	SYSTEM_STATUS_FLAG_POWER_SAVING_ON                                 uint32 = 0x1
	BATTERY_LIFE_UNKNOWN                                               uint32 = 0xffffffff
	ACTCTX_FLAG_PROCESSOR_ARCHITECTURE_VALID                           uint32 = 0x1
	ACTCTX_FLAG_LANGID_VALID                                           uint32 = 0x2
	ACTCTX_FLAG_ASSEMBLY_DIRECTORY_VALID                               uint32 = 0x4
	ACTCTX_FLAG_RESOURCE_NAME_VALID                                    uint32 = 0x8
	ACTCTX_FLAG_SET_PROCESS_DEFAULT                                    uint32 = 0x10
	ACTCTX_FLAG_APPLICATION_NAME_VALID                                 uint32 = 0x20
	ACTCTX_FLAG_SOURCE_IS_ASSEMBLYREF                                  uint32 = 0x40
	ACTCTX_FLAG_HMODULE_VALID                                          uint32 = 0x80
	DEACTIVATE_ACTCTX_FLAG_FORCE_EARLY_DEACTIVATION                    uint32 = 0x1
	FIND_ACTCTX_SECTION_KEY_RETURN_HACTCTX                             uint32 = 0x1
	FIND_ACTCTX_SECTION_KEY_RETURN_FLAGS                               uint32 = 0x2
	FIND_ACTCTX_SECTION_KEY_RETURN_ASSEMBLY_METADATA                   uint32 = 0x4
	ACTIVATION_CONTEXT_BASIC_INFORMATION_DEFINED                       uint32 = 0x1
	QUERY_ACTCTX_FLAG_USE_ACTIVE_ACTCTX                                uint32 = 0x4
	QUERY_ACTCTX_FLAG_ACTCTX_IS_HMODULE                                uint32 = 0x8
	QUERY_ACTCTX_FLAG_ACTCTX_IS_ADDRESS                                uint32 = 0x10
	QUERY_ACTCTX_FLAG_NO_ADDREF                                        uint32 = 0x80000000
	RESTART_MAX_CMD_LINE                                               uint32 = 0x400
	RECOVERY_DEFAULT_PING_INTERVAL                                     uint32 = 0x1388
	FILE_RENAME_FLAG_REPLACE_IF_EXISTS                                 uint32 = 0x1
	FILE_RENAME_FLAG_POSIX_SEMANTICS                                   uint32 = 0x2
	FILE_RENAME_FLAG_SUPPRESS_PIN_STATE_INHERITANCE                    uint32 = 0x4
	FILE_DISPOSITION_FLAG_DO_NOT_DELETE                                uint32 = 0x0
	FILE_DISPOSITION_FLAG_DELETE                                       uint32 = 0x1
	FILE_DISPOSITION_FLAG_POSIX_SEMANTICS                              uint32 = 0x2
	FILE_DISPOSITION_FLAG_FORCE_IMAGE_SECTION_CHECK                    uint32 = 0x4
	FILE_DISPOSITION_FLAG_ON_CLOSE                                     uint32 = 0x8
	FILE_DISPOSITION_FLAG_IGNORE_READONLY_ATTRIBUTE                    uint32 = 0x10
	STORAGE_INFO_FLAGS_ALIGNED_DEVICE                                  uint32 = 0x1
	STORAGE_INFO_FLAGS_PARTITION_ALIGNED_ON_DEVICE                     uint32 = 0x2
	STORAGE_INFO_OFFSET_UNKNOWN                                        uint32 = 0xffffffff
	REMOTE_PROTOCOL_INFO_FLAG_LOOPBACK                                 uint32 = 0x1
	REMOTE_PROTOCOL_INFO_FLAG_OFFLINE                                  uint32 = 0x2
	REMOTE_PROTOCOL_INFO_FLAG_PERSISTENT_HANDLE                        uint32 = 0x4
	RPI_FLAG_SMB2_SHARECAP_TIMEWARP                                    uint32 = 0x2
	RPI_FLAG_SMB2_SHARECAP_DFS                                         uint32 = 0x8
	RPI_FLAG_SMB2_SHARECAP_CONTINUOUS_AVAILABILITY                     uint32 = 0x10
	RPI_FLAG_SMB2_SHARECAP_SCALEOUT                                    uint32 = 0x20
	RPI_FLAG_SMB2_SHARECAP_CLUSTER                                     uint32 = 0x40
	RPI_SMB2_FLAG_SERVERCAP_DFS                                        uint32 = 0x1
	RPI_SMB2_FLAG_SERVERCAP_LEASING                                    uint32 = 0x2
	RPI_SMB2_FLAG_SERVERCAP_LARGEMTU                                   uint32 = 0x4
	RPI_SMB2_FLAG_SERVERCAP_MULTICHANNEL                               uint32 = 0x8
	RPI_SMB2_FLAG_SERVERCAP_PERSISTENT_HANDLES                         uint32 = 0x10
	RPI_SMB2_FLAG_SERVERCAP_DIRECTORY_LEASING                          uint32 = 0x20
	MICROSOFT_WINDOWS_WINBASE_H_DEFINE_INTERLOCKED_CPLUSPLUS_OVERLOADS uint32 = 0x0
	MICROSOFT_WINBASE_H_DEFINE_INTERLOCKED_CPLUSPLUS_OVERLOADS         uint32 = 0x0
	CODEINTEGRITY_OPTION_ENABLED                                       uint32 = 0x1
	CODEINTEGRITY_OPTION_TESTSIGN                                      uint32 = 0x2
	CODEINTEGRITY_OPTION_UMCI_ENABLED                                  uint32 = 0x4
	CODEINTEGRITY_OPTION_UMCI_AUDITMODE_ENABLED                        uint32 = 0x8
	CODEINTEGRITY_OPTION_UMCI_EXCLUSIONPATHS_ENABLED                   uint32 = 0x10
	CODEINTEGRITY_OPTION_TEST_BUILD                                    uint32 = 0x20
	CODEINTEGRITY_OPTION_PREPRODUCTION_BUILD                           uint32 = 0x40
	CODEINTEGRITY_OPTION_DEBUGMODE_ENABLED                             uint32 = 0x80
	CODEINTEGRITY_OPTION_FLIGHT_BUILD                                  uint32 = 0x100
	CODEINTEGRITY_OPTION_FLIGHTING_ENABLED                             uint32 = 0x200
	CODEINTEGRITY_OPTION_HVCI_KMCI_ENABLED                             uint32 = 0x400
	CODEINTEGRITY_OPTION_HVCI_KMCI_AUDITMODE_ENABLED                   uint32 = 0x800
	CODEINTEGRITY_OPTION_HVCI_KMCI_STRICTMODE_ENABLED                  uint32 = 0x1000
	CODEINTEGRITY_OPTION_HVCI_IUM_ENABLED                              uint32 = 0x2000
	FILE_MAXIMUM_DISPOSITION                                           uint32 = 0x5
	FILE_DIRECTORY_FILE                                                uint32 = 0x1
	FILE_WRITE_THROUGH                                                 uint32 = 0x2
	FILE_SEQUENTIAL_ONLY                                               uint32 = 0x4
	FILE_NO_INTERMEDIATE_BUFFERING                                     uint32 = 0x8
	FILE_SYNCHRONOUS_IO_ALERT                                          uint32 = 0x10
	FILE_SYNCHRONOUS_IO_NONALERT                                       uint32 = 0x20
	FILE_NON_DIRECTORY_FILE                                            uint32 = 0x40
	FILE_CREATE_TREE_CONNECTION                                        uint32 = 0x80
	FILE_COMPLETE_IF_OPLOCKED                                          uint32 = 0x100
	FILE_NO_EA_KNOWLEDGE                                               uint32 = 0x200
	FILE_OPEN_REMOTE_INSTANCE                                          uint32 = 0x400
	FILE_RANDOM_ACCESS                                                 uint32 = 0x800
	FILE_DELETE_ON_CLOSE                                               uint32 = 0x1000
	FILE_OPEN_BY_FILE_ID                                               uint32 = 0x2000
	FILE_OPEN_FOR_BACKUP_INTENT                                        uint32 = 0x4000
	FILE_NO_COMPRESSION                                                uint32 = 0x8000
	FILE_OPEN_REQUIRING_OPLOCK                                         uint32 = 0x10000
	FILE_RESERVE_OPFILTER                                              uint32 = 0x100000
	FILE_OPEN_REPARSE_POINT                                            uint32 = 0x200000
	FILE_OPEN_NO_RECALL                                                uint32 = 0x400000
	FILE_OPEN_FOR_FREE_SPACE_QUERY                                     uint32 = 0x800000
	FILE_VALID_OPTION_FLAGS                                            uint32 = 0xffffff
	FILE_VALID_PIPE_OPTION_FLAGS                                       uint32 = 0x32
	FILE_VALID_MAILSLOT_OPTION_FLAGS                                   uint32 = 0x32
	FILE_VALID_SET_FLAGS                                               uint32 = 0x36
	FILE_SUPERSEDED                                                    uint32 = 0x0
	FILE_OPENED                                                        uint32 = 0x1
	FILE_CREATED                                                       uint32 = 0x2
	FILE_OVERWRITTEN                                                   uint32 = 0x3
	FILE_EXISTS                                                        uint32 = 0x4
	FILE_DOES_NOT_EXIST                                                uint32 = 0x5
	WINWATCHNOTIFY_START                                               uint32 = 0x0
	WINWATCHNOTIFY_STOP                                                uint32 = 0x1
	WINWATCHNOTIFY_DESTROY                                             uint32 = 0x2
	WINWATCHNOTIFY_CHANGING                                            uint32 = 0x3
	WINWATCHNOTIFY_CHANGED                                             uint32 = 0x4
	RSC_FLAG_INF                                                       uint32 = 0x1
	RSC_FLAG_SKIPDISKSPACECHECK                                        uint32 = 0x2
	RSC_FLAG_QUIET                                                     uint32 = 0x4
	RSC_FLAG_NGCONV                                                    uint32 = 0x8
	RSC_FLAG_UPDHLPDLLS                                                uint32 = 0x10
	RSC_FLAG_DELAYREGISTEROCX                                          uint32 = 0x200
	RSC_FLAG_SETUPAPI                                                  uint32 = 0x400
	ALINF_QUIET                                                        uint32 = 0x4
	ALINF_NGCONV                                                       uint32 = 0x8
	ALINF_UPDHLPDLLS                                                   uint32 = 0x10
	ALINF_BKINSTALL                                                    uint32 = 0x20
	ALINF_ROLLBACK                                                     uint32 = 0x40
	ALINF_CHECKBKDATA                                                  uint32 = 0x80
	ALINF_ROLLBKDOALL                                                  uint32 = 0x100
	ALINF_DELAYREGISTEROCX                                             uint32 = 0x200
	AIF_WARNIFSKIP                                                     uint32 = 0x1
	AIF_NOSKIP                                                         uint32 = 0x2
	AIF_NOVERSIONCHECK                                                 uint32 = 0x4
	AIF_FORCE_FILE_IN_USE                                              uint32 = 0x8
	AIF_NOOVERWRITE                                                    uint32 = 0x10
	AIF_NO_VERSION_DIALOG                                              uint32 = 0x20
	AIF_REPLACEONLY                                                    uint32 = 0x400
	AIF_NOLANGUAGECHECK                                                uint32 = 0x10000000
	AIF_QUIET                                                          uint32 = 0x20000000
	IE4_RESTORE                                                        uint32 = 0x1
	IE4_BACKNEW                                                        uint32 = 0x2
	IE4_NODELETENEW                                                    uint32 = 0x4
	IE4_NOMESSAGES                                                     uint32 = 0x8
	IE4_NOPROGRESS                                                     uint32 = 0x10
	IE4_NOENUMKEY                                                      uint32 = 0x20
	IE4_NO_CRC_MAPPING                                                 uint32 = 0x40
	IE4_REGSECTION                                                     uint32 = 0x80
	IE4_FRDOALL                                                        uint32 = 0x100
	IE4_UPDREFCNT                                                      uint32 = 0x200
	IE4_USEREFCNT                                                      uint32 = 0x400
	IE4_EXTRAINCREFCNT                                                 uint32 = 0x800
	IE4_REMOVREGBKDATA                                                 uint32 = 0x1000
	ARSR_RESTORE                                                       uint32 = 0x1
	ARSR_NOMESSAGES                                                    uint32 = 0x8
	ARSR_REGSECTION                                                    uint32 = 0x80
	ARSR_REMOVREGBKDATA                                                uint32 = 0x1000
	REG_SAVE_LOG_KEY                                                   string = "RegSaveLogFile"
	REG_RESTORE_LOG_KEY                                                string = "RegRestoreLogFile"
	AFSR_RESTORE                                                       uint32 = 0x1
	AFSR_BACKNEW                                                       uint32 = 0x2
	AFSR_NODELETENEW                                                   uint32 = 0x4
	AFSR_NOMESSAGES                                                    uint32 = 0x8
	AFSR_NOPROGRESS                                                    uint32 = 0x10
	AFSR_UPDREFCNT                                                     uint32 = 0x200
	AFSR_USEREFCNT                                                     uint32 = 0x400
	AFSR_EXTRAINCREFCNT                                                uint32 = 0x800
	AADBE_ADD_ENTRY                                                    uint32 = 0x1
	AADBE_DEL_ENTRY                                                    uint32 = 0x2
	ADN_DEL_IF_EMPTY                                                   uint32 = 0x1
	ADN_DONT_DEL_SUBDIRS                                               uint32 = 0x2
	ADN_DONT_DEL_DIR                                                   uint32 = 0x4
	ADN_DEL_UNC_PATHS                                                  uint32 = 0x8
	LIS_QUIET                                                          uint32 = 0x1
	LIS_NOGRPCONV                                                      uint32 = 0x2
	RUNCMDS_QUIET                                                      uint32 = 0x1
	RUNCMDS_NOWAIT                                                     uint32 = 0x2
	RUNCMDS_DELAYPOSTCMD                                               uint32 = 0x4
	IME_MAXPROCESS                                                     uint32 = 0x20
	CP_HWND                                                            uint32 = 0x0
	CP_OPEN                                                            uint32 = 0x1
	CP_DIRECT                                                          uint32 = 0x2
	CP_LEVEL                                                           uint32 = 0x3
	MCW_DEFAULT                                                        uint32 = 0x0
	MCW_RECT                                                           uint32 = 0x1
	MCW_WINDOW                                                         uint32 = 0x2
	MCW_SCREEN                                                         uint32 = 0x4
	MCW_VERTICAL                                                       uint32 = 0x8
	MCW_HIDDEN                                                         uint32 = 0x10
	IME_MODE_ALPHANUMERIC                                              uint32 = 0x1
	IME_MODE_SBCSCHAR                                                  uint32 = 0x2
	IME_MODE_KATAKANA                                                  uint32 = 0x2
	IME_MODE_HIRAGANA                                                  uint32 = 0x4
	IME_MODE_HANJACONVERT                                              uint32 = 0x4
	IME_MODE_DBCSCHAR                                                  uint32 = 0x10
	IME_MODE_ROMAN                                                     uint32 = 0x20
	IME_MODE_NOROMAN                                                   uint32 = 0x40
	IME_MODE_CODEINPUT                                                 uint32 = 0x80
	IME_MODE_NOCODEINPUT                                               uint32 = 0x100
	IME_GETIMECAPS                                                     uint32 = 0x3
	IME_SETOPEN                                                        uint32 = 0x4
	IME_GETOPEN                                                        uint32 = 0x5
	IME_GETVERSION                                                     uint32 = 0x7
	IME_SETCONVERSIONWINDOW                                            uint32 = 0x8
	IME_MOVEIMEWINDOW                                                  uint32 = 0x8
	IME_SETCONVERSIONMODE                                              uint32 = 0x10
	IME_GETCONVERSIONMODE                                              uint32 = 0x11
	IME_SET_MODE                                                       uint32 = 0x12
	IME_SENDVKEY                                                       uint32 = 0x13
	IME_ENTERWORDREGISTERMODE                                          uint32 = 0x18
	IME_SETCONVERSIONFONTEX                                            uint32 = 0x19
	IME_BANJAtoJUNJA                                                   uint32 = 0x13
	IME_JUNJAtoBANJA                                                   uint32 = 0x14
	IME_JOHABtoKS                                                      uint32 = 0x15
	IME_KStoJOHAB                                                      uint32 = 0x16
	IMEA_INIT                                                          uint32 = 0x1
	IMEA_NEXT                                                          uint32 = 0x2
	IMEA_PREV                                                          uint32 = 0x3
	IME_REQUEST_CONVERT                                                uint32 = 0x1
	IME_ENABLE_CONVERT                                                 uint32 = 0x2
	INTERIM_WINDOW                                                     uint32 = 0x0
	MODE_WINDOW                                                        uint32 = 0x1
	HANJA_WINDOW                                                       uint32 = 0x2
	IME_RS_ERROR                                                       uint32 = 0x1
	IME_RS_NOIME                                                       uint32 = 0x2
	IME_RS_TOOLONG                                                     uint32 = 0x5
	IME_RS_ILLEGAL                                                     uint32 = 0x6
	IME_RS_NOTFOUND                                                    uint32 = 0x7
	IME_RS_NOROOM                                                      uint32 = 0xa
	IME_RS_DISKERROR                                                   uint32 = 0xe
	IME_RS_INVALID                                                     uint32 = 0x11
	IME_RS_NEST                                                        uint32 = 0x12
	IME_RS_SYSTEMMODAL                                                 uint32 = 0x13
	WM_IME_REPORT                                                      uint32 = 0x280
	IR_STRINGSTART                                                     uint32 = 0x100
	IR_STRINGEND                                                       uint32 = 0x101
	IR_OPENCONVERT                                                     uint32 = 0x120
	IR_CHANGECONVERT                                                   uint32 = 0x121
	IR_CLOSECONVERT                                                    uint32 = 0x122
	IR_FULLCONVERT                                                     uint32 = 0x123
	IR_IMESELECT                                                       uint32 = 0x130
	IR_STRING                                                          uint32 = 0x140
	IR_DBCSCHAR                                                        uint32 = 0x160
	IR_UNDETERMINE                                                     uint32 = 0x170
	IR_STRINGEX                                                        uint32 = 0x180
	IR_MODEINFO                                                        uint32 = 0x190
	WM_WNT_CONVERTREQUESTEX                                            uint32 = 0x109
	WM_CONVERTREQUEST                                                  uint32 = 0x10a
	WM_CONVERTRESULT                                                   uint32 = 0x10b
	WM_INTERIM                                                         uint32 = 0x10c
	WM_IMEKEYDOWN                                                      uint32 = 0x290
	WM_IMEKEYUP                                                        uint32 = 0x291
	DELAYLOAD_GPA_FAILURE                                              uint32 = 0x4
	DELETE_BROWSING_HISTORY_HISTORY                                    uint32 = 0x1
	DELETE_BROWSING_HISTORY_COOKIES                                    uint32 = 0x2
	DELETE_BROWSING_HISTORY_TIF                                        uint32 = 0x4
	DELETE_BROWSING_HISTORY_FORMDATA                                   uint32 = 0x8
	DELETE_BROWSING_HISTORY_PASSWORDS                                  uint32 = 0x10
	DELETE_BROWSING_HISTORY_PRESERVEFAVORITES                          uint32 = 0x20
	DELETE_BROWSING_HISTORY_DOWNLOADHISTORY                            uint32 = 0x40
)

var (
	CATID_DeleteBrowsingHistory = syscall.GUID{0x31CAF6E4, 0xD6AA, 0x4090,
		[8]byte{0xA0, 0x50, 0xA5, 0xAC, 0x89, 0x72, 0xE9, 0xEF}}
)

// enums

// enum
type TDIENTITY_ENTITY_TYPE uint32

const (
	GENERIC_ENTITY TDIENTITY_ENTITY_TYPE = 0
	AT_ENTITY      TDIENTITY_ENTITY_TYPE = 640
	CL_NL_ENTITY   TDIENTITY_ENTITY_TYPE = 769
	CO_NL_ENTITY   TDIENTITY_ENTITY_TYPE = 768
	CL_TL_ENTITY   TDIENTITY_ENTITY_TYPE = 1025
	CO_TL_ENTITY   TDIENTITY_ENTITY_TYPE = 1024
	ER_ENTITY      TDIENTITY_ENTITY_TYPE = 896
	IF_ENTITY      TDIENTITY_ENTITY_TYPE = 512
)

// enum
type FILE_INFORMATION_CLASS int32

const (
	FileDirectoryInformation FILE_INFORMATION_CLASS = 1
)

// enum
type SYSTEM_INFORMATION_CLASS int32

const (
	SystemBasicInformation                SYSTEM_INFORMATION_CLASS = 0
	SystemPerformanceInformation          SYSTEM_INFORMATION_CLASS = 2
	SystemTimeOfDayInformation            SYSTEM_INFORMATION_CLASS = 3
	SystemProcessInformation              SYSTEM_INFORMATION_CLASS = 5
	SystemProcessorPerformanceInformation SYSTEM_INFORMATION_CLASS = 8
	SystemInterruptInformation            SYSTEM_INFORMATION_CLASS = 23
	SystemExceptionInformation            SYSTEM_INFORMATION_CLASS = 33
	SystemRegistryQuotaInformation        SYSTEM_INFORMATION_CLASS = 37
	SystemLookasideInformation            SYSTEM_INFORMATION_CLASS = 45
	SystemCodeIntegrityInformation        SYSTEM_INFORMATION_CLASS = 103
	SystemPolicyInformation               SYSTEM_INFORMATION_CLASS = 134
)

// enum
type OBJECT_INFORMATION_CLASS int32

const (
	ObjectBasicInformation OBJECT_INFORMATION_CLASS = 0
	ObjectTypeInformation  OBJECT_INFORMATION_CLASS = 2
)

// enum
type KEY_SET_INFORMATION_CLASS int32

const (
	KeyWriteTimeInformation         KEY_SET_INFORMATION_CLASS = 0
	KeyWow64FlagsInformation        KEY_SET_INFORMATION_CLASS = 1
	KeyControlFlagsInformation      KEY_SET_INFORMATION_CLASS = 2
	KeySetVirtualizationInformation KEY_SET_INFORMATION_CLASS = 3
	KeySetDebugInformation          KEY_SET_INFORMATION_CLASS = 4
	KeySetHandleTagsInformation     KEY_SET_INFORMATION_CLASS = 5
	MaxKeySetInfoClass              KEY_SET_INFORMATION_CLASS = 6
)

// enum
type WINSTATIONINFOCLASS int32

const (
	WinStationInformation WINSTATIONINFOCLASS = 8
)

// enum
type CameraUIControlMode int32

const (
	Browse CameraUIControlMode = 0
	Linear CameraUIControlMode = 1
)

// enum
type CameraUIControlLinearSelectionMode int32

const (
	Single   CameraUIControlLinearSelectionMode = 0
	Multiple CameraUIControlLinearSelectionMode = 1
)

// enum
type CameraUIControlCaptureMode int32

const (
	PhotoOrVideo CameraUIControlCaptureMode = 0
	Photo        CameraUIControlCaptureMode = 1
	Video        CameraUIControlCaptureMode = 2
)

// enum
type CameraUIControlPhotoFormat int32

const (
	Jpeg   CameraUIControlPhotoFormat = 0
	Png    CameraUIControlPhotoFormat = 1
	JpegXR CameraUIControlPhotoFormat = 2
)

// enum
type CameraUIControlVideoFormat int32

const (
	Mp4 CameraUIControlVideoFormat = 0
	Wmv CameraUIControlVideoFormat = 1
)

// enum
type CameraUIControlViewType int32

const (
	SingleItem CameraUIControlViewType = 0
	ItemList   CameraUIControlViewType = 1
)

// enum
type FEATURE_CHANGE_TIME int32

const (
	FEATURE_CHANGE_TIME_READ          FEATURE_CHANGE_TIME = 0
	FEATURE_CHANGE_TIME_MODULE_RELOAD FEATURE_CHANGE_TIME = 1
	FEATURE_CHANGE_TIME_SESSION       FEATURE_CHANGE_TIME = 2
	FEATURE_CHANGE_TIME_REBOOT        FEATURE_CHANGE_TIME = 3
)

// enum
type FEATURE_ENABLED_STATE int32

const (
	FEATURE_ENABLED_STATE_DEFAULT  FEATURE_ENABLED_STATE = 0
	FEATURE_ENABLED_STATE_DISABLED FEATURE_ENABLED_STATE = 1
	FEATURE_ENABLED_STATE_ENABLED  FEATURE_ENABLED_STATE = 2
)

// enum
type TDI_TL_IO_CONTROL_TYPE int32

const (
	EndpointIoControlType   TDI_TL_IO_CONTROL_TYPE = 0
	SetSockOptIoControlType TDI_TL_IO_CONTROL_TYPE = 1
	GetSockOptIoControlType TDI_TL_IO_CONTROL_TYPE = 2
	SocketIoControlType     TDI_TL_IO_CONTROL_TYPE = 3
)

// enum
type WLDP_HOST int32

const (
	WLDP_HOST_RUNDLL32 WLDP_HOST = 0
	WLDP_HOST_SVCHOST  WLDP_HOST = 1
	WLDP_HOST_MAX      WLDP_HOST = 2
)

// enum
type WLDP_HOST_ID int32

const (
	WLDP_HOST_ID_UNKNOWN    WLDP_HOST_ID = 0
	WLDP_HOST_ID_GLOBAL     WLDP_HOST_ID = 1
	WLDP_HOST_ID_VBA        WLDP_HOST_ID = 2
	WLDP_HOST_ID_WSH        WLDP_HOST_ID = 3
	WLDP_HOST_ID_POWERSHELL WLDP_HOST_ID = 4
	WLDP_HOST_ID_IE         WLDP_HOST_ID = 5
	WLDP_HOST_ID_MSI        WLDP_HOST_ID = 6
	WLDP_HOST_ID_ALL        WLDP_HOST_ID = 7
	WLDP_HOST_ID_MAX        WLDP_HOST_ID = 8
)

// enum
type DECISION_LOCATION int32

const (
	DECISION_LOCATION_REFRESH_GLOBAL_DATA         DECISION_LOCATION = 0
	DECISION_LOCATION_PARAMETER_VALIDATION        DECISION_LOCATION = 1
	DECISION_LOCATION_AUDIT                       DECISION_LOCATION = 2
	DECISION_LOCATION_FAILED_CONVERT_GUID         DECISION_LOCATION = 3
	DECISION_LOCATION_ENTERPRISE_DEFINED_CLASS_ID DECISION_LOCATION = 4
	DECISION_LOCATION_GLOBAL_BUILT_IN_LIST        DECISION_LOCATION = 5
	DECISION_LOCATION_PROVIDER_BUILT_IN_LIST      DECISION_LOCATION = 6
	DECISION_LOCATION_ENFORCE_STATE_LIST          DECISION_LOCATION = 7
	DECISION_LOCATION_NOT_FOUND                   DECISION_LOCATION = 8
	DECISION_LOCATION_UNKNOWN                     DECISION_LOCATION = 9
)

// enum
type WLDP_KEY int32

const (
	KEY_UNKNOWN  WLDP_KEY = 0
	KEY_OVERRIDE WLDP_KEY = 1
	KEY_ALL_KEYS WLDP_KEY = 2
)

// enum
type VALUENAME int32

const (
	VALUENAME_UNKNOWN                     VALUENAME = 0
	VALUENAME_ENTERPRISE_DEFINED_CLASS_ID VALUENAME = 1
	VALUENAME_BUILT_IN_LIST               VALUENAME = 2
)

// enum
type WLDP_WINDOWS_LOCKDOWN_MODE int32

const (
	WLDP_WINDOWS_LOCKDOWN_MODE_UNLOCKED WLDP_WINDOWS_LOCKDOWN_MODE = 0
	WLDP_WINDOWS_LOCKDOWN_MODE_TRIAL    WLDP_WINDOWS_LOCKDOWN_MODE = 1
	WLDP_WINDOWS_LOCKDOWN_MODE_LOCKED   WLDP_WINDOWS_LOCKDOWN_MODE = 2
	WLDP_WINDOWS_LOCKDOWN_MODE_MAX      WLDP_WINDOWS_LOCKDOWN_MODE = 3
)

// enum
type WLDP_WINDOWS_LOCKDOWN_RESTRICTION int32

const (
	WLDP_WINDOWS_LOCKDOWN_RESTRICTION_NONE               WLDP_WINDOWS_LOCKDOWN_RESTRICTION = 0
	WLDP_WINDOWS_LOCKDOWN_RESTRICTION_NOUNLOCK           WLDP_WINDOWS_LOCKDOWN_RESTRICTION = 1
	WLDP_WINDOWS_LOCKDOWN_RESTRICTION_NOUNLOCK_PERMANENT WLDP_WINDOWS_LOCKDOWN_RESTRICTION = 2
	WLDP_WINDOWS_LOCKDOWN_RESTRICTION_MAX                WLDP_WINDOWS_LOCKDOWN_RESTRICTION = 3
)

// enum
type WLDP_POLICY_SETTING int32

const (
	WLDP_POLICY_SETTING_AV_PERF_MODE WLDP_POLICY_SETTING = 1000
)

// structs

type D3DHAL_CALLBACKS_ struct {
}

type D3DHAL_GLOBALDRIVERDATA_ struct {
}

type IMAGE_THUNK_DATA64_U1 struct {
	Data [1]uint64
}

func (this *IMAGE_THUNK_DATA64_U1) ForwarderString() *uint64 {
	return (*uint64)(unsafe.Pointer(this))
}

func (this *IMAGE_THUNK_DATA64_U1) ForwarderStringVal() uint64 {
	return *(*uint64)(unsafe.Pointer(this))
}

func (this *IMAGE_THUNK_DATA64_U1) Function() *uint64 {
	return (*uint64)(unsafe.Pointer(this))
}

func (this *IMAGE_THUNK_DATA64_U1) FunctionVal() uint64 {
	return *(*uint64)(unsafe.Pointer(this))
}

func (this *IMAGE_THUNK_DATA64_U1) Ordinal() *uint64 {
	return (*uint64)(unsafe.Pointer(this))
}

func (this *IMAGE_THUNK_DATA64_U1) OrdinalVal() uint64 {
	return *(*uint64)(unsafe.Pointer(this))
}

func (this *IMAGE_THUNK_DATA64_U1) AddressOfData() *uint64 {
	return (*uint64)(unsafe.Pointer(this))
}

func (this *IMAGE_THUNK_DATA64_U1) AddressOfDataVal() uint64 {
	return *(*uint64)(unsafe.Pointer(this))
}

type IMAGE_THUNK_DATA64 struct {
	U1 IMAGE_THUNK_DATA64_U1
}

type IMAGE_THUNK_DATA32_U1 struct {
	Data [1]uint32
}

func (this *IMAGE_THUNK_DATA32_U1) ForwarderString() *uint32 {
	return (*uint32)(unsafe.Pointer(this))
}

func (this *IMAGE_THUNK_DATA32_U1) ForwarderStringVal() uint32 {
	return *(*uint32)(unsafe.Pointer(this))
}

func (this *IMAGE_THUNK_DATA32_U1) Function() *uint32 {
	return (*uint32)(unsafe.Pointer(this))
}

func (this *IMAGE_THUNK_DATA32_U1) FunctionVal() uint32 {
	return *(*uint32)(unsafe.Pointer(this))
}

func (this *IMAGE_THUNK_DATA32_U1) Ordinal() *uint32 {
	return (*uint32)(unsafe.Pointer(this))
}

func (this *IMAGE_THUNK_DATA32_U1) OrdinalVal() uint32 {
	return *(*uint32)(unsafe.Pointer(this))
}

func (this *IMAGE_THUNK_DATA32_U1) AddressOfData() *uint32 {
	return (*uint32)(unsafe.Pointer(this))
}

func (this *IMAGE_THUNK_DATA32_U1) AddressOfDataVal() uint32 {
	return *(*uint32)(unsafe.Pointer(this))
}

type IMAGE_THUNK_DATA32 struct {
	U1 IMAGE_THUNK_DATA32_U1
}

type IMAGE_DELAYLOAD_DESCRIPTOR_Attributes_Anonymous struct {
	Bitfield_ uint32
}

type IMAGE_DELAYLOAD_DESCRIPTOR_Attributes struct {
	IMAGE_DELAYLOAD_DESCRIPTOR_Attributes_Anonymous
}

func (this *IMAGE_DELAYLOAD_DESCRIPTOR_Attributes) AllAttributes() *uint32 {
	return (*uint32)(unsafe.Pointer(this))
}

func (this *IMAGE_DELAYLOAD_DESCRIPTOR_Attributes) AllAttributesVal() uint32 {
	return *(*uint32)(unsafe.Pointer(this))
}

func (this *IMAGE_DELAYLOAD_DESCRIPTOR_Attributes) Anonymous() *IMAGE_DELAYLOAD_DESCRIPTOR_Attributes_Anonymous {
	return (*IMAGE_DELAYLOAD_DESCRIPTOR_Attributes_Anonymous)(unsafe.Pointer(this))
}

func (this *IMAGE_DELAYLOAD_DESCRIPTOR_Attributes) AnonymousVal() IMAGE_DELAYLOAD_DESCRIPTOR_Attributes_Anonymous {
	return *(*IMAGE_DELAYLOAD_DESCRIPTOR_Attributes_Anonymous)(unsafe.Pointer(this))
}

type IMAGE_DELAYLOAD_DESCRIPTOR struct {
	Attributes                 IMAGE_DELAYLOAD_DESCRIPTOR_Attributes
	DllNameRVA                 uint32
	ModuleHandleRVA            uint32
	ImportAddressTableRVA      uint32
	ImportNameTableRVA         uint32
	BoundImportAddressTableRVA uint32
	UnloadInformationTableRVA  uint32
	TimeDateStamp              uint32
}

type CUSTOM_SYSTEM_EVENT_TRIGGER_CONFIG struct {
	Size      uint32
	TriggerId PWSTR
}

type JIT_DEBUG_INFO struct {
	DwSize                  uint32
	DwProcessorArchitecture uint32
	DwThreadID              uint32
	DwReserved0             uint32
	LpExceptionAddress      uint64
	LpExceptionRecord       uint64
	LpContextRecord         uint64
}

type HW_PROFILE_INFOA struct {
	DwDockInfo      uint32
	SzHwProfileGuid [39]CHAR
	SzHwProfileName [80]CHAR
}

type HW_PROFILE_INFO = HW_PROFILE_INFOW
type HW_PROFILE_INFOW struct {
	DwDockInfo      uint32
	SzHwProfileGuid [39]uint16
	SzHwProfileName [80]uint16
}

type ACTCTX_SECTION_KEYED_DATA_2600 struct {
	CbSize                    uint32
	UlDataFormatVersion       uint32
	LpData                    unsafe.Pointer
	UlLength                  uint32
	LpSectionGlobalData       unsafe.Pointer
	UlSectionGlobalDataLength uint32
	LpSectionBase             unsafe.Pointer
	UlSectionTotalLength      uint32
	HActCtx                   HANDLE
	UlAssemblyRosterIndex     uint32
}

type ACTCTX_SECTION_KEYED_DATA_ASSEMBLY_METADATA struct {
	LpInformation             unsafe.Pointer
	LpSectionBase             unsafe.Pointer
	UlSectionLength           uint32
	LpSectionGlobalDataBase   unsafe.Pointer
	UlSectionGlobalDataLength uint32
}

type ACTIVATION_CONTEXT_BASIC_INFORMATION struct {
	HActCtx HANDLE
	DwFlags uint32
}

type FILE_CASE_SENSITIVE_INFO struct {
	Flags uint32
}

type FILE_DISPOSITION_INFO_EX struct {
	Flags uint32
}

type CLIENT_ID struct {
	UniqueProcess HANDLE
	UniqueThread  HANDLE
}

type LDR_DATA_TABLE_ENTRY_Anonymous struct {
	Data [1]uint64
}

func (this *LDR_DATA_TABLE_ENTRY_Anonymous) CheckSum() *uint32 {
	return (*uint32)(unsafe.Pointer(this))
}

func (this *LDR_DATA_TABLE_ENTRY_Anonymous) CheckSumVal() uint32 {
	return *(*uint32)(unsafe.Pointer(this))
}

func (this *LDR_DATA_TABLE_ENTRY_Anonymous) Reserved6() *unsafe.Pointer {
	return (*unsafe.Pointer)(unsafe.Pointer(this))
}

func (this *LDR_DATA_TABLE_ENTRY_Anonymous) Reserved6Val() unsafe.Pointer {
	return *(*unsafe.Pointer)(unsafe.Pointer(this))
}

type LDR_DATA_TABLE_ENTRY struct {
	Reserved1          [2]unsafe.Pointer
	InMemoryOrderLinks LIST_ENTRY
	Reserved2          [2]unsafe.Pointer
	DllBase            unsafe.Pointer
	Reserved3          [2]unsafe.Pointer
	FullDllName        UNICODE_STRING
	Reserved4          [8]byte
	Reserved5          [3]unsafe.Pointer
	LDR_DATA_TABLE_ENTRY_Anonymous
	TimeDateStamp uint32
}

type OBJECT_ATTRIBUTES struct {
	Length                   uint32
	RootDirectory            HANDLE
	ObjectName               *UNICODE_STRING
	Attributes               uint32
	SecurityDescriptor       unsafe.Pointer
	SecurityQualityOfService unsafe.Pointer
}

type IO_STATUS_BLOCK_Anonymous struct {
	Data [1]uint64
}

func (this *IO_STATUS_BLOCK_Anonymous) Status() *NTSTATUS {
	return (*NTSTATUS)(unsafe.Pointer(this))
}

func (this *IO_STATUS_BLOCK_Anonymous) StatusVal() NTSTATUS {
	return *(*NTSTATUS)(unsafe.Pointer(this))
}

func (this *IO_STATUS_BLOCK_Anonymous) Pointer() *unsafe.Pointer {
	return (*unsafe.Pointer)(unsafe.Pointer(this))
}

func (this *IO_STATUS_BLOCK_Anonymous) PointerVal() unsafe.Pointer {
	return *(*unsafe.Pointer)(unsafe.Pointer(this))
}

type IO_STATUS_BLOCK struct {
	IO_STATUS_BLOCK_Anonymous
	Information uintptr
}

type SYSTEM_PROCESSOR_PERFORMANCE_INFORMATION struct {
	IdleTime   int64
	KernelTime int64
	UserTime   int64
	Reserved1  [2]int64
	Reserved2  uint32
}

type SYSTEM_PROCESS_INFORMATION struct {
	NextEntryOffset        uint32
	NumberOfThreads        uint32
	Reserved1              [48]byte
	ImageName              UNICODE_STRING
	BasePriority           int32
	UniqueProcessId        HANDLE
	Reserved2              unsafe.Pointer
	HandleCount            uint32
	SessionId              uint32
	Reserved3              unsafe.Pointer
	PeakVirtualSize        uintptr
	VirtualSize            uintptr
	Reserved4              uint32
	PeakWorkingSetSize     uintptr
	WorkingSetSize         uintptr
	Reserved5              unsafe.Pointer
	QuotaPagedPoolUsage    uintptr
	Reserved6              unsafe.Pointer
	QuotaNonPagedPoolUsage uintptr
	PagefileUsage          uintptr
	PeakPagefileUsage      uintptr
	PrivatePageCount       uintptr
	Reserved7              [6]int64
}

type SYSTEM_THREAD_INFORMATION struct {
	Reserved1    [3]int64
	Reserved2    uint32
	StartAddress unsafe.Pointer
	ClientId     CLIENT_ID
	Priority     int32
	BasePriority int32
	Reserved3    uint32
	ThreadState  uint32
	WaitReason   uint32
}

type SYSTEM_REGISTRY_QUOTA_INFORMATION struct {
	RegistryQuotaAllowed uint32
	RegistryQuotaUsed    uint32
	Reserved1            unsafe.Pointer
}

type SYSTEM_BASIC_INFORMATION struct {
	Reserved1          [24]byte
	Reserved2          [4]unsafe.Pointer
	NumberOfProcessors int8
}

type SYSTEM_TIMEOFDAY_INFORMATION struct {
	Reserved1 [48]byte
}

type SYSTEM_PERFORMANCE_INFORMATION struct {
	Reserved1 [312]byte
}

type SYSTEM_EXCEPTION_INFORMATION struct {
	Reserved1 [16]byte
}

type SYSTEM_LOOKASIDE_INFORMATION struct {
	Reserved1 [32]byte
}

type SYSTEM_INTERRUPT_INFORMATION struct {
	Reserved1 [24]byte
}

type SYSTEM_POLICY_INFORMATION struct {
	Reserved1 [2]unsafe.Pointer
	Reserved2 [3]uint32
}

type THREAD_NAME_INFORMATION struct {
	ThreadName UNICODE_STRING
}

type SYSTEM_CODEINTEGRITY_INFORMATION struct {
	Length               uint32
	CodeIntegrityOptions uint32
}

type PUBLIC_OBJECT_BASIC_INFORMATION struct {
	Attributes    uint32
	GrantedAccess uint32
	HandleCount   uint32
	PointerCount  uint32
	Reserved      [10]uint32
}

type PUBLIC_OBJECT_TYPE_INFORMATION struct {
	TypeName UNICODE_STRING
	Reserved [22]uint32
}

type KEY_VALUE_ENTRY struct {
	ValueName  *UNICODE_STRING
	DataLength uint32
	DataOffset uint32
	Type       uint32
}

type WINSTATIONINFORMATIONW struct {
	Reserved2 [70]byte
	LogonId   uint32
	Reserved3 [1140]byte
}

type CameraUIControl struct {
}

type EditionUpgradeHelper struct {
}

type EditionUpgradeBroker struct {
}

type FEATURE_ERROR struct {
	Hr                              HRESULT
	LineNumber                      uint16
	File                            PSTR
	Process                         PSTR
	Module                          PSTR
	CallerReturnAddressOffset       uint32
	CallerModule                    PSTR
	Message                         PSTR
	OriginLineNumber                uint16
	OriginFile                      PSTR
	OriginModule                    PSTR
	OriginCallerReturnAddressOffset uint32
	OriginCallerModule              PSTR
	OriginName                      PSTR
}

type DCICMD struct {
	DwCommand  uint32
	DwParam1   uint32
	DwParam2   uint32
	DwVersion  uint32
	DwReserved uint32
}

type DCICREATEINPUT struct {
	Cmd           DCICMD
	DwCompression uint32
	DwMask        [3]uint32
	DwWidth       uint32
	DwHeight      uint32
	DwDCICaps     uint32
	DwBitCount    uint32
	LpSurface     unsafe.Pointer
}

type DCISURFACEINFO struct {
	DwSize         uint32
	DwDCICaps      uint32
	DwCompression  uint32
	DwMask         [3]uint32
	DwWidth        uint32
	DwHeight       uint32
	LStride        int32
	DwBitCount     uint32
	DwOffSurface   uintptr
	WSelSurface    uint16
	WReserved      uint16
	DwReserved1    uint32
	DwReserved2    uint32
	DwReserved3    uint32
	BeginAccess    uintptr
	EndAccess      uintptr
	DestroySurface uintptr
}

type DCIENUMINPUT struct {
	Cmd          DCICMD
	RSrc         RECT
	RDst         RECT
	EnumCallback uintptr
	LpContext    unsafe.Pointer
}

type DCIOFFSCREEN struct {
	DciInfo        DCISURFACEINFO
	Draw           uintptr
	SetClipList    uintptr
	SetDestination uintptr
}

type DCIOVERLAY struct {
	DciInfo          DCISURFACEINFO
	DwChromakeyValue uint32
	DwChromakeyMask  uint32
}

type STRENTRYA struct {
	PszName  PSTR
	PszValue PSTR
}

type STRENTRY = STRENTRYW
type STRENTRYW struct {
	PszName  PWSTR
	PszValue PWSTR
}

type STRTABLEA struct {
	CEntries uint32
	Pse      *STRENTRYA
}

type STRTABLE = STRTABLEW
type STRTABLEW struct {
	CEntries uint32
	Pse      *STRENTRYW
}

type CABINFOA struct {
	PszCab     PSTR
	PszInf     PSTR
	PszSection PSTR
	SzSrcPath  [260]CHAR
	DwFlags    uint32
}

type CABINFO = CABINFOW
type CABINFOW struct {
	PszCab     PWSTR
	PszInf     PWSTR
	PszSection PWSTR
	SzSrcPath  [260]uint16
	DwFlags    uint32
}

type PERUSERSECTIONA struct {
	SzGUID        [59]CHAR
	SzDispName    [128]CHAR
	SzLocale      [10]CHAR
	SzStub        [1040]CHAR
	SzVersion     [32]CHAR
	SzCompID      [128]CHAR
	DwIsInstalled uint32
	BRollback     BOOL
}

type PERUSERSECTION = PERUSERSECTIONW
type PERUSERSECTIONW struct {
	SzGUID        [59]uint16
	SzDispName    [128]uint16
	SzLocale      [10]uint16
	SzStub        [1040]uint16
	SzVersion     [32]uint16
	SzCompID      [128]uint16
	DwIsInstalled uint32
	BRollback     BOOL
}

type IMESTRUCT struct {
	Fnc       uint32
	WParam    WPARAM
	WCount    uint32
	DchSource uint32
	DchDest   uint32
	LParam1   LPARAM
	LParam2   LPARAM
	LParam3   LPARAM
}

type UNDETERMINESTRUCT struct {
	DwSize             uint32
	UDefIMESize        uint32
	UDefIMEPos         uint32
	UUndetTextLen      uint32
	UUndetTextPos      uint32
	UUndetAttrPos      uint32
	UCursorPos         uint32
	UDeltaStart        uint32
	UDetermineTextLen  uint32
	UDetermineTextPos  uint32
	UDetermineDelimPos uint32
	UYomiTextLen       uint32
	UYomiTextPos       uint32
	UYomiDelimPos      uint32
}

type STRINGEXSTRUCT struct {
	DwSize             uint32
	UDeterminePos      uint32
	UDetermineDelimPos uint32
	UYomiPos           uint32
	UYomiDelimPos      uint32
}

type DATETIME struct {
	Year  uint16
	Month uint16
	Day   uint16
	Hour  uint16
	Min   uint16
	Sec   uint16
}

type IMEPROA struct {
	HWnd          HWND
	InstDate      DATETIME
	WVersion      uint32
	SzDescription [50]byte
	SzName        [80]byte
	SzOptions     [30]byte
}

type IMEPRO = IMEPROW
type IMEPROW struct {
	HWnd          HWND
	InstDate      DATETIME
	WVersion      uint32
	SzDescription [50]uint16
	SzName        [80]uint16
	SzOptions     [30]uint16
}

type JAVA_TRUST struct {
	CbSize                 uint32
	Flag                   uint32
	FAllActiveXPermissions BOOL
	FAllPermissions        BOOL
	DwEncodingType         uint32
	PbJavaPermissions      *byte
	CbJavaPermissions      uint32
	PbSigner               *byte
	CbSigner               uint32
	PwszZone               PWSTR
	GuidZone               syscall.GUID
	HVerify                HRESULT
}

type TDIEntityID struct {
	Tei_entity   TDIENTITY_ENTITY_TYPE
	Tei_instance uint32
}

type TDIObjectID struct {
	Toi_entity TDIEntityID
	Toi_class  uint32
	Toi_type   uint32
	Toi_id     uint32
}

type Tcp_request_query_information_ex_xp struct {
	ID      TDIObjectID
	Context [2]uintptr
}

type Tcp_request_query_information_ex32_xp struct {
	ID      TDIObjectID
	Context [4]uint32
}

type Tcp_request_query_information_ex_w2k struct {
	ID      TDIObjectID
	Context [16]byte
}

type Tcp_request_set_information_ex struct {
	ID         TDIObjectID
	BufferSize uint32
	Buffer     [1]byte
}

type TDI_TL_IO_CONTROL_ENDPOINT_Anonymous struct {
	Data [1]uint32
}

func (this *TDI_TL_IO_CONTROL_ENDPOINT_Anonymous) IoControlCode() *uint32 {
	return (*uint32)(unsafe.Pointer(this))
}

func (this *TDI_TL_IO_CONTROL_ENDPOINT_Anonymous) IoControlCodeVal() uint32 {
	return *(*uint32)(unsafe.Pointer(this))
}

func (this *TDI_TL_IO_CONTROL_ENDPOINT_Anonymous) OptionName() *uint32 {
	return (*uint32)(unsafe.Pointer(this))
}

func (this *TDI_TL_IO_CONTROL_ENDPOINT_Anonymous) OptionNameVal() uint32 {
	return *(*uint32)(unsafe.Pointer(this))
}

type TDI_TL_IO_CONTROL_ENDPOINT struct {
	Type  TDI_TL_IO_CONTROL_TYPE
	Level uint32
	TDI_TL_IO_CONTROL_ENDPOINT_Anonymous
	InputBuffer        unsafe.Pointer
	InputBufferLength  uint32
	OutputBuffer       unsafe.Pointer
	OutputBufferLength uint32
}

type WLDP_HOST_INFORMATION struct {
	DwRevision uint32
	DwHostId   WLDP_HOST_ID
	SzSource   PWSTR
	HSource    HANDLE
}

type WLDP_DEVICE_SECURITY_INFORMATION struct {
	UnlockIdSize         uint32
	UnlockId             *byte
	ManufacturerIDLength uint32
	ManufacturerID       PWSTR
}

type DefaultBrowserSyncSettings struct {
}

type DELAYLOAD_PROC_DESCRIPTOR_Description struct {
	Data [1]uint64
}

func (this *DELAYLOAD_PROC_DESCRIPTOR_Description) Name() *PSTR {
	return (*PSTR)(unsafe.Pointer(this))
}

func (this *DELAYLOAD_PROC_DESCRIPTOR_Description) NameVal() PSTR {
	return *(*PSTR)(unsafe.Pointer(this))
}

func (this *DELAYLOAD_PROC_DESCRIPTOR_Description) Ordinal() *uint32 {
	return (*uint32)(unsafe.Pointer(this))
}

func (this *DELAYLOAD_PROC_DESCRIPTOR_Description) OrdinalVal() uint32 {
	return *(*uint32)(unsafe.Pointer(this))
}

type DELAYLOAD_PROC_DESCRIPTOR struct {
	ImportDescribedByName uint32
	Description           DELAYLOAD_PROC_DESCRIPTOR_Description
}

type DELAYLOAD_INFO struct {
	Size                uint32
	DelayloadDescriptor *IMAGE_DELAYLOAD_DESCRIPTOR
	ThunkAddress        *IMAGE_THUNK_DATA64
	TargetDllName       PSTR
	TargetApiDescriptor DELAYLOAD_PROC_DESCRIPTOR
	TargetModuleBase    unsafe.Pointer
	Unused              unsafe.Pointer
	LastError           uint32
}

// func types

type PFIBER_CALLOUT_ROUTINE = uintptr
type PFIBER_CALLOUT_ROUTINE_func = func(lpParameter unsafe.Pointer) unsafe.Pointer

type PQUERYACTCTXW_FUNC = uintptr
type PQUERYACTCTXW_FUNC_func = func(dwFlags uint32, hActCtx HANDLE, pvSubInstance unsafe.Pointer, ulInfoClass uint32, pvBuffer unsafe.Pointer, cbBuffer uintptr, pcbWrittenOrRequired *uintptr) BOOL

type APPLICATION_RECOVERY_CALLBACK = uintptr
type APPLICATION_RECOVERY_CALLBACK_func = func(pvParameter unsafe.Pointer) uint32

type PIO_APC_ROUTINE = uintptr
type PIO_APC_ROUTINE_func = func(ApcContext unsafe.Pointer, IoStatusBlock *IO_STATUS_BLOCK, Reserved uint32)

type PWINSTATIONQUERYINFORMATIONW = uintptr
type PWINSTATIONQUERYINFORMATIONW_func = func(param0 HANDLE, param1 uint32, param2 WINSTATIONINFOCLASS, param3 unsafe.Pointer, param4 uint32, param5 *uint32) BOOLEAN

type PFEATURE_STATE_CHANGE_CALLBACK = uintptr
type PFEATURE_STATE_CHANGE_CALLBACK_func = func(context unsafe.Pointer)

type ENUM_CALLBACK = uintptr
type ENUM_CALLBACK_func = func(lpSurfaceInfo *DCISURFACEINFO, lpContext unsafe.Pointer)

type WINWATCHNOTIFYPROC = uintptr
type WINWATCHNOTIFYPROC_func = func(hww HWINWATCH, hwnd HWND, code uint32, lParam LPARAM)

type REGINSTALLA = uintptr
type REGINSTALLA_func = func(hm HINSTANCE, pszSection PSTR, pstTable *STRTABLEA) HRESULT

type PWLDP_SETDYNAMICCODETRUST_API = uintptr
type PWLDP_SETDYNAMICCODETRUST_API_func = func(hFileHandle HANDLE) HRESULT

type PWLDP_ISDYNAMICCODEPOLICYENABLED_API = uintptr
type PWLDP_ISDYNAMICCODEPOLICYENABLED_API_func = func(pbEnabled *BOOL) HRESULT

type PWLDP_QUERYDYNAMICODETRUST_API = uintptr
type PWLDP_QUERYDYNAMICODETRUST_API_func = func(fileHandle HANDLE, baseImage unsafe.Pointer, imageSize uint32) HRESULT

type PWLDP_QUERYWINDOWSLOCKDOWNMODE_API = uintptr
type PWLDP_QUERYWINDOWSLOCKDOWNMODE_API_func = func(lockdownMode *WLDP_WINDOWS_LOCKDOWN_MODE) HRESULT

type PWLDP_QUERYDEVICESECURITYINFORMATION_API = uintptr
type PWLDP_QUERYDEVICESECURITYINFORMATION_API_func = func(information *WLDP_DEVICE_SECURITY_INFORMATION, informationLength uint32, returnLength *uint32) HRESULT

type PWLDP_QUERYWINDOWSLOCKDOWNRESTRICTION_API = uintptr
type PWLDP_QUERYWINDOWSLOCKDOWNRESTRICTION_API_func = func(LockdownRestriction *WLDP_WINDOWS_LOCKDOWN_RESTRICTION) HRESULT

type PWLDP_SETWINDOWSLOCKDOWNRESTRICTION_API = uintptr
type PWLDP_SETWINDOWSLOCKDOWNRESTRICTION_API_func = func(LockdownRestriction WLDP_WINDOWS_LOCKDOWN_RESTRICTION) HRESULT

type PWLDP_ISAPPAPPROVEDBYPOLICY_API = uintptr
type PWLDP_ISAPPAPPROVEDBYPOLICY_API_func = func(PackageFamilyName PWSTR, PackageVersion uint64) HRESULT

type PWLDP_QUERYPOLICYSETTINGENABLED_API = uintptr
type PWLDP_QUERYPOLICYSETTINGENABLED_API_func = func(Setting WLDP_POLICY_SETTING, Enabled *BOOL) HRESULT

type PWLDP_QUERYPOLICYSETTINGENABLED2_API = uintptr
type PWLDP_QUERYPOLICYSETTINGENABLED2_API_func = func(Setting PWSTR, Enabled *BOOL) HRESULT

type PWLDP_ISWCOSPRODUCTIONCONFIGURATION_API = uintptr
type PWLDP_ISWCOSPRODUCTIONCONFIGURATION_API_func = func(IsProductionConfiguration *BOOL) HRESULT

type PWLDP_RESETWCOSPRODUCTIONCONFIGURATION_API = uintptr
type PWLDP_RESETWCOSPRODUCTIONCONFIGURATION_API_func = func() HRESULT

type PWLDP_ISPRODUCTIONCONFIGURATION_API = uintptr
type PWLDP_ISPRODUCTIONCONFIGURATION_API_func = func(IsProductionConfiguration *BOOL) HRESULT

type PWLDP_RESETPRODUCTIONCONFIGURATION_API = uintptr
type PWLDP_RESETPRODUCTIONCONFIGURATION_API_func = func() HRESULT

type PDELAYLOAD_FAILURE_DLL_CALLBACK = uintptr
type PDELAYLOAD_FAILURE_DLL_CALLBACK_func = func(NotificationReason uint32, DelayloadInfo *DELAYLOAD_INFO) unsafe.Pointer

// interfaces

// 1BFA0C2C-FBCD-4776-BDA4-88BF974E74F4
var IID_ICameraUIControlEventCallback = syscall.GUID{0x1BFA0C2C, 0xFBCD, 0x4776,
	[8]byte{0xBD, 0xA4, 0x88, 0xBF, 0x97, 0x4E, 0x74, 0xF4}}

type ICameraUIControlEventCallbackInterface interface {
	IUnknownInterface
	OnStartupComplete()
	OnSuspendComplete()
	OnItemCaptured(pszPath PWSTR)
	OnItemDeleted(pszPath PWSTR)
	OnClosed()
}

type ICameraUIControlEventCallbackVtbl struct {
	IUnknownVtbl
	OnStartupComplete uintptr
	OnSuspendComplete uintptr
	OnItemCaptured    uintptr
	OnItemDeleted     uintptr
	OnClosed          uintptr
}

type ICameraUIControlEventCallback struct {
	IUnknown
}

func (this *ICameraUIControlEventCallback) Vtbl() *ICameraUIControlEventCallbackVtbl {
	return (*ICameraUIControlEventCallbackVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *ICameraUIControlEventCallback) OnStartupComplete() {
	_, _, _ = syscall.SyscallN(this.Vtbl().OnStartupComplete, uintptr(unsafe.Pointer(this)))
}

func (this *ICameraUIControlEventCallback) OnSuspendComplete() {
	_, _, _ = syscall.SyscallN(this.Vtbl().OnSuspendComplete, uintptr(unsafe.Pointer(this)))
}

func (this *ICameraUIControlEventCallback) OnItemCaptured(pszPath PWSTR) {
	_, _, _ = syscall.SyscallN(this.Vtbl().OnItemCaptured, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pszPath)))
}

func (this *ICameraUIControlEventCallback) OnItemDeleted(pszPath PWSTR) {
	_, _, _ = syscall.SyscallN(this.Vtbl().OnItemDeleted, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pszPath)))
}

func (this *ICameraUIControlEventCallback) OnClosed() {
	_, _, _ = syscall.SyscallN(this.Vtbl().OnClosed, uintptr(unsafe.Pointer(this)))
}

// B8733ADF-3D68-4B8F-BB08-E28A0BED0376
var IID_ICameraUIControl = syscall.GUID{0xB8733ADF, 0x3D68, 0x4B8F,
	[8]byte{0xBB, 0x08, 0xE2, 0x8A, 0x0B, 0xED, 0x03, 0x76}}

type ICameraUIControlInterface interface {
	IUnknownInterface
	Show(pWindow *IUnknown, mode CameraUIControlMode, selectionMode CameraUIControlLinearSelectionMode, captureMode CameraUIControlCaptureMode, photoFormat CameraUIControlPhotoFormat, videoFormat CameraUIControlVideoFormat, bHasCloseButton BOOL, pEventCallback *ICameraUIControlEventCallback) HRESULT
	Close() HRESULT
	Suspend(pbDeferralRequired *BOOL) HRESULT
	Resume() HRESULT
	GetCurrentViewType(pViewType *CameraUIControlViewType) HRESULT
	GetActiveItem(pbstrActiveItemPath *BSTR) HRESULT
	GetSelectedItems(ppSelectedItemPaths **SAFEARRAY) HRESULT
	RemoveCapturedItem(pszPath PWSTR) HRESULT
}

type ICameraUIControlVtbl struct {
	IUnknownVtbl
	Show               uintptr
	Close              uintptr
	Suspend            uintptr
	Resume             uintptr
	GetCurrentViewType uintptr
	GetActiveItem      uintptr
	GetSelectedItems   uintptr
	RemoveCapturedItem uintptr
}

type ICameraUIControl struct {
	IUnknown
}

func (this *ICameraUIControl) Vtbl() *ICameraUIControlVtbl {
	return (*ICameraUIControlVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *ICameraUIControl) Show(pWindow *IUnknown, mode CameraUIControlMode, selectionMode CameraUIControlLinearSelectionMode, captureMode CameraUIControlCaptureMode, photoFormat CameraUIControlPhotoFormat, videoFormat CameraUIControlVideoFormat, bHasCloseButton BOOL, pEventCallback *ICameraUIControlEventCallback) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Show, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pWindow)), uintptr(mode), uintptr(selectionMode), uintptr(captureMode), uintptr(photoFormat), uintptr(videoFormat), uintptr(bHasCloseButton), uintptr(unsafe.Pointer(pEventCallback)))
	return HRESULT(ret)
}

func (this *ICameraUIControl) Close() HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Close, uintptr(unsafe.Pointer(this)))
	return HRESULT(ret)
}

func (this *ICameraUIControl) Suspend(pbDeferralRequired *BOOL) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Suspend, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pbDeferralRequired)))
	return HRESULT(ret)
}

func (this *ICameraUIControl) Resume() HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Resume, uintptr(unsafe.Pointer(this)))
	return HRESULT(ret)
}

func (this *ICameraUIControl) GetCurrentViewType(pViewType *CameraUIControlViewType) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetCurrentViewType, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pViewType)))
	return HRESULT(ret)
}

func (this *ICameraUIControl) GetActiveItem(pbstrActiveItemPath *BSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetActiveItem, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pbstrActiveItemPath)))
	return HRESULT(ret)
}

func (this *ICameraUIControl) GetSelectedItems(ppSelectedItemPaths **SAFEARRAY) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetSelectedItems, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(ppSelectedItemPaths)))
	return HRESULT(ret)
}

func (this *ICameraUIControl) RemoveCapturedItem(pszPath PWSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().RemoveCapturedItem, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pszPath)))
	return HRESULT(ret)
}

// D3E9E342-5DEB-43B6-849E-6913B85D503A
var IID_IEditionUpgradeHelper = syscall.GUID{0xD3E9E342, 0x5DEB, 0x43B6,
	[8]byte{0x84, 0x9E, 0x69, 0x13, 0xB8, 0x5D, 0x50, 0x3A}}

type IEditionUpgradeHelperInterface interface {
	IUnknownInterface
	CanUpgrade(isAllowed *BOOL) HRESULT
	UpdateOperatingSystem(contentId PWSTR) HRESULT
	ShowProductKeyUI() HRESULT
	GetOsProductContentId(contentId *PWSTR) HRESULT
	GetGenuineLocalStatus(isGenuine *BOOL) HRESULT
}

type IEditionUpgradeHelperVtbl struct {
	IUnknownVtbl
	CanUpgrade            uintptr
	UpdateOperatingSystem uintptr
	ShowProductKeyUI      uintptr
	GetOsProductContentId uintptr
	GetGenuineLocalStatus uintptr
}

type IEditionUpgradeHelper struct {
	IUnknown
}

func (this *IEditionUpgradeHelper) Vtbl() *IEditionUpgradeHelperVtbl {
	return (*IEditionUpgradeHelperVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IEditionUpgradeHelper) CanUpgrade(isAllowed *BOOL) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().CanUpgrade, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(isAllowed)))
	return HRESULT(ret)
}

func (this *IEditionUpgradeHelper) UpdateOperatingSystem(contentId PWSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().UpdateOperatingSystem, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(contentId)))
	return HRESULT(ret)
}

func (this *IEditionUpgradeHelper) ShowProductKeyUI() HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().ShowProductKeyUI, uintptr(unsafe.Pointer(this)))
	return HRESULT(ret)
}

func (this *IEditionUpgradeHelper) GetOsProductContentId(contentId *PWSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetOsProductContentId, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(contentId)))
	return HRESULT(ret)
}

func (this *IEditionUpgradeHelper) GetGenuineLocalStatus(isGenuine *BOOL) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetGenuineLocalStatus, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(isGenuine)))
	return HRESULT(ret)
}

// F342D19E-CC22-4648-BB5D-03CCF75B47C5
var IID_IWindowsLockModeHelper = syscall.GUID{0xF342D19E, 0xCC22, 0x4648,
	[8]byte{0xBB, 0x5D, 0x03, 0xCC, 0xF7, 0x5B, 0x47, 0xC5}}

type IWindowsLockModeHelperInterface interface {
	IUnknownInterface
	GetSMode(isSmode *BOOL) HRESULT
}

type IWindowsLockModeHelperVtbl struct {
	IUnknownVtbl
	GetSMode uintptr
}

type IWindowsLockModeHelper struct {
	IUnknown
}

func (this *IWindowsLockModeHelper) Vtbl() *IWindowsLockModeHelperVtbl {
	return (*IWindowsLockModeHelperVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IWindowsLockModeHelper) GetSMode(isSmode *BOOL) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetSMode, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(isSmode)))
	return HRESULT(ret)
}

// FF19CBCF-9455-4937-B872-6B7929A460AF
var IID_IEditionUpgradeBroker = syscall.GUID{0xFF19CBCF, 0x9455, 0x4937,
	[8]byte{0xB8, 0x72, 0x6B, 0x79, 0x29, 0xA4, 0x60, 0xAF}}

type IEditionUpgradeBrokerInterface interface {
	IUnknownInterface
	InitializeParentWindow(parentHandle uint32) HRESULT
	UpdateOperatingSystem(parameter BSTR) HRESULT
	ShowProductKeyUI() HRESULT
	CanUpgrade() HRESULT
}

type IEditionUpgradeBrokerVtbl struct {
	IUnknownVtbl
	InitializeParentWindow uintptr
	UpdateOperatingSystem  uintptr
	ShowProductKeyUI       uintptr
	CanUpgrade             uintptr
}

type IEditionUpgradeBroker struct {
	IUnknown
}

func (this *IEditionUpgradeBroker) Vtbl() *IEditionUpgradeBrokerVtbl {
	return (*IEditionUpgradeBrokerVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IEditionUpgradeBroker) InitializeParentWindow(parentHandle uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().InitializeParentWindow, uintptr(unsafe.Pointer(this)), uintptr(parentHandle))
	return HRESULT(ret)
}

func (this *IEditionUpgradeBroker) UpdateOperatingSystem(parameter BSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().UpdateOperatingSystem, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(parameter)))
	return HRESULT(ret)
}

func (this *IEditionUpgradeBroker) ShowProductKeyUI() HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().ShowProductKeyUI, uintptr(unsafe.Pointer(this)))
	return HRESULT(ret)
}

func (this *IEditionUpgradeBroker) CanUpgrade() HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().CanUpgrade, uintptr(unsafe.Pointer(this)))
	return HRESULT(ret)
}

// B524F93F-80D5-4EC7-AE9E-D66E93ADE1FA
var IID_IContainerActivationHelper = syscall.GUID{0xB524F93F, 0x80D5, 0x4EC7,
	[8]byte{0xAE, 0x9E, 0xD6, 0x6E, 0x93, 0xAD, 0xE1, 0xFA}}

type IContainerActivationHelperInterface interface {
	IUnknownInterface
	CanActivateClientVM(isAllowed *int16) HRESULT
}

type IContainerActivationHelperVtbl struct {
	IUnknownVtbl
	CanActivateClientVM uintptr
}

type IContainerActivationHelper struct {
	IUnknown
}

func (this *IContainerActivationHelper) Vtbl() *IContainerActivationHelperVtbl {
	return (*IContainerActivationHelperVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IContainerActivationHelper) CanActivateClientVM(isAllowed *int16) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().CanActivateClientVM, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(isAllowed)))
	return HRESULT(ret)
}

// C39948F0-6142-44FD-98CA-E1681A8D68B5
var IID_IClipServiceNotificationHelper = syscall.GUID{0xC39948F0, 0x6142, 0x44FD,
	[8]byte{0x98, 0xCA, 0xE1, 0x68, 0x1A, 0x8D, 0x68, 0xB5}}

type IClipServiceNotificationHelperInterface interface {
	IUnknownInterface
	ShowToast(titleText BSTR, bodyText BSTR, packageName BSTR, appId BSTR, launchCommand BSTR) HRESULT
}

type IClipServiceNotificationHelperVtbl struct {
	IUnknownVtbl
	ShowToast uintptr
}

type IClipServiceNotificationHelper struct {
	IUnknown
}

func (this *IClipServiceNotificationHelper) Vtbl() *IClipServiceNotificationHelperVtbl {
	return (*IClipServiceNotificationHelperVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IClipServiceNotificationHelper) ShowToast(titleText BSTR, bodyText BSTR, packageName BSTR, appId BSTR, launchCommand BSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().ShowToast, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(titleText)), uintptr(unsafe.Pointer(bodyText)), uintptr(unsafe.Pointer(packageName)), uintptr(unsafe.Pointer(appId)), uintptr(unsafe.Pointer(launchCommand)))
	return HRESULT(ret)
}

// 7A27FAAD-5AE6-4255-9030-C530936292E3
var IID_IDefaultBrowserSyncSettings = syscall.GUID{0x7A27FAAD, 0x5AE6, 0x4255,
	[8]byte{0x90, 0x30, 0xC5, 0x30, 0x93, 0x62, 0x92, 0xE3}}

type IDefaultBrowserSyncSettingsInterface interface {
	IUnknownInterface
	IsEnabled() BOOL
}

type IDefaultBrowserSyncSettingsVtbl struct {
	IUnknownVtbl
	IsEnabled uintptr
}

type IDefaultBrowserSyncSettings struct {
	IUnknown
}

func (this *IDefaultBrowserSyncSettings) Vtbl() *IDefaultBrowserSyncSettingsVtbl {
	return (*IDefaultBrowserSyncSettingsVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IDefaultBrowserSyncSettings) IsEnabled() BOOL {
	ret, _, _ := syscall.SyscallN(this.Vtbl().IsEnabled, uintptr(unsafe.Pointer(this)))
	return BOOL(ret)
}

// CF38ED4B-2BE7-4461-8B5E-9A466DC82AE3
var IID_IDeleteBrowsingHistory = syscall.GUID{0xCF38ED4B, 0x2BE7, 0x4461,
	[8]byte{0x8B, 0x5E, 0x9A, 0x46, 0x6D, 0xC8, 0x2A, 0xE3}}

type IDeleteBrowsingHistoryInterface interface {
	IUnknownInterface
	DeleteBrowsingHistory(dwFlags uint32) HRESULT
}

type IDeleteBrowsingHistoryVtbl struct {
	IUnknownVtbl
	DeleteBrowsingHistory uintptr
}

type IDeleteBrowsingHistory struct {
	IUnknown
}

func (this *IDeleteBrowsingHistory) Vtbl() *IDeleteBrowsingHistoryVtbl {
	return (*IDeleteBrowsingHistoryVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IDeleteBrowsingHistory) DeleteBrowsingHistory(dwFlags uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().DeleteBrowsingHistory, uintptr(unsafe.Pointer(this)), uintptr(dwFlags))
	return HRESULT(ret)
}

var (
	pQueryThreadCycleTime                uintptr
	pQueryProcessCycleTime               uintptr
	pQueryIdleProcessorCycleTime         uintptr
	pQueryIdleProcessorCycleTimeEx       uintptr
	pQueryUnbiasedInterruptTime          uintptr
	pGlobalCompact                       uintptr
	pGlobalFix                           uintptr
	pGlobalUnfix                         uintptr
	pGlobalWire                          uintptr
	pGlobalUnWire                        uintptr
	pLocalShrink                         uintptr
	pLocalCompact                        uintptr
	pSetEnvironmentStringsA              uintptr
	pSetHandleCount                      uintptr
	pRequestDeviceWakeup                 uintptr
	pCancelDeviceWakeupRequest           uintptr
	pSetMessageWaitingIndicator          uintptr
	pMulDiv                              uintptr
	pGetSystemRegistryQuota              uintptr
	pFileTimeToDosDateTime               uintptr
	pDosDateTimeToFileTime               uintptr
	pLopen_                              uintptr
	pLcreat_                             uintptr
	pLread_                              uintptr
	pLwrite_                             uintptr
	pHread_                              uintptr
	pHwrite_                             uintptr
	pLclose_                             uintptr
	pLlseek_                             uintptr
	pSignalObjectAndWait                 uintptr
	pOpenMutexA                          uintptr
	pOpenSemaphoreA                      uintptr
	pCreateWaitableTimerA                uintptr
	pOpenWaitableTimerA                  uintptr
	pCreateWaitableTimerExA              uintptr
	pGetFirmwareEnvironmentVariableA     uintptr
	pGetFirmwareEnvironmentVariableW     uintptr
	pGetFirmwareEnvironmentVariableExA   uintptr
	pGetFirmwareEnvironmentVariableExW   uintptr
	pSetFirmwareEnvironmentVariableA     uintptr
	pSetFirmwareEnvironmentVariableW     uintptr
	pSetFirmwareEnvironmentVariableExA   uintptr
	pSetFirmwareEnvironmentVariableExW   uintptr
	pIsNativeVhdBoot                     uintptr
	pGetProfileIntA                      uintptr
	pGetProfileIntW                      uintptr
	pGetProfileStringA                   uintptr
	pGetProfileStringW                   uintptr
	pWriteProfileStringA                 uintptr
	pWriteProfileStringW                 uintptr
	pGetProfileSectionA                  uintptr
	pGetProfileSectionW                  uintptr
	pWriteProfileSectionA                uintptr
	pWriteProfileSectionW                uintptr
	pGetPrivateProfileIntA               uintptr
	pGetPrivateProfileIntW               uintptr
	pGetPrivateProfileStringA            uintptr
	pGetPrivateProfileStringW            uintptr
	pWritePrivateProfileStringA          uintptr
	pWritePrivateProfileStringW          uintptr
	pGetPrivateProfileSectionA           uintptr
	pGetPrivateProfileSectionW           uintptr
	pWritePrivateProfileSectionA         uintptr
	pWritePrivateProfileSectionW         uintptr
	pGetPrivateProfileSectionNamesA      uintptr
	pGetPrivateProfileSectionNamesW      uintptr
	pGetPrivateProfileStructA            uintptr
	pGetPrivateProfileStructW            uintptr
	pWritePrivateProfileStructA          uintptr
	pWritePrivateProfileStructW          uintptr
	pIsBadHugeReadPtr                    uintptr
	pIsBadHugeWritePtr                   uintptr
	pGetComputerNameA                    uintptr
	pGetComputerNameW                    uintptr
	pDnsHostnameToComputerNameA          uintptr
	pDnsHostnameToComputerNameW          uintptr
	pGetUserNameA                        uintptr
	pGetUserNameW                        uintptr
	pIsTokenUntrusted                    uintptr
	pCancelTimerQueueTimer               uintptr
	pGetCurrentHwProfileA                uintptr
	pGetCurrentHwProfileW                uintptr
	pReplacePartitionUnit                uintptr
	pGetThreadEnabledXStateFeatures      uintptr
	pEnableProcessOptionalXStateFeatures uintptr
	pUaw_lstrcmpW                        uintptr
	pUaw_lstrcmpiW                       uintptr
	pUaw_lstrlenW                        uintptr
	pUaw_wcschr                          uintptr
	pUaw_wcscpy                          uintptr
	pUaw_wcsicmp                         uintptr
	pUaw_wcslen                          uintptr
	pUaw_wcsrchr                         uintptr
	pSendIMEMessageExA                   uintptr
	pSendIMEMessageExW                   uintptr
	pIMPGetIMEA                          uintptr
	pIMPGetIMEW                          uintptr
	pIMPQueryIMEA                        uintptr
	pIMPQueryIMEW                        uintptr
	pIMPSetIMEA                          uintptr
	pIMPSetIMEW                          uintptr
	pWINNLSGetIMEHotkey                  uintptr
	pWINNLSEnableIME                     uintptr
	pWINNLSGetEnableStatus               uintptr
)

func QueryThreadCycleTime(ThreadHandle HANDLE, CycleTime *uint64) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pQueryThreadCycleTime, libKernel32, "QueryThreadCycleTime")
	ret, _, err := syscall.SyscallN(addr, ThreadHandle, uintptr(unsafe.Pointer(CycleTime)))
	return BOOL(ret), WIN32_ERROR(err)
}

func QueryProcessCycleTime(ProcessHandle HANDLE, CycleTime *uint64) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pQueryProcessCycleTime, libKernel32, "QueryProcessCycleTime")
	ret, _, err := syscall.SyscallN(addr, ProcessHandle, uintptr(unsafe.Pointer(CycleTime)))
	return BOOL(ret), WIN32_ERROR(err)
}

func QueryIdleProcessorCycleTime(BufferLength *uint32, ProcessorIdleCycleTime *uint64) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pQueryIdleProcessorCycleTime, libKernel32, "QueryIdleProcessorCycleTime")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(BufferLength)), uintptr(unsafe.Pointer(ProcessorIdleCycleTime)))
	return BOOL(ret), WIN32_ERROR(err)
}

func QueryIdleProcessorCycleTimeEx(Group uint16, BufferLength *uint32, ProcessorIdleCycleTime *uint64) BOOL {
	addr := lazyAddr(&pQueryIdleProcessorCycleTimeEx, libKernel32, "QueryIdleProcessorCycleTimeEx")
	ret, _, _ := syscall.SyscallN(addr, uintptr(Group), uintptr(unsafe.Pointer(BufferLength)), uintptr(unsafe.Pointer(ProcessorIdleCycleTime)))
	return BOOL(ret)
}

func QueryUnbiasedInterruptTime(UnbiasedTime *uint64) BOOL {
	addr := lazyAddr(&pQueryUnbiasedInterruptTime, libKernel32, "QueryUnbiasedInterruptTime")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(UnbiasedTime)))
	return BOOL(ret)
}

func GlobalCompact(dwMinFree uint32) uintptr {
	addr := lazyAddr(&pGlobalCompact, libKernel32, "GlobalCompact")
	ret, _, _ := syscall.SyscallN(addr, uintptr(dwMinFree))
	return ret
}

func GlobalFix(hMem uintptr) {
	addr := lazyAddr(&pGlobalFix, libKernel32, "GlobalFix")
	syscall.SyscallN(addr, hMem)
}

func GlobalUnfix(hMem uintptr) {
	addr := lazyAddr(&pGlobalUnfix, libKernel32, "GlobalUnfix")
	syscall.SyscallN(addr, hMem)
}

func GlobalWire(hMem uintptr) unsafe.Pointer {
	addr := lazyAddr(&pGlobalWire, libKernel32, "GlobalWire")
	ret, _, _ := syscall.SyscallN(addr, hMem)
	return (unsafe.Pointer)(ret)
}

func GlobalUnWire(hMem uintptr) BOOL {
	addr := lazyAddr(&pGlobalUnWire, libKernel32, "GlobalUnWire")
	ret, _, _ := syscall.SyscallN(addr, hMem)
	return BOOL(ret)
}

func LocalShrink(hMem uintptr, cbNewSize uint32) uintptr {
	addr := lazyAddr(&pLocalShrink, libKernel32, "LocalShrink")
	ret, _, _ := syscall.SyscallN(addr, hMem, uintptr(cbNewSize))
	return ret
}

func LocalCompact(uMinFree uint32) uintptr {
	addr := lazyAddr(&pLocalCompact, libKernel32, "LocalCompact")
	ret, _, _ := syscall.SyscallN(addr, uintptr(uMinFree))
	return ret
}

func SetEnvironmentStringsA(NewEnvironment PSTR) BOOL {
	addr := lazyAddr(&pSetEnvironmentStringsA, libKernel32, "SetEnvironmentStringsA")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(NewEnvironment)))
	return BOOL(ret)
}

func SetHandleCount(uNumber uint32) uint32 {
	addr := lazyAddr(&pSetHandleCount, libKernel32, "SetHandleCount")
	ret, _, _ := syscall.SyscallN(addr, uintptr(uNumber))
	return uint32(ret)
}

func RequestDeviceWakeup(hDevice HANDLE) BOOL {
	addr := lazyAddr(&pRequestDeviceWakeup, libKernel32, "RequestDeviceWakeup")
	ret, _, _ := syscall.SyscallN(addr, hDevice)
	return BOOL(ret)
}

func CancelDeviceWakeupRequest(hDevice HANDLE) BOOL {
	addr := lazyAddr(&pCancelDeviceWakeupRequest, libKernel32, "CancelDeviceWakeupRequest")
	ret, _, _ := syscall.SyscallN(addr, hDevice)
	return BOOL(ret)
}

func SetMessageWaitingIndicator(hMsgIndicator HANDLE, ulMsgCount uint32) BOOL {
	addr := lazyAddr(&pSetMessageWaitingIndicator, libKernel32, "SetMessageWaitingIndicator")
	ret, _, _ := syscall.SyscallN(addr, hMsgIndicator, uintptr(ulMsgCount))
	return BOOL(ret)
}

func MulDiv(nNumber int32, nNumerator int32, nDenominator int32) int32 {
	addr := lazyAddr(&pMulDiv, libKernel32, "MulDiv")
	ret, _, _ := syscall.SyscallN(addr, uintptr(nNumber), uintptr(nNumerator), uintptr(nDenominator))
	return int32(ret)
}

func GetSystemRegistryQuota(pdwQuotaAllowed *uint32, pdwQuotaUsed *uint32) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pGetSystemRegistryQuota, libKernel32, "GetSystemRegistryQuota")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pdwQuotaAllowed)), uintptr(unsafe.Pointer(pdwQuotaUsed)))
	return BOOL(ret), WIN32_ERROR(err)
}

func FileTimeToDosDateTime(lpFileTime *FILETIME, lpFatDate *uint16, lpFatTime *uint16) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pFileTimeToDosDateTime, libKernel32, "FileTimeToDosDateTime")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpFileTime)), uintptr(unsafe.Pointer(lpFatDate)), uintptr(unsafe.Pointer(lpFatTime)))
	return BOOL(ret), WIN32_ERROR(err)
}

func DosDateTimeToFileTime(wFatDate uint16, wFatTime uint16, lpFileTime *FILETIME) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pDosDateTimeToFileTime, libKernel32, "DosDateTimeToFileTime")
	ret, _, err := syscall.SyscallN(addr, uintptr(wFatDate), uintptr(wFatTime), uintptr(unsafe.Pointer(lpFileTime)))
	return BOOL(ret), WIN32_ERROR(err)
}

func Lopen_(lpPathName PSTR, iReadWrite int32) int32 {
	addr := lazyAddr(&pLopen_, libKernel32, "_lopen")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpPathName)), uintptr(iReadWrite))
	return int32(ret)
}

func Lcreat_(lpPathName PSTR, iAttribute int32) (int32, WIN32_ERROR) {
	addr := lazyAddr(&pLcreat_, libKernel32, "_lcreat")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpPathName)), uintptr(iAttribute))
	return int32(ret), WIN32_ERROR(err)
}

func Lread_(hFile int32, lpBuffer unsafe.Pointer, uBytes uint32) uint32 {
	addr := lazyAddr(&pLread_, libKernel32, "_lread")
	ret, _, _ := syscall.SyscallN(addr, uintptr(hFile), uintptr(lpBuffer), uintptr(uBytes))
	return uint32(ret)
}

func Lwrite_(hFile int32, lpBuffer PSTR, uBytes uint32) (uint32, WIN32_ERROR) {
	addr := lazyAddr(&pLwrite_, libKernel32, "_lwrite")
	ret, _, err := syscall.SyscallN(addr, uintptr(hFile), uintptr(unsafe.Pointer(lpBuffer)), uintptr(uBytes))
	return uint32(ret), WIN32_ERROR(err)
}

func Hread_(hFile int32, lpBuffer unsafe.Pointer, lBytes int32) int32 {
	addr := lazyAddr(&pHread_, libKernel32, "_hread")
	ret, _, _ := syscall.SyscallN(addr, uintptr(hFile), uintptr(lpBuffer), uintptr(lBytes))
	return int32(ret)
}

func Hwrite_(hFile int32, lpBuffer PSTR, lBytes int32) int32 {
	addr := lazyAddr(&pHwrite_, libKernel32, "_hwrite")
	ret, _, _ := syscall.SyscallN(addr, uintptr(hFile), uintptr(unsafe.Pointer(lpBuffer)), uintptr(lBytes))
	return int32(ret)
}

func Lclose_(hFile int32) int32 {
	addr := lazyAddr(&pLclose_, libKernel32, "_lclose")
	ret, _, _ := syscall.SyscallN(addr, uintptr(hFile))
	return int32(ret)
}

func Llseek_(hFile int32, lOffset int32, iOrigin int32) (int32, WIN32_ERROR) {
	addr := lazyAddr(&pLlseek_, libKernel32, "_llseek")
	ret, _, err := syscall.SyscallN(addr, uintptr(hFile), uintptr(lOffset), uintptr(iOrigin))
	return int32(ret), WIN32_ERROR(err)
}

func SignalObjectAndWait(hObjectToSignal HANDLE, hObjectToWaitOn HANDLE, dwMilliseconds uint32, bAlertable BOOL) (uint32, WIN32_ERROR) {
	addr := lazyAddr(&pSignalObjectAndWait, libKernel32, "SignalObjectAndWait")
	ret, _, err := syscall.SyscallN(addr, hObjectToSignal, hObjectToWaitOn, uintptr(dwMilliseconds), uintptr(bAlertable))
	return uint32(ret), WIN32_ERROR(err)
}

func OpenMutexA(dwDesiredAccess uint32, bInheritHandle BOOL, lpName PSTR) HANDLE {
	addr := lazyAddr(&pOpenMutexA, libKernel32, "OpenMutexA")
	ret, _, _ := syscall.SyscallN(addr, uintptr(dwDesiredAccess), uintptr(bInheritHandle), uintptr(unsafe.Pointer(lpName)))
	return ret
}

func OpenSemaphoreA(dwDesiredAccess uint32, bInheritHandle BOOL, lpName PSTR) HANDLE {
	addr := lazyAddr(&pOpenSemaphoreA, libKernel32, "OpenSemaphoreA")
	ret, _, _ := syscall.SyscallN(addr, uintptr(dwDesiredAccess), uintptr(bInheritHandle), uintptr(unsafe.Pointer(lpName)))
	return ret
}

func CreateWaitableTimerA(lpTimerAttributes *SECURITY_ATTRIBUTES, bManualReset BOOL, lpTimerName PSTR) HANDLE {
	addr := lazyAddr(&pCreateWaitableTimerA, libKernel32, "CreateWaitableTimerA")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpTimerAttributes)), uintptr(bManualReset), uintptr(unsafe.Pointer(lpTimerName)))
	return ret
}

func OpenWaitableTimerA(dwDesiredAccess uint32, bInheritHandle BOOL, lpTimerName PSTR) HANDLE {
	addr := lazyAddr(&pOpenWaitableTimerA, libKernel32, "OpenWaitableTimerA")
	ret, _, _ := syscall.SyscallN(addr, uintptr(dwDesiredAccess), uintptr(bInheritHandle), uintptr(unsafe.Pointer(lpTimerName)))
	return ret
}

func CreateWaitableTimerExA(lpTimerAttributes *SECURITY_ATTRIBUTES, lpTimerName PSTR, dwFlags uint32, dwDesiredAccess uint32) HANDLE {
	addr := lazyAddr(&pCreateWaitableTimerExA, libKernel32, "CreateWaitableTimerExA")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpTimerAttributes)), uintptr(unsafe.Pointer(lpTimerName)), uintptr(dwFlags), uintptr(dwDesiredAccess))
	return ret
}

func GetFirmwareEnvironmentVariableA(lpName PSTR, lpGuid PSTR, pBuffer unsafe.Pointer, nSize uint32) (uint32, WIN32_ERROR) {
	addr := lazyAddr(&pGetFirmwareEnvironmentVariableA, libKernel32, "GetFirmwareEnvironmentVariableA")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpName)), uintptr(unsafe.Pointer(lpGuid)), uintptr(pBuffer), uintptr(nSize))
	return uint32(ret), WIN32_ERROR(err)
}

var GetFirmwareEnvironmentVariable = GetFirmwareEnvironmentVariableW

func GetFirmwareEnvironmentVariableW(lpName PWSTR, lpGuid PWSTR, pBuffer unsafe.Pointer, nSize uint32) (uint32, WIN32_ERROR) {
	addr := lazyAddr(&pGetFirmwareEnvironmentVariableW, libKernel32, "GetFirmwareEnvironmentVariableW")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpName)), uintptr(unsafe.Pointer(lpGuid)), uintptr(pBuffer), uintptr(nSize))
	return uint32(ret), WIN32_ERROR(err)
}

func GetFirmwareEnvironmentVariableExA(lpName PSTR, lpGuid PSTR, pBuffer unsafe.Pointer, nSize uint32, pdwAttribubutes *uint32) (uint32, WIN32_ERROR) {
	addr := lazyAddr(&pGetFirmwareEnvironmentVariableExA, libKernel32, "GetFirmwareEnvironmentVariableExA")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpName)), uintptr(unsafe.Pointer(lpGuid)), uintptr(pBuffer), uintptr(nSize), uintptr(unsafe.Pointer(pdwAttribubutes)))
	return uint32(ret), WIN32_ERROR(err)
}

var GetFirmwareEnvironmentVariableEx = GetFirmwareEnvironmentVariableExW

func GetFirmwareEnvironmentVariableExW(lpName PWSTR, lpGuid PWSTR, pBuffer unsafe.Pointer, nSize uint32, pdwAttribubutes *uint32) (uint32, WIN32_ERROR) {
	addr := lazyAddr(&pGetFirmwareEnvironmentVariableExW, libKernel32, "GetFirmwareEnvironmentVariableExW")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpName)), uintptr(unsafe.Pointer(lpGuid)), uintptr(pBuffer), uintptr(nSize), uintptr(unsafe.Pointer(pdwAttribubutes)))
	return uint32(ret), WIN32_ERROR(err)
}

func SetFirmwareEnvironmentVariableA(lpName PSTR, lpGuid PSTR, pValue unsafe.Pointer, nSize uint32) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pSetFirmwareEnvironmentVariableA, libKernel32, "SetFirmwareEnvironmentVariableA")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpName)), uintptr(unsafe.Pointer(lpGuid)), uintptr(pValue), uintptr(nSize))
	return BOOL(ret), WIN32_ERROR(err)
}

var SetFirmwareEnvironmentVariable = SetFirmwareEnvironmentVariableW

func SetFirmwareEnvironmentVariableW(lpName PWSTR, lpGuid PWSTR, pValue unsafe.Pointer, nSize uint32) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pSetFirmwareEnvironmentVariableW, libKernel32, "SetFirmwareEnvironmentVariableW")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpName)), uintptr(unsafe.Pointer(lpGuid)), uintptr(pValue), uintptr(nSize))
	return BOOL(ret), WIN32_ERROR(err)
}

func SetFirmwareEnvironmentVariableExA(lpName PSTR, lpGuid PSTR, pValue unsafe.Pointer, nSize uint32, dwAttributes uint32) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pSetFirmwareEnvironmentVariableExA, libKernel32, "SetFirmwareEnvironmentVariableExA")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpName)), uintptr(unsafe.Pointer(lpGuid)), uintptr(pValue), uintptr(nSize), uintptr(dwAttributes))
	return BOOL(ret), WIN32_ERROR(err)
}

var SetFirmwareEnvironmentVariableEx = SetFirmwareEnvironmentVariableExW

func SetFirmwareEnvironmentVariableExW(lpName PWSTR, lpGuid PWSTR, pValue unsafe.Pointer, nSize uint32, dwAttributes uint32) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pSetFirmwareEnvironmentVariableExW, libKernel32, "SetFirmwareEnvironmentVariableExW")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpName)), uintptr(unsafe.Pointer(lpGuid)), uintptr(pValue), uintptr(nSize), uintptr(dwAttributes))
	return BOOL(ret), WIN32_ERROR(err)
}

func IsNativeVhdBoot(NativeVhdBoot *BOOL) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pIsNativeVhdBoot, libKernel32, "IsNativeVhdBoot")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(NativeVhdBoot)))
	return BOOL(ret), WIN32_ERROR(err)
}

func GetProfileIntA(lpAppName PSTR, lpKeyName PSTR, nDefault int32) uint32 {
	addr := lazyAddr(&pGetProfileIntA, libKernel32, "GetProfileIntA")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpAppName)), uintptr(unsafe.Pointer(lpKeyName)), uintptr(nDefault))
	return uint32(ret)
}

var GetProfileInt = GetProfileIntW

func GetProfileIntW(lpAppName PWSTR, lpKeyName PWSTR, nDefault int32) uint32 {
	addr := lazyAddr(&pGetProfileIntW, libKernel32, "GetProfileIntW")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpAppName)), uintptr(unsafe.Pointer(lpKeyName)), uintptr(nDefault))
	return uint32(ret)
}

func GetProfileStringA(lpAppName PSTR, lpKeyName PSTR, lpDefault PSTR, lpReturnedString PSTR, nSize uint32) uint32 {
	addr := lazyAddr(&pGetProfileStringA, libKernel32, "GetProfileStringA")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpAppName)), uintptr(unsafe.Pointer(lpKeyName)), uintptr(unsafe.Pointer(lpDefault)), uintptr(unsafe.Pointer(lpReturnedString)), uintptr(nSize))
	return uint32(ret)
}

var GetProfileString = GetProfileStringW

func GetProfileStringW(lpAppName PWSTR, lpKeyName PWSTR, lpDefault PWSTR, lpReturnedString PWSTR, nSize uint32) uint32 {
	addr := lazyAddr(&pGetProfileStringW, libKernel32, "GetProfileStringW")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpAppName)), uintptr(unsafe.Pointer(lpKeyName)), uintptr(unsafe.Pointer(lpDefault)), uintptr(unsafe.Pointer(lpReturnedString)), uintptr(nSize))
	return uint32(ret)
}

func WriteProfileStringA(lpAppName PSTR, lpKeyName PSTR, lpString PSTR) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pWriteProfileStringA, libKernel32, "WriteProfileStringA")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpAppName)), uintptr(unsafe.Pointer(lpKeyName)), uintptr(unsafe.Pointer(lpString)))
	return BOOL(ret), WIN32_ERROR(err)
}

var WriteProfileString = WriteProfileStringW

func WriteProfileStringW(lpAppName PWSTR, lpKeyName PWSTR, lpString PWSTR) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pWriteProfileStringW, libKernel32, "WriteProfileStringW")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpAppName)), uintptr(unsafe.Pointer(lpKeyName)), uintptr(unsafe.Pointer(lpString)))
	return BOOL(ret), WIN32_ERROR(err)
}

func GetProfileSectionA(lpAppName PSTR, lpReturnedString PSTR, nSize uint32) uint32 {
	addr := lazyAddr(&pGetProfileSectionA, libKernel32, "GetProfileSectionA")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpAppName)), uintptr(unsafe.Pointer(lpReturnedString)), uintptr(nSize))
	return uint32(ret)
}

var GetProfileSection = GetProfileSectionW

func GetProfileSectionW(lpAppName PWSTR, lpReturnedString PWSTR, nSize uint32) uint32 {
	addr := lazyAddr(&pGetProfileSectionW, libKernel32, "GetProfileSectionW")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpAppName)), uintptr(unsafe.Pointer(lpReturnedString)), uintptr(nSize))
	return uint32(ret)
}

func WriteProfileSectionA(lpAppName PSTR, lpString PSTR) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pWriteProfileSectionA, libKernel32, "WriteProfileSectionA")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpAppName)), uintptr(unsafe.Pointer(lpString)))
	return BOOL(ret), WIN32_ERROR(err)
}

var WriteProfileSection = WriteProfileSectionW

func WriteProfileSectionW(lpAppName PWSTR, lpString PWSTR) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pWriteProfileSectionW, libKernel32, "WriteProfileSectionW")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpAppName)), uintptr(unsafe.Pointer(lpString)))
	return BOOL(ret), WIN32_ERROR(err)
}

func GetPrivateProfileIntA(lpAppName PSTR, lpKeyName PSTR, nDefault int32, lpFileName PSTR) uint32 {
	addr := lazyAddr(&pGetPrivateProfileIntA, libKernel32, "GetPrivateProfileIntA")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpAppName)), uintptr(unsafe.Pointer(lpKeyName)), uintptr(nDefault), uintptr(unsafe.Pointer(lpFileName)))
	return uint32(ret)
}

var GetPrivateProfileInt = GetPrivateProfileIntW

func GetPrivateProfileIntW(lpAppName PWSTR, lpKeyName PWSTR, nDefault int32, lpFileName PWSTR) uint32 {
	addr := lazyAddr(&pGetPrivateProfileIntW, libKernel32, "GetPrivateProfileIntW")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpAppName)), uintptr(unsafe.Pointer(lpKeyName)), uintptr(nDefault), uintptr(unsafe.Pointer(lpFileName)))
	return uint32(ret)
}

func GetPrivateProfileStringA(lpAppName PSTR, lpKeyName PSTR, lpDefault PSTR, lpReturnedString PSTR, nSize uint32, lpFileName PSTR) (uint32, WIN32_ERROR) {
	addr := lazyAddr(&pGetPrivateProfileStringA, libKernel32, "GetPrivateProfileStringA")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpAppName)), uintptr(unsafe.Pointer(lpKeyName)), uintptr(unsafe.Pointer(lpDefault)), uintptr(unsafe.Pointer(lpReturnedString)), uintptr(nSize), uintptr(unsafe.Pointer(lpFileName)))
	return uint32(ret), WIN32_ERROR(err)
}

var GetPrivateProfileString = GetPrivateProfileStringW

func GetPrivateProfileStringW(lpAppName PWSTR, lpKeyName PWSTR, lpDefault PWSTR, lpReturnedString PWSTR, nSize uint32, lpFileName PWSTR) (uint32, WIN32_ERROR) {
	addr := lazyAddr(&pGetPrivateProfileStringW, libKernel32, "GetPrivateProfileStringW")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpAppName)), uintptr(unsafe.Pointer(lpKeyName)), uintptr(unsafe.Pointer(lpDefault)), uintptr(unsafe.Pointer(lpReturnedString)), uintptr(nSize), uintptr(unsafe.Pointer(lpFileName)))
	return uint32(ret), WIN32_ERROR(err)
}

func WritePrivateProfileStringA(lpAppName PSTR, lpKeyName PSTR, lpString PSTR, lpFileName PSTR) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pWritePrivateProfileStringA, libKernel32, "WritePrivateProfileStringA")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpAppName)), uintptr(unsafe.Pointer(lpKeyName)), uintptr(unsafe.Pointer(lpString)), uintptr(unsafe.Pointer(lpFileName)))
	return BOOL(ret), WIN32_ERROR(err)
}

var WritePrivateProfileString = WritePrivateProfileStringW

func WritePrivateProfileStringW(lpAppName PWSTR, lpKeyName PWSTR, lpString PWSTR, lpFileName PWSTR) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pWritePrivateProfileStringW, libKernel32, "WritePrivateProfileStringW")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpAppName)), uintptr(unsafe.Pointer(lpKeyName)), uintptr(unsafe.Pointer(lpString)), uintptr(unsafe.Pointer(lpFileName)))
	return BOOL(ret), WIN32_ERROR(err)
}

func GetPrivateProfileSectionA(lpAppName PSTR, lpReturnedString PSTR, nSize uint32, lpFileName PSTR) uint32 {
	addr := lazyAddr(&pGetPrivateProfileSectionA, libKernel32, "GetPrivateProfileSectionA")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpAppName)), uintptr(unsafe.Pointer(lpReturnedString)), uintptr(nSize), uintptr(unsafe.Pointer(lpFileName)))
	return uint32(ret)
}

var GetPrivateProfileSection = GetPrivateProfileSectionW

func GetPrivateProfileSectionW(lpAppName PWSTR, lpReturnedString PWSTR, nSize uint32, lpFileName PWSTR) uint32 {
	addr := lazyAddr(&pGetPrivateProfileSectionW, libKernel32, "GetPrivateProfileSectionW")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpAppName)), uintptr(unsafe.Pointer(lpReturnedString)), uintptr(nSize), uintptr(unsafe.Pointer(lpFileName)))
	return uint32(ret)
}

func WritePrivateProfileSectionA(lpAppName PSTR, lpString PSTR, lpFileName PSTR) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pWritePrivateProfileSectionA, libKernel32, "WritePrivateProfileSectionA")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpAppName)), uintptr(unsafe.Pointer(lpString)), uintptr(unsafe.Pointer(lpFileName)))
	return BOOL(ret), WIN32_ERROR(err)
}

var WritePrivateProfileSection = WritePrivateProfileSectionW

func WritePrivateProfileSectionW(lpAppName PWSTR, lpString PWSTR, lpFileName PWSTR) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pWritePrivateProfileSectionW, libKernel32, "WritePrivateProfileSectionW")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpAppName)), uintptr(unsafe.Pointer(lpString)), uintptr(unsafe.Pointer(lpFileName)))
	return BOOL(ret), WIN32_ERROR(err)
}

func GetPrivateProfileSectionNamesA(lpszReturnBuffer PSTR, nSize uint32, lpFileName PSTR) uint32 {
	addr := lazyAddr(&pGetPrivateProfileSectionNamesA, libKernel32, "GetPrivateProfileSectionNamesA")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpszReturnBuffer)), uintptr(nSize), uintptr(unsafe.Pointer(lpFileName)))
	return uint32(ret)
}

var GetPrivateProfileSectionNames = GetPrivateProfileSectionNamesW

func GetPrivateProfileSectionNamesW(lpszReturnBuffer PWSTR, nSize uint32, lpFileName PWSTR) uint32 {
	addr := lazyAddr(&pGetPrivateProfileSectionNamesW, libKernel32, "GetPrivateProfileSectionNamesW")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpszReturnBuffer)), uintptr(nSize), uintptr(unsafe.Pointer(lpFileName)))
	return uint32(ret)
}

func GetPrivateProfileStructA(lpszSection PSTR, lpszKey PSTR, lpStruct unsafe.Pointer, uSizeStruct uint32, szFile PSTR) BOOL {
	addr := lazyAddr(&pGetPrivateProfileStructA, libKernel32, "GetPrivateProfileStructA")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpszSection)), uintptr(unsafe.Pointer(lpszKey)), uintptr(lpStruct), uintptr(uSizeStruct), uintptr(unsafe.Pointer(szFile)))
	return BOOL(ret)
}

var GetPrivateProfileStruct = GetPrivateProfileStructW

func GetPrivateProfileStructW(lpszSection PWSTR, lpszKey PWSTR, lpStruct unsafe.Pointer, uSizeStruct uint32, szFile PWSTR) BOOL {
	addr := lazyAddr(&pGetPrivateProfileStructW, libKernel32, "GetPrivateProfileStructW")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpszSection)), uintptr(unsafe.Pointer(lpszKey)), uintptr(lpStruct), uintptr(uSizeStruct), uintptr(unsafe.Pointer(szFile)))
	return BOOL(ret)
}

func WritePrivateProfileStructA(lpszSection PSTR, lpszKey PSTR, lpStruct unsafe.Pointer, uSizeStruct uint32, szFile PSTR) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pWritePrivateProfileStructA, libKernel32, "WritePrivateProfileStructA")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpszSection)), uintptr(unsafe.Pointer(lpszKey)), uintptr(lpStruct), uintptr(uSizeStruct), uintptr(unsafe.Pointer(szFile)))
	return BOOL(ret), WIN32_ERROR(err)
}

var WritePrivateProfileStruct = WritePrivateProfileStructW

func WritePrivateProfileStructW(lpszSection PWSTR, lpszKey PWSTR, lpStruct unsafe.Pointer, uSizeStruct uint32, szFile PWSTR) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pWritePrivateProfileStructW, libKernel32, "WritePrivateProfileStructW")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpszSection)), uintptr(unsafe.Pointer(lpszKey)), uintptr(lpStruct), uintptr(uSizeStruct), uintptr(unsafe.Pointer(szFile)))
	return BOOL(ret), WIN32_ERROR(err)
}

func IsBadHugeReadPtr(lp unsafe.Pointer, ucb uintptr) BOOL {
	addr := lazyAddr(&pIsBadHugeReadPtr, libKernel32, "IsBadHugeReadPtr")
	ret, _, _ := syscall.SyscallN(addr, uintptr(lp), ucb)
	return BOOL(ret)
}

func IsBadHugeWritePtr(lp unsafe.Pointer, ucb uintptr) BOOL {
	addr := lazyAddr(&pIsBadHugeWritePtr, libKernel32, "IsBadHugeWritePtr")
	ret, _, _ := syscall.SyscallN(addr, uintptr(lp), ucb)
	return BOOL(ret)
}

func GetComputerNameA(lpBuffer PSTR, nSize *uint32) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pGetComputerNameA, libKernel32, "GetComputerNameA")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpBuffer)), uintptr(unsafe.Pointer(nSize)))
	return BOOL(ret), WIN32_ERROR(err)
}

var GetComputerName = GetComputerNameW

func GetComputerNameW(lpBuffer PWSTR, nSize *uint32) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pGetComputerNameW, libKernel32, "GetComputerNameW")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpBuffer)), uintptr(unsafe.Pointer(nSize)))
	return BOOL(ret), WIN32_ERROR(err)
}

func DnsHostnameToComputerNameA(Hostname PSTR, ComputerName PSTR, nSize *uint32) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pDnsHostnameToComputerNameA, libKernel32, "DnsHostnameToComputerNameA")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(Hostname)), uintptr(unsafe.Pointer(ComputerName)), uintptr(unsafe.Pointer(nSize)))
	return BOOL(ret), WIN32_ERROR(err)
}

var DnsHostnameToComputerName = DnsHostnameToComputerNameW

func DnsHostnameToComputerNameW(Hostname PWSTR, ComputerName PWSTR, nSize *uint32) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pDnsHostnameToComputerNameW, libKernel32, "DnsHostnameToComputerNameW")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(Hostname)), uintptr(unsafe.Pointer(ComputerName)), uintptr(unsafe.Pointer(nSize)))
	return BOOL(ret), WIN32_ERROR(err)
}

func GetUserNameA(lpBuffer PSTR, pcbBuffer *uint32) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pGetUserNameA, libAdvapi32, "GetUserNameA")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpBuffer)), uintptr(unsafe.Pointer(pcbBuffer)))
	return BOOL(ret), WIN32_ERROR(err)
}

var GetUserName = GetUserNameW

func GetUserNameW(lpBuffer PWSTR, pcbBuffer *uint32) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pGetUserNameW, libAdvapi32, "GetUserNameW")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpBuffer)), uintptr(unsafe.Pointer(pcbBuffer)))
	return BOOL(ret), WIN32_ERROR(err)
}

func IsTokenUntrusted(TokenHandle HANDLE) BOOL {
	addr := lazyAddr(&pIsTokenUntrusted, libAdvapi32, "IsTokenUntrusted")
	ret, _, _ := syscall.SyscallN(addr, TokenHandle)
	return BOOL(ret)
}

func CancelTimerQueueTimer(TimerQueue HANDLE, Timer HANDLE) BOOL {
	addr := lazyAddr(&pCancelTimerQueueTimer, libKernel32, "CancelTimerQueueTimer")
	ret, _, _ := syscall.SyscallN(addr, TimerQueue, Timer)
	return BOOL(ret)
}

func GetCurrentHwProfileA(lpHwProfileInfo *HW_PROFILE_INFOA) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pGetCurrentHwProfileA, libAdvapi32, "GetCurrentHwProfileA")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpHwProfileInfo)))
	return BOOL(ret), WIN32_ERROR(err)
}

var GetCurrentHwProfile = GetCurrentHwProfileW

func GetCurrentHwProfileW(lpHwProfileInfo *HW_PROFILE_INFOW) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pGetCurrentHwProfileW, libAdvapi32, "GetCurrentHwProfileW")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpHwProfileInfo)))
	return BOOL(ret), WIN32_ERROR(err)
}

func ReplacePartitionUnit(TargetPartition PWSTR, SparePartition PWSTR, Flags uint32) BOOL {
	addr := lazyAddr(&pReplacePartitionUnit, libKernel32, "ReplacePartitionUnit")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(TargetPartition)), uintptr(unsafe.Pointer(SparePartition)), uintptr(Flags))
	return BOOL(ret)
}

func GetThreadEnabledXStateFeatures() uint64 {
	addr := lazyAddr(&pGetThreadEnabledXStateFeatures, libKernel32, "GetThreadEnabledXStateFeatures")
	ret, _, _ := syscall.SyscallN(addr)
	return uint64(ret)
}

func EnableProcessOptionalXStateFeatures(Features uint64) BOOL {
	addr := lazyAddr(&pEnableProcessOptionalXStateFeatures, libKernel32, "EnableProcessOptionalXStateFeatures")
	ret, _, _ := syscall.SyscallN(addr, uintptr(Features))
	return BOOL(ret)
}

func Uaw_lstrcmpW(String1 *uint16, String2 *uint16) int32 {
	addr := lazyAddr(&pUaw_lstrcmpW, libKernel32, "uaw_lstrcmpW")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(String1)), uintptr(unsafe.Pointer(String2)))
	return int32(ret)
}

func Uaw_lstrcmpiW(String1 *uint16, String2 *uint16) int32 {
	addr := lazyAddr(&pUaw_lstrcmpiW, libKernel32, "uaw_lstrcmpiW")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(String1)), uintptr(unsafe.Pointer(String2)))
	return int32(ret)
}

func Uaw_lstrlenW(String *uint16) int32 {
	addr := lazyAddr(&pUaw_lstrlenW, libKernel32, "uaw_lstrlenW")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(String)))
	return int32(ret)
}

func Uaw_wcschr(String *uint16, Character uint16) *uint16 {
	addr := lazyAddr(&pUaw_wcschr, libKernel32, "uaw_wcschr")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(String)), uintptr(Character))
	return (*uint16)(unsafe.Pointer(ret))
}

func Uaw_wcscpy(Destination *uint16, Source *uint16) *uint16 {
	addr := lazyAddr(&pUaw_wcscpy, libKernel32, "uaw_wcscpy")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(Destination)), uintptr(unsafe.Pointer(Source)))
	return (*uint16)(unsafe.Pointer(ret))
}

func Uaw_wcsicmp(String1 *uint16, String2 *uint16) int32 {
	addr := lazyAddr(&pUaw_wcsicmp, libKernel32, "uaw_wcsicmp")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(String1)), uintptr(unsafe.Pointer(String2)))
	return int32(ret)
}

func Uaw_wcslen(String *uint16) uintptr {
	addr := lazyAddr(&pUaw_wcslen, libKernel32, "uaw_wcslen")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(String)))
	return ret
}

func Uaw_wcsrchr(String *uint16, Character uint16) *uint16 {
	addr := lazyAddr(&pUaw_wcsrchr, libKernel32, "uaw_wcsrchr")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(String)), uintptr(Character))
	return (*uint16)(unsafe.Pointer(ret))
}

func SendIMEMessageExA(param0 HWND, param1 LPARAM) LRESULT {
	addr := lazyAddr(&pSendIMEMessageExA, libUser32, "SendIMEMessageExA")
	ret, _, _ := syscall.SyscallN(addr, param0, param1)
	return ret
}

var SendIMEMessageEx = SendIMEMessageExW

func SendIMEMessageExW(param0 HWND, param1 LPARAM) LRESULT {
	addr := lazyAddr(&pSendIMEMessageExW, libUser32, "SendIMEMessageExW")
	ret, _, _ := syscall.SyscallN(addr, param0, param1)
	return ret
}

func IMPGetIMEA(param0 HWND, param1 *IMEPROA) BOOL {
	addr := lazyAddr(&pIMPGetIMEA, libUser32, "IMPGetIMEA")
	ret, _, _ := syscall.SyscallN(addr, param0, uintptr(unsafe.Pointer(param1)))
	return BOOL(ret)
}

var IMPGetIME = IMPGetIMEW

func IMPGetIMEW(param0 HWND, param1 *IMEPROW) BOOL {
	addr := lazyAddr(&pIMPGetIMEW, libUser32, "IMPGetIMEW")
	ret, _, _ := syscall.SyscallN(addr, param0, uintptr(unsafe.Pointer(param1)))
	return BOOL(ret)
}

func IMPQueryIMEA(param0 *IMEPROA) BOOL {
	addr := lazyAddr(&pIMPQueryIMEA, libUser32, "IMPQueryIMEA")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(param0)))
	return BOOL(ret)
}

var IMPQueryIME = IMPQueryIMEW

func IMPQueryIMEW(param0 *IMEPROW) BOOL {
	addr := lazyAddr(&pIMPQueryIMEW, libUser32, "IMPQueryIMEW")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(param0)))
	return BOOL(ret)
}

func IMPSetIMEA(param0 HWND, param1 *IMEPROA) BOOL {
	addr := lazyAddr(&pIMPSetIMEA, libUser32, "IMPSetIMEA")
	ret, _, _ := syscall.SyscallN(addr, param0, uintptr(unsafe.Pointer(param1)))
	return BOOL(ret)
}

var IMPSetIME = IMPSetIMEW

func IMPSetIMEW(param0 HWND, param1 *IMEPROW) BOOL {
	addr := lazyAddr(&pIMPSetIMEW, libUser32, "IMPSetIMEW")
	ret, _, _ := syscall.SyscallN(addr, param0, uintptr(unsafe.Pointer(param1)))
	return BOOL(ret)
}

func WINNLSGetIMEHotkey(param0 HWND) uint32 {
	addr := lazyAddr(&pWINNLSGetIMEHotkey, libUser32, "WINNLSGetIMEHotkey")
	ret, _, _ := syscall.SyscallN(addr, param0)
	return uint32(ret)
}

func WINNLSEnableIME(param0 HWND, param1 BOOL) BOOL {
	addr := lazyAddr(&pWINNLSEnableIME, libUser32, "WINNLSEnableIME")
	ret, _, _ := syscall.SyscallN(addr, param0, uintptr(param1))
	return BOOL(ret)
}

func WINNLSGetEnableStatus(param0 HWND) BOOL {
	addr := lazyAddr(&pWINNLSGetEnableStatus, libUser32, "WINNLSGetEnableStatus")
	ret, _, _ := syscall.SyscallN(addr, param0)
	return BOOL(ret)
}
