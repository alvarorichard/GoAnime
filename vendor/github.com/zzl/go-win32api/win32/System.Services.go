package win32

import (
	"syscall"
	"unsafe"
)

type (
	SERVICE_STATUS_HANDLE = uintptr
)

const (
	SERVICE_ALL_ACCESS                                  uint32 = 0xf01ff
	SC_MANAGER_ALL_ACCESS                               uint32 = 0xf003f
	SERVICES_ACTIVE_DATABASEW                           string = "ServicesActive"
	SERVICES_FAILED_DATABASEW                           string = "ServicesFailed"
	SERVICES_ACTIVE_DATABASEA                           string = "ServicesActive"
	SERVICES_FAILED_DATABASEA                           string = "ServicesFailed"
	SERVICES_ACTIVE_DATABASE                            string = "ServicesActive"
	SERVICES_FAILED_DATABASE                            string = "ServicesFailed"
	SERVICE_NO_CHANGE                                   uint32 = 0xffffffff
	SERVICE_CONTROL_STOP                                uint32 = 0x1
	SERVICE_CONTROL_PAUSE                               uint32 = 0x2
	SERVICE_CONTROL_CONTINUE                            uint32 = 0x3
	SERVICE_CONTROL_INTERROGATE                         uint32 = 0x4
	SERVICE_CONTROL_SHUTDOWN                            uint32 = 0x5
	SERVICE_CONTROL_PARAMCHANGE                         uint32 = 0x6
	SERVICE_CONTROL_NETBINDADD                          uint32 = 0x7
	SERVICE_CONTROL_NETBINDREMOVE                       uint32 = 0x8
	SERVICE_CONTROL_NETBINDENABLE                       uint32 = 0x9
	SERVICE_CONTROL_NETBINDDISABLE                      uint32 = 0xa
	SERVICE_CONTROL_DEVICEEVENT                         uint32 = 0xb
	SERVICE_CONTROL_HARDWAREPROFILECHANGE               uint32 = 0xc
	SERVICE_CONTROL_POWEREVENT                          uint32 = 0xd
	SERVICE_CONTROL_SESSIONCHANGE                       uint32 = 0xe
	SERVICE_CONTROL_PRESHUTDOWN                         uint32 = 0xf
	SERVICE_CONTROL_TIMECHANGE                          uint32 = 0x10
	SERVICE_CONTROL_TRIGGEREVENT                        uint32 = 0x20
	SERVICE_CONTROL_LOWRESOURCES                        uint32 = 0x60
	SERVICE_CONTROL_SYSTEMLOWRESOURCES                  uint32 = 0x61
	SERVICE_ACCEPT_STOP                                 uint32 = 0x1
	SERVICE_ACCEPT_PAUSE_CONTINUE                       uint32 = 0x2
	SERVICE_ACCEPT_SHUTDOWN                             uint32 = 0x4
	SERVICE_ACCEPT_PARAMCHANGE                          uint32 = 0x8
	SERVICE_ACCEPT_NETBINDCHANGE                        uint32 = 0x10
	SERVICE_ACCEPT_HARDWAREPROFILECHANGE                uint32 = 0x20
	SERVICE_ACCEPT_POWEREVENT                           uint32 = 0x40
	SERVICE_ACCEPT_SESSIONCHANGE                        uint32 = 0x80
	SERVICE_ACCEPT_PRESHUTDOWN                          uint32 = 0x100
	SERVICE_ACCEPT_TIMECHANGE                           uint32 = 0x200
	SERVICE_ACCEPT_TRIGGEREVENT                         uint32 = 0x400
	SERVICE_ACCEPT_USER_LOGOFF                          uint32 = 0x800
	SERVICE_ACCEPT_LOWRESOURCES                         uint32 = 0x2000
	SERVICE_ACCEPT_SYSTEMLOWRESOURCES                   uint32 = 0x4000
	SC_MANAGER_CONNECT                                  uint32 = 0x1
	SC_MANAGER_CREATE_SERVICE                           uint32 = 0x2
	SC_MANAGER_ENUMERATE_SERVICE                        uint32 = 0x4
	SC_MANAGER_LOCK                                     uint32 = 0x8
	SC_MANAGER_QUERY_LOCK_STATUS                        uint32 = 0x10
	SC_MANAGER_MODIFY_BOOT_CONFIG                       uint32 = 0x20
	SERVICE_QUERY_CONFIG                                uint32 = 0x1
	SERVICE_CHANGE_CONFIG                               uint32 = 0x2
	SERVICE_QUERY_STATUS                                uint32 = 0x4
	SERVICE_ENUMERATE_DEPENDENTS                        uint32 = 0x8
	SERVICE_START                                       uint32 = 0x10
	SERVICE_STOP                                        uint32 = 0x20
	SERVICE_PAUSE_CONTINUE                              uint32 = 0x40
	SERVICE_INTERROGATE                                 uint32 = 0x80
	SERVICE_USER_DEFINED_CONTROL                        uint32 = 0x100
	SERVICE_NOTIFY_STATUS_CHANGE_1                      uint32 = 0x1
	SERVICE_NOTIFY_STATUS_CHANGE_2                      uint32 = 0x2
	SERVICE_NOTIFY_STATUS_CHANGE                        uint32 = 0x2
	SERVICE_STOP_REASON_FLAG_MIN                        uint32 = 0x0
	SERVICE_STOP_REASON_FLAG_UNPLANNED                  uint32 = 0x10000000
	SERVICE_STOP_REASON_FLAG_CUSTOM                     uint32 = 0x20000000
	SERVICE_STOP_REASON_FLAG_PLANNED                    uint32 = 0x40000000
	SERVICE_STOP_REASON_FLAG_MAX                        uint32 = 0x80000000
	SERVICE_STOP_REASON_MAJOR_MIN                       uint32 = 0x0
	SERVICE_STOP_REASON_MAJOR_OTHER                     uint32 = 0x10000
	SERVICE_STOP_REASON_MAJOR_HARDWARE                  uint32 = 0x20000
	SERVICE_STOP_REASON_MAJOR_OPERATINGSYSTEM           uint32 = 0x30000
	SERVICE_STOP_REASON_MAJOR_SOFTWARE                  uint32 = 0x40000
	SERVICE_STOP_REASON_MAJOR_APPLICATION               uint32 = 0x50000
	SERVICE_STOP_REASON_MAJOR_NONE                      uint32 = 0x60000
	SERVICE_STOP_REASON_MAJOR_MAX                       uint32 = 0x70000
	SERVICE_STOP_REASON_MAJOR_MIN_CUSTOM                uint32 = 0x400000
	SERVICE_STOP_REASON_MAJOR_MAX_CUSTOM                uint32 = 0xff0000
	SERVICE_STOP_REASON_MINOR_MIN                       uint32 = 0x0
	SERVICE_STOP_REASON_MINOR_OTHER                     uint32 = 0x1
	SERVICE_STOP_REASON_MINOR_MAINTENANCE               uint32 = 0x2
	SERVICE_STOP_REASON_MINOR_INSTALLATION              uint32 = 0x3
	SERVICE_STOP_REASON_MINOR_UPGRADE                   uint32 = 0x4
	SERVICE_STOP_REASON_MINOR_RECONFIG                  uint32 = 0x5
	SERVICE_STOP_REASON_MINOR_HUNG                      uint32 = 0x6
	SERVICE_STOP_REASON_MINOR_UNSTABLE                  uint32 = 0x7
	SERVICE_STOP_REASON_MINOR_DISK                      uint32 = 0x8
	SERVICE_STOP_REASON_MINOR_NETWORKCARD               uint32 = 0x9
	SERVICE_STOP_REASON_MINOR_ENVIRONMENT               uint32 = 0xa
	SERVICE_STOP_REASON_MINOR_HARDWARE_DRIVER           uint32 = 0xb
	SERVICE_STOP_REASON_MINOR_OTHERDRIVER               uint32 = 0xc
	SERVICE_STOP_REASON_MINOR_SERVICEPACK               uint32 = 0xd
	SERVICE_STOP_REASON_MINOR_SOFTWARE_UPDATE           uint32 = 0xe
	SERVICE_STOP_REASON_MINOR_SECURITYFIX               uint32 = 0xf
	SERVICE_STOP_REASON_MINOR_SECURITY                  uint32 = 0x10
	SERVICE_STOP_REASON_MINOR_NETWORK_CONNECTIVITY      uint32 = 0x11
	SERVICE_STOP_REASON_MINOR_WMI                       uint32 = 0x12
	SERVICE_STOP_REASON_MINOR_SERVICEPACK_UNINSTALL     uint32 = 0x13
	SERVICE_STOP_REASON_MINOR_SOFTWARE_UPDATE_UNINSTALL uint32 = 0x14
	SERVICE_STOP_REASON_MINOR_SECURITYFIX_UNINSTALL     uint32 = 0x15
	SERVICE_STOP_REASON_MINOR_MMC                       uint32 = 0x16
	SERVICE_STOP_REASON_MINOR_NONE                      uint32 = 0x17
	SERVICE_STOP_REASON_MINOR_MEMOTYLIMIT               uint32 = 0x18
	SERVICE_STOP_REASON_MINOR_MAX                       uint32 = 0x19
	SERVICE_STOP_REASON_MINOR_MIN_CUSTOM                uint32 = 0x100
	SERVICE_STOP_REASON_MINOR_MAX_CUSTOM                uint32 = 0xffff
	SERVICE_CONTROL_STATUS_REASON_INFO                  uint32 = 0x1
	SERVICE_SID_TYPE_NONE                               uint32 = 0x0
	SERVICE_SID_TYPE_UNRESTRICTED                       uint32 = 0x1
	SERVICE_TRIGGER_TYPE_CUSTOM_SYSTEM_STATE_CHANGE     uint32 = 0x7
	SERVICE_TRIGGER_TYPE_AGGREGATE                      uint32 = 0x1e
	SERVICE_START_REASON_DEMAND                         uint32 = 0x1
	SERVICE_START_REASON_AUTO                           uint32 = 0x2
	SERVICE_START_REASON_TRIGGER                        uint32 = 0x4
	SERVICE_START_REASON_RESTART_ON_FAILURE             uint32 = 0x8
	SERVICE_START_REASON_DELAYEDAUTO                    uint32 = 0x10
	SERVICE_DYNAMIC_INFORMATION_LEVEL_START_REASON      uint32 = 0x1
	SERVICE_LAUNCH_PROTECTED_NONE                       uint32 = 0x0
	SERVICE_LAUNCH_PROTECTED_WINDOWS                    uint32 = 0x1
	SERVICE_LAUNCH_PROTECTED_WINDOWS_LIGHT              uint32 = 0x2
	SERVICE_LAUNCH_PROTECTED_ANTIMALWARE_LIGHT          uint32 = 0x3
	SERVICE_TRIGGER_STARTED_ARGUMENT                    string = "TriggerStarted"
	SC_AGGREGATE_STORAGE_KEY                            string = "System\\CurrentControlSet\\Control\\ServiceAggregatedEvents"
)

var (
	NETWORK_MANAGER_FIRST_IP_ADDRESS_ARRIVAL_GUID = syscall.GUID{0x4F27F2DE, 0x14E2, 0x430B,
		[8]byte{0xA5, 0x49, 0x7C, 0xD4, 0x8C, 0xBC, 0x82, 0x45}}

	NETWORK_MANAGER_LAST_IP_ADDRESS_REMOVAL_GUID = syscall.GUID{0xCC4BA62A, 0x162E, 0x4648,
		[8]byte{0x84, 0x7A, 0xB6, 0xBD, 0xF9, 0x93, 0xE3, 0x35}}

	DOMAIN_JOIN_GUID = syscall.GUID{0x1CE20ABA, 0x9851, 0x4421,
		[8]byte{0x94, 0x30, 0x1D, 0xDE, 0xB7, 0x66, 0xE8, 0x09}}

	DOMAIN_LEAVE_GUID = syscall.GUID{0xDDAF516E, 0x58C2, 0x4866,
		[8]byte{0x95, 0x74, 0xC3, 0xB6, 0x15, 0xD4, 0x2E, 0xA1}}

	FIREWALL_PORT_OPEN_GUID = syscall.GUID{0xB7569E07, 0x8421, 0x4EE0,
		[8]byte{0xAD, 0x10, 0x86, 0x91, 0x5A, 0xFD, 0xAD, 0x09}}

	FIREWALL_PORT_CLOSE_GUID = syscall.GUID{0xA144ED38, 0x8E12, 0x4DE4,
		[8]byte{0x9D, 0x96, 0xE6, 0x47, 0x40, 0xB1, 0xA5, 0x24}}

	MACHINE_POLICY_PRESENT_GUID = syscall.GUID{0x659FCAE6, 0x5BDB, 0x4DA9,
		[8]byte{0xB1, 0xFF, 0xCA, 0x2A, 0x17, 0x8D, 0x46, 0xE0}}

	USER_POLICY_PRESENT_GUID = syscall.GUID{0x54FB46C8, 0xF089, 0x464C,
		[8]byte{0xB1, 0xFD, 0x59, 0xD1, 0xB6, 0x2C, 0x3B, 0x50}}

	RPC_INTERFACE_EVENT_GUID = syscall.GUID{0xBC90D167, 0x9470, 0x4139,
		[8]byte{0xA9, 0xBA, 0xBE, 0x0B, 0xBB, 0xF5, 0xB7, 0x4D}}

	NAMED_PIPE_EVENT_GUID = syscall.GUID{0x1F81D131, 0x3FAC, 0x4537,
		[8]byte{0x9E, 0x0C, 0x7E, 0x7B, 0x0C, 0x2F, 0x4B, 0x55}}

	CUSTOM_SYSTEM_STATE_CHANGE_EVENT_GUID = syscall.GUID{0x2D7A2816, 0x0C5E, 0x45FC,
		[8]byte{0x9C, 0xE7, 0x57, 0x0E, 0x5E, 0xCD, 0xE9, 0xC9}}
)

// enums

// enum
type ENUM_SERVICE_STATE uint32

const (
	SERVICE_ACTIVE    ENUM_SERVICE_STATE = 1
	SERVICE_INACTIVE  ENUM_SERVICE_STATE = 2
	SERVICE_STATE_ALL ENUM_SERVICE_STATE = 3
)

// enum
type SERVICE_ERROR uint32

const (
	SERVICE_ERROR_CRITICAL SERVICE_ERROR = 3
	SERVICE_ERROR_IGNORE   SERVICE_ERROR = 0
	SERVICE_ERROR_NORMAL   SERVICE_ERROR = 1
	SERVICE_ERROR_SEVERE   SERVICE_ERROR = 2
)

// enum
type SERVICE_CONFIG uint32

const (
	SERVICE_CONFIG_DELAYED_AUTO_START_INFO  SERVICE_CONFIG = 3
	SERVICE_CONFIG_DESCRIPTION              SERVICE_CONFIG = 1
	SERVICE_CONFIG_FAILURE_ACTIONS          SERVICE_CONFIG = 2
	SERVICE_CONFIG_FAILURE_ACTIONS_FLAG     SERVICE_CONFIG = 4
	SERVICE_CONFIG_PREFERRED_NODE           SERVICE_CONFIG = 9
	SERVICE_CONFIG_PRESHUTDOWN_INFO         SERVICE_CONFIG = 7
	SERVICE_CONFIG_REQUIRED_PRIVILEGES_INFO SERVICE_CONFIG = 6
	SERVICE_CONFIG_SERVICE_SID_INFO         SERVICE_CONFIG = 5
	SERVICE_CONFIG_TRIGGER_INFO             SERVICE_CONFIG = 8
	SERVICE_CONFIG_LAUNCH_PROTECTED         SERVICE_CONFIG = 12
)

// enum
// flags
type ENUM_SERVICE_TYPE uint32

const (
	SERVICE_DRIVER              ENUM_SERVICE_TYPE = 11
	SERVICE_KERNEL_DRIVER       ENUM_SERVICE_TYPE = 1
	SERVICE_WIN32               ENUM_SERVICE_TYPE = 48
	SERVICE_WIN32_SHARE_PROCESS ENUM_SERVICE_TYPE = 32
	SERVICE_ADAPTER             ENUM_SERVICE_TYPE = 4
	SERVICE_FILE_SYSTEM_DRIVER  ENUM_SERVICE_TYPE = 2
	SERVICE_RECOGNIZER_DRIVER   ENUM_SERVICE_TYPE = 8
	SERVICE_WIN32_OWN_PROCESS   ENUM_SERVICE_TYPE = 16
	SERVICE_USER_OWN_PROCESS    ENUM_SERVICE_TYPE = 80
	SERVICE_USER_SHARE_PROCESS  ENUM_SERVICE_TYPE = 96
)

// enum
type SERVICE_START_TYPE uint32

const (
	SERVICE_AUTO_START   SERVICE_START_TYPE = 2
	SERVICE_BOOT_START   SERVICE_START_TYPE = 0
	SERVICE_DEMAND_START SERVICE_START_TYPE = 3
	SERVICE_DISABLED     SERVICE_START_TYPE = 4
	SERVICE_SYSTEM_START SERVICE_START_TYPE = 1
)

// enum
// flags
type SERVICE_NOTIFY uint32

const (
	SERVICE_NOTIFY_CREATED          SERVICE_NOTIFY = 128
	SERVICE_NOTIFY_CONTINUE_PENDING SERVICE_NOTIFY = 16
	SERVICE_NOTIFY_DELETE_PENDING   SERVICE_NOTIFY = 512
	SERVICE_NOTIFY_DELETED          SERVICE_NOTIFY = 256
	SERVICE_NOTIFY_PAUSE_PENDING    SERVICE_NOTIFY = 32
	SERVICE_NOTIFY_PAUSED           SERVICE_NOTIFY = 64
	SERVICE_NOTIFY_RUNNING          SERVICE_NOTIFY = 8
	SERVICE_NOTIFY_START_PENDING    SERVICE_NOTIFY = 2
	SERVICE_NOTIFY_STOP_PENDING     SERVICE_NOTIFY = 4
	SERVICE_NOTIFY_STOPPED          SERVICE_NOTIFY = 1
)

// enum
type SERVICE_RUNS_IN_PROCESS uint32

const (
	SERVICE_RUNS_IN_NON_SYSTEM_OR_NOT_RUNNING SERVICE_RUNS_IN_PROCESS = 0
	SERVICE_RUNS_IN_SYSTEM_PROCESS            SERVICE_RUNS_IN_PROCESS = 1
)

// enum
type SERVICE_TRIGGER_ACTION uint32

const (
	SERVICE_TRIGGER_ACTION_SERVICE_START SERVICE_TRIGGER_ACTION = 1
	SERVICE_TRIGGER_ACTION_SERVICE_STOP  SERVICE_TRIGGER_ACTION = 2
)

// enum
type SERVICE_TRIGGER_TYPE uint32

const (
	SERVICE_TRIGGER_TYPE_CUSTOM                   SERVICE_TRIGGER_TYPE = 20
	SERVICE_TRIGGER_TYPE_DEVICE_INTERFACE_ARRIVAL SERVICE_TRIGGER_TYPE = 1
	SERVICE_TRIGGER_TYPE_DOMAIN_JOIN              SERVICE_TRIGGER_TYPE = 3
	SERVICE_TRIGGER_TYPE_FIREWALL_PORT_EVENT      SERVICE_TRIGGER_TYPE = 4
	SERVICE_TRIGGER_TYPE_GROUP_POLICY             SERVICE_TRIGGER_TYPE = 5
	SERVICE_TRIGGER_TYPE_IP_ADDRESS_AVAILABILITY  SERVICE_TRIGGER_TYPE = 2
	SERVICE_TRIGGER_TYPE_NETWORK_ENDPOINT         SERVICE_TRIGGER_TYPE = 6
)

// enum
type SERVICE_TRIGGER_SPECIFIC_DATA_ITEM_DATA_TYPE uint32

const (
	SERVICE_TRIGGER_DATA_TYPE_BINARY      SERVICE_TRIGGER_SPECIFIC_DATA_ITEM_DATA_TYPE = 1
	SERVICE_TRIGGER_DATA_TYPE_STRING      SERVICE_TRIGGER_SPECIFIC_DATA_ITEM_DATA_TYPE = 2
	SERVICE_TRIGGER_DATA_TYPE_LEVEL       SERVICE_TRIGGER_SPECIFIC_DATA_ITEM_DATA_TYPE = 3
	SERVICE_TRIGGER_DATA_TYPE_KEYWORD_ANY SERVICE_TRIGGER_SPECIFIC_DATA_ITEM_DATA_TYPE = 4
	SERVICE_TRIGGER_DATA_TYPE_KEYWORD_ALL SERVICE_TRIGGER_SPECIFIC_DATA_ITEM_DATA_TYPE = 5
)

// enum
type SERVICE_STATUS_CURRENT_STATE uint32

const (
	SERVICE_CONTINUE_PENDING SERVICE_STATUS_CURRENT_STATE = 5
	SERVICE_PAUSE_PENDING    SERVICE_STATUS_CURRENT_STATE = 6
	SERVICE_PAUSED           SERVICE_STATUS_CURRENT_STATE = 7
	SERVICE_RUNNING          SERVICE_STATUS_CURRENT_STATE = 4
	SERVICE_START_PENDING    SERVICE_STATUS_CURRENT_STATE = 2
	SERVICE_STOP_PENDING     SERVICE_STATUS_CURRENT_STATE = 3
	SERVICE_STOPPED          SERVICE_STATUS_CURRENT_STATE = 1
)

// enum
type SC_ACTION_TYPE int32

const (
	SC_ACTION_NONE        SC_ACTION_TYPE = 0
	SC_ACTION_RESTART     SC_ACTION_TYPE = 1
	SC_ACTION_REBOOT      SC_ACTION_TYPE = 2
	SC_ACTION_RUN_COMMAND SC_ACTION_TYPE = 3
	SC_ACTION_OWN_RESTART SC_ACTION_TYPE = 4
)

// enum
type SC_STATUS_TYPE int32

const (
	SC_STATUS_PROCESS_INFO SC_STATUS_TYPE = 0
)

// enum
type SC_ENUM_TYPE int32

const (
	SC_ENUM_PROCESS_INFO SC_ENUM_TYPE = 0
)

// enum
type SC_EVENT_TYPE int32

const (
	SC_EVENT_DATABASE_CHANGE SC_EVENT_TYPE = 0
	SC_EVENT_PROPERTY_CHANGE SC_EVENT_TYPE = 1
	SC_EVENT_STATUS_CHANGE   SC_EVENT_TYPE = 2
)

// enum
type SERVICE_REGISTRY_STATE_TYPE int32

const (
	ServiceRegistryStateParameters SERVICE_REGISTRY_STATE_TYPE = 0
	ServiceRegistryStatePersistent SERVICE_REGISTRY_STATE_TYPE = 1
	MaxServiceRegistryStateType    SERVICE_REGISTRY_STATE_TYPE = 2
)

// enum
type SERVICE_DIRECTORY_TYPE int32

const (
	ServiceDirectoryPersistentState SERVICE_DIRECTORY_TYPE = 0
	ServiceDirectoryTypeMax         SERVICE_DIRECTORY_TYPE = 1
)

// enum
type SERVICE_SHARED_REGISTRY_STATE_TYPE int32

const (
	ServiceSharedRegistryPersistentState SERVICE_SHARED_REGISTRY_STATE_TYPE = 0
)

// enum
type SERVICE_SHARED_DIRECTORY_TYPE int32

const (
	ServiceSharedDirectoryPersistentState SERVICE_SHARED_DIRECTORY_TYPE = 0
)

// structs

type SERVICE_TRIGGER_CUSTOM_STATE_ID struct {
	Data [2]uint32
}

type SERVICE_CUSTOM_SYSTEM_STATE_CHANGE_DATA_ITEM_U_S struct {
	DataOffset uint32
	Data       [1]byte
}

type SERVICE_CUSTOM_SYSTEM_STATE_CHANGE_DATA_ITEM_U struct {
	Data [2]uint32
}

func (this *SERVICE_CUSTOM_SYSTEM_STATE_CHANGE_DATA_ITEM_U) CustomStateId() *SERVICE_TRIGGER_CUSTOM_STATE_ID {
	return (*SERVICE_TRIGGER_CUSTOM_STATE_ID)(unsafe.Pointer(this))
}

func (this *SERVICE_CUSTOM_SYSTEM_STATE_CHANGE_DATA_ITEM_U) CustomStateIdVal() SERVICE_TRIGGER_CUSTOM_STATE_ID {
	return *(*SERVICE_TRIGGER_CUSTOM_STATE_ID)(unsafe.Pointer(this))
}

func (this *SERVICE_CUSTOM_SYSTEM_STATE_CHANGE_DATA_ITEM_U) S() *SERVICE_CUSTOM_SYSTEM_STATE_CHANGE_DATA_ITEM_U_S {
	return (*SERVICE_CUSTOM_SYSTEM_STATE_CHANGE_DATA_ITEM_U_S)(unsafe.Pointer(this))
}

func (this *SERVICE_CUSTOM_SYSTEM_STATE_CHANGE_DATA_ITEM_U) SVal() SERVICE_CUSTOM_SYSTEM_STATE_CHANGE_DATA_ITEM_U_S {
	return *(*SERVICE_CUSTOM_SYSTEM_STATE_CHANGE_DATA_ITEM_U_S)(unsafe.Pointer(this))
}

type SERVICE_CUSTOM_SYSTEM_STATE_CHANGE_DATA_ITEM struct {
	U SERVICE_CUSTOM_SYSTEM_STATE_CHANGE_DATA_ITEM_U
}

type SERVICE_DESCRIPTIONA struct {
	LpDescription PSTR
}

type SERVICE_DESCRIPTION = SERVICE_DESCRIPTIONW
type SERVICE_DESCRIPTIONW struct {
	LpDescription PWSTR
}

type SC_ACTION struct {
	Type  SC_ACTION_TYPE
	Delay uint32
}

type SERVICE_FAILURE_ACTIONSA struct {
	DwResetPeriod uint32
	LpRebootMsg   PSTR
	LpCommand     PSTR
	CActions      uint32
	LpsaActions   *SC_ACTION
}

type SERVICE_FAILURE_ACTIONS = SERVICE_FAILURE_ACTIONSW
type SERVICE_FAILURE_ACTIONSW struct {
	DwResetPeriod uint32
	LpRebootMsg   PWSTR
	LpCommand     PWSTR
	CActions      uint32
	LpsaActions   *SC_ACTION
}

type SERVICE_DELAYED_AUTO_START_INFO struct {
	FDelayedAutostart BOOL
}

type SERVICE_FAILURE_ACTIONS_FLAG struct {
	FFailureActionsOnNonCrashFailures BOOL
}

type SERVICE_SID_INFO struct {
	DwServiceSidType uint32
}

type SERVICE_REQUIRED_PRIVILEGES_INFOA struct {
	PmszRequiredPrivileges PSTR
}

type SERVICE_REQUIRED_PRIVILEGES_INFO = SERVICE_REQUIRED_PRIVILEGES_INFOW
type SERVICE_REQUIRED_PRIVILEGES_INFOW struct {
	PmszRequiredPrivileges PWSTR
}

type SERVICE_PRESHUTDOWN_INFO struct {
	DwPreshutdownTimeout uint32
}

type SERVICE_TRIGGER_SPECIFIC_DATA_ITEM struct {
	DwDataType SERVICE_TRIGGER_SPECIFIC_DATA_ITEM_DATA_TYPE
	CbData     uint32
	PData      *byte
}

type SERVICE_TRIGGER struct {
	DwTriggerType   SERVICE_TRIGGER_TYPE
	DwAction        SERVICE_TRIGGER_ACTION
	PTriggerSubtype *syscall.GUID
	CDataItems      uint32
	PDataItems      *SERVICE_TRIGGER_SPECIFIC_DATA_ITEM
}

type SERVICE_TRIGGER_INFO struct {
	CTriggers uint32
	PTriggers *SERVICE_TRIGGER
	PReserved *byte
}

type SERVICE_PREFERRED_NODE_INFO struct {
	UsPreferredNode uint16
	FDelete         BOOLEAN
}

type SERVICE_TIMECHANGE_INFO struct {
	LiNewTime int64
	LiOldTime int64
}

type SERVICE_LAUNCH_PROTECTED_INFO struct {
	DwLaunchProtected uint32
}

type SERVICE_STATUS struct {
	DwServiceType             ENUM_SERVICE_TYPE
	DwCurrentState            SERVICE_STATUS_CURRENT_STATE
	DwControlsAccepted        uint32
	DwWin32ExitCode           uint32
	DwServiceSpecificExitCode uint32
	DwCheckPoint              uint32
	DwWaitHint                uint32
}

type SERVICE_STATUS_PROCESS struct {
	DwServiceType             ENUM_SERVICE_TYPE
	DwCurrentState            SERVICE_STATUS_CURRENT_STATE
	DwControlsAccepted        uint32
	DwWin32ExitCode           uint32
	DwServiceSpecificExitCode uint32
	DwCheckPoint              uint32
	DwWaitHint                uint32
	DwProcessId               uint32
	DwServiceFlags            SERVICE_RUNS_IN_PROCESS
}

type ENUM_SERVICE_STATUSA struct {
	LpServiceName PSTR
	LpDisplayName PSTR
	ServiceStatus SERVICE_STATUS
}

type ENUM_SERVICE_STATUS = ENUM_SERVICE_STATUSW
type ENUM_SERVICE_STATUSW struct {
	LpServiceName PWSTR
	LpDisplayName PWSTR
	ServiceStatus SERVICE_STATUS
}

type ENUM_SERVICE_STATUS_PROCESSA struct {
	LpServiceName        PSTR
	LpDisplayName        PSTR
	ServiceStatusProcess SERVICE_STATUS_PROCESS
}

type ENUM_SERVICE_STATUS_PROCESS = ENUM_SERVICE_STATUS_PROCESSW
type ENUM_SERVICE_STATUS_PROCESSW struct {
	LpServiceName        PWSTR
	LpDisplayName        PWSTR
	ServiceStatusProcess SERVICE_STATUS_PROCESS
}

type QUERY_SERVICE_LOCK_STATUSA struct {
	FIsLocked      uint32
	LpLockOwner    PSTR
	DwLockDuration uint32
}

type QUERY_SERVICE_LOCK_STATUS = QUERY_SERVICE_LOCK_STATUSW
type QUERY_SERVICE_LOCK_STATUSW struct {
	FIsLocked      uint32
	LpLockOwner    PWSTR
	DwLockDuration uint32
}

type QUERY_SERVICE_CONFIGA struct {
	DwServiceType      ENUM_SERVICE_TYPE
	DwStartType        SERVICE_START_TYPE
	DwErrorControl     SERVICE_ERROR
	LpBinaryPathName   PSTR
	LpLoadOrderGroup   PSTR
	DwTagId            uint32
	LpDependencies     PSTR
	LpServiceStartName PSTR
	LpDisplayName      PSTR
}

type QUERY_SERVICE_CONFIG = QUERY_SERVICE_CONFIGW
type QUERY_SERVICE_CONFIGW struct {
	DwServiceType      ENUM_SERVICE_TYPE
	DwStartType        SERVICE_START_TYPE
	DwErrorControl     SERVICE_ERROR
	LpBinaryPathName   PWSTR
	LpLoadOrderGroup   PWSTR
	DwTagId            uint32
	LpDependencies     PWSTR
	LpServiceStartName PWSTR
	LpDisplayName      PWSTR
}

type SERVICE_TABLE_ENTRYA struct {
	LpServiceName PSTR
	LpServiceProc LPSERVICE_MAIN_FUNCTIONA
}

type SERVICE_TABLE_ENTRY = SERVICE_TABLE_ENTRYW
type SERVICE_TABLE_ENTRYW struct {
	LpServiceName PWSTR
	LpServiceProc LPSERVICE_MAIN_FUNCTIONW
}

type SERVICE_NOTIFY_1 struct {
	DwVersion            uint32
	PfnNotifyCallback    PFN_SC_NOTIFY_CALLBACK
	PContext             unsafe.Pointer
	DwNotificationStatus uint32
	ServiceStatus        SERVICE_STATUS_PROCESS
}

type SERVICE_NOTIFY_2A struct {
	DwVersion               uint32
	PfnNotifyCallback       PFN_SC_NOTIFY_CALLBACK
	PContext                unsafe.Pointer
	DwNotificationStatus    uint32
	ServiceStatus           SERVICE_STATUS_PROCESS
	DwNotificationTriggered uint32
	PszServiceNames         PSTR
}

type SERVICE_NOTIFY_2 = SERVICE_NOTIFY_2W
type SERVICE_NOTIFY_2W struct {
	DwVersion               uint32
	PfnNotifyCallback       PFN_SC_NOTIFY_CALLBACK
	PContext                unsafe.Pointer
	DwNotificationStatus    uint32
	ServiceStatus           SERVICE_STATUS_PROCESS
	DwNotificationTriggered uint32
	PszServiceNames         PWSTR
}

type SERVICE_CONTROL_STATUS_REASON_PARAMSA struct {
	DwReason      uint32
	PszComment    PSTR
	ServiceStatus SERVICE_STATUS_PROCESS
}

type SERVICE_CONTROL_STATUS_REASON_PARAMS = SERVICE_CONTROL_STATUS_REASON_PARAMSW
type SERVICE_CONTROL_STATUS_REASON_PARAMSW struct {
	DwReason      uint32
	PszComment    PWSTR
	ServiceStatus SERVICE_STATUS_PROCESS
}

type SERVICE_START_REASON struct {
	DwReason uint32
}

type SC_NOTIFICATION_REGISTRATION_ struct {
}

// func types

type SERVICE_MAIN_FUNCTIONW = uintptr
type SERVICE_MAIN_FUNCTIONW_func = func(dwNumServicesArgs uint32, lpServiceArgVectors *PWSTR)

type SERVICE_MAIN_FUNCTIONA = uintptr
type SERVICE_MAIN_FUNCTIONA_func = func(dwNumServicesArgs uint32, lpServiceArgVectors **int8)

type LPSERVICE_MAIN_FUNCTIONW = uintptr
type LPSERVICE_MAIN_FUNCTIONW_func = func(dwNumServicesArgs uint32, lpServiceArgVectors *PWSTR)

type LPSERVICE_MAIN_FUNCTIONA = uintptr
type LPSERVICE_MAIN_FUNCTIONA_func = func(dwNumServicesArgs uint32, lpServiceArgVectors *PSTR)

type HANDLER_FUNCTION = uintptr
type HANDLER_FUNCTION_func = func(dwControl uint32)

type HANDLER_FUNCTION_EX = uintptr
type HANDLER_FUNCTION_EX_func = func(dwControl uint32, dwEventType uint32, lpEventData unsafe.Pointer, lpContext unsafe.Pointer) uint32

type LPHANDLER_FUNCTION = uintptr
type LPHANDLER_FUNCTION_func = func(dwControl uint32)

type LPHANDLER_FUNCTION_EX = uintptr
type LPHANDLER_FUNCTION_EX_func = func(dwControl uint32, dwEventType uint32, lpEventData unsafe.Pointer, lpContext unsafe.Pointer) uint32

type PFN_SC_NOTIFY_CALLBACK = uintptr
type PFN_SC_NOTIFY_CALLBACK_func = func(pParameter unsafe.Pointer)

type PSC_NOTIFICATION_CALLBACK = uintptr
type PSC_NOTIFICATION_CALLBACK_func = func(dwNotify uint32, pCallbackContext unsafe.Pointer)

var (
	pSetServiceBits                 uintptr
	pChangeServiceConfigA           uintptr
	pChangeServiceConfigW           uintptr
	pChangeServiceConfig2A          uintptr
	pChangeServiceConfig2W          uintptr
	pCloseServiceHandle             uintptr
	pControlService                 uintptr
	pCreateServiceA                 uintptr
	pCreateServiceW                 uintptr
	pDeleteService                  uintptr
	pEnumDependentServicesA         uintptr
	pEnumDependentServicesW         uintptr
	pEnumServicesStatusA            uintptr
	pEnumServicesStatusW            uintptr
	pEnumServicesStatusExA          uintptr
	pEnumServicesStatusExW          uintptr
	pGetServiceKeyNameA             uintptr
	pGetServiceKeyNameW             uintptr
	pGetServiceDisplayNameA         uintptr
	pGetServiceDisplayNameW         uintptr
	pLockServiceDatabase            uintptr
	pNotifyBootConfigStatus         uintptr
	pOpenSCManagerA                 uintptr
	pOpenSCManagerW                 uintptr
	pOpenServiceA                   uintptr
	pOpenServiceW                   uintptr
	pQueryServiceConfigA            uintptr
	pQueryServiceConfigW            uintptr
	pQueryServiceConfig2A           uintptr
	pQueryServiceConfig2W           uintptr
	pQueryServiceLockStatusA        uintptr
	pQueryServiceLockStatusW        uintptr
	pQueryServiceObjectSecurity     uintptr
	pQueryServiceStatus             uintptr
	pQueryServiceStatusEx           uintptr
	pRegisterServiceCtrlHandlerA    uintptr
	pRegisterServiceCtrlHandlerW    uintptr
	pRegisterServiceCtrlHandlerExA  uintptr
	pRegisterServiceCtrlHandlerExW  uintptr
	pSetServiceObjectSecurity       uintptr
	pSetServiceStatus               uintptr
	pStartServiceCtrlDispatcherA    uintptr
	pStartServiceCtrlDispatcherW    uintptr
	pStartServiceA                  uintptr
	pStartServiceW                  uintptr
	pUnlockServiceDatabase          uintptr
	pNotifyServiceStatusChangeA     uintptr
	pNotifyServiceStatusChangeW     uintptr
	pControlServiceExA              uintptr
	pControlServiceExW              uintptr
	pQueryServiceDynamicInformation uintptr
	pWaitServiceState               uintptr
)

func SetServiceBits(hServiceStatus SERVICE_STATUS_HANDLE, dwServiceBits uint32, bSetBitsOn BOOL, bUpdateImmediately BOOL) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pSetServiceBits, libAdvapi32, "SetServiceBits")
	ret, _, err := syscall.SyscallN(addr, hServiceStatus, uintptr(dwServiceBits), uintptr(bSetBitsOn), uintptr(bUpdateImmediately))
	return BOOL(ret), WIN32_ERROR(err)
}

func ChangeServiceConfigA(hService SC_HANDLE, dwServiceType uint32, dwStartType SERVICE_START_TYPE, dwErrorControl SERVICE_ERROR, lpBinaryPathName PSTR, lpLoadOrderGroup PSTR, lpdwTagId *uint32, lpDependencies PSTR, lpServiceStartName PSTR, lpPassword PSTR, lpDisplayName PSTR) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pChangeServiceConfigA, libAdvapi32, "ChangeServiceConfigA")
	ret, _, err := syscall.SyscallN(addr, hService, uintptr(dwServiceType), uintptr(dwStartType), uintptr(dwErrorControl), uintptr(unsafe.Pointer(lpBinaryPathName)), uintptr(unsafe.Pointer(lpLoadOrderGroup)), uintptr(unsafe.Pointer(lpdwTagId)), uintptr(unsafe.Pointer(lpDependencies)), uintptr(unsafe.Pointer(lpServiceStartName)), uintptr(unsafe.Pointer(lpPassword)), uintptr(unsafe.Pointer(lpDisplayName)))
	return BOOL(ret), WIN32_ERROR(err)
}

var ChangeServiceConfig = ChangeServiceConfigW

func ChangeServiceConfigW(hService SC_HANDLE, dwServiceType uint32, dwStartType SERVICE_START_TYPE, dwErrorControl SERVICE_ERROR, lpBinaryPathName PWSTR, lpLoadOrderGroup PWSTR, lpdwTagId *uint32, lpDependencies PWSTR, lpServiceStartName PWSTR, lpPassword PWSTR, lpDisplayName PWSTR) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pChangeServiceConfigW, libAdvapi32, "ChangeServiceConfigW")
	ret, _, err := syscall.SyscallN(addr, hService, uintptr(dwServiceType), uintptr(dwStartType), uintptr(dwErrorControl), uintptr(unsafe.Pointer(lpBinaryPathName)), uintptr(unsafe.Pointer(lpLoadOrderGroup)), uintptr(unsafe.Pointer(lpdwTagId)), uintptr(unsafe.Pointer(lpDependencies)), uintptr(unsafe.Pointer(lpServiceStartName)), uintptr(unsafe.Pointer(lpPassword)), uintptr(unsafe.Pointer(lpDisplayName)))
	return BOOL(ret), WIN32_ERROR(err)
}

func ChangeServiceConfig2A(hService SC_HANDLE, dwInfoLevel SERVICE_CONFIG, lpInfo unsafe.Pointer) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pChangeServiceConfig2A, libAdvapi32, "ChangeServiceConfig2A")
	ret, _, err := syscall.SyscallN(addr, hService, uintptr(dwInfoLevel), uintptr(lpInfo))
	return BOOL(ret), WIN32_ERROR(err)
}

var ChangeServiceConfig2 = ChangeServiceConfig2W

func ChangeServiceConfig2W(hService SC_HANDLE, dwInfoLevel SERVICE_CONFIG, lpInfo unsafe.Pointer) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pChangeServiceConfig2W, libAdvapi32, "ChangeServiceConfig2W")
	ret, _, err := syscall.SyscallN(addr, hService, uintptr(dwInfoLevel), uintptr(lpInfo))
	return BOOL(ret), WIN32_ERROR(err)
}

func CloseServiceHandle(hSCObject SC_HANDLE) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pCloseServiceHandle, libAdvapi32, "CloseServiceHandle")
	ret, _, err := syscall.SyscallN(addr, hSCObject)
	return BOOL(ret), WIN32_ERROR(err)
}

func ControlService(hService SC_HANDLE, dwControl uint32, lpServiceStatus *SERVICE_STATUS) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pControlService, libAdvapi32, "ControlService")
	ret, _, err := syscall.SyscallN(addr, hService, uintptr(dwControl), uintptr(unsafe.Pointer(lpServiceStatus)))
	return BOOL(ret), WIN32_ERROR(err)
}

func CreateServiceA(hSCManager SC_HANDLE, lpServiceName PSTR, lpDisplayName PSTR, dwDesiredAccess uint32, dwServiceType ENUM_SERVICE_TYPE, dwStartType SERVICE_START_TYPE, dwErrorControl SERVICE_ERROR, lpBinaryPathName PSTR, lpLoadOrderGroup PSTR, lpdwTagId *uint32, lpDependencies PSTR, lpServiceStartName PSTR, lpPassword PSTR) (SC_HANDLE, WIN32_ERROR) {
	addr := lazyAddr(&pCreateServiceA, libAdvapi32, "CreateServiceA")
	ret, _, err := syscall.SyscallN(addr, hSCManager, uintptr(unsafe.Pointer(lpServiceName)), uintptr(unsafe.Pointer(lpDisplayName)), uintptr(dwDesiredAccess), uintptr(dwServiceType), uintptr(dwStartType), uintptr(dwErrorControl), uintptr(unsafe.Pointer(lpBinaryPathName)), uintptr(unsafe.Pointer(lpLoadOrderGroup)), uintptr(unsafe.Pointer(lpdwTagId)), uintptr(unsafe.Pointer(lpDependencies)), uintptr(unsafe.Pointer(lpServiceStartName)), uintptr(unsafe.Pointer(lpPassword)))
	return ret, WIN32_ERROR(err)
}

var CreateService = CreateServiceW

func CreateServiceW(hSCManager SC_HANDLE, lpServiceName PWSTR, lpDisplayName PWSTR, dwDesiredAccess uint32, dwServiceType ENUM_SERVICE_TYPE, dwStartType SERVICE_START_TYPE, dwErrorControl SERVICE_ERROR, lpBinaryPathName PWSTR, lpLoadOrderGroup PWSTR, lpdwTagId *uint32, lpDependencies PWSTR, lpServiceStartName PWSTR, lpPassword PWSTR) (SC_HANDLE, WIN32_ERROR) {
	addr := lazyAddr(&pCreateServiceW, libAdvapi32, "CreateServiceW")
	ret, _, err := syscall.SyscallN(addr, hSCManager, uintptr(unsafe.Pointer(lpServiceName)), uintptr(unsafe.Pointer(lpDisplayName)), uintptr(dwDesiredAccess), uintptr(dwServiceType), uintptr(dwStartType), uintptr(dwErrorControl), uintptr(unsafe.Pointer(lpBinaryPathName)), uintptr(unsafe.Pointer(lpLoadOrderGroup)), uintptr(unsafe.Pointer(lpdwTagId)), uintptr(unsafe.Pointer(lpDependencies)), uintptr(unsafe.Pointer(lpServiceStartName)), uintptr(unsafe.Pointer(lpPassword)))
	return ret, WIN32_ERROR(err)
}

func DeleteService(hService SC_HANDLE) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pDeleteService, libAdvapi32, "DeleteService")
	ret, _, err := syscall.SyscallN(addr, hService)
	return BOOL(ret), WIN32_ERROR(err)
}

func EnumDependentServicesA(hService SC_HANDLE, dwServiceState ENUM_SERVICE_STATE, lpServices *ENUM_SERVICE_STATUSA, cbBufSize uint32, pcbBytesNeeded *uint32, lpServicesReturned *uint32) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pEnumDependentServicesA, libAdvapi32, "EnumDependentServicesA")
	ret, _, err := syscall.SyscallN(addr, hService, uintptr(dwServiceState), uintptr(unsafe.Pointer(lpServices)), uintptr(cbBufSize), uintptr(unsafe.Pointer(pcbBytesNeeded)), uintptr(unsafe.Pointer(lpServicesReturned)))
	return BOOL(ret), WIN32_ERROR(err)
}

var EnumDependentServices = EnumDependentServicesW

func EnumDependentServicesW(hService SC_HANDLE, dwServiceState ENUM_SERVICE_STATE, lpServices *ENUM_SERVICE_STATUSW, cbBufSize uint32, pcbBytesNeeded *uint32, lpServicesReturned *uint32) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pEnumDependentServicesW, libAdvapi32, "EnumDependentServicesW")
	ret, _, err := syscall.SyscallN(addr, hService, uintptr(dwServiceState), uintptr(unsafe.Pointer(lpServices)), uintptr(cbBufSize), uintptr(unsafe.Pointer(pcbBytesNeeded)), uintptr(unsafe.Pointer(lpServicesReturned)))
	return BOOL(ret), WIN32_ERROR(err)
}

func EnumServicesStatusA(hSCManager SC_HANDLE, dwServiceType ENUM_SERVICE_TYPE, dwServiceState ENUM_SERVICE_STATE, lpServices *ENUM_SERVICE_STATUSA, cbBufSize uint32, pcbBytesNeeded *uint32, lpServicesReturned *uint32, lpResumeHandle *uint32) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pEnumServicesStatusA, libAdvapi32, "EnumServicesStatusA")
	ret, _, err := syscall.SyscallN(addr, hSCManager, uintptr(dwServiceType), uintptr(dwServiceState), uintptr(unsafe.Pointer(lpServices)), uintptr(cbBufSize), uintptr(unsafe.Pointer(pcbBytesNeeded)), uintptr(unsafe.Pointer(lpServicesReturned)), uintptr(unsafe.Pointer(lpResumeHandle)))
	return BOOL(ret), WIN32_ERROR(err)
}

var EnumServicesStatus = EnumServicesStatusW

func EnumServicesStatusW(hSCManager SC_HANDLE, dwServiceType ENUM_SERVICE_TYPE, dwServiceState ENUM_SERVICE_STATE, lpServices *ENUM_SERVICE_STATUSW, cbBufSize uint32, pcbBytesNeeded *uint32, lpServicesReturned *uint32, lpResumeHandle *uint32) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pEnumServicesStatusW, libAdvapi32, "EnumServicesStatusW")
	ret, _, err := syscall.SyscallN(addr, hSCManager, uintptr(dwServiceType), uintptr(dwServiceState), uintptr(unsafe.Pointer(lpServices)), uintptr(cbBufSize), uintptr(unsafe.Pointer(pcbBytesNeeded)), uintptr(unsafe.Pointer(lpServicesReturned)), uintptr(unsafe.Pointer(lpResumeHandle)))
	return BOOL(ret), WIN32_ERROR(err)
}

func EnumServicesStatusExA(hSCManager SC_HANDLE, InfoLevel SC_ENUM_TYPE, dwServiceType ENUM_SERVICE_TYPE, dwServiceState ENUM_SERVICE_STATE, lpServices *byte, cbBufSize uint32, pcbBytesNeeded *uint32, lpServicesReturned *uint32, lpResumeHandle *uint32, pszGroupName PSTR) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pEnumServicesStatusExA, libAdvapi32, "EnumServicesStatusExA")
	ret, _, err := syscall.SyscallN(addr, hSCManager, uintptr(InfoLevel), uintptr(dwServiceType), uintptr(dwServiceState), uintptr(unsafe.Pointer(lpServices)), uintptr(cbBufSize), uintptr(unsafe.Pointer(pcbBytesNeeded)), uintptr(unsafe.Pointer(lpServicesReturned)), uintptr(unsafe.Pointer(lpResumeHandle)), uintptr(unsafe.Pointer(pszGroupName)))
	return BOOL(ret), WIN32_ERROR(err)
}

var EnumServicesStatusEx = EnumServicesStatusExW

func EnumServicesStatusExW(hSCManager SC_HANDLE, InfoLevel SC_ENUM_TYPE, dwServiceType ENUM_SERVICE_TYPE, dwServiceState ENUM_SERVICE_STATE, lpServices *byte, cbBufSize uint32, pcbBytesNeeded *uint32, lpServicesReturned *uint32, lpResumeHandle *uint32, pszGroupName PWSTR) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pEnumServicesStatusExW, libAdvapi32, "EnumServicesStatusExW")
	ret, _, err := syscall.SyscallN(addr, hSCManager, uintptr(InfoLevel), uintptr(dwServiceType), uintptr(dwServiceState), uintptr(unsafe.Pointer(lpServices)), uintptr(cbBufSize), uintptr(unsafe.Pointer(pcbBytesNeeded)), uintptr(unsafe.Pointer(lpServicesReturned)), uintptr(unsafe.Pointer(lpResumeHandle)), uintptr(unsafe.Pointer(pszGroupName)))
	return BOOL(ret), WIN32_ERROR(err)
}

func GetServiceKeyNameA(hSCManager SC_HANDLE, lpDisplayName PSTR, lpServiceName PSTR, lpcchBuffer *uint32) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pGetServiceKeyNameA, libAdvapi32, "GetServiceKeyNameA")
	ret, _, err := syscall.SyscallN(addr, hSCManager, uintptr(unsafe.Pointer(lpDisplayName)), uintptr(unsafe.Pointer(lpServiceName)), uintptr(unsafe.Pointer(lpcchBuffer)))
	return BOOL(ret), WIN32_ERROR(err)
}

var GetServiceKeyName = GetServiceKeyNameW

func GetServiceKeyNameW(hSCManager SC_HANDLE, lpDisplayName PWSTR, lpServiceName PWSTR, lpcchBuffer *uint32) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pGetServiceKeyNameW, libAdvapi32, "GetServiceKeyNameW")
	ret, _, err := syscall.SyscallN(addr, hSCManager, uintptr(unsafe.Pointer(lpDisplayName)), uintptr(unsafe.Pointer(lpServiceName)), uintptr(unsafe.Pointer(lpcchBuffer)))
	return BOOL(ret), WIN32_ERROR(err)
}

func GetServiceDisplayNameA(hSCManager SC_HANDLE, lpServiceName PSTR, lpDisplayName PSTR, lpcchBuffer *uint32) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pGetServiceDisplayNameA, libAdvapi32, "GetServiceDisplayNameA")
	ret, _, err := syscall.SyscallN(addr, hSCManager, uintptr(unsafe.Pointer(lpServiceName)), uintptr(unsafe.Pointer(lpDisplayName)), uintptr(unsafe.Pointer(lpcchBuffer)))
	return BOOL(ret), WIN32_ERROR(err)
}

var GetServiceDisplayName = GetServiceDisplayNameW

func GetServiceDisplayNameW(hSCManager SC_HANDLE, lpServiceName PWSTR, lpDisplayName PWSTR, lpcchBuffer *uint32) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pGetServiceDisplayNameW, libAdvapi32, "GetServiceDisplayNameW")
	ret, _, err := syscall.SyscallN(addr, hSCManager, uintptr(unsafe.Pointer(lpServiceName)), uintptr(unsafe.Pointer(lpDisplayName)), uintptr(unsafe.Pointer(lpcchBuffer)))
	return BOOL(ret), WIN32_ERROR(err)
}

func LockServiceDatabase(hSCManager SC_HANDLE) (unsafe.Pointer, WIN32_ERROR) {
	addr := lazyAddr(&pLockServiceDatabase, libAdvapi32, "LockServiceDatabase")
	ret, _, err := syscall.SyscallN(addr, hSCManager)
	return (unsafe.Pointer)(ret), WIN32_ERROR(err)
}

func NotifyBootConfigStatus(BootAcceptable BOOL) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pNotifyBootConfigStatus, libAdvapi32, "NotifyBootConfigStatus")
	ret, _, err := syscall.SyscallN(addr, uintptr(BootAcceptable))
	return BOOL(ret), WIN32_ERROR(err)
}

func OpenSCManagerA(lpMachineName PSTR, lpDatabaseName PSTR, dwDesiredAccess uint32) (SC_HANDLE, WIN32_ERROR) {
	addr := lazyAddr(&pOpenSCManagerA, libAdvapi32, "OpenSCManagerA")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpMachineName)), uintptr(unsafe.Pointer(lpDatabaseName)), uintptr(dwDesiredAccess))
	return ret, WIN32_ERROR(err)
}

var OpenSCManager = OpenSCManagerW

func OpenSCManagerW(lpMachineName PWSTR, lpDatabaseName PWSTR, dwDesiredAccess uint32) (SC_HANDLE, WIN32_ERROR) {
	addr := lazyAddr(&pOpenSCManagerW, libAdvapi32, "OpenSCManagerW")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpMachineName)), uintptr(unsafe.Pointer(lpDatabaseName)), uintptr(dwDesiredAccess))
	return ret, WIN32_ERROR(err)
}

func OpenServiceA(hSCManager SC_HANDLE, lpServiceName PSTR, dwDesiredAccess uint32) (SC_HANDLE, WIN32_ERROR) {
	addr := lazyAddr(&pOpenServiceA, libAdvapi32, "OpenServiceA")
	ret, _, err := syscall.SyscallN(addr, hSCManager, uintptr(unsafe.Pointer(lpServiceName)), uintptr(dwDesiredAccess))
	return ret, WIN32_ERROR(err)
}

var OpenService = OpenServiceW

func OpenServiceW(hSCManager SC_HANDLE, lpServiceName PWSTR, dwDesiredAccess uint32) (SC_HANDLE, WIN32_ERROR) {
	addr := lazyAddr(&pOpenServiceW, libAdvapi32, "OpenServiceW")
	ret, _, err := syscall.SyscallN(addr, hSCManager, uintptr(unsafe.Pointer(lpServiceName)), uintptr(dwDesiredAccess))
	return ret, WIN32_ERROR(err)
}

func QueryServiceConfigA(hService SC_HANDLE, lpServiceConfig *QUERY_SERVICE_CONFIGA, cbBufSize uint32, pcbBytesNeeded *uint32) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pQueryServiceConfigA, libAdvapi32, "QueryServiceConfigA")
	ret, _, err := syscall.SyscallN(addr, hService, uintptr(unsafe.Pointer(lpServiceConfig)), uintptr(cbBufSize), uintptr(unsafe.Pointer(pcbBytesNeeded)))
	return BOOL(ret), WIN32_ERROR(err)
}

var QueryServiceConfig = QueryServiceConfigW

func QueryServiceConfigW(hService SC_HANDLE, lpServiceConfig *QUERY_SERVICE_CONFIGW, cbBufSize uint32, pcbBytesNeeded *uint32) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pQueryServiceConfigW, libAdvapi32, "QueryServiceConfigW")
	ret, _, err := syscall.SyscallN(addr, hService, uintptr(unsafe.Pointer(lpServiceConfig)), uintptr(cbBufSize), uintptr(unsafe.Pointer(pcbBytesNeeded)))
	return BOOL(ret), WIN32_ERROR(err)
}

func QueryServiceConfig2A(hService SC_HANDLE, dwInfoLevel SERVICE_CONFIG, lpBuffer *byte, cbBufSize uint32, pcbBytesNeeded *uint32) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pQueryServiceConfig2A, libAdvapi32, "QueryServiceConfig2A")
	ret, _, err := syscall.SyscallN(addr, hService, uintptr(dwInfoLevel), uintptr(unsafe.Pointer(lpBuffer)), uintptr(cbBufSize), uintptr(unsafe.Pointer(pcbBytesNeeded)))
	return BOOL(ret), WIN32_ERROR(err)
}

var QueryServiceConfig2 = QueryServiceConfig2W

func QueryServiceConfig2W(hService SC_HANDLE, dwInfoLevel SERVICE_CONFIG, lpBuffer *byte, cbBufSize uint32, pcbBytesNeeded *uint32) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pQueryServiceConfig2W, libAdvapi32, "QueryServiceConfig2W")
	ret, _, err := syscall.SyscallN(addr, hService, uintptr(dwInfoLevel), uintptr(unsafe.Pointer(lpBuffer)), uintptr(cbBufSize), uintptr(unsafe.Pointer(pcbBytesNeeded)))
	return BOOL(ret), WIN32_ERROR(err)
}

func QueryServiceLockStatusA(hSCManager SC_HANDLE, lpLockStatus *QUERY_SERVICE_LOCK_STATUSA, cbBufSize uint32, pcbBytesNeeded *uint32) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pQueryServiceLockStatusA, libAdvapi32, "QueryServiceLockStatusA")
	ret, _, err := syscall.SyscallN(addr, hSCManager, uintptr(unsafe.Pointer(lpLockStatus)), uintptr(cbBufSize), uintptr(unsafe.Pointer(pcbBytesNeeded)))
	return BOOL(ret), WIN32_ERROR(err)
}

var QueryServiceLockStatus = QueryServiceLockStatusW

func QueryServiceLockStatusW(hSCManager SC_HANDLE, lpLockStatus *QUERY_SERVICE_LOCK_STATUSW, cbBufSize uint32, pcbBytesNeeded *uint32) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pQueryServiceLockStatusW, libAdvapi32, "QueryServiceLockStatusW")
	ret, _, err := syscall.SyscallN(addr, hSCManager, uintptr(unsafe.Pointer(lpLockStatus)), uintptr(cbBufSize), uintptr(unsafe.Pointer(pcbBytesNeeded)))
	return BOOL(ret), WIN32_ERROR(err)
}

func QueryServiceObjectSecurity(hService SC_HANDLE, dwSecurityInformation uint32, lpSecurityDescriptor PSECURITY_DESCRIPTOR, cbBufSize uint32, pcbBytesNeeded *uint32) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pQueryServiceObjectSecurity, libAdvapi32, "QueryServiceObjectSecurity")
	ret, _, err := syscall.SyscallN(addr, hService, uintptr(dwSecurityInformation), uintptr(unsafe.Pointer(lpSecurityDescriptor)), uintptr(cbBufSize), uintptr(unsafe.Pointer(pcbBytesNeeded)))
	return BOOL(ret), WIN32_ERROR(err)
}

func QueryServiceStatus(hService SC_HANDLE, lpServiceStatus *SERVICE_STATUS) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pQueryServiceStatus, libAdvapi32, "QueryServiceStatus")
	ret, _, err := syscall.SyscallN(addr, hService, uintptr(unsafe.Pointer(lpServiceStatus)))
	return BOOL(ret), WIN32_ERROR(err)
}

func QueryServiceStatusEx(hService SC_HANDLE, InfoLevel SC_STATUS_TYPE, lpBuffer *byte, cbBufSize uint32, pcbBytesNeeded *uint32) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pQueryServiceStatusEx, libAdvapi32, "QueryServiceStatusEx")
	ret, _, err := syscall.SyscallN(addr, hService, uintptr(InfoLevel), uintptr(unsafe.Pointer(lpBuffer)), uintptr(cbBufSize), uintptr(unsafe.Pointer(pcbBytesNeeded)))
	return BOOL(ret), WIN32_ERROR(err)
}

func RegisterServiceCtrlHandlerA(lpServiceName PSTR, lpHandlerProc LPHANDLER_FUNCTION) (SERVICE_STATUS_HANDLE, WIN32_ERROR) {
	addr := lazyAddr(&pRegisterServiceCtrlHandlerA, libAdvapi32, "RegisterServiceCtrlHandlerA")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpServiceName)), lpHandlerProc)
	return ret, WIN32_ERROR(err)
}

var RegisterServiceCtrlHandler = RegisterServiceCtrlHandlerW

func RegisterServiceCtrlHandlerW(lpServiceName PWSTR, lpHandlerProc LPHANDLER_FUNCTION) (SERVICE_STATUS_HANDLE, WIN32_ERROR) {
	addr := lazyAddr(&pRegisterServiceCtrlHandlerW, libAdvapi32, "RegisterServiceCtrlHandlerW")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpServiceName)), lpHandlerProc)
	return ret, WIN32_ERROR(err)
}

func RegisterServiceCtrlHandlerExA(lpServiceName PSTR, lpHandlerProc LPHANDLER_FUNCTION_EX, lpContext unsafe.Pointer) (SERVICE_STATUS_HANDLE, WIN32_ERROR) {
	addr := lazyAddr(&pRegisterServiceCtrlHandlerExA, libAdvapi32, "RegisterServiceCtrlHandlerExA")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpServiceName)), lpHandlerProc, uintptr(lpContext))
	return ret, WIN32_ERROR(err)
}

var RegisterServiceCtrlHandlerEx = RegisterServiceCtrlHandlerExW

func RegisterServiceCtrlHandlerExW(lpServiceName PWSTR, lpHandlerProc LPHANDLER_FUNCTION_EX, lpContext unsafe.Pointer) (SERVICE_STATUS_HANDLE, WIN32_ERROR) {
	addr := lazyAddr(&pRegisterServiceCtrlHandlerExW, libAdvapi32, "RegisterServiceCtrlHandlerExW")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpServiceName)), lpHandlerProc, uintptr(lpContext))
	return ret, WIN32_ERROR(err)
}

func SetServiceObjectSecurity(hService SC_HANDLE, dwSecurityInformation OBJECT_SECURITY_INFORMATION, lpSecurityDescriptor PSECURITY_DESCRIPTOR) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pSetServiceObjectSecurity, libAdvapi32, "SetServiceObjectSecurity")
	ret, _, err := syscall.SyscallN(addr, hService, uintptr(dwSecurityInformation), uintptr(unsafe.Pointer(lpSecurityDescriptor)))
	return BOOL(ret), WIN32_ERROR(err)
}

func SetServiceStatus(hServiceStatus SERVICE_STATUS_HANDLE, lpServiceStatus *SERVICE_STATUS) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pSetServiceStatus, libAdvapi32, "SetServiceStatus")
	ret, _, err := syscall.SyscallN(addr, hServiceStatus, uintptr(unsafe.Pointer(lpServiceStatus)))
	return BOOL(ret), WIN32_ERROR(err)
}

func StartServiceCtrlDispatcherA(lpServiceStartTable *SERVICE_TABLE_ENTRYA) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pStartServiceCtrlDispatcherA, libAdvapi32, "StartServiceCtrlDispatcherA")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpServiceStartTable)))
	return BOOL(ret), WIN32_ERROR(err)
}

var StartServiceCtrlDispatcher = StartServiceCtrlDispatcherW

func StartServiceCtrlDispatcherW(lpServiceStartTable *SERVICE_TABLE_ENTRYW) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pStartServiceCtrlDispatcherW, libAdvapi32, "StartServiceCtrlDispatcherW")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpServiceStartTable)))
	return BOOL(ret), WIN32_ERROR(err)
}

func StartServiceA(hService SC_HANDLE, dwNumServiceArgs uint32, lpServiceArgVectors *PSTR) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pStartServiceA, libAdvapi32, "StartServiceA")
	ret, _, err := syscall.SyscallN(addr, hService, uintptr(dwNumServiceArgs), uintptr(unsafe.Pointer(lpServiceArgVectors)))
	return BOOL(ret), WIN32_ERROR(err)
}

var StartService = StartServiceW

func StartServiceW(hService SC_HANDLE, dwNumServiceArgs uint32, lpServiceArgVectors *PWSTR) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pStartServiceW, libAdvapi32, "StartServiceW")
	ret, _, err := syscall.SyscallN(addr, hService, uintptr(dwNumServiceArgs), uintptr(unsafe.Pointer(lpServiceArgVectors)))
	return BOOL(ret), WIN32_ERROR(err)
}

func UnlockServiceDatabase(ScLock unsafe.Pointer) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pUnlockServiceDatabase, libAdvapi32, "UnlockServiceDatabase")
	ret, _, err := syscall.SyscallN(addr, uintptr(ScLock))
	return BOOL(ret), WIN32_ERROR(err)
}

func NotifyServiceStatusChangeA(hService SC_HANDLE, dwNotifyMask SERVICE_NOTIFY, pNotifyBuffer *SERVICE_NOTIFY_2A) uint32 {
	addr := lazyAddr(&pNotifyServiceStatusChangeA, libAdvapi32, "NotifyServiceStatusChangeA")
	ret, _, _ := syscall.SyscallN(addr, hService, uintptr(dwNotifyMask), uintptr(unsafe.Pointer(pNotifyBuffer)))
	return uint32(ret)
}

var NotifyServiceStatusChange = NotifyServiceStatusChangeW

func NotifyServiceStatusChangeW(hService SC_HANDLE, dwNotifyMask SERVICE_NOTIFY, pNotifyBuffer *SERVICE_NOTIFY_2W) uint32 {
	addr := lazyAddr(&pNotifyServiceStatusChangeW, libAdvapi32, "NotifyServiceStatusChangeW")
	ret, _, _ := syscall.SyscallN(addr, hService, uintptr(dwNotifyMask), uintptr(unsafe.Pointer(pNotifyBuffer)))
	return uint32(ret)
}

func ControlServiceExA(hService SC_HANDLE, dwControl uint32, dwInfoLevel uint32, pControlParams unsafe.Pointer) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pControlServiceExA, libAdvapi32, "ControlServiceExA")
	ret, _, err := syscall.SyscallN(addr, hService, uintptr(dwControl), uintptr(dwInfoLevel), uintptr(pControlParams))
	return BOOL(ret), WIN32_ERROR(err)
}

var ControlServiceEx = ControlServiceExW

func ControlServiceExW(hService SC_HANDLE, dwControl uint32, dwInfoLevel uint32, pControlParams unsafe.Pointer) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pControlServiceExW, libAdvapi32, "ControlServiceExW")
	ret, _, err := syscall.SyscallN(addr, hService, uintptr(dwControl), uintptr(dwInfoLevel), uintptr(pControlParams))
	return BOOL(ret), WIN32_ERROR(err)
}

func QueryServiceDynamicInformation(hServiceStatus SERVICE_STATUS_HANDLE, dwInfoLevel uint32, ppDynamicInfo unsafe.Pointer) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pQueryServiceDynamicInformation, libAdvapi32, "QueryServiceDynamicInformation")
	ret, _, err := syscall.SyscallN(addr, hServiceStatus, uintptr(dwInfoLevel), uintptr(ppDynamicInfo))
	return BOOL(ret), WIN32_ERROR(err)
}

func WaitServiceState(hService SC_HANDLE, dwNotify uint32, dwTimeout uint32, hCancelEvent HANDLE) uint32 {
	addr := lazyAddr(&pWaitServiceState, libAdvapi32, "WaitServiceState")
	ret, _, _ := syscall.SyscallN(addr, hService, uintptr(dwNotify), uintptr(dwTimeout), hCancelEvent)
	return uint32(ret)
}
