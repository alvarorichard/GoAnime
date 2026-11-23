package win32

import (
	"syscall"
	"unsafe"
)

const (
	PKEY_PIDSTR_MAX uint32 = 0xa
)

// enums

// enum
// flags
type GETPROPERTYSTOREFLAGS uint32

const (
	GPS_DEFAULT                 GETPROPERTYSTOREFLAGS = 0
	GPS_HANDLERPROPERTIESONLY   GETPROPERTYSTOREFLAGS = 1
	GPS_READWRITE               GETPROPERTYSTOREFLAGS = 2
	GPS_TEMPORARY               GETPROPERTYSTOREFLAGS = 4
	GPS_FASTPROPERTIESONLY      GETPROPERTYSTOREFLAGS = 8
	GPS_OPENSLOWITEM            GETPROPERTYSTOREFLAGS = 16
	GPS_DELAYCREATION           GETPROPERTYSTOREFLAGS = 32
	GPS_BESTEFFORT              GETPROPERTYSTOREFLAGS = 64
	GPS_NO_OPLOCK               GETPROPERTYSTOREFLAGS = 128
	GPS_PREFERQUERYPROPERTIES   GETPROPERTYSTOREFLAGS = 256
	GPS_EXTRINSICPROPERTIES     GETPROPERTYSTOREFLAGS = 512
	GPS_EXTRINSICPROPERTIESONLY GETPROPERTYSTOREFLAGS = 1024
	GPS_VOLATILEPROPERTIES      GETPROPERTYSTOREFLAGS = 2048
	GPS_VOLATILEPROPERTIESONLY  GETPROPERTYSTOREFLAGS = 4096
	GPS_MASK_VALID              GETPROPERTYSTOREFLAGS = 8191
)

// enum
// flags
type PKA_FLAGS uint32

const (
	PKA_SET    PKA_FLAGS = 0
	PKA_APPEND PKA_FLAGS = 1
	PKA_DELETE PKA_FLAGS = 2
)

// enum
type PSC_STATE int32

const (
	PSC_NORMAL      PSC_STATE = 0
	PSC_NOTINSOURCE PSC_STATE = 1
	PSC_DIRTY       PSC_STATE = 2
	PSC_READONLY    PSC_STATE = 3
)

// enum
type PROPENUMTYPE int32

const (
	PET_DISCRETEVALUE PROPENUMTYPE = 0
	PET_RANGEDVALUE   PROPENUMTYPE = 1
	PET_DEFAULTVALUE  PROPENUMTYPE = 2
	PET_ENDRANGE      PROPENUMTYPE = 3
)

// enum
// flags
type PROPDESC_TYPE_FLAGS uint32

const (
	PDTF_DEFAULT                   PROPDESC_TYPE_FLAGS = 0
	PDTF_MULTIPLEVALUES            PROPDESC_TYPE_FLAGS = 1
	PDTF_ISINNATE                  PROPDESC_TYPE_FLAGS = 2
	PDTF_ISGROUP                   PROPDESC_TYPE_FLAGS = 4
	PDTF_CANGROUPBY                PROPDESC_TYPE_FLAGS = 8
	PDTF_CANSTACKBY                PROPDESC_TYPE_FLAGS = 16
	PDTF_ISTREEPROPERTY            PROPDESC_TYPE_FLAGS = 32
	PDTF_INCLUDEINFULLTEXTQUERY    PROPDESC_TYPE_FLAGS = 64
	PDTF_ISVIEWABLE                PROPDESC_TYPE_FLAGS = 128
	PDTF_ISQUERYABLE               PROPDESC_TYPE_FLAGS = 256
	PDTF_CANBEPURGED               PROPDESC_TYPE_FLAGS = 512
	PDTF_SEARCHRAWVALUE            PROPDESC_TYPE_FLAGS = 1024
	PDTF_DONTCOERCEEMPTYSTRINGS    PROPDESC_TYPE_FLAGS = 2048
	PDTF_ALWAYSINSUPPLEMENTALSTORE PROPDESC_TYPE_FLAGS = 4096
	PDTF_ISSYSTEMPROPERTY          PROPDESC_TYPE_FLAGS = 2147483648
	PDTF_MASK_ALL                  PROPDESC_TYPE_FLAGS = 2147491839
)

// enum
// flags
type PROPDESC_VIEW_FLAGS uint32

const (
	PDVF_DEFAULT             PROPDESC_VIEW_FLAGS = 0
	PDVF_CENTERALIGN         PROPDESC_VIEW_FLAGS = 1
	PDVF_RIGHTALIGN          PROPDESC_VIEW_FLAGS = 2
	PDVF_BEGINNEWGROUP       PROPDESC_VIEW_FLAGS = 4
	PDVF_FILLAREA            PROPDESC_VIEW_FLAGS = 8
	PDVF_SORTDESCENDING      PROPDESC_VIEW_FLAGS = 16
	PDVF_SHOWONLYIFPRESENT   PROPDESC_VIEW_FLAGS = 32
	PDVF_SHOWBYDEFAULT       PROPDESC_VIEW_FLAGS = 64
	PDVF_SHOWINPRIMARYLIST   PROPDESC_VIEW_FLAGS = 128
	PDVF_SHOWINSECONDARYLIST PROPDESC_VIEW_FLAGS = 256
	PDVF_HIDELABEL           PROPDESC_VIEW_FLAGS = 512
	PDVF_HIDDEN              PROPDESC_VIEW_FLAGS = 2048
	PDVF_CANWRAP             PROPDESC_VIEW_FLAGS = 4096
	PDVF_MASK_ALL            PROPDESC_VIEW_FLAGS = 7167
)

// enum
type PROPDESC_DISPLAYTYPE int32

const (
	PDDT_STRING     PROPDESC_DISPLAYTYPE = 0
	PDDT_NUMBER     PROPDESC_DISPLAYTYPE = 1
	PDDT_BOOLEAN    PROPDESC_DISPLAYTYPE = 2
	PDDT_DATETIME   PROPDESC_DISPLAYTYPE = 3
	PDDT_ENUMERATED PROPDESC_DISPLAYTYPE = 4
)

// enum
type PROPDESC_GROUPING_RANGE int32

const (
	PDGR_DISCRETE     PROPDESC_GROUPING_RANGE = 0
	PDGR_ALPHANUMERIC PROPDESC_GROUPING_RANGE = 1
	PDGR_SIZE         PROPDESC_GROUPING_RANGE = 2
	PDGR_DYNAMIC      PROPDESC_GROUPING_RANGE = 3
	PDGR_DATE         PROPDESC_GROUPING_RANGE = 4
	PDGR_PERCENT      PROPDESC_GROUPING_RANGE = 5
	PDGR_ENUMERATED   PROPDESC_GROUPING_RANGE = 6
)

// enum
// flags
type PROPDESC_FORMAT_FLAGS uint32

const (
	PDFF_DEFAULT              PROPDESC_FORMAT_FLAGS = 0
	PDFF_PREFIXNAME           PROPDESC_FORMAT_FLAGS = 1
	PDFF_FILENAME             PROPDESC_FORMAT_FLAGS = 2
	PDFF_ALWAYSKB             PROPDESC_FORMAT_FLAGS = 4
	PDFF_RESERVED_RIGHTTOLEFT PROPDESC_FORMAT_FLAGS = 8
	PDFF_SHORTTIME            PROPDESC_FORMAT_FLAGS = 16
	PDFF_LONGTIME             PROPDESC_FORMAT_FLAGS = 32
	PDFF_HIDETIME             PROPDESC_FORMAT_FLAGS = 64
	PDFF_SHORTDATE            PROPDESC_FORMAT_FLAGS = 128
	PDFF_LONGDATE             PROPDESC_FORMAT_FLAGS = 256
	PDFF_HIDEDATE             PROPDESC_FORMAT_FLAGS = 512
	PDFF_RELATIVEDATE         PROPDESC_FORMAT_FLAGS = 1024
	PDFF_USEEDITINVITATION    PROPDESC_FORMAT_FLAGS = 2048
	PDFF_READONLY             PROPDESC_FORMAT_FLAGS = 4096
	PDFF_NOAUTOREADINGORDER   PROPDESC_FORMAT_FLAGS = 8192
)

// enum
type PROPDESC_SORTDESCRIPTION int32

const (
	PDSD_GENERAL          PROPDESC_SORTDESCRIPTION = 0
	PDSD_A_Z              PROPDESC_SORTDESCRIPTION = 1
	PDSD_LOWEST_HIGHEST   PROPDESC_SORTDESCRIPTION = 2
	PDSD_SMALLEST_BIGGEST PROPDESC_SORTDESCRIPTION = 3
	PDSD_OLDEST_NEWEST    PROPDESC_SORTDESCRIPTION = 4
)

// enum
type PROPDESC_RELATIVEDESCRIPTION_TYPE int32

const (
	PDRDT_GENERAL  PROPDESC_RELATIVEDESCRIPTION_TYPE = 0
	PDRDT_DATE     PROPDESC_RELATIVEDESCRIPTION_TYPE = 1
	PDRDT_SIZE     PROPDESC_RELATIVEDESCRIPTION_TYPE = 2
	PDRDT_COUNT    PROPDESC_RELATIVEDESCRIPTION_TYPE = 3
	PDRDT_REVISION PROPDESC_RELATIVEDESCRIPTION_TYPE = 4
	PDRDT_LENGTH   PROPDESC_RELATIVEDESCRIPTION_TYPE = 5
	PDRDT_DURATION PROPDESC_RELATIVEDESCRIPTION_TYPE = 6
	PDRDT_SPEED    PROPDESC_RELATIVEDESCRIPTION_TYPE = 7
	PDRDT_RATE     PROPDESC_RELATIVEDESCRIPTION_TYPE = 8
	PDRDT_RATING   PROPDESC_RELATIVEDESCRIPTION_TYPE = 9
	PDRDT_PRIORITY PROPDESC_RELATIVEDESCRIPTION_TYPE = 10
)

// enum
type PROPDESC_AGGREGATION_TYPE int32

const (
	PDAT_DEFAULT   PROPDESC_AGGREGATION_TYPE = 0
	PDAT_FIRST     PROPDESC_AGGREGATION_TYPE = 1
	PDAT_SUM       PROPDESC_AGGREGATION_TYPE = 2
	PDAT_AVERAGE   PROPDESC_AGGREGATION_TYPE = 3
	PDAT_DATERANGE PROPDESC_AGGREGATION_TYPE = 4
	PDAT_UNION     PROPDESC_AGGREGATION_TYPE = 5
	PDAT_MAX       PROPDESC_AGGREGATION_TYPE = 6
	PDAT_MIN       PROPDESC_AGGREGATION_TYPE = 7
)

// enum
type PROPDESC_CONDITION_TYPE int32

const (
	PDCOT_NONE     PROPDESC_CONDITION_TYPE = 0
	PDCOT_STRING   PROPDESC_CONDITION_TYPE = 1
	PDCOT_SIZE     PROPDESC_CONDITION_TYPE = 2
	PDCOT_DATETIME PROPDESC_CONDITION_TYPE = 3
	PDCOT_BOOLEAN  PROPDESC_CONDITION_TYPE = 4
	PDCOT_NUMBER   PROPDESC_CONDITION_TYPE = 5
)

// enum
// flags
type PROPDESC_SEARCHINFO_FLAGS uint32

const (
	PDSIF_DEFAULT         PROPDESC_SEARCHINFO_FLAGS = 0
	PDSIF_ININVERTEDINDEX PROPDESC_SEARCHINFO_FLAGS = 1
	PDSIF_ISCOLUMN        PROPDESC_SEARCHINFO_FLAGS = 2
	PDSIF_ISCOLUMNSPARSE  PROPDESC_SEARCHINFO_FLAGS = 4
	PDSIF_ALWAYSINCLUDE   PROPDESC_SEARCHINFO_FLAGS = 8
	PDSIF_USEFORTYPEAHEAD PROPDESC_SEARCHINFO_FLAGS = 16
)

// enum
type PROPDESC_COLUMNINDEX_TYPE int32

const (
	PDCIT_NONE         PROPDESC_COLUMNINDEX_TYPE = 0
	PDCIT_ONDISK       PROPDESC_COLUMNINDEX_TYPE = 1
	PDCIT_INMEMORY     PROPDESC_COLUMNINDEX_TYPE = 2
	PDCIT_ONDEMAND     PROPDESC_COLUMNINDEX_TYPE = 3
	PDCIT_ONDISKALL    PROPDESC_COLUMNINDEX_TYPE = 4
	PDCIT_ONDISKVECTOR PROPDESC_COLUMNINDEX_TYPE = 5
)

// enum
type PROPDESC_ENUMFILTER int32

const (
	PDEF_ALL             PROPDESC_ENUMFILTER = 0
	PDEF_SYSTEM          PROPDESC_ENUMFILTER = 1
	PDEF_NONSYSTEM       PROPDESC_ENUMFILTER = 2
	PDEF_VIEWABLE        PROPDESC_ENUMFILTER = 3
	PDEF_QUERYABLE       PROPDESC_ENUMFILTER = 4
	PDEF_INFULLTEXTQUERY PROPDESC_ENUMFILTER = 5
	PDEF_COLUMN          PROPDESC_ENUMFILTER = 6
)

// enum
type PERSIST_SPROPSTORE_FLAGS_ int32

const (
	FPSPS_DEFAULT                   PERSIST_SPROPSTORE_FLAGS_ = 0
	FPSPS_READONLY                  PERSIST_SPROPSTORE_FLAGS_ = 1
	FPSPS_TREAT_NEW_VALUES_AS_DIRTY PERSIST_SPROPSTORE_FLAGS_ = 2
)

// enum
// flags
type PSTIME_FLAGS uint32

const (
	PSTF_UTC   PSTIME_FLAGS = 0
	PSTF_LOCAL PSTIME_FLAGS = 1
)

// enum
type PROPVAR_COMPARE_UNIT int32

const (
	PVCU_DEFAULT PROPVAR_COMPARE_UNIT = 0
	PVCU_SECOND  PROPVAR_COMPARE_UNIT = 1
	PVCU_MINUTE  PROPVAR_COMPARE_UNIT = 2
	PVCU_HOUR    PROPVAR_COMPARE_UNIT = 3
	PVCU_DAY     PROPVAR_COMPARE_UNIT = 4
	PVCU_MONTH   PROPVAR_COMPARE_UNIT = 5
	PVCU_YEAR    PROPVAR_COMPARE_UNIT = 6
)

// enum
// flags
type PROPVAR_COMPARE_FLAGS uint32

const (
	PVCF_DEFAULT                       PROPVAR_COMPARE_FLAGS = 0
	PVCF_TREATEMPTYASGREATERTHAN       PROPVAR_COMPARE_FLAGS = 1
	PVCF_USESTRCMP                     PROPVAR_COMPARE_FLAGS = 2
	PVCF_USESTRCMPC                    PROPVAR_COMPARE_FLAGS = 4
	PVCF_USESTRCMPI                    PROPVAR_COMPARE_FLAGS = 8
	PVCF_USESTRCMPIC                   PROPVAR_COMPARE_FLAGS = 16
	PVCF_DIGITSASNUMBERS_CASESENSITIVE PROPVAR_COMPARE_FLAGS = 32
)

// enum
// flags
type PROPVAR_CHANGE_FLAGS uint32

const (
	PVCHF_DEFAULT        PROPVAR_CHANGE_FLAGS = 0
	PVCHF_NOVALUEPROP    PROPVAR_CHANGE_FLAGS = 1
	PVCHF_ALPHABOOL      PROPVAR_CHANGE_FLAGS = 2
	PVCHF_NOUSEROVERRIDE PROPVAR_CHANGE_FLAGS = 4
	PVCHF_LOCALBOOL      PROPVAR_CHANGE_FLAGS = 8
	PVCHF_NOHEXSTRING    PROPVAR_CHANGE_FLAGS = 16
)

// enum
// flags
type DRAWPROGRESSFLAGS uint32

const (
	DPF_NONE             DRAWPROGRESSFLAGS = 0
	DPF_MARQUEE          DRAWPROGRESSFLAGS = 1
	DPF_MARQUEE_COMPLETE DRAWPROGRESSFLAGS = 2
	DPF_ERROR            DRAWPROGRESSFLAGS = 4
	DPF_WARNING          DRAWPROGRESSFLAGS = 8
	DPF_STOPPED          DRAWPROGRESSFLAGS = 16
)

// enum
// flags
type SYNC_TRANSFER_STATUS uint32

const (
	STS_NONE                   SYNC_TRANSFER_STATUS = 0
	STS_NEEDSUPLOAD            SYNC_TRANSFER_STATUS = 1
	STS_NEEDSDOWNLOAD          SYNC_TRANSFER_STATUS = 2
	STS_TRANSFERRING           SYNC_TRANSFER_STATUS = 4
	STS_PAUSED                 SYNC_TRANSFER_STATUS = 8
	STS_HASERROR               SYNC_TRANSFER_STATUS = 16
	STS_FETCHING_METADATA      SYNC_TRANSFER_STATUS = 32
	STS_USER_REQUESTED_REFRESH SYNC_TRANSFER_STATUS = 64
	STS_HASWARNING             SYNC_TRANSFER_STATUS = 128
	STS_EXCLUDED               SYNC_TRANSFER_STATUS = 256
	STS_INCOMPLETE             SYNC_TRANSFER_STATUS = 512
	STS_PLACEHOLDER_IFEMPTY    SYNC_TRANSFER_STATUS = 1024
)

// enum
// flags
type PLACEHOLDER_STATES uint32

const (
	PS_NONE                            PLACEHOLDER_STATES = 0
	PS_MARKED_FOR_OFFLINE_AVAILABILITY PLACEHOLDER_STATES = 1
	PS_FULL_PRIMARY_STREAM_AVAILABLE   PLACEHOLDER_STATES = 2
	PS_CREATE_FILE_ACCESSIBLE          PLACEHOLDER_STATES = 4
	PS_CLOUDFILE_PLACEHOLDER           PLACEHOLDER_STATES = 8
	PS_DEFAULT                         PLACEHOLDER_STATES = 7
	PS_ALL                             PLACEHOLDER_STATES = 15
)

// enum
// flags
type PROPERTYUI_NAME_FLAGS uint32

const (
	PUIFNF_DEFAULT  PROPERTYUI_NAME_FLAGS = 0
	PUIFNF_MNEMONIC PROPERTYUI_NAME_FLAGS = 1
)

// enum
// flags
type PROPERTYUI_FLAGS uint32

const (
	PUIF_DEFAULT          PROPERTYUI_FLAGS = 0
	PUIF_RIGHTALIGN       PROPERTYUI_FLAGS = 1
	PUIF_NOLABELININFOTIP PROPERTYUI_FLAGS = 2
)

// enum
// flags
type PROPERTYUI_FORMAT_FLAGS uint32

const (
	PUIFFDF_DEFAULT      PROPERTYUI_FORMAT_FLAGS = 0
	PUIFFDF_RIGHTTOLEFT  PROPERTYUI_FORMAT_FLAGS = 1
	PUIFFDF_SHORTFORMAT  PROPERTYUI_FORMAT_FLAGS = 2
	PUIFFDF_NOTIME       PROPERTYUI_FORMAT_FLAGS = 4
	PUIFFDF_FRIENDLYDATE PROPERTYUI_FORMAT_FLAGS = 8
)

// enum
type PDOPSTATUS int32

const (
	PDOPS_RUNNING   PDOPSTATUS = 1
	PDOPS_PAUSED    PDOPSTATUS = 2
	PDOPS_CANCELLED PDOPSTATUS = 3
	PDOPS_STOPPED   PDOPSTATUS = 4
	PDOPS_ERRORS    PDOPSTATUS = 5
)

// enum
// flags
type SYNC_ENGINE_STATE_FLAGS uint32

const (
	SESF_NONE                          SYNC_ENGINE_STATE_FLAGS = 0
	SESF_SERVICE_QUOTA_NEARING_LIMIT   SYNC_ENGINE_STATE_FLAGS = 1
	SESF_SERVICE_QUOTA_EXCEEDED_LIMIT  SYNC_ENGINE_STATE_FLAGS = 2
	SESF_AUTHENTICATION_ERROR          SYNC_ENGINE_STATE_FLAGS = 4
	SESF_PAUSED_DUE_TO_METERED_NETWORK SYNC_ENGINE_STATE_FLAGS = 8
	SESF_PAUSED_DUE_TO_DISK_SPACE_FULL SYNC_ENGINE_STATE_FLAGS = 16
	SESF_PAUSED_DUE_TO_CLIENT_POLICY   SYNC_ENGINE_STATE_FLAGS = 32
	SESF_PAUSED_DUE_TO_SERVICE_POLICY  SYNC_ENGINE_STATE_FLAGS = 64
	SESF_SERVICE_UNAVAILABLE           SYNC_ENGINE_STATE_FLAGS = 128
	SESF_PAUSED_DUE_TO_USER_REQUEST    SYNC_ENGINE_STATE_FLAGS = 256
	SESF_ALL_FLAGS                     SYNC_ENGINE_STATE_FLAGS = 511
)

// structs

type PROPERTYKEY struct {
	Fmtid syscall.GUID
	Pid   uint32
}

type InMemoryPropertyStore struct {
}

type InMemoryPropertyStoreMarshalByValue struct {
}

type PropertySystem struct {
}

type SERIALIZEDPROPSTORAGE struct {
}

type PROPPRG struct {
	FlPrg           uint16
	FlPrgInit       uint16
	AchTitle        [30]CHAR
	AchCmdLine      [128]CHAR
	AchWorkDir      [64]CHAR
	WHotKey         uint16
	AchIconFile     [80]CHAR
	WIconIndex      uint16
	DwEnhModeFlags  uint32
	DwRealModeFlags uint32
	AchOtherFile    [80]CHAR
	AchPIFFile      [260]CHAR
}

// interfaces

// B7D14566-0509-4CCE-A71F-0A554233BD9B
var IID_IInitializeWithFile = syscall.GUID{0xB7D14566, 0x0509, 0x4CCE,
	[8]byte{0xA7, 0x1F, 0x0A, 0x55, 0x42, 0x33, 0xBD, 0x9B}}

type IInitializeWithFileInterface interface {
	IUnknownInterface
	Initialize(pszFilePath PWSTR, grfMode uint32) HRESULT
}

type IInitializeWithFileVtbl struct {
	IUnknownVtbl
	Initialize uintptr
}

type IInitializeWithFile struct {
	IUnknown
}

func (this *IInitializeWithFile) Vtbl() *IInitializeWithFileVtbl {
	return (*IInitializeWithFileVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IInitializeWithFile) Initialize(pszFilePath PWSTR, grfMode uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Initialize, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pszFilePath)), uintptr(grfMode))
	return HRESULT(ret)
}

// B824B49D-22AC-4161-AC8A-9916E8FA3F7F
var IID_IInitializeWithStream = syscall.GUID{0xB824B49D, 0x22AC, 0x4161,
	[8]byte{0xAC, 0x8A, 0x99, 0x16, 0xE8, 0xFA, 0x3F, 0x7F}}

type IInitializeWithStreamInterface interface {
	IUnknownInterface
	Initialize(pstream *IStream, grfMode uint32) HRESULT
}

type IInitializeWithStreamVtbl struct {
	IUnknownVtbl
	Initialize uintptr
}

type IInitializeWithStream struct {
	IUnknown
}

func (this *IInitializeWithStream) Vtbl() *IInitializeWithStreamVtbl {
	return (*IInitializeWithStreamVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IInitializeWithStream) Initialize(pstream *IStream, grfMode uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Initialize, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pstream)), uintptr(grfMode))
	return HRESULT(ret)
}

// 886D8EEB-8CF2-4446-8D02-CDBA1DBDCF99
var IID_IPropertyStore = syscall.GUID{0x886D8EEB, 0x8CF2, 0x4446,
	[8]byte{0x8D, 0x02, 0xCD, 0xBA, 0x1D, 0xBD, 0xCF, 0x99}}

type IPropertyStoreInterface interface {
	IUnknownInterface
	GetCount(cProps *uint32) HRESULT
	GetAt(iProp uint32, pkey *PROPERTYKEY) HRESULT
	GetValue(key *PROPERTYKEY, pv *PROPVARIANT) HRESULT
	SetValue(key *PROPERTYKEY, propvar *PROPVARIANT) HRESULT
	Commit() HRESULT
}

type IPropertyStoreVtbl struct {
	IUnknownVtbl
	GetCount uintptr
	GetAt    uintptr
	GetValue uintptr
	SetValue uintptr
	Commit   uintptr
}

type IPropertyStore struct {
	IUnknown
}

func (this *IPropertyStore) Vtbl() *IPropertyStoreVtbl {
	return (*IPropertyStoreVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IPropertyStore) GetCount(cProps *uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetCount, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(cProps)))
	return HRESULT(ret)
}

func (this *IPropertyStore) GetAt(iProp uint32, pkey *PROPERTYKEY) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetAt, uintptr(unsafe.Pointer(this)), uintptr(iProp), uintptr(unsafe.Pointer(pkey)))
	return HRESULT(ret)
}

func (this *IPropertyStore) GetValue(key *PROPERTYKEY, pv *PROPVARIANT) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetValue, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(key)), uintptr(unsafe.Pointer(pv)))
	return HRESULT(ret)
}

func (this *IPropertyStore) SetValue(key *PROPERTYKEY, propvar *PROPVARIANT) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetValue, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(key)), uintptr(unsafe.Pointer(propvar)))
	return HRESULT(ret)
}

func (this *IPropertyStore) Commit() HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Commit, uintptr(unsafe.Pointer(this)))
	return HRESULT(ret)
}

// 71604B0F-97B0-4764-8577-2F13E98A1422
var IID_INamedPropertyStore = syscall.GUID{0x71604B0F, 0x97B0, 0x4764,
	[8]byte{0x85, 0x77, 0x2F, 0x13, 0xE9, 0x8A, 0x14, 0x22}}

type INamedPropertyStoreInterface interface {
	IUnknownInterface
	GetNamedValue(pszName PWSTR, ppropvar *PROPVARIANT) HRESULT
	SetNamedValue(pszName PWSTR, propvar *PROPVARIANT) HRESULT
	GetNameCount(pdwCount *uint32) HRESULT
	GetNameAt(iProp uint32, pbstrName *BSTR) HRESULT
}

type INamedPropertyStoreVtbl struct {
	IUnknownVtbl
	GetNamedValue uintptr
	SetNamedValue uintptr
	GetNameCount  uintptr
	GetNameAt     uintptr
}

type INamedPropertyStore struct {
	IUnknown
}

func (this *INamedPropertyStore) Vtbl() *INamedPropertyStoreVtbl {
	return (*INamedPropertyStoreVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *INamedPropertyStore) GetNamedValue(pszName PWSTR, ppropvar *PROPVARIANT) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetNamedValue, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pszName)), uintptr(unsafe.Pointer(ppropvar)))
	return HRESULT(ret)
}

func (this *INamedPropertyStore) SetNamedValue(pszName PWSTR, propvar *PROPVARIANT) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetNamedValue, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pszName)), uintptr(unsafe.Pointer(propvar)))
	return HRESULT(ret)
}

func (this *INamedPropertyStore) GetNameCount(pdwCount *uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetNameCount, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pdwCount)))
	return HRESULT(ret)
}

func (this *INamedPropertyStore) GetNameAt(iProp uint32, pbstrName *BSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetNameAt, uintptr(unsafe.Pointer(this)), uintptr(iProp), uintptr(unsafe.Pointer(pbstrName)))
	return HRESULT(ret)
}

// FC0CA0A7-C316-4FD2-9031-3E628E6D4F23
var IID_IObjectWithPropertyKey = syscall.GUID{0xFC0CA0A7, 0xC316, 0x4FD2,
	[8]byte{0x90, 0x31, 0x3E, 0x62, 0x8E, 0x6D, 0x4F, 0x23}}

type IObjectWithPropertyKeyInterface interface {
	IUnknownInterface
	SetPropertyKey(key *PROPERTYKEY) HRESULT
	GetPropertyKey(pkey *PROPERTYKEY) HRESULT
}

type IObjectWithPropertyKeyVtbl struct {
	IUnknownVtbl
	SetPropertyKey uintptr
	GetPropertyKey uintptr
}

type IObjectWithPropertyKey struct {
	IUnknown
}

func (this *IObjectWithPropertyKey) Vtbl() *IObjectWithPropertyKeyVtbl {
	return (*IObjectWithPropertyKeyVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IObjectWithPropertyKey) SetPropertyKey(key *PROPERTYKEY) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetPropertyKey, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(key)))
	return HRESULT(ret)
}

func (this *IObjectWithPropertyKey) GetPropertyKey(pkey *PROPERTYKEY) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetPropertyKey, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pkey)))
	return HRESULT(ret)
}

// F917BC8A-1BBA-4478-A245-1BDE03EB9431
var IID_IPropertyChange = syscall.GUID{0xF917BC8A, 0x1BBA, 0x4478,
	[8]byte{0xA2, 0x45, 0x1B, 0xDE, 0x03, 0xEB, 0x94, 0x31}}

type IPropertyChangeInterface interface {
	IObjectWithPropertyKeyInterface
	ApplyToPropVariant(propvarIn *PROPVARIANT, ppropvarOut *PROPVARIANT) HRESULT
}

type IPropertyChangeVtbl struct {
	IObjectWithPropertyKeyVtbl
	ApplyToPropVariant uintptr
}

type IPropertyChange struct {
	IObjectWithPropertyKey
}

func (this *IPropertyChange) Vtbl() *IPropertyChangeVtbl {
	return (*IPropertyChangeVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IPropertyChange) ApplyToPropVariant(propvarIn *PROPVARIANT, ppropvarOut *PROPVARIANT) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().ApplyToPropVariant, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(propvarIn)), uintptr(unsafe.Pointer(ppropvarOut)))
	return HRESULT(ret)
}

// 380F5CAD-1B5E-42F2-805D-637FD392D31E
var IID_IPropertyChangeArray = syscall.GUID{0x380F5CAD, 0x1B5E, 0x42F2,
	[8]byte{0x80, 0x5D, 0x63, 0x7F, 0xD3, 0x92, 0xD3, 0x1E}}

type IPropertyChangeArrayInterface interface {
	IUnknownInterface
	GetCount(pcOperations *uint32) HRESULT
	GetAt(iIndex uint32, riid *syscall.GUID, ppv unsafe.Pointer) HRESULT
	InsertAt(iIndex uint32, ppropChange *IPropertyChange) HRESULT
	Append(ppropChange *IPropertyChange) HRESULT
	AppendOrReplace(ppropChange *IPropertyChange) HRESULT
	RemoveAt(iIndex uint32) HRESULT
	IsKeyInArray(key *PROPERTYKEY) HRESULT
}

type IPropertyChangeArrayVtbl struct {
	IUnknownVtbl
	GetCount        uintptr
	GetAt           uintptr
	InsertAt        uintptr
	Append          uintptr
	AppendOrReplace uintptr
	RemoveAt        uintptr
	IsKeyInArray    uintptr
}

type IPropertyChangeArray struct {
	IUnknown
}

func (this *IPropertyChangeArray) Vtbl() *IPropertyChangeArrayVtbl {
	return (*IPropertyChangeArrayVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IPropertyChangeArray) GetCount(pcOperations *uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetCount, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pcOperations)))
	return HRESULT(ret)
}

func (this *IPropertyChangeArray) GetAt(iIndex uint32, riid *syscall.GUID, ppv unsafe.Pointer) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetAt, uintptr(unsafe.Pointer(this)), uintptr(iIndex), uintptr(unsafe.Pointer(riid)), uintptr(ppv))
	return HRESULT(ret)
}

func (this *IPropertyChangeArray) InsertAt(iIndex uint32, ppropChange *IPropertyChange) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().InsertAt, uintptr(unsafe.Pointer(this)), uintptr(iIndex), uintptr(unsafe.Pointer(ppropChange)))
	return HRESULT(ret)
}

func (this *IPropertyChangeArray) Append(ppropChange *IPropertyChange) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Append, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(ppropChange)))
	return HRESULT(ret)
}

func (this *IPropertyChangeArray) AppendOrReplace(ppropChange *IPropertyChange) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().AppendOrReplace, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(ppropChange)))
	return HRESULT(ret)
}

func (this *IPropertyChangeArray) RemoveAt(iIndex uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().RemoveAt, uintptr(unsafe.Pointer(this)), uintptr(iIndex))
	return HRESULT(ret)
}

func (this *IPropertyChangeArray) IsKeyInArray(key *PROPERTYKEY) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().IsKeyInArray, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(key)))
	return HRESULT(ret)
}

// C8E2D566-186E-4D49-BF41-6909EAD56ACC
var IID_IPropertyStoreCapabilities = syscall.GUID{0xC8E2D566, 0x186E, 0x4D49,
	[8]byte{0xBF, 0x41, 0x69, 0x09, 0xEA, 0xD5, 0x6A, 0xCC}}

type IPropertyStoreCapabilitiesInterface interface {
	IUnknownInterface
	IsPropertyWritable(key *PROPERTYKEY) HRESULT
}

type IPropertyStoreCapabilitiesVtbl struct {
	IUnknownVtbl
	IsPropertyWritable uintptr
}

type IPropertyStoreCapabilities struct {
	IUnknown
}

func (this *IPropertyStoreCapabilities) Vtbl() *IPropertyStoreCapabilitiesVtbl {
	return (*IPropertyStoreCapabilitiesVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IPropertyStoreCapabilities) IsPropertyWritable(key *PROPERTYKEY) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().IsPropertyWritable, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(key)))
	return HRESULT(ret)
}

// 3017056D-9A91-4E90-937D-746C72ABBF4F
var IID_IPropertyStoreCache = syscall.GUID{0x3017056D, 0x9A91, 0x4E90,
	[8]byte{0x93, 0x7D, 0x74, 0x6C, 0x72, 0xAB, 0xBF, 0x4F}}

type IPropertyStoreCacheInterface interface {
	IPropertyStoreInterface
	GetState(key *PROPERTYKEY, pstate *PSC_STATE) HRESULT
	GetValueAndState(key *PROPERTYKEY, ppropvar *PROPVARIANT, pstate *PSC_STATE) HRESULT
	SetState(key *PROPERTYKEY, state PSC_STATE) HRESULT
	SetValueAndState(key *PROPERTYKEY, ppropvar *PROPVARIANT, state PSC_STATE) HRESULT
}

type IPropertyStoreCacheVtbl struct {
	IPropertyStoreVtbl
	GetState         uintptr
	GetValueAndState uintptr
	SetState         uintptr
	SetValueAndState uintptr
}

type IPropertyStoreCache struct {
	IPropertyStore
}

func (this *IPropertyStoreCache) Vtbl() *IPropertyStoreCacheVtbl {
	return (*IPropertyStoreCacheVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IPropertyStoreCache) GetState(key *PROPERTYKEY, pstate *PSC_STATE) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetState, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(key)), uintptr(unsafe.Pointer(pstate)))
	return HRESULT(ret)
}

func (this *IPropertyStoreCache) GetValueAndState(key *PROPERTYKEY, ppropvar *PROPVARIANT, pstate *PSC_STATE) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetValueAndState, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(key)), uintptr(unsafe.Pointer(ppropvar)), uintptr(unsafe.Pointer(pstate)))
	return HRESULT(ret)
}

func (this *IPropertyStoreCache) SetState(key *PROPERTYKEY, state PSC_STATE) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetState, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(key)), uintptr(state))
	return HRESULT(ret)
}

func (this *IPropertyStoreCache) SetValueAndState(key *PROPERTYKEY, ppropvar *PROPVARIANT, state PSC_STATE) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetValueAndState, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(key)), uintptr(unsafe.Pointer(ppropvar)), uintptr(state))
	return HRESULT(ret)
}

// 11E1FBF9-2D56-4A6B-8DB3-7CD193A471F2
var IID_IPropertyEnumType = syscall.GUID{0x11E1FBF9, 0x2D56, 0x4A6B,
	[8]byte{0x8D, 0xB3, 0x7C, 0xD1, 0x93, 0xA4, 0x71, 0xF2}}

type IPropertyEnumTypeInterface interface {
	IUnknownInterface
	GetEnumType(penumtype *PROPENUMTYPE) HRESULT
	GetValue(ppropvar *PROPVARIANT) HRESULT
	GetRangeMinValue(ppropvarMin *PROPVARIANT) HRESULT
	GetRangeSetValue(ppropvarSet *PROPVARIANT) HRESULT
	GetDisplayText(ppszDisplay *PWSTR) HRESULT
}

type IPropertyEnumTypeVtbl struct {
	IUnknownVtbl
	GetEnumType      uintptr
	GetValue         uintptr
	GetRangeMinValue uintptr
	GetRangeSetValue uintptr
	GetDisplayText   uintptr
}

type IPropertyEnumType struct {
	IUnknown
}

func (this *IPropertyEnumType) Vtbl() *IPropertyEnumTypeVtbl {
	return (*IPropertyEnumTypeVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IPropertyEnumType) GetEnumType(penumtype *PROPENUMTYPE) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetEnumType, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(penumtype)))
	return HRESULT(ret)
}

func (this *IPropertyEnumType) GetValue(ppropvar *PROPVARIANT) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetValue, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(ppropvar)))
	return HRESULT(ret)
}

func (this *IPropertyEnumType) GetRangeMinValue(ppropvarMin *PROPVARIANT) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetRangeMinValue, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(ppropvarMin)))
	return HRESULT(ret)
}

func (this *IPropertyEnumType) GetRangeSetValue(ppropvarSet *PROPVARIANT) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetRangeSetValue, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(ppropvarSet)))
	return HRESULT(ret)
}

func (this *IPropertyEnumType) GetDisplayText(ppszDisplay *PWSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetDisplayText, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(ppszDisplay)))
	return HRESULT(ret)
}

// 9B6E051C-5DDD-4321-9070-FE2ACB55E794
var IID_IPropertyEnumType2 = syscall.GUID{0x9B6E051C, 0x5DDD, 0x4321,
	[8]byte{0x90, 0x70, 0xFE, 0x2A, 0xCB, 0x55, 0xE7, 0x94}}

type IPropertyEnumType2Interface interface {
	IPropertyEnumTypeInterface
	GetImageReference(ppszImageRes *PWSTR) HRESULT
}

type IPropertyEnumType2Vtbl struct {
	IPropertyEnumTypeVtbl
	GetImageReference uintptr
}

type IPropertyEnumType2 struct {
	IPropertyEnumType
}

func (this *IPropertyEnumType2) Vtbl() *IPropertyEnumType2Vtbl {
	return (*IPropertyEnumType2Vtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IPropertyEnumType2) GetImageReference(ppszImageRes *PWSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetImageReference, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(ppszImageRes)))
	return HRESULT(ret)
}

// A99400F4-3D84-4557-94BA-1242FB2CC9A6
var IID_IPropertyEnumTypeList = syscall.GUID{0xA99400F4, 0x3D84, 0x4557,
	[8]byte{0x94, 0xBA, 0x12, 0x42, 0xFB, 0x2C, 0xC9, 0xA6}}

type IPropertyEnumTypeListInterface interface {
	IUnknownInterface
	GetCount(pctypes *uint32) HRESULT
	GetAt(itype uint32, riid *syscall.GUID, ppv unsafe.Pointer) HRESULT
	GetConditionAt(nIndex uint32, riid *syscall.GUID, ppv unsafe.Pointer) HRESULT
	FindMatchingIndex(propvarCmp *PROPVARIANT, pnIndex *uint32) HRESULT
}

type IPropertyEnumTypeListVtbl struct {
	IUnknownVtbl
	GetCount          uintptr
	GetAt             uintptr
	GetConditionAt    uintptr
	FindMatchingIndex uintptr
}

type IPropertyEnumTypeList struct {
	IUnknown
}

func (this *IPropertyEnumTypeList) Vtbl() *IPropertyEnumTypeListVtbl {
	return (*IPropertyEnumTypeListVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IPropertyEnumTypeList) GetCount(pctypes *uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetCount, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pctypes)))
	return HRESULT(ret)
}

func (this *IPropertyEnumTypeList) GetAt(itype uint32, riid *syscall.GUID, ppv unsafe.Pointer) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetAt, uintptr(unsafe.Pointer(this)), uintptr(itype), uintptr(unsafe.Pointer(riid)), uintptr(ppv))
	return HRESULT(ret)
}

func (this *IPropertyEnumTypeList) GetConditionAt(nIndex uint32, riid *syscall.GUID, ppv unsafe.Pointer) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetConditionAt, uintptr(unsafe.Pointer(this)), uintptr(nIndex), uintptr(unsafe.Pointer(riid)), uintptr(ppv))
	return HRESULT(ret)
}

func (this *IPropertyEnumTypeList) FindMatchingIndex(propvarCmp *PROPVARIANT, pnIndex *uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().FindMatchingIndex, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(propvarCmp)), uintptr(unsafe.Pointer(pnIndex)))
	return HRESULT(ret)
}

// 6F79D558-3E96-4549-A1D1-7D75D2288814
var IID_IPropertyDescription = syscall.GUID{0x6F79D558, 0x3E96, 0x4549,
	[8]byte{0xA1, 0xD1, 0x7D, 0x75, 0xD2, 0x28, 0x88, 0x14}}

type IPropertyDescriptionInterface interface {
	IUnknownInterface
	GetPropertyKey(pkey *PROPERTYKEY) HRESULT
	GetCanonicalName(ppszName *PWSTR) HRESULT
	GetPropertyType(pvartype *uint16) HRESULT
	GetDisplayName(ppszName *PWSTR) HRESULT
	GetEditInvitation(ppszInvite *PWSTR) HRESULT
	GetTypeFlags(mask PROPDESC_TYPE_FLAGS, ppdtFlags *PROPDESC_TYPE_FLAGS) HRESULT
	GetViewFlags(ppdvFlags *PROPDESC_VIEW_FLAGS) HRESULT
	GetDefaultColumnWidth(pcxChars *uint32) HRESULT
	GetDisplayType(pdisplaytype *PROPDESC_DISPLAYTYPE) HRESULT
	GetColumnState(pcsFlags *uint32) HRESULT
	GetGroupingRange(pgr *PROPDESC_GROUPING_RANGE) HRESULT
	GetRelativeDescriptionType(prdt *PROPDESC_RELATIVEDESCRIPTION_TYPE) HRESULT
	GetRelativeDescription(propvar1 *PROPVARIANT, propvar2 *PROPVARIANT, ppszDesc1 *PWSTR, ppszDesc2 *PWSTR) HRESULT
	GetSortDescription(psd *PROPDESC_SORTDESCRIPTION) HRESULT
	GetSortDescriptionLabel(fDescending BOOL, ppszDescription *PWSTR) HRESULT
	GetAggregationType(paggtype *PROPDESC_AGGREGATION_TYPE) HRESULT
	GetConditionType(pcontype *PROPDESC_CONDITION_TYPE, popDefault unsafe.Pointer) HRESULT
	GetEnumTypeList(riid *syscall.GUID, ppv unsafe.Pointer) HRESULT
	CoerceToCanonicalValue(ppropvar *PROPVARIANT) HRESULT
	FormatForDisplay(propvar *PROPVARIANT, pdfFlags PROPDESC_FORMAT_FLAGS, ppszDisplay *PWSTR) HRESULT
	IsValueCanonical(propvar *PROPVARIANT) HRESULT
}

type IPropertyDescriptionVtbl struct {
	IUnknownVtbl
	GetPropertyKey             uintptr
	GetCanonicalName           uintptr
	GetPropertyType            uintptr
	GetDisplayName             uintptr
	GetEditInvitation          uintptr
	GetTypeFlags               uintptr
	GetViewFlags               uintptr
	GetDefaultColumnWidth      uintptr
	GetDisplayType             uintptr
	GetColumnState             uintptr
	GetGroupingRange           uintptr
	GetRelativeDescriptionType uintptr
	GetRelativeDescription     uintptr
	GetSortDescription         uintptr
	GetSortDescriptionLabel    uintptr
	GetAggregationType         uintptr
	GetConditionType           uintptr
	GetEnumTypeList            uintptr
	CoerceToCanonicalValue     uintptr
	FormatForDisplay           uintptr
	IsValueCanonical           uintptr
}

type IPropertyDescription struct {
	IUnknown
}

func (this *IPropertyDescription) Vtbl() *IPropertyDescriptionVtbl {
	return (*IPropertyDescriptionVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IPropertyDescription) GetPropertyKey(pkey *PROPERTYKEY) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetPropertyKey, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pkey)))
	return HRESULT(ret)
}

func (this *IPropertyDescription) GetCanonicalName(ppszName *PWSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetCanonicalName, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(ppszName)))
	return HRESULT(ret)
}

func (this *IPropertyDescription) GetPropertyType(pvartype *uint16) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetPropertyType, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pvartype)))
	return HRESULT(ret)
}

func (this *IPropertyDescription) GetDisplayName(ppszName *PWSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetDisplayName, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(ppszName)))
	return HRESULT(ret)
}

func (this *IPropertyDescription) GetEditInvitation(ppszInvite *PWSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetEditInvitation, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(ppszInvite)))
	return HRESULT(ret)
}

func (this *IPropertyDescription) GetTypeFlags(mask PROPDESC_TYPE_FLAGS, ppdtFlags *PROPDESC_TYPE_FLAGS) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetTypeFlags, uintptr(unsafe.Pointer(this)), uintptr(mask), uintptr(unsafe.Pointer(ppdtFlags)))
	return HRESULT(ret)
}

func (this *IPropertyDescription) GetViewFlags(ppdvFlags *PROPDESC_VIEW_FLAGS) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetViewFlags, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(ppdvFlags)))
	return HRESULT(ret)
}

func (this *IPropertyDescription) GetDefaultColumnWidth(pcxChars *uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetDefaultColumnWidth, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pcxChars)))
	return HRESULT(ret)
}

func (this *IPropertyDescription) GetDisplayType(pdisplaytype *PROPDESC_DISPLAYTYPE) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetDisplayType, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pdisplaytype)))
	return HRESULT(ret)
}

func (this *IPropertyDescription) GetColumnState(pcsFlags *uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetColumnState, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pcsFlags)))
	return HRESULT(ret)
}

func (this *IPropertyDescription) GetGroupingRange(pgr *PROPDESC_GROUPING_RANGE) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetGroupingRange, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pgr)))
	return HRESULT(ret)
}

func (this *IPropertyDescription) GetRelativeDescriptionType(prdt *PROPDESC_RELATIVEDESCRIPTION_TYPE) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetRelativeDescriptionType, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(prdt)))
	return HRESULT(ret)
}

func (this *IPropertyDescription) GetRelativeDescription(propvar1 *PROPVARIANT, propvar2 *PROPVARIANT, ppszDesc1 *PWSTR, ppszDesc2 *PWSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetRelativeDescription, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(propvar1)), uintptr(unsafe.Pointer(propvar2)), uintptr(unsafe.Pointer(ppszDesc1)), uintptr(unsafe.Pointer(ppszDesc2)))
	return HRESULT(ret)
}

func (this *IPropertyDescription) GetSortDescription(psd *PROPDESC_SORTDESCRIPTION) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetSortDescription, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(psd)))
	return HRESULT(ret)
}

func (this *IPropertyDescription) GetSortDescriptionLabel(fDescending BOOL, ppszDescription *PWSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetSortDescriptionLabel, uintptr(unsafe.Pointer(this)), uintptr(fDescending), uintptr(unsafe.Pointer(ppszDescription)))
	return HRESULT(ret)
}

func (this *IPropertyDescription) GetAggregationType(paggtype *PROPDESC_AGGREGATION_TYPE) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetAggregationType, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(paggtype)))
	return HRESULT(ret)
}

func (this *IPropertyDescription) GetConditionType(pcontype *PROPDESC_CONDITION_TYPE, popDefault unsafe.Pointer) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetConditionType, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pcontype)), uintptr(popDefault))
	return HRESULT(ret)
}

func (this *IPropertyDescription) GetEnumTypeList(riid *syscall.GUID, ppv unsafe.Pointer) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetEnumTypeList, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(riid)), uintptr(ppv))
	return HRESULT(ret)
}

func (this *IPropertyDescription) CoerceToCanonicalValue(ppropvar *PROPVARIANT) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().CoerceToCanonicalValue, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(ppropvar)))
	return HRESULT(ret)
}

func (this *IPropertyDescription) FormatForDisplay(propvar *PROPVARIANT, pdfFlags PROPDESC_FORMAT_FLAGS, ppszDisplay *PWSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().FormatForDisplay, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(propvar)), uintptr(pdfFlags), uintptr(unsafe.Pointer(ppszDisplay)))
	return HRESULT(ret)
}

func (this *IPropertyDescription) IsValueCanonical(propvar *PROPVARIANT) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().IsValueCanonical, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(propvar)))
	return HRESULT(ret)
}

// 57D2EDED-5062-400E-B107-5DAE79FE57A6
var IID_IPropertyDescription2 = syscall.GUID{0x57D2EDED, 0x5062, 0x400E,
	[8]byte{0xB1, 0x07, 0x5D, 0xAE, 0x79, 0xFE, 0x57, 0xA6}}

type IPropertyDescription2Interface interface {
	IPropertyDescriptionInterface
	GetImageReferenceForValue(propvar *PROPVARIANT, ppszImageRes *PWSTR) HRESULT
}

type IPropertyDescription2Vtbl struct {
	IPropertyDescriptionVtbl
	GetImageReferenceForValue uintptr
}

type IPropertyDescription2 struct {
	IPropertyDescription
}

func (this *IPropertyDescription2) Vtbl() *IPropertyDescription2Vtbl {
	return (*IPropertyDescription2Vtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IPropertyDescription2) GetImageReferenceForValue(propvar *PROPVARIANT, ppszImageRes *PWSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetImageReferenceForValue, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(propvar)), uintptr(unsafe.Pointer(ppszImageRes)))
	return HRESULT(ret)
}

// F67104FC-2AF9-46FD-B32D-243C1404F3D1
var IID_IPropertyDescriptionAliasInfo = syscall.GUID{0xF67104FC, 0x2AF9, 0x46FD,
	[8]byte{0xB3, 0x2D, 0x24, 0x3C, 0x14, 0x04, 0xF3, 0xD1}}

type IPropertyDescriptionAliasInfoInterface interface {
	IPropertyDescriptionInterface
	GetSortByAlias(riid *syscall.GUID, ppv unsafe.Pointer) HRESULT
	GetAdditionalSortByAliases(riid *syscall.GUID, ppv unsafe.Pointer) HRESULT
}

type IPropertyDescriptionAliasInfoVtbl struct {
	IPropertyDescriptionVtbl
	GetSortByAlias             uintptr
	GetAdditionalSortByAliases uintptr
}

type IPropertyDescriptionAliasInfo struct {
	IPropertyDescription
}

func (this *IPropertyDescriptionAliasInfo) Vtbl() *IPropertyDescriptionAliasInfoVtbl {
	return (*IPropertyDescriptionAliasInfoVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IPropertyDescriptionAliasInfo) GetSortByAlias(riid *syscall.GUID, ppv unsafe.Pointer) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetSortByAlias, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(riid)), uintptr(ppv))
	return HRESULT(ret)
}

func (this *IPropertyDescriptionAliasInfo) GetAdditionalSortByAliases(riid *syscall.GUID, ppv unsafe.Pointer) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetAdditionalSortByAliases, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(riid)), uintptr(ppv))
	return HRESULT(ret)
}

// 078F91BD-29A2-440F-924E-46A291524520
var IID_IPropertyDescriptionSearchInfo = syscall.GUID{0x078F91BD, 0x29A2, 0x440F,
	[8]byte{0x92, 0x4E, 0x46, 0xA2, 0x91, 0x52, 0x45, 0x20}}

type IPropertyDescriptionSearchInfoInterface interface {
	IPropertyDescriptionInterface
	GetSearchInfoFlags(ppdsiFlags *PROPDESC_SEARCHINFO_FLAGS) HRESULT
	GetColumnIndexType(ppdciType *PROPDESC_COLUMNINDEX_TYPE) HRESULT
	GetProjectionString(ppszProjection *PWSTR) HRESULT
	GetMaxSize(pcbMaxSize *uint32) HRESULT
}

type IPropertyDescriptionSearchInfoVtbl struct {
	IPropertyDescriptionVtbl
	GetSearchInfoFlags  uintptr
	GetColumnIndexType  uintptr
	GetProjectionString uintptr
	GetMaxSize          uintptr
}

type IPropertyDescriptionSearchInfo struct {
	IPropertyDescription
}

func (this *IPropertyDescriptionSearchInfo) Vtbl() *IPropertyDescriptionSearchInfoVtbl {
	return (*IPropertyDescriptionSearchInfoVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IPropertyDescriptionSearchInfo) GetSearchInfoFlags(ppdsiFlags *PROPDESC_SEARCHINFO_FLAGS) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetSearchInfoFlags, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(ppdsiFlags)))
	return HRESULT(ret)
}

func (this *IPropertyDescriptionSearchInfo) GetColumnIndexType(ppdciType *PROPDESC_COLUMNINDEX_TYPE) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetColumnIndexType, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(ppdciType)))
	return HRESULT(ret)
}

func (this *IPropertyDescriptionSearchInfo) GetProjectionString(ppszProjection *PWSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetProjectionString, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(ppszProjection)))
	return HRESULT(ret)
}

func (this *IPropertyDescriptionSearchInfo) GetMaxSize(pcbMaxSize *uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetMaxSize, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pcbMaxSize)))
	return HRESULT(ret)
}

// 507393F4-2A3D-4A60-B59E-D9C75716C2DD
var IID_IPropertyDescriptionRelatedPropertyInfo = syscall.GUID{0x507393F4, 0x2A3D, 0x4A60,
	[8]byte{0xB5, 0x9E, 0xD9, 0xC7, 0x57, 0x16, 0xC2, 0xDD}}

type IPropertyDescriptionRelatedPropertyInfoInterface interface {
	IPropertyDescriptionInterface
	GetRelatedProperty(pszRelationshipName PWSTR, riid *syscall.GUID, ppv unsafe.Pointer) HRESULT
}

type IPropertyDescriptionRelatedPropertyInfoVtbl struct {
	IPropertyDescriptionVtbl
	GetRelatedProperty uintptr
}

type IPropertyDescriptionRelatedPropertyInfo struct {
	IPropertyDescription
}

func (this *IPropertyDescriptionRelatedPropertyInfo) Vtbl() *IPropertyDescriptionRelatedPropertyInfoVtbl {
	return (*IPropertyDescriptionRelatedPropertyInfoVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IPropertyDescriptionRelatedPropertyInfo) GetRelatedProperty(pszRelationshipName PWSTR, riid *syscall.GUID, ppv unsafe.Pointer) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetRelatedProperty, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pszRelationshipName)), uintptr(unsafe.Pointer(riid)), uintptr(ppv))
	return HRESULT(ret)
}

// CA724E8A-C3E6-442B-88A4-6FB0DB8035A3
var IID_IPropertySystem = syscall.GUID{0xCA724E8A, 0xC3E6, 0x442B,
	[8]byte{0x88, 0xA4, 0x6F, 0xB0, 0xDB, 0x80, 0x35, 0xA3}}

type IPropertySystemInterface interface {
	IUnknownInterface
	GetPropertyDescription(propkey *PROPERTYKEY, riid *syscall.GUID, ppv unsafe.Pointer) HRESULT
	GetPropertyDescriptionByName(pszCanonicalName PWSTR, riid *syscall.GUID, ppv unsafe.Pointer) HRESULT
	GetPropertyDescriptionListFromString(pszPropList PWSTR, riid *syscall.GUID, ppv unsafe.Pointer) HRESULT
	EnumeratePropertyDescriptions(filterOn PROPDESC_ENUMFILTER, riid *syscall.GUID, ppv unsafe.Pointer) HRESULT
	FormatForDisplay(key *PROPERTYKEY, propvar *PROPVARIANT, pdff PROPDESC_FORMAT_FLAGS, pszText PWSTR, cchText uint32) HRESULT
	FormatForDisplayAlloc(key *PROPERTYKEY, propvar *PROPVARIANT, pdff PROPDESC_FORMAT_FLAGS, ppszDisplay *PWSTR) HRESULT
	RegisterPropertySchema(pszPath PWSTR) HRESULT
	UnregisterPropertySchema(pszPath PWSTR) HRESULT
	RefreshPropertySchema() HRESULT
}

type IPropertySystemVtbl struct {
	IUnknownVtbl
	GetPropertyDescription               uintptr
	GetPropertyDescriptionByName         uintptr
	GetPropertyDescriptionListFromString uintptr
	EnumeratePropertyDescriptions        uintptr
	FormatForDisplay                     uintptr
	FormatForDisplayAlloc                uintptr
	RegisterPropertySchema               uintptr
	UnregisterPropertySchema             uintptr
	RefreshPropertySchema                uintptr
}

type IPropertySystem struct {
	IUnknown
}

func (this *IPropertySystem) Vtbl() *IPropertySystemVtbl {
	return (*IPropertySystemVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IPropertySystem) GetPropertyDescription(propkey *PROPERTYKEY, riid *syscall.GUID, ppv unsafe.Pointer) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetPropertyDescription, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(propkey)), uintptr(unsafe.Pointer(riid)), uintptr(ppv))
	return HRESULT(ret)
}

func (this *IPropertySystem) GetPropertyDescriptionByName(pszCanonicalName PWSTR, riid *syscall.GUID, ppv unsafe.Pointer) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetPropertyDescriptionByName, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pszCanonicalName)), uintptr(unsafe.Pointer(riid)), uintptr(ppv))
	return HRESULT(ret)
}

func (this *IPropertySystem) GetPropertyDescriptionListFromString(pszPropList PWSTR, riid *syscall.GUID, ppv unsafe.Pointer) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetPropertyDescriptionListFromString, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pszPropList)), uintptr(unsafe.Pointer(riid)), uintptr(ppv))
	return HRESULT(ret)
}

func (this *IPropertySystem) EnumeratePropertyDescriptions(filterOn PROPDESC_ENUMFILTER, riid *syscall.GUID, ppv unsafe.Pointer) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().EnumeratePropertyDescriptions, uintptr(unsafe.Pointer(this)), uintptr(filterOn), uintptr(unsafe.Pointer(riid)), uintptr(ppv))
	return HRESULT(ret)
}

func (this *IPropertySystem) FormatForDisplay(key *PROPERTYKEY, propvar *PROPVARIANT, pdff PROPDESC_FORMAT_FLAGS, pszText PWSTR, cchText uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().FormatForDisplay, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(key)), uintptr(unsafe.Pointer(propvar)), uintptr(pdff), uintptr(unsafe.Pointer(pszText)), uintptr(cchText))
	return HRESULT(ret)
}

func (this *IPropertySystem) FormatForDisplayAlloc(key *PROPERTYKEY, propvar *PROPVARIANT, pdff PROPDESC_FORMAT_FLAGS, ppszDisplay *PWSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().FormatForDisplayAlloc, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(key)), uintptr(unsafe.Pointer(propvar)), uintptr(pdff), uintptr(unsafe.Pointer(ppszDisplay)))
	return HRESULT(ret)
}

func (this *IPropertySystem) RegisterPropertySchema(pszPath PWSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().RegisterPropertySchema, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pszPath)))
	return HRESULT(ret)
}

func (this *IPropertySystem) UnregisterPropertySchema(pszPath PWSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().UnregisterPropertySchema, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pszPath)))
	return HRESULT(ret)
}

func (this *IPropertySystem) RefreshPropertySchema() HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().RefreshPropertySchema, uintptr(unsafe.Pointer(this)))
	return HRESULT(ret)
}

// 1F9FC1D0-C39B-4B26-817F-011967D3440E
var IID_IPropertyDescriptionList = syscall.GUID{0x1F9FC1D0, 0xC39B, 0x4B26,
	[8]byte{0x81, 0x7F, 0x01, 0x19, 0x67, 0xD3, 0x44, 0x0E}}

type IPropertyDescriptionListInterface interface {
	IUnknownInterface
	GetCount(pcElem *uint32) HRESULT
	GetAt(iElem uint32, riid *syscall.GUID, ppv unsafe.Pointer) HRESULT
}

type IPropertyDescriptionListVtbl struct {
	IUnknownVtbl
	GetCount uintptr
	GetAt    uintptr
}

type IPropertyDescriptionList struct {
	IUnknown
}

func (this *IPropertyDescriptionList) Vtbl() *IPropertyDescriptionListVtbl {
	return (*IPropertyDescriptionListVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IPropertyDescriptionList) GetCount(pcElem *uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetCount, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pcElem)))
	return HRESULT(ret)
}

func (this *IPropertyDescriptionList) GetAt(iElem uint32, riid *syscall.GUID, ppv unsafe.Pointer) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetAt, uintptr(unsafe.Pointer(this)), uintptr(iElem), uintptr(unsafe.Pointer(riid)), uintptr(ppv))
	return HRESULT(ret)
}

// BC110B6D-57E8-4148-A9C6-91015AB2F3A5
var IID_IPropertyStoreFactory = syscall.GUID{0xBC110B6D, 0x57E8, 0x4148,
	[8]byte{0xA9, 0xC6, 0x91, 0x01, 0x5A, 0xB2, 0xF3, 0xA5}}

type IPropertyStoreFactoryInterface interface {
	IUnknownInterface
	GetPropertyStore(flags GETPROPERTYSTOREFLAGS, pUnkFactory *IUnknown, riid *syscall.GUID, ppv unsafe.Pointer) HRESULT
	GetPropertyStoreForKeys(rgKeys *PROPERTYKEY, cKeys uint32, flags GETPROPERTYSTOREFLAGS, riid *syscall.GUID, ppv unsafe.Pointer) HRESULT
}

type IPropertyStoreFactoryVtbl struct {
	IUnknownVtbl
	GetPropertyStore        uintptr
	GetPropertyStoreForKeys uintptr
}

type IPropertyStoreFactory struct {
	IUnknown
}

func (this *IPropertyStoreFactory) Vtbl() *IPropertyStoreFactoryVtbl {
	return (*IPropertyStoreFactoryVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IPropertyStoreFactory) GetPropertyStore(flags GETPROPERTYSTOREFLAGS, pUnkFactory *IUnknown, riid *syscall.GUID, ppv unsafe.Pointer) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetPropertyStore, uintptr(unsafe.Pointer(this)), uintptr(flags), uintptr(unsafe.Pointer(pUnkFactory)), uintptr(unsafe.Pointer(riid)), uintptr(ppv))
	return HRESULT(ret)
}

func (this *IPropertyStoreFactory) GetPropertyStoreForKeys(rgKeys *PROPERTYKEY, cKeys uint32, flags GETPROPERTYSTOREFLAGS, riid *syscall.GUID, ppv unsafe.Pointer) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetPropertyStoreForKeys, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(rgKeys)), uintptr(cKeys), uintptr(flags), uintptr(unsafe.Pointer(riid)), uintptr(ppv))
	return HRESULT(ret)
}

// 40D4577F-E237-4BDB-BD69-58F089431B6A
var IID_IDelayedPropertyStoreFactory = syscall.GUID{0x40D4577F, 0xE237, 0x4BDB,
	[8]byte{0xBD, 0x69, 0x58, 0xF0, 0x89, 0x43, 0x1B, 0x6A}}

type IDelayedPropertyStoreFactoryInterface interface {
	IPropertyStoreFactoryInterface
	GetDelayedPropertyStore(flags GETPROPERTYSTOREFLAGS, dwStoreId uint32, riid *syscall.GUID, ppv unsafe.Pointer) HRESULT
}

type IDelayedPropertyStoreFactoryVtbl struct {
	IPropertyStoreFactoryVtbl
	GetDelayedPropertyStore uintptr
}

type IDelayedPropertyStoreFactory struct {
	IPropertyStoreFactory
}

func (this *IDelayedPropertyStoreFactory) Vtbl() *IDelayedPropertyStoreFactoryVtbl {
	return (*IDelayedPropertyStoreFactoryVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IDelayedPropertyStoreFactory) GetDelayedPropertyStore(flags GETPROPERTYSTOREFLAGS, dwStoreId uint32, riid *syscall.GUID, ppv unsafe.Pointer) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetDelayedPropertyStore, uintptr(unsafe.Pointer(this)), uintptr(flags), uintptr(dwStoreId), uintptr(unsafe.Pointer(riid)), uintptr(ppv))
	return HRESULT(ret)
}

// E318AD57-0AA0-450F-ACA5-6FAB7103D917
var IID_IPersistSerializedPropStorage = syscall.GUID{0xE318AD57, 0x0AA0, 0x450F,
	[8]byte{0xAC, 0xA5, 0x6F, 0xAB, 0x71, 0x03, 0xD9, 0x17}}

type IPersistSerializedPropStorageInterface interface {
	IUnknownInterface
	SetFlags(flags int32) HRESULT
	SetPropertyStorage(psps *SERIALIZEDPROPSTORAGE, cb uint32) HRESULT
	GetPropertyStorage(ppsps **SERIALIZEDPROPSTORAGE, pcb *uint32) HRESULT
}

type IPersistSerializedPropStorageVtbl struct {
	IUnknownVtbl
	SetFlags           uintptr
	SetPropertyStorage uintptr
	GetPropertyStorage uintptr
}

type IPersistSerializedPropStorage struct {
	IUnknown
}

func (this *IPersistSerializedPropStorage) Vtbl() *IPersistSerializedPropStorageVtbl {
	return (*IPersistSerializedPropStorageVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IPersistSerializedPropStorage) SetFlags(flags int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetFlags, uintptr(unsafe.Pointer(this)), uintptr(flags))
	return HRESULT(ret)
}

func (this *IPersistSerializedPropStorage) SetPropertyStorage(psps *SERIALIZEDPROPSTORAGE, cb uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetPropertyStorage, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(psps)), uintptr(cb))
	return HRESULT(ret)
}

func (this *IPersistSerializedPropStorage) GetPropertyStorage(ppsps **SERIALIZEDPROPSTORAGE, pcb *uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetPropertyStorage, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(ppsps)), uintptr(unsafe.Pointer(pcb)))
	return HRESULT(ret)
}

// 77EFFA68-4F98-4366-BA72-573B3D880571
var IID_IPersistSerializedPropStorage2 = syscall.GUID{0x77EFFA68, 0x4F98, 0x4366,
	[8]byte{0xBA, 0x72, 0x57, 0x3B, 0x3D, 0x88, 0x05, 0x71}}

type IPersistSerializedPropStorage2Interface interface {
	IPersistSerializedPropStorageInterface
	GetPropertyStorageSize(pcb *uint32) HRESULT
	GetPropertyStorageBuffer(psps *SERIALIZEDPROPSTORAGE, cb uint32, pcbWritten *uint32) HRESULT
}

type IPersistSerializedPropStorage2Vtbl struct {
	IPersistSerializedPropStorageVtbl
	GetPropertyStorageSize   uintptr
	GetPropertyStorageBuffer uintptr
}

type IPersistSerializedPropStorage2 struct {
	IPersistSerializedPropStorage
}

func (this *IPersistSerializedPropStorage2) Vtbl() *IPersistSerializedPropStorage2Vtbl {
	return (*IPersistSerializedPropStorage2Vtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IPersistSerializedPropStorage2) GetPropertyStorageSize(pcb *uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetPropertyStorageSize, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pcb)))
	return HRESULT(ret)
}

func (this *IPersistSerializedPropStorage2) GetPropertyStorageBuffer(psps *SERIALIZEDPROPSTORAGE, cb uint32, pcbWritten *uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetPropertyStorageBuffer, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(psps)), uintptr(cb), uintptr(unsafe.Pointer(pcbWritten)))
	return HRESULT(ret)
}

// FA955FD9-38BE-4879-A6CE-824CF52D609F
var IID_IPropertySystemChangeNotify = syscall.GUID{0xFA955FD9, 0x38BE, 0x4879,
	[8]byte{0xA6, 0xCE, 0x82, 0x4C, 0xF5, 0x2D, 0x60, 0x9F}}

type IPropertySystemChangeNotifyInterface interface {
	IUnknownInterface
	SchemaRefreshed() HRESULT
}

type IPropertySystemChangeNotifyVtbl struct {
	IUnknownVtbl
	SchemaRefreshed uintptr
}

type IPropertySystemChangeNotify struct {
	IUnknown
}

func (this *IPropertySystemChangeNotify) Vtbl() *IPropertySystemChangeNotifyVtbl {
	return (*IPropertySystemChangeNotifyVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IPropertySystemChangeNotify) SchemaRefreshed() HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SchemaRefreshed, uintptr(unsafe.Pointer(this)))
	return HRESULT(ret)
}

// 75121952-E0D0-43E5-9380-1D80483ACF72
var IID_ICreateObject = syscall.GUID{0x75121952, 0xE0D0, 0x43E5,
	[8]byte{0x93, 0x80, 0x1D, 0x80, 0x48, 0x3A, 0xCF, 0x72}}

type ICreateObjectInterface interface {
	IUnknownInterface
	CreateObject(clsid *syscall.GUID, pUnkOuter *IUnknown, riid *syscall.GUID, ppv unsafe.Pointer) HRESULT
}

type ICreateObjectVtbl struct {
	IUnknownVtbl
	CreateObject uintptr
}

type ICreateObject struct {
	IUnknown
}

func (this *ICreateObject) Vtbl() *ICreateObjectVtbl {
	return (*ICreateObjectVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *ICreateObject) CreateObject(clsid *syscall.GUID, pUnkOuter *IUnknown, riid *syscall.GUID, ppv unsafe.Pointer) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().CreateObject, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(clsid)), uintptr(unsafe.Pointer(pUnkOuter)), uintptr(unsafe.Pointer(riid)), uintptr(ppv))
	return HRESULT(ret)
}

// 757A7D9F-919A-4118-99D7-DBB208C8CC66
var IID_IPropertyUI = syscall.GUID{0x757A7D9F, 0x919A, 0x4118,
	[8]byte{0x99, 0xD7, 0xDB, 0xB2, 0x08, 0xC8, 0xCC, 0x66}}

type IPropertyUIInterface interface {
	IUnknownInterface
	ParsePropertyName(pszName PWSTR, pfmtid *syscall.GUID, ppid *uint32, pchEaten *uint32) HRESULT
	GetCannonicalName(fmtid *syscall.GUID, pid uint32, pwszText PWSTR, cchText uint32) HRESULT
	GetDisplayName(fmtid *syscall.GUID, pid uint32, flags PROPERTYUI_NAME_FLAGS, pwszText PWSTR, cchText uint32) HRESULT
	GetPropertyDescription(fmtid *syscall.GUID, pid uint32, pwszText PWSTR, cchText uint32) HRESULT
	GetDefaultWidth(fmtid *syscall.GUID, pid uint32, pcxChars *uint32) HRESULT
	GetFlags(fmtid *syscall.GUID, pid uint32, pflags *PROPERTYUI_FLAGS) HRESULT
	FormatForDisplay(fmtid *syscall.GUID, pid uint32, ppropvar *PROPVARIANT, puiff PROPERTYUI_FORMAT_FLAGS, pwszText PWSTR, cchText uint32) HRESULT
	GetHelpInfo(fmtid *syscall.GUID, pid uint32, pwszHelpFile PWSTR, cch uint32, puHelpID *uint32) HRESULT
}

type IPropertyUIVtbl struct {
	IUnknownVtbl
	ParsePropertyName      uintptr
	GetCannonicalName      uintptr
	GetDisplayName         uintptr
	GetPropertyDescription uintptr
	GetDefaultWidth        uintptr
	GetFlags               uintptr
	FormatForDisplay       uintptr
	GetHelpInfo            uintptr
}

type IPropertyUI struct {
	IUnknown
}

func (this *IPropertyUI) Vtbl() *IPropertyUIVtbl {
	return (*IPropertyUIVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IPropertyUI) ParsePropertyName(pszName PWSTR, pfmtid *syscall.GUID, ppid *uint32, pchEaten *uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().ParsePropertyName, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pszName)), uintptr(unsafe.Pointer(pfmtid)), uintptr(unsafe.Pointer(ppid)), uintptr(unsafe.Pointer(pchEaten)))
	return HRESULT(ret)
}

func (this *IPropertyUI) GetCannonicalName(fmtid *syscall.GUID, pid uint32, pwszText PWSTR, cchText uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetCannonicalName, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(fmtid)), uintptr(pid), uintptr(unsafe.Pointer(pwszText)), uintptr(cchText))
	return HRESULT(ret)
}

func (this *IPropertyUI) GetDisplayName(fmtid *syscall.GUID, pid uint32, flags PROPERTYUI_NAME_FLAGS, pwszText PWSTR, cchText uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetDisplayName, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(fmtid)), uintptr(pid), uintptr(flags), uintptr(unsafe.Pointer(pwszText)), uintptr(cchText))
	return HRESULT(ret)
}

func (this *IPropertyUI) GetPropertyDescription(fmtid *syscall.GUID, pid uint32, pwszText PWSTR, cchText uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetPropertyDescription, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(fmtid)), uintptr(pid), uintptr(unsafe.Pointer(pwszText)), uintptr(cchText))
	return HRESULT(ret)
}

func (this *IPropertyUI) GetDefaultWidth(fmtid *syscall.GUID, pid uint32, pcxChars *uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetDefaultWidth, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(fmtid)), uintptr(pid), uintptr(unsafe.Pointer(pcxChars)))
	return HRESULT(ret)
}

func (this *IPropertyUI) GetFlags(fmtid *syscall.GUID, pid uint32, pflags *PROPERTYUI_FLAGS) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetFlags, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(fmtid)), uintptr(pid), uintptr(unsafe.Pointer(pflags)))
	return HRESULT(ret)
}

func (this *IPropertyUI) FormatForDisplay(fmtid *syscall.GUID, pid uint32, ppropvar *PROPVARIANT, puiff PROPERTYUI_FORMAT_FLAGS, pwszText PWSTR, cchText uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().FormatForDisplay, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(fmtid)), uintptr(pid), uintptr(unsafe.Pointer(ppropvar)), uintptr(puiff), uintptr(unsafe.Pointer(pwszText)), uintptr(cchText))
	return HRESULT(ret)
}

func (this *IPropertyUI) GetHelpInfo(fmtid *syscall.GUID, pid uint32, pwszHelpFile PWSTR, cch uint32, puHelpID *uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetHelpInfo, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(fmtid)), uintptr(pid), uintptr(unsafe.Pointer(pwszHelpFile)), uintptr(cch), uintptr(unsafe.Pointer(puHelpID)))
	return HRESULT(ret)
}

var (
	pSHGetPropertyStoreFromIDList      uintptr
	pSHGetPropertyStoreFromParsingName uintptr
	pSHAddDefaultPropertiesByExt       uintptr
	pPifMgr_OpenProperties             uintptr
	pPifMgr_GetProperties              uintptr
	pPifMgr_SetProperties              uintptr
	pPifMgr_CloseProperties            uintptr
	pSHPropStgCreate                   uintptr
	pSHPropStgReadMultiple             uintptr
	pSHPropStgWriteMultiple            uintptr
	pSHGetPropertyStoreForWindow       uintptr
)

func SHGetPropertyStoreFromIDList(pidl *ITEMIDLIST, flags GETPROPERTYSTOREFLAGS, riid *syscall.GUID, ppv unsafe.Pointer) HRESULT {
	addr := lazyAddr(&pSHGetPropertyStoreFromIDList, libShell32, "SHGetPropertyStoreFromIDList")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pidl)), uintptr(flags), uintptr(unsafe.Pointer(riid)), uintptr(ppv))
	return HRESULT(ret)
}

func SHGetPropertyStoreFromParsingName(pszPath PWSTR, pbc *IBindCtx, flags GETPROPERTYSTOREFLAGS, riid *syscall.GUID, ppv unsafe.Pointer) HRESULT {
	addr := lazyAddr(&pSHGetPropertyStoreFromParsingName, libShell32, "SHGetPropertyStoreFromParsingName")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pszPath)), uintptr(unsafe.Pointer(pbc)), uintptr(flags), uintptr(unsafe.Pointer(riid)), uintptr(ppv))
	return HRESULT(ret)
}

func SHAddDefaultPropertiesByExt(pszExt PWSTR, pPropStore *IPropertyStore) HRESULT {
	addr := lazyAddr(&pSHAddDefaultPropertiesByExt, libShell32, "SHAddDefaultPropertiesByExt")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pszExt)), uintptr(unsafe.Pointer(pPropStore)))
	return HRESULT(ret)
}

func PifMgr_OpenProperties(pszApp PWSTR, pszPIF PWSTR, hInf uint32, flOpt uint32) HANDLE {
	addr := lazyAddr(&pPifMgr_OpenProperties, libShell32, "PifMgr_OpenProperties")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pszApp)), uintptr(unsafe.Pointer(pszPIF)), uintptr(hInf), uintptr(flOpt))
	return ret
}

func PifMgr_GetProperties(hProps HANDLE, pszGroup PSTR, lpProps unsafe.Pointer, cbProps int32, flOpt uint32) int32 {
	addr := lazyAddr(&pPifMgr_GetProperties, libShell32, "PifMgr_GetProperties")
	ret, _, _ := syscall.SyscallN(addr, hProps, uintptr(unsafe.Pointer(pszGroup)), uintptr(lpProps), uintptr(cbProps), uintptr(flOpt))
	return int32(ret)
}

func PifMgr_SetProperties(hProps HANDLE, pszGroup PSTR, lpProps unsafe.Pointer, cbProps int32, flOpt uint32) int32 {
	addr := lazyAddr(&pPifMgr_SetProperties, libShell32, "PifMgr_SetProperties")
	ret, _, _ := syscall.SyscallN(addr, hProps, uintptr(unsafe.Pointer(pszGroup)), uintptr(lpProps), uintptr(cbProps), uintptr(flOpt))
	return int32(ret)
}

func PifMgr_CloseProperties(hProps HANDLE, flOpt uint32) HANDLE {
	addr := lazyAddr(&pPifMgr_CloseProperties, libShell32, "PifMgr_CloseProperties")
	ret, _, _ := syscall.SyscallN(addr, hProps, uintptr(flOpt))
	return ret
}

func SHPropStgCreate(psstg *IPropertySetStorage, fmtid *syscall.GUID, pclsid *syscall.GUID, grfFlags uint32, grfMode uint32, dwDisposition uint32, ppstg **IPropertyStorage, puCodePage *uint32) HRESULT {
	addr := lazyAddr(&pSHPropStgCreate, libShell32, "SHPropStgCreate")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(psstg)), uintptr(unsafe.Pointer(fmtid)), uintptr(unsafe.Pointer(pclsid)), uintptr(grfFlags), uintptr(grfMode), uintptr(dwDisposition), uintptr(unsafe.Pointer(ppstg)), uintptr(unsafe.Pointer(puCodePage)))
	return HRESULT(ret)
}

func SHPropStgReadMultiple(pps *IPropertyStorage, uCodePage uint32, cpspec uint32, rgpspec *PROPSPEC, rgvar *PROPVARIANT) HRESULT {
	addr := lazyAddr(&pSHPropStgReadMultiple, libShell32, "SHPropStgReadMultiple")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pps)), uintptr(uCodePage), uintptr(cpspec), uintptr(unsafe.Pointer(rgpspec)), uintptr(unsafe.Pointer(rgvar)))
	return HRESULT(ret)
}

func SHPropStgWriteMultiple(pps *IPropertyStorage, puCodePage *uint32, cpspec uint32, rgpspec *PROPSPEC, rgvar *PROPVARIANT, propidNameFirst uint32) HRESULT {
	addr := lazyAddr(&pSHPropStgWriteMultiple, libShell32, "SHPropStgWriteMultiple")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pps)), uintptr(unsafe.Pointer(puCodePage)), uintptr(cpspec), uintptr(unsafe.Pointer(rgpspec)), uintptr(unsafe.Pointer(rgvar)), uintptr(propidNameFirst))
	return HRESULT(ret)
}

func SHGetPropertyStoreForWindow(hwnd HWND, riid *syscall.GUID, ppv unsafe.Pointer) HRESULT {
	addr := lazyAddr(&pSHGetPropertyStoreForWindow, libShell32, "SHGetPropertyStoreForWindow")
	ret, _, _ := syscall.SyscallN(addr, hwnd, uintptr(unsafe.Pointer(riid)), uintptr(ppv))
	return HRESULT(ret)
}
