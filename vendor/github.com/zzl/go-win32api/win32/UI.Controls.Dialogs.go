package win32

import (
	"syscall"
	"unsafe"
)

const (
	OFN_SHAREFALLTHROUGH       uint32 = 0x2
	OFN_SHARENOWARN            uint32 = 0x1
	OFN_SHAREWARN              uint32 = 0x0
	CDM_FIRST                  uint32 = 0x464
	CDM_LAST                   uint32 = 0x4c8
	CDM_GETSPEC                uint32 = 0x464
	CDM_GETFILEPATH            uint32 = 0x465
	CDM_GETFOLDERPATH          uint32 = 0x466
	CDM_GETFOLDERIDLIST        uint32 = 0x467
	CDM_SETCONTROLTEXT         uint32 = 0x468
	CDM_HIDECONTROL            uint32 = 0x469
	CDM_SETDEFEXT              uint32 = 0x46a
	FR_RAW                     uint32 = 0x20000
	FR_SHOWWRAPAROUND          uint32 = 0x40000
	FR_NOWRAPAROUND            uint32 = 0x80000
	FR_WRAPAROUND              uint32 = 0x100000
	FRM_FIRST                  uint32 = 0x464
	FRM_LAST                   uint32 = 0x4c8
	FRM_SETOPERATIONRESULT     uint32 = 0x464
	FRM_SETOPERATIONRESULTTEXT uint32 = 0x465
	PS_OPENTYPE_FONTTYPE       uint32 = 0x10000
	TT_OPENTYPE_FONTTYPE       uint32 = 0x20000
	TYPE1_FONTTYPE             uint32 = 0x40000
	SYMBOL_FONTTYPE            uint32 = 0x80000
	WM_CHOOSEFONT_GETLOGFONT   uint32 = 0x401
	WM_CHOOSEFONT_SETLOGFONT   uint32 = 0x465
	WM_CHOOSEFONT_SETFLAGS     uint32 = 0x466
	LBSELCHSTRINGA             string = "commdlg_LBSelChangedNotify"
	SHAREVISTRINGA             string = "commdlg_ShareViolation"
	FILEOKSTRINGA              string = "commdlg_FileNameOK"
	COLOROKSTRINGA             string = "commdlg_ColorOK"
	SETRGBSTRINGA              string = "commdlg_SetRGBColor"
	HELPMSGSTRINGA             string = "commdlg_help"
	FINDMSGSTRINGA             string = "commdlg_FindReplace"
	LBSELCHSTRINGW             string = "commdlg_LBSelChangedNotify"
	SHAREVISTRINGW             string = "commdlg_ShareViolation"
	FILEOKSTRINGW              string = "commdlg_FileNameOK"
	COLOROKSTRINGW             string = "commdlg_ColorOK"
	SETRGBSTRINGW              string = "commdlg_SetRGBColor"
	HELPMSGSTRINGW             string = "commdlg_help"
	FINDMSGSTRINGW             string = "commdlg_FindReplace"
	LBSELCHSTRING              string = "commdlg_LBSelChangedNotify"
	SHAREVISTRING              string = "commdlg_ShareViolation"
	FILEOKSTRING               string = "commdlg_FileNameOK"
	COLOROKSTRING              string = "commdlg_ColorOK"
	SETRGBSTRING               string = "commdlg_SetRGBColor"
	HELPMSGSTRING              string = "commdlg_help"
	FINDMSGSTRING              string = "commdlg_FindReplace"
	CD_LBSELNOITEMS            int32  = -1
	CD_LBSELCHANGE             uint32 = 0x0
	CD_LBSELSUB                uint32 = 0x1
	CD_LBSELADD                uint32 = 0x2
	START_PAGE_GENERAL         uint32 = 0xffffffff
	PD_RESULT_CANCEL           uint32 = 0x0
	PD_RESULT_PRINT            uint32 = 0x1
	PD_RESULT_APPLY            uint32 = 0x2
	DN_DEFAULTPRN              uint32 = 0x1
	WM_PSD_FULLPAGERECT        uint32 = 0x401
	WM_PSD_MINMARGINRECT       uint32 = 0x402
	WM_PSD_MARGINRECT          uint32 = 0x403
	WM_PSD_GREEKTEXTRECT       uint32 = 0x404
	WM_PSD_ENVSTAMPRECT        uint32 = 0x405
	WM_PSD_YAFULLPAGERECT      uint32 = 0x406
	DLG_COLOR                  uint32 = 0xa
	COLOR_HUESCROLL            uint32 = 0x2bc
	COLOR_SATSCROLL            uint32 = 0x2bd
	COLOR_LUMSCROLL            uint32 = 0x2be
	COLOR_HUE                  uint32 = 0x2bf
	COLOR_SAT                  uint32 = 0x2c0
	COLOR_LUM                  uint32 = 0x2c1
	COLOR_RED                  uint32 = 0x2c2
	COLOR_GREEN                uint32 = 0x2c3
	COLOR_BLUE                 uint32 = 0x2c4
	COLOR_CURRENT              uint32 = 0x2c5
	COLOR_RAINBOW              uint32 = 0x2c6
	COLOR_SAVE                 uint32 = 0x2c7
	COLOR_ADD                  uint32 = 0x2c8
	COLOR_SOLID                uint32 = 0x2c9
	COLOR_TUNE                 uint32 = 0x2ca
	COLOR_SCHEMES              uint32 = 0x2cb
	COLOR_ELEMENT              uint32 = 0x2cc
	COLOR_SAMPLES              uint32 = 0x2cd
	COLOR_PALETTE              uint32 = 0x2ce
	COLOR_MIX                  uint32 = 0x2cf
	COLOR_BOX1                 uint32 = 0x2d0
	COLOR_CUSTOM1              uint32 = 0x2d1
	COLOR_HUEACCEL             uint32 = 0x2d3
	COLOR_SATACCEL             uint32 = 0x2d4
	COLOR_LUMACCEL             uint32 = 0x2d5
	COLOR_REDACCEL             uint32 = 0x2d6
	COLOR_GREENACCEL           uint32 = 0x2d7
	COLOR_BLUEACCEL            uint32 = 0x2d8
	COLOR_SOLID_LEFT           uint32 = 0x2da
	COLOR_SOLID_RIGHT          uint32 = 0x2db
	NUM_BASIC_COLORS           uint32 = 0x30
	NUM_CUSTOM_COLORS          uint32 = 0x10
)

// enums

// enum
type COMMON_DLG_ERRORS uint32

const (
	CDERR_DIALOGFAILURE    COMMON_DLG_ERRORS = 65535
	CDERR_GENERALCODES     COMMON_DLG_ERRORS = 0
	CDERR_STRUCTSIZE       COMMON_DLG_ERRORS = 1
	CDERR_INITIALIZATION   COMMON_DLG_ERRORS = 2
	CDERR_NOTEMPLATE       COMMON_DLG_ERRORS = 3
	CDERR_NOHINSTANCE      COMMON_DLG_ERRORS = 4
	CDERR_LOADSTRFAILURE   COMMON_DLG_ERRORS = 5
	CDERR_FINDRESFAILURE   COMMON_DLG_ERRORS = 6
	CDERR_LOADRESFAILURE   COMMON_DLG_ERRORS = 7
	CDERR_LOCKRESFAILURE   COMMON_DLG_ERRORS = 8
	CDERR_MEMALLOCFAILURE  COMMON_DLG_ERRORS = 9
	CDERR_MEMLOCKFAILURE   COMMON_DLG_ERRORS = 10
	CDERR_NOHOOK           COMMON_DLG_ERRORS = 11
	CDERR_REGISTERMSGFAIL  COMMON_DLG_ERRORS = 12
	PDERR_PRINTERCODES     COMMON_DLG_ERRORS = 4096
	PDERR_SETUPFAILURE     COMMON_DLG_ERRORS = 4097
	PDERR_PARSEFAILURE     COMMON_DLG_ERRORS = 4098
	PDERR_RETDEFFAILURE    COMMON_DLG_ERRORS = 4099
	PDERR_LOADDRVFAILURE   COMMON_DLG_ERRORS = 4100
	PDERR_GETDEVMODEFAIL   COMMON_DLG_ERRORS = 4101
	PDERR_INITFAILURE      COMMON_DLG_ERRORS = 4102
	PDERR_NODEVICES        COMMON_DLG_ERRORS = 4103
	PDERR_NODEFAULTPRN     COMMON_DLG_ERRORS = 4104
	PDERR_DNDMMISMATCH     COMMON_DLG_ERRORS = 4105
	PDERR_CREATEICFAILURE  COMMON_DLG_ERRORS = 4106
	PDERR_PRINTERNOTFOUND  COMMON_DLG_ERRORS = 4107
	PDERR_DEFAULTDIFFERENT COMMON_DLG_ERRORS = 4108
	CFERR_CHOOSEFONTCODES  COMMON_DLG_ERRORS = 8192
	CFERR_NOFONTS          COMMON_DLG_ERRORS = 8193
	CFERR_MAXLESSTHANMIN   COMMON_DLG_ERRORS = 8194
	FNERR_FILENAMECODES    COMMON_DLG_ERRORS = 12288
	FNERR_SUBCLASSFAILURE  COMMON_DLG_ERRORS = 12289
	FNERR_INVALIDFILENAME  COMMON_DLG_ERRORS = 12290
	FNERR_BUFFERTOOSMALL   COMMON_DLG_ERRORS = 12291
	FRERR_FINDREPLACECODES COMMON_DLG_ERRORS = 16384
	FRERR_BUFFERLENGTHZERO COMMON_DLG_ERRORS = 16385
	CCERR_CHOOSECOLORCODES COMMON_DLG_ERRORS = 20480
)

// enum
// flags
type OPEN_FILENAME_FLAGS uint32

const (
	OFN_READONLY             OPEN_FILENAME_FLAGS = 1
	OFN_OVERWRITEPROMPT      OPEN_FILENAME_FLAGS = 2
	OFN_HIDEREADONLY         OPEN_FILENAME_FLAGS = 4
	OFN_NOCHANGEDIR          OPEN_FILENAME_FLAGS = 8
	OFN_SHOWHELP             OPEN_FILENAME_FLAGS = 16
	OFN_ENABLEHOOK           OPEN_FILENAME_FLAGS = 32
	OFN_ENABLETEMPLATE       OPEN_FILENAME_FLAGS = 64
	OFN_ENABLETEMPLATEHANDLE OPEN_FILENAME_FLAGS = 128
	OFN_NOVALIDATE           OPEN_FILENAME_FLAGS = 256
	OFN_ALLOWMULTISELECT     OPEN_FILENAME_FLAGS = 512
	OFN_EXTENSIONDIFFERENT   OPEN_FILENAME_FLAGS = 1024
	OFN_PATHMUSTEXIST        OPEN_FILENAME_FLAGS = 2048
	OFN_FILEMUSTEXIST        OPEN_FILENAME_FLAGS = 4096
	OFN_CREATEPROMPT         OPEN_FILENAME_FLAGS = 8192
	OFN_SHAREAWARE           OPEN_FILENAME_FLAGS = 16384
	OFN_NOREADONLYRETURN     OPEN_FILENAME_FLAGS = 32768
	OFN_NOTESTFILECREATE     OPEN_FILENAME_FLAGS = 65536
	OFN_NONETWORKBUTTON      OPEN_FILENAME_FLAGS = 131072
	OFN_NOLONGNAMES          OPEN_FILENAME_FLAGS = 262144
	OFN_EXPLORER             OPEN_FILENAME_FLAGS = 524288
	OFN_NODEREFERENCELINKS   OPEN_FILENAME_FLAGS = 1048576
	OFN_LONGNAMES            OPEN_FILENAME_FLAGS = 2097152
	OFN_ENABLEINCLUDENOTIFY  OPEN_FILENAME_FLAGS = 4194304
	OFN_ENABLESIZING         OPEN_FILENAME_FLAGS = 8388608
	OFN_DONTADDTORECENT      OPEN_FILENAME_FLAGS = 33554432
	OFN_FORCESHOWHIDDEN      OPEN_FILENAME_FLAGS = 268435456
)

// enum
// flags
type OPEN_FILENAME_FLAGS_EX uint32

const (
	OFN_EX_NONE        OPEN_FILENAME_FLAGS_EX = 0
	OFN_EX_NOPLACESBAR OPEN_FILENAME_FLAGS_EX = 1
)

// enum
// flags
type PAGESETUPDLG_FLAGS uint32

const (
	PSD_DEFAULTMINMARGINS             PAGESETUPDLG_FLAGS = 0
	PSD_DISABLEMARGINS                PAGESETUPDLG_FLAGS = 16
	PSD_DISABLEORIENTATION            PAGESETUPDLG_FLAGS = 256
	PSD_DISABLEPAGEPAINTING           PAGESETUPDLG_FLAGS = 524288
	PSD_DISABLEPAPER                  PAGESETUPDLG_FLAGS = 512
	PSD_DISABLEPRINTER                PAGESETUPDLG_FLAGS = 32
	PSD_ENABLEPAGEPAINTHOOK           PAGESETUPDLG_FLAGS = 262144
	PSD_ENABLEPAGESETUPHOOK           PAGESETUPDLG_FLAGS = 8192
	PSD_ENABLEPAGESETUPTEMPLATE       PAGESETUPDLG_FLAGS = 32768
	PSD_ENABLEPAGESETUPTEMPLATEHANDLE PAGESETUPDLG_FLAGS = 131072
	PSD_INHUNDREDTHSOFMILLIMETERS     PAGESETUPDLG_FLAGS = 8
	PSD_INTHOUSANDTHSOFINCHES         PAGESETUPDLG_FLAGS = 4
	PSD_INWININIINTLMEASURE           PAGESETUPDLG_FLAGS = 0
	PSD_MARGINS                       PAGESETUPDLG_FLAGS = 2
	PSD_MINMARGINS                    PAGESETUPDLG_FLAGS = 1
	PSD_NONETWORKBUTTON               PAGESETUPDLG_FLAGS = 2097152
	PSD_NOWARNING                     PAGESETUPDLG_FLAGS = 128
	PSD_RETURNDEFAULT                 PAGESETUPDLG_FLAGS = 1024
	PSD_SHOWHELP                      PAGESETUPDLG_FLAGS = 2048
)

// enum
// flags
type CHOOSEFONT_FLAGS uint32

const (
	CF_APPLY                CHOOSEFONT_FLAGS = 512
	CF_ANSIONLY             CHOOSEFONT_FLAGS = 1024
	CF_BOTH                 CHOOSEFONT_FLAGS = 3
	CF_EFFECTS              CHOOSEFONT_FLAGS = 256
	CF_ENABLEHOOK           CHOOSEFONT_FLAGS = 8
	CF_ENABLETEMPLATE       CHOOSEFONT_FLAGS = 16
	CF_ENABLETEMPLATEHANDLE CHOOSEFONT_FLAGS = 32
	CF_FIXEDPITCHONLY       CHOOSEFONT_FLAGS = 16384
	CF_FORCEFONTEXIST       CHOOSEFONT_FLAGS = 65536
	CF_INACTIVEFONTS        CHOOSEFONT_FLAGS = 33554432
	CF_INITTOLOGFONTSTRUCT  CHOOSEFONT_FLAGS = 64
	CF_LIMITSIZE            CHOOSEFONT_FLAGS = 8192
	CF_NOOEMFONTS           CHOOSEFONT_FLAGS = 2048
	CF_NOFACESEL            CHOOSEFONT_FLAGS = 524288
	CF_NOSCRIPTSEL          CHOOSEFONT_FLAGS = 8388608
	CF_NOSIMULATIONS        CHOOSEFONT_FLAGS = 4096
	CF_NOSIZESEL            CHOOSEFONT_FLAGS = 2097152
	CF_NOSTYLESEL           CHOOSEFONT_FLAGS = 1048576
	CF_NOVECTORFONTS        CHOOSEFONT_FLAGS = 2048
	CF_NOVERTFONTS          CHOOSEFONT_FLAGS = 16777216
	CF_PRINTERFONTS         CHOOSEFONT_FLAGS = 2
	CF_SCALABLEONLY         CHOOSEFONT_FLAGS = 131072
	CF_SCREENFONTS          CHOOSEFONT_FLAGS = 1
	CF_SCRIPTSONLY          CHOOSEFONT_FLAGS = 1024
	CF_SELECTSCRIPT         CHOOSEFONT_FLAGS = 4194304
	CF_SHOWHELP             CHOOSEFONT_FLAGS = 4
	CF_TTONLY               CHOOSEFONT_FLAGS = 262144
	CF_USESTYLE             CHOOSEFONT_FLAGS = 128
	CF_WYSIWYG              CHOOSEFONT_FLAGS = 32768
)

// enum
// flags
type FINDREPLACE_FLAGS uint32

const (
	FR_DIALOGTERM           FINDREPLACE_FLAGS = 64
	FR_DOWN                 FINDREPLACE_FLAGS = 1
	FR_ENABLEHOOK           FINDREPLACE_FLAGS = 256
	FR_ENABLETEMPLATE       FINDREPLACE_FLAGS = 512
	FR_ENABLETEMPLATEHANDLE FINDREPLACE_FLAGS = 8192
	FR_FINDNEXT             FINDREPLACE_FLAGS = 8
	FR_HIDEUPDOWN           FINDREPLACE_FLAGS = 16384
	FR_HIDEMATCHCASE        FINDREPLACE_FLAGS = 32768
	FR_HIDEWHOLEWORD        FINDREPLACE_FLAGS = 65536
	FR_MATCHCASE            FINDREPLACE_FLAGS = 4
	FR_NOMATCHCASE          FINDREPLACE_FLAGS = 2048
	FR_NOUPDOWN             FINDREPLACE_FLAGS = 1024
	FR_NOWHOLEWORD          FINDREPLACE_FLAGS = 4096
	FR_REPLACE              FINDREPLACE_FLAGS = 16
	FR_REPLACEALL           FINDREPLACE_FLAGS = 32
	FR_SHOWHELP             FINDREPLACE_FLAGS = 128
	FR_WHOLEWORD            FINDREPLACE_FLAGS = 2
)

// enum
// flags
type PRINTDLGEX_FLAGS uint32

const (
	PD_ALLPAGES                   PRINTDLGEX_FLAGS = 0
	PD_COLLATE                    PRINTDLGEX_FLAGS = 16
	PD_CURRENTPAGE                PRINTDLGEX_FLAGS = 4194304
	PD_DISABLEPRINTTOFILE         PRINTDLGEX_FLAGS = 524288
	PD_ENABLEPRINTTEMPLATE        PRINTDLGEX_FLAGS = 16384
	PD_ENABLEPRINTTEMPLATEHANDLE  PRINTDLGEX_FLAGS = 65536
	PD_EXCLUSIONFLAGS             PRINTDLGEX_FLAGS = 16777216
	PD_HIDEPRINTTOFILE            PRINTDLGEX_FLAGS = 1048576
	PD_NOCURRENTPAGE              PRINTDLGEX_FLAGS = 8388608
	PD_NOPAGENUMS                 PRINTDLGEX_FLAGS = 8
	PD_NOSELECTION                PRINTDLGEX_FLAGS = 4
	PD_NOWARNING                  PRINTDLGEX_FLAGS = 128
	PD_PAGENUMS                   PRINTDLGEX_FLAGS = 2
	PD_PRINTTOFILE                PRINTDLGEX_FLAGS = 32
	PD_RETURNDC                   PRINTDLGEX_FLAGS = 256
	PD_RETURNDEFAULT              PRINTDLGEX_FLAGS = 1024
	PD_RETURNIC                   PRINTDLGEX_FLAGS = 512
	PD_SELECTION                  PRINTDLGEX_FLAGS = 1
	PD_USEDEVMODECOPIES           PRINTDLGEX_FLAGS = 262144
	PD_USEDEVMODECOPIESANDCOLLATE PRINTDLGEX_FLAGS = 262144
	PD_USELARGETEMPLATE           PRINTDLGEX_FLAGS = 268435456
	PD_ENABLEPRINTHOOK            PRINTDLGEX_FLAGS = 4096
	PD_ENABLESETUPHOOK            PRINTDLGEX_FLAGS = 8192
	PD_ENABLESETUPTEMPLATE        PRINTDLGEX_FLAGS = 32768
	PD_ENABLESETUPTEMPLATEHANDLE  PRINTDLGEX_FLAGS = 131072
	PD_NONETWORKBUTTON            PRINTDLGEX_FLAGS = 2097152
	PD_PRINTSETUP                 PRINTDLGEX_FLAGS = 64
	PD_SHOWHELP                   PRINTDLGEX_FLAGS = 2048
)

// enum
// flags
type CHOOSEFONT_FONT_TYPE uint16

const (
	BOLD_FONTTYPE      CHOOSEFONT_FONT_TYPE = 256
	ITALIC_FONTTYPE    CHOOSEFONT_FONT_TYPE = 512
	PRINTER_FONTTYPE   CHOOSEFONT_FONT_TYPE = 16384
	REGULAR_FONTTYPE   CHOOSEFONT_FONT_TYPE = 1024
	SCREEN_FONTTYPE    CHOOSEFONT_FONT_TYPE = 8192
	SIMULATED_FONTTYPE CHOOSEFONT_FONT_TYPE = 32768
)

// structs

type OPENFILENAME_NT4A struct {
	LStructSize       uint32
	HwndOwner         HWND
	HInstance         HINSTANCE
	LpstrFilter       PSTR
	LpstrCustomFilter PSTR
	NMaxCustFilter    uint32
	NFilterIndex      uint32
	LpstrFile         PSTR
	NMaxFile          uint32
	LpstrFileTitle    PSTR
	NMaxFileTitle     uint32
	LpstrInitialDir   PSTR
	LpstrTitle        PSTR
	Flags             uint32
	NFileOffset       uint16
	NFileExtension    uint16
	LpstrDefExt       PSTR
	LCustData         LPARAM
	LpfnHook          LPOFNHOOKPROC
	LpTemplateName    PSTR
}

type OPENFILENAME_NT4 = OPENFILENAME_NT4W
type OPENFILENAME_NT4W struct {
	LStructSize       uint32
	HwndOwner         HWND
	HInstance         HINSTANCE
	LpstrFilter       PWSTR
	LpstrCustomFilter PWSTR
	NMaxCustFilter    uint32
	NFilterIndex      uint32
	LpstrFile         PWSTR
	NMaxFile          uint32
	LpstrFileTitle    PWSTR
	NMaxFileTitle     uint32
	LpstrInitialDir   PWSTR
	LpstrTitle        PWSTR
	Flags             uint32
	NFileOffset       uint16
	NFileExtension    uint16
	LpstrDefExt       PWSTR
	LCustData         LPARAM
	LpfnHook          LPOFNHOOKPROC
	LpTemplateName    PWSTR
}

type OPENFILENAMEA struct {
	LStructSize       uint32
	HwndOwner         HWND
	HInstance         HINSTANCE
	LpstrFilter       PSTR
	LpstrCustomFilter PSTR
	NMaxCustFilter    uint32
	NFilterIndex      uint32
	LpstrFile         PSTR
	NMaxFile          uint32
	LpstrFileTitle    PSTR
	NMaxFileTitle     uint32
	LpstrInitialDir   PSTR
	LpstrTitle        PSTR
	Flags             OPEN_FILENAME_FLAGS
	NFileOffset       uint16
	NFileExtension    uint16
	LpstrDefExt       PSTR
	LCustData         LPARAM
	LpfnHook          LPOFNHOOKPROC
	LpTemplateName    PSTR
	PvReserved        unsafe.Pointer
	DwReserved        uint32
	FlagsEx           OPEN_FILENAME_FLAGS_EX
}

type OPENFILENAME = OPENFILENAMEW
type OPENFILENAMEW struct {
	LStructSize       uint32
	HwndOwner         HWND
	HInstance         HINSTANCE
	LpstrFilter       PWSTR
	LpstrCustomFilter PWSTR
	NMaxCustFilter    uint32
	NFilterIndex      uint32
	LpstrFile         PWSTR
	NMaxFile          uint32
	LpstrFileTitle    PWSTR
	NMaxFileTitle     uint32
	LpstrInitialDir   PWSTR
	LpstrTitle        PWSTR
	Flags             OPEN_FILENAME_FLAGS
	NFileOffset       uint16
	NFileExtension    uint16
	LpstrDefExt       PWSTR
	LCustData         LPARAM
	LpfnHook          LPOFNHOOKPROC
	LpTemplateName    PWSTR
	PvReserved        unsafe.Pointer
	DwReserved        uint32
	FlagsEx           OPEN_FILENAME_FLAGS_EX
}

type OFNOTIFYA struct {
	Hdr     NMHDR
	LpOFN   *OPENFILENAMEA
	PszFile PSTR
}

type OFNOTIFY = OFNOTIFYW
type OFNOTIFYW struct {
	Hdr     NMHDR
	LpOFN   *OPENFILENAMEW
	PszFile PWSTR
}

type OFNOTIFYEXA struct {
	Hdr   NMHDR
	LpOFN *OPENFILENAMEA
	Psf   unsafe.Pointer
	Pidl  unsafe.Pointer
}

type OFNOTIFYEX = OFNOTIFYEXW
type OFNOTIFYEXW struct {
	Hdr   NMHDR
	LpOFN *OPENFILENAMEW
	Psf   unsafe.Pointer
	Pidl  unsafe.Pointer
}

type CHOOSECOLORA struct {
	LStructSize    uint32
	HwndOwner      HWND
	HInstance      HWND
	RgbResult      uint32
	LpCustColors   *uint32
	Flags          uint32
	LCustData      LPARAM
	LpfnHook       LPCCHOOKPROC
	LpTemplateName PSTR
}

type CHOOSECOLOR = CHOOSECOLORW
type CHOOSECOLORW struct {
	LStructSize    uint32
	HwndOwner      HWND
	HInstance      HWND
	RgbResult      uint32
	LpCustColors   *uint32
	Flags          uint32
	LCustData      LPARAM
	LpfnHook       LPCCHOOKPROC
	LpTemplateName PWSTR
}

type FINDREPLACEA struct {
	LStructSize      uint32
	HwndOwner        HWND
	HInstance        HINSTANCE
	Flags            FINDREPLACE_FLAGS
	LpstrFindWhat    PSTR
	LpstrReplaceWith PSTR
	WFindWhatLen     uint16
	WReplaceWithLen  uint16
	LCustData        LPARAM
	LpfnHook         LPFRHOOKPROC
	LpTemplateName   PSTR
}

type FINDREPLACE = FINDREPLACEW
type FINDREPLACEW struct {
	LStructSize      uint32
	HwndOwner        HWND
	HInstance        HINSTANCE
	Flags            FINDREPLACE_FLAGS
	LpstrFindWhat    PWSTR
	LpstrReplaceWith PWSTR
	WFindWhatLen     uint16
	WReplaceWithLen  uint16
	LCustData        LPARAM
	LpfnHook         LPFRHOOKPROC
	LpTemplateName   PWSTR
}

type CHOOSEFONTA struct {
	LStructSize            uint32
	HwndOwner              HWND
	HDC                    HDC
	LpLogFont              *LOGFONTA
	IPointSize             int32
	Flags                  CHOOSEFONT_FLAGS
	RgbColors              uint32
	LCustData              LPARAM
	LpfnHook               LPCFHOOKPROC
	LpTemplateName         PSTR
	HInstance              HINSTANCE
	LpszStyle              PSTR
	NFontType              CHOOSEFONT_FONT_TYPE
	MISSING_ALIGNMENT_____ uint16
	NSizeMin               int32
	NSizeMax               int32
}

type CHOOSEFONT = CHOOSEFONTW
type CHOOSEFONTW struct {
	LStructSize            uint32
	HwndOwner              HWND
	HDC                    HDC
	LpLogFont              *LOGFONTW
	IPointSize             int32
	Flags                  CHOOSEFONT_FLAGS
	RgbColors              uint32
	LCustData              LPARAM
	LpfnHook               LPCFHOOKPROC
	LpTemplateName         PWSTR
	HInstance              HINSTANCE
	LpszStyle              PWSTR
	NFontType              CHOOSEFONT_FONT_TYPE
	MISSING_ALIGNMENT_____ uint16
	NSizeMin               int32
	NSizeMax               int32
}

type PRINTDLGA struct {
	LStructSize         uint32
	HwndOwner           HWND
	HDevMode            uintptr
	HDevNames           uintptr
	HDC                 HDC
	Flags               PRINTDLGEX_FLAGS
	NFromPage           uint16
	NToPage             uint16
	NMinPage            uint16
	NMaxPage            uint16
	NCopies             uint16
	HInstance           HINSTANCE
	LCustData           LPARAM
	LpfnPrintHook       LPPRINTHOOKPROC
	LpfnSetupHook       LPSETUPHOOKPROC
	LpPrintTemplateName PSTR
	LpSetupTemplateName PSTR
	HPrintTemplate      uintptr
	HSetupTemplate      uintptr
}

type PRINTDLG = PRINTDLGW
type PRINTDLGW struct {
	LStructSize         uint32
	HwndOwner           HWND
	HDevMode            uintptr
	HDevNames           uintptr
	HDC                 HDC
	Flags               PRINTDLGEX_FLAGS
	NFromPage           uint16
	NToPage             uint16
	NMinPage            uint16
	NMaxPage            uint16
	NCopies             uint16
	HInstance           HINSTANCE
	LCustData           LPARAM
	LpfnPrintHook       LPPRINTHOOKPROC
	LpfnSetupHook       LPSETUPHOOKPROC
	LpPrintTemplateName PWSTR
	LpSetupTemplateName PWSTR
	HPrintTemplate      uintptr
	HSetupTemplate      uintptr
}

type PRINTPAGERANGE struct {
	NFromPage uint32
	NToPage   uint32
}

type PRINTDLGEXA struct {
	LStructSize         uint32
	HwndOwner           HWND
	HDevMode            uintptr
	HDevNames           uintptr
	HDC                 HDC
	Flags               PRINTDLGEX_FLAGS
	Flags2              uint32
	ExclusionFlags      uint32
	NPageRanges         uint32
	NMaxPageRanges      uint32
	LpPageRanges        *PRINTPAGERANGE
	NMinPage            uint32
	NMaxPage            uint32
	NCopies             uint32
	HInstance           HINSTANCE
	LpPrintTemplateName PSTR
	LpCallback          *IUnknown
	NPropertyPages      uint32
	LphPropertyPages    *HPROPSHEETPAGE
	NStartPage          uint32
	DwResultAction      uint32
}

type PRINTDLGEX = PRINTDLGEXW
type PRINTDLGEXW struct {
	LStructSize         uint32
	HwndOwner           HWND
	HDevMode            uintptr
	HDevNames           uintptr
	HDC                 HDC
	Flags               PRINTDLGEX_FLAGS
	Flags2              uint32
	ExclusionFlags      uint32
	NPageRanges         uint32
	NMaxPageRanges      uint32
	LpPageRanges        *PRINTPAGERANGE
	NMinPage            uint32
	NMaxPage            uint32
	NCopies             uint32
	HInstance           HINSTANCE
	LpPrintTemplateName PWSTR
	LpCallback          *IUnknown
	NPropertyPages      uint32
	LphPropertyPages    *HPROPSHEETPAGE
	NStartPage          uint32
	DwResultAction      uint32
}

type DEVNAMES struct {
	WDriverOffset uint16
	WDeviceOffset uint16
	WOutputOffset uint16
	WDefault      uint16
}

type PAGESETUPDLGA struct {
	LStructSize             uint32
	HwndOwner               HWND
	HDevMode                uintptr
	HDevNames               uintptr
	Flags                   PAGESETUPDLG_FLAGS
	PtPaperSize             POINT
	RtMinMargin             RECT
	RtMargin                RECT
	HInstance               HINSTANCE
	LCustData               LPARAM
	LpfnPageSetupHook       LPPAGESETUPHOOK
	LpfnPagePaintHook       LPPAGEPAINTHOOK
	LpPageSetupTemplateName PSTR
	HPageSetupTemplate      uintptr
}

type PAGESETUPDLG = PAGESETUPDLGW
type PAGESETUPDLGW struct {
	LStructSize             uint32
	HwndOwner               HWND
	HDevMode                uintptr
	HDevNames               uintptr
	Flags                   PAGESETUPDLG_FLAGS
	PtPaperSize             POINT
	RtMinMargin             RECT
	RtMargin                RECT
	HInstance               HINSTANCE
	LCustData               LPARAM
	LpfnPageSetupHook       LPPAGESETUPHOOK
	LpfnPagePaintHook       LPPAGEPAINTHOOK
	LpPageSetupTemplateName PWSTR
	HPageSetupTemplate      uintptr
}

// func types

type LPOFNHOOKPROC = uintptr
type LPOFNHOOKPROC_func = func(param0 HWND, param1 uint32, param2 WPARAM, param3 LPARAM) uintptr

type LPCCHOOKPROC = uintptr
type LPCCHOOKPROC_func = func(param0 HWND, param1 uint32, param2 WPARAM, param3 LPARAM) uintptr

type LPFRHOOKPROC = uintptr
type LPFRHOOKPROC_func = func(param0 HWND, param1 uint32, param2 WPARAM, param3 LPARAM) uintptr

type LPCFHOOKPROC = uintptr
type LPCFHOOKPROC_func = func(param0 HWND, param1 uint32, param2 WPARAM, param3 LPARAM) uintptr

type LPPRINTHOOKPROC = uintptr
type LPPRINTHOOKPROC_func = func(param0 HWND, param1 uint32, param2 WPARAM, param3 LPARAM) uintptr

type LPSETUPHOOKPROC = uintptr
type LPSETUPHOOKPROC_func = func(param0 HWND, param1 uint32, param2 WPARAM, param3 LPARAM) uintptr

type LPPAGEPAINTHOOK = uintptr
type LPPAGEPAINTHOOK_func = func(param0 HWND, param1 uint32, param2 WPARAM, param3 LPARAM) uintptr

type LPPAGESETUPHOOK = uintptr
type LPPAGESETUPHOOK_func = func(param0 HWND, param1 uint32, param2 WPARAM, param3 LPARAM) uintptr

// interfaces

// 5852A2C3-6530-11D1-B6A3-0000F8757BF9
var IID_IPrintDialogCallback = syscall.GUID{0x5852A2C3, 0x6530, 0x11D1,
	[8]byte{0xB6, 0xA3, 0x00, 0x00, 0xF8, 0x75, 0x7B, 0xF9}}

type IPrintDialogCallbackInterface interface {
	IUnknownInterface
	InitDone() HRESULT
	SelectionChange() HRESULT
	HandleMessage(hDlg HWND, uMsg uint32, wParam WPARAM, lParam LPARAM, pResult *LRESULT) HRESULT
}

type IPrintDialogCallbackVtbl struct {
	IUnknownVtbl
	InitDone        uintptr
	SelectionChange uintptr
	HandleMessage   uintptr
}

type IPrintDialogCallback struct {
	IUnknown
}

func (this *IPrintDialogCallback) Vtbl() *IPrintDialogCallbackVtbl {
	return (*IPrintDialogCallbackVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IPrintDialogCallback) InitDone() HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().InitDone, uintptr(unsafe.Pointer(this)))
	return HRESULT(ret)
}

func (this *IPrintDialogCallback) SelectionChange() HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SelectionChange, uintptr(unsafe.Pointer(this)))
	return HRESULT(ret)
}

func (this *IPrintDialogCallback) HandleMessage(hDlg HWND, uMsg uint32, wParam WPARAM, lParam LPARAM, pResult *LRESULT) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().HandleMessage, uintptr(unsafe.Pointer(this)), hDlg, uintptr(uMsg), wParam, lParam, uintptr(unsafe.Pointer(pResult)))
	return HRESULT(ret)
}

// 509AAEDA-5639-11D1-B6A1-0000F8757BF9
var IID_IPrintDialogServices = syscall.GUID{0x509AAEDA, 0x5639, 0x11D1,
	[8]byte{0xB6, 0xA1, 0x00, 0x00, 0xF8, 0x75, 0x7B, 0xF9}}

type IPrintDialogServicesInterface interface {
	IUnknownInterface
	GetCurrentDevMode(pDevMode *DEVMODEA, pcbSize *uint32) HRESULT
	GetCurrentPrinterName(pPrinterName PWSTR, pcchSize *uint32) HRESULT
	GetCurrentPortName(pPortName PWSTR, pcchSize *uint32) HRESULT
}

type IPrintDialogServicesVtbl struct {
	IUnknownVtbl
	GetCurrentDevMode     uintptr
	GetCurrentPrinterName uintptr
	GetCurrentPortName    uintptr
}

type IPrintDialogServices struct {
	IUnknown
}

func (this *IPrintDialogServices) Vtbl() *IPrintDialogServicesVtbl {
	return (*IPrintDialogServicesVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IPrintDialogServices) GetCurrentDevMode(pDevMode *DEVMODEA, pcbSize *uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetCurrentDevMode, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pDevMode)), uintptr(unsafe.Pointer(pcbSize)))
	return HRESULT(ret)
}

func (this *IPrintDialogServices) GetCurrentPrinterName(pPrinterName PWSTR, pcchSize *uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetCurrentPrinterName, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pPrinterName)), uintptr(unsafe.Pointer(pcchSize)))
	return HRESULT(ret)
}

func (this *IPrintDialogServices) GetCurrentPortName(pPortName PWSTR, pcchSize *uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetCurrentPortName, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pPortName)), uintptr(unsafe.Pointer(pcchSize)))
	return HRESULT(ret)
}

var (
	pGetOpenFileNameA     uintptr
	pGetOpenFileNameW     uintptr
	pGetSaveFileNameA     uintptr
	pGetSaveFileNameW     uintptr
	pGetFileTitleA        uintptr
	pGetFileTitleW        uintptr
	pChooseColorA         uintptr
	pChooseColorW         uintptr
	pFindTextA            uintptr
	pFindTextW            uintptr
	pReplaceTextA         uintptr
	pReplaceTextW         uintptr
	pChooseFontA          uintptr
	pChooseFontW          uintptr
	pPrintDlgA            uintptr
	pPrintDlgW            uintptr
	pPrintDlgExA          uintptr
	pPrintDlgExW          uintptr
	pCommDlgExtendedError uintptr
	pPageSetupDlgA        uintptr
	pPageSetupDlgW        uintptr
)

func GetOpenFileNameA(param0 *OPENFILENAMEA) BOOL {
	addr := lazyAddr(&pGetOpenFileNameA, libComdlg32, "GetOpenFileNameA")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(param0)))
	return BOOL(ret)
}

var GetOpenFileName = GetOpenFileNameW

func GetOpenFileNameW(param0 *OPENFILENAMEW) BOOL {
	addr := lazyAddr(&pGetOpenFileNameW, libComdlg32, "GetOpenFileNameW")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(param0)))
	return BOOL(ret)
}

func GetSaveFileNameA(param0 *OPENFILENAMEA) BOOL {
	addr := lazyAddr(&pGetSaveFileNameA, libComdlg32, "GetSaveFileNameA")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(param0)))
	return BOOL(ret)
}

var GetSaveFileName = GetSaveFileNameW

func GetSaveFileNameW(param0 *OPENFILENAMEW) BOOL {
	addr := lazyAddr(&pGetSaveFileNameW, libComdlg32, "GetSaveFileNameW")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(param0)))
	return BOOL(ret)
}

func GetFileTitleA(param0 PSTR, Buf PSTR, cchSize uint16) int16 {
	addr := lazyAddr(&pGetFileTitleA, libComdlg32, "GetFileTitleA")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(param0)), uintptr(unsafe.Pointer(Buf)), uintptr(cchSize))
	return int16(ret)
}

var GetFileTitle = GetFileTitleW

func GetFileTitleW(param0 PWSTR, Buf PWSTR, cchSize uint16) int16 {
	addr := lazyAddr(&pGetFileTitleW, libComdlg32, "GetFileTitleW")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(param0)), uintptr(unsafe.Pointer(Buf)), uintptr(cchSize))
	return int16(ret)
}

func ChooseColorA(param0 *CHOOSECOLORA) BOOL {
	addr := lazyAddr(&pChooseColorA, libComdlg32, "ChooseColorA")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(param0)))
	return BOOL(ret)
}

var ChooseColor = ChooseColorW

func ChooseColorW(param0 *CHOOSECOLORW) BOOL {
	addr := lazyAddr(&pChooseColorW, libComdlg32, "ChooseColorW")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(param0)))
	return BOOL(ret)
}

func FindTextA(param0 *FINDREPLACEA) HWND {
	addr := lazyAddr(&pFindTextA, libComdlg32, "FindTextA")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(param0)))
	return ret
}

var FindText = FindTextW

func FindTextW(param0 *FINDREPLACEW) HWND {
	addr := lazyAddr(&pFindTextW, libComdlg32, "FindTextW")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(param0)))
	return ret
}

func ReplaceTextA(param0 *FINDREPLACEA) HWND {
	addr := lazyAddr(&pReplaceTextA, libComdlg32, "ReplaceTextA")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(param0)))
	return ret
}

var ReplaceText = ReplaceTextW

func ReplaceTextW(param0 *FINDREPLACEW) HWND {
	addr := lazyAddr(&pReplaceTextW, libComdlg32, "ReplaceTextW")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(param0)))
	return ret
}

func ChooseFontA(param0 *CHOOSEFONTA) BOOL {
	addr := lazyAddr(&pChooseFontA, libComdlg32, "ChooseFontA")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(param0)))
	return BOOL(ret)
}

var ChooseFont = ChooseFontW

func ChooseFontW(param0 *CHOOSEFONTW) BOOL {
	addr := lazyAddr(&pChooseFontW, libComdlg32, "ChooseFontW")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(param0)))
	return BOOL(ret)
}

func PrintDlgA(pPD *PRINTDLGA) BOOL {
	addr := lazyAddr(&pPrintDlgA, libComdlg32, "PrintDlgA")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pPD)))
	return BOOL(ret)
}

var PrintDlg = PrintDlgW

func PrintDlgW(pPD *PRINTDLGW) BOOL {
	addr := lazyAddr(&pPrintDlgW, libComdlg32, "PrintDlgW")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pPD)))
	return BOOL(ret)
}

func PrintDlgExA(pPD *PRINTDLGEXA) HRESULT {
	addr := lazyAddr(&pPrintDlgExA, libComdlg32, "PrintDlgExA")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pPD)))
	return HRESULT(ret)
}

var PrintDlgEx = PrintDlgExW

func PrintDlgExW(pPD *PRINTDLGEXW) HRESULT {
	addr := lazyAddr(&pPrintDlgExW, libComdlg32, "PrintDlgExW")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pPD)))
	return HRESULT(ret)
}

func CommDlgExtendedError() COMMON_DLG_ERRORS {
	addr := lazyAddr(&pCommDlgExtendedError, libComdlg32, "CommDlgExtendedError")
	ret, _, _ := syscall.SyscallN(addr)
	return COMMON_DLG_ERRORS(ret)
}

func PageSetupDlgA(param0 *PAGESETUPDLGA) BOOL {
	addr := lazyAddr(&pPageSetupDlgA, libComdlg32, "PageSetupDlgA")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(param0)))
	return BOOL(ret)
}

var PageSetupDlg = PageSetupDlgW

func PageSetupDlgW(param0 *PAGESETUPDLGW) BOOL {
	addr := lazyAddr(&pPageSetupDlgW, libComdlg32, "PageSetupDlgW")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(param0)))
	return BOOL(ret)
}
