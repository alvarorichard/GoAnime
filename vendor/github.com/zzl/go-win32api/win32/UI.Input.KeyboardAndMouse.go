package win32

import (
	"syscall"
	"unsafe"
)

const (
	EXTENDED_BIT                      uint32 = 0x1000000
	DONTCARE_BIT                      uint32 = 0x2000000
	FAKE_KEYSTROKE                    uint32 = 0x2000000
	KBDBASE                           uint32 = 0x0
	KBDSHIFT                          uint32 = 0x1
	KBDCTRL                           uint32 = 0x2
	KBDALT                            uint32 = 0x4
	KBDKANA                           uint32 = 0x8
	KBDROYA                           uint32 = 0x10
	KBDLOYA                           uint32 = 0x20
	KBDGRPSELTAP                      uint32 = 0x80
	GRAVE                             uint32 = 0x300
	ACUTE                             uint32 = 0x301
	CIRCUMFLEX                        uint32 = 0x302
	TILDE                             uint32 = 0x303
	MACRON                            uint32 = 0x304
	OVERSCORE                         uint32 = 0x305
	BREVE                             uint32 = 0x306
	DOT_ABOVE                         uint32 = 0x307
	UMLAUT                            uint32 = 0x308
	DIARESIS                          uint32 = 0x308
	HOOK_ABOVE                        uint32 = 0x309
	RING                              uint32 = 0x30a
	DOUBLE_ACUTE                      uint32 = 0x30b
	HACEK                             uint32 = 0x30c
	CEDILLA                           uint32 = 0x327
	OGONEK                            uint32 = 0x328
	TONOS                             uint32 = 0x384
	DIARESIS_TONOS                    uint32 = 0x385
	WszGRAVE                          string = "̀"
	WszACUTE                          string = "́"
	WszCIRCUMFLEX                     string = "̂"
	WszTILDE                          string = "̃"
	WszMACRON                         string = "̄"
	WszOVERSCORE                      string = "̅"
	WszBREVE                          string = "̆"
	WszDOT_ABOVE                      string = "̇"
	WszUMLAUT                         string = "̈"
	WszHOOK_ABOVE                     string = "̉"
	WszRING                           string = "̊"
	WszDOUBLE_ACUTE                   string = "̋"
	WszHACEK                          string = "̌"
	WszCEDILLA                        string = "̧"
	WszOGONEK                         string = "̨"
	WszTONOS                          string = "΄"
	WszDIARESIS_TONOS                 string = "΅"
	SHFT_INVALID                      uint32 = 0xf
	WCH_NONE                          uint32 = 0xf000
	WCH_DEAD                          uint32 = 0xf001
	WCH_LGTR                          uint32 = 0xf002
	CAPLOK                            uint32 = 0x1
	SGCAPS                            uint32 = 0x2
	CAPLOKALTGR                       uint32 = 0x4
	KANALOK                           uint32 = 0x8
	GRPSELTAP                         uint32 = 0x80
	DKF_DEAD                          uint32 = 0x1
	KBD_VERSION                       uint32 = 0x1
	KLLF_ALTGR                        uint32 = 0x1
	KLLF_SHIFTLOCK                    uint32 = 0x2
	KLLF_LRM_RLM                      uint32 = 0x4
	KLLF_GLOBAL_ATTRS                 uint32 = 0x2
	KBDTABLE_MULTI_MAX                uint32 = 0x8
	KEYBOARD_TYPE_GENERIC_101         uint32 = 0x4
	KEYBOARD_TYPE_JAPAN               uint32 = 0x7
	KEYBOARD_TYPE_KOREA               uint32 = 0x8
	KEYBOARD_TYPE_UNKNOWN             uint32 = 0x51
	NLSKBD_OEM_MICROSOFT              uint32 = 0x0
	NLSKBD_OEM_AX                     uint32 = 0x1
	NLSKBD_OEM_EPSON                  uint32 = 0x4
	NLSKBD_OEM_FUJITSU                uint32 = 0x5
	NLSKBD_OEM_IBM                    uint32 = 0x7
	NLSKBD_OEM_MATSUSHITA             uint32 = 0xa
	NLSKBD_OEM_NEC                    uint32 = 0xd
	NLSKBD_OEM_TOSHIBA                uint32 = 0x12
	NLSKBD_OEM_DEC                    uint32 = 0x18
	MICROSOFT_KBD_101_TYPE            uint32 = 0x0
	MICROSOFT_KBD_AX_TYPE             uint32 = 0x1
	MICROSOFT_KBD_106_TYPE            uint32 = 0x2
	MICROSOFT_KBD_002_TYPE            uint32 = 0x3
	MICROSOFT_KBD_001_TYPE            uint32 = 0x4
	MICROSOFT_KBD_FUNC                uint32 = 0xc
	AX_KBD_DESKTOP_TYPE               uint32 = 0x1
	FMR_KBD_JIS_TYPE                  uint32 = 0x0
	FMR_KBD_OASYS_TYPE                uint32 = 0x1
	FMV_KBD_OASYS_TYPE                uint32 = 0x2
	NEC_KBD_NORMAL_TYPE               uint32 = 0x1
	NEC_KBD_N_MODE_TYPE               uint32 = 0x2
	NEC_KBD_H_MODE_TYPE               uint32 = 0x3
	NEC_KBD_LAPTOP_TYPE               uint32 = 0x4
	NEC_KBD_106_TYPE                  uint32 = 0x5
	TOSHIBA_KBD_DESKTOP_TYPE          uint32 = 0xd
	TOSHIBA_KBD_LAPTOP_TYPE           uint32 = 0xf
	DEC_KBD_ANSI_LAYOUT_TYPE          uint32 = 0x1
	DEC_KBD_JIS_LAYOUT_TYPE           uint32 = 0x2
	MICROSOFT_KBD_101A_TYPE           uint32 = 0x0
	MICROSOFT_KBD_101B_TYPE           uint32 = 0x4
	MICROSOFT_KBD_101C_TYPE           uint32 = 0x5
	MICROSOFT_KBD_103_TYPE            uint32 = 0x6
	NLSKBD_INFO_SEND_IME_NOTIFICATION uint32 = 0x1
	NLSKBD_INFO_ACCESSIBILITY_KEYMAP  uint32 = 0x2
	NLSKBD_INFO_EMURATE_101_KEYBOARD  uint32 = 0x10
	NLSKBD_INFO_EMURATE_106_KEYBOARD  uint32 = 0x20
	KBDNLS_TYPE_NULL                  uint32 = 0x0
	KBDNLS_TYPE_NORMAL                uint32 = 0x1
	KBDNLS_TYPE_TOGGLE                uint32 = 0x2
	KBDNLS_INDEX_NORMAL               uint32 = 0x1
	KBDNLS_INDEX_ALT                  uint32 = 0x2
	KBDNLS_NULL                       uint32 = 0x0
	KBDNLS_NOEVENT                    uint32 = 0x1
	KBDNLS_SEND_BASE_VK               uint32 = 0x2
	KBDNLS_SEND_PARAM_VK              uint32 = 0x3
	KBDNLS_KANALOCK                   uint32 = 0x4
	KBDNLS_ALPHANUM                   uint32 = 0x5
	KBDNLS_HIRAGANA                   uint32 = 0x6
	KBDNLS_KATAKANA                   uint32 = 0x7
	KBDNLS_SBCSDBCS                   uint32 = 0x8
	KBDNLS_ROMAN                      uint32 = 0x9
	KBDNLS_CODEINPUT                  uint32 = 0xa
	KBDNLS_HELP_OR_END                uint32 = 0xb
	KBDNLS_HOME_OR_CLEAR              uint32 = 0xc
	KBDNLS_NUMPAD                     uint32 = 0xd
	KBDNLS_KANAEVENT                  uint32 = 0xe
	KBDNLS_CONV_OR_NONCONV            uint32 = 0xf
	KBD_TYPE                          uint32 = 0x4
	VK__none_                         uint32 = 0xff
	VK_ABNT_C1                        uint32 = 0xc1
	VK_ABNT_C2                        uint32 = 0xc2
	SCANCODE_LSHIFT                   uint32 = 0x2a
	SCANCODE_RSHIFT                   uint32 = 0x36
	SCANCODE_CTRL                     uint32 = 0x1d
	SCANCODE_ALT                      uint32 = 0x38
	SCANCODE_NUMPAD_FIRST             uint32 = 0x47
	SCANCODE_NUMPAD_LAST              uint32 = 0x52
	SCANCODE_LWIN                     uint32 = 0x5b
	SCANCODE_RWIN                     uint32 = 0x5c
	SCANCODE_THAI_LAYOUT_TOGGLE       uint32 = 0x29
	VK_DBE_ALPHANUMERIC               uint32 = 0xf0
	VK_DBE_KATAKANA                   uint32 = 0xf1
	VK_DBE_HIRAGANA                   uint32 = 0xf2
	VK_DBE_SBCSCHAR                   uint32 = 0xf3
	VK_DBE_DBCSCHAR                   uint32 = 0xf4
	VK_DBE_ROMAN                      uint32 = 0xf5
	VK_DBE_NOROMAN                    uint32 = 0xf6
	VK_DBE_ENTERWORDREGISTERMODE      uint32 = 0xf7
	VK_DBE_ENTERIMECONFIGMODE         uint32 = 0xf8
	VK_DBE_FLUSHSTRING                uint32 = 0xf9
	VK_DBE_CODEINPUT                  uint32 = 0xfa
	VK_DBE_NOCODEINPUT                uint32 = 0xfb
	VK_DBE_DETERMINESTRING            uint32 = 0xfc
	VK_DBE_ENTERDLGCONVERSIONMODE     uint32 = 0xfd
)

// enums

// enum
// flags
type HOT_KEY_MODIFIERS uint32

const (
	MOD_ALT      HOT_KEY_MODIFIERS = 1
	MOD_CONTROL  HOT_KEY_MODIFIERS = 2
	MOD_NOREPEAT HOT_KEY_MODIFIERS = 16384
	MOD_SHIFT    HOT_KEY_MODIFIERS = 4
	MOD_WIN      HOT_KEY_MODIFIERS = 8
)

// enum
type ACTIVATE_KEYBOARD_LAYOUT_FLAGS uint32

const (
	KLF_REORDER       ACTIVATE_KEYBOARD_LAYOUT_FLAGS = 8
	KLF_RESET         ACTIVATE_KEYBOARD_LAYOUT_FLAGS = 1073741824
	KLF_SETFORPROCESS ACTIVATE_KEYBOARD_LAYOUT_FLAGS = 256
	KLF_SHIFTLOCK     ACTIVATE_KEYBOARD_LAYOUT_FLAGS = 65536
	KLF_ACTIVATE      ACTIVATE_KEYBOARD_LAYOUT_FLAGS = 1
	KLF_NOTELLSHELL   ACTIVATE_KEYBOARD_LAYOUT_FLAGS = 128
	KLF_REPLACELANG   ACTIVATE_KEYBOARD_LAYOUT_FLAGS = 16
	KLF_SUBSTITUTE_OK ACTIVATE_KEYBOARD_LAYOUT_FLAGS = 2
)

// enum
type GET_MOUSE_MOVE_POINTS_EX_RESOLUTION uint32

const (
	GMMP_USE_DISPLAY_POINTS         GET_MOUSE_MOVE_POINTS_EX_RESOLUTION = 1
	GMMP_USE_HIGH_RESOLUTION_POINTS GET_MOUSE_MOVE_POINTS_EX_RESOLUTION = 2
)

// enum
// flags
type KEYBD_EVENT_FLAGS uint32

const (
	KEYEVENTF_EXTENDEDKEY KEYBD_EVENT_FLAGS = 1
	KEYEVENTF_KEYUP       KEYBD_EVENT_FLAGS = 2
	KEYEVENTF_SCANCODE    KEYBD_EVENT_FLAGS = 8
	KEYEVENTF_UNICODE     KEYBD_EVENT_FLAGS = 4
)

// enum
// flags
type MOUSE_EVENT_FLAGS uint32

const (
	MOUSEEVENTF_ABSOLUTE        MOUSE_EVENT_FLAGS = 32768
	MOUSEEVENTF_LEFTDOWN        MOUSE_EVENT_FLAGS = 2
	MOUSEEVENTF_LEFTUP          MOUSE_EVENT_FLAGS = 4
	MOUSEEVENTF_MIDDLEDOWN      MOUSE_EVENT_FLAGS = 32
	MOUSEEVENTF_MIDDLEUP        MOUSE_EVENT_FLAGS = 64
	MOUSEEVENTF_MOVE            MOUSE_EVENT_FLAGS = 1
	MOUSEEVENTF_RIGHTDOWN       MOUSE_EVENT_FLAGS = 8
	MOUSEEVENTF_RIGHTUP         MOUSE_EVENT_FLAGS = 16
	MOUSEEVENTF_WHEEL           MOUSE_EVENT_FLAGS = 2048
	MOUSEEVENTF_XDOWN           MOUSE_EVENT_FLAGS = 128
	MOUSEEVENTF_XUP             MOUSE_EVENT_FLAGS = 256
	MOUSEEVENTF_HWHEEL          MOUSE_EVENT_FLAGS = 4096
	MOUSEEVENTF_MOVE_NOCOALESCE MOUSE_EVENT_FLAGS = 8192
	MOUSEEVENTF_VIRTUALDESK     MOUSE_EVENT_FLAGS = 16384
)

// enum
type INPUT_TYPE uint32

const (
	INPUT_MOUSE    INPUT_TYPE = 0
	INPUT_KEYBOARD INPUT_TYPE = 1
	INPUT_HARDWARE INPUT_TYPE = 2
)

// enum
// flags
type TRACKMOUSEEVENT_FLAGS uint32

const (
	TME_CANCEL    TRACKMOUSEEVENT_FLAGS = 2147483648
	TME_HOVER     TRACKMOUSEEVENT_FLAGS = 1
	TME_LEAVE     TRACKMOUSEEVENT_FLAGS = 2
	TME_NONCLIENT TRACKMOUSEEVENT_FLAGS = 16
	TME_QUERY     TRACKMOUSEEVENT_FLAGS = 1073741824
)

// enum
type VIRTUAL_KEY uint16

const (
	VK_0                               VIRTUAL_KEY = 48
	VK_1                               VIRTUAL_KEY = 49
	VK_2                               VIRTUAL_KEY = 50
	VK_3                               VIRTUAL_KEY = 51
	VK_4                               VIRTUAL_KEY = 52
	VK_5                               VIRTUAL_KEY = 53
	VK_6                               VIRTUAL_KEY = 54
	VK_7                               VIRTUAL_KEY = 55
	VK_8                               VIRTUAL_KEY = 56
	VK_9                               VIRTUAL_KEY = 57
	VK_A                               VIRTUAL_KEY = 65
	VK_B                               VIRTUAL_KEY = 66
	VK_C                               VIRTUAL_KEY = 67
	VK_D                               VIRTUAL_KEY = 68
	VK_E                               VIRTUAL_KEY = 69
	VK_F                               VIRTUAL_KEY = 70
	VK_G                               VIRTUAL_KEY = 71
	VK_H                               VIRTUAL_KEY = 72
	VK_I                               VIRTUAL_KEY = 73
	VK_J                               VIRTUAL_KEY = 74
	VK_K                               VIRTUAL_KEY = 75
	VK_L                               VIRTUAL_KEY = 76
	VK_M                               VIRTUAL_KEY = 77
	VK_N                               VIRTUAL_KEY = 78
	VK_O                               VIRTUAL_KEY = 79
	VK_P                               VIRTUAL_KEY = 80
	VK_Q                               VIRTUAL_KEY = 81
	VK_R                               VIRTUAL_KEY = 82
	VK_S                               VIRTUAL_KEY = 83
	VK_T                               VIRTUAL_KEY = 84
	VK_U                               VIRTUAL_KEY = 85
	VK_V                               VIRTUAL_KEY = 86
	VK_W                               VIRTUAL_KEY = 87
	VK_X                               VIRTUAL_KEY = 88
	VK_Y                               VIRTUAL_KEY = 89
	VK_Z                               VIRTUAL_KEY = 90
	VK_LBUTTON                         VIRTUAL_KEY = 1
	VK_RBUTTON                         VIRTUAL_KEY = 2
	VK_CANCEL                          VIRTUAL_KEY = 3
	VK_MBUTTON                         VIRTUAL_KEY = 4
	VK_XBUTTON1                        VIRTUAL_KEY = 5
	VK_XBUTTON2                        VIRTUAL_KEY = 6
	VK_BACK                            VIRTUAL_KEY = 8
	VK_TAB                             VIRTUAL_KEY = 9
	VK_CLEAR                           VIRTUAL_KEY = 12
	VK_RETURN                          VIRTUAL_KEY = 13
	VK_SHIFT                           VIRTUAL_KEY = 16
	VK_CONTROL                         VIRTUAL_KEY = 17
	VK_MENU                            VIRTUAL_KEY = 18
	VK_PAUSE                           VIRTUAL_KEY = 19
	VK_CAPITAL                         VIRTUAL_KEY = 20
	VK_KANA                            VIRTUAL_KEY = 21
	VK_HANGEUL                         VIRTUAL_KEY = 21
	VK_HANGUL                          VIRTUAL_KEY = 21
	VK_IME_ON                          VIRTUAL_KEY = 22
	VK_JUNJA                           VIRTUAL_KEY = 23
	VK_FINAL                           VIRTUAL_KEY = 24
	VK_HANJA                           VIRTUAL_KEY = 25
	VK_KANJI                           VIRTUAL_KEY = 25
	VK_IME_OFF                         VIRTUAL_KEY = 26
	VK_ESCAPE                          VIRTUAL_KEY = 27
	VK_CONVERT                         VIRTUAL_KEY = 28
	VK_NONCONVERT                      VIRTUAL_KEY = 29
	VK_ACCEPT                          VIRTUAL_KEY = 30
	VK_MODECHANGE                      VIRTUAL_KEY = 31
	VK_SPACE                           VIRTUAL_KEY = 32
	VK_PRIOR                           VIRTUAL_KEY = 33
	VK_NEXT                            VIRTUAL_KEY = 34
	VK_END                             VIRTUAL_KEY = 35
	VK_HOME                            VIRTUAL_KEY = 36
	VK_LEFT                            VIRTUAL_KEY = 37
	VK_UP                              VIRTUAL_KEY = 38
	VK_RIGHT                           VIRTUAL_KEY = 39
	VK_DOWN                            VIRTUAL_KEY = 40
	VK_SELECT                          VIRTUAL_KEY = 41
	VK_PRINT                           VIRTUAL_KEY = 42
	VK_EXECUTE                         VIRTUAL_KEY = 43
	VK_SNAPSHOT                        VIRTUAL_KEY = 44
	VK_INSERT                          VIRTUAL_KEY = 45
	VK_DELETE                          VIRTUAL_KEY = 46
	VK_HELP                            VIRTUAL_KEY = 47
	VK_LWIN                            VIRTUAL_KEY = 91
	VK_RWIN                            VIRTUAL_KEY = 92
	VK_APPS                            VIRTUAL_KEY = 93
	VK_SLEEP                           VIRTUAL_KEY = 95
	VK_NUMPAD0                         VIRTUAL_KEY = 96
	VK_NUMPAD1                         VIRTUAL_KEY = 97
	VK_NUMPAD2                         VIRTUAL_KEY = 98
	VK_NUMPAD3                         VIRTUAL_KEY = 99
	VK_NUMPAD4                         VIRTUAL_KEY = 100
	VK_NUMPAD5                         VIRTUAL_KEY = 101
	VK_NUMPAD6                         VIRTUAL_KEY = 102
	VK_NUMPAD7                         VIRTUAL_KEY = 103
	VK_NUMPAD8                         VIRTUAL_KEY = 104
	VK_NUMPAD9                         VIRTUAL_KEY = 105
	VK_MULTIPLY                        VIRTUAL_KEY = 106
	VK_ADD                             VIRTUAL_KEY = 107
	VK_SEPARATOR                       VIRTUAL_KEY = 108
	VK_SUBTRACT                        VIRTUAL_KEY = 109
	VK_DECIMAL                         VIRTUAL_KEY = 110
	VK_DIVIDE                          VIRTUAL_KEY = 111
	VK_F1                              VIRTUAL_KEY = 112
	VK_F2                              VIRTUAL_KEY = 113
	VK_F3                              VIRTUAL_KEY = 114
	VK_F4                              VIRTUAL_KEY = 115
	VK_F5                              VIRTUAL_KEY = 116
	VK_F6                              VIRTUAL_KEY = 117
	VK_F7                              VIRTUAL_KEY = 118
	VK_F8                              VIRTUAL_KEY = 119
	VK_F9                              VIRTUAL_KEY = 120
	VK_F10                             VIRTUAL_KEY = 121
	VK_F11                             VIRTUAL_KEY = 122
	VK_F12                             VIRTUAL_KEY = 123
	VK_F13                             VIRTUAL_KEY = 124
	VK_F14                             VIRTUAL_KEY = 125
	VK_F15                             VIRTUAL_KEY = 126
	VK_F16                             VIRTUAL_KEY = 127
	VK_F17                             VIRTUAL_KEY = 128
	VK_F18                             VIRTUAL_KEY = 129
	VK_F19                             VIRTUAL_KEY = 130
	VK_F20                             VIRTUAL_KEY = 131
	VK_F21                             VIRTUAL_KEY = 132
	VK_F22                             VIRTUAL_KEY = 133
	VK_F23                             VIRTUAL_KEY = 134
	VK_F24                             VIRTUAL_KEY = 135
	VK_NAVIGATION_VIEW                 VIRTUAL_KEY = 136
	VK_NAVIGATION_MENU                 VIRTUAL_KEY = 137
	VK_NAVIGATION_UP                   VIRTUAL_KEY = 138
	VK_NAVIGATION_DOWN                 VIRTUAL_KEY = 139
	VK_NAVIGATION_LEFT                 VIRTUAL_KEY = 140
	VK_NAVIGATION_RIGHT                VIRTUAL_KEY = 141
	VK_NAVIGATION_ACCEPT               VIRTUAL_KEY = 142
	VK_NAVIGATION_CANCEL               VIRTUAL_KEY = 143
	VK_NUMLOCK                         VIRTUAL_KEY = 144
	VK_SCROLL                          VIRTUAL_KEY = 145
	VK_OEM_NEC_EQUAL                   VIRTUAL_KEY = 146
	VK_OEM_FJ_JISHO                    VIRTUAL_KEY = 146
	VK_OEM_FJ_MASSHOU                  VIRTUAL_KEY = 147
	VK_OEM_FJ_TOUROKU                  VIRTUAL_KEY = 148
	VK_OEM_FJ_LOYA                     VIRTUAL_KEY = 149
	VK_OEM_FJ_ROYA                     VIRTUAL_KEY = 150
	VK_LSHIFT                          VIRTUAL_KEY = 160
	VK_RSHIFT                          VIRTUAL_KEY = 161
	VK_LCONTROL                        VIRTUAL_KEY = 162
	VK_RCONTROL                        VIRTUAL_KEY = 163
	VK_LMENU                           VIRTUAL_KEY = 164
	VK_RMENU                           VIRTUAL_KEY = 165
	VK_BROWSER_BACK                    VIRTUAL_KEY = 166
	VK_BROWSER_FORWARD                 VIRTUAL_KEY = 167
	VK_BROWSER_REFRESH                 VIRTUAL_KEY = 168
	VK_BROWSER_STOP                    VIRTUAL_KEY = 169
	VK_BROWSER_SEARCH                  VIRTUAL_KEY = 170
	VK_BROWSER_FAVORITES               VIRTUAL_KEY = 171
	VK_BROWSER_HOME                    VIRTUAL_KEY = 172
	VK_VOLUME_MUTE                     VIRTUAL_KEY = 173
	VK_VOLUME_DOWN                     VIRTUAL_KEY = 174
	VK_VOLUME_UP                       VIRTUAL_KEY = 175
	VK_MEDIA_NEXT_TRACK                VIRTUAL_KEY = 176
	VK_MEDIA_PREV_TRACK                VIRTUAL_KEY = 177
	VK_MEDIA_STOP                      VIRTUAL_KEY = 178
	VK_MEDIA_PLAY_PAUSE                VIRTUAL_KEY = 179
	VK_LAUNCH_MAIL                     VIRTUAL_KEY = 180
	VK_LAUNCH_MEDIA_SELECT             VIRTUAL_KEY = 181
	VK_LAUNCH_APP1                     VIRTUAL_KEY = 182
	VK_LAUNCH_APP2                     VIRTUAL_KEY = 183
	VK_OEM_1                           VIRTUAL_KEY = 186
	VK_OEM_PLUS                        VIRTUAL_KEY = 187
	VK_OEM_COMMA                       VIRTUAL_KEY = 188
	VK_OEM_MINUS                       VIRTUAL_KEY = 189
	VK_OEM_PERIOD                      VIRTUAL_KEY = 190
	VK_OEM_2                           VIRTUAL_KEY = 191
	VK_OEM_3                           VIRTUAL_KEY = 192
	VK_GAMEPAD_A                       VIRTUAL_KEY = 195
	VK_GAMEPAD_B                       VIRTUAL_KEY = 196
	VK_GAMEPAD_X                       VIRTUAL_KEY = 197
	VK_GAMEPAD_Y                       VIRTUAL_KEY = 198
	VK_GAMEPAD_RIGHT_SHOULDER          VIRTUAL_KEY = 199
	VK_GAMEPAD_LEFT_SHOULDER           VIRTUAL_KEY = 200
	VK_GAMEPAD_LEFT_TRIGGER            VIRTUAL_KEY = 201
	VK_GAMEPAD_RIGHT_TRIGGER           VIRTUAL_KEY = 202
	VK_GAMEPAD_DPAD_UP                 VIRTUAL_KEY = 203
	VK_GAMEPAD_DPAD_DOWN               VIRTUAL_KEY = 204
	VK_GAMEPAD_DPAD_LEFT               VIRTUAL_KEY = 205
	VK_GAMEPAD_DPAD_RIGHT              VIRTUAL_KEY = 206
	VK_GAMEPAD_MENU                    VIRTUAL_KEY = 207
	VK_GAMEPAD_VIEW                    VIRTUAL_KEY = 208
	VK_GAMEPAD_LEFT_THUMBSTICK_BUTTON  VIRTUAL_KEY = 209
	VK_GAMEPAD_RIGHT_THUMBSTICK_BUTTON VIRTUAL_KEY = 210
	VK_GAMEPAD_LEFT_THUMBSTICK_UP      VIRTUAL_KEY = 211
	VK_GAMEPAD_LEFT_THUMBSTICK_DOWN    VIRTUAL_KEY = 212
	VK_GAMEPAD_LEFT_THUMBSTICK_RIGHT   VIRTUAL_KEY = 213
	VK_GAMEPAD_LEFT_THUMBSTICK_LEFT    VIRTUAL_KEY = 214
	VK_GAMEPAD_RIGHT_THUMBSTICK_UP     VIRTUAL_KEY = 215
	VK_GAMEPAD_RIGHT_THUMBSTICK_DOWN   VIRTUAL_KEY = 216
	VK_GAMEPAD_RIGHT_THUMBSTICK_RIGHT  VIRTUAL_KEY = 217
	VK_GAMEPAD_RIGHT_THUMBSTICK_LEFT   VIRTUAL_KEY = 218
	VK_OEM_4                           VIRTUAL_KEY = 219
	VK_OEM_5                           VIRTUAL_KEY = 220
	VK_OEM_6                           VIRTUAL_KEY = 221
	VK_OEM_7                           VIRTUAL_KEY = 222
	VK_OEM_8                           VIRTUAL_KEY = 223
	VK_OEM_AX                          VIRTUAL_KEY = 225
	VK_OEM_102                         VIRTUAL_KEY = 226
	VK_ICO_HELP                        VIRTUAL_KEY = 227
	VK_ICO_00                          VIRTUAL_KEY = 228
	VK_PROCESSKEY                      VIRTUAL_KEY = 229
	VK_ICO_CLEAR                       VIRTUAL_KEY = 230
	VK_PACKET                          VIRTUAL_KEY = 231
	VK_OEM_RESET                       VIRTUAL_KEY = 233
	VK_OEM_JUMP                        VIRTUAL_KEY = 234
	VK_OEM_PA1                         VIRTUAL_KEY = 235
	VK_OEM_PA2                         VIRTUAL_KEY = 236
	VK_OEM_PA3                         VIRTUAL_KEY = 237
	VK_OEM_WSCTRL                      VIRTUAL_KEY = 238
	VK_OEM_CUSEL                       VIRTUAL_KEY = 239
	VK_OEM_ATTN                        VIRTUAL_KEY = 240
	VK_OEM_FINISH                      VIRTUAL_KEY = 241
	VK_OEM_COPY                        VIRTUAL_KEY = 242
	VK_OEM_AUTO                        VIRTUAL_KEY = 243
	VK_OEM_ENLW                        VIRTUAL_KEY = 244
	VK_OEM_BACKTAB                     VIRTUAL_KEY = 245
	VK_ATTN                            VIRTUAL_KEY = 246
	VK_CRSEL                           VIRTUAL_KEY = 247
	VK_EXSEL                           VIRTUAL_KEY = 248
	VK_EREOF                           VIRTUAL_KEY = 249
	VK_PLAY                            VIRTUAL_KEY = 250
	VK_ZOOM                            VIRTUAL_KEY = 251
	VK_NONAME                          VIRTUAL_KEY = 252
	VK_PA1                             VIRTUAL_KEY = 253
	VK_OEM_CLEAR                       VIRTUAL_KEY = 254
)

// structs

type VK_TO_BIT struct {
	Vk      byte
	ModBits byte
}

type MODIFIERS struct {
	PVkToBit    *VK_TO_BIT
	WMaxModBits uint16
	ModNumber   [1]byte
}

type VSC_VK struct {
	Vsc byte
	Vk  uint16
}

type VK_VSC struct {
	Vk  byte
	Vsc byte
}

type VK_TO_WCHARS1 struct {
	VirtualKey byte
	Attributes byte
	Wch        [1]uint16
}

type VK_TO_WCHARS2 struct {
	VirtualKey byte
	Attributes byte
	Wch        [2]uint16
}

type VK_TO_WCHARS3 struct {
	VirtualKey byte
	Attributes byte
	Wch        [3]uint16
}

type VK_TO_WCHARS4 struct {
	VirtualKey byte
	Attributes byte
	Wch        [4]uint16
}

type VK_TO_WCHARS5 struct {
	VirtualKey byte
	Attributes byte
	Wch        [5]uint16
}

type VK_TO_WCHARS6 struct {
	VirtualKey byte
	Attributes byte
	Wch        [6]uint16
}

type VK_TO_WCHARS7 struct {
	VirtualKey byte
	Attributes byte
	Wch        [7]uint16
}

type VK_TO_WCHARS8 struct {
	VirtualKey byte
	Attributes byte
	Wch        [8]uint16
}

type VK_TO_WCHARS9 struct {
	VirtualKey byte
	Attributes byte
	Wch        [9]uint16
}

type VK_TO_WCHARS10 struct {
	VirtualKey byte
	Attributes byte
	Wch        [10]uint16
}

type VK_TO_WCHAR_TABLE struct {
	PVkToWchars    *VK_TO_WCHARS1
	NModifications byte
	CbSize         byte
}

type DEADKEY struct {
	DwBoth      uint32
	WchComposed uint16
	UFlags      uint16
}

type LIGATURE1 struct {
	VirtualKey         byte
	ModificationNumber uint16
	Wch                [1]uint16
}

type LIGATURE2 struct {
	VirtualKey         byte
	ModificationNumber uint16
	Wch                [2]uint16
}

type LIGATURE3 struct {
	VirtualKey         byte
	ModificationNumber uint16
	Wch                [3]uint16
}

type LIGATURE4 struct {
	VirtualKey         byte
	ModificationNumber uint16
	Wch                [4]uint16
}

type LIGATURE5 struct {
	VirtualKey         byte
	ModificationNumber uint16
	Wch                [5]uint16
}

type VSC_LPWSTR struct {
	Vsc  byte
	Pwsz PWSTR
}

type TagKbdLayer struct {
	PCharModifiers  *MODIFIERS
	PVkToWcharTable *VK_TO_WCHAR_TABLE
	PDeadKey        *DEADKEY
	PKeyNames       *VSC_LPWSTR
	PKeyNamesExt    *VSC_LPWSTR
	PKeyNamesDead   **uint16
	PusVSCtoVK      *uint16
	BMaxVSCtoVK     byte
	PVSCtoVK_E0     *VSC_VK
	PVSCtoVK_E1     *VSC_VK
	FLocaleFlags    uint32
	NLgMax          byte
	CbLgEntry       byte
	PLigature       *LIGATURE1
	DwType          uint32
	DwSubType       uint32
}

type VK_FUNCTION_PARAM_ struct {
	NLSFEProcIndex byte
	NLSFEProcParam uint32
}

type VK_TO_FUNCTION_TABLE_ struct {
	Vk               byte
	NLSFEProcType    byte
	NLSFEProcCurrent byte
	NLSFEProcSwitch  byte
	NLSFEProc        [8]VK_FUNCTION_PARAM_
	NLSFEProcAlt     [8]VK_FUNCTION_PARAM_
}

type TagKbdNlsLayer struct {
	OEMIdentifier     uint16
	LayoutInformation uint16
	NumOfVkToF        uint32
	PVkToF            *VK_TO_FUNCTION_TABLE_
	NumOfMouseVKey    int32
	PusMouseVKey      *uint16
}

type KBDTABLE_DESC struct {
	WszDllName [32]uint16
	DwType     uint32
	DwSubType  uint32
}

type KBDTABLE_MULTI struct {
	NTables    uint32
	AKbdTables [8]KBDTABLE_DESC
}

type KBD_TYPE_INFO struct {
	DwVersion uint32
	DwType    uint32
	DwSubType uint32
}

type MOUSEMOVEPOINT struct {
	X           int32
	Y           int32
	Time        uint32
	DwExtraInfo uintptr
}

type TRACKMOUSEEVENT struct {
	CbSize      uint32
	DwFlags     TRACKMOUSEEVENT_FLAGS
	HwndTrack   HWND
	DwHoverTime uint32
}

type MOUSEINPUT struct {
	Dx          int32
	Dy          int32
	MouseData   int32
	DwFlags     MOUSE_EVENT_FLAGS
	Time        uint32
	DwExtraInfo uintptr
}

type KEYBDINPUT struct {
	WVk         VIRTUAL_KEY
	WScan       uint16
	DwFlags     KEYBD_EVENT_FLAGS
	Time        uint32
	DwExtraInfo uintptr
}

type HARDWAREINPUT struct {
	UMsg    uint32
	WParamL uint16
	WParamH uint16
}

type INPUT_Anonymous struct {
	Data [4]uint64
}

func (this *INPUT_Anonymous) Mi() *MOUSEINPUT {
	return (*MOUSEINPUT)(unsafe.Pointer(this))
}

func (this *INPUT_Anonymous) MiVal() MOUSEINPUT {
	return *(*MOUSEINPUT)(unsafe.Pointer(this))
}

func (this *INPUT_Anonymous) Ki() *KEYBDINPUT {
	return (*KEYBDINPUT)(unsafe.Pointer(this))
}

func (this *INPUT_Anonymous) KiVal() KEYBDINPUT {
	return *(*KEYBDINPUT)(unsafe.Pointer(this))
}

func (this *INPUT_Anonymous) Hi() *HARDWAREINPUT {
	return (*HARDWAREINPUT)(unsafe.Pointer(this))
}

func (this *INPUT_Anonymous) HiVal() HARDWAREINPUT {
	return *(*HARDWAREINPUT)(unsafe.Pointer(this))
}

type INPUT struct {
	Type_ INPUT_TYPE
	INPUT_Anonymous
}

type LASTINPUTINFO struct {
	CbSize uint32
	DwTime uint32
}

var (
	pTrackMouseEvent_       uintptr
	pLoadKeyboardLayoutA    uintptr
	pLoadKeyboardLayoutW    uintptr
	pActivateKeyboardLayout uintptr
	pToUnicodeEx            uintptr
	pUnloadKeyboardLayout   uintptr
	pGetKeyboardLayoutNameA uintptr
	pGetKeyboardLayoutNameW uintptr
	pGetKeyboardLayoutList  uintptr
	pGetKeyboardLayout      uintptr
	pGetMouseMovePointsEx   uintptr
	pTrackMouseEvent        uintptr
	pRegisterHotKey         uintptr
	pUnregisterHotKey       uintptr
	pSwapMouseButton        uintptr
	pGetDoubleClickTime     uintptr
	pSetDoubleClickTime     uintptr
	pSetFocus               uintptr
	pGetActiveWindow        uintptr
	pGetFocus               uintptr
	pGetKBCodePage          uintptr
	pGetKeyState            uintptr
	pGetAsyncKeyState       uintptr
	pGetKeyboardState       uintptr
	pSetKeyboardState       uintptr
	pGetKeyNameTextA        uintptr
	pGetKeyNameTextW        uintptr
	pGetKeyboardType        uintptr
	pToAscii                uintptr
	pToAsciiEx              uintptr
	pToUnicode              uintptr
	pOemKeyScan             uintptr
	pVkKeyScanA             uintptr
	pVkKeyScanW             uintptr
	pVkKeyScanExA           uintptr
	pVkKeyScanExW           uintptr
	pKeybd_event            uintptr
	pMouse_event            uintptr
	pSendInput              uintptr
	pGetLastInputInfo       uintptr
	pMapVirtualKeyA         uintptr
	pMapVirtualKeyW         uintptr
	pMapVirtualKeyExA       uintptr
	pMapVirtualKeyExW       uintptr
	pGetCapture             uintptr
	pSetCapture             uintptr
	pReleaseCapture         uintptr
	pEnableWindow           uintptr
	pIsWindowEnabled        uintptr
	pDragDetect             uintptr
	pSetActiveWindow        uintptr
	pBlockInput             uintptr
)

func TrackMouseEvent_(lpEventTrack *TRACKMOUSEEVENT) BOOL {
	addr := lazyAddr(&pTrackMouseEvent_, libComctl32, "_TrackMouseEvent")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpEventTrack)))
	return BOOL(ret)
}

func LoadKeyboardLayoutA(pwszKLID PSTR, Flags ACTIVATE_KEYBOARD_LAYOUT_FLAGS) (unsafe.Pointer, WIN32_ERROR) {
	addr := lazyAddr(&pLoadKeyboardLayoutA, libUser32, "LoadKeyboardLayoutA")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pwszKLID)), uintptr(Flags))
	return (unsafe.Pointer)(ret), WIN32_ERROR(err)
}

var LoadKeyboardLayout = LoadKeyboardLayoutW

func LoadKeyboardLayoutW(pwszKLID PWSTR, Flags ACTIVATE_KEYBOARD_LAYOUT_FLAGS) (unsafe.Pointer, WIN32_ERROR) {
	addr := lazyAddr(&pLoadKeyboardLayoutW, libUser32, "LoadKeyboardLayoutW")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pwszKLID)), uintptr(Flags))
	return (unsafe.Pointer)(ret), WIN32_ERROR(err)
}

func ActivateKeyboardLayout(hkl unsafe.Pointer, Flags ACTIVATE_KEYBOARD_LAYOUT_FLAGS) (unsafe.Pointer, WIN32_ERROR) {
	addr := lazyAddr(&pActivateKeyboardLayout, libUser32, "ActivateKeyboardLayout")
	ret, _, err := syscall.SyscallN(addr, uintptr(hkl), uintptr(Flags))
	return (unsafe.Pointer)(ret), WIN32_ERROR(err)
}

func ToUnicodeEx(wVirtKey uint32, wScanCode uint32, lpKeyState *byte, pwszBuff PWSTR, cchBuff int32, wFlags uint32, dwhkl unsafe.Pointer) int32 {
	addr := lazyAddr(&pToUnicodeEx, libUser32, "ToUnicodeEx")
	ret, _, _ := syscall.SyscallN(addr, uintptr(wVirtKey), uintptr(wScanCode), uintptr(unsafe.Pointer(lpKeyState)), uintptr(unsafe.Pointer(pwszBuff)), uintptr(cchBuff), uintptr(wFlags), uintptr(dwhkl))
	return int32(ret)
}

func UnloadKeyboardLayout(hkl unsafe.Pointer) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pUnloadKeyboardLayout, libUser32, "UnloadKeyboardLayout")
	ret, _, err := syscall.SyscallN(addr, uintptr(hkl))
	return BOOL(ret), WIN32_ERROR(err)
}

func GetKeyboardLayoutNameA(pwszKLID PSTR) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pGetKeyboardLayoutNameA, libUser32, "GetKeyboardLayoutNameA")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pwszKLID)))
	return BOOL(ret), WIN32_ERROR(err)
}

var GetKeyboardLayoutName = GetKeyboardLayoutNameW

func GetKeyboardLayoutNameW(pwszKLID PWSTR) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pGetKeyboardLayoutNameW, libUser32, "GetKeyboardLayoutNameW")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pwszKLID)))
	return BOOL(ret), WIN32_ERROR(err)
}

func GetKeyboardLayoutList(nBuff int32, lpList unsafe.Pointer) (int32, WIN32_ERROR) {
	addr := lazyAddr(&pGetKeyboardLayoutList, libUser32, "GetKeyboardLayoutList")
	ret, _, err := syscall.SyscallN(addr, uintptr(nBuff), uintptr(lpList))
	return int32(ret), WIN32_ERROR(err)
}

func GetKeyboardLayout(idThread uint32) unsafe.Pointer {
	addr := lazyAddr(&pGetKeyboardLayout, libUser32, "GetKeyboardLayout")
	ret, _, _ := syscall.SyscallN(addr, uintptr(idThread))
	return (unsafe.Pointer)(ret)
}

func GetMouseMovePointsEx(cbSize uint32, lppt *MOUSEMOVEPOINT, lpptBuf *MOUSEMOVEPOINT, nBufPoints int32, resolution GET_MOUSE_MOVE_POINTS_EX_RESOLUTION) (int32, WIN32_ERROR) {
	addr := lazyAddr(&pGetMouseMovePointsEx, libUser32, "GetMouseMovePointsEx")
	ret, _, err := syscall.SyscallN(addr, uintptr(cbSize), uintptr(unsafe.Pointer(lppt)), uintptr(unsafe.Pointer(lpptBuf)), uintptr(nBufPoints), uintptr(resolution))
	return int32(ret), WIN32_ERROR(err)
}

func TrackMouseEvent(lpEventTrack *TRACKMOUSEEVENT) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pTrackMouseEvent, libUser32, "TrackMouseEvent")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpEventTrack)))
	return BOOL(ret), WIN32_ERROR(err)
}

func RegisterHotKey(hWnd HWND, id int32, fsModifiers HOT_KEY_MODIFIERS, vk uint32) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pRegisterHotKey, libUser32, "RegisterHotKey")
	ret, _, err := syscall.SyscallN(addr, hWnd, uintptr(id), uintptr(fsModifiers), uintptr(vk))
	return BOOL(ret), WIN32_ERROR(err)
}

func UnregisterHotKey(hWnd HWND, id int32) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pUnregisterHotKey, libUser32, "UnregisterHotKey")
	ret, _, err := syscall.SyscallN(addr, hWnd, uintptr(id))
	return BOOL(ret), WIN32_ERROR(err)
}

func SwapMouseButton(fSwap BOOL) BOOL {
	addr := lazyAddr(&pSwapMouseButton, libUser32, "SwapMouseButton")
	ret, _, _ := syscall.SyscallN(addr, uintptr(fSwap))
	return BOOL(ret)
}

func GetDoubleClickTime() uint32 {
	addr := lazyAddr(&pGetDoubleClickTime, libUser32, "GetDoubleClickTime")
	ret, _, _ := syscall.SyscallN(addr)
	return uint32(ret)
}

func SetDoubleClickTime(param0 uint32) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pSetDoubleClickTime, libUser32, "SetDoubleClickTime")
	ret, _, err := syscall.SyscallN(addr, uintptr(param0))
	return BOOL(ret), WIN32_ERROR(err)
}

func SetFocus(hWnd HWND) (HWND, WIN32_ERROR) {
	addr := lazyAddr(&pSetFocus, libUser32, "SetFocus")
	ret, _, err := syscall.SyscallN(addr, hWnd)
	return ret, WIN32_ERROR(err)
}

func GetActiveWindow() HWND {
	addr := lazyAddr(&pGetActiveWindow, libUser32, "GetActiveWindow")
	ret, _, _ := syscall.SyscallN(addr)
	return ret
}

func GetFocus() HWND {
	addr := lazyAddr(&pGetFocus, libUser32, "GetFocus")
	ret, _, _ := syscall.SyscallN(addr)
	return ret
}

func GetKBCodePage() uint32 {
	addr := lazyAddr(&pGetKBCodePage, libUser32, "GetKBCodePage")
	ret, _, _ := syscall.SyscallN(addr)
	return uint32(ret)
}

func GetKeyState(nVirtKey int32) int16 {
	addr := lazyAddr(&pGetKeyState, libUser32, "GetKeyState")
	ret, _, _ := syscall.SyscallN(addr, uintptr(nVirtKey))
	return int16(ret)
}

func GetAsyncKeyState(vKey int32) int16 {
	addr := lazyAddr(&pGetAsyncKeyState, libUser32, "GetAsyncKeyState")
	ret, _, _ := syscall.SyscallN(addr, uintptr(vKey))
	return int16(ret)
}

func GetKeyboardState(lpKeyState *byte) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pGetKeyboardState, libUser32, "GetKeyboardState")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpKeyState)))
	return BOOL(ret), WIN32_ERROR(err)
}

func SetKeyboardState(lpKeyState *byte) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pSetKeyboardState, libUser32, "SetKeyboardState")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpKeyState)))
	return BOOL(ret), WIN32_ERROR(err)
}

func GetKeyNameTextA(lParam int32, lpString PSTR, cchSize int32) (int32, WIN32_ERROR) {
	addr := lazyAddr(&pGetKeyNameTextA, libUser32, "GetKeyNameTextA")
	ret, _, err := syscall.SyscallN(addr, uintptr(lParam), uintptr(unsafe.Pointer(lpString)), uintptr(cchSize))
	return int32(ret), WIN32_ERROR(err)
}

var GetKeyNameText = GetKeyNameTextW

func GetKeyNameTextW(lParam int32, lpString PWSTR, cchSize int32) (int32, WIN32_ERROR) {
	addr := lazyAddr(&pGetKeyNameTextW, libUser32, "GetKeyNameTextW")
	ret, _, err := syscall.SyscallN(addr, uintptr(lParam), uintptr(unsafe.Pointer(lpString)), uintptr(cchSize))
	return int32(ret), WIN32_ERROR(err)
}

func GetKeyboardType(nTypeFlag int32) (int32, WIN32_ERROR) {
	addr := lazyAddr(&pGetKeyboardType, libUser32, "GetKeyboardType")
	ret, _, err := syscall.SyscallN(addr, uintptr(nTypeFlag))
	return int32(ret), WIN32_ERROR(err)
}

func ToAscii(uVirtKey uint32, uScanCode uint32, lpKeyState *byte, lpChar *uint16, uFlags uint32) int32 {
	addr := lazyAddr(&pToAscii, libUser32, "ToAscii")
	ret, _, _ := syscall.SyscallN(addr, uintptr(uVirtKey), uintptr(uScanCode), uintptr(unsafe.Pointer(lpKeyState)), uintptr(unsafe.Pointer(lpChar)), uintptr(uFlags))
	return int32(ret)
}

func ToAsciiEx(uVirtKey uint32, uScanCode uint32, lpKeyState *byte, lpChar *uint16, uFlags uint32, dwhkl unsafe.Pointer) int32 {
	addr := lazyAddr(&pToAsciiEx, libUser32, "ToAsciiEx")
	ret, _, _ := syscall.SyscallN(addr, uintptr(uVirtKey), uintptr(uScanCode), uintptr(unsafe.Pointer(lpKeyState)), uintptr(unsafe.Pointer(lpChar)), uintptr(uFlags), uintptr(dwhkl))
	return int32(ret)
}

func ToUnicode(wVirtKey uint32, wScanCode uint32, lpKeyState *byte, pwszBuff PWSTR, cchBuff int32, wFlags uint32) int32 {
	addr := lazyAddr(&pToUnicode, libUser32, "ToUnicode")
	ret, _, _ := syscall.SyscallN(addr, uintptr(wVirtKey), uintptr(wScanCode), uintptr(unsafe.Pointer(lpKeyState)), uintptr(unsafe.Pointer(pwszBuff)), uintptr(cchBuff), uintptr(wFlags))
	return int32(ret)
}

func OemKeyScan(wOemChar uint16) uint32 {
	addr := lazyAddr(&pOemKeyScan, libUser32, "OemKeyScan")
	ret, _, _ := syscall.SyscallN(addr, uintptr(wOemChar))
	return uint32(ret)
}

func VkKeyScanA(ch CHAR) int16 {
	addr := lazyAddr(&pVkKeyScanA, libUser32, "VkKeyScanA")
	ret, _, _ := syscall.SyscallN(addr, uintptr(ch))
	return int16(ret)
}

var VkKeyScan = VkKeyScanW

func VkKeyScanW(ch uint16) int16 {
	addr := lazyAddr(&pVkKeyScanW, libUser32, "VkKeyScanW")
	ret, _, _ := syscall.SyscallN(addr, uintptr(ch))
	return int16(ret)
}

func VkKeyScanExA(ch CHAR, dwhkl unsafe.Pointer) int16 {
	addr := lazyAddr(&pVkKeyScanExA, libUser32, "VkKeyScanExA")
	ret, _, _ := syscall.SyscallN(addr, uintptr(ch), uintptr(dwhkl))
	return int16(ret)
}

var VkKeyScanEx = VkKeyScanExW

func VkKeyScanExW(ch uint16, dwhkl unsafe.Pointer) int16 {
	addr := lazyAddr(&pVkKeyScanExW, libUser32, "VkKeyScanExW")
	ret, _, _ := syscall.SyscallN(addr, uintptr(ch), uintptr(dwhkl))
	return int16(ret)
}

func Keybd_event(bVk byte, bScan byte, dwFlags KEYBD_EVENT_FLAGS, dwExtraInfo uintptr) {
	addr := lazyAddr(&pKeybd_event, libUser32, "keybd_event")
	syscall.SyscallN(addr, uintptr(bVk), uintptr(bScan), uintptr(dwFlags), dwExtraInfo)
}

func Mouse_event(dwFlags MOUSE_EVENT_FLAGS, dx int32, dy int32, dwData uint32, dwExtraInfo uintptr) {
	addr := lazyAddr(&pMouse_event, libUser32, "mouse_event")
	syscall.SyscallN(addr, uintptr(dwFlags), uintptr(dx), uintptr(dy), uintptr(dwData), dwExtraInfo)
}

func SendInput(cInputs uint32, pInputs *INPUT, cbSize int32) (uint32, WIN32_ERROR) {
	addr := lazyAddr(&pSendInput, libUser32, "SendInput")
	ret, _, err := syscall.SyscallN(addr, uintptr(cInputs), uintptr(unsafe.Pointer(pInputs)), uintptr(cbSize))
	return uint32(ret), WIN32_ERROR(err)
}

func GetLastInputInfo(plii *LASTINPUTINFO) BOOL {
	addr := lazyAddr(&pGetLastInputInfo, libUser32, "GetLastInputInfo")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(plii)))
	return BOOL(ret)
}

func MapVirtualKeyA(uCode uint32, uMapType uint32) uint32 {
	addr := lazyAddr(&pMapVirtualKeyA, libUser32, "MapVirtualKeyA")
	ret, _, _ := syscall.SyscallN(addr, uintptr(uCode), uintptr(uMapType))
	return uint32(ret)
}

var MapVirtualKey = MapVirtualKeyW

func MapVirtualKeyW(uCode uint32, uMapType uint32) uint32 {
	addr := lazyAddr(&pMapVirtualKeyW, libUser32, "MapVirtualKeyW")
	ret, _, _ := syscall.SyscallN(addr, uintptr(uCode), uintptr(uMapType))
	return uint32(ret)
}

func MapVirtualKeyExA(uCode uint32, uMapType uint32, dwhkl unsafe.Pointer) uint32 {
	addr := lazyAddr(&pMapVirtualKeyExA, libUser32, "MapVirtualKeyExA")
	ret, _, _ := syscall.SyscallN(addr, uintptr(uCode), uintptr(uMapType), uintptr(dwhkl))
	return uint32(ret)
}

var MapVirtualKeyEx = MapVirtualKeyExW

func MapVirtualKeyExW(uCode uint32, uMapType uint32, dwhkl unsafe.Pointer) uint32 {
	addr := lazyAddr(&pMapVirtualKeyExW, libUser32, "MapVirtualKeyExW")
	ret, _, _ := syscall.SyscallN(addr, uintptr(uCode), uintptr(uMapType), uintptr(dwhkl))
	return uint32(ret)
}

func GetCapture() HWND {
	addr := lazyAddr(&pGetCapture, libUser32, "GetCapture")
	ret, _, _ := syscall.SyscallN(addr)
	return ret
}

func SetCapture(hWnd HWND) HWND {
	addr := lazyAddr(&pSetCapture, libUser32, "SetCapture")
	ret, _, _ := syscall.SyscallN(addr, hWnd)
	return ret
}

func ReleaseCapture() (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pReleaseCapture, libUser32, "ReleaseCapture")
	ret, _, err := syscall.SyscallN(addr)
	return BOOL(ret), WIN32_ERROR(err)
}

func EnableWindow(hWnd HWND, bEnable BOOL) BOOL {
	addr := lazyAddr(&pEnableWindow, libUser32, "EnableWindow")
	ret, _, _ := syscall.SyscallN(addr, hWnd, uintptr(bEnable))
	return BOOL(ret)
}

func IsWindowEnabled(hWnd HWND) BOOL {
	addr := lazyAddr(&pIsWindowEnabled, libUser32, "IsWindowEnabled")
	ret, _, _ := syscall.SyscallN(addr, hWnd)
	return BOOL(ret)
}

func DragDetect(hwnd HWND, pt POINT) BOOL {
	addr := lazyAddr(&pDragDetect, libUser32, "DragDetect")
	ret, _, _ := syscall.SyscallN(addr, hwnd, *(*uintptr)(unsafe.Pointer(&pt)))
	return BOOL(ret)
}

func SetActiveWindow(hWnd HWND) (HWND, WIN32_ERROR) {
	addr := lazyAddr(&pSetActiveWindow, libUser32, "SetActiveWindow")
	ret, _, err := syscall.SyscallN(addr, hWnd)
	return ret, WIN32_ERROR(err)
}

func BlockInput(fBlockIt BOOL) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pBlockInput, libUser32, "BlockInput")
	ret, _, err := syscall.SyscallN(addr, uintptr(fBlockIt))
	return BOOL(ret), WIN32_ERROR(err)
}
