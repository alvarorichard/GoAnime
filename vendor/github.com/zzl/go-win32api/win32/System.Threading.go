package win32

import (
	"syscall"
	"unsafe"
)

type (
	TimerQueueHandle             = uintptr
	PTP_POOL                     = uintptr
	NamespaceHandle              = uintptr
	BoundaryDescriptorHandle     = uintptr
	LPPROC_THREAD_ATTRIBUTE_LIST = unsafe.Pointer
)

const (
	WAIT_OBJECT_0                                         uint32 = 0x0
	WAIT_ABANDONED                                        uint32 = 0x80
	WAIT_ABANDONED_0                                      uint32 = 0x80
	WAIT_IO_COMPLETION                                    uint32 = 0xc0
	PRIVATE_NAMESPACE_FLAG_DESTROY                        uint32 = 0x1
	PROC_THREAD_ATTRIBUTE_REPLACE_VALUE                   uint32 = 0x1
	THREAD_POWER_THROTTLING_CURRENT_VERSION               uint32 = 0x1
	THREAD_POWER_THROTTLING_EXECUTION_SPEED               uint32 = 0x1
	THREAD_POWER_THROTTLING_VALID_FLAGS                   uint32 = 0x1
	PME_CURRENT_VERSION                                   uint32 = 0x1
	PME_FAILFAST_ON_COMMIT_FAIL_DISABLE                   uint32 = 0x0
	PME_FAILFAST_ON_COMMIT_FAIL_ENABLE                    uint32 = 0x1
	PROCESS_POWER_THROTTLING_CURRENT_VERSION              uint32 = 0x1
	PROCESS_POWER_THROTTLING_EXECUTION_SPEED              uint32 = 0x1
	PROCESS_POWER_THROTTLING_IGNORE_TIMER_RESOLUTION      uint32 = 0x4
	PROCESS_LEAP_SECOND_INFO_FLAG_ENABLE_SIXTY_SECOND     uint32 = 0x1
	PROCESS_LEAP_SECOND_INFO_VALID_FLAGS                  uint32 = 0x1
	INIT_ONCE_CHECK_ONLY                                  uint32 = 0x1
	INIT_ONCE_ASYNC                                       uint32 = 0x2
	INIT_ONCE_INIT_FAILED                                 uint32 = 0x4
	INIT_ONCE_CTX_RESERVED_BITS                           uint32 = 0x2
	CONDITION_VARIABLE_LOCKMODE_SHARED                    uint32 = 0x1
	MUTEX_MODIFY_STATE                                    uint32 = 0x1
	CREATE_MUTEX_INITIAL_OWNER                            uint32 = 0x1
	CREATE_WAITABLE_TIMER_MANUAL_RESET                    uint32 = 0x1
	CREATE_WAITABLE_TIMER_HIGH_RESOLUTION                 uint32 = 0x2
	SYNCHRONIZATION_BARRIER_FLAGS_SPIN_ONLY               uint32 = 0x1
	SYNCHRONIZATION_BARRIER_FLAGS_BLOCK_ONLY              uint32 = 0x2
	SYNCHRONIZATION_BARRIER_FLAGS_NO_DELETE               uint32 = 0x4
	PROC_THREAD_ATTRIBUTE_PARENT_PROCESS                  uint32 = 0x20000
	PROC_THREAD_ATTRIBUTE_HANDLE_LIST                     uint32 = 0x20002
	PROC_THREAD_ATTRIBUTE_GROUP_AFFINITY                  uint32 = 0x30003
	PROC_THREAD_ATTRIBUTE_PREFERRED_NODE                  uint32 = 0x20004
	PROC_THREAD_ATTRIBUTE_IDEAL_PROCESSOR                 uint32 = 0x30005
	PROC_THREAD_ATTRIBUTE_UMS_THREAD                      uint32 = 0x30006
	PROC_THREAD_ATTRIBUTE_MITIGATION_POLICY               uint32 = 0x20007
	PROC_THREAD_ATTRIBUTE_SECURITY_CAPABILITIES           uint32 = 0x20009
	PROC_THREAD_ATTRIBUTE_PROTECTION_LEVEL                uint32 = 0x2000b
	PROC_THREAD_ATTRIBUTE_PSEUDOCONSOLE                   uint32 = 0x20016
	PROC_THREAD_ATTRIBUTE_MACHINE_TYPE                    uint32 = 0x20019
	PROC_THREAD_ATTRIBUTE_ENABLE_OPTIONAL_XSTATE_FEATURES uint32 = 0x3001b
	PROC_THREAD_ATTRIBUTE_WIN32K_FILTER                   uint32 = 0x20010
	PROC_THREAD_ATTRIBUTE_JOB_LIST                        uint32 = 0x2000d
	PROC_THREAD_ATTRIBUTE_CHILD_PROCESS_POLICY            uint32 = 0x2000e
	PROC_THREAD_ATTRIBUTE_ALL_APPLICATION_PACKAGES_POLICY uint32 = 0x2000f
	PROC_THREAD_ATTRIBUTE_DESKTOP_APP_POLICY              uint32 = 0x20012
	PROC_THREAD_ATTRIBUTE_MITIGATION_AUDIT_POLICY         uint32 = 0x20018
	PROC_THREAD_ATTRIBUTE_COMPONENT_FILTER                uint32 = 0x2001a
)

// enums

// enum
// flags
type THREAD_CREATION_FLAGS uint32

const (
	THREAD_CREATE_RUN_IMMEDIATELY     THREAD_CREATION_FLAGS = 0
	THREAD_CREATE_SUSPENDED           THREAD_CREATION_FLAGS = 4
	STACK_SIZE_PARAM_IS_A_RESERVATION THREAD_CREATION_FLAGS = 65536
)

// enum
type THREAD_PRIORITY int32

const (
	THREAD_MODE_BACKGROUND_BEGIN  THREAD_PRIORITY = 65536
	THREAD_MODE_BACKGROUND_END    THREAD_PRIORITY = 131072
	THREAD_PRIORITY_ABOVE_NORMAL  THREAD_PRIORITY = 1
	THREAD_PRIORITY_BELOW_NORMAL  THREAD_PRIORITY = -1
	THREAD_PRIORITY_HIGHEST       THREAD_PRIORITY = 2
	THREAD_PRIORITY_IDLE          THREAD_PRIORITY = -15
	THREAD_PRIORITY_MIN           THREAD_PRIORITY = -2
	THREAD_PRIORITY_LOWEST        THREAD_PRIORITY = -2
	THREAD_PRIORITY_NORMAL        THREAD_PRIORITY = 0
	THREAD_PRIORITY_TIME_CRITICAL THREAD_PRIORITY = 15
)

// enum
// flags
type WORKER_THREAD_FLAGS uint32

const (
	WT_EXECUTEDEFAULT            WORKER_THREAD_FLAGS = 0
	WT_EXECUTEINIOTHREAD         WORKER_THREAD_FLAGS = 1
	WT_EXECUTEINPERSISTENTTHREAD WORKER_THREAD_FLAGS = 128
	WT_EXECUTEINWAITTHREAD       WORKER_THREAD_FLAGS = 4
	WT_EXECUTELONGFUNCTION       WORKER_THREAD_FLAGS = 16
	WT_EXECUTEONLYONCE           WORKER_THREAD_FLAGS = 8
	WT_TRANSFER_IMPERSONATION    WORKER_THREAD_FLAGS = 256
	WT_EXECUTEINTIMERTHREAD      WORKER_THREAD_FLAGS = 32
)

// enum
// flags
type CREATE_EVENT uint32

const (
	CREATE_EVENT_INITIAL_SET  CREATE_EVENT = 2
	CREATE_EVENT_MANUAL_RESET CREATE_EVENT = 1
)

// enum
type CREATE_PROCESS_LOGON_FLAGS uint32

const (
	LOGON_WITH_PROFILE        CREATE_PROCESS_LOGON_FLAGS = 1
	LOGON_NETCREDENTIALS_ONLY CREATE_PROCESS_LOGON_FLAGS = 2
)

// enum
type PROCESS_AFFINITY_AUTO_UPDATE_FLAGS uint32

const (
	PROCESS_AFFINITY_DISABLE_AUTO_UPDATE PROCESS_AFFINITY_AUTO_UPDATE_FLAGS = 0
	PROCESS_AFFINITY_ENABLE_AUTO_UPDATE  PROCESS_AFFINITY_AUTO_UPDATE_FLAGS = 1
)

// enum
// flags
type PROCESS_DEP_FLAGS uint32

const (
	PROCESS_DEP_ENABLE                      PROCESS_DEP_FLAGS = 1
	PROCESS_DEP_DISABLE_ATL_THUNK_EMULATION PROCESS_DEP_FLAGS = 2
	PROCESS_DEP_NONE                        PROCESS_DEP_FLAGS = 0
)

// enum
type PROCESS_NAME_FORMAT uint32

const (
	PROCESS_NAME_WIN32  PROCESS_NAME_FORMAT = 0
	PROCESS_NAME_NATIVE PROCESS_NAME_FORMAT = 1
)

// enum
type PROCESSOR_FEATURE_ID uint32

const (
	PF_ARM_64BIT_LOADSTORE_ATOMIC            PROCESSOR_FEATURE_ID = 25
	PF_ARM_DIVIDE_INSTRUCTION_AVAILABLE      PROCESSOR_FEATURE_ID = 24
	PF_ARM_EXTERNAL_CACHE_AVAILABLE          PROCESSOR_FEATURE_ID = 26
	PF_ARM_FMAC_INSTRUCTIONS_AVAILABLE       PROCESSOR_FEATURE_ID = 27
	PF_ARM_VFP_32_REGISTERS_AVAILABLE        PROCESSOR_FEATURE_ID = 18
	PF_3DNOW_INSTRUCTIONS_AVAILABLE          PROCESSOR_FEATURE_ID = 7
	PF_CHANNELS_ENABLED                      PROCESSOR_FEATURE_ID = 16
	PF_COMPARE_EXCHANGE_DOUBLE               PROCESSOR_FEATURE_ID = 2
	PF_COMPARE_EXCHANGE128                   PROCESSOR_FEATURE_ID = 14
	PF_COMPARE64_EXCHANGE128                 PROCESSOR_FEATURE_ID = 15
	PF_FASTFAIL_AVAILABLE                    PROCESSOR_FEATURE_ID = 23
	PF_FLOATING_POINT_EMULATED               PROCESSOR_FEATURE_ID = 1
	PF_FLOATING_POINT_PRECISION_ERRATA       PROCESSOR_FEATURE_ID = 0
	PF_MMX_INSTRUCTIONS_AVAILABLE            PROCESSOR_FEATURE_ID = 3
	PF_NX_ENABLED                            PROCESSOR_FEATURE_ID = 12
	PF_PAE_ENABLED                           PROCESSOR_FEATURE_ID = 9
	PF_RDTSC_INSTRUCTION_AVAILABLE           PROCESSOR_FEATURE_ID = 8
	PF_RDWRFSGSBASE_AVAILABLE                PROCESSOR_FEATURE_ID = 22
	PF_SECOND_LEVEL_ADDRESS_TRANSLATION      PROCESSOR_FEATURE_ID = 20
	PF_SSE3_INSTRUCTIONS_AVAILABLE           PROCESSOR_FEATURE_ID = 13
	PF_VIRT_FIRMWARE_ENABLED                 PROCESSOR_FEATURE_ID = 21
	PF_XMMI_INSTRUCTIONS_AVAILABLE           PROCESSOR_FEATURE_ID = 6
	PF_XMMI64_INSTRUCTIONS_AVAILABLE         PROCESSOR_FEATURE_ID = 10
	PF_XSAVE_ENABLED                         PROCESSOR_FEATURE_ID = 17
	PF_ARM_V8_INSTRUCTIONS_AVAILABLE         PROCESSOR_FEATURE_ID = 29
	PF_ARM_V8_CRYPTO_INSTRUCTIONS_AVAILABLE  PROCESSOR_FEATURE_ID = 30
	PF_ARM_V8_CRC32_INSTRUCTIONS_AVAILABLE   PROCESSOR_FEATURE_ID = 31
	PF_ARM_V81_ATOMIC_INSTRUCTIONS_AVAILABLE PROCESSOR_FEATURE_ID = 34
)

// enum
type GET_GUI_RESOURCES_FLAGS uint32

const (
	GR_GDIOBJECTS       GET_GUI_RESOURCES_FLAGS = 0
	GR_GDIOBJECTS_PEAK  GET_GUI_RESOURCES_FLAGS = 2
	GR_USEROBJECTS      GET_GUI_RESOURCES_FLAGS = 1
	GR_USEROBJECTS_PEAK GET_GUI_RESOURCES_FLAGS = 4
)

// enum
// flags
type STARTUPINFOW_FLAGS uint32

const (
	STARTF_FORCEONFEEDBACK  STARTUPINFOW_FLAGS = 64
	STARTF_FORCEOFFFEEDBACK STARTUPINFOW_FLAGS = 128
	STARTF_PREVENTPINNING   STARTUPINFOW_FLAGS = 8192
	STARTF_RUNFULLSCREEN    STARTUPINFOW_FLAGS = 32
	STARTF_TITLEISAPPID     STARTUPINFOW_FLAGS = 4096
	STARTF_TITLEISLINKNAME  STARTUPINFOW_FLAGS = 2048
	STARTF_UNTRUSTEDSOURCE  STARTUPINFOW_FLAGS = 32768
	STARTF_USECOUNTCHARS    STARTUPINFOW_FLAGS = 8
	STARTF_USEFILLATTRIBUTE STARTUPINFOW_FLAGS = 16
	STARTF_USEHOTKEY        STARTUPINFOW_FLAGS = 512
	STARTF_USEPOSITION      STARTUPINFOW_FLAGS = 4
	STARTF_USESHOWWINDOW    STARTUPINFOW_FLAGS = 1
	STARTF_USESIZE          STARTUPINFOW_FLAGS = 2
	STARTF_USESTDHANDLES    STARTUPINFOW_FLAGS = 256
)

// enum
type MEMORY_PRIORITY uint32

const (
	MEMORY_PRIORITY_VERY_LOW     MEMORY_PRIORITY = 1
	MEMORY_PRIORITY_LOW          MEMORY_PRIORITY = 2
	MEMORY_PRIORITY_MEDIUM       MEMORY_PRIORITY = 3
	MEMORY_PRIORITY_BELOW_NORMAL MEMORY_PRIORITY = 4
	MEMORY_PRIORITY_NORMAL       MEMORY_PRIORITY = 5
)

// enum
type PROCESS_PROTECTION_LEVEL uint32

const (
	PROTECTION_LEVEL_WINTCB_LIGHT      PROCESS_PROTECTION_LEVEL = 0
	PROTECTION_LEVEL_WINDOWS           PROCESS_PROTECTION_LEVEL = 1
	PROTECTION_LEVEL_WINDOWS_LIGHT     PROCESS_PROTECTION_LEVEL = 2
	PROTECTION_LEVEL_ANTIMALWARE_LIGHT PROCESS_PROTECTION_LEVEL = 3
	PROTECTION_LEVEL_LSA_LIGHT         PROCESS_PROTECTION_LEVEL = 4
	PROTECTION_LEVEL_WINTCB            PROCESS_PROTECTION_LEVEL = 5
	PROTECTION_LEVEL_CODEGEN_LIGHT     PROCESS_PROTECTION_LEVEL = 6
	PROTECTION_LEVEL_AUTHENTICODE      PROCESS_PROTECTION_LEVEL = 7
	PROTECTION_LEVEL_PPL_APP           PROCESS_PROTECTION_LEVEL = 8
	PROTECTION_LEVEL_NONE              PROCESS_PROTECTION_LEVEL = 4294967294
)

// enum
type POWER_REQUEST_CONTEXT_FLAGS uint32

const (
	POWER_REQUEST_CONTEXT_DETAILED_STRING POWER_REQUEST_CONTEXT_FLAGS = 2
	POWER_REQUEST_CONTEXT_SIMPLE_STRING   POWER_REQUEST_CONTEXT_FLAGS = 1
)

// enum
// flags
type THREAD_ACCESS_RIGHTS uint32

const (
	THREAD_TERMINATE                 THREAD_ACCESS_RIGHTS = 1
	THREAD_SUSPEND_RESUME            THREAD_ACCESS_RIGHTS = 2
	THREAD_GET_CONTEXT               THREAD_ACCESS_RIGHTS = 8
	THREAD_SET_CONTEXT               THREAD_ACCESS_RIGHTS = 16
	THREAD_SET_INFORMATION           THREAD_ACCESS_RIGHTS = 32
	THREAD_QUERY_INFORMATION         THREAD_ACCESS_RIGHTS = 64
	THREAD_SET_THREAD_TOKEN          THREAD_ACCESS_RIGHTS = 128
	THREAD_IMPERSONATE               THREAD_ACCESS_RIGHTS = 256
	THREAD_DIRECT_IMPERSONATION      THREAD_ACCESS_RIGHTS = 512
	THREAD_SET_LIMITED_INFORMATION   THREAD_ACCESS_RIGHTS = 1024
	THREAD_QUERY_LIMITED_INFORMATION THREAD_ACCESS_RIGHTS = 2048
	THREAD_RESUME                    THREAD_ACCESS_RIGHTS = 4096
	THREAD_ALL_ACCESS                THREAD_ACCESS_RIGHTS = 2097151
	THREAD_DELETE                    THREAD_ACCESS_RIGHTS = 65536
	THREAD_READ_CONTROL              THREAD_ACCESS_RIGHTS = 131072
	THREAD_WRITE_DAC                 THREAD_ACCESS_RIGHTS = 262144
	THREAD_WRITE_OWNER               THREAD_ACCESS_RIGHTS = 524288
	THREAD_SYNCHRONIZE               THREAD_ACCESS_RIGHTS = 1048576
	THREAD_STANDARD_RIGHTS_REQUIRED  THREAD_ACCESS_RIGHTS = 983040
)

// enum
// flags
type PROCESS_CREATION_FLAGS uint32

const (
	DEBUG_PROCESS                    PROCESS_CREATION_FLAGS = 1
	DEBUG_ONLY_THIS_PROCESS          PROCESS_CREATION_FLAGS = 2
	CREATE_SUSPENDED                 PROCESS_CREATION_FLAGS = 4
	DETACHED_PROCESS                 PROCESS_CREATION_FLAGS = 8
	CREATE_NEW_CONSOLE               PROCESS_CREATION_FLAGS = 16
	NORMAL_PRIORITY_CLASS            PROCESS_CREATION_FLAGS = 32
	IDLE_PRIORITY_CLASS              PROCESS_CREATION_FLAGS = 64
	HIGH_PRIORITY_CLASS              PROCESS_CREATION_FLAGS = 128
	REALTIME_PRIORITY_CLASS          PROCESS_CREATION_FLAGS = 256
	CREATE_NEW_PROCESS_GROUP         PROCESS_CREATION_FLAGS = 512
	CREATE_UNICODE_ENVIRONMENT       PROCESS_CREATION_FLAGS = 1024
	CREATE_SEPARATE_WOW_VDM          PROCESS_CREATION_FLAGS = 2048
	CREATE_SHARED_WOW_VDM            PROCESS_CREATION_FLAGS = 4096
	CREATE_FORCEDOS                  PROCESS_CREATION_FLAGS = 8192
	BELOW_NORMAL_PRIORITY_CLASS      PROCESS_CREATION_FLAGS = 16384
	ABOVE_NORMAL_PRIORITY_CLASS      PROCESS_CREATION_FLAGS = 32768
	INHERIT_PARENT_AFFINITY          PROCESS_CREATION_FLAGS = 65536
	INHERIT_CALLER_PRIORITY          PROCESS_CREATION_FLAGS = 131072
	CREATE_PROTECTED_PROCESS         PROCESS_CREATION_FLAGS = 262144
	EXTENDED_STARTUPINFO_PRESENT     PROCESS_CREATION_FLAGS = 524288
	PROCESS_MODE_BACKGROUND_BEGIN    PROCESS_CREATION_FLAGS = 1048576
	PROCESS_MODE_BACKGROUND_END      PROCESS_CREATION_FLAGS = 2097152
	CREATE_SECURE_PROCESS            PROCESS_CREATION_FLAGS = 4194304
	CREATE_BREAKAWAY_FROM_JOB        PROCESS_CREATION_FLAGS = 16777216
	CREATE_PRESERVE_CODE_AUTHZ_LEVEL PROCESS_CREATION_FLAGS = 33554432
	CREATE_DEFAULT_ERROR_MODE        PROCESS_CREATION_FLAGS = 67108864
	CREATE_NO_WINDOW                 PROCESS_CREATION_FLAGS = 134217728
	PROFILE_USER                     PROCESS_CREATION_FLAGS = 268435456
	PROFILE_KERNEL                   PROCESS_CREATION_FLAGS = 536870912
	PROFILE_SERVER                   PROCESS_CREATION_FLAGS = 1073741824
	CREATE_IGNORE_SYSTEM_DEFAULT     PROCESS_CREATION_FLAGS = 2147483648
)

// enum
// flags
type PROCESS_ACCESS_RIGHTS uint32

const (
	PROCESS_TERMINATE                 PROCESS_ACCESS_RIGHTS = 1
	PROCESS_CREATE_THREAD             PROCESS_ACCESS_RIGHTS = 2
	PROCESS_SET_SESSIONID             PROCESS_ACCESS_RIGHTS = 4
	PROCESS_VM_OPERATION              PROCESS_ACCESS_RIGHTS = 8
	PROCESS_VM_READ                   PROCESS_ACCESS_RIGHTS = 16
	PROCESS_VM_WRITE                  PROCESS_ACCESS_RIGHTS = 32
	PROCESS_DUP_HANDLE                PROCESS_ACCESS_RIGHTS = 64
	PROCESS_CREATE_PROCESS            PROCESS_ACCESS_RIGHTS = 128
	PROCESS_SET_QUOTA                 PROCESS_ACCESS_RIGHTS = 256
	PROCESS_SET_INFORMATION           PROCESS_ACCESS_RIGHTS = 512
	PROCESS_QUERY_INFORMATION         PROCESS_ACCESS_RIGHTS = 1024
	PROCESS_SUSPEND_RESUME            PROCESS_ACCESS_RIGHTS = 2048
	PROCESS_QUERY_LIMITED_INFORMATION PROCESS_ACCESS_RIGHTS = 4096
	PROCESS_SET_LIMITED_INFORMATION   PROCESS_ACCESS_RIGHTS = 8192
	PROCESS_ALL_ACCESS                PROCESS_ACCESS_RIGHTS = 2097151
	PROCESS_DELETE                    PROCESS_ACCESS_RIGHTS = 65536
	PROCESS_READ_CONTROL              PROCESS_ACCESS_RIGHTS = 131072
	PROCESS_WRITE_DAC                 PROCESS_ACCESS_RIGHTS = 262144
	PROCESS_WRITE_OWNER               PROCESS_ACCESS_RIGHTS = 524288
	PROCESS_SYNCHRONIZE               PROCESS_ACCESS_RIGHTS = 1048576
	PROCESS_STANDARD_RIGHTS_REQUIRED  PROCESS_ACCESS_RIGHTS = 983040
)

// enum
type QUEUE_USER_APC_FLAGS int32

const (
	QUEUE_USER_APC_FLAGS_NONE             QUEUE_USER_APC_FLAGS = 0
	QUEUE_USER_APC_FLAGS_SPECIAL_USER_APC QUEUE_USER_APC_FLAGS = 1
)

// enum
type THREAD_INFORMATION_CLASS int32

const (
	ThreadMemoryPriority      THREAD_INFORMATION_CLASS = 0
	ThreadAbsoluteCpuPriority THREAD_INFORMATION_CLASS = 1
	ThreadDynamicCodePolicy   THREAD_INFORMATION_CLASS = 2
	ThreadPowerThrottling     THREAD_INFORMATION_CLASS = 3
	ThreadInformationClassMax THREAD_INFORMATION_CLASS = 4
)

// enum
type PROCESS_INFORMATION_CLASS int32

const (
	ProcessMemoryPriority        PROCESS_INFORMATION_CLASS = 0
	ProcessMemoryExhaustionInfo  PROCESS_INFORMATION_CLASS = 1
	ProcessAppMemoryInfo         PROCESS_INFORMATION_CLASS = 2
	ProcessInPrivateInfo         PROCESS_INFORMATION_CLASS = 3
	ProcessPowerThrottling       PROCESS_INFORMATION_CLASS = 4
	ProcessReservedValue1        PROCESS_INFORMATION_CLASS = 5
	ProcessTelemetryCoverageInfo PROCESS_INFORMATION_CLASS = 6
	ProcessProtectionLevelInfo   PROCESS_INFORMATION_CLASS = 7
	ProcessLeapSecondInfo        PROCESS_INFORMATION_CLASS = 8
	ProcessMachineTypeInfo       PROCESS_INFORMATION_CLASS = 9
	ProcessInformationClassMax   PROCESS_INFORMATION_CLASS = 10
)

// enum
// flags
type MACHINE_ATTRIBUTES uint32

const (
	UserEnabled    MACHINE_ATTRIBUTES = 1
	KernelEnabled  MACHINE_ATTRIBUTES = 2
	Wow64Container MACHINE_ATTRIBUTES = 4
)

// enum
type PROCESS_MEMORY_EXHAUSTION_TYPE int32

const (
	PMETypeFailFastOnCommitFailure PROCESS_MEMORY_EXHAUSTION_TYPE = 0
	PMETypeMax                     PROCESS_MEMORY_EXHAUSTION_TYPE = 1
)

// enum
type AVRT_PRIORITY int32

const (
	AVRT_PRIORITY_VERYLOW  AVRT_PRIORITY = -2
	AVRT_PRIORITY_LOW      AVRT_PRIORITY = -1
	AVRT_PRIORITY_NORMAL   AVRT_PRIORITY = 0
	AVRT_PRIORITY_HIGH     AVRT_PRIORITY = 1
	AVRT_PRIORITY_CRITICAL AVRT_PRIORITY = 2
)

// enum
type PROCESS_MITIGATION_POLICY int32

const (
	ProcessDEPPolicy                   PROCESS_MITIGATION_POLICY = 0
	ProcessASLRPolicy                  PROCESS_MITIGATION_POLICY = 1
	ProcessDynamicCodePolicy           PROCESS_MITIGATION_POLICY = 2
	ProcessStrictHandleCheckPolicy     PROCESS_MITIGATION_POLICY = 3
	ProcessSystemCallDisablePolicy     PROCESS_MITIGATION_POLICY = 4
	ProcessMitigationOptionsMask       PROCESS_MITIGATION_POLICY = 5
	ProcessExtensionPointDisablePolicy PROCESS_MITIGATION_POLICY = 6
	ProcessControlFlowGuardPolicy      PROCESS_MITIGATION_POLICY = 7
	ProcessSignaturePolicy             PROCESS_MITIGATION_POLICY = 8
	ProcessFontDisablePolicy           PROCESS_MITIGATION_POLICY = 9
	ProcessImageLoadPolicy             PROCESS_MITIGATION_POLICY = 10
	ProcessSystemCallFilterPolicy      PROCESS_MITIGATION_POLICY = 11
	ProcessPayloadRestrictionPolicy    PROCESS_MITIGATION_POLICY = 12
	ProcessChildProcessPolicy          PROCESS_MITIGATION_POLICY = 13
	ProcessSideChannelIsolationPolicy  PROCESS_MITIGATION_POLICY = 14
	ProcessUserShadowStackPolicy       PROCESS_MITIGATION_POLICY = 15
	ProcessRedirectionTrustPolicy      PROCESS_MITIGATION_POLICY = 16
	MaxProcessMitigationPolicy         PROCESS_MITIGATION_POLICY = 17
)

// enum
type RTL_UMS_THREAD_INFO_CLASS int32

const (
	UmsThreadInvalidInfoClass RTL_UMS_THREAD_INFO_CLASS = 0
	UmsThreadUserContext      RTL_UMS_THREAD_INFO_CLASS = 1
	UmsThreadPriority         RTL_UMS_THREAD_INFO_CLASS = 2
	UmsThreadAffinity         RTL_UMS_THREAD_INFO_CLASS = 3
	UmsThreadTeb              RTL_UMS_THREAD_INFO_CLASS = 4
	UmsThreadIsSuspended      RTL_UMS_THREAD_INFO_CLASS = 5
	UmsThreadIsTerminated     RTL_UMS_THREAD_INFO_CLASS = 6
	UmsThreadMaxInfoClass     RTL_UMS_THREAD_INFO_CLASS = 7
)

// enum
type TP_CALLBACK_PRIORITY int32

const (
	TP_CALLBACK_PRIORITY_HIGH    TP_CALLBACK_PRIORITY = 0
	TP_CALLBACK_PRIORITY_NORMAL  TP_CALLBACK_PRIORITY = 1
	TP_CALLBACK_PRIORITY_LOW     TP_CALLBACK_PRIORITY = 2
	TP_CALLBACK_PRIORITY_INVALID TP_CALLBACK_PRIORITY = 3
	TP_CALLBACK_PRIORITY_COUNT   TP_CALLBACK_PRIORITY = 3
)

// enum
type PROC_THREAD_ATTRIBUTE_NUM uint32

const (
	ProcThreadAttributeParentProcess                PROC_THREAD_ATTRIBUTE_NUM = 0
	ProcThreadAttributeHandleList                   PROC_THREAD_ATTRIBUTE_NUM = 2
	ProcThreadAttributeGroupAffinity                PROC_THREAD_ATTRIBUTE_NUM = 3
	ProcThreadAttributePreferredNode                PROC_THREAD_ATTRIBUTE_NUM = 4
	ProcThreadAttributeIdealProcessor               PROC_THREAD_ATTRIBUTE_NUM = 5
	ProcThreadAttributeUmsThread                    PROC_THREAD_ATTRIBUTE_NUM = 6
	ProcThreadAttributeMitigationPolicy             PROC_THREAD_ATTRIBUTE_NUM = 7
	ProcThreadAttributeSecurityCapabilities         PROC_THREAD_ATTRIBUTE_NUM = 9
	ProcThreadAttributeProtectionLevel              PROC_THREAD_ATTRIBUTE_NUM = 11
	ProcThreadAttributeJobList                      PROC_THREAD_ATTRIBUTE_NUM = 13
	ProcThreadAttributeChildProcessPolicy           PROC_THREAD_ATTRIBUTE_NUM = 14
	ProcThreadAttributeAllApplicationPackagesPolicy PROC_THREAD_ATTRIBUTE_NUM = 15
	ProcThreadAttributeWin32kFilter                 PROC_THREAD_ATTRIBUTE_NUM = 16
	ProcThreadAttributeSafeOpenPromptOriginClaim    PROC_THREAD_ATTRIBUTE_NUM = 17
	ProcThreadAttributeDesktopAppPolicy             PROC_THREAD_ATTRIBUTE_NUM = 18
	ProcThreadAttributePseudoConsole                PROC_THREAD_ATTRIBUTE_NUM = 22
	ProcThreadAttributeMitigationAuditPolicy        PROC_THREAD_ATTRIBUTE_NUM = 24
	ProcThreadAttributeMachineType                  PROC_THREAD_ATTRIBUTE_NUM = 25
	ProcThreadAttributeComponentFilter              PROC_THREAD_ATTRIBUTE_NUM = 26
	ProcThreadAttributeEnableOptionalXStateFeatures PROC_THREAD_ATTRIBUTE_NUM = 27
)

// enum
type PROCESSINFOCLASS int32

const (
	ProcessBasicInformation   PROCESSINFOCLASS = 0
	ProcessDebugPort          PROCESSINFOCLASS = 7
	ProcessWow64Information   PROCESSINFOCLASS = 26
	ProcessImageFileName      PROCESSINFOCLASS = 27
	ProcessBreakOnTermination PROCESSINFOCLASS = 29
)

// enum
type THREADINFOCLASS int32

const (
	ThreadIsIoPending     THREADINFOCLASS = 16
	ThreadNameInformation THREADINFOCLASS = 38
)

// structs

type TP_CALLBACK_INSTANCE struct {
}

type TP_WORK struct {
}

type TP_TIMER struct {
}

type TP_WAIT struct {
}

type TP_IO struct {
}

type REASON_CONTEXT_Reason_Detailed struct {
	LocalizedReasonModule HINSTANCE
	LocalizedReasonId     uint32
	ReasonStringCount     uint32
	ReasonStrings         *PWSTR
}

type REASON_CONTEXT_Reason struct {
	Data [3]uint64
}

func (this *REASON_CONTEXT_Reason) Detailed() *REASON_CONTEXT_Reason_Detailed {
	return (*REASON_CONTEXT_Reason_Detailed)(unsafe.Pointer(this))
}

func (this *REASON_CONTEXT_Reason) DetailedVal() REASON_CONTEXT_Reason_Detailed {
	return *(*REASON_CONTEXT_Reason_Detailed)(unsafe.Pointer(this))
}

func (this *REASON_CONTEXT_Reason) SimpleReasonString() *PWSTR {
	return (*PWSTR)(unsafe.Pointer(this))
}

func (this *REASON_CONTEXT_Reason) SimpleReasonStringVal() PWSTR {
	return *(*PWSTR)(unsafe.Pointer(this))
}

type REASON_CONTEXT struct {
	Version uint32
	Flags   POWER_REQUEST_CONTEXT_FLAGS
	Reason  REASON_CONTEXT_Reason
}

type PROCESS_INFORMATION struct {
	HProcess    HANDLE
	HThread     HANDLE
	DwProcessId uint32
	DwThreadId  uint32
}

type STARTUPINFOA struct {
	Cb              uint32
	LpReserved      PSTR
	LpDesktop       PSTR
	LpTitle         PSTR
	DwX             uint32
	DwY             uint32
	DwXSize         uint32
	DwYSize         uint32
	DwXCountChars   uint32
	DwYCountChars   uint32
	DwFillAttribute uint32
	DwFlags         STARTUPINFOW_FLAGS
	WShowWindow     uint16
	CbReserved2     uint16
	LpReserved2     *byte
	HStdInput       HANDLE
	HStdOutput      HANDLE
	HStdError       HANDLE
}

type STARTUPINFO = STARTUPINFOW
type STARTUPINFOW struct {
	Cb              uint32
	LpReserved      PWSTR
	LpDesktop       PWSTR
	LpTitle         PWSTR
	DwX             uint32
	DwY             uint32
	DwXSize         uint32
	DwYSize         uint32
	DwXCountChars   uint32
	DwYCountChars   uint32
	DwFillAttribute uint32
	DwFlags         STARTUPINFOW_FLAGS
	WShowWindow     uint16
	CbReserved2     uint16
	LpReserved2     *byte
	HStdInput       HANDLE
	HStdOutput      HANDLE
	HStdError       HANDLE
}

type MEMORY_PRIORITY_INFORMATION struct {
	MemoryPriority MEMORY_PRIORITY
}

type THREAD_POWER_THROTTLING_STATE struct {
	Version     uint32
	ControlMask uint32
	StateMask   uint32
}

type APP_MEMORY_INFORMATION struct {
	AvailableCommit        uint64
	PrivateCommitUsage     uint64
	PeakPrivateCommitUsage uint64
	TotalCommitUsage       uint64
}

type PROCESS_MACHINE_INFORMATION struct {
	ProcessMachine    uint16
	Res0              uint16
	MachineAttributes MACHINE_ATTRIBUTES
}

type PROCESS_MEMORY_EXHAUSTION_INFO struct {
	Version  uint16
	Reserved uint16
	Type     PROCESS_MEMORY_EXHAUSTION_TYPE
	Value    uintptr
}

type PROCESS_POWER_THROTTLING_STATE struct {
	Version     uint32
	ControlMask uint32
	StateMask   uint32
}

type PROCESS_PROTECTION_LEVEL_INFORMATION struct {
	ProtectionLevel PROCESS_PROTECTION_LEVEL
}

type PROCESS_LEAP_SECOND_INFO struct {
	Flags    uint32
	Reserved uint32
}

type PROCESS_DYNAMIC_EH_CONTINUATION_TARGET struct {
	TargetAddress uintptr
	Flags         uintptr
}

type PROCESS_DYNAMIC_EH_CONTINUATION_TARGETS_INFORMATION struct {
	NumberOfTargets uint16
	Reserved        uint16
	Reserved2       uint32
	Targets         *PROCESS_DYNAMIC_EH_CONTINUATION_TARGET
}

type PROCESS_DYNAMIC_ENFORCED_ADDRESS_RANGE struct {
	BaseAddress uintptr
	Size        uintptr
	Flags       uint32
}

type PROCESS_DYNAMIC_ENFORCED_ADDRESS_RANGES_INFORMATION struct {
	NumberOfRanges uint16
	Reserved       uint16
	Reserved2      uint32
	Ranges         *PROCESS_DYNAMIC_ENFORCED_ADDRESS_RANGE
}

type IO_COUNTERS struct {
	ReadOperationCount  uint64
	WriteOperationCount uint64
	OtherOperationCount uint64
	ReadTransferCount   uint64
	WriteTransferCount  uint64
	OtherTransferCount  uint64
}

type RTL_RUN_ONCE struct {
	Data [1]uint64
}

func (this *RTL_RUN_ONCE) Ptr() *unsafe.Pointer {
	return (*unsafe.Pointer)(unsafe.Pointer(this))
}

func (this *RTL_RUN_ONCE) PtrVal() unsafe.Pointer {
	return *(*unsafe.Pointer)(unsafe.Pointer(this))
}

type RTL_BARRIER struct {
	Reserved1 uint32
	Reserved2 uint32
	Reserved3 [2]uintptr
	Reserved4 uint32
	Reserved5 uint32
}

type RTL_CRITICAL_SECTION_DEBUG struct {
	Type                      uint16
	CreatorBackTraceIndex     uint16
	CriticalSection           *RTL_CRITICAL_SECTION
	ProcessLocksList          LIST_ENTRY
	EntryCount                uint32
	ContentionCount           uint32
	Flags                     uint32
	CreatorBackTraceIndexHigh uint16
	SpareWORD                 uint16
}

type RTL_CRITICAL_SECTION struct {
	DebugInfo      *RTL_CRITICAL_SECTION_DEBUG
	LockCount      int32
	RecursionCount int32
	OwningThread   HANDLE
	LockSemaphore  HANDLE
	SpinCount      uintptr
}

type RTL_SRWLOCK struct {
	Ptr unsafe.Pointer
}

type RTL_CONDITION_VARIABLE struct {
	Ptr unsafe.Pointer
}

type TP_POOL_STACK_INFORMATION struct {
	StackReserve uintptr
	StackCommit  uintptr
}

type TP_CALLBACK_ENVIRON_V3_ACTIVATION_CONTEXT struct {
}

type TP_CALLBACK_ENVIRON_V3_U_S struct {
	Bitfield_ uint32
}

type TP_CALLBACK_ENVIRON_V3_U struct {
	Data [1]uint32
}

func (this *TP_CALLBACK_ENVIRON_V3_U) Flags() *uint32 {
	return (*uint32)(unsafe.Pointer(this))
}

func (this *TP_CALLBACK_ENVIRON_V3_U) FlagsVal() uint32 {
	return *(*uint32)(unsafe.Pointer(this))
}

func (this *TP_CALLBACK_ENVIRON_V3_U) S() *TP_CALLBACK_ENVIRON_V3_U_S {
	return (*TP_CALLBACK_ENVIRON_V3_U_S)(unsafe.Pointer(this))
}

func (this *TP_CALLBACK_ENVIRON_V3_U) SVal() TP_CALLBACK_ENVIRON_V3_U_S {
	return *(*TP_CALLBACK_ENVIRON_V3_U_S)(unsafe.Pointer(this))
}

type TP_CALLBACK_ENVIRON_V3 struct {
	Version                    uint32
	Pool                       PTP_POOL
	CleanupGroup               uintptr
	CleanupGroupCancelCallback PTP_CLEANUP_GROUP_CANCEL_CALLBACK
	RaceDll                    unsafe.Pointer
	ActivationContext          uintptr
	FinalizationCallback       PTP_SIMPLE_CALLBACK
	U                          TP_CALLBACK_ENVIRON_V3_U
	CallbackPriority           TP_CALLBACK_PRIORITY
	Size                       uint32
}

type UMS_SCHEDULER_STARTUP_INFO struct {
	UmsVersion     uint32
	CompletionList unsafe.Pointer
	SchedulerProc  PRTL_UMS_SCHEDULER_ENTRY_POINT
	SchedulerParam unsafe.Pointer
}

type UMS_SYSTEM_THREAD_INFORMATION_Anonymous_Anonymous struct {
	Bitfield_ uint32
}

type UMS_SYSTEM_THREAD_INFORMATION_Anonymous struct {
	UMS_SYSTEM_THREAD_INFORMATION_Anonymous_Anonymous
}

func (this *UMS_SYSTEM_THREAD_INFORMATION_Anonymous) Anonymous() *UMS_SYSTEM_THREAD_INFORMATION_Anonymous_Anonymous {
	return (*UMS_SYSTEM_THREAD_INFORMATION_Anonymous_Anonymous)(unsafe.Pointer(this))
}

func (this *UMS_SYSTEM_THREAD_INFORMATION_Anonymous) AnonymousVal() UMS_SYSTEM_THREAD_INFORMATION_Anonymous_Anonymous {
	return *(*UMS_SYSTEM_THREAD_INFORMATION_Anonymous_Anonymous)(unsafe.Pointer(this))
}

func (this *UMS_SYSTEM_THREAD_INFORMATION_Anonymous) ThreadUmsFlags() *uint32 {
	return (*uint32)(unsafe.Pointer(this))
}

func (this *UMS_SYSTEM_THREAD_INFORMATION_Anonymous) ThreadUmsFlagsVal() uint32 {
	return *(*uint32)(unsafe.Pointer(this))
}

type UMS_SYSTEM_THREAD_INFORMATION struct {
	UmsVersion uint32
	UMS_SYSTEM_THREAD_INFORMATION_Anonymous
}

type STARTUPINFOEXA struct {
	StartupInfo     STARTUPINFOA
	LpAttributeList LPPROC_THREAD_ATTRIBUTE_LIST
}

type STARTUPINFOEX = STARTUPINFOEXW
type STARTUPINFOEXW struct {
	StartupInfo     STARTUPINFOW
	LpAttributeList LPPROC_THREAD_ATTRIBUTE_LIST
}

type PEB_LDR_DATA struct {
	Reserved1               [8]byte
	Reserved2               [3]unsafe.Pointer
	InMemoryOrderModuleList LIST_ENTRY
}

type RTL_USER_PROCESS_PARAMETERS struct {
	Reserved1     [16]byte
	Reserved2     [10]unsafe.Pointer
	ImagePathName UNICODE_STRING
	CommandLine   UNICODE_STRING
}

type PEB struct {
	Reserved1              [2]byte
	BeingDebugged          byte
	Reserved2              [1]byte
	Reserved3              [2]unsafe.Pointer
	Ldr                    *PEB_LDR_DATA
	ProcessParameters      *RTL_USER_PROCESS_PARAMETERS
	Reserved4              [3]unsafe.Pointer
	AtlThunkSListPtr       unsafe.Pointer
	Reserved5              unsafe.Pointer
	Reserved6              uint32
	Reserved7              unsafe.Pointer
	Reserved8              uint32
	AtlThunkSListPtr32     uint32
	Reserved9              [45]unsafe.Pointer
	Reserved10             [96]byte
	PostProcessInitRoutine PPS_POST_PROCESS_INIT_ROUTINE
	Reserved11             [128]byte
	Reserved12             [1]unsafe.Pointer
	SessionId              uint32
}

type PROCESS_BASIC_INFORMATION struct {
	Reserved1       unsafe.Pointer
	PebBaseAddress  *PEB
	Reserved2       [2]unsafe.Pointer
	UniqueProcessId uintptr
	Reserved3       unsafe.Pointer
}

// func types

type LPTHREAD_START_ROUTINE = uintptr
type LPTHREAD_START_ROUTINE_func = func(lpThreadParameter unsafe.Pointer) uint32

type PINIT_ONCE_FN = uintptr
type PINIT_ONCE_FN_func = func(InitOnce *RTL_RUN_ONCE, Parameter unsafe.Pointer, Context unsafe.Pointer) BOOL

type PTIMERAPCROUTINE = uintptr
type PTIMERAPCROUTINE_func = func(lpArgToCompletionRoutine unsafe.Pointer, dwTimerLowValue uint32, dwTimerHighValue uint32)

type PTP_WIN32_IO_CALLBACK = uintptr
type PTP_WIN32_IO_CALLBACK_func = func(Instance *TP_CALLBACK_INSTANCE, Context unsafe.Pointer, Overlapped unsafe.Pointer, IoResult uint32, NumberOfBytesTransferred uintptr, Io *TP_IO)

type PRTL_UMS_SCHEDULER_ENTRY_POINT = uintptr
type PRTL_UMS_SCHEDULER_ENTRY_POINT_func = func(Reason RTL_UMS_SCHEDULER_REASON, ActivationPayload uintptr, SchedulerParam unsafe.Pointer)

type WAITORTIMERCALLBACK = uintptr
type WAITORTIMERCALLBACK_func = func(param0 unsafe.Pointer, param1 BOOLEAN)

type PFLS_CALLBACK_FUNCTION = uintptr
type PFLS_CALLBACK_FUNCTION_func = func(lpFlsData unsafe.Pointer)

type PTP_SIMPLE_CALLBACK = uintptr
type PTP_SIMPLE_CALLBACK_func = func(Instance *TP_CALLBACK_INSTANCE, Context unsafe.Pointer)

type PTP_CLEANUP_GROUP_CANCEL_CALLBACK = uintptr
type PTP_CLEANUP_GROUP_CANCEL_CALLBACK_func = func(ObjectContext unsafe.Pointer, CleanupContext unsafe.Pointer)

type PTP_WORK_CALLBACK = uintptr
type PTP_WORK_CALLBACK_func = func(Instance *TP_CALLBACK_INSTANCE, Context unsafe.Pointer, Work *TP_WORK)

type PTP_TIMER_CALLBACK = uintptr
type PTP_TIMER_CALLBACK_func = func(Instance *TP_CALLBACK_INSTANCE, Context unsafe.Pointer, Timer *TP_TIMER)

type PTP_WAIT_CALLBACK = uintptr
type PTP_WAIT_CALLBACK_func = func(Instance *TP_CALLBACK_INSTANCE, Context unsafe.Pointer, Wait *TP_WAIT, WaitResult uint32)

type LPFIBER_START_ROUTINE = uintptr
type LPFIBER_START_ROUTINE_func = func(lpFiberParameter unsafe.Pointer)

type PPS_POST_PROCESS_INIT_ROUTINE = uintptr
type PPS_POST_PROCESS_INIT_ROUTINE_func = func()

var (
	pGetProcessWorkingSetSize                     uintptr
	pSetProcessWorkingSetSize                     uintptr
	pFlsAlloc                                     uintptr
	pFlsGetValue                                  uintptr
	pFlsSetValue                                  uintptr
	pFlsFree                                      uintptr
	pIsThreadAFiber                               uintptr
	pInitializeSRWLock                            uintptr
	pReleaseSRWLockExclusive                      uintptr
	pReleaseSRWLockShared                         uintptr
	pAcquireSRWLockExclusive                      uintptr
	pAcquireSRWLockShared                         uintptr
	pTryAcquireSRWLockExclusive                   uintptr
	pTryAcquireSRWLockShared                      uintptr
	pInitializeCriticalSection                    uintptr
	pEnterCriticalSection                         uintptr
	pLeaveCriticalSection                         uintptr
	pInitializeCriticalSectionAndSpinCount        uintptr
	pInitializeCriticalSectionEx                  uintptr
	pSetCriticalSectionSpinCount                  uintptr
	pTryEnterCriticalSection                      uintptr
	pDeleteCriticalSection                        uintptr
	pInitOnceInitialize                           uintptr
	pInitOnceExecuteOnce                          uintptr
	pInitOnceBeginInitialize                      uintptr
	pInitOnceComplete                             uintptr
	pInitializeConditionVariable                  uintptr
	pWakeConditionVariable                        uintptr
	pWakeAllConditionVariable                     uintptr
	pSleepConditionVariableCS                     uintptr
	pSleepConditionVariableSRW                    uintptr
	pSetEvent                                     uintptr
	pResetEvent                                   uintptr
	pReleaseSemaphore                             uintptr
	pReleaseMutex                                 uintptr
	pWaitForSingleObject                          uintptr
	pSleepEx                                      uintptr
	pWaitForSingleObjectEx                        uintptr
	pWaitForMultipleObjectsEx                     uintptr
	pCreateMutexA                                 uintptr
	pCreateMutexW                                 uintptr
	pOpenMutexW                                   uintptr
	pCreateEventA                                 uintptr
	pCreateEventW                                 uintptr
	pOpenEventA                                   uintptr
	pOpenEventW                                   uintptr
	pOpenSemaphoreW                               uintptr
	pOpenWaitableTimerW                           uintptr
	pSetWaitableTimerEx                           uintptr
	pSetWaitableTimer                             uintptr
	pCancelWaitableTimer                          uintptr
	pCreateMutexExA                               uintptr
	pCreateMutexExW                               uintptr
	pCreateEventExA                               uintptr
	pCreateEventExW                               uintptr
	pCreateSemaphoreExW                           uintptr
	pCreateWaitableTimerExW                       uintptr
	pEnterSynchronizationBarrier                  uintptr
	pInitializeSynchronizationBarrier             uintptr
	pDeleteSynchronizationBarrier                 uintptr
	pSleep                                        uintptr
	pWaitForMultipleObjects                       uintptr
	pCreateSemaphoreW                             uintptr
	pCreateWaitableTimerW                         uintptr
	pInitializeSListHead                          uintptr
	pInterlockedPopEntrySList                     uintptr
	pInterlockedPushEntrySList                    uintptr
	pInterlockedPushListSListEx                   uintptr
	pInterlockedFlushSList                        uintptr
	pQueryDepthSList                              uintptr
	pQueueUserAPC                                 uintptr
	pQueueUserAPC2                                uintptr
	pGetProcessTimes                              uintptr
	pGetCurrentProcess                            uintptr
	pGetCurrentProcessId                          uintptr
	pExitProcess                                  uintptr
	pTerminateProcess                             uintptr
	pGetExitCodeProcess                           uintptr
	pSwitchToThread                               uintptr
	pCreateThread                                 uintptr
	pCreateRemoteThread                           uintptr
	pGetCurrentThread                             uintptr
	pGetCurrentThreadId                           uintptr
	pOpenThread                                   uintptr
	pSetThreadPriority                            uintptr
	pSetThreadPriorityBoost                       uintptr
	pGetThreadPriorityBoost                       uintptr
	pGetThreadPriority                            uintptr
	pExitThread                                   uintptr
	pTerminateThread                              uintptr
	pGetExitCodeThread                            uintptr
	pSuspendThread                                uintptr
	pResumeThread                                 uintptr
	pTlsAlloc                                     uintptr
	pTlsGetValue                                  uintptr
	pTlsSetValue                                  uintptr
	pTlsFree                                      uintptr
	pCreateProcessA                               uintptr
	pCreateProcessW                               uintptr
	pSetProcessShutdownParameters                 uintptr
	pGetProcessVersion                            uintptr
	pGetStartupInfoW                              uintptr
	pCreateProcessAsUserW                         uintptr
	pSetThreadToken                               uintptr
	pOpenProcessToken                             uintptr
	pOpenThreadToken                              uintptr
	pSetPriorityClass                             uintptr
	pGetPriorityClass                             uintptr
	pSetThreadStackGuarantee                      uintptr
	pGetProcessId                                 uintptr
	pGetThreadId                                  uintptr
	pFlushProcessWriteBuffers                     uintptr
	pGetProcessIdOfThread                         uintptr
	pInitializeProcThreadAttributeList            uintptr
	pDeleteProcThreadAttributeList                uintptr
	pUpdateProcThreadAttribute                    uintptr
	pSetProcessDynamicEHContinuationTargets       uintptr
	pSetProcessDynamicEnforcedCetCompatibleRanges uintptr
	pSetProcessAffinityUpdateMode                 uintptr
	pQueryProcessAffinityUpdateMode               uintptr
	pCreateRemoteThreadEx                         uintptr
	pGetCurrentThreadStackLimits                  uintptr
	pGetProcessMitigationPolicy                   uintptr
	pSetProcessMitigationPolicy                   uintptr
	pGetThreadTimes                               uintptr
	pOpenProcess                                  uintptr
	pIsProcessorFeaturePresent                    uintptr
	pGetProcessHandleCount                        uintptr
	pGetCurrentProcessorNumber                    uintptr
	pSetThreadIdealProcessorEx                    uintptr
	pGetThreadIdealProcessorEx                    uintptr
	pGetCurrentProcessorNumberEx                  uintptr
	pGetProcessPriorityBoost                      uintptr
	pSetProcessPriorityBoost                      uintptr
	pGetThreadIOPendingFlag                       uintptr
	pGetSystemTimes                               uintptr
	pGetThreadInformation                         uintptr
	pSetThreadInformation                         uintptr
	pIsProcessCritical                            uintptr
	pSetProtectedPolicy                           uintptr
	pQueryProtectedPolicy                         uintptr
	pSetThreadIdealProcessor                      uintptr
	pSetProcessInformation                        uintptr
	pGetProcessInformation                        uintptr
	pGetProcessDefaultCpuSets                     uintptr
	pSetProcessDefaultCpuSets                     uintptr
	pGetThreadSelectedCpuSets                     uintptr
	pSetThreadSelectedCpuSets                     uintptr
	pCreateProcessAsUserA                         uintptr
	pGetProcessShutdownParameters                 uintptr
	pGetProcessDefaultCpuSetMasks                 uintptr
	pSetProcessDefaultCpuSetMasks                 uintptr
	pGetThreadSelectedCpuSetMasks                 uintptr
	pSetThreadSelectedCpuSetMasks                 uintptr
	pGetMachineTypeAttributes                     uintptr
	pSetThreadDescription                         uintptr
	pGetThreadDescription                         uintptr
	pQueueUserWorkItem                            uintptr
	pUnregisterWaitEx                             uintptr
	pCreateTimerQueue                             uintptr
	pCreateTimerQueueTimer                        uintptr
	pChangeTimerQueueTimer                        uintptr
	pDeleteTimerQueueTimer                        uintptr
	pDeleteTimerQueue                             uintptr
	pDeleteTimerQueueEx                           uintptr
	pCreateThreadpool                             uintptr
	pSetThreadpoolThreadMaximum                   uintptr
	pSetThreadpoolThreadMinimum                   uintptr
	pSetThreadpoolStackInformation                uintptr
	pQueryThreadpoolStackInformation              uintptr
	pCloseThreadpool                              uintptr
	pCreateThreadpoolCleanupGroup                 uintptr
	pCloseThreadpoolCleanupGroupMembers           uintptr
	pCloseThreadpoolCleanupGroup                  uintptr
	pSetEventWhenCallbackReturns                  uintptr
	pReleaseSemaphoreWhenCallbackReturns          uintptr
	pReleaseMutexWhenCallbackReturns              uintptr
	pLeaveCriticalSectionWhenCallbackReturns      uintptr
	pFreeLibraryWhenCallbackReturns               uintptr
	pCallbackMayRunLong                           uintptr
	pDisassociateCurrentThreadFromCallback        uintptr
	pTrySubmitThreadpoolCallback                  uintptr
	pCreateThreadpoolWork                         uintptr
	pSubmitThreadpoolWork                         uintptr
	pWaitForThreadpoolWorkCallbacks               uintptr
	pCloseThreadpoolWork                          uintptr
	pCreateThreadpoolTimer                        uintptr
	pSetThreadpoolTimer                           uintptr
	pIsThreadpoolTimerSet                         uintptr
	pWaitForThreadpoolTimerCallbacks              uintptr
	pCloseThreadpoolTimer                         uintptr
	pCreateThreadpoolWait                         uintptr
	pSetThreadpoolWait                            uintptr
	pWaitForThreadpoolWaitCallbacks               uintptr
	pCloseThreadpoolWait                          uintptr
	pCreateThreadpoolIo                           uintptr
	pStartThreadpoolIo                            uintptr
	pCancelThreadpoolIo                           uintptr
	pWaitForThreadpoolIoCallbacks                 uintptr
	pCloseThreadpoolIo                            uintptr
	pSetThreadpoolTimerEx                         uintptr
	pSetThreadpoolWaitEx                          uintptr
	pIsWow64Process                               uintptr
	pIsWow64Process2                              uintptr
	pWow64SuspendThread                           uintptr
	pCreatePrivateNamespaceW                      uintptr
	pOpenPrivateNamespaceW                        uintptr
	pClosePrivateNamespace                        uintptr
	pCreateBoundaryDescriptorW                    uintptr
	pAddSIDToBoundaryDescriptor                   uintptr
	pDeleteBoundaryDescriptor                     uintptr
	pGetNumaHighestNodeNumber                     uintptr
	pGetNumaNodeProcessorMaskEx                   uintptr
	pGetNumaNodeProcessorMask2                    uintptr
	pGetNumaProximityNodeEx                       uintptr
	pGetProcessGroupAffinity                      uintptr
	pGetThreadGroupAffinity                       uintptr
	pSetThreadGroupAffinity                       uintptr
	pAttachThreadInput                            uintptr
	pWaitForInputIdle                             uintptr
	pGetGuiResources                              uintptr
	pIsImmersiveProcess                           uintptr
	pSetProcessRestrictionExemption               uintptr
	pGetProcessAffinityMask                       uintptr
	pSetProcessAffinityMask                       uintptr
	pGetProcessIoCounters                         uintptr
	pSwitchToFiber                                uintptr
	pDeleteFiber                                  uintptr
	pConvertFiberToThread                         uintptr
	pCreateFiberEx                                uintptr
	pConvertThreadToFiberEx                       uintptr
	pCreateFiber                                  uintptr
	pConvertThreadToFiber                         uintptr
	pCreateUmsCompletionList                      uintptr
	pDequeueUmsCompletionListItems                uintptr
	pGetUmsCompletionListEvent                    uintptr
	pExecuteUmsThread                             uintptr
	pUmsThreadYield                               uintptr
	pDeleteUmsCompletionList                      uintptr
	pGetCurrentUmsThread                          uintptr
	pGetNextUmsListItem                           uintptr
	pQueryUmsThreadInformation                    uintptr
	pSetUmsThreadInformation                      uintptr
	pDeleteUmsThreadContext                       uintptr
	pCreateUmsThreadContext                       uintptr
	pEnterUmsSchedulingMode                       uintptr
	pGetUmsSystemThreadInformation                uintptr
	pSetThreadAffinityMask                        uintptr
	pSetProcessDEPPolicy                          uintptr
	pGetProcessDEPPolicy                          uintptr
	pPulseEvent                                   uintptr
	pWinExec                                      uintptr
	pCreateSemaphoreA                             uintptr
	pCreateSemaphoreExA                           uintptr
	pQueryFullProcessImageNameA                   uintptr
	pQueryFullProcessImageNameW                   uintptr
	pGetStartupInfoA                              uintptr
	pCreateProcessWithLogonW                      uintptr
	pCreateProcessWithTokenW                      uintptr
	pRegisterWaitForSingleObject                  uintptr
	pUnregisterWait                               uintptr
	pSetTimerQueueTimer                           uintptr
	pCreatePrivateNamespaceA                      uintptr
	pOpenPrivateNamespaceA                        uintptr
	pCreateBoundaryDescriptorA                    uintptr
	pAddIntegrityLabelToBoundaryDescriptor        uintptr
	pGetActiveProcessorGroupCount                 uintptr
	pGetMaximumProcessorGroupCount                uintptr
	pGetActiveProcessorCount                      uintptr
	pGetMaximumProcessorCount                     uintptr
	pGetNumaProcessorNode                         uintptr
	pGetNumaNodeNumberFromHandle                  uintptr
	pGetNumaProcessorNodeEx                       uintptr
	pGetNumaNodeProcessorMask                     uintptr
	pGetNumaAvailableMemoryNode                   uintptr
	pGetNumaAvailableMemoryNodeEx                 uintptr
	pGetNumaProximityNode                         uintptr
)

func GetProcessWorkingSetSize(hProcess HANDLE, lpMinimumWorkingSetSize *uintptr, lpMaximumWorkingSetSize *uintptr) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pGetProcessWorkingSetSize, libKernel32, "GetProcessWorkingSetSize")
	ret, _, err := syscall.SyscallN(addr, hProcess, uintptr(unsafe.Pointer(lpMinimumWorkingSetSize)), uintptr(unsafe.Pointer(lpMaximumWorkingSetSize)))
	return BOOL(ret), WIN32_ERROR(err)
}

func SetProcessWorkingSetSize(hProcess HANDLE, dwMinimumWorkingSetSize uintptr, dwMaximumWorkingSetSize uintptr) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pSetProcessWorkingSetSize, libKernel32, "SetProcessWorkingSetSize")
	ret, _, err := syscall.SyscallN(addr, hProcess, dwMinimumWorkingSetSize, dwMaximumWorkingSetSize)
	return BOOL(ret), WIN32_ERROR(err)
}

func FlsAlloc(lpCallback PFLS_CALLBACK_FUNCTION) (uint32, WIN32_ERROR) {
	addr := lazyAddr(&pFlsAlloc, libKernel32, "FlsAlloc")
	ret, _, err := syscall.SyscallN(addr, lpCallback)
	return uint32(ret), WIN32_ERROR(err)
}

func FlsGetValue(dwFlsIndex uint32) (unsafe.Pointer, WIN32_ERROR) {
	addr := lazyAddr(&pFlsGetValue, libKernel32, "FlsGetValue")
	ret, _, err := syscall.SyscallN(addr, uintptr(dwFlsIndex))
	return (unsafe.Pointer)(ret), WIN32_ERROR(err)
}

func FlsSetValue(dwFlsIndex uint32, lpFlsData unsafe.Pointer) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pFlsSetValue, libKernel32, "FlsSetValue")
	ret, _, err := syscall.SyscallN(addr, uintptr(dwFlsIndex), uintptr(lpFlsData))
	return BOOL(ret), WIN32_ERROR(err)
}

func FlsFree(dwFlsIndex uint32) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pFlsFree, libKernel32, "FlsFree")
	ret, _, err := syscall.SyscallN(addr, uintptr(dwFlsIndex))
	return BOOL(ret), WIN32_ERROR(err)
}

func IsThreadAFiber() BOOL {
	addr := lazyAddr(&pIsThreadAFiber, libKernel32, "IsThreadAFiber")
	ret, _, _ := syscall.SyscallN(addr)
	return BOOL(ret)
}

func InitializeSRWLock(SRWLock *RTL_SRWLOCK) {
	addr := lazyAddr(&pInitializeSRWLock, libKernel32, "InitializeSRWLock")
	syscall.SyscallN(addr, uintptr(unsafe.Pointer(SRWLock)))
}

func ReleaseSRWLockExclusive(SRWLock *RTL_SRWLOCK) {
	addr := lazyAddr(&pReleaseSRWLockExclusive, libKernel32, "ReleaseSRWLockExclusive")
	syscall.SyscallN(addr, uintptr(unsafe.Pointer(SRWLock)))
}

func ReleaseSRWLockShared(SRWLock *RTL_SRWLOCK) {
	addr := lazyAddr(&pReleaseSRWLockShared, libKernel32, "ReleaseSRWLockShared")
	syscall.SyscallN(addr, uintptr(unsafe.Pointer(SRWLock)))
}

func AcquireSRWLockExclusive(SRWLock *RTL_SRWLOCK) {
	addr := lazyAddr(&pAcquireSRWLockExclusive, libKernel32, "AcquireSRWLockExclusive")
	syscall.SyscallN(addr, uintptr(unsafe.Pointer(SRWLock)))
}

func AcquireSRWLockShared(SRWLock *RTL_SRWLOCK) {
	addr := lazyAddr(&pAcquireSRWLockShared, libKernel32, "AcquireSRWLockShared")
	syscall.SyscallN(addr, uintptr(unsafe.Pointer(SRWLock)))
}

func TryAcquireSRWLockExclusive(SRWLock *RTL_SRWLOCK) BOOLEAN {
	addr := lazyAddr(&pTryAcquireSRWLockExclusive, libKernel32, "TryAcquireSRWLockExclusive")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(SRWLock)))
	return BOOLEAN(ret)
}

func TryAcquireSRWLockShared(SRWLock *RTL_SRWLOCK) BOOLEAN {
	addr := lazyAddr(&pTryAcquireSRWLockShared, libKernel32, "TryAcquireSRWLockShared")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(SRWLock)))
	return BOOLEAN(ret)
}

func InitializeCriticalSection(lpCriticalSection *RTL_CRITICAL_SECTION) {
	addr := lazyAddr(&pInitializeCriticalSection, libKernel32, "InitializeCriticalSection")
	syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpCriticalSection)))
}

func EnterCriticalSection(lpCriticalSection *RTL_CRITICAL_SECTION) {
	addr := lazyAddr(&pEnterCriticalSection, libKernel32, "EnterCriticalSection")
	syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpCriticalSection)))
}

func LeaveCriticalSection(lpCriticalSection *RTL_CRITICAL_SECTION) {
	addr := lazyAddr(&pLeaveCriticalSection, libKernel32, "LeaveCriticalSection")
	syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpCriticalSection)))
}

func InitializeCriticalSectionAndSpinCount(lpCriticalSection *RTL_CRITICAL_SECTION, dwSpinCount uint32) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pInitializeCriticalSectionAndSpinCount, libKernel32, "InitializeCriticalSectionAndSpinCount")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpCriticalSection)), uintptr(dwSpinCount))
	return BOOL(ret), WIN32_ERROR(err)
}

func InitializeCriticalSectionEx(lpCriticalSection *RTL_CRITICAL_SECTION, dwSpinCount uint32, Flags uint32) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pInitializeCriticalSectionEx, libKernel32, "InitializeCriticalSectionEx")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpCriticalSection)), uintptr(dwSpinCount), uintptr(Flags))
	return BOOL(ret), WIN32_ERROR(err)
}

func SetCriticalSectionSpinCount(lpCriticalSection *RTL_CRITICAL_SECTION, dwSpinCount uint32) uint32 {
	addr := lazyAddr(&pSetCriticalSectionSpinCount, libKernel32, "SetCriticalSectionSpinCount")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpCriticalSection)), uintptr(dwSpinCount))
	return uint32(ret)
}

func TryEnterCriticalSection(lpCriticalSection *RTL_CRITICAL_SECTION) BOOL {
	addr := lazyAddr(&pTryEnterCriticalSection, libKernel32, "TryEnterCriticalSection")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpCriticalSection)))
	return BOOL(ret)
}

func DeleteCriticalSection(lpCriticalSection *RTL_CRITICAL_SECTION) {
	addr := lazyAddr(&pDeleteCriticalSection, libKernel32, "DeleteCriticalSection")
	syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpCriticalSection)))
}

func InitOnceInitialize(InitOnce *RTL_RUN_ONCE) {
	addr := lazyAddr(&pInitOnceInitialize, libKernel32, "InitOnceInitialize")
	syscall.SyscallN(addr, uintptr(unsafe.Pointer(InitOnce)))
}

func InitOnceExecuteOnce(InitOnce *RTL_RUN_ONCE, InitFn PINIT_ONCE_FN, Parameter unsafe.Pointer, Context unsafe.Pointer) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pInitOnceExecuteOnce, libKernel32, "InitOnceExecuteOnce")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(InitOnce)), InitFn, uintptr(Parameter), uintptr(Context))
	return BOOL(ret), WIN32_ERROR(err)
}

func InitOnceBeginInitialize(lpInitOnce *RTL_RUN_ONCE, dwFlags uint32, fPending *BOOL, lpContext unsafe.Pointer) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pInitOnceBeginInitialize, libKernel32, "InitOnceBeginInitialize")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpInitOnce)), uintptr(dwFlags), uintptr(unsafe.Pointer(fPending)), uintptr(lpContext))
	return BOOL(ret), WIN32_ERROR(err)
}

func InitOnceComplete(lpInitOnce *RTL_RUN_ONCE, dwFlags uint32, lpContext unsafe.Pointer) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pInitOnceComplete, libKernel32, "InitOnceComplete")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpInitOnce)), uintptr(dwFlags), uintptr(lpContext))
	return BOOL(ret), WIN32_ERROR(err)
}

func InitializeConditionVariable(ConditionVariable *RTL_CONDITION_VARIABLE) {
	addr := lazyAddr(&pInitializeConditionVariable, libKernel32, "InitializeConditionVariable")
	syscall.SyscallN(addr, uintptr(unsafe.Pointer(ConditionVariable)))
}

func WakeConditionVariable(ConditionVariable *RTL_CONDITION_VARIABLE) {
	addr := lazyAddr(&pWakeConditionVariable, libKernel32, "WakeConditionVariable")
	syscall.SyscallN(addr, uintptr(unsafe.Pointer(ConditionVariable)))
}

func WakeAllConditionVariable(ConditionVariable *RTL_CONDITION_VARIABLE) {
	addr := lazyAddr(&pWakeAllConditionVariable, libKernel32, "WakeAllConditionVariable")
	syscall.SyscallN(addr, uintptr(unsafe.Pointer(ConditionVariable)))
}

func SleepConditionVariableCS(ConditionVariable *RTL_CONDITION_VARIABLE, CriticalSection *RTL_CRITICAL_SECTION, dwMilliseconds uint32) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pSleepConditionVariableCS, libKernel32, "SleepConditionVariableCS")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(ConditionVariable)), uintptr(unsafe.Pointer(CriticalSection)), uintptr(dwMilliseconds))
	return BOOL(ret), WIN32_ERROR(err)
}

func SleepConditionVariableSRW(ConditionVariable *RTL_CONDITION_VARIABLE, SRWLock *RTL_SRWLOCK, dwMilliseconds uint32, Flags uint32) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pSleepConditionVariableSRW, libKernel32, "SleepConditionVariableSRW")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(ConditionVariable)), uintptr(unsafe.Pointer(SRWLock)), uintptr(dwMilliseconds), uintptr(Flags))
	return BOOL(ret), WIN32_ERROR(err)
}

func SetEvent(hEvent HANDLE) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pSetEvent, libKernel32, "SetEvent")
	ret, _, err := syscall.SyscallN(addr, hEvent)
	return BOOL(ret), WIN32_ERROR(err)
}

func ResetEvent(hEvent HANDLE) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pResetEvent, libKernel32, "ResetEvent")
	ret, _, err := syscall.SyscallN(addr, hEvent)
	return BOOL(ret), WIN32_ERROR(err)
}

func ReleaseSemaphore(hSemaphore HANDLE, lReleaseCount int32, lpPreviousCount *int32) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pReleaseSemaphore, libKernel32, "ReleaseSemaphore")
	ret, _, err := syscall.SyscallN(addr, hSemaphore, uintptr(lReleaseCount), uintptr(unsafe.Pointer(lpPreviousCount)))
	return BOOL(ret), WIN32_ERROR(err)
}

func ReleaseMutex(hMutex HANDLE) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pReleaseMutex, libKernel32, "ReleaseMutex")
	ret, _, err := syscall.SyscallN(addr, hMutex)
	return BOOL(ret), WIN32_ERROR(err)
}

func WaitForSingleObject(hHandle HANDLE, dwMilliseconds uint32) (uint32, WIN32_ERROR) {
	addr := lazyAddr(&pWaitForSingleObject, libKernel32, "WaitForSingleObject")
	ret, _, err := syscall.SyscallN(addr, hHandle, uintptr(dwMilliseconds))
	return uint32(ret), WIN32_ERROR(err)
}

func SleepEx(dwMilliseconds uint32, bAlertable BOOL) uint32 {
	addr := lazyAddr(&pSleepEx, libKernel32, "SleepEx")
	ret, _, _ := syscall.SyscallN(addr, uintptr(dwMilliseconds), uintptr(bAlertable))
	return uint32(ret)
}

func WaitForSingleObjectEx(hHandle HANDLE, dwMilliseconds uint32, bAlertable BOOL) (uint32, WIN32_ERROR) {
	addr := lazyAddr(&pWaitForSingleObjectEx, libKernel32, "WaitForSingleObjectEx")
	ret, _, err := syscall.SyscallN(addr, hHandle, uintptr(dwMilliseconds), uintptr(bAlertable))
	return uint32(ret), WIN32_ERROR(err)
}

func WaitForMultipleObjectsEx(nCount uint32, lpHandles *HANDLE, bWaitAll BOOL, dwMilliseconds uint32, bAlertable BOOL) (uint32, WIN32_ERROR) {
	addr := lazyAddr(&pWaitForMultipleObjectsEx, libKernel32, "WaitForMultipleObjectsEx")
	ret, _, err := syscall.SyscallN(addr, uintptr(nCount), uintptr(unsafe.Pointer(lpHandles)), uintptr(bWaitAll), uintptr(dwMilliseconds), uintptr(bAlertable))
	return uint32(ret), WIN32_ERROR(err)
}

func CreateMutexA(lpMutexAttributes *SECURITY_ATTRIBUTES, bInitialOwner BOOL, lpName PSTR) (HANDLE, WIN32_ERROR) {
	addr := lazyAddr(&pCreateMutexA, libKernel32, "CreateMutexA")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpMutexAttributes)), uintptr(bInitialOwner), uintptr(unsafe.Pointer(lpName)))
	return ret, WIN32_ERROR(err)
}

var CreateMutex = CreateMutexW

func CreateMutexW(lpMutexAttributes *SECURITY_ATTRIBUTES, bInitialOwner BOOL, lpName PWSTR) (HANDLE, WIN32_ERROR) {
	addr := lazyAddr(&pCreateMutexW, libKernel32, "CreateMutexW")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpMutexAttributes)), uintptr(bInitialOwner), uintptr(unsafe.Pointer(lpName)))
	return ret, WIN32_ERROR(err)
}

func OpenMutexW(dwDesiredAccess uint32, bInheritHandle BOOL, lpName PWSTR) (HANDLE, WIN32_ERROR) {
	addr := lazyAddr(&pOpenMutexW, libKernel32, "OpenMutexW")
	ret, _, err := syscall.SyscallN(addr, uintptr(dwDesiredAccess), uintptr(bInheritHandle), uintptr(unsafe.Pointer(lpName)))
	return ret, WIN32_ERROR(err)
}

func CreateEventA(lpEventAttributes *SECURITY_ATTRIBUTES, bManualReset BOOL, bInitialState BOOL, lpName PSTR) (HANDLE, WIN32_ERROR) {
	addr := lazyAddr(&pCreateEventA, libKernel32, "CreateEventA")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpEventAttributes)), uintptr(bManualReset), uintptr(bInitialState), uintptr(unsafe.Pointer(lpName)))
	return ret, WIN32_ERROR(err)
}

var CreateEvent = CreateEventW

func CreateEventW(lpEventAttributes *SECURITY_ATTRIBUTES, bManualReset BOOL, bInitialState BOOL, lpName PWSTR) (HANDLE, WIN32_ERROR) {
	addr := lazyAddr(&pCreateEventW, libKernel32, "CreateEventW")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpEventAttributes)), uintptr(bManualReset), uintptr(bInitialState), uintptr(unsafe.Pointer(lpName)))
	return ret, WIN32_ERROR(err)
}

func OpenEventA(dwDesiredAccess uint32, bInheritHandle BOOL, lpName PSTR) (HANDLE, WIN32_ERROR) {
	addr := lazyAddr(&pOpenEventA, libKernel32, "OpenEventA")
	ret, _, err := syscall.SyscallN(addr, uintptr(dwDesiredAccess), uintptr(bInheritHandle), uintptr(unsafe.Pointer(lpName)))
	return ret, WIN32_ERROR(err)
}

var OpenEvent = OpenEventW

func OpenEventW(dwDesiredAccess uint32, bInheritHandle BOOL, lpName PWSTR) (HANDLE, WIN32_ERROR) {
	addr := lazyAddr(&pOpenEventW, libKernel32, "OpenEventW")
	ret, _, err := syscall.SyscallN(addr, uintptr(dwDesiredAccess), uintptr(bInheritHandle), uintptr(unsafe.Pointer(lpName)))
	return ret, WIN32_ERROR(err)
}

func OpenSemaphoreW(dwDesiredAccess uint32, bInheritHandle BOOL, lpName PWSTR) (HANDLE, WIN32_ERROR) {
	addr := lazyAddr(&pOpenSemaphoreW, libKernel32, "OpenSemaphoreW")
	ret, _, err := syscall.SyscallN(addr, uintptr(dwDesiredAccess), uintptr(bInheritHandle), uintptr(unsafe.Pointer(lpName)))
	return ret, WIN32_ERROR(err)
}

func OpenWaitableTimerW(dwDesiredAccess uint32, bInheritHandle BOOL, lpTimerName PWSTR) (HANDLE, WIN32_ERROR) {
	addr := lazyAddr(&pOpenWaitableTimerW, libKernel32, "OpenWaitableTimerW")
	ret, _, err := syscall.SyscallN(addr, uintptr(dwDesiredAccess), uintptr(bInheritHandle), uintptr(unsafe.Pointer(lpTimerName)))
	return ret, WIN32_ERROR(err)
}

func SetWaitableTimerEx(hTimer HANDLE, lpDueTime *int64, lPeriod int32, pfnCompletionRoutine PTIMERAPCROUTINE, lpArgToCompletionRoutine unsafe.Pointer, WakeContext *REASON_CONTEXT, TolerableDelay uint32) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pSetWaitableTimerEx, libKernel32, "SetWaitableTimerEx")
	ret, _, err := syscall.SyscallN(addr, hTimer, uintptr(unsafe.Pointer(lpDueTime)), uintptr(lPeriod), pfnCompletionRoutine, uintptr(lpArgToCompletionRoutine), uintptr(unsafe.Pointer(WakeContext)), uintptr(TolerableDelay))
	return BOOL(ret), WIN32_ERROR(err)
}

func SetWaitableTimer(hTimer HANDLE, lpDueTime *int64, lPeriod int32, pfnCompletionRoutine PTIMERAPCROUTINE, lpArgToCompletionRoutine unsafe.Pointer, fResume BOOL) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pSetWaitableTimer, libKernel32, "SetWaitableTimer")
	ret, _, err := syscall.SyscallN(addr, hTimer, uintptr(unsafe.Pointer(lpDueTime)), uintptr(lPeriod), pfnCompletionRoutine, uintptr(lpArgToCompletionRoutine), uintptr(fResume))
	return BOOL(ret), WIN32_ERROR(err)
}

func CancelWaitableTimer(hTimer HANDLE) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pCancelWaitableTimer, libKernel32, "CancelWaitableTimer")
	ret, _, err := syscall.SyscallN(addr, hTimer)
	return BOOL(ret), WIN32_ERROR(err)
}

func CreateMutexExA(lpMutexAttributes *SECURITY_ATTRIBUTES, lpName PSTR, dwFlags uint32, dwDesiredAccess uint32) (HANDLE, WIN32_ERROR) {
	addr := lazyAddr(&pCreateMutexExA, libKernel32, "CreateMutexExA")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpMutexAttributes)), uintptr(unsafe.Pointer(lpName)), uintptr(dwFlags), uintptr(dwDesiredAccess))
	return ret, WIN32_ERROR(err)
}

var CreateMutexEx = CreateMutexExW

func CreateMutexExW(lpMutexAttributes *SECURITY_ATTRIBUTES, lpName PWSTR, dwFlags uint32, dwDesiredAccess uint32) (HANDLE, WIN32_ERROR) {
	addr := lazyAddr(&pCreateMutexExW, libKernel32, "CreateMutexExW")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpMutexAttributes)), uintptr(unsafe.Pointer(lpName)), uintptr(dwFlags), uintptr(dwDesiredAccess))
	return ret, WIN32_ERROR(err)
}

func CreateEventExA(lpEventAttributes *SECURITY_ATTRIBUTES, lpName PSTR, dwFlags CREATE_EVENT, dwDesiredAccess uint32) (HANDLE, WIN32_ERROR) {
	addr := lazyAddr(&pCreateEventExA, libKernel32, "CreateEventExA")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpEventAttributes)), uintptr(unsafe.Pointer(lpName)), uintptr(dwFlags), uintptr(dwDesiredAccess))
	return ret, WIN32_ERROR(err)
}

var CreateEventEx = CreateEventExW

func CreateEventExW(lpEventAttributes *SECURITY_ATTRIBUTES, lpName PWSTR, dwFlags CREATE_EVENT, dwDesiredAccess uint32) (HANDLE, WIN32_ERROR) {
	addr := lazyAddr(&pCreateEventExW, libKernel32, "CreateEventExW")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpEventAttributes)), uintptr(unsafe.Pointer(lpName)), uintptr(dwFlags), uintptr(dwDesiredAccess))
	return ret, WIN32_ERROR(err)
}

var CreateSemaphoreEx = CreateSemaphoreExW

func CreateSemaphoreExW(lpSemaphoreAttributes *SECURITY_ATTRIBUTES, lInitialCount int32, lMaximumCount int32, lpName PWSTR, dwFlags uint32, dwDesiredAccess uint32) (HANDLE, WIN32_ERROR) {
	addr := lazyAddr(&pCreateSemaphoreExW, libKernel32, "CreateSemaphoreExW")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpSemaphoreAttributes)), uintptr(lInitialCount), uintptr(lMaximumCount), uintptr(unsafe.Pointer(lpName)), uintptr(dwFlags), uintptr(dwDesiredAccess))
	return ret, WIN32_ERROR(err)
}

func CreateWaitableTimerExW(lpTimerAttributes *SECURITY_ATTRIBUTES, lpTimerName PWSTR, dwFlags uint32, dwDesiredAccess uint32) (HANDLE, WIN32_ERROR) {
	addr := lazyAddr(&pCreateWaitableTimerExW, libKernel32, "CreateWaitableTimerExW")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpTimerAttributes)), uintptr(unsafe.Pointer(lpTimerName)), uintptr(dwFlags), uintptr(dwDesiredAccess))
	return ret, WIN32_ERROR(err)
}

func EnterSynchronizationBarrier(lpBarrier *RTL_BARRIER, dwFlags uint32) BOOL {
	addr := lazyAddr(&pEnterSynchronizationBarrier, libKernel32, "EnterSynchronizationBarrier")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpBarrier)), uintptr(dwFlags))
	return BOOL(ret)
}

func InitializeSynchronizationBarrier(lpBarrier *RTL_BARRIER, lTotalThreads int32, lSpinCount int32) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pInitializeSynchronizationBarrier, libKernel32, "InitializeSynchronizationBarrier")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpBarrier)), uintptr(lTotalThreads), uintptr(lSpinCount))
	return BOOL(ret), WIN32_ERROR(err)
}

func DeleteSynchronizationBarrier(lpBarrier *RTL_BARRIER) BOOL {
	addr := lazyAddr(&pDeleteSynchronizationBarrier, libKernel32, "DeleteSynchronizationBarrier")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpBarrier)))
	return BOOL(ret)
}

func Sleep(dwMilliseconds uint32) {
	addr := lazyAddr(&pSleep, libKernel32, "Sleep")
	syscall.SyscallN(addr, uintptr(dwMilliseconds))
}

func WaitForMultipleObjects(nCount uint32, lpHandles *HANDLE, bWaitAll BOOL, dwMilliseconds uint32) (uint32, WIN32_ERROR) {
	addr := lazyAddr(&pWaitForMultipleObjects, libKernel32, "WaitForMultipleObjects")
	ret, _, err := syscall.SyscallN(addr, uintptr(nCount), uintptr(unsafe.Pointer(lpHandles)), uintptr(bWaitAll), uintptr(dwMilliseconds))
	return uint32(ret), WIN32_ERROR(err)
}

var CreateSemaphore = CreateSemaphoreW

func CreateSemaphoreW(lpSemaphoreAttributes *SECURITY_ATTRIBUTES, lInitialCount int32, lMaximumCount int32, lpName PWSTR) (HANDLE, WIN32_ERROR) {
	addr := lazyAddr(&pCreateSemaphoreW, libKernel32, "CreateSemaphoreW")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpSemaphoreAttributes)), uintptr(lInitialCount), uintptr(lMaximumCount), uintptr(unsafe.Pointer(lpName)))
	return ret, WIN32_ERROR(err)
}

func CreateWaitableTimerW(lpTimerAttributes *SECURITY_ATTRIBUTES, bManualReset BOOL, lpTimerName PWSTR) (HANDLE, WIN32_ERROR) {
	addr := lazyAddr(&pCreateWaitableTimerW, libKernel32, "CreateWaitableTimerW")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpTimerAttributes)), uintptr(bManualReset), uintptr(unsafe.Pointer(lpTimerName)))
	return ret, WIN32_ERROR(err)
}

func InitializeSListHead(ListHead *SLIST_HEADER) {
	addr := lazyAddr(&pInitializeSListHead, libKernel32, "InitializeSListHead")
	syscall.SyscallN(addr, uintptr(unsafe.Pointer(ListHead)))
}

func InterlockedPopEntrySList(ListHead *SLIST_HEADER) *SLIST_ENTRY {
	addr := lazyAddr(&pInterlockedPopEntrySList, libKernel32, "InterlockedPopEntrySList")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(ListHead)))
	return (*SLIST_ENTRY)(unsafe.Pointer(ret))
}

func InterlockedPushEntrySList(ListHead *SLIST_HEADER, ListEntry *SLIST_ENTRY) *SLIST_ENTRY {
	addr := lazyAddr(&pInterlockedPushEntrySList, libKernel32, "InterlockedPushEntrySList")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(ListHead)), uintptr(unsafe.Pointer(ListEntry)))
	return (*SLIST_ENTRY)(unsafe.Pointer(ret))
}

func InterlockedPushListSListEx(ListHead *SLIST_HEADER, List *SLIST_ENTRY, ListEnd *SLIST_ENTRY, Count uint32) *SLIST_ENTRY {
	addr := lazyAddr(&pInterlockedPushListSListEx, libKernel32, "InterlockedPushListSListEx")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(ListHead)), uintptr(unsafe.Pointer(List)), uintptr(unsafe.Pointer(ListEnd)), uintptr(Count))
	return (*SLIST_ENTRY)(unsafe.Pointer(ret))
}

func InterlockedFlushSList(ListHead *SLIST_HEADER) *SLIST_ENTRY {
	addr := lazyAddr(&pInterlockedFlushSList, libKernel32, "InterlockedFlushSList")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(ListHead)))
	return (*SLIST_ENTRY)(unsafe.Pointer(ret))
}

func QueryDepthSList(ListHead *SLIST_HEADER) uint16 {
	addr := lazyAddr(&pQueryDepthSList, libKernel32, "QueryDepthSList")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(ListHead)))
	return uint16(ret)
}

func QueueUserAPC(pfnAPC PAPCFUNC, hThread HANDLE, dwData uintptr) (uint32, WIN32_ERROR) {
	addr := lazyAddr(&pQueueUserAPC, libKernel32, "QueueUserAPC")
	ret, _, err := syscall.SyscallN(addr, pfnAPC, hThread, dwData)
	return uint32(ret), WIN32_ERROR(err)
}

func QueueUserAPC2(ApcRoutine PAPCFUNC, Thread HANDLE, Data uintptr, Flags QUEUE_USER_APC_FLAGS) BOOL {
	addr := lazyAddr(&pQueueUserAPC2, libKernel32, "QueueUserAPC2")
	ret, _, _ := syscall.SyscallN(addr, ApcRoutine, Thread, Data, uintptr(Flags))
	return BOOL(ret)
}

func GetProcessTimes(hProcess HANDLE, lpCreationTime *FILETIME, lpExitTime *FILETIME, lpKernelTime *FILETIME, lpUserTime *FILETIME) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pGetProcessTimes, libKernel32, "GetProcessTimes")
	ret, _, err := syscall.SyscallN(addr, hProcess, uintptr(unsafe.Pointer(lpCreationTime)), uintptr(unsafe.Pointer(lpExitTime)), uintptr(unsafe.Pointer(lpKernelTime)), uintptr(unsafe.Pointer(lpUserTime)))
	return BOOL(ret), WIN32_ERROR(err)
}

func GetCurrentProcess() HANDLE {
	addr := lazyAddr(&pGetCurrentProcess, libKernel32, "GetCurrentProcess")
	ret, _, _ := syscall.SyscallN(addr)
	return ret
}

func GetCurrentProcessId() uint32 {
	addr := lazyAddr(&pGetCurrentProcessId, libKernel32, "GetCurrentProcessId")
	ret, _, _ := syscall.SyscallN(addr)
	return uint32(ret)
}

func ExitProcess(uExitCode uint32) {
	addr := lazyAddr(&pExitProcess, libKernel32, "ExitProcess")
	syscall.SyscallN(addr, uintptr(uExitCode))
}

func TerminateProcess(hProcess HANDLE, uExitCode uint32) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pTerminateProcess, libKernel32, "TerminateProcess")
	ret, _, err := syscall.SyscallN(addr, hProcess, uintptr(uExitCode))
	return BOOL(ret), WIN32_ERROR(err)
}

func GetExitCodeProcess(hProcess HANDLE, lpExitCode *uint32) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pGetExitCodeProcess, libKernel32, "GetExitCodeProcess")
	ret, _, err := syscall.SyscallN(addr, hProcess, uintptr(unsafe.Pointer(lpExitCode)))
	return BOOL(ret), WIN32_ERROR(err)
}

func SwitchToThread() BOOL {
	addr := lazyAddr(&pSwitchToThread, libKernel32, "SwitchToThread")
	ret, _, _ := syscall.SyscallN(addr)
	return BOOL(ret)
}

func CreateThread(lpThreadAttributes *SECURITY_ATTRIBUTES, dwStackSize uintptr, lpStartAddress LPTHREAD_START_ROUTINE, lpParameter unsafe.Pointer, dwCreationFlags THREAD_CREATION_FLAGS, lpThreadId *uint32) (HANDLE, WIN32_ERROR) {
	addr := lazyAddr(&pCreateThread, libKernel32, "CreateThread")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpThreadAttributes)), dwStackSize, lpStartAddress, uintptr(lpParameter), uintptr(dwCreationFlags), uintptr(unsafe.Pointer(lpThreadId)))
	return ret, WIN32_ERROR(err)
}

func CreateRemoteThread(hProcess HANDLE, lpThreadAttributes *SECURITY_ATTRIBUTES, dwStackSize uintptr, lpStartAddress LPTHREAD_START_ROUTINE, lpParameter unsafe.Pointer, dwCreationFlags uint32, lpThreadId *uint32) (HANDLE, WIN32_ERROR) {
	addr := lazyAddr(&pCreateRemoteThread, libKernel32, "CreateRemoteThread")
	ret, _, err := syscall.SyscallN(addr, hProcess, uintptr(unsafe.Pointer(lpThreadAttributes)), dwStackSize, lpStartAddress, uintptr(lpParameter), uintptr(dwCreationFlags), uintptr(unsafe.Pointer(lpThreadId)))
	return ret, WIN32_ERROR(err)
}

func GetCurrentThread() HANDLE {
	addr := lazyAddr(&pGetCurrentThread, libKernel32, "GetCurrentThread")
	ret, _, _ := syscall.SyscallN(addr)
	return ret
}

func GetCurrentThreadId() uint32 {
	addr := lazyAddr(&pGetCurrentThreadId, libKernel32, "GetCurrentThreadId")
	ret, _, _ := syscall.SyscallN(addr)
	return uint32(ret)
}

func OpenThread(dwDesiredAccess THREAD_ACCESS_RIGHTS, bInheritHandle BOOL, dwThreadId uint32) (HANDLE, WIN32_ERROR) {
	addr := lazyAddr(&pOpenThread, libKernel32, "OpenThread")
	ret, _, err := syscall.SyscallN(addr, uintptr(dwDesiredAccess), uintptr(bInheritHandle), uintptr(dwThreadId))
	return ret, WIN32_ERROR(err)
}

func SetThreadPriority(hThread HANDLE, nPriority THREAD_PRIORITY) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pSetThreadPriority, libKernel32, "SetThreadPriority")
	ret, _, err := syscall.SyscallN(addr, hThread, uintptr(nPriority))
	return BOOL(ret), WIN32_ERROR(err)
}

func SetThreadPriorityBoost(hThread HANDLE, bDisablePriorityBoost BOOL) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pSetThreadPriorityBoost, libKernel32, "SetThreadPriorityBoost")
	ret, _, err := syscall.SyscallN(addr, hThread, uintptr(bDisablePriorityBoost))
	return BOOL(ret), WIN32_ERROR(err)
}

func GetThreadPriorityBoost(hThread HANDLE, pDisablePriorityBoost *BOOL) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pGetThreadPriorityBoost, libKernel32, "GetThreadPriorityBoost")
	ret, _, err := syscall.SyscallN(addr, hThread, uintptr(unsafe.Pointer(pDisablePriorityBoost)))
	return BOOL(ret), WIN32_ERROR(err)
}

func GetThreadPriority(hThread HANDLE) (int32, WIN32_ERROR) {
	addr := lazyAddr(&pGetThreadPriority, libKernel32, "GetThreadPriority")
	ret, _, err := syscall.SyscallN(addr, hThread)
	return int32(ret), WIN32_ERROR(err)
}

func ExitThread(dwExitCode uint32) {
	addr := lazyAddr(&pExitThread, libKernel32, "ExitThread")
	syscall.SyscallN(addr, uintptr(dwExitCode))
}

func TerminateThread(hThread HANDLE, dwExitCode uint32) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pTerminateThread, libKernel32, "TerminateThread")
	ret, _, err := syscall.SyscallN(addr, hThread, uintptr(dwExitCode))
	return BOOL(ret), WIN32_ERROR(err)
}

func GetExitCodeThread(hThread HANDLE, lpExitCode *uint32) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pGetExitCodeThread, libKernel32, "GetExitCodeThread")
	ret, _, err := syscall.SyscallN(addr, hThread, uintptr(unsafe.Pointer(lpExitCode)))
	return BOOL(ret), WIN32_ERROR(err)
}

func SuspendThread(hThread HANDLE) (uint32, WIN32_ERROR) {
	addr := lazyAddr(&pSuspendThread, libKernel32, "SuspendThread")
	ret, _, err := syscall.SyscallN(addr, hThread)
	return uint32(ret), WIN32_ERROR(err)
}

func ResumeThread(hThread HANDLE) (uint32, WIN32_ERROR) {
	addr := lazyAddr(&pResumeThread, libKernel32, "ResumeThread")
	ret, _, err := syscall.SyscallN(addr, hThread)
	return uint32(ret), WIN32_ERROR(err)
}

func TlsAlloc() (uint32, WIN32_ERROR) {
	addr := lazyAddr(&pTlsAlloc, libKernel32, "TlsAlloc")
	ret, _, err := syscall.SyscallN(addr)
	return uint32(ret), WIN32_ERROR(err)
}

func TlsGetValue(dwTlsIndex uint32) (unsafe.Pointer, WIN32_ERROR) {
	addr := lazyAddr(&pTlsGetValue, libKernel32, "TlsGetValue")
	ret, _, err := syscall.SyscallN(addr, uintptr(dwTlsIndex))
	return (unsafe.Pointer)(ret), WIN32_ERROR(err)
}

func TlsSetValue(dwTlsIndex uint32, lpTlsValue unsafe.Pointer) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pTlsSetValue, libKernel32, "TlsSetValue")
	ret, _, err := syscall.SyscallN(addr, uintptr(dwTlsIndex), uintptr(lpTlsValue))
	return BOOL(ret), WIN32_ERROR(err)
}

func TlsFree(dwTlsIndex uint32) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pTlsFree, libKernel32, "TlsFree")
	ret, _, err := syscall.SyscallN(addr, uintptr(dwTlsIndex))
	return BOOL(ret), WIN32_ERROR(err)
}

func CreateProcessA(lpApplicationName PSTR, lpCommandLine PSTR, lpProcessAttributes *SECURITY_ATTRIBUTES, lpThreadAttributes *SECURITY_ATTRIBUTES, bInheritHandles BOOL, dwCreationFlags PROCESS_CREATION_FLAGS, lpEnvironment unsafe.Pointer, lpCurrentDirectory PSTR, lpStartupInfo *STARTUPINFOA, lpProcessInformation *PROCESS_INFORMATION) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pCreateProcessA, libKernel32, "CreateProcessA")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpApplicationName)), uintptr(unsafe.Pointer(lpCommandLine)), uintptr(unsafe.Pointer(lpProcessAttributes)), uintptr(unsafe.Pointer(lpThreadAttributes)), uintptr(bInheritHandles), uintptr(dwCreationFlags), uintptr(lpEnvironment), uintptr(unsafe.Pointer(lpCurrentDirectory)), uintptr(unsafe.Pointer(lpStartupInfo)), uintptr(unsafe.Pointer(lpProcessInformation)))
	return BOOL(ret), WIN32_ERROR(err)
}

var CreateProcess = CreateProcessW

func CreateProcessW(lpApplicationName PWSTR, lpCommandLine PWSTR, lpProcessAttributes *SECURITY_ATTRIBUTES, lpThreadAttributes *SECURITY_ATTRIBUTES, bInheritHandles BOOL, dwCreationFlags PROCESS_CREATION_FLAGS, lpEnvironment unsafe.Pointer, lpCurrentDirectory PWSTR, lpStartupInfo *STARTUPINFOW, lpProcessInformation *PROCESS_INFORMATION) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pCreateProcessW, libKernel32, "CreateProcessW")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpApplicationName)), uintptr(unsafe.Pointer(lpCommandLine)), uintptr(unsafe.Pointer(lpProcessAttributes)), uintptr(unsafe.Pointer(lpThreadAttributes)), uintptr(bInheritHandles), uintptr(dwCreationFlags), uintptr(lpEnvironment), uintptr(unsafe.Pointer(lpCurrentDirectory)), uintptr(unsafe.Pointer(lpStartupInfo)), uintptr(unsafe.Pointer(lpProcessInformation)))
	return BOOL(ret), WIN32_ERROR(err)
}

func SetProcessShutdownParameters(dwLevel uint32, dwFlags uint32) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pSetProcessShutdownParameters, libKernel32, "SetProcessShutdownParameters")
	ret, _, err := syscall.SyscallN(addr, uintptr(dwLevel), uintptr(dwFlags))
	return BOOL(ret), WIN32_ERROR(err)
}

func GetProcessVersion(ProcessId uint32) (uint32, WIN32_ERROR) {
	addr := lazyAddr(&pGetProcessVersion, libKernel32, "GetProcessVersion")
	ret, _, err := syscall.SyscallN(addr, uintptr(ProcessId))
	return uint32(ret), WIN32_ERROR(err)
}

var GetStartupInfo = GetStartupInfoW

func GetStartupInfoW(lpStartupInfo *STARTUPINFOW) {
	addr := lazyAddr(&pGetStartupInfoW, libKernel32, "GetStartupInfoW")
	syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpStartupInfo)))
}

var CreateProcessAsUser = CreateProcessAsUserW

func CreateProcessAsUserW(hToken HANDLE, lpApplicationName PWSTR, lpCommandLine PWSTR, lpProcessAttributes *SECURITY_ATTRIBUTES, lpThreadAttributes *SECURITY_ATTRIBUTES, bInheritHandles BOOL, dwCreationFlags uint32, lpEnvironment unsafe.Pointer, lpCurrentDirectory PWSTR, lpStartupInfo *STARTUPINFOW, lpProcessInformation *PROCESS_INFORMATION) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pCreateProcessAsUserW, libAdvapi32, "CreateProcessAsUserW")
	ret, _, err := syscall.SyscallN(addr, hToken, uintptr(unsafe.Pointer(lpApplicationName)), uintptr(unsafe.Pointer(lpCommandLine)), uintptr(unsafe.Pointer(lpProcessAttributes)), uintptr(unsafe.Pointer(lpThreadAttributes)), uintptr(bInheritHandles), uintptr(dwCreationFlags), uintptr(lpEnvironment), uintptr(unsafe.Pointer(lpCurrentDirectory)), uintptr(unsafe.Pointer(lpStartupInfo)), uintptr(unsafe.Pointer(lpProcessInformation)))
	return BOOL(ret), WIN32_ERROR(err)
}

func SetThreadToken(Thread *HANDLE, Token HANDLE) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pSetThreadToken, libAdvapi32, "SetThreadToken")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(Thread)), Token)
	return BOOL(ret), WIN32_ERROR(err)
}

func OpenProcessToken(ProcessHandle HANDLE, DesiredAccess TOKEN_ACCESS_MASK, TokenHandle *HANDLE) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pOpenProcessToken, libAdvapi32, "OpenProcessToken")
	ret, _, err := syscall.SyscallN(addr, ProcessHandle, uintptr(DesiredAccess), uintptr(unsafe.Pointer(TokenHandle)))
	return BOOL(ret), WIN32_ERROR(err)
}

func OpenThreadToken(ThreadHandle HANDLE, DesiredAccess TOKEN_ACCESS_MASK, OpenAsSelf BOOL, TokenHandle *HANDLE) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pOpenThreadToken, libAdvapi32, "OpenThreadToken")
	ret, _, err := syscall.SyscallN(addr, ThreadHandle, uintptr(DesiredAccess), uintptr(OpenAsSelf), uintptr(unsafe.Pointer(TokenHandle)))
	return BOOL(ret), WIN32_ERROR(err)
}

func SetPriorityClass(hProcess HANDLE, dwPriorityClass PROCESS_CREATION_FLAGS) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pSetPriorityClass, libKernel32, "SetPriorityClass")
	ret, _, err := syscall.SyscallN(addr, hProcess, uintptr(dwPriorityClass))
	return BOOL(ret), WIN32_ERROR(err)
}

func GetPriorityClass(hProcess HANDLE) (uint32, WIN32_ERROR) {
	addr := lazyAddr(&pGetPriorityClass, libKernel32, "GetPriorityClass")
	ret, _, err := syscall.SyscallN(addr, hProcess)
	return uint32(ret), WIN32_ERROR(err)
}

func SetThreadStackGuarantee(StackSizeInBytes *uint32) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pSetThreadStackGuarantee, libKernel32, "SetThreadStackGuarantee")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(StackSizeInBytes)))
	return BOOL(ret), WIN32_ERROR(err)
}

func GetProcessId(Process HANDLE) (uint32, WIN32_ERROR) {
	addr := lazyAddr(&pGetProcessId, libKernel32, "GetProcessId")
	ret, _, err := syscall.SyscallN(addr, Process)
	return uint32(ret), WIN32_ERROR(err)
}

func GetThreadId(Thread HANDLE) (uint32, WIN32_ERROR) {
	addr := lazyAddr(&pGetThreadId, libKernel32, "GetThreadId")
	ret, _, err := syscall.SyscallN(addr, Thread)
	return uint32(ret), WIN32_ERROR(err)
}

func FlushProcessWriteBuffers() {
	addr := lazyAddr(&pFlushProcessWriteBuffers, libKernel32, "FlushProcessWriteBuffers")
	syscall.SyscallN(addr)
}

func GetProcessIdOfThread(Thread HANDLE) (uint32, WIN32_ERROR) {
	addr := lazyAddr(&pGetProcessIdOfThread, libKernel32, "GetProcessIdOfThread")
	ret, _, err := syscall.SyscallN(addr, Thread)
	return uint32(ret), WIN32_ERROR(err)
}

func InitializeProcThreadAttributeList(lpAttributeList LPPROC_THREAD_ATTRIBUTE_LIST, dwAttributeCount uint32, dwFlags uint32, lpSize *uintptr) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pInitializeProcThreadAttributeList, libKernel32, "InitializeProcThreadAttributeList")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpAttributeList)), uintptr(dwAttributeCount), uintptr(dwFlags), uintptr(unsafe.Pointer(lpSize)))
	return BOOL(ret), WIN32_ERROR(err)
}

func DeleteProcThreadAttributeList(lpAttributeList LPPROC_THREAD_ATTRIBUTE_LIST) {
	addr := lazyAddr(&pDeleteProcThreadAttributeList, libKernel32, "DeleteProcThreadAttributeList")
	syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpAttributeList)))
}

func UpdateProcThreadAttribute(lpAttributeList LPPROC_THREAD_ATTRIBUTE_LIST, dwFlags uint32, Attribute uintptr, lpValue unsafe.Pointer, cbSize uintptr, lpPreviousValue unsafe.Pointer, lpReturnSize *uintptr) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pUpdateProcThreadAttribute, libKernel32, "UpdateProcThreadAttribute")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpAttributeList)), uintptr(dwFlags), Attribute, uintptr(lpValue), cbSize, uintptr(lpPreviousValue), uintptr(unsafe.Pointer(lpReturnSize)))
	return BOOL(ret), WIN32_ERROR(err)
}

func SetProcessDynamicEHContinuationTargets(Process HANDLE, NumberOfTargets uint16, Targets *PROCESS_DYNAMIC_EH_CONTINUATION_TARGET) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pSetProcessDynamicEHContinuationTargets, libKernel32, "SetProcessDynamicEHContinuationTargets")
	ret, _, err := syscall.SyscallN(addr, Process, uintptr(NumberOfTargets), uintptr(unsafe.Pointer(Targets)))
	return BOOL(ret), WIN32_ERROR(err)
}

func SetProcessDynamicEnforcedCetCompatibleRanges(Process HANDLE, NumberOfRanges uint16, Ranges *PROCESS_DYNAMIC_ENFORCED_ADDRESS_RANGE) BOOL {
	addr := lazyAddr(&pSetProcessDynamicEnforcedCetCompatibleRanges, libKernel32, "SetProcessDynamicEnforcedCetCompatibleRanges")
	ret, _, _ := syscall.SyscallN(addr, Process, uintptr(NumberOfRanges), uintptr(unsafe.Pointer(Ranges)))
	return BOOL(ret)
}

func SetProcessAffinityUpdateMode(hProcess HANDLE, dwFlags PROCESS_AFFINITY_AUTO_UPDATE_FLAGS) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pSetProcessAffinityUpdateMode, libKernel32, "SetProcessAffinityUpdateMode")
	ret, _, err := syscall.SyscallN(addr, hProcess, uintptr(dwFlags))
	return BOOL(ret), WIN32_ERROR(err)
}

func QueryProcessAffinityUpdateMode(hProcess HANDLE, lpdwFlags *PROCESS_AFFINITY_AUTO_UPDATE_FLAGS) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pQueryProcessAffinityUpdateMode, libKernel32, "QueryProcessAffinityUpdateMode")
	ret, _, err := syscall.SyscallN(addr, hProcess, uintptr(unsafe.Pointer(lpdwFlags)))
	return BOOL(ret), WIN32_ERROR(err)
}

func CreateRemoteThreadEx(hProcess HANDLE, lpThreadAttributes *SECURITY_ATTRIBUTES, dwStackSize uintptr, lpStartAddress LPTHREAD_START_ROUTINE, lpParameter unsafe.Pointer, dwCreationFlags uint32, lpAttributeList LPPROC_THREAD_ATTRIBUTE_LIST, lpThreadId *uint32) (HANDLE, WIN32_ERROR) {
	addr := lazyAddr(&pCreateRemoteThreadEx, libKernel32, "CreateRemoteThreadEx")
	ret, _, err := syscall.SyscallN(addr, hProcess, uintptr(unsafe.Pointer(lpThreadAttributes)), dwStackSize, lpStartAddress, uintptr(lpParameter), uintptr(dwCreationFlags), uintptr(unsafe.Pointer(lpAttributeList)), uintptr(unsafe.Pointer(lpThreadId)))
	return ret, WIN32_ERROR(err)
}

func GetCurrentThreadStackLimits(LowLimit *uintptr, HighLimit *uintptr) {
	addr := lazyAddr(&pGetCurrentThreadStackLimits, libKernel32, "GetCurrentThreadStackLimits")
	syscall.SyscallN(addr, uintptr(unsafe.Pointer(LowLimit)), uintptr(unsafe.Pointer(HighLimit)))
}

func GetProcessMitigationPolicy(hProcess HANDLE, MitigationPolicy PROCESS_MITIGATION_POLICY, lpBuffer unsafe.Pointer, dwLength uintptr) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pGetProcessMitigationPolicy, libKernel32, "GetProcessMitigationPolicy")
	ret, _, err := syscall.SyscallN(addr, hProcess, uintptr(MitigationPolicy), uintptr(lpBuffer), dwLength)
	return BOOL(ret), WIN32_ERROR(err)
}

func SetProcessMitigationPolicy(MitigationPolicy PROCESS_MITIGATION_POLICY, lpBuffer unsafe.Pointer, dwLength uintptr) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pSetProcessMitigationPolicy, libKernel32, "SetProcessMitigationPolicy")
	ret, _, err := syscall.SyscallN(addr, uintptr(MitigationPolicy), uintptr(lpBuffer), dwLength)
	return BOOL(ret), WIN32_ERROR(err)
}

func GetThreadTimes(hThread HANDLE, lpCreationTime *FILETIME, lpExitTime *FILETIME, lpKernelTime *FILETIME, lpUserTime *FILETIME) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pGetThreadTimes, libKernel32, "GetThreadTimes")
	ret, _, err := syscall.SyscallN(addr, hThread, uintptr(unsafe.Pointer(lpCreationTime)), uintptr(unsafe.Pointer(lpExitTime)), uintptr(unsafe.Pointer(lpKernelTime)), uintptr(unsafe.Pointer(lpUserTime)))
	return BOOL(ret), WIN32_ERROR(err)
}

func OpenProcess(dwDesiredAccess PROCESS_ACCESS_RIGHTS, bInheritHandle BOOL, dwProcessId uint32) (HANDLE, WIN32_ERROR) {
	addr := lazyAddr(&pOpenProcess, libKernel32, "OpenProcess")
	ret, _, err := syscall.SyscallN(addr, uintptr(dwDesiredAccess), uintptr(bInheritHandle), uintptr(dwProcessId))
	return ret, WIN32_ERROR(err)
}

func IsProcessorFeaturePresent(ProcessorFeature PROCESSOR_FEATURE_ID) BOOL {
	addr := lazyAddr(&pIsProcessorFeaturePresent, libKernel32, "IsProcessorFeaturePresent")
	ret, _, _ := syscall.SyscallN(addr, uintptr(ProcessorFeature))
	return BOOL(ret)
}

func GetProcessHandleCount(hProcess HANDLE, pdwHandleCount *uint32) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pGetProcessHandleCount, libKernel32, "GetProcessHandleCount")
	ret, _, err := syscall.SyscallN(addr, hProcess, uintptr(unsafe.Pointer(pdwHandleCount)))
	return BOOL(ret), WIN32_ERROR(err)
}

func GetCurrentProcessorNumber() uint32 {
	addr := lazyAddr(&pGetCurrentProcessorNumber, libKernel32, "GetCurrentProcessorNumber")
	ret, _, _ := syscall.SyscallN(addr)
	return uint32(ret)
}

func SetThreadIdealProcessorEx(hThread HANDLE, lpIdealProcessor *PROCESSOR_NUMBER, lpPreviousIdealProcessor *PROCESSOR_NUMBER) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pSetThreadIdealProcessorEx, libKernel32, "SetThreadIdealProcessorEx")
	ret, _, err := syscall.SyscallN(addr, hThread, uintptr(unsafe.Pointer(lpIdealProcessor)), uintptr(unsafe.Pointer(lpPreviousIdealProcessor)))
	return BOOL(ret), WIN32_ERROR(err)
}

func GetThreadIdealProcessorEx(hThread HANDLE, lpIdealProcessor *PROCESSOR_NUMBER) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pGetThreadIdealProcessorEx, libKernel32, "GetThreadIdealProcessorEx")
	ret, _, err := syscall.SyscallN(addr, hThread, uintptr(unsafe.Pointer(lpIdealProcessor)))
	return BOOL(ret), WIN32_ERROR(err)
}

func GetCurrentProcessorNumberEx(ProcNumber *PROCESSOR_NUMBER) {
	addr := lazyAddr(&pGetCurrentProcessorNumberEx, libKernel32, "GetCurrentProcessorNumberEx")
	syscall.SyscallN(addr, uintptr(unsafe.Pointer(ProcNumber)))
}

func GetProcessPriorityBoost(hProcess HANDLE, pDisablePriorityBoost *BOOL) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pGetProcessPriorityBoost, libKernel32, "GetProcessPriorityBoost")
	ret, _, err := syscall.SyscallN(addr, hProcess, uintptr(unsafe.Pointer(pDisablePriorityBoost)))
	return BOOL(ret), WIN32_ERROR(err)
}

func SetProcessPriorityBoost(hProcess HANDLE, bDisablePriorityBoost BOOL) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pSetProcessPriorityBoost, libKernel32, "SetProcessPriorityBoost")
	ret, _, err := syscall.SyscallN(addr, hProcess, uintptr(bDisablePriorityBoost))
	return BOOL(ret), WIN32_ERROR(err)
}

func GetThreadIOPendingFlag(hThread HANDLE, lpIOIsPending *BOOL) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pGetThreadIOPendingFlag, libKernel32, "GetThreadIOPendingFlag")
	ret, _, err := syscall.SyscallN(addr, hThread, uintptr(unsafe.Pointer(lpIOIsPending)))
	return BOOL(ret), WIN32_ERROR(err)
}

func GetSystemTimes(lpIdleTime *FILETIME, lpKernelTime *FILETIME, lpUserTime *FILETIME) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pGetSystemTimes, libKernel32, "GetSystemTimes")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpIdleTime)), uintptr(unsafe.Pointer(lpKernelTime)), uintptr(unsafe.Pointer(lpUserTime)))
	return BOOL(ret), WIN32_ERROR(err)
}

func GetThreadInformation(hThread HANDLE, ThreadInformationClass THREAD_INFORMATION_CLASS, ThreadInformation unsafe.Pointer, ThreadInformationSize uint32) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pGetThreadInformation, libKernel32, "GetThreadInformation")
	ret, _, err := syscall.SyscallN(addr, hThread, uintptr(ThreadInformationClass), uintptr(ThreadInformation), uintptr(ThreadInformationSize))
	return BOOL(ret), WIN32_ERROR(err)
}

func SetThreadInformation(hThread HANDLE, ThreadInformationClass THREAD_INFORMATION_CLASS, ThreadInformation unsafe.Pointer, ThreadInformationSize uint32) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pSetThreadInformation, libKernel32, "SetThreadInformation")
	ret, _, err := syscall.SyscallN(addr, hThread, uintptr(ThreadInformationClass), uintptr(ThreadInformation), uintptr(ThreadInformationSize))
	return BOOL(ret), WIN32_ERROR(err)
}

func IsProcessCritical(hProcess HANDLE, Critical *BOOL) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pIsProcessCritical, libKernel32, "IsProcessCritical")
	ret, _, err := syscall.SyscallN(addr, hProcess, uintptr(unsafe.Pointer(Critical)))
	return BOOL(ret), WIN32_ERROR(err)
}

func SetProtectedPolicy(PolicyGuid *syscall.GUID, PolicyValue uintptr, OldPolicyValue *uintptr) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pSetProtectedPolicy, libKernel32, "SetProtectedPolicy")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(PolicyGuid)), PolicyValue, uintptr(unsafe.Pointer(OldPolicyValue)))
	return BOOL(ret), WIN32_ERROR(err)
}

func QueryProtectedPolicy(PolicyGuid *syscall.GUID, PolicyValue *uintptr) BOOL {
	addr := lazyAddr(&pQueryProtectedPolicy, libKernel32, "QueryProtectedPolicy")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(PolicyGuid)), uintptr(unsafe.Pointer(PolicyValue)))
	return BOOL(ret)
}

func SetThreadIdealProcessor(hThread HANDLE, dwIdealProcessor uint32) (uint32, WIN32_ERROR) {
	addr := lazyAddr(&pSetThreadIdealProcessor, libKernel32, "SetThreadIdealProcessor")
	ret, _, err := syscall.SyscallN(addr, hThread, uintptr(dwIdealProcessor))
	return uint32(ret), WIN32_ERROR(err)
}

func SetProcessInformation(hProcess HANDLE, ProcessInformationClass PROCESS_INFORMATION_CLASS, ProcessInformation unsafe.Pointer, ProcessInformationSize uint32) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pSetProcessInformation, libKernel32, "SetProcessInformation")
	ret, _, err := syscall.SyscallN(addr, hProcess, uintptr(ProcessInformationClass), uintptr(ProcessInformation), uintptr(ProcessInformationSize))
	return BOOL(ret), WIN32_ERROR(err)
}

func GetProcessInformation(hProcess HANDLE, ProcessInformationClass PROCESS_INFORMATION_CLASS, ProcessInformation unsafe.Pointer, ProcessInformationSize uint32) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pGetProcessInformation, libKernel32, "GetProcessInformation")
	ret, _, err := syscall.SyscallN(addr, hProcess, uintptr(ProcessInformationClass), uintptr(ProcessInformation), uintptr(ProcessInformationSize))
	return BOOL(ret), WIN32_ERROR(err)
}

func GetProcessDefaultCpuSets(Process HANDLE, CpuSetIds *uint32, CpuSetIdCount uint32, RequiredIdCount *uint32) BOOL {
	addr := lazyAddr(&pGetProcessDefaultCpuSets, libKernel32, "GetProcessDefaultCpuSets")
	ret, _, _ := syscall.SyscallN(addr, Process, uintptr(unsafe.Pointer(CpuSetIds)), uintptr(CpuSetIdCount), uintptr(unsafe.Pointer(RequiredIdCount)))
	return BOOL(ret)
}

func SetProcessDefaultCpuSets(Process HANDLE, CpuSetIds *uint32, CpuSetIdCount uint32) BOOL {
	addr := lazyAddr(&pSetProcessDefaultCpuSets, libKernel32, "SetProcessDefaultCpuSets")
	ret, _, _ := syscall.SyscallN(addr, Process, uintptr(unsafe.Pointer(CpuSetIds)), uintptr(CpuSetIdCount))
	return BOOL(ret)
}

func GetThreadSelectedCpuSets(Thread HANDLE, CpuSetIds *uint32, CpuSetIdCount uint32, RequiredIdCount *uint32) BOOL {
	addr := lazyAddr(&pGetThreadSelectedCpuSets, libKernel32, "GetThreadSelectedCpuSets")
	ret, _, _ := syscall.SyscallN(addr, Thread, uintptr(unsafe.Pointer(CpuSetIds)), uintptr(CpuSetIdCount), uintptr(unsafe.Pointer(RequiredIdCount)))
	return BOOL(ret)
}

func SetThreadSelectedCpuSets(Thread HANDLE, CpuSetIds *uint32, CpuSetIdCount uint32) BOOL {
	addr := lazyAddr(&pSetThreadSelectedCpuSets, libKernel32, "SetThreadSelectedCpuSets")
	ret, _, _ := syscall.SyscallN(addr, Thread, uintptr(unsafe.Pointer(CpuSetIds)), uintptr(CpuSetIdCount))
	return BOOL(ret)
}

func CreateProcessAsUserA(hToken HANDLE, lpApplicationName PSTR, lpCommandLine PSTR, lpProcessAttributes *SECURITY_ATTRIBUTES, lpThreadAttributes *SECURITY_ATTRIBUTES, bInheritHandles BOOL, dwCreationFlags uint32, lpEnvironment unsafe.Pointer, lpCurrentDirectory PSTR, lpStartupInfo *STARTUPINFOA, lpProcessInformation *PROCESS_INFORMATION) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pCreateProcessAsUserA, libAdvapi32, "CreateProcessAsUserA")
	ret, _, err := syscall.SyscallN(addr, hToken, uintptr(unsafe.Pointer(lpApplicationName)), uintptr(unsafe.Pointer(lpCommandLine)), uintptr(unsafe.Pointer(lpProcessAttributes)), uintptr(unsafe.Pointer(lpThreadAttributes)), uintptr(bInheritHandles), uintptr(dwCreationFlags), uintptr(lpEnvironment), uintptr(unsafe.Pointer(lpCurrentDirectory)), uintptr(unsafe.Pointer(lpStartupInfo)), uintptr(unsafe.Pointer(lpProcessInformation)))
	return BOOL(ret), WIN32_ERROR(err)
}

func GetProcessShutdownParameters(lpdwLevel *uint32, lpdwFlags *uint32) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pGetProcessShutdownParameters, libKernel32, "GetProcessShutdownParameters")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpdwLevel)), uintptr(unsafe.Pointer(lpdwFlags)))
	return BOOL(ret), WIN32_ERROR(err)
}

func GetProcessDefaultCpuSetMasks(Process HANDLE, CpuSetMasks *GROUP_AFFINITY, CpuSetMaskCount uint16, RequiredMaskCount *uint16) BOOL {
	addr := lazyAddr(&pGetProcessDefaultCpuSetMasks, libKernel32, "GetProcessDefaultCpuSetMasks")
	ret, _, _ := syscall.SyscallN(addr, Process, uintptr(unsafe.Pointer(CpuSetMasks)), uintptr(CpuSetMaskCount), uintptr(unsafe.Pointer(RequiredMaskCount)))
	return BOOL(ret)
}

func SetProcessDefaultCpuSetMasks(Process HANDLE, CpuSetMasks *GROUP_AFFINITY, CpuSetMaskCount uint16) BOOL {
	addr := lazyAddr(&pSetProcessDefaultCpuSetMasks, libKernel32, "SetProcessDefaultCpuSetMasks")
	ret, _, _ := syscall.SyscallN(addr, Process, uintptr(unsafe.Pointer(CpuSetMasks)), uintptr(CpuSetMaskCount))
	return BOOL(ret)
}

func GetThreadSelectedCpuSetMasks(Thread HANDLE, CpuSetMasks *GROUP_AFFINITY, CpuSetMaskCount uint16, RequiredMaskCount *uint16) BOOL {
	addr := lazyAddr(&pGetThreadSelectedCpuSetMasks, libKernel32, "GetThreadSelectedCpuSetMasks")
	ret, _, _ := syscall.SyscallN(addr, Thread, uintptr(unsafe.Pointer(CpuSetMasks)), uintptr(CpuSetMaskCount), uintptr(unsafe.Pointer(RequiredMaskCount)))
	return BOOL(ret)
}

func SetThreadSelectedCpuSetMasks(Thread HANDLE, CpuSetMasks *GROUP_AFFINITY, CpuSetMaskCount uint16) BOOL {
	addr := lazyAddr(&pSetThreadSelectedCpuSetMasks, libKernel32, "SetThreadSelectedCpuSetMasks")
	ret, _, _ := syscall.SyscallN(addr, Thread, uintptr(unsafe.Pointer(CpuSetMasks)), uintptr(CpuSetMaskCount))
	return BOOL(ret)
}

func GetMachineTypeAttributes(Machine uint16, MachineTypeAttributes *MACHINE_ATTRIBUTES) HRESULT {
	addr := lazyAddr(&pGetMachineTypeAttributes, libKernel32, "GetMachineTypeAttributes")
	ret, _, _ := syscall.SyscallN(addr, uintptr(Machine), uintptr(unsafe.Pointer(MachineTypeAttributes)))
	return HRESULT(ret)
}

func SetThreadDescription(hThread HANDLE, lpThreadDescription PWSTR) HRESULT {
	addr := lazyAddr(&pSetThreadDescription, libKernel32, "SetThreadDescription")
	ret, _, _ := syscall.SyscallN(addr, hThread, uintptr(unsafe.Pointer(lpThreadDescription)))
	return HRESULT(ret)
}

func GetThreadDescription(hThread HANDLE, ppszThreadDescription *PWSTR) HRESULT {
	addr := lazyAddr(&pGetThreadDescription, libKernel32, "GetThreadDescription")
	ret, _, _ := syscall.SyscallN(addr, hThread, uintptr(unsafe.Pointer(ppszThreadDescription)))
	return HRESULT(ret)
}

func QueueUserWorkItem(Function LPTHREAD_START_ROUTINE, Context unsafe.Pointer, Flags WORKER_THREAD_FLAGS) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pQueueUserWorkItem, libKernel32, "QueueUserWorkItem")
	ret, _, err := syscall.SyscallN(addr, Function, uintptr(Context), uintptr(Flags))
	return BOOL(ret), WIN32_ERROR(err)
}

func UnregisterWaitEx(WaitHandle HANDLE, CompletionEvent HANDLE) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pUnregisterWaitEx, libKernel32, "UnregisterWaitEx")
	ret, _, err := syscall.SyscallN(addr, WaitHandle, CompletionEvent)
	return BOOL(ret), WIN32_ERROR(err)
}

func CreateTimerQueue() (HANDLE, WIN32_ERROR) {
	addr := lazyAddr(&pCreateTimerQueue, libKernel32, "CreateTimerQueue")
	ret, _, err := syscall.SyscallN(addr)
	return ret, WIN32_ERROR(err)
}

func CreateTimerQueueTimer(phNewTimer *HANDLE, TimerQueue HANDLE, Callback WAITORTIMERCALLBACK, Parameter unsafe.Pointer, DueTime uint32, Period uint32, Flags WORKER_THREAD_FLAGS) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pCreateTimerQueueTimer, libKernel32, "CreateTimerQueueTimer")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(phNewTimer)), TimerQueue, Callback, uintptr(Parameter), uintptr(DueTime), uintptr(Period), uintptr(Flags))
	return BOOL(ret), WIN32_ERROR(err)
}

func ChangeTimerQueueTimer(TimerQueue HANDLE, Timer HANDLE, DueTime uint32, Period uint32) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pChangeTimerQueueTimer, libKernel32, "ChangeTimerQueueTimer")
	ret, _, err := syscall.SyscallN(addr, TimerQueue, Timer, uintptr(DueTime), uintptr(Period))
	return BOOL(ret), WIN32_ERROR(err)
}

func DeleteTimerQueueTimer(TimerQueue HANDLE, Timer HANDLE, CompletionEvent HANDLE) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pDeleteTimerQueueTimer, libKernel32, "DeleteTimerQueueTimer")
	ret, _, err := syscall.SyscallN(addr, TimerQueue, Timer, CompletionEvent)
	return BOOL(ret), WIN32_ERROR(err)
}

func DeleteTimerQueue(TimerQueue HANDLE) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pDeleteTimerQueue, libKernel32, "DeleteTimerQueue")
	ret, _, err := syscall.SyscallN(addr, TimerQueue)
	return BOOL(ret), WIN32_ERROR(err)
}

func DeleteTimerQueueEx(TimerQueue HANDLE, CompletionEvent HANDLE) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pDeleteTimerQueueEx, libKernel32, "DeleteTimerQueueEx")
	ret, _, err := syscall.SyscallN(addr, TimerQueue, CompletionEvent)
	return BOOL(ret), WIN32_ERROR(err)
}

func CreateThreadpool(reserved unsafe.Pointer) (PTP_POOL, WIN32_ERROR) {
	addr := lazyAddr(&pCreateThreadpool, libKernel32, "CreateThreadpool")
	ret, _, err := syscall.SyscallN(addr, uintptr(reserved))
	return ret, WIN32_ERROR(err)
}

func SetThreadpoolThreadMaximum(ptpp PTP_POOL, cthrdMost uint32) {
	addr := lazyAddr(&pSetThreadpoolThreadMaximum, libKernel32, "SetThreadpoolThreadMaximum")
	syscall.SyscallN(addr, ptpp, uintptr(cthrdMost))
}

func SetThreadpoolThreadMinimum(ptpp PTP_POOL, cthrdMic uint32) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pSetThreadpoolThreadMinimum, libKernel32, "SetThreadpoolThreadMinimum")
	ret, _, err := syscall.SyscallN(addr, ptpp, uintptr(cthrdMic))
	return BOOL(ret), WIN32_ERROR(err)
}

func SetThreadpoolStackInformation(ptpp PTP_POOL, ptpsi *TP_POOL_STACK_INFORMATION) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pSetThreadpoolStackInformation, libKernel32, "SetThreadpoolStackInformation")
	ret, _, err := syscall.SyscallN(addr, ptpp, uintptr(unsafe.Pointer(ptpsi)))
	return BOOL(ret), WIN32_ERROR(err)
}

func QueryThreadpoolStackInformation(ptpp PTP_POOL, ptpsi *TP_POOL_STACK_INFORMATION) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pQueryThreadpoolStackInformation, libKernel32, "QueryThreadpoolStackInformation")
	ret, _, err := syscall.SyscallN(addr, ptpp, uintptr(unsafe.Pointer(ptpsi)))
	return BOOL(ret), WIN32_ERROR(err)
}

func CloseThreadpool(ptpp PTP_POOL) {
	addr := lazyAddr(&pCloseThreadpool, libKernel32, "CloseThreadpool")
	syscall.SyscallN(addr, ptpp)
}

func CreateThreadpoolCleanupGroup() (uintptr, WIN32_ERROR) {
	addr := lazyAddr(&pCreateThreadpoolCleanupGroup, libKernel32, "CreateThreadpoolCleanupGroup")
	ret, _, err := syscall.SyscallN(addr)
	return ret, WIN32_ERROR(err)
}

func CloseThreadpoolCleanupGroupMembers(ptpcg uintptr, fCancelPendingCallbacks BOOL, pvCleanupContext unsafe.Pointer) {
	addr := lazyAddr(&pCloseThreadpoolCleanupGroupMembers, libKernel32, "CloseThreadpoolCleanupGroupMembers")
	syscall.SyscallN(addr, ptpcg, uintptr(fCancelPendingCallbacks), uintptr(pvCleanupContext))
}

func CloseThreadpoolCleanupGroup(ptpcg uintptr) {
	addr := lazyAddr(&pCloseThreadpoolCleanupGroup, libKernel32, "CloseThreadpoolCleanupGroup")
	syscall.SyscallN(addr, ptpcg)
}

func SetEventWhenCallbackReturns(pci *TP_CALLBACK_INSTANCE, evt HANDLE) {
	addr := lazyAddr(&pSetEventWhenCallbackReturns, libKernel32, "SetEventWhenCallbackReturns")
	syscall.SyscallN(addr, uintptr(unsafe.Pointer(pci)), evt)
}

func ReleaseSemaphoreWhenCallbackReturns(pci *TP_CALLBACK_INSTANCE, sem HANDLE, crel uint32) {
	addr := lazyAddr(&pReleaseSemaphoreWhenCallbackReturns, libKernel32, "ReleaseSemaphoreWhenCallbackReturns")
	syscall.SyscallN(addr, uintptr(unsafe.Pointer(pci)), sem, uintptr(crel))
}

func ReleaseMutexWhenCallbackReturns(pci *TP_CALLBACK_INSTANCE, mut HANDLE) {
	addr := lazyAddr(&pReleaseMutexWhenCallbackReturns, libKernel32, "ReleaseMutexWhenCallbackReturns")
	syscall.SyscallN(addr, uintptr(unsafe.Pointer(pci)), mut)
}

func LeaveCriticalSectionWhenCallbackReturns(pci *TP_CALLBACK_INSTANCE, pcs *RTL_CRITICAL_SECTION) {
	addr := lazyAddr(&pLeaveCriticalSectionWhenCallbackReturns, libKernel32, "LeaveCriticalSectionWhenCallbackReturns")
	syscall.SyscallN(addr, uintptr(unsafe.Pointer(pci)), uintptr(unsafe.Pointer(pcs)))
}

func FreeLibraryWhenCallbackReturns(pci *TP_CALLBACK_INSTANCE, mod HINSTANCE) {
	addr := lazyAddr(&pFreeLibraryWhenCallbackReturns, libKernel32, "FreeLibraryWhenCallbackReturns")
	syscall.SyscallN(addr, uintptr(unsafe.Pointer(pci)), mod)
}

func CallbackMayRunLong(pci *TP_CALLBACK_INSTANCE) BOOL {
	addr := lazyAddr(&pCallbackMayRunLong, libKernel32, "CallbackMayRunLong")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pci)))
	return BOOL(ret)
}

func DisassociateCurrentThreadFromCallback(pci *TP_CALLBACK_INSTANCE) {
	addr := lazyAddr(&pDisassociateCurrentThreadFromCallback, libKernel32, "DisassociateCurrentThreadFromCallback")
	syscall.SyscallN(addr, uintptr(unsafe.Pointer(pci)))
}

func TrySubmitThreadpoolCallback(pfns PTP_SIMPLE_CALLBACK, pv unsafe.Pointer, pcbe *TP_CALLBACK_ENVIRON_V3) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pTrySubmitThreadpoolCallback, libKernel32, "TrySubmitThreadpoolCallback")
	ret, _, err := syscall.SyscallN(addr, pfns, uintptr(pv), uintptr(unsafe.Pointer(pcbe)))
	return BOOL(ret), WIN32_ERROR(err)
}

func CreateThreadpoolWork(pfnwk PTP_WORK_CALLBACK, pv unsafe.Pointer, pcbe *TP_CALLBACK_ENVIRON_V3) (*TP_WORK, WIN32_ERROR) {
	addr := lazyAddr(&pCreateThreadpoolWork, libKernel32, "CreateThreadpoolWork")
	ret, _, err := syscall.SyscallN(addr, pfnwk, uintptr(pv), uintptr(unsafe.Pointer(pcbe)))
	return (*TP_WORK)(unsafe.Pointer(ret)), WIN32_ERROR(err)
}

func SubmitThreadpoolWork(pwk *TP_WORK) {
	addr := lazyAddr(&pSubmitThreadpoolWork, libKernel32, "SubmitThreadpoolWork")
	syscall.SyscallN(addr, uintptr(unsafe.Pointer(pwk)))
}

func WaitForThreadpoolWorkCallbacks(pwk *TP_WORK, fCancelPendingCallbacks BOOL) {
	addr := lazyAddr(&pWaitForThreadpoolWorkCallbacks, libKernel32, "WaitForThreadpoolWorkCallbacks")
	syscall.SyscallN(addr, uintptr(unsafe.Pointer(pwk)), uintptr(fCancelPendingCallbacks))
}

func CloseThreadpoolWork(pwk *TP_WORK) {
	addr := lazyAddr(&pCloseThreadpoolWork, libKernel32, "CloseThreadpoolWork")
	syscall.SyscallN(addr, uintptr(unsafe.Pointer(pwk)))
}

func CreateThreadpoolTimer(pfnti PTP_TIMER_CALLBACK, pv unsafe.Pointer, pcbe *TP_CALLBACK_ENVIRON_V3) (*TP_TIMER, WIN32_ERROR) {
	addr := lazyAddr(&pCreateThreadpoolTimer, libKernel32, "CreateThreadpoolTimer")
	ret, _, err := syscall.SyscallN(addr, pfnti, uintptr(pv), uintptr(unsafe.Pointer(pcbe)))
	return (*TP_TIMER)(unsafe.Pointer(ret)), WIN32_ERROR(err)
}

func SetThreadpoolTimer(pti *TP_TIMER, pftDueTime *FILETIME, msPeriod uint32, msWindowLength uint32) {
	addr := lazyAddr(&pSetThreadpoolTimer, libKernel32, "SetThreadpoolTimer")
	syscall.SyscallN(addr, uintptr(unsafe.Pointer(pti)), uintptr(unsafe.Pointer(pftDueTime)), uintptr(msPeriod), uintptr(msWindowLength))
}

func IsThreadpoolTimerSet(pti *TP_TIMER) BOOL {
	addr := lazyAddr(&pIsThreadpoolTimerSet, libKernel32, "IsThreadpoolTimerSet")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pti)))
	return BOOL(ret)
}

func WaitForThreadpoolTimerCallbacks(pti *TP_TIMER, fCancelPendingCallbacks BOOL) {
	addr := lazyAddr(&pWaitForThreadpoolTimerCallbacks, libKernel32, "WaitForThreadpoolTimerCallbacks")
	syscall.SyscallN(addr, uintptr(unsafe.Pointer(pti)), uintptr(fCancelPendingCallbacks))
}

func CloseThreadpoolTimer(pti *TP_TIMER) {
	addr := lazyAddr(&pCloseThreadpoolTimer, libKernel32, "CloseThreadpoolTimer")
	syscall.SyscallN(addr, uintptr(unsafe.Pointer(pti)))
}

func CreateThreadpoolWait(pfnwa PTP_WAIT_CALLBACK, pv unsafe.Pointer, pcbe *TP_CALLBACK_ENVIRON_V3) (*TP_WAIT, WIN32_ERROR) {
	addr := lazyAddr(&pCreateThreadpoolWait, libKernel32, "CreateThreadpoolWait")
	ret, _, err := syscall.SyscallN(addr, pfnwa, uintptr(pv), uintptr(unsafe.Pointer(pcbe)))
	return (*TP_WAIT)(unsafe.Pointer(ret)), WIN32_ERROR(err)
}

func SetThreadpoolWait(pwa *TP_WAIT, h HANDLE, pftTimeout *FILETIME) {
	addr := lazyAddr(&pSetThreadpoolWait, libKernel32, "SetThreadpoolWait")
	syscall.SyscallN(addr, uintptr(unsafe.Pointer(pwa)), h, uintptr(unsafe.Pointer(pftTimeout)))
}

func WaitForThreadpoolWaitCallbacks(pwa *TP_WAIT, fCancelPendingCallbacks BOOL) {
	addr := lazyAddr(&pWaitForThreadpoolWaitCallbacks, libKernel32, "WaitForThreadpoolWaitCallbacks")
	syscall.SyscallN(addr, uintptr(unsafe.Pointer(pwa)), uintptr(fCancelPendingCallbacks))
}

func CloseThreadpoolWait(pwa *TP_WAIT) {
	addr := lazyAddr(&pCloseThreadpoolWait, libKernel32, "CloseThreadpoolWait")
	syscall.SyscallN(addr, uintptr(unsafe.Pointer(pwa)))
}

func CreateThreadpoolIo(fl HANDLE, pfnio PTP_WIN32_IO_CALLBACK, pv unsafe.Pointer, pcbe *TP_CALLBACK_ENVIRON_V3) (*TP_IO, WIN32_ERROR) {
	addr := lazyAddr(&pCreateThreadpoolIo, libKernel32, "CreateThreadpoolIo")
	ret, _, err := syscall.SyscallN(addr, fl, pfnio, uintptr(pv), uintptr(unsafe.Pointer(pcbe)))
	return (*TP_IO)(unsafe.Pointer(ret)), WIN32_ERROR(err)
}

func StartThreadpoolIo(pio *TP_IO) {
	addr := lazyAddr(&pStartThreadpoolIo, libKernel32, "StartThreadpoolIo")
	syscall.SyscallN(addr, uintptr(unsafe.Pointer(pio)))
}

func CancelThreadpoolIo(pio *TP_IO) {
	addr := lazyAddr(&pCancelThreadpoolIo, libKernel32, "CancelThreadpoolIo")
	syscall.SyscallN(addr, uintptr(unsafe.Pointer(pio)))
}

func WaitForThreadpoolIoCallbacks(pio *TP_IO, fCancelPendingCallbacks BOOL) {
	addr := lazyAddr(&pWaitForThreadpoolIoCallbacks, libKernel32, "WaitForThreadpoolIoCallbacks")
	syscall.SyscallN(addr, uintptr(unsafe.Pointer(pio)), uintptr(fCancelPendingCallbacks))
}

func CloseThreadpoolIo(pio *TP_IO) {
	addr := lazyAddr(&pCloseThreadpoolIo, libKernel32, "CloseThreadpoolIo")
	syscall.SyscallN(addr, uintptr(unsafe.Pointer(pio)))
}

func SetThreadpoolTimerEx(pti *TP_TIMER, pftDueTime *FILETIME, msPeriod uint32, msWindowLength uint32) BOOL {
	addr := lazyAddr(&pSetThreadpoolTimerEx, libKernel32, "SetThreadpoolTimerEx")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pti)), uintptr(unsafe.Pointer(pftDueTime)), uintptr(msPeriod), uintptr(msWindowLength))
	return BOOL(ret)
}

func SetThreadpoolWaitEx(pwa *TP_WAIT, h HANDLE, pftTimeout *FILETIME, Reserved unsafe.Pointer) BOOL {
	addr := lazyAddr(&pSetThreadpoolWaitEx, libKernel32, "SetThreadpoolWaitEx")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pwa)), h, uintptr(unsafe.Pointer(pftTimeout)), uintptr(Reserved))
	return BOOL(ret)
}

func IsWow64Process(hProcess HANDLE, Wow64Process *BOOL) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pIsWow64Process, libKernel32, "IsWow64Process")
	ret, _, err := syscall.SyscallN(addr, hProcess, uintptr(unsafe.Pointer(Wow64Process)))
	return BOOL(ret), WIN32_ERROR(err)
}

func IsWow64Process2(hProcess HANDLE, pProcessMachine *uint16, pNativeMachine *uint16) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pIsWow64Process2, libKernel32, "IsWow64Process2")
	ret, _, err := syscall.SyscallN(addr, hProcess, uintptr(unsafe.Pointer(pProcessMachine)), uintptr(unsafe.Pointer(pNativeMachine)))
	return BOOL(ret), WIN32_ERROR(err)
}

func Wow64SuspendThread(hThread HANDLE) (uint32, WIN32_ERROR) {
	addr := lazyAddr(&pWow64SuspendThread, libKernel32, "Wow64SuspendThread")
	ret, _, err := syscall.SyscallN(addr, hThread)
	return uint32(ret), WIN32_ERROR(err)
}

var CreatePrivateNamespace = CreatePrivateNamespaceW

func CreatePrivateNamespaceW(lpPrivateNamespaceAttributes *SECURITY_ATTRIBUTES, lpBoundaryDescriptor unsafe.Pointer, lpAliasPrefix PWSTR) NamespaceHandle {
	addr := lazyAddr(&pCreatePrivateNamespaceW, libKernel32, "CreatePrivateNamespaceW")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpPrivateNamespaceAttributes)), uintptr(lpBoundaryDescriptor), uintptr(unsafe.Pointer(lpAliasPrefix)))
	return ret
}

var OpenPrivateNamespace = OpenPrivateNamespaceW

func OpenPrivateNamespaceW(lpBoundaryDescriptor unsafe.Pointer, lpAliasPrefix PWSTR) NamespaceHandle {
	addr := lazyAddr(&pOpenPrivateNamespaceW, libKernel32, "OpenPrivateNamespaceW")
	ret, _, _ := syscall.SyscallN(addr, uintptr(lpBoundaryDescriptor), uintptr(unsafe.Pointer(lpAliasPrefix)))
	return ret
}

func ClosePrivateNamespace(Handle NamespaceHandle, Flags uint32) (BOOLEAN, WIN32_ERROR) {
	addr := lazyAddr(&pClosePrivateNamespace, libKernel32, "ClosePrivateNamespace")
	ret, _, err := syscall.SyscallN(addr, Handle, uintptr(Flags))
	return BOOLEAN(ret), WIN32_ERROR(err)
}

var CreateBoundaryDescriptor = CreateBoundaryDescriptorW

func CreateBoundaryDescriptorW(Name PWSTR, Flags uint32) BoundaryDescriptorHandle {
	addr := lazyAddr(&pCreateBoundaryDescriptorW, libKernel32, "CreateBoundaryDescriptorW")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(Name)), uintptr(Flags))
	return ret
}

func AddSIDToBoundaryDescriptor(BoundaryDescriptor *HANDLE, RequiredSid PSID) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pAddSIDToBoundaryDescriptor, libKernel32, "AddSIDToBoundaryDescriptor")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(BoundaryDescriptor)), uintptr(unsafe.Pointer(RequiredSid)))
	return BOOL(ret), WIN32_ERROR(err)
}

func DeleteBoundaryDescriptor(BoundaryDescriptor BoundaryDescriptorHandle) {
	addr := lazyAddr(&pDeleteBoundaryDescriptor, libKernel32, "DeleteBoundaryDescriptor")
	syscall.SyscallN(addr, BoundaryDescriptor)
}

func GetNumaHighestNodeNumber(HighestNodeNumber *uint32) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pGetNumaHighestNodeNumber, libKernel32, "GetNumaHighestNodeNumber")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(HighestNodeNumber)))
	return BOOL(ret), WIN32_ERROR(err)
}

func GetNumaNodeProcessorMaskEx(Node uint16, ProcessorMask *GROUP_AFFINITY) BOOL {
	addr := lazyAddr(&pGetNumaNodeProcessorMaskEx, libKernel32, "GetNumaNodeProcessorMaskEx")
	ret, _, _ := syscall.SyscallN(addr, uintptr(Node), uintptr(unsafe.Pointer(ProcessorMask)))
	return BOOL(ret)
}

func GetNumaNodeProcessorMask2(NodeNumber uint16, ProcessorMasks *GROUP_AFFINITY, ProcessorMaskCount uint16, RequiredMaskCount *uint16) BOOL {
	addr := lazyAddr(&pGetNumaNodeProcessorMask2, libKernel32, "GetNumaNodeProcessorMask2")
	ret, _, _ := syscall.SyscallN(addr, uintptr(NodeNumber), uintptr(unsafe.Pointer(ProcessorMasks)), uintptr(ProcessorMaskCount), uintptr(unsafe.Pointer(RequiredMaskCount)))
	return BOOL(ret)
}

func GetNumaProximityNodeEx(ProximityId uint32, NodeNumber *uint16) BOOL {
	addr := lazyAddr(&pGetNumaProximityNodeEx, libKernel32, "GetNumaProximityNodeEx")
	ret, _, _ := syscall.SyscallN(addr, uintptr(ProximityId), uintptr(unsafe.Pointer(NodeNumber)))
	return BOOL(ret)
}

func GetProcessGroupAffinity(hProcess HANDLE, GroupCount *uint16, GroupArray *uint16) BOOL {
	addr := lazyAddr(&pGetProcessGroupAffinity, libKernel32, "GetProcessGroupAffinity")
	ret, _, _ := syscall.SyscallN(addr, hProcess, uintptr(unsafe.Pointer(GroupCount)), uintptr(unsafe.Pointer(GroupArray)))
	return BOOL(ret)
}

func GetThreadGroupAffinity(hThread HANDLE, GroupAffinity *GROUP_AFFINITY) BOOL {
	addr := lazyAddr(&pGetThreadGroupAffinity, libKernel32, "GetThreadGroupAffinity")
	ret, _, _ := syscall.SyscallN(addr, hThread, uintptr(unsafe.Pointer(GroupAffinity)))
	return BOOL(ret)
}

func SetThreadGroupAffinity(hThread HANDLE, GroupAffinity *GROUP_AFFINITY, PreviousGroupAffinity *GROUP_AFFINITY) BOOL {
	addr := lazyAddr(&pSetThreadGroupAffinity, libKernel32, "SetThreadGroupAffinity")
	ret, _, _ := syscall.SyscallN(addr, hThread, uintptr(unsafe.Pointer(GroupAffinity)), uintptr(unsafe.Pointer(PreviousGroupAffinity)))
	return BOOL(ret)
}

func AttachThreadInput(idAttach uint32, idAttachTo uint32, fAttach BOOL) BOOL {
	addr := lazyAddr(&pAttachThreadInput, libUser32, "AttachThreadInput")
	ret, _, _ := syscall.SyscallN(addr, uintptr(idAttach), uintptr(idAttachTo), uintptr(fAttach))
	return BOOL(ret)
}

func WaitForInputIdle(hProcess HANDLE, dwMilliseconds uint32) uint32 {
	addr := lazyAddr(&pWaitForInputIdle, libUser32, "WaitForInputIdle")
	ret, _, _ := syscall.SyscallN(addr, hProcess, uintptr(dwMilliseconds))
	return uint32(ret)
}

func GetGuiResources(hProcess HANDLE, uiFlags GET_GUI_RESOURCES_FLAGS) (uint32, WIN32_ERROR) {
	addr := lazyAddr(&pGetGuiResources, libUser32, "GetGuiResources")
	ret, _, err := syscall.SyscallN(addr, hProcess, uintptr(uiFlags))
	return uint32(ret), WIN32_ERROR(err)
}

func IsImmersiveProcess(hProcess HANDLE) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pIsImmersiveProcess, libUser32, "IsImmersiveProcess")
	ret, _, err := syscall.SyscallN(addr, hProcess)
	return BOOL(ret), WIN32_ERROR(err)
}

func SetProcessRestrictionExemption(fEnableExemption BOOL) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pSetProcessRestrictionExemption, libUser32, "SetProcessRestrictionExemption")
	ret, _, err := syscall.SyscallN(addr, uintptr(fEnableExemption))
	return BOOL(ret), WIN32_ERROR(err)
}

func GetProcessAffinityMask(hProcess HANDLE, lpProcessAffinityMask *uintptr, lpSystemAffinityMask *uintptr) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pGetProcessAffinityMask, libKernel32, "GetProcessAffinityMask")
	ret, _, err := syscall.SyscallN(addr, hProcess, uintptr(unsafe.Pointer(lpProcessAffinityMask)), uintptr(unsafe.Pointer(lpSystemAffinityMask)))
	return BOOL(ret), WIN32_ERROR(err)
}

func SetProcessAffinityMask(hProcess HANDLE, dwProcessAffinityMask uintptr) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pSetProcessAffinityMask, libKernel32, "SetProcessAffinityMask")
	ret, _, err := syscall.SyscallN(addr, hProcess, dwProcessAffinityMask)
	return BOOL(ret), WIN32_ERROR(err)
}

func GetProcessIoCounters(hProcess HANDLE, lpIoCounters *IO_COUNTERS) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pGetProcessIoCounters, libKernel32, "GetProcessIoCounters")
	ret, _, err := syscall.SyscallN(addr, hProcess, uintptr(unsafe.Pointer(lpIoCounters)))
	return BOOL(ret), WIN32_ERROR(err)
}

func SwitchToFiber(lpFiber unsafe.Pointer) {
	addr := lazyAddr(&pSwitchToFiber, libKernel32, "SwitchToFiber")
	syscall.SyscallN(addr, uintptr(lpFiber))
}

func DeleteFiber(lpFiber unsafe.Pointer) {
	addr := lazyAddr(&pDeleteFiber, libKernel32, "DeleteFiber")
	syscall.SyscallN(addr, uintptr(lpFiber))
}

func ConvertFiberToThread() (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pConvertFiberToThread, libKernel32, "ConvertFiberToThread")
	ret, _, err := syscall.SyscallN(addr)
	return BOOL(ret), WIN32_ERROR(err)
}

func CreateFiberEx(dwStackCommitSize uintptr, dwStackReserveSize uintptr, dwFlags uint32, lpStartAddress LPFIBER_START_ROUTINE, lpParameter unsafe.Pointer) (unsafe.Pointer, WIN32_ERROR) {
	addr := lazyAddr(&pCreateFiberEx, libKernel32, "CreateFiberEx")
	ret, _, err := syscall.SyscallN(addr, dwStackCommitSize, dwStackReserveSize, uintptr(dwFlags), lpStartAddress, uintptr(lpParameter))
	return (unsafe.Pointer)(ret), WIN32_ERROR(err)
}

func ConvertThreadToFiberEx(lpParameter unsafe.Pointer, dwFlags uint32) (unsafe.Pointer, WIN32_ERROR) {
	addr := lazyAddr(&pConvertThreadToFiberEx, libKernel32, "ConvertThreadToFiberEx")
	ret, _, err := syscall.SyscallN(addr, uintptr(lpParameter), uintptr(dwFlags))
	return (unsafe.Pointer)(ret), WIN32_ERROR(err)
}

func CreateFiber(dwStackSize uintptr, lpStartAddress LPFIBER_START_ROUTINE, lpParameter unsafe.Pointer) (unsafe.Pointer, WIN32_ERROR) {
	addr := lazyAddr(&pCreateFiber, libKernel32, "CreateFiber")
	ret, _, err := syscall.SyscallN(addr, dwStackSize, lpStartAddress, uintptr(lpParameter))
	return (unsafe.Pointer)(ret), WIN32_ERROR(err)
}

func ConvertThreadToFiber(lpParameter unsafe.Pointer) (unsafe.Pointer, WIN32_ERROR) {
	addr := lazyAddr(&pConvertThreadToFiber, libKernel32, "ConvertThreadToFiber")
	ret, _, err := syscall.SyscallN(addr, uintptr(lpParameter))
	return (unsafe.Pointer)(ret), WIN32_ERROR(err)
}

func CreateUmsCompletionList(UmsCompletionList unsafe.Pointer) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pCreateUmsCompletionList, libKernel32, "CreateUmsCompletionList")
	ret, _, err := syscall.SyscallN(addr, uintptr(UmsCompletionList))
	return BOOL(ret), WIN32_ERROR(err)
}

func DequeueUmsCompletionListItems(UmsCompletionList unsafe.Pointer, WaitTimeOut uint32, UmsThreadList unsafe.Pointer) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pDequeueUmsCompletionListItems, libKernel32, "DequeueUmsCompletionListItems")
	ret, _, err := syscall.SyscallN(addr, uintptr(UmsCompletionList), uintptr(WaitTimeOut), uintptr(UmsThreadList))
	return BOOL(ret), WIN32_ERROR(err)
}

func GetUmsCompletionListEvent(UmsCompletionList unsafe.Pointer, UmsCompletionEvent *HANDLE) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pGetUmsCompletionListEvent, libKernel32, "GetUmsCompletionListEvent")
	ret, _, err := syscall.SyscallN(addr, uintptr(UmsCompletionList), uintptr(unsafe.Pointer(UmsCompletionEvent)))
	return BOOL(ret), WIN32_ERROR(err)
}

func ExecuteUmsThread(UmsThread unsafe.Pointer) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pExecuteUmsThread, libKernel32, "ExecuteUmsThread")
	ret, _, err := syscall.SyscallN(addr, uintptr(UmsThread))
	return BOOL(ret), WIN32_ERROR(err)
}

func UmsThreadYield(SchedulerParam unsafe.Pointer) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pUmsThreadYield, libKernel32, "UmsThreadYield")
	ret, _, err := syscall.SyscallN(addr, uintptr(SchedulerParam))
	return BOOL(ret), WIN32_ERROR(err)
}

func DeleteUmsCompletionList(UmsCompletionList unsafe.Pointer) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pDeleteUmsCompletionList, libKernel32, "DeleteUmsCompletionList")
	ret, _, err := syscall.SyscallN(addr, uintptr(UmsCompletionList))
	return BOOL(ret), WIN32_ERROR(err)
}

func GetCurrentUmsThread() (unsafe.Pointer, WIN32_ERROR) {
	addr := lazyAddr(&pGetCurrentUmsThread, libKernel32, "GetCurrentUmsThread")
	ret, _, err := syscall.SyscallN(addr)
	return (unsafe.Pointer)(ret), WIN32_ERROR(err)
}

func GetNextUmsListItem(UmsContext unsafe.Pointer) (unsafe.Pointer, WIN32_ERROR) {
	addr := lazyAddr(&pGetNextUmsListItem, libKernel32, "GetNextUmsListItem")
	ret, _, err := syscall.SyscallN(addr, uintptr(UmsContext))
	return (unsafe.Pointer)(ret), WIN32_ERROR(err)
}

func QueryUmsThreadInformation(UmsThread unsafe.Pointer, UmsThreadInfoClass RTL_UMS_THREAD_INFO_CLASS, UmsThreadInformation unsafe.Pointer, UmsThreadInformationLength uint32, ReturnLength *uint32) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pQueryUmsThreadInformation, libKernel32, "QueryUmsThreadInformation")
	ret, _, err := syscall.SyscallN(addr, uintptr(UmsThread), uintptr(UmsThreadInfoClass), uintptr(UmsThreadInformation), uintptr(UmsThreadInformationLength), uintptr(unsafe.Pointer(ReturnLength)))
	return BOOL(ret), WIN32_ERROR(err)
}

func SetUmsThreadInformation(UmsThread unsafe.Pointer, UmsThreadInfoClass RTL_UMS_THREAD_INFO_CLASS, UmsThreadInformation unsafe.Pointer, UmsThreadInformationLength uint32) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pSetUmsThreadInformation, libKernel32, "SetUmsThreadInformation")
	ret, _, err := syscall.SyscallN(addr, uintptr(UmsThread), uintptr(UmsThreadInfoClass), uintptr(UmsThreadInformation), uintptr(UmsThreadInformationLength))
	return BOOL(ret), WIN32_ERROR(err)
}

func DeleteUmsThreadContext(UmsThread unsafe.Pointer) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pDeleteUmsThreadContext, libKernel32, "DeleteUmsThreadContext")
	ret, _, err := syscall.SyscallN(addr, uintptr(UmsThread))
	return BOOL(ret), WIN32_ERROR(err)
}

func CreateUmsThreadContext(lpUmsThread unsafe.Pointer) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pCreateUmsThreadContext, libKernel32, "CreateUmsThreadContext")
	ret, _, err := syscall.SyscallN(addr, uintptr(lpUmsThread))
	return BOOL(ret), WIN32_ERROR(err)
}

func EnterUmsSchedulingMode(SchedulerStartupInfo *UMS_SCHEDULER_STARTUP_INFO) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pEnterUmsSchedulingMode, libKernel32, "EnterUmsSchedulingMode")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(SchedulerStartupInfo)))
	return BOOL(ret), WIN32_ERROR(err)
}

func GetUmsSystemThreadInformation(ThreadHandle HANDLE, SystemThreadInfo *UMS_SYSTEM_THREAD_INFORMATION) BOOL {
	addr := lazyAddr(&pGetUmsSystemThreadInformation, libKernel32, "GetUmsSystemThreadInformation")
	ret, _, _ := syscall.SyscallN(addr, ThreadHandle, uintptr(unsafe.Pointer(SystemThreadInfo)))
	return BOOL(ret)
}

func SetThreadAffinityMask(hThread HANDLE, dwThreadAffinityMask uintptr) (uintptr, WIN32_ERROR) {
	addr := lazyAddr(&pSetThreadAffinityMask, libKernel32, "SetThreadAffinityMask")
	ret, _, err := syscall.SyscallN(addr, hThread, dwThreadAffinityMask)
	return ret, WIN32_ERROR(err)
}

func SetProcessDEPPolicy(dwFlags PROCESS_DEP_FLAGS) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pSetProcessDEPPolicy, libKernel32, "SetProcessDEPPolicy")
	ret, _, err := syscall.SyscallN(addr, uintptr(dwFlags))
	return BOOL(ret), WIN32_ERROR(err)
}

func GetProcessDEPPolicy(hProcess HANDLE, lpFlags *uint32, lpPermanent *BOOL) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pGetProcessDEPPolicy, libKernel32, "GetProcessDEPPolicy")
	ret, _, err := syscall.SyscallN(addr, hProcess, uintptr(unsafe.Pointer(lpFlags)), uintptr(unsafe.Pointer(lpPermanent)))
	return BOOL(ret), WIN32_ERROR(err)
}

func PulseEvent(hEvent HANDLE) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pPulseEvent, libKernel32, "PulseEvent")
	ret, _, err := syscall.SyscallN(addr, hEvent)
	return BOOL(ret), WIN32_ERROR(err)
}

func WinExec(lpCmdLine PSTR, uCmdShow uint32) uint32 {
	addr := lazyAddr(&pWinExec, libKernel32, "WinExec")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpCmdLine)), uintptr(uCmdShow))
	return uint32(ret)
}

func CreateSemaphoreA(lpSemaphoreAttributes *SECURITY_ATTRIBUTES, lInitialCount int32, lMaximumCount int32, lpName PSTR) (HANDLE, WIN32_ERROR) {
	addr := lazyAddr(&pCreateSemaphoreA, libKernel32, "CreateSemaphoreA")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpSemaphoreAttributes)), uintptr(lInitialCount), uintptr(lMaximumCount), uintptr(unsafe.Pointer(lpName)))
	return ret, WIN32_ERROR(err)
}

func CreateSemaphoreExA(lpSemaphoreAttributes *SECURITY_ATTRIBUTES, lInitialCount int32, lMaximumCount int32, lpName PSTR, dwFlags uint32, dwDesiredAccess uint32) (HANDLE, WIN32_ERROR) {
	addr := lazyAddr(&pCreateSemaphoreExA, libKernel32, "CreateSemaphoreExA")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpSemaphoreAttributes)), uintptr(lInitialCount), uintptr(lMaximumCount), uintptr(unsafe.Pointer(lpName)), uintptr(dwFlags), uintptr(dwDesiredAccess))
	return ret, WIN32_ERROR(err)
}

func QueryFullProcessImageNameA(hProcess HANDLE, dwFlags PROCESS_NAME_FORMAT, lpExeName PSTR, lpdwSize *uint32) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pQueryFullProcessImageNameA, libKernel32, "QueryFullProcessImageNameA")
	ret, _, err := syscall.SyscallN(addr, hProcess, uintptr(dwFlags), uintptr(unsafe.Pointer(lpExeName)), uintptr(unsafe.Pointer(lpdwSize)))
	return BOOL(ret), WIN32_ERROR(err)
}

var QueryFullProcessImageName = QueryFullProcessImageNameW

func QueryFullProcessImageNameW(hProcess HANDLE, dwFlags PROCESS_NAME_FORMAT, lpExeName PWSTR, lpdwSize *uint32) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pQueryFullProcessImageNameW, libKernel32, "QueryFullProcessImageNameW")
	ret, _, err := syscall.SyscallN(addr, hProcess, uintptr(dwFlags), uintptr(unsafe.Pointer(lpExeName)), uintptr(unsafe.Pointer(lpdwSize)))
	return BOOL(ret), WIN32_ERROR(err)
}

func GetStartupInfoA(lpStartupInfo *STARTUPINFOA) {
	addr := lazyAddr(&pGetStartupInfoA, libKernel32, "GetStartupInfoA")
	syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpStartupInfo)))
}

func CreateProcessWithLogonW(lpUsername PWSTR, lpDomain PWSTR, lpPassword PWSTR, dwLogonFlags CREATE_PROCESS_LOGON_FLAGS, lpApplicationName PWSTR, lpCommandLine PWSTR, dwCreationFlags uint32, lpEnvironment unsafe.Pointer, lpCurrentDirectory PWSTR, lpStartupInfo *STARTUPINFOW, lpProcessInformation *PROCESS_INFORMATION) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pCreateProcessWithLogonW, libAdvapi32, "CreateProcessWithLogonW")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpUsername)), uintptr(unsafe.Pointer(lpDomain)), uintptr(unsafe.Pointer(lpPassword)), uintptr(dwLogonFlags), uintptr(unsafe.Pointer(lpApplicationName)), uintptr(unsafe.Pointer(lpCommandLine)), uintptr(dwCreationFlags), uintptr(lpEnvironment), uintptr(unsafe.Pointer(lpCurrentDirectory)), uintptr(unsafe.Pointer(lpStartupInfo)), uintptr(unsafe.Pointer(lpProcessInformation)))
	return BOOL(ret), WIN32_ERROR(err)
}

func CreateProcessWithTokenW(hToken HANDLE, dwLogonFlags CREATE_PROCESS_LOGON_FLAGS, lpApplicationName PWSTR, lpCommandLine PWSTR, dwCreationFlags uint32, lpEnvironment unsafe.Pointer, lpCurrentDirectory PWSTR, lpStartupInfo *STARTUPINFOW, lpProcessInformation *PROCESS_INFORMATION) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pCreateProcessWithTokenW, libAdvapi32, "CreateProcessWithTokenW")
	ret, _, err := syscall.SyscallN(addr, hToken, uintptr(dwLogonFlags), uintptr(unsafe.Pointer(lpApplicationName)), uintptr(unsafe.Pointer(lpCommandLine)), uintptr(dwCreationFlags), uintptr(lpEnvironment), uintptr(unsafe.Pointer(lpCurrentDirectory)), uintptr(unsafe.Pointer(lpStartupInfo)), uintptr(unsafe.Pointer(lpProcessInformation)))
	return BOOL(ret), WIN32_ERROR(err)
}

func RegisterWaitForSingleObject(phNewWaitObject *HANDLE, hObject HANDLE, Callback WAITORTIMERCALLBACK, Context unsafe.Pointer, dwMilliseconds uint32, dwFlags WORKER_THREAD_FLAGS) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pRegisterWaitForSingleObject, libKernel32, "RegisterWaitForSingleObject")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(phNewWaitObject)), hObject, Callback, uintptr(Context), uintptr(dwMilliseconds), uintptr(dwFlags))
	return BOOL(ret), WIN32_ERROR(err)
}

func UnregisterWait(WaitHandle HANDLE) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pUnregisterWait, libKernel32, "UnregisterWait")
	ret, _, err := syscall.SyscallN(addr, WaitHandle)
	return BOOL(ret), WIN32_ERROR(err)
}

func SetTimerQueueTimer(TimerQueue HANDLE, Callback WAITORTIMERCALLBACK, Parameter unsafe.Pointer, DueTime uint32, Period uint32, PreferIo BOOL) HANDLE {
	addr := lazyAddr(&pSetTimerQueueTimer, libKernel32, "SetTimerQueueTimer")
	ret, _, _ := syscall.SyscallN(addr, TimerQueue, Callback, uintptr(Parameter), uintptr(DueTime), uintptr(Period), uintptr(PreferIo))
	return ret
}

func CreatePrivateNamespaceA(lpPrivateNamespaceAttributes *SECURITY_ATTRIBUTES, lpBoundaryDescriptor unsafe.Pointer, lpAliasPrefix PSTR) (NamespaceHandle, WIN32_ERROR) {
	addr := lazyAddr(&pCreatePrivateNamespaceA, libKernel32, "CreatePrivateNamespaceA")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpPrivateNamespaceAttributes)), uintptr(lpBoundaryDescriptor), uintptr(unsafe.Pointer(lpAliasPrefix)))
	return ret, WIN32_ERROR(err)
}

func OpenPrivateNamespaceA(lpBoundaryDescriptor unsafe.Pointer, lpAliasPrefix PSTR) NamespaceHandle {
	addr := lazyAddr(&pOpenPrivateNamespaceA, libKernel32, "OpenPrivateNamespaceA")
	ret, _, _ := syscall.SyscallN(addr, uintptr(lpBoundaryDescriptor), uintptr(unsafe.Pointer(lpAliasPrefix)))
	return ret
}

func CreateBoundaryDescriptorA(Name PSTR, Flags uint32) (BoundaryDescriptorHandle, WIN32_ERROR) {
	addr := lazyAddr(&pCreateBoundaryDescriptorA, libKernel32, "CreateBoundaryDescriptorA")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(Name)), uintptr(Flags))
	return ret, WIN32_ERROR(err)
}

func AddIntegrityLabelToBoundaryDescriptor(BoundaryDescriptor *HANDLE, IntegrityLabel PSID) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pAddIntegrityLabelToBoundaryDescriptor, libKernel32, "AddIntegrityLabelToBoundaryDescriptor")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(BoundaryDescriptor)), uintptr(unsafe.Pointer(IntegrityLabel)))
	return BOOL(ret), WIN32_ERROR(err)
}

func GetActiveProcessorGroupCount() uint16 {
	addr := lazyAddr(&pGetActiveProcessorGroupCount, libKernel32, "GetActiveProcessorGroupCount")
	ret, _, _ := syscall.SyscallN(addr)
	return uint16(ret)
}

func GetMaximumProcessorGroupCount() uint16 {
	addr := lazyAddr(&pGetMaximumProcessorGroupCount, libKernel32, "GetMaximumProcessorGroupCount")
	ret, _, _ := syscall.SyscallN(addr)
	return uint16(ret)
}

func GetActiveProcessorCount(GroupNumber uint16) (uint32, WIN32_ERROR) {
	addr := lazyAddr(&pGetActiveProcessorCount, libKernel32, "GetActiveProcessorCount")
	ret, _, err := syscall.SyscallN(addr, uintptr(GroupNumber))
	return uint32(ret), WIN32_ERROR(err)
}

func GetMaximumProcessorCount(GroupNumber uint16) (uint32, WIN32_ERROR) {
	addr := lazyAddr(&pGetMaximumProcessorCount, libKernel32, "GetMaximumProcessorCount")
	ret, _, err := syscall.SyscallN(addr, uintptr(GroupNumber))
	return uint32(ret), WIN32_ERROR(err)
}

func GetNumaProcessorNode(Processor byte, NodeNumber *byte) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pGetNumaProcessorNode, libKernel32, "GetNumaProcessorNode")
	ret, _, err := syscall.SyscallN(addr, uintptr(Processor), uintptr(unsafe.Pointer(NodeNumber)))
	return BOOL(ret), WIN32_ERROR(err)
}

func GetNumaNodeNumberFromHandle(hFile HANDLE, NodeNumber *uint16) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pGetNumaNodeNumberFromHandle, libKernel32, "GetNumaNodeNumberFromHandle")
	ret, _, err := syscall.SyscallN(addr, hFile, uintptr(unsafe.Pointer(NodeNumber)))
	return BOOL(ret), WIN32_ERROR(err)
}

func GetNumaProcessorNodeEx(Processor *PROCESSOR_NUMBER, NodeNumber *uint16) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pGetNumaProcessorNodeEx, libKernel32, "GetNumaProcessorNodeEx")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(Processor)), uintptr(unsafe.Pointer(NodeNumber)))
	return BOOL(ret), WIN32_ERROR(err)
}

func GetNumaNodeProcessorMask(Node byte, ProcessorMask *uint64) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pGetNumaNodeProcessorMask, libKernel32, "GetNumaNodeProcessorMask")
	ret, _, err := syscall.SyscallN(addr, uintptr(Node), uintptr(unsafe.Pointer(ProcessorMask)))
	return BOOL(ret), WIN32_ERROR(err)
}

func GetNumaAvailableMemoryNode(Node byte, AvailableBytes *uint64) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pGetNumaAvailableMemoryNode, libKernel32, "GetNumaAvailableMemoryNode")
	ret, _, err := syscall.SyscallN(addr, uintptr(Node), uintptr(unsafe.Pointer(AvailableBytes)))
	return BOOL(ret), WIN32_ERROR(err)
}

func GetNumaAvailableMemoryNodeEx(Node uint16, AvailableBytes *uint64) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pGetNumaAvailableMemoryNodeEx, libKernel32, "GetNumaAvailableMemoryNodeEx")
	ret, _, err := syscall.SyscallN(addr, uintptr(Node), uintptr(unsafe.Pointer(AvailableBytes)))
	return BOOL(ret), WIN32_ERROR(err)
}

func GetNumaProximityNode(ProximityId uint32, NodeNumber *byte) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pGetNumaProximityNode, libKernel32, "GetNumaProximityNode")
	ret, _, err := syscall.SyscallN(addr, uintptr(ProximityId), uintptr(unsafe.Pointer(NodeNumber)))
	return BOOL(ret), WIN32_ERROR(err)
}
