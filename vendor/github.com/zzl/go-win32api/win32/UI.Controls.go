package win32

import (
	"syscall"
	"unsafe"
)

type (
	HPROPSHEETPAGE          = uintptr
	HIMAGELIST              = uintptr
	HSYNTHETICPOINTERDEVICE = uintptr
	HTREEITEM               = uintptr
	HDSA                    = uintptr
	HDPA                    = uintptr
)

const (
	TVI_ROOT                      HTREEITEM = ^HTREEITEM(0xffff)
	TVI_FIRST                     HTREEITEM = ^HTREEITEM(0xfffe)
	TVI_LAST                      HTREEITEM = ^HTREEITEM(0xfffd)
	TVI_SORT                      HTREEITEM = ^HTREEITEM(0xfffc)
	BCN_FIRST                     uint32    = 0xfffffb1e
	EM_SCROLLCARET                uint32    = 0xb7
	EM_SETLIMITTEXT               uint32    = 0xc5
	EM_GETLIMITTEXT               uint32    = 0xd5
	EM_POSFROMCHAR                uint32    = 0xd6
	EM_CHARFROMPOS                uint32    = 0xd7
	CtlFirst                      uint32    = 0x400
	CtlLast                       uint32    = 0x4ff
	Psh1                          uint32    = 0x400
	Psh2                          uint32    = 0x401
	Psh3                          uint32    = 0x402
	Psh4                          uint32    = 0x403
	Psh5                          uint32    = 0x404
	Psh6                          uint32    = 0x405
	Psh7                          uint32    = 0x406
	Psh8                          uint32    = 0x407
	Psh9                          uint32    = 0x408
	Psh10                         uint32    = 0x409
	Psh11                         uint32    = 0x40a
	Psh12                         uint32    = 0x40b
	Psh13                         uint32    = 0x40c
	Psh14                         uint32    = 0x40d
	Psh15                         uint32    = 0x40e
	PshHelp                       uint32    = 0x40e
	Psh16                         uint32    = 0x40f
	Chx1                          uint32    = 0x410
	Chx2                          uint32    = 0x411
	Chx3                          uint32    = 0x412
	Chx4                          uint32    = 0x413
	Chx5                          uint32    = 0x414
	Chx6                          uint32    = 0x415
	Chx7                          uint32    = 0x416
	Chx8                          uint32    = 0x417
	Chx9                          uint32    = 0x418
	Chx10                         uint32    = 0x419
	Chx11                         uint32    = 0x41a
	Chx12                         uint32    = 0x41b
	Chx13                         uint32    = 0x41c
	Chx14                         uint32    = 0x41d
	Chx15                         uint32    = 0x41e
	Chx16                         uint32    = 0x41f
	Rad1                          uint32    = 0x420
	Rad2                          uint32    = 0x421
	Rad3                          uint32    = 0x422
	Rad4                          uint32    = 0x423
	Rad5                          uint32    = 0x424
	Rad6                          uint32    = 0x425
	Rad7                          uint32    = 0x426
	Rad8                          uint32    = 0x427
	Rad9                          uint32    = 0x428
	Rad10                         uint32    = 0x429
	Rad11                         uint32    = 0x42a
	Rad12                         uint32    = 0x42b
	Rad13                         uint32    = 0x42c
	Rad14                         uint32    = 0x42d
	Rad15                         uint32    = 0x42e
	Rad16                         uint32    = 0x42f
	Grp1                          uint32    = 0x430
	Grp2                          uint32    = 0x431
	Grp3                          uint32    = 0x432
	Grp4                          uint32    = 0x433
	Frm1                          uint32    = 0x434
	Frm2                          uint32    = 0x435
	Frm3                          uint32    = 0x436
	Frm4                          uint32    = 0x437
	Rct1                          uint32    = 0x438
	Rct2                          uint32    = 0x439
	Rct3                          uint32    = 0x43a
	Rct4                          uint32    = 0x43b
	Ico1                          uint32    = 0x43c
	Ico2                          uint32    = 0x43d
	Ico3                          uint32    = 0x43e
	Ico4                          uint32    = 0x43f
	Stc1                          uint32    = 0x440
	Stc2                          uint32    = 0x441
	Stc3                          uint32    = 0x442
	Stc4                          uint32    = 0x443
	Stc5                          uint32    = 0x444
	Stc6                          uint32    = 0x445
	Stc7                          uint32    = 0x446
	Stc8                          uint32    = 0x447
	Stc9                          uint32    = 0x448
	Stc10                         uint32    = 0x449
	Stc11                         uint32    = 0x44a
	Stc12                         uint32    = 0x44b
	Stc13                         uint32    = 0x44c
	Stc14                         uint32    = 0x44d
	Stc15                         uint32    = 0x44e
	Stc16                         uint32    = 0x44f
	Stc17                         uint32    = 0x450
	Stc18                         uint32    = 0x451
	Stc19                         uint32    = 0x452
	Stc20                         uint32    = 0x453
	Stc21                         uint32    = 0x454
	Stc22                         uint32    = 0x455
	Stc23                         uint32    = 0x456
	Stc24                         uint32    = 0x457
	Stc25                         uint32    = 0x458
	Stc26                         uint32    = 0x459
	Stc27                         uint32    = 0x45a
	Stc28                         uint32    = 0x45b
	Stc29                         uint32    = 0x45c
	Stc30                         uint32    = 0x45d
	Stc31                         uint32    = 0x45e
	Stc32                         uint32    = 0x45f
	Lst1                          uint32    = 0x460
	Lst2                          uint32    = 0x461
	Lst3                          uint32    = 0x462
	Lst4                          uint32    = 0x463
	Lst5                          uint32    = 0x464
	Lst6                          uint32    = 0x465
	Lst7                          uint32    = 0x466
	Lst8                          uint32    = 0x467
	Lst9                          uint32    = 0x468
	Lst10                         uint32    = 0x469
	Lst11                         uint32    = 0x46a
	Lst12                         uint32    = 0x46b
	Lst13                         uint32    = 0x46c
	Lst14                         uint32    = 0x46d
	Lst15                         uint32    = 0x46e
	Lst16                         uint32    = 0x46f
	Cmb1                          uint32    = 0x470
	Cmb2                          uint32    = 0x471
	Cmb3                          uint32    = 0x472
	Cmb4                          uint32    = 0x473
	Cmb5                          uint32    = 0x474
	Cmb6                          uint32    = 0x475
	Cmb7                          uint32    = 0x476
	Cmb8                          uint32    = 0x477
	Cmb9                          uint32    = 0x478
	Cmb10                         uint32    = 0x479
	Cmb11                         uint32    = 0x47a
	Cmb12                         uint32    = 0x47b
	Cmb13                         uint32    = 0x47c
	Cmb14                         uint32    = 0x47d
	Cmb15                         uint32    = 0x47e
	Cmb16                         uint32    = 0x47f
	Edt1                          uint32    = 0x480
	Edt2                          uint32    = 0x481
	Edt3                          uint32    = 0x482
	Edt4                          uint32    = 0x483
	Edt5                          uint32    = 0x484
	Edt6                          uint32    = 0x485
	Edt7                          uint32    = 0x486
	Edt8                          uint32    = 0x487
	Edt9                          uint32    = 0x488
	Edt10                         uint32    = 0x489
	Edt11                         uint32    = 0x48a
	Edt12                         uint32    = 0x48b
	Edt13                         uint32    = 0x48c
	Edt14                         uint32    = 0x48d
	Edt15                         uint32    = 0x48e
	Edt16                         uint32    = 0x48f
	Scr1                          uint32    = 0x490
	Scr2                          uint32    = 0x491
	Scr3                          uint32    = 0x492
	Scr4                          uint32    = 0x493
	Scr5                          uint32    = 0x494
	Scr6                          uint32    = 0x495
	Scr7                          uint32    = 0x496
	Scr8                          uint32    = 0x497
	Ctl1                          uint32    = 0x4a0
	FILEOPENORD                   uint32    = 0x600
	MULTIFILEOPENORD              uint32    = 0x601
	PRINTDLGORD                   uint32    = 0x602
	PRNSETUPDLGORD                uint32    = 0x603
	FINDDLGORD                    uint32    = 0x604
	REPLACEDLGORD                 uint32    = 0x605
	FONTDLGORD                    uint32    = 0x606
	FORMATDLGORD31                uint32    = 0x607
	FORMATDLGORD30                uint32    = 0x608
	RUNDLGORD                     uint32    = 0x609
	PAGESETUPDLGORD               uint32    = 0x60a
	NEWFILEOPENORD                uint32    = 0x60b
	PRINTDLGEXORD                 uint32    = 0x60d
	PAGESETUPDLGORDMOTIF          uint32    = 0x60e
	COLORMGMTDLGORD               uint32    = 0x60f
	NEWFILEOPENV2ORD              uint32    = 0x610
	NEWFILEOPENV3ORD              uint32    = 0x611
	NEWFORMATDLGWITHLINK          uint32    = 0x637
	IDC_MANAGE_LINK               uint32    = 0x638
	WM_CTLCOLOR                   uint32    = 0x19
	ILDRF_IMAGELOWQUALITY         uint32    = 0x1
	ILDRF_OVERLAYLOWQUALITY       uint32    = 0x10
	ILR_DEFAULT                   uint32    = 0x0
	ILR_HORIZONTAL_LEFT           uint32    = 0x0
	ILR_HORIZONTAL_CENTER         uint32    = 0x1
	ILR_HORIZONTAL_RIGHT          uint32    = 0x2
	ILR_VERTICAL_TOP              uint32    = 0x0
	ILR_VERTICAL_CENTER           uint32    = 0x10
	ILR_VERTICAL_BOTTOM           uint32    = 0x20
	ILR_SCALE_CLIP                uint32    = 0x0
	ILR_SCALE_ASPECTRATIO         uint32    = 0x100
	ILGOS_ALWAYS                  uint32    = 0x0
	ILGOS_FROMSTANDBY             uint32    = 0x1
	ILFIP_ALWAYS                  uint32    = 0x0
	ILFIP_FROMSTANDBY             uint32    = 0x1
	ILDI_PURGE                    uint32    = 0x1
	ILDI_STANDBY                  uint32    = 0x2
	ILDI_RESETACCESS              uint32    = 0x4
	ILDI_QUERYACCESS              uint32    = 0x8
	MAXPROPPAGES                  uint32    = 0x64
	PSP_DEFAULT                   uint32    = 0x0
	PSP_DLGINDIRECT               uint32    = 0x1
	PSP_USEHICON                  uint32    = 0x2
	PSP_USEICONID                 uint32    = 0x4
	PSP_USETITLE                  uint32    = 0x8
	PSP_RTLREADING                uint32    = 0x10
	PSP_HASHELP                   uint32    = 0x20
	PSP_USEREFPARENT              uint32    = 0x40
	PSP_USECALLBACK               uint32    = 0x80
	PSP_PREMATURE                 uint32    = 0x400
	PSP_HIDEHEADER                uint32    = 0x800
	PSP_USEHEADERTITLE            uint32    = 0x1000
	PSP_USEHEADERSUBTITLE         uint32    = 0x2000
	PSP_USEFUSIONCONTEXT          uint32    = 0x4000
	PSH_DEFAULT                   uint32    = 0x0
	PSH_PROPTITLE                 uint32    = 0x1
	PSH_USEHICON                  uint32    = 0x2
	PSH_USEICONID                 uint32    = 0x4
	PSH_PROPSHEETPAGE             uint32    = 0x8
	PSH_WIZARDHASFINISH           uint32    = 0x10
	PSH_WIZARD                    uint32    = 0x20
	PSH_USEPSTARTPAGE             uint32    = 0x40
	PSH_NOAPPLYNOW                uint32    = 0x80
	PSH_USECALLBACK               uint32    = 0x100
	PSH_HASHELP                   uint32    = 0x200
	PSH_MODELESS                  uint32    = 0x400
	PSH_RTLREADING                uint32    = 0x800
	PSH_WIZARDCONTEXTHELP         uint32    = 0x1000
	PSH_WIZARD97                  uint32    = 0x2000
	PSH_WATERMARK                 uint32    = 0x8000
	PSH_USEHBMWATERMARK           uint32    = 0x10000
	PSH_USEHPLWATERMARK           uint32    = 0x20000
	PSH_STRETCHWATERMARK          uint32    = 0x40000
	PSH_HEADER                    uint32    = 0x80000
	PSH_USEHBMHEADER              uint32    = 0x100000
	PSH_USEPAGELANG               uint32    = 0x200000
	PSH_WIZARD_LITE               uint32    = 0x400000
	PSH_NOCONTEXTHELP             uint32    = 0x2000000
	PSH_AEROWIZARD                uint32    = 0x4000
	PSH_RESIZABLE                 uint32    = 0x4000000
	PSH_HEADERBITMAP              uint32    = 0x8000000
	PSH_NOMARGIN                  uint32    = 0x10000000
	PSCB_INITIALIZED              uint32    = 0x1
	PSCB_PRECREATE                uint32    = 0x2
	PSCB_BUTTONPRESSED            uint32    = 0x3
	PSNRET_NOERROR                uint32    = 0x0
	PSNRET_INVALID                uint32    = 0x1
	PSNRET_INVALID_NOCHANGEPAGE   uint32    = 0x2
	PSNRET_MESSAGEHANDLED         uint32    = 0x3
	PSM_SETCURSEL                 uint32    = 0x465
	PSM_REMOVEPAGE                uint32    = 0x466
	PSM_ADDPAGE                   uint32    = 0x467
	PSM_CHANGED                   uint32    = 0x468
	PSM_RESTARTWINDOWS            uint32    = 0x469
	PSM_REBOOTSYSTEM              uint32    = 0x46a
	PSM_CANCELTOCLOSE             uint32    = 0x46b
	PSM_QUERYSIBLINGS             uint32    = 0x46c
	PSM_UNCHANGED                 uint32    = 0x46d
	PSM_APPLY                     uint32    = 0x46e
	PSM_SETTITLEA                 uint32    = 0x46f
	PSM_SETTITLEW                 uint32    = 0x478
	PSM_SETTITLE                  uint32    = 0x478
	PSM_SETWIZBUTTONS             uint32    = 0x470
	PSWIZB_BACK                   uint32    = 0x1
	PSWIZB_NEXT                   uint32    = 0x2
	PSWIZB_FINISH                 uint32    = 0x4
	PSWIZB_DISABLEDFINISH         uint32    = 0x8
	PSWIZBF_ELEVATIONREQUIRED     uint32    = 0x1
	PSWIZB_CANCEL                 uint32    = 0x10
	PSM_PRESSBUTTON               uint32    = 0x471
	PSBTN_BACK                    uint32    = 0x0
	PSBTN_NEXT                    uint32    = 0x1
	PSBTN_FINISH                  uint32    = 0x2
	PSBTN_OK                      uint32    = 0x3
	PSBTN_APPLYNOW                uint32    = 0x4
	PSBTN_CANCEL                  uint32    = 0x5
	PSBTN_HELP                    uint32    = 0x6
	PSBTN_MAX                     uint32    = 0x6
	PSM_SETCURSELID               uint32    = 0x472
	PSM_SETFINISHTEXTA            uint32    = 0x473
	PSM_SETFINISHTEXTW            uint32    = 0x479
	PSM_SETFINISHTEXT             uint32    = 0x479
	PSM_GETTABCONTROL             uint32    = 0x474
	PSM_ISDIALOGMESSAGE           uint32    = 0x475
	PSM_GETCURRENTPAGEHWND        uint32    = 0x476
	PSM_INSERTPAGE                uint32    = 0x477
	PSM_SETHEADERTITLEA           uint32    = 0x47d
	PSM_SETHEADERTITLEW           uint32    = 0x47e
	PSM_SETHEADERTITLE            uint32    = 0x47e
	PSM_SETHEADERSUBTITLEA        uint32    = 0x47f
	PSM_SETHEADERSUBTITLEW        uint32    = 0x480
	PSM_SETHEADERSUBTITLE         uint32    = 0x480
	PSM_HWNDTOINDEX               uint32    = 0x481
	PSM_INDEXTOHWND               uint32    = 0x482
	PSM_PAGETOINDEX               uint32    = 0x483
	PSM_INDEXTOPAGE               uint32    = 0x484
	PSM_IDTOINDEX                 uint32    = 0x485
	PSM_INDEXTOID                 uint32    = 0x486
	PSM_GETRESULT                 uint32    = 0x487
	PSM_RECALCPAGESIZES           uint32    = 0x488
	PSM_SETNEXTTEXTW              uint32    = 0x489
	PSM_SETNEXTTEXT               uint32    = 0x489
	PSWIZB_SHOW                   uint32    = 0x0
	PSWIZB_RESTORE                uint32    = 0x1
	PSM_SHOWWIZBUTTONS            uint32    = 0x48a
	PSM_ENABLEWIZBUTTONS          uint32    = 0x48b
	PSM_SETBUTTONTEXTW            uint32    = 0x48c
	PSM_SETBUTTONTEXT             uint32    = 0x48c
	ID_PSRESTARTWINDOWS           uint32    = 0x2
	WIZ_CXDLG                     uint32    = 0x114
	WIZ_CYDLG                     uint32    = 0x8c
	WIZ_CXBMP                     uint32    = 0x50
	WIZ_BODYX                     uint32    = 0x5c
	WIZ_BODYCX                    uint32    = 0xb8
	PROP_SM_CXDLG                 uint32    = 0xd4
	PROP_SM_CYDLG                 uint32    = 0xbc
	PROP_MED_CXDLG                uint32    = 0xe3
	PROP_MED_CYDLG                uint32    = 0xd7
	PROP_LG_CXDLG                 uint32    = 0xfc
	PROP_LG_CYDLG                 uint32    = 0xda
	DA_LAST                       uint32    = 0x7fffffff
	DA_ERR                        int32     = -1
	DSA_APPEND                    uint32    = 0x7fffffff
	DSA_ERR                       int32     = -1
	DPAM_SORTED                   uint32    = 0x1
	DPAM_NORMAL                   uint32    = 0x2
	DPAM_UNION                    uint32    = 0x4
	DPAM_INTERSECT                uint32    = 0x8
	DPAS_SORTED                   uint32    = 0x1
	DPAS_INSERTBEFORE             uint32    = 0x2
	DPAS_INSERTAFTER              uint32    = 0x4
	DPA_APPEND                    uint32    = 0x7fffffff
	DPA_ERR                       int32     = -1
	MAX_THEMECOLOR                uint32    = 0x40
	MAX_THEMESIZE                 uint32    = 0x40
	DTBG_CLIPRECT                 uint32    = 0x1
	DTBG_DRAWSOLID                uint32    = 0x2
	DTBG_OMITBORDER               uint32    = 0x4
	DTBG_OMITCONTENT              uint32    = 0x8
	DTBG_COMPUTINGREGION          uint32    = 0x10
	DTBG_MIRRORDC                 uint32    = 0x20
	DTBG_NOMIRROR                 uint32    = 0x40
	DTT_GRAYED                    uint32    = 0x1
	DTT_FLAGS2VALIDBITS           uint32    = 0x1
	HTTB_BACKGROUNDSEG            uint32    = 0x0
	HTTB_FIXEDBORDER              uint32    = 0x2
	HTTB_CAPTION                  uint32    = 0x4
	HTTB_RESIZINGBORDER_LEFT      uint32    = 0x10
	HTTB_RESIZINGBORDER_TOP       uint32    = 0x20
	HTTB_RESIZINGBORDER_RIGHT     uint32    = 0x40
	HTTB_RESIZINGBORDER_BOTTOM    uint32    = 0x80
	HTTB_SIZINGTEMPLATE           uint32    = 0x100
	HTTB_SYSTEMSIZINGMARGINS      uint32    = 0x200
	MAX_INTLIST_COUNT             uint32    = 0x192
	ETDT_DISABLE                  uint32    = 0x1
	ETDT_ENABLE                   uint32    = 0x2
	ETDT_USETABTEXTURE            uint32    = 0x4
	ETDT_USEAEROWIZARDTABTEXTURE  uint32    = 0x8
	SZ_THDOCPROP_DISPLAYNAME      string    = "DisplayName"
	SZ_THDOCPROP_CANONICALNAME    string    = "ThemeName"
	SZ_THDOCPROP_TOOLTIP          string    = "ToolTip"
	SZ_THDOCPROP_AUTHOR           string    = "author"
	WTNCA_NODRAWCAPTION           uint32    = 0x1
	WTNCA_NODRAWICON              uint32    = 0x2
	WTNCA_NOSYSMENU               uint32    = 0x4
	WTNCA_NOMIRRORHELP            uint32    = 0x8
	ODT_HEADER                    uint32    = 0x64
	LVM_FIRST                     uint32    = 0x1000
	TV_FIRST                      uint32    = 0x1100
	HDM_FIRST                     uint32    = 0x1200
	TCM_FIRST                     uint32    = 0x1300
	PGM_FIRST                     uint32    = 0x1400
	ECM_FIRST                     uint32    = 0x1500
	BCM_FIRST                     uint32    = 0x1600
	CBM_FIRST                     uint32    = 0x1700
	CCM_FIRST                     uint32    = 0x2000
	CCM_LAST                      uint32    = 0x2200
	CCM_SETBKCOLOR                uint32    = 0x2001
	CCM_SETCOLORSCHEME            uint32    = 0x2002
	CCM_GETCOLORSCHEME            uint32    = 0x2003
	CCM_GETDROPTARGET             uint32    = 0x2004
	CCM_SETUNICODEFORMAT          uint32    = 0x2005
	CCM_GETUNICODEFORMAT          uint32    = 0x2006
	COMCTL32_VERSION              uint32    = 0x6
	CCM_SETVERSION                uint32    = 0x2007
	CCM_GETVERSION                uint32    = 0x2008
	CCM_SETNOTIFYWINDOW           uint32    = 0x2009
	CCM_SETWINDOWTHEME            uint32    = 0x200b
	CCM_DPISCALE                  uint32    = 0x200c
	INFOTIPSIZE                   uint32    = 0x400
	MSGF_COMMCTRL_BEGINDRAG       uint32    = 0x4200
	MSGF_COMMCTRL_SIZEHEADER      uint32    = 0x4201
	MSGF_COMMCTRL_DRAGSELECT      uint32    = 0x4202
	MSGF_COMMCTRL_TOOLBARCUST     uint32    = 0x4203
	CDRF_DODEFAULT                uint32    = 0x0
	CDRF_NEWFONT                  uint32    = 0x2
	CDRF_SKIPDEFAULT              uint32    = 0x4
	CDRF_DOERASE                  uint32    = 0x8
	CDRF_SKIPPOSTPAINT            uint32    = 0x100
	CDRF_NOTIFYPOSTPAINT          uint32    = 0x10
	CDRF_NOTIFYITEMDRAW           uint32    = 0x20
	CDRF_NOTIFYSUBITEMDRAW        uint32    = 0x20
	CDRF_NOTIFYPOSTERASE          uint32    = 0x40
	CDDS_POSTERASE                uint32    = 0x4
	CDDS_ITEM                     uint32    = 0x10000
	CDIS_SELECTED                 uint32    = 0x1
	CDIS_GRAYED                   uint32    = 0x2
	CDIS_DISABLED                 uint32    = 0x4
	CDIS_CHECKED                  uint32    = 0x8
	CDIS_FOCUS                    uint32    = 0x10
	CDIS_DEFAULT                  uint32    = 0x20
	CDIS_HOT                      uint32    = 0x40
	CDIS_MARKED                   uint32    = 0x80
	CDIS_INDETERMINATE            uint32    = 0x100
	CDIS_SHOWKEYBOARDCUES         uint32    = 0x200
	CDIS_NEARHOT                  uint32    = 0x400
	CDIS_OTHERSIDEHOT             uint32    = 0x800
	CDIS_DROPHILITED              uint32    = 0x1000
	NM_GETCUSTOMSPLITRECT         uint32    = 0xfffffb21
	CLR_NONE                      int32     = -1
	CLR_DEFAULT                   int32     = -16777216
	ILD_TRANSPARENT               uint32    = 0x1
	ILD_IMAGE                     uint32    = 0x20
	ILD_ROP                       uint32    = 0x40
	ILD_BLEND25                   uint32    = 0x2
	ILD_OVERLAYMASK               uint32    = 0xf00
	ILD_PRESERVEALPHA             uint32    = 0x1000
	ILD_SCALE                     uint32    = 0x2000
	ILD_DPISCALE                  uint32    = 0x4000
	ILD_ASYNC                     uint32    = 0x8000
	CLR_HILIGHT                   int32     = -16777216
	ILS_NORMAL                    uint32    = 0x0
	ILS_GLOW                      uint32    = 0x1
	ILS_SHADOW                    uint32    = 0x2
	ILS_SATURATE                  uint32    = 0x4
	ILS_ALPHA                     uint32    = 0x8
	ILGT_NORMAL                   uint32    = 0x0
	ILGT_ASYNC                    uint32    = 0x1
	ILP_NORMAL                    uint32    = 0x0
	ILP_DOWNLEVEL                 uint32    = 0x1
	WC_HEADERA                    string    = "SysHeader32"
	WC_HEADERW                    string    = "SysHeader32"
	WC_HEADER                     string    = "SysHeader32"
	HDS_HORZ                      uint32    = 0x0
	HDS_BUTTONS                   uint32    = 0x2
	HDS_HOTTRACK                  uint32    = 0x4
	HDS_HIDDEN                    uint32    = 0x8
	HDS_DRAGDROP                  uint32    = 0x40
	HDS_FULLDRAG                  uint32    = 0x80
	HDS_FILTERBAR                 uint32    = 0x100
	HDS_FLAT                      uint32    = 0x200
	HDS_CHECKBOXES                uint32    = 0x400
	HDS_NOSIZING                  uint32    = 0x800
	HDS_OVERFLOW                  uint32    = 0x1000
	HDFT_ISSTRING                 uint32    = 0x0
	HDFT_ISNUMBER                 uint32    = 0x1
	HDFT_ISDATE                   uint32    = 0x2
	HDFT_HASNOVALUE               uint32    = 0x8000
	HDF_LEFT                      uint32    = 0x0
	HDF_RIGHT                     uint32    = 0x1
	HDF_CENTER                    uint32    = 0x2
	HDF_JUSTIFYMASK               uint32    = 0x3
	HDF_RTLREADING                uint32    = 0x4
	HDF_BITMAP                    uint32    = 0x2000
	HDF_STRING                    uint32    = 0x4000
	HDF_OWNERDRAW                 uint32    = 0x8000
	HDF_IMAGE                     uint32    = 0x800
	HDF_BITMAP_ON_RIGHT           uint32    = 0x1000
	HDF_SORTUP                    uint32    = 0x400
	HDF_SORTDOWN                  uint32    = 0x200
	HDF_CHECKBOX                  uint32    = 0x40
	HDF_CHECKED                   uint32    = 0x80
	HDF_FIXEDWIDTH                uint32    = 0x100
	HDF_SPLITBUTTON               uint32    = 0x1000000
	HDIS_FOCUSED                  uint32    = 0x1
	HDM_GETITEMCOUNT              uint32    = 0x1200
	HDM_INSERTITEMA               uint32    = 0x1201
	HDM_INSERTITEMW               uint32    = 0x120a
	HDM_INSERTITEM                uint32    = 0x120a
	HDM_DELETEITEM                uint32    = 0x1202
	HDM_GETITEMA                  uint32    = 0x1203
	HDM_GETITEMW                  uint32    = 0x120b
	HDM_GETITEM                   uint32    = 0x120b
	HDM_SETITEMA                  uint32    = 0x1204
	HDM_SETITEMW                  uint32    = 0x120c
	HDM_SETITEM                   uint32    = 0x120c
	HDM_LAYOUT                    uint32    = 0x1205
	HHT_NOWHERE                   uint32    = 0x1
	HHT_ONHEADER                  uint32    = 0x2
	HHT_ONDIVIDER                 uint32    = 0x4
	HHT_ONDIVOPEN                 uint32    = 0x8
	HHT_ONFILTER                  uint32    = 0x10
	HHT_ONFILTERBUTTON            uint32    = 0x20
	HHT_ABOVE                     uint32    = 0x100
	HHT_BELOW                     uint32    = 0x200
	HHT_TORIGHT                   uint32    = 0x400
	HHT_TOLEFT                    uint32    = 0x800
	HHT_ONITEMSTATEICON           uint32    = 0x1000
	HHT_ONDROPDOWN                uint32    = 0x2000
	HHT_ONOVERFLOW                uint32    = 0x4000
	HDSIL_NORMAL                  uint32    = 0x0
	HDSIL_STATE                   uint32    = 0x1
	HDM_HITTEST                   uint32    = 0x1206
	HDM_GETITEMRECT               uint32    = 0x1207
	HDM_SETIMAGELIST              uint32    = 0x1208
	HDM_GETIMAGELIST              uint32    = 0x1209
	HDM_ORDERTOINDEX              uint32    = 0x120f
	HDM_CREATEDRAGIMAGE           uint32    = 0x1210
	HDM_GETORDERARRAY             uint32    = 0x1211
	HDM_SETORDERARRAY             uint32    = 0x1212
	HDM_SETHOTDIVIDER             uint32    = 0x1213
	HDM_SETBITMAPMARGIN           uint32    = 0x1214
	HDM_GETBITMAPMARGIN           uint32    = 0x1215
	HDM_SETUNICODEFORMAT          uint32    = 0x2005
	HDM_GETUNICODEFORMAT          uint32    = 0x2006
	HDM_SETFILTERCHANGETIMEOUT    uint32    = 0x1216
	HDM_EDITFILTER                uint32    = 0x1217
	HDM_CLEARFILTER               uint32    = 0x1218
	HDM_GETITEMDROPDOWNRECT       uint32    = 0x1219
	HDM_GETOVERFLOWRECT           uint32    = 0x121a
	HDM_GETFOCUSEDITEM            uint32    = 0x121b
	HDM_SETFOCUSEDITEM            uint32    = 0x121c
	TOOLBARCLASSNAMEW             string    = "ToolbarWindow32"
	TOOLBARCLASSNAMEA             string    = "ToolbarWindow32"
	TOOLBARCLASSNAME              string    = "ToolbarWindow32"
	CMB_MASKED                    uint32    = 0x2
	TBSTATE_CHECKED               uint32    = 0x1
	TBSTATE_PRESSED               uint32    = 0x2
	TBSTATE_ENABLED               uint32    = 0x4
	TBSTATE_HIDDEN                uint32    = 0x8
	TBSTATE_INDETERMINATE         uint32    = 0x10
	TBSTATE_WRAP                  uint32    = 0x20
	TBSTATE_ELLIPSES              uint32    = 0x40
	TBSTATE_MARKED                uint32    = 0x80
	TBSTYLE_BUTTON                uint32    = 0x0
	TBSTYLE_SEP                   uint32    = 0x1
	TBSTYLE_CHECK                 uint32    = 0x2
	TBSTYLE_GROUP                 uint32    = 0x4
	TBSTYLE_DROPDOWN              uint32    = 0x8
	TBSTYLE_AUTOSIZE              uint32    = 0x10
	TBSTYLE_NOPREFIX              uint32    = 0x20
	TBSTYLE_TOOLTIPS              uint32    = 0x100
	TBSTYLE_WRAPABLE              uint32    = 0x200
	TBSTYLE_ALTDRAG               uint32    = 0x400
	TBSTYLE_FLAT                  uint32    = 0x800
	TBSTYLE_LIST                  uint32    = 0x1000
	TBSTYLE_CUSTOMERASE           uint32    = 0x2000
	TBSTYLE_REGISTERDROP          uint32    = 0x4000
	TBSTYLE_TRANSPARENT           uint32    = 0x8000
	TBSTYLE_EX_DRAWDDARROWS       uint32    = 0x1
	BTNS_BUTTON                   uint32    = 0x0
	BTNS_SEP                      uint32    = 0x1
	BTNS_CHECK                    uint32    = 0x2
	BTNS_GROUP                    uint32    = 0x4
	BTNS_DROPDOWN                 uint32    = 0x8
	BTNS_AUTOSIZE                 uint32    = 0x10
	BTNS_NOPREFIX                 uint32    = 0x20
	BTNS_SHOWTEXT                 uint32    = 0x40
	BTNS_WHOLEDROPDOWN            uint32    = 0x80
	TBSTYLE_EX_MIXEDBUTTONS       uint32    = 0x8
	TBSTYLE_EX_HIDECLIPPEDBUTTONS uint32    = 0x10
	TBSTYLE_EX_MULTICOLUMN        uint32    = 0x2
	TBSTYLE_EX_VERTICAL           uint32    = 0x4
	TBSTYLE_EX_DOUBLEBUFFER       uint32    = 0x80
	TBCDRF_NOEDGES                uint32    = 0x10000
	TBCDRF_HILITEHOTTRACK         uint32    = 0x20000
	TBCDRF_NOOFFSET               uint32    = 0x40000
	TBCDRF_NOMARK                 uint32    = 0x80000
	TBCDRF_NOETCHEDEFFECT         uint32    = 0x100000
	TBCDRF_BLENDICON              uint32    = 0x200000
	TBCDRF_NOBACKGROUND           uint32    = 0x400000
	TBCDRF_USECDCOLORS            uint32    = 0x800000
	TB_ENABLEBUTTON               uint32    = 0x401
	TB_CHECKBUTTON                uint32    = 0x402
	TB_PRESSBUTTON                uint32    = 0x403
	TB_HIDEBUTTON                 uint32    = 0x404
	TB_INDETERMINATE              uint32    = 0x405
	TB_MARKBUTTON                 uint32    = 0x406
	TB_ISBUTTONENABLED            uint32    = 0x409
	TB_ISBUTTONCHECKED            uint32    = 0x40a
	TB_ISBUTTONPRESSED            uint32    = 0x40b
	TB_ISBUTTONHIDDEN             uint32    = 0x40c
	TB_ISBUTTONINDETERMINATE      uint32    = 0x40d
	TB_ISBUTTONHIGHLIGHTED        uint32    = 0x40e
	TB_SETSTATE                   uint32    = 0x411
	TB_GETSTATE                   uint32    = 0x412
	TB_ADDBITMAP                  uint32    = 0x413
	IDB_STD_SMALL_COLOR           uint32    = 0x0
	IDB_STD_LARGE_COLOR           uint32    = 0x1
	IDB_VIEW_SMALL_COLOR          uint32    = 0x4
	IDB_VIEW_LARGE_COLOR          uint32    = 0x5
	IDB_HIST_SMALL_COLOR          uint32    = 0x8
	IDB_HIST_LARGE_COLOR          uint32    = 0x9
	IDB_HIST_NORMAL               uint32    = 0xc
	IDB_HIST_HOT                  uint32    = 0xd
	IDB_HIST_DISABLED             uint32    = 0xe
	IDB_HIST_PRESSED              uint32    = 0xf
	STD_CUT                       uint32    = 0x0
	STD_COPY                      uint32    = 0x1
	STD_PASTE                     uint32    = 0x2
	STD_UNDO                      uint32    = 0x3
	STD_REDOW                     uint32    = 0x4
	STD_DELETE                    uint32    = 0x5
	STD_FILENEW                   uint32    = 0x6
	STD_FILEOPEN                  uint32    = 0x7
	STD_FILESAVE                  uint32    = 0x8
	STD_PRINTPRE                  uint32    = 0x9
	STD_PROPERTIES                uint32    = 0xa
	STD_HELP                      uint32    = 0xb
	STD_FIND                      uint32    = 0xc
	STD_REPLACE                   uint32    = 0xd
	STD_PRINT                     uint32    = 0xe
	VIEW_LARGEICONS               uint32    = 0x0
	VIEW_SMALLICONS               uint32    = 0x1
	VIEW_LIST                     uint32    = 0x2
	VIEW_DETAILS                  uint32    = 0x3
	VIEW_SORTNAME                 uint32    = 0x4
	VIEW_SORTSIZE                 uint32    = 0x5
	VIEW_SORTDATE                 uint32    = 0x6
	VIEW_SORTTYPE                 uint32    = 0x7
	VIEW_PARENTFOLDER             uint32    = 0x8
	VIEW_NETCONNECT               uint32    = 0x9
	VIEW_NETDISCONNECT            uint32    = 0xa
	VIEW_NEWFOLDER                uint32    = 0xb
	VIEW_VIEWMENU                 uint32    = 0xc
	HIST_BACK                     uint32    = 0x0
	HIST_FORWARD                  uint32    = 0x1
	HIST_FAVORITES                uint32    = 0x2
	HIST_ADDTOFAVORITES           uint32    = 0x3
	HIST_VIEWTREE                 uint32    = 0x4
	TB_ADDBUTTONSA                uint32    = 0x414
	TB_INSERTBUTTONA              uint32    = 0x415
	TB_DELETEBUTTON               uint32    = 0x416
	TB_GETBUTTON                  uint32    = 0x417
	TB_BUTTONCOUNT                uint32    = 0x418
	TB_COMMANDTOINDEX             uint32    = 0x419
	TB_SAVERESTOREA               uint32    = 0x41a
	TB_SAVERESTOREW               uint32    = 0x44c
	TB_CUSTOMIZE                  uint32    = 0x41b
	TB_ADDSTRINGA                 uint32    = 0x41c
	TB_ADDSTRINGW                 uint32    = 0x44d
	TB_GETITEMRECT                uint32    = 0x41d
	TB_BUTTONSTRUCTSIZE           uint32    = 0x41e
	TB_SETBUTTONSIZE              uint32    = 0x41f
	TB_SETBITMAPSIZE              uint32    = 0x420
	TB_AUTOSIZE                   uint32    = 0x421
	TB_GETTOOLTIPS                uint32    = 0x423
	TB_SETTOOLTIPS                uint32    = 0x424
	TB_SETPARENT                  uint32    = 0x425
	TB_SETROWS                    uint32    = 0x427
	TB_GETROWS                    uint32    = 0x428
	TB_SETCMDID                   uint32    = 0x42a
	TB_CHANGEBITMAP               uint32    = 0x42b
	TB_GETBITMAP                  uint32    = 0x42c
	TB_GETBUTTONTEXTA             uint32    = 0x42d
	TB_GETBUTTONTEXTW             uint32    = 0x44b
	TB_REPLACEBITMAP              uint32    = 0x42e
	TB_SETINDENT                  uint32    = 0x42f
	TB_SETIMAGELIST               uint32    = 0x430
	TB_GETIMAGELIST               uint32    = 0x431
	TB_LOADIMAGES                 uint32    = 0x432
	TB_GETRECT                    uint32    = 0x433
	TB_SETHOTIMAGELIST            uint32    = 0x434
	TB_GETHOTIMAGELIST            uint32    = 0x435
	TB_SETDISABLEDIMAGELIST       uint32    = 0x436
	TB_GETDISABLEDIMAGELIST       uint32    = 0x437
	TB_SETSTYLE                   uint32    = 0x438
	TB_GETSTYLE                   uint32    = 0x439
	TB_GETBUTTONSIZE              uint32    = 0x43a
	TB_SETBUTTONWIDTH             uint32    = 0x43b
	TB_SETMAXTEXTROWS             uint32    = 0x43c
	TB_GETTEXTROWS                uint32    = 0x43d
	TB_GETBUTTONTEXT              uint32    = 0x44b
	TB_SAVERESTORE                uint32    = 0x44c
	TB_ADDSTRING                  uint32    = 0x44d
	TB_GETOBJECT                  uint32    = 0x43e
	TB_GETHOTITEM                 uint32    = 0x447
	TB_SETHOTITEM                 uint32    = 0x448
	TB_SETANCHORHIGHLIGHT         uint32    = 0x449
	TB_GETANCHORHIGHLIGHT         uint32    = 0x44a
	TB_MAPACCELERATORA            uint32    = 0x44e
	TB_GETINSERTMARK              uint32    = 0x44f
	TB_SETINSERTMARK              uint32    = 0x450
	TB_INSERTMARKHITTEST          uint32    = 0x451
	TB_MOVEBUTTON                 uint32    = 0x452
	TB_GETMAXSIZE                 uint32    = 0x453
	TB_SETEXTENDEDSTYLE           uint32    = 0x454
	TB_GETEXTENDEDSTYLE           uint32    = 0x455
	TB_GETPADDING                 uint32    = 0x456
	TB_SETPADDING                 uint32    = 0x457
	TB_SETINSERTMARKCOLOR         uint32    = 0x458
	TB_GETINSERTMARKCOLOR         uint32    = 0x459
	TB_SETCOLORSCHEME             uint32    = 0x2002
	TB_GETCOLORSCHEME             uint32    = 0x2003
	TB_SETUNICODEFORMAT           uint32    = 0x2005
	TB_GETUNICODEFORMAT           uint32    = 0x2006
	TB_MAPACCELERATORW            uint32    = 0x45a
	TB_MAPACCELERATOR             uint32    = 0x45a
	TBBF_LARGE                    uint32    = 0x1
	TB_GETBITMAPFLAGS             uint32    = 0x429
	TB_GETBUTTONINFOW             uint32    = 0x43f
	TB_SETBUTTONINFOW             uint32    = 0x440
	TB_GETBUTTONINFOA             uint32    = 0x441
	TB_SETBUTTONINFOA             uint32    = 0x442
	TB_GETBUTTONINFO              uint32    = 0x43f
	TB_SETBUTTONINFO              uint32    = 0x440
	TB_INSERTBUTTONW              uint32    = 0x443
	TB_ADDBUTTONSW                uint32    = 0x444
	TB_HITTEST                    uint32    = 0x445
	TB_INSERTBUTTON               uint32    = 0x443
	TB_ADDBUTTONS                 uint32    = 0x444
	TB_SETDRAWTEXTFLAGS           uint32    = 0x446
	TB_GETSTRINGW                 uint32    = 0x45b
	TB_GETSTRINGA                 uint32    = 0x45c
	TB_GETSTRING                  uint32    = 0x45b
	TB_SETBOUNDINGSIZE            uint32    = 0x45d
	TB_SETHOTITEM2                uint32    = 0x45e
	TB_HASACCELERATOR             uint32    = 0x45f
	TB_SETLISTGAP                 uint32    = 0x460
	TB_GETIMAGELISTCOUNT          uint32    = 0x462
	TB_GETIDEALSIZE               uint32    = 0x463
	TBMF_PAD                      uint32    = 0x1
	TBMF_BARPAD                   uint32    = 0x2
	TBMF_BUTTONSPACING            uint32    = 0x4
	TB_GETMETRICS                 uint32    = 0x465
	TB_SETMETRICS                 uint32    = 0x466
	TB_GETITEMDROPDOWNRECT        uint32    = 0x467
	TB_SETPRESSEDIMAGELIST        uint32    = 0x468
	TB_GETPRESSEDIMAGELIST        uint32    = 0x469
	TB_SETWINDOWTHEME             uint32    = 0x200b
	TBNRF_HIDEHELP                uint32    = 0x1
	TBNRF_ENDCUSTOMIZE            uint32    = 0x2
	TBDDRET_DEFAULT               uint32    = 0x0
	TBDDRET_NODEFAULT             uint32    = 0x1
	TBDDRET_TREATPRESSED          uint32    = 0x2
	REBARCLASSNAMEW               string    = "ReBarWindow32"
	REBARCLASSNAMEA               string    = "ReBarWindow32"
	REBARCLASSNAME                string    = "ReBarWindow32"
	RBIM_IMAGELIST                uint32    = 0x1
	RBS_TOOLTIPS                  uint32    = 0x100
	RBS_VARHEIGHT                 uint32    = 0x200
	RBS_BANDBORDERS               uint32    = 0x400
	RBS_FIXEDORDER                uint32    = 0x800
	RBS_REGISTERDROP              uint32    = 0x1000
	RBS_AUTOSIZE                  uint32    = 0x2000
	RBS_VERTICALGRIPPER           uint32    = 0x4000
	RBS_DBLCLKTOGGLE              uint32    = 0x8000
	RBBS_BREAK                    uint32    = 0x1
	RBBS_FIXEDSIZE                uint32    = 0x2
	RBBS_CHILDEDGE                uint32    = 0x4
	RBBS_HIDDEN                   uint32    = 0x8
	RBBS_NOVERT                   uint32    = 0x10
	RBBS_FIXEDBMP                 uint32    = 0x20
	RBBS_VARIABLEHEIGHT           uint32    = 0x40
	RBBS_GRIPPERALWAYS            uint32    = 0x80
	RBBS_NOGRIPPER                uint32    = 0x100
	RBBS_USECHEVRON               uint32    = 0x200
	RBBS_HIDETITLE                uint32    = 0x400
	RBBS_TOPALIGN                 uint32    = 0x800
	RBBIM_STYLE                   uint32    = 0x1
	RBBIM_COLORS                  uint32    = 0x2
	RBBIM_TEXT                    uint32    = 0x4
	RBBIM_IMAGE                   uint32    = 0x8
	RBBIM_CHILD                   uint32    = 0x10
	RBBIM_CHILDSIZE               uint32    = 0x20
	RBBIM_SIZE                    uint32    = 0x40
	RBBIM_BACKGROUND              uint32    = 0x80
	RBBIM_ID                      uint32    = 0x100
	RBBIM_IDEALSIZE               uint32    = 0x200
	RBBIM_LPARAM                  uint32    = 0x400
	RBBIM_HEADERSIZE              uint32    = 0x800
	RBBIM_CHEVRONLOCATION         uint32    = 0x1000
	RBBIM_CHEVRONSTATE            uint32    = 0x2000
	RB_INSERTBANDA                uint32    = 0x401
	RB_DELETEBAND                 uint32    = 0x402
	RB_GETBARINFO                 uint32    = 0x403
	RB_SETBARINFO                 uint32    = 0x404
	RB_SETBANDINFOA               uint32    = 0x406
	RB_SETPARENT                  uint32    = 0x407
	RB_HITTEST                    uint32    = 0x408
	RB_GETRECT                    uint32    = 0x409
	RB_INSERTBANDW                uint32    = 0x40a
	RB_SETBANDINFOW               uint32    = 0x40b
	RB_GETBANDCOUNT               uint32    = 0x40c
	RB_GETROWCOUNT                uint32    = 0x40d
	RB_GETROWHEIGHT               uint32    = 0x40e
	RB_IDTOINDEX                  uint32    = 0x410
	RB_GETTOOLTIPS                uint32    = 0x411
	RB_SETTOOLTIPS                uint32    = 0x412
	RB_SETBKCOLOR                 uint32    = 0x413
	RB_GETBKCOLOR                 uint32    = 0x414
	RB_SETTEXTCOLOR               uint32    = 0x415
	RB_GETTEXTCOLOR               uint32    = 0x416
	RBSTR_CHANGERECT              uint32    = 0x1
	RB_SIZETORECT                 uint32    = 0x417
	RB_SETCOLORSCHEME             uint32    = 0x2002
	RB_GETCOLORSCHEME             uint32    = 0x2003
	RB_INSERTBAND                 uint32    = 0x40a
	RB_SETBANDINFO                uint32    = 0x40b
	RB_BEGINDRAG                  uint32    = 0x418
	RB_ENDDRAG                    uint32    = 0x419
	RB_DRAGMOVE                   uint32    = 0x41a
	RB_GETBARHEIGHT               uint32    = 0x41b
	RB_GETBANDINFOW               uint32    = 0x41c
	RB_GETBANDINFOA               uint32    = 0x41d
	RB_GETBANDINFO                uint32    = 0x41c
	RB_MINIMIZEBAND               uint32    = 0x41e
	RB_MAXIMIZEBAND               uint32    = 0x41f
	RB_GETDROPTARGET              uint32    = 0x2004
	RB_GETBANDBORDERS             uint32    = 0x422
	RB_SHOWBAND                   uint32    = 0x423
	RB_SETPALETTE                 uint32    = 0x425
	RB_GETPALETTE                 uint32    = 0x426
	RB_MOVEBAND                   uint32    = 0x427
	RB_SETUNICODEFORMAT           uint32    = 0x2005
	RB_GETUNICODEFORMAT           uint32    = 0x2006
	RB_GETBANDMARGINS             uint32    = 0x428
	RB_SETWINDOWTHEME             uint32    = 0x200b
	RB_SETEXTENDEDSTYLE           uint32    = 0x429
	RB_GETEXTENDEDSTYLE           uint32    = 0x42a
	RB_PUSHCHEVRON                uint32    = 0x42b
	RB_SETBANDWIDTH               uint32    = 0x42c
	RBAB_AUTOSIZE                 uint32    = 0x1
	RBAB_ADDBAND                  uint32    = 0x2
	RBHT_NOWHERE                  uint32    = 0x1
	RBHT_CAPTION                  uint32    = 0x2
	RBHT_CLIENT                   uint32    = 0x3
	RBHT_GRABBER                  uint32    = 0x4
	RBHT_CHEVRON                  uint32    = 0x8
	RBHT_SPLITTER                 uint32    = 0x10
	TOOLTIPS_CLASSW               string    = "tooltips_class32"
	TOOLTIPS_CLASSA               string    = "tooltips_class32"
	TOOLTIPS_CLASS                string    = "tooltips_class32"
	TTS_ALWAYSTIP                 uint32    = 0x1
	TTS_NOPREFIX                  uint32    = 0x2
	TTS_NOANIMATE                 uint32    = 0x10
	TTS_NOFADE                    uint32    = 0x20
	TTS_BALLOON                   uint32    = 0x40
	TTS_CLOSE                     uint32    = 0x80
	TTS_USEVISUALSTYLE            uint32    = 0x100
	TTF_DI_SETITEM                uint32    = 0x8000
	TTDT_AUTOMATIC                uint32    = 0x0
	TTDT_RESHOW                   uint32    = 0x1
	TTDT_AUTOPOP                  uint32    = 0x2
	TTDT_INITIAL                  uint32    = 0x3
	TTM_ACTIVATE                  uint32    = 0x401
	TTM_SETDELAYTIME              uint32    = 0x403
	TTM_ADDTOOLA                  uint32    = 0x404
	TTM_ADDTOOLW                  uint32    = 0x432
	TTM_DELTOOLA                  uint32    = 0x405
	TTM_DELTOOLW                  uint32    = 0x433
	TTM_NEWTOOLRECTA              uint32    = 0x406
	TTM_NEWTOOLRECTW              uint32    = 0x434
	TTM_RELAYEVENT                uint32    = 0x407
	TTM_GETTOOLINFOA              uint32    = 0x408
	TTM_GETTOOLINFOW              uint32    = 0x435
	TTM_SETTOOLINFOA              uint32    = 0x409
	TTM_SETTOOLINFOW              uint32    = 0x436
	TTM_HITTESTA                  uint32    = 0x40a
	TTM_HITTESTW                  uint32    = 0x437
	TTM_GETTEXTA                  uint32    = 0x40b
	TTM_GETTEXTW                  uint32    = 0x438
	TTM_UPDATETIPTEXTA            uint32    = 0x40c
	TTM_UPDATETIPTEXTW            uint32    = 0x439
	TTM_GETTOOLCOUNT              uint32    = 0x40d
	TTM_ENUMTOOLSA                uint32    = 0x40e
	TTM_ENUMTOOLSW                uint32    = 0x43a
	TTM_GETCURRENTTOOLA           uint32    = 0x40f
	TTM_GETCURRENTTOOLW           uint32    = 0x43b
	TTM_WINDOWFROMPOINT           uint32    = 0x410
	TTM_TRACKACTIVATE             uint32    = 0x411
	TTM_TRACKPOSITION             uint32    = 0x412
	TTM_SETTIPBKCOLOR             uint32    = 0x413
	TTM_SETTIPTEXTCOLOR           uint32    = 0x414
	TTM_GETDELAYTIME              uint32    = 0x415
	TTM_GETTIPBKCOLOR             uint32    = 0x416
	TTM_GETTIPTEXTCOLOR           uint32    = 0x417
	TTM_SETMAXTIPWIDTH            uint32    = 0x418
	TTM_GETMAXTIPWIDTH            uint32    = 0x419
	TTM_SETMARGIN                 uint32    = 0x41a
	TTM_GETMARGIN                 uint32    = 0x41b
	TTM_POP                       uint32    = 0x41c
	TTM_UPDATE                    uint32    = 0x41d
	TTM_GETBUBBLESIZE             uint32    = 0x41e
	TTM_ADJUSTRECT                uint32    = 0x41f
	TTM_SETTITLEA                 uint32    = 0x420
	TTM_SETTITLEW                 uint32    = 0x421
	TTM_POPUP                     uint32    = 0x422
	TTM_GETTITLE                  uint32    = 0x423
	TTM_ADDTOOL                   uint32    = 0x432
	TTM_DELTOOL                   uint32    = 0x433
	TTM_NEWTOOLRECT               uint32    = 0x434
	TTM_GETTOOLINFO               uint32    = 0x435
	TTM_SETTOOLINFO               uint32    = 0x436
	TTM_HITTEST                   uint32    = 0x437
	TTM_GETTEXT                   uint32    = 0x438
	TTM_UPDATETIPTEXT             uint32    = 0x439
	TTM_ENUMTOOLS                 uint32    = 0x43a
	TTM_GETCURRENTTOOL            uint32    = 0x43b
	TTM_SETTITLE                  uint32    = 0x421
	TTM_SETWINDOWTHEME            uint32    = 0x200b
	SBARS_SIZEGRIP                uint32    = 0x100
	SBARS_TOOLTIPS                uint32    = 0x800
	SBT_TOOLTIPS                  uint32    = 0x800
	STATUSCLASSNAMEW              string    = "msctls_statusbar32"
	STATUSCLASSNAMEA              string    = "msctls_statusbar32"
	STATUSCLASSNAME               string    = "msctls_statusbar32"
	SB_SETTEXTA                   uint32    = 0x401
	SB_SETTEXTW                   uint32    = 0x40b
	SB_GETTEXTA                   uint32    = 0x402
	SB_GETTEXTW                   uint32    = 0x40d
	SB_GETTEXTLENGTHA             uint32    = 0x403
	SB_GETTEXTLENGTHW             uint32    = 0x40c
	SB_GETTEXT                    uint32    = 0x40d
	SB_SETTEXT                    uint32    = 0x40b
	SB_GETTEXTLENGTH              uint32    = 0x40c
	SB_SETPARTS                   uint32    = 0x404
	SB_GETPARTS                   uint32    = 0x406
	SB_GETBORDERS                 uint32    = 0x407
	SB_SETMINHEIGHT               uint32    = 0x408
	SB_SIMPLE                     uint32    = 0x409
	SB_GETRECT                    uint32    = 0x40a
	SB_ISSIMPLE                   uint32    = 0x40e
	SB_SETICON                    uint32    = 0x40f
	SB_SETTIPTEXTA                uint32    = 0x410
	SB_SETTIPTEXTW                uint32    = 0x411
	SB_GETTIPTEXTA                uint32    = 0x412
	SB_GETTIPTEXTW                uint32    = 0x413
	SB_GETICON                    uint32    = 0x414
	SB_SETUNICODEFORMAT           uint32    = 0x2005
	SB_GETUNICODEFORMAT           uint32    = 0x2006
	SBT_OWNERDRAW                 uint32    = 0x1000
	SBT_NOBORDERS                 uint32    = 0x100
	SBT_POPOUT                    uint32    = 0x200
	SBT_RTLREADING                uint32    = 0x400
	SBT_NOTABPARSING              uint32    = 0x800
	SB_SETBKCOLOR                 uint32    = 0x2001
	SB_SIMPLEID                   uint32    = 0xff
	TRACKBAR_CLASSA               string    = "msctls_trackbar32"
	TRACKBAR_CLASSW               string    = "msctls_trackbar32"
	TRACKBAR_CLASS                string    = "msctls_trackbar32"
	TBS_AUTOTICKS                 uint32    = 0x1
	TBS_VERT                      uint32    = 0x2
	TBS_HORZ                      uint32    = 0x0
	TBS_TOP                       uint32    = 0x4
	TBS_BOTTOM                    uint32    = 0x0
	TBS_LEFT                      uint32    = 0x4
	TBS_RIGHT                     uint32    = 0x0
	TBS_BOTH                      uint32    = 0x8
	TBS_NOTICKS                   uint32    = 0x10
	TBS_ENABLESELRANGE            uint32    = 0x20
	TBS_FIXEDLENGTH               uint32    = 0x40
	TBS_NOTHUMB                   uint32    = 0x80
	TBS_TOOLTIPS                  uint32    = 0x100
	TBS_REVERSED                  uint32    = 0x200
	TBS_DOWNISLEFT                uint32    = 0x400
	TBS_NOTIFYBEFOREMOVE          uint32    = 0x800
	TBS_TRANSPARENTBKGND          uint32    = 0x1000
	TBM_GETRANGEMIN               uint32    = 0x401
	TBM_GETRANGEMAX               uint32    = 0x402
	TBM_GETTIC                    uint32    = 0x403
	TBM_SETTIC                    uint32    = 0x404
	TBM_SETPOS                    uint32    = 0x405
	TBM_SETRANGE                  uint32    = 0x406
	TBM_SETRANGEMIN               uint32    = 0x407
	TBM_SETRANGEMAX               uint32    = 0x408
	TBM_CLEARTICS                 uint32    = 0x409
	TBM_SETSEL                    uint32    = 0x40a
	TBM_SETSELSTART               uint32    = 0x40b
	TBM_SETSELEND                 uint32    = 0x40c
	TBM_GETPTICS                  uint32    = 0x40e
	TBM_GETTICPOS                 uint32    = 0x40f
	TBM_GETNUMTICS                uint32    = 0x410
	TBM_GETSELSTART               uint32    = 0x411
	TBM_GETSELEND                 uint32    = 0x412
	TBM_CLEARSEL                  uint32    = 0x413
	TBM_SETTICFREQ                uint32    = 0x414
	TBM_SETPAGESIZE               uint32    = 0x415
	TBM_GETPAGESIZE               uint32    = 0x416
	TBM_SETLINESIZE               uint32    = 0x417
	TBM_GETLINESIZE               uint32    = 0x418
	TBM_GETTHUMBRECT              uint32    = 0x419
	TBM_GETCHANNELRECT            uint32    = 0x41a
	TBM_SETTHUMBLENGTH            uint32    = 0x41b
	TBM_GETTHUMBLENGTH            uint32    = 0x41c
	TBM_SETTOOLTIPS               uint32    = 0x41d
	TBM_GETTOOLTIPS               uint32    = 0x41e
	TBM_SETTIPSIDE                uint32    = 0x41f
	TBTS_TOP                      uint32    = 0x0
	TBTS_LEFT                     uint32    = 0x1
	TBTS_BOTTOM                   uint32    = 0x2
	TBTS_RIGHT                    uint32    = 0x3
	TBM_SETBUDDY                  uint32    = 0x420
	TBM_GETBUDDY                  uint32    = 0x421
	TBM_SETPOSNOTIFY              uint32    = 0x422
	TBM_SETUNICODEFORMAT          uint32    = 0x2005
	TBM_GETUNICODEFORMAT          uint32    = 0x2006
	TB_LINEUP                     uint32    = 0x0
	TB_LINEDOWN                   uint32    = 0x1
	TB_PAGEUP                     uint32    = 0x2
	TB_PAGEDOWN                   uint32    = 0x3
	TB_THUMBPOSITION              uint32    = 0x4
	TB_THUMBTRACK                 uint32    = 0x5
	TB_TOP                        uint32    = 0x6
	TB_BOTTOM                     uint32    = 0x7
	TB_ENDTRACK                   uint32    = 0x8
	TBCD_TICS                     uint32    = 0x1
	TBCD_THUMB                    uint32    = 0x2
	TBCD_CHANNEL                  uint32    = 0x3
	DL_CURSORSET                  uint32    = 0x0
	DL_STOPCURSOR                 uint32    = 0x1
	DL_COPYCURSOR                 uint32    = 0x2
	DL_MOVECURSOR                 uint32    = 0x3
	DRAGLISTMSGSTRING             string    = "commctrl_DragListMsg"
	UPDOWN_CLASSA                 string    = "msctls_updown32"
	UPDOWN_CLASSW                 string    = "msctls_updown32"
	UPDOWN_CLASS                  string    = "msctls_updown32"
	UD_MAXVAL                     uint32    = 0x7fff
	UDS_WRAP                      uint32    = 0x1
	UDS_SETBUDDYINT               uint32    = 0x2
	UDS_ALIGNRIGHT                uint32    = 0x4
	UDS_ALIGNLEFT                 uint32    = 0x8
	UDS_AUTOBUDDY                 uint32    = 0x10
	UDS_ARROWKEYS                 uint32    = 0x20
	UDS_HORZ                      uint32    = 0x40
	UDS_NOTHOUSANDS               uint32    = 0x80
	UDS_HOTTRACK                  uint32    = 0x100
	UDM_SETRANGE                  uint32    = 0x465
	UDM_GETRANGE                  uint32    = 0x466
	UDM_SETPOS                    uint32    = 0x467
	UDM_GETPOS                    uint32    = 0x468
	UDM_SETBUDDY                  uint32    = 0x469
	UDM_GETBUDDY                  uint32    = 0x46a
	UDM_SETACCEL                  uint32    = 0x46b
	UDM_GETACCEL                  uint32    = 0x46c
	UDM_SETBASE                   uint32    = 0x46d
	UDM_GETBASE                   uint32    = 0x46e
	UDM_SETRANGE32                uint32    = 0x46f
	UDM_GETRANGE32                uint32    = 0x470
	UDM_SETUNICODEFORMAT          uint32    = 0x2005
	UDM_GETUNICODEFORMAT          uint32    = 0x2006
	UDM_SETPOS32                  uint32    = 0x471
	UDM_GETPOS32                  uint32    = 0x472
	PROGRESS_CLASSA               string    = "msctls_progress32"
	PROGRESS_CLASSW               string    = "msctls_progress32"
	PROGRESS_CLASS                string    = "msctls_progress32"
	PBS_SMOOTH                    uint32    = 0x1
	PBS_VERTICAL                  uint32    = 0x4
	PBM_SETRANGE                  uint32    = 0x401
	PBM_SETPOS                    uint32    = 0x402
	PBM_DELTAPOS                  uint32    = 0x403
	PBM_SETSTEP                   uint32    = 0x404
	PBM_STEPIT                    uint32    = 0x405
	PBM_SETRANGE32                uint32    = 0x406
	PBM_GETRANGE                  uint32    = 0x407
	PBM_GETPOS                    uint32    = 0x408
	PBM_SETBARCOLOR               uint32    = 0x409
	PBM_SETBKCOLOR                uint32    = 0x2001
	PBS_MARQUEE                   uint32    = 0x8
	PBM_SETMARQUEE                uint32    = 0x40a
	PBS_SMOOTHREVERSE             uint32    = 0x10
	PBM_GETSTEP                   uint32    = 0x40d
	PBM_GETBKCOLOR                uint32    = 0x40e
	PBM_GETBARCOLOR               uint32    = 0x40f
	PBM_SETSTATE                  uint32    = 0x410
	PBM_GETSTATE                  uint32    = 0x411
	PBST_NORMAL                   uint32    = 0x1
	PBST_ERROR                    uint32    = 0x2
	PBST_PAUSED                   uint32    = 0x3
	HOTKEYF_SHIFT                 uint32    = 0x1
	HOTKEYF_CONTROL               uint32    = 0x2
	HOTKEYF_ALT                   uint32    = 0x4
	HOTKEYF_EXT                   uint32    = 0x80
	HKCOMB_NONE                   uint32    = 0x1
	HKCOMB_S                      uint32    = 0x2
	HKCOMB_C                      uint32    = 0x4
	HKCOMB_A                      uint32    = 0x8
	HKCOMB_SC                     uint32    = 0x10
	HKCOMB_SA                     uint32    = 0x20
	HKCOMB_CA                     uint32    = 0x40
	HKCOMB_SCA                    uint32    = 0x80
	HKM_SETHOTKEY                 uint32    = 0x401
	HKM_GETHOTKEY                 uint32    = 0x402
	HKM_SETRULES                  uint32    = 0x403
	HOTKEY_CLASSA                 string    = "msctls_hotkey32"
	HOTKEY_CLASSW                 string    = "msctls_hotkey32"
	HOTKEY_CLASS                  string    = "msctls_hotkey32"
	CCS_TOP                       int32     = 1
	CCS_NOMOVEY                   int32     = 2
	CCS_BOTTOM                    int32     = 3
	CCS_NORESIZE                  int32     = 4
	CCS_NOPARENTALIGN             int32     = 8
	CCS_ADJUSTABLE                int32     = 32
	CCS_NODIVIDER                 int32     = 64
	CCS_VERT                      int32     = 128
	INVALID_LINK_INDEX            int32     = -1
	MAX_LINKID_TEXT               uint32    = 0x30
	WC_LINK                       string    = "SysLink"
	LWS_TRANSPARENT               uint32    = 0x1
	LWS_IGNORERETURN              uint32    = 0x2
	LWS_NOPREFIX                  uint32    = 0x4
	LWS_USEVISUALSTYLE            uint32    = 0x8
	LWS_USECUSTOMTEXT             uint32    = 0x10
	LWS_RIGHT                     uint32    = 0x20
	LIF_ITEMINDEX                 uint32    = 0x1
	LIF_STATE                     uint32    = 0x2
	LIF_ITEMID                    uint32    = 0x4
	LIF_URL                       uint32    = 0x8
	LIS_FOCUSED                   uint32    = 0x1
	LIS_ENABLED                   uint32    = 0x2
	LIS_VISITED                   uint32    = 0x4
	LIS_HOTTRACK                  uint32    = 0x8
	LIS_DEFAULTCOLORS             uint32    = 0x10
	LM_HITTEST                    uint32    = 0x700
	LM_GETIDEALHEIGHT             uint32    = 0x701
	LM_SETITEM                    uint32    = 0x702
	LM_GETITEM                    uint32    = 0x703
	LM_GETIDEALSIZE               uint32    = 0x701
	WC_LISTVIEWA                  string    = "SysListView32"
	WC_LISTVIEWW                  string    = "SysListView32"
	WC_LISTVIEW                   string    = "SysListView32"
	LVS_ICON                      uint32    = 0x0
	LVS_REPORT                    uint32    = 0x1
	LVS_SMALLICON                 uint32    = 0x2
	LVS_LIST                      uint32    = 0x3
	LVS_TYPEMASK                  uint32    = 0x3
	LVS_SINGLESEL                 uint32    = 0x4
	LVS_SHOWSELALWAYS             uint32    = 0x8
	LVS_SORTASCENDING             uint32    = 0x10
	LVS_SORTDESCENDING            uint32    = 0x20
	LVS_SHAREIMAGELISTS           uint32    = 0x40
	LVS_NOLABELWRAP               uint32    = 0x80
	LVS_AUTOARRANGE               uint32    = 0x100
	LVS_EDITLABELS                uint32    = 0x200
	LVS_OWNERDATA                 uint32    = 0x1000
	LVS_NOSCROLL                  uint32    = 0x2000
	LVS_TYPESTYLEMASK             uint32    = 0xfc00
	LVS_ALIGNTOP                  uint32    = 0x0
	LVS_ALIGNLEFT                 uint32    = 0x800
	LVS_ALIGNMASK                 uint32    = 0xc00
	LVS_OWNERDRAWFIXED            uint32    = 0x400
	LVS_NOCOLUMNHEADER            uint32    = 0x4000
	LVS_NOSORTHEADER              uint32    = 0x8000
	LVM_SETUNICODEFORMAT          uint32    = 0x2005
	LVM_GETUNICODEFORMAT          uint32    = 0x2006
	LVM_GETBKCOLOR                uint32    = 0x1000
	LVM_SETBKCOLOR                uint32    = 0x1001
	LVM_GETIMAGELIST              uint32    = 0x1002
	LVSIL_NORMAL                  uint32    = 0x0
	LVSIL_SMALL                   uint32    = 0x1
	LVSIL_STATE                   uint32    = 0x2
	LVSIL_GROUPHEADER             uint32    = 0x3
	LVM_SETIMAGELIST              uint32    = 0x1003
	LVM_GETITEMCOUNT              uint32    = 0x1004
	LVIF_TEXT                     uint32    = 0x1
	LVIF_IMAGE                    uint32    = 0x2
	LVIF_PARAM                    uint32    = 0x4
	LVIF_STATE                    uint32    = 0x8
	LVIF_INDENT                   uint32    = 0x10
	LVIF_NORECOMPUTE              uint32    = 0x800
	LVIF_GROUPID                  uint32    = 0x100
	LVIF_COLUMNS                  uint32    = 0x200
	LVIF_COLFMT                   uint32    = 0x10000
	LVIS_FOCUSED                  uint32    = 0x1
	LVIS_SELECTED                 uint32    = 0x2
	LVIS_CUT                      uint32    = 0x4
	LVIS_DROPHILITED              uint32    = 0x8
	LVIS_GLOW                     uint32    = 0x10
	LVIS_ACTIVATING               uint32    = 0x20
	LVIS_OVERLAYMASK              uint32    = 0xf00
	LVIS_STATEIMAGEMASK           uint32    = 0xf000
	I_INDENTCALLBACK              int32     = -1
	I_IMAGECALLBACK               int32     = -1
	I_IMAGENONE                   int32     = -2
	LVM_GETITEMA                  uint32    = 0x1005
	LVM_GETITEMW                  uint32    = 0x104b
	LVM_GETITEM                   uint32    = 0x104b
	LVM_SETITEMA                  uint32    = 0x1006
	LVM_SETITEMW                  uint32    = 0x104c
	LVM_SETITEM                   uint32    = 0x104c
	LVM_INSERTITEMA               uint32    = 0x1007
	LVM_INSERTITEMW               uint32    = 0x104d
	LVM_INSERTITEM                uint32    = 0x104d
	LVM_DELETEITEM                uint32    = 0x1008
	LVM_DELETEALLITEMS            uint32    = 0x1009
	LVM_GETCALLBACKMASK           uint32    = 0x100a
	LVM_SETCALLBACKMASK           uint32    = 0x100b
	LVNI_ALL                      uint32    = 0x0
	LVNI_FOCUSED                  uint32    = 0x1
	LVNI_SELECTED                 uint32    = 0x2
	LVNI_CUT                      uint32    = 0x4
	LVNI_DROPHILITED              uint32    = 0x8
	LVNI_VISIBLEORDER             uint32    = 0x10
	LVNI_PREVIOUS                 uint32    = 0x20
	LVNI_VISIBLEONLY              uint32    = 0x40
	LVNI_SAMEGROUPONLY            uint32    = 0x80
	LVNI_ABOVE                    uint32    = 0x100
	LVNI_BELOW                    uint32    = 0x200
	LVNI_TOLEFT                   uint32    = 0x400
	LVNI_TORIGHT                  uint32    = 0x800
	LVM_GETNEXTITEM               uint32    = 0x100c
	LVM_FINDITEMA                 uint32    = 0x100d
	LVM_FINDITEMW                 uint32    = 0x1053
	LVM_FINDITEM                  uint32    = 0x1053
	LVIR_BOUNDS                   uint32    = 0x0
	LVIR_ICON                     uint32    = 0x1
	LVIR_LABEL                    uint32    = 0x2
	LVIR_SELECTBOUNDS             uint32    = 0x3
	LVM_GETITEMRECT               uint32    = 0x100e
	LVM_SETITEMPOSITION           uint32    = 0x100f
	LVM_GETITEMPOSITION           uint32    = 0x1010
	LVM_GETSTRINGWIDTHA           uint32    = 0x1011
	LVM_GETSTRINGWIDTHW           uint32    = 0x1057
	LVM_GETSTRINGWIDTH            uint32    = 0x1057
	LVM_HITTEST                   uint32    = 0x1012
	LVM_ENSUREVISIBLE             uint32    = 0x1013
	LVM_SCROLL                    uint32    = 0x1014
	LVM_REDRAWITEMS               uint32    = 0x1015
	LVA_DEFAULT                   uint32    = 0x0
	LVA_ALIGNLEFT                 uint32    = 0x1
	LVA_ALIGNTOP                  uint32    = 0x2
	LVA_SNAPTOGRID                uint32    = 0x5
	LVM_ARRANGE                   uint32    = 0x1016
	LVM_EDITLABELA                uint32    = 0x1017
	LVM_EDITLABELW                uint32    = 0x1076
	LVM_EDITLABEL                 uint32    = 0x1076
	LVM_GETEDITCONTROL            uint32    = 0x1018
	LVCFMT_LINE_BREAK             uint32    = 0x100000
	LVCFMT_FILL                   uint32    = 0x200000
	LVCFMT_WRAP                   uint32    = 0x400000
	LVCFMT_NO_TITLE               uint32    = 0x800000
	LVM_GETCOLUMNA                uint32    = 0x1019
	LVM_GETCOLUMNW                uint32    = 0x105f
	LVM_GETCOLUMN                 uint32    = 0x105f
	LVM_SETCOLUMNA                uint32    = 0x101a
	LVM_SETCOLUMNW                uint32    = 0x1060
	LVM_SETCOLUMN                 uint32    = 0x1060
	LVM_INSERTCOLUMNA             uint32    = 0x101b
	LVM_INSERTCOLUMNW             uint32    = 0x1061
	LVM_INSERTCOLUMN              uint32    = 0x1061
	LVM_DELETECOLUMN              uint32    = 0x101c
	LVM_GETCOLUMNWIDTH            uint32    = 0x101d
	LVSCW_AUTOSIZE                int32     = -1
	LVSCW_AUTOSIZE_USEHEADER      int32     = -2
	LVM_SETCOLUMNWIDTH            uint32    = 0x101e
	LVM_GETHEADER                 uint32    = 0x101f
	LVM_CREATEDRAGIMAGE           uint32    = 0x1021
	LVM_GETVIEWRECT               uint32    = 0x1022
	LVM_GETTEXTCOLOR              uint32    = 0x1023
	LVM_SETTEXTCOLOR              uint32    = 0x1024
	LVM_GETTEXTBKCOLOR            uint32    = 0x1025
	LVM_SETTEXTBKCOLOR            uint32    = 0x1026
	LVM_GETTOPINDEX               uint32    = 0x1027
	LVM_GETCOUNTPERPAGE           uint32    = 0x1028
	LVM_GETORIGIN                 uint32    = 0x1029
	LVM_UPDATE                    uint32    = 0x102a
	LVM_SETITEMSTATE              uint32    = 0x102b
	LVM_GETITEMSTATE              uint32    = 0x102c
	LVM_GETITEMTEXTA              uint32    = 0x102d
	LVM_GETITEMTEXTW              uint32    = 0x1073
	LVM_GETITEMTEXT               uint32    = 0x1073
	LVM_SETITEMTEXTA              uint32    = 0x102e
	LVM_SETITEMTEXTW              uint32    = 0x1074
	LVM_SETITEMTEXT               uint32    = 0x1074
	LVSICF_NOINVALIDATEALL        uint32    = 0x1
	LVSICF_NOSCROLL               uint32    = 0x2
	LVM_SETITEMCOUNT              uint32    = 0x102f
	LVM_SORTITEMS                 uint32    = 0x1030
	LVM_SETITEMPOSITION32         uint32    = 0x1031
	LVM_GETSELECTEDCOUNT          uint32    = 0x1032
	LVM_GETITEMSPACING            uint32    = 0x1033
	LVM_GETISEARCHSTRINGA         uint32    = 0x1034
	LVM_GETISEARCHSTRINGW         uint32    = 0x1075
	LVM_GETISEARCHSTRING          uint32    = 0x1075
	LVM_SETICONSPACING            uint32    = 0x1035
	LVM_SETEXTENDEDLISTVIEWSTYLE  uint32    = 0x1036
	LVM_GETEXTENDEDLISTVIEWSTYLE  uint32    = 0x1037
	LVS_EX_GRIDLINES              uint32    = 0x1
	LVS_EX_SUBITEMIMAGES          uint32    = 0x2
	LVS_EX_CHECKBOXES             uint32    = 0x4
	LVS_EX_TRACKSELECT            uint32    = 0x8
	LVS_EX_HEADERDRAGDROP         uint32    = 0x10
	LVS_EX_FULLROWSELECT          uint32    = 0x20
	LVS_EX_ONECLICKACTIVATE       uint32    = 0x40
	LVS_EX_TWOCLICKACTIVATE       uint32    = 0x80
	LVS_EX_FLATSB                 uint32    = 0x100
	LVS_EX_REGIONAL               uint32    = 0x200
	LVS_EX_INFOTIP                uint32    = 0x400
	LVS_EX_UNDERLINEHOT           uint32    = 0x800
	LVS_EX_UNDERLINECOLD          uint32    = 0x1000
	LVS_EX_MULTIWORKAREAS         uint32    = 0x2000
	LVS_EX_LABELTIP               uint32    = 0x4000
	LVS_EX_BORDERSELECT           uint32    = 0x8000
	LVS_EX_DOUBLEBUFFER           uint32    = 0x10000
	LVS_EX_HIDELABELS             uint32    = 0x20000
	LVS_EX_SINGLEROW              uint32    = 0x40000
	LVS_EX_SNAPTOGRID             uint32    = 0x80000
	LVS_EX_SIMPLESELECT           uint32    = 0x100000
	LVS_EX_JUSTIFYCOLUMNS         uint32    = 0x200000
	LVS_EX_TRANSPARENTBKGND       uint32    = 0x400000
	LVS_EX_TRANSPARENTSHADOWTEXT  uint32    = 0x800000
	LVS_EX_AUTOAUTOARRANGE        uint32    = 0x1000000
	LVS_EX_HEADERINALLVIEWS       uint32    = 0x2000000
	LVS_EX_AUTOCHECKSELECT        uint32    = 0x8000000
	LVS_EX_AUTOSIZECOLUMNS        uint32    = 0x10000000
	LVS_EX_COLUMNSNAPPOINTS       uint32    = 0x40000000
	LVS_EX_COLUMNOVERFLOW         uint32    = 0x80000000
	LVM_GETSUBITEMRECT            uint32    = 0x1038
	LVM_SUBITEMHITTEST            uint32    = 0x1039
	LVM_SETCOLUMNORDERARRAY       uint32    = 0x103a
	LVM_GETCOLUMNORDERARRAY       uint32    = 0x103b
	LVM_SETHOTITEM                uint32    = 0x103c
	LVM_GETHOTITEM                uint32    = 0x103d
	LVM_SETHOTCURSOR              uint32    = 0x103e
	LVM_GETHOTCURSOR              uint32    = 0x103f
	LVM_APPROXIMATEVIEWRECT       uint32    = 0x1040
	LV_MAX_WORKAREAS              uint32    = 0x10
	LVM_SETWORKAREAS              uint32    = 0x1041
	LVM_GETWORKAREAS              uint32    = 0x1046
	LVM_GETNUMBEROFWORKAREAS      uint32    = 0x1049
	LVM_GETSELECTIONMARK          uint32    = 0x1042
	LVM_SETSELECTIONMARK          uint32    = 0x1043
	LVM_SETHOVERTIME              uint32    = 0x1047
	LVM_GETHOVERTIME              uint32    = 0x1048
	LVM_SETTOOLTIPS               uint32    = 0x104a
	LVM_GETTOOLTIPS               uint32    = 0x104e
	LVM_SORTITEMSEX               uint32    = 0x1051
	LVBKIF_SOURCE_NONE            uint32    = 0x0
	LVBKIF_SOURCE_HBITMAP         uint32    = 0x1
	LVBKIF_SOURCE_URL             uint32    = 0x2
	LVBKIF_SOURCE_MASK            uint32    = 0x3
	LVBKIF_STYLE_NORMAL           uint32    = 0x0
	LVBKIF_STYLE_TILE             uint32    = 0x10
	LVBKIF_STYLE_MASK             uint32    = 0x10
	LVBKIF_FLAG_TILEOFFSET        uint32    = 0x100
	LVBKIF_TYPE_WATERMARK         uint32    = 0x10000000
	LVBKIF_FLAG_ALPHABLEND        uint32    = 0x20000000
	LVM_SETBKIMAGEA               uint32    = 0x1044
	LVM_SETBKIMAGEW               uint32    = 0x108a
	LVM_GETBKIMAGEA               uint32    = 0x1045
	LVM_GETBKIMAGEW               uint32    = 0x108b
	LVM_SETSELECTEDCOLUMN         uint32    = 0x108c
	LV_VIEW_ICON                  uint32    = 0x0
	LV_VIEW_DETAILS               uint32    = 0x1
	LV_VIEW_SMALLICON             uint32    = 0x2
	LV_VIEW_LIST                  uint32    = 0x3
	LV_VIEW_TILE                  uint32    = 0x4
	LV_VIEW_MAX                   uint32    = 0x4
	LVM_SETVIEW                   uint32    = 0x108e
	LVM_GETVIEW                   uint32    = 0x108f
	LVGF_ALIGN                    uint32    = 0x8
	LVGF_GROUPID                  uint32    = 0x10
	LVGF_SUBTITLE                 uint32    = 0x100
	LVGF_TASK                     uint32    = 0x200
	LVGF_DESCRIPTIONTOP           uint32    = 0x400
	LVGF_DESCRIPTIONBOTTOM        uint32    = 0x800
	LVGF_TITLEIMAGE               uint32    = 0x1000
	LVGF_EXTENDEDIMAGE            uint32    = 0x2000
	LVGF_ITEMS                    uint32    = 0x4000
	LVGF_SUBSET                   uint32    = 0x8000
	LVGF_SUBSETITEMS              uint32    = 0x10000
	LVGS_NORMAL                   uint32    = 0x0
	LVGS_COLLAPSED                uint32    = 0x1
	LVGS_HIDDEN                   uint32    = 0x2
	LVGS_NOHEADER                 uint32    = 0x4
	LVGS_COLLAPSIBLE              uint32    = 0x8
	LVGS_FOCUSED                  uint32    = 0x10
	LVGS_SELECTED                 uint32    = 0x20
	LVGS_SUBSETED                 uint32    = 0x40
	LVGS_SUBSETLINKFOCUSED        uint32    = 0x80
	LVGA_FOOTER_LEFT              uint32    = 0x8
	LVGA_FOOTER_CENTER            uint32    = 0x10
	LVGA_FOOTER_RIGHT             uint32    = 0x20
	LVM_INSERTGROUP               uint32    = 0x1091
	LVM_SETGROUPINFO              uint32    = 0x1093
	LVM_GETGROUPINFO              uint32    = 0x1095
	LVM_REMOVEGROUP               uint32    = 0x1096
	LVM_MOVEGROUP                 uint32    = 0x1097
	LVM_GETGROUPCOUNT             uint32    = 0x1098
	LVM_GETGROUPINFOBYINDEX       uint32    = 0x1099
	LVM_MOVEITEMTOGROUP           uint32    = 0x109a
	LVGGR_GROUP                   uint32    = 0x0
	LVGGR_HEADER                  uint32    = 0x1
	LVGGR_LABEL                   uint32    = 0x2
	LVGGR_SUBSETLINK              uint32    = 0x3
	LVM_GETGROUPRECT              uint32    = 0x1062
	LVGMF_NONE                    uint32    = 0x0
	LVGMF_BORDERSIZE              uint32    = 0x1
	LVGMF_BORDERCOLOR             uint32    = 0x2
	LVGMF_TEXTCOLOR               uint32    = 0x4
	LVM_SETGROUPMETRICS           uint32    = 0x109b
	LVM_GETGROUPMETRICS           uint32    = 0x109c
	LVM_ENABLEGROUPVIEW           uint32    = 0x109d
	LVM_SORTGROUPS                uint32    = 0x109e
	LVM_INSERTGROUPSORTED         uint32    = 0x109f
	LVM_REMOVEALLGROUPS           uint32    = 0x10a0
	LVM_HASGROUP                  uint32    = 0x10a1
	LVM_GETGROUPSTATE             uint32    = 0x105c
	LVM_GETFOCUSEDGROUP           uint32    = 0x105d
	LVTVIF_AUTOSIZE               uint32    = 0x0
	LVTVIF_FIXEDWIDTH             uint32    = 0x1
	LVTVIF_FIXEDHEIGHT            uint32    = 0x2
	LVTVIF_FIXEDSIZE              uint32    = 0x3
	LVTVIM_TILESIZE               uint32    = 0x1
	LVTVIM_COLUMNS                uint32    = 0x2
	LVTVIM_LABELMARGIN            uint32    = 0x4
	LVM_SETTILEVIEWINFO           uint32    = 0x10a2
	LVM_GETTILEVIEWINFO           uint32    = 0x10a3
	LVM_SETTILEINFO               uint32    = 0x10a4
	LVM_GETTILEINFO               uint32    = 0x10a5
	LVIM_AFTER                    uint32    = 0x1
	LVM_SETINSERTMARK             uint32    = 0x10a6
	LVM_GETINSERTMARK             uint32    = 0x10a7
	LVM_INSERTMARKHITTEST         uint32    = 0x10a8
	LVM_GETINSERTMARKRECT         uint32    = 0x10a9
	LVM_SETINSERTMARKCOLOR        uint32    = 0x10aa
	LVM_GETINSERTMARKCOLOR        uint32    = 0x10ab
	LVM_SETINFOTIP                uint32    = 0x10ad
	LVM_GETSELECTEDCOLUMN         uint32    = 0x10ae
	LVM_ISGROUPVIEWENABLED        uint32    = 0x10af
	LVM_GETOUTLINECOLOR           uint32    = 0x10b0
	LVM_SETOUTLINECOLOR           uint32    = 0x10b1
	LVM_CANCELEDITLABEL           uint32    = 0x10b3
	LVM_MAPINDEXTOID              uint32    = 0x10b4
	LVM_MAPIDTOINDEX              uint32    = 0x10b5
	LVM_ISITEMVISIBLE             uint32    = 0x10b6
	LVM_GETEMPTYTEXT              uint32    = 0x10cc
	LVM_GETFOOTERRECT             uint32    = 0x10cd
	LVFF_ITEMCOUNT                uint32    = 0x1
	LVM_GETFOOTERINFO             uint32    = 0x10ce
	LVM_GETFOOTERITEMRECT         uint32    = 0x10cf
	LVFIS_FOCUSED                 uint32    = 0x1
	LVM_GETFOOTERITEM             uint32    = 0x10d0
	LVM_GETITEMINDEXRECT          uint32    = 0x10d1
	LVM_SETITEMINDEXSTATE         uint32    = 0x10d2
	LVM_GETNEXTITEMINDEX          uint32    = 0x10d3
	LVM_SETBKIMAGE                uint32    = 0x108a
	LVM_GETBKIMAGE                uint32    = 0x108b
	LVKF_ALT                      uint32    = 0x1
	LVKF_CONTROL                  uint32    = 0x2
	LVKF_SHIFT                    uint32    = 0x4
	LVCDRF_NOSELECT               uint32    = 0x10000
	LVCDRF_NOGROUPFRAME           uint32    = 0x20000
	LVIF_DI_SETITEM               uint32    = 0x1000
	LVGIT_UNFOLDED                uint32    = 0x1
	LVNSCH_DEFAULT                int32     = -1
	LVNSCH_ERROR                  int32     = -2
	LVNSCH_IGNORE                 int32     = -3
	WC_TREEVIEWA                  string    = "SysTreeView32"
	WC_TREEVIEWW                  string    = "SysTreeView32"
	WC_TREEVIEW                   string    = "SysTreeView32"
	TVS_HASBUTTONS                uint32    = 0x1
	TVS_HASLINES                  uint32    = 0x2
	TVS_LINESATROOT               uint32    = 0x4
	TVS_EDITLABELS                uint32    = 0x8
	TVS_DISABLEDRAGDROP           uint32    = 0x10
	TVS_SHOWSELALWAYS             uint32    = 0x20
	TVS_RTLREADING                uint32    = 0x40
	TVS_NOTOOLTIPS                uint32    = 0x80
	TVS_CHECKBOXES                uint32    = 0x100
	TVS_TRACKSELECT               uint32    = 0x200
	TVS_SINGLEEXPAND              uint32    = 0x400
	TVS_INFOTIP                   uint32    = 0x800
	TVS_FULLROWSELECT             uint32    = 0x1000
	TVS_NOSCROLL                  uint32    = 0x2000
	TVS_NONEVENHEIGHT             uint32    = 0x4000
	TVS_NOHSCROLL                 uint32    = 0x8000
	TVS_EX_NOSINGLECOLLAPSE       uint32    = 0x1
	TVS_EX_MULTISELECT            uint32    = 0x2
	TVS_EX_DOUBLEBUFFER           uint32    = 0x4
	TVS_EX_NOINDENTSTATE          uint32    = 0x8
	TVS_EX_RICHTOOLTIP            uint32    = 0x10
	TVS_EX_AUTOHSCROLL            uint32    = 0x20
	TVS_EX_FADEINOUTEXPANDOS      uint32    = 0x40
	TVS_EX_PARTIALCHECKBOXES      uint32    = 0x80
	TVS_EX_EXCLUSIONCHECKBOXES    uint32    = 0x100
	TVS_EX_DIMMEDCHECKBOXES       uint32    = 0x200
	TVS_EX_DRAWIMAGEASYNC         uint32    = 0x400
	TVIS_SELECTED                 uint32    = 0x2
	TVIS_CUT                      uint32    = 0x4
	TVIS_DROPHILITED              uint32    = 0x8
	TVIS_BOLD                     uint32    = 0x10
	TVIS_EXPANDED                 uint32    = 0x20
	TVIS_EXPANDEDONCE             uint32    = 0x40
	TVIS_EXPANDPARTIAL            uint32    = 0x80
	TVIS_OVERLAYMASK              uint32    = 0xf00
	TVIS_STATEIMAGEMASK           uint32    = 0xf000
	TVIS_USERMASK                 uint32    = 0xf000
	TVIS_EX_FLAT                  uint32    = 0x1
	TVIS_EX_DISABLED              uint32    = 0x2
	TVIS_EX_ALL                   uint32    = 0x2
	TVM_INSERTITEMA               uint32    = 0x1100
	TVM_INSERTITEMW               uint32    = 0x1132
	TVM_INSERTITEM                uint32    = 0x1132
	TVM_DELETEITEM                uint32    = 0x1101
	TVM_EXPAND                    uint32    = 0x1102
	TVE_COLLAPSE                  uint32    = 0x1
	TVE_EXPAND                    uint32    = 0x2
	TVE_TOGGLE                    uint32    = 0x3
	TVE_EXPANDPARTIAL             uint32    = 0x4000
	TVE_COLLAPSERESET             uint32    = 0x8000
	TVM_GETITEMRECT               uint32    = 0x1104
	TVM_GETCOUNT                  uint32    = 0x1105
	TVM_GETINDENT                 uint32    = 0x1106
	TVM_SETINDENT                 uint32    = 0x1107
	TVM_GETIMAGELIST              uint32    = 0x1108
	TVSIL_NORMAL                  uint32    = 0x0
	TVSIL_STATE                   uint32    = 0x2
	TVM_SETIMAGELIST              uint32    = 0x1109
	TVM_GETNEXTITEM               uint32    = 0x110a
	TVGN_ROOT                     uint32    = 0x0
	TVGN_NEXT                     uint32    = 0x1
	TVGN_PREVIOUS                 uint32    = 0x2
	TVGN_PARENT                   uint32    = 0x3
	TVGN_CHILD                    uint32    = 0x4
	TVGN_FIRSTVISIBLE             uint32    = 0x5
	TVGN_NEXTVISIBLE              uint32    = 0x6
	TVGN_PREVIOUSVISIBLE          uint32    = 0x7
	TVGN_DROPHILITE               uint32    = 0x8
	TVGN_CARET                    uint32    = 0x9
	TVGN_LASTVISIBLE              uint32    = 0xa
	TVGN_NEXTSELECTED             uint32    = 0xb
	TVSI_NOSINGLEEXPAND           uint32    = 0x8000
	TVM_SELECTITEM                uint32    = 0x110b
	TVM_GETITEMA                  uint32    = 0x110c
	TVM_GETITEMW                  uint32    = 0x113e
	TVM_GETITEM                   uint32    = 0x113e
	TVM_SETITEMA                  uint32    = 0x110d
	TVM_SETITEMW                  uint32    = 0x113f
	TVM_SETITEM                   uint32    = 0x113f
	TVM_EDITLABELA                uint32    = 0x110e
	TVM_EDITLABELW                uint32    = 0x1141
	TVM_EDITLABEL                 uint32    = 0x1141
	TVM_GETEDITCONTROL            uint32    = 0x110f
	TVM_GETVISIBLECOUNT           uint32    = 0x1110
	TVM_HITTEST                   uint32    = 0x1111
	TVM_CREATEDRAGIMAGE           uint32    = 0x1112
	TVM_SORTCHILDREN              uint32    = 0x1113
	TVM_ENSUREVISIBLE             uint32    = 0x1114
	TVM_SORTCHILDRENCB            uint32    = 0x1115
	TVM_ENDEDITLABELNOW           uint32    = 0x1116
	TVM_GETISEARCHSTRINGA         uint32    = 0x1117
	TVM_GETISEARCHSTRINGW         uint32    = 0x1140
	TVM_GETISEARCHSTRING          uint32    = 0x1140
	TVM_SETTOOLTIPS               uint32    = 0x1118
	TVM_GETTOOLTIPS               uint32    = 0x1119
	TVM_SETINSERTMARK             uint32    = 0x111a
	TVM_SETUNICODEFORMAT          uint32    = 0x2005
	TVM_GETUNICODEFORMAT          uint32    = 0x2006
	TVM_SETITEMHEIGHT             uint32    = 0x111b
	TVM_GETITEMHEIGHT             uint32    = 0x111c
	TVM_SETBKCOLOR                uint32    = 0x111d
	TVM_SETTEXTCOLOR              uint32    = 0x111e
	TVM_GETBKCOLOR                uint32    = 0x111f
	TVM_GETTEXTCOLOR              uint32    = 0x1120
	TVM_SETSCROLLTIME             uint32    = 0x1121
	TVM_GETSCROLLTIME             uint32    = 0x1122
	TVM_SETINSERTMARKCOLOR        uint32    = 0x1125
	TVM_GETINSERTMARKCOLOR        uint32    = 0x1126
	TVM_SETBORDER                 uint32    = 0x1123
	TVSBF_XBORDER                 uint32    = 0x1
	TVSBF_YBORDER                 uint32    = 0x2
	TVM_GETITEMSTATE              uint32    = 0x1127
	TVM_SETLINECOLOR              uint32    = 0x1128
	TVM_GETLINECOLOR              uint32    = 0x1129
	TVM_MAPACCIDTOHTREEITEM       uint32    = 0x112a
	TVM_MAPHTREEITEMTOACCID       uint32    = 0x112b
	TVM_SETEXTENDEDSTYLE          uint32    = 0x112c
	TVM_GETEXTENDEDSTYLE          uint32    = 0x112d
	TVM_SETAUTOSCROLLINFO         uint32    = 0x113b
	TVM_SETHOT                    uint32    = 0x113a
	TVM_GETSELECTEDCOUNT          uint32    = 0x1146
	TVM_SHOWINFOTIP               uint32    = 0x1147
	TVM_GETITEMPARTRECT           uint32    = 0x1148
	TVC_UNKNOWN                   uint32    = 0x0
	TVC_BYMOUSE                   uint32    = 0x1
	TVC_BYKEYBOARD                uint32    = 0x2
	TVNRET_DEFAULT                uint32    = 0x0
	TVNRET_SKIPOLD                uint32    = 0x1
	TVNRET_SKIPNEW                uint32    = 0x2
	TVCDRF_NOIMAGES               uint32    = 0x10000
	WC_COMBOBOXEXW                string    = "ComboBoxEx32"
	WC_COMBOBOXEXA                string    = "ComboBoxEx32"
	WC_COMBOBOXEX                 string    = "ComboBoxEx32"
	CBEM_INSERTITEMA              uint32    = 0x401
	CBEM_SETIMAGELIST             uint32    = 0x402
	CBEM_GETIMAGELIST             uint32    = 0x403
	CBEM_GETITEMA                 uint32    = 0x404
	CBEM_SETITEMA                 uint32    = 0x405
	CBEM_GETCOMBOCONTROL          uint32    = 0x406
	CBEM_GETEDITCONTROL           uint32    = 0x407
	CBEM_SETEXSTYLE               uint32    = 0x408
	CBEM_SETEXTENDEDSTYLE         uint32    = 0x40e
	CBEM_GETEXSTYLE               uint32    = 0x409
	CBEM_GETEXTENDEDSTYLE         uint32    = 0x409
	CBEM_SETUNICODEFORMAT         uint32    = 0x2005
	CBEM_GETUNICODEFORMAT         uint32    = 0x2006
	CBEM_HASEDITCHANGED           uint32    = 0x40a
	CBEM_INSERTITEMW              uint32    = 0x40b
	CBEM_SETITEMW                 uint32    = 0x40c
	CBEM_GETITEMW                 uint32    = 0x40d
	CBEM_INSERTITEM               uint32    = 0x40b
	CBEM_SETITEM                  uint32    = 0x40c
	CBEM_GETITEM                  uint32    = 0x40d
	CBEM_SETWINDOWTHEME           uint32    = 0x200b
	CBES_EX_NOEDITIMAGE           uint32    = 0x1
	CBES_EX_NOEDITIMAGEINDENT     uint32    = 0x2
	CBES_EX_PATHWORDBREAKPROC     uint32    = 0x4
	CBES_EX_NOSIZELIMIT           uint32    = 0x8
	CBES_EX_CASESENSITIVE         uint32    = 0x10
	CBES_EX_TEXTENDELLIPSIS       uint32    = 0x20
	CBENF_KILLFOCUS               uint32    = 0x1
	CBENF_RETURN                  uint32    = 0x2
	CBENF_ESCAPE                  uint32    = 0x3
	CBENF_DROPDOWN                uint32    = 0x4
	CBEMAXSTRLEN                  uint32    = 0x104
	WC_TABCONTROLA                string    = "SysTabControl32"
	WC_TABCONTROLW                string    = "SysTabControl32"
	WC_TABCONTROL                 string    = "SysTabControl32"
	TCS_SCROLLOPPOSITE            uint32    = 0x1
	TCS_BOTTOM                    uint32    = 0x2
	TCS_RIGHT                     uint32    = 0x2
	TCS_MULTISELECT               uint32    = 0x4
	TCS_FLATBUTTONS               uint32    = 0x8
	TCS_FORCEICONLEFT             uint32    = 0x10
	TCS_FORCELABELLEFT            uint32    = 0x20
	TCS_HOTTRACK                  uint32    = 0x40
	TCS_VERTICAL                  uint32    = 0x80
	TCS_TABS                      uint32    = 0x0
	TCS_BUTTONS                   uint32    = 0x100
	TCS_SINGLELINE                uint32    = 0x0
	TCS_MULTILINE                 uint32    = 0x200
	TCS_RIGHTJUSTIFY              uint32    = 0x0
	TCS_FIXEDWIDTH                uint32    = 0x400
	TCS_RAGGEDRIGHT               uint32    = 0x800
	TCS_FOCUSONBUTTONDOWN         uint32    = 0x1000
	TCS_OWNERDRAWFIXED            uint32    = 0x2000
	TCS_TOOLTIPS                  uint32    = 0x4000
	TCS_FOCUSNEVER                uint32    = 0x8000
	TCS_EX_FLATSEPARATORS         uint32    = 0x1
	TCS_EX_REGISTERDROP           uint32    = 0x2
	TCM_GETIMAGELIST              uint32    = 0x1302
	TCM_SETIMAGELIST              uint32    = 0x1303
	TCM_GETITEMCOUNT              uint32    = 0x1304
	TCIS_BUTTONPRESSED            uint32    = 0x1
	TCIS_HIGHLIGHTED              uint32    = 0x2
	TCM_GETITEMA                  uint32    = 0x1305
	TCM_GETITEMW                  uint32    = 0x133c
	TCM_GETITEM                   uint32    = 0x133c
	TCM_SETITEMA                  uint32    = 0x1306
	TCM_SETITEMW                  uint32    = 0x133d
	TCM_SETITEM                   uint32    = 0x133d
	TCM_INSERTITEMA               uint32    = 0x1307
	TCM_INSERTITEMW               uint32    = 0x133e
	TCM_INSERTITEM                uint32    = 0x133e
	TCM_DELETEITEM                uint32    = 0x1308
	TCM_DELETEALLITEMS            uint32    = 0x1309
	TCM_GETITEMRECT               uint32    = 0x130a
	TCM_GETCURSEL                 uint32    = 0x130b
	TCM_SETCURSEL                 uint32    = 0x130c
	TCM_HITTEST                   uint32    = 0x130d
	TCM_SETITEMEXTRA              uint32    = 0x130e
	TCM_ADJUSTRECT                uint32    = 0x1328
	TCM_SETITEMSIZE               uint32    = 0x1329
	TCM_REMOVEIMAGE               uint32    = 0x132a
	TCM_SETPADDING                uint32    = 0x132b
	TCM_GETROWCOUNT               uint32    = 0x132c
	TCM_GETTOOLTIPS               uint32    = 0x132d
	TCM_SETTOOLTIPS               uint32    = 0x132e
	TCM_GETCURFOCUS               uint32    = 0x132f
	TCM_SETCURFOCUS               uint32    = 0x1330
	TCM_SETMINTABWIDTH            uint32    = 0x1331
	TCM_DESELECTALL               uint32    = 0x1332
	TCM_HIGHLIGHTITEM             uint32    = 0x1333
	TCM_SETEXTENDEDSTYLE          uint32    = 0x1334
	TCM_GETEXTENDEDSTYLE          uint32    = 0x1335
	TCM_SETUNICODEFORMAT          uint32    = 0x2005
	TCM_GETUNICODEFORMAT          uint32    = 0x2006
	ANIMATE_CLASSW                string    = "SysAnimate32"
	ANIMATE_CLASSA                string    = "SysAnimate32"
	ANIMATE_CLASS                 string    = "SysAnimate32"
	ACS_CENTER                    uint32    = 0x1
	ACS_TRANSPARENT               uint32    = 0x2
	ACS_AUTOPLAY                  uint32    = 0x4
	ACS_TIMER                     uint32    = 0x8
	ACM_OPENA                     uint32    = 0x464
	ACM_OPENW                     uint32    = 0x467
	ACM_OPEN                      uint32    = 0x467
	ACM_PLAY                      uint32    = 0x465
	ACM_STOP                      uint32    = 0x466
	ACM_ISPLAYING                 uint32    = 0x468
	ACN_START                     uint32    = 0x1
	ACN_STOP                      uint32    = 0x2
	MONTHCAL_CLASSW               string    = "SysMonthCal32"
	MONTHCAL_CLASSA               string    = "SysMonthCal32"
	MONTHCAL_CLASS                string    = "SysMonthCal32"
	MCM_FIRST                     uint32    = 0x1000
	MCM_GETCURSEL                 uint32    = 0x1001
	MCM_SETCURSEL                 uint32    = 0x1002
	MCM_GETMAXSELCOUNT            uint32    = 0x1003
	MCM_SETMAXSELCOUNT            uint32    = 0x1004
	MCM_GETSELRANGE               uint32    = 0x1005
	MCM_SETSELRANGE               uint32    = 0x1006
	MCM_GETMONTHRANGE             uint32    = 0x1007
	MCM_SETDAYSTATE               uint32    = 0x1008
	MCM_GETMINREQRECT             uint32    = 0x1009
	MCM_SETCOLOR                  uint32    = 0x100a
	MCM_GETCOLOR                  uint32    = 0x100b
	MCSC_BACKGROUND               uint32    = 0x0
	MCSC_TEXT                     uint32    = 0x1
	MCSC_TITLEBK                  uint32    = 0x2
	MCSC_TITLETEXT                uint32    = 0x3
	MCSC_MONTHBK                  uint32    = 0x4
	MCSC_TRAILINGTEXT             uint32    = 0x5
	MCM_SETTODAY                  uint32    = 0x100c
	MCM_GETTODAY                  uint32    = 0x100d
	MCM_HITTEST                   uint32    = 0x100e
	MCHT_TITLE                    uint32    = 0x10000
	MCHT_CALENDAR                 uint32    = 0x20000
	MCHT_TODAYLINK                uint32    = 0x30000
	MCHT_CALENDARCONTROL          uint32    = 0x100000
	MCHT_NEXT                     uint32    = 0x1000000
	MCHT_PREV                     uint32    = 0x2000000
	MCHT_NOWHERE                  uint32    = 0x0
	MCHT_TITLEBK                  uint32    = 0x10000
	MCHT_CALENDARBK               uint32    = 0x20000
	MCM_SETFIRSTDAYOFWEEK         uint32    = 0x100f
	MCM_GETFIRSTDAYOFWEEK         uint32    = 0x1010
	MCM_GETRANGE                  uint32    = 0x1011
	MCM_SETRANGE                  uint32    = 0x1012
	MCM_GETMONTHDELTA             uint32    = 0x1013
	MCM_SETMONTHDELTA             uint32    = 0x1014
	MCM_GETMAXTODAYWIDTH          uint32    = 0x1015
	MCM_SETUNICODEFORMAT          uint32    = 0x2005
	MCM_GETUNICODEFORMAT          uint32    = 0x2006
	MCMV_MONTH                    uint32    = 0x0
	MCMV_YEAR                     uint32    = 0x1
	MCMV_DECADE                   uint32    = 0x2
	MCMV_CENTURY                  uint32    = 0x3
	MCMV_MAX                      uint32    = 0x3
	MCM_GETCURRENTVIEW            uint32    = 0x1016
	MCM_GETCALENDARCOUNT          uint32    = 0x1017
	MCM_GETCALENDARGRIDINFO       uint32    = 0x1018
	MCM_GETCALID                  uint32    = 0x101b
	MCM_SETCALID                  uint32    = 0x101c
	MCM_SIZERECTTOMIN             uint32    = 0x101d
	MCM_SETCALENDARBORDER         uint32    = 0x101e
	MCM_GETCALENDARBORDER         uint32    = 0x101f
	MCM_SETCURRENTVIEW            uint32    = 0x1020
	MCS_DAYSTATE                  uint32    = 0x1
	MCS_MULTISELECT               uint32    = 0x2
	MCS_WEEKNUMBERS               uint32    = 0x4
	MCS_NOTODAYCIRCLE             uint32    = 0x8
	MCS_NOTODAY                   uint32    = 0x10
	MCS_NOTRAILINGDATES           uint32    = 0x40
	MCS_SHORTDAYSOFWEEK           uint32    = 0x80
	MCS_NOSELCHANGEONNAV          uint32    = 0x100
	GMR_VISIBLE                   uint32    = 0x0
	GMR_DAYSTATE                  uint32    = 0x1
	DATETIMEPICK_CLASSW           string    = "SysDateTimePick32"
	DATETIMEPICK_CLASSA           string    = "SysDateTimePick32"
	DATETIMEPICK_CLASS            string    = "SysDateTimePick32"
	DTM_FIRST                     uint32    = 0x1000
	DTM_GETSYSTEMTIME             uint32    = 0x1001
	DTM_SETSYSTEMTIME             uint32    = 0x1002
	DTM_GETRANGE                  uint32    = 0x1003
	DTM_SETRANGE                  uint32    = 0x1004
	DTM_SETFORMATA                uint32    = 0x1005
	DTM_SETFORMATW                uint32    = 0x1032
	DTM_SETFORMAT                 uint32    = 0x1032
	DTM_SETMCCOLOR                uint32    = 0x1006
	DTM_GETMCCOLOR                uint32    = 0x1007
	DTM_GETMONTHCAL               uint32    = 0x1008
	DTM_SETMCFONT                 uint32    = 0x1009
	DTM_GETMCFONT                 uint32    = 0x100a
	DTM_SETMCSTYLE                uint32    = 0x100b
	DTM_GETMCSTYLE                uint32    = 0x100c
	DTM_CLOSEMONTHCAL             uint32    = 0x100d
	DTM_GETDATETIMEPICKERINFO     uint32    = 0x100e
	DTM_GETIDEALSIZE              uint32    = 0x100f
	DTS_UPDOWN                    uint32    = 0x1
	DTS_SHOWNONE                  uint32    = 0x2
	DTS_SHORTDATEFORMAT           uint32    = 0x0
	DTS_LONGDATEFORMAT            uint32    = 0x4
	DTS_SHORTDATECENTURYFORMAT    uint32    = 0xc
	DTS_TIMEFORMAT                uint32    = 0x9
	DTS_APPCANPARSE               uint32    = 0x10
	DTS_RIGHTALIGN                uint32    = 0x20
	GDTR_MIN                      uint32    = 0x1
	GDTR_MAX                      uint32    = 0x2
	GDT_ERROR                     int32     = -1
	GDT_VALID                     uint32    = 0x0
	GDT_NONE                      uint32    = 0x1
	IPM_CLEARADDRESS              uint32    = 0x464
	IPM_SETADDRESS                uint32    = 0x465
	IPM_GETADDRESS                uint32    = 0x466
	IPM_SETRANGE                  uint32    = 0x467
	IPM_SETFOCUS                  uint32    = 0x468
	IPM_ISBLANK                   uint32    = 0x469
	WC_IPADDRESSW                 string    = "SysIPAddress32"
	WC_IPADDRESSA                 string    = "SysIPAddress32"
	WC_IPADDRESS                  string    = "SysIPAddress32"
	WC_PAGESCROLLERW              string    = "SysPager"
	WC_PAGESCROLLERA              string    = "SysPager"
	WC_PAGESCROLLER               string    = "SysPager"
	PGS_VERT                      uint32    = 0x0
	PGS_HORZ                      uint32    = 0x1
	PGS_AUTOSCROLL                uint32    = 0x2
	PGS_DRAGNDROP                 uint32    = 0x4
	PGF_INVISIBLE                 uint32    = 0x0
	PGF_NORMAL                    uint32    = 0x1
	PGF_GRAYED                    uint32    = 0x2
	PGF_DEPRESSED                 uint32    = 0x4
	PGF_HOT                       uint32    = 0x8
	PGB_TOPORLEFT                 uint32    = 0x0
	PGB_BOTTOMORRIGHT             uint32    = 0x1
	PGM_SETCHILD                  uint32    = 0x1401
	PGM_RECALCSIZE                uint32    = 0x1402
	PGM_FORWARDMOUSE              uint32    = 0x1403
	PGM_SETBKCOLOR                uint32    = 0x1404
	PGM_GETBKCOLOR                uint32    = 0x1405
	PGM_SETBORDER                 uint32    = 0x1406
	PGM_GETBORDER                 uint32    = 0x1407
	PGM_SETPOS                    uint32    = 0x1408
	PGM_GETPOS                    uint32    = 0x1409
	PGM_SETBUTTONSIZE             uint32    = 0x140a
	PGM_GETBUTTONSIZE             uint32    = 0x140b
	PGM_GETBUTTONSTATE            uint32    = 0x140c
	PGM_GETDROPTARGET             uint32    = 0x2004
	PGM_SETSCROLLINFO             uint32    = 0x140d
	WC_NATIVEFONTCTLW             string    = "NativeFontCtl"
	WC_NATIVEFONTCTLA             string    = "NativeFontCtl"
	WC_NATIVEFONTCTL              string    = "NativeFontCtl"
	NFS_EDIT                      uint32    = 0x1
	NFS_STATIC                    uint32    = 0x2
	NFS_LISTCOMBO                 uint32    = 0x4
	NFS_BUTTON                    uint32    = 0x8
	NFS_ALL                       uint32    = 0x10
	NFS_USEFONTASSOC              uint32    = 0x20
	WC_BUTTONA                    string    = "Button"
	WC_BUTTONW                    string    = "Button"
	WC_BUTTON                     string    = "Button"
	BCM_GETIDEALSIZE              uint32    = 0x1601
	BCM_SETIMAGELIST              uint32    = 0x1602
	BCM_GETIMAGELIST              uint32    = 0x1603
	BCM_SETTEXTMARGIN             uint32    = 0x1604
	BCM_GETTEXTMARGIN             uint32    = 0x1605
	BCN_HOTITEMCHANGE             uint32    = 0xfffffb1f
	BST_HOT                       uint32    = 0x200
	BST_DROPDOWNPUSHED            uint32    = 0x400
	BS_SPLITBUTTON                int32     = 12
	BS_DEFSPLITBUTTON             int32     = 13
	BS_COMMANDLINK                int32     = 14
	BS_DEFCOMMANDLINK             int32     = 15
	BCSIF_GLYPH                   uint32    = 0x1
	BCSIF_IMAGE                   uint32    = 0x2
	BCSIF_STYLE                   uint32    = 0x4
	BCSIF_SIZE                    uint32    = 0x8
	BCSS_NOSPLIT                  uint32    = 0x1
	BCSS_STRETCH                  uint32    = 0x2
	BCSS_ALIGNLEFT                uint32    = 0x4
	BCSS_IMAGE                    uint32    = 0x8
	BCM_SETDROPDOWNSTATE          uint32    = 0x1606
	BCM_SETSPLITINFO              uint32    = 0x1607
	BCM_GETSPLITINFO              uint32    = 0x1608
	BCM_SETNOTE                   uint32    = 0x1609
	BCM_GETNOTE                   uint32    = 0x160a
	BCM_GETNOTELENGTH             uint32    = 0x160b
	BCM_SETSHIELD                 uint32    = 0x160c
	BCN_DROPDOWN                  uint32    = 0xfffffb20
	WC_STATICA                    string    = "Static"
	WC_STATICW                    string    = "Static"
	WC_STATIC                     string    = "Static"
	WC_EDITA                      string    = "Edit"
	WC_EDITW                      string    = "Edit"
	WC_EDIT                       string    = "Edit"
	ES_EX_ALLOWEOL_CR             int32     = 1
	ES_EX_ALLOWEOL_LF             int32     = 2
	ES_EX_CONVERT_EOL_ON_PASTE    int32     = 4
	ES_EX_ZOOMABLE                int32     = 16
	EM_SETCUEBANNER               uint32    = 0x1501
	EM_GETCUEBANNER               uint32    = 0x1502
	EM_SHOWBALLOONTIP             uint32    = 0x1503
	EM_HIDEBALLOONTIP             uint32    = 0x1504
	EM_SETHILITE                  uint32    = 0x1505
	EM_GETHILITE                  uint32    = 0x1506
	EM_NOSETFOCUS                 uint32    = 0x1507
	EM_TAKEFOCUS                  uint32    = 0x1508
	EM_SETEXTENDEDSTYLE           uint32    = 0x150a
	EM_GETEXTENDEDSTYLE           uint32    = 0x150b
	EM_SETENDOFLINE               uint32    = 0x150c
	EM_GETENDOFLINE               uint32    = 0x150d
	EM_ENABLESEARCHWEB            uint32    = 0x150e
	EM_SEARCHWEB                  uint32    = 0x150f
	EM_SETCARETINDEX              uint32    = 0x1511
	EM_GETCARETINDEX              uint32    = 0x1512
	EM_FILELINEFROMCHAR           uint32    = 0x1513
	EM_FILELINEINDEX              uint32    = 0x1514
	EM_FILELINELENGTH             uint32    = 0x1515
	EM_GETFILELINE                uint32    = 0x1516
	EM_GETFILELINECOUNT           uint32    = 0x1517
	WC_LISTBOXA                   string    = "ListBox"
	WC_LISTBOXW                   string    = "ListBox"
	WC_LISTBOX                    string    = "ListBox"
	WC_COMBOBOXA                  string    = "ComboBox"
	WC_COMBOBOXW                  string    = "ComboBox"
	WC_COMBOBOX                   string    = "ComboBox"
	CB_SETMINVISIBLE              uint32    = 0x1701
	CB_GETMINVISIBLE              uint32    = 0x1702
	CB_SETCUEBANNER               uint32    = 0x1703
	CB_GETCUEBANNER               uint32    = 0x1704
	WC_SCROLLBARA                 string    = "ScrollBar"
	WC_SCROLLBARW                 string    = "ScrollBar"
	WC_SCROLLBAR                  string    = "ScrollBar"
	WM_MOUSEHOVER                 uint32    = 0x2a1
	WM_MOUSELEAVE                 uint32    = 0x2a3
	HOVER_DEFAULT                 uint32    = 0xffffffff
	WSB_PROP_MASK                 int32     = 4095
	FSB_FLAT_MODE                 uint32    = 0x2
	FSB_ENCARTA_MODE              uint32    = 0x1
	FSB_REGULAR_MODE              uint32    = 0x0
	VSCLASS_LINK                  string    = "LINK"
	VSCLASS_EMPTYMARKUP           string    = "EMPTYMARKUP"
	VSCLASS_STATIC                string    = "STATIC"
	VSCLASS_PAGE                  string    = "PAGE"
	VSCLASS_MONTHCAL              string    = "MONTHCAL"
	VSCLASS_CLOCK                 string    = "CLOCK"
	VSCLASS_TRAYNOTIFY            string    = "TRAYNOTIFY"
	VSCLASS_TASKBAR               string    = "TASKBAR"
	VSCLASS_TASKBAND              string    = "TASKBAND"
	VSCLASS_STARTPANEL            string    = "STARTPANEL"
	VSCLASS_MENUBAND              string    = "MENUBAND"
	CCHCCCLASS                    uint32    = 0x20
	CCHCCDESC                     uint32    = 0x20
	CCHCCTEXT                     uint32    = 0x100
	CCF_NOTEXT                    uint32    = 0x1
	EM_GETSEL                     uint32    = 0xb0
	EM_SETSEL                     uint32    = 0xb1
	EM_GETRECT                    uint32    = 0xb2
	EM_SETRECT                    uint32    = 0xb3
	EM_SETRECTNP                  uint32    = 0xb4
	EM_SCROLL                     uint32    = 0xb5
	EM_LINESCROLL                 uint32    = 0xb6
	EM_GETMODIFY                  uint32    = 0xb8
	EM_SETMODIFY                  uint32    = 0xb9
	EM_GETLINECOUNT               uint32    = 0xba
	EM_LINEINDEX                  uint32    = 0xbb
	EM_SETHANDLE                  uint32    = 0xbc
	EM_GETHANDLE                  uint32    = 0xbd
	EM_GETTHUMB                   uint32    = 0xbe
	EM_LINELENGTH                 uint32    = 0xc1
	EM_REPLACESEL                 uint32    = 0xc2
	EM_GETLINE                    uint32    = 0xc4
	EM_LIMITTEXT                  uint32    = 0xc5
	EM_CANUNDO                    uint32    = 0xc6
	EM_UNDO                       uint32    = 0xc7
	EM_FMTLINES                   uint32    = 0xc8
	EM_LINEFROMCHAR               uint32    = 0xc9
	EM_SETTABSTOPS                uint32    = 0xcb
	EM_SETPASSWORDCHAR            uint32    = 0xcc
	EM_EMPTYUNDOBUFFER            uint32    = 0xcd
	EM_GETFIRSTVISIBLELINE        uint32    = 0xce
	EM_SETREADONLY                uint32    = 0xcf
	EM_SETWORDBREAKPROC           uint32    = 0xd0
	EM_GETWORDBREAKPROC           uint32    = 0xd1
	EM_GETPASSWORDCHAR            uint32    = 0xd2
	EM_SETMARGINS                 uint32    = 0xd3
	EM_GETMARGINS                 uint32    = 0xd4
	EM_SETIMESTATUS               uint32    = 0xd8
	EM_GETIMESTATUS               uint32    = 0xd9
	EM_ENABLEFEATURE              uint32    = 0xda
)

// enums

// enum
type THEME_PROPERTY_SYMBOL_ID uint32

const (
	TMT_RESERVEDLOW             THEME_PROPERTY_SYMBOL_ID = 0
	TMT_RESERVEDHIGH            THEME_PROPERTY_SYMBOL_ID = 7999
	TMT_DIBDATA                 THEME_PROPERTY_SYMBOL_ID = 2
	TMT_GLYPHDIBDATA            THEME_PROPERTY_SYMBOL_ID = 8
	TMT_ENUM                    THEME_PROPERTY_SYMBOL_ID = 200
	TMT_STRING                  THEME_PROPERTY_SYMBOL_ID = 201
	TMT_INT                     THEME_PROPERTY_SYMBOL_ID = 202
	TMT_BOOL                    THEME_PROPERTY_SYMBOL_ID = 203
	TMT_COLOR                   THEME_PROPERTY_SYMBOL_ID = 204
	TMT_MARGINS                 THEME_PROPERTY_SYMBOL_ID = 205
	TMT_FILENAME                THEME_PROPERTY_SYMBOL_ID = 206
	TMT_SIZE                    THEME_PROPERTY_SYMBOL_ID = 207
	TMT_POSITION                THEME_PROPERTY_SYMBOL_ID = 208
	TMT_RECT                    THEME_PROPERTY_SYMBOL_ID = 209
	TMT_FONT                    THEME_PROPERTY_SYMBOL_ID = 210
	TMT_INTLIST                 THEME_PROPERTY_SYMBOL_ID = 211
	TMT_HBITMAP                 THEME_PROPERTY_SYMBOL_ID = 212
	TMT_DISKSTREAM              THEME_PROPERTY_SYMBOL_ID = 213
	TMT_STREAM                  THEME_PROPERTY_SYMBOL_ID = 214
	TMT_BITMAPREF               THEME_PROPERTY_SYMBOL_ID = 215
	TMT_FLOAT                   THEME_PROPERTY_SYMBOL_ID = 216
	TMT_FLOATLIST               THEME_PROPERTY_SYMBOL_ID = 217
	TMT_COLORSCHEMES            THEME_PROPERTY_SYMBOL_ID = 401
	TMT_SIZES                   THEME_PROPERTY_SYMBOL_ID = 402
	TMT_CHARSET                 THEME_PROPERTY_SYMBOL_ID = 403
	TMT_NAME                    THEME_PROPERTY_SYMBOL_ID = 600
	TMT_DISPLAYNAME             THEME_PROPERTY_SYMBOL_ID = 601
	TMT_TOOLTIP                 THEME_PROPERTY_SYMBOL_ID = 602
	TMT_COMPANY                 THEME_PROPERTY_SYMBOL_ID = 603
	TMT_AUTHOR                  THEME_PROPERTY_SYMBOL_ID = 604
	TMT_COPYRIGHT               THEME_PROPERTY_SYMBOL_ID = 605
	TMT_URL                     THEME_PROPERTY_SYMBOL_ID = 606
	TMT_VERSION                 THEME_PROPERTY_SYMBOL_ID = 607
	TMT_DESCRIPTION             THEME_PROPERTY_SYMBOL_ID = 608
	TMT_FIRST_RCSTRING_NAME     THEME_PROPERTY_SYMBOL_ID = 601
	TMT_LAST_RCSTRING_NAME      THEME_PROPERTY_SYMBOL_ID = 608
	TMT_CAPTIONFONT             THEME_PROPERTY_SYMBOL_ID = 801
	TMT_SMALLCAPTIONFONT        THEME_PROPERTY_SYMBOL_ID = 802
	TMT_MENUFONT                THEME_PROPERTY_SYMBOL_ID = 803
	TMT_STATUSFONT              THEME_PROPERTY_SYMBOL_ID = 804
	TMT_MSGBOXFONT              THEME_PROPERTY_SYMBOL_ID = 805
	TMT_ICONTITLEFONT           THEME_PROPERTY_SYMBOL_ID = 806
	TMT_HEADING1FONT            THEME_PROPERTY_SYMBOL_ID = 807
	TMT_HEADING2FONT            THEME_PROPERTY_SYMBOL_ID = 808
	TMT_BODYFONT                THEME_PROPERTY_SYMBOL_ID = 809
	TMT_FIRSTFONT               THEME_PROPERTY_SYMBOL_ID = 801
	TMT_LASTFONT                THEME_PROPERTY_SYMBOL_ID = 809
	TMT_FLATMENUS               THEME_PROPERTY_SYMBOL_ID = 1001
	TMT_FIRSTBOOL               THEME_PROPERTY_SYMBOL_ID = 1001
	TMT_LASTBOOL                THEME_PROPERTY_SYMBOL_ID = 1001
	TMT_SIZINGBORDERWIDTH       THEME_PROPERTY_SYMBOL_ID = 1201
	TMT_SCROLLBARWIDTH          THEME_PROPERTY_SYMBOL_ID = 1202
	TMT_SCROLLBARHEIGHT         THEME_PROPERTY_SYMBOL_ID = 1203
	TMT_CAPTIONBARWIDTH         THEME_PROPERTY_SYMBOL_ID = 1204
	TMT_CAPTIONBARHEIGHT        THEME_PROPERTY_SYMBOL_ID = 1205
	TMT_SMCAPTIONBARWIDTH       THEME_PROPERTY_SYMBOL_ID = 1206
	TMT_SMCAPTIONBARHEIGHT      THEME_PROPERTY_SYMBOL_ID = 1207
	TMT_MENUBARWIDTH            THEME_PROPERTY_SYMBOL_ID = 1208
	TMT_MENUBARHEIGHT           THEME_PROPERTY_SYMBOL_ID = 1209
	TMT_PADDEDBORDERWIDTH       THEME_PROPERTY_SYMBOL_ID = 1210
	TMT_FIRSTSIZE               THEME_PROPERTY_SYMBOL_ID = 1201
	TMT_LASTSIZE                THEME_PROPERTY_SYMBOL_ID = 1210
	TMT_MINCOLORDEPTH           THEME_PROPERTY_SYMBOL_ID = 1301
	TMT_FIRSTINT                THEME_PROPERTY_SYMBOL_ID = 1301
	TMT_LASTINT                 THEME_PROPERTY_SYMBOL_ID = 1301
	TMT_CSSNAME                 THEME_PROPERTY_SYMBOL_ID = 1401
	TMT_XMLNAME                 THEME_PROPERTY_SYMBOL_ID = 1402
	TMT_LASTUPDATED             THEME_PROPERTY_SYMBOL_ID = 1403
	TMT_ALIAS                   THEME_PROPERTY_SYMBOL_ID = 1404
	TMT_FIRSTSTRING             THEME_PROPERTY_SYMBOL_ID = 1401
	TMT_LASTSTRING              THEME_PROPERTY_SYMBOL_ID = 1404
	TMT_SCROLLBAR               THEME_PROPERTY_SYMBOL_ID = 1601
	TMT_BACKGROUND              THEME_PROPERTY_SYMBOL_ID = 1602
	TMT_ACTIVECAPTION           THEME_PROPERTY_SYMBOL_ID = 1603
	TMT_INACTIVECAPTION         THEME_PROPERTY_SYMBOL_ID = 1604
	TMT_MENU                    THEME_PROPERTY_SYMBOL_ID = 1605
	TMT_WINDOW                  THEME_PROPERTY_SYMBOL_ID = 1606
	TMT_WINDOWFRAME             THEME_PROPERTY_SYMBOL_ID = 1607
	TMT_MENUTEXT                THEME_PROPERTY_SYMBOL_ID = 1608
	TMT_WINDOWTEXT              THEME_PROPERTY_SYMBOL_ID = 1609
	TMT_CAPTIONTEXT             THEME_PROPERTY_SYMBOL_ID = 1610
	TMT_ACTIVEBORDER            THEME_PROPERTY_SYMBOL_ID = 1611
	TMT_INACTIVEBORDER          THEME_PROPERTY_SYMBOL_ID = 1612
	TMT_APPWORKSPACE            THEME_PROPERTY_SYMBOL_ID = 1613
	TMT_HIGHLIGHT               THEME_PROPERTY_SYMBOL_ID = 1614
	TMT_HIGHLIGHTTEXT           THEME_PROPERTY_SYMBOL_ID = 1615
	TMT_BTNFACE                 THEME_PROPERTY_SYMBOL_ID = 1616
	TMT_BTNSHADOW               THEME_PROPERTY_SYMBOL_ID = 1617
	TMT_GRAYTEXT                THEME_PROPERTY_SYMBOL_ID = 1618
	TMT_BTNTEXT                 THEME_PROPERTY_SYMBOL_ID = 1619
	TMT_INACTIVECAPTIONTEXT     THEME_PROPERTY_SYMBOL_ID = 1620
	TMT_BTNHIGHLIGHT            THEME_PROPERTY_SYMBOL_ID = 1621
	TMT_DKSHADOW3D              THEME_PROPERTY_SYMBOL_ID = 1622
	TMT_LIGHT3D                 THEME_PROPERTY_SYMBOL_ID = 1623
	TMT_INFOTEXT                THEME_PROPERTY_SYMBOL_ID = 1624
	TMT_INFOBK                  THEME_PROPERTY_SYMBOL_ID = 1625
	TMT_BUTTONALTERNATEFACE     THEME_PROPERTY_SYMBOL_ID = 1626
	TMT_HOTTRACKING             THEME_PROPERTY_SYMBOL_ID = 1627
	TMT_GRADIENTACTIVECAPTION   THEME_PROPERTY_SYMBOL_ID = 1628
	TMT_GRADIENTINACTIVECAPTION THEME_PROPERTY_SYMBOL_ID = 1629
	TMT_MENUHILIGHT             THEME_PROPERTY_SYMBOL_ID = 1630
	TMT_MENUBAR                 THEME_PROPERTY_SYMBOL_ID = 1631
	TMT_FIRSTCOLOR              THEME_PROPERTY_SYMBOL_ID = 1601
	TMT_LASTCOLOR               THEME_PROPERTY_SYMBOL_ID = 1631
	TMT_FROMHUE1                THEME_PROPERTY_SYMBOL_ID = 1801
	TMT_FROMHUE2                THEME_PROPERTY_SYMBOL_ID = 1802
	TMT_FROMHUE3                THEME_PROPERTY_SYMBOL_ID = 1803
	TMT_FROMHUE4                THEME_PROPERTY_SYMBOL_ID = 1804
	TMT_FROMHUE5                THEME_PROPERTY_SYMBOL_ID = 1805
	TMT_TOHUE1                  THEME_PROPERTY_SYMBOL_ID = 1806
	TMT_TOHUE2                  THEME_PROPERTY_SYMBOL_ID = 1807
	TMT_TOHUE3                  THEME_PROPERTY_SYMBOL_ID = 1808
	TMT_TOHUE4                  THEME_PROPERTY_SYMBOL_ID = 1809
	TMT_TOHUE5                  THEME_PROPERTY_SYMBOL_ID = 1810
	TMT_FROMCOLOR1              THEME_PROPERTY_SYMBOL_ID = 2001
	TMT_FROMCOLOR2              THEME_PROPERTY_SYMBOL_ID = 2002
	TMT_FROMCOLOR3              THEME_PROPERTY_SYMBOL_ID = 2003
	TMT_FROMCOLOR4              THEME_PROPERTY_SYMBOL_ID = 2004
	TMT_FROMCOLOR5              THEME_PROPERTY_SYMBOL_ID = 2005
	TMT_TOCOLOR1                THEME_PROPERTY_SYMBOL_ID = 2006
	TMT_TOCOLOR2                THEME_PROPERTY_SYMBOL_ID = 2007
	TMT_TOCOLOR3                THEME_PROPERTY_SYMBOL_ID = 2008
	TMT_TOCOLOR4                THEME_PROPERTY_SYMBOL_ID = 2009
	TMT_TOCOLOR5                THEME_PROPERTY_SYMBOL_ID = 2010
	TMT_TRANSPARENT             THEME_PROPERTY_SYMBOL_ID = 2201
	TMT_AUTOSIZE                THEME_PROPERTY_SYMBOL_ID = 2202
	TMT_BORDERONLY              THEME_PROPERTY_SYMBOL_ID = 2203
	TMT_COMPOSITED              THEME_PROPERTY_SYMBOL_ID = 2204
	TMT_BGFILL                  THEME_PROPERTY_SYMBOL_ID = 2205
	TMT_GLYPHTRANSPARENT        THEME_PROPERTY_SYMBOL_ID = 2206
	TMT_GLYPHONLY               THEME_PROPERTY_SYMBOL_ID = 2207
	TMT_ALWAYSSHOWSIZINGBAR     THEME_PROPERTY_SYMBOL_ID = 2208
	TMT_MIRRORIMAGE             THEME_PROPERTY_SYMBOL_ID = 2209
	TMT_UNIFORMSIZING           THEME_PROPERTY_SYMBOL_ID = 2210
	TMT_INTEGRALSIZING          THEME_PROPERTY_SYMBOL_ID = 2211
	TMT_SOURCEGROW              THEME_PROPERTY_SYMBOL_ID = 2212
	TMT_SOURCESHRINK            THEME_PROPERTY_SYMBOL_ID = 2213
	TMT_DRAWBORDERS             THEME_PROPERTY_SYMBOL_ID = 2214
	TMT_NOETCHEDEFFECT          THEME_PROPERTY_SYMBOL_ID = 2215
	TMT_TEXTAPPLYOVERLAY        THEME_PROPERTY_SYMBOL_ID = 2216
	TMT_TEXTGLOW                THEME_PROPERTY_SYMBOL_ID = 2217
	TMT_TEXTITALIC              THEME_PROPERTY_SYMBOL_ID = 2218
	TMT_COMPOSITEDOPAQUE        THEME_PROPERTY_SYMBOL_ID = 2219
	TMT_LOCALIZEDMIRRORIMAGE    THEME_PROPERTY_SYMBOL_ID = 2220
	TMT_IMAGECOUNT              THEME_PROPERTY_SYMBOL_ID = 2401
	TMT_ALPHALEVEL              THEME_PROPERTY_SYMBOL_ID = 2402
	TMT_BORDERSIZE              THEME_PROPERTY_SYMBOL_ID = 2403
	TMT_ROUNDCORNERWIDTH        THEME_PROPERTY_SYMBOL_ID = 2404
	TMT_ROUNDCORNERHEIGHT       THEME_PROPERTY_SYMBOL_ID = 2405
	TMT_GRADIENTRATIO1          THEME_PROPERTY_SYMBOL_ID = 2406
	TMT_GRADIENTRATIO2          THEME_PROPERTY_SYMBOL_ID = 2407
	TMT_GRADIENTRATIO3          THEME_PROPERTY_SYMBOL_ID = 2408
	TMT_GRADIENTRATIO4          THEME_PROPERTY_SYMBOL_ID = 2409
	TMT_GRADIENTRATIO5          THEME_PROPERTY_SYMBOL_ID = 2410
	TMT_PROGRESSCHUNKSIZE       THEME_PROPERTY_SYMBOL_ID = 2411
	TMT_PROGRESSSPACESIZE       THEME_PROPERTY_SYMBOL_ID = 2412
	TMT_SATURATION              THEME_PROPERTY_SYMBOL_ID = 2413
	TMT_TEXTBORDERSIZE          THEME_PROPERTY_SYMBOL_ID = 2414
	TMT_ALPHATHRESHOLD          THEME_PROPERTY_SYMBOL_ID = 2415
	TMT_WIDTH                   THEME_PROPERTY_SYMBOL_ID = 2416
	TMT_HEIGHT                  THEME_PROPERTY_SYMBOL_ID = 2417
	TMT_GLYPHINDEX              THEME_PROPERTY_SYMBOL_ID = 2418
	TMT_TRUESIZESTRETCHMARK     THEME_PROPERTY_SYMBOL_ID = 2419
	TMT_MINDPI1                 THEME_PROPERTY_SYMBOL_ID = 2420
	TMT_MINDPI2                 THEME_PROPERTY_SYMBOL_ID = 2421
	TMT_MINDPI3                 THEME_PROPERTY_SYMBOL_ID = 2422
	TMT_MINDPI4                 THEME_PROPERTY_SYMBOL_ID = 2423
	TMT_MINDPI5                 THEME_PROPERTY_SYMBOL_ID = 2424
	TMT_TEXTGLOWSIZE            THEME_PROPERTY_SYMBOL_ID = 2425
	TMT_FRAMESPERSECOND         THEME_PROPERTY_SYMBOL_ID = 2426
	TMT_PIXELSPERFRAME          THEME_PROPERTY_SYMBOL_ID = 2427
	TMT_ANIMATIONDELAY          THEME_PROPERTY_SYMBOL_ID = 2428
	TMT_GLOWINTENSITY           THEME_PROPERTY_SYMBOL_ID = 2429
	TMT_OPACITY                 THEME_PROPERTY_SYMBOL_ID = 2430
	TMT_COLORIZATIONCOLOR       THEME_PROPERTY_SYMBOL_ID = 2431
	TMT_COLORIZATIONOPACITY     THEME_PROPERTY_SYMBOL_ID = 2432
	TMT_MINDPI6                 THEME_PROPERTY_SYMBOL_ID = 2433
	TMT_MINDPI7                 THEME_PROPERTY_SYMBOL_ID = 2434
	TMT_GLYPHFONT               THEME_PROPERTY_SYMBOL_ID = 2601
	TMT_IMAGEFILE               THEME_PROPERTY_SYMBOL_ID = 3001
	TMT_IMAGEFILE1              THEME_PROPERTY_SYMBOL_ID = 3002
	TMT_IMAGEFILE2              THEME_PROPERTY_SYMBOL_ID = 3003
	TMT_IMAGEFILE3              THEME_PROPERTY_SYMBOL_ID = 3004
	TMT_IMAGEFILE4              THEME_PROPERTY_SYMBOL_ID = 3005
	TMT_IMAGEFILE5              THEME_PROPERTY_SYMBOL_ID = 3006
	TMT_GLYPHIMAGEFILE          THEME_PROPERTY_SYMBOL_ID = 3008
	TMT_IMAGEFILE6              THEME_PROPERTY_SYMBOL_ID = 3009
	TMT_IMAGEFILE7              THEME_PROPERTY_SYMBOL_ID = 3010
	TMT_TEXT                    THEME_PROPERTY_SYMBOL_ID = 3201
	TMT_CLASSICVALUE            THEME_PROPERTY_SYMBOL_ID = 3202
	TMT_OFFSET                  THEME_PROPERTY_SYMBOL_ID = 3401
	TMT_TEXTSHADOWOFFSET        THEME_PROPERTY_SYMBOL_ID = 3402
	TMT_MINSIZE                 THEME_PROPERTY_SYMBOL_ID = 3403
	TMT_MINSIZE1                THEME_PROPERTY_SYMBOL_ID = 3404
	TMT_MINSIZE2                THEME_PROPERTY_SYMBOL_ID = 3405
	TMT_MINSIZE3                THEME_PROPERTY_SYMBOL_ID = 3406
	TMT_MINSIZE4                THEME_PROPERTY_SYMBOL_ID = 3407
	TMT_MINSIZE5                THEME_PROPERTY_SYMBOL_ID = 3408
	TMT_NORMALSIZE              THEME_PROPERTY_SYMBOL_ID = 3409
	TMT_MINSIZE6                THEME_PROPERTY_SYMBOL_ID = 3410
	TMT_MINSIZE7                THEME_PROPERTY_SYMBOL_ID = 3411
	TMT_SIZINGMARGINS           THEME_PROPERTY_SYMBOL_ID = 3601
	TMT_CONTENTMARGINS          THEME_PROPERTY_SYMBOL_ID = 3602
	TMT_CAPTIONMARGINS          THEME_PROPERTY_SYMBOL_ID = 3603
	TMT_BORDERCOLOR             THEME_PROPERTY_SYMBOL_ID = 3801
	TMT_FILLCOLOR               THEME_PROPERTY_SYMBOL_ID = 3802
	TMT_TEXTCOLOR               THEME_PROPERTY_SYMBOL_ID = 3803
	TMT_EDGELIGHTCOLOR          THEME_PROPERTY_SYMBOL_ID = 3804
	TMT_EDGEHIGHLIGHTCOLOR      THEME_PROPERTY_SYMBOL_ID = 3805
	TMT_EDGESHADOWCOLOR         THEME_PROPERTY_SYMBOL_ID = 3806
	TMT_EDGEDKSHADOWCOLOR       THEME_PROPERTY_SYMBOL_ID = 3807
	TMT_EDGEFILLCOLOR           THEME_PROPERTY_SYMBOL_ID = 3808
	TMT_TRANSPARENTCOLOR        THEME_PROPERTY_SYMBOL_ID = 3809
	TMT_GRADIENTCOLOR1          THEME_PROPERTY_SYMBOL_ID = 3810
	TMT_GRADIENTCOLOR2          THEME_PROPERTY_SYMBOL_ID = 3811
	TMT_GRADIENTCOLOR3          THEME_PROPERTY_SYMBOL_ID = 3812
	TMT_GRADIENTCOLOR4          THEME_PROPERTY_SYMBOL_ID = 3813
	TMT_GRADIENTCOLOR5          THEME_PROPERTY_SYMBOL_ID = 3814
	TMT_SHADOWCOLOR             THEME_PROPERTY_SYMBOL_ID = 3815
	TMT_GLOWCOLOR               THEME_PROPERTY_SYMBOL_ID = 3816
	TMT_TEXTBORDERCOLOR         THEME_PROPERTY_SYMBOL_ID = 3817
	TMT_TEXTSHADOWCOLOR         THEME_PROPERTY_SYMBOL_ID = 3818
	TMT_GLYPHTEXTCOLOR          THEME_PROPERTY_SYMBOL_ID = 3819
	TMT_GLYPHTRANSPARENTCOLOR   THEME_PROPERTY_SYMBOL_ID = 3820
	TMT_FILLCOLORHINT           THEME_PROPERTY_SYMBOL_ID = 3821
	TMT_BORDERCOLORHINT         THEME_PROPERTY_SYMBOL_ID = 3822
	TMT_ACCENTCOLORHINT         THEME_PROPERTY_SYMBOL_ID = 3823
	TMT_TEXTCOLORHINT           THEME_PROPERTY_SYMBOL_ID = 3824
	TMT_HEADING1TEXTCOLOR       THEME_PROPERTY_SYMBOL_ID = 3825
	TMT_HEADING2TEXTCOLOR       THEME_PROPERTY_SYMBOL_ID = 3826
	TMT_BODYTEXTCOLOR           THEME_PROPERTY_SYMBOL_ID = 3827
	TMT_BGTYPE                  THEME_PROPERTY_SYMBOL_ID = 4001
	TMT_BORDERTYPE              THEME_PROPERTY_SYMBOL_ID = 4002
	TMT_FILLTYPE                THEME_PROPERTY_SYMBOL_ID = 4003
	TMT_SIZINGTYPE              THEME_PROPERTY_SYMBOL_ID = 4004
	TMT_HALIGN                  THEME_PROPERTY_SYMBOL_ID = 4005
	TMT_CONTENTALIGNMENT        THEME_PROPERTY_SYMBOL_ID = 4006
	TMT_VALIGN                  THEME_PROPERTY_SYMBOL_ID = 4007
	TMT_OFFSETTYPE              THEME_PROPERTY_SYMBOL_ID = 4008
	TMT_ICONEFFECT              THEME_PROPERTY_SYMBOL_ID = 4009
	TMT_TEXTSHADOWTYPE          THEME_PROPERTY_SYMBOL_ID = 4010
	TMT_IMAGELAYOUT             THEME_PROPERTY_SYMBOL_ID = 4011
	TMT_GLYPHTYPE               THEME_PROPERTY_SYMBOL_ID = 4012
	TMT_IMAGESELECTTYPE         THEME_PROPERTY_SYMBOL_ID = 4013
	TMT_GLYPHFONTSIZINGTYPE     THEME_PROPERTY_SYMBOL_ID = 4014
	TMT_TRUESIZESCALINGTYPE     THEME_PROPERTY_SYMBOL_ID = 4015
	TMT_USERPICTURE             THEME_PROPERTY_SYMBOL_ID = 5001
	TMT_DEFAULTPANESIZE         THEME_PROPERTY_SYMBOL_ID = 5002
	TMT_BLENDCOLOR              THEME_PROPERTY_SYMBOL_ID = 5003
	TMT_CUSTOMSPLITRECT         THEME_PROPERTY_SYMBOL_ID = 5004
	TMT_ANIMATIONBUTTONRECT     THEME_PROPERTY_SYMBOL_ID = 5005
	TMT_ANIMATIONDURATION       THEME_PROPERTY_SYMBOL_ID = 5006
	TMT_TRANSITIONDURATIONS     THEME_PROPERTY_SYMBOL_ID = 6000
	TMT_SCALEDBACKGROUND        THEME_PROPERTY_SYMBOL_ID = 7001
	TMT_ATLASIMAGE              THEME_PROPERTY_SYMBOL_ID = 8000
	TMT_ATLASINPUTIMAGE         THEME_PROPERTY_SYMBOL_ID = 8001
	TMT_ATLASRECT               THEME_PROPERTY_SYMBOL_ID = 8002
)

// enum
type DRAGLISTINFO_NOTIFICATION_FLAGS uint32

const (
	DL_BEGINDRAG  DRAGLISTINFO_NOTIFICATION_FLAGS = 1157
	DL_CANCELDRAG DRAGLISTINFO_NOTIFICATION_FLAGS = 1160
	DL_DRAGGING   DRAGLISTINFO_NOTIFICATION_FLAGS = 1158
	DL_DROPPED    DRAGLISTINFO_NOTIFICATION_FLAGS = 1159
)

// enum
type WORD_BREAK_ACTION uint32

const (
	WB_CLASSIFY      WORD_BREAK_ACTION = 3
	WB_ISDELIMITER   WORD_BREAK_ACTION = 2
	WB_LEFT          WORD_BREAK_ACTION = 0
	WB_LEFTBREAK     WORD_BREAK_ACTION = 6
	WB_MOVEWORDLEFT  WORD_BREAK_ACTION = 4
	WB_MOVEWORDRIGHT WORD_BREAK_ACTION = 5
	WB_RIGHT         WORD_BREAK_ACTION = 1
	WB_RIGHTBREAK    WORD_BREAK_ACTION = 7
)

// enum
type DPAMM_MESSAGE uint32

const (
	DPAMM_MERGE  DPAMM_MESSAGE = 1
	DPAMM_DELETE DPAMM_MESSAGE = 2
	DPAMM_INSERT DPAMM_MESSAGE = 3
)

// enum
// flags
type DLG_DIR_LIST_FILE_TYPE uint32

const (
	DDL_ARCHIVE   DLG_DIR_LIST_FILE_TYPE = 32
	DDL_DIRECTORY DLG_DIR_LIST_FILE_TYPE = 16
	DDL_DRIVES    DLG_DIR_LIST_FILE_TYPE = 16384
	DDL_EXCLUSIVE DLG_DIR_LIST_FILE_TYPE = 32768
	DDL_HIDDEN    DLG_DIR_LIST_FILE_TYPE = 2
	DDL_READONLY  DLG_DIR_LIST_FILE_TYPE = 1
	DDL_READWRITE DLG_DIR_LIST_FILE_TYPE = 0
	DDL_SYSTEM    DLG_DIR_LIST_FILE_TYPE = 4
	DDL_POSTMSGS  DLG_DIR_LIST_FILE_TYPE = 8192
)

// enum
// flags
type OPEN_THEME_DATA_FLAGS uint32

const (
	OTD_FORCE_RECT_SIZING OPEN_THEME_DATA_FLAGS = 1
	OTD_NONCLIENT         OPEN_THEME_DATA_FLAGS = 2
)

// enum
type GET_THEME_BITMAP_FLAGS uint32

const (
	GBF_DIRECT    GET_THEME_BITMAP_FLAGS = 1
	GBF_COPY      GET_THEME_BITMAP_FLAGS = 2
	GBF_VALIDBITS GET_THEME_BITMAP_FLAGS = 3
)

// enum
type ENABLE_SCROLL_BAR_ARROWS uint32

const (
	ESB_DISABLE_BOTH  ENABLE_SCROLL_BAR_ARROWS = 3
	ESB_DISABLE_DOWN  ENABLE_SCROLL_BAR_ARROWS = 2
	ESB_DISABLE_LEFT  ENABLE_SCROLL_BAR_ARROWS = 1
	ESB_DISABLE_LTUP  ENABLE_SCROLL_BAR_ARROWS = 1
	ESB_DISABLE_RIGHT ENABLE_SCROLL_BAR_ARROWS = 2
	ESB_DISABLE_RTDN  ENABLE_SCROLL_BAR_ARROWS = 2
	ESB_DISABLE_UP    ENABLE_SCROLL_BAR_ARROWS = 1
	ESB_ENABLE_BOTH   ENABLE_SCROLL_BAR_ARROWS = 0
)

// enum
// flags
type IMAGE_LIST_DRAW_STYLE uint32

const (
	ILD_BLEND    IMAGE_LIST_DRAW_STYLE = 4
	ILD_BLEND50  IMAGE_LIST_DRAW_STYLE = 4
	ILD_FOCUS    IMAGE_LIST_DRAW_STYLE = 2
	ILD_MASK     IMAGE_LIST_DRAW_STYLE = 16
	ILD_NORMAL   IMAGE_LIST_DRAW_STYLE = 0
	ILD_SELECTED IMAGE_LIST_DRAW_STYLE = 4
)

// enum
type WSB_PROP int32

const (
	WSB_PROP_CXHSCROLL WSB_PROP = 2
	WSB_PROP_CXHTHUMB  WSB_PROP = 16
	WSB_PROP_CXVSCROLL WSB_PROP = 8
	WSB_PROP_CYHSCROLL WSB_PROP = 4
	WSB_PROP_CYVSCROLL WSB_PROP = 1
	WSB_PROP_CYVTHUMB  WSB_PROP = 32
	WSB_PROP_HBKGCOLOR WSB_PROP = 128
	WSB_PROP_HSTYLE    WSB_PROP = 512
	WSB_PROP_PALETTE   WSB_PROP = 2048
	WSB_PROP_VBKGCOLOR WSB_PROP = 64
	WSB_PROP_VSTYLE    WSB_PROP = 256
	WSB_PROP_WINSTYLE  WSB_PROP = 1024
)

// enum
type PSPCB_MESSAGE uint32

const (
	PSPCB_ADDREF        PSPCB_MESSAGE = 0
	PSPCB_CREATE        PSPCB_MESSAGE = 2
	PSPCB_RELEASE       PSPCB_MESSAGE = 1
	PSPCB_SI_INITDIALOG PSPCB_MESSAGE = 1025
)

// enum
type HEADER_CONTROL_NOTIFICATION_BUTTON uint32

const (
	HEADER_CONTROL_NOTIFICATION_BUTTON_LEFT   HEADER_CONTROL_NOTIFICATION_BUTTON = 0
	HEADER_CONTROL_NOTIFICATION_BUTTON_RIGHT  HEADER_CONTROL_NOTIFICATION_BUTTON = 1
	HEADER_CONTROL_NOTIFICATION_BUTTON_MIDDLE HEADER_CONTROL_NOTIFICATION_BUTTON = 2
)

// enum
type IMAGE_LIST_COPY_FLAGS uint32

const (
	ILCF_MOVE IMAGE_LIST_COPY_FLAGS = 0
	ILCF_SWAP IMAGE_LIST_COPY_FLAGS = 1
)

// enum
type DLG_BUTTON_CHECK_STATE uint32

const (
	BST_CHECKED       DLG_BUTTON_CHECK_STATE = 1
	BST_INDETERMINATE DLG_BUTTON_CHECK_STATE = 2
	BST_UNCHECKED     DLG_BUTTON_CHECK_STATE = 0
)

// enum
// flags
type DRAW_THEME_PARENT_BACKGROUND_FLAGS uint32

const (
	DTPB_WINDOWDC          DRAW_THEME_PARENT_BACKGROUND_FLAGS = 1
	DTPB_USECTLCOLORSTATIC DRAW_THEME_PARENT_BACKGROUND_FLAGS = 2
	DTPB_USEERASEBKGND     DRAW_THEME_PARENT_BACKGROUND_FLAGS = 4
)

// enum
type IMAGE_LIST_ITEM_FLAGS uint32

const (
	ILIF_ALPHA      IMAGE_LIST_ITEM_FLAGS = 1
	ILIF_LOWQUALITY IMAGE_LIST_ITEM_FLAGS = 2
)

// enum
// flags
type HDI_MASK uint32

const (
	HDI_WIDTH      HDI_MASK = 1
	HDI_HEIGHT     HDI_MASK = 1
	HDI_TEXT       HDI_MASK = 2
	HDI_FORMAT     HDI_MASK = 4
	HDI_LPARAM     HDI_MASK = 8
	HDI_BITMAP     HDI_MASK = 16
	HDI_IMAGE      HDI_MASK = 32
	HDI_DI_SETITEM HDI_MASK = 64
	HDI_ORDER      HDI_MASK = 128
	HDI_FILTER     HDI_MASK = 256
	HDI_STATE      HDI_MASK = 512
)

// enum
// flags
type NMREBAR_MASK_FLAGS uint32

const (
	RBNM_ID     NMREBAR_MASK_FLAGS = 1
	RBNM_LPARAM NMREBAR_MASK_FLAGS = 4
	RBNM_STYLE  NMREBAR_MASK_FLAGS = 2
)

// enum
type EDITBALLOONTIP_ICON uint32

const (
	TTI_ERROR         EDITBALLOONTIP_ICON = 3
	TTI_INFO          EDITBALLOONTIP_ICON = 1
	TTI_NONE          EDITBALLOONTIP_ICON = 0
	TTI_WARNING       EDITBALLOONTIP_ICON = 2
	TTI_INFO_LARGE    EDITBALLOONTIP_ICON = 4
	TTI_WARNING_LARGE EDITBALLOONTIP_ICON = 5
	TTI_ERROR_LARGE   EDITBALLOONTIP_ICON = 6
)

// enum
// flags
type LVCOLUMNW_FORMAT uint32

const (
	LVCFMT_LEFT            LVCOLUMNW_FORMAT = 0
	LVCFMT_RIGHT           LVCOLUMNW_FORMAT = 1
	LVCFMT_CENTER          LVCOLUMNW_FORMAT = 2
	LVCFMT_JUSTIFYMASK     LVCOLUMNW_FORMAT = 3
	LVCFMT_IMAGE           LVCOLUMNW_FORMAT = 2048
	LVCFMT_BITMAP_ON_RIGHT LVCOLUMNW_FORMAT = 4096
	LVCFMT_COL_HAS_IMAGES  LVCOLUMNW_FORMAT = 32768
	LVCFMT_FIXED_WIDTH     LVCOLUMNW_FORMAT = 256
	LVCFMT_NO_DPI_SCALE    LVCOLUMNW_FORMAT = 262144
	LVCFMT_FIXED_RATIO     LVCOLUMNW_FORMAT = 524288
	LVCFMT_SPLITBUTTON     LVCOLUMNW_FORMAT = 16777216
)

// enum
// flags
type NMPGSCROLL_KEYS uint16

const (
	PGK_NONE    NMPGSCROLL_KEYS = 0
	PGK_SHIFT   NMPGSCROLL_KEYS = 1
	PGK_CONTROL NMPGSCROLL_KEYS = 2
	PGK_MENU    NMPGSCROLL_KEYS = 4
)

// enum
// flags
type COMBOBOX_EX_ITEM_FLAGS uint32

const (
	CBEIF_DI_SETITEM    COMBOBOX_EX_ITEM_FLAGS = 268435456
	CBEIF_IMAGE         COMBOBOX_EX_ITEM_FLAGS = 2
	CBEIF_INDENT        COMBOBOX_EX_ITEM_FLAGS = 16
	CBEIF_LPARAM        COMBOBOX_EX_ITEM_FLAGS = 32
	CBEIF_OVERLAY       COMBOBOX_EX_ITEM_FLAGS = 8
	CBEIF_SELECTEDIMAGE COMBOBOX_EX_ITEM_FLAGS = 4
	CBEIF_TEXT          COMBOBOX_EX_ITEM_FLAGS = 1
)

// enum
type TVITEMEXW_CHILDREN int32

const (
	I_ZERO             TVITEMEXW_CHILDREN = 0
	I_ONE_OR_MORE      TVITEMEXW_CHILDREN = 1
	I_CHILDRENCALLBACK TVITEMEXW_CHILDREN = -1
	I_CHILDRENAUTO     TVITEMEXW_CHILDREN = -2
)

// enum
// flags
type TVITEM_MASK uint32

const (
	TVIF_CHILDREN      TVITEM_MASK = 64
	TVIF_DI_SETITEM    TVITEM_MASK = 4096
	TVIF_HANDLE        TVITEM_MASK = 16
	TVIF_IMAGE         TVITEM_MASK = 2
	TVIF_PARAM         TVITEM_MASK = 4
	TVIF_SELECTEDIMAGE TVITEM_MASK = 32
	TVIF_STATE         TVITEM_MASK = 8
	TVIF_TEXT          TVITEM_MASK = 1
	TVIF_EXPANDEDIMAGE TVITEM_MASK = 512
	TVIF_INTEGRAL      TVITEM_MASK = 128
	TVIF_STATEEX       TVITEM_MASK = 256
)

// enum
// flags
type TCITEMHEADERA_MASK uint32

const (
	TCIF_IMAGE      TCITEMHEADERA_MASK = 2
	TCIF_RTLREADING TCITEMHEADERA_MASK = 4
	TCIF_TEXT       TCITEMHEADERA_MASK = 1
	TCIF_PARAM      TCITEMHEADERA_MASK = 8
	TCIF_STATE      TCITEMHEADERA_MASK = 16
)

// enum
type TCHITTESTINFO_FLAGS uint32

const (
	TCHT_NOWHERE     TCHITTESTINFO_FLAGS = 1
	TCHT_ONITEM      TCHITTESTINFO_FLAGS = 6
	TCHT_ONITEMICON  TCHITTESTINFO_FLAGS = 2
	TCHT_ONITEMLABEL TCHITTESTINFO_FLAGS = 4
)

// enum
type COMBOBOXINFO_BUTTON_STATE uint32

const (
	STATE_SYSTEM_INVISIBLE   COMBOBOXINFO_BUTTON_STATE = 32768
	STATE_SYSTEM_PRESSED     COMBOBOXINFO_BUTTON_STATE = 8
	STATE_SYSTEM_FOCUSABLE   COMBOBOXINFO_BUTTON_STATE = 1048576
	STATE_SYSTEM_OFFSCREEN   COMBOBOXINFO_BUTTON_STATE = 65536
	STATE_SYSTEM_UNAVAILABLE COMBOBOXINFO_BUTTON_STATE = 1
)

// enum
type NMCUSTOMDRAW_DRAW_STAGE uint32

const (
	CDDS_POSTPAINT     NMCUSTOMDRAW_DRAW_STAGE = 2
	CDDS_PREERASE      NMCUSTOMDRAW_DRAW_STAGE = 3
	CDDS_PREPAINT      NMCUSTOMDRAW_DRAW_STAGE = 1
	CDDS_ITEMPOSTERASE NMCUSTOMDRAW_DRAW_STAGE = 65540
	CDDS_ITEMPOSTPAINT NMCUSTOMDRAW_DRAW_STAGE = 65538
	CDDS_ITEMPREERASE  NMCUSTOMDRAW_DRAW_STAGE = 65539
	CDDS_ITEMPREPAINT  NMCUSTOMDRAW_DRAW_STAGE = 65537
	CDDS_SUBITEM       NMCUSTOMDRAW_DRAW_STAGE = 131072
)

// enum
type MCGRIDINFO_PART uint32

const (
	MCGIP_CALENDARCONTROL MCGRIDINFO_PART = 0
	MCGIP_NEXT            MCGRIDINFO_PART = 1
	MCGIP_PREV            MCGRIDINFO_PART = 2
	MCGIP_FOOTER          MCGRIDINFO_PART = 3
	MCGIP_CALENDAR        MCGRIDINFO_PART = 4
	MCGIP_CALENDARHEADER  MCGRIDINFO_PART = 5
	MCGIP_CALENDARBODY    MCGRIDINFO_PART = 6
	MCGIP_CALENDARROW     MCGRIDINFO_PART = 7
	MCGIP_CALENDARCELL    MCGRIDINFO_PART = 8
)

// enum
type LVITEMA_GROUP_ID int32

const (
	I_GROUPIDCALLBACK LVITEMA_GROUP_ID = -1
	I_GROUPIDNONE     LVITEMA_GROUP_ID = -2
)

// enum
// flags
type NMTBHOTITEM_FLAGS uint32

const (
	HICF_ACCELERATOR    NMTBHOTITEM_FLAGS = 4
	HICF_ARROWKEYS      NMTBHOTITEM_FLAGS = 2
	HICF_DUPACCEL       NMTBHOTITEM_FLAGS = 8
	HICF_ENTERING       NMTBHOTITEM_FLAGS = 16
	HICF_LEAVING        NMTBHOTITEM_FLAGS = 32
	HICF_LMOUSE         NMTBHOTITEM_FLAGS = 128
	HICF_MOUSE          NMTBHOTITEM_FLAGS = 1
	HICF_OTHER          NMTBHOTITEM_FLAGS = 0
	HICF_RESELECT       NMTBHOTITEM_FLAGS = 64
	HICF_TOGGLEDROPDOWN NMTBHOTITEM_FLAGS = 256
)

// enum
// flags
type TTTOOLINFO_FLAGS uint32

const (
	TTF_ABSOLUTE    TTTOOLINFO_FLAGS = 128
	TTF_CENTERTIP   TTTOOLINFO_FLAGS = 2
	TTF_IDISHWND    TTTOOLINFO_FLAGS = 1
	TTF_PARSELINKS  TTTOOLINFO_FLAGS = 4096
	TTF_RTLREADING  TTTOOLINFO_FLAGS = 4
	TTF_SUBCLASS    TTTOOLINFO_FLAGS = 16
	TTF_TRACK       TTTOOLINFO_FLAGS = 32
	TTF_TRANSPARENT TTTOOLINFO_FLAGS = 256
)

// enum
type LVTILEVIEWINFO_FLAGS uint32

const (
	LVTVIF_EXTENDED LVTILEVIEWINFO_FLAGS = 4
)

// enum
type NMPGSCROLL_DIR uint32

const (
	PGF_SCROLLDOWN  NMPGSCROLL_DIR = 2
	PGF_SCROLLLEFT  NMPGSCROLL_DIR = 4
	PGF_SCROLLRIGHT NMPGSCROLL_DIR = 8
	PGF_SCROLLUP    NMPGSCROLL_DIR = 1
)

// enum
// flags
type LVCOLUMNW_MASK uint32

const (
	LVCF_FMT          LVCOLUMNW_MASK = 1
	LVCF_WIDTH        LVCOLUMNW_MASK = 2
	LVCF_TEXT         LVCOLUMNW_MASK = 4
	LVCF_SUBITEM      LVCOLUMNW_MASK = 8
	LVCF_IMAGE        LVCOLUMNW_MASK = 16
	LVCF_ORDER        LVCOLUMNW_MASK = 32
	LVCF_MINWIDTH     LVCOLUMNW_MASK = 64
	LVCF_DEFAULTWIDTH LVCOLUMNW_MASK = 128
	LVCF_IDEALWIDTH   LVCOLUMNW_MASK = 256
)

// enum
// flags
type LVFINDINFOW_FLAGS uint32

const (
	LVFI_PARAM     LVFINDINFOW_FLAGS = 1
	LVFI_PARTIAL   LVFINDINFOW_FLAGS = 8
	LVFI_STRING    LVFINDINFOW_FLAGS = 2
	LVFI_SUBSTRING LVFINDINFOW_FLAGS = 4
	LVFI_WRAP      LVFINDINFOW_FLAGS = 32
	LVFI_NEARESTXY LVFINDINFOW_FLAGS = 64
)

// enum
type BUTTON_IMAGELIST_ALIGN uint32

const (
	BUTTON_IMAGELIST_ALIGN_LEFT   BUTTON_IMAGELIST_ALIGN = 0
	BUTTON_IMAGELIST_ALIGN_RIGHT  BUTTON_IMAGELIST_ALIGN = 1
	BUTTON_IMAGELIST_ALIGN_TOP    BUTTON_IMAGELIST_ALIGN = 2
	BUTTON_IMAGELIST_ALIGN_BOTTOM BUTTON_IMAGELIST_ALIGN = 3
	BUTTON_IMAGELIST_ALIGN_CENTER BUTTON_IMAGELIST_ALIGN = 4
)

// enum
// flags
type TBBUTTONINFOW_MASK uint32

const (
	TBIF_BYINDEX TBBUTTONINFOW_MASK = 2147483648
	TBIF_COMMAND TBBUTTONINFOW_MASK = 32
	TBIF_IMAGE   TBBUTTONINFOW_MASK = 1
	TBIF_LPARAM  TBBUTTONINFOW_MASK = 16
	TBIF_SIZE    TBBUTTONINFOW_MASK = 64
	TBIF_STATE   TBBUTTONINFOW_MASK = 4
	TBIF_STYLE   TBBUTTONINFOW_MASK = 8
	TBIF_TEXT    TBBUTTONINFOW_MASK = 2
)

// enum
type TBINSERTMARK_FLAGS uint32

const (
	TBIMHT_NONE       TBINSERTMARK_FLAGS = 0
	TBIMHT_AFTER      TBINSERTMARK_FLAGS = 1
	TBIMHT_BACKGROUND TBINSERTMARK_FLAGS = 2
)

// enum
// flags
type LVGROUP_MASK uint32

const (
	LVGF_NONE   LVGROUP_MASK = 0
	LVGF_HEADER LVGROUP_MASK = 1
	LVGF_FOOTER LVGROUP_MASK = 2
	LVGF_STATE  LVGROUP_MASK = 4
)

// enum
// flags
type BP_PAINTPARAMS_FLAGS uint32

const (
	BPPF_ERASE     BP_PAINTPARAMS_FLAGS = 1
	BPPF_NOCLIP    BP_PAINTPARAMS_FLAGS = 2
	BPPF_NONCLIENT BP_PAINTPARAMS_FLAGS = 4
)

// enum
// flags
type TVHITTESTINFO_FLAGS uint32

const (
	TVHT_ABOVE           TVHITTESTINFO_FLAGS = 256
	TVHT_BELOW           TVHITTESTINFO_FLAGS = 512
	TVHT_NOWHERE         TVHITTESTINFO_FLAGS = 1
	TVHT_ONITEM          TVHITTESTINFO_FLAGS = 70
	TVHT_ONITEMBUTTON    TVHITTESTINFO_FLAGS = 16
	TVHT_ONITEMICON      TVHITTESTINFO_FLAGS = 2
	TVHT_ONITEMINDENT    TVHITTESTINFO_FLAGS = 8
	TVHT_ONITEMLABEL     TVHITTESTINFO_FLAGS = 4
	TVHT_ONITEMRIGHT     TVHITTESTINFO_FLAGS = 32
	TVHT_ONITEMSTATEICON TVHITTESTINFO_FLAGS = 64
	TVHT_TOLEFT          TVHITTESTINFO_FLAGS = 2048
	TVHT_TORIGHT         TVHITTESTINFO_FLAGS = 1024
)

// enum
type DRAWITEMSTRUCT_CTL_TYPE uint32

const (
	ODT_BUTTON   DRAWITEMSTRUCT_CTL_TYPE = 4
	ODT_COMBOBOX DRAWITEMSTRUCT_CTL_TYPE = 3
	ODT_LISTBOX  DRAWITEMSTRUCT_CTL_TYPE = 2
	ODT_LISTVIEW DRAWITEMSTRUCT_CTL_TYPE = 102
	ODT_MENU     DRAWITEMSTRUCT_CTL_TYPE = 1
	ODT_STATIC   DRAWITEMSTRUCT_CTL_TYPE = 5
	ODT_TAB      DRAWITEMSTRUCT_CTL_TYPE = 101
)

// enum
type NMPGCALCSIZE_FLAGS uint32

const (
	PGF_CALCHEIGHT NMPGCALCSIZE_FLAGS = 2
	PGF_CALCWIDTH  NMPGCALCSIZE_FLAGS = 1
)

// enum
type NMLVCUSTOMDRAW_ALIGN uint32

const (
	LVGA_HEADER_CENTER NMLVCUSTOMDRAW_ALIGN = 2
	LVGA_HEADER_LEFT   NMLVCUSTOMDRAW_ALIGN = 1
	LVGA_HEADER_RIGHT  NMLVCUSTOMDRAW_ALIGN = 4
)

// enum
// flags
type MCGRIDINFO_FLAGS uint32

const (
	MCGIF_DATE MCGRIDINFO_FLAGS = 1
	MCGIF_RECT MCGRIDINFO_FLAGS = 2
	MCGIF_NAME MCGRIDINFO_FLAGS = 4
)

// enum
// flags
type LVHITTESTINFO_FLAGS uint32

const (
	LVHT_ABOVE               LVHITTESTINFO_FLAGS = 8
	LVHT_BELOW               LVHITTESTINFO_FLAGS = 16
	LVHT_NOWHERE             LVHITTESTINFO_FLAGS = 1
	LVHT_ONITEMICON          LVHITTESTINFO_FLAGS = 2
	LVHT_ONITEMLABEL         LVHITTESTINFO_FLAGS = 4
	LVHT_ONITEMSTATEICON     LVHITTESTINFO_FLAGS = 8
	LVHT_TOLEFT              LVHITTESTINFO_FLAGS = 64
	LVHT_TORIGHT             LVHITTESTINFO_FLAGS = 32
	LVHT_EX_GROUP_HEADER     LVHITTESTINFO_FLAGS = 268435456
	LVHT_EX_GROUP_FOOTER     LVHITTESTINFO_FLAGS = 536870912
	LVHT_EX_GROUP_COLLAPSE   LVHITTESTINFO_FLAGS = 1073741824
	LVHT_EX_GROUP_BACKGROUND LVHITTESTINFO_FLAGS = 2147483648
	LVHT_EX_GROUP_STATEICON  LVHITTESTINFO_FLAGS = 16777216
	LVHT_EX_GROUP_SUBSETLINK LVHITTESTINFO_FLAGS = 33554432
	LVHT_EX_GROUP            LVHITTESTINFO_FLAGS = 4076863488
	LVHT_EX_ONCONTENTS       LVHITTESTINFO_FLAGS = 67108864
	LVHT_EX_FOOTER           LVHITTESTINFO_FLAGS = 134217728
)

// enum
// flags
type INITCOMMONCONTROLSEX_ICC uint32

const (
	ICC_ANIMATE_CLASS      INITCOMMONCONTROLSEX_ICC = 128
	ICC_BAR_CLASSES        INITCOMMONCONTROLSEX_ICC = 4
	ICC_COOL_CLASSES       INITCOMMONCONTROLSEX_ICC = 1024
	ICC_DATE_CLASSES       INITCOMMONCONTROLSEX_ICC = 256
	ICC_HOTKEY_CLASS       INITCOMMONCONTROLSEX_ICC = 64
	ICC_INTERNET_CLASSES   INITCOMMONCONTROLSEX_ICC = 2048
	ICC_LINK_CLASS         INITCOMMONCONTROLSEX_ICC = 32768
	ICC_LISTVIEW_CLASSES   INITCOMMONCONTROLSEX_ICC = 1
	ICC_NATIVEFNTCTL_CLASS INITCOMMONCONTROLSEX_ICC = 8192
	ICC_PAGESCROLLER_CLASS INITCOMMONCONTROLSEX_ICC = 4096
	ICC_PROGRESS_CLASS     INITCOMMONCONTROLSEX_ICC = 32
	ICC_STANDARD_CLASSES   INITCOMMONCONTROLSEX_ICC = 16384
	ICC_TAB_CLASSES        INITCOMMONCONTROLSEX_ICC = 8
	ICC_TREEVIEW_CLASSES   INITCOMMONCONTROLSEX_ICC = 2
	ICC_UPDOWN_CLASS       INITCOMMONCONTROLSEX_ICC = 16
	ICC_USEREX_CLASSES     INITCOMMONCONTROLSEX_ICC = 512
	ICC_WIN95_CLASSES      INITCOMMONCONTROLSEX_ICC = 255
)

// enum
type NMLVCUSTOMDRAW_ITEM_TYPE uint32

const (
	LVCDI_ITEM      NMLVCUSTOMDRAW_ITEM_TYPE = 0
	LVCDI_GROUP     NMLVCUSTOMDRAW_ITEM_TYPE = 1
	LVCDI_ITEMSLIST NMLVCUSTOMDRAW_ITEM_TYPE = 2
)

// enum
// flags
type NMTBDISPINFOW_MASK uint32

const (
	TBNF_IMAGE      NMTBDISPINFOW_MASK = 1
	TBNF_TEXT       NMTBDISPINFOW_MASK = 2
	TBNF_DI_SETITEM NMTBDISPINFOW_MASK = 268435456
)

// enum
type NMLVEMPTYMARKUP_FLAGS uint32

const (
	EMF_CENTERED NMLVEMPTYMARKUP_FLAGS = 1
)

// enum
type LVFOOTERITEM_MASK uint32

const (
	LVFIF_TEXT  LVFOOTERITEM_MASK = 1
	LVFIF_STATE LVFOOTERITEM_MASK = 2
)

// enum
// flags
type IMAGELIST_CREATION_FLAGS uint32

const (
	ILC_MASK             IMAGELIST_CREATION_FLAGS = 1
	ILC_COLOR            IMAGELIST_CREATION_FLAGS = 0
	ILC_COLORDDB         IMAGELIST_CREATION_FLAGS = 254
	ILC_COLOR4           IMAGELIST_CREATION_FLAGS = 4
	ILC_COLOR8           IMAGELIST_CREATION_FLAGS = 8
	ILC_COLOR16          IMAGELIST_CREATION_FLAGS = 16
	ILC_COLOR24          IMAGELIST_CREATION_FLAGS = 24
	ILC_COLOR32          IMAGELIST_CREATION_FLAGS = 32
	ILC_PALETTE          IMAGELIST_CREATION_FLAGS = 2048
	ILC_MIRROR           IMAGELIST_CREATION_FLAGS = 8192
	ILC_PERITEMMIRROR    IMAGELIST_CREATION_FLAGS = 32768
	ILC_ORIGINALSIZE     IMAGELIST_CREATION_FLAGS = 65536
	ILC_HIGHQUALITYSCALE IMAGELIST_CREATION_FLAGS = 131072
)

// enum
type TVITEMPART int32

const (
	TVGIPR_BUTTON TVITEMPART = 1
)

// enum
type EC_ENDOFLINE int32

const (
	EC_ENDOFLINE_DETECTFROMCONTENT EC_ENDOFLINE = 0
	EC_ENDOFLINE_CRLF              EC_ENDOFLINE = 1
	EC_ENDOFLINE_CR                EC_ENDOFLINE = 2
	EC_ENDOFLINE_LF                EC_ENDOFLINE = 3
)

// enum
type EC_SEARCHWEB_ENTRYPOINT int32

const (
	EC_SEARCHWEB_ENTRYPOINT_EXTERNAL    EC_SEARCHWEB_ENTRYPOINT = 0
	EC_SEARCHWEB_ENTRYPOINT_CONTEXTMENU EC_SEARCHWEB_ENTRYPOINT = 1
)

// enum
type TASKDIALOG_FLAGS int32

const (
	TDF_ENABLE_HYPERLINKS           TASKDIALOG_FLAGS = 1
	TDF_USE_HICON_MAIN              TASKDIALOG_FLAGS = 2
	TDF_USE_HICON_FOOTER            TASKDIALOG_FLAGS = 4
	TDF_ALLOW_DIALOG_CANCELLATION   TASKDIALOG_FLAGS = 8
	TDF_USE_COMMAND_LINKS           TASKDIALOG_FLAGS = 16
	TDF_USE_COMMAND_LINKS_NO_ICON   TASKDIALOG_FLAGS = 32
	TDF_EXPAND_FOOTER_AREA          TASKDIALOG_FLAGS = 64
	TDF_EXPANDED_BY_DEFAULT         TASKDIALOG_FLAGS = 128
	TDF_VERIFICATION_FLAG_CHECKED   TASKDIALOG_FLAGS = 256
	TDF_SHOW_PROGRESS_BAR           TASKDIALOG_FLAGS = 512
	TDF_SHOW_MARQUEE_PROGRESS_BAR   TASKDIALOG_FLAGS = 1024
	TDF_CALLBACK_TIMER              TASKDIALOG_FLAGS = 2048
	TDF_POSITION_RELATIVE_TO_WINDOW TASKDIALOG_FLAGS = 4096
	TDF_RTL_LAYOUT                  TASKDIALOG_FLAGS = 8192
	TDF_NO_DEFAULT_RADIO_BUTTON     TASKDIALOG_FLAGS = 16384
	TDF_CAN_BE_MINIMIZED            TASKDIALOG_FLAGS = 32768
	TDF_NO_SET_FOREGROUND           TASKDIALOG_FLAGS = 65536
	TDF_SIZE_TO_CONTENT             TASKDIALOG_FLAGS = 16777216
)

// enum
type TASKDIALOG_MESSAGES int32

const (
	TDM_NAVIGATE_PAGE                       TASKDIALOG_MESSAGES = 1125
	TDM_CLICK_BUTTON                        TASKDIALOG_MESSAGES = 1126
	TDM_SET_MARQUEE_PROGRESS_BAR            TASKDIALOG_MESSAGES = 1127
	TDM_SET_PROGRESS_BAR_STATE              TASKDIALOG_MESSAGES = 1128
	TDM_SET_PROGRESS_BAR_RANGE              TASKDIALOG_MESSAGES = 1129
	TDM_SET_PROGRESS_BAR_POS                TASKDIALOG_MESSAGES = 1130
	TDM_SET_PROGRESS_BAR_MARQUEE            TASKDIALOG_MESSAGES = 1131
	TDM_SET_ELEMENT_TEXT                    TASKDIALOG_MESSAGES = 1132
	TDM_CLICK_RADIO_BUTTON                  TASKDIALOG_MESSAGES = 1134
	TDM_ENABLE_BUTTON                       TASKDIALOG_MESSAGES = 1135
	TDM_ENABLE_RADIO_BUTTON                 TASKDIALOG_MESSAGES = 1136
	TDM_CLICK_VERIFICATION                  TASKDIALOG_MESSAGES = 1137
	TDM_UPDATE_ELEMENT_TEXT                 TASKDIALOG_MESSAGES = 1138
	TDM_SET_BUTTON_ELEVATION_REQUIRED_STATE TASKDIALOG_MESSAGES = 1139
	TDM_UPDATE_ICON                         TASKDIALOG_MESSAGES = 1140
)

// enum
type TASKDIALOG_NOTIFICATIONS int32

const (
	TDN_CREATED                TASKDIALOG_NOTIFICATIONS = 0
	TDN_NAVIGATED              TASKDIALOG_NOTIFICATIONS = 1
	TDN_BUTTON_CLICKED         TASKDIALOG_NOTIFICATIONS = 2
	TDN_HYPERLINK_CLICKED      TASKDIALOG_NOTIFICATIONS = 3
	TDN_TIMER                  TASKDIALOG_NOTIFICATIONS = 4
	TDN_DESTROYED              TASKDIALOG_NOTIFICATIONS = 5
	TDN_RADIO_BUTTON_CLICKED   TASKDIALOG_NOTIFICATIONS = 6
	TDN_DIALOG_CONSTRUCTED     TASKDIALOG_NOTIFICATIONS = 7
	TDN_VERIFICATION_CLICKED   TASKDIALOG_NOTIFICATIONS = 8
	TDN_HELP                   TASKDIALOG_NOTIFICATIONS = 9
	TDN_EXPANDO_BUTTON_CLICKED TASKDIALOG_NOTIFICATIONS = 10
)

// enum
type TASKDIALOG_ELEMENTS int32

const (
	TDE_CONTENT              TASKDIALOG_ELEMENTS = 0
	TDE_EXPANDED_INFORMATION TASKDIALOG_ELEMENTS = 1
	TDE_FOOTER               TASKDIALOG_ELEMENTS = 2
	TDE_MAIN_INSTRUCTION     TASKDIALOG_ELEMENTS = 3
)

// enum
type TASKDIALOG_ICON_ELEMENTS int32

const (
	TDIE_ICON_MAIN   TASKDIALOG_ICON_ELEMENTS = 0
	TDIE_ICON_FOOTER TASKDIALOG_ICON_ELEMENTS = 1
)

// enum
type TASKDIALOG_COMMON_BUTTON_FLAGS int32

const (
	TDCBF_OK_BUTTON     TASKDIALOG_COMMON_BUTTON_FLAGS = 1
	TDCBF_YES_BUTTON    TASKDIALOG_COMMON_BUTTON_FLAGS = 2
	TDCBF_NO_BUTTON     TASKDIALOG_COMMON_BUTTON_FLAGS = 4
	TDCBF_CANCEL_BUTTON TASKDIALOG_COMMON_BUTTON_FLAGS = 8
	TDCBF_RETRY_BUTTON  TASKDIALOG_COMMON_BUTTON_FLAGS = 16
	TDCBF_CLOSE_BUTTON  TASKDIALOG_COMMON_BUTTON_FLAGS = 32
)

// enum
type LI_METRIC_ int32

const (
	LIM_SMALL LI_METRIC_ = 0
	LIM_LARGE LI_METRIC_ = 1
)

// enum
type TA_PROPERTY int32

const (
	TAP_FLAGS              TA_PROPERTY = 0
	TAP_TRANSFORMCOUNT     TA_PROPERTY = 1
	TAP_STAGGERDELAY       TA_PROPERTY = 2
	TAP_STAGGERDELAYCAP    TA_PROPERTY = 3
	TAP_STAGGERDELAYFACTOR TA_PROPERTY = 4
	TAP_ZORDER             TA_PROPERTY = 5
)

// enum
// flags
type TA_PROPERTY_FLAG uint32

const (
	TAPF_NONE            TA_PROPERTY_FLAG = 0
	TAPF_HASSTAGGER      TA_PROPERTY_FLAG = 1
	TAPF_ISRTLAWARE      TA_PROPERTY_FLAG = 2
	TAPF_ALLOWCOLLECTION TA_PROPERTY_FLAG = 4
	TAPF_HASBACKGROUND   TA_PROPERTY_FLAG = 8
	TAPF_HASPERSPECTIVE  TA_PROPERTY_FLAG = 16
)

// enum
type TA_TRANSFORM_TYPE int32

const (
	TATT_TRANSLATE_2D TA_TRANSFORM_TYPE = 0
	TATT_SCALE_2D     TA_TRANSFORM_TYPE = 1
	TATT_OPACITY      TA_TRANSFORM_TYPE = 2
	TATT_CLIP         TA_TRANSFORM_TYPE = 3
)

// enum
type TA_TRANSFORM_FLAG int32

const (
	TATF_NONE              TA_TRANSFORM_FLAG = 0
	TATF_TARGETVALUES_USER TA_TRANSFORM_FLAG = 1
	TATF_HASINITIALVALUES  TA_TRANSFORM_FLAG = 2
	TATF_HASORIGINVALUES   TA_TRANSFORM_FLAG = 4
)

// enum
type TA_TIMINGFUNCTION_TYPE int32

const (
	TTFT_UNDEFINED    TA_TIMINGFUNCTION_TYPE = 0
	TTFT_CUBIC_BEZIER TA_TIMINGFUNCTION_TYPE = 1
)

// enum
type THEMESIZE int32

const (
	TS_MIN  THEMESIZE = 0
	TS_TRUE THEMESIZE = 1
	TS_DRAW THEMESIZE = 2
)

// enum
type PROPERTYORIGIN int32

const (
	PO_STATE    PROPERTYORIGIN = 0
	PO_PART     PROPERTYORIGIN = 1
	PO_CLASS    PROPERTYORIGIN = 2
	PO_GLOBAL   PROPERTYORIGIN = 3
	PO_NOTFOUND PROPERTYORIGIN = 4
)

// enum
type WINDOWTHEMEATTRIBUTETYPE int32

const (
	WTA_NONCLIENT WINDOWTHEMEATTRIBUTETYPE = 1
)

// enum
type BP_BUFFERFORMAT int32

const (
	BPBF_COMPATIBLEBITMAP BP_BUFFERFORMAT = 0
	BPBF_DIB              BP_BUFFERFORMAT = 1
	BPBF_TOPDOWNDIB       BP_BUFFERFORMAT = 2
	BPBF_TOPDOWNMONODIB   BP_BUFFERFORMAT = 3
)

// enum
type BP_ANIMATIONSTYLE int32

const (
	BPAS_NONE   BP_ANIMATIONSTYLE = 0
	BPAS_LINEAR BP_ANIMATIONSTYLE = 1
	BPAS_CUBIC  BP_ANIMATIONSTYLE = 2
	BPAS_SINE   BP_ANIMATIONSTYLE = 3
)

// enum
type BGTYPE int32

const (
	BT_IMAGEFILE  BGTYPE = 0
	BT_BORDERFILL BGTYPE = 1
	BT_NONE       BGTYPE = 2
)

// enum
type IMAGELAYOUT int32

const (
	IL_VERTICAL   IMAGELAYOUT = 0
	IL_HORIZONTAL IMAGELAYOUT = 1
)

// enum
type BORDERTYPE int32

const (
	BT_RECT      BORDERTYPE = 0
	BT_ROUNDRECT BORDERTYPE = 1
	BT_ELLIPSE   BORDERTYPE = 2
)

// enum
type FILLTYPE int32

const (
	FT_SOLID          FILLTYPE = 0
	FT_VERTGRADIENT   FILLTYPE = 1
	FT_HORZGRADIENT   FILLTYPE = 2
	FT_RADIALGRADIENT FILLTYPE = 3
	FT_TILEIMAGE      FILLTYPE = 4
)

// enum
type SIZINGTYPE int32

const (
	ST_TRUESIZE SIZINGTYPE = 0
	ST_STRETCH  SIZINGTYPE = 1
	ST_TILE     SIZINGTYPE = 2
)

// enum
type HALIGN int32

const (
	HA_LEFT   HALIGN = 0
	HA_CENTER HALIGN = 1
	HA_RIGHT  HALIGN = 2
)

// enum
type CONTENTALIGNMENT int32

const (
	CA_LEFT   CONTENTALIGNMENT = 0
	CA_CENTER CONTENTALIGNMENT = 1
	CA_RIGHT  CONTENTALIGNMENT = 2
)

// enum
type VALIGN int32

const (
	VA_TOP    VALIGN = 0
	VA_CENTER VALIGN = 1
	VA_BOTTOM VALIGN = 2
)

// enum
type OFFSETTYPE int32

const (
	OT_TOPLEFT           OFFSETTYPE = 0
	OT_TOPRIGHT          OFFSETTYPE = 1
	OT_TOPMIDDLE         OFFSETTYPE = 2
	OT_BOTTOMLEFT        OFFSETTYPE = 3
	OT_BOTTOMRIGHT       OFFSETTYPE = 4
	OT_BOTTOMMIDDLE      OFFSETTYPE = 5
	OT_MIDDLELEFT        OFFSETTYPE = 6
	OT_MIDDLERIGHT       OFFSETTYPE = 7
	OT_LEFTOFCAPTION     OFFSETTYPE = 8
	OT_RIGHTOFCAPTION    OFFSETTYPE = 9
	OT_LEFTOFLASTBUTTON  OFFSETTYPE = 10
	OT_RIGHTOFLASTBUTTON OFFSETTYPE = 11
	OT_ABOVELASTBUTTON   OFFSETTYPE = 12
	OT_BELOWLASTBUTTON   OFFSETTYPE = 13
)

// enum
type ICONEFFECT int32

const (
	ICE_NONE   ICONEFFECT = 0
	ICE_GLOW   ICONEFFECT = 1
	ICE_SHADOW ICONEFFECT = 2
	ICE_PULSE  ICONEFFECT = 3
	ICE_ALPHA  ICONEFFECT = 4
)

// enum
type TEXTSHADOWTYPE int32

const (
	TST_NONE       TEXTSHADOWTYPE = 0
	TST_SINGLE     TEXTSHADOWTYPE = 1
	TST_CONTINUOUS TEXTSHADOWTYPE = 2
)

// enum
type GLYPHTYPE int32

const (
	GT_NONE       GLYPHTYPE = 0
	GT_IMAGEGLYPH GLYPHTYPE = 1
	GT_FONTGLYPH  GLYPHTYPE = 2
)

// enum
type IMAGESELECTTYPE int32

const (
	IST_NONE IMAGESELECTTYPE = 0
	IST_SIZE IMAGESELECTTYPE = 1
	IST_DPI  IMAGESELECTTYPE = 2
)

// enum
type TRUESIZESCALINGTYPE int32

const (
	TSST_NONE TRUESIZESCALINGTYPE = 0
	TSST_SIZE TRUESIZESCALINGTYPE = 1
	TSST_DPI  TRUESIZESCALINGTYPE = 2
)

// enum
type GLYPHFONTSIZINGTYPE int32

const (
	GFST_NONE GLYPHFONTSIZINGTYPE = 0
	GFST_SIZE GLYPHFONTSIZINGTYPE = 1
	GFST_DPI  GLYPHFONTSIZINGTYPE = 2
)

// enum
type LINKPARTS int32

const (
	LP_HYPERLINK LINKPARTS = 1
)

// enum
type HYPERLINKSTATES int32

const (
	HLS_NORMALTEXT HYPERLINKSTATES = 1
	HLS_LINKTEXT   HYPERLINKSTATES = 2
)

// enum
type EMPTYMARKUPPARTS int32

const (
	EMP_MARKUPTEXT EMPTYMARKUPPARTS = 1
)

// enum
type MARKUPTEXTSTATES int32

const (
	EMT_NORMALTEXT MARKUPTEXTSTATES = 1
	EMT_LINKTEXT   MARKUPTEXTSTATES = 2
)

// enum
type STATICPARTS int32

const (
	STAT_TEXT STATICPARTS = 1
)

// enum
type PAGEPARTS int32

const (
	PGRP_UP       PAGEPARTS = 1
	PGRP_DOWN     PAGEPARTS = 2
	PGRP_UPHORZ   PAGEPARTS = 3
	PGRP_DOWNHORZ PAGEPARTS = 4
)

// enum
type MONTHCALPARTS int32

const (
	MC_BACKGROUND            MONTHCALPARTS = 1
	MC_BORDERS               MONTHCALPARTS = 2
	MC_GRIDBACKGROUND        MONTHCALPARTS = 3
	MC_COLHEADERSPLITTER     MONTHCALPARTS = 4
	MC_GRIDCELLBACKGROUND    MONTHCALPARTS = 5
	MC_GRIDCELL              MONTHCALPARTS = 6
	MC_GRIDCELLUPPER         MONTHCALPARTS = 7
	MC_TRAILINGGRIDCELL      MONTHCALPARTS = 8
	MC_TRAILINGGRIDCELLUPPER MONTHCALPARTS = 9
	MC_NAVNEXT               MONTHCALPARTS = 10
	MC_NAVPREV               MONTHCALPARTS = 11
)

// enum
type GRIDCELLBACKGROUNDSTATES int32

const (
	MCGCB_SELECTED           GRIDCELLBACKGROUNDSTATES = 1
	MCGCB_HOT                GRIDCELLBACKGROUNDSTATES = 2
	MCGCB_SELECTEDHOT        GRIDCELLBACKGROUNDSTATES = 3
	MCGCB_SELECTEDNOTFOCUSED GRIDCELLBACKGROUNDSTATES = 4
	MCGCB_TODAY              GRIDCELLBACKGROUNDSTATES = 5
	MCGCB_TODAYSELECTED      GRIDCELLBACKGROUNDSTATES = 6
)

// enum
type GRIDCELLSTATES int32

const (
	MCGC_HOT           GRIDCELLSTATES = 1
	MCGC_HASSTATE      GRIDCELLSTATES = 2
	MCGC_HASSTATEHOT   GRIDCELLSTATES = 3
	MCGC_TODAY         GRIDCELLSTATES = 4
	MCGC_TODAYSELECTED GRIDCELLSTATES = 5
	MCGC_SELECTED      GRIDCELLSTATES = 6
	MCGC_SELECTEDHOT   GRIDCELLSTATES = 7
)

// enum
type GRIDCELLUPPERSTATES int32

const (
	MCGCU_HOT         GRIDCELLUPPERSTATES = 1
	MCGCU_HASSTATE    GRIDCELLUPPERSTATES = 2
	MCGCU_HASSTATEHOT GRIDCELLUPPERSTATES = 3
	MCGCU_SELECTED    GRIDCELLUPPERSTATES = 4
	MCGCU_SELECTEDHOT GRIDCELLUPPERSTATES = 5
)

// enum
type TRAILINGGRIDCELLSTATES int32

const (
	MCTGC_HOT           TRAILINGGRIDCELLSTATES = 1
	MCTGC_HASSTATE      TRAILINGGRIDCELLSTATES = 2
	MCTGC_HASSTATEHOT   TRAILINGGRIDCELLSTATES = 3
	MCTGC_TODAY         TRAILINGGRIDCELLSTATES = 4
	MCTGC_TODAYSELECTED TRAILINGGRIDCELLSTATES = 5
	MCTGC_SELECTED      TRAILINGGRIDCELLSTATES = 6
	MCTGC_SELECTEDHOT   TRAILINGGRIDCELLSTATES = 7
)

// enum
type TRAILINGGRIDCELLUPPERSTATES int32

const (
	MCTGCU_HOT         TRAILINGGRIDCELLUPPERSTATES = 1
	MCTGCU_HASSTATE    TRAILINGGRIDCELLUPPERSTATES = 2
	MCTGCU_HASSTATEHOT TRAILINGGRIDCELLUPPERSTATES = 3
	MCTGCU_SELECTED    TRAILINGGRIDCELLUPPERSTATES = 4
	MCTGCU_SELECTEDHOT TRAILINGGRIDCELLUPPERSTATES = 5
)

// enum
type NAVNEXTSTATES int32

const (
	MCNN_NORMAL   NAVNEXTSTATES = 1
	MCNN_HOT      NAVNEXTSTATES = 2
	MCNN_PRESSED  NAVNEXTSTATES = 3
	MCNN_DISABLED NAVNEXTSTATES = 4
)

// enum
type NAVPREVSTATES int32

const (
	MCNP_NORMAL   NAVPREVSTATES = 1
	MCNP_HOT      NAVPREVSTATES = 2
	MCNP_PRESSED  NAVPREVSTATES = 3
	MCNP_DISABLED NAVPREVSTATES = 4
)

// enum
type CLOCKPARTS int32

const (
	CLP_TIME CLOCKPARTS = 1
)

// enum
type CLOCKSTATES int32

const (
	CLS_NORMAL  CLOCKSTATES = 1
	CLS_HOT     CLOCKSTATES = 2
	CLS_PRESSED CLOCKSTATES = 3
)

// enum
type TRAYNOTIFYPARTS int32

const (
	TNP_BACKGROUND     TRAYNOTIFYPARTS = 1
	TNP_ANIMBACKGROUND TRAYNOTIFYPARTS = 2
)

// enum
type TASKBARPARTS int32

const (
	TBP_BACKGROUNDBOTTOM TASKBARPARTS = 1
	TBP_BACKGROUNDRIGHT  TASKBARPARTS = 2
	TBP_BACKGROUNDTOP    TASKBARPARTS = 3
	TBP_BACKGROUNDLEFT   TASKBARPARTS = 4
	TBP_SIZINGBARBOTTOM  TASKBARPARTS = 5
	TBP_SIZINGBARRIGHT   TASKBARPARTS = 6
	TBP_SIZINGBARTOP     TASKBARPARTS = 7
	TBP_SIZINGBARLEFT    TASKBARPARTS = 8
)

// enum
type TASKBANDPARTS int32

const (
	TDP_GROUPCOUNT           TASKBANDPARTS = 1
	TDP_FLASHBUTTON          TASKBANDPARTS = 2
	TDP_FLASHBUTTONGROUPMENU TASKBANDPARTS = 3
)

// enum
type STARTPANELPARTS int32

const (
	SPP_USERPANE                  STARTPANELPARTS = 1
	SPP_MOREPROGRAMS              STARTPANELPARTS = 2
	SPP_MOREPROGRAMSARROW         STARTPANELPARTS = 3
	SPP_PROGLIST                  STARTPANELPARTS = 4
	SPP_PROGLISTSEPARATOR         STARTPANELPARTS = 5
	SPP_PLACESLIST                STARTPANELPARTS = 6
	SPP_PLACESLISTSEPARATOR       STARTPANELPARTS = 7
	SPP_LOGOFF                    STARTPANELPARTS = 8
	SPP_LOGOFFBUTTONS             STARTPANELPARTS = 9
	SPP_USERPICTURE               STARTPANELPARTS = 10
	SPP_PREVIEW                   STARTPANELPARTS = 11
	SPP_MOREPROGRAMSTAB           STARTPANELPARTS = 12
	SPP_NSCHOST                   STARTPANELPARTS = 13
	SPP_SOFTWAREEXPLORER          STARTPANELPARTS = 14
	SPP_OPENBOX                   STARTPANELPARTS = 15
	SPP_SEARCHVIEW                STARTPANELPARTS = 16
	SPP_MOREPROGRAMSARROWBACK     STARTPANELPARTS = 17
	SPP_TOPMATCH                  STARTPANELPARTS = 18
	SPP_LOGOFFSPLITBUTTONDROPDOWN STARTPANELPARTS = 19
)

// enum
type MOREPROGRAMSTABSTATES int32

const (
	SPMPT_NORMAL   MOREPROGRAMSTABSTATES = 1
	SPMPT_HOT      MOREPROGRAMSTABSTATES = 2
	SPMPT_SELECTED MOREPROGRAMSTABSTATES = 3
	SPMPT_DISABLED MOREPROGRAMSTABSTATES = 4
	SPMPT_FOCUSED  MOREPROGRAMSTABSTATES = 5
)

// enum
type SOFTWAREEXPLORERSTATES int32

const (
	SPSE_NORMAL   SOFTWAREEXPLORERSTATES = 1
	SPSE_HOT      SOFTWAREEXPLORERSTATES = 2
	SPSE_SELECTED SOFTWAREEXPLORERSTATES = 3
	SPSE_DISABLED SOFTWAREEXPLORERSTATES = 4
	SPSE_FOCUSED  SOFTWAREEXPLORERSTATES = 5
)

// enum
type OPENBOXSTATES int32

const (
	SPOB_NORMAL   OPENBOXSTATES = 1
	SPOB_HOT      OPENBOXSTATES = 2
	SPOB_SELECTED OPENBOXSTATES = 3
	SPOB_DISABLED OPENBOXSTATES = 4
	SPOB_FOCUSED  OPENBOXSTATES = 5
)

// enum
type MOREPROGRAMSARROWSTATES int32

const (
	SPS_NORMAL  MOREPROGRAMSARROWSTATES = 1
	SPS_HOT     MOREPROGRAMSARROWSTATES = 2
	SPS_PRESSED MOREPROGRAMSARROWSTATES = 3
)

// enum
type MOREPROGRAMSARROWBACKSTATES int32

const (
	SPSB_NORMAL  MOREPROGRAMSARROWBACKSTATES = 1
	SPSB_HOT     MOREPROGRAMSARROWBACKSTATES = 2
	SPSB_PRESSED MOREPROGRAMSARROWBACKSTATES = 3
)

// enum
type LOGOFFBUTTONSSTATES int32

const (
	SPLS_NORMAL  LOGOFFBUTTONSSTATES = 1
	SPLS_HOT     LOGOFFBUTTONSSTATES = 2
	SPLS_PRESSED LOGOFFBUTTONSSTATES = 3
)

// enum
type MENUBANDPARTS int32

const (
	MDP_NEWAPPBUTTON MENUBANDPARTS = 1
	MDP_SEPERATOR    MENUBANDPARTS = 2
)

// enum
type MENUBANDSTATES int32

const (
	MDS_NORMAL     MENUBANDSTATES = 1
	MDS_HOT        MENUBANDSTATES = 2
	MDS_PRESSED    MENUBANDSTATES = 3
	MDS_DISABLED   MENUBANDSTATES = 4
	MDS_CHECKED    MENUBANDSTATES = 5
	MDS_HOTCHECKED MENUBANDSTATES = 6
)

// enum
type POINTER_FEEDBACK_MODE int32

const (
	POINTER_FEEDBACK_DEFAULT  POINTER_FEEDBACK_MODE = 1
	POINTER_FEEDBACK_INDIRECT POINTER_FEEDBACK_MODE = 2
	POINTER_FEEDBACK_NONE     POINTER_FEEDBACK_MODE = 3
)

// enum
type FEEDBACK_TYPE int32

const (
	FEEDBACK_TOUCH_CONTACTVISUALIZATION FEEDBACK_TYPE = 1
	FEEDBACK_PEN_BARRELVISUALIZATION    FEEDBACK_TYPE = 2
	FEEDBACK_PEN_TAP                    FEEDBACK_TYPE = 3
	FEEDBACK_PEN_DOUBLETAP              FEEDBACK_TYPE = 4
	FEEDBACK_PEN_PRESSANDHOLD           FEEDBACK_TYPE = 5
	FEEDBACK_PEN_RIGHTTAP               FEEDBACK_TYPE = 6
	FEEDBACK_TOUCH_TAP                  FEEDBACK_TYPE = 7
	FEEDBACK_TOUCH_DOUBLETAP            FEEDBACK_TYPE = 8
	FEEDBACK_TOUCH_PRESSANDHOLD         FEEDBACK_TYPE = 9
	FEEDBACK_TOUCH_RIGHTTAP             FEEDBACK_TYPE = 10
	FEEDBACK_GESTURE_PRESSANDTAP        FEEDBACK_TYPE = 11
	FEEDBACK_MAX                        FEEDBACK_TYPE = -1
)

// enum
type POINTER_DEVICE_TYPE int32

const (
	POINTER_DEVICE_TYPE_INTEGRATED_PEN POINTER_DEVICE_TYPE = 1
	POINTER_DEVICE_TYPE_EXTERNAL_PEN   POINTER_DEVICE_TYPE = 2
	POINTER_DEVICE_TYPE_TOUCH          POINTER_DEVICE_TYPE = 3
	POINTER_DEVICE_TYPE_TOUCH_PAD      POINTER_DEVICE_TYPE = 4
	POINTER_DEVICE_TYPE_MAX            POINTER_DEVICE_TYPE = -1
)

// enum
type POINTER_DEVICE_CURSOR_TYPE int32

const (
	POINTER_DEVICE_CURSOR_TYPE_UNKNOWN POINTER_DEVICE_CURSOR_TYPE = 0
	POINTER_DEVICE_CURSOR_TYPE_TIP     POINTER_DEVICE_CURSOR_TYPE = 1
	POINTER_DEVICE_CURSOR_TYPE_ERASER  POINTER_DEVICE_CURSOR_TYPE = 2
	POINTER_DEVICE_CURSOR_TYPE_MAX     POINTER_DEVICE_CURSOR_TYPE = -1
)

// structs

type PROPSHEETPAGEA_V1_Anonymous1 struct {
	Data [1]uint64
}

func (this *PROPSHEETPAGEA_V1_Anonymous1) PszTemplate() *PSTR {
	return (*PSTR)(unsafe.Pointer(this))
}

func (this *PROPSHEETPAGEA_V1_Anonymous1) PszTemplateVal() PSTR {
	return *(*PSTR)(unsafe.Pointer(this))
}

func (this *PROPSHEETPAGEA_V1_Anonymous1) PResource() **DLGTEMPLATE {
	return (**DLGTEMPLATE)(unsafe.Pointer(this))
}

func (this *PROPSHEETPAGEA_V1_Anonymous1) PResourceVal() *DLGTEMPLATE {
	return *(**DLGTEMPLATE)(unsafe.Pointer(this))
}

type PROPSHEETPAGEA_V1_Anonymous2 struct {
	Data [1]uint64
}

func (this *PROPSHEETPAGEA_V1_Anonymous2) HIcon() *HICON {
	return (*HICON)(unsafe.Pointer(this))
}

func (this *PROPSHEETPAGEA_V1_Anonymous2) HIconVal() HICON {
	return *(*HICON)(unsafe.Pointer(this))
}

func (this *PROPSHEETPAGEA_V1_Anonymous2) PszIcon() *PSTR {
	return (*PSTR)(unsafe.Pointer(this))
}

func (this *PROPSHEETPAGEA_V1_Anonymous2) PszIconVal() PSTR {
	return *(*PSTR)(unsafe.Pointer(this))
}

type PROPSHEETPAGEA_V1 struct {
	DwSize    uint32
	DwFlags   uint32
	HInstance HINSTANCE
	PROPSHEETPAGEA_V1_Anonymous1
	PROPSHEETPAGEA_V1_Anonymous2
	PszTitle    PSTR
	PfnDlgProc  DLGPROC
	LParam      LPARAM
	PfnCallback LPFNPSPCALLBACKA
	PcRefParent *uint32
}

type PROPSHEETPAGEA_V2_Anonymous1 struct {
	Data [1]uint64
}

func (this *PROPSHEETPAGEA_V2_Anonymous1) PszTemplate() *PSTR {
	return (*PSTR)(unsafe.Pointer(this))
}

func (this *PROPSHEETPAGEA_V2_Anonymous1) PszTemplateVal() PSTR {
	return *(*PSTR)(unsafe.Pointer(this))
}

func (this *PROPSHEETPAGEA_V2_Anonymous1) PResource() **DLGTEMPLATE {
	return (**DLGTEMPLATE)(unsafe.Pointer(this))
}

func (this *PROPSHEETPAGEA_V2_Anonymous1) PResourceVal() *DLGTEMPLATE {
	return *(**DLGTEMPLATE)(unsafe.Pointer(this))
}

type PROPSHEETPAGEA_V2_Anonymous2 struct {
	Data [1]uint64
}

func (this *PROPSHEETPAGEA_V2_Anonymous2) HIcon() *HICON {
	return (*HICON)(unsafe.Pointer(this))
}

func (this *PROPSHEETPAGEA_V2_Anonymous2) HIconVal() HICON {
	return *(*HICON)(unsafe.Pointer(this))
}

func (this *PROPSHEETPAGEA_V2_Anonymous2) PszIcon() *PSTR {
	return (*PSTR)(unsafe.Pointer(this))
}

func (this *PROPSHEETPAGEA_V2_Anonymous2) PszIconVal() PSTR {
	return *(*PSTR)(unsafe.Pointer(this))
}

type PROPSHEETPAGEA_V2 struct {
	DwSize    uint32
	DwFlags   uint32
	HInstance HINSTANCE
	PROPSHEETPAGEA_V2_Anonymous1
	PROPSHEETPAGEA_V2_Anonymous2
	PszTitle          PSTR
	PfnDlgProc        DLGPROC
	LParam            LPARAM
	PfnCallback       LPFNPSPCALLBACKA
	PcRefParent       *uint32
	PszHeaderTitle    PSTR
	PszHeaderSubTitle PSTR
}

type PROPSHEETPAGEA_V3_Anonymous1 struct {
	Data [1]uint64
}

func (this *PROPSHEETPAGEA_V3_Anonymous1) PszTemplate() *PSTR {
	return (*PSTR)(unsafe.Pointer(this))
}

func (this *PROPSHEETPAGEA_V3_Anonymous1) PszTemplateVal() PSTR {
	return *(*PSTR)(unsafe.Pointer(this))
}

func (this *PROPSHEETPAGEA_V3_Anonymous1) PResource() **DLGTEMPLATE {
	return (**DLGTEMPLATE)(unsafe.Pointer(this))
}

func (this *PROPSHEETPAGEA_V3_Anonymous1) PResourceVal() *DLGTEMPLATE {
	return *(**DLGTEMPLATE)(unsafe.Pointer(this))
}

type PROPSHEETPAGEA_V3_Anonymous2 struct {
	Data [1]uint64
}

func (this *PROPSHEETPAGEA_V3_Anonymous2) HIcon() *HICON {
	return (*HICON)(unsafe.Pointer(this))
}

func (this *PROPSHEETPAGEA_V3_Anonymous2) HIconVal() HICON {
	return *(*HICON)(unsafe.Pointer(this))
}

func (this *PROPSHEETPAGEA_V3_Anonymous2) PszIcon() *PSTR {
	return (*PSTR)(unsafe.Pointer(this))
}

func (this *PROPSHEETPAGEA_V3_Anonymous2) PszIconVal() PSTR {
	return *(*PSTR)(unsafe.Pointer(this))
}

type PROPSHEETPAGEA_V3 struct {
	DwSize    uint32
	DwFlags   uint32
	HInstance HINSTANCE
	PROPSHEETPAGEA_V3_Anonymous1
	PROPSHEETPAGEA_V3_Anonymous2
	PszTitle          PSTR
	PfnDlgProc        DLGPROC
	LParam            LPARAM
	PfnCallback       LPFNPSPCALLBACKA
	PcRefParent       *uint32
	PszHeaderTitle    PSTR
	PszHeaderSubTitle PSTR
	HActCtx           HANDLE
}

type PROPSHEETPAGEA_Anonymous1 struct {
	Data [1]uint64
}

func (this *PROPSHEETPAGEA_Anonymous1) PszTemplate() *PSTR {
	return (*PSTR)(unsafe.Pointer(this))
}

func (this *PROPSHEETPAGEA_Anonymous1) PszTemplateVal() PSTR {
	return *(*PSTR)(unsafe.Pointer(this))
}

func (this *PROPSHEETPAGEA_Anonymous1) PResource() **DLGTEMPLATE {
	return (**DLGTEMPLATE)(unsafe.Pointer(this))
}

func (this *PROPSHEETPAGEA_Anonymous1) PResourceVal() *DLGTEMPLATE {
	return *(**DLGTEMPLATE)(unsafe.Pointer(this))
}

type PROPSHEETPAGEA_Anonymous2 struct {
	Data [1]uint64
}

func (this *PROPSHEETPAGEA_Anonymous2) HIcon() *HICON {
	return (*HICON)(unsafe.Pointer(this))
}

func (this *PROPSHEETPAGEA_Anonymous2) HIconVal() HICON {
	return *(*HICON)(unsafe.Pointer(this))
}

func (this *PROPSHEETPAGEA_Anonymous2) PszIcon() *PSTR {
	return (*PSTR)(unsafe.Pointer(this))
}

func (this *PROPSHEETPAGEA_Anonymous2) PszIconVal() PSTR {
	return *(*PSTR)(unsafe.Pointer(this))
}

type PROPSHEETPAGEA_Anonymous3 struct {
	Data [1]uint64
}

func (this *PROPSHEETPAGEA_Anonymous3) HbmHeader() *HBITMAP {
	return (*HBITMAP)(unsafe.Pointer(this))
}

func (this *PROPSHEETPAGEA_Anonymous3) HbmHeaderVal() HBITMAP {
	return *(*HBITMAP)(unsafe.Pointer(this))
}

func (this *PROPSHEETPAGEA_Anonymous3) PszbmHeader() *PSTR {
	return (*PSTR)(unsafe.Pointer(this))
}

func (this *PROPSHEETPAGEA_Anonymous3) PszbmHeaderVal() PSTR {
	return *(*PSTR)(unsafe.Pointer(this))
}

type PROPSHEETPAGEA struct {
	DwSize    uint32
	DwFlags   uint32
	HInstance HINSTANCE
	PROPSHEETPAGEA_Anonymous1
	PROPSHEETPAGEA_Anonymous2
	PszTitle          PSTR
	PfnDlgProc        DLGPROC
	LParam            LPARAM
	PfnCallback       LPFNPSPCALLBACKA
	PcRefParent       *uint32
	PszHeaderTitle    PSTR
	PszHeaderSubTitle PSTR
	HActCtx           HANDLE
	PROPSHEETPAGEA_Anonymous3
}

type PROPSHEETPAGEW_V1_Anonymous1 struct {
	Data [1]uint64
}

func (this *PROPSHEETPAGEW_V1_Anonymous1) PszTemplate() *PWSTR {
	return (*PWSTR)(unsafe.Pointer(this))
}

func (this *PROPSHEETPAGEW_V1_Anonymous1) PszTemplateVal() PWSTR {
	return *(*PWSTR)(unsafe.Pointer(this))
}

func (this *PROPSHEETPAGEW_V1_Anonymous1) PResource() **DLGTEMPLATE {
	return (**DLGTEMPLATE)(unsafe.Pointer(this))
}

func (this *PROPSHEETPAGEW_V1_Anonymous1) PResourceVal() *DLGTEMPLATE {
	return *(**DLGTEMPLATE)(unsafe.Pointer(this))
}

type PROPSHEETPAGEW_V1_Anonymous2 struct {
	Data [1]uint64
}

func (this *PROPSHEETPAGEW_V1_Anonymous2) HIcon() *HICON {
	return (*HICON)(unsafe.Pointer(this))
}

func (this *PROPSHEETPAGEW_V1_Anonymous2) HIconVal() HICON {
	return *(*HICON)(unsafe.Pointer(this))
}

func (this *PROPSHEETPAGEW_V1_Anonymous2) PszIcon() *PWSTR {
	return (*PWSTR)(unsafe.Pointer(this))
}

func (this *PROPSHEETPAGEW_V1_Anonymous2) PszIconVal() PWSTR {
	return *(*PWSTR)(unsafe.Pointer(this))
}

type PROPSHEETPAGEW_V1 struct {
	DwSize    uint32
	DwFlags   uint32
	HInstance HINSTANCE
	PROPSHEETPAGEW_V1_Anonymous1
	PROPSHEETPAGEW_V1_Anonymous2
	PszTitle    PWSTR
	PfnDlgProc  DLGPROC
	LParam      LPARAM
	PfnCallback LPFNPSPCALLBACKW
	PcRefParent *uint32
}

type PROPSHEETPAGEW_V2_Anonymous1 struct {
	Data [1]uint64
}

func (this *PROPSHEETPAGEW_V2_Anonymous1) PszTemplate() *PWSTR {
	return (*PWSTR)(unsafe.Pointer(this))
}

func (this *PROPSHEETPAGEW_V2_Anonymous1) PszTemplateVal() PWSTR {
	return *(*PWSTR)(unsafe.Pointer(this))
}

func (this *PROPSHEETPAGEW_V2_Anonymous1) PResource() **DLGTEMPLATE {
	return (**DLGTEMPLATE)(unsafe.Pointer(this))
}

func (this *PROPSHEETPAGEW_V2_Anonymous1) PResourceVal() *DLGTEMPLATE {
	return *(**DLGTEMPLATE)(unsafe.Pointer(this))
}

type PROPSHEETPAGEW_V2_Anonymous2 struct {
	Data [1]uint64
}

func (this *PROPSHEETPAGEW_V2_Anonymous2) HIcon() *HICON {
	return (*HICON)(unsafe.Pointer(this))
}

func (this *PROPSHEETPAGEW_V2_Anonymous2) HIconVal() HICON {
	return *(*HICON)(unsafe.Pointer(this))
}

func (this *PROPSHEETPAGEW_V2_Anonymous2) PszIcon() *PWSTR {
	return (*PWSTR)(unsafe.Pointer(this))
}

func (this *PROPSHEETPAGEW_V2_Anonymous2) PszIconVal() PWSTR {
	return *(*PWSTR)(unsafe.Pointer(this))
}

type PROPSHEETPAGEW_V2 struct {
	DwSize    uint32
	DwFlags   uint32
	HInstance HINSTANCE
	PROPSHEETPAGEW_V2_Anonymous1
	PROPSHEETPAGEW_V2_Anonymous2
	PszTitle          PWSTR
	PfnDlgProc        DLGPROC
	LParam            LPARAM
	PfnCallback       LPFNPSPCALLBACKW
	PcRefParent       *uint32
	PszHeaderTitle    PWSTR
	PszHeaderSubTitle PWSTR
}

type PROPSHEETPAGEW_V3_Anonymous1 struct {
	Data [1]uint64
}

func (this *PROPSHEETPAGEW_V3_Anonymous1) PszTemplate() *PWSTR {
	return (*PWSTR)(unsafe.Pointer(this))
}

func (this *PROPSHEETPAGEW_V3_Anonymous1) PszTemplateVal() PWSTR {
	return *(*PWSTR)(unsafe.Pointer(this))
}

func (this *PROPSHEETPAGEW_V3_Anonymous1) PResource() **DLGTEMPLATE {
	return (**DLGTEMPLATE)(unsafe.Pointer(this))
}

func (this *PROPSHEETPAGEW_V3_Anonymous1) PResourceVal() *DLGTEMPLATE {
	return *(**DLGTEMPLATE)(unsafe.Pointer(this))
}

type PROPSHEETPAGEW_V3_Anonymous2 struct {
	Data [1]uint64
}

func (this *PROPSHEETPAGEW_V3_Anonymous2) HIcon() *HICON {
	return (*HICON)(unsafe.Pointer(this))
}

func (this *PROPSHEETPAGEW_V3_Anonymous2) HIconVal() HICON {
	return *(*HICON)(unsafe.Pointer(this))
}

func (this *PROPSHEETPAGEW_V3_Anonymous2) PszIcon() *PWSTR {
	return (*PWSTR)(unsafe.Pointer(this))
}

func (this *PROPSHEETPAGEW_V3_Anonymous2) PszIconVal() PWSTR {
	return *(*PWSTR)(unsafe.Pointer(this))
}

type PROPSHEETPAGEW_V3 struct {
	DwSize    uint32
	DwFlags   uint32
	HInstance HINSTANCE
	PROPSHEETPAGEW_V3_Anonymous1
	PROPSHEETPAGEW_V3_Anonymous2
	PszTitle          PWSTR
	PfnDlgProc        DLGPROC
	LParam            LPARAM
	PfnCallback       LPFNPSPCALLBACKW
	PcRefParent       *uint32
	PszHeaderTitle    PWSTR
	PszHeaderSubTitle PWSTR
	HActCtx           HANDLE
}

type PROPSHEETPAGEW_Anonymous1 struct {
	Data [1]uint64
}

func (this *PROPSHEETPAGEW_Anonymous1) PszTemplate() *PWSTR {
	return (*PWSTR)(unsafe.Pointer(this))
}

func (this *PROPSHEETPAGEW_Anonymous1) PszTemplateVal() PWSTR {
	return *(*PWSTR)(unsafe.Pointer(this))
}

func (this *PROPSHEETPAGEW_Anonymous1) PResource() **DLGTEMPLATE {
	return (**DLGTEMPLATE)(unsafe.Pointer(this))
}

func (this *PROPSHEETPAGEW_Anonymous1) PResourceVal() *DLGTEMPLATE {
	return *(**DLGTEMPLATE)(unsafe.Pointer(this))
}

type PROPSHEETPAGEW_Anonymous2 struct {
	Data [1]uint64
}

func (this *PROPSHEETPAGEW_Anonymous2) HIcon() *HICON {
	return (*HICON)(unsafe.Pointer(this))
}

func (this *PROPSHEETPAGEW_Anonymous2) HIconVal() HICON {
	return *(*HICON)(unsafe.Pointer(this))
}

func (this *PROPSHEETPAGEW_Anonymous2) PszIcon() *PWSTR {
	return (*PWSTR)(unsafe.Pointer(this))
}

func (this *PROPSHEETPAGEW_Anonymous2) PszIconVal() PWSTR {
	return *(*PWSTR)(unsafe.Pointer(this))
}

type PROPSHEETPAGEW_Anonymous3 struct {
	Data [1]uint64
}

func (this *PROPSHEETPAGEW_Anonymous3) HbmHeader() *HBITMAP {
	return (*HBITMAP)(unsafe.Pointer(this))
}

func (this *PROPSHEETPAGEW_Anonymous3) HbmHeaderVal() HBITMAP {
	return *(*HBITMAP)(unsafe.Pointer(this))
}

func (this *PROPSHEETPAGEW_Anonymous3) PszbmHeader() *PWSTR {
	return (*PWSTR)(unsafe.Pointer(this))
}

func (this *PROPSHEETPAGEW_Anonymous3) PszbmHeaderVal() PWSTR {
	return *(*PWSTR)(unsafe.Pointer(this))
}

type PROPSHEETPAGE = PROPSHEETPAGEW
type PROPSHEETPAGEW struct {
	DwSize    uint32
	DwFlags   uint32
	HInstance HINSTANCE
	PROPSHEETPAGEW_Anonymous1
	PROPSHEETPAGEW_Anonymous2
	PszTitle          PWSTR
	PfnDlgProc        DLGPROC
	LParam            LPARAM
	PfnCallback       LPFNPSPCALLBACKW
	PcRefParent       *uint32
	PszHeaderTitle    PWSTR
	PszHeaderSubTitle PWSTR
	HActCtx           HANDLE
	PROPSHEETPAGEW_Anonymous3
}

type PROPSHEETHEADERA_V1_Anonymous1 struct {
	Data [1]uint64
}

func (this *PROPSHEETHEADERA_V1_Anonymous1) HIcon() *HICON {
	return (*HICON)(unsafe.Pointer(this))
}

func (this *PROPSHEETHEADERA_V1_Anonymous1) HIconVal() HICON {
	return *(*HICON)(unsafe.Pointer(this))
}

func (this *PROPSHEETHEADERA_V1_Anonymous1) PszIcon() *PSTR {
	return (*PSTR)(unsafe.Pointer(this))
}

func (this *PROPSHEETHEADERA_V1_Anonymous1) PszIconVal() PSTR {
	return *(*PSTR)(unsafe.Pointer(this))
}

type PROPSHEETHEADERA_V1_Anonymous2 struct {
	Data [1]uint64
}

func (this *PROPSHEETHEADERA_V1_Anonymous2) NStartPage() *uint32 {
	return (*uint32)(unsafe.Pointer(this))
}

func (this *PROPSHEETHEADERA_V1_Anonymous2) NStartPageVal() uint32 {
	return *(*uint32)(unsafe.Pointer(this))
}

func (this *PROPSHEETHEADERA_V1_Anonymous2) PStartPage() *PSTR {
	return (*PSTR)(unsafe.Pointer(this))
}

func (this *PROPSHEETHEADERA_V1_Anonymous2) PStartPageVal() PSTR {
	return *(*PSTR)(unsafe.Pointer(this))
}

type PROPSHEETHEADERA_V1_Anonymous3 struct {
	Data [1]uint64
}

func (this *PROPSHEETHEADERA_V1_Anonymous3) Ppsp() **PROPSHEETPAGEA {
	return (**PROPSHEETPAGEA)(unsafe.Pointer(this))
}

func (this *PROPSHEETHEADERA_V1_Anonymous3) PpspVal() *PROPSHEETPAGEA {
	return *(**PROPSHEETPAGEA)(unsafe.Pointer(this))
}

func (this *PROPSHEETHEADERA_V1_Anonymous3) Phpage() **HPROPSHEETPAGE {
	return (**HPROPSHEETPAGE)(unsafe.Pointer(this))
}

func (this *PROPSHEETHEADERA_V1_Anonymous3) PhpageVal() *HPROPSHEETPAGE {
	return *(**HPROPSHEETPAGE)(unsafe.Pointer(this))
}

type PROPSHEETHEADERA_V1 struct {
	DwSize     uint32
	DwFlags    uint32
	HwndParent HWND
	HInstance  HINSTANCE
	PROPSHEETHEADERA_V1_Anonymous1
	PszCaption PSTR
	NPages     uint32
	PROPSHEETHEADERA_V1_Anonymous2
	PROPSHEETHEADERA_V1_Anonymous3
	PfnCallback PFNPROPSHEETCALLBACK
}

type PROPSHEETHEADERA_V2_Anonymous1 struct {
	Data [1]uint64
}

func (this *PROPSHEETHEADERA_V2_Anonymous1) HIcon() *HICON {
	return (*HICON)(unsafe.Pointer(this))
}

func (this *PROPSHEETHEADERA_V2_Anonymous1) HIconVal() HICON {
	return *(*HICON)(unsafe.Pointer(this))
}

func (this *PROPSHEETHEADERA_V2_Anonymous1) PszIcon() *PSTR {
	return (*PSTR)(unsafe.Pointer(this))
}

func (this *PROPSHEETHEADERA_V2_Anonymous1) PszIconVal() PSTR {
	return *(*PSTR)(unsafe.Pointer(this))
}

type PROPSHEETHEADERA_V2_Anonymous2 struct {
	Data [1]uint64
}

func (this *PROPSHEETHEADERA_V2_Anonymous2) NStartPage() *uint32 {
	return (*uint32)(unsafe.Pointer(this))
}

func (this *PROPSHEETHEADERA_V2_Anonymous2) NStartPageVal() uint32 {
	return *(*uint32)(unsafe.Pointer(this))
}

func (this *PROPSHEETHEADERA_V2_Anonymous2) PStartPage() *PSTR {
	return (*PSTR)(unsafe.Pointer(this))
}

func (this *PROPSHEETHEADERA_V2_Anonymous2) PStartPageVal() PSTR {
	return *(*PSTR)(unsafe.Pointer(this))
}

type PROPSHEETHEADERA_V2_Anonymous3 struct {
	Data [1]uint64
}

func (this *PROPSHEETHEADERA_V2_Anonymous3) Ppsp() **PROPSHEETPAGEA {
	return (**PROPSHEETPAGEA)(unsafe.Pointer(this))
}

func (this *PROPSHEETHEADERA_V2_Anonymous3) PpspVal() *PROPSHEETPAGEA {
	return *(**PROPSHEETPAGEA)(unsafe.Pointer(this))
}

func (this *PROPSHEETHEADERA_V2_Anonymous3) Phpage() **HPROPSHEETPAGE {
	return (**HPROPSHEETPAGE)(unsafe.Pointer(this))
}

func (this *PROPSHEETHEADERA_V2_Anonymous3) PhpageVal() *HPROPSHEETPAGE {
	return *(**HPROPSHEETPAGE)(unsafe.Pointer(this))
}

type PROPSHEETHEADERA_V2_Anonymous4 struct {
	Data [1]uint64
}

func (this *PROPSHEETHEADERA_V2_Anonymous4) HbmWatermark() *HBITMAP {
	return (*HBITMAP)(unsafe.Pointer(this))
}

func (this *PROPSHEETHEADERA_V2_Anonymous4) HbmWatermarkVal() HBITMAP {
	return *(*HBITMAP)(unsafe.Pointer(this))
}

func (this *PROPSHEETHEADERA_V2_Anonymous4) PszbmWatermark() *PSTR {
	return (*PSTR)(unsafe.Pointer(this))
}

func (this *PROPSHEETHEADERA_V2_Anonymous4) PszbmWatermarkVal() PSTR {
	return *(*PSTR)(unsafe.Pointer(this))
}

type PROPSHEETHEADERA_V2_Anonymous5 struct {
	Data [1]uint64
}

func (this *PROPSHEETHEADERA_V2_Anonymous5) HbmHeader() *HBITMAP {
	return (*HBITMAP)(unsafe.Pointer(this))
}

func (this *PROPSHEETHEADERA_V2_Anonymous5) HbmHeaderVal() HBITMAP {
	return *(*HBITMAP)(unsafe.Pointer(this))
}

func (this *PROPSHEETHEADERA_V2_Anonymous5) PszbmHeader() *PSTR {
	return (*PSTR)(unsafe.Pointer(this))
}

func (this *PROPSHEETHEADERA_V2_Anonymous5) PszbmHeaderVal() PSTR {
	return *(*PSTR)(unsafe.Pointer(this))
}

type PROPSHEETHEADERA_V2 struct {
	DwSize     uint32
	DwFlags    uint32
	HwndParent HWND
	HInstance  HINSTANCE
	PROPSHEETHEADERA_V2_Anonymous1
	PszCaption PSTR
	NPages     uint32
	PROPSHEETHEADERA_V2_Anonymous2
	PROPSHEETHEADERA_V2_Anonymous3
	PfnCallback PFNPROPSHEETCALLBACK
	PROPSHEETHEADERA_V2_Anonymous4
	HplWatermark HPALETTE
	PROPSHEETHEADERA_V2_Anonymous5
}

type PROPSHEETHEADERW_V1_Anonymous1 struct {
	Data [1]uint64
}

func (this *PROPSHEETHEADERW_V1_Anonymous1) HIcon() *HICON {
	return (*HICON)(unsafe.Pointer(this))
}

func (this *PROPSHEETHEADERW_V1_Anonymous1) HIconVal() HICON {
	return *(*HICON)(unsafe.Pointer(this))
}

func (this *PROPSHEETHEADERW_V1_Anonymous1) PszIcon() *PWSTR {
	return (*PWSTR)(unsafe.Pointer(this))
}

func (this *PROPSHEETHEADERW_V1_Anonymous1) PszIconVal() PWSTR {
	return *(*PWSTR)(unsafe.Pointer(this))
}

type PROPSHEETHEADERW_V1_Anonymous2 struct {
	Data [1]uint64
}

func (this *PROPSHEETHEADERW_V1_Anonymous2) NStartPage() *uint32 {
	return (*uint32)(unsafe.Pointer(this))
}

func (this *PROPSHEETHEADERW_V1_Anonymous2) NStartPageVal() uint32 {
	return *(*uint32)(unsafe.Pointer(this))
}

func (this *PROPSHEETHEADERW_V1_Anonymous2) PStartPage() *PWSTR {
	return (*PWSTR)(unsafe.Pointer(this))
}

func (this *PROPSHEETHEADERW_V1_Anonymous2) PStartPageVal() PWSTR {
	return *(*PWSTR)(unsafe.Pointer(this))
}

type PROPSHEETHEADERW_V1_Anonymous3 struct {
	Data [1]uint64
}

func (this *PROPSHEETHEADERW_V1_Anonymous3) Ppsp() **PROPSHEETPAGEW {
	return (**PROPSHEETPAGEW)(unsafe.Pointer(this))
}

func (this *PROPSHEETHEADERW_V1_Anonymous3) PpspVal() *PROPSHEETPAGEW {
	return *(**PROPSHEETPAGEW)(unsafe.Pointer(this))
}

func (this *PROPSHEETHEADERW_V1_Anonymous3) Phpage() **HPROPSHEETPAGE {
	return (**HPROPSHEETPAGE)(unsafe.Pointer(this))
}

func (this *PROPSHEETHEADERW_V1_Anonymous3) PhpageVal() *HPROPSHEETPAGE {
	return *(**HPROPSHEETPAGE)(unsafe.Pointer(this))
}

type PROPSHEETHEADERW_V1 struct {
	DwSize     uint32
	DwFlags    uint32
	HwndParent HWND
	HInstance  HINSTANCE
	PROPSHEETHEADERW_V1_Anonymous1
	PszCaption PWSTR
	NPages     uint32
	PROPSHEETHEADERW_V1_Anonymous2
	PROPSHEETHEADERW_V1_Anonymous3
	PfnCallback PFNPROPSHEETCALLBACK
}

type PROPSHEETHEADERW_V2_Anonymous1 struct {
	Data [1]uint64
}

func (this *PROPSHEETHEADERW_V2_Anonymous1) HIcon() *HICON {
	return (*HICON)(unsafe.Pointer(this))
}

func (this *PROPSHEETHEADERW_V2_Anonymous1) HIconVal() HICON {
	return *(*HICON)(unsafe.Pointer(this))
}

func (this *PROPSHEETHEADERW_V2_Anonymous1) PszIcon() *PWSTR {
	return (*PWSTR)(unsafe.Pointer(this))
}

func (this *PROPSHEETHEADERW_V2_Anonymous1) PszIconVal() PWSTR {
	return *(*PWSTR)(unsafe.Pointer(this))
}

type PROPSHEETHEADERW_V2_Anonymous2 struct {
	Data [1]uint64
}

func (this *PROPSHEETHEADERW_V2_Anonymous2) NStartPage() *uint32 {
	return (*uint32)(unsafe.Pointer(this))
}

func (this *PROPSHEETHEADERW_V2_Anonymous2) NStartPageVal() uint32 {
	return *(*uint32)(unsafe.Pointer(this))
}

func (this *PROPSHEETHEADERW_V2_Anonymous2) PStartPage() *PWSTR {
	return (*PWSTR)(unsafe.Pointer(this))
}

func (this *PROPSHEETHEADERW_V2_Anonymous2) PStartPageVal() PWSTR {
	return *(*PWSTR)(unsafe.Pointer(this))
}

type PROPSHEETHEADERW_V2_Anonymous3 struct {
	Data [1]uint64
}

func (this *PROPSHEETHEADERW_V2_Anonymous3) Ppsp() **PROPSHEETPAGEW {
	return (**PROPSHEETPAGEW)(unsafe.Pointer(this))
}

func (this *PROPSHEETHEADERW_V2_Anonymous3) PpspVal() *PROPSHEETPAGEW {
	return *(**PROPSHEETPAGEW)(unsafe.Pointer(this))
}

func (this *PROPSHEETHEADERW_V2_Anonymous3) Phpage() **HPROPSHEETPAGE {
	return (**HPROPSHEETPAGE)(unsafe.Pointer(this))
}

func (this *PROPSHEETHEADERW_V2_Anonymous3) PhpageVal() *HPROPSHEETPAGE {
	return *(**HPROPSHEETPAGE)(unsafe.Pointer(this))
}

type PROPSHEETHEADERW_V2_Anonymous4 struct {
	Data [1]uint64
}

func (this *PROPSHEETHEADERW_V2_Anonymous4) HbmWatermark() *HBITMAP {
	return (*HBITMAP)(unsafe.Pointer(this))
}

func (this *PROPSHEETHEADERW_V2_Anonymous4) HbmWatermarkVal() HBITMAP {
	return *(*HBITMAP)(unsafe.Pointer(this))
}

func (this *PROPSHEETHEADERW_V2_Anonymous4) PszbmWatermark() *PWSTR {
	return (*PWSTR)(unsafe.Pointer(this))
}

func (this *PROPSHEETHEADERW_V2_Anonymous4) PszbmWatermarkVal() PWSTR {
	return *(*PWSTR)(unsafe.Pointer(this))
}

type PROPSHEETHEADERW_V2_Anonymous5 struct {
	Data [1]uint64
}

func (this *PROPSHEETHEADERW_V2_Anonymous5) HbmHeader() *HBITMAP {
	return (*HBITMAP)(unsafe.Pointer(this))
}

func (this *PROPSHEETHEADERW_V2_Anonymous5) HbmHeaderVal() HBITMAP {
	return *(*HBITMAP)(unsafe.Pointer(this))
}

func (this *PROPSHEETHEADERW_V2_Anonymous5) PszbmHeader() *PWSTR {
	return (*PWSTR)(unsafe.Pointer(this))
}

func (this *PROPSHEETHEADERW_V2_Anonymous5) PszbmHeaderVal() PWSTR {
	return *(*PWSTR)(unsafe.Pointer(this))
}

type PROPSHEETHEADERW_V2 struct {
	DwSize     uint32
	DwFlags    uint32
	HwndParent HWND
	HInstance  HINSTANCE
	PROPSHEETHEADERW_V2_Anonymous1
	PszCaption PWSTR
	NPages     uint32
	PROPSHEETHEADERW_V2_Anonymous2
	PROPSHEETHEADERW_V2_Anonymous3
	PfnCallback PFNPROPSHEETCALLBACK
	PROPSHEETHEADERW_V2_Anonymous4
	HplWatermark HPALETTE
	PROPSHEETHEADERW_V2_Anonymous5
}

type PSHNOTIFY struct {
	Hdr    NMHDR
	LParam LPARAM
}

type INITCOMMONCONTROLSEX struct {
	DwSize uint32
	DwICC  INITCOMMONCONTROLSEX_ICC
}

type COLORSCHEME struct {
	DwSize          uint32
	ClrBtnHighlight uint32
	ClrBtnShadow    uint32
}

type NMTOOLTIPSCREATED struct {
	Hdr          NMHDR
	HwndToolTips HWND
}

type NMMOUSE struct {
	Hdr        NMHDR
	DwItemSpec uintptr
	DwItemData uintptr
	Pt         POINT
	DwHitInfo  LPARAM
}

type NMOBJECTNOTIFY struct {
	Hdr     NMHDR
	IItem   int32
	Piid    *syscall.GUID
	PObject unsafe.Pointer
	HResult HRESULT
	DwFlags uint32
}

type NMKEY struct {
	Hdr    NMHDR
	NVKey  uint32
	UFlags uint32
}

type NMCHAR struct {
	Hdr        NMHDR
	Ch         uint32
	DwItemPrev uint32
	DwItemNext uint32
}

type NMCUSTOMTEXT struct {
	Hdr      NMHDR
	HDC      HDC
	LpString PWSTR
	NCount   int32
	LpRect   *RECT
	UFormat  uint32
	FLink    BOOL
}

type NMCUSTOMDRAW struct {
	Hdr         NMHDR
	DwDrawStage NMCUSTOMDRAW_DRAW_STAGE
	Hdc         HDC
	Rc          RECT
	DwItemSpec  uintptr
	UItemState  uint32
	LItemlParam LPARAM
}

type NMTTCUSTOMDRAW struct {
	Nmcd       NMCUSTOMDRAW
	UDrawFlags uint32
}

type NMCUSTOMSPLITRECTINFO struct {
	Hdr      NMHDR
	RcClient RECT
	RcButton RECT
	RcSplit  RECT
}

type IMAGELISTDRAWPARAMS struct {
	CbSize   uint32
	Himl     HIMAGELIST
	I        int32
	HdcDst   HDC
	X        int32
	Y        int32
	Cx       int32
	Cy       int32
	XBitmap  int32
	YBitmap  int32
	RgbBk    uint32
	RgbFg    uint32
	FStyle   uint32
	DwRop    uint32
	FState   uint32
	Frame    uint32
	CrEffect uint32
}

type IMAGEINFO struct {
	HbmImage HBITMAP
	HbmMask  HBITMAP
	Unused1  int32
	Unused2  int32
	RcImage  RECT
}

type HD_TEXTFILTERA struct {
	PszText    PSTR
	CchTextMax int32
}

type HD_TEXTFILTER = HD_TEXTFILTERW
type HD_TEXTFILTERW struct {
	PszText    PWSTR
	CchTextMax int32
}

type HDITEMA struct {
	Mask       HDI_MASK
	Cxy        int32
	PszText    PSTR
	Hbm        HBITMAP
	CchTextMax int32
	Fmt        int32
	LParam     LPARAM
	IImage     int32
	IOrder     int32
	Type_      uint32
	PvFilter   unsafe.Pointer
	State      uint32
}

type HDITEM = HDITEMW
type HDITEMW struct {
	Mask       HDI_MASK
	Cxy        int32
	PszText    PWSTR
	Hbm        HBITMAP
	CchTextMax int32
	Fmt        int32
	LParam     LPARAM
	IImage     int32
	IOrder     int32
	Type_      uint32
	PvFilter   unsafe.Pointer
	State      uint32
}

type HDLAYOUT struct {
	Prc   *RECT
	Pwpos *WINDOWPOS
}

type HDHITTESTINFO struct {
	Pt    POINT
	Flags uint32
	IItem int32
}

type NMHEADERA struct {
	Hdr     NMHDR
	IItem   int32
	IButton HEADER_CONTROL_NOTIFICATION_BUTTON
	Pitem   *HDITEMA
}

type NMHEADER = NMHEADERW
type NMHEADERW struct {
	Hdr     NMHDR
	IItem   int32
	IButton HEADER_CONTROL_NOTIFICATION_BUTTON
	Pitem   *HDITEMW
}

type NMHDDISPINFO = NMHDDISPINFOW
type NMHDDISPINFOW struct {
	Hdr        NMHDR
	IItem      int32
	Mask       HDI_MASK
	PszText    PWSTR
	CchTextMax int32
	IImage     int32
	LParam     LPARAM
}

type NMHDDISPINFOA struct {
	Hdr        NMHDR
	IItem      int32
	Mask       HDI_MASK
	PszText    PSTR
	CchTextMax int32
	IImage     int32
	LParam     LPARAM
}

type NMHDFILTERBTNCLICK struct {
	Hdr   NMHDR
	IItem int32
	Rc    RECT
}

type TBBUTTON struct {
	IBitmap   int32
	IdCommand int32
	FsState   byte
	FsStyle   byte
	BReserved [6]byte
	DwData    uintptr
	IString   uintptr
}

type COLORMAP struct {
	From uint32
	To   uint32
}

type NMTBCUSTOMDRAW struct {
	Nmcd                 NMCUSTOMDRAW
	HbrMonoDither        HBRUSH
	HbrLines             HBRUSH
	HpenLines            HPEN
	ClrText              uint32
	ClrMark              uint32
	ClrTextHighlight     uint32
	ClrBtnFace           uint32
	ClrBtnHighlight      uint32
	ClrHighlightHotTrack uint32
	RcText               RECT
	NStringBkMode        int32
	NHLStringBkMode      int32
	IListGap             int32
}

type TBADDBITMAP struct {
	HInst HINSTANCE
	NID   uintptr
}

type TBSAVEPARAMSA struct {
	Hkr          HKEY
	PszSubKey    PSTR
	PszValueName PSTR
}

type TBSAVEPARAMS = TBSAVEPARAMSW
type TBSAVEPARAMSW struct {
	Hkr          HKEY
	PszSubKey    PWSTR
	PszValueName PWSTR
}

type TBINSERTMARK struct {
	IButton int32
	DwFlags TBINSERTMARK_FLAGS
}

type TBREPLACEBITMAP struct {
	HInstOld HINSTANCE
	NIDOld   uintptr
	HInstNew HINSTANCE
	NIDNew   uintptr
	NButtons int32
}

type TBBUTTONINFOA struct {
	CbSize    uint32
	DwMask    TBBUTTONINFOW_MASK
	IdCommand int32
	IImage    int32
	FsState   byte
	FsStyle   byte
	Cx        uint16
	LParam    uintptr
	PszText   PSTR
	CchText   int32
}

type TBBUTTONINFO = TBBUTTONINFOW
type TBBUTTONINFOW struct {
	CbSize    uint32
	DwMask    TBBUTTONINFOW_MASK
	IdCommand int32
	IImage    int32
	FsState   byte
	FsStyle   byte
	Cx        uint16
	LParam    uintptr
	PszText   PWSTR
	CchText   int32
}

type TBMETRICS struct {
	CbSize          uint32
	DwMask          uint32
	CxPad           int32
	CyPad           int32
	CxBarPad        int32
	CyBarPad        int32
	CxButtonSpacing int32
	CyButtonSpacing int32
}

type NMTBHOTITEM struct {
	Hdr     NMHDR
	IdOld   int32
	IdNew   int32
	DwFlags NMTBHOTITEM_FLAGS
}

type NMTBSAVE struct {
	Hdr      NMHDR
	PData    *uint32
	PCurrent *uint32
	CbData   uint32
	IItem    int32
	CButtons int32
	TbButton TBBUTTON
}

type NMTBRESTORE struct {
	Hdr              NMHDR
	PData            *uint32
	PCurrent         *uint32
	CbData           uint32
	IItem            int32
	CButtons         int32
	CbBytesPerRecord int32
	TbButton         TBBUTTON
}

type NMTBGETINFOTIPA struct {
	Hdr        NMHDR
	PszText    PSTR
	CchTextMax int32
	IItem      int32
	LParam     LPARAM
}

type NMTBGETINFOTIP = NMTBGETINFOTIPW
type NMTBGETINFOTIPW struct {
	Hdr        NMHDR
	PszText    PWSTR
	CchTextMax int32
	IItem      int32
	LParam     LPARAM
}

type NMTBDISPINFOA struct {
	Hdr       NMHDR
	DwMask    NMTBDISPINFOW_MASK
	IdCommand int32
	LParam    uintptr
	IImage    int32
	PszText   PSTR
	CchText   int32
}

type NMTBDISPINFO = NMTBDISPINFOW
type NMTBDISPINFOW struct {
	Hdr       NMHDR
	DwMask    NMTBDISPINFOW_MASK
	IdCommand int32
	LParam    uintptr
	IImage    int32
	PszText   PWSTR
	CchText   int32
}

type NMTOOLBARA struct {
	Hdr      NMHDR
	IItem    int32
	TbButton TBBUTTON
	CchText  int32
	PszText  PSTR
	RcButton RECT
}

type NMTOOLBAR = NMTOOLBARW
type NMTOOLBARW struct {
	Hdr      NMHDR
	IItem    int32
	TbButton TBBUTTON
	CchText  int32
	PszText  PWSTR
	RcButton RECT
}

type REBARINFO struct {
	CbSize uint32
	FMask  uint32
	Himl   HIMAGELIST
}

type REBARBANDINFOA struct {
	CbSize            uint32
	FMask             uint32
	FStyle            uint32
	ClrFore           uint32
	ClrBack           uint32
	LpText            PSTR
	Cch               uint32
	IImage            int32
	HwndChild         HWND
	CxMinChild        uint32
	CyMinChild        uint32
	Cx                uint32
	HbmBack           HBITMAP
	WID               uint32
	CyChild           uint32
	CyMaxChild        uint32
	CyIntegral        uint32
	CxIdeal           uint32
	LParam            LPARAM
	CxHeader          uint32
	RcChevronLocation RECT
	UChevronState     uint32
}

type REBARBANDINFO = REBARBANDINFOW
type REBARBANDINFOW struct {
	CbSize            uint32
	FMask             uint32
	FStyle            uint32
	ClrFore           uint32
	ClrBack           uint32
	LpText            PWSTR
	Cch               uint32
	IImage            int32
	HwndChild         HWND
	CxMinChild        uint32
	CyMinChild        uint32
	Cx                uint32
	HbmBack           HBITMAP
	WID               uint32
	CyChild           uint32
	CyMaxChild        uint32
	CyIntegral        uint32
	CxIdeal           uint32
	LParam            LPARAM
	CxHeader          uint32
	RcChevronLocation RECT
	UChevronState     uint32
}

type NMREBARCHILDSIZE struct {
	Hdr     NMHDR
	UBand   uint32
	WID     uint32
	RcChild RECT
	RcBand  RECT
}

type NMREBAR struct {
	Hdr    NMHDR
	DwMask NMREBAR_MASK_FLAGS
	UBand  uint32
	FStyle uint32
	WID    uint32
	LParam LPARAM
}

type NMRBAUTOSIZE struct {
	Hdr      NMHDR
	FChanged BOOL
	RcTarget RECT
	RcActual RECT
}

type NMREBARCHEVRON struct {
	Hdr      NMHDR
	UBand    uint32
	WID      uint32
	LParam   LPARAM
	Rc       RECT
	LParamNM LPARAM
}

type NMREBARSPLITTER struct {
	Hdr      NMHDR
	RcSizing RECT
}

type NMREBARAUTOBREAK struct {
	Hdr           NMHDR
	UBand         uint32
	WID           uint32
	LParam        LPARAM
	UMsg          uint32
	FStyleCurrent uint32
	FAutoBreak    BOOL
}

type RBHITTESTINFO struct {
	Pt    POINT
	Flags uint32
	IBand int32
}

type TTTOOLINFOA struct {
	CbSize     uint32
	UFlags     TTTOOLINFO_FLAGS
	Hwnd       HWND
	UId        uintptr
	Rect       RECT
	Hinst      HINSTANCE
	LpszText   PSTR
	LParam     LPARAM
	LpReserved unsafe.Pointer
}

type TTTOOLINFO = TTTOOLINFOW
type TTTOOLINFOW struct {
	CbSize     uint32
	UFlags     TTTOOLINFO_FLAGS
	Hwnd       HWND
	UId        uintptr
	Rect       RECT
	Hinst      HINSTANCE
	LpszText   PWSTR
	LParam     LPARAM
	LpReserved unsafe.Pointer
}

type TTGETTITLE struct {
	DwSize       uint32
	UTitleBitmap uint32
	Cch          uint32
	PszTitle     PWSTR
}

type TTHITTESTINFOA struct {
	Hwnd HWND
	Pt   POINT
	Ti   TTTOOLINFOA
}

type TTHITTESTINFO = TTHITTESTINFOW
type TTHITTESTINFOW struct {
	Hwnd HWND
	Pt   POINT
	Ti   TTTOOLINFOW
}

type NMTTDISPINFOA struct {
	Hdr      NMHDR
	LpszText PSTR
	SzText   [80]CHAR
	Hinst    HINSTANCE
	UFlags   uint32
	LParam   LPARAM
}

type NMTTDISPINFO = NMTTDISPINFOW
type NMTTDISPINFOW struct {
	Hdr      NMHDR
	LpszText PWSTR
	SzText   [80]uint16
	Hinst    HINSTANCE
	UFlags   uint32
	LParam   LPARAM
}

type NMTRBTHUMBPOSCHANGING struct {
	Hdr     NMHDR
	DwPos   uint32
	NReason int32
}

type DRAGLISTINFO struct {
	UNotification DRAGLISTINFO_NOTIFICATION_FLAGS
	HWnd          HWND
	PtCursor      POINT
}

type UDACCEL struct {
	NSec uint32
	NInc uint32
}

type NMUPDOWN struct {
	Hdr    NMHDR
	IPos   int32
	IDelta int32
}

type PBRANGE struct {
	ILow  int32
	IHigh int32
}

type LITEM struct {
	Mask      uint32
	ILink     int32
	State     uint32
	StateMask uint32
	SzID      [48]uint16
	SzUrl     [2084]uint16
}

type LHITTESTINFO struct {
	Pt   POINT
	Item LITEM
}

type NMLINK struct {
	Hdr  NMHDR
	Item LITEM
}

type LVITEMA struct {
	Mask       uint32
	IItem      int32
	ISubItem   int32
	State      uint32
	StateMask  uint32
	PszText    PSTR
	CchTextMax int32
	IImage     int32
	LParam     LPARAM
	IIndent    int32
	IGroupId   LVITEMA_GROUP_ID
	CColumns   uint32
	PuColumns  *uint32
	PiColFmt   *int32
	IGroup     int32
}

type LVITEM = LVITEMW
type LVITEMW struct {
	Mask       uint32
	IItem      int32
	ISubItem   int32
	State      uint32
	StateMask  uint32
	PszText    PWSTR
	CchTextMax int32
	IImage     int32
	LParam     LPARAM
	IIndent    int32
	IGroupId   LVITEMA_GROUP_ID
	CColumns   uint32
	PuColumns  *uint32
	PiColFmt   *int32
	IGroup     int32
}

type LVFINDINFOA struct {
	Flags       LVFINDINFOW_FLAGS
	Psz         PSTR
	LParam      LPARAM
	Pt          POINT
	VkDirection uint32
}

type LVFINDINFO = LVFINDINFOW
type LVFINDINFOW struct {
	Flags       LVFINDINFOW_FLAGS
	Psz         PWSTR
	LParam      LPARAM
	Pt          POINT
	VkDirection uint32
}

type LVHITTESTINFO struct {
	Pt       POINT
	Flags    LVHITTESTINFO_FLAGS
	IItem    int32
	ISubItem int32
	IGroup   int32
}

type LVCOLUMNA struct {
	Mask       LVCOLUMNW_MASK
	Fmt        LVCOLUMNW_FORMAT
	Cx         int32
	PszText    PSTR
	CchTextMax int32
	ISubItem   int32
	IImage     int32
	IOrder     int32
	CxMin      int32
	CxDefault  int32
	CxIdeal    int32
}

type LVCOLUMN = LVCOLUMNW
type LVCOLUMNW struct {
	Mask       LVCOLUMNW_MASK
	Fmt        LVCOLUMNW_FORMAT
	Cx         int32
	PszText    PWSTR
	CchTextMax int32
	ISubItem   int32
	IImage     int32
	IOrder     int32
	CxMin      int32
	CxDefault  int32
	CxIdeal    int32
}

type LVBKIMAGEA struct {
	UlFlags        uint32
	Hbm            HBITMAP
	PszImage       PSTR
	CchImageMax    uint32
	XOffsetPercent int32
	YOffsetPercent int32
}

type LVBKIMAGE = LVBKIMAGEW
type LVBKIMAGEW struct {
	UlFlags        uint32
	Hbm            HBITMAP
	PszImage       PWSTR
	CchImageMax    uint32
	XOffsetPercent int32
	YOffsetPercent int32
}

type LVGROUP struct {
	CbSize               uint32
	Mask                 LVGROUP_MASK
	PszHeader            PWSTR
	CchHeader            int32
	PszFooter            PWSTR
	CchFooter            int32
	IGroupId             int32
	StateMask            uint32
	State                uint32
	UAlign               uint32
	PszSubtitle          PWSTR
	CchSubtitle          uint32
	PszTask              PWSTR
	CchTask              uint32
	PszDescriptionTop    PWSTR
	CchDescriptionTop    uint32
	PszDescriptionBottom PWSTR
	CchDescriptionBottom uint32
	ITitleImage          int32
	IExtendedImage       int32
	IFirstItem           int32
	CItems               uint32
	PszSubsetTitle       PWSTR
	CchSubsetTitle       uint32
}

type LVGROUPMETRICS struct {
	CbSize   uint32
	Mask     uint32
	Left     uint32
	Top      uint32
	Right    uint32
	Bottom   uint32
	CrLeft   uint32
	CrTop    uint32
	CrRight  uint32
	CrBottom uint32
	CrHeader uint32
	CrFooter uint32
}

type LVINSERTGROUPSORTED struct {
	PfnGroupCompare PFNLVGROUPCOMPARE
	PvData          unsafe.Pointer
	LvGroup         LVGROUP
}

type LVTILEVIEWINFO struct {
	CbSize        uint32
	DwMask        uint32
	DwFlags       LVTILEVIEWINFO_FLAGS
	SizeTile      SIZE
	CLines        int32
	RcLabelMargin RECT
}

type LVTILEINFO struct {
	CbSize    uint32
	IItem     int32
	CColumns  uint32
	PuColumns *uint32
	PiColFmt  *int32
}

type LVINSERTMARK struct {
	CbSize     uint32
	DwFlags    uint32
	IItem      int32
	DwReserved uint32
}

type LVSETINFOTIP struct {
	CbSize   uint32
	DwFlags  uint32
	PszText  PWSTR
	IItem    int32
	ISubItem int32
}

type LVFOOTERINFO struct {
	Mask       uint32
	PszText    PWSTR
	CchTextMax int32
	CItems     uint32
}

type LVFOOTERITEM struct {
	Mask       LVFOOTERITEM_MASK
	IItem      int32
	PszText    PWSTR
	CchTextMax int32
	State      uint32
	StateMask  uint32
}

type LVITEMINDEX struct {
	IItem  int32
	IGroup int32
}

type NMLISTVIEW struct {
	Hdr       NMHDR
	IItem     int32
	ISubItem  int32
	UNewState uint32
	UOldState uint32
	UChanged  uint32
	PtAction  POINT
	LParam    LPARAM
}

type NMITEMACTIVATE struct {
	Hdr       NMHDR
	IItem     int32
	ISubItem  int32
	UNewState uint32
	UOldState uint32
	UChanged  uint32
	PtAction  POINT
	LParam    LPARAM
	UKeyFlags uint32
}

type NMLVCUSTOMDRAW struct {
	Nmcd        NMCUSTOMDRAW
	ClrText     uint32
	ClrTextBk   uint32
	ISubItem    int32
	DwItemType  NMLVCUSTOMDRAW_ITEM_TYPE
	ClrFace     uint32
	IIconEffect int32
	IIconPhase  int32
	IPartId     int32
	IStateId    int32
	RcText      RECT
	UAlign      NMLVCUSTOMDRAW_ALIGN
}

type NMLVCACHEHINT struct {
	Hdr   NMHDR
	IFrom int32
	ITo   int32
}

type NMLVFINDITEMA struct {
	Hdr    NMHDR
	IStart int32
	Lvfi   LVFINDINFOA
}

type NMLVFINDITEM = NMLVFINDITEMW
type NMLVFINDITEMW struct {
	Hdr    NMHDR
	IStart int32
	Lvfi   LVFINDINFOW
}

type NMLVODSTATECHANGE struct {
	Hdr       NMHDR
	IFrom     int32
	ITo       int32
	UNewState uint32
	UOldState uint32
}

type NMLVDISPINFOA struct {
	Hdr  NMHDR
	Item LVITEMA
}

type NMLVDISPINFO = NMLVDISPINFOW
type NMLVDISPINFOW struct {
	Hdr  NMHDR
	Item LVITEMW
}

type NMLVKEYDOWN struct {
	Hdr   NMHDR
	WVKey uint16
	Flags uint32
}

type NMLVLINK struct {
	Hdr      NMHDR
	Link     LITEM
	IItem    int32
	ISubItem int32
}

type NMLVGETINFOTIPA struct {
	Hdr        NMHDR
	DwFlags    uint32
	PszText    PSTR
	CchTextMax int32
	IItem      int32
	ISubItem   int32
	LParam     LPARAM
}

type NMLVGETINFOTIP = NMLVGETINFOTIPW
type NMLVGETINFOTIPW struct {
	Hdr        NMHDR
	DwFlags    uint32
	PszText    PWSTR
	CchTextMax int32
	IItem      int32
	ISubItem   int32
	LParam     LPARAM
}

type NMLVSCROLL struct {
	Hdr NMHDR
	Dx  int32
	Dy  int32
}

type NMLVEMPTYMARKUP struct {
	Hdr      NMHDR
	DwFlags  NMLVEMPTYMARKUP_FLAGS
	SzMarkup [2084]uint16
}

type NMTVSTATEIMAGECHANGING struct {
	Hdr                 NMHDR
	Hti                 HTREEITEM
	IOldStateImageIndex int32
	INewStateImageIndex int32
}

type TVITEMA struct {
	Mask           TVITEM_MASK
	HItem          HTREEITEM
	State          uint32
	StateMask      uint32
	PszText        PSTR
	CchTextMax     int32
	IImage         int32
	ISelectedImage int32
	CChildren      TVITEMEXW_CHILDREN
	LParam         LPARAM
}

type TVITEM = TVITEMW
type TVITEMW struct {
	Mask           TVITEM_MASK
	HItem          HTREEITEM
	State          uint32
	StateMask      uint32
	PszText        PWSTR
	CchTextMax     int32
	IImage         int32
	ISelectedImage int32
	CChildren      TVITEMEXW_CHILDREN
	LParam         LPARAM
}

type TVITEMEXA struct {
	Mask           TVITEM_MASK
	HItem          HTREEITEM
	State          uint32
	StateMask      uint32
	PszText        PSTR
	CchTextMax     int32
	IImage         int32
	ISelectedImage int32
	CChildren      TVITEMEXW_CHILDREN
	LParam         LPARAM
	IIntegral      int32
	UStateEx       uint32
	Hwnd           HWND
	IExpandedImage int32
	IReserved      int32
}

type TVITEMEX = TVITEMEXW
type TVITEMEXW struct {
	Mask           TVITEM_MASK
	HItem          HTREEITEM
	State          uint32
	StateMask      uint32
	PszText        PWSTR
	CchTextMax     int32
	IImage         int32
	ISelectedImage int32
	CChildren      TVITEMEXW_CHILDREN
	LParam         LPARAM
	IIntegral      int32
	UStateEx       uint32
	Hwnd           HWND
	IExpandedImage int32
	IReserved      int32
}

type TVINSERTSTRUCTA_Anonymous struct {
	Data [10]uint64
}

func (this *TVINSERTSTRUCTA_Anonymous) Itemex() *TVITEMEXA {
	return (*TVITEMEXA)(unsafe.Pointer(this))
}

func (this *TVINSERTSTRUCTA_Anonymous) ItemexVal() TVITEMEXA {
	return *(*TVITEMEXA)(unsafe.Pointer(this))
}

func (this *TVINSERTSTRUCTA_Anonymous) Item() *TVITEMA {
	return (*TVITEMA)(unsafe.Pointer(this))
}

func (this *TVINSERTSTRUCTA_Anonymous) ItemVal() TVITEMA {
	return *(*TVITEMA)(unsafe.Pointer(this))
}

type TVINSERTSTRUCTA struct {
	HParent      HTREEITEM
	HInsertAfter HTREEITEM
	TVINSERTSTRUCTA_Anonymous
}

type TVINSERTSTRUCTW_Anonymous struct {
	Data [10]uint64
}

func (this *TVINSERTSTRUCTW_Anonymous) Itemex() *TVITEMEXW {
	return (*TVITEMEXW)(unsafe.Pointer(this))
}

func (this *TVINSERTSTRUCTW_Anonymous) ItemexVal() TVITEMEXW {
	return *(*TVITEMEXW)(unsafe.Pointer(this))
}

func (this *TVINSERTSTRUCTW_Anonymous) Item() *TVITEMW {
	return (*TVITEMW)(unsafe.Pointer(this))
}

func (this *TVINSERTSTRUCTW_Anonymous) ItemVal() TVITEMW {
	return *(*TVITEMW)(unsafe.Pointer(this))
}

type TVINSERTSTRUCT = TVINSERTSTRUCTW
type TVINSERTSTRUCTW struct {
	HParent      HTREEITEM
	HInsertAfter HTREEITEM
	TVINSERTSTRUCTW_Anonymous
}

type TVHITTESTINFO struct {
	Pt    POINT
	Flags TVHITTESTINFO_FLAGS
	HItem HTREEITEM
}

type TVGETITEMPARTRECTINFO struct {
	Hti    HTREEITEM
	Prc    *RECT
	PartID TVITEMPART
}

type TVSORTCB struct {
	HParent     HTREEITEM
	LpfnCompare PFNTVCOMPARE
	LParam      LPARAM
}

type NMTREEVIEWA struct {
	Hdr     NMHDR
	Action  uint32
	ItemOld TVITEMA
	ItemNew TVITEMA
	PtDrag  POINT
}

type NMTREEVIEW = NMTREEVIEWW
type NMTREEVIEWW struct {
	Hdr     NMHDR
	Action  uint32
	ItemOld TVITEMW
	ItemNew TVITEMW
	PtDrag  POINT
}

type NMTVDISPINFOA struct {
	Hdr  NMHDR
	Item TVITEMA
}

type NMTVDISPINFO = NMTVDISPINFOW
type NMTVDISPINFOW struct {
	Hdr  NMHDR
	Item TVITEMW
}

type NMTVDISPINFOEXA struct {
	Hdr  NMHDR
	Item TVITEMEXA
}

type NMTVDISPINFOEX = NMTVDISPINFOEXW
type NMTVDISPINFOEXW struct {
	Hdr  NMHDR
	Item TVITEMEXW
}

type NMTVKEYDOWN struct {
	Hdr   NMHDR
	WVKey uint16
	Flags uint32
}

type NMTVCUSTOMDRAW struct {
	Nmcd      NMCUSTOMDRAW
	ClrText   uint32
	ClrTextBk uint32
	ILevel    int32
}

type NMTVGETINFOTIPA struct {
	Hdr        NMHDR
	PszText    PSTR
	CchTextMax int32
	HItem      HTREEITEM
	LParam     LPARAM
}

type NMTVGETINFOTIP = NMTVGETINFOTIPW
type NMTVGETINFOTIPW struct {
	Hdr        NMHDR
	PszText    PWSTR
	CchTextMax int32
	HItem      HTREEITEM
	LParam     LPARAM
}

type NMTVITEMCHANGE struct {
	Hdr       NMHDR
	UChanged  uint32
	HItem     HTREEITEM
	UStateNew uint32
	UStateOld uint32
	LParam    LPARAM
}

type NMTVASYNCDRAW struct {
	Hdr            NMHDR
	Pimldp         *IMAGELISTDRAWPARAMS
	Hr             HRESULT
	HItem          HTREEITEM
	LParam         LPARAM
	DwRetFlags     uint32
	IRetImageIndex int32
}

type COMBOBOXEXITEMA struct {
	Mask           COMBOBOX_EX_ITEM_FLAGS
	IItem          uintptr
	PszText        PSTR
	CchTextMax     int32
	IImage         int32
	ISelectedImage int32
	IOverlay       int32
	IIndent        int32
	LParam         LPARAM
}

type COMBOBOXEXITEM = COMBOBOXEXITEMW
type COMBOBOXEXITEMW struct {
	Mask           COMBOBOX_EX_ITEM_FLAGS
	IItem          uintptr
	PszText        PWSTR
	CchTextMax     int32
	IImage         int32
	ISelectedImage int32
	IOverlay       int32
	IIndent        int32
	LParam         LPARAM
}

type NMCOMBOBOXEXA struct {
	Hdr    NMHDR
	CeItem COMBOBOXEXITEMA
}

type NMCOMBOBOXEX = NMCOMBOBOXEXW
type NMCOMBOBOXEXW struct {
	Hdr    NMHDR
	CeItem COMBOBOXEXITEMW
}

type NMCBEDRAGBEGIN = NMCBEDRAGBEGINW
type NMCBEDRAGBEGINW struct {
	Hdr     NMHDR
	IItemid int32
	SzText  [260]uint16
}

type NMCBEDRAGBEGINA struct {
	Hdr     NMHDR
	IItemid int32
	SzText  [260]CHAR
}

type NMCBEENDEDIT = NMCBEENDEDITW
type NMCBEENDEDITW struct {
	Hdr           NMHDR
	FChanged      BOOL
	INewSelection int32
	SzText        [260]uint16
	IWhy          int32
}

type NMCBEENDEDITA struct {
	Hdr           NMHDR
	FChanged      BOOL
	INewSelection int32
	SzText        [260]CHAR
	IWhy          int32
}

type TCITEMHEADERA struct {
	Mask        TCITEMHEADERA_MASK
	LpReserved1 uint32
	LpReserved2 uint32
	PszText     PSTR
	CchTextMax  int32
	IImage      int32
}

type TCITEMHEADER = TCITEMHEADERW
type TCITEMHEADERW struct {
	Mask        TCITEMHEADERA_MASK
	LpReserved1 uint32
	LpReserved2 uint32
	PszText     PWSTR
	CchTextMax  int32
	IImage      int32
}

type TCITEMA struct {
	Mask        TCITEMHEADERA_MASK
	DwState     uint32
	DwStateMask uint32
	PszText     PSTR
	CchTextMax  int32
	IImage      int32
	LParam      LPARAM
}

type TCITEM = TCITEMW
type TCITEMW struct {
	Mask        TCITEMHEADERA_MASK
	DwState     uint32
	DwStateMask uint32
	PszText     PWSTR
	CchTextMax  int32
	IImage      int32
	LParam      LPARAM
}

type TCHITTESTINFO struct {
	Pt    POINT
	Flags TCHITTESTINFO_FLAGS
}

type NMTCKEYDOWN struct {
	Hdr   NMHDR
	WVKey uint16
	Flags uint32
}

type MCHITTESTINFO struct {
	CbSize  uint32
	Pt      POINT
	UHit    uint32
	St      SYSTEMTIME
	Rc      RECT
	IOffset int32
	IRow    int32
	ICol    int32
}

type MCGRIDINFO struct {
	CbSize    uint32
	DwPart    MCGRIDINFO_PART
	DwFlags   MCGRIDINFO_FLAGS
	ICalendar int32
	IRow      int32
	ICol      int32
	BSelected BOOL
	StStart   SYSTEMTIME
	StEnd     SYSTEMTIME
	Rc        RECT
	PszName   PWSTR
	CchName   uintptr
}

type NMSELCHANGE struct {
	Nmhdr      NMHDR
	StSelStart SYSTEMTIME
	StSelEnd   SYSTEMTIME
}

type NMDAYSTATE struct {
	Nmhdr       NMHDR
	StStart     SYSTEMTIME
	CDayState   int32
	PrgDayState *uint32
}

type NMVIEWCHANGE struct {
	Nmhdr     NMHDR
	DwOldView uint32
	DwNewView uint32
}

type DATETIMEPICKERINFO struct {
	CbSize       uint32
	RcCheck      RECT
	StateCheck   uint32
	RcButton     RECT
	StateButton  uint32
	HwndEdit     HWND
	HwndUD       HWND
	HwndDropDown HWND
}

type NMDATETIMECHANGE struct {
	Nmhdr   NMHDR
	DwFlags uint32
	St      SYSTEMTIME
}

type NMDATETIMESTRINGA struct {
	Nmhdr         NMHDR
	PszUserString PSTR
	St            SYSTEMTIME
	DwFlags       uint32
}

type NMDATETIMESTRING = NMDATETIMESTRINGW
type NMDATETIMESTRINGW struct {
	Nmhdr         NMHDR
	PszUserString PWSTR
	St            SYSTEMTIME
	DwFlags       uint32
}

type NMDATETIMEWMKEYDOWNA struct {
	Nmhdr     NMHDR
	NVirtKey  int32
	PszFormat PSTR
	St        SYSTEMTIME
}

type NMDATETIMEWMKEYDOWN = NMDATETIMEWMKEYDOWNW
type NMDATETIMEWMKEYDOWNW struct {
	Nmhdr     NMHDR
	NVirtKey  int32
	PszFormat PWSTR
	St        SYSTEMTIME
}

type NMDATETIMEFORMATA struct {
	Nmhdr      NMHDR
	PszFormat  PSTR
	St         SYSTEMTIME
	PszDisplay PSTR
	SzDisplay  [64]CHAR
}

type NMDATETIMEFORMAT = NMDATETIMEFORMATW
type NMDATETIMEFORMATW struct {
	Nmhdr      NMHDR
	PszFormat  PWSTR
	St         SYSTEMTIME
	PszDisplay PWSTR
	SzDisplay  [64]uint16
}

type NMDATETIMEFORMATQUERYA struct {
	Nmhdr     NMHDR
	PszFormat PSTR
	SzMax     SIZE
}

type NMDATETIMEFORMATQUERY = NMDATETIMEFORMATQUERYW
type NMDATETIMEFORMATQUERYW struct {
	Nmhdr     NMHDR
	PszFormat PWSTR
	SzMax     SIZE
}

type NMIPADDRESS struct {
	Hdr    NMHDR
	IField int32
	IValue int32
}

type NMPGSCROLL struct {
	Hdr      NMHDR
	FwKeys   NMPGSCROLL_KEYS
	RcParent RECT
	IDir     NMPGSCROLL_DIR
	IXpos    int32
	IYpos    int32
	IScroll  int32
}

type NMPGCALCSIZE struct {
	Hdr     NMHDR
	DwFlag  NMPGCALCSIZE_FLAGS
	IWidth  int32
	IHeight int32
}

type NMPGHOTITEM struct {
	Hdr     NMHDR
	IdOld   int32
	IdNew   int32
	DwFlags uint32
}

type BUTTON_IMAGELIST struct {
	Himl   HIMAGELIST
	Margin RECT
	UAlign BUTTON_IMAGELIST_ALIGN
}

type NMBCHOTITEM struct {
	Hdr     NMHDR
	DwFlags NMTBHOTITEM_FLAGS
}

type BUTTON_SPLITINFO struct {
	Mask        uint32
	HimlGlyph   HIMAGELIST
	USplitStyle uint32
	Size        SIZE
}

type NMBCDROPDOWN struct {
	Hdr      NMHDR
	RcButton RECT
}

type EDITBALLOONTIP struct {
	CbStruct uint32
	PszTitle PWSTR
	PszText  PWSTR
	TtiIcon  EDITBALLOONTIP_ICON
}

type NMSEARCHWEB struct {
	Hdr             NMHDR
	Entrypoint      EC_SEARCHWEB_ENTRYPOINT
	HasQueryText    BOOL
	InvokeSucceeded BOOL
}

type TASKDIALOG_BUTTON struct {
	NButtonID     int32
	PszButtonText PWSTR
}

type TASKDIALOGCONFIG_Anonymous1 struct {
	Data [1]uint64
}

func (this *TASKDIALOGCONFIG_Anonymous1) HMainIcon() *HICON {
	return (*HICON)(unsafe.Pointer(this))
}

func (this *TASKDIALOGCONFIG_Anonymous1) HMainIconVal() HICON {
	return *(*HICON)(unsafe.Pointer(this))
}

func (this *TASKDIALOGCONFIG_Anonymous1) PszMainIcon() *PWSTR {
	return (*PWSTR)(unsafe.Pointer(this))
}

func (this *TASKDIALOGCONFIG_Anonymous1) PszMainIconVal() PWSTR {
	return *(*PWSTR)(unsafe.Pointer(this))
}

type TASKDIALOGCONFIG_Anonymous2 struct {
	Data [1]uint64
}

func (this *TASKDIALOGCONFIG_Anonymous2) HFooterIcon() *HICON {
	return (*HICON)(unsafe.Pointer(this))
}

func (this *TASKDIALOGCONFIG_Anonymous2) HFooterIconVal() HICON {
	return *(*HICON)(unsafe.Pointer(this))
}

func (this *TASKDIALOGCONFIG_Anonymous2) PszFooterIcon() *PWSTR {
	return (*PWSTR)(unsafe.Pointer(this))
}

func (this *TASKDIALOGCONFIG_Anonymous2) PszFooterIconVal() PWSTR {
	return *(*PWSTR)(unsafe.Pointer(this))
}

type TASKDIALOGCONFIG struct {
	CbSize          uint32
	HwndParent      HWND
	HInstance       HINSTANCE
	DwFlags         TASKDIALOG_FLAGS
	DwCommonButtons TASKDIALOG_COMMON_BUTTON_FLAGS
	PszWindowTitle  PWSTR
	TASKDIALOGCONFIG_Anonymous1
	PszMainInstruction      PWSTR
	PszContent              PWSTR
	CButtons                uint32
	PButtons                *TASKDIALOG_BUTTON
	NDefaultButton          int32
	CRadioButtons           uint32
	PRadioButtons           *TASKDIALOG_BUTTON
	NDefaultRadioButton     int32
	PszVerificationText     PWSTR
	PszExpandedInformation  PWSTR
	PszExpandedControlText  PWSTR
	PszCollapsedControlText PWSTR
	TASKDIALOGCONFIG_Anonymous2
	PszFooter      PWSTR
	PfCallback     PFTASKDIALOGCALLBACK
	LpCallbackData uintptr
	CxWidth        uint32
}

type DPASTREAMINFO struct {
	IPos   int32
	PvItem unsafe.Pointer
}

type ImageList struct {
}

type IMAGELISTSTATS struct {
	CbSize   uint32
	CAlloc   int32
	CUsed    int32
	CStandby int32
}

type TA_TRANSFORM struct {
	ETransformType     TA_TRANSFORM_TYPE
	DwTimingFunctionId uint32
	DwStartTime        uint32
	DwDurationTime     uint32
	EFlags             TA_TRANSFORM_FLAG
}

type TA_TRANSFORM_2D struct {
	Header    TA_TRANSFORM
	RX        float32
	RY        float32
	RInitialX float32
	RInitialY float32
	ROriginX  float32
	ROriginY  float32
}

type TA_TRANSFORM_OPACITY struct {
	Header          TA_TRANSFORM
	ROpacity        float32
	RInitialOpacity float32
}

type TA_TRANSFORM_CLIP struct {
	Header         TA_TRANSFORM
	RLeft          float32
	RTop           float32
	RRight         float32
	RBottom        float32
	RInitialLeft   float32
	RInitialTop    float32
	RInitialRight  float32
	RInitialBottom float32
}

type TA_TIMINGFUNCTION struct {
	ETimingFunctionType TA_TIMINGFUNCTION_TYPE
}

type TA_CUBIC_BEZIER struct {
	Header TA_TIMINGFUNCTION
	RX0    float32
	RY0    float32
	RX1    float32
	RY1    float32
}

type DTBGOPTS struct {
	DwSize  uint32
	DwFlags uint32
	RcClip  RECT
}

type MARGINS struct {
	CxLeftWidth    int32
	CxRightWidth   int32
	CyTopHeight    int32
	CyBottomHeight int32
}

type INTLIST struct {
	IValueCount int32
	IValues     [402]int32
}

type WTA_OPTIONS struct {
	DwFlags uint32
	DwMask  uint32
}

type DTTOPTS struct {
	DwSize              uint32
	DwFlags             uint32
	CrText              uint32
	CrBorder            uint32
	CrShadow            uint32
	ITextShadowType     int32
	PtShadowOffset      POINT
	IBorderSize         int32
	IFontPropId         int32
	IColorPropId        int32
	IStateId            int32
	FApplyOverlay       BOOL
	IGlowSize           int32
	PfnDrawTextCallback DTT_CALLBACK_PROC
	LParam              LPARAM
}

type BP_ANIMATIONPARAMS struct {
	CbSize     uint32
	DwFlags    uint32
	Style      BP_ANIMATIONSTYLE
	DwDuration uint32
}

type BP_PAINTPARAMS struct {
	CbSize         uint32
	DwFlags        BP_PAINTPARAMS_FLAGS
	PrcExclude     *RECT
	PBlendFunction *BLENDFUNCTION
}

type CCSTYLEA struct {
	FlStyle    uint32
	FlExtStyle uint32
	SzText     [256]CHAR
	Lgid       uint16
	WReserved1 uint16
}

type CCSTYLE = CCSTYLEW
type CCSTYLEW struct {
	FlStyle    uint32
	FlExtStyle uint32
	SzText     [256]uint16
	Lgid       uint16
	WReserved1 uint16
}

type CCSTYLEFLAGA struct {
	FlStyle     uint32
	FlStyleMask uint32
	PszStyle    PSTR
}

type CCSTYLEFLAG = CCSTYLEFLAGW
type CCSTYLEFLAGW struct {
	FlStyle     uint32
	FlStyleMask uint32
	PszStyle    PWSTR
}

type CCINFOA struct {
	SzClass           [32]CHAR
	FlOptions         uint32
	SzDesc            [32]CHAR
	CxDefault         uint32
	CyDefault         uint32
	FlStyleDefault    uint32
	FlExtStyleDefault uint32
	FlCtrlTypeMask    uint32
	SzTextDefault     [256]CHAR
	CStyleFlags       int32
	AStyleFlags       *CCSTYLEFLAGA
	LpfnStyle         LPFNCCSTYLEA
	LpfnSizeToText    LPFNCCSIZETOTEXTA
	DwReserved1       uint32
	DwReserved2       uint32
}

type CCINFO = CCINFOW
type CCINFOW struct {
	SzClass           [32]uint16
	FlOptions         uint32
	SzDesc            [32]uint16
	CxDefault         uint32
	CyDefault         uint32
	FlStyleDefault    uint32
	FlExtStyleDefault uint32
	FlCtrlTypeMask    uint32
	CStyleFlags       int32
	AStyleFlags       *CCSTYLEFLAGW
	SzTextDefault     [256]uint16
	LpfnStyle         LPFNCCSTYLEW
	LpfnSizeToText    LPFNCCSIZETOTEXTW
	DwReserved1       uint32
	DwReserved2       uint32
}

type NMHDR struct {
	HwndFrom HWND
	IdFrom   uintptr
	Code     uint32
}

type MEASUREITEMSTRUCT struct {
	CtlType    uint32
	CtlID      uint32
	ItemID     uint32
	ItemWidth  uint32
	ItemHeight uint32
	ItemData   uintptr
}

type DRAWITEMSTRUCT struct {
	CtlType    DRAWITEMSTRUCT_CTL_TYPE
	CtlID      uint32
	ItemID     uint32
	ItemAction uint32
	ItemState  uint32
	HwndItem   HWND
	HDC        HDC
	RcItem     RECT
	ItemData   uintptr
}

type DELETEITEMSTRUCT struct {
	CtlType  DRAWITEMSTRUCT_CTL_TYPE
	CtlID    uint32
	ItemID   uint32
	HwndItem HWND
	ItemData uintptr
}

type COMPAREITEMSTRUCT struct {
	CtlType    uint32
	CtlID      uint32
	HwndItem   HWND
	ItemID1    uint32
	ItemData1  uintptr
	ItemID2    uint32
	ItemData2  uintptr
	DwLocaleId uint32
}

type USAGE_PROPERTIES struct {
	Level           uint16
	Page            uint16
	Usage           uint16
	LogicalMinimum  int32
	LogicalMaximum  int32
	Unit            uint16
	Exponent        uint16
	Count           byte
	PhysicalMinimum int32
	PhysicalMaximum int32
}

type POINTER_TYPE_INFO_Anonymous struct {
	Data [18]uint64
}

func (this *POINTER_TYPE_INFO_Anonymous) TouchInfo() *POINTER_TOUCH_INFO {
	return (*POINTER_TOUCH_INFO)(unsafe.Pointer(this))
}

func (this *POINTER_TYPE_INFO_Anonymous) TouchInfoVal() POINTER_TOUCH_INFO {
	return *(*POINTER_TOUCH_INFO)(unsafe.Pointer(this))
}

func (this *POINTER_TYPE_INFO_Anonymous) PenInfo() *POINTER_PEN_INFO {
	return (*POINTER_PEN_INFO)(unsafe.Pointer(this))
}

func (this *POINTER_TYPE_INFO_Anonymous) PenInfoVal() POINTER_PEN_INFO {
	return *(*POINTER_PEN_INFO)(unsafe.Pointer(this))
}

type POINTER_TYPE_INFO struct {
	Type_ POINTER_INPUT_TYPE
	POINTER_TYPE_INFO_Anonymous
}

type TOUCH_HIT_TESTING_PROXIMITY_EVALUATION struct {
	Score         uint16
	AdjustedPoint POINT
}

type TOUCH_HIT_TESTING_INPUT struct {
	PointerId              uint32
	Point                  POINT
	BoundingBox            RECT
	NonOccludedBoundingBox RECT
	Orientation            uint32
}

type COMBOBOXINFO struct {
	CbSize      uint32
	RcItem      RECT
	RcButton    RECT
	StateButton COMBOBOXINFO_BUTTON_STATE
	HwndCombo   HWND
	HwndItem    HWND
	HwndList    HWND
}

type POINTER_DEVICE_INFO struct {
	DisplayOrientation uint32
	Device             HANDLE
	PointerDeviceType  POINTER_DEVICE_TYPE
	Monitor            HMONITOR
	StartingCursorId   uint32
	MaxActiveContacts  uint16
	ProductString      [520]uint16
}

type POINTER_DEVICE_PROPERTY struct {
	LogicalMin   int32
	LogicalMax   int32
	PhysicalMin  int32
	PhysicalMax  int32
	Unit         uint32
	UnitExponent uint32
	UsagePageId  uint16
	UsageId      uint16
}

type POINTER_DEVICE_CURSOR_INFO struct {
	CursorId uint32
	Cursor   POINTER_DEVICE_CURSOR_TYPE
}

// func types

type LPFNPSPCALLBACKA = uintptr
type LPFNPSPCALLBACKA_func = func(hwnd HWND, uMsg PSPCB_MESSAGE, ppsp *PROPSHEETPAGEA) uint32

type LPFNPSPCALLBACKW = uintptr
type LPFNPSPCALLBACKW_func = func(hwnd HWND, uMsg PSPCB_MESSAGE, ppsp *PROPSHEETPAGEW) uint32

type PFNPROPSHEETCALLBACK = uintptr
type PFNPROPSHEETCALLBACK_func = func(param0 HWND, param1 uint32, param2 LPARAM) int32

type LPFNSVADDPROPSHEETPAGE = uintptr
type LPFNSVADDPROPSHEETPAGE_func = func(param0 HPROPSHEETPAGE, param1 LPARAM) BOOL

type LPFNADDPROPSHEETPAGES = uintptr
type LPFNADDPROPSHEETPAGES_func = func(param0 unsafe.Pointer, param1 LPFNSVADDPROPSHEETPAGE, param2 LPARAM) BOOL

type PFNLVCOMPARE = uintptr
type PFNLVCOMPARE_func = func(param0 LPARAM, param1 LPARAM, param2 LPARAM) int32

type PFNLVGROUPCOMPARE = uintptr
type PFNLVGROUPCOMPARE_func = func(param0 int32, param1 int32, param2 unsafe.Pointer) int32

type PFNTVCOMPARE = uintptr
type PFNTVCOMPARE_func = func(lParam1 LPARAM, lParam2 LPARAM, lParamSort LPARAM) int32

type PFTASKDIALOGCALLBACK = uintptr
type PFTASKDIALOGCALLBACK_func = func(hwnd HWND, msg uint32, wParam WPARAM, lParam LPARAM, lpRefData uintptr) HRESULT

type PFNDAENUMCALLBACK = uintptr
type PFNDAENUMCALLBACK_func = func(p unsafe.Pointer, pData unsafe.Pointer) int32

type PFNDAENUMCALLBACKCONST = uintptr
type PFNDAENUMCALLBACKCONST_func = func(p unsafe.Pointer, pData unsafe.Pointer) int32

type PFNDACOMPARE = uintptr
type PFNDACOMPARE_func = func(p1 unsafe.Pointer, p2 unsafe.Pointer, lParam LPARAM) int32

type PFNDACOMPARECONST = uintptr
type PFNDACOMPARECONST_func = func(p1 unsafe.Pointer, p2 unsafe.Pointer, lParam LPARAM) int32

type PFNDPASTREAM = uintptr
type PFNDPASTREAM_func = func(pinfo *DPASTREAMINFO, pstream *IStream, pvInstData unsafe.Pointer) HRESULT

type PFNDPAMERGE = uintptr
type PFNDPAMERGE_func = func(uMsg DPAMM_MESSAGE, pvDest unsafe.Pointer, pvSrc unsafe.Pointer, lParam LPARAM) unsafe.Pointer

type PFNDPAMERGECONST = uintptr
type PFNDPAMERGECONST_func = func(uMsg DPAMM_MESSAGE, pvDest unsafe.Pointer, pvSrc unsafe.Pointer, lParam LPARAM) unsafe.Pointer

type DTT_CALLBACK_PROC = uintptr
type DTT_CALLBACK_PROC_func = func(hdc HDC, pszText PWSTR, cchText int32, prc *RECT, dwFlags uint32, lParam LPARAM) int32

type LPFNCCSTYLEA = uintptr
type LPFNCCSTYLEA_func = func(hwndParent HWND, pccs *CCSTYLEA) BOOL

type LPFNCCSTYLEW = uintptr
type LPFNCCSTYLEW_func = func(hwndParent HWND, pccs *CCSTYLEW) BOOL

type LPFNCCSIZETOTEXTA = uintptr
type LPFNCCSIZETOTEXTA_func = func(flStyle uint32, flExtStyle uint32, hfont HFONT, pszText PSTR) int32

type LPFNCCSIZETOTEXTW = uintptr
type LPFNCCSIZETOTEXTW_func = func(flStyle uint32, flExtStyle uint32, hfont HFONT, pszText PWSTR) int32

type LPFNCCINFOA = uintptr
type LPFNCCINFOA_func = func(acci *CCINFOA) uint32

type LPFNCCINFOW = uintptr
type LPFNCCINFOW_func = func(acci *CCINFOW) uint32

type EDITWORDBREAKPROCA = uintptr
type EDITWORDBREAKPROCA_func = func(lpch PSTR, ichCurrent int32, cch int32, code WORD_BREAK_ACTION) int32

type EDITWORDBREAKPROCW = uintptr
type EDITWORDBREAKPROCW_func = func(lpch PWSTR, ichCurrent int32, cch int32, code WORD_BREAK_ACTION) int32

// interfaces

// 46EB5926-582E-4017-9FDF-E8998DAA0950
var IID_IImageList = syscall.GUID{0x46EB5926, 0x582E, 0x4017,
	[8]byte{0x9F, 0xDF, 0xE8, 0x99, 0x8D, 0xAA, 0x09, 0x50}}

type IImageListInterface interface {
	IUnknownInterface
	Add(hbmImage HBITMAP, hbmMask HBITMAP, pi *int32) HRESULT
	ReplaceIcon(i int32, hicon HICON, pi *int32) HRESULT
	SetOverlayImage(iImage int32, iOverlay int32) HRESULT
	Replace(i int32, hbmImage HBITMAP, hbmMask HBITMAP) HRESULT
	AddMasked(hbmImage HBITMAP, crMask uint32, pi *int32) HRESULT
	Draw(pimldp *IMAGELISTDRAWPARAMS) HRESULT
	Remove(i int32) HRESULT
	GetIcon(i int32, flags uint32, picon *HICON) HRESULT
	GetImageInfo(i int32, pImageInfo *IMAGEINFO) HRESULT
	Copy(iDst int32, punkSrc *IUnknown, iSrc int32, uFlags uint32) HRESULT
	Merge(i1 int32, punk2 *IUnknown, i2 int32, dx int32, dy int32, riid *syscall.GUID, ppv unsafe.Pointer) HRESULT
	Clone(riid *syscall.GUID, ppv unsafe.Pointer) HRESULT
	GetImageRect(i int32, prc *RECT) HRESULT
	GetIconSize(cx *int32, cy *int32) HRESULT
	SetIconSize(cx int32, cy int32) HRESULT
	GetImageCount(pi *int32) HRESULT
	SetImageCount(uNewCount uint32) HRESULT
	SetBkColor(clrBk uint32, pclr *uint32) HRESULT
	GetBkColor(pclr *uint32) HRESULT
	BeginDrag(iTrack int32, dxHotspot int32, dyHotspot int32) HRESULT
	EndDrag() HRESULT
	DragEnter(hwndLock HWND, x int32, y int32) HRESULT
	DragLeave(hwndLock HWND) HRESULT
	DragMove(x int32, y int32) HRESULT
	SetDragCursorImage(punk *IUnknown, iDrag int32, dxHotspot int32, dyHotspot int32) HRESULT
	DragShowNolock(fShow BOOL) HRESULT
	GetDragImage(ppt *POINT, pptHotspot *POINT, riid *syscall.GUID, ppv unsafe.Pointer) HRESULT
	GetItemFlags(i int32, dwFlags *IMAGE_LIST_ITEM_FLAGS) HRESULT
	GetOverlayImage(iOverlay int32, piIndex *int32) HRESULT
}

type IImageListVtbl struct {
	IUnknownVtbl
	Add                uintptr
	ReplaceIcon        uintptr
	SetOverlayImage    uintptr
	Replace            uintptr
	AddMasked          uintptr
	Draw               uintptr
	Remove             uintptr
	GetIcon            uintptr
	GetImageInfo       uintptr
	Copy               uintptr
	Merge              uintptr
	Clone              uintptr
	GetImageRect       uintptr
	GetIconSize        uintptr
	SetIconSize        uintptr
	GetImageCount      uintptr
	SetImageCount      uintptr
	SetBkColor         uintptr
	GetBkColor         uintptr
	BeginDrag          uintptr
	EndDrag            uintptr
	DragEnter          uintptr
	DragLeave          uintptr
	DragMove           uintptr
	SetDragCursorImage uintptr
	DragShowNolock     uintptr
	GetDragImage       uintptr
	GetItemFlags       uintptr
	GetOverlayImage    uintptr
}

type IImageList struct {
	IUnknown
}

func (this *IImageList) Vtbl() *IImageListVtbl {
	return (*IImageListVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IImageList) Add(hbmImage HBITMAP, hbmMask HBITMAP, pi *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Add, uintptr(unsafe.Pointer(this)), hbmImage, hbmMask, uintptr(unsafe.Pointer(pi)))
	return HRESULT(ret)
}

func (this *IImageList) ReplaceIcon(i int32, hicon HICON, pi *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().ReplaceIcon, uintptr(unsafe.Pointer(this)), uintptr(i), hicon, uintptr(unsafe.Pointer(pi)))
	return HRESULT(ret)
}

func (this *IImageList) SetOverlayImage(iImage int32, iOverlay int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetOverlayImage, uintptr(unsafe.Pointer(this)), uintptr(iImage), uintptr(iOverlay))
	return HRESULT(ret)
}

func (this *IImageList) Replace(i int32, hbmImage HBITMAP, hbmMask HBITMAP) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Replace, uintptr(unsafe.Pointer(this)), uintptr(i), hbmImage, hbmMask)
	return HRESULT(ret)
}

func (this *IImageList) AddMasked(hbmImage HBITMAP, crMask uint32, pi *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().AddMasked, uintptr(unsafe.Pointer(this)), hbmImage, uintptr(crMask), uintptr(unsafe.Pointer(pi)))
	return HRESULT(ret)
}

func (this *IImageList) Draw(pimldp *IMAGELISTDRAWPARAMS) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Draw, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pimldp)))
	return HRESULT(ret)
}

func (this *IImageList) Remove(i int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Remove, uintptr(unsafe.Pointer(this)), uintptr(i))
	return HRESULT(ret)
}

func (this *IImageList) GetIcon(i int32, flags uint32, picon *HICON) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetIcon, uintptr(unsafe.Pointer(this)), uintptr(i), uintptr(flags), uintptr(unsafe.Pointer(picon)))
	return HRESULT(ret)
}

func (this *IImageList) GetImageInfo(i int32, pImageInfo *IMAGEINFO) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetImageInfo, uintptr(unsafe.Pointer(this)), uintptr(i), uintptr(unsafe.Pointer(pImageInfo)))
	return HRESULT(ret)
}

func (this *IImageList) Copy(iDst int32, punkSrc *IUnknown, iSrc int32, uFlags uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Copy, uintptr(unsafe.Pointer(this)), uintptr(iDst), uintptr(unsafe.Pointer(punkSrc)), uintptr(iSrc), uintptr(uFlags))
	return HRESULT(ret)
}

func (this *IImageList) Merge(i1 int32, punk2 *IUnknown, i2 int32, dx int32, dy int32, riid *syscall.GUID, ppv unsafe.Pointer) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Merge, uintptr(unsafe.Pointer(this)), uintptr(i1), uintptr(unsafe.Pointer(punk2)), uintptr(i2), uintptr(dx), uintptr(dy), uintptr(unsafe.Pointer(riid)), uintptr(ppv))
	return HRESULT(ret)
}

func (this *IImageList) Clone(riid *syscall.GUID, ppv unsafe.Pointer) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Clone, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(riid)), uintptr(ppv))
	return HRESULT(ret)
}

func (this *IImageList) GetImageRect(i int32, prc *RECT) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetImageRect, uintptr(unsafe.Pointer(this)), uintptr(i), uintptr(unsafe.Pointer(prc)))
	return HRESULT(ret)
}

func (this *IImageList) GetIconSize(cx *int32, cy *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetIconSize, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(cx)), uintptr(unsafe.Pointer(cy)))
	return HRESULT(ret)
}

func (this *IImageList) SetIconSize(cx int32, cy int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetIconSize, uintptr(unsafe.Pointer(this)), uintptr(cx), uintptr(cy))
	return HRESULT(ret)
}

func (this *IImageList) GetImageCount(pi *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetImageCount, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pi)))
	return HRESULT(ret)
}

func (this *IImageList) SetImageCount(uNewCount uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetImageCount, uintptr(unsafe.Pointer(this)), uintptr(uNewCount))
	return HRESULT(ret)
}

func (this *IImageList) SetBkColor(clrBk uint32, pclr *uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetBkColor, uintptr(unsafe.Pointer(this)), uintptr(clrBk), uintptr(unsafe.Pointer(pclr)))
	return HRESULT(ret)
}

func (this *IImageList) GetBkColor(pclr *uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetBkColor, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pclr)))
	return HRESULT(ret)
}

func (this *IImageList) BeginDrag(iTrack int32, dxHotspot int32, dyHotspot int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().BeginDrag, uintptr(unsafe.Pointer(this)), uintptr(iTrack), uintptr(dxHotspot), uintptr(dyHotspot))
	return HRESULT(ret)
}

func (this *IImageList) EndDrag() HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().EndDrag, uintptr(unsafe.Pointer(this)))
	return HRESULT(ret)
}

func (this *IImageList) DragEnter(hwndLock HWND, x int32, y int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().DragEnter, uintptr(unsafe.Pointer(this)), hwndLock, uintptr(x), uintptr(y))
	return HRESULT(ret)
}

func (this *IImageList) DragLeave(hwndLock HWND) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().DragLeave, uintptr(unsafe.Pointer(this)), hwndLock)
	return HRESULT(ret)
}

func (this *IImageList) DragMove(x int32, y int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().DragMove, uintptr(unsafe.Pointer(this)), uintptr(x), uintptr(y))
	return HRESULT(ret)
}

func (this *IImageList) SetDragCursorImage(punk *IUnknown, iDrag int32, dxHotspot int32, dyHotspot int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetDragCursorImage, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(punk)), uintptr(iDrag), uintptr(dxHotspot), uintptr(dyHotspot))
	return HRESULT(ret)
}

func (this *IImageList) DragShowNolock(fShow BOOL) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().DragShowNolock, uintptr(unsafe.Pointer(this)), uintptr(fShow))
	return HRESULT(ret)
}

func (this *IImageList) GetDragImage(ppt *POINT, pptHotspot *POINT, riid *syscall.GUID, ppv unsafe.Pointer) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetDragImage, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(ppt)), uintptr(unsafe.Pointer(pptHotspot)), uintptr(unsafe.Pointer(riid)), uintptr(ppv))
	return HRESULT(ret)
}

func (this *IImageList) GetItemFlags(i int32, dwFlags *IMAGE_LIST_ITEM_FLAGS) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetItemFlags, uintptr(unsafe.Pointer(this)), uintptr(i), uintptr(unsafe.Pointer(dwFlags)))
	return HRESULT(ret)
}

func (this *IImageList) GetOverlayImage(iOverlay int32, piIndex *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetOverlayImage, uintptr(unsafe.Pointer(this)), uintptr(iOverlay), uintptr(unsafe.Pointer(piIndex)))
	return HRESULT(ret)
}

// 192B9D83-50FC-457B-90A0-2B82A8B5DAE1
var IID_IImageList2 = syscall.GUID{0x192B9D83, 0x50FC, 0x457B,
	[8]byte{0x90, 0xA0, 0x2B, 0x82, 0xA8, 0xB5, 0xDA, 0xE1}}

type IImageList2Interface interface {
	IImageListInterface
	Resize(cxNewIconSize int32, cyNewIconSize int32) HRESULT
	GetOriginalSize(iImage int32, dwFlags uint32, pcx *int32, pcy *int32) HRESULT
	SetOriginalSize(iImage int32, cx int32, cy int32) HRESULT
	SetCallback(punk *IUnknown) HRESULT
	GetCallback(riid *syscall.GUID, ppv unsafe.Pointer) HRESULT
	ForceImagePresent(iImage int32, dwFlags uint32) HRESULT
	DiscardImages(iFirstImage int32, iLastImage int32, dwFlags uint32) HRESULT
	PreloadImages(pimldp *IMAGELISTDRAWPARAMS) HRESULT
	GetStatistics(pils *IMAGELISTSTATS) HRESULT
	Initialize(cx int32, cy int32, flags IMAGELIST_CREATION_FLAGS, cInitial int32, cGrow int32) HRESULT
	Replace2(i int32, hbmImage HBITMAP, hbmMask HBITMAP, punk *IUnknown, dwFlags uint32) HRESULT
	ReplaceFromImageList(i int32, pil *IImageList, iSrc int32, punk *IUnknown, dwFlags uint32) HRESULT
}

type IImageList2Vtbl struct {
	IImageListVtbl
	Resize               uintptr
	GetOriginalSize      uintptr
	SetOriginalSize      uintptr
	SetCallback          uintptr
	GetCallback          uintptr
	ForceImagePresent    uintptr
	DiscardImages        uintptr
	PreloadImages        uintptr
	GetStatistics        uintptr
	Initialize           uintptr
	Replace2             uintptr
	ReplaceFromImageList uintptr
}

type IImageList2 struct {
	IImageList
}

func (this *IImageList2) Vtbl() *IImageList2Vtbl {
	return (*IImageList2Vtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IImageList2) Resize(cxNewIconSize int32, cyNewIconSize int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Resize, uintptr(unsafe.Pointer(this)), uintptr(cxNewIconSize), uintptr(cyNewIconSize))
	return HRESULT(ret)
}

func (this *IImageList2) GetOriginalSize(iImage int32, dwFlags uint32, pcx *int32, pcy *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetOriginalSize, uintptr(unsafe.Pointer(this)), uintptr(iImage), uintptr(dwFlags), uintptr(unsafe.Pointer(pcx)), uintptr(unsafe.Pointer(pcy)))
	return HRESULT(ret)
}

func (this *IImageList2) SetOriginalSize(iImage int32, cx int32, cy int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetOriginalSize, uintptr(unsafe.Pointer(this)), uintptr(iImage), uintptr(cx), uintptr(cy))
	return HRESULT(ret)
}

func (this *IImageList2) SetCallback(punk *IUnknown) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetCallback, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(punk)))
	return HRESULT(ret)
}

func (this *IImageList2) GetCallback(riid *syscall.GUID, ppv unsafe.Pointer) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetCallback, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(riid)), uintptr(ppv))
	return HRESULT(ret)
}

func (this *IImageList2) ForceImagePresent(iImage int32, dwFlags uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().ForceImagePresent, uintptr(unsafe.Pointer(this)), uintptr(iImage), uintptr(dwFlags))
	return HRESULT(ret)
}

func (this *IImageList2) DiscardImages(iFirstImage int32, iLastImage int32, dwFlags uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().DiscardImages, uintptr(unsafe.Pointer(this)), uintptr(iFirstImage), uintptr(iLastImage), uintptr(dwFlags))
	return HRESULT(ret)
}

func (this *IImageList2) PreloadImages(pimldp *IMAGELISTDRAWPARAMS) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().PreloadImages, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pimldp)))
	return HRESULT(ret)
}

func (this *IImageList2) GetStatistics(pils *IMAGELISTSTATS) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetStatistics, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pils)))
	return HRESULT(ret)
}

func (this *IImageList2) Initialize(cx int32, cy int32, flags IMAGELIST_CREATION_FLAGS, cInitial int32, cGrow int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Initialize, uintptr(unsafe.Pointer(this)), uintptr(cx), uintptr(cy), uintptr(flags), uintptr(cInitial), uintptr(cGrow))
	return HRESULT(ret)
}

func (this *IImageList2) Replace2(i int32, hbmImage HBITMAP, hbmMask HBITMAP, punk *IUnknown, dwFlags uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Replace2, uintptr(unsafe.Pointer(this)), uintptr(i), hbmImage, hbmMask, uintptr(unsafe.Pointer(punk)), uintptr(dwFlags))
	return HRESULT(ret)
}

func (this *IImageList2) ReplaceFromImageList(i int32, pil *IImageList, iSrc int32, punk *IUnknown, dwFlags uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().ReplaceFromImageList, uintptr(unsafe.Pointer(this)), uintptr(i), uintptr(unsafe.Pointer(pil)), uintptr(iSrc), uintptr(unsafe.Pointer(punk)), uintptr(dwFlags))
	return HRESULT(ret)
}

var (
	pCreatePropertySheetPageA               uintptr
	pCreatePropertySheetPageW               uintptr
	pDestroyPropertySheetPage               uintptr
	pPropertySheetA                         uintptr
	pPropertySheetW                         uintptr
	pInitCommonControls                     uintptr
	pInitCommonControlsEx                   uintptr
	pImageList_Create                       uintptr
	pImageList_Destroy                      uintptr
	pImageList_GetImageCount                uintptr
	pImageList_SetImageCount                uintptr
	pImageList_Add                          uintptr
	pImageList_ReplaceIcon                  uintptr
	pImageList_SetBkColor                   uintptr
	pImageList_GetBkColor                   uintptr
	pImageList_SetOverlayImage              uintptr
	pImageList_Draw                         uintptr
	pImageList_Replace                      uintptr
	pImageList_AddMasked                    uintptr
	pImageList_DrawEx                       uintptr
	pImageList_DrawIndirect                 uintptr
	pImageList_Remove                       uintptr
	pImageList_GetIcon                      uintptr
	pImageList_LoadImageA                   uintptr
	pImageList_LoadImageW                   uintptr
	pImageList_Copy                         uintptr
	pImageList_BeginDrag                    uintptr
	pImageList_EndDrag                      uintptr
	pImageList_DragEnter                    uintptr
	pImageList_DragLeave                    uintptr
	pImageList_DragMove                     uintptr
	pImageList_SetDragCursorImage           uintptr
	pImageList_DragShowNolock               uintptr
	pImageList_GetDragImage                 uintptr
	pImageList_Read                         uintptr
	pImageList_Write                        uintptr
	pImageList_ReadEx                       uintptr
	pImageList_WriteEx                      uintptr
	pImageList_GetIconSize                  uintptr
	pImageList_SetIconSize                  uintptr
	pImageList_GetImageInfo                 uintptr
	pImageList_Merge                        uintptr
	pImageList_Duplicate                    uintptr
	pHIMAGELIST_QueryInterface              uintptr
	pCreateToolbarEx                        uintptr
	pCreateMappedBitmap                     uintptr
	pDrawStatusTextA                        uintptr
	pDrawStatusTextW                        uintptr
	pCreateStatusWindowA                    uintptr
	pCreateStatusWindowW                    uintptr
	pMenuHelp                               uintptr
	pShowHideMenuCtl                        uintptr
	pGetEffectiveClientRect                 uintptr
	pMakeDragList                           uintptr
	pDrawInsert                             uintptr
	pLBItemFromPt                           uintptr
	pCreateUpDownControl                    uintptr
	pTaskDialogIndirect                     uintptr
	pTaskDialog                             uintptr
	pInitMUILanguage                        uintptr
	pGetMUILanguage                         uintptr
	pDSA_Create                             uintptr
	pDSA_Destroy                            uintptr
	pDSA_DestroyCallback                    uintptr
	pDSA_DeleteItem                         uintptr
	pDSA_DeleteAllItems                     uintptr
	pDSA_EnumCallback                       uintptr
	pDSA_InsertItem                         uintptr
	pDSA_GetItemPtr                         uintptr
	pDSA_GetItem                            uintptr
	pDSA_SetItem                            uintptr
	pDSA_Clone                              uintptr
	pDSA_GetSize                            uintptr
	pDSA_Sort                               uintptr
	pDPA_Create                             uintptr
	pDPA_CreateEx                           uintptr
	pDPA_Clone                              uintptr
	pDPA_Destroy                            uintptr
	pDPA_DestroyCallback                    uintptr
	pDPA_DeletePtr                          uintptr
	pDPA_DeleteAllPtrs                      uintptr
	pDPA_EnumCallback                       uintptr
	pDPA_Grow                               uintptr
	pDPA_InsertPtr                          uintptr
	pDPA_SetPtr                             uintptr
	pDPA_GetPtr                             uintptr
	pDPA_GetPtrIndex                        uintptr
	pDPA_GetSize                            uintptr
	pDPA_Sort                               uintptr
	pDPA_LoadStream                         uintptr
	pDPA_SaveStream                         uintptr
	pDPA_Merge                              uintptr
	pDPA_Search                             uintptr
	pStr_SetPtrW                            uintptr
	pFlatSB_EnableScrollBar                 uintptr
	pFlatSB_ShowScrollBar                   uintptr
	pFlatSB_GetScrollRange                  uintptr
	pFlatSB_GetScrollInfo                   uintptr
	pFlatSB_GetScrollPos                    uintptr
	pFlatSB_GetScrollProp                   uintptr
	pFlatSB_SetScrollPos                    uintptr
	pFlatSB_SetScrollInfo                   uintptr
	pFlatSB_SetScrollRange                  uintptr
	pFlatSB_SetScrollProp                   uintptr
	pInitializeFlatSB                       uintptr
	pUninitializeFlatSB                     uintptr
	pLoadIconMetric                         uintptr
	pLoadIconWithScaleDown                  uintptr
	pDrawShadowText                         uintptr
	pImageList_CoCreateInstance             uintptr
	pBeginPanningFeedback                   uintptr
	pUpdatePanningFeedback                  uintptr
	pEndPanningFeedback                     uintptr
	pGetThemeAnimationProperty              uintptr
	pGetThemeAnimationTransform             uintptr
	pGetThemeTimingFunction                 uintptr
	pOpenThemeData                          uintptr
	pOpenThemeDataEx                        uintptr
	pCloseThemeData                         uintptr
	pDrawThemeBackground                    uintptr
	pDrawThemeBackgroundEx                  uintptr
	pDrawThemeText                          uintptr
	pGetThemeBackgroundContentRect          uintptr
	pGetThemeBackgroundExtent               uintptr
	pGetThemeBackgroundRegion               uintptr
	pGetThemePartSize                       uintptr
	pGetThemeTextExtent                     uintptr
	pGetThemeTextMetrics                    uintptr
	pHitTestThemeBackground                 uintptr
	pDrawThemeEdge                          uintptr
	pDrawThemeIcon                          uintptr
	pIsThemePartDefined                     uintptr
	pIsThemeBackgroundPartiallyTransparent  uintptr
	pGetThemeColor                          uintptr
	pGetThemeMetric                         uintptr
	pGetThemeString                         uintptr
	pGetThemeBool                           uintptr
	pGetThemeInt                            uintptr
	pGetThemeEnumValue                      uintptr
	pGetThemePosition                       uintptr
	pGetThemeFont                           uintptr
	pGetThemeRect                           uintptr
	pGetThemeMargins                        uintptr
	pGetThemeIntList                        uintptr
	pGetThemePropertyOrigin                 uintptr
	pSetWindowTheme                         uintptr
	pGetThemeFilename                       uintptr
	pGetThemeSysColor                       uintptr
	pGetThemeSysColorBrush                  uintptr
	pGetThemeSysBool                        uintptr
	pGetThemeSysSize                        uintptr
	pGetThemeSysFont                        uintptr
	pGetThemeSysString                      uintptr
	pGetThemeSysInt                         uintptr
	pIsThemeActive                          uintptr
	pIsAppThemed                            uintptr
	pGetWindowTheme                         uintptr
	pEnableThemeDialogTexture               uintptr
	pIsThemeDialogTextureEnabled            uintptr
	pGetThemeAppProperties                  uintptr
	pSetThemeAppProperties                  uintptr
	pGetCurrentThemeName                    uintptr
	pGetThemeDocumentationProperty          uintptr
	pDrawThemeParentBackground              uintptr
	pEnableTheming                          uintptr
	pDrawThemeParentBackgroundEx            uintptr
	pSetWindowThemeAttribute                uintptr
	pDrawThemeTextEx                        uintptr
	pGetThemeBitmap                         uintptr
	pGetThemeStream                         uintptr
	pBufferedPaintInit                      uintptr
	pBufferedPaintUnInit                    uintptr
	pBeginBufferedPaint                     uintptr
	pEndBufferedPaint                       uintptr
	pGetBufferedPaintTargetRect             uintptr
	pGetBufferedPaintTargetDC               uintptr
	pGetBufferedPaintDC                     uintptr
	pGetBufferedPaintBits                   uintptr
	pBufferedPaintClear                     uintptr
	pBufferedPaintSetAlpha                  uintptr
	pBufferedPaintStopAllAnimations         uintptr
	pBeginBufferedAnimation                 uintptr
	pEndBufferedAnimation                   uintptr
	pBufferedPaintRenderAnimation           uintptr
	pIsCompositionActive                    uintptr
	pGetThemeTransitionDuration             uintptr
	pCheckDlgButton                         uintptr
	pCheckRadioButton                       uintptr
	pIsDlgButtonChecked                     uintptr
	pIsCharLowerW                           uintptr
	pCreateSyntheticPointerDevice           uintptr
	pDestroySyntheticPointerDevice          uintptr
	pRegisterTouchHitTestingWindow          uintptr
	pEvaluateProximityToRect                uintptr
	pEvaluateProximityToPolygon             uintptr
	pPackTouchHitTestingProximityEvaluation uintptr
	pGetWindowFeedbackSetting               uintptr
	pSetWindowFeedbackSetting               uintptr
	pSetScrollPos                           uintptr
	pSetScrollRange                         uintptr
	pShowScrollBar                          uintptr
	pEnableScrollBar                        uintptr
	pDlgDirListA                            uintptr
	pDlgDirListW                            uintptr
	pDlgDirSelectExA                        uintptr
	pDlgDirSelectExW                        uintptr
	pDlgDirListComboBoxA                    uintptr
	pDlgDirListComboBoxW                    uintptr
	pDlgDirSelectComboBoxExA                uintptr
	pDlgDirSelectComboBoxExW                uintptr
	pSetScrollInfo                          uintptr
	pGetComboBoxInfo                        uintptr
	pGetListBoxInfo                         uintptr
	pRegisterPointerDeviceNotifications     uintptr
)

func CreatePropertySheetPageA(constPropSheetPagePointer *PROPSHEETPAGEA) HPROPSHEETPAGE {
	addr := lazyAddr(&pCreatePropertySheetPageA, libComctl32, "CreatePropertySheetPageA")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(constPropSheetPagePointer)))
	return ret
}

var CreatePropertySheetPage = CreatePropertySheetPageW

func CreatePropertySheetPageW(constPropSheetPagePointer *PROPSHEETPAGEW) HPROPSHEETPAGE {
	addr := lazyAddr(&pCreatePropertySheetPageW, libComctl32, "CreatePropertySheetPageW")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(constPropSheetPagePointer)))
	return ret
}

func DestroyPropertySheetPage(param0 HPROPSHEETPAGE) BOOL {
	addr := lazyAddr(&pDestroyPropertySheetPage, libComctl32, "DestroyPropertySheetPage")
	ret, _, _ := syscall.SyscallN(addr, param0)
	return BOOL(ret)
}

func PropertySheetA(param0 *PROPSHEETHEADERA_V2) (uintptr, WIN32_ERROR) {
	addr := lazyAddr(&pPropertySheetA, libComctl32, "PropertySheetA")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(param0)))
	return ret, WIN32_ERROR(err)
}

var PropertySheet = PropertySheetW

func PropertySheetW(param0 *PROPSHEETHEADERW_V2) (uintptr, WIN32_ERROR) {
	addr := lazyAddr(&pPropertySheetW, libComctl32, "PropertySheetW")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(param0)))
	return ret, WIN32_ERROR(err)
}

func InitCommonControls() {
	addr := lazyAddr(&pInitCommonControls, libComctl32, "InitCommonControls")
	syscall.SyscallN(addr)
}

func InitCommonControlsEx(picce *INITCOMMONCONTROLSEX) BOOL {
	addr := lazyAddr(&pInitCommonControlsEx, libComctl32, "InitCommonControlsEx")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(picce)))
	return BOOL(ret)
}

func ImageList_Create(cx int32, cy int32, flags IMAGELIST_CREATION_FLAGS, cInitial int32, cGrow int32) HIMAGELIST {
	addr := lazyAddr(&pImageList_Create, libComctl32, "ImageList_Create")
	ret, _, _ := syscall.SyscallN(addr, uintptr(cx), uintptr(cy), uintptr(flags), uintptr(cInitial), uintptr(cGrow))
	return ret
}

func ImageList_Destroy(himl HIMAGELIST) BOOL {
	addr := lazyAddr(&pImageList_Destroy, libComctl32, "ImageList_Destroy")
	ret, _, _ := syscall.SyscallN(addr, himl)
	return BOOL(ret)
}

func ImageList_GetImageCount(himl HIMAGELIST) int32 {
	addr := lazyAddr(&pImageList_GetImageCount, libComctl32, "ImageList_GetImageCount")
	ret, _, _ := syscall.SyscallN(addr, himl)
	return int32(ret)
}

func ImageList_SetImageCount(himl HIMAGELIST, uNewCount uint32) BOOL {
	addr := lazyAddr(&pImageList_SetImageCount, libComctl32, "ImageList_SetImageCount")
	ret, _, _ := syscall.SyscallN(addr, himl, uintptr(uNewCount))
	return BOOL(ret)
}

func ImageList_Add(himl HIMAGELIST, hbmImage HBITMAP, hbmMask HBITMAP) int32 {
	addr := lazyAddr(&pImageList_Add, libComctl32, "ImageList_Add")
	ret, _, _ := syscall.SyscallN(addr, himl, hbmImage, hbmMask)
	return int32(ret)
}

func ImageList_ReplaceIcon(himl HIMAGELIST, i int32, hicon HICON) int32 {
	addr := lazyAddr(&pImageList_ReplaceIcon, libComctl32, "ImageList_ReplaceIcon")
	ret, _, _ := syscall.SyscallN(addr, himl, uintptr(i), hicon)
	return int32(ret)
}

func ImageList_SetBkColor(himl HIMAGELIST, clrBk uint32) uint32 {
	addr := lazyAddr(&pImageList_SetBkColor, libComctl32, "ImageList_SetBkColor")
	ret, _, _ := syscall.SyscallN(addr, himl, uintptr(clrBk))
	return uint32(ret)
}

func ImageList_GetBkColor(himl HIMAGELIST) uint32 {
	addr := lazyAddr(&pImageList_GetBkColor, libComctl32, "ImageList_GetBkColor")
	ret, _, _ := syscall.SyscallN(addr, himl)
	return uint32(ret)
}

func ImageList_SetOverlayImage(himl HIMAGELIST, iImage int32, iOverlay int32) BOOL {
	addr := lazyAddr(&pImageList_SetOverlayImage, libComctl32, "ImageList_SetOverlayImage")
	ret, _, _ := syscall.SyscallN(addr, himl, uintptr(iImage), uintptr(iOverlay))
	return BOOL(ret)
}

func ImageList_Draw(himl HIMAGELIST, i int32, hdcDst HDC, x int32, y int32, fStyle IMAGE_LIST_DRAW_STYLE) BOOL {
	addr := lazyAddr(&pImageList_Draw, libComctl32, "ImageList_Draw")
	ret, _, _ := syscall.SyscallN(addr, himl, uintptr(i), hdcDst, uintptr(x), uintptr(y), uintptr(fStyle))
	return BOOL(ret)
}

func ImageList_Replace(himl HIMAGELIST, i int32, hbmImage HBITMAP, hbmMask HBITMAP) BOOL {
	addr := lazyAddr(&pImageList_Replace, libComctl32, "ImageList_Replace")
	ret, _, _ := syscall.SyscallN(addr, himl, uintptr(i), hbmImage, hbmMask)
	return BOOL(ret)
}

func ImageList_AddMasked(himl HIMAGELIST, hbmImage HBITMAP, crMask uint32) int32 {
	addr := lazyAddr(&pImageList_AddMasked, libComctl32, "ImageList_AddMasked")
	ret, _, _ := syscall.SyscallN(addr, himl, hbmImage, uintptr(crMask))
	return int32(ret)
}

func ImageList_DrawEx(himl HIMAGELIST, i int32, hdcDst HDC, x int32, y int32, dx int32, dy int32, rgbBk uint32, rgbFg uint32, fStyle IMAGE_LIST_DRAW_STYLE) BOOL {
	addr := lazyAddr(&pImageList_DrawEx, libComctl32, "ImageList_DrawEx")
	ret, _, _ := syscall.SyscallN(addr, himl, uintptr(i), hdcDst, uintptr(x), uintptr(y), uintptr(dx), uintptr(dy), uintptr(rgbBk), uintptr(rgbFg), uintptr(fStyle))
	return BOOL(ret)
}

func ImageList_DrawIndirect(pimldp *IMAGELISTDRAWPARAMS) BOOL {
	addr := lazyAddr(&pImageList_DrawIndirect, libComctl32, "ImageList_DrawIndirect")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pimldp)))
	return BOOL(ret)
}

func ImageList_Remove(himl HIMAGELIST, i int32) BOOL {
	addr := lazyAddr(&pImageList_Remove, libComctl32, "ImageList_Remove")
	ret, _, _ := syscall.SyscallN(addr, himl, uintptr(i))
	return BOOL(ret)
}

func ImageList_GetIcon(himl HIMAGELIST, i int32, flags uint32) HICON {
	addr := lazyAddr(&pImageList_GetIcon, libComctl32, "ImageList_GetIcon")
	ret, _, _ := syscall.SyscallN(addr, himl, uintptr(i), uintptr(flags))
	return ret
}

func ImageList_LoadImageA(hi HINSTANCE, lpbmp PSTR, cx int32, cGrow int32, crMask uint32, uType uint32, uFlags IMAGE_FLAGS) HIMAGELIST {
	addr := lazyAddr(&pImageList_LoadImageA, libComctl32, "ImageList_LoadImageA")
	ret, _, _ := syscall.SyscallN(addr, hi, uintptr(unsafe.Pointer(lpbmp)), uintptr(cx), uintptr(cGrow), uintptr(crMask), uintptr(uType), uintptr(uFlags))
	return ret
}

var ImageList_LoadImage = ImageList_LoadImageW

func ImageList_LoadImageW(hi HINSTANCE, lpbmp PWSTR, cx int32, cGrow int32, crMask uint32, uType uint32, uFlags IMAGE_FLAGS) HIMAGELIST {
	addr := lazyAddr(&pImageList_LoadImageW, libComctl32, "ImageList_LoadImageW")
	ret, _, _ := syscall.SyscallN(addr, hi, uintptr(unsafe.Pointer(lpbmp)), uintptr(cx), uintptr(cGrow), uintptr(crMask), uintptr(uType), uintptr(uFlags))
	return ret
}

func ImageList_Copy(himlDst HIMAGELIST, iDst int32, himlSrc HIMAGELIST, iSrc int32, uFlags IMAGE_LIST_COPY_FLAGS) BOOL {
	addr := lazyAddr(&pImageList_Copy, libComctl32, "ImageList_Copy")
	ret, _, _ := syscall.SyscallN(addr, himlDst, uintptr(iDst), himlSrc, uintptr(iSrc), uintptr(uFlags))
	return BOOL(ret)
}

func ImageList_BeginDrag(himlTrack HIMAGELIST, iTrack int32, dxHotspot int32, dyHotspot int32) BOOL {
	addr := lazyAddr(&pImageList_BeginDrag, libComctl32, "ImageList_BeginDrag")
	ret, _, _ := syscall.SyscallN(addr, himlTrack, uintptr(iTrack), uintptr(dxHotspot), uintptr(dyHotspot))
	return BOOL(ret)
}

func ImageList_EndDrag() {
	addr := lazyAddr(&pImageList_EndDrag, libComctl32, "ImageList_EndDrag")
	syscall.SyscallN(addr)
}

func ImageList_DragEnter(hwndLock HWND, x int32, y int32) BOOL {
	addr := lazyAddr(&pImageList_DragEnter, libComctl32, "ImageList_DragEnter")
	ret, _, _ := syscall.SyscallN(addr, hwndLock, uintptr(x), uintptr(y))
	return BOOL(ret)
}

func ImageList_DragLeave(hwndLock HWND) BOOL {
	addr := lazyAddr(&pImageList_DragLeave, libComctl32, "ImageList_DragLeave")
	ret, _, _ := syscall.SyscallN(addr, hwndLock)
	return BOOL(ret)
}

func ImageList_DragMove(x int32, y int32) BOOL {
	addr := lazyAddr(&pImageList_DragMove, libComctl32, "ImageList_DragMove")
	ret, _, _ := syscall.SyscallN(addr, uintptr(x), uintptr(y))
	return BOOL(ret)
}

func ImageList_SetDragCursorImage(himlDrag HIMAGELIST, iDrag int32, dxHotspot int32, dyHotspot int32) BOOL {
	addr := lazyAddr(&pImageList_SetDragCursorImage, libComctl32, "ImageList_SetDragCursorImage")
	ret, _, _ := syscall.SyscallN(addr, himlDrag, uintptr(iDrag), uintptr(dxHotspot), uintptr(dyHotspot))
	return BOOL(ret)
}

func ImageList_DragShowNolock(fShow BOOL) BOOL {
	addr := lazyAddr(&pImageList_DragShowNolock, libComctl32, "ImageList_DragShowNolock")
	ret, _, _ := syscall.SyscallN(addr, uintptr(fShow))
	return BOOL(ret)
}

func ImageList_GetDragImage(ppt *POINT, pptHotspot *POINT) HIMAGELIST {
	addr := lazyAddr(&pImageList_GetDragImage, libComctl32, "ImageList_GetDragImage")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(ppt)), uintptr(unsafe.Pointer(pptHotspot)))
	return ret
}

func ImageList_Read(pstm *IStream) HIMAGELIST {
	addr := lazyAddr(&pImageList_Read, libComctl32, "ImageList_Read")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pstm)))
	return ret
}

func ImageList_Write(himl HIMAGELIST, pstm *IStream) BOOL {
	addr := lazyAddr(&pImageList_Write, libComctl32, "ImageList_Write")
	ret, _, _ := syscall.SyscallN(addr, himl, uintptr(unsafe.Pointer(pstm)))
	return BOOL(ret)
}

func ImageList_ReadEx(dwFlags uint32, pstm *IStream, riid *syscall.GUID, ppv unsafe.Pointer) HRESULT {
	addr := lazyAddr(&pImageList_ReadEx, libComctl32, "ImageList_ReadEx")
	ret, _, _ := syscall.SyscallN(addr, uintptr(dwFlags), uintptr(unsafe.Pointer(pstm)), uintptr(unsafe.Pointer(riid)), uintptr(ppv))
	return HRESULT(ret)
}

func ImageList_WriteEx(himl HIMAGELIST, dwFlags uint32, pstm *IStream) HRESULT {
	addr := lazyAddr(&pImageList_WriteEx, libComctl32, "ImageList_WriteEx")
	ret, _, _ := syscall.SyscallN(addr, himl, uintptr(dwFlags), uintptr(unsafe.Pointer(pstm)))
	return HRESULT(ret)
}

func ImageList_GetIconSize(himl HIMAGELIST, cx *int32, cy *int32) BOOL {
	addr := lazyAddr(&pImageList_GetIconSize, libComctl32, "ImageList_GetIconSize")
	ret, _, _ := syscall.SyscallN(addr, himl, uintptr(unsafe.Pointer(cx)), uintptr(unsafe.Pointer(cy)))
	return BOOL(ret)
}

func ImageList_SetIconSize(himl HIMAGELIST, cx int32, cy int32) BOOL {
	addr := lazyAddr(&pImageList_SetIconSize, libComctl32, "ImageList_SetIconSize")
	ret, _, _ := syscall.SyscallN(addr, himl, uintptr(cx), uintptr(cy))
	return BOOL(ret)
}

func ImageList_GetImageInfo(himl HIMAGELIST, i int32, pImageInfo *IMAGEINFO) BOOL {
	addr := lazyAddr(&pImageList_GetImageInfo, libComctl32, "ImageList_GetImageInfo")
	ret, _, _ := syscall.SyscallN(addr, himl, uintptr(i), uintptr(unsafe.Pointer(pImageInfo)))
	return BOOL(ret)
}

func ImageList_Merge(himl1 HIMAGELIST, i1 int32, himl2 HIMAGELIST, i2 int32, dx int32, dy int32) HIMAGELIST {
	addr := lazyAddr(&pImageList_Merge, libComctl32, "ImageList_Merge")
	ret, _, _ := syscall.SyscallN(addr, himl1, uintptr(i1), himl2, uintptr(i2), uintptr(dx), uintptr(dy))
	return ret
}

func ImageList_Duplicate(himl HIMAGELIST) HIMAGELIST {
	addr := lazyAddr(&pImageList_Duplicate, libComctl32, "ImageList_Duplicate")
	ret, _, _ := syscall.SyscallN(addr, himl)
	return ret
}

func HIMAGELIST_QueryInterface(himl HIMAGELIST, riid *syscall.GUID, ppv unsafe.Pointer) HRESULT {
	addr := lazyAddr(&pHIMAGELIST_QueryInterface, libComctl32, "HIMAGELIST_QueryInterface")
	ret, _, _ := syscall.SyscallN(addr, himl, uintptr(unsafe.Pointer(riid)), uintptr(ppv))
	return HRESULT(ret)
}

func CreateToolbarEx(hwnd HWND, ws uint32, wID uint32, nBitmaps int32, hBMInst HINSTANCE, wBMID uintptr, lpButtons *TBBUTTON, iNumButtons int32, dxButton int32, dyButton int32, dxBitmap int32, dyBitmap int32, uStructSize uint32) (HWND, WIN32_ERROR) {
	addr := lazyAddr(&pCreateToolbarEx, libComctl32, "CreateToolbarEx")
	ret, _, err := syscall.SyscallN(addr, hwnd, uintptr(ws), uintptr(wID), uintptr(nBitmaps), hBMInst, wBMID, uintptr(unsafe.Pointer(lpButtons)), uintptr(iNumButtons), uintptr(dxButton), uintptr(dyButton), uintptr(dxBitmap), uintptr(dyBitmap), uintptr(uStructSize))
	return ret, WIN32_ERROR(err)
}

func CreateMappedBitmap(hInstance HINSTANCE, idBitmap uintptr, wFlags uint32, lpColorMap *COLORMAP, iNumMaps int32) (HBITMAP, WIN32_ERROR) {
	addr := lazyAddr(&pCreateMappedBitmap, libComctl32, "CreateMappedBitmap")
	ret, _, err := syscall.SyscallN(addr, hInstance, idBitmap, uintptr(wFlags), uintptr(unsafe.Pointer(lpColorMap)), uintptr(iNumMaps))
	return ret, WIN32_ERROR(err)
}

func DrawStatusTextA(hDC HDC, lprc *RECT, pszText PSTR, uFlags uint32) {
	addr := lazyAddr(&pDrawStatusTextA, libComctl32, "DrawStatusTextA")
	syscall.SyscallN(addr, hDC, uintptr(unsafe.Pointer(lprc)), uintptr(unsafe.Pointer(pszText)), uintptr(uFlags))
}

var DrawStatusText = DrawStatusTextW

func DrawStatusTextW(hDC HDC, lprc *RECT, pszText PWSTR, uFlags uint32) {
	addr := lazyAddr(&pDrawStatusTextW, libComctl32, "DrawStatusTextW")
	syscall.SyscallN(addr, hDC, uintptr(unsafe.Pointer(lprc)), uintptr(unsafe.Pointer(pszText)), uintptr(uFlags))
}

func CreateStatusWindowA(style int32, lpszText PSTR, hwndParent HWND, wID uint32) (HWND, WIN32_ERROR) {
	addr := lazyAddr(&pCreateStatusWindowA, libComctl32, "CreateStatusWindowA")
	ret, _, err := syscall.SyscallN(addr, uintptr(style), uintptr(unsafe.Pointer(lpszText)), hwndParent, uintptr(wID))
	return ret, WIN32_ERROR(err)
}

var CreateStatusWindow = CreateStatusWindowW

func CreateStatusWindowW(style int32, lpszText PWSTR, hwndParent HWND, wID uint32) (HWND, WIN32_ERROR) {
	addr := lazyAddr(&pCreateStatusWindowW, libComctl32, "CreateStatusWindowW")
	ret, _, err := syscall.SyscallN(addr, uintptr(style), uintptr(unsafe.Pointer(lpszText)), hwndParent, uintptr(wID))
	return ret, WIN32_ERROR(err)
}

func MenuHelp(uMsg uint32, wParam WPARAM, lParam LPARAM, hMainMenu HMENU, hInst HINSTANCE, hwndStatus HWND, lpwIDs *uint32) {
	addr := lazyAddr(&pMenuHelp, libComctl32, "MenuHelp")
	syscall.SyscallN(addr, uintptr(uMsg), wParam, lParam, hMainMenu, hInst, hwndStatus, uintptr(unsafe.Pointer(lpwIDs)))
}

func ShowHideMenuCtl(hWnd HWND, uFlags uintptr, lpInfo *int32) BOOL {
	addr := lazyAddr(&pShowHideMenuCtl, libComctl32, "ShowHideMenuCtl")
	ret, _, _ := syscall.SyscallN(addr, hWnd, uFlags, uintptr(unsafe.Pointer(lpInfo)))
	return BOOL(ret)
}

func GetEffectiveClientRect(hWnd HWND, lprc *RECT, lpInfo *int32) {
	addr := lazyAddr(&pGetEffectiveClientRect, libComctl32, "GetEffectiveClientRect")
	syscall.SyscallN(addr, hWnd, uintptr(unsafe.Pointer(lprc)), uintptr(unsafe.Pointer(lpInfo)))
}

func MakeDragList(hLB HWND) BOOL {
	addr := lazyAddr(&pMakeDragList, libComctl32, "MakeDragList")
	ret, _, _ := syscall.SyscallN(addr, hLB)
	return BOOL(ret)
}

func DrawInsert(handParent HWND, hLB HWND, nItem int32) {
	addr := lazyAddr(&pDrawInsert, libComctl32, "DrawInsert")
	syscall.SyscallN(addr, handParent, hLB, uintptr(nItem))
}

func LBItemFromPt(hLB HWND, pt POINT, bAutoScroll BOOL) int32 {
	addr := lazyAddr(&pLBItemFromPt, libComctl32, "LBItemFromPt")
	ret, _, _ := syscall.SyscallN(addr, hLB, *(*uintptr)(unsafe.Pointer(&pt)), uintptr(bAutoScroll))
	return int32(ret)
}

func CreateUpDownControl(dwStyle uint32, x int32, y int32, cx int32, cy int32, hParent HWND, nID int32, hInst HINSTANCE, hBuddy HWND, nUpper int32, nLower int32, nPos int32) HWND {
	addr := lazyAddr(&pCreateUpDownControl, libComctl32, "CreateUpDownControl")
	ret, _, _ := syscall.SyscallN(addr, uintptr(dwStyle), uintptr(x), uintptr(y), uintptr(cx), uintptr(cy), hParent, uintptr(nID), hInst, hBuddy, uintptr(nUpper), uintptr(nLower), uintptr(nPos))
	return ret
}

func TaskDialogIndirect(pTaskConfig *TASKDIALOGCONFIG, pnButton *int32, pnRadioButton *int32, pfVerificationFlagChecked *BOOL) HRESULT {
	addr := lazyAddr(&pTaskDialogIndirect, libComctl32, "TaskDialogIndirect")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pTaskConfig)), uintptr(unsafe.Pointer(pnButton)), uintptr(unsafe.Pointer(pnRadioButton)), uintptr(unsafe.Pointer(pfVerificationFlagChecked)))
	return HRESULT(ret)
}

func TaskDialog(hwndOwner HWND, hInstance HINSTANCE, pszWindowTitle PWSTR, pszMainInstruction PWSTR, pszContent PWSTR, dwCommonButtons TASKDIALOG_COMMON_BUTTON_FLAGS, pszIcon PWSTR, pnButton *int32) HRESULT {
	addr := lazyAddr(&pTaskDialog, libComctl32, "TaskDialog")
	ret, _, _ := syscall.SyscallN(addr, hwndOwner, hInstance, uintptr(unsafe.Pointer(pszWindowTitle)), uintptr(unsafe.Pointer(pszMainInstruction)), uintptr(unsafe.Pointer(pszContent)), uintptr(dwCommonButtons), uintptr(unsafe.Pointer(pszIcon)), uintptr(unsafe.Pointer(pnButton)))
	return HRESULT(ret)
}

func InitMUILanguage(uiLang uint16) {
	addr := lazyAddr(&pInitMUILanguage, libComctl32, "InitMUILanguage")
	syscall.SyscallN(addr, uintptr(uiLang))
}

func GetMUILanguage() uint16 {
	addr := lazyAddr(&pGetMUILanguage, libComctl32, "GetMUILanguage")
	ret, _, _ := syscall.SyscallN(addr)
	return uint16(ret)
}

func DSA_Create(cbItem int32, cItemGrow int32) HDSA {
	addr := lazyAddr(&pDSA_Create, libComctl32, "DSA_Create")
	ret, _, _ := syscall.SyscallN(addr, uintptr(cbItem), uintptr(cItemGrow))
	return ret
}

func DSA_Destroy(hdsa HDSA) BOOL {
	addr := lazyAddr(&pDSA_Destroy, libComctl32, "DSA_Destroy")
	ret, _, _ := syscall.SyscallN(addr, hdsa)
	return BOOL(ret)
}

func DSA_DestroyCallback(hdsa HDSA, pfnCB PFNDAENUMCALLBACK, pData unsafe.Pointer) {
	addr := lazyAddr(&pDSA_DestroyCallback, libComctl32, "DSA_DestroyCallback")
	syscall.SyscallN(addr, hdsa, pfnCB, uintptr(pData))
}

func DSA_DeleteItem(hdsa HDSA, i int32) BOOL {
	addr := lazyAddr(&pDSA_DeleteItem, libComctl32, "DSA_DeleteItem")
	ret, _, _ := syscall.SyscallN(addr, hdsa, uintptr(i))
	return BOOL(ret)
}

func DSA_DeleteAllItems(hdsa HDSA) BOOL {
	addr := lazyAddr(&pDSA_DeleteAllItems, libComctl32, "DSA_DeleteAllItems")
	ret, _, _ := syscall.SyscallN(addr, hdsa)
	return BOOL(ret)
}

func DSA_EnumCallback(hdsa HDSA, pfnCB PFNDAENUMCALLBACK, pData unsafe.Pointer) {
	addr := lazyAddr(&pDSA_EnumCallback, libComctl32, "DSA_EnumCallback")
	syscall.SyscallN(addr, hdsa, pfnCB, uintptr(pData))
}

func DSA_InsertItem(hdsa HDSA, i int32, pitem unsafe.Pointer) int32 {
	addr := lazyAddr(&pDSA_InsertItem, libComctl32, "DSA_InsertItem")
	ret, _, _ := syscall.SyscallN(addr, hdsa, uintptr(i), uintptr(pitem))
	return int32(ret)
}

func DSA_GetItemPtr(hdsa HDSA, i int32) unsafe.Pointer {
	addr := lazyAddr(&pDSA_GetItemPtr, libComctl32, "DSA_GetItemPtr")
	ret, _, _ := syscall.SyscallN(addr, hdsa, uintptr(i))
	return (unsafe.Pointer)(ret)
}

func DSA_GetItem(hdsa HDSA, i int32, pitem unsafe.Pointer) BOOL {
	addr := lazyAddr(&pDSA_GetItem, libComctl32, "DSA_GetItem")
	ret, _, _ := syscall.SyscallN(addr, hdsa, uintptr(i), uintptr(pitem))
	return BOOL(ret)
}

func DSA_SetItem(hdsa HDSA, i int32, pitem unsafe.Pointer) BOOL {
	addr := lazyAddr(&pDSA_SetItem, libComctl32, "DSA_SetItem")
	ret, _, _ := syscall.SyscallN(addr, hdsa, uintptr(i), uintptr(pitem))
	return BOOL(ret)
}

func DSA_Clone(hdsa HDSA) HDSA {
	addr := lazyAddr(&pDSA_Clone, libComctl32, "DSA_Clone")
	ret, _, _ := syscall.SyscallN(addr, hdsa)
	return ret
}

func DSA_GetSize(hdsa HDSA) uint64 {
	addr := lazyAddr(&pDSA_GetSize, libComctl32, "DSA_GetSize")
	ret, _, _ := syscall.SyscallN(addr, hdsa)
	return uint64(ret)
}

func DSA_Sort(pdsa HDSA, pfnCompare PFNDACOMPARE, lParam LPARAM) BOOL {
	addr := lazyAddr(&pDSA_Sort, libComctl32, "DSA_Sort")
	ret, _, _ := syscall.SyscallN(addr, pdsa, pfnCompare, lParam)
	return BOOL(ret)
}

func DPA_Create(cItemGrow int32) HDPA {
	addr := lazyAddr(&pDPA_Create, libComctl32, "DPA_Create")
	ret, _, _ := syscall.SyscallN(addr, uintptr(cItemGrow))
	return ret
}

func DPA_CreateEx(cpGrow int32, hheap HANDLE) HDPA {
	addr := lazyAddr(&pDPA_CreateEx, libComctl32, "DPA_CreateEx")
	ret, _, _ := syscall.SyscallN(addr, uintptr(cpGrow), hheap)
	return ret
}

func DPA_Clone(hdpa HDPA, hdpaNew HDPA) HDPA {
	addr := lazyAddr(&pDPA_Clone, libComctl32, "DPA_Clone")
	ret, _, _ := syscall.SyscallN(addr, hdpa, hdpaNew)
	return ret
}

func DPA_Destroy(hdpa HDPA) BOOL {
	addr := lazyAddr(&pDPA_Destroy, libComctl32, "DPA_Destroy")
	ret, _, _ := syscall.SyscallN(addr, hdpa)
	return BOOL(ret)
}

func DPA_DestroyCallback(hdpa HDPA, pfnCB PFNDAENUMCALLBACK, pData unsafe.Pointer) {
	addr := lazyAddr(&pDPA_DestroyCallback, libComctl32, "DPA_DestroyCallback")
	syscall.SyscallN(addr, hdpa, pfnCB, uintptr(pData))
}

func DPA_DeletePtr(hdpa HDPA, i int32) unsafe.Pointer {
	addr := lazyAddr(&pDPA_DeletePtr, libComctl32, "DPA_DeletePtr")
	ret, _, _ := syscall.SyscallN(addr, hdpa, uintptr(i))
	return (unsafe.Pointer)(ret)
}

func DPA_DeleteAllPtrs(hdpa HDPA) BOOL {
	addr := lazyAddr(&pDPA_DeleteAllPtrs, libComctl32, "DPA_DeleteAllPtrs")
	ret, _, _ := syscall.SyscallN(addr, hdpa)
	return BOOL(ret)
}

func DPA_EnumCallback(hdpa HDPA, pfnCB PFNDAENUMCALLBACK, pData unsafe.Pointer) {
	addr := lazyAddr(&pDPA_EnumCallback, libComctl32, "DPA_EnumCallback")
	syscall.SyscallN(addr, hdpa, pfnCB, uintptr(pData))
}

func DPA_Grow(pdpa HDPA, cp int32) BOOL {
	addr := lazyAddr(&pDPA_Grow, libComctl32, "DPA_Grow")
	ret, _, _ := syscall.SyscallN(addr, pdpa, uintptr(cp))
	return BOOL(ret)
}

func DPA_InsertPtr(hdpa HDPA, i int32, p unsafe.Pointer) int32 {
	addr := lazyAddr(&pDPA_InsertPtr, libComctl32, "DPA_InsertPtr")
	ret, _, _ := syscall.SyscallN(addr, hdpa, uintptr(i), uintptr(p))
	return int32(ret)
}

func DPA_SetPtr(hdpa HDPA, i int32, p unsafe.Pointer) BOOL {
	addr := lazyAddr(&pDPA_SetPtr, libComctl32, "DPA_SetPtr")
	ret, _, _ := syscall.SyscallN(addr, hdpa, uintptr(i), uintptr(p))
	return BOOL(ret)
}

func DPA_GetPtr(hdpa HDPA, i uintptr) unsafe.Pointer {
	addr := lazyAddr(&pDPA_GetPtr, libComctl32, "DPA_GetPtr")
	ret, _, _ := syscall.SyscallN(addr, hdpa, i)
	return (unsafe.Pointer)(ret)
}

func DPA_GetPtrIndex(hdpa HDPA, p unsafe.Pointer) int32 {
	addr := lazyAddr(&pDPA_GetPtrIndex, libComctl32, "DPA_GetPtrIndex")
	ret, _, _ := syscall.SyscallN(addr, hdpa, uintptr(p))
	return int32(ret)
}

func DPA_GetSize(hdpa HDPA) uint64 {
	addr := lazyAddr(&pDPA_GetSize, libComctl32, "DPA_GetSize")
	ret, _, _ := syscall.SyscallN(addr, hdpa)
	return uint64(ret)
}

func DPA_Sort(hdpa HDPA, pfnCompare PFNDACOMPARE, lParam LPARAM) BOOL {
	addr := lazyAddr(&pDPA_Sort, libComctl32, "DPA_Sort")
	ret, _, _ := syscall.SyscallN(addr, hdpa, pfnCompare, lParam)
	return BOOL(ret)
}

func DPA_LoadStream(phdpa *HDPA, pfn PFNDPASTREAM, pstream *IStream, pvInstData unsafe.Pointer) HRESULT {
	addr := lazyAddr(&pDPA_LoadStream, libComctl32, "DPA_LoadStream")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(phdpa)), pfn, uintptr(unsafe.Pointer(pstream)), uintptr(pvInstData))
	return HRESULT(ret)
}

func DPA_SaveStream(hdpa HDPA, pfn PFNDPASTREAM, pstream *IStream, pvInstData unsafe.Pointer) HRESULT {
	addr := lazyAddr(&pDPA_SaveStream, libComctl32, "DPA_SaveStream")
	ret, _, _ := syscall.SyscallN(addr, hdpa, pfn, uintptr(unsafe.Pointer(pstream)), uintptr(pvInstData))
	return HRESULT(ret)
}

func DPA_Merge(hdpaDest HDPA, hdpaSrc HDPA, dwFlags uint32, pfnCompare PFNDACOMPARE, pfnMerge PFNDPAMERGE, lParam LPARAM) BOOL {
	addr := lazyAddr(&pDPA_Merge, libComctl32, "DPA_Merge")
	ret, _, _ := syscall.SyscallN(addr, hdpaDest, hdpaSrc, uintptr(dwFlags), pfnCompare, pfnMerge, lParam)
	return BOOL(ret)
}

func DPA_Search(hdpa HDPA, pFind unsafe.Pointer, iStart int32, pfnCompare PFNDACOMPARE, lParam LPARAM, options uint32) int32 {
	addr := lazyAddr(&pDPA_Search, libComctl32, "DPA_Search")
	ret, _, _ := syscall.SyscallN(addr, hdpa, uintptr(pFind), uintptr(iStart), pfnCompare, lParam, uintptr(options))
	return int32(ret)
}

func Str_SetPtrW(ppsz *PWSTR, psz PWSTR) BOOL {
	addr := lazyAddr(&pStr_SetPtrW, libComctl32, "Str_SetPtrW")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(ppsz)), uintptr(unsafe.Pointer(psz)))
	return BOOL(ret)
}

func FlatSB_EnableScrollBar(param0 HWND, param1 int32, param2 uint32) BOOL {
	addr := lazyAddr(&pFlatSB_EnableScrollBar, libComctl32, "FlatSB_EnableScrollBar")
	ret, _, _ := syscall.SyscallN(addr, param0, uintptr(param1), uintptr(param2))
	return BOOL(ret)
}

func FlatSB_ShowScrollBar(param0 HWND, code SCROLLBAR_CONSTANTS, param2 BOOL) BOOL {
	addr := lazyAddr(&pFlatSB_ShowScrollBar, libComctl32, "FlatSB_ShowScrollBar")
	ret, _, _ := syscall.SyscallN(addr, param0, uintptr(code), uintptr(param2))
	return BOOL(ret)
}

func FlatSB_GetScrollRange(param0 HWND, code SCROLLBAR_CONSTANTS, param2 *int32, param3 *int32) BOOL {
	addr := lazyAddr(&pFlatSB_GetScrollRange, libComctl32, "FlatSB_GetScrollRange")
	ret, _, _ := syscall.SyscallN(addr, param0, uintptr(code), uintptr(unsafe.Pointer(param2)), uintptr(unsafe.Pointer(param3)))
	return BOOL(ret)
}

func FlatSB_GetScrollInfo(param0 HWND, code SCROLLBAR_CONSTANTS, param2 *SCROLLINFO) BOOL {
	addr := lazyAddr(&pFlatSB_GetScrollInfo, libComctl32, "FlatSB_GetScrollInfo")
	ret, _, _ := syscall.SyscallN(addr, param0, uintptr(code), uintptr(unsafe.Pointer(param2)))
	return BOOL(ret)
}

func FlatSB_GetScrollPos(param0 HWND, code SCROLLBAR_CONSTANTS) int32 {
	addr := lazyAddr(&pFlatSB_GetScrollPos, libComctl32, "FlatSB_GetScrollPos")
	ret, _, _ := syscall.SyscallN(addr, param0, uintptr(code))
	return int32(ret)
}

func FlatSB_GetScrollProp(param0 HWND, propIndex WSB_PROP, param2 *int32) BOOL {
	addr := lazyAddr(&pFlatSB_GetScrollProp, libComctl32, "FlatSB_GetScrollProp")
	ret, _, _ := syscall.SyscallN(addr, param0, uintptr(propIndex), uintptr(unsafe.Pointer(param2)))
	return BOOL(ret)
}

func FlatSB_SetScrollPos(param0 HWND, code SCROLLBAR_CONSTANTS, pos int32, fRedraw BOOL) int32 {
	addr := lazyAddr(&pFlatSB_SetScrollPos, libComctl32, "FlatSB_SetScrollPos")
	ret, _, _ := syscall.SyscallN(addr, param0, uintptr(code), uintptr(pos), uintptr(fRedraw))
	return int32(ret)
}

func FlatSB_SetScrollInfo(param0 HWND, code SCROLLBAR_CONSTANTS, psi *SCROLLINFO, fRedraw BOOL) int32 {
	addr := lazyAddr(&pFlatSB_SetScrollInfo, libComctl32, "FlatSB_SetScrollInfo")
	ret, _, _ := syscall.SyscallN(addr, param0, uintptr(code), uintptr(unsafe.Pointer(psi)), uintptr(fRedraw))
	return int32(ret)
}

func FlatSB_SetScrollRange(param0 HWND, code SCROLLBAR_CONSTANTS, min int32, max int32, fRedraw BOOL) int32 {
	addr := lazyAddr(&pFlatSB_SetScrollRange, libComctl32, "FlatSB_SetScrollRange")
	ret, _, _ := syscall.SyscallN(addr, param0, uintptr(code), uintptr(min), uintptr(max), uintptr(fRedraw))
	return int32(ret)
}

func FlatSB_SetScrollProp(param0 HWND, index WSB_PROP, newValue uintptr, param3 BOOL) BOOL {
	addr := lazyAddr(&pFlatSB_SetScrollProp, libComctl32, "FlatSB_SetScrollProp")
	ret, _, _ := syscall.SyscallN(addr, param0, uintptr(index), newValue, uintptr(param3))
	return BOOL(ret)
}

func InitializeFlatSB(param0 HWND) BOOL {
	addr := lazyAddr(&pInitializeFlatSB, libComctl32, "InitializeFlatSB")
	ret, _, _ := syscall.SyscallN(addr, param0)
	return BOOL(ret)
}

func UninitializeFlatSB(param0 HWND) HRESULT {
	addr := lazyAddr(&pUninitializeFlatSB, libComctl32, "UninitializeFlatSB")
	ret, _, _ := syscall.SyscallN(addr, param0)
	return HRESULT(ret)
}

func LoadIconMetric(hinst HINSTANCE, pszName PWSTR, lims LI_METRIC_, phico *HICON) HRESULT {
	addr := lazyAddr(&pLoadIconMetric, libComctl32, "LoadIconMetric")
	ret, _, _ := syscall.SyscallN(addr, hinst, uintptr(unsafe.Pointer(pszName)), uintptr(lims), uintptr(unsafe.Pointer(phico)))
	return HRESULT(ret)
}

func LoadIconWithScaleDown(hinst HINSTANCE, pszName PWSTR, cx int32, cy int32, phico *HICON) HRESULT {
	addr := lazyAddr(&pLoadIconWithScaleDown, libComctl32, "LoadIconWithScaleDown")
	ret, _, _ := syscall.SyscallN(addr, hinst, uintptr(unsafe.Pointer(pszName)), uintptr(cx), uintptr(cy), uintptr(unsafe.Pointer(phico)))
	return HRESULT(ret)
}

func DrawShadowText(hdc HDC, pszText PWSTR, cch uint32, prc *RECT, dwFlags uint32, crText uint32, crShadow uint32, ixOffset int32, iyOffset int32) int32 {
	addr := lazyAddr(&pDrawShadowText, libComctl32, "DrawShadowText")
	ret, _, _ := syscall.SyscallN(addr, hdc, uintptr(unsafe.Pointer(pszText)), uintptr(cch), uintptr(unsafe.Pointer(prc)), uintptr(dwFlags), uintptr(crText), uintptr(crShadow), uintptr(ixOffset), uintptr(iyOffset))
	return int32(ret)
}

func ImageList_CoCreateInstance(rclsid *syscall.GUID, punkOuter *IUnknown, riid *syscall.GUID, ppv unsafe.Pointer) HRESULT {
	addr := lazyAddr(&pImageList_CoCreateInstance, libComctl32, "ImageList_CoCreateInstance")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(rclsid)), uintptr(unsafe.Pointer(punkOuter)), uintptr(unsafe.Pointer(riid)), uintptr(ppv))
	return HRESULT(ret)
}

func BeginPanningFeedback(hwnd HWND) BOOL {
	addr := lazyAddr(&pBeginPanningFeedback, libUxtheme, "BeginPanningFeedback")
	ret, _, _ := syscall.SyscallN(addr, hwnd)
	return BOOL(ret)
}

func UpdatePanningFeedback(hwnd HWND, lTotalOverpanOffsetX int32, lTotalOverpanOffsetY int32, fInInertia BOOL) BOOL {
	addr := lazyAddr(&pUpdatePanningFeedback, libUxtheme, "UpdatePanningFeedback")
	ret, _, _ := syscall.SyscallN(addr, hwnd, uintptr(lTotalOverpanOffsetX), uintptr(lTotalOverpanOffsetY), uintptr(fInInertia))
	return BOOL(ret)
}

func EndPanningFeedback(hwnd HWND, fAnimateBack BOOL) BOOL {
	addr := lazyAddr(&pEndPanningFeedback, libUxtheme, "EndPanningFeedback")
	ret, _, _ := syscall.SyscallN(addr, hwnd, uintptr(fAnimateBack))
	return BOOL(ret)
}

func GetThemeAnimationProperty(hTheme uintptr, iStoryboardId int32, iTargetId int32, eProperty TA_PROPERTY, pvProperty unsafe.Pointer, cbSize uint32, pcbSizeOut *uint32) HRESULT {
	addr := lazyAddr(&pGetThemeAnimationProperty, libUxtheme, "GetThemeAnimationProperty")
	ret, _, _ := syscall.SyscallN(addr, hTheme, uintptr(iStoryboardId), uintptr(iTargetId), uintptr(eProperty), uintptr(pvProperty), uintptr(cbSize), uintptr(unsafe.Pointer(pcbSizeOut)))
	return HRESULT(ret)
}

func GetThemeAnimationTransform(hTheme uintptr, iStoryboardId int32, iTargetId int32, dwTransformIndex uint32, pTransform *TA_TRANSFORM, cbSize uint32, pcbSizeOut *uint32) HRESULT {
	addr := lazyAddr(&pGetThemeAnimationTransform, libUxtheme, "GetThemeAnimationTransform")
	ret, _, _ := syscall.SyscallN(addr, hTheme, uintptr(iStoryboardId), uintptr(iTargetId), uintptr(dwTransformIndex), uintptr(unsafe.Pointer(pTransform)), uintptr(cbSize), uintptr(unsafe.Pointer(pcbSizeOut)))
	return HRESULT(ret)
}

func GetThemeTimingFunction(hTheme uintptr, iTimingFunctionId int32, pTimingFunction *TA_TIMINGFUNCTION, cbSize uint32, pcbSizeOut *uint32) HRESULT {
	addr := lazyAddr(&pGetThemeTimingFunction, libUxtheme, "GetThemeTimingFunction")
	ret, _, _ := syscall.SyscallN(addr, hTheme, uintptr(iTimingFunctionId), uintptr(unsafe.Pointer(pTimingFunction)), uintptr(cbSize), uintptr(unsafe.Pointer(pcbSizeOut)))
	return HRESULT(ret)
}

func OpenThemeData(hwnd HWND, pszClassList PWSTR) uintptr {
	addr := lazyAddr(&pOpenThemeData, libUxtheme, "OpenThemeData")
	ret, _, _ := syscall.SyscallN(addr, hwnd, uintptr(unsafe.Pointer(pszClassList)))
	return ret
}

func OpenThemeDataEx(hwnd HWND, pszClassList PWSTR, dwFlags OPEN_THEME_DATA_FLAGS) uintptr {
	addr := lazyAddr(&pOpenThemeDataEx, libUxtheme, "OpenThemeDataEx")
	ret, _, _ := syscall.SyscallN(addr, hwnd, uintptr(unsafe.Pointer(pszClassList)), uintptr(dwFlags))
	return ret
}

func CloseThemeData(hTheme uintptr) HRESULT {
	addr := lazyAddr(&pCloseThemeData, libUxtheme, "CloseThemeData")
	ret, _, _ := syscall.SyscallN(addr, hTheme)
	return HRESULT(ret)
}

func DrawThemeBackground(hTheme uintptr, hdc HDC, iPartId int32, iStateId int32, pRect *RECT, pClipRect *RECT) HRESULT {
	addr := lazyAddr(&pDrawThemeBackground, libUxtheme, "DrawThemeBackground")
	ret, _, _ := syscall.SyscallN(addr, hTheme, hdc, uintptr(iPartId), uintptr(iStateId), uintptr(unsafe.Pointer(pRect)), uintptr(unsafe.Pointer(pClipRect)))
	return HRESULT(ret)
}

func DrawThemeBackgroundEx(hTheme uintptr, hdc HDC, iPartId int32, iStateId int32, pRect *RECT, pOptions *DTBGOPTS) HRESULT {
	addr := lazyAddr(&pDrawThemeBackgroundEx, libUxtheme, "DrawThemeBackgroundEx")
	ret, _, _ := syscall.SyscallN(addr, hTheme, hdc, uintptr(iPartId), uintptr(iStateId), uintptr(unsafe.Pointer(pRect)), uintptr(unsafe.Pointer(pOptions)))
	return HRESULT(ret)
}

func DrawThemeText(hTheme uintptr, hdc HDC, iPartId int32, iStateId int32, pszText PWSTR, cchText int32, dwTextFlags uint32, dwTextFlags2 uint32, pRect *RECT) HRESULT {
	addr := lazyAddr(&pDrawThemeText, libUxtheme, "DrawThemeText")
	ret, _, _ := syscall.SyscallN(addr, hTheme, hdc, uintptr(iPartId), uintptr(iStateId), uintptr(unsafe.Pointer(pszText)), uintptr(cchText), uintptr(dwTextFlags), uintptr(dwTextFlags2), uintptr(unsafe.Pointer(pRect)))
	return HRESULT(ret)
}

func GetThemeBackgroundContentRect(hTheme uintptr, hdc HDC, iPartId int32, iStateId int32, pBoundingRect *RECT, pContentRect *RECT) HRESULT {
	addr := lazyAddr(&pGetThemeBackgroundContentRect, libUxtheme, "GetThemeBackgroundContentRect")
	ret, _, _ := syscall.SyscallN(addr, hTheme, hdc, uintptr(iPartId), uintptr(iStateId), uintptr(unsafe.Pointer(pBoundingRect)), uintptr(unsafe.Pointer(pContentRect)))
	return HRESULT(ret)
}

func GetThemeBackgroundExtent(hTheme uintptr, hdc HDC, iPartId int32, iStateId int32, pContentRect *RECT, pExtentRect *RECT) HRESULT {
	addr := lazyAddr(&pGetThemeBackgroundExtent, libUxtheme, "GetThemeBackgroundExtent")
	ret, _, _ := syscall.SyscallN(addr, hTheme, hdc, uintptr(iPartId), uintptr(iStateId), uintptr(unsafe.Pointer(pContentRect)), uintptr(unsafe.Pointer(pExtentRect)))
	return HRESULT(ret)
}

func GetThemeBackgroundRegion(hTheme uintptr, hdc HDC, iPartId int32, iStateId int32, pRect *RECT, pRegion *HRGN) HRESULT {
	addr := lazyAddr(&pGetThemeBackgroundRegion, libUxtheme, "GetThemeBackgroundRegion")
	ret, _, _ := syscall.SyscallN(addr, hTheme, hdc, uintptr(iPartId), uintptr(iStateId), uintptr(unsafe.Pointer(pRect)), uintptr(unsafe.Pointer(pRegion)))
	return HRESULT(ret)
}

func GetThemePartSize(hTheme uintptr, hdc HDC, iPartId int32, iStateId int32, prc *RECT, eSize THEMESIZE, psz *SIZE) HRESULT {
	addr := lazyAddr(&pGetThemePartSize, libUxtheme, "GetThemePartSize")
	ret, _, _ := syscall.SyscallN(addr, hTheme, hdc, uintptr(iPartId), uintptr(iStateId), uintptr(unsafe.Pointer(prc)), uintptr(eSize), uintptr(unsafe.Pointer(psz)))
	return HRESULT(ret)
}

func GetThemeTextExtent(hTheme uintptr, hdc HDC, iPartId int32, iStateId int32, pszText PWSTR, cchCharCount int32, dwTextFlags uint32, pBoundingRect *RECT, pExtentRect *RECT) HRESULT {
	addr := lazyAddr(&pGetThemeTextExtent, libUxtheme, "GetThemeTextExtent")
	ret, _, _ := syscall.SyscallN(addr, hTheme, hdc, uintptr(iPartId), uintptr(iStateId), uintptr(unsafe.Pointer(pszText)), uintptr(cchCharCount), uintptr(dwTextFlags), uintptr(unsafe.Pointer(pBoundingRect)), uintptr(unsafe.Pointer(pExtentRect)))
	return HRESULT(ret)
}

func GetThemeTextMetrics(hTheme uintptr, hdc HDC, iPartId int32, iStateId int32, ptm *TEXTMETRICW) HRESULT {
	addr := lazyAddr(&pGetThemeTextMetrics, libUxtheme, "GetThemeTextMetrics")
	ret, _, _ := syscall.SyscallN(addr, hTheme, hdc, uintptr(iPartId), uintptr(iStateId), uintptr(unsafe.Pointer(ptm)))
	return HRESULT(ret)
}

func HitTestThemeBackground(hTheme uintptr, hdc HDC, iPartId int32, iStateId int32, dwOptions uint32, pRect *RECT, hrgn HRGN, ptTest POINT, pwHitTestCode *uint16) HRESULT {
	addr := lazyAddr(&pHitTestThemeBackground, libUxtheme, "HitTestThemeBackground")
	ret, _, _ := syscall.SyscallN(addr, hTheme, hdc, uintptr(iPartId), uintptr(iStateId), uintptr(dwOptions), uintptr(unsafe.Pointer(pRect)), hrgn, *(*uintptr)(unsafe.Pointer(&ptTest)), uintptr(unsafe.Pointer(pwHitTestCode)))
	return HRESULT(ret)
}

func DrawThemeEdge(hTheme uintptr, hdc HDC, iPartId int32, iStateId int32, pDestRect *RECT, uEdge uint32, uFlags uint32, pContentRect *RECT) HRESULT {
	addr := lazyAddr(&pDrawThemeEdge, libUxtheme, "DrawThemeEdge")
	ret, _, _ := syscall.SyscallN(addr, hTheme, hdc, uintptr(iPartId), uintptr(iStateId), uintptr(unsafe.Pointer(pDestRect)), uintptr(uEdge), uintptr(uFlags), uintptr(unsafe.Pointer(pContentRect)))
	return HRESULT(ret)
}

func DrawThemeIcon(hTheme uintptr, hdc HDC, iPartId int32, iStateId int32, pRect *RECT, himl HIMAGELIST, iImageIndex int32) HRESULT {
	addr := lazyAddr(&pDrawThemeIcon, libUxtheme, "DrawThemeIcon")
	ret, _, _ := syscall.SyscallN(addr, hTheme, hdc, uintptr(iPartId), uintptr(iStateId), uintptr(unsafe.Pointer(pRect)), himl, uintptr(iImageIndex))
	return HRESULT(ret)
}

func IsThemePartDefined(hTheme uintptr, iPartId int32, iStateId int32) BOOL {
	addr := lazyAddr(&pIsThemePartDefined, libUxtheme, "IsThemePartDefined")
	ret, _, _ := syscall.SyscallN(addr, hTheme, uintptr(iPartId), uintptr(iStateId))
	return BOOL(ret)
}

func IsThemeBackgroundPartiallyTransparent(hTheme uintptr, iPartId int32, iStateId int32) BOOL {
	addr := lazyAddr(&pIsThemeBackgroundPartiallyTransparent, libUxtheme, "IsThemeBackgroundPartiallyTransparent")
	ret, _, _ := syscall.SyscallN(addr, hTheme, uintptr(iPartId), uintptr(iStateId))
	return BOOL(ret)
}

func GetThemeColor(hTheme uintptr, iPartId int32, iStateId int32, iPropId int32, pColor *uint32) HRESULT {
	addr := lazyAddr(&pGetThemeColor, libUxtheme, "GetThemeColor")
	ret, _, _ := syscall.SyscallN(addr, hTheme, uintptr(iPartId), uintptr(iStateId), uintptr(iPropId), uintptr(unsafe.Pointer(pColor)))
	return HRESULT(ret)
}

func GetThemeMetric(hTheme uintptr, hdc HDC, iPartId int32, iStateId int32, iPropId THEME_PROPERTY_SYMBOL_ID, piVal *int32) HRESULT {
	addr := lazyAddr(&pGetThemeMetric, libUxtheme, "GetThemeMetric")
	ret, _, _ := syscall.SyscallN(addr, hTheme, hdc, uintptr(iPartId), uintptr(iStateId), uintptr(iPropId), uintptr(unsafe.Pointer(piVal)))
	return HRESULT(ret)
}

func GetThemeString(hTheme uintptr, iPartId int32, iStateId int32, iPropId int32, pszBuff PWSTR, cchMaxBuffChars int32) HRESULT {
	addr := lazyAddr(&pGetThemeString, libUxtheme, "GetThemeString")
	ret, _, _ := syscall.SyscallN(addr, hTheme, uintptr(iPartId), uintptr(iStateId), uintptr(iPropId), uintptr(unsafe.Pointer(pszBuff)), uintptr(cchMaxBuffChars))
	return HRESULT(ret)
}

func GetThemeBool(hTheme uintptr, iPartId int32, iStateId int32, iPropId THEME_PROPERTY_SYMBOL_ID, pfVal *BOOL) HRESULT {
	addr := lazyAddr(&pGetThemeBool, libUxtheme, "GetThemeBool")
	ret, _, _ := syscall.SyscallN(addr, hTheme, uintptr(iPartId), uintptr(iStateId), uintptr(iPropId), uintptr(unsafe.Pointer(pfVal)))
	return HRESULT(ret)
}

func GetThemeInt(hTheme uintptr, iPartId int32, iStateId int32, iPropId int32, piVal *int32) HRESULT {
	addr := lazyAddr(&pGetThemeInt, libUxtheme, "GetThemeInt")
	ret, _, _ := syscall.SyscallN(addr, hTheme, uintptr(iPartId), uintptr(iStateId), uintptr(iPropId), uintptr(unsafe.Pointer(piVal)))
	return HRESULT(ret)
}

func GetThemeEnumValue(hTheme uintptr, iPartId int32, iStateId int32, iPropId int32, piVal *int32) HRESULT {
	addr := lazyAddr(&pGetThemeEnumValue, libUxtheme, "GetThemeEnumValue")
	ret, _, _ := syscall.SyscallN(addr, hTheme, uintptr(iPartId), uintptr(iStateId), uintptr(iPropId), uintptr(unsafe.Pointer(piVal)))
	return HRESULT(ret)
}

func GetThemePosition(hTheme uintptr, iPartId int32, iStateId int32, iPropId int32, pPoint *POINT) HRESULT {
	addr := lazyAddr(&pGetThemePosition, libUxtheme, "GetThemePosition")
	ret, _, _ := syscall.SyscallN(addr, hTheme, uintptr(iPartId), uintptr(iStateId), uintptr(iPropId), uintptr(unsafe.Pointer(pPoint)))
	return HRESULT(ret)
}

func GetThemeFont(hTheme uintptr, hdc HDC, iPartId int32, iStateId int32, iPropId int32, pFont *LOGFONTW) HRESULT {
	addr := lazyAddr(&pGetThemeFont, libUxtheme, "GetThemeFont")
	ret, _, _ := syscall.SyscallN(addr, hTheme, hdc, uintptr(iPartId), uintptr(iStateId), uintptr(iPropId), uintptr(unsafe.Pointer(pFont)))
	return HRESULT(ret)
}

func GetThemeRect(hTheme uintptr, iPartId int32, iStateId int32, iPropId int32, pRect *RECT) HRESULT {
	addr := lazyAddr(&pGetThemeRect, libUxtheme, "GetThemeRect")
	ret, _, _ := syscall.SyscallN(addr, hTheme, uintptr(iPartId), uintptr(iStateId), uintptr(iPropId), uintptr(unsafe.Pointer(pRect)))
	return HRESULT(ret)
}

func GetThemeMargins(hTheme uintptr, hdc HDC, iPartId int32, iStateId int32, iPropId int32, prc *RECT, pMargins *MARGINS) HRESULT {
	addr := lazyAddr(&pGetThemeMargins, libUxtheme, "GetThemeMargins")
	ret, _, _ := syscall.SyscallN(addr, hTheme, hdc, uintptr(iPartId), uintptr(iStateId), uintptr(iPropId), uintptr(unsafe.Pointer(prc)), uintptr(unsafe.Pointer(pMargins)))
	return HRESULT(ret)
}

func GetThemeIntList(hTheme uintptr, iPartId int32, iStateId int32, iPropId int32, pIntList *INTLIST) HRESULT {
	addr := lazyAddr(&pGetThemeIntList, libUxtheme, "GetThemeIntList")
	ret, _, _ := syscall.SyscallN(addr, hTheme, uintptr(iPartId), uintptr(iStateId), uintptr(iPropId), uintptr(unsafe.Pointer(pIntList)))
	return HRESULT(ret)
}

func GetThemePropertyOrigin(hTheme uintptr, iPartId int32, iStateId int32, iPropId int32, pOrigin *PROPERTYORIGIN) HRESULT {
	addr := lazyAddr(&pGetThemePropertyOrigin, libUxtheme, "GetThemePropertyOrigin")
	ret, _, _ := syscall.SyscallN(addr, hTheme, uintptr(iPartId), uintptr(iStateId), uintptr(iPropId), uintptr(unsafe.Pointer(pOrigin)))
	return HRESULT(ret)
}

func SetWindowTheme(hwnd HWND, pszSubAppName PWSTR, pszSubIdList PWSTR) HRESULT {
	addr := lazyAddr(&pSetWindowTheme, libUxtheme, "SetWindowTheme")
	ret, _, _ := syscall.SyscallN(addr, hwnd, uintptr(unsafe.Pointer(pszSubAppName)), uintptr(unsafe.Pointer(pszSubIdList)))
	return HRESULT(ret)
}

func GetThemeFilename(hTheme uintptr, iPartId int32, iStateId int32, iPropId int32, pszThemeFileName PWSTR, cchMaxBuffChars int32) HRESULT {
	addr := lazyAddr(&pGetThemeFilename, libUxtheme, "GetThemeFilename")
	ret, _, _ := syscall.SyscallN(addr, hTheme, uintptr(iPartId), uintptr(iStateId), uintptr(iPropId), uintptr(unsafe.Pointer(pszThemeFileName)), uintptr(cchMaxBuffChars))
	return HRESULT(ret)
}

func GetThemeSysColor(hTheme uintptr, iColorId int32) uint32 {
	addr := lazyAddr(&pGetThemeSysColor, libUxtheme, "GetThemeSysColor")
	ret, _, _ := syscall.SyscallN(addr, hTheme, uintptr(iColorId))
	return uint32(ret)
}

func GetThemeSysColorBrush(hTheme uintptr, iColorId THEME_PROPERTY_SYMBOL_ID) HBRUSH {
	addr := lazyAddr(&pGetThemeSysColorBrush, libUxtheme, "GetThemeSysColorBrush")
	ret, _, _ := syscall.SyscallN(addr, hTheme, uintptr(iColorId))
	return ret
}

func GetThemeSysBool(hTheme uintptr, iBoolId int32) BOOL {
	addr := lazyAddr(&pGetThemeSysBool, libUxtheme, "GetThemeSysBool")
	ret, _, _ := syscall.SyscallN(addr, hTheme, uintptr(iBoolId))
	return BOOL(ret)
}

func GetThemeSysSize(hTheme uintptr, iSizeId int32) int32 {
	addr := lazyAddr(&pGetThemeSysSize, libUxtheme, "GetThemeSysSize")
	ret, _, _ := syscall.SyscallN(addr, hTheme, uintptr(iSizeId))
	return int32(ret)
}

func GetThemeSysFont(hTheme uintptr, iFontId THEME_PROPERTY_SYMBOL_ID, plf *LOGFONTW) HRESULT {
	addr := lazyAddr(&pGetThemeSysFont, libUxtheme, "GetThemeSysFont")
	ret, _, _ := syscall.SyscallN(addr, hTheme, uintptr(iFontId), uintptr(unsafe.Pointer(plf)))
	return HRESULT(ret)
}

func GetThemeSysString(hTheme uintptr, iStringId THEME_PROPERTY_SYMBOL_ID, pszStringBuff PWSTR, cchMaxStringChars int32) HRESULT {
	addr := lazyAddr(&pGetThemeSysString, libUxtheme, "GetThemeSysString")
	ret, _, _ := syscall.SyscallN(addr, hTheme, uintptr(iStringId), uintptr(unsafe.Pointer(pszStringBuff)), uintptr(cchMaxStringChars))
	return HRESULT(ret)
}

func GetThemeSysInt(hTheme uintptr, iIntId int32, piValue *int32) HRESULT {
	addr := lazyAddr(&pGetThemeSysInt, libUxtheme, "GetThemeSysInt")
	ret, _, _ := syscall.SyscallN(addr, hTheme, uintptr(iIntId), uintptr(unsafe.Pointer(piValue)))
	return HRESULT(ret)
}

func IsThemeActive() BOOL {
	addr := lazyAddr(&pIsThemeActive, libUxtheme, "IsThemeActive")
	ret, _, _ := syscall.SyscallN(addr)
	return BOOL(ret)
}

func IsAppThemed() BOOL {
	addr := lazyAddr(&pIsAppThemed, libUxtheme, "IsAppThemed")
	ret, _, _ := syscall.SyscallN(addr)
	return BOOL(ret)
}

func GetWindowTheme(hwnd HWND) uintptr {
	addr := lazyAddr(&pGetWindowTheme, libUxtheme, "GetWindowTheme")
	ret, _, _ := syscall.SyscallN(addr, hwnd)
	return ret
}

func EnableThemeDialogTexture(hwnd HWND, dwFlags uint32) HRESULT {
	addr := lazyAddr(&pEnableThemeDialogTexture, libUxtheme, "EnableThemeDialogTexture")
	ret, _, _ := syscall.SyscallN(addr, hwnd, uintptr(dwFlags))
	return HRESULT(ret)
}

func IsThemeDialogTextureEnabled(hwnd HWND) BOOL {
	addr := lazyAddr(&pIsThemeDialogTextureEnabled, libUxtheme, "IsThemeDialogTextureEnabled")
	ret, _, _ := syscall.SyscallN(addr, hwnd)
	return BOOL(ret)
}

func GetThemeAppProperties() uint32 {
	addr := lazyAddr(&pGetThemeAppProperties, libUxtheme, "GetThemeAppProperties")
	ret, _, _ := syscall.SyscallN(addr)
	return uint32(ret)
}

func SetThemeAppProperties(dwFlags uint32) {
	addr := lazyAddr(&pSetThemeAppProperties, libUxtheme, "SetThemeAppProperties")
	syscall.SyscallN(addr, uintptr(dwFlags))
}

func GetCurrentThemeName(pszThemeFileName PWSTR, cchMaxNameChars int32, pszColorBuff PWSTR, cchMaxColorChars int32, pszSizeBuff PWSTR, cchMaxSizeChars int32) HRESULT {
	addr := lazyAddr(&pGetCurrentThemeName, libUxtheme, "GetCurrentThemeName")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pszThemeFileName)), uintptr(cchMaxNameChars), uintptr(unsafe.Pointer(pszColorBuff)), uintptr(cchMaxColorChars), uintptr(unsafe.Pointer(pszSizeBuff)), uintptr(cchMaxSizeChars))
	return HRESULT(ret)
}

func GetThemeDocumentationProperty(pszThemeName PWSTR, pszPropertyName PWSTR, pszValueBuff PWSTR, cchMaxValChars int32) HRESULT {
	addr := lazyAddr(&pGetThemeDocumentationProperty, libUxtheme, "GetThemeDocumentationProperty")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pszThemeName)), uintptr(unsafe.Pointer(pszPropertyName)), uintptr(unsafe.Pointer(pszValueBuff)), uintptr(cchMaxValChars))
	return HRESULT(ret)
}

func DrawThemeParentBackground(hwnd HWND, hdc HDC, prc *RECT) HRESULT {
	addr := lazyAddr(&pDrawThemeParentBackground, libUxtheme, "DrawThemeParentBackground")
	ret, _, _ := syscall.SyscallN(addr, hwnd, hdc, uintptr(unsafe.Pointer(prc)))
	return HRESULT(ret)
}

func EnableTheming(fEnable BOOL) HRESULT {
	addr := lazyAddr(&pEnableTheming, libUxtheme, "EnableTheming")
	ret, _, _ := syscall.SyscallN(addr, uintptr(fEnable))
	return HRESULT(ret)
}

func DrawThemeParentBackgroundEx(hwnd HWND, hdc HDC, dwFlags DRAW_THEME_PARENT_BACKGROUND_FLAGS, prc *RECT) HRESULT {
	addr := lazyAddr(&pDrawThemeParentBackgroundEx, libUxtheme, "DrawThemeParentBackgroundEx")
	ret, _, _ := syscall.SyscallN(addr, hwnd, hdc, uintptr(dwFlags), uintptr(unsafe.Pointer(prc)))
	return HRESULT(ret)
}

func SetWindowThemeAttribute(hwnd HWND, eAttribute WINDOWTHEMEATTRIBUTETYPE, pvAttribute unsafe.Pointer, cbAttribute uint32) HRESULT {
	addr := lazyAddr(&pSetWindowThemeAttribute, libUxtheme, "SetWindowThemeAttribute")
	ret, _, _ := syscall.SyscallN(addr, hwnd, uintptr(eAttribute), uintptr(pvAttribute), uintptr(cbAttribute))
	return HRESULT(ret)
}

func DrawThemeTextEx(hTheme uintptr, hdc HDC, iPartId int32, iStateId int32, pszText PWSTR, cchText int32, dwTextFlags uint32, pRect *RECT, pOptions *DTTOPTS) HRESULT {
	addr := lazyAddr(&pDrawThemeTextEx, libUxtheme, "DrawThemeTextEx")
	ret, _, _ := syscall.SyscallN(addr, hTheme, hdc, uintptr(iPartId), uintptr(iStateId), uintptr(unsafe.Pointer(pszText)), uintptr(cchText), uintptr(dwTextFlags), uintptr(unsafe.Pointer(pRect)), uintptr(unsafe.Pointer(pOptions)))
	return HRESULT(ret)
}

func GetThemeBitmap(hTheme uintptr, iPartId int32, iStateId int32, iPropId THEME_PROPERTY_SYMBOL_ID, dwFlags GET_THEME_BITMAP_FLAGS, phBitmap *HBITMAP) HRESULT {
	addr := lazyAddr(&pGetThemeBitmap, libUxtheme, "GetThemeBitmap")
	ret, _, _ := syscall.SyscallN(addr, hTheme, uintptr(iPartId), uintptr(iStateId), uintptr(iPropId), uintptr(dwFlags), uintptr(unsafe.Pointer(phBitmap)))
	return HRESULT(ret)
}

func GetThemeStream(hTheme uintptr, iPartId int32, iStateId int32, iPropId int32, ppvStream unsafe.Pointer, pcbStream *uint32, hInst HINSTANCE) HRESULT {
	addr := lazyAddr(&pGetThemeStream, libUxtheme, "GetThemeStream")
	ret, _, _ := syscall.SyscallN(addr, hTheme, uintptr(iPartId), uintptr(iStateId), uintptr(iPropId), uintptr(ppvStream), uintptr(unsafe.Pointer(pcbStream)), hInst)
	return HRESULT(ret)
}

func BufferedPaintInit() HRESULT {
	addr := lazyAddr(&pBufferedPaintInit, libUxtheme, "BufferedPaintInit")
	ret, _, _ := syscall.SyscallN(addr)
	return HRESULT(ret)
}

func BufferedPaintUnInit() HRESULT {
	addr := lazyAddr(&pBufferedPaintUnInit, libUxtheme, "BufferedPaintUnInit")
	ret, _, _ := syscall.SyscallN(addr)
	return HRESULT(ret)
}

func BeginBufferedPaint(hdcTarget HDC, prcTarget *RECT, dwFormat BP_BUFFERFORMAT, pPaintParams *BP_PAINTPARAMS, phdc *HDC) (uintptr, WIN32_ERROR) {
	addr := lazyAddr(&pBeginBufferedPaint, libUxtheme, "BeginBufferedPaint")
	ret, _, err := syscall.SyscallN(addr, hdcTarget, uintptr(unsafe.Pointer(prcTarget)), uintptr(dwFormat), uintptr(unsafe.Pointer(pPaintParams)), uintptr(unsafe.Pointer(phdc)))
	return ret, WIN32_ERROR(err)
}

func EndBufferedPaint(hBufferedPaint uintptr, fUpdateTarget BOOL) HRESULT {
	addr := lazyAddr(&pEndBufferedPaint, libUxtheme, "EndBufferedPaint")
	ret, _, _ := syscall.SyscallN(addr, hBufferedPaint, uintptr(fUpdateTarget))
	return HRESULT(ret)
}

func GetBufferedPaintTargetRect(hBufferedPaint uintptr, prc *RECT) HRESULT {
	addr := lazyAddr(&pGetBufferedPaintTargetRect, libUxtheme, "GetBufferedPaintTargetRect")
	ret, _, _ := syscall.SyscallN(addr, hBufferedPaint, uintptr(unsafe.Pointer(prc)))
	return HRESULT(ret)
}

func GetBufferedPaintTargetDC(hBufferedPaint uintptr) HDC {
	addr := lazyAddr(&pGetBufferedPaintTargetDC, libUxtheme, "GetBufferedPaintTargetDC")
	ret, _, _ := syscall.SyscallN(addr, hBufferedPaint)
	return ret
}

func GetBufferedPaintDC(hBufferedPaint uintptr) HDC {
	addr := lazyAddr(&pGetBufferedPaintDC, libUxtheme, "GetBufferedPaintDC")
	ret, _, _ := syscall.SyscallN(addr, hBufferedPaint)
	return ret
}

func GetBufferedPaintBits(hBufferedPaint uintptr, ppbBuffer **RGBQUAD, pcxRow *int32) HRESULT {
	addr := lazyAddr(&pGetBufferedPaintBits, libUxtheme, "GetBufferedPaintBits")
	ret, _, _ := syscall.SyscallN(addr, hBufferedPaint, uintptr(unsafe.Pointer(ppbBuffer)), uintptr(unsafe.Pointer(pcxRow)))
	return HRESULT(ret)
}

func BufferedPaintClear(hBufferedPaint uintptr, prc *RECT) HRESULT {
	addr := lazyAddr(&pBufferedPaintClear, libUxtheme, "BufferedPaintClear")
	ret, _, _ := syscall.SyscallN(addr, hBufferedPaint, uintptr(unsafe.Pointer(prc)))
	return HRESULT(ret)
}

func BufferedPaintSetAlpha(hBufferedPaint uintptr, prc *RECT, alpha byte) HRESULT {
	addr := lazyAddr(&pBufferedPaintSetAlpha, libUxtheme, "BufferedPaintSetAlpha")
	ret, _, _ := syscall.SyscallN(addr, hBufferedPaint, uintptr(unsafe.Pointer(prc)), uintptr(alpha))
	return HRESULT(ret)
}

func BufferedPaintStopAllAnimations(hwnd HWND) HRESULT {
	addr := lazyAddr(&pBufferedPaintStopAllAnimations, libUxtheme, "BufferedPaintStopAllAnimations")
	ret, _, _ := syscall.SyscallN(addr, hwnd)
	return HRESULT(ret)
}

func BeginBufferedAnimation(hwnd HWND, hdcTarget HDC, prcTarget *RECT, dwFormat BP_BUFFERFORMAT, pPaintParams *BP_PAINTPARAMS, pAnimationParams *BP_ANIMATIONPARAMS, phdcFrom *HDC, phdcTo *HDC) uintptr {
	addr := lazyAddr(&pBeginBufferedAnimation, libUxtheme, "BeginBufferedAnimation")
	ret, _, _ := syscall.SyscallN(addr, hwnd, hdcTarget, uintptr(unsafe.Pointer(prcTarget)), uintptr(dwFormat), uintptr(unsafe.Pointer(pPaintParams)), uintptr(unsafe.Pointer(pAnimationParams)), uintptr(unsafe.Pointer(phdcFrom)), uintptr(unsafe.Pointer(phdcTo)))
	return ret
}

func EndBufferedAnimation(hbpAnimation uintptr, fUpdateTarget BOOL) HRESULT {
	addr := lazyAddr(&pEndBufferedAnimation, libUxtheme, "EndBufferedAnimation")
	ret, _, _ := syscall.SyscallN(addr, hbpAnimation, uintptr(fUpdateTarget))
	return HRESULT(ret)
}

func BufferedPaintRenderAnimation(hwnd HWND, hdcTarget HDC) BOOL {
	addr := lazyAddr(&pBufferedPaintRenderAnimation, libUxtheme, "BufferedPaintRenderAnimation")
	ret, _, _ := syscall.SyscallN(addr, hwnd, hdcTarget)
	return BOOL(ret)
}

func IsCompositionActive() BOOL {
	addr := lazyAddr(&pIsCompositionActive, libUxtheme, "IsCompositionActive")
	ret, _, _ := syscall.SyscallN(addr)
	return BOOL(ret)
}

func GetThemeTransitionDuration(hTheme uintptr, iPartId int32, iStateIdFrom int32, iStateIdTo int32, iPropId int32, pdwDuration *uint32) HRESULT {
	addr := lazyAddr(&pGetThemeTransitionDuration, libUxtheme, "GetThemeTransitionDuration")
	ret, _, _ := syscall.SyscallN(addr, hTheme, uintptr(iPartId), uintptr(iStateIdFrom), uintptr(iStateIdTo), uintptr(iPropId), uintptr(unsafe.Pointer(pdwDuration)))
	return HRESULT(ret)
}

func CheckDlgButton(hDlg HWND, nIDButton int32, uCheck DLG_BUTTON_CHECK_STATE) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pCheckDlgButton, libUser32, "CheckDlgButton")
	ret, _, err := syscall.SyscallN(addr, hDlg, uintptr(nIDButton), uintptr(uCheck))
	return BOOL(ret), WIN32_ERROR(err)
}

func CheckRadioButton(hDlg HWND, nIDFirstButton int32, nIDLastButton int32, nIDCheckButton int32) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pCheckRadioButton, libUser32, "CheckRadioButton")
	ret, _, err := syscall.SyscallN(addr, hDlg, uintptr(nIDFirstButton), uintptr(nIDLastButton), uintptr(nIDCheckButton))
	return BOOL(ret), WIN32_ERROR(err)
}

func IsDlgButtonChecked(hDlg HWND, nIDButton int32) uint32 {
	addr := lazyAddr(&pIsDlgButtonChecked, libUser32, "IsDlgButtonChecked")
	ret, _, _ := syscall.SyscallN(addr, hDlg, uintptr(nIDButton))
	return uint32(ret)
}

func IsCharLowerW(ch uint16) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pIsCharLowerW, libUser32, "IsCharLowerW")
	ret, _, err := syscall.SyscallN(addr, uintptr(ch))
	return BOOL(ret), WIN32_ERROR(err)
}

func CreateSyntheticPointerDevice(pointerType POINTER_INPUT_TYPE, maxCount uint32, mode POINTER_FEEDBACK_MODE) (HSYNTHETICPOINTERDEVICE, WIN32_ERROR) {
	addr := lazyAddr(&pCreateSyntheticPointerDevice, libUser32, "CreateSyntheticPointerDevice")
	ret, _, err := syscall.SyscallN(addr, uintptr(pointerType), uintptr(maxCount), uintptr(mode))
	return ret, WIN32_ERROR(err)
}

func DestroySyntheticPointerDevice(device HSYNTHETICPOINTERDEVICE) {
	addr := lazyAddr(&pDestroySyntheticPointerDevice, libUser32, "DestroySyntheticPointerDevice")
	syscall.SyscallN(addr, device)
}

func RegisterTouchHitTestingWindow(hwnd HWND, value uint32) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pRegisterTouchHitTestingWindow, libUser32, "RegisterTouchHitTestingWindow")
	ret, _, err := syscall.SyscallN(addr, hwnd, uintptr(value))
	return BOOL(ret), WIN32_ERROR(err)
}

func EvaluateProximityToRect(controlBoundingBox *RECT, pHitTestingInput *TOUCH_HIT_TESTING_INPUT, pProximityEval *TOUCH_HIT_TESTING_PROXIMITY_EVALUATION) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pEvaluateProximityToRect, libUser32, "EvaluateProximityToRect")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(controlBoundingBox)), uintptr(unsafe.Pointer(pHitTestingInput)), uintptr(unsafe.Pointer(pProximityEval)))
	return BOOL(ret), WIN32_ERROR(err)
}

func EvaluateProximityToPolygon(numVertices uint32, controlPolygon *POINT, pHitTestingInput *TOUCH_HIT_TESTING_INPUT, pProximityEval *TOUCH_HIT_TESTING_PROXIMITY_EVALUATION) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pEvaluateProximityToPolygon, libUser32, "EvaluateProximityToPolygon")
	ret, _, err := syscall.SyscallN(addr, uintptr(numVertices), uintptr(unsafe.Pointer(controlPolygon)), uintptr(unsafe.Pointer(pHitTestingInput)), uintptr(unsafe.Pointer(pProximityEval)))
	return BOOL(ret), WIN32_ERROR(err)
}

func PackTouchHitTestingProximityEvaluation(pHitTestingInput *TOUCH_HIT_TESTING_INPUT, pProximityEval *TOUCH_HIT_TESTING_PROXIMITY_EVALUATION) (LRESULT, WIN32_ERROR) {
	addr := lazyAddr(&pPackTouchHitTestingProximityEvaluation, libUser32, "PackTouchHitTestingProximityEvaluation")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pHitTestingInput)), uintptr(unsafe.Pointer(pProximityEval)))
	return ret, WIN32_ERROR(err)
}

func GetWindowFeedbackSetting(hwnd HWND, feedback FEEDBACK_TYPE, dwFlags uint32, pSize *uint32, config unsafe.Pointer) BOOL {
	addr := lazyAddr(&pGetWindowFeedbackSetting, libUser32, "GetWindowFeedbackSetting")
	ret, _, _ := syscall.SyscallN(addr, hwnd, uintptr(feedback), uintptr(dwFlags), uintptr(unsafe.Pointer(pSize)), uintptr(config))
	return BOOL(ret)
}

func SetWindowFeedbackSetting(hwnd HWND, feedback FEEDBACK_TYPE, dwFlags uint32, size uint32, configuration unsafe.Pointer) BOOL {
	addr := lazyAddr(&pSetWindowFeedbackSetting, libUser32, "SetWindowFeedbackSetting")
	ret, _, _ := syscall.SyscallN(addr, hwnd, uintptr(feedback), uintptr(dwFlags), uintptr(size), uintptr(configuration))
	return BOOL(ret)
}

func SetScrollPos(hWnd HWND, nBar SCROLLBAR_CONSTANTS, nPos int32, bRedraw BOOL) (int32, WIN32_ERROR) {
	addr := lazyAddr(&pSetScrollPos, libUser32, "SetScrollPos")
	ret, _, err := syscall.SyscallN(addr, hWnd, uintptr(nBar), uintptr(nPos), uintptr(bRedraw))
	return int32(ret), WIN32_ERROR(err)
}

func SetScrollRange(hWnd HWND, nBar SCROLLBAR_CONSTANTS, nMinPos int32, nMaxPos int32, bRedraw BOOL) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pSetScrollRange, libUser32, "SetScrollRange")
	ret, _, err := syscall.SyscallN(addr, hWnd, uintptr(nBar), uintptr(nMinPos), uintptr(nMaxPos), uintptr(bRedraw))
	return BOOL(ret), WIN32_ERROR(err)
}

func ShowScrollBar(hWnd HWND, wBar SCROLLBAR_CONSTANTS, bShow BOOL) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pShowScrollBar, libUser32, "ShowScrollBar")
	ret, _, err := syscall.SyscallN(addr, hWnd, uintptr(wBar), uintptr(bShow))
	return BOOL(ret), WIN32_ERROR(err)
}

func EnableScrollBar(hWnd HWND, wSBflags SCROLLBAR_CONSTANTS, wArrows ENABLE_SCROLL_BAR_ARROWS) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pEnableScrollBar, libUser32, "EnableScrollBar")
	ret, _, err := syscall.SyscallN(addr, hWnd, uintptr(wSBflags), uintptr(wArrows))
	return BOOL(ret), WIN32_ERROR(err)
}

func DlgDirListA(hDlg HWND, lpPathSpec PSTR, nIDListBox int32, nIDStaticPath int32, uFileType DLG_DIR_LIST_FILE_TYPE) int32 {
	addr := lazyAddr(&pDlgDirListA, libUser32, "DlgDirListA")
	ret, _, _ := syscall.SyscallN(addr, hDlg, uintptr(unsafe.Pointer(lpPathSpec)), uintptr(nIDListBox), uintptr(nIDStaticPath), uintptr(uFileType))
	return int32(ret)
}

var DlgDirList = DlgDirListW

func DlgDirListW(hDlg HWND, lpPathSpec PWSTR, nIDListBox int32, nIDStaticPath int32, uFileType DLG_DIR_LIST_FILE_TYPE) int32 {
	addr := lazyAddr(&pDlgDirListW, libUser32, "DlgDirListW")
	ret, _, _ := syscall.SyscallN(addr, hDlg, uintptr(unsafe.Pointer(lpPathSpec)), uintptr(nIDListBox), uintptr(nIDStaticPath), uintptr(uFileType))
	return int32(ret)
}

func DlgDirSelectExA(hwndDlg HWND, lpString PSTR, chCount int32, idListBox int32) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pDlgDirSelectExA, libUser32, "DlgDirSelectExA")
	ret, _, err := syscall.SyscallN(addr, hwndDlg, uintptr(unsafe.Pointer(lpString)), uintptr(chCount), uintptr(idListBox))
	return BOOL(ret), WIN32_ERROR(err)
}

var DlgDirSelectEx = DlgDirSelectExW

func DlgDirSelectExW(hwndDlg HWND, lpString PWSTR, chCount int32, idListBox int32) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pDlgDirSelectExW, libUser32, "DlgDirSelectExW")
	ret, _, err := syscall.SyscallN(addr, hwndDlg, uintptr(unsafe.Pointer(lpString)), uintptr(chCount), uintptr(idListBox))
	return BOOL(ret), WIN32_ERROR(err)
}

func DlgDirListComboBoxA(hDlg HWND, lpPathSpec PSTR, nIDComboBox int32, nIDStaticPath int32, uFiletype DLG_DIR_LIST_FILE_TYPE) (int32, WIN32_ERROR) {
	addr := lazyAddr(&pDlgDirListComboBoxA, libUser32, "DlgDirListComboBoxA")
	ret, _, err := syscall.SyscallN(addr, hDlg, uintptr(unsafe.Pointer(lpPathSpec)), uintptr(nIDComboBox), uintptr(nIDStaticPath), uintptr(uFiletype))
	return int32(ret), WIN32_ERROR(err)
}

var DlgDirListComboBox = DlgDirListComboBoxW

func DlgDirListComboBoxW(hDlg HWND, lpPathSpec PWSTR, nIDComboBox int32, nIDStaticPath int32, uFiletype DLG_DIR_LIST_FILE_TYPE) (int32, WIN32_ERROR) {
	addr := lazyAddr(&pDlgDirListComboBoxW, libUser32, "DlgDirListComboBoxW")
	ret, _, err := syscall.SyscallN(addr, hDlg, uintptr(unsafe.Pointer(lpPathSpec)), uintptr(nIDComboBox), uintptr(nIDStaticPath), uintptr(uFiletype))
	return int32(ret), WIN32_ERROR(err)
}

func DlgDirSelectComboBoxExA(hwndDlg HWND, lpString PSTR, cchOut int32, idComboBox int32) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pDlgDirSelectComboBoxExA, libUser32, "DlgDirSelectComboBoxExA")
	ret, _, err := syscall.SyscallN(addr, hwndDlg, uintptr(unsafe.Pointer(lpString)), uintptr(cchOut), uintptr(idComboBox))
	return BOOL(ret), WIN32_ERROR(err)
}

var DlgDirSelectComboBoxEx = DlgDirSelectComboBoxExW

func DlgDirSelectComboBoxExW(hwndDlg HWND, lpString PWSTR, cchOut int32, idComboBox int32) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pDlgDirSelectComboBoxExW, libUser32, "DlgDirSelectComboBoxExW")
	ret, _, err := syscall.SyscallN(addr, hwndDlg, uintptr(unsafe.Pointer(lpString)), uintptr(cchOut), uintptr(idComboBox))
	return BOOL(ret), WIN32_ERROR(err)
}

func SetScrollInfo(hwnd HWND, nBar SCROLLBAR_CONSTANTS, lpsi *SCROLLINFO, redraw BOOL) int32 {
	addr := lazyAddr(&pSetScrollInfo, libUser32, "SetScrollInfo")
	ret, _, _ := syscall.SyscallN(addr, hwnd, uintptr(nBar), uintptr(unsafe.Pointer(lpsi)), uintptr(redraw))
	return int32(ret)
}

func GetComboBoxInfo(hwndCombo HWND, pcbi *COMBOBOXINFO) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pGetComboBoxInfo, libUser32, "GetComboBoxInfo")
	ret, _, err := syscall.SyscallN(addr, hwndCombo, uintptr(unsafe.Pointer(pcbi)))
	return BOOL(ret), WIN32_ERROR(err)
}

func GetListBoxInfo(hwnd HWND) uint32 {
	addr := lazyAddr(&pGetListBoxInfo, libUser32, "GetListBoxInfo")
	ret, _, _ := syscall.SyscallN(addr, hwnd)
	return uint32(ret)
}

func RegisterPointerDeviceNotifications(window HWND, notifyRange BOOL) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pRegisterPointerDeviceNotifications, libUser32, "RegisterPointerDeviceNotifications")
	ret, _, err := syscall.SyscallN(addr, window, uintptr(notifyRange))
	return BOOL(ret), WIN32_ERROR(err)
}
