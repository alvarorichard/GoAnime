package win32

import (
	"syscall"
	"unsafe"
)

type (
	HPOWERNOTIFY = uintptr
)

const (
	BATTERY_UNKNOWN_CAPACITY                  uint32 = 0xffffffff
	UNKNOWN_CAPACITY                          uint32 = 0xffffffff
	BATTERY_SYSTEM_BATTERY                    uint32 = 0x80000000
	BATTERY_CAPACITY_RELATIVE                 uint32 = 0x40000000
	BATTERY_IS_SHORT_TERM                     uint32 = 0x20000000
	BATTERY_SEALED                            uint32 = 0x10000000
	BATTERY_SET_CHARGE_SUPPORTED              uint32 = 0x1
	BATTERY_SET_DISCHARGE_SUPPORTED           uint32 = 0x2
	BATTERY_SET_CHARGINGSOURCE_SUPPORTED      uint32 = 0x4
	BATTERY_SET_CHARGER_ID_SUPPORTED          uint32 = 0x8
	BATTERY_UNKNOWN_TIME                      uint32 = 0xffffffff
	BATTERY_UNKNOWN_CURRENT                   uint32 = 0xffffffff
	UNKNOWN_CURRENT                           uint32 = 0xffffffff
	BATTERY_USB_CHARGER_STATUS_FN_DEFAULT_USB uint32 = 0x1
	BATTERY_USB_CHARGER_STATUS_UCM_PD         uint32 = 0x2
	BATTERY_UNKNOWN_VOLTAGE                   uint32 = 0xffffffff
	BATTERY_UNKNOWN_RATE                      uint32 = 0x80000000
	UNKNOWN_RATE                              uint32 = 0x80000000
	UNKNOWN_VOLTAGE                           uint32 = 0xffffffff
	BATTERY_POWER_ON_LINE                     uint32 = 0x1
	BATTERY_DISCHARGING                       uint32 = 0x2
	BATTERY_CHARGING                          uint32 = 0x4
	BATTERY_CRITICAL                          uint32 = 0x8
	MAX_BATTERY_STRING_SIZE                   uint32 = 0x80
	IOCTL_BATTERY_QUERY_TAG                   uint32 = 0x294040
	IOCTL_BATTERY_QUERY_INFORMATION           uint32 = 0x294044
	IOCTL_BATTERY_SET_INFORMATION             uint32 = 0x298048
	IOCTL_BATTERY_QUERY_STATUS                uint32 = 0x29404c
	IOCTL_BATTERY_CHARGING_SOURCE_CHANGE      uint32 = 0x294050
	BATTERY_TAG_INVALID                       uint32 = 0x0
	MAX_ACTIVE_COOLING_LEVELS                 uint32 = 0xa
	ACTIVE_COOLING                            uint32 = 0x0
	PASSIVE_COOLING                           uint32 = 0x1
	TZ_ACTIVATION_REASON_THERMAL              uint32 = 0x1
	TZ_ACTIVATION_REASON_CURRENT              uint32 = 0x2
	THERMAL_POLICY_VERSION_1                  uint32 = 0x1
	THERMAL_POLICY_VERSION_2                  uint32 = 0x2
	IOCTL_THERMAL_QUERY_INFORMATION           uint32 = 0x294080
	IOCTL_THERMAL_SET_COOLING_POLICY          uint32 = 0x298084
	IOCTL_RUN_ACTIVE_COOLING_METHOD           uint32 = 0x298088
	IOCTL_THERMAL_SET_PASSIVE_LIMIT           uint32 = 0x29808c
	IOCTL_THERMAL_READ_TEMPERATURE            uint32 = 0x294090
	IOCTL_THERMAL_READ_POLICY                 uint32 = 0x294094
	IOCTL_QUERY_LID                           uint32 = 0x2940c0
	IOCTL_NOTIFY_SWITCH_EVENT                 uint32 = 0x294100
	IOCTL_GET_SYS_BUTTON_CAPS                 uint32 = 0x294140
	IOCTL_GET_SYS_BUTTON_EVENT                uint32 = 0x294144
	SYS_BUTTON_POWER                          uint32 = 0x1
	SYS_BUTTON_SLEEP                          uint32 = 0x2
	SYS_BUTTON_LID                            uint32 = 0x4
	SYS_BUTTON_WAKE                           uint32 = 0x80000000
	SYS_BUTTON_LID_STATE_MASK                 uint32 = 0x30000
	SYS_BUTTON_LID_OPEN                       uint32 = 0x10000
	SYS_BUTTON_LID_CLOSED                     uint32 = 0x20000
	SYS_BUTTON_LID_INITIAL                    uint32 = 0x40000
	SYS_BUTTON_LID_CHANGED                    uint32 = 0x80000
	IOCTL_GET_PROCESSOR_OBJ_INFO              uint32 = 0x294180
	THERMAL_COOLING_INTERFACE_VERSION         uint32 = 0x1
	THERMAL_DEVICE_INTERFACE_VERSION          uint32 = 0x1
	IOCTL_SET_SYS_MESSAGE_INDICATOR           uint32 = 0x2981c0
	IOCTL_SET_WAKE_ALARM_VALUE                uint32 = 0x298200
	IOCTL_SET_WAKE_ALARM_POLICY               uint32 = 0x298204
	IOCTL_GET_WAKE_ALARM_VALUE                uint32 = 0x29c208
	IOCTL_GET_WAKE_ALARM_POLICY               uint32 = 0x29c20c
	ACPI_TIME_ADJUST_DAYLIGHT                 uint32 = 0x1
	ACPI_TIME_IN_DAYLIGHT                     uint32 = 0x2
	ACPI_TIME_ZONE_UNKNOWN                    uint32 = 0x7ff
	IOCTL_ACPI_GET_REAL_TIME                  uint32 = 0x294210
	IOCTL_ACPI_SET_REAL_TIME                  uint32 = 0x298214
	IOCTL_GET_WAKE_ALARM_SYSTEM_POWERSTATE    uint32 = 0x294218
	BATTERY_MINIPORT_UPDATE_DATA_VER_1        uint32 = 0x1
	BATTERY_MINIPORT_UPDATE_DATA_VER_2        uint32 = 0x2
	BATTERY_CLASS_MAJOR_VERSION               uint32 = 0x1
	BATTERY_CLASS_MINOR_VERSION               uint32 = 0x0
	BATTERY_CLASS_MINOR_VERSION_1             uint32 = 0x1
	IOCTL_EMI_GET_VERSION                     uint32 = 0x224000
	IOCTL_EMI_GET_METADATA_SIZE               uint32 = 0x224004
	IOCTL_EMI_GET_METADATA                    uint32 = 0x224008
	IOCTL_EMI_GET_MEASUREMENT                 uint32 = 0x22400c
	EMI_NAME_MAX                              uint32 = 0x10
	EMI_VERSION_V1                            uint32 = 0x1
	EMI_VERSION_V2                            uint32 = 0x2
	EFFECTIVE_POWER_MODE_V1                   uint32 = 0x1
	EFFECTIVE_POWER_MODE_V2                   uint32 = 0x2
	EnableSysTrayBatteryMeter                 uint32 = 0x1
	EnableMultiBatteryDisplay                 uint32 = 0x2
	EnablePasswordLogon                       uint32 = 0x4
	EnableWakeOnRing                          uint32 = 0x8
	EnableVideoDimDisplay                     uint32 = 0x10
	POWER_ATTRIBUTE_HIDE                      uint32 = 0x1
	POWER_ATTRIBUTE_SHOW_AOAC                 uint32 = 0x2
	DEVICEPOWER_HARDWAREID                    uint32 = 0x80000000
	DEVICEPOWER_AND_OPERATION                 uint32 = 0x40000000
	DEVICEPOWER_FILTER_DEVICES_PRESENT        uint32 = 0x20000000
	DEVICEPOWER_FILTER_HARDWARE               uint32 = 0x10000000
	DEVICEPOWER_FILTER_WAKEENABLED            uint32 = 0x8000000
	DEVICEPOWER_FILTER_WAKEPROGRAMMABLE       uint32 = 0x4000000
	DEVICEPOWER_FILTER_ON_NAME                uint32 = 0x2000000
	DEVICEPOWER_SET_WAKEENABLED               uint32 = 0x1
	DEVICEPOWER_CLEAR_WAKEENABLED             uint32 = 0x2
	PDCAP_S0_SUPPORTED                        uint32 = 0x10000
	PDCAP_S1_SUPPORTED                        uint32 = 0x20000
	PDCAP_S2_SUPPORTED                        uint32 = 0x40000
	PDCAP_S3_SUPPORTED                        uint32 = 0x80000
	PDCAP_WAKE_FROM_S0_SUPPORTED              uint32 = 0x100000
	PDCAP_WAKE_FROM_S1_SUPPORTED              uint32 = 0x200000
	PDCAP_WAKE_FROM_S2_SUPPORTED              uint32 = 0x400000
	PDCAP_WAKE_FROM_S3_SUPPORTED              uint32 = 0x800000
	PDCAP_S4_SUPPORTED                        uint32 = 0x1000000
	PDCAP_S5_SUPPORTED                        uint32 = 0x2000000
	THERMAL_EVENT_VERSION                     uint32 = 0x1
)

var (
	GUID_DEVICE_BATTERY = syscall.GUID{0x72631E54, 0x78A4, 0x11D0,
		[8]byte{0xBC, 0xF7, 0x00, 0xAA, 0x00, 0xB7, 0xB3, 0x2A}}

	GUID_DEVICE_APPLICATIONLAUNCH_BUTTON = syscall.GUID{0x629758EE, 0x986E, 0x4D9E,
		[8]byte{0x8E, 0x47, 0xDE, 0x27, 0xF8, 0xAB, 0x05, 0x4D}}

	GUID_DEVICE_SYS_BUTTON = syscall.GUID{0x4AFA3D53, 0x74A7, 0x11D0,
		[8]byte{0xBE, 0x5E, 0x00, 0xA0, 0xC9, 0x06, 0x28, 0x57}}

	GUID_DEVICE_LID = syscall.GUID{0x4AFA3D52, 0x74A7, 0x11D0,
		[8]byte{0xBE, 0x5E, 0x00, 0xA0, 0xC9, 0x06, 0x28, 0x57}}

	GUID_DEVICE_THERMAL_ZONE = syscall.GUID{0x4AFA3D51, 0x74A7, 0x11D0,
		[8]byte{0xBE, 0x5E, 0x00, 0xA0, 0xC9, 0x06, 0x28, 0x57}}

	GUID_DEVICE_FAN = syscall.GUID{0x05ECD13D, 0x81DA, 0x4A2A,
		[8]byte{0x8A, 0x4C, 0x52, 0x4F, 0x23, 0xDD, 0x4D, 0xC9}}

	GUID_DEVICE_PROCESSOR = syscall.GUID{0x97FADB10, 0x4E33, 0x40AE,
		[8]byte{0x35, 0x9C, 0x8B, 0xEF, 0x02, 0x9D, 0xBD, 0xD0}}

	GUID_DEVICE_MEMORY = syscall.GUID{0x3FD0F03D, 0x92E0, 0x45FB,
		[8]byte{0xB7, 0x5C, 0x5E, 0xD8, 0xFF, 0xB0, 0x10, 0x21}}

	GUID_DEVICE_ACPI_TIME = syscall.GUID{0x97F99BF6, 0x4497, 0x4F18,
		[8]byte{0xBB, 0x22, 0x4B, 0x9F, 0xB2, 0xFB, 0xEF, 0x9C}}

	GUID_DEVICE_MESSAGE_INDICATOR = syscall.GUID{0xCD48A365, 0xFA94, 0x4CE2,
		[8]byte{0xA2, 0x32, 0xA1, 0xB7, 0x64, 0xE5, 0xD8, 0xB4}}

	GUID_CLASS_INPUT = syscall.GUID{0x4D1E55B2, 0xF16F, 0x11CF,
		[8]byte{0x88, 0xCB, 0x00, 0x11, 0x11, 0x00, 0x00, 0x30}}

	GUID_DEVINTERFACE_THERMAL_COOLING = syscall.GUID{0xDBE4373D, 0x3C81, 0x40CB,
		[8]byte{0xAC, 0xE4, 0xE0, 0xE5, 0xD0, 0x5F, 0x0C, 0x9F}}

	GUID_DEVINTERFACE_THERMAL_MANAGER = syscall.GUID{0x927EC093, 0x69A4, 0x4BC0,
		[8]byte{0xBD, 0x02, 0x71, 0x16, 0x64, 0x71, 0x44, 0x63}}

	BATTERY_STATUS_WMI_GUID = syscall.GUID{0xFC4670D1, 0xEBBF, 0x416E,
		[8]byte{0x87, 0xCE, 0x37, 0x4A, 0x4E, 0xBC, 0x11, 0x1A}}

	BATTERY_RUNTIME_WMI_GUID = syscall.GUID{0x535A3767, 0x1AC2, 0x49BC,
		[8]byte{0xA0, 0x77, 0x3F, 0x7A, 0x02, 0xE4, 0x0A, 0xEC}}

	BATTERY_TEMPERATURE_WMI_GUID = syscall.GUID{0x1A52A14D, 0xADCE, 0x4A44,
		[8]byte{0x9A, 0x3E, 0xC8, 0xD8, 0xF1, 0x5F, 0xF2, 0xC2}}

	BATTERY_FULL_CHARGED_CAPACITY_WMI_GUID = syscall.GUID{0x40B40565, 0x96F7, 0x4435,
		[8]byte{0x86, 0x94, 0x97, 0xE0, 0xE4, 0x39, 0x59, 0x05}}

	BATTERY_CYCLE_COUNT_WMI_GUID = syscall.GUID{0xEF98DB24, 0x0014, 0x4C25,
		[8]byte{0xA5, 0x0B, 0xC7, 0x24, 0xAE, 0x5C, 0xD3, 0x71}}

	BATTERY_STATIC_DATA_WMI_GUID = syscall.GUID{0x05E1E463, 0xE4E2, 0x4EA9,
		[8]byte{0x80, 0xCB, 0x9B, 0xD4, 0xB3, 0xCA, 0x06, 0x55}}

	BATTERY_STATUS_CHANGE_WMI_GUID = syscall.GUID{0xCDDFA0C3, 0x7C5B, 0x4E43,
		[8]byte{0xA0, 0x34, 0x05, 0x9F, 0xA5, 0xB8, 0x43, 0x64}}

	BATTERY_TAG_CHANGE_WMI_GUID = syscall.GUID{0x5E1F6E19, 0x8786, 0x4D23,
		[8]byte{0x94, 0xFC, 0x9E, 0x74, 0x6B, 0xD5, 0xD8, 0x88}}

	GUID_DEVICE_ENERGY_METER = syscall.GUID{0x45BD8344, 0x7ED6, 0x49CF,
		[8]byte{0xA4, 0x40, 0xC2, 0x76, 0xC9, 0x33, 0xB0, 0x53}}
)

// enums

// enum
type POWER_PLATFORM_ROLE_VERSION uint32

const (
	POWER_PLATFORM_ROLE_V1 POWER_PLATFORM_ROLE_VERSION = 1
	POWER_PLATFORM_ROLE_V2 POWER_PLATFORM_ROLE_VERSION = 2
)

// enum
type POWER_SETTING_REGISTER_NOTIFICATION_FLAGS uint32

const (
	DEVICE_NOTIFY_SERVICE_HANDLE POWER_SETTING_REGISTER_NOTIFICATION_FLAGS = 1
	DEVICE_NOTIFY_CALLBACK       POWER_SETTING_REGISTER_NOTIFICATION_FLAGS = 2
	DEVICE_NOTIFY_WINDOW_HANDLE  POWER_SETTING_REGISTER_NOTIFICATION_FLAGS = 0
)

// enum
// flags
type EXECUTION_STATE uint32

const (
	ES_AWAYMODE_REQUIRED EXECUTION_STATE = 64
	ES_CONTINUOUS        EXECUTION_STATE = 2147483648
	ES_DISPLAY_REQUIRED  EXECUTION_STATE = 2
	ES_SYSTEM_REQUIRED   EXECUTION_STATE = 1
	ES_USER_PRESENT      EXECUTION_STATE = 4
)

// enum
// flags
type POWER_ACTION_POLICY_EVENT_CODE uint32

const (
	POWER_FORCE_TRIGGER_RESET     POWER_ACTION_POLICY_EVENT_CODE = 2147483648
	POWER_LEVEL_USER_NOTIFY_EXEC  POWER_ACTION_POLICY_EVENT_CODE = 4
	POWER_LEVEL_USER_NOTIFY_SOUND POWER_ACTION_POLICY_EVENT_CODE = 2
	POWER_LEVEL_USER_NOTIFY_TEXT  POWER_ACTION_POLICY_EVENT_CODE = 1
	POWER_USER_NOTIFY_BUTTON      POWER_ACTION_POLICY_EVENT_CODE = 8
	POWER_USER_NOTIFY_SHUTDOWN    POWER_ACTION_POLICY_EVENT_CODE = 16
)

// enum
type EFFECTIVE_POWER_MODE int32

const (
	EffectivePowerModeBatterySaver    EFFECTIVE_POWER_MODE = 0
	EffectivePowerModeBetterBattery   EFFECTIVE_POWER_MODE = 1
	EffectivePowerModeBalanced        EFFECTIVE_POWER_MODE = 2
	EffectivePowerModeHighPerformance EFFECTIVE_POWER_MODE = 3
	EffectivePowerModeMaxPerformance  EFFECTIVE_POWER_MODE = 4
	EffectivePowerModeGameMode        EFFECTIVE_POWER_MODE = 5
	EffectivePowerModeMixedReality    EFFECTIVE_POWER_MODE = 6
)

// enum
type POWER_DATA_ACCESSOR int32

const (
	ACCESS_AC_POWER_SETTING_INDEX               POWER_DATA_ACCESSOR = 0
	ACCESS_DC_POWER_SETTING_INDEX               POWER_DATA_ACCESSOR = 1
	ACCESS_FRIENDLY_NAME                        POWER_DATA_ACCESSOR = 2
	ACCESS_DESCRIPTION                          POWER_DATA_ACCESSOR = 3
	ACCESS_POSSIBLE_POWER_SETTING               POWER_DATA_ACCESSOR = 4
	ACCESS_POSSIBLE_POWER_SETTING_FRIENDLY_NAME POWER_DATA_ACCESSOR = 5
	ACCESS_POSSIBLE_POWER_SETTING_DESCRIPTION   POWER_DATA_ACCESSOR = 6
	ACCESS_DEFAULT_AC_POWER_SETTING             POWER_DATA_ACCESSOR = 7
	ACCESS_DEFAULT_DC_POWER_SETTING             POWER_DATA_ACCESSOR = 8
	ACCESS_POSSIBLE_VALUE_MIN                   POWER_DATA_ACCESSOR = 9
	ACCESS_POSSIBLE_VALUE_MAX                   POWER_DATA_ACCESSOR = 10
	ACCESS_POSSIBLE_VALUE_INCREMENT             POWER_DATA_ACCESSOR = 11
	ACCESS_POSSIBLE_VALUE_UNITS                 POWER_DATA_ACCESSOR = 12
	ACCESS_ICON_RESOURCE                        POWER_DATA_ACCESSOR = 13
	ACCESS_DEFAULT_SECURITY_DESCRIPTOR          POWER_DATA_ACCESSOR = 14
	ACCESS_ATTRIBUTES                           POWER_DATA_ACCESSOR = 15
	ACCESS_SCHEME                               POWER_DATA_ACCESSOR = 16
	ACCESS_SUBGROUP                             POWER_DATA_ACCESSOR = 17
	ACCESS_INDIVIDUAL_SETTING                   POWER_DATA_ACCESSOR = 18
	ACCESS_ACTIVE_SCHEME                        POWER_DATA_ACCESSOR = 19
	ACCESS_CREATE_SCHEME                        POWER_DATA_ACCESSOR = 20
	ACCESS_AC_POWER_SETTING_MAX                 POWER_DATA_ACCESSOR = 21
	ACCESS_DC_POWER_SETTING_MAX                 POWER_DATA_ACCESSOR = 22
	ACCESS_AC_POWER_SETTING_MIN                 POWER_DATA_ACCESSOR = 23
	ACCESS_DC_POWER_SETTING_MIN                 POWER_DATA_ACCESSOR = 24
	ACCESS_PROFILE                              POWER_DATA_ACCESSOR = 25
	ACCESS_OVERLAY_SCHEME                       POWER_DATA_ACCESSOR = 26
	ACCESS_ACTIVE_OVERLAY_SCHEME                POWER_DATA_ACCESSOR = 27
)

// enum
type BATTERY_QUERY_INFORMATION_LEVEL int32

const (
	BatteryInformation            BATTERY_QUERY_INFORMATION_LEVEL = 0
	BatteryGranularityInformation BATTERY_QUERY_INFORMATION_LEVEL = 1
	BatteryTemperature            BATTERY_QUERY_INFORMATION_LEVEL = 2
	BatteryEstimatedTime          BATTERY_QUERY_INFORMATION_LEVEL = 3
	BatteryDeviceName             BATTERY_QUERY_INFORMATION_LEVEL = 4
	BatteryManufactureDate        BATTERY_QUERY_INFORMATION_LEVEL = 5
	BatteryManufactureName        BATTERY_QUERY_INFORMATION_LEVEL = 6
	BatteryUniqueID               BATTERY_QUERY_INFORMATION_LEVEL = 7
	BatterySerialNumber           BATTERY_QUERY_INFORMATION_LEVEL = 8
)

// enum
type BATTERY_CHARGING_SOURCE_TYPE int32

const (
	BatteryChargingSourceType_AC       BATTERY_CHARGING_SOURCE_TYPE = 1
	BatteryChargingSourceType_USB      BATTERY_CHARGING_SOURCE_TYPE = 2
	BatteryChargingSourceType_Wireless BATTERY_CHARGING_SOURCE_TYPE = 3
	BatteryChargingSourceType_Max      BATTERY_CHARGING_SOURCE_TYPE = 4
)

// enum
type USB_CHARGER_PORT int32

const (
	UsbChargerPort_Legacy USB_CHARGER_PORT = 0
	UsbChargerPort_TypeC  USB_CHARGER_PORT = 1
	UsbChargerPort_Max    USB_CHARGER_PORT = 2
)

// enum
type BATTERY_SET_INFORMATION_LEVEL int32

const (
	BatteryCriticalBias   BATTERY_SET_INFORMATION_LEVEL = 0
	BatteryCharge         BATTERY_SET_INFORMATION_LEVEL = 1
	BatteryDischarge      BATTERY_SET_INFORMATION_LEVEL = 2
	BatteryChargingSource BATTERY_SET_INFORMATION_LEVEL = 3
	BatteryChargerId      BATTERY_SET_INFORMATION_LEVEL = 4
	BatteryChargerStatus  BATTERY_SET_INFORMATION_LEVEL = 5
)

// enum
type EMI_MEASUREMENT_UNIT int32

const (
	EmiMeasurementUnitPicowattHours EMI_MEASUREMENT_UNIT = 0
)

// enum
type SYSTEM_POWER_STATE int32

const (
	PowerSystemUnspecified SYSTEM_POWER_STATE = 0
	PowerSystemWorking     SYSTEM_POWER_STATE = 1
	PowerSystemSleeping1   SYSTEM_POWER_STATE = 2
	PowerSystemSleeping2   SYSTEM_POWER_STATE = 3
	PowerSystemSleeping3   SYSTEM_POWER_STATE = 4
	PowerSystemHibernate   SYSTEM_POWER_STATE = 5
	PowerSystemShutdown    SYSTEM_POWER_STATE = 6
	PowerSystemMaximum     SYSTEM_POWER_STATE = 7
)

// enum
type POWER_ACTION int32

const (
	PowerActionNone          POWER_ACTION = 0
	PowerActionReserved      POWER_ACTION = 1
	PowerActionSleep         POWER_ACTION = 2
	PowerActionHibernate     POWER_ACTION = 3
	PowerActionShutdown      POWER_ACTION = 4
	PowerActionShutdownReset POWER_ACTION = 5
	PowerActionShutdownOff   POWER_ACTION = 6
	PowerActionWarmEject     POWER_ACTION = 7
	PowerActionDisplayOff    POWER_ACTION = 8
)

// enum
type DEVICE_POWER_STATE int32

const (
	PowerDeviceUnspecified DEVICE_POWER_STATE = 0
	PowerDeviceD0          DEVICE_POWER_STATE = 1
	PowerDeviceD1          DEVICE_POWER_STATE = 2
	PowerDeviceD2          DEVICE_POWER_STATE = 3
	PowerDeviceD3          DEVICE_POWER_STATE = 4
	PowerDeviceMaximum     DEVICE_POWER_STATE = 5
)

// enum
type LATENCY_TIME int32

const (
	LT_DONT_CARE      LATENCY_TIME = 0
	LT_LOWEST_LATENCY LATENCY_TIME = 1
)

// enum
type POWER_REQUEST_TYPE int32

const (
	PowerRequestDisplayRequired   POWER_REQUEST_TYPE = 0
	PowerRequestSystemRequired    POWER_REQUEST_TYPE = 1
	PowerRequestAwayModeRequired  POWER_REQUEST_TYPE = 2
	PowerRequestExecutionRequired POWER_REQUEST_TYPE = 3
)

// enum
type POWER_INFORMATION_LEVEL int32

const (
	SystemPowerPolicyAc                POWER_INFORMATION_LEVEL = 0
	SystemPowerPolicyDc                POWER_INFORMATION_LEVEL = 1
	VerifySystemPolicyAc               POWER_INFORMATION_LEVEL = 2
	VerifySystemPolicyDc               POWER_INFORMATION_LEVEL = 3
	SystemPowerCapabilities            POWER_INFORMATION_LEVEL = 4
	SystemBatteryState                 POWER_INFORMATION_LEVEL = 5
	SystemPowerStateHandler            POWER_INFORMATION_LEVEL = 6
	ProcessorStateHandler              POWER_INFORMATION_LEVEL = 7
	SystemPowerPolicyCurrent           POWER_INFORMATION_LEVEL = 8
	AdministratorPowerPolicy           POWER_INFORMATION_LEVEL = 9
	SystemReserveHiberFile             POWER_INFORMATION_LEVEL = 10
	ProcessorInformation               POWER_INFORMATION_LEVEL = 11
	SystemPowerInformation             POWER_INFORMATION_LEVEL = 12
	ProcessorStateHandler2             POWER_INFORMATION_LEVEL = 13
	LastWakeTime                       POWER_INFORMATION_LEVEL = 14
	LastSleepTime                      POWER_INFORMATION_LEVEL = 15
	SystemExecutionState               POWER_INFORMATION_LEVEL = 16
	SystemPowerStateNotifyHandler      POWER_INFORMATION_LEVEL = 17
	ProcessorPowerPolicyAc             POWER_INFORMATION_LEVEL = 18
	ProcessorPowerPolicyDc             POWER_INFORMATION_LEVEL = 19
	VerifyProcessorPowerPolicyAc       POWER_INFORMATION_LEVEL = 20
	VerifyProcessorPowerPolicyDc       POWER_INFORMATION_LEVEL = 21
	ProcessorPowerPolicyCurrent        POWER_INFORMATION_LEVEL = 22
	SystemPowerStateLogging            POWER_INFORMATION_LEVEL = 23
	SystemPowerLoggingEntry            POWER_INFORMATION_LEVEL = 24
	SetPowerSettingValue               POWER_INFORMATION_LEVEL = 25
	NotifyUserPowerSetting             POWER_INFORMATION_LEVEL = 26
	PowerInformationLevelUnused0       POWER_INFORMATION_LEVEL = 27
	SystemMonitorHiberBootPowerOff     POWER_INFORMATION_LEVEL = 28
	SystemVideoState                   POWER_INFORMATION_LEVEL = 29
	TraceApplicationPowerMessage       POWER_INFORMATION_LEVEL = 30
	TraceApplicationPowerMessageEnd    POWER_INFORMATION_LEVEL = 31
	ProcessorPerfStates                POWER_INFORMATION_LEVEL = 32
	ProcessorIdleStates                POWER_INFORMATION_LEVEL = 33
	ProcessorCap                       POWER_INFORMATION_LEVEL = 34
	SystemWakeSource                   POWER_INFORMATION_LEVEL = 35
	SystemHiberFileInformation         POWER_INFORMATION_LEVEL = 36
	TraceServicePowerMessage           POWER_INFORMATION_LEVEL = 37
	ProcessorLoad                      POWER_INFORMATION_LEVEL = 38
	PowerShutdownNotification          POWER_INFORMATION_LEVEL = 39
	MonitorCapabilities                POWER_INFORMATION_LEVEL = 40
	SessionPowerInit                   POWER_INFORMATION_LEVEL = 41
	SessionDisplayState                POWER_INFORMATION_LEVEL = 42
	PowerRequestCreate                 POWER_INFORMATION_LEVEL = 43
	PowerRequestAction                 POWER_INFORMATION_LEVEL = 44
	GetPowerRequestList                POWER_INFORMATION_LEVEL = 45
	ProcessorInformationEx             POWER_INFORMATION_LEVEL = 46
	NotifyUserModeLegacyPowerEvent     POWER_INFORMATION_LEVEL = 47
	GroupPark                          POWER_INFORMATION_LEVEL = 48
	ProcessorIdleDomains               POWER_INFORMATION_LEVEL = 49
	WakeTimerList                      POWER_INFORMATION_LEVEL = 50
	SystemHiberFileSize                POWER_INFORMATION_LEVEL = 51
	ProcessorIdleStatesHv              POWER_INFORMATION_LEVEL = 52
	ProcessorPerfStatesHv              POWER_INFORMATION_LEVEL = 53
	ProcessorPerfCapHv                 POWER_INFORMATION_LEVEL = 54
	ProcessorSetIdle                   POWER_INFORMATION_LEVEL = 55
	LogicalProcessorIdling             POWER_INFORMATION_LEVEL = 56
	UserPresence                       POWER_INFORMATION_LEVEL = 57
	PowerSettingNotificationName       POWER_INFORMATION_LEVEL = 58
	GetPowerSettingValue               POWER_INFORMATION_LEVEL = 59
	IdleResiliency                     POWER_INFORMATION_LEVEL = 60
	SessionRITState                    POWER_INFORMATION_LEVEL = 61
	SessionConnectNotification         POWER_INFORMATION_LEVEL = 62
	SessionPowerCleanup                POWER_INFORMATION_LEVEL = 63
	SessionLockState                   POWER_INFORMATION_LEVEL = 64
	SystemHiberbootState               POWER_INFORMATION_LEVEL = 65
	PlatformInformation                POWER_INFORMATION_LEVEL = 66
	PdcInvocation                      POWER_INFORMATION_LEVEL = 67
	MonitorInvocation                  POWER_INFORMATION_LEVEL = 68
	FirmwareTableInformationRegistered POWER_INFORMATION_LEVEL = 69
	SetShutdownSelectedTime            POWER_INFORMATION_LEVEL = 70
	SuspendResumeInvocation            POWER_INFORMATION_LEVEL = 71
	PlmPowerRequestCreate              POWER_INFORMATION_LEVEL = 72
	ScreenOff                          POWER_INFORMATION_LEVEL = 73
	CsDeviceNotification               POWER_INFORMATION_LEVEL = 74
	PlatformRole                       POWER_INFORMATION_LEVEL = 75
	LastResumePerformance              POWER_INFORMATION_LEVEL = 76
	DisplayBurst                       POWER_INFORMATION_LEVEL = 77
	ExitLatencySamplingPercentage      POWER_INFORMATION_LEVEL = 78
	RegisterSpmPowerSettings           POWER_INFORMATION_LEVEL = 79
	PlatformIdleStates                 POWER_INFORMATION_LEVEL = 80
	ProcessorIdleVeto                  POWER_INFORMATION_LEVEL = 81
	PlatformIdleVeto                   POWER_INFORMATION_LEVEL = 82
	SystemBatteryStatePrecise          POWER_INFORMATION_LEVEL = 83
	ThermalEvent                       POWER_INFORMATION_LEVEL = 84
	PowerRequestActionInternal         POWER_INFORMATION_LEVEL = 85
	BatteryDeviceState                 POWER_INFORMATION_LEVEL = 86
	PowerInformationInternal           POWER_INFORMATION_LEVEL = 87
	ThermalStandby                     POWER_INFORMATION_LEVEL = 88
	SystemHiberFileType                POWER_INFORMATION_LEVEL = 89
	PhysicalPowerButtonPress           POWER_INFORMATION_LEVEL = 90
	QueryPotentialDripsConstraint      POWER_INFORMATION_LEVEL = 91
	EnergyTrackerCreate                POWER_INFORMATION_LEVEL = 92
	EnergyTrackerQuery                 POWER_INFORMATION_LEVEL = 93
	UpdateBlackBoxRecorder             POWER_INFORMATION_LEVEL = 94
	SessionAllowExternalDmaDevices     POWER_INFORMATION_LEVEL = 95
	SendSuspendResumeNotification      POWER_INFORMATION_LEVEL = 96
	PowerInformationLevelMaximum       POWER_INFORMATION_LEVEL = 97
)

// enum
type SYSTEM_POWER_CONDITION int32

const (
	PoAc               SYSTEM_POWER_CONDITION = 0
	PoDc               SYSTEM_POWER_CONDITION = 1
	PoHot              SYSTEM_POWER_CONDITION = 2
	PoConditionMaximum SYSTEM_POWER_CONDITION = 3
)

// enum
type POWER_PLATFORM_ROLE int32

const (
	PlatformRoleUnspecified       POWER_PLATFORM_ROLE = 0
	PlatformRoleDesktop           POWER_PLATFORM_ROLE = 1
	PlatformRoleMobile            POWER_PLATFORM_ROLE = 2
	PlatformRoleWorkstation       POWER_PLATFORM_ROLE = 3
	PlatformRoleEnterpriseServer  POWER_PLATFORM_ROLE = 4
	PlatformRoleSOHOServer        POWER_PLATFORM_ROLE = 5
	PlatformRoleAppliancePC       POWER_PLATFORM_ROLE = 6
	PlatformRolePerformanceServer POWER_PLATFORM_ROLE = 7
	PlatformRoleSlate             POWER_PLATFORM_ROLE = 8
	PlatformRoleMaximum           POWER_PLATFORM_ROLE = 9
)

// structs

type GLOBAL_MACHINE_POWER_POLICY struct {
	Revision                    uint32
	LidOpenWakeAc               SYSTEM_POWER_STATE
	LidOpenWakeDc               SYSTEM_POWER_STATE
	BroadcastCapacityResolution uint32
}

type GLOBAL_USER_POWER_POLICY struct {
	Revision        uint32
	PowerButtonAc   POWER_ACTION_POLICY
	PowerButtonDc   POWER_ACTION_POLICY
	SleepButtonAc   POWER_ACTION_POLICY
	SleepButtonDc   POWER_ACTION_POLICY
	LidCloseAc      POWER_ACTION_POLICY
	LidCloseDc      POWER_ACTION_POLICY
	DischargePolicy [4]SYSTEM_POWER_LEVEL
	GlobalFlags     uint32
}

type GLOBAL_POWER_POLICY struct {
	User GLOBAL_USER_POWER_POLICY
	Mach GLOBAL_MACHINE_POWER_POLICY
}

type MACHINE_POWER_POLICY struct {
	Revision              uint32
	MinSleepAc            SYSTEM_POWER_STATE
	MinSleepDc            SYSTEM_POWER_STATE
	ReducedLatencySleepAc SYSTEM_POWER_STATE
	ReducedLatencySleepDc SYSTEM_POWER_STATE
	DozeTimeoutAc         uint32
	DozeTimeoutDc         uint32
	DozeS4TimeoutAc       uint32
	DozeS4TimeoutDc       uint32
	MinThrottleAc         byte
	MinThrottleDc         byte
	Pad1                  [2]byte
	OverThrottledAc       POWER_ACTION_POLICY
	OverThrottledDc       POWER_ACTION_POLICY
}

type MACHINE_PROCESSOR_POWER_POLICY struct {
	Revision          uint32
	ProcessorPolicyAc PROCESSOR_POWER_POLICY
	ProcessorPolicyDc PROCESSOR_POWER_POLICY
}

type USER_POWER_POLICY struct {
	Revision               uint32
	IdleAc                 POWER_ACTION_POLICY
	IdleDc                 POWER_ACTION_POLICY
	IdleTimeoutAc          uint32
	IdleTimeoutDc          uint32
	IdleSensitivityAc      byte
	IdleSensitivityDc      byte
	ThrottlePolicyAc       byte
	ThrottlePolicyDc       byte
	MaxSleepAc             SYSTEM_POWER_STATE
	MaxSleepDc             SYSTEM_POWER_STATE
	Reserved               [2]uint32
	VideoTimeoutAc         uint32
	VideoTimeoutDc         uint32
	SpindownTimeoutAc      uint32
	SpindownTimeoutDc      uint32
	OptimizeForPowerAc     BOOLEAN
	OptimizeForPowerDc     BOOLEAN
	FanThrottleToleranceAc byte
	FanThrottleToleranceDc byte
	ForcedThrottleAc       byte
	ForcedThrottleDc       byte
}

type POWER_POLICY struct {
	User USER_POWER_POLICY
	Mach MACHINE_POWER_POLICY
}

type DEVICE_NOTIFY_SUBSCRIBE_PARAMETERS struct {
	Callback PDEVICE_NOTIFY_CALLBACK_ROUTINE
	Context  unsafe.Pointer
}

type THERMAL_EVENT struct {
	Version              uint32
	Size                 uint32
	Type                 uint32
	Temperature          uint32
	TripPointTemperature uint32
	Initiator            PWSTR
}

type BATTERY_QUERY_INFORMATION struct {
	BatteryTag       uint32
	InformationLevel BATTERY_QUERY_INFORMATION_LEVEL
	AtRate           uint32
}

type BATTERY_INFORMATION struct {
	Capabilities        uint32
	Technology          byte
	Reserved            [3]byte
	Chemistry           [4]byte
	DesignedCapacity    uint32
	FullChargedCapacity uint32
	DefaultAlert1       uint32
	DefaultAlert2       uint32
	CriticalBias        uint32
	CycleCount          uint32
}

type BATTERY_CHARGING_SOURCE struct {
	Type       BATTERY_CHARGING_SOURCE_TYPE
	MaxCurrent uint32
}

type BATTERY_CHARGING_SOURCE_INFORMATION struct {
	Type         BATTERY_CHARGING_SOURCE_TYPE
	SourceOnline BOOLEAN
}

type BATTERY_SET_INFORMATION struct {
	BatteryTag       uint32
	InformationLevel BATTERY_SET_INFORMATION_LEVEL
	Buffer           [1]byte
}

type BATTERY_CHARGER_STATUS struct {
	Type   BATTERY_CHARGING_SOURCE_TYPE
	VaData [1]uint32
}

type BATTERY_USB_CHARGER_STATUS struct {
	Type                   BATTERY_CHARGING_SOURCE_TYPE
	Reserved               uint32
	Flags                  uint32
	MaxCurrent             uint32
	Voltage                uint32
	PortType               USB_CHARGER_PORT
	PortId                 uint64
	PowerSourceInformation unsafe.Pointer
	OemCharger             syscall.GUID
}

type BATTERY_WAIT_STATUS struct {
	BatteryTag   uint32
	Timeout      uint32
	PowerState   uint32
	LowCapacity  uint32
	HighCapacity uint32
}

type BATTERY_STATUS struct {
	PowerState uint32
	Capacity   uint32
	Voltage    uint32
	Rate       int32
}

type BATTERY_MANUFACTURE_DATE struct {
	Day   byte
	Month byte
	Year  uint16
}

type THERMAL_INFORMATION struct {
	ThermalStamp         uint32
	ThermalConstant1     uint32
	ThermalConstant2     uint32
	Processors           uintptr
	SamplingPeriod       uint32
	CurrentTemperature   uint32
	PassiveTripPoint     uint32
	CriticalTripPoint    uint32
	ActiveTripPointCount byte
	ActiveTripPoint      [10]uint32
}

type THERMAL_WAIT_READ struct {
	Timeout         uint32
	LowTemperature  uint32
	HighTemperature uint32
}

type THERMAL_POLICY struct {
	Version           uint32
	WaitForUpdate     BOOLEAN
	Hibernate         BOOLEAN
	Critical          BOOLEAN
	ThermalStandby    BOOLEAN
	ActivationReasons uint32
	PassiveLimit      uint32
	ActiveLevel       uint32
	OverThrottled     BOOLEAN
}

type PROCESSOR_OBJECT_INFO struct {
	PhysicalID  uint32
	PBlkAddress uint32
	PBlkLength  byte
}

type PROCESSOR_OBJECT_INFO_EX struct {
	PhysicalID    uint32
	PBlkAddress   uint32
	PBlkLength    byte
	InitialApicId uint32
}

type WAKE_ALARM_INFORMATION struct {
	TimerIdentifier uint32
	Timeout         uint32
}

type ACPI_REAL_TIME struct {
	Year         uint16
	Month        byte
	Day          byte
	Hour         byte
	Minute       byte
	Second       byte
	Valid        byte
	Milliseconds uint16
	TimeZone     int16
	DayLight     byte
	Reserved1    [3]byte
}

type EMI_VERSION struct {
	EmiVersion uint16
}

type EMI_METADATA_SIZE struct {
	MetadataSize uint32
}

type EMI_CHANNEL_MEASUREMENT_DATA struct {
	AbsoluteEnergy uint64
	AbsoluteTime   uint64
}

type EMI_METADATA_V1 struct {
	MeasurementUnit         EMI_MEASUREMENT_UNIT
	HardwareOEM             [16]uint16
	HardwareModel           [16]uint16
	HardwareRevision        uint16
	MeteredHardwareNameSize uint16
	MeteredHardwareName     [1]uint16
}

type EMI_CHANNEL_V2 struct {
	MeasurementUnit EMI_MEASUREMENT_UNIT
	ChannelNameSize uint16
	ChannelName     [1]uint16
}

type EMI_METADATA_V2 struct {
	HardwareOEM      [16]uint16
	HardwareModel    [16]uint16
	HardwareRevision uint16
	ChannelCount     uint16
	Channels         [1]EMI_CHANNEL_V2
}

type EMI_MEASUREMENT_DATA_V2 struct {
	ChannelData [1]EMI_CHANNEL_MEASUREMENT_DATA
}

type CM_POWER_DATA struct {
	PD_Size                 uint32
	PD_MostRecentPowerState DEVICE_POWER_STATE
	PD_Capabilities         uint32
	PD_D1Latency            uint32
	PD_D2Latency            uint32
	PD_D3Latency            uint32
	PD_PowerStateMapping    [7]DEVICE_POWER_STATE
	PD_DeepestSystemWake    SYSTEM_POWER_STATE
}

type SET_POWER_SETTING_VALUE struct {
	Version        uint32
	Guid           syscall.GUID
	PowerCondition SYSTEM_POWER_CONDITION
	DataLength     uint32
	Data           [1]byte
}

type BATTERY_REPORTING_SCALE struct {
	Granularity uint32
	Capacity    uint32
}

type POWER_ACTION_POLICY struct {
	Action    POWER_ACTION
	Flags     uint32
	EventCode POWER_ACTION_POLICY_EVENT_CODE
}

type SYSTEM_POWER_LEVEL struct {
	Enable         BOOLEAN
	Spare          [3]byte
	BatteryLevel   uint32
	PowerPolicy    POWER_ACTION_POLICY
	MinSystemState SYSTEM_POWER_STATE
}

type SYSTEM_POWER_POLICY struct {
	Revision                    uint32
	PowerButton                 POWER_ACTION_POLICY
	SleepButton                 POWER_ACTION_POLICY
	LidClose                    POWER_ACTION_POLICY
	LidOpenWake                 SYSTEM_POWER_STATE
	Reserved                    uint32
	Idle                        POWER_ACTION_POLICY
	IdleTimeout                 uint32
	IdleSensitivity             byte
	DynamicThrottle             byte
	Spare2                      [2]byte
	MinSleep                    SYSTEM_POWER_STATE
	MaxSleep                    SYSTEM_POWER_STATE
	ReducedLatencySleep         SYSTEM_POWER_STATE
	WinLogonFlags               uint32
	Spare3                      uint32
	DozeS4Timeout               uint32
	BroadcastCapacityResolution uint32
	DischargePolicy             [4]SYSTEM_POWER_LEVEL
	VideoTimeout                uint32
	VideoDimDisplay             BOOLEAN
	VideoReserved               [3]uint32
	SpindownTimeout             uint32
	OptimizeForPower            BOOLEAN
	FanThrottleTolerance        byte
	ForcedThrottle              byte
	MinThrottle                 byte
	OverThrottled               POWER_ACTION_POLICY
}

type PROCESSOR_POWER_POLICY_INFO struct {
	TimeCheck      uint32
	DemoteLimit    uint32
	PromoteLimit   uint32
	DemotePercent  byte
	PromotePercent byte
	Spare          [2]byte
	Bitfield_      uint32
}

type PROCESSOR_POWER_POLICY struct {
	Revision        uint32
	DynamicThrottle byte
	Spare           [3]byte
	Bitfield_       uint32
	PolicyCount     uint32
	Policy          [3]PROCESSOR_POWER_POLICY_INFO
}

type ADMINISTRATOR_POWER_POLICY struct {
	MinSleep           SYSTEM_POWER_STATE
	MaxSleep           SYSTEM_POWER_STATE
	MinVideoTimeout    uint32
	MaxVideoTimeout    uint32
	MinSpindownTimeout uint32
	MaxSpindownTimeout uint32
}

type SYSTEM_POWER_CAPABILITIES struct {
	PowerButtonPresent        BOOLEAN
	SleepButtonPresent        BOOLEAN
	LidPresent                BOOLEAN
	SystemS1                  BOOLEAN
	SystemS2                  BOOLEAN
	SystemS3                  BOOLEAN
	SystemS4                  BOOLEAN
	SystemS5                  BOOLEAN
	HiberFilePresent          BOOLEAN
	FullWake                  BOOLEAN
	VideoDimPresent           BOOLEAN
	ApmPresent                BOOLEAN
	UpsPresent                BOOLEAN
	ThermalControl            BOOLEAN
	ProcessorThrottle         BOOLEAN
	ProcessorMinThrottle      byte
	ProcessorMaxThrottle      byte
	FastSystemS4              BOOLEAN
	Hiberboot                 BOOLEAN
	WakeAlarmPresent          BOOLEAN
	AoAc                      BOOLEAN
	DiskSpinDown              BOOLEAN
	HiberFileType             byte
	AoAcConnectivitySupported BOOLEAN
	Spare3                    [6]byte
	SystemBatteriesPresent    BOOLEAN
	BatteriesAreShortTerm     BOOLEAN
	BatteryScale              [3]BATTERY_REPORTING_SCALE
	AcOnLineWake              SYSTEM_POWER_STATE
	SoftLidWake               SYSTEM_POWER_STATE
	RtcWake                   SYSTEM_POWER_STATE
	MinDeviceWakeState        SYSTEM_POWER_STATE
	DefaultLowLatencyWake     SYSTEM_POWER_STATE
}

type SYSTEM_BATTERY_STATE struct {
	AcOnLine          BOOLEAN
	BatteryPresent    BOOLEAN
	Charging          BOOLEAN
	Discharging       BOOLEAN
	Spare1            [3]BOOLEAN
	Tag               byte
	MaxCapacity       uint32
	RemainingCapacity uint32
	Rate              uint32
	EstimatedTime     uint32
	DefaultAlert1     uint32
	DefaultAlert2     uint32
}

type POWERBROADCAST_SETTING struct {
	PowerSetting syscall.GUID
	DataLength   uint32
	Data         [1]byte
}

type SYSTEM_POWER_STATUS struct {
	ACLineStatus        byte
	BatteryFlag         byte
	BatteryLifePercent  byte
	SystemStatusFlag    byte
	BatteryLifeTime     uint32
	BatteryFullLifeTime uint32
}

// func types

type EFFECTIVE_POWER_MODE_CALLBACK = uintptr
type EFFECTIVE_POWER_MODE_CALLBACK_func = func(Mode EFFECTIVE_POWER_MODE, Context unsafe.Pointer)

type PWRSCHEMESENUMPROC_V1 = uintptr
type PWRSCHEMESENUMPROC_V1_func = func(Index uint32, NameSize uint32, Name *int8, DescriptionSize uint32, Description *int8, Policy *POWER_POLICY, Context LPARAM) BOOLEAN

type PWRSCHEMESENUMPROC = uintptr
type PWRSCHEMESENUMPROC_func = func(Index uint32, NameSize uint32, Name PWSTR, DescriptionSize uint32, Description PWSTR, Policy *POWER_POLICY, Context LPARAM) BOOLEAN

type PDEVICE_NOTIFY_CALLBACK_ROUTINE = uintptr
type PDEVICE_NOTIFY_CALLBACK_ROUTINE_func = func(Context unsafe.Pointer, Type uint32, Setting unsafe.Pointer) uint32

var (
	pRegisterPowerSettingNotification    uintptr
	pUnregisterPowerSettingNotification  uintptr
	pRegisterSuspendResumeNotification   uintptr
	pUnregisterSuspendResumeNotification uintptr
	pRequestWakeupLatency                uintptr
	pIsSystemResumeAutomatic             uintptr
	pSetThreadExecutionState             uintptr
	pPowerCreateRequest                  uintptr
	pPowerSetRequest                     uintptr
	pPowerClearRequest                   uintptr
	pGetDevicePowerState                 uintptr
	pSetSystemPowerState                 uintptr
	pGetSystemPowerStatus                uintptr
)

func RegisterPowerSettingNotification(hRecipient HANDLE, PowerSettingGuid *syscall.GUID, Flags uint32) (HPOWERNOTIFY, WIN32_ERROR) {
	addr := lazyAddr(&pRegisterPowerSettingNotification, libUser32, "RegisterPowerSettingNotification")
	ret, _, err := syscall.SyscallN(addr, hRecipient, uintptr(unsafe.Pointer(PowerSettingGuid)), uintptr(Flags))
	return ret, WIN32_ERROR(err)
}

func UnregisterPowerSettingNotification(Handle HPOWERNOTIFY) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pUnregisterPowerSettingNotification, libUser32, "UnregisterPowerSettingNotification")
	ret, _, err := syscall.SyscallN(addr, Handle)
	return BOOL(ret), WIN32_ERROR(err)
}

func RegisterSuspendResumeNotification(hRecipient HANDLE, Flags uint32) (HPOWERNOTIFY, WIN32_ERROR) {
	addr := lazyAddr(&pRegisterSuspendResumeNotification, libUser32, "RegisterSuspendResumeNotification")
	ret, _, err := syscall.SyscallN(addr, hRecipient, uintptr(Flags))
	return ret, WIN32_ERROR(err)
}

func UnregisterSuspendResumeNotification(Handle HPOWERNOTIFY) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pUnregisterSuspendResumeNotification, libUser32, "UnregisterSuspendResumeNotification")
	ret, _, err := syscall.SyscallN(addr, Handle)
	return BOOL(ret), WIN32_ERROR(err)
}

func RequestWakeupLatency(latency LATENCY_TIME) BOOL {
	addr := lazyAddr(&pRequestWakeupLatency, libKernel32, "RequestWakeupLatency")
	ret, _, _ := syscall.SyscallN(addr, uintptr(latency))
	return BOOL(ret)
}

func IsSystemResumeAutomatic() BOOL {
	addr := lazyAddr(&pIsSystemResumeAutomatic, libKernel32, "IsSystemResumeAutomatic")
	ret, _, _ := syscall.SyscallN(addr)
	return BOOL(ret)
}

func SetThreadExecutionState(esFlags EXECUTION_STATE) EXECUTION_STATE {
	addr := lazyAddr(&pSetThreadExecutionState, libKernel32, "SetThreadExecutionState")
	ret, _, _ := syscall.SyscallN(addr, uintptr(esFlags))
	return EXECUTION_STATE(ret)
}

func PowerCreateRequest(Context *REASON_CONTEXT) (HANDLE, WIN32_ERROR) {
	addr := lazyAddr(&pPowerCreateRequest, libKernel32, "PowerCreateRequest")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(Context)))
	return ret, WIN32_ERROR(err)
}

func PowerSetRequest(PowerRequest HANDLE, RequestType POWER_REQUEST_TYPE) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pPowerSetRequest, libKernel32, "PowerSetRequest")
	ret, _, err := syscall.SyscallN(addr, PowerRequest, uintptr(RequestType))
	return BOOL(ret), WIN32_ERROR(err)
}

func PowerClearRequest(PowerRequest HANDLE, RequestType POWER_REQUEST_TYPE) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pPowerClearRequest, libKernel32, "PowerClearRequest")
	ret, _, err := syscall.SyscallN(addr, PowerRequest, uintptr(RequestType))
	return BOOL(ret), WIN32_ERROR(err)
}

func GetDevicePowerState(hDevice HANDLE, pfOn *BOOL) BOOL {
	addr := lazyAddr(&pGetDevicePowerState, libKernel32, "GetDevicePowerState")
	ret, _, _ := syscall.SyscallN(addr, hDevice, uintptr(unsafe.Pointer(pfOn)))
	return BOOL(ret)
}

func SetSystemPowerState(fSuspend BOOL, fForce BOOL) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pSetSystemPowerState, libKernel32, "SetSystemPowerState")
	ret, _, err := syscall.SyscallN(addr, uintptr(fSuspend), uintptr(fForce))
	return BOOL(ret), WIN32_ERROR(err)
}

func GetSystemPowerStatus(lpSystemPowerStatus *SYSTEM_POWER_STATUS) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pGetSystemPowerStatus, libKernel32, "GetSystemPowerStatus")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpSystemPowerStatus)))
	return BOOL(ret), WIN32_ERROR(err)
}
