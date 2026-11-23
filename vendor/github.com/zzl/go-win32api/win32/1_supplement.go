package win32

import (
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

type HKL = uintptr
type HTASK = uintptr

type CONDITION_OPERATION = int32

type IID = syscall.GUID

type POINTER_TOUCH_INFO struct {
	Data [18]uint64
}

type POINTER_PEN_INFO struct {
	Data [15]uint64
}

type NETRESOURCEA struct {
}

//
type BYTE = byte
type CLSID = syscall.GUID
type LANGID = uint16
type UINT16 = uint16
type ULONG_PTR = uintptr
type COLORREF = uint32

type ATOM = uint16

type HGLOBAL = uintptr

type DWORD = uint32

type HREFTYPE = DWORD

type DISPID = int32
type MEMBERID = DISPID

type VARIANT_BOOL = int16

type TOOLINFO struct {
	CbSize     uint32
	UFlags     uint32
	Hwnd       HWND
	UId        uintptr
	Rect       RECT
	Hinst      HINSTANCE
	LpszText   *uint16
	LParam     uintptr
	LpReserved unsafe.Pointer
}

//
const (
	DATA_E_FORMATETC = DV_E_FORMATETC
)

//
const DTN_FIRST = ^uint32(739)  // -740
const DTN_FIRST2 = ^uint32(752) // -753

// Notifications
const (
	DTN_DATETIMECHANGE = DTN_FIRST2 - 6
	DTN_USERSTRING     = DTN_FIRST - 5
	DTN_WMKEYDOWN      = DTN_FIRST - 4
	DTN_FORMAT         = DTN_FIRST - 3
	DTN_FORMATQUERY    = DTN_FIRST - 2
	DTN_DROPDOWN       = DTN_FIRST2 - 1
	DTN_CLOSEUP        = DTN_FIRST2
)

// WM_NOTITY messages
const (
	NM_FIRST              = 0
	NM_OUTOFMEMORY        = ^uint32(0)  // NM_FIRST - 1
	NM_CLICK              = ^uint32(1)  // NM_FIRST - 2
	NM_DBLCLK             = ^uint32(2)  // NM_FIRST - 3
	NM_RETURN             = ^uint32(3)  // NM_FIRST - 4
	NM_RCLICK             = ^uint32(4)  // NM_FIRST - 5
	NM_RDBLCLK            = ^uint32(5)  // NM_FIRST - 6
	NM_SETFOCUS           = ^uint32(6)  // NM_FIRST - 7
	NM_KILLFOCUS          = ^uint32(7)  // NM_FIRST - 8
	NM_CUSTOMDRAW         = ^uint32(11) // NM_FIRST - 12
	NM_HOVER              = ^uint32(12) // NM_FIRST - 13
	NM_NCHITTEST          = ^uint32(13) // NM_FIRST - 14
	NM_KEYDOWN            = ^uint32(14) // NM_FIRST - 15
	NM_RELEASEDCAPTURE    = ^uint32(15) // NM_FIRST - 16
	NM_SETCURSOR          = ^uint32(16) // NM_FIRST - 17
	NM_CHAR               = ^uint32(17) // NM_FIRST - 18
	NM_TOOLTIPSCREATED    = ^uint32(18) // NM_FIRST - 19
	NM_LAST               = ^uint32(98) // NM_FIRST - 99
	TRBN_THUMBPOSCHANGING = 0xfffffa22  // TRBN_FIRST - 1
)

// ListView notifications
const (
	LVN_FIRST = ^uint32(99) // -100

	LVN_ITEMCHANGING      = LVN_FIRST - 0
	LVN_ITEMCHANGED       = LVN_FIRST - 1
	LVN_INSERTITEM        = LVN_FIRST - 2
	LVN_DELETEITEM        = LVN_FIRST - 3
	LVN_DELETEALLITEMS    = LVN_FIRST - 4
	LVN_BEGINLABELEDIT    = LVN_FIRST - 75
	LVN_ENDLABELEDIT      = LVN_FIRST - 76
	LVN_COLUMNCLICK       = LVN_FIRST - 8
	LVN_BEGINDRAG         = LVN_FIRST - 9
	LVN_BEGINRDRAG        = LVN_FIRST - 11
	LVN_ODCACHEHINT       = LVN_FIRST - 13
	LVN_ODFINDITEM        = LVN_FIRST - 79
	LVN_ITEMACTIVATE      = LVN_FIRST - 14
	LVN_ODSTATECHANGED    = LVN_FIRST - 15
	LVN_HOTTRACK          = LVN_FIRST - 21
	LVN_GETDISPINFO       = LVN_FIRST - 77
	LVN_SETDISPINFO       = LVN_FIRST - 78
	LVN_KEYDOWN           = LVN_FIRST - 55
	LVN_MARQUEEBEGIN      = LVN_FIRST - 56
	LVN_GETINFOTIP        = LVN_FIRST - 58
	LVN_INCREMENTALSEARCH = LVN_FIRST - 63
	LVN_BEGINSCROLL       = LVN_FIRST - 80
	LVN_ENDSCROLL         = LVN_FIRST - 81
)

const (
	RBN_FIRST         = 0xFFFFFCC1
	RBN_HEIGHTCHANGE  = RBN_FIRST - 0
	RBN_GETOBJECT     = RBN_FIRST - 1
	RBN_LAYOUTCHANGED = RBN_FIRST - 2
	RBN_AUTOSIZE      = RBN_FIRST - 3
	RBN_BEGINDRAG     = RBN_FIRST - 4
	RBN_ENDDRAG       = RBN_FIRST - 5
	RBN_DELETINGBAND  = RBN_FIRST - 6
	RBN_DELETEDBAND   = RBN_FIRST - 7
	RBN_CHILDSIZE     = RBN_FIRST - 8
	RBN_CHEVRONPUSHED = RBN_FIRST - 10
	RBN_SPLITTERDRAG  = RBN_FIRST - 11
	RBN_MINMAX        = RBN_FIRST - 21
	RBN_AUTOBREAK     = RBN_FIRST - 22
)

const (
	WVR_REDRAW = WVR_HREDRAW | WVR_VREDRAW
)

const (
	EP_EDITTEXT             = 1
	EP_CARET                = 2
	EP_BACKGROUND           = 3
	EP_PASSWORD             = 4
	EP_BACKGROUNDWITHBORDER = 5
	EP_EDITBORDER_NOSCROLL  = 6
	EP_EDITBORDER_HSCROLL   = 7
	EP_EDITBORDER_VSCROLL   = 8
	EP_EDITBORDER_HVSCROLL  = 9
)

const (
	ETS_NORMAL    = 1
	ETS_HOT       = 2
	ETS_SELECTED  = 3
	ETS_DISABLED  = 4
	ETS_FOCUSED   = 5
	ETS_READONLY  = 6
	ETS_ASSIST    = 7
	ETS_CUEBANNER = 8
)

// TrackBar (Slider) messages
const (
	TBM_GETPOS = WM_USER
)

const TCN_FIRST uint32 = 0xFFFFFDDA

const (
	TCN_KEYDOWN     = TCN_FIRST - 0
	TCN_SELCHANGE   = TCN_FIRST - 1
	TCN_SELCHANGING = TCN_FIRST - 2
	TCN_GETOBJECT   = TCN_FIRST - 3
	TCN_FOCUSCHANGE = TCN_FIRST - 4
)

const MENU_BARITEM = 8
const MBI_NORMAL = 1
const MBI_HOT = 2
const MBI_PUSHED = 3

const TRUE BOOL = 1
const FALSE BOOL = 0

const (
	CLR_NONE_U    = ^uint32(-CLR_NONE - 1)
	CLR_DEFAULT_U = ^uint32(-CLR_DEFAULT - 1)
)

// ToolBar notifications
const (
	TBN_FIRST       = 0xFFFFFD44 //-700
	TBN_DROPDOWN    = TBN_FIRST - 10
	TBN_GETINFOTIP  = TBN_FIRST - 19
	TBN_BEGINADJUST = TBN_FIRST - 3
	TBN_QUERYINSERT = TBN_FIRST - 6
)

const (
	TTN_FIRST       = ^uint32(519)
	TTN_GETDISPINFO = TTN_FIRST - 10
	TTN_SHOW        = TTN_FIRST - 1
	TTN_POP         = TTN_FIRST - 2
	TTN_LINKCLICK   = TTN_FIRST - 3
)

// TreeView notifications
const (
	TVN_FIRST = ^uint32(399)

	TVN_SELCHANGING    = TVN_FIRST - 50
	TVN_SELCHANGED     = TVN_FIRST - 51
	TVN_GETDISPINFO    = TVN_FIRST - 52
	TVN_ITEMEXPANDING  = TVN_FIRST - 54
	TVN_ITEMEXPANDED   = TVN_FIRST - 55
	TVN_BEGINDRAG      = TVN_FIRST - 56
	TVN_BEGINRDRAG     = TVN_FIRST - 57
	TVN_DELETEITEM     = TVN_FIRST - 58
	TVN_BEGINLABELEDIT = TVN_FIRST - 59
	TVN_ENDLABELEDIT   = TVN_FIRST - 60
	TVN_KEYDOWN        = TVN_FIRST - 12
	TVN_GETINFOTIP     = TVN_FIRST - 14
	TVN_SINGLEEXPAND   = TVN_FIRST - 15
	TVN_ITEMCHANGING   = TVN_FIRST - 17
	TVN_ITEMCHANGED    = TVN_FIRST - 19
	TVN_ASYNCDRAW      = TVN_FIRST - 20

	NM_TVSTATEIMAGECHANGING = ^uint32(23) //NM_FIRST-24
)

// SHGetStockIconInfo flags
const (
	//SHGSI_ICONLOCATION  = 0
	SHGSI_ICON          = 0x000000100
	SHGSI_SYSICONINDEX  = 0x000004000
	SHGSI_LINKOVERLAY   = 0x000008000
	SHGSI_SELECTED      = 0x000010000
	SHGSI_LARGEICON     = 0x000000000
	SHGSI_SMALLICON     = 0x000000001
	SHGSI_SHELLICONSIZE = 0x000000004
)

// Button parts
const (
	BP_PUSHBUTTON       = 1
	BP_RADIOBUTTON      = 2
	BP_CHECKBOX         = 3
	BP_GROUPBOX         = 4
	BP_USERBUTTON       = 5
	BP_COMMANDLINK      = 6
	BP_COMMANDLINKGLYPH = 7
)

// CheckBox states
const (
	CBS_UNCHECKEDNORMAL   = 1
	CBS_UNCHECKEDHOT      = 2
	CBS_UNCHECKEDPRESSED  = 3
	CBS_UNCHECKEDDISABLED = 4
	CBS_CHECKEDNORMAL     = 5
	CBS_CHECKEDHOT        = 6
	CBS_CHECKEDPRESSED    = 7
	CBS_CHECKEDDISABLED   = 8
	CBS_MIXEDNORMAL       = 9
	CBS_MIXEDHOT          = 10
	CBS_MIXEDPRESSED      = 11
	CBS_MIXEDDISABLED     = 12
	CBS_IMPLICITNORMAL    = 13
	CBS_IMPLICITHOT       = 14
	CBS_IMPLICITPRESSED   = 15
	CBS_IMPLICITDISABLED  = 16
	CBS_EXCLUDEDNORMAL    = 17
	CBS_EXCLUDEDHOT       = 18
	CBS_EXCLUDEDPRESSED   = 19
	CBS_EXCLUDEDDISABLED  = 20
)

const (
	SP_GRIPPER            = 3
	LOCALE_INVARIANT      = 127
	LOCALE_USER_DEFAULT   = 1024
	LOCALE_SYSTEM_DEFAULT = 2048
)

const (
	VARIANT_TRUE  VARIANT_BOOL = -1
	VARIANT_FALSE VARIANT_BOOL = 0
)

//
const VK_WHEEL_DOWN VIRTUAL_KEY = 0x9E
const VK_WHEEL_UP VIRTUAL_KEY = 0x9F
const VK_WHEEL_LEFT VIRTUAL_KEY = 0x9C
const VK_WHEEL_RIGHT VIRTUAL_KEY = 0x9D

//
var (
	IID_NULL syscall.GUID
)

//
func RGB(r, g, b byte) COLORREF {
	return COLORREF(r) | (COLORREF(g) << 8) | (COLORREF(b) << 16)
}

func MAKEINTRESOURCE(id uint16) *uint16 {
	return (*uint16)(unsafe.Pointer(uintptr(id)))
}

func MAKELONG(lo, hi uint16) uint32 {
	return uint32(uint32(lo) | ((uint32(hi)) << 16))
}

func GET_X_LPARAM(dw DWORD) int32 {
	return int32(int16(LOWORD(dw)))
}

func GET_Y_LPARAM(dw DWORD) int32 {
	return int32(int16(HIWORD(dw)))
}

func LOBYTE(w uint16) byte {
	return byte(w)
}

func HIBYTE(w uint16) byte {
	return byte(w >> 8 & 0xff)
}

func MAKEWORD(lo, hi byte) uint16 {
	return uint16(uint16(lo) | ((uint16(hi)) << 8))
}

func LOWORD(dw uint32) uint16 {
	return uint16(dw)
}

func HIWORD(dw uint32) uint16 {
	return uint16(dw >> 16 & 0xffff)
}

func SUCCEEDED(hr HRESULT) bool {
	return hr >= 0
}

func FAILED(hr HRESULT) bool {
	return hr < 0
}

//
var (
	copyMemory        *windows.LazyProc
	imageList_AddIcon *windows.LazyProc
)

//
func init() {
	copyMemory = libKernel32.NewProc("RtlMoveMemory")
	imageList_AddIcon = libComctl32.NewProc("ImageList_AddIcon")

	//
	lazyAddr(&pTlsGetValue, libKernel32, "TlsGetValue")
}

func CopyMemory(Destination unsafe.Pointer, Source unsafe.Pointer, Length uint32) WIN32_ERROR {
	_, _, err := syscall.SyscallN(copyMemory.Addr(),
		uintptr(Destination),
		uintptr(Source),
		uintptr(Length))
	return WIN32_ERROR(err)
}

func ImageList_AddIcon(himl HIMAGELIST, hicon HICON) (int32, WIN32_ERROR) {
	ret, _, err := syscall.SyscallN(imageList_AddIcon.Addr(), himl, hicon)
	return int32(ret), WIN32_ERROR(err)
}

//
type IUnknownObject interface {
	GetIUnknown() *IUnknown
}

func (this *IUnknown) GetIUnknown() *IUnknown {
	return this
}

type IDispatchObject interface {
	GetIDispatch_() *IDispatch
}

func (this *IDispatch) GetIDispatch_() *IDispatch {
	return this
}

//
func TlsGetValueAlt(dwTlsIndex uint32) uintptr {
	ret, _, _ := syscall.SyscallN(pTlsGetValue, uintptr(dwTlsIndex))
	return ret
}
