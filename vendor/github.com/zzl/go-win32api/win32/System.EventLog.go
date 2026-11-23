package win32

import (
	"syscall"
	"unsafe"
)

type (
	EventLogHandle    = uintptr
	EventSourceHandle = uintptr
)

const (
	EVT_VARIANT_TYPE_MASK  uint32 = 0x7f
	EVT_VARIANT_TYPE_ARRAY uint32 = 0x80
	EVT_READ_ACCESS        uint32 = 0x1
	EVT_WRITE_ACCESS       uint32 = 0x2
	EVT_CLEAR_ACCESS       uint32 = 0x4
	EVT_ALL_ACCESS         uint32 = 0x7
)

// enums

// enum
type REPORT_EVENT_TYPE uint16

const (
	EVENTLOG_SUCCESS          REPORT_EVENT_TYPE = 0
	EVENTLOG_AUDIT_FAILURE    REPORT_EVENT_TYPE = 16
	EVENTLOG_AUDIT_SUCCESS    REPORT_EVENT_TYPE = 8
	EVENTLOG_ERROR_TYPE       REPORT_EVENT_TYPE = 1
	EVENTLOG_INFORMATION_TYPE REPORT_EVENT_TYPE = 4
	EVENTLOG_WARNING_TYPE     REPORT_EVENT_TYPE = 2
)

// enum
type READ_EVENT_LOG_READ_FLAGS uint32

const (
	EVENTLOG_SEEK_READ       READ_EVENT_LOG_READ_FLAGS = 2
	EVENTLOG_SEQUENTIAL_READ READ_EVENT_LOG_READ_FLAGS = 1
)

// enum
type EVT_VARIANT_TYPE int32

const (
	EvtVarTypeNull       EVT_VARIANT_TYPE = 0
	EvtVarTypeString     EVT_VARIANT_TYPE = 1
	EvtVarTypeAnsiString EVT_VARIANT_TYPE = 2
	EvtVarTypeSByte      EVT_VARIANT_TYPE = 3
	EvtVarTypeByte       EVT_VARIANT_TYPE = 4
	EvtVarTypeInt16      EVT_VARIANT_TYPE = 5
	EvtVarTypeUInt16     EVT_VARIANT_TYPE = 6
	EvtVarTypeInt32      EVT_VARIANT_TYPE = 7
	EvtVarTypeUInt32     EVT_VARIANT_TYPE = 8
	EvtVarTypeInt64      EVT_VARIANT_TYPE = 9
	EvtVarTypeUInt64     EVT_VARIANT_TYPE = 10
	EvtVarTypeSingle     EVT_VARIANT_TYPE = 11
	EvtVarTypeDouble     EVT_VARIANT_TYPE = 12
	EvtVarTypeBoolean    EVT_VARIANT_TYPE = 13
	EvtVarTypeBinary     EVT_VARIANT_TYPE = 14
	EvtVarTypeGuid       EVT_VARIANT_TYPE = 15
	EvtVarTypeSizeT      EVT_VARIANT_TYPE = 16
	EvtVarTypeFileTime   EVT_VARIANT_TYPE = 17
	EvtVarTypeSysTime    EVT_VARIANT_TYPE = 18
	EvtVarTypeSid        EVT_VARIANT_TYPE = 19
	EvtVarTypeHexInt32   EVT_VARIANT_TYPE = 20
	EvtVarTypeHexInt64   EVT_VARIANT_TYPE = 21
	EvtVarTypeEvtHandle  EVT_VARIANT_TYPE = 32
	EvtVarTypeEvtXml     EVT_VARIANT_TYPE = 35
)

// enum
type EVT_LOGIN_CLASS int32

const (
	EvtRpcLogin EVT_LOGIN_CLASS = 1
)

// enum
type EVT_RPC_LOGIN_FLAGS int32

const (
	EvtRpcLoginAuthDefault   EVT_RPC_LOGIN_FLAGS = 0
	EvtRpcLoginAuthNegotiate EVT_RPC_LOGIN_FLAGS = 1
	EvtRpcLoginAuthKerberos  EVT_RPC_LOGIN_FLAGS = 2
	EvtRpcLoginAuthNTLM      EVT_RPC_LOGIN_FLAGS = 3
)

// enum
type EVT_QUERY_FLAGS int32

const (
	EvtQueryChannelPath         EVT_QUERY_FLAGS = 1
	EvtQueryFilePath            EVT_QUERY_FLAGS = 2
	EvtQueryForwardDirection    EVT_QUERY_FLAGS = 256
	EvtQueryReverseDirection    EVT_QUERY_FLAGS = 512
	EvtQueryTolerateQueryErrors EVT_QUERY_FLAGS = 4096
)

// enum
type EVT_SEEK_FLAGS int32

const (
	EvtSeekRelativeToFirst    EVT_SEEK_FLAGS = 1
	EvtSeekRelativeToLast     EVT_SEEK_FLAGS = 2
	EvtSeekRelativeToCurrent  EVT_SEEK_FLAGS = 3
	EvtSeekRelativeToBookmark EVT_SEEK_FLAGS = 4
	EvtSeekOriginMask         EVT_SEEK_FLAGS = 7
	EvtSeekStrict             EVT_SEEK_FLAGS = 65536
)

// enum
type EVT_SUBSCRIBE_FLAGS int32

const (
	EvtSubscribeToFutureEvents      EVT_SUBSCRIBE_FLAGS = 1
	EvtSubscribeStartAtOldestRecord EVT_SUBSCRIBE_FLAGS = 2
	EvtSubscribeStartAfterBookmark  EVT_SUBSCRIBE_FLAGS = 3
	EvtSubscribeOriginMask          EVT_SUBSCRIBE_FLAGS = 3
	EvtSubscribeTolerateQueryErrors EVT_SUBSCRIBE_FLAGS = 4096
	EvtSubscribeStrict              EVT_SUBSCRIBE_FLAGS = 65536
)

// enum
type EVT_SUBSCRIBE_NOTIFY_ACTION int32

const (
	EvtSubscribeActionError   EVT_SUBSCRIBE_NOTIFY_ACTION = 0
	EvtSubscribeActionDeliver EVT_SUBSCRIBE_NOTIFY_ACTION = 1
)

// enum
type EVT_SYSTEM_PROPERTY_ID int32

const (
	EvtSystemProviderName      EVT_SYSTEM_PROPERTY_ID = 0
	EvtSystemProviderGuid      EVT_SYSTEM_PROPERTY_ID = 1
	EvtSystemEventID           EVT_SYSTEM_PROPERTY_ID = 2
	EvtSystemQualifiers        EVT_SYSTEM_PROPERTY_ID = 3
	EvtSystemLevel             EVT_SYSTEM_PROPERTY_ID = 4
	EvtSystemTask              EVT_SYSTEM_PROPERTY_ID = 5
	EvtSystemOpcode            EVT_SYSTEM_PROPERTY_ID = 6
	EvtSystemKeywords          EVT_SYSTEM_PROPERTY_ID = 7
	EvtSystemTimeCreated       EVT_SYSTEM_PROPERTY_ID = 8
	EvtSystemEventRecordId     EVT_SYSTEM_PROPERTY_ID = 9
	EvtSystemActivityID        EVT_SYSTEM_PROPERTY_ID = 10
	EvtSystemRelatedActivityID EVT_SYSTEM_PROPERTY_ID = 11
	EvtSystemProcessID         EVT_SYSTEM_PROPERTY_ID = 12
	EvtSystemThreadID          EVT_SYSTEM_PROPERTY_ID = 13
	EvtSystemChannel           EVT_SYSTEM_PROPERTY_ID = 14
	EvtSystemComputer          EVT_SYSTEM_PROPERTY_ID = 15
	EvtSystemUserID            EVT_SYSTEM_PROPERTY_ID = 16
	EvtSystemVersion           EVT_SYSTEM_PROPERTY_ID = 17
	EvtSystemPropertyIdEND     EVT_SYSTEM_PROPERTY_ID = 18
)

// enum
type EVT_RENDER_CONTEXT_FLAGS int32

const (
	EvtRenderContextValues EVT_RENDER_CONTEXT_FLAGS = 0
	EvtRenderContextSystem EVT_RENDER_CONTEXT_FLAGS = 1
	EvtRenderContextUser   EVT_RENDER_CONTEXT_FLAGS = 2
)

// enum
type EVT_RENDER_FLAGS int32

const (
	EvtRenderEventValues EVT_RENDER_FLAGS = 0
	EvtRenderEventXml    EVT_RENDER_FLAGS = 1
	EvtRenderBookmark    EVT_RENDER_FLAGS = 2
)

// enum
type EVT_FORMAT_MESSAGE_FLAGS int32

const (
	EvtFormatMessageEvent    EVT_FORMAT_MESSAGE_FLAGS = 1
	EvtFormatMessageLevel    EVT_FORMAT_MESSAGE_FLAGS = 2
	EvtFormatMessageTask     EVT_FORMAT_MESSAGE_FLAGS = 3
	EvtFormatMessageOpcode   EVT_FORMAT_MESSAGE_FLAGS = 4
	EvtFormatMessageKeyword  EVT_FORMAT_MESSAGE_FLAGS = 5
	EvtFormatMessageChannel  EVT_FORMAT_MESSAGE_FLAGS = 6
	EvtFormatMessageProvider EVT_FORMAT_MESSAGE_FLAGS = 7
	EvtFormatMessageId       EVT_FORMAT_MESSAGE_FLAGS = 8
	EvtFormatMessageXml      EVT_FORMAT_MESSAGE_FLAGS = 9
)

// enum
type EVT_OPEN_LOG_FLAGS int32

const (
	EvtOpenChannelPath EVT_OPEN_LOG_FLAGS = 1
	EvtOpenFilePath    EVT_OPEN_LOG_FLAGS = 2
)

// enum
type EVT_LOG_PROPERTY_ID int32

const (
	EvtLogCreationTime       EVT_LOG_PROPERTY_ID = 0
	EvtLogLastAccessTime     EVT_LOG_PROPERTY_ID = 1
	EvtLogLastWriteTime      EVT_LOG_PROPERTY_ID = 2
	EvtLogFileSize           EVT_LOG_PROPERTY_ID = 3
	EvtLogAttributes         EVT_LOG_PROPERTY_ID = 4
	EvtLogNumberOfLogRecords EVT_LOG_PROPERTY_ID = 5
	EvtLogOldestRecordNumber EVT_LOG_PROPERTY_ID = 6
	EvtLogFull               EVT_LOG_PROPERTY_ID = 7
)

// enum
type EVT_EXPORTLOG_FLAGS int32

const (
	EvtExportLogChannelPath         EVT_EXPORTLOG_FLAGS = 1
	EvtExportLogFilePath            EVT_EXPORTLOG_FLAGS = 2
	EvtExportLogTolerateQueryErrors EVT_EXPORTLOG_FLAGS = 4096
	EvtExportLogOverwrite           EVT_EXPORTLOG_FLAGS = 8192
)

// enum
type EVT_CHANNEL_CONFIG_PROPERTY_ID int32

const (
	EvtChannelConfigEnabled               EVT_CHANNEL_CONFIG_PROPERTY_ID = 0
	EvtChannelConfigIsolation             EVT_CHANNEL_CONFIG_PROPERTY_ID = 1
	EvtChannelConfigType                  EVT_CHANNEL_CONFIG_PROPERTY_ID = 2
	EvtChannelConfigOwningPublisher       EVT_CHANNEL_CONFIG_PROPERTY_ID = 3
	EvtChannelConfigClassicEventlog       EVT_CHANNEL_CONFIG_PROPERTY_ID = 4
	EvtChannelConfigAccess                EVT_CHANNEL_CONFIG_PROPERTY_ID = 5
	EvtChannelLoggingConfigRetention      EVT_CHANNEL_CONFIG_PROPERTY_ID = 6
	EvtChannelLoggingConfigAutoBackup     EVT_CHANNEL_CONFIG_PROPERTY_ID = 7
	EvtChannelLoggingConfigMaxSize        EVT_CHANNEL_CONFIG_PROPERTY_ID = 8
	EvtChannelLoggingConfigLogFilePath    EVT_CHANNEL_CONFIG_PROPERTY_ID = 9
	EvtChannelPublishingConfigLevel       EVT_CHANNEL_CONFIG_PROPERTY_ID = 10
	EvtChannelPublishingConfigKeywords    EVT_CHANNEL_CONFIG_PROPERTY_ID = 11
	EvtChannelPublishingConfigControlGuid EVT_CHANNEL_CONFIG_PROPERTY_ID = 12
	EvtChannelPublishingConfigBufferSize  EVT_CHANNEL_CONFIG_PROPERTY_ID = 13
	EvtChannelPublishingConfigMinBuffers  EVT_CHANNEL_CONFIG_PROPERTY_ID = 14
	EvtChannelPublishingConfigMaxBuffers  EVT_CHANNEL_CONFIG_PROPERTY_ID = 15
	EvtChannelPublishingConfigLatency     EVT_CHANNEL_CONFIG_PROPERTY_ID = 16
	EvtChannelPublishingConfigClockType   EVT_CHANNEL_CONFIG_PROPERTY_ID = 17
	EvtChannelPublishingConfigSidType     EVT_CHANNEL_CONFIG_PROPERTY_ID = 18
	EvtChannelPublisherList               EVT_CHANNEL_CONFIG_PROPERTY_ID = 19
	EvtChannelPublishingConfigFileMax     EVT_CHANNEL_CONFIG_PROPERTY_ID = 20
	EvtChannelConfigPropertyIdEND         EVT_CHANNEL_CONFIG_PROPERTY_ID = 21
)

// enum
type EVT_CHANNEL_TYPE int32

const (
	EvtChannelTypeAdmin       EVT_CHANNEL_TYPE = 0
	EvtChannelTypeOperational EVT_CHANNEL_TYPE = 1
	EvtChannelTypeAnalytic    EVT_CHANNEL_TYPE = 2
	EvtChannelTypeDebug       EVT_CHANNEL_TYPE = 3
)

// enum
type EVT_CHANNEL_ISOLATION_TYPE int32

const (
	EvtChannelIsolationTypeApplication EVT_CHANNEL_ISOLATION_TYPE = 0
	EvtChannelIsolationTypeSystem      EVT_CHANNEL_ISOLATION_TYPE = 1
	EvtChannelIsolationTypeCustom      EVT_CHANNEL_ISOLATION_TYPE = 2
)

// enum
type EVT_CHANNEL_CLOCK_TYPE int32

const (
	EvtChannelClockTypeSystemTime EVT_CHANNEL_CLOCK_TYPE = 0
	EvtChannelClockTypeQPC        EVT_CHANNEL_CLOCK_TYPE = 1
)

// enum
type EVT_CHANNEL_SID_TYPE int32

const (
	EvtChannelSidTypeNone       EVT_CHANNEL_SID_TYPE = 0
	EvtChannelSidTypePublishing EVT_CHANNEL_SID_TYPE = 1
)

// enum
type EVT_CHANNEL_REFERENCE_FLAGS int32

const (
	EvtChannelReferenceImported EVT_CHANNEL_REFERENCE_FLAGS = 1
)

// enum
type EVT_PUBLISHER_METADATA_PROPERTY_ID int32

const (
	EvtPublisherMetadataPublisherGuid             EVT_PUBLISHER_METADATA_PROPERTY_ID = 0
	EvtPublisherMetadataResourceFilePath          EVT_PUBLISHER_METADATA_PROPERTY_ID = 1
	EvtPublisherMetadataParameterFilePath         EVT_PUBLISHER_METADATA_PROPERTY_ID = 2
	EvtPublisherMetadataMessageFilePath           EVT_PUBLISHER_METADATA_PROPERTY_ID = 3
	EvtPublisherMetadataHelpLink                  EVT_PUBLISHER_METADATA_PROPERTY_ID = 4
	EvtPublisherMetadataPublisherMessageID        EVT_PUBLISHER_METADATA_PROPERTY_ID = 5
	EvtPublisherMetadataChannelReferences         EVT_PUBLISHER_METADATA_PROPERTY_ID = 6
	EvtPublisherMetadataChannelReferencePath      EVT_PUBLISHER_METADATA_PROPERTY_ID = 7
	EvtPublisherMetadataChannelReferenceIndex     EVT_PUBLISHER_METADATA_PROPERTY_ID = 8
	EvtPublisherMetadataChannelReferenceID        EVT_PUBLISHER_METADATA_PROPERTY_ID = 9
	EvtPublisherMetadataChannelReferenceFlags     EVT_PUBLISHER_METADATA_PROPERTY_ID = 10
	EvtPublisherMetadataChannelReferenceMessageID EVT_PUBLISHER_METADATA_PROPERTY_ID = 11
	EvtPublisherMetadataLevels                    EVT_PUBLISHER_METADATA_PROPERTY_ID = 12
	EvtPublisherMetadataLevelName                 EVT_PUBLISHER_METADATA_PROPERTY_ID = 13
	EvtPublisherMetadataLevelValue                EVT_PUBLISHER_METADATA_PROPERTY_ID = 14
	EvtPublisherMetadataLevelMessageID            EVT_PUBLISHER_METADATA_PROPERTY_ID = 15
	EvtPublisherMetadataTasks                     EVT_PUBLISHER_METADATA_PROPERTY_ID = 16
	EvtPublisherMetadataTaskName                  EVT_PUBLISHER_METADATA_PROPERTY_ID = 17
	EvtPublisherMetadataTaskEventGuid             EVT_PUBLISHER_METADATA_PROPERTY_ID = 18
	EvtPublisherMetadataTaskValue                 EVT_PUBLISHER_METADATA_PROPERTY_ID = 19
	EvtPublisherMetadataTaskMessageID             EVT_PUBLISHER_METADATA_PROPERTY_ID = 20
	EvtPublisherMetadataOpcodes                   EVT_PUBLISHER_METADATA_PROPERTY_ID = 21
	EvtPublisherMetadataOpcodeName                EVT_PUBLISHER_METADATA_PROPERTY_ID = 22
	EvtPublisherMetadataOpcodeValue               EVT_PUBLISHER_METADATA_PROPERTY_ID = 23
	EvtPublisherMetadataOpcodeMessageID           EVT_PUBLISHER_METADATA_PROPERTY_ID = 24
	EvtPublisherMetadataKeywords                  EVT_PUBLISHER_METADATA_PROPERTY_ID = 25
	EvtPublisherMetadataKeywordName               EVT_PUBLISHER_METADATA_PROPERTY_ID = 26
	EvtPublisherMetadataKeywordValue              EVT_PUBLISHER_METADATA_PROPERTY_ID = 27
	EvtPublisherMetadataKeywordMessageID          EVT_PUBLISHER_METADATA_PROPERTY_ID = 28
	EvtPublisherMetadataPropertyIdEND             EVT_PUBLISHER_METADATA_PROPERTY_ID = 29
)

// enum
type EVT_EVENT_METADATA_PROPERTY_ID int32

const (
	EventMetadataEventID          EVT_EVENT_METADATA_PROPERTY_ID = 0
	EventMetadataEventVersion     EVT_EVENT_METADATA_PROPERTY_ID = 1
	EventMetadataEventChannel     EVT_EVENT_METADATA_PROPERTY_ID = 2
	EventMetadataEventLevel       EVT_EVENT_METADATA_PROPERTY_ID = 3
	EventMetadataEventOpcode      EVT_EVENT_METADATA_PROPERTY_ID = 4
	EventMetadataEventTask        EVT_EVENT_METADATA_PROPERTY_ID = 5
	EventMetadataEventKeyword     EVT_EVENT_METADATA_PROPERTY_ID = 6
	EventMetadataEventMessageID   EVT_EVENT_METADATA_PROPERTY_ID = 7
	EventMetadataEventTemplate    EVT_EVENT_METADATA_PROPERTY_ID = 8
	EvtEventMetadataPropertyIdEND EVT_EVENT_METADATA_PROPERTY_ID = 9
)

// enum
type EVT_QUERY_PROPERTY_ID int32

const (
	EvtQueryNames         EVT_QUERY_PROPERTY_ID = 0
	EvtQueryStatuses      EVT_QUERY_PROPERTY_ID = 1
	EvtQueryPropertyIdEND EVT_QUERY_PROPERTY_ID = 2
)

// enum
type EVT_EVENT_PROPERTY_ID int32

const (
	EvtEventQueryIDs      EVT_EVENT_PROPERTY_ID = 0
	EvtEventPath          EVT_EVENT_PROPERTY_ID = 1
	EvtEventPropertyIdEND EVT_EVENT_PROPERTY_ID = 2
)

// structs

type EVT_VARIANT_Anonymous struct {
	Data [1]uint64
}

func (this *EVT_VARIANT_Anonymous) BooleanVal() *BOOL {
	return (*BOOL)(unsafe.Pointer(this))
}

func (this *EVT_VARIANT_Anonymous) BooleanValVal() BOOL {
	return *(*BOOL)(unsafe.Pointer(this))
}

func (this *EVT_VARIANT_Anonymous) SByteVal() *int8 {
	return (*int8)(unsafe.Pointer(this))
}

func (this *EVT_VARIANT_Anonymous) SByteValVal() int8 {
	return *(*int8)(unsafe.Pointer(this))
}

func (this *EVT_VARIANT_Anonymous) Int16Val() *int16 {
	return (*int16)(unsafe.Pointer(this))
}

func (this *EVT_VARIANT_Anonymous) Int16ValVal() int16 {
	return *(*int16)(unsafe.Pointer(this))
}

func (this *EVT_VARIANT_Anonymous) Int32Val() *int32 {
	return (*int32)(unsafe.Pointer(this))
}

func (this *EVT_VARIANT_Anonymous) Int32ValVal() int32 {
	return *(*int32)(unsafe.Pointer(this))
}

func (this *EVT_VARIANT_Anonymous) Int64Val() *int64 {
	return (*int64)(unsafe.Pointer(this))
}

func (this *EVT_VARIANT_Anonymous) Int64ValVal() int64 {
	return *(*int64)(unsafe.Pointer(this))
}

func (this *EVT_VARIANT_Anonymous) ByteVal() *byte {
	return (*byte)(unsafe.Pointer(this))
}

func (this *EVT_VARIANT_Anonymous) ByteValVal() byte {
	return *(*byte)(unsafe.Pointer(this))
}

func (this *EVT_VARIANT_Anonymous) UInt16Val() *uint16 {
	return (*uint16)(unsafe.Pointer(this))
}

func (this *EVT_VARIANT_Anonymous) UInt16ValVal() uint16 {
	return *(*uint16)(unsafe.Pointer(this))
}

func (this *EVT_VARIANT_Anonymous) UInt32Val() *uint32 {
	return (*uint32)(unsafe.Pointer(this))
}

func (this *EVT_VARIANT_Anonymous) UInt32ValVal() uint32 {
	return *(*uint32)(unsafe.Pointer(this))
}

func (this *EVT_VARIANT_Anonymous) UInt64Val() *uint64 {
	return (*uint64)(unsafe.Pointer(this))
}

func (this *EVT_VARIANT_Anonymous) UInt64ValVal() uint64 {
	return *(*uint64)(unsafe.Pointer(this))
}

func (this *EVT_VARIANT_Anonymous) SingleVal() *float32 {
	return (*float32)(unsafe.Pointer(this))
}

func (this *EVT_VARIANT_Anonymous) SingleValVal() float32 {
	return *(*float32)(unsafe.Pointer(this))
}

func (this *EVT_VARIANT_Anonymous) DoubleVal() *float64 {
	return (*float64)(unsafe.Pointer(this))
}

func (this *EVT_VARIANT_Anonymous) DoubleValVal() float64 {
	return *(*float64)(unsafe.Pointer(this))
}

func (this *EVT_VARIANT_Anonymous) FileTimeVal() *uint64 {
	return (*uint64)(unsafe.Pointer(this))
}

func (this *EVT_VARIANT_Anonymous) FileTimeValVal() uint64 {
	return *(*uint64)(unsafe.Pointer(this))
}

func (this *EVT_VARIANT_Anonymous) SysTimeVal() **SYSTEMTIME {
	return (**SYSTEMTIME)(unsafe.Pointer(this))
}

func (this *EVT_VARIANT_Anonymous) SysTimeValVal() *SYSTEMTIME {
	return *(**SYSTEMTIME)(unsafe.Pointer(this))
}

func (this *EVT_VARIANT_Anonymous) GuidVal() **syscall.GUID {
	return (**syscall.GUID)(unsafe.Pointer(this))
}

func (this *EVT_VARIANT_Anonymous) GuidValVal() *syscall.GUID {
	return *(**syscall.GUID)(unsafe.Pointer(this))
}

func (this *EVT_VARIANT_Anonymous) StringVal() *PWSTR {
	return (*PWSTR)(unsafe.Pointer(this))
}

func (this *EVT_VARIANT_Anonymous) StringValVal() PWSTR {
	return *(*PWSTR)(unsafe.Pointer(this))
}

func (this *EVT_VARIANT_Anonymous) AnsiStringVal() *PSTR {
	return (*PSTR)(unsafe.Pointer(this))
}

func (this *EVT_VARIANT_Anonymous) AnsiStringValVal() PSTR {
	return *(*PSTR)(unsafe.Pointer(this))
}

func (this *EVT_VARIANT_Anonymous) BinaryVal() **byte {
	return (**byte)(unsafe.Pointer(this))
}

func (this *EVT_VARIANT_Anonymous) BinaryValVal() *byte {
	return *(**byte)(unsafe.Pointer(this))
}

func (this *EVT_VARIANT_Anonymous) SidVal() *PSID {
	return (*PSID)(unsafe.Pointer(this))
}

func (this *EVT_VARIANT_Anonymous) SidValVal() PSID {
	return *(*PSID)(unsafe.Pointer(this))
}

func (this *EVT_VARIANT_Anonymous) SizeTVal() *uintptr {
	return (*uintptr)(unsafe.Pointer(this))
}

func (this *EVT_VARIANT_Anonymous) SizeTValVal() uintptr {
	return *(*uintptr)(unsafe.Pointer(this))
}

func (this *EVT_VARIANT_Anonymous) BooleanArr() **BOOL {
	return (**BOOL)(unsafe.Pointer(this))
}

func (this *EVT_VARIANT_Anonymous) BooleanArrVal() *BOOL {
	return *(**BOOL)(unsafe.Pointer(this))
}

func (this *EVT_VARIANT_Anonymous) SByteArr() **int8 {
	return (**int8)(unsafe.Pointer(this))
}

func (this *EVT_VARIANT_Anonymous) SByteArrVal() *int8 {
	return *(**int8)(unsafe.Pointer(this))
}

func (this *EVT_VARIANT_Anonymous) Int16Arr() **int16 {
	return (**int16)(unsafe.Pointer(this))
}

func (this *EVT_VARIANT_Anonymous) Int16ArrVal() *int16 {
	return *(**int16)(unsafe.Pointer(this))
}

func (this *EVT_VARIANT_Anonymous) Int32Arr() **int32 {
	return (**int32)(unsafe.Pointer(this))
}

func (this *EVT_VARIANT_Anonymous) Int32ArrVal() *int32 {
	return *(**int32)(unsafe.Pointer(this))
}

func (this *EVT_VARIANT_Anonymous) Int64Arr() **int64 {
	return (**int64)(unsafe.Pointer(this))
}

func (this *EVT_VARIANT_Anonymous) Int64ArrVal() *int64 {
	return *(**int64)(unsafe.Pointer(this))
}

func (this *EVT_VARIANT_Anonymous) ByteArr() **byte {
	return (**byte)(unsafe.Pointer(this))
}

func (this *EVT_VARIANT_Anonymous) ByteArrVal() *byte {
	return *(**byte)(unsafe.Pointer(this))
}

func (this *EVT_VARIANT_Anonymous) UInt16Arr() **uint16 {
	return (**uint16)(unsafe.Pointer(this))
}

func (this *EVT_VARIANT_Anonymous) UInt16ArrVal() *uint16 {
	return *(**uint16)(unsafe.Pointer(this))
}

func (this *EVT_VARIANT_Anonymous) UInt32Arr() **uint32 {
	return (**uint32)(unsafe.Pointer(this))
}

func (this *EVT_VARIANT_Anonymous) UInt32ArrVal() *uint32 {
	return *(**uint32)(unsafe.Pointer(this))
}

func (this *EVT_VARIANT_Anonymous) UInt64Arr() **uint64 {
	return (**uint64)(unsafe.Pointer(this))
}

func (this *EVT_VARIANT_Anonymous) UInt64ArrVal() *uint64 {
	return *(**uint64)(unsafe.Pointer(this))
}

func (this *EVT_VARIANT_Anonymous) SingleArr() **float32 {
	return (**float32)(unsafe.Pointer(this))
}

func (this *EVT_VARIANT_Anonymous) SingleArrVal() *float32 {
	return *(**float32)(unsafe.Pointer(this))
}

func (this *EVT_VARIANT_Anonymous) DoubleArr() **float64 {
	return (**float64)(unsafe.Pointer(this))
}

func (this *EVT_VARIANT_Anonymous) DoubleArrVal() *float64 {
	return *(**float64)(unsafe.Pointer(this))
}

func (this *EVT_VARIANT_Anonymous) FileTimeArr() **FILETIME {
	return (**FILETIME)(unsafe.Pointer(this))
}

func (this *EVT_VARIANT_Anonymous) FileTimeArrVal() *FILETIME {
	return *(**FILETIME)(unsafe.Pointer(this))
}

func (this *EVT_VARIANT_Anonymous) SysTimeArr() **SYSTEMTIME {
	return (**SYSTEMTIME)(unsafe.Pointer(this))
}

func (this *EVT_VARIANT_Anonymous) SysTimeArrVal() *SYSTEMTIME {
	return *(**SYSTEMTIME)(unsafe.Pointer(this))
}

func (this *EVT_VARIANT_Anonymous) GuidArr() **syscall.GUID {
	return (**syscall.GUID)(unsafe.Pointer(this))
}

func (this *EVT_VARIANT_Anonymous) GuidArrVal() *syscall.GUID {
	return *(**syscall.GUID)(unsafe.Pointer(this))
}

func (this *EVT_VARIANT_Anonymous) StringArr() **PWSTR {
	return (**PWSTR)(unsafe.Pointer(this))
}

func (this *EVT_VARIANT_Anonymous) StringArrVal() *PWSTR {
	return *(**PWSTR)(unsafe.Pointer(this))
}

func (this *EVT_VARIANT_Anonymous) AnsiStringArr() **PSTR {
	return (**PSTR)(unsafe.Pointer(this))
}

func (this *EVT_VARIANT_Anonymous) AnsiStringArrVal() *PSTR {
	return *(**PSTR)(unsafe.Pointer(this))
}

func (this *EVT_VARIANT_Anonymous) SidArr() **PSID {
	return (**PSID)(unsafe.Pointer(this))
}

func (this *EVT_VARIANT_Anonymous) SidArrVal() *PSID {
	return *(**PSID)(unsafe.Pointer(this))
}

func (this *EVT_VARIANT_Anonymous) SizeTArr() **uintptr {
	return (**uintptr)(unsafe.Pointer(this))
}

func (this *EVT_VARIANT_Anonymous) SizeTArrVal() *uintptr {
	return *(**uintptr)(unsafe.Pointer(this))
}

func (this *EVT_VARIANT_Anonymous) EvtHandleVal() *uintptr {
	return (*uintptr)(unsafe.Pointer(this))
}

func (this *EVT_VARIANT_Anonymous) EvtHandleValVal() uintptr {
	return *(*uintptr)(unsafe.Pointer(this))
}

func (this *EVT_VARIANT_Anonymous) XmlVal() *PWSTR {
	return (*PWSTR)(unsafe.Pointer(this))
}

func (this *EVT_VARIANT_Anonymous) XmlValVal() PWSTR {
	return *(*PWSTR)(unsafe.Pointer(this))
}

func (this *EVT_VARIANT_Anonymous) XmlValArr() **PWSTR {
	return (**PWSTR)(unsafe.Pointer(this))
}

func (this *EVT_VARIANT_Anonymous) XmlValArrVal() *PWSTR {
	return *(**PWSTR)(unsafe.Pointer(this))
}

type EVT_VARIANT struct {
	EVT_VARIANT_Anonymous
	Count uint32
	Type  uint32
}

type EVT_RPC_LOGIN struct {
	Server   PWSTR
	User     PWSTR
	Domain   PWSTR
	Password PWSTR
	Flags    uint32
}

type EVENTLOGRECORD struct {
	Length              uint32
	Reserved            uint32
	RecordNumber        uint32
	TimeGenerated       uint32
	TimeWritten         uint32
	EventID             uint32
	EventType           REPORT_EVENT_TYPE
	NumStrings          uint16
	EventCategory       uint16
	ReservedFlags       uint16
	ClosingRecordNumber uint32
	StringOffset        uint32
	UserSidLength       uint32
	UserSidOffset       uint32
	DataLength          uint32
	DataOffset          uint32
}

type EVENTSFORLOGFILE struct {
	UlSize           uint32
	SzLogicalLogFile [256]uint16
	UlNumRecords     uint32
	PEventLogRecords [1]EVENTLOGRECORD
}

type EVENTLOG_FULL_INFORMATION struct {
	DwFull uint32
}

// func types

type EVT_SUBSCRIBE_CALLBACK = uintptr
type EVT_SUBSCRIBE_CALLBACK_func = func(Action EVT_SUBSCRIBE_NOTIFY_ACTION, UserContext unsafe.Pointer, Event uintptr) uint32

var (
	pClearEventLogA             uintptr
	pClearEventLogW             uintptr
	pBackupEventLogA            uintptr
	pBackupEventLogW            uintptr
	pCloseEventLog              uintptr
	pDeregisterEventSource      uintptr
	pNotifyChangeEventLog       uintptr
	pGetNumberOfEventLogRecords uintptr
	pGetOldestEventLogRecord    uintptr
	pOpenEventLogA              uintptr
	pOpenEventLogW              uintptr
	pRegisterEventSourceA       uintptr
	pRegisterEventSourceW       uintptr
	pOpenBackupEventLogA        uintptr
	pOpenBackupEventLogW        uintptr
	pReadEventLogA              uintptr
	pReadEventLogW              uintptr
	pReportEventA               uintptr
	pReportEventW               uintptr
	pGetEventLogInformation     uintptr
)

func ClearEventLogA(hEventLog EventLogHandle, lpBackupFileName PSTR) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pClearEventLogA, libAdvapi32, "ClearEventLogA")
	ret, _, err := syscall.SyscallN(addr, hEventLog, uintptr(unsafe.Pointer(lpBackupFileName)))
	return BOOL(ret), WIN32_ERROR(err)
}

var ClearEventLog = ClearEventLogW

func ClearEventLogW(hEventLog EventLogHandle, lpBackupFileName PWSTR) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pClearEventLogW, libAdvapi32, "ClearEventLogW")
	ret, _, err := syscall.SyscallN(addr, hEventLog, uintptr(unsafe.Pointer(lpBackupFileName)))
	return BOOL(ret), WIN32_ERROR(err)
}

func BackupEventLogA(hEventLog EventLogHandle, lpBackupFileName PSTR) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pBackupEventLogA, libAdvapi32, "BackupEventLogA")
	ret, _, err := syscall.SyscallN(addr, hEventLog, uintptr(unsafe.Pointer(lpBackupFileName)))
	return BOOL(ret), WIN32_ERROR(err)
}

var BackupEventLog = BackupEventLogW

func BackupEventLogW(hEventLog EventLogHandle, lpBackupFileName PWSTR) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pBackupEventLogW, libAdvapi32, "BackupEventLogW")
	ret, _, err := syscall.SyscallN(addr, hEventLog, uintptr(unsafe.Pointer(lpBackupFileName)))
	return BOOL(ret), WIN32_ERROR(err)
}

func CloseEventLog(hEventLog EventLogHandle) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pCloseEventLog, libAdvapi32, "CloseEventLog")
	ret, _, err := syscall.SyscallN(addr, hEventLog)
	return BOOL(ret), WIN32_ERROR(err)
}

func DeregisterEventSource(hEventLog EventSourceHandle) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pDeregisterEventSource, libAdvapi32, "DeregisterEventSource")
	ret, _, err := syscall.SyscallN(addr, hEventLog)
	return BOOL(ret), WIN32_ERROR(err)
}

func NotifyChangeEventLog(hEventLog EventLogHandle, hEvent HANDLE) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pNotifyChangeEventLog, libAdvapi32, "NotifyChangeEventLog")
	ret, _, err := syscall.SyscallN(addr, hEventLog, hEvent)
	return BOOL(ret), WIN32_ERROR(err)
}

func GetNumberOfEventLogRecords(hEventLog EventLogHandle, NumberOfRecords *uint32) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pGetNumberOfEventLogRecords, libAdvapi32, "GetNumberOfEventLogRecords")
	ret, _, err := syscall.SyscallN(addr, hEventLog, uintptr(unsafe.Pointer(NumberOfRecords)))
	return BOOL(ret), WIN32_ERROR(err)
}

func GetOldestEventLogRecord(hEventLog EventLogHandle, OldestRecord *uint32) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pGetOldestEventLogRecord, libAdvapi32, "GetOldestEventLogRecord")
	ret, _, err := syscall.SyscallN(addr, hEventLog, uintptr(unsafe.Pointer(OldestRecord)))
	return BOOL(ret), WIN32_ERROR(err)
}

func OpenEventLogA(lpUNCServerName PSTR, lpSourceName PSTR) (EventLogHandle, WIN32_ERROR) {
	addr := lazyAddr(&pOpenEventLogA, libAdvapi32, "OpenEventLogA")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpUNCServerName)), uintptr(unsafe.Pointer(lpSourceName)))
	return ret, WIN32_ERROR(err)
}

var OpenEventLog = OpenEventLogW

func OpenEventLogW(lpUNCServerName PWSTR, lpSourceName PWSTR) (EventLogHandle, WIN32_ERROR) {
	addr := lazyAddr(&pOpenEventLogW, libAdvapi32, "OpenEventLogW")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpUNCServerName)), uintptr(unsafe.Pointer(lpSourceName)))
	return ret, WIN32_ERROR(err)
}

func RegisterEventSourceA(lpUNCServerName PSTR, lpSourceName PSTR) (EventSourceHandle, WIN32_ERROR) {
	addr := lazyAddr(&pRegisterEventSourceA, libAdvapi32, "RegisterEventSourceA")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpUNCServerName)), uintptr(unsafe.Pointer(lpSourceName)))
	return ret, WIN32_ERROR(err)
}

var RegisterEventSource = RegisterEventSourceW

func RegisterEventSourceW(lpUNCServerName PWSTR, lpSourceName PWSTR) (EventSourceHandle, WIN32_ERROR) {
	addr := lazyAddr(&pRegisterEventSourceW, libAdvapi32, "RegisterEventSourceW")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpUNCServerName)), uintptr(unsafe.Pointer(lpSourceName)))
	return ret, WIN32_ERROR(err)
}

func OpenBackupEventLogA(lpUNCServerName PSTR, lpFileName PSTR) (EventLogHandle, WIN32_ERROR) {
	addr := lazyAddr(&pOpenBackupEventLogA, libAdvapi32, "OpenBackupEventLogA")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpUNCServerName)), uintptr(unsafe.Pointer(lpFileName)))
	return ret, WIN32_ERROR(err)
}

var OpenBackupEventLog = OpenBackupEventLogW

func OpenBackupEventLogW(lpUNCServerName PWSTR, lpFileName PWSTR) (EventLogHandle, WIN32_ERROR) {
	addr := lazyAddr(&pOpenBackupEventLogW, libAdvapi32, "OpenBackupEventLogW")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpUNCServerName)), uintptr(unsafe.Pointer(lpFileName)))
	return ret, WIN32_ERROR(err)
}

func ReadEventLogA(hEventLog EventLogHandle, dwReadFlags READ_EVENT_LOG_READ_FLAGS, dwRecordOffset uint32, lpBuffer unsafe.Pointer, nNumberOfBytesToRead uint32, pnBytesRead *uint32, pnMinNumberOfBytesNeeded *uint32) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pReadEventLogA, libAdvapi32, "ReadEventLogA")
	ret, _, err := syscall.SyscallN(addr, hEventLog, uintptr(dwReadFlags), uintptr(dwRecordOffset), uintptr(lpBuffer), uintptr(nNumberOfBytesToRead), uintptr(unsafe.Pointer(pnBytesRead)), uintptr(unsafe.Pointer(pnMinNumberOfBytesNeeded)))
	return BOOL(ret), WIN32_ERROR(err)
}

var ReadEventLog = ReadEventLogW

func ReadEventLogW(hEventLog EventLogHandle, dwReadFlags READ_EVENT_LOG_READ_FLAGS, dwRecordOffset uint32, lpBuffer unsafe.Pointer, nNumberOfBytesToRead uint32, pnBytesRead *uint32, pnMinNumberOfBytesNeeded *uint32) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pReadEventLogW, libAdvapi32, "ReadEventLogW")
	ret, _, err := syscall.SyscallN(addr, hEventLog, uintptr(dwReadFlags), uintptr(dwRecordOffset), uintptr(lpBuffer), uintptr(nNumberOfBytesToRead), uintptr(unsafe.Pointer(pnBytesRead)), uintptr(unsafe.Pointer(pnMinNumberOfBytesNeeded)))
	return BOOL(ret), WIN32_ERROR(err)
}

func ReportEventA(hEventLog EventSourceHandle, wType REPORT_EVENT_TYPE, wCategory uint16, dwEventID uint32, lpUserSid PSID, wNumStrings uint16, dwDataSize uint32, lpStrings *PSTR, lpRawData unsafe.Pointer) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pReportEventA, libAdvapi32, "ReportEventA")
	ret, _, err := syscall.SyscallN(addr, hEventLog, uintptr(wType), uintptr(wCategory), uintptr(dwEventID), uintptr(unsafe.Pointer(lpUserSid)), uintptr(wNumStrings), uintptr(dwDataSize), uintptr(unsafe.Pointer(lpStrings)), uintptr(lpRawData))
	return BOOL(ret), WIN32_ERROR(err)
}

var ReportEvent = ReportEventW

func ReportEventW(hEventLog EventSourceHandle, wType REPORT_EVENT_TYPE, wCategory uint16, dwEventID uint32, lpUserSid PSID, wNumStrings uint16, dwDataSize uint32, lpStrings *PWSTR, lpRawData unsafe.Pointer) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pReportEventW, libAdvapi32, "ReportEventW")
	ret, _, err := syscall.SyscallN(addr, hEventLog, uintptr(wType), uintptr(wCategory), uintptr(dwEventID), uintptr(unsafe.Pointer(lpUserSid)), uintptr(wNumStrings), uintptr(dwDataSize), uintptr(unsafe.Pointer(lpStrings)), uintptr(lpRawData))
	return BOOL(ret), WIN32_ERROR(err)
}

func GetEventLogInformation(hEventLog EventLogHandle, dwInfoLevel uint32, lpBuffer unsafe.Pointer, cbBufSize uint32, pcbBytesNeeded *uint32) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pGetEventLogInformation, libAdvapi32, "GetEventLogInformation")
	ret, _, err := syscall.SyscallN(addr, hEventLog, uintptr(dwInfoLevel), uintptr(lpBuffer), uintptr(cbBufSize), uintptr(unsafe.Pointer(pcbBytesNeeded)))
	return BOOL(ret), WIN32_ERROR(err)
}
