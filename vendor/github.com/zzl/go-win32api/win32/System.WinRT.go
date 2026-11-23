package win32

import (
	"syscall"
	"unsafe"
)

type (
	HSTRING                                = uintptr
	HSTRING_BUFFER                         = uintptr
	ROPARAMIIDHANDLE                       = uintptr
	APARTMENT_SHUTDOWN_REGISTRATION_COOKIE = uintptr
)

const (
	MAX_ERROR_MESSAGE_CHARS                             uint32 = 0x200
	CastingSourceInfo_Property_PreferredSourceUriScheme string = "PreferredSourceUriScheme"
	CastingSourceInfo_Property_CastingTypes             string = "CastingTypes"
	CastingSourceInfo_Property_ProtectedMedia           string = "ProtectedMedia"
)

// enums

// enum
type ACTIVATIONTYPE int32

const (
	ACTIVATIONTYPE_UNCATEGORIZED ACTIVATIONTYPE = 0
	ACTIVATIONTYPE_FROM_MONIKER  ACTIVATIONTYPE = 1
	ACTIVATIONTYPE_FROM_DATA     ACTIVATIONTYPE = 2
	ACTIVATIONTYPE_FROM_STORAGE  ACTIVATIONTYPE = 4
	ACTIVATIONTYPE_FROM_STREAM   ACTIVATIONTYPE = 8
	ACTIVATIONTYPE_FROM_FILE     ACTIVATIONTYPE = 16
)

// enum
type AgileReferenceOptions int32

const (
	AGILEREFERENCE_DEFAULT        AgileReferenceOptions = 0
	AGILEREFERENCE_DELAYEDMARSHAL AgileReferenceOptions = 1
)

// enum
type TrustLevel int32

const (
	BaseTrust    TrustLevel = 0
	PartialTrust TrustLevel = 1
	FullTrust    TrustLevel = 2
)

// enum
type DISPATCHERQUEUE_THREAD_APARTMENTTYPE int32

const (
	DQTAT_COM_NONE DISPATCHERQUEUE_THREAD_APARTMENTTYPE = 0
	DQTAT_COM_ASTA DISPATCHERQUEUE_THREAD_APARTMENTTYPE = 1
	DQTAT_COM_STA  DISPATCHERQUEUE_THREAD_APARTMENTTYPE = 2
)

// enum
type DISPATCHERQUEUE_THREAD_TYPE int32

const (
	DQTYPE_THREAD_DEDICATED DISPATCHERQUEUE_THREAD_TYPE = 1
	DQTYPE_THREAD_CURRENT   DISPATCHERQUEUE_THREAD_TYPE = 2
)

// enum
type CASTING_CONNECTION_ERROR_STATUS int32

const (
	CASTING_CONNECTION_ERROR_STATUS_SUCCEEDED                 CASTING_CONNECTION_ERROR_STATUS = 0
	CASTING_CONNECTION_ERROR_STATUS_DEVICE_DID_NOT_RESPOND    CASTING_CONNECTION_ERROR_STATUS = 1
	CASTING_CONNECTION_ERROR_STATUS_DEVICE_ERROR              CASTING_CONNECTION_ERROR_STATUS = 2
	CASTING_CONNECTION_ERROR_STATUS_DEVICE_LOCKED             CASTING_CONNECTION_ERROR_STATUS = 3
	CASTING_CONNECTION_ERROR_STATUS_PROTECTED_PLAYBACK_FAILED CASTING_CONNECTION_ERROR_STATUS = 4
	CASTING_CONNECTION_ERROR_STATUS_INVALID_CASTING_SOURCE    CASTING_CONNECTION_ERROR_STATUS = 5
	CASTING_CONNECTION_ERROR_STATUS_UNKNOWN                   CASTING_CONNECTION_ERROR_STATUS = 6
)

// enum
type CASTING_CONNECTION_STATE int32

const (
	CASTING_CONNECTION_STATE_DISCONNECTED  CASTING_CONNECTION_STATE = 0
	CASTING_CONNECTION_STATE_CONNECTED     CASTING_CONNECTION_STATE = 1
	CASTING_CONNECTION_STATE_RENDERING     CASTING_CONNECTION_STATE = 2
	CASTING_CONNECTION_STATE_DISCONNECTING CASTING_CONNECTION_STATE = 3
	CASTING_CONNECTION_STATE_CONNECTING    CASTING_CONNECTION_STATE = 4
)

// enum
type RO_INIT_TYPE int32

const (
	RO_INIT_SINGLETHREADED RO_INIT_TYPE = 0
	RO_INIT_MULTITHREADED  RO_INIT_TYPE = 1
)

// enum
// flags
type RO_ERROR_REPORTING_FLAGS uint32

const (
	RO_ERROR_REPORTING_NONE                 RO_ERROR_REPORTING_FLAGS = 0
	RO_ERROR_REPORTING_SUPPRESSEXCEPTIONS   RO_ERROR_REPORTING_FLAGS = 1
	RO_ERROR_REPORTING_FORCEEXCEPTIONS      RO_ERROR_REPORTING_FLAGS = 2
	RO_ERROR_REPORTING_USESETERRORINFO      RO_ERROR_REPORTING_FLAGS = 4
	RO_ERROR_REPORTING_SUPPRESSSETERRORINFO RO_ERROR_REPORTING_FLAGS = 8
)

// enum
type BSOS_OPTIONS int32

const (
	BSOS_DEFAULT                 BSOS_OPTIONS = 0
	BSOS_PREFERDESTINATIONSTREAM BSOS_OPTIONS = 1
)

// structs

type EventRegistrationToken struct {
	Value int64
}

type HSTRING_HEADER struct {
	Flags    uint32
	Length   uint32
	Padding1 uint32
	Padding2 uint32
	Data     uintptr
}

type ServerInformation struct {
	DwServerPid       uint32
	DwServerTid       uint32
	Ui64ServerAddress uint64
}

type DispatcherQueueOptions struct {
	DwSize        uint32
	ThreadType    DISPATCHERQUEUE_THREAD_TYPE
	ApartmentType DISPATCHERQUEUE_THREAD_APARTMENTTYPE
}

type RO_REGISTRATION_COOKIE_ struct {
}

// func types

type PINSPECT_HSTRING_CALLBACK = uintptr
type PINSPECT_HSTRING_CALLBACK_func = func(context unsafe.Pointer, readAddress uintptr, length uint32, buffer *byte) HRESULT

type PINSPECT_HSTRING_CALLBACK2 = uintptr
type PINSPECT_HSTRING_CALLBACK2_func = func(context unsafe.Pointer, readAddress uint64, length uint32, buffer *byte) HRESULT

type PINSPECT_MEMORY_CALLBACK = uintptr
type PINSPECT_MEMORY_CALLBACK_func = func(context unsafe.Pointer, readAddress uintptr, length uint32, buffer *byte) HRESULT

// interfaces

// C03F6A43-65A4-9818-987E-E0B810D2A6F2
var IID_IAgileReference = syscall.GUID{0xC03F6A43, 0x65A4, 0x9818,
	[8]byte{0x98, 0x7E, 0xE0, 0xB8, 0x10, 0xD2, 0xA6, 0xF2}}

type IAgileReferenceInterface interface {
	IUnknownInterface
	Resolve(riid *syscall.GUID, ppvObjectReference unsafe.Pointer) HRESULT
}

type IAgileReferenceVtbl struct {
	IUnknownVtbl
	Resolve uintptr
}

type IAgileReference struct {
	IUnknown
}

func (this *IAgileReference) Vtbl() *IAgileReferenceVtbl {
	return (*IAgileReferenceVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IAgileReference) Resolve(riid *syscall.GUID, ppvObjectReference unsafe.Pointer) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Resolve, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(riid)), uintptr(ppvObjectReference))
	return HRESULT(ret)
}

// A2F05A09-27A2-42B5-BC0E-AC163EF49D9B
var IID_IApartmentShutdown = syscall.GUID{0xA2F05A09, 0x27A2, 0x42B5,
	[8]byte{0xBC, 0x0E, 0xAC, 0x16, 0x3E, 0xF4, 0x9D, 0x9B}}

type IApartmentShutdownInterface interface {
	IUnknownInterface
	OnUninitialize(ui64ApartmentIdentifier uint64)
}

type IApartmentShutdownVtbl struct {
	IUnknownVtbl
	OnUninitialize uintptr
}

type IApartmentShutdown struct {
	IUnknown
}

func (this *IApartmentShutdown) Vtbl() *IApartmentShutdownVtbl {
	return (*IApartmentShutdownVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IApartmentShutdown) OnUninitialize(ui64ApartmentIdentifier uint64) {
	_, _, _ = syscall.SyscallN(this.Vtbl().OnUninitialize, uintptr(unsafe.Pointer(this)), uintptr(ui64ApartmentIdentifier))
}

// 5C4EE536-6A98-4B86-A170-587013D6FD4B
var IID_ISpatialInteractionManagerInterop = syscall.GUID{0x5C4EE536, 0x6A98, 0x4B86,
	[8]byte{0xA1, 0x70, 0x58, 0x70, 0x13, 0xD6, 0xFD, 0x4B}}

type ISpatialInteractionManagerInteropInterface interface {
	IInspectableInterface
	GetForWindow(window HWND, riid *syscall.GUID, spatialInteractionManager unsafe.Pointer) HRESULT
}

type ISpatialInteractionManagerInteropVtbl struct {
	IInspectableVtbl
	GetForWindow uintptr
}

type ISpatialInteractionManagerInterop struct {
	IInspectable
}

func (this *ISpatialInteractionManagerInterop) Vtbl() *ISpatialInteractionManagerInteropVtbl {
	return (*ISpatialInteractionManagerInteropVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *ISpatialInteractionManagerInterop) GetForWindow(window HWND, riid *syscall.GUID, spatialInteractionManager unsafe.Pointer) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetForWindow, uintptr(unsafe.Pointer(this)), window, uintptr(unsafe.Pointer(riid)), uintptr(spatialInteractionManager))
	return HRESULT(ret)
}

// 5C4EE536-6A98-4B86-A170-587013D6FD4B
var IID_IHolographicSpaceInterop = syscall.GUID{0x5C4EE536, 0x6A98, 0x4B86,
	[8]byte{0xA1, 0x70, 0x58, 0x70, 0x13, 0xD6, 0xFD, 0x4B}}

type IHolographicSpaceInteropInterface interface {
	IInspectableInterface
	CreateForWindow(window HWND, riid *syscall.GUID, holographicSpace unsafe.Pointer) HRESULT
}

type IHolographicSpaceInteropVtbl struct {
	IInspectableVtbl
	CreateForWindow uintptr
}

type IHolographicSpaceInterop struct {
	IInspectable
}

func (this *IHolographicSpaceInterop) Vtbl() *IHolographicSpaceInteropVtbl {
	return (*IHolographicSpaceInteropVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IHolographicSpaceInterop) CreateForWindow(window HWND, riid *syscall.GUID, holographicSpace unsafe.Pointer) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().CreateForWindow, uintptr(unsafe.Pointer(this)), window, uintptr(unsafe.Pointer(riid)), uintptr(holographicSpace))
	return HRESULT(ret)
}

// AF86E2E0-B12D-4C6A-9C5A-D7AA65101E90
var IID_IInspectable = syscall.GUID{0xAF86E2E0, 0xB12D, 0x4C6A,
	[8]byte{0x9C, 0x5A, 0xD7, 0xAA, 0x65, 0x10, 0x1E, 0x90}}

type IInspectableInterface interface {
	IUnknownInterface
	GetIids(iidCount *uint32, iids **syscall.GUID) HRESULT
	GetRuntimeClassName(className *HSTRING) HRESULT
	GetTrustLevel(trustLevel *TrustLevel) HRESULT
}

type IInspectableVtbl struct {
	IUnknownVtbl
	GetIids             uintptr
	GetRuntimeClassName uintptr
	GetTrustLevel       uintptr
}

type IInspectable struct {
	IUnknown
}

func (this *IInspectable) Vtbl() *IInspectableVtbl {
	return (*IInspectableVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IInspectable) GetIids(iidCount *uint32, iids **syscall.GUID) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetIids, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(iidCount)), uintptr(unsafe.Pointer(iids)))
	return HRESULT(ret)
}

func (this *IInspectable) GetRuntimeClassName(className *HSTRING) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetRuntimeClassName, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(className)))
	return HRESULT(ret)
}

func (this *IInspectable) GetTrustLevel(trustLevel *TrustLevel) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetTrustLevel, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(trustLevel)))
	return HRESULT(ret)
}

// D3EE12AD-3865-4362-9746-B75A682DF0E6
var IID_IAccountsSettingsPaneInterop = syscall.GUID{0xD3EE12AD, 0x3865, 0x4362,
	[8]byte{0x97, 0x46, 0xB7, 0x5A, 0x68, 0x2D, 0xF0, 0xE6}}

type IAccountsSettingsPaneInteropInterface interface {
	IInspectableInterface
	GetForWindow(appWindow HWND, riid *syscall.GUID, accountsSettingsPane unsafe.Pointer) HRESULT
	ShowManageAccountsForWindowAsync(appWindow HWND, riid *syscall.GUID, asyncAction unsafe.Pointer) HRESULT
	ShowAddAccountForWindowAsync(appWindow HWND, riid *syscall.GUID, asyncAction unsafe.Pointer) HRESULT
}

type IAccountsSettingsPaneInteropVtbl struct {
	IInspectableVtbl
	GetForWindow                     uintptr
	ShowManageAccountsForWindowAsync uintptr
	ShowAddAccountForWindowAsync     uintptr
}

type IAccountsSettingsPaneInterop struct {
	IInspectable
}

func (this *IAccountsSettingsPaneInterop) Vtbl() *IAccountsSettingsPaneInteropVtbl {
	return (*IAccountsSettingsPaneInteropVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IAccountsSettingsPaneInterop) GetForWindow(appWindow HWND, riid *syscall.GUID, accountsSettingsPane unsafe.Pointer) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetForWindow, uintptr(unsafe.Pointer(this)), appWindow, uintptr(unsafe.Pointer(riid)), uintptr(accountsSettingsPane))
	return HRESULT(ret)
}

func (this *IAccountsSettingsPaneInterop) ShowManageAccountsForWindowAsync(appWindow HWND, riid *syscall.GUID, asyncAction unsafe.Pointer) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().ShowManageAccountsForWindowAsync, uintptr(unsafe.Pointer(this)), appWindow, uintptr(unsafe.Pointer(riid)), uintptr(asyncAction))
	return HRESULT(ret)
}

func (this *IAccountsSettingsPaneInterop) ShowAddAccountForWindowAsync(appWindow HWND, riid *syscall.GUID, asyncAction unsafe.Pointer) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().ShowAddAccountForWindowAsync, uintptr(unsafe.Pointer(this)), appWindow, uintptr(unsafe.Pointer(riid)), uintptr(asyncAction))
	return HRESULT(ret)
}

// 65219584-F9CB-4AE3-81F9-A28A6CA450D9
var IID_IAppServiceConnectionExtendedExecution = syscall.GUID{0x65219584, 0xF9CB, 0x4AE3,
	[8]byte{0x81, 0xF9, 0xA2, 0x8A, 0x6C, 0xA4, 0x50, 0xD9}}

type IAppServiceConnectionExtendedExecutionInterface interface {
	IUnknownInterface
	OpenForExtendedExecutionAsync(riid *syscall.GUID, operation unsafe.Pointer) HRESULT
}

type IAppServiceConnectionExtendedExecutionVtbl struct {
	IUnknownVtbl
	OpenForExtendedExecutionAsync uintptr
}

type IAppServiceConnectionExtendedExecution struct {
	IUnknown
}

func (this *IAppServiceConnectionExtendedExecution) Vtbl() *IAppServiceConnectionExtendedExecutionVtbl {
	return (*IAppServiceConnectionExtendedExecutionVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IAppServiceConnectionExtendedExecution) OpenForExtendedExecutionAsync(riid *syscall.GUID, operation unsafe.Pointer) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().OpenForExtendedExecutionAsync, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(riid)), uintptr(operation))
	return HRESULT(ret)
}

// 152B8A3B-B9B9-4685-B56E-974847BC7545
var IID_ICorrelationVectorSource = syscall.GUID{0x152B8A3B, 0xB9B9, 0x4685,
	[8]byte{0xB5, 0x6E, 0x97, 0x48, 0x47, 0xBC, 0x75, 0x45}}

type ICorrelationVectorSourceInterface interface {
	IUnknownInterface
	Get_CorrelationVector(cv *HSTRING) HRESULT
}

type ICorrelationVectorSourceVtbl struct {
	IUnknownVtbl
	Get_CorrelationVector uintptr
}

type ICorrelationVectorSource struct {
	IUnknown
}

func (this *ICorrelationVectorSource) Vtbl() *ICorrelationVectorSourceVtbl {
	return (*ICorrelationVectorSourceVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *ICorrelationVectorSource) Get_CorrelationVector(cv *HSTRING) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CorrelationVector, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(cv)))
	return HRESULT(ret)
}

// C79A6CB7-BEBD-47A6-A2AD-4D45AD79C7BC
var IID_ICastingEventHandler = syscall.GUID{0xC79A6CB7, 0xBEBD, 0x47A6,
	[8]byte{0xA2, 0xAD, 0x4D, 0x45, 0xAD, 0x79, 0xC7, 0xBC}}

type ICastingEventHandlerInterface interface {
	IUnknownInterface
	OnStateChanged(newState CASTING_CONNECTION_STATE) HRESULT
	OnError(errorStatus CASTING_CONNECTION_ERROR_STATUS, errorMessage PWSTR) HRESULT
}

type ICastingEventHandlerVtbl struct {
	IUnknownVtbl
	OnStateChanged uintptr
	OnError        uintptr
}

type ICastingEventHandler struct {
	IUnknown
}

func (this *ICastingEventHandler) Vtbl() *ICastingEventHandlerVtbl {
	return (*ICastingEventHandlerVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *ICastingEventHandler) OnStateChanged(newState CASTING_CONNECTION_STATE) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().OnStateChanged, uintptr(unsafe.Pointer(this)), uintptr(newState))
	return HRESULT(ret)
}

func (this *ICastingEventHandler) OnError(errorStatus CASTING_CONNECTION_ERROR_STATUS, errorMessage PWSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().OnError, uintptr(unsafe.Pointer(this)), uintptr(errorStatus), uintptr(unsafe.Pointer(errorMessage)))
	return HRESULT(ret)
}

// F0A56423-A664-4FBD-8B43-409A45E8D9A1
var IID_ICastingController = syscall.GUID{0xF0A56423, 0xA664, 0x4FBD,
	[8]byte{0x8B, 0x43, 0x40, 0x9A, 0x45, 0xE8, 0xD9, 0xA1}}

type ICastingControllerInterface interface {
	IUnknownInterface
	Initialize(castingEngine *IUnknown, castingSource *IUnknown) HRESULT
	Connect() HRESULT
	Disconnect() HRESULT
	Advise(eventHandler *ICastingEventHandler, cookie *uint32) HRESULT
	UnAdvise(cookie uint32) HRESULT
}

type ICastingControllerVtbl struct {
	IUnknownVtbl
	Initialize uintptr
	Connect    uintptr
	Disconnect uintptr
	Advise     uintptr
	UnAdvise   uintptr
}

type ICastingController struct {
	IUnknown
}

func (this *ICastingController) Vtbl() *ICastingControllerVtbl {
	return (*ICastingControllerVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *ICastingController) Initialize(castingEngine *IUnknown, castingSource *IUnknown) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Initialize, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(castingEngine)), uintptr(unsafe.Pointer(castingSource)))
	return HRESULT(ret)
}

func (this *ICastingController) Connect() HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Connect, uintptr(unsafe.Pointer(this)))
	return HRESULT(ret)
}

func (this *ICastingController) Disconnect() HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Disconnect, uintptr(unsafe.Pointer(this)))
	return HRESULT(ret)
}

func (this *ICastingController) Advise(eventHandler *ICastingEventHandler, cookie *uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Advise, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(eventHandler)), uintptr(unsafe.Pointer(cookie)))
	return HRESULT(ret)
}

func (this *ICastingController) UnAdvise(cookie uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().UnAdvise, uintptr(unsafe.Pointer(this)), uintptr(cookie))
	return HRESULT(ret)
}

// 45101AB7-7C3A-4BCE-9500-12C09024B298
var IID_ICastingSourceInfo = syscall.GUID{0x45101AB7, 0x7C3A, 0x4BCE,
	[8]byte{0x95, 0x00, 0x12, 0xC0, 0x90, 0x24, 0xB2, 0x98}}

type ICastingSourceInfoInterface interface {
	IUnknownInterface
	GetController(controller **ICastingController) HRESULT
	GetProperties(props **INamedPropertyStore) HRESULT
}

type ICastingSourceInfoVtbl struct {
	IUnknownVtbl
	GetController uintptr
	GetProperties uintptr
}

type ICastingSourceInfo struct {
	IUnknown
}

func (this *ICastingSourceInfo) Vtbl() *ICastingSourceInfoVtbl {
	return (*ICastingSourceInfoVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *ICastingSourceInfo) GetController(controller **ICastingController) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetController, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(controller)))
	return HRESULT(ret)
}

func (this *ICastingSourceInfo) GetProperties(props **INamedPropertyStore) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetProperties, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(props)))
	return HRESULT(ret)
}

// 5AD8CBA7-4C01-4DAC-9074-827894292D63
var IID_IDragDropManagerInterop = syscall.GUID{0x5AD8CBA7, 0x4C01, 0x4DAC,
	[8]byte{0x90, 0x74, 0x82, 0x78, 0x94, 0x29, 0x2D, 0x63}}

type IDragDropManagerInteropInterface interface {
	IInspectableInterface
	GetForWindow(hwnd HWND, riid *syscall.GUID, ppv unsafe.Pointer) HRESULT
}

type IDragDropManagerInteropVtbl struct {
	IInspectableVtbl
	GetForWindow uintptr
}

type IDragDropManagerInterop struct {
	IInspectable
}

func (this *IDragDropManagerInterop) Vtbl() *IDragDropManagerInteropVtbl {
	return (*IDragDropManagerInteropVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IDragDropManagerInterop) GetForWindow(hwnd HWND, riid *syscall.GUID, ppv unsafe.Pointer) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetForWindow, uintptr(unsafe.Pointer(this)), hwnd, uintptr(unsafe.Pointer(riid)), uintptr(ppv))
	return HRESULT(ret)
}

// 75CF2C57-9195-4931-8332-F0B409E916AF
var IID_IInputPaneInterop = syscall.GUID{0x75CF2C57, 0x9195, 0x4931,
	[8]byte{0x83, 0x32, 0xF0, 0xB4, 0x09, 0xE9, 0x16, 0xAF}}

type IInputPaneInteropInterface interface {
	IInspectableInterface
	GetForWindow(appWindow HWND, riid *syscall.GUID, inputPane unsafe.Pointer) HRESULT
}

type IInputPaneInteropVtbl struct {
	IInspectableVtbl
	GetForWindow uintptr
}

type IInputPaneInterop struct {
	IInspectable
}

func (this *IInputPaneInterop) Vtbl() *IInputPaneInteropVtbl {
	return (*IInputPaneInteropVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IInputPaneInterop) GetForWindow(appWindow HWND, riid *syscall.GUID, inputPane unsafe.Pointer) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetForWindow, uintptr(unsafe.Pointer(this)), appWindow, uintptr(unsafe.Pointer(riid)), uintptr(inputPane))
	return HRESULT(ret)
}

// 24394699-1F2C-4EB3-8CD7-0EC1DA42A540
var IID_IPlayToManagerInterop = syscall.GUID{0x24394699, 0x1F2C, 0x4EB3,
	[8]byte{0x8C, 0xD7, 0x0E, 0xC1, 0xDA, 0x42, 0xA5, 0x40}}

type IPlayToManagerInteropInterface interface {
	IInspectableInterface
	GetForWindow(appWindow HWND, riid *syscall.GUID, playToManager unsafe.Pointer) HRESULT
	ShowPlayToUIForWindow(appWindow HWND) HRESULT
}

type IPlayToManagerInteropVtbl struct {
	IInspectableVtbl
	GetForWindow          uintptr
	ShowPlayToUIForWindow uintptr
}

type IPlayToManagerInterop struct {
	IInspectable
}

func (this *IPlayToManagerInterop) Vtbl() *IPlayToManagerInteropVtbl {
	return (*IPlayToManagerInteropVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IPlayToManagerInterop) GetForWindow(appWindow HWND, riid *syscall.GUID, playToManager unsafe.Pointer) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetForWindow, uintptr(unsafe.Pointer(this)), appWindow, uintptr(unsafe.Pointer(riid)), uintptr(playToManager))
	return HRESULT(ret)
}

func (this *IPlayToManagerInterop) ShowPlayToUIForWindow(appWindow HWND) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().ShowPlayToUIForWindow, uintptr(unsafe.Pointer(this)), appWindow)
	return HRESULT(ret)
}

// 83C78B3C-D88B-4950-AA6E-22B8D22AABD3
var IID_ICorrelationVectorInformation = syscall.GUID{0x83C78B3C, 0xD88B, 0x4950,
	[8]byte{0xAA, 0x6E, 0x22, 0xB8, 0xD2, 0x2A, 0xAB, 0xD3}}

type ICorrelationVectorInformationInterface interface {
	IInspectableInterface
	Get_LastCorrelationVectorForThread(cv *HSTRING) HRESULT
	Get_NextCorrelationVectorForThread(cv *HSTRING) HRESULT
	Put_NextCorrelationVectorForThread(cv HSTRING) HRESULT
}

type ICorrelationVectorInformationVtbl struct {
	IInspectableVtbl
	Get_LastCorrelationVectorForThread uintptr
	Get_NextCorrelationVectorForThread uintptr
	Put_NextCorrelationVectorForThread uintptr
}

type ICorrelationVectorInformation struct {
	IInspectable
}

func (this *ICorrelationVectorInformation) Vtbl() *ICorrelationVectorInformationVtbl {
	return (*ICorrelationVectorInformationVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *ICorrelationVectorInformation) Get_LastCorrelationVectorForThread(cv *HSTRING) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_LastCorrelationVectorForThread, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(cv)))
	return HRESULT(ret)
}

func (this *ICorrelationVectorInformation) Get_NextCorrelationVectorForThread(cv *HSTRING) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_NextCorrelationVectorForThread, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(cv)))
	return HRESULT(ret)
}

func (this *ICorrelationVectorInformation) Put_NextCorrelationVectorForThread(cv HSTRING) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Put_NextCorrelationVectorForThread, uintptr(unsafe.Pointer(this)), cv)
	return HRESULT(ret)
}

// 3694DBF9-8F68-44BE-8FF5-195C98EDE8A6
var IID_IUIViewSettingsInterop = syscall.GUID{0x3694DBF9, 0x8F68, 0x44BE,
	[8]byte{0x8F, 0xF5, 0x19, 0x5C, 0x98, 0xED, 0xE8, 0xA6}}

type IUIViewSettingsInteropInterface interface {
	IInspectableInterface
	GetForWindow(hwnd HWND, riid *syscall.GUID, ppv unsafe.Pointer) HRESULT
}

type IUIViewSettingsInteropVtbl struct {
	IInspectableVtbl
	GetForWindow uintptr
}

type IUIViewSettingsInterop struct {
	IInspectable
}

func (this *IUIViewSettingsInterop) Vtbl() *IUIViewSettingsInteropVtbl {
	return (*IUIViewSettingsInteropVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IUIViewSettingsInterop) GetForWindow(hwnd HWND, riid *syscall.GUID, ppv unsafe.Pointer) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetForWindow, uintptr(unsafe.Pointer(this)), hwnd, uintptr(unsafe.Pointer(riid)), uintptr(ppv))
	return HRESULT(ret)
}

// 1ADE314D-0E0A-40D9-824C-9A088A50059F
var IID_IUserActivityInterop = syscall.GUID{0x1ADE314D, 0x0E0A, 0x40D9,
	[8]byte{0x82, 0x4C, 0x9A, 0x08, 0x8A, 0x50, 0x05, 0x9F}}

type IUserActivityInteropInterface interface {
	IInspectableInterface
	CreateSessionForWindow(window HWND, iid *syscall.GUID, value unsafe.Pointer) HRESULT
}

type IUserActivityInteropVtbl struct {
	IInspectableVtbl
	CreateSessionForWindow uintptr
}

type IUserActivityInterop struct {
	IInspectable
}

func (this *IUserActivityInterop) Vtbl() *IUserActivityInteropVtbl {
	return (*IUserActivityInteropVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IUserActivityInterop) CreateSessionForWindow(window HWND, iid *syscall.GUID, value unsafe.Pointer) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().CreateSessionForWindow, uintptr(unsafe.Pointer(this)), window, uintptr(unsafe.Pointer(iid)), uintptr(value))
	return HRESULT(ret)
}

// C15DF8BC-8844-487A-B85B-7578E0F61419
var IID_IUserActivitySourceHostInterop = syscall.GUID{0xC15DF8BC, 0x8844, 0x487A,
	[8]byte{0xB8, 0x5B, 0x75, 0x78, 0xE0, 0xF6, 0x14, 0x19}}

type IUserActivitySourceHostInteropInterface interface {
	IInspectableInterface
	SetActivitySourceHost(activitySourceHost HSTRING) HRESULT
}

type IUserActivitySourceHostInteropVtbl struct {
	IInspectableVtbl
	SetActivitySourceHost uintptr
}

type IUserActivitySourceHostInterop struct {
	IInspectable
}

func (this *IUserActivitySourceHostInterop) Vtbl() *IUserActivitySourceHostInteropVtbl {
	return (*IUserActivitySourceHostInteropVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IUserActivitySourceHostInterop) SetActivitySourceHost(activitySourceHost HSTRING) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetActivitySourceHost, uintptr(unsafe.Pointer(this)), activitySourceHost)
	return HRESULT(ret)
}

// DD69F876-9699-4715-9095-E37EA30DFA1B
var IID_IUserActivityRequestManagerInterop = syscall.GUID{0xDD69F876, 0x9699, 0x4715,
	[8]byte{0x90, 0x95, 0xE3, 0x7E, 0xA3, 0x0D, 0xFA, 0x1B}}

type IUserActivityRequestManagerInteropInterface interface {
	IInspectableInterface
	GetForWindow(window HWND, iid *syscall.GUID, value unsafe.Pointer) HRESULT
}

type IUserActivityRequestManagerInteropVtbl struct {
	IInspectableVtbl
	GetForWindow uintptr
}

type IUserActivityRequestManagerInterop struct {
	IInspectable
}

func (this *IUserActivityRequestManagerInterop) Vtbl() *IUserActivityRequestManagerInteropVtbl {
	return (*IUserActivityRequestManagerInteropVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IUserActivityRequestManagerInterop) GetForWindow(window HWND, iid *syscall.GUID, value unsafe.Pointer) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetForWindow, uintptr(unsafe.Pointer(this)), window, uintptr(unsafe.Pointer(iid)), uintptr(value))
	return HRESULT(ret)
}

// 39E050C3-4E74-441A-8DC0-B81104DF949C
var IID_IUserConsentVerifierInterop = syscall.GUID{0x39E050C3, 0x4E74, 0x441A,
	[8]byte{0x8D, 0xC0, 0xB8, 0x11, 0x04, 0xDF, 0x94, 0x9C}}

type IUserConsentVerifierInteropInterface interface {
	IInspectableInterface
	RequestVerificationForWindowAsync(appWindow HWND, message HSTRING, riid *syscall.GUID, asyncOperation unsafe.Pointer) HRESULT
}

type IUserConsentVerifierInteropVtbl struct {
	IInspectableVtbl
	RequestVerificationForWindowAsync uintptr
}

type IUserConsentVerifierInterop struct {
	IInspectable
}

func (this *IUserConsentVerifierInterop) Vtbl() *IUserConsentVerifierInteropVtbl {
	return (*IUserConsentVerifierInteropVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IUserConsentVerifierInterop) RequestVerificationForWindowAsync(appWindow HWND, message HSTRING, riid *syscall.GUID, asyncOperation unsafe.Pointer) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().RequestVerificationForWindowAsync, uintptr(unsafe.Pointer(this)), appWindow, message, uintptr(unsafe.Pointer(riid)), uintptr(asyncOperation))
	return HRESULT(ret)
}

// F4B8E804-811E-4436-B69C-44CB67B72084
var IID_IWebAuthenticationCoreManagerInterop = syscall.GUID{0xF4B8E804, 0x811E, 0x4436,
	[8]byte{0xB6, 0x9C, 0x44, 0xCB, 0x67, 0xB7, 0x20, 0x84}}

type IWebAuthenticationCoreManagerInteropInterface interface {
	IInspectableInterface
	RequestTokenForWindowAsync(appWindow HWND, request *IInspectable, riid *syscall.GUID, asyncInfo unsafe.Pointer) HRESULT
	RequestTokenWithWebAccountForWindowAsync(appWindow HWND, request *IInspectable, webAccount *IInspectable, riid *syscall.GUID, asyncInfo unsafe.Pointer) HRESULT
}

type IWebAuthenticationCoreManagerInteropVtbl struct {
	IInspectableVtbl
	RequestTokenForWindowAsync               uintptr
	RequestTokenWithWebAccountForWindowAsync uintptr
}

type IWebAuthenticationCoreManagerInterop struct {
	IInspectable
}

func (this *IWebAuthenticationCoreManagerInterop) Vtbl() *IWebAuthenticationCoreManagerInteropVtbl {
	return (*IWebAuthenticationCoreManagerInteropVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IWebAuthenticationCoreManagerInterop) RequestTokenForWindowAsync(appWindow HWND, request *IInspectable, riid *syscall.GUID, asyncInfo unsafe.Pointer) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().RequestTokenForWindowAsync, uintptr(unsafe.Pointer(this)), appWindow, uintptr(unsafe.Pointer(request)), uintptr(unsafe.Pointer(riid)), uintptr(asyncInfo))
	return HRESULT(ret)
}

func (this *IWebAuthenticationCoreManagerInterop) RequestTokenWithWebAccountForWindowAsync(appWindow HWND, request *IInspectable, webAccount *IInspectable, riid *syscall.GUID, asyncInfo unsafe.Pointer) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().RequestTokenWithWebAccountForWindowAsync, uintptr(unsafe.Pointer(this)), appWindow, uintptr(unsafe.Pointer(request)), uintptr(unsafe.Pointer(webAccount)), uintptr(unsafe.Pointer(riid)), uintptr(asyncInfo))
	return HRESULT(ret)
}

// 82BA7092-4C88-427D-A7BC-16DD93FEB67E
var IID_IRestrictedErrorInfo = syscall.GUID{0x82BA7092, 0x4C88, 0x427D,
	[8]byte{0xA7, 0xBC, 0x16, 0xDD, 0x93, 0xFE, 0xB6, 0x7E}}

type IRestrictedErrorInfoInterface interface {
	IUnknownInterface
	GetErrorDetails(description *BSTR, error *HRESULT, restrictedDescription *BSTR, capabilitySid *BSTR) HRESULT
	GetReference(reference *BSTR) HRESULT
}

type IRestrictedErrorInfoVtbl struct {
	IUnknownVtbl
	GetErrorDetails uintptr
	GetReference    uintptr
}

type IRestrictedErrorInfo struct {
	IUnknown
}

func (this *IRestrictedErrorInfo) Vtbl() *IRestrictedErrorInfoVtbl {
	return (*IRestrictedErrorInfoVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IRestrictedErrorInfo) GetErrorDetails(description *BSTR, error *HRESULT, restrictedDescription *BSTR, capabilitySid *BSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetErrorDetails, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(description)), uintptr(unsafe.Pointer(error)), uintptr(unsafe.Pointer(restrictedDescription)), uintptr(unsafe.Pointer(capabilitySid)))
	return HRESULT(ret)
}

func (this *IRestrictedErrorInfo) GetReference(reference *BSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetReference, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(reference)))
	return HRESULT(ret)
}

// 04A2DBF3-DF83-116C-0946-0812ABF6E07D
var IID_ILanguageExceptionErrorInfo = syscall.GUID{0x04A2DBF3, 0xDF83, 0x116C,
	[8]byte{0x09, 0x46, 0x08, 0x12, 0xAB, 0xF6, 0xE0, 0x7D}}

type ILanguageExceptionErrorInfoInterface interface {
	IUnknownInterface
	GetLanguageException(languageException **IUnknown) HRESULT
}

type ILanguageExceptionErrorInfoVtbl struct {
	IUnknownVtbl
	GetLanguageException uintptr
}

type ILanguageExceptionErrorInfo struct {
	IUnknown
}

func (this *ILanguageExceptionErrorInfo) Vtbl() *ILanguageExceptionErrorInfoVtbl {
	return (*ILanguageExceptionErrorInfoVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *ILanguageExceptionErrorInfo) GetLanguageException(languageException **IUnknown) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetLanguageException, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(languageException)))
	return HRESULT(ret)
}

// FEB5A271-A6CD-45CE-880A-696706BADC65
var IID_ILanguageExceptionTransform = syscall.GUID{0xFEB5A271, 0xA6CD, 0x45CE,
	[8]byte{0x88, 0x0A, 0x69, 0x67, 0x06, 0xBA, 0xDC, 0x65}}

type ILanguageExceptionTransformInterface interface {
	IUnknownInterface
	GetTransformedRestrictedErrorInfo(restrictedErrorInfo **IRestrictedErrorInfo) HRESULT
}

type ILanguageExceptionTransformVtbl struct {
	IUnknownVtbl
	GetTransformedRestrictedErrorInfo uintptr
}

type ILanguageExceptionTransform struct {
	IUnknown
}

func (this *ILanguageExceptionTransform) Vtbl() *ILanguageExceptionTransformVtbl {
	return (*ILanguageExceptionTransformVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *ILanguageExceptionTransform) GetTransformedRestrictedErrorInfo(restrictedErrorInfo **IRestrictedErrorInfo) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetTransformedRestrictedErrorInfo, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(restrictedErrorInfo)))
	return HRESULT(ret)
}

// CBE53FB5-F967-4258-8D34-42F5E25833DE
var IID_ILanguageExceptionStackBackTrace = syscall.GUID{0xCBE53FB5, 0xF967, 0x4258,
	[8]byte{0x8D, 0x34, 0x42, 0xF5, 0xE2, 0x58, 0x33, 0xDE}}

type ILanguageExceptionStackBackTraceInterface interface {
	IUnknownInterface
	GetStackBackTrace(maxFramesToCapture uint32, stackBackTrace *uintptr, framesCaptured *uint32) HRESULT
}

type ILanguageExceptionStackBackTraceVtbl struct {
	IUnknownVtbl
	GetStackBackTrace uintptr
}

type ILanguageExceptionStackBackTrace struct {
	IUnknown
}

func (this *ILanguageExceptionStackBackTrace) Vtbl() *ILanguageExceptionStackBackTraceVtbl {
	return (*ILanguageExceptionStackBackTraceVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *ILanguageExceptionStackBackTrace) GetStackBackTrace(maxFramesToCapture uint32, stackBackTrace *uintptr, framesCaptured *uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetStackBackTrace, uintptr(unsafe.Pointer(this)), uintptr(maxFramesToCapture), uintptr(unsafe.Pointer(stackBackTrace)), uintptr(unsafe.Pointer(framesCaptured)))
	return HRESULT(ret)
}

// 5746E5C4-5B97-424C-B620-2822915734DD
var IID_ILanguageExceptionErrorInfo2 = syscall.GUID{0x5746E5C4, 0x5B97, 0x424C,
	[8]byte{0xB6, 0x20, 0x28, 0x22, 0x91, 0x57, 0x34, 0xDD}}

type ILanguageExceptionErrorInfo2Interface interface {
	ILanguageExceptionErrorInfoInterface
	GetPreviousLanguageExceptionErrorInfo(previousLanguageExceptionErrorInfo **ILanguageExceptionErrorInfo2) HRESULT
	CapturePropagationContext(languageException *IUnknown) HRESULT
	GetPropagationContextHead(propagatedLanguageExceptionErrorInfoHead **ILanguageExceptionErrorInfo2) HRESULT
}

type ILanguageExceptionErrorInfo2Vtbl struct {
	ILanguageExceptionErrorInfoVtbl
	GetPreviousLanguageExceptionErrorInfo uintptr
	CapturePropagationContext             uintptr
	GetPropagationContextHead             uintptr
}

type ILanguageExceptionErrorInfo2 struct {
	ILanguageExceptionErrorInfo
}

func (this *ILanguageExceptionErrorInfo2) Vtbl() *ILanguageExceptionErrorInfo2Vtbl {
	return (*ILanguageExceptionErrorInfo2Vtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *ILanguageExceptionErrorInfo2) GetPreviousLanguageExceptionErrorInfo(previousLanguageExceptionErrorInfo **ILanguageExceptionErrorInfo2) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetPreviousLanguageExceptionErrorInfo, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(previousLanguageExceptionErrorInfo)))
	return HRESULT(ret)
}

func (this *ILanguageExceptionErrorInfo2) CapturePropagationContext(languageException *IUnknown) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().CapturePropagationContext, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(languageException)))
	return HRESULT(ret)
}

func (this *ILanguageExceptionErrorInfo2) GetPropagationContextHead(propagatedLanguageExceptionErrorInfoHead **ILanguageExceptionErrorInfo2) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetPropagationContextHead, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(propagatedLanguageExceptionErrorInfoHead)))
	return HRESULT(ret)
}

// 00000035-0000-0000-C000-000000000046
var IID_IActivationFactory = syscall.GUID{0x00000035, 0x0000, 0x0000,
	[8]byte{0xC0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x46}}

type IActivationFactoryInterface interface {
	IInspectableInterface
	ActivateInstance(instance **IInspectable) HRESULT
}

type IActivationFactoryVtbl struct {
	IInspectableVtbl
	ActivateInstance uintptr
}

type IActivationFactory struct {
	IInspectable
}

func (this *IActivationFactory) Vtbl() *IActivationFactoryVtbl {
	return (*IActivationFactoryVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IActivationFactory) ActivateInstance(instance **IInspectable) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().ActivateInstance, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(instance)))
	return HRESULT(ret)
}

// 905A0FEF-BC53-11DF-8C49-001E4FC686DA
var IID_IBufferByteAccess = syscall.GUID{0x905A0FEF, 0xBC53, 0x11DF,
	[8]byte{0x8C, 0x49, 0x00, 0x1E, 0x4F, 0xC6, 0x86, 0xDA}}

type IBufferByteAccessInterface interface {
	IUnknownInterface
	Buffer(value **byte) HRESULT
}

type IBufferByteAccessVtbl struct {
	IUnknownVtbl
	Buffer uintptr
}

type IBufferByteAccess struct {
	IUnknown
}

func (this *IBufferByteAccess) Vtbl() *IBufferByteAccessVtbl {
	return (*IBufferByteAccessVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IBufferByteAccess) Buffer(value **byte) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Buffer, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(value)))
	return HRESULT(ret)
}

// 00000000-0000-0000-0000-000000000000
var IID_IRoSimpleMetaDataBuilder = syscall.GUID{0x00000000, 0x0000, 0x0000,
	[8]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}}

type IRoSimpleMetaDataBuilderInterface interface {
	SetWinRtInterface(iid syscall.GUID) HRESULT
	SetDelegate(iid syscall.GUID) HRESULT
	SetInterfaceGroupSimpleDefault(name PWSTR, defaultInterfaceName PWSTR, defaultInterfaceIID *syscall.GUID) HRESULT
	SetInterfaceGroupParameterizedDefault(name PWSTR, elementCount uint32, defaultInterfaceNameElements *PWSTR) HRESULT
	SetRuntimeClassSimpleDefault(name PWSTR, defaultInterfaceName PWSTR, defaultInterfaceIID *syscall.GUID) HRESULT
	SetRuntimeClassParameterizedDefault(name PWSTR, elementCount uint32, defaultInterfaceNameElements *PWSTR) HRESULT
	SetStruct(name PWSTR, numFields uint32, fieldTypeNames *PWSTR) HRESULT
	SetEnum(name PWSTR, baseType PWSTR) HRESULT
	SetParameterizedInterface(piid syscall.GUID, numArgs uint32) HRESULT
	SetParameterizedDelegate(piid syscall.GUID, numArgs uint32) HRESULT
}

type IRoSimpleMetaDataBuilderVtbl struct {
	SetWinRtInterface                     uintptr
	SetDelegate                           uintptr
	SetInterfaceGroupSimpleDefault        uintptr
	SetInterfaceGroupParameterizedDefault uintptr
	SetRuntimeClassSimpleDefault          uintptr
	SetRuntimeClassParameterizedDefault   uintptr
	SetStruct                             uintptr
	SetEnum                               uintptr
	SetParameterizedInterface             uintptr
	SetParameterizedDelegate              uintptr
}

type IRoSimpleMetaDataBuilder struct {
	LpVtbl *[1024]uintptr
}

func (this *IRoSimpleMetaDataBuilder) Vtbl() *IRoSimpleMetaDataBuilderVtbl {
	return (*IRoSimpleMetaDataBuilderVtbl)(unsafe.Pointer(this.LpVtbl))
}

func (this *IRoSimpleMetaDataBuilder) SetWinRtInterface(iid syscall.GUID) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetWinRtInterface, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(&iid)))
	return HRESULT(ret)
}

func (this *IRoSimpleMetaDataBuilder) SetDelegate(iid syscall.GUID) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetDelegate, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(&iid)))
	return HRESULT(ret)
}

func (this *IRoSimpleMetaDataBuilder) SetInterfaceGroupSimpleDefault(name PWSTR, defaultInterfaceName PWSTR, defaultInterfaceIID *syscall.GUID) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetInterfaceGroupSimpleDefault, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(name)), uintptr(unsafe.Pointer(defaultInterfaceName)), uintptr(unsafe.Pointer(defaultInterfaceIID)))
	return HRESULT(ret)
}

func (this *IRoSimpleMetaDataBuilder) SetInterfaceGroupParameterizedDefault(name PWSTR, elementCount uint32, defaultInterfaceNameElements *PWSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetInterfaceGroupParameterizedDefault, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(name)), uintptr(elementCount), uintptr(unsafe.Pointer(defaultInterfaceNameElements)))
	return HRESULT(ret)
}

func (this *IRoSimpleMetaDataBuilder) SetRuntimeClassSimpleDefault(name PWSTR, defaultInterfaceName PWSTR, defaultInterfaceIID *syscall.GUID) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetRuntimeClassSimpleDefault, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(name)), uintptr(unsafe.Pointer(defaultInterfaceName)), uintptr(unsafe.Pointer(defaultInterfaceIID)))
	return HRESULT(ret)
}

func (this *IRoSimpleMetaDataBuilder) SetRuntimeClassParameterizedDefault(name PWSTR, elementCount uint32, defaultInterfaceNameElements *PWSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetRuntimeClassParameterizedDefault, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(name)), uintptr(elementCount), uintptr(unsafe.Pointer(defaultInterfaceNameElements)))
	return HRESULT(ret)
}

func (this *IRoSimpleMetaDataBuilder) SetStruct(name PWSTR, numFields uint32, fieldTypeNames *PWSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetStruct, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(name)), uintptr(numFields), uintptr(unsafe.Pointer(fieldTypeNames)))
	return HRESULT(ret)
}

func (this *IRoSimpleMetaDataBuilder) SetEnum(name PWSTR, baseType PWSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetEnum, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(name)), uintptr(unsafe.Pointer(baseType)))
	return HRESULT(ret)
}

func (this *IRoSimpleMetaDataBuilder) SetParameterizedInterface(piid syscall.GUID, numArgs uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetParameterizedInterface, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(&piid)), uintptr(numArgs))
	return HRESULT(ret)
}

func (this *IRoSimpleMetaDataBuilder) SetParameterizedDelegate(piid syscall.GUID, numArgs uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetParameterizedDelegate, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(&piid)), uintptr(numArgs))
	return HRESULT(ret)
}

// 00000000-0000-0000-0000-000000000000
var IID_IRoMetaDataLocator = syscall.GUID{0x00000000, 0x0000, 0x0000,
	[8]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}}

type IRoMetaDataLocatorInterface interface {
	Locate(nameElement PWSTR, metaDataDestination *IRoSimpleMetaDataBuilder) HRESULT
}

type IRoMetaDataLocatorVtbl struct {
	Locate uintptr
}

type IRoMetaDataLocator struct {
	LpVtbl *[1024]uintptr
}

func (this *IRoMetaDataLocator) Vtbl() *IRoMetaDataLocatorVtbl {
	return (*IRoMetaDataLocatorVtbl)(unsafe.Pointer(this.LpVtbl))
}

func (this *IRoMetaDataLocator) Locate(nameElement PWSTR, metaDataDestination *IRoSimpleMetaDataBuilder) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Locate, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(nameElement)), uintptr(unsafe.Pointer(metaDataDestination)))
	return HRESULT(ret)
}

// 5B0D3235-4DBA-4D44-865E-8F1D0E4FD04D
var IID_IMemoryBufferByteAccess = syscall.GUID{0x5B0D3235, 0x4DBA, 0x4D44,
	[8]byte{0x86, 0x5E, 0x8F, 0x1D, 0x0E, 0x4F, 0xD0, 0x4D}}

type IMemoryBufferByteAccessInterface interface {
	IUnknownInterface
	GetBuffer(value **byte, capacity *uint32) HRESULT
}

type IMemoryBufferByteAccessVtbl struct {
	IUnknownVtbl
	GetBuffer uintptr
}

type IMemoryBufferByteAccess struct {
	IUnknown
}

func (this *IMemoryBufferByteAccess) Vtbl() *IMemoryBufferByteAccessVtbl {
	return (*IMemoryBufferByteAccessVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IMemoryBufferByteAccess) GetBuffer(value **byte, capacity *uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetBuffer, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(value)), uintptr(unsafe.Pointer(capacity)))
	return HRESULT(ret)
}

// 00000037-0000-0000-C000-000000000046
var IID_IWeakReference = syscall.GUID{0x00000037, 0x0000, 0x0000,
	[8]byte{0xC0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x46}}

type IWeakReferenceInterface interface {
	IUnknownInterface
	Resolve(riid *syscall.GUID, objectReference unsafe.Pointer) HRESULT
}

type IWeakReferenceVtbl struct {
	IUnknownVtbl
	Resolve uintptr
}

type IWeakReference struct {
	IUnknown
}

func (this *IWeakReference) Vtbl() *IWeakReferenceVtbl {
	return (*IWeakReferenceVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IWeakReference) Resolve(riid *syscall.GUID, objectReference unsafe.Pointer) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Resolve, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(riid)), uintptr(objectReference))
	return HRESULT(ret)
}

// 00000038-0000-0000-C000-000000000046
var IID_IWeakReferenceSource = syscall.GUID{0x00000038, 0x0000, 0x0000,
	[8]byte{0xC0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x46}}

type IWeakReferenceSourceInterface interface {
	IUnknownInterface
	GetWeakReference(weakReference **IWeakReference) HRESULT
}

type IWeakReferenceSourceVtbl struct {
	IUnknownVtbl
	GetWeakReference uintptr
}

type IWeakReferenceSource struct {
	IUnknown
}

func (this *IWeakReferenceSource) Vtbl() *IWeakReferenceSourceVtbl {
	return (*IWeakReferenceSourceVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IWeakReferenceSource) GetWeakReference(weakReference **IWeakReference) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetWeakReference, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(weakReference)))
	return HRESULT(ret)
}

// DDB0472D-C911-4A1F-86D9-DC3D71A95F5A
var IID_ISystemMediaTransportControlsInterop = syscall.GUID{0xDDB0472D, 0xC911, 0x4A1F,
	[8]byte{0x86, 0xD9, 0xDC, 0x3D, 0x71, 0xA9, 0x5F, 0x5A}}

type ISystemMediaTransportControlsInteropInterface interface {
	IInspectableInterface
	GetForWindow(appWindow HWND, riid *syscall.GUID, mediaTransportControl unsafe.Pointer) HRESULT
}

type ISystemMediaTransportControlsInteropVtbl struct {
	IInspectableVtbl
	GetForWindow uintptr
}

type ISystemMediaTransportControlsInterop struct {
	IInspectable
}

func (this *ISystemMediaTransportControlsInterop) Vtbl() *ISystemMediaTransportControlsInteropVtbl {
	return (*ISystemMediaTransportControlsInteropVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *ISystemMediaTransportControlsInterop) GetForWindow(appWindow HWND, riid *syscall.GUID, mediaTransportControl unsafe.Pointer) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetForWindow, uintptr(unsafe.Pointer(this)), appWindow, uintptr(unsafe.Pointer(riid)), uintptr(mediaTransportControl))
	return HRESULT(ret)
}

// 6571A721-643D-43D4-ACA4-6B6F5F30F1AD
var IID_IShareWindowCommandEventArgsInterop = syscall.GUID{0x6571A721, 0x643D, 0x43D4,
	[8]byte{0xAC, 0xA4, 0x6B, 0x6F, 0x5F, 0x30, 0xF1, 0xAD}}

type IShareWindowCommandEventArgsInteropInterface interface {
	IUnknownInterface
	GetWindow(value *HWND) HRESULT
}

type IShareWindowCommandEventArgsInteropVtbl struct {
	IUnknownVtbl
	GetWindow uintptr
}

type IShareWindowCommandEventArgsInterop struct {
	IUnknown
}

func (this *IShareWindowCommandEventArgsInterop) Vtbl() *IShareWindowCommandEventArgsInteropVtbl {
	return (*IShareWindowCommandEventArgsInteropVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IShareWindowCommandEventArgsInterop) GetWindow(value *HWND) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetWindow, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(value)))
	return HRESULT(ret)
}

// 461A191F-8424-43A6-A0FA-3451A22F56AB
var IID_IShareWindowCommandSourceInterop = syscall.GUID{0x461A191F, 0x8424, 0x43A6,
	[8]byte{0xA0, 0xFA, 0x34, 0x51, 0xA2, 0x2F, 0x56, 0xAB}}

type IShareWindowCommandSourceInteropInterface interface {
	IUnknownInterface
	GetForWindow(appWindow HWND, riid *syscall.GUID, shareWindowCommandSource unsafe.Pointer) HRESULT
}

type IShareWindowCommandSourceInteropVtbl struct {
	IUnknownVtbl
	GetForWindow uintptr
}

type IShareWindowCommandSourceInterop struct {
	IUnknown
}

func (this *IShareWindowCommandSourceInterop) Vtbl() *IShareWindowCommandSourceInteropVtbl {
	return (*IShareWindowCommandSourceInteropVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IShareWindowCommandSourceInterop) GetForWindow(appWindow HWND, riid *syscall.GUID, shareWindowCommandSource unsafe.Pointer) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetForWindow, uintptr(unsafe.Pointer(this)), appWindow, uintptr(unsafe.Pointer(riid)), uintptr(shareWindowCommandSource))
	return HRESULT(ret)
}

// F5F84C8F-CFD0-4CD6-B66B-C5D26FF1689D
var IID_IMessageDispatcher = syscall.GUID{0xF5F84C8F, 0xCFD0, 0x4CD6,
	[8]byte{0xB6, 0x6B, 0xC5, 0xD2, 0x6F, 0xF1, 0x68, 0x9D}}

type IMessageDispatcherInterface interface {
	IInspectableInterface
	PumpMessages() HRESULT
}

type IMessageDispatcherVtbl struct {
	IInspectableVtbl
	PumpMessages uintptr
}

type IMessageDispatcher struct {
	IInspectable
}

func (this *IMessageDispatcher) Vtbl() *IMessageDispatcherVtbl {
	return (*IMessageDispatcherVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IMessageDispatcher) PumpMessages() HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().PumpMessages, uintptr(unsafe.Pointer(this)))
	return HRESULT(ret)
}

// 45D64A29-A63E-4CB6-B498-5781D298CB4F
var IID_ICoreWindowInterop = syscall.GUID{0x45D64A29, 0xA63E, 0x4CB6,
	[8]byte{0xB4, 0x98, 0x57, 0x81, 0xD2, 0x98, 0xCB, 0x4F}}

type ICoreWindowInteropInterface interface {
	IUnknownInterface
	Get_WindowHandle(hwnd *HWND) HRESULT
	Put_MessageHandled(value byte) HRESULT
}

type ICoreWindowInteropVtbl struct {
	IUnknownVtbl
	Get_WindowHandle   uintptr
	Put_MessageHandled uintptr
}

type ICoreWindowInterop struct {
	IUnknown
}

func (this *ICoreWindowInterop) Vtbl() *ICoreWindowInteropVtbl {
	return (*ICoreWindowInteropVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *ICoreWindowInterop) Get_WindowHandle(hwnd *HWND) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_WindowHandle, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(hwnd)))
	return HRESULT(ret)
}

func (this *ICoreWindowInterop) Put_MessageHandled(value byte) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Put_MessageHandled, uintptr(unsafe.Pointer(this)), uintptr(value))
	return HRESULT(ret)
}

// 40BFE3E3-B75A-4479-AC96-475365749BB8
var IID_ICoreInputInterop = syscall.GUID{0x40BFE3E3, 0xB75A, 0x4479,
	[8]byte{0xAC, 0x96, 0x47, 0x53, 0x65, 0x74, 0x9B, 0xB8}}

type ICoreInputInteropInterface interface {
	IUnknownInterface
	SetInputSource(value *IUnknown) HRESULT
	Put_MessageHandled(value byte) HRESULT
}

type ICoreInputInteropVtbl struct {
	IUnknownVtbl
	SetInputSource     uintptr
	Put_MessageHandled uintptr
}

type ICoreInputInterop struct {
	IUnknown
}

func (this *ICoreInputInterop) Vtbl() *ICoreInputInteropVtbl {
	return (*ICoreInputInteropVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *ICoreInputInterop) SetInputSource(value *IUnknown) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetInputSource, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(value)))
	return HRESULT(ret)
}

func (this *ICoreInputInterop) Put_MessageHandled(value byte) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Put_MessageHandled, uintptr(unsafe.Pointer(this)), uintptr(value))
	return HRESULT(ret)
}

// 0576AB31-A310-4C40-BA31-FD37E0298DFA
var IID_ICoreWindowComponentInterop = syscall.GUID{0x0576AB31, 0xA310, 0x4C40,
	[8]byte{0xBA, 0x31, 0xFD, 0x37, 0xE0, 0x29, 0x8D, 0xFA}}

type ICoreWindowComponentInteropInterface interface {
	IUnknownInterface
	ConfigureComponentInput(hostViewInstanceId uint32, hwndHost HWND, inputSourceVisual *IUnknown) HRESULT
	GetViewInstanceId(componentViewInstanceId *uint32) HRESULT
}

type ICoreWindowComponentInteropVtbl struct {
	IUnknownVtbl
	ConfigureComponentInput uintptr
	GetViewInstanceId       uintptr
}

type ICoreWindowComponentInterop struct {
	IUnknown
}

func (this *ICoreWindowComponentInterop) Vtbl() *ICoreWindowComponentInteropVtbl {
	return (*ICoreWindowComponentInteropVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *ICoreWindowComponentInterop) ConfigureComponentInput(hostViewInstanceId uint32, hwndHost HWND, inputSourceVisual *IUnknown) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().ConfigureComponentInput, uintptr(unsafe.Pointer(this)), uintptr(hostViewInstanceId), hwndHost, uintptr(unsafe.Pointer(inputSourceVisual)))
	return HRESULT(ret)
}

func (this *ICoreWindowComponentInterop) GetViewInstanceId(componentViewInstanceId *uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetViewInstanceId, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(componentViewInstanceId)))
	return HRESULT(ret)
}

// 7A5B6FD1-CD73-4B6C-9CF4-2E869EAF470A
var IID_ICoreWindowAdapterInterop = syscall.GUID{0x7A5B6FD1, 0xCD73, 0x4B6C,
	[8]byte{0x9C, 0xF4, 0x2E, 0x86, 0x9E, 0xAF, 0x47, 0x0A}}

type ICoreWindowAdapterInteropInterface interface {
	IInspectableInterface
	Get_AppActivationClientAdapter(value **IUnknown) HRESULT
	Get_ApplicationViewClientAdapter(value **IUnknown) HRESULT
	Get_CoreApplicationViewClientAdapter(value **IUnknown) HRESULT
	Get_HoloViewClientAdapter(value **IUnknown) HRESULT
	Get_PositionerClientAdapter(value **IUnknown) HRESULT
	Get_SystemNavigationClientAdapter(value **IUnknown) HRESULT
	Get_TitleBarClientAdapter(value **IUnknown) HRESULT
	SetWindowClientAdapter(value *IUnknown) HRESULT
}

type ICoreWindowAdapterInteropVtbl struct {
	IInspectableVtbl
	Get_AppActivationClientAdapter       uintptr
	Get_ApplicationViewClientAdapter     uintptr
	Get_CoreApplicationViewClientAdapter uintptr
	Get_HoloViewClientAdapter            uintptr
	Get_PositionerClientAdapter          uintptr
	Get_SystemNavigationClientAdapter    uintptr
	Get_TitleBarClientAdapter            uintptr
	SetWindowClientAdapter               uintptr
}

type ICoreWindowAdapterInterop struct {
	IInspectable
}

func (this *ICoreWindowAdapterInterop) Vtbl() *ICoreWindowAdapterInteropVtbl {
	return (*ICoreWindowAdapterInteropVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *ICoreWindowAdapterInterop) Get_AppActivationClientAdapter(value **IUnknown) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_AppActivationClientAdapter, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(value)))
	return HRESULT(ret)
}

func (this *ICoreWindowAdapterInterop) Get_ApplicationViewClientAdapter(value **IUnknown) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_ApplicationViewClientAdapter, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(value)))
	return HRESULT(ret)
}

func (this *ICoreWindowAdapterInterop) Get_CoreApplicationViewClientAdapter(value **IUnknown) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CoreApplicationViewClientAdapter, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(value)))
	return HRESULT(ret)
}

func (this *ICoreWindowAdapterInterop) Get_HoloViewClientAdapter(value **IUnknown) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_HoloViewClientAdapter, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(value)))
	return HRESULT(ret)
}

func (this *ICoreWindowAdapterInterop) Get_PositionerClientAdapter(value **IUnknown) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_PositionerClientAdapter, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(value)))
	return HRESULT(ret)
}

func (this *ICoreWindowAdapterInterop) Get_SystemNavigationClientAdapter(value **IUnknown) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_SystemNavigationClientAdapter, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(value)))
	return HRESULT(ret)
}

func (this *ICoreWindowAdapterInterop) Get_TitleBarClientAdapter(value **IUnknown) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_TitleBarClientAdapter, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(value)))
	return HRESULT(ret)
}

func (this *ICoreWindowAdapterInterop) SetWindowClientAdapter(value *IUnknown) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetWindowClientAdapter, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(value)))
	return HRESULT(ret)
}

var (
	pCoDecodeProxy                       uintptr
	pRoGetAgileReference                 uintptr
	pHSTRING_UserSize                    uintptr
	pHSTRING_UserMarshal                 uintptr
	pHSTRING_UserUnmarshal               uintptr
	pHSTRING_UserFree                    uintptr
	pHSTRING_UserSize64                  uintptr
	pHSTRING_UserMarshal64               uintptr
	pHSTRING_UserUnmarshal64             uintptr
	pHSTRING_UserFree64                  uintptr
	pWindowsCreateString                 uintptr
	pWindowsCreateStringReference        uintptr
	pWindowsDeleteString                 uintptr
	pWindowsDuplicateString              uintptr
	pWindowsGetStringLen                 uintptr
	pWindowsGetStringRawBuffer           uintptr
	pWindowsIsStringEmpty                uintptr
	pWindowsStringHasEmbeddedNull        uintptr
	pWindowsCompareStringOrdinal         uintptr
	pWindowsSubstring                    uintptr
	pWindowsSubstringWithSpecifiedLength uintptr
	pWindowsConcatString                 uintptr
	pWindowsReplaceString                uintptr
	pWindowsTrimStringStart              uintptr
	pWindowsTrimStringEnd                uintptr
	pWindowsPreallocateStringBuffer      uintptr
	pWindowsPromoteStringBuffer          uintptr
	pWindowsDeleteStringBuffer           uintptr
	pWindowsInspectString                uintptr
	pRoInitialize                        uintptr
	pRoUninitialize                      uintptr
	pRoActivateInstance                  uintptr
	pRoRegisterActivationFactories       uintptr
	pRoRevokeActivationFactories         uintptr
	pRoGetActivationFactory              uintptr
	pRoRegisterForApartmentShutdown      uintptr
	pRoUnregisterForApartmentShutdown    uintptr
	pRoGetApartmentIdentifier            uintptr
)

func CoDecodeProxy(dwClientPid uint32, ui64ProxyAddress uint64, pServerInformation *ServerInformation) HRESULT {
	addr := lazyAddr(&pCoDecodeProxy, libOle32, "CoDecodeProxy")
	ret, _, _ := syscall.SyscallN(addr, uintptr(dwClientPid), uintptr(ui64ProxyAddress), uintptr(unsafe.Pointer(pServerInformation)))
	return HRESULT(ret)
}

func RoGetAgileReference(options AgileReferenceOptions, riid *syscall.GUID, pUnk *IUnknown, ppAgileReference **IAgileReference) HRESULT {
	addr := lazyAddr(&pRoGetAgileReference, libOle32, "RoGetAgileReference")
	ret, _, _ := syscall.SyscallN(addr, uintptr(options), uintptr(unsafe.Pointer(riid)), uintptr(unsafe.Pointer(pUnk)), uintptr(unsafe.Pointer(ppAgileReference)))
	return HRESULT(ret)
}

func HSTRING_UserSize(param0 *uint32, param1 uint32, param2 *HSTRING) uint32 {
	addr := lazyAddr(&pHSTRING_UserSize, libApi_ms_win_core_winrt_string_l1_1_0, "HSTRING_UserSize")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(param0)), uintptr(param1), uintptr(unsafe.Pointer(param2)))
	return uint32(ret)
}

func HSTRING_UserMarshal(param0 *uint32, param1 *byte, param2 *HSTRING) *byte {
	addr := lazyAddr(&pHSTRING_UserMarshal, libApi_ms_win_core_winrt_string_l1_1_0, "HSTRING_UserMarshal")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(param0)), uintptr(unsafe.Pointer(param1)), uintptr(unsafe.Pointer(param2)))
	return (*byte)(unsafe.Pointer(ret))
}

func HSTRING_UserUnmarshal(param0 *uint32, param1 *byte, param2 *HSTRING) *byte {
	addr := lazyAddr(&pHSTRING_UserUnmarshal, libApi_ms_win_core_winrt_string_l1_1_0, "HSTRING_UserUnmarshal")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(param0)), uintptr(unsafe.Pointer(param1)), uintptr(unsafe.Pointer(param2)))
	return (*byte)(unsafe.Pointer(ret))
}

func HSTRING_UserFree(param0 *uint32, param1 *HSTRING) {
	addr := lazyAddr(&pHSTRING_UserFree, libApi_ms_win_core_winrt_string_l1_1_0, "HSTRING_UserFree")
	syscall.SyscallN(addr, uintptr(unsafe.Pointer(param0)), uintptr(unsafe.Pointer(param1)))
}

func HSTRING_UserSize64(param0 *uint32, param1 uint32, param2 *HSTRING) uint32 {
	addr := lazyAddr(&pHSTRING_UserSize64, libApi_ms_win_core_winrt_string_l1_1_0, "HSTRING_UserSize64")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(param0)), uintptr(param1), uintptr(unsafe.Pointer(param2)))
	return uint32(ret)
}

func HSTRING_UserMarshal64(param0 *uint32, param1 *byte, param2 *HSTRING) *byte {
	addr := lazyAddr(&pHSTRING_UserMarshal64, libApi_ms_win_core_winrt_string_l1_1_0, "HSTRING_UserMarshal64")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(param0)), uintptr(unsafe.Pointer(param1)), uintptr(unsafe.Pointer(param2)))
	return (*byte)(unsafe.Pointer(ret))
}

func HSTRING_UserUnmarshal64(param0 *uint32, param1 *byte, param2 *HSTRING) *byte {
	addr := lazyAddr(&pHSTRING_UserUnmarshal64, libApi_ms_win_core_winrt_string_l1_1_0, "HSTRING_UserUnmarshal64")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(param0)), uintptr(unsafe.Pointer(param1)), uintptr(unsafe.Pointer(param2)))
	return (*byte)(unsafe.Pointer(ret))
}

func HSTRING_UserFree64(param0 *uint32, param1 *HSTRING) {
	addr := lazyAddr(&pHSTRING_UserFree64, libApi_ms_win_core_winrt_string_l1_1_0, "HSTRING_UserFree64")
	syscall.SyscallN(addr, uintptr(unsafe.Pointer(param0)), uintptr(unsafe.Pointer(param1)))
}

func WindowsCreateString(sourceString PWSTR, length uint32, string *HSTRING) HRESULT {
	addr := lazyAddr(&pWindowsCreateString, libApi_ms_win_core_winrt_string_l1_1_0, "WindowsCreateString")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(sourceString)), uintptr(length), uintptr(unsafe.Pointer(string)))
	return HRESULT(ret)
}

func WindowsCreateStringReference(sourceString PWSTR, length uint32, hstringHeader *HSTRING_HEADER, string *HSTRING) HRESULT {
	addr := lazyAddr(&pWindowsCreateStringReference, libApi_ms_win_core_winrt_string_l1_1_0, "WindowsCreateStringReference")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(sourceString)), uintptr(length), uintptr(unsafe.Pointer(hstringHeader)), uintptr(unsafe.Pointer(string)))
	return HRESULT(ret)
}

func WindowsDeleteString(string HSTRING) HRESULT {
	addr := lazyAddr(&pWindowsDeleteString, libApi_ms_win_core_winrt_string_l1_1_0, "WindowsDeleteString")
	ret, _, _ := syscall.SyscallN(addr, string)
	return HRESULT(ret)
}

func WindowsDuplicateString(string HSTRING, newString *HSTRING) HRESULT {
	addr := lazyAddr(&pWindowsDuplicateString, libApi_ms_win_core_winrt_string_l1_1_0, "WindowsDuplicateString")
	ret, _, _ := syscall.SyscallN(addr, string, uintptr(unsafe.Pointer(newString)))
	return HRESULT(ret)
}

func WindowsGetStringLen(string HSTRING) uint32 {
	addr := lazyAddr(&pWindowsGetStringLen, libApi_ms_win_core_winrt_string_l1_1_0, "WindowsGetStringLen")
	ret, _, _ := syscall.SyscallN(addr, string)
	return uint32(ret)
}

func WindowsGetStringRawBuffer(string HSTRING, length *uint32) PWSTR {
	addr := lazyAddr(&pWindowsGetStringRawBuffer, libApi_ms_win_core_winrt_string_l1_1_0, "WindowsGetStringRawBuffer")
	ret, _, _ := syscall.SyscallN(addr, string, uintptr(unsafe.Pointer(length)))
	return (PWSTR)(unsafe.Pointer(ret))
}

func WindowsIsStringEmpty(string HSTRING) BOOL {
	addr := lazyAddr(&pWindowsIsStringEmpty, libApi_ms_win_core_winrt_string_l1_1_0, "WindowsIsStringEmpty")
	ret, _, _ := syscall.SyscallN(addr, string)
	return BOOL(ret)
}

func WindowsStringHasEmbeddedNull(string HSTRING, hasEmbedNull *BOOL) HRESULT {
	addr := lazyAddr(&pWindowsStringHasEmbeddedNull, libApi_ms_win_core_winrt_string_l1_1_0, "WindowsStringHasEmbeddedNull")
	ret, _, _ := syscall.SyscallN(addr, string, uintptr(unsafe.Pointer(hasEmbedNull)))
	return HRESULT(ret)
}

func WindowsCompareStringOrdinal(string1 HSTRING, string2 HSTRING, result *int32) HRESULT {
	addr := lazyAddr(&pWindowsCompareStringOrdinal, libApi_ms_win_core_winrt_string_l1_1_0, "WindowsCompareStringOrdinal")
	ret, _, _ := syscall.SyscallN(addr, string1, string2, uintptr(unsafe.Pointer(result)))
	return HRESULT(ret)
}

func WindowsSubstring(string HSTRING, startIndex uint32, newString *HSTRING) HRESULT {
	addr := lazyAddr(&pWindowsSubstring, libApi_ms_win_core_winrt_string_l1_1_0, "WindowsSubstring")
	ret, _, _ := syscall.SyscallN(addr, string, uintptr(startIndex), uintptr(unsafe.Pointer(newString)))
	return HRESULT(ret)
}

func WindowsSubstringWithSpecifiedLength(string HSTRING, startIndex uint32, length uint32, newString *HSTRING) HRESULT {
	addr := lazyAddr(&pWindowsSubstringWithSpecifiedLength, libApi_ms_win_core_winrt_string_l1_1_0, "WindowsSubstringWithSpecifiedLength")
	ret, _, _ := syscall.SyscallN(addr, string, uintptr(startIndex), uintptr(length), uintptr(unsafe.Pointer(newString)))
	return HRESULT(ret)
}

func WindowsConcatString(string1 HSTRING, string2 HSTRING, newString *HSTRING) HRESULT {
	addr := lazyAddr(&pWindowsConcatString, libApi_ms_win_core_winrt_string_l1_1_0, "WindowsConcatString")
	ret, _, _ := syscall.SyscallN(addr, string1, string2, uintptr(unsafe.Pointer(newString)))
	return HRESULT(ret)
}

func WindowsReplaceString(string HSTRING, stringReplaced HSTRING, stringReplaceWith HSTRING, newString *HSTRING) HRESULT {
	addr := lazyAddr(&pWindowsReplaceString, libApi_ms_win_core_winrt_string_l1_1_0, "WindowsReplaceString")
	ret, _, _ := syscall.SyscallN(addr, string, stringReplaced, stringReplaceWith, uintptr(unsafe.Pointer(newString)))
	return HRESULT(ret)
}

func WindowsTrimStringStart(string HSTRING, trimString HSTRING, newString *HSTRING) HRESULT {
	addr := lazyAddr(&pWindowsTrimStringStart, libApi_ms_win_core_winrt_string_l1_1_0, "WindowsTrimStringStart")
	ret, _, _ := syscall.SyscallN(addr, string, trimString, uintptr(unsafe.Pointer(newString)))
	return HRESULT(ret)
}

func WindowsTrimStringEnd(string HSTRING, trimString HSTRING, newString *HSTRING) HRESULT {
	addr := lazyAddr(&pWindowsTrimStringEnd, libApi_ms_win_core_winrt_string_l1_1_0, "WindowsTrimStringEnd")
	ret, _, _ := syscall.SyscallN(addr, string, trimString, uintptr(unsafe.Pointer(newString)))
	return HRESULT(ret)
}

func WindowsPreallocateStringBuffer(length uint32, charBuffer **uint16, bufferHandle *HSTRING_BUFFER) HRESULT {
	addr := lazyAddr(&pWindowsPreallocateStringBuffer, libApi_ms_win_core_winrt_string_l1_1_0, "WindowsPreallocateStringBuffer")
	ret, _, _ := syscall.SyscallN(addr, uintptr(length), uintptr(unsafe.Pointer(charBuffer)), uintptr(unsafe.Pointer(bufferHandle)))
	return HRESULT(ret)
}

func WindowsPromoteStringBuffer(bufferHandle HSTRING_BUFFER, string *HSTRING) HRESULT {
	addr := lazyAddr(&pWindowsPromoteStringBuffer, libApi_ms_win_core_winrt_string_l1_1_0, "WindowsPromoteStringBuffer")
	ret, _, _ := syscall.SyscallN(addr, bufferHandle, uintptr(unsafe.Pointer(string)))
	return HRESULT(ret)
}

func WindowsDeleteStringBuffer(bufferHandle HSTRING_BUFFER) HRESULT {
	addr := lazyAddr(&pWindowsDeleteStringBuffer, libApi_ms_win_core_winrt_string_l1_1_0, "WindowsDeleteStringBuffer")
	ret, _, _ := syscall.SyscallN(addr, bufferHandle)
	return HRESULT(ret)
}

func WindowsInspectString(targetHString uintptr, machine uint16, callback PINSPECT_HSTRING_CALLBACK, context unsafe.Pointer, length *uint32, targetStringAddress *uintptr) HRESULT {
	addr := lazyAddr(&pWindowsInspectString, libApi_ms_win_core_winrt_string_l1_1_0, "WindowsInspectString")
	ret, _, _ := syscall.SyscallN(addr, targetHString, uintptr(machine), callback, uintptr(context), uintptr(unsafe.Pointer(length)), uintptr(unsafe.Pointer(targetStringAddress)))
	return HRESULT(ret)
}

func RoInitialize(initType RO_INIT_TYPE) HRESULT {
	addr := lazyAddr(&pRoInitialize, libApi_ms_win_core_winrt_l1_1_0, "RoInitialize")
	ret, _, _ := syscall.SyscallN(addr, uintptr(initType))
	return HRESULT(ret)
}

func RoUninitialize() {
	addr := lazyAddr(&pRoUninitialize, libApi_ms_win_core_winrt_l1_1_0, "RoUninitialize")
	syscall.SyscallN(addr)
}

func RoActivateInstance(activatableClassId HSTRING, instance **IInspectable) HRESULT {
	addr := lazyAddr(&pRoActivateInstance, libApi_ms_win_core_winrt_l1_1_0, "RoActivateInstance")
	ret, _, _ := syscall.SyscallN(addr, activatableClassId, uintptr(unsafe.Pointer(instance)))
	return HRESULT(ret)
}

func RoRegisterActivationFactories(activatableClassIds *HSTRING, activationFactoryCallbacks *uintptr, count uint32, cookie *uintptr) HRESULT {
	addr := lazyAddr(&pRoRegisterActivationFactories, libApi_ms_win_core_winrt_l1_1_0, "RoRegisterActivationFactories")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(activatableClassIds)), uintptr(unsafe.Pointer(activationFactoryCallbacks)), uintptr(count), uintptr(unsafe.Pointer(cookie)))
	return HRESULT(ret)
}

func RoRevokeActivationFactories(cookie uintptr) {
	addr := lazyAddr(&pRoRevokeActivationFactories, libApi_ms_win_core_winrt_l1_1_0, "RoRevokeActivationFactories")
	syscall.SyscallN(addr, cookie)
}

func RoGetActivationFactory(activatableClassId HSTRING, iid *syscall.GUID, factory unsafe.Pointer) HRESULT {
	addr := lazyAddr(&pRoGetActivationFactory, libApi_ms_win_core_winrt_l1_1_0, "RoGetActivationFactory")
	ret, _, _ := syscall.SyscallN(addr, activatableClassId, uintptr(unsafe.Pointer(iid)), uintptr(factory))
	return HRESULT(ret)
}

func RoRegisterForApartmentShutdown(callbackObject *IApartmentShutdown, apartmentIdentifier *uint64, regCookie *APARTMENT_SHUTDOWN_REGISTRATION_COOKIE) HRESULT {
	addr := lazyAddr(&pRoRegisterForApartmentShutdown, libApi_ms_win_core_winrt_l1_1_0, "RoRegisterForApartmentShutdown")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(callbackObject)), uintptr(unsafe.Pointer(apartmentIdentifier)), uintptr(unsafe.Pointer(regCookie)))
	return HRESULT(ret)
}

func RoUnregisterForApartmentShutdown(regCookie APARTMENT_SHUTDOWN_REGISTRATION_COOKIE) HRESULT {
	addr := lazyAddr(&pRoUnregisterForApartmentShutdown, libApi_ms_win_core_winrt_l1_1_0, "RoUnregisterForApartmentShutdown")
	ret, _, _ := syscall.SyscallN(addr, regCookie)
	return HRESULT(ret)
}

func RoGetApartmentIdentifier(apartmentIdentifier *uint64) HRESULT {
	addr := lazyAddr(&pRoGetApartmentIdentifier, libApi_ms_win_core_winrt_l1_1_0, "RoGetApartmentIdentifier")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(apartmentIdentifier)))
	return HRESULT(ret)
}
