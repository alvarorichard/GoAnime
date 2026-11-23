package win32

import (
	"syscall"
	"unsafe"
)

const (
	CTL_E_ILLEGALFUNCTIONCALL                  int32   = -2146828283
	CONNECT_E_FIRST                            int32   = -2147220992
	SELFREG_E_FIRST                            int32   = -2147220992
	PERPROP_E_FIRST                            int32   = -2147220992
	OLECMDERR_E_FIRST                          HRESULT = -2147221248
	OLECMDERR_E_DISABLED                       HRESULT = -2147221247
	OLECMDERR_E_NOHELP                         HRESULT = -2147221246
	OLECMDERR_E_CANCELED                       HRESULT = -2147221245
	OLECMDERR_E_UNKNOWNGROUP                   HRESULT = -2147221244
	CONNECT_E_NOCONNECTION                     HRESULT = -2147220992
	CONNECT_E_ADVISELIMIT                      HRESULT = -2147220991
	CONNECT_E_CANNOTCONNECT                    HRESULT = -2147220990
	CONNECT_E_OVERRIDDEN                       HRESULT = -2147220989
	SELFREG_E_TYPELIB                          HRESULT = -2147220992
	SELFREG_E_CLASS                            HRESULT = -2147220991
	PERPROP_E_NOPAGEAVAILABLE                  HRESULT = -2147220992
	PICTYPE_UNINITIALIZED                      int32   = -1
	PICTYPE_NONE                               uint32  = 0x0
	PICTYPE_BITMAP                             uint32  = 0x1
	PICTYPE_METAFILE                           uint32  = 0x2
	PICTYPE_ICON                               uint32  = 0x3
	PICTYPE_ENHMETAFILE                        uint32  = 0x4
	CONNECT_E_LAST                             HRESULT = -2147220977
	CONNECT_S_FIRST                            HRESULT = 262656
	CONNECT_S_LAST                             HRESULT = 262671
	SELFREG_E_LAST                             HRESULT = -2147220977
	SELFREG_S_FIRST                            HRESULT = 262656
	SELFREG_S_LAST                             HRESULT = 262671
	PERPROP_E_LAST                             HRESULT = -2147220977
	PERPROP_S_FIRST                            HRESULT = 262656
	PERPROP_S_LAST                             HRESULT = 262671
	OLEIVERB_PROPERTIES                        int32   = -7
	VT_STREAMED_PROPSET                        uint32  = 0x49
	VT_STORED_PROPSET                          uint32  = 0x4a
	VT_BLOB_PROPSET                            uint32  = 0x4b
	VT_VERBOSE_ENUM                            uint32  = 0x4c
	OCM__BASE                                  uint32  = 0x2000
	LP_DEFAULT                                 uint32  = 0x0
	LP_MONOCHROME                              uint32  = 0x1
	LP_VGACOLOR                                uint32  = 0x2
	LP_COLOR                                   uint32  = 0x4
	DISPID_AUTOSIZE                            int32   = -500
	DISPID_BACKCOLOR                           int32   = -501
	DISPID_BACKSTYLE                           int32   = -502
	DISPID_BORDERCOLOR                         int32   = -503
	DISPID_BORDERSTYLE                         int32   = -504
	DISPID_BORDERWIDTH                         int32   = -505
	DISPID_DRAWMODE                            int32   = -507
	DISPID_DRAWSTYLE                           int32   = -508
	DISPID_DRAWWIDTH                           int32   = -509
	DISPID_FILLCOLOR                           int32   = -510
	DISPID_FILLSTYLE                           int32   = -511
	DISPID_FONT                                int32   = -512
	DISPID_FORECOLOR                           int32   = -513
	DISPID_ENABLED                             int32   = -514
	DISPID_HWND                                int32   = -515
	DISPID_TABSTOP                             int32   = -516
	DISPID_TEXT                                int32   = -517
	DISPID_CAPTION                             int32   = -518
	DISPID_BORDERVISIBLE                       int32   = -519
	DISPID_APPEARANCE                          int32   = -520
	DISPID_MOUSEPOINTER                        int32   = -521
	DISPID_MOUSEICON                           int32   = -522
	DISPID_PICTURE                             int32   = -523
	DISPID_VALID                               int32   = -524
	DISPID_READYSTATE                          int32   = -525
	DISPID_LISTINDEX                           int32   = -526
	DISPID_SELECTED                            int32   = -527
	DISPID_LIST                                int32   = -528
	DISPID_COLUMN                              int32   = -529
	DISPID_LISTCOUNT                           int32   = -531
	DISPID_MULTISELECT                         int32   = -532
	DISPID_MAXLENGTH                           int32   = -533
	DISPID_PASSWORDCHAR                        int32   = -534
	DISPID_SCROLLBARS                          int32   = -535
	DISPID_WORDWRAP                            int32   = -536
	DISPID_MULTILINE                           int32   = -537
	DISPID_NUMBEROFROWS                        int32   = -538
	DISPID_NUMBEROFCOLUMNS                     int32   = -539
	DISPID_DISPLAYSTYLE                        int32   = -540
	DISPID_GROUPNAME                           int32   = -541
	DISPID_IMEMODE                             int32   = -542
	DISPID_ACCELERATOR                         int32   = -543
	DISPID_ENTERKEYBEHAVIOR                    int32   = -544
	DISPID_TABKEYBEHAVIOR                      int32   = -545
	DISPID_SELTEXT                             int32   = -546
	DISPID_SELSTART                            int32   = -547
	DISPID_SELLENGTH                           int32   = -548
	DISPID_REFRESH                             int32   = -550
	DISPID_DOCLICK                             int32   = -551
	DISPID_ABOUTBOX                            int32   = -552
	DISPID_ADDITEM                             int32   = -553
	DISPID_CLEAR                               int32   = -554
	DISPID_REMOVEITEM                          int32   = -555
	DISPID_CLICK                               int32   = -600
	DISPID_DBLCLICK                            int32   = -601
	DISPID_KEYDOWN                             int32   = -602
	DISPID_KEYPRESS                            int32   = -603
	DISPID_KEYUP                               int32   = -604
	DISPID_MOUSEDOWN                           int32   = -605
	DISPID_MOUSEMOVE                           int32   = -606
	DISPID_MOUSEUP                             int32   = -607
	DISPID_ERROREVENT                          int32   = -608
	DISPID_READYSTATECHANGE                    int32   = -609
	DISPID_CLICK_VALUE                         int32   = -610
	DISPID_RIGHTTOLEFT                         int32   = -611
	DISPID_TOPTOBOTTOM                         int32   = -612
	DISPID_THIS                                int32   = -613
	DISPID_AMBIENT_BACKCOLOR                   int32   = -701
	DISPID_AMBIENT_DISPLAYNAME                 int32   = -702
	DISPID_AMBIENT_FONT                        int32   = -703
	DISPID_AMBIENT_FORECOLOR                   int32   = -704
	DISPID_AMBIENT_LOCALEID                    int32   = -705
	DISPID_AMBIENT_MESSAGEREFLECT              int32   = -706
	DISPID_AMBIENT_SCALEUNITS                  int32   = -707
	DISPID_AMBIENT_TEXTALIGN                   int32   = -708
	DISPID_AMBIENT_USERMODE                    int32   = -709
	DISPID_AMBIENT_UIDEAD                      int32   = -710
	DISPID_AMBIENT_SHOWGRABHANDLES             int32   = -711
	DISPID_AMBIENT_SHOWHATCHING                int32   = -712
	DISPID_AMBIENT_DISPLAYASDEFAULT            int32   = -713
	DISPID_AMBIENT_SUPPORTSMNEMONICS           int32   = -714
	DISPID_AMBIENT_AUTOCLIP                    int32   = -715
	DISPID_AMBIENT_APPEARANCE                  int32   = -716
	DISPID_AMBIENT_CODEPAGE                    int32   = -725
	DISPID_AMBIENT_PALETTE                     int32   = -726
	DISPID_AMBIENT_CHARSET                     int32   = -727
	DISPID_AMBIENT_TRANSFERPRIORITY            int32   = -728
	DISPID_AMBIENT_RIGHTTOLEFT                 int32   = -732
	DISPID_AMBIENT_TOPTOBOTTOM                 int32   = -733
	DISPID_Name                                int32   = -800
	DISPID_Delete                              int32   = -801
	DISPID_Object                              int32   = -802
	DISPID_Parent                              int32   = -803
	DISPID_FONT_NAME                           uint32  = 0x0
	DISPID_FONT_SIZE                           uint32  = 0x2
	DISPID_FONT_BOLD                           uint32  = 0x3
	DISPID_FONT_ITALIC                         uint32  = 0x4
	DISPID_FONT_UNDER                          uint32  = 0x5
	DISPID_FONT_STRIKE                         uint32  = 0x6
	DISPID_FONT_WEIGHT                         uint32  = 0x7
	DISPID_FONT_CHARSET                        uint32  = 0x8
	DISPID_FONT_CHANGED                        uint32  = 0x9
	DISPID_PICT_HANDLE                         uint32  = 0x0
	DISPID_PICT_HPAL                           uint32  = 0x2
	DISPID_PICT_TYPE                           uint32  = 0x3
	DISPID_PICT_WIDTH                          uint32  = 0x4
	DISPID_PICT_HEIGHT                         uint32  = 0x5
	DISPID_PICT_RENDER                         uint32  = 0x6
	STDOLE_TLB                                 string  = "stdole2.tlb"
	STDTYPE_TLB                                string  = "stdole2.tlb"
	GC_WCH_SIBLING                             int32   = 1
	TIFLAGS_EXTENDDISPATCHONLY                 uint32  = 0x1
	OLECMDERR_E_NOTSUPPORTED                   int32   = -2147221248
	MSOCMDERR_E_FIRST                          int32   = -2147221248
	MSOCMDERR_E_NOTSUPPORTED                   int32   = -2147221248
	MSOCMDERR_E_DISABLED                       int32   = -2147221247
	MSOCMDERR_E_NOHELP                         int32   = -2147221246
	MSOCMDERR_E_CANCELED                       int32   = -2147221245
	MSOCMDERR_E_UNKNOWNGROUP                   int32   = -2147221244
	OLECMD_TASKDLGID_ONBEFOREUNLOAD            uint32  = 0x1
	OLECMDARGINDEX_SHOWPAGEACTIONMENU_HWND     uint32  = 0x0
	OLECMDARGINDEX_SHOWPAGEACTIONMENU_X        uint32  = 0x1
	OLECMDARGINDEX_SHOWPAGEACTIONMENU_Y        uint32  = 0x2
	OLECMDARGINDEX_ACTIVEXINSTALL_PUBLISHER    uint32  = 0x0
	OLECMDARGINDEX_ACTIVEXINSTALL_DISPLAYNAME  uint32  = 0x1
	OLECMDARGINDEX_ACTIVEXINSTALL_CLSID        uint32  = 0x2
	OLECMDARGINDEX_ACTIVEXINSTALL_INSTALLSCOPE uint32  = 0x3
	OLECMDARGINDEX_ACTIVEXINSTALL_SOURCEURL    uint32  = 0x4
	INSTALL_SCOPE_INVALID                      uint32  = 0x0
	INSTALL_SCOPE_MACHINE                      uint32  = 0x1
	INSTALL_SCOPE_USER                         uint32  = 0x2
	MK_ALT                                     uint32  = 0x20
	DROPEFFECT_NONE                            uint32  = 0x0
	DROPEFFECT_COPY                            uint32  = 0x1
	DROPEFFECT_MOVE                            uint32  = 0x2
	DROPEFFECT_LINK                            uint32  = 0x4
	DROPEFFECT_SCROLL                          uint32  = 0x80000000
	DD_DEFSCROLLINSET                          uint32  = 0xb
	DD_DEFSCROLLDELAY                          uint32  = 0x32
	DD_DEFSCROLLINTERVAL                       uint32  = 0x32
	DD_DEFDRAGDELAY                            uint32  = 0xc8
	DD_DEFDRAGMINDIST                          uint32  = 0x2
	OT_LINK                                    int32   = 1
	OT_EMBEDDED                                int32   = 2
	OT_STATIC                                  int32   = 3
	OLEVERB_PRIMARY                            uint32  = 0x0
	OF_SET                                     uint32  = 0x1
	OF_GET                                     uint32  = 0x2
	OF_HANDLER                                 uint32  = 0x4
	WIN32                                      uint32  = 0x64
	OLEIVERB_PRIMARY                           int32   = 0
	OLEIVERB_SHOW                              int32   = -1
	OLEIVERB_OPEN                              int32   = -2
	OLEIVERB_HIDE                              int32   = -3
	OLEIVERB_UIACTIVATE                        int32   = -4
	OLEIVERB_INPLACEACTIVATE                   int32   = -5
	OLEIVERB_DISCARDUNDOSTATE                  int32   = -6
	EMBDHLP_INPROC_HANDLER                     int32   = 0
	EMBDHLP_INPROC_SERVER                      int32   = 1
	EMBDHLP_CREATENOW                          int32   = 0
	EMBDHLP_DELAYCREATE                        int32   = 65536
	OLECREATE_LEAVERUNNING                     uint32  = 0x1
	IDC_OLEUIHELP                              uint32  = 0x63
	IDC_IO_CREATENEW                           uint32  = 0x834
	IDC_IO_CREATEFROMFILE                      uint32  = 0x835
	IDC_IO_LINKFILE                            uint32  = 0x836
	IDC_IO_OBJECTTYPELIST                      uint32  = 0x837
	IDC_IO_DISPLAYASICON                       uint32  = 0x838
	IDC_IO_CHANGEICON                          uint32  = 0x839
	IDC_IO_FILE                                uint32  = 0x83a
	IDC_IO_FILEDISPLAY                         uint32  = 0x83b
	IDC_IO_RESULTIMAGE                         uint32  = 0x83c
	IDC_IO_RESULTTEXT                          uint32  = 0x83d
	IDC_IO_ICONDISPLAY                         uint32  = 0x83e
	IDC_IO_OBJECTTYPETEXT                      uint32  = 0x83f
	IDC_IO_FILETEXT                            uint32  = 0x840
	IDC_IO_FILETYPE                            uint32  = 0x841
	IDC_IO_INSERTCONTROL                       uint32  = 0x842
	IDC_IO_ADDCONTROL                          uint32  = 0x843
	IDC_IO_CONTROLTYPELIST                     uint32  = 0x844
	IDC_PS_PASTE                               uint32  = 0x1f4
	IDC_PS_PASTELINK                           uint32  = 0x1f5
	IDC_PS_SOURCETEXT                          uint32  = 0x1f6
	IDC_PS_PASTELIST                           uint32  = 0x1f7
	IDC_PS_PASTELINKLIST                       uint32  = 0x1f8
	IDC_PS_DISPLAYLIST                         uint32  = 0x1f9
	IDC_PS_DISPLAYASICON                       uint32  = 0x1fa
	IDC_PS_ICONDISPLAY                         uint32  = 0x1fb
	IDC_PS_CHANGEICON                          uint32  = 0x1fc
	IDC_PS_RESULTIMAGE                         uint32  = 0x1fd
	IDC_PS_RESULTTEXT                          uint32  = 0x1fe
	IDC_CI_GROUP                               uint32  = 0x78
	IDC_CI_CURRENT                             uint32  = 0x79
	IDC_CI_CURRENTICON                         uint32  = 0x7a
	IDC_CI_DEFAULT                             uint32  = 0x7b
	IDC_CI_DEFAULTICON                         uint32  = 0x7c
	IDC_CI_FROMFILE                            uint32  = 0x7d
	IDC_CI_FROMFILEEDIT                        uint32  = 0x7e
	IDC_CI_ICONLIST                            uint32  = 0x7f
	IDC_CI_LABEL                               uint32  = 0x80
	IDC_CI_LABELEDIT                           uint32  = 0x81
	IDC_CI_BROWSE                              uint32  = 0x82
	IDC_CI_ICONDISPLAY                         uint32  = 0x83
	IDC_CV_OBJECTTYPE                          uint32  = 0x96
	IDC_CV_DISPLAYASICON                       uint32  = 0x98
	IDC_CV_CHANGEICON                          uint32  = 0x99
	IDC_CV_ACTIVATELIST                        uint32  = 0x9a
	IDC_CV_CONVERTTO                           uint32  = 0x9b
	IDC_CV_ACTIVATEAS                          uint32  = 0x9c
	IDC_CV_RESULTTEXT                          uint32  = 0x9d
	IDC_CV_CONVERTLIST                         uint32  = 0x9e
	IDC_CV_ICONDISPLAY                         uint32  = 0xa5
	IDC_EL_CHANGESOURCE                        uint32  = 0xc9
	IDC_EL_AUTOMATIC                           uint32  = 0xca
	IDC_EL_CANCELLINK                          uint32  = 0xd1
	IDC_EL_UPDATENOW                           uint32  = 0xd2
	IDC_EL_OPENSOURCE                          uint32  = 0xd3
	IDC_EL_MANUAL                              uint32  = 0xd4
	IDC_EL_LINKSOURCE                          uint32  = 0xd8
	IDC_EL_LINKTYPE                            uint32  = 0xd9
	IDC_EL_LINKSLISTBOX                        uint32  = 0xce
	IDC_EL_COL1                                uint32  = 0xdc
	IDC_EL_COL2                                uint32  = 0xdd
	IDC_EL_COL3                                uint32  = 0xde
	IDC_BZ_RETRY                               uint32  = 0x258
	IDC_BZ_ICON                                uint32  = 0x259
	IDC_BZ_MESSAGE1                            uint32  = 0x25a
	IDC_BZ_SWITCHTO                            uint32  = 0x25c
	IDC_UL_METER                               uint32  = 0x405
	IDC_UL_STOP                                uint32  = 0x406
	IDC_UL_PERCENT                             uint32  = 0x407
	IDC_UL_PROGRESS                            uint32  = 0x408
	IDC_PU_LINKS                               uint32  = 0x384
	IDC_PU_TEXT                                uint32  = 0x385
	IDC_PU_CONVERT                             uint32  = 0x386
	IDC_PU_ICON                                uint32  = 0x38c
	IDC_GP_OBJECTNAME                          uint32  = 0x3f1
	IDC_GP_OBJECTTYPE                          uint32  = 0x3f2
	IDC_GP_OBJECTSIZE                          uint32  = 0x3f3
	IDC_GP_CONVERT                             uint32  = 0x3f5
	IDC_GP_OBJECTICON                          uint32  = 0x3f6
	IDC_GP_OBJECTLOCATION                      uint32  = 0x3fe
	IDC_VP_PERCENT                             uint32  = 0x3e8
	IDC_VP_CHANGEICON                          uint32  = 0x3e9
	IDC_VP_EDITABLE                            uint32  = 0x3ea
	IDC_VP_ASICON                              uint32  = 0x3eb
	IDC_VP_RELATIVE                            uint32  = 0x3ed
	IDC_VP_SPIN                                uint32  = 0x3ee
	IDC_VP_SCALETXT                            uint32  = 0x40a
	IDC_VP_ICONDISPLAY                         uint32  = 0x3fd
	IDC_VP_RESULTIMAGE                         uint32  = 0x409
	IDC_LP_OPENSOURCE                          uint32  = 0x3ee
	IDC_LP_UPDATENOW                           uint32  = 0x3ef
	IDC_LP_BREAKLINK                           uint32  = 0x3f0
	IDC_LP_LINKSOURCE                          uint32  = 0x3f4
	IDC_LP_CHANGESOURCE                        uint32  = 0x3f7
	IDC_LP_AUTOMATIC                           uint32  = 0x3f8
	IDC_LP_MANUAL                              uint32  = 0x3f9
	IDC_LP_DATE                                uint32  = 0x3fa
	IDC_LP_TIME                                uint32  = 0x3fb
	IDD_INSERTOBJECT                           uint32  = 0x3e8
	IDD_CHANGEICON                             uint32  = 0x3e9
	IDD_CONVERT                                uint32  = 0x3ea
	IDD_PASTESPECIAL                           uint32  = 0x3eb
	IDD_EDITLINKS                              uint32  = 0x3ec
	IDD_BUSY                                   uint32  = 0x3ee
	IDD_UPDATELINKS                            uint32  = 0x3ef
	IDD_CHANGESOURCE                           uint32  = 0x3f1
	IDD_INSERTFILEBROWSE                       uint32  = 0x3f2
	IDD_CHANGEICONBROWSE                       uint32  = 0x3f3
	IDD_CONVERTONLY                            uint32  = 0x3f4
	IDD_CHANGESOURCE4                          uint32  = 0x3f5
	IDD_GNRLPROPS                              uint32  = 0x44c
	IDD_VIEWPROPS                              uint32  = 0x44d
	IDD_LINKPROPS                              uint32  = 0x44e
	IDD_CONVERT4                               uint32  = 0x44f
	IDD_CONVERTONLY4                           uint32  = 0x450
	IDD_EDITLINKS4                             uint32  = 0x451
	IDD_GNRLPROPS4                             uint32  = 0x452
	IDD_LINKPROPS4                             uint32  = 0x453
	IDD_PASTESPECIAL4                          uint32  = 0x454
	IDD_CANNOTUPDATELINK                       uint32  = 0x3f0
	IDD_LINKSOURCEUNAVAILABLE                  uint32  = 0x3fc
	IDD_SERVERNOTFOUND                         uint32  = 0x3ff
	IDD_OUTOFMEMORY                            uint32  = 0x400
	IDD_SERVERNOTREGW                          uint32  = 0x3fd
	IDD_LINKTYPECHANGEDW                       uint32  = 0x3fe
	IDD_SERVERNOTREGA                          uint32  = 0x401
	IDD_LINKTYPECHANGEDA                       uint32  = 0x402
	IDD_SERVERNOTREG                           uint32  = 0x3fd
	IDD_LINKTYPECHANGED                        uint32  = 0x3fe
	OLESTDDELIM                                string  = "\\"
	SZOLEUI_MSG_HELP                           string  = "OLEUI_MSG_HELP"
	SZOLEUI_MSG_ENDDIALOG                      string  = "OLEUI_MSG_ENDDIALOG"
	SZOLEUI_MSG_BROWSE                         string  = "OLEUI_MSG_BROWSE"
	SZOLEUI_MSG_CHANGEICON                     string  = "OLEUI_MSG_CHANGEICON"
	SZOLEUI_MSG_CLOSEBUSYDIALOG                string  = "OLEUI_MSG_CLOSEBUSYDIALOG"
	SZOLEUI_MSG_CONVERT                        string  = "OLEUI_MSG_CONVERT"
	SZOLEUI_MSG_CHANGESOURCE                   string  = "OLEUI_MSG_CHANGESOURCE"
	SZOLEUI_MSG_ADDCONTROL                     string  = "OLEUI_MSG_ADDCONTROL"
	SZOLEUI_MSG_BROWSE_OFN                     string  = "OLEUI_MSG_BROWSE_OFN"
	ID_BROWSE_CHANGEICON                       uint32  = 0x1
	ID_BROWSE_INSERTFILE                       uint32  = 0x2
	ID_BROWSE_ADDCONTROL                       uint32  = 0x3
	ID_BROWSE_CHANGESOURCE                     uint32  = 0x4
	OLEUI_FALSE                                uint32  = 0x0
	OLEUI_SUCCESS                              uint32  = 0x1
	OLEUI_OK                                   uint32  = 0x1
	OLEUI_CANCEL                               uint32  = 0x2
	OLEUI_ERR_STANDARDMIN                      uint32  = 0x64
	OLEUI_ERR_OLEMEMALLOC                      uint32  = 0x64
	OLEUI_ERR_STRUCTURENULL                    uint32  = 0x65
	OLEUI_ERR_STRUCTUREINVALID                 uint32  = 0x66
	OLEUI_ERR_CBSTRUCTINCORRECT                uint32  = 0x67
	OLEUI_ERR_HWNDOWNERINVALID                 uint32  = 0x68
	OLEUI_ERR_LPSZCAPTIONINVALID               uint32  = 0x69
	OLEUI_ERR_LPFNHOOKINVALID                  uint32  = 0x6a
	OLEUI_ERR_HINSTANCEINVALID                 uint32  = 0x6b
	OLEUI_ERR_LPSZTEMPLATEINVALID              uint32  = 0x6c
	OLEUI_ERR_HRESOURCEINVALID                 uint32  = 0x6d
	OLEUI_ERR_FINDTEMPLATEFAILURE              uint32  = 0x6e
	OLEUI_ERR_LOADTEMPLATEFAILURE              uint32  = 0x6f
	OLEUI_ERR_DIALOGFAILURE                    uint32  = 0x70
	OLEUI_ERR_LOCALMEMALLOC                    uint32  = 0x71
	OLEUI_ERR_GLOBALMEMALLOC                   uint32  = 0x72
	OLEUI_ERR_LOADSTRING                       uint32  = 0x73
	OLEUI_ERR_STANDARDMAX                      uint32  = 0x74
	IOF_SHOWHELP                               int32   = 1
	IOF_SELECTCREATENEW                        int32   = 2
	IOF_SELECTCREATEFROMFILE                   int32   = 4
	IOF_CHECKLINK                              int32   = 8
	IOF_CHECKDISPLAYASICON                     int32   = 16
	IOF_CREATENEWOBJECT                        int32   = 32
	IOF_CREATEFILEOBJECT                       int32   = 64
	IOF_CREATELINKOBJECT                       int32   = 128
	IOF_DISABLELINK                            int32   = 256
	IOF_VERIFYSERVERSEXIST                     int32   = 512
	IOF_DISABLEDISPLAYASICON                   int32   = 1024
	IOF_HIDECHANGEICON                         int32   = 2048
	IOF_SHOWINSERTCONTROL                      int32   = 4096
	IOF_SELECTCREATECONTROL                    int32   = 8192
	OLEUI_IOERR_LPSZFILEINVALID                uint32  = 0x74
	OLEUI_IOERR_LPSZLABELINVALID               uint32  = 0x75
	OLEUI_IOERR_HICONINVALID                   uint32  = 0x76
	OLEUI_IOERR_LPFORMATETCINVALID             uint32  = 0x77
	OLEUI_IOERR_PPVOBJINVALID                  uint32  = 0x78
	OLEUI_IOERR_LPIOLECLIENTSITEINVALID        uint32  = 0x79
	OLEUI_IOERR_LPISTORAGEINVALID              uint32  = 0x7a
	OLEUI_IOERR_SCODEHASERROR                  uint32  = 0x7b
	OLEUI_IOERR_LPCLSIDEXCLUDEINVALID          uint32  = 0x7c
	OLEUI_IOERR_CCHFILEINVALID                 uint32  = 0x7d
	PS_MAXLINKTYPES                            uint32  = 0x8
	PSF_SHOWHELP                               int32   = 1
	PSF_SELECTPASTE                            int32   = 2
	PSF_SELECTPASTELINK                        int32   = 4
	PSF_CHECKDISPLAYASICON                     int32   = 8
	PSF_DISABLEDISPLAYASICON                   int32   = 16
	PSF_HIDECHANGEICON                         int32   = 32
	PSF_STAYONCLIPBOARDCHANGE                  int32   = 64
	PSF_NOREFRESHDATAOBJECT                    int32   = 128
	OLEUI_IOERR_SRCDATAOBJECTINVALID           uint32  = 0x74
	OLEUI_IOERR_ARRPASTEENTRIESINVALID         uint32  = 0x75
	OLEUI_IOERR_ARRLINKTYPESINVALID            uint32  = 0x76
	OLEUI_PSERR_CLIPBOARDCHANGED               uint32  = 0x77
	OLEUI_PSERR_GETCLIPBOARDFAILED             uint32  = 0x78
	OLEUI_ELERR_LINKCNTRNULL                   uint32  = 0x74
	OLEUI_ELERR_LINKCNTRINVALID                uint32  = 0x75
	ELF_SHOWHELP                               int32   = 1
	ELF_DISABLEUPDATENOW                       int32   = 2
	ELF_DISABLEOPENSOURCE                      int32   = 4
	ELF_DISABLECHANGESOURCE                    int32   = 8
	ELF_DISABLECANCELLINK                      int32   = 16
	CIF_SHOWHELP                               int32   = 1
	CIF_SELECTCURRENT                          int32   = 2
	CIF_SELECTDEFAULT                          int32   = 4
	CIF_SELECTFROMFILE                         int32   = 8
	CIF_USEICONEXE                             int32   = 16
	OLEUI_CIERR_MUSTHAVECLSID                  uint32  = 0x74
	OLEUI_CIERR_MUSTHAVECURRENTMETAFILE        uint32  = 0x75
	OLEUI_CIERR_SZICONEXEINVALID               uint32  = 0x76
	PROP_HWND_CHGICONDLG                       string  = "HWND_CIDLG"
	CF_SHOWHELPBUTTON                          int32   = 1
	CF_SETCONVERTDEFAULT                       int32   = 2
	CF_SETACTIVATEDEFAULT                      int32   = 4
	CF_SELECTCONVERTTO                         int32   = 8
	CF_SELECTACTIVATEAS                        int32   = 16
	CF_DISABLEDISPLAYASICON                    int32   = 32
	CF_DISABLEACTIVATEAS                       int32   = 64
	CF_HIDECHANGEICON                          int32   = 128
	CF_CONVERTONLY                             int32   = 256
	OLEUI_CTERR_CLASSIDINVALID                 uint32  = 0x75
	OLEUI_CTERR_DVASPECTINVALID                uint32  = 0x76
	OLEUI_CTERR_CBFORMATINVALID                uint32  = 0x77
	OLEUI_CTERR_HMETAPICTINVALID               uint32  = 0x78
	OLEUI_CTERR_STRINGINVALID                  uint32  = 0x79
	BZ_DISABLECANCELBUTTON                     int32   = 1
	BZ_DISABLESWITCHTOBUTTON                   int32   = 2
	BZ_DISABLERETRYBUTTON                      int32   = 4
	BZ_NOTRESPONDINGDIALOG                     int32   = 8
	OLEUI_BZERR_HTASKINVALID                   uint32  = 0x74
	OLEUI_BZ_SWITCHTOSELECTED                  uint32  = 0x75
	OLEUI_BZ_RETRYSELECTED                     uint32  = 0x76
	OLEUI_BZ_CALLUNBLOCKED                     uint32  = 0x77
	CSF_SHOWHELP                               int32   = 1
	CSF_VALIDSOURCE                            int32   = 2
	CSF_ONLYGETSOURCE                          int32   = 4
	CSF_EXPLORER                               int32   = 8
	OLEUI_CSERR_LINKCNTRNULL                   uint32  = 0x74
	OLEUI_CSERR_LINKCNTRINVALID                uint32  = 0x75
	OLEUI_CSERR_FROMNOTNULL                    uint32  = 0x76
	OLEUI_CSERR_TONOTNULL                      uint32  = 0x77
	OLEUI_CSERR_SOURCENULL                     uint32  = 0x78
	OLEUI_CSERR_SOURCEINVALID                  uint32  = 0x79
	OLEUI_CSERR_SOURCEPARSERROR                uint32  = 0x7a
	OLEUI_CSERR_SOURCEPARSEERROR               uint32  = 0x7a
	VPF_SELECTRELATIVE                         int32   = 1
	VPF_DISABLERELATIVE                        int32   = 2
	VPF_DISABLESCALE                           int32   = 4
	OPF_OBJECTISLINK                           int32   = 1
	OPF_NOFILLDEFAULT                          int32   = 2
	OPF_SHOWHELP                               int32   = 4
	OPF_DISABLECONVERT                         int32   = 8
	OLEUI_OPERR_SUBPROPNULL                    uint32  = 0x74
	OLEUI_OPERR_SUBPROPINVALID                 uint32  = 0x75
	OLEUI_OPERR_PROPSHEETNULL                  uint32  = 0x76
	OLEUI_OPERR_PROPSHEETINVALID               uint32  = 0x77
	OLEUI_OPERR_SUPPROP                        uint32  = 0x78
	OLEUI_OPERR_PROPSINVALID                   uint32  = 0x79
	OLEUI_OPERR_PAGESINCORRECT                 uint32  = 0x7a
	OLEUI_OPERR_INVALIDPAGES                   uint32  = 0x7b
	OLEUI_OPERR_NOTSUPPORTED                   uint32  = 0x7c
	OLEUI_OPERR_DLGPROCNOTNULL                 uint32  = 0x7d
	OLEUI_OPERR_LPARAMNOTZERO                  uint32  = 0x7e
	OLEUI_GPERR_STRINGINVALID                  uint32  = 0x7f
	OLEUI_GPERR_CLASSIDINVALID                 uint32  = 0x80
	OLEUI_GPERR_LPCLSIDEXCLUDEINVALID          uint32  = 0x81
	OLEUI_GPERR_CBFORMATINVALID                uint32  = 0x82
	OLEUI_VPERR_METAPICTINVALID                uint32  = 0x83
	OLEUI_VPERR_DVASPECTINVALID                uint32  = 0x84
	OLEUI_LPERR_LINKCNTRNULL                   uint32  = 0x85
	OLEUI_LPERR_LINKCNTRINVALID                uint32  = 0x86
	OLEUI_OPERR_PROPERTYSHEET                  uint32  = 0x87
	OLEUI_OPERR_OBJINFOINVALID                 uint32  = 0x88
	OLEUI_OPERR_LINKINFOINVALID                uint32  = 0x89
	OLEUI_QUERY_GETCLASSID                     uint32  = 0xff00
	OLEUI_QUERY_LINKBROKEN                     uint32  = 0xff01
	FADF_AUTO                                  uint32  = 0x1
	FADF_STATIC                                uint32  = 0x2
	FADF_EMBEDDED                              uint32  = 0x4
	FADF_FIXEDSIZE                             uint32  = 0x10
	FADF_RECORD                                uint32  = 0x20
	FADF_HAVEIID                               uint32  = 0x40
	FADF_HAVEVARTYPE                           uint32  = 0x80
	FADF_BSTR                                  uint32  = 0x100
	FADF_UNKNOWN                               uint32  = 0x200
	FADF_DISPATCH                              uint32  = 0x400
	FADF_VARIANT                               uint32  = 0x800
	FADF_RESERVED                              uint32  = 0xf008
	PARAMFLAG_NONE                             uint32  = 0x0
	PARAMFLAG_FIN                              uint32  = 0x1
	PARAMFLAG_FOUT                             uint32  = 0x2
	PARAMFLAG_FLCID                            uint32  = 0x4
	PARAMFLAG_FRETVAL                          uint32  = 0x8
	PARAMFLAG_FOPT                             uint32  = 0x10
	PARAMFLAG_FHASDEFAULT                      uint32  = 0x20
	PARAMFLAG_FHASCUSTDATA                     uint32  = 0x40
	IDLFLAG_NONE                               uint32  = 0x0
	IDLFLAG_FIN                                uint32  = 0x1
	IDLFLAG_FOUT                               uint32  = 0x2
	IDLFLAG_FLCID                              uint32  = 0x4
	IDLFLAG_FRETVAL                            uint32  = 0x8
	IMPLTYPEFLAG_FDEFAULT                      uint32  = 0x1
	IMPLTYPEFLAG_FSOURCE                       uint32  = 0x2
	IMPLTYPEFLAG_FRESTRICTED                   uint32  = 0x4
	IMPLTYPEFLAG_FDEFAULTVTABLE                uint32  = 0x8
	DISPID_UNKNOWN                             int32   = -1
	DISPID_VALUE                               uint32  = 0x0
	DISPID_PROPERTYPUT                         int32   = -3
	DISPID_NEWENUM                             int32   = -4
	DISPID_EVALUATE                            int32   = -5
	DISPID_CONSTRUCTOR                         int32   = -6
	DISPID_DESTRUCTOR                          int32   = -7
	DISPID_COLLECT                             int32   = -8
	STDOLE_MAJORVERNUM                         uint32  = 0x1
	STDOLE_MINORVERNUM                         uint32  = 0x0
	STDOLE_LCID                                uint32  = 0x0
	STDOLE2_MAJORVERNUM                        uint32  = 0x2
	STDOLE2_MINORVERNUM                        uint32  = 0x0
	STDOLE2_LCID                               uint32  = 0x0
	VARIANT_NOVALUEPROP                        uint32  = 0x1
	VARIANT_ALPHABOOL                          uint32  = 0x2
	VARIANT_NOUSEROVERRIDE                     uint32  = 0x4
	VARIANT_CALENDAR_HIJRI                     uint32  = 0x8
	VARIANT_LOCALBOOL                          uint32  = 0x10
	VARIANT_CALENDAR_THAI                      uint32  = 0x20
	VARIANT_CALENDAR_GREGORIAN                 uint32  = 0x40
	VARIANT_USE_NLS                            uint32  = 0x80
	LOCALE_USE_NLS                             uint32  = 0x10000000
	VTDATEGRE_MAX                              uint32  = 0x2d2481
	VTDATEGRE_MIN                              int32   = -657434
	NUMPRS_LEADING_WHITE                       uint32  = 0x1
	NUMPRS_TRAILING_WHITE                      uint32  = 0x2
	NUMPRS_LEADING_PLUS                        uint32  = 0x4
	NUMPRS_TRAILING_PLUS                       uint32  = 0x8
	NUMPRS_LEADING_MINUS                       uint32  = 0x10
	NUMPRS_TRAILING_MINUS                      uint32  = 0x20
	NUMPRS_HEX_OCT                             uint32  = 0x40
	NUMPRS_PARENS                              uint32  = 0x80
	NUMPRS_DECIMAL                             uint32  = 0x100
	NUMPRS_THOUSANDS                           uint32  = 0x200
	NUMPRS_CURRENCY                            uint32  = 0x400
	NUMPRS_EXPONENT                            uint32  = 0x800
	NUMPRS_USE_ALL                             uint32  = 0x1000
	NUMPRS_STD                                 uint32  = 0x1fff
	NUMPRS_NEG                                 uint32  = 0x10000
	NUMPRS_INEXACT                             uint32  = 0x20000
	VARCMP_LT                                  uint32  = 0x0
	VARCMP_EQ                                  uint32  = 0x1
	VARCMP_GT                                  uint32  = 0x2
	VARCMP_NULL                                uint32  = 0x3
	MEMBERID_NIL                               int32   = -1
	ID_DEFAULTINST                             int32   = -2
	DISPATCH_METHOD                            uint32  = 0x1
	DISPATCH_PROPERTYGET                       uint32  = 0x2
	DISPATCH_PROPERTYPUT                       uint32  = 0x4
	DISPATCH_PROPERTYPUTREF                    uint32  = 0x8
	LOAD_TLB_AS_32BIT                          uint32  = 0x20
	LOAD_TLB_AS_64BIT                          uint32  = 0x40
	ACTIVEOBJECT_STRONG                        uint32  = 0x0
	ACTIVEOBJECT_WEAK                          uint32  = 0x1
	FdexNameCaseSensitive                      int32   = 1
	FdexNameEnsure                             int32   = 2
	FdexNameImplicit                           int32   = 4
	FdexNameCaseInsensitive                    int32   = 8
	FdexNameInternal                           int32   = 16
	FdexNameNoDynamicProperties                int32   = 32
	FdexPropCanGet                             int32   = 1
	FdexPropCannotGet                          int32   = 2
	FdexPropCanPut                             int32   = 4
	FdexPropCannotPut                          int32   = 8
	FdexPropCanPutRef                          int32   = 16
	FdexPropCannotPutRef                       int32   = 32
	FdexPropNoSideEffects                      int32   = 64
	FdexPropDynamicType                        int32   = 128
	FdexPropCanCall                            int32   = 256
	FdexPropCannotCall                         int32   = 512
	FdexPropCanConstruct                       int32   = 1024
	FdexPropCannotConstruct                    int32   = 2048
	FdexPropCanSourceEvents                    int32   = 4096
	FdexPropCannotSourceEvents                 int32   = 8192
	FdexEnumDefault                            int32   = 1
	FdexEnumAll                                int32   = 2
	DISPATCH_CONSTRUCT                         uint32  = 0x4000
	DISPID_STARTENUM                           int32   = -1
)

var (
	CLSID_CFontPropPage = syscall.GUID{0x0BE35200, 0x8F91, 0x11CE,
		[8]byte{0x9D, 0xE3, 0x00, 0xAA, 0x00, 0x4B, 0xB8, 0x51}}

	CLSID_CColorPropPage = syscall.GUID{0x0BE35201, 0x8F91, 0x11CE,
		[8]byte{0x9D, 0xE3, 0x00, 0xAA, 0x00, 0x4B, 0xB8, 0x51}}

	CLSID_CPicturePropPage = syscall.GUID{0x0BE35202, 0x8F91, 0x11CE,
		[8]byte{0x9D, 0xE3, 0x00, 0xAA, 0x00, 0x4B, 0xB8, 0x51}}

	CLSID_PersistPropset = syscall.GUID{0xFB8F0821, 0x0164, 0x101B,
		[8]byte{0x84, 0xED, 0x08, 0x00, 0x2B, 0x2E, 0xC7, 0x13}}

	CLSID_ConvertVBX = syscall.GUID{0xFB8F0822, 0x0164, 0x101B,
		[8]byte{0x84, 0xED, 0x08, 0x00, 0x2B, 0x2E, 0xC7, 0x13}}

	CLSID_StdFont = syscall.GUID{0x0BE35203, 0x8F91, 0x11CE,
		[8]byte{0x9D, 0xE3, 0x00, 0xAA, 0x00, 0x4B, 0xB8, 0x51}}

	CLSID_StdPicture = syscall.GUID{0x0BE35204, 0x8F91, 0x11CE,
		[8]byte{0x9D, 0xE3, 0x00, 0xAA, 0x00, 0x4B, 0xB8, 0x51}}

	GUID_HIMETRIC = syscall.GUID{0x66504300, 0xBE0F, 0x101A,
		[8]byte{0x8B, 0xBB, 0x00, 0xAA, 0x00, 0x30, 0x0C, 0xAB}}

	GUID_COLOR = syscall.GUID{0x66504301, 0xBE0F, 0x101A,
		[8]byte{0x8B, 0xBB, 0x00, 0xAA, 0x00, 0x30, 0x0C, 0xAB}}

	GUID_XPOSPIXEL = syscall.GUID{0x66504302, 0xBE0F, 0x101A,
		[8]byte{0x8B, 0xBB, 0x00, 0xAA, 0x00, 0x30, 0x0C, 0xAB}}

	GUID_YPOSPIXEL = syscall.GUID{0x66504303, 0xBE0F, 0x101A,
		[8]byte{0x8B, 0xBB, 0x00, 0xAA, 0x00, 0x30, 0x0C, 0xAB}}

	GUID_XSIZEPIXEL = syscall.GUID{0x66504304, 0xBE0F, 0x101A,
		[8]byte{0x8B, 0xBB, 0x00, 0xAA, 0x00, 0x30, 0x0C, 0xAB}}

	GUID_YSIZEPIXEL = syscall.GUID{0x66504305, 0xBE0F, 0x101A,
		[8]byte{0x8B, 0xBB, 0x00, 0xAA, 0x00, 0x30, 0x0C, 0xAB}}

	GUID_XPOS = syscall.GUID{0x66504306, 0xBE0F, 0x101A,
		[8]byte{0x8B, 0xBB, 0x00, 0xAA, 0x00, 0x30, 0x0C, 0xAB}}

	GUID_YPOS = syscall.GUID{0x66504307, 0xBE0F, 0x101A,
		[8]byte{0x8B, 0xBB, 0x00, 0xAA, 0x00, 0x30, 0x0C, 0xAB}}

	GUID_XSIZE = syscall.GUID{0x66504308, 0xBE0F, 0x101A,
		[8]byte{0x8B, 0xBB, 0x00, 0xAA, 0x00, 0x30, 0x0C, 0xAB}}

	GUID_YSIZE = syscall.GUID{0x66504309, 0xBE0F, 0x101A,
		[8]byte{0x8B, 0xBB, 0x00, 0xAA, 0x00, 0x30, 0x0C, 0xAB}}

	GUID_TRISTATE = syscall.GUID{0x6650430A, 0xBE0F, 0x101A,
		[8]byte{0x8B, 0xBB, 0x00, 0xAA, 0x00, 0x30, 0x0C, 0xAB}}

	GUID_OPTIONVALUEEXCLUSIVE = syscall.GUID{0x6650430B, 0xBE0F, 0x101A,
		[8]byte{0x8B, 0xBB, 0x00, 0xAA, 0x00, 0x30, 0x0C, 0xAB}}

	GUID_CHECKVALUEEXCLUSIVE = syscall.GUID{0x6650430C, 0xBE0F, 0x101A,
		[8]byte{0x8B, 0xBB, 0x00, 0xAA, 0x00, 0x30, 0x0C, 0xAB}}

	GUID_FONTNAME = syscall.GUID{0x6650430D, 0xBE0F, 0x101A,
		[8]byte{0x8B, 0xBB, 0x00, 0xAA, 0x00, 0x30, 0x0C, 0xAB}}

	GUID_FONTSIZE = syscall.GUID{0x6650430E, 0xBE0F, 0x101A,
		[8]byte{0x8B, 0xBB, 0x00, 0xAA, 0x00, 0x30, 0x0C, 0xAB}}

	GUID_FONTBOLD = syscall.GUID{0x6650430F, 0xBE0F, 0x101A,
		[8]byte{0x8B, 0xBB, 0x00, 0xAA, 0x00, 0x30, 0x0C, 0xAB}}

	GUID_FONTITALIC = syscall.GUID{0x66504310, 0xBE0F, 0x101A,
		[8]byte{0x8B, 0xBB, 0x00, 0xAA, 0x00, 0x30, 0x0C, 0xAB}}

	GUID_FONTUNDERSCORE = syscall.GUID{0x66504311, 0xBE0F, 0x101A,
		[8]byte{0x8B, 0xBB, 0x00, 0xAA, 0x00, 0x30, 0x0C, 0xAB}}

	GUID_FONTSTRIKETHROUGH = syscall.GUID{0x66504312, 0xBE0F, 0x101A,
		[8]byte{0x8B, 0xBB, 0x00, 0xAA, 0x00, 0x30, 0x0C, 0xAB}}

	GUID_HANDLE = syscall.GUID{0x66504313, 0xBE0F, 0x101A,
		[8]byte{0x8B, 0xBB, 0x00, 0xAA, 0x00, 0x30, 0x0C, 0xAB}}

	SID_VariantConversion = syscall.GUID{0x1F101481, 0xBCCD, 0x11D0,
		[8]byte{0x93, 0x36, 0x00, 0xA0, 0xC9, 0x0D, 0xCA, 0xA9}}

	SID_GetCaller = syscall.GUID{0x4717CC40, 0xBCB9, 0x11D0,
		[8]byte{0x93, 0x36, 0x00, 0xA0, 0xC9, 0x0D, 0xCA, 0xA9}}

	SID_ProvideRuntimeContext = syscall.GUID{0x74A5040C, 0xDD0C, 0x48F0,
		[8]byte{0xAC, 0x85, 0x19, 0x4C, 0x32, 0x59, 0x18, 0x0A}}
)

// enums

// enum
// flags
type UPDFCACHE_FLAGS uint32

const (
	UPDFCACHE_ALL                  UPDFCACHE_FLAGS = 2147483647
	UPDFCACHE_ALLBUTNODATACACHE    UPDFCACHE_FLAGS = 2147483646
	UPDFCACHE_NORMALCACHE          UPDFCACHE_FLAGS = 8
	UPDFCACHE_IFBLANK              UPDFCACHE_FLAGS = 16
	UPDFCACHE_ONLYIFBLANK          UPDFCACHE_FLAGS = 2147483648
	UPDFCACHE_NODATACACHE          UPDFCACHE_FLAGS = 1
	UPDFCACHE_ONSAVECACHE          UPDFCACHE_FLAGS = 2
	UPDFCACHE_ONSTOPCACHE          UPDFCACHE_FLAGS = 4
	UPDFCACHE_IFBLANKORONSAVECACHE UPDFCACHE_FLAGS = 18
)

// enum
type ENUM_CONTROLS_WHICH_FLAGS uint32

const (
	GCW_WCH_SIBLING    ENUM_CONTROLS_WHICH_FLAGS = 1
	GC_WCH_CONTAINER   ENUM_CONTROLS_WHICH_FLAGS = 2
	GC_WCH_CONTAINED   ENUM_CONTROLS_WHICH_FLAGS = 3
	GC_WCH_ALL         ENUM_CONTROLS_WHICH_FLAGS = 4
	GC_WCH_FREVERSEDIR ENUM_CONTROLS_WHICH_FLAGS = 134217728
	GC_WCH_FONLYAFTER  ENUM_CONTROLS_WHICH_FLAGS = 268435456
	GC_WCH_FONLYBEFORE ENUM_CONTROLS_WHICH_FLAGS = 536870912
	GC_WCH_FSELECTED   ENUM_CONTROLS_WHICH_FLAGS = 1073741824
)

// enum
type MULTICLASSINFO_FLAGS uint32

const (
	MULTICLASSINFO_GETTYPEINFO           MULTICLASSINFO_FLAGS = 1
	MULTICLASSINFO_GETNUMRESERVEDDISPIDS MULTICLASSINFO_FLAGS = 2
	MULTICLASSINFO_GETIIDPRIMARY         MULTICLASSINFO_FLAGS = 4
	MULTICLASSINFO_GETIIDSOURCE          MULTICLASSINFO_FLAGS = 8
)

// enum
type VARENUM int32

const (
	VT_EMPTY            VARENUM = 0
	VT_NULL             VARENUM = 1
	VT_I2               VARENUM = 2
	VT_I4               VARENUM = 3
	VT_R4               VARENUM = 4
	VT_R8               VARENUM = 5
	VT_CY               VARENUM = 6
	VT_DATE             VARENUM = 7
	VT_BSTR             VARENUM = 8
	VT_DISPATCH         VARENUM = 9
	VT_ERROR            VARENUM = 10
	VT_BOOL             VARENUM = 11
	VT_VARIANT          VARENUM = 12
	VT_UNKNOWN          VARENUM = 13
	VT_DECIMAL          VARENUM = 14
	VT_I1               VARENUM = 16
	VT_UI1              VARENUM = 17
	VT_UI2              VARENUM = 18
	VT_UI4              VARENUM = 19
	VT_I8               VARENUM = 20
	VT_UI8              VARENUM = 21
	VT_INT              VARENUM = 22
	VT_UINT             VARENUM = 23
	VT_VOID             VARENUM = 24
	VT_HRESULT          VARENUM = 25
	VT_PTR              VARENUM = 26
	VT_SAFEARRAY        VARENUM = 27
	VT_CARRAY           VARENUM = 28
	VT_USERDEFINED      VARENUM = 29
	VT_LPSTR            VARENUM = 30
	VT_LPWSTR           VARENUM = 31
	VT_RECORD           VARENUM = 36
	VT_INT_PTR          VARENUM = 37
	VT_UINT_PTR         VARENUM = 38
	VT_FILETIME         VARENUM = 64
	VT_BLOB             VARENUM = 65
	VT_STREAM           VARENUM = 66
	VT_STORAGE          VARENUM = 67
	VT_STREAMED_OBJECT  VARENUM = 68
	VT_STORED_OBJECT    VARENUM = 69
	VT_BLOB_OBJECT      VARENUM = 70
	VT_CF               VARENUM = 71
	VT_CLSID            VARENUM = 72
	VT_VERSIONED_STREAM VARENUM = 73
	VT_BSTR_BLOB        VARENUM = 4095
	VT_VECTOR           VARENUM = 4096
	VT_ARRAY            VARENUM = 8192
	VT_BYREF            VARENUM = 16384
	VT_RESERVED         VARENUM = 32768
	VT_ILLEGAL          VARENUM = 65535
	VT_ILLEGALMASKED    VARENUM = 4095
	VT_TYPEMASK         VARENUM = 4095
)

// enum
type SF_TYPE int32

const (
	SF_ERROR    SF_TYPE = 10
	SF_I1       SF_TYPE = 16
	SF_I2       SF_TYPE = 2
	SF_I4       SF_TYPE = 3
	SF_I8       SF_TYPE = 20
	SF_BSTR     SF_TYPE = 8
	SF_UNKNOWN  SF_TYPE = 13
	SF_DISPATCH SF_TYPE = 9
	SF_VARIANT  SF_TYPE = 12
	SF_RECORD   SF_TYPE = 36
	SF_HAVEIID  SF_TYPE = 32781
)

// enum
type TYPEFLAGS int32

const (
	TYPEFLAG_FAPPOBJECT     TYPEFLAGS = 1
	TYPEFLAG_FCANCREATE     TYPEFLAGS = 2
	TYPEFLAG_FLICENSED      TYPEFLAGS = 4
	TYPEFLAG_FPREDECLID     TYPEFLAGS = 8
	TYPEFLAG_FHIDDEN        TYPEFLAGS = 16
	TYPEFLAG_FCONTROL       TYPEFLAGS = 32
	TYPEFLAG_FDUAL          TYPEFLAGS = 64
	TYPEFLAG_FNONEXTENSIBLE TYPEFLAGS = 128
	TYPEFLAG_FOLEAUTOMATION TYPEFLAGS = 256
	TYPEFLAG_FRESTRICTED    TYPEFLAGS = 512
	TYPEFLAG_FAGGREGATABLE  TYPEFLAGS = 1024
	TYPEFLAG_FREPLACEABLE   TYPEFLAGS = 2048
	TYPEFLAG_FDISPATCHABLE  TYPEFLAGS = 4096
	TYPEFLAG_FREVERSEBIND   TYPEFLAGS = 8192
	TYPEFLAG_FPROXY         TYPEFLAGS = 16384
)

// enum
type FUNCFLAGS int32

const (
	FUNCFLAG_FRESTRICTED       FUNCFLAGS = 1
	FUNCFLAG_FSOURCE           FUNCFLAGS = 2
	FUNCFLAG_FBINDABLE         FUNCFLAGS = 4
	FUNCFLAG_FREQUESTEDIT      FUNCFLAGS = 8
	FUNCFLAG_FDISPLAYBIND      FUNCFLAGS = 16
	FUNCFLAG_FDEFAULTBIND      FUNCFLAGS = 32
	FUNCFLAG_FHIDDEN           FUNCFLAGS = 64
	FUNCFLAG_FUSESGETLASTERROR FUNCFLAGS = 128
	FUNCFLAG_FDEFAULTCOLLELEM  FUNCFLAGS = 256
	FUNCFLAG_FUIDEFAULT        FUNCFLAGS = 512
	FUNCFLAG_FNONBROWSABLE     FUNCFLAGS = 1024
	FUNCFLAG_FREPLACEABLE      FUNCFLAGS = 2048
	FUNCFLAG_FIMMEDIATEBIND    FUNCFLAGS = 4096
)

// enum
type VARFLAGS int32

const (
	VARFLAG_FREADONLY        VARFLAGS = 1
	VARFLAG_FSOURCE          VARFLAGS = 2
	VARFLAG_FBINDABLE        VARFLAGS = 4
	VARFLAG_FREQUESTEDIT     VARFLAGS = 8
	VARFLAG_FDISPLAYBIND     VARFLAGS = 16
	VARFLAG_FDEFAULTBIND     VARFLAGS = 32
	VARFLAG_FHIDDEN          VARFLAGS = 64
	VARFLAG_FRESTRICTED      VARFLAGS = 128
	VARFLAG_FDEFAULTCOLLELEM VARFLAGS = 256
	VARFLAG_FUIDEFAULT       VARFLAGS = 512
	VARFLAG_FNONBROWSABLE    VARFLAGS = 1024
	VARFLAG_FREPLACEABLE     VARFLAGS = 2048
	VARFLAG_FIMMEDIATEBIND   VARFLAGS = 4096
)

// enum
type LIBFLAGS int32

const (
	LIBFLAG_FRESTRICTED   LIBFLAGS = 1
	LIBFLAG_FCONTROL      LIBFLAGS = 2
	LIBFLAG_FHIDDEN       LIBFLAGS = 4
	LIBFLAG_FHASDISKIMAGE LIBFLAGS = 8
)

// enum
type CHANGEKIND int32

const (
	CHANGEKIND_ADDMEMBER        CHANGEKIND = 0
	CHANGEKIND_DELETEMEMBER     CHANGEKIND = 1
	CHANGEKIND_SETNAMES         CHANGEKIND = 2
	CHANGEKIND_SETDOCUMENTATION CHANGEKIND = 3
	CHANGEKIND_GENERAL          CHANGEKIND = 4
	CHANGEKIND_INVALIDATE       CHANGEKIND = 5
	CHANGEKIND_CHANGEFAILED     CHANGEKIND = 6
	CHANGEKIND_MAX              CHANGEKIND = 7
)

// enum
type DISCARDCACHE int32

const (
	DISCARDCACHE_SAVEIFDIRTY DISCARDCACHE = 0
	DISCARDCACHE_NOSAVE      DISCARDCACHE = 1
)

// enum
type OLEGETMONIKER int32

const (
	OLEGETMONIKER_ONLYIFTHERE OLEGETMONIKER = 1
	OLEGETMONIKER_FORCEASSIGN OLEGETMONIKER = 2
	OLEGETMONIKER_UNASSIGN    OLEGETMONIKER = 3
	OLEGETMONIKER_TEMPFORUSER OLEGETMONIKER = 4
)

// enum
type OLEWHICHMK int32

const (
	OLEWHICHMK_CONTAINER OLEWHICHMK = 1
	OLEWHICHMK_OBJREL    OLEWHICHMK = 2
	OLEWHICHMK_OBJFULL   OLEWHICHMK = 3
)

// enum
type USERCLASSTYPE int32

const (
	USERCLASSTYPE_FULL    USERCLASSTYPE = 1
	USERCLASSTYPE_SHORT   USERCLASSTYPE = 2
	USERCLASSTYPE_APPNAME USERCLASSTYPE = 3
)

// enum
type OLEMISC int32

const (
	OLEMISC_RECOMPOSEONRESIZE            OLEMISC = 1
	OLEMISC_ONLYICONIC                   OLEMISC = 2
	OLEMISC_INSERTNOTREPLACE             OLEMISC = 4
	OLEMISC_STATIC                       OLEMISC = 8
	OLEMISC_CANTLINKINSIDE               OLEMISC = 16
	OLEMISC_CANLINKBYOLE1                OLEMISC = 32
	OLEMISC_ISLINKOBJECT                 OLEMISC = 64
	OLEMISC_INSIDEOUT                    OLEMISC = 128
	OLEMISC_ACTIVATEWHENVISIBLE          OLEMISC = 256
	OLEMISC_RENDERINGISDEVICEINDEPENDENT OLEMISC = 512
	OLEMISC_INVISIBLEATRUNTIME           OLEMISC = 1024
	OLEMISC_ALWAYSRUN                    OLEMISC = 2048
	OLEMISC_ACTSLIKEBUTTON               OLEMISC = 4096
	OLEMISC_ACTSLIKELABEL                OLEMISC = 8192
	OLEMISC_NOUIACTIVATE                 OLEMISC = 16384
	OLEMISC_ALIGNABLE                    OLEMISC = 32768
	OLEMISC_SIMPLEFRAME                  OLEMISC = 65536
	OLEMISC_SETCLIENTSITEFIRST           OLEMISC = 131072
	OLEMISC_IMEMODE                      OLEMISC = 262144
	OLEMISC_IGNOREACTIVATEWHENVISIBLE    OLEMISC = 524288
	OLEMISC_WANTSTOMENUMERGE             OLEMISC = 1048576
	OLEMISC_SUPPORTSMULTILEVELUNDO       OLEMISC = 2097152
)

// enum
type OLECLOSE int32

const (
	OLECLOSE_SAVEIFDIRTY OLECLOSE = 0
	OLECLOSE_NOSAVE      OLECLOSE = 1
	OLECLOSE_PROMPTSAVE  OLECLOSE = 2
)

// enum
type OLERENDER int32

const (
	OLERENDER_NONE   OLERENDER = 0
	OLERENDER_DRAW   OLERENDER = 1
	OLERENDER_FORMAT OLERENDER = 2
	OLERENDER_ASIS   OLERENDER = 3
)

// enum
type OLEUPDATE int32

const (
	OLEUPDATE_ALWAYS OLEUPDATE = 1
	OLEUPDATE_ONCALL OLEUPDATE = 3
)

// enum
type OLELINKBIND int32

const (
	OLELINKBIND_EVENIFCLASSDIFF OLELINKBIND = 1
)

// enum
type BINDSPEED int32

const (
	BINDSPEED_INDEFINITE BINDSPEED = 1
	BINDSPEED_MODERATE   BINDSPEED = 2
	BINDSPEED_IMMEDIATE  BINDSPEED = 3
)

// enum
type OLECONTF int32

const (
	OLECONTF_EMBEDDINGS    OLECONTF = 1
	OLECONTF_LINKS         OLECONTF = 2
	OLECONTF_OTHERS        OLECONTF = 4
	OLECONTF_ONLYUSER      OLECONTF = 8
	OLECONTF_ONLYIFRUNNING OLECONTF = 16
)

// enum
type OLEVERBATTRIB int32

const (
	OLEVERBATTRIB_NEVERDIRTIES    OLEVERBATTRIB = 1
	OLEVERBATTRIB_ONCONTAINERMENU OLEVERBATTRIB = 2
)

// enum
type REGKIND int32

const (
	REGKIND_DEFAULT  REGKIND = 0
	REGKIND_REGISTER REGKIND = 1
	REGKIND_NONE     REGKIND = 2
)

// enum
type UASFLAGS int32

const (
	UAS_NORMAL         UASFLAGS = 0
	UAS_BLOCKED        UASFLAGS = 1
	UAS_NOPARENTENABLE UASFLAGS = 2
	UAS_MASK           UASFLAGS = 3
)

// enum
type READYSTATE int32

const (
	READYSTATE_UNINITIALIZED READYSTATE = 0
	READYSTATE_LOADING       READYSTATE = 1
	READYSTATE_LOADED        READYSTATE = 2
	READYSTATE_INTERACTIVE   READYSTATE = 3
	READYSTATE_COMPLETE      READYSTATE = 4
)

// enum
type GUIDKIND int32

const (
	GUIDKIND_DEFAULT_SOURCE_DISP_IID GUIDKIND = 1
)

// enum
type CTRLINFO int32

const (
	CTRLINFO_EATS_RETURN CTRLINFO = 1
	CTRLINFO_EATS_ESCAPE CTRLINFO = 2
)

// enum
type XFORMCOORDS int32

const (
	XFORMCOORDS_POSITION            XFORMCOORDS = 1
	XFORMCOORDS_SIZE                XFORMCOORDS = 2
	XFORMCOORDS_HIMETRICTOCONTAINER XFORMCOORDS = 4
	XFORMCOORDS_CONTAINERTOHIMETRIC XFORMCOORDS = 8
	XFORMCOORDS_EVENTCOMPAT         XFORMCOORDS = 16
)

// enum
type PROPPAGESTATUS int32

const (
	PROPPAGESTATUS_DIRTY    PROPPAGESTATUS = 1
	PROPPAGESTATUS_VALIDATE PROPPAGESTATUS = 2
	PROPPAGESTATUS_CLEAN    PROPPAGESTATUS = 4
)

// enum
type PictureAttributes int32

const (
	PICTURE_SCALABLE    PictureAttributes = 1
	PICTURE_TRANSPARENT PictureAttributes = 2
)

// enum
type ACTIVATEFLAGS int32

const (
	ACTIVATE_WINDOWLESS ACTIVATEFLAGS = 1
)

// enum
type OLEDCFLAGS int32

const (
	OLEDC_NODRAW     OLEDCFLAGS = 1
	OLEDC_PAINTBKGND OLEDCFLAGS = 2
	OLEDC_OFFSCREEN  OLEDCFLAGS = 4
)

// enum
type VIEWSTATUS int32

const (
	VIEWSTATUS_OPAQUE              VIEWSTATUS = 1
	VIEWSTATUS_SOLIDBKGND          VIEWSTATUS = 2
	VIEWSTATUS_DVASPECTOPAQUE      VIEWSTATUS = 4
	VIEWSTATUS_DVASPECTTRANSPARENT VIEWSTATUS = 8
	VIEWSTATUS_SURFACE             VIEWSTATUS = 16
	VIEWSTATUS_3DSURFACE           VIEWSTATUS = 32
)

// enum
type HITRESULT int32

const (
	HITRESULT_OUTSIDE     HITRESULT = 0
	HITRESULT_TRANSPARENT HITRESULT = 1
	HITRESULT_CLOSE       HITRESULT = 2
	HITRESULT_HIT         HITRESULT = 3
)

// enum
type DVASPECT2 int32

const (
	DVASPECT_OPAQUE      DVASPECT2 = 16
	DVASPECT_TRANSPARENT DVASPECT2 = 32
)

// enum
type ExtentMode int32

const (
	DVEXTENT_CONTENT  ExtentMode = 0
	DVEXTENT_INTEGRAL ExtentMode = 1
)

// enum
type AspectInfoFlag int32

const (
	DVASPECTINFOFLAG_CANOPTIMIZE AspectInfoFlag = 1
)

// enum
type POINTERINACTIVE int32

const (
	POINTERINACTIVE_ACTIVATEONENTRY   POINTERINACTIVE = 1
	POINTERINACTIVE_DEACTIVATEONLEAVE POINTERINACTIVE = 2
	POINTERINACTIVE_ACTIVATEONDRAG    POINTERINACTIVE = 4
)

// enum
type PROPBAG2_TYPE int32

const (
	PROPBAG2_TYPE_UNDEFINED PROPBAG2_TYPE = 0
	PROPBAG2_TYPE_DATA      PROPBAG2_TYPE = 1
	PROPBAG2_TYPE_URL       PROPBAG2_TYPE = 2
	PROPBAG2_TYPE_OBJECT    PROPBAG2_TYPE = 3
	PROPBAG2_TYPE_STREAM    PROPBAG2_TYPE = 4
	PROPBAG2_TYPE_STORAGE   PROPBAG2_TYPE = 5
	PROPBAG2_TYPE_MONIKER   PROPBAG2_TYPE = 6
)

// enum
type QACONTAINERFLAGS int32

const (
	QACONTAINER_SHOWHATCHING      QACONTAINERFLAGS = 1
	QACONTAINER_SHOWGRABHANDLES   QACONTAINERFLAGS = 2
	QACONTAINER_USERMODE          QACONTAINERFLAGS = 4
	QACONTAINER_DISPLAYASDEFAULT  QACONTAINERFLAGS = 8
	QACONTAINER_UIDEAD            QACONTAINERFLAGS = 16
	QACONTAINER_AUTOCLIP          QACONTAINERFLAGS = 32
	QACONTAINER_MESSAGEREFLECT    QACONTAINERFLAGS = 64
	QACONTAINER_SUPPORTSMNEMONICS QACONTAINERFLAGS = 128
)

// enum
type OLE_TRISTATE int32

const (
	TriUnchecked OLE_TRISTATE = 0
	TriChecked   OLE_TRISTATE = 1
	TriGray      OLE_TRISTATE = 2
)

// enum
type DOCMISC int32

const (
	DOCMISC_CANCREATEMULTIPLEVIEWS   DOCMISC = 1
	DOCMISC_SUPPORTCOMPLEXRECTANGLES DOCMISC = 2
	DOCMISC_CANTOPENEDIT             DOCMISC = 4
	DOCMISC_NOFILESUPPORT            DOCMISC = 8
)

// enum
// flags
type PRINTFLAG uint32

const (
	PRINTFLAG_MAYBOTHERUSER        PRINTFLAG = 1
	PRINTFLAG_PROMPTUSER           PRINTFLAG = 2
	PRINTFLAG_USERMAYCHANGEPRINTER PRINTFLAG = 4
	PRINTFLAG_RECOMPOSETODEVICE    PRINTFLAG = 8
	PRINTFLAG_DONTACTUALLYPRINT    PRINTFLAG = 16
	PRINTFLAG_FORCEPROPERTIES      PRINTFLAG = 32
	PRINTFLAG_PRINTTOFILE          PRINTFLAG = 64
)

// enum
type OLECMDF int32

const (
	OLECMDF_SUPPORTED         OLECMDF = 1
	OLECMDF_ENABLED           OLECMDF = 2
	OLECMDF_LATCHED           OLECMDF = 4
	OLECMDF_NINCHED           OLECMDF = 8
	OLECMDF_INVISIBLE         OLECMDF = 16
	OLECMDF_DEFHIDEONCTXTMENU OLECMDF = 32
)

// enum
type OLECMDTEXTF int32

const (
	OLECMDTEXTF_NONE   OLECMDTEXTF = 0
	OLECMDTEXTF_NAME   OLECMDTEXTF = 1
	OLECMDTEXTF_STATUS OLECMDTEXTF = 2
)

// enum
type OLECMDEXECOPT int32

const (
	OLECMDEXECOPT_DODEFAULT      OLECMDEXECOPT = 0
	OLECMDEXECOPT_PROMPTUSER     OLECMDEXECOPT = 1
	OLECMDEXECOPT_DONTPROMPTUSER OLECMDEXECOPT = 2
	OLECMDEXECOPT_SHOWHELP       OLECMDEXECOPT = 3
)

// enum
type OLECMDID int32

const (
	OLECMDID_OPEN                           OLECMDID = 1
	OLECMDID_NEW                            OLECMDID = 2
	OLECMDID_SAVE                           OLECMDID = 3
	OLECMDID_SAVEAS                         OLECMDID = 4
	OLECMDID_SAVECOPYAS                     OLECMDID = 5
	OLECMDID_PRINT                          OLECMDID = 6
	OLECMDID_PRINTPREVIEW                   OLECMDID = 7
	OLECMDID_PAGESETUP                      OLECMDID = 8
	OLECMDID_SPELL                          OLECMDID = 9
	OLECMDID_PROPERTIES                     OLECMDID = 10
	OLECMDID_CUT                            OLECMDID = 11
	OLECMDID_COPY                           OLECMDID = 12
	OLECMDID_PASTE                          OLECMDID = 13
	OLECMDID_PASTESPECIAL                   OLECMDID = 14
	OLECMDID_UNDO                           OLECMDID = 15
	OLECMDID_REDO                           OLECMDID = 16
	OLECMDID_SELECTALL                      OLECMDID = 17
	OLECMDID_CLEARSELECTION                 OLECMDID = 18
	OLECMDID_ZOOM                           OLECMDID = 19
	OLECMDID_GETZOOMRANGE                   OLECMDID = 20
	OLECMDID_UPDATECOMMANDS                 OLECMDID = 21
	OLECMDID_REFRESH                        OLECMDID = 22
	OLECMDID_STOP                           OLECMDID = 23
	OLECMDID_HIDETOOLBARS                   OLECMDID = 24
	OLECMDID_SETPROGRESSMAX                 OLECMDID = 25
	OLECMDID_SETPROGRESSPOS                 OLECMDID = 26
	OLECMDID_SETPROGRESSTEXT                OLECMDID = 27
	OLECMDID_SETTITLE                       OLECMDID = 28
	OLECMDID_SETDOWNLOADSTATE               OLECMDID = 29
	OLECMDID_STOPDOWNLOAD                   OLECMDID = 30
	OLECMDID_ONTOOLBARACTIVATED             OLECMDID = 31
	OLECMDID_FIND                           OLECMDID = 32
	OLECMDID_DELETE                         OLECMDID = 33
	OLECMDID_HTTPEQUIV                      OLECMDID = 34
	OLECMDID_HTTPEQUIV_DONE                 OLECMDID = 35
	OLECMDID_ENABLE_INTERACTION             OLECMDID = 36
	OLECMDID_ONUNLOAD                       OLECMDID = 37
	OLECMDID_PROPERTYBAG2                   OLECMDID = 38
	OLECMDID_PREREFRESH                     OLECMDID = 39
	OLECMDID_SHOWSCRIPTERROR                OLECMDID = 40
	OLECMDID_SHOWMESSAGE                    OLECMDID = 41
	OLECMDID_SHOWFIND                       OLECMDID = 42
	OLECMDID_SHOWPAGESETUP                  OLECMDID = 43
	OLECMDID_SHOWPRINT                      OLECMDID = 44
	OLECMDID_CLOSE                          OLECMDID = 45
	OLECMDID_ALLOWUILESSSAVEAS              OLECMDID = 46
	OLECMDID_DONTDOWNLOADCSS                OLECMDID = 47
	OLECMDID_UPDATEPAGESTATUS               OLECMDID = 48
	OLECMDID_PRINT2                         OLECMDID = 49
	OLECMDID_PRINTPREVIEW2                  OLECMDID = 50
	OLECMDID_SETPRINTTEMPLATE               OLECMDID = 51
	OLECMDID_GETPRINTTEMPLATE               OLECMDID = 52
	OLECMDID_PAGEACTIONBLOCKED              OLECMDID = 55
	OLECMDID_PAGEACTIONUIQUERY              OLECMDID = 56
	OLECMDID_FOCUSVIEWCONTROLS              OLECMDID = 57
	OLECMDID_FOCUSVIEWCONTROLSQUERY         OLECMDID = 58
	OLECMDID_SHOWPAGEACTIONMENU             OLECMDID = 59
	OLECMDID_ADDTRAVELENTRY                 OLECMDID = 60
	OLECMDID_UPDATETRAVELENTRY              OLECMDID = 61
	OLECMDID_UPDATEBACKFORWARDSTATE         OLECMDID = 62
	OLECMDID_OPTICAL_ZOOM                   OLECMDID = 63
	OLECMDID_OPTICAL_GETZOOMRANGE           OLECMDID = 64
	OLECMDID_WINDOWSTATECHANGED             OLECMDID = 65
	OLECMDID_ACTIVEXINSTALLSCOPE            OLECMDID = 66
	OLECMDID_UPDATETRAVELENTRY_DATARECOVERY OLECMDID = 67
	OLECMDID_SHOWTASKDLG                    OLECMDID = 68
	OLECMDID_POPSTATEEVENT                  OLECMDID = 69
	OLECMDID_VIEWPORT_MODE                  OLECMDID = 70
	OLECMDID_LAYOUT_VIEWPORT_WIDTH          OLECMDID = 71
	OLECMDID_VISUAL_VIEWPORT_EXCLUDE_BOTTOM OLECMDID = 72
	OLECMDID_USER_OPTICAL_ZOOM              OLECMDID = 73
	OLECMDID_PAGEAVAILABLE                  OLECMDID = 74
	OLECMDID_GETUSERSCALABLE                OLECMDID = 75
	OLECMDID_UPDATE_CARET                   OLECMDID = 76
	OLECMDID_ENABLE_VISIBILITY              OLECMDID = 77
	OLECMDID_MEDIA_PLAYBACK                 OLECMDID = 78
	OLECMDID_SETFAVICON                     OLECMDID = 79
	OLECMDID_SET_HOST_FULLSCREENMODE        OLECMDID = 80
	OLECMDID_EXITFULLSCREEN                 OLECMDID = 81
	OLECMDID_SCROLLCOMPLETE                 OLECMDID = 82
	OLECMDID_ONBEFOREUNLOAD                 OLECMDID = 83
	OLECMDID_SHOWMESSAGE_BLOCKABLE          OLECMDID = 84
	OLECMDID_SHOWTASKDLG_BLOCKABLE          OLECMDID = 85
)

// enum
type MEDIAPLAYBACK_STATE int32

const (
	MEDIAPLAYBACK_RESUME              MEDIAPLAYBACK_STATE = 0
	MEDIAPLAYBACK_PAUSE               MEDIAPLAYBACK_STATE = 1
	MEDIAPLAYBACK_PAUSE_AND_SUSPEND   MEDIAPLAYBACK_STATE = 2
	MEDIAPLAYBACK_RESUME_FROM_SUSPEND MEDIAPLAYBACK_STATE = 3
)

// enum
type IGNOREMIME int32

const (
	IGNOREMIME_PROMPT IGNOREMIME = 1
	IGNOREMIME_TEXT   IGNOREMIME = 2
)

// enum
type WPCSETTING int32

const (
	WPCSETTING_LOGGING_ENABLED      WPCSETTING = 1
	WPCSETTING_FILEDOWNLOAD_BLOCKED WPCSETTING = 2
)

// enum
type OLECMDID_REFRESHFLAG int32

const (
	OLECMDIDF_REFRESH_NORMAL                              OLECMDID_REFRESHFLAG = 0
	OLECMDIDF_REFRESH_IFEXPIRED                           OLECMDID_REFRESHFLAG = 1
	OLECMDIDF_REFRESH_CONTINUE                            OLECMDID_REFRESHFLAG = 2
	OLECMDIDF_REFRESH_COMPLETELY                          OLECMDID_REFRESHFLAG = 3
	OLECMDIDF_REFRESH_NO_CACHE                            OLECMDID_REFRESHFLAG = 4
	OLECMDIDF_REFRESH_RELOAD                              OLECMDID_REFRESHFLAG = 5
	OLECMDIDF_REFRESH_LEVELMASK                           OLECMDID_REFRESHFLAG = 255
	OLECMDIDF_REFRESH_CLEARUSERINPUT                      OLECMDID_REFRESHFLAG = 4096
	OLECMDIDF_REFRESH_PROMPTIFOFFLINE                     OLECMDID_REFRESHFLAG = 8192
	OLECMDIDF_REFRESH_THROUGHSCRIPT                       OLECMDID_REFRESHFLAG = 16384
	OLECMDIDF_REFRESH_SKIPBEFOREUNLOADEVENT               OLECMDID_REFRESHFLAG = 32768
	OLECMDIDF_REFRESH_PAGEACTION_ACTIVEXINSTALL           OLECMDID_REFRESHFLAG = 65536
	OLECMDIDF_REFRESH_PAGEACTION_FILEDOWNLOAD             OLECMDID_REFRESHFLAG = 131072
	OLECMDIDF_REFRESH_PAGEACTION_LOCALMACHINE             OLECMDID_REFRESHFLAG = 262144
	OLECMDIDF_REFRESH_PAGEACTION_POPUPWINDOW              OLECMDID_REFRESHFLAG = 524288
	OLECMDIDF_REFRESH_PAGEACTION_PROTLOCKDOWNLOCALMACHINE OLECMDID_REFRESHFLAG = 1048576
	OLECMDIDF_REFRESH_PAGEACTION_PROTLOCKDOWNTRUSTED      OLECMDID_REFRESHFLAG = 2097152
	OLECMDIDF_REFRESH_PAGEACTION_PROTLOCKDOWNINTRANET     OLECMDID_REFRESHFLAG = 4194304
	OLECMDIDF_REFRESH_PAGEACTION_PROTLOCKDOWNINTERNET     OLECMDID_REFRESHFLAG = 8388608
	OLECMDIDF_REFRESH_PAGEACTION_PROTLOCKDOWNRESTRICTED   OLECMDID_REFRESHFLAG = 16777216
	OLECMDIDF_REFRESH_PAGEACTION_MIXEDCONTENT             OLECMDID_REFRESHFLAG = 33554432
	OLECMDIDF_REFRESH_PAGEACTION_INVALID_CERT             OLECMDID_REFRESHFLAG = 67108864
	OLECMDIDF_REFRESH_PAGEACTION_ALLOW_VERSION            OLECMDID_REFRESHFLAG = 134217728
)

// enum
type OLECMDID_PAGEACTIONFLAG int32

const (
	OLECMDIDF_PAGEACTION_FILEDOWNLOAD                       OLECMDID_PAGEACTIONFLAG = 1
	OLECMDIDF_PAGEACTION_ACTIVEXINSTALL                     OLECMDID_PAGEACTIONFLAG = 2
	OLECMDIDF_PAGEACTION_ACTIVEXTRUSTFAIL                   OLECMDID_PAGEACTIONFLAG = 4
	OLECMDIDF_PAGEACTION_ACTIVEXUSERDISABLE                 OLECMDID_PAGEACTIONFLAG = 8
	OLECMDIDF_PAGEACTION_ACTIVEXDISALLOW                    OLECMDID_PAGEACTIONFLAG = 16
	OLECMDIDF_PAGEACTION_ACTIVEXUNSAFE                      OLECMDID_PAGEACTIONFLAG = 32
	OLECMDIDF_PAGEACTION_POPUPWINDOW                        OLECMDID_PAGEACTIONFLAG = 64
	OLECMDIDF_PAGEACTION_LOCALMACHINE                       OLECMDID_PAGEACTIONFLAG = 128
	OLECMDIDF_PAGEACTION_MIMETEXTPLAIN                      OLECMDID_PAGEACTIONFLAG = 256
	OLECMDIDF_PAGEACTION_SCRIPTNAVIGATE                     OLECMDID_PAGEACTIONFLAG = 512
	OLECMDIDF_PAGEACTION_SCRIPTNAVIGATE_ACTIVEXINSTALL      OLECMDID_PAGEACTIONFLAG = 512
	OLECMDIDF_PAGEACTION_PROTLOCKDOWNLOCALMACHINE           OLECMDID_PAGEACTIONFLAG = 1024
	OLECMDIDF_PAGEACTION_PROTLOCKDOWNTRUSTED                OLECMDID_PAGEACTIONFLAG = 2048
	OLECMDIDF_PAGEACTION_PROTLOCKDOWNINTRANET               OLECMDID_PAGEACTIONFLAG = 4096
	OLECMDIDF_PAGEACTION_PROTLOCKDOWNINTERNET               OLECMDID_PAGEACTIONFLAG = 8192
	OLECMDIDF_PAGEACTION_PROTLOCKDOWNRESTRICTED             OLECMDID_PAGEACTIONFLAG = 16384
	OLECMDIDF_PAGEACTION_PROTLOCKDOWNDENY                   OLECMDID_PAGEACTIONFLAG = 32768
	OLECMDIDF_PAGEACTION_POPUPALLOWED                       OLECMDID_PAGEACTIONFLAG = 65536
	OLECMDIDF_PAGEACTION_SCRIPTPROMPT                       OLECMDID_PAGEACTIONFLAG = 131072
	OLECMDIDF_PAGEACTION_ACTIVEXUSERAPPROVAL                OLECMDID_PAGEACTIONFLAG = 262144
	OLECMDIDF_PAGEACTION_MIXEDCONTENT                       OLECMDID_PAGEACTIONFLAG = 524288
	OLECMDIDF_PAGEACTION_INVALID_CERT                       OLECMDID_PAGEACTIONFLAG = 1048576
	OLECMDIDF_PAGEACTION_INTRANETZONEREQUEST                OLECMDID_PAGEACTIONFLAG = 2097152
	OLECMDIDF_PAGEACTION_XSSFILTERED                        OLECMDID_PAGEACTIONFLAG = 4194304
	OLECMDIDF_PAGEACTION_SPOOFABLEIDNHOST                   OLECMDID_PAGEACTIONFLAG = 8388608
	OLECMDIDF_PAGEACTION_ACTIVEX_EPM_INCOMPATIBLE           OLECMDID_PAGEACTIONFLAG = 16777216
	OLECMDIDF_PAGEACTION_SCRIPTNAVIGATE_ACTIVEXUSERAPPROVAL OLECMDID_PAGEACTIONFLAG = 33554432
	OLECMDIDF_PAGEACTION_WPCBLOCKED                         OLECMDID_PAGEACTIONFLAG = 67108864
	OLECMDIDF_PAGEACTION_WPCBLOCKED_ACTIVEX                 OLECMDID_PAGEACTIONFLAG = 134217728
	OLECMDIDF_PAGEACTION_EXTENSION_COMPAT_BLOCKED           OLECMDID_PAGEACTIONFLAG = 268435456
	OLECMDIDF_PAGEACTION_NORESETACTIVEX                     OLECMDID_PAGEACTIONFLAG = 536870912
	OLECMDIDF_PAGEACTION_GENERIC_STATE                      OLECMDID_PAGEACTIONFLAG = 1073741824
	OLECMDIDF_PAGEACTION_RESET                              OLECMDID_PAGEACTIONFLAG = -2147483648
)

// enum
type OLECMDID_BROWSERSTATEFLAG int32

const (
	OLECMDIDF_BROWSERSTATE_EXTENSIONSOFF     OLECMDID_BROWSERSTATEFLAG = 1
	OLECMDIDF_BROWSERSTATE_IESECURITY        OLECMDID_BROWSERSTATEFLAG = 2
	OLECMDIDF_BROWSERSTATE_PROTECTEDMODE_OFF OLECMDID_BROWSERSTATEFLAG = 4
	OLECMDIDF_BROWSERSTATE_RESET             OLECMDID_BROWSERSTATEFLAG = 8
	OLECMDIDF_BROWSERSTATE_REQUIRESACTIVEX   OLECMDID_BROWSERSTATEFLAG = 16
	OLECMDIDF_BROWSERSTATE_DESKTOPHTMLDIALOG OLECMDID_BROWSERSTATEFLAG = 32
	OLECMDIDF_BROWSERSTATE_BLOCKEDVERSION    OLECMDID_BROWSERSTATEFLAG = 64
)

// enum
type OLECMDID_OPTICAL_ZOOMFLAG int32

const (
	OLECMDIDF_OPTICAL_ZOOM_NOPERSIST       OLECMDID_OPTICAL_ZOOMFLAG = 1
	OLECMDIDF_OPTICAL_ZOOM_NOLAYOUT        OLECMDID_OPTICAL_ZOOMFLAG = 16
	OLECMDIDF_OPTICAL_ZOOM_NOTRANSIENT     OLECMDID_OPTICAL_ZOOMFLAG = 32
	OLECMDIDF_OPTICAL_ZOOM_RELOADFORNEWTAB OLECMDID_OPTICAL_ZOOMFLAG = 64
)

// enum
type PAGEACTION_UI int32

const (
	PAGEACTION_UI_DEFAULT  PAGEACTION_UI = 0
	PAGEACTION_UI_MODAL    PAGEACTION_UI = 1
	PAGEACTION_UI_MODELESS PAGEACTION_UI = 2
	PAGEACTION_UI_SILENT   PAGEACTION_UI = 3
)

// enum
type OLECMDID_WINDOWSTATE_FLAG int32

const (
	OLECMDIDF_WINDOWSTATE_USERVISIBLE       OLECMDID_WINDOWSTATE_FLAG = 1
	OLECMDIDF_WINDOWSTATE_ENABLED           OLECMDID_WINDOWSTATE_FLAG = 2
	OLECMDIDF_WINDOWSTATE_USERVISIBLE_VALID OLECMDID_WINDOWSTATE_FLAG = 65536
	OLECMDIDF_WINDOWSTATE_ENABLED_VALID     OLECMDID_WINDOWSTATE_FLAG = 131072
)

// enum
type OLECMDID_VIEWPORT_MODE_FLAG int32

const (
	OLECMDIDF_VIEWPORTMODE_FIXED_LAYOUT_WIDTH          OLECMDID_VIEWPORT_MODE_FLAG = 1
	OLECMDIDF_VIEWPORTMODE_EXCLUDE_VISUAL_BOTTOM       OLECMDID_VIEWPORT_MODE_FLAG = 2
	OLECMDIDF_VIEWPORTMODE_FIXED_LAYOUT_WIDTH_VALID    OLECMDID_VIEWPORT_MODE_FLAG = 65536
	OLECMDIDF_VIEWPORTMODE_EXCLUDE_VISUAL_BOTTOM_VALID OLECMDID_VIEWPORT_MODE_FLAG = 131072
)

// enum
type OLEUIPASTEFLAG int32

const (
	OLEUIPASTE_ENABLEICON  OLEUIPASTEFLAG = 2048
	OLEUIPASTE_PASTEONLY   OLEUIPASTEFLAG = 0
	OLEUIPASTE_PASTE       OLEUIPASTEFLAG = 512
	OLEUIPASTE_LINKANYTYPE OLEUIPASTEFLAG = 1024
	OLEUIPASTE_LINKTYPE1   OLEUIPASTEFLAG = 1
	OLEUIPASTE_LINKTYPE2   OLEUIPASTEFLAG = 2
	OLEUIPASTE_LINKTYPE3   OLEUIPASTEFLAG = 4
	OLEUIPASTE_LINKTYPE4   OLEUIPASTEFLAG = 8
	OLEUIPASTE_LINKTYPE5   OLEUIPASTEFLAG = 16
	OLEUIPASTE_LINKTYPE6   OLEUIPASTEFLAG = 32
	OLEUIPASTE_LINKTYPE7   OLEUIPASTEFLAG = 64
	OLEUIPASTE_LINKTYPE8   OLEUIPASTEFLAG = 128
)

// structs

type WireSAFEARR_BSTR_ struct {
	Size  uint32
	ABstr **FLAGGED_WORD_BLOB
}

type WireSAFEARR_UNKNOWN_ struct {
	Size      uint32
	ApUnknown **IUnknown
}

type WireSAFEARR_DISPATCH_ struct {
	Size       uint32
	ApDispatch **IDispatch
}

type WireSAFEARR_VARIANT_ struct {
	Size     uint32
	AVariant **WireVARIANT_
}

type WireSAFEARR_BRECORD_ struct {
	Size    uint32
	ARecord **WireBRECORD_
}

type WireSAFEARR_HAVEIID_ struct {
	Size      uint32
	ApUnknown **IUnknown
	Iid       syscall.GUID
}

type WireSAFEARRAY_UNION_U_ struct {
	Data [4]uint64
}

func (this *WireSAFEARRAY_UNION_U_) BstrStr() *WireSAFEARR_BSTR_ {
	return (*WireSAFEARR_BSTR_)(unsafe.Pointer(this))
}

func (this *WireSAFEARRAY_UNION_U_) BstrStrVal() WireSAFEARR_BSTR_ {
	return *(*WireSAFEARR_BSTR_)(unsafe.Pointer(this))
}

func (this *WireSAFEARRAY_UNION_U_) UnknownStr() *WireSAFEARR_UNKNOWN_ {
	return (*WireSAFEARR_UNKNOWN_)(unsafe.Pointer(this))
}

func (this *WireSAFEARRAY_UNION_U_) UnknownStrVal() WireSAFEARR_UNKNOWN_ {
	return *(*WireSAFEARR_UNKNOWN_)(unsafe.Pointer(this))
}

func (this *WireSAFEARRAY_UNION_U_) DispatchStr() *WireSAFEARR_DISPATCH_ {
	return (*WireSAFEARR_DISPATCH_)(unsafe.Pointer(this))
}

func (this *WireSAFEARRAY_UNION_U_) DispatchStrVal() WireSAFEARR_DISPATCH_ {
	return *(*WireSAFEARR_DISPATCH_)(unsafe.Pointer(this))
}

func (this *WireSAFEARRAY_UNION_U_) VariantStr() *WireSAFEARR_VARIANT_ {
	return (*WireSAFEARR_VARIANT_)(unsafe.Pointer(this))
}

func (this *WireSAFEARRAY_UNION_U_) VariantStrVal() WireSAFEARR_VARIANT_ {
	return *(*WireSAFEARR_VARIANT_)(unsafe.Pointer(this))
}

func (this *WireSAFEARRAY_UNION_U_) RecordStr() *WireSAFEARR_BRECORD_ {
	return (*WireSAFEARR_BRECORD_)(unsafe.Pointer(this))
}

func (this *WireSAFEARRAY_UNION_U_) RecordStrVal() WireSAFEARR_BRECORD_ {
	return *(*WireSAFEARR_BRECORD_)(unsafe.Pointer(this))
}

func (this *WireSAFEARRAY_UNION_U_) HaveIidStr() *WireSAFEARR_HAVEIID_ {
	return (*WireSAFEARR_HAVEIID_)(unsafe.Pointer(this))
}

func (this *WireSAFEARRAY_UNION_U_) HaveIidStrVal() WireSAFEARR_HAVEIID_ {
	return *(*WireSAFEARR_HAVEIID_)(unsafe.Pointer(this))
}

func (this *WireSAFEARRAY_UNION_U_) ByteStr() *BYTE_SIZEDARR {
	return (*BYTE_SIZEDARR)(unsafe.Pointer(this))
}

func (this *WireSAFEARRAY_UNION_U_) ByteStrVal() BYTE_SIZEDARR {
	return *(*BYTE_SIZEDARR)(unsafe.Pointer(this))
}

func (this *WireSAFEARRAY_UNION_U_) WordStr() *SHORT_SIZEDARR {
	return (*SHORT_SIZEDARR)(unsafe.Pointer(this))
}

func (this *WireSAFEARRAY_UNION_U_) WordStrVal() SHORT_SIZEDARR {
	return *(*SHORT_SIZEDARR)(unsafe.Pointer(this))
}

func (this *WireSAFEARRAY_UNION_U_) LongStr() *LONG_SIZEDARR {
	return (*LONG_SIZEDARR)(unsafe.Pointer(this))
}

func (this *WireSAFEARRAY_UNION_U_) LongStrVal() LONG_SIZEDARR {
	return *(*LONG_SIZEDARR)(unsafe.Pointer(this))
}

func (this *WireSAFEARRAY_UNION_U_) HyperStr() *HYPER_SIZEDARR {
	return (*HYPER_SIZEDARR)(unsafe.Pointer(this))
}

func (this *WireSAFEARRAY_UNION_U_) HyperStrVal() HYPER_SIZEDARR {
	return *(*HYPER_SIZEDARR)(unsafe.Pointer(this))
}

type WireSAFEARRAY_UNION_ struct {
	SfType uint32
	U      WireSAFEARRAY_UNION_U_
}

type WireSAFEARRAY_ struct {
	CDims         uint16
	FFeatures     uint16
	CbElements    uint32
	CLocks        uint32
	UArrayStructs WireSAFEARRAY_UNION_
	Rgsabound     [1]SAFEARRAYBOUND
}

type WireBRECORD_ struct {
	FFlags   uint32
	ClSize   uint32
	PRecInfo *IRecordInfo
	PRecord  *byte
}

type WireVARIANT_Anonymous_ struct {
	Data [2]uint64
}

func (this *WireVARIANT_Anonymous_) LlVal() *int64 {
	return (*int64)(unsafe.Pointer(this))
}

func (this *WireVARIANT_Anonymous_) LlValVal() int64 {
	return *(*int64)(unsafe.Pointer(this))
}

func (this *WireVARIANT_Anonymous_) LVal() *int32 {
	return (*int32)(unsafe.Pointer(this))
}

func (this *WireVARIANT_Anonymous_) LValVal() int32 {
	return *(*int32)(unsafe.Pointer(this))
}

func (this *WireVARIANT_Anonymous_) BVal() *byte {
	return (*byte)(unsafe.Pointer(this))
}

func (this *WireVARIANT_Anonymous_) BValVal() byte {
	return *(*byte)(unsafe.Pointer(this))
}

func (this *WireVARIANT_Anonymous_) IVal() *int16 {
	return (*int16)(unsafe.Pointer(this))
}

func (this *WireVARIANT_Anonymous_) IValVal() int16 {
	return *(*int16)(unsafe.Pointer(this))
}

func (this *WireVARIANT_Anonymous_) FltVal() *float32 {
	return (*float32)(unsafe.Pointer(this))
}

func (this *WireVARIANT_Anonymous_) FltValVal() float32 {
	return *(*float32)(unsafe.Pointer(this))
}

func (this *WireVARIANT_Anonymous_) DblVal() *float64 {
	return (*float64)(unsafe.Pointer(this))
}

func (this *WireVARIANT_Anonymous_) DblValVal() float64 {
	return *(*float64)(unsafe.Pointer(this))
}

func (this *WireVARIANT_Anonymous_) BoolVal() *int16 {
	return (*int16)(unsafe.Pointer(this))
}

func (this *WireVARIANT_Anonymous_) BoolValVal() int16 {
	return *(*int16)(unsafe.Pointer(this))
}

func (this *WireVARIANT_Anonymous_) Scode() *int32 {
	return (*int32)(unsafe.Pointer(this))
}

func (this *WireVARIANT_Anonymous_) ScodeVal() int32 {
	return *(*int32)(unsafe.Pointer(this))
}

func (this *WireVARIANT_Anonymous_) CyVal() *CY {
	return (*CY)(unsafe.Pointer(this))
}

func (this *WireVARIANT_Anonymous_) CyValVal() CY {
	return *(*CY)(unsafe.Pointer(this))
}

func (this *WireVARIANT_Anonymous_) Date() *float64 {
	return (*float64)(unsafe.Pointer(this))
}

func (this *WireVARIANT_Anonymous_) DateVal() float64 {
	return *(*float64)(unsafe.Pointer(this))
}

func (this *WireVARIANT_Anonymous_) BstrVal() **FLAGGED_WORD_BLOB {
	return (**FLAGGED_WORD_BLOB)(unsafe.Pointer(this))
}

func (this *WireVARIANT_Anonymous_) BstrValVal() *FLAGGED_WORD_BLOB {
	return *(**FLAGGED_WORD_BLOB)(unsafe.Pointer(this))
}

func (this *WireVARIANT_Anonymous_) PunkVal() **IUnknown {
	return (**IUnknown)(unsafe.Pointer(this))
}

func (this *WireVARIANT_Anonymous_) PunkValVal() *IUnknown {
	return *(**IUnknown)(unsafe.Pointer(this))
}

func (this *WireVARIANT_Anonymous_) PdispVal() **IDispatch {
	return (**IDispatch)(unsafe.Pointer(this))
}

func (this *WireVARIANT_Anonymous_) PdispValVal() *IDispatch {
	return *(**IDispatch)(unsafe.Pointer(this))
}

func (this *WireVARIANT_Anonymous_) Parray() ***WireSAFEARRAY_ {
	return (***WireSAFEARRAY_)(unsafe.Pointer(this))
}

func (this *WireVARIANT_Anonymous_) ParrayVal() **WireSAFEARRAY_ {
	return *(***WireSAFEARRAY_)(unsafe.Pointer(this))
}

func (this *WireVARIANT_Anonymous_) BrecVal() **WireBRECORD_ {
	return (**WireBRECORD_)(unsafe.Pointer(this))
}

func (this *WireVARIANT_Anonymous_) BrecValVal() *WireBRECORD_ {
	return *(**WireBRECORD_)(unsafe.Pointer(this))
}

func (this *WireVARIANT_Anonymous_) PbVal() **byte {
	return (**byte)(unsafe.Pointer(this))
}

func (this *WireVARIANT_Anonymous_) PbValVal() *byte {
	return *(**byte)(unsafe.Pointer(this))
}

func (this *WireVARIANT_Anonymous_) PiVal() **int16 {
	return (**int16)(unsafe.Pointer(this))
}

func (this *WireVARIANT_Anonymous_) PiValVal() *int16 {
	return *(**int16)(unsafe.Pointer(this))
}

func (this *WireVARIANT_Anonymous_) PlVal() **int32 {
	return (**int32)(unsafe.Pointer(this))
}

func (this *WireVARIANT_Anonymous_) PlValVal() *int32 {
	return *(**int32)(unsafe.Pointer(this))
}

func (this *WireVARIANT_Anonymous_) PllVal() **int64 {
	return (**int64)(unsafe.Pointer(this))
}

func (this *WireVARIANT_Anonymous_) PllValVal() *int64 {
	return *(**int64)(unsafe.Pointer(this))
}

func (this *WireVARIANT_Anonymous_) PfltVal() **float32 {
	return (**float32)(unsafe.Pointer(this))
}

func (this *WireVARIANT_Anonymous_) PfltValVal() *float32 {
	return *(**float32)(unsafe.Pointer(this))
}

func (this *WireVARIANT_Anonymous_) PdblVal() **float64 {
	return (**float64)(unsafe.Pointer(this))
}

func (this *WireVARIANT_Anonymous_) PdblValVal() *float64 {
	return *(**float64)(unsafe.Pointer(this))
}

func (this *WireVARIANT_Anonymous_) PboolVal() **int16 {
	return (**int16)(unsafe.Pointer(this))
}

func (this *WireVARIANT_Anonymous_) PboolValVal() *int16 {
	return *(**int16)(unsafe.Pointer(this))
}

func (this *WireVARIANT_Anonymous_) Pscode() **int32 {
	return (**int32)(unsafe.Pointer(this))
}

func (this *WireVARIANT_Anonymous_) PscodeVal() *int32 {
	return *(**int32)(unsafe.Pointer(this))
}

func (this *WireVARIANT_Anonymous_) PcyVal() **CY {
	return (**CY)(unsafe.Pointer(this))
}

func (this *WireVARIANT_Anonymous_) PcyValVal() *CY {
	return *(**CY)(unsafe.Pointer(this))
}

func (this *WireVARIANT_Anonymous_) Pdate() **float64 {
	return (**float64)(unsafe.Pointer(this))
}

func (this *WireVARIANT_Anonymous_) PdateVal() *float64 {
	return *(**float64)(unsafe.Pointer(this))
}

func (this *WireVARIANT_Anonymous_) PbstrVal() ***FLAGGED_WORD_BLOB {
	return (***FLAGGED_WORD_BLOB)(unsafe.Pointer(this))
}

func (this *WireVARIANT_Anonymous_) PbstrValVal() **FLAGGED_WORD_BLOB {
	return *(***FLAGGED_WORD_BLOB)(unsafe.Pointer(this))
}

func (this *WireVARIANT_Anonymous_) PpunkVal() ***IUnknown {
	return (***IUnknown)(unsafe.Pointer(this))
}

func (this *WireVARIANT_Anonymous_) PpunkValVal() **IUnknown {
	return *(***IUnknown)(unsafe.Pointer(this))
}

func (this *WireVARIANT_Anonymous_) PpdispVal() ***IDispatch {
	return (***IDispatch)(unsafe.Pointer(this))
}

func (this *WireVARIANT_Anonymous_) PpdispValVal() **IDispatch {
	return *(***IDispatch)(unsafe.Pointer(this))
}

func (this *WireVARIANT_Anonymous_) Pparray() ****WireSAFEARRAY_ {
	return (****WireSAFEARRAY_)(unsafe.Pointer(this))
}

func (this *WireVARIANT_Anonymous_) PparrayVal() ***WireSAFEARRAY_ {
	return *(****WireSAFEARRAY_)(unsafe.Pointer(this))
}

func (this *WireVARIANT_Anonymous_) PvarVal() ***WireVARIANT_ {
	return (***WireVARIANT_)(unsafe.Pointer(this))
}

func (this *WireVARIANT_Anonymous_) PvarValVal() **WireVARIANT_ {
	return *(***WireVARIANT_)(unsafe.Pointer(this))
}

func (this *WireVARIANT_Anonymous_) CVal() *CHAR {
	return (*CHAR)(unsafe.Pointer(this))
}

func (this *WireVARIANT_Anonymous_) CValVal() CHAR {
	return *(*CHAR)(unsafe.Pointer(this))
}

func (this *WireVARIANT_Anonymous_) UiVal() *uint16 {
	return (*uint16)(unsafe.Pointer(this))
}

func (this *WireVARIANT_Anonymous_) UiValVal() uint16 {
	return *(*uint16)(unsafe.Pointer(this))
}

func (this *WireVARIANT_Anonymous_) UlVal() *uint32 {
	return (*uint32)(unsafe.Pointer(this))
}

func (this *WireVARIANT_Anonymous_) UlValVal() uint32 {
	return *(*uint32)(unsafe.Pointer(this))
}

func (this *WireVARIANT_Anonymous_) UllVal() *uint64 {
	return (*uint64)(unsafe.Pointer(this))
}

func (this *WireVARIANT_Anonymous_) UllValVal() uint64 {
	return *(*uint64)(unsafe.Pointer(this))
}

func (this *WireVARIANT_Anonymous_) IntVal() *int32 {
	return (*int32)(unsafe.Pointer(this))
}

func (this *WireVARIANT_Anonymous_) IntValVal() int32 {
	return *(*int32)(unsafe.Pointer(this))
}

func (this *WireVARIANT_Anonymous_) UintVal() *uint32 {
	return (*uint32)(unsafe.Pointer(this))
}

func (this *WireVARIANT_Anonymous_) UintValVal() uint32 {
	return *(*uint32)(unsafe.Pointer(this))
}

func (this *WireVARIANT_Anonymous_) DecVal() *DECIMAL {
	return (*DECIMAL)(unsafe.Pointer(this))
}

func (this *WireVARIANT_Anonymous_) DecValVal() DECIMAL {
	return *(*DECIMAL)(unsafe.Pointer(this))
}

func (this *WireVARIANT_Anonymous_) PdecVal() **DECIMAL {
	return (**DECIMAL)(unsafe.Pointer(this))
}

func (this *WireVARIANT_Anonymous_) PdecValVal() *DECIMAL {
	return *(**DECIMAL)(unsafe.Pointer(this))
}

func (this *WireVARIANT_Anonymous_) PcVal() *PSTR {
	return (*PSTR)(unsafe.Pointer(this))
}

func (this *WireVARIANT_Anonymous_) PcValVal() PSTR {
	return *(*PSTR)(unsafe.Pointer(this))
}

func (this *WireVARIANT_Anonymous_) PuiVal() **uint16 {
	return (**uint16)(unsafe.Pointer(this))
}

func (this *WireVARIANT_Anonymous_) PuiValVal() *uint16 {
	return *(**uint16)(unsafe.Pointer(this))
}

func (this *WireVARIANT_Anonymous_) PulVal() **uint32 {
	return (**uint32)(unsafe.Pointer(this))
}

func (this *WireVARIANT_Anonymous_) PulValVal() *uint32 {
	return *(**uint32)(unsafe.Pointer(this))
}

func (this *WireVARIANT_Anonymous_) PullVal() **uint64 {
	return (**uint64)(unsafe.Pointer(this))
}

func (this *WireVARIANT_Anonymous_) PullValVal() *uint64 {
	return *(**uint64)(unsafe.Pointer(this))
}

func (this *WireVARIANT_Anonymous_) PintVal() **int32 {
	return (**int32)(unsafe.Pointer(this))
}

func (this *WireVARIANT_Anonymous_) PintValVal() *int32 {
	return *(**int32)(unsafe.Pointer(this))
}

func (this *WireVARIANT_Anonymous_) PuintVal() **uint32 {
	return (**uint32)(unsafe.Pointer(this))
}

func (this *WireVARIANT_Anonymous_) PuintValVal() *uint32 {
	return *(**uint32)(unsafe.Pointer(this))
}

type WireVARIANT_ struct {
	ClSize      uint32
	RpcReserved uint32
	Vt          uint16
	WReserved1  uint16
	WReserved2  uint16
	WReserved3  uint16
	WireVARIANT_Anonymous_
}

type ARRAYDESC struct {
	TdescElem TYPEDESC
	CDims     uint16
	Rgbounds  [1]SAFEARRAYBOUND
}

type PARAMDESCEX struct {
	CBytes          uint32
	VarDefaultValue VARIANT
}

type PARAMDESC struct {
	Pparamdescex *PARAMDESCEX
	WParamFlags  uint16
}

type CLEANLOCALSTORAGE struct {
	PInterface *IUnknown
	PStorage   unsafe.Pointer
	Flags      uint32
}

type OBJECTDESCRIPTOR struct {
	CbSize             uint32
	Clsid              syscall.GUID
	DwDrawAspect       uint32
	Sizel              SIZE
	Pointl             POINTL
	DwStatus           uint32
	DwFullUserTypeName uint32
	DwSrcOfCopy        uint32
}

type OIFI struct {
	Cb            uint32
	FMDIApp       BOOL
	HwndFrame     HWND
	Haccel        HACCEL
	CAccelEntries uint32
}

type OleMenuGroupWidths struct {
	Width [6]int32
}

type OLEVERB struct {
	LVerb        int32
	LpszVerbName PWSTR
	FuFlags      uint32
	GrfAttribs   uint32
}

type NUMPARSE struct {
	CDig       int32
	DwInFlags  uint32
	DwOutFlags uint32
	CchUsed    int32
	NBaseShift int32
	NPwr10     int32
}

type UDATE struct {
	St         SYSTEMTIME
	WDayOfYear uint16
}

type PARAMDATA struct {
	SzName PWSTR
	Vt     uint16
}

type METHODDATA struct {
	SzName   PWSTR
	Ppdata   *PARAMDATA
	Dispid   int32
	IMeth    uint32
	Cc       CALLCONV
	CArgs    uint32
	WFlags   uint16
	VtReturn uint16
}

type INTERFACEDATA struct {
	Pmethdata *METHODDATA
	CMembers  uint32
}

type LICINFO struct {
	CbLicInfo        int32
	FRuntimeKeyAvail BOOL
	FLicVerified     BOOL
}

type CONTROLINFO struct {
	Cb      uint32
	HAccel  HACCEL
	CAccel  uint16
	DwFlags uint32
}

type POINTF struct {
	X float32
	Y float32
}

type PROPPAGEINFO struct {
	Cb            uint32
	PszTitle      PWSTR
	Size          SIZE
	PszDocString  PWSTR
	PszHelpFile   PWSTR
	DwHelpContext uint32
}

type CAUUID struct {
	CElems uint32
	PElems *syscall.GUID
}

type ExtentInfo struct {
	Cb            uint32
	DwExtentMode  uint32
	SizelProposed SIZE
}

type AspectInfo struct {
	Cb      uint32
	DwFlags uint32
}

type CALPOLESTR struct {
	CElems uint32
	PElems *PWSTR
}

type CADWORD struct {
	CElems uint32
	PElems *uint32
}

type QACONTAINER struct {
	CbSize              uint32
	PClientSite         *IOleClientSite
	PAdviseSink         *IAdviseSinkEx
	PPropertyNotifySink *IPropertyNotifySink
	PUnkEventSink       *IUnknown
	DwAmbientFlags      uint32
	ColorFore           uint32
	ColorBack           uint32
	PFont               *IFont
	PUndoMgr            *IOleUndoManager
	DwAppearance        uint32
	Lcid                int32
	Hpal                HPALETTE
	PBindHost           *IBindHost
	POleControlSite     *IOleControlSite
	PServiceProvider    *IServiceProvider
}

type QACONTROL struct {
	CbSize                    uint32
	DwMiscStatus              uint32
	DwViewStatus              uint32
	DwEventCookie             uint32
	DwPropNotifyCookie        uint32
	DwPointerActivationPolicy uint32
}

type OCPFIPARAMS struct {
	CbStructSize          uint32
	HWndOwner             HWND
	X                     int32
	Y                     int32
	LpszCaption           PWSTR
	CObjects              uint32
	LplpUnk               **IUnknown
	CPages                uint32
	LpPages               *syscall.GUID
	Lcid                  uint32
	DispidInitialProperty int32
}

type FONTDESC struct {
	CbSizeofstruct uint32
	LpstrName      PWSTR
	CySize         CY
	SWeight        int16
	SCharset       int16
	FItalic        BOOL
	FUnderline     BOOL
	FStrikethrough BOOL
}

type PICTDESC_Anonymous_Bmp struct {
	Hbitmap HBITMAP
	Hpal    HPALETTE
}

type PICTDESC_Anonymous_Wmf struct {
	Hmeta HMETAFILE
	XExt  int32
	YExt  int32
}

type PICTDESC_Anonymous_Icon struct {
	Hicon HICON
}

type PICTDESC_Anonymous_Emf struct {
	Hemf HENHMETAFILE
}

type PICTDESC_Anonymous struct {
	Data [2]uint64
}

func (this *PICTDESC_Anonymous) Bmp() *PICTDESC_Anonymous_Bmp {
	return (*PICTDESC_Anonymous_Bmp)(unsafe.Pointer(this))
}

func (this *PICTDESC_Anonymous) BmpVal() PICTDESC_Anonymous_Bmp {
	return *(*PICTDESC_Anonymous_Bmp)(unsafe.Pointer(this))
}

func (this *PICTDESC_Anonymous) Wmf() *PICTDESC_Anonymous_Wmf {
	return (*PICTDESC_Anonymous_Wmf)(unsafe.Pointer(this))
}

func (this *PICTDESC_Anonymous) WmfVal() PICTDESC_Anonymous_Wmf {
	return *(*PICTDESC_Anonymous_Wmf)(unsafe.Pointer(this))
}

func (this *PICTDESC_Anonymous) Icon() *PICTDESC_Anonymous_Icon {
	return (*PICTDESC_Anonymous_Icon)(unsafe.Pointer(this))
}

func (this *PICTDESC_Anonymous) IconVal() PICTDESC_Anonymous_Icon {
	return *(*PICTDESC_Anonymous_Icon)(unsafe.Pointer(this))
}

func (this *PICTDESC_Anonymous) Emf() *PICTDESC_Anonymous_Emf {
	return (*PICTDESC_Anonymous_Emf)(unsafe.Pointer(this))
}

func (this *PICTDESC_Anonymous) EmfVal() PICTDESC_Anonymous_Emf {
	return *(*PICTDESC_Anonymous_Emf)(unsafe.Pointer(this))
}

type PICTDESC struct {
	CbSizeofstruct uint32
	PicType        uint32
	PICTDESC_Anonymous
}

type PAGERANGE struct {
	NFromPage int32
	NToPage   int32
}

type PAGESET struct {
	CbStruct   uint32
	FOddPages  BOOL
	FEvenPages BOOL
	CPageRange uint32
	RgPages    [1]PAGERANGE
}

type OLECMD struct {
	CmdID uint32
	Cmdf  uint32
}

type OLECMDTEXT struct {
	Cmdtextf uint32
	CwActual uint32
	CwBuf    uint32
	Rgwz     [1]uint16
}

type OLEUIINSERTOBJECT = OLEUIINSERTOBJECTW
type OLEUIINSERTOBJECTW struct {
	CbStruct         uint32
	DwFlags          uint32
	HWndOwner        HWND
	LpszCaption      PWSTR
	LpfnHook         LPFNOLEUIHOOK
	LCustData        LPARAM
	HInstance        HINSTANCE
	LpszTemplate     PWSTR
	HResource        HRSRC
	Clsid            syscall.GUID
	LpszFile         PWSTR
	CchFile          uint32
	CClsidExclude    uint32
	LpClsidExclude   *syscall.GUID
	Iid              syscall.GUID
	OleRender        uint32
	LpFormatEtc      *FORMATETC
	LpIOleClientSite *IOleClientSite
	LpIStorage       *IStorage
	PpvObj           unsafe.Pointer
	Sc               int32
	HMetaPict        uintptr
}

type OLEUIINSERTOBJECTA struct {
	CbStruct         uint32
	DwFlags          uint32
	HWndOwner        HWND
	LpszCaption      PSTR
	LpfnHook         LPFNOLEUIHOOK
	LCustData        LPARAM
	HInstance        HINSTANCE
	LpszTemplate     PSTR
	HResource        HRSRC
	Clsid            syscall.GUID
	LpszFile         PSTR
	CchFile          uint32
	CClsidExclude    uint32
	LpClsidExclude   *syscall.GUID
	Iid              syscall.GUID
	OleRender        uint32
	LpFormatEtc      *FORMATETC
	LpIOleClientSite *IOleClientSite
	LpIStorage       *IStorage
	PpvObj           unsafe.Pointer
	Sc               int32
	HMetaPict        uintptr
}

type OLEUIPASTEENTRY = OLEUIPASTEENTRYW
type OLEUIPASTEENTRYW struct {
	Fmtetc          FORMATETC
	LpstrFormatName PWSTR
	LpstrResultText PWSTR
	DwFlags         uint32
	DwScratchSpace  uint32
}

type OLEUIPASTEENTRYA struct {
	Fmtetc          FORMATETC
	LpstrFormatName PSTR
	LpstrResultText PSTR
	DwFlags         uint32
	DwScratchSpace  uint32
}

type OLEUIPASTESPECIAL = OLEUIPASTESPECIALW
type OLEUIPASTESPECIALW struct {
	CbStruct        uint32
	DwFlags         uint32
	HWndOwner       HWND
	LpszCaption     PWSTR
	LpfnHook        LPFNOLEUIHOOK
	LCustData       LPARAM
	HInstance       HINSTANCE
	LpszTemplate    PWSTR
	HResource       HRSRC
	LpSrcDataObj    *IDataObject
	ArrPasteEntries *OLEUIPASTEENTRYW
	CPasteEntries   int32
	ArrLinkTypes    *uint32
	CLinkTypes      int32
	CClsidExclude   uint32
	LpClsidExclude  *syscall.GUID
	NSelectedIndex  int32
	FLink           BOOL
	HMetaPict       uintptr
	Sizel           SIZE
}

type OLEUIPASTESPECIALA struct {
	CbStruct        uint32
	DwFlags         uint32
	HWndOwner       HWND
	LpszCaption     PSTR
	LpfnHook        LPFNOLEUIHOOK
	LCustData       LPARAM
	HInstance       HINSTANCE
	LpszTemplate    PSTR
	HResource       HRSRC
	LpSrcDataObj    *IDataObject
	ArrPasteEntries *OLEUIPASTEENTRYA
	CPasteEntries   int32
	ArrLinkTypes    *uint32
	CLinkTypes      int32
	CClsidExclude   uint32
	LpClsidExclude  *syscall.GUID
	NSelectedIndex  int32
	FLink           BOOL
	HMetaPict       uintptr
	Sizel           SIZE
}

type OLEUIEDITLINKS = OLEUIEDITLINKSW
type OLEUIEDITLINKSW struct {
	CbStruct             uint32
	DwFlags              uint32
	HWndOwner            HWND
	LpszCaption          PWSTR
	LpfnHook             LPFNOLEUIHOOK
	LCustData            LPARAM
	HInstance            HINSTANCE
	LpszTemplate         PWSTR
	HResource            HRSRC
	LpOleUILinkContainer *IOleUILinkContainerW
}

type OLEUIEDITLINKSA struct {
	CbStruct             uint32
	DwFlags              uint32
	HWndOwner            HWND
	LpszCaption          PSTR
	LpfnHook             LPFNOLEUIHOOK
	LCustData            LPARAM
	HInstance            HINSTANCE
	LpszTemplate         PSTR
	HResource            HRSRC
	LpOleUILinkContainer *IOleUILinkContainerA
}

type OLEUICHANGEICON = OLEUICHANGEICONW
type OLEUICHANGEICONW struct {
	CbStruct     uint32
	DwFlags      uint32
	HWndOwner    HWND
	LpszCaption  PWSTR
	LpfnHook     LPFNOLEUIHOOK
	LCustData    LPARAM
	HInstance    HINSTANCE
	LpszTemplate PWSTR
	HResource    HRSRC
	HMetaPict    uintptr
	Clsid        syscall.GUID
	SzIconExe    [260]uint16
	CchIconExe   int32
}

type OLEUICHANGEICONA struct {
	CbStruct     uint32
	DwFlags      uint32
	HWndOwner    HWND
	LpszCaption  PSTR
	LpfnHook     LPFNOLEUIHOOK
	LCustData    LPARAM
	HInstance    HINSTANCE
	LpszTemplate PSTR
	HResource    HRSRC
	HMetaPict    uintptr
	Clsid        syscall.GUID
	SzIconExe    [260]CHAR
	CchIconExe   int32
}

type OLEUICONVERT = OLEUICONVERTW
type OLEUICONVERTW struct {
	CbStruct             uint32
	DwFlags              uint32
	HWndOwner            HWND
	LpszCaption          PWSTR
	LpfnHook             LPFNOLEUIHOOK
	LCustData            LPARAM
	HInstance            HINSTANCE
	LpszTemplate         PWSTR
	HResource            HRSRC
	Clsid                syscall.GUID
	ClsidConvertDefault  syscall.GUID
	ClsidActivateDefault syscall.GUID
	ClsidNew             syscall.GUID
	DvAspect             uint32
	WFormat              uint16
	FIsLinkedObject      BOOL
	HMetaPict            uintptr
	LpszUserType         PWSTR
	FObjectsIconChanged  BOOL
	LpszDefLabel         PWSTR
	CClsidExclude        uint32
	LpClsidExclude       *syscall.GUID
}

type OLEUICONVERTA struct {
	CbStruct             uint32
	DwFlags              uint32
	HWndOwner            HWND
	LpszCaption          PSTR
	LpfnHook             LPFNOLEUIHOOK
	LCustData            LPARAM
	HInstance            HINSTANCE
	LpszTemplate         PSTR
	HResource            HRSRC
	Clsid                syscall.GUID
	ClsidConvertDefault  syscall.GUID
	ClsidActivateDefault syscall.GUID
	ClsidNew             syscall.GUID
	DvAspect             uint32
	WFormat              uint16
	FIsLinkedObject      BOOL
	HMetaPict            uintptr
	LpszUserType         PSTR
	FObjectsIconChanged  BOOL
	LpszDefLabel         PSTR
	CClsidExclude        uint32
	LpClsidExclude       *syscall.GUID
}

type OLEUIBUSY = OLEUIBUSYW
type OLEUIBUSYW struct {
	CbStruct     uint32
	DwFlags      uint32
	HWndOwner    HWND
	LpszCaption  PWSTR
	LpfnHook     LPFNOLEUIHOOK
	LCustData    LPARAM
	HInstance    HINSTANCE
	LpszTemplate PWSTR
	HResource    HRSRC
	HTask        HTASK
	LphWndDialog *HWND
}

type OLEUIBUSYA struct {
	CbStruct     uint32
	DwFlags      uint32
	HWndOwner    HWND
	LpszCaption  PSTR
	LpfnHook     LPFNOLEUIHOOK
	LCustData    LPARAM
	HInstance    HINSTANCE
	LpszTemplate PSTR
	HResource    HRSRC
	HTask        HTASK
	LphWndDialog *HWND
}

type OLEUICHANGESOURCE = OLEUICHANGESOURCEW
type OLEUICHANGESOURCEW struct {
	CbStruct             uint32
	DwFlags              uint32
	HWndOwner            HWND
	LpszCaption          PWSTR
	LpfnHook             LPFNOLEUIHOOK
	LCustData            LPARAM
	HInstance            HINSTANCE
	LpszTemplate         PWSTR
	HResource            HRSRC
	LpOFN                *OPENFILENAMEW
	DwReserved1          [4]uint32
	LpOleUILinkContainer *IOleUILinkContainerW
	DwLink               uint32
	LpszDisplayName      PWSTR
	NFileLength          uint32
	LpszFrom             PWSTR
	LpszTo               PWSTR
}

type OLEUICHANGESOURCEA struct {
	CbStruct             uint32
	DwFlags              uint32
	HWndOwner            HWND
	LpszCaption          PSTR
	LpfnHook             LPFNOLEUIHOOK
	LCustData            LPARAM
	HInstance            HINSTANCE
	LpszTemplate         PSTR
	HResource            HRSRC
	LpOFN                *OPENFILENAMEA
	DwReserved1          [4]uint32
	LpOleUILinkContainer *IOleUILinkContainerA
	DwLink               uint32
	LpszDisplayName      PSTR
	NFileLength          uint32
	LpszFrom             PSTR
	LpszTo               PSTR
}

type OLEUIGNRLPROPS = OLEUIGNRLPROPSW
type OLEUIGNRLPROPSW struct {
	CbStruct    uint32
	DwFlags     uint32
	DwReserved1 [2]uint32
	LpfnHook    LPFNOLEUIHOOK
	LCustData   LPARAM
	DwReserved2 [3]uint32
	LpOP        *OLEUIOBJECTPROPSW
}

type OLEUIGNRLPROPSA struct {
	CbStruct    uint32
	DwFlags     uint32
	DwReserved1 [2]uint32
	LpfnHook    LPFNOLEUIHOOK
	LCustData   LPARAM
	DwReserved2 [3]uint32
	LpOP        *OLEUIOBJECTPROPSA
}

type OLEUIVIEWPROPS = OLEUIVIEWPROPSW
type OLEUIVIEWPROPSW struct {
	CbStruct    uint32
	DwFlags     uint32
	DwReserved1 [2]uint32
	LpfnHook    LPFNOLEUIHOOK
	LCustData   LPARAM
	DwReserved2 [3]uint32
	LpOP        *OLEUIOBJECTPROPSW
	NScaleMin   int32
	NScaleMax   int32
}

type OLEUIVIEWPROPSA struct {
	CbStruct    uint32
	DwFlags     uint32
	DwReserved1 [2]uint32
	LpfnHook    LPFNOLEUIHOOK
	LCustData   LPARAM
	DwReserved2 [3]uint32
	LpOP        *OLEUIOBJECTPROPSA
	NScaleMin   int32
	NScaleMax   int32
}

type OLEUILINKPROPS = OLEUILINKPROPSW
type OLEUILINKPROPSW struct {
	CbStruct    uint32
	DwFlags     uint32
	DwReserved1 [2]uint32
	LpfnHook    LPFNOLEUIHOOK
	LCustData   LPARAM
	DwReserved2 [3]uint32
	LpOP        *OLEUIOBJECTPROPSW
}

type OLEUILINKPROPSA struct {
	CbStruct    uint32
	DwFlags     uint32
	DwReserved1 [2]uint32
	LpfnHook    LPFNOLEUIHOOK
	LCustData   LPARAM
	DwReserved2 [3]uint32
	LpOP        *OLEUIOBJECTPROPSA
}

type OLEUIOBJECTPROPS = OLEUIOBJECTPROPSW
type OLEUIOBJECTPROPSW struct {
	CbStruct   uint32
	DwFlags    uint32
	LpPS       *PROPSHEETHEADERW_V2
	DwObject   uint32
	LpObjInfo  *IOleUIObjInfoW
	DwLink     uint32
	LpLinkInfo *IOleUILinkInfoW
	LpGP       *OLEUIGNRLPROPSW
	LpVP       *OLEUIVIEWPROPSW
	LpLP       *OLEUILINKPROPSW
}

type OLEUIOBJECTPROPSA struct {
	CbStruct   uint32
	DwFlags    uint32
	LpPS       *PROPSHEETHEADERA_V2
	DwObject   uint32
	LpObjInfo  *IOleUIObjInfoA
	DwLink     uint32
	LpLinkInfo *IOleUILinkInfoA
	LpGP       *OLEUIGNRLPROPSA
	LpVP       *OLEUIVIEWPROPSA
	LpLP       *OLEUILINKPROPSA
}

// func types

type LPFNOLEUIHOOK = uintptr
type LPFNOLEUIHOOK_func = func(param0 HWND, param1 uint32, param2 WPARAM, param3 LPARAM) uint32

// interfaces

// 00020405-0000-0000-C000-000000000046
var IID_ICreateTypeInfo = syscall.GUID{0x00020405, 0x0000, 0x0000,
	[8]byte{0xC0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x46}}

type ICreateTypeInfoInterface interface {
	IUnknownInterface
	SetGuid(guid *syscall.GUID) HRESULT
	SetTypeFlags(uTypeFlags uint32) HRESULT
	SetDocString(pStrDoc PWSTR) HRESULT
	SetHelpContext(dwHelpContext uint32) HRESULT
	SetVersion(wMajorVerNum uint16, wMinorVerNum uint16) HRESULT
	AddRefTypeInfo(pTInfo *ITypeInfo, phRefType *uint32) HRESULT
	AddFuncDesc(index uint32, pFuncDesc *FUNCDESC) HRESULT
	AddImplType(index uint32, hRefType uint32) HRESULT
	SetImplTypeFlags(index uint32, implTypeFlags int32) HRESULT
	SetAlignment(cbAlignment uint16) HRESULT
	SetSchema(pStrSchema PWSTR) HRESULT
	AddVarDesc(index uint32, pVarDesc *VARDESC) HRESULT
	SetFuncAndParamNames(index uint32, rgszNames *PWSTR, cNames uint32) HRESULT
	SetVarName(index uint32, szName PWSTR) HRESULT
	SetTypeDescAlias(pTDescAlias *TYPEDESC) HRESULT
	DefineFuncAsDllEntry(index uint32, szDllName PWSTR, szProcName PWSTR) HRESULT
	SetFuncDocString(index uint32, szDocString PWSTR) HRESULT
	SetVarDocString(index uint32, szDocString PWSTR) HRESULT
	SetFuncHelpContext(index uint32, dwHelpContext uint32) HRESULT
	SetVarHelpContext(index uint32, dwHelpContext uint32) HRESULT
	SetMops(index uint32, bstrMops BSTR) HRESULT
	SetTypeIdldesc(pIdlDesc *IDLDESC) HRESULT
	LayOut() HRESULT
}

type ICreateTypeInfoVtbl struct {
	IUnknownVtbl
	SetGuid              uintptr
	SetTypeFlags         uintptr
	SetDocString         uintptr
	SetHelpContext       uintptr
	SetVersion           uintptr
	AddRefTypeInfo       uintptr
	AddFuncDesc          uintptr
	AddImplType          uintptr
	SetImplTypeFlags     uintptr
	SetAlignment         uintptr
	SetSchema            uintptr
	AddVarDesc           uintptr
	SetFuncAndParamNames uintptr
	SetVarName           uintptr
	SetTypeDescAlias     uintptr
	DefineFuncAsDllEntry uintptr
	SetFuncDocString     uintptr
	SetVarDocString      uintptr
	SetFuncHelpContext   uintptr
	SetVarHelpContext    uintptr
	SetMops              uintptr
	SetTypeIdldesc       uintptr
	LayOut               uintptr
}

type ICreateTypeInfo struct {
	IUnknown
}

func (this *ICreateTypeInfo) Vtbl() *ICreateTypeInfoVtbl {
	return (*ICreateTypeInfoVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *ICreateTypeInfo) SetGuid(guid *syscall.GUID) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetGuid, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(guid)))
	return HRESULT(ret)
}

func (this *ICreateTypeInfo) SetTypeFlags(uTypeFlags uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetTypeFlags, uintptr(unsafe.Pointer(this)), uintptr(uTypeFlags))
	return HRESULT(ret)
}

func (this *ICreateTypeInfo) SetDocString(pStrDoc PWSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetDocString, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pStrDoc)))
	return HRESULT(ret)
}

func (this *ICreateTypeInfo) SetHelpContext(dwHelpContext uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetHelpContext, uintptr(unsafe.Pointer(this)), uintptr(dwHelpContext))
	return HRESULT(ret)
}

func (this *ICreateTypeInfo) SetVersion(wMajorVerNum uint16, wMinorVerNum uint16) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetVersion, uintptr(unsafe.Pointer(this)), uintptr(wMajorVerNum), uintptr(wMinorVerNum))
	return HRESULT(ret)
}

func (this *ICreateTypeInfo) AddRefTypeInfo(pTInfo *ITypeInfo, phRefType *uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().AddRefTypeInfo, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pTInfo)), uintptr(unsafe.Pointer(phRefType)))
	return HRESULT(ret)
}

func (this *ICreateTypeInfo) AddFuncDesc(index uint32, pFuncDesc *FUNCDESC) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().AddFuncDesc, uintptr(unsafe.Pointer(this)), uintptr(index), uintptr(unsafe.Pointer(pFuncDesc)))
	return HRESULT(ret)
}

func (this *ICreateTypeInfo) AddImplType(index uint32, hRefType uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().AddImplType, uintptr(unsafe.Pointer(this)), uintptr(index), uintptr(hRefType))
	return HRESULT(ret)
}

func (this *ICreateTypeInfo) SetImplTypeFlags(index uint32, implTypeFlags int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetImplTypeFlags, uintptr(unsafe.Pointer(this)), uintptr(index), uintptr(implTypeFlags))
	return HRESULT(ret)
}

func (this *ICreateTypeInfo) SetAlignment(cbAlignment uint16) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetAlignment, uintptr(unsafe.Pointer(this)), uintptr(cbAlignment))
	return HRESULT(ret)
}

func (this *ICreateTypeInfo) SetSchema(pStrSchema PWSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetSchema, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pStrSchema)))
	return HRESULT(ret)
}

func (this *ICreateTypeInfo) AddVarDesc(index uint32, pVarDesc *VARDESC) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().AddVarDesc, uintptr(unsafe.Pointer(this)), uintptr(index), uintptr(unsafe.Pointer(pVarDesc)))
	return HRESULT(ret)
}

func (this *ICreateTypeInfo) SetFuncAndParamNames(index uint32, rgszNames *PWSTR, cNames uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetFuncAndParamNames, uintptr(unsafe.Pointer(this)), uintptr(index), uintptr(unsafe.Pointer(rgszNames)), uintptr(cNames))
	return HRESULT(ret)
}

func (this *ICreateTypeInfo) SetVarName(index uint32, szName PWSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetVarName, uintptr(unsafe.Pointer(this)), uintptr(index), uintptr(unsafe.Pointer(szName)))
	return HRESULT(ret)
}

func (this *ICreateTypeInfo) SetTypeDescAlias(pTDescAlias *TYPEDESC) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetTypeDescAlias, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pTDescAlias)))
	return HRESULT(ret)
}

func (this *ICreateTypeInfo) DefineFuncAsDllEntry(index uint32, szDllName PWSTR, szProcName PWSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().DefineFuncAsDllEntry, uintptr(unsafe.Pointer(this)), uintptr(index), uintptr(unsafe.Pointer(szDllName)), uintptr(unsafe.Pointer(szProcName)))
	return HRESULT(ret)
}

func (this *ICreateTypeInfo) SetFuncDocString(index uint32, szDocString PWSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetFuncDocString, uintptr(unsafe.Pointer(this)), uintptr(index), uintptr(unsafe.Pointer(szDocString)))
	return HRESULT(ret)
}

func (this *ICreateTypeInfo) SetVarDocString(index uint32, szDocString PWSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetVarDocString, uintptr(unsafe.Pointer(this)), uintptr(index), uintptr(unsafe.Pointer(szDocString)))
	return HRESULT(ret)
}

func (this *ICreateTypeInfo) SetFuncHelpContext(index uint32, dwHelpContext uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetFuncHelpContext, uintptr(unsafe.Pointer(this)), uintptr(index), uintptr(dwHelpContext))
	return HRESULT(ret)
}

func (this *ICreateTypeInfo) SetVarHelpContext(index uint32, dwHelpContext uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetVarHelpContext, uintptr(unsafe.Pointer(this)), uintptr(index), uintptr(dwHelpContext))
	return HRESULT(ret)
}

func (this *ICreateTypeInfo) SetMops(index uint32, bstrMops BSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetMops, uintptr(unsafe.Pointer(this)), uintptr(index), uintptr(unsafe.Pointer(bstrMops)))
	return HRESULT(ret)
}

func (this *ICreateTypeInfo) SetTypeIdldesc(pIdlDesc *IDLDESC) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetTypeIdldesc, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pIdlDesc)))
	return HRESULT(ret)
}

func (this *ICreateTypeInfo) LayOut() HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().LayOut, uintptr(unsafe.Pointer(this)))
	return HRESULT(ret)
}

// 0002040E-0000-0000-C000-000000000046
var IID_ICreateTypeInfo2 = syscall.GUID{0x0002040E, 0x0000, 0x0000,
	[8]byte{0xC0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x46}}

type ICreateTypeInfo2Interface interface {
	ICreateTypeInfoInterface
	DeleteFuncDesc(index uint32) HRESULT
	DeleteFuncDescByMemId(memid int32, invKind INVOKEKIND) HRESULT
	DeleteVarDesc(index uint32) HRESULT
	DeleteVarDescByMemId(memid int32) HRESULT
	DeleteImplType(index uint32) HRESULT
	SetCustData(guid *syscall.GUID, pVarVal *VARIANT) HRESULT
	SetFuncCustData(index uint32, guid *syscall.GUID, pVarVal *VARIANT) HRESULT
	SetParamCustData(indexFunc uint32, indexParam uint32, guid *syscall.GUID, pVarVal *VARIANT) HRESULT
	SetVarCustData(index uint32, guid *syscall.GUID, pVarVal *VARIANT) HRESULT
	SetImplTypeCustData(index uint32, guid *syscall.GUID, pVarVal *VARIANT) HRESULT
	SetHelpStringContext(dwHelpStringContext uint32) HRESULT
	SetFuncHelpStringContext(index uint32, dwHelpStringContext uint32) HRESULT
	SetVarHelpStringContext(index uint32, dwHelpStringContext uint32) HRESULT
	Invalidate() HRESULT
	SetName(szName PWSTR) HRESULT
}

type ICreateTypeInfo2Vtbl struct {
	ICreateTypeInfoVtbl
	DeleteFuncDesc           uintptr
	DeleteFuncDescByMemId    uintptr
	DeleteVarDesc            uintptr
	DeleteVarDescByMemId     uintptr
	DeleteImplType           uintptr
	SetCustData              uintptr
	SetFuncCustData          uintptr
	SetParamCustData         uintptr
	SetVarCustData           uintptr
	SetImplTypeCustData      uintptr
	SetHelpStringContext     uintptr
	SetFuncHelpStringContext uintptr
	SetVarHelpStringContext  uintptr
	Invalidate               uintptr
	SetName                  uintptr
}

type ICreateTypeInfo2 struct {
	ICreateTypeInfo
}

func (this *ICreateTypeInfo2) Vtbl() *ICreateTypeInfo2Vtbl {
	return (*ICreateTypeInfo2Vtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *ICreateTypeInfo2) DeleteFuncDesc(index uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().DeleteFuncDesc, uintptr(unsafe.Pointer(this)), uintptr(index))
	return HRESULT(ret)
}

func (this *ICreateTypeInfo2) DeleteFuncDescByMemId(memid int32, invKind INVOKEKIND) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().DeleteFuncDescByMemId, uintptr(unsafe.Pointer(this)), uintptr(memid), uintptr(invKind))
	return HRESULT(ret)
}

func (this *ICreateTypeInfo2) DeleteVarDesc(index uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().DeleteVarDesc, uintptr(unsafe.Pointer(this)), uintptr(index))
	return HRESULT(ret)
}

func (this *ICreateTypeInfo2) DeleteVarDescByMemId(memid int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().DeleteVarDescByMemId, uintptr(unsafe.Pointer(this)), uintptr(memid))
	return HRESULT(ret)
}

func (this *ICreateTypeInfo2) DeleteImplType(index uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().DeleteImplType, uintptr(unsafe.Pointer(this)), uintptr(index))
	return HRESULT(ret)
}

func (this *ICreateTypeInfo2) SetCustData(guid *syscall.GUID, pVarVal *VARIANT) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetCustData, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(guid)), uintptr(unsafe.Pointer(pVarVal)))
	return HRESULT(ret)
}

func (this *ICreateTypeInfo2) SetFuncCustData(index uint32, guid *syscall.GUID, pVarVal *VARIANT) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetFuncCustData, uintptr(unsafe.Pointer(this)), uintptr(index), uintptr(unsafe.Pointer(guid)), uintptr(unsafe.Pointer(pVarVal)))
	return HRESULT(ret)
}

func (this *ICreateTypeInfo2) SetParamCustData(indexFunc uint32, indexParam uint32, guid *syscall.GUID, pVarVal *VARIANT) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetParamCustData, uintptr(unsafe.Pointer(this)), uintptr(indexFunc), uintptr(indexParam), uintptr(unsafe.Pointer(guid)), uintptr(unsafe.Pointer(pVarVal)))
	return HRESULT(ret)
}

func (this *ICreateTypeInfo2) SetVarCustData(index uint32, guid *syscall.GUID, pVarVal *VARIANT) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetVarCustData, uintptr(unsafe.Pointer(this)), uintptr(index), uintptr(unsafe.Pointer(guid)), uintptr(unsafe.Pointer(pVarVal)))
	return HRESULT(ret)
}

func (this *ICreateTypeInfo2) SetImplTypeCustData(index uint32, guid *syscall.GUID, pVarVal *VARIANT) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetImplTypeCustData, uintptr(unsafe.Pointer(this)), uintptr(index), uintptr(unsafe.Pointer(guid)), uintptr(unsafe.Pointer(pVarVal)))
	return HRESULT(ret)
}

func (this *ICreateTypeInfo2) SetHelpStringContext(dwHelpStringContext uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetHelpStringContext, uintptr(unsafe.Pointer(this)), uintptr(dwHelpStringContext))
	return HRESULT(ret)
}

func (this *ICreateTypeInfo2) SetFuncHelpStringContext(index uint32, dwHelpStringContext uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetFuncHelpStringContext, uintptr(unsafe.Pointer(this)), uintptr(index), uintptr(dwHelpStringContext))
	return HRESULT(ret)
}

func (this *ICreateTypeInfo2) SetVarHelpStringContext(index uint32, dwHelpStringContext uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetVarHelpStringContext, uintptr(unsafe.Pointer(this)), uintptr(index), uintptr(dwHelpStringContext))
	return HRESULT(ret)
}

func (this *ICreateTypeInfo2) Invalidate() HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Invalidate, uintptr(unsafe.Pointer(this)))
	return HRESULT(ret)
}

func (this *ICreateTypeInfo2) SetName(szName PWSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetName, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(szName)))
	return HRESULT(ret)
}

// 00020406-0000-0000-C000-000000000046
var IID_ICreateTypeLib = syscall.GUID{0x00020406, 0x0000, 0x0000,
	[8]byte{0xC0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x46}}

type ICreateTypeLibInterface interface {
	IUnknownInterface
	CreateTypeInfo(szName PWSTR, tkind TYPEKIND, ppCTInfo **ICreateTypeInfo) HRESULT
	SetName(szName PWSTR) HRESULT
	SetVersion(wMajorVerNum uint16, wMinorVerNum uint16) HRESULT
	SetGuid(guid *syscall.GUID) HRESULT
	SetDocString(szDoc PWSTR) HRESULT
	SetHelpFileName(szHelpFileName PWSTR) HRESULT
	SetHelpContext(dwHelpContext uint32) HRESULT
	SetLcid(lcid uint32) HRESULT
	SetLibFlags(uLibFlags uint32) HRESULT
	SaveAllChanges() HRESULT
}

type ICreateTypeLibVtbl struct {
	IUnknownVtbl
	CreateTypeInfo  uintptr
	SetName         uintptr
	SetVersion      uintptr
	SetGuid         uintptr
	SetDocString    uintptr
	SetHelpFileName uintptr
	SetHelpContext  uintptr
	SetLcid         uintptr
	SetLibFlags     uintptr
	SaveAllChanges  uintptr
}

type ICreateTypeLib struct {
	IUnknown
}

func (this *ICreateTypeLib) Vtbl() *ICreateTypeLibVtbl {
	return (*ICreateTypeLibVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *ICreateTypeLib) CreateTypeInfo(szName PWSTR, tkind TYPEKIND, ppCTInfo **ICreateTypeInfo) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().CreateTypeInfo, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(szName)), uintptr(tkind), uintptr(unsafe.Pointer(ppCTInfo)))
	return HRESULT(ret)
}

func (this *ICreateTypeLib) SetName(szName PWSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetName, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(szName)))
	return HRESULT(ret)
}

func (this *ICreateTypeLib) SetVersion(wMajorVerNum uint16, wMinorVerNum uint16) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetVersion, uintptr(unsafe.Pointer(this)), uintptr(wMajorVerNum), uintptr(wMinorVerNum))
	return HRESULT(ret)
}

func (this *ICreateTypeLib) SetGuid(guid *syscall.GUID) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetGuid, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(guid)))
	return HRESULT(ret)
}

func (this *ICreateTypeLib) SetDocString(szDoc PWSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetDocString, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(szDoc)))
	return HRESULT(ret)
}

func (this *ICreateTypeLib) SetHelpFileName(szHelpFileName PWSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetHelpFileName, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(szHelpFileName)))
	return HRESULT(ret)
}

func (this *ICreateTypeLib) SetHelpContext(dwHelpContext uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetHelpContext, uintptr(unsafe.Pointer(this)), uintptr(dwHelpContext))
	return HRESULT(ret)
}

func (this *ICreateTypeLib) SetLcid(lcid uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetLcid, uintptr(unsafe.Pointer(this)), uintptr(lcid))
	return HRESULT(ret)
}

func (this *ICreateTypeLib) SetLibFlags(uLibFlags uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetLibFlags, uintptr(unsafe.Pointer(this)), uintptr(uLibFlags))
	return HRESULT(ret)
}

func (this *ICreateTypeLib) SaveAllChanges() HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SaveAllChanges, uintptr(unsafe.Pointer(this)))
	return HRESULT(ret)
}

// 0002040F-0000-0000-C000-000000000046
var IID_ICreateTypeLib2 = syscall.GUID{0x0002040F, 0x0000, 0x0000,
	[8]byte{0xC0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x46}}

type ICreateTypeLib2Interface interface {
	ICreateTypeLibInterface
	DeleteTypeInfo(szName PWSTR) HRESULT
	SetCustData(guid *syscall.GUID, pVarVal *VARIANT) HRESULT
	SetHelpStringContext(dwHelpStringContext uint32) HRESULT
	SetHelpStringDll(szFileName PWSTR) HRESULT
}

type ICreateTypeLib2Vtbl struct {
	ICreateTypeLibVtbl
	DeleteTypeInfo       uintptr
	SetCustData          uintptr
	SetHelpStringContext uintptr
	SetHelpStringDll     uintptr
}

type ICreateTypeLib2 struct {
	ICreateTypeLib
}

func (this *ICreateTypeLib2) Vtbl() *ICreateTypeLib2Vtbl {
	return (*ICreateTypeLib2Vtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *ICreateTypeLib2) DeleteTypeInfo(szName PWSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().DeleteTypeInfo, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(szName)))
	return HRESULT(ret)
}

func (this *ICreateTypeLib2) SetCustData(guid *syscall.GUID, pVarVal *VARIANT) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetCustData, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(guid)), uintptr(unsafe.Pointer(pVarVal)))
	return HRESULT(ret)
}

func (this *ICreateTypeLib2) SetHelpStringContext(dwHelpStringContext uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetHelpStringContext, uintptr(unsafe.Pointer(this)), uintptr(dwHelpStringContext))
	return HRESULT(ret)
}

func (this *ICreateTypeLib2) SetHelpStringDll(szFileName PWSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetHelpStringDll, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(szFileName)))
	return HRESULT(ret)
}

// 00020404-0000-0000-C000-000000000046
var IID_IEnumVARIANT = syscall.GUID{0x00020404, 0x0000, 0x0000,
	[8]byte{0xC0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x46}}

type IEnumVARIANTInterface interface {
	IUnknownInterface
	Next(celt uint32, rgVar *VARIANT, pCeltFetched *uint32) HRESULT
	Skip(celt uint32) HRESULT
	Reset() HRESULT
	Clone(ppEnum **IEnumVARIANT) HRESULT
}

type IEnumVARIANTVtbl struct {
	IUnknownVtbl
	Next  uintptr
	Skip  uintptr
	Reset uintptr
	Clone uintptr
}

type IEnumVARIANT struct {
	IUnknown
}

func (this *IEnumVARIANT) Vtbl() *IEnumVARIANTVtbl {
	return (*IEnumVARIANTVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IEnumVARIANT) Next(celt uint32, rgVar *VARIANT, pCeltFetched *uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Next, uintptr(unsafe.Pointer(this)), uintptr(celt), uintptr(unsafe.Pointer(rgVar)), uintptr(unsafe.Pointer(pCeltFetched)))
	return HRESULT(ret)
}

func (this *IEnumVARIANT) Skip(celt uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Skip, uintptr(unsafe.Pointer(this)), uintptr(celt))
	return HRESULT(ret)
}

func (this *IEnumVARIANT) Reset() HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Reset, uintptr(unsafe.Pointer(this)))
	return HRESULT(ret)
}

func (this *IEnumVARIANT) Clone(ppEnum **IEnumVARIANT) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Clone, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(ppEnum)))
	return HRESULT(ret)
}

// 00020410-0000-0000-C000-000000000046
var IID_ITypeChangeEvents = syscall.GUID{0x00020410, 0x0000, 0x0000,
	[8]byte{0xC0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x46}}

type ITypeChangeEventsInterface interface {
	IUnknownInterface
	RequestTypeChange(changeKind CHANGEKIND, pTInfoBefore *ITypeInfo, pStrName PWSTR, pfCancel *int32) HRESULT
	AfterTypeChange(changeKind CHANGEKIND, pTInfoAfter *ITypeInfo, pStrName PWSTR) HRESULT
}

type ITypeChangeEventsVtbl struct {
	IUnknownVtbl
	RequestTypeChange uintptr
	AfterTypeChange   uintptr
}

type ITypeChangeEvents struct {
	IUnknown
}

func (this *ITypeChangeEvents) Vtbl() *ITypeChangeEventsVtbl {
	return (*ITypeChangeEventsVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *ITypeChangeEvents) RequestTypeChange(changeKind CHANGEKIND, pTInfoBefore *ITypeInfo, pStrName PWSTR, pfCancel *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().RequestTypeChange, uintptr(unsafe.Pointer(this)), uintptr(changeKind), uintptr(unsafe.Pointer(pTInfoBefore)), uintptr(unsafe.Pointer(pStrName)), uintptr(unsafe.Pointer(pfCancel)))
	return HRESULT(ret)
}

func (this *ITypeChangeEvents) AfterTypeChange(changeKind CHANGEKIND, pTInfoAfter *ITypeInfo, pStrName PWSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().AfterTypeChange, uintptr(unsafe.Pointer(this)), uintptr(changeKind), uintptr(unsafe.Pointer(pTInfoAfter)), uintptr(unsafe.Pointer(pStrName)))
	return HRESULT(ret)
}

// 22F03340-547D-101B-8E65-08002B2BD119
var IID_ICreateErrorInfo = syscall.GUID{0x22F03340, 0x547D, 0x101B,
	[8]byte{0x8E, 0x65, 0x08, 0x00, 0x2B, 0x2B, 0xD1, 0x19}}

type ICreateErrorInfoInterface interface {
	IUnknownInterface
	SetGUID(rguid *syscall.GUID) HRESULT
	SetSource(szSource PWSTR) HRESULT
	SetDescription(szDescription PWSTR) HRESULT
	SetHelpFile(szHelpFile PWSTR) HRESULT
	SetHelpContext(dwHelpContext uint32) HRESULT
}

type ICreateErrorInfoVtbl struct {
	IUnknownVtbl
	SetGUID        uintptr
	SetSource      uintptr
	SetDescription uintptr
	SetHelpFile    uintptr
	SetHelpContext uintptr
}

type ICreateErrorInfo struct {
	IUnknown
}

func (this *ICreateErrorInfo) Vtbl() *ICreateErrorInfoVtbl {
	return (*ICreateErrorInfoVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *ICreateErrorInfo) SetGUID(rguid *syscall.GUID) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetGUID, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(rguid)))
	return HRESULT(ret)
}

func (this *ICreateErrorInfo) SetSource(szSource PWSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetSource, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(szSource)))
	return HRESULT(ret)
}

func (this *ICreateErrorInfo) SetDescription(szDescription PWSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetDescription, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(szDescription)))
	return HRESULT(ret)
}

func (this *ICreateErrorInfo) SetHelpFile(szHelpFile PWSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetHelpFile, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(szHelpFile)))
	return HRESULT(ret)
}

func (this *ICreateErrorInfo) SetHelpContext(dwHelpContext uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetHelpContext, uintptr(unsafe.Pointer(this)), uintptr(dwHelpContext))
	return HRESULT(ret)
}

// 0000002E-0000-0000-C000-000000000046
var IID_ITypeFactory = syscall.GUID{0x0000002E, 0x0000, 0x0000,
	[8]byte{0xC0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x46}}

type ITypeFactoryInterface interface {
	IUnknownInterface
	CreateFromTypeInfo(pTypeInfo *ITypeInfo, riid *syscall.GUID, ppv **IUnknown) HRESULT
}

type ITypeFactoryVtbl struct {
	IUnknownVtbl
	CreateFromTypeInfo uintptr
}

type ITypeFactory struct {
	IUnknown
}

func (this *ITypeFactory) Vtbl() *ITypeFactoryVtbl {
	return (*ITypeFactoryVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *ITypeFactory) CreateFromTypeInfo(pTypeInfo *ITypeInfo, riid *syscall.GUID, ppv **IUnknown) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().CreateFromTypeInfo, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pTypeInfo)), uintptr(unsafe.Pointer(riid)), uintptr(unsafe.Pointer(ppv)))
	return HRESULT(ret)
}

// 0000002D-0000-0000-C000-000000000046
var IID_ITypeMarshal = syscall.GUID{0x0000002D, 0x0000, 0x0000,
	[8]byte{0xC0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x46}}

type ITypeMarshalInterface interface {
	IUnknownInterface
	Size(pvType unsafe.Pointer, dwDestContext uint32, pvDestContext unsafe.Pointer, pSize *uint32) HRESULT
	Marshal(pvType unsafe.Pointer, dwDestContext uint32, pvDestContext unsafe.Pointer, cbBufferLength uint32, pBuffer *byte, pcbWritten *uint32) HRESULT
	Unmarshal(pvType unsafe.Pointer, dwFlags uint32, cbBufferLength uint32, pBuffer *byte, pcbRead *uint32) HRESULT
	Free(pvType unsafe.Pointer) HRESULT
}

type ITypeMarshalVtbl struct {
	IUnknownVtbl
	Size      uintptr
	Marshal   uintptr
	Unmarshal uintptr
	Free      uintptr
}

type ITypeMarshal struct {
	IUnknown
}

func (this *ITypeMarshal) Vtbl() *ITypeMarshalVtbl {
	return (*ITypeMarshalVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *ITypeMarshal) Size(pvType unsafe.Pointer, dwDestContext uint32, pvDestContext unsafe.Pointer, pSize *uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Size, uintptr(unsafe.Pointer(this)), uintptr(pvType), uintptr(dwDestContext), uintptr(pvDestContext), uintptr(unsafe.Pointer(pSize)))
	return HRESULT(ret)
}

func (this *ITypeMarshal) Marshal(pvType unsafe.Pointer, dwDestContext uint32, pvDestContext unsafe.Pointer, cbBufferLength uint32, pBuffer *byte, pcbWritten *uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Marshal, uintptr(unsafe.Pointer(this)), uintptr(pvType), uintptr(dwDestContext), uintptr(pvDestContext), uintptr(cbBufferLength), uintptr(unsafe.Pointer(pBuffer)), uintptr(unsafe.Pointer(pcbWritten)))
	return HRESULT(ret)
}

func (this *ITypeMarshal) Unmarshal(pvType unsafe.Pointer, dwFlags uint32, cbBufferLength uint32, pBuffer *byte, pcbRead *uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Unmarshal, uintptr(unsafe.Pointer(this)), uintptr(pvType), uintptr(dwFlags), uintptr(cbBufferLength), uintptr(unsafe.Pointer(pBuffer)), uintptr(unsafe.Pointer(pcbRead)))
	return HRESULT(ret)
}

func (this *ITypeMarshal) Free(pvType unsafe.Pointer) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Free, uintptr(unsafe.Pointer(this)), uintptr(pvType))
	return HRESULT(ret)
}

// 0000002F-0000-0000-C000-000000000046
var IID_IRecordInfo = syscall.GUID{0x0000002F, 0x0000, 0x0000,
	[8]byte{0xC0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x46}}

type IRecordInfoInterface interface {
	IUnknownInterface
	RecordInit(pvNew unsafe.Pointer) HRESULT
	RecordClear(pvExisting unsafe.Pointer) HRESULT
	RecordCopy(pvExisting unsafe.Pointer, pvNew unsafe.Pointer) HRESULT
	GetGuid(pguid *syscall.GUID) HRESULT
	GetName(pbstrName *BSTR) HRESULT
	GetSize(pcbSize *uint32) HRESULT
	GetTypeInfo(ppTypeInfo **ITypeInfo) HRESULT
	GetField(pvData unsafe.Pointer, szFieldName PWSTR, pvarField *VARIANT) HRESULT
	GetFieldNoCopy(pvData unsafe.Pointer, szFieldName PWSTR, pvarField *VARIANT, ppvDataCArray unsafe.Pointer) HRESULT
	PutField(wFlags uint32, pvData unsafe.Pointer, szFieldName PWSTR, pvarField *VARIANT) HRESULT
	PutFieldNoCopy(wFlags uint32, pvData unsafe.Pointer, szFieldName PWSTR, pvarField *VARIANT) HRESULT
	GetFieldNames(pcNames *uint32, rgBstrNames *BSTR) HRESULT
	IsMatchingType(pRecordInfo *IRecordInfo) BOOL
	RecordCreate() unsafe.Pointer
	RecordCreateCopy(pvSource unsafe.Pointer, ppvDest unsafe.Pointer) HRESULT
	RecordDestroy(pvRecord unsafe.Pointer) HRESULT
}

type IRecordInfoVtbl struct {
	IUnknownVtbl
	RecordInit       uintptr
	RecordClear      uintptr
	RecordCopy       uintptr
	GetGuid          uintptr
	GetName          uintptr
	GetSize          uintptr
	GetTypeInfo      uintptr
	GetField         uintptr
	GetFieldNoCopy   uintptr
	PutField         uintptr
	PutFieldNoCopy   uintptr
	GetFieldNames    uintptr
	IsMatchingType   uintptr
	RecordCreate     uintptr
	RecordCreateCopy uintptr
	RecordDestroy    uintptr
}

type IRecordInfo struct {
	IUnknown
}

func (this *IRecordInfo) Vtbl() *IRecordInfoVtbl {
	return (*IRecordInfoVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IRecordInfo) RecordInit(pvNew unsafe.Pointer) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().RecordInit, uintptr(unsafe.Pointer(this)), uintptr(pvNew))
	return HRESULT(ret)
}

func (this *IRecordInfo) RecordClear(pvExisting unsafe.Pointer) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().RecordClear, uintptr(unsafe.Pointer(this)), uintptr(pvExisting))
	return HRESULT(ret)
}

func (this *IRecordInfo) RecordCopy(pvExisting unsafe.Pointer, pvNew unsafe.Pointer) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().RecordCopy, uintptr(unsafe.Pointer(this)), uintptr(pvExisting), uintptr(pvNew))
	return HRESULT(ret)
}

func (this *IRecordInfo) GetGuid(pguid *syscall.GUID) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetGuid, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pguid)))
	return HRESULT(ret)
}

func (this *IRecordInfo) GetName(pbstrName *BSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetName, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pbstrName)))
	return HRESULT(ret)
}

func (this *IRecordInfo) GetSize(pcbSize *uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetSize, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pcbSize)))
	return HRESULT(ret)
}

func (this *IRecordInfo) GetTypeInfo(ppTypeInfo **ITypeInfo) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetTypeInfo, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(ppTypeInfo)))
	return HRESULT(ret)
}

func (this *IRecordInfo) GetField(pvData unsafe.Pointer, szFieldName PWSTR, pvarField *VARIANT) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetField, uintptr(unsafe.Pointer(this)), uintptr(pvData), uintptr(unsafe.Pointer(szFieldName)), uintptr(unsafe.Pointer(pvarField)))
	return HRESULT(ret)
}

func (this *IRecordInfo) GetFieldNoCopy(pvData unsafe.Pointer, szFieldName PWSTR, pvarField *VARIANT, ppvDataCArray unsafe.Pointer) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetFieldNoCopy, uintptr(unsafe.Pointer(this)), uintptr(pvData), uintptr(unsafe.Pointer(szFieldName)), uintptr(unsafe.Pointer(pvarField)), uintptr(ppvDataCArray))
	return HRESULT(ret)
}

func (this *IRecordInfo) PutField(wFlags uint32, pvData unsafe.Pointer, szFieldName PWSTR, pvarField *VARIANT) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().PutField, uintptr(unsafe.Pointer(this)), uintptr(wFlags), uintptr(pvData), uintptr(unsafe.Pointer(szFieldName)), uintptr(unsafe.Pointer(pvarField)))
	return HRESULT(ret)
}

func (this *IRecordInfo) PutFieldNoCopy(wFlags uint32, pvData unsafe.Pointer, szFieldName PWSTR, pvarField *VARIANT) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().PutFieldNoCopy, uintptr(unsafe.Pointer(this)), uintptr(wFlags), uintptr(pvData), uintptr(unsafe.Pointer(szFieldName)), uintptr(unsafe.Pointer(pvarField)))
	return HRESULT(ret)
}

func (this *IRecordInfo) GetFieldNames(pcNames *uint32, rgBstrNames *BSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetFieldNames, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pcNames)), uintptr(unsafe.Pointer(rgBstrNames)))
	return HRESULT(ret)
}

func (this *IRecordInfo) IsMatchingType(pRecordInfo *IRecordInfo) BOOL {
	ret, _, _ := syscall.SyscallN(this.Vtbl().IsMatchingType, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pRecordInfo)))
	return BOOL(ret)
}

func (this *IRecordInfo) RecordCreate() unsafe.Pointer {
	ret, _, _ := syscall.SyscallN(this.Vtbl().RecordCreate, uintptr(unsafe.Pointer(this)))
	return (unsafe.Pointer)(ret)
}

func (this *IRecordInfo) RecordCreateCopy(pvSource unsafe.Pointer, ppvDest unsafe.Pointer) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().RecordCreateCopy, uintptr(unsafe.Pointer(this)), uintptr(pvSource), uintptr(ppvDest))
	return HRESULT(ret)
}

func (this *IRecordInfo) RecordDestroy(pvRecord unsafe.Pointer) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().RecordDestroy, uintptr(unsafe.Pointer(this)), uintptr(pvRecord))
	return HRESULT(ret)
}

// 00000111-0000-0000-C000-000000000046
var IID_IOleAdviseHolder = syscall.GUID{0x00000111, 0x0000, 0x0000,
	[8]byte{0xC0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x46}}

type IOleAdviseHolderInterface interface {
	IUnknownInterface
	Advise(pAdvise *IAdviseSink, pdwConnection *uint32) HRESULT
	Unadvise(dwConnection uint32) HRESULT
	EnumAdvise(ppenumAdvise **IEnumSTATDATA) HRESULT
	SendOnRename(pmk *IMoniker) HRESULT
	SendOnSave() HRESULT
	SendOnClose() HRESULT
}

type IOleAdviseHolderVtbl struct {
	IUnknownVtbl
	Advise       uintptr
	Unadvise     uintptr
	EnumAdvise   uintptr
	SendOnRename uintptr
	SendOnSave   uintptr
	SendOnClose  uintptr
}

type IOleAdviseHolder struct {
	IUnknown
}

func (this *IOleAdviseHolder) Vtbl() *IOleAdviseHolderVtbl {
	return (*IOleAdviseHolderVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IOleAdviseHolder) Advise(pAdvise *IAdviseSink, pdwConnection *uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Advise, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pAdvise)), uintptr(unsafe.Pointer(pdwConnection)))
	return HRESULT(ret)
}

func (this *IOleAdviseHolder) Unadvise(dwConnection uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Unadvise, uintptr(unsafe.Pointer(this)), uintptr(dwConnection))
	return HRESULT(ret)
}

func (this *IOleAdviseHolder) EnumAdvise(ppenumAdvise **IEnumSTATDATA) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().EnumAdvise, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(ppenumAdvise)))
	return HRESULT(ret)
}

func (this *IOleAdviseHolder) SendOnRename(pmk *IMoniker) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SendOnRename, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pmk)))
	return HRESULT(ret)
}

func (this *IOleAdviseHolder) SendOnSave() HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SendOnSave, uintptr(unsafe.Pointer(this)))
	return HRESULT(ret)
}

func (this *IOleAdviseHolder) SendOnClose() HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SendOnClose, uintptr(unsafe.Pointer(this)))
	return HRESULT(ret)
}

// 0000011E-0000-0000-C000-000000000046
var IID_IOleCache = syscall.GUID{0x0000011E, 0x0000, 0x0000,
	[8]byte{0xC0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x46}}

type IOleCacheInterface interface {
	IUnknownInterface
	Cache(pformatetc *FORMATETC, advf uint32, pdwConnection *uint32) HRESULT
	Uncache(dwConnection uint32) HRESULT
	EnumCache(ppenumSTATDATA **IEnumSTATDATA) HRESULT
	InitCache(pDataObject *IDataObject) HRESULT
	SetData(pformatetc *FORMATETC, pmedium *STGMEDIUM, fRelease BOOL) HRESULT
}

type IOleCacheVtbl struct {
	IUnknownVtbl
	Cache     uintptr
	Uncache   uintptr
	EnumCache uintptr
	InitCache uintptr
	SetData   uintptr
}

type IOleCache struct {
	IUnknown
}

func (this *IOleCache) Vtbl() *IOleCacheVtbl {
	return (*IOleCacheVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IOleCache) Cache(pformatetc *FORMATETC, advf uint32, pdwConnection *uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Cache, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pformatetc)), uintptr(advf), uintptr(unsafe.Pointer(pdwConnection)))
	return HRESULT(ret)
}

func (this *IOleCache) Uncache(dwConnection uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Uncache, uintptr(unsafe.Pointer(this)), uintptr(dwConnection))
	return HRESULT(ret)
}

func (this *IOleCache) EnumCache(ppenumSTATDATA **IEnumSTATDATA) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().EnumCache, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(ppenumSTATDATA)))
	return HRESULT(ret)
}

func (this *IOleCache) InitCache(pDataObject *IDataObject) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().InitCache, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pDataObject)))
	return HRESULT(ret)
}

func (this *IOleCache) SetData(pformatetc *FORMATETC, pmedium *STGMEDIUM, fRelease BOOL) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetData, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pformatetc)), uintptr(unsafe.Pointer(pmedium)), uintptr(fRelease))
	return HRESULT(ret)
}

// 00000128-0000-0000-C000-000000000046
var IID_IOleCache2 = syscall.GUID{0x00000128, 0x0000, 0x0000,
	[8]byte{0xC0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x46}}

type IOleCache2Interface interface {
	IOleCacheInterface
	UpdateCache(pDataObject *IDataObject, grfUpdf UPDFCACHE_FLAGS, pReserved unsafe.Pointer) HRESULT
	DiscardCache(dwDiscardOptions uint32) HRESULT
}

type IOleCache2Vtbl struct {
	IOleCacheVtbl
	UpdateCache  uintptr
	DiscardCache uintptr
}

type IOleCache2 struct {
	IOleCache
}

func (this *IOleCache2) Vtbl() *IOleCache2Vtbl {
	return (*IOleCache2Vtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IOleCache2) UpdateCache(pDataObject *IDataObject, grfUpdf UPDFCACHE_FLAGS, pReserved unsafe.Pointer) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().UpdateCache, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pDataObject)), uintptr(grfUpdf), uintptr(pReserved))
	return HRESULT(ret)
}

func (this *IOleCache2) DiscardCache(dwDiscardOptions uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().DiscardCache, uintptr(unsafe.Pointer(this)), uintptr(dwDiscardOptions))
	return HRESULT(ret)
}

// 00000129-0000-0000-C000-000000000046
var IID_IOleCacheControl = syscall.GUID{0x00000129, 0x0000, 0x0000,
	[8]byte{0xC0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x46}}

type IOleCacheControlInterface interface {
	IUnknownInterface
	OnRun(pDataObject *IDataObject) HRESULT
	OnStop() HRESULT
}

type IOleCacheControlVtbl struct {
	IUnknownVtbl
	OnRun  uintptr
	OnStop uintptr
}

type IOleCacheControl struct {
	IUnknown
}

func (this *IOleCacheControl) Vtbl() *IOleCacheControlVtbl {
	return (*IOleCacheControlVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IOleCacheControl) OnRun(pDataObject *IDataObject) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().OnRun, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pDataObject)))
	return HRESULT(ret)
}

func (this *IOleCacheControl) OnStop() HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().OnStop, uintptr(unsafe.Pointer(this)))
	return HRESULT(ret)
}

// 0000011A-0000-0000-C000-000000000046
var IID_IParseDisplayName = syscall.GUID{0x0000011A, 0x0000, 0x0000,
	[8]byte{0xC0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x46}}

type IParseDisplayNameInterface interface {
	IUnknownInterface
	ParseDisplayName(pbc *IBindCtx, pszDisplayName PWSTR, pchEaten *uint32, ppmkOut **IMoniker) HRESULT
}

type IParseDisplayNameVtbl struct {
	IUnknownVtbl
	ParseDisplayName uintptr
}

type IParseDisplayName struct {
	IUnknown
}

func (this *IParseDisplayName) Vtbl() *IParseDisplayNameVtbl {
	return (*IParseDisplayNameVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IParseDisplayName) ParseDisplayName(pbc *IBindCtx, pszDisplayName PWSTR, pchEaten *uint32, ppmkOut **IMoniker) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().ParseDisplayName, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pbc)), uintptr(unsafe.Pointer(pszDisplayName)), uintptr(unsafe.Pointer(pchEaten)), uintptr(unsafe.Pointer(ppmkOut)))
	return HRESULT(ret)
}

// 0000011B-0000-0000-C000-000000000046
var IID_IOleContainer = syscall.GUID{0x0000011B, 0x0000, 0x0000,
	[8]byte{0xC0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x46}}

type IOleContainerInterface interface {
	IParseDisplayNameInterface
	EnumObjects(grfFlags uint32, ppenum **IEnumUnknown) HRESULT
	LockContainer(fLock BOOL) HRESULT
}

type IOleContainerVtbl struct {
	IParseDisplayNameVtbl
	EnumObjects   uintptr
	LockContainer uintptr
}

type IOleContainer struct {
	IParseDisplayName
}

func (this *IOleContainer) Vtbl() *IOleContainerVtbl {
	return (*IOleContainerVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IOleContainer) EnumObjects(grfFlags uint32, ppenum **IEnumUnknown) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().EnumObjects, uintptr(unsafe.Pointer(this)), uintptr(grfFlags), uintptr(unsafe.Pointer(ppenum)))
	return HRESULT(ret)
}

func (this *IOleContainer) LockContainer(fLock BOOL) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().LockContainer, uintptr(unsafe.Pointer(this)), uintptr(fLock))
	return HRESULT(ret)
}

// 00000118-0000-0000-C000-000000000046
var IID_IOleClientSite = syscall.GUID{0x00000118, 0x0000, 0x0000,
	[8]byte{0xC0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x46}}

type IOleClientSiteInterface interface {
	IUnknownInterface
	SaveObject() HRESULT
	GetMoniker(dwAssign uint32, dwWhichMoniker uint32, ppmk **IMoniker) HRESULT
	GetContainer(ppContainer **IOleContainer) HRESULT
	ShowObject() HRESULT
	OnShowWindow(fShow BOOL) HRESULT
	RequestNewObjectLayout() HRESULT
}

type IOleClientSiteVtbl struct {
	IUnknownVtbl
	SaveObject             uintptr
	GetMoniker             uintptr
	GetContainer           uintptr
	ShowObject             uintptr
	OnShowWindow           uintptr
	RequestNewObjectLayout uintptr
}

type IOleClientSite struct {
	IUnknown
}

func (this *IOleClientSite) Vtbl() *IOleClientSiteVtbl {
	return (*IOleClientSiteVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IOleClientSite) SaveObject() HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SaveObject, uintptr(unsafe.Pointer(this)))
	return HRESULT(ret)
}

func (this *IOleClientSite) GetMoniker(dwAssign uint32, dwWhichMoniker uint32, ppmk **IMoniker) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetMoniker, uintptr(unsafe.Pointer(this)), uintptr(dwAssign), uintptr(dwWhichMoniker), uintptr(unsafe.Pointer(ppmk)))
	return HRESULT(ret)
}

func (this *IOleClientSite) GetContainer(ppContainer **IOleContainer) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetContainer, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(ppContainer)))
	return HRESULT(ret)
}

func (this *IOleClientSite) ShowObject() HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().ShowObject, uintptr(unsafe.Pointer(this)))
	return HRESULT(ret)
}

func (this *IOleClientSite) OnShowWindow(fShow BOOL) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().OnShowWindow, uintptr(unsafe.Pointer(this)), uintptr(fShow))
	return HRESULT(ret)
}

func (this *IOleClientSite) RequestNewObjectLayout() HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().RequestNewObjectLayout, uintptr(unsafe.Pointer(this)))
	return HRESULT(ret)
}

// 00000112-0000-0000-C000-000000000046
var IID_IOleObject = syscall.GUID{0x00000112, 0x0000, 0x0000,
	[8]byte{0xC0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x46}}

type IOleObjectInterface interface {
	IUnknownInterface
	SetClientSite(pClientSite *IOleClientSite) HRESULT
	GetClientSite(ppClientSite **IOleClientSite) HRESULT
	SetHostNames(szContainerApp PWSTR, szContainerObj PWSTR) HRESULT
	Close(dwSaveOption uint32) HRESULT
	SetMoniker(dwWhichMoniker uint32, pmk *IMoniker) HRESULT
	GetMoniker(dwAssign uint32, dwWhichMoniker uint32, ppmk **IMoniker) HRESULT
	InitFromData(pDataObject *IDataObject, fCreation BOOL, dwReserved uint32) HRESULT
	GetClipboardData(dwReserved uint32, ppDataObject **IDataObject) HRESULT
	DoVerb(iVerb int32, lpmsg *MSG, pActiveSite *IOleClientSite, lindex int32, hwndParent HWND, lprcPosRect *RECT) HRESULT
	EnumVerbs(ppEnumOleVerb **IEnumOLEVERB) HRESULT
	Update() HRESULT
	IsUpToDate() HRESULT
	GetUserClassID(pClsid *syscall.GUID) HRESULT
	GetUserType(dwFormOfType uint32, pszUserType *PWSTR) HRESULT
	SetExtent(dwDrawAspect uint32, psizel *SIZE) HRESULT
	GetExtent(dwDrawAspect uint32, psizel *SIZE) HRESULT
	Advise(pAdvSink *IAdviseSink, pdwConnection *uint32) HRESULT
	Unadvise(dwConnection uint32) HRESULT
	EnumAdvise(ppenumAdvise **IEnumSTATDATA) HRESULT
	GetMiscStatus(dwAspect uint32, pdwStatus *uint32) HRESULT
	SetColorScheme(pLogpal *LOGPALETTE) HRESULT
}

type IOleObjectVtbl struct {
	IUnknownVtbl
	SetClientSite    uintptr
	GetClientSite    uintptr
	SetHostNames     uintptr
	Close            uintptr
	SetMoniker       uintptr
	GetMoniker       uintptr
	InitFromData     uintptr
	GetClipboardData uintptr
	DoVerb           uintptr
	EnumVerbs        uintptr
	Update           uintptr
	IsUpToDate       uintptr
	GetUserClassID   uintptr
	GetUserType      uintptr
	SetExtent        uintptr
	GetExtent        uintptr
	Advise           uintptr
	Unadvise         uintptr
	EnumAdvise       uintptr
	GetMiscStatus    uintptr
	SetColorScheme   uintptr
}

type IOleObject struct {
	IUnknown
}

func (this *IOleObject) Vtbl() *IOleObjectVtbl {
	return (*IOleObjectVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IOleObject) SetClientSite(pClientSite *IOleClientSite) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetClientSite, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pClientSite)))
	return HRESULT(ret)
}

func (this *IOleObject) GetClientSite(ppClientSite **IOleClientSite) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetClientSite, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(ppClientSite)))
	return HRESULT(ret)
}

func (this *IOleObject) SetHostNames(szContainerApp PWSTR, szContainerObj PWSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetHostNames, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(szContainerApp)), uintptr(unsafe.Pointer(szContainerObj)))
	return HRESULT(ret)
}

func (this *IOleObject) Close(dwSaveOption uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Close, uintptr(unsafe.Pointer(this)), uintptr(dwSaveOption))
	return HRESULT(ret)
}

func (this *IOleObject) SetMoniker(dwWhichMoniker uint32, pmk *IMoniker) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetMoniker, uintptr(unsafe.Pointer(this)), uintptr(dwWhichMoniker), uintptr(unsafe.Pointer(pmk)))
	return HRESULT(ret)
}

func (this *IOleObject) GetMoniker(dwAssign uint32, dwWhichMoniker uint32, ppmk **IMoniker) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetMoniker, uintptr(unsafe.Pointer(this)), uintptr(dwAssign), uintptr(dwWhichMoniker), uintptr(unsafe.Pointer(ppmk)))
	return HRESULT(ret)
}

func (this *IOleObject) InitFromData(pDataObject *IDataObject, fCreation BOOL, dwReserved uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().InitFromData, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pDataObject)), uintptr(fCreation), uintptr(dwReserved))
	return HRESULT(ret)
}

func (this *IOleObject) GetClipboardData(dwReserved uint32, ppDataObject **IDataObject) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetClipboardData, uintptr(unsafe.Pointer(this)), uintptr(dwReserved), uintptr(unsafe.Pointer(ppDataObject)))
	return HRESULT(ret)
}

func (this *IOleObject) DoVerb(iVerb int32, lpmsg *MSG, pActiveSite *IOleClientSite, lindex int32, hwndParent HWND, lprcPosRect *RECT) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().DoVerb, uintptr(unsafe.Pointer(this)), uintptr(iVerb), uintptr(unsafe.Pointer(lpmsg)), uintptr(unsafe.Pointer(pActiveSite)), uintptr(lindex), hwndParent, uintptr(unsafe.Pointer(lprcPosRect)))
	return HRESULT(ret)
}

func (this *IOleObject) EnumVerbs(ppEnumOleVerb **IEnumOLEVERB) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().EnumVerbs, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(ppEnumOleVerb)))
	return HRESULT(ret)
}

func (this *IOleObject) Update() HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Update, uintptr(unsafe.Pointer(this)))
	return HRESULT(ret)
}

func (this *IOleObject) IsUpToDate() HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().IsUpToDate, uintptr(unsafe.Pointer(this)))
	return HRESULT(ret)
}

func (this *IOleObject) GetUserClassID(pClsid *syscall.GUID) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetUserClassID, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pClsid)))
	return HRESULT(ret)
}

func (this *IOleObject) GetUserType(dwFormOfType uint32, pszUserType *PWSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetUserType, uintptr(unsafe.Pointer(this)), uintptr(dwFormOfType), uintptr(unsafe.Pointer(pszUserType)))
	return HRESULT(ret)
}

func (this *IOleObject) SetExtent(dwDrawAspect uint32, psizel *SIZE) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetExtent, uintptr(unsafe.Pointer(this)), uintptr(dwDrawAspect), uintptr(unsafe.Pointer(psizel)))
	return HRESULT(ret)
}

func (this *IOleObject) GetExtent(dwDrawAspect uint32, psizel *SIZE) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetExtent, uintptr(unsafe.Pointer(this)), uintptr(dwDrawAspect), uintptr(unsafe.Pointer(psizel)))
	return HRESULT(ret)
}

func (this *IOleObject) Advise(pAdvSink *IAdviseSink, pdwConnection *uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Advise, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pAdvSink)), uintptr(unsafe.Pointer(pdwConnection)))
	return HRESULT(ret)
}

func (this *IOleObject) Unadvise(dwConnection uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Unadvise, uintptr(unsafe.Pointer(this)), uintptr(dwConnection))
	return HRESULT(ret)
}

func (this *IOleObject) EnumAdvise(ppenumAdvise **IEnumSTATDATA) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().EnumAdvise, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(ppenumAdvise)))
	return HRESULT(ret)
}

func (this *IOleObject) GetMiscStatus(dwAspect uint32, pdwStatus *uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetMiscStatus, uintptr(unsafe.Pointer(this)), uintptr(dwAspect), uintptr(unsafe.Pointer(pdwStatus)))
	return HRESULT(ret)
}

func (this *IOleObject) SetColorScheme(pLogpal *LOGPALETTE) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetColorScheme, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pLogpal)))
	return HRESULT(ret)
}

// 00000114-0000-0000-C000-000000000046
var IID_IOleWindow = syscall.GUID{0x00000114, 0x0000, 0x0000,
	[8]byte{0xC0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x46}}

type IOleWindowInterface interface {
	IUnknownInterface
	GetWindow(phwnd *HWND) HRESULT
	ContextSensitiveHelp(fEnterMode BOOL) HRESULT
}

type IOleWindowVtbl struct {
	IUnknownVtbl
	GetWindow            uintptr
	ContextSensitiveHelp uintptr
}

type IOleWindow struct {
	IUnknown
}

func (this *IOleWindow) Vtbl() *IOleWindowVtbl {
	return (*IOleWindowVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IOleWindow) GetWindow(phwnd *HWND) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetWindow, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(phwnd)))
	return HRESULT(ret)
}

func (this *IOleWindow) ContextSensitiveHelp(fEnterMode BOOL) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().ContextSensitiveHelp, uintptr(unsafe.Pointer(this)), uintptr(fEnterMode))
	return HRESULT(ret)
}

// 0000011D-0000-0000-C000-000000000046
var IID_IOleLink = syscall.GUID{0x0000011D, 0x0000, 0x0000,
	[8]byte{0xC0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x46}}

type IOleLinkInterface interface {
	IUnknownInterface
	SetUpdateOptions(dwUpdateOpt uint32) HRESULT
	GetUpdateOptions(pdwUpdateOpt *uint32) HRESULT
	SetSourceMoniker(pmk *IMoniker, rclsid *syscall.GUID) HRESULT
	GetSourceMoniker(ppmk **IMoniker) HRESULT
	SetSourceDisplayName(pszStatusText PWSTR) HRESULT
	GetSourceDisplayName(ppszDisplayName *PWSTR) HRESULT
	BindToSource(bindflags uint32, pbc *IBindCtx) HRESULT
	BindIfRunning() HRESULT
	GetBoundSource(ppunk **IUnknown) HRESULT
	UnbindSource() HRESULT
	Update(pbc *IBindCtx) HRESULT
}

type IOleLinkVtbl struct {
	IUnknownVtbl
	SetUpdateOptions     uintptr
	GetUpdateOptions     uintptr
	SetSourceMoniker     uintptr
	GetSourceMoniker     uintptr
	SetSourceDisplayName uintptr
	GetSourceDisplayName uintptr
	BindToSource         uintptr
	BindIfRunning        uintptr
	GetBoundSource       uintptr
	UnbindSource         uintptr
	Update               uintptr
}

type IOleLink struct {
	IUnknown
}

func (this *IOleLink) Vtbl() *IOleLinkVtbl {
	return (*IOleLinkVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IOleLink) SetUpdateOptions(dwUpdateOpt uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetUpdateOptions, uintptr(unsafe.Pointer(this)), uintptr(dwUpdateOpt))
	return HRESULT(ret)
}

func (this *IOleLink) GetUpdateOptions(pdwUpdateOpt *uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetUpdateOptions, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pdwUpdateOpt)))
	return HRESULT(ret)
}

func (this *IOleLink) SetSourceMoniker(pmk *IMoniker, rclsid *syscall.GUID) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetSourceMoniker, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pmk)), uintptr(unsafe.Pointer(rclsid)))
	return HRESULT(ret)
}

func (this *IOleLink) GetSourceMoniker(ppmk **IMoniker) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetSourceMoniker, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(ppmk)))
	return HRESULT(ret)
}

func (this *IOleLink) SetSourceDisplayName(pszStatusText PWSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetSourceDisplayName, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pszStatusText)))
	return HRESULT(ret)
}

func (this *IOleLink) GetSourceDisplayName(ppszDisplayName *PWSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetSourceDisplayName, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(ppszDisplayName)))
	return HRESULT(ret)
}

func (this *IOleLink) BindToSource(bindflags uint32, pbc *IBindCtx) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().BindToSource, uintptr(unsafe.Pointer(this)), uintptr(bindflags), uintptr(unsafe.Pointer(pbc)))
	return HRESULT(ret)
}

func (this *IOleLink) BindIfRunning() HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().BindIfRunning, uintptr(unsafe.Pointer(this)))
	return HRESULT(ret)
}

func (this *IOleLink) GetBoundSource(ppunk **IUnknown) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetBoundSource, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(ppunk)))
	return HRESULT(ret)
}

func (this *IOleLink) UnbindSource() HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().UnbindSource, uintptr(unsafe.Pointer(this)))
	return HRESULT(ret)
}

func (this *IOleLink) Update(pbc *IBindCtx) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Update, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pbc)))
	return HRESULT(ret)
}

// 0000011C-0000-0000-C000-000000000046
var IID_IOleItemContainer = syscall.GUID{0x0000011C, 0x0000, 0x0000,
	[8]byte{0xC0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x46}}

type IOleItemContainerInterface interface {
	IOleContainerInterface
	GetObject(pszItem PWSTR, dwSpeedNeeded uint32, pbc *IBindCtx, riid *syscall.GUID, ppvObject unsafe.Pointer) HRESULT
	GetObjectStorage(pszItem PWSTR, pbc *IBindCtx, riid *syscall.GUID, ppvStorage unsafe.Pointer) HRESULT
	IsRunning(pszItem PWSTR) HRESULT
}

type IOleItemContainerVtbl struct {
	IOleContainerVtbl
	GetObject        uintptr
	GetObjectStorage uintptr
	IsRunning        uintptr
}

type IOleItemContainer struct {
	IOleContainer
}

func (this *IOleItemContainer) Vtbl() *IOleItemContainerVtbl {
	return (*IOleItemContainerVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IOleItemContainer) GetObject(pszItem PWSTR, dwSpeedNeeded uint32, pbc *IBindCtx, riid *syscall.GUID, ppvObject unsafe.Pointer) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetObject, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pszItem)), uintptr(dwSpeedNeeded), uintptr(unsafe.Pointer(pbc)), uintptr(unsafe.Pointer(riid)), uintptr(ppvObject))
	return HRESULT(ret)
}

func (this *IOleItemContainer) GetObjectStorage(pszItem PWSTR, pbc *IBindCtx, riid *syscall.GUID, ppvStorage unsafe.Pointer) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetObjectStorage, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pszItem)), uintptr(unsafe.Pointer(pbc)), uintptr(unsafe.Pointer(riid)), uintptr(ppvStorage))
	return HRESULT(ret)
}

func (this *IOleItemContainer) IsRunning(pszItem PWSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().IsRunning, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pszItem)))
	return HRESULT(ret)
}

// 00000115-0000-0000-C000-000000000046
var IID_IOleInPlaceUIWindow = syscall.GUID{0x00000115, 0x0000, 0x0000,
	[8]byte{0xC0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x46}}

type IOleInPlaceUIWindowInterface interface {
	IOleWindowInterface
	GetBorder(lprectBorder *RECT) HRESULT
	RequestBorderSpace(pborderwidths *RECT) HRESULT
	SetBorderSpace(pborderwidths *RECT) HRESULT
	SetActiveObject(pActiveObject *IOleInPlaceActiveObject, pszObjName PWSTR) HRESULT
}

type IOleInPlaceUIWindowVtbl struct {
	IOleWindowVtbl
	GetBorder          uintptr
	RequestBorderSpace uintptr
	SetBorderSpace     uintptr
	SetActiveObject    uintptr
}

type IOleInPlaceUIWindow struct {
	IOleWindow
}

func (this *IOleInPlaceUIWindow) Vtbl() *IOleInPlaceUIWindowVtbl {
	return (*IOleInPlaceUIWindowVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IOleInPlaceUIWindow) GetBorder(lprectBorder *RECT) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetBorder, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(lprectBorder)))
	return HRESULT(ret)
}

func (this *IOleInPlaceUIWindow) RequestBorderSpace(pborderwidths *RECT) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().RequestBorderSpace, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pborderwidths)))
	return HRESULT(ret)
}

func (this *IOleInPlaceUIWindow) SetBorderSpace(pborderwidths *RECT) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetBorderSpace, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pborderwidths)))
	return HRESULT(ret)
}

func (this *IOleInPlaceUIWindow) SetActiveObject(pActiveObject *IOleInPlaceActiveObject, pszObjName PWSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetActiveObject, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pActiveObject)), uintptr(unsafe.Pointer(pszObjName)))
	return HRESULT(ret)
}

// 00000117-0000-0000-C000-000000000046
var IID_IOleInPlaceActiveObject = syscall.GUID{0x00000117, 0x0000, 0x0000,
	[8]byte{0xC0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x46}}

type IOleInPlaceActiveObjectInterface interface {
	IOleWindowInterface
	TranslateAccelerator(lpmsg *MSG) HRESULT
	OnFrameWindowActivate(fActivate BOOL) HRESULT
	OnDocWindowActivate(fActivate BOOL) HRESULT
	ResizeBorder(prcBorder *RECT, pUIWindow *IOleInPlaceUIWindow, fFrameWindow BOOL) HRESULT
	EnableModeless(fEnable BOOL) HRESULT
}

type IOleInPlaceActiveObjectVtbl struct {
	IOleWindowVtbl
	TranslateAccelerator  uintptr
	OnFrameWindowActivate uintptr
	OnDocWindowActivate   uintptr
	ResizeBorder          uintptr
	EnableModeless        uintptr
}

type IOleInPlaceActiveObject struct {
	IOleWindow
}

func (this *IOleInPlaceActiveObject) Vtbl() *IOleInPlaceActiveObjectVtbl {
	return (*IOleInPlaceActiveObjectVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IOleInPlaceActiveObject) TranslateAccelerator(lpmsg *MSG) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().TranslateAccelerator, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(lpmsg)))
	return HRESULT(ret)
}

func (this *IOleInPlaceActiveObject) OnFrameWindowActivate(fActivate BOOL) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().OnFrameWindowActivate, uintptr(unsafe.Pointer(this)), uintptr(fActivate))
	return HRESULT(ret)
}

func (this *IOleInPlaceActiveObject) OnDocWindowActivate(fActivate BOOL) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().OnDocWindowActivate, uintptr(unsafe.Pointer(this)), uintptr(fActivate))
	return HRESULT(ret)
}

func (this *IOleInPlaceActiveObject) ResizeBorder(prcBorder *RECT, pUIWindow *IOleInPlaceUIWindow, fFrameWindow BOOL) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().ResizeBorder, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(prcBorder)), uintptr(unsafe.Pointer(pUIWindow)), uintptr(fFrameWindow))
	return HRESULT(ret)
}

func (this *IOleInPlaceActiveObject) EnableModeless(fEnable BOOL) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().EnableModeless, uintptr(unsafe.Pointer(this)), uintptr(fEnable))
	return HRESULT(ret)
}

// 00000116-0000-0000-C000-000000000046
var IID_IOleInPlaceFrame = syscall.GUID{0x00000116, 0x0000, 0x0000,
	[8]byte{0xC0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x46}}

type IOleInPlaceFrameInterface interface {
	IOleInPlaceUIWindowInterface
	InsertMenus(hmenuShared HMENU, lpMenuWidths *OleMenuGroupWidths) HRESULT
	SetMenu(hmenuShared HMENU, holemenu uintptr, hwndActiveObject HWND) HRESULT
	RemoveMenus(hmenuShared HMENU) HRESULT
	SetStatusText(pszStatusText PWSTR) HRESULT
	EnableModeless(fEnable BOOL) HRESULT
	TranslateAccelerator(lpmsg *MSG, wID uint16) HRESULT
}

type IOleInPlaceFrameVtbl struct {
	IOleInPlaceUIWindowVtbl
	InsertMenus          uintptr
	SetMenu              uintptr
	RemoveMenus          uintptr
	SetStatusText        uintptr
	EnableModeless       uintptr
	TranslateAccelerator uintptr
}

type IOleInPlaceFrame struct {
	IOleInPlaceUIWindow
}

func (this *IOleInPlaceFrame) Vtbl() *IOleInPlaceFrameVtbl {
	return (*IOleInPlaceFrameVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IOleInPlaceFrame) InsertMenus(hmenuShared HMENU, lpMenuWidths *OleMenuGroupWidths) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().InsertMenus, uintptr(unsafe.Pointer(this)), hmenuShared, uintptr(unsafe.Pointer(lpMenuWidths)))
	return HRESULT(ret)
}

func (this *IOleInPlaceFrame) SetMenu(hmenuShared HMENU, holemenu uintptr, hwndActiveObject HWND) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetMenu, uintptr(unsafe.Pointer(this)), hmenuShared, holemenu, hwndActiveObject)
	return HRESULT(ret)
}

func (this *IOleInPlaceFrame) RemoveMenus(hmenuShared HMENU) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().RemoveMenus, uintptr(unsafe.Pointer(this)), hmenuShared)
	return HRESULT(ret)
}

func (this *IOleInPlaceFrame) SetStatusText(pszStatusText PWSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetStatusText, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pszStatusText)))
	return HRESULT(ret)
}

func (this *IOleInPlaceFrame) EnableModeless(fEnable BOOL) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().EnableModeless, uintptr(unsafe.Pointer(this)), uintptr(fEnable))
	return HRESULT(ret)
}

func (this *IOleInPlaceFrame) TranslateAccelerator(lpmsg *MSG, wID uint16) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().TranslateAccelerator, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(lpmsg)), uintptr(wID))
	return HRESULT(ret)
}

// 00000113-0000-0000-C000-000000000046
var IID_IOleInPlaceObject = syscall.GUID{0x00000113, 0x0000, 0x0000,
	[8]byte{0xC0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x46}}

type IOleInPlaceObjectInterface interface {
	IOleWindowInterface
	InPlaceDeactivate() HRESULT
	UIDeactivate() HRESULT
	SetObjectRects(lprcPosRect *RECT, lprcClipRect *RECT) HRESULT
	ReactivateAndUndo() HRESULT
}

type IOleInPlaceObjectVtbl struct {
	IOleWindowVtbl
	InPlaceDeactivate uintptr
	UIDeactivate      uintptr
	SetObjectRects    uintptr
	ReactivateAndUndo uintptr
}

type IOleInPlaceObject struct {
	IOleWindow
}

func (this *IOleInPlaceObject) Vtbl() *IOleInPlaceObjectVtbl {
	return (*IOleInPlaceObjectVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IOleInPlaceObject) InPlaceDeactivate() HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().InPlaceDeactivate, uintptr(unsafe.Pointer(this)))
	return HRESULT(ret)
}

func (this *IOleInPlaceObject) UIDeactivate() HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().UIDeactivate, uintptr(unsafe.Pointer(this)))
	return HRESULT(ret)
}

func (this *IOleInPlaceObject) SetObjectRects(lprcPosRect *RECT, lprcClipRect *RECT) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetObjectRects, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(lprcPosRect)), uintptr(unsafe.Pointer(lprcClipRect)))
	return HRESULT(ret)
}

func (this *IOleInPlaceObject) ReactivateAndUndo() HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().ReactivateAndUndo, uintptr(unsafe.Pointer(this)))
	return HRESULT(ret)
}

// 00000119-0000-0000-C000-000000000046
var IID_IOleInPlaceSite = syscall.GUID{0x00000119, 0x0000, 0x0000,
	[8]byte{0xC0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x46}}

type IOleInPlaceSiteInterface interface {
	IOleWindowInterface
	CanInPlaceActivate() HRESULT
	OnInPlaceActivate() HRESULT
	OnUIActivate() HRESULT
	GetWindowContext(ppFrame **IOleInPlaceFrame, ppDoc **IOleInPlaceUIWindow, lprcPosRect *RECT, lprcClipRect *RECT, lpFrameInfo *OIFI) HRESULT
	Scroll(scrollExtant SIZE) HRESULT
	OnUIDeactivate(fUndoable BOOL) HRESULT
	OnInPlaceDeactivate() HRESULT
	DiscardUndoState() HRESULT
	DeactivateAndUndo() HRESULT
	OnPosRectChange(lprcPosRect *RECT) HRESULT
}

type IOleInPlaceSiteVtbl struct {
	IOleWindowVtbl
	CanInPlaceActivate  uintptr
	OnInPlaceActivate   uintptr
	OnUIActivate        uintptr
	GetWindowContext    uintptr
	Scroll              uintptr
	OnUIDeactivate      uintptr
	OnInPlaceDeactivate uintptr
	DiscardUndoState    uintptr
	DeactivateAndUndo   uintptr
	OnPosRectChange     uintptr
}

type IOleInPlaceSite struct {
	IOleWindow
}

func (this *IOleInPlaceSite) Vtbl() *IOleInPlaceSiteVtbl {
	return (*IOleInPlaceSiteVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IOleInPlaceSite) CanInPlaceActivate() HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().CanInPlaceActivate, uintptr(unsafe.Pointer(this)))
	return HRESULT(ret)
}

func (this *IOleInPlaceSite) OnInPlaceActivate() HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().OnInPlaceActivate, uintptr(unsafe.Pointer(this)))
	return HRESULT(ret)
}

func (this *IOleInPlaceSite) OnUIActivate() HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().OnUIActivate, uintptr(unsafe.Pointer(this)))
	return HRESULT(ret)
}

func (this *IOleInPlaceSite) GetWindowContext(ppFrame **IOleInPlaceFrame, ppDoc **IOleInPlaceUIWindow, lprcPosRect *RECT, lprcClipRect *RECT, lpFrameInfo *OIFI) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetWindowContext, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(ppFrame)), uintptr(unsafe.Pointer(ppDoc)), uintptr(unsafe.Pointer(lprcPosRect)), uintptr(unsafe.Pointer(lprcClipRect)), uintptr(unsafe.Pointer(lpFrameInfo)))
	return HRESULT(ret)
}

func (this *IOleInPlaceSite) Scroll(scrollExtant SIZE) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Scroll, uintptr(unsafe.Pointer(this)), *(*uintptr)(unsafe.Pointer(&scrollExtant)))
	return HRESULT(ret)
}

func (this *IOleInPlaceSite) OnUIDeactivate(fUndoable BOOL) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().OnUIDeactivate, uintptr(unsafe.Pointer(this)), uintptr(fUndoable))
	return HRESULT(ret)
}

func (this *IOleInPlaceSite) OnInPlaceDeactivate() HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().OnInPlaceDeactivate, uintptr(unsafe.Pointer(this)))
	return HRESULT(ret)
}

func (this *IOleInPlaceSite) DiscardUndoState() HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().DiscardUndoState, uintptr(unsafe.Pointer(this)))
	return HRESULT(ret)
}

func (this *IOleInPlaceSite) DeactivateAndUndo() HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().DeactivateAndUndo, uintptr(unsafe.Pointer(this)))
	return HRESULT(ret)
}

func (this *IOleInPlaceSite) OnPosRectChange(lprcPosRect *RECT) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().OnPosRectChange, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(lprcPosRect)))
	return HRESULT(ret)
}

// 0000012A-0000-0000-C000-000000000046
var IID_IContinue = syscall.GUID{0x0000012A, 0x0000, 0x0000,
	[8]byte{0xC0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x46}}

type IContinueInterface interface {
	IUnknownInterface
	FContinue() HRESULT
}

type IContinueVtbl struct {
	IUnknownVtbl
	FContinue uintptr
}

type IContinue struct {
	IUnknown
}

func (this *IContinue) Vtbl() *IContinueVtbl {
	return (*IContinueVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IContinue) FContinue() HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().FContinue, uintptr(unsafe.Pointer(this)))
	return HRESULT(ret)
}

// 0000010D-0000-0000-C000-000000000046
var IID_IViewObject = syscall.GUID{0x0000010D, 0x0000, 0x0000,
	[8]byte{0xC0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x46}}

type IViewObjectInterface interface {
	IUnknownInterface
	Draw(dwDrawAspect uint32, lindex int32, pvAspect unsafe.Pointer, ptd *DVTARGETDEVICE, hdcTargetDev HDC, hdcDraw HDC, lprcBounds *RECTL, lprcWBounds *RECTL, pfnContinue uintptr, dwContinue uintptr) HRESULT
	GetColorSet(dwDrawAspect uint32, lindex int32, pvAspect unsafe.Pointer, ptd *DVTARGETDEVICE, hicTargetDev HDC, ppColorSet **LOGPALETTE) HRESULT
	Freeze(dwDrawAspect uint32, lindex int32, pvAspect unsafe.Pointer, pdwFreeze *uint32) HRESULT
	Unfreeze(dwFreeze uint32) HRESULT
	SetAdvise(aspects uint32, advf uint32, pAdvSink *IAdviseSink) HRESULT
	GetAdvise(pAspects *uint32, pAdvf *uint32, ppAdvSink **IAdviseSink) HRESULT
}

type IViewObjectVtbl struct {
	IUnknownVtbl
	Draw        uintptr
	GetColorSet uintptr
	Freeze      uintptr
	Unfreeze    uintptr
	SetAdvise   uintptr
	GetAdvise   uintptr
}

type IViewObject struct {
	IUnknown
}

func (this *IViewObject) Vtbl() *IViewObjectVtbl {
	return (*IViewObjectVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IViewObject) Draw(dwDrawAspect uint32, lindex int32, pvAspect unsafe.Pointer, ptd *DVTARGETDEVICE, hdcTargetDev HDC, hdcDraw HDC, lprcBounds *RECTL, lprcWBounds *RECTL, pfnContinue uintptr, dwContinue uintptr) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Draw, uintptr(unsafe.Pointer(this)), uintptr(dwDrawAspect), uintptr(lindex), uintptr(pvAspect), uintptr(unsafe.Pointer(ptd)), hdcTargetDev, hdcDraw, uintptr(unsafe.Pointer(lprcBounds)), uintptr(unsafe.Pointer(lprcWBounds)), pfnContinue, dwContinue)
	return HRESULT(ret)
}

func (this *IViewObject) GetColorSet(dwDrawAspect uint32, lindex int32, pvAspect unsafe.Pointer, ptd *DVTARGETDEVICE, hicTargetDev HDC, ppColorSet **LOGPALETTE) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetColorSet, uintptr(unsafe.Pointer(this)), uintptr(dwDrawAspect), uintptr(lindex), uintptr(pvAspect), uintptr(unsafe.Pointer(ptd)), hicTargetDev, uintptr(unsafe.Pointer(ppColorSet)))
	return HRESULT(ret)
}

func (this *IViewObject) Freeze(dwDrawAspect uint32, lindex int32, pvAspect unsafe.Pointer, pdwFreeze *uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Freeze, uintptr(unsafe.Pointer(this)), uintptr(dwDrawAspect), uintptr(lindex), uintptr(pvAspect), uintptr(unsafe.Pointer(pdwFreeze)))
	return HRESULT(ret)
}

func (this *IViewObject) Unfreeze(dwFreeze uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Unfreeze, uintptr(unsafe.Pointer(this)), uintptr(dwFreeze))
	return HRESULT(ret)
}

func (this *IViewObject) SetAdvise(aspects uint32, advf uint32, pAdvSink *IAdviseSink) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetAdvise, uintptr(unsafe.Pointer(this)), uintptr(aspects), uintptr(advf), uintptr(unsafe.Pointer(pAdvSink)))
	return HRESULT(ret)
}

func (this *IViewObject) GetAdvise(pAspects *uint32, pAdvf *uint32, ppAdvSink **IAdviseSink) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetAdvise, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pAspects)), uintptr(unsafe.Pointer(pAdvf)), uintptr(unsafe.Pointer(ppAdvSink)))
	return HRESULT(ret)
}

// 00000127-0000-0000-C000-000000000046
var IID_IViewObject2 = syscall.GUID{0x00000127, 0x0000, 0x0000,
	[8]byte{0xC0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x46}}

type IViewObject2Interface interface {
	IViewObjectInterface
	GetExtent(dwDrawAspect uint32, lindex int32, ptd *DVTARGETDEVICE, lpsizel *SIZE) HRESULT
}

type IViewObject2Vtbl struct {
	IViewObjectVtbl
	GetExtent uintptr
}

type IViewObject2 struct {
	IViewObject
}

func (this *IViewObject2) Vtbl() *IViewObject2Vtbl {
	return (*IViewObject2Vtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IViewObject2) GetExtent(dwDrawAspect uint32, lindex int32, ptd *DVTARGETDEVICE, lpsizel *SIZE) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetExtent, uintptr(unsafe.Pointer(this)), uintptr(dwDrawAspect), uintptr(lindex), uintptr(unsafe.Pointer(ptd)), uintptr(unsafe.Pointer(lpsizel)))
	return HRESULT(ret)
}

// 00000121-0000-0000-C000-000000000046
var IID_IDropSource = syscall.GUID{0x00000121, 0x0000, 0x0000,
	[8]byte{0xC0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x46}}

type IDropSourceInterface interface {
	IUnknownInterface
	QueryContinueDrag(fEscapePressed BOOL, grfKeyState uint32) HRESULT
	GiveFeedback(dwEffect uint32) HRESULT
}

type IDropSourceVtbl struct {
	IUnknownVtbl
	QueryContinueDrag uintptr
	GiveFeedback      uintptr
}

type IDropSource struct {
	IUnknown
}

func (this *IDropSource) Vtbl() *IDropSourceVtbl {
	return (*IDropSourceVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IDropSource) QueryContinueDrag(fEscapePressed BOOL, grfKeyState uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().QueryContinueDrag, uintptr(unsafe.Pointer(this)), uintptr(fEscapePressed), uintptr(grfKeyState))
	return HRESULT(ret)
}

func (this *IDropSource) GiveFeedback(dwEffect uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GiveFeedback, uintptr(unsafe.Pointer(this)), uintptr(dwEffect))
	return HRESULT(ret)
}

// 00000122-0000-0000-C000-000000000046
var IID_IDropTarget = syscall.GUID{0x00000122, 0x0000, 0x0000,
	[8]byte{0xC0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x46}}

type IDropTargetInterface interface {
	IUnknownInterface
	DragEnter(pDataObj *IDataObject, grfKeyState uint32, pt POINTL, pdwEffect *uint32) HRESULT
	DragOver(grfKeyState uint32, pt POINTL, pdwEffect *uint32) HRESULT
	DragLeave() HRESULT
	Drop(pDataObj *IDataObject, grfKeyState uint32, pt POINTL, pdwEffect *uint32) HRESULT
}

type IDropTargetVtbl struct {
	IUnknownVtbl
	DragEnter uintptr
	DragOver  uintptr
	DragLeave uintptr
	Drop      uintptr
}

type IDropTarget struct {
	IUnknown
}

func (this *IDropTarget) Vtbl() *IDropTargetVtbl {
	return (*IDropTargetVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IDropTarget) DragEnter(pDataObj *IDataObject, grfKeyState uint32, pt POINTL, pdwEffect *uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().DragEnter, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pDataObj)), uintptr(grfKeyState), *(*uintptr)(unsafe.Pointer(&pt)), uintptr(unsafe.Pointer(pdwEffect)))
	return HRESULT(ret)
}

func (this *IDropTarget) DragOver(grfKeyState uint32, pt POINTL, pdwEffect *uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().DragOver, uintptr(unsafe.Pointer(this)), uintptr(grfKeyState), *(*uintptr)(unsafe.Pointer(&pt)), uintptr(unsafe.Pointer(pdwEffect)))
	return HRESULT(ret)
}

func (this *IDropTarget) DragLeave() HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().DragLeave, uintptr(unsafe.Pointer(this)))
	return HRESULT(ret)
}

func (this *IDropTarget) Drop(pDataObj *IDataObject, grfKeyState uint32, pt POINTL, pdwEffect *uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Drop, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pDataObj)), uintptr(grfKeyState), *(*uintptr)(unsafe.Pointer(&pt)), uintptr(unsafe.Pointer(pdwEffect)))
	return HRESULT(ret)
}

// 0000012B-0000-0000-C000-000000000046
var IID_IDropSourceNotify = syscall.GUID{0x0000012B, 0x0000, 0x0000,
	[8]byte{0xC0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x46}}

type IDropSourceNotifyInterface interface {
	IUnknownInterface
	DragEnterTarget(hwndTarget HWND) HRESULT
	DragLeaveTarget() HRESULT
}

type IDropSourceNotifyVtbl struct {
	IUnknownVtbl
	DragEnterTarget uintptr
	DragLeaveTarget uintptr
}

type IDropSourceNotify struct {
	IUnknown
}

func (this *IDropSourceNotify) Vtbl() *IDropSourceNotifyVtbl {
	return (*IDropSourceNotifyVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IDropSourceNotify) DragEnterTarget(hwndTarget HWND) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().DragEnterTarget, uintptr(unsafe.Pointer(this)), hwndTarget)
	return HRESULT(ret)
}

func (this *IDropSourceNotify) DragLeaveTarget() HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().DragLeaveTarget, uintptr(unsafe.Pointer(this)))
	return HRESULT(ret)
}

// 390E3878-FD55-4E18-819D-4682081C0CFD
var IID_IEnterpriseDropTarget = syscall.GUID{0x390E3878, 0xFD55, 0x4E18,
	[8]byte{0x81, 0x9D, 0x46, 0x82, 0x08, 0x1C, 0x0C, 0xFD}}

type IEnterpriseDropTargetInterface interface {
	IUnknownInterface
	SetDropSourceEnterpriseId(identity PWSTR) HRESULT
	IsEvaluatingEdpPolicy(value *BOOL) HRESULT
}

type IEnterpriseDropTargetVtbl struct {
	IUnknownVtbl
	SetDropSourceEnterpriseId uintptr
	IsEvaluatingEdpPolicy     uintptr
}

type IEnterpriseDropTarget struct {
	IUnknown
}

func (this *IEnterpriseDropTarget) Vtbl() *IEnterpriseDropTargetVtbl {
	return (*IEnterpriseDropTargetVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IEnterpriseDropTarget) SetDropSourceEnterpriseId(identity PWSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetDropSourceEnterpriseId, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(identity)))
	return HRESULT(ret)
}

func (this *IEnterpriseDropTarget) IsEvaluatingEdpPolicy(value *BOOL) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().IsEvaluatingEdpPolicy, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(value)))
	return HRESULT(ret)
}

// 00000104-0000-0000-C000-000000000046
var IID_IEnumOLEVERB = syscall.GUID{0x00000104, 0x0000, 0x0000,
	[8]byte{0xC0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x46}}

type IEnumOLEVERBInterface interface {
	IUnknownInterface
	Next(celt uint32, rgelt *OLEVERB, pceltFetched *uint32) HRESULT
	Skip(celt uint32) HRESULT
	Reset() HRESULT
	Clone(ppenum **IEnumOLEVERB) HRESULT
}

type IEnumOLEVERBVtbl struct {
	IUnknownVtbl
	Next  uintptr
	Skip  uintptr
	Reset uintptr
	Clone uintptr
}

type IEnumOLEVERB struct {
	IUnknown
}

func (this *IEnumOLEVERB) Vtbl() *IEnumOLEVERBVtbl {
	return (*IEnumOLEVERBVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IEnumOLEVERB) Next(celt uint32, rgelt *OLEVERB, pceltFetched *uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Next, uintptr(unsafe.Pointer(this)), uintptr(celt), uintptr(unsafe.Pointer(rgelt)), uintptr(unsafe.Pointer(pceltFetched)))
	return HRESULT(ret)
}

func (this *IEnumOLEVERB) Skip(celt uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Skip, uintptr(unsafe.Pointer(this)), uintptr(celt))
	return HRESULT(ret)
}

func (this *IEnumOLEVERB) Reset() HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Reset, uintptr(unsafe.Pointer(this)))
	return HRESULT(ret)
}

func (this *IEnumOLEVERB) Clone(ppenum **IEnumOLEVERB) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Clone, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(ppenum)))
	return HRESULT(ret)
}

// B196B28F-BAB4-101A-B69C-00AA00341D07
var IID_IClassFactory2 = syscall.GUID{0xB196B28F, 0xBAB4, 0x101A,
	[8]byte{0xB6, 0x9C, 0x00, 0xAA, 0x00, 0x34, 0x1D, 0x07}}

type IClassFactory2Interface interface {
	IClassFactoryInterface
	GetLicInfo(pLicInfo *LICINFO) HRESULT
	RequestLicKey(dwReserved uint32, pBstrKey *BSTR) HRESULT
	CreateInstanceLic(pUnkOuter *IUnknown, pUnkReserved *IUnknown, riid *syscall.GUID, bstrKey BSTR, ppvObj unsafe.Pointer) HRESULT
}

type IClassFactory2Vtbl struct {
	IClassFactoryVtbl
	GetLicInfo        uintptr
	RequestLicKey     uintptr
	CreateInstanceLic uintptr
}

type IClassFactory2 struct {
	IClassFactory
}

func (this *IClassFactory2) Vtbl() *IClassFactory2Vtbl {
	return (*IClassFactory2Vtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IClassFactory2) GetLicInfo(pLicInfo *LICINFO) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetLicInfo, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pLicInfo)))
	return HRESULT(ret)
}

func (this *IClassFactory2) RequestLicKey(dwReserved uint32, pBstrKey *BSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().RequestLicKey, uintptr(unsafe.Pointer(this)), uintptr(dwReserved), uintptr(unsafe.Pointer(pBstrKey)))
	return HRESULT(ret)
}

func (this *IClassFactory2) CreateInstanceLic(pUnkOuter *IUnknown, pUnkReserved *IUnknown, riid *syscall.GUID, bstrKey BSTR, ppvObj unsafe.Pointer) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().CreateInstanceLic, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pUnkOuter)), uintptr(unsafe.Pointer(pUnkReserved)), uintptr(unsafe.Pointer(riid)), uintptr(unsafe.Pointer(bstrKey)), uintptr(ppvObj))
	return HRESULT(ret)
}

// B196B283-BAB4-101A-B69C-00AA00341D07
var IID_IProvideClassInfo = syscall.GUID{0xB196B283, 0xBAB4, 0x101A,
	[8]byte{0xB6, 0x9C, 0x00, 0xAA, 0x00, 0x34, 0x1D, 0x07}}

type IProvideClassInfoInterface interface {
	IUnknownInterface
	GetClassInfo(ppTI **ITypeInfo) HRESULT
}

type IProvideClassInfoVtbl struct {
	IUnknownVtbl
	GetClassInfo uintptr
}

type IProvideClassInfo struct {
	IUnknown
}

func (this *IProvideClassInfo) Vtbl() *IProvideClassInfoVtbl {
	return (*IProvideClassInfoVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IProvideClassInfo) GetClassInfo(ppTI **ITypeInfo) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetClassInfo, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(ppTI)))
	return HRESULT(ret)
}

// A6BC3AC0-DBAA-11CE-9DE3-00AA004BB851
var IID_IProvideClassInfo2 = syscall.GUID{0xA6BC3AC0, 0xDBAA, 0x11CE,
	[8]byte{0x9D, 0xE3, 0x00, 0xAA, 0x00, 0x4B, 0xB8, 0x51}}

type IProvideClassInfo2Interface interface {
	IProvideClassInfoInterface
	GetGUID(dwGuidKind uint32, pGUID *syscall.GUID) HRESULT
}

type IProvideClassInfo2Vtbl struct {
	IProvideClassInfoVtbl
	GetGUID uintptr
}

type IProvideClassInfo2 struct {
	IProvideClassInfo
}

func (this *IProvideClassInfo2) Vtbl() *IProvideClassInfo2Vtbl {
	return (*IProvideClassInfo2Vtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IProvideClassInfo2) GetGUID(dwGuidKind uint32, pGUID *syscall.GUID) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetGUID, uintptr(unsafe.Pointer(this)), uintptr(dwGuidKind), uintptr(unsafe.Pointer(pGUID)))
	return HRESULT(ret)
}

// A7ABA9C1-8983-11CF-8F20-00805F2CD064
var IID_IProvideMultipleClassInfo = syscall.GUID{0xA7ABA9C1, 0x8983, 0x11CF,
	[8]byte{0x8F, 0x20, 0x00, 0x80, 0x5F, 0x2C, 0xD0, 0x64}}

type IProvideMultipleClassInfoInterface interface {
	IProvideClassInfo2Interface
	GetMultiTypeInfoCount(pcti *uint32) HRESULT
	GetInfoOfIndex(iti uint32, dwFlags MULTICLASSINFO_FLAGS, pptiCoClass **ITypeInfo, pdwTIFlags *uint32, pcdispidReserved *uint32, piidPrimary *syscall.GUID, piidSource *syscall.GUID) HRESULT
}

type IProvideMultipleClassInfoVtbl struct {
	IProvideClassInfo2Vtbl
	GetMultiTypeInfoCount uintptr
	GetInfoOfIndex        uintptr
}

type IProvideMultipleClassInfo struct {
	IProvideClassInfo2
}

func (this *IProvideMultipleClassInfo) Vtbl() *IProvideMultipleClassInfoVtbl {
	return (*IProvideMultipleClassInfoVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IProvideMultipleClassInfo) GetMultiTypeInfoCount(pcti *uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetMultiTypeInfoCount, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pcti)))
	return HRESULT(ret)
}

func (this *IProvideMultipleClassInfo) GetInfoOfIndex(iti uint32, dwFlags MULTICLASSINFO_FLAGS, pptiCoClass **ITypeInfo, pdwTIFlags *uint32, pcdispidReserved *uint32, piidPrimary *syscall.GUID, piidSource *syscall.GUID) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetInfoOfIndex, uintptr(unsafe.Pointer(this)), uintptr(iti), uintptr(dwFlags), uintptr(unsafe.Pointer(pptiCoClass)), uintptr(unsafe.Pointer(pdwTIFlags)), uintptr(unsafe.Pointer(pcdispidReserved)), uintptr(unsafe.Pointer(piidPrimary)), uintptr(unsafe.Pointer(piidSource)))
	return HRESULT(ret)
}

// B196B288-BAB4-101A-B69C-00AA00341D07
var IID_IOleControl = syscall.GUID{0xB196B288, 0xBAB4, 0x101A,
	[8]byte{0xB6, 0x9C, 0x00, 0xAA, 0x00, 0x34, 0x1D, 0x07}}

type IOleControlInterface interface {
	IUnknownInterface
	GetControlInfo(pCI *CONTROLINFO) HRESULT
	OnMnemonic(pMsg *MSG) HRESULT
	OnAmbientPropertyChange(dispID int32) HRESULT
	FreezeEvents(bFreeze BOOL) HRESULT
}

type IOleControlVtbl struct {
	IUnknownVtbl
	GetControlInfo          uintptr
	OnMnemonic              uintptr
	OnAmbientPropertyChange uintptr
	FreezeEvents            uintptr
}

type IOleControl struct {
	IUnknown
}

func (this *IOleControl) Vtbl() *IOleControlVtbl {
	return (*IOleControlVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IOleControl) GetControlInfo(pCI *CONTROLINFO) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetControlInfo, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pCI)))
	return HRESULT(ret)
}

func (this *IOleControl) OnMnemonic(pMsg *MSG) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().OnMnemonic, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pMsg)))
	return HRESULT(ret)
}

func (this *IOleControl) OnAmbientPropertyChange(dispID int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().OnAmbientPropertyChange, uintptr(unsafe.Pointer(this)), uintptr(dispID))
	return HRESULT(ret)
}

func (this *IOleControl) FreezeEvents(bFreeze BOOL) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().FreezeEvents, uintptr(unsafe.Pointer(this)), uintptr(bFreeze))
	return HRESULT(ret)
}

// B196B289-BAB4-101A-B69C-00AA00341D07
var IID_IOleControlSite = syscall.GUID{0xB196B289, 0xBAB4, 0x101A,
	[8]byte{0xB6, 0x9C, 0x00, 0xAA, 0x00, 0x34, 0x1D, 0x07}}

type IOleControlSiteInterface interface {
	IUnknownInterface
	OnControlInfoChanged() HRESULT
	LockInPlaceActive(fLock BOOL) HRESULT
	GetExtendedControl(ppDisp **IDispatch) HRESULT
	TransformCoords(pPtlHimetric *POINTL, pPtfContainer *POINTF, dwFlags XFORMCOORDS) HRESULT
	TranslateAccelerator(pMsg *MSG, grfModifiers uint32) HRESULT
	OnFocus(fGotFocus BOOL) HRESULT
	ShowPropertyFrame() HRESULT
}

type IOleControlSiteVtbl struct {
	IUnknownVtbl
	OnControlInfoChanged uintptr
	LockInPlaceActive    uintptr
	GetExtendedControl   uintptr
	TransformCoords      uintptr
	TranslateAccelerator uintptr
	OnFocus              uintptr
	ShowPropertyFrame    uintptr
}

type IOleControlSite struct {
	IUnknown
}

func (this *IOleControlSite) Vtbl() *IOleControlSiteVtbl {
	return (*IOleControlSiteVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IOleControlSite) OnControlInfoChanged() HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().OnControlInfoChanged, uintptr(unsafe.Pointer(this)))
	return HRESULT(ret)
}

func (this *IOleControlSite) LockInPlaceActive(fLock BOOL) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().LockInPlaceActive, uintptr(unsafe.Pointer(this)), uintptr(fLock))
	return HRESULT(ret)
}

func (this *IOleControlSite) GetExtendedControl(ppDisp **IDispatch) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetExtendedControl, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(ppDisp)))
	return HRESULT(ret)
}

func (this *IOleControlSite) TransformCoords(pPtlHimetric *POINTL, pPtfContainer *POINTF, dwFlags XFORMCOORDS) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().TransformCoords, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pPtlHimetric)), uintptr(unsafe.Pointer(pPtfContainer)), uintptr(dwFlags))
	return HRESULT(ret)
}

func (this *IOleControlSite) TranslateAccelerator(pMsg *MSG, grfModifiers uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().TranslateAccelerator, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pMsg)), uintptr(grfModifiers))
	return HRESULT(ret)
}

func (this *IOleControlSite) OnFocus(fGotFocus BOOL) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().OnFocus, uintptr(unsafe.Pointer(this)), uintptr(fGotFocus))
	return HRESULT(ret)
}

func (this *IOleControlSite) ShowPropertyFrame() HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().ShowPropertyFrame, uintptr(unsafe.Pointer(this)))
	return HRESULT(ret)
}

// B196B28D-BAB4-101A-B69C-00AA00341D07
var IID_IPropertyPage = syscall.GUID{0xB196B28D, 0xBAB4, 0x101A,
	[8]byte{0xB6, 0x9C, 0x00, 0xAA, 0x00, 0x34, 0x1D, 0x07}}

type IPropertyPageInterface interface {
	IUnknownInterface
	SetPageSite(pPageSite *IPropertyPageSite) HRESULT
	Activate(hWndParent HWND, pRect *RECT, bModal BOOL) HRESULT
	Deactivate() HRESULT
	GetPageInfo(pPageInfo *PROPPAGEINFO) HRESULT
	SetObjects(cObjects uint32, ppUnk **IUnknown) HRESULT
	Show(nCmdShow uint32) HRESULT
	Move(pRect *RECT) HRESULT
	IsPageDirty() HRESULT
	Apply() HRESULT
	Help(pszHelpDir PWSTR) HRESULT
	TranslateAccelerator(pMsg *MSG) HRESULT
}

type IPropertyPageVtbl struct {
	IUnknownVtbl
	SetPageSite          uintptr
	Activate             uintptr
	Deactivate           uintptr
	GetPageInfo          uintptr
	SetObjects           uintptr
	Show                 uintptr
	Move                 uintptr
	IsPageDirty          uintptr
	Apply                uintptr
	Help                 uintptr
	TranslateAccelerator uintptr
}

type IPropertyPage struct {
	IUnknown
}

func (this *IPropertyPage) Vtbl() *IPropertyPageVtbl {
	return (*IPropertyPageVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IPropertyPage) SetPageSite(pPageSite *IPropertyPageSite) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetPageSite, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pPageSite)))
	return HRESULT(ret)
}

func (this *IPropertyPage) Activate(hWndParent HWND, pRect *RECT, bModal BOOL) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Activate, uintptr(unsafe.Pointer(this)), hWndParent, uintptr(unsafe.Pointer(pRect)), uintptr(bModal))
	return HRESULT(ret)
}

func (this *IPropertyPage) Deactivate() HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Deactivate, uintptr(unsafe.Pointer(this)))
	return HRESULT(ret)
}

func (this *IPropertyPage) GetPageInfo(pPageInfo *PROPPAGEINFO) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetPageInfo, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pPageInfo)))
	return HRESULT(ret)
}

func (this *IPropertyPage) SetObjects(cObjects uint32, ppUnk **IUnknown) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetObjects, uintptr(unsafe.Pointer(this)), uintptr(cObjects), uintptr(unsafe.Pointer(ppUnk)))
	return HRESULT(ret)
}

func (this *IPropertyPage) Show(nCmdShow uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Show, uintptr(unsafe.Pointer(this)), uintptr(nCmdShow))
	return HRESULT(ret)
}

func (this *IPropertyPage) Move(pRect *RECT) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Move, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pRect)))
	return HRESULT(ret)
}

func (this *IPropertyPage) IsPageDirty() HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().IsPageDirty, uintptr(unsafe.Pointer(this)))
	return HRESULT(ret)
}

func (this *IPropertyPage) Apply() HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Apply, uintptr(unsafe.Pointer(this)))
	return HRESULT(ret)
}

func (this *IPropertyPage) Help(pszHelpDir PWSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Help, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pszHelpDir)))
	return HRESULT(ret)
}

func (this *IPropertyPage) TranslateAccelerator(pMsg *MSG) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().TranslateAccelerator, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pMsg)))
	return HRESULT(ret)
}

// 01E44665-24AC-101B-84ED-08002B2EC713
var IID_IPropertyPage2 = syscall.GUID{0x01E44665, 0x24AC, 0x101B,
	[8]byte{0x84, 0xED, 0x08, 0x00, 0x2B, 0x2E, 0xC7, 0x13}}

type IPropertyPage2Interface interface {
	IPropertyPageInterface
	EditProperty(dispID int32) HRESULT
}

type IPropertyPage2Vtbl struct {
	IPropertyPageVtbl
	EditProperty uintptr
}

type IPropertyPage2 struct {
	IPropertyPage
}

func (this *IPropertyPage2) Vtbl() *IPropertyPage2Vtbl {
	return (*IPropertyPage2Vtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IPropertyPage2) EditProperty(dispID int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().EditProperty, uintptr(unsafe.Pointer(this)), uintptr(dispID))
	return HRESULT(ret)
}

// B196B28C-BAB4-101A-B69C-00AA00341D07
var IID_IPropertyPageSite = syscall.GUID{0xB196B28C, 0xBAB4, 0x101A,
	[8]byte{0xB6, 0x9C, 0x00, 0xAA, 0x00, 0x34, 0x1D, 0x07}}

type IPropertyPageSiteInterface interface {
	IUnknownInterface
	OnStatusChange(dwFlags PROPPAGESTATUS) HRESULT
	GetLocaleID(pLocaleID *uint32) HRESULT
	GetPageContainer(ppUnk **IUnknown) HRESULT
	TranslateAccelerator(pMsg *MSG) HRESULT
}

type IPropertyPageSiteVtbl struct {
	IUnknownVtbl
	OnStatusChange       uintptr
	GetLocaleID          uintptr
	GetPageContainer     uintptr
	TranslateAccelerator uintptr
}

type IPropertyPageSite struct {
	IUnknown
}

func (this *IPropertyPageSite) Vtbl() *IPropertyPageSiteVtbl {
	return (*IPropertyPageSiteVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IPropertyPageSite) OnStatusChange(dwFlags PROPPAGESTATUS) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().OnStatusChange, uintptr(unsafe.Pointer(this)), uintptr(dwFlags))
	return HRESULT(ret)
}

func (this *IPropertyPageSite) GetLocaleID(pLocaleID *uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetLocaleID, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pLocaleID)))
	return HRESULT(ret)
}

func (this *IPropertyPageSite) GetPageContainer(ppUnk **IUnknown) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetPageContainer, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(ppUnk)))
	return HRESULT(ret)
}

func (this *IPropertyPageSite) TranslateAccelerator(pMsg *MSG) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().TranslateAccelerator, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pMsg)))
	return HRESULT(ret)
}

// 9BFBBC02-EFF1-101A-84ED-00AA00341D07
var IID_IPropertyNotifySink = syscall.GUID{0x9BFBBC02, 0xEFF1, 0x101A,
	[8]byte{0x84, 0xED, 0x00, 0xAA, 0x00, 0x34, 0x1D, 0x07}}

type IPropertyNotifySinkInterface interface {
	IUnknownInterface
	OnChanged(dispID int32) HRESULT
	OnRequestEdit(dispID int32) HRESULT
}

type IPropertyNotifySinkVtbl struct {
	IUnknownVtbl
	OnChanged     uintptr
	OnRequestEdit uintptr
}

type IPropertyNotifySink struct {
	IUnknown
}

func (this *IPropertyNotifySink) Vtbl() *IPropertyNotifySinkVtbl {
	return (*IPropertyNotifySinkVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IPropertyNotifySink) OnChanged(dispID int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().OnChanged, uintptr(unsafe.Pointer(this)), uintptr(dispID))
	return HRESULT(ret)
}

func (this *IPropertyNotifySink) OnRequestEdit(dispID int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().OnRequestEdit, uintptr(unsafe.Pointer(this)), uintptr(dispID))
	return HRESULT(ret)
}

// B196B28B-BAB4-101A-B69C-00AA00341D07
var IID_ISpecifyPropertyPages = syscall.GUID{0xB196B28B, 0xBAB4, 0x101A,
	[8]byte{0xB6, 0x9C, 0x00, 0xAA, 0x00, 0x34, 0x1D, 0x07}}

type ISpecifyPropertyPagesInterface interface {
	IUnknownInterface
	GetPages(pPages *CAUUID) HRESULT
}

type ISpecifyPropertyPagesVtbl struct {
	IUnknownVtbl
	GetPages uintptr
}

type ISpecifyPropertyPages struct {
	IUnknown
}

func (this *ISpecifyPropertyPages) Vtbl() *ISpecifyPropertyPagesVtbl {
	return (*ISpecifyPropertyPagesVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *ISpecifyPropertyPages) GetPages(pPages *CAUUID) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetPages, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pPages)))
	return HRESULT(ret)
}

// 37D84F60-42CB-11CE-8135-00AA004BB851
var IID_IPersistPropertyBag = syscall.GUID{0x37D84F60, 0x42CB, 0x11CE,
	[8]byte{0x81, 0x35, 0x00, 0xAA, 0x00, 0x4B, 0xB8, 0x51}}

type IPersistPropertyBagInterface interface {
	IPersistInterface
	InitNew() HRESULT
	Load(pPropBag *IPropertyBag, pErrorLog *IErrorLog) HRESULT
	Save(pPropBag *IPropertyBag, fClearDirty BOOL, fSaveAllProperties BOOL) HRESULT
}

type IPersistPropertyBagVtbl struct {
	IPersistVtbl
	InitNew uintptr
	Load    uintptr
	Save    uintptr
}

type IPersistPropertyBag struct {
	IPersist
}

func (this *IPersistPropertyBag) Vtbl() *IPersistPropertyBagVtbl {
	return (*IPersistPropertyBagVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IPersistPropertyBag) InitNew() HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().InitNew, uintptr(unsafe.Pointer(this)))
	return HRESULT(ret)
}

func (this *IPersistPropertyBag) Load(pPropBag *IPropertyBag, pErrorLog *IErrorLog) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Load, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pPropBag)), uintptr(unsafe.Pointer(pErrorLog)))
	return HRESULT(ret)
}

func (this *IPersistPropertyBag) Save(pPropBag *IPropertyBag, fClearDirty BOOL, fSaveAllProperties BOOL) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Save, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pPropBag)), uintptr(fClearDirty), uintptr(fSaveAllProperties))
	return HRESULT(ret)
}

// 742B0E01-14E6-101B-914E-00AA00300CAB
var IID_ISimpleFrameSite = syscall.GUID{0x742B0E01, 0x14E6, 0x101B,
	[8]byte{0x91, 0x4E, 0x00, 0xAA, 0x00, 0x30, 0x0C, 0xAB}}

type ISimpleFrameSiteInterface interface {
	IUnknownInterface
	PreMessageFilter(hWnd HWND, msg uint32, wp WPARAM, lp LPARAM, plResult *LRESULT, pdwCookie *uint32) HRESULT
	PostMessageFilter(hWnd HWND, msg uint32, wp WPARAM, lp LPARAM, plResult *LRESULT, dwCookie uint32) HRESULT
}

type ISimpleFrameSiteVtbl struct {
	IUnknownVtbl
	PreMessageFilter  uintptr
	PostMessageFilter uintptr
}

type ISimpleFrameSite struct {
	IUnknown
}

func (this *ISimpleFrameSite) Vtbl() *ISimpleFrameSiteVtbl {
	return (*ISimpleFrameSiteVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *ISimpleFrameSite) PreMessageFilter(hWnd HWND, msg uint32, wp WPARAM, lp LPARAM, plResult *LRESULT, pdwCookie *uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().PreMessageFilter, uintptr(unsafe.Pointer(this)), hWnd, uintptr(msg), wp, lp, uintptr(unsafe.Pointer(plResult)), uintptr(unsafe.Pointer(pdwCookie)))
	return HRESULT(ret)
}

func (this *ISimpleFrameSite) PostMessageFilter(hWnd HWND, msg uint32, wp WPARAM, lp LPARAM, plResult *LRESULT, dwCookie uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().PostMessageFilter, uintptr(unsafe.Pointer(this)), hWnd, uintptr(msg), wp, lp, uintptr(unsafe.Pointer(plResult)), uintptr(dwCookie))
	return HRESULT(ret)
}

// BEF6E002-A874-101A-8BBA-00AA00300CAB
var IID_IFont = syscall.GUID{0xBEF6E002, 0xA874, 0x101A,
	[8]byte{0x8B, 0xBA, 0x00, 0xAA, 0x00, 0x30, 0x0C, 0xAB}}

type IFontInterface interface {
	IUnknownInterface
	Get_Name(pName *BSTR) HRESULT
	Put_Name(name BSTR) HRESULT
	Get_Size(pSize *CY) HRESULT
	Put_Size(size CY) HRESULT
	Get_Bold(pBold *BOOL) HRESULT
	Put_Bold(bold BOOL) HRESULT
	Get_Italic(pItalic *BOOL) HRESULT
	Put_Italic(italic BOOL) HRESULT
	Get_Underline(pUnderline *BOOL) HRESULT
	Put_Underline(underline BOOL) HRESULT
	Get_Strikethrough(pStrikethrough *BOOL) HRESULT
	Put_Strikethrough(strikethrough BOOL) HRESULT
	Get_Weight(pWeight *int16) HRESULT
	Put_Weight(weight int16) HRESULT
	Get_Charset(pCharset *int16) HRESULT
	Put_Charset(charset int16) HRESULT
	Get_hFont(phFont *HFONT) HRESULT
	Clone(ppFont **IFont) HRESULT
	IsEqual(pFontOther *IFont) HRESULT
	SetRatio(cyLogical int32, cyHimetric int32) HRESULT
	QueryTextMetrics(pTM *TEXTMETRICW) HRESULT
	AddRefHfont(hFont HFONT) HRESULT
	ReleaseHfont(hFont HFONT) HRESULT
	SetHdc(hDC HDC) HRESULT
}

type IFontVtbl struct {
	IUnknownVtbl
	Get_Name          uintptr
	Put_Name          uintptr
	Get_Size          uintptr
	Put_Size          uintptr
	Get_Bold          uintptr
	Put_Bold          uintptr
	Get_Italic        uintptr
	Put_Italic        uintptr
	Get_Underline     uintptr
	Put_Underline     uintptr
	Get_Strikethrough uintptr
	Put_Strikethrough uintptr
	Get_Weight        uintptr
	Put_Weight        uintptr
	Get_Charset       uintptr
	Put_Charset       uintptr
	Get_hFont         uintptr
	Clone             uintptr
	IsEqual           uintptr
	SetRatio          uintptr
	QueryTextMetrics  uintptr
	AddRefHfont       uintptr
	ReleaseHfont      uintptr
	SetHdc            uintptr
}

type IFont struct {
	IUnknown
}

func (this *IFont) Vtbl() *IFontVtbl {
	return (*IFontVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IFont) Get_Name(pName *BSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_Name, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pName)))
	return HRESULT(ret)
}

func (this *IFont) Put_Name(name BSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Put_Name, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(name)))
	return HRESULT(ret)
}

func (this *IFont) Get_Size(pSize *CY) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_Size, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pSize)))
	return HRESULT(ret)
}

func (this *IFont) Put_Size(size CY) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Put_Size, uintptr(unsafe.Pointer(this)), *(*uintptr)(unsafe.Pointer(&size)))
	return HRESULT(ret)
}

func (this *IFont) Get_Bold(pBold *BOOL) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_Bold, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pBold)))
	return HRESULT(ret)
}

func (this *IFont) Put_Bold(bold BOOL) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Put_Bold, uintptr(unsafe.Pointer(this)), uintptr(bold))
	return HRESULT(ret)
}

func (this *IFont) Get_Italic(pItalic *BOOL) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_Italic, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pItalic)))
	return HRESULT(ret)
}

func (this *IFont) Put_Italic(italic BOOL) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Put_Italic, uintptr(unsafe.Pointer(this)), uintptr(italic))
	return HRESULT(ret)
}

func (this *IFont) Get_Underline(pUnderline *BOOL) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_Underline, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pUnderline)))
	return HRESULT(ret)
}

func (this *IFont) Put_Underline(underline BOOL) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Put_Underline, uintptr(unsafe.Pointer(this)), uintptr(underline))
	return HRESULT(ret)
}

func (this *IFont) Get_Strikethrough(pStrikethrough *BOOL) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_Strikethrough, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pStrikethrough)))
	return HRESULT(ret)
}

func (this *IFont) Put_Strikethrough(strikethrough BOOL) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Put_Strikethrough, uintptr(unsafe.Pointer(this)), uintptr(strikethrough))
	return HRESULT(ret)
}

func (this *IFont) Get_Weight(pWeight *int16) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_Weight, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pWeight)))
	return HRESULT(ret)
}

func (this *IFont) Put_Weight(weight int16) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Put_Weight, uintptr(unsafe.Pointer(this)), uintptr(weight))
	return HRESULT(ret)
}

func (this *IFont) Get_Charset(pCharset *int16) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_Charset, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pCharset)))
	return HRESULT(ret)
}

func (this *IFont) Put_Charset(charset int16) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Put_Charset, uintptr(unsafe.Pointer(this)), uintptr(charset))
	return HRESULT(ret)
}

func (this *IFont) Get_hFont(phFont *HFONT) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_hFont, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(phFont)))
	return HRESULT(ret)
}

func (this *IFont) Clone(ppFont **IFont) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Clone, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(ppFont)))
	return HRESULT(ret)
}

func (this *IFont) IsEqual(pFontOther *IFont) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().IsEqual, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pFontOther)))
	return HRESULT(ret)
}

func (this *IFont) SetRatio(cyLogical int32, cyHimetric int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetRatio, uintptr(unsafe.Pointer(this)), uintptr(cyLogical), uintptr(cyHimetric))
	return HRESULT(ret)
}

func (this *IFont) QueryTextMetrics(pTM *TEXTMETRICW) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().QueryTextMetrics, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pTM)))
	return HRESULT(ret)
}

func (this *IFont) AddRefHfont(hFont HFONT) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().AddRefHfont, uintptr(unsafe.Pointer(this)), hFont)
	return HRESULT(ret)
}

func (this *IFont) ReleaseHfont(hFont HFONT) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().ReleaseHfont, uintptr(unsafe.Pointer(this)), hFont)
	return HRESULT(ret)
}

func (this *IFont) SetHdc(hDC HDC) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetHdc, uintptr(unsafe.Pointer(this)), hDC)
	return HRESULT(ret)
}

// 7BF80980-BF32-101A-8BBB-00AA00300CAB
var IID_IPicture = syscall.GUID{0x7BF80980, 0xBF32, 0x101A,
	[8]byte{0x8B, 0xBB, 0x00, 0xAA, 0x00, 0x30, 0x0C, 0xAB}}

type IPictureInterface interface {
	IUnknownInterface
	Get_Handle(pHandle *uint32) HRESULT
	Get_hPal(phPal *uint32) HRESULT
	Get_Type(pType *int16) HRESULT
	Get_Width(pWidth *int32) HRESULT
	Get_Height(pHeight *int32) HRESULT
	Render(hDC HDC, x int32, y int32, cx int32, cy int32, xSrc int32, ySrc int32, cxSrc int32, cySrc int32, pRcWBounds *RECT) HRESULT
	Set_hPal(hPal uint32) HRESULT
	Get_CurDC(phDC *HDC) HRESULT
	SelectPicture(hDCIn HDC, phDCOut *HDC, phBmpOut *uint32) HRESULT
	Get_KeepOriginalFormat(pKeep *BOOL) HRESULT
	Put_KeepOriginalFormat(keep BOOL) HRESULT
	PictureChanged() HRESULT
	SaveAsFile(pStream *IStream, fSaveMemCopy BOOL, pCbSize *int32) HRESULT
	Get_Attributes(pDwAttr *uint32) HRESULT
}

type IPictureVtbl struct {
	IUnknownVtbl
	Get_Handle             uintptr
	Get_hPal               uintptr
	Get_Type               uintptr
	Get_Width              uintptr
	Get_Height             uintptr
	Render                 uintptr
	Set_hPal               uintptr
	Get_CurDC              uintptr
	SelectPicture          uintptr
	Get_KeepOriginalFormat uintptr
	Put_KeepOriginalFormat uintptr
	PictureChanged         uintptr
	SaveAsFile             uintptr
	Get_Attributes         uintptr
}

type IPicture struct {
	IUnknown
}

func (this *IPicture) Vtbl() *IPictureVtbl {
	return (*IPictureVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IPicture) Get_Handle(pHandle *uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_Handle, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pHandle)))
	return HRESULT(ret)
}

func (this *IPicture) Get_hPal(phPal *uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_hPal, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(phPal)))
	return HRESULT(ret)
}

func (this *IPicture) Get_Type(pType *int16) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_Type, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pType)))
	return HRESULT(ret)
}

func (this *IPicture) Get_Width(pWidth *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_Width, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pWidth)))
	return HRESULT(ret)
}

func (this *IPicture) Get_Height(pHeight *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_Height, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pHeight)))
	return HRESULT(ret)
}

func (this *IPicture) Render(hDC HDC, x int32, y int32, cx int32, cy int32, xSrc int32, ySrc int32, cxSrc int32, cySrc int32, pRcWBounds *RECT) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Render, uintptr(unsafe.Pointer(this)), hDC, uintptr(x), uintptr(y), uintptr(cx), uintptr(cy), uintptr(xSrc), uintptr(ySrc), uintptr(cxSrc), uintptr(cySrc), uintptr(unsafe.Pointer(pRcWBounds)))
	return HRESULT(ret)
}

func (this *IPicture) Set_hPal(hPal uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Set_hPal, uintptr(unsafe.Pointer(this)), uintptr(hPal))
	return HRESULT(ret)
}

func (this *IPicture) Get_CurDC(phDC *HDC) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CurDC, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(phDC)))
	return HRESULT(ret)
}

func (this *IPicture) SelectPicture(hDCIn HDC, phDCOut *HDC, phBmpOut *uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SelectPicture, uintptr(unsafe.Pointer(this)), hDCIn, uintptr(unsafe.Pointer(phDCOut)), uintptr(unsafe.Pointer(phBmpOut)))
	return HRESULT(ret)
}

func (this *IPicture) Get_KeepOriginalFormat(pKeep *BOOL) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_KeepOriginalFormat, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pKeep)))
	return HRESULT(ret)
}

func (this *IPicture) Put_KeepOriginalFormat(keep BOOL) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Put_KeepOriginalFormat, uintptr(unsafe.Pointer(this)), uintptr(keep))
	return HRESULT(ret)
}

func (this *IPicture) PictureChanged() HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().PictureChanged, uintptr(unsafe.Pointer(this)))
	return HRESULT(ret)
}

func (this *IPicture) SaveAsFile(pStream *IStream, fSaveMemCopy BOOL, pCbSize *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SaveAsFile, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pStream)), uintptr(fSaveMemCopy), uintptr(unsafe.Pointer(pCbSize)))
	return HRESULT(ret)
}

func (this *IPicture) Get_Attributes(pDwAttr *uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_Attributes, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pDwAttr)))
	return HRESULT(ret)
}

// F5185DD8-2012-4B0B-AAD9-F052C6BD482B
var IID_IPicture2 = syscall.GUID{0xF5185DD8, 0x2012, 0x4B0B,
	[8]byte{0xAA, 0xD9, 0xF0, 0x52, 0xC6, 0xBD, 0x48, 0x2B}}

type IPicture2Interface interface {
	IUnknownInterface
	Get_Handle(pHandle *uintptr) HRESULT
	Get_hPal(phPal *uintptr) HRESULT
	Get_Type(pType *int16) HRESULT
	Get_Width(pWidth *int32) HRESULT
	Get_Height(pHeight *int32) HRESULT
	Render(hDC HDC, x int32, y int32, cx int32, cy int32, xSrc int32, ySrc int32, cxSrc int32, cySrc int32, pRcWBounds *RECT) HRESULT
	Set_hPal(hPal uintptr) HRESULT
	Get_CurDC(phDC *HDC) HRESULT
	SelectPicture(hDCIn HDC, phDCOut *HDC, phBmpOut *uintptr) HRESULT
	Get_KeepOriginalFormat(pKeep *BOOL) HRESULT
	Put_KeepOriginalFormat(keep BOOL) HRESULT
	PictureChanged() HRESULT
	SaveAsFile(pStream *IStream, fSaveMemCopy BOOL, pCbSize *int32) HRESULT
	Get_Attributes(pDwAttr *uint32) HRESULT
}

type IPicture2Vtbl struct {
	IUnknownVtbl
	Get_Handle             uintptr
	Get_hPal               uintptr
	Get_Type               uintptr
	Get_Width              uintptr
	Get_Height             uintptr
	Render                 uintptr
	Set_hPal               uintptr
	Get_CurDC              uintptr
	SelectPicture          uintptr
	Get_KeepOriginalFormat uintptr
	Put_KeepOriginalFormat uintptr
	PictureChanged         uintptr
	SaveAsFile             uintptr
	Get_Attributes         uintptr
}

type IPicture2 struct {
	IUnknown
}

func (this *IPicture2) Vtbl() *IPicture2Vtbl {
	return (*IPicture2Vtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IPicture2) Get_Handle(pHandle *uintptr) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_Handle, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pHandle)))
	return HRESULT(ret)
}

func (this *IPicture2) Get_hPal(phPal *uintptr) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_hPal, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(phPal)))
	return HRESULT(ret)
}

func (this *IPicture2) Get_Type(pType *int16) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_Type, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pType)))
	return HRESULT(ret)
}

func (this *IPicture2) Get_Width(pWidth *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_Width, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pWidth)))
	return HRESULT(ret)
}

func (this *IPicture2) Get_Height(pHeight *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_Height, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pHeight)))
	return HRESULT(ret)
}

func (this *IPicture2) Render(hDC HDC, x int32, y int32, cx int32, cy int32, xSrc int32, ySrc int32, cxSrc int32, cySrc int32, pRcWBounds *RECT) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Render, uintptr(unsafe.Pointer(this)), hDC, uintptr(x), uintptr(y), uintptr(cx), uintptr(cy), uintptr(xSrc), uintptr(ySrc), uintptr(cxSrc), uintptr(cySrc), uintptr(unsafe.Pointer(pRcWBounds)))
	return HRESULT(ret)
}

func (this *IPicture2) Set_hPal(hPal uintptr) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Set_hPal, uintptr(unsafe.Pointer(this)), hPal)
	return HRESULT(ret)
}

func (this *IPicture2) Get_CurDC(phDC *HDC) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_CurDC, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(phDC)))
	return HRESULT(ret)
}

func (this *IPicture2) SelectPicture(hDCIn HDC, phDCOut *HDC, phBmpOut *uintptr) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SelectPicture, uintptr(unsafe.Pointer(this)), hDCIn, uintptr(unsafe.Pointer(phDCOut)), uintptr(unsafe.Pointer(phBmpOut)))
	return HRESULT(ret)
}

func (this *IPicture2) Get_KeepOriginalFormat(pKeep *BOOL) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_KeepOriginalFormat, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pKeep)))
	return HRESULT(ret)
}

func (this *IPicture2) Put_KeepOriginalFormat(keep BOOL) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Put_KeepOriginalFormat, uintptr(unsafe.Pointer(this)), uintptr(keep))
	return HRESULT(ret)
}

func (this *IPicture2) PictureChanged() HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().PictureChanged, uintptr(unsafe.Pointer(this)))
	return HRESULT(ret)
}

func (this *IPicture2) SaveAsFile(pStream *IStream, fSaveMemCopy BOOL, pCbSize *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SaveAsFile, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pStream)), uintptr(fSaveMemCopy), uintptr(unsafe.Pointer(pCbSize)))
	return HRESULT(ret)
}

func (this *IPicture2) Get_Attributes(pDwAttr *uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Get_Attributes, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pDwAttr)))
	return HRESULT(ret)
}

// 4EF6100A-AF88-11D0-9846-00C04FC29993
var IID_IFontEventsDisp = syscall.GUID{0x4EF6100A, 0xAF88, 0x11D0,
	[8]byte{0x98, 0x46, 0x00, 0xC0, 0x4F, 0xC2, 0x99, 0x93}}

type IFontEventsDispInterface interface {
	IDispatchInterface
}

type IFontEventsDispVtbl struct {
	IDispatchVtbl
}

type IFontEventsDisp struct {
	IDispatch
}

func (this *IFontEventsDisp) Vtbl() *IFontEventsDispVtbl {
	return (*IFontEventsDispVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

// BEF6E003-A874-101A-8BBA-00AA00300CAB
var IID_IFontDisp = syscall.GUID{0xBEF6E003, 0xA874, 0x101A,
	[8]byte{0x8B, 0xBA, 0x00, 0xAA, 0x00, 0x30, 0x0C, 0xAB}}

type IFontDispInterface interface {
	IDispatchInterface
}

type IFontDispVtbl struct {
	IDispatchVtbl
}

type IFontDisp struct {
	IDispatch
}

func (this *IFontDisp) Vtbl() *IFontDispVtbl {
	return (*IFontDispVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

// 7BF80981-BF32-101A-8BBB-00AA00300CAB
var IID_IPictureDisp = syscall.GUID{0x7BF80981, 0xBF32, 0x101A,
	[8]byte{0x8B, 0xBB, 0x00, 0xAA, 0x00, 0x30, 0x0C, 0xAB}}

type IPictureDispInterface interface {
	IDispatchInterface
}

type IPictureDispVtbl struct {
	IDispatchVtbl
}

type IPictureDisp struct {
	IDispatch
}

func (this *IPictureDisp) Vtbl() *IPictureDispVtbl {
	return (*IPictureDispVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

// 1C2056CC-5EF4-101B-8BC8-00AA003E3B29
var IID_IOleInPlaceObjectWindowless = syscall.GUID{0x1C2056CC, 0x5EF4, 0x101B,
	[8]byte{0x8B, 0xC8, 0x00, 0xAA, 0x00, 0x3E, 0x3B, 0x29}}

type IOleInPlaceObjectWindowlessInterface interface {
	IOleInPlaceObjectInterface
	OnWindowMessage(msg uint32, wParam WPARAM, lParam LPARAM, plResult *LRESULT) HRESULT
	GetDropTarget(ppDropTarget **IDropTarget) HRESULT
}

type IOleInPlaceObjectWindowlessVtbl struct {
	IOleInPlaceObjectVtbl
	OnWindowMessage uintptr
	GetDropTarget   uintptr
}

type IOleInPlaceObjectWindowless struct {
	IOleInPlaceObject
}

func (this *IOleInPlaceObjectWindowless) Vtbl() *IOleInPlaceObjectWindowlessVtbl {
	return (*IOleInPlaceObjectWindowlessVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IOleInPlaceObjectWindowless) OnWindowMessage(msg uint32, wParam WPARAM, lParam LPARAM, plResult *LRESULT) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().OnWindowMessage, uintptr(unsafe.Pointer(this)), uintptr(msg), wParam, lParam, uintptr(unsafe.Pointer(plResult)))
	return HRESULT(ret)
}

func (this *IOleInPlaceObjectWindowless) GetDropTarget(ppDropTarget **IDropTarget) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetDropTarget, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(ppDropTarget)))
	return HRESULT(ret)
}

// 9C2CAD80-3424-11CF-B670-00AA004CD6D8
var IID_IOleInPlaceSiteEx = syscall.GUID{0x9C2CAD80, 0x3424, 0x11CF,
	[8]byte{0xB6, 0x70, 0x00, 0xAA, 0x00, 0x4C, 0xD6, 0xD8}}

type IOleInPlaceSiteExInterface interface {
	IOleInPlaceSiteInterface
	OnInPlaceActivateEx(pfNoRedraw *BOOL, dwFlags uint32) HRESULT
	OnInPlaceDeactivateEx(fNoRedraw BOOL) HRESULT
	RequestUIActivate() HRESULT
}

type IOleInPlaceSiteExVtbl struct {
	IOleInPlaceSiteVtbl
	OnInPlaceActivateEx   uintptr
	OnInPlaceDeactivateEx uintptr
	RequestUIActivate     uintptr
}

type IOleInPlaceSiteEx struct {
	IOleInPlaceSite
}

func (this *IOleInPlaceSiteEx) Vtbl() *IOleInPlaceSiteExVtbl {
	return (*IOleInPlaceSiteExVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IOleInPlaceSiteEx) OnInPlaceActivateEx(pfNoRedraw *BOOL, dwFlags uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().OnInPlaceActivateEx, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pfNoRedraw)), uintptr(dwFlags))
	return HRESULT(ret)
}

func (this *IOleInPlaceSiteEx) OnInPlaceDeactivateEx(fNoRedraw BOOL) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().OnInPlaceDeactivateEx, uintptr(unsafe.Pointer(this)), uintptr(fNoRedraw))
	return HRESULT(ret)
}

func (this *IOleInPlaceSiteEx) RequestUIActivate() HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().RequestUIActivate, uintptr(unsafe.Pointer(this)))
	return HRESULT(ret)
}

// 922EADA0-3424-11CF-B670-00AA004CD6D8
var IID_IOleInPlaceSiteWindowless = syscall.GUID{0x922EADA0, 0x3424, 0x11CF,
	[8]byte{0xB6, 0x70, 0x00, 0xAA, 0x00, 0x4C, 0xD6, 0xD8}}

type IOleInPlaceSiteWindowlessInterface interface {
	IOleInPlaceSiteExInterface
	CanWindowlessActivate() HRESULT
	GetCapture() HRESULT
	SetCapture(fCapture BOOL) HRESULT
	GetFocus() HRESULT
	SetFocus(fFocus BOOL) HRESULT
	GetDC(pRect *RECT, grfFlags uint32, phDC *HDC) HRESULT
	ReleaseDC(hDC HDC) HRESULT
	InvalidateRect(pRect *RECT, fErase BOOL) HRESULT
	InvalidateRgn(hRGN HRGN, fErase BOOL) HRESULT
	ScrollRect(dx int32, dy int32, pRectScroll *RECT, pRectClip *RECT) HRESULT
	AdjustRect(prc *RECT) HRESULT
	OnDefWindowMessage(msg uint32, wParam WPARAM, lParam LPARAM, plResult *LRESULT) HRESULT
}

type IOleInPlaceSiteWindowlessVtbl struct {
	IOleInPlaceSiteExVtbl
	CanWindowlessActivate uintptr
	GetCapture            uintptr
	SetCapture            uintptr
	GetFocus              uintptr
	SetFocus              uintptr
	GetDC                 uintptr
	ReleaseDC             uintptr
	InvalidateRect        uintptr
	InvalidateRgn         uintptr
	ScrollRect            uintptr
	AdjustRect            uintptr
	OnDefWindowMessage    uintptr
}

type IOleInPlaceSiteWindowless struct {
	IOleInPlaceSiteEx
}

func (this *IOleInPlaceSiteWindowless) Vtbl() *IOleInPlaceSiteWindowlessVtbl {
	return (*IOleInPlaceSiteWindowlessVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IOleInPlaceSiteWindowless) CanWindowlessActivate() HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().CanWindowlessActivate, uintptr(unsafe.Pointer(this)))
	return HRESULT(ret)
}

func (this *IOleInPlaceSiteWindowless) GetCapture() HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetCapture, uintptr(unsafe.Pointer(this)))
	return HRESULT(ret)
}

func (this *IOleInPlaceSiteWindowless) SetCapture(fCapture BOOL) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetCapture, uintptr(unsafe.Pointer(this)), uintptr(fCapture))
	return HRESULT(ret)
}

func (this *IOleInPlaceSiteWindowless) GetFocus() HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetFocus, uintptr(unsafe.Pointer(this)))
	return HRESULT(ret)
}

func (this *IOleInPlaceSiteWindowless) SetFocus(fFocus BOOL) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetFocus, uintptr(unsafe.Pointer(this)), uintptr(fFocus))
	return HRESULT(ret)
}

func (this *IOleInPlaceSiteWindowless) GetDC(pRect *RECT, grfFlags uint32, phDC *HDC) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetDC, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pRect)), uintptr(grfFlags), uintptr(unsafe.Pointer(phDC)))
	return HRESULT(ret)
}

func (this *IOleInPlaceSiteWindowless) ReleaseDC(hDC HDC) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().ReleaseDC, uintptr(unsafe.Pointer(this)), hDC)
	return HRESULT(ret)
}

func (this *IOleInPlaceSiteWindowless) InvalidateRect(pRect *RECT, fErase BOOL) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().InvalidateRect, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pRect)), uintptr(fErase))
	return HRESULT(ret)
}

func (this *IOleInPlaceSiteWindowless) InvalidateRgn(hRGN HRGN, fErase BOOL) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().InvalidateRgn, uintptr(unsafe.Pointer(this)), hRGN, uintptr(fErase))
	return HRESULT(ret)
}

func (this *IOleInPlaceSiteWindowless) ScrollRect(dx int32, dy int32, pRectScroll *RECT, pRectClip *RECT) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().ScrollRect, uintptr(unsafe.Pointer(this)), uintptr(dx), uintptr(dy), uintptr(unsafe.Pointer(pRectScroll)), uintptr(unsafe.Pointer(pRectClip)))
	return HRESULT(ret)
}

func (this *IOleInPlaceSiteWindowless) AdjustRect(prc *RECT) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().AdjustRect, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(prc)))
	return HRESULT(ret)
}

func (this *IOleInPlaceSiteWindowless) OnDefWindowMessage(msg uint32, wParam WPARAM, lParam LPARAM, plResult *LRESULT) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().OnDefWindowMessage, uintptr(unsafe.Pointer(this)), uintptr(msg), wParam, lParam, uintptr(unsafe.Pointer(plResult)))
	return HRESULT(ret)
}

// 3AF24292-0C96-11CE-A0CF-00AA00600AB8
var IID_IViewObjectEx = syscall.GUID{0x3AF24292, 0x0C96, 0x11CE,
	[8]byte{0xA0, 0xCF, 0x00, 0xAA, 0x00, 0x60, 0x0A, 0xB8}}

type IViewObjectExInterface interface {
	IViewObject2Interface
	GetRect(dwAspect uint32, pRect *RECTL) HRESULT
	GetViewStatus(pdwStatus *uint32) HRESULT
	QueryHitPoint(dwAspect uint32, pRectBounds *RECT, ptlLoc POINT, lCloseHint int32, pHitResult *uint32) HRESULT
	QueryHitRect(dwAspect uint32, pRectBounds *RECT, pRectLoc *RECT, lCloseHint int32, pHitResult *uint32) HRESULT
	GetNaturalExtent(dwAspect DVASPECT, lindex int32, ptd *DVTARGETDEVICE, hicTargetDev HDC, pExtentInfo *ExtentInfo, pSizel *SIZE) HRESULT
}

type IViewObjectExVtbl struct {
	IViewObject2Vtbl
	GetRect          uintptr
	GetViewStatus    uintptr
	QueryHitPoint    uintptr
	QueryHitRect     uintptr
	GetNaturalExtent uintptr
}

type IViewObjectEx struct {
	IViewObject2
}

func (this *IViewObjectEx) Vtbl() *IViewObjectExVtbl {
	return (*IViewObjectExVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IViewObjectEx) GetRect(dwAspect uint32, pRect *RECTL) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetRect, uintptr(unsafe.Pointer(this)), uintptr(dwAspect), uintptr(unsafe.Pointer(pRect)))
	return HRESULT(ret)
}

func (this *IViewObjectEx) GetViewStatus(pdwStatus *uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetViewStatus, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pdwStatus)))
	return HRESULT(ret)
}

func (this *IViewObjectEx) QueryHitPoint(dwAspect uint32, pRectBounds *RECT, ptlLoc POINT, lCloseHint int32, pHitResult *uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().QueryHitPoint, uintptr(unsafe.Pointer(this)), uintptr(dwAspect), uintptr(unsafe.Pointer(pRectBounds)), *(*uintptr)(unsafe.Pointer(&ptlLoc)), uintptr(lCloseHint), uintptr(unsafe.Pointer(pHitResult)))
	return HRESULT(ret)
}

func (this *IViewObjectEx) QueryHitRect(dwAspect uint32, pRectBounds *RECT, pRectLoc *RECT, lCloseHint int32, pHitResult *uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().QueryHitRect, uintptr(unsafe.Pointer(this)), uintptr(dwAspect), uintptr(unsafe.Pointer(pRectBounds)), uintptr(unsafe.Pointer(pRectLoc)), uintptr(lCloseHint), uintptr(unsafe.Pointer(pHitResult)))
	return HRESULT(ret)
}

func (this *IViewObjectEx) GetNaturalExtent(dwAspect DVASPECT, lindex int32, ptd *DVTARGETDEVICE, hicTargetDev HDC, pExtentInfo *ExtentInfo, pSizel *SIZE) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetNaturalExtent, uintptr(unsafe.Pointer(this)), uintptr(dwAspect), uintptr(lindex), uintptr(unsafe.Pointer(ptd)), hicTargetDev, uintptr(unsafe.Pointer(pExtentInfo)), uintptr(unsafe.Pointer(pSizel)))
	return HRESULT(ret)
}

// 894AD3B0-EF97-11CE-9BC9-00AA00608E01
var IID_IOleUndoUnit = syscall.GUID{0x894AD3B0, 0xEF97, 0x11CE,
	[8]byte{0x9B, 0xC9, 0x00, 0xAA, 0x00, 0x60, 0x8E, 0x01}}

type IOleUndoUnitInterface interface {
	IUnknownInterface
	Do(pUndoManager *IOleUndoManager) HRESULT
	GetDescription(pBstr *BSTR) HRESULT
	GetUnitType(pClsid *syscall.GUID, plID *int32) HRESULT
	OnNextAdd() HRESULT
}

type IOleUndoUnitVtbl struct {
	IUnknownVtbl
	Do             uintptr
	GetDescription uintptr
	GetUnitType    uintptr
	OnNextAdd      uintptr
}

type IOleUndoUnit struct {
	IUnknown
}

func (this *IOleUndoUnit) Vtbl() *IOleUndoUnitVtbl {
	return (*IOleUndoUnitVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IOleUndoUnit) Do(pUndoManager *IOleUndoManager) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Do, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pUndoManager)))
	return HRESULT(ret)
}

func (this *IOleUndoUnit) GetDescription(pBstr *BSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetDescription, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pBstr)))
	return HRESULT(ret)
}

func (this *IOleUndoUnit) GetUnitType(pClsid *syscall.GUID, plID *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetUnitType, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pClsid)), uintptr(unsafe.Pointer(plID)))
	return HRESULT(ret)
}

func (this *IOleUndoUnit) OnNextAdd() HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().OnNextAdd, uintptr(unsafe.Pointer(this)))
	return HRESULT(ret)
}

// A1FAF330-EF97-11CE-9BC9-00AA00608E01
var IID_IOleParentUndoUnit = syscall.GUID{0xA1FAF330, 0xEF97, 0x11CE,
	[8]byte{0x9B, 0xC9, 0x00, 0xAA, 0x00, 0x60, 0x8E, 0x01}}

type IOleParentUndoUnitInterface interface {
	IOleUndoUnitInterface
	Open(pPUU *IOleParentUndoUnit) HRESULT
	Close(pPUU *IOleParentUndoUnit, fCommit BOOL) HRESULT
	Add(pUU *IOleUndoUnit) HRESULT
	FindUnit(pUU *IOleUndoUnit) HRESULT
	GetParentState(pdwState *uint32) HRESULT
}

type IOleParentUndoUnitVtbl struct {
	IOleUndoUnitVtbl
	Open           uintptr
	Close          uintptr
	Add            uintptr
	FindUnit       uintptr
	GetParentState uintptr
}

type IOleParentUndoUnit struct {
	IOleUndoUnit
}

func (this *IOleParentUndoUnit) Vtbl() *IOleParentUndoUnitVtbl {
	return (*IOleParentUndoUnitVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IOleParentUndoUnit) Open(pPUU *IOleParentUndoUnit) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Open, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pPUU)))
	return HRESULT(ret)
}

func (this *IOleParentUndoUnit) Close(pPUU *IOleParentUndoUnit, fCommit BOOL) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Close, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pPUU)), uintptr(fCommit))
	return HRESULT(ret)
}

func (this *IOleParentUndoUnit) Add(pUU *IOleUndoUnit) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Add, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pUU)))
	return HRESULT(ret)
}

func (this *IOleParentUndoUnit) FindUnit(pUU *IOleUndoUnit) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().FindUnit, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pUU)))
	return HRESULT(ret)
}

func (this *IOleParentUndoUnit) GetParentState(pdwState *uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetParentState, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pdwState)))
	return HRESULT(ret)
}

// B3E7C340-EF97-11CE-9BC9-00AA00608E01
var IID_IEnumOleUndoUnits = syscall.GUID{0xB3E7C340, 0xEF97, 0x11CE,
	[8]byte{0x9B, 0xC9, 0x00, 0xAA, 0x00, 0x60, 0x8E, 0x01}}

type IEnumOleUndoUnitsInterface interface {
	IUnknownInterface
	Next(cElt uint32, rgElt **IOleUndoUnit, pcEltFetched *uint32) HRESULT
	Skip(cElt uint32) HRESULT
	Reset() HRESULT
	Clone(ppEnum **IEnumOleUndoUnits) HRESULT
}

type IEnumOleUndoUnitsVtbl struct {
	IUnknownVtbl
	Next  uintptr
	Skip  uintptr
	Reset uintptr
	Clone uintptr
}

type IEnumOleUndoUnits struct {
	IUnknown
}

func (this *IEnumOleUndoUnits) Vtbl() *IEnumOleUndoUnitsVtbl {
	return (*IEnumOleUndoUnitsVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IEnumOleUndoUnits) Next(cElt uint32, rgElt **IOleUndoUnit, pcEltFetched *uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Next, uintptr(unsafe.Pointer(this)), uintptr(cElt), uintptr(unsafe.Pointer(rgElt)), uintptr(unsafe.Pointer(pcEltFetched)))
	return HRESULT(ret)
}

func (this *IEnumOleUndoUnits) Skip(cElt uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Skip, uintptr(unsafe.Pointer(this)), uintptr(cElt))
	return HRESULT(ret)
}

func (this *IEnumOleUndoUnits) Reset() HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Reset, uintptr(unsafe.Pointer(this)))
	return HRESULT(ret)
}

func (this *IEnumOleUndoUnits) Clone(ppEnum **IEnumOleUndoUnits) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Clone, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(ppEnum)))
	return HRESULT(ret)
}

// D001F200-EF97-11CE-9BC9-00AA00608E01
var IID_IOleUndoManager = syscall.GUID{0xD001F200, 0xEF97, 0x11CE,
	[8]byte{0x9B, 0xC9, 0x00, 0xAA, 0x00, 0x60, 0x8E, 0x01}}

type IOleUndoManagerInterface interface {
	IUnknownInterface
	Open(pPUU *IOleParentUndoUnit) HRESULT
	Close(pPUU *IOleParentUndoUnit, fCommit BOOL) HRESULT
	Add(pUU *IOleUndoUnit) HRESULT
	GetOpenParentState(pdwState *uint32) HRESULT
	DiscardFrom(pUU *IOleUndoUnit) HRESULT
	UndoTo(pUU *IOleUndoUnit) HRESULT
	RedoTo(pUU *IOleUndoUnit) HRESULT
	EnumUndoable(ppEnum **IEnumOleUndoUnits) HRESULT
	EnumRedoable(ppEnum **IEnumOleUndoUnits) HRESULT
	GetLastUndoDescription(pBstr *BSTR) HRESULT
	GetLastRedoDescription(pBstr *BSTR) HRESULT
	Enable(fEnable BOOL) HRESULT
}

type IOleUndoManagerVtbl struct {
	IUnknownVtbl
	Open                   uintptr
	Close                  uintptr
	Add                    uintptr
	GetOpenParentState     uintptr
	DiscardFrom            uintptr
	UndoTo                 uintptr
	RedoTo                 uintptr
	EnumUndoable           uintptr
	EnumRedoable           uintptr
	GetLastUndoDescription uintptr
	GetLastRedoDescription uintptr
	Enable                 uintptr
}

type IOleUndoManager struct {
	IUnknown
}

func (this *IOleUndoManager) Vtbl() *IOleUndoManagerVtbl {
	return (*IOleUndoManagerVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IOleUndoManager) Open(pPUU *IOleParentUndoUnit) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Open, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pPUU)))
	return HRESULT(ret)
}

func (this *IOleUndoManager) Close(pPUU *IOleParentUndoUnit, fCommit BOOL) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Close, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pPUU)), uintptr(fCommit))
	return HRESULT(ret)
}

func (this *IOleUndoManager) Add(pUU *IOleUndoUnit) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Add, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pUU)))
	return HRESULT(ret)
}

func (this *IOleUndoManager) GetOpenParentState(pdwState *uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetOpenParentState, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pdwState)))
	return HRESULT(ret)
}

func (this *IOleUndoManager) DiscardFrom(pUU *IOleUndoUnit) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().DiscardFrom, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pUU)))
	return HRESULT(ret)
}

func (this *IOleUndoManager) UndoTo(pUU *IOleUndoUnit) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().UndoTo, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pUU)))
	return HRESULT(ret)
}

func (this *IOleUndoManager) RedoTo(pUU *IOleUndoUnit) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().RedoTo, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pUU)))
	return HRESULT(ret)
}

func (this *IOleUndoManager) EnumUndoable(ppEnum **IEnumOleUndoUnits) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().EnumUndoable, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(ppEnum)))
	return HRESULT(ret)
}

func (this *IOleUndoManager) EnumRedoable(ppEnum **IEnumOleUndoUnits) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().EnumRedoable, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(ppEnum)))
	return HRESULT(ret)
}

func (this *IOleUndoManager) GetLastUndoDescription(pBstr *BSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetLastUndoDescription, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pBstr)))
	return HRESULT(ret)
}

func (this *IOleUndoManager) GetLastRedoDescription(pBstr *BSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetLastRedoDescription, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pBstr)))
	return HRESULT(ret)
}

func (this *IOleUndoManager) Enable(fEnable BOOL) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Enable, uintptr(unsafe.Pointer(this)), uintptr(fEnable))
	return HRESULT(ret)
}

// 55980BA0-35AA-11CF-B671-00AA004CD6D8
var IID_IPointerInactive = syscall.GUID{0x55980BA0, 0x35AA, 0x11CF,
	[8]byte{0xB6, 0x71, 0x00, 0xAA, 0x00, 0x4C, 0xD6, 0xD8}}

type IPointerInactiveInterface interface {
	IUnknownInterface
	GetActivationPolicy(pdwPolicy *uint32) HRESULT
	OnInactiveMouseMove(pRectBounds *RECT, x int32, y int32, grfKeyState uint32) HRESULT
	OnInactiveSetCursor(pRectBounds *RECT, x int32, y int32, dwMouseMsg uint32, fSetAlways BOOL) HRESULT
}

type IPointerInactiveVtbl struct {
	IUnknownVtbl
	GetActivationPolicy uintptr
	OnInactiveMouseMove uintptr
	OnInactiveSetCursor uintptr
}

type IPointerInactive struct {
	IUnknown
}

func (this *IPointerInactive) Vtbl() *IPointerInactiveVtbl {
	return (*IPointerInactiveVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IPointerInactive) GetActivationPolicy(pdwPolicy *uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetActivationPolicy, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pdwPolicy)))
	return HRESULT(ret)
}

func (this *IPointerInactive) OnInactiveMouseMove(pRectBounds *RECT, x int32, y int32, grfKeyState uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().OnInactiveMouseMove, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pRectBounds)), uintptr(x), uintptr(y), uintptr(grfKeyState))
	return HRESULT(ret)
}

func (this *IPointerInactive) OnInactiveSetCursor(pRectBounds *RECT, x int32, y int32, dwMouseMsg uint32, fSetAlways BOOL) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().OnInactiveSetCursor, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pRectBounds)), uintptr(x), uintptr(y), uintptr(dwMouseMsg), uintptr(fSetAlways))
	return HRESULT(ret)
}

// FC4801A3-2BA9-11CF-A229-00AA003D7352
var IID_IObjectWithSite = syscall.GUID{0xFC4801A3, 0x2BA9, 0x11CF,
	[8]byte{0xA2, 0x29, 0x00, 0xAA, 0x00, 0x3D, 0x73, 0x52}}

type IObjectWithSiteInterface interface {
	IUnknownInterface
	SetSite(pUnkSite *IUnknown) HRESULT
	GetSite(riid *syscall.GUID, ppvSite unsafe.Pointer) HRESULT
}

type IObjectWithSiteVtbl struct {
	IUnknownVtbl
	SetSite uintptr
	GetSite uintptr
}

type IObjectWithSite struct {
	IUnknown
}

func (this *IObjectWithSite) Vtbl() *IObjectWithSiteVtbl {
	return (*IObjectWithSiteVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IObjectWithSite) SetSite(pUnkSite *IUnknown) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetSite, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pUnkSite)))
	return HRESULT(ret)
}

func (this *IObjectWithSite) GetSite(riid *syscall.GUID, ppvSite unsafe.Pointer) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetSite, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(riid)), uintptr(ppvSite))
	return HRESULT(ret)
}

// 376BD3AA-3845-101B-84ED-08002B2EC713
var IID_IPerPropertyBrowsing = syscall.GUID{0x376BD3AA, 0x3845, 0x101B,
	[8]byte{0x84, 0xED, 0x08, 0x00, 0x2B, 0x2E, 0xC7, 0x13}}

type IPerPropertyBrowsingInterface interface {
	IUnknownInterface
	GetDisplayString(dispID int32, pBstr *BSTR) HRESULT
	MapPropertyToPage(dispID int32, pClsid *syscall.GUID) HRESULT
	GetPredefinedStrings(dispID int32, pCaStringsOut *CALPOLESTR, pCaCookiesOut *CADWORD) HRESULT
	GetPredefinedValue(dispID int32, dwCookie uint32, pVarOut *VARIANT) HRESULT
}

type IPerPropertyBrowsingVtbl struct {
	IUnknownVtbl
	GetDisplayString     uintptr
	MapPropertyToPage    uintptr
	GetPredefinedStrings uintptr
	GetPredefinedValue   uintptr
}

type IPerPropertyBrowsing struct {
	IUnknown
}

func (this *IPerPropertyBrowsing) Vtbl() *IPerPropertyBrowsingVtbl {
	return (*IPerPropertyBrowsingVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IPerPropertyBrowsing) GetDisplayString(dispID int32, pBstr *BSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetDisplayString, uintptr(unsafe.Pointer(this)), uintptr(dispID), uintptr(unsafe.Pointer(pBstr)))
	return HRESULT(ret)
}

func (this *IPerPropertyBrowsing) MapPropertyToPage(dispID int32, pClsid *syscall.GUID) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().MapPropertyToPage, uintptr(unsafe.Pointer(this)), uintptr(dispID), uintptr(unsafe.Pointer(pClsid)))
	return HRESULT(ret)
}

func (this *IPerPropertyBrowsing) GetPredefinedStrings(dispID int32, pCaStringsOut *CALPOLESTR, pCaCookiesOut *CADWORD) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetPredefinedStrings, uintptr(unsafe.Pointer(this)), uintptr(dispID), uintptr(unsafe.Pointer(pCaStringsOut)), uintptr(unsafe.Pointer(pCaCookiesOut)))
	return HRESULT(ret)
}

func (this *IPerPropertyBrowsing) GetPredefinedValue(dispID int32, dwCookie uint32, pVarOut *VARIANT) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetPredefinedValue, uintptr(unsafe.Pointer(this)), uintptr(dispID), uintptr(dwCookie), uintptr(unsafe.Pointer(pVarOut)))
	return HRESULT(ret)
}

// 22F55881-280B-11D0-A8A9-00A0C90C2004
var IID_IPersistPropertyBag2 = syscall.GUID{0x22F55881, 0x280B, 0x11D0,
	[8]byte{0xA8, 0xA9, 0x00, 0xA0, 0xC9, 0x0C, 0x20, 0x04}}

type IPersistPropertyBag2Interface interface {
	IPersistInterface
	InitNew() HRESULT
	Load(pPropBag *IPropertyBag2, pErrLog *IErrorLog) HRESULT
	Save(pPropBag *IPropertyBag2, fClearDirty BOOL, fSaveAllProperties BOOL) HRESULT
	IsDirty() HRESULT
}

type IPersistPropertyBag2Vtbl struct {
	IPersistVtbl
	InitNew uintptr
	Load    uintptr
	Save    uintptr
	IsDirty uintptr
}

type IPersistPropertyBag2 struct {
	IPersist
}

func (this *IPersistPropertyBag2) Vtbl() *IPersistPropertyBag2Vtbl {
	return (*IPersistPropertyBag2Vtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IPersistPropertyBag2) InitNew() HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().InitNew, uintptr(unsafe.Pointer(this)))
	return HRESULT(ret)
}

func (this *IPersistPropertyBag2) Load(pPropBag *IPropertyBag2, pErrLog *IErrorLog) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Load, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pPropBag)), uintptr(unsafe.Pointer(pErrLog)))
	return HRESULT(ret)
}

func (this *IPersistPropertyBag2) Save(pPropBag *IPropertyBag2, fClearDirty BOOL, fSaveAllProperties BOOL) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Save, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pPropBag)), uintptr(fClearDirty), uintptr(fSaveAllProperties))
	return HRESULT(ret)
}

func (this *IPersistPropertyBag2) IsDirty() HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().IsDirty, uintptr(unsafe.Pointer(this)))
	return HRESULT(ret)
}

// 3AF24290-0C96-11CE-A0CF-00AA00600AB8
var IID_IAdviseSinkEx = syscall.GUID{0x3AF24290, 0x0C96, 0x11CE,
	[8]byte{0xA0, 0xCF, 0x00, 0xAA, 0x00, 0x60, 0x0A, 0xB8}}

type IAdviseSinkExInterface interface {
	IAdviseSinkInterface
	OnViewStatusChange(dwViewStatus uint32)
}

type IAdviseSinkExVtbl struct {
	IAdviseSinkVtbl
	OnViewStatusChange uintptr
}

type IAdviseSinkEx struct {
	IAdviseSink
}

func (this *IAdviseSinkEx) Vtbl() *IAdviseSinkExVtbl {
	return (*IAdviseSinkExVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IAdviseSinkEx) OnViewStatusChange(dwViewStatus uint32) {
	_, _, _ = syscall.SyscallN(this.Vtbl().OnViewStatusChange, uintptr(unsafe.Pointer(this)), uintptr(dwViewStatus))
}

// CF51ED10-62FE-11CF-BF86-00A0C9034836
var IID_IQuickActivate = syscall.GUID{0xCF51ED10, 0x62FE, 0x11CF,
	[8]byte{0xBF, 0x86, 0x00, 0xA0, 0xC9, 0x03, 0x48, 0x36}}

type IQuickActivateInterface interface {
	IUnknownInterface
	QuickActivate(pQaContainer *QACONTAINER, pQaControl *QACONTROL) HRESULT
	SetContentExtent(pSizel *SIZE) HRESULT
	GetContentExtent(pSizel *SIZE) HRESULT
}

type IQuickActivateVtbl struct {
	IUnknownVtbl
	QuickActivate    uintptr
	SetContentExtent uintptr
	GetContentExtent uintptr
}

type IQuickActivate struct {
	IUnknown
}

func (this *IQuickActivate) Vtbl() *IQuickActivateVtbl {
	return (*IQuickActivateVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IQuickActivate) QuickActivate(pQaContainer *QACONTAINER, pQaControl *QACONTROL) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().QuickActivate, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pQaContainer)), uintptr(unsafe.Pointer(pQaControl)))
	return HRESULT(ret)
}

func (this *IQuickActivate) SetContentExtent(pSizel *SIZE) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetContentExtent, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pSizel)))
	return HRESULT(ret)
}

func (this *IQuickActivate) GetContentExtent(pSizel *SIZE) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetContentExtent, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pSizel)))
	return HRESULT(ret)
}

// 40A050A0-3C31-101B-A82E-08002B2B2337
var IID_IVBGetControl = syscall.GUID{0x40A050A0, 0x3C31, 0x101B,
	[8]byte{0xA8, 0x2E, 0x08, 0x00, 0x2B, 0x2B, 0x23, 0x37}}

type IVBGetControlInterface interface {
	IUnknownInterface
	EnumControls(dwOleContF OLECONTF, dwWhich ENUM_CONTROLS_WHICH_FLAGS, ppenumUnk **IEnumUnknown) HRESULT
}

type IVBGetControlVtbl struct {
	IUnknownVtbl
	EnumControls uintptr
}

type IVBGetControl struct {
	IUnknown
}

func (this *IVBGetControl) Vtbl() *IVBGetControlVtbl {
	return (*IVBGetControlVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IVBGetControl) EnumControls(dwOleContF OLECONTF, dwWhich ENUM_CONTROLS_WHICH_FLAGS, ppenumUnk **IEnumUnknown) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().EnumControls, uintptr(unsafe.Pointer(this)), uintptr(dwOleContF), uintptr(dwWhich), uintptr(unsafe.Pointer(ppenumUnk)))
	return HRESULT(ret)
}

// 8A701DA0-4FEB-101B-A82E-08002B2B2337
var IID_IGetOleObject = syscall.GUID{0x8A701DA0, 0x4FEB, 0x101B,
	[8]byte{0xA8, 0x2E, 0x08, 0x00, 0x2B, 0x2B, 0x23, 0x37}}

type IGetOleObjectInterface interface {
	IUnknownInterface
	GetOleObject(riid *syscall.GUID, ppvObj unsafe.Pointer) HRESULT
}

type IGetOleObjectVtbl struct {
	IUnknownVtbl
	GetOleObject uintptr
}

type IGetOleObject struct {
	IUnknown
}

func (this *IGetOleObject) Vtbl() *IGetOleObjectVtbl {
	return (*IGetOleObjectVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IGetOleObject) GetOleObject(riid *syscall.GUID, ppvObj unsafe.Pointer) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetOleObject, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(riid)), uintptr(ppvObj))
	return HRESULT(ret)
}

// 9849FD60-3768-101B-8D72-AE6164FFE3CF
var IID_IVBFormat = syscall.GUID{0x9849FD60, 0x3768, 0x101B,
	[8]byte{0x8D, 0x72, 0xAE, 0x61, 0x64, 0xFF, 0xE3, 0xCF}}

type IVBFormatInterface interface {
	IUnknownInterface
	Format(vData *VARIANT, bstrFormat BSTR, lpBuffer unsafe.Pointer, cb uint16, lcid int32, sFirstDayOfWeek int16, sFirstWeekOfYear uint16, rcb *uint16) HRESULT
}

type IVBFormatVtbl struct {
	IUnknownVtbl
	Format uintptr
}

type IVBFormat struct {
	IUnknown
}

func (this *IVBFormat) Vtbl() *IVBFormatVtbl {
	return (*IVBFormatVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IVBFormat) Format(vData *VARIANT, bstrFormat BSTR, lpBuffer unsafe.Pointer, cb uint16, lcid int32, sFirstDayOfWeek int16, sFirstWeekOfYear uint16, rcb *uint16) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Format, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(vData)), uintptr(unsafe.Pointer(bstrFormat)), uintptr(lpBuffer), uintptr(cb), uintptr(lcid), uintptr(sFirstDayOfWeek), uintptr(sFirstWeekOfYear), uintptr(unsafe.Pointer(rcb)))
	return HRESULT(ret)
}

// 91733A60-3F4C-101B-A3F6-00AA0034E4E9
var IID_IGetVBAObject = syscall.GUID{0x91733A60, 0x3F4C, 0x101B,
	[8]byte{0xA3, 0xF6, 0x00, 0xAA, 0x00, 0x34, 0xE4, 0xE9}}

type IGetVBAObjectInterface interface {
	IUnknownInterface
	GetObject(riid *syscall.GUID, ppvObj unsafe.Pointer, dwReserved uint32) HRESULT
}

type IGetVBAObjectVtbl struct {
	IUnknownVtbl
	GetObject uintptr
}

type IGetVBAObject struct {
	IUnknown
}

func (this *IGetVBAObject) Vtbl() *IGetVBAObjectVtbl {
	return (*IGetVBAObjectVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IGetVBAObject) GetObject(riid *syscall.GUID, ppvObj unsafe.Pointer, dwReserved uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetObject, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(riid)), uintptr(ppvObj), uintptr(dwReserved))
	return HRESULT(ret)
}

// B722BCC5-4E68-101B-A2BC-00AA00404770
var IID_IOleDocument = syscall.GUID{0xB722BCC5, 0x4E68, 0x101B,
	[8]byte{0xA2, 0xBC, 0x00, 0xAA, 0x00, 0x40, 0x47, 0x70}}

type IOleDocumentInterface interface {
	IUnknownInterface
	CreateView(pIPSite *IOleInPlaceSite, pstm *IStream, dwReserved uint32, ppView **IOleDocumentView) HRESULT
	GetDocMiscStatus(pdwStatus *uint32) HRESULT
	EnumViews(ppEnum **IEnumOleDocumentViews, ppView **IOleDocumentView) HRESULT
}

type IOleDocumentVtbl struct {
	IUnknownVtbl
	CreateView       uintptr
	GetDocMiscStatus uintptr
	EnumViews        uintptr
}

type IOleDocument struct {
	IUnknown
}

func (this *IOleDocument) Vtbl() *IOleDocumentVtbl {
	return (*IOleDocumentVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IOleDocument) CreateView(pIPSite *IOleInPlaceSite, pstm *IStream, dwReserved uint32, ppView **IOleDocumentView) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().CreateView, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pIPSite)), uintptr(unsafe.Pointer(pstm)), uintptr(dwReserved), uintptr(unsafe.Pointer(ppView)))
	return HRESULT(ret)
}

func (this *IOleDocument) GetDocMiscStatus(pdwStatus *uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetDocMiscStatus, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pdwStatus)))
	return HRESULT(ret)
}

func (this *IOleDocument) EnumViews(ppEnum **IEnumOleDocumentViews, ppView **IOleDocumentView) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().EnumViews, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(ppEnum)), uintptr(unsafe.Pointer(ppView)))
	return HRESULT(ret)
}

// B722BCC7-4E68-101B-A2BC-00AA00404770
var IID_IOleDocumentSite = syscall.GUID{0xB722BCC7, 0x4E68, 0x101B,
	[8]byte{0xA2, 0xBC, 0x00, 0xAA, 0x00, 0x40, 0x47, 0x70}}

type IOleDocumentSiteInterface interface {
	IUnknownInterface
	ActivateMe(pViewToActivate *IOleDocumentView) HRESULT
}

type IOleDocumentSiteVtbl struct {
	IUnknownVtbl
	ActivateMe uintptr
}

type IOleDocumentSite struct {
	IUnknown
}

func (this *IOleDocumentSite) Vtbl() *IOleDocumentSiteVtbl {
	return (*IOleDocumentSiteVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IOleDocumentSite) ActivateMe(pViewToActivate *IOleDocumentView) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().ActivateMe, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pViewToActivate)))
	return HRESULT(ret)
}

// B722BCC6-4E68-101B-A2BC-00AA00404770
var IID_IOleDocumentView = syscall.GUID{0xB722BCC6, 0x4E68, 0x101B,
	[8]byte{0xA2, 0xBC, 0x00, 0xAA, 0x00, 0x40, 0x47, 0x70}}

type IOleDocumentViewInterface interface {
	IUnknownInterface
	SetInPlaceSite(pIPSite *IOleInPlaceSite) HRESULT
	GetInPlaceSite(ppIPSite **IOleInPlaceSite) HRESULT
	GetDocument(ppunk **IUnknown) HRESULT
	SetRect(prcView *RECT) HRESULT
	GetRect(prcView *RECT) HRESULT
	SetRectComplex(prcView *RECT, prcHScroll *RECT, prcVScroll *RECT, prcSizeBox *RECT) HRESULT
	Show(fShow BOOL) HRESULT
	UIActivate(fUIActivate BOOL) HRESULT
	Open() HRESULT
	CloseView(dwReserved uint32) HRESULT
	SaveViewState(pstm *IStream) HRESULT
	ApplyViewState(pstm *IStream) HRESULT
	Clone(pIPSiteNew *IOleInPlaceSite, ppViewNew **IOleDocumentView) HRESULT
}

type IOleDocumentViewVtbl struct {
	IUnknownVtbl
	SetInPlaceSite uintptr
	GetInPlaceSite uintptr
	GetDocument    uintptr
	SetRect        uintptr
	GetRect        uintptr
	SetRectComplex uintptr
	Show           uintptr
	UIActivate     uintptr
	Open           uintptr
	CloseView      uintptr
	SaveViewState  uintptr
	ApplyViewState uintptr
	Clone          uintptr
}

type IOleDocumentView struct {
	IUnknown
}

func (this *IOleDocumentView) Vtbl() *IOleDocumentViewVtbl {
	return (*IOleDocumentViewVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IOleDocumentView) SetInPlaceSite(pIPSite *IOleInPlaceSite) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetInPlaceSite, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pIPSite)))
	return HRESULT(ret)
}

func (this *IOleDocumentView) GetInPlaceSite(ppIPSite **IOleInPlaceSite) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetInPlaceSite, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(ppIPSite)))
	return HRESULT(ret)
}

func (this *IOleDocumentView) GetDocument(ppunk **IUnknown) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetDocument, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(ppunk)))
	return HRESULT(ret)
}

func (this *IOleDocumentView) SetRect(prcView *RECT) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetRect, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(prcView)))
	return HRESULT(ret)
}

func (this *IOleDocumentView) GetRect(prcView *RECT) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetRect, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(prcView)))
	return HRESULT(ret)
}

func (this *IOleDocumentView) SetRectComplex(prcView *RECT, prcHScroll *RECT, prcVScroll *RECT, prcSizeBox *RECT) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetRectComplex, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(prcView)), uintptr(unsafe.Pointer(prcHScroll)), uintptr(unsafe.Pointer(prcVScroll)), uintptr(unsafe.Pointer(prcSizeBox)))
	return HRESULT(ret)
}

func (this *IOleDocumentView) Show(fShow BOOL) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Show, uintptr(unsafe.Pointer(this)), uintptr(fShow))
	return HRESULT(ret)
}

func (this *IOleDocumentView) UIActivate(fUIActivate BOOL) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().UIActivate, uintptr(unsafe.Pointer(this)), uintptr(fUIActivate))
	return HRESULT(ret)
}

func (this *IOleDocumentView) Open() HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Open, uintptr(unsafe.Pointer(this)))
	return HRESULT(ret)
}

func (this *IOleDocumentView) CloseView(dwReserved uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().CloseView, uintptr(unsafe.Pointer(this)), uintptr(dwReserved))
	return HRESULT(ret)
}

func (this *IOleDocumentView) SaveViewState(pstm *IStream) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SaveViewState, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pstm)))
	return HRESULT(ret)
}

func (this *IOleDocumentView) ApplyViewState(pstm *IStream) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().ApplyViewState, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pstm)))
	return HRESULT(ret)
}

func (this *IOleDocumentView) Clone(pIPSiteNew *IOleInPlaceSite, ppViewNew **IOleDocumentView) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Clone, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pIPSiteNew)), uintptr(unsafe.Pointer(ppViewNew)))
	return HRESULT(ret)
}

// B722BCC8-4E68-101B-A2BC-00AA00404770
var IID_IEnumOleDocumentViews = syscall.GUID{0xB722BCC8, 0x4E68, 0x101B,
	[8]byte{0xA2, 0xBC, 0x00, 0xAA, 0x00, 0x40, 0x47, 0x70}}

type IEnumOleDocumentViewsInterface interface {
	IUnknownInterface
	Next(cViews uint32, rgpView **IOleDocumentView, pcFetched *uint32) HRESULT
	Skip(cViews uint32) HRESULT
	Reset() HRESULT
	Clone(ppEnum **IEnumOleDocumentViews) HRESULT
}

type IEnumOleDocumentViewsVtbl struct {
	IUnknownVtbl
	Next  uintptr
	Skip  uintptr
	Reset uintptr
	Clone uintptr
}

type IEnumOleDocumentViews struct {
	IUnknown
}

func (this *IEnumOleDocumentViews) Vtbl() *IEnumOleDocumentViewsVtbl {
	return (*IEnumOleDocumentViewsVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IEnumOleDocumentViews) Next(cViews uint32, rgpView **IOleDocumentView, pcFetched *uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Next, uintptr(unsafe.Pointer(this)), uintptr(cViews), uintptr(unsafe.Pointer(rgpView)), uintptr(unsafe.Pointer(pcFetched)))
	return HRESULT(ret)
}

func (this *IEnumOleDocumentViews) Skip(cViews uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Skip, uintptr(unsafe.Pointer(this)), uintptr(cViews))
	return HRESULT(ret)
}

func (this *IEnumOleDocumentViews) Reset() HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Reset, uintptr(unsafe.Pointer(this)))
	return HRESULT(ret)
}

func (this *IEnumOleDocumentViews) Clone(ppEnum **IEnumOleDocumentViews) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Clone, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(ppEnum)))
	return HRESULT(ret)
}

// B722BCCA-4E68-101B-A2BC-00AA00404770
var IID_IContinueCallback = syscall.GUID{0xB722BCCA, 0x4E68, 0x101B,
	[8]byte{0xA2, 0xBC, 0x00, 0xAA, 0x00, 0x40, 0x47, 0x70}}

type IContinueCallbackInterface interface {
	IUnknownInterface
	FContinue() HRESULT
	FContinuePrinting(nCntPrinted int32, nCurPage int32, pwszPrintStatus PWSTR) HRESULT
}

type IContinueCallbackVtbl struct {
	IUnknownVtbl
	FContinue         uintptr
	FContinuePrinting uintptr
}

type IContinueCallback struct {
	IUnknown
}

func (this *IContinueCallback) Vtbl() *IContinueCallbackVtbl {
	return (*IContinueCallbackVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IContinueCallback) FContinue() HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().FContinue, uintptr(unsafe.Pointer(this)))
	return HRESULT(ret)
}

func (this *IContinueCallback) FContinuePrinting(nCntPrinted int32, nCurPage int32, pwszPrintStatus PWSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().FContinuePrinting, uintptr(unsafe.Pointer(this)), uintptr(nCntPrinted), uintptr(nCurPage), uintptr(unsafe.Pointer(pwszPrintStatus)))
	return HRESULT(ret)
}

// B722BCC9-4E68-101B-A2BC-00AA00404770
var IID_IPrint = syscall.GUID{0xB722BCC9, 0x4E68, 0x101B,
	[8]byte{0xA2, 0xBC, 0x00, 0xAA, 0x00, 0x40, 0x47, 0x70}}

type IPrintInterface interface {
	IUnknownInterface
	SetInitialPageNum(nFirstPage int32) HRESULT
	GetPageInfo(pnFirstPage *int32, pcPages *int32) HRESULT
	Print(grfFlags uint32, pptd **DVTARGETDEVICE, ppPageSet **PAGESET, pstgmOptions *STGMEDIUM, pcallback *IContinueCallback, nFirstPage int32, pcPagesPrinted *int32, pnLastPage *int32) HRESULT
}

type IPrintVtbl struct {
	IUnknownVtbl
	SetInitialPageNum uintptr
	GetPageInfo       uintptr
	Print             uintptr
}

type IPrint struct {
	IUnknown
}

func (this *IPrint) Vtbl() *IPrintVtbl {
	return (*IPrintVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IPrint) SetInitialPageNum(nFirstPage int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetInitialPageNum, uintptr(unsafe.Pointer(this)), uintptr(nFirstPage))
	return HRESULT(ret)
}

func (this *IPrint) GetPageInfo(pnFirstPage *int32, pcPages *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetPageInfo, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pnFirstPage)), uintptr(unsafe.Pointer(pcPages)))
	return HRESULT(ret)
}

func (this *IPrint) Print(grfFlags uint32, pptd **DVTARGETDEVICE, ppPageSet **PAGESET, pstgmOptions *STGMEDIUM, pcallback *IContinueCallback, nFirstPage int32, pcPagesPrinted *int32, pnLastPage *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Print, uintptr(unsafe.Pointer(this)), uintptr(grfFlags), uintptr(unsafe.Pointer(pptd)), uintptr(unsafe.Pointer(ppPageSet)), uintptr(unsafe.Pointer(pstgmOptions)), uintptr(unsafe.Pointer(pcallback)), uintptr(nFirstPage), uintptr(unsafe.Pointer(pcPagesPrinted)), uintptr(unsafe.Pointer(pnLastPage)))
	return HRESULT(ret)
}

// B722BCCB-4E68-101B-A2BC-00AA00404770
var IID_IOleCommandTarget = syscall.GUID{0xB722BCCB, 0x4E68, 0x101B,
	[8]byte{0xA2, 0xBC, 0x00, 0xAA, 0x00, 0x40, 0x47, 0x70}}

type IOleCommandTargetInterface interface {
	IUnknownInterface
	QueryStatus(pguidCmdGroup *syscall.GUID, cCmds uint32, prgCmds *OLECMD, pCmdText *OLECMDTEXT) HRESULT
	Exec(pguidCmdGroup *syscall.GUID, nCmdID uint32, nCmdexecopt uint32, pvaIn *VARIANT, pvaOut *VARIANT) HRESULT
}

type IOleCommandTargetVtbl struct {
	IUnknownVtbl
	QueryStatus uintptr
	Exec        uintptr
}

type IOleCommandTarget struct {
	IUnknown
}

func (this *IOleCommandTarget) Vtbl() *IOleCommandTargetVtbl {
	return (*IOleCommandTargetVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IOleCommandTarget) QueryStatus(pguidCmdGroup *syscall.GUID, cCmds uint32, prgCmds *OLECMD, pCmdText *OLECMDTEXT) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().QueryStatus, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pguidCmdGroup)), uintptr(cCmds), uintptr(unsafe.Pointer(prgCmds)), uintptr(unsafe.Pointer(pCmdText)))
	return HRESULT(ret)
}

func (this *IOleCommandTarget) Exec(pguidCmdGroup *syscall.GUID, nCmdID uint32, nCmdexecopt uint32, pvaIn *VARIANT, pvaOut *VARIANT) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().Exec, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pguidCmdGroup)), uintptr(nCmdID), uintptr(nCmdexecopt), uintptr(unsafe.Pointer(pvaIn)), uintptr(unsafe.Pointer(pvaOut)))
	return HRESULT(ret)
}

// 41B68150-904C-4E17-A0BA-A438182E359D
var IID_IZoomEvents = syscall.GUID{0x41B68150, 0x904C, 0x4E17,
	[8]byte{0xA0, 0xBA, 0xA4, 0x38, 0x18, 0x2E, 0x35, 0x9D}}

type IZoomEventsInterface interface {
	IUnknownInterface
	OnZoomPercentChanged(ulZoomPercent uint32) HRESULT
}

type IZoomEventsVtbl struct {
	IUnknownVtbl
	OnZoomPercentChanged uintptr
}

type IZoomEvents struct {
	IUnknown
}

func (this *IZoomEvents) Vtbl() *IZoomEventsVtbl {
	return (*IZoomEventsVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IZoomEvents) OnZoomPercentChanged(ulZoomPercent uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().OnZoomPercentChanged, uintptr(unsafe.Pointer(this)), uintptr(ulZoomPercent))
	return HRESULT(ret)
}

// D81F90A3-8156-44F7-AD28-5ABB87003274
var IID_IProtectFocus = syscall.GUID{0xD81F90A3, 0x8156, 0x44F7,
	[8]byte{0xAD, 0x28, 0x5A, 0xBB, 0x87, 0x00, 0x32, 0x74}}

type IProtectFocusInterface interface {
	IUnknownInterface
	AllowFocusChange(pfAllow *BOOL) HRESULT
}

type IProtectFocusVtbl struct {
	IUnknownVtbl
	AllowFocusChange uintptr
}

type IProtectFocus struct {
	IUnknown
}

func (this *IProtectFocus) Vtbl() *IProtectFocusVtbl {
	return (*IProtectFocusVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IProtectFocus) AllowFocusChange(pfAllow *BOOL) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().AllowFocusChange, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pfAllow)))
	return HRESULT(ret)
}

// 73C105EE-9DFF-4A07-B83C-7EFF290C266E
var IID_IProtectedModeMenuServices = syscall.GUID{0x73C105EE, 0x9DFF, 0x4A07,
	[8]byte{0xB8, 0x3C, 0x7E, 0xFF, 0x29, 0x0C, 0x26, 0x6E}}

type IProtectedModeMenuServicesInterface interface {
	IUnknownInterface
	CreateMenu(phMenu *HMENU) HRESULT
	LoadMenu(pszModuleName PWSTR, pszMenuName PWSTR, phMenu *HMENU) HRESULT
	LoadMenuID(pszModuleName PWSTR, wResourceID uint16, phMenu *HMENU) HRESULT
}

type IProtectedModeMenuServicesVtbl struct {
	IUnknownVtbl
	CreateMenu uintptr
	LoadMenu   uintptr
	LoadMenuID uintptr
}

type IProtectedModeMenuServices struct {
	IUnknown
}

func (this *IProtectedModeMenuServices) Vtbl() *IProtectedModeMenuServicesVtbl {
	return (*IProtectedModeMenuServicesVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IProtectedModeMenuServices) CreateMenu(phMenu *HMENU) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().CreateMenu, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(phMenu)))
	return HRESULT(ret)
}

func (this *IProtectedModeMenuServices) LoadMenu(pszModuleName PWSTR, pszMenuName PWSTR, phMenu *HMENU) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().LoadMenu, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pszModuleName)), uintptr(unsafe.Pointer(pszMenuName)), uintptr(unsafe.Pointer(phMenu)))
	return HRESULT(ret)
}

func (this *IProtectedModeMenuServices) LoadMenuID(pszModuleName PWSTR, wResourceID uint16, phMenu *HMENU) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().LoadMenuID, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pszModuleName)), uintptr(wResourceID), uintptr(unsafe.Pointer(phMenu)))
	return HRESULT(ret)
}

// 00000000-0000-0000-0000-000000000000
var IID_IOleUILinkContainerW = syscall.GUID{0x00000000, 0x0000, 0x0000,
	[8]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}}

type IOleUILinkContainerWInterface interface {
	IUnknownInterface
	GetNextLink(dwLink uint32) uint32
	SetLinkUpdateOptions(dwLink uint32, dwUpdateOpt uint32) HRESULT
	GetLinkUpdateOptions(dwLink uint32, lpdwUpdateOpt *uint32) HRESULT
	SetLinkSource(dwLink uint32, lpszDisplayName PWSTR, lenFileName uint32, pchEaten *uint32, fValidateSource BOOL) HRESULT
	GetLinkSource(dwLink uint32, lplpszDisplayName *PWSTR, lplenFileName *uint32, lplpszFullLinkType *PWSTR, lplpszShortLinkType *PWSTR, lpfSourceAvailable *BOOL, lpfIsSelected *BOOL) HRESULT
	OpenLinkSource(dwLink uint32) HRESULT
	UpdateLink(dwLink uint32, fErrorMessage BOOL, fReserved BOOL) HRESULT
	CancelLink(dwLink uint32) HRESULT
}

type IOleUILinkContainerWVtbl struct {
	IUnknownVtbl
	GetNextLink          uintptr
	SetLinkUpdateOptions uintptr
	GetLinkUpdateOptions uintptr
	SetLinkSource        uintptr
	GetLinkSource        uintptr
	OpenLinkSource       uintptr
	UpdateLink           uintptr
	CancelLink           uintptr
}

type IOleUILinkContainerW struct {
	IUnknown
}

func (this *IOleUILinkContainerW) Vtbl() *IOleUILinkContainerWVtbl {
	return (*IOleUILinkContainerWVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IOleUILinkContainerW) GetNextLink(dwLink uint32) uint32 {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetNextLink, uintptr(unsafe.Pointer(this)), uintptr(dwLink))
	return uint32(ret)
}

func (this *IOleUILinkContainerW) SetLinkUpdateOptions(dwLink uint32, dwUpdateOpt uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetLinkUpdateOptions, uintptr(unsafe.Pointer(this)), uintptr(dwLink), uintptr(dwUpdateOpt))
	return HRESULT(ret)
}

func (this *IOleUILinkContainerW) GetLinkUpdateOptions(dwLink uint32, lpdwUpdateOpt *uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetLinkUpdateOptions, uintptr(unsafe.Pointer(this)), uintptr(dwLink), uintptr(unsafe.Pointer(lpdwUpdateOpt)))
	return HRESULT(ret)
}

func (this *IOleUILinkContainerW) SetLinkSource(dwLink uint32, lpszDisplayName PWSTR, lenFileName uint32, pchEaten *uint32, fValidateSource BOOL) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetLinkSource, uintptr(unsafe.Pointer(this)), uintptr(dwLink), uintptr(unsafe.Pointer(lpszDisplayName)), uintptr(lenFileName), uintptr(unsafe.Pointer(pchEaten)), uintptr(fValidateSource))
	return HRESULT(ret)
}

func (this *IOleUILinkContainerW) GetLinkSource(dwLink uint32, lplpszDisplayName *PWSTR, lplenFileName *uint32, lplpszFullLinkType *PWSTR, lplpszShortLinkType *PWSTR, lpfSourceAvailable *BOOL, lpfIsSelected *BOOL) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetLinkSource, uintptr(unsafe.Pointer(this)), uintptr(dwLink), uintptr(unsafe.Pointer(lplpszDisplayName)), uintptr(unsafe.Pointer(lplenFileName)), uintptr(unsafe.Pointer(lplpszFullLinkType)), uintptr(unsafe.Pointer(lplpszShortLinkType)), uintptr(unsafe.Pointer(lpfSourceAvailable)), uintptr(unsafe.Pointer(lpfIsSelected)))
	return HRESULT(ret)
}

func (this *IOleUILinkContainerW) OpenLinkSource(dwLink uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().OpenLinkSource, uintptr(unsafe.Pointer(this)), uintptr(dwLink))
	return HRESULT(ret)
}

func (this *IOleUILinkContainerW) UpdateLink(dwLink uint32, fErrorMessage BOOL, fReserved BOOL) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().UpdateLink, uintptr(unsafe.Pointer(this)), uintptr(dwLink), uintptr(fErrorMessage), uintptr(fReserved))
	return HRESULT(ret)
}

func (this *IOleUILinkContainerW) CancelLink(dwLink uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().CancelLink, uintptr(unsafe.Pointer(this)), uintptr(dwLink))
	return HRESULT(ret)
}

// 00000000-0000-0000-0000-000000000000
var IID_IOleUILinkContainerA = syscall.GUID{0x00000000, 0x0000, 0x0000,
	[8]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}}

type IOleUILinkContainerAInterface interface {
	IUnknownInterface
	GetNextLink(dwLink uint32) uint32
	SetLinkUpdateOptions(dwLink uint32, dwUpdateOpt uint32) HRESULT
	GetLinkUpdateOptions(dwLink uint32, lpdwUpdateOpt *uint32) HRESULT
	SetLinkSource(dwLink uint32, lpszDisplayName PSTR, lenFileName uint32, pchEaten *uint32, fValidateSource BOOL) HRESULT
	GetLinkSource(dwLink uint32, lplpszDisplayName *PSTR, lplenFileName *uint32, lplpszFullLinkType *PSTR, lplpszShortLinkType *PSTR, lpfSourceAvailable *BOOL, lpfIsSelected *BOOL) HRESULT
	OpenLinkSource(dwLink uint32) HRESULT
	UpdateLink(dwLink uint32, fErrorMessage BOOL, fReserved BOOL) HRESULT
	CancelLink(dwLink uint32) HRESULT
}

type IOleUILinkContainerAVtbl struct {
	IUnknownVtbl
	GetNextLink          uintptr
	SetLinkUpdateOptions uintptr
	GetLinkUpdateOptions uintptr
	SetLinkSource        uintptr
	GetLinkSource        uintptr
	OpenLinkSource       uintptr
	UpdateLink           uintptr
	CancelLink           uintptr
}

type IOleUILinkContainerA struct {
	IUnknown
}

func (this *IOleUILinkContainerA) Vtbl() *IOleUILinkContainerAVtbl {
	return (*IOleUILinkContainerAVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IOleUILinkContainerA) GetNextLink(dwLink uint32) uint32 {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetNextLink, uintptr(unsafe.Pointer(this)), uintptr(dwLink))
	return uint32(ret)
}

func (this *IOleUILinkContainerA) SetLinkUpdateOptions(dwLink uint32, dwUpdateOpt uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetLinkUpdateOptions, uintptr(unsafe.Pointer(this)), uintptr(dwLink), uintptr(dwUpdateOpt))
	return HRESULT(ret)
}

func (this *IOleUILinkContainerA) GetLinkUpdateOptions(dwLink uint32, lpdwUpdateOpt *uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetLinkUpdateOptions, uintptr(unsafe.Pointer(this)), uintptr(dwLink), uintptr(unsafe.Pointer(lpdwUpdateOpt)))
	return HRESULT(ret)
}

func (this *IOleUILinkContainerA) SetLinkSource(dwLink uint32, lpszDisplayName PSTR, lenFileName uint32, pchEaten *uint32, fValidateSource BOOL) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetLinkSource, uintptr(unsafe.Pointer(this)), uintptr(dwLink), uintptr(unsafe.Pointer(lpszDisplayName)), uintptr(lenFileName), uintptr(unsafe.Pointer(pchEaten)), uintptr(fValidateSource))
	return HRESULT(ret)
}

func (this *IOleUILinkContainerA) GetLinkSource(dwLink uint32, lplpszDisplayName *PSTR, lplenFileName *uint32, lplpszFullLinkType *PSTR, lplpszShortLinkType *PSTR, lpfSourceAvailable *BOOL, lpfIsSelected *BOOL) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetLinkSource, uintptr(unsafe.Pointer(this)), uintptr(dwLink), uintptr(unsafe.Pointer(lplpszDisplayName)), uintptr(unsafe.Pointer(lplenFileName)), uintptr(unsafe.Pointer(lplpszFullLinkType)), uintptr(unsafe.Pointer(lplpszShortLinkType)), uintptr(unsafe.Pointer(lpfSourceAvailable)), uintptr(unsafe.Pointer(lpfIsSelected)))
	return HRESULT(ret)
}

func (this *IOleUILinkContainerA) OpenLinkSource(dwLink uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().OpenLinkSource, uintptr(unsafe.Pointer(this)), uintptr(dwLink))
	return HRESULT(ret)
}

func (this *IOleUILinkContainerA) UpdateLink(dwLink uint32, fErrorMessage BOOL, fReserved BOOL) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().UpdateLink, uintptr(unsafe.Pointer(this)), uintptr(dwLink), uintptr(fErrorMessage), uintptr(fReserved))
	return HRESULT(ret)
}

func (this *IOleUILinkContainerA) CancelLink(dwLink uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().CancelLink, uintptr(unsafe.Pointer(this)), uintptr(dwLink))
	return HRESULT(ret)
}

// 00000000-0000-0000-0000-000000000000
var IID_IOleUIObjInfoW = syscall.GUID{0x00000000, 0x0000, 0x0000,
	[8]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}}

type IOleUIObjInfoWInterface interface {
	IUnknownInterface
	GetObjectInfo(dwObject uint32, lpdwObjSize *uint32, lplpszLabel *PWSTR, lplpszType *PWSTR, lplpszShortType *PWSTR, lplpszLocation *PWSTR) HRESULT
	GetConvertInfo(dwObject uint32, lpClassID *syscall.GUID, lpwFormat *uint16, lpConvertDefaultClassID *syscall.GUID, lplpClsidExclude **syscall.GUID, lpcClsidExclude *uint32) HRESULT
	ConvertObject(dwObject uint32, clsidNew *syscall.GUID) HRESULT
	GetViewInfo(dwObject uint32, phMetaPict *uintptr, pdvAspect *uint32, pnCurrentScale *int32) HRESULT
	SetViewInfo(dwObject uint32, hMetaPict uintptr, dvAspect uint32, nCurrentScale int32, bRelativeToOrig BOOL) HRESULT
}

type IOleUIObjInfoWVtbl struct {
	IUnknownVtbl
	GetObjectInfo  uintptr
	GetConvertInfo uintptr
	ConvertObject  uintptr
	GetViewInfo    uintptr
	SetViewInfo    uintptr
}

type IOleUIObjInfoW struct {
	IUnknown
}

func (this *IOleUIObjInfoW) Vtbl() *IOleUIObjInfoWVtbl {
	return (*IOleUIObjInfoWVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IOleUIObjInfoW) GetObjectInfo(dwObject uint32, lpdwObjSize *uint32, lplpszLabel *PWSTR, lplpszType *PWSTR, lplpszShortType *PWSTR, lplpszLocation *PWSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetObjectInfo, uintptr(unsafe.Pointer(this)), uintptr(dwObject), uintptr(unsafe.Pointer(lpdwObjSize)), uintptr(unsafe.Pointer(lplpszLabel)), uintptr(unsafe.Pointer(lplpszType)), uintptr(unsafe.Pointer(lplpszShortType)), uintptr(unsafe.Pointer(lplpszLocation)))
	return HRESULT(ret)
}

func (this *IOleUIObjInfoW) GetConvertInfo(dwObject uint32, lpClassID *syscall.GUID, lpwFormat *uint16, lpConvertDefaultClassID *syscall.GUID, lplpClsidExclude **syscall.GUID, lpcClsidExclude *uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetConvertInfo, uintptr(unsafe.Pointer(this)), uintptr(dwObject), uintptr(unsafe.Pointer(lpClassID)), uintptr(unsafe.Pointer(lpwFormat)), uintptr(unsafe.Pointer(lpConvertDefaultClassID)), uintptr(unsafe.Pointer(lplpClsidExclude)), uintptr(unsafe.Pointer(lpcClsidExclude)))
	return HRESULT(ret)
}

func (this *IOleUIObjInfoW) ConvertObject(dwObject uint32, clsidNew *syscall.GUID) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().ConvertObject, uintptr(unsafe.Pointer(this)), uintptr(dwObject), uintptr(unsafe.Pointer(clsidNew)))
	return HRESULT(ret)
}

func (this *IOleUIObjInfoW) GetViewInfo(dwObject uint32, phMetaPict *uintptr, pdvAspect *uint32, pnCurrentScale *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetViewInfo, uintptr(unsafe.Pointer(this)), uintptr(dwObject), uintptr(unsafe.Pointer(phMetaPict)), uintptr(unsafe.Pointer(pdvAspect)), uintptr(unsafe.Pointer(pnCurrentScale)))
	return HRESULT(ret)
}

func (this *IOleUIObjInfoW) SetViewInfo(dwObject uint32, hMetaPict uintptr, dvAspect uint32, nCurrentScale int32, bRelativeToOrig BOOL) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetViewInfo, uintptr(unsafe.Pointer(this)), uintptr(dwObject), hMetaPict, uintptr(dvAspect), uintptr(nCurrentScale), uintptr(bRelativeToOrig))
	return HRESULT(ret)
}

// 00000000-0000-0000-0000-000000000000
var IID_IOleUIObjInfoA = syscall.GUID{0x00000000, 0x0000, 0x0000,
	[8]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}}

type IOleUIObjInfoAInterface interface {
	IUnknownInterface
	GetObjectInfo(dwObject uint32, lpdwObjSize *uint32, lplpszLabel *PSTR, lplpszType *PSTR, lplpszShortType *PSTR, lplpszLocation *PSTR) HRESULT
	GetConvertInfo(dwObject uint32, lpClassID *syscall.GUID, lpwFormat *uint16, lpConvertDefaultClassID *syscall.GUID, lplpClsidExclude **syscall.GUID, lpcClsidExclude *uint32) HRESULT
	ConvertObject(dwObject uint32, clsidNew *syscall.GUID) HRESULT
	GetViewInfo(dwObject uint32, phMetaPict *uintptr, pdvAspect *uint32, pnCurrentScale *int32) HRESULT
	SetViewInfo(dwObject uint32, hMetaPict uintptr, dvAspect uint32, nCurrentScale int32, bRelativeToOrig BOOL) HRESULT
}

type IOleUIObjInfoAVtbl struct {
	IUnknownVtbl
	GetObjectInfo  uintptr
	GetConvertInfo uintptr
	ConvertObject  uintptr
	GetViewInfo    uintptr
	SetViewInfo    uintptr
}

type IOleUIObjInfoA struct {
	IUnknown
}

func (this *IOleUIObjInfoA) Vtbl() *IOleUIObjInfoAVtbl {
	return (*IOleUIObjInfoAVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IOleUIObjInfoA) GetObjectInfo(dwObject uint32, lpdwObjSize *uint32, lplpszLabel *PSTR, lplpszType *PSTR, lplpszShortType *PSTR, lplpszLocation *PSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetObjectInfo, uintptr(unsafe.Pointer(this)), uintptr(dwObject), uintptr(unsafe.Pointer(lpdwObjSize)), uintptr(unsafe.Pointer(lplpszLabel)), uintptr(unsafe.Pointer(lplpszType)), uintptr(unsafe.Pointer(lplpszShortType)), uintptr(unsafe.Pointer(lplpszLocation)))
	return HRESULT(ret)
}

func (this *IOleUIObjInfoA) GetConvertInfo(dwObject uint32, lpClassID *syscall.GUID, lpwFormat *uint16, lpConvertDefaultClassID *syscall.GUID, lplpClsidExclude **syscall.GUID, lpcClsidExclude *uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetConvertInfo, uintptr(unsafe.Pointer(this)), uintptr(dwObject), uintptr(unsafe.Pointer(lpClassID)), uintptr(unsafe.Pointer(lpwFormat)), uintptr(unsafe.Pointer(lpConvertDefaultClassID)), uintptr(unsafe.Pointer(lplpClsidExclude)), uintptr(unsafe.Pointer(lpcClsidExclude)))
	return HRESULT(ret)
}

func (this *IOleUIObjInfoA) ConvertObject(dwObject uint32, clsidNew *syscall.GUID) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().ConvertObject, uintptr(unsafe.Pointer(this)), uintptr(dwObject), uintptr(unsafe.Pointer(clsidNew)))
	return HRESULT(ret)
}

func (this *IOleUIObjInfoA) GetViewInfo(dwObject uint32, phMetaPict *uintptr, pdvAspect *uint32, pnCurrentScale *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetViewInfo, uintptr(unsafe.Pointer(this)), uintptr(dwObject), uintptr(unsafe.Pointer(phMetaPict)), uintptr(unsafe.Pointer(pdvAspect)), uintptr(unsafe.Pointer(pnCurrentScale)))
	return HRESULT(ret)
}

func (this *IOleUIObjInfoA) SetViewInfo(dwObject uint32, hMetaPict uintptr, dvAspect uint32, nCurrentScale int32, bRelativeToOrig BOOL) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().SetViewInfo, uintptr(unsafe.Pointer(this)), uintptr(dwObject), hMetaPict, uintptr(dvAspect), uintptr(nCurrentScale), uintptr(bRelativeToOrig))
	return HRESULT(ret)
}

// 00000000-0000-0000-0000-000000000000
var IID_IOleUILinkInfoW = syscall.GUID{0x00000000, 0x0000, 0x0000,
	[8]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}}

type IOleUILinkInfoWInterface interface {
	IOleUILinkContainerWInterface
	GetLastUpdate(dwLink uint32, lpLastUpdate *FILETIME) HRESULT
}

type IOleUILinkInfoWVtbl struct {
	IOleUILinkContainerWVtbl
	GetLastUpdate uintptr
}

type IOleUILinkInfoW struct {
	IOleUILinkContainerW
}

func (this *IOleUILinkInfoW) Vtbl() *IOleUILinkInfoWVtbl {
	return (*IOleUILinkInfoWVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IOleUILinkInfoW) GetLastUpdate(dwLink uint32, lpLastUpdate *FILETIME) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetLastUpdate, uintptr(unsafe.Pointer(this)), uintptr(dwLink), uintptr(unsafe.Pointer(lpLastUpdate)))
	return HRESULT(ret)
}

// 00000000-0000-0000-0000-000000000000
var IID_IOleUILinkInfoA = syscall.GUID{0x00000000, 0x0000, 0x0000,
	[8]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}}

type IOleUILinkInfoAInterface interface {
	IOleUILinkContainerAInterface
	GetLastUpdate(dwLink uint32, lpLastUpdate *FILETIME) HRESULT
}

type IOleUILinkInfoAVtbl struct {
	IOleUILinkContainerAVtbl
	GetLastUpdate uintptr
}

type IOleUILinkInfoA struct {
	IOleUILinkContainerA
}

func (this *IOleUILinkInfoA) Vtbl() *IOleUILinkInfoAVtbl {
	return (*IOleUILinkInfoAVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IOleUILinkInfoA) GetLastUpdate(dwLink uint32, lpLastUpdate *FILETIME) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetLastUpdate, uintptr(unsafe.Pointer(this)), uintptr(dwLink), uintptr(unsafe.Pointer(lpLastUpdate)))
	return HRESULT(ret)
}

// A6EF9860-C720-11D0-9337-00A0C90DCAA9
var IID_IDispatchEx = syscall.GUID{0xA6EF9860, 0xC720, 0x11D0,
	[8]byte{0x93, 0x37, 0x00, 0xA0, 0xC9, 0x0D, 0xCA, 0xA9}}

type IDispatchExInterface interface {
	IDispatchInterface
	GetDispID(bstrName BSTR, grfdex uint32, pid *int32) HRESULT
	InvokeEx(id int32, lcid uint32, wFlags uint16, pdp *DISPPARAMS, pvarRes *VARIANT, pei *EXCEPINFO, pspCaller *IServiceProvider) HRESULT
	DeleteMemberByName(bstrName BSTR, grfdex uint32) HRESULT
	DeleteMemberByDispID(id int32) HRESULT
	GetMemberProperties(id int32, grfdexFetch uint32, pgrfdex *uint32) HRESULT
	GetMemberName(id int32, pbstrName *BSTR) HRESULT
	GetNextDispID(grfdex uint32, id int32, pid *int32) HRESULT
	GetNameSpaceParent(ppunk **IUnknown) HRESULT
}

type IDispatchExVtbl struct {
	IDispatchVtbl
	GetDispID            uintptr
	InvokeEx             uintptr
	DeleteMemberByName   uintptr
	DeleteMemberByDispID uintptr
	GetMemberProperties  uintptr
	GetMemberName        uintptr
	GetNextDispID        uintptr
	GetNameSpaceParent   uintptr
}

type IDispatchEx struct {
	IDispatch
}

func (this *IDispatchEx) Vtbl() *IDispatchExVtbl {
	return (*IDispatchExVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IDispatchEx) GetDispID(bstrName BSTR, grfdex uint32, pid *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetDispID, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(bstrName)), uintptr(grfdex), uintptr(unsafe.Pointer(pid)))
	return HRESULT(ret)
}

func (this *IDispatchEx) InvokeEx(id int32, lcid uint32, wFlags uint16, pdp *DISPPARAMS, pvarRes *VARIANT, pei *EXCEPINFO, pspCaller *IServiceProvider) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().InvokeEx, uintptr(unsafe.Pointer(this)), uintptr(id), uintptr(lcid), uintptr(wFlags), uintptr(unsafe.Pointer(pdp)), uintptr(unsafe.Pointer(pvarRes)), uintptr(unsafe.Pointer(pei)), uintptr(unsafe.Pointer(pspCaller)))
	return HRESULT(ret)
}

func (this *IDispatchEx) DeleteMemberByName(bstrName BSTR, grfdex uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().DeleteMemberByName, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(bstrName)), uintptr(grfdex))
	return HRESULT(ret)
}

func (this *IDispatchEx) DeleteMemberByDispID(id int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().DeleteMemberByDispID, uintptr(unsafe.Pointer(this)), uintptr(id))
	return HRESULT(ret)
}

func (this *IDispatchEx) GetMemberProperties(id int32, grfdexFetch uint32, pgrfdex *uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetMemberProperties, uintptr(unsafe.Pointer(this)), uintptr(id), uintptr(grfdexFetch), uintptr(unsafe.Pointer(pgrfdex)))
	return HRESULT(ret)
}

func (this *IDispatchEx) GetMemberName(id int32, pbstrName *BSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetMemberName, uintptr(unsafe.Pointer(this)), uintptr(id), uintptr(unsafe.Pointer(pbstrName)))
	return HRESULT(ret)
}

func (this *IDispatchEx) GetNextDispID(grfdex uint32, id int32, pid *int32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetNextDispID, uintptr(unsafe.Pointer(this)), uintptr(grfdex), uintptr(id), uintptr(unsafe.Pointer(pid)))
	return HRESULT(ret)
}

func (this *IDispatchEx) GetNameSpaceParent(ppunk **IUnknown) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetNameSpaceParent, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(ppunk)))
	return HRESULT(ret)
}

// A6EF9861-C720-11D0-9337-00A0C90DCAA9
var IID_IDispError = syscall.GUID{0xA6EF9861, 0xC720, 0x11D0,
	[8]byte{0x93, 0x37, 0x00, 0xA0, 0xC9, 0x0D, 0xCA, 0xA9}}

type IDispErrorInterface interface {
	IUnknownInterface
	QueryErrorInfo(guidErrorType syscall.GUID, ppde **IDispError) HRESULT
	GetNext(ppde **IDispError) HRESULT
	GetHresult(phr *HRESULT) HRESULT
	GetSource(pbstrSource *BSTR) HRESULT
	GetHelpInfo(pbstrFileName *BSTR, pdwContext *uint32) HRESULT
	GetDescription(pbstrDescription *BSTR) HRESULT
}

type IDispErrorVtbl struct {
	IUnknownVtbl
	QueryErrorInfo uintptr
	GetNext        uintptr
	GetHresult     uintptr
	GetSource      uintptr
	GetHelpInfo    uintptr
	GetDescription uintptr
}

type IDispError struct {
	IUnknown
}

func (this *IDispError) Vtbl() *IDispErrorVtbl {
	return (*IDispErrorVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IDispError) QueryErrorInfo(guidErrorType syscall.GUID, ppde **IDispError) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().QueryErrorInfo, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(&guidErrorType)), uintptr(unsafe.Pointer(ppde)))
	return HRESULT(ret)
}

func (this *IDispError) GetNext(ppde **IDispError) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetNext, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(ppde)))
	return HRESULT(ret)
}

func (this *IDispError) GetHresult(phr *HRESULT) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetHresult, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(phr)))
	return HRESULT(ret)
}

func (this *IDispError) GetSource(pbstrSource *BSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetSource, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pbstrSource)))
	return HRESULT(ret)
}

func (this *IDispError) GetHelpInfo(pbstrFileName *BSTR, pdwContext *uint32) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetHelpInfo, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pbstrFileName)), uintptr(unsafe.Pointer(pdwContext)))
	return HRESULT(ret)
}

func (this *IDispError) GetDescription(pbstrDescription *BSTR) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetDescription, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pbstrDescription)))
	return HRESULT(ret)
}

// A6EF9862-C720-11D0-9337-00A0C90DCAA9
var IID_IVariantChangeType = syscall.GUID{0xA6EF9862, 0xC720, 0x11D0,
	[8]byte{0x93, 0x37, 0x00, 0xA0, 0xC9, 0x0D, 0xCA, 0xA9}}

type IVariantChangeTypeInterface interface {
	IUnknownInterface
	ChangeType(pvarDst *VARIANT, pvarSrc *VARIANT, lcid uint32, vtNew uint16) HRESULT
}

type IVariantChangeTypeVtbl struct {
	IUnknownVtbl
	ChangeType uintptr
}

type IVariantChangeType struct {
	IUnknown
}

func (this *IVariantChangeType) Vtbl() *IVariantChangeTypeVtbl {
	return (*IVariantChangeTypeVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IVariantChangeType) ChangeType(pvarDst *VARIANT, pvarSrc *VARIANT, lcid uint32, vtNew uint16) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().ChangeType, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pvarDst)), uintptr(unsafe.Pointer(pvarSrc)), uintptr(lcid), uintptr(vtNew))
	return HRESULT(ret)
}

// CA04B7E6-0D21-11D1-8CC5-00C04FC2B085
var IID_IObjectIdentity = syscall.GUID{0xCA04B7E6, 0x0D21, 0x11D1,
	[8]byte{0x8C, 0xC5, 0x00, 0xC0, 0x4F, 0xC2, 0xB0, 0x85}}

type IObjectIdentityInterface interface {
	IUnknownInterface
	IsEqualObject(punk *IUnknown) HRESULT
}

type IObjectIdentityVtbl struct {
	IUnknownVtbl
	IsEqualObject uintptr
}

type IObjectIdentity struct {
	IUnknown
}

func (this *IObjectIdentity) Vtbl() *IObjectIdentityVtbl {
	return (*IObjectIdentityVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IObjectIdentity) IsEqualObject(punk *IUnknown) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().IsEqualObject, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(punk)))
	return HRESULT(ret)
}

// C5598E60-B307-11D1-B27D-006008C3FBFB
var IID_ICanHandleException = syscall.GUID{0xC5598E60, 0xB307, 0x11D1,
	[8]byte{0xB2, 0x7D, 0x00, 0x60, 0x08, 0xC3, 0xFB, 0xFB}}

type ICanHandleExceptionInterface interface {
	IUnknownInterface
	CanHandleException(pExcepInfo *EXCEPINFO, pvar *VARIANT) HRESULT
}

type ICanHandleExceptionVtbl struct {
	IUnknownVtbl
	CanHandleException uintptr
}

type ICanHandleException struct {
	IUnknown
}

func (this *ICanHandleException) Vtbl() *ICanHandleExceptionVtbl {
	return (*ICanHandleExceptionVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *ICanHandleException) CanHandleException(pExcepInfo *EXCEPINFO, pvar *VARIANT) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().CanHandleException, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pExcepInfo)), uintptr(unsafe.Pointer(pvar)))
	return HRESULT(ret)
}

// 10E2414A-EC59-49D2-BC51-5ADD2C36FEBC
var IID_IProvideRuntimeContext = syscall.GUID{0x10E2414A, 0xEC59, 0x49D2,
	[8]byte{0xBC, 0x51, 0x5A, 0xDD, 0x2C, 0x36, 0xFE, 0xBC}}

type IProvideRuntimeContextInterface interface {
	IUnknownInterface
	GetCurrentSourceContext(pdwContext *uintptr, pfExecutingGlobalCode *int16) HRESULT
}

type IProvideRuntimeContextVtbl struct {
	IUnknownVtbl
	GetCurrentSourceContext uintptr
}

type IProvideRuntimeContext struct {
	IUnknown
}

func (this *IProvideRuntimeContext) Vtbl() *IProvideRuntimeContextVtbl {
	return (*IProvideRuntimeContextVtbl)(unsafe.Pointer(this.IUnknown.LpVtbl))
}

func (this *IProvideRuntimeContext) GetCurrentSourceContext(pdwContext *uintptr, pfExecutingGlobalCode *int16) HRESULT {
	ret, _, _ := syscall.SyscallN(this.Vtbl().GetCurrentSourceContext, uintptr(unsafe.Pointer(this)), uintptr(unsafe.Pointer(pdwContext)), uintptr(unsafe.Pointer(pfExecutingGlobalCode)))
	return HRESULT(ret)
}

var (
	pDosDateTimeToVariantTime          uintptr
	pVariantTimeToDosDateTime          uintptr
	pSystemTimeToVariantTime           uintptr
	pVariantTimeToSystemTime           uintptr
	pSafeArrayAllocDescriptor          uintptr
	pSafeArrayAllocDescriptorEx        uintptr
	pSafeArrayAllocData                uintptr
	pSafeArrayCreate                   uintptr
	pSafeArrayCreateEx                 uintptr
	pSafeArrayCopyData                 uintptr
	pSafeArrayReleaseDescriptor        uintptr
	pSafeArrayDestroyDescriptor        uintptr
	pSafeArrayReleaseData              uintptr
	pSafeArrayDestroyData              uintptr
	pSafeArrayAddRef                   uintptr
	pSafeArrayDestroy                  uintptr
	pSafeArrayRedim                    uintptr
	pSafeArrayGetDim                   uintptr
	pSafeArrayGetElemsize              uintptr
	pSafeArrayGetUBound                uintptr
	pSafeArrayGetLBound                uintptr
	pSafeArrayLock                     uintptr
	pSafeArrayUnlock                   uintptr
	pSafeArrayAccessData               uintptr
	pSafeArrayUnaccessData             uintptr
	pSafeArrayGetElement               uintptr
	pSafeArrayPutElement               uintptr
	pSafeArrayCopy                     uintptr
	pSafeArrayPtrOfIndex               uintptr
	pSafeArraySetRecordInfo            uintptr
	pSafeArrayGetRecordInfo            uintptr
	pSafeArraySetIID                   uintptr
	pSafeArrayGetIID                   uintptr
	pSafeArrayGetVartype               uintptr
	pSafeArrayCreateVector             uintptr
	pSafeArrayCreateVectorEx           uintptr
	pVariantInit                       uintptr
	pVariantClear                      uintptr
	pVariantCopy                       uintptr
	pVariantCopyInd                    uintptr
	pVariantChangeType                 uintptr
	pVariantChangeTypeEx               uintptr
	pVectorFromBstr                    uintptr
	pBstrFromVector                    uintptr
	pVarUI1FromI2                      uintptr
	pVarUI1FromI4                      uintptr
	pVarUI1FromI8                      uintptr
	pVarUI1FromR4                      uintptr
	pVarUI1FromR8                      uintptr
	pVarUI1FromCy                      uintptr
	pVarUI1FromDate                    uintptr
	pVarUI1FromStr                     uintptr
	pVarUI1FromDisp                    uintptr
	pVarUI1FromBool                    uintptr
	pVarUI1FromI1                      uintptr
	pVarUI1FromUI2                     uintptr
	pVarUI1FromUI4                     uintptr
	pVarUI1FromUI8                     uintptr
	pVarUI1FromDec                     uintptr
	pVarI2FromUI1                      uintptr
	pVarI2FromI4                       uintptr
	pVarI2FromI8                       uintptr
	pVarI2FromR4                       uintptr
	pVarI2FromR8                       uintptr
	pVarI2FromCy                       uintptr
	pVarI2FromDate                     uintptr
	pVarI2FromStr                      uintptr
	pVarI2FromDisp                     uintptr
	pVarI2FromBool                     uintptr
	pVarI2FromI1                       uintptr
	pVarI2FromUI2                      uintptr
	pVarI2FromUI4                      uintptr
	pVarI2FromUI8                      uintptr
	pVarI2FromDec                      uintptr
	pVarI4FromUI1                      uintptr
	pVarI4FromI2                       uintptr
	pVarI4FromI8                       uintptr
	pVarI4FromR4                       uintptr
	pVarI4FromR8                       uintptr
	pVarI4FromCy                       uintptr
	pVarI4FromDate                     uintptr
	pVarI4FromStr                      uintptr
	pVarI4FromDisp                     uintptr
	pVarI4FromBool                     uintptr
	pVarI4FromI1                       uintptr
	pVarI4FromUI2                      uintptr
	pVarI4FromUI4                      uintptr
	pVarI4FromUI8                      uintptr
	pVarI4FromDec                      uintptr
	pVarI8FromUI1                      uintptr
	pVarI8FromI2                       uintptr
	pVarI8FromR4                       uintptr
	pVarI8FromR8                       uintptr
	pVarI8FromCy                       uintptr
	pVarI8FromDate                     uintptr
	pVarI8FromStr                      uintptr
	pVarI8FromDisp                     uintptr
	pVarI8FromBool                     uintptr
	pVarI8FromI1                       uintptr
	pVarI8FromUI2                      uintptr
	pVarI8FromUI4                      uintptr
	pVarI8FromUI8                      uintptr
	pVarI8FromDec                      uintptr
	pVarR4FromUI1                      uintptr
	pVarR4FromI2                       uintptr
	pVarR4FromI4                       uintptr
	pVarR4FromI8                       uintptr
	pVarR4FromR8                       uintptr
	pVarR4FromCy                       uintptr
	pVarR4FromDate                     uintptr
	pVarR4FromStr                      uintptr
	pVarR4FromDisp                     uintptr
	pVarR4FromBool                     uintptr
	pVarR4FromI1                       uintptr
	pVarR4FromUI2                      uintptr
	pVarR4FromUI4                      uintptr
	pVarR4FromUI8                      uintptr
	pVarR4FromDec                      uintptr
	pVarR8FromUI1                      uintptr
	pVarR8FromI2                       uintptr
	pVarR8FromI4                       uintptr
	pVarR8FromI8                       uintptr
	pVarR8FromR4                       uintptr
	pVarR8FromCy                       uintptr
	pVarR8FromDate                     uintptr
	pVarR8FromStr                      uintptr
	pVarR8FromDisp                     uintptr
	pVarR8FromBool                     uintptr
	pVarR8FromI1                       uintptr
	pVarR8FromUI2                      uintptr
	pVarR8FromUI4                      uintptr
	pVarR8FromUI8                      uintptr
	pVarR8FromDec                      uintptr
	pVarDateFromUI1                    uintptr
	pVarDateFromI2                     uintptr
	pVarDateFromI4                     uintptr
	pVarDateFromI8                     uintptr
	pVarDateFromR4                     uintptr
	pVarDateFromR8                     uintptr
	pVarDateFromCy                     uintptr
	pVarDateFromStr                    uintptr
	pVarDateFromDisp                   uintptr
	pVarDateFromBool                   uintptr
	pVarDateFromI1                     uintptr
	pVarDateFromUI2                    uintptr
	pVarDateFromUI4                    uintptr
	pVarDateFromUI8                    uintptr
	pVarDateFromDec                    uintptr
	pVarCyFromUI1                      uintptr
	pVarCyFromI2                       uintptr
	pVarCyFromI4                       uintptr
	pVarCyFromI8                       uintptr
	pVarCyFromR4                       uintptr
	pVarCyFromR8                       uintptr
	pVarCyFromDate                     uintptr
	pVarCyFromStr                      uintptr
	pVarCyFromDisp                     uintptr
	pVarCyFromBool                     uintptr
	pVarCyFromI1                       uintptr
	pVarCyFromUI2                      uintptr
	pVarCyFromUI4                      uintptr
	pVarCyFromUI8                      uintptr
	pVarCyFromDec                      uintptr
	pVarBstrFromUI1                    uintptr
	pVarBstrFromI2                     uintptr
	pVarBstrFromI4                     uintptr
	pVarBstrFromI8                     uintptr
	pVarBstrFromR4                     uintptr
	pVarBstrFromR8                     uintptr
	pVarBstrFromCy                     uintptr
	pVarBstrFromDate                   uintptr
	pVarBstrFromDisp                   uintptr
	pVarBstrFromBool                   uintptr
	pVarBstrFromI1                     uintptr
	pVarBstrFromUI2                    uintptr
	pVarBstrFromUI4                    uintptr
	pVarBstrFromUI8                    uintptr
	pVarBstrFromDec                    uintptr
	pVarBoolFromUI1                    uintptr
	pVarBoolFromI2                     uintptr
	pVarBoolFromI4                     uintptr
	pVarBoolFromI8                     uintptr
	pVarBoolFromR4                     uintptr
	pVarBoolFromR8                     uintptr
	pVarBoolFromDate                   uintptr
	pVarBoolFromCy                     uintptr
	pVarBoolFromStr                    uintptr
	pVarBoolFromDisp                   uintptr
	pVarBoolFromI1                     uintptr
	pVarBoolFromUI2                    uintptr
	pVarBoolFromUI4                    uintptr
	pVarBoolFromUI8                    uintptr
	pVarBoolFromDec                    uintptr
	pVarI1FromUI1                      uintptr
	pVarI1FromI2                       uintptr
	pVarI1FromI4                       uintptr
	pVarI1FromI8                       uintptr
	pVarI1FromR4                       uintptr
	pVarI1FromR8                       uintptr
	pVarI1FromDate                     uintptr
	pVarI1FromCy                       uintptr
	pVarI1FromStr                      uintptr
	pVarI1FromDisp                     uintptr
	pVarI1FromBool                     uintptr
	pVarI1FromUI2                      uintptr
	pVarI1FromUI4                      uintptr
	pVarI1FromUI8                      uintptr
	pVarI1FromDec                      uintptr
	pVarUI2FromUI1                     uintptr
	pVarUI2FromI2                      uintptr
	pVarUI2FromI4                      uintptr
	pVarUI2FromI8                      uintptr
	pVarUI2FromR4                      uintptr
	pVarUI2FromR8                      uintptr
	pVarUI2FromDate                    uintptr
	pVarUI2FromCy                      uintptr
	pVarUI2FromStr                     uintptr
	pVarUI2FromDisp                    uintptr
	pVarUI2FromBool                    uintptr
	pVarUI2FromI1                      uintptr
	pVarUI2FromUI4                     uintptr
	pVarUI2FromUI8                     uintptr
	pVarUI2FromDec                     uintptr
	pVarUI4FromUI1                     uintptr
	pVarUI4FromI2                      uintptr
	pVarUI4FromI4                      uintptr
	pVarUI4FromI8                      uintptr
	pVarUI4FromR4                      uintptr
	pVarUI4FromR8                      uintptr
	pVarUI4FromDate                    uintptr
	pVarUI4FromCy                      uintptr
	pVarUI4FromStr                     uintptr
	pVarUI4FromDisp                    uintptr
	pVarUI4FromBool                    uintptr
	pVarUI4FromI1                      uintptr
	pVarUI4FromUI2                     uintptr
	pVarUI4FromUI8                     uintptr
	pVarUI4FromDec                     uintptr
	pVarUI8FromUI1                     uintptr
	pVarUI8FromI2                      uintptr
	pVarUI8FromI8                      uintptr
	pVarUI8FromR4                      uintptr
	pVarUI8FromR8                      uintptr
	pVarUI8FromCy                      uintptr
	pVarUI8FromDate                    uintptr
	pVarUI8FromStr                     uintptr
	pVarUI8FromDisp                    uintptr
	pVarUI8FromBool                    uintptr
	pVarUI8FromI1                      uintptr
	pVarUI8FromUI2                     uintptr
	pVarUI8FromUI4                     uintptr
	pVarUI8FromDec                     uintptr
	pVarDecFromUI1                     uintptr
	pVarDecFromI2                      uintptr
	pVarDecFromI4                      uintptr
	pVarDecFromI8                      uintptr
	pVarDecFromR4                      uintptr
	pVarDecFromR8                      uintptr
	pVarDecFromDate                    uintptr
	pVarDecFromCy                      uintptr
	pVarDecFromStr                     uintptr
	pVarDecFromDisp                    uintptr
	pVarDecFromBool                    uintptr
	pVarDecFromI1                      uintptr
	pVarDecFromUI2                     uintptr
	pVarDecFromUI4                     uintptr
	pVarDecFromUI8                     uintptr
	pVarParseNumFromStr                uintptr
	pVarNumFromParseNum                uintptr
	pVarAdd                            uintptr
	pVarAnd                            uintptr
	pVarCat                            uintptr
	pVarDiv                            uintptr
	pVarEqv                            uintptr
	pVarIdiv                           uintptr
	pVarImp                            uintptr
	pVarMod                            uintptr
	pVarMul                            uintptr
	pVarOr                             uintptr
	pVarPow                            uintptr
	pVarSub                            uintptr
	pVarXor                            uintptr
	pVarAbs                            uintptr
	pVarFix                            uintptr
	pVarInt                            uintptr
	pVarNeg                            uintptr
	pVarNot                            uintptr
	pVarRound                          uintptr
	pVarCmp                            uintptr
	pVarDecAdd                         uintptr
	pVarDecDiv                         uintptr
	pVarDecMul                         uintptr
	pVarDecSub                         uintptr
	pVarDecAbs                         uintptr
	pVarDecFix                         uintptr
	pVarDecInt                         uintptr
	pVarDecNeg                         uintptr
	pVarDecRound                       uintptr
	pVarDecCmp                         uintptr
	pVarDecCmpR8                       uintptr
	pVarCyAdd                          uintptr
	pVarCyMul                          uintptr
	pVarCyMulI4                        uintptr
	pVarCyMulI8                        uintptr
	pVarCySub                          uintptr
	pVarCyAbs                          uintptr
	pVarCyFix                          uintptr
	pVarCyInt                          uintptr
	pVarCyNeg                          uintptr
	pVarCyRound                        uintptr
	pVarCyCmp                          uintptr
	pVarCyCmpR8                        uintptr
	pVarBstrCat                        uintptr
	pVarBstrCmp                        uintptr
	pVarR8Pow                          uintptr
	pVarR4CmpR8                        uintptr
	pVarR8Round                        uintptr
	pVarDateFromUdate                  uintptr
	pVarDateFromUdateEx                uintptr
	pVarUdateFromDate                  uintptr
	pGetAltMonthNames                  uintptr
	pVarFormat                         uintptr
	pVarFormatDateTime                 uintptr
	pVarFormatNumber                   uintptr
	pVarFormatPercent                  uintptr
	pVarFormatCurrency                 uintptr
	pVarWeekdayName                    uintptr
	pVarMonthName                      uintptr
	pVarFormatFromTokens               uintptr
	pVarTokenizeFormatString           uintptr
	pLHashValOfNameSysA                uintptr
	pLHashValOfNameSys                 uintptr
	pLoadTypeLib                       uintptr
	pLoadTypeLibEx                     uintptr
	pLoadRegTypeLib                    uintptr
	pQueryPathOfRegTypeLib             uintptr
	pRegisterTypeLib                   uintptr
	pUnRegisterTypeLib                 uintptr
	pRegisterTypeLibForUser            uintptr
	pUnRegisterTypeLibForUser          uintptr
	pCreateTypeLib                     uintptr
	pCreateTypeLib2                    uintptr
	pDispGetParam                      uintptr
	pDispGetIDsOfNames                 uintptr
	pDispInvoke                        uintptr
	pCreateDispTypeInfo                uintptr
	pCreateStdDispatch                 uintptr
	pDispCallFunc                      uintptr
	pRegisterActiveObject              uintptr
	pRevokeActiveObject                uintptr
	pGetActiveObject                   uintptr
	pCreateErrorInfo                   uintptr
	pGetRecordInfoFromTypeInfo         uintptr
	pGetRecordInfoFromGuids            uintptr
	pOaBuildVersion                    uintptr
	pClearCustData                     uintptr
	pOaEnablePerUserTLibRegistration   uintptr
	pOleBuildVersion                   uintptr
	pOleInitialize                     uintptr
	pOleUninitialize                   uintptr
	pOleQueryLinkFromData              uintptr
	pOleQueryCreateFromData            uintptr
	pOleCreate                         uintptr
	pOleCreateEx                       uintptr
	pOleCreateFromData                 uintptr
	pOleCreateFromDataEx               uintptr
	pOleCreateLinkFromData             uintptr
	pOleCreateLinkFromDataEx           uintptr
	pOleCreateStaticFromData           uintptr
	pOleCreateLink                     uintptr
	pOleCreateLinkEx                   uintptr
	pOleCreateLinkToFile               uintptr
	pOleCreateLinkToFileEx             uintptr
	pOleCreateFromFile                 uintptr
	pOleCreateFromFileEx               uintptr
	pOleLoad                           uintptr
	pOleSave                           uintptr
	pOleLoadFromStream                 uintptr
	pOleSaveToStream                   uintptr
	pOleSetContainedObject             uintptr
	pOleNoteObjectVisible              uintptr
	pRegisterDragDrop                  uintptr
	pRevokeDragDrop                    uintptr
	pDoDragDrop                        uintptr
	pOleSetClipboard                   uintptr
	pOleGetClipboard                   uintptr
	pOleGetClipboardWithEnterpriseInfo uintptr
	pOleFlushClipboard                 uintptr
	pOleIsCurrentClipboard             uintptr
	pOleCreateMenuDescriptor           uintptr
	pOleSetMenuDescriptor              uintptr
	pOleDestroyMenuDescriptor          uintptr
	pOleTranslateAccelerator           uintptr
	pOleDuplicateData                  uintptr
	pOleDraw                           uintptr
	pOleRun                            uintptr
	pOleIsRunning                      uintptr
	pOleLockRunning                    uintptr
	pReleaseStgMedium                  uintptr
	pCreateOleAdviseHolder             uintptr
	pOleCreateDefaultHandler           uintptr
	pOleCreateEmbeddingHelper          uintptr
	pIsAccelerator                     uintptr
	pOleGetIconOfFile                  uintptr
	pOleGetIconOfClass                 uintptr
	pOleMetafilePictFromIconAndLabel   uintptr
	pOleRegGetUserType                 uintptr
	pOleRegGetMiscStatus               uintptr
	pOleRegEnumFormatEtc               uintptr
	pOleRegEnumVerbs                   uintptr
	pOleDoAutoConvert                  uintptr
	pOleGetAutoConvert                 uintptr
	pOleSetAutoConvert                 uintptr
	pHRGN_UserSize                     uintptr
	pHRGN_UserMarshal                  uintptr
	pHRGN_UserUnmarshal                uintptr
	pHRGN_UserFree                     uintptr
	pOleCreatePropertyFrame            uintptr
	pOleCreatePropertyFrameIndirect    uintptr
	pOleTranslateColor                 uintptr
	pOleCreateFontIndirect             uintptr
	pOleCreatePictureIndirect          uintptr
	pOleLoadPicture                    uintptr
	pOleLoadPictureEx                  uintptr
	pOleLoadPicturePath                uintptr
	pOleLoadPictureFile                uintptr
	pOleLoadPictureFileEx              uintptr
	pOleSavePictureFile                uintptr
	pOleIconToCursor                   uintptr
)

func DosDateTimeToVariantTime(wDosDate uint16, wDosTime uint16, pvtime *float64) int32 {
	addr := lazyAddr(&pDosDateTimeToVariantTime, libOleaut32, "DosDateTimeToVariantTime")
	ret, _, _ := syscall.SyscallN(addr, uintptr(wDosDate), uintptr(wDosTime), uintptr(unsafe.Pointer(pvtime)))
	return int32(ret)
}

func VariantTimeToDosDateTime(vtime float64, pwDosDate *uint16, pwDosTime *uint16) int32 {
	addr := lazyAddr(&pVariantTimeToDosDateTime, libOleaut32, "VariantTimeToDosDateTime")
	ret, _, _ := syscall.SyscallN(addr, uintptr(vtime), uintptr(unsafe.Pointer(pwDosDate)), uintptr(unsafe.Pointer(pwDosTime)))
	return int32(ret)
}

func SystemTimeToVariantTime(lpSystemTime *SYSTEMTIME, pvtime *float64) int32 {
	addr := lazyAddr(&pSystemTimeToVariantTime, libOleaut32, "SystemTimeToVariantTime")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpSystemTime)), uintptr(unsafe.Pointer(pvtime)))
	return int32(ret)
}

func VariantTimeToSystemTime(vtime float64, lpSystemTime *SYSTEMTIME) int32 {
	addr := lazyAddr(&pVariantTimeToSystemTime, libOleaut32, "VariantTimeToSystemTime")
	ret, _, _ := syscall.SyscallN(addr, uintptr(vtime), uintptr(unsafe.Pointer(lpSystemTime)))
	return int32(ret)
}

func SafeArrayAllocDescriptor(cDims uint32, ppsaOut **SAFEARRAY) HRESULT {
	addr := lazyAddr(&pSafeArrayAllocDescriptor, libOleaut32, "SafeArrayAllocDescriptor")
	ret, _, _ := syscall.SyscallN(addr, uintptr(cDims), uintptr(unsafe.Pointer(ppsaOut)))
	return HRESULT(ret)
}

func SafeArrayAllocDescriptorEx(vt uint16, cDims uint32, ppsaOut **SAFEARRAY) HRESULT {
	addr := lazyAddr(&pSafeArrayAllocDescriptorEx, libOleaut32, "SafeArrayAllocDescriptorEx")
	ret, _, _ := syscall.SyscallN(addr, uintptr(vt), uintptr(cDims), uintptr(unsafe.Pointer(ppsaOut)))
	return HRESULT(ret)
}

func SafeArrayAllocData(psa *SAFEARRAY) HRESULT {
	addr := lazyAddr(&pSafeArrayAllocData, libOleaut32, "SafeArrayAllocData")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(psa)))
	return HRESULT(ret)
}

func SafeArrayCreate(vt uint16, cDims uint32, rgsabound *SAFEARRAYBOUND) *SAFEARRAY {
	addr := lazyAddr(&pSafeArrayCreate, libOleaut32, "SafeArrayCreate")
	ret, _, _ := syscall.SyscallN(addr, uintptr(vt), uintptr(cDims), uintptr(unsafe.Pointer(rgsabound)))
	return (*SAFEARRAY)(unsafe.Pointer(ret))
}

func SafeArrayCreateEx(vt uint16, cDims uint32, rgsabound *SAFEARRAYBOUND, pvExtra unsafe.Pointer) *SAFEARRAY {
	addr := lazyAddr(&pSafeArrayCreateEx, libOleaut32, "SafeArrayCreateEx")
	ret, _, _ := syscall.SyscallN(addr, uintptr(vt), uintptr(cDims), uintptr(unsafe.Pointer(rgsabound)), uintptr(pvExtra))
	return (*SAFEARRAY)(unsafe.Pointer(ret))
}

func SafeArrayCopyData(psaSource *SAFEARRAY, psaTarget *SAFEARRAY) HRESULT {
	addr := lazyAddr(&pSafeArrayCopyData, libOleaut32, "SafeArrayCopyData")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(psaSource)), uintptr(unsafe.Pointer(psaTarget)))
	return HRESULT(ret)
}

func SafeArrayReleaseDescriptor(psa *SAFEARRAY) {
	addr := lazyAddr(&pSafeArrayReleaseDescriptor, libOleaut32, "SafeArrayReleaseDescriptor")
	syscall.SyscallN(addr, uintptr(unsafe.Pointer(psa)))
}

func SafeArrayDestroyDescriptor(psa *SAFEARRAY) HRESULT {
	addr := lazyAddr(&pSafeArrayDestroyDescriptor, libOleaut32, "SafeArrayDestroyDescriptor")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(psa)))
	return HRESULT(ret)
}

func SafeArrayReleaseData(pData unsafe.Pointer) {
	addr := lazyAddr(&pSafeArrayReleaseData, libOleaut32, "SafeArrayReleaseData")
	syscall.SyscallN(addr, uintptr(pData))
}

func SafeArrayDestroyData(psa *SAFEARRAY) HRESULT {
	addr := lazyAddr(&pSafeArrayDestroyData, libOleaut32, "SafeArrayDestroyData")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(psa)))
	return HRESULT(ret)
}

func SafeArrayAddRef(psa *SAFEARRAY, ppDataToRelease unsafe.Pointer) HRESULT {
	addr := lazyAddr(&pSafeArrayAddRef, libOleaut32, "SafeArrayAddRef")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(psa)), uintptr(ppDataToRelease))
	return HRESULT(ret)
}

func SafeArrayDestroy(psa *SAFEARRAY) HRESULT {
	addr := lazyAddr(&pSafeArrayDestroy, libOleaut32, "SafeArrayDestroy")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(psa)))
	return HRESULT(ret)
}

func SafeArrayRedim(psa *SAFEARRAY, psaboundNew *SAFEARRAYBOUND) HRESULT {
	addr := lazyAddr(&pSafeArrayRedim, libOleaut32, "SafeArrayRedim")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(psa)), uintptr(unsafe.Pointer(psaboundNew)))
	return HRESULT(ret)
}

func SafeArrayGetDim(psa *SAFEARRAY) uint32 {
	addr := lazyAddr(&pSafeArrayGetDim, libOleaut32, "SafeArrayGetDim")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(psa)))
	return uint32(ret)
}

func SafeArrayGetElemsize(psa *SAFEARRAY) uint32 {
	addr := lazyAddr(&pSafeArrayGetElemsize, libOleaut32, "SafeArrayGetElemsize")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(psa)))
	return uint32(ret)
}

func SafeArrayGetUBound(psa *SAFEARRAY, nDim uint32, plUbound *int32) HRESULT {
	addr := lazyAddr(&pSafeArrayGetUBound, libOleaut32, "SafeArrayGetUBound")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(psa)), uintptr(nDim), uintptr(unsafe.Pointer(plUbound)))
	return HRESULT(ret)
}

func SafeArrayGetLBound(psa *SAFEARRAY, nDim uint32, plLbound *int32) HRESULT {
	addr := lazyAddr(&pSafeArrayGetLBound, libOleaut32, "SafeArrayGetLBound")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(psa)), uintptr(nDim), uintptr(unsafe.Pointer(plLbound)))
	return HRESULT(ret)
}

func SafeArrayLock(psa *SAFEARRAY) HRESULT {
	addr := lazyAddr(&pSafeArrayLock, libOleaut32, "SafeArrayLock")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(psa)))
	return HRESULT(ret)
}

func SafeArrayUnlock(psa *SAFEARRAY) HRESULT {
	addr := lazyAddr(&pSafeArrayUnlock, libOleaut32, "SafeArrayUnlock")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(psa)))
	return HRESULT(ret)
}

func SafeArrayAccessData(psa *SAFEARRAY, ppvData unsafe.Pointer) HRESULT {
	addr := lazyAddr(&pSafeArrayAccessData, libOleaut32, "SafeArrayAccessData")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(psa)), uintptr(ppvData))
	return HRESULT(ret)
}

func SafeArrayUnaccessData(psa *SAFEARRAY) HRESULT {
	addr := lazyAddr(&pSafeArrayUnaccessData, libOleaut32, "SafeArrayUnaccessData")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(psa)))
	return HRESULT(ret)
}

func SafeArrayGetElement(psa *SAFEARRAY, rgIndices *int32, pv unsafe.Pointer) HRESULT {
	addr := lazyAddr(&pSafeArrayGetElement, libOleaut32, "SafeArrayGetElement")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(psa)), uintptr(unsafe.Pointer(rgIndices)), uintptr(pv))
	return HRESULT(ret)
}

func SafeArrayPutElement(psa *SAFEARRAY, rgIndices *int32, pv unsafe.Pointer) HRESULT {
	addr := lazyAddr(&pSafeArrayPutElement, libOleaut32, "SafeArrayPutElement")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(psa)), uintptr(unsafe.Pointer(rgIndices)), uintptr(pv))
	return HRESULT(ret)
}

func SafeArrayCopy(psa *SAFEARRAY, ppsaOut **SAFEARRAY) HRESULT {
	addr := lazyAddr(&pSafeArrayCopy, libOleaut32, "SafeArrayCopy")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(psa)), uintptr(unsafe.Pointer(ppsaOut)))
	return HRESULT(ret)
}

func SafeArrayPtrOfIndex(psa *SAFEARRAY, rgIndices *int32, ppvData unsafe.Pointer) HRESULT {
	addr := lazyAddr(&pSafeArrayPtrOfIndex, libOleaut32, "SafeArrayPtrOfIndex")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(psa)), uintptr(unsafe.Pointer(rgIndices)), uintptr(ppvData))
	return HRESULT(ret)
}

func SafeArraySetRecordInfo(psa *SAFEARRAY, prinfo *IRecordInfo) HRESULT {
	addr := lazyAddr(&pSafeArraySetRecordInfo, libOleaut32, "SafeArraySetRecordInfo")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(psa)), uintptr(unsafe.Pointer(prinfo)))
	return HRESULT(ret)
}

func SafeArrayGetRecordInfo(psa *SAFEARRAY, prinfo **IRecordInfo) HRESULT {
	addr := lazyAddr(&pSafeArrayGetRecordInfo, libOleaut32, "SafeArrayGetRecordInfo")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(psa)), uintptr(unsafe.Pointer(prinfo)))
	return HRESULT(ret)
}

func SafeArraySetIID(psa *SAFEARRAY, guid *syscall.GUID) HRESULT {
	addr := lazyAddr(&pSafeArraySetIID, libOleaut32, "SafeArraySetIID")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(psa)), uintptr(unsafe.Pointer(guid)))
	return HRESULT(ret)
}

func SafeArrayGetIID(psa *SAFEARRAY, pguid *syscall.GUID) HRESULT {
	addr := lazyAddr(&pSafeArrayGetIID, libOleaut32, "SafeArrayGetIID")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(psa)), uintptr(unsafe.Pointer(pguid)))
	return HRESULT(ret)
}

func SafeArrayGetVartype(psa *SAFEARRAY, pvt *uint16) HRESULT {
	addr := lazyAddr(&pSafeArrayGetVartype, libOleaut32, "SafeArrayGetVartype")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(psa)), uintptr(unsafe.Pointer(pvt)))
	return HRESULT(ret)
}

func SafeArrayCreateVector(vt uint16, lLbound int32, cElements uint32) *SAFEARRAY {
	addr := lazyAddr(&pSafeArrayCreateVector, libOleaut32, "SafeArrayCreateVector")
	ret, _, _ := syscall.SyscallN(addr, uintptr(vt), uintptr(lLbound), uintptr(cElements))
	return (*SAFEARRAY)(unsafe.Pointer(ret))
}

func SafeArrayCreateVectorEx(vt uint16, lLbound int32, cElements uint32, pvExtra unsafe.Pointer) *SAFEARRAY {
	addr := lazyAddr(&pSafeArrayCreateVectorEx, libOleaut32, "SafeArrayCreateVectorEx")
	ret, _, _ := syscall.SyscallN(addr, uintptr(vt), uintptr(lLbound), uintptr(cElements), uintptr(pvExtra))
	return (*SAFEARRAY)(unsafe.Pointer(ret))
}

func VariantInit(pvarg *VARIANT) {
	addr := lazyAddr(&pVariantInit, libOleaut32, "VariantInit")
	syscall.SyscallN(addr, uintptr(unsafe.Pointer(pvarg)))
}

func VariantClear(pvarg *VARIANT) HRESULT {
	addr := lazyAddr(&pVariantClear, libOleaut32, "VariantClear")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pvarg)))
	return HRESULT(ret)
}

func VariantCopy(pvargDest *VARIANT, pvargSrc *VARIANT) HRESULT {
	addr := lazyAddr(&pVariantCopy, libOleaut32, "VariantCopy")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pvargDest)), uintptr(unsafe.Pointer(pvargSrc)))
	return HRESULT(ret)
}

func VariantCopyInd(pvarDest *VARIANT, pvargSrc *VARIANT) HRESULT {
	addr := lazyAddr(&pVariantCopyInd, libOleaut32, "VariantCopyInd")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pvarDest)), uintptr(unsafe.Pointer(pvargSrc)))
	return HRESULT(ret)
}

func VariantChangeType(pvargDest *VARIANT, pvarSrc *VARIANT, wFlags uint16, vt uint16) HRESULT {
	addr := lazyAddr(&pVariantChangeType, libOleaut32, "VariantChangeType")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pvargDest)), uintptr(unsafe.Pointer(pvarSrc)), uintptr(wFlags), uintptr(vt))
	return HRESULT(ret)
}

func VariantChangeTypeEx(pvargDest *VARIANT, pvarSrc *VARIANT, lcid uint32, wFlags uint16, vt uint16) HRESULT {
	addr := lazyAddr(&pVariantChangeTypeEx, libOleaut32, "VariantChangeTypeEx")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pvargDest)), uintptr(unsafe.Pointer(pvarSrc)), uintptr(lcid), uintptr(wFlags), uintptr(vt))
	return HRESULT(ret)
}

func VectorFromBstr(bstr BSTR, ppsa **SAFEARRAY) HRESULT {
	addr := lazyAddr(&pVectorFromBstr, libOleaut32, "VectorFromBstr")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(bstr)), uintptr(unsafe.Pointer(ppsa)))
	return HRESULT(ret)
}

func BstrFromVector(psa *SAFEARRAY, pbstr *BSTR) HRESULT {
	addr := lazyAddr(&pBstrFromVector, libOleaut32, "BstrFromVector")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(psa)), uintptr(unsafe.Pointer(pbstr)))
	return HRESULT(ret)
}

func VarUI1FromI2(sIn int16, pbOut *byte) HRESULT {
	addr := lazyAddr(&pVarUI1FromI2, libOleaut32, "VarUI1FromI2")
	ret, _, _ := syscall.SyscallN(addr, uintptr(sIn), uintptr(unsafe.Pointer(pbOut)))
	return HRESULT(ret)
}

func VarUI1FromI4(lIn int32, pbOut *byte) HRESULT {
	addr := lazyAddr(&pVarUI1FromI4, libOleaut32, "VarUI1FromI4")
	ret, _, _ := syscall.SyscallN(addr, uintptr(lIn), uintptr(unsafe.Pointer(pbOut)))
	return HRESULT(ret)
}

func VarUI1FromI8(i64In int64, pbOut *byte) HRESULT {
	addr := lazyAddr(&pVarUI1FromI8, libOleaut32, "VarUI1FromI8")
	ret, _, _ := syscall.SyscallN(addr, uintptr(i64In), uintptr(unsafe.Pointer(pbOut)))
	return HRESULT(ret)
}

func VarUI1FromR4(fltIn float32, pbOut *byte) HRESULT {
	addr := lazyAddr(&pVarUI1FromR4, libOleaut32, "VarUI1FromR4")
	ret, _, _ := syscall.SyscallN(addr, uintptr(fltIn), uintptr(unsafe.Pointer(pbOut)))
	return HRESULT(ret)
}

func VarUI1FromR8(dblIn float64, pbOut *byte) HRESULT {
	addr := lazyAddr(&pVarUI1FromR8, libOleaut32, "VarUI1FromR8")
	ret, _, _ := syscall.SyscallN(addr, uintptr(dblIn), uintptr(unsafe.Pointer(pbOut)))
	return HRESULT(ret)
}

func VarUI1FromCy(cyIn CY, pbOut *byte) HRESULT {
	addr := lazyAddr(&pVarUI1FromCy, libOleaut32, "VarUI1FromCy")
	ret, _, _ := syscall.SyscallN(addr, *(*uintptr)(unsafe.Pointer(&cyIn)), uintptr(unsafe.Pointer(pbOut)))
	return HRESULT(ret)
}

func VarUI1FromDate(dateIn float64, pbOut *byte) HRESULT {
	addr := lazyAddr(&pVarUI1FromDate, libOleaut32, "VarUI1FromDate")
	ret, _, _ := syscall.SyscallN(addr, uintptr(dateIn), uintptr(unsafe.Pointer(pbOut)))
	return HRESULT(ret)
}

func VarUI1FromStr(strIn PWSTR, lcid uint32, dwFlags uint32, pbOut *byte) HRESULT {
	addr := lazyAddr(&pVarUI1FromStr, libOleaut32, "VarUI1FromStr")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(strIn)), uintptr(lcid), uintptr(dwFlags), uintptr(unsafe.Pointer(pbOut)))
	return HRESULT(ret)
}

func VarUI1FromDisp(pdispIn *IDispatch, lcid uint32, pbOut *byte) HRESULT {
	addr := lazyAddr(&pVarUI1FromDisp, libOleaut32, "VarUI1FromDisp")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pdispIn)), uintptr(lcid), uintptr(unsafe.Pointer(pbOut)))
	return HRESULT(ret)
}

func VarUI1FromBool(boolIn int16, pbOut *byte) HRESULT {
	addr := lazyAddr(&pVarUI1FromBool, libOleaut32, "VarUI1FromBool")
	ret, _, _ := syscall.SyscallN(addr, uintptr(boolIn), uintptr(unsafe.Pointer(pbOut)))
	return HRESULT(ret)
}

func VarUI1FromI1(cIn CHAR, pbOut *byte) HRESULT {
	addr := lazyAddr(&pVarUI1FromI1, libOleaut32, "VarUI1FromI1")
	ret, _, _ := syscall.SyscallN(addr, uintptr(cIn), uintptr(unsafe.Pointer(pbOut)))
	return HRESULT(ret)
}

func VarUI1FromUI2(uiIn uint16, pbOut *byte) HRESULT {
	addr := lazyAddr(&pVarUI1FromUI2, libOleaut32, "VarUI1FromUI2")
	ret, _, _ := syscall.SyscallN(addr, uintptr(uiIn), uintptr(unsafe.Pointer(pbOut)))
	return HRESULT(ret)
}

func VarUI1FromUI4(ulIn uint32, pbOut *byte) HRESULT {
	addr := lazyAddr(&pVarUI1FromUI4, libOleaut32, "VarUI1FromUI4")
	ret, _, _ := syscall.SyscallN(addr, uintptr(ulIn), uintptr(unsafe.Pointer(pbOut)))
	return HRESULT(ret)
}

func VarUI1FromUI8(ui64In uint64, pbOut *byte) HRESULT {
	addr := lazyAddr(&pVarUI1FromUI8, libOleaut32, "VarUI1FromUI8")
	ret, _, _ := syscall.SyscallN(addr, uintptr(ui64In), uintptr(unsafe.Pointer(pbOut)))
	return HRESULT(ret)
}

func VarUI1FromDec(pdecIn *DECIMAL, pbOut *byte) HRESULT {
	addr := lazyAddr(&pVarUI1FromDec, libOleaut32, "VarUI1FromDec")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pdecIn)), uintptr(unsafe.Pointer(pbOut)))
	return HRESULT(ret)
}

func VarI2FromUI1(bIn byte, psOut *int16) HRESULT {
	addr := lazyAddr(&pVarI2FromUI1, libOleaut32, "VarI2FromUI1")
	ret, _, _ := syscall.SyscallN(addr, uintptr(bIn), uintptr(unsafe.Pointer(psOut)))
	return HRESULT(ret)
}

func VarI2FromI4(lIn int32, psOut *int16) HRESULT {
	addr := lazyAddr(&pVarI2FromI4, libOleaut32, "VarI2FromI4")
	ret, _, _ := syscall.SyscallN(addr, uintptr(lIn), uintptr(unsafe.Pointer(psOut)))
	return HRESULT(ret)
}

func VarI2FromI8(i64In int64, psOut *int16) HRESULT {
	addr := lazyAddr(&pVarI2FromI8, libOleaut32, "VarI2FromI8")
	ret, _, _ := syscall.SyscallN(addr, uintptr(i64In), uintptr(unsafe.Pointer(psOut)))
	return HRESULT(ret)
}

func VarI2FromR4(fltIn float32, psOut *int16) HRESULT {
	addr := lazyAddr(&pVarI2FromR4, libOleaut32, "VarI2FromR4")
	ret, _, _ := syscall.SyscallN(addr, uintptr(fltIn), uintptr(unsafe.Pointer(psOut)))
	return HRESULT(ret)
}

func VarI2FromR8(dblIn float64, psOut *int16) HRESULT {
	addr := lazyAddr(&pVarI2FromR8, libOleaut32, "VarI2FromR8")
	ret, _, _ := syscall.SyscallN(addr, uintptr(dblIn), uintptr(unsafe.Pointer(psOut)))
	return HRESULT(ret)
}

func VarI2FromCy(cyIn CY, psOut *int16) HRESULT {
	addr := lazyAddr(&pVarI2FromCy, libOleaut32, "VarI2FromCy")
	ret, _, _ := syscall.SyscallN(addr, *(*uintptr)(unsafe.Pointer(&cyIn)), uintptr(unsafe.Pointer(psOut)))
	return HRESULT(ret)
}

func VarI2FromDate(dateIn float64, psOut *int16) HRESULT {
	addr := lazyAddr(&pVarI2FromDate, libOleaut32, "VarI2FromDate")
	ret, _, _ := syscall.SyscallN(addr, uintptr(dateIn), uintptr(unsafe.Pointer(psOut)))
	return HRESULT(ret)
}

func VarI2FromStr(strIn PWSTR, lcid uint32, dwFlags uint32, psOut *int16) HRESULT {
	addr := lazyAddr(&pVarI2FromStr, libOleaut32, "VarI2FromStr")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(strIn)), uintptr(lcid), uintptr(dwFlags), uintptr(unsafe.Pointer(psOut)))
	return HRESULT(ret)
}

func VarI2FromDisp(pdispIn *IDispatch, lcid uint32, psOut *int16) HRESULT {
	addr := lazyAddr(&pVarI2FromDisp, libOleaut32, "VarI2FromDisp")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pdispIn)), uintptr(lcid), uintptr(unsafe.Pointer(psOut)))
	return HRESULT(ret)
}

func VarI2FromBool(boolIn int16, psOut *int16) HRESULT {
	addr := lazyAddr(&pVarI2FromBool, libOleaut32, "VarI2FromBool")
	ret, _, _ := syscall.SyscallN(addr, uintptr(boolIn), uintptr(unsafe.Pointer(psOut)))
	return HRESULT(ret)
}

func VarI2FromI1(cIn CHAR, psOut *int16) HRESULT {
	addr := lazyAddr(&pVarI2FromI1, libOleaut32, "VarI2FromI1")
	ret, _, _ := syscall.SyscallN(addr, uintptr(cIn), uintptr(unsafe.Pointer(psOut)))
	return HRESULT(ret)
}

func VarI2FromUI2(uiIn uint16, psOut *int16) HRESULT {
	addr := lazyAddr(&pVarI2FromUI2, libOleaut32, "VarI2FromUI2")
	ret, _, _ := syscall.SyscallN(addr, uintptr(uiIn), uintptr(unsafe.Pointer(psOut)))
	return HRESULT(ret)
}

func VarI2FromUI4(ulIn uint32, psOut *int16) HRESULT {
	addr := lazyAddr(&pVarI2FromUI4, libOleaut32, "VarI2FromUI4")
	ret, _, _ := syscall.SyscallN(addr, uintptr(ulIn), uintptr(unsafe.Pointer(psOut)))
	return HRESULT(ret)
}

func VarI2FromUI8(ui64In uint64, psOut *int16) HRESULT {
	addr := lazyAddr(&pVarI2FromUI8, libOleaut32, "VarI2FromUI8")
	ret, _, _ := syscall.SyscallN(addr, uintptr(ui64In), uintptr(unsafe.Pointer(psOut)))
	return HRESULT(ret)
}

func VarI2FromDec(pdecIn *DECIMAL, psOut *int16) HRESULT {
	addr := lazyAddr(&pVarI2FromDec, libOleaut32, "VarI2FromDec")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pdecIn)), uintptr(unsafe.Pointer(psOut)))
	return HRESULT(ret)
}

func VarI4FromUI1(bIn byte, plOut *int32) HRESULT {
	addr := lazyAddr(&pVarI4FromUI1, libOleaut32, "VarI4FromUI1")
	ret, _, _ := syscall.SyscallN(addr, uintptr(bIn), uintptr(unsafe.Pointer(plOut)))
	return HRESULT(ret)
}

func VarI4FromI2(sIn int16, plOut *int32) HRESULT {
	addr := lazyAddr(&pVarI4FromI2, libOleaut32, "VarI4FromI2")
	ret, _, _ := syscall.SyscallN(addr, uintptr(sIn), uintptr(unsafe.Pointer(plOut)))
	return HRESULT(ret)
}

func VarI4FromI8(i64In int64, plOut *int32) HRESULT {
	addr := lazyAddr(&pVarI4FromI8, libOleaut32, "VarI4FromI8")
	ret, _, _ := syscall.SyscallN(addr, uintptr(i64In), uintptr(unsafe.Pointer(plOut)))
	return HRESULT(ret)
}

func VarI4FromR4(fltIn float32, plOut *int32) HRESULT {
	addr := lazyAddr(&pVarI4FromR4, libOleaut32, "VarI4FromR4")
	ret, _, _ := syscall.SyscallN(addr, uintptr(fltIn), uintptr(unsafe.Pointer(plOut)))
	return HRESULT(ret)
}

func VarI4FromR8(dblIn float64, plOut *int32) HRESULT {
	addr := lazyAddr(&pVarI4FromR8, libOleaut32, "VarI4FromR8")
	ret, _, _ := syscall.SyscallN(addr, uintptr(dblIn), uintptr(unsafe.Pointer(plOut)))
	return HRESULT(ret)
}

func VarI4FromCy(cyIn CY, plOut *int32) HRESULT {
	addr := lazyAddr(&pVarI4FromCy, libOleaut32, "VarI4FromCy")
	ret, _, _ := syscall.SyscallN(addr, *(*uintptr)(unsafe.Pointer(&cyIn)), uintptr(unsafe.Pointer(plOut)))
	return HRESULT(ret)
}

func VarI4FromDate(dateIn float64, plOut *int32) HRESULT {
	addr := lazyAddr(&pVarI4FromDate, libOleaut32, "VarI4FromDate")
	ret, _, _ := syscall.SyscallN(addr, uintptr(dateIn), uintptr(unsafe.Pointer(plOut)))
	return HRESULT(ret)
}

func VarI4FromStr(strIn PWSTR, lcid uint32, dwFlags uint32, plOut *int32) HRESULT {
	addr := lazyAddr(&pVarI4FromStr, libOleaut32, "VarI4FromStr")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(strIn)), uintptr(lcid), uintptr(dwFlags), uintptr(unsafe.Pointer(plOut)))
	return HRESULT(ret)
}

func VarI4FromDisp(pdispIn *IDispatch, lcid uint32, plOut *int32) HRESULT {
	addr := lazyAddr(&pVarI4FromDisp, libOleaut32, "VarI4FromDisp")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pdispIn)), uintptr(lcid), uintptr(unsafe.Pointer(plOut)))
	return HRESULT(ret)
}

func VarI4FromBool(boolIn int16, plOut *int32) HRESULT {
	addr := lazyAddr(&pVarI4FromBool, libOleaut32, "VarI4FromBool")
	ret, _, _ := syscall.SyscallN(addr, uintptr(boolIn), uintptr(unsafe.Pointer(plOut)))
	return HRESULT(ret)
}

func VarI4FromI1(cIn CHAR, plOut *int32) HRESULT {
	addr := lazyAddr(&pVarI4FromI1, libOleaut32, "VarI4FromI1")
	ret, _, _ := syscall.SyscallN(addr, uintptr(cIn), uintptr(unsafe.Pointer(plOut)))
	return HRESULT(ret)
}

func VarI4FromUI2(uiIn uint16, plOut *int32) HRESULT {
	addr := lazyAddr(&pVarI4FromUI2, libOleaut32, "VarI4FromUI2")
	ret, _, _ := syscall.SyscallN(addr, uintptr(uiIn), uintptr(unsafe.Pointer(plOut)))
	return HRESULT(ret)
}

func VarI4FromUI4(ulIn uint32, plOut *int32) HRESULT {
	addr := lazyAddr(&pVarI4FromUI4, libOleaut32, "VarI4FromUI4")
	ret, _, _ := syscall.SyscallN(addr, uintptr(ulIn), uintptr(unsafe.Pointer(plOut)))
	return HRESULT(ret)
}

func VarI4FromUI8(ui64In uint64, plOut *int32) HRESULT {
	addr := lazyAddr(&pVarI4FromUI8, libOleaut32, "VarI4FromUI8")
	ret, _, _ := syscall.SyscallN(addr, uintptr(ui64In), uintptr(unsafe.Pointer(plOut)))
	return HRESULT(ret)
}

func VarI4FromDec(pdecIn *DECIMAL, plOut *int32) HRESULT {
	addr := lazyAddr(&pVarI4FromDec, libOleaut32, "VarI4FromDec")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pdecIn)), uintptr(unsafe.Pointer(plOut)))
	return HRESULT(ret)
}

func VarI8FromUI1(bIn byte, pi64Out *int64) HRESULT {
	addr := lazyAddr(&pVarI8FromUI1, libOleaut32, "VarI8FromUI1")
	ret, _, _ := syscall.SyscallN(addr, uintptr(bIn), uintptr(unsafe.Pointer(pi64Out)))
	return HRESULT(ret)
}

func VarI8FromI2(sIn int16, pi64Out *int64) HRESULT {
	addr := lazyAddr(&pVarI8FromI2, libOleaut32, "VarI8FromI2")
	ret, _, _ := syscall.SyscallN(addr, uintptr(sIn), uintptr(unsafe.Pointer(pi64Out)))
	return HRESULT(ret)
}

func VarI8FromR4(fltIn float32, pi64Out *int64) HRESULT {
	addr := lazyAddr(&pVarI8FromR4, libOleaut32, "VarI8FromR4")
	ret, _, _ := syscall.SyscallN(addr, uintptr(fltIn), uintptr(unsafe.Pointer(pi64Out)))
	return HRESULT(ret)
}

func VarI8FromR8(dblIn float64, pi64Out *int64) HRESULT {
	addr := lazyAddr(&pVarI8FromR8, libOleaut32, "VarI8FromR8")
	ret, _, _ := syscall.SyscallN(addr, uintptr(dblIn), uintptr(unsafe.Pointer(pi64Out)))
	return HRESULT(ret)
}

func VarI8FromCy(cyIn CY, pi64Out *int64) HRESULT {
	addr := lazyAddr(&pVarI8FromCy, libOleaut32, "VarI8FromCy")
	ret, _, _ := syscall.SyscallN(addr, *(*uintptr)(unsafe.Pointer(&cyIn)), uintptr(unsafe.Pointer(pi64Out)))
	return HRESULT(ret)
}

func VarI8FromDate(dateIn float64, pi64Out *int64) HRESULT {
	addr := lazyAddr(&pVarI8FromDate, libOleaut32, "VarI8FromDate")
	ret, _, _ := syscall.SyscallN(addr, uintptr(dateIn), uintptr(unsafe.Pointer(pi64Out)))
	return HRESULT(ret)
}

func VarI8FromStr(strIn PWSTR, lcid uint32, dwFlags uint32, pi64Out *int64) HRESULT {
	addr := lazyAddr(&pVarI8FromStr, libOleaut32, "VarI8FromStr")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(strIn)), uintptr(lcid), uintptr(dwFlags), uintptr(unsafe.Pointer(pi64Out)))
	return HRESULT(ret)
}

func VarI8FromDisp(pdispIn *IDispatch, lcid uint32, pi64Out *int64) HRESULT {
	addr := lazyAddr(&pVarI8FromDisp, libOleaut32, "VarI8FromDisp")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pdispIn)), uintptr(lcid), uintptr(unsafe.Pointer(pi64Out)))
	return HRESULT(ret)
}

func VarI8FromBool(boolIn int16, pi64Out *int64) HRESULT {
	addr := lazyAddr(&pVarI8FromBool, libOleaut32, "VarI8FromBool")
	ret, _, _ := syscall.SyscallN(addr, uintptr(boolIn), uintptr(unsafe.Pointer(pi64Out)))
	return HRESULT(ret)
}

func VarI8FromI1(cIn CHAR, pi64Out *int64) HRESULT {
	addr := lazyAddr(&pVarI8FromI1, libOleaut32, "VarI8FromI1")
	ret, _, _ := syscall.SyscallN(addr, uintptr(cIn), uintptr(unsafe.Pointer(pi64Out)))
	return HRESULT(ret)
}

func VarI8FromUI2(uiIn uint16, pi64Out *int64) HRESULT {
	addr := lazyAddr(&pVarI8FromUI2, libOleaut32, "VarI8FromUI2")
	ret, _, _ := syscall.SyscallN(addr, uintptr(uiIn), uintptr(unsafe.Pointer(pi64Out)))
	return HRESULT(ret)
}

func VarI8FromUI4(ulIn uint32, pi64Out *int64) HRESULT {
	addr := lazyAddr(&pVarI8FromUI4, libOleaut32, "VarI8FromUI4")
	ret, _, _ := syscall.SyscallN(addr, uintptr(ulIn), uintptr(unsafe.Pointer(pi64Out)))
	return HRESULT(ret)
}

func VarI8FromUI8(ui64In uint64, pi64Out *int64) HRESULT {
	addr := lazyAddr(&pVarI8FromUI8, libOleaut32, "VarI8FromUI8")
	ret, _, _ := syscall.SyscallN(addr, uintptr(ui64In), uintptr(unsafe.Pointer(pi64Out)))
	return HRESULT(ret)
}

func VarI8FromDec(pdecIn *DECIMAL, pi64Out *int64) HRESULT {
	addr := lazyAddr(&pVarI8FromDec, libOleaut32, "VarI8FromDec")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pdecIn)), uintptr(unsafe.Pointer(pi64Out)))
	return HRESULT(ret)
}

func VarR4FromUI1(bIn byte, pfltOut *float32) HRESULT {
	addr := lazyAddr(&pVarR4FromUI1, libOleaut32, "VarR4FromUI1")
	ret, _, _ := syscall.SyscallN(addr, uintptr(bIn), uintptr(unsafe.Pointer(pfltOut)))
	return HRESULT(ret)
}

func VarR4FromI2(sIn int16, pfltOut *float32) HRESULT {
	addr := lazyAddr(&pVarR4FromI2, libOleaut32, "VarR4FromI2")
	ret, _, _ := syscall.SyscallN(addr, uintptr(sIn), uintptr(unsafe.Pointer(pfltOut)))
	return HRESULT(ret)
}

func VarR4FromI4(lIn int32, pfltOut *float32) HRESULT {
	addr := lazyAddr(&pVarR4FromI4, libOleaut32, "VarR4FromI4")
	ret, _, _ := syscall.SyscallN(addr, uintptr(lIn), uintptr(unsafe.Pointer(pfltOut)))
	return HRESULT(ret)
}

func VarR4FromI8(i64In int64, pfltOut *float32) HRESULT {
	addr := lazyAddr(&pVarR4FromI8, libOleaut32, "VarR4FromI8")
	ret, _, _ := syscall.SyscallN(addr, uintptr(i64In), uintptr(unsafe.Pointer(pfltOut)))
	return HRESULT(ret)
}

func VarR4FromR8(dblIn float64, pfltOut *float32) HRESULT {
	addr := lazyAddr(&pVarR4FromR8, libOleaut32, "VarR4FromR8")
	ret, _, _ := syscall.SyscallN(addr, uintptr(dblIn), uintptr(unsafe.Pointer(pfltOut)))
	return HRESULT(ret)
}

func VarR4FromCy(cyIn CY, pfltOut *float32) HRESULT {
	addr := lazyAddr(&pVarR4FromCy, libOleaut32, "VarR4FromCy")
	ret, _, _ := syscall.SyscallN(addr, *(*uintptr)(unsafe.Pointer(&cyIn)), uintptr(unsafe.Pointer(pfltOut)))
	return HRESULT(ret)
}

func VarR4FromDate(dateIn float64, pfltOut *float32) HRESULT {
	addr := lazyAddr(&pVarR4FromDate, libOleaut32, "VarR4FromDate")
	ret, _, _ := syscall.SyscallN(addr, uintptr(dateIn), uintptr(unsafe.Pointer(pfltOut)))
	return HRESULT(ret)
}

func VarR4FromStr(strIn PWSTR, lcid uint32, dwFlags uint32, pfltOut *float32) HRESULT {
	addr := lazyAddr(&pVarR4FromStr, libOleaut32, "VarR4FromStr")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(strIn)), uintptr(lcid), uintptr(dwFlags), uintptr(unsafe.Pointer(pfltOut)))
	return HRESULT(ret)
}

func VarR4FromDisp(pdispIn *IDispatch, lcid uint32, pfltOut *float32) HRESULT {
	addr := lazyAddr(&pVarR4FromDisp, libOleaut32, "VarR4FromDisp")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pdispIn)), uintptr(lcid), uintptr(unsafe.Pointer(pfltOut)))
	return HRESULT(ret)
}

func VarR4FromBool(boolIn int16, pfltOut *float32) HRESULT {
	addr := lazyAddr(&pVarR4FromBool, libOleaut32, "VarR4FromBool")
	ret, _, _ := syscall.SyscallN(addr, uintptr(boolIn), uintptr(unsafe.Pointer(pfltOut)))
	return HRESULT(ret)
}

func VarR4FromI1(cIn CHAR, pfltOut *float32) HRESULT {
	addr := lazyAddr(&pVarR4FromI1, libOleaut32, "VarR4FromI1")
	ret, _, _ := syscall.SyscallN(addr, uintptr(cIn), uintptr(unsafe.Pointer(pfltOut)))
	return HRESULT(ret)
}

func VarR4FromUI2(uiIn uint16, pfltOut *float32) HRESULT {
	addr := lazyAddr(&pVarR4FromUI2, libOleaut32, "VarR4FromUI2")
	ret, _, _ := syscall.SyscallN(addr, uintptr(uiIn), uintptr(unsafe.Pointer(pfltOut)))
	return HRESULT(ret)
}

func VarR4FromUI4(ulIn uint32, pfltOut *float32) HRESULT {
	addr := lazyAddr(&pVarR4FromUI4, libOleaut32, "VarR4FromUI4")
	ret, _, _ := syscall.SyscallN(addr, uintptr(ulIn), uintptr(unsafe.Pointer(pfltOut)))
	return HRESULT(ret)
}

func VarR4FromUI8(ui64In uint64, pfltOut *float32) HRESULT {
	addr := lazyAddr(&pVarR4FromUI8, libOleaut32, "VarR4FromUI8")
	ret, _, _ := syscall.SyscallN(addr, uintptr(ui64In), uintptr(unsafe.Pointer(pfltOut)))
	return HRESULT(ret)
}

func VarR4FromDec(pdecIn *DECIMAL, pfltOut *float32) HRESULT {
	addr := lazyAddr(&pVarR4FromDec, libOleaut32, "VarR4FromDec")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pdecIn)), uintptr(unsafe.Pointer(pfltOut)))
	return HRESULT(ret)
}

func VarR8FromUI1(bIn byte, pdblOut *float64) HRESULT {
	addr := lazyAddr(&pVarR8FromUI1, libOleaut32, "VarR8FromUI1")
	ret, _, _ := syscall.SyscallN(addr, uintptr(bIn), uintptr(unsafe.Pointer(pdblOut)))
	return HRESULT(ret)
}

func VarR8FromI2(sIn int16, pdblOut *float64) HRESULT {
	addr := lazyAddr(&pVarR8FromI2, libOleaut32, "VarR8FromI2")
	ret, _, _ := syscall.SyscallN(addr, uintptr(sIn), uintptr(unsafe.Pointer(pdblOut)))
	return HRESULT(ret)
}

func VarR8FromI4(lIn int32, pdblOut *float64) HRESULT {
	addr := lazyAddr(&pVarR8FromI4, libOleaut32, "VarR8FromI4")
	ret, _, _ := syscall.SyscallN(addr, uintptr(lIn), uintptr(unsafe.Pointer(pdblOut)))
	return HRESULT(ret)
}

func VarR8FromI8(i64In int64, pdblOut *float64) HRESULT {
	addr := lazyAddr(&pVarR8FromI8, libOleaut32, "VarR8FromI8")
	ret, _, _ := syscall.SyscallN(addr, uintptr(i64In), uintptr(unsafe.Pointer(pdblOut)))
	return HRESULT(ret)
}

func VarR8FromR4(fltIn float32, pdblOut *float64) HRESULT {
	addr := lazyAddr(&pVarR8FromR4, libOleaut32, "VarR8FromR4")
	ret, _, _ := syscall.SyscallN(addr, uintptr(fltIn), uintptr(unsafe.Pointer(pdblOut)))
	return HRESULT(ret)
}

func VarR8FromCy(cyIn CY, pdblOut *float64) HRESULT {
	addr := lazyAddr(&pVarR8FromCy, libOleaut32, "VarR8FromCy")
	ret, _, _ := syscall.SyscallN(addr, *(*uintptr)(unsafe.Pointer(&cyIn)), uintptr(unsafe.Pointer(pdblOut)))
	return HRESULT(ret)
}

func VarR8FromDate(dateIn float64, pdblOut *float64) HRESULT {
	addr := lazyAddr(&pVarR8FromDate, libOleaut32, "VarR8FromDate")
	ret, _, _ := syscall.SyscallN(addr, uintptr(dateIn), uintptr(unsafe.Pointer(pdblOut)))
	return HRESULT(ret)
}

func VarR8FromStr(strIn PWSTR, lcid uint32, dwFlags uint32, pdblOut *float64) HRESULT {
	addr := lazyAddr(&pVarR8FromStr, libOleaut32, "VarR8FromStr")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(strIn)), uintptr(lcid), uintptr(dwFlags), uintptr(unsafe.Pointer(pdblOut)))
	return HRESULT(ret)
}

func VarR8FromDisp(pdispIn *IDispatch, lcid uint32, pdblOut *float64) HRESULT {
	addr := lazyAddr(&pVarR8FromDisp, libOleaut32, "VarR8FromDisp")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pdispIn)), uintptr(lcid), uintptr(unsafe.Pointer(pdblOut)))
	return HRESULT(ret)
}

func VarR8FromBool(boolIn int16, pdblOut *float64) HRESULT {
	addr := lazyAddr(&pVarR8FromBool, libOleaut32, "VarR8FromBool")
	ret, _, _ := syscall.SyscallN(addr, uintptr(boolIn), uintptr(unsafe.Pointer(pdblOut)))
	return HRESULT(ret)
}

func VarR8FromI1(cIn CHAR, pdblOut *float64) HRESULT {
	addr := lazyAddr(&pVarR8FromI1, libOleaut32, "VarR8FromI1")
	ret, _, _ := syscall.SyscallN(addr, uintptr(cIn), uintptr(unsafe.Pointer(pdblOut)))
	return HRESULT(ret)
}

func VarR8FromUI2(uiIn uint16, pdblOut *float64) HRESULT {
	addr := lazyAddr(&pVarR8FromUI2, libOleaut32, "VarR8FromUI2")
	ret, _, _ := syscall.SyscallN(addr, uintptr(uiIn), uintptr(unsafe.Pointer(pdblOut)))
	return HRESULT(ret)
}

func VarR8FromUI4(ulIn uint32, pdblOut *float64) HRESULT {
	addr := lazyAddr(&pVarR8FromUI4, libOleaut32, "VarR8FromUI4")
	ret, _, _ := syscall.SyscallN(addr, uintptr(ulIn), uintptr(unsafe.Pointer(pdblOut)))
	return HRESULT(ret)
}

func VarR8FromUI8(ui64In uint64, pdblOut *float64) HRESULT {
	addr := lazyAddr(&pVarR8FromUI8, libOleaut32, "VarR8FromUI8")
	ret, _, _ := syscall.SyscallN(addr, uintptr(ui64In), uintptr(unsafe.Pointer(pdblOut)))
	return HRESULT(ret)
}

func VarR8FromDec(pdecIn *DECIMAL, pdblOut *float64) HRESULT {
	addr := lazyAddr(&pVarR8FromDec, libOleaut32, "VarR8FromDec")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pdecIn)), uintptr(unsafe.Pointer(pdblOut)))
	return HRESULT(ret)
}

func VarDateFromUI1(bIn byte, pdateOut *float64) HRESULT {
	addr := lazyAddr(&pVarDateFromUI1, libOleaut32, "VarDateFromUI1")
	ret, _, _ := syscall.SyscallN(addr, uintptr(bIn), uintptr(unsafe.Pointer(pdateOut)))
	return HRESULT(ret)
}

func VarDateFromI2(sIn int16, pdateOut *float64) HRESULT {
	addr := lazyAddr(&pVarDateFromI2, libOleaut32, "VarDateFromI2")
	ret, _, _ := syscall.SyscallN(addr, uintptr(sIn), uintptr(unsafe.Pointer(pdateOut)))
	return HRESULT(ret)
}

func VarDateFromI4(lIn int32, pdateOut *float64) HRESULT {
	addr := lazyAddr(&pVarDateFromI4, libOleaut32, "VarDateFromI4")
	ret, _, _ := syscall.SyscallN(addr, uintptr(lIn), uintptr(unsafe.Pointer(pdateOut)))
	return HRESULT(ret)
}

func VarDateFromI8(i64In int64, pdateOut *float64) HRESULT {
	addr := lazyAddr(&pVarDateFromI8, libOleaut32, "VarDateFromI8")
	ret, _, _ := syscall.SyscallN(addr, uintptr(i64In), uintptr(unsafe.Pointer(pdateOut)))
	return HRESULT(ret)
}

func VarDateFromR4(fltIn float32, pdateOut *float64) HRESULT {
	addr := lazyAddr(&pVarDateFromR4, libOleaut32, "VarDateFromR4")
	ret, _, _ := syscall.SyscallN(addr, uintptr(fltIn), uintptr(unsafe.Pointer(pdateOut)))
	return HRESULT(ret)
}

func VarDateFromR8(dblIn float64, pdateOut *float64) HRESULT {
	addr := lazyAddr(&pVarDateFromR8, libOleaut32, "VarDateFromR8")
	ret, _, _ := syscall.SyscallN(addr, uintptr(dblIn), uintptr(unsafe.Pointer(pdateOut)))
	return HRESULT(ret)
}

func VarDateFromCy(cyIn CY, pdateOut *float64) HRESULT {
	addr := lazyAddr(&pVarDateFromCy, libOleaut32, "VarDateFromCy")
	ret, _, _ := syscall.SyscallN(addr, *(*uintptr)(unsafe.Pointer(&cyIn)), uintptr(unsafe.Pointer(pdateOut)))
	return HRESULT(ret)
}

func VarDateFromStr(strIn PWSTR, lcid uint32, dwFlags uint32, pdateOut *float64) HRESULT {
	addr := lazyAddr(&pVarDateFromStr, libOleaut32, "VarDateFromStr")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(strIn)), uintptr(lcid), uintptr(dwFlags), uintptr(unsafe.Pointer(pdateOut)))
	return HRESULT(ret)
}

func VarDateFromDisp(pdispIn *IDispatch, lcid uint32, pdateOut *float64) HRESULT {
	addr := lazyAddr(&pVarDateFromDisp, libOleaut32, "VarDateFromDisp")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pdispIn)), uintptr(lcid), uintptr(unsafe.Pointer(pdateOut)))
	return HRESULT(ret)
}

func VarDateFromBool(boolIn int16, pdateOut *float64) HRESULT {
	addr := lazyAddr(&pVarDateFromBool, libOleaut32, "VarDateFromBool")
	ret, _, _ := syscall.SyscallN(addr, uintptr(boolIn), uintptr(unsafe.Pointer(pdateOut)))
	return HRESULT(ret)
}

func VarDateFromI1(cIn CHAR, pdateOut *float64) HRESULT {
	addr := lazyAddr(&pVarDateFromI1, libOleaut32, "VarDateFromI1")
	ret, _, _ := syscall.SyscallN(addr, uintptr(cIn), uintptr(unsafe.Pointer(pdateOut)))
	return HRESULT(ret)
}

func VarDateFromUI2(uiIn uint16, pdateOut *float64) HRESULT {
	addr := lazyAddr(&pVarDateFromUI2, libOleaut32, "VarDateFromUI2")
	ret, _, _ := syscall.SyscallN(addr, uintptr(uiIn), uintptr(unsafe.Pointer(pdateOut)))
	return HRESULT(ret)
}

func VarDateFromUI4(ulIn uint32, pdateOut *float64) HRESULT {
	addr := lazyAddr(&pVarDateFromUI4, libOleaut32, "VarDateFromUI4")
	ret, _, _ := syscall.SyscallN(addr, uintptr(ulIn), uintptr(unsafe.Pointer(pdateOut)))
	return HRESULT(ret)
}

func VarDateFromUI8(ui64In uint64, pdateOut *float64) HRESULT {
	addr := lazyAddr(&pVarDateFromUI8, libOleaut32, "VarDateFromUI8")
	ret, _, _ := syscall.SyscallN(addr, uintptr(ui64In), uintptr(unsafe.Pointer(pdateOut)))
	return HRESULT(ret)
}

func VarDateFromDec(pdecIn *DECIMAL, pdateOut *float64) HRESULT {
	addr := lazyAddr(&pVarDateFromDec, libOleaut32, "VarDateFromDec")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pdecIn)), uintptr(unsafe.Pointer(pdateOut)))
	return HRESULT(ret)
}

func VarCyFromUI1(bIn byte, pcyOut *CY) HRESULT {
	addr := lazyAddr(&pVarCyFromUI1, libOleaut32, "VarCyFromUI1")
	ret, _, _ := syscall.SyscallN(addr, uintptr(bIn), uintptr(unsafe.Pointer(pcyOut)))
	return HRESULT(ret)
}

func VarCyFromI2(sIn int16, pcyOut *CY) HRESULT {
	addr := lazyAddr(&pVarCyFromI2, libOleaut32, "VarCyFromI2")
	ret, _, _ := syscall.SyscallN(addr, uintptr(sIn), uintptr(unsafe.Pointer(pcyOut)))
	return HRESULT(ret)
}

func VarCyFromI4(lIn int32, pcyOut *CY) HRESULT {
	addr := lazyAddr(&pVarCyFromI4, libOleaut32, "VarCyFromI4")
	ret, _, _ := syscall.SyscallN(addr, uintptr(lIn), uintptr(unsafe.Pointer(pcyOut)))
	return HRESULT(ret)
}

func VarCyFromI8(i64In int64, pcyOut *CY) HRESULT {
	addr := lazyAddr(&pVarCyFromI8, libOleaut32, "VarCyFromI8")
	ret, _, _ := syscall.SyscallN(addr, uintptr(i64In), uintptr(unsafe.Pointer(pcyOut)))
	return HRESULT(ret)
}

func VarCyFromR4(fltIn float32, pcyOut *CY) HRESULT {
	addr := lazyAddr(&pVarCyFromR4, libOleaut32, "VarCyFromR4")
	ret, _, _ := syscall.SyscallN(addr, uintptr(fltIn), uintptr(unsafe.Pointer(pcyOut)))
	return HRESULT(ret)
}

func VarCyFromR8(dblIn float64, pcyOut *CY) HRESULT {
	addr := lazyAddr(&pVarCyFromR8, libOleaut32, "VarCyFromR8")
	ret, _, _ := syscall.SyscallN(addr, uintptr(dblIn), uintptr(unsafe.Pointer(pcyOut)))
	return HRESULT(ret)
}

func VarCyFromDate(dateIn float64, pcyOut *CY) HRESULT {
	addr := lazyAddr(&pVarCyFromDate, libOleaut32, "VarCyFromDate")
	ret, _, _ := syscall.SyscallN(addr, uintptr(dateIn), uintptr(unsafe.Pointer(pcyOut)))
	return HRESULT(ret)
}

func VarCyFromStr(strIn PWSTR, lcid uint32, dwFlags uint32, pcyOut *CY) HRESULT {
	addr := lazyAddr(&pVarCyFromStr, libOleaut32, "VarCyFromStr")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(strIn)), uintptr(lcid), uintptr(dwFlags), uintptr(unsafe.Pointer(pcyOut)))
	return HRESULT(ret)
}

func VarCyFromDisp(pdispIn *IDispatch, lcid uint32, pcyOut *CY) HRESULT {
	addr := lazyAddr(&pVarCyFromDisp, libOleaut32, "VarCyFromDisp")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pdispIn)), uintptr(lcid), uintptr(unsafe.Pointer(pcyOut)))
	return HRESULT(ret)
}

func VarCyFromBool(boolIn int16, pcyOut *CY) HRESULT {
	addr := lazyAddr(&pVarCyFromBool, libOleaut32, "VarCyFromBool")
	ret, _, _ := syscall.SyscallN(addr, uintptr(boolIn), uintptr(unsafe.Pointer(pcyOut)))
	return HRESULT(ret)
}

func VarCyFromI1(cIn CHAR, pcyOut *CY) HRESULT {
	addr := lazyAddr(&pVarCyFromI1, libOleaut32, "VarCyFromI1")
	ret, _, _ := syscall.SyscallN(addr, uintptr(cIn), uintptr(unsafe.Pointer(pcyOut)))
	return HRESULT(ret)
}

func VarCyFromUI2(uiIn uint16, pcyOut *CY) HRESULT {
	addr := lazyAddr(&pVarCyFromUI2, libOleaut32, "VarCyFromUI2")
	ret, _, _ := syscall.SyscallN(addr, uintptr(uiIn), uintptr(unsafe.Pointer(pcyOut)))
	return HRESULT(ret)
}

func VarCyFromUI4(ulIn uint32, pcyOut *CY) HRESULT {
	addr := lazyAddr(&pVarCyFromUI4, libOleaut32, "VarCyFromUI4")
	ret, _, _ := syscall.SyscallN(addr, uintptr(ulIn), uintptr(unsafe.Pointer(pcyOut)))
	return HRESULT(ret)
}

func VarCyFromUI8(ui64In uint64, pcyOut *CY) HRESULT {
	addr := lazyAddr(&pVarCyFromUI8, libOleaut32, "VarCyFromUI8")
	ret, _, _ := syscall.SyscallN(addr, uintptr(ui64In), uintptr(unsafe.Pointer(pcyOut)))
	return HRESULT(ret)
}

func VarCyFromDec(pdecIn *DECIMAL, pcyOut *CY) HRESULT {
	addr := lazyAddr(&pVarCyFromDec, libOleaut32, "VarCyFromDec")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pdecIn)), uintptr(unsafe.Pointer(pcyOut)))
	return HRESULT(ret)
}

func VarBstrFromUI1(bVal byte, lcid uint32, dwFlags uint32, pbstrOut *BSTR) HRESULT {
	addr := lazyAddr(&pVarBstrFromUI1, libOleaut32, "VarBstrFromUI1")
	ret, _, _ := syscall.SyscallN(addr, uintptr(bVal), uintptr(lcid), uintptr(dwFlags), uintptr(unsafe.Pointer(pbstrOut)))
	return HRESULT(ret)
}

func VarBstrFromI2(iVal int16, lcid uint32, dwFlags uint32, pbstrOut *BSTR) HRESULT {
	addr := lazyAddr(&pVarBstrFromI2, libOleaut32, "VarBstrFromI2")
	ret, _, _ := syscall.SyscallN(addr, uintptr(iVal), uintptr(lcid), uintptr(dwFlags), uintptr(unsafe.Pointer(pbstrOut)))
	return HRESULT(ret)
}

func VarBstrFromI4(lIn int32, lcid uint32, dwFlags uint32, pbstrOut *BSTR) HRESULT {
	addr := lazyAddr(&pVarBstrFromI4, libOleaut32, "VarBstrFromI4")
	ret, _, _ := syscall.SyscallN(addr, uintptr(lIn), uintptr(lcid), uintptr(dwFlags), uintptr(unsafe.Pointer(pbstrOut)))
	return HRESULT(ret)
}

func VarBstrFromI8(i64In int64, lcid uint32, dwFlags uint32, pbstrOut *BSTR) HRESULT {
	addr := lazyAddr(&pVarBstrFromI8, libOleaut32, "VarBstrFromI8")
	ret, _, _ := syscall.SyscallN(addr, uintptr(i64In), uintptr(lcid), uintptr(dwFlags), uintptr(unsafe.Pointer(pbstrOut)))
	return HRESULT(ret)
}

func VarBstrFromR4(fltIn float32, lcid uint32, dwFlags uint32, pbstrOut *BSTR) HRESULT {
	addr := lazyAddr(&pVarBstrFromR4, libOleaut32, "VarBstrFromR4")
	ret, _, _ := syscall.SyscallN(addr, uintptr(fltIn), uintptr(lcid), uintptr(dwFlags), uintptr(unsafe.Pointer(pbstrOut)))
	return HRESULT(ret)
}

func VarBstrFromR8(dblIn float64, lcid uint32, dwFlags uint32, pbstrOut *BSTR) HRESULT {
	addr := lazyAddr(&pVarBstrFromR8, libOleaut32, "VarBstrFromR8")
	ret, _, _ := syscall.SyscallN(addr, uintptr(dblIn), uintptr(lcid), uintptr(dwFlags), uintptr(unsafe.Pointer(pbstrOut)))
	return HRESULT(ret)
}

func VarBstrFromCy(cyIn CY, lcid uint32, dwFlags uint32, pbstrOut *BSTR) HRESULT {
	addr := lazyAddr(&pVarBstrFromCy, libOleaut32, "VarBstrFromCy")
	ret, _, _ := syscall.SyscallN(addr, *(*uintptr)(unsafe.Pointer(&cyIn)), uintptr(lcid), uintptr(dwFlags), uintptr(unsafe.Pointer(pbstrOut)))
	return HRESULT(ret)
}

func VarBstrFromDate(dateIn float64, lcid uint32, dwFlags uint32, pbstrOut *BSTR) HRESULT {
	addr := lazyAddr(&pVarBstrFromDate, libOleaut32, "VarBstrFromDate")
	ret, _, _ := syscall.SyscallN(addr, uintptr(dateIn), uintptr(lcid), uintptr(dwFlags), uintptr(unsafe.Pointer(pbstrOut)))
	return HRESULT(ret)
}

func VarBstrFromDisp(pdispIn *IDispatch, lcid uint32, dwFlags uint32, pbstrOut *BSTR) HRESULT {
	addr := lazyAddr(&pVarBstrFromDisp, libOleaut32, "VarBstrFromDisp")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pdispIn)), uintptr(lcid), uintptr(dwFlags), uintptr(unsafe.Pointer(pbstrOut)))
	return HRESULT(ret)
}

func VarBstrFromBool(boolIn int16, lcid uint32, dwFlags uint32, pbstrOut *BSTR) HRESULT {
	addr := lazyAddr(&pVarBstrFromBool, libOleaut32, "VarBstrFromBool")
	ret, _, _ := syscall.SyscallN(addr, uintptr(boolIn), uintptr(lcid), uintptr(dwFlags), uintptr(unsafe.Pointer(pbstrOut)))
	return HRESULT(ret)
}

func VarBstrFromI1(cIn CHAR, lcid uint32, dwFlags uint32, pbstrOut *BSTR) HRESULT {
	addr := lazyAddr(&pVarBstrFromI1, libOleaut32, "VarBstrFromI1")
	ret, _, _ := syscall.SyscallN(addr, uintptr(cIn), uintptr(lcid), uintptr(dwFlags), uintptr(unsafe.Pointer(pbstrOut)))
	return HRESULT(ret)
}

func VarBstrFromUI2(uiIn uint16, lcid uint32, dwFlags uint32, pbstrOut *BSTR) HRESULT {
	addr := lazyAddr(&pVarBstrFromUI2, libOleaut32, "VarBstrFromUI2")
	ret, _, _ := syscall.SyscallN(addr, uintptr(uiIn), uintptr(lcid), uintptr(dwFlags), uintptr(unsafe.Pointer(pbstrOut)))
	return HRESULT(ret)
}

func VarBstrFromUI4(ulIn uint32, lcid uint32, dwFlags uint32, pbstrOut *BSTR) HRESULT {
	addr := lazyAddr(&pVarBstrFromUI4, libOleaut32, "VarBstrFromUI4")
	ret, _, _ := syscall.SyscallN(addr, uintptr(ulIn), uintptr(lcid), uintptr(dwFlags), uintptr(unsafe.Pointer(pbstrOut)))
	return HRESULT(ret)
}

func VarBstrFromUI8(ui64In uint64, lcid uint32, dwFlags uint32, pbstrOut *BSTR) HRESULT {
	addr := lazyAddr(&pVarBstrFromUI8, libOleaut32, "VarBstrFromUI8")
	ret, _, _ := syscall.SyscallN(addr, uintptr(ui64In), uintptr(lcid), uintptr(dwFlags), uintptr(unsafe.Pointer(pbstrOut)))
	return HRESULT(ret)
}

func VarBstrFromDec(pdecIn *DECIMAL, lcid uint32, dwFlags uint32, pbstrOut *BSTR) HRESULT {
	addr := lazyAddr(&pVarBstrFromDec, libOleaut32, "VarBstrFromDec")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pdecIn)), uintptr(lcid), uintptr(dwFlags), uintptr(unsafe.Pointer(pbstrOut)))
	return HRESULT(ret)
}

func VarBoolFromUI1(bIn byte, pboolOut *int16) HRESULT {
	addr := lazyAddr(&pVarBoolFromUI1, libOleaut32, "VarBoolFromUI1")
	ret, _, _ := syscall.SyscallN(addr, uintptr(bIn), uintptr(unsafe.Pointer(pboolOut)))
	return HRESULT(ret)
}

func VarBoolFromI2(sIn int16, pboolOut *int16) HRESULT {
	addr := lazyAddr(&pVarBoolFromI2, libOleaut32, "VarBoolFromI2")
	ret, _, _ := syscall.SyscallN(addr, uintptr(sIn), uintptr(unsafe.Pointer(pboolOut)))
	return HRESULT(ret)
}

func VarBoolFromI4(lIn int32, pboolOut *int16) HRESULT {
	addr := lazyAddr(&pVarBoolFromI4, libOleaut32, "VarBoolFromI4")
	ret, _, _ := syscall.SyscallN(addr, uintptr(lIn), uintptr(unsafe.Pointer(pboolOut)))
	return HRESULT(ret)
}

func VarBoolFromI8(i64In int64, pboolOut *int16) HRESULT {
	addr := lazyAddr(&pVarBoolFromI8, libOleaut32, "VarBoolFromI8")
	ret, _, _ := syscall.SyscallN(addr, uintptr(i64In), uintptr(unsafe.Pointer(pboolOut)))
	return HRESULT(ret)
}

func VarBoolFromR4(fltIn float32, pboolOut *int16) HRESULT {
	addr := lazyAddr(&pVarBoolFromR4, libOleaut32, "VarBoolFromR4")
	ret, _, _ := syscall.SyscallN(addr, uintptr(fltIn), uintptr(unsafe.Pointer(pboolOut)))
	return HRESULT(ret)
}

func VarBoolFromR8(dblIn float64, pboolOut *int16) HRESULT {
	addr := lazyAddr(&pVarBoolFromR8, libOleaut32, "VarBoolFromR8")
	ret, _, _ := syscall.SyscallN(addr, uintptr(dblIn), uintptr(unsafe.Pointer(pboolOut)))
	return HRESULT(ret)
}

func VarBoolFromDate(dateIn float64, pboolOut *int16) HRESULT {
	addr := lazyAddr(&pVarBoolFromDate, libOleaut32, "VarBoolFromDate")
	ret, _, _ := syscall.SyscallN(addr, uintptr(dateIn), uintptr(unsafe.Pointer(pboolOut)))
	return HRESULT(ret)
}

func VarBoolFromCy(cyIn CY, pboolOut *int16) HRESULT {
	addr := lazyAddr(&pVarBoolFromCy, libOleaut32, "VarBoolFromCy")
	ret, _, _ := syscall.SyscallN(addr, *(*uintptr)(unsafe.Pointer(&cyIn)), uintptr(unsafe.Pointer(pboolOut)))
	return HRESULT(ret)
}

func VarBoolFromStr(strIn PWSTR, lcid uint32, dwFlags uint32, pboolOut *int16) HRESULT {
	addr := lazyAddr(&pVarBoolFromStr, libOleaut32, "VarBoolFromStr")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(strIn)), uintptr(lcid), uintptr(dwFlags), uintptr(unsafe.Pointer(pboolOut)))
	return HRESULT(ret)
}

func VarBoolFromDisp(pdispIn *IDispatch, lcid uint32, pboolOut *int16) HRESULT {
	addr := lazyAddr(&pVarBoolFromDisp, libOleaut32, "VarBoolFromDisp")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pdispIn)), uintptr(lcid), uintptr(unsafe.Pointer(pboolOut)))
	return HRESULT(ret)
}

func VarBoolFromI1(cIn CHAR, pboolOut *int16) HRESULT {
	addr := lazyAddr(&pVarBoolFromI1, libOleaut32, "VarBoolFromI1")
	ret, _, _ := syscall.SyscallN(addr, uintptr(cIn), uintptr(unsafe.Pointer(pboolOut)))
	return HRESULT(ret)
}

func VarBoolFromUI2(uiIn uint16, pboolOut *int16) HRESULT {
	addr := lazyAddr(&pVarBoolFromUI2, libOleaut32, "VarBoolFromUI2")
	ret, _, _ := syscall.SyscallN(addr, uintptr(uiIn), uintptr(unsafe.Pointer(pboolOut)))
	return HRESULT(ret)
}

func VarBoolFromUI4(ulIn uint32, pboolOut *int16) HRESULT {
	addr := lazyAddr(&pVarBoolFromUI4, libOleaut32, "VarBoolFromUI4")
	ret, _, _ := syscall.SyscallN(addr, uintptr(ulIn), uintptr(unsafe.Pointer(pboolOut)))
	return HRESULT(ret)
}

func VarBoolFromUI8(i64In uint64, pboolOut *int16) HRESULT {
	addr := lazyAddr(&pVarBoolFromUI8, libOleaut32, "VarBoolFromUI8")
	ret, _, _ := syscall.SyscallN(addr, uintptr(i64In), uintptr(unsafe.Pointer(pboolOut)))
	return HRESULT(ret)
}

func VarBoolFromDec(pdecIn *DECIMAL, pboolOut *int16) HRESULT {
	addr := lazyAddr(&pVarBoolFromDec, libOleaut32, "VarBoolFromDec")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pdecIn)), uintptr(unsafe.Pointer(pboolOut)))
	return HRESULT(ret)
}

func VarI1FromUI1(bIn byte, pcOut PSTR) HRESULT {
	addr := lazyAddr(&pVarI1FromUI1, libOleaut32, "VarI1FromUI1")
	ret, _, _ := syscall.SyscallN(addr, uintptr(bIn), uintptr(unsafe.Pointer(pcOut)))
	return HRESULT(ret)
}

func VarI1FromI2(uiIn int16, pcOut PSTR) HRESULT {
	addr := lazyAddr(&pVarI1FromI2, libOleaut32, "VarI1FromI2")
	ret, _, _ := syscall.SyscallN(addr, uintptr(uiIn), uintptr(unsafe.Pointer(pcOut)))
	return HRESULT(ret)
}

func VarI1FromI4(lIn int32, pcOut PSTR) HRESULT {
	addr := lazyAddr(&pVarI1FromI4, libOleaut32, "VarI1FromI4")
	ret, _, _ := syscall.SyscallN(addr, uintptr(lIn), uintptr(unsafe.Pointer(pcOut)))
	return HRESULT(ret)
}

func VarI1FromI8(i64In int64, pcOut PSTR) HRESULT {
	addr := lazyAddr(&pVarI1FromI8, libOleaut32, "VarI1FromI8")
	ret, _, _ := syscall.SyscallN(addr, uintptr(i64In), uintptr(unsafe.Pointer(pcOut)))
	return HRESULT(ret)
}

func VarI1FromR4(fltIn float32, pcOut PSTR) HRESULT {
	addr := lazyAddr(&pVarI1FromR4, libOleaut32, "VarI1FromR4")
	ret, _, _ := syscall.SyscallN(addr, uintptr(fltIn), uintptr(unsafe.Pointer(pcOut)))
	return HRESULT(ret)
}

func VarI1FromR8(dblIn float64, pcOut PSTR) HRESULT {
	addr := lazyAddr(&pVarI1FromR8, libOleaut32, "VarI1FromR8")
	ret, _, _ := syscall.SyscallN(addr, uintptr(dblIn), uintptr(unsafe.Pointer(pcOut)))
	return HRESULT(ret)
}

func VarI1FromDate(dateIn float64, pcOut PSTR) HRESULT {
	addr := lazyAddr(&pVarI1FromDate, libOleaut32, "VarI1FromDate")
	ret, _, _ := syscall.SyscallN(addr, uintptr(dateIn), uintptr(unsafe.Pointer(pcOut)))
	return HRESULT(ret)
}

func VarI1FromCy(cyIn CY, pcOut PSTR) HRESULT {
	addr := lazyAddr(&pVarI1FromCy, libOleaut32, "VarI1FromCy")
	ret, _, _ := syscall.SyscallN(addr, *(*uintptr)(unsafe.Pointer(&cyIn)), uintptr(unsafe.Pointer(pcOut)))
	return HRESULT(ret)
}

func VarI1FromStr(strIn PWSTR, lcid uint32, dwFlags uint32, pcOut PSTR) HRESULT {
	addr := lazyAddr(&pVarI1FromStr, libOleaut32, "VarI1FromStr")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(strIn)), uintptr(lcid), uintptr(dwFlags), uintptr(unsafe.Pointer(pcOut)))
	return HRESULT(ret)
}

func VarI1FromDisp(pdispIn *IDispatch, lcid uint32, pcOut PSTR) HRESULT {
	addr := lazyAddr(&pVarI1FromDisp, libOleaut32, "VarI1FromDisp")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pdispIn)), uintptr(lcid), uintptr(unsafe.Pointer(pcOut)))
	return HRESULT(ret)
}

func VarI1FromBool(boolIn int16, pcOut PSTR) HRESULT {
	addr := lazyAddr(&pVarI1FromBool, libOleaut32, "VarI1FromBool")
	ret, _, _ := syscall.SyscallN(addr, uintptr(boolIn), uintptr(unsafe.Pointer(pcOut)))
	return HRESULT(ret)
}

func VarI1FromUI2(uiIn uint16, pcOut PSTR) HRESULT {
	addr := lazyAddr(&pVarI1FromUI2, libOleaut32, "VarI1FromUI2")
	ret, _, _ := syscall.SyscallN(addr, uintptr(uiIn), uintptr(unsafe.Pointer(pcOut)))
	return HRESULT(ret)
}

func VarI1FromUI4(ulIn uint32, pcOut PSTR) HRESULT {
	addr := lazyAddr(&pVarI1FromUI4, libOleaut32, "VarI1FromUI4")
	ret, _, _ := syscall.SyscallN(addr, uintptr(ulIn), uintptr(unsafe.Pointer(pcOut)))
	return HRESULT(ret)
}

func VarI1FromUI8(i64In uint64, pcOut PSTR) HRESULT {
	addr := lazyAddr(&pVarI1FromUI8, libOleaut32, "VarI1FromUI8")
	ret, _, _ := syscall.SyscallN(addr, uintptr(i64In), uintptr(unsafe.Pointer(pcOut)))
	return HRESULT(ret)
}

func VarI1FromDec(pdecIn *DECIMAL, pcOut PSTR) HRESULT {
	addr := lazyAddr(&pVarI1FromDec, libOleaut32, "VarI1FromDec")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pdecIn)), uintptr(unsafe.Pointer(pcOut)))
	return HRESULT(ret)
}

func VarUI2FromUI1(bIn byte, puiOut *uint16) HRESULT {
	addr := lazyAddr(&pVarUI2FromUI1, libOleaut32, "VarUI2FromUI1")
	ret, _, _ := syscall.SyscallN(addr, uintptr(bIn), uintptr(unsafe.Pointer(puiOut)))
	return HRESULT(ret)
}

func VarUI2FromI2(uiIn int16, puiOut *uint16) HRESULT {
	addr := lazyAddr(&pVarUI2FromI2, libOleaut32, "VarUI2FromI2")
	ret, _, _ := syscall.SyscallN(addr, uintptr(uiIn), uintptr(unsafe.Pointer(puiOut)))
	return HRESULT(ret)
}

func VarUI2FromI4(lIn int32, puiOut *uint16) HRESULT {
	addr := lazyAddr(&pVarUI2FromI4, libOleaut32, "VarUI2FromI4")
	ret, _, _ := syscall.SyscallN(addr, uintptr(lIn), uintptr(unsafe.Pointer(puiOut)))
	return HRESULT(ret)
}

func VarUI2FromI8(i64In int64, puiOut *uint16) HRESULT {
	addr := lazyAddr(&pVarUI2FromI8, libOleaut32, "VarUI2FromI8")
	ret, _, _ := syscall.SyscallN(addr, uintptr(i64In), uintptr(unsafe.Pointer(puiOut)))
	return HRESULT(ret)
}

func VarUI2FromR4(fltIn float32, puiOut *uint16) HRESULT {
	addr := lazyAddr(&pVarUI2FromR4, libOleaut32, "VarUI2FromR4")
	ret, _, _ := syscall.SyscallN(addr, uintptr(fltIn), uintptr(unsafe.Pointer(puiOut)))
	return HRESULT(ret)
}

func VarUI2FromR8(dblIn float64, puiOut *uint16) HRESULT {
	addr := lazyAddr(&pVarUI2FromR8, libOleaut32, "VarUI2FromR8")
	ret, _, _ := syscall.SyscallN(addr, uintptr(dblIn), uintptr(unsafe.Pointer(puiOut)))
	return HRESULT(ret)
}

func VarUI2FromDate(dateIn float64, puiOut *uint16) HRESULT {
	addr := lazyAddr(&pVarUI2FromDate, libOleaut32, "VarUI2FromDate")
	ret, _, _ := syscall.SyscallN(addr, uintptr(dateIn), uintptr(unsafe.Pointer(puiOut)))
	return HRESULT(ret)
}

func VarUI2FromCy(cyIn CY, puiOut *uint16) HRESULT {
	addr := lazyAddr(&pVarUI2FromCy, libOleaut32, "VarUI2FromCy")
	ret, _, _ := syscall.SyscallN(addr, *(*uintptr)(unsafe.Pointer(&cyIn)), uintptr(unsafe.Pointer(puiOut)))
	return HRESULT(ret)
}

func VarUI2FromStr(strIn PWSTR, lcid uint32, dwFlags uint32, puiOut *uint16) HRESULT {
	addr := lazyAddr(&pVarUI2FromStr, libOleaut32, "VarUI2FromStr")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(strIn)), uintptr(lcid), uintptr(dwFlags), uintptr(unsafe.Pointer(puiOut)))
	return HRESULT(ret)
}

func VarUI2FromDisp(pdispIn *IDispatch, lcid uint32, puiOut *uint16) HRESULT {
	addr := lazyAddr(&pVarUI2FromDisp, libOleaut32, "VarUI2FromDisp")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pdispIn)), uintptr(lcid), uintptr(unsafe.Pointer(puiOut)))
	return HRESULT(ret)
}

func VarUI2FromBool(boolIn int16, puiOut *uint16) HRESULT {
	addr := lazyAddr(&pVarUI2FromBool, libOleaut32, "VarUI2FromBool")
	ret, _, _ := syscall.SyscallN(addr, uintptr(boolIn), uintptr(unsafe.Pointer(puiOut)))
	return HRESULT(ret)
}

func VarUI2FromI1(cIn CHAR, puiOut *uint16) HRESULT {
	addr := lazyAddr(&pVarUI2FromI1, libOleaut32, "VarUI2FromI1")
	ret, _, _ := syscall.SyscallN(addr, uintptr(cIn), uintptr(unsafe.Pointer(puiOut)))
	return HRESULT(ret)
}

func VarUI2FromUI4(ulIn uint32, puiOut *uint16) HRESULT {
	addr := lazyAddr(&pVarUI2FromUI4, libOleaut32, "VarUI2FromUI4")
	ret, _, _ := syscall.SyscallN(addr, uintptr(ulIn), uintptr(unsafe.Pointer(puiOut)))
	return HRESULT(ret)
}

func VarUI2FromUI8(i64In uint64, puiOut *uint16) HRESULT {
	addr := lazyAddr(&pVarUI2FromUI8, libOleaut32, "VarUI2FromUI8")
	ret, _, _ := syscall.SyscallN(addr, uintptr(i64In), uintptr(unsafe.Pointer(puiOut)))
	return HRESULT(ret)
}

func VarUI2FromDec(pdecIn *DECIMAL, puiOut *uint16) HRESULT {
	addr := lazyAddr(&pVarUI2FromDec, libOleaut32, "VarUI2FromDec")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pdecIn)), uintptr(unsafe.Pointer(puiOut)))
	return HRESULT(ret)
}

func VarUI4FromUI1(bIn byte, pulOut *uint32) HRESULT {
	addr := lazyAddr(&pVarUI4FromUI1, libOleaut32, "VarUI4FromUI1")
	ret, _, _ := syscall.SyscallN(addr, uintptr(bIn), uintptr(unsafe.Pointer(pulOut)))
	return HRESULT(ret)
}

func VarUI4FromI2(uiIn int16, pulOut *uint32) HRESULT {
	addr := lazyAddr(&pVarUI4FromI2, libOleaut32, "VarUI4FromI2")
	ret, _, _ := syscall.SyscallN(addr, uintptr(uiIn), uintptr(unsafe.Pointer(pulOut)))
	return HRESULT(ret)
}

func VarUI4FromI4(lIn int32, pulOut *uint32) HRESULT {
	addr := lazyAddr(&pVarUI4FromI4, libOleaut32, "VarUI4FromI4")
	ret, _, _ := syscall.SyscallN(addr, uintptr(lIn), uintptr(unsafe.Pointer(pulOut)))
	return HRESULT(ret)
}

func VarUI4FromI8(i64In int64, plOut *uint32) HRESULT {
	addr := lazyAddr(&pVarUI4FromI8, libOleaut32, "VarUI4FromI8")
	ret, _, _ := syscall.SyscallN(addr, uintptr(i64In), uintptr(unsafe.Pointer(plOut)))
	return HRESULT(ret)
}

func VarUI4FromR4(fltIn float32, pulOut *uint32) HRESULT {
	addr := lazyAddr(&pVarUI4FromR4, libOleaut32, "VarUI4FromR4")
	ret, _, _ := syscall.SyscallN(addr, uintptr(fltIn), uintptr(unsafe.Pointer(pulOut)))
	return HRESULT(ret)
}

func VarUI4FromR8(dblIn float64, pulOut *uint32) HRESULT {
	addr := lazyAddr(&pVarUI4FromR8, libOleaut32, "VarUI4FromR8")
	ret, _, _ := syscall.SyscallN(addr, uintptr(dblIn), uintptr(unsafe.Pointer(pulOut)))
	return HRESULT(ret)
}

func VarUI4FromDate(dateIn float64, pulOut *uint32) HRESULT {
	addr := lazyAddr(&pVarUI4FromDate, libOleaut32, "VarUI4FromDate")
	ret, _, _ := syscall.SyscallN(addr, uintptr(dateIn), uintptr(unsafe.Pointer(pulOut)))
	return HRESULT(ret)
}

func VarUI4FromCy(cyIn CY, pulOut *uint32) HRESULT {
	addr := lazyAddr(&pVarUI4FromCy, libOleaut32, "VarUI4FromCy")
	ret, _, _ := syscall.SyscallN(addr, *(*uintptr)(unsafe.Pointer(&cyIn)), uintptr(unsafe.Pointer(pulOut)))
	return HRESULT(ret)
}

func VarUI4FromStr(strIn PWSTR, lcid uint32, dwFlags uint32, pulOut *uint32) HRESULT {
	addr := lazyAddr(&pVarUI4FromStr, libOleaut32, "VarUI4FromStr")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(strIn)), uintptr(lcid), uintptr(dwFlags), uintptr(unsafe.Pointer(pulOut)))
	return HRESULT(ret)
}

func VarUI4FromDisp(pdispIn *IDispatch, lcid uint32, pulOut *uint32) HRESULT {
	addr := lazyAddr(&pVarUI4FromDisp, libOleaut32, "VarUI4FromDisp")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pdispIn)), uintptr(lcid), uintptr(unsafe.Pointer(pulOut)))
	return HRESULT(ret)
}

func VarUI4FromBool(boolIn int16, pulOut *uint32) HRESULT {
	addr := lazyAddr(&pVarUI4FromBool, libOleaut32, "VarUI4FromBool")
	ret, _, _ := syscall.SyscallN(addr, uintptr(boolIn), uintptr(unsafe.Pointer(pulOut)))
	return HRESULT(ret)
}

func VarUI4FromI1(cIn CHAR, pulOut *uint32) HRESULT {
	addr := lazyAddr(&pVarUI4FromI1, libOleaut32, "VarUI4FromI1")
	ret, _, _ := syscall.SyscallN(addr, uintptr(cIn), uintptr(unsafe.Pointer(pulOut)))
	return HRESULT(ret)
}

func VarUI4FromUI2(uiIn uint16, pulOut *uint32) HRESULT {
	addr := lazyAddr(&pVarUI4FromUI2, libOleaut32, "VarUI4FromUI2")
	ret, _, _ := syscall.SyscallN(addr, uintptr(uiIn), uintptr(unsafe.Pointer(pulOut)))
	return HRESULT(ret)
}

func VarUI4FromUI8(ui64In uint64, plOut *uint32) HRESULT {
	addr := lazyAddr(&pVarUI4FromUI8, libOleaut32, "VarUI4FromUI8")
	ret, _, _ := syscall.SyscallN(addr, uintptr(ui64In), uintptr(unsafe.Pointer(plOut)))
	return HRESULT(ret)
}

func VarUI4FromDec(pdecIn *DECIMAL, pulOut *uint32) HRESULT {
	addr := lazyAddr(&pVarUI4FromDec, libOleaut32, "VarUI4FromDec")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pdecIn)), uintptr(unsafe.Pointer(pulOut)))
	return HRESULT(ret)
}

func VarUI8FromUI1(bIn byte, pi64Out *uint64) HRESULT {
	addr := lazyAddr(&pVarUI8FromUI1, libOleaut32, "VarUI8FromUI1")
	ret, _, _ := syscall.SyscallN(addr, uintptr(bIn), uintptr(unsafe.Pointer(pi64Out)))
	return HRESULT(ret)
}

func VarUI8FromI2(sIn int16, pi64Out *uint64) HRESULT {
	addr := lazyAddr(&pVarUI8FromI2, libOleaut32, "VarUI8FromI2")
	ret, _, _ := syscall.SyscallN(addr, uintptr(sIn), uintptr(unsafe.Pointer(pi64Out)))
	return HRESULT(ret)
}

func VarUI8FromI8(ui64In int64, pi64Out *uint64) HRESULT {
	addr := lazyAddr(&pVarUI8FromI8, libOleaut32, "VarUI8FromI8")
	ret, _, _ := syscall.SyscallN(addr, uintptr(ui64In), uintptr(unsafe.Pointer(pi64Out)))
	return HRESULT(ret)
}

func VarUI8FromR4(fltIn float32, pi64Out *uint64) HRESULT {
	addr := lazyAddr(&pVarUI8FromR4, libOleaut32, "VarUI8FromR4")
	ret, _, _ := syscall.SyscallN(addr, uintptr(fltIn), uintptr(unsafe.Pointer(pi64Out)))
	return HRESULT(ret)
}

func VarUI8FromR8(dblIn float64, pi64Out *uint64) HRESULT {
	addr := lazyAddr(&pVarUI8FromR8, libOleaut32, "VarUI8FromR8")
	ret, _, _ := syscall.SyscallN(addr, uintptr(dblIn), uintptr(unsafe.Pointer(pi64Out)))
	return HRESULT(ret)
}

func VarUI8FromCy(cyIn CY, pi64Out *uint64) HRESULT {
	addr := lazyAddr(&pVarUI8FromCy, libOleaut32, "VarUI8FromCy")
	ret, _, _ := syscall.SyscallN(addr, *(*uintptr)(unsafe.Pointer(&cyIn)), uintptr(unsafe.Pointer(pi64Out)))
	return HRESULT(ret)
}

func VarUI8FromDate(dateIn float64, pi64Out *uint64) HRESULT {
	addr := lazyAddr(&pVarUI8FromDate, libOleaut32, "VarUI8FromDate")
	ret, _, _ := syscall.SyscallN(addr, uintptr(dateIn), uintptr(unsafe.Pointer(pi64Out)))
	return HRESULT(ret)
}

func VarUI8FromStr(strIn PWSTR, lcid uint32, dwFlags uint32, pi64Out *uint64) HRESULT {
	addr := lazyAddr(&pVarUI8FromStr, libOleaut32, "VarUI8FromStr")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(strIn)), uintptr(lcid), uintptr(dwFlags), uintptr(unsafe.Pointer(pi64Out)))
	return HRESULT(ret)
}

func VarUI8FromDisp(pdispIn *IDispatch, lcid uint32, pi64Out *uint64) HRESULT {
	addr := lazyAddr(&pVarUI8FromDisp, libOleaut32, "VarUI8FromDisp")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pdispIn)), uintptr(lcid), uintptr(unsafe.Pointer(pi64Out)))
	return HRESULT(ret)
}

func VarUI8FromBool(boolIn int16, pi64Out *uint64) HRESULT {
	addr := lazyAddr(&pVarUI8FromBool, libOleaut32, "VarUI8FromBool")
	ret, _, _ := syscall.SyscallN(addr, uintptr(boolIn), uintptr(unsafe.Pointer(pi64Out)))
	return HRESULT(ret)
}

func VarUI8FromI1(cIn CHAR, pi64Out *uint64) HRESULT {
	addr := lazyAddr(&pVarUI8FromI1, libOleaut32, "VarUI8FromI1")
	ret, _, _ := syscall.SyscallN(addr, uintptr(cIn), uintptr(unsafe.Pointer(pi64Out)))
	return HRESULT(ret)
}

func VarUI8FromUI2(uiIn uint16, pi64Out *uint64) HRESULT {
	addr := lazyAddr(&pVarUI8FromUI2, libOleaut32, "VarUI8FromUI2")
	ret, _, _ := syscall.SyscallN(addr, uintptr(uiIn), uintptr(unsafe.Pointer(pi64Out)))
	return HRESULT(ret)
}

func VarUI8FromUI4(ulIn uint32, pi64Out *uint64) HRESULT {
	addr := lazyAddr(&pVarUI8FromUI4, libOleaut32, "VarUI8FromUI4")
	ret, _, _ := syscall.SyscallN(addr, uintptr(ulIn), uintptr(unsafe.Pointer(pi64Out)))
	return HRESULT(ret)
}

func VarUI8FromDec(pdecIn *DECIMAL, pi64Out *uint64) HRESULT {
	addr := lazyAddr(&pVarUI8FromDec, libOleaut32, "VarUI8FromDec")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pdecIn)), uintptr(unsafe.Pointer(pi64Out)))
	return HRESULT(ret)
}

func VarDecFromUI1(bIn byte, pdecOut *DECIMAL) HRESULT {
	addr := lazyAddr(&pVarDecFromUI1, libOleaut32, "VarDecFromUI1")
	ret, _, _ := syscall.SyscallN(addr, uintptr(bIn), uintptr(unsafe.Pointer(pdecOut)))
	return HRESULT(ret)
}

func VarDecFromI2(uiIn int16, pdecOut *DECIMAL) HRESULT {
	addr := lazyAddr(&pVarDecFromI2, libOleaut32, "VarDecFromI2")
	ret, _, _ := syscall.SyscallN(addr, uintptr(uiIn), uintptr(unsafe.Pointer(pdecOut)))
	return HRESULT(ret)
}

func VarDecFromI4(lIn int32, pdecOut *DECIMAL) HRESULT {
	addr := lazyAddr(&pVarDecFromI4, libOleaut32, "VarDecFromI4")
	ret, _, _ := syscall.SyscallN(addr, uintptr(lIn), uintptr(unsafe.Pointer(pdecOut)))
	return HRESULT(ret)
}

func VarDecFromI8(i64In int64, pdecOut *DECIMAL) HRESULT {
	addr := lazyAddr(&pVarDecFromI8, libOleaut32, "VarDecFromI8")
	ret, _, _ := syscall.SyscallN(addr, uintptr(i64In), uintptr(unsafe.Pointer(pdecOut)))
	return HRESULT(ret)
}

func VarDecFromR4(fltIn float32, pdecOut *DECIMAL) HRESULT {
	addr := lazyAddr(&pVarDecFromR4, libOleaut32, "VarDecFromR4")
	ret, _, _ := syscall.SyscallN(addr, uintptr(fltIn), uintptr(unsafe.Pointer(pdecOut)))
	return HRESULT(ret)
}

func VarDecFromR8(dblIn float64, pdecOut *DECIMAL) HRESULT {
	addr := lazyAddr(&pVarDecFromR8, libOleaut32, "VarDecFromR8")
	ret, _, _ := syscall.SyscallN(addr, uintptr(dblIn), uintptr(unsafe.Pointer(pdecOut)))
	return HRESULT(ret)
}

func VarDecFromDate(dateIn float64, pdecOut *DECIMAL) HRESULT {
	addr := lazyAddr(&pVarDecFromDate, libOleaut32, "VarDecFromDate")
	ret, _, _ := syscall.SyscallN(addr, uintptr(dateIn), uintptr(unsafe.Pointer(pdecOut)))
	return HRESULT(ret)
}

func VarDecFromCy(cyIn CY, pdecOut *DECIMAL) HRESULT {
	addr := lazyAddr(&pVarDecFromCy, libOleaut32, "VarDecFromCy")
	ret, _, _ := syscall.SyscallN(addr, *(*uintptr)(unsafe.Pointer(&cyIn)), uintptr(unsafe.Pointer(pdecOut)))
	return HRESULT(ret)
}

func VarDecFromStr(strIn PWSTR, lcid uint32, dwFlags uint32, pdecOut *DECIMAL) HRESULT {
	addr := lazyAddr(&pVarDecFromStr, libOleaut32, "VarDecFromStr")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(strIn)), uintptr(lcid), uintptr(dwFlags), uintptr(unsafe.Pointer(pdecOut)))
	return HRESULT(ret)
}

func VarDecFromDisp(pdispIn *IDispatch, lcid uint32, pdecOut *DECIMAL) HRESULT {
	addr := lazyAddr(&pVarDecFromDisp, libOleaut32, "VarDecFromDisp")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pdispIn)), uintptr(lcid), uintptr(unsafe.Pointer(pdecOut)))
	return HRESULT(ret)
}

func VarDecFromBool(boolIn int16, pdecOut *DECIMAL) HRESULT {
	addr := lazyAddr(&pVarDecFromBool, libOleaut32, "VarDecFromBool")
	ret, _, _ := syscall.SyscallN(addr, uintptr(boolIn), uintptr(unsafe.Pointer(pdecOut)))
	return HRESULT(ret)
}

func VarDecFromI1(cIn CHAR, pdecOut *DECIMAL) HRESULT {
	addr := lazyAddr(&pVarDecFromI1, libOleaut32, "VarDecFromI1")
	ret, _, _ := syscall.SyscallN(addr, uintptr(cIn), uintptr(unsafe.Pointer(pdecOut)))
	return HRESULT(ret)
}

func VarDecFromUI2(uiIn uint16, pdecOut *DECIMAL) HRESULT {
	addr := lazyAddr(&pVarDecFromUI2, libOleaut32, "VarDecFromUI2")
	ret, _, _ := syscall.SyscallN(addr, uintptr(uiIn), uintptr(unsafe.Pointer(pdecOut)))
	return HRESULT(ret)
}

func VarDecFromUI4(ulIn uint32, pdecOut *DECIMAL) HRESULT {
	addr := lazyAddr(&pVarDecFromUI4, libOleaut32, "VarDecFromUI4")
	ret, _, _ := syscall.SyscallN(addr, uintptr(ulIn), uintptr(unsafe.Pointer(pdecOut)))
	return HRESULT(ret)
}

func VarDecFromUI8(ui64In uint64, pdecOut *DECIMAL) HRESULT {
	addr := lazyAddr(&pVarDecFromUI8, libOleaut32, "VarDecFromUI8")
	ret, _, _ := syscall.SyscallN(addr, uintptr(ui64In), uintptr(unsafe.Pointer(pdecOut)))
	return HRESULT(ret)
}

func VarParseNumFromStr(strIn PWSTR, lcid uint32, dwFlags uint32, pnumprs *NUMPARSE, rgbDig *byte) HRESULT {
	addr := lazyAddr(&pVarParseNumFromStr, libOleaut32, "VarParseNumFromStr")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(strIn)), uintptr(lcid), uintptr(dwFlags), uintptr(unsafe.Pointer(pnumprs)), uintptr(unsafe.Pointer(rgbDig)))
	return HRESULT(ret)
}

func VarNumFromParseNum(pnumprs *NUMPARSE, rgbDig *byte, dwVtBits uint32, pvar *VARIANT) HRESULT {
	addr := lazyAddr(&pVarNumFromParseNum, libOleaut32, "VarNumFromParseNum")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pnumprs)), uintptr(unsafe.Pointer(rgbDig)), uintptr(dwVtBits), uintptr(unsafe.Pointer(pvar)))
	return HRESULT(ret)
}

func VarAdd(pvarLeft *VARIANT, pvarRight *VARIANT, pvarResult *VARIANT) HRESULT {
	addr := lazyAddr(&pVarAdd, libOleaut32, "VarAdd")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pvarLeft)), uintptr(unsafe.Pointer(pvarRight)), uintptr(unsafe.Pointer(pvarResult)))
	return HRESULT(ret)
}

func VarAnd(pvarLeft *VARIANT, pvarRight *VARIANT, pvarResult *VARIANT) HRESULT {
	addr := lazyAddr(&pVarAnd, libOleaut32, "VarAnd")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pvarLeft)), uintptr(unsafe.Pointer(pvarRight)), uintptr(unsafe.Pointer(pvarResult)))
	return HRESULT(ret)
}

func VarCat(pvarLeft *VARIANT, pvarRight *VARIANT, pvarResult *VARIANT) HRESULT {
	addr := lazyAddr(&pVarCat, libOleaut32, "VarCat")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pvarLeft)), uintptr(unsafe.Pointer(pvarRight)), uintptr(unsafe.Pointer(pvarResult)))
	return HRESULT(ret)
}

func VarDiv(pvarLeft *VARIANT, pvarRight *VARIANT, pvarResult *VARIANT) HRESULT {
	addr := lazyAddr(&pVarDiv, libOleaut32, "VarDiv")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pvarLeft)), uintptr(unsafe.Pointer(pvarRight)), uintptr(unsafe.Pointer(pvarResult)))
	return HRESULT(ret)
}

func VarEqv(pvarLeft *VARIANT, pvarRight *VARIANT, pvarResult *VARIANT) HRESULT {
	addr := lazyAddr(&pVarEqv, libOleaut32, "VarEqv")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pvarLeft)), uintptr(unsafe.Pointer(pvarRight)), uintptr(unsafe.Pointer(pvarResult)))
	return HRESULT(ret)
}

func VarIdiv(pvarLeft *VARIANT, pvarRight *VARIANT, pvarResult *VARIANT) HRESULT {
	addr := lazyAddr(&pVarIdiv, libOleaut32, "VarIdiv")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pvarLeft)), uintptr(unsafe.Pointer(pvarRight)), uintptr(unsafe.Pointer(pvarResult)))
	return HRESULT(ret)
}

func VarImp(pvarLeft *VARIANT, pvarRight *VARIANT, pvarResult *VARIANT) HRESULT {
	addr := lazyAddr(&pVarImp, libOleaut32, "VarImp")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pvarLeft)), uintptr(unsafe.Pointer(pvarRight)), uintptr(unsafe.Pointer(pvarResult)))
	return HRESULT(ret)
}

func VarMod(pvarLeft *VARIANT, pvarRight *VARIANT, pvarResult *VARIANT) HRESULT {
	addr := lazyAddr(&pVarMod, libOleaut32, "VarMod")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pvarLeft)), uintptr(unsafe.Pointer(pvarRight)), uintptr(unsafe.Pointer(pvarResult)))
	return HRESULT(ret)
}

func VarMul(pvarLeft *VARIANT, pvarRight *VARIANT, pvarResult *VARIANT) HRESULT {
	addr := lazyAddr(&pVarMul, libOleaut32, "VarMul")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pvarLeft)), uintptr(unsafe.Pointer(pvarRight)), uintptr(unsafe.Pointer(pvarResult)))
	return HRESULT(ret)
}

func VarOr(pvarLeft *VARIANT, pvarRight *VARIANT, pvarResult *VARIANT) HRESULT {
	addr := lazyAddr(&pVarOr, libOleaut32, "VarOr")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pvarLeft)), uintptr(unsafe.Pointer(pvarRight)), uintptr(unsafe.Pointer(pvarResult)))
	return HRESULT(ret)
}

func VarPow(pvarLeft *VARIANT, pvarRight *VARIANT, pvarResult *VARIANT) HRESULT {
	addr := lazyAddr(&pVarPow, libOleaut32, "VarPow")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pvarLeft)), uintptr(unsafe.Pointer(pvarRight)), uintptr(unsafe.Pointer(pvarResult)))
	return HRESULT(ret)
}

func VarSub(pvarLeft *VARIANT, pvarRight *VARIANT, pvarResult *VARIANT) HRESULT {
	addr := lazyAddr(&pVarSub, libOleaut32, "VarSub")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pvarLeft)), uintptr(unsafe.Pointer(pvarRight)), uintptr(unsafe.Pointer(pvarResult)))
	return HRESULT(ret)
}

func VarXor(pvarLeft *VARIANT, pvarRight *VARIANT, pvarResult *VARIANT) HRESULT {
	addr := lazyAddr(&pVarXor, libOleaut32, "VarXor")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pvarLeft)), uintptr(unsafe.Pointer(pvarRight)), uintptr(unsafe.Pointer(pvarResult)))
	return HRESULT(ret)
}

func VarAbs(pvarIn *VARIANT, pvarResult *VARIANT) HRESULT {
	addr := lazyAddr(&pVarAbs, libOleaut32, "VarAbs")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pvarIn)), uintptr(unsafe.Pointer(pvarResult)))
	return HRESULT(ret)
}

func VarFix(pvarIn *VARIANT, pvarResult *VARIANT) HRESULT {
	addr := lazyAddr(&pVarFix, libOleaut32, "VarFix")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pvarIn)), uintptr(unsafe.Pointer(pvarResult)))
	return HRESULT(ret)
}

func VarInt(pvarIn *VARIANT, pvarResult *VARIANT) HRESULT {
	addr := lazyAddr(&pVarInt, libOleaut32, "VarInt")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pvarIn)), uintptr(unsafe.Pointer(pvarResult)))
	return HRESULT(ret)
}

func VarNeg(pvarIn *VARIANT, pvarResult *VARIANT) HRESULT {
	addr := lazyAddr(&pVarNeg, libOleaut32, "VarNeg")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pvarIn)), uintptr(unsafe.Pointer(pvarResult)))
	return HRESULT(ret)
}

func VarNot(pvarIn *VARIANT, pvarResult *VARIANT) HRESULT {
	addr := lazyAddr(&pVarNot, libOleaut32, "VarNot")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pvarIn)), uintptr(unsafe.Pointer(pvarResult)))
	return HRESULT(ret)
}

func VarRound(pvarIn *VARIANT, cDecimals int32, pvarResult *VARIANT) HRESULT {
	addr := lazyAddr(&pVarRound, libOleaut32, "VarRound")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pvarIn)), uintptr(cDecimals), uintptr(unsafe.Pointer(pvarResult)))
	return HRESULT(ret)
}

func VarCmp(pvarLeft *VARIANT, pvarRight *VARIANT, lcid uint32, dwFlags uint32) HRESULT {
	addr := lazyAddr(&pVarCmp, libOleaut32, "VarCmp")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pvarLeft)), uintptr(unsafe.Pointer(pvarRight)), uintptr(lcid), uintptr(dwFlags))
	return HRESULT(ret)
}

func VarDecAdd(pdecLeft *DECIMAL, pdecRight *DECIMAL, pdecResult *DECIMAL) HRESULT {
	addr := lazyAddr(&pVarDecAdd, libOleaut32, "VarDecAdd")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pdecLeft)), uintptr(unsafe.Pointer(pdecRight)), uintptr(unsafe.Pointer(pdecResult)))
	return HRESULT(ret)
}

func VarDecDiv(pdecLeft *DECIMAL, pdecRight *DECIMAL, pdecResult *DECIMAL) HRESULT {
	addr := lazyAddr(&pVarDecDiv, libOleaut32, "VarDecDiv")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pdecLeft)), uintptr(unsafe.Pointer(pdecRight)), uintptr(unsafe.Pointer(pdecResult)))
	return HRESULT(ret)
}

func VarDecMul(pdecLeft *DECIMAL, pdecRight *DECIMAL, pdecResult *DECIMAL) HRESULT {
	addr := lazyAddr(&pVarDecMul, libOleaut32, "VarDecMul")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pdecLeft)), uintptr(unsafe.Pointer(pdecRight)), uintptr(unsafe.Pointer(pdecResult)))
	return HRESULT(ret)
}

func VarDecSub(pdecLeft *DECIMAL, pdecRight *DECIMAL, pdecResult *DECIMAL) HRESULT {
	addr := lazyAddr(&pVarDecSub, libOleaut32, "VarDecSub")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pdecLeft)), uintptr(unsafe.Pointer(pdecRight)), uintptr(unsafe.Pointer(pdecResult)))
	return HRESULT(ret)
}

func VarDecAbs(pdecIn *DECIMAL, pdecResult *DECIMAL) HRESULT {
	addr := lazyAddr(&pVarDecAbs, libOleaut32, "VarDecAbs")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pdecIn)), uintptr(unsafe.Pointer(pdecResult)))
	return HRESULT(ret)
}

func VarDecFix(pdecIn *DECIMAL, pdecResult *DECIMAL) HRESULT {
	addr := lazyAddr(&pVarDecFix, libOleaut32, "VarDecFix")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pdecIn)), uintptr(unsafe.Pointer(pdecResult)))
	return HRESULT(ret)
}

func VarDecInt(pdecIn *DECIMAL, pdecResult *DECIMAL) HRESULT {
	addr := lazyAddr(&pVarDecInt, libOleaut32, "VarDecInt")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pdecIn)), uintptr(unsafe.Pointer(pdecResult)))
	return HRESULT(ret)
}

func VarDecNeg(pdecIn *DECIMAL, pdecResult *DECIMAL) HRESULT {
	addr := lazyAddr(&pVarDecNeg, libOleaut32, "VarDecNeg")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pdecIn)), uintptr(unsafe.Pointer(pdecResult)))
	return HRESULT(ret)
}

func VarDecRound(pdecIn *DECIMAL, cDecimals int32, pdecResult *DECIMAL) HRESULT {
	addr := lazyAddr(&pVarDecRound, libOleaut32, "VarDecRound")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pdecIn)), uintptr(cDecimals), uintptr(unsafe.Pointer(pdecResult)))
	return HRESULT(ret)
}

func VarDecCmp(pdecLeft *DECIMAL, pdecRight *DECIMAL) HRESULT {
	addr := lazyAddr(&pVarDecCmp, libOleaut32, "VarDecCmp")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pdecLeft)), uintptr(unsafe.Pointer(pdecRight)))
	return HRESULT(ret)
}

func VarDecCmpR8(pdecLeft *DECIMAL, dblRight float64) HRESULT {
	addr := lazyAddr(&pVarDecCmpR8, libOleaut32, "VarDecCmpR8")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pdecLeft)), uintptr(dblRight))
	return HRESULT(ret)
}

func VarCyAdd(cyLeft CY, cyRight CY, pcyResult *CY) HRESULT {
	addr := lazyAddr(&pVarCyAdd, libOleaut32, "VarCyAdd")
	ret, _, _ := syscall.SyscallN(addr, *(*uintptr)(unsafe.Pointer(&cyLeft)), *(*uintptr)(unsafe.Pointer(&cyRight)), uintptr(unsafe.Pointer(pcyResult)))
	return HRESULT(ret)
}

func VarCyMul(cyLeft CY, cyRight CY, pcyResult *CY) HRESULT {
	addr := lazyAddr(&pVarCyMul, libOleaut32, "VarCyMul")
	ret, _, _ := syscall.SyscallN(addr, *(*uintptr)(unsafe.Pointer(&cyLeft)), *(*uintptr)(unsafe.Pointer(&cyRight)), uintptr(unsafe.Pointer(pcyResult)))
	return HRESULT(ret)
}

func VarCyMulI4(cyLeft CY, lRight int32, pcyResult *CY) HRESULT {
	addr := lazyAddr(&pVarCyMulI4, libOleaut32, "VarCyMulI4")
	ret, _, _ := syscall.SyscallN(addr, *(*uintptr)(unsafe.Pointer(&cyLeft)), uintptr(lRight), uintptr(unsafe.Pointer(pcyResult)))
	return HRESULT(ret)
}

func VarCyMulI8(cyLeft CY, lRight int64, pcyResult *CY) HRESULT {
	addr := lazyAddr(&pVarCyMulI8, libOleaut32, "VarCyMulI8")
	ret, _, _ := syscall.SyscallN(addr, *(*uintptr)(unsafe.Pointer(&cyLeft)), uintptr(lRight), uintptr(unsafe.Pointer(pcyResult)))
	return HRESULT(ret)
}

func VarCySub(cyLeft CY, cyRight CY, pcyResult *CY) HRESULT {
	addr := lazyAddr(&pVarCySub, libOleaut32, "VarCySub")
	ret, _, _ := syscall.SyscallN(addr, *(*uintptr)(unsafe.Pointer(&cyLeft)), *(*uintptr)(unsafe.Pointer(&cyRight)), uintptr(unsafe.Pointer(pcyResult)))
	return HRESULT(ret)
}

func VarCyAbs(cyIn CY, pcyResult *CY) HRESULT {
	addr := lazyAddr(&pVarCyAbs, libOleaut32, "VarCyAbs")
	ret, _, _ := syscall.SyscallN(addr, *(*uintptr)(unsafe.Pointer(&cyIn)), uintptr(unsafe.Pointer(pcyResult)))
	return HRESULT(ret)
}

func VarCyFix(cyIn CY, pcyResult *CY) HRESULT {
	addr := lazyAddr(&pVarCyFix, libOleaut32, "VarCyFix")
	ret, _, _ := syscall.SyscallN(addr, *(*uintptr)(unsafe.Pointer(&cyIn)), uintptr(unsafe.Pointer(pcyResult)))
	return HRESULT(ret)
}

func VarCyInt(cyIn CY, pcyResult *CY) HRESULT {
	addr := lazyAddr(&pVarCyInt, libOleaut32, "VarCyInt")
	ret, _, _ := syscall.SyscallN(addr, *(*uintptr)(unsafe.Pointer(&cyIn)), uintptr(unsafe.Pointer(pcyResult)))
	return HRESULT(ret)
}

func VarCyNeg(cyIn CY, pcyResult *CY) HRESULT {
	addr := lazyAddr(&pVarCyNeg, libOleaut32, "VarCyNeg")
	ret, _, _ := syscall.SyscallN(addr, *(*uintptr)(unsafe.Pointer(&cyIn)), uintptr(unsafe.Pointer(pcyResult)))
	return HRESULT(ret)
}

func VarCyRound(cyIn CY, cDecimals int32, pcyResult *CY) HRESULT {
	addr := lazyAddr(&pVarCyRound, libOleaut32, "VarCyRound")
	ret, _, _ := syscall.SyscallN(addr, *(*uintptr)(unsafe.Pointer(&cyIn)), uintptr(cDecimals), uintptr(unsafe.Pointer(pcyResult)))
	return HRESULT(ret)
}

func VarCyCmp(cyLeft CY, cyRight CY) HRESULT {
	addr := lazyAddr(&pVarCyCmp, libOleaut32, "VarCyCmp")
	ret, _, _ := syscall.SyscallN(addr, *(*uintptr)(unsafe.Pointer(&cyLeft)), *(*uintptr)(unsafe.Pointer(&cyRight)))
	return HRESULT(ret)
}

func VarCyCmpR8(cyLeft CY, dblRight float64) HRESULT {
	addr := lazyAddr(&pVarCyCmpR8, libOleaut32, "VarCyCmpR8")
	ret, _, _ := syscall.SyscallN(addr, *(*uintptr)(unsafe.Pointer(&cyLeft)), uintptr(dblRight))
	return HRESULT(ret)
}

func VarBstrCat(bstrLeft BSTR, bstrRight BSTR, pbstrResult **uint16) HRESULT {
	addr := lazyAddr(&pVarBstrCat, libOleaut32, "VarBstrCat")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(bstrLeft)), uintptr(unsafe.Pointer(bstrRight)), uintptr(unsafe.Pointer(pbstrResult)))
	return HRESULT(ret)
}

func VarBstrCmp(bstrLeft BSTR, bstrRight BSTR, lcid uint32, dwFlags uint32) HRESULT {
	addr := lazyAddr(&pVarBstrCmp, libOleaut32, "VarBstrCmp")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(bstrLeft)), uintptr(unsafe.Pointer(bstrRight)), uintptr(lcid), uintptr(dwFlags))
	return HRESULT(ret)
}

func VarR8Pow(dblLeft float64, dblRight float64, pdblResult *float64) HRESULT {
	addr := lazyAddr(&pVarR8Pow, libOleaut32, "VarR8Pow")
	ret, _, _ := syscall.SyscallN(addr, uintptr(dblLeft), uintptr(dblRight), uintptr(unsafe.Pointer(pdblResult)))
	return HRESULT(ret)
}

func VarR4CmpR8(fltLeft float32, dblRight float64) HRESULT {
	addr := lazyAddr(&pVarR4CmpR8, libOleaut32, "VarR4CmpR8")
	ret, _, _ := syscall.SyscallN(addr, uintptr(fltLeft), uintptr(dblRight))
	return HRESULT(ret)
}

func VarR8Round(dblIn float64, cDecimals int32, pdblResult *float64) HRESULT {
	addr := lazyAddr(&pVarR8Round, libOleaut32, "VarR8Round")
	ret, _, _ := syscall.SyscallN(addr, uintptr(dblIn), uintptr(cDecimals), uintptr(unsafe.Pointer(pdblResult)))
	return HRESULT(ret)
}

func VarDateFromUdate(pudateIn *UDATE, dwFlags uint32, pdateOut *float64) HRESULT {
	addr := lazyAddr(&pVarDateFromUdate, libOleaut32, "VarDateFromUdate")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pudateIn)), uintptr(dwFlags), uintptr(unsafe.Pointer(pdateOut)))
	return HRESULT(ret)
}

func VarDateFromUdateEx(pudateIn *UDATE, lcid uint32, dwFlags uint32, pdateOut *float64) HRESULT {
	addr := lazyAddr(&pVarDateFromUdateEx, libOleaut32, "VarDateFromUdateEx")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pudateIn)), uintptr(lcid), uintptr(dwFlags), uintptr(unsafe.Pointer(pdateOut)))
	return HRESULT(ret)
}

func VarUdateFromDate(dateIn float64, dwFlags uint32, pudateOut *UDATE) HRESULT {
	addr := lazyAddr(&pVarUdateFromDate, libOleaut32, "VarUdateFromDate")
	ret, _, _ := syscall.SyscallN(addr, uintptr(dateIn), uintptr(dwFlags), uintptr(unsafe.Pointer(pudateOut)))
	return HRESULT(ret)
}

func GetAltMonthNames(lcid uint32, prgp **PWSTR) HRESULT {
	addr := lazyAddr(&pGetAltMonthNames, libOleaut32, "GetAltMonthNames")
	ret, _, _ := syscall.SyscallN(addr, uintptr(lcid), uintptr(unsafe.Pointer(prgp)))
	return HRESULT(ret)
}

func VarFormat(pvarIn *VARIANT, pstrFormat PWSTR, iFirstDay int32, iFirstWeek int32, dwFlags uint32, pbstrOut *BSTR) HRESULT {
	addr := lazyAddr(&pVarFormat, libOleaut32, "VarFormat")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pvarIn)), uintptr(unsafe.Pointer(pstrFormat)), uintptr(iFirstDay), uintptr(iFirstWeek), uintptr(dwFlags), uintptr(unsafe.Pointer(pbstrOut)))
	return HRESULT(ret)
}

func VarFormatDateTime(pvarIn *VARIANT, iNamedFormat int32, dwFlags uint32, pbstrOut *BSTR) HRESULT {
	addr := lazyAddr(&pVarFormatDateTime, libOleaut32, "VarFormatDateTime")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pvarIn)), uintptr(iNamedFormat), uintptr(dwFlags), uintptr(unsafe.Pointer(pbstrOut)))
	return HRESULT(ret)
}

func VarFormatNumber(pvarIn *VARIANT, iNumDig int32, iIncLead int32, iUseParens int32, iGroup int32, dwFlags uint32, pbstrOut *BSTR) HRESULT {
	addr := lazyAddr(&pVarFormatNumber, libOleaut32, "VarFormatNumber")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pvarIn)), uintptr(iNumDig), uintptr(iIncLead), uintptr(iUseParens), uintptr(iGroup), uintptr(dwFlags), uintptr(unsafe.Pointer(pbstrOut)))
	return HRESULT(ret)
}

func VarFormatPercent(pvarIn *VARIANT, iNumDig int32, iIncLead int32, iUseParens int32, iGroup int32, dwFlags uint32, pbstrOut *BSTR) HRESULT {
	addr := lazyAddr(&pVarFormatPercent, libOleaut32, "VarFormatPercent")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pvarIn)), uintptr(iNumDig), uintptr(iIncLead), uintptr(iUseParens), uintptr(iGroup), uintptr(dwFlags), uintptr(unsafe.Pointer(pbstrOut)))
	return HRESULT(ret)
}

func VarFormatCurrency(pvarIn *VARIANT, iNumDig int32, iIncLead int32, iUseParens int32, iGroup int32, dwFlags uint32, pbstrOut *BSTR) HRESULT {
	addr := lazyAddr(&pVarFormatCurrency, libOleaut32, "VarFormatCurrency")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pvarIn)), uintptr(iNumDig), uintptr(iIncLead), uintptr(iUseParens), uintptr(iGroup), uintptr(dwFlags), uintptr(unsafe.Pointer(pbstrOut)))
	return HRESULT(ret)
}

func VarWeekdayName(iWeekday int32, fAbbrev int32, iFirstDay int32, dwFlags uint32, pbstrOut *BSTR) HRESULT {
	addr := lazyAddr(&pVarWeekdayName, libOleaut32, "VarWeekdayName")
	ret, _, _ := syscall.SyscallN(addr, uintptr(iWeekday), uintptr(fAbbrev), uintptr(iFirstDay), uintptr(dwFlags), uintptr(unsafe.Pointer(pbstrOut)))
	return HRESULT(ret)
}

func VarMonthName(iMonth int32, fAbbrev int32, dwFlags uint32, pbstrOut *BSTR) HRESULT {
	addr := lazyAddr(&pVarMonthName, libOleaut32, "VarMonthName")
	ret, _, _ := syscall.SyscallN(addr, uintptr(iMonth), uintptr(fAbbrev), uintptr(dwFlags), uintptr(unsafe.Pointer(pbstrOut)))
	return HRESULT(ret)
}

func VarFormatFromTokens(pvarIn *VARIANT, pstrFormat PWSTR, pbTokCur *byte, dwFlags uint32, pbstrOut *BSTR, lcid uint32) HRESULT {
	addr := lazyAddr(&pVarFormatFromTokens, libOleaut32, "VarFormatFromTokens")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pvarIn)), uintptr(unsafe.Pointer(pstrFormat)), uintptr(unsafe.Pointer(pbTokCur)), uintptr(dwFlags), uintptr(unsafe.Pointer(pbstrOut)), uintptr(lcid))
	return HRESULT(ret)
}

func VarTokenizeFormatString(pstrFormat PWSTR, rgbTok *byte, cbTok int32, iFirstDay int32, iFirstWeek int32, lcid uint32, pcbActual *int32) HRESULT {
	addr := lazyAddr(&pVarTokenizeFormatString, libOleaut32, "VarTokenizeFormatString")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pstrFormat)), uintptr(unsafe.Pointer(rgbTok)), uintptr(cbTok), uintptr(iFirstDay), uintptr(iFirstWeek), uintptr(lcid), uintptr(unsafe.Pointer(pcbActual)))
	return HRESULT(ret)
}

func LHashValOfNameSysA(syskind SYSKIND, lcid uint32, szName PSTR) uint32 {
	addr := lazyAddr(&pLHashValOfNameSysA, libOleaut32, "LHashValOfNameSysA")
	ret, _, _ := syscall.SyscallN(addr, uintptr(syskind), uintptr(lcid), uintptr(unsafe.Pointer(szName)))
	return uint32(ret)
}

func LHashValOfNameSys(syskind SYSKIND, lcid uint32, szName PWSTR) uint32 {
	addr := lazyAddr(&pLHashValOfNameSys, libOleaut32, "LHashValOfNameSys")
	ret, _, _ := syscall.SyscallN(addr, uintptr(syskind), uintptr(lcid), uintptr(unsafe.Pointer(szName)))
	return uint32(ret)
}

func LoadTypeLib(szFile PWSTR, pptlib **ITypeLib) HRESULT {
	addr := lazyAddr(&pLoadTypeLib, libOleaut32, "LoadTypeLib")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(szFile)), uintptr(unsafe.Pointer(pptlib)))
	return HRESULT(ret)
}

func LoadTypeLibEx(szFile PWSTR, regkind REGKIND, pptlib **ITypeLib) HRESULT {
	addr := lazyAddr(&pLoadTypeLibEx, libOleaut32, "LoadTypeLibEx")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(szFile)), uintptr(regkind), uintptr(unsafe.Pointer(pptlib)))
	return HRESULT(ret)
}

func LoadRegTypeLib(rguid *syscall.GUID, wVerMajor uint16, wVerMinor uint16, lcid uint32, pptlib **ITypeLib) HRESULT {
	addr := lazyAddr(&pLoadRegTypeLib, libOleaut32, "LoadRegTypeLib")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(rguid)), uintptr(wVerMajor), uintptr(wVerMinor), uintptr(lcid), uintptr(unsafe.Pointer(pptlib)))
	return HRESULT(ret)
}

func QueryPathOfRegTypeLib(guid *syscall.GUID, wMaj uint16, wMin uint16, lcid uint32, lpbstrPathName **uint16) HRESULT {
	addr := lazyAddr(&pQueryPathOfRegTypeLib, libOleaut32, "QueryPathOfRegTypeLib")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(guid)), uintptr(wMaj), uintptr(wMin), uintptr(lcid), uintptr(unsafe.Pointer(lpbstrPathName)))
	return HRESULT(ret)
}

func RegisterTypeLib(ptlib *ITypeLib, szFullPath PWSTR, szHelpDir PWSTR) HRESULT {
	addr := lazyAddr(&pRegisterTypeLib, libOleaut32, "RegisterTypeLib")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(ptlib)), uintptr(unsafe.Pointer(szFullPath)), uintptr(unsafe.Pointer(szHelpDir)))
	return HRESULT(ret)
}

func UnRegisterTypeLib(libID *syscall.GUID, wVerMajor uint16, wVerMinor uint16, lcid uint32, syskind SYSKIND) HRESULT {
	addr := lazyAddr(&pUnRegisterTypeLib, libOleaut32, "UnRegisterTypeLib")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(libID)), uintptr(wVerMajor), uintptr(wVerMinor), uintptr(lcid), uintptr(syskind))
	return HRESULT(ret)
}

func RegisterTypeLibForUser(ptlib *ITypeLib, szFullPath PWSTR, szHelpDir PWSTR) HRESULT {
	addr := lazyAddr(&pRegisterTypeLibForUser, libOleaut32, "RegisterTypeLibForUser")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(ptlib)), uintptr(unsafe.Pointer(szFullPath)), uintptr(unsafe.Pointer(szHelpDir)))
	return HRESULT(ret)
}

func UnRegisterTypeLibForUser(libID *syscall.GUID, wMajorVerNum uint16, wMinorVerNum uint16, lcid uint32, syskind SYSKIND) HRESULT {
	addr := lazyAddr(&pUnRegisterTypeLibForUser, libOleaut32, "UnRegisterTypeLibForUser")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(libID)), uintptr(wMajorVerNum), uintptr(wMinorVerNum), uintptr(lcid), uintptr(syskind))
	return HRESULT(ret)
}

func CreateTypeLib(syskind SYSKIND, szFile PWSTR, ppctlib **ICreateTypeLib) HRESULT {
	addr := lazyAddr(&pCreateTypeLib, libOleaut32, "CreateTypeLib")
	ret, _, _ := syscall.SyscallN(addr, uintptr(syskind), uintptr(unsafe.Pointer(szFile)), uintptr(unsafe.Pointer(ppctlib)))
	return HRESULT(ret)
}

func CreateTypeLib2(syskind SYSKIND, szFile PWSTR, ppctlib **ICreateTypeLib2) HRESULT {
	addr := lazyAddr(&pCreateTypeLib2, libOleaut32, "CreateTypeLib2")
	ret, _, _ := syscall.SyscallN(addr, uintptr(syskind), uintptr(unsafe.Pointer(szFile)), uintptr(unsafe.Pointer(ppctlib)))
	return HRESULT(ret)
}

func DispGetParam(pdispparams *DISPPARAMS, position uint32, vtTarg uint16, pvarResult *VARIANT, puArgErr *uint32) HRESULT {
	addr := lazyAddr(&pDispGetParam, libOleaut32, "DispGetParam")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pdispparams)), uintptr(position), uintptr(vtTarg), uintptr(unsafe.Pointer(pvarResult)), uintptr(unsafe.Pointer(puArgErr)))
	return HRESULT(ret)
}

func DispGetIDsOfNames(ptinfo *ITypeInfo, rgszNames *PWSTR, cNames uint32, rgdispid *int32) HRESULT {
	addr := lazyAddr(&pDispGetIDsOfNames, libOleaut32, "DispGetIDsOfNames")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(ptinfo)), uintptr(unsafe.Pointer(rgszNames)), uintptr(cNames), uintptr(unsafe.Pointer(rgdispid)))
	return HRESULT(ret)
}

func DispInvoke(_this unsafe.Pointer, ptinfo *ITypeInfo, dispidMember int32, wFlags uint16, pparams *DISPPARAMS, pvarResult *VARIANT, pexcepinfo *EXCEPINFO, puArgErr *uint32) HRESULT {
	addr := lazyAddr(&pDispInvoke, libOleaut32, "DispInvoke")
	ret, _, _ := syscall.SyscallN(addr, uintptr(_this), uintptr(unsafe.Pointer(ptinfo)), uintptr(dispidMember), uintptr(wFlags), uintptr(unsafe.Pointer(pparams)), uintptr(unsafe.Pointer(pvarResult)), uintptr(unsafe.Pointer(pexcepinfo)), uintptr(unsafe.Pointer(puArgErr)))
	return HRESULT(ret)
}

func CreateDispTypeInfo(pidata *INTERFACEDATA, lcid uint32, pptinfo **ITypeInfo) HRESULT {
	addr := lazyAddr(&pCreateDispTypeInfo, libOleaut32, "CreateDispTypeInfo")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pidata)), uintptr(lcid), uintptr(unsafe.Pointer(pptinfo)))
	return HRESULT(ret)
}

func CreateStdDispatch(punkOuter *IUnknown, pvThis unsafe.Pointer, ptinfo *ITypeInfo, ppunkStdDisp **IUnknown) HRESULT {
	addr := lazyAddr(&pCreateStdDispatch, libOleaut32, "CreateStdDispatch")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(punkOuter)), uintptr(pvThis), uintptr(unsafe.Pointer(ptinfo)), uintptr(unsafe.Pointer(ppunkStdDisp)))
	return HRESULT(ret)
}

func DispCallFunc(pvInstance unsafe.Pointer, oVft uintptr, cc CALLCONV, vtReturn uint16, cActuals uint32, prgvt *uint16, prgpvarg **VARIANT, pvargResult *VARIANT) HRESULT {
	addr := lazyAddr(&pDispCallFunc, libOleaut32, "DispCallFunc")
	ret, _, _ := syscall.SyscallN(addr, uintptr(pvInstance), oVft, uintptr(cc), uintptr(vtReturn), uintptr(cActuals), uintptr(unsafe.Pointer(prgvt)), uintptr(unsafe.Pointer(prgpvarg)), uintptr(unsafe.Pointer(pvargResult)))
	return HRESULT(ret)
}

func RegisterActiveObject(punk *IUnknown, rclsid *syscall.GUID, dwFlags uint32, pdwRegister *uint32) HRESULT {
	addr := lazyAddr(&pRegisterActiveObject, libOleaut32, "RegisterActiveObject")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(punk)), uintptr(unsafe.Pointer(rclsid)), uintptr(dwFlags), uintptr(unsafe.Pointer(pdwRegister)))
	return HRESULT(ret)
}

func RevokeActiveObject(dwRegister uint32, pvReserved unsafe.Pointer) HRESULT {
	addr := lazyAddr(&pRevokeActiveObject, libOleaut32, "RevokeActiveObject")
	ret, _, _ := syscall.SyscallN(addr, uintptr(dwRegister), uintptr(pvReserved))
	return HRESULT(ret)
}

func GetActiveObject(rclsid *syscall.GUID, pvReserved unsafe.Pointer, ppunk **IUnknown) HRESULT {
	addr := lazyAddr(&pGetActiveObject, libOleaut32, "GetActiveObject")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(rclsid)), uintptr(pvReserved), uintptr(unsafe.Pointer(ppunk)))
	return HRESULT(ret)
}

func CreateErrorInfo(pperrinfo **ICreateErrorInfo) HRESULT {
	addr := lazyAddr(&pCreateErrorInfo, libOleaut32, "CreateErrorInfo")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pperrinfo)))
	return HRESULT(ret)
}

func GetRecordInfoFromTypeInfo(pTypeInfo *ITypeInfo, ppRecInfo **IRecordInfo) HRESULT {
	addr := lazyAddr(&pGetRecordInfoFromTypeInfo, libOleaut32, "GetRecordInfoFromTypeInfo")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pTypeInfo)), uintptr(unsafe.Pointer(ppRecInfo)))
	return HRESULT(ret)
}

func GetRecordInfoFromGuids(rGuidTypeLib *syscall.GUID, uVerMajor uint32, uVerMinor uint32, lcid uint32, rGuidTypeInfo *syscall.GUID, ppRecInfo **IRecordInfo) HRESULT {
	addr := lazyAddr(&pGetRecordInfoFromGuids, libOleaut32, "GetRecordInfoFromGuids")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(rGuidTypeLib)), uintptr(uVerMajor), uintptr(uVerMinor), uintptr(lcid), uintptr(unsafe.Pointer(rGuidTypeInfo)), uintptr(unsafe.Pointer(ppRecInfo)))
	return HRESULT(ret)
}

func OaBuildVersion() uint32 {
	addr := lazyAddr(&pOaBuildVersion, libOleaut32, "OaBuildVersion")
	ret, _, _ := syscall.SyscallN(addr)
	return uint32(ret)
}

func ClearCustData(pCustData *CUSTDATA) {
	addr := lazyAddr(&pClearCustData, libOleaut32, "ClearCustData")
	syscall.SyscallN(addr, uintptr(unsafe.Pointer(pCustData)))
}

func OaEnablePerUserTLibRegistration() {
	addr := lazyAddr(&pOaEnablePerUserTLibRegistration, libOleaut32, "OaEnablePerUserTLibRegistration")
	syscall.SyscallN(addr)
}

func OleBuildVersion() uint32 {
	addr := lazyAddr(&pOleBuildVersion, libOle32, "OleBuildVersion")
	ret, _, _ := syscall.SyscallN(addr)
	return uint32(ret)
}

func OleInitialize(pvReserved unsafe.Pointer) HRESULT {
	addr := lazyAddr(&pOleInitialize, libOle32, "OleInitialize")
	ret, _, _ := syscall.SyscallN(addr, uintptr(pvReserved))
	return HRESULT(ret)
}

func OleUninitialize() {
	addr := lazyAddr(&pOleUninitialize, libOle32, "OleUninitialize")
	syscall.SyscallN(addr)
}

func OleQueryLinkFromData(pSrcDataObject *IDataObject) HRESULT {
	addr := lazyAddr(&pOleQueryLinkFromData, libOle32, "OleQueryLinkFromData")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pSrcDataObject)))
	return HRESULT(ret)
}

func OleQueryCreateFromData(pSrcDataObject *IDataObject) HRESULT {
	addr := lazyAddr(&pOleQueryCreateFromData, libOle32, "OleQueryCreateFromData")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pSrcDataObject)))
	return HRESULT(ret)
}

func OleCreate(rclsid *syscall.GUID, riid *syscall.GUID, renderopt uint32, pFormatEtc *FORMATETC, pClientSite *IOleClientSite, pStg *IStorage, ppvObj unsafe.Pointer) HRESULT {
	addr := lazyAddr(&pOleCreate, libOle32, "OleCreate")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(rclsid)), uintptr(unsafe.Pointer(riid)), uintptr(renderopt), uintptr(unsafe.Pointer(pFormatEtc)), uintptr(unsafe.Pointer(pClientSite)), uintptr(unsafe.Pointer(pStg)), uintptr(ppvObj))
	return HRESULT(ret)
}

func OleCreateEx(rclsid *syscall.GUID, riid *syscall.GUID, dwFlags uint32, renderopt uint32, cFormats uint32, rgAdvf *uint32, rgFormatEtc *FORMATETC, lpAdviseSink *IAdviseSink, rgdwConnection *uint32, pClientSite *IOleClientSite, pStg *IStorage, ppvObj unsafe.Pointer) HRESULT {
	addr := lazyAddr(&pOleCreateEx, libOle32, "OleCreateEx")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(rclsid)), uintptr(unsafe.Pointer(riid)), uintptr(dwFlags), uintptr(renderopt), uintptr(cFormats), uintptr(unsafe.Pointer(rgAdvf)), uintptr(unsafe.Pointer(rgFormatEtc)), uintptr(unsafe.Pointer(lpAdviseSink)), uintptr(unsafe.Pointer(rgdwConnection)), uintptr(unsafe.Pointer(pClientSite)), uintptr(unsafe.Pointer(pStg)), uintptr(ppvObj))
	return HRESULT(ret)
}

func OleCreateFromData(pSrcDataObj *IDataObject, riid *syscall.GUID, renderopt uint32, pFormatEtc *FORMATETC, pClientSite *IOleClientSite, pStg *IStorage, ppvObj unsafe.Pointer) HRESULT {
	addr := lazyAddr(&pOleCreateFromData, libOle32, "OleCreateFromData")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pSrcDataObj)), uintptr(unsafe.Pointer(riid)), uintptr(renderopt), uintptr(unsafe.Pointer(pFormatEtc)), uintptr(unsafe.Pointer(pClientSite)), uintptr(unsafe.Pointer(pStg)), uintptr(ppvObj))
	return HRESULT(ret)
}

func OleCreateFromDataEx(pSrcDataObj *IDataObject, riid *syscall.GUID, dwFlags uint32, renderopt uint32, cFormats uint32, rgAdvf *uint32, rgFormatEtc *FORMATETC, lpAdviseSink *IAdviseSink, rgdwConnection *uint32, pClientSite *IOleClientSite, pStg *IStorage, ppvObj unsafe.Pointer) HRESULT {
	addr := lazyAddr(&pOleCreateFromDataEx, libOle32, "OleCreateFromDataEx")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pSrcDataObj)), uintptr(unsafe.Pointer(riid)), uintptr(dwFlags), uintptr(renderopt), uintptr(cFormats), uintptr(unsafe.Pointer(rgAdvf)), uintptr(unsafe.Pointer(rgFormatEtc)), uintptr(unsafe.Pointer(lpAdviseSink)), uintptr(unsafe.Pointer(rgdwConnection)), uintptr(unsafe.Pointer(pClientSite)), uintptr(unsafe.Pointer(pStg)), uintptr(ppvObj))
	return HRESULT(ret)
}

func OleCreateLinkFromData(pSrcDataObj *IDataObject, riid *syscall.GUID, renderopt uint32, pFormatEtc *FORMATETC, pClientSite *IOleClientSite, pStg *IStorage, ppvObj unsafe.Pointer) HRESULT {
	addr := lazyAddr(&pOleCreateLinkFromData, libOle32, "OleCreateLinkFromData")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pSrcDataObj)), uintptr(unsafe.Pointer(riid)), uintptr(renderopt), uintptr(unsafe.Pointer(pFormatEtc)), uintptr(unsafe.Pointer(pClientSite)), uintptr(unsafe.Pointer(pStg)), uintptr(ppvObj))
	return HRESULT(ret)
}

func OleCreateLinkFromDataEx(pSrcDataObj *IDataObject, riid *syscall.GUID, dwFlags uint32, renderopt uint32, cFormats uint32, rgAdvf *uint32, rgFormatEtc *FORMATETC, lpAdviseSink *IAdviseSink, rgdwConnection *uint32, pClientSite *IOleClientSite, pStg *IStorage, ppvObj unsafe.Pointer) HRESULT {
	addr := lazyAddr(&pOleCreateLinkFromDataEx, libOle32, "OleCreateLinkFromDataEx")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pSrcDataObj)), uintptr(unsafe.Pointer(riid)), uintptr(dwFlags), uintptr(renderopt), uintptr(cFormats), uintptr(unsafe.Pointer(rgAdvf)), uintptr(unsafe.Pointer(rgFormatEtc)), uintptr(unsafe.Pointer(lpAdviseSink)), uintptr(unsafe.Pointer(rgdwConnection)), uintptr(unsafe.Pointer(pClientSite)), uintptr(unsafe.Pointer(pStg)), uintptr(ppvObj))
	return HRESULT(ret)
}

func OleCreateStaticFromData(pSrcDataObj *IDataObject, iid *syscall.GUID, renderopt uint32, pFormatEtc *FORMATETC, pClientSite *IOleClientSite, pStg *IStorage, ppvObj unsafe.Pointer) HRESULT {
	addr := lazyAddr(&pOleCreateStaticFromData, libOle32, "OleCreateStaticFromData")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pSrcDataObj)), uintptr(unsafe.Pointer(iid)), uintptr(renderopt), uintptr(unsafe.Pointer(pFormatEtc)), uintptr(unsafe.Pointer(pClientSite)), uintptr(unsafe.Pointer(pStg)), uintptr(ppvObj))
	return HRESULT(ret)
}

func OleCreateLink(pmkLinkSrc *IMoniker, riid *syscall.GUID, renderopt uint32, lpFormatEtc *FORMATETC, pClientSite *IOleClientSite, pStg *IStorage, ppvObj unsafe.Pointer) HRESULT {
	addr := lazyAddr(&pOleCreateLink, libOle32, "OleCreateLink")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pmkLinkSrc)), uintptr(unsafe.Pointer(riid)), uintptr(renderopt), uintptr(unsafe.Pointer(lpFormatEtc)), uintptr(unsafe.Pointer(pClientSite)), uintptr(unsafe.Pointer(pStg)), uintptr(ppvObj))
	return HRESULT(ret)
}

func OleCreateLinkEx(pmkLinkSrc *IMoniker, riid *syscall.GUID, dwFlags uint32, renderopt uint32, cFormats uint32, rgAdvf *uint32, rgFormatEtc *FORMATETC, lpAdviseSink *IAdviseSink, rgdwConnection *uint32, pClientSite *IOleClientSite, pStg *IStorage, ppvObj unsafe.Pointer) HRESULT {
	addr := lazyAddr(&pOleCreateLinkEx, libOle32, "OleCreateLinkEx")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pmkLinkSrc)), uintptr(unsafe.Pointer(riid)), uintptr(dwFlags), uintptr(renderopt), uintptr(cFormats), uintptr(unsafe.Pointer(rgAdvf)), uintptr(unsafe.Pointer(rgFormatEtc)), uintptr(unsafe.Pointer(lpAdviseSink)), uintptr(unsafe.Pointer(rgdwConnection)), uintptr(unsafe.Pointer(pClientSite)), uintptr(unsafe.Pointer(pStg)), uintptr(ppvObj))
	return HRESULT(ret)
}

func OleCreateLinkToFile(lpszFileName PWSTR, riid *syscall.GUID, renderopt uint32, lpFormatEtc *FORMATETC, pClientSite *IOleClientSite, pStg *IStorage, ppvObj unsafe.Pointer) HRESULT {
	addr := lazyAddr(&pOleCreateLinkToFile, libOle32, "OleCreateLinkToFile")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpszFileName)), uintptr(unsafe.Pointer(riid)), uintptr(renderopt), uintptr(unsafe.Pointer(lpFormatEtc)), uintptr(unsafe.Pointer(pClientSite)), uintptr(unsafe.Pointer(pStg)), uintptr(ppvObj))
	return HRESULT(ret)
}

func OleCreateLinkToFileEx(lpszFileName PWSTR, riid *syscall.GUID, dwFlags uint32, renderopt uint32, cFormats uint32, rgAdvf *uint32, rgFormatEtc *FORMATETC, lpAdviseSink *IAdviseSink, rgdwConnection *uint32, pClientSite *IOleClientSite, pStg *IStorage, ppvObj unsafe.Pointer) HRESULT {
	addr := lazyAddr(&pOleCreateLinkToFileEx, libOle32, "OleCreateLinkToFileEx")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpszFileName)), uintptr(unsafe.Pointer(riid)), uintptr(dwFlags), uintptr(renderopt), uintptr(cFormats), uintptr(unsafe.Pointer(rgAdvf)), uintptr(unsafe.Pointer(rgFormatEtc)), uintptr(unsafe.Pointer(lpAdviseSink)), uintptr(unsafe.Pointer(rgdwConnection)), uintptr(unsafe.Pointer(pClientSite)), uintptr(unsafe.Pointer(pStg)), uintptr(ppvObj))
	return HRESULT(ret)
}

func OleCreateFromFile(rclsid *syscall.GUID, lpszFileName PWSTR, riid *syscall.GUID, renderopt uint32, lpFormatEtc *FORMATETC, pClientSite *IOleClientSite, pStg *IStorage, ppvObj unsafe.Pointer) HRESULT {
	addr := lazyAddr(&pOleCreateFromFile, libOle32, "OleCreateFromFile")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(rclsid)), uintptr(unsafe.Pointer(lpszFileName)), uintptr(unsafe.Pointer(riid)), uintptr(renderopt), uintptr(unsafe.Pointer(lpFormatEtc)), uintptr(unsafe.Pointer(pClientSite)), uintptr(unsafe.Pointer(pStg)), uintptr(ppvObj))
	return HRESULT(ret)
}

func OleCreateFromFileEx(rclsid *syscall.GUID, lpszFileName PWSTR, riid *syscall.GUID, dwFlags uint32, renderopt uint32, cFormats uint32, rgAdvf *uint32, rgFormatEtc *FORMATETC, lpAdviseSink *IAdviseSink, rgdwConnection *uint32, pClientSite *IOleClientSite, pStg *IStorage, ppvObj unsafe.Pointer) HRESULT {
	addr := lazyAddr(&pOleCreateFromFileEx, libOle32, "OleCreateFromFileEx")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(rclsid)), uintptr(unsafe.Pointer(lpszFileName)), uintptr(unsafe.Pointer(riid)), uintptr(dwFlags), uintptr(renderopt), uintptr(cFormats), uintptr(unsafe.Pointer(rgAdvf)), uintptr(unsafe.Pointer(rgFormatEtc)), uintptr(unsafe.Pointer(lpAdviseSink)), uintptr(unsafe.Pointer(rgdwConnection)), uintptr(unsafe.Pointer(pClientSite)), uintptr(unsafe.Pointer(pStg)), uintptr(ppvObj))
	return HRESULT(ret)
}

func OleLoad(pStg *IStorage, riid *syscall.GUID, pClientSite *IOleClientSite, ppvObj unsafe.Pointer) HRESULT {
	addr := lazyAddr(&pOleLoad, libOle32, "OleLoad")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pStg)), uintptr(unsafe.Pointer(riid)), uintptr(unsafe.Pointer(pClientSite)), uintptr(ppvObj))
	return HRESULT(ret)
}

func OleSave(pPS *IPersistStorage, pStg *IStorage, fSameAsLoad BOOL) HRESULT {
	addr := lazyAddr(&pOleSave, libOle32, "OleSave")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pPS)), uintptr(unsafe.Pointer(pStg)), uintptr(fSameAsLoad))
	return HRESULT(ret)
}

func OleLoadFromStream(pStm *IStream, iidInterface *syscall.GUID, ppvObj unsafe.Pointer) HRESULT {
	addr := lazyAddr(&pOleLoadFromStream, libOle32, "OleLoadFromStream")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pStm)), uintptr(unsafe.Pointer(iidInterface)), uintptr(ppvObj))
	return HRESULT(ret)
}

func OleSaveToStream(pPStm *IPersistStream, pStm *IStream) HRESULT {
	addr := lazyAddr(&pOleSaveToStream, libOle32, "OleSaveToStream")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pPStm)), uintptr(unsafe.Pointer(pStm)))
	return HRESULT(ret)
}

func OleSetContainedObject(pUnknown *IUnknown, fContained BOOL) HRESULT {
	addr := lazyAddr(&pOleSetContainedObject, libOle32, "OleSetContainedObject")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pUnknown)), uintptr(fContained))
	return HRESULT(ret)
}

func OleNoteObjectVisible(pUnknown *IUnknown, fVisible BOOL) HRESULT {
	addr := lazyAddr(&pOleNoteObjectVisible, libOle32, "OleNoteObjectVisible")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pUnknown)), uintptr(fVisible))
	return HRESULT(ret)
}

func RegisterDragDrop(hwnd HWND, pDropTarget *IDropTarget) HRESULT {
	addr := lazyAddr(&pRegisterDragDrop, libOle32, "RegisterDragDrop")
	ret, _, _ := syscall.SyscallN(addr, hwnd, uintptr(unsafe.Pointer(pDropTarget)))
	return HRESULT(ret)
}

func RevokeDragDrop(hwnd HWND) HRESULT {
	addr := lazyAddr(&pRevokeDragDrop, libOle32, "RevokeDragDrop")
	ret, _, _ := syscall.SyscallN(addr, hwnd)
	return HRESULT(ret)
}

func DoDragDrop(pDataObj *IDataObject, pDropSource *IDropSource, dwOKEffects uint32, pdwEffect *uint32) HRESULT {
	addr := lazyAddr(&pDoDragDrop, libOle32, "DoDragDrop")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pDataObj)), uintptr(unsafe.Pointer(pDropSource)), uintptr(dwOKEffects), uintptr(unsafe.Pointer(pdwEffect)))
	return HRESULT(ret)
}

func OleSetClipboard(pDataObj *IDataObject) HRESULT {
	addr := lazyAddr(&pOleSetClipboard, libOle32, "OleSetClipboard")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pDataObj)))
	return HRESULT(ret)
}

func OleGetClipboard(ppDataObj **IDataObject) HRESULT {
	addr := lazyAddr(&pOleGetClipboard, libOle32, "OleGetClipboard")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(ppDataObj)))
	return HRESULT(ret)
}

func OleGetClipboardWithEnterpriseInfo(dataObject **IDataObject, dataEnterpriseId *PWSTR, sourceDescription *PWSTR, targetDescription *PWSTR, dataDescription *PWSTR) HRESULT {
	addr := lazyAddr(&pOleGetClipboardWithEnterpriseInfo, libOle32, "OleGetClipboardWithEnterpriseInfo")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(dataObject)), uintptr(unsafe.Pointer(dataEnterpriseId)), uintptr(unsafe.Pointer(sourceDescription)), uintptr(unsafe.Pointer(targetDescription)), uintptr(unsafe.Pointer(dataDescription)))
	return HRESULT(ret)
}

func OleFlushClipboard() HRESULT {
	addr := lazyAddr(&pOleFlushClipboard, libOle32, "OleFlushClipboard")
	ret, _, _ := syscall.SyscallN(addr)
	return HRESULT(ret)
}

func OleIsCurrentClipboard(pDataObj *IDataObject) HRESULT {
	addr := lazyAddr(&pOleIsCurrentClipboard, libOle32, "OleIsCurrentClipboard")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pDataObj)))
	return HRESULT(ret)
}

func OleCreateMenuDescriptor(hmenuCombined HMENU, lpMenuWidths *OleMenuGroupWidths) uintptr {
	addr := lazyAddr(&pOleCreateMenuDescriptor, libOle32, "OleCreateMenuDescriptor")
	ret, _, _ := syscall.SyscallN(addr, hmenuCombined, uintptr(unsafe.Pointer(lpMenuWidths)))
	return ret
}

func OleSetMenuDescriptor(holemenu uintptr, hwndFrame HWND, hwndActiveObject HWND, lpFrame *IOleInPlaceFrame, lpActiveObj *IOleInPlaceActiveObject) HRESULT {
	addr := lazyAddr(&pOleSetMenuDescriptor, libOle32, "OleSetMenuDescriptor")
	ret, _, _ := syscall.SyscallN(addr, holemenu, hwndFrame, hwndActiveObject, uintptr(unsafe.Pointer(lpFrame)), uintptr(unsafe.Pointer(lpActiveObj)))
	return HRESULT(ret)
}

func OleDestroyMenuDescriptor(holemenu uintptr) HRESULT {
	addr := lazyAddr(&pOleDestroyMenuDescriptor, libOle32, "OleDestroyMenuDescriptor")
	ret, _, _ := syscall.SyscallN(addr, holemenu)
	return HRESULT(ret)
}

func OleTranslateAccelerator(lpFrame *IOleInPlaceFrame, lpFrameInfo *OIFI, lpmsg *MSG) HRESULT {
	addr := lazyAddr(&pOleTranslateAccelerator, libOle32, "OleTranslateAccelerator")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpFrame)), uintptr(unsafe.Pointer(lpFrameInfo)), uintptr(unsafe.Pointer(lpmsg)))
	return HRESULT(ret)
}

func OleDuplicateData(hSrc HANDLE, cfFormat uint16, uiFlags uint32) HANDLE {
	addr := lazyAddr(&pOleDuplicateData, libOle32, "OleDuplicateData")
	ret, _, _ := syscall.SyscallN(addr, hSrc, uintptr(cfFormat), uintptr(uiFlags))
	return ret
}

func OleDraw(pUnknown *IUnknown, dwAspect uint32, hdcDraw HDC, lprcBounds *RECT) HRESULT {
	addr := lazyAddr(&pOleDraw, libOle32, "OleDraw")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pUnknown)), uintptr(dwAspect), hdcDraw, uintptr(unsafe.Pointer(lprcBounds)))
	return HRESULT(ret)
}

func OleRun(pUnknown *IUnknown) HRESULT {
	addr := lazyAddr(&pOleRun, libOle32, "OleRun")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pUnknown)))
	return HRESULT(ret)
}

func OleIsRunning(pObject *IOleObject) BOOL {
	addr := lazyAddr(&pOleIsRunning, libOle32, "OleIsRunning")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pObject)))
	return BOOL(ret)
}

func OleLockRunning(pUnknown *IUnknown, fLock BOOL, fLastUnlockCloses BOOL) HRESULT {
	addr := lazyAddr(&pOleLockRunning, libOle32, "OleLockRunning")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pUnknown)), uintptr(fLock), uintptr(fLastUnlockCloses))
	return HRESULT(ret)
}

func ReleaseStgMedium(param0 *STGMEDIUM) {
	addr := lazyAddr(&pReleaseStgMedium, libOle32, "ReleaseStgMedium")
	syscall.SyscallN(addr, uintptr(unsafe.Pointer(param0)))
}

func CreateOleAdviseHolder(ppOAHolder **IOleAdviseHolder) HRESULT {
	addr := lazyAddr(&pCreateOleAdviseHolder, libOle32, "CreateOleAdviseHolder")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(ppOAHolder)))
	return HRESULT(ret)
}

func OleCreateDefaultHandler(clsid *syscall.GUID, pUnkOuter *IUnknown, riid *syscall.GUID, lplpObj unsafe.Pointer) HRESULT {
	addr := lazyAddr(&pOleCreateDefaultHandler, libOle32, "OleCreateDefaultHandler")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(clsid)), uintptr(unsafe.Pointer(pUnkOuter)), uintptr(unsafe.Pointer(riid)), uintptr(lplpObj))
	return HRESULT(ret)
}

func OleCreateEmbeddingHelper(clsid *syscall.GUID, pUnkOuter *IUnknown, flags uint32, pCF *IClassFactory, riid *syscall.GUID, lplpObj unsafe.Pointer) HRESULT {
	addr := lazyAddr(&pOleCreateEmbeddingHelper, libOle32, "OleCreateEmbeddingHelper")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(clsid)), uintptr(unsafe.Pointer(pUnkOuter)), uintptr(flags), uintptr(unsafe.Pointer(pCF)), uintptr(unsafe.Pointer(riid)), uintptr(lplpObj))
	return HRESULT(ret)
}

func IsAccelerator(hAccel HACCEL, cAccelEntries int32, lpMsg *MSG, lpwCmd *uint16) BOOL {
	addr := lazyAddr(&pIsAccelerator, libOle32, "IsAccelerator")
	ret, _, _ := syscall.SyscallN(addr, hAccel, uintptr(cAccelEntries), uintptr(unsafe.Pointer(lpMsg)), uintptr(unsafe.Pointer(lpwCmd)))
	return BOOL(ret)
}

func OleGetIconOfFile(lpszPath PWSTR, fUseFileAsLabel BOOL) uintptr {
	addr := lazyAddr(&pOleGetIconOfFile, libOle32, "OleGetIconOfFile")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpszPath)), uintptr(fUseFileAsLabel))
	return ret
}

func OleGetIconOfClass(rclsid *syscall.GUID, lpszLabel PWSTR, fUseTypeAsLabel BOOL) uintptr {
	addr := lazyAddr(&pOleGetIconOfClass, libOle32, "OleGetIconOfClass")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(rclsid)), uintptr(unsafe.Pointer(lpszLabel)), uintptr(fUseTypeAsLabel))
	return ret
}

func OleMetafilePictFromIconAndLabel(hIcon HICON, lpszLabel PWSTR, lpszSourceFile PWSTR, iIconIndex uint32) (uintptr, WIN32_ERROR) {
	addr := lazyAddr(&pOleMetafilePictFromIconAndLabel, libOle32, "OleMetafilePictFromIconAndLabel")
	ret, _, err := syscall.SyscallN(addr, hIcon, uintptr(unsafe.Pointer(lpszLabel)), uintptr(unsafe.Pointer(lpszSourceFile)), uintptr(iIconIndex))
	return ret, WIN32_ERROR(err)
}

func OleRegGetUserType(clsid *syscall.GUID, dwFormOfType uint32, pszUserType *PWSTR) HRESULT {
	addr := lazyAddr(&pOleRegGetUserType, libOle32, "OleRegGetUserType")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(clsid)), uintptr(dwFormOfType), uintptr(unsafe.Pointer(pszUserType)))
	return HRESULT(ret)
}

func OleRegGetMiscStatus(clsid *syscall.GUID, dwAspect uint32, pdwStatus *uint32) HRESULT {
	addr := lazyAddr(&pOleRegGetMiscStatus, libOle32, "OleRegGetMiscStatus")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(clsid)), uintptr(dwAspect), uintptr(unsafe.Pointer(pdwStatus)))
	return HRESULT(ret)
}

func OleRegEnumFormatEtc(clsid *syscall.GUID, dwDirection uint32, ppenum **IEnumFORMATETC) HRESULT {
	addr := lazyAddr(&pOleRegEnumFormatEtc, libOle32, "OleRegEnumFormatEtc")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(clsid)), uintptr(dwDirection), uintptr(unsafe.Pointer(ppenum)))
	return HRESULT(ret)
}

func OleRegEnumVerbs(clsid *syscall.GUID, ppenum **IEnumOLEVERB) HRESULT {
	addr := lazyAddr(&pOleRegEnumVerbs, libOle32, "OleRegEnumVerbs")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(clsid)), uintptr(unsafe.Pointer(ppenum)))
	return HRESULT(ret)
}

func OleDoAutoConvert(pStg *IStorage, pClsidNew *syscall.GUID) HRESULT {
	addr := lazyAddr(&pOleDoAutoConvert, libOle32, "OleDoAutoConvert")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pStg)), uintptr(unsafe.Pointer(pClsidNew)))
	return HRESULT(ret)
}

func OleGetAutoConvert(clsidOld *syscall.GUID, pClsidNew *syscall.GUID) HRESULT {
	addr := lazyAddr(&pOleGetAutoConvert, libOle32, "OleGetAutoConvert")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(clsidOld)), uintptr(unsafe.Pointer(pClsidNew)))
	return HRESULT(ret)
}

func OleSetAutoConvert(clsidOld *syscall.GUID, clsidNew *syscall.GUID) HRESULT {
	addr := lazyAddr(&pOleSetAutoConvert, libOle32, "OleSetAutoConvert")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(clsidOld)), uintptr(unsafe.Pointer(clsidNew)))
	return HRESULT(ret)
}

func HRGN_UserSize(param0 *uint32, param1 uint32, param2 *HRGN) uint32 {
	addr := lazyAddr(&pHRGN_UserSize, libOle32, "HRGN_UserSize")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(param0)), uintptr(param1), uintptr(unsafe.Pointer(param2)))
	return uint32(ret)
}

func HRGN_UserMarshal(param0 *uint32, param1 *byte, param2 *HRGN) *byte {
	addr := lazyAddr(&pHRGN_UserMarshal, libOle32, "HRGN_UserMarshal")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(param0)), uintptr(unsafe.Pointer(param1)), uintptr(unsafe.Pointer(param2)))
	return (*byte)(unsafe.Pointer(ret))
}

func HRGN_UserUnmarshal(param0 *uint32, param1 *byte, param2 *HRGN) *byte {
	addr := lazyAddr(&pHRGN_UserUnmarshal, libOle32, "HRGN_UserUnmarshal")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(param0)), uintptr(unsafe.Pointer(param1)), uintptr(unsafe.Pointer(param2)))
	return (*byte)(unsafe.Pointer(ret))
}

func HRGN_UserFree(param0 *uint32, param1 *HRGN) {
	addr := lazyAddr(&pHRGN_UserFree, libOle32, "HRGN_UserFree")
	syscall.SyscallN(addr, uintptr(unsafe.Pointer(param0)), uintptr(unsafe.Pointer(param1)))
}

func OleCreatePropertyFrame(hwndOwner HWND, x uint32, y uint32, lpszCaption PWSTR, cObjects uint32, ppUnk **IUnknown, cPages uint32, pPageClsID *syscall.GUID, lcid uint32, dwReserved uint32, pvReserved unsafe.Pointer) HRESULT {
	addr := lazyAddr(&pOleCreatePropertyFrame, libOleaut32, "OleCreatePropertyFrame")
	ret, _, _ := syscall.SyscallN(addr, hwndOwner, uintptr(x), uintptr(y), uintptr(unsafe.Pointer(lpszCaption)), uintptr(cObjects), uintptr(unsafe.Pointer(ppUnk)), uintptr(cPages), uintptr(unsafe.Pointer(pPageClsID)), uintptr(lcid), uintptr(dwReserved), uintptr(pvReserved))
	return HRESULT(ret)
}

func OleCreatePropertyFrameIndirect(lpParams *OCPFIPARAMS) HRESULT {
	addr := lazyAddr(&pOleCreatePropertyFrameIndirect, libOleaut32, "OleCreatePropertyFrameIndirect")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpParams)))
	return HRESULT(ret)
}

func OleTranslateColor(clr uint32, hpal HPALETTE, lpcolorref *uint32) HRESULT {
	addr := lazyAddr(&pOleTranslateColor, libOleaut32, "OleTranslateColor")
	ret, _, _ := syscall.SyscallN(addr, uintptr(clr), hpal, uintptr(unsafe.Pointer(lpcolorref)))
	return HRESULT(ret)
}

func OleCreateFontIndirect(lpFontDesc *FONTDESC, riid *syscall.GUID, lplpvObj unsafe.Pointer) HRESULT {
	addr := lazyAddr(&pOleCreateFontIndirect, libOleaut32, "OleCreateFontIndirect")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpFontDesc)), uintptr(unsafe.Pointer(riid)), uintptr(lplpvObj))
	return HRESULT(ret)
}

func OleCreatePictureIndirect(lpPictDesc *PICTDESC, riid *syscall.GUID, fOwn BOOL, lplpvObj unsafe.Pointer) HRESULT {
	addr := lazyAddr(&pOleCreatePictureIndirect, libOleaut32, "OleCreatePictureIndirect")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpPictDesc)), uintptr(unsafe.Pointer(riid)), uintptr(fOwn), uintptr(lplpvObj))
	return HRESULT(ret)
}

func OleLoadPicture(lpstream *IStream, lSize int32, fRunmode BOOL, riid *syscall.GUID, lplpvObj unsafe.Pointer) HRESULT {
	addr := lazyAddr(&pOleLoadPicture, libOleaut32, "OleLoadPicture")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpstream)), uintptr(lSize), uintptr(fRunmode), uintptr(unsafe.Pointer(riid)), uintptr(lplpvObj))
	return HRESULT(ret)
}

func OleLoadPictureEx(lpstream *IStream, lSize int32, fRunmode BOOL, riid *syscall.GUID, xSizeDesired uint32, ySizeDesired uint32, dwFlags uint32, lplpvObj unsafe.Pointer) HRESULT {
	addr := lazyAddr(&pOleLoadPictureEx, libOleaut32, "OleLoadPictureEx")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpstream)), uintptr(lSize), uintptr(fRunmode), uintptr(unsafe.Pointer(riid)), uintptr(xSizeDesired), uintptr(ySizeDesired), uintptr(dwFlags), uintptr(lplpvObj))
	return HRESULT(ret)
}

func OleLoadPicturePath(szURLorPath PWSTR, punkCaller *IUnknown, dwReserved uint32, clrReserved uint32, riid *syscall.GUID, ppvRet unsafe.Pointer) HRESULT {
	addr := lazyAddr(&pOleLoadPicturePath, libOleaut32, "OleLoadPicturePath")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(szURLorPath)), uintptr(unsafe.Pointer(punkCaller)), uintptr(dwReserved), uintptr(clrReserved), uintptr(unsafe.Pointer(riid)), uintptr(ppvRet))
	return HRESULT(ret)
}

func OleLoadPictureFile(varFileName VARIANT, lplpdispPicture **IDispatch) HRESULT {
	addr := lazyAddr(&pOleLoadPictureFile, libOleaut32, "OleLoadPictureFile")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(&varFileName)), uintptr(unsafe.Pointer(lplpdispPicture)))
	return HRESULT(ret)
}

func OleLoadPictureFileEx(varFileName VARIANT, xSizeDesired uint32, ySizeDesired uint32, dwFlags uint32, lplpdispPicture **IDispatch) HRESULT {
	addr := lazyAddr(&pOleLoadPictureFileEx, libOleaut32, "OleLoadPictureFileEx")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(&varFileName)), uintptr(xSizeDesired), uintptr(ySizeDesired), uintptr(dwFlags), uintptr(unsafe.Pointer(lplpdispPicture)))
	return HRESULT(ret)
}

func OleSavePictureFile(lpdispPicture *IDispatch, bstrFileName BSTR) HRESULT {
	addr := lazyAddr(&pOleSavePictureFile, libOleaut32, "OleSavePictureFile")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpdispPicture)), uintptr(unsafe.Pointer(bstrFileName)))
	return HRESULT(ret)
}

func OleIconToCursor(hinstExe HINSTANCE, hIcon HICON) HCURSOR {
	addr := lazyAddr(&pOleIconToCursor, libOleaut32, "OleIconToCursor")
	ret, _, _ := syscall.SyscallN(addr, hinstExe, hIcon)
	return ret
}
