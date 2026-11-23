package win32

import (
	"syscall"
	"unsafe"
)

type (
	HHOOK   = uintptr
	HICON   = uintptr
	HMENU   = uintptr
	HCURSOR = uintptr
	HACCEL  = uintptr
)

const (
	WM_DEVICECHANGE                                            uint32  = 0x219
	BSM_VXDS                                                   uint32  = 0x1
	BSM_NETDRIVER                                              uint32  = 0x2
	BSM_INSTALLABLEDRIVERS                                     uint32  = 0x4
	WM_CONTEXTMENU                                             uint32  = 0x7b
	WM_UNICHAR                                                 uint32  = 0x109
	WM_PRINTCLIENT                                             uint32  = 0x318
	WM_NOTIFY                                                  uint32  = 0x4e
	DIFFERENCE                                                 uint32  = 0xb
	RT_MANIFEST                                                uint32  = 0x18
	CREATEPROCESS_MANIFEST_RESOURCE_ID                         uint32  = 0x1
	ISOLATIONAWARE_MANIFEST_RESOURCE_ID                        uint32  = 0x2
	ISOLATIONAWARE_NOSTATICIMPORT_MANIFEST_RESOURCE_ID         uint32  = 0x3
	ISOLATIONPOLICY_MANIFEST_RESOURCE_ID                       uint32  = 0x4
	ISOLATIONPOLICY_BROWSER_MANIFEST_RESOURCE_ID               uint32  = 0x5
	MINIMUM_RESERVED_MANIFEST_RESOURCE_ID                      uint32  = 0x1
	MAXIMUM_RESERVED_MANIFEST_RESOURCE_ID                      uint32  = 0x10
	SB_LINEUP                                                  uint32  = 0x0
	SB_LINELEFT                                                uint32  = 0x0
	SB_LINEDOWN                                                uint32  = 0x1
	SB_LINERIGHT                                               uint32  = 0x1
	SB_PAGEUP                                                  uint32  = 0x2
	SB_PAGELEFT                                                uint32  = 0x2
	SB_PAGEDOWN                                                uint32  = 0x3
	SB_PAGERIGHT                                               uint32  = 0x3
	SB_THUMBPOSITION                                           uint32  = 0x4
	SB_THUMBTRACK                                              uint32  = 0x5
	SB_TOP                                                     uint32  = 0x6
	SB_LEFT                                                    uint32  = 0x6
	SB_BOTTOM                                                  uint32  = 0x7
	SB_RIGHT                                                   uint32  = 0x7
	SB_ENDSCROLL                                               uint32  = 0x8
	HIDE_WINDOW                                                uint32  = 0x0
	SHOW_OPENWINDOW                                            uint32  = 0x1
	SHOW_ICONWINDOW                                            uint32  = 0x2
	SHOW_FULLSCREEN                                            uint32  = 0x3
	SHOW_OPENNOACTIVATE                                        uint32  = 0x4
	KF_EXTENDED                                                uint32  = 0x100
	KF_DLGMODE                                                 uint32  = 0x800
	KF_MENUMODE                                                uint32  = 0x1000
	KF_ALTDOWN                                                 uint32  = 0x2000
	KF_REPEAT                                                  uint32  = 0x4000
	KF_UP                                                      uint32  = 0x8000
	WH_MIN                                                     int32   = -1
	WH_HARDWARE                                                uint32  = 0x8
	WH_MAX                                                     uint32  = 0xe
	WH_MINHOOK                                                 int32   = -1
	WH_MAXHOOK                                                 uint32  = 0xe
	HC_ACTION                                                  uint32  = 0x0
	HC_GETNEXT                                                 uint32  = 0x1
	HC_SKIP                                                    uint32  = 0x2
	HC_NOREMOVE                                                uint32  = 0x3
	HC_NOREM                                                   uint32  = 0x3
	HC_SYSMODALON                                              uint32  = 0x4
	HC_SYSMODALOFF                                             uint32  = 0x5
	HCBT_MOVESIZE                                              uint32  = 0x0
	HCBT_MINMAX                                                uint32  = 0x1
	HCBT_QS                                                    uint32  = 0x2
	HCBT_CREATEWND                                             uint32  = 0x3
	HCBT_DESTROYWND                                            uint32  = 0x4
	HCBT_ACTIVATE                                              uint32  = 0x5
	HCBT_CLICKSKIPPED                                          uint32  = 0x6
	HCBT_KEYSKIPPED                                            uint32  = 0x7
	HCBT_SYSCOMMAND                                            uint32  = 0x8
	HCBT_SETFOCUS                                              uint32  = 0x9
	WTS_CONSOLE_CONNECT                                        uint32  = 0x1
	WTS_CONSOLE_DISCONNECT                                     uint32  = 0x2
	WTS_REMOTE_CONNECT                                         uint32  = 0x3
	WTS_REMOTE_DISCONNECT                                      uint32  = 0x4
	WTS_SESSION_LOGON                                          uint32  = 0x5
	WTS_SESSION_LOGOFF                                         uint32  = 0x6
	WTS_SESSION_LOCK                                           uint32  = 0x7
	WTS_SESSION_UNLOCK                                         uint32  = 0x8
	WTS_SESSION_REMOTE_CONTROL                                 uint32  = 0x9
	WTS_SESSION_CREATE                                         uint32  = 0xa
	WTS_SESSION_TERMINATE                                      uint32  = 0xb
	MSGF_DIALOGBOX                                             uint32  = 0x0
	MSGF_MESSAGEBOX                                            uint32  = 0x1
	MSGF_MENU                                                  uint32  = 0x2
	MSGF_SCROLLBAR                                             uint32  = 0x5
	MSGF_NEXTWINDOW                                            uint32  = 0x6
	MSGF_MAX                                                   uint32  = 0x8
	MSGF_USER                                                  uint32  = 0x1000
	HSHELL_WINDOWCREATED                                       uint32  = 0x1
	HSHELL_WINDOWDESTROYED                                     uint32  = 0x2
	HSHELL_ACTIVATESHELLWINDOW                                 uint32  = 0x3
	HSHELL_WINDOWACTIVATED                                     uint32  = 0x4
	HSHELL_GETMINRECT                                          uint32  = 0x5
	HSHELL_REDRAW                                              uint32  = 0x6
	HSHELL_TASKMAN                                             uint32  = 0x7
	HSHELL_LANGUAGE                                            uint32  = 0x8
	HSHELL_SYSMENU                                             uint32  = 0x9
	HSHELL_ENDTASK                                             uint32  = 0xa
	HSHELL_ACCESSIBILITYSTATE                                  uint32  = 0xb
	HSHELL_APPCOMMAND                                          uint32  = 0xc
	HSHELL_WINDOWREPLACED                                      uint32  = 0xd
	HSHELL_WINDOWREPLACING                                     uint32  = 0xe
	HSHELL_MONITORCHANGED                                      uint32  = 0x10
	HSHELL_HIGHBIT                                             uint32  = 0x8000
	FAPPCOMMAND_MOUSE                                          uint32  = 0x8000
	FAPPCOMMAND_KEY                                            uint32  = 0x0
	FAPPCOMMAND_OEM                                            uint32  = 0x1000
	FAPPCOMMAND_MASK                                           uint32  = 0xf000
	LLMHF_INJECTED                                             uint32  = 0x1
	LLMHF_LOWER_IL_INJECTED                                    uint32  = 0x2
	HKL_PREV                                                   uint32  = 0x0
	HKL_NEXT                                                   uint32  = 0x1
	INPUTLANGCHANGE_SYSCHARSET                                 uint32  = 0x1
	INPUTLANGCHANGE_FORWARD                                    uint32  = 0x2
	INPUTLANGCHANGE_BACKWARD                                   uint32  = 0x4
	KL_NAMELENGTH                                              uint32  = 0x9
	DESKTOP_READOBJECTS                                        int32   = 1
	DESKTOP_CREATEWINDOW                                       int32   = 2
	DESKTOP_CREATEMENU                                         int32   = 4
	DESKTOP_HOOKCONTROL                                        int32   = 8
	DESKTOP_JOURNALRECORD                                      int32   = 16
	DESKTOP_JOURNALPLAYBACK                                    int32   = 32
	DESKTOP_ENUMERATE                                          int32   = 64
	DESKTOP_WRITEOBJECTS                                       int32   = 128
	DESKTOP_SWITCHDESKTOP                                      int32   = 256
	DF_ALLOWOTHERACCOUNTHOOK                                   int32   = 1
	WINSTA_ENUMDESKTOPS                                        int32   = 1
	WINSTA_READATTRIBUTES                                      int32   = 2
	WINSTA_ACCESSCLIPBOARD                                     int32   = 4
	WINSTA_CREATEDESKTOP                                       int32   = 8
	WINSTA_WRITEATTRIBUTES                                     int32   = 16
	WINSTA_ACCESSGLOBALATOMS                                   int32   = 32
	WINSTA_EXITWINDOWS                                         int32   = 64
	WINSTA_ENUMERATE                                           int32   = 256
	WINSTA_READSCREEN                                          int32   = 512
	CWF_CREATE_ONLY                                            uint32  = 0x1
	WSF_VISIBLE                                                int32   = 1
	UOI_TIMERPROC_EXCEPTION_SUPPRESSION                        uint32  = 0x7
	WM_NULL                                                    uint32  = 0x0
	WM_CREATE                                                  uint32  = 0x1
	WM_DESTROY                                                 uint32  = 0x2
	WM_MOVE                                                    uint32  = 0x3
	WM_SIZE                                                    uint32  = 0x5
	WM_ACTIVATE                                                uint32  = 0x6
	WA_INACTIVE                                                uint32  = 0x0
	WA_ACTIVE                                                  uint32  = 0x1
	WA_CLICKACTIVE                                             uint32  = 0x2
	WM_SETFOCUS                                                uint32  = 0x7
	WM_KILLFOCUS                                               uint32  = 0x8
	WM_ENABLE                                                  uint32  = 0xa
	WM_SETREDRAW                                               uint32  = 0xb
	WM_SETTEXT                                                 uint32  = 0xc
	WM_GETTEXT                                                 uint32  = 0xd
	WM_GETTEXTLENGTH                                           uint32  = 0xe
	WM_PAINT                                                   uint32  = 0xf
	WM_CLOSE                                                   uint32  = 0x10
	WM_QUERYENDSESSION                                         uint32  = 0x11
	WM_QUERYOPEN                                               uint32  = 0x13
	WM_ENDSESSION                                              uint32  = 0x16
	WM_QUIT                                                    uint32  = 0x12
	WM_ERASEBKGND                                              uint32  = 0x14
	WM_SYSCOLORCHANGE                                          uint32  = 0x15
	WM_SHOWWINDOW                                              uint32  = 0x18
	WM_WININICHANGE                                            uint32  = 0x1a
	WM_SETTINGCHANGE                                           uint32  = 0x1a
	WM_DEVMODECHANGE                                           uint32  = 0x1b
	WM_ACTIVATEAPP                                             uint32  = 0x1c
	WM_FONTCHANGE                                              uint32  = 0x1d
	WM_TIMECHANGE                                              uint32  = 0x1e
	WM_CANCELMODE                                              uint32  = 0x1f
	WM_SETCURSOR                                               uint32  = 0x20
	WM_MOUSEACTIVATE                                           uint32  = 0x21
	WM_CHILDACTIVATE                                           uint32  = 0x22
	WM_QUEUESYNC                                               uint32  = 0x23
	WM_GETMINMAXINFO                                           uint32  = 0x24
	WM_PAINTICON                                               uint32  = 0x26
	WM_ICONERASEBKGND                                          uint32  = 0x27
	WM_NEXTDLGCTL                                              uint32  = 0x28
	WM_SPOOLERSTATUS                                           uint32  = 0x2a
	WM_DRAWITEM                                                uint32  = 0x2b
	WM_MEASUREITEM                                             uint32  = 0x2c
	WM_DELETEITEM                                              uint32  = 0x2d
	WM_VKEYTOITEM                                              uint32  = 0x2e
	WM_CHARTOITEM                                              uint32  = 0x2f
	WM_SETFONT                                                 uint32  = 0x30
	WM_GETFONT                                                 uint32  = 0x31
	WM_SETHOTKEY                                               uint32  = 0x32
	WM_GETHOTKEY                                               uint32  = 0x33
	WM_QUERYDRAGICON                                           uint32  = 0x37
	WM_COMPAREITEM                                             uint32  = 0x39
	WM_GETOBJECT                                               uint32  = 0x3d
	WM_COMPACTING                                              uint32  = 0x41
	WM_COMMNOTIFY                                              uint32  = 0x44
	WM_WINDOWPOSCHANGING                                       uint32  = 0x46
	WM_WINDOWPOSCHANGED                                        uint32  = 0x47
	WM_POWER                                                   uint32  = 0x48
	PWR_OK                                                     uint32  = 0x1
	PWR_FAIL                                                   int32   = -1
	PWR_SUSPENDREQUEST                                         uint32  = 0x1
	PWR_SUSPENDRESUME                                          uint32  = 0x2
	PWR_CRITICALRESUME                                         uint32  = 0x3
	WM_COPYDATA                                                uint32  = 0x4a
	WM_CANCELJOURNAL                                           uint32  = 0x4b
	WM_INPUTLANGCHANGEREQUEST                                  uint32  = 0x50
	WM_INPUTLANGCHANGE                                         uint32  = 0x51
	WM_TCARD                                                   uint32  = 0x52
	WM_HELP                                                    uint32  = 0x53
	WM_USERCHANGED                                             uint32  = 0x54
	WM_NOTIFYFORMAT                                            uint32  = 0x55
	NFR_ANSI                                                   uint32  = 0x1
	NFR_UNICODE                                                uint32  = 0x2
	NF_QUERY                                                   uint32  = 0x3
	NF_REQUERY                                                 uint32  = 0x4
	WM_STYLECHANGING                                           uint32  = 0x7c
	WM_STYLECHANGED                                            uint32  = 0x7d
	WM_DISPLAYCHANGE                                           uint32  = 0x7e
	WM_GETICON                                                 uint32  = 0x7f
	WM_SETICON                                                 uint32  = 0x80
	WM_NCCREATE                                                uint32  = 0x81
	WM_NCDESTROY                                               uint32  = 0x82
	WM_NCCALCSIZE                                              uint32  = 0x83
	WM_NCHITTEST                                               uint32  = 0x84
	WM_NCPAINT                                                 uint32  = 0x85
	WM_NCACTIVATE                                              uint32  = 0x86
	WM_GETDLGCODE                                              uint32  = 0x87
	WM_SYNCPAINT                                               uint32  = 0x88
	WM_NCMOUSEMOVE                                             uint32  = 0xa0
	WM_NCLBUTTONDOWN                                           uint32  = 0xa1
	WM_NCLBUTTONUP                                             uint32  = 0xa2
	WM_NCLBUTTONDBLCLK                                         uint32  = 0xa3
	WM_NCRBUTTONDOWN                                           uint32  = 0xa4
	WM_NCRBUTTONUP                                             uint32  = 0xa5
	WM_NCRBUTTONDBLCLK                                         uint32  = 0xa6
	WM_NCMBUTTONDOWN                                           uint32  = 0xa7
	WM_NCMBUTTONUP                                             uint32  = 0xa8
	WM_NCMBUTTONDBLCLK                                         uint32  = 0xa9
	WM_NCXBUTTONDOWN                                           uint32  = 0xab
	WM_NCXBUTTONUP                                             uint32  = 0xac
	WM_NCXBUTTONDBLCLK                                         uint32  = 0xad
	WM_INPUT_DEVICE_CHANGE                                     uint32  = 0xfe
	WM_INPUT                                                   uint32  = 0xff
	WM_KEYFIRST                                                uint32  = 0x100
	WM_KEYDOWN                                                 uint32  = 0x100
	WM_KEYUP                                                   uint32  = 0x101
	WM_CHAR                                                    uint32  = 0x102
	WM_DEADCHAR                                                uint32  = 0x103
	WM_SYSKEYDOWN                                              uint32  = 0x104
	WM_SYSKEYUP                                                uint32  = 0x105
	WM_SYSCHAR                                                 uint32  = 0x106
	WM_SYSDEADCHAR                                             uint32  = 0x107
	WM_KEYLAST                                                 uint32  = 0x109
	UNICODE_NOCHAR                                             uint32  = 0xffff
	WM_IME_STARTCOMPOSITION                                    uint32  = 0x10d
	WM_IME_ENDCOMPOSITION                                      uint32  = 0x10e
	WM_IME_COMPOSITION                                         uint32  = 0x10f
	WM_IME_KEYLAST                                             uint32  = 0x10f
	WM_INITDIALOG                                              uint32  = 0x110
	WM_COMMAND                                                 uint32  = 0x111
	WM_SYSCOMMAND                                              uint32  = 0x112
	WM_TIMER                                                   uint32  = 0x113
	WM_HSCROLL                                                 uint32  = 0x114
	WM_VSCROLL                                                 uint32  = 0x115
	WM_INITMENU                                                uint32  = 0x116
	WM_INITMENUPOPUP                                           uint32  = 0x117
	WM_GESTURE                                                 uint32  = 0x119
	WM_GESTURENOTIFY                                           uint32  = 0x11a
	WM_MENUSELECT                                              uint32  = 0x11f
	WM_MENUCHAR                                                uint32  = 0x120
	WM_ENTERIDLE                                               uint32  = 0x121
	WM_MENURBUTTONUP                                           uint32  = 0x122
	WM_MENUDRAG                                                uint32  = 0x123
	WM_MENUGETOBJECT                                           uint32  = 0x124
	WM_UNINITMENUPOPUP                                         uint32  = 0x125
	WM_MENUCOMMAND                                             uint32  = 0x126
	WM_CHANGEUISTATE                                           uint32  = 0x127
	WM_UPDATEUISTATE                                           uint32  = 0x128
	WM_QUERYUISTATE                                            uint32  = 0x129
	UIS_SET                                                    uint32  = 0x1
	UIS_CLEAR                                                  uint32  = 0x2
	UIS_INITIALIZE                                             uint32  = 0x3
	UISF_HIDEFOCUS                                             uint32  = 0x1
	UISF_HIDEACCEL                                             uint32  = 0x2
	UISF_ACTIVE                                                uint32  = 0x4
	WM_CTLCOLORMSGBOX                                          uint32  = 0x132
	WM_CTLCOLOREDIT                                            uint32  = 0x133
	WM_CTLCOLORLISTBOX                                         uint32  = 0x134
	WM_CTLCOLORBTN                                             uint32  = 0x135
	WM_CTLCOLORDLG                                             uint32  = 0x136
	WM_CTLCOLORSCROLLBAR                                       uint32  = 0x137
	WM_CTLCOLORSTATIC                                          uint32  = 0x138
	MN_GETHMENU                                                uint32  = 0x1e1
	WM_MOUSEFIRST                                              uint32  = 0x200
	WM_MOUSEMOVE                                               uint32  = 0x200
	WM_LBUTTONDOWN                                             uint32  = 0x201
	WM_LBUTTONUP                                               uint32  = 0x202
	WM_LBUTTONDBLCLK                                           uint32  = 0x203
	WM_RBUTTONDOWN                                             uint32  = 0x204
	WM_RBUTTONUP                                               uint32  = 0x205
	WM_RBUTTONDBLCLK                                           uint32  = 0x206
	WM_MBUTTONDOWN                                             uint32  = 0x207
	WM_MBUTTONUP                                               uint32  = 0x208
	WM_MBUTTONDBLCLK                                           uint32  = 0x209
	WM_MOUSEWHEEL                                              uint32  = 0x20a
	WM_XBUTTONDOWN                                             uint32  = 0x20b
	WM_XBUTTONUP                                               uint32  = 0x20c
	WM_XBUTTONDBLCLK                                           uint32  = 0x20d
	WM_MOUSEHWHEEL                                             uint32  = 0x20e
	WM_MOUSELAST                                               uint32  = 0x20e
	WHEEL_DELTA                                                uint32  = 0x78
	WM_PARENTNOTIFY                                            uint32  = 0x210
	WM_ENTERMENULOOP                                           uint32  = 0x211
	WM_EXITMENULOOP                                            uint32  = 0x212
	WM_NEXTMENU                                                uint32  = 0x213
	WM_SIZING                                                  uint32  = 0x214
	WM_CAPTURECHANGED                                          uint32  = 0x215
	WM_MOVING                                                  uint32  = 0x216
	WM_POWERBROADCAST                                          uint32  = 0x218
	PBT_APMQUERYSUSPEND                                        uint32  = 0x0
	PBT_APMQUERYSTANDBY                                        uint32  = 0x1
	PBT_APMQUERYSUSPENDFAILED                                  uint32  = 0x2
	PBT_APMQUERYSTANDBYFAILED                                  uint32  = 0x3
	PBT_APMSUSPEND                                             uint32  = 0x4
	PBT_APMSTANDBY                                             uint32  = 0x5
	PBT_APMRESUMECRITICAL                                      uint32  = 0x6
	PBT_APMRESUMESUSPEND                                       uint32  = 0x7
	PBT_APMRESUMESTANDBY                                       uint32  = 0x8
	PBTF_APMRESUMEFROMFAILURE                                  uint32  = 0x1
	PBT_APMBATTERYLOW                                          uint32  = 0x9
	PBT_APMPOWERSTATUSCHANGE                                   uint32  = 0xa
	PBT_APMOEMEVENT                                            uint32  = 0xb
	PBT_APMRESUMEAUTOMATIC                                     uint32  = 0x12
	PBT_POWERSETTINGCHANGE                                     uint32  = 0x8013
	WM_MDICREATE                                               uint32  = 0x220
	WM_MDIDESTROY                                              uint32  = 0x221
	WM_MDIACTIVATE                                             uint32  = 0x222
	WM_MDIRESTORE                                              uint32  = 0x223
	WM_MDINEXT                                                 uint32  = 0x224
	WM_MDIMAXIMIZE                                             uint32  = 0x225
	WM_MDITILE                                                 uint32  = 0x226
	WM_MDICASCADE                                              uint32  = 0x227
	WM_MDIICONARRANGE                                          uint32  = 0x228
	WM_MDIGETACTIVE                                            uint32  = 0x229
	WM_MDISETMENU                                              uint32  = 0x230
	WM_ENTERSIZEMOVE                                           uint32  = 0x231
	WM_EXITSIZEMOVE                                            uint32  = 0x232
	WM_DROPFILES                                               uint32  = 0x233
	WM_MDIREFRESHMENU                                          uint32  = 0x234
	WM_POINTERDEVICECHANGE                                     uint32  = 0x238
	WM_POINTERDEVICEINRANGE                                    uint32  = 0x239
	WM_POINTERDEVICEOUTOFRANGE                                 uint32  = 0x23a
	WM_TOUCH                                                   uint32  = 0x240
	WM_NCPOINTERUPDATE                                         uint32  = 0x241
	WM_NCPOINTERDOWN                                           uint32  = 0x242
	WM_NCPOINTERUP                                             uint32  = 0x243
	WM_POINTERUPDATE                                           uint32  = 0x245
	WM_POINTERDOWN                                             uint32  = 0x246
	WM_POINTERUP                                               uint32  = 0x247
	WM_POINTERENTER                                            uint32  = 0x249
	WM_POINTERLEAVE                                            uint32  = 0x24a
	WM_POINTERACTIVATE                                         uint32  = 0x24b
	WM_POINTERCAPTURECHANGED                                   uint32  = 0x24c
	WM_TOUCHHITTESTING                                         uint32  = 0x24d
	WM_POINTERWHEEL                                            uint32  = 0x24e
	WM_POINTERHWHEEL                                           uint32  = 0x24f
	DM_POINTERHITTEST                                          uint32  = 0x250
	WM_POINTERROUTEDTO                                         uint32  = 0x251
	WM_POINTERROUTEDAWAY                                       uint32  = 0x252
	WM_POINTERROUTEDRELEASED                                   uint32  = 0x253
	WM_IME_SETCONTEXT                                          uint32  = 0x281
	WM_IME_NOTIFY                                              uint32  = 0x282
	WM_IME_CONTROL                                             uint32  = 0x283
	WM_IME_COMPOSITIONFULL                                     uint32  = 0x284
	WM_IME_SELECT                                              uint32  = 0x285
	WM_IME_CHAR                                                uint32  = 0x286
	WM_IME_REQUEST                                             uint32  = 0x288
	WM_IME_KEYDOWN                                             uint32  = 0x290
	WM_IME_KEYUP                                               uint32  = 0x291
	WM_NCMOUSEHOVER                                            uint32  = 0x2a0
	WM_NCMOUSELEAVE                                            uint32  = 0x2a2
	WM_WTSSESSION_CHANGE                                       uint32  = 0x2b1
	WM_TABLET_FIRST                                            uint32  = 0x2c0
	WM_TABLET_LAST                                             uint32  = 0x2df
	WM_DPICHANGED                                              uint32  = 0x2e0
	WM_DPICHANGED_BEFOREPARENT                                 uint32  = 0x2e2
	WM_DPICHANGED_AFTERPARENT                                  uint32  = 0x2e3
	WM_GETDPISCALEDSIZE                                        uint32  = 0x2e4
	WM_CUT                                                     uint32  = 0x300
	WM_COPY                                                    uint32  = 0x301
	WM_PASTE                                                   uint32  = 0x302
	WM_CLEAR                                                   uint32  = 0x303
	WM_UNDO                                                    uint32  = 0x304
	WM_RENDERFORMAT                                            uint32  = 0x305
	WM_RENDERALLFORMATS                                        uint32  = 0x306
	WM_DESTROYCLIPBOARD                                        uint32  = 0x307
	WM_DRAWCLIPBOARD                                           uint32  = 0x308
	WM_PAINTCLIPBOARD                                          uint32  = 0x309
	WM_VSCROLLCLIPBOARD                                        uint32  = 0x30a
	WM_SIZECLIPBOARD                                           uint32  = 0x30b
	WM_ASKCBFORMATNAME                                         uint32  = 0x30c
	WM_CHANGECBCHAIN                                           uint32  = 0x30d
	WM_HSCROLLCLIPBOARD                                        uint32  = 0x30e
	WM_QUERYNEWPALETTE                                         uint32  = 0x30f
	WM_PALETTEISCHANGING                                       uint32  = 0x310
	WM_PALETTECHANGED                                          uint32  = 0x311
	WM_HOTKEY                                                  uint32  = 0x312
	WM_PRINT                                                   uint32  = 0x317
	WM_APPCOMMAND                                              uint32  = 0x319
	WM_THEMECHANGED                                            uint32  = 0x31a
	WM_CLIPBOARDUPDATE                                         uint32  = 0x31d
	WM_DWMCOMPOSITIONCHANGED                                   uint32  = 0x31e
	WM_DWMNCRENDERINGCHANGED                                   uint32  = 0x31f
	WM_DWMCOLORIZATIONCOLORCHANGED                             uint32  = 0x320
	WM_DWMWINDOWMAXIMIZEDCHANGE                                uint32  = 0x321
	WM_DWMSENDICONICTHUMBNAIL                                  uint32  = 0x323
	WM_DWMSENDICONICLIVEPREVIEWBITMAP                          uint32  = 0x326
	WM_GETTITLEBARINFOEX                                       uint32  = 0x33f
	WM_HANDHELDFIRST                                           uint32  = 0x358
	WM_HANDHELDLAST                                            uint32  = 0x35f
	WM_AFXFIRST                                                uint32  = 0x360
	WM_AFXLAST                                                 uint32  = 0x37f
	WM_PENWINFIRST                                             uint32  = 0x380
	WM_PENWINLAST                                              uint32  = 0x38f
	WM_APP                                                     uint32  = 0x8000
	WM_USER                                                    uint32  = 0x400
	WMSZ_LEFT                                                  uint32  = 0x1
	WMSZ_RIGHT                                                 uint32  = 0x2
	WMSZ_TOP                                                   uint32  = 0x3
	WMSZ_TOPLEFT                                               uint32  = 0x4
	WMSZ_TOPRIGHT                                              uint32  = 0x5
	WMSZ_BOTTOM                                                uint32  = 0x6
	WMSZ_BOTTOMLEFT                                            uint32  = 0x7
	WMSZ_BOTTOMRIGHT                                           uint32  = 0x8
	HTERROR                                                    int32   = -2
	HTTRANSPARENT                                              int32   = -1
	HTNOWHERE                                                  uint32  = 0x0
	HTCLIENT                                                   uint32  = 0x1
	HTCAPTION                                                  uint32  = 0x2
	HTSYSMENU                                                  uint32  = 0x3
	HTGROWBOX                                                  uint32  = 0x4
	HTSIZE                                                     uint32  = 0x4
	HTMENU                                                     uint32  = 0x5
	HTHSCROLL                                                  uint32  = 0x6
	HTVSCROLL                                                  uint32  = 0x7
	HTMINBUTTON                                                uint32  = 0x8
	HTMAXBUTTON                                                uint32  = 0x9
	HTLEFT                                                     uint32  = 0xa
	HTRIGHT                                                    uint32  = 0xb
	HTTOP                                                      uint32  = 0xc
	HTTOPLEFT                                                  uint32  = 0xd
	HTTOPRIGHT                                                 uint32  = 0xe
	HTBOTTOM                                                   uint32  = 0xf
	HTBOTTOMLEFT                                               uint32  = 0x10
	HTBOTTOMRIGHT                                              uint32  = 0x11
	HTBORDER                                                   uint32  = 0x12
	HTREDUCE                                                   uint32  = 0x8
	HTZOOM                                                     uint32  = 0x9
	HTSIZEFIRST                                                uint32  = 0xa
	HTSIZELAST                                                 uint32  = 0x11
	HTOBJECT                                                   uint32  = 0x13
	HTCLOSE                                                    uint32  = 0x14
	HTHELP                                                     uint32  = 0x15
	MA_ACTIVATE                                                uint32  = 0x1
	MA_ACTIVATEANDEAT                                          uint32  = 0x2
	MA_NOACTIVATE                                              uint32  = 0x3
	MA_NOACTIVATEANDEAT                                        uint32  = 0x4
	ICON_SMALL                                                 uint32  = 0x0
	ICON_BIG                                                   uint32  = 0x1
	ICON_SMALL2                                                uint32  = 0x2
	SIZE_RESTORED                                              uint32  = 0x0
	SIZE_MINIMIZED                                             uint32  = 0x1
	SIZE_MAXIMIZED                                             uint32  = 0x2
	SIZE_MAXSHOW                                               uint32  = 0x3
	SIZE_MAXHIDE                                               uint32  = 0x4
	SIZENORMAL                                                 uint32  = 0x0
	SIZEICONIC                                                 uint32  = 0x1
	SIZEFULLSCREEN                                             uint32  = 0x2
	SIZEZOOMSHOW                                               uint32  = 0x3
	SIZEZOOMHIDE                                               uint32  = 0x4
	WVR_ALIGNTOP                                               uint32  = 0x10
	WVR_ALIGNLEFT                                              uint32  = 0x20
	WVR_ALIGNBOTTOM                                            uint32  = 0x40
	WVR_ALIGNRIGHT                                             uint32  = 0x80
	WVR_HREDRAW                                                uint32  = 0x100
	WVR_VREDRAW                                                uint32  = 0x200
	WVR_VALIDRECTS                                             uint32  = 0x400
	MK_LBUTTON                                                 uint32  = 0x1
	MK_RBUTTON                                                 uint32  = 0x2
	MK_SHIFT                                                   uint32  = 0x4
	MK_CONTROL                                                 uint32  = 0x8
	MK_MBUTTON                                                 uint32  = 0x10
	MK_XBUTTON1                                                uint32  = 0x20
	MK_XBUTTON2                                                uint32  = 0x40
	PRF_CHECKVISIBLE                                           int32   = 1
	PRF_NONCLIENT                                              int32   = 2
	PRF_CLIENT                                                 int32   = 4
	PRF_ERASEBKGND                                             int32   = 8
	PRF_CHILDREN                                               int32   = 16
	PRF_OWNED                                                  int32   = 32
	IDANI_OPEN                                                 uint32  = 0x1
	IDANI_CAPTION                                              uint32  = 0x3
	FVIRTKEY                                                   uint32  = 0x1
	FNOINVERT                                                  uint32  = 0x2
	FSHIFT                                                     uint32  = 0x4
	FCONTROL                                                   uint32  = 0x8
	FALT                                                       uint32  = 0x10
	ODA_DRAWENTIRE                                             uint32  = 0x1
	ODA_SELECT                                                 uint32  = 0x2
	ODA_FOCUS                                                  uint32  = 0x4
	ODS_SELECTED                                               uint32  = 0x1
	ODS_GRAYED                                                 uint32  = 0x2
	ODS_DISABLED                                               uint32  = 0x4
	ODS_CHECKED                                                uint32  = 0x8
	ODS_FOCUS                                                  uint32  = 0x10
	ODS_DEFAULT                                                uint32  = 0x20
	ODS_COMBOBOXEDIT                                           uint32  = 0x1000
	ODS_HOTLIGHT                                               uint32  = 0x40
	ODS_INACTIVE                                               uint32  = 0x80
	ODS_NOACCEL                                                uint32  = 0x100
	ODS_NOFOCUSRECT                                            uint32  = 0x200
	IDHOT_SNAPWINDOW                                           int32   = -1
	IDHOT_SNAPDESKTOP                                          int32   = -2
	ENDSESSION_CLOSEAPP                                        uint32  = 0x1
	ENDSESSION_CRITICAL                                        uint32  = 0x40000000
	ENDSESSION_LOGOFF                                          uint32  = 0x80000000
	EWX_FORCE                                                  uint32  = 0x4
	EWX_FORCEIFHUNG                                            uint32  = 0x10
	EWX_QUICKRESOLVE                                           uint32  = 0x20
	EWX_BOOTOPTIONS                                            uint32  = 0x1000000
	EWX_ARSO                                                   uint32  = 0x4000000
	EWX_CHECK_SAFE_FOR_SERVER                                  uint32  = 0x8000000
	EWX_SYSTEM_INITIATED                                       uint32  = 0x10000000
	BROADCAST_QUERY_DENY                                       uint32  = 0x424d5144
	DEVICE_NOTIFY_ALL_INTERFACE_CLASSES                        uint32  = 0x4
	HWND_MESSAGE                                               HWND    = ^HWND(0x2)
	ISMEX_NOSEND                                               uint32  = 0x0
	ISMEX_SEND                                                 uint32  = 0x1
	ISMEX_NOTIFY                                               uint32  = 0x2
	ISMEX_CALLBACK                                             uint32  = 0x4
	ISMEX_REPLIED                                              uint32  = 0x8
	HWND_DESKTOP                                               HWND    = 0
	PW_RENDERFULLCONTENT                                       uint32  = 0x2
	HWND_TOP                                                   HWND    = 0
	HWND_BOTTOM                                                HWND    = 1
	HWND_TOPMOST                                               HWND    = ^HWND(0x0)
	HWND_NOTOPMOST                                             HWND    = ^HWND(0x1)
	DLGWINDOWEXTRA                                             uint32  = 0x1e
	POINTER_MOD_SHIFT                                          uint32  = 0x4
	POINTER_MOD_CTRL                                           uint32  = 0x8
	TOUCH_FLAG_NONE                                            uint32  = 0x0
	TOUCH_MASK_NONE                                            uint32  = 0x0
	TOUCH_MASK_CONTACTAREA                                     uint32  = 0x1
	TOUCH_MASK_ORIENTATION                                     uint32  = 0x2
	TOUCH_MASK_PRESSURE                                        uint32  = 0x4
	PEN_FLAG_NONE                                              uint32  = 0x0
	PEN_FLAG_BARREL                                            uint32  = 0x1
	PEN_FLAG_INVERTED                                          uint32  = 0x2
	PEN_FLAG_ERASER                                            uint32  = 0x4
	PEN_MASK_NONE                                              uint32  = 0x0
	PEN_MASK_PRESSURE                                          uint32  = 0x1
	PEN_MASK_ROTATION                                          uint32  = 0x2
	PEN_MASK_TILT_X                                            uint32  = 0x4
	PEN_MASK_TILT_Y                                            uint32  = 0x8
	POINTER_MESSAGE_FLAG_NEW                                   uint32  = 0x1
	POINTER_MESSAGE_FLAG_INRANGE                               uint32  = 0x2
	POINTER_MESSAGE_FLAG_INCONTACT                             uint32  = 0x4
	POINTER_MESSAGE_FLAG_FIRSTBUTTON                           uint32  = 0x10
	POINTER_MESSAGE_FLAG_SECONDBUTTON                          uint32  = 0x20
	POINTER_MESSAGE_FLAG_THIRDBUTTON                           uint32  = 0x40
	POINTER_MESSAGE_FLAG_FOURTHBUTTON                          uint32  = 0x80
	POINTER_MESSAGE_FLAG_FIFTHBUTTON                           uint32  = 0x100
	POINTER_MESSAGE_FLAG_PRIMARY                               uint32  = 0x2000
	POINTER_MESSAGE_FLAG_CONFIDENCE                            uint32  = 0x4000
	POINTER_MESSAGE_FLAG_CANCELED                              uint32  = 0x8000
	PA_ACTIVATE                                                uint32  = 0x1
	PA_NOACTIVATE                                              uint32  = 0x3
	MAX_TOUCH_COUNT                                            uint32  = 0x100
	TOUCH_HIT_TESTING_DEFAULT                                  uint32  = 0x0
	TOUCH_HIT_TESTING_CLIENT                                   uint32  = 0x1
	TOUCH_HIT_TESTING_NONE                                     uint32  = 0x2
	TOUCH_HIT_TESTING_PROXIMITY_CLOSEST                        uint32  = 0x0
	TOUCH_HIT_TESTING_PROXIMITY_FARTHEST                       uint32  = 0xfff
	GWFS_INCLUDE_ANCESTORS                                     uint32  = 0x1
	MAPVK_VK_TO_VSC                                            uint32  = 0x0
	MAPVK_VSC_TO_VK                                            uint32  = 0x1
	MAPVK_VK_TO_CHAR                                           uint32  = 0x2
	MAPVK_VSC_TO_VK_EX                                         uint32  = 0x3
	MAPVK_VK_TO_VSC_EX                                         uint32  = 0x4
	QS_TOUCH                                                   uint32  = 0x800
	QS_POINTER                                                 uint32  = 0x1000
	USER_TIMER_MAXIMUM                                         uint32  = 0x7fffffff
	USER_TIMER_MINIMUM                                         uint32  = 0xa
	TIMERV_DEFAULT_COALESCING                                  uint32  = 0x0
	TIMERV_NO_COALESCING                                       uint32  = 0xffffffff
	TIMERV_COALESCING_MIN                                      uint32  = 0x1
	TIMERV_COALESCING_MAX                                      uint32  = 0x7ffffff5
	SM_RESERVED1                                               uint32  = 0x18
	SM_RESERVED2                                               uint32  = 0x19
	SM_RESERVED3                                               uint32  = 0x1a
	SM_RESERVED4                                               uint32  = 0x1b
	SM_CMETRICS                                                uint32  = 0x4c
	SM_CARETBLINKINGENABLED                                    uint32  = 0x2002
	PMB_ACTIVE                                                 uint32  = 0x1
	MNC_IGNORE                                                 uint32  = 0x0
	MNC_CLOSE                                                  uint32  = 0x1
	MNC_EXECUTE                                                uint32  = 0x2
	MNC_SELECT                                                 uint32  = 0x3
	MND_CONTINUE                                               uint32  = 0x0
	MND_ENDMENU                                                uint32  = 0x1
	MNGO_NOINTERFACE                                           uint32  = 0x0
	MNGO_NOERROR                                               uint32  = 0x1
	DOF_EXECUTABLE                                             uint32  = 0x8001
	DOF_DOCUMENT                                               uint32  = 0x8002
	DOF_DIRECTORY                                              uint32  = 0x8003
	DOF_MULTIPLE                                               uint32  = 0x8004
	DOF_PROGMAN                                                uint32  = 0x1
	DOF_SHELLDATA                                              uint32  = 0x2
	DO_DROPFILE                                                int32   = 1162627398
	DO_PRINTFILE                                               int32   = 1414419024
	ASFW_ANY                                                   uint32  = 0xffffffff
	DCX_EXCLUDEUPDATE                                          int32   = 256
	HELPINFO_WINDOW                                            uint32  = 0x1
	HELPINFO_MENUITEM                                          uint32  = 0x2
	CTLCOLOR_MSGBOX                                            uint32  = 0x0
	CTLCOLOR_EDIT                                              uint32  = 0x1
	CTLCOLOR_LISTBOX                                           uint32  = 0x2
	CTLCOLOR_BTN                                               uint32  = 0x3
	CTLCOLOR_DLG                                               uint32  = 0x4
	CTLCOLOR_SCROLLBAR                                         uint32  = 0x5
	CTLCOLOR_STATIC                                            uint32  = 0x6
	CTLCOLOR_MAX                                               uint32  = 0x7
	COLOR_BTNHIGHLIGHT                                         uint32  = 0x14
	COLOR_BTNHILIGHT                                           uint32  = 0x14
	GW_MAX                                                     uint32  = 0x5
	SC_SIZE                                                    uint32  = 0xf000
	SC_MOVE                                                    uint32  = 0xf010
	SC_MINIMIZE                                                uint32  = 0xf020
	SC_MAXIMIZE                                                uint32  = 0xf030
	SC_NEXTWINDOW                                              uint32  = 0xf040
	SC_PREVWINDOW                                              uint32  = 0xf050
	SC_CLOSE                                                   uint32  = 0xf060
	SC_VSCROLL                                                 uint32  = 0xf070
	SC_HSCROLL                                                 uint32  = 0xf080
	SC_MOUSEMENU                                               uint32  = 0xf090
	SC_KEYMENU                                                 uint32  = 0xf100
	SC_ARRANGE                                                 uint32  = 0xf110
	SC_RESTORE                                                 uint32  = 0xf120
	SC_TASKLIST                                                uint32  = 0xf130
	SC_HOTKEY                                                  uint32  = 0xf150
	SC_DEFAULT                                                 uint32  = 0xf160
	SC_MONITORPOWER                                            uint32  = 0xf170
	SC_CONTEXTHELP                                             uint32  = 0xf180
	SC_SEPARATOR                                               uint32  = 0xf00f
	SCF_ISSECURE                                               uint32  = 0x1
	SC_ICON                                                    uint32  = 0xf020
	SC_ZOOM                                                    uint32  = 0xf030
	CURSOR_CREATION_SCALING_NONE                               uint32  = 0x1
	CURSOR_CREATION_SCALING_DEFAULT                            uint32  = 0x2
	IMAGE_ENHMETAFILE                                          uint32  = 0x3
	LR_COLOR                                                   uint32  = 0x2
	RES_ICON                                                   uint32  = 0x1
	RES_CURSOR                                                 uint32  = 0x2
	OBM_CLOSE                                                  uint32  = 0x7ff2
	OBM_UPARROW                                                uint32  = 0x7ff1
	OBM_DNARROW                                                uint32  = 0x7ff0
	OBM_RGARROW                                                uint32  = 0x7fef
	OBM_LFARROW                                                uint32  = 0x7fee
	OBM_REDUCE                                                 uint32  = 0x7fed
	OBM_ZOOM                                                   uint32  = 0x7fec
	OBM_RESTORE                                                uint32  = 0x7feb
	OBM_REDUCED                                                uint32  = 0x7fea
	OBM_ZOOMD                                                  uint32  = 0x7fe9
	OBM_RESTORED                                               uint32  = 0x7fe8
	OBM_UPARROWD                                               uint32  = 0x7fe7
	OBM_DNARROWD                                               uint32  = 0x7fe6
	OBM_RGARROWD                                               uint32  = 0x7fe5
	OBM_LFARROWD                                               uint32  = 0x7fe4
	OBM_MNARROW                                                uint32  = 0x7fe3
	OBM_COMBO                                                  uint32  = 0x7fe2
	OBM_UPARROWI                                               uint32  = 0x7fe1
	OBM_DNARROWI                                               uint32  = 0x7fe0
	OBM_RGARROWI                                               uint32  = 0x7fdf
	OBM_LFARROWI                                               uint32  = 0x7fde
	OBM_OLD_CLOSE                                              uint32  = 0x7fff
	OBM_SIZE                                                   uint32  = 0x7ffe
	OBM_OLD_UPARROW                                            uint32  = 0x7ffd
	OBM_OLD_DNARROW                                            uint32  = 0x7ffc
	OBM_OLD_RGARROW                                            uint32  = 0x7ffb
	OBM_OLD_LFARROW                                            uint32  = 0x7ffa
	OBM_BTSIZE                                                 uint32  = 0x7ff9
	OBM_CHECK                                                  uint32  = 0x7ff8
	OBM_CHECKBOXES                                             uint32  = 0x7ff7
	OBM_BTNCORNERS                                             uint32  = 0x7ff6
	OBM_OLD_REDUCE                                             uint32  = 0x7ff5
	OBM_OLD_ZOOM                                               uint32  = 0x7ff4
	OBM_OLD_RESTORE                                            uint32  = 0x7ff3
	OCR_SIZE                                                   uint32  = 0x7f80
	OCR_ICON                                                   uint32  = 0x7f81
	OCR_ICOCUR                                                 uint32  = 0x7f87
	OIC_SAMPLE                                                 uint32  = 0x7f00
	OIC_HAND                                                   uint32  = 0x7f01
	OIC_QUES                                                   uint32  = 0x7f02
	OIC_BANG                                                   uint32  = 0x7f03
	OIC_NOTE                                                   uint32  = 0x7f04
	OIC_WINLOGO                                                uint32  = 0x7f05
	OIC_WARNING                                                uint32  = 0x7f03
	OIC_ERROR                                                  uint32  = 0x7f01
	OIC_INFORMATION                                            uint32  = 0x7f04
	OIC_SHIELD                                                 uint32  = 0x7f06
	ORD_LANGDRIVER                                             uint32  = 0x1
	IDI_WARNING                                                uint32  = 0x7f03
	IDI_ERROR                                                  uint32  = 0x7f01
	IDI_INFORMATION                                            uint32  = 0x7f04
	ES_LEFT                                                    int32   = 0
	ES_CENTER                                                  int32   = 1
	ES_RIGHT                                                   int32   = 2
	ES_MULTILINE                                               int32   = 4
	ES_UPPERCASE                                               int32   = 8
	ES_LOWERCASE                                               int32   = 16
	ES_PASSWORD                                                int32   = 32
	ES_AUTOVSCROLL                                             int32   = 64
	ES_AUTOHSCROLL                                             int32   = 128
	ES_NOHIDESEL                                               int32   = 256
	ES_OEMCONVERT                                              int32   = 1024
	ES_READONLY                                                int32   = 2048
	ES_WANTRETURN                                              int32   = 4096
	ES_NUMBER                                                  int32   = 8192
	EN_SETFOCUS                                                uint32  = 0x100
	EN_KILLFOCUS                                               uint32  = 0x200
	EN_CHANGE                                                  uint32  = 0x300
	EN_UPDATE                                                  uint32  = 0x400
	EN_ERRSPACE                                                uint32  = 0x500
	EN_MAXTEXT                                                 uint32  = 0x501
	EN_HSCROLL                                                 uint32  = 0x601
	EN_VSCROLL                                                 uint32  = 0x602
	EN_ALIGN_LTR_EC                                            uint32  = 0x700
	EN_ALIGN_RTL_EC                                            uint32  = 0x701
	EN_BEFORE_PASTE                                            uint32  = 0x800
	EN_AFTER_PASTE                                             uint32  = 0x801
	EC_LEFTMARGIN                                              uint32  = 0x1
	EC_RIGHTMARGIN                                             uint32  = 0x2
	EC_USEFONTINFO                                             uint32  = 0xffff
	EMSIS_COMPOSITIONSTRING                                    uint32  = 0x1
	EIMES_GETCOMPSTRATONCE                                     uint32  = 0x1
	EIMES_CANCELCOMPSTRINFOCUS                                 uint32  = 0x2
	EIMES_COMPLETECOMPSTRKILLFOCUS                             uint32  = 0x4
	BS_PUSHBUTTON                                              int32   = 0
	BS_DEFPUSHBUTTON                                           int32   = 1
	BS_CHECKBOX                                                int32   = 2
	BS_AUTOCHECKBOX                                            int32   = 3
	BS_RADIOBUTTON                                             int32   = 4
	BS_3STATE                                                  int32   = 5
	BS_AUTO3STATE                                              int32   = 6
	BS_GROUPBOX                                                int32   = 7
	BS_USERBUTTON                                              int32   = 8
	BS_AUTORADIOBUTTON                                         int32   = 9
	BS_PUSHBOX                                                 int32   = 10
	BS_OWNERDRAW                                               int32   = 11
	BS_TYPEMASK                                                int32   = 15
	BS_LEFTTEXT                                                int32   = 32
	BS_TEXT                                                    int32   = 0
	BS_ICON                                                    int32   = 64
	BS_BITMAP                                                  int32   = 128
	BS_LEFT                                                    int32   = 256
	BS_RIGHT                                                   int32   = 512
	BS_CENTER                                                  int32   = 768
	BS_TOP                                                     int32   = 1024
	BS_BOTTOM                                                  int32   = 2048
	BS_VCENTER                                                 int32   = 3072
	BS_PUSHLIKE                                                int32   = 4096
	BS_MULTILINE                                               int32   = 8192
	BS_NOTIFY                                                  int32   = 16384
	BS_FLAT                                                    int32   = 32768
	BS_RIGHTBUTTON                                             int32   = 32
	BN_CLICKED                                                 uint32  = 0x0
	BN_PAINT                                                   uint32  = 0x1
	BN_HILITE                                                  uint32  = 0x2
	BN_UNHILITE                                                uint32  = 0x3
	BN_DISABLE                                                 uint32  = 0x4
	BN_DOUBLECLICKED                                           uint32  = 0x5
	BN_PUSHED                                                  uint32  = 0x2
	BN_UNPUSHED                                                uint32  = 0x3
	BN_DBLCLK                                                  uint32  = 0x5
	BN_SETFOCUS                                                uint32  = 0x6
	BN_KILLFOCUS                                               uint32  = 0x7
	BM_GETCHECK                                                uint32  = 0xf0
	BM_SETCHECK                                                uint32  = 0xf1
	BM_GETSTATE                                                uint32  = 0xf2
	BM_SETSTATE                                                uint32  = 0xf3
	BM_SETSTYLE                                                uint32  = 0xf4
	BM_CLICK                                                   uint32  = 0xf5
	BM_GETIMAGE                                                uint32  = 0xf6
	BM_SETIMAGE                                                uint32  = 0xf7
	BM_SETDONTCLICK                                            uint32  = 0xf8
	BST_PUSHED                                                 uint32  = 0x4
	BST_FOCUS                                                  uint32  = 0x8
	SS_LEFT                                                    int32   = 0
	SS_CENTER                                                  int32   = 1
	SS_RIGHT                                                   int32   = 2
	SS_ICON                                                    int32   = 3
	SS_BLACKRECT                                               int32   = 4
	SS_GRAYRECT                                                int32   = 5
	SS_WHITERECT                                               int32   = 6
	SS_BLACKFRAME                                              int32   = 7
	SS_GRAYFRAME                                               int32   = 8
	SS_WHITEFRAME                                              int32   = 9
	SS_USERITEM                                                int32   = 10
	SS_SIMPLE                                                  int32   = 11
	SS_LEFTNOWORDWRAP                                          int32   = 12
	SS_OWNERDRAW                                               int32   = 13
	SS_BITMAP                                                  int32   = 14
	SS_ENHMETAFILE                                             int32   = 15
	SS_ETCHEDHORZ                                              int32   = 16
	SS_ETCHEDVERT                                              int32   = 17
	SS_ETCHEDFRAME                                             int32   = 18
	SS_TYPEMASK                                                int32   = 31
	SS_REALSIZECONTROL                                         int32   = 64
	SS_NOPREFIX                                                int32   = 128
	SS_NOTIFY                                                  int32   = 256
	SS_CENTERIMAGE                                             int32   = 512
	SS_RIGHTJUST                                               int32   = 1024
	SS_REALSIZEIMAGE                                           int32   = 2048
	SS_SUNKEN                                                  int32   = 4096
	SS_EDITCONTROL                                             int32   = 8192
	SS_ENDELLIPSIS                                             int32   = 16384
	SS_PATHELLIPSIS                                            int32   = 32768
	SS_WORDELLIPSIS                                            int32   = 49152
	SS_ELLIPSISMASK                                            int32   = 49152
	STM_SETICON                                                uint32  = 0x170
	STM_GETICON                                                uint32  = 0x171
	STM_SETIMAGE                                               uint32  = 0x172
	STM_GETIMAGE                                               uint32  = 0x173
	STN_CLICKED                                                uint32  = 0x0
	STN_DBLCLK                                                 uint32  = 0x1
	STN_ENABLE                                                 uint32  = 0x2
	STN_DISABLE                                                uint32  = 0x3
	STM_MSGMAX                                                 uint32  = 0x174
	DWL_MSGRESULT                                              uint32  = 0x0
	DWL_DLGPROC                                                uint32  = 0x4
	DWL_USER                                                   uint32  = 0x8
	DWLP_MSGRESULT                                             uint32  = 0x0
	DS_ABSALIGN                                                int32   = 1
	DS_SYSMODAL                                                int32   = 2
	DS_LOCALEDIT                                               int32   = 32
	DS_SETFONT                                                 int32   = 64
	DS_MODALFRAME                                              int32   = 128
	DS_NOIDLEMSG                                               int32   = 256
	DS_SETFOREGROUND                                           int32   = 512
	DS_3DLOOK                                                  int32   = 4
	DS_FIXEDSYS                                                int32   = 8
	DS_NOFAILCREATE                                            int32   = 16
	DS_CONTROL                                                 int32   = 1024
	DS_CENTER                                                  int32   = 2048
	DS_CENTERMOUSE                                             int32   = 4096
	DS_CONTEXTHELP                                             int32   = 8192
	DS_USEPIXELS                                               int32   = 32768
	DM_GETDEFID                                                uint32  = 0x400
	DM_SETDEFID                                                uint32  = 0x401
	DM_REPOSITION                                              uint32  = 0x402
	DC_HASDEFID                                                uint32  = 0x534b
	DLGC_WANTARROWS                                            uint32  = 0x1
	DLGC_WANTTAB                                               uint32  = 0x2
	DLGC_WANTALLKEYS                                           uint32  = 0x4
	DLGC_WANTMESSAGE                                           uint32  = 0x4
	DLGC_HASSETSEL                                             uint32  = 0x8
	DLGC_DEFPUSHBUTTON                                         uint32  = 0x10
	DLGC_UNDEFPUSHBUTTON                                       uint32  = 0x20
	DLGC_RADIOBUTTON                                           uint32  = 0x40
	DLGC_WANTCHARS                                             uint32  = 0x80
	DLGC_STATIC                                                uint32  = 0x100
	DLGC_BUTTON                                                uint32  = 0x2000
	LB_CTLCODE                                                 int32   = 0
	LB_OKAY                                                    uint32  = 0x0
	LB_ERR                                                     int32   = -1
	LB_ERRSPACE                                                int32   = -2
	LBN_ERRSPACE                                               int32   = -2
	LBN_SELCHANGE                                              uint32  = 0x1
	LBN_DBLCLK                                                 uint32  = 0x2
	LBN_SELCANCEL                                              uint32  = 0x3
	LBN_SETFOCUS                                               uint32  = 0x4
	LBN_KILLFOCUS                                              uint32  = 0x5
	LB_ADDSTRING                                               uint32  = 0x180
	LB_INSERTSTRING                                            uint32  = 0x181
	LB_DELETESTRING                                            uint32  = 0x182
	LB_SELITEMRANGEEX                                          uint32  = 0x183
	LB_RESETCONTENT                                            uint32  = 0x184
	LB_SETSEL                                                  uint32  = 0x185
	LB_SETCURSEL                                               uint32  = 0x186
	LB_GETSEL                                                  uint32  = 0x187
	LB_GETCURSEL                                               uint32  = 0x188
	LB_GETTEXT                                                 uint32  = 0x189
	LB_GETTEXTLEN                                              uint32  = 0x18a
	LB_GETCOUNT                                                uint32  = 0x18b
	LB_SELECTSTRING                                            uint32  = 0x18c
	LB_DIR                                                     uint32  = 0x18d
	LB_GETTOPINDEX                                             uint32  = 0x18e
	LB_FINDSTRING                                              uint32  = 0x18f
	LB_GETSELCOUNT                                             uint32  = 0x190
	LB_GETSELITEMS                                             uint32  = 0x191
	LB_SETTABSTOPS                                             uint32  = 0x192
	LB_GETHORIZONTALEXTENT                                     uint32  = 0x193
	LB_SETHORIZONTALEXTENT                                     uint32  = 0x194
	LB_SETCOLUMNWIDTH                                          uint32  = 0x195
	LB_ADDFILE                                                 uint32  = 0x196
	LB_SETTOPINDEX                                             uint32  = 0x197
	LB_GETITEMRECT                                             uint32  = 0x198
	LB_GETITEMDATA                                             uint32  = 0x199
	LB_SETITEMDATA                                             uint32  = 0x19a
	LB_SELITEMRANGE                                            uint32  = 0x19b
	LB_SETANCHORINDEX                                          uint32  = 0x19c
	LB_GETANCHORINDEX                                          uint32  = 0x19d
	LB_SETCARETINDEX                                           uint32  = 0x19e
	LB_GETCARETINDEX                                           uint32  = 0x19f
	LB_SETITEMHEIGHT                                           uint32  = 0x1a0
	LB_GETITEMHEIGHT                                           uint32  = 0x1a1
	LB_FINDSTRINGEXACT                                         uint32  = 0x1a2
	LB_SETLOCALE                                               uint32  = 0x1a5
	LB_GETLOCALE                                               uint32  = 0x1a6
	LB_SETCOUNT                                                uint32  = 0x1a7
	LB_INITSTORAGE                                             uint32  = 0x1a8
	LB_ITEMFROMPOINT                                           uint32  = 0x1a9
	LB_MULTIPLEADDSTRING                                       uint32  = 0x1b1
	LB_GETLISTBOXINFO                                          uint32  = 0x1b2
	LB_MSGMAX                                                  uint32  = 0x1b3
	LBS_NOTIFY                                                 int32   = 1
	LBS_SORT                                                   int32   = 2
	LBS_NOREDRAW                                               int32   = 4
	LBS_MULTIPLESEL                                            int32   = 8
	LBS_OWNERDRAWFIXED                                         int32   = 16
	LBS_OWNERDRAWVARIABLE                                      int32   = 32
	LBS_HASSTRINGS                                             int32   = 64
	LBS_USETABSTOPS                                            int32   = 128
	LBS_NOINTEGRALHEIGHT                                       int32   = 256
	LBS_MULTICOLUMN                                            int32   = 512
	LBS_WANTKEYBOARDINPUT                                      int32   = 1024
	LBS_EXTENDEDSEL                                            int32   = 2048
	LBS_DISABLENOSCROLL                                        int32   = 4096
	LBS_NODATA                                                 int32   = 8192
	LBS_NOSEL                                                  int32   = 16384
	LBS_COMBOBOX                                               int32   = 32768
	CB_OKAY                                                    uint32  = 0x0
	CB_ERR                                                     int32   = -1
	CB_ERRSPACE                                                int32   = -2
	CBN_ERRSPACE                                               int32   = -1
	CBN_SELCHANGE                                              uint32  = 0x1
	CBN_DBLCLK                                                 uint32  = 0x2
	CBN_SETFOCUS                                               uint32  = 0x3
	CBN_KILLFOCUS                                              uint32  = 0x4
	CBN_EDITCHANGE                                             uint32  = 0x5
	CBN_EDITUPDATE                                             uint32  = 0x6
	CBN_DROPDOWN                                               uint32  = 0x7
	CBN_CLOSEUP                                                uint32  = 0x8
	CBN_SELENDOK                                               uint32  = 0x9
	CBN_SELENDCANCEL                                           uint32  = 0xa
	CBS_SIMPLE                                                 int32   = 1
	CBS_DROPDOWN                                               int32   = 2
	CBS_DROPDOWNLIST                                           int32   = 3
	CBS_OWNERDRAWFIXED                                         int32   = 16
	CBS_OWNERDRAWVARIABLE                                      int32   = 32
	CBS_AUTOHSCROLL                                            int32   = 64
	CBS_OEMCONVERT                                             int32   = 128
	CBS_SORT                                                   int32   = 256
	CBS_HASSTRINGS                                             int32   = 512
	CBS_NOINTEGRALHEIGHT                                       int32   = 1024
	CBS_DISABLENOSCROLL                                        int32   = 2048
	CBS_UPPERCASE                                              int32   = 8192
	CBS_LOWERCASE                                              int32   = 16384
	CB_GETEDITSEL                                              uint32  = 0x140
	CB_LIMITTEXT                                               uint32  = 0x141
	CB_SETEDITSEL                                              uint32  = 0x142
	CB_ADDSTRING                                               uint32  = 0x143
	CB_DELETESTRING                                            uint32  = 0x144
	CB_DIR                                                     uint32  = 0x145
	CB_GETCOUNT                                                uint32  = 0x146
	CB_GETCURSEL                                               uint32  = 0x147
	CB_GETLBTEXT                                               uint32  = 0x148
	CB_GETLBTEXTLEN                                            uint32  = 0x149
	CB_INSERTSTRING                                            uint32  = 0x14a
	CB_RESETCONTENT                                            uint32  = 0x14b
	CB_FINDSTRING                                              uint32  = 0x14c
	CB_SELECTSTRING                                            uint32  = 0x14d
	CB_SETCURSEL                                               uint32  = 0x14e
	CB_SHOWDROPDOWN                                            uint32  = 0x14f
	CB_GETITEMDATA                                             uint32  = 0x150
	CB_SETITEMDATA                                             uint32  = 0x151
	CB_GETDROPPEDCONTROLRECT                                   uint32  = 0x152
	CB_SETITEMHEIGHT                                           uint32  = 0x153
	CB_GETITEMHEIGHT                                           uint32  = 0x154
	CB_SETEXTENDEDUI                                           uint32  = 0x155
	CB_GETEXTENDEDUI                                           uint32  = 0x156
	CB_GETDROPPEDSTATE                                         uint32  = 0x157
	CB_FINDSTRINGEXACT                                         uint32  = 0x158
	CB_SETLOCALE                                               uint32  = 0x159
	CB_GETLOCALE                                               uint32  = 0x15a
	CB_GETTOPINDEX                                             uint32  = 0x15b
	CB_SETTOPINDEX                                             uint32  = 0x15c
	CB_GETHORIZONTALEXTENT                                     uint32  = 0x15d
	CB_SETHORIZONTALEXTENT                                     uint32  = 0x15e
	CB_GETDROPPEDWIDTH                                         uint32  = 0x15f
	CB_SETDROPPEDWIDTH                                         uint32  = 0x160
	CB_INITSTORAGE                                             uint32  = 0x161
	CB_MULTIPLEADDSTRING                                       uint32  = 0x163
	CB_GETCOMBOBOXINFO                                         uint32  = 0x164
	CB_MSGMAX                                                  uint32  = 0x165
	SBS_HORZ                                                   int32   = 0
	SBS_VERT                                                   int32   = 1
	SBS_TOPALIGN                                               int32   = 2
	SBS_LEFTALIGN                                              int32   = 2
	SBS_BOTTOMALIGN                                            int32   = 4
	SBS_RIGHTALIGN                                             int32   = 4
	SBS_SIZEBOXTOPLEFTALIGN                                    int32   = 2
	SBS_SIZEBOXBOTTOMRIGHTALIGN                                int32   = 4
	SBS_SIZEBOX                                                int32   = 8
	SBS_SIZEGRIP                                               int32   = 16
	SBM_SETPOS                                                 uint32  = 0xe0
	SBM_GETPOS                                                 uint32  = 0xe1
	SBM_SETRANGE                                               uint32  = 0xe2
	SBM_SETRANGEREDRAW                                         uint32  = 0xe6
	SBM_GETRANGE                                               uint32  = 0xe3
	SBM_ENABLE_ARROWS                                          uint32  = 0xe4
	SBM_SETSCROLLINFO                                          uint32  = 0xe9
	SBM_GETSCROLLINFO                                          uint32  = 0xea
	SBM_GETSCROLLBARINFO                                       uint32  = 0xeb
	MDIS_ALLCHILDSTYLES                                        uint32  = 0x1
	HELP_CONTEXT                                               int32   = 1
	HELP_QUIT                                                  int32   = 2
	HELP_INDEX                                                 int32   = 3
	HELP_CONTENTS                                              int32   = 3
	HELP_HELPONHELP                                            int32   = 4
	HELP_SETINDEX                                              int32   = 5
	HELP_SETCONTENTS                                           int32   = 5
	HELP_CONTEXTPOPUP                                          int32   = 8
	HELP_FORCEFILE                                             int32   = 9
	HELP_KEY                                                   int32   = 257
	HELP_COMMAND                                               int32   = 258
	HELP_PARTIALKEY                                            int32   = 261
	HELP_MULTIKEY                                              int32   = 513
	HELP_SETWINPOS                                             int32   = 515
	HELP_CONTEXTMENU                                           uint32  = 0xa
	HELP_FINDER                                                uint32  = 0xb
	HELP_WM_HELP                                               uint32  = 0xc
	HELP_SETPOPUP_POS                                          uint32  = 0xd
	HELP_TCARD                                                 uint32  = 0x8000
	HELP_TCARD_DATA                                            uint32  = 0x10
	HELP_TCARD_OTHER_CALLER                                    uint32  = 0x11
	IDH_NO_HELP                                                uint32  = 0x6f18
	IDH_MISSING_CONTEXT                                        uint32  = 0x6f19
	IDH_GENERIC_HELP_BUTTON                                    uint32  = 0x6f1a
	IDH_OK                                                     uint32  = 0x6f1b
	IDH_CANCEL                                                 uint32  = 0x6f1c
	IDH_HELP                                                   uint32  = 0x6f1d
	MAX_TOUCH_PREDICTION_FILTER_TAPS                           uint32  = 0x3
	TOUCHPREDICTIONPARAMETERS_DEFAULT_LATENCY                  uint32  = 0x8
	TOUCHPREDICTIONPARAMETERS_DEFAULT_SAMPLETIME               uint32  = 0x8
	TOUCHPREDICTIONPARAMETERS_DEFAULT_USE_HW_TIMESTAMP         uint32  = 0x1
	TOUCHPREDICTIONPARAMETERS_DEFAULT_RLS_DELTA                float32 = 0.001
	TOUCHPREDICTIONPARAMETERS_DEFAULT_RLS_LAMBDA_MIN           float32 = 0.9
	TOUCHPREDICTIONPARAMETERS_DEFAULT_RLS_LAMBDA_MAX           float32 = 0.999
	TOUCHPREDICTIONPARAMETERS_DEFAULT_RLS_LAMBDA_LEARNING_RATE float32 = 0.001
	TOUCHPREDICTIONPARAMETERS_DEFAULT_RLS_EXPO_SMOOTH_ALPHA    float32 = 0.99
	MAX_LOGICALDPIOVERRIDE                                     uint32  = 0x2
	MIN_LOGICALDPIOVERRIDE                                     int32   = -2
	FE_FONTSMOOTHINGSTANDARD                                   uint32  = 0x1
	FE_FONTSMOOTHINGCLEARTYPE                                  uint32  = 0x2
	FE_FONTSMOOTHINGORIENTATIONBGR                             uint32  = 0x0
	FE_FONTSMOOTHINGORIENTATIONRGB                             uint32  = 0x1
	CONTACTVISUALIZATION_OFF                                   uint32  = 0x0
	CONTACTVISUALIZATION_ON                                    uint32  = 0x1
	CONTACTVISUALIZATION_PRESENTATIONMODE                      uint32  = 0x2
	GESTUREVISUALIZATION_OFF                                   uint32  = 0x0
	GESTUREVISUALIZATION_ON                                    uint32  = 0x1f
	GESTUREVISUALIZATION_TAP                                   uint32  = 0x1
	GESTUREVISUALIZATION_DOUBLETAP                             uint32  = 0x2
	GESTUREVISUALIZATION_PRESSANDTAP                           uint32  = 0x4
	GESTUREVISUALIZATION_PRESSANDHOLD                          uint32  = 0x8
	GESTUREVISUALIZATION_RIGHTTAP                              uint32  = 0x10
	MOUSEWHEEL_ROUTING_FOCUS                                   uint32  = 0x0
	MOUSEWHEEL_ROUTING_HYBRID                                  uint32  = 0x1
	MOUSEWHEEL_ROUTING_MOUSE_POS                               uint32  = 0x2
	PENVISUALIZATION_ON                                        uint32  = 0x23
	PENVISUALIZATION_OFF                                       uint32  = 0x0
	PENVISUALIZATION_TAP                                       uint32  = 0x1
	PENVISUALIZATION_DOUBLETAP                                 uint32  = 0x2
	PENVISUALIZATION_CURSOR                                    uint32  = 0x20
	PENARBITRATIONTYPE_NONE                                    uint32  = 0x0
	PENARBITRATIONTYPE_WIN8                                    uint32  = 0x1
	PENARBITRATIONTYPE_FIS                                     uint32  = 0x2
	PENARBITRATIONTYPE_SPT                                     uint32  = 0x3
	PENARBITRATIONTYPE_MAX                                     uint32  = 0x4
	METRICS_USEDEFAULT                                         int32   = -1
	ARW_STARTMASK                                              int32   = 3
	ARW_STARTRIGHT                                             int32   = 1
	ARW_STARTTOP                                               int32   = 2
	ARW_LEFT                                                   int32   = 0
	ARW_RIGHT                                                  int32   = 0
	ARW_UP                                                     int32   = 4
	ARW_DOWN                                                   int32   = 4
	ARW_HIDE                                                   int32   = 8
	HCF_LOGONDESKTOP                                           uint32  = 0x100
	HCF_DEFAULTDESKTOP                                         uint32  = 0x200
	EDS_RAWMODE                                                uint32  = 0x2
	EDS_ROTATEDMODE                                            uint32  = 0x4
	EDD_GET_DEVICE_INTERFACE_NAME                              uint32  = 0x1
	FKF_FILTERKEYSON                                           uint32  = 0x1
	FKF_AVAILABLE                                              uint32  = 0x2
	FKF_HOTKEYACTIVE                                           uint32  = 0x4
	FKF_CONFIRMHOTKEY                                          uint32  = 0x8
	FKF_HOTKEYSOUND                                            uint32  = 0x10
	FKF_INDICATOR                                              uint32  = 0x20
	FKF_CLICKON                                                uint32  = 0x40
	MKF_MOUSEKEYSON                                            uint32  = 0x1
	MKF_AVAILABLE                                              uint32  = 0x2
	MKF_HOTKEYACTIVE                                           uint32  = 0x4
	MKF_CONFIRMHOTKEY                                          uint32  = 0x8
	MKF_HOTKEYSOUND                                            uint32  = 0x10
	MKF_INDICATOR                                              uint32  = 0x20
	MKF_MODIFIERS                                              uint32  = 0x40
	MKF_REPLACENUMBERS                                         uint32  = 0x80
	MKF_LEFTBUTTONSEL                                          uint32  = 0x10000000
	MKF_RIGHTBUTTONSEL                                         uint32  = 0x20000000
	MKF_LEFTBUTTONDOWN                                         uint32  = 0x1000000
	MKF_RIGHTBUTTONDOWN                                        uint32  = 0x2000000
	MKF_MOUSEMODE                                              uint32  = 0x80000000
	TKF_TOGGLEKEYSON                                           uint32  = 0x1
	TKF_AVAILABLE                                              uint32  = 0x2
	TKF_HOTKEYACTIVE                                           uint32  = 0x4
	TKF_CONFIRMHOTKEY                                          uint32  = 0x8
	TKF_HOTKEYSOUND                                            uint32  = 0x10
	TKF_INDICATOR                                              uint32  = 0x20
	MONITORINFOF_PRIMARY                                       uint32  = 0x1
	WINEVENT_OUTOFCONTEXT                                      uint32  = 0x0
	WINEVENT_SKIPOWNTHREAD                                     uint32  = 0x1
	WINEVENT_SKIPOWNPROCESS                                    uint32  = 0x2
	WINEVENT_INCONTEXT                                         uint32  = 0x4
	CHILDID_SELF                                               uint32  = 0x0
	INDEXID_OBJECT                                             uint32  = 0x0
	INDEXID_CONTAINER                                          uint32  = 0x0
	EVENT_MIN                                                  uint32  = 0x1
	EVENT_MAX                                                  uint32  = 0x7fffffff
	EVENT_SYSTEM_SOUND                                         uint32  = 0x1
	EVENT_SYSTEM_ALERT                                         uint32  = 0x2
	EVENT_SYSTEM_FOREGROUND                                    uint32  = 0x3
	EVENT_SYSTEM_MENUSTART                                     uint32  = 0x4
	EVENT_SYSTEM_MENUEND                                       uint32  = 0x5
	EVENT_SYSTEM_MENUPOPUPSTART                                uint32  = 0x6
	EVENT_SYSTEM_MENUPOPUPEND                                  uint32  = 0x7
	EVENT_SYSTEM_CAPTURESTART                                  uint32  = 0x8
	EVENT_SYSTEM_CAPTUREEND                                    uint32  = 0x9
	EVENT_SYSTEM_MOVESIZESTART                                 uint32  = 0xa
	EVENT_SYSTEM_MOVESIZEEND                                   uint32  = 0xb
	EVENT_SYSTEM_CONTEXTHELPSTART                              uint32  = 0xc
	EVENT_SYSTEM_CONTEXTHELPEND                                uint32  = 0xd
	EVENT_SYSTEM_DRAGDROPSTART                                 uint32  = 0xe
	EVENT_SYSTEM_DRAGDROPEND                                   uint32  = 0xf
	EVENT_SYSTEM_DIALOGSTART                                   uint32  = 0x10
	EVENT_SYSTEM_DIALOGEND                                     uint32  = 0x11
	EVENT_SYSTEM_SCROLLINGSTART                                uint32  = 0x12
	EVENT_SYSTEM_SCROLLINGEND                                  uint32  = 0x13
	EVENT_SYSTEM_SWITCHSTART                                   uint32  = 0x14
	EVENT_SYSTEM_SWITCHEND                                     uint32  = 0x15
	EVENT_SYSTEM_MINIMIZESTART                                 uint32  = 0x16
	EVENT_SYSTEM_MINIMIZEEND                                   uint32  = 0x17
	EVENT_SYSTEM_DESKTOPSWITCH                                 uint32  = 0x20
	EVENT_SYSTEM_SWITCHER_APPGRABBED                           uint32  = 0x24
	EVENT_SYSTEM_SWITCHER_APPOVERTARGET                        uint32  = 0x25
	EVENT_SYSTEM_SWITCHER_APPDROPPED                           uint32  = 0x26
	EVENT_SYSTEM_SWITCHER_CANCELLED                            uint32  = 0x27
	EVENT_SYSTEM_IME_KEY_NOTIFICATION                          uint32  = 0x29
	EVENT_SYSTEM_END                                           uint32  = 0xff
	EVENT_OEM_DEFINED_START                                    uint32  = 0x101
	EVENT_OEM_DEFINED_END                                      uint32  = 0x1ff
	EVENT_UIA_EVENTID_START                                    uint32  = 0x4e00
	EVENT_UIA_EVENTID_END                                      uint32  = 0x4eff
	EVENT_UIA_PROPID_START                                     uint32  = 0x7500
	EVENT_UIA_PROPID_END                                       uint32  = 0x75ff
	EVENT_CONSOLE_CARET                                        uint32  = 0x4001
	EVENT_CONSOLE_UPDATE_REGION                                uint32  = 0x4002
	EVENT_CONSOLE_UPDATE_SIMPLE                                uint32  = 0x4003
	EVENT_CONSOLE_UPDATE_SCROLL                                uint32  = 0x4004
	EVENT_CONSOLE_LAYOUT                                       uint32  = 0x4005
	EVENT_CONSOLE_START_APPLICATION                            uint32  = 0x4006
	EVENT_CONSOLE_END_APPLICATION                              uint32  = 0x4007
	CONSOLE_APPLICATION_16BIT                                  uint32  = 0x0
	CONSOLE_CARET_SELECTION                                    uint32  = 0x1
	CONSOLE_CARET_VISIBLE                                      uint32  = 0x2
	EVENT_CONSOLE_END                                          uint32  = 0x40ff
	EVENT_OBJECT_CREATE                                        uint32  = 0x8000
	EVENT_OBJECT_DESTROY                                       uint32  = 0x8001
	EVENT_OBJECT_SHOW                                          uint32  = 0x8002
	EVENT_OBJECT_HIDE                                          uint32  = 0x8003
	EVENT_OBJECT_REORDER                                       uint32  = 0x8004
	EVENT_OBJECT_FOCUS                                         uint32  = 0x8005
	EVENT_OBJECT_SELECTION                                     uint32  = 0x8006
	EVENT_OBJECT_SELECTIONADD                                  uint32  = 0x8007
	EVENT_OBJECT_SELECTIONREMOVE                               uint32  = 0x8008
	EVENT_OBJECT_SELECTIONWITHIN                               uint32  = 0x8009
	EVENT_OBJECT_STATECHANGE                                   uint32  = 0x800a
	EVENT_OBJECT_LOCATIONCHANGE                                uint32  = 0x800b
	EVENT_OBJECT_NAMECHANGE                                    uint32  = 0x800c
	EVENT_OBJECT_DESCRIPTIONCHANGE                             uint32  = 0x800d
	EVENT_OBJECT_VALUECHANGE                                   uint32  = 0x800e
	EVENT_OBJECT_PARENTCHANGE                                  uint32  = 0x800f
	EVENT_OBJECT_HELPCHANGE                                    uint32  = 0x8010
	EVENT_OBJECT_DEFACTIONCHANGE                               uint32  = 0x8011
	EVENT_OBJECT_ACCELERATORCHANGE                             uint32  = 0x8012
	EVENT_OBJECT_INVOKED                                       uint32  = 0x8013
	EVENT_OBJECT_TEXTSELECTIONCHANGED                          uint32  = 0x8014
	EVENT_OBJECT_CONTENTSCROLLED                               uint32  = 0x8015
	EVENT_SYSTEM_ARRANGMENTPREVIEW                             uint32  = 0x8016
	EVENT_OBJECT_CLOAKED                                       uint32  = 0x8017
	EVENT_OBJECT_UNCLOAKED                                     uint32  = 0x8018
	EVENT_OBJECT_LIVEREGIONCHANGED                             uint32  = 0x8019
	EVENT_OBJECT_HOSTEDOBJECTSINVALIDATED                      uint32  = 0x8020
	EVENT_OBJECT_DRAGSTART                                     uint32  = 0x8021
	EVENT_OBJECT_DRAGCANCEL                                    uint32  = 0x8022
	EVENT_OBJECT_DRAGCOMPLETE                                  uint32  = 0x8023
	EVENT_OBJECT_DRAGENTER                                     uint32  = 0x8024
	EVENT_OBJECT_DRAGLEAVE                                     uint32  = 0x8025
	EVENT_OBJECT_DRAGDROPPED                                   uint32  = 0x8026
	EVENT_OBJECT_IME_SHOW                                      uint32  = 0x8027
	EVENT_OBJECT_IME_HIDE                                      uint32  = 0x8028
	EVENT_OBJECT_IME_CHANGE                                    uint32  = 0x8029
	EVENT_OBJECT_TEXTEDIT_CONVERSIONTARGETCHANGED              uint32  = 0x8030
	EVENT_OBJECT_END                                           uint32  = 0x80ff
	EVENT_AIA_START                                            uint32  = 0xa000
	EVENT_AIA_END                                              uint32  = 0xafff
	SOUND_SYSTEM_STARTUP                                       uint32  = 0x1
	SOUND_SYSTEM_SHUTDOWN                                      uint32  = 0x2
	SOUND_SYSTEM_BEEP                                          uint32  = 0x3
	SOUND_SYSTEM_ERROR                                         uint32  = 0x4
	SOUND_SYSTEM_QUESTION                                      uint32  = 0x5
	SOUND_SYSTEM_WARNING                                       uint32  = 0x6
	SOUND_SYSTEM_INFORMATION                                   uint32  = 0x7
	SOUND_SYSTEM_MAXIMIZE                                      uint32  = 0x8
	SOUND_SYSTEM_MINIMIZE                                      uint32  = 0x9
	SOUND_SYSTEM_RESTOREUP                                     uint32  = 0xa
	SOUND_SYSTEM_RESTOREDOWN                                   uint32  = 0xb
	SOUND_SYSTEM_APPSTART                                      uint32  = 0xc
	SOUND_SYSTEM_FAULT                                         uint32  = 0xd
	SOUND_SYSTEM_APPEND                                        uint32  = 0xe
	SOUND_SYSTEM_MENUCOMMAND                                   uint32  = 0xf
	SOUND_SYSTEM_MENUPOPUP                                     uint32  = 0x10
	CSOUND_SYSTEM                                              uint32  = 0x10
	CALERT_SYSTEM                                              uint32  = 0x6
	GUI_16BITTASK                                              uint32  = 0x0
	USER_DEFAULT_SCREEN_DPI                                    uint32  = 0x60
	STATE_SYSTEM_SELECTED                                      uint32  = 0x2
	STATE_SYSTEM_FOCUSED                                       uint32  = 0x4
	STATE_SYSTEM_CHECKED                                       uint32  = 0x10
	STATE_SYSTEM_MIXED                                         uint32  = 0x20
	STATE_SYSTEM_INDETERMINATE                                 uint32  = 0x20
	STATE_SYSTEM_READONLY                                      uint32  = 0x40
	STATE_SYSTEM_HOTTRACKED                                    uint32  = 0x80
	STATE_SYSTEM_DEFAULT                                       uint32  = 0x100
	STATE_SYSTEM_EXPANDED                                      uint32  = 0x200
	STATE_SYSTEM_COLLAPSED                                     uint32  = 0x400
	STATE_SYSTEM_BUSY                                          uint32  = 0x800
	STATE_SYSTEM_FLOATING                                      uint32  = 0x1000
	STATE_SYSTEM_MARQUEED                                      uint32  = 0x2000
	STATE_SYSTEM_ANIMATED                                      uint32  = 0x4000
	STATE_SYSTEM_SIZEABLE                                      uint32  = 0x20000
	STATE_SYSTEM_MOVEABLE                                      uint32  = 0x40000
	STATE_SYSTEM_SELFVOICING                                   uint32  = 0x80000
	STATE_SYSTEM_SELECTABLE                                    uint32  = 0x200000
	STATE_SYSTEM_LINKED                                        uint32  = 0x400000
	STATE_SYSTEM_TRAVERSED                                     uint32  = 0x800000
	STATE_SYSTEM_MULTISELECTABLE                               uint32  = 0x1000000
	STATE_SYSTEM_EXTSELECTABLE                                 uint32  = 0x2000000
	STATE_SYSTEM_ALERT_LOW                                     uint32  = 0x4000000
	STATE_SYSTEM_ALERT_MEDIUM                                  uint32  = 0x8000000
	STATE_SYSTEM_ALERT_HIGH                                    uint32  = 0x10000000
	STATE_SYSTEM_PROTECTED                                     uint32  = 0x20000000
	STATE_SYSTEM_VALID                                         uint32  = 0x3fffffff
	CCHILDREN_TITLEBAR                                         uint32  = 0x5
	CCHILDREN_SCROLLBAR                                        uint32  = 0x5
	RIM_INPUT                                                  uint32  = 0x0
	RIM_INPUTSINK                                              uint32  = 0x1
	RIM_TYPEMAX                                                uint32  = 0x2
	RI_MOUSE_LEFT_BUTTON_DOWN                                  uint32  = 0x1
	RI_MOUSE_LEFT_BUTTON_UP                                    uint32  = 0x2
	RI_MOUSE_RIGHT_BUTTON_DOWN                                 uint32  = 0x4
	RI_MOUSE_RIGHT_BUTTON_UP                                   uint32  = 0x8
	RI_MOUSE_MIDDLE_BUTTON_DOWN                                uint32  = 0x10
	RI_MOUSE_MIDDLE_BUTTON_UP                                  uint32  = 0x20
	RI_MOUSE_BUTTON_1_DOWN                                     uint32  = 0x1
	RI_MOUSE_BUTTON_1_UP                                       uint32  = 0x2
	RI_MOUSE_BUTTON_2_DOWN                                     uint32  = 0x4
	RI_MOUSE_BUTTON_2_UP                                       uint32  = 0x8
	RI_MOUSE_BUTTON_3_DOWN                                     uint32  = 0x10
	RI_MOUSE_BUTTON_3_UP                                       uint32  = 0x20
	RI_MOUSE_BUTTON_4_DOWN                                     uint32  = 0x40
	RI_MOUSE_BUTTON_4_UP                                       uint32  = 0x80
	RI_MOUSE_BUTTON_5_DOWN                                     uint32  = 0x100
	RI_MOUSE_BUTTON_5_UP                                       uint32  = 0x200
	RI_MOUSE_WHEEL                                             uint32  = 0x400
	RI_MOUSE_HWHEEL                                            uint32  = 0x800
	RI_KEY_MAKE                                                uint32  = 0x0
	RI_KEY_BREAK                                               uint32  = 0x1
	RI_KEY_E0                                                  uint32  = 0x2
	RI_KEY_E1                                                  uint32  = 0x4
	RI_KEY_TERMSRV_SET_LED                                     uint32  = 0x8
	RI_KEY_TERMSRV_SHADOW                                      uint32  = 0x10
	RIDEV_EXMODEMASK                                           uint32  = 0xf0
	GIDC_ARRIVAL                                               uint32  = 0x1
	GIDC_REMOVAL                                               uint32  = 0x2
	POINTER_DEVICE_PRODUCT_STRING_MAX                          uint32  = 0x208
	PDC_ARRIVAL                                                uint32  = 0x1
	PDC_REMOVAL                                                uint32  = 0x2
	PDC_ORIENTATION_0                                          uint32  = 0x4
	PDC_ORIENTATION_90                                         uint32  = 0x8
	PDC_ORIENTATION_180                                        uint32  = 0x10
	PDC_ORIENTATION_270                                        uint32  = 0x20
	PDC_MODE_DEFAULT                                           uint32  = 0x40
	PDC_MODE_CENTERED                                          uint32  = 0x80
	PDC_MAPPING_CHANGE                                         uint32  = 0x100
	PDC_RESOLUTION                                             uint32  = 0x200
	PDC_ORIGIN                                                 uint32  = 0x400
	PDC_MODE_ASPECTRATIOPRESERVED                              uint32  = 0x800
	GF_BEGIN                                                   uint32  = 0x1
	GF_INERTIA                                                 uint32  = 0x2
	GF_END                                                     uint32  = 0x4
	GESTURECONFIGMAXCOUNT                                      uint32  = 0x100
	GCF_INCLUDE_ANCESTORS                                      uint32  = 0x1
	NID_INTEGRATED_TOUCH                                       uint32  = 0x1
	NID_EXTERNAL_TOUCH                                         uint32  = 0x2
	NID_INTEGRATED_PEN                                         uint32  = 0x4
	NID_EXTERNAL_PEN                                           uint32  = 0x8
	NID_MULTI_INPUT                                            uint32  = 0x40
	NID_READY                                                  uint32  = 0x80
	MAX_STR_BLOCKREASON                                        uint32  = 0x100
	STRSAFE_USE_SECURE_CRT                                     uint32  = 0x0
	STRSAFE_MAX_CCH                                            uint32  = 0x7fffffff
	STRSAFE_IGNORE_NULLS                                       uint32  = 0x100
	STRSAFE_FILL_BEHIND_NULL                                   uint32  = 0x200
	STRSAFE_FILL_ON_FAILURE                                    uint32  = 0x400
	STRSAFE_NULL_ON_FAILURE                                    uint32  = 0x800
	STRSAFE_NO_TRUNCATION                                      uint32  = 0x1000
	STRSAFE_E_INSUFFICIENT_BUFFER                              HRESULT = -2147024774
	STRSAFE_E_INVALID_PARAMETER                                HRESULT = -2147024809
	STRSAFE_E_END_OF_FILE                                      HRESULT = -2147024858
	WARNING_CYCLOMATIC_COMPLEXITY__                            uint32  = 0x703e
	WARNING_USING_UNINIT_VAR__                                 uint32  = 0x1771
	WARNING_RETURN_UNINIT_VAR__                                uint32  = 0x17d5
	WARNING_DEREF_NULL_PTR__                                   uint32  = 0x177b
	WARNING_MISSING_ZERO_TERMINATION2__                        uint32  = 0x17a6
	WARNING_INVALID_PARAM_VALUE_1__                            uint32  = 0x18f3
	WARNING_INCORRECT_ANNOTATION__                             uint32  = 0x6597
	WARNING_POTENTIAL_BUFFER_OVERFLOW_HIGH_PRIORITY__          uint32  = 0x659f
	WARNING_PRECONDITION_NULLTERMINATION_VIOLATION__           uint32  = 0x65b3
	WARNING_POSTCONDITION_NULLTERMINATION_VIOLATION__          uint32  = 0x65b4
	WARNING_HIGH_PRIORITY_OVERFLOW_POSTCONDITION__             uint32  = 0x65bd
	WARNING_RANGE_POSTCONDITION_VIOLATION__                    uint32  = 0x65cd
	WARNING_POTENTIAL_RANGE_POSTCONDITION_VIOLATION__          uint32  = 0x65d7
	WARNING_INVALID_PARAM_VALUE_3__                            uint32  = 0x6e17
	WARNING_RETURNING_BAD_RESULT__                             uint32  = 0x6e24
	WARNING_BANNED_API_USAGE__                                 uint32  = 0x702f
	WARNING_POST_EXPECTED__                                    uint32  = 0x6e32
	HBMMENU_CALLBACK                                           HBITMAP = ^HBITMAP(0x0)
	HBMMENU_SYSTEM                                             HBITMAP = 1
	HBMMENU_MBAR_RESTORE                                       HBITMAP = 2
	HBMMENU_MBAR_MINIMIZE                                      HBITMAP = 3
	HBMMENU_MBAR_CLOSE                                         HBITMAP = 5
	HBMMENU_MBAR_CLOSE_D                                       HBITMAP = 6
	HBMMENU_MBAR_MINIMIZE_D                                    HBITMAP = 7
	HBMMENU_POPUP_CLOSE                                        HBITMAP = 8
	HBMMENU_POPUP_RESTORE                                      HBITMAP = 9
	HBMMENU_POPUP_MAXIMIZE                                     HBITMAP = 10
	HBMMENU_POPUP_MINIMIZE                                     HBITMAP = 11
	CW_USEDEFAULT                                              int32   = -2147483648
	LBS_STANDARD                                               int32   = 10485763
)

var (
	RT_CURSOR       = PWSTR(unsafe.Pointer(uintptr(1)))
	RT_BITMAP       = PWSTR(unsafe.Pointer(uintptr(2)))
	RT_ICON         = PWSTR(unsafe.Pointer(uintptr(3)))
	RT_MENU         = PWSTR(unsafe.Pointer(uintptr(4)))
	RT_DIALOG       = PWSTR(unsafe.Pointer(uintptr(5)))
	RT_FONTDIR      = PWSTR(unsafe.Pointer(uintptr(7)))
	RT_FONT         = PWSTR(unsafe.Pointer(uintptr(8)))
	RT_ACCELERATOR  = PWSTR(unsafe.Pointer(uintptr(9)))
	RT_MESSAGETABLE = PWSTR(unsafe.Pointer(uintptr(11)))
	RT_VERSION      = PWSTR(unsafe.Pointer(uintptr(16)))
	RT_DLGINCLUDE   = PWSTR(unsafe.Pointer(uintptr(17)))
	RT_PLUGPLAY     = PWSTR(unsafe.Pointer(uintptr(19)))
	RT_VXD          = PWSTR(unsafe.Pointer(uintptr(20)))
	RT_ANICURSOR    = PWSTR(unsafe.Pointer(uintptr(21)))
	RT_ANIICON      = PWSTR(unsafe.Pointer(uintptr(22)))
	RT_HTML         = PWSTR(unsafe.Pointer(uintptr(23)))
	IDC_ARROW       = PWSTR(unsafe.Pointer(uintptr(32512)))
	IDC_IBEAM       = PWSTR(unsafe.Pointer(uintptr(32513)))
	IDC_WAIT        = PWSTR(unsafe.Pointer(uintptr(32514)))
	IDC_CROSS       = PWSTR(unsafe.Pointer(uintptr(32515)))
	IDC_UPARROW     = PWSTR(unsafe.Pointer(uintptr(32516)))
	IDC_SIZE        = PWSTR(unsafe.Pointer(uintptr(32640)))
	IDC_ICON        = PWSTR(unsafe.Pointer(uintptr(32641)))
	IDC_SIZENWSE    = PWSTR(unsafe.Pointer(uintptr(32642)))
	IDC_SIZENESW    = PWSTR(unsafe.Pointer(uintptr(32643)))
	IDC_SIZEWE      = PWSTR(unsafe.Pointer(uintptr(32644)))
	IDC_SIZENS      = PWSTR(unsafe.Pointer(uintptr(32645)))
	IDC_SIZEALL     = PWSTR(unsafe.Pointer(uintptr(32646)))
	IDC_NO          = PWSTR(unsafe.Pointer(uintptr(32648)))
	IDC_HAND        = PWSTR(unsafe.Pointer(uintptr(32649)))
	IDC_APPSTARTING = PWSTR(unsafe.Pointer(uintptr(32650)))
	IDC_HELP        = PWSTR(unsafe.Pointer(uintptr(32651)))
	IDC_PIN         = PWSTR(unsafe.Pointer(uintptr(32671)))
	IDC_PERSON      = PWSTR(unsafe.Pointer(uintptr(32672)))
	IDI_APPLICATION = PWSTR(unsafe.Pointer(uintptr(0x7f00)))
	IDI_HAND        = PWSTR(unsafe.Pointer(uintptr(0x7f01)))
	IDI_QUESTION    = PWSTR(unsafe.Pointer(uintptr(0x7f02)))
	IDI_EXCLAMATION = PWSTR(unsafe.Pointer(uintptr(0x7f03)))
	IDI_ASTERISK    = PWSTR(unsafe.Pointer(uintptr(0x7f04)))
	IDI_WINLOGO     = PWSTR(unsafe.Pointer(uintptr(0x7f05)))
	IDI_SHIELD      = PWSTR(unsafe.Pointer(uintptr(0x7f06)))
)

// enums

// enum
// flags
type WNDCLASS_STYLES uint32

const (
	CS_VREDRAW         WNDCLASS_STYLES = 1
	CS_HREDRAW         WNDCLASS_STYLES = 2
	CS_DBLCLKS         WNDCLASS_STYLES = 8
	CS_OWNDC           WNDCLASS_STYLES = 32
	CS_CLASSDC         WNDCLASS_STYLES = 64
	CS_PARENTDC        WNDCLASS_STYLES = 128
	CS_NOCLOSE         WNDCLASS_STYLES = 512
	CS_SAVEBITS        WNDCLASS_STYLES = 2048
	CS_BYTEALIGNCLIENT WNDCLASS_STYLES = 4096
	CS_BYTEALIGNWINDOW WNDCLASS_STYLES = 8192
	CS_GLOBALCLASS     WNDCLASS_STYLES = 16384
	CS_IME             WNDCLASS_STYLES = 65536
	CS_DROPSHADOW      WNDCLASS_STYLES = 131072
)

// enum
// flags
type CWP_FLAGS uint32

const (
	CWP_ALL             CWP_FLAGS = 0
	CWP_SKIPINVISIBLE   CWP_FLAGS = 1
	CWP_SKIPDISABLED    CWP_FLAGS = 2
	CWP_SKIPTRANSPARENT CWP_FLAGS = 4
)

// enum
// flags
type MESSAGEBOX_STYLE uint32

const (
	MB_ABORTRETRYIGNORE          MESSAGEBOX_STYLE = 2
	MB_CANCELTRYCONTINUE         MESSAGEBOX_STYLE = 6
	MB_HELP                      MESSAGEBOX_STYLE = 16384
	MB_OK                        MESSAGEBOX_STYLE = 0
	MB_OKCANCEL                  MESSAGEBOX_STYLE = 1
	MB_RETRYCANCEL               MESSAGEBOX_STYLE = 5
	MB_YESNO                     MESSAGEBOX_STYLE = 4
	MB_YESNOCANCEL               MESSAGEBOX_STYLE = 3
	MB_ICONHAND                  MESSAGEBOX_STYLE = 16
	MB_ICONQUESTION              MESSAGEBOX_STYLE = 32
	MB_ICONEXCLAMATION           MESSAGEBOX_STYLE = 48
	MB_ICONASTERISK              MESSAGEBOX_STYLE = 64
	MB_USERICON                  MESSAGEBOX_STYLE = 128
	MB_ICONWARNING               MESSAGEBOX_STYLE = 48
	MB_ICONERROR                 MESSAGEBOX_STYLE = 16
	MB_ICONINFORMATION           MESSAGEBOX_STYLE = 64
	MB_ICONSTOP                  MESSAGEBOX_STYLE = 16
	MB_DEFBUTTON1                MESSAGEBOX_STYLE = 0
	MB_DEFBUTTON2                MESSAGEBOX_STYLE = 256
	MB_DEFBUTTON3                MESSAGEBOX_STYLE = 512
	MB_DEFBUTTON4                MESSAGEBOX_STYLE = 768
	MB_APPLMODAL                 MESSAGEBOX_STYLE = 0
	MB_SYSTEMMODAL               MESSAGEBOX_STYLE = 4096
	MB_TASKMODAL                 MESSAGEBOX_STYLE = 8192
	MB_NOFOCUS                   MESSAGEBOX_STYLE = 32768
	MB_SETFOREGROUND             MESSAGEBOX_STYLE = 65536
	MB_DEFAULT_DESKTOP_ONLY      MESSAGEBOX_STYLE = 131072
	MB_TOPMOST                   MESSAGEBOX_STYLE = 262144
	MB_RIGHT                     MESSAGEBOX_STYLE = 524288
	MB_RTLREADING                MESSAGEBOX_STYLE = 1048576
	MB_SERVICE_NOTIFICATION      MESSAGEBOX_STYLE = 2097152
	MB_SERVICE_NOTIFICATION_NT3X MESSAGEBOX_STYLE = 262144
	MB_TYPEMASK                  MESSAGEBOX_STYLE = 15
	MB_ICONMASK                  MESSAGEBOX_STYLE = 240
	MB_DEFMASK                   MESSAGEBOX_STYLE = 3840
	MB_MODEMASK                  MESSAGEBOX_STYLE = 12288
	MB_MISCMASK                  MESSAGEBOX_STYLE = 49152
)

// enum
// flags
type MENU_ITEM_FLAGS uint32

const (
	MF_BYCOMMAND       MENU_ITEM_FLAGS = 0
	MF_BYPOSITION      MENU_ITEM_FLAGS = 1024
	MF_BITMAP          MENU_ITEM_FLAGS = 4
	MF_CHECKED         MENU_ITEM_FLAGS = 8
	MF_DISABLED        MENU_ITEM_FLAGS = 2
	MF_ENABLED         MENU_ITEM_FLAGS = 0
	MF_GRAYED          MENU_ITEM_FLAGS = 1
	MF_MENUBARBREAK    MENU_ITEM_FLAGS = 32
	MF_MENUBREAK       MENU_ITEM_FLAGS = 64
	MF_OWNERDRAW       MENU_ITEM_FLAGS = 256
	MF_POPUP           MENU_ITEM_FLAGS = 16
	MF_SEPARATOR       MENU_ITEM_FLAGS = 2048
	MF_STRING          MENU_ITEM_FLAGS = 0
	MF_UNCHECKED       MENU_ITEM_FLAGS = 0
	MF_INSERT          MENU_ITEM_FLAGS = 0
	MF_CHANGE          MENU_ITEM_FLAGS = 128
	MF_APPEND          MENU_ITEM_FLAGS = 256
	MF_DELETE          MENU_ITEM_FLAGS = 512
	MF_REMOVE          MENU_ITEM_FLAGS = 4096
	MF_USECHECKBITMAPS MENU_ITEM_FLAGS = 512
	MF_UNHILITE        MENU_ITEM_FLAGS = 0
	MF_HILITE          MENU_ITEM_FLAGS = 128
	MF_DEFAULT         MENU_ITEM_FLAGS = 4096
	MF_SYSMENU         MENU_ITEM_FLAGS = 8192
	MF_HELP            MENU_ITEM_FLAGS = 16384
	MF_RIGHTJUSTIFY    MENU_ITEM_FLAGS = 16384
	MF_MOUSESELECT     MENU_ITEM_FLAGS = 32768
	MF_END             MENU_ITEM_FLAGS = 128
)

// enum
// flags
type SHOW_WINDOW_CMD uint32

const (
	SW_FORCEMINIMIZE   SHOW_WINDOW_CMD = 11
	SW_HIDE            SHOW_WINDOW_CMD = 0
	SW_MAXIMIZE        SHOW_WINDOW_CMD = 3
	SW_MINIMIZE        SHOW_WINDOW_CMD = 6
	SW_RESTORE         SHOW_WINDOW_CMD = 9
	SW_SHOW            SHOW_WINDOW_CMD = 5
	SW_SHOWDEFAULT     SHOW_WINDOW_CMD = 10
	SW_SHOWMAXIMIZED   SHOW_WINDOW_CMD = 3
	SW_SHOWMINIMIZED   SHOW_WINDOW_CMD = 2
	SW_SHOWMINNOACTIVE SHOW_WINDOW_CMD = 7
	SW_SHOWNA          SHOW_WINDOW_CMD = 8
	SW_SHOWNOACTIVATE  SHOW_WINDOW_CMD = 4
	SW_SHOWNORMAL      SHOW_WINDOW_CMD = 1
	SW_NORMAL          SHOW_WINDOW_CMD = 1
	SW_MAX             SHOW_WINDOW_CMD = 11
	SW_PARENTCLOSING   SHOW_WINDOW_CMD = 1
	SW_OTHERZOOM       SHOW_WINDOW_CMD = 2
	SW_PARENTOPENING   SHOW_WINDOW_CMD = 3
	SW_OTHERUNZOOM     SHOW_WINDOW_CMD = 4
	SW_SCROLLCHILDREN  SHOW_WINDOW_CMD = 1
	SW_INVALIDATE      SHOW_WINDOW_CMD = 2
	SW_ERASE           SHOW_WINDOW_CMD = 4
	SW_SMOOTHSCROLL    SHOW_WINDOW_CMD = 16
)

// enum
// flags
type SYSTEM_PARAMETERS_INFO_ACTION uint32

const (
	SPI_GETBEEP                      SYSTEM_PARAMETERS_INFO_ACTION = 1
	SPI_SETBEEP                      SYSTEM_PARAMETERS_INFO_ACTION = 2
	SPI_GETMOUSE                     SYSTEM_PARAMETERS_INFO_ACTION = 3
	SPI_SETMOUSE                     SYSTEM_PARAMETERS_INFO_ACTION = 4
	SPI_GETBORDER                    SYSTEM_PARAMETERS_INFO_ACTION = 5
	SPI_SETBORDER                    SYSTEM_PARAMETERS_INFO_ACTION = 6
	SPI_GETKEYBOARDSPEED             SYSTEM_PARAMETERS_INFO_ACTION = 10
	SPI_SETKEYBOARDSPEED             SYSTEM_PARAMETERS_INFO_ACTION = 11
	SPI_LANGDRIVER                   SYSTEM_PARAMETERS_INFO_ACTION = 12
	SPI_ICONHORIZONTALSPACING        SYSTEM_PARAMETERS_INFO_ACTION = 13
	SPI_GETSCREENSAVETIMEOUT         SYSTEM_PARAMETERS_INFO_ACTION = 14
	SPI_SETSCREENSAVETIMEOUT         SYSTEM_PARAMETERS_INFO_ACTION = 15
	SPI_GETSCREENSAVEACTIVE          SYSTEM_PARAMETERS_INFO_ACTION = 16
	SPI_SETSCREENSAVEACTIVE          SYSTEM_PARAMETERS_INFO_ACTION = 17
	SPI_GETGRIDGRANULARITY           SYSTEM_PARAMETERS_INFO_ACTION = 18
	SPI_SETGRIDGRANULARITY           SYSTEM_PARAMETERS_INFO_ACTION = 19
	SPI_SETDESKWALLPAPER             SYSTEM_PARAMETERS_INFO_ACTION = 20
	SPI_SETDESKPATTERN               SYSTEM_PARAMETERS_INFO_ACTION = 21
	SPI_GETKEYBOARDDELAY             SYSTEM_PARAMETERS_INFO_ACTION = 22
	SPI_SETKEYBOARDDELAY             SYSTEM_PARAMETERS_INFO_ACTION = 23
	SPI_ICONVERTICALSPACING          SYSTEM_PARAMETERS_INFO_ACTION = 24
	SPI_GETICONTITLEWRAP             SYSTEM_PARAMETERS_INFO_ACTION = 25
	SPI_SETICONTITLEWRAP             SYSTEM_PARAMETERS_INFO_ACTION = 26
	SPI_GETMENUDROPALIGNMENT         SYSTEM_PARAMETERS_INFO_ACTION = 27
	SPI_SETMENUDROPALIGNMENT         SYSTEM_PARAMETERS_INFO_ACTION = 28
	SPI_SETDOUBLECLKWIDTH            SYSTEM_PARAMETERS_INFO_ACTION = 29
	SPI_SETDOUBLECLKHEIGHT           SYSTEM_PARAMETERS_INFO_ACTION = 30
	SPI_GETICONTITLELOGFONT          SYSTEM_PARAMETERS_INFO_ACTION = 31
	SPI_SETDOUBLECLICKTIME           SYSTEM_PARAMETERS_INFO_ACTION = 32
	SPI_SETMOUSEBUTTONSWAP           SYSTEM_PARAMETERS_INFO_ACTION = 33
	SPI_SETICONTITLELOGFONT          SYSTEM_PARAMETERS_INFO_ACTION = 34
	SPI_GETFASTTASKSWITCH            SYSTEM_PARAMETERS_INFO_ACTION = 35
	SPI_SETFASTTASKSWITCH            SYSTEM_PARAMETERS_INFO_ACTION = 36
	SPI_SETDRAGFULLWINDOWS           SYSTEM_PARAMETERS_INFO_ACTION = 37
	SPI_GETDRAGFULLWINDOWS           SYSTEM_PARAMETERS_INFO_ACTION = 38
	SPI_GETNONCLIENTMETRICS          SYSTEM_PARAMETERS_INFO_ACTION = 41
	SPI_SETNONCLIENTMETRICS          SYSTEM_PARAMETERS_INFO_ACTION = 42
	SPI_GETMINIMIZEDMETRICS          SYSTEM_PARAMETERS_INFO_ACTION = 43
	SPI_SETMINIMIZEDMETRICS          SYSTEM_PARAMETERS_INFO_ACTION = 44
	SPI_GETICONMETRICS               SYSTEM_PARAMETERS_INFO_ACTION = 45
	SPI_SETICONMETRICS               SYSTEM_PARAMETERS_INFO_ACTION = 46
	SPI_SETWORKAREA                  SYSTEM_PARAMETERS_INFO_ACTION = 47
	SPI_GETWORKAREA                  SYSTEM_PARAMETERS_INFO_ACTION = 48
	SPI_SETPENWINDOWS                SYSTEM_PARAMETERS_INFO_ACTION = 49
	SPI_GETHIGHCONTRAST              SYSTEM_PARAMETERS_INFO_ACTION = 66
	SPI_SETHIGHCONTRAST              SYSTEM_PARAMETERS_INFO_ACTION = 67
	SPI_GETKEYBOARDPREF              SYSTEM_PARAMETERS_INFO_ACTION = 68
	SPI_SETKEYBOARDPREF              SYSTEM_PARAMETERS_INFO_ACTION = 69
	SPI_GETSCREENREADER              SYSTEM_PARAMETERS_INFO_ACTION = 70
	SPI_SETSCREENREADER              SYSTEM_PARAMETERS_INFO_ACTION = 71
	SPI_GETANIMATION                 SYSTEM_PARAMETERS_INFO_ACTION = 72
	SPI_SETANIMATION                 SYSTEM_PARAMETERS_INFO_ACTION = 73
	SPI_GETFONTSMOOTHING             SYSTEM_PARAMETERS_INFO_ACTION = 74
	SPI_SETFONTSMOOTHING             SYSTEM_PARAMETERS_INFO_ACTION = 75
	SPI_SETDRAGWIDTH                 SYSTEM_PARAMETERS_INFO_ACTION = 76
	SPI_SETDRAGHEIGHT                SYSTEM_PARAMETERS_INFO_ACTION = 77
	SPI_SETHANDHELD                  SYSTEM_PARAMETERS_INFO_ACTION = 78
	SPI_GETLOWPOWERTIMEOUT           SYSTEM_PARAMETERS_INFO_ACTION = 79
	SPI_GETPOWEROFFTIMEOUT           SYSTEM_PARAMETERS_INFO_ACTION = 80
	SPI_SETLOWPOWERTIMEOUT           SYSTEM_PARAMETERS_INFO_ACTION = 81
	SPI_SETPOWEROFFTIMEOUT           SYSTEM_PARAMETERS_INFO_ACTION = 82
	SPI_GETLOWPOWERACTIVE            SYSTEM_PARAMETERS_INFO_ACTION = 83
	SPI_GETPOWEROFFACTIVE            SYSTEM_PARAMETERS_INFO_ACTION = 84
	SPI_SETLOWPOWERACTIVE            SYSTEM_PARAMETERS_INFO_ACTION = 85
	SPI_SETPOWEROFFACTIVE            SYSTEM_PARAMETERS_INFO_ACTION = 86
	SPI_SETCURSORS                   SYSTEM_PARAMETERS_INFO_ACTION = 87
	SPI_SETICONS                     SYSTEM_PARAMETERS_INFO_ACTION = 88
	SPI_GETDEFAULTINPUTLANG          SYSTEM_PARAMETERS_INFO_ACTION = 89
	SPI_SETDEFAULTINPUTLANG          SYSTEM_PARAMETERS_INFO_ACTION = 90
	SPI_SETLANGTOGGLE                SYSTEM_PARAMETERS_INFO_ACTION = 91
	SPI_GETWINDOWSEXTENSION          SYSTEM_PARAMETERS_INFO_ACTION = 92
	SPI_SETMOUSETRAILS               SYSTEM_PARAMETERS_INFO_ACTION = 93
	SPI_GETMOUSETRAILS               SYSTEM_PARAMETERS_INFO_ACTION = 94
	SPI_SETSCREENSAVERRUNNING        SYSTEM_PARAMETERS_INFO_ACTION = 97
	SPI_SCREENSAVERRUNNING           SYSTEM_PARAMETERS_INFO_ACTION = 97
	SPI_GETFILTERKEYS                SYSTEM_PARAMETERS_INFO_ACTION = 50
	SPI_SETFILTERKEYS                SYSTEM_PARAMETERS_INFO_ACTION = 51
	SPI_GETTOGGLEKEYS                SYSTEM_PARAMETERS_INFO_ACTION = 52
	SPI_SETTOGGLEKEYS                SYSTEM_PARAMETERS_INFO_ACTION = 53
	SPI_GETMOUSEKEYS                 SYSTEM_PARAMETERS_INFO_ACTION = 54
	SPI_SETMOUSEKEYS                 SYSTEM_PARAMETERS_INFO_ACTION = 55
	SPI_GETSHOWSOUNDS                SYSTEM_PARAMETERS_INFO_ACTION = 56
	SPI_SETSHOWSOUNDS                SYSTEM_PARAMETERS_INFO_ACTION = 57
	SPI_GETSTICKYKEYS                SYSTEM_PARAMETERS_INFO_ACTION = 58
	SPI_SETSTICKYKEYS                SYSTEM_PARAMETERS_INFO_ACTION = 59
	SPI_GETACCESSTIMEOUT             SYSTEM_PARAMETERS_INFO_ACTION = 60
	SPI_SETACCESSTIMEOUT             SYSTEM_PARAMETERS_INFO_ACTION = 61
	SPI_GETSERIALKEYS                SYSTEM_PARAMETERS_INFO_ACTION = 62
	SPI_SETSERIALKEYS                SYSTEM_PARAMETERS_INFO_ACTION = 63
	SPI_GETSOUNDSENTRY               SYSTEM_PARAMETERS_INFO_ACTION = 64
	SPI_SETSOUNDSENTRY               SYSTEM_PARAMETERS_INFO_ACTION = 65
	SPI_GETSNAPTODEFBUTTON           SYSTEM_PARAMETERS_INFO_ACTION = 95
	SPI_SETSNAPTODEFBUTTON           SYSTEM_PARAMETERS_INFO_ACTION = 96
	SPI_GETMOUSEHOVERWIDTH           SYSTEM_PARAMETERS_INFO_ACTION = 98
	SPI_SETMOUSEHOVERWIDTH           SYSTEM_PARAMETERS_INFO_ACTION = 99
	SPI_GETMOUSEHOVERHEIGHT          SYSTEM_PARAMETERS_INFO_ACTION = 100
	SPI_SETMOUSEHOVERHEIGHT          SYSTEM_PARAMETERS_INFO_ACTION = 101
	SPI_GETMOUSEHOVERTIME            SYSTEM_PARAMETERS_INFO_ACTION = 102
	SPI_SETMOUSEHOVERTIME            SYSTEM_PARAMETERS_INFO_ACTION = 103
	SPI_GETWHEELSCROLLLINES          SYSTEM_PARAMETERS_INFO_ACTION = 104
	SPI_SETWHEELSCROLLLINES          SYSTEM_PARAMETERS_INFO_ACTION = 105
	SPI_GETMENUSHOWDELAY             SYSTEM_PARAMETERS_INFO_ACTION = 106
	SPI_SETMENUSHOWDELAY             SYSTEM_PARAMETERS_INFO_ACTION = 107
	SPI_GETWHEELSCROLLCHARS          SYSTEM_PARAMETERS_INFO_ACTION = 108
	SPI_SETWHEELSCROLLCHARS          SYSTEM_PARAMETERS_INFO_ACTION = 109
	SPI_GETSHOWIMEUI                 SYSTEM_PARAMETERS_INFO_ACTION = 110
	SPI_SETSHOWIMEUI                 SYSTEM_PARAMETERS_INFO_ACTION = 111
	SPI_GETMOUSESPEED                SYSTEM_PARAMETERS_INFO_ACTION = 112
	SPI_SETMOUSESPEED                SYSTEM_PARAMETERS_INFO_ACTION = 113
	SPI_GETSCREENSAVERRUNNING        SYSTEM_PARAMETERS_INFO_ACTION = 114
	SPI_GETDESKWALLPAPER             SYSTEM_PARAMETERS_INFO_ACTION = 115
	SPI_GETAUDIODESCRIPTION          SYSTEM_PARAMETERS_INFO_ACTION = 116
	SPI_SETAUDIODESCRIPTION          SYSTEM_PARAMETERS_INFO_ACTION = 117
	SPI_GETSCREENSAVESECURE          SYSTEM_PARAMETERS_INFO_ACTION = 118
	SPI_SETSCREENSAVESECURE          SYSTEM_PARAMETERS_INFO_ACTION = 119
	SPI_GETHUNGAPPTIMEOUT            SYSTEM_PARAMETERS_INFO_ACTION = 120
	SPI_SETHUNGAPPTIMEOUT            SYSTEM_PARAMETERS_INFO_ACTION = 121
	SPI_GETWAITTOKILLTIMEOUT         SYSTEM_PARAMETERS_INFO_ACTION = 122
	SPI_SETWAITTOKILLTIMEOUT         SYSTEM_PARAMETERS_INFO_ACTION = 123
	SPI_GETWAITTOKILLSERVICETIMEOUT  SYSTEM_PARAMETERS_INFO_ACTION = 124
	SPI_SETWAITTOKILLSERVICETIMEOUT  SYSTEM_PARAMETERS_INFO_ACTION = 125
	SPI_GETMOUSEDOCKTHRESHOLD        SYSTEM_PARAMETERS_INFO_ACTION = 126
	SPI_SETMOUSEDOCKTHRESHOLD        SYSTEM_PARAMETERS_INFO_ACTION = 127
	SPI_GETPENDOCKTHRESHOLD          SYSTEM_PARAMETERS_INFO_ACTION = 128
	SPI_SETPENDOCKTHRESHOLD          SYSTEM_PARAMETERS_INFO_ACTION = 129
	SPI_GETWINARRANGING              SYSTEM_PARAMETERS_INFO_ACTION = 130
	SPI_SETWINARRANGING              SYSTEM_PARAMETERS_INFO_ACTION = 131
	SPI_GETMOUSEDRAGOUTTHRESHOLD     SYSTEM_PARAMETERS_INFO_ACTION = 132
	SPI_SETMOUSEDRAGOUTTHRESHOLD     SYSTEM_PARAMETERS_INFO_ACTION = 133
	SPI_GETPENDRAGOUTTHRESHOLD       SYSTEM_PARAMETERS_INFO_ACTION = 134
	SPI_SETPENDRAGOUTTHRESHOLD       SYSTEM_PARAMETERS_INFO_ACTION = 135
	SPI_GETMOUSESIDEMOVETHRESHOLD    SYSTEM_PARAMETERS_INFO_ACTION = 136
	SPI_SETMOUSESIDEMOVETHRESHOLD    SYSTEM_PARAMETERS_INFO_ACTION = 137
	SPI_GETPENSIDEMOVETHRESHOLD      SYSTEM_PARAMETERS_INFO_ACTION = 138
	SPI_SETPENSIDEMOVETHRESHOLD      SYSTEM_PARAMETERS_INFO_ACTION = 139
	SPI_GETDRAGFROMMAXIMIZE          SYSTEM_PARAMETERS_INFO_ACTION = 140
	SPI_SETDRAGFROMMAXIMIZE          SYSTEM_PARAMETERS_INFO_ACTION = 141
	SPI_GETSNAPSIZING                SYSTEM_PARAMETERS_INFO_ACTION = 142
	SPI_SETSNAPSIZING                SYSTEM_PARAMETERS_INFO_ACTION = 143
	SPI_GETDOCKMOVING                SYSTEM_PARAMETERS_INFO_ACTION = 144
	SPI_SETDOCKMOVING                SYSTEM_PARAMETERS_INFO_ACTION = 145
	SPI_GETTOUCHPREDICTIONPARAMETERS SYSTEM_PARAMETERS_INFO_ACTION = 156
	SPI_SETTOUCHPREDICTIONPARAMETERS SYSTEM_PARAMETERS_INFO_ACTION = 157
	SPI_GETLOGICALDPIOVERRIDE        SYSTEM_PARAMETERS_INFO_ACTION = 158
	SPI_SETLOGICALDPIOVERRIDE        SYSTEM_PARAMETERS_INFO_ACTION = 159
	SPI_GETMENURECT                  SYSTEM_PARAMETERS_INFO_ACTION = 162
	SPI_SETMENURECT                  SYSTEM_PARAMETERS_INFO_ACTION = 163
	SPI_GETACTIVEWINDOWTRACKING      SYSTEM_PARAMETERS_INFO_ACTION = 4096
	SPI_SETACTIVEWINDOWTRACKING      SYSTEM_PARAMETERS_INFO_ACTION = 4097
	SPI_GETMENUANIMATION             SYSTEM_PARAMETERS_INFO_ACTION = 4098
	SPI_SETMENUANIMATION             SYSTEM_PARAMETERS_INFO_ACTION = 4099
	SPI_GETCOMBOBOXANIMATION         SYSTEM_PARAMETERS_INFO_ACTION = 4100
	SPI_SETCOMBOBOXANIMATION         SYSTEM_PARAMETERS_INFO_ACTION = 4101
	SPI_GETLISTBOXSMOOTHSCROLLING    SYSTEM_PARAMETERS_INFO_ACTION = 4102
	SPI_SETLISTBOXSMOOTHSCROLLING    SYSTEM_PARAMETERS_INFO_ACTION = 4103
	SPI_GETGRADIENTCAPTIONS          SYSTEM_PARAMETERS_INFO_ACTION = 4104
	SPI_SETGRADIENTCAPTIONS          SYSTEM_PARAMETERS_INFO_ACTION = 4105
	SPI_GETKEYBOARDCUES              SYSTEM_PARAMETERS_INFO_ACTION = 4106
	SPI_SETKEYBOARDCUES              SYSTEM_PARAMETERS_INFO_ACTION = 4107
	SPI_GETMENUUNDERLINES            SYSTEM_PARAMETERS_INFO_ACTION = 4106
	SPI_SETMENUUNDERLINES            SYSTEM_PARAMETERS_INFO_ACTION = 4107
	SPI_GETACTIVEWNDTRKZORDER        SYSTEM_PARAMETERS_INFO_ACTION = 4108
	SPI_SETACTIVEWNDTRKZORDER        SYSTEM_PARAMETERS_INFO_ACTION = 4109
	SPI_GETHOTTRACKING               SYSTEM_PARAMETERS_INFO_ACTION = 4110
	SPI_SETHOTTRACKING               SYSTEM_PARAMETERS_INFO_ACTION = 4111
	SPI_GETMENUFADE                  SYSTEM_PARAMETERS_INFO_ACTION = 4114
	SPI_SETMENUFADE                  SYSTEM_PARAMETERS_INFO_ACTION = 4115
	SPI_GETSELECTIONFADE             SYSTEM_PARAMETERS_INFO_ACTION = 4116
	SPI_SETSELECTIONFADE             SYSTEM_PARAMETERS_INFO_ACTION = 4117
	SPI_GETTOOLTIPANIMATION          SYSTEM_PARAMETERS_INFO_ACTION = 4118
	SPI_SETTOOLTIPANIMATION          SYSTEM_PARAMETERS_INFO_ACTION = 4119
	SPI_GETTOOLTIPFADE               SYSTEM_PARAMETERS_INFO_ACTION = 4120
	SPI_SETTOOLTIPFADE               SYSTEM_PARAMETERS_INFO_ACTION = 4121
	SPI_GETCURSORSHADOW              SYSTEM_PARAMETERS_INFO_ACTION = 4122
	SPI_SETCURSORSHADOW              SYSTEM_PARAMETERS_INFO_ACTION = 4123
	SPI_GETMOUSESONAR                SYSTEM_PARAMETERS_INFO_ACTION = 4124
	SPI_SETMOUSESONAR                SYSTEM_PARAMETERS_INFO_ACTION = 4125
	SPI_GETMOUSECLICKLOCK            SYSTEM_PARAMETERS_INFO_ACTION = 4126
	SPI_SETMOUSECLICKLOCK            SYSTEM_PARAMETERS_INFO_ACTION = 4127
	SPI_GETMOUSEVANISH               SYSTEM_PARAMETERS_INFO_ACTION = 4128
	SPI_SETMOUSEVANISH               SYSTEM_PARAMETERS_INFO_ACTION = 4129
	SPI_GETFLATMENU                  SYSTEM_PARAMETERS_INFO_ACTION = 4130
	SPI_SETFLATMENU                  SYSTEM_PARAMETERS_INFO_ACTION = 4131
	SPI_GETDROPSHADOW                SYSTEM_PARAMETERS_INFO_ACTION = 4132
	SPI_SETDROPSHADOW                SYSTEM_PARAMETERS_INFO_ACTION = 4133
	SPI_GETBLOCKSENDINPUTRESETS      SYSTEM_PARAMETERS_INFO_ACTION = 4134
	SPI_SETBLOCKSENDINPUTRESETS      SYSTEM_PARAMETERS_INFO_ACTION = 4135
	SPI_GETUIEFFECTS                 SYSTEM_PARAMETERS_INFO_ACTION = 4158
	SPI_SETUIEFFECTS                 SYSTEM_PARAMETERS_INFO_ACTION = 4159
	SPI_GETDISABLEOVERLAPPEDCONTENT  SYSTEM_PARAMETERS_INFO_ACTION = 4160
	SPI_SETDISABLEOVERLAPPEDCONTENT  SYSTEM_PARAMETERS_INFO_ACTION = 4161
	SPI_GETCLIENTAREAANIMATION       SYSTEM_PARAMETERS_INFO_ACTION = 4162
	SPI_SETCLIENTAREAANIMATION       SYSTEM_PARAMETERS_INFO_ACTION = 4163
	SPI_GETCLEARTYPE                 SYSTEM_PARAMETERS_INFO_ACTION = 4168
	SPI_SETCLEARTYPE                 SYSTEM_PARAMETERS_INFO_ACTION = 4169
	SPI_GETSPEECHRECOGNITION         SYSTEM_PARAMETERS_INFO_ACTION = 4170
	SPI_SETSPEECHRECOGNITION         SYSTEM_PARAMETERS_INFO_ACTION = 4171
	SPI_GETCARETBROWSING             SYSTEM_PARAMETERS_INFO_ACTION = 4172
	SPI_SETCARETBROWSING             SYSTEM_PARAMETERS_INFO_ACTION = 4173
	SPI_GETTHREADLOCALINPUTSETTINGS  SYSTEM_PARAMETERS_INFO_ACTION = 4174
	SPI_SETTHREADLOCALINPUTSETTINGS  SYSTEM_PARAMETERS_INFO_ACTION = 4175
	SPI_GETSYSTEMLANGUAGEBAR         SYSTEM_PARAMETERS_INFO_ACTION = 4176
	SPI_SETSYSTEMLANGUAGEBAR         SYSTEM_PARAMETERS_INFO_ACTION = 4177
	SPI_GETFOREGROUNDLOCKTIMEOUT     SYSTEM_PARAMETERS_INFO_ACTION = 8192
	SPI_SETFOREGROUNDLOCKTIMEOUT     SYSTEM_PARAMETERS_INFO_ACTION = 8193
	SPI_GETACTIVEWNDTRKTIMEOUT       SYSTEM_PARAMETERS_INFO_ACTION = 8194
	SPI_SETACTIVEWNDTRKTIMEOUT       SYSTEM_PARAMETERS_INFO_ACTION = 8195
	SPI_GETFOREGROUNDFLASHCOUNT      SYSTEM_PARAMETERS_INFO_ACTION = 8196
	SPI_SETFOREGROUNDFLASHCOUNT      SYSTEM_PARAMETERS_INFO_ACTION = 8197
	SPI_GETCARETWIDTH                SYSTEM_PARAMETERS_INFO_ACTION = 8198
	SPI_SETCARETWIDTH                SYSTEM_PARAMETERS_INFO_ACTION = 8199
	SPI_GETMOUSECLICKLOCKTIME        SYSTEM_PARAMETERS_INFO_ACTION = 8200
	SPI_SETMOUSECLICKLOCKTIME        SYSTEM_PARAMETERS_INFO_ACTION = 8201
	SPI_GETFONTSMOOTHINGTYPE         SYSTEM_PARAMETERS_INFO_ACTION = 8202
	SPI_SETFONTSMOOTHINGTYPE         SYSTEM_PARAMETERS_INFO_ACTION = 8203
	SPI_GETFONTSMOOTHINGCONTRAST     SYSTEM_PARAMETERS_INFO_ACTION = 8204
	SPI_SETFONTSMOOTHINGCONTRAST     SYSTEM_PARAMETERS_INFO_ACTION = 8205
	SPI_GETFOCUSBORDERWIDTH          SYSTEM_PARAMETERS_INFO_ACTION = 8206
	SPI_SETFOCUSBORDERWIDTH          SYSTEM_PARAMETERS_INFO_ACTION = 8207
	SPI_GETFOCUSBORDERHEIGHT         SYSTEM_PARAMETERS_INFO_ACTION = 8208
	SPI_SETFOCUSBORDERHEIGHT         SYSTEM_PARAMETERS_INFO_ACTION = 8209
	SPI_GETFONTSMOOTHINGORIENTATION  SYSTEM_PARAMETERS_INFO_ACTION = 8210
	SPI_SETFONTSMOOTHINGORIENTATION  SYSTEM_PARAMETERS_INFO_ACTION = 8211
	SPI_GETMINIMUMHITRADIUS          SYSTEM_PARAMETERS_INFO_ACTION = 8212
	SPI_SETMINIMUMHITRADIUS          SYSTEM_PARAMETERS_INFO_ACTION = 8213
	SPI_GETMESSAGEDURATION           SYSTEM_PARAMETERS_INFO_ACTION = 8214
	SPI_SETMESSAGEDURATION           SYSTEM_PARAMETERS_INFO_ACTION = 8215
	SPI_GETCONTACTVISUALIZATION      SYSTEM_PARAMETERS_INFO_ACTION = 8216
	SPI_SETCONTACTVISUALIZATION      SYSTEM_PARAMETERS_INFO_ACTION = 8217
	SPI_GETGESTUREVISUALIZATION      SYSTEM_PARAMETERS_INFO_ACTION = 8218
	SPI_SETGESTUREVISUALIZATION      SYSTEM_PARAMETERS_INFO_ACTION = 8219
	SPI_GETMOUSEWHEELROUTING         SYSTEM_PARAMETERS_INFO_ACTION = 8220
	SPI_SETMOUSEWHEELROUTING         SYSTEM_PARAMETERS_INFO_ACTION = 8221
	SPI_GETPENVISUALIZATION          SYSTEM_PARAMETERS_INFO_ACTION = 8222
	SPI_SETPENVISUALIZATION          SYSTEM_PARAMETERS_INFO_ACTION = 8223
	SPI_GETPENARBITRATIONTYPE        SYSTEM_PARAMETERS_INFO_ACTION = 8224
	SPI_SETPENARBITRATIONTYPE        SYSTEM_PARAMETERS_INFO_ACTION = 8225
	SPI_GETCARETTIMEOUT              SYSTEM_PARAMETERS_INFO_ACTION = 8226
	SPI_SETCARETTIMEOUT              SYSTEM_PARAMETERS_INFO_ACTION = 8227
	SPI_GETHANDEDNESS                SYSTEM_PARAMETERS_INFO_ACTION = 8228
	SPI_SETHANDEDNESS                SYSTEM_PARAMETERS_INFO_ACTION = 8229
)

// enum
// flags
type TRACK_POPUP_MENU_FLAGS uint32

const (
	TPM_LEFTBUTTON      TRACK_POPUP_MENU_FLAGS = 0
	TPM_RIGHTBUTTON     TRACK_POPUP_MENU_FLAGS = 2
	TPM_LEFTALIGN       TRACK_POPUP_MENU_FLAGS = 0
	TPM_CENTERALIGN     TRACK_POPUP_MENU_FLAGS = 4
	TPM_RIGHTALIGN      TRACK_POPUP_MENU_FLAGS = 8
	TPM_TOPALIGN        TRACK_POPUP_MENU_FLAGS = 0
	TPM_VCENTERALIGN    TRACK_POPUP_MENU_FLAGS = 16
	TPM_BOTTOMALIGN     TRACK_POPUP_MENU_FLAGS = 32
	TPM_HORIZONTAL      TRACK_POPUP_MENU_FLAGS = 0
	TPM_VERTICAL        TRACK_POPUP_MENU_FLAGS = 64
	TPM_NONOTIFY        TRACK_POPUP_MENU_FLAGS = 128
	TPM_RETURNCMD       TRACK_POPUP_MENU_FLAGS = 256
	TPM_RECURSE         TRACK_POPUP_MENU_FLAGS = 1
	TPM_HORPOSANIMATION TRACK_POPUP_MENU_FLAGS = 1024
	TPM_HORNEGANIMATION TRACK_POPUP_MENU_FLAGS = 2048
	TPM_VERPOSANIMATION TRACK_POPUP_MENU_FLAGS = 4096
	TPM_VERNEGANIMATION TRACK_POPUP_MENU_FLAGS = 8192
	TPM_NOANIMATION     TRACK_POPUP_MENU_FLAGS = 16384
	TPM_LAYOUTRTL       TRACK_POPUP_MENU_FLAGS = 32768
	TPM_WORKAREA        TRACK_POPUP_MENU_FLAGS = 65536
)

// enum
// flags
type WINDOW_EX_STYLE uint32

const (
	WS_EX_DLGMODALFRAME       WINDOW_EX_STYLE = 1
	WS_EX_NOPARENTNOTIFY      WINDOW_EX_STYLE = 4
	WS_EX_TOPMOST             WINDOW_EX_STYLE = 8
	WS_EX_ACCEPTFILES         WINDOW_EX_STYLE = 16
	WS_EX_TRANSPARENT         WINDOW_EX_STYLE = 32
	WS_EX_MDICHILD            WINDOW_EX_STYLE = 64
	WS_EX_TOOLWINDOW          WINDOW_EX_STYLE = 128
	WS_EX_WINDOWEDGE          WINDOW_EX_STYLE = 256
	WS_EX_CLIENTEDGE          WINDOW_EX_STYLE = 512
	WS_EX_CONTEXTHELP         WINDOW_EX_STYLE = 1024
	WS_EX_RIGHT               WINDOW_EX_STYLE = 4096
	WS_EX_LEFT                WINDOW_EX_STYLE = 0
	WS_EX_RTLREADING          WINDOW_EX_STYLE = 8192
	WS_EX_LTRREADING          WINDOW_EX_STYLE = 0
	WS_EX_LEFTSCROLLBAR       WINDOW_EX_STYLE = 16384
	WS_EX_RIGHTSCROLLBAR      WINDOW_EX_STYLE = 0
	WS_EX_CONTROLPARENT       WINDOW_EX_STYLE = 65536
	WS_EX_STATICEDGE          WINDOW_EX_STYLE = 131072
	WS_EX_APPWINDOW           WINDOW_EX_STYLE = 262144
	WS_EX_OVERLAPPEDWINDOW    WINDOW_EX_STYLE = 768
	WS_EX_PALETTEWINDOW       WINDOW_EX_STYLE = 392
	WS_EX_LAYERED             WINDOW_EX_STYLE = 524288
	WS_EX_NOINHERITLAYOUT     WINDOW_EX_STYLE = 1048576
	WS_EX_NOREDIRECTIONBITMAP WINDOW_EX_STYLE = 2097152
	WS_EX_LAYOUTRTL           WINDOW_EX_STYLE = 4194304
	WS_EX_COMPOSITED          WINDOW_EX_STYLE = 33554432
	WS_EX_NOACTIVATE          WINDOW_EX_STYLE = 134217728
)

// enum
// flags
type WINDOW_STYLE uint32

const (
	WS_OVERLAPPED       WINDOW_STYLE = 0
	WS_POPUP            WINDOW_STYLE = 2147483648
	WS_CHILD            WINDOW_STYLE = 1073741824
	WS_MINIMIZE         WINDOW_STYLE = 536870912
	WS_VISIBLE          WINDOW_STYLE = 268435456
	WS_DISABLED         WINDOW_STYLE = 134217728
	WS_CLIPSIBLINGS     WINDOW_STYLE = 67108864
	WS_CLIPCHILDREN     WINDOW_STYLE = 33554432
	WS_MAXIMIZE         WINDOW_STYLE = 16777216
	WS_CAPTION          WINDOW_STYLE = 12582912
	WS_BORDER           WINDOW_STYLE = 8388608
	WS_DLGFRAME         WINDOW_STYLE = 4194304
	WS_VSCROLL          WINDOW_STYLE = 2097152
	WS_HSCROLL          WINDOW_STYLE = 1048576
	WS_SYSMENU          WINDOW_STYLE = 524288
	WS_THICKFRAME       WINDOW_STYLE = 262144
	WS_GROUP            WINDOW_STYLE = 131072
	WS_TABSTOP          WINDOW_STYLE = 65536
	WS_MINIMIZEBOX      WINDOW_STYLE = 131072
	WS_MAXIMIZEBOX      WINDOW_STYLE = 65536
	WS_TILED            WINDOW_STYLE = 0
	WS_ICONIC           WINDOW_STYLE = 536870912
	WS_SIZEBOX          WINDOW_STYLE = 262144
	WS_TILEDWINDOW      WINDOW_STYLE = 13565952
	WS_OVERLAPPEDWINDOW WINDOW_STYLE = 13565952
	WS_POPUPWINDOW      WINDOW_STYLE = 2156396544
	WS_CHILDWINDOW      WINDOW_STYLE = 1073741824
	WS_ACTIVECAPTION    WINDOW_STYLE = 1
)

// enum
type OBJECT_IDENTIFIER int32

const (
	OBJID_WINDOW            OBJECT_IDENTIFIER = 0
	OBJID_SYSMENU           OBJECT_IDENTIFIER = -1
	OBJID_TITLEBAR          OBJECT_IDENTIFIER = -2
	OBJID_MENU              OBJECT_IDENTIFIER = -3
	OBJID_CLIENT            OBJECT_IDENTIFIER = -4
	OBJID_VSCROLL           OBJECT_IDENTIFIER = -5
	OBJID_HSCROLL           OBJECT_IDENTIFIER = -6
	OBJID_SIZEGRIP          OBJECT_IDENTIFIER = -7
	OBJID_CARET             OBJECT_IDENTIFIER = -8
	OBJID_CURSOR            OBJECT_IDENTIFIER = -9
	OBJID_ALERT             OBJECT_IDENTIFIER = -10
	OBJID_SOUND             OBJECT_IDENTIFIER = -11
	OBJID_QUERYCLASSNAMEIDX OBJECT_IDENTIFIER = -12
	OBJID_NATIVEOM          OBJECT_IDENTIFIER = -16
)

// enum
// flags
type MENU_ITEM_TYPE uint32

const (
	MFT_BITMAP       MENU_ITEM_TYPE = 4
	MFT_MENUBARBREAK MENU_ITEM_TYPE = 32
	MFT_MENUBREAK    MENU_ITEM_TYPE = 64
	MFT_OWNERDRAW    MENU_ITEM_TYPE = 256
	MFT_RADIOCHECK   MENU_ITEM_TYPE = 512
	MFT_RIGHTJUSTIFY MENU_ITEM_TYPE = 16384
	MFT_RIGHTORDER   MENU_ITEM_TYPE = 8192
	MFT_SEPARATOR    MENU_ITEM_TYPE = 2048
	MFT_STRING       MENU_ITEM_TYPE = 0
)

// enum
type MESSAGEBOX_RESULT int32

const (
	IDOK       MESSAGEBOX_RESULT = 1
	IDCANCEL   MESSAGEBOX_RESULT = 2
	IDABORT    MESSAGEBOX_RESULT = 3
	IDRETRY    MESSAGEBOX_RESULT = 4
	IDIGNORE   MESSAGEBOX_RESULT = 5
	IDYES      MESSAGEBOX_RESULT = 6
	IDNO       MESSAGEBOX_RESULT = 7
	IDCLOSE    MESSAGEBOX_RESULT = 8
	IDHELP     MESSAGEBOX_RESULT = 9
	IDTRYAGAIN MESSAGEBOX_RESULT = 10
	IDCONTINUE MESSAGEBOX_RESULT = 11
	IDASYNC    MESSAGEBOX_RESULT = 32001
	IDTIMEOUT  MESSAGEBOX_RESULT = 32000
)

// enum
// flags
type MENU_ITEM_STATE uint32

const (
	MFS_GRAYED    MENU_ITEM_STATE = 3
	MFS_DISABLED  MENU_ITEM_STATE = 3
	MFS_CHECKED   MENU_ITEM_STATE = 8
	MFS_HILITE    MENU_ITEM_STATE = 128
	MFS_ENABLED   MENU_ITEM_STATE = 0
	MFS_UNCHECKED MENU_ITEM_STATE = 0
	MFS_UNHILITE  MENU_ITEM_STATE = 0
	MFS_DEFAULT   MENU_ITEM_STATE = 4096
)

// enum
// flags
type SCROLLBAR_CONSTANTS uint32

const (
	SB_CTL  SCROLLBAR_CONSTANTS = 2
	SB_HORZ SCROLLBAR_CONSTANTS = 0
	SB_VERT SCROLLBAR_CONSTANTS = 1
	SB_BOTH SCROLLBAR_CONSTANTS = 3
)

// enum
type GET_CLASS_LONG_INDEX int32

const (
	GCW_ATOM           GET_CLASS_LONG_INDEX = -32
	GCL_CBCLSEXTRA     GET_CLASS_LONG_INDEX = -20
	GCL_CBWNDEXTRA     GET_CLASS_LONG_INDEX = -18
	GCL_HBRBACKGROUND  GET_CLASS_LONG_INDEX = -10
	GCL_HCURSOR        GET_CLASS_LONG_INDEX = -12
	GCL_HICON          GET_CLASS_LONG_INDEX = -14
	GCL_HICONSM        GET_CLASS_LONG_INDEX = -34
	GCL_HMODULE        GET_CLASS_LONG_INDEX = -16
	GCL_MENUNAME       GET_CLASS_LONG_INDEX = -8
	GCL_STYLE          GET_CLASS_LONG_INDEX = -26
	GCL_WNDPROC        GET_CLASS_LONG_INDEX = -24
	GCLP_HBRBACKGROUND GET_CLASS_LONG_INDEX = -10
	GCLP_HCURSOR       GET_CLASS_LONG_INDEX = -12
	GCLP_HICON         GET_CLASS_LONG_INDEX = -14
	GCLP_HICONSM       GET_CLASS_LONG_INDEX = -34
	GCLP_HMODULE       GET_CLASS_LONG_INDEX = -16
	GCLP_MENUNAME      GET_CLASS_LONG_INDEX = -8
	GCLP_WNDPROC       GET_CLASS_LONG_INDEX = -24
)

// enum
type UPDATE_LAYERED_WINDOW_FLAGS uint32

const (
	ULW_ALPHA       UPDATE_LAYERED_WINDOW_FLAGS = 2
	ULW_COLORKEY    UPDATE_LAYERED_WINDOW_FLAGS = 1
	ULW_OPAQUE      UPDATE_LAYERED_WINDOW_FLAGS = 4
	ULW_EX_NORESIZE UPDATE_LAYERED_WINDOW_FLAGS = 8
)

// enum
type WINDOW_LONG_PTR_INDEX int32

const (
	GWL_EXSTYLE     WINDOW_LONG_PTR_INDEX = -20
	GWLP_HINSTANCE  WINDOW_LONG_PTR_INDEX = -6
	GWLP_HWNDPARENT WINDOW_LONG_PTR_INDEX = -8
	GWLP_ID         WINDOW_LONG_PTR_INDEX = -12
	GWL_STYLE       WINDOW_LONG_PTR_INDEX = -16
	GWLP_USERDATA   WINDOW_LONG_PTR_INDEX = -21
	GWLP_WNDPROC    WINDOW_LONG_PTR_INDEX = -4
	GWL_HINSTANCE   WINDOW_LONG_PTR_INDEX = -6
	GWL_ID          WINDOW_LONG_PTR_INDEX = -12
	GWL_USERDATA    WINDOW_LONG_PTR_INDEX = -21
	GWL_WNDPROC     WINDOW_LONG_PTR_INDEX = -4
	GWL_HWNDPARENT  WINDOW_LONG_PTR_INDEX = -8
)

// enum
// flags
type ANIMATE_WINDOW_FLAGS uint32

const (
	AW_ACTIVATE     ANIMATE_WINDOW_FLAGS = 131072
	AW_BLEND        ANIMATE_WINDOW_FLAGS = 524288
	AW_CENTER       ANIMATE_WINDOW_FLAGS = 16
	AW_HIDE         ANIMATE_WINDOW_FLAGS = 65536
	AW_HOR_POSITIVE ANIMATE_WINDOW_FLAGS = 1
	AW_HOR_NEGATIVE ANIMATE_WINDOW_FLAGS = 2
	AW_SLIDE        ANIMATE_WINDOW_FLAGS = 262144
	AW_VER_POSITIVE ANIMATE_WINDOW_FLAGS = 4
	AW_VER_NEGATIVE ANIMATE_WINDOW_FLAGS = 8
)

// enum
type CHANGE_WINDOW_MESSAGE_FILTER_FLAGS uint32

const (
	MSGFLT_ADD    CHANGE_WINDOW_MESSAGE_FILTER_FLAGS = 1
	MSGFLT_REMOVE CHANGE_WINDOW_MESSAGE_FILTER_FLAGS = 2
)

// enum
type GDI_IMAGE_TYPE uint32

const (
	IMAGE_BITMAP GDI_IMAGE_TYPE = 0
	IMAGE_CURSOR GDI_IMAGE_TYPE = 2
	IMAGE_ICON   GDI_IMAGE_TYPE = 1
)

// enum
type WINDOWS_HOOK_ID int32

const (
	WH_CALLWNDPROC     WINDOWS_HOOK_ID = 4
	WH_CALLWNDPROCRET  WINDOWS_HOOK_ID = 12
	WH_CBT             WINDOWS_HOOK_ID = 5
	WH_DEBUG           WINDOWS_HOOK_ID = 9
	WH_FOREGROUNDIDLE  WINDOWS_HOOK_ID = 11
	WH_GETMESSAGE      WINDOWS_HOOK_ID = 3
	WH_JOURNALPLAYBACK WINDOWS_HOOK_ID = 1
	WH_JOURNALRECORD   WINDOWS_HOOK_ID = 0
	WH_KEYBOARD        WINDOWS_HOOK_ID = 2
	WH_KEYBOARD_LL     WINDOWS_HOOK_ID = 13
	WH_MOUSE           WINDOWS_HOOK_ID = 7
	WH_MOUSE_LL        WINDOWS_HOOK_ID = 14
	WH_MSGFILTER       WINDOWS_HOOK_ID = -1
	WH_SHELL           WINDOWS_HOOK_ID = 10
	WH_SYSMSGFILTER    WINDOWS_HOOK_ID = 6
)

// enum
// flags
type IMAGE_FLAGS uint32

const (
	LR_CREATEDIBSECTION IMAGE_FLAGS = 8192
	LR_DEFAULTCOLOR     IMAGE_FLAGS = 0
	LR_DEFAULTSIZE      IMAGE_FLAGS = 64
	LR_LOADFROMFILE     IMAGE_FLAGS = 16
	LR_LOADMAP3DCOLORS  IMAGE_FLAGS = 4096
	LR_LOADTRANSPARENT  IMAGE_FLAGS = 32
	LR_MONOCHROME       IMAGE_FLAGS = 1
	LR_SHARED           IMAGE_FLAGS = 32768
	LR_VGACOLOR         IMAGE_FLAGS = 128
	LR_COPYDELETEORG    IMAGE_FLAGS = 8
	LR_COPYFROMRESOURCE IMAGE_FLAGS = 16384
	LR_COPYRETURNORG    IMAGE_FLAGS = 4
)

// enum
// flags
type SYSTEM_PARAMETERS_INFO_UPDATE_FLAGS uint32

const (
	SPIF_UPDATEINIFILE    SYSTEM_PARAMETERS_INFO_UPDATE_FLAGS = 1
	SPIF_SENDCHANGE       SYSTEM_PARAMETERS_INFO_UPDATE_FLAGS = 2
	SPIF_SENDWININICHANGE SYSTEM_PARAMETERS_INFO_UPDATE_FLAGS = 2
)

// enum
// flags
type SET_WINDOW_POS_FLAGS uint32

const (
	SWP_ASYNCWINDOWPOS SET_WINDOW_POS_FLAGS = 16384
	SWP_DEFERERASE     SET_WINDOW_POS_FLAGS = 8192
	SWP_DRAWFRAME      SET_WINDOW_POS_FLAGS = 32
	SWP_FRAMECHANGED   SET_WINDOW_POS_FLAGS = 32
	SWP_HIDEWINDOW     SET_WINDOW_POS_FLAGS = 128
	SWP_NOACTIVATE     SET_WINDOW_POS_FLAGS = 16
	SWP_NOCOPYBITS     SET_WINDOW_POS_FLAGS = 256
	SWP_NOMOVE         SET_WINDOW_POS_FLAGS = 2
	SWP_NOOWNERZORDER  SET_WINDOW_POS_FLAGS = 512
	SWP_NOREDRAW       SET_WINDOW_POS_FLAGS = 8
	SWP_NOREPOSITION   SET_WINDOW_POS_FLAGS = 512
	SWP_NOSENDCHANGING SET_WINDOW_POS_FLAGS = 1024
	SWP_NOSIZE         SET_WINDOW_POS_FLAGS = 1
	SWP_NOZORDER       SET_WINDOW_POS_FLAGS = 4
	SWP_SHOWWINDOW     SET_WINDOW_POS_FLAGS = 64
	SWP__NOOWNERZORDER SET_WINDOW_POS_FLAGS = 512
)

// enum
// flags
type MSG_WAIT_FOR_MULTIPLE_OBJECTS_EX_FLAGS uint32

const (
	MWMO_NONE           MSG_WAIT_FOR_MULTIPLE_OBJECTS_EX_FLAGS = 0
	MWMO_ALERTABLE      MSG_WAIT_FOR_MULTIPLE_OBJECTS_EX_FLAGS = 2
	MWMO_INPUTAVAILABLE MSG_WAIT_FOR_MULTIPLE_OBJECTS_EX_FLAGS = 4
	MWMO_WAITALL        MSG_WAIT_FOR_MULTIPLE_OBJECTS_EX_FLAGS = 1
)

// enum
// flags
type QUEUE_STATUS_FLAGS uint32

const (
	QS_ALLEVENTS      QUEUE_STATUS_FLAGS = 1215
	QS_ALLINPUT       QUEUE_STATUS_FLAGS = 1279
	QS_ALLPOSTMESSAGE QUEUE_STATUS_FLAGS = 256
	QS_HOTKEY         QUEUE_STATUS_FLAGS = 128
	QS_INPUT          QUEUE_STATUS_FLAGS = 1031
	QS_KEY            QUEUE_STATUS_FLAGS = 1
	QS_MOUSE          QUEUE_STATUS_FLAGS = 6
	QS_MOUSEBUTTON    QUEUE_STATUS_FLAGS = 4
	QS_MOUSEMOVE      QUEUE_STATUS_FLAGS = 2
	QS_PAINT          QUEUE_STATUS_FLAGS = 32
	QS_POSTMESSAGE    QUEUE_STATUS_FLAGS = 8
	QS_RAWINPUT       QUEUE_STATUS_FLAGS = 1024
	QS_SENDMESSAGE    QUEUE_STATUS_FLAGS = 64
	QS_TIMER          QUEUE_STATUS_FLAGS = 16
)

// enum
type SYSTEM_CURSOR_ID uint32

const (
	OCR_APPSTARTING SYSTEM_CURSOR_ID = 32650
	OCR_NORMAL      SYSTEM_CURSOR_ID = 32512
	OCR_CROSS       SYSTEM_CURSOR_ID = 32515
	OCR_HAND        SYSTEM_CURSOR_ID = 32649
	OCR_HELP        SYSTEM_CURSOR_ID = 32651
	OCR_IBEAM       SYSTEM_CURSOR_ID = 32513
	OCR_NO          SYSTEM_CURSOR_ID = 32648
	OCR_SIZEALL     SYSTEM_CURSOR_ID = 32646
	OCR_SIZENESW    SYSTEM_CURSOR_ID = 32643
	OCR_SIZENS      SYSTEM_CURSOR_ID = 32645
	OCR_SIZENWSE    SYSTEM_CURSOR_ID = 32642
	OCR_SIZEWE      SYSTEM_CURSOR_ID = 32644
	OCR_UP          SYSTEM_CURSOR_ID = 32516
	OCR_WAIT        SYSTEM_CURSOR_ID = 32514
)

// enum
// flags
type LAYERED_WINDOW_ATTRIBUTES_FLAGS uint32

const (
	LWA_ALPHA    LAYERED_WINDOW_ATTRIBUTES_FLAGS = 2
	LWA_COLORKEY LAYERED_WINDOW_ATTRIBUTES_FLAGS = 1
)

// enum
// flags
type SEND_MESSAGE_TIMEOUT_FLAGS uint32

const (
	SMTO_ABORTIFHUNG        SEND_MESSAGE_TIMEOUT_FLAGS = 2
	SMTO_BLOCK              SEND_MESSAGE_TIMEOUT_FLAGS = 1
	SMTO_NORMAL             SEND_MESSAGE_TIMEOUT_FLAGS = 0
	SMTO_NOTIMEOUTIFNOTHUNG SEND_MESSAGE_TIMEOUT_FLAGS = 8
	SMTO_ERRORONEXIT        SEND_MESSAGE_TIMEOUT_FLAGS = 32
)

// enum
// flags
type PEEK_MESSAGE_REMOVE_TYPE uint32

const (
	PM_NOREMOVE       PEEK_MESSAGE_REMOVE_TYPE = 0
	PM_REMOVE         PEEK_MESSAGE_REMOVE_TYPE = 1
	PM_NOYIELD        PEEK_MESSAGE_REMOVE_TYPE = 2
	PM_QS_INPUT       PEEK_MESSAGE_REMOVE_TYPE = 67567616
	PM_QS_POSTMESSAGE PEEK_MESSAGE_REMOVE_TYPE = 9961472
	PM_QS_PAINT       PEEK_MESSAGE_REMOVE_TYPE = 2097152
	PM_QS_SENDMESSAGE PEEK_MESSAGE_REMOVE_TYPE = 4194304
)

// enum
type SYS_COLOR_INDEX uint32

const (
	COLOR_3DDKSHADOW              SYS_COLOR_INDEX = 21
	COLOR_3DFACE                  SYS_COLOR_INDEX = 15
	COLOR_3DHIGHLIGHT             SYS_COLOR_INDEX = 20
	COLOR_3DHILIGHT               SYS_COLOR_INDEX = 20
	COLOR_3DLIGHT                 SYS_COLOR_INDEX = 22
	COLOR_3DSHADOW                SYS_COLOR_INDEX = 16
	COLOR_ACTIVEBORDER            SYS_COLOR_INDEX = 10
	COLOR_ACTIVECAPTION           SYS_COLOR_INDEX = 2
	COLOR_APPWORKSPACE            SYS_COLOR_INDEX = 12
	COLOR_BACKGROUND              SYS_COLOR_INDEX = 1
	COLOR_BTNFACE                 SYS_COLOR_INDEX = 15
	COLOR_BTNHIGHLIGHT_           SYS_COLOR_INDEX = 20
	COLOR_BTNHILIGHT_             SYS_COLOR_INDEX = 20
	COLOR_BTNSHADOW               SYS_COLOR_INDEX = 16
	COLOR_BTNTEXT                 SYS_COLOR_INDEX = 18
	COLOR_CAPTIONTEXT             SYS_COLOR_INDEX = 9
	COLOR_DESKTOP                 SYS_COLOR_INDEX = 1
	COLOR_GRADIENTACTIVECAPTION   SYS_COLOR_INDEX = 27
	COLOR_GRADIENTINACTIVECAPTION SYS_COLOR_INDEX = 28
	COLOR_GRAYTEXT                SYS_COLOR_INDEX = 17
	COLOR_HIGHLIGHT               SYS_COLOR_INDEX = 13
	COLOR_HIGHLIGHTTEXT           SYS_COLOR_INDEX = 14
	COLOR_HOTLIGHT                SYS_COLOR_INDEX = 26
	COLOR_INACTIVEBORDER          SYS_COLOR_INDEX = 11
	COLOR_INACTIVECAPTION         SYS_COLOR_INDEX = 3
	COLOR_INACTIVECAPTIONTEXT     SYS_COLOR_INDEX = 19
	COLOR_INFOBK                  SYS_COLOR_INDEX = 24
	COLOR_INFOTEXT                SYS_COLOR_INDEX = 23
	COLOR_MENU                    SYS_COLOR_INDEX = 4
	COLOR_MENUHILIGHT             SYS_COLOR_INDEX = 29
	COLOR_MENUBAR                 SYS_COLOR_INDEX = 30
	COLOR_MENUTEXT                SYS_COLOR_INDEX = 7
	COLOR_SCROLLBAR               SYS_COLOR_INDEX = 0
	COLOR_WINDOW                  SYS_COLOR_INDEX = 5
	COLOR_WINDOWFRAME             SYS_COLOR_INDEX = 6
	COLOR_WINDOWTEXT              SYS_COLOR_INDEX = 8
)

// enum
type GET_WINDOW_CMD uint32

const (
	GW_CHILD        GET_WINDOW_CMD = 5
	GW_ENABLEDPOPUP GET_WINDOW_CMD = 6
	GW_HWNDFIRST    GET_WINDOW_CMD = 0
	GW_HWNDLAST     GET_WINDOW_CMD = 1
	GW_HWNDNEXT     GET_WINDOW_CMD = 2
	GW_HWNDPREV     GET_WINDOW_CMD = 3
	GW_OWNER        GET_WINDOW_CMD = 4
)

// enum
type SYSTEM_METRICS_INDEX uint32

const (
	SM_ARRANGE                     SYSTEM_METRICS_INDEX = 56
	SM_CLEANBOOT                   SYSTEM_METRICS_INDEX = 67
	SM_CMONITORS                   SYSTEM_METRICS_INDEX = 80
	SM_CMOUSEBUTTONS               SYSTEM_METRICS_INDEX = 43
	SM_CONVERTIBLESLATEMODE        SYSTEM_METRICS_INDEX = 8195
	SM_CXBORDER                    SYSTEM_METRICS_INDEX = 5
	SM_CXCURSOR                    SYSTEM_METRICS_INDEX = 13
	SM_CXDLGFRAME                  SYSTEM_METRICS_INDEX = 7
	SM_CXDOUBLECLK                 SYSTEM_METRICS_INDEX = 36
	SM_CXDRAG                      SYSTEM_METRICS_INDEX = 68
	SM_CXEDGE                      SYSTEM_METRICS_INDEX = 45
	SM_CXFIXEDFRAME                SYSTEM_METRICS_INDEX = 7
	SM_CXFOCUSBORDER               SYSTEM_METRICS_INDEX = 83
	SM_CXFRAME                     SYSTEM_METRICS_INDEX = 32
	SM_CXFULLSCREEN                SYSTEM_METRICS_INDEX = 16
	SM_CXHSCROLL                   SYSTEM_METRICS_INDEX = 21
	SM_CXHTHUMB                    SYSTEM_METRICS_INDEX = 10
	SM_CXICON                      SYSTEM_METRICS_INDEX = 11
	SM_CXICONSPACING               SYSTEM_METRICS_INDEX = 38
	SM_CXMAXIMIZED                 SYSTEM_METRICS_INDEX = 61
	SM_CXMAXTRACK                  SYSTEM_METRICS_INDEX = 59
	SM_CXMENUCHECK                 SYSTEM_METRICS_INDEX = 71
	SM_CXMENUSIZE                  SYSTEM_METRICS_INDEX = 54
	SM_CXMIN                       SYSTEM_METRICS_INDEX = 28
	SM_CXMINIMIZED                 SYSTEM_METRICS_INDEX = 57
	SM_CXMINSPACING                SYSTEM_METRICS_INDEX = 47
	SM_CXMINTRACK                  SYSTEM_METRICS_INDEX = 34
	SM_CXPADDEDBORDER              SYSTEM_METRICS_INDEX = 92
	SM_CXSCREEN                    SYSTEM_METRICS_INDEX = 0
	SM_CXSIZE                      SYSTEM_METRICS_INDEX = 30
	SM_CXSIZEFRAME                 SYSTEM_METRICS_INDEX = 32
	SM_CXSMICON                    SYSTEM_METRICS_INDEX = 49
	SM_CXSMSIZE                    SYSTEM_METRICS_INDEX = 52
	SM_CXVIRTUALSCREEN             SYSTEM_METRICS_INDEX = 78
	SM_CXVSCROLL                   SYSTEM_METRICS_INDEX = 2
	SM_CYBORDER                    SYSTEM_METRICS_INDEX = 6
	SM_CYCAPTION                   SYSTEM_METRICS_INDEX = 4
	SM_CYCURSOR                    SYSTEM_METRICS_INDEX = 14
	SM_CYDLGFRAME                  SYSTEM_METRICS_INDEX = 8
	SM_CYDOUBLECLK                 SYSTEM_METRICS_INDEX = 37
	SM_CYDRAG                      SYSTEM_METRICS_INDEX = 69
	SM_CYEDGE                      SYSTEM_METRICS_INDEX = 46
	SM_CYFIXEDFRAME                SYSTEM_METRICS_INDEX = 8
	SM_CYFOCUSBORDER               SYSTEM_METRICS_INDEX = 84
	SM_CYFRAME                     SYSTEM_METRICS_INDEX = 33
	SM_CYFULLSCREEN                SYSTEM_METRICS_INDEX = 17
	SM_CYHSCROLL                   SYSTEM_METRICS_INDEX = 3
	SM_CYICON                      SYSTEM_METRICS_INDEX = 12
	SM_CYICONSPACING               SYSTEM_METRICS_INDEX = 39
	SM_CYKANJIWINDOW               SYSTEM_METRICS_INDEX = 18
	SM_CYMAXIMIZED                 SYSTEM_METRICS_INDEX = 62
	SM_CYMAXTRACK                  SYSTEM_METRICS_INDEX = 60
	SM_CYMENU                      SYSTEM_METRICS_INDEX = 15
	SM_CYMENUCHECK                 SYSTEM_METRICS_INDEX = 72
	SM_CYMENUSIZE                  SYSTEM_METRICS_INDEX = 55
	SM_CYMIN                       SYSTEM_METRICS_INDEX = 29
	SM_CYMINIMIZED                 SYSTEM_METRICS_INDEX = 58
	SM_CYMINSPACING                SYSTEM_METRICS_INDEX = 48
	SM_CYMINTRACK                  SYSTEM_METRICS_INDEX = 35
	SM_CYSCREEN                    SYSTEM_METRICS_INDEX = 1
	SM_CYSIZE                      SYSTEM_METRICS_INDEX = 31
	SM_CYSIZEFRAME                 SYSTEM_METRICS_INDEX = 33
	SM_CYSMCAPTION                 SYSTEM_METRICS_INDEX = 51
	SM_CYSMICON                    SYSTEM_METRICS_INDEX = 50
	SM_CYSMSIZE                    SYSTEM_METRICS_INDEX = 53
	SM_CYVIRTUALSCREEN             SYSTEM_METRICS_INDEX = 79
	SM_CYVSCROLL                   SYSTEM_METRICS_INDEX = 20
	SM_CYVTHUMB                    SYSTEM_METRICS_INDEX = 9
	SM_DBCSENABLED                 SYSTEM_METRICS_INDEX = 42
	SM_DEBUG                       SYSTEM_METRICS_INDEX = 22
	SM_DIGITIZER                   SYSTEM_METRICS_INDEX = 94
	SM_IMMENABLED                  SYSTEM_METRICS_INDEX = 82
	SM_MAXIMUMTOUCHES              SYSTEM_METRICS_INDEX = 95
	SM_MEDIACENTER                 SYSTEM_METRICS_INDEX = 87
	SM_MENUDROPALIGNMENT           SYSTEM_METRICS_INDEX = 40
	SM_MIDEASTENABLED              SYSTEM_METRICS_INDEX = 74
	SM_MOUSEPRESENT                SYSTEM_METRICS_INDEX = 19
	SM_MOUSEHORIZONTALWHEELPRESENT SYSTEM_METRICS_INDEX = 91
	SM_MOUSEWHEELPRESENT           SYSTEM_METRICS_INDEX = 75
	SM_NETWORK                     SYSTEM_METRICS_INDEX = 63
	SM_PENWINDOWS                  SYSTEM_METRICS_INDEX = 41
	SM_REMOTECONTROL               SYSTEM_METRICS_INDEX = 8193
	SM_REMOTESESSION               SYSTEM_METRICS_INDEX = 4096
	SM_SAMEDISPLAYFORMAT           SYSTEM_METRICS_INDEX = 81
	SM_SECURE                      SYSTEM_METRICS_INDEX = 44
	SM_SERVERR2                    SYSTEM_METRICS_INDEX = 89
	SM_SHOWSOUNDS                  SYSTEM_METRICS_INDEX = 70
	SM_SHUTTINGDOWN                SYSTEM_METRICS_INDEX = 8192
	SM_SLOWMACHINE                 SYSTEM_METRICS_INDEX = 73
	SM_STARTER                     SYSTEM_METRICS_INDEX = 88
	SM_SWAPBUTTON                  SYSTEM_METRICS_INDEX = 23
	SM_SYSTEMDOCKED                SYSTEM_METRICS_INDEX = 8196
	SM_TABLETPC                    SYSTEM_METRICS_INDEX = 86
	SM_XVIRTUALSCREEN              SYSTEM_METRICS_INDEX = 76
	SM_YVIRTUALSCREEN              SYSTEM_METRICS_INDEX = 77
)

// enum
type GET_ANCESTOR_FLAGS uint32

const (
	GA_PARENT    GET_ANCESTOR_FLAGS = 1
	GA_ROOT      GET_ANCESTOR_FLAGS = 2
	GA_ROOTOWNER GET_ANCESTOR_FLAGS = 3
)

// enum
type TILE_WINDOWS_HOW uint32

const (
	MDITILE_HORIZONTAL TILE_WINDOWS_HOW = 1
	MDITILE_VERTICAL   TILE_WINDOWS_HOW = 0
)

// enum
type WINDOW_DISPLAY_AFFINITY uint32

const (
	WDA_NONE               WINDOW_DISPLAY_AFFINITY = 0
	WDA_MONITOR            WINDOW_DISPLAY_AFFINITY = 1
	WDA_EXCLUDEFROMCAPTURE WINDOW_DISPLAY_AFFINITY = 17
)

// enum
type FOREGROUND_WINDOW_LOCK_CODE uint32

const (
	LSFW_LOCK   FOREGROUND_WINDOW_LOCK_CODE = 1
	LSFW_UNLOCK FOREGROUND_WINDOW_LOCK_CODE = 2
)

// enum
// flags
type CASCADE_WINDOWS_HOW uint32

const (
	MDITILE_SKIPDISABLED CASCADE_WINDOWS_HOW = 2
	MDITILE_ZORDER       CASCADE_WINDOWS_HOW = 4
)

// enum
type WINDOW_MESSAGE_FILTER_ACTION uint32

const (
	MSGFLT_ALLOW    WINDOW_MESSAGE_FILTER_ACTION = 1
	MSGFLT_DISALLOW WINDOW_MESSAGE_FILTER_ACTION = 2
	MSGFLT_RESET    WINDOW_MESSAGE_FILTER_ACTION = 0
)

// enum
// flags
type GET_MENU_DEFAULT_ITEM_FLAGS uint32

const (
	GMDI_GOINTOPOPUPS GET_MENU_DEFAULT_ITEM_FLAGS = 2
	GMDI_USEDISABLED  GET_MENU_DEFAULT_ITEM_FLAGS = 1
)

// enum
type MSGFLTINFO_STATUS uint32

const (
	MSGFLTINFO_NONE                     MSGFLTINFO_STATUS = 0
	MSGFLTINFO_ALLOWED_HIGHER           MSGFLTINFO_STATUS = 3
	MSGFLTINFO_ALREADYALLOWED_FORWND    MSGFLTINFO_STATUS = 1
	MSGFLTINFO_ALREADYDISALLOWED_FORWND MSGFLTINFO_STATUS = 2
)

// enum
// flags
type MOUSEHOOKSTRUCTEX_MOUSE_DATA uint32

const (
	XBUTTON1 MOUSEHOOKSTRUCTEX_MOUSE_DATA = 1
	XBUTTON2 MOUSEHOOKSTRUCTEX_MOUSE_DATA = 2
)

// enum
// flags
type MENU_ITEM_MASK uint32

const (
	MIIM_BITMAP     MENU_ITEM_MASK = 128
	MIIM_CHECKMARKS MENU_ITEM_MASK = 8
	MIIM_DATA       MENU_ITEM_MASK = 32
	MIIM_FTYPE      MENU_ITEM_MASK = 256
	MIIM_ID         MENU_ITEM_MASK = 2
	MIIM_STATE      MENU_ITEM_MASK = 1
	MIIM_STRING     MENU_ITEM_MASK = 64
	MIIM_SUBMENU    MENU_ITEM_MASK = 4
	MIIM_TYPE       MENU_ITEM_MASK = 16
)

// enum
// flags
type FLASHWINFO_FLAGS uint32

const (
	FLASHW_ALL       FLASHWINFO_FLAGS = 3
	FLASHW_CAPTION   FLASHWINFO_FLAGS = 1
	FLASHW_STOP      FLASHWINFO_FLAGS = 0
	FLASHW_TIMER     FLASHWINFO_FLAGS = 4
	FLASHW_TIMERNOFG FLASHWINFO_FLAGS = 12
	FLASHW_TRAY      FLASHWINFO_FLAGS = 2
)

// enum
type CURSORINFO_FLAGS uint32

const (
	CURSOR_SHOWING    CURSORINFO_FLAGS = 1
	CURSOR_SUPPRESSED CURSORINFO_FLAGS = 2
)

// enum
// flags
type MENUINFO_STYLE uint32

const (
	MNS_AUTODISMISS MENUINFO_STYLE = 268435456
	MNS_CHECKORBMP  MENUINFO_STYLE = 67108864
	MNS_DRAGDROP    MENUINFO_STYLE = 536870912
	MNS_MODELESS    MENUINFO_STYLE = 1073741824
	MNS_NOCHECK     MENUINFO_STYLE = 2147483648
	MNS_NOTIFYBYPOS MENUINFO_STYLE = 134217728
)

// enum
// flags
type WINDOWPLACEMENT_FLAGS uint32

const (
	WPF_ASYNCWINDOWPLACEMENT WINDOWPLACEMENT_FLAGS = 4
	WPF_RESTORETOMAXIMIZED   WINDOWPLACEMENT_FLAGS = 2
	WPF_SETMINPOSITION       WINDOWPLACEMENT_FLAGS = 1
)

// enum
// flags
type MENUINFO_MASK uint32

const (
	MIM_APPLYTOSUBMENUS MENUINFO_MASK = 2147483648
	MIM_BACKGROUND      MENUINFO_MASK = 2
	MIM_HELPID          MENUINFO_MASK = 4
	MIM_MAXHEIGHT       MENUINFO_MASK = 1
	MIM_MENUDATA        MENUINFO_MASK = 8
	MIM_STYLE           MENUINFO_MASK = 16
)

// enum
type MINIMIZEDMETRICS_ARRANGE int32

const (
	ARW_BOTTOMLEFT  MINIMIZEDMETRICS_ARRANGE = 0
	ARW_BOTTOMRIGHT MINIMIZEDMETRICS_ARRANGE = 1
	ARW_TOPLEFT     MINIMIZEDMETRICS_ARRANGE = 2
	ARW_TOPRIGHT    MINIMIZEDMETRICS_ARRANGE = 3
)

// enum
// flags
type SCROLLINFO_MASK uint32

const (
	SIF_ALL             SCROLLINFO_MASK = 23
	SIF_DISABLENOSCROLL SCROLLINFO_MASK = 8
	SIF_PAGE            SCROLLINFO_MASK = 2
	SIF_POS             SCROLLINFO_MASK = 4
	SIF_RANGE           SCROLLINFO_MASK = 1
	SIF_TRACKPOS        SCROLLINFO_MASK = 16
)

// enum
type MENUGETOBJECTINFO_FLAGS uint32

const (
	MNGOF_BOTTOMGAP MENUGETOBJECTINFO_FLAGS = 2
	MNGOF_TOPGAP    MENUGETOBJECTINFO_FLAGS = 1
)

// enum
// flags
type GUITHREADINFO_FLAGS uint32

const (
	GUI_CARETBLINKING  GUITHREADINFO_FLAGS = 1
	GUI_INMENUMODE     GUITHREADINFO_FLAGS = 4
	GUI_INMOVESIZE     GUITHREADINFO_FLAGS = 2
	GUI_POPUPMENUMODE  GUITHREADINFO_FLAGS = 16
	GUI_SYSTEMMENUMODE GUITHREADINFO_FLAGS = 8
)

// enum
// flags
type KBDLLHOOKSTRUCT_FLAGS uint32

const (
	LLKHF_EXTENDED          KBDLLHOOKSTRUCT_FLAGS = 1
	LLKHF_ALTDOWN           KBDLLHOOKSTRUCT_FLAGS = 32
	LLKHF_UP                KBDLLHOOKSTRUCT_FLAGS = 128
	LLKHF_INJECTED          KBDLLHOOKSTRUCT_FLAGS = 16
	LLKHF_LOWER_IL_INJECTED KBDLLHOOKSTRUCT_FLAGS = 2
)

// enum
// flags
type DI_FLAGS uint32

const (
	DI_MASK        DI_FLAGS = 1
	DI_IMAGE       DI_FLAGS = 2
	DI_NORMAL      DI_FLAGS = 3
	DI_COMPAT      DI_FLAGS = 4
	DI_DEFAULTSIZE DI_FLAGS = 8
	DI_NOMIRROR    DI_FLAGS = 16
)

// enum
type POINTER_INPUT_TYPE int32

const (
	PT_POINTER  POINTER_INPUT_TYPE = 1
	PT_TOUCH    POINTER_INPUT_TYPE = 2
	PT_PEN      POINTER_INPUT_TYPE = 3
	PT_MOUSE    POINTER_INPUT_TYPE = 4
	PT_TOUCHPAD POINTER_INPUT_TYPE = 5
)

// enum
type EDIT_CONTROL_FEATURE int32

const (
	EDIT_CONTROL_FEATURE_ENTERPRISE_DATA_PROTECTION_PASTE_SUPPORT EDIT_CONTROL_FEATURE = 0
	EDIT_CONTROL_FEATURE_PASTE_NOTIFICATIONS                      EDIT_CONTROL_FEATURE = 1
)

// enum
type HANDEDNESS int32

const (
	HANDEDNESS_LEFT  HANDEDNESS = 0
	HANDEDNESS_RIGHT HANDEDNESS = 1
)

// enum
type MrmPlatformVersion int32

const (
	MrmPlatformVersion_Default         MrmPlatformVersion = 0
	MrmPlatformVersion_Windows10_0_0_0 MrmPlatformVersion = 17432576
	MrmPlatformVersion_Windows10_0_0_5 MrmPlatformVersion = 17432581
)

// enum
type MrmPackagingMode int32

const (
	MrmPackagingModeStandaloneFile MrmPackagingMode = 0
	MrmPackagingModeAutoSplit      MrmPackagingMode = 1
	MrmPackagingModeResourcePack   MrmPackagingMode = 2
)

// enum
type MrmPackagingOptions int32

const (
	MrmPackagingOptionsNone                        MrmPackagingOptions = 0
	MrmPackagingOptionsOmitSchemaFromResourcePacks MrmPackagingOptions = 1
	MrmPackagingOptionsSplitLanguageVariants       MrmPackagingOptions = 2
)

// enum
type MrmDumpType int32

const (
	MrmDumpType_Basic    MrmDumpType = 0
	MrmDumpType_Detailed MrmDumpType = 1
	MrmDumpType_Schema   MrmDumpType = 2
)

// enum
type MrmResourceIndexerMessageSeverity int32

const (
	MrmResourceIndexerMessageSeverityVerbose MrmResourceIndexerMessageSeverity = 0
	MrmResourceIndexerMessageSeverityInfo    MrmResourceIndexerMessageSeverity = 1
	MrmResourceIndexerMessageSeverityWarning MrmResourceIndexerMessageSeverity = 2
	MrmResourceIndexerMessageSeverityError   MrmResourceIndexerMessageSeverity = 3
)

// enum
type MrmIndexerFlags int32

const (
	MrmIndexerFlagsNone                  MrmIndexerFlags = 0
	MrmIndexerFlagsAutoMerge             MrmIndexerFlags = 1
	MrmIndexerFlagsCreateContentChecksum MrmIndexerFlags = 2
)

// structs

type MESSAGE_RESOURCE_ENTRY struct {
	Length uint16
	Flags  uint16
	Text   [1]byte
}

type MESSAGE_RESOURCE_BLOCK struct {
	LowId           uint32
	HighId          uint32
	OffsetToEntries uint32
}

type MESSAGE_RESOURCE_DATA struct {
	NumberOfBlocks uint32
	Blocks         [1]MESSAGE_RESOURCE_BLOCK
}

type CBT_CREATEWNDA struct {
	Lpcs            *CREATESTRUCTA
	HwndInsertAfter HWND
}

type CBT_CREATEWND = CBT_CREATEWNDW
type CBT_CREATEWNDW struct {
	Lpcs            *CREATESTRUCTW
	HwndInsertAfter HWND
}

type CBTACTIVATESTRUCT struct {
	FMouse     BOOL
	HWndActive HWND
}

type SHELLHOOKINFO struct {
	Hwnd HWND
	Rc   RECT
}

type EVENTMSG struct {
	Message uint32
	ParamL  uint32
	ParamH  uint32
	Time    uint32
	Hwnd    HWND
}

type CWPSTRUCT struct {
	LParam  LPARAM
	WParam  WPARAM
	Message uint32
	Hwnd    HWND
}

type CWPRETSTRUCT struct {
	LResult LRESULT
	LParam  LPARAM
	WParam  WPARAM
	Message uint32
	Hwnd    HWND
}

type KBDLLHOOKSTRUCT struct {
	VkCode      uint32
	ScanCode    uint32
	Flags       KBDLLHOOKSTRUCT_FLAGS
	Time        uint32
	DwExtraInfo uintptr
}

type MSLLHOOKSTRUCT struct {
	Pt          POINT
	MouseData   MOUSEHOOKSTRUCTEX_MOUSE_DATA
	Flags       uint32
	Time        uint32
	DwExtraInfo uintptr
}

type DEBUGHOOKINFO struct {
	IdThread          uint32
	IdThreadInstaller uint32
	LParam            LPARAM
	WParam            WPARAM
	Code              int32
}

type MOUSEHOOKSTRUCT struct {
	Pt           POINT
	Hwnd         HWND
	WHitTestCode uint32
	DwExtraInfo  uintptr
}

type MOUSEHOOKSTRUCTEX struct {
	MOUSEHOOKSTRUCT
	MouseData MOUSEHOOKSTRUCTEX_MOUSE_DATA
}

type HARDWAREHOOKSTRUCT struct {
	Hwnd    HWND
	Message uint32
	WParam  WPARAM
	LParam  LPARAM
}

type WNDCLASSEXA struct {
	CbSize        uint32
	Style         WNDCLASS_STYLES
	LpfnWndProc   WNDPROC
	CbClsExtra    int32
	CbWndExtra    int32
	HInstance     HINSTANCE
	HIcon         HICON
	HCursor       HCURSOR
	HbrBackground HBRUSH
	LpszMenuName  PSTR
	LpszClassName PSTR
	HIconSm       HICON
}

type WNDCLASSEX = WNDCLASSEXW
type WNDCLASSEXW struct {
	CbSize        uint32
	Style         WNDCLASS_STYLES
	LpfnWndProc   WNDPROC
	CbClsExtra    int32
	CbWndExtra    int32
	HInstance     HINSTANCE
	HIcon         HICON
	HCursor       HCURSOR
	HbrBackground HBRUSH
	LpszMenuName  PWSTR
	LpszClassName PWSTR
	HIconSm       HICON
}

type WNDCLASSA struct {
	Style         WNDCLASS_STYLES
	LpfnWndProc   WNDPROC
	CbClsExtra    int32
	CbWndExtra    int32
	HInstance     HINSTANCE
	HIcon         HICON
	HCursor       HCURSOR
	HbrBackground HBRUSH
	LpszMenuName  PSTR
	LpszClassName PSTR
}

type WNDCLASS = WNDCLASSW
type WNDCLASSW struct {
	Style         WNDCLASS_STYLES
	LpfnWndProc   WNDPROC
	CbClsExtra    int32
	CbWndExtra    int32
	HInstance     HINSTANCE
	HIcon         HICON
	HCursor       HCURSOR
	HbrBackground HBRUSH
	LpszMenuName  PWSTR
	LpszClassName PWSTR
}

type MSG struct {
	Hwnd    HWND
	Message uint32
	WParam  WPARAM
	LParam  LPARAM
	Time    uint32
	Pt      POINT
}

type MINMAXINFO struct {
	PtReserved     POINT
	PtMaxSize      POINT
	PtMaxPosition  POINT
	PtMinTrackSize POINT
	PtMaxTrackSize POINT
}

type MDINEXTMENU struct {
	HmenuIn   HMENU
	HmenuNext HMENU
	HwndNext  HWND
}

type WINDOWPOS struct {
	Hwnd            HWND
	HwndInsertAfter HWND
	X               int32
	Y               int32
	Cx              int32
	Cy              int32
	Flags           SET_WINDOW_POS_FLAGS
}

type NCCALCSIZE_PARAMS struct {
	Rgrc  [3]RECT
	Lppos *WINDOWPOS
}

type ACCEL struct {
	FVirt byte
	Key   uint16
	Cmd   uint16
}

type CREATESTRUCTA struct {
	LpCreateParams unsafe.Pointer
	HInstance      HINSTANCE
	HMenu          HMENU
	HwndParent     HWND
	Cy             int32
	Cx             int32
	Y              int32
	X              int32
	Style          int32
	LpszName       PSTR
	LpszClass      PSTR
	DwExStyle      uint32
}

type CREATESTRUCT = CREATESTRUCTW
type CREATESTRUCTW struct {
	LpCreateParams unsafe.Pointer
	HInstance      HINSTANCE
	HMenu          HMENU
	HwndParent     HWND
	Cy             int32
	Cx             int32
	Y              int32
	X              int32
	Style          int32
	LpszName       PWSTR
	LpszClass      PWSTR
	DwExStyle      uint32
}

type WINDOWPLACEMENT struct {
	Length           uint32
	Flags            WINDOWPLACEMENT_FLAGS
	ShowCmd          SHOW_WINDOW_CMD
	PtMinPosition    POINT
	PtMaxPosition    POINT
	RcNormalPosition RECT
}

type STYLESTRUCT struct {
	StyleOld uint32
	StyleNew uint32
}

type UPDATELAYEREDWINDOWINFO struct {
	CbSize   uint32
	HdcDst   HDC
	PptDst   *POINT
	Psize    *SIZE
	HdcSrc   HDC
	PptSrc   *POINT
	CrKey    uint32
	Pblend   *BLENDFUNCTION
	DwFlags  UPDATE_LAYERED_WINDOW_FLAGS
	PrcDirty *RECT
}

type FLASHWINFO struct {
	CbSize    uint32
	Hwnd      HWND
	DwFlags   FLASHWINFO_FLAGS
	UCount    uint32
	DwTimeout uint32
}

type DLGTEMPLATE struct {
	Style           uint32
	DwExtendedStyle uint32
	Cdit            uint16
	X               int16
	Y               int16
	Cx              int16
	Cy              int16
}

type DLGITEMTEMPLATE struct {
	Style           uint32
	DwExtendedStyle uint32
	X               int16
	Y               int16
	Cx              int16
	Cy              int16
	Id              uint16
}

type TPMPARAMS struct {
	CbSize    uint32
	RcExclude RECT
}

type MENUINFO struct {
	CbSize          uint32
	FMask           MENUINFO_MASK
	DwStyle         MENUINFO_STYLE
	CyMax           uint32
	HbrBack         HBRUSH
	DwContextHelpID uint32
	DwMenuData      uintptr
}

type MENUGETOBJECTINFO struct {
	DwFlags MENUGETOBJECTINFO_FLAGS
	UPos    uint32
	Hmenu   HMENU
	Riid    unsafe.Pointer
	PvObj   unsafe.Pointer
}

type MENUITEMINFOA struct {
	CbSize        uint32
	FMask         MENU_ITEM_MASK
	FType         MENU_ITEM_TYPE
	FState        MENU_ITEM_STATE
	WID           uint32
	HSubMenu      HMENU
	HbmpChecked   HBITMAP
	HbmpUnchecked HBITMAP
	DwItemData    uintptr
	DwTypeData    PSTR
	Cch           uint32
	HbmpItem      HBITMAP
}

type MENUITEMINFO = MENUITEMINFOW
type MENUITEMINFOW struct {
	CbSize        uint32
	FMask         MENU_ITEM_MASK
	FType         MENU_ITEM_TYPE
	FState        MENU_ITEM_STATE
	WID           uint32
	HSubMenu      HMENU
	HbmpChecked   HBITMAP
	HbmpUnchecked HBITMAP
	DwItemData    uintptr
	DwTypeData    PWSTR
	Cch           uint32
	HbmpItem      HBITMAP
}

type DROPSTRUCT struct {
	HwndSource    HWND
	HwndSink      HWND
	WFmt          uint32
	DwData        uintptr
	PtDrop        POINT
	DwControlData uint32
}

type MSGBOXPARAMSA struct {
	CbSize             uint32
	HwndOwner          HWND
	HInstance          HINSTANCE
	LpszText           PSTR
	LpszCaption        PSTR
	DwStyle            MESSAGEBOX_STYLE
	LpszIcon           PSTR
	DwContextHelpId    uintptr
	LpfnMsgBoxCallback MSGBOXCALLBACK
	DwLanguageId       uint32
}

type MSGBOXPARAMS = MSGBOXPARAMSW
type MSGBOXPARAMSW struct {
	CbSize             uint32
	HwndOwner          HWND
	HInstance          HINSTANCE
	LpszText           PWSTR
	LpszCaption        PWSTR
	DwStyle            MESSAGEBOX_STYLE
	LpszIcon           PWSTR
	DwContextHelpId    uintptr
	LpfnMsgBoxCallback MSGBOXCALLBACK
	DwLanguageId       uint32
}

type MENUITEMTEMPLATEHEADER struct {
	VersionNumber uint16
	Offset        uint16
}

type MENUITEMTEMPLATE struct {
	MtOption uint16
	MtID     uint16
	MtString [1]uint16
}

type ICONINFO struct {
	FIcon    BOOL
	XHotspot uint32
	YHotspot uint32
	HbmMask  HBITMAP
	HbmColor HBITMAP
}

type CURSORSHAPE struct {
	XHotSpot  int32
	YHotSpot  int32
	Cx        int32
	Cy        int32
	CbWidth   int32
	Planes    byte
	BitsPixel byte
}

type ICONINFOEXA struct {
	CbSize    uint32
	FIcon     BOOL
	XHotspot  uint32
	YHotspot  uint32
	HbmMask   HBITMAP
	HbmColor  HBITMAP
	WResID    uint16
	SzModName [260]CHAR
	SzResName [260]CHAR
}

type ICONINFOEX = ICONINFOEXW
type ICONINFOEXW struct {
	CbSize    uint32
	FIcon     BOOL
	XHotspot  uint32
	YHotspot  uint32
	HbmMask   HBITMAP
	HbmColor  HBITMAP
	WResID    uint16
	SzModName [260]uint16
	SzResName [260]uint16
}

type SCROLLINFO struct {
	CbSize    uint32
	FMask     SCROLLINFO_MASK
	NMin      int32
	NMax      int32
	NPage     uint32
	NPos      int32
	NTrackPos int32
}

type MDICREATESTRUCTA struct {
	SzClass PSTR
	SzTitle PSTR
	HOwner  HANDLE
	X       int32
	Y       int32
	Cx      int32
	Cy      int32
	Style   WINDOW_STYLE
	LParam  LPARAM
}

type MDICREATESTRUCT = MDICREATESTRUCTW
type MDICREATESTRUCTW struct {
	SzClass PWSTR
	SzTitle PWSTR
	HOwner  HANDLE
	X       int32
	Y       int32
	Cx      int32
	Cy      int32
	Style   WINDOW_STYLE
	LParam  LPARAM
}

type CLIENTCREATESTRUCT struct {
	HWindowMenu  HANDLE
	IdFirstChild uint32
}

type TouchPredictionParameters struct {
	CbSize          uint32
	DwLatency       uint32
	DwSampleTime    uint32
	BUseHWTimeStamp uint32
}

type NONCLIENTMETRICSA struct {
	CbSize             uint32
	IBorderWidth       int32
	IScrollWidth       int32
	IScrollHeight      int32
	ICaptionWidth      int32
	ICaptionHeight     int32
	LfCaptionFont      LOGFONTA
	ISmCaptionWidth    int32
	ISmCaptionHeight   int32
	LfSmCaptionFont    LOGFONTA
	IMenuWidth         int32
	IMenuHeight        int32
	LfMenuFont         LOGFONTA
	LfStatusFont       LOGFONTA
	LfMessageFont      LOGFONTA
	IPaddedBorderWidth int32
}

type NONCLIENTMETRICS = NONCLIENTMETRICSW
type NONCLIENTMETRICSW struct {
	CbSize             uint32
	IBorderWidth       int32
	IScrollWidth       int32
	IScrollHeight      int32
	ICaptionWidth      int32
	ICaptionHeight     int32
	LfCaptionFont      LOGFONTW
	ISmCaptionWidth    int32
	ISmCaptionHeight   int32
	LfSmCaptionFont    LOGFONTW
	IMenuWidth         int32
	IMenuHeight        int32
	LfMenuFont         LOGFONTW
	LfStatusFont       LOGFONTW
	LfMessageFont      LOGFONTW
	IPaddedBorderWidth int32
}

type MINIMIZEDMETRICS struct {
	CbSize   uint32
	IWidth   int32
	IHorzGap int32
	IVertGap int32
	IArrange MINIMIZEDMETRICS_ARRANGE
}

type ICONMETRICSA struct {
	CbSize       uint32
	IHorzSpacing int32
	IVertSpacing int32
	ITitleWrap   int32
	LfFont       LOGFONTA
}

type ICONMETRICS = ICONMETRICSW
type ICONMETRICSW struct {
	CbSize       uint32
	IHorzSpacing int32
	IVertSpacing int32
	ITitleWrap   int32
	LfFont       LOGFONTW
}

type ANIMATIONINFO struct {
	CbSize      uint32
	IMinAnimate int32
}

type AUDIODESCRIPTION struct {
	CbSize  uint32
	Enabled BOOL
	Locale  uint32
}

type GUITHREADINFO struct {
	CbSize        uint32
	Flags         GUITHREADINFO_FLAGS
	HwndActive    HWND
	HwndFocus     HWND
	HwndCapture   HWND
	HwndMenuOwner HWND
	HwndMoveSize  HWND
	HwndCaret     HWND
	RcCaret       RECT
}

type CURSORINFO struct {
	CbSize      uint32
	Flags       CURSORINFO_FLAGS
	HCursor     HCURSOR
	PtScreenPos POINT
}

type WINDOWINFO struct {
	CbSize          uint32
	RcWindow        RECT
	RcClient        RECT
	DwStyle         uint32
	DwExStyle       uint32
	DwWindowStatus  uint32
	CxWindowBorders uint32
	CyWindowBorders uint32
	AtomWindowType  uint16
	WCreatorVersion uint16
}

type TITLEBARINFO struct {
	CbSize     uint32
	RcTitleBar RECT
	Rgstate    [6]uint32
}

type TITLEBARINFOEX struct {
	CbSize     uint32
	RcTitleBar RECT
	Rgstate    [6]uint32
	Rgrect     [6]RECT
}

type MENUBARINFO struct {
	CbSize    uint32
	RcBar     RECT
	HMenu     HMENU
	HwndMenu  HWND
	Bitfield_ int32
}

type SCROLLBARINFO struct {
	CbSize        uint32
	RcScrollBar   RECT
	DxyLineButton int32
	XyThumbTop    int32
	XyThumbBottom int32
	Reserved      int32
	Rgstate       [6]uint32
}

type ALTTABINFO struct {
	CbSize    uint32
	CItems    int32
	CColumns  int32
	CRows     int32
	IColFocus int32
	IRowFocus int32
	CxItem    int32
	CyItem    int32
	PtStart   POINT
}

type CHANGEFILTERSTRUCT struct {
	CbSize    uint32
	ExtStatus MSGFLTINFO_STATUS
}

type IndexedResourceQualifier struct {
	Name  PWSTR
	Value PWSTR
}

type MrmResourceIndexerHandle struct {
	Handle unsafe.Pointer
}

type MrmResourceIndexerMessage struct {
	Severity MrmResourceIndexerMessageSeverity
	Id       uint32
	Text     PWSTR
}

// func types

type WNDPROC = uintptr
type WNDPROC_func = func(param0 HWND, param1 uint32, param2 WPARAM, param3 LPARAM) LRESULT

type DLGPROC = uintptr
type DLGPROC_func = func(param0 HWND, param1 uint32, param2 WPARAM, param3 LPARAM) uintptr

type TIMERPROC = uintptr
type TIMERPROC_func = func(param0 HWND, param1 uint32, param2 uintptr, param3 uint32)

type WNDENUMPROC = uintptr
type WNDENUMPROC_func = func(param0 HWND, param1 LPARAM) BOOL

type HOOKPROC = uintptr
type HOOKPROC_func = func(code int32, wParam WPARAM, lParam LPARAM) LRESULT

type SENDASYNCPROC = uintptr
type SENDASYNCPROC_func = func(param0 HWND, param1 uint32, param2 uintptr, param3 LRESULT)

type PROPENUMPROCA = uintptr
type PROPENUMPROCA_func = func(param0 HWND, param1 PSTR, param2 HANDLE) BOOL

type PROPENUMPROCW = uintptr
type PROPENUMPROCW_func = func(param0 HWND, param1 PWSTR, param2 HANDLE) BOOL

type PROPENUMPROCEXA = uintptr
type PROPENUMPROCEXA_func = func(param0 HWND, param1 PSTR, param2 HANDLE, param3 uintptr) BOOL

type PROPENUMPROCEXW = uintptr
type PROPENUMPROCEXW_func = func(param0 HWND, param1 PWSTR, param2 HANDLE, param3 uintptr) BOOL

type NAMEENUMPROCA = uintptr
type NAMEENUMPROCA_func = func(param0 PSTR, param1 LPARAM) BOOL

type NAMEENUMPROCW = uintptr
type NAMEENUMPROCW_func = func(param0 PWSTR, param1 LPARAM) BOOL

type PREGISTERCLASSNAMEW = uintptr
type PREGISTERCLASSNAMEW_func = func(param0 PWSTR) BOOLEAN

type MSGBOXCALLBACK = uintptr
type MSGBOXCALLBACK_func = func(lpHelpInfo *HELPINFO)

var (
	pLoadStringA                   uintptr
	pLoadStringW                   uintptr
	pWvsprintfA                    uintptr
	pWvsprintfW                    uintptr
	pWsprintfA                     uintptr
	pWsprintfW                     uintptr
	pIsHungAppWindow               uintptr
	pDisableProcessWindowsGhosting uintptr
	pRegisterWindowMessageA        uintptr
	pRegisterWindowMessageW        uintptr
	pGetMessageA                   uintptr
	pGetMessageW                   uintptr
	pTranslateMessage              uintptr
	pDispatchMessageA              uintptr
	pDispatchMessageW              uintptr
	pSetMessageQueue               uintptr
	pPeekMessageA                  uintptr
	pPeekMessageW                  uintptr
	pGetMessagePos                 uintptr
	pGetMessageTime                uintptr
	pGetMessageExtraInfo           uintptr
	pIsWow64Message                uintptr
	pSetMessageExtraInfo           uintptr
	pSendMessageA                  uintptr
	pSendMessageW                  uintptr
	pSendMessageTimeoutA           uintptr
	pSendMessageTimeoutW           uintptr
	pSendNotifyMessageA            uintptr
	pSendNotifyMessageW            uintptr
	pSendMessageCallbackA          uintptr
	pSendMessageCallbackW          uintptr
	pRegisterDeviceNotificationA   uintptr
	pRegisterDeviceNotificationW   uintptr
	pPostMessageA                  uintptr
	pPostMessageW                  uintptr
	pPostThreadMessageA            uintptr
	pPostThreadMessageW            uintptr
	pReplyMessage                  uintptr
	pWaitMessage                   uintptr
	pDefWindowProcA                uintptr
	pDefWindowProcW                uintptr
	pPostQuitMessage               uintptr
	pCallWindowProcA               uintptr
	pCallWindowProcW               uintptr
	pInSendMessage                 uintptr
	pInSendMessageEx               uintptr
	pRegisterClassA                uintptr
	pRegisterClassW                uintptr
	pUnregisterClassA              uintptr
	pUnregisterClassW              uintptr
	pGetClassInfoA                 uintptr
	pGetClassInfoW                 uintptr
	pRegisterClassExA              uintptr
	pRegisterClassExW              uintptr
	pGetClassInfoExA               uintptr
	pGetClassInfoExW               uintptr
	pCreateWindowExA               uintptr
	pCreateWindowExW               uintptr
	pIsWindow                      uintptr
	pIsMenu                        uintptr
	pIsChild                       uintptr
	pDestroyWindow                 uintptr
	pShowWindow                    uintptr
	pAnimateWindow                 uintptr
	pUpdateLayeredWindow           uintptr
	pUpdateLayeredWindowIndirect   uintptr
	pGetLayeredWindowAttributes    uintptr
	pSetLayeredWindowAttributes    uintptr
	pShowWindowAsync               uintptr
	pFlashWindow                   uintptr
	pFlashWindowEx                 uintptr
	pShowOwnedPopups               uintptr
	pOpenIcon                      uintptr
	pCloseWindow                   uintptr
	pMoveWindow                    uintptr
	pSetWindowPos                  uintptr
	pGetWindowPlacement            uintptr
	pSetWindowPlacement            uintptr
	pGetWindowDisplayAffinity      uintptr
	pSetWindowDisplayAffinity      uintptr
	pBeginDeferWindowPos           uintptr
	pDeferWindowPos                uintptr
	pEndDeferWindowPos             uintptr
	pIsWindowVisible               uintptr
	pIsIconic                      uintptr
	pAnyPopup                      uintptr
	pBringWindowToTop              uintptr
	pIsZoomed                      uintptr
	pCreateDialogParamA            uintptr
	pCreateDialogParamW            uintptr
	pCreateDialogIndirectParamA    uintptr
	pCreateDialogIndirectParamW    uintptr
	pDialogBoxParamA               uintptr
	pDialogBoxParamW               uintptr
	pDialogBoxIndirectParamA       uintptr
	pDialogBoxIndirectParamW       uintptr
	pEndDialog                     uintptr
	pGetDlgItem                    uintptr
	pSetDlgItemInt                 uintptr
	pGetDlgItemInt                 uintptr
	pSetDlgItemTextA               uintptr
	pSetDlgItemTextW               uintptr
	pGetDlgItemTextA               uintptr
	pGetDlgItemTextW               uintptr
	pSendDlgItemMessageA           uintptr
	pSendDlgItemMessageW           uintptr
	pGetNextDlgGroupItem           uintptr
	pGetNextDlgTabItem             uintptr
	pGetDlgCtrlID                  uintptr
	pGetDialogBaseUnits            uintptr
	pDefDlgProcA                   uintptr
	pDefDlgProcW                   uintptr
	pCallMsgFilterA                uintptr
	pCallMsgFilterW                uintptr
	pCharToOemA                    uintptr
	pCharToOemW                    uintptr
	pOemToCharA                    uintptr
	pOemToCharW                    uintptr
	pCharToOemBuffA                uintptr
	pCharToOemBuffW                uintptr
	pOemToCharBuffA                uintptr
	pOemToCharBuffW                uintptr
	pCharUpperA                    uintptr
	pCharUpperW                    uintptr
	pCharUpperBuffA                uintptr
	pCharUpperBuffW                uintptr
	pCharLowerA                    uintptr
	pCharLowerW                    uintptr
	pCharLowerBuffA                uintptr
	pCharLowerBuffW                uintptr
	pCharNextA                     uintptr
	pCharNextW                     uintptr
	pCharPrevA                     uintptr
	pCharPrevW                     uintptr
	pCharNextExA                   uintptr
	pCharPrevExA                   uintptr
	pIsCharAlphaA                  uintptr
	pIsCharAlphaW                  uintptr
	pIsCharAlphaNumericA           uintptr
	pIsCharAlphaNumericW           uintptr
	pIsCharUpperA                  uintptr
	pIsCharUpperW                  uintptr
	pIsCharLowerA                  uintptr
	pGetInputState                 uintptr
	pGetQueueStatus                uintptr
	pMsgWaitForMultipleObjects     uintptr
	pMsgWaitForMultipleObjectsEx   uintptr
	pSetTimer                      uintptr
	pSetCoalescableTimer           uintptr
	pKillTimer                     uintptr
	pIsWindowUnicode               uintptr
	pLoadAcceleratorsA             uintptr
	pLoadAcceleratorsW             uintptr
	pCreateAcceleratorTableA       uintptr
	pCreateAcceleratorTableW       uintptr
	pDestroyAcceleratorTable       uintptr
	pCopyAcceleratorTableA         uintptr
	pCopyAcceleratorTableW         uintptr
	pTranslateAcceleratorA         uintptr
	pTranslateAcceleratorW         uintptr
	pGetSystemMetrics              uintptr
	pLoadMenuA                     uintptr
	pLoadMenuW                     uintptr
	pLoadMenuIndirectA             uintptr
	pLoadMenuIndirectW             uintptr
	pGetMenu                       uintptr
	pSetMenu                       uintptr
	pChangeMenuA                   uintptr
	pChangeMenuW                   uintptr
	pHiliteMenuItem                uintptr
	pGetMenuStringA                uintptr
	pGetMenuStringW                uintptr
	pGetMenuState                  uintptr
	pDrawMenuBar                   uintptr
	pGetSystemMenu                 uintptr
	pCreateMenu                    uintptr
	pCreatePopupMenu               uintptr
	pDestroyMenu                   uintptr
	pCheckMenuItem                 uintptr
	pEnableMenuItem                uintptr
	pGetSubMenu                    uintptr
	pGetMenuItemID                 uintptr
	pGetMenuItemCount              uintptr
	pInsertMenuA                   uintptr
	pInsertMenuW                   uintptr
	pAppendMenuA                   uintptr
	pAppendMenuW                   uintptr
	pModifyMenuA                   uintptr
	pModifyMenuW                   uintptr
	pRemoveMenu                    uintptr
	pDeleteMenu                    uintptr
	pSetMenuItemBitmaps            uintptr
	pGetMenuCheckMarkDimensions    uintptr
	pTrackPopupMenu                uintptr
	pTrackPopupMenuEx              uintptr
	pCalculatePopupWindowPosition  uintptr
	pGetMenuInfo                   uintptr
	pSetMenuInfo                   uintptr
	pEndMenu                       uintptr
	pInsertMenuItemA               uintptr
	pInsertMenuItemW               uintptr
	pGetMenuItemInfoA              uintptr
	pGetMenuItemInfoW              uintptr
	pSetMenuItemInfoA              uintptr
	pSetMenuItemInfoW              uintptr
	pGetMenuDefaultItem            uintptr
	pSetMenuDefaultItem            uintptr
	pGetMenuItemRect               uintptr
	pMenuItemFromPoint             uintptr
	pDragObject                    uintptr
	pDrawIcon                      uintptr
	pGetForegroundWindow           uintptr
	pSwitchToThisWindow            uintptr
	pSetForegroundWindow           uintptr
	pAllowSetForegroundWindow      uintptr
	pLockSetForegroundWindow       uintptr
	pScrollWindow                  uintptr
	pScrollDC                      uintptr
	pScrollWindowEx                uintptr
	pGetScrollPos                  uintptr
	pGetScrollRange                uintptr
	pSetPropA                      uintptr
	pSetPropW                      uintptr
	pGetPropA                      uintptr
	pGetPropW                      uintptr
	pRemovePropA                   uintptr
	pRemovePropW                   uintptr
	pEnumPropsExA                  uintptr
	pEnumPropsExW                  uintptr
	pEnumPropsA                    uintptr
	pEnumPropsW                    uintptr
	pSetWindowTextA                uintptr
	pSetWindowTextW                uintptr
	pGetWindowTextA                uintptr
	pGetWindowTextW                uintptr
	pGetWindowTextLengthA          uintptr
	pGetWindowTextLengthW          uintptr
	pGetClientRect                 uintptr
	pGetWindowRect                 uintptr
	pAdjustWindowRect              uintptr
	pAdjustWindowRectEx            uintptr
	pMessageBoxA                   uintptr
	pMessageBoxW                   uintptr
	pMessageBoxExA                 uintptr
	pMessageBoxExW                 uintptr
	pMessageBoxIndirectA           uintptr
	pMessageBoxIndirectW           uintptr
	pShowCursor                    uintptr
	pSetCursorPos                  uintptr
	pSetPhysicalCursorPos          uintptr
	pSetCursor                     uintptr
	pGetCursorPos                  uintptr
	pGetPhysicalCursorPos          uintptr
	pGetClipCursor                 uintptr
	pGetCursor                     uintptr
	pCreateCaret                   uintptr
	pGetCaretBlinkTime             uintptr
	pSetCaretBlinkTime             uintptr
	pDestroyCaret                  uintptr
	pHideCaret                     uintptr
	pShowCaret                     uintptr
	pSetCaretPos                   uintptr
	pGetCaretPos                   uintptr
	pLogicalToPhysicalPoint        uintptr
	pPhysicalToLogicalPoint        uintptr
	pWindowFromPoint               uintptr
	pWindowFromPhysicalPoint       uintptr
	pChildWindowFromPoint          uintptr
	pClipCursor                    uintptr
	pChildWindowFromPointEx        uintptr
	pGetSysColor                   uintptr
	pSetSysColors                  uintptr
	pGetWindowWord                 uintptr
	pSetWindowWord                 uintptr
	pGetWindowLongA                uintptr
	pGetWindowLongW                uintptr
	pSetWindowLongA                uintptr
	pSetWindowLongW                uintptr
	pGetWindowLongPtrA             uintptr
	pGetWindowLongPtrW             uintptr
	pSetWindowLongPtrA             uintptr
	pSetWindowLongPtrW             uintptr
	pGetClassWord                  uintptr
	pSetClassWord                  uintptr
	pGetClassLongA                 uintptr
	pGetClassLongW                 uintptr
	pSetClassLongA                 uintptr
	pSetClassLongW                 uintptr
	pGetClassLongPtrA              uintptr
	pGetClassLongPtrW              uintptr
	pSetClassLongPtrA              uintptr
	pSetClassLongPtrW              uintptr
	pGetProcessDefaultLayout       uintptr
	pSetProcessDefaultLayout       uintptr
	pGetDesktopWindow              uintptr
	pGetParent                     uintptr
	pSetParent                     uintptr
	pEnumChildWindows              uintptr
	pFindWindowA                   uintptr
	pFindWindowW                   uintptr
	pFindWindowExA                 uintptr
	pFindWindowExW                 uintptr
	pGetShellWindow                uintptr
	pRegisterShellHookWindow       uintptr
	pDeregisterShellHookWindow     uintptr
	pEnumWindows                   uintptr
	pEnumThreadWindows             uintptr
	pGetClassNameA                 uintptr
	pGetClassNameW                 uintptr
	pGetTopWindow                  uintptr
	pGetWindowThreadProcessId      uintptr
	pIsGUIThread                   uintptr
	pGetLastActivePopup            uintptr
	pGetWindow                     uintptr
	pSetWindowsHookA               uintptr
	pSetWindowsHookW               uintptr
	pUnhookWindowsHook             uintptr
	pSetWindowsHookExA             uintptr
	pSetWindowsHookExW             uintptr
	pUnhookWindowsHookEx           uintptr
	pCallNextHookEx                uintptr
	pCheckMenuRadioItem            uintptr
	pLoadCursorA                   uintptr
	pLoadCursorW                   uintptr
	pLoadCursorFromFileA           uintptr
	pLoadCursorFromFileW           uintptr
	pCreateCursor                  uintptr
	pDestroyCursor                 uintptr
	pSetSystemCursor               uintptr
	pLoadIconA                     uintptr
	pLoadIconW                     uintptr
	pPrivateExtractIconsA          uintptr
	pPrivateExtractIconsW          uintptr
	pCreateIcon                    uintptr
	pDestroyIcon                   uintptr
	pLookupIconIdFromDirectory     uintptr
	pLookupIconIdFromDirectoryEx   uintptr
	pCreateIconFromResource        uintptr
	pCreateIconFromResourceEx      uintptr
	pLoadImageA                    uintptr
	pLoadImageW                    uintptr
	pCopyImage                     uintptr
	pDrawIconEx                    uintptr
	pCreateIconIndirect            uintptr
	pCopyIcon                      uintptr
	pGetIconInfo                   uintptr
	pGetIconInfoExA                uintptr
	pGetIconInfoExW                uintptr
	pIsDialogMessageA              uintptr
	pIsDialogMessageW              uintptr
	pMapDialogRect                 uintptr
	pGetScrollInfo                 uintptr
	pDefFrameProcA                 uintptr
	pDefFrameProcW                 uintptr
	pDefMDIChildProcA              uintptr
	pDefMDIChildProcW              uintptr
	pTranslateMDISysAccel          uintptr
	pArrangeIconicWindows          uintptr
	pCreateMDIWindowA              uintptr
	pCreateMDIWindowW              uintptr
	pTileWindows                   uintptr
	pCascadeWindows                uintptr
	pSystemParametersInfoA         uintptr
	pSystemParametersInfoW         uintptr
	pSoundSentry                   uintptr
	pSetDebugErrorLevel            uintptr
	pInternalGetWindowText         uintptr
	pCancelShutdown                uintptr
	pGetGUIThreadInfo              uintptr
	pSetProcessDPIAware            uintptr
	pIsProcessDPIAware             uintptr
	pInheritWindowMonitor          uintptr
	pGetWindowModuleFileNameA      uintptr
	pGetWindowModuleFileNameW      uintptr
	pGetCursorInfo                 uintptr
	pGetWindowInfo                 uintptr
	pGetTitleBarInfo               uintptr
	pGetMenuBarInfo                uintptr
	pGetScrollBarInfo              uintptr
	pGetAncestor                   uintptr
	pRealChildWindowFromPoint      uintptr
	pRealGetWindowClassA           uintptr
	pRealGetWindowClassW           uintptr
	pGetAltTabInfoA                uintptr
	pGetAltTabInfoW                uintptr
	pChangeWindowMessageFilter     uintptr
	pChangeWindowMessageFilterEx   uintptr
)

func LoadStringA(hInstance HINSTANCE, uID uint32, lpBuffer PSTR, cchBufferMax int32) (int32, WIN32_ERROR) {
	addr := lazyAddr(&pLoadStringA, libUser32, "LoadStringA")
	ret, _, err := syscall.SyscallN(addr, hInstance, uintptr(uID), uintptr(unsafe.Pointer(lpBuffer)), uintptr(cchBufferMax))
	return int32(ret), WIN32_ERROR(err)
}

var LoadString = LoadStringW

func LoadStringW(hInstance HINSTANCE, uID uint32, lpBuffer PWSTR, cchBufferMax int32) (int32, WIN32_ERROR) {
	addr := lazyAddr(&pLoadStringW, libUser32, "LoadStringW")
	ret, _, err := syscall.SyscallN(addr, hInstance, uintptr(uID), uintptr(unsafe.Pointer(lpBuffer)), uintptr(cchBufferMax))
	return int32(ret), WIN32_ERROR(err)
}

func WvsprintfA(param0 PSTR, param1 PSTR, arglist *int8) (int32, WIN32_ERROR) {
	addr := lazyAddr(&pWvsprintfA, libUser32, "wvsprintfA")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(param0)), uintptr(unsafe.Pointer(param1)), uintptr(unsafe.Pointer(arglist)))
	return int32(ret), WIN32_ERROR(err)
}

var Wvsprintf = WvsprintfW

func WvsprintfW(param0 PWSTR, param1 PWSTR, arglist *int8) (int32, WIN32_ERROR) {
	addr := lazyAddr(&pWvsprintfW, libUser32, "wvsprintfW")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(param0)), uintptr(unsafe.Pointer(param1)), uintptr(unsafe.Pointer(arglist)))
	return int32(ret), WIN32_ERROR(err)
}

func WsprintfA(param0 PSTR, param1 PSTR) (int32, WIN32_ERROR) {
	addr := lazyAddr(&pWsprintfA, libUser32, "wsprintfA")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(param0)), uintptr(unsafe.Pointer(param1)))
	return int32(ret), WIN32_ERROR(err)
}

var Wsprintf = WsprintfW

func WsprintfW(param0 PWSTR, param1 PWSTR) (int32, WIN32_ERROR) {
	addr := lazyAddr(&pWsprintfW, libUser32, "wsprintfW")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(param0)), uintptr(unsafe.Pointer(param1)))
	return int32(ret), WIN32_ERROR(err)
}

func IsHungAppWindow(hwnd HWND) BOOL {
	addr := lazyAddr(&pIsHungAppWindow, libUser32, "IsHungAppWindow")
	ret, _, _ := syscall.SyscallN(addr, hwnd)
	return BOOL(ret)
}

func DisableProcessWindowsGhosting() {
	addr := lazyAddr(&pDisableProcessWindowsGhosting, libUser32, "DisableProcessWindowsGhosting")
	syscall.SyscallN(addr)
}

func RegisterWindowMessageA(lpString PSTR) (uint32, WIN32_ERROR) {
	addr := lazyAddr(&pRegisterWindowMessageA, libUser32, "RegisterWindowMessageA")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpString)))
	return uint32(ret), WIN32_ERROR(err)
}

var RegisterWindowMessage = RegisterWindowMessageW

func RegisterWindowMessageW(lpString PWSTR) (uint32, WIN32_ERROR) {
	addr := lazyAddr(&pRegisterWindowMessageW, libUser32, "RegisterWindowMessageW")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpString)))
	return uint32(ret), WIN32_ERROR(err)
}

func GetMessageA(lpMsg *MSG, hWnd HWND, wMsgFilterMin uint32, wMsgFilterMax uint32) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pGetMessageA, libUser32, "GetMessageA")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpMsg)), hWnd, uintptr(wMsgFilterMin), uintptr(wMsgFilterMax))
	return BOOL(ret), WIN32_ERROR(err)
}

var GetMessage = GetMessageW

func GetMessageW(lpMsg *MSG, hWnd HWND, wMsgFilterMin uint32, wMsgFilterMax uint32) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pGetMessageW, libUser32, "GetMessageW")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpMsg)), hWnd, uintptr(wMsgFilterMin), uintptr(wMsgFilterMax))
	return BOOL(ret), WIN32_ERROR(err)
}

func TranslateMessage(lpMsg *MSG) BOOL {
	addr := lazyAddr(&pTranslateMessage, libUser32, "TranslateMessage")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpMsg)))
	return BOOL(ret)
}

func DispatchMessageA(lpMsg *MSG) LRESULT {
	addr := lazyAddr(&pDispatchMessageA, libUser32, "DispatchMessageA")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpMsg)))
	return ret
}

var DispatchMessage = DispatchMessageW

func DispatchMessageW(lpMsg *MSG) LRESULT {
	addr := lazyAddr(&pDispatchMessageW, libUser32, "DispatchMessageW")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpMsg)))
	return ret
}

func SetMessageQueue(cMessagesMax int32) BOOL {
	addr := lazyAddr(&pSetMessageQueue, libUser32, "SetMessageQueue")
	ret, _, _ := syscall.SyscallN(addr, uintptr(cMessagesMax))
	return BOOL(ret)
}

func PeekMessageA(lpMsg *MSG, hWnd HWND, wMsgFilterMin uint32, wMsgFilterMax uint32, wRemoveMsg PEEK_MESSAGE_REMOVE_TYPE) BOOL {
	addr := lazyAddr(&pPeekMessageA, libUser32, "PeekMessageA")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpMsg)), hWnd, uintptr(wMsgFilterMin), uintptr(wMsgFilterMax), uintptr(wRemoveMsg))
	return BOOL(ret)
}

var PeekMessage = PeekMessageW

func PeekMessageW(lpMsg *MSG, hWnd HWND, wMsgFilterMin uint32, wMsgFilterMax uint32, wRemoveMsg PEEK_MESSAGE_REMOVE_TYPE) BOOL {
	addr := lazyAddr(&pPeekMessageW, libUser32, "PeekMessageW")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpMsg)), hWnd, uintptr(wMsgFilterMin), uintptr(wMsgFilterMax), uintptr(wRemoveMsg))
	return BOOL(ret)
}

func GetMessagePos() uint32 {
	addr := lazyAddr(&pGetMessagePos, libUser32, "GetMessagePos")
	ret, _, _ := syscall.SyscallN(addr)
	return uint32(ret)
}

func GetMessageTime() int32 {
	addr := lazyAddr(&pGetMessageTime, libUser32, "GetMessageTime")
	ret, _, _ := syscall.SyscallN(addr)
	return int32(ret)
}

func GetMessageExtraInfo() LPARAM {
	addr := lazyAddr(&pGetMessageExtraInfo, libUser32, "GetMessageExtraInfo")
	ret, _, _ := syscall.SyscallN(addr)
	return ret
}

func IsWow64Message() BOOL {
	addr := lazyAddr(&pIsWow64Message, libUser32, "IsWow64Message")
	ret, _, _ := syscall.SyscallN(addr)
	return BOOL(ret)
}

func SetMessageExtraInfo(lParam LPARAM) LPARAM {
	addr := lazyAddr(&pSetMessageExtraInfo, libUser32, "SetMessageExtraInfo")
	ret, _, _ := syscall.SyscallN(addr, lParam)
	return ret
}

func SendMessageA(hWnd HWND, Msg uint32, wParam WPARAM, lParam LPARAM) (LRESULT, WIN32_ERROR) {
	addr := lazyAddr(&pSendMessageA, libUser32, "SendMessageA")
	ret, _, err := syscall.SyscallN(addr, hWnd, uintptr(Msg), wParam, lParam)
	return ret, WIN32_ERROR(err)
}

var SendMessage = SendMessageW

func SendMessageW(hWnd HWND, Msg uint32, wParam WPARAM, lParam LPARAM) (LRESULT, WIN32_ERROR) {
	addr := lazyAddr(&pSendMessageW, libUser32, "SendMessageW")
	ret, _, err := syscall.SyscallN(addr, hWnd, uintptr(Msg), wParam, lParam)
	return ret, WIN32_ERROR(err)
}

func SendMessageTimeoutA(hWnd HWND, Msg uint32, wParam WPARAM, lParam LPARAM, fuFlags SEND_MESSAGE_TIMEOUT_FLAGS, uTimeout uint32, lpdwResult *uintptr) (LRESULT, WIN32_ERROR) {
	addr := lazyAddr(&pSendMessageTimeoutA, libUser32, "SendMessageTimeoutA")
	ret, _, err := syscall.SyscallN(addr, hWnd, uintptr(Msg), wParam, lParam, uintptr(fuFlags), uintptr(uTimeout), uintptr(unsafe.Pointer(lpdwResult)))
	return ret, WIN32_ERROR(err)
}

var SendMessageTimeout = SendMessageTimeoutW

func SendMessageTimeoutW(hWnd HWND, Msg uint32, wParam WPARAM, lParam LPARAM, fuFlags SEND_MESSAGE_TIMEOUT_FLAGS, uTimeout uint32, lpdwResult *uintptr) (LRESULT, WIN32_ERROR) {
	addr := lazyAddr(&pSendMessageTimeoutW, libUser32, "SendMessageTimeoutW")
	ret, _, err := syscall.SyscallN(addr, hWnd, uintptr(Msg), wParam, lParam, uintptr(fuFlags), uintptr(uTimeout), uintptr(unsafe.Pointer(lpdwResult)))
	return ret, WIN32_ERROR(err)
}

func SendNotifyMessageA(hWnd HWND, Msg uint32, wParam WPARAM, lParam LPARAM) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pSendNotifyMessageA, libUser32, "SendNotifyMessageA")
	ret, _, err := syscall.SyscallN(addr, hWnd, uintptr(Msg), wParam, lParam)
	return BOOL(ret), WIN32_ERROR(err)
}

var SendNotifyMessage = SendNotifyMessageW

func SendNotifyMessageW(hWnd HWND, Msg uint32, wParam WPARAM, lParam LPARAM) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pSendNotifyMessageW, libUser32, "SendNotifyMessageW")
	ret, _, err := syscall.SyscallN(addr, hWnd, uintptr(Msg), wParam, lParam)
	return BOOL(ret), WIN32_ERROR(err)
}

func SendMessageCallbackA(hWnd HWND, Msg uint32, wParam WPARAM, lParam LPARAM, lpResultCallBack SENDASYNCPROC, dwData uintptr) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pSendMessageCallbackA, libUser32, "SendMessageCallbackA")
	ret, _, err := syscall.SyscallN(addr, hWnd, uintptr(Msg), wParam, lParam, lpResultCallBack, dwData)
	return BOOL(ret), WIN32_ERROR(err)
}

var SendMessageCallback = SendMessageCallbackW

func SendMessageCallbackW(hWnd HWND, Msg uint32, wParam WPARAM, lParam LPARAM, lpResultCallBack SENDASYNCPROC, dwData uintptr) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pSendMessageCallbackW, libUser32, "SendMessageCallbackW")
	ret, _, err := syscall.SyscallN(addr, hWnd, uintptr(Msg), wParam, lParam, lpResultCallBack, dwData)
	return BOOL(ret), WIN32_ERROR(err)
}

func RegisterDeviceNotificationA(hRecipient HANDLE, NotificationFilter unsafe.Pointer, Flags POWER_SETTING_REGISTER_NOTIFICATION_FLAGS) (unsafe.Pointer, WIN32_ERROR) {
	addr := lazyAddr(&pRegisterDeviceNotificationA, libUser32, "RegisterDeviceNotificationA")
	ret, _, err := syscall.SyscallN(addr, hRecipient, uintptr(NotificationFilter), uintptr(Flags))
	return (unsafe.Pointer)(ret), WIN32_ERROR(err)
}

var RegisterDeviceNotification = RegisterDeviceNotificationW

func RegisterDeviceNotificationW(hRecipient HANDLE, NotificationFilter unsafe.Pointer, Flags POWER_SETTING_REGISTER_NOTIFICATION_FLAGS) (unsafe.Pointer, WIN32_ERROR) {
	addr := lazyAddr(&pRegisterDeviceNotificationW, libUser32, "RegisterDeviceNotificationW")
	ret, _, err := syscall.SyscallN(addr, hRecipient, uintptr(NotificationFilter), uintptr(Flags))
	return (unsafe.Pointer)(ret), WIN32_ERROR(err)
}

func PostMessageA(hWnd HWND, Msg uint32, wParam WPARAM, lParam LPARAM) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pPostMessageA, libUser32, "PostMessageA")
	ret, _, err := syscall.SyscallN(addr, hWnd, uintptr(Msg), wParam, lParam)
	return BOOL(ret), WIN32_ERROR(err)
}

var PostMessage = PostMessageW

func PostMessageW(hWnd HWND, Msg uint32, wParam WPARAM, lParam LPARAM) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pPostMessageW, libUser32, "PostMessageW")
	ret, _, err := syscall.SyscallN(addr, hWnd, uintptr(Msg), wParam, lParam)
	return BOOL(ret), WIN32_ERROR(err)
}

func PostThreadMessageA(idThread uint32, Msg uint32, wParam WPARAM, lParam LPARAM) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pPostThreadMessageA, libUser32, "PostThreadMessageA")
	ret, _, err := syscall.SyscallN(addr, uintptr(idThread), uintptr(Msg), wParam, lParam)
	return BOOL(ret), WIN32_ERROR(err)
}

var PostThreadMessage = PostThreadMessageW

func PostThreadMessageW(idThread uint32, Msg uint32, wParam WPARAM, lParam LPARAM) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pPostThreadMessageW, libUser32, "PostThreadMessageW")
	ret, _, err := syscall.SyscallN(addr, uintptr(idThread), uintptr(Msg), wParam, lParam)
	return BOOL(ret), WIN32_ERROR(err)
}

func ReplyMessage(lResult LRESULT) BOOL {
	addr := lazyAddr(&pReplyMessage, libUser32, "ReplyMessage")
	ret, _, _ := syscall.SyscallN(addr, lResult)
	return BOOL(ret)
}

func WaitMessage() (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pWaitMessage, libUser32, "WaitMessage")
	ret, _, err := syscall.SyscallN(addr)
	return BOOL(ret), WIN32_ERROR(err)
}

func DefWindowProcA(hWnd HWND, Msg uint32, wParam WPARAM, lParam LPARAM) LRESULT {
	addr := lazyAddr(&pDefWindowProcA, libUser32, "DefWindowProcA")
	ret, _, _ := syscall.SyscallN(addr, hWnd, uintptr(Msg), wParam, lParam)
	return ret
}

var DefWindowProc = DefWindowProcW

func DefWindowProcW(hWnd HWND, Msg uint32, wParam WPARAM, lParam LPARAM) LRESULT {
	addr := lazyAddr(&pDefWindowProcW, libUser32, "DefWindowProcW")
	ret, _, _ := syscall.SyscallN(addr, hWnd, uintptr(Msg), wParam, lParam)
	return ret
}

func PostQuitMessage(nExitCode int32) {
	addr := lazyAddr(&pPostQuitMessage, libUser32, "PostQuitMessage")
	syscall.SyscallN(addr, uintptr(nExitCode))
}

func CallWindowProcA(lpPrevWndFunc WNDPROC, hWnd HWND, Msg uint32, wParam WPARAM, lParam LPARAM) LRESULT {
	addr := lazyAddr(&pCallWindowProcA, libUser32, "CallWindowProcA")
	ret, _, _ := syscall.SyscallN(addr, lpPrevWndFunc, hWnd, uintptr(Msg), wParam, lParam)
	return ret
}

var CallWindowProc = CallWindowProcW

func CallWindowProcW(lpPrevWndFunc WNDPROC, hWnd HWND, Msg uint32, wParam WPARAM, lParam LPARAM) LRESULT {
	addr := lazyAddr(&pCallWindowProcW, libUser32, "CallWindowProcW")
	ret, _, _ := syscall.SyscallN(addr, lpPrevWndFunc, hWnd, uintptr(Msg), wParam, lParam)
	return ret
}

func InSendMessage() BOOL {
	addr := lazyAddr(&pInSendMessage, libUser32, "InSendMessage")
	ret, _, _ := syscall.SyscallN(addr)
	return BOOL(ret)
}

func InSendMessageEx(lpReserved unsafe.Pointer) uint32 {
	addr := lazyAddr(&pInSendMessageEx, libUser32, "InSendMessageEx")
	ret, _, _ := syscall.SyscallN(addr, uintptr(lpReserved))
	return uint32(ret)
}

func RegisterClassA(lpWndClass *WNDCLASSA) (uint16, WIN32_ERROR) {
	addr := lazyAddr(&pRegisterClassA, libUser32, "RegisterClassA")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpWndClass)))
	return uint16(ret), WIN32_ERROR(err)
}

var RegisterClass = RegisterClassW

func RegisterClassW(lpWndClass *WNDCLASSW) (uint16, WIN32_ERROR) {
	addr := lazyAddr(&pRegisterClassW, libUser32, "RegisterClassW")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpWndClass)))
	return uint16(ret), WIN32_ERROR(err)
}

func UnregisterClassA(lpClassName PSTR, hInstance HINSTANCE) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pUnregisterClassA, libUser32, "UnregisterClassA")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpClassName)), hInstance)
	return BOOL(ret), WIN32_ERROR(err)
}

var UnregisterClass = UnregisterClassW

func UnregisterClassW(lpClassName PWSTR, hInstance HINSTANCE) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pUnregisterClassW, libUser32, "UnregisterClassW")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpClassName)), hInstance)
	return BOOL(ret), WIN32_ERROR(err)
}

func GetClassInfoA(hInstance HINSTANCE, lpClassName PSTR, lpWndClass *WNDCLASSA) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pGetClassInfoA, libUser32, "GetClassInfoA")
	ret, _, err := syscall.SyscallN(addr, hInstance, uintptr(unsafe.Pointer(lpClassName)), uintptr(unsafe.Pointer(lpWndClass)))
	return BOOL(ret), WIN32_ERROR(err)
}

var GetClassInfo = GetClassInfoW

func GetClassInfoW(hInstance HINSTANCE, lpClassName PWSTR, lpWndClass *WNDCLASSW) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pGetClassInfoW, libUser32, "GetClassInfoW")
	ret, _, err := syscall.SyscallN(addr, hInstance, uintptr(unsafe.Pointer(lpClassName)), uintptr(unsafe.Pointer(lpWndClass)))
	return BOOL(ret), WIN32_ERROR(err)
}

func RegisterClassExA(param0 *WNDCLASSEXA) (uint16, WIN32_ERROR) {
	addr := lazyAddr(&pRegisterClassExA, libUser32, "RegisterClassExA")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(param0)))
	return uint16(ret), WIN32_ERROR(err)
}

var RegisterClassEx = RegisterClassExW

func RegisterClassExW(param0 *WNDCLASSEXW) (uint16, WIN32_ERROR) {
	addr := lazyAddr(&pRegisterClassExW, libUser32, "RegisterClassExW")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(param0)))
	return uint16(ret), WIN32_ERROR(err)
}

func GetClassInfoExA(hInstance HINSTANCE, lpszClass PSTR, lpwcx *WNDCLASSEXA) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pGetClassInfoExA, libUser32, "GetClassInfoExA")
	ret, _, err := syscall.SyscallN(addr, hInstance, uintptr(unsafe.Pointer(lpszClass)), uintptr(unsafe.Pointer(lpwcx)))
	return BOOL(ret), WIN32_ERROR(err)
}

var GetClassInfoEx = GetClassInfoExW

func GetClassInfoExW(hInstance HINSTANCE, lpszClass PWSTR, lpwcx *WNDCLASSEXW) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pGetClassInfoExW, libUser32, "GetClassInfoExW")
	ret, _, err := syscall.SyscallN(addr, hInstance, uintptr(unsafe.Pointer(lpszClass)), uintptr(unsafe.Pointer(lpwcx)))
	return BOOL(ret), WIN32_ERROR(err)
}

func CreateWindowExA(dwExStyle WINDOW_EX_STYLE, lpClassName PSTR, lpWindowName PSTR, dwStyle WINDOW_STYLE, X int32, Y int32, nWidth int32, nHeight int32, hWndParent HWND, hMenu HMENU, hInstance HINSTANCE, lpParam unsafe.Pointer) (HWND, WIN32_ERROR) {
	addr := lazyAddr(&pCreateWindowExA, libUser32, "CreateWindowExA")
	ret, _, err := syscall.SyscallN(addr, uintptr(dwExStyle), uintptr(unsafe.Pointer(lpClassName)), uintptr(unsafe.Pointer(lpWindowName)), uintptr(dwStyle), uintptr(X), uintptr(Y), uintptr(nWidth), uintptr(nHeight), hWndParent, hMenu, hInstance, uintptr(lpParam))
	return ret, WIN32_ERROR(err)
}

var CreateWindowEx = CreateWindowExW

func CreateWindowExW(dwExStyle WINDOW_EX_STYLE, lpClassName PWSTR, lpWindowName PWSTR, dwStyle WINDOW_STYLE, X int32, Y int32, nWidth int32, nHeight int32, hWndParent HWND, hMenu HMENU, hInstance HINSTANCE, lpParam unsafe.Pointer) (HWND, WIN32_ERROR) {
	addr := lazyAddr(&pCreateWindowExW, libUser32, "CreateWindowExW")
	ret, _, err := syscall.SyscallN(addr, uintptr(dwExStyle), uintptr(unsafe.Pointer(lpClassName)), uintptr(unsafe.Pointer(lpWindowName)), uintptr(dwStyle), uintptr(X), uintptr(Y), uintptr(nWidth), uintptr(nHeight), hWndParent, hMenu, hInstance, uintptr(lpParam))
	return ret, WIN32_ERROR(err)
}

func IsWindow(hWnd HWND) BOOL {
	addr := lazyAddr(&pIsWindow, libUser32, "IsWindow")
	ret, _, _ := syscall.SyscallN(addr, hWnd)
	return BOOL(ret)
}

func IsMenu(hMenu HMENU) BOOL {
	addr := lazyAddr(&pIsMenu, libUser32, "IsMenu")
	ret, _, _ := syscall.SyscallN(addr, hMenu)
	return BOOL(ret)
}

func IsChild(hWndParent HWND, hWnd HWND) BOOL {
	addr := lazyAddr(&pIsChild, libUser32, "IsChild")
	ret, _, _ := syscall.SyscallN(addr, hWndParent, hWnd)
	return BOOL(ret)
}

func DestroyWindow(hWnd HWND) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pDestroyWindow, libUser32, "DestroyWindow")
	ret, _, err := syscall.SyscallN(addr, hWnd)
	return BOOL(ret), WIN32_ERROR(err)
}

func ShowWindow(hWnd HWND, nCmdShow SHOW_WINDOW_CMD) BOOL {
	addr := lazyAddr(&pShowWindow, libUser32, "ShowWindow")
	ret, _, _ := syscall.SyscallN(addr, hWnd, uintptr(nCmdShow))
	return BOOL(ret)
}

func AnimateWindow(hWnd HWND, dwTime uint32, dwFlags ANIMATE_WINDOW_FLAGS) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pAnimateWindow, libUser32, "AnimateWindow")
	ret, _, err := syscall.SyscallN(addr, hWnd, uintptr(dwTime), uintptr(dwFlags))
	return BOOL(ret), WIN32_ERROR(err)
}

func UpdateLayeredWindow(hWnd HWND, hdcDst HDC, pptDst *POINT, psize *SIZE, hdcSrc HDC, pptSrc *POINT, crKey uint32, pblend *BLENDFUNCTION, dwFlags UPDATE_LAYERED_WINDOW_FLAGS) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pUpdateLayeredWindow, libUser32, "UpdateLayeredWindow")
	ret, _, err := syscall.SyscallN(addr, hWnd, hdcDst, uintptr(unsafe.Pointer(pptDst)), uintptr(unsafe.Pointer(psize)), hdcSrc, uintptr(unsafe.Pointer(pptSrc)), uintptr(crKey), uintptr(unsafe.Pointer(pblend)), uintptr(dwFlags))
	return BOOL(ret), WIN32_ERROR(err)
}

func UpdateLayeredWindowIndirect(hWnd HWND, pULWInfo *UPDATELAYEREDWINDOWINFO) BOOL {
	addr := lazyAddr(&pUpdateLayeredWindowIndirect, libUser32, "UpdateLayeredWindowIndirect")
	ret, _, _ := syscall.SyscallN(addr, hWnd, uintptr(unsafe.Pointer(pULWInfo)))
	return BOOL(ret)
}

func GetLayeredWindowAttributes(hwnd HWND, pcrKey *uint32, pbAlpha *byte, pdwFlags *LAYERED_WINDOW_ATTRIBUTES_FLAGS) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pGetLayeredWindowAttributes, libUser32, "GetLayeredWindowAttributes")
	ret, _, err := syscall.SyscallN(addr, hwnd, uintptr(unsafe.Pointer(pcrKey)), uintptr(unsafe.Pointer(pbAlpha)), uintptr(unsafe.Pointer(pdwFlags)))
	return BOOL(ret), WIN32_ERROR(err)
}

func SetLayeredWindowAttributes(hwnd HWND, crKey uint32, bAlpha byte, dwFlags LAYERED_WINDOW_ATTRIBUTES_FLAGS) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pSetLayeredWindowAttributes, libUser32, "SetLayeredWindowAttributes")
	ret, _, err := syscall.SyscallN(addr, hwnd, uintptr(crKey), uintptr(bAlpha), uintptr(dwFlags))
	return BOOL(ret), WIN32_ERROR(err)
}

func ShowWindowAsync(hWnd HWND, nCmdShow SHOW_WINDOW_CMD) BOOL {
	addr := lazyAddr(&pShowWindowAsync, libUser32, "ShowWindowAsync")
	ret, _, _ := syscall.SyscallN(addr, hWnd, uintptr(nCmdShow))
	return BOOL(ret)
}

func FlashWindow(hWnd HWND, bInvert BOOL) BOOL {
	addr := lazyAddr(&pFlashWindow, libUser32, "FlashWindow")
	ret, _, _ := syscall.SyscallN(addr, hWnd, uintptr(bInvert))
	return BOOL(ret)
}

func FlashWindowEx(pfwi *FLASHWINFO) BOOL {
	addr := lazyAddr(&pFlashWindowEx, libUser32, "FlashWindowEx")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pfwi)))
	return BOOL(ret)
}

func ShowOwnedPopups(hWnd HWND, fShow BOOL) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pShowOwnedPopups, libUser32, "ShowOwnedPopups")
	ret, _, err := syscall.SyscallN(addr, hWnd, uintptr(fShow))
	return BOOL(ret), WIN32_ERROR(err)
}

func OpenIcon(hWnd HWND) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pOpenIcon, libUser32, "OpenIcon")
	ret, _, err := syscall.SyscallN(addr, hWnd)
	return BOOL(ret), WIN32_ERROR(err)
}

func CloseWindow(hWnd HWND) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pCloseWindow, libUser32, "CloseWindow")
	ret, _, err := syscall.SyscallN(addr, hWnd)
	return BOOL(ret), WIN32_ERROR(err)
}

func MoveWindow(hWnd HWND, X int32, Y int32, nWidth int32, nHeight int32, bRepaint BOOL) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pMoveWindow, libUser32, "MoveWindow")
	ret, _, err := syscall.SyscallN(addr, hWnd, uintptr(X), uintptr(Y), uintptr(nWidth), uintptr(nHeight), uintptr(bRepaint))
	return BOOL(ret), WIN32_ERROR(err)
}

func SetWindowPos(hWnd HWND, hWndInsertAfter HWND, X int32, Y int32, cx int32, cy int32, uFlags SET_WINDOW_POS_FLAGS) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pSetWindowPos, libUser32, "SetWindowPos")
	ret, _, err := syscall.SyscallN(addr, hWnd, hWndInsertAfter, uintptr(X), uintptr(Y), uintptr(cx), uintptr(cy), uintptr(uFlags))
	return BOOL(ret), WIN32_ERROR(err)
}

func GetWindowPlacement(hWnd HWND, lpwndpl *WINDOWPLACEMENT) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pGetWindowPlacement, libUser32, "GetWindowPlacement")
	ret, _, err := syscall.SyscallN(addr, hWnd, uintptr(unsafe.Pointer(lpwndpl)))
	return BOOL(ret), WIN32_ERROR(err)
}

func SetWindowPlacement(hWnd HWND, lpwndpl *WINDOWPLACEMENT) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pSetWindowPlacement, libUser32, "SetWindowPlacement")
	ret, _, err := syscall.SyscallN(addr, hWnd, uintptr(unsafe.Pointer(lpwndpl)))
	return BOOL(ret), WIN32_ERROR(err)
}

func GetWindowDisplayAffinity(hWnd HWND, pdwAffinity *uint32) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pGetWindowDisplayAffinity, libUser32, "GetWindowDisplayAffinity")
	ret, _, err := syscall.SyscallN(addr, hWnd, uintptr(unsafe.Pointer(pdwAffinity)))
	return BOOL(ret), WIN32_ERROR(err)
}

func SetWindowDisplayAffinity(hWnd HWND, dwAffinity WINDOW_DISPLAY_AFFINITY) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pSetWindowDisplayAffinity, libUser32, "SetWindowDisplayAffinity")
	ret, _, err := syscall.SyscallN(addr, hWnd, uintptr(dwAffinity))
	return BOOL(ret), WIN32_ERROR(err)
}

func BeginDeferWindowPos(nNumWindows int32) (uintptr, WIN32_ERROR) {
	addr := lazyAddr(&pBeginDeferWindowPos, libUser32, "BeginDeferWindowPos")
	ret, _, err := syscall.SyscallN(addr, uintptr(nNumWindows))
	return ret, WIN32_ERROR(err)
}

func DeferWindowPos(hWinPosInfo uintptr, hWnd HWND, hWndInsertAfter HWND, x int32, y int32, cx int32, cy int32, uFlags SET_WINDOW_POS_FLAGS) (uintptr, WIN32_ERROR) {
	addr := lazyAddr(&pDeferWindowPos, libUser32, "DeferWindowPos")
	ret, _, err := syscall.SyscallN(addr, hWinPosInfo, hWnd, hWndInsertAfter, uintptr(x), uintptr(y), uintptr(cx), uintptr(cy), uintptr(uFlags))
	return ret, WIN32_ERROR(err)
}

func EndDeferWindowPos(hWinPosInfo uintptr) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pEndDeferWindowPos, libUser32, "EndDeferWindowPos")
	ret, _, err := syscall.SyscallN(addr, hWinPosInfo)
	return BOOL(ret), WIN32_ERROR(err)
}

func IsWindowVisible(hWnd HWND) BOOL {
	addr := lazyAddr(&pIsWindowVisible, libUser32, "IsWindowVisible")
	ret, _, _ := syscall.SyscallN(addr, hWnd)
	return BOOL(ret)
}

func IsIconic(hWnd HWND) BOOL {
	addr := lazyAddr(&pIsIconic, libUser32, "IsIconic")
	ret, _, _ := syscall.SyscallN(addr, hWnd)
	return BOOL(ret)
}

func AnyPopup() BOOL {
	addr := lazyAddr(&pAnyPopup, libUser32, "AnyPopup")
	ret, _, _ := syscall.SyscallN(addr)
	return BOOL(ret)
}

func BringWindowToTop(hWnd HWND) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pBringWindowToTop, libUser32, "BringWindowToTop")
	ret, _, err := syscall.SyscallN(addr, hWnd)
	return BOOL(ret), WIN32_ERROR(err)
}

func IsZoomed(hWnd HWND) BOOL {
	addr := lazyAddr(&pIsZoomed, libUser32, "IsZoomed")
	ret, _, _ := syscall.SyscallN(addr, hWnd)
	return BOOL(ret)
}

func CreateDialogParamA(hInstance HINSTANCE, lpTemplateName PSTR, hWndParent HWND, lpDialogFunc DLGPROC, dwInitParam LPARAM) (HWND, WIN32_ERROR) {
	addr := lazyAddr(&pCreateDialogParamA, libUser32, "CreateDialogParamA")
	ret, _, err := syscall.SyscallN(addr, hInstance, uintptr(unsafe.Pointer(lpTemplateName)), hWndParent, lpDialogFunc, dwInitParam)
	return ret, WIN32_ERROR(err)
}

var CreateDialogParam = CreateDialogParamW

func CreateDialogParamW(hInstance HINSTANCE, lpTemplateName PWSTR, hWndParent HWND, lpDialogFunc DLGPROC, dwInitParam LPARAM) (HWND, WIN32_ERROR) {
	addr := lazyAddr(&pCreateDialogParamW, libUser32, "CreateDialogParamW")
	ret, _, err := syscall.SyscallN(addr, hInstance, uintptr(unsafe.Pointer(lpTemplateName)), hWndParent, lpDialogFunc, dwInitParam)
	return ret, WIN32_ERROR(err)
}

func CreateDialogIndirectParamA(hInstance HINSTANCE, lpTemplate *DLGTEMPLATE, hWndParent HWND, lpDialogFunc DLGPROC, dwInitParam LPARAM) (HWND, WIN32_ERROR) {
	addr := lazyAddr(&pCreateDialogIndirectParamA, libUser32, "CreateDialogIndirectParamA")
	ret, _, err := syscall.SyscallN(addr, hInstance, uintptr(unsafe.Pointer(lpTemplate)), hWndParent, lpDialogFunc, dwInitParam)
	return ret, WIN32_ERROR(err)
}

var CreateDialogIndirectParam = CreateDialogIndirectParamW

func CreateDialogIndirectParamW(hInstance HINSTANCE, lpTemplate *DLGTEMPLATE, hWndParent HWND, lpDialogFunc DLGPROC, dwInitParam LPARAM) (HWND, WIN32_ERROR) {
	addr := lazyAddr(&pCreateDialogIndirectParamW, libUser32, "CreateDialogIndirectParamW")
	ret, _, err := syscall.SyscallN(addr, hInstance, uintptr(unsafe.Pointer(lpTemplate)), hWndParent, lpDialogFunc, dwInitParam)
	return ret, WIN32_ERROR(err)
}

func DialogBoxParamA(hInstance HINSTANCE, lpTemplateName PSTR, hWndParent HWND, lpDialogFunc DLGPROC, dwInitParam LPARAM) (uintptr, WIN32_ERROR) {
	addr := lazyAddr(&pDialogBoxParamA, libUser32, "DialogBoxParamA")
	ret, _, err := syscall.SyscallN(addr, hInstance, uintptr(unsafe.Pointer(lpTemplateName)), hWndParent, lpDialogFunc, dwInitParam)
	return ret, WIN32_ERROR(err)
}

var DialogBoxParam = DialogBoxParamW

func DialogBoxParamW(hInstance HINSTANCE, lpTemplateName PWSTR, hWndParent HWND, lpDialogFunc DLGPROC, dwInitParam LPARAM) (uintptr, WIN32_ERROR) {
	addr := lazyAddr(&pDialogBoxParamW, libUser32, "DialogBoxParamW")
	ret, _, err := syscall.SyscallN(addr, hInstance, uintptr(unsafe.Pointer(lpTemplateName)), hWndParent, lpDialogFunc, dwInitParam)
	return ret, WIN32_ERROR(err)
}

func DialogBoxIndirectParamA(hInstance HINSTANCE, hDialogTemplate *DLGTEMPLATE, hWndParent HWND, lpDialogFunc DLGPROC, dwInitParam LPARAM) (uintptr, WIN32_ERROR) {
	addr := lazyAddr(&pDialogBoxIndirectParamA, libUser32, "DialogBoxIndirectParamA")
	ret, _, err := syscall.SyscallN(addr, hInstance, uintptr(unsafe.Pointer(hDialogTemplate)), hWndParent, lpDialogFunc, dwInitParam)
	return ret, WIN32_ERROR(err)
}

var DialogBoxIndirectParam = DialogBoxIndirectParamW

func DialogBoxIndirectParamW(hInstance HINSTANCE, hDialogTemplate *DLGTEMPLATE, hWndParent HWND, lpDialogFunc DLGPROC, dwInitParam LPARAM) (uintptr, WIN32_ERROR) {
	addr := lazyAddr(&pDialogBoxIndirectParamW, libUser32, "DialogBoxIndirectParamW")
	ret, _, err := syscall.SyscallN(addr, hInstance, uintptr(unsafe.Pointer(hDialogTemplate)), hWndParent, lpDialogFunc, dwInitParam)
	return ret, WIN32_ERROR(err)
}

func EndDialog(hDlg HWND, nResult uintptr) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pEndDialog, libUser32, "EndDialog")
	ret, _, err := syscall.SyscallN(addr, hDlg, nResult)
	return BOOL(ret), WIN32_ERROR(err)
}

func GetDlgItem(hDlg HWND, nIDDlgItem int32) (HWND, WIN32_ERROR) {
	addr := lazyAddr(&pGetDlgItem, libUser32, "GetDlgItem")
	ret, _, err := syscall.SyscallN(addr, hDlg, uintptr(nIDDlgItem))
	return ret, WIN32_ERROR(err)
}

func SetDlgItemInt(hDlg HWND, nIDDlgItem int32, uValue uint32, bSigned BOOL) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pSetDlgItemInt, libUser32, "SetDlgItemInt")
	ret, _, err := syscall.SyscallN(addr, hDlg, uintptr(nIDDlgItem), uintptr(uValue), uintptr(bSigned))
	return BOOL(ret), WIN32_ERROR(err)
}

func GetDlgItemInt(hDlg HWND, nIDDlgItem int32, lpTranslated *BOOL, bSigned BOOL) (uint32, WIN32_ERROR) {
	addr := lazyAddr(&pGetDlgItemInt, libUser32, "GetDlgItemInt")
	ret, _, err := syscall.SyscallN(addr, hDlg, uintptr(nIDDlgItem), uintptr(unsafe.Pointer(lpTranslated)), uintptr(bSigned))
	return uint32(ret), WIN32_ERROR(err)
}

func SetDlgItemTextA(hDlg HWND, nIDDlgItem int32, lpString PSTR) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pSetDlgItemTextA, libUser32, "SetDlgItemTextA")
	ret, _, err := syscall.SyscallN(addr, hDlg, uintptr(nIDDlgItem), uintptr(unsafe.Pointer(lpString)))
	return BOOL(ret), WIN32_ERROR(err)
}

var SetDlgItemText = SetDlgItemTextW

func SetDlgItemTextW(hDlg HWND, nIDDlgItem int32, lpString PWSTR) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pSetDlgItemTextW, libUser32, "SetDlgItemTextW")
	ret, _, err := syscall.SyscallN(addr, hDlg, uintptr(nIDDlgItem), uintptr(unsafe.Pointer(lpString)))
	return BOOL(ret), WIN32_ERROR(err)
}

func GetDlgItemTextA(hDlg HWND, nIDDlgItem int32, lpString PSTR, cchMax int32) (uint32, WIN32_ERROR) {
	addr := lazyAddr(&pGetDlgItemTextA, libUser32, "GetDlgItemTextA")
	ret, _, err := syscall.SyscallN(addr, hDlg, uintptr(nIDDlgItem), uintptr(unsafe.Pointer(lpString)), uintptr(cchMax))
	return uint32(ret), WIN32_ERROR(err)
}

var GetDlgItemText = GetDlgItemTextW

func GetDlgItemTextW(hDlg HWND, nIDDlgItem int32, lpString PWSTR, cchMax int32) (uint32, WIN32_ERROR) {
	addr := lazyAddr(&pGetDlgItemTextW, libUser32, "GetDlgItemTextW")
	ret, _, err := syscall.SyscallN(addr, hDlg, uintptr(nIDDlgItem), uintptr(unsafe.Pointer(lpString)), uintptr(cchMax))
	return uint32(ret), WIN32_ERROR(err)
}

func SendDlgItemMessageA(hDlg HWND, nIDDlgItem int32, Msg uint32, wParam WPARAM, lParam LPARAM) LRESULT {
	addr := lazyAddr(&pSendDlgItemMessageA, libUser32, "SendDlgItemMessageA")
	ret, _, _ := syscall.SyscallN(addr, hDlg, uintptr(nIDDlgItem), uintptr(Msg), wParam, lParam)
	return ret
}

var SendDlgItemMessage = SendDlgItemMessageW

func SendDlgItemMessageW(hDlg HWND, nIDDlgItem int32, Msg uint32, wParam WPARAM, lParam LPARAM) LRESULT {
	addr := lazyAddr(&pSendDlgItemMessageW, libUser32, "SendDlgItemMessageW")
	ret, _, _ := syscall.SyscallN(addr, hDlg, uintptr(nIDDlgItem), uintptr(Msg), wParam, lParam)
	return ret
}

func GetNextDlgGroupItem(hDlg HWND, hCtl HWND, bPrevious BOOL) (HWND, WIN32_ERROR) {
	addr := lazyAddr(&pGetNextDlgGroupItem, libUser32, "GetNextDlgGroupItem")
	ret, _, err := syscall.SyscallN(addr, hDlg, hCtl, uintptr(bPrevious))
	return ret, WIN32_ERROR(err)
}

func GetNextDlgTabItem(hDlg HWND, hCtl HWND, bPrevious BOOL) (HWND, WIN32_ERROR) {
	addr := lazyAddr(&pGetNextDlgTabItem, libUser32, "GetNextDlgTabItem")
	ret, _, err := syscall.SyscallN(addr, hDlg, hCtl, uintptr(bPrevious))
	return ret, WIN32_ERROR(err)
}

func GetDlgCtrlID(hWnd HWND) (int32, WIN32_ERROR) {
	addr := lazyAddr(&pGetDlgCtrlID, libUser32, "GetDlgCtrlID")
	ret, _, err := syscall.SyscallN(addr, hWnd)
	return int32(ret), WIN32_ERROR(err)
}

func GetDialogBaseUnits() int32 {
	addr := lazyAddr(&pGetDialogBaseUnits, libUser32, "GetDialogBaseUnits")
	ret, _, _ := syscall.SyscallN(addr)
	return int32(ret)
}

func DefDlgProcA(hDlg HWND, Msg uint32, wParam WPARAM, lParam LPARAM) LRESULT {
	addr := lazyAddr(&pDefDlgProcA, libUser32, "DefDlgProcA")
	ret, _, _ := syscall.SyscallN(addr, hDlg, uintptr(Msg), wParam, lParam)
	return ret
}

var DefDlgProc = DefDlgProcW

func DefDlgProcW(hDlg HWND, Msg uint32, wParam WPARAM, lParam LPARAM) LRESULT {
	addr := lazyAddr(&pDefDlgProcW, libUser32, "DefDlgProcW")
	ret, _, _ := syscall.SyscallN(addr, hDlg, uintptr(Msg), wParam, lParam)
	return ret
}

func CallMsgFilterA(lpMsg *MSG, nCode int32) BOOL {
	addr := lazyAddr(&pCallMsgFilterA, libUser32, "CallMsgFilterA")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpMsg)), uintptr(nCode))
	return BOOL(ret)
}

var CallMsgFilter = CallMsgFilterW

func CallMsgFilterW(lpMsg *MSG, nCode int32) BOOL {
	addr := lazyAddr(&pCallMsgFilterW, libUser32, "CallMsgFilterW")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpMsg)), uintptr(nCode))
	return BOOL(ret)
}

func CharToOemA(pSrc PSTR, pDst PSTR) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pCharToOemA, libUser32, "CharToOemA")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pSrc)), uintptr(unsafe.Pointer(pDst)))
	return BOOL(ret), WIN32_ERROR(err)
}

var CharToOem = CharToOemW

func CharToOemW(pSrc PWSTR, pDst PSTR) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pCharToOemW, libUser32, "CharToOemW")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pSrc)), uintptr(unsafe.Pointer(pDst)))
	return BOOL(ret), WIN32_ERROR(err)
}

func OemToCharA(pSrc PSTR, pDst PSTR) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pOemToCharA, libUser32, "OemToCharA")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pSrc)), uintptr(unsafe.Pointer(pDst)))
	return BOOL(ret), WIN32_ERROR(err)
}

var OemToChar = OemToCharW

func OemToCharW(pSrc PSTR, pDst PWSTR) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pOemToCharW, libUser32, "OemToCharW")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pSrc)), uintptr(unsafe.Pointer(pDst)))
	return BOOL(ret), WIN32_ERROR(err)
}

func CharToOemBuffA(lpszSrc PSTR, lpszDst PSTR, cchDstLength uint32) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pCharToOemBuffA, libUser32, "CharToOemBuffA")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpszSrc)), uintptr(unsafe.Pointer(lpszDst)), uintptr(cchDstLength))
	return BOOL(ret), WIN32_ERROR(err)
}

var CharToOemBuff = CharToOemBuffW

func CharToOemBuffW(lpszSrc PWSTR, lpszDst PSTR, cchDstLength uint32) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pCharToOemBuffW, libUser32, "CharToOemBuffW")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpszSrc)), uintptr(unsafe.Pointer(lpszDst)), uintptr(cchDstLength))
	return BOOL(ret), WIN32_ERROR(err)
}

func OemToCharBuffA(lpszSrc PSTR, lpszDst PSTR, cchDstLength uint32) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pOemToCharBuffA, libUser32, "OemToCharBuffA")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpszSrc)), uintptr(unsafe.Pointer(lpszDst)), uintptr(cchDstLength))
	return BOOL(ret), WIN32_ERROR(err)
}

var OemToCharBuff = OemToCharBuffW

func OemToCharBuffW(lpszSrc PSTR, lpszDst PWSTR, cchDstLength uint32) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pOemToCharBuffW, libUser32, "OemToCharBuffW")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpszSrc)), uintptr(unsafe.Pointer(lpszDst)), uintptr(cchDstLength))
	return BOOL(ret), WIN32_ERROR(err)
}

func CharUpperA(lpsz PSTR) (PSTR, WIN32_ERROR) {
	addr := lazyAddr(&pCharUpperA, libUser32, "CharUpperA")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpsz)))
	return (PSTR)(unsafe.Pointer(ret)), WIN32_ERROR(err)
}

var CharUpper = CharUpperW

func CharUpperW(lpsz PWSTR) (PWSTR, WIN32_ERROR) {
	addr := lazyAddr(&pCharUpperW, libUser32, "CharUpperW")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpsz)))
	return (PWSTR)(unsafe.Pointer(ret)), WIN32_ERROR(err)
}

func CharUpperBuffA(lpsz PSTR, cchLength uint32) uint32 {
	addr := lazyAddr(&pCharUpperBuffA, libUser32, "CharUpperBuffA")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpsz)), uintptr(cchLength))
	return uint32(ret)
}

var CharUpperBuff = CharUpperBuffW

func CharUpperBuffW(lpsz PWSTR, cchLength uint32) uint32 {
	addr := lazyAddr(&pCharUpperBuffW, libUser32, "CharUpperBuffW")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpsz)), uintptr(cchLength))
	return uint32(ret)
}

func CharLowerA(lpsz PSTR) (PSTR, WIN32_ERROR) {
	addr := lazyAddr(&pCharLowerA, libUser32, "CharLowerA")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpsz)))
	return (PSTR)(unsafe.Pointer(ret)), WIN32_ERROR(err)
}

var CharLower = CharLowerW

func CharLowerW(lpsz PWSTR) (PWSTR, WIN32_ERROR) {
	addr := lazyAddr(&pCharLowerW, libUser32, "CharLowerW")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpsz)))
	return (PWSTR)(unsafe.Pointer(ret)), WIN32_ERROR(err)
}

func CharLowerBuffA(lpsz PSTR, cchLength uint32) uint32 {
	addr := lazyAddr(&pCharLowerBuffA, libUser32, "CharLowerBuffA")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpsz)), uintptr(cchLength))
	return uint32(ret)
}

var CharLowerBuff = CharLowerBuffW

func CharLowerBuffW(lpsz PWSTR, cchLength uint32) uint32 {
	addr := lazyAddr(&pCharLowerBuffW, libUser32, "CharLowerBuffW")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpsz)), uintptr(cchLength))
	return uint32(ret)
}

func CharNextA(lpsz PSTR) PSTR {
	addr := lazyAddr(&pCharNextA, libUser32, "CharNextA")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpsz)))
	return (PSTR)(unsafe.Pointer(ret))
}

var CharNext = CharNextW

func CharNextW(lpsz PWSTR) PWSTR {
	addr := lazyAddr(&pCharNextW, libUser32, "CharNextW")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpsz)))
	return (PWSTR)(unsafe.Pointer(ret))
}

func CharPrevA(lpszStart PSTR, lpszCurrent PSTR) PSTR {
	addr := lazyAddr(&pCharPrevA, libUser32, "CharPrevA")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpszStart)), uintptr(unsafe.Pointer(lpszCurrent)))
	return (PSTR)(unsafe.Pointer(ret))
}

var CharPrev = CharPrevW

func CharPrevW(lpszStart PWSTR, lpszCurrent PWSTR) PWSTR {
	addr := lazyAddr(&pCharPrevW, libUser32, "CharPrevW")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpszStart)), uintptr(unsafe.Pointer(lpszCurrent)))
	return (PWSTR)(unsafe.Pointer(ret))
}

func CharNextExA(CodePage uint16, lpCurrentChar PSTR, dwFlags uint32) PSTR {
	addr := lazyAddr(&pCharNextExA, libUser32, "CharNextExA")
	ret, _, _ := syscall.SyscallN(addr, uintptr(CodePage), uintptr(unsafe.Pointer(lpCurrentChar)), uintptr(dwFlags))
	return (PSTR)(unsafe.Pointer(ret))
}

func CharPrevExA(CodePage uint16, lpStart PSTR, lpCurrentChar PSTR, dwFlags uint32) PSTR {
	addr := lazyAddr(&pCharPrevExA, libUser32, "CharPrevExA")
	ret, _, _ := syscall.SyscallN(addr, uintptr(CodePage), uintptr(unsafe.Pointer(lpStart)), uintptr(unsafe.Pointer(lpCurrentChar)), uintptr(dwFlags))
	return (PSTR)(unsafe.Pointer(ret))
}

func IsCharAlphaA(ch CHAR) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pIsCharAlphaA, libUser32, "IsCharAlphaA")
	ret, _, err := syscall.SyscallN(addr, uintptr(ch))
	return BOOL(ret), WIN32_ERROR(err)
}

var IsCharAlpha = IsCharAlphaW

func IsCharAlphaW(ch uint16) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pIsCharAlphaW, libUser32, "IsCharAlphaW")
	ret, _, err := syscall.SyscallN(addr, uintptr(ch))
	return BOOL(ret), WIN32_ERROR(err)
}

func IsCharAlphaNumericA(ch CHAR) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pIsCharAlphaNumericA, libUser32, "IsCharAlphaNumericA")
	ret, _, err := syscall.SyscallN(addr, uintptr(ch))
	return BOOL(ret), WIN32_ERROR(err)
}

var IsCharAlphaNumeric = IsCharAlphaNumericW

func IsCharAlphaNumericW(ch uint16) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pIsCharAlphaNumericW, libUser32, "IsCharAlphaNumericW")
	ret, _, err := syscall.SyscallN(addr, uintptr(ch))
	return BOOL(ret), WIN32_ERROR(err)
}

func IsCharUpperA(ch CHAR) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pIsCharUpperA, libUser32, "IsCharUpperA")
	ret, _, err := syscall.SyscallN(addr, uintptr(ch))
	return BOOL(ret), WIN32_ERROR(err)
}

var IsCharUpper = IsCharUpperW

func IsCharUpperW(ch uint16) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pIsCharUpperW, libUser32, "IsCharUpperW")
	ret, _, err := syscall.SyscallN(addr, uintptr(ch))
	return BOOL(ret), WIN32_ERROR(err)
}

func IsCharLowerA(ch CHAR) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pIsCharLowerA, libUser32, "IsCharLowerA")
	ret, _, err := syscall.SyscallN(addr, uintptr(ch))
	return BOOL(ret), WIN32_ERROR(err)
}

func GetInputState() BOOL {
	addr := lazyAddr(&pGetInputState, libUser32, "GetInputState")
	ret, _, _ := syscall.SyscallN(addr)
	return BOOL(ret)
}

func GetQueueStatus(flags QUEUE_STATUS_FLAGS) uint32 {
	addr := lazyAddr(&pGetQueueStatus, libUser32, "GetQueueStatus")
	ret, _, _ := syscall.SyscallN(addr, uintptr(flags))
	return uint32(ret)
}

func MsgWaitForMultipleObjects(nCount uint32, pHandles *HANDLE, fWaitAll BOOL, dwMilliseconds uint32, dwWakeMask QUEUE_STATUS_FLAGS) (uint32, WIN32_ERROR) {
	addr := lazyAddr(&pMsgWaitForMultipleObjects, libUser32, "MsgWaitForMultipleObjects")
	ret, _, err := syscall.SyscallN(addr, uintptr(nCount), uintptr(unsafe.Pointer(pHandles)), uintptr(fWaitAll), uintptr(dwMilliseconds), uintptr(dwWakeMask))
	return uint32(ret), WIN32_ERROR(err)
}

func MsgWaitForMultipleObjectsEx(nCount uint32, pHandles *HANDLE, dwMilliseconds uint32, dwWakeMask QUEUE_STATUS_FLAGS, dwFlags MSG_WAIT_FOR_MULTIPLE_OBJECTS_EX_FLAGS) (uint32, WIN32_ERROR) {
	addr := lazyAddr(&pMsgWaitForMultipleObjectsEx, libUser32, "MsgWaitForMultipleObjectsEx")
	ret, _, err := syscall.SyscallN(addr, uintptr(nCount), uintptr(unsafe.Pointer(pHandles)), uintptr(dwMilliseconds), uintptr(dwWakeMask), uintptr(dwFlags))
	return uint32(ret), WIN32_ERROR(err)
}

func SetTimer(hWnd HWND, nIDEvent uintptr, uElapse uint32, lpTimerFunc TIMERPROC) (uintptr, WIN32_ERROR) {
	addr := lazyAddr(&pSetTimer, libUser32, "SetTimer")
	ret, _, err := syscall.SyscallN(addr, hWnd, nIDEvent, uintptr(uElapse), lpTimerFunc)
	return ret, WIN32_ERROR(err)
}

func SetCoalescableTimer(hWnd HWND, nIDEvent uintptr, uElapse uint32, lpTimerFunc TIMERPROC, uToleranceDelay uint32) (uintptr, WIN32_ERROR) {
	addr := lazyAddr(&pSetCoalescableTimer, libUser32, "SetCoalescableTimer")
	ret, _, err := syscall.SyscallN(addr, hWnd, nIDEvent, uintptr(uElapse), lpTimerFunc, uintptr(uToleranceDelay))
	return ret, WIN32_ERROR(err)
}

func KillTimer(hWnd HWND, uIDEvent uintptr) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pKillTimer, libUser32, "KillTimer")
	ret, _, err := syscall.SyscallN(addr, hWnd, uIDEvent)
	return BOOL(ret), WIN32_ERROR(err)
}

func IsWindowUnicode(hWnd HWND) BOOL {
	addr := lazyAddr(&pIsWindowUnicode, libUser32, "IsWindowUnicode")
	ret, _, _ := syscall.SyscallN(addr, hWnd)
	return BOOL(ret)
}

func LoadAcceleratorsA(hInstance HINSTANCE, lpTableName PSTR) (HACCEL, WIN32_ERROR) {
	addr := lazyAddr(&pLoadAcceleratorsA, libUser32, "LoadAcceleratorsA")
	ret, _, err := syscall.SyscallN(addr, hInstance, uintptr(unsafe.Pointer(lpTableName)))
	return ret, WIN32_ERROR(err)
}

var LoadAccelerators = LoadAcceleratorsW

func LoadAcceleratorsW(hInstance HINSTANCE, lpTableName PWSTR) (HACCEL, WIN32_ERROR) {
	addr := lazyAddr(&pLoadAcceleratorsW, libUser32, "LoadAcceleratorsW")
	ret, _, err := syscall.SyscallN(addr, hInstance, uintptr(unsafe.Pointer(lpTableName)))
	return ret, WIN32_ERROR(err)
}

func CreateAcceleratorTableA(paccel *ACCEL, cAccel int32) (HACCEL, WIN32_ERROR) {
	addr := lazyAddr(&pCreateAcceleratorTableA, libUser32, "CreateAcceleratorTableA")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(paccel)), uintptr(cAccel))
	return ret, WIN32_ERROR(err)
}

var CreateAcceleratorTable = CreateAcceleratorTableW

func CreateAcceleratorTableW(paccel *ACCEL, cAccel int32) (HACCEL, WIN32_ERROR) {
	addr := lazyAddr(&pCreateAcceleratorTableW, libUser32, "CreateAcceleratorTableW")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(paccel)), uintptr(cAccel))
	return ret, WIN32_ERROR(err)
}

func DestroyAcceleratorTable(hAccel HACCEL) BOOL {
	addr := lazyAddr(&pDestroyAcceleratorTable, libUser32, "DestroyAcceleratorTable")
	ret, _, _ := syscall.SyscallN(addr, hAccel)
	return BOOL(ret)
}

func CopyAcceleratorTableA(hAccelSrc HACCEL, lpAccelDst *ACCEL, cAccelEntries int32) int32 {
	addr := lazyAddr(&pCopyAcceleratorTableA, libUser32, "CopyAcceleratorTableA")
	ret, _, _ := syscall.SyscallN(addr, hAccelSrc, uintptr(unsafe.Pointer(lpAccelDst)), uintptr(cAccelEntries))
	return int32(ret)
}

var CopyAcceleratorTable = CopyAcceleratorTableW

func CopyAcceleratorTableW(hAccelSrc HACCEL, lpAccelDst *ACCEL, cAccelEntries int32) int32 {
	addr := lazyAddr(&pCopyAcceleratorTableW, libUser32, "CopyAcceleratorTableW")
	ret, _, _ := syscall.SyscallN(addr, hAccelSrc, uintptr(unsafe.Pointer(lpAccelDst)), uintptr(cAccelEntries))
	return int32(ret)
}

func TranslateAcceleratorA(hWnd HWND, hAccTable HACCEL, lpMsg *MSG) (int32, WIN32_ERROR) {
	addr := lazyAddr(&pTranslateAcceleratorA, libUser32, "TranslateAcceleratorA")
	ret, _, err := syscall.SyscallN(addr, hWnd, hAccTable, uintptr(unsafe.Pointer(lpMsg)))
	return int32(ret), WIN32_ERROR(err)
}

var TranslateAccelerator = TranslateAcceleratorW

func TranslateAcceleratorW(hWnd HWND, hAccTable HACCEL, lpMsg *MSG) (int32, WIN32_ERROR) {
	addr := lazyAddr(&pTranslateAcceleratorW, libUser32, "TranslateAcceleratorW")
	ret, _, err := syscall.SyscallN(addr, hWnd, hAccTable, uintptr(unsafe.Pointer(lpMsg)))
	return int32(ret), WIN32_ERROR(err)
}

func GetSystemMetrics(nIndex SYSTEM_METRICS_INDEX) (int32, WIN32_ERROR) {
	addr := lazyAddr(&pGetSystemMetrics, libUser32, "GetSystemMetrics")
	ret, _, err := syscall.SyscallN(addr, uintptr(nIndex))
	return int32(ret), WIN32_ERROR(err)
}

func LoadMenuA(hInstance HINSTANCE, lpMenuName PSTR) (HMENU, WIN32_ERROR) {
	addr := lazyAddr(&pLoadMenuA, libUser32, "LoadMenuA")
	ret, _, err := syscall.SyscallN(addr, hInstance, uintptr(unsafe.Pointer(lpMenuName)))
	return ret, WIN32_ERROR(err)
}

var LoadMenu = LoadMenuW

func LoadMenuW(hInstance HINSTANCE, lpMenuName PWSTR) (HMENU, WIN32_ERROR) {
	addr := lazyAddr(&pLoadMenuW, libUser32, "LoadMenuW")
	ret, _, err := syscall.SyscallN(addr, hInstance, uintptr(unsafe.Pointer(lpMenuName)))
	return ret, WIN32_ERROR(err)
}

func LoadMenuIndirectA(lpMenuTemplate unsafe.Pointer) (HMENU, WIN32_ERROR) {
	addr := lazyAddr(&pLoadMenuIndirectA, libUser32, "LoadMenuIndirectA")
	ret, _, err := syscall.SyscallN(addr, uintptr(lpMenuTemplate))
	return ret, WIN32_ERROR(err)
}

var LoadMenuIndirect = LoadMenuIndirectW

func LoadMenuIndirectW(lpMenuTemplate unsafe.Pointer) (HMENU, WIN32_ERROR) {
	addr := lazyAddr(&pLoadMenuIndirectW, libUser32, "LoadMenuIndirectW")
	ret, _, err := syscall.SyscallN(addr, uintptr(lpMenuTemplate))
	return ret, WIN32_ERROR(err)
}

func GetMenu(hWnd HWND) HMENU {
	addr := lazyAddr(&pGetMenu, libUser32, "GetMenu")
	ret, _, _ := syscall.SyscallN(addr, hWnd)
	return ret
}

func SetMenu(hWnd HWND, hMenu HMENU) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pSetMenu, libUser32, "SetMenu")
	ret, _, err := syscall.SyscallN(addr, hWnd, hMenu)
	return BOOL(ret), WIN32_ERROR(err)
}

func ChangeMenuA(hMenu HMENU, cmd uint32, lpszNewItem PSTR, cmdInsert uint32, flags uint32) BOOL {
	addr := lazyAddr(&pChangeMenuA, libUser32, "ChangeMenuA")
	ret, _, _ := syscall.SyscallN(addr, hMenu, uintptr(cmd), uintptr(unsafe.Pointer(lpszNewItem)), uintptr(cmdInsert), uintptr(flags))
	return BOOL(ret)
}

var ChangeMenu = ChangeMenuW

func ChangeMenuW(hMenu HMENU, cmd uint32, lpszNewItem PWSTR, cmdInsert uint32, flags uint32) BOOL {
	addr := lazyAddr(&pChangeMenuW, libUser32, "ChangeMenuW")
	ret, _, _ := syscall.SyscallN(addr, hMenu, uintptr(cmd), uintptr(unsafe.Pointer(lpszNewItem)), uintptr(cmdInsert), uintptr(flags))
	return BOOL(ret)
}

func HiliteMenuItem(hWnd HWND, hMenu HMENU, uIDHiliteItem uint32, uHilite uint32) BOOL {
	addr := lazyAddr(&pHiliteMenuItem, libUser32, "HiliteMenuItem")
	ret, _, _ := syscall.SyscallN(addr, hWnd, hMenu, uintptr(uIDHiliteItem), uintptr(uHilite))
	return BOOL(ret)
}

func GetMenuStringA(hMenu HMENU, uIDItem uint32, lpString PSTR, cchMax int32, flags MENU_ITEM_FLAGS) int32 {
	addr := lazyAddr(&pGetMenuStringA, libUser32, "GetMenuStringA")
	ret, _, _ := syscall.SyscallN(addr, hMenu, uintptr(uIDItem), uintptr(unsafe.Pointer(lpString)), uintptr(cchMax), uintptr(flags))
	return int32(ret)
}

var GetMenuString = GetMenuStringW

func GetMenuStringW(hMenu HMENU, uIDItem uint32, lpString PWSTR, cchMax int32, flags MENU_ITEM_FLAGS) int32 {
	addr := lazyAddr(&pGetMenuStringW, libUser32, "GetMenuStringW")
	ret, _, _ := syscall.SyscallN(addr, hMenu, uintptr(uIDItem), uintptr(unsafe.Pointer(lpString)), uintptr(cchMax), uintptr(flags))
	return int32(ret)
}

func GetMenuState(hMenu HMENU, uId uint32, uFlags MENU_ITEM_FLAGS) uint32 {
	addr := lazyAddr(&pGetMenuState, libUser32, "GetMenuState")
	ret, _, _ := syscall.SyscallN(addr, hMenu, uintptr(uId), uintptr(uFlags))
	return uint32(ret)
}

func DrawMenuBar(hWnd HWND) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pDrawMenuBar, libUser32, "DrawMenuBar")
	ret, _, err := syscall.SyscallN(addr, hWnd)
	return BOOL(ret), WIN32_ERROR(err)
}

func GetSystemMenu(hWnd HWND, bRevert BOOL) HMENU {
	addr := lazyAddr(&pGetSystemMenu, libUser32, "GetSystemMenu")
	ret, _, _ := syscall.SyscallN(addr, hWnd, uintptr(bRevert))
	return ret
}

func CreateMenu() (HMENU, WIN32_ERROR) {
	addr := lazyAddr(&pCreateMenu, libUser32, "CreateMenu")
	ret, _, err := syscall.SyscallN(addr)
	return ret, WIN32_ERROR(err)
}

func CreatePopupMenu() (HMENU, WIN32_ERROR) {
	addr := lazyAddr(&pCreatePopupMenu, libUser32, "CreatePopupMenu")
	ret, _, err := syscall.SyscallN(addr)
	return ret, WIN32_ERROR(err)
}

func DestroyMenu(hMenu HMENU) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pDestroyMenu, libUser32, "DestroyMenu")
	ret, _, err := syscall.SyscallN(addr, hMenu)
	return BOOL(ret), WIN32_ERROR(err)
}

func CheckMenuItem(hMenu HMENU, uIDCheckItem uint32, uCheck uint32) uint32 {
	addr := lazyAddr(&pCheckMenuItem, libUser32, "CheckMenuItem")
	ret, _, _ := syscall.SyscallN(addr, hMenu, uintptr(uIDCheckItem), uintptr(uCheck))
	return uint32(ret)
}

func EnableMenuItem(hMenu HMENU, uIDEnableItem uint32, uEnable MENU_ITEM_FLAGS) BOOL {
	addr := lazyAddr(&pEnableMenuItem, libUser32, "EnableMenuItem")
	ret, _, _ := syscall.SyscallN(addr, hMenu, uintptr(uIDEnableItem), uintptr(uEnable))
	return BOOL(ret)
}

func GetSubMenu(hMenu HMENU, nPos int32) HMENU {
	addr := lazyAddr(&pGetSubMenu, libUser32, "GetSubMenu")
	ret, _, _ := syscall.SyscallN(addr, hMenu, uintptr(nPos))
	return ret
}

func GetMenuItemID(hMenu HMENU, nPos int32) uint32 {
	addr := lazyAddr(&pGetMenuItemID, libUser32, "GetMenuItemID")
	ret, _, _ := syscall.SyscallN(addr, hMenu, uintptr(nPos))
	return uint32(ret)
}

func GetMenuItemCount(hMenu HMENU) (int32, WIN32_ERROR) {
	addr := lazyAddr(&pGetMenuItemCount, libUser32, "GetMenuItemCount")
	ret, _, err := syscall.SyscallN(addr, hMenu)
	return int32(ret), WIN32_ERROR(err)
}

func InsertMenuA(hMenu HMENU, uPosition uint32, uFlags MENU_ITEM_FLAGS, uIDNewItem uintptr, lpNewItem PSTR) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pInsertMenuA, libUser32, "InsertMenuA")
	ret, _, err := syscall.SyscallN(addr, hMenu, uintptr(uPosition), uintptr(uFlags), uIDNewItem, uintptr(unsafe.Pointer(lpNewItem)))
	return BOOL(ret), WIN32_ERROR(err)
}

var InsertMenu = InsertMenuW

func InsertMenuW(hMenu HMENU, uPosition uint32, uFlags MENU_ITEM_FLAGS, uIDNewItem uintptr, lpNewItem PWSTR) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pInsertMenuW, libUser32, "InsertMenuW")
	ret, _, err := syscall.SyscallN(addr, hMenu, uintptr(uPosition), uintptr(uFlags), uIDNewItem, uintptr(unsafe.Pointer(lpNewItem)))
	return BOOL(ret), WIN32_ERROR(err)
}

func AppendMenuA(hMenu HMENU, uFlags MENU_ITEM_FLAGS, uIDNewItem uintptr, lpNewItem PSTR) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pAppendMenuA, libUser32, "AppendMenuA")
	ret, _, err := syscall.SyscallN(addr, hMenu, uintptr(uFlags), uIDNewItem, uintptr(unsafe.Pointer(lpNewItem)))
	return BOOL(ret), WIN32_ERROR(err)
}

var AppendMenu = AppendMenuW

func AppendMenuW(hMenu HMENU, uFlags MENU_ITEM_FLAGS, uIDNewItem uintptr, lpNewItem PWSTR) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pAppendMenuW, libUser32, "AppendMenuW")
	ret, _, err := syscall.SyscallN(addr, hMenu, uintptr(uFlags), uIDNewItem, uintptr(unsafe.Pointer(lpNewItem)))
	return BOOL(ret), WIN32_ERROR(err)
}

func ModifyMenuA(hMnu HMENU, uPosition uint32, uFlags MENU_ITEM_FLAGS, uIDNewItem uintptr, lpNewItem PSTR) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pModifyMenuA, libUser32, "ModifyMenuA")
	ret, _, err := syscall.SyscallN(addr, hMnu, uintptr(uPosition), uintptr(uFlags), uIDNewItem, uintptr(unsafe.Pointer(lpNewItem)))
	return BOOL(ret), WIN32_ERROR(err)
}

var ModifyMenu = ModifyMenuW

func ModifyMenuW(hMnu HMENU, uPosition uint32, uFlags MENU_ITEM_FLAGS, uIDNewItem uintptr, lpNewItem PWSTR) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pModifyMenuW, libUser32, "ModifyMenuW")
	ret, _, err := syscall.SyscallN(addr, hMnu, uintptr(uPosition), uintptr(uFlags), uIDNewItem, uintptr(unsafe.Pointer(lpNewItem)))
	return BOOL(ret), WIN32_ERROR(err)
}

func RemoveMenu(hMenu HMENU, uPosition uint32, uFlags MENU_ITEM_FLAGS) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pRemoveMenu, libUser32, "RemoveMenu")
	ret, _, err := syscall.SyscallN(addr, hMenu, uintptr(uPosition), uintptr(uFlags))
	return BOOL(ret), WIN32_ERROR(err)
}

func DeleteMenu(hMenu HMENU, uPosition uint32, uFlags MENU_ITEM_FLAGS) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pDeleteMenu, libUser32, "DeleteMenu")
	ret, _, err := syscall.SyscallN(addr, hMenu, uintptr(uPosition), uintptr(uFlags))
	return BOOL(ret), WIN32_ERROR(err)
}

func SetMenuItemBitmaps(hMenu HMENU, uPosition uint32, uFlags MENU_ITEM_FLAGS, hBitmapUnchecked HBITMAP, hBitmapChecked HBITMAP) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pSetMenuItemBitmaps, libUser32, "SetMenuItemBitmaps")
	ret, _, err := syscall.SyscallN(addr, hMenu, uintptr(uPosition), uintptr(uFlags), hBitmapUnchecked, hBitmapChecked)
	return BOOL(ret), WIN32_ERROR(err)
}

func GetMenuCheckMarkDimensions() int32 {
	addr := lazyAddr(&pGetMenuCheckMarkDimensions, libUser32, "GetMenuCheckMarkDimensions")
	ret, _, _ := syscall.SyscallN(addr)
	return int32(ret)
}

func TrackPopupMenu(hMenu HMENU, uFlags TRACK_POPUP_MENU_FLAGS, x int32, y int32, nReserved int32, hWnd HWND, prcRect *RECT) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pTrackPopupMenu, libUser32, "TrackPopupMenu")
	ret, _, err := syscall.SyscallN(addr, hMenu, uintptr(uFlags), uintptr(x), uintptr(y), uintptr(nReserved), hWnd, uintptr(unsafe.Pointer(prcRect)))
	return BOOL(ret), WIN32_ERROR(err)
}

func TrackPopupMenuEx(hMenu HMENU, uFlags uint32, x int32, y int32, hwnd HWND, lptpm *TPMPARAMS) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pTrackPopupMenuEx, libUser32, "TrackPopupMenuEx")
	ret, _, err := syscall.SyscallN(addr, hMenu, uintptr(uFlags), uintptr(x), uintptr(y), hwnd, uintptr(unsafe.Pointer(lptpm)))
	return BOOL(ret), WIN32_ERROR(err)
}

func CalculatePopupWindowPosition(anchorPoint *POINT, windowSize *SIZE, flags uint32, excludeRect *RECT, popupWindowPosition *RECT) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pCalculatePopupWindowPosition, libUser32, "CalculatePopupWindowPosition")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(anchorPoint)), uintptr(unsafe.Pointer(windowSize)), uintptr(flags), uintptr(unsafe.Pointer(excludeRect)), uintptr(unsafe.Pointer(popupWindowPosition)))
	return BOOL(ret), WIN32_ERROR(err)
}

func GetMenuInfo(param0 HMENU, param1 *MENUINFO) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pGetMenuInfo, libUser32, "GetMenuInfo")
	ret, _, err := syscall.SyscallN(addr, param0, uintptr(unsafe.Pointer(param1)))
	return BOOL(ret), WIN32_ERROR(err)
}

func SetMenuInfo(param0 HMENU, param1 *MENUINFO) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pSetMenuInfo, libUser32, "SetMenuInfo")
	ret, _, err := syscall.SyscallN(addr, param0, uintptr(unsafe.Pointer(param1)))
	return BOOL(ret), WIN32_ERROR(err)
}

func EndMenu() (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pEndMenu, libUser32, "EndMenu")
	ret, _, err := syscall.SyscallN(addr)
	return BOOL(ret), WIN32_ERROR(err)
}

func InsertMenuItemA(hmenu HMENU, item uint32, fByPosition BOOL, lpmi *MENUITEMINFOA) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pInsertMenuItemA, libUser32, "InsertMenuItemA")
	ret, _, err := syscall.SyscallN(addr, hmenu, uintptr(item), uintptr(fByPosition), uintptr(unsafe.Pointer(lpmi)))
	return BOOL(ret), WIN32_ERROR(err)
}

var InsertMenuItem = InsertMenuItemW

func InsertMenuItemW(hmenu HMENU, item uint32, fByPosition BOOL, lpmi *MENUITEMINFOW) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pInsertMenuItemW, libUser32, "InsertMenuItemW")
	ret, _, err := syscall.SyscallN(addr, hmenu, uintptr(item), uintptr(fByPosition), uintptr(unsafe.Pointer(lpmi)))
	return BOOL(ret), WIN32_ERROR(err)
}

func GetMenuItemInfoA(hmenu HMENU, item uint32, fByPosition BOOL, lpmii *MENUITEMINFOA) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pGetMenuItemInfoA, libUser32, "GetMenuItemInfoA")
	ret, _, err := syscall.SyscallN(addr, hmenu, uintptr(item), uintptr(fByPosition), uintptr(unsafe.Pointer(lpmii)))
	return BOOL(ret), WIN32_ERROR(err)
}

var GetMenuItemInfo = GetMenuItemInfoW

func GetMenuItemInfoW(hmenu HMENU, item uint32, fByPosition BOOL, lpmii *MENUITEMINFOW) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pGetMenuItemInfoW, libUser32, "GetMenuItemInfoW")
	ret, _, err := syscall.SyscallN(addr, hmenu, uintptr(item), uintptr(fByPosition), uintptr(unsafe.Pointer(lpmii)))
	return BOOL(ret), WIN32_ERROR(err)
}

func SetMenuItemInfoA(hmenu HMENU, item uint32, fByPositon BOOL, lpmii *MENUITEMINFOA) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pSetMenuItemInfoA, libUser32, "SetMenuItemInfoA")
	ret, _, err := syscall.SyscallN(addr, hmenu, uintptr(item), uintptr(fByPositon), uintptr(unsafe.Pointer(lpmii)))
	return BOOL(ret), WIN32_ERROR(err)
}

var SetMenuItemInfo = SetMenuItemInfoW

func SetMenuItemInfoW(hmenu HMENU, item uint32, fByPositon BOOL, lpmii *MENUITEMINFOW) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pSetMenuItemInfoW, libUser32, "SetMenuItemInfoW")
	ret, _, err := syscall.SyscallN(addr, hmenu, uintptr(item), uintptr(fByPositon), uintptr(unsafe.Pointer(lpmii)))
	return BOOL(ret), WIN32_ERROR(err)
}

func GetMenuDefaultItem(hMenu HMENU, fByPos uint32, gmdiFlags GET_MENU_DEFAULT_ITEM_FLAGS) (uint32, WIN32_ERROR) {
	addr := lazyAddr(&pGetMenuDefaultItem, libUser32, "GetMenuDefaultItem")
	ret, _, err := syscall.SyscallN(addr, hMenu, uintptr(fByPos), uintptr(gmdiFlags))
	return uint32(ret), WIN32_ERROR(err)
}

func SetMenuDefaultItem(hMenu HMENU, uItem uint32, fByPos uint32) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pSetMenuDefaultItem, libUser32, "SetMenuDefaultItem")
	ret, _, err := syscall.SyscallN(addr, hMenu, uintptr(uItem), uintptr(fByPos))
	return BOOL(ret), WIN32_ERROR(err)
}

func GetMenuItemRect(hWnd HWND, hMenu HMENU, uItem uint32, lprcItem *RECT) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pGetMenuItemRect, libUser32, "GetMenuItemRect")
	ret, _, err := syscall.SyscallN(addr, hWnd, hMenu, uintptr(uItem), uintptr(unsafe.Pointer(lprcItem)))
	return BOOL(ret), WIN32_ERROR(err)
}

func MenuItemFromPoint(hWnd HWND, hMenu HMENU, ptScreen POINT) int32 {
	addr := lazyAddr(&pMenuItemFromPoint, libUser32, "MenuItemFromPoint")
	ret, _, _ := syscall.SyscallN(addr, hWnd, hMenu, *(*uintptr)(unsafe.Pointer(&ptScreen)))
	return int32(ret)
}

func DragObject(hwndParent HWND, hwndFrom HWND, fmt uint32, data uintptr, hcur HCURSOR) uint32 {
	addr := lazyAddr(&pDragObject, libUser32, "DragObject")
	ret, _, _ := syscall.SyscallN(addr, hwndParent, hwndFrom, uintptr(fmt), data, hcur)
	return uint32(ret)
}

func DrawIcon(hDC HDC, X int32, Y int32, hIcon HICON) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pDrawIcon, libUser32, "DrawIcon")
	ret, _, err := syscall.SyscallN(addr, hDC, uintptr(X), uintptr(Y), hIcon)
	return BOOL(ret), WIN32_ERROR(err)
}

func GetForegroundWindow() HWND {
	addr := lazyAddr(&pGetForegroundWindow, libUser32, "GetForegroundWindow")
	ret, _, _ := syscall.SyscallN(addr)
	return ret
}

func SwitchToThisWindow(hwnd HWND, fUnknown BOOL) {
	addr := lazyAddr(&pSwitchToThisWindow, libUser32, "SwitchToThisWindow")
	syscall.SyscallN(addr, hwnd, uintptr(fUnknown))
}

func SetForegroundWindow(hWnd HWND) BOOL {
	addr := lazyAddr(&pSetForegroundWindow, libUser32, "SetForegroundWindow")
	ret, _, _ := syscall.SyscallN(addr, hWnd)
	return BOOL(ret)
}

func AllowSetForegroundWindow(dwProcessId uint32) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pAllowSetForegroundWindow, libUser32, "AllowSetForegroundWindow")
	ret, _, err := syscall.SyscallN(addr, uintptr(dwProcessId))
	return BOOL(ret), WIN32_ERROR(err)
}

func LockSetForegroundWindow(uLockCode FOREGROUND_WINDOW_LOCK_CODE) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pLockSetForegroundWindow, libUser32, "LockSetForegroundWindow")
	ret, _, err := syscall.SyscallN(addr, uintptr(uLockCode))
	return BOOL(ret), WIN32_ERROR(err)
}

func ScrollWindow(hWnd HWND, XAmount int32, YAmount int32, lpRect *RECT, lpClipRect *RECT) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pScrollWindow, libUser32, "ScrollWindow")
	ret, _, err := syscall.SyscallN(addr, hWnd, uintptr(XAmount), uintptr(YAmount), uintptr(unsafe.Pointer(lpRect)), uintptr(unsafe.Pointer(lpClipRect)))
	return BOOL(ret), WIN32_ERROR(err)
}

func ScrollDC(hDC HDC, dx int32, dy int32, lprcScroll *RECT, lprcClip *RECT, hrgnUpdate HRGN, lprcUpdate *RECT) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pScrollDC, libUser32, "ScrollDC")
	ret, _, err := syscall.SyscallN(addr, hDC, uintptr(dx), uintptr(dy), uintptr(unsafe.Pointer(lprcScroll)), uintptr(unsafe.Pointer(lprcClip)), hrgnUpdate, uintptr(unsafe.Pointer(lprcUpdate)))
	return BOOL(ret), WIN32_ERROR(err)
}

func ScrollWindowEx(hWnd HWND, dx int32, dy int32, prcScroll *RECT, prcClip *RECT, hrgnUpdate HRGN, prcUpdate *RECT, flags SHOW_WINDOW_CMD) (int32, WIN32_ERROR) {
	addr := lazyAddr(&pScrollWindowEx, libUser32, "ScrollWindowEx")
	ret, _, err := syscall.SyscallN(addr, hWnd, uintptr(dx), uintptr(dy), uintptr(unsafe.Pointer(prcScroll)), uintptr(unsafe.Pointer(prcClip)), hrgnUpdate, uintptr(unsafe.Pointer(prcUpdate)), uintptr(flags))
	return int32(ret), WIN32_ERROR(err)
}

func GetScrollPos(hWnd HWND, nBar SCROLLBAR_CONSTANTS) (int32, WIN32_ERROR) {
	addr := lazyAddr(&pGetScrollPos, libUser32, "GetScrollPos")
	ret, _, err := syscall.SyscallN(addr, hWnd, uintptr(nBar))
	return int32(ret), WIN32_ERROR(err)
}

func GetScrollRange(hWnd HWND, nBar SCROLLBAR_CONSTANTS, lpMinPos *int32, lpMaxPos *int32) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pGetScrollRange, libUser32, "GetScrollRange")
	ret, _, err := syscall.SyscallN(addr, hWnd, uintptr(nBar), uintptr(unsafe.Pointer(lpMinPos)), uintptr(unsafe.Pointer(lpMaxPos)))
	return BOOL(ret), WIN32_ERROR(err)
}

func SetPropA(hWnd HWND, lpString PSTR, hData HANDLE) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pSetPropA, libUser32, "SetPropA")
	ret, _, err := syscall.SyscallN(addr, hWnd, uintptr(unsafe.Pointer(lpString)), hData)
	return BOOL(ret), WIN32_ERROR(err)
}

var SetProp = SetPropW

func SetPropW(hWnd HWND, lpString PWSTR, hData HANDLE) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pSetPropW, libUser32, "SetPropW")
	ret, _, err := syscall.SyscallN(addr, hWnd, uintptr(unsafe.Pointer(lpString)), hData)
	return BOOL(ret), WIN32_ERROR(err)
}

func GetPropA(hWnd HWND, lpString PSTR) HANDLE {
	addr := lazyAddr(&pGetPropA, libUser32, "GetPropA")
	ret, _, _ := syscall.SyscallN(addr, hWnd, uintptr(unsafe.Pointer(lpString)))
	return ret
}

var GetProp = GetPropW

func GetPropW(hWnd HWND, lpString PWSTR) HANDLE {
	addr := lazyAddr(&pGetPropW, libUser32, "GetPropW")
	ret, _, _ := syscall.SyscallN(addr, hWnd, uintptr(unsafe.Pointer(lpString)))
	return ret
}

func RemovePropA(hWnd HWND, lpString PSTR) (HANDLE, WIN32_ERROR) {
	addr := lazyAddr(&pRemovePropA, libUser32, "RemovePropA")
	ret, _, err := syscall.SyscallN(addr, hWnd, uintptr(unsafe.Pointer(lpString)))
	return ret, WIN32_ERROR(err)
}

var RemoveProp = RemovePropW

func RemovePropW(hWnd HWND, lpString PWSTR) (HANDLE, WIN32_ERROR) {
	addr := lazyAddr(&pRemovePropW, libUser32, "RemovePropW")
	ret, _, err := syscall.SyscallN(addr, hWnd, uintptr(unsafe.Pointer(lpString)))
	return ret, WIN32_ERROR(err)
}

func EnumPropsExA(hWnd HWND, lpEnumFunc PROPENUMPROCEXA, lParam LPARAM) int32 {
	addr := lazyAddr(&pEnumPropsExA, libUser32, "EnumPropsExA")
	ret, _, _ := syscall.SyscallN(addr, hWnd, lpEnumFunc, lParam)
	return int32(ret)
}

var EnumPropsEx = EnumPropsExW

func EnumPropsExW(hWnd HWND, lpEnumFunc PROPENUMPROCEXW, lParam LPARAM) int32 {
	addr := lazyAddr(&pEnumPropsExW, libUser32, "EnumPropsExW")
	ret, _, _ := syscall.SyscallN(addr, hWnd, lpEnumFunc, lParam)
	return int32(ret)
}

func EnumPropsA(hWnd HWND, lpEnumFunc PROPENUMPROCA) int32 {
	addr := lazyAddr(&pEnumPropsA, libUser32, "EnumPropsA")
	ret, _, _ := syscall.SyscallN(addr, hWnd, lpEnumFunc)
	return int32(ret)
}

var EnumProps = EnumPropsW

func EnumPropsW(hWnd HWND, lpEnumFunc PROPENUMPROCW) int32 {
	addr := lazyAddr(&pEnumPropsW, libUser32, "EnumPropsW")
	ret, _, _ := syscall.SyscallN(addr, hWnd, lpEnumFunc)
	return int32(ret)
}

func SetWindowTextA(hWnd HWND, lpString PSTR) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pSetWindowTextA, libUser32, "SetWindowTextA")
	ret, _, err := syscall.SyscallN(addr, hWnd, uintptr(unsafe.Pointer(lpString)))
	return BOOL(ret), WIN32_ERROR(err)
}

var SetWindowText = SetWindowTextW

func SetWindowTextW(hWnd HWND, lpString PWSTR) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pSetWindowTextW, libUser32, "SetWindowTextW")
	ret, _, err := syscall.SyscallN(addr, hWnd, uintptr(unsafe.Pointer(lpString)))
	return BOOL(ret), WIN32_ERROR(err)
}

func GetWindowTextA(hWnd HWND, lpString PSTR, nMaxCount int32) (int32, WIN32_ERROR) {
	addr := lazyAddr(&pGetWindowTextA, libUser32, "GetWindowTextA")
	ret, _, err := syscall.SyscallN(addr, hWnd, uintptr(unsafe.Pointer(lpString)), uintptr(nMaxCount))
	return int32(ret), WIN32_ERROR(err)
}

var GetWindowText = GetWindowTextW

func GetWindowTextW(hWnd HWND, lpString PWSTR, nMaxCount int32) (int32, WIN32_ERROR) {
	addr := lazyAddr(&pGetWindowTextW, libUser32, "GetWindowTextW")
	ret, _, err := syscall.SyscallN(addr, hWnd, uintptr(unsafe.Pointer(lpString)), uintptr(nMaxCount))
	return int32(ret), WIN32_ERROR(err)
}

func GetWindowTextLengthA(hWnd HWND) (int32, WIN32_ERROR) {
	addr := lazyAddr(&pGetWindowTextLengthA, libUser32, "GetWindowTextLengthA")
	ret, _, err := syscall.SyscallN(addr, hWnd)
	return int32(ret), WIN32_ERROR(err)
}

var GetWindowTextLength = GetWindowTextLengthW

func GetWindowTextLengthW(hWnd HWND) (int32, WIN32_ERROR) {
	addr := lazyAddr(&pGetWindowTextLengthW, libUser32, "GetWindowTextLengthW")
	ret, _, err := syscall.SyscallN(addr, hWnd)
	return int32(ret), WIN32_ERROR(err)
}

func GetClientRect(hWnd HWND, lpRect *RECT) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pGetClientRect, libUser32, "GetClientRect")
	ret, _, err := syscall.SyscallN(addr, hWnd, uintptr(unsafe.Pointer(lpRect)))
	return BOOL(ret), WIN32_ERROR(err)
}

func GetWindowRect(hWnd HWND, lpRect *RECT) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pGetWindowRect, libUser32, "GetWindowRect")
	ret, _, err := syscall.SyscallN(addr, hWnd, uintptr(unsafe.Pointer(lpRect)))
	return BOOL(ret), WIN32_ERROR(err)
}

func AdjustWindowRect(lpRect *RECT, dwStyle WINDOW_STYLE, bMenu BOOL) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pAdjustWindowRect, libUser32, "AdjustWindowRect")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpRect)), uintptr(dwStyle), uintptr(bMenu))
	return BOOL(ret), WIN32_ERROR(err)
}

func AdjustWindowRectEx(lpRect *RECT, dwStyle WINDOW_STYLE, bMenu BOOL, dwExStyle WINDOW_EX_STYLE) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pAdjustWindowRectEx, libUser32, "AdjustWindowRectEx")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpRect)), uintptr(dwStyle), uintptr(bMenu), uintptr(dwExStyle))
	return BOOL(ret), WIN32_ERROR(err)
}

func MessageBoxA(hWnd HWND, lpText PSTR, lpCaption PSTR, uType MESSAGEBOX_STYLE) (MESSAGEBOX_RESULT, WIN32_ERROR) {
	addr := lazyAddr(&pMessageBoxA, libUser32, "MessageBoxA")
	ret, _, err := syscall.SyscallN(addr, hWnd, uintptr(unsafe.Pointer(lpText)), uintptr(unsafe.Pointer(lpCaption)), uintptr(uType))
	return MESSAGEBOX_RESULT(ret), WIN32_ERROR(err)
}

var MessageBox = MessageBoxW

func MessageBoxW(hWnd HWND, lpText PWSTR, lpCaption PWSTR, uType MESSAGEBOX_STYLE) (MESSAGEBOX_RESULT, WIN32_ERROR) {
	addr := lazyAddr(&pMessageBoxW, libUser32, "MessageBoxW")
	ret, _, err := syscall.SyscallN(addr, hWnd, uintptr(unsafe.Pointer(lpText)), uintptr(unsafe.Pointer(lpCaption)), uintptr(uType))
	return MESSAGEBOX_RESULT(ret), WIN32_ERROR(err)
}

func MessageBoxExA(hWnd HWND, lpText PSTR, lpCaption PSTR, uType MESSAGEBOX_STYLE, wLanguageId uint16) (MESSAGEBOX_RESULT, WIN32_ERROR) {
	addr := lazyAddr(&pMessageBoxExA, libUser32, "MessageBoxExA")
	ret, _, err := syscall.SyscallN(addr, hWnd, uintptr(unsafe.Pointer(lpText)), uintptr(unsafe.Pointer(lpCaption)), uintptr(uType), uintptr(wLanguageId))
	return MESSAGEBOX_RESULT(ret), WIN32_ERROR(err)
}

var MessageBoxEx = MessageBoxExW

func MessageBoxExW(hWnd HWND, lpText PWSTR, lpCaption PWSTR, uType MESSAGEBOX_STYLE, wLanguageId uint16) (MESSAGEBOX_RESULT, WIN32_ERROR) {
	addr := lazyAddr(&pMessageBoxExW, libUser32, "MessageBoxExW")
	ret, _, err := syscall.SyscallN(addr, hWnd, uintptr(unsafe.Pointer(lpText)), uintptr(unsafe.Pointer(lpCaption)), uintptr(uType), uintptr(wLanguageId))
	return MESSAGEBOX_RESULT(ret), WIN32_ERROR(err)
}

func MessageBoxIndirectA(lpmbp *MSGBOXPARAMSA) MESSAGEBOX_RESULT {
	addr := lazyAddr(&pMessageBoxIndirectA, libUser32, "MessageBoxIndirectA")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpmbp)))
	return MESSAGEBOX_RESULT(ret)
}

var MessageBoxIndirect = MessageBoxIndirectW

func MessageBoxIndirectW(lpmbp *MSGBOXPARAMSW) MESSAGEBOX_RESULT {
	addr := lazyAddr(&pMessageBoxIndirectW, libUser32, "MessageBoxIndirectW")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpmbp)))
	return MESSAGEBOX_RESULT(ret)
}

func ShowCursor(bShow BOOL) int32 {
	addr := lazyAddr(&pShowCursor, libUser32, "ShowCursor")
	ret, _, _ := syscall.SyscallN(addr, uintptr(bShow))
	return int32(ret)
}

func SetCursorPos(X int32, Y int32) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pSetCursorPos, libUser32, "SetCursorPos")
	ret, _, err := syscall.SyscallN(addr, uintptr(X), uintptr(Y))
	return BOOL(ret), WIN32_ERROR(err)
}

func SetPhysicalCursorPos(X int32, Y int32) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pSetPhysicalCursorPos, libUser32, "SetPhysicalCursorPos")
	ret, _, err := syscall.SyscallN(addr, uintptr(X), uintptr(Y))
	return BOOL(ret), WIN32_ERROR(err)
}

func SetCursor(hCursor HCURSOR) HCURSOR {
	addr := lazyAddr(&pSetCursor, libUser32, "SetCursor")
	ret, _, _ := syscall.SyscallN(addr, hCursor)
	return ret
}

func GetCursorPos(lpPoint *POINT) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pGetCursorPos, libUser32, "GetCursorPos")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpPoint)))
	return BOOL(ret), WIN32_ERROR(err)
}

func GetPhysicalCursorPos(lpPoint *POINT) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pGetPhysicalCursorPos, libUser32, "GetPhysicalCursorPos")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpPoint)))
	return BOOL(ret), WIN32_ERROR(err)
}

func GetClipCursor(lpRect *RECT) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pGetClipCursor, libUser32, "GetClipCursor")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpRect)))
	return BOOL(ret), WIN32_ERROR(err)
}

func GetCursor() HCURSOR {
	addr := lazyAddr(&pGetCursor, libUser32, "GetCursor")
	ret, _, _ := syscall.SyscallN(addr)
	return ret
}

func CreateCaret(hWnd HWND, hBitmap HBITMAP, nWidth int32, nHeight int32) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pCreateCaret, libUser32, "CreateCaret")
	ret, _, err := syscall.SyscallN(addr, hWnd, hBitmap, uintptr(nWidth), uintptr(nHeight))
	return BOOL(ret), WIN32_ERROR(err)
}

func GetCaretBlinkTime() (uint32, WIN32_ERROR) {
	addr := lazyAddr(&pGetCaretBlinkTime, libUser32, "GetCaretBlinkTime")
	ret, _, err := syscall.SyscallN(addr)
	return uint32(ret), WIN32_ERROR(err)
}

func SetCaretBlinkTime(uMSeconds uint32) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pSetCaretBlinkTime, libUser32, "SetCaretBlinkTime")
	ret, _, err := syscall.SyscallN(addr, uintptr(uMSeconds))
	return BOOL(ret), WIN32_ERROR(err)
}

func DestroyCaret() (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pDestroyCaret, libUser32, "DestroyCaret")
	ret, _, err := syscall.SyscallN(addr)
	return BOOL(ret), WIN32_ERROR(err)
}

func HideCaret(hWnd HWND) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pHideCaret, libUser32, "HideCaret")
	ret, _, err := syscall.SyscallN(addr, hWnd)
	return BOOL(ret), WIN32_ERROR(err)
}

func ShowCaret(hWnd HWND) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pShowCaret, libUser32, "ShowCaret")
	ret, _, err := syscall.SyscallN(addr, hWnd)
	return BOOL(ret), WIN32_ERROR(err)
}

func SetCaretPos(X int32, Y int32) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pSetCaretPos, libUser32, "SetCaretPos")
	ret, _, err := syscall.SyscallN(addr, uintptr(X), uintptr(Y))
	return BOOL(ret), WIN32_ERROR(err)
}

func GetCaretPos(lpPoint *POINT) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pGetCaretPos, libUser32, "GetCaretPos")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpPoint)))
	return BOOL(ret), WIN32_ERROR(err)
}

func LogicalToPhysicalPoint(hWnd HWND, lpPoint *POINT) BOOL {
	addr := lazyAddr(&pLogicalToPhysicalPoint, libUser32, "LogicalToPhysicalPoint")
	ret, _, _ := syscall.SyscallN(addr, hWnd, uintptr(unsafe.Pointer(lpPoint)))
	return BOOL(ret)
}

func PhysicalToLogicalPoint(hWnd HWND, lpPoint *POINT) BOOL {
	addr := lazyAddr(&pPhysicalToLogicalPoint, libUser32, "PhysicalToLogicalPoint")
	ret, _, _ := syscall.SyscallN(addr, hWnd, uintptr(unsafe.Pointer(lpPoint)))
	return BOOL(ret)
}

func WindowFromPoint(Point POINT) HWND {
	addr := lazyAddr(&pWindowFromPoint, libUser32, "WindowFromPoint")
	ret, _, _ := syscall.SyscallN(addr, *(*uintptr)(unsafe.Pointer(&Point)))
	return ret
}

func WindowFromPhysicalPoint(Point POINT) HWND {
	addr := lazyAddr(&pWindowFromPhysicalPoint, libUser32, "WindowFromPhysicalPoint")
	ret, _, _ := syscall.SyscallN(addr, *(*uintptr)(unsafe.Pointer(&Point)))
	return ret
}

func ChildWindowFromPoint(hWndParent HWND, Point POINT) HWND {
	addr := lazyAddr(&pChildWindowFromPoint, libUser32, "ChildWindowFromPoint")
	ret, _, _ := syscall.SyscallN(addr, hWndParent, *(*uintptr)(unsafe.Pointer(&Point)))
	return ret
}

func ClipCursor(lpRect *RECT) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pClipCursor, libUser32, "ClipCursor")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpRect)))
	return BOOL(ret), WIN32_ERROR(err)
}

func ChildWindowFromPointEx(hwnd HWND, pt POINT, flags CWP_FLAGS) HWND {
	addr := lazyAddr(&pChildWindowFromPointEx, libUser32, "ChildWindowFromPointEx")
	ret, _, _ := syscall.SyscallN(addr, hwnd, *(*uintptr)(unsafe.Pointer(&pt)), uintptr(flags))
	return ret
}

func GetSysColor(nIndex SYS_COLOR_INDEX) uint32 {
	addr := lazyAddr(&pGetSysColor, libUser32, "GetSysColor")
	ret, _, _ := syscall.SyscallN(addr, uintptr(nIndex))
	return uint32(ret)
}

func SetSysColors(cElements int32, lpaElements *int32, lpaRgbValues *uint32) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pSetSysColors, libUser32, "SetSysColors")
	ret, _, err := syscall.SyscallN(addr, uintptr(cElements), uintptr(unsafe.Pointer(lpaElements)), uintptr(unsafe.Pointer(lpaRgbValues)))
	return BOOL(ret), WIN32_ERROR(err)
}

func GetWindowWord(hWnd HWND, nIndex int32) uint16 {
	addr := lazyAddr(&pGetWindowWord, libUser32, "GetWindowWord")
	ret, _, _ := syscall.SyscallN(addr, hWnd, uintptr(nIndex))
	return uint16(ret)
}

func SetWindowWord(hWnd HWND, nIndex int32, wNewWord uint16) uint16 {
	addr := lazyAddr(&pSetWindowWord, libUser32, "SetWindowWord")
	ret, _, _ := syscall.SyscallN(addr, hWnd, uintptr(nIndex), uintptr(wNewWord))
	return uint16(ret)
}

func GetWindowLongA(hWnd HWND, nIndex WINDOW_LONG_PTR_INDEX) (int32, WIN32_ERROR) {
	addr := lazyAddr(&pGetWindowLongA, libUser32, "GetWindowLongA")
	ret, _, err := syscall.SyscallN(addr, hWnd, uintptr(nIndex))
	return int32(ret), WIN32_ERROR(err)
}

var GetWindowLong = GetWindowLongW

func GetWindowLongW(hWnd HWND, nIndex WINDOW_LONG_PTR_INDEX) (int32, WIN32_ERROR) {
	addr := lazyAddr(&pGetWindowLongW, libUser32, "GetWindowLongW")
	ret, _, err := syscall.SyscallN(addr, hWnd, uintptr(nIndex))
	return int32(ret), WIN32_ERROR(err)
}

func SetWindowLongA(hWnd HWND, nIndex WINDOW_LONG_PTR_INDEX, dwNewLong int32) (int32, WIN32_ERROR) {
	addr := lazyAddr(&pSetWindowLongA, libUser32, "SetWindowLongA")
	ret, _, err := syscall.SyscallN(addr, hWnd, uintptr(nIndex), uintptr(dwNewLong))
	return int32(ret), WIN32_ERROR(err)
}

var SetWindowLong = SetWindowLongW

func SetWindowLongW(hWnd HWND, nIndex WINDOW_LONG_PTR_INDEX, dwNewLong int32) (int32, WIN32_ERROR) {
	addr := lazyAddr(&pSetWindowLongW, libUser32, "SetWindowLongW")
	ret, _, err := syscall.SyscallN(addr, hWnd, uintptr(nIndex), uintptr(dwNewLong))
	return int32(ret), WIN32_ERROR(err)
}

func GetWindowLongPtrA(hWnd HWND, nIndex WINDOW_LONG_PTR_INDEX) (uintptr, WIN32_ERROR) {
	addr := lazyAddr(&pGetWindowLongPtrA, libUser32, "GetWindowLongPtrA")
	ret, _, err := syscall.SyscallN(addr, hWnd, uintptr(nIndex))
	return ret, WIN32_ERROR(err)
}

var GetWindowLongPtr = GetWindowLongPtrW

func GetWindowLongPtrW(hWnd HWND, nIndex WINDOW_LONG_PTR_INDEX) (uintptr, WIN32_ERROR) {
	addr := lazyAddr(&pGetWindowLongPtrW, libUser32, "GetWindowLongPtrW")
	ret, _, err := syscall.SyscallN(addr, hWnd, uintptr(nIndex))
	return ret, WIN32_ERROR(err)
}

func SetWindowLongPtrA(hWnd HWND, nIndex WINDOW_LONG_PTR_INDEX, dwNewLong uintptr) (uintptr, WIN32_ERROR) {
	addr := lazyAddr(&pSetWindowLongPtrA, libUser32, "SetWindowLongPtrA")
	ret, _, err := syscall.SyscallN(addr, hWnd, uintptr(nIndex), dwNewLong)
	return ret, WIN32_ERROR(err)
}

var SetWindowLongPtr = SetWindowLongPtrW

func SetWindowLongPtrW(hWnd HWND, nIndex WINDOW_LONG_PTR_INDEX, dwNewLong uintptr) (uintptr, WIN32_ERROR) {
	addr := lazyAddr(&pSetWindowLongPtrW, libUser32, "SetWindowLongPtrW")
	ret, _, err := syscall.SyscallN(addr, hWnd, uintptr(nIndex), dwNewLong)
	return ret, WIN32_ERROR(err)
}

func GetClassWord(hWnd HWND, nIndex int32) (uint16, WIN32_ERROR) {
	addr := lazyAddr(&pGetClassWord, libUser32, "GetClassWord")
	ret, _, err := syscall.SyscallN(addr, hWnd, uintptr(nIndex))
	return uint16(ret), WIN32_ERROR(err)
}

func SetClassWord(hWnd HWND, nIndex int32, wNewWord uint16) (uint16, WIN32_ERROR) {
	addr := lazyAddr(&pSetClassWord, libUser32, "SetClassWord")
	ret, _, err := syscall.SyscallN(addr, hWnd, uintptr(nIndex), uintptr(wNewWord))
	return uint16(ret), WIN32_ERROR(err)
}

func GetClassLongA(hWnd HWND, nIndex GET_CLASS_LONG_INDEX) (uint32, WIN32_ERROR) {
	addr := lazyAddr(&pGetClassLongA, libUser32, "GetClassLongA")
	ret, _, err := syscall.SyscallN(addr, hWnd, uintptr(nIndex))
	return uint32(ret), WIN32_ERROR(err)
}

var GetClassLong = GetClassLongW

func GetClassLongW(hWnd HWND, nIndex GET_CLASS_LONG_INDEX) (uint32, WIN32_ERROR) {
	addr := lazyAddr(&pGetClassLongW, libUser32, "GetClassLongW")
	ret, _, err := syscall.SyscallN(addr, hWnd, uintptr(nIndex))
	return uint32(ret), WIN32_ERROR(err)
}

func SetClassLongA(hWnd HWND, nIndex GET_CLASS_LONG_INDEX, dwNewLong int32) (uint32, WIN32_ERROR) {
	addr := lazyAddr(&pSetClassLongA, libUser32, "SetClassLongA")
	ret, _, err := syscall.SyscallN(addr, hWnd, uintptr(nIndex), uintptr(dwNewLong))
	return uint32(ret), WIN32_ERROR(err)
}

var SetClassLong = SetClassLongW

func SetClassLongW(hWnd HWND, nIndex GET_CLASS_LONG_INDEX, dwNewLong int32) (uint32, WIN32_ERROR) {
	addr := lazyAddr(&pSetClassLongW, libUser32, "SetClassLongW")
	ret, _, err := syscall.SyscallN(addr, hWnd, uintptr(nIndex), uintptr(dwNewLong))
	return uint32(ret), WIN32_ERROR(err)
}

func GetClassLongPtrA(hWnd HWND, nIndex GET_CLASS_LONG_INDEX) (uintptr, WIN32_ERROR) {
	addr := lazyAddr(&pGetClassLongPtrA, libUser32, "GetClassLongPtrA")
	ret, _, err := syscall.SyscallN(addr, hWnd, uintptr(nIndex))
	return ret, WIN32_ERROR(err)
}

var GetClassLongPtr = GetClassLongPtrW

func GetClassLongPtrW(hWnd HWND, nIndex GET_CLASS_LONG_INDEX) (uintptr, WIN32_ERROR) {
	addr := lazyAddr(&pGetClassLongPtrW, libUser32, "GetClassLongPtrW")
	ret, _, err := syscall.SyscallN(addr, hWnd, uintptr(nIndex))
	return ret, WIN32_ERROR(err)
}

func SetClassLongPtrA(hWnd HWND, nIndex GET_CLASS_LONG_INDEX, dwNewLong uintptr) (uintptr, WIN32_ERROR) {
	addr := lazyAddr(&pSetClassLongPtrA, libUser32, "SetClassLongPtrA")
	ret, _, err := syscall.SyscallN(addr, hWnd, uintptr(nIndex), dwNewLong)
	return ret, WIN32_ERROR(err)
}

var SetClassLongPtr = SetClassLongPtrW

func SetClassLongPtrW(hWnd HWND, nIndex GET_CLASS_LONG_INDEX, dwNewLong uintptr) (uintptr, WIN32_ERROR) {
	addr := lazyAddr(&pSetClassLongPtrW, libUser32, "SetClassLongPtrW")
	ret, _, err := syscall.SyscallN(addr, hWnd, uintptr(nIndex), dwNewLong)
	return ret, WIN32_ERROR(err)
}

func GetProcessDefaultLayout(pdwDefaultLayout *uint32) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pGetProcessDefaultLayout, libUser32, "GetProcessDefaultLayout")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pdwDefaultLayout)))
	return BOOL(ret), WIN32_ERROR(err)
}

func SetProcessDefaultLayout(dwDefaultLayout uint32) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pSetProcessDefaultLayout, libUser32, "SetProcessDefaultLayout")
	ret, _, err := syscall.SyscallN(addr, uintptr(dwDefaultLayout))
	return BOOL(ret), WIN32_ERROR(err)
}

func GetDesktopWindow() HWND {
	addr := lazyAddr(&pGetDesktopWindow, libUser32, "GetDesktopWindow")
	ret, _, _ := syscall.SyscallN(addr)
	return ret
}

func GetParent(hWnd HWND) (HWND, WIN32_ERROR) {
	addr := lazyAddr(&pGetParent, libUser32, "GetParent")
	ret, _, err := syscall.SyscallN(addr, hWnd)
	return ret, WIN32_ERROR(err)
}

func SetParent(hWndChild HWND, hWndNewParent HWND) (HWND, WIN32_ERROR) {
	addr := lazyAddr(&pSetParent, libUser32, "SetParent")
	ret, _, err := syscall.SyscallN(addr, hWndChild, hWndNewParent)
	return ret, WIN32_ERROR(err)
}

func EnumChildWindows(hWndParent HWND, lpEnumFunc WNDENUMPROC, lParam LPARAM) BOOL {
	addr := lazyAddr(&pEnumChildWindows, libUser32, "EnumChildWindows")
	ret, _, _ := syscall.SyscallN(addr, hWndParent, lpEnumFunc, lParam)
	return BOOL(ret)
}

func FindWindowA(lpClassName PSTR, lpWindowName PSTR) (HWND, WIN32_ERROR) {
	addr := lazyAddr(&pFindWindowA, libUser32, "FindWindowA")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpClassName)), uintptr(unsafe.Pointer(lpWindowName)))
	return ret, WIN32_ERROR(err)
}

var FindWindow = FindWindowW

func FindWindowW(lpClassName PWSTR, lpWindowName PWSTR) (HWND, WIN32_ERROR) {
	addr := lazyAddr(&pFindWindowW, libUser32, "FindWindowW")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpClassName)), uintptr(unsafe.Pointer(lpWindowName)))
	return ret, WIN32_ERROR(err)
}

func FindWindowExA(hWndParent HWND, hWndChildAfter HWND, lpszClass PSTR, lpszWindow PSTR) (HWND, WIN32_ERROR) {
	addr := lazyAddr(&pFindWindowExA, libUser32, "FindWindowExA")
	ret, _, err := syscall.SyscallN(addr, hWndParent, hWndChildAfter, uintptr(unsafe.Pointer(lpszClass)), uintptr(unsafe.Pointer(lpszWindow)))
	return ret, WIN32_ERROR(err)
}

var FindWindowEx = FindWindowExW

func FindWindowExW(hWndParent HWND, hWndChildAfter HWND, lpszClass PWSTR, lpszWindow PWSTR) (HWND, WIN32_ERROR) {
	addr := lazyAddr(&pFindWindowExW, libUser32, "FindWindowExW")
	ret, _, err := syscall.SyscallN(addr, hWndParent, hWndChildAfter, uintptr(unsafe.Pointer(lpszClass)), uintptr(unsafe.Pointer(lpszWindow)))
	return ret, WIN32_ERROR(err)
}

func GetShellWindow() HWND {
	addr := lazyAddr(&pGetShellWindow, libUser32, "GetShellWindow")
	ret, _, _ := syscall.SyscallN(addr)
	return ret
}

func RegisterShellHookWindow(hwnd HWND) BOOL {
	addr := lazyAddr(&pRegisterShellHookWindow, libUser32, "RegisterShellHookWindow")
	ret, _, _ := syscall.SyscallN(addr, hwnd)
	return BOOL(ret)
}

func DeregisterShellHookWindow(hwnd HWND) BOOL {
	addr := lazyAddr(&pDeregisterShellHookWindow, libUser32, "DeregisterShellHookWindow")
	ret, _, _ := syscall.SyscallN(addr, hwnd)
	return BOOL(ret)
}

func EnumWindows(lpEnumFunc WNDENUMPROC, lParam LPARAM) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pEnumWindows, libUser32, "EnumWindows")
	ret, _, err := syscall.SyscallN(addr, lpEnumFunc, lParam)
	return BOOL(ret), WIN32_ERROR(err)
}

func EnumThreadWindows(dwThreadId uint32, lpfn WNDENUMPROC, lParam LPARAM) BOOL {
	addr := lazyAddr(&pEnumThreadWindows, libUser32, "EnumThreadWindows")
	ret, _, _ := syscall.SyscallN(addr, uintptr(dwThreadId), lpfn, lParam)
	return BOOL(ret)
}

func GetClassNameA(hWnd HWND, lpClassName PSTR, nMaxCount int32) (int32, WIN32_ERROR) {
	addr := lazyAddr(&pGetClassNameA, libUser32, "GetClassNameA")
	ret, _, err := syscall.SyscallN(addr, hWnd, uintptr(unsafe.Pointer(lpClassName)), uintptr(nMaxCount))
	return int32(ret), WIN32_ERROR(err)
}

var GetClassName = GetClassNameW

func GetClassNameW(hWnd HWND, lpClassName PWSTR, nMaxCount int32) (int32, WIN32_ERROR) {
	addr := lazyAddr(&pGetClassNameW, libUser32, "GetClassNameW")
	ret, _, err := syscall.SyscallN(addr, hWnd, uintptr(unsafe.Pointer(lpClassName)), uintptr(nMaxCount))
	return int32(ret), WIN32_ERROR(err)
}

func GetTopWindow(hWnd HWND) (HWND, WIN32_ERROR) {
	addr := lazyAddr(&pGetTopWindow, libUser32, "GetTopWindow")
	ret, _, err := syscall.SyscallN(addr, hWnd)
	return ret, WIN32_ERROR(err)
}

func GetWindowThreadProcessId(hWnd HWND, lpdwProcessId *uint32) uint32 {
	addr := lazyAddr(&pGetWindowThreadProcessId, libUser32, "GetWindowThreadProcessId")
	ret, _, _ := syscall.SyscallN(addr, hWnd, uintptr(unsafe.Pointer(lpdwProcessId)))
	return uint32(ret)
}

func IsGUIThread(bConvert BOOL) BOOL {
	addr := lazyAddr(&pIsGUIThread, libUser32, "IsGUIThread")
	ret, _, _ := syscall.SyscallN(addr, uintptr(bConvert))
	return BOOL(ret)
}

func GetLastActivePopup(hWnd HWND) HWND {
	addr := lazyAddr(&pGetLastActivePopup, libUser32, "GetLastActivePopup")
	ret, _, _ := syscall.SyscallN(addr, hWnd)
	return ret
}

func GetWindow(hWnd HWND, uCmd GET_WINDOW_CMD) (HWND, WIN32_ERROR) {
	addr := lazyAddr(&pGetWindow, libUser32, "GetWindow")
	ret, _, err := syscall.SyscallN(addr, hWnd, uintptr(uCmd))
	return ret, WIN32_ERROR(err)
}

func SetWindowsHookA(nFilterType int32, pfnFilterProc HOOKPROC) HHOOK {
	addr := lazyAddr(&pSetWindowsHookA, libUser32, "SetWindowsHookA")
	ret, _, _ := syscall.SyscallN(addr, uintptr(nFilterType), pfnFilterProc)
	return ret
}

var SetWindowsHook = SetWindowsHookW

func SetWindowsHookW(nFilterType int32, pfnFilterProc HOOKPROC) HHOOK {
	addr := lazyAddr(&pSetWindowsHookW, libUser32, "SetWindowsHookW")
	ret, _, _ := syscall.SyscallN(addr, uintptr(nFilterType), pfnFilterProc)
	return ret
}

func UnhookWindowsHook(nCode int32, pfnFilterProc HOOKPROC) BOOL {
	addr := lazyAddr(&pUnhookWindowsHook, libUser32, "UnhookWindowsHook")
	ret, _, _ := syscall.SyscallN(addr, uintptr(nCode), pfnFilterProc)
	return BOOL(ret)
}

func SetWindowsHookExA(idHook WINDOWS_HOOK_ID, lpfn HOOKPROC, hmod HINSTANCE, dwThreadId uint32) (HHOOK, WIN32_ERROR) {
	addr := lazyAddr(&pSetWindowsHookExA, libUser32, "SetWindowsHookExA")
	ret, _, err := syscall.SyscallN(addr, uintptr(idHook), lpfn, hmod, uintptr(dwThreadId))
	return ret, WIN32_ERROR(err)
}

var SetWindowsHookEx = SetWindowsHookExW

func SetWindowsHookExW(idHook WINDOWS_HOOK_ID, lpfn HOOKPROC, hmod HINSTANCE, dwThreadId uint32) (HHOOK, WIN32_ERROR) {
	addr := lazyAddr(&pSetWindowsHookExW, libUser32, "SetWindowsHookExW")
	ret, _, err := syscall.SyscallN(addr, uintptr(idHook), lpfn, hmod, uintptr(dwThreadId))
	return ret, WIN32_ERROR(err)
}

func UnhookWindowsHookEx(hhk HHOOK) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pUnhookWindowsHookEx, libUser32, "UnhookWindowsHookEx")
	ret, _, err := syscall.SyscallN(addr, hhk)
	return BOOL(ret), WIN32_ERROR(err)
}

func CallNextHookEx(hhk HHOOK, nCode int32, wParam WPARAM, lParam LPARAM) LRESULT {
	addr := lazyAddr(&pCallNextHookEx, libUser32, "CallNextHookEx")
	ret, _, _ := syscall.SyscallN(addr, hhk, uintptr(nCode), wParam, lParam)
	return ret
}

func CheckMenuRadioItem(hmenu HMENU, first uint32, last uint32, check uint32, flags uint32) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pCheckMenuRadioItem, libUser32, "CheckMenuRadioItem")
	ret, _, err := syscall.SyscallN(addr, hmenu, uintptr(first), uintptr(last), uintptr(check), uintptr(flags))
	return BOOL(ret), WIN32_ERROR(err)
}

func LoadCursorA(hInstance HINSTANCE, lpCursorName PSTR) (HCURSOR, WIN32_ERROR) {
	addr := lazyAddr(&pLoadCursorA, libUser32, "LoadCursorA")
	ret, _, err := syscall.SyscallN(addr, hInstance, uintptr(unsafe.Pointer(lpCursorName)))
	return ret, WIN32_ERROR(err)
}

var LoadCursor = LoadCursorW

func LoadCursorW(hInstance HINSTANCE, lpCursorName PWSTR) (HCURSOR, WIN32_ERROR) {
	addr := lazyAddr(&pLoadCursorW, libUser32, "LoadCursorW")
	ret, _, err := syscall.SyscallN(addr, hInstance, uintptr(unsafe.Pointer(lpCursorName)))
	return ret, WIN32_ERROR(err)
}

func LoadCursorFromFileA(lpFileName PSTR) (HCURSOR, WIN32_ERROR) {
	addr := lazyAddr(&pLoadCursorFromFileA, libUser32, "LoadCursorFromFileA")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpFileName)))
	return ret, WIN32_ERROR(err)
}

var LoadCursorFromFile = LoadCursorFromFileW

func LoadCursorFromFileW(lpFileName PWSTR) (HCURSOR, WIN32_ERROR) {
	addr := lazyAddr(&pLoadCursorFromFileW, libUser32, "LoadCursorFromFileW")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpFileName)))
	return ret, WIN32_ERROR(err)
}

func CreateCursor(hInst HINSTANCE, xHotSpot int32, yHotSpot int32, nWidth int32, nHeight int32, pvANDPlane unsafe.Pointer, pvXORPlane unsafe.Pointer) (HCURSOR, WIN32_ERROR) {
	addr := lazyAddr(&pCreateCursor, libUser32, "CreateCursor")
	ret, _, err := syscall.SyscallN(addr, hInst, uintptr(xHotSpot), uintptr(yHotSpot), uintptr(nWidth), uintptr(nHeight), uintptr(pvANDPlane), uintptr(pvXORPlane))
	return ret, WIN32_ERROR(err)
}

func DestroyCursor(hCursor HCURSOR) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pDestroyCursor, libUser32, "DestroyCursor")
	ret, _, err := syscall.SyscallN(addr, hCursor)
	return BOOL(ret), WIN32_ERROR(err)
}

func SetSystemCursor(hcur HCURSOR, id SYSTEM_CURSOR_ID) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pSetSystemCursor, libUser32, "SetSystemCursor")
	ret, _, err := syscall.SyscallN(addr, hcur, uintptr(id))
	return BOOL(ret), WIN32_ERROR(err)
}

func LoadIconA(hInstance HINSTANCE, lpIconName PSTR) (HICON, WIN32_ERROR) {
	addr := lazyAddr(&pLoadIconA, libUser32, "LoadIconA")
	ret, _, err := syscall.SyscallN(addr, hInstance, uintptr(unsafe.Pointer(lpIconName)))
	return ret, WIN32_ERROR(err)
}

var LoadIcon = LoadIconW

func LoadIconW(hInstance HINSTANCE, lpIconName PWSTR) (HICON, WIN32_ERROR) {
	addr := lazyAddr(&pLoadIconW, libUser32, "LoadIconW")
	ret, _, err := syscall.SyscallN(addr, hInstance, uintptr(unsafe.Pointer(lpIconName)))
	return ret, WIN32_ERROR(err)
}

func PrivateExtractIconsA(szFileName PSTR, nIconIndex int32, cxIcon int32, cyIcon int32, phicon *HICON, piconid *uint32, nIcons uint32, flags uint32) uint32 {
	addr := lazyAddr(&pPrivateExtractIconsA, libUser32, "PrivateExtractIconsA")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(szFileName)), uintptr(nIconIndex), uintptr(cxIcon), uintptr(cyIcon), uintptr(unsafe.Pointer(phicon)), uintptr(unsafe.Pointer(piconid)), uintptr(nIcons), uintptr(flags))
	return uint32(ret)
}

var PrivateExtractIcons = PrivateExtractIconsW

func PrivateExtractIconsW(szFileName PWSTR, nIconIndex int32, cxIcon int32, cyIcon int32, phicon *HICON, piconid *uint32, nIcons uint32, flags uint32) uint32 {
	addr := lazyAddr(&pPrivateExtractIconsW, libUser32, "PrivateExtractIconsW")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(szFileName)), uintptr(nIconIndex), uintptr(cxIcon), uintptr(cyIcon), uintptr(unsafe.Pointer(phicon)), uintptr(unsafe.Pointer(piconid)), uintptr(nIcons), uintptr(flags))
	return uint32(ret)
}

func CreateIcon(hInstance HINSTANCE, nWidth int32, nHeight int32, cPlanes byte, cBitsPixel byte, lpbANDbits *byte, lpbXORbits *byte) (HICON, WIN32_ERROR) {
	addr := lazyAddr(&pCreateIcon, libUser32, "CreateIcon")
	ret, _, err := syscall.SyscallN(addr, hInstance, uintptr(nWidth), uintptr(nHeight), uintptr(cPlanes), uintptr(cBitsPixel), uintptr(unsafe.Pointer(lpbANDbits)), uintptr(unsafe.Pointer(lpbXORbits)))
	return ret, WIN32_ERROR(err)
}

func DestroyIcon(hIcon HICON) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pDestroyIcon, libUser32, "DestroyIcon")
	ret, _, err := syscall.SyscallN(addr, hIcon)
	return BOOL(ret), WIN32_ERROR(err)
}

func LookupIconIdFromDirectory(presbits *byte, fIcon BOOL) (int32, WIN32_ERROR) {
	addr := lazyAddr(&pLookupIconIdFromDirectory, libUser32, "LookupIconIdFromDirectory")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(presbits)), uintptr(fIcon))
	return int32(ret), WIN32_ERROR(err)
}

func LookupIconIdFromDirectoryEx(presbits *byte, fIcon BOOL, cxDesired int32, cyDesired int32, Flags IMAGE_FLAGS) (int32, WIN32_ERROR) {
	addr := lazyAddr(&pLookupIconIdFromDirectoryEx, libUser32, "LookupIconIdFromDirectoryEx")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(presbits)), uintptr(fIcon), uintptr(cxDesired), uintptr(cyDesired), uintptr(Flags))
	return int32(ret), WIN32_ERROR(err)
}

func CreateIconFromResource(presbits *byte, dwResSize uint32, fIcon BOOL, dwVer uint32) (HICON, WIN32_ERROR) {
	addr := lazyAddr(&pCreateIconFromResource, libUser32, "CreateIconFromResource")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(presbits)), uintptr(dwResSize), uintptr(fIcon), uintptr(dwVer))
	return ret, WIN32_ERROR(err)
}

func CreateIconFromResourceEx(presbits *byte, dwResSize uint32, fIcon BOOL, dwVer uint32, cxDesired int32, cyDesired int32, Flags IMAGE_FLAGS) (HICON, WIN32_ERROR) {
	addr := lazyAddr(&pCreateIconFromResourceEx, libUser32, "CreateIconFromResourceEx")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(presbits)), uintptr(dwResSize), uintptr(fIcon), uintptr(dwVer), uintptr(cxDesired), uintptr(cyDesired), uintptr(Flags))
	return ret, WIN32_ERROR(err)
}

func LoadImageA(hInst HINSTANCE, name PSTR, type_ GDI_IMAGE_TYPE, cx int32, cy int32, fuLoad IMAGE_FLAGS) (HANDLE, WIN32_ERROR) {
	addr := lazyAddr(&pLoadImageA, libUser32, "LoadImageA")
	ret, _, err := syscall.SyscallN(addr, hInst, uintptr(unsafe.Pointer(name)), uintptr(type_), uintptr(cx), uintptr(cy), uintptr(fuLoad))
	return ret, WIN32_ERROR(err)
}

var LoadImage = LoadImageW

func LoadImageW(hInst HINSTANCE, name PWSTR, type_ GDI_IMAGE_TYPE, cx int32, cy int32, fuLoad IMAGE_FLAGS) (HANDLE, WIN32_ERROR) {
	addr := lazyAddr(&pLoadImageW, libUser32, "LoadImageW")
	ret, _, err := syscall.SyscallN(addr, hInst, uintptr(unsafe.Pointer(name)), uintptr(type_), uintptr(cx), uintptr(cy), uintptr(fuLoad))
	return ret, WIN32_ERROR(err)
}

func CopyImage(h HANDLE, type_ GDI_IMAGE_TYPE, cx int32, cy int32, flags IMAGE_FLAGS) (HANDLE, WIN32_ERROR) {
	addr := lazyAddr(&pCopyImage, libUser32, "CopyImage")
	ret, _, err := syscall.SyscallN(addr, h, uintptr(type_), uintptr(cx), uintptr(cy), uintptr(flags))
	return ret, WIN32_ERROR(err)
}

func DrawIconEx(hdc HDC, xLeft int32, yTop int32, hIcon HICON, cxWidth int32, cyWidth int32, istepIfAniCur uint32, hbrFlickerFreeDraw HBRUSH, diFlags DI_FLAGS) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pDrawIconEx, libUser32, "DrawIconEx")
	ret, _, err := syscall.SyscallN(addr, hdc, uintptr(xLeft), uintptr(yTop), hIcon, uintptr(cxWidth), uintptr(cyWidth), uintptr(istepIfAniCur), hbrFlickerFreeDraw, uintptr(diFlags))
	return BOOL(ret), WIN32_ERROR(err)
}

func CreateIconIndirect(piconinfo *ICONINFO) (HICON, WIN32_ERROR) {
	addr := lazyAddr(&pCreateIconIndirect, libUser32, "CreateIconIndirect")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(piconinfo)))
	return ret, WIN32_ERROR(err)
}

func CopyIcon(hIcon HICON) (HICON, WIN32_ERROR) {
	addr := lazyAddr(&pCopyIcon, libUser32, "CopyIcon")
	ret, _, err := syscall.SyscallN(addr, hIcon)
	return ret, WIN32_ERROR(err)
}

func GetIconInfo(hIcon HICON, piconinfo *ICONINFO) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pGetIconInfo, libUser32, "GetIconInfo")
	ret, _, err := syscall.SyscallN(addr, hIcon, uintptr(unsafe.Pointer(piconinfo)))
	return BOOL(ret), WIN32_ERROR(err)
}

func GetIconInfoExA(hicon HICON, piconinfo *ICONINFOEXA) BOOL {
	addr := lazyAddr(&pGetIconInfoExA, libUser32, "GetIconInfoExA")
	ret, _, _ := syscall.SyscallN(addr, hicon, uintptr(unsafe.Pointer(piconinfo)))
	return BOOL(ret)
}

var GetIconInfoEx = GetIconInfoExW

func GetIconInfoExW(hicon HICON, piconinfo *ICONINFOEXW) BOOL {
	addr := lazyAddr(&pGetIconInfoExW, libUser32, "GetIconInfoExW")
	ret, _, _ := syscall.SyscallN(addr, hicon, uintptr(unsafe.Pointer(piconinfo)))
	return BOOL(ret)
}

func IsDialogMessageA(hDlg HWND, lpMsg *MSG) BOOL {
	addr := lazyAddr(&pIsDialogMessageA, libUser32, "IsDialogMessageA")
	ret, _, _ := syscall.SyscallN(addr, hDlg, uintptr(unsafe.Pointer(lpMsg)))
	return BOOL(ret)
}

var IsDialogMessage = IsDialogMessageW

func IsDialogMessageW(hDlg HWND, lpMsg *MSG) BOOL {
	addr := lazyAddr(&pIsDialogMessageW, libUser32, "IsDialogMessageW")
	ret, _, _ := syscall.SyscallN(addr, hDlg, uintptr(unsafe.Pointer(lpMsg)))
	return BOOL(ret)
}

func MapDialogRect(hDlg HWND, lpRect *RECT) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pMapDialogRect, libUser32, "MapDialogRect")
	ret, _, err := syscall.SyscallN(addr, hDlg, uintptr(unsafe.Pointer(lpRect)))
	return BOOL(ret), WIN32_ERROR(err)
}

func GetScrollInfo(hwnd HWND, nBar SCROLLBAR_CONSTANTS, lpsi *SCROLLINFO) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pGetScrollInfo, libUser32, "GetScrollInfo")
	ret, _, err := syscall.SyscallN(addr, hwnd, uintptr(nBar), uintptr(unsafe.Pointer(lpsi)))
	return BOOL(ret), WIN32_ERROR(err)
}

func DefFrameProcA(hWnd HWND, hWndMDIClient HWND, uMsg uint32, wParam WPARAM, lParam LPARAM) LRESULT {
	addr := lazyAddr(&pDefFrameProcA, libUser32, "DefFrameProcA")
	ret, _, _ := syscall.SyscallN(addr, hWnd, hWndMDIClient, uintptr(uMsg), wParam, lParam)
	return ret
}

var DefFrameProc = DefFrameProcW

func DefFrameProcW(hWnd HWND, hWndMDIClient HWND, uMsg uint32, wParam WPARAM, lParam LPARAM) LRESULT {
	addr := lazyAddr(&pDefFrameProcW, libUser32, "DefFrameProcW")
	ret, _, _ := syscall.SyscallN(addr, hWnd, hWndMDIClient, uintptr(uMsg), wParam, lParam)
	return ret
}

func DefMDIChildProcA(hWnd HWND, uMsg uint32, wParam WPARAM, lParam LPARAM) LRESULT {
	addr := lazyAddr(&pDefMDIChildProcA, libUser32, "DefMDIChildProcA")
	ret, _, _ := syscall.SyscallN(addr, hWnd, uintptr(uMsg), wParam, lParam)
	return ret
}

var DefMDIChildProc = DefMDIChildProcW

func DefMDIChildProcW(hWnd HWND, uMsg uint32, wParam WPARAM, lParam LPARAM) LRESULT {
	addr := lazyAddr(&pDefMDIChildProcW, libUser32, "DefMDIChildProcW")
	ret, _, _ := syscall.SyscallN(addr, hWnd, uintptr(uMsg), wParam, lParam)
	return ret
}

func TranslateMDISysAccel(hWndClient HWND, lpMsg *MSG) BOOL {
	addr := lazyAddr(&pTranslateMDISysAccel, libUser32, "TranslateMDISysAccel")
	ret, _, _ := syscall.SyscallN(addr, hWndClient, uintptr(unsafe.Pointer(lpMsg)))
	return BOOL(ret)
}

func ArrangeIconicWindows(hWnd HWND) (uint32, WIN32_ERROR) {
	addr := lazyAddr(&pArrangeIconicWindows, libUser32, "ArrangeIconicWindows")
	ret, _, err := syscall.SyscallN(addr, hWnd)
	return uint32(ret), WIN32_ERROR(err)
}

func CreateMDIWindowA(lpClassName PSTR, lpWindowName PSTR, dwStyle WINDOW_STYLE, X int32, Y int32, nWidth int32, nHeight int32, hWndParent HWND, hInstance HINSTANCE, lParam LPARAM) (HWND, WIN32_ERROR) {
	addr := lazyAddr(&pCreateMDIWindowA, libUser32, "CreateMDIWindowA")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpClassName)), uintptr(unsafe.Pointer(lpWindowName)), uintptr(dwStyle), uintptr(X), uintptr(Y), uintptr(nWidth), uintptr(nHeight), hWndParent, hInstance, lParam)
	return ret, WIN32_ERROR(err)
}

var CreateMDIWindow = CreateMDIWindowW

func CreateMDIWindowW(lpClassName PWSTR, lpWindowName PWSTR, dwStyle WINDOW_STYLE, X int32, Y int32, nWidth int32, nHeight int32, hWndParent HWND, hInstance HINSTANCE, lParam LPARAM) (HWND, WIN32_ERROR) {
	addr := lazyAddr(&pCreateMDIWindowW, libUser32, "CreateMDIWindowW")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpClassName)), uintptr(unsafe.Pointer(lpWindowName)), uintptr(dwStyle), uintptr(X), uintptr(Y), uintptr(nWidth), uintptr(nHeight), hWndParent, hInstance, lParam)
	return ret, WIN32_ERROR(err)
}

func TileWindows(hwndParent HWND, wHow TILE_WINDOWS_HOW, lpRect *RECT, cKids uint32, lpKids *HWND) (uint16, WIN32_ERROR) {
	addr := lazyAddr(&pTileWindows, libUser32, "TileWindows")
	ret, _, err := syscall.SyscallN(addr, hwndParent, uintptr(wHow), uintptr(unsafe.Pointer(lpRect)), uintptr(cKids), uintptr(unsafe.Pointer(lpKids)))
	return uint16(ret), WIN32_ERROR(err)
}

func CascadeWindows(hwndParent HWND, wHow CASCADE_WINDOWS_HOW, lpRect *RECT, cKids uint32, lpKids *HWND) (uint16, WIN32_ERROR) {
	addr := lazyAddr(&pCascadeWindows, libUser32, "CascadeWindows")
	ret, _, err := syscall.SyscallN(addr, hwndParent, uintptr(wHow), uintptr(unsafe.Pointer(lpRect)), uintptr(cKids), uintptr(unsafe.Pointer(lpKids)))
	return uint16(ret), WIN32_ERROR(err)
}

func SystemParametersInfoA(uiAction SYSTEM_PARAMETERS_INFO_ACTION, uiParam uint32, pvParam unsafe.Pointer, fWinIni SYSTEM_PARAMETERS_INFO_UPDATE_FLAGS) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pSystemParametersInfoA, libUser32, "SystemParametersInfoA")
	ret, _, err := syscall.SyscallN(addr, uintptr(uiAction), uintptr(uiParam), uintptr(pvParam), uintptr(fWinIni))
	return BOOL(ret), WIN32_ERROR(err)
}

var SystemParametersInfo = SystemParametersInfoW

func SystemParametersInfoW(uiAction SYSTEM_PARAMETERS_INFO_ACTION, uiParam uint32, pvParam unsafe.Pointer, fWinIni SYSTEM_PARAMETERS_INFO_UPDATE_FLAGS) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pSystemParametersInfoW, libUser32, "SystemParametersInfoW")
	ret, _, err := syscall.SyscallN(addr, uintptr(uiAction), uintptr(uiParam), uintptr(pvParam), uintptr(fWinIni))
	return BOOL(ret), WIN32_ERROR(err)
}

func SoundSentry() BOOL {
	addr := lazyAddr(&pSoundSentry, libUser32, "SoundSentry")
	ret, _, _ := syscall.SyscallN(addr)
	return BOOL(ret)
}

func SetDebugErrorLevel(dwLevel uint32) {
	addr := lazyAddr(&pSetDebugErrorLevel, libUser32, "SetDebugErrorLevel")
	syscall.SyscallN(addr, uintptr(dwLevel))
}

func InternalGetWindowText(hWnd HWND, pString PWSTR, cchMaxCount int32) (int32, WIN32_ERROR) {
	addr := lazyAddr(&pInternalGetWindowText, libUser32, "InternalGetWindowText")
	ret, _, err := syscall.SyscallN(addr, hWnd, uintptr(unsafe.Pointer(pString)), uintptr(cchMaxCount))
	return int32(ret), WIN32_ERROR(err)
}

func CancelShutdown() BOOL {
	addr := lazyAddr(&pCancelShutdown, libUser32, "CancelShutdown")
	ret, _, _ := syscall.SyscallN(addr)
	return BOOL(ret)
}

func GetGUIThreadInfo(idThread uint32, pgui *GUITHREADINFO) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pGetGUIThreadInfo, libUser32, "GetGUIThreadInfo")
	ret, _, err := syscall.SyscallN(addr, uintptr(idThread), uintptr(unsafe.Pointer(pgui)))
	return BOOL(ret), WIN32_ERROR(err)
}

func SetProcessDPIAware() BOOL {
	addr := lazyAddr(&pSetProcessDPIAware, libUser32, "SetProcessDPIAware")
	ret, _, _ := syscall.SyscallN(addr)
	return BOOL(ret)
}

func IsProcessDPIAware() BOOL {
	addr := lazyAddr(&pIsProcessDPIAware, libUser32, "IsProcessDPIAware")
	ret, _, _ := syscall.SyscallN(addr)
	return BOOL(ret)
}

func InheritWindowMonitor(hwnd HWND, hwndInherit HWND) BOOL {
	addr := lazyAddr(&pInheritWindowMonitor, libUser32, "InheritWindowMonitor")
	ret, _, _ := syscall.SyscallN(addr, hwnd, hwndInherit)
	return BOOL(ret)
}

func GetWindowModuleFileNameA(hwnd HWND, pszFileName PSTR, cchFileNameMax uint32) uint32 {
	addr := lazyAddr(&pGetWindowModuleFileNameA, libUser32, "GetWindowModuleFileNameA")
	ret, _, _ := syscall.SyscallN(addr, hwnd, uintptr(unsafe.Pointer(pszFileName)), uintptr(cchFileNameMax))
	return uint32(ret)
}

var GetWindowModuleFileName = GetWindowModuleFileNameW

func GetWindowModuleFileNameW(hwnd HWND, pszFileName PWSTR, cchFileNameMax uint32) uint32 {
	addr := lazyAddr(&pGetWindowModuleFileNameW, libUser32, "GetWindowModuleFileNameW")
	ret, _, _ := syscall.SyscallN(addr, hwnd, uintptr(unsafe.Pointer(pszFileName)), uintptr(cchFileNameMax))
	return uint32(ret)
}

func GetCursorInfo(pci *CURSORINFO) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pGetCursorInfo, libUser32, "GetCursorInfo")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pci)))
	return BOOL(ret), WIN32_ERROR(err)
}

func GetWindowInfo(hwnd HWND, pwi *WINDOWINFO) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pGetWindowInfo, libUser32, "GetWindowInfo")
	ret, _, err := syscall.SyscallN(addr, hwnd, uintptr(unsafe.Pointer(pwi)))
	return BOOL(ret), WIN32_ERROR(err)
}

func GetTitleBarInfo(hwnd HWND, pti *TITLEBARINFO) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pGetTitleBarInfo, libUser32, "GetTitleBarInfo")
	ret, _, err := syscall.SyscallN(addr, hwnd, uintptr(unsafe.Pointer(pti)))
	return BOOL(ret), WIN32_ERROR(err)
}

func GetMenuBarInfo(hwnd HWND, idObject OBJECT_IDENTIFIER, idItem int32, pmbi *MENUBARINFO) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pGetMenuBarInfo, libUser32, "GetMenuBarInfo")
	ret, _, err := syscall.SyscallN(addr, hwnd, uintptr(idObject), uintptr(idItem), uintptr(unsafe.Pointer(pmbi)))
	return BOOL(ret), WIN32_ERROR(err)
}

func GetScrollBarInfo(hwnd HWND, idObject OBJECT_IDENTIFIER, psbi *SCROLLBARINFO) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pGetScrollBarInfo, libUser32, "GetScrollBarInfo")
	ret, _, err := syscall.SyscallN(addr, hwnd, uintptr(idObject), uintptr(unsafe.Pointer(psbi)))
	return BOOL(ret), WIN32_ERROR(err)
}

func GetAncestor(hwnd HWND, gaFlags GET_ANCESTOR_FLAGS) HWND {
	addr := lazyAddr(&pGetAncestor, libUser32, "GetAncestor")
	ret, _, _ := syscall.SyscallN(addr, hwnd, uintptr(gaFlags))
	return ret
}

func RealChildWindowFromPoint(hwndParent HWND, ptParentClientCoords POINT) HWND {
	addr := lazyAddr(&pRealChildWindowFromPoint, libUser32, "RealChildWindowFromPoint")
	ret, _, _ := syscall.SyscallN(addr, hwndParent, *(*uintptr)(unsafe.Pointer(&ptParentClientCoords)))
	return ret
}

func RealGetWindowClassA(hwnd HWND, ptszClassName PSTR, cchClassNameMax uint32) uint32 {
	addr := lazyAddr(&pRealGetWindowClassA, libUser32, "RealGetWindowClassA")
	ret, _, _ := syscall.SyscallN(addr, hwnd, uintptr(unsafe.Pointer(ptszClassName)), uintptr(cchClassNameMax))
	return uint32(ret)
}

var RealGetWindowClass = RealGetWindowClassW

func RealGetWindowClassW(hwnd HWND, ptszClassName PWSTR, cchClassNameMax uint32) (uint32, WIN32_ERROR) {
	addr := lazyAddr(&pRealGetWindowClassW, libUser32, "RealGetWindowClassW")
	ret, _, err := syscall.SyscallN(addr, hwnd, uintptr(unsafe.Pointer(ptszClassName)), uintptr(cchClassNameMax))
	return uint32(ret), WIN32_ERROR(err)
}

func GetAltTabInfoA(hwnd HWND, iItem int32, pati *ALTTABINFO, pszItemText PSTR, cchItemText uint32) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pGetAltTabInfoA, libUser32, "GetAltTabInfoA")
	ret, _, err := syscall.SyscallN(addr, hwnd, uintptr(iItem), uintptr(unsafe.Pointer(pati)), uintptr(unsafe.Pointer(pszItemText)), uintptr(cchItemText))
	return BOOL(ret), WIN32_ERROR(err)
}

var GetAltTabInfo = GetAltTabInfoW

func GetAltTabInfoW(hwnd HWND, iItem int32, pati *ALTTABINFO, pszItemText PWSTR, cchItemText uint32) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pGetAltTabInfoW, libUser32, "GetAltTabInfoW")
	ret, _, err := syscall.SyscallN(addr, hwnd, uintptr(iItem), uintptr(unsafe.Pointer(pati)), uintptr(unsafe.Pointer(pszItemText)), uintptr(cchItemText))
	return BOOL(ret), WIN32_ERROR(err)
}

func ChangeWindowMessageFilter(message uint32, dwFlag CHANGE_WINDOW_MESSAGE_FILTER_FLAGS) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pChangeWindowMessageFilter, libUser32, "ChangeWindowMessageFilter")
	ret, _, err := syscall.SyscallN(addr, uintptr(message), uintptr(dwFlag))
	return BOOL(ret), WIN32_ERROR(err)
}

func ChangeWindowMessageFilterEx(hwnd HWND, message uint32, action WINDOW_MESSAGE_FILTER_ACTION, pChangeFilterStruct *CHANGEFILTERSTRUCT) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pChangeWindowMessageFilterEx, libUser32, "ChangeWindowMessageFilterEx")
	ret, _, err := syscall.SyscallN(addr, hwnd, uintptr(message), uintptr(action), uintptr(unsafe.Pointer(pChangeFilterStruct)))
	return BOOL(ret), WIN32_ERROR(err)
}
