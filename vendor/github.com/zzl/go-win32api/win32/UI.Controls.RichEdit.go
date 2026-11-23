package win32

import (
	"syscall"
	"unsafe"
)

const (
	CchTextLimitDefault               uint32  = 0x7fff
	MSFTEDIT_CLASS                    string  = "RICHEDIT50W"
	CERICHEDIT_CLASSA                 string  = "RichEditCEA"
	CERICHEDIT_CLASSW                 string  = "RichEditCEW"
	RICHEDIT_CLASSA                   string  = "RichEdit20A"
	RICHEDIT_CLASS10A                 string  = "RICHEDIT"
	RICHEDIT_CLASSW                   string  = "RichEdit20W"
	RICHEDIT_CLASS                    string  = "RichEdit20W"
	EM_CANPASTE                       uint32  = 0x432
	EM_DISPLAYBAND                    uint32  = 0x433
	EM_EXGETSEL                       uint32  = 0x434
	EM_EXLIMITTEXT                    uint32  = 0x435
	EM_EXLINEFROMCHAR                 uint32  = 0x436
	EM_EXSETSEL                       uint32  = 0x437
	EM_FINDTEXT                       uint32  = 0x438
	EM_FORMATRANGE                    uint32  = 0x439
	EM_GETCHARFORMAT                  uint32  = 0x43a
	EM_GETEVENTMASK                   uint32  = 0x43b
	EM_GETOLEINTERFACE                uint32  = 0x43c
	EM_GETPARAFORMAT                  uint32  = 0x43d
	EM_GETSELTEXT                     uint32  = 0x43e
	EM_HIDESELECTION                  uint32  = 0x43f
	EM_PASTESPECIAL                   uint32  = 0x440
	EM_REQUESTRESIZE                  uint32  = 0x441
	EM_SELECTIONTYPE                  uint32  = 0x442
	EM_SETBKGNDCOLOR                  uint32  = 0x443
	EM_SETCHARFORMAT                  uint32  = 0x444
	EM_SETEVENTMASK                   uint32  = 0x445
	EM_SETOLECALLBACK                 uint32  = 0x446
	EM_SETPARAFORMAT                  uint32  = 0x447
	EM_SETTARGETDEVICE                uint32  = 0x448
	EM_STREAMIN                       uint32  = 0x449
	EM_STREAMOUT                      uint32  = 0x44a
	EM_GETTEXTRANGE                   uint32  = 0x44b
	EM_FINDWORDBREAK                  uint32  = 0x44c
	EM_SETOPTIONS                     uint32  = 0x44d
	EM_GETOPTIONS                     uint32  = 0x44e
	EM_FINDTEXTEX                     uint32  = 0x44f
	EM_GETWORDBREAKPROCEX             uint32  = 0x450
	EM_SETWORDBREAKPROCEX             uint32  = 0x451
	EM_SETUNDOLIMIT                   uint32  = 0x452
	EM_REDO                           uint32  = 0x454
	EM_CANREDO                        uint32  = 0x455
	EM_GETUNDONAME                    uint32  = 0x456
	EM_GETREDONAME                    uint32  = 0x457
	EM_STOPGROUPTYPING                uint32  = 0x458
	EM_SETTEXTMODE                    uint32  = 0x459
	EM_GETTEXTMODE                    uint32  = 0x45a
	EM_AUTOURLDETECT                  uint32  = 0x45b
	AURL_ENABLEURL                    uint32  = 0x1
	AURL_ENABLEEMAILADDR              uint32  = 0x2
	AURL_ENABLETELNO                  uint32  = 0x4
	AURL_ENABLEEAURLS                 uint32  = 0x8
	AURL_ENABLEDRIVELETTERS           uint32  = 0x10
	AURL_DISABLEMIXEDLGC              uint32  = 0x20
	EM_GETAUTOURLDETECT               uint32  = 0x45c
	EM_SETPALETTE                     uint32  = 0x45d
	EM_GETTEXTEX                      uint32  = 0x45e
	EM_GETTEXTLENGTHEX                uint32  = 0x45f
	EM_SHOWSCROLLBAR                  uint32  = 0x460
	EM_SETTEXTEX                      uint32  = 0x461
	EM_SETPUNCTUATION                 uint32  = 0x464
	EM_GETPUNCTUATION                 uint32  = 0x465
	EM_SETWORDWRAPMODE                uint32  = 0x466
	EM_GETWORDWRAPMODE                uint32  = 0x467
	EM_SETIMECOLOR                    uint32  = 0x468
	EM_GETIMECOLOR                    uint32  = 0x469
	EM_SETIMEOPTIONS                  uint32  = 0x46a
	EM_GETIMEOPTIONS                  uint32  = 0x46b
	EM_CONVPOSITION                   uint32  = 0x46c
	EM_SETLANGOPTIONS                 uint32  = 0x478
	EM_GETLANGOPTIONS                 uint32  = 0x479
	EM_GETIMECOMPMODE                 uint32  = 0x47a
	EM_FINDTEXTW                      uint32  = 0x47b
	EM_FINDTEXTEXW                    uint32  = 0x47c
	EM_RECONVERSION                   uint32  = 0x47d
	EM_SETIMEMODEBIAS                 uint32  = 0x47e
	EM_GETIMEMODEBIAS                 uint32  = 0x47f
	EM_SETBIDIOPTIONS                 uint32  = 0x4c8
	EM_GETBIDIOPTIONS                 uint32  = 0x4c9
	EM_SETTYPOGRAPHYOPTIONS           uint32  = 0x4ca
	EM_GETTYPOGRAPHYOPTIONS           uint32  = 0x4cb
	EM_SETEDITSTYLE                   uint32  = 0x4cc
	EM_GETEDITSTYLE                   uint32  = 0x4cd
	SES_EMULATESYSEDIT                uint32  = 0x1
	SES_BEEPONMAXTEXT                 uint32  = 0x2
	SES_EXTENDBACKCOLOR               uint32  = 0x4
	SES_MAPCPS                        uint32  = 0x8
	SES_HYPERLINKTOOLTIPS             uint32  = 0x8
	SES_EMULATE10                     uint32  = 0x10
	SES_DEFAULTLATINLIGA              uint32  = 0x10
	SES_USECRLF                       uint32  = 0x20
	SES_NOFOCUSLINKNOTIFY             uint32  = 0x20
	SES_USEAIMM                       uint32  = 0x40
	SES_NOIME                         uint32  = 0x80
	SES_ALLOWBEEPS                    uint32  = 0x100
	SES_UPPERCASE                     uint32  = 0x200
	SES_LOWERCASE                     uint32  = 0x400
	SES_NOINPUTSEQUENCECHK            uint32  = 0x800
	SES_BIDI                          uint32  = 0x1000
	SES_SCROLLONKILLFOCUS             uint32  = 0x2000
	SES_XLTCRCRLFTOCR                 uint32  = 0x4000
	SES_DRAFTMODE                     uint32  = 0x8000
	SES_USECTF                        uint32  = 0x10000
	SES_HIDEGRIDLINES                 uint32  = 0x20000
	SES_USEATFONT                     uint32  = 0x40000
	SES_CUSTOMLOOK                    uint32  = 0x80000
	SES_LBSCROLLNOTIFY                uint32  = 0x100000
	SES_CTFALLOWEMBED                 uint32  = 0x200000
	SES_CTFALLOWSMARTTAG              uint32  = 0x400000
	SES_CTFALLOWPROOFING              uint32  = 0x800000
	SES_LOGICALCARET                  uint32  = 0x1000000
	SES_WORDDRAGDROP                  uint32  = 0x2000000
	SES_SMARTDRAGDROP                 uint32  = 0x4000000
	SES_MULTISELECT                   uint32  = 0x8000000
	SES_CTFNOLOCK                     uint32  = 0x10000000
	SES_NOEALINEHEIGHTADJUST          uint32  = 0x20000000
	SES_MAX                           uint32  = 0x20000000
	IMF_AUTOKEYBOARD                  uint32  = 0x1
	IMF_AUTOFONT                      uint32  = 0x2
	IMF_IMECANCELCOMPLETE             uint32  = 0x4
	IMF_IMEALWAYSSENDNOTIFY           uint32  = 0x8
	IMF_AUTOFONTSIZEADJUST            uint32  = 0x10
	IMF_UIFONTS                       uint32  = 0x20
	IMF_NOIMPLICITLANG                uint32  = 0x40
	IMF_DUALFONT                      uint32  = 0x80
	IMF_NOKBDLIDFIXUP                 uint32  = 0x200
	IMF_NORTFFONTSUBSTITUTE           uint32  = 0x400
	IMF_SPELLCHECKING                 uint32  = 0x800
	IMF_TKBPREDICTION                 uint32  = 0x1000
	IMF_IMEUIINTEGRATION              uint32  = 0x2000
	ICM_NOTOPEN                       uint32  = 0x0
	ICM_LEVEL3                        uint32  = 0x1
	ICM_LEVEL2                        uint32  = 0x2
	ICM_LEVEL2_5                      uint32  = 0x3
	ICM_LEVEL2_SUI                    uint32  = 0x4
	ICM_CTF                           uint32  = 0x5
	TO_ADVANCEDTYPOGRAPHY             uint32  = 0x1
	TO_SIMPLELINEBREAK                uint32  = 0x2
	TO_DISABLECUSTOMTEXTOUT           uint32  = 0x4
	TO_ADVANCEDLAYOUT                 uint32  = 0x8
	EM_OUTLINE                        uint32  = 0x4dc
	EM_GETSCROLLPOS                   uint32  = 0x4dd
	EM_SETSCROLLPOS                   uint32  = 0x4de
	EM_SETFONTSIZE                    uint32  = 0x4df
	EM_GETZOOM                        uint32  = 0x4e0
	EM_SETZOOM                        uint32  = 0x4e1
	EM_GETVIEWKIND                    uint32  = 0x4e2
	EM_SETVIEWKIND                    uint32  = 0x4e3
	EM_GETPAGE                        uint32  = 0x4e4
	EM_SETPAGE                        uint32  = 0x4e5
	EM_GETHYPHENATEINFO               uint32  = 0x4e6
	EM_SETHYPHENATEINFO               uint32  = 0x4e7
	EM_GETPAGEROTATE                  uint32  = 0x4eb
	EM_SETPAGEROTATE                  uint32  = 0x4ec
	EM_GETCTFMODEBIAS                 uint32  = 0x4ed
	EM_SETCTFMODEBIAS                 uint32  = 0x4ee
	EM_GETCTFOPENSTATUS               uint32  = 0x4f0
	EM_SETCTFOPENSTATUS               uint32  = 0x4f1
	EM_GETIMECOMPTEXT                 uint32  = 0x4f2
	EM_ISIME                          uint32  = 0x4f3
	EM_GETIMEPROPERTY                 uint32  = 0x4f4
	EM_GETQUERYRTFOBJ                 uint32  = 0x50d
	EM_SETQUERYRTFOBJ                 uint32  = 0x50e
	EPR_0                             uint32  = 0x0
	EPR_270                           uint32  = 0x1
	EPR_180                           uint32  = 0x2
	EPR_90                            uint32  = 0x3
	EPR_SE                            uint32  = 0x5
	CTFMODEBIAS_DEFAULT               uint32  = 0x0
	CTFMODEBIAS_FILENAME              uint32  = 0x1
	CTFMODEBIAS_NAME                  uint32  = 0x2
	CTFMODEBIAS_READING               uint32  = 0x3
	CTFMODEBIAS_DATETIME              uint32  = 0x4
	CTFMODEBIAS_CONVERSATION          uint32  = 0x5
	CTFMODEBIAS_NUMERIC               uint32  = 0x6
	CTFMODEBIAS_HIRAGANA              uint32  = 0x7
	CTFMODEBIAS_KATAKANA              uint32  = 0x8
	CTFMODEBIAS_HANGUL                uint32  = 0x9
	CTFMODEBIAS_HALFWIDTHKATAKANA     uint32  = 0xa
	CTFMODEBIAS_FULLWIDTHALPHANUMERIC uint32  = 0xb
	CTFMODEBIAS_HALFWIDTHALPHANUMERIC uint32  = 0xc
	IMF_SMODE_PLAURALCLAUSE           uint32  = 0x1
	IMF_SMODE_NONE                    uint32  = 0x2
	EMO_EXIT                          uint32  = 0x0
	EMO_ENTER                         uint32  = 0x1
	EMO_PROMOTE                       uint32  = 0x2
	EMO_EXPAND                        uint32  = 0x3
	EMO_MOVESELECTION                 uint32  = 0x4
	EMO_GETVIEWMODE                   uint32  = 0x5
	EMO_EXPANDSELECTION               uint32  = 0x0
	EMO_EXPANDDOCUMENT                uint32  = 0x1
	VM_NORMAL                         uint32  = 0x4
	VM_OUTLINE                        uint32  = 0x2
	VM_PAGE                           uint32  = 0x9
	EM_INSERTTABLE                    uint32  = 0x4e8
	EM_GETAUTOCORRECTPROC             uint32  = 0x4e9
	EM_SETAUTOCORRECTPROC             uint32  = 0x4ea
	EM_CALLAUTOCORRECTPROC            uint32  = 0x4ff
	ATP_NOCHANGE                      uint32  = 0x0
	ATP_CHANGE                        uint32  = 0x1
	ATP_NODELIMITER                   uint32  = 0x2
	ATP_REPLACEALLTEXT                uint32  = 0x4
	EM_GETTABLEPARMS                  uint32  = 0x509
	EM_SETEDITSTYLEEX                 uint32  = 0x513
	EM_GETEDITSTYLEEX                 uint32  = 0x514
	SES_EX_NOTABLE                    uint32  = 0x4
	SES_EX_NOMATH                     uint32  = 0x40
	SES_EX_HANDLEFRIENDLYURL          uint32  = 0x100
	SES_EX_NOTHEMING                  uint32  = 0x80000
	SES_EX_NOACETATESELECTION         uint32  = 0x100000
	SES_EX_USESINGLELINE              uint32  = 0x200000
	SES_EX_MULTITOUCH                 uint32  = 0x8000000
	SES_EX_HIDETEMPFORMAT             uint32  = 0x10000000
	SES_EX_USEMOUSEWPARAM             uint32  = 0x20000000
	EM_GETSTORYTYPE                   uint32  = 0x522
	EM_SETSTORYTYPE                   uint32  = 0x523
	EM_GETELLIPSISMODE                uint32  = 0x531
	EM_SETELLIPSISMODE                uint32  = 0x532
	ELLIPSIS_MASK                     uint32  = 0x3
	ELLIPSIS_NONE                     uint32  = 0x0
	ELLIPSIS_END                      uint32  = 0x1
	ELLIPSIS_WORD                     uint32  = 0x3
	EM_SETTABLEPARMS                  uint32  = 0x533
	EM_GETTOUCHOPTIONS                uint32  = 0x536
	EM_SETTOUCHOPTIONS                uint32  = 0x537
	EM_INSERTIMAGE                    uint32  = 0x53a
	EM_SETUIANAME                     uint32  = 0x540
	EM_GETELLIPSISSTATE               uint32  = 0x542
	RTO_SHOWHANDLES                   uint32  = 0x1
	RTO_DISABLEHANDLES                uint32  = 0x2
	RTO_READINGMODE                   uint32  = 0x3
	EN_MSGFILTER                      uint32  = 0x700
	EN_REQUESTRESIZE                  uint32  = 0x701
	EN_SELCHANGE                      uint32  = 0x702
	EN_DROPFILES                      uint32  = 0x703
	EN_PROTECTED                      uint32  = 0x704
	EN_CORRECTTEXT                    uint32  = 0x705
	EN_STOPNOUNDO                     uint32  = 0x706
	EN_IMECHANGE                      uint32  = 0x707
	EN_SAVECLIPBOARD                  uint32  = 0x708
	EN_OLEOPFAILED                    uint32  = 0x709
	EN_OBJECTPOSITIONS                uint32  = 0x70a
	EN_LINK                           uint32  = 0x70b
	EN_DRAGDROPDONE                   uint32  = 0x70c
	EN_PARAGRAPHEXPANDED              uint32  = 0x70d
	EN_PAGECHANGE                     uint32  = 0x70e
	EN_LOWFIRTF                       uint32  = 0x70f
	EN_ALIGNLTR                       uint32  = 0x710
	EN_ALIGNRTL                       uint32  = 0x711
	EN_CLIPFORMAT                     uint32  = 0x712
	EN_STARTCOMPOSITION               uint32  = 0x713
	EN_ENDCOMPOSITION                 uint32  = 0x714
	ENM_NONE                          uint32  = 0x0
	ENM_CHANGE                        uint32  = 0x1
	ENM_UPDATE                        uint32  = 0x2
	ENM_SCROLL                        uint32  = 0x4
	ENM_SCROLLEVENTS                  uint32  = 0x8
	ENM_DRAGDROPDONE                  uint32  = 0x10
	ENM_PARAGRAPHEXPANDED             uint32  = 0x20
	ENM_PAGECHANGE                    uint32  = 0x40
	ENM_CLIPFORMAT                    uint32  = 0x80
	ENM_KEYEVENTS                     uint32  = 0x10000
	ENM_MOUSEEVENTS                   uint32  = 0x20000
	ENM_REQUESTRESIZE                 uint32  = 0x40000
	ENM_SELCHANGE                     uint32  = 0x80000
	ENM_DROPFILES                     uint32  = 0x100000
	ENM_PROTECTED                     uint32  = 0x200000
	ENM_CORRECTTEXT                   uint32  = 0x400000
	ENM_IMECHANGE                     uint32  = 0x800000
	ENM_LANGCHANGE                    uint32  = 0x1000000
	ENM_OBJECTPOSITIONS               uint32  = 0x2000000
	ENM_LINK                          uint32  = 0x4000000
	ENM_LOWFIRTF                      uint32  = 0x8000000
	ENM_STARTCOMPOSITION              uint32  = 0x10000000
	ENM_ENDCOMPOSITION                uint32  = 0x20000000
	ENM_GROUPTYPINGCHANGE             uint32  = 0x40000000
	ENM_HIDELINKTOOLTIP               uint32  = 0x80000000
	ES_SAVESEL                        uint32  = 0x8000
	ES_SUNKEN                         uint32  = 0x4000
	ES_DISABLENOSCROLL                uint32  = 0x2000
	ES_SELECTIONBAR                   uint32  = 0x1000000
	ES_NOOLEDRAGDROP                  uint32  = 0x8
	ES_EX_NOCALLOLEINIT               uint32  = 0x0
	ES_VERTICAL                       uint32  = 0x400000
	ES_NOIME                          uint32  = 0x80000
	ES_SELFIME                        uint32  = 0x40000
	ECO_AUTOWORDSELECTION             uint32  = 0x1
	ECO_AUTOVSCROLL                   uint32  = 0x40
	ECO_AUTOHSCROLL                   uint32  = 0x80
	ECO_NOHIDESEL                     uint32  = 0x100
	ECO_READONLY                      uint32  = 0x800
	ECO_WANTRETURN                    uint32  = 0x1000
	ECO_SAVESEL                       uint32  = 0x8000
	ECO_SELECTIONBAR                  uint32  = 0x1000000
	ECO_VERTICAL                      uint32  = 0x400000
	ECOOP_SET                         uint32  = 0x1
	ECOOP_OR                          uint32  = 0x2
	ECOOP_AND                         uint32  = 0x3
	ECOOP_XOR                         uint32  = 0x4
	WB_MOVEWORDPREV                   uint32  = 0x4
	WB_MOVEWORDNEXT                   uint32  = 0x5
	WB_PREVBREAK                      uint32  = 0x6
	WB_NEXTBREAK                      uint32  = 0x7
	PC_FOLLOWING                      uint32  = 0x1
	PC_LEADING                        uint32  = 0x2
	PC_OVERFLOW                       uint32  = 0x3
	PC_DELIMITER                      uint32  = 0x4
	WBF_WORDWRAP                      uint32  = 0x10
	WBF_WORDBREAK                     uint32  = 0x20
	WBF_OVERFLOW                      uint32  = 0x40
	WBF_LEVEL1                        uint32  = 0x80
	WBF_LEVEL2                        uint32  = 0x100
	WBF_CUSTOM                        uint32  = 0x200
	IMF_FORCENONE                     uint32  = 0x1
	IMF_FORCEENABLE                   uint32  = 0x2
	IMF_FORCEDISABLE                  uint32  = 0x4
	IMF_CLOSESTATUSWINDOW             uint32  = 0x8
	IMF_VERTICAL                      uint32  = 0x20
	IMF_FORCEACTIVE                   uint32  = 0x40
	IMF_FORCEINACTIVE                 uint32  = 0x80
	IMF_FORCEREMEMBER                 uint32  = 0x100
	IMF_MULTIPLEEDIT                  uint32  = 0x400
	YHeightCharPtsMost                uint32  = 0x666
	SCF_SELECTION                     uint32  = 0x1
	SCF_WORD                          uint32  = 0x2
	SCF_DEFAULT                       uint32  = 0x0
	SCF_ALL                           uint32  = 0x4
	SCF_USEUIRULES                    uint32  = 0x8
	SCF_ASSOCIATEFONT                 uint32  = 0x10
	SCF_NOKBUPDATE                    uint32  = 0x20
	SCF_ASSOCIATEFONT2                uint32  = 0x40
	SCF_SMARTFONT                     uint32  = 0x80
	SCF_CHARREPFROMLCID               uint32  = 0x100
	SPF_DONTSETDEFAULT                uint32  = 0x2
	SPF_SETDEFAULT                    uint32  = 0x4
	SF_TEXT                           uint32  = 0x1
	SF_RTF                            uint32  = 0x2
	SF_RTFNOOBJS                      uint32  = 0x3
	SF_TEXTIZED                       uint32  = 0x4
	SF_UNICODE                        uint32  = 0x10
	SF_USECODEPAGE                    uint32  = 0x20
	SF_NCRFORNONASCII                 uint32  = 0x40
	SFF_WRITEXTRAPAR                  uint32  = 0x80
	SFF_SELECTION                     uint32  = 0x8000
	SFF_PLAINRTF                      uint32  = 0x4000
	SFF_PERSISTVIEWSCALE              uint32  = 0x2000
	SFF_KEEPDOCINFO                   uint32  = 0x1000
	SFF_PWD                           uint32  = 0x800
	SF_RTFVAL                         uint32  = 0x700
	MAX_TAB_STOPS                     uint32  = 0x20
	LDefaultTab                       uint32  = 0x2d0
	MAX_TABLE_CELLS                   uint32  = 0x3f
	PFM_SPACEBEFORE                   uint32  = 0x40
	PFM_SPACEAFTER                    uint32  = 0x80
	PFM_LINESPACING                   uint32  = 0x100
	PFM_STYLE                         uint32  = 0x400
	PFM_BORDER                        uint32  = 0x800
	PFM_SHADING                       uint32  = 0x1000
	PFM_NUMBERINGSTYLE                uint32  = 0x2000
	PFM_NUMBERINGTAB                  uint32  = 0x4000
	PFM_NUMBERINGSTART                uint32  = 0x8000
	PFM_KEEP                          uint32  = 0x20000
	PFM_KEEPNEXT                      uint32  = 0x40000
	PFM_PAGEBREAKBEFORE               uint32  = 0x80000
	PFM_NOLINENUMBER                  uint32  = 0x100000
	PFM_NOWIDOWCONTROL                uint32  = 0x200000
	PFM_DONOTHYPHEN                   uint32  = 0x400000
	PFM_SIDEBYSIDE                    uint32  = 0x800000
	PFM_COLLAPSED                     uint32  = 0x1000000
	PFM_OUTLINELEVEL                  uint32  = 0x2000000
	PFM_BOX                           uint32  = 0x4000000
	PFM_RESERVED2                     uint32  = 0x8000000
	PFM_TABLEROWDELIMITER             uint32  = 0x10000000
	PFM_TEXTWRAPPINGBREAK             uint32  = 0x20000000
	PFM_TABLE                         uint32  = 0x40000000
	PFN_BULLET                        uint32  = 0x1
	PFN_ARABIC                        uint32  = 0x2
	PFN_LCLETTER                      uint32  = 0x3
	PFN_UCLETTER                      uint32  = 0x4
	PFN_LCROMAN                       uint32  = 0x5
	PFN_UCROMAN                       uint32  = 0x6
	PFA_JUSTIFY                       uint32  = 0x4
	PFA_FULL_INTERWORD                uint32  = 0x4
	GCMF_GRIPPER                      uint32  = 0x1
	GCMF_SPELLING                     uint32  = 0x2
	GCMF_TOUCHMENU                    uint32  = 0x4000
	GCMF_MOUSEMENU                    uint32  = 0x2000
	OLEOP_DOVERB                      uint32  = 0x1
	CF_RTF                            string  = "Rich Text Format"
	CF_RTFNOOBJS                      string  = "Rich Text Format Without Objects"
	CF_RETEXTOBJ                      string  = "RichEdit Text and Objects"
	ST_DEFAULT                        uint32  = 0x0
	ST_KEEPUNDO                       uint32  = 0x1
	ST_SELECTION                      uint32  = 0x2
	ST_NEWCHARS                       uint32  = 0x4
	ST_UNICODE                        uint32  = 0x8
	BOM_DEFPARADIR                    uint32  = 0x1
	BOM_PLAINTEXT                     uint32  = 0x2
	BOM_NEUTRALOVERRIDE               uint32  = 0x4
	BOM_CONTEXTREADING                uint32  = 0x8
	BOM_CONTEXTALIGNMENT              uint32  = 0x10
	BOM_LEGACYBIDICLASS               uint32  = 0x40
	BOM_UNICODEBIDI                   uint32  = 0x80
	BOE_RTLDIR                        uint32  = 0x1
	BOE_PLAINTEXT                     uint32  = 0x2
	BOE_NEUTRALOVERRIDE               uint32  = 0x4
	BOE_CONTEXTREADING                uint32  = 0x8
	BOE_CONTEXTALIGNMENT              uint32  = 0x10
	BOE_FORCERECALC                   uint32  = 0x20
	BOE_LEGACYBIDICLASS               uint32  = 0x40
	BOE_UNICODEBIDI                   uint32  = 0x80
	FR_MATCHDIAC                      uint32  = 0x20000000
	FR_MATCHKASHIDA                   uint32  = 0x40000000
	FR_MATCHALEFHAMZA                 uint32  = 0x80000000
	RICHEDIT60_CLASS                  string  = "RICHEDIT60W"
	PFA_FULL_NEWSPAPER                uint32  = 0x5
	PFA_FULL_INTERLETTER              uint32  = 0x6
	PFA_FULL_SCALED                   uint32  = 0x7
	PFA_FULL_GLYPHS                   uint32  = 0x8
	AURL_ENABLEEA                     uint32  = 0x1
	GCM_TOUCHMENU                     uint32  = 0x4000
	GCM_MOUSEMENU                     uint32  = 0x2000
	S_MSG_KEY_IGNORED                 HRESULT = 262657
	TXTBIT_RICHTEXT                   uint32  = 0x1
	TXTBIT_MULTILINE                  uint32  = 0x2
	TXTBIT_READONLY                   uint32  = 0x4
	TXTBIT_SHOWACCELERATOR            uint32  = 0x8
	TXTBIT_USEPASSWORD                uint32  = 0x10
	TXTBIT_HIDESELECTION              uint32  = 0x20
	TXTBIT_SAVESELECTION              uint32  = 0x40
	TXTBIT_AUTOWORDSEL                uint32  = 0x80
	TXTBIT_VERTICAL                   uint32  = 0x100
	TXTBIT_SELBARCHANGE               uint32  = 0x200
	TXTBIT_WORDWRAP                   uint32  = 0x400
	TXTBIT_ALLOWBEEP                  uint32  = 0x800
	TXTBIT_DISABLEDRAG                uint32  = 0x1000
	TXTBIT_VIEWINSETCHANGE            uint32  = 0x2000
	TXTBIT_BACKSTYLECHANGE            uint32  = 0x4000
	TXTBIT_MAXLENGTHCHANGE            uint32  = 0x8000
	TXTBIT_SCROLLBARCHANGE            uint32  = 0x10000
	TXTBIT_CHARFORMATCHANGE           uint32  = 0x20000
	TXTBIT_PARAFORMATCHANGE           uint32  = 0x40000
	TXTBIT_EXTENTCHANGE               uint32  = 0x80000
	TXTBIT_CLIENTRECTCHANGE           uint32  = 0x100000
	TXTBIT_USECURRENTBKG              uint32  = 0x200000
	TXTBIT_NOTHREADREFCOUNT           uint32  = 0x400000
	TXTBIT_SHOWPASSWORD               uint32  = 0x800000
	TXTBIT_D2DDWRITE                  uint32  = 0x1000000
	TXTBIT_D2DSIMPLETYPOGRAPHY        uint32  = 0x2000000
	TXTBIT_D2DPIXELSNAPPED            uint32  = 0x4000000
	TXTBIT_D2DSUBPIXELLINES           uint32  = 0x8000000
	TXTBIT_FLASHLASTPASSWORDCHAR      uint32  = 0x10000000
	TXTBIT_ADVANCEDINPUT              uint32  = 0x20000000
	TXES_ISDIALOG                     uint32  = 0x1
	REO_NULL                          int32   = 0
	REO_READWRITEMASK                 int32   = 2047
	RECO_PASTE                        int32   = 0
	RECO_DROP                         int32   = 1
	RECO_COPY                         int32   = 2
	RECO_CUT                          int32   = 3
	RECO_DRAG                         int32   = 4
)

// enums

// enum
// flags
type CFM_MASK uint32

const (
	CFM_SUBSCRIPT     CFM_MASK = 196608
	CFM_SUPERSCRIPT   CFM_MASK = 196608
	CFM_EFFECTS       CFM_MASK = 1073741887
	CFM_ALL           CFM_MASK = 4160749631
	CFM_BOLD          CFM_MASK = 1
	CFM_CHARSET       CFM_MASK = 134217728
	CFM_COLOR         CFM_MASK = 1073741824
	CFM_FACE          CFM_MASK = 536870912
	CFM_ITALIC        CFM_MASK = 2
	CFM_OFFSET        CFM_MASK = 268435456
	CFM_PROTECTED     CFM_MASK = 16
	CFM_SIZE          CFM_MASK = 2147483648
	CFM_STRIKEOUT     CFM_MASK = 8
	CFM_UNDERLINE     CFM_MASK = 4
	CFM_LINK          CFM_MASK = 32
	CFM_SMALLCAPS     CFM_MASK = 64
	CFM_ALLCAPS       CFM_MASK = 128
	CFM_HIDDEN        CFM_MASK = 256
	CFM_OUTLINE       CFM_MASK = 512
	CFM_SHADOW        CFM_MASK = 1024
	CFM_EMBOSS        CFM_MASK = 2048
	CFM_IMPRINT       CFM_MASK = 4096
	CFM_DISABLED      CFM_MASK = 8192
	CFM_REVISED       CFM_MASK = 16384
	CFM_REVAUTHOR     CFM_MASK = 32768
	CFM_ANIMATION     CFM_MASK = 262144
	CFM_STYLE         CFM_MASK = 524288
	CFM_KERNING       CFM_MASK = 1048576
	CFM_SPACING       CFM_MASK = 2097152
	CFM_WEIGHT        CFM_MASK = 4194304
	CFM_UNDERLINETYPE CFM_MASK = 8388608
	CFM_COOKIE        CFM_MASK = 16777216
	CFM_LCID          CFM_MASK = 33554432
	CFM_BACKCOLOR     CFM_MASK = 67108864
	CFM_EFFECTS2      CFM_MASK = 1141080063
	CFM_ALL2          CFM_MASK = 4294967295
	CFM_FONTBOUND     CFM_MASK = 1048576
	CFM_LINKPROTECTED CFM_MASK = 8388608
	CFM_EXTENDED      CFM_MASK = 33554432
	CFM_MATHNOBUILDUP CFM_MASK = 134217728
	CFM_MATH          CFM_MASK = 268435456
	CFM_MATHORDINARY  CFM_MASK = 536870912
	CFM_ALLEFFECTS    CFM_MASK = 2115207167
)

// enum
// flags
type CFE_EFFECTS uint32

const (
	CFE_ALLCAPS       CFE_EFFECTS = 128
	CFE_AUTOBACKCOLOR CFE_EFFECTS = 67108864
	CFE_DISABLED      CFE_EFFECTS = 8192
	CFE_EMBOSS        CFE_EFFECTS = 2048
	CFE_HIDDEN        CFE_EFFECTS = 256
	CFE_IMPRINT       CFE_EFFECTS = 4096
	CFE_OUTLINE       CFE_EFFECTS = 512
	CFE_REVISED       CFE_EFFECTS = 16384
	CFE_SHADOW        CFE_EFFECTS = 1024
	CFE_SMALLCAPS     CFE_EFFECTS = 64
	CFE_AUTOCOLOR     CFE_EFFECTS = 1073741824
	CFE_BOLD          CFE_EFFECTS = 1
	CFE_ITALIC        CFE_EFFECTS = 2
	CFE_STRIKEOUT     CFE_EFFECTS = 8
	CFE_UNDERLINE_    CFE_EFFECTS = 4
	CFE_PROTECTED     CFE_EFFECTS = 16
	CFE_LINK          CFE_EFFECTS = 32
	CFE_SUBSCRIPT     CFE_EFFECTS = 65536
	CFE_SUPERSCRIPT   CFE_EFFECTS = 131072
	CFE_FONTBOUND     CFE_EFFECTS = 1048576
	CFE_LINKPROTECTED CFE_EFFECTS = 8388608
	CFE_EXTENDED      CFE_EFFECTS = 33554432
	CFE_MATHNOBUILDUP CFE_EFFECTS = 134217728
	CFE_MATH          CFE_EFFECTS = 268435456
	CFE_MATHORDINARY  CFE_EFFECTS = 536870912
)

// enum
// flags
type PARAFORMAT_MASK uint32

const (
	PFM_ALIGNMENT    PARAFORMAT_MASK = 8
	PFM_NUMBERING    PARAFORMAT_MASK = 32
	PFM_OFFSET       PARAFORMAT_MASK = 4
	PFM_OFFSETINDENT PARAFORMAT_MASK = 2147483648
	PFM_RIGHTINDENT  PARAFORMAT_MASK = 2
	PFM_RTLPARA      PARAFORMAT_MASK = 65536
	PFM_STARTINDENT  PARAFORMAT_MASK = 1
	PFM_TABSTOPS     PARAFORMAT_MASK = 16
)

// enum
// flags
type RICH_EDIT_GET_CONTEXT_MENU_SEL_TYPE uint16

const (
	SEL_EMPTY          RICH_EDIT_GET_CONTEXT_MENU_SEL_TYPE = 0
	SEL_TEXT           RICH_EDIT_GET_CONTEXT_MENU_SEL_TYPE = 1
	SEL_OBJECT         RICH_EDIT_GET_CONTEXT_MENU_SEL_TYPE = 2
	SEL_MULTICHAR      RICH_EDIT_GET_CONTEXT_MENU_SEL_TYPE = 4
	SEL_MULTIOBJECT    RICH_EDIT_GET_CONTEXT_MENU_SEL_TYPE = 8
	GCM_RIGHTMOUSEDROP RICH_EDIT_GET_CONTEXT_MENU_SEL_TYPE = 32768
)

// enum
// flags
type RICH_EDIT_GET_OBJECT_FLAGS uint32

const (
	REO_GETOBJ_POLEOBJ        RICH_EDIT_GET_OBJECT_FLAGS = 1
	REO_GETOBJ_PSTG           RICH_EDIT_GET_OBJECT_FLAGS = 2
	REO_GETOBJ_POLESITE       RICH_EDIT_GET_OBJECT_FLAGS = 4
	REO_GETOBJ_NO_INTERFACES  RICH_EDIT_GET_OBJECT_FLAGS = 0
	REO_GETOBJ_ALL_INTERFACES RICH_EDIT_GET_OBJECT_FLAGS = 7
)

// enum
// flags
type PARAFORMAT_BORDERS uint16

const (
	PARAFORMAT_BORDERS_LEFT      PARAFORMAT_BORDERS = 1
	PARAFORMAT_BORDERS_RIGHT     PARAFORMAT_BORDERS = 2
	PARAFORMAT_BORDERS_TOP       PARAFORMAT_BORDERS = 4
	PARAFORMAT_BORDERS_BOTTOM    PARAFORMAT_BORDERS = 8
	PARAFORMAT_BORDERS_INSIDE    PARAFORMAT_BORDERS = 16
	PARAFORMAT_BORDERS_OUTSIDE   PARAFORMAT_BORDERS = 32
	PARAFORMAT_BORDERS_AUTOCOLOR PARAFORMAT_BORDERS = 64
)

// enum
type PARAFORMAT_SHADING_STYLE uint16

const (
	PARAFORMAT_SHADING_STYLE_NONE            PARAFORMAT_SHADING_STYLE = 0
	PARAFORMAT_SHADING_STYLE_DARK_HORIZ      PARAFORMAT_SHADING_STYLE = 1
	PARAFORMAT_SHADING_STYLE_DARK_VERT       PARAFORMAT_SHADING_STYLE = 2
	PARAFORMAT_SHADING_STYLE_DARK_DOWN_DIAG  PARAFORMAT_SHADING_STYLE = 3
	PARAFORMAT_SHADING_STYLE_DARK_UP_DIAG    PARAFORMAT_SHADING_STYLE = 4
	PARAFORMAT_SHADING_STYLE_DARK_GRID       PARAFORMAT_SHADING_STYLE = 5
	PARAFORMAT_SHADING_STYLE_DARK_TRELLIS    PARAFORMAT_SHADING_STYLE = 6
	PARAFORMAT_SHADING_STYLE_LIGHT_HORZ      PARAFORMAT_SHADING_STYLE = 7
	PARAFORMAT_SHADING_STYLE_LIGHT_VERT      PARAFORMAT_SHADING_STYLE = 8
	PARAFORMAT_SHADING_STYLE_LIGHT_DOWN_DIAG PARAFORMAT_SHADING_STYLE = 9
	PARAFORMAT_SHADING_STYLE_LIGHT_UP_DIAG   PARAFORMAT_SHADING_STYLE = 10
	PARAFORMAT_SHADING_STYLE_LIGHT_GRID      PARAFORMAT_SHADING_STYLE = 11
	PARAFORMAT_SHADING_STYLE_LIGHT_TRELLIS   PARAFORMAT_SHADING_STYLE = 12
)

// enum
type GETTEXTEX_FLAGS uint32

const (
	GT_DEFAULT      GETTEXTEX_FLAGS = 0
	GT_NOHIDDENTEXT GETTEXTEX_FLAGS = 8
	GT_RAWTEXT      GETTEXTEX_FLAGS = 4
	GT_SELECTION    GETTEXTEX_FLAGS = 2
	GT_USECRLF      GETTEXTEX_FLAGS = 1
)

// enum
type ENDCOMPOSITIONNOTIFY_CODE uint32

const (
	ECN_ENDCOMPOSITION ENDCOMPOSITIONNOTIFY_CODE = 1
	ECN_NEWTEXT        ENDCOMPOSITIONNOTIFY_CODE = 2
)

// enum
type IMECOMPTEXT_FLAGS uint32

const (
	ICT_RESULTREADSTR IMECOMPTEXT_FLAGS = 1
)

// enum
// flags
type GETTEXTLENGTHEX_FLAGS uint32

const (
	GTL_DEFAULT  GETTEXTLENGTHEX_FLAGS = 0
	GTL_USECRLF  GETTEXTLENGTHEX_FLAGS = 1
	GTL_PRECISE  GETTEXTLENGTHEX_FLAGS = 2
	GTL_CLOSE    GETTEXTLENGTHEX_FLAGS = 4
	GTL_NUMCHARS GETTEXTLENGTHEX_FLAGS = 8
	GTL_NUMBYTES GETTEXTLENGTHEX_FLAGS = 16
)

// enum
// flags
type REOBJECT_FLAGS uint32

const (
	REO_ALIGNTORIGHT    REOBJECT_FLAGS = 256
	REO_BELOWBASELINE   REOBJECT_FLAGS = 2
	REO_BLANK           REOBJECT_FLAGS = 16
	REO_CANROTATE       REOBJECT_FLAGS = 128
	REO_DONTNEEDPALETTE REOBJECT_FLAGS = 32
	REO_DYNAMICSIZE     REOBJECT_FLAGS = 8
	REO_GETMETAFILE     REOBJECT_FLAGS = 4194304
	REO_HILITED         REOBJECT_FLAGS = 16777216
	REO_INPLACEACTIVE   REOBJECT_FLAGS = 33554432
	REO_INVERTEDSELECT  REOBJECT_FLAGS = 4
	REO_LINK            REOBJECT_FLAGS = 2147483648
	REO_LINKAVAILABLE   REOBJECT_FLAGS = 8388608
	REO_OPEN            REOBJECT_FLAGS = 67108864
	REO_OWNERDRAWSELECT REOBJECT_FLAGS = 64
	REO_RESIZABLE       REOBJECT_FLAGS = 1
	REO_SELECTED        REOBJECT_FLAGS = 134217728
	REO_STATIC          REOBJECT_FLAGS = 1073741824
	REO_USEASBACKGROUND REOBJECT_FLAGS = 1024
	REO_WRAPTEXTAROUND  REOBJECT_FLAGS = 512
)

// enum
type PARAFORMAT_NUMBERING_STYLE uint16

const (
	PFNS_PAREN     PARAFORMAT_NUMBERING_STYLE = 0
	PFNS_PARENS    PARAFORMAT_NUMBERING_STYLE = 256
	PFNS_PERIOD    PARAFORMAT_NUMBERING_STYLE = 512
	PFNS_PLAIN     PARAFORMAT_NUMBERING_STYLE = 768
	PFNS_NONUMBER  PARAFORMAT_NUMBERING_STYLE = 1024
	PFNS_NEWNUMBER PARAFORMAT_NUMBERING_STYLE = 32768
)

// enum
type PARAFORMAT_ALIGNMENT uint16

const (
	PFA_CENTER PARAFORMAT_ALIGNMENT = 3
	PFA_LEFT   PARAFORMAT_ALIGNMENT = 1
	PFA_RIGHT  PARAFORMAT_ALIGNMENT = 2
)

// enum
type TEXTMODE int32

const (
	TM_PLAINTEXT       TEXTMODE = 1
	TM_RICHTEXT        TEXTMODE = 2
	TM_SINGLELEVELUNDO TEXTMODE = 4
	TM_MULTILEVELUNDO  TEXTMODE = 8
	TM_SINGLECODEPAGE  TEXTMODE = 16
	TM_MULTICODEPAGE   TEXTMODE = 32
)

// enum
type UNDONAMEID int32

const (
	UID_UNKNOWN   UNDONAMEID = 0
	UID_TYPING    UNDONAMEID = 1
	UID_DELETE    UNDONAMEID = 2
	UID_DRAGDROP  UNDONAMEID = 3
	UID_CUT       UNDONAMEID = 4
	UID_PASTE     UNDONAMEID = 5
	UID_AUTOTABLE UNDONAMEID = 6
)

// enum
type KHYPH int32

const (
	KhyphNil          KHYPH = 0
	KhyphNormal       KHYPH = 1
	KhyphAddBefore    KHYPH = 2
	KhyphChangeBefore KHYPH = 3
	KhyphDeleteBefore KHYPH = 4
	KhyphChangeAfter  KHYPH = 5
	KhyphDelAndChange KHYPH = 6
)

// enum
type TXTBACKSTYLE int32

const (
	TXTBACK_TRANSPARENT TXTBACKSTYLE = 0
	TXTBACK_OPAQUE      TXTBACKSTYLE = 1
)

// enum
type TXTHITRESULT int32

const (
	TXTHITRESULT_NOHIT       TXTHITRESULT = 0
	TXTHITRESULT_TRANSPARENT TXTHITRESULT = 1
	TXTHITRESULT_CLOSE       TXTHITRESULT = 2
	TXTHITRESULT_HIT         TXTHITRESULT = 3
)

// enum
type TXTNATURALSIZE int32

const (
	TXTNS_FITTOCONTENT2   TXTNATURALSIZE = 0
	TXTNS_FITTOCONTENT    TXTNATURALSIZE = 1
	TXTNS_ROUNDTOLINE     TXTNATURALSIZE = 2
	TXTNS_FITTOCONTENT3   TXTNATURALSIZE = 3
	TXTNS_FITTOCONTENTWSP TXTNATURALSIZE = 4
	TXTNS_INCLUDELASTLINE TXTNATURALSIZE = 1073741824
	TXTNS_EMU             TXTNATURALSIZE = -2147483648
)

// enum
type TXTVIEW int32

const (
	TXTVIEW_ACTIVE   TXTVIEW = 0
	TXTVIEW_INACTIVE TXTVIEW = -1
)

// enum
type CHANGETYPE int32

const (
	CN_GENERIC     CHANGETYPE = 0
	CN_TEXTCHANGED CHANGETYPE = 1
	CN_NEWUNDO     CHANGETYPE = 2
	CN_NEWREDO     CHANGETYPE = 4
)

// enum
type CARET_FLAGS int32

const (
	CARET_NONE     CARET_FLAGS = 0
	CARET_CUSTOM   CARET_FLAGS = 1
	CARET_RTL      CARET_FLAGS = 2
	CARET_ITALIC   CARET_FLAGS = 32
	CARET_NULL     CARET_FLAGS = 64
	CARET_ROTATE90 CARET_FLAGS = 128
)

// enum
type TomConstants int32

const (
	TomFalse                           TomConstants = 0
	TomTrue                            TomConstants = -1
	TomUndefined                       TomConstants = -9999999
	TomToggle                          TomConstants = -9999998
	TomAutoColor                       TomConstants = -9999997
	TomDefault                         TomConstants = -9999996
	TomSuspend                         TomConstants = -9999995
	TomResume                          TomConstants = -9999994
	TomApplyNow                        TomConstants = 0
	TomApplyLater                      TomConstants = 1
	TomTrackParms                      TomConstants = 2
	TomCacheParms                      TomConstants = 3
	TomApplyTmp                        TomConstants = 4
	TomDisableSmartFont                TomConstants = 8
	TomEnableSmartFont                 TomConstants = 9
	TomUsePoints                       TomConstants = 10
	TomUseTwips                        TomConstants = 11
	TomBackward                        TomConstants = -1073741823
	TomForward                         TomConstants = 1073741823
	TomMove                            TomConstants = 0
	TomExtend                          TomConstants = 1
	TomNoSelection                     TomConstants = 0
	TomSelectionIP                     TomConstants = 1
	TomSelectionNormal                 TomConstants = 2
	TomSelectionFrame                  TomConstants = 3
	TomSelectionColumn                 TomConstants = 4
	TomSelectionRow                    TomConstants = 5
	TomSelectionBlock                  TomConstants = 6
	TomSelectionInlineShape            TomConstants = 7
	TomSelectionShape                  TomConstants = 8
	TomSelStartActive                  TomConstants = 1
	TomSelAtEOL                        TomConstants = 2
	TomSelOvertype                     TomConstants = 4
	TomSelActive                       TomConstants = 8
	TomSelReplace                      TomConstants = 16
	TomEnd                             TomConstants = 0
	TomStart                           TomConstants = 32
	TomCollapseEnd                     TomConstants = 0
	TomCollapseStart                   TomConstants = 1
	TomClientCoord                     TomConstants = 256
	TomAllowOffClient                  TomConstants = 512
	TomTransform                       TomConstants = 1024
	TomObjectArg                       TomConstants = 2048
	TomAtEnd                           TomConstants = 4096
	TomNone                            TomConstants = 0
	TomSingle                          TomConstants = 1
	TomWords                           TomConstants = 2
	TomDouble                          TomConstants = 3
	TomDotted                          TomConstants = 4
	TomDash                            TomConstants = 5
	TomDashDot                         TomConstants = 6
	TomDashDotDot                      TomConstants = 7
	TomWave                            TomConstants = 8
	TomThick                           TomConstants = 9
	TomHair                            TomConstants = 10
	TomDoubleWave                      TomConstants = 11
	TomHeavyWave                       TomConstants = 12
	TomLongDash                        TomConstants = 13
	TomThickDash                       TomConstants = 14
	TomThickDashDot                    TomConstants = 15
	TomThickDashDotDot                 TomConstants = 16
	TomThickDotted                     TomConstants = 17
	TomThickLongDash                   TomConstants = 18
	TomLineSpaceSingle                 TomConstants = 0
	TomLineSpace1pt5                   TomConstants = 1
	TomLineSpaceDouble                 TomConstants = 2
	TomLineSpaceAtLeast                TomConstants = 3
	TomLineSpaceExactly                TomConstants = 4
	TomLineSpaceMultiple               TomConstants = 5
	TomLineSpacePercent                TomConstants = 6
	TomAlignLeft                       TomConstants = 0
	TomAlignCenter                     TomConstants = 1
	TomAlignRight                      TomConstants = 2
	TomAlignJustify                    TomConstants = 3
	TomAlignDecimal                    TomConstants = 3
	TomAlignBar                        TomConstants = 4
	TomDefaultTab                      TomConstants = 5
	TomAlignInterWord                  TomConstants = 3
	TomAlignNewspaper                  TomConstants = 4
	TomAlignInterLetter                TomConstants = 5
	TomAlignScaled                     TomConstants = 6
	TomSpaces                          TomConstants = 0
	TomDots                            TomConstants = 1
	TomDashes                          TomConstants = 2
	TomLines                           TomConstants = 3
	TomThickLines                      TomConstants = 4
	TomEquals                          TomConstants = 5
	TomTabBack                         TomConstants = -3
	TomTabNext                         TomConstants = -2
	TomTabHere                         TomConstants = -1
	TomListNone                        TomConstants = 0
	TomListBullet                      TomConstants = 1
	TomListNumberAsArabic              TomConstants = 2
	TomListNumberAsLCLetter            TomConstants = 3
	TomListNumberAsUCLetter            TomConstants = 4
	TomListNumberAsLCRoman             TomConstants = 5
	TomListNumberAsUCRoman             TomConstants = 6
	TomListNumberAsSequence            TomConstants = 7
	TomListNumberedCircle              TomConstants = 8
	TomListNumberedBlackCircleWingding TomConstants = 9
	TomListNumberedWhiteCircleWingding TomConstants = 10
	TomListNumberedArabicWide          TomConstants = 11
	TomListNumberedChS                 TomConstants = 12
	TomListNumberedChT                 TomConstants = 13
	TomListNumberedJpnChS              TomConstants = 14
	TomListNumberedJpnKor              TomConstants = 15
	TomListNumberedArabic1             TomConstants = 16
	TomListNumberedArabic2             TomConstants = 17
	TomListNumberedHebrew              TomConstants = 18
	TomListNumberedThaiAlpha           TomConstants = 19
	TomListNumberedThaiNum             TomConstants = 20
	TomListNumberedHindiAlpha          TomConstants = 21
	TomListNumberedHindiAlpha1         TomConstants = 22
	TomListNumberedHindiNum            TomConstants = 23
	TomListParentheses                 TomConstants = 65536
	TomListPeriod                      TomConstants = 131072
	TomListPlain                       TomConstants = 196608
	TomListNoNumber                    TomConstants = 262144
	TomListMinus                       TomConstants = 524288
	TomIgnoreNumberStyle               TomConstants = 16777216
	TomParaStyleNormal                 TomConstants = -1
	TomParaStyleHeading1               TomConstants = -2
	TomParaStyleHeading2               TomConstants = -3
	TomParaStyleHeading3               TomConstants = -4
	TomParaStyleHeading4               TomConstants = -5
	TomParaStyleHeading5               TomConstants = -6
	TomParaStyleHeading6               TomConstants = -7
	TomParaStyleHeading7               TomConstants = -8
	TomParaStyleHeading8               TomConstants = -9
	TomParaStyleHeading9               TomConstants = -10
	TomCharacter                       TomConstants = 1
	TomWord                            TomConstants = 2
	TomSentence                        TomConstants = 3
	TomParagraph                       TomConstants = 4
	TomLine                            TomConstants = 5
	TomStory                           TomConstants = 6
	TomScreen                          TomConstants = 7
	TomSection                         TomConstants = 8
	TomTableColumn                     TomConstants = 9
	TomColumn                          TomConstants = 9
	TomRow                             TomConstants = 10
	TomWindow                          TomConstants = 11
	TomCell                            TomConstants = 12
	TomCharFormat                      TomConstants = 13
	TomParaFormat                      TomConstants = 14
	TomTable                           TomConstants = 15
	TomObject                          TomConstants = 16
	TomPage                            TomConstants = 17
	TomHardParagraph                   TomConstants = 18
	TomCluster                         TomConstants = 19
	TomInlineObject                    TomConstants = 20
	TomInlineObjectArg                 TomConstants = 21
	TomLeafLine                        TomConstants = 22
	TomLayoutColumn                    TomConstants = 23
	TomProcessId                       TomConstants = 1073741825
	TomMatchWord                       TomConstants = 2
	TomMatchCase                       TomConstants = 4
	TomMatchPattern                    TomConstants = 8
	TomUnknownStory                    TomConstants = 0
	TomMainTextStory                   TomConstants = 1
	TomFootnotesStory                  TomConstants = 2
	TomEndnotesStory                   TomConstants = 3
	TomCommentsStory                   TomConstants = 4
	TomTextFrameStory                  TomConstants = 5
	TomEvenPagesHeaderStory            TomConstants = 6
	TomPrimaryHeaderStory              TomConstants = 7
	TomEvenPagesFooterStory            TomConstants = 8
	TomPrimaryFooterStory              TomConstants = 9
	TomFirstPageHeaderStory            TomConstants = 10
	TomFirstPageFooterStory            TomConstants = 11
	TomScratchStory                    TomConstants = 127
	TomFindStory                       TomConstants = 128
	TomReplaceStory                    TomConstants = 129
	TomStoryInactive                   TomConstants = 0
	TomStoryActiveDisplay              TomConstants = 1
	TomStoryActiveUI                   TomConstants = 2
	TomStoryActiveDisplayUI            TomConstants = 3
	TomNoAnimation                     TomConstants = 0
	TomLasVegasLights                  TomConstants = 1
	TomBlinkingBackground              TomConstants = 2
	TomSparkleText                     TomConstants = 3
	TomMarchingBlackAnts               TomConstants = 4
	TomMarchingRedAnts                 TomConstants = 5
	TomShimmer                         TomConstants = 6
	TomWipeDown                        TomConstants = 7
	TomWipeRight                       TomConstants = 8
	TomAnimationMax                    TomConstants = 8
	TomLowerCase                       TomConstants = 0
	TomUpperCase                       TomConstants = 1
	TomTitleCase                       TomConstants = 2
	TomSentenceCase                    TomConstants = 4
	TomToggleCase                      TomConstants = 5
	TomReadOnly                        TomConstants = 256
	TomShareDenyRead                   TomConstants = 512
	TomShareDenyWrite                  TomConstants = 1024
	TomPasteFile                       TomConstants = 4096
	TomCreateNew                       TomConstants = 16
	TomCreateAlways                    TomConstants = 32
	TomOpenExisting                    TomConstants = 48
	TomOpenAlways                      TomConstants = 64
	TomTruncateExisting                TomConstants = 80
	TomRTF                             TomConstants = 1
	TomText                            TomConstants = 2
	TomHTML                            TomConstants = 3
	TomWordDocument                    TomConstants = 4
	TomBold                            TomConstants = -2147483647
	TomItalic                          TomConstants = -2147483646
	TomUnderline                       TomConstants = -2147483644
	TomStrikeout                       TomConstants = -2147483640
	TomProtected                       TomConstants = -2147483632
	TomLink                            TomConstants = -2147483616
	TomSmallCaps                       TomConstants = -2147483584
	TomAllCaps                         TomConstants = -2147483520
	TomHidden                          TomConstants = -2147483392
	TomOutline                         TomConstants = -2147483136
	TomShadow                          TomConstants = -2147482624
	TomEmboss                          TomConstants = -2147481600
	TomImprint                         TomConstants = -2147479552
	TomDisabled                        TomConstants = -2147475456
	TomRevised                         TomConstants = -2147467264
	TomSubscriptCF                     TomConstants = -2147418112
	TomSuperscriptCF                   TomConstants = -2147352576
	TomFontBound                       TomConstants = -2146435072
	TomLinkProtected                   TomConstants = -2139095040
	TomInlineObjectStart               TomConstants = -2130706432
	TomExtendedChar                    TomConstants = -2113929216
	TomAutoBackColor                   TomConstants = -2080374784
	TomMathZoneNoBuildUp               TomConstants = -2013265920
	TomMathZone                        TomConstants = -1879048192
	TomMathZoneOrdinary                TomConstants = -1610612736
	TomAutoTextColor                   TomConstants = -1073741824
	TomMathZoneDisplay                 TomConstants = 262144
	TomParaEffectRTL                   TomConstants = 1
	TomParaEffectKeep                  TomConstants = 2
	TomParaEffectKeepNext              TomConstants = 4
	TomParaEffectPageBreakBefore       TomConstants = 8
	TomParaEffectNoLineNumber          TomConstants = 16
	TomParaEffectNoWidowControl        TomConstants = 32
	TomParaEffectDoNotHyphen           TomConstants = 64
	TomParaEffectSideBySide            TomConstants = 128
	TomParaEffectCollapsed             TomConstants = 256
	TomParaEffectOutlineLevel          TomConstants = 512
	TomParaEffectBox                   TomConstants = 1024
	TomParaEffectTableRowDelimiter     TomConstants = 4096
	TomParaEffectTable                 TomConstants = 16384
	TomModWidthPairs                   TomConstants = 1
	TomModWidthSpace                   TomConstants = 2
	TomAutoSpaceAlpha                  TomConstants = 4
	TomAutoSpaceNumeric                TomConstants = 8
	TomAutoSpaceParens                 TomConstants = 16
	TomEmbeddedFont                    TomConstants = 32
	TomDoublestrike                    TomConstants = 64
	TomOverlapping                     TomConstants = 128
	TomNormalCaret                     TomConstants = 0
	TomKoreanBlockCaret                TomConstants = 1
	TomNullCaret                       TomConstants = 2
	TomIncludeInset                    TomConstants = 1
	TomUnicodeBiDi                     TomConstants = 1
	TomMathCFCheck                     TomConstants = 4
	TomUnlink                          TomConstants = 8
	TomUnhide                          TomConstants = 16
	TomCheckTextLimit                  TomConstants = 32
	TomIgnoreCurrentFont               TomConstants = 0
	TomMatchCharRep                    TomConstants = 1
	TomMatchFontSignature              TomConstants = 2
	TomMatchAscii                      TomConstants = 4
	TomGetHeightOnly                   TomConstants = 8
	TomMatchMathFont                   TomConstants = 16
	TomCharset                         TomConstants = -2147483648
	TomCharRepFromLcid                 TomConstants = 1073741824
	TomAnsi                            TomConstants = 0
	TomEastEurope                      TomConstants = 1
	TomCyrillic                        TomConstants = 2
	TomGreek                           TomConstants = 3
	TomTurkish                         TomConstants = 4
	TomHebrew                          TomConstants = 5
	TomArabic                          TomConstants = 6
	TomBaltic                          TomConstants = 7
	TomVietnamese                      TomConstants = 8
	TomDefaultCharRep                  TomConstants = 9
	TomSymbol                          TomConstants = 10
	TomThai                            TomConstants = 11
	TomShiftJIS                        TomConstants = 12
	TomGB2312                          TomConstants = 13
	TomHangul                          TomConstants = 14
	TomBIG5                            TomConstants = 15
	TomPC437                           TomConstants = 16
	TomOEM                             TomConstants = 17
	TomMac                             TomConstants = 18
	TomArmenian                        TomConstants = 19
	TomSyriac                          TomConstants = 20
	TomThaana                          TomConstants = 21
	TomDevanagari                      TomConstants = 22
	TomBengali                         TomConstants = 23
	TomGurmukhi                        TomConstants = 24
	TomGujarati                        TomConstants = 25
	TomOriya                           TomConstants = 26
	TomTamil                           TomConstants = 27
	TomTelugu                          TomConstants = 28
	TomKannada                         TomConstants = 29
	TomMalayalam                       TomConstants = 30
	TomSinhala                         TomConstants = 31
	TomLao                             TomConstants = 32
	TomTibetan                         TomConstants = 33
	TomMyanmar                         TomConstants = 34
	TomGeorgian                        TomConstants = 35
	TomJamo                            TomConstants = 36
	TomEthiopic                        TomConstants = 37
	TomCherokee                        TomConstants = 38
	TomAboriginal                      TomConstants = 39
	TomOgham                           TomConstants = 40
	TomRunic                           TomConstants = 41
	TomKhmer                           TomConstants = 42
	TomMongolian                       TomConstants = 43
	TomBraille                         TomConstants = 44
	TomYi                              TomConstants = 45
	TomLimbu                           TomConstants = 46
	TomTaiLe                           TomConstants = 47
	TomNewTaiLue                       TomConstants = 48
	TomSylotiNagri                     TomConstants = 49
	TomKharoshthi                      TomConstants = 50
	TomKayahli                         TomConstants = 51
	TomUsymbol                         TomConstants = 52
	TomEmoji                           TomConstants = 53
	TomGlagolitic                      TomConstants = 54
	TomLisu                            TomConstants = 55
	TomVai                             TomConstants = 56
	TomNKo                             TomConstants = 57
	TomOsmanya                         TomConstants = 58
	TomPhagsPa                         TomConstants = 59
	TomGothic                          TomConstants = 60
	TomDeseret                         TomConstants = 61
	TomTifinagh                        TomConstants = 62
	TomCharRepMax                      TomConstants = 63
	TomRE10Mode                        TomConstants = 1
	TomUseAtFont                       TomConstants = 2
	TomTextFlowMask                    TomConstants = 12
	TomTextFlowES                      TomConstants = 0
	TomTextFlowSW                      TomConstants = 4
	TomTextFlowWN                      TomConstants = 8
	TomTextFlowNE                      TomConstants = 12
	TomNoIME                           TomConstants = 524288
	TomSelfIME                         TomConstants = 262144
	TomNoUpScroll                      TomConstants = 65536
	TomNoVpScroll                      TomConstants = 262144
	TomNoLink                          TomConstants = 0
	TomClientLink                      TomConstants = 1
	TomFriendlyLinkName                TomConstants = 2
	TomFriendlyLinkAddress             TomConstants = 3
	TomAutoLinkURL                     TomConstants = 4
	TomAutoLinkEmail                   TomConstants = 5
	TomAutoLinkPhone                   TomConstants = 6
	TomAutoLinkPath                    TomConstants = 7
	TomCompressNone                    TomConstants = 0
	TomCompressPunctuation             TomConstants = 1
	TomCompressPunctuationAndKana      TomConstants = 2
	TomCompressMax                     TomConstants = 2
	TomUnderlinePositionAuto           TomConstants = 0
	TomUnderlinePositionBelow          TomConstants = 1
	TomUnderlinePositionAbove          TomConstants = 2
	TomUnderlinePositionMax            TomConstants = 2
	TomFontAlignmentAuto               TomConstants = 0
	TomFontAlignmentTop                TomConstants = 1
	TomFontAlignmentBaseline           TomConstants = 2
	TomFontAlignmentBottom             TomConstants = 3
	TomFontAlignmentCenter             TomConstants = 4
	TomFontAlignmentMax                TomConstants = 4
	TomRubyBelow                       TomConstants = 128
	TomRubyAlignCenter                 TomConstants = 0
	TomRubyAlign010                    TomConstants = 1
	TomRubyAlign121                    TomConstants = 2
	TomRubyAlignLeft                   TomConstants = 3
	TomRubyAlignRight                  TomConstants = 4
	TomLimitsDefault                   TomConstants = 0
	TomLimitsUnderOver                 TomConstants = 1
	TomLimitsSubSup                    TomConstants = 2
	TomUpperLimitAsSuperScript         TomConstants = 3
	TomLimitsOpposite                  TomConstants = 4
	TomShowLLimPlaceHldr               TomConstants = 8
	TomShowULimPlaceHldr               TomConstants = 16
	TomDontGrowWithContent             TomConstants = 64
	TomGrowWithContent                 TomConstants = 128
	TomSubSupAlign                     TomConstants = 1
	TomLimitAlignMask                  TomConstants = 3
	TomLimitAlignCenter                TomConstants = 0
	TomLimitAlignLeft                  TomConstants = 1
	TomLimitAlignRight                 TomConstants = 2
	TomShowDegPlaceHldr                TomConstants = 8
	TomAlignDefault                    TomConstants = 0
	TomAlignMatchAscentDescent         TomConstants = 2
	TomMathVariant                     TomConstants = 32
	TomStyleDefault                    TomConstants = 0
	TomStyleScriptScriptCramped        TomConstants = 1
	TomStyleScriptScript               TomConstants = 2
	TomStyleScriptCramped              TomConstants = 3
	TomStyleScript                     TomConstants = 4
	TomStyleTextCramped                TomConstants = 5
	TomStyleText                       TomConstants = 6
	TomStyleDisplayCramped             TomConstants = 7
	TomStyleDisplay                    TomConstants = 8
	TomMathRelSize                     TomConstants = 64
	TomDecDecSize                      TomConstants = 254
	TomDecSize                         TomConstants = 255
	TomIncSize                         TomConstants = 65
	TomIncIncSize                      TomConstants = 66
	TomGravityUI                       TomConstants = 0
	TomGravityBack                     TomConstants = 1
	TomGravityFore                     TomConstants = 2
	TomGravityIn                       TomConstants = 3
	TomGravityOut                      TomConstants = 4
	TomGravityBackward                 TomConstants = 536870912
	TomGravityForward                  TomConstants = 1073741824
	TomAdjustCRLF                      TomConstants = 1
	TomUseCRLF                         TomConstants = 2
	TomTextize                         TomConstants = 4
	TomAllowFinalEOP                   TomConstants = 8
	TomFoldMathAlpha                   TomConstants = 16
	TomNoHidden                        TomConstants = 32
	TomIncludeNumbering                TomConstants = 64
	TomTranslateTableCell              TomConstants = 128
	TomNoMathZoneBrackets              TomConstants = 256
	TomConvertMathChar                 TomConstants = 512
	TomNoUCGreekItalic                 TomConstants = 1024
	TomAllowMathBold                   TomConstants = 2048
	TomLanguageTag                     TomConstants = 4096
	TomConvertRTF                      TomConstants = 8192
	TomApplyRtfDocProps                TomConstants = 16384
	TomPhantomShow                     TomConstants = 1
	TomPhantomZeroWidth                TomConstants = 2
	TomPhantomZeroAscent               TomConstants = 4
	TomPhantomZeroDescent              TomConstants = 8
	TomPhantomTransparent              TomConstants = 16
	TomPhantomASmash                   TomConstants = 5
	TomPhantomDSmash                   TomConstants = 9
	TomPhantomHSmash                   TomConstants = 3
	TomPhantomSmash                    TomConstants = 13
	TomPhantomHorz                     TomConstants = 12
	TomPhantomVert                     TomConstants = 2
	TomBoxHideTop                      TomConstants = 1
	TomBoxHideBottom                   TomConstants = 2
	TomBoxHideLeft                     TomConstants = 4
	TomBoxHideRight                    TomConstants = 8
	TomBoxStrikeH                      TomConstants = 16
	TomBoxStrikeV                      TomConstants = 32
	TomBoxStrikeTLBR                   TomConstants = 64
	TomBoxStrikeBLTR                   TomConstants = 128
	TomBoxAlignCenter                  TomConstants = 1
	TomSpaceMask                       TomConstants = 28
	TomSpaceDefault                    TomConstants = 0
	TomSpaceUnary                      TomConstants = 4
	TomSpaceBinary                     TomConstants = 8
	TomSpaceRelational                 TomConstants = 12
	TomSpaceSkip                       TomConstants = 16
	TomSpaceOrd                        TomConstants = 20
	TomSpaceDifferential               TomConstants = 24
	TomSizeText                        TomConstants = 32
	TomSizeScript                      TomConstants = 64
	TomSizeScriptScript                TomConstants = 96
	TomNoBreak                         TomConstants = 128
	TomTransparentForPositioning       TomConstants = 256
	TomTransparentForSpacing           TomConstants = 512
	TomStretchCharBelow                TomConstants = 0
	TomStretchCharAbove                TomConstants = 1
	TomStretchBaseBelow                TomConstants = 2
	TomStretchBaseAbove                TomConstants = 3
	TomMatrixAlignMask                 TomConstants = 3
	TomMatrixAlignCenter               TomConstants = 0
	TomMatrixAlignTopRow               TomConstants = 1
	TomMatrixAlignBottomRow            TomConstants = 3
	TomShowMatPlaceHldr                TomConstants = 8
	TomEqArrayLayoutWidth              TomConstants = 1
	TomEqArrayAlignMask                TomConstants = 12
	TomEqArrayAlignCenter              TomConstants = 0
	TomEqArrayAlignTopRow              TomConstants = 4
	TomEqArrayAlignBottomRow           TomConstants = 12
	TomMathManualBreakMask             TomConstants = 127
	TomMathBreakLeft                   TomConstants = 125
	TomMathBreakCenter                 TomConstants = 126
	TomMathBreakRight                  TomConstants = 127
	TomMathEqAlign                     TomConstants = 128
	TomMathArgShadingStart             TomConstants = 593
	TomMathArgShadingEnd               TomConstants = 594
	TomMathObjShadingStart             TomConstants = 595
	TomMathObjShadingEnd               TomConstants = 596
	TomFunctionTypeNone                TomConstants = 0
	TomFunctionTypeTakesArg            TomConstants = 1
	TomFunctionTypeTakesLim            TomConstants = 2
	TomFunctionTypeTakesLim2           TomConstants = 3
	TomFunctionTypeIsLim               TomConstants = 4
	TomMathParaAlignDefault            TomConstants = 0
	TomMathParaAlignCenterGroup        TomConstants = 1
	TomMathParaAlignCenter             TomConstants = 2
	TomMathParaAlignLeft               TomConstants = 3
	TomMathParaAlignRight              TomConstants = 4
	TomMathDispAlignMask               TomConstants = 3
	TomMathDispAlignCenterGroup        TomConstants = 0
	TomMathDispAlignCenter             TomConstants = 1
	TomMathDispAlignLeft               TomConstants = 2
	TomMathDispAlignRight              TomConstants = 3
	TomMathDispIntUnderOver            TomConstants = 4
	TomMathDispFracTeX                 TomConstants = 8
	TomMathDispNaryGrow                TomConstants = 16
	TomMathDocEmptyArgMask             TomConstants = 96
	TomMathDocEmptyArgAuto             TomConstants = 0
	TomMathDocEmptyArgAlways           TomConstants = 32
	TomMathDocEmptyArgNever            TomConstants = 64
	TomMathDocSbSpOpUnchanged          TomConstants = 128
	TomMathDocDiffMask                 TomConstants = 768
	TomMathDocDiffDefault              TomConstants = 0
	TomMathDocDiffUpright              TomConstants = 256
	TomMathDocDiffItalic               TomConstants = 512
	TomMathDocDiffOpenItalic           TomConstants = 768
	TomMathDispNarySubSup              TomConstants = 1024
	TomMathDispDef                     TomConstants = 2048
	TomMathEnableRtl                   TomConstants = 4096
	TomMathBrkBinMask                  TomConstants = 196608
	TomMathBrkBinBefore                TomConstants = 0
	TomMathBrkBinAfter                 TomConstants = 65536
	TomMathBrkBinDup                   TomConstants = 131072
	TomMathBrkBinSubMask               TomConstants = 786432
	TomMathBrkBinSubMM                 TomConstants = 0
	TomMathBrkBinSubPM                 TomConstants = 262144
	TomMathBrkBinSubMP                 TomConstants = 524288
	TomSelRange                        TomConstants = 597
	TomHstring                         TomConstants = 596
	TomFontPropTeXStyle                TomConstants = 828
	TomFontPropAlign                   TomConstants = 829
	TomFontStretch                     TomConstants = 830
	TomFontStyle                       TomConstants = 831
	TomFontStyleUpright                TomConstants = 0
	TomFontStyleOblique                TomConstants = 1
	TomFontStyleItalic                 TomConstants = 2
	TomFontStretchDefault              TomConstants = 0
	TomFontStretchUltraCondensed       TomConstants = 1
	TomFontStretchExtraCondensed       TomConstants = 2
	TomFontStretchCondensed            TomConstants = 3
	TomFontStretchSemiCondensed        TomConstants = 4
	TomFontStretchNormal               TomConstants = 5
	TomFontStretchSemiExpanded         TomConstants = 6
	TomFontStretchExpanded             TomConstants = 7
	TomFontStretchExtraExpanded        TomConstants = 8
	TomFontStretchUltraExpanded        TomConstants = 9
	TomFontWeightDefault               TomConstants = 0
	TomFontWeightThin                  TomConstants = 100
	TomFontWeightExtraLight            TomConstants = 200
	TomFontWeightLight                 TomConstants = 300
	TomFontWeightNormal                TomConstants = 400
	TomFontWeightRegular               TomConstants = 400
	TomFontWeightMedium                TomConstants = 500
	TomFontWeightSemiBold              TomConstants = 600
	TomFontWeightBold                  TomConstants = 700
	TomFontWeightExtraBold             TomConstants = 800
	TomFontWeightBlack                 TomConstants = 900
	TomFontWeightHeavy                 TomConstants = 900
	TomFontWeightExtraBlack            TomConstants = 950
	TomParaPropMathAlign               TomConstants = 1079
	TomDocMathBuild                    TomConstants = 128
	TomMathLMargin                     TomConstants = 129
	TomMathRMargin                     TomConstants = 130
	TomMathWrapIndent                  TomConstants = 131
	TomMathWrapRight                   TomConstants = 132
	TomMathPostSpace                   TomConstants = 134
	TomMathPreSpace                    TomConstants = 133
	TomMathInterSpace                  TomConstants = 135
	TomMathIntraSpace                  TomConstants = 136
	TomCanCopy                         TomConstants = 137
	TomCanRedo                         TomConstants = 138
	TomCanUndo                         TomConstants = 139
	TomUndoLimit                       TomConstants = 140
	TomDocAutoLink                     TomConstants = 141
	TomEllipsisMode                    TomConstants = 142
	TomEllipsisState                   TomConstants = 143
	TomEllipsisNone                    TomConstants = 0
	TomEllipsisEnd                     TomConstants = 1
	TomEllipsisWord                    TomConstants = 3
	TomEllipsisPresent                 TomConstants = 1
	TomVTopCell                        TomConstants = 1
	TomVLowCell                        TomConstants = 2
	TomHStartCell                      TomConstants = 4
	TomHContCell                       TomConstants = 8
	TomRowUpdate                       TomConstants = 1
	TomRowApplyDefault                 TomConstants = 0
	TomCellStructureChangeOnly         TomConstants = 1
	TomRowHeightActual                 TomConstants = 2059
)

// enum
type OBJECTTYPE int32

const (
	TomSimpleText       OBJECTTYPE = 0
	TomRuby             OBJECTTYPE = 1
	TomHorzVert         OBJECTTYPE = 2
	TomWarichu          OBJECTTYPE = 3
	TomEq               OBJECTTYPE = 9
	TomMath             OBJECTTYPE = 10
	TomAccent           OBJECTTYPE = 10
	TomBox              OBJECTTYPE = 11
	TomBoxedFormula     OBJECTTYPE = 12
	TomBrackets         OBJECTTYPE = 13
	TomBracketsWithSeps OBJECTTYPE = 14
	TomEquationArray    OBJECTTYPE = 15
	TomFraction         OBJECTTYPE = 16
	TomFunctionApply    OBJECTTYPE = 17
	TomLeftSubSup       OBJECTTYPE = 18
	TomLowerLimit       OBJECTTYPE = 19
	TomMatrix           OBJECTTYPE = 20
	TomNary             OBJECTTYPE = 21
	TomOpChar           OBJECTTYPE = 22
	TomOverbar          OBJECTTYPE = 23
	TomPhantom          OBJECTTYPE = 24
	TomRadical          OBJECTTYPE = 25
	TomSlashedFraction  OBJECTTYPE = 26
	TomStack            OBJECTTYPE = 27
	TomStretchStack     OBJECTTYPE = 28
	TomSubscript        OBJECTTYPE = 29
	TomSubSup           OBJECTTYPE = 30
	TomSuperscript      OBJECTTYPE = 31
	TomUnderbar         OBJECTTYPE = 32
	TomUpperLimit       OBJECTTYPE = 33
	TomObjectMax        OBJECTTYPE = 33
)

// enum
type MANCODE int32

const (
	MBOLD  MANCODE = 16
	MITAL  MANCODE = 32
	MGREEK MANCODE = 64
	MROMN  MANCODE = 0
	MSCRP  MANCODE = 1
	MFRAK  MANCODE = 2
	MOPEN  MANCODE = 3
	MSANS  MANCODE = 4
	MMONO  MANCODE = 5
	MMATH  MANCODE = 6
	MISOL  MANCODE = 7
	MINIT  MANCODE = 8
	MTAIL  MANCODE = 9
	MSTRCH MANCODE = 10
	MLOOP  MANCODE = 11
	MOPENA MANCODE = 12
)

// structs

type IMECOMPTEXT struct {
	Cb    int32
	Flags IMECOMPTEXT_FLAGS
}

type TABLEROWPARMS struct {
	CbRow        byte
	CbCell       byte
	CCell        byte
	CRow         byte
	DxCellMargin int32
	DxIndent     int32
	DyHeight     int32
	Bitfield_    uint32
	CpStartRow   int32
	BTableLevel  byte
	ICell        byte
}

type TABLECELLPARMS struct {
	DxWidth      int32
	Bitfield_    uint16
	WShading     uint16
	DxBrdrLeft   int16
	DyBrdrTop    int16
	DxBrdrRight  int16
	DyBrdrBottom int16
	CrBrdrLeft   uint32
	CrBrdrTop    uint32
	CrBrdrRight  uint32
	CrBrdrBottom uint32
	CrBackPat    uint32
	CrForePat    uint32
}

type RICHEDIT_IMAGE_PARAMETERS struct {
	XWidth            int32
	YHeight           int32
	Ascent            int32
	Type              TEXT_ALIGN_OPTIONS
	PwszAlternateText PWSTR
	PIStream          *IStream
}

type ENDCOMPOSITIONNOTIFY struct {
	Nmhdr  NMHDR
	DwCode ENDCOMPOSITIONNOTIFY_CODE
}

type CHARFORMATA struct {
	CbSize          uint32
	DwMask          CFM_MASK
	DwEffects       CFE_EFFECTS
	YHeight         int32
	YOffset         int32
	CrTextColor     uint32
	BCharSet        byte
	BPitchAndFamily byte
	SzFaceName      [32]CHAR
}

type CHARFORMAT = CHARFORMATW
type CHARFORMATW struct {
	CbSize          uint32
	DwMask          CFM_MASK
	DwEffects       CFE_EFFECTS
	YHeight         int32
	YOffset         int32
	CrTextColor     uint32
	BCharSet        byte
	BPitchAndFamily byte
	SzFaceName      [32]uint16
}

type CHARFORMAT2W_Anonymous struct {
	Data [1]uint32
}

func (this *CHARFORMAT2W_Anonymous) DwReserved() *uint32 {
	return (*uint32)(unsafe.Pointer(this))
}

func (this *CHARFORMAT2W_Anonymous) DwReservedVal() uint32 {
	return *(*uint32)(unsafe.Pointer(this))
}

func (this *CHARFORMAT2W_Anonymous) DwCookie() *uint32 {
	return (*uint32)(unsafe.Pointer(this))
}

func (this *CHARFORMAT2W_Anonymous) DwCookieVal() uint32 {
	return *(*uint32)(unsafe.Pointer(this))
}

type CHARFORMAT2 = CHARFORMAT2W
type CHARFORMAT2W struct {
	CHARFORMATW
	WWeight     uint16
	SSpacing    int16
	CrBackColor uint32
	Lcid        uint32
	CHARFORMAT2W_Anonymous
	SStyle          int16
	WKerning        uint16
	BUnderlineType  byte
	BAnimation      byte
	BRevAuthor      byte
	BUnderlineColor byte
}

type CHARFORMAT2A_Anonymous struct {
	Data [1]uint32
}

func (this *CHARFORMAT2A_Anonymous) DwReserved() *uint32 {
	return (*uint32)(unsafe.Pointer(this))
}

func (this *CHARFORMAT2A_Anonymous) DwReservedVal() uint32 {
	return *(*uint32)(unsafe.Pointer(this))
}

func (this *CHARFORMAT2A_Anonymous) DwCookie() *uint32 {
	return (*uint32)(unsafe.Pointer(this))
}

func (this *CHARFORMAT2A_Anonymous) DwCookieVal() uint32 {
	return *(*uint32)(unsafe.Pointer(this))
}

type CHARFORMAT2A struct {
	CHARFORMATA
	WWeight     uint16
	SSpacing    int16
	CrBackColor uint32
	Lcid        uint32
	CHARFORMAT2A_Anonymous
	SStyle          int16
	WKerning        uint16
	BUnderlineType  byte
	BAnimation      byte
	BRevAuthor      byte
	BUnderlineColor byte
}

type CHARRANGE struct {
	CpMin int32
	CpMax int32
}

type TEXTRANGEA struct {
	Chrg      CHARRANGE
	LpstrText PSTR
}

type TEXTRANGE = TEXTRANGEW
type TEXTRANGEW struct {
	Chrg      CHARRANGE
	LpstrText PWSTR
}

type EDITSTREAM struct {
	DwCookie    uintptr
	DwError     uint32
	PfnCallback EDITSTREAMCALLBACK
}

type FINDTEXTA struct {
	Chrg      CHARRANGE
	LpstrText PSTR
}

type FINDTEXT = FINDTEXTW
type FINDTEXTW struct {
	Chrg      CHARRANGE
	LpstrText PWSTR
}

type FINDTEXTEXA struct {
	Chrg      CHARRANGE
	LpstrText PSTR
	ChrgText  CHARRANGE
}

type FINDTEXTEX = FINDTEXTEXW
type FINDTEXTEXW struct {
	Chrg      CHARRANGE
	LpstrText PWSTR
	ChrgText  CHARRANGE
}

type FORMATRANGE struct {
	Hdc       HDC
	HdcTarget HDC
	Rc        RECT
	RcPage    RECT
	Chrg      CHARRANGE
}

type PARAFORMAT_Anonymous struct {
	Data [1]uint16
}

func (this *PARAFORMAT_Anonymous) WReserved() *uint16 {
	return (*uint16)(unsafe.Pointer(this))
}

func (this *PARAFORMAT_Anonymous) WReservedVal() uint16 {
	return *(*uint16)(unsafe.Pointer(this))
}

func (this *PARAFORMAT_Anonymous) WEffects() *uint16 {
	return (*uint16)(unsafe.Pointer(this))
}

func (this *PARAFORMAT_Anonymous) WEffectsVal() uint16 {
	return *(*uint16)(unsafe.Pointer(this))
}

type PARAFORMAT struct {
	CbSize     uint32
	DwMask     PARAFORMAT_MASK
	WNumbering uint16
	PARAFORMAT_Anonymous
	DxStartIndent int32
	DxRightIndent int32
	DxOffset      int32
	WAlignment    PARAFORMAT_ALIGNMENT
	CTabCount     int16
	RgxTabs       [32]uint32
}

type PARAFORMAT2 struct {
	PARAFORMAT
	DySpaceBefore    int32
	DySpaceAfter     int32
	DyLineSpacing    int32
	SStyle           int16
	BLineSpacingRule byte
	BOutlineLevel    byte
	WShadingWeight   uint16
	WShadingStyle    PARAFORMAT_SHADING_STYLE
	WNumberingStart  uint16
	WNumberingStyle  PARAFORMAT_NUMBERING_STYLE
	WNumberingTab    uint16
	WBorderSpace     uint16
	WBorderWidth     uint16
	WBorders         PARAFORMAT_BORDERS
}

type MSGFILTER struct {
	Nmhdr  NMHDR
	Msg    uint32
	WParam WPARAM
	LParam LPARAM
}

type REQRESIZE struct {
	Nmhdr NMHDR
	Rc    RECT
}

type SELCHANGE struct {
	Nmhdr  NMHDR
	Chrg   CHARRANGE
	Seltyp RICH_EDIT_GET_CONTEXT_MENU_SEL_TYPE
}

type Grouptypingchange_ struct {
	Nmhdr        NMHDR
	FGroupTyping BOOL
}

type CLIPBOARDFORMAT struct {
	Nmhdr NMHDR
	Cf    uint16
}

type GETCONTEXTMENUEX struct {
	Chrg       CHARRANGE
	DwFlags    uint32
	Pt         POINT
	PvReserved unsafe.Pointer
}

type ENDROPFILES struct {
	Nmhdr      NMHDR
	HDrop      HANDLE
	Cp         int32
	FProtected BOOL
}

type ENPROTECTED struct {
	Nmhdr  NMHDR
	Msg    uint32
	WParam WPARAM
	LParam LPARAM
	Chrg   CHARRANGE
}

type ENSAVECLIPBOARD struct {
	Nmhdr        NMHDR
	CObjectCount int32
	Cch          int32
}

type ENOLEOPFAILED struct {
	Nmhdr NMHDR
	Iob   int32
	LOper int32
	Hr    HRESULT
}

type OBJECTPOSITIONS struct {
	Nmhdr        NMHDR
	CObjectCount int32
	PcpPositions *int32
}

type ENLINK struct {
	Nmhdr  NMHDR
	Msg    uint32
	WParam WPARAM
	LParam LPARAM
	Chrg   CHARRANGE
}

type ENLOWFIRTF struct {
	Nmhdr     NMHDR
	SzControl PSTR
}

type ENCORRECTTEXT struct {
	Nmhdr  NMHDR
	Chrg   CHARRANGE
	Seltyp RICH_EDIT_GET_CONTEXT_MENU_SEL_TYPE
}

type PUNCTUATION struct {
	ISize         uint32
	SzPunctuation PSTR
}

type COMPCOLOR struct {
	CrText       uint32
	CrBackground uint32
	DwEffects    uint32
}

type REPASTESPECIAL struct {
	DwAspect DVASPECT
	DwParam  uintptr
}

type SETTEXTEX struct {
	Flags    uint32
	Codepage uint32
}

type GETTEXTEX struct {
	Cb            uint32
	Flags         GETTEXTEX_FLAGS
	Codepage      uint32
	LpDefaultChar PSTR
	LpUsedDefChar *int32
}

type GETTEXTLENGTHEX struct {
	Flags    GETTEXTLENGTHEX_FLAGS
	Codepage uint32
}

type BIDIOPTIONS struct {
	CbSize   uint32
	WMask    uint16
	WEffects uint16
}

type Hyphresult struct {
	Khyph   KHYPH
	IchHyph int32
	ChHyph  uint16
}

type HYPHENATEINFO struct {
	CbSize          int16
	DxHyphenateZone int16
	PfnHyphenate    uintptr
}

type CHANGENOTIFY struct {
	DwChangeType CHANGETYPE
	PvCookieData unsafe.Pointer
}

type CARET_INFO struct {
	Data [1]uint64
}

func (this *CARET_INFO) Hbitmap() *HBITMAP {
	return (*HBITMAP)(unsafe.Pointer(this))
}

func (this *CARET_INFO) HbitmapVal() HBITMAP {
	return *(*HBITMAP)(unsafe.Pointer(this))
}

func (this *CARET_INFO) CaretFlags() *CARET_FLAGS {
	return (*CARET_FLAGS)(unsafe.Pointer(this))
}

func (this *CARET_INFO) CaretFlagsVal() CARET_FLAGS {
	return *(*CARET_FLAGS)(unsafe.Pointer(this))
}

type REOBJECT struct {
	CbStruct uint32
	Cp       int32
	Clsid    syscall.GUID
	Poleobj  *IOleObject
	Pstg     *IStorage
	Polesite *IOleClientSite
	Sizel    SIZE
	Dvaspect uint32
	DwFlags  REOBJECT_FLAGS
	DwUser   uint32
}

// func types

type AutoCorrectProc = uintptr
type AutoCorrectProc_func = func(langid uint16, pszBefore PWSTR, pszAfter PWSTR, cchAfter int32, pcchReplaced *int32) int32

type EDITWORDBREAKPROCEX = uintptr
type EDITWORDBREAKPROCEX_func = func(pchText PSTR, cchText int32, bCharSet byte, action int32) int32

type EDITSTREAMCALLBACK = uintptr
type EDITSTREAMCALLBACK_func = func(dwCookie uintptr, pbBuff *byte, cb int32, pcb *int32) uint32

type PCreateTextServices = uintptr
type PCreateTextServices_func = func(punkOuter *IUnknown, pITextHost *ITextHost, ppUnk **IUnknown) HRESULT

type PShutdownTextServices = uintptr
type PShutdownTextServices_func = func(pTextServices *IUnknown) HRESULT

// interfaces

// 00000000-0000-0000-0000-000000000000
var IID_ITextServices = syscall.GUID{0x00000000, 0x0000, 0x0000,
	[8]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}}

type ITextServicesInterface interface {
	IUnknownInterface
	TxSendMessage(msg uint32, wparam WPARAM, lparam LPARAM, plresult *LRESULT) HRESULT
	TxDraw(dwDrawAspect DVASPECT, lindex int32, pvAspect unsafe.Pointer, ptd *DVTARGETDEVICE, hdcDraw HDC, hicTargetDev HDC, lprcBounds *RECTL, lprcWBounds *RECTL, lprcUpdate *RECT, pfnContinue uintptr, dwContinue uint32, lViewId int32) HRESULT
	TxGetHScroll(plMin *int32, plMax *int32, plPos *int32, plPage *int32, pfEnabled *BOOL) HRESULT
	TxGetVScroll(plMin *int32, plMax *int32, plPos *int32, plPage *int32, pfEnabled *BOOL) HRESULT
	OnTxSetCursor(dwDrawAspect DVASPECT, lindex int32, pvAspect unsafe.Pointer, ptd *DVTARGETDEVICE, hdcDraw HDC, hicTargetDev HDC, lprcClient *RECT, x int32, y int32) HRESULT
	TxQueryHitPoint(dwDrawAspect DVASPECT, lindex int32, pvAspect unsafe.Pointer, ptd *DVTARGETDEVICE, hdcDraw HDC, hicTargetDev HDC, lprcClient *RECT, x int32, y int32, pHitResult *uint32) HRESULT
	OnTxInPlaceActivate(prcClient *RECT) HRESULT
	OnTxInPlaceDeactivate() HRESULT
	OnTxUIActivate() HRESULT
	OnTxUIDeactivate() HRESULT
	TxGetText(pbstrText *BSTR) HRESULT
	TxSetText(pszText PWSTR) HRESULT
	TxGetCurTargetX(param0 *int32) HRESULT
	TxGetBaseLinePos(param0 *int32) HRESULT
	TxGetNaturalSize(dwAspect uint32, hdcDraw HDC, hicTargetDev HDC, ptd *DVTARGETDEVICE, dwMode uint32, psizelExtent *SIZE, pwidth *int32, pheight *int32) HRESULT
	TxGetDropTarget(ppDropTarget **IDropTarget) HRESULT
	OnTxPropertyBitsChange(dwMask uint32, dwBits uint32) HRESULT
	TxGetCachedSize(pdwWidth *uint32, pdwHeight *uint32) HRESULT
}

type ITextServicesVtbl struct {
	IUnknownVtbl
	TxSendMessage          uintptr
	TxDraw                 uintptr
	TxGetHScroll           uintptr
	TxGetVScroll           uintptr
	OnTxSetCursor          uintptr
	TxQueryHitPoint        uintptr
	OnTxInPlaceActivate    uintptr
	OnTxInPlaceDeactivate  uintptr
	OnTxUIActivate         uintptr
	OnTxUIDeactivate       uintptr
	TxGetText              uintptr
	TxSetText              uintptr
	TxGetCurTargetX        uintptr
	TxGetBaseLinePos       uintptr
	TxGetNaturalSize       uintptr
	TxGetDropTarget        uintptr
	OnTxPropertyBitsChange uintptr
	TxGetCachedSize        uintptr
}

type ITextServices struct {
	IUnknown
}

func (this *ITextServices) Vtbl() *ITextServicesVtbl {
	return (*ITextServicesVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *ITextServices) TxSendMessage(msg uint32, wparam WPARAM, lparam LPARAM, plresult *LRESULT) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().TxSendMessage, uintptr(unsafe.Pointer(this)), uintptr(msg), wparam, lparam, uintptr(unsafe.Pointer(plresult)))
	return HRESULT(ret)
}

func (this *ITextServices) TxDraw(dwDrawAspect DVASPECT, lindex int32, pvAspect unsafe.Pointer, ptd *DVTARGETDEVICE, hdcDraw HDC, hicTargetDev HDC, lprcBounds *RECTL, lprcWBounds *RECTL, lprcUpdate *RECT, pfnContinue uintptr, dwContinue uint32, lViewId int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().TxDraw, uintptr(unsafe.Pointer(this)), uintptr(dwDrawAspect), uintptr(lindex), uintptr(pvAspect), uintptr(unsafe.Pointer(ptd)), hdcDraw, hicTargetDev, uintptr(unsafe.Pointer(lprcBounds)), uintptr(unsafe.Pointer(lprcWBounds)), uintptr(unsafe.Pointer(lprcUpdate)), pfnContinue, uintptr(dwContinue), uintptr(lViewId))
	return HRESULT(ret)
}

func (this *ITextServices) TxGetHScroll(plMin *int32, plMax *int32, plPos *int32, plPage *int32, pfEnabled *BOOL) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().TxGetHScroll, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(plMin)), uintptr(unsafe.Pointer(plMax)), uintptr(unsafe.Pointer(plPos)), uintptr(unsafe.Pointer(plPage)), uintptr(unsafe.Pointer(pfEnabled)))
	return HRESULT(ret)
}

func (this *ITextServices) TxGetVScroll(plMin *int32, plMax *int32, plPos *int32, plPage *int32, pfEnabled *BOOL) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().TxGetVScroll, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(plMin)), uintptr(unsafe.Pointer(plMax)), uintptr(unsafe.Pointer(plPos)), uintptr(unsafe.Pointer(plPage)), uintptr(unsafe.Pointer(pfEnabled)))
	return HRESULT(ret)
}

func (this *ITextServices) OnTxSetCursor(dwDrawAspect DVASPECT, lindex int32, pvAspect unsafe.Pointer, ptd *DVTARGETDEVICE, hdcDraw HDC, hicTargetDev HDC, lprcClient *RECT, x int32, y int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().OnTxSetCursor, uintptr(unsafe.Pointer(this)), uintptr(dwDrawAspect), uintptr(lindex), uintptr(pvAspect), uintptr(unsafe.Pointer(ptd)), hdcDraw, hicTargetDev, uintptr(unsafe.Pointer(lprcClient)), uintptr(x), uintptr(y))
	return HRESULT(ret)
}

func (this *ITextServices) TxQueryHitPoint(dwDrawAspect DVASPECT, lindex int32, pvAspect unsafe.Pointer, ptd *DVTARGETDEVICE, hdcDraw HDC, hicTargetDev HDC, lprcClient *RECT, x int32, y int32, pHitResult *uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().TxQueryHitPoint, uintptr(unsafe.Pointer(this)), uintptr(dwDrawAspect), uintptr(lindex), uintptr(pvAspect), uintptr(unsafe.Pointer(ptd)), hdcDraw, hicTargetDev, uintptr(unsafe.Pointer(lprcClient)), uintptr(x), uintptr(y), uintptr(unsafe.Pointer(pHitResult)))
	return HRESULT(ret)
}

func (this *ITextServices) OnTxInPlaceActivate(prcClient *RECT) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().OnTxInPlaceActivate, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(prcClient)))
	return HRESULT(ret)
}

func (this *ITextServices) OnTxInPlaceDeactivate() HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().OnTxInPlaceDeactivate, uintptr(unsafe.Pointer(this)))
	return HRESULT(ret)
}

func (this *ITextServices) OnTxUIActivate() HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().OnTxUIActivate, uintptr(unsafe.Pointer(this)))
	return HRESULT(ret)
}

func (this *ITextServices) OnTxUIDeactivate() HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().OnTxUIDeactivate, uintptr(unsafe.Pointer(this)))
	return HRESULT(ret)
}

func (this *ITextServices) TxGetText(pbstrText *BSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().TxGetText, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pbstrText)))
	return HRESULT(ret)
}

func (this *ITextServices) TxSetText(pszText PWSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().TxSetText, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pszText)))
	return HRESULT(ret)
}

func (this *ITextServices) TxGetCurTargetX(param0 *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().TxGetCurTargetX, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(param0)))
	return HRESULT(ret)
}

func (this *ITextServices) TxGetBaseLinePos(param0 *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().TxGetBaseLinePos, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(param0)))
	return HRESULT(ret)
}

func (this *ITextServices) TxGetNaturalSize(dwAspect uint32, hdcDraw HDC, hicTargetDev HDC, ptd *DVTARGETDEVICE, dwMode uint32, psizelExtent *SIZE, pwidth *int32, pheight *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().TxGetNaturalSize, uintptr(unsafe.Pointer(this)), uintptr(dwAspect), hdcDraw, hicTargetDev, uintptr(unsafe.Pointer(ptd)), uintptr(dwMode), uintptr(unsafe.Pointer(psizelExtent)), uintptr(unsafe.Pointer(pwidth)), uintptr(unsafe.Pointer(pheight)))
	return HRESULT(ret)
}

func (this *ITextServices) TxGetDropTarget(ppDropTarget **IDropTarget) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().TxGetDropTarget, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(ppDropTarget)))
	return HRESULT(ret)
}

func (this *ITextServices) OnTxPropertyBitsChange(dwMask uint32, dwBits uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().OnTxPropertyBitsChange, uintptr(unsafe.Pointer(this)), uintptr(dwMask), uintptr(dwBits))
	return HRESULT(ret)
}

func (this *ITextServices) TxGetCachedSize(pdwWidth *uint32, pdwHeight *uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().TxGetCachedSize, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pdwWidth)), uintptr(unsafe.Pointer(pdwHeight)))
	return HRESULT(ret)
}

// 00000000-0000-0000-0000-000000000000
var IID_ITextHost = syscall.GUID{0x00000000, 0x0000, 0x0000,
	[8]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}}

type ITextHostInterface interface {
	IUnknownInterface
	TxGetDC() HDC
	TxReleaseDC(hdc HDC) int32
	TxShowScrollBar(fnBar int32, fShow BOOL) BOOL
	TxEnableScrollBar(fuSBFlags SCROLLBAR_CONSTANTS, fuArrowflags ENABLE_SCROLL_BAR_ARROWS) BOOL
	TxSetScrollRange(fnBar int32, nMinPos int32, nMaxPos int32, fRedraw BOOL) BOOL
	TxSetScrollPos(fnBar int32, nPos int32, fRedraw BOOL) BOOL
	TxInvalidateRect(prc *RECT, fMode BOOL)
	TxViewChange(fUpdate BOOL)
	TxCreateCaret(hbmp HBITMAP, xWidth int32, yHeight int32) BOOL
	TxShowCaret(fShow BOOL) BOOL
	TxSetCaretPos(x int32, y int32) BOOL
	TxSetTimer(idTimer uint32, uTimeout uint32) BOOL
	TxKillTimer(idTimer uint32)
	TxScrollWindowEx(dx int32, dy int32, lprcScroll *RECT, lprcClip *RECT, hrgnUpdate HRGN, lprcUpdate *RECT, fuScroll SHOW_WINDOW_CMD)
	TxSetCapture(fCapture BOOL)
	TxSetFocus()
	TxSetCursor(hcur HCURSOR, fText BOOL)
	TxScreenToClient(lppt *POINT) BOOL
	TxClientToScreen(lppt *POINT) BOOL
	TxActivate(plOldState *int32) HRESULT
	TxDeactivate(lNewState int32) HRESULT
	TxGetClientRect(prc *RECT) HRESULT
	TxGetViewInset(prc *RECT) HRESULT
	TxGetCharFormat(ppCF **CHARFORMATW) HRESULT
	TxGetParaFormat(ppPF **PARAFORMAT) HRESULT
	TxGetSysColor(nIndex int32) uint32
	TxGetBackStyle(pstyle *TXTBACKSTYLE) HRESULT
	TxGetMaxLength(plength *uint32) HRESULT
	TxGetScrollBars(pdwScrollBar *uint32) HRESULT
	TxGetPasswordChar(pch *int8) HRESULT
	TxGetAcceleratorPos(pcp *int32) HRESULT
	TxGetExtent(lpExtent *SIZE) HRESULT
	OnTxCharFormatChange(pCF *CHARFORMATW) HRESULT
	OnTxParaFormatChange(pPF *PARAFORMAT) HRESULT
	TxGetPropertyBits(dwMask uint32, pdwBits *uint32) HRESULT
	TxNotify(iNotify uint32, pv unsafe.Pointer) HRESULT
	TxImmGetContext() HIMC
	TxImmReleaseContext(himc HIMC)
	TxGetSelectionBarWidth(lSelBarWidth *int32) HRESULT
}

type ITextHostVtbl struct {
	IUnknownVtbl
	TxGetDC                uintptr
	TxReleaseDC            uintptr
	TxShowScrollBar        uintptr
	TxEnableScrollBar      uintptr
	TxSetScrollRange       uintptr
	TxSetScrollPos         uintptr
	TxInvalidateRect       uintptr
	TxViewChange           uintptr
	TxCreateCaret          uintptr
	TxShowCaret            uintptr
	TxSetCaretPos          uintptr
	TxSetTimer             uintptr
	TxKillTimer            uintptr
	TxScrollWindowEx       uintptr
	TxSetCapture           uintptr
	TxSetFocus             uintptr
	TxSetCursor            uintptr
	TxScreenToClient       uintptr
	TxClientToScreen       uintptr
	TxActivate             uintptr
	TxDeactivate           uintptr
	TxGetClientRect        uintptr
	TxGetViewInset         uintptr
	TxGetCharFormat        uintptr
	TxGetParaFormat        uintptr
	TxGetSysColor          uintptr
	TxGetBackStyle         uintptr
	TxGetMaxLength         uintptr
	TxGetScrollBars        uintptr
	TxGetPasswordChar      uintptr
	TxGetAcceleratorPos    uintptr
	TxGetExtent            uintptr
	OnTxCharFormatChange   uintptr
	OnTxParaFormatChange   uintptr
	TxGetPropertyBits      uintptr
	TxNotify               uintptr
	TxImmGetContext        uintptr
	TxImmReleaseContext    uintptr
	TxGetSelectionBarWidth uintptr
}

type ITextHost struct {
	IUnknown
}

func (this *ITextHost) Vtbl() *ITextHostVtbl {
	return (*ITextHostVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *ITextHost) TxGetDC() HDC {
	ret, _, _ := syscall.SyscallN(this.Vtbl().TxGetDC, uintptr(unsafe.Pointer(this)))
	return ret
}

func (this *ITextHost) TxReleaseDC(hdc HDC) int32 {
	ret, _, _ := syscall.SyscallN(this.Vtbl().TxReleaseDC, uintptr(unsafe.Pointer(this)), hdc)
	return int32(ret)
}

func (this *ITextHost) TxShowScrollBar(fnBar int32, fShow BOOL) BOOL {
	ret, _, _ := syscall.SyscallN(this.Vtbl().TxShowScrollBar, uintptr(unsafe.Pointer(this)), uintptr(fnBar), uintptr(fShow))
	return BOOL(ret)
}

func (this *ITextHost) TxEnableScrollBar(fuSBFlags SCROLLBAR_CONSTANTS, fuArrowflags ENABLE_SCROLL_BAR_ARROWS) BOOL {
	ret, _, _ := syscall.SyscallN(this.Vtbl().TxEnableScrollBar, uintptr(unsafe.Pointer(this)), uintptr(fuSBFlags), uintptr(fuArrowflags))
	return BOOL(ret)
}

func (this *ITextHost) TxSetScrollRange(fnBar int32, nMinPos int32, nMaxPos int32, fRedraw BOOL) BOOL {
	ret, _, _ := syscall.SyscallN(this.Vtbl().TxSetScrollRange, uintptr(unsafe.Pointer(this)), uintptr(fnBar), uintptr(nMinPos), uintptr(nMaxPos), uintptr(fRedraw))
	return BOOL(ret)
}

func (this *ITextHost) TxSetScrollPos(fnBar int32, nPos int32, fRedraw BOOL) BOOL {
	ret, _, _ := syscall.SyscallN(this.Vtbl().TxSetScrollPos, uintptr(unsafe.Pointer(this)), uintptr(fnBar), uintptr(nPos), uintptr(fRedraw))
	return BOOL(ret)
}

func (this *ITextHost) TxInvalidateRect(prc *RECT, fMode BOOL) {
	_, _, _ = syscall.SyscallN(this.Vtbl().TxInvalidateRect, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(prc)), uintptr(fMode))
}

func (this *ITextHost) TxViewChange(fUpdate BOOL) {
	_, _, _ = syscall.SyscallN(this.Vtbl().TxViewChange, uintptr(unsafe.Pointer(this)), uintptr(fUpdate))
}

func (this *ITextHost) TxCreateCaret(hbmp HBITMAP, xWidth int32, yHeight int32) BOOL {
	ret, _, _ := syscall.SyscallN(this.Vtbl().TxCreateCaret, uintptr(unsafe.Pointer(this)), hbmp, uintptr(xWidth), uintptr(yHeight))
	return BOOL(ret)
}

func (this *ITextHost) TxShowCaret(fShow BOOL) BOOL {
	ret, _, _ := syscall.SyscallN(this.Vtbl().TxShowCaret, uintptr(unsafe.Pointer(this)), uintptr(fShow))
	return BOOL(ret)
}

func (this *ITextHost) TxSetCaretPos(x int32, y int32) BOOL {
	ret, _, _ := syscall.SyscallN(this.Vtbl().TxSetCaretPos, uintptr(unsafe.Pointer(this)), uintptr(x), uintptr(y))
	return BOOL(ret)
}

func (this *ITextHost) TxSetTimer(idTimer uint32, uTimeout uint32) BOOL {
	ret, _, _ := syscall.SyscallN(this.Vtbl().TxSetTimer, uintptr(unsafe.Pointer(this)), uintptr(idTimer), uintptr(uTimeout))
	return BOOL(ret)
}

func (this *ITextHost) TxKillTimer(idTimer uint32) {
	_, _, _ = syscall.SyscallN(this.Vtbl().TxKillTimer, uintptr(unsafe.Pointer(this)), uintptr(idTimer))
}

func (this *ITextHost) TxScrollWindowEx(dx int32, dy int32, lprcScroll *RECT, lprcClip *RECT, hrgnUpdate HRGN, lprcUpdate *RECT, fuScroll SHOW_WINDOW_CMD) {
	_, _, _ = syscall.SyscallN(this.Vtbl().TxScrollWindowEx, uintptr(unsafe.Pointer(this)), uintptr(dx), uintptr(dy), uintptr(unsafe.Pointer(lprcScroll)), uintptr(unsafe.Pointer(lprcClip)), hrgnUpdate, uintptr(unsafe.Pointer(lprcUpdate)), uintptr(fuScroll))
}

func (this *ITextHost) TxSetCapture(fCapture BOOL) {
	_, _, _ = syscall.SyscallN(this.Vtbl().TxSetCapture, uintptr(unsafe.Pointer(this)), uintptr(fCapture))
}

func (this *ITextHost) TxSetFocus() {
	_, _, _ = syscall.SyscallN(this.Vtbl().TxSetFocus, uintptr(unsafe.Pointer(this)))
}

func (this *ITextHost) TxSetCursor(hcur HCURSOR, fText BOOL) {
	_, _, _ = syscall.SyscallN(this.Vtbl().TxSetCursor, uintptr(unsafe.Pointer(this)), hcur, uintptr(fText))
}

func (this *ITextHost) TxScreenToClient(lppt *POINT) BOOL {
	ret, _, _ := syscall.SyscallN(this.Vtbl().TxScreenToClient, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(lppt)))
	return BOOL(ret)
}

func (this *ITextHost) TxClientToScreen(lppt *POINT) BOOL {
	ret, _, _ := syscall.SyscallN(this.Vtbl().TxClientToScreen, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(lppt)))
	return BOOL(ret)
}

func (this *ITextHost) TxActivate(plOldState *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().TxActivate, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(plOldState)))
	return HRESULT(ret)
}

func (this *ITextHost) TxDeactivate(lNewState int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().TxDeactivate, uintptr(unsafe.Pointer(this)), uintptr(lNewState))
	return HRESULT(ret)
}

func (this *ITextHost) TxGetClientRect(prc *RECT) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().TxGetClientRect, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(prc)))
	return HRESULT(ret)
}

func (this *ITextHost) TxGetViewInset(prc *RECT) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().TxGetViewInset, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(prc)))
	return HRESULT(ret)
}

func (this *ITextHost) TxGetCharFormat(ppCF **CHARFORMATW) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().TxGetCharFormat, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(ppCF)))
	return HRESULT(ret)
}

func (this *ITextHost) TxGetParaFormat(ppPF **PARAFORMAT) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().TxGetParaFormat, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(ppPF)))
	return HRESULT(ret)
}

func (this *ITextHost) TxGetSysColor(nIndex int32) uint32 {
	ret, _, _ := syscall.SyscallN(this.Vtbl().TxGetSysColor, uintptr(unsafe.Pointer(this)), uintptr(nIndex))
	return uint32(ret)
}

func (this *ITextHost) TxGetBackStyle(pstyle *TXTBACKSTYLE) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().TxGetBackStyle, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pstyle)))
	return HRESULT(ret)
}

func (this *ITextHost) TxGetMaxLength(plength *uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().TxGetMaxLength, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(plength)))
	return HRESULT(ret)
}

func (this *ITextHost) TxGetScrollBars(pdwScrollBar *uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().TxGetScrollBars, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pdwScrollBar)))
	return HRESULT(ret)
}

func (this *ITextHost) TxGetPasswordChar(pch *int8) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().TxGetPasswordChar, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pch)))
	return HRESULT(ret)
}

func (this *ITextHost) TxGetAcceleratorPos(pcp *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().TxGetAcceleratorPos, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pcp)))
	return HRESULT(ret)
}

func (this *ITextHost) TxGetExtent(lpExtent *SIZE) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().TxGetExtent, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(lpExtent)))
	return HRESULT(ret)
}

func (this *ITextHost) OnTxCharFormatChange(pCF *CHARFORMATW) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().OnTxCharFormatChange, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pCF)))
	return HRESULT(ret)
}

func (this *ITextHost) OnTxParaFormatChange(pPF *PARAFORMAT) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().OnTxParaFormatChange, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pPF)))
	return HRESULT(ret)
}

func (this *ITextHost) TxGetPropertyBits(dwMask uint32, pdwBits *uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().TxGetPropertyBits, uintptr(unsafe.Pointer(this)), uintptr(dwMask), uintptr(unsafe.Pointer(pdwBits)))
	return HRESULT(ret)
}

func (this *ITextHost) TxNotify(iNotify uint32, pv unsafe.Pointer) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().TxNotify, uintptr(unsafe.Pointer(this)), uintptr(iNotify), uintptr(pv))
	return HRESULT(ret)
}

func (this *ITextHost) TxImmGetContext() HIMC {
	ret, _, _ := syscall.SyscallN(this.Vtbl().TxImmGetContext, uintptr(unsafe.Pointer(this)))
	return ret
}

func (this *ITextHost) TxImmReleaseContext(himc HIMC) {
	_, _, _ = syscall.SyscallN(this.Vtbl().TxImmReleaseContext, uintptr(unsafe.Pointer(this)), himc)
}

func (this *ITextHost) TxGetSelectionBarWidth(lSelBarWidth *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().TxGetSelectionBarWidth, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(lSelBarWidth)))
	return HRESULT(ret)
}

// 00000000-0000-0000-0000-000000000000
var IID_IRicheditUiaOverrides = syscall.GUID{0x00000000, 0x0000, 0x0000,
	[8]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}}

type IRicheditUiaOverridesInterface interface {
	IUnknownInterface
	GetPropertyOverrideValue(propertyId int32, pRetValue *VARIANT) HRESULT
}

type IRicheditUiaOverridesVtbl struct {
	IUnknownVtbl
	GetPropertyOverrideValue uintptr
}

type IRicheditUiaOverrides struct {
	IUnknown
}

func (this *IRicheditUiaOverrides) Vtbl() *IRicheditUiaOverridesVtbl {
	return (*IRicheditUiaOverridesVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IRicheditUiaOverrides) GetPropertyOverrideValue(propertyId int32, pRetValue *VARIANT) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetPropertyOverrideValue, uintptr(unsafe.Pointer(this)), uintptr(propertyId), uintptr(unsafe.Pointer(pRetValue)))
	return HRESULT(ret)
}

// 00000000-0000-0000-0000-000000000000
var IID_ITextHost2 = syscall.GUID{0x00000000, 0x0000, 0x0000,
	[8]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}}

type ITextHost2Interface interface {
	ITextHostInterface
	TxIsDoubleClickPending() BOOL
	TxGetWindow(phwnd *HWND) HRESULT
	TxSetForegroundWindow() HRESULT
	TxGetPalette() HPALETTE
	TxGetEastAsianFlags(pFlags *int32) HRESULT
	TxSetCursor2(hcur HCURSOR, bText BOOL) HCURSOR
	TxFreeTextServicesNotification()
	TxGetEditStyle(dwItem uint32, pdwData *uint32) HRESULT
	TxGetWindowStyles(pdwStyle *uint32, pdwExStyle *uint32) HRESULT
	TxShowDropCaret(fShow BOOL, hdc HDC, prc *RECT) HRESULT
	TxDestroyCaret() HRESULT
	TxGetHorzExtent(plHorzExtent *int32) HRESULT
}

type ITextHost2Vtbl struct {
	ITextHostVtbl
	TxIsDoubleClickPending         uintptr
	TxGetWindow                    uintptr
	TxSetForegroundWindow          uintptr
	TxGetPalette                   uintptr
	TxGetEastAsianFlags            uintptr
	TxSetCursor2                   uintptr
	TxFreeTextServicesNotification uintptr
	TxGetEditStyle                 uintptr
	TxGetWindowStyles              uintptr
	TxShowDropCaret                uintptr
	TxDestroyCaret                 uintptr
	TxGetHorzExtent                uintptr
}

type ITextHost2 struct {
	ITextHost
}

func (this *ITextHost2) Vtbl() *ITextHost2Vtbl {
	return (*ITextHost2Vtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *ITextHost2) TxIsDoubleClickPending() BOOL {
	ret, _, _ := syscall.SyscallN(this.Vtbl().TxIsDoubleClickPending, uintptr(unsafe.Pointer(this)))
	return BOOL(ret)
}

func (this *ITextHost2) TxGetWindow(phwnd *HWND) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().TxGetWindow, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(phwnd)))
	return HRESULT(ret)
}

func (this *ITextHost2) TxSetForegroundWindow() HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().TxSetForegroundWindow, uintptr(unsafe.Pointer(this)))
	return HRESULT(ret)
}

func (this *ITextHost2) TxGetPalette() HPALETTE {
	ret, _, _ := syscall.SyscallN(this.Vtbl().TxGetPalette, uintptr(unsafe.Pointer(this)))
	return ret
}

func (this *ITextHost2) TxGetEastAsianFlags(pFlags *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().TxGetEastAsianFlags, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pFlags)))
	return HRESULT(ret)
}

func (this *ITextHost2) TxSetCursor2(hcur HCURSOR, bText BOOL) HCURSOR {
	ret, _, _ := syscall.SyscallN(this.Vtbl().TxSetCursor2, uintptr(unsafe.Pointer(this)), hcur, uintptr(bText))
	return ret
}

func (this *ITextHost2) TxFreeTextServicesNotification() {
	_, _, _ = syscall.SyscallN(this.Vtbl().TxFreeTextServicesNotification, uintptr(unsafe.Pointer(this)))
}

func (this *ITextHost2) TxGetEditStyle(dwItem uint32, pdwData *uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().TxGetEditStyle, uintptr(unsafe.Pointer(this)), uintptr(dwItem), uintptr(unsafe.Pointer(pdwData)))
	return HRESULT(ret)
}

func (this *ITextHost2) TxGetWindowStyles(pdwStyle *uint32, pdwExStyle *uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().TxGetWindowStyles, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pdwStyle)), uintptr(unsafe.Pointer(pdwExStyle)))
	return HRESULT(ret)
}

func (this *ITextHost2) TxShowDropCaret(fShow BOOL, hdc HDC, prc *RECT) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().TxShowDropCaret, uintptr(unsafe.Pointer(this)), uintptr(fShow), hdc, uintptr(unsafe.Pointer(prc)))
	return HRESULT(ret)
}

func (this *ITextHost2) TxDestroyCaret() HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().TxDestroyCaret, uintptr(unsafe.Pointer(this)))
	return HRESULT(ret)
}

func (this *ITextHost2) TxGetHorzExtent(plHorzExtent *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().TxGetHorzExtent, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(plHorzExtent)))
	return HRESULT(ret)
}

// 00000000-0000-0000-0000-000000000000
var IID_ITextServices2 = syscall.GUID{0x00000000, 0x0000, 0x0000,
	[8]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}}

type ITextServices2Interface interface {
	ITextServicesInterface
	TxGetNaturalSize2(dwAspect uint32, hdcDraw HDC, hicTargetDev HDC, ptd *DVTARGETDEVICE, dwMode uint32, psizelExtent *SIZE, pwidth *int32, pheight *int32, pascent *int32) HRESULT
	TxDrawD2D(pRenderTarget unsafe.Pointer, lprcBounds *RECTL, lprcUpdate *RECT, lViewId int32) HRESULT
}

type ITextServices2Vtbl struct {
	ITextServicesVtbl
	TxGetNaturalSize2 uintptr
	TxDrawD2D         uintptr
}

type ITextServices2 struct {
	ITextServices
}

func (this *ITextServices2) Vtbl() *ITextServices2Vtbl {
	return (*ITextServices2Vtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *ITextServices2) TxGetNaturalSize2(dwAspect uint32, hdcDraw HDC, hicTargetDev HDC, ptd *DVTARGETDEVICE, dwMode uint32, psizelExtent *SIZE, pwidth *int32, pheight *int32, pascent *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().TxGetNaturalSize2, uintptr(unsafe.Pointer(this)), uintptr(dwAspect), hdcDraw, hicTargetDev, uintptr(unsafe.Pointer(ptd)), uintptr(dwMode), uintptr(unsafe.Pointer(psizelExtent)), uintptr(unsafe.Pointer(pwidth)), uintptr(unsafe.Pointer(pheight)), uintptr(unsafe.Pointer(pascent)))
	return HRESULT(ret)
}

func (this *ITextServices2) TxDrawD2D(pRenderTarget unsafe.Pointer, lprcBounds *RECTL, lprcUpdate *RECT, lViewId int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().TxDrawD2D, uintptr(unsafe.Pointer(this)), uintptr(pRenderTarget), uintptr(unsafe.Pointer(lprcBounds)), uintptr(unsafe.Pointer(lprcUpdate)), uintptr(lViewId))
	return HRESULT(ret)
}

// 00020D00-0000-0000-C000-000000000046
var IID_IRichEditOle = syscall.GUID{0x00020D00, 0x0000, 0x0000,
	[8]byte{0xC0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x46}}

type IRichEditOleInterface interface {
	IUnknownInterface
	GetClientSite(lplpolesite **IOleClientSite) HRESULT
	GetObjectCount() int32
	GetLinkCount() int32
	GetObject(iob int32, lpreobject *REOBJECT, dwFlags RICH_EDIT_GET_OBJECT_FLAGS) HRESULT
	InsertObject(lpreobject *REOBJECT) HRESULT
	ConvertObject(iob int32, rclsidNew *syscall.GUID, lpstrUserTypeNew PSTR) HRESULT
	ActivateAs(rclsid *syscall.GUID, rclsidAs *syscall.GUID) HRESULT
	SetHostNames(lpstrContainerApp PSTR, lpstrContainerObj PSTR) HRESULT
	SetLinkAvailable(iob int32, fAvailable BOOL) HRESULT
	SetDvaspect(iob int32, dvaspect uint32) HRESULT
	HandsOffStorage(iob int32) HRESULT
	SaveCompleted(iob int32, lpstg *IStorage) HRESULT
	InPlaceDeactivate() HRESULT
	ContextSensitiveHelp(fEnterMode BOOL) HRESULT
	GetClipboardData(lpchrg *CHARRANGE, reco uint32, lplpdataobj **IDataObject) HRESULT
	ImportDataObject(lpdataobj *IDataObject, cf uint16, hMetaPict uintptr) HRESULT
}

type IRichEditOleVtbl struct {
	IUnknownVtbl
	GetClientSite        uintptr
	GetObjectCount       uintptr
	GetLinkCount         uintptr
	GetObject            uintptr
	InsertObject         uintptr
	ConvertObject        uintptr
	ActivateAs           uintptr
	SetHostNames         uintptr
	SetLinkAvailable     uintptr
	SetDvaspect          uintptr
	HandsOffStorage      uintptr
	SaveCompleted        uintptr
	InPlaceDeactivate    uintptr
	ContextSensitiveHelp uintptr
	GetClipboardData     uintptr
	ImportDataObject     uintptr
}

type IRichEditOle struct {
	IUnknown
}

func (this *IRichEditOle) Vtbl() *IRichEditOleVtbl {
	return (*IRichEditOleVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IRichEditOle) GetClientSite(lplpolesite **IOleClientSite) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetClientSite, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(lplpolesite)))
	return HRESULT(ret)
}

func (this *IRichEditOle) GetObjectCount() int32 {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetObjectCount, uintptr(unsafe.Pointer(this)))
	return int32(ret)
}

func (this *IRichEditOle) GetLinkCount() int32 {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetLinkCount, uintptr(unsafe.Pointer(this)))
	return int32(ret)
}

func (this *IRichEditOle) GetObject(iob int32, lpreobject *REOBJECT, dwFlags RICH_EDIT_GET_OBJECT_FLAGS) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetObject, uintptr(unsafe.Pointer(this)), uintptr(iob), uintptr(unsafe.Pointer(lpreobject)), uintptr(dwFlags))
	return HRESULT(ret)
}

func (this *IRichEditOle) InsertObject(lpreobject *REOBJECT) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().InsertObject, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(lpreobject)))
	return HRESULT(ret)
}

func (this *IRichEditOle) ConvertObject(iob int32, rclsidNew *syscall.GUID, lpstrUserTypeNew PSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().ConvertObject, uintptr(unsafe.Pointer(this)), uintptr(iob), uintptr(unsafe.Pointer(rclsidNew)), uintptr(unsafe.Pointer(lpstrUserTypeNew)))
	return HRESULT(ret)
}

func (this *IRichEditOle) ActivateAs(rclsid *syscall.GUID, rclsidAs *syscall.GUID) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().ActivateAs, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(rclsid)), uintptr(unsafe.Pointer(rclsidAs)))
	return HRESULT(ret)
}

func (this *IRichEditOle) SetHostNames(lpstrContainerApp PSTR, lpstrContainerObj PSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetHostNames, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(lpstrContainerApp)), uintptr(unsafe.Pointer(lpstrContainerObj)))
	return HRESULT(ret)
}

func (this *IRichEditOle) SetLinkAvailable(iob int32, fAvailable BOOL) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetLinkAvailable, uintptr(unsafe.Pointer(this)), uintptr(iob), uintptr(fAvailable))
	return HRESULT(ret)
}

func (this *IRichEditOle) SetDvaspect(iob int32, dvaspect uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetDvaspect, uintptr(unsafe.Pointer(this)), uintptr(iob), uintptr(dvaspect))
	return HRESULT(ret)
}

func (this *IRichEditOle) HandsOffStorage(iob int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().HandsOffStorage, uintptr(unsafe.Pointer(this)), uintptr(iob))
	return HRESULT(ret)
}

func (this *IRichEditOle) SaveCompleted(iob int32, lpstg *IStorage) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SaveCompleted, uintptr(unsafe.Pointer(this)), uintptr(iob), uintptr(unsafe.Pointer(lpstg)))
	return HRESULT(ret)
}

func (this *IRichEditOle) InPlaceDeactivate() HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().InPlaceDeactivate, uintptr(unsafe.Pointer(this)))
	return HRESULT(ret)
}

func (this *IRichEditOle) ContextSensitiveHelp(fEnterMode BOOL) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().ContextSensitiveHelp, uintptr(unsafe.Pointer(this)), uintptr(fEnterMode))
	return HRESULT(ret)
}

func (this *IRichEditOle) GetClipboardData(lpchrg *CHARRANGE, reco uint32, lplpdataobj **IDataObject) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetClipboardData, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(lpchrg)), uintptr(reco), uintptr(unsafe.Pointer(lplpdataobj)))
	return HRESULT(ret)
}

func (this *IRichEditOle) ImportDataObject(lpdataobj *IDataObject, cf uint16, hMetaPict uintptr) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().ImportDataObject, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(lpdataobj)), uintptr(cf), hMetaPict)
	return HRESULT(ret)
}

// 00020D03-0000-0000-C000-000000000046
var IID_IRichEditOleCallback = syscall.GUID{0x00020D03, 0x0000, 0x0000,
	[8]byte{0xC0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x46}}

type IRichEditOleCallbackInterface interface {
	IUnknownInterface
	GetNewStorage(lplpstg **IStorage) HRESULT
	GetInPlaceContext(lplpFrame **IOleInPlaceFrame, lplpDoc **IOleInPlaceUIWindow, lpFrameInfo *OIFI) HRESULT
	ShowContainerUI(fShow BOOL) HRESULT
	QueryInsertObject(lpclsid *syscall.GUID, lpstg *IStorage, cp int32) HRESULT
	DeleteObject(lpoleobj *IOleObject) HRESULT
	QueryAcceptData(lpdataobj *IDataObject, lpcfFormat *uint16, reco uint32, fReally BOOL, hMetaPict uintptr) HRESULT
	ContextSensitiveHelp(fEnterMode BOOL) HRESULT
	GetClipboardData(lpchrg *CHARRANGE, reco uint32, lplpdataobj **IDataObject) HRESULT
	GetDragDropEffect(fDrag BOOL, grfKeyState uint32, pdwEffect *uint32) HRESULT
	GetContextMenu(seltype RICH_EDIT_GET_CONTEXT_MENU_SEL_TYPE, lpoleobj *IOleObject, lpchrg *CHARRANGE, lphmenu *HMENU) HRESULT
}

type IRichEditOleCallbackVtbl struct {
	IUnknownVtbl
	GetNewStorage        uintptr
	GetInPlaceContext    uintptr
	ShowContainerUI      uintptr
	QueryInsertObject    uintptr
	DeleteObject         uintptr
	QueryAcceptData      uintptr
	ContextSensitiveHelp uintptr
	GetClipboardData     uintptr
	GetDragDropEffect    uintptr
	GetContextMenu       uintptr
}

type IRichEditOleCallback struct {
	IUnknown
}

func (this *IRichEditOleCallback) Vtbl() *IRichEditOleCallbackVtbl {
	return (*IRichEditOleCallbackVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IRichEditOleCallback) GetNewStorage(lplpstg **IStorage) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetNewStorage, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(lplpstg)))
	return HRESULT(ret)
}

func (this *IRichEditOleCallback) GetInPlaceContext(lplpFrame **IOleInPlaceFrame, lplpDoc **IOleInPlaceUIWindow, lpFrameInfo *OIFI) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetInPlaceContext, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(lplpFrame)), uintptr(unsafe.Pointer(lplpDoc)), uintptr(unsafe.Pointer(lpFrameInfo)))
	return HRESULT(ret)
}

func (this *IRichEditOleCallback) ShowContainerUI(fShow BOOL) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().ShowContainerUI, uintptr(unsafe.Pointer(this)), uintptr(fShow))
	return HRESULT(ret)
}

func (this *IRichEditOleCallback) QueryInsertObject(lpclsid *syscall.GUID, lpstg *IStorage, cp int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().QueryInsertObject, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(lpclsid)), uintptr(unsafe.Pointer(lpstg)), uintptr(cp))
	return HRESULT(ret)
}

func (this *IRichEditOleCallback) DeleteObject(lpoleobj *IOleObject) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().DeleteObject, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(lpoleobj)))
	return HRESULT(ret)
}

func (this *IRichEditOleCallback) QueryAcceptData(lpdataobj *IDataObject, lpcfFormat *uint16, reco uint32, fReally BOOL, hMetaPict uintptr) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().QueryAcceptData, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(lpdataobj)), uintptr(unsafe.Pointer(lpcfFormat)), uintptr(reco), uintptr(fReally), hMetaPict)
	return HRESULT(ret)
}

func (this *IRichEditOleCallback) ContextSensitiveHelp(fEnterMode BOOL) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().ContextSensitiveHelp, uintptr(unsafe.Pointer(this)), uintptr(fEnterMode))
	return HRESULT(ret)
}

func (this *IRichEditOleCallback) GetClipboardData(lpchrg *CHARRANGE, reco uint32, lplpdataobj **IDataObject) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetClipboardData, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(lpchrg)), uintptr(reco), uintptr(unsafe.Pointer(lplpdataobj)))
	return HRESULT(ret)
}

func (this *IRichEditOleCallback) GetDragDropEffect(fDrag BOOL, grfKeyState uint32, pdwEffect *uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetDragDropEffect, uintptr(unsafe.Pointer(this)), uintptr(fDrag), uintptr(grfKeyState), uintptr(unsafe.Pointer(pdwEffect)))
	return HRESULT(ret)
}

func (this *IRichEditOleCallback) GetContextMenu(seltype RICH_EDIT_GET_CONTEXT_MENU_SEL_TYPE, lpoleobj *IOleObject, lpchrg *CHARRANGE, lphmenu *HMENU) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetContextMenu, uintptr(unsafe.Pointer(this)), uintptr(seltype), uintptr(unsafe.Pointer(lpoleobj)), uintptr(unsafe.Pointer(lpchrg)), uintptr(unsafe.Pointer(lphmenu)))
	return HRESULT(ret)
}

// 8CC497C0-A1DF-11CE-8098-00AA0047BE5D
var IID_ITextDocument = syscall.GUID{0x8CC497C0, 0xA1DF, 0x11CE,
	[8]byte{0x80, 0x98, 0x00, 0xAA, 0x00, 0x47, 0xBE, 0x5D}}

type ITextDocumentInterface interface {
	IDispatchInterface
	GetName(pName *BSTR) HRESULT
	GetSelection(ppSel **ITextSelection) HRESULT
	GetStoryCount(pCount *int32) HRESULT
	GetStoryRanges(ppStories **ITextStoryRanges) HRESULT
	GetSaved(pValue *int32) HRESULT
	SetSaved(Value TomConstants) HRESULT
	GetDefaultTabStop(pValue *float32) HRESULT
	SetDefaultTabStop(Value float32) HRESULT
	New() HRESULT
	Open(pVar *VARIANT, Flags int32, CodePage int32) HRESULT
	Save(pVar *VARIANT, Flags int32, CodePage int32) HRESULT
	Freeze(pCount *int32) HRESULT
	Unfreeze(pCount *int32) HRESULT
	BeginEditCollection() HRESULT
	EndEditCollection() HRESULT
	Undo(Count int32, pCount *int32) HRESULT
	Redo(Count int32, pCount *int32) HRESULT
	Range(cpActive int32, cpAnchor int32, ppRange **ITextRange) HRESULT
	RangeFromPoint(x int32, y int32, ppRange **ITextRange) HRESULT
}

type ITextDocumentVtbl struct {
	IDispatchVtbl
	GetName             uintptr
	GetSelection        uintptr
	GetStoryCount       uintptr
	GetStoryRanges      uintptr
	GetSaved            uintptr
	SetSaved            uintptr
	GetDefaultTabStop   uintptr
	SetDefaultTabStop   uintptr
	New                 uintptr
	Open                uintptr
	Save                uintptr
	Freeze              uintptr
	Unfreeze            uintptr
	BeginEditCollection uintptr
	EndEditCollection   uintptr
	Undo                uintptr
	Redo                uintptr
	Range               uintptr
	RangeFromPoint      uintptr
}

type ITextDocument struct {
	IDispatch
}

func (this *ITextDocument) Vtbl() *ITextDocumentVtbl {
	return (*ITextDocumentVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *ITextDocument) GetName(pName *BSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetName, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pName)))
	return HRESULT(ret)
}

func (this *ITextDocument) GetSelection(ppSel **ITextSelection) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetSelection, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(ppSel)))
	return HRESULT(ret)
}

func (this *ITextDocument) GetStoryCount(pCount *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetStoryCount, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pCount)))
	return HRESULT(ret)
}

func (this *ITextDocument) GetStoryRanges(ppStories **ITextStoryRanges) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetStoryRanges, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(ppStories)))
	return HRESULT(ret)
}

func (this *ITextDocument) GetSaved(pValue *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetSaved, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pValue)))
	return HRESULT(ret)
}

func (this *ITextDocument) SetSaved(Value TomConstants) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetSaved, uintptr(unsafe.Pointer(this)), uintptr(Value))
	return HRESULT(ret)
}

func (this *ITextDocument) GetDefaultTabStop(pValue *float32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetDefaultTabStop, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pValue)))
	return HRESULT(ret)
}

func (this *ITextDocument) SetDefaultTabStop(Value float32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetDefaultTabStop, uintptr(unsafe.Pointer(this)), uintptr(Value))
	return HRESULT(ret)
}

func (this *ITextDocument) New() HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().New, uintptr(unsafe.Pointer(this)))
	return HRESULT(ret)
}

func (this *ITextDocument) Open(pVar *VARIANT, Flags int32, CodePage int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Open, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pVar)), uintptr(Flags), uintptr(CodePage))
	return HRESULT(ret)
}

func (this *ITextDocument) Save(pVar *VARIANT, Flags int32, CodePage int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Save, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pVar)), uintptr(Flags), uintptr(CodePage))
	return HRESULT(ret)
}

func (this *ITextDocument) Freeze(pCount *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Freeze, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pCount)))
	return HRESULT(ret)
}

func (this *ITextDocument) Unfreeze(pCount *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Unfreeze, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pCount)))
	return HRESULT(ret)
}

func (this *ITextDocument) BeginEditCollection() HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().BeginEditCollection, uintptr(unsafe.Pointer(this)))
	return HRESULT(ret)
}

func (this *ITextDocument) EndEditCollection() HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().EndEditCollection, uintptr(unsafe.Pointer(this)))
	return HRESULT(ret)
}

func (this *ITextDocument) Undo(Count int32, pCount *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Undo, uintptr(unsafe.Pointer(this)), uintptr(Count), uintptr(unsafe.Pointer(pCount)))
	return HRESULT(ret)
}

func (this *ITextDocument) Redo(Count int32, pCount *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Redo, uintptr(unsafe.Pointer(this)), uintptr(Count), uintptr(unsafe.Pointer(pCount)))
	return HRESULT(ret)
}

func (this *ITextDocument) Range(cpActive int32, cpAnchor int32, ppRange **ITextRange) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Range, uintptr(unsafe.Pointer(this)), uintptr(cpActive), uintptr(cpAnchor), uintptr(unsafe.Pointer(ppRange)))
	return HRESULT(ret)
}

func (this *ITextDocument) RangeFromPoint(x int32, y int32, ppRange **ITextRange) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().RangeFromPoint, uintptr(unsafe.Pointer(this)), uintptr(x), uintptr(y), uintptr(unsafe.Pointer(ppRange)))
	return HRESULT(ret)
}

// 8CC497C2-A1DF-11CE-8098-00AA0047BE5D
var IID_ITextRange = syscall.GUID{0x8CC497C2, 0xA1DF, 0x11CE,
	[8]byte{0x80, 0x98, 0x00, 0xAA, 0x00, 0x47, 0xBE, 0x5D}}

type ITextRangeInterface interface {
	IDispatchInterface
	GetText(pbstr *BSTR) HRESULT
	SetText(bstr BSTR) HRESULT
	GetChar(pChar *int32) HRESULT
	SetChar(Char int32) HRESULT
	GetDuplicate(ppRange **ITextRange) HRESULT
	GetFormattedText(ppRange **ITextRange) HRESULT
	SetFormattedText(pRange *ITextRange) HRESULT
	GetStart(pcpFirst *int32) HRESULT
	SetStart(cpFirst int32) HRESULT
	GetEnd(pcpLim *int32) HRESULT
	SetEnd(cpLim int32) HRESULT
	GetFont(ppFont **ITextFont) HRESULT
	SetFont(pFont *ITextFont) HRESULT
	GetPara(ppPara **ITextPara) HRESULT
	SetPara(pPara *ITextPara) HRESULT
	GetStoryLength(pCount *int32) HRESULT
	GetStoryType(pValue *int32) HRESULT
	Collapse(bStart int32) HRESULT
	Expand(Unit int32, pDelta *int32) HRESULT
	GetIndex(Unit int32, pIndex *int32) HRESULT
	SetIndex(Unit int32, Index int32, Extend int32) HRESULT
	SetRange(cpAnchor int32, cpActive int32) HRESULT
	InRange(pRange *ITextRange, pValue *int32) HRESULT
	InStory(pRange *ITextRange, pValue *int32) HRESULT
	IsEqual(pRange *ITextRange, pValue *int32) HRESULT
	Select() HRESULT
	StartOf(Unit int32, Extend int32, pDelta *int32) HRESULT
	EndOf(Unit int32, Extend int32, pDelta *int32) HRESULT
	Move(Unit int32, Count int32, pDelta *int32) HRESULT
	MoveStart(Unit int32, Count int32, pDelta *int32) HRESULT
	MoveEnd(Unit int32, Count int32, pDelta *int32) HRESULT
	MoveWhile(Cset *VARIANT, Count int32, pDelta *int32) HRESULT
	MoveStartWhile(Cset *VARIANT, Count int32, pDelta *int32) HRESULT
	MoveEndWhile(Cset *VARIANT, Count int32, pDelta *int32) HRESULT
	MoveUntil(Cset *VARIANT, Count int32, pDelta *int32) HRESULT
	MoveStartUntil(Cset *VARIANT, Count int32, pDelta *int32) HRESULT
	MoveEndUntil(Cset *VARIANT, Count int32, pDelta *int32) HRESULT
	FindText(bstr BSTR, Count int32, Flags int32, pLength *int32) HRESULT
	FindTextStart(bstr BSTR, Count int32, Flags int32, pLength *int32) HRESULT
	FindTextEnd(bstr BSTR, Count int32, Flags int32, pLength *int32) HRESULT
	Delete(Unit int32, Count int32, pDelta *int32) HRESULT
	Cut(pVar *VARIANT) HRESULT
	Copy(pVar *VARIANT) HRESULT
	Paste(pVar *VARIANT, Format int32) HRESULT
	CanPaste(pVar *VARIANT, Format int32, pValue *int32) HRESULT
	CanEdit(pValue *int32) HRESULT
	ChangeCase(Type int32) HRESULT
	GetPoint(Type int32, px *int32, py *int32) HRESULT
	SetPoint(x int32, y int32, Type int32, Extend int32) HRESULT
	ScrollIntoView(Value int32) HRESULT
	GetEmbeddedObject(ppObject **IUnknown) HRESULT
}

type ITextRangeVtbl struct {
	IDispatchVtbl
	GetText           uintptr
	SetText           uintptr
	GetChar           uintptr
	SetChar           uintptr
	GetDuplicate      uintptr
	GetFormattedText  uintptr
	SetFormattedText  uintptr
	GetStart          uintptr
	SetStart          uintptr
	GetEnd            uintptr
	SetEnd            uintptr
	GetFont           uintptr
	SetFont           uintptr
	GetPara           uintptr
	SetPara           uintptr
	GetStoryLength    uintptr
	GetStoryType      uintptr
	Collapse          uintptr
	Expand            uintptr
	GetIndex          uintptr
	SetIndex          uintptr
	SetRange          uintptr
	InRange           uintptr
	InStory           uintptr
	IsEqual           uintptr
	Select            uintptr
	StartOf           uintptr
	EndOf             uintptr
	Move              uintptr
	MoveStart         uintptr
	MoveEnd           uintptr
	MoveWhile         uintptr
	MoveStartWhile    uintptr
	MoveEndWhile      uintptr
	MoveUntil         uintptr
	MoveStartUntil    uintptr
	MoveEndUntil      uintptr
	FindText          uintptr
	FindTextStart     uintptr
	FindTextEnd       uintptr
	Delete            uintptr
	Cut               uintptr
	Copy              uintptr
	Paste             uintptr
	CanPaste          uintptr
	CanEdit           uintptr
	ChangeCase        uintptr
	GetPoint          uintptr
	SetPoint          uintptr
	ScrollIntoView    uintptr
	GetEmbeddedObject uintptr
}

type ITextRange struct {
	IDispatch
}

func (this *ITextRange) Vtbl() *ITextRangeVtbl {
	return (*ITextRangeVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *ITextRange) GetText(pbstr *BSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetText, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pbstr)))
	return HRESULT(ret)
}

func (this *ITextRange) SetText(bstr BSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetText, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(bstr)))
	return HRESULT(ret)
}

func (this *ITextRange) GetChar(pChar *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetChar, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pChar)))
	return HRESULT(ret)
}

func (this *ITextRange) SetChar(Char int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetChar, uintptr(unsafe.Pointer(this)), uintptr(Char))
	return HRESULT(ret)
}

func (this *ITextRange) GetDuplicate(ppRange **ITextRange) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetDuplicate, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(ppRange)))
	return HRESULT(ret)
}

func (this *ITextRange) GetFormattedText(ppRange **ITextRange) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetFormattedText, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(ppRange)))
	return HRESULT(ret)
}

func (this *ITextRange) SetFormattedText(pRange *ITextRange) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetFormattedText, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pRange)))
	return HRESULT(ret)
}

func (this *ITextRange) GetStart(pcpFirst *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetStart, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pcpFirst)))
	return HRESULT(ret)
}

func (this *ITextRange) SetStart(cpFirst int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetStart, uintptr(unsafe.Pointer(this)), uintptr(cpFirst))
	return HRESULT(ret)
}

func (this *ITextRange) GetEnd(pcpLim *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetEnd, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pcpLim)))
	return HRESULT(ret)
}

func (this *ITextRange) SetEnd(cpLim int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetEnd, uintptr(unsafe.Pointer(this)), uintptr(cpLim))
	return HRESULT(ret)
}

func (this *ITextRange) GetFont(ppFont **ITextFont) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetFont, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(ppFont)))
	return HRESULT(ret)
}

func (this *ITextRange) SetFont(pFont *ITextFont) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetFont, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pFont)))
	return HRESULT(ret)
}

func (this *ITextRange) GetPara(ppPara **ITextPara) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetPara, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(ppPara)))
	return HRESULT(ret)
}

func (this *ITextRange) SetPara(pPara *ITextPara) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetPara, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pPara)))
	return HRESULT(ret)
}

func (this *ITextRange) GetStoryLength(pCount *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetStoryLength, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pCount)))
	return HRESULT(ret)
}

func (this *ITextRange) GetStoryType(pValue *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetStoryType, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pValue)))
	return HRESULT(ret)
}

func (this *ITextRange) Collapse(bStart int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Collapse, uintptr(unsafe.Pointer(this)), uintptr(bStart))
	return HRESULT(ret)
}

func (this *ITextRange) Expand(Unit int32, pDelta *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Expand, uintptr(unsafe.Pointer(this)), uintptr(Unit), uintptr(unsafe.Pointer(pDelta)))
	return HRESULT(ret)
}

func (this *ITextRange) GetIndex(Unit int32, pIndex *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetIndex, uintptr(unsafe.Pointer(this)), uintptr(Unit), uintptr(unsafe.Pointer(pIndex)))
	return HRESULT(ret)
}

func (this *ITextRange) SetIndex(Unit int32, Index int32, Extend int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetIndex, uintptr(unsafe.Pointer(this)), uintptr(Unit), uintptr(Index), uintptr(Extend))
	return HRESULT(ret)
}

func (this *ITextRange) SetRange(cpAnchor int32, cpActive int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetRange, uintptr(unsafe.Pointer(this)), uintptr(cpAnchor), uintptr(cpActive))
	return HRESULT(ret)
}

func (this *ITextRange) InRange(pRange *ITextRange, pValue *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().InRange, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pRange)), uintptr(unsafe.Pointer(pValue)))
	return HRESULT(ret)
}

func (this *ITextRange) InStory(pRange *ITextRange, pValue *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().InStory, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pRange)), uintptr(unsafe.Pointer(pValue)))
	return HRESULT(ret)
}

func (this *ITextRange) IsEqual(pRange *ITextRange, pValue *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().IsEqual, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pRange)), uintptr(unsafe.Pointer(pValue)))
	return HRESULT(ret)
}

func (this *ITextRange) Select() HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Select, uintptr(unsafe.Pointer(this)))
	return HRESULT(ret)
}

func (this *ITextRange) StartOf(Unit int32, Extend int32, pDelta *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().StartOf, uintptr(unsafe.Pointer(this)), uintptr(Unit), uintptr(Extend), uintptr(unsafe.Pointer(pDelta)))
	return HRESULT(ret)
}

func (this *ITextRange) EndOf(Unit int32, Extend int32, pDelta *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().EndOf, uintptr(unsafe.Pointer(this)), uintptr(Unit), uintptr(Extend), uintptr(unsafe.Pointer(pDelta)))
	return HRESULT(ret)
}

func (this *ITextRange) Move(Unit int32, Count int32, pDelta *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Move, uintptr(unsafe.Pointer(this)), uintptr(Unit), uintptr(Count), uintptr(unsafe.Pointer(pDelta)))
	return HRESULT(ret)
}

func (this *ITextRange) MoveStart(Unit int32, Count int32, pDelta *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().MoveStart, uintptr(unsafe.Pointer(this)), uintptr(Unit), uintptr(Count), uintptr(unsafe.Pointer(pDelta)))
	return HRESULT(ret)
}

func (this *ITextRange) MoveEnd(Unit int32, Count int32, pDelta *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().MoveEnd, uintptr(unsafe.Pointer(this)), uintptr(Unit), uintptr(Count), uintptr(unsafe.Pointer(pDelta)))
	return HRESULT(ret)
}

func (this *ITextRange) MoveWhile(Cset *VARIANT, Count int32, pDelta *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().MoveWhile, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(Cset)), uintptr(Count), uintptr(unsafe.Pointer(pDelta)))
	return HRESULT(ret)
}

func (this *ITextRange) MoveStartWhile(Cset *VARIANT, Count int32, pDelta *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().MoveStartWhile, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(Cset)), uintptr(Count), uintptr(unsafe.Pointer(pDelta)))
	return HRESULT(ret)
}

func (this *ITextRange) MoveEndWhile(Cset *VARIANT, Count int32, pDelta *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().MoveEndWhile, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(Cset)), uintptr(Count), uintptr(unsafe.Pointer(pDelta)))
	return HRESULT(ret)
}

func (this *ITextRange) MoveUntil(Cset *VARIANT, Count int32, pDelta *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().MoveUntil, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(Cset)), uintptr(Count), uintptr(unsafe.Pointer(pDelta)))
	return HRESULT(ret)
}

func (this *ITextRange) MoveStartUntil(Cset *VARIANT, Count int32, pDelta *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().MoveStartUntil, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(Cset)), uintptr(Count), uintptr(unsafe.Pointer(pDelta)))
	return HRESULT(ret)
}

func (this *ITextRange) MoveEndUntil(Cset *VARIANT, Count int32, pDelta *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().MoveEndUntil, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(Cset)), uintptr(Count), uintptr(unsafe.Pointer(pDelta)))
	return HRESULT(ret)
}

func (this *ITextRange) FindText(bstr BSTR, Count int32, Flags int32, pLength *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().FindText, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(bstr)), uintptr(Count), uintptr(Flags), uintptr(unsafe.Pointer(pLength)))
	return HRESULT(ret)
}

func (this *ITextRange) FindTextStart(bstr BSTR, Count int32, Flags int32, pLength *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().FindTextStart, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(bstr)), uintptr(Count), uintptr(Flags), uintptr(unsafe.Pointer(pLength)))
	return HRESULT(ret)
}

func (this *ITextRange) FindTextEnd(bstr BSTR, Count int32, Flags int32, pLength *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().FindTextEnd, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(bstr)), uintptr(Count), uintptr(Flags), uintptr(unsafe.Pointer(pLength)))
	return HRESULT(ret)
}

func (this *ITextRange) Delete(Unit int32, Count int32, pDelta *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Delete, uintptr(unsafe.Pointer(this)), uintptr(Unit), uintptr(Count), uintptr(unsafe.Pointer(pDelta)))
	return HRESULT(ret)
}

func (this *ITextRange) Cut(pVar *VARIANT) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Cut, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pVar)))
	return HRESULT(ret)
}

func (this *ITextRange) Copy(pVar *VARIANT) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Copy, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pVar)))
	return HRESULT(ret)
}

func (this *ITextRange) Paste(pVar *VARIANT, Format int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Paste, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pVar)), uintptr(Format))
	return HRESULT(ret)
}

func (this *ITextRange) CanPaste(pVar *VARIANT, Format int32, pValue *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().CanPaste, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pVar)), uintptr(Format), uintptr(unsafe.Pointer(pValue)))
	return HRESULT(ret)
}

func (this *ITextRange) CanEdit(pValue *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().CanEdit, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pValue)))
	return HRESULT(ret)
}

func (this *ITextRange) ChangeCase(Type int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().ChangeCase, uintptr(unsafe.Pointer(this)), uintptr(Type))
	return HRESULT(ret)
}

func (this *ITextRange) GetPoint(Type int32, px *int32, py *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetPoint, uintptr(unsafe.Pointer(this)), uintptr(Type), uintptr(unsafe.Pointer(px)), uintptr(unsafe.Pointer(py)))
	return HRESULT(ret)
}

func (this *ITextRange) SetPoint(x int32, y int32, Type int32, Extend int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetPoint, uintptr(unsafe.Pointer(this)), uintptr(x), uintptr(y), uintptr(Type), uintptr(Extend))
	return HRESULT(ret)
}

func (this *ITextRange) ScrollIntoView(Value int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().ScrollIntoView, uintptr(unsafe.Pointer(this)), uintptr(Value))
	return HRESULT(ret)
}

func (this *ITextRange) GetEmbeddedObject(ppObject **IUnknown) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetEmbeddedObject, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(ppObject)))
	return HRESULT(ret)
}

// 8CC497C1-A1DF-11CE-8098-00AA0047BE5D
var IID_ITextSelection = syscall.GUID{0x8CC497C1, 0xA1DF, 0x11CE,
	[8]byte{0x80, 0x98, 0x00, 0xAA, 0x00, 0x47, 0xBE, 0x5D}}

type ITextSelectionInterface interface {
	ITextRangeInterface
	GetFlags(pFlags *int32) HRESULT
	SetFlags(Flags int32) HRESULT
	GetType(pType *int32) HRESULT
	MoveLeft(Unit int32, Count int32, Extend int32, pDelta *int32) HRESULT
	MoveRight(Unit int32, Count int32, Extend int32, pDelta *int32) HRESULT
	MoveUp(Unit int32, Count int32, Extend int32, pDelta *int32) HRESULT
	MoveDown(Unit int32, Count int32, Extend int32, pDelta *int32) HRESULT
	HomeKey(Unit TomConstants, Extend int32, pDelta *int32) HRESULT
	EndKey(Unit int32, Extend int32, pDelta *int32) HRESULT
	TypeText(bstr BSTR) HRESULT
}

type ITextSelectionVtbl struct {
	ITextRangeVtbl
	GetFlags  uintptr
	SetFlags  uintptr
	GetType   uintptr
	MoveLeft  uintptr
	MoveRight uintptr
	MoveUp    uintptr
	MoveDown  uintptr
	HomeKey   uintptr
	EndKey    uintptr
	TypeText  uintptr
}

type ITextSelection struct {
	ITextRange
}

func (this *ITextSelection) Vtbl() *ITextSelectionVtbl {
	return (*ITextSelectionVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *ITextSelection) GetFlags(pFlags *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetFlags, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pFlags)))
	return HRESULT(ret)
}

func (this *ITextSelection) SetFlags(Flags int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetFlags, uintptr(unsafe.Pointer(this)), uintptr(Flags))
	return HRESULT(ret)
}

func (this *ITextSelection) GetType(pType *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetType, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pType)))
	return HRESULT(ret)
}

func (this *ITextSelection) MoveLeft(Unit int32, Count int32, Extend int32, pDelta *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().MoveLeft, uintptr(unsafe.Pointer(this)), uintptr(Unit), uintptr(Count), uintptr(Extend), uintptr(unsafe.Pointer(pDelta)))
	return HRESULT(ret)
}

func (this *ITextSelection) MoveRight(Unit int32, Count int32, Extend int32, pDelta *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().MoveRight, uintptr(unsafe.Pointer(this)), uintptr(Unit), uintptr(Count), uintptr(Extend), uintptr(unsafe.Pointer(pDelta)))
	return HRESULT(ret)
}

func (this *ITextSelection) MoveUp(Unit int32, Count int32, Extend int32, pDelta *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().MoveUp, uintptr(unsafe.Pointer(this)), uintptr(Unit), uintptr(Count), uintptr(Extend), uintptr(unsafe.Pointer(pDelta)))
	return HRESULT(ret)
}

func (this *ITextSelection) MoveDown(Unit int32, Count int32, Extend int32, pDelta *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().MoveDown, uintptr(unsafe.Pointer(this)), uintptr(Unit), uintptr(Count), uintptr(Extend), uintptr(unsafe.Pointer(pDelta)))
	return HRESULT(ret)
}

func (this *ITextSelection) HomeKey(Unit TomConstants, Extend int32, pDelta *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().HomeKey, uintptr(unsafe.Pointer(this)), uintptr(Unit), uintptr(Extend), uintptr(unsafe.Pointer(pDelta)))
	return HRESULT(ret)
}

func (this *ITextSelection) EndKey(Unit int32, Extend int32, pDelta *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().EndKey, uintptr(unsafe.Pointer(this)), uintptr(Unit), uintptr(Extend), uintptr(unsafe.Pointer(pDelta)))
	return HRESULT(ret)
}

func (this *ITextSelection) TypeText(bstr BSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().TypeText, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(bstr)))
	return HRESULT(ret)
}

// 8CC497C3-A1DF-11CE-8098-00AA0047BE5D
var IID_ITextFont = syscall.GUID{0x8CC497C3, 0xA1DF, 0x11CE,
	[8]byte{0x80, 0x98, 0x00, 0xAA, 0x00, 0x47, 0xBE, 0x5D}}

type ITextFontInterface interface {
	IDispatchInterface
	GetDuplicate(ppFont **ITextFont) HRESULT
	SetDuplicate(pFont *ITextFont) HRESULT
	CanChange(pValue *int32) HRESULT
	IsEqual(pFont *ITextFont, pValue *int32) HRESULT
	Reset(Value TomConstants) HRESULT
	GetStyle(pValue *int32) HRESULT
	SetStyle(Value int32) HRESULT
	GetAllCaps(pValue *int32) HRESULT
	SetAllCaps(Value int32) HRESULT
	GetAnimation(pValue *int32) HRESULT
	SetAnimation(Value int32) HRESULT
	GetBackColor(pValue *int32) HRESULT
	SetBackColor(Value int32) HRESULT
	GetBold(pValue *int32) HRESULT
	SetBold(Value int32) HRESULT
	GetEmboss(pValue *int32) HRESULT
	SetEmboss(Value int32) HRESULT
	GetForeColor(pValue *int32) HRESULT
	SetForeColor(Value int32) HRESULT
	GetHidden(pValue *int32) HRESULT
	SetHidden(Value int32) HRESULT
	GetEngrave(pValue *int32) HRESULT
	SetEngrave(Value int32) HRESULT
	GetItalic(pValue *int32) HRESULT
	SetItalic(Value int32) HRESULT
	GetKerning(pValue *float32) HRESULT
	SetKerning(Value float32) HRESULT
	GetLanguageID(pValue *int32) HRESULT
	SetLanguageID(Value int32) HRESULT
	GetName(pbstr *BSTR) HRESULT
	SetName(bstr BSTR) HRESULT
	GetOutline(pValue *int32) HRESULT
	SetOutline(Value int32) HRESULT
	GetPosition(pValue *float32) HRESULT
	SetPosition(Value float32) HRESULT
	GetProtected(pValue *int32) HRESULT
	SetProtected(Value int32) HRESULT
	GetShadow(pValue *int32) HRESULT
	SetShadow(Value int32) HRESULT
	GetSize(pValue *float32) HRESULT
	SetSize(Value float32) HRESULT
	GetSmallCaps(pValue *int32) HRESULT
	SetSmallCaps(Value int32) HRESULT
	GetSpacing(pValue *float32) HRESULT
	SetSpacing(Value float32) HRESULT
	GetStrikeThrough(pValue *int32) HRESULT
	SetStrikeThrough(Value int32) HRESULT
	GetSubscript(pValue *int32) HRESULT
	SetSubscript(Value int32) HRESULT
	GetSuperscript(pValue *int32) HRESULT
	SetSuperscript(Value int32) HRESULT
	GetUnderline(pValue *int32) HRESULT
	SetUnderline(Value int32) HRESULT
	GetWeight(pValue *int32) HRESULT
	SetWeight(Value int32) HRESULT
}

type ITextFontVtbl struct {
	IDispatchVtbl
	GetDuplicate     uintptr
	SetDuplicate     uintptr
	CanChange        uintptr
	IsEqual          uintptr
	Reset            uintptr
	GetStyle         uintptr
	SetStyle         uintptr
	GetAllCaps       uintptr
	SetAllCaps       uintptr
	GetAnimation     uintptr
	SetAnimation     uintptr
	GetBackColor     uintptr
	SetBackColor     uintptr
	GetBold          uintptr
	SetBold          uintptr
	GetEmboss        uintptr
	SetEmboss        uintptr
	GetForeColor     uintptr
	SetForeColor     uintptr
	GetHidden        uintptr
	SetHidden        uintptr
	GetEngrave       uintptr
	SetEngrave       uintptr
	GetItalic        uintptr
	SetItalic        uintptr
	GetKerning       uintptr
	SetKerning       uintptr
	GetLanguageID    uintptr
	SetLanguageID    uintptr
	GetName          uintptr
	SetName          uintptr
	GetOutline       uintptr
	SetOutline       uintptr
	GetPosition      uintptr
	SetPosition      uintptr
	GetProtected     uintptr
	SetProtected     uintptr
	GetShadow        uintptr
	SetShadow        uintptr
	GetSize          uintptr
	SetSize          uintptr
	GetSmallCaps     uintptr
	SetSmallCaps     uintptr
	GetSpacing       uintptr
	SetSpacing       uintptr
	GetStrikeThrough uintptr
	SetStrikeThrough uintptr
	GetSubscript     uintptr
	SetSubscript     uintptr
	GetSuperscript   uintptr
	SetSuperscript   uintptr
	GetUnderline     uintptr
	SetUnderline     uintptr
	GetWeight        uintptr
	SetWeight        uintptr
}

type ITextFont struct {
	IDispatch
}

func (this *ITextFont) Vtbl() *ITextFontVtbl {
	return (*ITextFontVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *ITextFont) GetDuplicate(ppFont **ITextFont) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetDuplicate, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(ppFont)))
	return HRESULT(ret)
}

func (this *ITextFont) SetDuplicate(pFont *ITextFont) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetDuplicate, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pFont)))
	return HRESULT(ret)
}

func (this *ITextFont) CanChange(pValue *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().CanChange, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pValue)))
	return HRESULT(ret)
}

func (this *ITextFont) IsEqual(pFont *ITextFont, pValue *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().IsEqual, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pFont)), uintptr(unsafe.Pointer(pValue)))
	return HRESULT(ret)
}

func (this *ITextFont) Reset(Value TomConstants) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Reset, uintptr(unsafe.Pointer(this)), uintptr(Value))
	return HRESULT(ret)
}

func (this *ITextFont) GetStyle(pValue *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetStyle, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pValue)))
	return HRESULT(ret)
}

func (this *ITextFont) SetStyle(Value int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetStyle, uintptr(unsafe.Pointer(this)), uintptr(Value))
	return HRESULT(ret)
}

func (this *ITextFont) GetAllCaps(pValue *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetAllCaps, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pValue)))
	return HRESULT(ret)
}

func (this *ITextFont) SetAllCaps(Value int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetAllCaps, uintptr(unsafe.Pointer(this)), uintptr(Value))
	return HRESULT(ret)
}

func (this *ITextFont) GetAnimation(pValue *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetAnimation, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pValue)))
	return HRESULT(ret)
}

func (this *ITextFont) SetAnimation(Value int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetAnimation, uintptr(unsafe.Pointer(this)), uintptr(Value))
	return HRESULT(ret)
}

func (this *ITextFont) GetBackColor(pValue *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetBackColor, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pValue)))
	return HRESULT(ret)
}

func (this *ITextFont) SetBackColor(Value int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetBackColor, uintptr(unsafe.Pointer(this)), uintptr(Value))
	return HRESULT(ret)
}

func (this *ITextFont) GetBold(pValue *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetBold, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pValue)))
	return HRESULT(ret)
}

func (this *ITextFont) SetBold(Value int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetBold, uintptr(unsafe.Pointer(this)), uintptr(Value))
	return HRESULT(ret)
}

func (this *ITextFont) GetEmboss(pValue *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetEmboss, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pValue)))
	return HRESULT(ret)
}

func (this *ITextFont) SetEmboss(Value int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetEmboss, uintptr(unsafe.Pointer(this)), uintptr(Value))
	return HRESULT(ret)
}

func (this *ITextFont) GetForeColor(pValue *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetForeColor, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pValue)))
	return HRESULT(ret)
}

func (this *ITextFont) SetForeColor(Value int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetForeColor, uintptr(unsafe.Pointer(this)), uintptr(Value))
	return HRESULT(ret)
}

func (this *ITextFont) GetHidden(pValue *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetHidden, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pValue)))
	return HRESULT(ret)
}

func (this *ITextFont) SetHidden(Value int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetHidden, uintptr(unsafe.Pointer(this)), uintptr(Value))
	return HRESULT(ret)
}

func (this *ITextFont) GetEngrave(pValue *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetEngrave, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pValue)))
	return HRESULT(ret)
}

func (this *ITextFont) SetEngrave(Value int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetEngrave, uintptr(unsafe.Pointer(this)), uintptr(Value))
	return HRESULT(ret)
}

func (this *ITextFont) GetItalic(pValue *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetItalic, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pValue)))
	return HRESULT(ret)
}

func (this *ITextFont) SetItalic(Value int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetItalic, uintptr(unsafe.Pointer(this)), uintptr(Value))
	return HRESULT(ret)
}

func (this *ITextFont) GetKerning(pValue *float32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetKerning, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pValue)))
	return HRESULT(ret)
}

func (this *ITextFont) SetKerning(Value float32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetKerning, uintptr(unsafe.Pointer(this)), uintptr(Value))
	return HRESULT(ret)
}

func (this *ITextFont) GetLanguageID(pValue *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetLanguageID, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pValue)))
	return HRESULT(ret)
}

func (this *ITextFont) SetLanguageID(Value int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetLanguageID, uintptr(unsafe.Pointer(this)), uintptr(Value))
	return HRESULT(ret)
}

func (this *ITextFont) GetName(pbstr *BSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetName, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pbstr)))
	return HRESULT(ret)
}

func (this *ITextFont) SetName(bstr BSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetName, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(bstr)))
	return HRESULT(ret)
}

func (this *ITextFont) GetOutline(pValue *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetOutline, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pValue)))
	return HRESULT(ret)
}

func (this *ITextFont) SetOutline(Value int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetOutline, uintptr(unsafe.Pointer(this)), uintptr(Value))
	return HRESULT(ret)
}

func (this *ITextFont) GetPosition(pValue *float32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetPosition, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pValue)))
	return HRESULT(ret)
}

func (this *ITextFont) SetPosition(Value float32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetPosition, uintptr(unsafe.Pointer(this)), uintptr(Value))
	return HRESULT(ret)
}

func (this *ITextFont) GetProtected(pValue *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetProtected, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pValue)))
	return HRESULT(ret)
}

func (this *ITextFont) SetProtected(Value int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetProtected, uintptr(unsafe.Pointer(this)), uintptr(Value))
	return HRESULT(ret)
}

func (this *ITextFont) GetShadow(pValue *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetShadow, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pValue)))
	return HRESULT(ret)
}

func (this *ITextFont) SetShadow(Value int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetShadow, uintptr(unsafe.Pointer(this)), uintptr(Value))
	return HRESULT(ret)
}

func (this *ITextFont) GetSize(pValue *float32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetSize, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pValue)))
	return HRESULT(ret)
}

func (this *ITextFont) SetSize(Value float32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetSize, uintptr(unsafe.Pointer(this)), uintptr(Value))
	return HRESULT(ret)
}

func (this *ITextFont) GetSmallCaps(pValue *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetSmallCaps, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pValue)))
	return HRESULT(ret)
}

func (this *ITextFont) SetSmallCaps(Value int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetSmallCaps, uintptr(unsafe.Pointer(this)), uintptr(Value))
	return HRESULT(ret)
}

func (this *ITextFont) GetSpacing(pValue *float32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetSpacing, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pValue)))
	return HRESULT(ret)
}

func (this *ITextFont) SetSpacing(Value float32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetSpacing, uintptr(unsafe.Pointer(this)), uintptr(Value))
	return HRESULT(ret)
}

func (this *ITextFont) GetStrikeThrough(pValue *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetStrikeThrough, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pValue)))
	return HRESULT(ret)
}

func (this *ITextFont) SetStrikeThrough(Value int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetStrikeThrough, uintptr(unsafe.Pointer(this)), uintptr(Value))
	return HRESULT(ret)
}

func (this *ITextFont) GetSubscript(pValue *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetSubscript, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pValue)))
	return HRESULT(ret)
}

func (this *ITextFont) SetSubscript(Value int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetSubscript, uintptr(unsafe.Pointer(this)), uintptr(Value))
	return HRESULT(ret)
}

func (this *ITextFont) GetSuperscript(pValue *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetSuperscript, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pValue)))
	return HRESULT(ret)
}

func (this *ITextFont) SetSuperscript(Value int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetSuperscript, uintptr(unsafe.Pointer(this)), uintptr(Value))
	return HRESULT(ret)
}

func (this *ITextFont) GetUnderline(pValue *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetUnderline, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pValue)))
	return HRESULT(ret)
}

func (this *ITextFont) SetUnderline(Value int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetUnderline, uintptr(unsafe.Pointer(this)), uintptr(Value))
	return HRESULT(ret)
}

func (this *ITextFont) GetWeight(pValue *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetWeight, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pValue)))
	return HRESULT(ret)
}

func (this *ITextFont) SetWeight(Value int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetWeight, uintptr(unsafe.Pointer(this)), uintptr(Value))
	return HRESULT(ret)
}

// 8CC497C4-A1DF-11CE-8098-00AA0047BE5D
var IID_ITextPara = syscall.GUID{0x8CC497C4, 0xA1DF, 0x11CE,
	[8]byte{0x80, 0x98, 0x00, 0xAA, 0x00, 0x47, 0xBE, 0x5D}}

type ITextParaInterface interface {
	IDispatchInterface
	GetDuplicate(ppPara **ITextPara) HRESULT
	SetDuplicate(pPara *ITextPara) HRESULT
	CanChange(pValue *int32) HRESULT
	IsEqual(pPara *ITextPara, pValue *int32) HRESULT
	Reset(Value int32) HRESULT
	GetStyle(pValue *int32) HRESULT
	SetStyle(Value int32) HRESULT
	GetAlignment(pValue *int32) HRESULT
	SetAlignment(Value int32) HRESULT
	GetHyphenation(pValue *TomConstants) HRESULT
	SetHyphenation(Value int32) HRESULT
	GetFirstLineIndent(pValue *float32) HRESULT
	GetKeepTogether(pValue *TomConstants) HRESULT
	SetKeepTogether(Value int32) HRESULT
	GetKeepWithNext(pValue *TomConstants) HRESULT
	SetKeepWithNext(Value int32) HRESULT
	GetLeftIndent(pValue *float32) HRESULT
	GetLineSpacing(pValue *float32) HRESULT
	GetLineSpacingRule(pValue *int32) HRESULT
	GetListAlignment(pValue *int32) HRESULT
	SetListAlignment(Value int32) HRESULT
	GetListLevelIndex(pValue *int32) HRESULT
	SetListLevelIndex(Value int32) HRESULT
	GetListStart(pValue *int32) HRESULT
	SetListStart(Value int32) HRESULT
	GetListTab(pValue *float32) HRESULT
	SetListTab(Value float32) HRESULT
	GetListType(pValue *int32) HRESULT
	SetListType(Value int32) HRESULT
	GetNoLineNumber(pValue *int32) HRESULT
	SetNoLineNumber(Value int32) HRESULT
	GetPageBreakBefore(pValue *int32) HRESULT
	SetPageBreakBefore(Value int32) HRESULT
	GetRightIndent(pValue *float32) HRESULT
	SetRightIndent(Value float32) HRESULT
	SetIndents(First float32, Left float32, Right float32) HRESULT
	SetLineSpacing(Rule int32, Spacing float32) HRESULT
	GetSpaceAfter(pValue *float32) HRESULT
	SetSpaceAfter(Value float32) HRESULT
	GetSpaceBefore(pValue *float32) HRESULT
	SetSpaceBefore(Value float32) HRESULT
	GetWidowControl(pValue *int32) HRESULT
	SetWidowControl(Value int32) HRESULT
	GetTabCount(pCount *int32) HRESULT
	AddTab(tbPos float32, tbAlign int32, tbLeader int32) HRESULT
	ClearAllTabs() HRESULT
	DeleteTab(tbPos float32) HRESULT
	GetTab(iTab int32, ptbPos *float32, ptbAlign *int32, ptbLeader *int32) HRESULT
}

type ITextParaVtbl struct {
	IDispatchVtbl
	GetDuplicate       uintptr
	SetDuplicate       uintptr
	CanChange          uintptr
	IsEqual            uintptr
	Reset              uintptr
	GetStyle           uintptr
	SetStyle           uintptr
	GetAlignment       uintptr
	SetAlignment       uintptr
	GetHyphenation     uintptr
	SetHyphenation     uintptr
	GetFirstLineIndent uintptr
	GetKeepTogether    uintptr
	SetKeepTogether    uintptr
	GetKeepWithNext    uintptr
	SetKeepWithNext    uintptr
	GetLeftIndent      uintptr
	GetLineSpacing     uintptr
	GetLineSpacingRule uintptr
	GetListAlignment   uintptr
	SetListAlignment   uintptr
	GetListLevelIndex  uintptr
	SetListLevelIndex  uintptr
	GetListStart       uintptr
	SetListStart       uintptr
	GetListTab         uintptr
	SetListTab         uintptr
	GetListType        uintptr
	SetListType        uintptr
	GetNoLineNumber    uintptr
	SetNoLineNumber    uintptr
	GetPageBreakBefore uintptr
	SetPageBreakBefore uintptr
	GetRightIndent     uintptr
	SetRightIndent     uintptr
	SetIndents         uintptr
	SetLineSpacing     uintptr
	GetSpaceAfter      uintptr
	SetSpaceAfter      uintptr
	GetSpaceBefore     uintptr
	SetSpaceBefore     uintptr
	GetWidowControl    uintptr
	SetWidowControl    uintptr
	GetTabCount        uintptr
	AddTab             uintptr
	ClearAllTabs       uintptr
	DeleteTab          uintptr
	GetTab             uintptr
}

type ITextPara struct {
	IDispatch
}

func (this *ITextPara) Vtbl() *ITextParaVtbl {
	return (*ITextParaVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *ITextPara) GetDuplicate(ppPara **ITextPara) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetDuplicate, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(ppPara)))
	return HRESULT(ret)
}

func (this *ITextPara) SetDuplicate(pPara *ITextPara) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetDuplicate, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pPara)))
	return HRESULT(ret)
}

func (this *ITextPara) CanChange(pValue *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().CanChange, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pValue)))
	return HRESULT(ret)
}

func (this *ITextPara) IsEqual(pPara *ITextPara, pValue *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().IsEqual, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pPara)), uintptr(unsafe.Pointer(pValue)))
	return HRESULT(ret)
}

func (this *ITextPara) Reset(Value int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Reset, uintptr(unsafe.Pointer(this)), uintptr(Value))
	return HRESULT(ret)
}

func (this *ITextPara) GetStyle(pValue *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetStyle, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pValue)))
	return HRESULT(ret)
}

func (this *ITextPara) SetStyle(Value int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetStyle, uintptr(unsafe.Pointer(this)), uintptr(Value))
	return HRESULT(ret)
}

func (this *ITextPara) GetAlignment(pValue *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetAlignment, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pValue)))
	return HRESULT(ret)
}

func (this *ITextPara) SetAlignment(Value int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetAlignment, uintptr(unsafe.Pointer(this)), uintptr(Value))
	return HRESULT(ret)
}

func (this *ITextPara) GetHyphenation(pValue *TomConstants) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetHyphenation, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pValue)))
	return HRESULT(ret)
}

func (this *ITextPara) SetHyphenation(Value int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetHyphenation, uintptr(unsafe.Pointer(this)), uintptr(Value))
	return HRESULT(ret)
}

func (this *ITextPara) GetFirstLineIndent(pValue *float32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetFirstLineIndent, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pValue)))
	return HRESULT(ret)
}

func (this *ITextPara) GetKeepTogether(pValue *TomConstants) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetKeepTogether, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pValue)))
	return HRESULT(ret)
}

func (this *ITextPara) SetKeepTogether(Value int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetKeepTogether, uintptr(unsafe.Pointer(this)), uintptr(Value))
	return HRESULT(ret)
}

func (this *ITextPara) GetKeepWithNext(pValue *TomConstants) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetKeepWithNext, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pValue)))
	return HRESULT(ret)
}

func (this *ITextPara) SetKeepWithNext(Value int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetKeepWithNext, uintptr(unsafe.Pointer(this)), uintptr(Value))
	return HRESULT(ret)
}

func (this *ITextPara) GetLeftIndent(pValue *float32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetLeftIndent, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pValue)))
	return HRESULT(ret)
}

func (this *ITextPara) GetLineSpacing(pValue *float32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetLineSpacing, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pValue)))
	return HRESULT(ret)
}

func (this *ITextPara) GetLineSpacingRule(pValue *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetLineSpacingRule, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pValue)))
	return HRESULT(ret)
}

func (this *ITextPara) GetListAlignment(pValue *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetListAlignment, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pValue)))
	return HRESULT(ret)
}

func (this *ITextPara) SetListAlignment(Value int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetListAlignment, uintptr(unsafe.Pointer(this)), uintptr(Value))
	return HRESULT(ret)
}

func (this *ITextPara) GetListLevelIndex(pValue *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetListLevelIndex, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pValue)))
	return HRESULT(ret)
}

func (this *ITextPara) SetListLevelIndex(Value int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetListLevelIndex, uintptr(unsafe.Pointer(this)), uintptr(Value))
	return HRESULT(ret)
}

func (this *ITextPara) GetListStart(pValue *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetListStart, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pValue)))
	return HRESULT(ret)
}

func (this *ITextPara) SetListStart(Value int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetListStart, uintptr(unsafe.Pointer(this)), uintptr(Value))
	return HRESULT(ret)
}

func (this *ITextPara) GetListTab(pValue *float32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetListTab, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pValue)))
	return HRESULT(ret)
}

func (this *ITextPara) SetListTab(Value float32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetListTab, uintptr(unsafe.Pointer(this)), uintptr(Value))
	return HRESULT(ret)
}

func (this *ITextPara) GetListType(pValue *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetListType, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pValue)))
	return HRESULT(ret)
}

func (this *ITextPara) SetListType(Value int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetListType, uintptr(unsafe.Pointer(this)), uintptr(Value))
	return HRESULT(ret)
}

func (this *ITextPara) GetNoLineNumber(pValue *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetNoLineNumber, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pValue)))
	return HRESULT(ret)
}

func (this *ITextPara) SetNoLineNumber(Value int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetNoLineNumber, uintptr(unsafe.Pointer(this)), uintptr(Value))
	return HRESULT(ret)
}

func (this *ITextPara) GetPageBreakBefore(pValue *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetPageBreakBefore, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pValue)))
	return HRESULT(ret)
}

func (this *ITextPara) SetPageBreakBefore(Value int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetPageBreakBefore, uintptr(unsafe.Pointer(this)), uintptr(Value))
	return HRESULT(ret)
}

func (this *ITextPara) GetRightIndent(pValue *float32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetRightIndent, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pValue)))
	return HRESULT(ret)
}

func (this *ITextPara) SetRightIndent(Value float32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetRightIndent, uintptr(unsafe.Pointer(this)), uintptr(Value))
	return HRESULT(ret)
}

func (this *ITextPara) SetIndents(First float32, Left float32, Right float32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetIndents, uintptr(unsafe.Pointer(this)), uintptr(First), uintptr(Left), uintptr(Right))
	return HRESULT(ret)
}

func (this *ITextPara) SetLineSpacing(Rule int32, Spacing float32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetLineSpacing, uintptr(unsafe.Pointer(this)), uintptr(Rule), uintptr(Spacing))
	return HRESULT(ret)
}

func (this *ITextPara) GetSpaceAfter(pValue *float32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetSpaceAfter, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pValue)))
	return HRESULT(ret)
}

func (this *ITextPara) SetSpaceAfter(Value float32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetSpaceAfter, uintptr(unsafe.Pointer(this)), uintptr(Value))
	return HRESULT(ret)
}

func (this *ITextPara) GetSpaceBefore(pValue *float32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetSpaceBefore, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pValue)))
	return HRESULT(ret)
}

func (this *ITextPara) SetSpaceBefore(Value float32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetSpaceBefore, uintptr(unsafe.Pointer(this)), uintptr(Value))
	return HRESULT(ret)
}

func (this *ITextPara) GetWidowControl(pValue *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetWidowControl, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pValue)))
	return HRESULT(ret)
}

func (this *ITextPara) SetWidowControl(Value int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetWidowControl, uintptr(unsafe.Pointer(this)), uintptr(Value))
	return HRESULT(ret)
}

func (this *ITextPara) GetTabCount(pCount *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetTabCount, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pCount)))
	return HRESULT(ret)
}

func (this *ITextPara) AddTab(tbPos float32, tbAlign int32, tbLeader int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().AddTab, uintptr(unsafe.Pointer(this)), uintptr(tbPos), uintptr(tbAlign), uintptr(tbLeader))
	return HRESULT(ret)
}

func (this *ITextPara) ClearAllTabs() HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().ClearAllTabs, uintptr(unsafe.Pointer(this)))
	return HRESULT(ret)
}

func (this *ITextPara) DeleteTab(tbPos float32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().DeleteTab, uintptr(unsafe.Pointer(this)), uintptr(tbPos))
	return HRESULT(ret)
}

func (this *ITextPara) GetTab(iTab int32, ptbPos *float32, ptbAlign *int32, ptbLeader *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetTab, uintptr(unsafe.Pointer(this)), uintptr(iTab), uintptr(unsafe.Pointer(ptbPos)), uintptr(unsafe.Pointer(ptbAlign)), uintptr(unsafe.Pointer(ptbLeader)))
	return HRESULT(ret)
}

// 8CC497C5-A1DF-11CE-8098-00AA0047BE5D
var IID_ITextStoryRanges = syscall.GUID{0x8CC497C5, 0xA1DF, 0x11CE,
	[8]byte{0x80, 0x98, 0x00, 0xAA, 0x00, 0x47, 0xBE, 0x5D}}

type ITextStoryRangesInterface interface {
	IDispatchInterface
	NewEnum_(ppunkEnum **IUnknown) HRESULT
	Item(Index int32, ppRange **ITextRange) HRESULT
	GetCount(pCount *int32) HRESULT
}

type ITextStoryRangesVtbl struct {
	IDispatchVtbl
	NewEnum_ uintptr
	Item     uintptr
	GetCount uintptr
}

type ITextStoryRanges struct {
	IDispatch
}

func (this *ITextStoryRanges) Vtbl() *ITextStoryRangesVtbl {
	return (*ITextStoryRangesVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *ITextStoryRanges) NewEnum_(ppunkEnum **IUnknown) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().NewEnum_, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(ppunkEnum)))
	return HRESULT(ret)
}

func (this *ITextStoryRanges) Item(Index int32, ppRange **ITextRange) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Item, uintptr(unsafe.Pointer(this)), uintptr(Index), uintptr(unsafe.Pointer(ppRange)))
	return HRESULT(ret)
}

func (this *ITextStoryRanges) GetCount(pCount *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetCount, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pCount)))
	return HRESULT(ret)
}

// C241F5E0-7206-11D8-A2C7-00A0D1D6C6B3
var IID_ITextDocument2 = syscall.GUID{0xC241F5E0, 0x7206, 0x11D8,
	[8]byte{0xA2, 0xC7, 0x00, 0xA0, 0xD1, 0xD6, 0xC6, 0xB3}}

type ITextDocument2Interface interface {
	ITextDocumentInterface
	GetCaretType(pValue *int32) HRESULT
	SetCaretType(Value int32) HRESULT
	GetDisplays(ppDisplays **ITextDisplays) HRESULT
	GetDocumentFont(ppFont **ITextFont2) HRESULT
	SetDocumentFont(pFont *ITextFont2) HRESULT
	GetDocumentPara(ppPara **ITextPara2) HRESULT
	SetDocumentPara(pPara *ITextPara2) HRESULT
	GetEastAsianFlags(pFlags *TomConstants) HRESULT
	GetGenerator(pbstr *BSTR) HRESULT
	SetIMEInProgress(Value int32) HRESULT
	GetNotificationMode(pValue *int32) HRESULT
	SetNotificationMode(Value int32) HRESULT
	GetSelection2(ppSel **ITextSelection2) HRESULT
	GetStoryRanges2(ppStories **ITextStoryRanges2) HRESULT
	GetTypographyOptions(pOptions *int32) HRESULT
	GetVersion(pValue *int32) HRESULT
	GetWindow(pHwnd *int64) HRESULT
	AttachMsgFilter(pFilter *IUnknown) HRESULT
	CheckTextLimit(cch int32, pcch *int32) HRESULT
	GetCallManager(ppVoid **IUnknown) HRESULT
	GetClientRect(Type TomConstants, pLeft *int32, pTop *int32, pRight *int32, pBottom *int32) HRESULT
	GetEffectColor(Index int32, pValue *int32) HRESULT
	GetImmContext(pContext *int64) HRESULT
	GetPreferredFont(cp int32, CharRep int32, Options int32, curCharRep int32, curFontSize int32, pbstr *BSTR, pPitchAndFamily *int32, pNewFontSize *int32) HRESULT
	GetProperty(Type int32, pValue *int32) HRESULT
	GetStrings(ppStrs **ITextStrings) HRESULT
	Notify(Notify int32) HRESULT
	Range2(cpActive int32, cpAnchor int32, ppRange **ITextRange2) HRESULT
	RangeFromPoint2(x int32, y int32, Type int32, ppRange **ITextRange2) HRESULT
	ReleaseCallManager(pVoid *IUnknown) HRESULT
	ReleaseImmContext(Context int64) HRESULT
	SetEffectColor(Index int32, Value int32) HRESULT
	SetProperty(Type int32, Value int32) HRESULT
	SetTypographyOptions(Options int32, Mask int32) HRESULT
	SysBeep() HRESULT
	Update(Value int32) HRESULT
	UpdateWindow() HRESULT
	GetMathProperties(pOptions *int32) HRESULT
	SetMathProperties(Options int32, Mask int32) HRESULT
	GetActiveStory(ppStory **ITextStory) HRESULT
	SetActiveStory(pStory *ITextStory) HRESULT
	GetMainStory(ppStory **ITextStory) HRESULT
	GetNewStory(ppStory **ITextStory) HRESULT
	GetStory(Index int32, ppStory **ITextStory) HRESULT
}

type ITextDocument2Vtbl struct {
	ITextDocumentVtbl
	GetCaretType         uintptr
	SetCaretType         uintptr
	GetDisplays          uintptr
	GetDocumentFont      uintptr
	SetDocumentFont      uintptr
	GetDocumentPara      uintptr
	SetDocumentPara      uintptr
	GetEastAsianFlags    uintptr
	GetGenerator         uintptr
	SetIMEInProgress     uintptr
	GetNotificationMode  uintptr
	SetNotificationMode  uintptr
	GetSelection2        uintptr
	GetStoryRanges2      uintptr
	GetTypographyOptions uintptr
	GetVersion           uintptr
	GetWindow            uintptr
	AttachMsgFilter      uintptr
	CheckTextLimit       uintptr
	GetCallManager       uintptr
	GetClientRect        uintptr
	GetEffectColor       uintptr
	GetImmContext        uintptr
	GetPreferredFont     uintptr
	GetProperty          uintptr
	GetStrings           uintptr
	Notify               uintptr
	Range2               uintptr
	RangeFromPoint2      uintptr
	ReleaseCallManager   uintptr
	ReleaseImmContext    uintptr
	SetEffectColor       uintptr
	SetProperty          uintptr
	SetTypographyOptions uintptr
	SysBeep              uintptr
	Update               uintptr
	UpdateWindow         uintptr
	GetMathProperties    uintptr
	SetMathProperties    uintptr
	GetActiveStory       uintptr
	SetActiveStory       uintptr
	GetMainStory         uintptr
	GetNewStory          uintptr
	GetStory             uintptr
}

type ITextDocument2 struct {
	ITextDocument
}

func (this *ITextDocument2) Vtbl() *ITextDocument2Vtbl {
	return (*ITextDocument2Vtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *ITextDocument2) GetCaretType(pValue *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetCaretType, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pValue)))
	return HRESULT(ret)
}

func (this *ITextDocument2) SetCaretType(Value int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetCaretType, uintptr(unsafe.Pointer(this)), uintptr(Value))
	return HRESULT(ret)
}

func (this *ITextDocument2) GetDisplays(ppDisplays **ITextDisplays) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetDisplays, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(ppDisplays)))
	return HRESULT(ret)
}

func (this *ITextDocument2) GetDocumentFont(ppFont **ITextFont2) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetDocumentFont, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(ppFont)))
	return HRESULT(ret)
}

func (this *ITextDocument2) SetDocumentFont(pFont *ITextFont2) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetDocumentFont, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pFont)))
	return HRESULT(ret)
}

func (this *ITextDocument2) GetDocumentPara(ppPara **ITextPara2) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetDocumentPara, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(ppPara)))
	return HRESULT(ret)
}

func (this *ITextDocument2) SetDocumentPara(pPara *ITextPara2) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetDocumentPara, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pPara)))
	return HRESULT(ret)
}

func (this *ITextDocument2) GetEastAsianFlags(pFlags *TomConstants) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetEastAsianFlags, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pFlags)))
	return HRESULT(ret)
}

func (this *ITextDocument2) GetGenerator(pbstr *BSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetGenerator, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pbstr)))
	return HRESULT(ret)
}

func (this *ITextDocument2) SetIMEInProgress(Value int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetIMEInProgress, uintptr(unsafe.Pointer(this)), uintptr(Value))
	return HRESULT(ret)
}

func (this *ITextDocument2) GetNotificationMode(pValue *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetNotificationMode, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pValue)))
	return HRESULT(ret)
}

func (this *ITextDocument2) SetNotificationMode(Value int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetNotificationMode, uintptr(unsafe.Pointer(this)), uintptr(Value))
	return HRESULT(ret)
}

func (this *ITextDocument2) GetSelection2(ppSel **ITextSelection2) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetSelection2, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(ppSel)))
	return HRESULT(ret)
}

func (this *ITextDocument2) GetStoryRanges2(ppStories **ITextStoryRanges2) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetStoryRanges2, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(ppStories)))
	return HRESULT(ret)
}

func (this *ITextDocument2) GetTypographyOptions(pOptions *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetTypographyOptions, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pOptions)))
	return HRESULT(ret)
}

func (this *ITextDocument2) GetVersion(pValue *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetVersion, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pValue)))
	return HRESULT(ret)
}

func (this *ITextDocument2) GetWindow(pHwnd *int64) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetWindow, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pHwnd)))
	return HRESULT(ret)
}

func (this *ITextDocument2) AttachMsgFilter(pFilter *IUnknown) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().AttachMsgFilter, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pFilter)))
	return HRESULT(ret)
}

func (this *ITextDocument2) CheckTextLimit(cch int32, pcch *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().CheckTextLimit, uintptr(unsafe.Pointer(this)), uintptr(cch), uintptr(unsafe.Pointer(pcch)))
	return HRESULT(ret)
}

func (this *ITextDocument2) GetCallManager(ppVoid **IUnknown) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetCallManager, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(ppVoid)))
	return HRESULT(ret)
}

func (this *ITextDocument2) GetClientRect(Type TomConstants, pLeft *int32, pTop *int32, pRight *int32, pBottom *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetClientRect, uintptr(unsafe.Pointer(this)), uintptr(Type), uintptr(unsafe.Pointer(pLeft)), uintptr(unsafe.Pointer(pTop)), uintptr(unsafe.Pointer(pRight)), uintptr(unsafe.Pointer(pBottom)))
	return HRESULT(ret)
}

func (this *ITextDocument2) GetEffectColor(Index int32, pValue *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetEffectColor, uintptr(unsafe.Pointer(this)), uintptr(Index), uintptr(unsafe.Pointer(pValue)))
	return HRESULT(ret)
}

func (this *ITextDocument2) GetImmContext(pContext *int64) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetImmContext, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pContext)))
	return HRESULT(ret)
}

func (this *ITextDocument2) GetPreferredFont(cp int32, CharRep int32, Options int32, curCharRep int32, curFontSize int32, pbstr *BSTR, pPitchAndFamily *int32, pNewFontSize *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetPreferredFont, uintptr(unsafe.Pointer(this)), uintptr(cp), uintptr(CharRep), uintptr(Options), uintptr(curCharRep), uintptr(curFontSize), uintptr(unsafe.Pointer(pbstr)), uintptr(unsafe.Pointer(pPitchAndFamily)), uintptr(unsafe.Pointer(pNewFontSize)))
	return HRESULT(ret)
}

func (this *ITextDocument2) GetProperty(Type int32, pValue *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetProperty, uintptr(unsafe.Pointer(this)), uintptr(Type), uintptr(unsafe.Pointer(pValue)))
	return HRESULT(ret)
}

func (this *ITextDocument2) GetStrings(ppStrs **ITextStrings) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetStrings, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(ppStrs)))
	return HRESULT(ret)
}

func (this *ITextDocument2) Notify(Notify int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Notify, uintptr(unsafe.Pointer(this)), uintptr(Notify))
	return HRESULT(ret)
}

func (this *ITextDocument2) Range2(cpActive int32, cpAnchor int32, ppRange **ITextRange2) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Range2, uintptr(unsafe.Pointer(this)), uintptr(cpActive), uintptr(cpAnchor), uintptr(unsafe.Pointer(ppRange)))
	return HRESULT(ret)
}

func (this *ITextDocument2) RangeFromPoint2(x int32, y int32, Type int32, ppRange **ITextRange2) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().RangeFromPoint2, uintptr(unsafe.Pointer(this)), uintptr(x), uintptr(y), uintptr(Type), uintptr(unsafe.Pointer(ppRange)))
	return HRESULT(ret)
}

func (this *ITextDocument2) ReleaseCallManager(pVoid *IUnknown) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().ReleaseCallManager, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pVoid)))
	return HRESULT(ret)
}

func (this *ITextDocument2) ReleaseImmContext(Context int64) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().ReleaseImmContext, uintptr(unsafe.Pointer(this)), uintptr(Context))
	return HRESULT(ret)
}

func (this *ITextDocument2) SetEffectColor(Index int32, Value int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetEffectColor, uintptr(unsafe.Pointer(this)), uintptr(Index), uintptr(Value))
	return HRESULT(ret)
}

func (this *ITextDocument2) SetProperty(Type int32, Value int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetProperty, uintptr(unsafe.Pointer(this)), uintptr(Type), uintptr(Value))
	return HRESULT(ret)
}

func (this *ITextDocument2) SetTypographyOptions(Options int32, Mask int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetTypographyOptions, uintptr(unsafe.Pointer(this)), uintptr(Options), uintptr(Mask))
	return HRESULT(ret)
}

func (this *ITextDocument2) SysBeep() HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SysBeep, uintptr(unsafe.Pointer(this)))
	return HRESULT(ret)
}

func (this *ITextDocument2) Update(Value int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Update, uintptr(unsafe.Pointer(this)), uintptr(Value))
	return HRESULT(ret)
}

func (this *ITextDocument2) UpdateWindow() HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().UpdateWindow, uintptr(unsafe.Pointer(this)))
	return HRESULT(ret)
}

func (this *ITextDocument2) GetMathProperties(pOptions *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetMathProperties, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pOptions)))
	return HRESULT(ret)
}

func (this *ITextDocument2) SetMathProperties(Options int32, Mask int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetMathProperties, uintptr(unsafe.Pointer(this)), uintptr(Options), uintptr(Mask))
	return HRESULT(ret)
}

func (this *ITextDocument2) GetActiveStory(ppStory **ITextStory) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetActiveStory, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(ppStory)))
	return HRESULT(ret)
}

func (this *ITextDocument2) SetActiveStory(pStory *ITextStory) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetActiveStory, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pStory)))
	return HRESULT(ret)
}

func (this *ITextDocument2) GetMainStory(ppStory **ITextStory) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetMainStory, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(ppStory)))
	return HRESULT(ret)
}

func (this *ITextDocument2) GetNewStory(ppStory **ITextStory) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetNewStory, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(ppStory)))
	return HRESULT(ret)
}

func (this *ITextDocument2) GetStory(Index int32, ppStory **ITextStory) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetStory, uintptr(unsafe.Pointer(this)), uintptr(Index), uintptr(unsafe.Pointer(ppStory)))
	return HRESULT(ret)
}

// C241F5E2-7206-11D8-A2C7-00A0D1D6C6B3
var IID_ITextRange2 = syscall.GUID{0xC241F5E2, 0x7206, 0x11D8,
	[8]byte{0xA2, 0xC7, 0x00, 0xA0, 0xD1, 0xD6, 0xC6, 0xB3}}

type ITextRange2Interface interface {
	ITextSelectionInterface
	GetCch(pcch *int32) HRESULT
	GetCells(ppCells **IUnknown) HRESULT
	GetColumn(ppColumn **IUnknown) HRESULT
	GetCount(pCount *int32) HRESULT
	GetDuplicate2(ppRange **ITextRange2) HRESULT
	GetFont2(ppFont **ITextFont2) HRESULT
	SetFont2(pFont *ITextFont2) HRESULT
	GetFormattedText2(ppRange **ITextRange2) HRESULT
	SetFormattedText2(pRange *ITextRange2) HRESULT
	GetGravity(pValue *int32) HRESULT
	SetGravity(Value int32) HRESULT
	GetPara2(ppPara **ITextPara2) HRESULT
	SetPara2(pPara *ITextPara2) HRESULT
	GetRow(ppRow **ITextRow) HRESULT
	GetStartPara(pValue *int32) HRESULT
	GetTable(ppTable **IUnknown) HRESULT
	GetURL(pbstr *BSTR) HRESULT
	SetURL(bstr BSTR) HRESULT
	AddSubrange(cp1 int32, cp2 int32, Activate int32) HRESULT
	BuildUpMath(Flags int32) HRESULT
	DeleteSubrange(cpFirst int32, cpLim int32) HRESULT
	Find(pRange *ITextRange2, Count int32, Flags int32, pDelta *int32) HRESULT
	GetChar2(pChar *int32, Offset int32) HRESULT
	GetDropCap(pcLine *int32, pPosition *int32) HRESULT
	GetInlineObject(pType *int32, pAlign *int32, pChar *int32, pChar1 *int32, pChar2 *int32, pCount *int32, pTeXStyle *int32, pcCol *int32, pLevel *int32) HRESULT
	GetProperty(Type int32, pValue *int32) HRESULT
	GetRect(Type int32, pLeft *int32, pTop *int32, pRight *int32, pBottom *int32, pHit *int32) HRESULT
	GetSubrange(iSubrange int32, pcpFirst *int32, pcpLim *int32) HRESULT
	GetText2(Flags int32, pbstr *BSTR) HRESULT
	HexToUnicode() HRESULT
	InsertTable(cCol int32, cRow int32, AutoFit int32) HRESULT
	Linearize(Flags int32) HRESULT
	SetActiveSubrange(cpAnchor int32, cpActive int32) HRESULT
	SetDropCap(cLine int32, Position int32) HRESULT
	SetProperty(Type int32, Value int32) HRESULT
	SetText2(Flags int32, bstr BSTR) HRESULT
	UnicodeToHex() HRESULT
	SetInlineObject(Type int32, Align int32, Char int32, Char1 int32, Char2 int32, Count int32, TeXStyle int32, cCol int32) HRESULT
	GetMathFunctionType(bstr BSTR, pValue *int32) HRESULT
	InsertImage(width int32, height int32, ascent int32, Type TEXT_ALIGN_OPTIONS, bstrAltText BSTR, pStream *IStream) HRESULT
}

type ITextRange2Vtbl struct {
	ITextSelectionVtbl
	GetCch              uintptr
	GetCells            uintptr
	GetColumn           uintptr
	GetCount            uintptr
	GetDuplicate2       uintptr
	GetFont2            uintptr
	SetFont2            uintptr
	GetFormattedText2   uintptr
	SetFormattedText2   uintptr
	GetGravity          uintptr
	SetGravity          uintptr
	GetPara2            uintptr
	SetPara2            uintptr
	GetRow              uintptr
	GetStartPara        uintptr
	GetTable            uintptr
	GetURL              uintptr
	SetURL              uintptr
	AddSubrange         uintptr
	BuildUpMath         uintptr
	DeleteSubrange      uintptr
	Find                uintptr
	GetChar2            uintptr
	GetDropCap          uintptr
	GetInlineObject     uintptr
	GetProperty         uintptr
	GetRect             uintptr
	GetSubrange         uintptr
	GetText2            uintptr
	HexToUnicode        uintptr
	InsertTable         uintptr
	Linearize           uintptr
	SetActiveSubrange   uintptr
	SetDropCap          uintptr
	SetProperty         uintptr
	SetText2            uintptr
	UnicodeToHex        uintptr
	SetInlineObject     uintptr
	GetMathFunctionType uintptr
	InsertImage         uintptr
}

type ITextRange2 struct {
	ITextSelection
}

func (this *ITextRange2) Vtbl() *ITextRange2Vtbl {
	return (*ITextRange2Vtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *ITextRange2) GetCch(pcch *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetCch, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pcch)))
	return HRESULT(ret)
}

func (this *ITextRange2) GetCells(ppCells **IUnknown) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetCells, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(ppCells)))
	return HRESULT(ret)
}

func (this *ITextRange2) GetColumn(ppColumn **IUnknown) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetColumn, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(ppColumn)))
	return HRESULT(ret)
}

func (this *ITextRange2) GetCount(pCount *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetCount, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pCount)))
	return HRESULT(ret)
}

func (this *ITextRange2) GetDuplicate2(ppRange **ITextRange2) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetDuplicate2, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(ppRange)))
	return HRESULT(ret)
}

func (this *ITextRange2) GetFont2(ppFont **ITextFont2) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetFont2, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(ppFont)))
	return HRESULT(ret)
}

func (this *ITextRange2) SetFont2(pFont *ITextFont2) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetFont2, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pFont)))
	return HRESULT(ret)
}

func (this *ITextRange2) GetFormattedText2(ppRange **ITextRange2) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetFormattedText2, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(ppRange)))
	return HRESULT(ret)
}

func (this *ITextRange2) SetFormattedText2(pRange *ITextRange2) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetFormattedText2, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pRange)))
	return HRESULT(ret)
}

func (this *ITextRange2) GetGravity(pValue *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetGravity, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pValue)))
	return HRESULT(ret)
}

func (this *ITextRange2) SetGravity(Value int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetGravity, uintptr(unsafe.Pointer(this)), uintptr(Value))
	return HRESULT(ret)
}

func (this *ITextRange2) GetPara2(ppPara **ITextPara2) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetPara2, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(ppPara)))
	return HRESULT(ret)
}

func (this *ITextRange2) SetPara2(pPara *ITextPara2) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetPara2, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pPara)))
	return HRESULT(ret)
}

func (this *ITextRange2) GetRow(ppRow **ITextRow) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetRow, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(ppRow)))
	return HRESULT(ret)
}

func (this *ITextRange2) GetStartPara(pValue *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetStartPara, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pValue)))
	return HRESULT(ret)
}

func (this *ITextRange2) GetTable(ppTable **IUnknown) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetTable, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(ppTable)))
	return HRESULT(ret)
}

func (this *ITextRange2) GetURL(pbstr *BSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetURL, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pbstr)))
	return HRESULT(ret)
}

func (this *ITextRange2) SetURL(bstr BSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetURL, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(bstr)))
	return HRESULT(ret)
}

func (this *ITextRange2) AddSubrange(cp1 int32, cp2 int32, Activate int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().AddSubrange, uintptr(unsafe.Pointer(this)), uintptr(cp1), uintptr(cp2), uintptr(Activate))
	return HRESULT(ret)
}

func (this *ITextRange2) BuildUpMath(Flags int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().BuildUpMath, uintptr(unsafe.Pointer(this)), uintptr(Flags))
	return HRESULT(ret)
}

func (this *ITextRange2) DeleteSubrange(cpFirst int32, cpLim int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().DeleteSubrange, uintptr(unsafe.Pointer(this)), uintptr(cpFirst), uintptr(cpLim))
	return HRESULT(ret)
}

func (this *ITextRange2) Find(pRange *ITextRange2, Count int32, Flags int32, pDelta *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Find, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pRange)), uintptr(Count), uintptr(Flags), uintptr(unsafe.Pointer(pDelta)))
	return HRESULT(ret)
}

func (this *ITextRange2) GetChar2(pChar *int32, Offset int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetChar2, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pChar)), uintptr(Offset))
	return HRESULT(ret)
}

func (this *ITextRange2) GetDropCap(pcLine *int32, pPosition *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetDropCap, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pcLine)), uintptr(unsafe.Pointer(pPosition)))
	return HRESULT(ret)
}

func (this *ITextRange2) GetInlineObject(pType *int32, pAlign *int32, pChar *int32, pChar1 *int32, pChar2 *int32, pCount *int32, pTeXStyle *int32, pcCol *int32, pLevel *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetInlineObject, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pType)), uintptr(unsafe.Pointer(pAlign)), uintptr(unsafe.Pointer(pChar)), uintptr(unsafe.Pointer(pChar1)), uintptr(unsafe.Pointer(pChar2)), uintptr(unsafe.Pointer(pCount)), uintptr(unsafe.Pointer(pTeXStyle)), uintptr(unsafe.Pointer(pcCol)), uintptr(unsafe.Pointer(pLevel)))
	return HRESULT(ret)
}

func (this *ITextRange2) GetProperty(Type int32, pValue *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetProperty, uintptr(unsafe.Pointer(this)), uintptr(Type), uintptr(unsafe.Pointer(pValue)))
	return HRESULT(ret)
}

func (this *ITextRange2) GetRect(Type int32, pLeft *int32, pTop *int32, pRight *int32, pBottom *int32, pHit *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetRect, uintptr(unsafe.Pointer(this)), uintptr(Type), uintptr(unsafe.Pointer(pLeft)), uintptr(unsafe.Pointer(pTop)), uintptr(unsafe.Pointer(pRight)), uintptr(unsafe.Pointer(pBottom)), uintptr(unsafe.Pointer(pHit)))
	return HRESULT(ret)
}

func (this *ITextRange2) GetSubrange(iSubrange int32, pcpFirst *int32, pcpLim *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetSubrange, uintptr(unsafe.Pointer(this)), uintptr(iSubrange), uintptr(unsafe.Pointer(pcpFirst)), uintptr(unsafe.Pointer(pcpLim)))
	return HRESULT(ret)
}

func (this *ITextRange2) GetText2(Flags int32, pbstr *BSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetText2, uintptr(unsafe.Pointer(this)), uintptr(Flags), uintptr(unsafe.Pointer(pbstr)))
	return HRESULT(ret)
}

func (this *ITextRange2) HexToUnicode() HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().HexToUnicode, uintptr(unsafe.Pointer(this)))
	return HRESULT(ret)
}

func (this *ITextRange2) InsertTable(cCol int32, cRow int32, AutoFit int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().InsertTable, uintptr(unsafe.Pointer(this)), uintptr(cCol), uintptr(cRow), uintptr(AutoFit))
	return HRESULT(ret)
}

func (this *ITextRange2) Linearize(Flags int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Linearize, uintptr(unsafe.Pointer(this)), uintptr(Flags))
	return HRESULT(ret)
}

func (this *ITextRange2) SetActiveSubrange(cpAnchor int32, cpActive int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetActiveSubrange, uintptr(unsafe.Pointer(this)), uintptr(cpAnchor), uintptr(cpActive))
	return HRESULT(ret)
}

func (this *ITextRange2) SetDropCap(cLine int32, Position int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetDropCap, uintptr(unsafe.Pointer(this)), uintptr(cLine), uintptr(Position))
	return HRESULT(ret)
}

func (this *ITextRange2) SetProperty(Type int32, Value int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetProperty, uintptr(unsafe.Pointer(this)), uintptr(Type), uintptr(Value))
	return HRESULT(ret)
}

func (this *ITextRange2) SetText2(Flags int32, bstr BSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetText2, uintptr(unsafe.Pointer(this)), uintptr(Flags), uintptr(unsafe.Pointer(bstr)))
	return HRESULT(ret)
}

func (this *ITextRange2) UnicodeToHex() HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().UnicodeToHex, uintptr(unsafe.Pointer(this)))
	return HRESULT(ret)
}

func (this *ITextRange2) SetInlineObject(Type int32, Align int32, Char int32, Char1 int32, Char2 int32, Count int32, TeXStyle int32, cCol int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetInlineObject, uintptr(unsafe.Pointer(this)), uintptr(Type), uintptr(Align), uintptr(Char), uintptr(Char1), uintptr(Char2), uintptr(Count), uintptr(TeXStyle), uintptr(cCol))
	return HRESULT(ret)
}

func (this *ITextRange2) GetMathFunctionType(bstr BSTR, pValue *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetMathFunctionType, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(bstr)), uintptr(unsafe.Pointer(pValue)))
	return HRESULT(ret)
}

func (this *ITextRange2) InsertImage(width int32, height int32, ascent int32, Type TEXT_ALIGN_OPTIONS, bstrAltText BSTR, pStream *IStream) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().InsertImage, uintptr(unsafe.Pointer(this)), uintptr(width), uintptr(height), uintptr(ascent), uintptr(Type), uintptr(unsafe.Pointer(bstrAltText)), uintptr(unsafe.Pointer(pStream)))
	return HRESULT(ret)
}

// C241F5E1-7206-11D8-A2C7-00A0D1D6C6B3
var IID_ITextSelection2 = syscall.GUID{0xC241F5E1, 0x7206, 0x11D8,
	[8]byte{0xA2, 0xC7, 0x00, 0xA0, 0xD1, 0xD6, 0xC6, 0xB3}}

type ITextSelection2Interface interface {
	ITextRange2Interface
}

type ITextSelection2Vtbl struct {
	ITextRange2Vtbl
}

type ITextSelection2 struct {
	ITextRange2
}

func (this *ITextSelection2) Vtbl() *ITextSelection2Vtbl {
	return (*ITextSelection2Vtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

// C241F5E3-7206-11D8-A2C7-00A0D1D6C6B3
var IID_ITextFont2 = syscall.GUID{0xC241F5E3, 0x7206, 0x11D8,
	[8]byte{0xA2, 0xC7, 0x00, 0xA0, 0xD1, 0xD6, 0xC6, 0xB3}}

type ITextFont2Interface interface {
	ITextFontInterface
	GetCount(pCount *int32) HRESULT
	GetAutoLigatures(pValue *int32) HRESULT
	SetAutoLigatures(Value int32) HRESULT
	GetAutospaceAlpha(pValue *int32) HRESULT
	SetAutospaceAlpha(Value int32) HRESULT
	GetAutospaceNumeric(pValue *int32) HRESULT
	SetAutospaceNumeric(Value int32) HRESULT
	GetAutospaceParens(pValue *int32) HRESULT
	SetAutospaceParens(Value int32) HRESULT
	GetCharRep(pValue *int32) HRESULT
	SetCharRep(Value int32) HRESULT
	GetCompressionMode(pValue *int32) HRESULT
	SetCompressionMode(Value int32) HRESULT
	GetCookie(pValue *int32) HRESULT
	SetCookie(Value int32) HRESULT
	GetDoubleStrike(pValue *int32) HRESULT
	SetDoubleStrike(Value int32) HRESULT
	GetDuplicate2(ppFont **ITextFont2) HRESULT
	SetDuplicate2(pFont *ITextFont2) HRESULT
	GetLinkType(pValue *int32) HRESULT
	GetMathZone(pValue *int32) HRESULT
	SetMathZone(Value int32) HRESULT
	GetModWidthPairs(pValue *int32) HRESULT
	SetModWidthPairs(Value int32) HRESULT
	GetModWidthSpace(pValue *int32) HRESULT
	SetModWidthSpace(Value int32) HRESULT
	GetOldNumbers(pValue *int32) HRESULT
	SetOldNumbers(Value int32) HRESULT
	GetOverlapping(pValue *int32) HRESULT
	SetOverlapping(Value int32) HRESULT
	GetPositionSubSuper(pValue *int32) HRESULT
	SetPositionSubSuper(Value int32) HRESULT
	GetScaling(pValue *int32) HRESULT
	SetScaling(Value int32) HRESULT
	GetSpaceExtension(pValue *float32) HRESULT
	SetSpaceExtension(Value float32) HRESULT
	GetUnderlinePositionMode(pValue *int32) HRESULT
	SetUnderlinePositionMode(Value int32) HRESULT
	GetEffects(pValue *int32, pMask *int32) HRESULT
	GetEffects2(pValue *int32, pMask *int32) HRESULT
	GetProperty(Type int32, pValue *int32) HRESULT
	GetPropertyInfo(Index int32, pType *int32, pValue *int32) HRESULT
	IsEqual2(pFont *ITextFont2, pB *int32) HRESULT
	SetEffects(Value int32, Mask int32) HRESULT
	SetEffects2(Value int32, Mask int32) HRESULT
	SetProperty(Type int32, Value int32) HRESULT
}

type ITextFont2Vtbl struct {
	ITextFontVtbl
	GetCount                 uintptr
	GetAutoLigatures         uintptr
	SetAutoLigatures         uintptr
	GetAutospaceAlpha        uintptr
	SetAutospaceAlpha        uintptr
	GetAutospaceNumeric      uintptr
	SetAutospaceNumeric      uintptr
	GetAutospaceParens       uintptr
	SetAutospaceParens       uintptr
	GetCharRep               uintptr
	SetCharRep               uintptr
	GetCompressionMode       uintptr
	SetCompressionMode       uintptr
	GetCookie                uintptr
	SetCookie                uintptr
	GetDoubleStrike          uintptr
	SetDoubleStrike          uintptr
	GetDuplicate2            uintptr
	SetDuplicate2            uintptr
	GetLinkType              uintptr
	GetMathZone              uintptr
	SetMathZone              uintptr
	GetModWidthPairs         uintptr
	SetModWidthPairs         uintptr
	GetModWidthSpace         uintptr
	SetModWidthSpace         uintptr
	GetOldNumbers            uintptr
	SetOldNumbers            uintptr
	GetOverlapping           uintptr
	SetOverlapping           uintptr
	GetPositionSubSuper      uintptr
	SetPositionSubSuper      uintptr
	GetScaling               uintptr
	SetScaling               uintptr
	GetSpaceExtension        uintptr
	SetSpaceExtension        uintptr
	GetUnderlinePositionMode uintptr
	SetUnderlinePositionMode uintptr
	GetEffects               uintptr
	GetEffects2              uintptr
	GetProperty              uintptr
	GetPropertyInfo          uintptr
	IsEqual2                 uintptr
	SetEffects               uintptr
	SetEffects2              uintptr
	SetProperty              uintptr
}

type ITextFont2 struct {
	ITextFont
}

func (this *ITextFont2) Vtbl() *ITextFont2Vtbl {
	return (*ITextFont2Vtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *ITextFont2) GetCount(pCount *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetCount, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pCount)))
	return HRESULT(ret)
}

func (this *ITextFont2) GetAutoLigatures(pValue *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetAutoLigatures, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pValue)))
	return HRESULT(ret)
}

func (this *ITextFont2) SetAutoLigatures(Value int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetAutoLigatures, uintptr(unsafe.Pointer(this)), uintptr(Value))
	return HRESULT(ret)
}

func (this *ITextFont2) GetAutospaceAlpha(pValue *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetAutospaceAlpha, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pValue)))
	return HRESULT(ret)
}

func (this *ITextFont2) SetAutospaceAlpha(Value int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetAutospaceAlpha, uintptr(unsafe.Pointer(this)), uintptr(Value))
	return HRESULT(ret)
}

func (this *ITextFont2) GetAutospaceNumeric(pValue *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetAutospaceNumeric, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pValue)))
	return HRESULT(ret)
}

func (this *ITextFont2) SetAutospaceNumeric(Value int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetAutospaceNumeric, uintptr(unsafe.Pointer(this)), uintptr(Value))
	return HRESULT(ret)
}

func (this *ITextFont2) GetAutospaceParens(pValue *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetAutospaceParens, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pValue)))
	return HRESULT(ret)
}

func (this *ITextFont2) SetAutospaceParens(Value int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetAutospaceParens, uintptr(unsafe.Pointer(this)), uintptr(Value))
	return HRESULT(ret)
}

func (this *ITextFont2) GetCharRep(pValue *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetCharRep, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pValue)))
	return HRESULT(ret)
}

func (this *ITextFont2) SetCharRep(Value int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetCharRep, uintptr(unsafe.Pointer(this)), uintptr(Value))
	return HRESULT(ret)
}

func (this *ITextFont2) GetCompressionMode(pValue *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetCompressionMode, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pValue)))
	return HRESULT(ret)
}

func (this *ITextFont2) SetCompressionMode(Value int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetCompressionMode, uintptr(unsafe.Pointer(this)), uintptr(Value))
	return HRESULT(ret)
}

func (this *ITextFont2) GetCookie(pValue *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetCookie, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pValue)))
	return HRESULT(ret)
}

func (this *ITextFont2) SetCookie(Value int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetCookie, uintptr(unsafe.Pointer(this)), uintptr(Value))
	return HRESULT(ret)
}

func (this *ITextFont2) GetDoubleStrike(pValue *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetDoubleStrike, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pValue)))
	return HRESULT(ret)
}

func (this *ITextFont2) SetDoubleStrike(Value int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetDoubleStrike, uintptr(unsafe.Pointer(this)), uintptr(Value))
	return HRESULT(ret)
}

func (this *ITextFont2) GetDuplicate2(ppFont **ITextFont2) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetDuplicate2, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(ppFont)))
	return HRESULT(ret)
}

func (this *ITextFont2) SetDuplicate2(pFont *ITextFont2) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetDuplicate2, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pFont)))
	return HRESULT(ret)
}

func (this *ITextFont2) GetLinkType(pValue *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetLinkType, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pValue)))
	return HRESULT(ret)
}

func (this *ITextFont2) GetMathZone(pValue *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetMathZone, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pValue)))
	return HRESULT(ret)
}

func (this *ITextFont2) SetMathZone(Value int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetMathZone, uintptr(unsafe.Pointer(this)), uintptr(Value))
	return HRESULT(ret)
}

func (this *ITextFont2) GetModWidthPairs(pValue *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetModWidthPairs, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pValue)))
	return HRESULT(ret)
}

func (this *ITextFont2) SetModWidthPairs(Value int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetModWidthPairs, uintptr(unsafe.Pointer(this)), uintptr(Value))
	return HRESULT(ret)
}

func (this *ITextFont2) GetModWidthSpace(pValue *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetModWidthSpace, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pValue)))
	return HRESULT(ret)
}

func (this *ITextFont2) SetModWidthSpace(Value int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetModWidthSpace, uintptr(unsafe.Pointer(this)), uintptr(Value))
	return HRESULT(ret)
}

func (this *ITextFont2) GetOldNumbers(pValue *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetOldNumbers, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pValue)))
	return HRESULT(ret)
}

func (this *ITextFont2) SetOldNumbers(Value int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetOldNumbers, uintptr(unsafe.Pointer(this)), uintptr(Value))
	return HRESULT(ret)
}

func (this *ITextFont2) GetOverlapping(pValue *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetOverlapping, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pValue)))
	return HRESULT(ret)
}

func (this *ITextFont2) SetOverlapping(Value int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetOverlapping, uintptr(unsafe.Pointer(this)), uintptr(Value))
	return HRESULT(ret)
}

func (this *ITextFont2) GetPositionSubSuper(pValue *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetPositionSubSuper, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pValue)))
	return HRESULT(ret)
}

func (this *ITextFont2) SetPositionSubSuper(Value int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetPositionSubSuper, uintptr(unsafe.Pointer(this)), uintptr(Value))
	return HRESULT(ret)
}

func (this *ITextFont2) GetScaling(pValue *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetScaling, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pValue)))
	return HRESULT(ret)
}

func (this *ITextFont2) SetScaling(Value int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetScaling, uintptr(unsafe.Pointer(this)), uintptr(Value))
	return HRESULT(ret)
}

func (this *ITextFont2) GetSpaceExtension(pValue *float32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetSpaceExtension, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pValue)))
	return HRESULT(ret)
}

func (this *ITextFont2) SetSpaceExtension(Value float32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetSpaceExtension, uintptr(unsafe.Pointer(this)), uintptr(Value))
	return HRESULT(ret)
}

func (this *ITextFont2) GetUnderlinePositionMode(pValue *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetUnderlinePositionMode, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pValue)))
	return HRESULT(ret)
}

func (this *ITextFont2) SetUnderlinePositionMode(Value int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetUnderlinePositionMode, uintptr(unsafe.Pointer(this)), uintptr(Value))
	return HRESULT(ret)
}

func (this *ITextFont2) GetEffects(pValue *int32, pMask *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetEffects, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pValue)), uintptr(unsafe.Pointer(pMask)))
	return HRESULT(ret)
}

func (this *ITextFont2) GetEffects2(pValue *int32, pMask *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetEffects2, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pValue)), uintptr(unsafe.Pointer(pMask)))
	return HRESULT(ret)
}

func (this *ITextFont2) GetProperty(Type int32, pValue *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetProperty, uintptr(unsafe.Pointer(this)), uintptr(Type), uintptr(unsafe.Pointer(pValue)))
	return HRESULT(ret)
}

func (this *ITextFont2) GetPropertyInfo(Index int32, pType *int32, pValue *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetPropertyInfo, uintptr(unsafe.Pointer(this)), uintptr(Index), uintptr(unsafe.Pointer(pType)), uintptr(unsafe.Pointer(pValue)))
	return HRESULT(ret)
}

func (this *ITextFont2) IsEqual2(pFont *ITextFont2, pB *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().IsEqual2, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pFont)), uintptr(unsafe.Pointer(pB)))
	return HRESULT(ret)
}

func (this *ITextFont2) SetEffects(Value int32, Mask int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetEffects, uintptr(unsafe.Pointer(this)), uintptr(Value), uintptr(Mask))
	return HRESULT(ret)
}

func (this *ITextFont2) SetEffects2(Value int32, Mask int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetEffects2, uintptr(unsafe.Pointer(this)), uintptr(Value), uintptr(Mask))
	return HRESULT(ret)
}

func (this *ITextFont2) SetProperty(Type int32, Value int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetProperty, uintptr(unsafe.Pointer(this)), uintptr(Type), uintptr(Value))
	return HRESULT(ret)
}

// C241F5E4-7206-11D8-A2C7-00A0D1D6C6B3
var IID_ITextPara2 = syscall.GUID{0xC241F5E4, 0x7206, 0x11D8,
	[8]byte{0xA2, 0xC7, 0x00, 0xA0, 0xD1, 0xD6, 0xC6, 0xB3}}

type ITextPara2Interface interface {
	ITextParaInterface
	GetBorders(ppBorders **IUnknown) HRESULT
	GetDuplicate2(ppPara **ITextPara2) HRESULT
	SetDuplicate2(pPara *ITextPara2) HRESULT
	GetFontAlignment(pValue *int32) HRESULT
	SetFontAlignment(Value int32) HRESULT
	GetHangingPunctuation(pValue *int32) HRESULT
	SetHangingPunctuation(Value int32) HRESULT
	GetSnapToGrid(pValue *int32) HRESULT
	SetSnapToGrid(Value int32) HRESULT
	GetTrimPunctuationAtStart(pValue *int32) HRESULT
	SetTrimPunctuationAtStart(Value int32) HRESULT
	GetEffects(pValue *int32, pMask *int32) HRESULT
	GetProperty(Type int32, pValue *int32) HRESULT
	IsEqual2(pPara *ITextPara2, pB *int32) HRESULT
	SetEffects(Value int32, Mask int32) HRESULT
	SetProperty(Type int32, Value int32) HRESULT
}

type ITextPara2Vtbl struct {
	ITextParaVtbl
	GetBorders                uintptr
	GetDuplicate2             uintptr
	SetDuplicate2             uintptr
	GetFontAlignment          uintptr
	SetFontAlignment          uintptr
	GetHangingPunctuation     uintptr
	SetHangingPunctuation     uintptr
	GetSnapToGrid             uintptr
	SetSnapToGrid             uintptr
	GetTrimPunctuationAtStart uintptr
	SetTrimPunctuationAtStart uintptr
	GetEffects                uintptr
	GetProperty               uintptr
	IsEqual2                  uintptr
	SetEffects                uintptr
	SetProperty               uintptr
}

type ITextPara2 struct {
	ITextPara
}

func (this *ITextPara2) Vtbl() *ITextPara2Vtbl {
	return (*ITextPara2Vtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *ITextPara2) GetBorders(ppBorders **IUnknown) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetBorders, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(ppBorders)))
	return HRESULT(ret)
}

func (this *ITextPara2) GetDuplicate2(ppPara **ITextPara2) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetDuplicate2, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(ppPara)))
	return HRESULT(ret)
}

func (this *ITextPara2) SetDuplicate2(pPara *ITextPara2) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetDuplicate2, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pPara)))
	return HRESULT(ret)
}

func (this *ITextPara2) GetFontAlignment(pValue *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetFontAlignment, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pValue)))
	return HRESULT(ret)
}

func (this *ITextPara2) SetFontAlignment(Value int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetFontAlignment, uintptr(unsafe.Pointer(this)), uintptr(Value))
	return HRESULT(ret)
}

func (this *ITextPara2) GetHangingPunctuation(pValue *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetHangingPunctuation, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pValue)))
	return HRESULT(ret)
}

func (this *ITextPara2) SetHangingPunctuation(Value int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetHangingPunctuation, uintptr(unsafe.Pointer(this)), uintptr(Value))
	return HRESULT(ret)
}

func (this *ITextPara2) GetSnapToGrid(pValue *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetSnapToGrid, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pValue)))
	return HRESULT(ret)
}

func (this *ITextPara2) SetSnapToGrid(Value int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetSnapToGrid, uintptr(unsafe.Pointer(this)), uintptr(Value))
	return HRESULT(ret)
}

func (this *ITextPara2) GetTrimPunctuationAtStart(pValue *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetTrimPunctuationAtStart, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pValue)))
	return HRESULT(ret)
}

func (this *ITextPara2) SetTrimPunctuationAtStart(Value int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetTrimPunctuationAtStart, uintptr(unsafe.Pointer(this)), uintptr(Value))
	return HRESULT(ret)
}

func (this *ITextPara2) GetEffects(pValue *int32, pMask *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetEffects, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pValue)), uintptr(unsafe.Pointer(pMask)))
	return HRESULT(ret)
}

func (this *ITextPara2) GetProperty(Type int32, pValue *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetProperty, uintptr(unsafe.Pointer(this)), uintptr(Type), uintptr(unsafe.Pointer(pValue)))
	return HRESULT(ret)
}

func (this *ITextPara2) IsEqual2(pPara *ITextPara2, pB *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().IsEqual2, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pPara)), uintptr(unsafe.Pointer(pB)))
	return HRESULT(ret)
}

func (this *ITextPara2) SetEffects(Value int32, Mask int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetEffects, uintptr(unsafe.Pointer(this)), uintptr(Value), uintptr(Mask))
	return HRESULT(ret)
}

func (this *ITextPara2) SetProperty(Type int32, Value int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetProperty, uintptr(unsafe.Pointer(this)), uintptr(Type), uintptr(Value))
	return HRESULT(ret)
}

// C241F5E5-7206-11D8-A2C7-00A0D1D6C6B3
var IID_ITextStoryRanges2 = syscall.GUID{0xC241F5E5, 0x7206, 0x11D8,
	[8]byte{0xA2, 0xC7, 0x00, 0xA0, 0xD1, 0xD6, 0xC6, 0xB3}}

type ITextStoryRanges2Interface interface {
	ITextStoryRangesInterface
	Item2(Index int32, ppRange **ITextRange2) HRESULT
}

type ITextStoryRanges2Vtbl struct {
	ITextStoryRangesVtbl
	Item2 uintptr
}

type ITextStoryRanges2 struct {
	ITextStoryRanges
}

func (this *ITextStoryRanges2) Vtbl() *ITextStoryRanges2Vtbl {
	return (*ITextStoryRanges2Vtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *ITextStoryRanges2) Item2(Index int32, ppRange **ITextRange2) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Item2, uintptr(unsafe.Pointer(this)), uintptr(Index), uintptr(unsafe.Pointer(ppRange)))
	return HRESULT(ret)
}

// C241F5F3-7206-11D8-A2C7-00A0D1D6C6B3
var IID_ITextStory = syscall.GUID{0xC241F5F3, 0x7206, 0x11D8,
	[8]byte{0xA2, 0xC7, 0x00, 0xA0, 0xD1, 0xD6, 0xC6, 0xB3}}

type ITextStoryInterface interface {
	IUnknownInterface
	GetActive(pValue *int32) HRESULT
	SetActive(Value int32) HRESULT
	GetDisplay(ppDisplay **IUnknown) HRESULT
	GetIndex(pValue *int32) HRESULT
	GetType(pValue *int32) HRESULT
	SetType(Value int32) HRESULT
	GetProperty(Type int32, pValue *int32) HRESULT
	GetRange(cpActive int32, cpAnchor int32, ppRange **ITextRange2) HRESULT
	GetText(Flags int32, pbstr *BSTR) HRESULT
	SetFormattedText(pUnk *IUnknown) HRESULT
	SetProperty(Type int32, Value int32) HRESULT
	SetText(Flags int32, bstr BSTR) HRESULT
}

type ITextStoryVtbl struct {
	IUnknownVtbl
	GetActive        uintptr
	SetActive        uintptr
	GetDisplay       uintptr
	GetIndex         uintptr
	GetType          uintptr
	SetType          uintptr
	GetProperty      uintptr
	GetRange         uintptr
	GetText          uintptr
	SetFormattedText uintptr
	SetProperty      uintptr
	SetText          uintptr
}

type ITextStory struct {
	IUnknown
}

func (this *ITextStory) Vtbl() *ITextStoryVtbl {
	return (*ITextStoryVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *ITextStory) GetActive(pValue *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetActive, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pValue)))
	return HRESULT(ret)
}

func (this *ITextStory) SetActive(Value int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetActive, uintptr(unsafe.Pointer(this)), uintptr(Value))
	return HRESULT(ret)
}

func (this *ITextStory) GetDisplay(ppDisplay **IUnknown) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetDisplay, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(ppDisplay)))
	return HRESULT(ret)
}

func (this *ITextStory) GetIndex(pValue *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetIndex, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pValue)))
	return HRESULT(ret)
}

func (this *ITextStory) GetType(pValue *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetType, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pValue)))
	return HRESULT(ret)
}

func (this *ITextStory) SetType(Value int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetType, uintptr(unsafe.Pointer(this)), uintptr(Value))
	return HRESULT(ret)
}

func (this *ITextStory) GetProperty(Type int32, pValue *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetProperty, uintptr(unsafe.Pointer(this)), uintptr(Type), uintptr(unsafe.Pointer(pValue)))
	return HRESULT(ret)
}

func (this *ITextStory) GetRange(cpActive int32, cpAnchor int32, ppRange **ITextRange2) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetRange, uintptr(unsafe.Pointer(this)), uintptr(cpActive), uintptr(cpAnchor), uintptr(unsafe.Pointer(ppRange)))
	return HRESULT(ret)
}

func (this *ITextStory) GetText(Flags int32, pbstr *BSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetText, uintptr(unsafe.Pointer(this)), uintptr(Flags), uintptr(unsafe.Pointer(pbstr)))
	return HRESULT(ret)
}

func (this *ITextStory) SetFormattedText(pUnk *IUnknown) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetFormattedText, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pUnk)))
	return HRESULT(ret)
}

func (this *ITextStory) SetProperty(Type int32, Value int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetProperty, uintptr(unsafe.Pointer(this)), uintptr(Type), uintptr(Value))
	return HRESULT(ret)
}

func (this *ITextStory) SetText(Flags int32, bstr BSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetText, uintptr(unsafe.Pointer(this)), uintptr(Flags), uintptr(unsafe.Pointer(bstr)))
	return HRESULT(ret)
}

// C241F5E7-7206-11D8-A2C7-00A0D1D6C6B3
var IID_ITextStrings = syscall.GUID{0xC241F5E7, 0x7206, 0x11D8,
	[8]byte{0xA2, 0xC7, 0x00, 0xA0, 0xD1, 0xD6, 0xC6, 0xB3}}

type ITextStringsInterface interface {
	IDispatchInterface
	Item(Index int32, ppRange **ITextRange2) HRESULT
	GetCount(pCount *int32) HRESULT
	Add(bstr BSTR) HRESULT
	Append(pRange *ITextRange2, iString int32) HRESULT
	Cat2(iString int32) HRESULT
	CatTop2(bstr BSTR) HRESULT
	DeleteRange(pRange *ITextRange2) HRESULT
	EncodeFunction(Type int32, Align int32, Char int32, Char1 int32, Char2 int32, Count int32, TeXStyle int32, cCol int32, pRange *ITextRange2) HRESULT
	GetCch(iString int32, pcch *int32) HRESULT
	InsertNullStr(iString int32) HRESULT
	MoveBoundary(iString int32, cch int32) HRESULT
	PrefixTop(bstr BSTR) HRESULT
	Remove(iString int32, cString int32) HRESULT
	SetFormattedText(pRangeD *ITextRange2, pRangeS *ITextRange2) HRESULT
	SetOpCp(iString int32, cp int32) HRESULT
	SuffixTop(bstr BSTR, pRange *ITextRange2) HRESULT
	Swap() HRESULT
}

type ITextStringsVtbl struct {
	IDispatchVtbl
	Item             uintptr
	GetCount         uintptr
	Add              uintptr
	Append           uintptr
	Cat2             uintptr
	CatTop2          uintptr
	DeleteRange      uintptr
	EncodeFunction   uintptr
	GetCch           uintptr
	InsertNullStr    uintptr
	MoveBoundary     uintptr
	PrefixTop        uintptr
	Remove           uintptr
	SetFormattedText uintptr
	SetOpCp          uintptr
	SuffixTop        uintptr
	Swap             uintptr
}

type ITextStrings struct {
	IDispatch
}

func (this *ITextStrings) Vtbl() *ITextStringsVtbl {
	return (*ITextStringsVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *ITextStrings) Item(Index int32, ppRange **ITextRange2) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Item, uintptr(unsafe.Pointer(this)), uintptr(Index), uintptr(unsafe.Pointer(ppRange)))
	return HRESULT(ret)
}

func (this *ITextStrings) GetCount(pCount *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetCount, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pCount)))
	return HRESULT(ret)
}

func (this *ITextStrings) Add(bstr BSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Add, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(bstr)))
	return HRESULT(ret)
}

func (this *ITextStrings) Append(pRange *ITextRange2, iString int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Append, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pRange)), uintptr(iString))
	return HRESULT(ret)
}

func (this *ITextStrings) Cat2(iString int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Cat2, uintptr(unsafe.Pointer(this)), uintptr(iString))
	return HRESULT(ret)
}

func (this *ITextStrings) CatTop2(bstr BSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().CatTop2, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(bstr)))
	return HRESULT(ret)
}

func (this *ITextStrings) DeleteRange(pRange *ITextRange2) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().DeleteRange, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pRange)))
	return HRESULT(ret)
}

func (this *ITextStrings) EncodeFunction(Type int32, Align int32, Char int32, Char1 int32, Char2 int32, Count int32, TeXStyle int32, cCol int32, pRange *ITextRange2) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().EncodeFunction, uintptr(unsafe.Pointer(this)), uintptr(Type), uintptr(Align), uintptr(Char), uintptr(Char1), uintptr(Char2), uintptr(Count), uintptr(TeXStyle), uintptr(cCol), uintptr(unsafe.Pointer(pRange)))
	return HRESULT(ret)
}

func (this *ITextStrings) GetCch(iString int32, pcch *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetCch, uintptr(unsafe.Pointer(this)), uintptr(iString), uintptr(unsafe.Pointer(pcch)))
	return HRESULT(ret)
}

func (this *ITextStrings) InsertNullStr(iString int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().InsertNullStr, uintptr(unsafe.Pointer(this)), uintptr(iString))
	return HRESULT(ret)
}

func (this *ITextStrings) MoveBoundary(iString int32, cch int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().MoveBoundary, uintptr(unsafe.Pointer(this)), uintptr(iString), uintptr(cch))
	return HRESULT(ret)
}

func (this *ITextStrings) PrefixTop(bstr BSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().PrefixTop, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(bstr)))
	return HRESULT(ret)
}

func (this *ITextStrings) Remove(iString int32, cString int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Remove, uintptr(unsafe.Pointer(this)), uintptr(iString), uintptr(cString))
	return HRESULT(ret)
}

func (this *ITextStrings) SetFormattedText(pRangeD *ITextRange2, pRangeS *ITextRange2) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetFormattedText, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pRangeD)), uintptr(unsafe.Pointer(pRangeS)))
	return HRESULT(ret)
}

func (this *ITextStrings) SetOpCp(iString int32, cp int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetOpCp, uintptr(unsafe.Pointer(this)), uintptr(iString), uintptr(cp))
	return HRESULT(ret)
}

func (this *ITextStrings) SuffixTop(bstr BSTR, pRange *ITextRange2) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SuffixTop, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(bstr)), uintptr(unsafe.Pointer(pRange)))
	return HRESULT(ret)
}

func (this *ITextStrings) Swap() HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Swap, uintptr(unsafe.Pointer(this)))
	return HRESULT(ret)
}

// C241F5EF-7206-11D8-A2C7-00A0D1D6C6B3
var IID_ITextRow = syscall.GUID{0xC241F5EF, 0x7206, 0x11D8,
	[8]byte{0xA2, 0xC7, 0x00, 0xA0, 0xD1, 0xD6, 0xC6, 0xB3}}

type ITextRowInterface interface {
	IDispatchInterface
	GetAlignment(pValue *int32) HRESULT
	SetAlignment(Value int32) HRESULT
	GetCellCount(pValue *int32) HRESULT
	SetCellCount(Value int32) HRESULT
	GetCellCountCache(pValue *int32) HRESULT
	SetCellCountCache(Value int32) HRESULT
	GetCellIndex(pValue *int32) HRESULT
	SetCellIndex(Value int32) HRESULT
	GetCellMargin(pValue *int32) HRESULT
	SetCellMargin(Value int32) HRESULT
	GetHeight(pValue *int32) HRESULT
	SetHeight(Value int32) HRESULT
	GetIndent(pValue *int32) HRESULT
	SetIndent(Value int32) HRESULT
	GetKeepTogether(pValue *int32) HRESULT
	SetKeepTogether(Value int32) HRESULT
	GetKeepWithNext(pValue *int32) HRESULT
	SetKeepWithNext(Value int32) HRESULT
	GetNestLevel(pValue *int32) HRESULT
	GetRTL(pValue *int32) HRESULT
	SetRTL(Value int32) HRESULT
	GetCellAlignment(pValue *int32) HRESULT
	SetCellAlignment(Value int32) HRESULT
	GetCellColorBack(pValue *int32) HRESULT
	SetCellColorBack(Value int32) HRESULT
	GetCellColorFore(pValue *int32) HRESULT
	SetCellColorFore(Value int32) HRESULT
	GetCellMergeFlags(pValue *int32) HRESULT
	SetCellMergeFlags(Value int32) HRESULT
	GetCellShading(pValue *int32) HRESULT
	SetCellShading(Value int32) HRESULT
	GetCellVerticalText(pValue *int32) HRESULT
	SetCellVerticalText(Value int32) HRESULT
	GetCellWidth(pValue *int32) HRESULT
	SetCellWidth(Value int32) HRESULT
	GetCellBorderColors(pcrLeft *int32, pcrTop *int32, pcrRight *int32, pcrBottom *int32) HRESULT
	GetCellBorderWidths(pduLeft *int32, pduTop *int32, pduRight *int32, pduBottom *int32) HRESULT
	SetCellBorderColors(crLeft int32, crTop int32, crRight int32, crBottom int32) HRESULT
	SetCellBorderWidths(duLeft int32, duTop int32, duRight int32, duBottom int32) HRESULT
	Apply(cRow int32, Flags TomConstants) HRESULT
	CanChange(pValue *int32) HRESULT
	GetProperty(Type int32, pValue *int32) HRESULT
	Insert(cRow int32) HRESULT
	IsEqual(pRow *ITextRow, pB *int32) HRESULT
	Reset(Value int32) HRESULT
	SetProperty(Type int32, Value int32) HRESULT
}

type ITextRowVtbl struct {
	IDispatchVtbl
	GetAlignment        uintptr
	SetAlignment        uintptr
	GetCellCount        uintptr
	SetCellCount        uintptr
	GetCellCountCache   uintptr
	SetCellCountCache   uintptr
	GetCellIndex        uintptr
	SetCellIndex        uintptr
	GetCellMargin       uintptr
	SetCellMargin       uintptr
	GetHeight           uintptr
	SetHeight           uintptr
	GetIndent           uintptr
	SetIndent           uintptr
	GetKeepTogether     uintptr
	SetKeepTogether     uintptr
	GetKeepWithNext     uintptr
	SetKeepWithNext     uintptr
	GetNestLevel        uintptr
	GetRTL              uintptr
	SetRTL              uintptr
	GetCellAlignment    uintptr
	SetCellAlignment    uintptr
	GetCellColorBack    uintptr
	SetCellColorBack    uintptr
	GetCellColorFore    uintptr
	SetCellColorFore    uintptr
	GetCellMergeFlags   uintptr
	SetCellMergeFlags   uintptr
	GetCellShading      uintptr
	SetCellShading      uintptr
	GetCellVerticalText uintptr
	SetCellVerticalText uintptr
	GetCellWidth        uintptr
	SetCellWidth        uintptr
	GetCellBorderColors uintptr
	GetCellBorderWidths uintptr
	SetCellBorderColors uintptr
	SetCellBorderWidths uintptr
	Apply               uintptr
	CanChange           uintptr
	GetProperty         uintptr
	Insert              uintptr
	IsEqual             uintptr
	Reset               uintptr
	SetProperty         uintptr
}

type ITextRow struct {
	IDispatch
}

func (this *ITextRow) Vtbl() *ITextRowVtbl {
	return (*ITextRowVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *ITextRow) GetAlignment(pValue *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetAlignment, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pValue)))
	return HRESULT(ret)
}

func (this *ITextRow) SetAlignment(Value int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetAlignment, uintptr(unsafe.Pointer(this)), uintptr(Value))
	return HRESULT(ret)
}

func (this *ITextRow) GetCellCount(pValue *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetCellCount, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pValue)))
	return HRESULT(ret)
}

func (this *ITextRow) SetCellCount(Value int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetCellCount, uintptr(unsafe.Pointer(this)), uintptr(Value))
	return HRESULT(ret)
}

func (this *ITextRow) GetCellCountCache(pValue *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetCellCountCache, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pValue)))
	return HRESULT(ret)
}

func (this *ITextRow) SetCellCountCache(Value int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetCellCountCache, uintptr(unsafe.Pointer(this)), uintptr(Value))
	return HRESULT(ret)
}

func (this *ITextRow) GetCellIndex(pValue *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetCellIndex, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pValue)))
	return HRESULT(ret)
}

func (this *ITextRow) SetCellIndex(Value int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetCellIndex, uintptr(unsafe.Pointer(this)), uintptr(Value))
	return HRESULT(ret)
}

func (this *ITextRow) GetCellMargin(pValue *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetCellMargin, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pValue)))
	return HRESULT(ret)
}

func (this *ITextRow) SetCellMargin(Value int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetCellMargin, uintptr(unsafe.Pointer(this)), uintptr(Value))
	return HRESULT(ret)
}

func (this *ITextRow) GetHeight(pValue *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetHeight, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pValue)))
	return HRESULT(ret)
}

func (this *ITextRow) SetHeight(Value int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetHeight, uintptr(unsafe.Pointer(this)), uintptr(Value))
	return HRESULT(ret)
}

func (this *ITextRow) GetIndent(pValue *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetIndent, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pValue)))
	return HRESULT(ret)
}

func (this *ITextRow) SetIndent(Value int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetIndent, uintptr(unsafe.Pointer(this)), uintptr(Value))
	return HRESULT(ret)
}

func (this *ITextRow) GetKeepTogether(pValue *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetKeepTogether, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pValue)))
	return HRESULT(ret)
}

func (this *ITextRow) SetKeepTogether(Value int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetKeepTogether, uintptr(unsafe.Pointer(this)), uintptr(Value))
	return HRESULT(ret)
}

func (this *ITextRow) GetKeepWithNext(pValue *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetKeepWithNext, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pValue)))
	return HRESULT(ret)
}

func (this *ITextRow) SetKeepWithNext(Value int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetKeepWithNext, uintptr(unsafe.Pointer(this)), uintptr(Value))
	return HRESULT(ret)
}

func (this *ITextRow) GetNestLevel(pValue *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetNestLevel, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pValue)))
	return HRESULT(ret)
}

func (this *ITextRow) GetRTL(pValue *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetRTL, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pValue)))
	return HRESULT(ret)
}

func (this *ITextRow) SetRTL(Value int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetRTL, uintptr(unsafe.Pointer(this)), uintptr(Value))
	return HRESULT(ret)
}

func (this *ITextRow) GetCellAlignment(pValue *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetCellAlignment, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pValue)))
	return HRESULT(ret)
}

func (this *ITextRow) SetCellAlignment(Value int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetCellAlignment, uintptr(unsafe.Pointer(this)), uintptr(Value))
	return HRESULT(ret)
}

func (this *ITextRow) GetCellColorBack(pValue *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetCellColorBack, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pValue)))
	return HRESULT(ret)
}

func (this *ITextRow) SetCellColorBack(Value int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetCellColorBack, uintptr(unsafe.Pointer(this)), uintptr(Value))
	return HRESULT(ret)
}

func (this *ITextRow) GetCellColorFore(pValue *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetCellColorFore, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pValue)))
	return HRESULT(ret)
}

func (this *ITextRow) SetCellColorFore(Value int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetCellColorFore, uintptr(unsafe.Pointer(this)), uintptr(Value))
	return HRESULT(ret)
}

func (this *ITextRow) GetCellMergeFlags(pValue *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetCellMergeFlags, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pValue)))
	return HRESULT(ret)
}

func (this *ITextRow) SetCellMergeFlags(Value int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetCellMergeFlags, uintptr(unsafe.Pointer(this)), uintptr(Value))
	return HRESULT(ret)
}

func (this *ITextRow) GetCellShading(pValue *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetCellShading, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pValue)))
	return HRESULT(ret)
}

func (this *ITextRow) SetCellShading(Value int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetCellShading, uintptr(unsafe.Pointer(this)), uintptr(Value))
	return HRESULT(ret)
}

func (this *ITextRow) GetCellVerticalText(pValue *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetCellVerticalText, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pValue)))
	return HRESULT(ret)
}

func (this *ITextRow) SetCellVerticalText(Value int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetCellVerticalText, uintptr(unsafe.Pointer(this)), uintptr(Value))
	return HRESULT(ret)
}

func (this *ITextRow) GetCellWidth(pValue *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetCellWidth, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pValue)))
	return HRESULT(ret)
}

func (this *ITextRow) SetCellWidth(Value int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetCellWidth, uintptr(unsafe.Pointer(this)), uintptr(Value))
	return HRESULT(ret)
}

func (this *ITextRow) GetCellBorderColors(pcrLeft *int32, pcrTop *int32, pcrRight *int32, pcrBottom *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetCellBorderColors, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pcrLeft)), uintptr(unsafe.Pointer(pcrTop)), uintptr(unsafe.Pointer(pcrRight)), uintptr(unsafe.Pointer(pcrBottom)))
	return HRESULT(ret)
}

func (this *ITextRow) GetCellBorderWidths(pduLeft *int32, pduTop *int32, pduRight *int32, pduBottom *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetCellBorderWidths, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pduLeft)), uintptr(unsafe.Pointer(pduTop)), uintptr(unsafe.Pointer(pduRight)), uintptr(unsafe.Pointer(pduBottom)))
	return HRESULT(ret)
}

func (this *ITextRow) SetCellBorderColors(crLeft int32, crTop int32, crRight int32, crBottom int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetCellBorderColors, uintptr(unsafe.Pointer(this)), uintptr(crLeft), uintptr(crTop), uintptr(crRight), uintptr(crBottom))
	return HRESULT(ret)
}

func (this *ITextRow) SetCellBorderWidths(duLeft int32, duTop int32, duRight int32, duBottom int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetCellBorderWidths, uintptr(unsafe.Pointer(this)), uintptr(duLeft), uintptr(duTop), uintptr(duRight), uintptr(duBottom))
	return HRESULT(ret)
}

func (this *ITextRow) Apply(cRow int32, Flags TomConstants) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Apply, uintptr(unsafe.Pointer(this)), uintptr(cRow), uintptr(Flags))
	return HRESULT(ret)
}

func (this *ITextRow) CanChange(pValue *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().CanChange, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pValue)))
	return HRESULT(ret)
}

func (this *ITextRow) GetProperty(Type int32, pValue *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetProperty, uintptr(unsafe.Pointer(this)), uintptr(Type), uintptr(unsafe.Pointer(pValue)))
	return HRESULT(ret)
}

func (this *ITextRow) Insert(cRow int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Insert, uintptr(unsafe.Pointer(this)), uintptr(cRow))
	return HRESULT(ret)
}

func (this *ITextRow) IsEqual(pRow *ITextRow, pB *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().IsEqual, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pRow)), uintptr(unsafe.Pointer(pB)))
	return HRESULT(ret)
}

func (this *ITextRow) Reset(Value int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Reset, uintptr(unsafe.Pointer(this)), uintptr(Value))
	return HRESULT(ret)
}

func (this *ITextRow) SetProperty(Type int32, Value int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetProperty, uintptr(unsafe.Pointer(this)), uintptr(Type), uintptr(Value))
	return HRESULT(ret)
}

// C241F5F2-7206-11D8-A2C7-00A0D1D6C6B3
var IID_ITextDisplays = syscall.GUID{0xC241F5F2, 0x7206, 0x11D8,
	[8]byte{0xA2, 0xC7, 0x00, 0xA0, 0xD1, 0xD6, 0xC6, 0xB3}}

type ITextDisplaysInterface interface {
	IDispatchInterface
}

type ITextDisplaysVtbl struct {
	IDispatchVtbl
}

type ITextDisplays struct {
	IDispatch
}

func (this *ITextDisplays) Vtbl() *ITextDisplaysVtbl {
	return (*ITextDisplaysVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

// 01C25500-4268-11D1-883A-3C8B00C10000
var IID_ITextDocument2Old = syscall.GUID{0x01C25500, 0x4268, 0x11D1,
	[8]byte{0x88, 0x3A, 0x3C, 0x8B, 0x00, 0xC1, 0x00, 0x00}}

type ITextDocument2OldInterface interface {
	ITextDocumentInterface
	AttachMsgFilter(pFilter *IUnknown) HRESULT
	SetEffectColor(Index int32, cr uint32) HRESULT
	GetEffectColor(Index int32, pcr *uint32) HRESULT
	GetCaretType(pCaretType *int32) HRESULT
	SetCaretType(CaretType int32) HRESULT
	GetImmContext(pContext *int64) HRESULT
	ReleaseImmContext(Context int64) HRESULT
	GetPreferredFont(cp int32, CharRep int32, Option int32, CharRepCur int32, curFontSize int32, pbstr *BSTR, pPitchAndFamily *int32, pNewFontSize *int32) HRESULT
	GetNotificationMode(pMode *int32) HRESULT
	SetNotificationMode(Mode int32) HRESULT
	GetClientRect(Type int32, pLeft *int32, pTop *int32, pRight *int32, pBottom *int32) HRESULT
	GetSelection2(ppSel **ITextSelection) HRESULT
	GetWindow(phWnd *int32) HRESULT
	GetFEFlags(pFlags *int32) HRESULT
	UpdateWindow() HRESULT
	CheckTextLimit(cch int32, pcch *int32) HRESULT
	IMEInProgress(Value int32) HRESULT
	SysBeep() HRESULT
	Update(Mode int32) HRESULT
	Notify(Notify int32) HRESULT
	GetDocumentFont(ppITextFont **ITextFont) HRESULT
	GetDocumentPara(ppITextPara **ITextPara) HRESULT
	GetCallManager(ppVoid **IUnknown) HRESULT
	ReleaseCallManager(pVoid *IUnknown) HRESULT
}

type ITextDocument2OldVtbl struct {
	ITextDocumentVtbl
	AttachMsgFilter     uintptr
	SetEffectColor      uintptr
	GetEffectColor      uintptr
	GetCaretType        uintptr
	SetCaretType        uintptr
	GetImmContext       uintptr
	ReleaseImmContext   uintptr
	GetPreferredFont    uintptr
	GetNotificationMode uintptr
	SetNotificationMode uintptr
	GetClientRect       uintptr
	GetSelection2       uintptr
	GetWindow           uintptr
	GetFEFlags          uintptr
	UpdateWindow        uintptr
	CheckTextLimit      uintptr
	IMEInProgress       uintptr
	SysBeep             uintptr
	Update              uintptr
	Notify              uintptr
	GetDocumentFont     uintptr
	GetDocumentPara     uintptr
	GetCallManager      uintptr
	ReleaseCallManager  uintptr
}

type ITextDocument2Old struct {
	ITextDocument
}

func (this *ITextDocument2Old) Vtbl() *ITextDocument2OldVtbl {
	return (*ITextDocument2OldVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *ITextDocument2Old) AttachMsgFilter(pFilter *IUnknown) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().AttachMsgFilter, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pFilter)))
	return HRESULT(ret)
}

func (this *ITextDocument2Old) SetEffectColor(Index int32, cr uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetEffectColor, uintptr(unsafe.Pointer(this)), uintptr(Index), uintptr(cr))
	return HRESULT(ret)
}

func (this *ITextDocument2Old) GetEffectColor(Index int32, pcr *uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetEffectColor, uintptr(unsafe.Pointer(this)), uintptr(Index), uintptr(unsafe.Pointer(pcr)))
	return HRESULT(ret)
}

func (this *ITextDocument2Old) GetCaretType(pCaretType *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetCaretType, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pCaretType)))
	return HRESULT(ret)
}

func (this *ITextDocument2Old) SetCaretType(CaretType int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetCaretType, uintptr(unsafe.Pointer(this)), uintptr(CaretType))
	return HRESULT(ret)
}

func (this *ITextDocument2Old) GetImmContext(pContext *int64) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetImmContext, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pContext)))
	return HRESULT(ret)
}

func (this *ITextDocument2Old) ReleaseImmContext(Context int64) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().ReleaseImmContext, uintptr(unsafe.Pointer(this)), uintptr(Context))
	return HRESULT(ret)
}

func (this *ITextDocument2Old) GetPreferredFont(cp int32, CharRep int32, Option int32, CharRepCur int32, curFontSize int32, pbstr *BSTR, pPitchAndFamily *int32, pNewFontSize *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetPreferredFont, uintptr(unsafe.Pointer(this)), uintptr(cp), uintptr(CharRep), uintptr(Option), uintptr(CharRepCur), uintptr(curFontSize), uintptr(unsafe.Pointer(pbstr)), uintptr(unsafe.Pointer(pPitchAndFamily)), uintptr(unsafe.Pointer(pNewFontSize)))
	return HRESULT(ret)
}

func (this *ITextDocument2Old) GetNotificationMode(pMode *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetNotificationMode, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pMode)))
	return HRESULT(ret)
}

func (this *ITextDocument2Old) SetNotificationMode(Mode int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetNotificationMode, uintptr(unsafe.Pointer(this)), uintptr(Mode))
	return HRESULT(ret)
}

func (this *ITextDocument2Old) GetClientRect(Type int32, pLeft *int32, pTop *int32, pRight *int32, pBottom *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetClientRect, uintptr(unsafe.Pointer(this)), uintptr(Type), uintptr(unsafe.Pointer(pLeft)), uintptr(unsafe.Pointer(pTop)), uintptr(unsafe.Pointer(pRight)), uintptr(unsafe.Pointer(pBottom)))
	return HRESULT(ret)
}

func (this *ITextDocument2Old) GetSelection2(ppSel **ITextSelection) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetSelection2, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(ppSel)))
	return HRESULT(ret)
}

func (this *ITextDocument2Old) GetWindow(phWnd *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetWindow, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(phWnd)))
	return HRESULT(ret)
}

func (this *ITextDocument2Old) GetFEFlags(pFlags *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetFEFlags, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pFlags)))
	return HRESULT(ret)
}

func (this *ITextDocument2Old) UpdateWindow() HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().UpdateWindow, uintptr(unsafe.Pointer(this)))
	return HRESULT(ret)
}

func (this *ITextDocument2Old) CheckTextLimit(cch int32, pcch *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().CheckTextLimit, uintptr(unsafe.Pointer(this)), uintptr(cch), uintptr(unsafe.Pointer(pcch)))
	return HRESULT(ret)
}

func (this *ITextDocument2Old) IMEInProgress(Value int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().IMEInProgress, uintptr(unsafe.Pointer(this)), uintptr(Value))
	return HRESULT(ret)
}

func (this *ITextDocument2Old) SysBeep() HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SysBeep, uintptr(unsafe.Pointer(this)))
	return HRESULT(ret)
}

func (this *ITextDocument2Old) Update(Mode int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Update, uintptr(unsafe.Pointer(this)), uintptr(Mode))
	return HRESULT(ret)
}

func (this *ITextDocument2Old) Notify(Notify int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Notify, uintptr(unsafe.Pointer(this)), uintptr(Notify))
	return HRESULT(ret)
}

func (this *ITextDocument2Old) GetDocumentFont(ppITextFont **ITextFont) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetDocumentFont, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(ppITextFont)))
	return HRESULT(ret)
}

func (this *ITextDocument2Old) GetDocumentPara(ppITextPara **ITextPara) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetDocumentPara, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(ppITextPara)))
	return HRESULT(ret)
}

func (this *ITextDocument2Old) GetCallManager(ppVoid **IUnknown) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetCallManager, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(ppVoid)))
	return HRESULT(ret)
}

func (this *ITextDocument2Old) ReleaseCallManager(pVoid *IUnknown) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().ReleaseCallManager, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pVoid)))
	return HRESULT(ret)
}
