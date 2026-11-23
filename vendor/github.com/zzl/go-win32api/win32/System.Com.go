package win32

import (
	"syscall"
	"unsafe"
)

type (
	CO_MTA_USAGE_COOKIE      = uintptr
	CO_DEVICE_CATALOG_COOKIE = uintptr
)

const (
	MARSHALINTERFACE_MIN                                   uint32 = 0x1f4
	ASYNC_MODE_COMPATIBILITY                               int32  = 1
	ASYNC_MODE_DEFAULT                                     int32  = 0
	STGTY_REPEAT                                           int32  = 256
	STG_TOEND                                              int32  = -1
	STG_LAYOUT_SEQUENTIAL                                  int32  = 0
	STG_LAYOUT_INTERLEAVED                                 int32  = 1
	CRT_INTERNAL_COMBASE_SYMBOL_PREFIX_                    string = "_"
	COM_RIGHTS_EXECUTE                                     uint32 = 0x1
	COM_RIGHTS_EXECUTE_LOCAL                               uint32 = 0x2
	COM_RIGHTS_EXECUTE_REMOTE                              uint32 = 0x4
	COM_RIGHTS_ACTIVATE_LOCAL                              uint32 = 0x8
	COM_RIGHTS_ACTIVATE_REMOTE                             uint32 = 0x10
	COM_RIGHTS_RESERVED1                                   uint32 = 0x20
	COM_RIGHTS_RESERVED2                                   uint32 = 0x40
	CWMO_MAX_HANDLES                                       uint32 = 0x38
	ROTREGFLAGS_ALLOWANYCLIENT                             uint32 = 0x1
	APPIDREGFLAGS_ACTIVATE_IUSERVER_INDESKTOP              uint32 = 0x1
	APPIDREGFLAGS_SECURE_SERVER_PROCESS_SD_AND_BIND        uint32 = 0x2
	APPIDREGFLAGS_ISSUE_ACTIVATION_RPC_AT_IDENTIFY         uint32 = 0x4
	APPIDREGFLAGS_IUSERVER_UNMODIFIED_LOGON_TOKEN          uint32 = 0x8
	APPIDREGFLAGS_IUSERVER_SELF_SID_IN_LAUNCH_PERMISSION   uint32 = 0x10
	APPIDREGFLAGS_IUSERVER_ACTIVATE_IN_CLIENT_SESSION_ONLY uint32 = 0x20
	APPIDREGFLAGS_RESERVED1                                uint32 = 0x40
	APPIDREGFLAGS_RESERVED2                                uint32 = 0x80
	APPIDREGFLAGS_RESERVED3                                uint32 = 0x100
	APPIDREGFLAGS_RESERVED4                                uint32 = 0x200
	APPIDREGFLAGS_RESERVED5                                uint32 = 0x400
	APPIDREGFLAGS_AAA_NO_IMPLICIT_ACTIVATE_AS_IU           uint32 = 0x800
	APPIDREGFLAGS_RESERVED7                                uint32 = 0x1000
	APPIDREGFLAGS_RESERVED8                                uint32 = 0x2000
	APPIDREGFLAGS_RESERVED9                                uint32 = 0x4000
	DCOMSCM_ACTIVATION_USE_ALL_AUTHNSERVICES               uint32 = 0x1
	DCOMSCM_ACTIVATION_DISALLOW_UNSECURE_CALL              uint32 = 0x2
	DCOMSCM_RESOLVE_USE_ALL_AUTHNSERVICES                  uint32 = 0x4
	DCOMSCM_RESOLVE_DISALLOW_UNSECURE_CALL                 uint32 = 0x8
	DCOMSCM_PING_USE_MID_AUTHNSERVICE                      uint32 = 0x10
	DCOMSCM_PING_DISALLOW_UNSECURE_CALL                    uint32 = 0x20
	MAXLSN                                                 uint64 = 0x7fffffffffffffff
	DMUS_ERRBASE                                           uint32 = 0x1000
)

// enums

// enum
// flags
type URI_CREATE_FLAGS uint32

const (
	Uri_CREATE_ALLOW_RELATIVE                 URI_CREATE_FLAGS = 1
	Uri_CREATE_ALLOW_IMPLICIT_WILDCARD_SCHEME URI_CREATE_FLAGS = 2
	Uri_CREATE_ALLOW_IMPLICIT_FILE_SCHEME     URI_CREATE_FLAGS = 4
	Uri_CREATE_NOFRAG                         URI_CREATE_FLAGS = 8
	Uri_CREATE_NO_CANONICALIZE                URI_CREATE_FLAGS = 16
	Uri_CREATE_CANONICALIZE                   URI_CREATE_FLAGS = 256
	Uri_CREATE_FILE_USE_DOS_PATH              URI_CREATE_FLAGS = 32
	Uri_CREATE_DECODE_EXTRA_INFO              URI_CREATE_FLAGS = 64
	Uri_CREATE_NO_DECODE_EXTRA_INFO           URI_CREATE_FLAGS = 128
	Uri_CREATE_CRACK_UNKNOWN_SCHEMES          URI_CREATE_FLAGS = 512
	Uri_CREATE_NO_CRACK_UNKNOWN_SCHEMES       URI_CREATE_FLAGS = 1024
	Uri_CREATE_PRE_PROCESS_HTML_URI           URI_CREATE_FLAGS = 2048
	Uri_CREATE_NO_PRE_PROCESS_HTML_URI        URI_CREATE_FLAGS = 4096
	Uri_CREATE_IE_SETTINGS                    URI_CREATE_FLAGS = 8192
	Uri_CREATE_NO_IE_SETTINGS                 URI_CREATE_FLAGS = 16384
	Uri_CREATE_NO_ENCODE_FORBIDDEN_CHARACTERS URI_CREATE_FLAGS = 32768
	Uri_CREATE_NORMALIZE_INTL_CHARACTERS      URI_CREATE_FLAGS = 65536
	Uri_CREATE_CANONICALIZE_ABSOLUTE          URI_CREATE_FLAGS = 131072
)

// enum
type RPC_C_AUTHN_LEVEL uint32

const (
	RPC_C_AUTHN_LEVEL_DEFAULT       RPC_C_AUTHN_LEVEL = 0
	RPC_C_AUTHN_LEVEL_NONE          RPC_C_AUTHN_LEVEL = 1
	RPC_C_AUTHN_LEVEL_CONNECT       RPC_C_AUTHN_LEVEL = 2
	RPC_C_AUTHN_LEVEL_CALL          RPC_C_AUTHN_LEVEL = 3
	RPC_C_AUTHN_LEVEL_PKT           RPC_C_AUTHN_LEVEL = 4
	RPC_C_AUTHN_LEVEL_PKT_INTEGRITY RPC_C_AUTHN_LEVEL = 5
	RPC_C_AUTHN_LEVEL_PKT_PRIVACY   RPC_C_AUTHN_LEVEL = 6
)

// enum
type RPC_C_IMP_LEVEL uint32

const (
	RPC_C_IMP_LEVEL_DEFAULT     RPC_C_IMP_LEVEL = 0
	RPC_C_IMP_LEVEL_ANONYMOUS   RPC_C_IMP_LEVEL = 1
	RPC_C_IMP_LEVEL_IDENTIFY    RPC_C_IMP_LEVEL = 2
	RPC_C_IMP_LEVEL_IMPERSONATE RPC_C_IMP_LEVEL = 3
	RPC_C_IMP_LEVEL_DELEGATE    RPC_C_IMP_LEVEL = 4
)

// enum
type DVASPECT int32

const (
	DVASPECT_CONTENT   DVASPECT = 1
	DVASPECT_THUMBNAIL DVASPECT = 2
	DVASPECT_ICON      DVASPECT = 4
	DVASPECT_DOCPRINT  DVASPECT = 8
)

// enum
// flags
type STGC uint32

const (
	STGC_DEFAULT                            STGC = 0
	STGC_OVERWRITE                          STGC = 1
	STGC_ONLYIFCURRENT                      STGC = 2
	STGC_DANGEROUSLYCOMMITMERELYTODISKCACHE STGC = 4
	STGC_CONSOLIDATE                        STGC = 8
)

// enum
type TYSPEC int32

const (
	TYSPEC_CLSID       TYSPEC = 0
	TYSPEC_FILEEXT     TYSPEC = 1
	TYSPEC_MIMETYPE    TYSPEC = 2
	TYSPEC_FILENAME    TYSPEC = 3
	TYSPEC_PROGID      TYSPEC = 4
	TYSPEC_PACKAGENAME TYSPEC = 5
	TYSPEC_OBJECTID    TYSPEC = 6
)

// enum
type REGCLS int32

const (
	REGCLS_SINGLEUSE      REGCLS = 0
	REGCLS_MULTIPLEUSE    REGCLS = 1
	REGCLS_MULTI_SEPARATE REGCLS = 2
	REGCLS_SUSPENDED      REGCLS = 4
	REGCLS_SURROGATE      REGCLS = 8
	REGCLS_AGILE          REGCLS = 16
)

// enum
type COINITBASE int32

const (
	COINITBASE_MULTITHREADED COINITBASE = 0
)

// enum
type MEMCTX int32

const (
	MEMCTX_TASK      MEMCTX = 1
	MEMCTX_SHARED    MEMCTX = 2
	MEMCTX_MACSYSTEM MEMCTX = 3
	MEMCTX_UNKNOWN   MEMCTX = -1
	MEMCTX_SAME      MEMCTX = -2
)

// enum
// flags
type CLSCTX uint32

const (
	CLSCTX_INPROC_SERVER          CLSCTX = 1
	CLSCTX_INPROC_HANDLER         CLSCTX = 2
	CLSCTX_LOCAL_SERVER           CLSCTX = 4
	CLSCTX_INPROC_SERVER16        CLSCTX = 8
	CLSCTX_REMOTE_SERVER          CLSCTX = 16
	CLSCTX_INPROC_HANDLER16       CLSCTX = 32
	CLSCTX_RESERVED1              CLSCTX = 64
	CLSCTX_RESERVED2              CLSCTX = 128
	CLSCTX_RESERVED3              CLSCTX = 256
	CLSCTX_RESERVED4              CLSCTX = 512
	CLSCTX_NO_CODE_DOWNLOAD       CLSCTX = 1024
	CLSCTX_RESERVED5              CLSCTX = 2048
	CLSCTX_NO_CUSTOM_MARSHAL      CLSCTX = 4096
	CLSCTX_ENABLE_CODE_DOWNLOAD   CLSCTX = 8192
	CLSCTX_NO_FAILURE_LOG         CLSCTX = 16384
	CLSCTX_DISABLE_AAA            CLSCTX = 32768
	CLSCTX_ENABLE_AAA             CLSCTX = 65536
	CLSCTX_FROM_DEFAULT_CONTEXT   CLSCTX = 131072
	CLSCTX_ACTIVATE_X86_SERVER    CLSCTX = 262144
	CLSCTX_ACTIVATE_32_BIT_SERVER CLSCTX = 262144
	CLSCTX_ACTIVATE_64_BIT_SERVER CLSCTX = 524288
	CLSCTX_ENABLE_CLOAKING        CLSCTX = 1048576
	CLSCTX_APPCONTAINER           CLSCTX = 4194304
	CLSCTX_ACTIVATE_AAA_AS_IU     CLSCTX = 8388608
	CLSCTX_RESERVED6              CLSCTX = 16777216
	CLSCTX_ACTIVATE_ARM32_SERVER  CLSCTX = 33554432
	CLSCTX_PS_DLL                 CLSCTX = 2147483648
	CLSCTX_ALL                    CLSCTX = 23
	CLSCTX_SERVER                 CLSCTX = 21
)

// enum
type MSHLFLAGS int32

const (
	MSHLFLAGS_NORMAL      MSHLFLAGS = 0
	MSHLFLAGS_TABLESTRONG MSHLFLAGS = 1
	MSHLFLAGS_TABLEWEAK   MSHLFLAGS = 2
	MSHLFLAGS_NOPING      MSHLFLAGS = 4
	MSHLFLAGS_RESERVED1   MSHLFLAGS = 8
	MSHLFLAGS_RESERVED2   MSHLFLAGS = 16
	MSHLFLAGS_RESERVED3   MSHLFLAGS = 32
	MSHLFLAGS_RESERVED4   MSHLFLAGS = 64
)

// enum
type MSHCTX int32

const (
	MSHCTX_LOCAL            MSHCTX = 0
	MSHCTX_NOSHAREDMEM      MSHCTX = 1
	MSHCTX_DIFFERENTMACHINE MSHCTX = 2
	MSHCTX_INPROC           MSHCTX = 3
	MSHCTX_CROSSCTX         MSHCTX = 4
	MSHCTX_CONTAINER        MSHCTX = 5
)

// enum
type EXTCONN int32

const (
	EXTCONN_STRONG   EXTCONN = 1
	EXTCONN_WEAK     EXTCONN = 2
	EXTCONN_CALLABLE EXTCONN = 4
)

// enum
type STGTY int32

const (
	STGTY_STORAGE   STGTY = 1
	STGTY_STREAM    STGTY = 2
	STGTY_LOCKBYTES STGTY = 3
	STGTY_PROPERTY  STGTY = 4
)

// enum
type STREAM_SEEK uint32

const (
	STREAM_SEEK_SET STREAM_SEEK = 0
	STREAM_SEEK_CUR STREAM_SEEK = 1
	STREAM_SEEK_END STREAM_SEEK = 2
)

// enum
type EOLE_AUTHENTICATION_CAPABILITIES int32

const (
	EOAC_NONE              EOLE_AUTHENTICATION_CAPABILITIES = 0
	EOAC_MUTUAL_AUTH       EOLE_AUTHENTICATION_CAPABILITIES = 1
	EOAC_STATIC_CLOAKING   EOLE_AUTHENTICATION_CAPABILITIES = 32
	EOAC_DYNAMIC_CLOAKING  EOLE_AUTHENTICATION_CAPABILITIES = 64
	EOAC_ANY_AUTHORITY     EOLE_AUTHENTICATION_CAPABILITIES = 128
	EOAC_MAKE_FULLSIC      EOLE_AUTHENTICATION_CAPABILITIES = 256
	EOAC_DEFAULT           EOLE_AUTHENTICATION_CAPABILITIES = 2048
	EOAC_SECURE_REFS       EOLE_AUTHENTICATION_CAPABILITIES = 2
	EOAC_ACCESS_CONTROL    EOLE_AUTHENTICATION_CAPABILITIES = 4
	EOAC_APPID             EOLE_AUTHENTICATION_CAPABILITIES = 8
	EOAC_DYNAMIC           EOLE_AUTHENTICATION_CAPABILITIES = 16
	EOAC_REQUIRE_FULLSIC   EOLE_AUTHENTICATION_CAPABILITIES = 512
	EOAC_AUTO_IMPERSONATE  EOLE_AUTHENTICATION_CAPABILITIES = 1024
	EOAC_DISABLE_AAA       EOLE_AUTHENTICATION_CAPABILITIES = 4096
	EOAC_NO_CUSTOM_MARSHAL EOLE_AUTHENTICATION_CAPABILITIES = 8192
	EOAC_RESERVED1         EOLE_AUTHENTICATION_CAPABILITIES = 16384
)

// enum
type RPCOPT_PROPERTIES int32

const (
	COMBND_RPCTIMEOUT      RPCOPT_PROPERTIES = 1
	COMBND_SERVER_LOCALITY RPCOPT_PROPERTIES = 2
	COMBND_RESERVED1       RPCOPT_PROPERTIES = 4
	COMBND_RESERVED2       RPCOPT_PROPERTIES = 5
	COMBND_RESERVED3       RPCOPT_PROPERTIES = 8
	COMBND_RESERVED4       RPCOPT_PROPERTIES = 16
)

// enum
type RPCOPT_SERVER_LOCALITY_VALUES int32

const (
	SERVER_LOCALITY_PROCESS_LOCAL RPCOPT_SERVER_LOCALITY_VALUES = 0
	SERVER_LOCALITY_MACHINE_LOCAL RPCOPT_SERVER_LOCALITY_VALUES = 1
	SERVER_LOCALITY_REMOTE        RPCOPT_SERVER_LOCALITY_VALUES = 2
)

// enum
type GLOBALOPT_PROPERTIES int32

const (
	COMGLB_EXCEPTION_HANDLING     GLOBALOPT_PROPERTIES = 1
	COMGLB_APPID                  GLOBALOPT_PROPERTIES = 2
	COMGLB_RPC_THREADPOOL_SETTING GLOBALOPT_PROPERTIES = 3
	COMGLB_RO_SETTINGS            GLOBALOPT_PROPERTIES = 4
	COMGLB_UNMARSHALING_POLICY    GLOBALOPT_PROPERTIES = 5
	COMGLB_PROPERTIES_RESERVED1   GLOBALOPT_PROPERTIES = 6
	COMGLB_PROPERTIES_RESERVED2   GLOBALOPT_PROPERTIES = 7
	COMGLB_PROPERTIES_RESERVED3   GLOBALOPT_PROPERTIES = 8
)

// enum
type GLOBALOPT_EH_VALUES int32

const (
	COMGLB_EXCEPTION_HANDLE             GLOBALOPT_EH_VALUES = 0
	COMGLB_EXCEPTION_DONOT_HANDLE_FATAL GLOBALOPT_EH_VALUES = 1
	COMGLB_EXCEPTION_DONOT_HANDLE       GLOBALOPT_EH_VALUES = 1
	COMGLB_EXCEPTION_DONOT_HANDLE_ANY   GLOBALOPT_EH_VALUES = 2
)

// enum
type GLOBALOPT_RPCTP_VALUES int32

const (
	COMGLB_RPC_THREADPOOL_SETTING_DEFAULT_POOL GLOBALOPT_RPCTP_VALUES = 0
	COMGLB_RPC_THREADPOOL_SETTING_PRIVATE_POOL GLOBALOPT_RPCTP_VALUES = 1
)

// enum
type GLOBALOPT_RO_FLAGS int32

const (
	COMGLB_STA_MODALLOOP_REMOVE_TOUCH_MESSAGES                    GLOBALOPT_RO_FLAGS = 1
	COMGLB_STA_MODALLOOP_SHARED_QUEUE_REMOVE_INPUT_MESSAGES       GLOBALOPT_RO_FLAGS = 2
	COMGLB_STA_MODALLOOP_SHARED_QUEUE_DONOT_REMOVE_INPUT_MESSAGES GLOBALOPT_RO_FLAGS = 4
	COMGLB_FAST_RUNDOWN                                           GLOBALOPT_RO_FLAGS = 8
	COMGLB_RESERVED1                                              GLOBALOPT_RO_FLAGS = 16
	COMGLB_RESERVED2                                              GLOBALOPT_RO_FLAGS = 32
	COMGLB_RESERVED3                                              GLOBALOPT_RO_FLAGS = 64
	COMGLB_STA_MODALLOOP_SHARED_QUEUE_REORDER_POINTER_MESSAGES    GLOBALOPT_RO_FLAGS = 128
	COMGLB_RESERVED4                                              GLOBALOPT_RO_FLAGS = 256
	COMGLB_RESERVED5                                              GLOBALOPT_RO_FLAGS = 512
	COMGLB_RESERVED6                                              GLOBALOPT_RO_FLAGS = 1024
)

// enum
type GLOBALOPT_UNMARSHALING_POLICY_VALUES int32

const (
	COMGLB_UNMARSHALING_POLICY_NORMAL GLOBALOPT_UNMARSHALING_POLICY_VALUES = 0
	COMGLB_UNMARSHALING_POLICY_STRONG GLOBALOPT_UNMARSHALING_POLICY_VALUES = 1
	COMGLB_UNMARSHALING_POLICY_HYBRID GLOBALOPT_UNMARSHALING_POLICY_VALUES = 2
)

// enum
type DCOM_CALL_STATE int32

const (
	DCOM_NONE          DCOM_CALL_STATE = 0
	DCOM_CALL_COMPLETE DCOM_CALL_STATE = 1
	DCOM_CALL_CANCELED DCOM_CALL_STATE = 2
)

// enum
type APTTYPEQUALIFIER int32

const (
	APTTYPEQUALIFIER_NONE               APTTYPEQUALIFIER = 0
	APTTYPEQUALIFIER_IMPLICIT_MTA       APTTYPEQUALIFIER = 1
	APTTYPEQUALIFIER_NA_ON_MTA          APTTYPEQUALIFIER = 2
	APTTYPEQUALIFIER_NA_ON_STA          APTTYPEQUALIFIER = 3
	APTTYPEQUALIFIER_NA_ON_IMPLICIT_MTA APTTYPEQUALIFIER = 4
	APTTYPEQUALIFIER_NA_ON_MAINSTA      APTTYPEQUALIFIER = 5
	APTTYPEQUALIFIER_APPLICATION_STA    APTTYPEQUALIFIER = 6
	APTTYPEQUALIFIER_RESERVED_1         APTTYPEQUALIFIER = 7
)

// enum
type APTTYPE int32

const (
	APTTYPE_CURRENT APTTYPE = -1
	APTTYPE_STA     APTTYPE = 0
	APTTYPE_MTA     APTTYPE = 1
	APTTYPE_NA      APTTYPE = 2
	APTTYPE_MAINSTA APTTYPE = 3
)

// enum
type THDTYPE int32

const (
	THDTYPE_BLOCKMESSAGES   THDTYPE = 0
	THDTYPE_PROCESSMESSAGES THDTYPE = 1
)

// enum
type CO_MARSHALING_CONTEXT_ATTRIBUTES int32

const (
	CO_MARSHALING_SOURCE_IS_APP_CONTAINER       CO_MARSHALING_CONTEXT_ATTRIBUTES = 0
	CO_MARSHALING_CONTEXT_ATTRIBUTE_RESERVED_1  CO_MARSHALING_CONTEXT_ATTRIBUTES = -2147483648
	CO_MARSHALING_CONTEXT_ATTRIBUTE_RESERVED_2  CO_MARSHALING_CONTEXT_ATTRIBUTES = -2147483647
	CO_MARSHALING_CONTEXT_ATTRIBUTE_RESERVED_3  CO_MARSHALING_CONTEXT_ATTRIBUTES = -2147483646
	CO_MARSHALING_CONTEXT_ATTRIBUTE_RESERVED_4  CO_MARSHALING_CONTEXT_ATTRIBUTES = -2147483645
	CO_MARSHALING_CONTEXT_ATTRIBUTE_RESERVED_5  CO_MARSHALING_CONTEXT_ATTRIBUTES = -2147483644
	CO_MARSHALING_CONTEXT_ATTRIBUTE_RESERVED_6  CO_MARSHALING_CONTEXT_ATTRIBUTES = -2147483643
	CO_MARSHALING_CONTEXT_ATTRIBUTE_RESERVED_7  CO_MARSHALING_CONTEXT_ATTRIBUTES = -2147483642
	CO_MARSHALING_CONTEXT_ATTRIBUTE_RESERVED_8  CO_MARSHALING_CONTEXT_ATTRIBUTES = -2147483641
	CO_MARSHALING_CONTEXT_ATTRIBUTE_RESERVED_9  CO_MARSHALING_CONTEXT_ATTRIBUTES = -2147483640
	CO_MARSHALING_CONTEXT_ATTRIBUTE_RESERVED_10 CO_MARSHALING_CONTEXT_ATTRIBUTES = -2147483639
	CO_MARSHALING_CONTEXT_ATTRIBUTE_RESERVED_11 CO_MARSHALING_CONTEXT_ATTRIBUTES = -2147483638
	CO_MARSHALING_CONTEXT_ATTRIBUTE_RESERVED_12 CO_MARSHALING_CONTEXT_ATTRIBUTES = -2147483637
	CO_MARSHALING_CONTEXT_ATTRIBUTE_RESERVED_13 CO_MARSHALING_CONTEXT_ATTRIBUTES = -2147483636
	CO_MARSHALING_CONTEXT_ATTRIBUTE_RESERVED_14 CO_MARSHALING_CONTEXT_ATTRIBUTES = -2147483635
	CO_MARSHALING_CONTEXT_ATTRIBUTE_RESERVED_15 CO_MARSHALING_CONTEXT_ATTRIBUTES = -2147483634
	CO_MARSHALING_CONTEXT_ATTRIBUTE_RESERVED_16 CO_MARSHALING_CONTEXT_ATTRIBUTES = -2147483633
	CO_MARSHALING_CONTEXT_ATTRIBUTE_RESERVED_17 CO_MARSHALING_CONTEXT_ATTRIBUTES = -2147483632
	CO_MARSHALING_CONTEXT_ATTRIBUTE_RESERVED_18 CO_MARSHALING_CONTEXT_ATTRIBUTES = -2147483631
)

// enum
type BIND_FLAGS int32

const (
	BIND_MAYBOTHERUSER     BIND_FLAGS = 1
	BIND_JUSTTESTEXISTENCE BIND_FLAGS = 2
)

// enum
type MKSYS int32

const (
	MKSYS_NONE             MKSYS = 0
	MKSYS_GENERICCOMPOSITE MKSYS = 1
	MKSYS_FILEMONIKER      MKSYS = 2
	MKSYS_ANTIMONIKER      MKSYS = 3
	MKSYS_ITEMMONIKER      MKSYS = 4
	MKSYS_POINTERMONIKER   MKSYS = 5
	MKSYS_CLASSMONIKER     MKSYS = 7
	MKSYS_OBJREFMONIKER    MKSYS = 8
	MKSYS_SESSIONMONIKER   MKSYS = 9
	MKSYS_LUAMONIKER       MKSYS = 10
)

// enum
type MKREDUCE int32

const (
	MKRREDUCE_ONE         MKREDUCE = 196608
	MKRREDUCE_TOUSER      MKREDUCE = 131072
	MKRREDUCE_THROUGHUSER MKREDUCE = 65536
	MKRREDUCE_ALL         MKREDUCE = 0
)

// enum
type ADVF int32

const (
	ADVF_NODATA            ADVF = 1
	ADVF_PRIMEFIRST        ADVF = 2
	ADVF_ONLYONCE          ADVF = 4
	ADVF_DATAONSTOP        ADVF = 64
	ADVFCACHE_NOHANDLER    ADVF = 8
	ADVFCACHE_FORCEBUILTIN ADVF = 16
	ADVFCACHE_ONSAVE       ADVF = 32
)

// enum
type TYMED int32

const (
	TYMED_HGLOBAL  TYMED = 1
	TYMED_FILE     TYMED = 2
	TYMED_ISTREAM  TYMED = 4
	TYMED_ISTORAGE TYMED = 8
	TYMED_GDI      TYMED = 16
	TYMED_MFPICT   TYMED = 32
	TYMED_ENHMF    TYMED = 64
	TYMED_NULL     TYMED = 0
)

// enum
type DATADIR int32

const (
	DATADIR_GET DATADIR = 1
	DATADIR_SET DATADIR = 2
)

// enum
type CALLTYPE int32

const (
	CALLTYPE_TOPLEVEL             CALLTYPE = 1
	CALLTYPE_NESTED               CALLTYPE = 2
	CALLTYPE_ASYNC                CALLTYPE = 3
	CALLTYPE_TOPLEVEL_CALLPENDING CALLTYPE = 4
	CALLTYPE_ASYNC_CALLPENDING    CALLTYPE = 5
)

// enum
type SERVERCALL int32

const (
	SERVERCALL_ISHANDLED  SERVERCALL = 0
	SERVERCALL_REJECTED   SERVERCALL = 1
	SERVERCALL_RETRYLATER SERVERCALL = 2
)

// enum
type PENDINGTYPE int32

const (
	PENDINGTYPE_TOPLEVEL PENDINGTYPE = 1
	PENDINGTYPE_NESTED   PENDINGTYPE = 2
)

// enum
type PENDINGMSG int32

const (
	PENDINGMSG_CANCELCALL     PENDINGMSG = 0
	PENDINGMSG_WAITNOPROCESS  PENDINGMSG = 1
	PENDINGMSG_WAITDEFPROCESS PENDINGMSG = 2
)

// enum
type ApplicationType int32

const (
	ServerApplication  ApplicationType = 0
	LibraryApplication ApplicationType = 1
)

// enum
type ShutdownType int32

const (
	IdleShutdown   ShutdownType = 0
	ForcedShutdown ShutdownType = 1
)

// enum
// flags
type COINIT uint32

const (
	COINIT_APARTMENTTHREADED COINIT = 2
	COINIT_MULTITHREADED     COINIT = 0
	COINIT_DISABLE_OLE1DDE   COINIT = 4
	COINIT_SPEED_OVER_MEMORY COINIT = 8
)

// enum
type COMSD int32

const (
	SD_LAUNCHPERMISSIONS  COMSD = 0
	SD_ACCESSPERMISSIONS  COMSD = 1
	SD_LAUNCHRESTRICTIONS COMSD = 2
	SD_ACCESSRESTRICTIONS COMSD = 3
)

// enum
type COWAIT_FLAGS int32

const (
	COWAIT_DEFAULT                  COWAIT_FLAGS = 0
	COWAIT_WAITALL                  COWAIT_FLAGS = 1
	COWAIT_ALERTABLE                COWAIT_FLAGS = 2
	COWAIT_INPUTAVAILABLE           COWAIT_FLAGS = 4
	COWAIT_DISPATCH_CALLS           COWAIT_FLAGS = 8
	COWAIT_DISPATCH_WINDOW_MESSAGES COWAIT_FLAGS = 16
)

// enum
type CWMO_FLAGS int32

const (
	CWMO_DEFAULT                  CWMO_FLAGS = 0
	CWMO_DISPATCH_CALLS           CWMO_FLAGS = 1
	CWMO_DISPATCH_WINDOW_MESSAGES CWMO_FLAGS = 2
)

// enum
type BINDINFOF int32

const (
	BINDINFOF_URLENCODESTGMEDDATA BINDINFOF = 1
	BINDINFOF_URLENCODEDEXTRAINFO BINDINFOF = 2
)

// enum
type Uri_PROPERTY int32

const (
	Uri_PROPERTY_ABSOLUTE_URI   Uri_PROPERTY = 0
	Uri_PROPERTY_STRING_START   Uri_PROPERTY = 0
	Uri_PROPERTY_AUTHORITY      Uri_PROPERTY = 1
	Uri_PROPERTY_DISPLAY_URI    Uri_PROPERTY = 2
	Uri_PROPERTY_DOMAIN         Uri_PROPERTY = 3
	Uri_PROPERTY_EXTENSION      Uri_PROPERTY = 4
	Uri_PROPERTY_FRAGMENT       Uri_PROPERTY = 5
	Uri_PROPERTY_HOST           Uri_PROPERTY = 6
	Uri_PROPERTY_PASSWORD       Uri_PROPERTY = 7
	Uri_PROPERTY_PATH           Uri_PROPERTY = 8
	Uri_PROPERTY_PATH_AND_QUERY Uri_PROPERTY = 9
	Uri_PROPERTY_QUERY          Uri_PROPERTY = 10
	Uri_PROPERTY_RAW_URI        Uri_PROPERTY = 11
	Uri_PROPERTY_SCHEME_NAME    Uri_PROPERTY = 12
	Uri_PROPERTY_USER_INFO      Uri_PROPERTY = 13
	Uri_PROPERTY_USER_NAME      Uri_PROPERTY = 14
	Uri_PROPERTY_STRING_LAST    Uri_PROPERTY = 14
	Uri_PROPERTY_HOST_TYPE      Uri_PROPERTY = 15
	Uri_PROPERTY_DWORD_START    Uri_PROPERTY = 15
	Uri_PROPERTY_PORT           Uri_PROPERTY = 16
	Uri_PROPERTY_SCHEME         Uri_PROPERTY = 17
	Uri_PROPERTY_ZONE           Uri_PROPERTY = 18
	Uri_PROPERTY_DWORD_LAST     Uri_PROPERTY = 18
)

// enum
type TYPEKIND int32

const (
	TKIND_ENUM      TYPEKIND = 0
	TKIND_RECORD    TYPEKIND = 1
	TKIND_MODULE    TYPEKIND = 2
	TKIND_INTERFACE TYPEKIND = 3
	TKIND_DISPATCH  TYPEKIND = 4
	TKIND_COCLASS   TYPEKIND = 5
	TKIND_ALIAS     TYPEKIND = 6
	TKIND_UNION     TYPEKIND = 7
	TKIND_MAX       TYPEKIND = 8
)

// enum
type CALLCONV int32

const (
	CC_FASTCALL   CALLCONV = 0
	CC_CDECL      CALLCONV = 1
	CC_MSCPASCAL  CALLCONV = 2
	CC_PASCAL     CALLCONV = 2
	CC_MACPASCAL  CALLCONV = 3
	CC_STDCALL    CALLCONV = 4
	CC_FPFASTCALL CALLCONV = 5
	CC_SYSCALL    CALLCONV = 6
	CC_MPWCDECL   CALLCONV = 7
	CC_MPWPASCAL  CALLCONV = 8
	CC_MAX        CALLCONV = 9
)

// enum
type FUNCKIND int32

const (
	FUNC_VIRTUAL     FUNCKIND = 0
	FUNC_PUREVIRTUAL FUNCKIND = 1
	FUNC_NONVIRTUAL  FUNCKIND = 2
	FUNC_STATIC      FUNCKIND = 3
	FUNC_DISPATCH    FUNCKIND = 4
)

// enum
type INVOKEKIND int32

const (
	INVOKE_FUNC           INVOKEKIND = 1
	INVOKE_PROPERTYGET    INVOKEKIND = 2
	INVOKE_PROPERTYPUT    INVOKEKIND = 4
	INVOKE_PROPERTYPUTREF INVOKEKIND = 8
)

// enum
type VARKIND int32

const (
	VAR_PERINSTANCE VARKIND = 0
	VAR_STATIC      VARKIND = 1
	VAR_CONST       VARKIND = 2
	VAR_DISPATCH    VARKIND = 3
)

// enum
type DESCKIND int32

const (
	DESCKIND_NONE           DESCKIND = 0
	DESCKIND_FUNCDESC       DESCKIND = 1
	DESCKIND_VARDESC        DESCKIND = 2
	DESCKIND_TYPECOMP       DESCKIND = 3
	DESCKIND_IMPLICITAPPOBJ DESCKIND = 4
	DESCKIND_MAX            DESCKIND = 5
)

// enum
type SYSKIND int32

const (
	SYS_WIN16 SYSKIND = 0
	SYS_WIN32 SYSKIND = 1
	SYS_MAC   SYSKIND = 2
	SYS_WIN64 SYSKIND = 3
)

// structs

type CY_Anonymous struct {
	Lo uint32
	Hi int32
}

type CY struct {
	CY_Anonymous
}

func (this *CY) Anonymous() *CY_Anonymous {
	return (*CY_Anonymous)(unsafe.Pointer(this))
}

func (this *CY) AnonymousVal() CY_Anonymous {
	return *(*CY_Anonymous)(unsafe.Pointer(this))
}

func (this *CY) Int64() *int64 {
	return (*int64)(unsafe.Pointer(this))
}

func (this *CY) Int64Val() int64 {
	return *(*int64)(unsafe.Pointer(this))
}

type CSPLATFORM struct {
	DwPlatformId    uint32
	DwVersionHi     uint32
	DwVersionLo     uint32
	DwProcessorArch uint32
}

type QUERYCONTEXT struct {
	DwContext   uint32
	Platform    CSPLATFORM
	Locale      uint32
	DwVersionHi uint32
	DwVersionLo uint32
}

type UCLSSPEC_Tagged_union_ByName struct {
	PPackageName PWSTR
	PolicyId     syscall.GUID
}

type UCLSSPEC_Tagged_union_ByObjectId struct {
	ObjectId syscall.GUID
	PolicyId syscall.GUID
}

type UCLSSPEC_Tagged_union struct {
	Data [4]uint64
}

func (this *UCLSSPEC_Tagged_union) Clsid() *syscall.GUID {
	return (*syscall.GUID)(unsafe.Pointer(this))
}

func (this *UCLSSPEC_Tagged_union) ClsidVal() syscall.GUID {
	return *(*syscall.GUID)(unsafe.Pointer(this))
}

func (this *UCLSSPEC_Tagged_union) PFileExt() *PWSTR {
	return (*PWSTR)(unsafe.Pointer(this))
}

func (this *UCLSSPEC_Tagged_union) PFileExtVal() PWSTR {
	return *(*PWSTR)(unsafe.Pointer(this))
}

func (this *UCLSSPEC_Tagged_union) PMimeType() *PWSTR {
	return (*PWSTR)(unsafe.Pointer(this))
}

func (this *UCLSSPEC_Tagged_union) PMimeTypeVal() PWSTR {
	return *(*PWSTR)(unsafe.Pointer(this))
}

func (this *UCLSSPEC_Tagged_union) PProgId() *PWSTR {
	return (*PWSTR)(unsafe.Pointer(this))
}

func (this *UCLSSPEC_Tagged_union) PProgIdVal() PWSTR {
	return *(*PWSTR)(unsafe.Pointer(this))
}

func (this *UCLSSPEC_Tagged_union) PFileName() *PWSTR {
	return (*PWSTR)(unsafe.Pointer(this))
}

func (this *UCLSSPEC_Tagged_union) PFileNameVal() PWSTR {
	return *(*PWSTR)(unsafe.Pointer(this))
}

func (this *UCLSSPEC_Tagged_union) ByName() *UCLSSPEC_Tagged_union_ByName {
	return (*UCLSSPEC_Tagged_union_ByName)(unsafe.Pointer(this))
}

func (this *UCLSSPEC_Tagged_union) ByNameVal() UCLSSPEC_Tagged_union_ByName {
	return *(*UCLSSPEC_Tagged_union_ByName)(unsafe.Pointer(this))
}

func (this *UCLSSPEC_Tagged_union) ByObjectId() *UCLSSPEC_Tagged_union_ByObjectId {
	return (*UCLSSPEC_Tagged_union_ByObjectId)(unsafe.Pointer(this))
}

func (this *UCLSSPEC_Tagged_union) ByObjectIdVal() UCLSSPEC_Tagged_union_ByObjectId {
	return *(*UCLSSPEC_Tagged_union_ByObjectId)(unsafe.Pointer(this))
}

type UCLSSPEC struct {
	Tyspec       uint32
	Tagged_union UCLSSPEC_Tagged_union
}

type COAUTHIDENTITY struct {
	User           *uint16
	UserLength     uint32
	Domain         *uint16
	DomainLength   uint32
	Password       *uint16
	PasswordLength uint32
	Flags          uint32
}

type COAUTHINFO struct {
	DwAuthnSvc           uint32
	DwAuthzSvc           uint32
	PwszServerPrincName  PWSTR
	DwAuthnLevel         uint32
	DwImpersonationLevel uint32
	PAuthIdentityData    *COAUTHIDENTITY
	DwCapabilities       uint32
}

type BYTE_BLOB struct {
	ClSize uint32
	AbData [1]byte
}

type WORD_BLOB struct {
	ClSize uint32
	AsData [1]uint16
}

type DWORD_BLOB struct {
	ClSize uint32
	AlData [1]uint32
}

type FLAGGED_BYTE_BLOB struct {
	FFlags uint32
	ClSize uint32
	AbData [1]byte
}

type FLAGGED_WORD_BLOB struct {
	FFlags uint32
	ClSize uint32
	AsData [1]uint16
}

type BYTE_SIZEDARR struct {
	ClSize uint32
	PData  *byte
}

type SHORT_SIZEDARR struct {
	ClSize uint32
	PData  *uint16
}

type LONG_SIZEDARR struct {
	ClSize uint32
	PData  *uint32
}

type HYPER_SIZEDARR struct {
	ClSize uint32
	PData  *int64
}

type BLOB struct {
	CbSize    uint32
	PBlobData *byte
}

type IEnumContextProps struct {
}

type IContext struct {
}

type COSERVERINFO struct {
	DwReserved1 uint32
	PwszName    PWSTR
	PAuthInfo   *COAUTHINFO
	DwReserved2 uint32
}

type MULTI_QI struct {
	PIID *syscall.GUID
	PItf *IUnknown
	Hr   HRESULT
}

type STATSTG struct {
	PwcsName          PWSTR
	Type_             uint32
	CbSize            uint64
	Mtime             FILETIME
	Ctime             FILETIME
	Atime             FILETIME
	GrfMode           uint32
	GrfLocksSupported uint32
	Clsid             syscall.GUID
	GrfStateBits      uint32
	Reserved          uint32
}

type RPCOLEMESSAGE struct {
	Reserved1          unsafe.Pointer
	DataRepresentation uint32
	Buffer             unsafe.Pointer
	CbBuffer           uint32
	IMethod            uint32
	Reserved2          [5]unsafe.Pointer
	RpcFlags           uint32
}

type SChannelHookCallInfo struct {
	Iid         syscall.GUID
	CbSize      uint32
	UCausality  syscall.GUID
	DwServerPid uint32
	IMethod     uint32
	PObject     unsafe.Pointer
}

type SOLE_AUTHENTICATION_SERVICE struct {
	DwAuthnSvc     uint32
	DwAuthzSvc     uint32
	PPrincipalName PWSTR
	Hr             HRESULT
}

type SOLE_AUTHENTICATION_INFO struct {
	DwAuthnSvc uint32
	DwAuthzSvc uint32
	PAuthInfo  unsafe.Pointer
}

type SOLE_AUTHENTICATION_LIST struct {
	CAuthInfo uint32
	AAuthInfo *SOLE_AUTHENTICATION_INFO
}

type MachineGlobalObjectTableRegistrationToken__ struct {
	Unused int32
}

type BIND_OPTS struct {
	CbStruct            uint32
	GrfFlags            uint32
	GrfMode             uint32
	DwTickCountDeadline uint32
}

type BIND_OPTS2 struct {
	BIND_OPTS
	DwTrackFlags   uint32
	DwClassContext uint32
	Locale         uint32
	PServerInfo    *COSERVERINFO
}

type BIND_OPTS3 struct {
	BIND_OPTS2
	Hwnd HWND
}

type DVTARGETDEVICE struct {
	TdSize             uint32
	TdDriverNameOffset uint16
	TdDeviceNameOffset uint16
	TdPortNameOffset   uint16
	TdExtDevmodeOffset uint16
	TdData             [1]byte
}

type FORMATETC struct {
	CfFormat uint16
	Ptd      *DVTARGETDEVICE
	DwAspect uint32
	Lindex   int32
	Tymed    uint32
}

type STATDATA struct {
	Formatetc    FORMATETC
	Advf         uint32
	PAdvSink     *IAdviseSink
	DwConnection uint32
}

type RemSTGMEDIUM struct {
	Tymed          uint32
	DwHandleType   uint32
	PData          uint32
	PUnkForRelease uint32
	CbData         uint32
	Data           [1]byte
}

type STGMEDIUM_Anonymous struct {
	Data [1]uint64
}

func (this *STGMEDIUM_Anonymous) HBitmap() *HBITMAP {
	return (*HBITMAP)(unsafe.Pointer(this))
}

func (this *STGMEDIUM_Anonymous) HBitmapVal() HBITMAP {
	return *(*HBITMAP)(unsafe.Pointer(this))
}

func (this *STGMEDIUM_Anonymous) HMetaFilePict() *unsafe.Pointer {
	return (*unsafe.Pointer)(unsafe.Pointer(this))
}

func (this *STGMEDIUM_Anonymous) HMetaFilePictVal() unsafe.Pointer {
	return *(*unsafe.Pointer)(unsafe.Pointer(this))
}

func (this *STGMEDIUM_Anonymous) HEnhMetaFile() *HENHMETAFILE {
	return (*HENHMETAFILE)(unsafe.Pointer(this))
}

func (this *STGMEDIUM_Anonymous) HEnhMetaFileVal() HENHMETAFILE {
	return *(*HENHMETAFILE)(unsafe.Pointer(this))
}

func (this *STGMEDIUM_Anonymous) HGlobal() *uintptr {
	return (*uintptr)(unsafe.Pointer(this))
}

func (this *STGMEDIUM_Anonymous) HGlobalVal() uintptr {
	return *(*uintptr)(unsafe.Pointer(this))
}

func (this *STGMEDIUM_Anonymous) LpszFileName() *PWSTR {
	return (*PWSTR)(unsafe.Pointer(this))
}

func (this *STGMEDIUM_Anonymous) LpszFileNameVal() PWSTR {
	return *(*PWSTR)(unsafe.Pointer(this))
}

func (this *STGMEDIUM_Anonymous) Pstm() **IStream {
	return (**IStream)(unsafe.Pointer(this))
}

func (this *STGMEDIUM_Anonymous) PstmVal() *IStream {
	return *(**IStream)(unsafe.Pointer(this))
}

func (this *STGMEDIUM_Anonymous) Pstg() **IStorage {
	return (**IStorage)(unsafe.Pointer(this))
}

func (this *STGMEDIUM_Anonymous) PstgVal() *IStorage {
	return *(**IStorage)(unsafe.Pointer(this))
}

type STGMEDIUM struct {
	Tymed uint32
	STGMEDIUM_Anonymous
	PUnkForRelease *IUnknown
}

type GDI_OBJECT_U struct {
	Data [1]uint64
}

func (this *GDI_OBJECT_U) HBitmap() **UserHBITMAP {
	return (**UserHBITMAP)(unsafe.Pointer(this))
}

func (this *GDI_OBJECT_U) HBitmapVal() *UserHBITMAP {
	return *(**UserHBITMAP)(unsafe.Pointer(this))
}

func (this *GDI_OBJECT_U) HPalette() **UserHPALETTE {
	return (**UserHPALETTE)(unsafe.Pointer(this))
}

func (this *GDI_OBJECT_U) HPaletteVal() *UserHPALETTE {
	return *(**UserHPALETTE)(unsafe.Pointer(this))
}

func (this *GDI_OBJECT_U) HGeneric() **UserHGLOBAL {
	return (**UserHGLOBAL)(unsafe.Pointer(this))
}

func (this *GDI_OBJECT_U) HGenericVal() *UserHGLOBAL {
	return *(**UserHGLOBAL)(unsafe.Pointer(this))
}

type GDI_OBJECT struct {
	ObjectType uint32
	U          GDI_OBJECT_U
}

type UserSTGMEDIUM_STGMEDIUM_UNION_U struct {
	Data [1]uint64
}

func (this *UserSTGMEDIUM_STGMEDIUM_UNION_U) HMetaFilePict() **UserHMETAFILEPICT {
	return (**UserHMETAFILEPICT)(unsafe.Pointer(this))
}

func (this *UserSTGMEDIUM_STGMEDIUM_UNION_U) HMetaFilePictVal() *UserHMETAFILEPICT {
	return *(**UserHMETAFILEPICT)(unsafe.Pointer(this))
}

func (this *UserSTGMEDIUM_STGMEDIUM_UNION_U) HHEnhMetaFile() **UserHENHMETAFILE {
	return (**UserHENHMETAFILE)(unsafe.Pointer(this))
}

func (this *UserSTGMEDIUM_STGMEDIUM_UNION_U) HHEnhMetaFileVal() *UserHENHMETAFILE {
	return *(**UserHENHMETAFILE)(unsafe.Pointer(this))
}

func (this *UserSTGMEDIUM_STGMEDIUM_UNION_U) HGdiHandle() **GDI_OBJECT {
	return (**GDI_OBJECT)(unsafe.Pointer(this))
}

func (this *UserSTGMEDIUM_STGMEDIUM_UNION_U) HGdiHandleVal() *GDI_OBJECT {
	return *(**GDI_OBJECT)(unsafe.Pointer(this))
}

func (this *UserSTGMEDIUM_STGMEDIUM_UNION_U) HGlobal() **UserHGLOBAL {
	return (**UserHGLOBAL)(unsafe.Pointer(this))
}

func (this *UserSTGMEDIUM_STGMEDIUM_UNION_U) HGlobalVal() *UserHGLOBAL {
	return *(**UserHGLOBAL)(unsafe.Pointer(this))
}

func (this *UserSTGMEDIUM_STGMEDIUM_UNION_U) LpszFileName() *PWSTR {
	return (*PWSTR)(unsafe.Pointer(this))
}

func (this *UserSTGMEDIUM_STGMEDIUM_UNION_U) LpszFileNameVal() PWSTR {
	return *(*PWSTR)(unsafe.Pointer(this))
}

func (this *UserSTGMEDIUM_STGMEDIUM_UNION_U) Pstm() **BYTE_BLOB {
	return (**BYTE_BLOB)(unsafe.Pointer(this))
}

func (this *UserSTGMEDIUM_STGMEDIUM_UNION_U) PstmVal() *BYTE_BLOB {
	return *(**BYTE_BLOB)(unsafe.Pointer(this))
}

func (this *UserSTGMEDIUM_STGMEDIUM_UNION_U) Pstg() **BYTE_BLOB {
	return (**BYTE_BLOB)(unsafe.Pointer(this))
}

func (this *UserSTGMEDIUM_STGMEDIUM_UNION_U) PstgVal() *BYTE_BLOB {
	return *(**BYTE_BLOB)(unsafe.Pointer(this))
}

type UserSTGMEDIUM_STGMEDIUM_UNION struct {
	Tymed uint32
	U     UserSTGMEDIUM_STGMEDIUM_UNION_U
}

type UserSTGMEDIUM struct {
	PUnkForRelease *IUnknown
}

type UserFLAG_STGMEDIUM struct {
	ContextFlags   int32
	FPassOwnership int32
	Stgmed         UserSTGMEDIUM
}

type FLAG_STGMEDIUM struct {
	ContextFlags   int32
	FPassOwnership int32
	Stgmed         STGMEDIUM
}

type INTERFACEINFO struct {
	PUnk    *IUnknown
	Iid     syscall.GUID
	WMethod uint16
}

type StorageLayout struct {
	LayoutType      uint32
	PwcsElementName PWSTR
	COffset         int64
	CBytes          int64
}

type CATEGORYINFO struct {
	Catid         syscall.GUID
	Lcid          uint32
	SzDescription [128]uint16
}

type ComCallData struct {
	DwDispid     uint32
	DwReserved   uint32
	PUserDefined unsafe.Pointer
}

type BINDINFO struct {
	CbSize             uint32
	SzExtraInfo        PWSTR
	StgmedData         STGMEDIUM
	GrfBindInfoF       uint32
	DwBindVerb         uint32
	SzCustomVerb       PWSTR
	CbstgmedData       uint32
	DwOptions          uint32
	DwOptionsFlags     uint32
	DwCodePage         uint32
	SecurityAttributes SECURITY_ATTRIBUTES
	Iid                syscall.GUID
	PUnk               *IUnknown
	DwReserved         uint32
}

type AUTHENTICATEINFO struct {
	DwFlags    uint32
	DwReserved uint32
}

type SAFEARRAYBOUND struct {
	CElements uint32
	LLbound   int32
}

type SAFEARRAY struct {
	CDims      uint16
	FFeatures  uint16
	CbElements uint32
	CLocks     uint32
	PvData     unsafe.Pointer
	Rgsabound  [1]SAFEARRAYBOUND
}

type VARIANT_Anonymous_Anonymous_Anonymous_Anonymous struct {
	PvRecord unsafe.Pointer
	PRecInfo *IRecordInfo
}

type VARIANT_Anonymous_Anonymous_Anonymous struct {
	VARIANT_Anonymous_Anonymous_Anonymous_Anonymous
}

func (this *VARIANT_Anonymous_Anonymous_Anonymous) LlVal() *int64 {
	return (*int64)(unsafe.Pointer(this))
}

func (this *VARIANT_Anonymous_Anonymous_Anonymous) LlValVal() int64 {
	return *(*int64)(unsafe.Pointer(this))
}

func (this *VARIANT_Anonymous_Anonymous_Anonymous) LVal() *int32 {
	return (*int32)(unsafe.Pointer(this))
}

func (this *VARIANT_Anonymous_Anonymous_Anonymous) LValVal() int32 {
	return *(*int32)(unsafe.Pointer(this))
}

func (this *VARIANT_Anonymous_Anonymous_Anonymous) BVal() *byte {
	return (*byte)(unsafe.Pointer(this))
}

func (this *VARIANT_Anonymous_Anonymous_Anonymous) BValVal() byte {
	return *(*byte)(unsafe.Pointer(this))
}

func (this *VARIANT_Anonymous_Anonymous_Anonymous) IVal() *int16 {
	return (*int16)(unsafe.Pointer(this))
}

func (this *VARIANT_Anonymous_Anonymous_Anonymous) IValVal() int16 {
	return *(*int16)(unsafe.Pointer(this))
}

func (this *VARIANT_Anonymous_Anonymous_Anonymous) FltVal() *float32 {
	return (*float32)(unsafe.Pointer(this))
}

func (this *VARIANT_Anonymous_Anonymous_Anonymous) FltValVal() float32 {
	return *(*float32)(unsafe.Pointer(this))
}

func (this *VARIANT_Anonymous_Anonymous_Anonymous) DblVal() *float64 {
	return (*float64)(unsafe.Pointer(this))
}

func (this *VARIANT_Anonymous_Anonymous_Anonymous) DblValVal() float64 {
	return *(*float64)(unsafe.Pointer(this))
}

func (this *VARIANT_Anonymous_Anonymous_Anonymous) BoolVal() *int16 {
	return (*int16)(unsafe.Pointer(this))
}

func (this *VARIANT_Anonymous_Anonymous_Anonymous) BoolValVal() int16 {
	return *(*int16)(unsafe.Pointer(this))
}

func (this *VARIANT_Anonymous_Anonymous_Anonymous) OBSOLETE__VARIANT_BOOL__() *int16 {
	return (*int16)(unsafe.Pointer(this))
}

func (this *VARIANT_Anonymous_Anonymous_Anonymous) OBSOLETE__VARIANT_BOOL__Val() int16 {
	return *(*int16)(unsafe.Pointer(this))
}

func (this *VARIANT_Anonymous_Anonymous_Anonymous) Scode() *int32 {
	return (*int32)(unsafe.Pointer(this))
}

func (this *VARIANT_Anonymous_Anonymous_Anonymous) ScodeVal() int32 {
	return *(*int32)(unsafe.Pointer(this))
}

func (this *VARIANT_Anonymous_Anonymous_Anonymous) CyVal() *CY {
	return (*CY)(unsafe.Pointer(this))
}

func (this *VARIANT_Anonymous_Anonymous_Anonymous) CyValVal() CY {
	return *(*CY)(unsafe.Pointer(this))
}

func (this *VARIANT_Anonymous_Anonymous_Anonymous) Date() *float64 {
	return (*float64)(unsafe.Pointer(this))
}

func (this *VARIANT_Anonymous_Anonymous_Anonymous) DateVal() float64 {
	return *(*float64)(unsafe.Pointer(this))
}

func (this *VARIANT_Anonymous_Anonymous_Anonymous) BstrVal() *BSTR {
	return (*BSTR)(unsafe.Pointer(this))
}

func (this *VARIANT_Anonymous_Anonymous_Anonymous) BstrValVal() BSTR {
	return *(*BSTR)(unsafe.Pointer(this))
}

func (this *VARIANT_Anonymous_Anonymous_Anonymous) PunkVal() **IUnknown {
	return (**IUnknown)(unsafe.Pointer(this))
}

func (this *VARIANT_Anonymous_Anonymous_Anonymous) PunkValVal() *IUnknown {
	return *(**IUnknown)(unsafe.Pointer(this))
}

func (this *VARIANT_Anonymous_Anonymous_Anonymous) PdispVal() **IDispatch {
	return (**IDispatch)(unsafe.Pointer(this))
}

func (this *VARIANT_Anonymous_Anonymous_Anonymous) PdispValVal() *IDispatch {
	return *(**IDispatch)(unsafe.Pointer(this))
}

func (this *VARIANT_Anonymous_Anonymous_Anonymous) Parray() **SAFEARRAY {
	return (**SAFEARRAY)(unsafe.Pointer(this))
}

func (this *VARIANT_Anonymous_Anonymous_Anonymous) ParrayVal() *SAFEARRAY {
	return *(**SAFEARRAY)(unsafe.Pointer(this))
}

func (this *VARIANT_Anonymous_Anonymous_Anonymous) PbVal() **byte {
	return (**byte)(unsafe.Pointer(this))
}

func (this *VARIANT_Anonymous_Anonymous_Anonymous) PbValVal() *byte {
	return *(**byte)(unsafe.Pointer(this))
}

func (this *VARIANT_Anonymous_Anonymous_Anonymous) PiVal() **int16 {
	return (**int16)(unsafe.Pointer(this))
}

func (this *VARIANT_Anonymous_Anonymous_Anonymous) PiValVal() *int16 {
	return *(**int16)(unsafe.Pointer(this))
}

func (this *VARIANT_Anonymous_Anonymous_Anonymous) PlVal() **int32 {
	return (**int32)(unsafe.Pointer(this))
}

func (this *VARIANT_Anonymous_Anonymous_Anonymous) PlValVal() *int32 {
	return *(**int32)(unsafe.Pointer(this))
}

func (this *VARIANT_Anonymous_Anonymous_Anonymous) PllVal() **int64 {
	return (**int64)(unsafe.Pointer(this))
}

func (this *VARIANT_Anonymous_Anonymous_Anonymous) PllValVal() *int64 {
	return *(**int64)(unsafe.Pointer(this))
}

func (this *VARIANT_Anonymous_Anonymous_Anonymous) PfltVal() **float32 {
	return (**float32)(unsafe.Pointer(this))
}

func (this *VARIANT_Anonymous_Anonymous_Anonymous) PfltValVal() *float32 {
	return *(**float32)(unsafe.Pointer(this))
}

func (this *VARIANT_Anonymous_Anonymous_Anonymous) PdblVal() **float64 {
	return (**float64)(unsafe.Pointer(this))
}

func (this *VARIANT_Anonymous_Anonymous_Anonymous) PdblValVal() *float64 {
	return *(**float64)(unsafe.Pointer(this))
}

func (this *VARIANT_Anonymous_Anonymous_Anonymous) PboolVal() **int16 {
	return (**int16)(unsafe.Pointer(this))
}

func (this *VARIANT_Anonymous_Anonymous_Anonymous) PboolValVal() *int16 {
	return *(**int16)(unsafe.Pointer(this))
}

func (this *VARIANT_Anonymous_Anonymous_Anonymous) OBSOLETE__VARIANT_PBOOL__() **int16 {
	return (**int16)(unsafe.Pointer(this))
}

func (this *VARIANT_Anonymous_Anonymous_Anonymous) OBSOLETE__VARIANT_PBOOL__Val() *int16 {
	return *(**int16)(unsafe.Pointer(this))
}

func (this *VARIANT_Anonymous_Anonymous_Anonymous) Pscode() **int32 {
	return (**int32)(unsafe.Pointer(this))
}

func (this *VARIANT_Anonymous_Anonymous_Anonymous) PscodeVal() *int32 {
	return *(**int32)(unsafe.Pointer(this))
}

func (this *VARIANT_Anonymous_Anonymous_Anonymous) PcyVal() **CY {
	return (**CY)(unsafe.Pointer(this))
}

func (this *VARIANT_Anonymous_Anonymous_Anonymous) PcyValVal() *CY {
	return *(**CY)(unsafe.Pointer(this))
}

func (this *VARIANT_Anonymous_Anonymous_Anonymous) Pdate() **float64 {
	return (**float64)(unsafe.Pointer(this))
}

func (this *VARIANT_Anonymous_Anonymous_Anonymous) PdateVal() *float64 {
	return *(**float64)(unsafe.Pointer(this))
}

func (this *VARIANT_Anonymous_Anonymous_Anonymous) PbstrVal() **BSTR {
	return (**BSTR)(unsafe.Pointer(this))
}

func (this *VARIANT_Anonymous_Anonymous_Anonymous) PbstrValVal() *BSTR {
	return *(**BSTR)(unsafe.Pointer(this))
}

func (this *VARIANT_Anonymous_Anonymous_Anonymous) PpunkVal() ***IUnknown {
	return (***IUnknown)(unsafe.Pointer(this))
}

func (this *VARIANT_Anonymous_Anonymous_Anonymous) PpunkValVal() **IUnknown {
	return *(***IUnknown)(unsafe.Pointer(this))
}

func (this *VARIANT_Anonymous_Anonymous_Anonymous) PpdispVal() ***IDispatch {
	return (***IDispatch)(unsafe.Pointer(this))
}

func (this *VARIANT_Anonymous_Anonymous_Anonymous) PpdispValVal() **IDispatch {
	return *(***IDispatch)(unsafe.Pointer(this))
}

func (this *VARIANT_Anonymous_Anonymous_Anonymous) Pparray() ***SAFEARRAY {
	return (***SAFEARRAY)(unsafe.Pointer(this))
}

func (this *VARIANT_Anonymous_Anonymous_Anonymous) PparrayVal() **SAFEARRAY {
	return *(***SAFEARRAY)(unsafe.Pointer(this))
}

func (this *VARIANT_Anonymous_Anonymous_Anonymous) PvarVal() **VARIANT {
	return (**VARIANT)(unsafe.Pointer(this))
}

func (this *VARIANT_Anonymous_Anonymous_Anonymous) PvarValVal() *VARIANT {
	return *(**VARIANT)(unsafe.Pointer(this))
}

func (this *VARIANT_Anonymous_Anonymous_Anonymous) Byref() *unsafe.Pointer {
	return (*unsafe.Pointer)(unsafe.Pointer(this))
}

func (this *VARIANT_Anonymous_Anonymous_Anonymous) ByrefVal() unsafe.Pointer {
	return *(*unsafe.Pointer)(unsafe.Pointer(this))
}

func (this *VARIANT_Anonymous_Anonymous_Anonymous) CVal() *CHAR {
	return (*CHAR)(unsafe.Pointer(this))
}

func (this *VARIANT_Anonymous_Anonymous_Anonymous) CValVal() CHAR {
	return *(*CHAR)(unsafe.Pointer(this))
}

func (this *VARIANT_Anonymous_Anonymous_Anonymous) UiVal() *uint16 {
	return (*uint16)(unsafe.Pointer(this))
}

func (this *VARIANT_Anonymous_Anonymous_Anonymous) UiValVal() uint16 {
	return *(*uint16)(unsafe.Pointer(this))
}

func (this *VARIANT_Anonymous_Anonymous_Anonymous) UlVal() *uint32 {
	return (*uint32)(unsafe.Pointer(this))
}

func (this *VARIANT_Anonymous_Anonymous_Anonymous) UlValVal() uint32 {
	return *(*uint32)(unsafe.Pointer(this))
}

func (this *VARIANT_Anonymous_Anonymous_Anonymous) UllVal() *uint64 {
	return (*uint64)(unsafe.Pointer(this))
}

func (this *VARIANT_Anonymous_Anonymous_Anonymous) UllValVal() uint64 {
	return *(*uint64)(unsafe.Pointer(this))
}

func (this *VARIANT_Anonymous_Anonymous_Anonymous) IntVal() *int32 {
	return (*int32)(unsafe.Pointer(this))
}

func (this *VARIANT_Anonymous_Anonymous_Anonymous) IntValVal() int32 {
	return *(*int32)(unsafe.Pointer(this))
}

func (this *VARIANT_Anonymous_Anonymous_Anonymous) UintVal() *uint32 {
	return (*uint32)(unsafe.Pointer(this))
}

func (this *VARIANT_Anonymous_Anonymous_Anonymous) UintValVal() uint32 {
	return *(*uint32)(unsafe.Pointer(this))
}

func (this *VARIANT_Anonymous_Anonymous_Anonymous) PdecVal() **DECIMAL {
	return (**DECIMAL)(unsafe.Pointer(this))
}

func (this *VARIANT_Anonymous_Anonymous_Anonymous) PdecValVal() *DECIMAL {
	return *(**DECIMAL)(unsafe.Pointer(this))
}

func (this *VARIANT_Anonymous_Anonymous_Anonymous) PcVal() *PSTR {
	return (*PSTR)(unsafe.Pointer(this))
}

func (this *VARIANT_Anonymous_Anonymous_Anonymous) PcValVal() PSTR {
	return *(*PSTR)(unsafe.Pointer(this))
}

func (this *VARIANT_Anonymous_Anonymous_Anonymous) PuiVal() **uint16 {
	return (**uint16)(unsafe.Pointer(this))
}

func (this *VARIANT_Anonymous_Anonymous_Anonymous) PuiValVal() *uint16 {
	return *(**uint16)(unsafe.Pointer(this))
}

func (this *VARIANT_Anonymous_Anonymous_Anonymous) PulVal() **uint32 {
	return (**uint32)(unsafe.Pointer(this))
}

func (this *VARIANT_Anonymous_Anonymous_Anonymous) PulValVal() *uint32 {
	return *(**uint32)(unsafe.Pointer(this))
}

func (this *VARIANT_Anonymous_Anonymous_Anonymous) PullVal() **uint64 {
	return (**uint64)(unsafe.Pointer(this))
}

func (this *VARIANT_Anonymous_Anonymous_Anonymous) PullValVal() *uint64 {
	return *(**uint64)(unsafe.Pointer(this))
}

func (this *VARIANT_Anonymous_Anonymous_Anonymous) PintVal() **int32 {
	return (**int32)(unsafe.Pointer(this))
}

func (this *VARIANT_Anonymous_Anonymous_Anonymous) PintValVal() *int32 {
	return *(**int32)(unsafe.Pointer(this))
}

func (this *VARIANT_Anonymous_Anonymous_Anonymous) PuintVal() **uint32 {
	return (**uint32)(unsafe.Pointer(this))
}

func (this *VARIANT_Anonymous_Anonymous_Anonymous) PuintValVal() *uint32 {
	return *(**uint32)(unsafe.Pointer(this))
}

func (this *VARIANT_Anonymous_Anonymous_Anonymous) Anonymous() *VARIANT_Anonymous_Anonymous_Anonymous_Anonymous {
	return (*VARIANT_Anonymous_Anonymous_Anonymous_Anonymous)(unsafe.Pointer(this))
}

func (this *VARIANT_Anonymous_Anonymous_Anonymous) AnonymousVal() VARIANT_Anonymous_Anonymous_Anonymous_Anonymous {
	return *(*VARIANT_Anonymous_Anonymous_Anonymous_Anonymous)(unsafe.Pointer(this))
}

type VARIANT_Anonymous_Anonymous struct {
	Vt         uint16
	WReserved1 uint16
	WReserved2 uint16
	WReserved3 uint16
	VARIANT_Anonymous_Anonymous_Anonymous
}

type VARIANT_Anonymous struct {
	VARIANT_Anonymous_Anonymous
}

func (this *VARIANT_Anonymous) Anonymous() *VARIANT_Anonymous_Anonymous {
	return (*VARIANT_Anonymous_Anonymous)(unsafe.Pointer(this))
}

func (this *VARIANT_Anonymous) AnonymousVal() VARIANT_Anonymous_Anonymous {
	return *(*VARIANT_Anonymous_Anonymous)(unsafe.Pointer(this))
}

func (this *VARIANT_Anonymous) DecVal() *DECIMAL {
	return (*DECIMAL)(unsafe.Pointer(this))
}

func (this *VARIANT_Anonymous) DecValVal() DECIMAL {
	return *(*DECIMAL)(unsafe.Pointer(this))
}

type VARIANT struct {
	VARIANT_Anonymous
}

type TYPEDESC_Anonymous struct {
	Data [1]uint64
}

func (this *TYPEDESC_Anonymous) Lptdesc() **TYPEDESC {
	return (**TYPEDESC)(unsafe.Pointer(this))
}

func (this *TYPEDESC_Anonymous) LptdescVal() *TYPEDESC {
	return *(**TYPEDESC)(unsafe.Pointer(this))
}

func (this *TYPEDESC_Anonymous) Lpadesc() **ARRAYDESC {
	return (**ARRAYDESC)(unsafe.Pointer(this))
}

func (this *TYPEDESC_Anonymous) LpadescVal() *ARRAYDESC {
	return *(**ARRAYDESC)(unsafe.Pointer(this))
}

func (this *TYPEDESC_Anonymous) Hreftype() *uint32 {
	return (*uint32)(unsafe.Pointer(this))
}

func (this *TYPEDESC_Anonymous) HreftypeVal() uint32 {
	return *(*uint32)(unsafe.Pointer(this))
}

type TYPEDESC struct {
	TYPEDESC_Anonymous
	Vt uint16
}

type IDLDESC struct {
	DwReserved uintptr
	WIDLFlags  uint16
}

type ELEMDESC_Anonymous struct {
	Data [2]uint64
}

func (this *ELEMDESC_Anonymous) Idldesc() *IDLDESC {
	return (*IDLDESC)(unsafe.Pointer(this))
}

func (this *ELEMDESC_Anonymous) IdldescVal() IDLDESC {
	return *(*IDLDESC)(unsafe.Pointer(this))
}

func (this *ELEMDESC_Anonymous) Paramdesc() *PARAMDESC {
	return (*PARAMDESC)(unsafe.Pointer(this))
}

func (this *ELEMDESC_Anonymous) ParamdescVal() PARAMDESC {
	return *(*PARAMDESC)(unsafe.Pointer(this))
}

type ELEMDESC struct {
	Tdesc TYPEDESC
	ELEMDESC_Anonymous
}

type TYPEATTR struct {
	Guid             syscall.GUID
	Lcid             uint32
	DwReserved       uint32
	MemidConstructor int32
	MemidDestructor  int32
	LpstrSchema      PWSTR
	CbSizeInstance   uint32
	Typekind         TYPEKIND
	CFuncs           uint16
	CVars            uint16
	CImplTypes       uint16
	CbSizeVft        uint16
	CbAlignment      uint16
	WTypeFlags       uint16
	WMajorVerNum     uint16
	WMinorVerNum     uint16
	TdescAlias       TYPEDESC
	IdldescType      IDLDESC
}

type DISPPARAMS struct {
	Rgvarg            *VARIANT
	RgdispidNamedArgs *int32
	CArgs             uint32
	CNamedArgs        uint32
}

type EXCEPINFO struct {
	WCode             uint16
	WReserved         uint16
	BstrSource        BSTR
	BstrDescription   BSTR
	BstrHelpFile      BSTR
	DwHelpContext     uint32
	PvReserved        unsafe.Pointer
	PfnDeferredFillIn LPEXCEPFINO_DEFERRED_FILLIN
	Scode             int32
}

type FUNCDESC struct {
	Memid             int32
	Lprgscode         *int32
	LprgelemdescParam *ELEMDESC
	Funckind          FUNCKIND
	Invkind           INVOKEKIND
	Callconv          CALLCONV
	CParams           int16
	CParamsOpt        int16
	OVft              int16
	CScodes           int16
	ElemdescFunc      ELEMDESC
	WFuncFlags        uint16
}

type VARDESC_Anonymous struct {
	Data [1]uint64
}

func (this *VARDESC_Anonymous) OInst() *uint32 {
	return (*uint32)(unsafe.Pointer(this))
}

func (this *VARDESC_Anonymous) OInstVal() uint32 {
	return *(*uint32)(unsafe.Pointer(this))
}

func (this *VARDESC_Anonymous) LpvarValue() **VARIANT {
	return (**VARIANT)(unsafe.Pointer(this))
}

func (this *VARDESC_Anonymous) LpvarValueVal() *VARIANT {
	return *(**VARIANT)(unsafe.Pointer(this))
}

type VARDESC struct {
	Memid       int32
	LpstrSchema PWSTR
	VARDESC_Anonymous
	ElemdescVar ELEMDESC
	WVarFlags   uint16
	Varkind     VARKIND
}

type CUSTDATAITEM struct {
	Guid     syscall.GUID
	VarValue VARIANT
}

type CUSTDATA struct {
	CCustData   uint32
	PrgCustData *CUSTDATAITEM
}

type BINDPTR struct {
	Data [1]uint64
}

func (this *BINDPTR) Lpfuncdesc() **FUNCDESC {
	return (**FUNCDESC)(unsafe.Pointer(this))
}

func (this *BINDPTR) LpfuncdescVal() *FUNCDESC {
	return *(**FUNCDESC)(unsafe.Pointer(this))
}

func (this *BINDPTR) Lpvardesc() **VARDESC {
	return (**VARDESC)(unsafe.Pointer(this))
}

func (this *BINDPTR) LpvardescVal() *VARDESC {
	return *(**VARDESC)(unsafe.Pointer(this))
}

func (this *BINDPTR) Lptcomp() **ITypeComp {
	return (**ITypeComp)(unsafe.Pointer(this))
}

func (this *BINDPTR) LptcompVal() *ITypeComp {
	return *(**ITypeComp)(unsafe.Pointer(this))
}

type TLIBATTR struct {
	Guid         syscall.GUID
	Lcid         uint32
	Syskind      SYSKIND
	WMajorVerNum uint16
	WMinorVerNum uint16
	WLibFlags    uint16
}

type CONNECTDATA struct {
	PUnk     *IUnknown
	DwCookie uint32
}

// func types

type LPEXCEPFINO_DEFERRED_FILLIN = uintptr
type LPEXCEPFINO_DEFERRED_FILLIN_func = func(pExcepInfo *EXCEPINFO) HRESULT

type LPFNGETCLASSOBJECT = uintptr
type LPFNGETCLASSOBJECT_func = func(param0 *syscall.GUID, param1 *syscall.GUID, param2 unsafe.Pointer) HRESULT

type LPFNCANUNLOADNOW = uintptr
type LPFNCANUNLOADNOW_func = func() HRESULT

type PFNCONTEXTCALL = uintptr
type PFNCONTEXTCALL_func = func(pParam *ComCallData) HRESULT

// interfaces

// 00000000-0000-0000-C000-000000000046
var IID_IUnknown = syscall.GUID{0x00000000, 0x0000, 0x0000,
	[8]byte{0xC0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x46}}

type IUnknownInterface interface {
	QueryInterface(riid *syscall.GUID, ppvObject unsafe.Pointer) HRESULT
	AddRef() uint32
	Release() uint32
}

type IUnknownVtbl struct {
	QueryInterface uintptr
	AddRef         uintptr
	Release        uintptr
}

type IUnknown struct {
	LpVtbl *[1024]uintptr
}

func (this *IUnknown) Vtbl() *IUnknownVtbl {
	return (*IUnknownVtbl)(unsafe.Pointer(this.LpVtbl))
}

func (this *IUnknown) QueryInterface(riid *syscall.GUID, ppvObject unsafe.Pointer) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().QueryInterface, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(riid)), uintptr(ppvObject))
	return HRESULT(ret)
}

func (this *IUnknown) AddRef() uint32 {
	ret, _, _ := syscall.SyscallN(this.Vtbl().AddRef, uintptr(unsafe.Pointer(this)))
	return uint32(ret)
}

func (this *IUnknown) Release() uint32 {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Release, uintptr(unsafe.Pointer(this)))
	return uint32(ret)
}

// 000E0000-0000-0000-C000-000000000046
var IID_AsyncIUnknown = syscall.GUID{0x000E0000, 0x0000, 0x0000,
	[8]byte{0xC0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x46}}

type AsyncIUnknownInterface interface {
	IUnknownInterface
	Begin_QueryInterface(riid *syscall.GUID) HRESULT
	Finish_QueryInterface(ppvObject unsafe.Pointer) HRESULT
	Begin_AddRef() HRESULT
	Finish_AddRef() uint32
	Begin_Release() HRESULT
	Finish_Release() uint32
}

type AsyncIUnknownVtbl struct {
	IUnknownVtbl
	Begin_QueryInterface  uintptr
	Finish_QueryInterface uintptr
	Begin_AddRef          uintptr
	Finish_AddRef         uintptr
	Begin_Release         uintptr
	Finish_Release        uintptr
}

type AsyncIUnknown struct {
	IUnknown
}

func (this *AsyncIUnknown) Vtbl() *AsyncIUnknownVtbl {
	return (*AsyncIUnknownVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *AsyncIUnknown) Begin_QueryInterface(riid *syscall.GUID) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Begin_QueryInterface, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(riid)))
	return HRESULT(ret)
}

func (this *AsyncIUnknown) Finish_QueryInterface(ppvObject unsafe.Pointer) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Finish_QueryInterface, uintptr(unsafe.Pointer(this)), uintptr(ppvObject))
	return HRESULT(ret)
}

func (this *AsyncIUnknown) Begin_AddRef() HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Begin_AddRef, uintptr(unsafe.Pointer(this)))
	return HRESULT(ret)
}

func (this *AsyncIUnknown) Finish_AddRef() uint32 {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Finish_AddRef, uintptr(unsafe.Pointer(this)))
	return uint32(ret)
}

func (this *AsyncIUnknown) Begin_Release() HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Begin_Release, uintptr(unsafe.Pointer(this)))
	return HRESULT(ret)
}

func (this *AsyncIUnknown) Finish_Release() uint32 {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Finish_Release, uintptr(unsafe.Pointer(this)))
	return uint32(ret)
}

// 00000001-0000-0000-C000-000000000046
var IID_IClassFactory = syscall.GUID{0x00000001, 0x0000, 0x0000,
	[8]byte{0xC0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x46}}

type IClassFactoryInterface interface {
	IUnknownInterface
	CreateInstance(pUnkOuter *IUnknown, riid *syscall.GUID, ppvObject unsafe.Pointer) HRESULT
	LockServer(fLock BOOL) HRESULT
}

type IClassFactoryVtbl struct {
	IUnknownVtbl
	CreateInstance uintptr
	LockServer     uintptr
}

type IClassFactory struct {
	IUnknown
}

func (this *IClassFactory) Vtbl() *IClassFactoryVtbl {
	return (*IClassFactoryVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IClassFactory) CreateInstance(pUnkOuter *IUnknown, riid *syscall.GUID, ppvObject unsafe.Pointer) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().CreateInstance, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pUnkOuter)), uintptr(unsafe.Pointer(riid)), uintptr(ppvObject))
	return HRESULT(ret)
}

func (this *IClassFactory) LockServer(fLock BOOL) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().LockServer, uintptr(unsafe.Pointer(this)), uintptr(fLock))
	return HRESULT(ret)
}

// ECC8691B-C1DB-4DC0-855E-65F6C551AF49
var IID_INoMarshal = syscall.GUID{0xECC8691B, 0xC1DB, 0x4DC0,
	[8]byte{0x85, 0x5E, 0x65, 0xF6, 0xC5, 0x51, 0xAF, 0x49}}

type INoMarshalInterface interface {
	IUnknownInterface
}

type INoMarshalVtbl struct {
	IUnknownVtbl
}

type INoMarshal struct {
	IUnknown
}

func (this *INoMarshal) Vtbl() *INoMarshalVtbl {
	return (*INoMarshalVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

// 94EA2B94-E9CC-49E0-C0FF-EE64CA8F5B90
var IID_IAgileObject = syscall.GUID{0x94EA2B94, 0xE9CC, 0x49E0,
	[8]byte{0xC0, 0xFF, 0xEE, 0x64, 0xCA, 0x8F, 0x5B, 0x90}}

type IAgileObjectInterface interface {
	IUnknownInterface
}

type IAgileObjectVtbl struct {
	IUnknownVtbl
}

type IAgileObject struct {
	IUnknown
}

func (this *IAgileObject) Vtbl() *IAgileObjectVtbl {
	return (*IAgileObjectVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

// 00000017-0000-0000-C000-000000000046
var IID_IActivationFilter = syscall.GUID{0x00000017, 0x0000, 0x0000,
	[8]byte{0xC0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x46}}

type IActivationFilterInterface interface {
	IUnknownInterface
	HandleActivation(dwActivationType uint32, rclsid *syscall.GUID, pReplacementClsId *syscall.GUID) HRESULT
}

type IActivationFilterVtbl struct {
	IUnknownVtbl
	HandleActivation uintptr
}

type IActivationFilter struct {
	IUnknown
}

func (this *IActivationFilter) Vtbl() *IActivationFilterVtbl {
	return (*IActivationFilterVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IActivationFilter) HandleActivation(dwActivationType uint32, rclsid *syscall.GUID, pReplacementClsId *syscall.GUID) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().HandleActivation, uintptr(unsafe.Pointer(this)), uintptr(dwActivationType), uintptr(unsafe.Pointer(rclsid)), uintptr(unsafe.Pointer(pReplacementClsId)))
	return HRESULT(ret)
}

// 00000002-0000-0000-C000-000000000046
var IID_IMalloc = syscall.GUID{0x00000002, 0x0000, 0x0000,
	[8]byte{0xC0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x46}}

type IMallocInterface interface {
	IUnknownInterface
	Alloc(cb uintptr) unsafe.Pointer
	Realloc(pv unsafe.Pointer, cb uintptr) unsafe.Pointer
	Free(pv unsafe.Pointer)
	GetSize(pv unsafe.Pointer) uintptr
	DidAlloc(pv unsafe.Pointer) int32
	HeapMinimize()
}

type IMallocVtbl struct {
	IUnknownVtbl
	Alloc        uintptr
	Realloc      uintptr
	Free         uintptr
	GetSize      uintptr
	DidAlloc     uintptr
	HeapMinimize uintptr
}

type IMalloc struct {
	IUnknown
}

func (this *IMalloc) Vtbl() *IMallocVtbl {
	return (*IMallocVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IMalloc) Alloc(cb uintptr) unsafe.Pointer {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Alloc, uintptr(unsafe.Pointer(this)), cb)
	return (unsafe.Pointer)(ret)
}

func (this *IMalloc) Realloc(pv unsafe.Pointer, cb uintptr) unsafe.Pointer {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Realloc, uintptr(unsafe.Pointer(this)), uintptr(pv), cb)
	return (unsafe.Pointer)(ret)
}

func (this *IMalloc) Free(pv unsafe.Pointer) {
	_, _, _ = syscall.SyscallN(this.Vtbl().Free, uintptr(unsafe.Pointer(this)), uintptr(pv))
}

func (this *IMalloc) GetSize(pv unsafe.Pointer) uintptr {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetSize, uintptr(unsafe.Pointer(this)), uintptr(pv))
	return ret
}

func (this *IMalloc) DidAlloc(pv unsafe.Pointer) int32 {
	ret, _, _ := syscall.SyscallN(this.Vtbl().DidAlloc, uintptr(unsafe.Pointer(this)), uintptr(pv))
	return int32(ret)
}

func (this *IMalloc) HeapMinimize() {
	_, _, _ = syscall.SyscallN(this.Vtbl().HeapMinimize, uintptr(unsafe.Pointer(this)))
}

// 00000018-0000-0000-C000-000000000046
var IID_IStdMarshalInfo = syscall.GUID{0x00000018, 0x0000, 0x0000,
	[8]byte{0xC0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x46}}

type IStdMarshalInfoInterface interface {
	IUnknownInterface
	GetClassForHandler(dwDestContext uint32, pvDestContext unsafe.Pointer, pClsid *syscall.GUID) HRESULT
}

type IStdMarshalInfoVtbl struct {
	IUnknownVtbl
	GetClassForHandler uintptr
}

type IStdMarshalInfo struct {
	IUnknown
}

func (this *IStdMarshalInfo) Vtbl() *IStdMarshalInfoVtbl {
	return (*IStdMarshalInfoVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IStdMarshalInfo) GetClassForHandler(dwDestContext uint32, pvDestContext unsafe.Pointer, pClsid *syscall.GUID) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetClassForHandler, uintptr(unsafe.Pointer(this)), uintptr(dwDestContext), uintptr(pvDestContext), uintptr(unsafe.Pointer(pClsid)))
	return HRESULT(ret)
}

// 00000019-0000-0000-C000-000000000046
var IID_IExternalConnection = syscall.GUID{0x00000019, 0x0000, 0x0000,
	[8]byte{0xC0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x46}}

type IExternalConnectionInterface interface {
	IUnknownInterface
	AddConnection(extconn uint32, reserved uint32) uint32
	ReleaseConnection(extconn uint32, reserved uint32, fLastReleaseCloses BOOL) uint32
}

type IExternalConnectionVtbl struct {
	IUnknownVtbl
	AddConnection     uintptr
	ReleaseConnection uintptr
}

type IExternalConnection struct {
	IUnknown
}

func (this *IExternalConnection) Vtbl() *IExternalConnectionVtbl {
	return (*IExternalConnectionVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IExternalConnection) AddConnection(extconn uint32, reserved uint32) uint32 {
	ret, _, _ := syscall.SyscallN(this.Vtbl().AddConnection, uintptr(unsafe.Pointer(this)), uintptr(extconn), uintptr(reserved))
	return uint32(ret)
}

func (this *IExternalConnection) ReleaseConnection(extconn uint32, reserved uint32, fLastReleaseCloses BOOL) uint32 {
	ret, _, _ := syscall.SyscallN(this.Vtbl().ReleaseConnection, uintptr(unsafe.Pointer(this)), uintptr(extconn), uintptr(reserved), uintptr(fLastReleaseCloses))
	return uint32(ret)
}

// 00000020-0000-0000-C000-000000000046
var IID_IMultiQI = syscall.GUID{0x00000020, 0x0000, 0x0000,
	[8]byte{0xC0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x46}}

type IMultiQIInterface interface {
	IUnknownInterface
	QueryMultipleInterfaces(cMQIs uint32, pMQIs *MULTI_QI) HRESULT
}

type IMultiQIVtbl struct {
	IUnknownVtbl
	QueryMultipleInterfaces uintptr
}

type IMultiQI struct {
	IUnknown
}

func (this *IMultiQI) Vtbl() *IMultiQIVtbl {
	return (*IMultiQIVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IMultiQI) QueryMultipleInterfaces(cMQIs uint32, pMQIs *MULTI_QI) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().QueryMultipleInterfaces, uintptr(unsafe.Pointer(this)), uintptr(cMQIs), uintptr(unsafe.Pointer(pMQIs)))
	return HRESULT(ret)
}

// 000E0020-0000-0000-C000-000000000046
var IID_AsyncIMultiQI = syscall.GUID{0x000E0020, 0x0000, 0x0000,
	[8]byte{0xC0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x46}}

type AsyncIMultiQIInterface interface {
	IUnknownInterface
	Begin_QueryMultipleInterfaces(cMQIs uint32, pMQIs *MULTI_QI) HRESULT
	Finish_QueryMultipleInterfaces(pMQIs *MULTI_QI) HRESULT
}

type AsyncIMultiQIVtbl struct {
	IUnknownVtbl
	Begin_QueryMultipleInterfaces  uintptr
	Finish_QueryMultipleInterfaces uintptr
}

type AsyncIMultiQI struct {
	IUnknown
}

func (this *AsyncIMultiQI) Vtbl() *AsyncIMultiQIVtbl {
	return (*AsyncIMultiQIVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *AsyncIMultiQI) Begin_QueryMultipleInterfaces(cMQIs uint32, pMQIs *MULTI_QI) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Begin_QueryMultipleInterfaces, uintptr(unsafe.Pointer(this)), uintptr(cMQIs), uintptr(unsafe.Pointer(pMQIs)))
	return HRESULT(ret)
}

func (this *AsyncIMultiQI) Finish_QueryMultipleInterfaces(pMQIs *MULTI_QI) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Finish_QueryMultipleInterfaces, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pMQIs)))
	return HRESULT(ret)
}

// 00000021-0000-0000-C000-000000000046
var IID_IInternalUnknown = syscall.GUID{0x00000021, 0x0000, 0x0000,
	[8]byte{0xC0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x46}}

type IInternalUnknownInterface interface {
	IUnknownInterface
	QueryInternalInterface(riid *syscall.GUID, ppv unsafe.Pointer) HRESULT
}

type IInternalUnknownVtbl struct {
	IUnknownVtbl
	QueryInternalInterface uintptr
}

type IInternalUnknown struct {
	IUnknown
}

func (this *IInternalUnknown) Vtbl() *IInternalUnknownVtbl {
	return (*IInternalUnknownVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IInternalUnknown) QueryInternalInterface(riid *syscall.GUID, ppv unsafe.Pointer) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().QueryInternalInterface, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(riid)), uintptr(ppv))
	return HRESULT(ret)
}

// 00000100-0000-0000-C000-000000000046
var IID_IEnumUnknown = syscall.GUID{0x00000100, 0x0000, 0x0000,
	[8]byte{0xC0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x46}}

type IEnumUnknownInterface interface {
	IUnknownInterface
	Next(celt uint32, rgelt **IUnknown, pceltFetched *uint32) HRESULT
	Skip(celt uint32) HRESULT
	Reset() HRESULT
	Clone(ppenum **IEnumUnknown) HRESULT
}

type IEnumUnknownVtbl struct {
	IUnknownVtbl
	Next  uintptr
	Skip  uintptr
	Reset uintptr
	Clone uintptr
}

type IEnumUnknown struct {
	IUnknown
}

func (this *IEnumUnknown) Vtbl() *IEnumUnknownVtbl {
	return (*IEnumUnknownVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IEnumUnknown) Next(celt uint32, rgelt **IUnknown, pceltFetched *uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Next, uintptr(unsafe.Pointer(this)), uintptr(celt), uintptr(unsafe.Pointer(rgelt)), uintptr(unsafe.Pointer(pceltFetched)))
	return HRESULT(ret)
}

func (this *IEnumUnknown) Skip(celt uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Skip, uintptr(unsafe.Pointer(this)), uintptr(celt))
	return HRESULT(ret)
}

func (this *IEnumUnknown) Reset() HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Reset, uintptr(unsafe.Pointer(this)))
	return HRESULT(ret)
}

func (this *IEnumUnknown) Clone(ppenum **IEnumUnknown) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Clone, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(ppenum)))
	return HRESULT(ret)
}

// 00000101-0000-0000-C000-000000000046
var IID_IEnumString = syscall.GUID{0x00000101, 0x0000, 0x0000,
	[8]byte{0xC0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x46}}

type IEnumStringInterface interface {
	IUnknownInterface
	Next(celt uint32, rgelt *PWSTR, pceltFetched *uint32) HRESULT
	Skip(celt uint32) HRESULT
	Reset() HRESULT
	Clone(ppenum **IEnumString) HRESULT
}

type IEnumStringVtbl struct {
	IUnknownVtbl
	Next  uintptr
	Skip  uintptr
	Reset uintptr
	Clone uintptr
}

type IEnumString struct {
	IUnknown
}

func (this *IEnumString) Vtbl() *IEnumStringVtbl {
	return (*IEnumStringVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IEnumString) Next(celt uint32, rgelt *PWSTR, pceltFetched *uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Next, uintptr(unsafe.Pointer(this)), uintptr(celt), uintptr(unsafe.Pointer(rgelt)), uintptr(unsafe.Pointer(pceltFetched)))
	return HRESULT(ret)
}

func (this *IEnumString) Skip(celt uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Skip, uintptr(unsafe.Pointer(this)), uintptr(celt))
	return HRESULT(ret)
}

func (this *IEnumString) Reset() HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Reset, uintptr(unsafe.Pointer(this)))
	return HRESULT(ret)
}

func (this *IEnumString) Clone(ppenum **IEnumString) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Clone, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(ppenum)))
	return HRESULT(ret)
}

// 0C733A30-2A1C-11CE-ADE5-00AA0044773D
var IID_ISequentialStream = syscall.GUID{0x0C733A30, 0x2A1C, 0x11CE,
	[8]byte{0xAD, 0xE5, 0x00, 0xAA, 0x00, 0x44, 0x77, 0x3D}}

type ISequentialStreamInterface interface {
	IUnknownInterface
	Read(pv unsafe.Pointer, cb uint32, pcbRead *uint32) HRESULT
	Write(pv unsafe.Pointer, cb uint32, pcbWritten *uint32) HRESULT
}

type ISequentialStreamVtbl struct {
	IUnknownVtbl
	Read  uintptr
	Write uintptr
}

type ISequentialStream struct {
	IUnknown
}

func (this *ISequentialStream) Vtbl() *ISequentialStreamVtbl {
	return (*ISequentialStreamVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *ISequentialStream) Read(pv unsafe.Pointer, cb uint32, pcbRead *uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Read, uintptr(unsafe.Pointer(this)), uintptr(pv), uintptr(cb), uintptr(unsafe.Pointer(pcbRead)))
	return HRESULT(ret)
}

func (this *ISequentialStream) Write(pv unsafe.Pointer, cb uint32, pcbWritten *uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Write, uintptr(unsafe.Pointer(this)), uintptr(pv), uintptr(cb), uintptr(unsafe.Pointer(pcbWritten)))
	return HRESULT(ret)
}

// 0000000C-0000-0000-C000-000000000046
var IID_IStream = syscall.GUID{0x0000000C, 0x0000, 0x0000,
	[8]byte{0xC0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x46}}

type IStreamInterface interface {
	ISequentialStreamInterface
	Seek(dlibMove int64, dwOrigin STREAM_SEEK, plibNewPosition *uint64) HRESULT
	SetSize(libNewSize uint64) HRESULT
	CopyTo(pstm *IStream, cb uint64, pcbRead *uint64, pcbWritten *uint64) HRESULT
	Commit(grfCommitFlags STGC) HRESULT
	Revert() HRESULT
	LockRegion(libOffset uint64, cb uint64, dwLockType uint32) HRESULT
	UnlockRegion(libOffset uint64, cb uint64, dwLockType uint32) HRESULT
	Stat(pstatstg *STATSTG, grfStatFlag uint32) HRESULT
	Clone(ppstm **IStream) HRESULT
}

type IStreamVtbl struct {
	ISequentialStreamVtbl
	Seek         uintptr
	SetSize      uintptr
	CopyTo       uintptr
	Commit       uintptr
	Revert       uintptr
	LockRegion   uintptr
	UnlockRegion uintptr
	Stat         uintptr
	Clone        uintptr
}

type IStream struct {
	ISequentialStream
}

func (this *IStream) Vtbl() *IStreamVtbl {
	return (*IStreamVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IStream) Seek(dlibMove int64, dwOrigin STREAM_SEEK, plibNewPosition *uint64) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Seek, uintptr(unsafe.Pointer(this)), uintptr(dlibMove), uintptr(dwOrigin), uintptr(unsafe.Pointer(plibNewPosition)))
	return HRESULT(ret)
}

func (this *IStream) SetSize(libNewSize uint64) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetSize, uintptr(unsafe.Pointer(this)), uintptr(libNewSize))
	return HRESULT(ret)
}

func (this *IStream) CopyTo(pstm *IStream, cb uint64, pcbRead *uint64, pcbWritten *uint64) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().CopyTo, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pstm)), uintptr(cb), uintptr(unsafe.Pointer(pcbRead)), uintptr(unsafe.Pointer(pcbWritten)))
	return HRESULT(ret)
}

func (this *IStream) Commit(grfCommitFlags STGC) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Commit, uintptr(unsafe.Pointer(this)), uintptr(grfCommitFlags))
	return HRESULT(ret)
}

func (this *IStream) Revert() HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Revert, uintptr(unsafe.Pointer(this)))
	return HRESULT(ret)
}

func (this *IStream) LockRegion(libOffset uint64, cb uint64, dwLockType uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().LockRegion, uintptr(unsafe.Pointer(this)), uintptr(libOffset), uintptr(cb), uintptr(dwLockType))
	return HRESULT(ret)
}

func (this *IStream) UnlockRegion(libOffset uint64, cb uint64, dwLockType uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().UnlockRegion, uintptr(unsafe.Pointer(this)), uintptr(libOffset), uintptr(cb), uintptr(dwLockType))
	return HRESULT(ret)
}

func (this *IStream) Stat(pstatstg *STATSTG, grfStatFlag uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Stat, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pstatstg)), uintptr(grfStatFlag))
	return HRESULT(ret)
}

func (this *IStream) Clone(ppstm **IStream) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Clone, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(ppstm)))
	return HRESULT(ret)
}

// D5F56B60-593B-101A-B569-08002B2DBF7A
var IID_IRpcChannelBuffer = syscall.GUID{0xD5F56B60, 0x593B, 0x101A,
	[8]byte{0xB5, 0x69, 0x08, 0x00, 0x2B, 0x2D, 0xBF, 0x7A}}

type IRpcChannelBufferInterface interface {
	IUnknownInterface
	GetBuffer(pMessage *RPCOLEMESSAGE, riid *syscall.GUID) HRESULT
	SendReceive(pMessage *RPCOLEMESSAGE, pStatus *uint32) HRESULT
	FreeBuffer(pMessage *RPCOLEMESSAGE) HRESULT
	GetDestCtx(pdwDestContext *uint32, ppvDestContext unsafe.Pointer) HRESULT
	IsConnected() HRESULT
}

type IRpcChannelBufferVtbl struct {
	IUnknownVtbl
	GetBuffer   uintptr
	SendReceive uintptr
	FreeBuffer  uintptr
	GetDestCtx  uintptr
	IsConnected uintptr
}

type IRpcChannelBuffer struct {
	IUnknown
}

func (this *IRpcChannelBuffer) Vtbl() *IRpcChannelBufferVtbl {
	return (*IRpcChannelBufferVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IRpcChannelBuffer) GetBuffer(pMessage *RPCOLEMESSAGE, riid *syscall.GUID) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetBuffer, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pMessage)), uintptr(unsafe.Pointer(riid)))
	return HRESULT(ret)
}

func (this *IRpcChannelBuffer) SendReceive(pMessage *RPCOLEMESSAGE, pStatus *uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SendReceive, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pMessage)), uintptr(unsafe.Pointer(pStatus)))
	return HRESULT(ret)
}

func (this *IRpcChannelBuffer) FreeBuffer(pMessage *RPCOLEMESSAGE) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().FreeBuffer, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pMessage)))
	return HRESULT(ret)
}

func (this *IRpcChannelBuffer) GetDestCtx(pdwDestContext *uint32, ppvDestContext unsafe.Pointer) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetDestCtx, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pdwDestContext)), uintptr(ppvDestContext))
	return HRESULT(ret)
}

func (this *IRpcChannelBuffer) IsConnected() HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().IsConnected, uintptr(unsafe.Pointer(this)))
	return HRESULT(ret)
}

// 594F31D0-7F19-11D0-B194-00A0C90DC8BF
var IID_IRpcChannelBuffer2 = syscall.GUID{0x594F31D0, 0x7F19, 0x11D0,
	[8]byte{0xB1, 0x94, 0x00, 0xA0, 0xC9, 0x0D, 0xC8, 0xBF}}

type IRpcChannelBuffer2Interface interface {
	IRpcChannelBufferInterface
	GetProtocolVersion(pdwVersion *uint32) HRESULT
}

type IRpcChannelBuffer2Vtbl struct {
	IRpcChannelBufferVtbl
	GetProtocolVersion uintptr
}

type IRpcChannelBuffer2 struct {
	IRpcChannelBuffer
}

func (this *IRpcChannelBuffer2) Vtbl() *IRpcChannelBuffer2Vtbl {
	return (*IRpcChannelBuffer2Vtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IRpcChannelBuffer2) GetProtocolVersion(pdwVersion *uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetProtocolVersion, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pdwVersion)))
	return HRESULT(ret)
}

// A5029FB6-3C34-11D1-9C99-00C04FB998AA
var IID_IAsyncRpcChannelBuffer = syscall.GUID{0xA5029FB6, 0x3C34, 0x11D1,
	[8]byte{0x9C, 0x99, 0x00, 0xC0, 0x4F, 0xB9, 0x98, 0xAA}}

type IAsyncRpcChannelBufferInterface interface {
	IRpcChannelBuffer2Interface
	Send(pMsg *RPCOLEMESSAGE, pSync *ISynchronize, pulStatus *uint32) HRESULT
	Receive(pMsg *RPCOLEMESSAGE, pulStatus *uint32) HRESULT
	GetDestCtxEx(pMsg *RPCOLEMESSAGE, pdwDestContext *uint32, ppvDestContext unsafe.Pointer) HRESULT
}

type IAsyncRpcChannelBufferVtbl struct {
	IRpcChannelBuffer2Vtbl
	Send         uintptr
	Receive      uintptr
	GetDestCtxEx uintptr
}

type IAsyncRpcChannelBuffer struct {
	IRpcChannelBuffer2
}

func (this *IAsyncRpcChannelBuffer) Vtbl() *IAsyncRpcChannelBufferVtbl {
	return (*IAsyncRpcChannelBufferVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IAsyncRpcChannelBuffer) Send(pMsg *RPCOLEMESSAGE, pSync *ISynchronize, pulStatus *uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Send, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pMsg)), uintptr(unsafe.Pointer(pSync)), uintptr(unsafe.Pointer(pulStatus)))
	return HRESULT(ret)
}

func (this *IAsyncRpcChannelBuffer) Receive(pMsg *RPCOLEMESSAGE, pulStatus *uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Receive, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pMsg)), uintptr(unsafe.Pointer(pulStatus)))
	return HRESULT(ret)
}

func (this *IAsyncRpcChannelBuffer) GetDestCtxEx(pMsg *RPCOLEMESSAGE, pdwDestContext *uint32, ppvDestContext unsafe.Pointer) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetDestCtxEx, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pMsg)), uintptr(unsafe.Pointer(pdwDestContext)), uintptr(ppvDestContext))
	return HRESULT(ret)
}

// 25B15600-0115-11D0-BF0D-00AA00B8DFD2
var IID_IRpcChannelBuffer3 = syscall.GUID{0x25B15600, 0x0115, 0x11D0,
	[8]byte{0xBF, 0x0D, 0x00, 0xAA, 0x00, 0xB8, 0xDF, 0xD2}}

type IRpcChannelBuffer3Interface interface {
	IRpcChannelBuffer2Interface
	Send(pMsg *RPCOLEMESSAGE, pulStatus *uint32) HRESULT
	Receive(pMsg *RPCOLEMESSAGE, ulSize uint32, pulStatus *uint32) HRESULT
	Cancel(pMsg *RPCOLEMESSAGE) HRESULT
	GetCallContext(pMsg *RPCOLEMESSAGE, riid *syscall.GUID, pInterface unsafe.Pointer) HRESULT
	GetDestCtxEx(pMsg *RPCOLEMESSAGE, pdwDestContext *uint32, ppvDestContext unsafe.Pointer) HRESULT
	GetState(pMsg *RPCOLEMESSAGE, pState *uint32) HRESULT
	RegisterAsync(pMsg *RPCOLEMESSAGE, pAsyncMgr *IAsyncManager) HRESULT
}

type IRpcChannelBuffer3Vtbl struct {
	IRpcChannelBuffer2Vtbl
	Send           uintptr
	Receive        uintptr
	Cancel         uintptr
	GetCallContext uintptr
	GetDestCtxEx   uintptr
	GetState       uintptr
	RegisterAsync  uintptr
}

type IRpcChannelBuffer3 struct {
	IRpcChannelBuffer2
}

func (this *IRpcChannelBuffer3) Vtbl() *IRpcChannelBuffer3Vtbl {
	return (*IRpcChannelBuffer3Vtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IRpcChannelBuffer3) Send(pMsg *RPCOLEMESSAGE, pulStatus *uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Send, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pMsg)), uintptr(unsafe.Pointer(pulStatus)))
	return HRESULT(ret)
}

func (this *IRpcChannelBuffer3) Receive(pMsg *RPCOLEMESSAGE, ulSize uint32, pulStatus *uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Receive, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pMsg)), uintptr(ulSize), uintptr(unsafe.Pointer(pulStatus)))
	return HRESULT(ret)
}

func (this *IRpcChannelBuffer3) Cancel(pMsg *RPCOLEMESSAGE) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Cancel, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pMsg)))
	return HRESULT(ret)
}

func (this *IRpcChannelBuffer3) GetCallContext(pMsg *RPCOLEMESSAGE, riid *syscall.GUID, pInterface unsafe.Pointer) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetCallContext, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pMsg)), uintptr(unsafe.Pointer(riid)), uintptr(pInterface))
	return HRESULT(ret)
}

func (this *IRpcChannelBuffer3) GetDestCtxEx(pMsg *RPCOLEMESSAGE, pdwDestContext *uint32, ppvDestContext unsafe.Pointer) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetDestCtxEx, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pMsg)), uintptr(unsafe.Pointer(pdwDestContext)), uintptr(ppvDestContext))
	return HRESULT(ret)
}

func (this *IRpcChannelBuffer3) GetState(pMsg *RPCOLEMESSAGE, pState *uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetState, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pMsg)), uintptr(unsafe.Pointer(pState)))
	return HRESULT(ret)
}

func (this *IRpcChannelBuffer3) RegisterAsync(pMsg *RPCOLEMESSAGE, pAsyncMgr *IAsyncManager) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().RegisterAsync, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pMsg)), uintptr(unsafe.Pointer(pAsyncMgr)))
	return HRESULT(ret)
}

// 58A08519-24C8-4935-B482-3FD823333A4F
var IID_IRpcSyntaxNegotiate = syscall.GUID{0x58A08519, 0x24C8, 0x4935,
	[8]byte{0xB4, 0x82, 0x3F, 0xD8, 0x23, 0x33, 0x3A, 0x4F}}

type IRpcSyntaxNegotiateInterface interface {
	IUnknownInterface
	NegotiateSyntax(pMsg *RPCOLEMESSAGE) HRESULT
}

type IRpcSyntaxNegotiateVtbl struct {
	IUnknownVtbl
	NegotiateSyntax uintptr
}

type IRpcSyntaxNegotiate struct {
	IUnknown
}

func (this *IRpcSyntaxNegotiate) Vtbl() *IRpcSyntaxNegotiateVtbl {
	return (*IRpcSyntaxNegotiateVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IRpcSyntaxNegotiate) NegotiateSyntax(pMsg *RPCOLEMESSAGE) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().NegotiateSyntax, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pMsg)))
	return HRESULT(ret)
}

// D5F56A34-593B-101A-B569-08002B2DBF7A
var IID_IRpcProxyBuffer = syscall.GUID{0xD5F56A34, 0x593B, 0x101A,
	[8]byte{0xB5, 0x69, 0x08, 0x00, 0x2B, 0x2D, 0xBF, 0x7A}}

type IRpcProxyBufferInterface interface {
	IUnknownInterface
	Connect(pRpcChannelBuffer *IRpcChannelBuffer) HRESULT
	Disconnect()
}

type IRpcProxyBufferVtbl struct {
	IUnknownVtbl
	Connect    uintptr
	Disconnect uintptr
}

type IRpcProxyBuffer struct {
	IUnknown
}

func (this *IRpcProxyBuffer) Vtbl() *IRpcProxyBufferVtbl {
	return (*IRpcProxyBufferVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IRpcProxyBuffer) Connect(pRpcChannelBuffer *IRpcChannelBuffer) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Connect, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pRpcChannelBuffer)))
	return HRESULT(ret)
}

func (this *IRpcProxyBuffer) Disconnect() {
	_, _, _ = syscall.SyscallN(this.Vtbl().Disconnect, uintptr(unsafe.Pointer(this)))
}

// D5F56AFC-593B-101A-B569-08002B2DBF7A
var IID_IRpcStubBuffer = syscall.GUID{0xD5F56AFC, 0x593B, 0x101A,
	[8]byte{0xB5, 0x69, 0x08, 0x00, 0x2B, 0x2D, 0xBF, 0x7A}}

type IRpcStubBufferInterface interface {
	IUnknownInterface
	Connect(pUnkServer *IUnknown) HRESULT
	Disconnect()
	Invoke(_prpcmsg *RPCOLEMESSAGE, _pRpcChannelBuffer *IRpcChannelBuffer) HRESULT
	IsIIDSupported(riid *syscall.GUID) *IRpcStubBuffer
	CountRefs() uint32
	DebugServerQueryInterface(ppv unsafe.Pointer) HRESULT
	DebugServerRelease(pv unsafe.Pointer)
}

type IRpcStubBufferVtbl struct {
	IUnknownVtbl
	Connect                   uintptr
	Disconnect                uintptr
	Invoke                    uintptr
	IsIIDSupported            uintptr
	CountRefs                 uintptr
	DebugServerQueryInterface uintptr
	DebugServerRelease        uintptr
}

type IRpcStubBuffer struct {
	IUnknown
}

func (this *IRpcStubBuffer) Vtbl() *IRpcStubBufferVtbl {
	return (*IRpcStubBufferVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IRpcStubBuffer) Connect(pUnkServer *IUnknown) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Connect, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pUnkServer)))
	return HRESULT(ret)
}

func (this *IRpcStubBuffer) Disconnect() {
	_, _, _ = syscall.SyscallN(this.Vtbl().Disconnect, uintptr(unsafe.Pointer(this)))
}

func (this *IRpcStubBuffer) Invoke(_prpcmsg *RPCOLEMESSAGE, _pRpcChannelBuffer *IRpcChannelBuffer) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Invoke, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(_prpcmsg)), uintptr(unsafe.Pointer(_pRpcChannelBuffer)))
	return HRESULT(ret)
}

func (this *IRpcStubBuffer) IsIIDSupported(riid *syscall.GUID) *IRpcStubBuffer {
	ret, _, _ := syscall.SyscallN(this.Vtbl().IsIIDSupported, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(riid)))
	return (*IRpcStubBuffer)(unsafe.Pointer(ret))
}

func (this *IRpcStubBuffer) CountRefs() uint32 {
	ret, _, _ := syscall.SyscallN(this.Vtbl().CountRefs, uintptr(unsafe.Pointer(this)))
	return uint32(ret)
}

func (this *IRpcStubBuffer) DebugServerQueryInterface(ppv unsafe.Pointer) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().DebugServerQueryInterface, uintptr(unsafe.Pointer(this)), uintptr(ppv))
	return HRESULT(ret)
}

func (this *IRpcStubBuffer) DebugServerRelease(pv unsafe.Pointer) {
	_, _, _ = syscall.SyscallN(this.Vtbl().DebugServerRelease, uintptr(unsafe.Pointer(this)), uintptr(pv))
}

// D5F569D0-593B-101A-B569-08002B2DBF7A
var IID_IPSFactoryBuffer = syscall.GUID{0xD5F569D0, 0x593B, 0x101A,
	[8]byte{0xB5, 0x69, 0x08, 0x00, 0x2B, 0x2D, 0xBF, 0x7A}}

type IPSFactoryBufferInterface interface {
	IUnknownInterface
	CreateProxy(pUnkOuter *IUnknown, riid *syscall.GUID, ppProxy **IRpcProxyBuffer, ppv unsafe.Pointer) HRESULT
	CreateStub(riid *syscall.GUID, pUnkServer *IUnknown, ppStub **IRpcStubBuffer) HRESULT
}

type IPSFactoryBufferVtbl struct {
	IUnknownVtbl
	CreateProxy uintptr
	CreateStub  uintptr
}

type IPSFactoryBuffer struct {
	IUnknown
}

func (this *IPSFactoryBuffer) Vtbl() *IPSFactoryBufferVtbl {
	return (*IPSFactoryBufferVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IPSFactoryBuffer) CreateProxy(pUnkOuter *IUnknown, riid *syscall.GUID, ppProxy **IRpcProxyBuffer, ppv unsafe.Pointer) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().CreateProxy, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pUnkOuter)), uintptr(unsafe.Pointer(riid)), uintptr(unsafe.Pointer(ppProxy)), uintptr(ppv))
	return HRESULT(ret)
}

func (this *IPSFactoryBuffer) CreateStub(riid *syscall.GUID, pUnkServer *IUnknown, ppStub **IRpcStubBuffer) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().CreateStub, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(riid)), uintptr(unsafe.Pointer(pUnkServer)), uintptr(unsafe.Pointer(ppStub)))
	return HRESULT(ret)
}

// 1008C4A0-7613-11CF-9AF1-0020AF6E72F4
var IID_IChannelHook = syscall.GUID{0x1008C4A0, 0x7613, 0x11CF,
	[8]byte{0x9A, 0xF1, 0x00, 0x20, 0xAF, 0x6E, 0x72, 0xF4}}

type IChannelHookInterface interface {
	IUnknownInterface
	ClientGetSize(uExtent *syscall.GUID, riid *syscall.GUID, pDataSize *uint32)
	ClientFillBuffer(uExtent *syscall.GUID, riid *syscall.GUID, pDataSize *uint32, pDataBuffer unsafe.Pointer)
	ClientNotify(uExtent *syscall.GUID, riid *syscall.GUID, cbDataSize uint32, pDataBuffer unsafe.Pointer, lDataRep uint32, hrFault HRESULT)
	ServerNotify(uExtent *syscall.GUID, riid *syscall.GUID, cbDataSize uint32, pDataBuffer unsafe.Pointer, lDataRep uint32)
	ServerGetSize(uExtent *syscall.GUID, riid *syscall.GUID, hrFault HRESULT, pDataSize *uint32)
	ServerFillBuffer(uExtent *syscall.GUID, riid *syscall.GUID, pDataSize *uint32, pDataBuffer unsafe.Pointer, hrFault HRESULT)
}

type IChannelHookVtbl struct {
	IUnknownVtbl
	ClientGetSize    uintptr
	ClientFillBuffer uintptr
	ClientNotify     uintptr
	ServerNotify     uintptr
	ServerGetSize    uintptr
	ServerFillBuffer uintptr
}

type IChannelHook struct {
	IUnknown
}

func (this *IChannelHook) Vtbl() *IChannelHookVtbl {
	return (*IChannelHookVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IChannelHook) ClientGetSize(uExtent *syscall.GUID, riid *syscall.GUID, pDataSize *uint32) {
	_, _, _ = syscall.SyscallN(this.Vtbl().ClientGetSize, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(uExtent)), uintptr(unsafe.Pointer(riid)), uintptr(unsafe.Pointer(pDataSize)))
}

func (this *IChannelHook) ClientFillBuffer(uExtent *syscall.GUID, riid *syscall.GUID, pDataSize *uint32, pDataBuffer unsafe.Pointer) {
	_, _, _ = syscall.SyscallN(this.Vtbl().ClientFillBuffer, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(uExtent)), uintptr(unsafe.Pointer(riid)), uintptr(unsafe.Pointer(pDataSize)), uintptr(pDataBuffer))
}

func (this *IChannelHook) ClientNotify(uExtent *syscall.GUID, riid *syscall.GUID, cbDataSize uint32, pDataBuffer unsafe.Pointer, lDataRep uint32, hrFault HRESULT) {
	_, _, _ = syscall.SyscallN(this.Vtbl().ClientNotify, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(uExtent)), uintptr(unsafe.Pointer(riid)), uintptr(cbDataSize), uintptr(pDataBuffer), uintptr(lDataRep), uintptr(hrFault))
}

func (this *IChannelHook) ServerNotify(uExtent *syscall.GUID, riid *syscall.GUID, cbDataSize uint32, pDataBuffer unsafe.Pointer, lDataRep uint32) {
	_, _, _ = syscall.SyscallN(this.Vtbl().ServerNotify, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(uExtent)), uintptr(unsafe.Pointer(riid)), uintptr(cbDataSize), uintptr(pDataBuffer), uintptr(lDataRep))
}

func (this *IChannelHook) ServerGetSize(uExtent *syscall.GUID, riid *syscall.GUID, hrFault HRESULT, pDataSize *uint32) {
	_, _, _ = syscall.SyscallN(this.Vtbl().ServerGetSize, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(uExtent)), uintptr(unsafe.Pointer(riid)), uintptr(hrFault), uintptr(unsafe.Pointer(pDataSize)))
}

func (this *IChannelHook) ServerFillBuffer(uExtent *syscall.GUID, riid *syscall.GUID, pDataSize *uint32, pDataBuffer unsafe.Pointer, hrFault HRESULT) {
	_, _, _ = syscall.SyscallN(this.Vtbl().ServerFillBuffer, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(uExtent)), uintptr(unsafe.Pointer(riid)), uintptr(unsafe.Pointer(pDataSize)), uintptr(pDataBuffer), uintptr(hrFault))
}

// 0000013D-0000-0000-C000-000000000046
var IID_IClientSecurity = syscall.GUID{0x0000013D, 0x0000, 0x0000,
	[8]byte{0xC0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x46}}

type IClientSecurityInterface interface {
	IUnknownInterface
	QueryBlanket(pProxy *IUnknown, pAuthnSvc *uint32, pAuthzSvc *uint32, pServerPrincName **uint16, pAuthnLevel *RPC_C_AUTHN_LEVEL, pImpLevel *RPC_C_IMP_LEVEL, pAuthInfo unsafe.Pointer, pCapabilites *EOLE_AUTHENTICATION_CAPABILITIES) HRESULT
	SetBlanket(pProxy *IUnknown, dwAuthnSvc uint32, dwAuthzSvc uint32, pServerPrincName PWSTR, dwAuthnLevel RPC_C_AUTHN_LEVEL, dwImpLevel RPC_C_IMP_LEVEL, pAuthInfo unsafe.Pointer, dwCapabilities EOLE_AUTHENTICATION_CAPABILITIES) HRESULT
	CopyProxy(pProxy *IUnknown, ppCopy **IUnknown) HRESULT
}

type IClientSecurityVtbl struct {
	IUnknownVtbl
	QueryBlanket uintptr
	SetBlanket   uintptr
	CopyProxy    uintptr
}

type IClientSecurity struct {
	IUnknown
}

func (this *IClientSecurity) Vtbl() *IClientSecurityVtbl {
	return (*IClientSecurityVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IClientSecurity) QueryBlanket(pProxy *IUnknown, pAuthnSvc *uint32, pAuthzSvc *uint32, pServerPrincName **uint16, pAuthnLevel *RPC_C_AUTHN_LEVEL, pImpLevel *RPC_C_IMP_LEVEL, pAuthInfo unsafe.Pointer, pCapabilites *EOLE_AUTHENTICATION_CAPABILITIES) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().QueryBlanket, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pProxy)), uintptr(unsafe.Pointer(pAuthnSvc)), uintptr(unsafe.Pointer(pAuthzSvc)), uintptr(unsafe.Pointer(pServerPrincName)), uintptr(unsafe.Pointer(pAuthnLevel)), uintptr(unsafe.Pointer(pImpLevel)), uintptr(pAuthInfo), uintptr(unsafe.Pointer(pCapabilites)))
	return HRESULT(ret)
}

func (this *IClientSecurity) SetBlanket(pProxy *IUnknown, dwAuthnSvc uint32, dwAuthzSvc uint32, pServerPrincName PWSTR, dwAuthnLevel RPC_C_AUTHN_LEVEL, dwImpLevel RPC_C_IMP_LEVEL, pAuthInfo unsafe.Pointer, dwCapabilities EOLE_AUTHENTICATION_CAPABILITIES) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetBlanket, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pProxy)), uintptr(dwAuthnSvc), uintptr(dwAuthzSvc), uintptr(unsafe.Pointer(pServerPrincName)), uintptr(dwAuthnLevel), uintptr(dwImpLevel), uintptr(pAuthInfo), uintptr(dwCapabilities))
	return HRESULT(ret)
}

func (this *IClientSecurity) CopyProxy(pProxy *IUnknown, ppCopy **IUnknown) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().CopyProxy, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pProxy)), uintptr(unsafe.Pointer(ppCopy)))
	return HRESULT(ret)
}

// 0000013E-0000-0000-C000-000000000046
var IID_IServerSecurity = syscall.GUID{0x0000013E, 0x0000, 0x0000,
	[8]byte{0xC0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x46}}

type IServerSecurityInterface interface {
	IUnknownInterface
	QueryBlanket(pAuthnSvc *uint32, pAuthzSvc *uint32, pServerPrincName **uint16, pAuthnLevel *uint32, pImpLevel *uint32, pPrivs unsafe.Pointer, pCapabilities *uint32) HRESULT
	ImpersonateClient() HRESULT
	RevertToSelf() HRESULT
	IsImpersonating() BOOL
}

type IServerSecurityVtbl struct {
	IUnknownVtbl
	QueryBlanket      uintptr
	ImpersonateClient uintptr
	RevertToSelf      uintptr
	IsImpersonating   uintptr
}

type IServerSecurity struct {
	IUnknown
}

func (this *IServerSecurity) Vtbl() *IServerSecurityVtbl {
	return (*IServerSecurityVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IServerSecurity) QueryBlanket(pAuthnSvc *uint32, pAuthzSvc *uint32, pServerPrincName **uint16, pAuthnLevel *uint32, pImpLevel *uint32, pPrivs unsafe.Pointer, pCapabilities *uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().QueryBlanket, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pAuthnSvc)), uintptr(unsafe.Pointer(pAuthzSvc)), uintptr(unsafe.Pointer(pServerPrincName)), uintptr(unsafe.Pointer(pAuthnLevel)), uintptr(unsafe.Pointer(pImpLevel)), uintptr(pPrivs), uintptr(unsafe.Pointer(pCapabilities)))
	return HRESULT(ret)
}

func (this *IServerSecurity) ImpersonateClient() HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().ImpersonateClient, uintptr(unsafe.Pointer(this)))
	return HRESULT(ret)
}

func (this *IServerSecurity) RevertToSelf() HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().RevertToSelf, uintptr(unsafe.Pointer(this)))
	return HRESULT(ret)
}

func (this *IServerSecurity) IsImpersonating() BOOL {
	ret, _, _ := syscall.SyscallN(this.Vtbl().IsImpersonating, uintptr(unsafe.Pointer(this)))
	return BOOL(ret)
}

// 00000144-0000-0000-C000-000000000046
var IID_IRpcOptions = syscall.GUID{0x00000144, 0x0000, 0x0000,
	[8]byte{0xC0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x46}}

type IRpcOptionsInterface interface {
	IUnknownInterface
	Set(pPrx *IUnknown, dwProperty RPCOPT_PROPERTIES, dwValue uintptr) HRESULT
	Query(pPrx *IUnknown, dwProperty RPCOPT_PROPERTIES, pdwValue *uintptr) HRESULT
}

type IRpcOptionsVtbl struct {
	IUnknownVtbl
	Set   uintptr
	Query uintptr
}

type IRpcOptions struct {
	IUnknown
}

func (this *IRpcOptions) Vtbl() *IRpcOptionsVtbl {
	return (*IRpcOptionsVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IRpcOptions) Set(pPrx *IUnknown, dwProperty RPCOPT_PROPERTIES, dwValue uintptr) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Set, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pPrx)), uintptr(dwProperty), dwValue)
	return HRESULT(ret)
}

func (this *IRpcOptions) Query(pPrx *IUnknown, dwProperty RPCOPT_PROPERTIES, pdwValue *uintptr) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Query, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pPrx)), uintptr(dwProperty), uintptr(unsafe.Pointer(pdwValue)))
	return HRESULT(ret)
}

// 0000015B-0000-0000-C000-000000000046
var IID_IGlobalOptions = syscall.GUID{0x0000015B, 0x0000, 0x0000,
	[8]byte{0xC0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x46}}

type IGlobalOptionsInterface interface {
	IUnknownInterface
	Set(dwProperty GLOBALOPT_PROPERTIES, dwValue uintptr) HRESULT
	Query(dwProperty GLOBALOPT_PROPERTIES, pdwValue *uintptr) HRESULT
}

type IGlobalOptionsVtbl struct {
	IUnknownVtbl
	Set   uintptr
	Query uintptr
}

type IGlobalOptions struct {
	IUnknown
}

func (this *IGlobalOptions) Vtbl() *IGlobalOptionsVtbl {
	return (*IGlobalOptionsVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IGlobalOptions) Set(dwProperty GLOBALOPT_PROPERTIES, dwValue uintptr) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Set, uintptr(unsafe.Pointer(this)), uintptr(dwProperty), dwValue)
	return HRESULT(ret)
}

func (this *IGlobalOptions) Query(dwProperty GLOBALOPT_PROPERTIES, pdwValue *uintptr) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Query, uintptr(unsafe.Pointer(this)), uintptr(dwProperty), uintptr(unsafe.Pointer(pdwValue)))
	return HRESULT(ret)
}

// 00000022-0000-0000-C000-000000000046
var IID_ISurrogate = syscall.GUID{0x00000022, 0x0000, 0x0000,
	[8]byte{0xC0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x46}}

type ISurrogateInterface interface {
	IUnknownInterface
	LoadDllServer(Clsid *syscall.GUID) HRESULT
	FreeSurrogate() HRESULT
}

type ISurrogateVtbl struct {
	IUnknownVtbl
	LoadDllServer uintptr
	FreeSurrogate uintptr
}

type ISurrogate struct {
	IUnknown
}

func (this *ISurrogate) Vtbl() *ISurrogateVtbl {
	return (*ISurrogateVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *ISurrogate) LoadDllServer(Clsid *syscall.GUID) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().LoadDllServer, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(Clsid)))
	return HRESULT(ret)
}

func (this *ISurrogate) FreeSurrogate() HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().FreeSurrogate, uintptr(unsafe.Pointer(this)))
	return HRESULT(ret)
}

// 00000146-0000-0000-C000-000000000046
var IID_IGlobalInterfaceTable = syscall.GUID{0x00000146, 0x0000, 0x0000,
	[8]byte{0xC0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x46}}

type IGlobalInterfaceTableInterface interface {
	IUnknownInterface
	RegisterInterfaceInGlobal(pUnk *IUnknown, riid *syscall.GUID, pdwCookie *uint32) HRESULT
	RevokeInterfaceFromGlobal(dwCookie uint32) HRESULT
	GetInterfaceFromGlobal(dwCookie uint32, riid *syscall.GUID, ppv unsafe.Pointer) HRESULT
}

type IGlobalInterfaceTableVtbl struct {
	IUnknownVtbl
	RegisterInterfaceInGlobal uintptr
	RevokeInterfaceFromGlobal uintptr
	GetInterfaceFromGlobal    uintptr
}

type IGlobalInterfaceTable struct {
	IUnknown
}

func (this *IGlobalInterfaceTable) Vtbl() *IGlobalInterfaceTableVtbl {
	return (*IGlobalInterfaceTableVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IGlobalInterfaceTable) RegisterInterfaceInGlobal(pUnk *IUnknown, riid *syscall.GUID, pdwCookie *uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().RegisterInterfaceInGlobal, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pUnk)), uintptr(unsafe.Pointer(riid)), uintptr(unsafe.Pointer(pdwCookie)))
	return HRESULT(ret)
}

func (this *IGlobalInterfaceTable) RevokeInterfaceFromGlobal(dwCookie uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().RevokeInterfaceFromGlobal, uintptr(unsafe.Pointer(this)), uintptr(dwCookie))
	return HRESULT(ret)
}

func (this *IGlobalInterfaceTable) GetInterfaceFromGlobal(dwCookie uint32, riid *syscall.GUID, ppv unsafe.Pointer) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetInterfaceFromGlobal, uintptr(unsafe.Pointer(this)), uintptr(dwCookie), uintptr(unsafe.Pointer(riid)), uintptr(ppv))
	return HRESULT(ret)
}

// 00000030-0000-0000-C000-000000000046
var IID_ISynchronize = syscall.GUID{0x00000030, 0x0000, 0x0000,
	[8]byte{0xC0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x46}}

type ISynchronizeInterface interface {
	IUnknownInterface
	Wait(dwFlags uint32, dwMilliseconds uint32) HRESULT
	Signal() HRESULT
	Reset() HRESULT
}

type ISynchronizeVtbl struct {
	IUnknownVtbl
	Wait   uintptr
	Signal uintptr
	Reset  uintptr
}

type ISynchronize struct {
	IUnknown
}

func (this *ISynchronize) Vtbl() *ISynchronizeVtbl {
	return (*ISynchronizeVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *ISynchronize) Wait(dwFlags uint32, dwMilliseconds uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Wait, uintptr(unsafe.Pointer(this)), uintptr(dwFlags), uintptr(dwMilliseconds))
	return HRESULT(ret)
}

func (this *ISynchronize) Signal() HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Signal, uintptr(unsafe.Pointer(this)))
	return HRESULT(ret)
}

func (this *ISynchronize) Reset() HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Reset, uintptr(unsafe.Pointer(this)))
	return HRESULT(ret)
}

// 00000031-0000-0000-C000-000000000046
var IID_ISynchronizeHandle = syscall.GUID{0x00000031, 0x0000, 0x0000,
	[8]byte{0xC0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x46}}

type ISynchronizeHandleInterface interface {
	IUnknownInterface
	GetHandle(ph *HANDLE) HRESULT
}

type ISynchronizeHandleVtbl struct {
	IUnknownVtbl
	GetHandle uintptr
}

type ISynchronizeHandle struct {
	IUnknown
}

func (this *ISynchronizeHandle) Vtbl() *ISynchronizeHandleVtbl {
	return (*ISynchronizeHandleVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *ISynchronizeHandle) GetHandle(ph *HANDLE) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetHandle, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(ph)))
	return HRESULT(ret)
}

// 00000032-0000-0000-C000-000000000046
var IID_ISynchronizeEvent = syscall.GUID{0x00000032, 0x0000, 0x0000,
	[8]byte{0xC0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x46}}

type ISynchronizeEventInterface interface {
	ISynchronizeHandleInterface
	SetEventHandle(ph *HANDLE) HRESULT
}

type ISynchronizeEventVtbl struct {
	ISynchronizeHandleVtbl
	SetEventHandle uintptr
}

type ISynchronizeEvent struct {
	ISynchronizeHandle
}

func (this *ISynchronizeEvent) Vtbl() *ISynchronizeEventVtbl {
	return (*ISynchronizeEventVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *ISynchronizeEvent) SetEventHandle(ph *HANDLE) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetEventHandle, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(ph)))
	return HRESULT(ret)
}

// 00000033-0000-0000-C000-000000000046
var IID_ISynchronizeContainer = syscall.GUID{0x00000033, 0x0000, 0x0000,
	[8]byte{0xC0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x46}}

type ISynchronizeContainerInterface interface {
	IUnknownInterface
	AddSynchronize(pSync *ISynchronize) HRESULT
	WaitMultiple(dwFlags uint32, dwTimeOut uint32, ppSync **ISynchronize) HRESULT
}

type ISynchronizeContainerVtbl struct {
	IUnknownVtbl
	AddSynchronize uintptr
	WaitMultiple   uintptr
}

type ISynchronizeContainer struct {
	IUnknown
}

func (this *ISynchronizeContainer) Vtbl() *ISynchronizeContainerVtbl {
	return (*ISynchronizeContainerVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *ISynchronizeContainer) AddSynchronize(pSync *ISynchronize) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().AddSynchronize, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pSync)))
	return HRESULT(ret)
}

func (this *ISynchronizeContainer) WaitMultiple(dwFlags uint32, dwTimeOut uint32, ppSync **ISynchronize) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().WaitMultiple, uintptr(unsafe.Pointer(this)), uintptr(dwFlags), uintptr(dwTimeOut), uintptr(unsafe.Pointer(ppSync)))
	return HRESULT(ret)
}

// 00000025-0000-0000-C000-000000000046
var IID_ISynchronizeMutex = syscall.GUID{0x00000025, 0x0000, 0x0000,
	[8]byte{0xC0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x46}}

type ISynchronizeMutexInterface interface {
	ISynchronizeInterface
	ReleaseMutex() HRESULT
}

type ISynchronizeMutexVtbl struct {
	ISynchronizeVtbl
	ReleaseMutex uintptr
}

type ISynchronizeMutex struct {
	ISynchronize
}

func (this *ISynchronizeMutex) Vtbl() *ISynchronizeMutexVtbl {
	return (*ISynchronizeMutexVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *ISynchronizeMutex) ReleaseMutex() HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().ReleaseMutex, uintptr(unsafe.Pointer(this)))
	return HRESULT(ret)
}

// 00000029-0000-0000-C000-000000000046
var IID_ICancelMethodCalls = syscall.GUID{0x00000029, 0x0000, 0x0000,
	[8]byte{0xC0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x46}}

type ICancelMethodCallsInterface interface {
	IUnknownInterface
	Cancel(ulSeconds uint32) HRESULT
	TestCancel() HRESULT
}

type ICancelMethodCallsVtbl struct {
	IUnknownVtbl
	Cancel     uintptr
	TestCancel uintptr
}

type ICancelMethodCalls struct {
	IUnknown
}

func (this *ICancelMethodCalls) Vtbl() *ICancelMethodCallsVtbl {
	return (*ICancelMethodCallsVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *ICancelMethodCalls) Cancel(ulSeconds uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Cancel, uintptr(unsafe.Pointer(this)), uintptr(ulSeconds))
	return HRESULT(ret)
}

func (this *ICancelMethodCalls) TestCancel() HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().TestCancel, uintptr(unsafe.Pointer(this)))
	return HRESULT(ret)
}

// 0000002A-0000-0000-C000-000000000046
var IID_IAsyncManager = syscall.GUID{0x0000002A, 0x0000, 0x0000,
	[8]byte{0xC0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x46}}

type IAsyncManagerInterface interface {
	IUnknownInterface
	CompleteCall(Result HRESULT) HRESULT
	GetCallContext(riid *syscall.GUID, pInterface unsafe.Pointer) HRESULT
	GetState(pulStateFlags *uint32) HRESULT
}

type IAsyncManagerVtbl struct {
	IUnknownVtbl
	CompleteCall   uintptr
	GetCallContext uintptr
	GetState       uintptr
}

type IAsyncManager struct {
	IUnknown
}

func (this *IAsyncManager) Vtbl() *IAsyncManagerVtbl {
	return (*IAsyncManagerVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IAsyncManager) CompleteCall(Result HRESULT) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().CompleteCall, uintptr(unsafe.Pointer(this)), uintptr(Result))
	return HRESULT(ret)
}

func (this *IAsyncManager) GetCallContext(riid *syscall.GUID, pInterface unsafe.Pointer) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetCallContext, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(riid)), uintptr(pInterface))
	return HRESULT(ret)
}

func (this *IAsyncManager) GetState(pulStateFlags *uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetState, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pulStateFlags)))
	return HRESULT(ret)
}

// 1C733A30-2A1C-11CE-ADE5-00AA0044773D
var IID_ICallFactory = syscall.GUID{0x1C733A30, 0x2A1C, 0x11CE,
	[8]byte{0xAD, 0xE5, 0x00, 0xAA, 0x00, 0x44, 0x77, 0x3D}}

type ICallFactoryInterface interface {
	IUnknownInterface
	CreateCall(riid *syscall.GUID, pCtrlUnk *IUnknown, riid2 *syscall.GUID, ppv **IUnknown) HRESULT
}

type ICallFactoryVtbl struct {
	IUnknownVtbl
	CreateCall uintptr
}

type ICallFactory struct {
	IUnknown
}

func (this *ICallFactory) Vtbl() *ICallFactoryVtbl {
	return (*ICallFactoryVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *ICallFactory) CreateCall(riid *syscall.GUID, pCtrlUnk *IUnknown, riid2 *syscall.GUID, ppv **IUnknown) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().CreateCall, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(riid)), uintptr(unsafe.Pointer(pCtrlUnk)), uintptr(unsafe.Pointer(riid2)), uintptr(unsafe.Pointer(ppv)))
	return HRESULT(ret)
}

// 00000149-0000-0000-C000-000000000046
var IID_IRpcHelper = syscall.GUID{0x00000149, 0x0000, 0x0000,
	[8]byte{0xC0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x46}}

type IRpcHelperInterface interface {
	IUnknownInterface
	GetDCOMProtocolVersion(pComVersion *uint32) HRESULT
	GetIIDFromOBJREF(pObjRef unsafe.Pointer, piid **syscall.GUID) HRESULT
}

type IRpcHelperVtbl struct {
	IUnknownVtbl
	GetDCOMProtocolVersion uintptr
	GetIIDFromOBJREF       uintptr
}

type IRpcHelper struct {
	IUnknown
}

func (this *IRpcHelper) Vtbl() *IRpcHelperVtbl {
	return (*IRpcHelperVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IRpcHelper) GetDCOMProtocolVersion(pComVersion *uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetDCOMProtocolVersion, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pComVersion)))
	return HRESULT(ret)
}

func (this *IRpcHelper) GetIIDFromOBJREF(pObjRef unsafe.Pointer, piid **syscall.GUID) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetIIDFromOBJREF, uintptr(unsafe.Pointer(this)), uintptr(pObjRef), uintptr(unsafe.Pointer(piid)))
	return HRESULT(ret)
}

// EB0CB9E8-7996-11D2-872E-0000F8080859
var IID_IReleaseMarshalBuffers = syscall.GUID{0xEB0CB9E8, 0x7996, 0x11D2,
	[8]byte{0x87, 0x2E, 0x00, 0x00, 0xF8, 0x08, 0x08, 0x59}}

type IReleaseMarshalBuffersInterface interface {
	IUnknownInterface
	ReleaseMarshalBuffer(pMsg *RPCOLEMESSAGE, dwFlags uint32, pChnl *IUnknown) HRESULT
}

type IReleaseMarshalBuffersVtbl struct {
	IUnknownVtbl
	ReleaseMarshalBuffer uintptr
}

type IReleaseMarshalBuffers struct {
	IUnknown
}

func (this *IReleaseMarshalBuffers) Vtbl() *IReleaseMarshalBuffersVtbl {
	return (*IReleaseMarshalBuffersVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IReleaseMarshalBuffers) ReleaseMarshalBuffer(pMsg *RPCOLEMESSAGE, dwFlags uint32, pChnl *IUnknown) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().ReleaseMarshalBuffer, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pMsg)), uintptr(dwFlags), uintptr(unsafe.Pointer(pChnl)))
	return HRESULT(ret)
}

// 0000002B-0000-0000-C000-000000000046
var IID_IWaitMultiple = syscall.GUID{0x0000002B, 0x0000, 0x0000,
	[8]byte{0xC0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x46}}

type IWaitMultipleInterface interface {
	IUnknownInterface
	WaitMultiple(timeout uint32, pSync **ISynchronize) HRESULT
	AddSynchronize(pSync *ISynchronize) HRESULT
}

type IWaitMultipleVtbl struct {
	IUnknownVtbl
	WaitMultiple   uintptr
	AddSynchronize uintptr
}

type IWaitMultiple struct {
	IUnknown
}

func (this *IWaitMultiple) Vtbl() *IWaitMultipleVtbl {
	return (*IWaitMultipleVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IWaitMultiple) WaitMultiple(timeout uint32, pSync **ISynchronize) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().WaitMultiple, uintptr(unsafe.Pointer(this)), uintptr(timeout), uintptr(unsafe.Pointer(pSync)))
	return HRESULT(ret)
}

func (this *IWaitMultiple) AddSynchronize(pSync *ISynchronize) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().AddSynchronize, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pSync)))
	return HRESULT(ret)
}

// 00000147-0000-0000-C000-000000000046
var IID_IAddrTrackingControl = syscall.GUID{0x00000147, 0x0000, 0x0000,
	[8]byte{0xC0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x46}}

type IAddrTrackingControlInterface interface {
	IUnknownInterface
	EnableCOMDynamicAddrTracking() HRESULT
	DisableCOMDynamicAddrTracking() HRESULT
}

type IAddrTrackingControlVtbl struct {
	IUnknownVtbl
	EnableCOMDynamicAddrTracking  uintptr
	DisableCOMDynamicAddrTracking uintptr
}

type IAddrTrackingControl struct {
	IUnknown
}

func (this *IAddrTrackingControl) Vtbl() *IAddrTrackingControlVtbl {
	return (*IAddrTrackingControlVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IAddrTrackingControl) EnableCOMDynamicAddrTracking() HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().EnableCOMDynamicAddrTracking, uintptr(unsafe.Pointer(this)))
	return HRESULT(ret)
}

func (this *IAddrTrackingControl) DisableCOMDynamicAddrTracking() HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().DisableCOMDynamicAddrTracking, uintptr(unsafe.Pointer(this)))
	return HRESULT(ret)
}

// 00000148-0000-0000-C000-000000000046
var IID_IAddrExclusionControl = syscall.GUID{0x00000148, 0x0000, 0x0000,
	[8]byte{0xC0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x46}}

type IAddrExclusionControlInterface interface {
	IUnknownInterface
	GetCurrentAddrExclusionList(riid *syscall.GUID, ppEnumerator unsafe.Pointer) HRESULT
	UpdateAddrExclusionList(pEnumerator *IUnknown) HRESULT
}

type IAddrExclusionControlVtbl struct {
	IUnknownVtbl
	GetCurrentAddrExclusionList uintptr
	UpdateAddrExclusionList     uintptr
}

type IAddrExclusionControl struct {
	IUnknown
}

func (this *IAddrExclusionControl) Vtbl() *IAddrExclusionControlVtbl {
	return (*IAddrExclusionControlVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IAddrExclusionControl) GetCurrentAddrExclusionList(riid *syscall.GUID, ppEnumerator unsafe.Pointer) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetCurrentAddrExclusionList, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(riid)), uintptr(ppEnumerator))
	return HRESULT(ret)
}

func (this *IAddrExclusionControl) UpdateAddrExclusionList(pEnumerator *IUnknown) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().UpdateAddrExclusionList, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pEnumerator)))
	return HRESULT(ret)
}

// DB2F3ACA-2F86-11D1-8E04-00C04FB9989A
var IID_IPipeByte = syscall.GUID{0xDB2F3ACA, 0x2F86, 0x11D1,
	[8]byte{0x8E, 0x04, 0x00, 0xC0, 0x4F, 0xB9, 0x98, 0x9A}}

type IPipeByteInterface interface {
	IUnknownInterface
	Pull(buf *byte, cRequest uint32, pcReturned *uint32) HRESULT
	Push(buf *byte, cSent uint32) HRESULT
}

type IPipeByteVtbl struct {
	IUnknownVtbl
	Pull uintptr
	Push uintptr
}

type IPipeByte struct {
	IUnknown
}

func (this *IPipeByte) Vtbl() *IPipeByteVtbl {
	return (*IPipeByteVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IPipeByte) Pull(buf *byte, cRequest uint32, pcReturned *uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Pull, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(buf)), uintptr(cRequest), uintptr(unsafe.Pointer(pcReturned)))
	return HRESULT(ret)
}

func (this *IPipeByte) Push(buf *byte, cSent uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Push, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(buf)), uintptr(cSent))
	return HRESULT(ret)
}

// DB2F3ACB-2F86-11D1-8E04-00C04FB9989A
var IID_AsyncIPipeByte = syscall.GUID{0xDB2F3ACB, 0x2F86, 0x11D1,
	[8]byte{0x8E, 0x04, 0x00, 0xC0, 0x4F, 0xB9, 0x98, 0x9A}}

type AsyncIPipeByteInterface interface {
	IUnknownInterface
	Begin_Pull(cRequest uint32) HRESULT
	Finish_Pull(buf *byte, pcReturned *uint32) HRESULT
	Begin_Push(buf *byte, cSent uint32) HRESULT
	Finish_Push() HRESULT
}

type AsyncIPipeByteVtbl struct {
	IUnknownVtbl
	Begin_Pull  uintptr
	Finish_Pull uintptr
	Begin_Push  uintptr
	Finish_Push uintptr
}

type AsyncIPipeByte struct {
	IUnknown
}

func (this *AsyncIPipeByte) Vtbl() *AsyncIPipeByteVtbl {
	return (*AsyncIPipeByteVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *AsyncIPipeByte) Begin_Pull(cRequest uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Begin_Pull, uintptr(unsafe.Pointer(this)), uintptr(cRequest))
	return HRESULT(ret)
}

func (this *AsyncIPipeByte) Finish_Pull(buf *byte, pcReturned *uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Finish_Pull, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(buf)), uintptr(unsafe.Pointer(pcReturned)))
	return HRESULT(ret)
}

func (this *AsyncIPipeByte) Begin_Push(buf *byte, cSent uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Begin_Push, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(buf)), uintptr(cSent))
	return HRESULT(ret)
}

func (this *AsyncIPipeByte) Finish_Push() HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Finish_Push, uintptr(unsafe.Pointer(this)))
	return HRESULT(ret)
}

// DB2F3ACC-2F86-11D1-8E04-00C04FB9989A
var IID_IPipeLong = syscall.GUID{0xDB2F3ACC, 0x2F86, 0x11D1,
	[8]byte{0x8E, 0x04, 0x00, 0xC0, 0x4F, 0xB9, 0x98, 0x9A}}

type IPipeLongInterface interface {
	IUnknownInterface
	Pull(buf *int32, cRequest uint32, pcReturned *uint32) HRESULT
	Push(buf *int32, cSent uint32) HRESULT
}

type IPipeLongVtbl struct {
	IUnknownVtbl
	Pull uintptr
	Push uintptr
}

type IPipeLong struct {
	IUnknown
}

func (this *IPipeLong) Vtbl() *IPipeLongVtbl {
	return (*IPipeLongVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IPipeLong) Pull(buf *int32, cRequest uint32, pcReturned *uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Pull, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(buf)), uintptr(cRequest), uintptr(unsafe.Pointer(pcReturned)))
	return HRESULT(ret)
}

func (this *IPipeLong) Push(buf *int32, cSent uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Push, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(buf)), uintptr(cSent))
	return HRESULT(ret)
}

// DB2F3ACD-2F86-11D1-8E04-00C04FB9989A
var IID_AsyncIPipeLong = syscall.GUID{0xDB2F3ACD, 0x2F86, 0x11D1,
	[8]byte{0x8E, 0x04, 0x00, 0xC0, 0x4F, 0xB9, 0x98, 0x9A}}

type AsyncIPipeLongInterface interface {
	IUnknownInterface
	Begin_Pull(cRequest uint32) HRESULT
	Finish_Pull(buf *int32, pcReturned *uint32) HRESULT
	Begin_Push(buf *int32, cSent uint32) HRESULT
	Finish_Push() HRESULT
}

type AsyncIPipeLongVtbl struct {
	IUnknownVtbl
	Begin_Pull  uintptr
	Finish_Pull uintptr
	Begin_Push  uintptr
	Finish_Push uintptr
}

type AsyncIPipeLong struct {
	IUnknown
}

func (this *AsyncIPipeLong) Vtbl() *AsyncIPipeLongVtbl {
	return (*AsyncIPipeLongVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *AsyncIPipeLong) Begin_Pull(cRequest uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Begin_Pull, uintptr(unsafe.Pointer(this)), uintptr(cRequest))
	return HRESULT(ret)
}

func (this *AsyncIPipeLong) Finish_Pull(buf *int32, pcReturned *uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Finish_Pull, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(buf)), uintptr(unsafe.Pointer(pcReturned)))
	return HRESULT(ret)
}

func (this *AsyncIPipeLong) Begin_Push(buf *int32, cSent uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Begin_Push, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(buf)), uintptr(cSent))
	return HRESULT(ret)
}

func (this *AsyncIPipeLong) Finish_Push() HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Finish_Push, uintptr(unsafe.Pointer(this)))
	return HRESULT(ret)
}

// DB2F3ACE-2F86-11D1-8E04-00C04FB9989A
var IID_IPipeDouble = syscall.GUID{0xDB2F3ACE, 0x2F86, 0x11D1,
	[8]byte{0x8E, 0x04, 0x00, 0xC0, 0x4F, 0xB9, 0x98, 0x9A}}

type IPipeDoubleInterface interface {
	IUnknownInterface
	Pull(buf *float64, cRequest uint32, pcReturned *uint32) HRESULT
	Push(buf *float64, cSent uint32) HRESULT
}

type IPipeDoubleVtbl struct {
	IUnknownVtbl
	Pull uintptr
	Push uintptr
}

type IPipeDouble struct {
	IUnknown
}

func (this *IPipeDouble) Vtbl() *IPipeDoubleVtbl {
	return (*IPipeDoubleVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IPipeDouble) Pull(buf *float64, cRequest uint32, pcReturned *uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Pull, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(buf)), uintptr(cRequest), uintptr(unsafe.Pointer(pcReturned)))
	return HRESULT(ret)
}

func (this *IPipeDouble) Push(buf *float64, cSent uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Push, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(buf)), uintptr(cSent))
	return HRESULT(ret)
}

// DB2F3ACF-2F86-11D1-8E04-00C04FB9989A
var IID_AsyncIPipeDouble = syscall.GUID{0xDB2F3ACF, 0x2F86, 0x11D1,
	[8]byte{0x8E, 0x04, 0x00, 0xC0, 0x4F, 0xB9, 0x98, 0x9A}}

type AsyncIPipeDoubleInterface interface {
	IUnknownInterface
	Begin_Pull(cRequest uint32) HRESULT
	Finish_Pull(buf *float64, pcReturned *uint32) HRESULT
	Begin_Push(buf *float64, cSent uint32) HRESULT
	Finish_Push() HRESULT
}

type AsyncIPipeDoubleVtbl struct {
	IUnknownVtbl
	Begin_Pull  uintptr
	Finish_Pull uintptr
	Begin_Push  uintptr
	Finish_Push uintptr
}

type AsyncIPipeDouble struct {
	IUnknown
}

func (this *AsyncIPipeDouble) Vtbl() *AsyncIPipeDoubleVtbl {
	return (*AsyncIPipeDoubleVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *AsyncIPipeDouble) Begin_Pull(cRequest uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Begin_Pull, uintptr(unsafe.Pointer(this)), uintptr(cRequest))
	return HRESULT(ret)
}

func (this *AsyncIPipeDouble) Finish_Pull(buf *float64, pcReturned *uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Finish_Pull, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(buf)), uintptr(unsafe.Pointer(pcReturned)))
	return HRESULT(ret)
}

func (this *AsyncIPipeDouble) Begin_Push(buf *float64, cSent uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Begin_Push, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(buf)), uintptr(cSent))
	return HRESULT(ret)
}

func (this *AsyncIPipeDouble) Finish_Push() HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Finish_Push, uintptr(unsafe.Pointer(this)))
	return HRESULT(ret)
}

// 000001CE-0000-0000-C000-000000000046
var IID_IComThreadingInfo = syscall.GUID{0x000001CE, 0x0000, 0x0000,
	[8]byte{0xC0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x46}}

type IComThreadingInfoInterface interface {
	IUnknownInterface
	GetCurrentApartmentType(pAptType *APTTYPE) HRESULT
	GetCurrentThreadType(pThreadType *THDTYPE) HRESULT
	GetCurrentLogicalThreadId(pguidLogicalThreadId *syscall.GUID) HRESULT
	SetCurrentLogicalThreadId(rguid *syscall.GUID) HRESULT
}

type IComThreadingInfoVtbl struct {
	IUnknownVtbl
	GetCurrentApartmentType   uintptr
	GetCurrentThreadType      uintptr
	GetCurrentLogicalThreadId uintptr
	SetCurrentLogicalThreadId uintptr
}

type IComThreadingInfo struct {
	IUnknown
}

func (this *IComThreadingInfo) Vtbl() *IComThreadingInfoVtbl {
	return (*IComThreadingInfoVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IComThreadingInfo) GetCurrentApartmentType(pAptType *APTTYPE) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetCurrentApartmentType, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pAptType)))
	return HRESULT(ret)
}

func (this *IComThreadingInfo) GetCurrentThreadType(pThreadType *THDTYPE) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetCurrentThreadType, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pThreadType)))
	return HRESULT(ret)
}

func (this *IComThreadingInfo) GetCurrentLogicalThreadId(pguidLogicalThreadId *syscall.GUID) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetCurrentLogicalThreadId, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pguidLogicalThreadId)))
	return HRESULT(ret)
}

func (this *IComThreadingInfo) SetCurrentLogicalThreadId(rguid *syscall.GUID) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetCurrentLogicalThreadId, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(rguid)))
	return HRESULT(ret)
}

// 72380D55-8D2B-43A3-8513-2B6EF31434E9
var IID_IProcessInitControl = syscall.GUID{0x72380D55, 0x8D2B, 0x43A3,
	[8]byte{0x85, 0x13, 0x2B, 0x6E, 0xF3, 0x14, 0x34, 0xE9}}

type IProcessInitControlInterface interface {
	IUnknownInterface
	ResetInitializerTimeout(dwSecondsRemaining uint32) HRESULT
}

type IProcessInitControlVtbl struct {
	IUnknownVtbl
	ResetInitializerTimeout uintptr
}

type IProcessInitControl struct {
	IUnknown
}

func (this *IProcessInitControl) Vtbl() *IProcessInitControlVtbl {
	return (*IProcessInitControlVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IProcessInitControl) ResetInitializerTimeout(dwSecondsRemaining uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().ResetInitializerTimeout, uintptr(unsafe.Pointer(this)), uintptr(dwSecondsRemaining))
	return HRESULT(ret)
}

// 00000040-0000-0000-C000-000000000046
var IID_IFastRundown = syscall.GUID{0x00000040, 0x0000, 0x0000,
	[8]byte{0xC0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x46}}

type IFastRundownInterface interface {
	IUnknownInterface
}

type IFastRundownVtbl struct {
	IUnknownVtbl
}

type IFastRundown struct {
	IUnknown
}

func (this *IFastRundown) Vtbl() *IFastRundownVtbl {
	return (*IFastRundownVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

// 26D709AC-F70B-4421-A96F-D2878FAFB00D
var IID_IMachineGlobalObjectTable = syscall.GUID{0x26D709AC, 0xF70B, 0x4421,
	[8]byte{0xA9, 0x6F, 0xD2, 0x87, 0x8F, 0xAF, 0xB0, 0x0D}}

type IMachineGlobalObjectTableInterface interface {
	IUnknownInterface
	RegisterObject(clsid *syscall.GUID, identifier PWSTR, object *IUnknown, token **MachineGlobalObjectTableRegistrationToken__) HRESULT
	GetObject(clsid *syscall.GUID, identifier PWSTR, riid *syscall.GUID, ppv unsafe.Pointer) HRESULT
	RevokeObject(token *MachineGlobalObjectTableRegistrationToken__) HRESULT
}

type IMachineGlobalObjectTableVtbl struct {
	IUnknownVtbl
	RegisterObject uintptr
	GetObject      uintptr
	RevokeObject   uintptr
}

type IMachineGlobalObjectTable struct {
	IUnknown
}

func (this *IMachineGlobalObjectTable) Vtbl() *IMachineGlobalObjectTableVtbl {
	return (*IMachineGlobalObjectTableVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IMachineGlobalObjectTable) RegisterObject(clsid *syscall.GUID, identifier PWSTR, object *IUnknown, token **MachineGlobalObjectTableRegistrationToken__) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().RegisterObject, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(clsid)), uintptr(unsafe.Pointer(identifier)), uintptr(unsafe.Pointer(object)), uintptr(unsafe.Pointer(token)))
	return HRESULT(ret)
}

func (this *IMachineGlobalObjectTable) GetObject(clsid *syscall.GUID, identifier PWSTR, riid *syscall.GUID, ppv unsafe.Pointer) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetObject, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(clsid)), uintptr(unsafe.Pointer(identifier)), uintptr(unsafe.Pointer(riid)), uintptr(ppv))
	return HRESULT(ret)
}

func (this *IMachineGlobalObjectTable) RevokeObject(token *MachineGlobalObjectTableRegistrationToken__) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().RevokeObject, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(token)))
	return HRESULT(ret)
}

// 0000001D-0000-0000-C000-000000000046
var IID_IMallocSpy = syscall.GUID{0x0000001D, 0x0000, 0x0000,
	[8]byte{0xC0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x46}}

type IMallocSpyInterface interface {
	IUnknownInterface
	PreAlloc(cbRequest uintptr) uintptr
	PostAlloc(pActual unsafe.Pointer) unsafe.Pointer
	PreFree(pRequest unsafe.Pointer, fSpyed BOOL) unsafe.Pointer
	PostFree(fSpyed BOOL)
	PreRealloc(pRequest unsafe.Pointer, cbRequest uintptr, ppNewRequest unsafe.Pointer, fSpyed BOOL) uintptr
	PostRealloc(pActual unsafe.Pointer, fSpyed BOOL) unsafe.Pointer
	PreGetSize(pRequest unsafe.Pointer, fSpyed BOOL) unsafe.Pointer
	PostGetSize(cbActual uintptr, fSpyed BOOL) uintptr
	PreDidAlloc(pRequest unsafe.Pointer, fSpyed BOOL) unsafe.Pointer
	PostDidAlloc(pRequest unsafe.Pointer, fSpyed BOOL, fActual int32) int32
	PreHeapMinimize()
	PostHeapMinimize()
}

type IMallocSpyVtbl struct {
	IUnknownVtbl
	PreAlloc         uintptr
	PostAlloc        uintptr
	PreFree          uintptr
	PostFree         uintptr
	PreRealloc       uintptr
	PostRealloc      uintptr
	PreGetSize       uintptr
	PostGetSize      uintptr
	PreDidAlloc      uintptr
	PostDidAlloc     uintptr
	PreHeapMinimize  uintptr
	PostHeapMinimize uintptr
}

type IMallocSpy struct {
	IUnknown
}

func (this *IMallocSpy) Vtbl() *IMallocSpyVtbl {
	return (*IMallocSpyVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IMallocSpy) PreAlloc(cbRequest uintptr) uintptr {
	ret, _, _ := syscall.SyscallN(this.Vtbl().PreAlloc, uintptr(unsafe.Pointer(this)), cbRequest)
	return ret
}

func (this *IMallocSpy) PostAlloc(pActual unsafe.Pointer) unsafe.Pointer {
	ret, _, _ := syscall.SyscallN(this.Vtbl().PostAlloc, uintptr(unsafe.Pointer(this)), uintptr(pActual))
	return (unsafe.Pointer)(ret)
}

func (this *IMallocSpy) PreFree(pRequest unsafe.Pointer, fSpyed BOOL) unsafe.Pointer {
	ret, _, _ := syscall.SyscallN(this.Vtbl().PreFree, uintptr(unsafe.Pointer(this)), uintptr(pRequest), uintptr(fSpyed))
	return (unsafe.Pointer)(ret)
}

func (this *IMallocSpy) PostFree(fSpyed BOOL) {
	_, _, _ = syscall.SyscallN(this.Vtbl().PostFree, uintptr(unsafe.Pointer(this)), uintptr(fSpyed))
}

func (this *IMallocSpy) PreRealloc(pRequest unsafe.Pointer, cbRequest uintptr, ppNewRequest unsafe.Pointer, fSpyed BOOL) uintptr {
	ret, _, _ := syscall.SyscallN(this.Vtbl().PreRealloc, uintptr(unsafe.Pointer(this)), uintptr(pRequest), cbRequest, uintptr(ppNewRequest), uintptr(fSpyed))
	return ret
}

func (this *IMallocSpy) PostRealloc(pActual unsafe.Pointer, fSpyed BOOL) unsafe.Pointer {
	ret, _, _ := syscall.SyscallN(this.Vtbl().PostRealloc, uintptr(unsafe.Pointer(this)), uintptr(pActual), uintptr(fSpyed))
	return (unsafe.Pointer)(ret)
}

func (this *IMallocSpy) PreGetSize(pRequest unsafe.Pointer, fSpyed BOOL) unsafe.Pointer {
	ret, _, _ := syscall.SyscallN(this.Vtbl().PreGetSize, uintptr(unsafe.Pointer(this)), uintptr(pRequest), uintptr(fSpyed))
	return (unsafe.Pointer)(ret)
}

func (this *IMallocSpy) PostGetSize(cbActual uintptr, fSpyed BOOL) uintptr {
	ret, _, _ := syscall.SyscallN(this.Vtbl().PostGetSize, uintptr(unsafe.Pointer(this)), cbActual, uintptr(fSpyed))
	return ret
}

func (this *IMallocSpy) PreDidAlloc(pRequest unsafe.Pointer, fSpyed BOOL) unsafe.Pointer {
	ret, _, _ := syscall.SyscallN(this.Vtbl().PreDidAlloc, uintptr(unsafe.Pointer(this)), uintptr(pRequest), uintptr(fSpyed))
	return (unsafe.Pointer)(ret)
}

func (this *IMallocSpy) PostDidAlloc(pRequest unsafe.Pointer, fSpyed BOOL, fActual int32) int32 {
	ret, _, _ := syscall.SyscallN(this.Vtbl().PostDidAlloc, uintptr(unsafe.Pointer(this)), uintptr(pRequest), uintptr(fSpyed), uintptr(fActual))
	return int32(ret)
}

func (this *IMallocSpy) PreHeapMinimize() {
	_, _, _ = syscall.SyscallN(this.Vtbl().PreHeapMinimize, uintptr(unsafe.Pointer(this)))
}

func (this *IMallocSpy) PostHeapMinimize() {
	_, _, _ = syscall.SyscallN(this.Vtbl().PostHeapMinimize, uintptr(unsafe.Pointer(this)))
}

// 0000000E-0000-0000-C000-000000000046
var IID_IBindCtx = syscall.GUID{0x0000000E, 0x0000, 0x0000,
	[8]byte{0xC0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x46}}

type IBindCtxInterface interface {
	IUnknownInterface
	RegisterObjectBound(punk *IUnknown) HRESULT
	RevokeObjectBound(punk *IUnknown) HRESULT
	ReleaseBoundObjects() HRESULT
	SetBindOptions(pbindopts *BIND_OPTS) HRESULT
	GetBindOptions(pbindopts *BIND_OPTS) HRESULT
	GetRunningObjectTable(pprot **IRunningObjectTable) HRESULT
	RegisterObjectParam(pszKey PWSTR, punk *IUnknown) HRESULT
	GetObjectParam(pszKey PWSTR, ppunk **IUnknown) HRESULT
	EnumObjectParam(ppenum **IEnumString) HRESULT
	RevokeObjectParam(pszKey PWSTR) HRESULT
}

type IBindCtxVtbl struct {
	IUnknownVtbl
	RegisterObjectBound   uintptr
	RevokeObjectBound     uintptr
	ReleaseBoundObjects   uintptr
	SetBindOptions        uintptr
	GetBindOptions        uintptr
	GetRunningObjectTable uintptr
	RegisterObjectParam   uintptr
	GetObjectParam        uintptr
	EnumObjectParam       uintptr
	RevokeObjectParam     uintptr
}

type IBindCtx struct {
	IUnknown
}

func (this *IBindCtx) Vtbl() *IBindCtxVtbl {
	return (*IBindCtxVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IBindCtx) RegisterObjectBound(punk *IUnknown) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().RegisterObjectBound, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(punk)))
	return HRESULT(ret)
}

func (this *IBindCtx) RevokeObjectBound(punk *IUnknown) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().RevokeObjectBound, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(punk)))
	return HRESULT(ret)
}

func (this *IBindCtx) ReleaseBoundObjects() HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().ReleaseBoundObjects, uintptr(unsafe.Pointer(this)))
	return HRESULT(ret)
}

func (this *IBindCtx) SetBindOptions(pbindopts *BIND_OPTS) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetBindOptions, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pbindopts)))
	return HRESULT(ret)
}

func (this *IBindCtx) GetBindOptions(pbindopts *BIND_OPTS) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetBindOptions, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pbindopts)))
	return HRESULT(ret)
}

func (this *IBindCtx) GetRunningObjectTable(pprot **IRunningObjectTable) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetRunningObjectTable, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pprot)))
	return HRESULT(ret)
}

func (this *IBindCtx) RegisterObjectParam(pszKey PWSTR, punk *IUnknown) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().RegisterObjectParam, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pszKey)), uintptr(unsafe.Pointer(punk)))
	return HRESULT(ret)
}

func (this *IBindCtx) GetObjectParam(pszKey PWSTR, ppunk **IUnknown) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetObjectParam, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pszKey)), uintptr(unsafe.Pointer(ppunk)))
	return HRESULT(ret)
}

func (this *IBindCtx) EnumObjectParam(ppenum **IEnumString) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().EnumObjectParam, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(ppenum)))
	return HRESULT(ret)
}

func (this *IBindCtx) RevokeObjectParam(pszKey PWSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().RevokeObjectParam, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pszKey)))
	return HRESULT(ret)
}

// 00000102-0000-0000-C000-000000000046
var IID_IEnumMoniker = syscall.GUID{0x00000102, 0x0000, 0x0000,
	[8]byte{0xC0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x46}}

type IEnumMonikerInterface interface {
	IUnknownInterface
	Next(celt uint32, rgelt **IMoniker, pceltFetched *uint32) HRESULT
	Skip(celt uint32) HRESULT
	Reset() HRESULT
	Clone(ppenum **IEnumMoniker) HRESULT
}

type IEnumMonikerVtbl struct {
	IUnknownVtbl
	Next  uintptr
	Skip  uintptr
	Reset uintptr
	Clone uintptr
}

type IEnumMoniker struct {
	IUnknown
}

func (this *IEnumMoniker) Vtbl() *IEnumMonikerVtbl {
	return (*IEnumMonikerVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IEnumMoniker) Next(celt uint32, rgelt **IMoniker, pceltFetched *uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Next, uintptr(unsafe.Pointer(this)), uintptr(celt), uintptr(unsafe.Pointer(rgelt)), uintptr(unsafe.Pointer(pceltFetched)))
	return HRESULT(ret)
}

func (this *IEnumMoniker) Skip(celt uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Skip, uintptr(unsafe.Pointer(this)), uintptr(celt))
	return HRESULT(ret)
}

func (this *IEnumMoniker) Reset() HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Reset, uintptr(unsafe.Pointer(this)))
	return HRESULT(ret)
}

func (this *IEnumMoniker) Clone(ppenum **IEnumMoniker) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Clone, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(ppenum)))
	return HRESULT(ret)
}

// 00000126-0000-0000-C000-000000000046
var IID_IRunnableObject = syscall.GUID{0x00000126, 0x0000, 0x0000,
	[8]byte{0xC0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x46}}

type IRunnableObjectInterface interface {
	IUnknownInterface
	GetRunningClass(lpClsid *syscall.GUID) HRESULT
	Run(pbc *IBindCtx) HRESULT
	IsRunning() BOOL
	LockRunning(fLock BOOL, fLastUnlockCloses BOOL) HRESULT
	SetContainedObject(fContained BOOL) HRESULT
}

type IRunnableObjectVtbl struct {
	IUnknownVtbl
	GetRunningClass    uintptr
	Run                uintptr
	IsRunning          uintptr
	LockRunning        uintptr
	SetContainedObject uintptr
}

type IRunnableObject struct {
	IUnknown
}

func (this *IRunnableObject) Vtbl() *IRunnableObjectVtbl {
	return (*IRunnableObjectVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IRunnableObject) GetRunningClass(lpClsid *syscall.GUID) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetRunningClass, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(lpClsid)))
	return HRESULT(ret)
}

func (this *IRunnableObject) Run(pbc *IBindCtx) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Run, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pbc)))
	return HRESULT(ret)
}

func (this *IRunnableObject) IsRunning() BOOL {
	ret, _, _ := syscall.SyscallN(this.Vtbl().IsRunning, uintptr(unsafe.Pointer(this)))
	return BOOL(ret)
}

func (this *IRunnableObject) LockRunning(fLock BOOL, fLastUnlockCloses BOOL) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().LockRunning, uintptr(unsafe.Pointer(this)), uintptr(fLock), uintptr(fLastUnlockCloses))
	return HRESULT(ret)
}

func (this *IRunnableObject) SetContainedObject(fContained BOOL) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetContainedObject, uintptr(unsafe.Pointer(this)), uintptr(fContained))
	return HRESULT(ret)
}

// 00000010-0000-0000-C000-000000000046
var IID_IRunningObjectTable = syscall.GUID{0x00000010, 0x0000, 0x0000,
	[8]byte{0xC0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x46}}

type IRunningObjectTableInterface interface {
	IUnknownInterface
	Register(grfFlags uint32, punkObject *IUnknown, pmkObjectName *IMoniker, pdwRegister *uint32) HRESULT
	Revoke(dwRegister uint32) HRESULT
	IsRunning(pmkObjectName *IMoniker) HRESULT
	GetObject(pmkObjectName *IMoniker, ppunkObject **IUnknown) HRESULT
	NoteChangeTime(dwRegister uint32, pfiletime *FILETIME) HRESULT
	GetTimeOfLastChange(pmkObjectName *IMoniker, pfiletime *FILETIME) HRESULT
	EnumRunning(ppenumMoniker **IEnumMoniker) HRESULT
}

type IRunningObjectTableVtbl struct {
	IUnknownVtbl
	Register            uintptr
	Revoke              uintptr
	IsRunning           uintptr
	GetObject           uintptr
	NoteChangeTime      uintptr
	GetTimeOfLastChange uintptr
	EnumRunning         uintptr
}

type IRunningObjectTable struct {
	IUnknown
}

func (this *IRunningObjectTable) Vtbl() *IRunningObjectTableVtbl {
	return (*IRunningObjectTableVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IRunningObjectTable) Register(grfFlags uint32, punkObject *IUnknown, pmkObjectName *IMoniker, pdwRegister *uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Register, uintptr(unsafe.Pointer(this)), uintptr(grfFlags), uintptr(unsafe.Pointer(punkObject)), uintptr(unsafe.Pointer(pmkObjectName)), uintptr(unsafe.Pointer(pdwRegister)))
	return HRESULT(ret)
}

func (this *IRunningObjectTable) Revoke(dwRegister uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Revoke, uintptr(unsafe.Pointer(this)), uintptr(dwRegister))
	return HRESULT(ret)
}

func (this *IRunningObjectTable) IsRunning(pmkObjectName *IMoniker) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().IsRunning, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pmkObjectName)))
	return HRESULT(ret)
}

func (this *IRunningObjectTable) GetObject(pmkObjectName *IMoniker, ppunkObject **IUnknown) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetObject, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pmkObjectName)), uintptr(unsafe.Pointer(ppunkObject)))
	return HRESULT(ret)
}

func (this *IRunningObjectTable) NoteChangeTime(dwRegister uint32, pfiletime *FILETIME) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().NoteChangeTime, uintptr(unsafe.Pointer(this)), uintptr(dwRegister), uintptr(unsafe.Pointer(pfiletime)))
	return HRESULT(ret)
}

func (this *IRunningObjectTable) GetTimeOfLastChange(pmkObjectName *IMoniker, pfiletime *FILETIME) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetTimeOfLastChange, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pmkObjectName)), uintptr(unsafe.Pointer(pfiletime)))
	return HRESULT(ret)
}

func (this *IRunningObjectTable) EnumRunning(ppenumMoniker **IEnumMoniker) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().EnumRunning, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(ppenumMoniker)))
	return HRESULT(ret)
}

// 0000010C-0000-0000-C000-000000000046
var IID_IPersist = syscall.GUID{0x0000010C, 0x0000, 0x0000,
	[8]byte{0xC0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x46}}

type IPersistInterface interface {
	IUnknownInterface
	GetClassID(pClassID *syscall.GUID) HRESULT
}

type IPersistVtbl struct {
	IUnknownVtbl
	GetClassID uintptr
}

type IPersist struct {
	IUnknown
}

func (this *IPersist) Vtbl() *IPersistVtbl {
	return (*IPersistVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IPersist) GetClassID(pClassID *syscall.GUID) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetClassID, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pClassID)))
	return HRESULT(ret)
}

// 00000109-0000-0000-C000-000000000046
var IID_IPersistStream = syscall.GUID{0x00000109, 0x0000, 0x0000,
	[8]byte{0xC0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x46}}

type IPersistStreamInterface interface {
	IPersistInterface
	IsDirty() HRESULT
	Load(pStm *IStream) HRESULT
	Save(pStm *IStream, fClearDirty BOOL) HRESULT
	GetSizeMax(pcbSize *uint64) HRESULT
}

type IPersistStreamVtbl struct {
	IPersistVtbl
	IsDirty    uintptr
	Load       uintptr
	Save       uintptr
	GetSizeMax uintptr
}

type IPersistStream struct {
	IPersist
}

func (this *IPersistStream) Vtbl() *IPersistStreamVtbl {
	return (*IPersistStreamVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IPersistStream) IsDirty() HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().IsDirty, uintptr(unsafe.Pointer(this)))
	return HRESULT(ret)
}

func (this *IPersistStream) Load(pStm *IStream) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Load, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pStm)))
	return HRESULT(ret)
}

func (this *IPersistStream) Save(pStm *IStream, fClearDirty BOOL) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Save, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pStm)), uintptr(fClearDirty))
	return HRESULT(ret)
}

func (this *IPersistStream) GetSizeMax(pcbSize *uint64) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetSizeMax, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pcbSize)))
	return HRESULT(ret)
}

// 0000000F-0000-0000-C000-000000000046
var IID_IMoniker = syscall.GUID{0x0000000F, 0x0000, 0x0000,
	[8]byte{0xC0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x46}}

type IMonikerInterface interface {
	IPersistStreamInterface
	BindToObject(pbc *IBindCtx, pmkToLeft *IMoniker, riidResult *syscall.GUID, ppvResult unsafe.Pointer) HRESULT
	BindToStorage(pbc *IBindCtx, pmkToLeft *IMoniker, riid *syscall.GUID, ppvObj unsafe.Pointer) HRESULT
	Reduce(pbc *IBindCtx, dwReduceHowFar uint32, ppmkToLeft **IMoniker, ppmkReduced **IMoniker) HRESULT
	ComposeWith(pmkRight *IMoniker, fOnlyIfNotGeneric BOOL, ppmkComposite **IMoniker) HRESULT
	Enum(fForward BOOL, ppenumMoniker **IEnumMoniker) HRESULT
	IsEqual(pmkOtherMoniker *IMoniker) HRESULT
	Hash(pdwHash *uint32) HRESULT
	IsRunning(pbc *IBindCtx, pmkToLeft *IMoniker, pmkNewlyRunning *IMoniker) HRESULT
	GetTimeOfLastChange(pbc *IBindCtx, pmkToLeft *IMoniker, pFileTime *FILETIME) HRESULT
	Inverse(ppmk **IMoniker) HRESULT
	CommonPrefixWith(pmkOther *IMoniker, ppmkPrefix **IMoniker) HRESULT
	RelativePathTo(pmkOther *IMoniker, ppmkRelPath **IMoniker) HRESULT
	GetDisplayName(pbc *IBindCtx, pmkToLeft *IMoniker, ppszDisplayName *PWSTR) HRESULT
	ParseDisplayName(pbc *IBindCtx, pmkToLeft *IMoniker, pszDisplayName PWSTR, pchEaten *uint32, ppmkOut **IMoniker) HRESULT
	IsSystemMoniker(pdwMksys *uint32) HRESULT
}

type IMonikerVtbl struct {
	IPersistStreamVtbl
	BindToObject        uintptr
	BindToStorage       uintptr
	Reduce              uintptr
	ComposeWith         uintptr
	Enum                uintptr
	IsEqual             uintptr
	Hash                uintptr
	IsRunning           uintptr
	GetTimeOfLastChange uintptr
	Inverse             uintptr
	CommonPrefixWith    uintptr
	RelativePathTo      uintptr
	GetDisplayName      uintptr
	ParseDisplayName    uintptr
	IsSystemMoniker     uintptr
}

type IMoniker struct {
	IPersistStream
}

func (this *IMoniker) Vtbl() *IMonikerVtbl {
	return (*IMonikerVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IMoniker) BindToObject(pbc *IBindCtx, pmkToLeft *IMoniker, riidResult *syscall.GUID, ppvResult unsafe.Pointer) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().BindToObject, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pbc)), uintptr(unsafe.Pointer(pmkToLeft)), uintptr(unsafe.Pointer(riidResult)), uintptr(ppvResult))
	return HRESULT(ret)
}

func (this *IMoniker) BindToStorage(pbc *IBindCtx, pmkToLeft *IMoniker, riid *syscall.GUID, ppvObj unsafe.Pointer) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().BindToStorage, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pbc)), uintptr(unsafe.Pointer(pmkToLeft)), uintptr(unsafe.Pointer(riid)), uintptr(ppvObj))
	return HRESULT(ret)
}

func (this *IMoniker) Reduce(pbc *IBindCtx, dwReduceHowFar uint32, ppmkToLeft **IMoniker, ppmkReduced **IMoniker) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Reduce, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pbc)), uintptr(dwReduceHowFar), uintptr(unsafe.Pointer(ppmkToLeft)), uintptr(unsafe.Pointer(ppmkReduced)))
	return HRESULT(ret)
}

func (this *IMoniker) ComposeWith(pmkRight *IMoniker, fOnlyIfNotGeneric BOOL, ppmkComposite **IMoniker) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().ComposeWith, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pmkRight)), uintptr(fOnlyIfNotGeneric), uintptr(unsafe.Pointer(ppmkComposite)))
	return HRESULT(ret)
}

func (this *IMoniker) Enum(fForward BOOL, ppenumMoniker **IEnumMoniker) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Enum, uintptr(unsafe.Pointer(this)), uintptr(fForward), uintptr(unsafe.Pointer(ppenumMoniker)))
	return HRESULT(ret)
}

func (this *IMoniker) IsEqual(pmkOtherMoniker *IMoniker) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().IsEqual, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pmkOtherMoniker)))
	return HRESULT(ret)
}

func (this *IMoniker) Hash(pdwHash *uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Hash, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pdwHash)))
	return HRESULT(ret)
}

func (this *IMoniker) IsRunning(pbc *IBindCtx, pmkToLeft *IMoniker, pmkNewlyRunning *IMoniker) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().IsRunning, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pbc)), uintptr(unsafe.Pointer(pmkToLeft)), uintptr(unsafe.Pointer(pmkNewlyRunning)))
	return HRESULT(ret)
}

func (this *IMoniker) GetTimeOfLastChange(pbc *IBindCtx, pmkToLeft *IMoniker, pFileTime *FILETIME) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetTimeOfLastChange, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pbc)), uintptr(unsafe.Pointer(pmkToLeft)), uintptr(unsafe.Pointer(pFileTime)))
	return HRESULT(ret)
}

func (this *IMoniker) Inverse(ppmk **IMoniker) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Inverse, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(ppmk)))
	return HRESULT(ret)
}

func (this *IMoniker) CommonPrefixWith(pmkOther *IMoniker, ppmkPrefix **IMoniker) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().CommonPrefixWith, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pmkOther)), uintptr(unsafe.Pointer(ppmkPrefix)))
	return HRESULT(ret)
}

func (this *IMoniker) RelativePathTo(pmkOther *IMoniker, ppmkRelPath **IMoniker) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().RelativePathTo, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pmkOther)), uintptr(unsafe.Pointer(ppmkRelPath)))
	return HRESULT(ret)
}

func (this *IMoniker) GetDisplayName(pbc *IBindCtx, pmkToLeft *IMoniker, ppszDisplayName *PWSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetDisplayName, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pbc)), uintptr(unsafe.Pointer(pmkToLeft)), uintptr(unsafe.Pointer(ppszDisplayName)))
	return HRESULT(ret)
}

func (this *IMoniker) ParseDisplayName(pbc *IBindCtx, pmkToLeft *IMoniker, pszDisplayName PWSTR, pchEaten *uint32, ppmkOut **IMoniker) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().ParseDisplayName, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pbc)), uintptr(unsafe.Pointer(pmkToLeft)), uintptr(unsafe.Pointer(pszDisplayName)), uintptr(unsafe.Pointer(pchEaten)), uintptr(unsafe.Pointer(ppmkOut)))
	return HRESULT(ret)
}

func (this *IMoniker) IsSystemMoniker(pdwMksys *uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().IsSystemMoniker, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pdwMksys)))
	return HRESULT(ret)
}

// F29F6BC0-5021-11CE-AA15-00006901293F
var IID_IROTData = syscall.GUID{0xF29F6BC0, 0x5021, 0x11CE,
	[8]byte{0xAA, 0x15, 0x00, 0x00, 0x69, 0x01, 0x29, 0x3F}}

type IROTDataInterface interface {
	IUnknownInterface
	GetComparisonData(pbData *byte, cbMax uint32, pcbData *uint32) HRESULT
}

type IROTDataVtbl struct {
	IUnknownVtbl
	GetComparisonData uintptr
}

type IROTData struct {
	IUnknown
}

func (this *IROTData) Vtbl() *IROTDataVtbl {
	return (*IROTDataVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IROTData) GetComparisonData(pbData *byte, cbMax uint32, pcbData *uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetComparisonData, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pbData)), uintptr(cbMax), uintptr(unsafe.Pointer(pcbData)))
	return HRESULT(ret)
}

// 0000010B-0000-0000-C000-000000000046
var IID_IPersistFile = syscall.GUID{0x0000010B, 0x0000, 0x0000,
	[8]byte{0xC0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x46}}

type IPersistFileInterface interface {
	IPersistInterface
	IsDirty() HRESULT
	Load(pszFileName PWSTR, dwMode uint32) HRESULT
	Save(pszFileName PWSTR, fRemember BOOL) HRESULT
	SaveCompleted(pszFileName PWSTR) HRESULT
	GetCurFile(ppszFileName *PWSTR) HRESULT
}

type IPersistFileVtbl struct {
	IPersistVtbl
	IsDirty       uintptr
	Load          uintptr
	Save          uintptr
	SaveCompleted uintptr
	GetCurFile    uintptr
}

type IPersistFile struct {
	IPersist
}

func (this *IPersistFile) Vtbl() *IPersistFileVtbl {
	return (*IPersistFileVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IPersistFile) IsDirty() HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().IsDirty, uintptr(unsafe.Pointer(this)))
	return HRESULT(ret)
}

func (this *IPersistFile) Load(pszFileName PWSTR, dwMode uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Load, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pszFileName)), uintptr(dwMode))
	return HRESULT(ret)
}

func (this *IPersistFile) Save(pszFileName PWSTR, fRemember BOOL) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Save, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pszFileName)), uintptr(fRemember))
	return HRESULT(ret)
}

func (this *IPersistFile) SaveCompleted(pszFileName PWSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SaveCompleted, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pszFileName)))
	return HRESULT(ret)
}

func (this *IPersistFile) GetCurFile(ppszFileName *PWSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetCurFile, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(ppszFileName)))
	return HRESULT(ret)
}

// 00000103-0000-0000-C000-000000000046
var IID_IEnumFORMATETC = syscall.GUID{0x00000103, 0x0000, 0x0000,
	[8]byte{0xC0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x46}}

type IEnumFORMATETCInterface interface {
	IUnknownInterface
	Next(celt uint32, rgelt *FORMATETC, pceltFetched *uint32) HRESULT
	Skip(celt uint32) HRESULT
	Reset() HRESULT
	Clone(ppenum **IEnumFORMATETC) HRESULT
}

type IEnumFORMATETCVtbl struct {
	IUnknownVtbl
	Next  uintptr
	Skip  uintptr
	Reset uintptr
	Clone uintptr
}

type IEnumFORMATETC struct {
	IUnknown
}

func (this *IEnumFORMATETC) Vtbl() *IEnumFORMATETCVtbl {
	return (*IEnumFORMATETCVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IEnumFORMATETC) Next(celt uint32, rgelt *FORMATETC, pceltFetched *uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Next, uintptr(unsafe.Pointer(this)), uintptr(celt), uintptr(unsafe.Pointer(rgelt)), uintptr(unsafe.Pointer(pceltFetched)))
	return HRESULT(ret)
}

func (this *IEnumFORMATETC) Skip(celt uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Skip, uintptr(unsafe.Pointer(this)), uintptr(celt))
	return HRESULT(ret)
}

func (this *IEnumFORMATETC) Reset() HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Reset, uintptr(unsafe.Pointer(this)))
	return HRESULT(ret)
}

func (this *IEnumFORMATETC) Clone(ppenum **IEnumFORMATETC) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Clone, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(ppenum)))
	return HRESULT(ret)
}

// 00000105-0000-0000-C000-000000000046
var IID_IEnumSTATDATA = syscall.GUID{0x00000105, 0x0000, 0x0000,
	[8]byte{0xC0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x46}}

type IEnumSTATDATAInterface interface {
	IUnknownInterface
	Next(celt uint32, rgelt *STATDATA, pceltFetched *uint32) HRESULT
	Skip(celt uint32) HRESULT
	Reset() HRESULT
	Clone(ppenum **IEnumSTATDATA) HRESULT
}

type IEnumSTATDATAVtbl struct {
	IUnknownVtbl
	Next  uintptr
	Skip  uintptr
	Reset uintptr
	Clone uintptr
}

type IEnumSTATDATA struct {
	IUnknown
}

func (this *IEnumSTATDATA) Vtbl() *IEnumSTATDATAVtbl {
	return (*IEnumSTATDATAVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IEnumSTATDATA) Next(celt uint32, rgelt *STATDATA, pceltFetched *uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Next, uintptr(unsafe.Pointer(this)), uintptr(celt), uintptr(unsafe.Pointer(rgelt)), uintptr(unsafe.Pointer(pceltFetched)))
	return HRESULT(ret)
}

func (this *IEnumSTATDATA) Skip(celt uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Skip, uintptr(unsafe.Pointer(this)), uintptr(celt))
	return HRESULT(ret)
}

func (this *IEnumSTATDATA) Reset() HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Reset, uintptr(unsafe.Pointer(this)))
	return HRESULT(ret)
}

func (this *IEnumSTATDATA) Clone(ppenum **IEnumSTATDATA) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Clone, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(ppenum)))
	return HRESULT(ret)
}

// 0000010F-0000-0000-C000-000000000046
var IID_IAdviseSink = syscall.GUID{0x0000010F, 0x0000, 0x0000,
	[8]byte{0xC0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x46}}

type IAdviseSinkInterface interface {
	IUnknownInterface
	OnDataChange(pFormatetc *FORMATETC, pStgmed *STGMEDIUM)
	OnViewChange(dwAspect uint32, lindex int32)
	OnRename(pmk *IMoniker)
	OnSave()
	OnClose()
}

type IAdviseSinkVtbl struct {
	IUnknownVtbl
	OnDataChange uintptr
	OnViewChange uintptr
	OnRename     uintptr
	OnSave       uintptr
	OnClose      uintptr
}

type IAdviseSink struct {
	IUnknown
}

func (this *IAdviseSink) Vtbl() *IAdviseSinkVtbl {
	return (*IAdviseSinkVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IAdviseSink) OnDataChange(pFormatetc *FORMATETC, pStgmed *STGMEDIUM) {
	_, _, _ = syscall.SyscallN(this.Vtbl().OnDataChange, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pFormatetc)), uintptr(unsafe.Pointer(pStgmed)))
}

func (this *IAdviseSink) OnViewChange(dwAspect uint32, lindex int32) {
	_, _, _ = syscall.SyscallN(this.Vtbl().OnViewChange, uintptr(unsafe.Pointer(this)), uintptr(dwAspect), uintptr(lindex))
}

func (this *IAdviseSink) OnRename(pmk *IMoniker) {
	_, _, _ = syscall.SyscallN(this.Vtbl().OnRename, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pmk)))
}

func (this *IAdviseSink) OnSave() {
	_, _, _ = syscall.SyscallN(this.Vtbl().OnSave, uintptr(unsafe.Pointer(this)))
}

func (this *IAdviseSink) OnClose() {
	_, _, _ = syscall.SyscallN(this.Vtbl().OnClose, uintptr(unsafe.Pointer(this)))
}

// 00000150-0000-0000-C000-000000000046
var IID_AsyncIAdviseSink = syscall.GUID{0x00000150, 0x0000, 0x0000,
	[8]byte{0xC0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x46}}

type AsyncIAdviseSinkInterface interface {
	IUnknownInterface
	Begin_OnDataChange(pFormatetc *FORMATETC, pStgmed *STGMEDIUM)
	Finish_OnDataChange()
	Begin_OnViewChange(dwAspect uint32, lindex int32)
	Finish_OnViewChange()
	Begin_OnRename(pmk *IMoniker)
	Finish_OnRename()
	Begin_OnSave()
	Finish_OnSave()
	Begin_OnClose()
	Finish_OnClose()
}

type AsyncIAdviseSinkVtbl struct {
	IUnknownVtbl
	Begin_OnDataChange  uintptr
	Finish_OnDataChange uintptr
	Begin_OnViewChange  uintptr
	Finish_OnViewChange uintptr
	Begin_OnRename      uintptr
	Finish_OnRename     uintptr
	Begin_OnSave        uintptr
	Finish_OnSave       uintptr
	Begin_OnClose       uintptr
	Finish_OnClose      uintptr
}

type AsyncIAdviseSink struct {
	IUnknown
}

func (this *AsyncIAdviseSink) Vtbl() *AsyncIAdviseSinkVtbl {
	return (*AsyncIAdviseSinkVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *AsyncIAdviseSink) Begin_OnDataChange(pFormatetc *FORMATETC, pStgmed *STGMEDIUM) {
	_, _, _ = syscall.SyscallN(this.Vtbl().Begin_OnDataChange, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pFormatetc)), uintptr(unsafe.Pointer(pStgmed)))
}

func (this *AsyncIAdviseSink) Finish_OnDataChange() {
	_, _, _ = syscall.SyscallN(this.Vtbl().Finish_OnDataChange, uintptr(unsafe.Pointer(this)))
}

func (this *AsyncIAdviseSink) Begin_OnViewChange(dwAspect uint32, lindex int32) {
	_, _, _ = syscall.SyscallN(this.Vtbl().Begin_OnViewChange, uintptr(unsafe.Pointer(this)), uintptr(dwAspect), uintptr(lindex))
}

func (this *AsyncIAdviseSink) Finish_OnViewChange() {
	_, _, _ = syscall.SyscallN(this.Vtbl().Finish_OnViewChange, uintptr(unsafe.Pointer(this)))
}

func (this *AsyncIAdviseSink) Begin_OnRename(pmk *IMoniker) {
	_, _, _ = syscall.SyscallN(this.Vtbl().Begin_OnRename, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pmk)))
}

func (this *AsyncIAdviseSink) Finish_OnRename() {
	_, _, _ = syscall.SyscallN(this.Vtbl().Finish_OnRename, uintptr(unsafe.Pointer(this)))
}

func (this *AsyncIAdviseSink) Begin_OnSave() {
	_, _, _ = syscall.SyscallN(this.Vtbl().Begin_OnSave, uintptr(unsafe.Pointer(this)))
}

func (this *AsyncIAdviseSink) Finish_OnSave() {
	_, _, _ = syscall.SyscallN(this.Vtbl().Finish_OnSave, uintptr(unsafe.Pointer(this)))
}

func (this *AsyncIAdviseSink) Begin_OnClose() {
	_, _, _ = syscall.SyscallN(this.Vtbl().Begin_OnClose, uintptr(unsafe.Pointer(this)))
}

func (this *AsyncIAdviseSink) Finish_OnClose() {
	_, _, _ = syscall.SyscallN(this.Vtbl().Finish_OnClose, uintptr(unsafe.Pointer(this)))
}

// 00000125-0000-0000-C000-000000000046
var IID_IAdviseSink2 = syscall.GUID{0x00000125, 0x0000, 0x0000,
	[8]byte{0xC0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x46}}

type IAdviseSink2Interface interface {
	IAdviseSinkInterface
	OnLinkSrcChange(pmk *IMoniker)
}

type IAdviseSink2Vtbl struct {
	IAdviseSinkVtbl
	OnLinkSrcChange uintptr
}

type IAdviseSink2 struct {
	IAdviseSink
}

func (this *IAdviseSink2) Vtbl() *IAdviseSink2Vtbl {
	return (*IAdviseSink2Vtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IAdviseSink2) OnLinkSrcChange(pmk *IMoniker) {
	_, _, _ = syscall.SyscallN(this.Vtbl().OnLinkSrcChange, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pmk)))
}

// 00000151-0000-0000-C000-000000000046
var IID_AsyncIAdviseSink2 = syscall.GUID{0x00000151, 0x0000, 0x0000,
	[8]byte{0xC0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x46}}

type AsyncIAdviseSink2Interface interface {
	AsyncIAdviseSinkInterface
	Begin_OnLinkSrcChange(pmk *IMoniker)
	Finish_OnLinkSrcChange()
}

type AsyncIAdviseSink2Vtbl struct {
	AsyncIAdviseSinkVtbl
	Begin_OnLinkSrcChange  uintptr
	Finish_OnLinkSrcChange uintptr
}

type AsyncIAdviseSink2 struct {
	AsyncIAdviseSink
}

func (this *AsyncIAdviseSink2) Vtbl() *AsyncIAdviseSink2Vtbl {
	return (*AsyncIAdviseSink2Vtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *AsyncIAdviseSink2) Begin_OnLinkSrcChange(pmk *IMoniker) {
	_, _, _ = syscall.SyscallN(this.Vtbl().Begin_OnLinkSrcChange, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pmk)))
}

func (this *AsyncIAdviseSink2) Finish_OnLinkSrcChange() {
	_, _, _ = syscall.SyscallN(this.Vtbl().Finish_OnLinkSrcChange, uintptr(unsafe.Pointer(this)))
}

// 0000010E-0000-0000-C000-000000000046
var IID_IDataObject = syscall.GUID{0x0000010E, 0x0000, 0x0000,
	[8]byte{0xC0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x46}}

type IDataObjectInterface interface {
	IUnknownInterface
	GetData(pformatetcIn *FORMATETC, pmedium *STGMEDIUM) HRESULT
	GetDataHere(pformatetc *FORMATETC, pmedium *STGMEDIUM) HRESULT
	QueryGetData(pformatetc *FORMATETC) HRESULT
	GetCanonicalFormatEtc(pformatectIn *FORMATETC, pformatetcOut *FORMATETC) HRESULT
	SetData(pformatetc *FORMATETC, pmedium *STGMEDIUM, fRelease BOOL) HRESULT
	EnumFormatEtc(dwDirection uint32, ppenumFormatEtc **IEnumFORMATETC) HRESULT
	DAdvise(pformatetc *FORMATETC, advf uint32, pAdvSink *IAdviseSink, pdwConnection *uint32) HRESULT
	DUnadvise(dwConnection uint32) HRESULT
	EnumDAdvise(ppenumAdvise **IEnumSTATDATA) HRESULT
}

type IDataObjectVtbl struct {
	IUnknownVtbl
	GetData               uintptr
	GetDataHere           uintptr
	QueryGetData          uintptr
	GetCanonicalFormatEtc uintptr
	SetData               uintptr
	EnumFormatEtc         uintptr
	DAdvise               uintptr
	DUnadvise             uintptr
	EnumDAdvise           uintptr
}

type IDataObject struct {
	IUnknown
}

func (this *IDataObject) Vtbl() *IDataObjectVtbl {
	return (*IDataObjectVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IDataObject) GetData(pformatetcIn *FORMATETC, pmedium *STGMEDIUM) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetData, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pformatetcIn)), uintptr(unsafe.Pointer(pmedium)))
	return HRESULT(ret)
}

func (this *IDataObject) GetDataHere(pformatetc *FORMATETC, pmedium *STGMEDIUM) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetDataHere, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pformatetc)), uintptr(unsafe.Pointer(pmedium)))
	return HRESULT(ret)
}

func (this *IDataObject) QueryGetData(pformatetc *FORMATETC) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().QueryGetData, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pformatetc)))
	return HRESULT(ret)
}

func (this *IDataObject) GetCanonicalFormatEtc(pformatectIn *FORMATETC, pformatetcOut *FORMATETC) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetCanonicalFormatEtc, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pformatectIn)), uintptr(unsafe.Pointer(pformatetcOut)))
	return HRESULT(ret)
}

func (this *IDataObject) SetData(pformatetc *FORMATETC, pmedium *STGMEDIUM, fRelease BOOL) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetData, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pformatetc)), uintptr(unsafe.Pointer(pmedium)), uintptr(fRelease))
	return HRESULT(ret)
}

func (this *IDataObject) EnumFormatEtc(dwDirection uint32, ppenumFormatEtc **IEnumFORMATETC) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().EnumFormatEtc, uintptr(unsafe.Pointer(this)), uintptr(dwDirection), uintptr(unsafe.Pointer(ppenumFormatEtc)))
	return HRESULT(ret)
}

func (this *IDataObject) DAdvise(pformatetc *FORMATETC, advf uint32, pAdvSink *IAdviseSink, pdwConnection *uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().DAdvise, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pformatetc)), uintptr(advf), uintptr(unsafe.Pointer(pAdvSink)), uintptr(unsafe.Pointer(pdwConnection)))
	return HRESULT(ret)
}

func (this *IDataObject) DUnadvise(dwConnection uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().DUnadvise, uintptr(unsafe.Pointer(this)), uintptr(dwConnection))
	return HRESULT(ret)
}

func (this *IDataObject) EnumDAdvise(ppenumAdvise **IEnumSTATDATA) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().EnumDAdvise, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(ppenumAdvise)))
	return HRESULT(ret)
}

// 00000110-0000-0000-C000-000000000046
var IID_IDataAdviseHolder = syscall.GUID{0x00000110, 0x0000, 0x0000,
	[8]byte{0xC0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x46}}

type IDataAdviseHolderInterface interface {
	IUnknownInterface
	Advise(pDataObject *IDataObject, pFetc *FORMATETC, advf uint32, pAdvise *IAdviseSink, pdwConnection *uint32) HRESULT
	Unadvise(dwConnection uint32) HRESULT
	EnumAdvise(ppenumAdvise **IEnumSTATDATA) HRESULT
	SendOnDataChange(pDataObject *IDataObject, dwReserved uint32, advf uint32) HRESULT
}

type IDataAdviseHolderVtbl struct {
	IUnknownVtbl
	Advise           uintptr
	Unadvise         uintptr
	EnumAdvise       uintptr
	SendOnDataChange uintptr
}

type IDataAdviseHolder struct {
	IUnknown
}

func (this *IDataAdviseHolder) Vtbl() *IDataAdviseHolderVtbl {
	return (*IDataAdviseHolderVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IDataAdviseHolder) Advise(pDataObject *IDataObject, pFetc *FORMATETC, advf uint32, pAdvise *IAdviseSink, pdwConnection *uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Advise, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pDataObject)), uintptr(unsafe.Pointer(pFetc)), uintptr(advf), uintptr(unsafe.Pointer(pAdvise)), uintptr(unsafe.Pointer(pdwConnection)))
	return HRESULT(ret)
}

func (this *IDataAdviseHolder) Unadvise(dwConnection uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Unadvise, uintptr(unsafe.Pointer(this)), uintptr(dwConnection))
	return HRESULT(ret)
}

func (this *IDataAdviseHolder) EnumAdvise(ppenumAdvise **IEnumSTATDATA) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().EnumAdvise, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(ppenumAdvise)))
	return HRESULT(ret)
}

func (this *IDataAdviseHolder) SendOnDataChange(pDataObject *IDataObject, dwReserved uint32, advf uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SendOnDataChange, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pDataObject)), uintptr(dwReserved), uintptr(advf))
	return HRESULT(ret)
}

// 00000140-0000-0000-C000-000000000046
var IID_IClassActivator = syscall.GUID{0x00000140, 0x0000, 0x0000,
	[8]byte{0xC0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x46}}

type IClassActivatorInterface interface {
	IUnknownInterface
	GetClassObject(rclsid *syscall.GUID, dwClassContext uint32, locale uint32, riid *syscall.GUID, ppv unsafe.Pointer) HRESULT
}

type IClassActivatorVtbl struct {
	IUnknownVtbl
	GetClassObject uintptr
}

type IClassActivator struct {
	IUnknown
}

func (this *IClassActivator) Vtbl() *IClassActivatorVtbl {
	return (*IClassActivatorVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IClassActivator) GetClassObject(rclsid *syscall.GUID, dwClassContext uint32, locale uint32, riid *syscall.GUID, ppv unsafe.Pointer) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetClassObject, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(rclsid)), uintptr(dwClassContext), uintptr(locale), uintptr(unsafe.Pointer(riid)), uintptr(ppv))
	return HRESULT(ret)
}

// A9D758A0-4617-11CF-95FC-00AA00680DB4
var IID_IProgressNotify = syscall.GUID{0xA9D758A0, 0x4617, 0x11CF,
	[8]byte{0x95, 0xFC, 0x00, 0xAA, 0x00, 0x68, 0x0D, 0xB4}}

type IProgressNotifyInterface interface {
	IUnknownInterface
	OnProgress(dwProgressCurrent uint32, dwProgressMaximum uint32, fAccurate BOOL, fOwner BOOL) HRESULT
}

type IProgressNotifyVtbl struct {
	IUnknownVtbl
	OnProgress uintptr
}

type IProgressNotify struct {
	IUnknown
}

func (this *IProgressNotify) Vtbl() *IProgressNotifyVtbl {
	return (*IProgressNotifyVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IProgressNotify) OnProgress(dwProgressCurrent uint32, dwProgressMaximum uint32, fAccurate BOOL, fOwner BOOL) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().OnProgress, uintptr(unsafe.Pointer(this)), uintptr(dwProgressCurrent), uintptr(dwProgressMaximum), uintptr(fAccurate), uintptr(fOwner))
	return HRESULT(ret)
}

// 30F3D47A-6447-11D1-8E3C-00C04FB9386D
var IID_IBlockingLock = syscall.GUID{0x30F3D47A, 0x6447, 0x11D1,
	[8]byte{0x8E, 0x3C, 0x00, 0xC0, 0x4F, 0xB9, 0x38, 0x6D}}

type IBlockingLockInterface interface {
	IUnknownInterface
	Lock(dwTimeout uint32) HRESULT
	Unlock() HRESULT
}

type IBlockingLockVtbl struct {
	IUnknownVtbl
	Lock   uintptr
	Unlock uintptr
}

type IBlockingLock struct {
	IUnknown
}

func (this *IBlockingLock) Vtbl() *IBlockingLockVtbl {
	return (*IBlockingLockVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IBlockingLock) Lock(dwTimeout uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Lock, uintptr(unsafe.Pointer(this)), uintptr(dwTimeout))
	return HRESULT(ret)
}

func (this *IBlockingLock) Unlock() HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Unlock, uintptr(unsafe.Pointer(this)))
	return HRESULT(ret)
}

// BC0BF6AE-8878-11D1-83E9-00C04FC2C6D4
var IID_ITimeAndNoticeControl = syscall.GUID{0xBC0BF6AE, 0x8878, 0x11D1,
	[8]byte{0x83, 0xE9, 0x00, 0xC0, 0x4F, 0xC2, 0xC6, 0xD4}}

type ITimeAndNoticeControlInterface interface {
	IUnknownInterface
	SuppressChanges(res1 uint32, res2 uint32) HRESULT
}

type ITimeAndNoticeControlVtbl struct {
	IUnknownVtbl
	SuppressChanges uintptr
}

type ITimeAndNoticeControl struct {
	IUnknown
}

func (this *ITimeAndNoticeControl) Vtbl() *ITimeAndNoticeControlVtbl {
	return (*ITimeAndNoticeControlVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *ITimeAndNoticeControl) SuppressChanges(res1 uint32, res2 uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SuppressChanges, uintptr(unsafe.Pointer(this)), uintptr(res1), uintptr(res2))
	return HRESULT(ret)
}

// 8D19C834-8879-11D1-83E9-00C04FC2C6D4
var IID_IOplockStorage = syscall.GUID{0x8D19C834, 0x8879, 0x11D1,
	[8]byte{0x83, 0xE9, 0x00, 0xC0, 0x4F, 0xC2, 0xC6, 0xD4}}

type IOplockStorageInterface interface {
	IUnknownInterface
	CreateStorageEx(pwcsName PWSTR, grfMode uint32, stgfmt uint32, grfAttrs uint32, riid *syscall.GUID, ppstgOpen unsafe.Pointer) HRESULT
	OpenStorageEx(pwcsName PWSTR, grfMode uint32, stgfmt uint32, grfAttrs uint32, riid *syscall.GUID, ppstgOpen unsafe.Pointer) HRESULT
}

type IOplockStorageVtbl struct {
	IUnknownVtbl
	CreateStorageEx uintptr
	OpenStorageEx   uintptr
}

type IOplockStorage struct {
	IUnknown
}

func (this *IOplockStorage) Vtbl() *IOplockStorageVtbl {
	return (*IOplockStorageVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IOplockStorage) CreateStorageEx(pwcsName PWSTR, grfMode uint32, stgfmt uint32, grfAttrs uint32, riid *syscall.GUID, ppstgOpen unsafe.Pointer) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().CreateStorageEx, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pwcsName)), uintptr(grfMode), uintptr(stgfmt), uintptr(grfAttrs), uintptr(unsafe.Pointer(riid)), uintptr(ppstgOpen))
	return HRESULT(ret)
}

func (this *IOplockStorage) OpenStorageEx(pwcsName PWSTR, grfMode uint32, stgfmt uint32, grfAttrs uint32, riid *syscall.GUID, ppstgOpen unsafe.Pointer) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().OpenStorageEx, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pwcsName)), uintptr(grfMode), uintptr(stgfmt), uintptr(grfAttrs), uintptr(unsafe.Pointer(riid)), uintptr(ppstgOpen))
	return HRESULT(ret)
}

// 00000026-0000-0000-C000-000000000046
var IID_IUrlMon = syscall.GUID{0x00000026, 0x0000, 0x0000,
	[8]byte{0xC0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x46}}

type IUrlMonInterface interface {
	IUnknownInterface
	AsyncGetClassBits(rclsid *syscall.GUID, pszTYPE PWSTR, pszExt PWSTR, dwFileVersionMS uint32, dwFileVersionLS uint32, pszCodeBase PWSTR, pbc *IBindCtx, dwClassContext uint32, riid *syscall.GUID, flags uint32) HRESULT
}

type IUrlMonVtbl struct {
	IUnknownVtbl
	AsyncGetClassBits uintptr
}

type IUrlMon struct {
	IUnknown
}

func (this *IUrlMon) Vtbl() *IUrlMonVtbl {
	return (*IUrlMonVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IUrlMon) AsyncGetClassBits(rclsid *syscall.GUID, pszTYPE PWSTR, pszExt PWSTR, dwFileVersionMS uint32, dwFileVersionLS uint32, pszCodeBase PWSTR, pbc *IBindCtx, dwClassContext uint32, riid *syscall.GUID, flags uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().AsyncGetClassBits, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(rclsid)), uintptr(unsafe.Pointer(pszTYPE)), uintptr(unsafe.Pointer(pszExt)), uintptr(dwFileVersionMS), uintptr(dwFileVersionLS), uintptr(unsafe.Pointer(pszCodeBase)), uintptr(unsafe.Pointer(pbc)), uintptr(dwClassContext), uintptr(unsafe.Pointer(riid)), uintptr(flags))
	return HRESULT(ret)
}

// 00000145-0000-0000-C000-000000000046
var IID_IForegroundTransfer = syscall.GUID{0x00000145, 0x0000, 0x0000,
	[8]byte{0xC0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x46}}

type IForegroundTransferInterface interface {
	IUnknownInterface
	AllowForegroundTransfer(lpvReserved unsafe.Pointer) HRESULT
}

type IForegroundTransferVtbl struct {
	IUnknownVtbl
	AllowForegroundTransfer uintptr
}

type IForegroundTransfer struct {
	IUnknown
}

func (this *IForegroundTransfer) Vtbl() *IForegroundTransferVtbl {
	return (*IForegroundTransferVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IForegroundTransfer) AllowForegroundTransfer(lpvReserved unsafe.Pointer) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().AllowForegroundTransfer, uintptr(unsafe.Pointer(this)), uintptr(lpvReserved))
	return HRESULT(ret)
}

// 000001D5-0000-0000-C000-000000000046
var IID_IProcessLock = syscall.GUID{0x000001D5, 0x0000, 0x0000,
	[8]byte{0xC0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x46}}

type IProcessLockInterface interface {
	IUnknownInterface
	AddRefOnProcess() uint32
	ReleaseRefOnProcess() uint32
}

type IProcessLockVtbl struct {
	IUnknownVtbl
	AddRefOnProcess     uintptr
	ReleaseRefOnProcess uintptr
}

type IProcessLock struct {
	IUnknown
}

func (this *IProcessLock) Vtbl() *IProcessLockVtbl {
	return (*IProcessLockVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IProcessLock) AddRefOnProcess() uint32 {
	ret, _, _ := syscall.SyscallN(this.Vtbl().AddRefOnProcess, uintptr(unsafe.Pointer(this)))
	return uint32(ret)
}

func (this *IProcessLock) ReleaseRefOnProcess() uint32 {
	ret, _, _ := syscall.SyscallN(this.Vtbl().ReleaseRefOnProcess, uintptr(unsafe.Pointer(this)))
	return uint32(ret)
}

// 000001D4-0000-0000-C000-000000000046
var IID_ISurrogateService = syscall.GUID{0x000001D4, 0x0000, 0x0000,
	[8]byte{0xC0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x46}}

type ISurrogateServiceInterface interface {
	IUnknownInterface
	Init(rguidProcessID *syscall.GUID, pProcessLock *IProcessLock, pfApplicationAware *BOOL) HRESULT
	ApplicationLaunch(rguidApplID *syscall.GUID, appType ApplicationType) HRESULT
	ApplicationFree(rguidApplID *syscall.GUID) HRESULT
	CatalogRefresh(ulReserved uint32) HRESULT
	ProcessShutdown(shutdownType ShutdownType) HRESULT
}

type ISurrogateServiceVtbl struct {
	IUnknownVtbl
	Init              uintptr
	ApplicationLaunch uintptr
	ApplicationFree   uintptr
	CatalogRefresh    uintptr
	ProcessShutdown   uintptr
}

type ISurrogateService struct {
	IUnknown
}

func (this *ISurrogateService) Vtbl() *ISurrogateServiceVtbl {
	return (*ISurrogateServiceVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *ISurrogateService) Init(rguidProcessID *syscall.GUID, pProcessLock *IProcessLock, pfApplicationAware *BOOL) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Init, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(rguidProcessID)), uintptr(unsafe.Pointer(pProcessLock)), uintptr(unsafe.Pointer(pfApplicationAware)))
	return HRESULT(ret)
}

func (this *ISurrogateService) ApplicationLaunch(rguidApplID *syscall.GUID, appType ApplicationType) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().ApplicationLaunch, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(rguidApplID)), uintptr(appType))
	return HRESULT(ret)
}

func (this *ISurrogateService) ApplicationFree(rguidApplID *syscall.GUID) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().ApplicationFree, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(rguidApplID)))
	return HRESULT(ret)
}

func (this *ISurrogateService) CatalogRefresh(ulReserved uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().CatalogRefresh, uintptr(unsafe.Pointer(this)), uintptr(ulReserved))
	return HRESULT(ret)
}

func (this *ISurrogateService) ProcessShutdown(shutdownType ShutdownType) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().ProcessShutdown, uintptr(unsafe.Pointer(this)), uintptr(shutdownType))
	return HRESULT(ret)
}

// 00000034-0000-0000-C000-000000000046
var IID_IInitializeSpy = syscall.GUID{0x00000034, 0x0000, 0x0000,
	[8]byte{0xC0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x46}}

type IInitializeSpyInterface interface {
	IUnknownInterface
	PreInitialize(dwCoInit uint32, dwCurThreadAptRefs uint32) HRESULT
	PostInitialize(hrCoInit HRESULT, dwCoInit uint32, dwNewThreadAptRefs uint32) HRESULT
	PreUninitialize(dwCurThreadAptRefs uint32) HRESULT
	PostUninitialize(dwNewThreadAptRefs uint32) HRESULT
}

type IInitializeSpyVtbl struct {
	IUnknownVtbl
	PreInitialize    uintptr
	PostInitialize   uintptr
	PreUninitialize  uintptr
	PostUninitialize uintptr
}

type IInitializeSpy struct {
	IUnknown
}

func (this *IInitializeSpy) Vtbl() *IInitializeSpyVtbl {
	return (*IInitializeSpyVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IInitializeSpy) PreInitialize(dwCoInit uint32, dwCurThreadAptRefs uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().PreInitialize, uintptr(unsafe.Pointer(this)), uintptr(dwCoInit), uintptr(dwCurThreadAptRefs))
	return HRESULT(ret)
}

func (this *IInitializeSpy) PostInitialize(hrCoInit HRESULT, dwCoInit uint32, dwNewThreadAptRefs uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().PostInitialize, uintptr(unsafe.Pointer(this)), uintptr(hrCoInit), uintptr(dwCoInit), uintptr(dwNewThreadAptRefs))
	return HRESULT(ret)
}

func (this *IInitializeSpy) PreUninitialize(dwCurThreadAptRefs uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().PreUninitialize, uintptr(unsafe.Pointer(this)), uintptr(dwCurThreadAptRefs))
	return HRESULT(ret)
}

func (this *IInitializeSpy) PostUninitialize(dwNewThreadAptRefs uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().PostUninitialize, uintptr(unsafe.Pointer(this)), uintptr(dwNewThreadAptRefs))
	return HRESULT(ret)
}

// 6D5140C1-7436-11CE-8034-00AA006009FA
var IID_IServiceProvider = syscall.GUID{0x6D5140C1, 0x7436, 0x11CE,
	[8]byte{0x80, 0x34, 0x00, 0xAA, 0x00, 0x60, 0x09, 0xFA}}

type IServiceProviderInterface interface {
	IUnknownInterface
	QueryService(guidService *syscall.GUID, riid *syscall.GUID, ppvObject unsafe.Pointer) HRESULT
}

type IServiceProviderVtbl struct {
	IUnknownVtbl
	QueryService uintptr
}

type IServiceProvider struct {
	IUnknown
}

func (this *IServiceProvider) Vtbl() *IServiceProviderVtbl {
	return (*IServiceProviderVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IServiceProvider) QueryService(guidService *syscall.GUID, riid *syscall.GUID, ppvObject unsafe.Pointer) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().QueryService, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(guidService)), uintptr(unsafe.Pointer(riid)), uintptr(ppvObject))
	return HRESULT(ret)
}

// 0002E000-0000-0000-C000-000000000046
var IID_IEnumGUID = syscall.GUID{0x0002E000, 0x0000, 0x0000,
	[8]byte{0xC0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x46}}

type IEnumGUIDInterface interface {
	IUnknownInterface
	Next(celt uint32, rgelt *syscall.GUID, pceltFetched *uint32) HRESULT
	Skip(celt uint32) HRESULT
	Reset() HRESULT
	Clone(ppenum **IEnumGUID) HRESULT
}

type IEnumGUIDVtbl struct {
	IUnknownVtbl
	Next  uintptr
	Skip  uintptr
	Reset uintptr
	Clone uintptr
}

type IEnumGUID struct {
	IUnknown
}

func (this *IEnumGUID) Vtbl() *IEnumGUIDVtbl {
	return (*IEnumGUIDVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IEnumGUID) Next(celt uint32, rgelt *syscall.GUID, pceltFetched *uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Next, uintptr(unsafe.Pointer(this)), uintptr(celt), uintptr(unsafe.Pointer(rgelt)), uintptr(unsafe.Pointer(pceltFetched)))
	return HRESULT(ret)
}

func (this *IEnumGUID) Skip(celt uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Skip, uintptr(unsafe.Pointer(this)), uintptr(celt))
	return HRESULT(ret)
}

func (this *IEnumGUID) Reset() HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Reset, uintptr(unsafe.Pointer(this)))
	return HRESULT(ret)
}

func (this *IEnumGUID) Clone(ppenum **IEnumGUID) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Clone, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(ppenum)))
	return HRESULT(ret)
}

// 0002E011-0000-0000-C000-000000000046
var IID_IEnumCATEGORYINFO = syscall.GUID{0x0002E011, 0x0000, 0x0000,
	[8]byte{0xC0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x46}}

type IEnumCATEGORYINFOInterface interface {
	IUnknownInterface
	Next(celt uint32, rgelt *CATEGORYINFO, pceltFetched *uint32) HRESULT
	Skip(celt uint32) HRESULT
	Reset() HRESULT
	Clone(ppenum **IEnumCATEGORYINFO) HRESULT
}

type IEnumCATEGORYINFOVtbl struct {
	IUnknownVtbl
	Next  uintptr
	Skip  uintptr
	Reset uintptr
	Clone uintptr
}

type IEnumCATEGORYINFO struct {
	IUnknown
}

func (this *IEnumCATEGORYINFO) Vtbl() *IEnumCATEGORYINFOVtbl {
	return (*IEnumCATEGORYINFOVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IEnumCATEGORYINFO) Next(celt uint32, rgelt *CATEGORYINFO, pceltFetched *uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Next, uintptr(unsafe.Pointer(this)), uintptr(celt), uintptr(unsafe.Pointer(rgelt)), uintptr(unsafe.Pointer(pceltFetched)))
	return HRESULT(ret)
}

func (this *IEnumCATEGORYINFO) Skip(celt uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Skip, uintptr(unsafe.Pointer(this)), uintptr(celt))
	return HRESULT(ret)
}

func (this *IEnumCATEGORYINFO) Reset() HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Reset, uintptr(unsafe.Pointer(this)))
	return HRESULT(ret)
}

func (this *IEnumCATEGORYINFO) Clone(ppenum **IEnumCATEGORYINFO) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Clone, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(ppenum)))
	return HRESULT(ret)
}

// 0002E012-0000-0000-C000-000000000046
var IID_ICatRegister = syscall.GUID{0x0002E012, 0x0000, 0x0000,
	[8]byte{0xC0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x46}}

type ICatRegisterInterface interface {
	IUnknownInterface
	RegisterCategories(cCategories uint32, rgCategoryInfo *CATEGORYINFO) HRESULT
	UnRegisterCategories(cCategories uint32, rgcatid *syscall.GUID) HRESULT
	RegisterClassImplCategories(rclsid *syscall.GUID, cCategories uint32, rgcatid *syscall.GUID) HRESULT
	UnRegisterClassImplCategories(rclsid *syscall.GUID, cCategories uint32, rgcatid *syscall.GUID) HRESULT
	RegisterClassReqCategories(rclsid *syscall.GUID, cCategories uint32, rgcatid *syscall.GUID) HRESULT
	UnRegisterClassReqCategories(rclsid *syscall.GUID, cCategories uint32, rgcatid *syscall.GUID) HRESULT
}

type ICatRegisterVtbl struct {
	IUnknownVtbl
	RegisterCategories            uintptr
	UnRegisterCategories          uintptr
	RegisterClassImplCategories   uintptr
	UnRegisterClassImplCategories uintptr
	RegisterClassReqCategories    uintptr
	UnRegisterClassReqCategories  uintptr
}

type ICatRegister struct {
	IUnknown
}

func (this *ICatRegister) Vtbl() *ICatRegisterVtbl {
	return (*ICatRegisterVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *ICatRegister) RegisterCategories(cCategories uint32, rgCategoryInfo *CATEGORYINFO) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().RegisterCategories, uintptr(unsafe.Pointer(this)), uintptr(cCategories), uintptr(unsafe.Pointer(rgCategoryInfo)))
	return HRESULT(ret)
}

func (this *ICatRegister) UnRegisterCategories(cCategories uint32, rgcatid *syscall.GUID) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().UnRegisterCategories, uintptr(unsafe.Pointer(this)), uintptr(cCategories), uintptr(unsafe.Pointer(rgcatid)))
	return HRESULT(ret)
}

func (this *ICatRegister) RegisterClassImplCategories(rclsid *syscall.GUID, cCategories uint32, rgcatid *syscall.GUID) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().RegisterClassImplCategories, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(rclsid)), uintptr(cCategories), uintptr(unsafe.Pointer(rgcatid)))
	return HRESULT(ret)
}

func (this *ICatRegister) UnRegisterClassImplCategories(rclsid *syscall.GUID, cCategories uint32, rgcatid *syscall.GUID) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().UnRegisterClassImplCategories, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(rclsid)), uintptr(cCategories), uintptr(unsafe.Pointer(rgcatid)))
	return HRESULT(ret)
}

func (this *ICatRegister) RegisterClassReqCategories(rclsid *syscall.GUID, cCategories uint32, rgcatid *syscall.GUID) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().RegisterClassReqCategories, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(rclsid)), uintptr(cCategories), uintptr(unsafe.Pointer(rgcatid)))
	return HRESULT(ret)
}

func (this *ICatRegister) UnRegisterClassReqCategories(rclsid *syscall.GUID, cCategories uint32, rgcatid *syscall.GUID) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().UnRegisterClassReqCategories, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(rclsid)), uintptr(cCategories), uintptr(unsafe.Pointer(rgcatid)))
	return HRESULT(ret)
}

// 0002E013-0000-0000-C000-000000000046
var IID_ICatInformation = syscall.GUID{0x0002E013, 0x0000, 0x0000,
	[8]byte{0xC0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x46}}

type ICatInformationInterface interface {
	IUnknownInterface
	EnumCategories(lcid uint32, ppenumCategoryInfo **IEnumCATEGORYINFO) HRESULT
	GetCategoryDesc(rcatid *syscall.GUID, lcid uint32, pszDesc *PWSTR) HRESULT
	EnumClassesOfCategories(cImplemented uint32, rgcatidImpl *syscall.GUID, cRequired uint32, rgcatidReq *syscall.GUID, ppenumClsid **IEnumGUID) HRESULT
	IsClassOfCategories(rclsid *syscall.GUID, cImplemented uint32, rgcatidImpl *syscall.GUID, cRequired uint32, rgcatidReq *syscall.GUID) HRESULT
	EnumImplCategoriesOfClass(rclsid *syscall.GUID, ppenumCatid **IEnumGUID) HRESULT
	EnumReqCategoriesOfClass(rclsid *syscall.GUID, ppenumCatid **IEnumGUID) HRESULT
}

type ICatInformationVtbl struct {
	IUnknownVtbl
	EnumCategories            uintptr
	GetCategoryDesc           uintptr
	EnumClassesOfCategories   uintptr
	IsClassOfCategories       uintptr
	EnumImplCategoriesOfClass uintptr
	EnumReqCategoriesOfClass  uintptr
}

type ICatInformation struct {
	IUnknown
}

func (this *ICatInformation) Vtbl() *ICatInformationVtbl {
	return (*ICatInformationVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *ICatInformation) EnumCategories(lcid uint32, ppenumCategoryInfo **IEnumCATEGORYINFO) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().EnumCategories, uintptr(unsafe.Pointer(this)), uintptr(lcid), uintptr(unsafe.Pointer(ppenumCategoryInfo)))
	return HRESULT(ret)
}

func (this *ICatInformation) GetCategoryDesc(rcatid *syscall.GUID, lcid uint32, pszDesc *PWSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetCategoryDesc, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(rcatid)), uintptr(lcid), uintptr(unsafe.Pointer(pszDesc)))
	return HRESULT(ret)
}

func (this *ICatInformation) EnumClassesOfCategories(cImplemented uint32, rgcatidImpl *syscall.GUID, cRequired uint32, rgcatidReq *syscall.GUID, ppenumClsid **IEnumGUID) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().EnumClassesOfCategories, uintptr(unsafe.Pointer(this)), uintptr(cImplemented), uintptr(unsafe.Pointer(rgcatidImpl)), uintptr(cRequired), uintptr(unsafe.Pointer(rgcatidReq)), uintptr(unsafe.Pointer(ppenumClsid)))
	return HRESULT(ret)
}

func (this *ICatInformation) IsClassOfCategories(rclsid *syscall.GUID, cImplemented uint32, rgcatidImpl *syscall.GUID, cRequired uint32, rgcatidReq *syscall.GUID) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().IsClassOfCategories, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(rclsid)), uintptr(cImplemented), uintptr(unsafe.Pointer(rgcatidImpl)), uintptr(cRequired), uintptr(unsafe.Pointer(rgcatidReq)))
	return HRESULT(ret)
}

func (this *ICatInformation) EnumImplCategoriesOfClass(rclsid *syscall.GUID, ppenumCatid **IEnumGUID) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().EnumImplCategoriesOfClass, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(rclsid)), uintptr(unsafe.Pointer(ppenumCatid)))
	return HRESULT(ret)
}

func (this *ICatInformation) EnumReqCategoriesOfClass(rclsid *syscall.GUID, ppenumCatid **IEnumGUID) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().EnumReqCategoriesOfClass, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(rclsid)), uintptr(unsafe.Pointer(ppenumCatid)))
	return HRESULT(ret)
}

// 000001DA-0000-0000-C000-000000000046
var IID_IContextCallback = syscall.GUID{0x000001DA, 0x0000, 0x0000,
	[8]byte{0xC0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x46}}

type IContextCallbackInterface interface {
	IUnknownInterface
	ContextCallback(pfnCallback PFNCONTEXTCALL, pParam *ComCallData, riid *syscall.GUID, iMethod int32, pUnk *IUnknown) HRESULT
}

type IContextCallbackVtbl struct {
	IUnknownVtbl
	ContextCallback uintptr
}

type IContextCallback struct {
	IUnknown
}

func (this *IContextCallback) Vtbl() *IContextCallbackVtbl {
	return (*IContextCallbackVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IContextCallback) ContextCallback(pfnCallback PFNCONTEXTCALL, pParam *ComCallData, riid *syscall.GUID, iMethod int32, pUnk *IUnknown) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().ContextCallback, uintptr(unsafe.Pointer(this)), pfnCallback, uintptr(unsafe.Pointer(pParam)), uintptr(unsafe.Pointer(riid)), uintptr(iMethod), uintptr(unsafe.Pointer(pUnk)))
	return HRESULT(ret)
}

// 79EAC9C0-BAF9-11CE-8C82-00AA004BA90B
var IID_IBinding = syscall.GUID{0x79EAC9C0, 0xBAF9, 0x11CE,
	[8]byte{0x8C, 0x82, 0x00, 0xAA, 0x00, 0x4B, 0xA9, 0x0B}}

type IBindingInterface interface {
	IUnknownInterface
	Abort() HRESULT
	Suspend() HRESULT
	Resume() HRESULT
	SetPriority(nPriority int32) HRESULT
	GetPriority(pnPriority *int32) HRESULT
	GetBindResult(pclsidProtocol *syscall.GUID, pdwResult *uint32, pszResult *PWSTR, pdwReserved *uint32) HRESULT
}

type IBindingVtbl struct {
	IUnknownVtbl
	Abort         uintptr
	Suspend       uintptr
	Resume        uintptr
	SetPriority   uintptr
	GetPriority   uintptr
	GetBindResult uintptr
}

type IBinding struct {
	IUnknown
}

func (this *IBinding) Vtbl() *IBindingVtbl {
	return (*IBindingVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IBinding) Abort() HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Abort, uintptr(unsafe.Pointer(this)))
	return HRESULT(ret)
}

func (this *IBinding) Suspend() HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Suspend, uintptr(unsafe.Pointer(this)))
	return HRESULT(ret)
}

func (this *IBinding) Resume() HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Resume, uintptr(unsafe.Pointer(this)))
	return HRESULT(ret)
}

func (this *IBinding) SetPriority(nPriority int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetPriority, uintptr(unsafe.Pointer(this)), uintptr(nPriority))
	return HRESULT(ret)
}

func (this *IBinding) GetPriority(pnPriority *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetPriority, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pnPriority)))
	return HRESULT(ret)
}

func (this *IBinding) GetBindResult(pclsidProtocol *syscall.GUID, pdwResult *uint32, pszResult *PWSTR, pdwReserved *uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetBindResult, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pclsidProtocol)), uintptr(unsafe.Pointer(pdwResult)), uintptr(unsafe.Pointer(pszResult)), uintptr(unsafe.Pointer(pdwReserved)))
	return HRESULT(ret)
}

// 79EAC9C1-BAF9-11CE-8C82-00AA004BA90B
var IID_IBindStatusCallback = syscall.GUID{0x79EAC9C1, 0xBAF9, 0x11CE,
	[8]byte{0x8C, 0x82, 0x00, 0xAA, 0x00, 0x4B, 0xA9, 0x0B}}

type IBindStatusCallbackInterface interface {
	IUnknownInterface
	OnStartBinding(dwReserved uint32, pib *IBinding) HRESULT
	GetPriority(pnPriority *int32) HRESULT
	OnLowResource(reserved uint32) HRESULT
	OnProgress(ulProgress uint32, ulProgressMax uint32, ulStatusCode uint32, szStatusText PWSTR) HRESULT
	OnStopBinding(hresult HRESULT, szError PWSTR) HRESULT
	GetBindInfo(grfBINDF *uint32, pbindinfo *BINDINFO) HRESULT
	OnDataAvailable(grfBSCF uint32, dwSize uint32, pformatetc *FORMATETC, pstgmed *STGMEDIUM) HRESULT
	OnObjectAvailable(riid *syscall.GUID, punk *IUnknown) HRESULT
}

type IBindStatusCallbackVtbl struct {
	IUnknownVtbl
	OnStartBinding    uintptr
	GetPriority       uintptr
	OnLowResource     uintptr
	OnProgress        uintptr
	OnStopBinding     uintptr
	GetBindInfo       uintptr
	OnDataAvailable   uintptr
	OnObjectAvailable uintptr
}

type IBindStatusCallback struct {
	IUnknown
}

func (this *IBindStatusCallback) Vtbl() *IBindStatusCallbackVtbl {
	return (*IBindStatusCallbackVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IBindStatusCallback) OnStartBinding(dwReserved uint32, pib *IBinding) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().OnStartBinding, uintptr(unsafe.Pointer(this)), uintptr(dwReserved), uintptr(unsafe.Pointer(pib)))
	return HRESULT(ret)
}

func (this *IBindStatusCallback) GetPriority(pnPriority *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetPriority, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pnPriority)))
	return HRESULT(ret)
}

func (this *IBindStatusCallback) OnLowResource(reserved uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().OnLowResource, uintptr(unsafe.Pointer(this)), uintptr(reserved))
	return HRESULT(ret)
}

func (this *IBindStatusCallback) OnProgress(ulProgress uint32, ulProgressMax uint32, ulStatusCode uint32, szStatusText PWSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().OnProgress, uintptr(unsafe.Pointer(this)), uintptr(ulProgress), uintptr(ulProgressMax), uintptr(ulStatusCode), uintptr(unsafe.Pointer(szStatusText)))
	return HRESULT(ret)
}

func (this *IBindStatusCallback) OnStopBinding(hresult HRESULT, szError PWSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().OnStopBinding, uintptr(unsafe.Pointer(this)), uintptr(hresult), uintptr(unsafe.Pointer(szError)))
	return HRESULT(ret)
}

func (this *IBindStatusCallback) GetBindInfo(grfBINDF *uint32, pbindinfo *BINDINFO) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetBindInfo, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(grfBINDF)), uintptr(unsafe.Pointer(pbindinfo)))
	return HRESULT(ret)
}

func (this *IBindStatusCallback) OnDataAvailable(grfBSCF uint32, dwSize uint32, pformatetc *FORMATETC, pstgmed *STGMEDIUM) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().OnDataAvailable, uintptr(unsafe.Pointer(this)), uintptr(grfBSCF), uintptr(dwSize), uintptr(unsafe.Pointer(pformatetc)), uintptr(unsafe.Pointer(pstgmed)))
	return HRESULT(ret)
}

func (this *IBindStatusCallback) OnObjectAvailable(riid *syscall.GUID, punk *IUnknown) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().OnObjectAvailable, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(riid)), uintptr(unsafe.Pointer(punk)))
	return HRESULT(ret)
}

// AAA74EF9-8EE7-4659-88D9-F8C504DA73CC
var IID_IBindStatusCallbackEx = syscall.GUID{0xAAA74EF9, 0x8EE7, 0x4659,
	[8]byte{0x88, 0xD9, 0xF8, 0xC5, 0x04, 0xDA, 0x73, 0xCC}}

type IBindStatusCallbackExInterface interface {
	IBindStatusCallbackInterface
	GetBindInfoEx(grfBINDF *uint32, pbindinfo *BINDINFO, grfBINDF2 *uint32, pdwReserved *uint32) HRESULT
}

type IBindStatusCallbackExVtbl struct {
	IBindStatusCallbackVtbl
	GetBindInfoEx uintptr
}

type IBindStatusCallbackEx struct {
	IBindStatusCallback
}

func (this *IBindStatusCallbackEx) Vtbl() *IBindStatusCallbackExVtbl {
	return (*IBindStatusCallbackExVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IBindStatusCallbackEx) GetBindInfoEx(grfBINDF *uint32, pbindinfo *BINDINFO, grfBINDF2 *uint32, pdwReserved *uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetBindInfoEx, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(grfBINDF)), uintptr(unsafe.Pointer(pbindinfo)), uintptr(unsafe.Pointer(grfBINDF2)), uintptr(unsafe.Pointer(pdwReserved)))
	return HRESULT(ret)
}

// 79EAC9D0-BAF9-11CE-8C82-00AA004BA90B
var IID_IAuthenticate = syscall.GUID{0x79EAC9D0, 0xBAF9, 0x11CE,
	[8]byte{0x8C, 0x82, 0x00, 0xAA, 0x00, 0x4B, 0xA9, 0x0B}}

type IAuthenticateInterface interface {
	IUnknownInterface
	Authenticate(phwnd *HWND, pszUsername *PWSTR, pszPassword *PWSTR) HRESULT
}

type IAuthenticateVtbl struct {
	IUnknownVtbl
	Authenticate uintptr
}

type IAuthenticate struct {
	IUnknown
}

func (this *IAuthenticate) Vtbl() *IAuthenticateVtbl {
	return (*IAuthenticateVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IAuthenticate) Authenticate(phwnd *HWND, pszUsername *PWSTR, pszPassword *PWSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Authenticate, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(phwnd)), uintptr(unsafe.Pointer(pszUsername)), uintptr(unsafe.Pointer(pszPassword)))
	return HRESULT(ret)
}

// 2AD1EDAF-D83D-48B5-9ADF-03DBE19F53BD
var IID_IAuthenticateEx = syscall.GUID{0x2AD1EDAF, 0xD83D, 0x48B5,
	[8]byte{0x9A, 0xDF, 0x03, 0xDB, 0xE1, 0x9F, 0x53, 0xBD}}

type IAuthenticateExInterface interface {
	IAuthenticateInterface
	AuthenticateEx(phwnd *HWND, pszUsername *PWSTR, pszPassword *PWSTR, pauthinfo *AUTHENTICATEINFO) HRESULT
}

type IAuthenticateExVtbl struct {
	IAuthenticateVtbl
	AuthenticateEx uintptr
}

type IAuthenticateEx struct {
	IAuthenticate
}

func (this *IAuthenticateEx) Vtbl() *IAuthenticateExVtbl {
	return (*IAuthenticateExVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IAuthenticateEx) AuthenticateEx(phwnd *HWND, pszUsername *PWSTR, pszPassword *PWSTR, pauthinfo *AUTHENTICATEINFO) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().AuthenticateEx, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(phwnd)), uintptr(unsafe.Pointer(pszUsername)), uintptr(unsafe.Pointer(pszPassword)), uintptr(unsafe.Pointer(pauthinfo)))
	return HRESULT(ret)
}

// A39EE748-6A27-4817-A6F2-13914BEF5890
var IID_IUri = syscall.GUID{0xA39EE748, 0x6A27, 0x4817,
	[8]byte{0xA6, 0xF2, 0x13, 0x91, 0x4B, 0xEF, 0x58, 0x90}}

type IUriInterface interface {
	IUnknownInterface
	GetPropertyBSTR(uriProp Uri_PROPERTY, pbstrProperty *BSTR, dwFlags uint32) HRESULT
	GetPropertyLength(uriProp Uri_PROPERTY, pcchProperty *uint32, dwFlags uint32) HRESULT
	GetPropertyDWORD(uriProp Uri_PROPERTY, pdwProperty *uint32, dwFlags uint32) HRESULT
	HasProperty(uriProp Uri_PROPERTY, pfHasProperty *BOOL) HRESULT
	GetAbsoluteUri(pbstrAbsoluteUri *BSTR) HRESULT
	GetAuthority(pbstrAuthority *BSTR) HRESULT
	GetDisplayUri(pbstrDisplayString *BSTR) HRESULT
	GetDomain(pbstrDomain *BSTR) HRESULT
	GetExtension(pbstrExtension *BSTR) HRESULT
	GetFragment(pbstrFragment *BSTR) HRESULT
	GetHost(pbstrHost *BSTR) HRESULT
	GetPassword(pbstrPassword *BSTR) HRESULT
	GetPath(pbstrPath *BSTR) HRESULT
	GetPathAndQuery(pbstrPathAndQuery *BSTR) HRESULT
	GetQuery(pbstrQuery *BSTR) HRESULT
	GetRawUri(pbstrRawUri *BSTR) HRESULT
	GetSchemeName(pbstrSchemeName *BSTR) HRESULT
	GetUserInfo(pbstrUserInfo *BSTR) HRESULT
	GetUserName(pbstrUserName *BSTR) HRESULT
	GetHostType(pdwHostType *uint32) HRESULT
	GetPort(pdwPort *uint32) HRESULT
	GetScheme(pdwScheme *uint32) HRESULT
	GetZone(pdwZone *uint32) HRESULT
	GetProperties(pdwFlags *uint32) HRESULT
	IsEqual(pUri *IUri, pfEqual *BOOL) HRESULT
}

type IUriVtbl struct {
	IUnknownVtbl
	GetPropertyBSTR   uintptr
	GetPropertyLength uintptr
	GetPropertyDWORD  uintptr
	HasProperty       uintptr
	GetAbsoluteUri    uintptr
	GetAuthority      uintptr
	GetDisplayUri     uintptr
	GetDomain         uintptr
	GetExtension      uintptr
	GetFragment       uintptr
	GetHost           uintptr
	GetPassword       uintptr
	GetPath           uintptr
	GetPathAndQuery   uintptr
	GetQuery          uintptr
	GetRawUri         uintptr
	GetSchemeName     uintptr
	GetUserInfo       uintptr
	GetUserName       uintptr
	GetHostType       uintptr
	GetPort           uintptr
	GetScheme         uintptr
	GetZone           uintptr
	GetProperties     uintptr
	IsEqual           uintptr
}

type IUri struct {
	IUnknown
}

func (this *IUri) Vtbl() *IUriVtbl {
	return (*IUriVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IUri) GetPropertyBSTR(uriProp Uri_PROPERTY, pbstrProperty *BSTR, dwFlags uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetPropertyBSTR, uintptr(unsafe.Pointer(this)), uintptr(uriProp), uintptr(unsafe.Pointer(pbstrProperty)), uintptr(dwFlags))
	return HRESULT(ret)
}

func (this *IUri) GetPropertyLength(uriProp Uri_PROPERTY, pcchProperty *uint32, dwFlags uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetPropertyLength, uintptr(unsafe.Pointer(this)), uintptr(uriProp), uintptr(unsafe.Pointer(pcchProperty)), uintptr(dwFlags))
	return HRESULT(ret)
}

func (this *IUri) GetPropertyDWORD(uriProp Uri_PROPERTY, pdwProperty *uint32, dwFlags uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetPropertyDWORD, uintptr(unsafe.Pointer(this)), uintptr(uriProp), uintptr(unsafe.Pointer(pdwProperty)), uintptr(dwFlags))
	return HRESULT(ret)
}

func (this *IUri) HasProperty(uriProp Uri_PROPERTY, pfHasProperty *BOOL) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().HasProperty, uintptr(unsafe.Pointer(this)), uintptr(uriProp), uintptr(unsafe.Pointer(pfHasProperty)))
	return HRESULT(ret)
}

func (this *IUri) GetAbsoluteUri(pbstrAbsoluteUri *BSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetAbsoluteUri, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pbstrAbsoluteUri)))
	return HRESULT(ret)
}

func (this *IUri) GetAuthority(pbstrAuthority *BSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetAuthority, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pbstrAuthority)))
	return HRESULT(ret)
}

func (this *IUri) GetDisplayUri(pbstrDisplayString *BSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetDisplayUri, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pbstrDisplayString)))
	return HRESULT(ret)
}

func (this *IUri) GetDomain(pbstrDomain *BSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetDomain, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pbstrDomain)))
	return HRESULT(ret)
}

func (this *IUri) GetExtension(pbstrExtension *BSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetExtension, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pbstrExtension)))
	return HRESULT(ret)
}

func (this *IUri) GetFragment(pbstrFragment *BSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetFragment, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pbstrFragment)))
	return HRESULT(ret)
}

func (this *IUri) GetHost(pbstrHost *BSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetHost, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pbstrHost)))
	return HRESULT(ret)
}

func (this *IUri) GetPassword(pbstrPassword *BSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetPassword, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pbstrPassword)))
	return HRESULT(ret)
}

func (this *IUri) GetPath(pbstrPath *BSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetPath, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pbstrPath)))
	return HRESULT(ret)
}

func (this *IUri) GetPathAndQuery(pbstrPathAndQuery *BSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetPathAndQuery, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pbstrPathAndQuery)))
	return HRESULT(ret)
}

func (this *IUri) GetQuery(pbstrQuery *BSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetQuery, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pbstrQuery)))
	return HRESULT(ret)
}

func (this *IUri) GetRawUri(pbstrRawUri *BSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetRawUri, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pbstrRawUri)))
	return HRESULT(ret)
}

func (this *IUri) GetSchemeName(pbstrSchemeName *BSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetSchemeName, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pbstrSchemeName)))
	return HRESULT(ret)
}

func (this *IUri) GetUserInfo(pbstrUserInfo *BSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetUserInfo, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pbstrUserInfo)))
	return HRESULT(ret)
}

func (this *IUri) GetUserName(pbstrUserName *BSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetUserName, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pbstrUserName)))
	return HRESULT(ret)
}

func (this *IUri) GetHostType(pdwHostType *uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetHostType, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pdwHostType)))
	return HRESULT(ret)
}

func (this *IUri) GetPort(pdwPort *uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetPort, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pdwPort)))
	return HRESULT(ret)
}

func (this *IUri) GetScheme(pdwScheme *uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetScheme, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pdwScheme)))
	return HRESULT(ret)
}

func (this *IUri) GetZone(pdwZone *uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetZone, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pdwZone)))
	return HRESULT(ret)
}

func (this *IUri) GetProperties(pdwFlags *uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetProperties, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pdwFlags)))
	return HRESULT(ret)
}

func (this *IUri) IsEqual(pUri *IUri, pfEqual *BOOL) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().IsEqual, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pUri)), uintptr(unsafe.Pointer(pfEqual)))
	return HRESULT(ret)
}

// 4221B2E1-8955-46C0-BD5B-DE9897565DE7
var IID_IUriBuilder = syscall.GUID{0x4221B2E1, 0x8955, 0x46C0,
	[8]byte{0xBD, 0x5B, 0xDE, 0x98, 0x97, 0x56, 0x5D, 0xE7}}

type IUriBuilderInterface interface {
	IUnknownInterface
	CreateUriSimple(dwAllowEncodingPropertyMask uint32, dwReserved uintptr, ppIUri **IUri) HRESULT
	CreateUri(dwCreateFlags uint32, dwAllowEncodingPropertyMask uint32, dwReserved uintptr, ppIUri **IUri) HRESULT
	CreateUriWithFlags(dwCreateFlags uint32, dwUriBuilderFlags uint32, dwAllowEncodingPropertyMask uint32, dwReserved uintptr, ppIUri **IUri) HRESULT
	GetIUri(ppIUri **IUri) HRESULT
	SetIUri(pIUri *IUri) HRESULT
	GetFragment(pcchFragment *uint32, ppwzFragment *PWSTR) HRESULT
	GetHost(pcchHost *uint32, ppwzHost *PWSTR) HRESULT
	GetPassword(pcchPassword *uint32, ppwzPassword *PWSTR) HRESULT
	GetPath(pcchPath *uint32, ppwzPath *PWSTR) HRESULT
	GetPort(pfHasPort *BOOL, pdwPort *uint32) HRESULT
	GetQuery(pcchQuery *uint32, ppwzQuery *PWSTR) HRESULT
	GetSchemeName(pcchSchemeName *uint32, ppwzSchemeName *PWSTR) HRESULT
	GetUserName(pcchUserName *uint32, ppwzUserName *PWSTR) HRESULT
	SetFragment(pwzNewValue PWSTR) HRESULT
	SetHost(pwzNewValue PWSTR) HRESULT
	SetPassword(pwzNewValue PWSTR) HRESULT
	SetPath(pwzNewValue PWSTR) HRESULT
	SetPort(fHasPort BOOL, dwNewValue uint32) HRESULT
	SetQuery(pwzNewValue PWSTR) HRESULT
	SetSchemeName(pwzNewValue PWSTR) HRESULT
	SetUserName(pwzNewValue PWSTR) HRESULT
	RemoveProperties(dwPropertyMask uint32) HRESULT
	HasBeenModified(pfModified *BOOL) HRESULT
}

type IUriBuilderVtbl struct {
	IUnknownVtbl
	CreateUriSimple    uintptr
	CreateUri          uintptr
	CreateUriWithFlags uintptr
	GetIUri            uintptr
	SetIUri            uintptr
	GetFragment        uintptr
	GetHost            uintptr
	GetPassword        uintptr
	GetPath            uintptr
	GetPort            uintptr
	GetQuery           uintptr
	GetSchemeName      uintptr
	GetUserName        uintptr
	SetFragment        uintptr
	SetHost            uintptr
	SetPassword        uintptr
	SetPath            uintptr
	SetPort            uintptr
	SetQuery           uintptr
	SetSchemeName      uintptr
	SetUserName        uintptr
	RemoveProperties   uintptr
	HasBeenModified    uintptr
}

type IUriBuilder struct {
	IUnknown
}

func (this *IUriBuilder) Vtbl() *IUriBuilderVtbl {
	return (*IUriBuilderVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IUriBuilder) CreateUriSimple(dwAllowEncodingPropertyMask uint32, dwReserved uintptr, ppIUri **IUri) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().CreateUriSimple, uintptr(unsafe.Pointer(this)), uintptr(dwAllowEncodingPropertyMask), dwReserved, uintptr(unsafe.Pointer(ppIUri)))
	return HRESULT(ret)
}

func (this *IUriBuilder) CreateUri(dwCreateFlags uint32, dwAllowEncodingPropertyMask uint32, dwReserved uintptr, ppIUri **IUri) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().CreateUri, uintptr(unsafe.Pointer(this)), uintptr(dwCreateFlags), uintptr(dwAllowEncodingPropertyMask), dwReserved, uintptr(unsafe.Pointer(ppIUri)))
	return HRESULT(ret)
}

func (this *IUriBuilder) CreateUriWithFlags(dwCreateFlags uint32, dwUriBuilderFlags uint32, dwAllowEncodingPropertyMask uint32, dwReserved uintptr, ppIUri **IUri) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().CreateUriWithFlags, uintptr(unsafe.Pointer(this)), uintptr(dwCreateFlags), uintptr(dwUriBuilderFlags), uintptr(dwAllowEncodingPropertyMask), dwReserved, uintptr(unsafe.Pointer(ppIUri)))
	return HRESULT(ret)
}

func (this *IUriBuilder) GetIUri(ppIUri **IUri) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetIUri, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(ppIUri)))
	return HRESULT(ret)
}

func (this *IUriBuilder) SetIUri(pIUri *IUri) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetIUri, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pIUri)))
	return HRESULT(ret)
}

func (this *IUriBuilder) GetFragment(pcchFragment *uint32, ppwzFragment *PWSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetFragment, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pcchFragment)), uintptr(unsafe.Pointer(ppwzFragment)))
	return HRESULT(ret)
}

func (this *IUriBuilder) GetHost(pcchHost *uint32, ppwzHost *PWSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetHost, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pcchHost)), uintptr(unsafe.Pointer(ppwzHost)))
	return HRESULT(ret)
}

func (this *IUriBuilder) GetPassword(pcchPassword *uint32, ppwzPassword *PWSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetPassword, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pcchPassword)), uintptr(unsafe.Pointer(ppwzPassword)))
	return HRESULT(ret)
}

func (this *IUriBuilder) GetPath(pcchPath *uint32, ppwzPath *PWSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetPath, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pcchPath)), uintptr(unsafe.Pointer(ppwzPath)))
	return HRESULT(ret)
}

func (this *IUriBuilder) GetPort(pfHasPort *BOOL, pdwPort *uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetPort, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pfHasPort)), uintptr(unsafe.Pointer(pdwPort)))
	return HRESULT(ret)
}

func (this *IUriBuilder) GetQuery(pcchQuery *uint32, ppwzQuery *PWSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetQuery, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pcchQuery)), uintptr(unsafe.Pointer(ppwzQuery)))
	return HRESULT(ret)
}

func (this *IUriBuilder) GetSchemeName(pcchSchemeName *uint32, ppwzSchemeName *PWSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetSchemeName, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pcchSchemeName)), uintptr(unsafe.Pointer(ppwzSchemeName)))
	return HRESULT(ret)
}

func (this *IUriBuilder) GetUserName(pcchUserName *uint32, ppwzUserName *PWSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetUserName, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pcchUserName)), uintptr(unsafe.Pointer(ppwzUserName)))
	return HRESULT(ret)
}

func (this *IUriBuilder) SetFragment(pwzNewValue PWSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetFragment, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pwzNewValue)))
	return HRESULT(ret)
}

func (this *IUriBuilder) SetHost(pwzNewValue PWSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetHost, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pwzNewValue)))
	return HRESULT(ret)
}

func (this *IUriBuilder) SetPassword(pwzNewValue PWSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetPassword, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pwzNewValue)))
	return HRESULT(ret)
}

func (this *IUriBuilder) SetPath(pwzNewValue PWSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetPath, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pwzNewValue)))
	return HRESULT(ret)
}

func (this *IUriBuilder) SetPort(fHasPort BOOL, dwNewValue uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetPort, uintptr(unsafe.Pointer(this)), uintptr(fHasPort), uintptr(dwNewValue))
	return HRESULT(ret)
}

func (this *IUriBuilder) SetQuery(pwzNewValue PWSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetQuery, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pwzNewValue)))
	return HRESULT(ret)
}

func (this *IUriBuilder) SetSchemeName(pwzNewValue PWSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetSchemeName, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pwzNewValue)))
	return HRESULT(ret)
}

func (this *IUriBuilder) SetUserName(pwzNewValue PWSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetUserName, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pwzNewValue)))
	return HRESULT(ret)
}

func (this *IUriBuilder) RemoveProperties(dwPropertyMask uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().RemoveProperties, uintptr(unsafe.Pointer(this)), uintptr(dwPropertyMask))
	return HRESULT(ret)
}

func (this *IUriBuilder) HasBeenModified(pfModified *BOOL) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().HasBeenModified, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pfModified)))
	return HRESULT(ret)
}

// FC4801A1-2BA9-11CF-A229-00AA003D7352
var IID_IBindHost = syscall.GUID{0xFC4801A1, 0x2BA9, 0x11CF,
	[8]byte{0xA2, 0x29, 0x00, 0xAA, 0x00, 0x3D, 0x73, 0x52}}

type IBindHostInterface interface {
	IUnknownInterface
	CreateMoniker(szName PWSTR, pBC *IBindCtx, ppmk **IMoniker, dwReserved uint32) HRESULT
	MonikerBindToStorage(pMk *IMoniker, pBC *IBindCtx, pBSC *IBindStatusCallback, riid *syscall.GUID, ppvObj unsafe.Pointer) HRESULT
	MonikerBindToObject(pMk *IMoniker, pBC *IBindCtx, pBSC *IBindStatusCallback, riid *syscall.GUID, ppvObj unsafe.Pointer) HRESULT
}

type IBindHostVtbl struct {
	IUnknownVtbl
	CreateMoniker        uintptr
	MonikerBindToStorage uintptr
	MonikerBindToObject  uintptr
}

type IBindHost struct {
	IUnknown
}

func (this *IBindHost) Vtbl() *IBindHostVtbl {
	return (*IBindHostVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IBindHost) CreateMoniker(szName PWSTR, pBC *IBindCtx, ppmk **IMoniker, dwReserved uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().CreateMoniker, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(szName)), uintptr(unsafe.Pointer(pBC)), uintptr(unsafe.Pointer(ppmk)), uintptr(dwReserved))
	return HRESULT(ret)
}

func (this *IBindHost) MonikerBindToStorage(pMk *IMoniker, pBC *IBindCtx, pBSC *IBindStatusCallback, riid *syscall.GUID, ppvObj unsafe.Pointer) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().MonikerBindToStorage, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pMk)), uintptr(unsafe.Pointer(pBC)), uintptr(unsafe.Pointer(pBSC)), uintptr(unsafe.Pointer(riid)), uintptr(ppvObj))
	return HRESULT(ret)
}

func (this *IBindHost) MonikerBindToObject(pMk *IMoniker, pBC *IBindCtx, pBSC *IBindStatusCallback, riid *syscall.GUID, ppvObj unsafe.Pointer) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().MonikerBindToObject, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pMk)), uintptr(unsafe.Pointer(pBC)), uintptr(unsafe.Pointer(pBSC)), uintptr(unsafe.Pointer(riid)), uintptr(ppvObj))
	return HRESULT(ret)
}

// 00020400-0000-0000-C000-000000000046
var IID_IDispatch = syscall.GUID{0x00020400, 0x0000, 0x0000,
	[8]byte{0xC0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x46}}

type IDispatchInterface interface {
	IUnknownInterface
	GetTypeInfoCount(pctinfo *uint32) HRESULT
	GetTypeInfo(iTInfo uint32, lcid uint32, ppTInfo **ITypeInfo) HRESULT
	GetIDsOfNames(riid *syscall.GUID, rgszNames *PWSTR, cNames uint32, lcid uint32, rgDispId *int32) HRESULT
	Invoke(dispIdMember int32, riid *syscall.GUID, lcid uint32, wFlags uint16, pDispParams *DISPPARAMS, pVarResult *VARIANT, pExcepInfo *EXCEPINFO, puArgErr *uint32) HRESULT
}

type IDispatchVtbl struct {
	IUnknownVtbl
	GetTypeInfoCount uintptr
	GetTypeInfo      uintptr
	GetIDsOfNames    uintptr
	Invoke           uintptr
}

type IDispatch struct {
	IUnknown
}

func (this *IDispatch) Vtbl() *IDispatchVtbl {
	return (*IDispatchVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IDispatch) GetTypeInfoCount(pctinfo *uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetTypeInfoCount, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pctinfo)))
	return HRESULT(ret)
}

func (this *IDispatch) GetTypeInfo(iTInfo uint32, lcid uint32, ppTInfo **ITypeInfo) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetTypeInfo, uintptr(unsafe.Pointer(this)), uintptr(iTInfo), uintptr(lcid), uintptr(unsafe.Pointer(ppTInfo)))
	return HRESULT(ret)
}

func (this *IDispatch) GetIDsOfNames(riid *syscall.GUID, rgszNames *PWSTR, cNames uint32, lcid uint32, rgDispId *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetIDsOfNames, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(riid)), uintptr(unsafe.Pointer(rgszNames)), uintptr(cNames), uintptr(lcid), uintptr(unsafe.Pointer(rgDispId)))
	return HRESULT(ret)
}

func (this *IDispatch) Invoke(dispIdMember int32, riid *syscall.GUID, lcid uint32, wFlags uint16, pDispParams *DISPPARAMS, pVarResult *VARIANT, pExcepInfo *EXCEPINFO, puArgErr *uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Invoke, uintptr(unsafe.Pointer(this)), uintptr(dispIdMember), uintptr(unsafe.Pointer(riid)), uintptr(lcid), uintptr(wFlags), uintptr(unsafe.Pointer(pDispParams)), uintptr(unsafe.Pointer(pVarResult)), uintptr(unsafe.Pointer(pExcepInfo)), uintptr(unsafe.Pointer(puArgErr)))
	return HRESULT(ret)
}

// 00020403-0000-0000-C000-000000000046
var IID_ITypeComp = syscall.GUID{0x00020403, 0x0000, 0x0000,
	[8]byte{0xC0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x46}}

type ITypeCompInterface interface {
	IUnknownInterface
	Bind(szName PWSTR, lHashVal uint32, wFlags uint16, ppTInfo **ITypeInfo, pDescKind *DESCKIND, pBindPtr *BINDPTR) HRESULT
	BindType(szName PWSTR, lHashVal uint32, ppTInfo **ITypeInfo, ppTComp **ITypeComp) HRESULT
}

type ITypeCompVtbl struct {
	IUnknownVtbl
	Bind     uintptr
	BindType uintptr
}

type ITypeComp struct {
	IUnknown
}

func (this *ITypeComp) Vtbl() *ITypeCompVtbl {
	return (*ITypeCompVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *ITypeComp) Bind(szName PWSTR, lHashVal uint32, wFlags uint16, ppTInfo **ITypeInfo, pDescKind *DESCKIND, pBindPtr *BINDPTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Bind, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(szName)), uintptr(lHashVal), uintptr(wFlags), uintptr(unsafe.Pointer(ppTInfo)), uintptr(unsafe.Pointer(pDescKind)), uintptr(unsafe.Pointer(pBindPtr)))
	return HRESULT(ret)
}

func (this *ITypeComp) BindType(szName PWSTR, lHashVal uint32, ppTInfo **ITypeInfo, ppTComp **ITypeComp) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().BindType, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(szName)), uintptr(lHashVal), uintptr(unsafe.Pointer(ppTInfo)), uintptr(unsafe.Pointer(ppTComp)))
	return HRESULT(ret)
}

// 00020401-0000-0000-C000-000000000046
var IID_ITypeInfo = syscall.GUID{0x00020401, 0x0000, 0x0000,
	[8]byte{0xC0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x46}}

type ITypeInfoInterface interface {
	IUnknownInterface
	GetTypeAttr(ppTypeAttr **TYPEATTR) HRESULT
	GetTypeComp(ppTComp **ITypeComp) HRESULT
	GetFuncDesc(index uint32, ppFuncDesc **FUNCDESC) HRESULT
	GetVarDesc(index uint32, ppVarDesc **VARDESC) HRESULT
	GetNames(memid int32, rgBstrNames *BSTR, cMaxNames uint32, pcNames *uint32) HRESULT
	GetRefTypeOfImplType(index uint32, pRefType *uint32) HRESULT
	GetImplTypeFlags(index uint32, pImplTypeFlags *int32) HRESULT
	GetIDsOfNames(rgszNames *PWSTR, cNames uint32, pMemId *int32) HRESULT
	Invoke(pvInstance unsafe.Pointer, memid int32, wFlags uint16, pDispParams *DISPPARAMS, pVarResult *VARIANT, pExcepInfo *EXCEPINFO, puArgErr *uint32) HRESULT
	GetDocumentation(memid int32, pBstrName *BSTR, pBstrDocString *BSTR, pdwHelpContext *uint32, pBstrHelpFile *BSTR) HRESULT
	GetDllEntry(memid int32, invKind INVOKEKIND, pBstrDllName *BSTR, pBstrName *BSTR, pwOrdinal *uint16) HRESULT
	GetRefTypeInfo(hRefType uint32, ppTInfo **ITypeInfo) HRESULT
	AddressOfMember(memid int32, invKind INVOKEKIND, ppv unsafe.Pointer) HRESULT
	CreateInstance(pUnkOuter *IUnknown, riid *syscall.GUID, ppvObj unsafe.Pointer) HRESULT
	GetMops(memid int32, pBstrMops *BSTR) HRESULT
	GetContainingTypeLib(ppTLib **ITypeLib, pIndex *uint32) HRESULT
	ReleaseTypeAttr(pTypeAttr *TYPEATTR)
	ReleaseFuncDesc(pFuncDesc *FUNCDESC)
	ReleaseVarDesc(pVarDesc *VARDESC)
}

type ITypeInfoVtbl struct {
	IUnknownVtbl
	GetTypeAttr          uintptr
	GetTypeComp          uintptr
	GetFuncDesc          uintptr
	GetVarDesc           uintptr
	GetNames             uintptr
	GetRefTypeOfImplType uintptr
	GetImplTypeFlags     uintptr
	GetIDsOfNames        uintptr
	Invoke               uintptr
	GetDocumentation     uintptr
	GetDllEntry          uintptr
	GetRefTypeInfo       uintptr
	AddressOfMember      uintptr
	CreateInstance       uintptr
	GetMops              uintptr
	GetContainingTypeLib uintptr
	ReleaseTypeAttr      uintptr
	ReleaseFuncDesc      uintptr
	ReleaseVarDesc       uintptr
}

type ITypeInfo struct {
	IUnknown
}

func (this *ITypeInfo) Vtbl() *ITypeInfoVtbl {
	return (*ITypeInfoVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *ITypeInfo) GetTypeAttr(ppTypeAttr **TYPEATTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetTypeAttr, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(ppTypeAttr)))
	return HRESULT(ret)
}

func (this *ITypeInfo) GetTypeComp(ppTComp **ITypeComp) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetTypeComp, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(ppTComp)))
	return HRESULT(ret)
}

func (this *ITypeInfo) GetFuncDesc(index uint32, ppFuncDesc **FUNCDESC) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetFuncDesc, uintptr(unsafe.Pointer(this)), uintptr(index), uintptr(unsafe.Pointer(ppFuncDesc)))
	return HRESULT(ret)
}

func (this *ITypeInfo) GetVarDesc(index uint32, ppVarDesc **VARDESC) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetVarDesc, uintptr(unsafe.Pointer(this)), uintptr(index), uintptr(unsafe.Pointer(ppVarDesc)))
	return HRESULT(ret)
}

func (this *ITypeInfo) GetNames(memid int32, rgBstrNames *BSTR, cMaxNames uint32, pcNames *uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetNames, uintptr(unsafe.Pointer(this)), uintptr(memid), uintptr(unsafe.Pointer(rgBstrNames)), uintptr(cMaxNames), uintptr(unsafe.Pointer(pcNames)))
	return HRESULT(ret)
}

func (this *ITypeInfo) GetRefTypeOfImplType(index uint32, pRefType *uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetRefTypeOfImplType, uintptr(unsafe.Pointer(this)), uintptr(index), uintptr(unsafe.Pointer(pRefType)))
	return HRESULT(ret)
}

func (this *ITypeInfo) GetImplTypeFlags(index uint32, pImplTypeFlags *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetImplTypeFlags, uintptr(unsafe.Pointer(this)), uintptr(index), uintptr(unsafe.Pointer(pImplTypeFlags)))
	return HRESULT(ret)
}

func (this *ITypeInfo) GetIDsOfNames(rgszNames *PWSTR, cNames uint32, pMemId *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetIDsOfNames, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(rgszNames)), uintptr(cNames), uintptr(unsafe.Pointer(pMemId)))
	return HRESULT(ret)
}

func (this *ITypeInfo) Invoke(pvInstance unsafe.Pointer, memid int32, wFlags uint16, pDispParams *DISPPARAMS, pVarResult *VARIANT, pExcepInfo *EXCEPINFO, puArgErr *uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Invoke, uintptr(unsafe.Pointer(this)), uintptr(pvInstance), uintptr(memid), uintptr(wFlags), uintptr(unsafe.Pointer(pDispParams)), uintptr(unsafe.Pointer(pVarResult)), uintptr(unsafe.Pointer(pExcepInfo)), uintptr(unsafe.Pointer(puArgErr)))
	return HRESULT(ret)
}

func (this *ITypeInfo) GetDocumentation(memid int32, pBstrName *BSTR, pBstrDocString *BSTR, pdwHelpContext *uint32, pBstrHelpFile *BSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetDocumentation, uintptr(unsafe.Pointer(this)), uintptr(memid), uintptr(unsafe.Pointer(pBstrName)), uintptr(unsafe.Pointer(pBstrDocString)), uintptr(unsafe.Pointer(pdwHelpContext)), uintptr(unsafe.Pointer(pBstrHelpFile)))
	return HRESULT(ret)
}

func (this *ITypeInfo) GetDllEntry(memid int32, invKind INVOKEKIND, pBstrDllName *BSTR, pBstrName *BSTR, pwOrdinal *uint16) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetDllEntry, uintptr(unsafe.Pointer(this)), uintptr(memid), uintptr(invKind), uintptr(unsafe.Pointer(pBstrDllName)), uintptr(unsafe.Pointer(pBstrName)), uintptr(unsafe.Pointer(pwOrdinal)))
	return HRESULT(ret)
}

func (this *ITypeInfo) GetRefTypeInfo(hRefType uint32, ppTInfo **ITypeInfo) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetRefTypeInfo, uintptr(unsafe.Pointer(this)), uintptr(hRefType), uintptr(unsafe.Pointer(ppTInfo)))
	return HRESULT(ret)
}

func (this *ITypeInfo) AddressOfMember(memid int32, invKind INVOKEKIND, ppv unsafe.Pointer) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().AddressOfMember, uintptr(unsafe.Pointer(this)), uintptr(memid), uintptr(invKind), uintptr(ppv))
	return HRESULT(ret)
}

func (this *ITypeInfo) CreateInstance(pUnkOuter *IUnknown, riid *syscall.GUID, ppvObj unsafe.Pointer) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().CreateInstance, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pUnkOuter)), uintptr(unsafe.Pointer(riid)), uintptr(ppvObj))
	return HRESULT(ret)
}

func (this *ITypeInfo) GetMops(memid int32, pBstrMops *BSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetMops, uintptr(unsafe.Pointer(this)), uintptr(memid), uintptr(unsafe.Pointer(pBstrMops)))
	return HRESULT(ret)
}

func (this *ITypeInfo) GetContainingTypeLib(ppTLib **ITypeLib, pIndex *uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetContainingTypeLib, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(ppTLib)), uintptr(unsafe.Pointer(pIndex)))
	return HRESULT(ret)
}

func (this *ITypeInfo) ReleaseTypeAttr(pTypeAttr *TYPEATTR) {
	_, _, _ = syscall.SyscallN(this.Vtbl().ReleaseTypeAttr, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pTypeAttr)))
}

func (this *ITypeInfo) ReleaseFuncDesc(pFuncDesc *FUNCDESC) {
	_, _, _ = syscall.SyscallN(this.Vtbl().ReleaseFuncDesc, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pFuncDesc)))
}

func (this *ITypeInfo) ReleaseVarDesc(pVarDesc *VARDESC) {
	_, _, _ = syscall.SyscallN(this.Vtbl().ReleaseVarDesc, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pVarDesc)))
}

// 00020412-0000-0000-C000-000000000046
var IID_ITypeInfo2 = syscall.GUID{0x00020412, 0x0000, 0x0000,
	[8]byte{0xC0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x46}}

type ITypeInfo2Interface interface {
	ITypeInfoInterface
	GetTypeKind(pTypeKind *TYPEKIND) HRESULT
	GetTypeFlags(pTypeFlags *uint32) HRESULT
	GetFuncIndexOfMemId(memid int32, invKind INVOKEKIND, pFuncIndex *uint32) HRESULT
	GetVarIndexOfMemId(memid int32, pVarIndex *uint32) HRESULT
	GetCustData(guid *syscall.GUID, pVarVal *VARIANT) HRESULT
	GetFuncCustData(index uint32, guid *syscall.GUID, pVarVal *VARIANT) HRESULT
	GetParamCustData(indexFunc uint32, indexParam uint32, guid *syscall.GUID, pVarVal *VARIANT) HRESULT
	GetVarCustData(index uint32, guid *syscall.GUID, pVarVal *VARIANT) HRESULT
	GetImplTypeCustData(index uint32, guid *syscall.GUID, pVarVal *VARIANT) HRESULT
	GetDocumentation2(memid int32, lcid uint32, pbstrHelpString *BSTR, pdwHelpStringContext *uint32, pbstrHelpStringDll *BSTR) HRESULT
	GetAllCustData(pCustData *CUSTDATA) HRESULT
	GetAllFuncCustData(index uint32, pCustData *CUSTDATA) HRESULT
	GetAllParamCustData(indexFunc uint32, indexParam uint32, pCustData *CUSTDATA) HRESULT
	GetAllVarCustData(index uint32, pCustData *CUSTDATA) HRESULT
	GetAllImplTypeCustData(index uint32, pCustData *CUSTDATA) HRESULT
}

type ITypeInfo2Vtbl struct {
	ITypeInfoVtbl
	GetTypeKind            uintptr
	GetTypeFlags           uintptr
	GetFuncIndexOfMemId    uintptr
	GetVarIndexOfMemId     uintptr
	GetCustData            uintptr
	GetFuncCustData        uintptr
	GetParamCustData       uintptr
	GetVarCustData         uintptr
	GetImplTypeCustData    uintptr
	GetDocumentation2      uintptr
	GetAllCustData         uintptr
	GetAllFuncCustData     uintptr
	GetAllParamCustData    uintptr
	GetAllVarCustData      uintptr
	GetAllImplTypeCustData uintptr
}

type ITypeInfo2 struct {
	ITypeInfo
}

func (this *ITypeInfo2) Vtbl() *ITypeInfo2Vtbl {
	return (*ITypeInfo2Vtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *ITypeInfo2) GetTypeKind(pTypeKind *TYPEKIND) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetTypeKind, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pTypeKind)))
	return HRESULT(ret)
}

func (this *ITypeInfo2) GetTypeFlags(pTypeFlags *uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetTypeFlags, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pTypeFlags)))
	return HRESULT(ret)
}

func (this *ITypeInfo2) GetFuncIndexOfMemId(memid int32, invKind INVOKEKIND, pFuncIndex *uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetFuncIndexOfMemId, uintptr(unsafe.Pointer(this)), uintptr(memid), uintptr(invKind), uintptr(unsafe.Pointer(pFuncIndex)))
	return HRESULT(ret)
}

func (this *ITypeInfo2) GetVarIndexOfMemId(memid int32, pVarIndex *uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetVarIndexOfMemId, uintptr(unsafe.Pointer(this)), uintptr(memid), uintptr(unsafe.Pointer(pVarIndex)))
	return HRESULT(ret)
}

func (this *ITypeInfo2) GetCustData(guid *syscall.GUID, pVarVal *VARIANT) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetCustData, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(guid)), uintptr(unsafe.Pointer(pVarVal)))
	return HRESULT(ret)
}

func (this *ITypeInfo2) GetFuncCustData(index uint32, guid *syscall.GUID, pVarVal *VARIANT) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetFuncCustData, uintptr(unsafe.Pointer(this)), uintptr(index), uintptr(unsafe.Pointer(guid)), uintptr(unsafe.Pointer(pVarVal)))
	return HRESULT(ret)
}

func (this *ITypeInfo2) GetParamCustData(indexFunc uint32, indexParam uint32, guid *syscall.GUID, pVarVal *VARIANT) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetParamCustData, uintptr(unsafe.Pointer(this)), uintptr(indexFunc), uintptr(indexParam), uintptr(unsafe.Pointer(guid)), uintptr(unsafe.Pointer(pVarVal)))
	return HRESULT(ret)
}

func (this *ITypeInfo2) GetVarCustData(index uint32, guid *syscall.GUID, pVarVal *VARIANT) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetVarCustData, uintptr(unsafe.Pointer(this)), uintptr(index), uintptr(unsafe.Pointer(guid)), uintptr(unsafe.Pointer(pVarVal)))
	return HRESULT(ret)
}

func (this *ITypeInfo2) GetImplTypeCustData(index uint32, guid *syscall.GUID, pVarVal *VARIANT) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetImplTypeCustData, uintptr(unsafe.Pointer(this)), uintptr(index), uintptr(unsafe.Pointer(guid)), uintptr(unsafe.Pointer(pVarVal)))
	return HRESULT(ret)
}

func (this *ITypeInfo2) GetDocumentation2(memid int32, lcid uint32, pbstrHelpString *BSTR, pdwHelpStringContext *uint32, pbstrHelpStringDll *BSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetDocumentation2, uintptr(unsafe.Pointer(this)), uintptr(memid), uintptr(lcid), uintptr(unsafe.Pointer(pbstrHelpString)), uintptr(unsafe.Pointer(pdwHelpStringContext)), uintptr(unsafe.Pointer(pbstrHelpStringDll)))
	return HRESULT(ret)
}

func (this *ITypeInfo2) GetAllCustData(pCustData *CUSTDATA) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetAllCustData, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pCustData)))
	return HRESULT(ret)
}

func (this *ITypeInfo2) GetAllFuncCustData(index uint32, pCustData *CUSTDATA) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetAllFuncCustData, uintptr(unsafe.Pointer(this)), uintptr(index), uintptr(unsafe.Pointer(pCustData)))
	return HRESULT(ret)
}

func (this *ITypeInfo2) GetAllParamCustData(indexFunc uint32, indexParam uint32, pCustData *CUSTDATA) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetAllParamCustData, uintptr(unsafe.Pointer(this)), uintptr(indexFunc), uintptr(indexParam), uintptr(unsafe.Pointer(pCustData)))
	return HRESULT(ret)
}

func (this *ITypeInfo2) GetAllVarCustData(index uint32, pCustData *CUSTDATA) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetAllVarCustData, uintptr(unsafe.Pointer(this)), uintptr(index), uintptr(unsafe.Pointer(pCustData)))
	return HRESULT(ret)
}

func (this *ITypeInfo2) GetAllImplTypeCustData(index uint32, pCustData *CUSTDATA) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetAllImplTypeCustData, uintptr(unsafe.Pointer(this)), uintptr(index), uintptr(unsafe.Pointer(pCustData)))
	return HRESULT(ret)
}

// 00020402-0000-0000-C000-000000000046
var IID_ITypeLib = syscall.GUID{0x00020402, 0x0000, 0x0000,
	[8]byte{0xC0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x46}}

type ITypeLibInterface interface {
	IUnknownInterface
	GetTypeInfoCount() uint32
	GetTypeInfo(index uint32, ppTInfo **ITypeInfo) HRESULT
	GetTypeInfoType(index uint32, pTKind *TYPEKIND) HRESULT
	GetTypeInfoOfGuid(guid *syscall.GUID, ppTinfo **ITypeInfo) HRESULT
	GetLibAttr(ppTLibAttr **TLIBATTR) HRESULT
	GetTypeComp(ppTComp **ITypeComp) HRESULT
	GetDocumentation(index int32, pBstrName *BSTR, pBstrDocString *BSTR, pdwHelpContext *uint32, pBstrHelpFile *BSTR) HRESULT
	IsName(szNameBuf PWSTR, lHashVal uint32, pfName *BOOL) HRESULT
	FindName(szNameBuf PWSTR, lHashVal uint32, ppTInfo **ITypeInfo, rgMemId *int32, pcFound *uint16) HRESULT
	ReleaseTLibAttr(pTLibAttr *TLIBATTR)
}

type ITypeLibVtbl struct {
	IUnknownVtbl
	GetTypeInfoCount  uintptr
	GetTypeInfo       uintptr
	GetTypeInfoType   uintptr
	GetTypeInfoOfGuid uintptr
	GetLibAttr        uintptr
	GetTypeComp       uintptr
	GetDocumentation  uintptr
	IsName            uintptr
	FindName          uintptr
	ReleaseTLibAttr   uintptr
}

type ITypeLib struct {
	IUnknown
}

func (this *ITypeLib) Vtbl() *ITypeLibVtbl {
	return (*ITypeLibVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *ITypeLib) GetTypeInfoCount() uint32 {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetTypeInfoCount, uintptr(unsafe.Pointer(this)))
	return uint32(ret)
}

func (this *ITypeLib) GetTypeInfo(index uint32, ppTInfo **ITypeInfo) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetTypeInfo, uintptr(unsafe.Pointer(this)), uintptr(index), uintptr(unsafe.Pointer(ppTInfo)))
	return HRESULT(ret)
}

func (this *ITypeLib) GetTypeInfoType(index uint32, pTKind *TYPEKIND) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetTypeInfoType, uintptr(unsafe.Pointer(this)), uintptr(index), uintptr(unsafe.Pointer(pTKind)))
	return HRESULT(ret)
}

func (this *ITypeLib) GetTypeInfoOfGuid(guid *syscall.GUID, ppTinfo **ITypeInfo) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetTypeInfoOfGuid, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(guid)), uintptr(unsafe.Pointer(ppTinfo)))
	return HRESULT(ret)
}

func (this *ITypeLib) GetLibAttr(ppTLibAttr **TLIBATTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetLibAttr, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(ppTLibAttr)))
	return HRESULT(ret)
}

func (this *ITypeLib) GetTypeComp(ppTComp **ITypeComp) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetTypeComp, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(ppTComp)))
	return HRESULT(ret)
}

func (this *ITypeLib) GetDocumentation(index int32, pBstrName *BSTR, pBstrDocString *BSTR, pdwHelpContext *uint32, pBstrHelpFile *BSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetDocumentation, uintptr(unsafe.Pointer(this)), uintptr(index), uintptr(unsafe.Pointer(pBstrName)), uintptr(unsafe.Pointer(pBstrDocString)), uintptr(unsafe.Pointer(pdwHelpContext)), uintptr(unsafe.Pointer(pBstrHelpFile)))
	return HRESULT(ret)
}

func (this *ITypeLib) IsName(szNameBuf PWSTR, lHashVal uint32, pfName *BOOL) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().IsName, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(szNameBuf)), uintptr(lHashVal), uintptr(unsafe.Pointer(pfName)))
	return HRESULT(ret)
}

func (this *ITypeLib) FindName(szNameBuf PWSTR, lHashVal uint32, ppTInfo **ITypeInfo, rgMemId *int32, pcFound *uint16) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().FindName, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(szNameBuf)), uintptr(lHashVal), uintptr(unsafe.Pointer(ppTInfo)), uintptr(unsafe.Pointer(rgMemId)), uintptr(unsafe.Pointer(pcFound)))
	return HRESULT(ret)
}

func (this *ITypeLib) ReleaseTLibAttr(pTLibAttr *TLIBATTR) {
	_, _, _ = syscall.SyscallN(this.Vtbl().ReleaseTLibAttr, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pTLibAttr)))
}

// 00020411-0000-0000-C000-000000000046
var IID_ITypeLib2 = syscall.GUID{0x00020411, 0x0000, 0x0000,
	[8]byte{0xC0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x46}}

type ITypeLib2Interface interface {
	ITypeLibInterface
	GetCustData(guid *syscall.GUID, pVarVal *VARIANT) HRESULT
	GetLibStatistics(pcUniqueNames *uint32, pcchUniqueNames *uint32) HRESULT
	GetDocumentation2(index int32, lcid uint32, pbstrHelpString *BSTR, pdwHelpStringContext *uint32, pbstrHelpStringDll *BSTR) HRESULT
	GetAllCustData(pCustData *CUSTDATA) HRESULT
}

type ITypeLib2Vtbl struct {
	ITypeLibVtbl
	GetCustData       uintptr
	GetLibStatistics  uintptr
	GetDocumentation2 uintptr
	GetAllCustData    uintptr
}

type ITypeLib2 struct {
	ITypeLib
}

func (this *ITypeLib2) Vtbl() *ITypeLib2Vtbl {
	return (*ITypeLib2Vtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *ITypeLib2) GetCustData(guid *syscall.GUID, pVarVal *VARIANT) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetCustData, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(guid)), uintptr(unsafe.Pointer(pVarVal)))
	return HRESULT(ret)
}

func (this *ITypeLib2) GetLibStatistics(pcUniqueNames *uint32, pcchUniqueNames *uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetLibStatistics, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pcUniqueNames)), uintptr(unsafe.Pointer(pcchUniqueNames)))
	return HRESULT(ret)
}

func (this *ITypeLib2) GetDocumentation2(index int32, lcid uint32, pbstrHelpString *BSTR, pdwHelpStringContext *uint32, pbstrHelpStringDll *BSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetDocumentation2, uintptr(unsafe.Pointer(this)), uintptr(index), uintptr(lcid), uintptr(unsafe.Pointer(pbstrHelpString)), uintptr(unsafe.Pointer(pdwHelpStringContext)), uintptr(unsafe.Pointer(pbstrHelpStringDll)))
	return HRESULT(ret)
}

func (this *ITypeLib2) GetAllCustData(pCustData *CUSTDATA) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetAllCustData, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pCustData)))
	return HRESULT(ret)
}

// 1CF2B120-547D-101B-8E65-08002B2BD119
var IID_IErrorInfo = syscall.GUID{0x1CF2B120, 0x547D, 0x101B,
	[8]byte{0x8E, 0x65, 0x08, 0x00, 0x2B, 0x2B, 0xD1, 0x19}}

type IErrorInfoInterface interface {
	IUnknownInterface
	GetGUID(pGUID *syscall.GUID) HRESULT
	GetSource(pBstrSource *BSTR) HRESULT
	GetDescription(pBstrDescription *BSTR) HRESULT
	GetHelpFile(pBstrHelpFile *BSTR) HRESULT
	GetHelpContext(pdwHelpContext *uint32) HRESULT
}

type IErrorInfoVtbl struct {
	IUnknownVtbl
	GetGUID        uintptr
	GetSource      uintptr
	GetDescription uintptr
	GetHelpFile    uintptr
	GetHelpContext uintptr
}

type IErrorInfo struct {
	IUnknown
}

func (this *IErrorInfo) Vtbl() *IErrorInfoVtbl {
	return (*IErrorInfoVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IErrorInfo) GetGUID(pGUID *syscall.GUID) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetGUID, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pGUID)))
	return HRESULT(ret)
}

func (this *IErrorInfo) GetSource(pBstrSource *BSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetSource, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pBstrSource)))
	return HRESULT(ret)
}

func (this *IErrorInfo) GetDescription(pBstrDescription *BSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetDescription, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pBstrDescription)))
	return HRESULT(ret)
}

func (this *IErrorInfo) GetHelpFile(pBstrHelpFile *BSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetHelpFile, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pBstrHelpFile)))
	return HRESULT(ret)
}

func (this *IErrorInfo) GetHelpContext(pdwHelpContext *uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetHelpContext, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pdwHelpContext)))
	return HRESULT(ret)
}

// DF0B3D60-548F-101B-8E65-08002B2BD119
var IID_ISupportErrorInfo = syscall.GUID{0xDF0B3D60, 0x548F, 0x101B,
	[8]byte{0x8E, 0x65, 0x08, 0x00, 0x2B, 0x2B, 0xD1, 0x19}}

type ISupportErrorInfoInterface interface {
	IUnknownInterface
	InterfaceSupportsErrorInfo(riid *syscall.GUID) HRESULT
}

type ISupportErrorInfoVtbl struct {
	IUnknownVtbl
	InterfaceSupportsErrorInfo uintptr
}

type ISupportErrorInfo struct {
	IUnknown
}

func (this *ISupportErrorInfo) Vtbl() *ISupportErrorInfoVtbl {
	return (*ISupportErrorInfoVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *ISupportErrorInfo) InterfaceSupportsErrorInfo(riid *syscall.GUID) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().InterfaceSupportsErrorInfo, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(riid)))
	return HRESULT(ret)
}

// 3127CA40-446E-11CE-8135-00AA004BB851
var IID_IErrorLog = syscall.GUID{0x3127CA40, 0x446E, 0x11CE,
	[8]byte{0x81, 0x35, 0x00, 0xAA, 0x00, 0x4B, 0xB8, 0x51}}

type IErrorLogInterface interface {
	IUnknownInterface
	AddError(pszPropName PWSTR, pExcepInfo *EXCEPINFO) HRESULT
}

type IErrorLogVtbl struct {
	IUnknownVtbl
	AddError uintptr
}

type IErrorLog struct {
	IUnknown
}

func (this *IErrorLog) Vtbl() *IErrorLogVtbl {
	return (*IErrorLogVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IErrorLog) AddError(pszPropName PWSTR, pExcepInfo *EXCEPINFO) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().AddError, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pszPropName)), uintptr(unsafe.Pointer(pExcepInfo)))
	return HRESULT(ret)
}

// ED6A8A2A-B160-4E77-8F73-AA7435CD5C27
var IID_ITypeLibRegistrationReader = syscall.GUID{0xED6A8A2A, 0xB160, 0x4E77,
	[8]byte{0x8F, 0x73, 0xAA, 0x74, 0x35, 0xCD, 0x5C, 0x27}}

type ITypeLibRegistrationReaderInterface interface {
	IUnknownInterface
	EnumTypeLibRegistrations(ppEnumUnknown **IEnumUnknown) HRESULT
}

type ITypeLibRegistrationReaderVtbl struct {
	IUnknownVtbl
	EnumTypeLibRegistrations uintptr
}

type ITypeLibRegistrationReader struct {
	IUnknown
}

func (this *ITypeLibRegistrationReader) Vtbl() *ITypeLibRegistrationReaderVtbl {
	return (*ITypeLibRegistrationReaderVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *ITypeLibRegistrationReader) EnumTypeLibRegistrations(ppEnumUnknown **IEnumUnknown) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().EnumTypeLibRegistrations, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(ppEnumUnknown)))
	return HRESULT(ret)
}

// 76A3E735-02DF-4A12-98EB-043AD3600AF3
var IID_ITypeLibRegistration = syscall.GUID{0x76A3E735, 0x02DF, 0x4A12,
	[8]byte{0x98, 0xEB, 0x04, 0x3A, 0xD3, 0x60, 0x0A, 0xF3}}

type ITypeLibRegistrationInterface interface {
	IUnknownInterface
	GetGuid(pGuid *syscall.GUID) HRESULT
	GetVersion(pVersion *BSTR) HRESULT
	GetLcid(pLcid *uint32) HRESULT
	GetWin32Path(pWin32Path *BSTR) HRESULT
	GetWin64Path(pWin64Path *BSTR) HRESULT
	GetDisplayName(pDisplayName *BSTR) HRESULT
	GetFlags(pFlags *uint32) HRESULT
	GetHelpDir(pHelpDir *BSTR) HRESULT
}

type ITypeLibRegistrationVtbl struct {
	IUnknownVtbl
	GetGuid        uintptr
	GetVersion     uintptr
	GetLcid        uintptr
	GetWin32Path   uintptr
	GetWin64Path   uintptr
	GetDisplayName uintptr
	GetFlags       uintptr
	GetHelpDir     uintptr
}

type ITypeLibRegistration struct {
	IUnknown
}

func (this *ITypeLibRegistration) Vtbl() *ITypeLibRegistrationVtbl {
	return (*ITypeLibRegistrationVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *ITypeLibRegistration) GetGuid(pGuid *syscall.GUID) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetGuid, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pGuid)))
	return HRESULT(ret)
}

func (this *ITypeLibRegistration) GetVersion(pVersion *BSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetVersion, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pVersion)))
	return HRESULT(ret)
}

func (this *ITypeLibRegistration) GetLcid(pLcid *uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetLcid, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pLcid)))
	return HRESULT(ret)
}

func (this *ITypeLibRegistration) GetWin32Path(pWin32Path *BSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetWin32Path, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pWin32Path)))
	return HRESULT(ret)
}

func (this *ITypeLibRegistration) GetWin64Path(pWin64Path *BSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetWin64Path, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pWin64Path)))
	return HRESULT(ret)
}

func (this *ITypeLibRegistration) GetDisplayName(pDisplayName *BSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetDisplayName, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pDisplayName)))
	return HRESULT(ret)
}

func (this *ITypeLibRegistration) GetFlags(pFlags *uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetFlags, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pFlags)))
	return HRESULT(ret)
}

func (this *ITypeLibRegistration) GetHelpDir(pHelpDir *BSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetHelpDir, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pHelpDir)))
	return HRESULT(ret)
}

// B196B287-BAB4-101A-B69C-00AA00341D07
var IID_IEnumConnections = syscall.GUID{0xB196B287, 0xBAB4, 0x101A,
	[8]byte{0xB6, 0x9C, 0x00, 0xAA, 0x00, 0x34, 0x1D, 0x07}}

type IEnumConnectionsInterface interface {
	IUnknownInterface
	Next(cConnections uint32, rgcd *CONNECTDATA, pcFetched *uint32) HRESULT
	Skip(cConnections uint32) HRESULT
	Reset() HRESULT
	Clone(ppEnum **IEnumConnections) HRESULT
}

type IEnumConnectionsVtbl struct {
	IUnknownVtbl
	Next  uintptr
	Skip  uintptr
	Reset uintptr
	Clone uintptr
}

type IEnumConnections struct {
	IUnknown
}

func (this *IEnumConnections) Vtbl() *IEnumConnectionsVtbl {
	return (*IEnumConnectionsVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IEnumConnections) Next(cConnections uint32, rgcd *CONNECTDATA, pcFetched *uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Next, uintptr(unsafe.Pointer(this)), uintptr(cConnections), uintptr(unsafe.Pointer(rgcd)), uintptr(unsafe.Pointer(pcFetched)))
	return HRESULT(ret)
}

func (this *IEnumConnections) Skip(cConnections uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Skip, uintptr(unsafe.Pointer(this)), uintptr(cConnections))
	return HRESULT(ret)
}

func (this *IEnumConnections) Reset() HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Reset, uintptr(unsafe.Pointer(this)))
	return HRESULT(ret)
}

func (this *IEnumConnections) Clone(ppEnum **IEnumConnections) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Clone, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(ppEnum)))
	return HRESULT(ret)
}

// B196B286-BAB4-101A-B69C-00AA00341D07
var IID_IConnectionPoint = syscall.GUID{0xB196B286, 0xBAB4, 0x101A,
	[8]byte{0xB6, 0x9C, 0x00, 0xAA, 0x00, 0x34, 0x1D, 0x07}}

type IConnectionPointInterface interface {
	IUnknownInterface
	GetConnectionInterface(pIID *syscall.GUID) HRESULT
	GetConnectionPointContainer(ppCPC **IConnectionPointContainer) HRESULT
	Advise(pUnkSink *IUnknown, pdwCookie *uint32) HRESULT
	Unadvise(dwCookie uint32) HRESULT
	EnumConnections(ppEnum **IEnumConnections) HRESULT
}

type IConnectionPointVtbl struct {
	IUnknownVtbl
	GetConnectionInterface      uintptr
	GetConnectionPointContainer uintptr
	Advise                      uintptr
	Unadvise                    uintptr
	EnumConnections             uintptr
}

type IConnectionPoint struct {
	IUnknown
}

func (this *IConnectionPoint) Vtbl() *IConnectionPointVtbl {
	return (*IConnectionPointVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IConnectionPoint) GetConnectionInterface(pIID *syscall.GUID) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetConnectionInterface, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pIID)))
	return HRESULT(ret)
}

func (this *IConnectionPoint) GetConnectionPointContainer(ppCPC **IConnectionPointContainer) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetConnectionPointContainer, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(ppCPC)))
	return HRESULT(ret)
}

func (this *IConnectionPoint) Advise(pUnkSink *IUnknown, pdwCookie *uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Advise, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pUnkSink)), uintptr(unsafe.Pointer(pdwCookie)))
	return HRESULT(ret)
}

func (this *IConnectionPoint) Unadvise(dwCookie uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Unadvise, uintptr(unsafe.Pointer(this)), uintptr(dwCookie))
	return HRESULT(ret)
}

func (this *IConnectionPoint) EnumConnections(ppEnum **IEnumConnections) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().EnumConnections, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(ppEnum)))
	return HRESULT(ret)
}

// B196B285-BAB4-101A-B69C-00AA00341D07
var IID_IEnumConnectionPoints = syscall.GUID{0xB196B285, 0xBAB4, 0x101A,
	[8]byte{0xB6, 0x9C, 0x00, 0xAA, 0x00, 0x34, 0x1D, 0x07}}

type IEnumConnectionPointsInterface interface {
	IUnknownInterface
	Next(cConnections uint32, ppCP **IConnectionPoint, pcFetched *uint32) HRESULT
	Skip(cConnections uint32) HRESULT
	Reset() HRESULT
	Clone(ppEnum **IEnumConnectionPoints) HRESULT
}

type IEnumConnectionPointsVtbl struct {
	IUnknownVtbl
	Next  uintptr
	Skip  uintptr
	Reset uintptr
	Clone uintptr
}

type IEnumConnectionPoints struct {
	IUnknown
}

func (this *IEnumConnectionPoints) Vtbl() *IEnumConnectionPointsVtbl {
	return (*IEnumConnectionPointsVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IEnumConnectionPoints) Next(cConnections uint32, ppCP **IConnectionPoint, pcFetched *uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Next, uintptr(unsafe.Pointer(this)), uintptr(cConnections), uintptr(unsafe.Pointer(ppCP)), uintptr(unsafe.Pointer(pcFetched)))
	return HRESULT(ret)
}

func (this *IEnumConnectionPoints) Skip(cConnections uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Skip, uintptr(unsafe.Pointer(this)), uintptr(cConnections))
	return HRESULT(ret)
}

func (this *IEnumConnectionPoints) Reset() HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Reset, uintptr(unsafe.Pointer(this)))
	return HRESULT(ret)
}

func (this *IEnumConnectionPoints) Clone(ppEnum **IEnumConnectionPoints) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Clone, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(ppEnum)))
	return HRESULT(ret)
}

// B196B284-BAB4-101A-B69C-00AA00341D07
var IID_IConnectionPointContainer = syscall.GUID{0xB196B284, 0xBAB4, 0x101A,
	[8]byte{0xB6, 0x9C, 0x00, 0xAA, 0x00, 0x34, 0x1D, 0x07}}

type IConnectionPointContainerInterface interface {
	IUnknownInterface
	EnumConnectionPoints(ppEnum **IEnumConnectionPoints) HRESULT
	FindConnectionPoint(riid *syscall.GUID, ppCP **IConnectionPoint) HRESULT
}

type IConnectionPointContainerVtbl struct {
	IUnknownVtbl
	EnumConnectionPoints uintptr
	FindConnectionPoint  uintptr
}

type IConnectionPointContainer struct {
	IUnknown
}

func (this *IConnectionPointContainer) Vtbl() *IConnectionPointContainerVtbl {
	return (*IConnectionPointContainerVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IConnectionPointContainer) EnumConnectionPoints(ppEnum **IEnumConnectionPoints) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().EnumConnectionPoints, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(ppEnum)))
	return HRESULT(ret)
}

func (this *IConnectionPointContainer) FindConnectionPoint(riid *syscall.GUID, ppCP **IConnectionPoint) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().FindConnectionPoint, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(riid)), uintptr(unsafe.Pointer(ppCP)))
	return HRESULT(ret)
}

// BD1AE5E0-A6AE-11CE-BD37-504200C10000
var IID_IPersistMemory = syscall.GUID{0xBD1AE5E0, 0xA6AE, 0x11CE,
	[8]byte{0xBD, 0x37, 0x50, 0x42, 0x00, 0xC1, 0x00, 0x00}}

type IPersistMemoryInterface interface {
	IPersistInterface
	IsDirty() HRESULT
	Load(pMem unsafe.Pointer, cbSize uint32) HRESULT
	Save(pMem unsafe.Pointer, fClearDirty BOOL, cbSize uint32) HRESULT
	GetSizeMax(pCbSize *uint32) HRESULT
	InitNew() HRESULT
}

type IPersistMemoryVtbl struct {
	IPersistVtbl
	IsDirty    uintptr
	Load       uintptr
	Save       uintptr
	GetSizeMax uintptr
	InitNew    uintptr
}

type IPersistMemory struct {
	IPersist
}

func (this *IPersistMemory) Vtbl() *IPersistMemoryVtbl {
	return (*IPersistMemoryVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IPersistMemory) IsDirty() HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().IsDirty, uintptr(unsafe.Pointer(this)))
	return HRESULT(ret)
}

func (this *IPersistMemory) Load(pMem unsafe.Pointer, cbSize uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Load, uintptr(unsafe.Pointer(this)), uintptr(pMem), uintptr(cbSize))
	return HRESULT(ret)
}

func (this *IPersistMemory) Save(pMem unsafe.Pointer, fClearDirty BOOL, cbSize uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Save, uintptr(unsafe.Pointer(this)), uintptr(pMem), uintptr(fClearDirty), uintptr(cbSize))
	return HRESULT(ret)
}

func (this *IPersistMemory) GetSizeMax(pCbSize *uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetSizeMax, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pCbSize)))
	return HRESULT(ret)
}

func (this *IPersistMemory) InitNew() HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().InitNew, uintptr(unsafe.Pointer(this)))
	return HRESULT(ret)
}

// 7FD52380-4E07-101B-AE2D-08002B2EC713
var IID_IPersistStreamInit = syscall.GUID{0x7FD52380, 0x4E07, 0x101B,
	[8]byte{0xAE, 0x2D, 0x08, 0x00, 0x2B, 0x2E, 0xC7, 0x13}}

type IPersistStreamInitInterface interface {
	IPersistInterface
	IsDirty() HRESULT
	Load(pStm *IStream) HRESULT
	Save(pStm *IStream, fClearDirty BOOL) HRESULT
	GetSizeMax(pCbSize *uint64) HRESULT
	InitNew() HRESULT
}

type IPersistStreamInitVtbl struct {
	IPersistVtbl
	IsDirty    uintptr
	Load       uintptr
	Save       uintptr
	GetSizeMax uintptr
	InitNew    uintptr
}

type IPersistStreamInit struct {
	IPersist
}

func (this *IPersistStreamInit) Vtbl() *IPersistStreamInitVtbl {
	return (*IPersistStreamInitVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IPersistStreamInit) IsDirty() HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().IsDirty, uintptr(unsafe.Pointer(this)))
	return HRESULT(ret)
}

func (this *IPersistStreamInit) Load(pStm *IStream) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Load, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pStm)))
	return HRESULT(ret)
}

func (this *IPersistStreamInit) Save(pStm *IStream, fClearDirty BOOL) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Save, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pStm)), uintptr(fClearDirty))
	return HRESULT(ret)
}

func (this *IPersistStreamInit) GetSizeMax(pCbSize *uint64) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetSizeMax, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pCbSize)))
	return HRESULT(ret)
}

func (this *IPersistStreamInit) InitNew() HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().InitNew, uintptr(unsafe.Pointer(this)))
	return HRESULT(ret)
}

var (
	pCoBuildVersion                    uintptr
	pCoInitialize                      uintptr
	pCoRegisterMallocSpy               uintptr
	pCoRevokeMallocSpy                 uintptr
	pCoRegisterInitializeSpy           uintptr
	pCoRevokeInitializeSpy             uintptr
	pCoGetSystemSecurityPermissions    uintptr
	pCoLoadLibrary                     uintptr
	pCoFreeLibrary                     uintptr
	pCoFreeAllLibraries                uintptr
	pCoAllowSetForegroundWindow        uintptr
	pDcomChannelSetHResult             uintptr
	pCoIsOle1Class                     uintptr
	pCLSIDFromProgIDEx                 uintptr
	pCoFileTimeToDosDateTime           uintptr
	pCoDosDateTimeToFileTime           uintptr
	pCoFileTimeNow                     uintptr
	pCoRegisterChannelHook             uintptr
	pCoTreatAsClass                    uintptr
	pCreateDataAdviseHolder            uintptr
	pCreateDataCache                   uintptr
	pCoInstall                         uintptr
	pBindMoniker                       uintptr
	pCoGetObject                       uintptr
	pMkParseDisplayName                uintptr
	pMonikerRelativePathTo             uintptr
	pMonikerCommonPrefixWith           uintptr
	pCreateBindCtx                     uintptr
	pCreateGenericComposite            uintptr
	pGetClassFile                      uintptr
	pCreateClassMoniker                uintptr
	pCreateFileMoniker                 uintptr
	pCreateItemMoniker                 uintptr
	pCreateAntiMoniker                 uintptr
	pCreatePointerMoniker              uintptr
	pCreateObjrefMoniker               uintptr
	pGetRunningObjectTable             uintptr
	pCreateStdProgressIndicator        uintptr
	pCoGetMalloc                       uintptr
	pCoUninitialize                    uintptr
	pCoGetCurrentProcess               uintptr
	pCoInitializeEx                    uintptr
	pCoGetCallerTID                    uintptr
	pCoGetCurrentLogicalThreadId       uintptr
	pCoGetContextToken                 uintptr
	pCoGetApartmentType                uintptr
	pCoIncrementMTAUsage               uintptr
	pCoDecrementMTAUsage               uintptr
	pCoAllowUnmarshalerCLSID           uintptr
	pCoGetObjectContext                uintptr
	pCoGetClassObject                  uintptr
	pCoRegisterClassObject             uintptr
	pCoRevokeClassObject               uintptr
	pCoResumeClassObjects              uintptr
	pCoSuspendClassObjects             uintptr
	pCoAddRefServerProcess             uintptr
	pCoReleaseServerProcess            uintptr
	pCoGetPSClsid                      uintptr
	pCoRegisterPSClsid                 uintptr
	pCoRegisterSurrogate               uintptr
	pCoDisconnectObject                uintptr
	pCoLockObjectExternal              uintptr
	pCoIsHandlerConnected              uintptr
	pCoCreateFreeThreadedMarshaler     uintptr
	pCoFreeUnusedLibraries             uintptr
	pCoFreeUnusedLibrariesEx           uintptr
	pCoDisconnectContext               uintptr
	pCoInitializeSecurity              uintptr
	pCoGetCallContext                  uintptr
	pCoQueryProxyBlanket               uintptr
	pCoSetProxyBlanket                 uintptr
	pCoCopyProxy                       uintptr
	pCoQueryClientBlanket              uintptr
	pCoImpersonateClient               uintptr
	pCoRevertToSelf                    uintptr
	pCoQueryAuthenticationServices     uintptr
	pCoSwitchCallContext               uintptr
	pCoCreateInstance                  uintptr
	pCoCreateInstanceEx                uintptr
	pCoCreateInstanceFromApp           uintptr
	pCoRegisterActivationFilter        uintptr
	pCoGetCancelObject                 uintptr
	pCoSetCancelObject                 uintptr
	pCoCancelCall                      uintptr
	pCoTestCancel                      uintptr
	pCoEnableCallCancellation          uintptr
	pCoDisableCallCancellation         uintptr
	pStringFromCLSID                   uintptr
	pCLSIDFromString                   uintptr
	pStringFromIID                     uintptr
	pIIDFromString                     uintptr
	pProgIDFromCLSID                   uintptr
	pCLSIDFromProgID                   uintptr
	pStringFromGUID2                   uintptr
	pCoCreateGuid                      uintptr
	pCoWaitForMultipleHandles          uintptr
	pCoWaitForMultipleObjects          uintptr
	pCoGetTreatAsClass                 uintptr
	pCoInvalidateRemoteMachineBindings uintptr
	pCoTaskMemAlloc                    uintptr
	pCoTaskMemRealloc                  uintptr
	pCoTaskMemFree                     uintptr
	pCoRegisterDeviceCatalog           uintptr
	pCoRevokeDeviceCatalog             uintptr
	pSetErrorInfo                      uintptr
	pGetErrorInfo                      uintptr
)

func CoBuildVersion() uint32 {
	addr := lazyAddr(&pCoBuildVersion, libOle32, "CoBuildVersion")
	ret, _, _ := syscall.SyscallN(addr)
	return uint32(ret)
}

func CoInitialize(pvReserved unsafe.Pointer) HRESULT {
	addr := lazyAddr(&pCoInitialize, libOle32, "CoInitialize")
	ret, _, _ := syscall.SyscallN(addr, uintptr(pvReserved))
	return HRESULT(ret)
}

func CoRegisterMallocSpy(pMallocSpy *IMallocSpy) HRESULT {
	addr := lazyAddr(&pCoRegisterMallocSpy, libOle32, "CoRegisterMallocSpy")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pMallocSpy)))
	return HRESULT(ret)
}

func CoRevokeMallocSpy() HRESULT {
	addr := lazyAddr(&pCoRevokeMallocSpy, libOle32, "CoRevokeMallocSpy")
	ret, _, _ := syscall.SyscallN(addr)
	return HRESULT(ret)
}

func CoRegisterInitializeSpy(pSpy *IInitializeSpy, puliCookie *uint64) HRESULT {
	addr := lazyAddr(&pCoRegisterInitializeSpy, libOle32, "CoRegisterInitializeSpy")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pSpy)), uintptr(unsafe.Pointer(puliCookie)))
	return HRESULT(ret)
}

func CoRevokeInitializeSpy(uliCookie uint64) HRESULT {
	addr := lazyAddr(&pCoRevokeInitializeSpy, libOle32, "CoRevokeInitializeSpy")
	ret, _, _ := syscall.SyscallN(addr, uintptr(uliCookie))
	return HRESULT(ret)
}

func CoGetSystemSecurityPermissions(comSDType COMSD, ppSD *PSECURITY_DESCRIPTOR) HRESULT {
	addr := lazyAddr(&pCoGetSystemSecurityPermissions, libOle32, "CoGetSystemSecurityPermissions")
	ret, _, _ := syscall.SyscallN(addr, uintptr(comSDType), uintptr(unsafe.Pointer(ppSD)))
	return HRESULT(ret)
}

func CoLoadLibrary(lpszLibName PWSTR, bAutoFree BOOL) HINSTANCE {
	addr := lazyAddr(&pCoLoadLibrary, libOle32, "CoLoadLibrary")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpszLibName)), uintptr(bAutoFree))
	return ret
}

func CoFreeLibrary(hInst HINSTANCE) {
	addr := lazyAddr(&pCoFreeLibrary, libOle32, "CoFreeLibrary")
	syscall.SyscallN(addr, hInst)
}

func CoFreeAllLibraries() {
	addr := lazyAddr(&pCoFreeAllLibraries, libOle32, "CoFreeAllLibraries")
	syscall.SyscallN(addr)
}

func CoAllowSetForegroundWindow(pUnk *IUnknown, lpvReserved unsafe.Pointer) HRESULT {
	addr := lazyAddr(&pCoAllowSetForegroundWindow, libOle32, "CoAllowSetForegroundWindow")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pUnk)), uintptr(lpvReserved))
	return HRESULT(ret)
}

func DcomChannelSetHResult(pvReserved unsafe.Pointer, pulReserved *uint32, appsHR HRESULT) HRESULT {
	addr := lazyAddr(&pDcomChannelSetHResult, libOle32, "DcomChannelSetHResult")
	ret, _, _ := syscall.SyscallN(addr, uintptr(pvReserved), uintptr(unsafe.Pointer(pulReserved)), uintptr(appsHR))
	return HRESULT(ret)
}

func CoIsOle1Class(rclsid *syscall.GUID) BOOL {
	addr := lazyAddr(&pCoIsOle1Class, libOle32, "CoIsOle1Class")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(rclsid)))
	return BOOL(ret)
}

func CLSIDFromProgIDEx(lpszProgID PWSTR, lpclsid *syscall.GUID) HRESULT {
	addr := lazyAddr(&pCLSIDFromProgIDEx, libOle32, "CLSIDFromProgIDEx")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpszProgID)), uintptr(unsafe.Pointer(lpclsid)))
	return HRESULT(ret)
}

func CoFileTimeToDosDateTime(lpFileTime *FILETIME, lpDosDate *uint16, lpDosTime *uint16) BOOL {
	addr := lazyAddr(&pCoFileTimeToDosDateTime, libOle32, "CoFileTimeToDosDateTime")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpFileTime)), uintptr(unsafe.Pointer(lpDosDate)), uintptr(unsafe.Pointer(lpDosTime)))
	return BOOL(ret)
}

func CoDosDateTimeToFileTime(nDosDate uint16, nDosTime uint16, lpFileTime *FILETIME) BOOL {
	addr := lazyAddr(&pCoDosDateTimeToFileTime, libOle32, "CoDosDateTimeToFileTime")
	ret, _, _ := syscall.SyscallN(addr, uintptr(nDosDate), uintptr(nDosTime), uintptr(unsafe.Pointer(lpFileTime)))
	return BOOL(ret)
}

func CoFileTimeNow(lpFileTime *FILETIME) HRESULT {
	addr := lazyAddr(&pCoFileTimeNow, libOle32, "CoFileTimeNow")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpFileTime)))
	return HRESULT(ret)
}

func CoRegisterChannelHook(ExtensionUuid *syscall.GUID, pChannelHook *IChannelHook) HRESULT {
	addr := lazyAddr(&pCoRegisterChannelHook, libOle32, "CoRegisterChannelHook")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(ExtensionUuid)), uintptr(unsafe.Pointer(pChannelHook)))
	return HRESULT(ret)
}

func CoTreatAsClass(clsidOld *syscall.GUID, clsidNew *syscall.GUID) HRESULT {
	addr := lazyAddr(&pCoTreatAsClass, libOle32, "CoTreatAsClass")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(clsidOld)), uintptr(unsafe.Pointer(clsidNew)))
	return HRESULT(ret)
}

func CreateDataAdviseHolder(ppDAHolder **IDataAdviseHolder) HRESULT {
	addr := lazyAddr(&pCreateDataAdviseHolder, libOle32, "CreateDataAdviseHolder")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(ppDAHolder)))
	return HRESULT(ret)
}

func CreateDataCache(pUnkOuter *IUnknown, rclsid *syscall.GUID, iid *syscall.GUID, ppv unsafe.Pointer) HRESULT {
	addr := lazyAddr(&pCreateDataCache, libOle32, "CreateDataCache")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pUnkOuter)), uintptr(unsafe.Pointer(rclsid)), uintptr(unsafe.Pointer(iid)), uintptr(ppv))
	return HRESULT(ret)
}

func CoInstall(pbc *IBindCtx, dwFlags uint32, pClassSpec *UCLSSPEC, pQuery *QUERYCONTEXT, pszCodeBase PWSTR) HRESULT {
	addr := lazyAddr(&pCoInstall, libOle32, "CoInstall")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pbc)), uintptr(dwFlags), uintptr(unsafe.Pointer(pClassSpec)), uintptr(unsafe.Pointer(pQuery)), uintptr(unsafe.Pointer(pszCodeBase)))
	return HRESULT(ret)
}

func BindMoniker(pmk *IMoniker, grfOpt uint32, iidResult *syscall.GUID, ppvResult unsafe.Pointer) HRESULT {
	addr := lazyAddr(&pBindMoniker, libOle32, "BindMoniker")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pmk)), uintptr(grfOpt), uintptr(unsafe.Pointer(iidResult)), uintptr(ppvResult))
	return HRESULT(ret)
}

func CoGetObject(pszName PWSTR, pBindOptions *BIND_OPTS, riid *syscall.GUID, ppv unsafe.Pointer) HRESULT {
	addr := lazyAddr(&pCoGetObject, libOle32, "CoGetObject")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pszName)), uintptr(unsafe.Pointer(pBindOptions)), uintptr(unsafe.Pointer(riid)), uintptr(ppv))
	return HRESULT(ret)
}

func MkParseDisplayName(pbc *IBindCtx, szUserName PWSTR, pchEaten *uint32, ppmk **IMoniker) HRESULT {
	addr := lazyAddr(&pMkParseDisplayName, libOle32, "MkParseDisplayName")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pbc)), uintptr(unsafe.Pointer(szUserName)), uintptr(unsafe.Pointer(pchEaten)), uintptr(unsafe.Pointer(ppmk)))
	return HRESULT(ret)
}

func MonikerRelativePathTo(pmkSrc *IMoniker, pmkDest *IMoniker, ppmkRelPath **IMoniker, dwReserved BOOL) HRESULT {
	addr := lazyAddr(&pMonikerRelativePathTo, libOle32, "MonikerRelativePathTo")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pmkSrc)), uintptr(unsafe.Pointer(pmkDest)), uintptr(unsafe.Pointer(ppmkRelPath)), uintptr(dwReserved))
	return HRESULT(ret)
}

func MonikerCommonPrefixWith(pmkThis *IMoniker, pmkOther *IMoniker, ppmkCommon **IMoniker) HRESULT {
	addr := lazyAddr(&pMonikerCommonPrefixWith, libOle32, "MonikerCommonPrefixWith")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pmkThis)), uintptr(unsafe.Pointer(pmkOther)), uintptr(unsafe.Pointer(ppmkCommon)))
	return HRESULT(ret)
}

func CreateBindCtx(reserved uint32, ppbc **IBindCtx) HRESULT {
	addr := lazyAddr(&pCreateBindCtx, libOle32, "CreateBindCtx")
	ret, _, _ := syscall.SyscallN(addr, uintptr(reserved), uintptr(unsafe.Pointer(ppbc)))
	return HRESULT(ret)
}

func CreateGenericComposite(pmkFirst *IMoniker, pmkRest *IMoniker, ppmkComposite **IMoniker) HRESULT {
	addr := lazyAddr(&pCreateGenericComposite, libOle32, "CreateGenericComposite")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pmkFirst)), uintptr(unsafe.Pointer(pmkRest)), uintptr(unsafe.Pointer(ppmkComposite)))
	return HRESULT(ret)
}

func GetClassFile(szFilename PWSTR, pclsid *syscall.GUID) HRESULT {
	addr := lazyAddr(&pGetClassFile, libOle32, "GetClassFile")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(szFilename)), uintptr(unsafe.Pointer(pclsid)))
	return HRESULT(ret)
}

func CreateClassMoniker(rclsid *syscall.GUID, ppmk **IMoniker) HRESULT {
	addr := lazyAddr(&pCreateClassMoniker, libOle32, "CreateClassMoniker")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(rclsid)), uintptr(unsafe.Pointer(ppmk)))
	return HRESULT(ret)
}

func CreateFileMoniker(lpszPathName PWSTR, ppmk **IMoniker) HRESULT {
	addr := lazyAddr(&pCreateFileMoniker, libOle32, "CreateFileMoniker")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpszPathName)), uintptr(unsafe.Pointer(ppmk)))
	return HRESULT(ret)
}

func CreateItemMoniker(lpszDelim PWSTR, lpszItem PWSTR, ppmk **IMoniker) HRESULT {
	addr := lazyAddr(&pCreateItemMoniker, libOle32, "CreateItemMoniker")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpszDelim)), uintptr(unsafe.Pointer(lpszItem)), uintptr(unsafe.Pointer(ppmk)))
	return HRESULT(ret)
}

func CreateAntiMoniker(ppmk **IMoniker) HRESULT {
	addr := lazyAddr(&pCreateAntiMoniker, libOle32, "CreateAntiMoniker")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(ppmk)))
	return HRESULT(ret)
}

func CreatePointerMoniker(punk *IUnknown, ppmk **IMoniker) HRESULT {
	addr := lazyAddr(&pCreatePointerMoniker, libOle32, "CreatePointerMoniker")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(punk)), uintptr(unsafe.Pointer(ppmk)))
	return HRESULT(ret)
}

func CreateObjrefMoniker(punk *IUnknown, ppmk **IMoniker) HRESULT {
	addr := lazyAddr(&pCreateObjrefMoniker, libOle32, "CreateObjrefMoniker")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(punk)), uintptr(unsafe.Pointer(ppmk)))
	return HRESULT(ret)
}

func GetRunningObjectTable(reserved uint32, pprot **IRunningObjectTable) HRESULT {
	addr := lazyAddr(&pGetRunningObjectTable, libOle32, "GetRunningObjectTable")
	ret, _, _ := syscall.SyscallN(addr, uintptr(reserved), uintptr(unsafe.Pointer(pprot)))
	return HRESULT(ret)
}

func CreateStdProgressIndicator(hwndParent HWND, pszTitle PWSTR, pIbscCaller *IBindStatusCallback, ppIbsc **IBindStatusCallback) HRESULT {
	addr := lazyAddr(&pCreateStdProgressIndicator, libOle32, "CreateStdProgressIndicator")
	ret, _, _ := syscall.SyscallN(addr, hwndParent, uintptr(unsafe.Pointer(pszTitle)), uintptr(unsafe.Pointer(pIbscCaller)), uintptr(unsafe.Pointer(ppIbsc)))
	return HRESULT(ret)
}

func CoGetMalloc(dwMemContext uint32, ppMalloc **IMalloc) HRESULT {
	addr := lazyAddr(&pCoGetMalloc, libOle32, "CoGetMalloc")
	ret, _, _ := syscall.SyscallN(addr, uintptr(dwMemContext), uintptr(unsafe.Pointer(ppMalloc)))
	return HRESULT(ret)
}

func CoUninitialize() {
	addr := lazyAddr(&pCoUninitialize, libOle32, "CoUninitialize")
	syscall.SyscallN(addr)
}

func CoGetCurrentProcess() uint32 {
	addr := lazyAddr(&pCoGetCurrentProcess, libOle32, "CoGetCurrentProcess")
	ret, _, _ := syscall.SyscallN(addr)
	return uint32(ret)
}

func CoInitializeEx(pvReserved unsafe.Pointer, dwCoInit COINIT) HRESULT {
	addr := lazyAddr(&pCoInitializeEx, libOle32, "CoInitializeEx")
	ret, _, _ := syscall.SyscallN(addr, uintptr(pvReserved), uintptr(dwCoInit))
	return HRESULT(ret)
}

func CoGetCallerTID(lpdwTID *uint32) HRESULT {
	addr := lazyAddr(&pCoGetCallerTID, libOle32, "CoGetCallerTID")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpdwTID)))
	return HRESULT(ret)
}

func CoGetCurrentLogicalThreadId(pguid *syscall.GUID) HRESULT {
	addr := lazyAddr(&pCoGetCurrentLogicalThreadId, libOle32, "CoGetCurrentLogicalThreadId")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pguid)))
	return HRESULT(ret)
}

func CoGetContextToken(pToken *uintptr) HRESULT {
	addr := lazyAddr(&pCoGetContextToken, libOle32, "CoGetContextToken")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pToken)))
	return HRESULT(ret)
}

func CoGetApartmentType(pAptType *APTTYPE, pAptQualifier *APTTYPEQUALIFIER) HRESULT {
	addr := lazyAddr(&pCoGetApartmentType, libOle32, "CoGetApartmentType")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pAptType)), uintptr(unsafe.Pointer(pAptQualifier)))
	return HRESULT(ret)
}

func CoIncrementMTAUsage(pCookie *CO_MTA_USAGE_COOKIE) HRESULT {
	addr := lazyAddr(&pCoIncrementMTAUsage, libOle32, "CoIncrementMTAUsage")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pCookie)))
	return HRESULT(ret)
}

func CoDecrementMTAUsage(Cookie CO_MTA_USAGE_COOKIE) HRESULT {
	addr := lazyAddr(&pCoDecrementMTAUsage, libOle32, "CoDecrementMTAUsage")
	ret, _, _ := syscall.SyscallN(addr, Cookie)
	return HRESULT(ret)
}

func CoAllowUnmarshalerCLSID(clsid *syscall.GUID) HRESULT {
	addr := lazyAddr(&pCoAllowUnmarshalerCLSID, libOle32, "CoAllowUnmarshalerCLSID")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(clsid)))
	return HRESULT(ret)
}

func CoGetObjectContext(riid *syscall.GUID, ppv unsafe.Pointer) HRESULT {
	addr := lazyAddr(&pCoGetObjectContext, libOle32, "CoGetObjectContext")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(riid)), uintptr(ppv))
	return HRESULT(ret)
}

func CoGetClassObject(rclsid *syscall.GUID, dwClsContext CLSCTX, pvReserved unsafe.Pointer, riid *syscall.GUID, ppv unsafe.Pointer) HRESULT {
	addr := lazyAddr(&pCoGetClassObject, libOle32, "CoGetClassObject")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(rclsid)), uintptr(dwClsContext), uintptr(pvReserved), uintptr(unsafe.Pointer(riid)), uintptr(ppv))
	return HRESULT(ret)
}

func CoRegisterClassObject(rclsid *syscall.GUID, pUnk *IUnknown, dwClsContext CLSCTX, flags REGCLS, lpdwRegister *uint32) HRESULT {
	addr := lazyAddr(&pCoRegisterClassObject, libOle32, "CoRegisterClassObject")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(rclsid)), uintptr(unsafe.Pointer(pUnk)), uintptr(dwClsContext), uintptr(flags), uintptr(unsafe.Pointer(lpdwRegister)))
	return HRESULT(ret)
}

func CoRevokeClassObject(dwRegister uint32) HRESULT {
	addr := lazyAddr(&pCoRevokeClassObject, libOle32, "CoRevokeClassObject")
	ret, _, _ := syscall.SyscallN(addr, uintptr(dwRegister))
	return HRESULT(ret)
}

func CoResumeClassObjects() HRESULT {
	addr := lazyAddr(&pCoResumeClassObjects, libOle32, "CoResumeClassObjects")
	ret, _, _ := syscall.SyscallN(addr)
	return HRESULT(ret)
}

func CoSuspendClassObjects() HRESULT {
	addr := lazyAddr(&pCoSuspendClassObjects, libOle32, "CoSuspendClassObjects")
	ret, _, _ := syscall.SyscallN(addr)
	return HRESULT(ret)
}

func CoAddRefServerProcess() uint32 {
	addr := lazyAddr(&pCoAddRefServerProcess, libOle32, "CoAddRefServerProcess")
	ret, _, _ := syscall.SyscallN(addr)
	return uint32(ret)
}

func CoReleaseServerProcess() uint32 {
	addr := lazyAddr(&pCoReleaseServerProcess, libOle32, "CoReleaseServerProcess")
	ret, _, _ := syscall.SyscallN(addr)
	return uint32(ret)
}

func CoGetPSClsid(riid *syscall.GUID, pClsid *syscall.GUID) HRESULT {
	addr := lazyAddr(&pCoGetPSClsid, libOle32, "CoGetPSClsid")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(riid)), uintptr(unsafe.Pointer(pClsid)))
	return HRESULT(ret)
}

func CoRegisterPSClsid(riid *syscall.GUID, rclsid *syscall.GUID) HRESULT {
	addr := lazyAddr(&pCoRegisterPSClsid, libOle32, "CoRegisterPSClsid")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(riid)), uintptr(unsafe.Pointer(rclsid)))
	return HRESULT(ret)
}

func CoRegisterSurrogate(pSurrogate *ISurrogate) HRESULT {
	addr := lazyAddr(&pCoRegisterSurrogate, libOle32, "CoRegisterSurrogate")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pSurrogate)))
	return HRESULT(ret)
}

func CoDisconnectObject(pUnk *IUnknown, dwReserved uint32) HRESULT {
	addr := lazyAddr(&pCoDisconnectObject, libOle32, "CoDisconnectObject")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pUnk)), uintptr(dwReserved))
	return HRESULT(ret)
}

func CoLockObjectExternal(pUnk *IUnknown, fLock BOOL, fLastUnlockReleases BOOL) HRESULT {
	addr := lazyAddr(&pCoLockObjectExternal, libOle32, "CoLockObjectExternal")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pUnk)), uintptr(fLock), uintptr(fLastUnlockReleases))
	return HRESULT(ret)
}

func CoIsHandlerConnected(pUnk *IUnknown) BOOL {
	addr := lazyAddr(&pCoIsHandlerConnected, libOle32, "CoIsHandlerConnected")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pUnk)))
	return BOOL(ret)
}

func CoCreateFreeThreadedMarshaler(punkOuter *IUnknown, ppunkMarshal **IUnknown) HRESULT {
	addr := lazyAddr(&pCoCreateFreeThreadedMarshaler, libOle32, "CoCreateFreeThreadedMarshaler")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(punkOuter)), uintptr(unsafe.Pointer(ppunkMarshal)))
	return HRESULT(ret)
}

func CoFreeUnusedLibraries() {
	addr := lazyAddr(&pCoFreeUnusedLibraries, libOle32, "CoFreeUnusedLibraries")
	syscall.SyscallN(addr)
}

func CoFreeUnusedLibrariesEx(dwUnloadDelay uint32, dwReserved uint32) {
	addr := lazyAddr(&pCoFreeUnusedLibrariesEx, libOle32, "CoFreeUnusedLibrariesEx")
	syscall.SyscallN(addr, uintptr(dwUnloadDelay), uintptr(dwReserved))
}

func CoDisconnectContext(dwTimeout uint32) HRESULT {
	addr := lazyAddr(&pCoDisconnectContext, libOle32, "CoDisconnectContext")
	ret, _, _ := syscall.SyscallN(addr, uintptr(dwTimeout))
	return HRESULT(ret)
}

func CoInitializeSecurity(pSecDesc PSECURITY_DESCRIPTOR, cAuthSvc int32, asAuthSvc *SOLE_AUTHENTICATION_SERVICE, pReserved1 unsafe.Pointer, dwAuthnLevel RPC_C_AUTHN_LEVEL, dwImpLevel RPC_C_IMP_LEVEL, pAuthList unsafe.Pointer, dwCapabilities EOLE_AUTHENTICATION_CAPABILITIES, pReserved3 unsafe.Pointer) HRESULT {
	addr := lazyAddr(&pCoInitializeSecurity, libOle32, "CoInitializeSecurity")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pSecDesc)), uintptr(cAuthSvc), uintptr(unsafe.Pointer(asAuthSvc)), uintptr(pReserved1), uintptr(dwAuthnLevel), uintptr(dwImpLevel), uintptr(pAuthList), uintptr(dwCapabilities), uintptr(pReserved3))
	return HRESULT(ret)
}

func CoGetCallContext(riid *syscall.GUID, ppInterface unsafe.Pointer) HRESULT {
	addr := lazyAddr(&pCoGetCallContext, libOle32, "CoGetCallContext")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(riid)), uintptr(ppInterface))
	return HRESULT(ret)
}

func CoQueryProxyBlanket(pProxy *IUnknown, pwAuthnSvc *uint32, pAuthzSvc *uint32, pServerPrincName *PWSTR, pAuthnLevel *uint32, pImpLevel *uint32, pAuthInfo unsafe.Pointer, pCapabilites *uint32) HRESULT {
	addr := lazyAddr(&pCoQueryProxyBlanket, libOle32, "CoQueryProxyBlanket")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pProxy)), uintptr(unsafe.Pointer(pwAuthnSvc)), uintptr(unsafe.Pointer(pAuthzSvc)), uintptr(unsafe.Pointer(pServerPrincName)), uintptr(unsafe.Pointer(pAuthnLevel)), uintptr(unsafe.Pointer(pImpLevel)), uintptr(pAuthInfo), uintptr(unsafe.Pointer(pCapabilites)))
	return HRESULT(ret)
}

func CoSetProxyBlanket(pProxy *IUnknown, dwAuthnSvc uint32, dwAuthzSvc uint32, pServerPrincName PWSTR, dwAuthnLevel RPC_C_AUTHN_LEVEL, dwImpLevel RPC_C_IMP_LEVEL, pAuthInfo unsafe.Pointer, dwCapabilities EOLE_AUTHENTICATION_CAPABILITIES) HRESULT {
	addr := lazyAddr(&pCoSetProxyBlanket, libOle32, "CoSetProxyBlanket")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pProxy)), uintptr(dwAuthnSvc), uintptr(dwAuthzSvc), uintptr(unsafe.Pointer(pServerPrincName)), uintptr(dwAuthnLevel), uintptr(dwImpLevel), uintptr(pAuthInfo), uintptr(dwCapabilities))
	return HRESULT(ret)
}

func CoCopyProxy(pProxy *IUnknown, ppCopy **IUnknown) HRESULT {
	addr := lazyAddr(&pCoCopyProxy, libOle32, "CoCopyProxy")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pProxy)), uintptr(unsafe.Pointer(ppCopy)))
	return HRESULT(ret)
}

func CoQueryClientBlanket(pAuthnSvc *uint32, pAuthzSvc *uint32, pServerPrincName *PWSTR, pAuthnLevel *uint32, pImpLevel *uint32, pPrivs unsafe.Pointer, pCapabilities *uint32) HRESULT {
	addr := lazyAddr(&pCoQueryClientBlanket, libOle32, "CoQueryClientBlanket")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pAuthnSvc)), uintptr(unsafe.Pointer(pAuthzSvc)), uintptr(unsafe.Pointer(pServerPrincName)), uintptr(unsafe.Pointer(pAuthnLevel)), uintptr(unsafe.Pointer(pImpLevel)), uintptr(pPrivs), uintptr(unsafe.Pointer(pCapabilities)))
	return HRESULT(ret)
}

func CoImpersonateClient() HRESULT {
	addr := lazyAddr(&pCoImpersonateClient, libOle32, "CoImpersonateClient")
	ret, _, _ := syscall.SyscallN(addr)
	return HRESULT(ret)
}

func CoRevertToSelf() HRESULT {
	addr := lazyAddr(&pCoRevertToSelf, libOle32, "CoRevertToSelf")
	ret, _, _ := syscall.SyscallN(addr)
	return HRESULT(ret)
}

func CoQueryAuthenticationServices(pcAuthSvc *uint32, asAuthSvc **SOLE_AUTHENTICATION_SERVICE) HRESULT {
	addr := lazyAddr(&pCoQueryAuthenticationServices, libOle32, "CoQueryAuthenticationServices")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pcAuthSvc)), uintptr(unsafe.Pointer(asAuthSvc)))
	return HRESULT(ret)
}

func CoSwitchCallContext(pNewObject *IUnknown, ppOldObject **IUnknown) HRESULT {
	addr := lazyAddr(&pCoSwitchCallContext, libOle32, "CoSwitchCallContext")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pNewObject)), uintptr(unsafe.Pointer(ppOldObject)))
	return HRESULT(ret)
}

func CoCreateInstance(rclsid *syscall.GUID, pUnkOuter *IUnknown, dwClsContext CLSCTX, riid *syscall.GUID, ppv unsafe.Pointer) HRESULT {
	addr := lazyAddr(&pCoCreateInstance, libOle32, "CoCreateInstance")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(rclsid)), uintptr(unsafe.Pointer(pUnkOuter)), uintptr(dwClsContext), uintptr(unsafe.Pointer(riid)), uintptr(ppv))
	return HRESULT(ret)
}

func CoCreateInstanceEx(Clsid *syscall.GUID, punkOuter *IUnknown, dwClsCtx CLSCTX, pServerInfo *COSERVERINFO, dwCount uint32, pResults *MULTI_QI) HRESULT {
	addr := lazyAddr(&pCoCreateInstanceEx, libOle32, "CoCreateInstanceEx")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(Clsid)), uintptr(unsafe.Pointer(punkOuter)), uintptr(dwClsCtx), uintptr(unsafe.Pointer(pServerInfo)), uintptr(dwCount), uintptr(unsafe.Pointer(pResults)))
	return HRESULT(ret)
}

func CoCreateInstanceFromApp(Clsid *syscall.GUID, punkOuter *IUnknown, dwClsCtx CLSCTX, reserved unsafe.Pointer, dwCount uint32, pResults *MULTI_QI) HRESULT {
	addr := lazyAddr(&pCoCreateInstanceFromApp, libOle32, "CoCreateInstanceFromApp")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(Clsid)), uintptr(unsafe.Pointer(punkOuter)), uintptr(dwClsCtx), uintptr(reserved), uintptr(dwCount), uintptr(unsafe.Pointer(pResults)))
	return HRESULT(ret)
}

func CoRegisterActivationFilter(pActivationFilter *IActivationFilter) HRESULT {
	addr := lazyAddr(&pCoRegisterActivationFilter, libOle32, "CoRegisterActivationFilter")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pActivationFilter)))
	return HRESULT(ret)
}

func CoGetCancelObject(dwThreadId uint32, iid *syscall.GUID, ppUnk unsafe.Pointer) HRESULT {
	addr := lazyAddr(&pCoGetCancelObject, libOle32, "CoGetCancelObject")
	ret, _, _ := syscall.SyscallN(addr, uintptr(dwThreadId), uintptr(unsafe.Pointer(iid)), uintptr(ppUnk))
	return HRESULT(ret)
}

func CoSetCancelObject(pUnk *IUnknown) HRESULT {
	addr := lazyAddr(&pCoSetCancelObject, libOle32, "CoSetCancelObject")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pUnk)))
	return HRESULT(ret)
}

func CoCancelCall(dwThreadId uint32, ulTimeout uint32) HRESULT {
	addr := lazyAddr(&pCoCancelCall, libOle32, "CoCancelCall")
	ret, _, _ := syscall.SyscallN(addr, uintptr(dwThreadId), uintptr(ulTimeout))
	return HRESULT(ret)
}

func CoTestCancel() HRESULT {
	addr := lazyAddr(&pCoTestCancel, libOle32, "CoTestCancel")
	ret, _, _ := syscall.SyscallN(addr)
	return HRESULT(ret)
}

func CoEnableCallCancellation(pReserved unsafe.Pointer) HRESULT {
	addr := lazyAddr(&pCoEnableCallCancellation, libOle32, "CoEnableCallCancellation")
	ret, _, _ := syscall.SyscallN(addr, uintptr(pReserved))
	return HRESULT(ret)
}

func CoDisableCallCancellation(pReserved unsafe.Pointer) HRESULT {
	addr := lazyAddr(&pCoDisableCallCancellation, libOle32, "CoDisableCallCancellation")
	ret, _, _ := syscall.SyscallN(addr, uintptr(pReserved))
	return HRESULT(ret)
}

func StringFromCLSID(rclsid *syscall.GUID, lplpsz *PWSTR) HRESULT {
	addr := lazyAddr(&pStringFromCLSID, libOle32, "StringFromCLSID")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(rclsid)), uintptr(unsafe.Pointer(lplpsz)))
	return HRESULT(ret)
}

func CLSIDFromString(lpsz PWSTR, pclsid *syscall.GUID) HRESULT {
	addr := lazyAddr(&pCLSIDFromString, libOle32, "CLSIDFromString")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpsz)), uintptr(unsafe.Pointer(pclsid)))
	return HRESULT(ret)
}

func StringFromIID(rclsid *syscall.GUID, lplpsz *PWSTR) HRESULT {
	addr := lazyAddr(&pStringFromIID, libOle32, "StringFromIID")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(rclsid)), uintptr(unsafe.Pointer(lplpsz)))
	return HRESULT(ret)
}

func IIDFromString(lpsz PWSTR, lpiid *syscall.GUID) HRESULT {
	addr := lazyAddr(&pIIDFromString, libOle32, "IIDFromString")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpsz)), uintptr(unsafe.Pointer(lpiid)))
	return HRESULT(ret)
}

func ProgIDFromCLSID(clsid *syscall.GUID, lplpszProgID *PWSTR) HRESULT {
	addr := lazyAddr(&pProgIDFromCLSID, libOle32, "ProgIDFromCLSID")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(clsid)), uintptr(unsafe.Pointer(lplpszProgID)))
	return HRESULT(ret)
}

func CLSIDFromProgID(lpszProgID PWSTR, lpclsid *syscall.GUID) HRESULT {
	addr := lazyAddr(&pCLSIDFromProgID, libOle32, "CLSIDFromProgID")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpszProgID)), uintptr(unsafe.Pointer(lpclsid)))
	return HRESULT(ret)
}

func StringFromGUID2(rguid *syscall.GUID, lpsz PWSTR, cchMax int32) int32 {
	addr := lazyAddr(&pStringFromGUID2, libOle32, "StringFromGUID2")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(rguid)), uintptr(unsafe.Pointer(lpsz)), uintptr(cchMax))
	return int32(ret)
}

func CoCreateGuid(pguid *syscall.GUID) HRESULT {
	addr := lazyAddr(&pCoCreateGuid, libOle32, "CoCreateGuid")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pguid)))
	return HRESULT(ret)
}

func CoWaitForMultipleHandles(dwFlags uint32, dwTimeout uint32, cHandles uint32, pHandles *HANDLE, lpdwindex *uint32) HRESULT {
	addr := lazyAddr(&pCoWaitForMultipleHandles, libOle32, "CoWaitForMultipleHandles")
	ret, _, _ := syscall.SyscallN(addr, uintptr(dwFlags), uintptr(dwTimeout), uintptr(cHandles), uintptr(unsafe.Pointer(pHandles)), uintptr(unsafe.Pointer(lpdwindex)))
	return HRESULT(ret)
}

func CoWaitForMultipleObjects(dwFlags uint32, dwTimeout uint32, cHandles uint32, pHandles *HANDLE, lpdwindex *uint32) HRESULT {
	addr := lazyAddr(&pCoWaitForMultipleObjects, libOle32, "CoWaitForMultipleObjects")
	ret, _, _ := syscall.SyscallN(addr, uintptr(dwFlags), uintptr(dwTimeout), uintptr(cHandles), uintptr(unsafe.Pointer(pHandles)), uintptr(unsafe.Pointer(lpdwindex)))
	return HRESULT(ret)
}

func CoGetTreatAsClass(clsidOld *syscall.GUID, pClsidNew *syscall.GUID) HRESULT {
	addr := lazyAddr(&pCoGetTreatAsClass, libOle32, "CoGetTreatAsClass")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(clsidOld)), uintptr(unsafe.Pointer(pClsidNew)))
	return HRESULT(ret)
}

func CoInvalidateRemoteMachineBindings(pszMachineName PWSTR) HRESULT {
	addr := lazyAddr(&pCoInvalidateRemoteMachineBindings, libOle32, "CoInvalidateRemoteMachineBindings")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pszMachineName)))
	return HRESULT(ret)
}

func CoTaskMemAlloc(cb uintptr) unsafe.Pointer {
	addr := lazyAddr(&pCoTaskMemAlloc, libOle32, "CoTaskMemAlloc")
	ret, _, _ := syscall.SyscallN(addr, cb)
	return (unsafe.Pointer)(ret)
}

func CoTaskMemRealloc(pv unsafe.Pointer, cb uintptr) unsafe.Pointer {
	addr := lazyAddr(&pCoTaskMemRealloc, libOle32, "CoTaskMemRealloc")
	ret, _, _ := syscall.SyscallN(addr, uintptr(pv), cb)
	return (unsafe.Pointer)(ret)
}

func CoTaskMemFree(pv unsafe.Pointer) {
	addr := lazyAddr(&pCoTaskMemFree, libOle32, "CoTaskMemFree")
	syscall.SyscallN(addr, uintptr(pv))
}

func CoRegisterDeviceCatalog(deviceInstanceId PWSTR, cookie *CO_DEVICE_CATALOG_COOKIE) HRESULT {
	addr := lazyAddr(&pCoRegisterDeviceCatalog, libOle32, "CoRegisterDeviceCatalog")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(deviceInstanceId)), uintptr(unsafe.Pointer(cookie)))
	return HRESULT(ret)
}

func CoRevokeDeviceCatalog(cookie CO_DEVICE_CATALOG_COOKIE) HRESULT {
	addr := lazyAddr(&pCoRevokeDeviceCatalog, libOle32, "CoRevokeDeviceCatalog")
	ret, _, _ := syscall.SyscallN(addr, cookie)
	return HRESULT(ret)
}

func SetErrorInfo(dwReserved uint32, perrinfo *IErrorInfo) HRESULT {
	addr := lazyAddr(&pSetErrorInfo, libOleaut32, "SetErrorInfo")
	ret, _, _ := syscall.SyscallN(addr, uintptr(dwReserved), uintptr(unsafe.Pointer(perrinfo)))
	return HRESULT(ret)
}

func GetErrorInfo(dwReserved uint32, pperrinfo **IErrorInfo) HRESULT {
	addr := lazyAddr(&pGetErrorInfo, libOleaut32, "GetErrorInfo")
	ret, _, _ := syscall.SyscallN(addr, uintptr(dwReserved), uintptr(unsafe.Pointer(pperrinfo)))
	return HRESULT(ret)
}
