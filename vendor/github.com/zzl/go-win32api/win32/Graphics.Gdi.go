package win32

import (
	"syscall"
	"unsafe"
)

type (
	HDC                     = uintptr
	CreatedHDC              = uintptr
	HBITMAP                 = uintptr
	HRGN                    = uintptr
	HPEN                    = uintptr
	HBRUSH                  = uintptr
	HFONT                   = uintptr
	HMETAFILE               = uintptr
	HENHMETAFILE            = uintptr
	HPALETTE                = uintptr
	HdcMetdataFileHandle    = uintptr
	HdcMetdataEnhFileHandle = uintptr
	HGDIOBJ                 = uintptr
	HMONITOR                = uintptr
)

const (
	GDI_ERROR                                       int32  = -1
	ERROR                                           uint32 = 0x0
	NULLREGION                                      uint32 = 0x1
	SIMPLEREGION                                    uint32 = 0x2
	COMPLEXREGION                                   uint32 = 0x3
	RGN_ERROR                                       uint32 = 0x0
	MAXSTRETCHBLTMODE                               uint32 = 0x4
	POLYFILL_LAST                                   uint32 = 0x2
	LAYOUT_BTT                                      uint32 = 0x2
	LAYOUT_VBH                                      uint32 = 0x4
	ASPECT_FILTERING                                uint32 = 0x1
	META_SETBKCOLOR                                 uint32 = 0x201
	META_SETBKMODE                                  uint32 = 0x102
	META_SETMAPMODE                                 uint32 = 0x103
	META_SETROP2                                    uint32 = 0x104
	META_SETRELABS                                  uint32 = 0x105
	META_SETPOLYFILLMODE                            uint32 = 0x106
	META_SETSTRETCHBLTMODE                          uint32 = 0x107
	META_SETTEXTCHAREXTRA                           uint32 = 0x108
	META_SETTEXTCOLOR                               uint32 = 0x209
	META_SETTEXTJUSTIFICATION                       uint32 = 0x20a
	META_SETWINDOWORG                               uint32 = 0x20b
	META_SETWINDOWEXT                               uint32 = 0x20c
	META_SETVIEWPORTORG                             uint32 = 0x20d
	META_SETVIEWPORTEXT                             uint32 = 0x20e
	META_OFFSETWINDOWORG                            uint32 = 0x20f
	META_SCALEWINDOWEXT                             uint32 = 0x410
	META_OFFSETVIEWPORTORG                          uint32 = 0x211
	META_SCALEVIEWPORTEXT                           uint32 = 0x412
	META_LINETO                                     uint32 = 0x213
	META_MOVETO                                     uint32 = 0x214
	META_EXCLUDECLIPRECT                            uint32 = 0x415
	META_INTERSECTCLIPRECT                          uint32 = 0x416
	META_ARC                                        uint32 = 0x817
	META_ELLIPSE                                    uint32 = 0x418
	META_FLOODFILL                                  uint32 = 0x419
	META_PIE                                        uint32 = 0x81a
	META_RECTANGLE                                  uint32 = 0x41b
	META_ROUNDRECT                                  uint32 = 0x61c
	META_PATBLT                                     uint32 = 0x61d
	META_SAVEDC                                     uint32 = 0x1e
	META_SETPIXEL                                   uint32 = 0x41f
	META_OFFSETCLIPRGN                              uint32 = 0x220
	META_TEXTOUT                                    uint32 = 0x521
	META_BITBLT                                     uint32 = 0x922
	META_STRETCHBLT                                 uint32 = 0xb23
	META_POLYGON                                    uint32 = 0x324
	META_POLYLINE                                   uint32 = 0x325
	META_ESCAPE                                     uint32 = 0x626
	META_RESTOREDC                                  uint32 = 0x127
	META_FILLREGION                                 uint32 = 0x228
	META_FRAMEREGION                                uint32 = 0x429
	META_INVERTREGION                               uint32 = 0x12a
	META_PAINTREGION                                uint32 = 0x12b
	META_SELECTCLIPREGION                           uint32 = 0x12c
	META_SELECTOBJECT                               uint32 = 0x12d
	META_SETTEXTALIGN                               uint32 = 0x12e
	META_CHORD                                      uint32 = 0x830
	META_SETMAPPERFLAGS                             uint32 = 0x231
	META_EXTTEXTOUT                                 uint32 = 0xa32
	META_SETDIBTODEV                                uint32 = 0xd33
	META_SELECTPALETTE                              uint32 = 0x234
	META_REALIZEPALETTE                             uint32 = 0x35
	META_ANIMATEPALETTE                             uint32 = 0x436
	META_SETPALENTRIES                              uint32 = 0x37
	META_POLYPOLYGON                                uint32 = 0x538
	META_RESIZEPALETTE                              uint32 = 0x139
	META_DIBBITBLT                                  uint32 = 0x940
	META_DIBSTRETCHBLT                              uint32 = 0xb41
	META_DIBCREATEPATTERNBRUSH                      uint32 = 0x142
	META_STRETCHDIB                                 uint32 = 0xf43
	META_EXTFLOODFILL                               uint32 = 0x548
	META_SETLAYOUT                                  uint32 = 0x149
	META_DELETEOBJECT                               uint32 = 0x1f0
	META_CREATEPALETTE                              uint32 = 0xf7
	META_CREATEPATTERNBRUSH                         uint32 = 0x1f9
	META_CREATEPENINDIRECT                          uint32 = 0x2fa
	META_CREATEFONTINDIRECT                         uint32 = 0x2fb
	META_CREATEBRUSHINDIRECT                        uint32 = 0x2fc
	META_CREATEREGION                               uint32 = 0x6ff
	NEWFRAME                                        uint32 = 0x1
	ABORTDOC                                        uint32 = 0x2
	NEXTBAND                                        uint32 = 0x3
	SETCOLORTABLE                                   uint32 = 0x4
	GETCOLORTABLE                                   uint32 = 0x5
	FLUSHOUTPUT                                     uint32 = 0x6
	DRAFTMODE                                       uint32 = 0x7
	QUERYESCSUPPORT                                 uint32 = 0x8
	SETABORTPROC                                    uint32 = 0x9
	STARTDOC                                        uint32 = 0xa
	ENDDOC                                          uint32 = 0xb
	GETPHYSPAGESIZE                                 uint32 = 0xc
	GETPRINTINGOFFSET                               uint32 = 0xd
	GETSCALINGFACTOR                                uint32 = 0xe
	MFCOMMENT                                       uint32 = 0xf
	GETPENWIDTH                                     uint32 = 0x10
	SETCOPYCOUNT                                    uint32 = 0x11
	SELECTPAPERSOURCE                               uint32 = 0x12
	DEVICEDATA                                      uint32 = 0x13
	PASSTHROUGH                                     uint32 = 0x13
	GETTECHNOLGY                                    uint32 = 0x14
	GETTECHNOLOGY                                   uint32 = 0x14
	SETLINECAP                                      uint32 = 0x15
	SETLINEJOIN                                     uint32 = 0x16
	SETMITERLIMIT                                   uint32 = 0x17
	BANDINFO                                        uint32 = 0x18
	DRAWPATTERNRECT                                 uint32 = 0x19
	GETVECTORPENSIZE                                uint32 = 0x1a
	GETVECTORBRUSHSIZE                              uint32 = 0x1b
	ENABLEDUPLEX                                    uint32 = 0x1c
	GETSETPAPERBINS                                 uint32 = 0x1d
	GETSETPRINTORIENT                               uint32 = 0x1e
	ENUMPAPERBINS                                   uint32 = 0x1f
	SETDIBSCALING                                   uint32 = 0x20
	EPSPRINTING                                     uint32 = 0x21
	ENUMPAPERMETRICS                                uint32 = 0x22
	GETSETPAPERMETRICS                              uint32 = 0x23
	POSTSCRIPT_DATA                                 uint32 = 0x25
	POSTSCRIPT_IGNORE                               uint32 = 0x26
	MOUSETRAILS                                     uint32 = 0x27
	GETDEVICEUNITS                                  uint32 = 0x2a
	GETEXTENDEDTEXTMETRICS                          uint32 = 0x100
	GETEXTENTTABLE                                  uint32 = 0x101
	GETPAIRKERNTABLE                                uint32 = 0x102
	GETTRACKKERNTABLE                               uint32 = 0x103
	EXTTEXTOUT                                      uint32 = 0x200
	GETFACENAME                                     uint32 = 0x201
	DOWNLOADFACE                                    uint32 = 0x202
	ENABLERELATIVEWIDTHS                            uint32 = 0x300
	ENABLEPAIRKERNING                               uint32 = 0x301
	SETKERNTRACK                                    uint32 = 0x302
	SETALLJUSTVALUES                                uint32 = 0x303
	SETCHARSET                                      uint32 = 0x304
	STRETCHBLT                                      uint32 = 0x800
	METAFILE_DRIVER                                 uint32 = 0x801
	GETSETSCREENPARAMS                              uint32 = 0xc00
	QUERYDIBSUPPORT                                 uint32 = 0xc01
	BEGIN_PATH                                      uint32 = 0x1000
	CLIP_TO_PATH                                    uint32 = 0x1001
	END_PATH                                        uint32 = 0x1002
	EXT_DEVICE_CAPS                                 uint32 = 0x1003
	RESTORE_CTM                                     uint32 = 0x1004
	SAVE_CTM                                        uint32 = 0x1005
	SET_ARC_DIRECTION                               uint32 = 0x1006
	SET_BACKGROUND_COLOR                            uint32 = 0x1007
	SET_POLY_MODE                                   uint32 = 0x1008
	SET_SCREEN_ANGLE                                uint32 = 0x1009
	SET_SPREAD                                      uint32 = 0x100a
	TRANSFORM_CTM                                   uint32 = 0x100b
	SET_CLIP_BOX                                    uint32 = 0x100c
	SET_BOUNDS                                      uint32 = 0x100d
	SET_MIRROR_MODE                                 uint32 = 0x100e
	OPENCHANNEL                                     uint32 = 0x100e
	DOWNLOADHEADER                                  uint32 = 0x100f
	CLOSECHANNEL                                    uint32 = 0x1010
	POSTSCRIPT_PASSTHROUGH                          uint32 = 0x1013
	ENCAPSULATED_POSTSCRIPT                         uint32 = 0x1014
	POSTSCRIPT_IDENTIFY                             uint32 = 0x1015
	POSTSCRIPT_INJECTION                            uint32 = 0x1016
	CHECKJPEGFORMAT                                 uint32 = 0x1017
	CHECKPNGFORMAT                                  uint32 = 0x1018
	GET_PS_FEATURESETTING                           uint32 = 0x1019
	GDIPLUS_TS_QUERYVER                             uint32 = 0x101a
	GDIPLUS_TS_RECORD                               uint32 = 0x101b
	MILCORE_TS_QUERYVER_RESULT_FALSE                uint32 = 0x0
	MILCORE_TS_QUERYVER_RESULT_TRUE                 uint32 = 0x7fffffff
	SPCLPASSTHROUGH2                                uint32 = 0x11d8
	PSIDENT_GDICENTRIC                              uint32 = 0x0
	PSIDENT_PSCENTRIC                               uint32 = 0x1
	PSINJECT_DLFONT                                 uint32 = 0xdddddddd
	FEATURESETTING_NUP                              uint32 = 0x0
	FEATURESETTING_OUTPUT                           uint32 = 0x1
	FEATURESETTING_PSLEVEL                          uint32 = 0x2
	FEATURESETTING_CUSTPAPER                        uint32 = 0x3
	FEATURESETTING_MIRROR                           uint32 = 0x4
	FEATURESETTING_NEGATIVE                         uint32 = 0x5
	FEATURESETTING_PROTOCOL                         uint32 = 0x6
	FEATURESETTING_PRIVATE_BEGIN                    uint32 = 0x1000
	FEATURESETTING_PRIVATE_END                      uint32 = 0x1fff
	PSPROTOCOL_ASCII                                uint32 = 0x0
	PSPROTOCOL_BCP                                  uint32 = 0x1
	PSPROTOCOL_TBCP                                 uint32 = 0x2
	PSPROTOCOL_BINARY                               uint32 = 0x3
	QDI_SETDIBITS                                   uint32 = 0x1
	QDI_GETDIBITS                                   uint32 = 0x2
	QDI_DIBTOSCREEN                                 uint32 = 0x4
	QDI_STRETCHDIB                                  uint32 = 0x8
	SP_NOTREPORTED                                  uint32 = 0x4000
	SP_ERROR                                        int32  = -1
	SP_APPABORT                                     int32  = -2
	SP_USERABORT                                    int32  = -3
	SP_OUTOFDISK                                    int32  = -4
	SP_OUTOFMEMORY                                  int32  = -5
	PR_JOBSTATUS                                    uint32 = 0x0
	LCS_CALIBRATED_RGB                              int32  = 0
	LCS_GM_BUSINESS                                 int32  = 1
	LCS_GM_GRAPHICS                                 int32  = 2
	LCS_GM_IMAGES                                   int32  = 4
	LCS_GM_ABS_COLORIMETRIC                         int32  = 8
	CM_OUT_OF_GAMUT                                 uint32 = 0xff
	CM_IN_GAMUT                                     uint32 = 0x0
	BI_RGB                                          int32  = 0
	BI_RLE8                                         int32  = 1
	BI_RLE4                                         int32  = 2
	BI_BITFIELDS                                    int32  = 3
	BI_JPEG                                         int32  = 4
	BI_PNG                                          int32  = 5
	TMPF_FIXED_PITCH                                uint32 = 0x1
	TMPF_VECTOR                                     uint32 = 0x2
	TMPF_DEVICE                                     uint32 = 0x8
	TMPF_TRUETYPE                                   uint32 = 0x4
	NTM_REGULAR                                     int32  = 64
	NTM_BOLD                                        int32  = 32
	NTM_ITALIC                                      int32  = 1
	NTM_NONNEGATIVE_AC                              uint32 = 0x10000
	NTM_PS_OPENTYPE                                 uint32 = 0x20000
	NTM_TT_OPENTYPE                                 uint32 = 0x40000
	NTM_MULTIPLEMASTER                              uint32 = 0x80000
	NTM_TYPE1                                       uint32 = 0x100000
	NTM_DSIG                                        uint32 = 0x200000
	LF_FACESIZE                                     uint32 = 0x20
	LF_FULLFACESIZE                                 uint32 = 0x40
	OUT_SCREEN_OUTLINE_PRECIS                       uint32 = 0x9
	CLEARTYPE_NATURAL_QUALITY                       uint32 = 0x6
	DEFAULT_PITCH                                   uint32 = 0x0
	FIXED_PITCH                                     uint32 = 0x1
	VARIABLE_PITCH                                  uint32 = 0x2
	MONO_FONT                                       uint32 = 0x8
	ANSI_CHARSET                                    uint32 = 0x0
	DEFAULT_CHARSET                                 uint32 = 0x1
	SYMBOL_CHARSET                                  uint32 = 0x2
	SHIFTJIS_CHARSET                                uint32 = 0x80
	HANGEUL_CHARSET                                 uint32 = 0x81
	HANGUL_CHARSET                                  uint32 = 0x81
	GB2312_CHARSET                                  uint32 = 0x86
	CHINESEBIG5_CHARSET                             uint32 = 0x88
	OEM_CHARSET                                     uint32 = 0xff
	JOHAB_CHARSET                                   uint32 = 0x82
	HEBREW_CHARSET                                  uint32 = 0xb1
	ARABIC_CHARSET                                  uint32 = 0xb2
	GREEK_CHARSET                                   uint32 = 0xa1
	TURKISH_CHARSET                                 uint32 = 0xa2
	VIETNAMESE_CHARSET                              uint32 = 0xa3
	THAI_CHARSET                                    uint32 = 0xde
	EASTEUROPE_CHARSET                              uint32 = 0xee
	RUSSIAN_CHARSET                                 uint32 = 0xcc
	MAC_CHARSET                                     uint32 = 0x4d
	BALTIC_CHARSET                                  uint32 = 0xba
	FS_LATIN1                                       int32  = 1
	FS_LATIN2                                       int32  = 2
	FS_CYRILLIC                                     int32  = 4
	FS_GREEK                                        int32  = 8
	FS_TURKISH                                      int32  = 16
	FS_HEBREW                                       int32  = 32
	FS_ARABIC                                       int32  = 64
	FS_BALTIC                                       int32  = 128
	FS_VIETNAMESE                                   int32  = 256
	FS_THAI                                         int32  = 65536
	FS_JISJAPAN                                     int32  = 131072
	FS_CHINESESIMP                                  int32  = 262144
	FS_WANSUNG                                      int32  = 524288
	FS_CHINESETRAD                                  int32  = 1048576
	FS_JOHAB                                        int32  = 2097152
	FS_SYMBOL                                       int32  = -2147483648
	FW_DONTCARE                                     uint32 = 0x0
	FW_THIN                                         uint32 = 0x64
	FW_EXTRALIGHT                                   uint32 = 0xc8
	FW_LIGHT                                        uint32 = 0x12c
	FW_NORMAL                                       uint32 = 0x190
	FW_MEDIUM                                       uint32 = 0x1f4
	FW_SEMIBOLD                                     uint32 = 0x258
	FW_BOLD                                         uint32 = 0x2bc
	FW_EXTRABOLD                                    uint32 = 0x320
	FW_HEAVY                                        uint32 = 0x384
	FW_ULTRALIGHT                                   uint32 = 0xc8
	FW_REGULAR                                      uint32 = 0x190
	FW_DEMIBOLD                                     uint32 = 0x258
	FW_ULTRABOLD                                    uint32 = 0x320
	FW_BLACK                                        uint32 = 0x384
	PANOSE_COUNT                                    uint32 = 0xa
	PAN_FAMILYTYPE_INDEX                            uint32 = 0x0
	PAN_SERIFSTYLE_INDEX                            uint32 = 0x1
	PAN_WEIGHT_INDEX                                uint32 = 0x2
	PAN_PROPORTION_INDEX                            uint32 = 0x3
	PAN_CONTRAST_INDEX                              uint32 = 0x4
	PAN_STROKEVARIATION_INDEX                       uint32 = 0x5
	PAN_ARMSTYLE_INDEX                              uint32 = 0x6
	PAN_LETTERFORM_INDEX                            uint32 = 0x7
	PAN_MIDLINE_INDEX                               uint32 = 0x8
	PAN_XHEIGHT_INDEX                               uint32 = 0x9
	PAN_CULTURE_LATIN                               uint32 = 0x0
	PAN_ANY                                         uint32 = 0x0
	PAN_NO_FIT                                      uint32 = 0x1
	PAN_FAMILY_TEXT_DISPLAY                         uint32 = 0x2
	PAN_FAMILY_SCRIPT                               uint32 = 0x3
	PAN_FAMILY_DECORATIVE                           uint32 = 0x4
	PAN_FAMILY_PICTORIAL                            uint32 = 0x5
	PAN_SERIF_COVE                                  uint32 = 0x2
	PAN_SERIF_OBTUSE_COVE                           uint32 = 0x3
	PAN_SERIF_SQUARE_COVE                           uint32 = 0x4
	PAN_SERIF_OBTUSE_SQUARE_COVE                    uint32 = 0x5
	PAN_SERIF_SQUARE                                uint32 = 0x6
	PAN_SERIF_THIN                                  uint32 = 0x7
	PAN_SERIF_BONE                                  uint32 = 0x8
	PAN_SERIF_EXAGGERATED                           uint32 = 0x9
	PAN_SERIF_TRIANGLE                              uint32 = 0xa
	PAN_SERIF_NORMAL_SANS                           uint32 = 0xb
	PAN_SERIF_OBTUSE_SANS                           uint32 = 0xc
	PAN_SERIF_PERP_SANS                             uint32 = 0xd
	PAN_SERIF_FLARED                                uint32 = 0xe
	PAN_SERIF_ROUNDED                               uint32 = 0xf
	PAN_WEIGHT_VERY_LIGHT                           uint32 = 0x2
	PAN_WEIGHT_LIGHT                                uint32 = 0x3
	PAN_WEIGHT_THIN                                 uint32 = 0x4
	PAN_WEIGHT_BOOK                                 uint32 = 0x5
	PAN_WEIGHT_MEDIUM                               uint32 = 0x6
	PAN_WEIGHT_DEMI                                 uint32 = 0x7
	PAN_WEIGHT_BOLD                                 uint32 = 0x8
	PAN_WEIGHT_HEAVY                                uint32 = 0x9
	PAN_WEIGHT_BLACK                                uint32 = 0xa
	PAN_WEIGHT_NORD                                 uint32 = 0xb
	PAN_PROP_OLD_STYLE                              uint32 = 0x2
	PAN_PROP_MODERN                                 uint32 = 0x3
	PAN_PROP_EVEN_WIDTH                             uint32 = 0x4
	PAN_PROP_EXPANDED                               uint32 = 0x5
	PAN_PROP_CONDENSED                              uint32 = 0x6
	PAN_PROP_VERY_EXPANDED                          uint32 = 0x7
	PAN_PROP_VERY_CONDENSED                         uint32 = 0x8
	PAN_PROP_MONOSPACED                             uint32 = 0x9
	PAN_CONTRAST_NONE                               uint32 = 0x2
	PAN_CONTRAST_VERY_LOW                           uint32 = 0x3
	PAN_CONTRAST_LOW                                uint32 = 0x4
	PAN_CONTRAST_MEDIUM_LOW                         uint32 = 0x5
	PAN_CONTRAST_MEDIUM                             uint32 = 0x6
	PAN_CONTRAST_MEDIUM_HIGH                        uint32 = 0x7
	PAN_CONTRAST_HIGH                               uint32 = 0x8
	PAN_CONTRAST_VERY_HIGH                          uint32 = 0x9
	PAN_STROKE_GRADUAL_DIAG                         uint32 = 0x2
	PAN_STROKE_GRADUAL_TRAN                         uint32 = 0x3
	PAN_STROKE_GRADUAL_VERT                         uint32 = 0x4
	PAN_STROKE_GRADUAL_HORZ                         uint32 = 0x5
	PAN_STROKE_RAPID_VERT                           uint32 = 0x6
	PAN_STROKE_RAPID_HORZ                           uint32 = 0x7
	PAN_STROKE_INSTANT_VERT                         uint32 = 0x8
	PAN_STRAIGHT_ARMS_HORZ                          uint32 = 0x2
	PAN_STRAIGHT_ARMS_WEDGE                         uint32 = 0x3
	PAN_STRAIGHT_ARMS_VERT                          uint32 = 0x4
	PAN_STRAIGHT_ARMS_SINGLE_SERIF                  uint32 = 0x5
	PAN_STRAIGHT_ARMS_DOUBLE_SERIF                  uint32 = 0x6
	PAN_BENT_ARMS_HORZ                              uint32 = 0x7
	PAN_BENT_ARMS_WEDGE                             uint32 = 0x8
	PAN_BENT_ARMS_VERT                              uint32 = 0x9
	PAN_BENT_ARMS_SINGLE_SERIF                      uint32 = 0xa
	PAN_BENT_ARMS_DOUBLE_SERIF                      uint32 = 0xb
	PAN_LETT_NORMAL_CONTACT                         uint32 = 0x2
	PAN_LETT_NORMAL_WEIGHTED                        uint32 = 0x3
	PAN_LETT_NORMAL_BOXED                           uint32 = 0x4
	PAN_LETT_NORMAL_FLATTENED                       uint32 = 0x5
	PAN_LETT_NORMAL_ROUNDED                         uint32 = 0x6
	PAN_LETT_NORMAL_OFF_CENTER                      uint32 = 0x7
	PAN_LETT_NORMAL_SQUARE                          uint32 = 0x8
	PAN_LETT_OBLIQUE_CONTACT                        uint32 = 0x9
	PAN_LETT_OBLIQUE_WEIGHTED                       uint32 = 0xa
	PAN_LETT_OBLIQUE_BOXED                          uint32 = 0xb
	PAN_LETT_OBLIQUE_FLATTENED                      uint32 = 0xc
	PAN_LETT_OBLIQUE_ROUNDED                        uint32 = 0xd
	PAN_LETT_OBLIQUE_OFF_CENTER                     uint32 = 0xe
	PAN_LETT_OBLIQUE_SQUARE                         uint32 = 0xf
	PAN_MIDLINE_STANDARD_TRIMMED                    uint32 = 0x2
	PAN_MIDLINE_STANDARD_POINTED                    uint32 = 0x3
	PAN_MIDLINE_STANDARD_SERIFED                    uint32 = 0x4
	PAN_MIDLINE_HIGH_TRIMMED                        uint32 = 0x5
	PAN_MIDLINE_HIGH_POINTED                        uint32 = 0x6
	PAN_MIDLINE_HIGH_SERIFED                        uint32 = 0x7
	PAN_MIDLINE_CONSTANT_TRIMMED                    uint32 = 0x8
	PAN_MIDLINE_CONSTANT_POINTED                    uint32 = 0x9
	PAN_MIDLINE_CONSTANT_SERIFED                    uint32 = 0xa
	PAN_MIDLINE_LOW_TRIMMED                         uint32 = 0xb
	PAN_MIDLINE_LOW_POINTED                         uint32 = 0xc
	PAN_MIDLINE_LOW_SERIFED                         uint32 = 0xd
	PAN_XHEIGHT_CONSTANT_SMALL                      uint32 = 0x2
	PAN_XHEIGHT_CONSTANT_STD                        uint32 = 0x3
	PAN_XHEIGHT_CONSTANT_LARGE                      uint32 = 0x4
	PAN_XHEIGHT_DUCKING_SMALL                       uint32 = 0x5
	PAN_XHEIGHT_DUCKING_STD                         uint32 = 0x6
	PAN_XHEIGHT_DUCKING_LARGE                       uint32 = 0x7
	ELF_VENDOR_SIZE                                 uint32 = 0x4
	ELF_VERSION                                     uint32 = 0x0
	ELF_CULTURE_LATIN                               uint32 = 0x0
	RASTER_FONTTYPE                                 uint32 = 0x1
	DEVICE_FONTTYPE                                 uint32 = 0x2
	TRUETYPE_FONTTYPE                               uint32 = 0x4
	PC_RESERVED                                     uint32 = 0x1
	PC_EXPLICIT                                     uint32 = 0x2
	PC_NOCOLLAPSE                                   uint32 = 0x4
	BKMODE_LAST                                     uint32 = 0x2
	GM_LAST                                         uint32 = 0x2
	PT_CLOSEFIGURE                                  uint32 = 0x1
	PT_LINETO                                       uint32 = 0x2
	PT_BEZIERTO                                     uint32 = 0x4
	PT_MOVETO                                       uint32 = 0x6
	ABSOLUTE                                        uint32 = 0x1
	RELATIVE                                        uint32 = 0x2
	STOCK_LAST                                      uint32 = 0x13
	CLR_INVALID                                     uint32 = 0xffffffff
	BS_SOLID                                        uint32 = 0x0
	BS_NULL                                         uint32 = 0x1
	BS_HOLLOW                                       uint32 = 0x1
	BS_HATCHED                                      uint32 = 0x2
	BS_PATTERN                                      uint32 = 0x3
	BS_INDEXED                                      uint32 = 0x4
	BS_DIBPATTERN                                   uint32 = 0x5
	BS_DIBPATTERNPT                                 uint32 = 0x6
	BS_PATTERN8X8                                   uint32 = 0x7
	BS_DIBPATTERN8X8                                uint32 = 0x8
	BS_MONOPATTERN                                  uint32 = 0x9
	HS_API_MAX                                      uint32 = 0xc
	DT_PLOTTER                                      uint32 = 0x0
	DT_RASDISPLAY                                   uint32 = 0x1
	DT_RASPRINTER                                   uint32 = 0x2
	DT_RASCAMERA                                    uint32 = 0x3
	DT_CHARSTREAM                                   uint32 = 0x4
	DT_METAFILE                                     uint32 = 0x5
	DT_DISPFILE                                     uint32 = 0x6
	CC_NONE                                         uint32 = 0x0
	CC_CIRCLES                                      uint32 = 0x1
	CC_PIE                                          uint32 = 0x2
	CC_CHORD                                        uint32 = 0x4
	CC_ELLIPSES                                     uint32 = 0x8
	CC_WIDE                                         uint32 = 0x10
	CC_STYLED                                       uint32 = 0x20
	CC_WIDESTYLED                                   uint32 = 0x40
	CC_INTERIORS                                    uint32 = 0x80
	CC_ROUNDRECT                                    uint32 = 0x100
	LC_NONE                                         uint32 = 0x0
	LC_POLYLINE                                     uint32 = 0x2
	LC_MARKER                                       uint32 = 0x4
	LC_POLYMARKER                                   uint32 = 0x8
	LC_WIDE                                         uint32 = 0x10
	LC_STYLED                                       uint32 = 0x20
	LC_WIDESTYLED                                   uint32 = 0x40
	LC_INTERIORS                                    uint32 = 0x80
	PC_NONE                                         uint32 = 0x0
	PC_POLYGON                                      uint32 = 0x1
	PC_RECTANGLE                                    uint32 = 0x2
	PC_WINDPOLYGON                                  uint32 = 0x4
	PC_TRAPEZOID                                    uint32 = 0x4
	PC_SCANLINE                                     uint32 = 0x8
	PC_WIDE                                         uint32 = 0x10
	PC_STYLED                                       uint32 = 0x20
	PC_WIDESTYLED                                   uint32 = 0x40
	PC_INTERIORS                                    uint32 = 0x80
	PC_POLYPOLYGON                                  uint32 = 0x100
	PC_PATHS                                        uint32 = 0x200
	CP_NONE                                         uint32 = 0x0
	CP_RECTANGLE                                    uint32 = 0x1
	CP_REGION                                       uint32 = 0x2
	TC_OP_CHARACTER                                 uint32 = 0x1
	TC_OP_STROKE                                    uint32 = 0x2
	TC_CP_STROKE                                    uint32 = 0x4
	TC_CR_90                                        uint32 = 0x8
	TC_CR_ANY                                       uint32 = 0x10
	TC_SF_X_YINDEP                                  uint32 = 0x20
	TC_SA_DOUBLE                                    uint32 = 0x40
	TC_SA_INTEGER                                   uint32 = 0x80
	TC_SA_CONTIN                                    uint32 = 0x100
	TC_EA_DOUBLE                                    uint32 = 0x200
	TC_IA_ABLE                                      uint32 = 0x400
	TC_UA_ABLE                                      uint32 = 0x800
	TC_SO_ABLE                                      uint32 = 0x1000
	TC_RA_ABLE                                      uint32 = 0x2000
	TC_VA_ABLE                                      uint32 = 0x4000
	TC_RESERVED                                     uint32 = 0x8000
	TC_SCROLLBLT                                    uint32 = 0x10000
	RC_BITBLT                                       uint32 = 0x1
	RC_BANDING                                      uint32 = 0x2
	RC_SCALING                                      uint32 = 0x4
	RC_BITMAP64                                     uint32 = 0x8
	RC_GDI20_OUTPUT                                 uint32 = 0x10
	RC_GDI20_STATE                                  uint32 = 0x20
	RC_SAVEBITMAP                                   uint32 = 0x40
	RC_DI_BITMAP                                    uint32 = 0x80
	RC_PALETTE                                      uint32 = 0x100
	RC_DIBTODEV                                     uint32 = 0x200
	RC_BIGFONT                                      uint32 = 0x400
	RC_STRETCHBLT                                   uint32 = 0x800
	RC_FLOODFILL                                    uint32 = 0x1000
	RC_STRETCHDIB                                   uint32 = 0x2000
	RC_OP_DX_OUTPUT                                 uint32 = 0x4000
	RC_DEVBITS                                      uint32 = 0x8000
	SB_NONE                                         uint32 = 0x0
	SB_CONST_ALPHA                                  uint32 = 0x1
	SB_PIXEL_ALPHA                                  uint32 = 0x2
	SB_PREMULT_ALPHA                                uint32 = 0x4
	SB_GRAD_RECT                                    uint32 = 0x10
	SB_GRAD_TRI                                     uint32 = 0x20
	CM_NONE                                         uint32 = 0x0
	CM_DEVICE_ICM                                   uint32 = 0x1
	CM_GAMMA_RAMP                                   uint32 = 0x2
	CM_CMYK_COLOR                                   uint32 = 0x4
	SYSPAL_ERROR                                    uint32 = 0x0
	CBM_INIT                                        int32  = 4
	CCHFORMNAME                                     uint32 = 0x20
	DM_SPECVERSION                                  uint32 = 0x401
	DM_ORIENTATION                                  int32  = 1
	DM_PAPERSIZE                                    int32  = 2
	DM_PAPERLENGTH                                  int32  = 4
	DM_PAPERWIDTH                                   int32  = 8
	DM_SCALE                                        int32  = 16
	DM_POSITION                                     int32  = 32
	DM_NUP                                          int32  = 64
	DM_DISPLAYORIENTATION                           int32  = 128
	DM_COPIES                                       int32  = 256
	DM_DEFAULTSOURCE                                int32  = 512
	DM_PRINTQUALITY                                 int32  = 1024
	DM_COLOR                                        int32  = 2048
	DM_DUPLEX                                       int32  = 4096
	DM_YRESOLUTION                                  int32  = 8192
	DM_TTOPTION                                     int32  = 16384
	DM_COLLATE                                      int32  = 32768
	DM_FORMNAME                                     int32  = 65536
	DM_LOGPIXELS                                    int32  = 131072
	DM_BITSPERPEL                                   int32  = 262144
	DM_PELSWIDTH                                    int32  = 524288
	DM_PELSHEIGHT                                   int32  = 1048576
	DM_DISPLAYFLAGS                                 int32  = 2097152
	DM_DISPLAYFREQUENCY                             int32  = 4194304
	DM_ICMMETHOD                                    int32  = 8388608
	DM_ICMINTENT                                    int32  = 16777216
	DM_MEDIATYPE                                    int32  = 33554432
	DM_DITHERTYPE                                   int32  = 67108864
	DM_PANNINGWIDTH                                 int32  = 134217728
	DM_PANNINGHEIGHT                                int32  = 268435456
	DM_DISPLAYFIXEDOUTPUT                           int32  = 536870912
	DMORIENT_PORTRAIT                               uint32 = 0x1
	DMORIENT_LANDSCAPE                              uint32 = 0x2
	DMPAPER_LETTER                                  uint32 = 0x1
	DMPAPER_LETTERSMALL                             uint32 = 0x2
	DMPAPER_TABLOID                                 uint32 = 0x3
	DMPAPER_LEDGER                                  uint32 = 0x4
	DMPAPER_LEGAL                                   uint32 = 0x5
	DMPAPER_STATEMENT                               uint32 = 0x6
	DMPAPER_EXECUTIVE                               uint32 = 0x7
	DMPAPER_A3                                      uint32 = 0x8
	DMPAPER_A4                                      uint32 = 0x9
	DMPAPER_A4SMALL                                 uint32 = 0xa
	DMPAPER_A5                                      uint32 = 0xb
	DMPAPER_B4                                      uint32 = 0xc
	DMPAPER_B5                                      uint32 = 0xd
	DMPAPER_FOLIO                                   uint32 = 0xe
	DMPAPER_QUARTO                                  uint32 = 0xf
	DMPAPER_10X14                                   uint32 = 0x10
	DMPAPER_11X17                                   uint32 = 0x11
	DMPAPER_NOTE                                    uint32 = 0x12
	DMPAPER_ENV_9                                   uint32 = 0x13
	DMPAPER_ENV_10                                  uint32 = 0x14
	DMPAPER_ENV_11                                  uint32 = 0x15
	DMPAPER_ENV_12                                  uint32 = 0x16
	DMPAPER_ENV_14                                  uint32 = 0x17
	DMPAPER_CSHEET                                  uint32 = 0x18
	DMPAPER_DSHEET                                  uint32 = 0x19
	DMPAPER_ESHEET                                  uint32 = 0x1a
	DMPAPER_ENV_DL                                  uint32 = 0x1b
	DMPAPER_ENV_C5                                  uint32 = 0x1c
	DMPAPER_ENV_C3                                  uint32 = 0x1d
	DMPAPER_ENV_C4                                  uint32 = 0x1e
	DMPAPER_ENV_C6                                  uint32 = 0x1f
	DMPAPER_ENV_C65                                 uint32 = 0x20
	DMPAPER_ENV_B4                                  uint32 = 0x21
	DMPAPER_ENV_B5                                  uint32 = 0x22
	DMPAPER_ENV_B6                                  uint32 = 0x23
	DMPAPER_ENV_ITALY                               uint32 = 0x24
	DMPAPER_ENV_MONARCH                             uint32 = 0x25
	DMPAPER_ENV_PERSONAL                            uint32 = 0x26
	DMPAPER_FANFOLD_US                              uint32 = 0x27
	DMPAPER_FANFOLD_STD_GERMAN                      uint32 = 0x28
	DMPAPER_FANFOLD_LGL_GERMAN                      uint32 = 0x29
	DMPAPER_ISO_B4                                  uint32 = 0x2a
	DMPAPER_JAPANESE_POSTCARD                       uint32 = 0x2b
	DMPAPER_9X11                                    uint32 = 0x2c
	DMPAPER_10X11                                   uint32 = 0x2d
	DMPAPER_15X11                                   uint32 = 0x2e
	DMPAPER_ENV_INVITE                              uint32 = 0x2f
	DMPAPER_RESERVED_48                             uint32 = 0x30
	DMPAPER_RESERVED_49                             uint32 = 0x31
	DMPAPER_LETTER_EXTRA                            uint32 = 0x32
	DMPAPER_LEGAL_EXTRA                             uint32 = 0x33
	DMPAPER_TABLOID_EXTRA                           uint32 = 0x34
	DMPAPER_A4_EXTRA                                uint32 = 0x35
	DMPAPER_LETTER_TRANSVERSE                       uint32 = 0x36
	DMPAPER_A4_TRANSVERSE                           uint32 = 0x37
	DMPAPER_LETTER_EXTRA_TRANSVERSE                 uint32 = 0x38
	DMPAPER_A_PLUS                                  uint32 = 0x39
	DMPAPER_B_PLUS                                  uint32 = 0x3a
	DMPAPER_LETTER_PLUS                             uint32 = 0x3b
	DMPAPER_A4_PLUS                                 uint32 = 0x3c
	DMPAPER_A5_TRANSVERSE                           uint32 = 0x3d
	DMPAPER_B5_TRANSVERSE                           uint32 = 0x3e
	DMPAPER_A3_EXTRA                                uint32 = 0x3f
	DMPAPER_A5_EXTRA                                uint32 = 0x40
	DMPAPER_B5_EXTRA                                uint32 = 0x41
	DMPAPER_A2                                      uint32 = 0x42
	DMPAPER_A3_TRANSVERSE                           uint32 = 0x43
	DMPAPER_A3_EXTRA_TRANSVERSE                     uint32 = 0x44
	DMPAPER_DBL_JAPANESE_POSTCARD                   uint32 = 0x45
	DMPAPER_A6                                      uint32 = 0x46
	DMPAPER_JENV_KAKU2                              uint32 = 0x47
	DMPAPER_JENV_KAKU3                              uint32 = 0x48
	DMPAPER_JENV_CHOU3                              uint32 = 0x49
	DMPAPER_JENV_CHOU4                              uint32 = 0x4a
	DMPAPER_LETTER_ROTATED                          uint32 = 0x4b
	DMPAPER_A3_ROTATED                              uint32 = 0x4c
	DMPAPER_A4_ROTATED                              uint32 = 0x4d
	DMPAPER_A5_ROTATED                              uint32 = 0x4e
	DMPAPER_B4_JIS_ROTATED                          uint32 = 0x4f
	DMPAPER_B5_JIS_ROTATED                          uint32 = 0x50
	DMPAPER_JAPANESE_POSTCARD_ROTATED               uint32 = 0x51
	DMPAPER_DBL_JAPANESE_POSTCARD_ROTATED           uint32 = 0x52
	DMPAPER_A6_ROTATED                              uint32 = 0x53
	DMPAPER_JENV_KAKU2_ROTATED                      uint32 = 0x54
	DMPAPER_JENV_KAKU3_ROTATED                      uint32 = 0x55
	DMPAPER_JENV_CHOU3_ROTATED                      uint32 = 0x56
	DMPAPER_JENV_CHOU4_ROTATED                      uint32 = 0x57
	DMPAPER_B6_JIS                                  uint32 = 0x58
	DMPAPER_B6_JIS_ROTATED                          uint32 = 0x59
	DMPAPER_12X11                                   uint32 = 0x5a
	DMPAPER_JENV_YOU4                               uint32 = 0x5b
	DMPAPER_JENV_YOU4_ROTATED                       uint32 = 0x5c
	DMPAPER_P16K                                    uint32 = 0x5d
	DMPAPER_P32K                                    uint32 = 0x5e
	DMPAPER_P32KBIG                                 uint32 = 0x5f
	DMPAPER_PENV_1                                  uint32 = 0x60
	DMPAPER_PENV_2                                  uint32 = 0x61
	DMPAPER_PENV_3                                  uint32 = 0x62
	DMPAPER_PENV_4                                  uint32 = 0x63
	DMPAPER_PENV_5                                  uint32 = 0x64
	DMPAPER_PENV_6                                  uint32 = 0x65
	DMPAPER_PENV_7                                  uint32 = 0x66
	DMPAPER_PENV_8                                  uint32 = 0x67
	DMPAPER_PENV_9                                  uint32 = 0x68
	DMPAPER_PENV_10                                 uint32 = 0x69
	DMPAPER_P16K_ROTATED                            uint32 = 0x6a
	DMPAPER_P32K_ROTATED                            uint32 = 0x6b
	DMPAPER_P32KBIG_ROTATED                         uint32 = 0x6c
	DMPAPER_PENV_1_ROTATED                          uint32 = 0x6d
	DMPAPER_PENV_2_ROTATED                          uint32 = 0x6e
	DMPAPER_PENV_3_ROTATED                          uint32 = 0x6f
	DMPAPER_PENV_4_ROTATED                          uint32 = 0x70
	DMPAPER_PENV_5_ROTATED                          uint32 = 0x71
	DMPAPER_PENV_6_ROTATED                          uint32 = 0x72
	DMPAPER_PENV_7_ROTATED                          uint32 = 0x73
	DMPAPER_PENV_8_ROTATED                          uint32 = 0x74
	DMPAPER_PENV_9_ROTATED                          uint32 = 0x75
	DMPAPER_PENV_10_ROTATED                         uint32 = 0x76
	DMPAPER_LAST                                    uint32 = 0x76
	DMPAPER_USER                                    uint32 = 0x100
	DMBIN_UPPER                                     uint32 = 0x1
	DMBIN_ONLYONE                                   uint32 = 0x1
	DMBIN_LOWER                                     uint32 = 0x2
	DMBIN_MIDDLE                                    uint32 = 0x3
	DMBIN_MANUAL                                    uint32 = 0x4
	DMBIN_ENVELOPE                                  uint32 = 0x5
	DMBIN_ENVMANUAL                                 uint32 = 0x6
	DMBIN_AUTO                                      uint32 = 0x7
	DMBIN_TRACTOR                                   uint32 = 0x8
	DMBIN_SMALLFMT                                  uint32 = 0x9
	DMBIN_LARGEFMT                                  uint32 = 0xa
	DMBIN_LARGECAPACITY                             uint32 = 0xb
	DMBIN_CASSETTE                                  uint32 = 0xe
	DMBIN_FORMSOURCE                                uint32 = 0xf
	DMBIN_LAST                                      uint32 = 0xf
	DMBIN_USER                                      uint32 = 0x100
	DMRES_DRAFT                                     int32  = -1
	DMRES_LOW                                       int32  = -2
	DMRES_MEDIUM                                    int32  = -3
	DMRES_HIGH                                      int32  = -4
	DMCOLOR_MONOCHROME                              uint32 = 0x1
	DMCOLOR_COLOR                                   uint32 = 0x2
	DMDUP_SIMPLEX                                   uint32 = 0x1
	DMDUP_VERTICAL                                  uint32 = 0x2
	DMDUP_HORIZONTAL                                uint32 = 0x3
	DMTT_BITMAP                                     uint32 = 0x1
	DMTT_DOWNLOAD                                   uint32 = 0x2
	DMTT_SUBDEV                                     uint32 = 0x3
	DMTT_DOWNLOAD_OUTLINE                           uint32 = 0x4
	DMCOLLATE_FALSE                                 uint32 = 0x0
	DMCOLLATE_TRUE                                  uint32 = 0x1
	DMDO_DEFAULT                                    uint32 = 0x0
	DMDO_90                                         uint32 = 0x1
	DMDO_180                                        uint32 = 0x2
	DMDO_270                                        uint32 = 0x3
	DMDFO_DEFAULT                                   uint32 = 0x0
	DMDFO_STRETCH                                   uint32 = 0x1
	DMDFO_CENTER                                    uint32 = 0x2
	DM_INTERLACED                                   uint32 = 0x2
	DMDISPLAYFLAGS_TEXTMODE                         uint32 = 0x4
	DMNUP_SYSTEM                                    uint32 = 0x1
	DMNUP_ONEUP                                     uint32 = 0x2
	DMICMMETHOD_NONE                                uint32 = 0x1
	DMICMMETHOD_SYSTEM                              uint32 = 0x2
	DMICMMETHOD_DRIVER                              uint32 = 0x3
	DMICMMETHOD_DEVICE                              uint32 = 0x4
	DMICMMETHOD_USER                                uint32 = 0x100
	DMICM_SATURATE                                  uint32 = 0x1
	DMICM_CONTRAST                                  uint32 = 0x2
	DMICM_COLORIMETRIC                              uint32 = 0x3
	DMICM_ABS_COLORIMETRIC                          uint32 = 0x4
	DMICM_USER                                      uint32 = 0x100
	DMMEDIA_STANDARD                                uint32 = 0x1
	DMMEDIA_TRANSPARENCY                            uint32 = 0x2
	DMMEDIA_GLOSSY                                  uint32 = 0x3
	DMMEDIA_USER                                    uint32 = 0x100
	DMDITHER_NONE                                   uint32 = 0x1
	DMDITHER_COARSE                                 uint32 = 0x2
	DMDITHER_FINE                                   uint32 = 0x3
	DMDITHER_LINEART                                uint32 = 0x4
	DMDITHER_ERRORDIFFUSION                         uint32 = 0x5
	DMDITHER_RESERVED6                              uint32 = 0x6
	DMDITHER_RESERVED7                              uint32 = 0x7
	DMDITHER_RESERVED8                              uint32 = 0x8
	DMDITHER_RESERVED9                              uint32 = 0x9
	DMDITHER_GRAYSCALE                              uint32 = 0xa
	DMDITHER_USER                                   uint32 = 0x100
	DISPLAY_DEVICE_ATTACHED_TO_DESKTOP              uint32 = 0x1
	DISPLAY_DEVICE_MULTI_DRIVER                     uint32 = 0x2
	DISPLAY_DEVICE_PRIMARY_DEVICE                   uint32 = 0x4
	DISPLAY_DEVICE_MIRRORING_DRIVER                 uint32 = 0x8
	DISPLAY_DEVICE_VGA_COMPATIBLE                   uint32 = 0x10
	DISPLAY_DEVICE_REMOVABLE                        uint32 = 0x20
	DISPLAY_DEVICE_ACC_DRIVER                       uint32 = 0x40
	DISPLAY_DEVICE_MODESPRUNED                      uint32 = 0x8000000
	DISPLAY_DEVICE_RDPUDD                           uint32 = 0x1000000
	DISPLAY_DEVICE_REMOTE                           uint32 = 0x4000000
	DISPLAY_DEVICE_DISCONNECT                       uint32 = 0x2000000
	DISPLAY_DEVICE_TS_COMPATIBLE                    uint32 = 0x200000
	DISPLAY_DEVICE_UNSAFE_MODES_ON                  uint32 = 0x80000
	DISPLAY_DEVICE_ACTIVE                           uint32 = 0x1
	DISPLAY_DEVICE_ATTACHED                         uint32 = 0x2
	DISPLAYCONFIG_MAXPATH                           uint32 = 0x400
	DISPLAYCONFIG_PATH_MODE_IDX_INVALID             uint32 = 0xffffffff
	DISPLAYCONFIG_PATH_TARGET_MODE_IDX_INVALID      uint32 = 0xffff
	DISPLAYCONFIG_PATH_DESKTOP_IMAGE_IDX_INVALID    uint32 = 0xffff
	DISPLAYCONFIG_PATH_SOURCE_MODE_IDX_INVALID      uint32 = 0xffff
	DISPLAYCONFIG_PATH_CLONE_GROUP_INVALID          uint32 = 0xffff
	DISPLAYCONFIG_SOURCE_IN_USE                     uint32 = 0x1
	DISPLAYCONFIG_TARGET_IN_USE                     uint32 = 0x1
	DISPLAYCONFIG_TARGET_FORCIBLE                   uint32 = 0x2
	DISPLAYCONFIG_TARGET_FORCED_AVAILABILITY_BOOT   uint32 = 0x4
	DISPLAYCONFIG_TARGET_FORCED_AVAILABILITY_PATH   uint32 = 0x8
	DISPLAYCONFIG_TARGET_FORCED_AVAILABILITY_SYSTEM uint32 = 0x10
	DISPLAYCONFIG_TARGET_IS_HMD                     uint32 = 0x20
	DISPLAYCONFIG_PATH_ACTIVE                       uint32 = 0x1
	DISPLAYCONFIG_PATH_PREFERRED_UNSCALED           uint32 = 0x4
	DISPLAYCONFIG_PATH_SUPPORT_VIRTUAL_MODE         uint32 = 0x8
	DISPLAYCONFIG_PATH_VALID_FLAGS                  uint32 = 0x1d
	QDC_ALL_PATHS                                   uint32 = 0x1
	QDC_ONLY_ACTIVE_PATHS                           uint32 = 0x2
	QDC_DATABASE_CURRENT                            uint32 = 0x4
	QDC_VIRTUAL_MODE_AWARE                          uint32 = 0x10
	QDC_INCLUDE_HMD                                 uint32 = 0x20
	QDC_VIRTUAL_REFRESH_RATE_AWARE                  uint32 = 0x40
	SDC_TOPOLOGY_INTERNAL                           uint32 = 0x1
	SDC_TOPOLOGY_CLONE                              uint32 = 0x2
	SDC_TOPOLOGY_EXTEND                             uint32 = 0x4
	SDC_TOPOLOGY_EXTERNAL                           uint32 = 0x8
	SDC_TOPOLOGY_SUPPLIED                           uint32 = 0x10
	SDC_USE_SUPPLIED_DISPLAY_CONFIG                 uint32 = 0x20
	SDC_VALIDATE                                    uint32 = 0x40
	SDC_APPLY                                       uint32 = 0x80
	SDC_NO_OPTIMIZATION                             uint32 = 0x100
	SDC_SAVE_TO_DATABASE                            uint32 = 0x200
	SDC_ALLOW_CHANGES                               uint32 = 0x400
	SDC_PATH_PERSIST_IF_REQUIRED                    uint32 = 0x800
	SDC_FORCE_MODE_ENUMERATION                      uint32 = 0x1000
	SDC_ALLOW_PATH_ORDER_CHANGES                    uint32 = 0x2000
	SDC_VIRTUAL_MODE_AWARE                          uint32 = 0x8000
	SDC_VIRTUAL_REFRESH_RATE_AWARE                  uint32 = 0x20000
	RDH_RECTANGLES                                  uint32 = 0x1
	SYSRGN                                          uint32 = 0x4
	TT_POLYGON_TYPE                                 uint32 = 0x18
	TT_PRIM_LINE                                    uint32 = 0x1
	TT_PRIM_QSPLINE                                 uint32 = 0x2
	TT_PRIM_CSPLINE                                 uint32 = 0x3
	GCP_DBCS                                        uint32 = 0x1
	GCP_ERROR                                       uint32 = 0x8000
	FLI_MASK                                        uint32 = 0x103b
	FLI_GLYPHS                                      int32  = 262144
	GCP_JUSTIFYIN                                   int32  = 2097152
	GCPCLASS_LATIN                                  uint32 = 0x1
	GCPCLASS_HEBREW                                 uint32 = 0x2
	GCPCLASS_ARABIC                                 uint32 = 0x2
	GCPCLASS_NEUTRAL                                uint32 = 0x3
	GCPCLASS_LOCALNUMBER                            uint32 = 0x4
	GCPCLASS_LATINNUMBER                            uint32 = 0x5
	GCPCLASS_LATINNUMERICTERMINATOR                 uint32 = 0x6
	GCPCLASS_LATINNUMERICSEPARATOR                  uint32 = 0x7
	GCPCLASS_NUMERICSEPARATOR                       uint32 = 0x8
	GCPCLASS_PREBOUNDLTR                            uint32 = 0x80
	GCPCLASS_PREBOUNDRTL                            uint32 = 0x40
	GCPCLASS_POSTBOUNDLTR                           uint32 = 0x20
	GCPCLASS_POSTBOUNDRTL                           uint32 = 0x10
	GCPGLYPH_LINKBEFORE                             uint32 = 0x8000
	GCPGLYPH_LINKAFTER                              uint32 = 0x4000
	TT_AVAILABLE                                    uint32 = 0x1
	TT_ENABLED                                      uint32 = 0x2
	DC_BINADJUST                                    uint32 = 0x13
	DC_EMF_COMPLIANT                                uint32 = 0x14
	DC_DATATYPE_PRODUCED                            uint32 = 0x15
	DC_MANUFACTURER                                 uint32 = 0x17
	DC_MODEL                                        uint32 = 0x18
	PRINTRATEUNIT_PPM                               uint32 = 0x1
	PRINTRATEUNIT_CPS                               uint32 = 0x2
	PRINTRATEUNIT_LPM                               uint32 = 0x3
	PRINTRATEUNIT_IPM                               uint32 = 0x4
	DCTT_BITMAP                                     int32  = 1
	DCTT_DOWNLOAD                                   int32  = 2
	DCTT_SUBDEV                                     int32  = 4
	DCTT_DOWNLOAD_OUTLINE                           int32  = 8
	DCBA_FACEUPNONE                                 uint32 = 0x0
	DCBA_FACEUPCENTER                               uint32 = 0x1
	DCBA_FACEUPLEFT                                 uint32 = 0x2
	DCBA_FACEUPRIGHT                                uint32 = 0x3
	DCBA_FACEDOWNNONE                               uint32 = 0x100
	DCBA_FACEDOWNCENTER                             uint32 = 0x101
	DCBA_FACEDOWNLEFT                               uint32 = 0x102
	DCBA_FACEDOWNRIGHT                              uint32 = 0x103
	GS_8BIT_INDICES                                 uint32 = 0x1
	GGI_MARK_NONEXISTING_GLYPHS                     uint32 = 0x1
	MM_MAX_NUMAXES                                  uint32 = 0x10
	MM_MAX_AXES_NAMELEN                             uint32 = 0x10
	GDIREGISTERDDRAWPACKETVERSION                   uint32 = 0x1
	AC_SRC_OVER                                     uint32 = 0x0
	AC_SRC_ALPHA                                    uint32 = 0x1
	GRADIENT_FILL_OP_FLAG                           uint32 = 0xff
	CA_NEGATIVE                                     uint32 = 0x1
	CA_LOG_FILTER                                   uint32 = 0x2
	ILLUMINANT_DEVICE_DEFAULT                       uint32 = 0x0
	ILLUMINANT_A                                    uint32 = 0x1
	ILLUMINANT_B                                    uint32 = 0x2
	ILLUMINANT_C                                    uint32 = 0x3
	ILLUMINANT_D50                                  uint32 = 0x4
	ILLUMINANT_D55                                  uint32 = 0x5
	ILLUMINANT_D65                                  uint32 = 0x6
	ILLUMINANT_D75                                  uint32 = 0x7
	ILLUMINANT_F2                                   uint32 = 0x8
	ILLUMINANT_MAX_INDEX                            uint32 = 0x8
	ILLUMINANT_TUNGSTEN                             uint32 = 0x1
	ILLUMINANT_DAYLIGHT                             uint32 = 0x3
	ILLUMINANT_FLUORESCENT                          uint32 = 0x8
	ILLUMINANT_NTSC                                 uint32 = 0x3
	DI_APPBANDING                                   uint32 = 0x1
	DI_ROPS_READ_DESTINATION                        uint32 = 0x2
	FONTMAPPER_MAX                                  uint32 = 0xa
	ENHMETA_SIGNATURE                               uint32 = 0x464d4520
	ENHMETA_STOCK_OBJECT                            uint32 = 0x80000000
	EMR_HEADER                                      uint32 = 0x1
	EMR_POLYBEZIER                                  uint32 = 0x2
	EMR_POLYGON                                     uint32 = 0x3
	EMR_POLYLINE                                    uint32 = 0x4
	EMR_POLYBEZIERTO                                uint32 = 0x5
	EMR_POLYLINETO                                  uint32 = 0x6
	EMR_POLYPOLYLINE                                uint32 = 0x7
	EMR_POLYPOLYGON                                 uint32 = 0x8
	EMR_SETWINDOWEXTEX                              uint32 = 0x9
	EMR_SETWINDOWORGEX                              uint32 = 0xa
	EMR_SETVIEWPORTEXTEX                            uint32 = 0xb
	EMR_SETVIEWPORTORGEX                            uint32 = 0xc
	EMR_SETBRUSHORGEX                               uint32 = 0xd
	EMR_EOF                                         uint32 = 0xe
	EMR_SETPIXELV                                   uint32 = 0xf
	EMR_SETMAPPERFLAGS                              uint32 = 0x10
	EMR_SETMAPMODE                                  uint32 = 0x11
	EMR_SETBKMODE                                   uint32 = 0x12
	EMR_SETPOLYFILLMODE                             uint32 = 0x13
	EMR_SETROP2                                     uint32 = 0x14
	EMR_SETSTRETCHBLTMODE                           uint32 = 0x15
	EMR_SETTEXTALIGN                                uint32 = 0x16
	EMR_SETCOLORADJUSTMENT                          uint32 = 0x17
	EMR_SETTEXTCOLOR                                uint32 = 0x18
	EMR_SETBKCOLOR                                  uint32 = 0x19
	EMR_OFFSETCLIPRGN                               uint32 = 0x1a
	EMR_MOVETOEX                                    uint32 = 0x1b
	EMR_SETMETARGN                                  uint32 = 0x1c
	EMR_EXCLUDECLIPRECT                             uint32 = 0x1d
	EMR_INTERSECTCLIPRECT                           uint32 = 0x1e
	EMR_SCALEVIEWPORTEXTEX                          uint32 = 0x1f
	EMR_SCALEWINDOWEXTEX                            uint32 = 0x20
	EMR_SAVEDC                                      uint32 = 0x21
	EMR_RESTOREDC                                   uint32 = 0x22
	EMR_SETWORLDTRANSFORM                           uint32 = 0x23
	EMR_MODIFYWORLDTRANSFORM                        uint32 = 0x24
	EMR_SELECTOBJECT                                uint32 = 0x25
	EMR_CREATEPEN                                   uint32 = 0x26
	EMR_CREATEBRUSHINDIRECT                         uint32 = 0x27
	EMR_DELETEOBJECT                                uint32 = 0x28
	EMR_ANGLEARC                                    uint32 = 0x29
	EMR_ELLIPSE                                     uint32 = 0x2a
	EMR_RECTANGLE                                   uint32 = 0x2b
	EMR_ROUNDRECT                                   uint32 = 0x2c
	EMR_ARC                                         uint32 = 0x2d
	EMR_CHORD                                       uint32 = 0x2e
	EMR_PIE                                         uint32 = 0x2f
	EMR_SELECTPALETTE                               uint32 = 0x30
	EMR_CREATEPALETTE                               uint32 = 0x31
	EMR_SETPALETTEENTRIES                           uint32 = 0x32
	EMR_RESIZEPALETTE                               uint32 = 0x33
	EMR_REALIZEPALETTE                              uint32 = 0x34
	EMR_EXTFLOODFILL                                uint32 = 0x35
	EMR_LINETO                                      uint32 = 0x36
	EMR_ARCTO                                       uint32 = 0x37
	EMR_POLYDRAW                                    uint32 = 0x38
	EMR_SETARCDIRECTION                             uint32 = 0x39
	EMR_SETMITERLIMIT                               uint32 = 0x3a
	EMR_BEGINPATH                                   uint32 = 0x3b
	EMR_ENDPATH                                     uint32 = 0x3c
	EMR_CLOSEFIGURE                                 uint32 = 0x3d
	EMR_FILLPATH                                    uint32 = 0x3e
	EMR_STROKEANDFILLPATH                           uint32 = 0x3f
	EMR_STROKEPATH                                  uint32 = 0x40
	EMR_FLATTENPATH                                 uint32 = 0x41
	EMR_WIDENPATH                                   uint32 = 0x42
	EMR_SELECTCLIPPATH                              uint32 = 0x43
	EMR_ABORTPATH                                   uint32 = 0x44
	EMR_GDICOMMENT                                  uint32 = 0x46
	EMR_FILLRGN                                     uint32 = 0x47
	EMR_FRAMERGN                                    uint32 = 0x48
	EMR_INVERTRGN                                   uint32 = 0x49
	EMR_PAINTRGN                                    uint32 = 0x4a
	EMR_EXTSELECTCLIPRGN                            uint32 = 0x4b
	EMR_BITBLT                                      uint32 = 0x4c
	EMR_STRETCHBLT                                  uint32 = 0x4d
	EMR_MASKBLT                                     uint32 = 0x4e
	EMR_PLGBLT                                      uint32 = 0x4f
	EMR_SETDIBITSTODEVICE                           uint32 = 0x50
	EMR_STRETCHDIBITS                               uint32 = 0x51
	EMR_EXTCREATEFONTINDIRECTW                      uint32 = 0x52
	EMR_EXTTEXTOUTA                                 uint32 = 0x53
	EMR_EXTTEXTOUTW                                 uint32 = 0x54
	EMR_POLYBEZIER16                                uint32 = 0x55
	EMR_POLYGON16                                   uint32 = 0x56
	EMR_POLYLINE16                                  uint32 = 0x57
	EMR_POLYBEZIERTO16                              uint32 = 0x58
	EMR_POLYLINETO16                                uint32 = 0x59
	EMR_POLYPOLYLINE16                              uint32 = 0x5a
	EMR_POLYPOLYGON16                               uint32 = 0x5b
	EMR_POLYDRAW16                                  uint32 = 0x5c
	EMR_CREATEMONOBRUSH                             uint32 = 0x5d
	EMR_CREATEDIBPATTERNBRUSHPT                     uint32 = 0x5e
	EMR_EXTCREATEPEN                                uint32 = 0x5f
	EMR_POLYTEXTOUTA                                uint32 = 0x60
	EMR_POLYTEXTOUTW                                uint32 = 0x61
	EMR_SETICMMODE                                  uint32 = 0x62
	EMR_CREATECOLORSPACE                            uint32 = 0x63
	EMR_SETCOLORSPACE                               uint32 = 0x64
	EMR_DELETECOLORSPACE                            uint32 = 0x65
	EMR_GLSRECORD                                   uint32 = 0x66
	EMR_GLSBOUNDEDRECORD                            uint32 = 0x67
	EMR_PIXELFORMAT                                 uint32 = 0x68
	EMR_RESERVED_105                                uint32 = 0x69
	EMR_RESERVED_106                                uint32 = 0x6a
	EMR_RESERVED_107                                uint32 = 0x6b
	EMR_RESERVED_108                                uint32 = 0x6c
	EMR_RESERVED_109                                uint32 = 0x6d
	EMR_RESERVED_110                                uint32 = 0x6e
	EMR_COLORCORRECTPALETTE                         uint32 = 0x6f
	EMR_SETICMPROFILEA                              uint32 = 0x70
	EMR_SETICMPROFILEW                              uint32 = 0x71
	EMR_ALPHABLEND                                  uint32 = 0x72
	EMR_SETLAYOUT                                   uint32 = 0x73
	EMR_TRANSPARENTBLT                              uint32 = 0x74
	EMR_RESERVED_117                                uint32 = 0x75
	EMR_GRADIENTFILL                                uint32 = 0x76
	EMR_RESERVED_119                                uint32 = 0x77
	EMR_RESERVED_120                                uint32 = 0x78
	EMR_COLORMATCHTOTARGETW                         uint32 = 0x79
	EMR_CREATECOLORSPACEW                           uint32 = 0x7a
	EMR_MIN                                         uint32 = 0x1
	EMR_MAX                                         uint32 = 0x7a
	SETICMPROFILE_EMBEDED                           uint32 = 0x1
	CREATECOLORSPACE_EMBEDED                        uint32 = 0x1
	COLORMATCHTOTARGET_EMBEDED                      uint32 = 0x1
	GDICOMMENT_IDENTIFIER                           uint32 = 0x43494447
	GDICOMMENT_WINDOWS_METAFILE                     uint32 = 0x80000001
	GDICOMMENT_BEGINGROUP                           uint32 = 0x2
	GDICOMMENT_ENDGROUP                             uint32 = 0x3
	GDICOMMENT_MULTIFORMATS                         uint32 = 0x40000004
	EPS_SIGNATURE                                   uint32 = 0x46535045
	GDICOMMENT_UNICODE_STRING                       uint32 = 0x40
	GDICOMMENT_UNICODE_END                          uint32 = 0x80
	WGL_FONT_LINES                                  uint32 = 0x0
	WGL_FONT_POLYGONS                               uint32 = 0x1
	LPD_DOUBLEBUFFER                                uint32 = 0x1
	LPD_STEREO                                      uint32 = 0x2
	LPD_SUPPORT_GDI                                 uint32 = 0x10
	LPD_SUPPORT_OPENGL                              uint32 = 0x20
	LPD_SHARE_DEPTH                                 uint32 = 0x40
	LPD_SHARE_STENCIL                               uint32 = 0x80
	LPD_SHARE_ACCUM                                 uint32 = 0x100
	LPD_SWAP_EXCHANGE                               uint32 = 0x200
	LPD_SWAP_COPY                                   uint32 = 0x400
	LPD_TRANSPARENT                                 uint32 = 0x1000
	LPD_TYPE_RGBA                                   uint32 = 0x0
	LPD_TYPE_COLORINDEX                             uint32 = 0x1
	WGL_SWAP_MAIN_PLANE                             uint32 = 0x1
	WGL_SWAP_OVERLAY1                               uint32 = 0x2
	WGL_SWAP_OVERLAY2                               uint32 = 0x4
	WGL_SWAP_OVERLAY3                               uint32 = 0x8
	WGL_SWAP_OVERLAY4                               uint32 = 0x10
	WGL_SWAP_OVERLAY5                               uint32 = 0x20
	WGL_SWAP_OVERLAY6                               uint32 = 0x40
	WGL_SWAP_OVERLAY7                               uint32 = 0x80
	WGL_SWAP_OVERLAY8                               uint32 = 0x100
	WGL_SWAP_OVERLAY9                               uint32 = 0x200
	WGL_SWAP_OVERLAY10                              uint32 = 0x400
	WGL_SWAP_OVERLAY11                              uint32 = 0x800
	WGL_SWAP_OVERLAY12                              uint32 = 0x1000
	WGL_SWAP_OVERLAY13                              uint32 = 0x2000
	WGL_SWAP_OVERLAY14                              uint32 = 0x4000
	WGL_SWAP_OVERLAY15                              uint32 = 0x8000
	WGL_SWAP_UNDERLAY1                              uint32 = 0x10000
	WGL_SWAP_UNDERLAY2                              uint32 = 0x20000
	WGL_SWAP_UNDERLAY3                              uint32 = 0x40000
	WGL_SWAP_UNDERLAY4                              uint32 = 0x80000
	WGL_SWAP_UNDERLAY5                              uint32 = 0x100000
	WGL_SWAP_UNDERLAY6                              uint32 = 0x200000
	WGL_SWAP_UNDERLAY7                              uint32 = 0x400000
	WGL_SWAP_UNDERLAY8                              uint32 = 0x800000
	WGL_SWAP_UNDERLAY9                              uint32 = 0x1000000
	WGL_SWAP_UNDERLAY10                             uint32 = 0x2000000
	WGL_SWAP_UNDERLAY11                             uint32 = 0x4000000
	WGL_SWAP_UNDERLAY12                             uint32 = 0x8000000
	WGL_SWAP_UNDERLAY13                             uint32 = 0x10000000
	WGL_SWAP_UNDERLAY14                             uint32 = 0x20000000
	WGL_SWAP_UNDERLAY15                             uint32 = 0x40000000
	WGL_SWAPMULTIPLE_MAX                            uint32 = 0x10
	NEWTRANSPARENT                                  uint32 = 0x3
	QUERYROPSUPPORT                                 uint32 = 0x28
	SELECTDIB                                       uint32 = 0x29
	SC_SCREENSAVE                                   uint32 = 0xf140
	TTFCFP_SUBSET                                   uint32 = 0x0
	TTFCFP_SUBSET1                                  uint32 = 0x1
	TTFCFP_DELTA                                    uint32 = 0x2
	TTFCFP_APPLE_PLATFORMID                         uint32 = 0x1
	TTFCFP_MS_PLATFORMID                            uint32 = 0x3
	TTFCFP_DONT_CARE                                uint32 = 0xffff
	TTFCFP_LANG_KEEP_ALL                            uint32 = 0x0
	TTFCFP_FLAGS_SUBSET                             uint32 = 0x1
	TTFCFP_FLAGS_COMPRESS                           uint32 = 0x2
	TTFCFP_FLAGS_TTC                                uint32 = 0x4
	TTFCFP_FLAGS_GLYPHLIST                          uint32 = 0x8
	TTFMFP_SUBSET                                   uint32 = 0x0
	TTFMFP_SUBSET1                                  uint32 = 0x1
	TTFMFP_DELTA                                    uint32 = 0x2
	ERR_GENERIC                                     uint32 = 0x3e8
	ERR_READOUTOFBOUNDS                             uint32 = 0x3e9
	ERR_WRITEOUTOFBOUNDS                            uint32 = 0x3ea
	ERR_READCONTROL                                 uint32 = 0x3eb
	ERR_WRITECONTROL                                uint32 = 0x3ec
	ERR_MEM                                         uint32 = 0x3ed
	ERR_FORMAT                                      uint32 = 0x3ee
	ERR_WOULD_GROW                                  uint32 = 0x3ef
	ERR_VERSION                                     uint32 = 0x3f0
	ERR_NO_GLYPHS                                   uint32 = 0x3f1
	ERR_INVALID_MERGE_FORMATS                       uint32 = 0x3f2
	ERR_INVALID_MERGE_CHECKSUMS                     uint32 = 0x3f3
	ERR_INVALID_MERGE_NUMGLYPHS                     uint32 = 0x3f4
	ERR_INVALID_DELTA_FORMAT                        uint32 = 0x3f5
	ERR_NOT_TTC                                     uint32 = 0x3f6
	ERR_INVALID_TTC_INDEX                           uint32 = 0x3f7
	ERR_MISSING_CMAP                                uint32 = 0x406
	ERR_MISSING_GLYF                                uint32 = 0x407
	ERR_MISSING_HEAD                                uint32 = 0x408
	ERR_MISSING_HHEA                                uint32 = 0x409
	ERR_MISSING_HMTX                                uint32 = 0x40a
	ERR_MISSING_LOCA                                uint32 = 0x40b
	ERR_MISSING_MAXP                                uint32 = 0x40c
	ERR_MISSING_NAME                                uint32 = 0x40d
	ERR_MISSING_POST                                uint32 = 0x40e
	ERR_MISSING_OS2                                 uint32 = 0x40f
	ERR_MISSING_VHEA                                uint32 = 0x410
	ERR_MISSING_VMTX                                uint32 = 0x411
	ERR_MISSING_HHEA_OR_VHEA                        uint32 = 0x412
	ERR_MISSING_HMTX_OR_VMTX                        uint32 = 0x413
	ERR_MISSING_EBDT                                uint32 = 0x414
	ERR_INVALID_CMAP                                uint32 = 0x424
	ERR_INVALID_GLYF                                uint32 = 0x425
	ERR_INVALID_HEAD                                uint32 = 0x426
	ERR_INVALID_HHEA                                uint32 = 0x427
	ERR_INVALID_HMTX                                uint32 = 0x428
	ERR_INVALID_LOCA                                uint32 = 0x429
	ERR_INVALID_MAXP                                uint32 = 0x42a
	ERR_INVALID_NAME                                uint32 = 0x42b
	ERR_INVALID_POST                                uint32 = 0x42c
	ERR_INVALID_OS2                                 uint32 = 0x42d
	ERR_INVALID_VHEA                                uint32 = 0x42e
	ERR_INVALID_VMTX                                uint32 = 0x42f
	ERR_INVALID_HHEA_OR_VHEA                        uint32 = 0x430
	ERR_INVALID_HMTX_OR_VMTX                        uint32 = 0x431
	ERR_INVALID_TTO                                 uint32 = 0x438
	ERR_INVALID_GSUB                                uint32 = 0x439
	ERR_INVALID_GPOS                                uint32 = 0x43a
	ERR_INVALID_GDEF                                uint32 = 0x43b
	ERR_INVALID_JSTF                                uint32 = 0x43c
	ERR_INVALID_BASE                                uint32 = 0x43d
	ERR_INVALID_EBLC                                uint32 = 0x43e
	ERR_INVALID_LTSH                                uint32 = 0x43f
	ERR_INVALID_VDMX                                uint32 = 0x440
	ERR_INVALID_HDMX                                uint32 = 0x441
	ERR_PARAMETER0                                  uint32 = 0x44c
	ERR_PARAMETER1                                  uint32 = 0x44d
	ERR_PARAMETER2                                  uint32 = 0x44e
	ERR_PARAMETER3                                  uint32 = 0x44f
	ERR_PARAMETER4                                  uint32 = 0x450
	ERR_PARAMETER5                                  uint32 = 0x451
	ERR_PARAMETER6                                  uint32 = 0x452
	ERR_PARAMETER7                                  uint32 = 0x453
	ERR_PARAMETER8                                  uint32 = 0x454
	ERR_PARAMETER9                                  uint32 = 0x455
	ERR_PARAMETER10                                 uint32 = 0x456
	ERR_PARAMETER11                                 uint32 = 0x457
	ERR_PARAMETER12                                 uint32 = 0x458
	ERR_PARAMETER13                                 uint32 = 0x459
	ERR_PARAMETER14                                 uint32 = 0x45a
	ERR_PARAMETER15                                 uint32 = 0x45b
	ERR_PARAMETER16                                 uint32 = 0x45c
	CHARSET_DEFAULT                                 uint32 = 0x1
	CHARSET_GLYPHIDX                                uint32 = 0x3
	TTEMBED_FAILIFVARIATIONSIMULATED                uint32 = 0x10
	TTEMBED_WEBOBJECT                               uint32 = 0x80
	TTEMBED_XORENCRYPTDATA                          uint32 = 0x10000000
	TTEMBED_VARIATIONSIMULATED                      uint32 = 0x1
	TTEMBED_EUDCEMBEDDED                            uint32 = 0x2
	TTEMBED_SUBSETCANCEL                            uint32 = 0x4
	TTLOAD_PRIVATE                                  uint32 = 0x1
	TTLOAD_EUDC_OVERWRITE                           uint32 = 0x2
	TTLOAD_EUDC_SET                                 uint32 = 0x4
	TTDELETE_DONTREMOVEFONT                         uint32 = 0x1
	E_NONE                                          int32  = 0
	E_API_NOTIMPL                                   int32  = 1
	E_CHARCODECOUNTINVALID                          int32  = 2
	E_CHARCODESETINVALID                            int32  = 3
	E_DEVICETRUETYPEFONT                            int32  = 4
	E_HDCINVALID                                    int32  = 6
	E_NOFREEMEMORY                                  int32  = 7
	E_FONTREFERENCEINVALID                          int32  = 8
	E_NOTATRUETYPEFONT                              int32  = 10
	E_ERRORACCESSINGFONTDATA                        int32  = 12
	E_ERRORACCESSINGFACENAME                        int32  = 13
	E_ERRORUNICODECONVERSION                        int32  = 17
	E_ERRORCONVERTINGCHARS                          int32  = 18
	E_EXCEPTION                                     int32  = 19
	E_RESERVEDPARAMNOTNULL                          int32  = 20
	E_CHARSETINVALID                                int32  = 21
	E_FILE_NOT_FOUND                                int32  = 23
	E_TTC_INDEX_OUT_OF_RANGE                        int32  = 24
	E_INPUTPARAMINVALID                             int32  = 25
	E_ERRORCOMPRESSINGFONTDATA                      int32  = 256
	E_FONTDATAINVALID                               int32  = 258
	E_NAMECHANGEFAILED                              int32  = 259
	E_FONTNOTEMBEDDABLE                             int32  = 260
	E_PRIVSINVALID                                  int32  = 261
	E_SUBSETTINGFAILED                              int32  = 262
	E_READFROMSTREAMFAILED                          int32  = 263
	E_SAVETOSTREAMFAILED                            int32  = 264
	E_NOOS2                                         int32  = 265
	E_T2NOFREEMEMORY                                int32  = 266
	E_ERRORREADINGFONTDATA                          int32  = 267
	E_FLAGSINVALID                                  int32  = 268
	E_ERRORCREATINGFONTFILE                         int32  = 269
	E_FONTALREADYEXISTS                             int32  = 270
	E_FONTNAMEALREADYEXISTS                         int32  = 271
	E_FONTINSTALLFAILED                             int32  = 272
	E_ERRORDECOMPRESSINGFONTDATA                    int32  = 273
	E_ERRORACCESSINGEXCLUDELIST                     int32  = 274
	E_FACENAMEINVALID                               int32  = 275
	E_STREAMINVALID                                 int32  = 276
	E_STATUSINVALID                                 int32  = 277
	E_PRIVSTATUSINVALID                             int32  = 278
	E_PERMISSIONSINVALID                            int32  = 279
	E_PBENABLEDINVALID                              int32  = 280
	E_SUBSETTINGEXCEPTION                           int32  = 281
	E_SUBSTRING_TEST_FAIL                           int32  = 282
	E_FONTVARIATIONSIMULATED                        int32  = 283
	E_FONTFAMILYNAMENOTINFULL                       int32  = 285
	E_ADDFONTFAILED                                 int32  = 512
	E_COULDNTCREATETEMPFILE                         int32  = 513
	E_FONTFILECREATEFAILED                          int32  = 515
	E_WINDOWSAPI                                    int32  = 516
	E_FONTFILENOTFOUND                              int32  = 517
	E_RESOURCEFILECREATEFAILED                      int32  = 518
	E_ERROREXPANDINGFONTDATA                        int32  = 519
	E_ERRORGETTINGDC                                int32  = 520
	E_EXCEPTIONINDECOMPRESSION                      int32  = 521
	E_EXCEPTIONINCOMPRESSION                        int32  = 522
)

// enums

// enum
type R2_MODE int32

const (
	R2_BLACK       R2_MODE = 1
	R2_NOTMERGEPEN R2_MODE = 2
	R2_MASKNOTPEN  R2_MODE = 3
	R2_NOTCOPYPEN  R2_MODE = 4
	R2_MASKPENNOT  R2_MODE = 5
	R2_NOT         R2_MODE = 6
	R2_XORPEN      R2_MODE = 7
	R2_NOTMASKPEN  R2_MODE = 8
	R2_MASKPEN     R2_MODE = 9
	R2_NOTXORPEN   R2_MODE = 10
	R2_NOP         R2_MODE = 11
	R2_MERGENOTPEN R2_MODE = 12
	R2_COPYPEN     R2_MODE = 13
	R2_MERGEPENNOT R2_MODE = 14
	R2_MERGEPEN    R2_MODE = 15
	R2_WHITE       R2_MODE = 16
	R2_LAST        R2_MODE = 16
)

// enum
type RGN_COMBINE_MODE int32

const (
	RGN_AND  RGN_COMBINE_MODE = 1
	RGN_OR   RGN_COMBINE_MODE = 2
	RGN_XOR  RGN_COMBINE_MODE = 3
	RGN_DIFF RGN_COMBINE_MODE = 4
	RGN_COPY RGN_COMBINE_MODE = 5
	RGN_MIN  RGN_COMBINE_MODE = 1
	RGN_MAX  RGN_COMBINE_MODE = 5
)

// enum
// flags
type ETO_OPTIONS uint32

const (
	ETO_OPAQUE            ETO_OPTIONS = 2
	ETO_CLIPPED           ETO_OPTIONS = 4
	ETO_GLYPH_INDEX       ETO_OPTIONS = 16
	ETO_RTLREADING        ETO_OPTIONS = 128
	ETO_NUMERICSLOCAL     ETO_OPTIONS = 1024
	ETO_NUMERICSLATIN     ETO_OPTIONS = 2048
	ETO_IGNORELANGUAGE    ETO_OPTIONS = 4096
	ETO_PDY               ETO_OPTIONS = 8192
	ETO_REVERSE_INDEX_MAP ETO_OPTIONS = 65536
)

// enum
type OBJ_TYPE int32

const (
	OBJ_PEN         OBJ_TYPE = 1
	OBJ_BRUSH       OBJ_TYPE = 2
	OBJ_DC          OBJ_TYPE = 3
	OBJ_METADC      OBJ_TYPE = 4
	OBJ_PAL         OBJ_TYPE = 5
	OBJ_FONT        OBJ_TYPE = 6
	OBJ_BITMAP      OBJ_TYPE = 7
	OBJ_REGION      OBJ_TYPE = 8
	OBJ_METAFILE    OBJ_TYPE = 9
	OBJ_MEMDC       OBJ_TYPE = 10
	OBJ_EXTPEN      OBJ_TYPE = 11
	OBJ_ENHMETADC   OBJ_TYPE = 12
	OBJ_ENHMETAFILE OBJ_TYPE = 13
	OBJ_COLORSPACE  OBJ_TYPE = 14
)

// enum
type ROP_CODE uint32

const (
	SRCCOPY        ROP_CODE = 13369376
	SRCPAINT       ROP_CODE = 15597702
	SRCAND         ROP_CODE = 8913094
	SRCINVERT      ROP_CODE = 6684742
	SRCERASE       ROP_CODE = 4457256
	NOTSRCCOPY     ROP_CODE = 3342344
	NOTSRCERASE    ROP_CODE = 1114278
	MERGECOPY      ROP_CODE = 12583114
	MERGEPAINT     ROP_CODE = 12255782
	PATCOPY        ROP_CODE = 15728673
	PATPAINT       ROP_CODE = 16452105
	PATINVERT      ROP_CODE = 5898313
	DSTINVERT      ROP_CODE = 5570569
	BLACKNESS      ROP_CODE = 66
	WHITENESS      ROP_CODE = 16711778
	NOMIRRORBITMAP ROP_CODE = 2147483648
	CAPTUREBLT     ROP_CODE = 1073741824
)

// enum
type DIB_USAGE uint32

const (
	DIB_RGB_COLORS DIB_USAGE = 0
	DIB_PAL_COLORS DIB_USAGE = 1
)

// enum
// flags
type DRAWEDGE_FLAGS uint32

const (
	BDR_RAISEDOUTER DRAWEDGE_FLAGS = 1
	BDR_SUNKENOUTER DRAWEDGE_FLAGS = 2
	BDR_RAISEDINNER DRAWEDGE_FLAGS = 4
	BDR_SUNKENINNER DRAWEDGE_FLAGS = 8
	BDR_OUTER       DRAWEDGE_FLAGS = 3
	BDR_INNER       DRAWEDGE_FLAGS = 12
	BDR_RAISED      DRAWEDGE_FLAGS = 5
	BDR_SUNKEN      DRAWEDGE_FLAGS = 10
	EDGE_RAISED     DRAWEDGE_FLAGS = 5
	EDGE_SUNKEN     DRAWEDGE_FLAGS = 10
	EDGE_ETCHED     DRAWEDGE_FLAGS = 6
	EDGE_BUMP       DRAWEDGE_FLAGS = 9
)

// enum
type DFC_TYPE uint32

const (
	DFC_CAPTION   DFC_TYPE = 1
	DFC_MENU      DFC_TYPE = 2
	DFC_SCROLL    DFC_TYPE = 3
	DFC_BUTTON    DFC_TYPE = 4
	DFC_POPUPMENU DFC_TYPE = 5
)

// enum
// flags
type DFCS_STATE uint32

const (
	DFCS_CAPTIONCLOSE        DFCS_STATE = 0
	DFCS_CAPTIONMIN          DFCS_STATE = 1
	DFCS_CAPTIONMAX          DFCS_STATE = 2
	DFCS_CAPTIONRESTORE      DFCS_STATE = 3
	DFCS_CAPTIONHELP         DFCS_STATE = 4
	DFCS_MENUARROW           DFCS_STATE = 0
	DFCS_MENUCHECK           DFCS_STATE = 1
	DFCS_MENUBULLET          DFCS_STATE = 2
	DFCS_MENUARROWRIGHT      DFCS_STATE = 4
	DFCS_SCROLLUP            DFCS_STATE = 0
	DFCS_SCROLLDOWN          DFCS_STATE = 1
	DFCS_SCROLLLEFT          DFCS_STATE = 2
	DFCS_SCROLLRIGHT         DFCS_STATE = 3
	DFCS_SCROLLCOMBOBOX      DFCS_STATE = 5
	DFCS_SCROLLSIZEGRIP      DFCS_STATE = 8
	DFCS_SCROLLSIZEGRIPRIGHT DFCS_STATE = 16
	DFCS_BUTTONCHECK         DFCS_STATE = 0
	DFCS_BUTTONRADIOIMAGE    DFCS_STATE = 1
	DFCS_BUTTONRADIOMASK     DFCS_STATE = 2
	DFCS_BUTTONRADIO         DFCS_STATE = 4
	DFCS_BUTTON3STATE        DFCS_STATE = 8
	DFCS_BUTTONPUSH          DFCS_STATE = 16
	DFCS_INACTIVE            DFCS_STATE = 256
	DFCS_PUSHED              DFCS_STATE = 512
	DFCS_CHECKED             DFCS_STATE = 1024
	DFCS_TRANSPARENT         DFCS_STATE = 2048
	DFCS_HOT                 DFCS_STATE = 4096
	DFCS_ADJUSTRECT          DFCS_STATE = 8192
	DFCS_FLAT                DFCS_STATE = 16384
	DFCS_MONO                DFCS_STATE = 32768
)

// enum
// flags
type CDS_TYPE uint32

const (
	CDS_FULLSCREEN           CDS_TYPE = 4
	CDS_GLOBAL               CDS_TYPE = 8
	CDS_NORESET              CDS_TYPE = 268435456
	CDS_RESET                CDS_TYPE = 1073741824
	CDS_SET_PRIMARY          CDS_TYPE = 16
	CDS_TEST                 CDS_TYPE = 2
	CDS_UPDATEREGISTRY       CDS_TYPE = 1
	CDS_VIDEOPARAMETERS      CDS_TYPE = 32
	CDS_ENABLE_UNSAFE_MODES  CDS_TYPE = 256
	CDS_DISABLE_UNSAFE_MODES CDS_TYPE = 512
	CDS_RESET_EX             CDS_TYPE = 536870912
)

// enum
type DISP_CHANGE int32

const (
	DISP_CHANGE_SUCCESSFUL  DISP_CHANGE = 0
	DISP_CHANGE_RESTART     DISP_CHANGE = 1
	DISP_CHANGE_FAILED      DISP_CHANGE = -1
	DISP_CHANGE_BADMODE     DISP_CHANGE = -2
	DISP_CHANGE_NOTUPDATED  DISP_CHANGE = -3
	DISP_CHANGE_BADFLAGS    DISP_CHANGE = -4
	DISP_CHANGE_BADPARAM    DISP_CHANGE = -5
	DISP_CHANGE_BADDUALVIEW DISP_CHANGE = -6
)

// enum
// flags
type DRAWSTATE_FLAGS uint32

const (
	DST_COMPLEX    DRAWSTATE_FLAGS = 0
	DST_TEXT       DRAWSTATE_FLAGS = 1
	DST_PREFIXTEXT DRAWSTATE_FLAGS = 2
	DST_ICON       DRAWSTATE_FLAGS = 3
	DST_BITMAP     DRAWSTATE_FLAGS = 4
	DSS_NORMAL     DRAWSTATE_FLAGS = 0
	DSS_UNION      DRAWSTATE_FLAGS = 16
	DSS_DISABLED   DRAWSTATE_FLAGS = 32
	DSS_MONO       DRAWSTATE_FLAGS = 128
	DSS_HIDEPREFIX DRAWSTATE_FLAGS = 512
	DSS_PREFIXONLY DRAWSTATE_FLAGS = 1024
	DSS_RIGHT      DRAWSTATE_FLAGS = 32768
)

// enum
// flags
type REDRAW_WINDOW_FLAGS uint32

const (
	RDW_INVALIDATE      REDRAW_WINDOW_FLAGS = 1
	RDW_INTERNALPAINT   REDRAW_WINDOW_FLAGS = 2
	RDW_ERASE           REDRAW_WINDOW_FLAGS = 4
	RDW_VALIDATE        REDRAW_WINDOW_FLAGS = 8
	RDW_NOINTERNALPAINT REDRAW_WINDOW_FLAGS = 16
	RDW_NOERASE         REDRAW_WINDOW_FLAGS = 32
	RDW_NOCHILDREN      REDRAW_WINDOW_FLAGS = 64
	RDW_ALLCHILDREN     REDRAW_WINDOW_FLAGS = 128
	RDW_UPDATENOW       REDRAW_WINDOW_FLAGS = 256
	RDW_ERASENOW        REDRAW_WINDOW_FLAGS = 512
	RDW_FRAME           REDRAW_WINDOW_FLAGS = 1024
	RDW_NOFRAME         REDRAW_WINDOW_FLAGS = 2048
)

// enum
type ENUM_DISPLAY_SETTINGS_MODE uint32

const (
	ENUM_CURRENT_SETTINGS  ENUM_DISPLAY_SETTINGS_MODE = 4294967295
	ENUM_REGISTRY_SETTINGS ENUM_DISPLAY_SETTINGS_MODE = 4294967294
)

// enum
type TEXT_ALIGN_OPTIONS uint32

const (
	TA_NOUPDATECP TEXT_ALIGN_OPTIONS = 0
	TA_UPDATECP   TEXT_ALIGN_OPTIONS = 1
	TA_LEFT       TEXT_ALIGN_OPTIONS = 0
	TA_RIGHT      TEXT_ALIGN_OPTIONS = 2
	TA_CENTER     TEXT_ALIGN_OPTIONS = 6
	TA_TOP        TEXT_ALIGN_OPTIONS = 0
	TA_BOTTOM     TEXT_ALIGN_OPTIONS = 8
	TA_BASELINE   TEXT_ALIGN_OPTIONS = 24
	TA_RTLREADING TEXT_ALIGN_OPTIONS = 256
	TA_MASK       TEXT_ALIGN_OPTIONS = 287
	VTA_BASELINE  TEXT_ALIGN_OPTIONS = 24
	VTA_LEFT      TEXT_ALIGN_OPTIONS = 8
	VTA_RIGHT     TEXT_ALIGN_OPTIONS = 0
	VTA_CENTER    TEXT_ALIGN_OPTIONS = 6
	VTA_BOTTOM    TEXT_ALIGN_OPTIONS = 2
	VTA_TOP       TEXT_ALIGN_OPTIONS = 0
)

// enum
// flags
type PEN_STYLE uint32

const (
	PS_GEOMETRIC     PEN_STYLE = 65536
	PS_COSMETIC      PEN_STYLE = 0
	PS_SOLID         PEN_STYLE = 0
	PS_DASH          PEN_STYLE = 1
	PS_DOT           PEN_STYLE = 2
	PS_DASHDOT       PEN_STYLE = 3
	PS_DASHDOTDOT    PEN_STYLE = 4
	PS_NULL          PEN_STYLE = 5
	PS_INSIDEFRAME   PEN_STYLE = 6
	PS_USERSTYLE     PEN_STYLE = 7
	PS_ALTERNATE     PEN_STYLE = 8
	PS_STYLE_MASK    PEN_STYLE = 15
	PS_ENDCAP_ROUND  PEN_STYLE = 0
	PS_ENDCAP_SQUARE PEN_STYLE = 256
	PS_ENDCAP_FLAT   PEN_STYLE = 512
	PS_ENDCAP_MASK   PEN_STYLE = 3840
	PS_JOIN_ROUND    PEN_STYLE = 0
	PS_JOIN_BEVEL    PEN_STYLE = 4096
	PS_JOIN_MITER    PEN_STYLE = 8192
	PS_JOIN_MASK     PEN_STYLE = 61440
	PS_TYPE_MASK     PEN_STYLE = 983040
)

// enum
// flags
type TTEMBED_FLAGS uint32

const (
	TTEMBED_EMBEDEUDC    TTEMBED_FLAGS = 32
	TTEMBED_RAW          TTEMBED_FLAGS = 0
	TTEMBED_SUBSET       TTEMBED_FLAGS = 1
	TTEMBED_TTCOMPRESSED TTEMBED_FLAGS = 4
)

// enum
// flags
type DRAW_TEXT_FORMAT uint32

const (
	DT_BOTTOM               DRAW_TEXT_FORMAT = 8
	DT_CALCRECT             DRAW_TEXT_FORMAT = 1024
	DT_CENTER               DRAW_TEXT_FORMAT = 1
	DT_EDITCONTROL          DRAW_TEXT_FORMAT = 8192
	DT_END_ELLIPSIS         DRAW_TEXT_FORMAT = 32768
	DT_EXPANDTABS           DRAW_TEXT_FORMAT = 64
	DT_EXTERNALLEADING      DRAW_TEXT_FORMAT = 512
	DT_HIDEPREFIX           DRAW_TEXT_FORMAT = 1048576
	DT_INTERNAL             DRAW_TEXT_FORMAT = 4096
	DT_LEFT                 DRAW_TEXT_FORMAT = 0
	DT_MODIFYSTRING         DRAW_TEXT_FORMAT = 65536
	DT_NOCLIP               DRAW_TEXT_FORMAT = 256
	DT_NOFULLWIDTHCHARBREAK DRAW_TEXT_FORMAT = 524288
	DT_NOPREFIX             DRAW_TEXT_FORMAT = 2048
	DT_PATH_ELLIPSIS        DRAW_TEXT_FORMAT = 16384
	DT_PREFIXONLY           DRAW_TEXT_FORMAT = 2097152
	DT_RIGHT                DRAW_TEXT_FORMAT = 2
	DT_RTLREADING           DRAW_TEXT_FORMAT = 131072
	DT_SINGLELINE           DRAW_TEXT_FORMAT = 32
	DT_TABSTOP              DRAW_TEXT_FORMAT = 128
	DT_TOP                  DRAW_TEXT_FORMAT = 0
	DT_VCENTER              DRAW_TEXT_FORMAT = 4
	DT_WORDBREAK            DRAW_TEXT_FORMAT = 16
	DT_WORD_ELLIPSIS        DRAW_TEXT_FORMAT = 262144
)

// enum
type EMBED_FONT_CHARSET uint32

const (
	CHARSET_UNICODE EMBED_FONT_CHARSET = 1
	CHARSET_SYMBOL  EMBED_FONT_CHARSET = 2
)

// enum
// flags
type GET_DCX_FLAGS uint32

const (
	DCX_WINDOW           GET_DCX_FLAGS = 1
	DCX_CACHE            GET_DCX_FLAGS = 2
	DCX_PARENTCLIP       GET_DCX_FLAGS = 32
	DCX_CLIPSIBLINGS     GET_DCX_FLAGS = 16
	DCX_CLIPCHILDREN     GET_DCX_FLAGS = 8
	DCX_NORESETATTRS     GET_DCX_FLAGS = 4
	DCX_LOCKWINDOWUPDATE GET_DCX_FLAGS = 1024
	DCX_EXCLUDERGN       GET_DCX_FLAGS = 64
	DCX_INTERSECTRGN     GET_DCX_FLAGS = 128
	DCX_INTERSECTUPDATE  GET_DCX_FLAGS = 512
	DCX_VALIDATE         GET_DCX_FLAGS = 2097152
)

// enum
type GET_GLYPH_OUTLINE_FORMAT uint32

const (
	GGO_BEZIER       GET_GLYPH_OUTLINE_FORMAT = 3
	GGO_BITMAP       GET_GLYPH_OUTLINE_FORMAT = 1
	GGO_GLYPH_INDEX  GET_GLYPH_OUTLINE_FORMAT = 128
	GGO_GRAY2_BITMAP GET_GLYPH_OUTLINE_FORMAT = 4
	GGO_GRAY4_BITMAP GET_GLYPH_OUTLINE_FORMAT = 5
	GGO_GRAY8_BITMAP GET_GLYPH_OUTLINE_FORMAT = 6
	GGO_METRICS      GET_GLYPH_OUTLINE_FORMAT = 0
	GGO_NATIVE       GET_GLYPH_OUTLINE_FORMAT = 2
	GGO_UNHINTED     GET_GLYPH_OUTLINE_FORMAT = 256
)

// enum
type SET_BOUNDS_RECT_FLAGS uint32

const (
	DCB_ACCUMULATE SET_BOUNDS_RECT_FLAGS = 2
	DCB_DISABLE    SET_BOUNDS_RECT_FLAGS = 8
	DCB_ENABLE     SET_BOUNDS_RECT_FLAGS = 4
	DCB_RESET      SET_BOUNDS_RECT_FLAGS = 1
)

// enum
type GET_STOCK_OBJECT_FLAGS uint32

const (
	BLACK_BRUSH         GET_STOCK_OBJECT_FLAGS = 4
	DKGRAY_BRUSH        GET_STOCK_OBJECT_FLAGS = 3
	DC_BRUSH            GET_STOCK_OBJECT_FLAGS = 18
	GRAY_BRUSH          GET_STOCK_OBJECT_FLAGS = 2
	HOLLOW_BRUSH        GET_STOCK_OBJECT_FLAGS = 5
	LTGRAY_BRUSH        GET_STOCK_OBJECT_FLAGS = 1
	NULL_BRUSH          GET_STOCK_OBJECT_FLAGS = 5
	WHITE_BRUSH         GET_STOCK_OBJECT_FLAGS = 0
	BLACK_PEN           GET_STOCK_OBJECT_FLAGS = 7
	DC_PEN              GET_STOCK_OBJECT_FLAGS = 19
	NULL_PEN            GET_STOCK_OBJECT_FLAGS = 8
	WHITE_PEN           GET_STOCK_OBJECT_FLAGS = 6
	ANSI_FIXED_FONT     GET_STOCK_OBJECT_FLAGS = 11
	ANSI_VAR_FONT       GET_STOCK_OBJECT_FLAGS = 12
	DEVICE_DEFAULT_FONT GET_STOCK_OBJECT_FLAGS = 14
	DEFAULT_GUI_FONT    GET_STOCK_OBJECT_FLAGS = 17
	OEM_FIXED_FONT      GET_STOCK_OBJECT_FLAGS = 10
	SYSTEM_FONT         GET_STOCK_OBJECT_FLAGS = 13
	SYSTEM_FIXED_FONT   GET_STOCK_OBJECT_FLAGS = 16
	DEFAULT_PALETTE     GET_STOCK_OBJECT_FLAGS = 15
)

// enum
type MODIFY_WORLD_TRANSFORM_MODE uint32

const (
	MWT_IDENTITY      MODIFY_WORLD_TRANSFORM_MODE = 1
	MWT_LEFTMULTIPLY  MODIFY_WORLD_TRANSFORM_MODE = 2
	MWT_RIGHTMULTIPLY MODIFY_WORLD_TRANSFORM_MODE = 3
)

// enum
// flags
type FONT_CLIP_PRECISION uint32

const (
	CLIP_CHARACTER_PRECIS FONT_CLIP_PRECISION = 1
	CLIP_DEFAULT_PRECIS   FONT_CLIP_PRECISION = 0
	CLIP_DFA_DISABLE      FONT_CLIP_PRECISION = 64
	CLIP_EMBEDDED         FONT_CLIP_PRECISION = 128
	CLIP_LH_ANGLES        FONT_CLIP_PRECISION = 16
	CLIP_MASK             FONT_CLIP_PRECISION = 15
	CLIP_STROKE_PRECIS    FONT_CLIP_PRECISION = 2
	CLIP_TT_ALWAYS        FONT_CLIP_PRECISION = 32
)

// enum
type CREATE_POLYGON_RGN_MODE uint32

const (
	ALTERNATE CREATE_POLYGON_RGN_MODE = 1
	WINDING   CREATE_POLYGON_RGN_MODE = 2
)

// enum
type EMBEDDED_FONT_PRIV_STATUS uint32

const (
	EMBED_PREVIEWPRINT EMBEDDED_FONT_PRIV_STATUS = 1
	EMBED_EDITABLE     EMBEDDED_FONT_PRIV_STATUS = 2
	EMBED_INSTALLABLE  EMBEDDED_FONT_PRIV_STATUS = 3
	EMBED_NOEMBEDDING  EMBEDDED_FONT_PRIV_STATUS = 4
)

// enum
type MONITOR_FROM_FLAGS uint32

const (
	MONITOR_DEFAULTTONEAREST MONITOR_FROM_FLAGS = 2
	MONITOR_DEFAULTTONULL    MONITOR_FROM_FLAGS = 0
	MONITOR_DEFAULTTOPRIMARY MONITOR_FROM_FLAGS = 1
)

// enum
type FONT_RESOURCE_CHARACTERISTICS uint32

const (
	FR_PRIVATE  FONT_RESOURCE_CHARACTERISTICS = 16
	FR_NOT_ENUM FONT_RESOURCE_CHARACTERISTICS = 32
)

// enum
// flags
type DC_LAYOUT uint32

const (
	LAYOUT_BITMAPORIENTATIONPRESERVED DC_LAYOUT = 8
	LAYOUT_RTL                        DC_LAYOUT = 1
)

// enum
type GET_DEVICE_CAPS_INDEX uint32

const (
	DRIVERVERSION   GET_DEVICE_CAPS_INDEX = 0
	TECHNOLOGY      GET_DEVICE_CAPS_INDEX = 2
	HORZSIZE        GET_DEVICE_CAPS_INDEX = 4
	VERTSIZE        GET_DEVICE_CAPS_INDEX = 6
	HORZRES         GET_DEVICE_CAPS_INDEX = 8
	VERTRES         GET_DEVICE_CAPS_INDEX = 10
	BITSPIXEL       GET_DEVICE_CAPS_INDEX = 12
	PLANES          GET_DEVICE_CAPS_INDEX = 14
	NUMBRUSHES      GET_DEVICE_CAPS_INDEX = 16
	NUMPENS         GET_DEVICE_CAPS_INDEX = 18
	NUMMARKERS      GET_DEVICE_CAPS_INDEX = 20
	NUMFONTS        GET_DEVICE_CAPS_INDEX = 22
	NUMCOLORS       GET_DEVICE_CAPS_INDEX = 24
	PDEVICESIZE     GET_DEVICE_CAPS_INDEX = 26
	CURVECAPS       GET_DEVICE_CAPS_INDEX = 28
	LINECAPS        GET_DEVICE_CAPS_INDEX = 30
	POLYGONALCAPS   GET_DEVICE_CAPS_INDEX = 32
	TEXTCAPS        GET_DEVICE_CAPS_INDEX = 34
	CLIPCAPS        GET_DEVICE_CAPS_INDEX = 36
	RASTERCAPS      GET_DEVICE_CAPS_INDEX = 38
	ASPECTX         GET_DEVICE_CAPS_INDEX = 40
	ASPECTY         GET_DEVICE_CAPS_INDEX = 42
	ASPECTXY        GET_DEVICE_CAPS_INDEX = 44
	LOGPIXELSX      GET_DEVICE_CAPS_INDEX = 88
	LOGPIXELSY      GET_DEVICE_CAPS_INDEX = 90
	SIZEPALETTE     GET_DEVICE_CAPS_INDEX = 104
	NUMRESERVED     GET_DEVICE_CAPS_INDEX = 106
	COLORRES        GET_DEVICE_CAPS_INDEX = 108
	PHYSICALWIDTH   GET_DEVICE_CAPS_INDEX = 110
	PHYSICALHEIGHT  GET_DEVICE_CAPS_INDEX = 111
	PHYSICALOFFSETX GET_DEVICE_CAPS_INDEX = 112
	PHYSICALOFFSETY GET_DEVICE_CAPS_INDEX = 113
	SCALINGFACTORX  GET_DEVICE_CAPS_INDEX = 114
	SCALINGFACTORY  GET_DEVICE_CAPS_INDEX = 115
	VREFRESH        GET_DEVICE_CAPS_INDEX = 116
	DESKTOPVERTRES  GET_DEVICE_CAPS_INDEX = 117
	DESKTOPHORZRES  GET_DEVICE_CAPS_INDEX = 118
	BLTALIGNMENT    GET_DEVICE_CAPS_INDEX = 119
	SHADEBLENDCAPS  GET_DEVICE_CAPS_INDEX = 120
	COLORMGMTCAPS   GET_DEVICE_CAPS_INDEX = 121
)

// enum
type FONT_OUTPUT_PRECISION uint32

const (
	OUT_CHARACTER_PRECIS FONT_OUTPUT_PRECISION = 2
	OUT_DEFAULT_PRECIS   FONT_OUTPUT_PRECISION = 0
	OUT_DEVICE_PRECIS    FONT_OUTPUT_PRECISION = 5
	OUT_OUTLINE_PRECIS   FONT_OUTPUT_PRECISION = 8
	OUT_PS_ONLY_PRECIS   FONT_OUTPUT_PRECISION = 10
	OUT_RASTER_PRECIS    FONT_OUTPUT_PRECISION = 6
	OUT_STRING_PRECIS    FONT_OUTPUT_PRECISION = 1
	OUT_STROKE_PRECIS    FONT_OUTPUT_PRECISION = 3
	OUT_TT_ONLY_PRECIS   FONT_OUTPUT_PRECISION = 7
	OUT_TT_PRECIS        FONT_OUTPUT_PRECISION = 4
)

// enum
type ARC_DIRECTION uint32

const (
	AD_COUNTERCLOCKWISE ARC_DIRECTION = 1
	AD_CLOCKWISE        ARC_DIRECTION = 2
)

// enum
// flags
type TTLOAD_EMBEDDED_FONT_STATUS uint32

const (
	TTLOAD_FONT_SUBSETTED     TTLOAD_EMBEDDED_FONT_STATUS = 1
	TTLOAD_FONT_IN_SYSSTARTUP TTLOAD_EMBEDDED_FONT_STATUS = 2
)

// enum
type STRETCH_BLT_MODE uint32

const (
	BLACKONWHITE        STRETCH_BLT_MODE = 1
	COLORONCOLOR        STRETCH_BLT_MODE = 3
	HALFTONE            STRETCH_BLT_MODE = 4
	STRETCH_ANDSCANS    STRETCH_BLT_MODE = 1
	STRETCH_DELETESCANS STRETCH_BLT_MODE = 3
	STRETCH_HALFTONE    STRETCH_BLT_MODE = 4
	STRETCH_ORSCANS     STRETCH_BLT_MODE = 2
	WHITEONBLACK        STRETCH_BLT_MODE = 2
)

// enum
type FONT_QUALITY uint32

const (
	ANTIALIASED_QUALITY    FONT_QUALITY = 4
	CLEARTYPE_QUALITY      FONT_QUALITY = 5
	DEFAULT_QUALITY        FONT_QUALITY = 0
	DRAFT_QUALITY          FONT_QUALITY = 1
	NONANTIALIASED_QUALITY FONT_QUALITY = 3
	PROOF_QUALITY          FONT_QUALITY = 2
)

// enum
type BACKGROUND_MODE uint32

const (
	OPAQUE      BACKGROUND_MODE = 2
	TRANSPARENT BACKGROUND_MODE = 1
)

// enum
// flags
type GET_CHARACTER_PLACEMENT_FLAGS uint32

const (
	GCP_CLASSIN         GET_CHARACTER_PLACEMENT_FLAGS = 524288
	GCP_DIACRITIC       GET_CHARACTER_PLACEMENT_FLAGS = 256
	GCP_DISPLAYZWG      GET_CHARACTER_PLACEMENT_FLAGS = 4194304
	GCP_GLYPHSHAPE      GET_CHARACTER_PLACEMENT_FLAGS = 16
	GCP_JUSTIFY         GET_CHARACTER_PLACEMENT_FLAGS = 65536
	GCP_KASHIDA         GET_CHARACTER_PLACEMENT_FLAGS = 1024
	GCP_LIGATE          GET_CHARACTER_PLACEMENT_FLAGS = 32
	GCP_MAXEXTENT       GET_CHARACTER_PLACEMENT_FLAGS = 1048576
	GCP_NEUTRALOVERRIDE GET_CHARACTER_PLACEMENT_FLAGS = 33554432
	GCP_NUMERICOVERRIDE GET_CHARACTER_PLACEMENT_FLAGS = 16777216
	GCP_NUMERICSLATIN   GET_CHARACTER_PLACEMENT_FLAGS = 67108864
	GCP_NUMERICSLOCAL   GET_CHARACTER_PLACEMENT_FLAGS = 134217728
	GCP_REORDER         GET_CHARACTER_PLACEMENT_FLAGS = 2
	GCP_SYMSWAPOFF      GET_CHARACTER_PLACEMENT_FLAGS = 8388608
	GCP_USEKERNING      GET_CHARACTER_PLACEMENT_FLAGS = 8
)

// enum
// flags
type DRAW_EDGE_FLAGS uint32

const (
	BF_ADJUST                  DRAW_EDGE_FLAGS = 8192
	BF_BOTTOM                  DRAW_EDGE_FLAGS = 8
	BF_BOTTOMLEFT              DRAW_EDGE_FLAGS = 9
	BF_BOTTOMRIGHT             DRAW_EDGE_FLAGS = 12
	BF_DIAGONAL                DRAW_EDGE_FLAGS = 16
	BF_DIAGONAL_ENDBOTTOMLEFT  DRAW_EDGE_FLAGS = 25
	BF_DIAGONAL_ENDBOTTOMRIGHT DRAW_EDGE_FLAGS = 28
	BF_DIAGONAL_ENDTOPLEFT     DRAW_EDGE_FLAGS = 19
	BF_DIAGONAL_ENDTOPRIGHT    DRAW_EDGE_FLAGS = 22
	BF_FLAT                    DRAW_EDGE_FLAGS = 16384
	BF_LEFT                    DRAW_EDGE_FLAGS = 1
	BF_MIDDLE                  DRAW_EDGE_FLAGS = 2048
	BF_MONO                    DRAW_EDGE_FLAGS = 32768
	BF_RECT                    DRAW_EDGE_FLAGS = 15
	BF_RIGHT                   DRAW_EDGE_FLAGS = 4
	BF_SOFT                    DRAW_EDGE_FLAGS = 4096
	BF_TOP                     DRAW_EDGE_FLAGS = 2
	BF_TOPLEFT                 DRAW_EDGE_FLAGS = 3
	BF_TOPRIGHT                DRAW_EDGE_FLAGS = 6
)

// enum
type FONT_LICENSE_PRIVS uint32

const (
	LICENSE_PREVIEWPRINT FONT_LICENSE_PRIVS = 4
	LICENSE_EDITABLE     FONT_LICENSE_PRIVS = 8
	LICENSE_INSTALLABLE  FONT_LICENSE_PRIVS = 0
	LICENSE_NOEMBEDDING  FONT_LICENSE_PRIVS = 2
	LICENSE_DEFAULT      FONT_LICENSE_PRIVS = 0
)

// enum
type GRADIENT_FILL uint32

const (
	GRADIENT_FILL_RECT_H   GRADIENT_FILL = 0
	GRADIENT_FILL_RECT_V   GRADIENT_FILL = 1
	GRADIENT_FILL_TRIANGLE GRADIENT_FILL = 2
)

// enum
type CREATE_FONT_PACKAGE_SUBSET_ENCODING uint32

const (
	TTFCFP_STD_MAC_CHAR_SET CREATE_FONT_PACKAGE_SUBSET_ENCODING = 0
	TTFCFP_SYMBOL_CHAR_SET  CREATE_FONT_PACKAGE_SUBSET_ENCODING = 0
	TTFCFP_UNICODE_CHAR_SET CREATE_FONT_PACKAGE_SUBSET_ENCODING = 1
)

// enum
type EXT_FLOOD_FILL_TYPE uint32

const (
	FLOODFILLBORDER  EXT_FLOOD_FILL_TYPE = 0
	FLOODFILLSURFACE EXT_FLOOD_FILL_TYPE = 1
)

// enum
type HATCH_BRUSH_STYLE uint32

const (
	HS_BDIAGONAL  HATCH_BRUSH_STYLE = 3
	HS_CROSS      HATCH_BRUSH_STYLE = 4
	HS_DIAGCROSS  HATCH_BRUSH_STYLE = 5
	HS_FDIAGONAL  HATCH_BRUSH_STYLE = 2
	HS_HORIZONTAL HATCH_BRUSH_STYLE = 0
	HS_VERTICAL   HATCH_BRUSH_STYLE = 1
)

// enum
// flags
type DRAW_CAPTION_FLAGS uint32

const (
	DC_ACTIVE   DRAW_CAPTION_FLAGS = 1
	DC_BUTTONS  DRAW_CAPTION_FLAGS = 4096
	DC_GRADIENT DRAW_CAPTION_FLAGS = 32
	DC_ICON     DRAW_CAPTION_FLAGS = 4
	DC_INBUTTON DRAW_CAPTION_FLAGS = 16
	DC_SMALLCAP DRAW_CAPTION_FLAGS = 2
	DC_TEXT     DRAW_CAPTION_FLAGS = 8
)

// enum
type SYSTEM_PALETTE_USE uint32

const (
	SYSPAL_NOSTATIC    SYSTEM_PALETTE_USE = 2
	SYSPAL_NOSTATIC256 SYSTEM_PALETTE_USE = 3
	SYSPAL_STATIC      SYSTEM_PALETTE_USE = 1
)

// enum
type GRAPHICS_MODE uint32

const (
	GM_COMPATIBLE GRAPHICS_MODE = 1
	GM_ADVANCED   GRAPHICS_MODE = 2
)

// enum
type FONT_PITCH_AND_FAMILY uint32

const (
	FF_DECORATIVE FONT_PITCH_AND_FAMILY = 80
	FF_DONTCARE   FONT_PITCH_AND_FAMILY = 0
	FF_MODERN     FONT_PITCH_AND_FAMILY = 48
	FF_ROMAN      FONT_PITCH_AND_FAMILY = 16
	FF_SCRIPT     FONT_PITCH_AND_FAMILY = 64
	FF_SWISS      FONT_PITCH_AND_FAMILY = 32
)

// enum
type CREATE_FONT_PACKAGE_SUBSET_PLATFORM uint32

const (
	TTFCFP_UNICODE_PLATFORMID CREATE_FONT_PACKAGE_SUBSET_PLATFORM = 0
	TTFCFP_ISO_PLATFORMID     CREATE_FONT_PACKAGE_SUBSET_PLATFORM = 2
)

// enum
type HDC_MAP_MODE uint32

const (
	MM_ANISOTROPIC HDC_MAP_MODE = 8
	MM_HIENGLISH   HDC_MAP_MODE = 5
	MM_HIMETRIC    HDC_MAP_MODE = 3
	MM_ISOTROPIC   HDC_MAP_MODE = 7
	MM_LOENGLISH   HDC_MAP_MODE = 4
	MM_LOMETRIC    HDC_MAP_MODE = 2
	MM_TEXT        HDC_MAP_MODE = 1
	MM_TWIPS       HDC_MAP_MODE = 6
)

// enum
type DISPLAYCONFIG_COLOR_ENCODING int32

const (
	DISPLAYCONFIG_COLOR_ENCODING_RGB          DISPLAYCONFIG_COLOR_ENCODING = 0
	DISPLAYCONFIG_COLOR_ENCODING_YCBCR444     DISPLAYCONFIG_COLOR_ENCODING = 1
	DISPLAYCONFIG_COLOR_ENCODING_YCBCR422     DISPLAYCONFIG_COLOR_ENCODING = 2
	DISPLAYCONFIG_COLOR_ENCODING_YCBCR420     DISPLAYCONFIG_COLOR_ENCODING = 3
	DISPLAYCONFIG_COLOR_ENCODING_INTENSITY    DISPLAYCONFIG_COLOR_ENCODING = 4
	DISPLAYCONFIG_COLOR_ENCODING_FORCE_UINT32 DISPLAYCONFIG_COLOR_ENCODING = -1
)

// structs

type MONITORINFOEXA struct {
	MonitorInfo MONITORINFO
	SzDevice    [32]CHAR
}

type MONITORINFOEX = MONITORINFOEXW
type MONITORINFOEXW struct {
	MonitorInfo MONITORINFO
	SzDevice    [32]uint16
}

type XFORM struct {
	EM11 float32
	EM12 float32
	EM21 float32
	EM22 float32
	EDx  float32
	EDy  float32
}

type BITMAP struct {
	BmType       int32
	BmWidth      int32
	BmHeight     int32
	BmWidthBytes int32
	BmPlanes     uint16
	BmBitsPixel  uint16
	BmBits       unsafe.Pointer
}

type RGBTRIPLE struct {
	RgbtBlue  byte
	RgbtGreen byte
	RgbtRed   byte
}

type RGBQUAD struct {
	RgbBlue     byte
	RgbGreen    byte
	RgbRed      byte
	RgbReserved byte
}

type CIEXYZ struct {
	CiexyzX int32
	CiexyzY int32
	CiexyzZ int32
}

type CIEXYZTRIPLE struct {
	CiexyzRed   CIEXYZ
	CiexyzGreen CIEXYZ
	CiexyzBlue  CIEXYZ
}

type BITMAPCOREHEADER struct {
	BcSize     uint32
	BcWidth    uint16
	BcHeight   uint16
	BcPlanes   uint16
	BcBitCount uint16
}

type BITMAPINFOHEADER struct {
	BiSize          uint32
	BiWidth         int32
	BiHeight        int32
	BiPlanes        uint16
	BiBitCount      uint16
	BiCompression   uint32
	BiSizeImage     uint32
	BiXPelsPerMeter int32
	BiYPelsPerMeter int32
	BiClrUsed       uint32
	BiClrImportant  uint32
}

type BITMAPV4HEADER struct {
	BV4Size          uint32
	BV4Width         int32
	BV4Height        int32
	BV4Planes        uint16
	BV4BitCount      uint16
	BV4V4Compression uint32
	BV4SizeImage     uint32
	BV4XPelsPerMeter int32
	BV4YPelsPerMeter int32
	BV4ClrUsed       uint32
	BV4ClrImportant  uint32
	BV4RedMask       uint32
	BV4GreenMask     uint32
	BV4BlueMask      uint32
	BV4AlphaMask     uint32
	BV4CSType        uint32
	BV4Endpoints     CIEXYZTRIPLE
	BV4GammaRed      uint32
	BV4GammaGreen    uint32
	BV4GammaBlue     uint32
}

type BITMAPV5HEADER struct {
	BV5Size          uint32
	BV5Width         int32
	BV5Height        int32
	BV5Planes        uint16
	BV5BitCount      uint16
	BV5Compression   uint32
	BV5SizeImage     uint32
	BV5XPelsPerMeter int32
	BV5YPelsPerMeter int32
	BV5ClrUsed       uint32
	BV5ClrImportant  uint32
	BV5RedMask       uint32
	BV5GreenMask     uint32
	BV5BlueMask      uint32
	BV5AlphaMask     uint32
	BV5CSType        uint32
	BV5Endpoints     CIEXYZTRIPLE
	BV5GammaRed      uint32
	BV5GammaGreen    uint32
	BV5GammaBlue     uint32
	BV5Intent        uint32
	BV5ProfileData   uint32
	BV5ProfileSize   uint32
	BV5Reserved      uint32
}

type BITMAPINFO struct {
	BmiHeader BITMAPINFOHEADER
	BmiColors [1]RGBQUAD
}

type BITMAPCOREINFO struct {
	BmciHeader BITMAPCOREHEADER
	BmciColors [1]RGBTRIPLE
}

type BITMAPFILEHEADER struct {
	BfType      uint16
	BfSize      uint32
	BfReserved1 uint16
	BfReserved2 uint16
	BfOffBits   uint32
}

type HANDLETABLE struct {
	ObjectHandle [1]HGDIOBJ
}

type METARECORD struct {
	RdSize     uint32
	RdFunction uint16
	RdParm     [1]uint16
}

type METAHEADER struct {
	MtType         uint16
	MtHeaderSize   uint16
	MtVersion      uint16
	MtSize         uint32
	MtNoObjects    uint16
	MtMaxRecord    uint32
	MtNoParameters uint16
}

type ENHMETARECORD struct {
	IType uint32
	NSize uint32
	DParm [1]uint32
}

type ENHMETAHEADER struct {
	IType          uint32
	NSize          uint32
	RclBounds      RECTL
	RclFrame       RECTL
	DSignature     uint32
	NVersion       uint32
	NBytes         uint32
	NRecords       uint32
	NHandles       uint16
	SReserved      uint16
	NDescription   uint32
	OffDescription uint32
	NPalEntries    uint32
	SzlDevice      SIZE
	SzlMillimeters SIZE
	CbPixelFormat  uint32
	OffPixelFormat uint32
	BOpenGL        uint32
	SzlMicrometers SIZE
}

type TEXTMETRICA struct {
	TmHeight           int32
	TmAscent           int32
	TmDescent          int32
	TmInternalLeading  int32
	TmExternalLeading  int32
	TmAveCharWidth     int32
	TmMaxCharWidth     int32
	TmWeight           int32
	TmOverhang         int32
	TmDigitizedAspectX int32
	TmDigitizedAspectY int32
	TmFirstChar        byte
	TmLastChar         byte
	TmDefaultChar      byte
	TmBreakChar        byte
	TmItalic           byte
	TmUnderlined       byte
	TmStruckOut        byte
	TmPitchAndFamily   byte
	TmCharSet          byte
}

type TEXTMETRIC = TEXTMETRICW
type TEXTMETRICW struct {
	TmHeight           int32
	TmAscent           int32
	TmDescent          int32
	TmInternalLeading  int32
	TmExternalLeading  int32
	TmAveCharWidth     int32
	TmMaxCharWidth     int32
	TmWeight           int32
	TmOverhang         int32
	TmDigitizedAspectX int32
	TmDigitizedAspectY int32
	TmFirstChar        uint16
	TmLastChar         uint16
	TmDefaultChar      uint16
	TmBreakChar        uint16
	TmItalic           byte
	TmUnderlined       byte
	TmStruckOut        byte
	TmPitchAndFamily   byte
	TmCharSet          byte
}

type NEWTEXTMETRICA struct {
	TmHeight           int32
	TmAscent           int32
	TmDescent          int32
	TmInternalLeading  int32
	TmExternalLeading  int32
	TmAveCharWidth     int32
	TmMaxCharWidth     int32
	TmWeight           int32
	TmOverhang         int32
	TmDigitizedAspectX int32
	TmDigitizedAspectY int32
	TmFirstChar        byte
	TmLastChar         byte
	TmDefaultChar      byte
	TmBreakChar        byte
	TmItalic           byte
	TmUnderlined       byte
	TmStruckOut        byte
	TmPitchAndFamily   byte
	TmCharSet          byte
	NtmFlags           uint32
	NtmSizeEM          uint32
	NtmCellHeight      uint32
	NtmAvgWidth        uint32
}

type NEWTEXTMETRIC = NEWTEXTMETRICW
type NEWTEXTMETRICW struct {
	TmHeight           int32
	TmAscent           int32
	TmDescent          int32
	TmInternalLeading  int32
	TmExternalLeading  int32
	TmAveCharWidth     int32
	TmMaxCharWidth     int32
	TmWeight           int32
	TmOverhang         int32
	TmDigitizedAspectX int32
	TmDigitizedAspectY int32
	TmFirstChar        uint16
	TmLastChar         uint16
	TmDefaultChar      uint16
	TmBreakChar        uint16
	TmItalic           byte
	TmUnderlined       byte
	TmStruckOut        byte
	TmPitchAndFamily   byte
	TmCharSet          byte
	NtmFlags           uint32
	NtmSizeEM          uint32
	NtmCellHeight      uint32
	NtmAvgWidth        uint32
}

type PELARRAY struct {
	PaXCount int32
	PaYCount int32
	PaXExt   int32
	PaYExt   int32
	PaRGBs   byte
}

type LOGBRUSH struct {
	LbStyle uint32
	LbColor uint32
	LbHatch uintptr
}

type LOGBRUSH32 struct {
	LbStyle uint32
	LbColor uint32
	LbHatch uint32
}

type LOGPEN struct {
	LopnStyle uint32
	LopnWidth POINT
	LopnColor uint32
}

type EXTLOGPEN struct {
	ElpPenStyle   uint32
	ElpWidth      uint32
	ElpBrushStyle uint32
	ElpColor      uint32
	ElpHatch      uintptr
	ElpNumEntries uint32
	ElpStyleEntry [1]uint32
}

type EXTLOGPEN32 struct {
	ElpPenStyle   uint32
	ElpWidth      uint32
	ElpBrushStyle uint32
	ElpColor      uint32
	ElpHatch      uint32
	ElpNumEntries uint32
	ElpStyleEntry [1]uint32
}

type PALETTEENTRY struct {
	PeRed   byte
	PeGreen byte
	PeBlue  byte
	PeFlags byte
}

type LOGPALETTE struct {
	PalVersion    uint16
	PalNumEntries uint16
	PalPalEntry   [1]PALETTEENTRY
}

type LOGFONTA struct {
	LfHeight         int32
	LfWidth          int32
	LfEscapement     int32
	LfOrientation    int32
	LfWeight         int32
	LfItalic         byte
	LfUnderline      byte
	LfStrikeOut      byte
	LfCharSet        byte
	LfOutPrecision   byte
	LfClipPrecision  byte
	LfQuality        byte
	LfPitchAndFamily byte
	LfFaceName       [32]CHAR
}

type LOGFONT = LOGFONTW
type LOGFONTW struct {
	LfHeight         int32
	LfWidth          int32
	LfEscapement     int32
	LfOrientation    int32
	LfWeight         int32
	LfItalic         byte
	LfUnderline      byte
	LfStrikeOut      byte
	LfCharSet        byte
	LfOutPrecision   byte
	LfClipPrecision  byte
	LfQuality        byte
	LfPitchAndFamily byte
	LfFaceName       [32]uint16
}

type ENUMLOGFONTA struct {
	ElfLogFont  LOGFONTA
	ElfFullName [64]byte
	ElfStyle    [32]byte
}

type ENUMLOGFONT = ENUMLOGFONTW
type ENUMLOGFONTW struct {
	ElfLogFont  LOGFONTW
	ElfFullName [64]uint16
	ElfStyle    [32]uint16
}

type ENUMLOGFONTEXA struct {
	ElfLogFont  LOGFONTA
	ElfFullName [64]byte
	ElfStyle    [32]byte
	ElfScript   [32]byte
}

type ENUMLOGFONTEX = ENUMLOGFONTEXW
type ENUMLOGFONTEXW struct {
	ElfLogFont  LOGFONTW
	ElfFullName [64]uint16
	ElfStyle    [32]uint16
	ElfScript   [32]uint16
}

type PANOSE struct {
	BFamilyType      byte
	BSerifStyle      byte
	BWeight          byte
	BProportion      byte
	BContrast        byte
	BStrokeVariation byte
	BArmStyle        byte
	BLetterform      byte
	BMidline         byte
	BXHeight         byte
}

type EXTLOGFONTA struct {
	ElfLogFont   LOGFONTA
	ElfFullName  [64]byte
	ElfStyle     [32]byte
	ElfVersion   uint32
	ElfStyleSize uint32
	ElfMatch     uint32
	ElfReserved  uint32
	ElfVendorId  [4]byte
	ElfCulture   uint32
	ElfPanose    PANOSE
}

type EXTLOGFONT = EXTLOGFONTW
type EXTLOGFONTW struct {
	ElfLogFont   LOGFONTW
	ElfFullName  [64]uint16
	ElfStyle     [32]uint16
	ElfVersion   uint32
	ElfStyleSize uint32
	ElfMatch     uint32
	ElfReserved  uint32
	ElfVendorId  [4]byte
	ElfCulture   uint32
	ElfPanose    PANOSE
}

type DEVMODEA_Anonymous1_Anonymous1 struct {
	DmOrientation   int16
	DmPaperSize     int16
	DmPaperLength   int16
	DmPaperWidth    int16
	DmScale         int16
	DmCopies        int16
	DmDefaultSource int16
	DmPrintQuality  int16
}

type DEVMODEA_Anonymous1_Anonymous2 struct {
	DmPosition           POINTL
	DmDisplayOrientation uint32
	DmDisplayFixedOutput uint32
}

type DEVMODEA_Anonymous1 struct {
	Data [4]uint32
}

func (this *DEVMODEA_Anonymous1) Anonymous1() *DEVMODEA_Anonymous1_Anonymous1 {
	return (*DEVMODEA_Anonymous1_Anonymous1)(unsafe.Pointer(this))
}

func (this *DEVMODEA_Anonymous1) Anonymous1Val() DEVMODEA_Anonymous1_Anonymous1 {
	return *(*DEVMODEA_Anonymous1_Anonymous1)(unsafe.Pointer(this))
}

func (this *DEVMODEA_Anonymous1) Anonymous2() *DEVMODEA_Anonymous1_Anonymous2 {
	return (*DEVMODEA_Anonymous1_Anonymous2)(unsafe.Pointer(this))
}

func (this *DEVMODEA_Anonymous1) Anonymous2Val() DEVMODEA_Anonymous1_Anonymous2 {
	return *(*DEVMODEA_Anonymous1_Anonymous2)(unsafe.Pointer(this))
}

type DEVMODEA_Anonymous2 struct {
	Data [1]uint32
}

func (this *DEVMODEA_Anonymous2) DmDisplayFlags() *uint32 {
	return (*uint32)(unsafe.Pointer(this))
}

func (this *DEVMODEA_Anonymous2) DmDisplayFlagsVal() uint32 {
	return *(*uint32)(unsafe.Pointer(this))
}

func (this *DEVMODEA_Anonymous2) DmNup() *uint32 {
	return (*uint32)(unsafe.Pointer(this))
}

func (this *DEVMODEA_Anonymous2) DmNupVal() uint32 {
	return *(*uint32)(unsafe.Pointer(this))
}

type DEVMODEA struct {
	DmDeviceName    [32]byte
	DmSpecVersion   uint16
	DmDriverVersion uint16
	DmSize          uint16
	DmDriverExtra   uint16
	DmFields        uint32
	DEVMODEA_Anonymous1
	DmColor       int16
	DmDuplex      int16
	DmYResolution int16
	DmTTOption    int16
	DmCollate     int16
	DmFormName    [32]byte
	DmLogPixels   uint16
	DmBitsPerPel  uint32
	DmPelsWidth   uint32
	DmPelsHeight  uint32
	DEVMODEA_Anonymous2
	DmDisplayFrequency uint32
	DmICMMethod        uint32
	DmICMIntent        uint32
	DmMediaType        uint32
	DmDitherType       uint32
	DmReserved1        uint32
	DmReserved2        uint32
	DmPanningWidth     uint32
	DmPanningHeight    uint32
}

type DEVMODEW_Anonymous1_Anonymous1 struct {
	DmOrientation   int16
	DmPaperSize     int16
	DmPaperLength   int16
	DmPaperWidth    int16
	DmScale         int16
	DmCopies        int16
	DmDefaultSource int16
	DmPrintQuality  int16
}

type DEVMODEW_Anonymous1_Anonymous2 struct {
	DmPosition           POINTL
	DmDisplayOrientation uint32
	DmDisplayFixedOutput uint32
}

type DEVMODEW_Anonymous1 struct {
	Data [4]uint32
}

func (this *DEVMODEW_Anonymous1) Anonymous1() *DEVMODEW_Anonymous1_Anonymous1 {
	return (*DEVMODEW_Anonymous1_Anonymous1)(unsafe.Pointer(this))
}

func (this *DEVMODEW_Anonymous1) Anonymous1Val() DEVMODEW_Anonymous1_Anonymous1 {
	return *(*DEVMODEW_Anonymous1_Anonymous1)(unsafe.Pointer(this))
}

func (this *DEVMODEW_Anonymous1) Anonymous2() *DEVMODEW_Anonymous1_Anonymous2 {
	return (*DEVMODEW_Anonymous1_Anonymous2)(unsafe.Pointer(this))
}

func (this *DEVMODEW_Anonymous1) Anonymous2Val() DEVMODEW_Anonymous1_Anonymous2 {
	return *(*DEVMODEW_Anonymous1_Anonymous2)(unsafe.Pointer(this))
}

type DEVMODEW_Anonymous2 struct {
	Data [1]uint32
}

func (this *DEVMODEW_Anonymous2) DmDisplayFlags() *uint32 {
	return (*uint32)(unsafe.Pointer(this))
}

func (this *DEVMODEW_Anonymous2) DmDisplayFlagsVal() uint32 {
	return *(*uint32)(unsafe.Pointer(this))
}

func (this *DEVMODEW_Anonymous2) DmNup() *uint32 {
	return (*uint32)(unsafe.Pointer(this))
}

func (this *DEVMODEW_Anonymous2) DmNupVal() uint32 {
	return *(*uint32)(unsafe.Pointer(this))
}

type DEVMODE = DEVMODEW
type DEVMODEW struct {
	DmDeviceName    [32]uint16
	DmSpecVersion   uint16
	DmDriverVersion uint16
	DmSize          uint16
	DmDriverExtra   uint16
	DmFields        uint32
	DEVMODEW_Anonymous1
	DmColor       int16
	DmDuplex      int16
	DmYResolution int16
	DmTTOption    int16
	DmCollate     int16
	DmFormName    [32]uint16
	DmLogPixels   uint16
	DmBitsPerPel  uint32
	DmPelsWidth   uint32
	DmPelsHeight  uint32
	DEVMODEW_Anonymous2
	DmDisplayFrequency uint32
	DmICMMethod        uint32
	DmICMIntent        uint32
	DmMediaType        uint32
	DmDitherType       uint32
	DmReserved1        uint32
	DmReserved2        uint32
	DmPanningWidth     uint32
	DmPanningHeight    uint32
}

type DISPLAY_DEVICEA struct {
	Cb           uint32
	DeviceName   [32]CHAR
	DeviceString [128]CHAR
	StateFlags   uint32
	DeviceID     [128]CHAR
	DeviceKey    [128]CHAR
}

type DISPLAY_DEVICE = DISPLAY_DEVICEW
type DISPLAY_DEVICEW struct {
	Cb           uint32
	DeviceName   [32]uint16
	DeviceString [128]uint16
	StateFlags   uint32
	DeviceID     [128]uint16
	DeviceKey    [128]uint16
}

type RGNDATAHEADER struct {
	DwSize   uint32
	IType    uint32
	NCount   uint32
	NRgnSize uint32
	RcBound  RECT
}

type RGNDATA struct {
	Rdh    RGNDATAHEADER
	Buffer [1]CHAR
}

type ABC struct {
	AbcA int32
	AbcB uint32
	AbcC int32
}

type ABCFLOAT struct {
	AbcfA float32
	AbcfB float32
	AbcfC float32
}

type OUTLINETEXTMETRICA struct {
	OtmSize                uint32
	OtmTextMetrics         TEXTMETRICA
	OtmFiller              byte
	OtmPanoseNumber        PANOSE
	OtmfsSelection         uint32
	OtmfsType              uint32
	OtmsCharSlopeRise      int32
	OtmsCharSlopeRun       int32
	OtmItalicAngle         int32
	OtmEMSquare            uint32
	OtmAscent              int32
	OtmDescent             int32
	OtmLineGap             uint32
	OtmsCapEmHeight        uint32
	OtmsXHeight            uint32
	OtmrcFontBox           RECT
	OtmMacAscent           int32
	OtmMacDescent          int32
	OtmMacLineGap          uint32
	OtmusMinimumPPEM       uint32
	OtmptSubscriptSize     POINT
	OtmptSubscriptOffset   POINT
	OtmptSuperscriptSize   POINT
	OtmptSuperscriptOffset POINT
	OtmsStrikeoutSize      uint32
	OtmsStrikeoutPosition  int32
	OtmsUnderscoreSize     int32
	OtmsUnderscorePosition int32
	OtmpFamilyName         PSTR
	OtmpFaceName           PSTR
	OtmpStyleName          PSTR
	OtmpFullName           PSTR
}

type OUTLINETEXTMETRIC = OUTLINETEXTMETRICW
type OUTLINETEXTMETRICW struct {
	OtmSize                uint32
	OtmTextMetrics         TEXTMETRICW
	OtmFiller              byte
	OtmPanoseNumber        PANOSE
	OtmfsSelection         uint32
	OtmfsType              uint32
	OtmsCharSlopeRise      int32
	OtmsCharSlopeRun       int32
	OtmItalicAngle         int32
	OtmEMSquare            uint32
	OtmAscent              int32
	OtmDescent             int32
	OtmLineGap             uint32
	OtmsCapEmHeight        uint32
	OtmsXHeight            uint32
	OtmrcFontBox           RECT
	OtmMacAscent           int32
	OtmMacDescent          int32
	OtmMacLineGap          uint32
	OtmusMinimumPPEM       uint32
	OtmptSubscriptSize     POINT
	OtmptSubscriptOffset   POINT
	OtmptSuperscriptSize   POINT
	OtmptSuperscriptOffset POINT
	OtmsStrikeoutSize      uint32
	OtmsStrikeoutPosition  int32
	OtmsUnderscoreSize     int32
	OtmsUnderscorePosition int32
	OtmpFamilyName         PSTR
	OtmpFaceName           PSTR
	OtmpStyleName          PSTR
	OtmpFullName           PSTR
}

type POLYTEXTA struct {
	X       int32
	Y       int32
	N       uint32
	Lpstr   PSTR
	UiFlags uint32
	Rcl     RECT
	Pdx     *int32
}

type POLYTEXT = POLYTEXTW
type POLYTEXTW struct {
	X       int32
	Y       int32
	N       uint32
	Lpstr   PWSTR
	UiFlags uint32
	Rcl     RECT
	Pdx     *int32
}

type FIXED struct {
	Fract uint16
	Value int16
}

type MAT2 struct {
	EM11 FIXED
	EM12 FIXED
	EM21 FIXED
	EM22 FIXED
}

type GLYPHMETRICS struct {
	GmBlackBoxX     uint32
	GmBlackBoxY     uint32
	GmptGlyphOrigin POINT
	GmCellIncX      int16
	GmCellIncY      int16
}

type POINTFX struct {
	X FIXED
	Y FIXED
}

type TTPOLYCURVE struct {
	WType uint16
	Cpfx  uint16
	Apfx  [1]POINTFX
}

type TTPOLYGONHEADER struct {
	Cb       uint32
	DwType   uint32
	PfxStart POINTFX
}

type GCP_RESULTSA struct {
	LStructSize uint32
	LpOutString PSTR
	LpOrder     *uint32
	LpDx        *int32
	LpCaretPos  *int32
	LpClass     PSTR
	LpGlyphs    PWSTR
	NGlyphs     uint32
	NMaxFit     int32
}

type GCP_RESULTS = GCP_RESULTSW
type GCP_RESULTSW struct {
	LStructSize uint32
	LpOutString PWSTR
	LpOrder     *uint32
	LpDx        *int32
	LpCaretPos  *int32
	LpClass     PSTR
	LpGlyphs    PWSTR
	NGlyphs     uint32
	NMaxFit     int32
}

type RASTERIZER_STATUS struct {
	NSize       int16
	WFlags      int16
	NLanguageID int16
}

type WCRANGE struct {
	WcLow   uint16
	CGlyphs uint16
}

type GLYPHSET struct {
	CbThis           uint32
	FlAccel          uint32
	CGlyphsSupported uint32
	CRanges          uint32
	Ranges           [1]WCRANGE
}

type DESIGNVECTOR struct {
	DvReserved uint32
	DvNumAxes  uint32
	DvValues   [16]int32
}

type AXISINFOA struct {
	AxMinValue int32
	AxMaxValue int32
	AxAxisName [16]byte
}

type AXISINFO = AXISINFOW
type AXISINFOW struct {
	AxMinValue int32
	AxMaxValue int32
	AxAxisName [16]uint16
}

type AXESLISTA struct {
	AxlReserved uint32
	AxlNumAxes  uint32
	AxlAxisInfo [16]AXISINFOA
}

type AXESLIST = AXESLISTW
type AXESLISTW struct {
	AxlReserved uint32
	AxlNumAxes  uint32
	AxlAxisInfo [16]AXISINFOW
}

type ENUMLOGFONTEXDVA struct {
	ElfEnumLogfontEx ENUMLOGFONTEXA
	ElfDesignVector  DESIGNVECTOR
}

type ENUMLOGFONTEXDV = ENUMLOGFONTEXDVW
type ENUMLOGFONTEXDVW struct {
	ElfEnumLogfontEx ENUMLOGFONTEXW
	ElfDesignVector  DESIGNVECTOR
}

type TRIVERTEX struct {
	X     int32
	Y     int32
	Red   uint16
	Green uint16
	Blue  uint16
	Alpha uint16
}

type GRADIENT_TRIANGLE struct {
	Vertex1 uint32
	Vertex2 uint32
	Vertex3 uint32
}

type GRADIENT_RECT struct {
	UpperLeft  uint32
	LowerRight uint32
}

type BLENDFUNCTION struct {
	BlendOp             byte
	BlendFlags          byte
	SourceConstantAlpha byte
	AlphaFormat         byte
}

type DIBSECTION struct {
	DsBm        BITMAP
	DsBmih      BITMAPINFOHEADER
	DsBitfields [3]uint32
	DshSection  HANDLE
	DsOffset    uint32
}

type COLORADJUSTMENT struct {
	CaSize            uint16
	CaFlags           uint16
	CaIlluminantIndex uint16
	CaRedGamma        uint16
	CaGreenGamma      uint16
	CaBlueGamma       uint16
	CaReferenceBlack  uint16
	CaReferenceWhite  uint16
	CaContrast        int16
	CaBrightness      int16
	CaColorfulness    int16
	CaRedGreenTint    int16
}

type KERNINGPAIR struct {
	WFirst      uint16
	WSecond     uint16
	IKernAmount int32
}

type EMR struct {
	IType uint32
	NSize uint32
}

type EMRTEXT struct {
	PtlReference POINTL
	NChars       uint32
	OffString    uint32
	FOptions     uint32
	Rcl          RECTL
	OffDx        uint32
}

type ABORTPATH struct {
	Emr EMR
}

type EMRSELECTCLIPPATH struct {
	Emr   EMR
	IMode uint32
}

type EMRSETMITERLIMIT struct {
	Emr         EMR
	EMiterLimit float32
}

type EMRRESTOREDC struct {
	Emr       EMR
	IRelative int32
}

type EMRSETARCDIRECTION struct {
	Emr           EMR
	IArcDirection uint32
}

type EMRSETMAPPERFLAGS struct {
	Emr     EMR
	DwFlags uint32
}

type EMRSETTEXTCOLOR struct {
	Emr     EMR
	CrColor uint32
}

type EMRSELECTOBJECT struct {
	Emr      EMR
	IhObject uint32
}

type EMRSELECTPALETTE struct {
	Emr   EMR
	IhPal uint32
}

type EMRRESIZEPALETTE struct {
	Emr      EMR
	IhPal    uint32
	CEntries uint32
}

type EMRSETPALETTEENTRIES struct {
	Emr         EMR
	IhPal       uint32
	IStart      uint32
	CEntries    uint32
	APalEntries [1]PALETTEENTRY
}

type EMRSETCOLORADJUSTMENT struct {
	Emr             EMR
	ColorAdjustment COLORADJUSTMENT
}

type EMRGDICOMMENT struct {
	Emr    EMR
	CbData uint32
	Data   [1]byte
}

type EMREOF struct {
	Emr           EMR
	NPalEntries   uint32
	OffPalEntries uint32
	NSizeLast     uint32
}

type EMRLINETO struct {
	Emr EMR
	Ptl POINTL
}

type EMROFFSETCLIPRGN struct {
	Emr       EMR
	PtlOffset POINTL
}

type EMRFILLPATH struct {
	Emr       EMR
	RclBounds RECTL
}

type EMREXCLUDECLIPRECT struct {
	Emr     EMR
	RclClip RECTL
}

type EMRSETVIEWPORTORGEX struct {
	Emr       EMR
	PtlOrigin POINTL
}

type EMRSETVIEWPORTEXTEX struct {
	Emr       EMR
	SzlExtent SIZE
}

type EMRSCALEVIEWPORTEXTEX struct {
	Emr    EMR
	XNum   int32
	XDenom int32
	YNum   int32
	YDenom int32
}

type EMRSETWORLDTRANSFORM struct {
	Emr   EMR
	Xform XFORM
}

type EMRMODIFYWORLDTRANSFORM struct {
	Emr   EMR
	Xform XFORM
	IMode uint32
}

type EMRSETPIXELV struct {
	Emr      EMR
	PtlPixel POINTL
	CrColor  uint32
}

type EMREXTFLOODFILL struct {
	Emr      EMR
	PtlStart POINTL
	CrColor  uint32
	IMode    uint32
}

type EMRELLIPSE struct {
	Emr    EMR
	RclBox RECTL
}

type EMRROUNDRECT struct {
	Emr       EMR
	RclBox    RECTL
	SzlCorner SIZE
}

type EMRARC struct {
	Emr      EMR
	RclBox   RECTL
	PtlStart POINTL
	PtlEnd   POINTL
}

type EMRANGLEARC struct {
	Emr         EMR
	PtlCenter   POINTL
	NRadius     uint32
	EStartAngle float32
	ESweepAngle float32
}

type EMRPOLYLINE struct {
	Emr       EMR
	RclBounds RECTL
	Cptl      uint32
	Aptl      [1]POINTL
}

type EMRPOLYLINE16 struct {
	Emr       EMR
	RclBounds RECTL
	Cpts      uint32
	Apts      [1]POINTS
}

type EMRPOLYDRAW struct {
	Emr       EMR
	RclBounds RECTL
	Cptl      uint32
	Aptl      [1]POINTL
	AbTypes   [1]byte
}

type EMRPOLYDRAW16 struct {
	Emr       EMR
	RclBounds RECTL
	Cpts      uint32
	Apts      [1]POINTS
	AbTypes   [1]byte
}

type EMRPOLYPOLYLINE struct {
	Emr         EMR
	RclBounds   RECTL
	NPolys      uint32
	Cptl        uint32
	APolyCounts [1]uint32
	Aptl        [1]POINTL
}

type EMRPOLYPOLYLINE16 struct {
	Emr         EMR
	RclBounds   RECTL
	NPolys      uint32
	Cpts        uint32
	APolyCounts [1]uint32
	Apts        [1]POINTS
}

type EMRINVERTRGN struct {
	Emr       EMR
	RclBounds RECTL
	CbRgnData uint32
	RgnData   [1]byte
}

type EMRFILLRGN struct {
	Emr       EMR
	RclBounds RECTL
	CbRgnData uint32
	IhBrush   uint32
	RgnData   [1]byte
}

type EMRFRAMERGN struct {
	Emr       EMR
	RclBounds RECTL
	CbRgnData uint32
	IhBrush   uint32
	SzlStroke SIZE
	RgnData   [1]byte
}

type EMREXTSELECTCLIPRGN struct {
	Emr       EMR
	CbRgnData uint32
	IMode     uint32
	RgnData   [1]byte
}

type EMREXTTEXTOUTA struct {
	Emr           EMR
	RclBounds     RECTL
	IGraphicsMode uint32
	ExScale       float32
	EyScale       float32
	Emrtext       EMRTEXT
}

type EMRPOLYTEXTOUTA struct {
	Emr           EMR
	RclBounds     RECTL
	IGraphicsMode uint32
	ExScale       float32
	EyScale       float32
	CStrings      int32
	Aemrtext      [1]EMRTEXT
}

type EMRBITBLT struct {
	Emr          EMR
	RclBounds    RECTL
	XDest        int32
	YDest        int32
	CxDest       int32
	CyDest       int32
	DwRop        uint32
	XSrc         int32
	YSrc         int32
	XformSrc     XFORM
	CrBkColorSrc uint32
	IUsageSrc    uint32
	OffBmiSrc    uint32
	CbBmiSrc     uint32
	OffBitsSrc   uint32
	CbBitsSrc    uint32
}

type EMRSTRETCHBLT struct {
	Emr          EMR
	RclBounds    RECTL
	XDest        int32
	YDest        int32
	CxDest       int32
	CyDest       int32
	DwRop        uint32
	XSrc         int32
	YSrc         int32
	XformSrc     XFORM
	CrBkColorSrc uint32
	IUsageSrc    uint32
	OffBmiSrc    uint32
	CbBmiSrc     uint32
	OffBitsSrc   uint32
	CbBitsSrc    uint32
	CxSrc        int32
	CySrc        int32
}

type EMRMASKBLT struct {
	Emr          EMR
	RclBounds    RECTL
	XDest        int32
	YDest        int32
	CxDest       int32
	CyDest       int32
	DwRop        uint32
	XSrc         int32
	YSrc         int32
	XformSrc     XFORM
	CrBkColorSrc uint32
	IUsageSrc    uint32
	OffBmiSrc    uint32
	CbBmiSrc     uint32
	OffBitsSrc   uint32
	CbBitsSrc    uint32
	XMask        int32
	YMask        int32
	IUsageMask   uint32
	OffBmiMask   uint32
	CbBmiMask    uint32
	OffBitsMask  uint32
	CbBitsMask   uint32
}

type EMRPLGBLT struct {
	Emr          EMR
	RclBounds    RECTL
	AptlDest     [3]POINTL
	XSrc         int32
	YSrc         int32
	CxSrc        int32
	CySrc        int32
	XformSrc     XFORM
	CrBkColorSrc uint32
	IUsageSrc    uint32
	OffBmiSrc    uint32
	CbBmiSrc     uint32
	OffBitsSrc   uint32
	CbBitsSrc    uint32
	XMask        int32
	YMask        int32
	IUsageMask   uint32
	OffBmiMask   uint32
	CbBmiMask    uint32
	OffBitsMask  uint32
	CbBitsMask   uint32
}

type EMRSETDIBITSTODEVICE struct {
	Emr        EMR
	RclBounds  RECTL
	XDest      int32
	YDest      int32
	XSrc       int32
	YSrc       int32
	CxSrc      int32
	CySrc      int32
	OffBmiSrc  uint32
	CbBmiSrc   uint32
	OffBitsSrc uint32
	CbBitsSrc  uint32
	IUsageSrc  uint32
	IStartScan uint32
	CScans     uint32
}

type EMRSTRETCHDIBITS struct {
	Emr        EMR
	RclBounds  RECTL
	XDest      int32
	YDest      int32
	XSrc       int32
	YSrc       int32
	CxSrc      int32
	CySrc      int32
	OffBmiSrc  uint32
	CbBmiSrc   uint32
	OffBitsSrc uint32
	CbBitsSrc  uint32
	IUsageSrc  uint32
	DwRop      uint32
	CxDest     int32
	CyDest     int32
}

type EMREXTCREATEFONTINDIRECTW struct {
	Emr    EMR
	IhFont uint32
	Elfw   EXTLOGFONTW
}

type EMRCREATEPALETTE struct {
	Emr   EMR
	IhPal uint32
	Lgpl  LOGPALETTE
}

type EMRCREATEPEN struct {
	Emr   EMR
	IhPen uint32
	Lopn  LOGPEN
}

type EMREXTCREATEPEN struct {
	Emr     EMR
	IhPen   uint32
	OffBmi  uint32
	CbBmi   uint32
	OffBits uint32
	CbBits  uint32
	Elp     EXTLOGPEN32
}

type EMRCREATEBRUSHINDIRECT struct {
	Emr     EMR
	IhBrush uint32
	Lb      LOGBRUSH32
}

type EMRCREATEMONOBRUSH struct {
	Emr     EMR
	IhBrush uint32
	IUsage  uint32
	OffBmi  uint32
	CbBmi   uint32
	OffBits uint32
	CbBits  uint32
}

type EMRCREATEDIBPATTERNBRUSHPT struct {
	Emr     EMR
	IhBrush uint32
	IUsage  uint32
	OffBmi  uint32
	CbBmi   uint32
	OffBits uint32
	CbBits  uint32
}

type EMRFORMAT struct {
	DSignature uint32
	NVersion   uint32
	CbData     uint32
	OffData    uint32
}

type EMRGLSRECORD struct {
	Emr    EMR
	CbData uint32
	Data   [1]byte
}

type EMRGLSBOUNDEDRECORD struct {
	Emr       EMR
	RclBounds RECTL
	CbData    uint32
	Data      [1]byte
}

type EMRSETCOLORSPACE struct {
	Emr  EMR
	IhCS uint32
}

type EMREXTESCAPE struct {
	Emr       EMR
	IEscape   int32
	CbEscData int32
	EscData   [1]byte
}

type EMRNAMEDESCAPE struct {
	Emr       EMR
	IEscape   int32
	CbDriver  int32
	CbEscData int32
	EscData   [1]byte
}

type EMRSETICMPROFILE struct {
	Emr     EMR
	DwFlags uint32
	CbName  uint32
	CbData  uint32
	Data    [1]byte
}

type COLORMATCHTOTARGET struct {
	Emr      EMR
	DwAction uint32
	DwFlags  uint32
	CbName   uint32
	CbData   uint32
	Data     [1]byte
}

type COLORCORRECTPALETTE struct {
	Emr         EMR
	IhPalette   uint32
	NFirstEntry uint32
	NPalEntries uint32
	NReserved   uint32
}

type EMRALPHABLEND struct {
	Emr          EMR
	RclBounds    RECTL
	XDest        int32
	YDest        int32
	CxDest       int32
	CyDest       int32
	DwRop        uint32
	XSrc         int32
	YSrc         int32
	XformSrc     XFORM
	CrBkColorSrc uint32
	IUsageSrc    uint32
	OffBmiSrc    uint32
	CbBmiSrc     uint32
	OffBitsSrc   uint32
	CbBitsSrc    uint32
	CxSrc        int32
	CySrc        int32
}

type EMRGRADIENTFILL struct {
	Emr       EMR
	RclBounds RECTL
	NVer      uint32
	NTri      uint32
	UlMode    GRADIENT_FILL
	Ver       [1]TRIVERTEX
}

type EMRTRANSPARENTBLT struct {
	Emr          EMR
	RclBounds    RECTL
	XDest        int32
	YDest        int32
	CxDest       int32
	CyDest       int32
	DwRop        uint32
	XSrc         int32
	YSrc         int32
	XformSrc     XFORM
	CrBkColorSrc uint32
	IUsageSrc    uint32
	OffBmiSrc    uint32
	CbBmiSrc     uint32
	OffBitsSrc   uint32
	CbBitsSrc    uint32
	CxSrc        int32
	CySrc        int32
}

type WGLSWAP struct {
	Hdc     HDC
	UiFlags uint32
}

type TTLOADINFO struct {
	UsStructSize uint16
	UsRefStrSize uint16
	PusRefStr    *uint16
}

type TTEMBEDINFO struct {
	UsStructSize  uint16
	UsRootStrSize uint16
	PusRootStr    *uint16
}

type TTVALIDATIONTESTSPARAMS struct {
	UlStructSize    uint32
	LTestFromSize   int32
	LTestToSize     int32
	UlCharSet       uint32
	UsReserved1     uint16
	UsCharCodeCount uint16
	PusCharCodeSet  *uint16
}

type TTVALIDATIONTESTSPARAMSEX struct {
	UlStructSize    uint32
	LTestFromSize   int32
	LTestToSize     int32
	UlCharSet       uint32
	UsReserved1     uint16
	UsCharCodeCount uint16
	PulCharCodeSet  *uint32
}

type PAINTSTRUCT struct {
	Hdc         HDC
	FErase      BOOL
	RcPaint     RECT
	FRestore    BOOL
	FIncUpdate  BOOL
	RgbReserved [32]byte
}

type DRAWTEXTPARAMS struct {
	CbSize        uint32
	ITabLength    int32
	ILeftMargin   int32
	IRightMargin  int32
	UiLengthDrawn uint32
}

type MONITORINFO struct {
	CbSize    uint32
	RcMonitor RECT
	RcWork    RECT
	DwFlags   uint32
}

// func types

type FONTENUMPROCA = uintptr
type FONTENUMPROCA_func = func(param0 *LOGFONTA, param1 *TEXTMETRICA, param2 uint32, param3 LPARAM) int32

type FONTENUMPROCW = uintptr
type FONTENUMPROCW_func = func(param0 *LOGFONTW, param1 *TEXTMETRICW, param2 uint32, param3 LPARAM) int32

type GOBJENUMPROC = uintptr
type GOBJENUMPROC_func = func(param0 unsafe.Pointer, param1 LPARAM) int32

type LINEDDAPROC = uintptr
type LINEDDAPROC_func = func(param0 int32, param1 int32, param2 LPARAM)

type LPFNDEVMODE = uintptr
type LPFNDEVMODE_func = func(param0 HWND, param1 HINSTANCE, param2 *DEVMODEA, param3 PSTR, param4 PSTR, param5 *DEVMODEA, param6 PSTR, param7 uint32) uint32

type LPFNDEVCAPS = uintptr
type LPFNDEVCAPS_func = func(param0 PSTR, param1 PSTR, param2 uint32, param3 PSTR, param4 *DEVMODEA) uint32

type MFENUMPROC = uintptr
type MFENUMPROC_func = func(hdc HDC, lpht *HANDLETABLE, lpMR *METARECORD, nObj int32, param4 LPARAM) int32

type ENHMFENUMPROC = uintptr
type ENHMFENUMPROC_func = func(hdc HDC, lpht *HANDLETABLE, lpmr *ENHMETARECORD, nHandles int32, data LPARAM) int32

type CFP_ALLOCPROC = uintptr
type CFP_ALLOCPROC_func = func(param0 uintptr) unsafe.Pointer

type CFP_REALLOCPROC = uintptr
type CFP_REALLOCPROC_func = func(param0 unsafe.Pointer, param1 uintptr) unsafe.Pointer

type CFP_FREEPROC = uintptr
type CFP_FREEPROC_func = func(param0 unsafe.Pointer)

type READEMBEDPROC = uintptr
type READEMBEDPROC_func = func(param0 unsafe.Pointer, param1 unsafe.Pointer, param2 uint32) uint32

type WRITEEMBEDPROC = uintptr
type WRITEEMBEDPROC_func = func(param0 unsafe.Pointer, param1 unsafe.Pointer, param2 uint32) uint32

type GRAYSTRINGPROC = uintptr
type GRAYSTRINGPROC_func = func(param0 HDC, param1 LPARAM, param2 int32) BOOL

type DRAWSTATEPROC = uintptr
type DRAWSTATEPROC_func = func(hdc HDC, lData LPARAM, wData WPARAM, cx int32, cy int32) BOOL

type MONITORENUMPROC = uintptr
type MONITORENUMPROC_func = func(param0 HMONITOR, param1 HDC, param2 *RECT, param3 LPARAM) BOOL

var (
	pGetObjectA                   uintptr
	pAddFontResourceA             uintptr
	pAddFontResourceW             uintptr
	pAnimatePalette               uintptr
	pArc                          uintptr
	pBitBlt                       uintptr
	pCancelDC                     uintptr
	pChord                        uintptr
	pCloseMetaFile                uintptr
	pCombineRgn                   uintptr
	pCopyMetaFileA                uintptr
	pCopyMetaFileW                uintptr
	pCreateBitmap                 uintptr
	pCreateBitmapIndirect         uintptr
	pCreateBrushIndirect          uintptr
	pCreateCompatibleBitmap       uintptr
	pCreateDiscardableBitmap      uintptr
	pCreateCompatibleDC           uintptr
	pCreateDCA                    uintptr
	pCreateDCW                    uintptr
	pCreateDIBitmap               uintptr
	pCreateDIBPatternBrush        uintptr
	pCreateDIBPatternBrushPt      uintptr
	pCreateEllipticRgn            uintptr
	pCreateEllipticRgnIndirect    uintptr
	pCreateFontIndirectA          uintptr
	pCreateFontIndirectW          uintptr
	pCreateFontA                  uintptr
	pCreateFontW                  uintptr
	pCreateHatchBrush             uintptr
	pCreateICA                    uintptr
	pCreateICW                    uintptr
	pCreateMetaFileA              uintptr
	pCreateMetaFileW              uintptr
	pCreatePalette                uintptr
	pCreatePen                    uintptr
	pCreatePenIndirect            uintptr
	pCreatePolyPolygonRgn         uintptr
	pCreatePatternBrush           uintptr
	pCreateRectRgn                uintptr
	pCreateRectRgnIndirect        uintptr
	pCreateRoundRectRgn           uintptr
	pCreateScalableFontResourceA  uintptr
	pCreateScalableFontResourceW  uintptr
	pCreateSolidBrush             uintptr
	pDeleteDC                     uintptr
	pDeleteMetaFile               uintptr
	pDeleteObject                 uintptr
	pDrawEscape                   uintptr
	pEllipse                      uintptr
	pEnumFontFamiliesExA          uintptr
	pEnumFontFamiliesExW          uintptr
	pEnumFontFamiliesA            uintptr
	pEnumFontFamiliesW            uintptr
	pEnumFontsA                   uintptr
	pEnumFontsW                   uintptr
	pEnumObjects                  uintptr
	pEqualRgn                     uintptr
	pExcludeClipRect              uintptr
	pExtCreateRegion              uintptr
	pExtFloodFill                 uintptr
	pFillRgn                      uintptr
	pFloodFill                    uintptr
	pFrameRgn                     uintptr
	pGetROP2                      uintptr
	pGetAspectRatioFilterEx       uintptr
	pGetBkColor                   uintptr
	pGetDCBrushColor              uintptr
	pGetDCPenColor                uintptr
	pGetBkMode                    uintptr
	pGetBitmapBits                uintptr
	pGetBitmapDimensionEx         uintptr
	pGetBoundsRect                uintptr
	pGetBrushOrgEx                uintptr
	pGetCharWidthA                uintptr
	pGetCharWidthW                uintptr
	pGetCharWidth32A              uintptr
	pGetCharWidth32W              uintptr
	pGetCharWidthFloatA           uintptr
	pGetCharWidthFloatW           uintptr
	pGetCharABCWidthsA            uintptr
	pGetCharABCWidthsW            uintptr
	pGetCharABCWidthsFloatA       uintptr
	pGetCharABCWidthsFloatW       uintptr
	pGetClipBox                   uintptr
	pGetClipRgn                   uintptr
	pGetMetaRgn                   uintptr
	pGetCurrentObject             uintptr
	pGetCurrentPositionEx         uintptr
	pGetDeviceCaps                uintptr
	pGetDIBits                    uintptr
	pGetFontData                  uintptr
	pGetGlyphOutlineA             uintptr
	pGetGlyphOutlineW             uintptr
	pGetGraphicsMode              uintptr
	pGetMapMode                   uintptr
	pGetMetaFileBitsEx            uintptr
	pGetMetaFileA                 uintptr
	pGetMetaFileW                 uintptr
	pGetNearestColor              uintptr
	pGetNearestPaletteIndex       uintptr
	pGetObjectType                uintptr
	pGetOutlineTextMetricsA       uintptr
	pGetOutlineTextMetricsW       uintptr
	pGetPaletteEntries            uintptr
	pGetPixel                     uintptr
	pGetPolyFillMode              uintptr
	pGetRasterizerCaps            uintptr
	pGetRandomRgn                 uintptr
	pGetRegionData                uintptr
	pGetRgnBox                    uintptr
	pGetStockObject               uintptr
	pGetStretchBltMode            uintptr
	pGetSystemPaletteEntries      uintptr
	pGetSystemPaletteUse          uintptr
	pGetTextCharacterExtra        uintptr
	pGetTextAlign                 uintptr
	pGetTextColor                 uintptr
	pGetTextExtentPointA          uintptr
	pGetTextExtentPointW          uintptr
	pGetTextExtentPoint32A        uintptr
	pGetTextExtentPoint32W        uintptr
	pGetTextExtentExPointA        uintptr
	pGetTextExtentExPointW        uintptr
	pGetFontLanguageInfo          uintptr
	pGetCharacterPlacementA       uintptr
	pGetCharacterPlacementW       uintptr
	pGetFontUnicodeRanges         uintptr
	pGetGlyphIndicesA             uintptr
	pGetGlyphIndicesW             uintptr
	pGetTextExtentPointI          uintptr
	pGetTextExtentExPointI        uintptr
	pGetCharWidthI                uintptr
	pGetCharABCWidthsI            uintptr
	pAddFontResourceExA           uintptr
	pAddFontResourceExW           uintptr
	pRemoveFontResourceExA        uintptr
	pRemoveFontResourceExW        uintptr
	pAddFontMemResourceEx         uintptr
	pRemoveFontMemResourceEx      uintptr
	pCreateFontIndirectExA        uintptr
	pCreateFontIndirectExW        uintptr
	pGetViewportExtEx             uintptr
	pGetViewportOrgEx             uintptr
	pGetWindowExtEx               uintptr
	pGetWindowOrgEx               uintptr
	pIntersectClipRect            uintptr
	pInvertRgn                    uintptr
	pLineDDA                      uintptr
	pLineTo                       uintptr
	pMaskBlt                      uintptr
	pPlgBlt                       uintptr
	pOffsetClipRgn                uintptr
	pOffsetRgn                    uintptr
	pPatBlt                       uintptr
	pPie                          uintptr
	pPlayMetaFile                 uintptr
	pPaintRgn                     uintptr
	pPolyPolygon                  uintptr
	pPtInRegion                   uintptr
	pPtVisible                    uintptr
	pRectInRegion                 uintptr
	pRectVisible                  uintptr
	pRectangle                    uintptr
	pRestoreDC                    uintptr
	pResetDCA                     uintptr
	pResetDCW                     uintptr
	pRealizePalette               uintptr
	pRemoveFontResourceA          uintptr
	pRemoveFontResourceW          uintptr
	pRoundRect                    uintptr
	pResizePalette                uintptr
	pSaveDC                       uintptr
	pSelectClipRgn                uintptr
	pExtSelectClipRgn             uintptr
	pSetMetaRgn                   uintptr
	pSelectObject                 uintptr
	pSelectPalette                uintptr
	pSetBkColor                   uintptr
	pSetDCBrushColor              uintptr
	pSetDCPenColor                uintptr
	pSetBkMode                    uintptr
	pSetBitmapBits                uintptr
	pSetBoundsRect                uintptr
	pSetDIBits                    uintptr
	pSetDIBitsToDevice            uintptr
	pSetMapperFlags               uintptr
	pSetGraphicsMode              uintptr
	pSetMapMode                   uintptr
	pSetLayout                    uintptr
	pGetLayout                    uintptr
	pSetMetaFileBitsEx            uintptr
	pSetPaletteEntries            uintptr
	pSetPixel                     uintptr
	pSetPixelV                    uintptr
	pSetPolyFillMode              uintptr
	pStretchBlt                   uintptr
	pSetRectRgn                   uintptr
	pStretchDIBits                uintptr
	pSetROP2                      uintptr
	pSetStretchBltMode            uintptr
	pSetSystemPaletteUse          uintptr
	pSetTextCharacterExtra        uintptr
	pSetTextColor                 uintptr
	pSetTextAlign                 uintptr
	pSetTextJustification         uintptr
	pUpdateColors                 uintptr
	pAlphaBlend                   uintptr
	pTransparentBlt               uintptr
	pGradientFill                 uintptr
	pGdiAlphaBlend                uintptr
	pGdiTransparentBlt            uintptr
	pGdiGradientFill              uintptr
	pPlayMetaFileRecord           uintptr
	pEnumMetaFile                 uintptr
	pCloseEnhMetaFile             uintptr
	pCopyEnhMetaFileA             uintptr
	pCopyEnhMetaFileW             uintptr
	pCreateEnhMetaFileA           uintptr
	pCreateEnhMetaFileW           uintptr
	pDeleteEnhMetaFile            uintptr
	pEnumEnhMetaFile              uintptr
	pGetEnhMetaFileA              uintptr
	pGetEnhMetaFileW              uintptr
	pGetEnhMetaFileBits           uintptr
	pGetEnhMetaFileDescriptionA   uintptr
	pGetEnhMetaFileDescriptionW   uintptr
	pGetEnhMetaFileHeader         uintptr
	pGetEnhMetaFilePaletteEntries uintptr
	pGetWinMetaFileBits           uintptr
	pPlayEnhMetaFile              uintptr
	pPlayEnhMetaFileRecord        uintptr
	pSetEnhMetaFileBits           uintptr
	pGdiComment                   uintptr
	pGetTextMetricsA              uintptr
	pGetTextMetricsW              uintptr
	pAngleArc                     uintptr
	pPolyPolyline                 uintptr
	pGetWorldTransform            uintptr
	pSetWorldTransform            uintptr
	pModifyWorldTransform         uintptr
	pCombineTransform             uintptr
	pCreateDIBSection             uintptr
	pGetDIBColorTable             uintptr
	pSetDIBColorTable             uintptr
	pSetColorAdjustment           uintptr
	pGetColorAdjustment           uintptr
	pCreateHalftonePalette        uintptr
	pAbortPath                    uintptr
	pArcTo                        uintptr
	pBeginPath                    uintptr
	pCloseFigure                  uintptr
	pEndPath                      uintptr
	pFillPath                     uintptr
	pFlattenPath                  uintptr
	pGetPath                      uintptr
	pPathToRegion                 uintptr
	pPolyDraw                     uintptr
	pSelectClipPath               uintptr
	pSetArcDirection              uintptr
	pSetMiterLimit                uintptr
	pStrokeAndFillPath            uintptr
	pStrokePath                   uintptr
	pWidenPath                    uintptr
	pExtCreatePen                 uintptr
	pGetMiterLimit                uintptr
	pGetArcDirection              uintptr
	pGetObjectW                   uintptr
	pMoveToEx                     uintptr
	pTextOutA                     uintptr
	pTextOutW                     uintptr
	pExtTextOutA                  uintptr
	pExtTextOutW                  uintptr
	pPolyTextOutA                 uintptr
	pPolyTextOutW                 uintptr
	pCreatePolygonRgn             uintptr
	pDPtoLP                       uintptr
	pLPtoDP                       uintptr
	pPolygon                      uintptr
	pPolyline                     uintptr
	pPolyBezier                   uintptr
	pPolyBezierTo                 uintptr
	pPolylineTo                   uintptr
	pSetViewportExtEx             uintptr
	pSetViewportOrgEx             uintptr
	pSetWindowExtEx               uintptr
	pSetWindowOrgEx               uintptr
	pOffsetViewportOrgEx          uintptr
	pOffsetWindowOrgEx            uintptr
	pScaleViewportExtEx           uintptr
	pScaleWindowExtEx             uintptr
	pSetBitmapDimensionEx         uintptr
	pSetBrushOrgEx                uintptr
	pGetTextFaceA                 uintptr
	pGetTextFaceW                 uintptr
	pGetKerningPairsA             uintptr
	pGetKerningPairsW             uintptr
	pGetDCOrgEx                   uintptr
	pFixBrushOrgEx                uintptr
	pUnrealizeObject              uintptr
	pGdiFlush                     uintptr
	pGdiSetBatchLimit             uintptr
	pGdiGetBatchLimit             uintptr
	pDrawEdge                     uintptr
	pDrawFrameControl             uintptr
	pDrawCaption                  uintptr
	pDrawAnimatedRects            uintptr
	pDrawTextA                    uintptr
	pDrawTextW                    uintptr
	pDrawTextExA                  uintptr
	pDrawTextExW                  uintptr
	pGrayStringA                  uintptr
	pGrayStringW                  uintptr
	pDrawStateA                   uintptr
	pDrawStateW                   uintptr
	pTabbedTextOutA               uintptr
	pTabbedTextOutW               uintptr
	pGetTabbedTextExtentA         uintptr
	pGetTabbedTextExtentW         uintptr
	pUpdateWindow                 uintptr
	pPaintDesktop                 uintptr
	pWindowFromDC                 uintptr
	pGetDC                        uintptr
	pGetDCEx                      uintptr
	pGetWindowDC                  uintptr
	pReleaseDC                    uintptr
	pBeginPaint                   uintptr
	pEndPaint                     uintptr
	pGetUpdateRect                uintptr
	pGetUpdateRgn                 uintptr
	pSetWindowRgn                 uintptr
	pGetWindowRgn                 uintptr
	pGetWindowRgnBox              uintptr
	pExcludeUpdateRgn             uintptr
	pInvalidateRect               uintptr
	pValidateRect                 uintptr
	pInvalidateRgn                uintptr
	pValidateRgn                  uintptr
	pRedrawWindow                 uintptr
	pLockWindowUpdate             uintptr
	pClientToScreen               uintptr
	pScreenToClient               uintptr
	pMapWindowPoints              uintptr
	pGetSysColorBrush             uintptr
	pDrawFocusRect                uintptr
	pFillRect                     uintptr
	pFrameRect                    uintptr
	pInvertRect                   uintptr
	pSetRect                      uintptr
	pSetRectEmpty                 uintptr
	pCopyRect                     uintptr
	pInflateRect                  uintptr
	pIntersectRect                uintptr
	pUnionRect                    uintptr
	pSubtractRect                 uintptr
	pOffsetRect                   uintptr
	pIsRectEmpty                  uintptr
	pEqualRect                    uintptr
	pPtInRect                     uintptr
	pLoadBitmapA                  uintptr
	pLoadBitmapW                  uintptr
	pChangeDisplaySettingsA       uintptr
	pChangeDisplaySettingsW       uintptr
	pChangeDisplaySettingsExA     uintptr
	pChangeDisplaySettingsExW     uintptr
	pEnumDisplaySettingsA         uintptr
	pEnumDisplaySettingsW         uintptr
	pEnumDisplaySettingsExA       uintptr
	pEnumDisplaySettingsExW       uintptr
	pEnumDisplayDevicesA          uintptr
	pEnumDisplayDevicesW          uintptr
	pMonitorFromPoint             uintptr
	pMonitorFromRect              uintptr
	pMonitorFromWindow            uintptr
	pGetMonitorInfoA              uintptr
	pGetMonitorInfoW              uintptr
	pEnumDisplayMonitors          uintptr
)

func GetObjectA(h HGDIOBJ, c int32, pv unsafe.Pointer) int32 {
	addr := lazyAddr(&pGetObjectA, libGdi32, "GetObjectA")
	ret, _, _ := syscall.SyscallN(addr, h, uintptr(c), uintptr(pv))
	return int32(ret)
}

func AddFontResourceA(param0 PSTR) int32 {
	addr := lazyAddr(&pAddFontResourceA, libGdi32, "AddFontResourceA")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(param0)))
	return int32(ret)
}

var AddFontResource = AddFontResourceW

func AddFontResourceW(param0 PWSTR) int32 {
	addr := lazyAddr(&pAddFontResourceW, libGdi32, "AddFontResourceW")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(param0)))
	return int32(ret)
}

func AnimatePalette(hPal HPALETTE, iStartIndex uint32, cEntries uint32, ppe *PALETTEENTRY) BOOL {
	addr := lazyAddr(&pAnimatePalette, libGdi32, "AnimatePalette")
	ret, _, _ := syscall.SyscallN(addr, hPal, uintptr(iStartIndex), uintptr(cEntries), uintptr(unsafe.Pointer(ppe)))
	return BOOL(ret)
}

func Arc(hdc HDC, x1 int32, y1 int32, x2 int32, y2 int32, x3 int32, y3 int32, x4 int32, y4 int32) BOOL {
	addr := lazyAddr(&pArc, libGdi32, "Arc")
	ret, _, _ := syscall.SyscallN(addr, hdc, uintptr(x1), uintptr(y1), uintptr(x2), uintptr(y2), uintptr(x3), uintptr(y3), uintptr(x4), uintptr(y4))
	return BOOL(ret)
}

func BitBlt(hdc HDC, x int32, y int32, cx int32, cy int32, hdcSrc HDC, x1 int32, y1 int32, rop ROP_CODE) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pBitBlt, libGdi32, "BitBlt")
	ret, _, err := syscall.SyscallN(addr, hdc, uintptr(x), uintptr(y), uintptr(cx), uintptr(cy), hdcSrc, uintptr(x1), uintptr(y1), uintptr(rop))
	return BOOL(ret), WIN32_ERROR(err)
}

func CancelDC(hdc HDC) BOOL {
	addr := lazyAddr(&pCancelDC, libGdi32, "CancelDC")
	ret, _, _ := syscall.SyscallN(addr, hdc)
	return BOOL(ret)
}

func Chord(hdc HDC, x1 int32, y1 int32, x2 int32, y2 int32, x3 int32, y3 int32, x4 int32, y4 int32) BOOL {
	addr := lazyAddr(&pChord, libGdi32, "Chord")
	ret, _, _ := syscall.SyscallN(addr, hdc, uintptr(x1), uintptr(y1), uintptr(x2), uintptr(y2), uintptr(x3), uintptr(y3), uintptr(x4), uintptr(y4))
	return BOOL(ret)
}

func CloseMetaFile(hdc HDC) HMETAFILE {
	addr := lazyAddr(&pCloseMetaFile, libGdi32, "CloseMetaFile")
	ret, _, _ := syscall.SyscallN(addr, hdc)
	return ret
}

func CombineRgn(hrgnDst HRGN, hrgnSrc1 HRGN, hrgnSrc2 HRGN, iMode RGN_COMBINE_MODE) int32 {
	addr := lazyAddr(&pCombineRgn, libGdi32, "CombineRgn")
	ret, _, _ := syscall.SyscallN(addr, hrgnDst, hrgnSrc1, hrgnSrc2, uintptr(iMode))
	return int32(ret)
}

func CopyMetaFileA(param0 HMETAFILE, param1 PSTR) HMETAFILE {
	addr := lazyAddr(&pCopyMetaFileA, libGdi32, "CopyMetaFileA")
	ret, _, _ := syscall.SyscallN(addr, param0, uintptr(unsafe.Pointer(param1)))
	return ret
}

var CopyMetaFile = CopyMetaFileW

func CopyMetaFileW(param0 HMETAFILE, param1 PWSTR) HMETAFILE {
	addr := lazyAddr(&pCopyMetaFileW, libGdi32, "CopyMetaFileW")
	ret, _, _ := syscall.SyscallN(addr, param0, uintptr(unsafe.Pointer(param1)))
	return ret
}

func CreateBitmap(nWidth int32, nHeight int32, nPlanes uint32, nBitCount uint32, lpBits unsafe.Pointer) HBITMAP {
	addr := lazyAddr(&pCreateBitmap, libGdi32, "CreateBitmap")
	ret, _, _ := syscall.SyscallN(addr, uintptr(nWidth), uintptr(nHeight), uintptr(nPlanes), uintptr(nBitCount), uintptr(lpBits))
	return ret
}

func CreateBitmapIndirect(pbm *BITMAP) HBITMAP {
	addr := lazyAddr(&pCreateBitmapIndirect, libGdi32, "CreateBitmapIndirect")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pbm)))
	return ret
}

func CreateBrushIndirect(plbrush *LOGBRUSH) HBRUSH {
	addr := lazyAddr(&pCreateBrushIndirect, libGdi32, "CreateBrushIndirect")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(plbrush)))
	return ret
}

func CreateCompatibleBitmap(hdc HDC, cx int32, cy int32) HBITMAP {
	addr := lazyAddr(&pCreateCompatibleBitmap, libGdi32, "CreateCompatibleBitmap")
	ret, _, _ := syscall.SyscallN(addr, hdc, uintptr(cx), uintptr(cy))
	return ret
}

func CreateDiscardableBitmap(hdc HDC, cx int32, cy int32) HBITMAP {
	addr := lazyAddr(&pCreateDiscardableBitmap, libGdi32, "CreateDiscardableBitmap")
	ret, _, _ := syscall.SyscallN(addr, hdc, uintptr(cx), uintptr(cy))
	return ret
}

func CreateCompatibleDC(hdc HDC) CreatedHDC {
	addr := lazyAddr(&pCreateCompatibleDC, libGdi32, "CreateCompatibleDC")
	ret, _, _ := syscall.SyscallN(addr, hdc)
	return ret
}

func CreateDCA(pwszDriver PSTR, pwszDevice PSTR, pszPort PSTR, pdm *DEVMODEA) CreatedHDC {
	addr := lazyAddr(&pCreateDCA, libGdi32, "CreateDCA")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pwszDriver)), uintptr(unsafe.Pointer(pwszDevice)), uintptr(unsafe.Pointer(pszPort)), uintptr(unsafe.Pointer(pdm)))
	return ret
}

var CreateDC = CreateDCW

func CreateDCW(pwszDriver PWSTR, pwszDevice PWSTR, pszPort PWSTR, pdm *DEVMODEW) CreatedHDC {
	addr := lazyAddr(&pCreateDCW, libGdi32, "CreateDCW")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pwszDriver)), uintptr(unsafe.Pointer(pwszDevice)), uintptr(unsafe.Pointer(pszPort)), uintptr(unsafe.Pointer(pdm)))
	return ret
}

func CreateDIBitmap(hdc HDC, pbmih *BITMAPINFOHEADER, flInit uint32, pjBits unsafe.Pointer, pbmi *BITMAPINFO, iUsage DIB_USAGE) HBITMAP {
	addr := lazyAddr(&pCreateDIBitmap, libGdi32, "CreateDIBitmap")
	ret, _, _ := syscall.SyscallN(addr, hdc, uintptr(unsafe.Pointer(pbmih)), uintptr(flInit), uintptr(pjBits), uintptr(unsafe.Pointer(pbmi)), uintptr(iUsage))
	return ret
}

func CreateDIBPatternBrush(h uintptr, iUsage DIB_USAGE) HBRUSH {
	addr := lazyAddr(&pCreateDIBPatternBrush, libGdi32, "CreateDIBPatternBrush")
	ret, _, _ := syscall.SyscallN(addr, h, uintptr(iUsage))
	return ret
}

func CreateDIBPatternBrushPt(lpPackedDIB unsafe.Pointer, iUsage DIB_USAGE) HBRUSH {
	addr := lazyAddr(&pCreateDIBPatternBrushPt, libGdi32, "CreateDIBPatternBrushPt")
	ret, _, _ := syscall.SyscallN(addr, uintptr(lpPackedDIB), uintptr(iUsage))
	return ret
}

func CreateEllipticRgn(x1 int32, y1 int32, x2 int32, y2 int32) HRGN {
	addr := lazyAddr(&pCreateEllipticRgn, libGdi32, "CreateEllipticRgn")
	ret, _, _ := syscall.SyscallN(addr, uintptr(x1), uintptr(y1), uintptr(x2), uintptr(y2))
	return ret
}

func CreateEllipticRgnIndirect(lprect *RECT) HRGN {
	addr := lazyAddr(&pCreateEllipticRgnIndirect, libGdi32, "CreateEllipticRgnIndirect")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lprect)))
	return ret
}

func CreateFontIndirectA(lplf *LOGFONTA) HFONT {
	addr := lazyAddr(&pCreateFontIndirectA, libGdi32, "CreateFontIndirectA")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lplf)))
	return ret
}

var CreateFontIndirect = CreateFontIndirectW

func CreateFontIndirectW(lplf *LOGFONTW) HFONT {
	addr := lazyAddr(&pCreateFontIndirectW, libGdi32, "CreateFontIndirectW")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lplf)))
	return ret
}

func CreateFontA(cHeight int32, cWidth int32, cEscapement int32, cOrientation int32, cWeight int32, bItalic uint32, bUnderline uint32, bStrikeOut uint32, iCharSet uint32, iOutPrecision FONT_OUTPUT_PRECISION, iClipPrecision FONT_CLIP_PRECISION, iQuality FONT_QUALITY, iPitchAndFamily FONT_PITCH_AND_FAMILY, pszFaceName PSTR) HFONT {
	addr := lazyAddr(&pCreateFontA, libGdi32, "CreateFontA")
	ret, _, _ := syscall.SyscallN(addr, uintptr(cHeight), uintptr(cWidth), uintptr(cEscapement), uintptr(cOrientation), uintptr(cWeight), uintptr(bItalic), uintptr(bUnderline), uintptr(bStrikeOut), uintptr(iCharSet), uintptr(iOutPrecision), uintptr(iClipPrecision), uintptr(iQuality), uintptr(iPitchAndFamily), uintptr(unsafe.Pointer(pszFaceName)))
	return ret
}

var CreateFont = CreateFontW

func CreateFontW(cHeight int32, cWidth int32, cEscapement int32, cOrientation int32, cWeight int32, bItalic uint32, bUnderline uint32, bStrikeOut uint32, iCharSet uint32, iOutPrecision FONT_OUTPUT_PRECISION, iClipPrecision FONT_CLIP_PRECISION, iQuality FONT_QUALITY, iPitchAndFamily FONT_PITCH_AND_FAMILY, pszFaceName PWSTR) HFONT {
	addr := lazyAddr(&pCreateFontW, libGdi32, "CreateFontW")
	ret, _, _ := syscall.SyscallN(addr, uintptr(cHeight), uintptr(cWidth), uintptr(cEscapement), uintptr(cOrientation), uintptr(cWeight), uintptr(bItalic), uintptr(bUnderline), uintptr(bStrikeOut), uintptr(iCharSet), uintptr(iOutPrecision), uintptr(iClipPrecision), uintptr(iQuality), uintptr(iPitchAndFamily), uintptr(unsafe.Pointer(pszFaceName)))
	return ret
}

func CreateHatchBrush(iHatch HATCH_BRUSH_STYLE, color uint32) HBRUSH {
	addr := lazyAddr(&pCreateHatchBrush, libGdi32, "CreateHatchBrush")
	ret, _, _ := syscall.SyscallN(addr, uintptr(iHatch), uintptr(color))
	return ret
}

func CreateICA(pszDriver PSTR, pszDevice PSTR, pszPort PSTR, pdm *DEVMODEA) CreatedHDC {
	addr := lazyAddr(&pCreateICA, libGdi32, "CreateICA")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pszDriver)), uintptr(unsafe.Pointer(pszDevice)), uintptr(unsafe.Pointer(pszPort)), uintptr(unsafe.Pointer(pdm)))
	return ret
}

var CreateIC = CreateICW

func CreateICW(pszDriver PWSTR, pszDevice PWSTR, pszPort PWSTR, pdm *DEVMODEW) CreatedHDC {
	addr := lazyAddr(&pCreateICW, libGdi32, "CreateICW")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pszDriver)), uintptr(unsafe.Pointer(pszDevice)), uintptr(unsafe.Pointer(pszPort)), uintptr(unsafe.Pointer(pdm)))
	return ret
}

func CreateMetaFileA(pszFile PSTR) HdcMetdataFileHandle {
	addr := lazyAddr(&pCreateMetaFileA, libGdi32, "CreateMetaFileA")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pszFile)))
	return ret
}

var CreateMetaFile = CreateMetaFileW

func CreateMetaFileW(pszFile PWSTR) HdcMetdataFileHandle {
	addr := lazyAddr(&pCreateMetaFileW, libGdi32, "CreateMetaFileW")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pszFile)))
	return ret
}

func CreatePalette(plpal *LOGPALETTE) HPALETTE {
	addr := lazyAddr(&pCreatePalette, libGdi32, "CreatePalette")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(plpal)))
	return ret
}

func CreatePen(iStyle PEN_STYLE, cWidth int32, color uint32) HPEN {
	addr := lazyAddr(&pCreatePen, libGdi32, "CreatePen")
	ret, _, _ := syscall.SyscallN(addr, uintptr(iStyle), uintptr(cWidth), uintptr(color))
	return ret
}

func CreatePenIndirect(plpen *LOGPEN) HPEN {
	addr := lazyAddr(&pCreatePenIndirect, libGdi32, "CreatePenIndirect")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(plpen)))
	return ret
}

func CreatePolyPolygonRgn(pptl *POINT, pc *int32, cPoly int32, iMode CREATE_POLYGON_RGN_MODE) HRGN {
	addr := lazyAddr(&pCreatePolyPolygonRgn, libGdi32, "CreatePolyPolygonRgn")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pptl)), uintptr(unsafe.Pointer(pc)), uintptr(cPoly), uintptr(iMode))
	return ret
}

func CreatePatternBrush(hbm HBITMAP) HBRUSH {
	addr := lazyAddr(&pCreatePatternBrush, libGdi32, "CreatePatternBrush")
	ret, _, _ := syscall.SyscallN(addr, hbm)
	return ret
}

func CreateRectRgn(x1 int32, y1 int32, x2 int32, y2 int32) HRGN {
	addr := lazyAddr(&pCreateRectRgn, libGdi32, "CreateRectRgn")
	ret, _, _ := syscall.SyscallN(addr, uintptr(x1), uintptr(y1), uintptr(x2), uintptr(y2))
	return ret
}

func CreateRectRgnIndirect(lprect *RECT) HRGN {
	addr := lazyAddr(&pCreateRectRgnIndirect, libGdi32, "CreateRectRgnIndirect")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lprect)))
	return ret
}

func CreateRoundRectRgn(x1 int32, y1 int32, x2 int32, y2 int32, w int32, h int32) HRGN {
	addr := lazyAddr(&pCreateRoundRectRgn, libGdi32, "CreateRoundRectRgn")
	ret, _, _ := syscall.SyscallN(addr, uintptr(x1), uintptr(y1), uintptr(x2), uintptr(y2), uintptr(w), uintptr(h))
	return ret
}

func CreateScalableFontResourceA(fdwHidden uint32, lpszFont PSTR, lpszFile PSTR, lpszPath PSTR) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pCreateScalableFontResourceA, libGdi32, "CreateScalableFontResourceA")
	ret, _, err := syscall.SyscallN(addr, uintptr(fdwHidden), uintptr(unsafe.Pointer(lpszFont)), uintptr(unsafe.Pointer(lpszFile)), uintptr(unsafe.Pointer(lpszPath)))
	return BOOL(ret), WIN32_ERROR(err)
}

var CreateScalableFontResource = CreateScalableFontResourceW

func CreateScalableFontResourceW(fdwHidden uint32, lpszFont PWSTR, lpszFile PWSTR, lpszPath PWSTR) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pCreateScalableFontResourceW, libGdi32, "CreateScalableFontResourceW")
	ret, _, err := syscall.SyscallN(addr, uintptr(fdwHidden), uintptr(unsafe.Pointer(lpszFont)), uintptr(unsafe.Pointer(lpszFile)), uintptr(unsafe.Pointer(lpszPath)))
	return BOOL(ret), WIN32_ERROR(err)
}

func CreateSolidBrush(color uint32) HBRUSH {
	addr := lazyAddr(&pCreateSolidBrush, libGdi32, "CreateSolidBrush")
	ret, _, _ := syscall.SyscallN(addr, uintptr(color))
	return ret
}

func DeleteDC(hdc CreatedHDC) BOOL {
	addr := lazyAddr(&pDeleteDC, libGdi32, "DeleteDC")
	ret, _, _ := syscall.SyscallN(addr, hdc)
	return BOOL(ret)
}

func DeleteMetaFile(hmf HMETAFILE) BOOL {
	addr := lazyAddr(&pDeleteMetaFile, libGdi32, "DeleteMetaFile")
	ret, _, _ := syscall.SyscallN(addr, hmf)
	return BOOL(ret)
}

func DeleteObject(ho HGDIOBJ) BOOL {
	addr := lazyAddr(&pDeleteObject, libGdi32, "DeleteObject")
	ret, _, _ := syscall.SyscallN(addr, ho)
	return BOOL(ret)
}

func DrawEscape(hdc HDC, iEscape int32, cjIn int32, lpIn PSTR) int32 {
	addr := lazyAddr(&pDrawEscape, libGdi32, "DrawEscape")
	ret, _, _ := syscall.SyscallN(addr, hdc, uintptr(iEscape), uintptr(cjIn), uintptr(unsafe.Pointer(lpIn)))
	return int32(ret)
}

func Ellipse(hdc HDC, left int32, top int32, right int32, bottom int32) BOOL {
	addr := lazyAddr(&pEllipse, libGdi32, "Ellipse")
	ret, _, _ := syscall.SyscallN(addr, hdc, uintptr(left), uintptr(top), uintptr(right), uintptr(bottom))
	return BOOL(ret)
}

func EnumFontFamiliesExA(hdc HDC, lpLogfont *LOGFONTA, lpProc FONTENUMPROCA, lParam LPARAM, dwFlags uint32) int32 {
	addr := lazyAddr(&pEnumFontFamiliesExA, libGdi32, "EnumFontFamiliesExA")
	ret, _, _ := syscall.SyscallN(addr, hdc, uintptr(unsafe.Pointer(lpLogfont)), lpProc, lParam, uintptr(dwFlags))
	return int32(ret)
}

var EnumFontFamiliesEx = EnumFontFamiliesExW

func EnumFontFamiliesExW(hdc HDC, lpLogfont *LOGFONTW, lpProc FONTENUMPROCW, lParam LPARAM, dwFlags uint32) int32 {
	addr := lazyAddr(&pEnumFontFamiliesExW, libGdi32, "EnumFontFamiliesExW")
	ret, _, _ := syscall.SyscallN(addr, hdc, uintptr(unsafe.Pointer(lpLogfont)), lpProc, lParam, uintptr(dwFlags))
	return int32(ret)
}

func EnumFontFamiliesA(hdc HDC, lpLogfont PSTR, lpProc FONTENUMPROCA, lParam LPARAM) int32 {
	addr := lazyAddr(&pEnumFontFamiliesA, libGdi32, "EnumFontFamiliesA")
	ret, _, _ := syscall.SyscallN(addr, hdc, uintptr(unsafe.Pointer(lpLogfont)), lpProc, lParam)
	return int32(ret)
}

var EnumFontFamilies = EnumFontFamiliesW

func EnumFontFamiliesW(hdc HDC, lpLogfont PWSTR, lpProc FONTENUMPROCW, lParam LPARAM) int32 {
	addr := lazyAddr(&pEnumFontFamiliesW, libGdi32, "EnumFontFamiliesW")
	ret, _, _ := syscall.SyscallN(addr, hdc, uintptr(unsafe.Pointer(lpLogfont)), lpProc, lParam)
	return int32(ret)
}

func EnumFontsA(hdc HDC, lpLogfont PSTR, lpProc FONTENUMPROCA, lParam LPARAM) int32 {
	addr := lazyAddr(&pEnumFontsA, libGdi32, "EnumFontsA")
	ret, _, _ := syscall.SyscallN(addr, hdc, uintptr(unsafe.Pointer(lpLogfont)), lpProc, lParam)
	return int32(ret)
}

var EnumFonts = EnumFontsW

func EnumFontsW(hdc HDC, lpLogfont PWSTR, lpProc FONTENUMPROCW, lParam LPARAM) int32 {
	addr := lazyAddr(&pEnumFontsW, libGdi32, "EnumFontsW")
	ret, _, _ := syscall.SyscallN(addr, hdc, uintptr(unsafe.Pointer(lpLogfont)), lpProc, lParam)
	return int32(ret)
}

func EnumObjects(hdc HDC, nType OBJ_TYPE, lpFunc GOBJENUMPROC, lParam LPARAM) int32 {
	addr := lazyAddr(&pEnumObjects, libGdi32, "EnumObjects")
	ret, _, _ := syscall.SyscallN(addr, hdc, uintptr(nType), lpFunc, lParam)
	return int32(ret)
}

func EqualRgn(hrgn1 HRGN, hrgn2 HRGN) BOOL {
	addr := lazyAddr(&pEqualRgn, libGdi32, "EqualRgn")
	ret, _, _ := syscall.SyscallN(addr, hrgn1, hrgn2)
	return BOOL(ret)
}

func ExcludeClipRect(hdc HDC, left int32, top int32, right int32, bottom int32) int32 {
	addr := lazyAddr(&pExcludeClipRect, libGdi32, "ExcludeClipRect")
	ret, _, _ := syscall.SyscallN(addr, hdc, uintptr(left), uintptr(top), uintptr(right), uintptr(bottom))
	return int32(ret)
}

func ExtCreateRegion(lpx *XFORM, nCount uint32, lpData *RGNDATA) HRGN {
	addr := lazyAddr(&pExtCreateRegion, libGdi32, "ExtCreateRegion")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpx)), uintptr(nCount), uintptr(unsafe.Pointer(lpData)))
	return ret
}

func ExtFloodFill(hdc HDC, x int32, y int32, color uint32, type_ EXT_FLOOD_FILL_TYPE) BOOL {
	addr := lazyAddr(&pExtFloodFill, libGdi32, "ExtFloodFill")
	ret, _, _ := syscall.SyscallN(addr, hdc, uintptr(x), uintptr(y), uintptr(color), uintptr(type_))
	return BOOL(ret)
}

func FillRgn(hdc HDC, hrgn HRGN, hbr HBRUSH) BOOL {
	addr := lazyAddr(&pFillRgn, libGdi32, "FillRgn")
	ret, _, _ := syscall.SyscallN(addr, hdc, hrgn, hbr)
	return BOOL(ret)
}

func FloodFill(hdc HDC, x int32, y int32, color uint32) BOOL {
	addr := lazyAddr(&pFloodFill, libGdi32, "FloodFill")
	ret, _, _ := syscall.SyscallN(addr, hdc, uintptr(x), uintptr(y), uintptr(color))
	return BOOL(ret)
}

func FrameRgn(hdc HDC, hrgn HRGN, hbr HBRUSH, w int32, h int32) BOOL {
	addr := lazyAddr(&pFrameRgn, libGdi32, "FrameRgn")
	ret, _, _ := syscall.SyscallN(addr, hdc, hrgn, hbr, uintptr(w), uintptr(h))
	return BOOL(ret)
}

func GetROP2(hdc HDC) int32 {
	addr := lazyAddr(&pGetROP2, libGdi32, "GetROP2")
	ret, _, _ := syscall.SyscallN(addr, hdc)
	return int32(ret)
}

func GetAspectRatioFilterEx(hdc HDC, lpsize *SIZE) BOOL {
	addr := lazyAddr(&pGetAspectRatioFilterEx, libGdi32, "GetAspectRatioFilterEx")
	ret, _, _ := syscall.SyscallN(addr, hdc, uintptr(unsafe.Pointer(lpsize)))
	return BOOL(ret)
}

func GetBkColor(hdc HDC) uint32 {
	addr := lazyAddr(&pGetBkColor, libGdi32, "GetBkColor")
	ret, _, _ := syscall.SyscallN(addr, hdc)
	return uint32(ret)
}

func GetDCBrushColor(hdc HDC) uint32 {
	addr := lazyAddr(&pGetDCBrushColor, libGdi32, "GetDCBrushColor")
	ret, _, _ := syscall.SyscallN(addr, hdc)
	return uint32(ret)
}

func GetDCPenColor(hdc HDC) uint32 {
	addr := lazyAddr(&pGetDCPenColor, libGdi32, "GetDCPenColor")
	ret, _, _ := syscall.SyscallN(addr, hdc)
	return uint32(ret)
}

func GetBkMode(hdc HDC) int32 {
	addr := lazyAddr(&pGetBkMode, libGdi32, "GetBkMode")
	ret, _, _ := syscall.SyscallN(addr, hdc)
	return int32(ret)
}

func GetBitmapBits(hbit HBITMAP, cb int32, lpvBits unsafe.Pointer) int32 {
	addr := lazyAddr(&pGetBitmapBits, libGdi32, "GetBitmapBits")
	ret, _, _ := syscall.SyscallN(addr, hbit, uintptr(cb), uintptr(lpvBits))
	return int32(ret)
}

func GetBitmapDimensionEx(hbit HBITMAP, lpsize *SIZE) BOOL {
	addr := lazyAddr(&pGetBitmapDimensionEx, libGdi32, "GetBitmapDimensionEx")
	ret, _, _ := syscall.SyscallN(addr, hbit, uintptr(unsafe.Pointer(lpsize)))
	return BOOL(ret)
}

func GetBoundsRect(hdc HDC, lprect *RECT, flags uint32) uint32 {
	addr := lazyAddr(&pGetBoundsRect, libGdi32, "GetBoundsRect")
	ret, _, _ := syscall.SyscallN(addr, hdc, uintptr(unsafe.Pointer(lprect)), uintptr(flags))
	return uint32(ret)
}

func GetBrushOrgEx(hdc HDC, lppt *POINT) BOOL {
	addr := lazyAddr(&pGetBrushOrgEx, libGdi32, "GetBrushOrgEx")
	ret, _, _ := syscall.SyscallN(addr, hdc, uintptr(unsafe.Pointer(lppt)))
	return BOOL(ret)
}

func GetCharWidthA(hdc HDC, iFirst uint32, iLast uint32, lpBuffer *int32) BOOL {
	addr := lazyAddr(&pGetCharWidthA, libGdi32, "GetCharWidthA")
	ret, _, _ := syscall.SyscallN(addr, hdc, uintptr(iFirst), uintptr(iLast), uintptr(unsafe.Pointer(lpBuffer)))
	return BOOL(ret)
}

var GetCharWidth = GetCharWidthW

func GetCharWidthW(hdc HDC, iFirst uint32, iLast uint32, lpBuffer *int32) BOOL {
	addr := lazyAddr(&pGetCharWidthW, libGdi32, "GetCharWidthW")
	ret, _, _ := syscall.SyscallN(addr, hdc, uintptr(iFirst), uintptr(iLast), uintptr(unsafe.Pointer(lpBuffer)))
	return BOOL(ret)
}

func GetCharWidth32A(hdc HDC, iFirst uint32, iLast uint32, lpBuffer *int32) BOOL {
	addr := lazyAddr(&pGetCharWidth32A, libGdi32, "GetCharWidth32A")
	ret, _, _ := syscall.SyscallN(addr, hdc, uintptr(iFirst), uintptr(iLast), uintptr(unsafe.Pointer(lpBuffer)))
	return BOOL(ret)
}

var GetCharWidth32 = GetCharWidth32W

func GetCharWidth32W(hdc HDC, iFirst uint32, iLast uint32, lpBuffer *int32) BOOL {
	addr := lazyAddr(&pGetCharWidth32W, libGdi32, "GetCharWidth32W")
	ret, _, _ := syscall.SyscallN(addr, hdc, uintptr(iFirst), uintptr(iLast), uintptr(unsafe.Pointer(lpBuffer)))
	return BOOL(ret)
}

func GetCharWidthFloatA(hdc HDC, iFirst uint32, iLast uint32, lpBuffer *float32) BOOL {
	addr := lazyAddr(&pGetCharWidthFloatA, libGdi32, "GetCharWidthFloatA")
	ret, _, _ := syscall.SyscallN(addr, hdc, uintptr(iFirst), uintptr(iLast), uintptr(unsafe.Pointer(lpBuffer)))
	return BOOL(ret)
}

var GetCharWidthFloat = GetCharWidthFloatW

func GetCharWidthFloatW(hdc HDC, iFirst uint32, iLast uint32, lpBuffer *float32) BOOL {
	addr := lazyAddr(&pGetCharWidthFloatW, libGdi32, "GetCharWidthFloatW")
	ret, _, _ := syscall.SyscallN(addr, hdc, uintptr(iFirst), uintptr(iLast), uintptr(unsafe.Pointer(lpBuffer)))
	return BOOL(ret)
}

func GetCharABCWidthsA(hdc HDC, wFirst uint32, wLast uint32, lpABC *ABC) BOOL {
	addr := lazyAddr(&pGetCharABCWidthsA, libGdi32, "GetCharABCWidthsA")
	ret, _, _ := syscall.SyscallN(addr, hdc, uintptr(wFirst), uintptr(wLast), uintptr(unsafe.Pointer(lpABC)))
	return BOOL(ret)
}

var GetCharABCWidths = GetCharABCWidthsW

func GetCharABCWidthsW(hdc HDC, wFirst uint32, wLast uint32, lpABC *ABC) BOOL {
	addr := lazyAddr(&pGetCharABCWidthsW, libGdi32, "GetCharABCWidthsW")
	ret, _, _ := syscall.SyscallN(addr, hdc, uintptr(wFirst), uintptr(wLast), uintptr(unsafe.Pointer(lpABC)))
	return BOOL(ret)
}

func GetCharABCWidthsFloatA(hdc HDC, iFirst uint32, iLast uint32, lpABC *ABCFLOAT) BOOL {
	addr := lazyAddr(&pGetCharABCWidthsFloatA, libGdi32, "GetCharABCWidthsFloatA")
	ret, _, _ := syscall.SyscallN(addr, hdc, uintptr(iFirst), uintptr(iLast), uintptr(unsafe.Pointer(lpABC)))
	return BOOL(ret)
}

var GetCharABCWidthsFloat = GetCharABCWidthsFloatW

func GetCharABCWidthsFloatW(hdc HDC, iFirst uint32, iLast uint32, lpABC *ABCFLOAT) BOOL {
	addr := lazyAddr(&pGetCharABCWidthsFloatW, libGdi32, "GetCharABCWidthsFloatW")
	ret, _, _ := syscall.SyscallN(addr, hdc, uintptr(iFirst), uintptr(iLast), uintptr(unsafe.Pointer(lpABC)))
	return BOOL(ret)
}

func GetClipBox(hdc HDC, lprect *RECT) int32 {
	addr := lazyAddr(&pGetClipBox, libGdi32, "GetClipBox")
	ret, _, _ := syscall.SyscallN(addr, hdc, uintptr(unsafe.Pointer(lprect)))
	return int32(ret)
}

func GetClipRgn(hdc HDC, hrgn HRGN) int32 {
	addr := lazyAddr(&pGetClipRgn, libGdi32, "GetClipRgn")
	ret, _, _ := syscall.SyscallN(addr, hdc, hrgn)
	return int32(ret)
}

func GetMetaRgn(hdc HDC, hrgn HRGN) int32 {
	addr := lazyAddr(&pGetMetaRgn, libGdi32, "GetMetaRgn")
	ret, _, _ := syscall.SyscallN(addr, hdc, hrgn)
	return int32(ret)
}

func GetCurrentObject(hdc HDC, type_ OBJ_TYPE) HGDIOBJ {
	addr := lazyAddr(&pGetCurrentObject, libGdi32, "GetCurrentObject")
	ret, _, _ := syscall.SyscallN(addr, hdc, uintptr(type_))
	return ret
}

func GetCurrentPositionEx(hdc HDC, lppt *POINT) BOOL {
	addr := lazyAddr(&pGetCurrentPositionEx, libGdi32, "GetCurrentPositionEx")
	ret, _, _ := syscall.SyscallN(addr, hdc, uintptr(unsafe.Pointer(lppt)))
	return BOOL(ret)
}

func GetDeviceCaps(hdc HDC, index GET_DEVICE_CAPS_INDEX) int32 {
	addr := lazyAddr(&pGetDeviceCaps, libGdi32, "GetDeviceCaps")
	ret, _, _ := syscall.SyscallN(addr, hdc, uintptr(index))
	return int32(ret)
}

func GetDIBits(hdc HDC, hbm HBITMAP, start uint32, cLines uint32, lpvBits unsafe.Pointer, lpbmi *BITMAPINFO, usage DIB_USAGE) int32 {
	addr := lazyAddr(&pGetDIBits, libGdi32, "GetDIBits")
	ret, _, _ := syscall.SyscallN(addr, hdc, hbm, uintptr(start), uintptr(cLines), uintptr(lpvBits), uintptr(unsafe.Pointer(lpbmi)), uintptr(usage))
	return int32(ret)
}

func GetFontData(hdc HDC, dwTable uint32, dwOffset uint32, pvBuffer unsafe.Pointer, cjBuffer uint32) uint32 {
	addr := lazyAddr(&pGetFontData, libGdi32, "GetFontData")
	ret, _, _ := syscall.SyscallN(addr, hdc, uintptr(dwTable), uintptr(dwOffset), uintptr(pvBuffer), uintptr(cjBuffer))
	return uint32(ret)
}

func GetGlyphOutlineA(hdc HDC, uChar uint32, fuFormat GET_GLYPH_OUTLINE_FORMAT, lpgm *GLYPHMETRICS, cjBuffer uint32, pvBuffer unsafe.Pointer, lpmat2 *MAT2) uint32 {
	addr := lazyAddr(&pGetGlyphOutlineA, libGdi32, "GetGlyphOutlineA")
	ret, _, _ := syscall.SyscallN(addr, hdc, uintptr(uChar), uintptr(fuFormat), uintptr(unsafe.Pointer(lpgm)), uintptr(cjBuffer), uintptr(pvBuffer), uintptr(unsafe.Pointer(lpmat2)))
	return uint32(ret)
}

var GetGlyphOutline = GetGlyphOutlineW

func GetGlyphOutlineW(hdc HDC, uChar uint32, fuFormat GET_GLYPH_OUTLINE_FORMAT, lpgm *GLYPHMETRICS, cjBuffer uint32, pvBuffer unsafe.Pointer, lpmat2 *MAT2) uint32 {
	addr := lazyAddr(&pGetGlyphOutlineW, libGdi32, "GetGlyphOutlineW")
	ret, _, _ := syscall.SyscallN(addr, hdc, uintptr(uChar), uintptr(fuFormat), uintptr(unsafe.Pointer(lpgm)), uintptr(cjBuffer), uintptr(pvBuffer), uintptr(unsafe.Pointer(lpmat2)))
	return uint32(ret)
}

func GetGraphicsMode(hdc HDC) int32 {
	addr := lazyAddr(&pGetGraphicsMode, libGdi32, "GetGraphicsMode")
	ret, _, _ := syscall.SyscallN(addr, hdc)
	return int32(ret)
}

func GetMapMode(hdc HDC) int32 {
	addr := lazyAddr(&pGetMapMode, libGdi32, "GetMapMode")
	ret, _, _ := syscall.SyscallN(addr, hdc)
	return int32(ret)
}

func GetMetaFileBitsEx(hMF HMETAFILE, cbBuffer uint32, lpData unsafe.Pointer) uint32 {
	addr := lazyAddr(&pGetMetaFileBitsEx, libGdi32, "GetMetaFileBitsEx")
	ret, _, _ := syscall.SyscallN(addr, hMF, uintptr(cbBuffer), uintptr(lpData))
	return uint32(ret)
}

func GetMetaFileA(lpName PSTR) HMETAFILE {
	addr := lazyAddr(&pGetMetaFileA, libGdi32, "GetMetaFileA")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpName)))
	return ret
}

var GetMetaFile = GetMetaFileW

func GetMetaFileW(lpName PWSTR) HMETAFILE {
	addr := lazyAddr(&pGetMetaFileW, libGdi32, "GetMetaFileW")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpName)))
	return ret
}

func GetNearestColor(hdc HDC, color uint32) uint32 {
	addr := lazyAddr(&pGetNearestColor, libGdi32, "GetNearestColor")
	ret, _, _ := syscall.SyscallN(addr, hdc, uintptr(color))
	return uint32(ret)
}

func GetNearestPaletteIndex(h HPALETTE, color uint32) uint32 {
	addr := lazyAddr(&pGetNearestPaletteIndex, libGdi32, "GetNearestPaletteIndex")
	ret, _, _ := syscall.SyscallN(addr, h, uintptr(color))
	return uint32(ret)
}

func GetObjectType(h HGDIOBJ) uint32 {
	addr := lazyAddr(&pGetObjectType, libGdi32, "GetObjectType")
	ret, _, _ := syscall.SyscallN(addr, h)
	return uint32(ret)
}

func GetOutlineTextMetricsA(hdc HDC, cjCopy uint32, potm *OUTLINETEXTMETRICA) uint32 {
	addr := lazyAddr(&pGetOutlineTextMetricsA, libGdi32, "GetOutlineTextMetricsA")
	ret, _, _ := syscall.SyscallN(addr, hdc, uintptr(cjCopy), uintptr(unsafe.Pointer(potm)))
	return uint32(ret)
}

var GetOutlineTextMetrics = GetOutlineTextMetricsW

func GetOutlineTextMetricsW(hdc HDC, cjCopy uint32, potm *OUTLINETEXTMETRICW) uint32 {
	addr := lazyAddr(&pGetOutlineTextMetricsW, libGdi32, "GetOutlineTextMetricsW")
	ret, _, _ := syscall.SyscallN(addr, hdc, uintptr(cjCopy), uintptr(unsafe.Pointer(potm)))
	return uint32(ret)
}

func GetPaletteEntries(hpal HPALETTE, iStart uint32, cEntries uint32, pPalEntries *PALETTEENTRY) uint32 {
	addr := lazyAddr(&pGetPaletteEntries, libGdi32, "GetPaletteEntries")
	ret, _, _ := syscall.SyscallN(addr, hpal, uintptr(iStart), uintptr(cEntries), uintptr(unsafe.Pointer(pPalEntries)))
	return uint32(ret)
}

func GetPixel(hdc HDC, x int32, y int32) uint32 {
	addr := lazyAddr(&pGetPixel, libGdi32, "GetPixel")
	ret, _, _ := syscall.SyscallN(addr, hdc, uintptr(x), uintptr(y))
	return uint32(ret)
}

func GetPolyFillMode(hdc HDC) int32 {
	addr := lazyAddr(&pGetPolyFillMode, libGdi32, "GetPolyFillMode")
	ret, _, _ := syscall.SyscallN(addr, hdc)
	return int32(ret)
}

func GetRasterizerCaps(lpraststat *RASTERIZER_STATUS, cjBytes uint32) BOOL {
	addr := lazyAddr(&pGetRasterizerCaps, libGdi32, "GetRasterizerCaps")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpraststat)), uintptr(cjBytes))
	return BOOL(ret)
}

func GetRandomRgn(hdc HDC, hrgn HRGN, i int32) int32 {
	addr := lazyAddr(&pGetRandomRgn, libGdi32, "GetRandomRgn")
	ret, _, _ := syscall.SyscallN(addr, hdc, hrgn, uintptr(i))
	return int32(ret)
}

func GetRegionData(hrgn HRGN, nCount uint32, lpRgnData *RGNDATA) uint32 {
	addr := lazyAddr(&pGetRegionData, libGdi32, "GetRegionData")
	ret, _, _ := syscall.SyscallN(addr, hrgn, uintptr(nCount), uintptr(unsafe.Pointer(lpRgnData)))
	return uint32(ret)
}

func GetRgnBox(hrgn HRGN, lprc *RECT) int32 {
	addr := lazyAddr(&pGetRgnBox, libGdi32, "GetRgnBox")
	ret, _, _ := syscall.SyscallN(addr, hrgn, uintptr(unsafe.Pointer(lprc)))
	return int32(ret)
}

func GetStockObject(i GET_STOCK_OBJECT_FLAGS) HGDIOBJ {
	addr := lazyAddr(&pGetStockObject, libGdi32, "GetStockObject")
	ret, _, _ := syscall.SyscallN(addr, uintptr(i))
	return ret
}

func GetStretchBltMode(hdc HDC) int32 {
	addr := lazyAddr(&pGetStretchBltMode, libGdi32, "GetStretchBltMode")
	ret, _, _ := syscall.SyscallN(addr, hdc)
	return int32(ret)
}

func GetSystemPaletteEntries(hdc HDC, iStart uint32, cEntries uint32, pPalEntries *PALETTEENTRY) uint32 {
	addr := lazyAddr(&pGetSystemPaletteEntries, libGdi32, "GetSystemPaletteEntries")
	ret, _, _ := syscall.SyscallN(addr, hdc, uintptr(iStart), uintptr(cEntries), uintptr(unsafe.Pointer(pPalEntries)))
	return uint32(ret)
}

func GetSystemPaletteUse(hdc HDC) uint32 {
	addr := lazyAddr(&pGetSystemPaletteUse, libGdi32, "GetSystemPaletteUse")
	ret, _, _ := syscall.SyscallN(addr, hdc)
	return uint32(ret)
}

func GetTextCharacterExtra(hdc HDC) int32 {
	addr := lazyAddr(&pGetTextCharacterExtra, libGdi32, "GetTextCharacterExtra")
	ret, _, _ := syscall.SyscallN(addr, hdc)
	return int32(ret)
}

func GetTextAlign(hdc HDC) uint32 {
	addr := lazyAddr(&pGetTextAlign, libGdi32, "GetTextAlign")
	ret, _, _ := syscall.SyscallN(addr, hdc)
	return uint32(ret)
}

func GetTextColor(hdc HDC) uint32 {
	addr := lazyAddr(&pGetTextColor, libGdi32, "GetTextColor")
	ret, _, _ := syscall.SyscallN(addr, hdc)
	return uint32(ret)
}

func GetTextExtentPointA(hdc HDC, lpString PSTR, c int32, lpsz *SIZE) BOOL {
	addr := lazyAddr(&pGetTextExtentPointA, libGdi32, "GetTextExtentPointA")
	ret, _, _ := syscall.SyscallN(addr, hdc, uintptr(unsafe.Pointer(lpString)), uintptr(c), uintptr(unsafe.Pointer(lpsz)))
	return BOOL(ret)
}

var GetTextExtentPoint = GetTextExtentPointW

func GetTextExtentPointW(hdc HDC, lpString PWSTR, c int32, lpsz *SIZE) BOOL {
	addr := lazyAddr(&pGetTextExtentPointW, libGdi32, "GetTextExtentPointW")
	ret, _, _ := syscall.SyscallN(addr, hdc, uintptr(unsafe.Pointer(lpString)), uintptr(c), uintptr(unsafe.Pointer(lpsz)))
	return BOOL(ret)
}

func GetTextExtentPoint32A(hdc HDC, lpString PSTR, c int32, psizl *SIZE) BOOL {
	addr := lazyAddr(&pGetTextExtentPoint32A, libGdi32, "GetTextExtentPoint32A")
	ret, _, _ := syscall.SyscallN(addr, hdc, uintptr(unsafe.Pointer(lpString)), uintptr(c), uintptr(unsafe.Pointer(psizl)))
	return BOOL(ret)
}

var GetTextExtentPoint32 = GetTextExtentPoint32W

func GetTextExtentPoint32W(hdc HDC, lpString PWSTR, c int32, psizl *SIZE) BOOL {
	addr := lazyAddr(&pGetTextExtentPoint32W, libGdi32, "GetTextExtentPoint32W")
	ret, _, _ := syscall.SyscallN(addr, hdc, uintptr(unsafe.Pointer(lpString)), uintptr(c), uintptr(unsafe.Pointer(psizl)))
	return BOOL(ret)
}

func GetTextExtentExPointA(hdc HDC, lpszString PSTR, cchString int32, nMaxExtent int32, lpnFit *int32, lpnDx *int32, lpSize *SIZE) BOOL {
	addr := lazyAddr(&pGetTextExtentExPointA, libGdi32, "GetTextExtentExPointA")
	ret, _, _ := syscall.SyscallN(addr, hdc, uintptr(unsafe.Pointer(lpszString)), uintptr(cchString), uintptr(nMaxExtent), uintptr(unsafe.Pointer(lpnFit)), uintptr(unsafe.Pointer(lpnDx)), uintptr(unsafe.Pointer(lpSize)))
	return BOOL(ret)
}

var GetTextExtentExPoint = GetTextExtentExPointW

func GetTextExtentExPointW(hdc HDC, lpszString PWSTR, cchString int32, nMaxExtent int32, lpnFit *int32, lpnDx *int32, lpSize *SIZE) BOOL {
	addr := lazyAddr(&pGetTextExtentExPointW, libGdi32, "GetTextExtentExPointW")
	ret, _, _ := syscall.SyscallN(addr, hdc, uintptr(unsafe.Pointer(lpszString)), uintptr(cchString), uintptr(nMaxExtent), uintptr(unsafe.Pointer(lpnFit)), uintptr(unsafe.Pointer(lpnDx)), uintptr(unsafe.Pointer(lpSize)))
	return BOOL(ret)
}

func GetFontLanguageInfo(hdc HDC) uint32 {
	addr := lazyAddr(&pGetFontLanguageInfo, libGdi32, "GetFontLanguageInfo")
	ret, _, _ := syscall.SyscallN(addr, hdc)
	return uint32(ret)
}

func GetCharacterPlacementA(hdc HDC, lpString PSTR, nCount int32, nMexExtent int32, lpResults *GCP_RESULTSA, dwFlags GET_CHARACTER_PLACEMENT_FLAGS) uint32 {
	addr := lazyAddr(&pGetCharacterPlacementA, libGdi32, "GetCharacterPlacementA")
	ret, _, _ := syscall.SyscallN(addr, hdc, uintptr(unsafe.Pointer(lpString)), uintptr(nCount), uintptr(nMexExtent), uintptr(unsafe.Pointer(lpResults)), uintptr(dwFlags))
	return uint32(ret)
}

var GetCharacterPlacement = GetCharacterPlacementW

func GetCharacterPlacementW(hdc HDC, lpString PWSTR, nCount int32, nMexExtent int32, lpResults *GCP_RESULTSW, dwFlags GET_CHARACTER_PLACEMENT_FLAGS) uint32 {
	addr := lazyAddr(&pGetCharacterPlacementW, libGdi32, "GetCharacterPlacementW")
	ret, _, _ := syscall.SyscallN(addr, hdc, uintptr(unsafe.Pointer(lpString)), uintptr(nCount), uintptr(nMexExtent), uintptr(unsafe.Pointer(lpResults)), uintptr(dwFlags))
	return uint32(ret)
}

func GetFontUnicodeRanges(hdc HDC, lpgs *GLYPHSET) uint32 {
	addr := lazyAddr(&pGetFontUnicodeRanges, libGdi32, "GetFontUnicodeRanges")
	ret, _, _ := syscall.SyscallN(addr, hdc, uintptr(unsafe.Pointer(lpgs)))
	return uint32(ret)
}

func GetGlyphIndicesA(hdc HDC, lpstr PSTR, c int32, pgi *uint16, fl uint32) uint32 {
	addr := lazyAddr(&pGetGlyphIndicesA, libGdi32, "GetGlyphIndicesA")
	ret, _, _ := syscall.SyscallN(addr, hdc, uintptr(unsafe.Pointer(lpstr)), uintptr(c), uintptr(unsafe.Pointer(pgi)), uintptr(fl))
	return uint32(ret)
}

var GetGlyphIndices = GetGlyphIndicesW

func GetGlyphIndicesW(hdc HDC, lpstr PWSTR, c int32, pgi *uint16, fl uint32) uint32 {
	addr := lazyAddr(&pGetGlyphIndicesW, libGdi32, "GetGlyphIndicesW")
	ret, _, _ := syscall.SyscallN(addr, hdc, uintptr(unsafe.Pointer(lpstr)), uintptr(c), uintptr(unsafe.Pointer(pgi)), uintptr(fl))
	return uint32(ret)
}

func GetTextExtentPointI(hdc HDC, pgiIn *uint16, cgi int32, psize *SIZE) BOOL {
	addr := lazyAddr(&pGetTextExtentPointI, libGdi32, "GetTextExtentPointI")
	ret, _, _ := syscall.SyscallN(addr, hdc, uintptr(unsafe.Pointer(pgiIn)), uintptr(cgi), uintptr(unsafe.Pointer(psize)))
	return BOOL(ret)
}

func GetTextExtentExPointI(hdc HDC, lpwszString *uint16, cwchString int32, nMaxExtent int32, lpnFit *int32, lpnDx *int32, lpSize *SIZE) BOOL {
	addr := lazyAddr(&pGetTextExtentExPointI, libGdi32, "GetTextExtentExPointI")
	ret, _, _ := syscall.SyscallN(addr, hdc, uintptr(unsafe.Pointer(lpwszString)), uintptr(cwchString), uintptr(nMaxExtent), uintptr(unsafe.Pointer(lpnFit)), uintptr(unsafe.Pointer(lpnDx)), uintptr(unsafe.Pointer(lpSize)))
	return BOOL(ret)
}

func GetCharWidthI(hdc HDC, giFirst uint32, cgi uint32, pgi *uint16, piWidths *int32) BOOL {
	addr := lazyAddr(&pGetCharWidthI, libGdi32, "GetCharWidthI")
	ret, _, _ := syscall.SyscallN(addr, hdc, uintptr(giFirst), uintptr(cgi), uintptr(unsafe.Pointer(pgi)), uintptr(unsafe.Pointer(piWidths)))
	return BOOL(ret)
}

func GetCharABCWidthsI(hdc HDC, giFirst uint32, cgi uint32, pgi *uint16, pabc *ABC) BOOL {
	addr := lazyAddr(&pGetCharABCWidthsI, libGdi32, "GetCharABCWidthsI")
	ret, _, _ := syscall.SyscallN(addr, hdc, uintptr(giFirst), uintptr(cgi), uintptr(unsafe.Pointer(pgi)), uintptr(unsafe.Pointer(pabc)))
	return BOOL(ret)
}

func AddFontResourceExA(name PSTR, fl FONT_RESOURCE_CHARACTERISTICS, res unsafe.Pointer) int32 {
	addr := lazyAddr(&pAddFontResourceExA, libGdi32, "AddFontResourceExA")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(name)), uintptr(fl), uintptr(res))
	return int32(ret)
}

var AddFontResourceEx = AddFontResourceExW

func AddFontResourceExW(name PWSTR, fl FONT_RESOURCE_CHARACTERISTICS, res unsafe.Pointer) int32 {
	addr := lazyAddr(&pAddFontResourceExW, libGdi32, "AddFontResourceExW")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(name)), uintptr(fl), uintptr(res))
	return int32(ret)
}

func RemoveFontResourceExA(name PSTR, fl uint32, pdv unsafe.Pointer) BOOL {
	addr := lazyAddr(&pRemoveFontResourceExA, libGdi32, "RemoveFontResourceExA")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(name)), uintptr(fl), uintptr(pdv))
	return BOOL(ret)
}

var RemoveFontResourceEx = RemoveFontResourceExW

func RemoveFontResourceExW(name PWSTR, fl uint32, pdv unsafe.Pointer) BOOL {
	addr := lazyAddr(&pRemoveFontResourceExW, libGdi32, "RemoveFontResourceExW")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(name)), uintptr(fl), uintptr(pdv))
	return BOOL(ret)
}

func AddFontMemResourceEx(pFileView unsafe.Pointer, cjSize uint32, pvResrved unsafe.Pointer, pNumFonts *uint32) HANDLE {
	addr := lazyAddr(&pAddFontMemResourceEx, libGdi32, "AddFontMemResourceEx")
	ret, _, _ := syscall.SyscallN(addr, uintptr(pFileView), uintptr(cjSize), uintptr(pvResrved), uintptr(unsafe.Pointer(pNumFonts)))
	return ret
}

func RemoveFontMemResourceEx(h HANDLE) BOOL {
	addr := lazyAddr(&pRemoveFontMemResourceEx, libGdi32, "RemoveFontMemResourceEx")
	ret, _, _ := syscall.SyscallN(addr, h)
	return BOOL(ret)
}

func CreateFontIndirectExA(param0 *ENUMLOGFONTEXDVA) HFONT {
	addr := lazyAddr(&pCreateFontIndirectExA, libGdi32, "CreateFontIndirectExA")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(param0)))
	return ret
}

var CreateFontIndirectEx = CreateFontIndirectExW

func CreateFontIndirectExW(param0 *ENUMLOGFONTEXDVW) HFONT {
	addr := lazyAddr(&pCreateFontIndirectExW, libGdi32, "CreateFontIndirectExW")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(param0)))
	return ret
}

func GetViewportExtEx(hdc HDC, lpsize *SIZE) BOOL {
	addr := lazyAddr(&pGetViewportExtEx, libGdi32, "GetViewportExtEx")
	ret, _, _ := syscall.SyscallN(addr, hdc, uintptr(unsafe.Pointer(lpsize)))
	return BOOL(ret)
}

func GetViewportOrgEx(hdc HDC, lppoint *POINT) BOOL {
	addr := lazyAddr(&pGetViewportOrgEx, libGdi32, "GetViewportOrgEx")
	ret, _, _ := syscall.SyscallN(addr, hdc, uintptr(unsafe.Pointer(lppoint)))
	return BOOL(ret)
}

func GetWindowExtEx(hdc HDC, lpsize *SIZE) BOOL {
	addr := lazyAddr(&pGetWindowExtEx, libGdi32, "GetWindowExtEx")
	ret, _, _ := syscall.SyscallN(addr, hdc, uintptr(unsafe.Pointer(lpsize)))
	return BOOL(ret)
}

func GetWindowOrgEx(hdc HDC, lppoint *POINT) BOOL {
	addr := lazyAddr(&pGetWindowOrgEx, libGdi32, "GetWindowOrgEx")
	ret, _, _ := syscall.SyscallN(addr, hdc, uintptr(unsafe.Pointer(lppoint)))
	return BOOL(ret)
}

func IntersectClipRect(hdc HDC, left int32, top int32, right int32, bottom int32) int32 {
	addr := lazyAddr(&pIntersectClipRect, libGdi32, "IntersectClipRect")
	ret, _, _ := syscall.SyscallN(addr, hdc, uintptr(left), uintptr(top), uintptr(right), uintptr(bottom))
	return int32(ret)
}

func InvertRgn(hdc HDC, hrgn HRGN) BOOL {
	addr := lazyAddr(&pInvertRgn, libGdi32, "InvertRgn")
	ret, _, _ := syscall.SyscallN(addr, hdc, hrgn)
	return BOOL(ret)
}

func LineDDA(xStart int32, yStart int32, xEnd int32, yEnd int32, lpProc LINEDDAPROC, data LPARAM) BOOL {
	addr := lazyAddr(&pLineDDA, libGdi32, "LineDDA")
	ret, _, _ := syscall.SyscallN(addr, uintptr(xStart), uintptr(yStart), uintptr(xEnd), uintptr(yEnd), lpProc, data)
	return BOOL(ret)
}

func LineTo(hdc HDC, x int32, y int32) BOOL {
	addr := lazyAddr(&pLineTo, libGdi32, "LineTo")
	ret, _, _ := syscall.SyscallN(addr, hdc, uintptr(x), uintptr(y))
	return BOOL(ret)
}

func MaskBlt(hdcDest HDC, xDest int32, yDest int32, width int32, height int32, hdcSrc HDC, xSrc int32, ySrc int32, hbmMask HBITMAP, xMask int32, yMask int32, rop uint32) BOOL {
	addr := lazyAddr(&pMaskBlt, libGdi32, "MaskBlt")
	ret, _, _ := syscall.SyscallN(addr, hdcDest, uintptr(xDest), uintptr(yDest), uintptr(width), uintptr(height), hdcSrc, uintptr(xSrc), uintptr(ySrc), hbmMask, uintptr(xMask), uintptr(yMask), uintptr(rop))
	return BOOL(ret)
}

func PlgBlt(hdcDest HDC, lpPoint *POINT, hdcSrc HDC, xSrc int32, ySrc int32, width int32, height int32, hbmMask HBITMAP, xMask int32, yMask int32) BOOL {
	addr := lazyAddr(&pPlgBlt, libGdi32, "PlgBlt")
	ret, _, _ := syscall.SyscallN(addr, hdcDest, uintptr(unsafe.Pointer(lpPoint)), hdcSrc, uintptr(xSrc), uintptr(ySrc), uintptr(width), uintptr(height), hbmMask, uintptr(xMask), uintptr(yMask))
	return BOOL(ret)
}

func OffsetClipRgn(hdc HDC, x int32, y int32) int32 {
	addr := lazyAddr(&pOffsetClipRgn, libGdi32, "OffsetClipRgn")
	ret, _, _ := syscall.SyscallN(addr, hdc, uintptr(x), uintptr(y))
	return int32(ret)
}

func OffsetRgn(hrgn HRGN, x int32, y int32) int32 {
	addr := lazyAddr(&pOffsetRgn, libGdi32, "OffsetRgn")
	ret, _, _ := syscall.SyscallN(addr, hrgn, uintptr(x), uintptr(y))
	return int32(ret)
}

func PatBlt(hdc HDC, x int32, y int32, w int32, h int32, rop ROP_CODE) BOOL {
	addr := lazyAddr(&pPatBlt, libGdi32, "PatBlt")
	ret, _, _ := syscall.SyscallN(addr, hdc, uintptr(x), uintptr(y), uintptr(w), uintptr(h), uintptr(rop))
	return BOOL(ret)
}

func Pie(hdc HDC, left int32, top int32, right int32, bottom int32, xr1 int32, yr1 int32, xr2 int32, yr2 int32) BOOL {
	addr := lazyAddr(&pPie, libGdi32, "Pie")
	ret, _, _ := syscall.SyscallN(addr, hdc, uintptr(left), uintptr(top), uintptr(right), uintptr(bottom), uintptr(xr1), uintptr(yr1), uintptr(xr2), uintptr(yr2))
	return BOOL(ret)
}

func PlayMetaFile(hdc HDC, hmf HMETAFILE) BOOL {
	addr := lazyAddr(&pPlayMetaFile, libGdi32, "PlayMetaFile")
	ret, _, _ := syscall.SyscallN(addr, hdc, hmf)
	return BOOL(ret)
}

func PaintRgn(hdc HDC, hrgn HRGN) BOOL {
	addr := lazyAddr(&pPaintRgn, libGdi32, "PaintRgn")
	ret, _, _ := syscall.SyscallN(addr, hdc, hrgn)
	return BOOL(ret)
}

func PolyPolygon(hdc HDC, apt *POINT, asz *int32, csz int32) BOOL {
	addr := lazyAddr(&pPolyPolygon, libGdi32, "PolyPolygon")
	ret, _, _ := syscall.SyscallN(addr, hdc, uintptr(unsafe.Pointer(apt)), uintptr(unsafe.Pointer(asz)), uintptr(csz))
	return BOOL(ret)
}

func PtInRegion(hrgn HRGN, x int32, y int32) BOOL {
	addr := lazyAddr(&pPtInRegion, libGdi32, "PtInRegion")
	ret, _, _ := syscall.SyscallN(addr, hrgn, uintptr(x), uintptr(y))
	return BOOL(ret)
}

func PtVisible(hdc HDC, x int32, y int32) BOOL {
	addr := lazyAddr(&pPtVisible, libGdi32, "PtVisible")
	ret, _, _ := syscall.SyscallN(addr, hdc, uintptr(x), uintptr(y))
	return BOOL(ret)
}

func RectInRegion(hrgn HRGN, lprect *RECT) BOOL {
	addr := lazyAddr(&pRectInRegion, libGdi32, "RectInRegion")
	ret, _, _ := syscall.SyscallN(addr, hrgn, uintptr(unsafe.Pointer(lprect)))
	return BOOL(ret)
}

func RectVisible(hdc HDC, lprect *RECT) BOOL {
	addr := lazyAddr(&pRectVisible, libGdi32, "RectVisible")
	ret, _, _ := syscall.SyscallN(addr, hdc, uintptr(unsafe.Pointer(lprect)))
	return BOOL(ret)
}

func Rectangle(hdc HDC, left int32, top int32, right int32, bottom int32) BOOL {
	addr := lazyAddr(&pRectangle, libGdi32, "Rectangle")
	ret, _, _ := syscall.SyscallN(addr, hdc, uintptr(left), uintptr(top), uintptr(right), uintptr(bottom))
	return BOOL(ret)
}

func RestoreDC(hdc HDC, nSavedDC int32) BOOL {
	addr := lazyAddr(&pRestoreDC, libGdi32, "RestoreDC")
	ret, _, _ := syscall.SyscallN(addr, hdc, uintptr(nSavedDC))
	return BOOL(ret)
}

func ResetDCA(hdc HDC, lpdm *DEVMODEA) HDC {
	addr := lazyAddr(&pResetDCA, libGdi32, "ResetDCA")
	ret, _, _ := syscall.SyscallN(addr, hdc, uintptr(unsafe.Pointer(lpdm)))
	return ret
}

var ResetDC = ResetDCW

func ResetDCW(hdc HDC, lpdm *DEVMODEW) HDC {
	addr := lazyAddr(&pResetDCW, libGdi32, "ResetDCW")
	ret, _, _ := syscall.SyscallN(addr, hdc, uintptr(unsafe.Pointer(lpdm)))
	return ret
}

func RealizePalette(hdc HDC) uint32 {
	addr := lazyAddr(&pRealizePalette, libGdi32, "RealizePalette")
	ret, _, _ := syscall.SyscallN(addr, hdc)
	return uint32(ret)
}

func RemoveFontResourceA(lpFileName PSTR) BOOL {
	addr := lazyAddr(&pRemoveFontResourceA, libGdi32, "RemoveFontResourceA")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpFileName)))
	return BOOL(ret)
}

var RemoveFontResource = RemoveFontResourceW

func RemoveFontResourceW(lpFileName PWSTR) BOOL {
	addr := lazyAddr(&pRemoveFontResourceW, libGdi32, "RemoveFontResourceW")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpFileName)))
	return BOOL(ret)
}

func RoundRect(hdc HDC, left int32, top int32, right int32, bottom int32, width int32, height int32) BOOL {
	addr := lazyAddr(&pRoundRect, libGdi32, "RoundRect")
	ret, _, _ := syscall.SyscallN(addr, hdc, uintptr(left), uintptr(top), uintptr(right), uintptr(bottom), uintptr(width), uintptr(height))
	return BOOL(ret)
}

func ResizePalette(hpal HPALETTE, n uint32) BOOL {
	addr := lazyAddr(&pResizePalette, libGdi32, "ResizePalette")
	ret, _, _ := syscall.SyscallN(addr, hpal, uintptr(n))
	return BOOL(ret)
}

func SaveDC(hdc HDC) int32 {
	addr := lazyAddr(&pSaveDC, libGdi32, "SaveDC")
	ret, _, _ := syscall.SyscallN(addr, hdc)
	return int32(ret)
}

func SelectClipRgn(hdc HDC, hrgn HRGN) int32 {
	addr := lazyAddr(&pSelectClipRgn, libGdi32, "SelectClipRgn")
	ret, _, _ := syscall.SyscallN(addr, hdc, hrgn)
	return int32(ret)
}

func ExtSelectClipRgn(hdc HDC, hrgn HRGN, mode RGN_COMBINE_MODE) int32 {
	addr := lazyAddr(&pExtSelectClipRgn, libGdi32, "ExtSelectClipRgn")
	ret, _, _ := syscall.SyscallN(addr, hdc, hrgn, uintptr(mode))
	return int32(ret)
}

func SetMetaRgn(hdc HDC) int32 {
	addr := lazyAddr(&pSetMetaRgn, libGdi32, "SetMetaRgn")
	ret, _, _ := syscall.SyscallN(addr, hdc)
	return int32(ret)
}

func SelectObject(hdc HDC, h HGDIOBJ) HGDIOBJ {
	addr := lazyAddr(&pSelectObject, libGdi32, "SelectObject")
	ret, _, _ := syscall.SyscallN(addr, hdc, h)
	return ret
}

func SelectPalette(hdc HDC, hPal HPALETTE, bForceBkgd BOOL) HPALETTE {
	addr := lazyAddr(&pSelectPalette, libGdi32, "SelectPalette")
	ret, _, _ := syscall.SyscallN(addr, hdc, hPal, uintptr(bForceBkgd))
	return ret
}

func SetBkColor(hdc HDC, color uint32) uint32 {
	addr := lazyAddr(&pSetBkColor, libGdi32, "SetBkColor")
	ret, _, _ := syscall.SyscallN(addr, hdc, uintptr(color))
	return uint32(ret)
}

func SetDCBrushColor(hdc HDC, color uint32) uint32 {
	addr := lazyAddr(&pSetDCBrushColor, libGdi32, "SetDCBrushColor")
	ret, _, _ := syscall.SyscallN(addr, hdc, uintptr(color))
	return uint32(ret)
}

func SetDCPenColor(hdc HDC, color uint32) uint32 {
	addr := lazyAddr(&pSetDCPenColor, libGdi32, "SetDCPenColor")
	ret, _, _ := syscall.SyscallN(addr, hdc, uintptr(color))
	return uint32(ret)
}

func SetBkMode(hdc HDC, mode BACKGROUND_MODE) int32 {
	addr := lazyAddr(&pSetBkMode, libGdi32, "SetBkMode")
	ret, _, _ := syscall.SyscallN(addr, hdc, uintptr(mode))
	return int32(ret)
}

func SetBitmapBits(hbm HBITMAP, cb uint32, pvBits unsafe.Pointer) int32 {
	addr := lazyAddr(&pSetBitmapBits, libGdi32, "SetBitmapBits")
	ret, _, _ := syscall.SyscallN(addr, hbm, uintptr(cb), uintptr(pvBits))
	return int32(ret)
}

func SetBoundsRect(hdc HDC, lprect *RECT, flags SET_BOUNDS_RECT_FLAGS) uint32 {
	addr := lazyAddr(&pSetBoundsRect, libGdi32, "SetBoundsRect")
	ret, _, _ := syscall.SyscallN(addr, hdc, uintptr(unsafe.Pointer(lprect)), uintptr(flags))
	return uint32(ret)
}

func SetDIBits(hdc HDC, hbm HBITMAP, start uint32, cLines uint32, lpBits unsafe.Pointer, lpbmi *BITMAPINFO, ColorUse DIB_USAGE) int32 {
	addr := lazyAddr(&pSetDIBits, libGdi32, "SetDIBits")
	ret, _, _ := syscall.SyscallN(addr, hdc, hbm, uintptr(start), uintptr(cLines), uintptr(lpBits), uintptr(unsafe.Pointer(lpbmi)), uintptr(ColorUse))
	return int32(ret)
}

func SetDIBitsToDevice(hdc HDC, xDest int32, yDest int32, w uint32, h uint32, xSrc int32, ySrc int32, StartScan uint32, cLines uint32, lpvBits unsafe.Pointer, lpbmi *BITMAPINFO, ColorUse DIB_USAGE) int32 {
	addr := lazyAddr(&pSetDIBitsToDevice, libGdi32, "SetDIBitsToDevice")
	ret, _, _ := syscall.SyscallN(addr, hdc, uintptr(xDest), uintptr(yDest), uintptr(w), uintptr(h), uintptr(xSrc), uintptr(ySrc), uintptr(StartScan), uintptr(cLines), uintptr(lpvBits), uintptr(unsafe.Pointer(lpbmi)), uintptr(ColorUse))
	return int32(ret)
}

func SetMapperFlags(hdc HDC, flags uint32) uint32 {
	addr := lazyAddr(&pSetMapperFlags, libGdi32, "SetMapperFlags")
	ret, _, _ := syscall.SyscallN(addr, hdc, uintptr(flags))
	return uint32(ret)
}

func SetGraphicsMode(hdc HDC, iMode GRAPHICS_MODE) int32 {
	addr := lazyAddr(&pSetGraphicsMode, libGdi32, "SetGraphicsMode")
	ret, _, _ := syscall.SyscallN(addr, hdc, uintptr(iMode))
	return int32(ret)
}

func SetMapMode(hdc HDC, iMode HDC_MAP_MODE) int32 {
	addr := lazyAddr(&pSetMapMode, libGdi32, "SetMapMode")
	ret, _, _ := syscall.SyscallN(addr, hdc, uintptr(iMode))
	return int32(ret)
}

func SetLayout(hdc HDC, l DC_LAYOUT) uint32 {
	addr := lazyAddr(&pSetLayout, libGdi32, "SetLayout")
	ret, _, _ := syscall.SyscallN(addr, hdc, uintptr(l))
	return uint32(ret)
}

func GetLayout(hdc HDC) (uint32, WIN32_ERROR) {
	addr := lazyAddr(&pGetLayout, libGdi32, "GetLayout")
	ret, _, err := syscall.SyscallN(addr, hdc)
	return uint32(ret), WIN32_ERROR(err)
}

func SetMetaFileBitsEx(cbBuffer uint32, lpData *byte) HMETAFILE {
	addr := lazyAddr(&pSetMetaFileBitsEx, libGdi32, "SetMetaFileBitsEx")
	ret, _, _ := syscall.SyscallN(addr, uintptr(cbBuffer), uintptr(unsafe.Pointer(lpData)))
	return ret
}

func SetPaletteEntries(hpal HPALETTE, iStart uint32, cEntries uint32, pPalEntries *PALETTEENTRY) uint32 {
	addr := lazyAddr(&pSetPaletteEntries, libGdi32, "SetPaletteEntries")
	ret, _, _ := syscall.SyscallN(addr, hpal, uintptr(iStart), uintptr(cEntries), uintptr(unsafe.Pointer(pPalEntries)))
	return uint32(ret)
}

func SetPixel(hdc HDC, x int32, y int32, color uint32) uint32 {
	addr := lazyAddr(&pSetPixel, libGdi32, "SetPixel")
	ret, _, _ := syscall.SyscallN(addr, hdc, uintptr(x), uintptr(y), uintptr(color))
	return uint32(ret)
}

func SetPixelV(hdc HDC, x int32, y int32, color uint32) BOOL {
	addr := lazyAddr(&pSetPixelV, libGdi32, "SetPixelV")
	ret, _, _ := syscall.SyscallN(addr, hdc, uintptr(x), uintptr(y), uintptr(color))
	return BOOL(ret)
}

func SetPolyFillMode(hdc HDC, mode CREATE_POLYGON_RGN_MODE) int32 {
	addr := lazyAddr(&pSetPolyFillMode, libGdi32, "SetPolyFillMode")
	ret, _, _ := syscall.SyscallN(addr, hdc, uintptr(mode))
	return int32(ret)
}

func StretchBlt(hdcDest HDC, xDest int32, yDest int32, wDest int32, hDest int32, hdcSrc HDC, xSrc int32, ySrc int32, wSrc int32, hSrc int32, rop ROP_CODE) BOOL {
	addr := lazyAddr(&pStretchBlt, libGdi32, "StretchBlt")
	ret, _, _ := syscall.SyscallN(addr, hdcDest, uintptr(xDest), uintptr(yDest), uintptr(wDest), uintptr(hDest), hdcSrc, uintptr(xSrc), uintptr(ySrc), uintptr(wSrc), uintptr(hSrc), uintptr(rop))
	return BOOL(ret)
}

func SetRectRgn(hrgn HRGN, left int32, top int32, right int32, bottom int32) BOOL {
	addr := lazyAddr(&pSetRectRgn, libGdi32, "SetRectRgn")
	ret, _, _ := syscall.SyscallN(addr, hrgn, uintptr(left), uintptr(top), uintptr(right), uintptr(bottom))
	return BOOL(ret)
}

func StretchDIBits(hdc HDC, xDest int32, yDest int32, DestWidth int32, DestHeight int32, xSrc int32, ySrc int32, SrcWidth int32, SrcHeight int32, lpBits unsafe.Pointer, lpbmi *BITMAPINFO, iUsage DIB_USAGE, rop ROP_CODE) int32 {
	addr := lazyAddr(&pStretchDIBits, libGdi32, "StretchDIBits")
	ret, _, _ := syscall.SyscallN(addr, hdc, uintptr(xDest), uintptr(yDest), uintptr(DestWidth), uintptr(DestHeight), uintptr(xSrc), uintptr(ySrc), uintptr(SrcWidth), uintptr(SrcHeight), uintptr(lpBits), uintptr(unsafe.Pointer(lpbmi)), uintptr(iUsage), uintptr(rop))
	return int32(ret)
}

func SetROP2(hdc HDC, rop2 R2_MODE) int32 {
	addr := lazyAddr(&pSetROP2, libGdi32, "SetROP2")
	ret, _, _ := syscall.SyscallN(addr, hdc, uintptr(rop2))
	return int32(ret)
}

func SetStretchBltMode(hdc HDC, mode STRETCH_BLT_MODE) int32 {
	addr := lazyAddr(&pSetStretchBltMode, libGdi32, "SetStretchBltMode")
	ret, _, _ := syscall.SyscallN(addr, hdc, uintptr(mode))
	return int32(ret)
}

func SetSystemPaletteUse(hdc HDC, use SYSTEM_PALETTE_USE) uint32 {
	addr := lazyAddr(&pSetSystemPaletteUse, libGdi32, "SetSystemPaletteUse")
	ret, _, _ := syscall.SyscallN(addr, hdc, uintptr(use))
	return uint32(ret)
}

func SetTextCharacterExtra(hdc HDC, extra int32) int32 {
	addr := lazyAddr(&pSetTextCharacterExtra, libGdi32, "SetTextCharacterExtra")
	ret, _, _ := syscall.SyscallN(addr, hdc, uintptr(extra))
	return int32(ret)
}

func SetTextColor(hdc HDC, color uint32) uint32 {
	addr := lazyAddr(&pSetTextColor, libGdi32, "SetTextColor")
	ret, _, _ := syscall.SyscallN(addr, hdc, uintptr(color))
	return uint32(ret)
}

func SetTextAlign(hdc HDC, align TEXT_ALIGN_OPTIONS) uint32 {
	addr := lazyAddr(&pSetTextAlign, libGdi32, "SetTextAlign")
	ret, _, _ := syscall.SyscallN(addr, hdc, uintptr(align))
	return uint32(ret)
}

func SetTextJustification(hdc HDC, extra int32, count int32) BOOL {
	addr := lazyAddr(&pSetTextJustification, libGdi32, "SetTextJustification")
	ret, _, _ := syscall.SyscallN(addr, hdc, uintptr(extra), uintptr(count))
	return BOOL(ret)
}

func UpdateColors(hdc HDC) BOOL {
	addr := lazyAddr(&pUpdateColors, libGdi32, "UpdateColors")
	ret, _, _ := syscall.SyscallN(addr, hdc)
	return BOOL(ret)
}

func AlphaBlend(hdcDest HDC, xoriginDest int32, yoriginDest int32, wDest int32, hDest int32, hdcSrc HDC, xoriginSrc int32, yoriginSrc int32, wSrc int32, hSrc int32, ftn BLENDFUNCTION) BOOL {
	addr := lazyAddr(&pAlphaBlend, libMsimg32, "AlphaBlend")
	ret, _, _ := syscall.SyscallN(addr, hdcDest, uintptr(xoriginDest), uintptr(yoriginDest), uintptr(wDest), uintptr(hDest), hdcSrc, uintptr(xoriginSrc), uintptr(yoriginSrc), uintptr(wSrc), uintptr(hSrc), *(*uintptr)(unsafe.Pointer(&ftn)))
	return BOOL(ret)
}

func TransparentBlt(hdcDest HDC, xoriginDest int32, yoriginDest int32, wDest int32, hDest int32, hdcSrc HDC, xoriginSrc int32, yoriginSrc int32, wSrc int32, hSrc int32, crTransparent uint32) BOOL {
	addr := lazyAddr(&pTransparentBlt, libMsimg32, "TransparentBlt")
	ret, _, _ := syscall.SyscallN(addr, hdcDest, uintptr(xoriginDest), uintptr(yoriginDest), uintptr(wDest), uintptr(hDest), hdcSrc, uintptr(xoriginSrc), uintptr(yoriginSrc), uintptr(wSrc), uintptr(hSrc), uintptr(crTransparent))
	return BOOL(ret)
}

func GradientFill(hdc HDC, pVertex *TRIVERTEX, nVertex uint32, pMesh unsafe.Pointer, nMesh uint32, ulMode GRADIENT_FILL) BOOL {
	addr := lazyAddr(&pGradientFill, libMsimg32, "GradientFill")
	ret, _, _ := syscall.SyscallN(addr, hdc, uintptr(unsafe.Pointer(pVertex)), uintptr(nVertex), uintptr(pMesh), uintptr(nMesh), uintptr(ulMode))
	return BOOL(ret)
}

func GdiAlphaBlend(hdcDest HDC, xoriginDest int32, yoriginDest int32, wDest int32, hDest int32, hdcSrc HDC, xoriginSrc int32, yoriginSrc int32, wSrc int32, hSrc int32, ftn BLENDFUNCTION) BOOL {
	addr := lazyAddr(&pGdiAlphaBlend, libGdi32, "GdiAlphaBlend")
	ret, _, _ := syscall.SyscallN(addr, hdcDest, uintptr(xoriginDest), uintptr(yoriginDest), uintptr(wDest), uintptr(hDest), hdcSrc, uintptr(xoriginSrc), uintptr(yoriginSrc), uintptr(wSrc), uintptr(hSrc), *(*uintptr)(unsafe.Pointer(&ftn)))
	return BOOL(ret)
}

func GdiTransparentBlt(hdcDest HDC, xoriginDest int32, yoriginDest int32, wDest int32, hDest int32, hdcSrc HDC, xoriginSrc int32, yoriginSrc int32, wSrc int32, hSrc int32, crTransparent uint32) BOOL {
	addr := lazyAddr(&pGdiTransparentBlt, libGdi32, "GdiTransparentBlt")
	ret, _, _ := syscall.SyscallN(addr, hdcDest, uintptr(xoriginDest), uintptr(yoriginDest), uintptr(wDest), uintptr(hDest), hdcSrc, uintptr(xoriginSrc), uintptr(yoriginSrc), uintptr(wSrc), uintptr(hSrc), uintptr(crTransparent))
	return BOOL(ret)
}

func GdiGradientFill(hdc HDC, pVertex *TRIVERTEX, nVertex uint32, pMesh unsafe.Pointer, nCount uint32, ulMode GRADIENT_FILL) BOOL {
	addr := lazyAddr(&pGdiGradientFill, libGdi32, "GdiGradientFill")
	ret, _, _ := syscall.SyscallN(addr, hdc, uintptr(unsafe.Pointer(pVertex)), uintptr(nVertex), uintptr(pMesh), uintptr(nCount), uintptr(ulMode))
	return BOOL(ret)
}

func PlayMetaFileRecord(hdc HDC, lpHandleTable *HANDLETABLE, lpMR *METARECORD, noObjs uint32) BOOL {
	addr := lazyAddr(&pPlayMetaFileRecord, libGdi32, "PlayMetaFileRecord")
	ret, _, _ := syscall.SyscallN(addr, hdc, uintptr(unsafe.Pointer(lpHandleTable)), uintptr(unsafe.Pointer(lpMR)), uintptr(noObjs))
	return BOOL(ret)
}

func EnumMetaFile(hdc HDC, hmf HMETAFILE, proc MFENUMPROC, param3 LPARAM) BOOL {
	addr := lazyAddr(&pEnumMetaFile, libGdi32, "EnumMetaFile")
	ret, _, _ := syscall.SyscallN(addr, hdc, hmf, proc, param3)
	return BOOL(ret)
}

func CloseEnhMetaFile(hdc HDC) HENHMETAFILE {
	addr := lazyAddr(&pCloseEnhMetaFile, libGdi32, "CloseEnhMetaFile")
	ret, _, _ := syscall.SyscallN(addr, hdc)
	return ret
}

func CopyEnhMetaFileA(hEnh HENHMETAFILE, lpFileName PSTR) HENHMETAFILE {
	addr := lazyAddr(&pCopyEnhMetaFileA, libGdi32, "CopyEnhMetaFileA")
	ret, _, _ := syscall.SyscallN(addr, hEnh, uintptr(unsafe.Pointer(lpFileName)))
	return ret
}

var CopyEnhMetaFile = CopyEnhMetaFileW

func CopyEnhMetaFileW(hEnh HENHMETAFILE, lpFileName PWSTR) HENHMETAFILE {
	addr := lazyAddr(&pCopyEnhMetaFileW, libGdi32, "CopyEnhMetaFileW")
	ret, _, _ := syscall.SyscallN(addr, hEnh, uintptr(unsafe.Pointer(lpFileName)))
	return ret
}

func CreateEnhMetaFileA(hdc HDC, lpFilename PSTR, lprc *RECT, lpDesc PSTR) HdcMetdataEnhFileHandle {
	addr := lazyAddr(&pCreateEnhMetaFileA, libGdi32, "CreateEnhMetaFileA")
	ret, _, _ := syscall.SyscallN(addr, hdc, uintptr(unsafe.Pointer(lpFilename)), uintptr(unsafe.Pointer(lprc)), uintptr(unsafe.Pointer(lpDesc)))
	return ret
}

var CreateEnhMetaFile = CreateEnhMetaFileW

func CreateEnhMetaFileW(hdc HDC, lpFilename PWSTR, lprc *RECT, lpDesc PWSTR) HdcMetdataEnhFileHandle {
	addr := lazyAddr(&pCreateEnhMetaFileW, libGdi32, "CreateEnhMetaFileW")
	ret, _, _ := syscall.SyscallN(addr, hdc, uintptr(unsafe.Pointer(lpFilename)), uintptr(unsafe.Pointer(lprc)), uintptr(unsafe.Pointer(lpDesc)))
	return ret
}

func DeleteEnhMetaFile(hmf HENHMETAFILE) BOOL {
	addr := lazyAddr(&pDeleteEnhMetaFile, libGdi32, "DeleteEnhMetaFile")
	ret, _, _ := syscall.SyscallN(addr, hmf)
	return BOOL(ret)
}

func EnumEnhMetaFile(hdc HDC, hmf HENHMETAFILE, proc ENHMFENUMPROC, param3 unsafe.Pointer, lpRect *RECT) BOOL {
	addr := lazyAddr(&pEnumEnhMetaFile, libGdi32, "EnumEnhMetaFile")
	ret, _, _ := syscall.SyscallN(addr, hdc, hmf, proc, uintptr(param3), uintptr(unsafe.Pointer(lpRect)))
	return BOOL(ret)
}

func GetEnhMetaFileA(lpName PSTR) HENHMETAFILE {
	addr := lazyAddr(&pGetEnhMetaFileA, libGdi32, "GetEnhMetaFileA")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpName)))
	return ret
}

var GetEnhMetaFile = GetEnhMetaFileW

func GetEnhMetaFileW(lpName PWSTR) HENHMETAFILE {
	addr := lazyAddr(&pGetEnhMetaFileW, libGdi32, "GetEnhMetaFileW")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpName)))
	return ret
}

func GetEnhMetaFileBits(hEMF HENHMETAFILE, nSize uint32, lpData *byte) uint32 {
	addr := lazyAddr(&pGetEnhMetaFileBits, libGdi32, "GetEnhMetaFileBits")
	ret, _, _ := syscall.SyscallN(addr, hEMF, uintptr(nSize), uintptr(unsafe.Pointer(lpData)))
	return uint32(ret)
}

func GetEnhMetaFileDescriptionA(hemf HENHMETAFILE, cchBuffer uint32, lpDescription PSTR) uint32 {
	addr := lazyAddr(&pGetEnhMetaFileDescriptionA, libGdi32, "GetEnhMetaFileDescriptionA")
	ret, _, _ := syscall.SyscallN(addr, hemf, uintptr(cchBuffer), uintptr(unsafe.Pointer(lpDescription)))
	return uint32(ret)
}

var GetEnhMetaFileDescription = GetEnhMetaFileDescriptionW

func GetEnhMetaFileDescriptionW(hemf HENHMETAFILE, cchBuffer uint32, lpDescription PWSTR) uint32 {
	addr := lazyAddr(&pGetEnhMetaFileDescriptionW, libGdi32, "GetEnhMetaFileDescriptionW")
	ret, _, _ := syscall.SyscallN(addr, hemf, uintptr(cchBuffer), uintptr(unsafe.Pointer(lpDescription)))
	return uint32(ret)
}

func GetEnhMetaFileHeader(hemf HENHMETAFILE, nSize uint32, lpEnhMetaHeader *ENHMETAHEADER) uint32 {
	addr := lazyAddr(&pGetEnhMetaFileHeader, libGdi32, "GetEnhMetaFileHeader")
	ret, _, _ := syscall.SyscallN(addr, hemf, uintptr(nSize), uintptr(unsafe.Pointer(lpEnhMetaHeader)))
	return uint32(ret)
}

func GetEnhMetaFilePaletteEntries(hemf HENHMETAFILE, nNumEntries uint32, lpPaletteEntries *PALETTEENTRY) uint32 {
	addr := lazyAddr(&pGetEnhMetaFilePaletteEntries, libGdi32, "GetEnhMetaFilePaletteEntries")
	ret, _, _ := syscall.SyscallN(addr, hemf, uintptr(nNumEntries), uintptr(unsafe.Pointer(lpPaletteEntries)))
	return uint32(ret)
}

func GetWinMetaFileBits(hemf HENHMETAFILE, cbData16 uint32, pData16 *byte, iMapMode int32, hdcRef HDC) uint32 {
	addr := lazyAddr(&pGetWinMetaFileBits, libGdi32, "GetWinMetaFileBits")
	ret, _, _ := syscall.SyscallN(addr, hemf, uintptr(cbData16), uintptr(unsafe.Pointer(pData16)), uintptr(iMapMode), hdcRef)
	return uint32(ret)
}

func PlayEnhMetaFile(hdc HDC, hmf HENHMETAFILE, lprect *RECT) BOOL {
	addr := lazyAddr(&pPlayEnhMetaFile, libGdi32, "PlayEnhMetaFile")
	ret, _, _ := syscall.SyscallN(addr, hdc, hmf, uintptr(unsafe.Pointer(lprect)))
	return BOOL(ret)
}

func PlayEnhMetaFileRecord(hdc HDC, pht *HANDLETABLE, pmr *ENHMETARECORD, cht uint32) BOOL {
	addr := lazyAddr(&pPlayEnhMetaFileRecord, libGdi32, "PlayEnhMetaFileRecord")
	ret, _, _ := syscall.SyscallN(addr, hdc, uintptr(unsafe.Pointer(pht)), uintptr(unsafe.Pointer(pmr)), uintptr(cht))
	return BOOL(ret)
}

func SetEnhMetaFileBits(nSize uint32, pb *byte) HENHMETAFILE {
	addr := lazyAddr(&pSetEnhMetaFileBits, libGdi32, "SetEnhMetaFileBits")
	ret, _, _ := syscall.SyscallN(addr, uintptr(nSize), uintptr(unsafe.Pointer(pb)))
	return ret
}

func GdiComment(hdc HDC, nSize uint32, lpData *byte) BOOL {
	addr := lazyAddr(&pGdiComment, libGdi32, "GdiComment")
	ret, _, _ := syscall.SyscallN(addr, hdc, uintptr(nSize), uintptr(unsafe.Pointer(lpData)))
	return BOOL(ret)
}

func GetTextMetricsA(hdc HDC, lptm *TEXTMETRICA) BOOL {
	addr := lazyAddr(&pGetTextMetricsA, libGdi32, "GetTextMetricsA")
	ret, _, _ := syscall.SyscallN(addr, hdc, uintptr(unsafe.Pointer(lptm)))
	return BOOL(ret)
}

var GetTextMetrics = GetTextMetricsW

func GetTextMetricsW(hdc HDC, lptm *TEXTMETRICW) BOOL {
	addr := lazyAddr(&pGetTextMetricsW, libGdi32, "GetTextMetricsW")
	ret, _, _ := syscall.SyscallN(addr, hdc, uintptr(unsafe.Pointer(lptm)))
	return BOOL(ret)
}

func AngleArc(hdc HDC, x int32, y int32, r uint32, StartAngle float32, SweepAngle float32) BOOL {
	addr := lazyAddr(&pAngleArc, libGdi32, "AngleArc")
	ret, _, _ := syscall.SyscallN(addr, hdc, uintptr(x), uintptr(y), uintptr(r), uintptr(StartAngle), uintptr(SweepAngle))
	return BOOL(ret)
}

func PolyPolyline(hdc HDC, apt *POINT, asz *uint32, csz uint32) BOOL {
	addr := lazyAddr(&pPolyPolyline, libGdi32, "PolyPolyline")
	ret, _, _ := syscall.SyscallN(addr, hdc, uintptr(unsafe.Pointer(apt)), uintptr(unsafe.Pointer(asz)), uintptr(csz))
	return BOOL(ret)
}

func GetWorldTransform(hdc HDC, lpxf *XFORM) BOOL {
	addr := lazyAddr(&pGetWorldTransform, libGdi32, "GetWorldTransform")
	ret, _, _ := syscall.SyscallN(addr, hdc, uintptr(unsafe.Pointer(lpxf)))
	return BOOL(ret)
}

func SetWorldTransform(hdc HDC, lpxf *XFORM) BOOL {
	addr := lazyAddr(&pSetWorldTransform, libGdi32, "SetWorldTransform")
	ret, _, _ := syscall.SyscallN(addr, hdc, uintptr(unsafe.Pointer(lpxf)))
	return BOOL(ret)
}

func ModifyWorldTransform(hdc HDC, lpxf *XFORM, mode MODIFY_WORLD_TRANSFORM_MODE) BOOL {
	addr := lazyAddr(&pModifyWorldTransform, libGdi32, "ModifyWorldTransform")
	ret, _, _ := syscall.SyscallN(addr, hdc, uintptr(unsafe.Pointer(lpxf)), uintptr(mode))
	return BOOL(ret)
}

func CombineTransform(lpxfOut *XFORM, lpxf1 *XFORM, lpxf2 *XFORM) BOOL {
	addr := lazyAddr(&pCombineTransform, libGdi32, "CombineTransform")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpxfOut)), uintptr(unsafe.Pointer(lpxf1)), uintptr(unsafe.Pointer(lpxf2)))
	return BOOL(ret)
}

func CreateDIBSection(hdc HDC, pbmi *BITMAPINFO, usage DIB_USAGE, ppvBits unsafe.Pointer, hSection HANDLE, offset uint32) (HBITMAP, WIN32_ERROR) {
	addr := lazyAddr(&pCreateDIBSection, libGdi32, "CreateDIBSection")
	ret, _, err := syscall.SyscallN(addr, hdc, uintptr(unsafe.Pointer(pbmi)), uintptr(usage), uintptr(ppvBits), hSection, uintptr(offset))
	return ret, WIN32_ERROR(err)
}

func GetDIBColorTable(hdc HDC, iStart uint32, cEntries uint32, prgbq *RGBQUAD) uint32 {
	addr := lazyAddr(&pGetDIBColorTable, libGdi32, "GetDIBColorTable")
	ret, _, _ := syscall.SyscallN(addr, hdc, uintptr(iStart), uintptr(cEntries), uintptr(unsafe.Pointer(prgbq)))
	return uint32(ret)
}

func SetDIBColorTable(hdc HDC, iStart uint32, cEntries uint32, prgbq *RGBQUAD) uint32 {
	addr := lazyAddr(&pSetDIBColorTable, libGdi32, "SetDIBColorTable")
	ret, _, _ := syscall.SyscallN(addr, hdc, uintptr(iStart), uintptr(cEntries), uintptr(unsafe.Pointer(prgbq)))
	return uint32(ret)
}

func SetColorAdjustment(hdc HDC, lpca *COLORADJUSTMENT) BOOL {
	addr := lazyAddr(&pSetColorAdjustment, libGdi32, "SetColorAdjustment")
	ret, _, _ := syscall.SyscallN(addr, hdc, uintptr(unsafe.Pointer(lpca)))
	return BOOL(ret)
}

func GetColorAdjustment(hdc HDC, lpca *COLORADJUSTMENT) BOOL {
	addr := lazyAddr(&pGetColorAdjustment, libGdi32, "GetColorAdjustment")
	ret, _, _ := syscall.SyscallN(addr, hdc, uintptr(unsafe.Pointer(lpca)))
	return BOOL(ret)
}

func CreateHalftonePalette(hdc HDC) HPALETTE {
	addr := lazyAddr(&pCreateHalftonePalette, libGdi32, "CreateHalftonePalette")
	ret, _, _ := syscall.SyscallN(addr, hdc)
	return ret
}

func AbortPath(hdc HDC) BOOL {
	addr := lazyAddr(&pAbortPath, libGdi32, "AbortPath")
	ret, _, _ := syscall.SyscallN(addr, hdc)
	return BOOL(ret)
}

func ArcTo(hdc HDC, left int32, top int32, right int32, bottom int32, xr1 int32, yr1 int32, xr2 int32, yr2 int32) BOOL {
	addr := lazyAddr(&pArcTo, libGdi32, "ArcTo")
	ret, _, _ := syscall.SyscallN(addr, hdc, uintptr(left), uintptr(top), uintptr(right), uintptr(bottom), uintptr(xr1), uintptr(yr1), uintptr(xr2), uintptr(yr2))
	return BOOL(ret)
}

func BeginPath(hdc HDC) BOOL {
	addr := lazyAddr(&pBeginPath, libGdi32, "BeginPath")
	ret, _, _ := syscall.SyscallN(addr, hdc)
	return BOOL(ret)
}

func CloseFigure(hdc HDC) BOOL {
	addr := lazyAddr(&pCloseFigure, libGdi32, "CloseFigure")
	ret, _, _ := syscall.SyscallN(addr, hdc)
	return BOOL(ret)
}

func EndPath(hdc HDC) BOOL {
	addr := lazyAddr(&pEndPath, libGdi32, "EndPath")
	ret, _, _ := syscall.SyscallN(addr, hdc)
	return BOOL(ret)
}

func FillPath(hdc HDC) BOOL {
	addr := lazyAddr(&pFillPath, libGdi32, "FillPath")
	ret, _, _ := syscall.SyscallN(addr, hdc)
	return BOOL(ret)
}

func FlattenPath(hdc HDC) BOOL {
	addr := lazyAddr(&pFlattenPath, libGdi32, "FlattenPath")
	ret, _, _ := syscall.SyscallN(addr, hdc)
	return BOOL(ret)
}

func GetPath(hdc HDC, apt *POINT, aj *byte, cpt int32) int32 {
	addr := lazyAddr(&pGetPath, libGdi32, "GetPath")
	ret, _, _ := syscall.SyscallN(addr, hdc, uintptr(unsafe.Pointer(apt)), uintptr(unsafe.Pointer(aj)), uintptr(cpt))
	return int32(ret)
}

func PathToRegion(hdc HDC) HRGN {
	addr := lazyAddr(&pPathToRegion, libGdi32, "PathToRegion")
	ret, _, _ := syscall.SyscallN(addr, hdc)
	return ret
}

func PolyDraw(hdc HDC, apt *POINT, aj *byte, cpt int32) BOOL {
	addr := lazyAddr(&pPolyDraw, libGdi32, "PolyDraw")
	ret, _, _ := syscall.SyscallN(addr, hdc, uintptr(unsafe.Pointer(apt)), uintptr(unsafe.Pointer(aj)), uintptr(cpt))
	return BOOL(ret)
}

func SelectClipPath(hdc HDC, mode RGN_COMBINE_MODE) BOOL {
	addr := lazyAddr(&pSelectClipPath, libGdi32, "SelectClipPath")
	ret, _, _ := syscall.SyscallN(addr, hdc, uintptr(mode))
	return BOOL(ret)
}

func SetArcDirection(hdc HDC, dir ARC_DIRECTION) int32 {
	addr := lazyAddr(&pSetArcDirection, libGdi32, "SetArcDirection")
	ret, _, _ := syscall.SyscallN(addr, hdc, uintptr(dir))
	return int32(ret)
}

func SetMiterLimit(hdc HDC, limit float32, old *float32) BOOL {
	addr := lazyAddr(&pSetMiterLimit, libGdi32, "SetMiterLimit")
	ret, _, _ := syscall.SyscallN(addr, hdc, uintptr(limit), uintptr(unsafe.Pointer(old)))
	return BOOL(ret)
}

func StrokeAndFillPath(hdc HDC) BOOL {
	addr := lazyAddr(&pStrokeAndFillPath, libGdi32, "StrokeAndFillPath")
	ret, _, _ := syscall.SyscallN(addr, hdc)
	return BOOL(ret)
}

func StrokePath(hdc HDC) BOOL {
	addr := lazyAddr(&pStrokePath, libGdi32, "StrokePath")
	ret, _, _ := syscall.SyscallN(addr, hdc)
	return BOOL(ret)
}

func WidenPath(hdc HDC) BOOL {
	addr := lazyAddr(&pWidenPath, libGdi32, "WidenPath")
	ret, _, _ := syscall.SyscallN(addr, hdc)
	return BOOL(ret)
}

func ExtCreatePen(iPenStyle PEN_STYLE, cWidth uint32, plbrush *LOGBRUSH, cStyle uint32, pstyle *uint32) HPEN {
	addr := lazyAddr(&pExtCreatePen, libGdi32, "ExtCreatePen")
	ret, _, _ := syscall.SyscallN(addr, uintptr(iPenStyle), uintptr(cWidth), uintptr(unsafe.Pointer(plbrush)), uintptr(cStyle), uintptr(unsafe.Pointer(pstyle)))
	return ret
}

func GetMiterLimit(hdc HDC, plimit *float32) BOOL {
	addr := lazyAddr(&pGetMiterLimit, libGdi32, "GetMiterLimit")
	ret, _, _ := syscall.SyscallN(addr, hdc, uintptr(unsafe.Pointer(plimit)))
	return BOOL(ret)
}

func GetArcDirection(hdc HDC) int32 {
	addr := lazyAddr(&pGetArcDirection, libGdi32, "GetArcDirection")
	ret, _, _ := syscall.SyscallN(addr, hdc)
	return int32(ret)
}

var GetObject = GetObjectW

func GetObjectW(h HGDIOBJ, c int32, pv unsafe.Pointer) int32 {
	addr := lazyAddr(&pGetObjectW, libGdi32, "GetObjectW")
	ret, _, _ := syscall.SyscallN(addr, h, uintptr(c), uintptr(pv))
	return int32(ret)
}

func MoveToEx(hdc HDC, x int32, y int32, lppt *POINT) BOOL {
	addr := lazyAddr(&pMoveToEx, libGdi32, "MoveToEx")
	ret, _, _ := syscall.SyscallN(addr, hdc, uintptr(x), uintptr(y), uintptr(unsafe.Pointer(lppt)))
	return BOOL(ret)
}

func TextOutA(hdc HDC, x int32, y int32, lpString PSTR, c int32) BOOL {
	addr := lazyAddr(&pTextOutA, libGdi32, "TextOutA")
	ret, _, _ := syscall.SyscallN(addr, hdc, uintptr(x), uintptr(y), uintptr(unsafe.Pointer(lpString)), uintptr(c))
	return BOOL(ret)
}

var TextOut = TextOutW

func TextOutW(hdc HDC, x int32, y int32, lpString PWSTR, c int32) BOOL {
	addr := lazyAddr(&pTextOutW, libGdi32, "TextOutW")
	ret, _, _ := syscall.SyscallN(addr, hdc, uintptr(x), uintptr(y), uintptr(unsafe.Pointer(lpString)), uintptr(c))
	return BOOL(ret)
}

func ExtTextOutA(hdc HDC, x int32, y int32, options ETO_OPTIONS, lprect *RECT, lpString PSTR, c uint32, lpDx *int32) BOOL {
	addr := lazyAddr(&pExtTextOutA, libGdi32, "ExtTextOutA")
	ret, _, _ := syscall.SyscallN(addr, hdc, uintptr(x), uintptr(y), uintptr(options), uintptr(unsafe.Pointer(lprect)), uintptr(unsafe.Pointer(lpString)), uintptr(c), uintptr(unsafe.Pointer(lpDx)))
	return BOOL(ret)
}

var ExtTextOut = ExtTextOutW

func ExtTextOutW(hdc HDC, x int32, y int32, options ETO_OPTIONS, lprect *RECT, lpString PWSTR, c uint32, lpDx *int32) BOOL {
	addr := lazyAddr(&pExtTextOutW, libGdi32, "ExtTextOutW")
	ret, _, _ := syscall.SyscallN(addr, hdc, uintptr(x), uintptr(y), uintptr(options), uintptr(unsafe.Pointer(lprect)), uintptr(unsafe.Pointer(lpString)), uintptr(c), uintptr(unsafe.Pointer(lpDx)))
	return BOOL(ret)
}

func PolyTextOutA(hdc HDC, ppt *POLYTEXTA, nstrings int32) BOOL {
	addr := lazyAddr(&pPolyTextOutA, libGdi32, "PolyTextOutA")
	ret, _, _ := syscall.SyscallN(addr, hdc, uintptr(unsafe.Pointer(ppt)), uintptr(nstrings))
	return BOOL(ret)
}

var PolyTextOut = PolyTextOutW

func PolyTextOutW(hdc HDC, ppt *POLYTEXTW, nstrings int32) BOOL {
	addr := lazyAddr(&pPolyTextOutW, libGdi32, "PolyTextOutW")
	ret, _, _ := syscall.SyscallN(addr, hdc, uintptr(unsafe.Pointer(ppt)), uintptr(nstrings))
	return BOOL(ret)
}

func CreatePolygonRgn(pptl *POINT, cPoint int32, iMode CREATE_POLYGON_RGN_MODE) HRGN {
	addr := lazyAddr(&pCreatePolygonRgn, libGdi32, "CreatePolygonRgn")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pptl)), uintptr(cPoint), uintptr(iMode))
	return ret
}

func DPtoLP(hdc HDC, lppt *POINT, c int32) BOOL {
	addr := lazyAddr(&pDPtoLP, libGdi32, "DPtoLP")
	ret, _, _ := syscall.SyscallN(addr, hdc, uintptr(unsafe.Pointer(lppt)), uintptr(c))
	return BOOL(ret)
}

func LPtoDP(hdc HDC, lppt *POINT, c int32) BOOL {
	addr := lazyAddr(&pLPtoDP, libGdi32, "LPtoDP")
	ret, _, _ := syscall.SyscallN(addr, hdc, uintptr(unsafe.Pointer(lppt)), uintptr(c))
	return BOOL(ret)
}

func Polygon(hdc HDC, apt *POINT, cpt int32) BOOL {
	addr := lazyAddr(&pPolygon, libGdi32, "Polygon")
	ret, _, _ := syscall.SyscallN(addr, hdc, uintptr(unsafe.Pointer(apt)), uintptr(cpt))
	return BOOL(ret)
}

func Polyline(hdc HDC, apt *POINT, cpt int32) BOOL {
	addr := lazyAddr(&pPolyline, libGdi32, "Polyline")
	ret, _, _ := syscall.SyscallN(addr, hdc, uintptr(unsafe.Pointer(apt)), uintptr(cpt))
	return BOOL(ret)
}

func PolyBezier(hdc HDC, apt *POINT, cpt uint32) BOOL {
	addr := lazyAddr(&pPolyBezier, libGdi32, "PolyBezier")
	ret, _, _ := syscall.SyscallN(addr, hdc, uintptr(unsafe.Pointer(apt)), uintptr(cpt))
	return BOOL(ret)
}

func PolyBezierTo(hdc HDC, apt *POINT, cpt uint32) BOOL {
	addr := lazyAddr(&pPolyBezierTo, libGdi32, "PolyBezierTo")
	ret, _, _ := syscall.SyscallN(addr, hdc, uintptr(unsafe.Pointer(apt)), uintptr(cpt))
	return BOOL(ret)
}

func PolylineTo(hdc HDC, apt *POINT, cpt uint32) BOOL {
	addr := lazyAddr(&pPolylineTo, libGdi32, "PolylineTo")
	ret, _, _ := syscall.SyscallN(addr, hdc, uintptr(unsafe.Pointer(apt)), uintptr(cpt))
	return BOOL(ret)
}

func SetViewportExtEx(hdc HDC, x int32, y int32, lpsz *SIZE) BOOL {
	addr := lazyAddr(&pSetViewportExtEx, libGdi32, "SetViewportExtEx")
	ret, _, _ := syscall.SyscallN(addr, hdc, uintptr(x), uintptr(y), uintptr(unsafe.Pointer(lpsz)))
	return BOOL(ret)
}

func SetViewportOrgEx(hdc HDC, x int32, y int32, lppt *POINT) BOOL {
	addr := lazyAddr(&pSetViewportOrgEx, libGdi32, "SetViewportOrgEx")
	ret, _, _ := syscall.SyscallN(addr, hdc, uintptr(x), uintptr(y), uintptr(unsafe.Pointer(lppt)))
	return BOOL(ret)
}

func SetWindowExtEx(hdc HDC, x int32, y int32, lpsz *SIZE) BOOL {
	addr := lazyAddr(&pSetWindowExtEx, libGdi32, "SetWindowExtEx")
	ret, _, _ := syscall.SyscallN(addr, hdc, uintptr(x), uintptr(y), uintptr(unsafe.Pointer(lpsz)))
	return BOOL(ret)
}

func SetWindowOrgEx(hdc HDC, x int32, y int32, lppt *POINT) BOOL {
	addr := lazyAddr(&pSetWindowOrgEx, libGdi32, "SetWindowOrgEx")
	ret, _, _ := syscall.SyscallN(addr, hdc, uintptr(x), uintptr(y), uintptr(unsafe.Pointer(lppt)))
	return BOOL(ret)
}

func OffsetViewportOrgEx(hdc HDC, x int32, y int32, lppt *POINT) BOOL {
	addr := lazyAddr(&pOffsetViewportOrgEx, libGdi32, "OffsetViewportOrgEx")
	ret, _, _ := syscall.SyscallN(addr, hdc, uintptr(x), uintptr(y), uintptr(unsafe.Pointer(lppt)))
	return BOOL(ret)
}

func OffsetWindowOrgEx(hdc HDC, x int32, y int32, lppt *POINT) BOOL {
	addr := lazyAddr(&pOffsetWindowOrgEx, libGdi32, "OffsetWindowOrgEx")
	ret, _, _ := syscall.SyscallN(addr, hdc, uintptr(x), uintptr(y), uintptr(unsafe.Pointer(lppt)))
	return BOOL(ret)
}

func ScaleViewportExtEx(hdc HDC, xn int32, dx int32, yn int32, yd int32, lpsz *SIZE) BOOL {
	addr := lazyAddr(&pScaleViewportExtEx, libGdi32, "ScaleViewportExtEx")
	ret, _, _ := syscall.SyscallN(addr, hdc, uintptr(xn), uintptr(dx), uintptr(yn), uintptr(yd), uintptr(unsafe.Pointer(lpsz)))
	return BOOL(ret)
}

func ScaleWindowExtEx(hdc HDC, xn int32, xd int32, yn int32, yd int32, lpsz *SIZE) BOOL {
	addr := lazyAddr(&pScaleWindowExtEx, libGdi32, "ScaleWindowExtEx")
	ret, _, _ := syscall.SyscallN(addr, hdc, uintptr(xn), uintptr(xd), uintptr(yn), uintptr(yd), uintptr(unsafe.Pointer(lpsz)))
	return BOOL(ret)
}

func SetBitmapDimensionEx(hbm HBITMAP, w int32, h int32, lpsz *SIZE) BOOL {
	addr := lazyAddr(&pSetBitmapDimensionEx, libGdi32, "SetBitmapDimensionEx")
	ret, _, _ := syscall.SyscallN(addr, hbm, uintptr(w), uintptr(h), uintptr(unsafe.Pointer(lpsz)))
	return BOOL(ret)
}

func SetBrushOrgEx(hdc HDC, x int32, y int32, lppt *POINT) BOOL {
	addr := lazyAddr(&pSetBrushOrgEx, libGdi32, "SetBrushOrgEx")
	ret, _, _ := syscall.SyscallN(addr, hdc, uintptr(x), uintptr(y), uintptr(unsafe.Pointer(lppt)))
	return BOOL(ret)
}

func GetTextFaceA(hdc HDC, c int32, lpName PSTR) int32 {
	addr := lazyAddr(&pGetTextFaceA, libGdi32, "GetTextFaceA")
	ret, _, _ := syscall.SyscallN(addr, hdc, uintptr(c), uintptr(unsafe.Pointer(lpName)))
	return int32(ret)
}

var GetTextFace = GetTextFaceW

func GetTextFaceW(hdc HDC, c int32, lpName PWSTR) int32 {
	addr := lazyAddr(&pGetTextFaceW, libGdi32, "GetTextFaceW")
	ret, _, _ := syscall.SyscallN(addr, hdc, uintptr(c), uintptr(unsafe.Pointer(lpName)))
	return int32(ret)
}

func GetKerningPairsA(hdc HDC, nPairs uint32, lpKernPair *KERNINGPAIR) uint32 {
	addr := lazyAddr(&pGetKerningPairsA, libGdi32, "GetKerningPairsA")
	ret, _, _ := syscall.SyscallN(addr, hdc, uintptr(nPairs), uintptr(unsafe.Pointer(lpKernPair)))
	return uint32(ret)
}

var GetKerningPairs = GetKerningPairsW

func GetKerningPairsW(hdc HDC, nPairs uint32, lpKernPair *KERNINGPAIR) uint32 {
	addr := lazyAddr(&pGetKerningPairsW, libGdi32, "GetKerningPairsW")
	ret, _, _ := syscall.SyscallN(addr, hdc, uintptr(nPairs), uintptr(unsafe.Pointer(lpKernPair)))
	return uint32(ret)
}

func GetDCOrgEx(hdc HDC, lppt *POINT) BOOL {
	addr := lazyAddr(&pGetDCOrgEx, libGdi32, "GetDCOrgEx")
	ret, _, _ := syscall.SyscallN(addr, hdc, uintptr(unsafe.Pointer(lppt)))
	return BOOL(ret)
}

func FixBrushOrgEx(hdc HDC, x int32, y int32, ptl *POINT) BOOL {
	addr := lazyAddr(&pFixBrushOrgEx, libGdi32, "FixBrushOrgEx")
	ret, _, _ := syscall.SyscallN(addr, hdc, uintptr(x), uintptr(y), uintptr(unsafe.Pointer(ptl)))
	return BOOL(ret)
}

func UnrealizeObject(h HGDIOBJ) BOOL {
	addr := lazyAddr(&pUnrealizeObject, libGdi32, "UnrealizeObject")
	ret, _, _ := syscall.SyscallN(addr, h)
	return BOOL(ret)
}

func GdiFlush() BOOL {
	addr := lazyAddr(&pGdiFlush, libGdi32, "GdiFlush")
	ret, _, _ := syscall.SyscallN(addr)
	return BOOL(ret)
}

func GdiSetBatchLimit(dw uint32) uint32 {
	addr := lazyAddr(&pGdiSetBatchLimit, libGdi32, "GdiSetBatchLimit")
	ret, _, _ := syscall.SyscallN(addr, uintptr(dw))
	return uint32(ret)
}

func GdiGetBatchLimit() uint32 {
	addr := lazyAddr(&pGdiGetBatchLimit, libGdi32, "GdiGetBatchLimit")
	ret, _, _ := syscall.SyscallN(addr)
	return uint32(ret)
}

func DrawEdge(hdc HDC, qrc *RECT, edge DRAWEDGE_FLAGS, grfFlags DRAW_EDGE_FLAGS) BOOL {
	addr := lazyAddr(&pDrawEdge, libUser32, "DrawEdge")
	ret, _, _ := syscall.SyscallN(addr, hdc, uintptr(unsafe.Pointer(qrc)), uintptr(edge), uintptr(grfFlags))
	return BOOL(ret)
}

func DrawFrameControl(param0 HDC, param1 *RECT, param2 DFC_TYPE, param3 DFCS_STATE) BOOL {
	addr := lazyAddr(&pDrawFrameControl, libUser32, "DrawFrameControl")
	ret, _, _ := syscall.SyscallN(addr, param0, uintptr(unsafe.Pointer(param1)), uintptr(param2), uintptr(param3))
	return BOOL(ret)
}

func DrawCaption(hwnd HWND, hdc HDC, lprect *RECT, flags DRAW_CAPTION_FLAGS) BOOL {
	addr := lazyAddr(&pDrawCaption, libUser32, "DrawCaption")
	ret, _, _ := syscall.SyscallN(addr, hwnd, hdc, uintptr(unsafe.Pointer(lprect)), uintptr(flags))
	return BOOL(ret)
}

func DrawAnimatedRects(hwnd HWND, idAni int32, lprcFrom *RECT, lprcTo *RECT) BOOL {
	addr := lazyAddr(&pDrawAnimatedRects, libUser32, "DrawAnimatedRects")
	ret, _, _ := syscall.SyscallN(addr, hwnd, uintptr(idAni), uintptr(unsafe.Pointer(lprcFrom)), uintptr(unsafe.Pointer(lprcTo)))
	return BOOL(ret)
}

func DrawTextA(hdc HDC, lpchText PSTR, cchText int32, lprc *RECT, format DRAW_TEXT_FORMAT) int32 {
	addr := lazyAddr(&pDrawTextA, libUser32, "DrawTextA")
	ret, _, _ := syscall.SyscallN(addr, hdc, uintptr(unsafe.Pointer(lpchText)), uintptr(cchText), uintptr(unsafe.Pointer(lprc)), uintptr(format))
	return int32(ret)
}

var DrawText = DrawTextW

func DrawTextW(hdc HDC, lpchText PWSTR, cchText int32, lprc *RECT, format DRAW_TEXT_FORMAT) int32 {
	addr := lazyAddr(&pDrawTextW, libUser32, "DrawTextW")
	ret, _, _ := syscall.SyscallN(addr, hdc, uintptr(unsafe.Pointer(lpchText)), uintptr(cchText), uintptr(unsafe.Pointer(lprc)), uintptr(format))
	return int32(ret)
}

func DrawTextExA(hdc HDC, lpchText PSTR, cchText int32, lprc *RECT, format DRAW_TEXT_FORMAT, lpdtp *DRAWTEXTPARAMS) int32 {
	addr := lazyAddr(&pDrawTextExA, libUser32, "DrawTextExA")
	ret, _, _ := syscall.SyscallN(addr, hdc, uintptr(unsafe.Pointer(lpchText)), uintptr(cchText), uintptr(unsafe.Pointer(lprc)), uintptr(format), uintptr(unsafe.Pointer(lpdtp)))
	return int32(ret)
}

var DrawTextEx = DrawTextExW

func DrawTextExW(hdc HDC, lpchText PWSTR, cchText int32, lprc *RECT, format DRAW_TEXT_FORMAT, lpdtp *DRAWTEXTPARAMS) int32 {
	addr := lazyAddr(&pDrawTextExW, libUser32, "DrawTextExW")
	ret, _, _ := syscall.SyscallN(addr, hdc, uintptr(unsafe.Pointer(lpchText)), uintptr(cchText), uintptr(unsafe.Pointer(lprc)), uintptr(format), uintptr(unsafe.Pointer(lpdtp)))
	return int32(ret)
}

func GrayStringA(hDC HDC, hBrush HBRUSH, lpOutputFunc GRAYSTRINGPROC, lpData LPARAM, nCount int32, X int32, Y int32, nWidth int32, nHeight int32) BOOL {
	addr := lazyAddr(&pGrayStringA, libUser32, "GrayStringA")
	ret, _, _ := syscall.SyscallN(addr, hDC, hBrush, lpOutputFunc, lpData, uintptr(nCount), uintptr(X), uintptr(Y), uintptr(nWidth), uintptr(nHeight))
	return BOOL(ret)
}

var GrayString = GrayStringW

func GrayStringW(hDC HDC, hBrush HBRUSH, lpOutputFunc GRAYSTRINGPROC, lpData LPARAM, nCount int32, X int32, Y int32, nWidth int32, nHeight int32) BOOL {
	addr := lazyAddr(&pGrayStringW, libUser32, "GrayStringW")
	ret, _, _ := syscall.SyscallN(addr, hDC, hBrush, lpOutputFunc, lpData, uintptr(nCount), uintptr(X), uintptr(Y), uintptr(nWidth), uintptr(nHeight))
	return BOOL(ret)
}

func DrawStateA(hdc HDC, hbrFore HBRUSH, qfnCallBack DRAWSTATEPROC, lData LPARAM, wData WPARAM, x int32, y int32, cx int32, cy int32, uFlags DRAWSTATE_FLAGS) BOOL {
	addr := lazyAddr(&pDrawStateA, libUser32, "DrawStateA")
	ret, _, _ := syscall.SyscallN(addr, hdc, hbrFore, qfnCallBack, lData, wData, uintptr(x), uintptr(y), uintptr(cx), uintptr(cy), uintptr(uFlags))
	return BOOL(ret)
}

var DrawState = DrawStateW

func DrawStateW(hdc HDC, hbrFore HBRUSH, qfnCallBack DRAWSTATEPROC, lData LPARAM, wData WPARAM, x int32, y int32, cx int32, cy int32, uFlags DRAWSTATE_FLAGS) BOOL {
	addr := lazyAddr(&pDrawStateW, libUser32, "DrawStateW")
	ret, _, _ := syscall.SyscallN(addr, hdc, hbrFore, qfnCallBack, lData, wData, uintptr(x), uintptr(y), uintptr(cx), uintptr(cy), uintptr(uFlags))
	return BOOL(ret)
}

func TabbedTextOutA(hdc HDC, x int32, y int32, lpString PSTR, chCount int32, nTabPositions int32, lpnTabStopPositions *int32, nTabOrigin int32) int32 {
	addr := lazyAddr(&pTabbedTextOutA, libUser32, "TabbedTextOutA")
	ret, _, _ := syscall.SyscallN(addr, hdc, uintptr(x), uintptr(y), uintptr(unsafe.Pointer(lpString)), uintptr(chCount), uintptr(nTabPositions), uintptr(unsafe.Pointer(lpnTabStopPositions)), uintptr(nTabOrigin))
	return int32(ret)
}

var TabbedTextOut = TabbedTextOutW

func TabbedTextOutW(hdc HDC, x int32, y int32, lpString PWSTR, chCount int32, nTabPositions int32, lpnTabStopPositions *int32, nTabOrigin int32) int32 {
	addr := lazyAddr(&pTabbedTextOutW, libUser32, "TabbedTextOutW")
	ret, _, _ := syscall.SyscallN(addr, hdc, uintptr(x), uintptr(y), uintptr(unsafe.Pointer(lpString)), uintptr(chCount), uintptr(nTabPositions), uintptr(unsafe.Pointer(lpnTabStopPositions)), uintptr(nTabOrigin))
	return int32(ret)
}

func GetTabbedTextExtentA(hdc HDC, lpString PSTR, chCount int32, nTabPositions int32, lpnTabStopPositions *int32) uint32 {
	addr := lazyAddr(&pGetTabbedTextExtentA, libUser32, "GetTabbedTextExtentA")
	ret, _, _ := syscall.SyscallN(addr, hdc, uintptr(unsafe.Pointer(lpString)), uintptr(chCount), uintptr(nTabPositions), uintptr(unsafe.Pointer(lpnTabStopPositions)))
	return uint32(ret)
}

var GetTabbedTextExtent = GetTabbedTextExtentW

func GetTabbedTextExtentW(hdc HDC, lpString PWSTR, chCount int32, nTabPositions int32, lpnTabStopPositions *int32) uint32 {
	addr := lazyAddr(&pGetTabbedTextExtentW, libUser32, "GetTabbedTextExtentW")
	ret, _, _ := syscall.SyscallN(addr, hdc, uintptr(unsafe.Pointer(lpString)), uintptr(chCount), uintptr(nTabPositions), uintptr(unsafe.Pointer(lpnTabStopPositions)))
	return uint32(ret)
}

func UpdateWindow(hWnd HWND) BOOL {
	addr := lazyAddr(&pUpdateWindow, libUser32, "UpdateWindow")
	ret, _, _ := syscall.SyscallN(addr, hWnd)
	return BOOL(ret)
}

func PaintDesktop(hdc HDC) BOOL {
	addr := lazyAddr(&pPaintDesktop, libUser32, "PaintDesktop")
	ret, _, _ := syscall.SyscallN(addr, hdc)
	return BOOL(ret)
}

func WindowFromDC(hDC HDC) HWND {
	addr := lazyAddr(&pWindowFromDC, libUser32, "WindowFromDC")
	ret, _, _ := syscall.SyscallN(addr, hDC)
	return ret
}

func GetDC(hWnd HWND) HDC {
	addr := lazyAddr(&pGetDC, libUser32, "GetDC")
	ret, _, _ := syscall.SyscallN(addr, hWnd)
	return ret
}

func GetDCEx(hWnd HWND, hrgnClip HRGN, flags GET_DCX_FLAGS) HDC {
	addr := lazyAddr(&pGetDCEx, libUser32, "GetDCEx")
	ret, _, _ := syscall.SyscallN(addr, hWnd, hrgnClip, uintptr(flags))
	return ret
}

func GetWindowDC(hWnd HWND) HDC {
	addr := lazyAddr(&pGetWindowDC, libUser32, "GetWindowDC")
	ret, _, _ := syscall.SyscallN(addr, hWnd)
	return ret
}

func ReleaseDC(hWnd HWND, hDC HDC) int32 {
	addr := lazyAddr(&pReleaseDC, libUser32, "ReleaseDC")
	ret, _, _ := syscall.SyscallN(addr, hWnd, hDC)
	return int32(ret)
}

func BeginPaint(hWnd HWND, lpPaint *PAINTSTRUCT) HDC {
	addr := lazyAddr(&pBeginPaint, libUser32, "BeginPaint")
	ret, _, _ := syscall.SyscallN(addr, hWnd, uintptr(unsafe.Pointer(lpPaint)))
	return ret
}

func EndPaint(hWnd HWND, lpPaint *PAINTSTRUCT) BOOL {
	addr := lazyAddr(&pEndPaint, libUser32, "EndPaint")
	ret, _, _ := syscall.SyscallN(addr, hWnd, uintptr(unsafe.Pointer(lpPaint)))
	return BOOL(ret)
}

func GetUpdateRect(hWnd HWND, lpRect *RECT, bErase BOOL) BOOL {
	addr := lazyAddr(&pGetUpdateRect, libUser32, "GetUpdateRect")
	ret, _, _ := syscall.SyscallN(addr, hWnd, uintptr(unsafe.Pointer(lpRect)), uintptr(bErase))
	return BOOL(ret)
}

func GetUpdateRgn(hWnd HWND, hRgn HRGN, bErase BOOL) int32 {
	addr := lazyAddr(&pGetUpdateRgn, libUser32, "GetUpdateRgn")
	ret, _, _ := syscall.SyscallN(addr, hWnd, hRgn, uintptr(bErase))
	return int32(ret)
}

func SetWindowRgn(hWnd HWND, hRgn HRGN, bRedraw BOOL) int32 {
	addr := lazyAddr(&pSetWindowRgn, libUser32, "SetWindowRgn")
	ret, _, _ := syscall.SyscallN(addr, hWnd, hRgn, uintptr(bRedraw))
	return int32(ret)
}

func GetWindowRgn(hWnd HWND, hRgn HRGN) int32 {
	addr := lazyAddr(&pGetWindowRgn, libUser32, "GetWindowRgn")
	ret, _, _ := syscall.SyscallN(addr, hWnd, hRgn)
	return int32(ret)
}

func GetWindowRgnBox(hWnd HWND, lprc *RECT) int32 {
	addr := lazyAddr(&pGetWindowRgnBox, libUser32, "GetWindowRgnBox")
	ret, _, _ := syscall.SyscallN(addr, hWnd, uintptr(unsafe.Pointer(lprc)))
	return int32(ret)
}

func ExcludeUpdateRgn(hDC HDC, hWnd HWND) int32 {
	addr := lazyAddr(&pExcludeUpdateRgn, libUser32, "ExcludeUpdateRgn")
	ret, _, _ := syscall.SyscallN(addr, hDC, hWnd)
	return int32(ret)
}

func InvalidateRect(hWnd HWND, lpRect *RECT, bErase BOOL) BOOL {
	addr := lazyAddr(&pInvalidateRect, libUser32, "InvalidateRect")
	ret, _, _ := syscall.SyscallN(addr, hWnd, uintptr(unsafe.Pointer(lpRect)), uintptr(bErase))
	return BOOL(ret)
}

func ValidateRect(hWnd HWND, lpRect *RECT) BOOL {
	addr := lazyAddr(&pValidateRect, libUser32, "ValidateRect")
	ret, _, _ := syscall.SyscallN(addr, hWnd, uintptr(unsafe.Pointer(lpRect)))
	return BOOL(ret)
}

func InvalidateRgn(hWnd HWND, hRgn HRGN, bErase BOOL) BOOL {
	addr := lazyAddr(&pInvalidateRgn, libUser32, "InvalidateRgn")
	ret, _, _ := syscall.SyscallN(addr, hWnd, hRgn, uintptr(bErase))
	return BOOL(ret)
}

func ValidateRgn(hWnd HWND, hRgn HRGN) BOOL {
	addr := lazyAddr(&pValidateRgn, libUser32, "ValidateRgn")
	ret, _, _ := syscall.SyscallN(addr, hWnd, hRgn)
	return BOOL(ret)
}

func RedrawWindow(hWnd HWND, lprcUpdate *RECT, hrgnUpdate HRGN, flags REDRAW_WINDOW_FLAGS) BOOL {
	addr := lazyAddr(&pRedrawWindow, libUser32, "RedrawWindow")
	ret, _, _ := syscall.SyscallN(addr, hWnd, uintptr(unsafe.Pointer(lprcUpdate)), hrgnUpdate, uintptr(flags))
	return BOOL(ret)
}

func LockWindowUpdate(hWndLock HWND) BOOL {
	addr := lazyAddr(&pLockWindowUpdate, libUser32, "LockWindowUpdate")
	ret, _, _ := syscall.SyscallN(addr, hWndLock)
	return BOOL(ret)
}

func ClientToScreen(hWnd HWND, lpPoint *POINT) BOOL {
	addr := lazyAddr(&pClientToScreen, libUser32, "ClientToScreen")
	ret, _, _ := syscall.SyscallN(addr, hWnd, uintptr(unsafe.Pointer(lpPoint)))
	return BOOL(ret)
}

func ScreenToClient(hWnd HWND, lpPoint *POINT) BOOL {
	addr := lazyAddr(&pScreenToClient, libUser32, "ScreenToClient")
	ret, _, _ := syscall.SyscallN(addr, hWnd, uintptr(unsafe.Pointer(lpPoint)))
	return BOOL(ret)
}

func MapWindowPoints(hWndFrom HWND, hWndTo HWND, lpPoints *POINT, cPoints uint32) int32 {
	addr := lazyAddr(&pMapWindowPoints, libUser32, "MapWindowPoints")
	ret, _, _ := syscall.SyscallN(addr, hWndFrom, hWndTo, uintptr(unsafe.Pointer(lpPoints)), uintptr(cPoints))
	return int32(ret)
}

func GetSysColorBrush(nIndex int32) HBRUSH {
	addr := lazyAddr(&pGetSysColorBrush, libUser32, "GetSysColorBrush")
	ret, _, _ := syscall.SyscallN(addr, uintptr(nIndex))
	return ret
}

func DrawFocusRect(hDC HDC, lprc *RECT) BOOL {
	addr := lazyAddr(&pDrawFocusRect, libUser32, "DrawFocusRect")
	ret, _, _ := syscall.SyscallN(addr, hDC, uintptr(unsafe.Pointer(lprc)))
	return BOOL(ret)
}

func FillRect(hDC HDC, lprc *RECT, hbr HBRUSH) int32 {
	addr := lazyAddr(&pFillRect, libUser32, "FillRect")
	ret, _, _ := syscall.SyscallN(addr, hDC, uintptr(unsafe.Pointer(lprc)), hbr)
	return int32(ret)
}

func FrameRect(hDC HDC, lprc *RECT, hbr HBRUSH) int32 {
	addr := lazyAddr(&pFrameRect, libUser32, "FrameRect")
	ret, _, _ := syscall.SyscallN(addr, hDC, uintptr(unsafe.Pointer(lprc)), hbr)
	return int32(ret)
}

func InvertRect(hDC HDC, lprc *RECT) BOOL {
	addr := lazyAddr(&pInvertRect, libUser32, "InvertRect")
	ret, _, _ := syscall.SyscallN(addr, hDC, uintptr(unsafe.Pointer(lprc)))
	return BOOL(ret)
}

func SetRect(lprc *RECT, xLeft int32, yTop int32, xRight int32, yBottom int32) BOOL {
	addr := lazyAddr(&pSetRect, libUser32, "SetRect")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lprc)), uintptr(xLeft), uintptr(yTop), uintptr(xRight), uintptr(yBottom))
	return BOOL(ret)
}

func SetRectEmpty(lprc *RECT) BOOL {
	addr := lazyAddr(&pSetRectEmpty, libUser32, "SetRectEmpty")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lprc)))
	return BOOL(ret)
}

func CopyRect(lprcDst *RECT, lprcSrc *RECT) BOOL {
	addr := lazyAddr(&pCopyRect, libUser32, "CopyRect")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lprcDst)), uintptr(unsafe.Pointer(lprcSrc)))
	return BOOL(ret)
}

func InflateRect(lprc *RECT, dx int32, dy int32) BOOL {
	addr := lazyAddr(&pInflateRect, libUser32, "InflateRect")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lprc)), uintptr(dx), uintptr(dy))
	return BOOL(ret)
}

func IntersectRect(lprcDst *RECT, lprcSrc1 *RECT, lprcSrc2 *RECT) BOOL {
	addr := lazyAddr(&pIntersectRect, libUser32, "IntersectRect")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lprcDst)), uintptr(unsafe.Pointer(lprcSrc1)), uintptr(unsafe.Pointer(lprcSrc2)))
	return BOOL(ret)
}

func UnionRect(lprcDst *RECT, lprcSrc1 *RECT, lprcSrc2 *RECT) BOOL {
	addr := lazyAddr(&pUnionRect, libUser32, "UnionRect")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lprcDst)), uintptr(unsafe.Pointer(lprcSrc1)), uintptr(unsafe.Pointer(lprcSrc2)))
	return BOOL(ret)
}

func SubtractRect(lprcDst *RECT, lprcSrc1 *RECT, lprcSrc2 *RECT) BOOL {
	addr := lazyAddr(&pSubtractRect, libUser32, "SubtractRect")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lprcDst)), uintptr(unsafe.Pointer(lprcSrc1)), uintptr(unsafe.Pointer(lprcSrc2)))
	return BOOL(ret)
}

func OffsetRect(lprc *RECT, dx int32, dy int32) BOOL {
	addr := lazyAddr(&pOffsetRect, libUser32, "OffsetRect")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lprc)), uintptr(dx), uintptr(dy))
	return BOOL(ret)
}

func IsRectEmpty(lprc *RECT) BOOL {
	addr := lazyAddr(&pIsRectEmpty, libUser32, "IsRectEmpty")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lprc)))
	return BOOL(ret)
}

func EqualRect(lprc1 *RECT, lprc2 *RECT) BOOL {
	addr := lazyAddr(&pEqualRect, libUser32, "EqualRect")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lprc1)), uintptr(unsafe.Pointer(lprc2)))
	return BOOL(ret)
}

func PtInRect(lprc *RECT, pt POINT) BOOL {
	addr := lazyAddr(&pPtInRect, libUser32, "PtInRect")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lprc)), *(*uintptr)(unsafe.Pointer(&pt)))
	return BOOL(ret)
}

func LoadBitmapA(hInstance HINSTANCE, lpBitmapName PSTR) HBITMAP {
	addr := lazyAddr(&pLoadBitmapA, libUser32, "LoadBitmapA")
	ret, _, _ := syscall.SyscallN(addr, hInstance, uintptr(unsafe.Pointer(lpBitmapName)))
	return ret
}

var LoadBitmap = LoadBitmapW

func LoadBitmapW(hInstance HINSTANCE, lpBitmapName PWSTR) HBITMAP {
	addr := lazyAddr(&pLoadBitmapW, libUser32, "LoadBitmapW")
	ret, _, _ := syscall.SyscallN(addr, hInstance, uintptr(unsafe.Pointer(lpBitmapName)))
	return ret
}

func ChangeDisplaySettingsA(lpDevMode *DEVMODEA, dwFlags CDS_TYPE) DISP_CHANGE {
	addr := lazyAddr(&pChangeDisplaySettingsA, libUser32, "ChangeDisplaySettingsA")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpDevMode)), uintptr(dwFlags))
	return DISP_CHANGE(ret)
}

var ChangeDisplaySettings = ChangeDisplaySettingsW

func ChangeDisplaySettingsW(lpDevMode *DEVMODEW, dwFlags CDS_TYPE) DISP_CHANGE {
	addr := lazyAddr(&pChangeDisplaySettingsW, libUser32, "ChangeDisplaySettingsW")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpDevMode)), uintptr(dwFlags))
	return DISP_CHANGE(ret)
}

func ChangeDisplaySettingsExA(lpszDeviceName PSTR, lpDevMode *DEVMODEA, hwnd HWND, dwflags CDS_TYPE, lParam unsafe.Pointer) DISP_CHANGE {
	addr := lazyAddr(&pChangeDisplaySettingsExA, libUser32, "ChangeDisplaySettingsExA")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpszDeviceName)), uintptr(unsafe.Pointer(lpDevMode)), hwnd, uintptr(dwflags), uintptr(lParam))
	return DISP_CHANGE(ret)
}

var ChangeDisplaySettingsEx = ChangeDisplaySettingsExW

func ChangeDisplaySettingsExW(lpszDeviceName PWSTR, lpDevMode *DEVMODEW, hwnd HWND, dwflags CDS_TYPE, lParam unsafe.Pointer) DISP_CHANGE {
	addr := lazyAddr(&pChangeDisplaySettingsExW, libUser32, "ChangeDisplaySettingsExW")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpszDeviceName)), uintptr(unsafe.Pointer(lpDevMode)), hwnd, uintptr(dwflags), uintptr(lParam))
	return DISP_CHANGE(ret)
}

func EnumDisplaySettingsA(lpszDeviceName PSTR, iModeNum ENUM_DISPLAY_SETTINGS_MODE, lpDevMode *DEVMODEA) BOOL {
	addr := lazyAddr(&pEnumDisplaySettingsA, libUser32, "EnumDisplaySettingsA")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpszDeviceName)), uintptr(iModeNum), uintptr(unsafe.Pointer(lpDevMode)))
	return BOOL(ret)
}

var EnumDisplaySettings = EnumDisplaySettingsW

func EnumDisplaySettingsW(lpszDeviceName PWSTR, iModeNum ENUM_DISPLAY_SETTINGS_MODE, lpDevMode *DEVMODEW) BOOL {
	addr := lazyAddr(&pEnumDisplaySettingsW, libUser32, "EnumDisplaySettingsW")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpszDeviceName)), uintptr(iModeNum), uintptr(unsafe.Pointer(lpDevMode)))
	return BOOL(ret)
}

func EnumDisplaySettingsExA(lpszDeviceName PSTR, iModeNum ENUM_DISPLAY_SETTINGS_MODE, lpDevMode *DEVMODEA, dwFlags uint32) BOOL {
	addr := lazyAddr(&pEnumDisplaySettingsExA, libUser32, "EnumDisplaySettingsExA")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpszDeviceName)), uintptr(iModeNum), uintptr(unsafe.Pointer(lpDevMode)), uintptr(dwFlags))
	return BOOL(ret)
}

var EnumDisplaySettingsEx = EnumDisplaySettingsExW

func EnumDisplaySettingsExW(lpszDeviceName PWSTR, iModeNum ENUM_DISPLAY_SETTINGS_MODE, lpDevMode *DEVMODEW, dwFlags uint32) BOOL {
	addr := lazyAddr(&pEnumDisplaySettingsExW, libUser32, "EnumDisplaySettingsExW")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpszDeviceName)), uintptr(iModeNum), uintptr(unsafe.Pointer(lpDevMode)), uintptr(dwFlags))
	return BOOL(ret)
}

func EnumDisplayDevicesA(lpDevice PSTR, iDevNum uint32, lpDisplayDevice *DISPLAY_DEVICEA, dwFlags uint32) BOOL {
	addr := lazyAddr(&pEnumDisplayDevicesA, libUser32, "EnumDisplayDevicesA")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpDevice)), uintptr(iDevNum), uintptr(unsafe.Pointer(lpDisplayDevice)), uintptr(dwFlags))
	return BOOL(ret)
}

var EnumDisplayDevices = EnumDisplayDevicesW

func EnumDisplayDevicesW(lpDevice PWSTR, iDevNum uint32, lpDisplayDevice *DISPLAY_DEVICEW, dwFlags uint32) BOOL {
	addr := lazyAddr(&pEnumDisplayDevicesW, libUser32, "EnumDisplayDevicesW")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpDevice)), uintptr(iDevNum), uintptr(unsafe.Pointer(lpDisplayDevice)), uintptr(dwFlags))
	return BOOL(ret)
}

func MonitorFromPoint(pt POINT, dwFlags MONITOR_FROM_FLAGS) HMONITOR {
	addr := lazyAddr(&pMonitorFromPoint, libUser32, "MonitorFromPoint")
	ret, _, _ := syscall.SyscallN(addr, *(*uintptr)(unsafe.Pointer(&pt)), uintptr(dwFlags))
	return ret
}

func MonitorFromRect(lprc *RECT, dwFlags MONITOR_FROM_FLAGS) HMONITOR {
	addr := lazyAddr(&pMonitorFromRect, libUser32, "MonitorFromRect")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lprc)), uintptr(dwFlags))
	return ret
}

func MonitorFromWindow(hwnd HWND, dwFlags MONITOR_FROM_FLAGS) HMONITOR {
	addr := lazyAddr(&pMonitorFromWindow, libUser32, "MonitorFromWindow")
	ret, _, _ := syscall.SyscallN(addr, hwnd, uintptr(dwFlags))
	return ret
}

func GetMonitorInfoA(hMonitor HMONITOR, lpmi *MONITORINFO) BOOL {
	addr := lazyAddr(&pGetMonitorInfoA, libUser32, "GetMonitorInfoA")
	ret, _, _ := syscall.SyscallN(addr, hMonitor, uintptr(unsafe.Pointer(lpmi)))
	return BOOL(ret)
}

var GetMonitorInfo = GetMonitorInfoW

func GetMonitorInfoW(hMonitor HMONITOR, lpmi *MONITORINFO) BOOL {
	addr := lazyAddr(&pGetMonitorInfoW, libUser32, "GetMonitorInfoW")
	ret, _, _ := syscall.SyscallN(addr, hMonitor, uintptr(unsafe.Pointer(lpmi)))
	return BOOL(ret)
}

func EnumDisplayMonitors(hdc HDC, lprcClip *RECT, lpfnEnum MONITORENUMPROC, dwData LPARAM) BOOL {
	addr := lazyAddr(&pEnumDisplayMonitors, libUser32, "EnumDisplayMonitors")
	ret, _, _ := syscall.SyscallN(addr, hdc, uintptr(unsafe.Pointer(lprcClip)), lpfnEnum, dwData)
	return BOOL(ret)
}
