package win32

import (
	"syscall"
	"unsafe"
)

type (
	HRAWINPUT = uintptr
)

// enums

// enum
type RAW_INPUT_DATA_COMMAND_FLAGS uint32

const (
	RID_HEADER RAW_INPUT_DATA_COMMAND_FLAGS = 268435461
	RID_INPUT  RAW_INPUT_DATA_COMMAND_FLAGS = 268435459
)

// enum
type RAW_INPUT_DEVICE_INFO_COMMAND uint32

const (
	RIDI_PREPARSEDDATA RAW_INPUT_DEVICE_INFO_COMMAND = 536870917
	RIDI_DEVICENAME    RAW_INPUT_DEVICE_INFO_COMMAND = 536870919
	RIDI_DEVICEINFO    RAW_INPUT_DEVICE_INFO_COMMAND = 536870923
)

// enum
type RID_DEVICE_INFO_TYPE uint32

const (
	RIM_TYPEMOUSE    RID_DEVICE_INFO_TYPE = 0
	RIM_TYPEKEYBOARD RID_DEVICE_INFO_TYPE = 1
	RIM_TYPEHID      RID_DEVICE_INFO_TYPE = 2
)

// enum
// flags
type RAWINPUTDEVICE_FLAGS uint32

const (
	RIDEV_REMOVE       RAWINPUTDEVICE_FLAGS = 1
	RIDEV_EXCLUDE      RAWINPUTDEVICE_FLAGS = 16
	RIDEV_PAGEONLY     RAWINPUTDEVICE_FLAGS = 32
	RIDEV_NOLEGACY     RAWINPUTDEVICE_FLAGS = 48
	RIDEV_INPUTSINK    RAWINPUTDEVICE_FLAGS = 256
	RIDEV_CAPTUREMOUSE RAWINPUTDEVICE_FLAGS = 512
	RIDEV_NOHOTKEYS    RAWINPUTDEVICE_FLAGS = 512
	RIDEV_APPKEYS      RAWINPUTDEVICE_FLAGS = 1024
	RIDEV_EXINPUTSINK  RAWINPUTDEVICE_FLAGS = 4096
	RIDEV_DEVNOTIFY    RAWINPUTDEVICE_FLAGS = 8192
)

// enum
type INPUT_MESSAGE_DEVICE_TYPE int32

const (
	IMDT_UNAVAILABLE INPUT_MESSAGE_DEVICE_TYPE = 0
	IMDT_KEYBOARD    INPUT_MESSAGE_DEVICE_TYPE = 1
	IMDT_MOUSE       INPUT_MESSAGE_DEVICE_TYPE = 2
	IMDT_TOUCH       INPUT_MESSAGE_DEVICE_TYPE = 4
	IMDT_PEN         INPUT_MESSAGE_DEVICE_TYPE = 8
	IMDT_TOUCHPAD    INPUT_MESSAGE_DEVICE_TYPE = 16
)

// enum
type INPUT_MESSAGE_ORIGIN_ID int32

const (
	IMO_UNAVAILABLE INPUT_MESSAGE_ORIGIN_ID = 0
	IMO_HARDWARE    INPUT_MESSAGE_ORIGIN_ID = 1
	IMO_INJECTED    INPUT_MESSAGE_ORIGIN_ID = 2
	IMO_SYSTEM      INPUT_MESSAGE_ORIGIN_ID = 4
)

// structs

type RAWINPUTHEADER struct {
	DwType  uint32
	DwSize  uint32
	HDevice HANDLE
	WParam  WPARAM
}

type RAWMOUSE_Anonymous_Anonymous struct {
	UsButtonFlags uint16
	UsButtonData  uint16
}

type RAWMOUSE_Anonymous struct {
	RAWMOUSE_Anonymous_Anonymous
}

func (this *RAWMOUSE_Anonymous) UlButtons() *uint32 {
	return (*uint32)(unsafe.Pointer(this))
}

func (this *RAWMOUSE_Anonymous) UlButtonsVal() uint32 {
	return *(*uint32)(unsafe.Pointer(this))
}

func (this *RAWMOUSE_Anonymous) Anonymous() *RAWMOUSE_Anonymous_Anonymous {
	return (*RAWMOUSE_Anonymous_Anonymous)(unsafe.Pointer(this))
}

func (this *RAWMOUSE_Anonymous) AnonymousVal() RAWMOUSE_Anonymous_Anonymous {
	return *(*RAWMOUSE_Anonymous_Anonymous)(unsafe.Pointer(this))
}

type RAWMOUSE struct {
	UsFlags uint16
	RAWMOUSE_Anonymous
	UlRawButtons       uint32
	LLastX             int32
	LLastY             int32
	UlExtraInformation uint32
}

type RAWKEYBOARD struct {
	MakeCode         uint16
	Flags            uint16
	Reserved         uint16
	VKey             uint16
	Message          uint32
	ExtraInformation uint32
}

type RAWHID struct {
	DwSizeHid uint32
	DwCount   uint32
	BRawData  [1]byte
}

type RAWINPUT_Data struct {
	Data [6]uint32
}

func (this *RAWINPUT_Data) Mouse() *RAWMOUSE {
	return (*RAWMOUSE)(unsafe.Pointer(this))
}

func (this *RAWINPUT_Data) MouseVal() RAWMOUSE {
	return *(*RAWMOUSE)(unsafe.Pointer(this))
}

func (this *RAWINPUT_Data) Keyboard() *RAWKEYBOARD {
	return (*RAWKEYBOARD)(unsafe.Pointer(this))
}

func (this *RAWINPUT_Data) KeyboardVal() RAWKEYBOARD {
	return *(*RAWKEYBOARD)(unsafe.Pointer(this))
}

func (this *RAWINPUT_Data) Hid() *RAWHID {
	return (*RAWHID)(unsafe.Pointer(this))
}

func (this *RAWINPUT_Data) HidVal() RAWHID {
	return *(*RAWHID)(unsafe.Pointer(this))
}

type RAWINPUT struct {
	Header RAWINPUTHEADER
	Data   RAWINPUT_Data
}

type RID_DEVICE_INFO_MOUSE struct {
	DwId                uint32
	DwNumberOfButtons   uint32
	DwSampleRate        uint32
	FHasHorizontalWheel BOOL
}

type RID_DEVICE_INFO_KEYBOARD struct {
	DwType                 uint32
	DwSubType              uint32
	DwKeyboardMode         uint32
	DwNumberOfFunctionKeys uint32
	DwNumberOfIndicators   uint32
	DwNumberOfKeysTotal    uint32
}

type RID_DEVICE_INFO_HID struct {
	DwVendorId      uint32
	DwProductId     uint32
	DwVersionNumber uint32
	UsUsagePage     uint16
	UsUsage         uint16
}

type RID_DEVICE_INFO_Anonymous struct {
	Data [6]uint32
}

func (this *RID_DEVICE_INFO_Anonymous) Mouse() *RID_DEVICE_INFO_MOUSE {
	return (*RID_DEVICE_INFO_MOUSE)(unsafe.Pointer(this))
}

func (this *RID_DEVICE_INFO_Anonymous) MouseVal() RID_DEVICE_INFO_MOUSE {
	return *(*RID_DEVICE_INFO_MOUSE)(unsafe.Pointer(this))
}

func (this *RID_DEVICE_INFO_Anonymous) Keyboard() *RID_DEVICE_INFO_KEYBOARD {
	return (*RID_DEVICE_INFO_KEYBOARD)(unsafe.Pointer(this))
}

func (this *RID_DEVICE_INFO_Anonymous) KeyboardVal() RID_DEVICE_INFO_KEYBOARD {
	return *(*RID_DEVICE_INFO_KEYBOARD)(unsafe.Pointer(this))
}

func (this *RID_DEVICE_INFO_Anonymous) Hid() *RID_DEVICE_INFO_HID {
	return (*RID_DEVICE_INFO_HID)(unsafe.Pointer(this))
}

func (this *RID_DEVICE_INFO_Anonymous) HidVal() RID_DEVICE_INFO_HID {
	return *(*RID_DEVICE_INFO_HID)(unsafe.Pointer(this))
}

type RID_DEVICE_INFO struct {
	CbSize uint32
	DwType RID_DEVICE_INFO_TYPE
	RID_DEVICE_INFO_Anonymous
}

type RAWINPUTDEVICE struct {
	UsUsagePage uint16
	UsUsage     uint16
	DwFlags     RAWINPUTDEVICE_FLAGS
	HwndTarget  HWND
}

type RAWINPUTDEVICELIST struct {
	HDevice HANDLE
	DwType  RID_DEVICE_INFO_TYPE
}

type INPUT_MESSAGE_SOURCE struct {
	DeviceType INPUT_MESSAGE_DEVICE_TYPE
	OriginId   INPUT_MESSAGE_ORIGIN_ID
}

var (
	pGetRawInputData              uintptr
	pGetRawInputDeviceInfoA       uintptr
	pGetRawInputDeviceInfoW       uintptr
	pGetRawInputBuffer            uintptr
	pRegisterRawInputDevices      uintptr
	pGetRegisteredRawInputDevices uintptr
	pGetRawInputDeviceList        uintptr
	pDefRawInputProc              uintptr
	pGetCurrentInputMessageSource uintptr
	pGetCIMSSM                    uintptr
)

func GetRawInputData(hRawInput HRAWINPUT, uiCommand RAW_INPUT_DATA_COMMAND_FLAGS, pData unsafe.Pointer, pcbSize *uint32, cbSizeHeader uint32) uint32 {
	addr := lazyAddr(&pGetRawInputData, libUser32, "GetRawInputData")
	ret, _, _ := syscall.SyscallN(addr, hRawInput, uintptr(uiCommand), uintptr(pData), uintptr(unsafe.Pointer(pcbSize)), uintptr(cbSizeHeader))
	return uint32(ret)
}

func GetRawInputDeviceInfoA(hDevice HANDLE, uiCommand RAW_INPUT_DEVICE_INFO_COMMAND, pData unsafe.Pointer, pcbSize *uint32) (uint32, WIN32_ERROR) {
	addr := lazyAddr(&pGetRawInputDeviceInfoA, libUser32, "GetRawInputDeviceInfoA")
	ret, _, err := syscall.SyscallN(addr, hDevice, uintptr(uiCommand), uintptr(pData), uintptr(unsafe.Pointer(pcbSize)))
	return uint32(ret), WIN32_ERROR(err)
}

var GetRawInputDeviceInfo = GetRawInputDeviceInfoW

func GetRawInputDeviceInfoW(hDevice HANDLE, uiCommand RAW_INPUT_DEVICE_INFO_COMMAND, pData unsafe.Pointer, pcbSize *uint32) (uint32, WIN32_ERROR) {
	addr := lazyAddr(&pGetRawInputDeviceInfoW, libUser32, "GetRawInputDeviceInfoW")
	ret, _, err := syscall.SyscallN(addr, hDevice, uintptr(uiCommand), uintptr(pData), uintptr(unsafe.Pointer(pcbSize)))
	return uint32(ret), WIN32_ERROR(err)
}

func GetRawInputBuffer(pData *RAWINPUT, pcbSize *uint32, cbSizeHeader uint32) (uint32, WIN32_ERROR) {
	addr := lazyAddr(&pGetRawInputBuffer, libUser32, "GetRawInputBuffer")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pData)), uintptr(unsafe.Pointer(pcbSize)), uintptr(cbSizeHeader))
	return uint32(ret), WIN32_ERROR(err)
}

func RegisterRawInputDevices(pRawInputDevices *RAWINPUTDEVICE, uiNumDevices uint32, cbSize uint32) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pRegisterRawInputDevices, libUser32, "RegisterRawInputDevices")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pRawInputDevices)), uintptr(uiNumDevices), uintptr(cbSize))
	return BOOL(ret), WIN32_ERROR(err)
}

func GetRegisteredRawInputDevices(pRawInputDevices *RAWINPUTDEVICE, puiNumDevices *uint32, cbSize uint32) (uint32, WIN32_ERROR) {
	addr := lazyAddr(&pGetRegisteredRawInputDevices, libUser32, "GetRegisteredRawInputDevices")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pRawInputDevices)), uintptr(unsafe.Pointer(puiNumDevices)), uintptr(cbSize))
	return uint32(ret), WIN32_ERROR(err)
}

func GetRawInputDeviceList(pRawInputDeviceList *RAWINPUTDEVICELIST, puiNumDevices *uint32, cbSize uint32) (uint32, WIN32_ERROR) {
	addr := lazyAddr(&pGetRawInputDeviceList, libUser32, "GetRawInputDeviceList")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pRawInputDeviceList)), uintptr(unsafe.Pointer(puiNumDevices)), uintptr(cbSize))
	return uint32(ret), WIN32_ERROR(err)
}

func DefRawInputProc(paRawInput **RAWINPUT, nInput int32, cbSizeHeader uint32) LRESULT {
	addr := lazyAddr(&pDefRawInputProc, libUser32, "DefRawInputProc")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(paRawInput)), uintptr(nInput), uintptr(cbSizeHeader))
	return ret
}

func GetCurrentInputMessageSource(inputMessageSource *INPUT_MESSAGE_SOURCE) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pGetCurrentInputMessageSource, libUser32, "GetCurrentInputMessageSource")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(inputMessageSource)))
	return BOOL(ret), WIN32_ERROR(err)
}

func GetCIMSSM(inputMessageSource *INPUT_MESSAGE_SOURCE) BOOL {
	addr := lazyAddr(&pGetCIMSSM, libUser32, "GetCIMSSM")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(inputMessageSource)))
	return BOOL(ret)
}
