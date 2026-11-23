package win32

import (
	"syscall"
	"unsafe"
)

type (
	HSZ       = uintptr
	HCONV     = uintptr
	HCONVLIST = uintptr
	HDDEDATA  = uintptr
)

const (
	WM_DDE_FIRST               uint32 = 0x3e0
	WM_DDE_INITIATE            uint32 = 0x3e0
	WM_DDE_TERMINATE           uint32 = 0x3e1
	WM_DDE_ADVISE              uint32 = 0x3e2
	WM_DDE_UNADVISE            uint32 = 0x3e3
	WM_DDE_ACK                 uint32 = 0x3e4
	WM_DDE_DATA                uint32 = 0x3e5
	WM_DDE_REQUEST             uint32 = 0x3e6
	WM_DDE_POKE                uint32 = 0x3e7
	WM_DDE_EXECUTE             uint32 = 0x3e8
	WM_DDE_LAST                uint32 = 0x3e8
	CADV_LATEACK               uint32 = 0xffff
	DDE_FACK                   uint32 = 0x8000
	DDE_FBUSY                  uint32 = 0x4000
	DDE_FDEFERUPD              uint32 = 0x4000
	DDE_FACKREQ                uint32 = 0x8000
	DDE_FRELEASE               uint32 = 0x2000
	DDE_FREQUESTED             uint32 = 0x1000
	DDE_FAPPSTATUS             uint32 = 0xff
	DDE_FNOTPROCESSED          uint32 = 0x0
	MSGF_DDEMGR                uint32 = 0x8001
	CP_WINANSI                 int32  = 1004
	CP_WINUNICODE              int32  = 1200
	CP_WINNEUTRAL              int32  = 1200
	XTYPF_NOBLOCK              uint32 = 0x2
	XTYPF_NODATA               uint32 = 0x4
	XTYPF_ACKREQ               uint32 = 0x8
	XCLASS_MASK                uint32 = 0xfc00
	XCLASS_BOOL                uint32 = 0x1000
	XCLASS_DATA                uint32 = 0x2000
	XCLASS_FLAGS               uint32 = 0x4000
	XCLASS_NOTIFICATION        uint32 = 0x8000
	XTYP_MASK                  uint32 = 0xf0
	XTYP_SHIFT                 uint32 = 0x4
	TIMEOUT_ASYNC              uint32 = 0xffffffff
	QID_SYNC                   uint32 = 0xffffffff
	SZDDESYS_TOPIC             string = "System"
	SZDDESYS_ITEM_TOPICS       string = "Topics"
	SZDDESYS_ITEM_SYSITEMS     string = "SysItems"
	SZDDESYS_ITEM_RTNMSG       string = "ReturnMessage"
	SZDDESYS_ITEM_STATUS       string = "Status"
	SZDDESYS_ITEM_FORMATS      string = "Formats"
	SZDDESYS_ITEM_HELP         string = "Help"
	SZDDE_ITEM_ITEMLIST        string = "TopicItemList"
	APPCMD_MASK                int32  = 4080
	APPCLASS_MASK              int32  = 15
	HDATA_APPOWNED             uint32 = 0x1
	DMLERR_NO_ERROR            uint32 = 0x0
	DMLERR_FIRST               uint32 = 0x4000
	DMLERR_ADVACKTIMEOUT       uint32 = 0x4000
	DMLERR_BUSY                uint32 = 0x4001
	DMLERR_DATAACKTIMEOUT      uint32 = 0x4002
	DMLERR_DLL_NOT_INITIALIZED uint32 = 0x4003
	DMLERR_DLL_USAGE           uint32 = 0x4004
	DMLERR_EXECACKTIMEOUT      uint32 = 0x4005
	DMLERR_INVALIDPARAMETER    uint32 = 0x4006
	DMLERR_LOW_MEMORY          uint32 = 0x4007
	DMLERR_MEMORY_ERROR        uint32 = 0x4008
	DMLERR_NOTPROCESSED        uint32 = 0x4009
	DMLERR_NO_CONV_ESTABLISHED uint32 = 0x400a
	DMLERR_POKEACKTIMEOUT      uint32 = 0x400b
	DMLERR_POSTMSG_FAILED      uint32 = 0x400c
	DMLERR_REENTRANCY          uint32 = 0x400d
	DMLERR_SERVER_DIED         uint32 = 0x400e
	DMLERR_SYS_ERROR           uint32 = 0x400f
	DMLERR_UNADVACKTIMEOUT     uint32 = 0x4010
	DMLERR_UNFOUND_QUEUE_ID    uint32 = 0x4011
	DMLERR_LAST                uint32 = 0x4011
	MH_CREATE                  uint32 = 0x1
	MH_KEEP                    uint32 = 0x2
	MH_DELETE                  uint32 = 0x3
	MH_CLEANUP                 uint32 = 0x4
	MAX_MONITORS               uint32 = 0x4
	MF_MASK                    uint32 = 0xff000000
)

// enums

// enum
type DDE_ENABLE_CALLBACK_CMD uint32

const (
	EC_ENABLEALL    DDE_ENABLE_CALLBACK_CMD = 0
	EC_ENABLEONE    DDE_ENABLE_CALLBACK_CMD = 128
	EC_DISABLE      DDE_ENABLE_CALLBACK_CMD = 8
	EC_QUERYWAITING DDE_ENABLE_CALLBACK_CMD = 2
)

// enum
// flags
type DDE_INITIALIZE_COMMAND uint32

const (
	APPCLASS_MONITOR          DDE_INITIALIZE_COMMAND = 1
	APPCLASS_STANDARD         DDE_INITIALIZE_COMMAND = 0
	APPCMD_CLIENTONLY         DDE_INITIALIZE_COMMAND = 16
	APPCMD_FILTERINITS        DDE_INITIALIZE_COMMAND = 32
	CBF_FAIL_ALLSVRXACTIONS   DDE_INITIALIZE_COMMAND = 258048
	CBF_FAIL_ADVISES          DDE_INITIALIZE_COMMAND = 16384
	CBF_FAIL_CONNECTIONS      DDE_INITIALIZE_COMMAND = 8192
	CBF_FAIL_EXECUTES         DDE_INITIALIZE_COMMAND = 32768
	CBF_FAIL_POKES            DDE_INITIALIZE_COMMAND = 65536
	CBF_FAIL_REQUESTS         DDE_INITIALIZE_COMMAND = 131072
	CBF_FAIL_SELFCONNECTIONS  DDE_INITIALIZE_COMMAND = 4096
	CBF_SKIP_ALLNOTIFICATIONS DDE_INITIALIZE_COMMAND = 3932160
	CBF_SKIP_CONNECT_CONFIRMS DDE_INITIALIZE_COMMAND = 262144
	CBF_SKIP_DISCONNECTS      DDE_INITIALIZE_COMMAND = 2097152
	CBF_SKIP_REGISTRATIONS    DDE_INITIALIZE_COMMAND = 524288
	CBF_SKIP_UNREGISTRATIONS  DDE_INITIALIZE_COMMAND = 1048576
	MF_CALLBACKS              DDE_INITIALIZE_COMMAND = 134217728
	MF_CONV                   DDE_INITIALIZE_COMMAND = 1073741824
	MF_ERRORS                 DDE_INITIALIZE_COMMAND = 268435456
	MF_HSZ_INFO               DDE_INITIALIZE_COMMAND = 16777216
	MF_LINKS                  DDE_INITIALIZE_COMMAND = 536870912
	MF_POSTMSGS               DDE_INITIALIZE_COMMAND = 67108864
	MF_SENDMSGS               DDE_INITIALIZE_COMMAND = 33554432
)

// enum
type DDE_NAME_SERVICE_CMD uint32

const (
	DNS_REGISTER   DDE_NAME_SERVICE_CMD = 1
	DNS_UNREGISTER DDE_NAME_SERVICE_CMD = 2
	DNS_FILTERON   DDE_NAME_SERVICE_CMD = 4
	DNS_FILTEROFF  DDE_NAME_SERVICE_CMD = 8
)

// enum
type DDE_CLIENT_TRANSACTION_TYPE uint32

const (
	XTYP_ADVSTART        DDE_CLIENT_TRANSACTION_TYPE = 4144
	XTYP_ADVSTOP         DDE_CLIENT_TRANSACTION_TYPE = 32832
	XTYP_EXECUTE         DDE_CLIENT_TRANSACTION_TYPE = 16464
	XTYP_POKE            DDE_CLIENT_TRANSACTION_TYPE = 16528
	XTYP_REQUEST         DDE_CLIENT_TRANSACTION_TYPE = 8368
	XTYP_ADVDATA         DDE_CLIENT_TRANSACTION_TYPE = 16400
	XTYP_ADVREQ          DDE_CLIENT_TRANSACTION_TYPE = 8226
	XTYP_CONNECT         DDE_CLIENT_TRANSACTION_TYPE = 4194
	XTYP_CONNECT_CONFIRM DDE_CLIENT_TRANSACTION_TYPE = 32882
	XTYP_DISCONNECT      DDE_CLIENT_TRANSACTION_TYPE = 32962
	XTYP_MONITOR         DDE_CLIENT_TRANSACTION_TYPE = 33010
	XTYP_REGISTER        DDE_CLIENT_TRANSACTION_TYPE = 32930
	XTYP_UNREGISTER      DDE_CLIENT_TRANSACTION_TYPE = 32978
	XTYP_WILDCONNECT     DDE_CLIENT_TRANSACTION_TYPE = 8418
	XTYP_XACT_COMPLETE   DDE_CLIENT_TRANSACTION_TYPE = 32896
)

// enum
type CONVINFO_CONVERSATION_STATE uint32

const (
	XST_ADVACKRCVD     CONVINFO_CONVERSATION_STATE = 13
	XST_ADVDATAACKRCVD CONVINFO_CONVERSATION_STATE = 16
	XST_ADVDATASENT    CONVINFO_CONVERSATION_STATE = 15
	XST_ADVSENT        CONVINFO_CONVERSATION_STATE = 11
	XST_CONNECTED      CONVINFO_CONVERSATION_STATE = 2
	XST_DATARCVD       CONVINFO_CONVERSATION_STATE = 6
	XST_EXECACKRCVD    CONVINFO_CONVERSATION_STATE = 10
	XST_EXECSENT       CONVINFO_CONVERSATION_STATE = 9
	XST_INCOMPLETE     CONVINFO_CONVERSATION_STATE = 1
	XST_INIT1          CONVINFO_CONVERSATION_STATE = 3
	XST_INIT2          CONVINFO_CONVERSATION_STATE = 4
	XST_NULL           CONVINFO_CONVERSATION_STATE = 0
	XST_POKEACKRCVD    CONVINFO_CONVERSATION_STATE = 8
	XST_POKESENT       CONVINFO_CONVERSATION_STATE = 7
	XST_REQSENT        CONVINFO_CONVERSATION_STATE = 5
	XST_UNADVACKRCVD   CONVINFO_CONVERSATION_STATE = 14
	XST_UNADVSENT      CONVINFO_CONVERSATION_STATE = 12
)

// enum
// flags
type CONVINFO_STATUS uint32

const (
	ST_ADVISE     CONVINFO_STATUS = 2
	ST_BLOCKED    CONVINFO_STATUS = 8
	ST_BLOCKNEXT  CONVINFO_STATUS = 128
	ST_CLIENT     CONVINFO_STATUS = 16
	ST_CONNECTED  CONVINFO_STATUS = 1
	ST_INLIST     CONVINFO_STATUS = 64
	ST_ISLOCAL    CONVINFO_STATUS = 4
	ST_ISSELF     CONVINFO_STATUS = 256
	ST_TERMINATED CONVINFO_STATUS = 32
)

// structs

type DDEACK struct {
	Bitfield_ uint16
}

type DDEADVISE struct {
	Bitfield_ uint16
	CfFormat  int16
}

type DDEDATA struct {
	Bitfield_ uint16
	CfFormat  int16
	Value     [1]byte
}

type DDEPOKE struct {
	Bitfield_ uint16
	CfFormat  int16
	Value     [1]byte
}

type DDELN struct {
	Bitfield_ uint16
	CfFormat  int16
}

type DDEUP struct {
	Bitfield_ uint16
	CfFormat  int16
	Rgb       [1]byte
}

type HSZPAIR struct {
	HszSvc   HSZ
	HszTopic HSZ
}

type CONVCONTEXT struct {
	Cb         uint32
	WFlags     uint32
	WCountryID uint32
	ICodePage  int32
	DwLangID   uint32
	DwSecurity uint32
	Qos        SECURITY_QUALITY_OF_SERVICE
}

type CONVINFO struct {
	Cb            uint32
	HUser         uintptr
	HConvPartner  HCONV
	HszSvcPartner HSZ
	HszServiceReq HSZ
	HszTopic      HSZ
	HszItem       HSZ
	WFmt          uint32
	WType         DDE_CLIENT_TRANSACTION_TYPE
	WStatus       CONVINFO_STATUS
	WConvst       CONVINFO_CONVERSATION_STATE
	WLastError    uint32
	HConvList     HCONVLIST
	ConvCtxt      CONVCONTEXT
	Hwnd          HWND
	HwndPartner   HWND
}

type DDEML_MSG_HOOK_DATA struct {
	UiLo   uintptr
	UiHi   uintptr
	CbData uint32
	Data   [8]uint32
}

type MONMSGSTRUCT struct {
	Cb     uint32
	HwndTo HWND
	DwTime uint32
	HTask  HANDLE
	WMsg   uint32
	WParam WPARAM
	LParam LPARAM
	Dmhd   DDEML_MSG_HOOK_DATA
}

type MONCBSTRUCT struct {
	Cb      uint32
	DwTime  uint32
	HTask   HANDLE
	DwRet   uint32
	WType   uint32
	WFmt    uint32
	HConv   HCONV
	Hsz1    HSZ
	Hsz2    HSZ
	HData   HDDEDATA
	DwData1 uintptr
	DwData2 uintptr
	Cc      CONVCONTEXT
	CbData  uint32
	Data    [8]uint32
}

type MONHSZSTRUCTA struct {
	Cb       uint32
	FsAction BOOL
	DwTime   uint32
	Hsz      HSZ
	HTask    HANDLE
	Str      [1]CHAR
}

type MONHSZSTRUCT = MONHSZSTRUCTW
type MONHSZSTRUCTW struct {
	Cb       uint32
	FsAction BOOL
	DwTime   uint32
	Hsz      HSZ
	HTask    HANDLE
	Str      [1]uint16
}

type MONERRSTRUCT struct {
	Cb         uint32
	WLastError uint32
	DwTime     uint32
	HTask      HANDLE
}

type MONLINKSTRUCT struct {
	Cb           uint32
	DwTime       uint32
	HTask        HANDLE
	FEstablished BOOL
	FNoData      BOOL
	HszSvc       HSZ
	HszTopic     HSZ
	HszItem      HSZ
	WFmt         uint32
	FServer      BOOL
	HConvServer  HCONV
	HConvClient  HCONV
}

type MONCONVSTRUCT struct {
	Cb          uint32
	FConnect    BOOL
	DwTime      uint32
	HTask       HANDLE
	HszSvc      HSZ
	HszTopic    HSZ
	HConvClient HCONV
	HConvServer HCONV
}

type METAFILEPICT struct {
	Mm   int32
	XExt int32
	YExt int32
	HMF  HMETAFILE
}

type COPYDATASTRUCT struct {
	DwData uintptr
	CbData uint32
	LpData unsafe.Pointer
}

// func types

type PFNCALLBACK = uintptr
type PFNCALLBACK_func = func(wType uint32, wFmt uint32, hConv HCONV, hsz1 HSZ, hsz2 HSZ, hData HDDEDATA, dwData1 uintptr, dwData2 uintptr) HDDEDATA

var (
	pDdeSetQualityOfService        uintptr
	pImpersonateDdeClientWindow    uintptr
	pPackDDElParam                 uintptr
	pUnpackDDElParam               uintptr
	pFreeDDElParam                 uintptr
	pReuseDDElParam                uintptr
	pDdeInitializeA                uintptr
	pDdeInitializeW                uintptr
	pDdeUninitialize               uintptr
	pDdeConnectList                uintptr
	pDdeQueryNextServer            uintptr
	pDdeDisconnectList             uintptr
	pDdeConnect                    uintptr
	pDdeDisconnect                 uintptr
	pDdeReconnect                  uintptr
	pDdeQueryConvInfo              uintptr
	pDdeSetUserHandle              uintptr
	pDdeAbandonTransaction         uintptr
	pDdePostAdvise                 uintptr
	pDdeEnableCallback             uintptr
	pDdeImpersonateClient          uintptr
	pDdeNameService                uintptr
	pDdeClientTransaction          uintptr
	pDdeCreateDataHandle           uintptr
	pDdeAddData                    uintptr
	pDdeGetData                    uintptr
	pDdeAccessData                 uintptr
	pDdeUnaccessData               uintptr
	pDdeFreeDataHandle             uintptr
	pDdeGetLastError               uintptr
	pDdeCreateStringHandleA        uintptr
	pDdeCreateStringHandleW        uintptr
	pDdeQueryStringA               uintptr
	pDdeQueryStringW               uintptr
	pDdeFreeStringHandle           uintptr
	pDdeKeepStringHandle           uintptr
	pDdeCmpStringHandles           uintptr
	pSetWinMetaFileBits            uintptr
	pOpenClipboard                 uintptr
	pCloseClipboard                uintptr
	pGetClipboardSequenceNumber    uintptr
	pGetClipboardOwner             uintptr
	pSetClipboardViewer            uintptr
	pGetClipboardViewer            uintptr
	pChangeClipboardChain          uintptr
	pSetClipboardData              uintptr
	pGetClipboardData              uintptr
	pRegisterClipboardFormatA      uintptr
	pRegisterClipboardFormatW      uintptr
	pCountClipboardFormats         uintptr
	pEnumClipboardFormats          uintptr
	pGetClipboardFormatNameA       uintptr
	pGetClipboardFormatNameW       uintptr
	pEmptyClipboard                uintptr
	pIsClipboardFormatAvailable    uintptr
	pGetPriorityClipboardFormat    uintptr
	pGetOpenClipboardWindow        uintptr
	pAddClipboardFormatListener    uintptr
	pRemoveClipboardFormatListener uintptr
	pGetUpdatedClipboardFormats    uintptr
	pGlobalDeleteAtom              uintptr
	pInitAtomTable                 uintptr
	pDeleteAtom                    uintptr
	pGlobalAddAtomA                uintptr
	pGlobalAddAtomW                uintptr
	pGlobalAddAtomExA              uintptr
	pGlobalAddAtomExW              uintptr
	pGlobalFindAtomA               uintptr
	pGlobalFindAtomW               uintptr
	pGlobalGetAtomNameA            uintptr
	pGlobalGetAtomNameW            uintptr
	pAddAtomA                      uintptr
	pAddAtomW                      uintptr
	pFindAtomA                     uintptr
	pFindAtomW                     uintptr
	pGetAtomNameA                  uintptr
	pGetAtomNameW                  uintptr
)

func DdeSetQualityOfService(hwndClient HWND, pqosNew *SECURITY_QUALITY_OF_SERVICE, pqosPrev *SECURITY_QUALITY_OF_SERVICE) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pDdeSetQualityOfService, libUser32, "DdeSetQualityOfService")
	ret, _, err := syscall.SyscallN(addr, hwndClient, uintptr(unsafe.Pointer(pqosNew)), uintptr(unsafe.Pointer(pqosPrev)))
	return BOOL(ret), WIN32_ERROR(err)
}

func ImpersonateDdeClientWindow(hWndClient HWND, hWndServer HWND) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pImpersonateDdeClientWindow, libUser32, "ImpersonateDdeClientWindow")
	ret, _, err := syscall.SyscallN(addr, hWndClient, hWndServer)
	return BOOL(ret), WIN32_ERROR(err)
}

func PackDDElParam(msg uint32, uiLo uintptr, uiHi uintptr) LPARAM {
	addr := lazyAddr(&pPackDDElParam, libUser32, "PackDDElParam")
	ret, _, _ := syscall.SyscallN(addr, uintptr(msg), uiLo, uiHi)
	return ret
}

func UnpackDDElParam(msg uint32, lParam LPARAM, puiLo *uintptr, puiHi *uintptr) BOOL {
	addr := lazyAddr(&pUnpackDDElParam, libUser32, "UnpackDDElParam")
	ret, _, _ := syscall.SyscallN(addr, uintptr(msg), lParam, uintptr(unsafe.Pointer(puiLo)), uintptr(unsafe.Pointer(puiHi)))
	return BOOL(ret)
}

func FreeDDElParam(msg uint32, lParam LPARAM) BOOL {
	addr := lazyAddr(&pFreeDDElParam, libUser32, "FreeDDElParam")
	ret, _, _ := syscall.SyscallN(addr, uintptr(msg), lParam)
	return BOOL(ret)
}

func ReuseDDElParam(lParam LPARAM, msgIn uint32, msgOut uint32, uiLo uintptr, uiHi uintptr) LPARAM {
	addr := lazyAddr(&pReuseDDElParam, libUser32, "ReuseDDElParam")
	ret, _, _ := syscall.SyscallN(addr, lParam, uintptr(msgIn), uintptr(msgOut), uiLo, uiHi)
	return ret
}

func DdeInitializeA(pidInst *uint32, pfnCallback PFNCALLBACK, afCmd DDE_INITIALIZE_COMMAND, ulRes uint32) uint32 {
	addr := lazyAddr(&pDdeInitializeA, libUser32, "DdeInitializeA")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pidInst)), pfnCallback, uintptr(afCmd), uintptr(ulRes))
	return uint32(ret)
}

var DdeInitialize = DdeInitializeW

func DdeInitializeW(pidInst *uint32, pfnCallback PFNCALLBACK, afCmd DDE_INITIALIZE_COMMAND, ulRes uint32) uint32 {
	addr := lazyAddr(&pDdeInitializeW, libUser32, "DdeInitializeW")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pidInst)), pfnCallback, uintptr(afCmd), uintptr(ulRes))
	return uint32(ret)
}

func DdeUninitialize(idInst uint32) BOOL {
	addr := lazyAddr(&pDdeUninitialize, libUser32, "DdeUninitialize")
	ret, _, _ := syscall.SyscallN(addr, uintptr(idInst))
	return BOOL(ret)
}

func DdeConnectList(idInst uint32, hszService HSZ, hszTopic HSZ, hConvList HCONVLIST, pCC *CONVCONTEXT) HCONVLIST {
	addr := lazyAddr(&pDdeConnectList, libUser32, "DdeConnectList")
	ret, _, _ := syscall.SyscallN(addr, uintptr(idInst), hszService, hszTopic, hConvList, uintptr(unsafe.Pointer(pCC)))
	return ret
}

func DdeQueryNextServer(hConvList HCONVLIST, hConvPrev HCONV) HCONV {
	addr := lazyAddr(&pDdeQueryNextServer, libUser32, "DdeQueryNextServer")
	ret, _, _ := syscall.SyscallN(addr, hConvList, hConvPrev)
	return ret
}

func DdeDisconnectList(hConvList HCONVLIST) BOOL {
	addr := lazyAddr(&pDdeDisconnectList, libUser32, "DdeDisconnectList")
	ret, _, _ := syscall.SyscallN(addr, hConvList)
	return BOOL(ret)
}

func DdeConnect(idInst uint32, hszService HSZ, hszTopic HSZ, pCC *CONVCONTEXT) HCONV {
	addr := lazyAddr(&pDdeConnect, libUser32, "DdeConnect")
	ret, _, _ := syscall.SyscallN(addr, uintptr(idInst), hszService, hszTopic, uintptr(unsafe.Pointer(pCC)))
	return ret
}

func DdeDisconnect(hConv HCONV) BOOL {
	addr := lazyAddr(&pDdeDisconnect, libUser32, "DdeDisconnect")
	ret, _, _ := syscall.SyscallN(addr, hConv)
	return BOOL(ret)
}

func DdeReconnect(hConv HCONV) HCONV {
	addr := lazyAddr(&pDdeReconnect, libUser32, "DdeReconnect")
	ret, _, _ := syscall.SyscallN(addr, hConv)
	return ret
}

func DdeQueryConvInfo(hConv HCONV, idTransaction uint32, pConvInfo *CONVINFO) uint32 {
	addr := lazyAddr(&pDdeQueryConvInfo, libUser32, "DdeQueryConvInfo")
	ret, _, _ := syscall.SyscallN(addr, hConv, uintptr(idTransaction), uintptr(unsafe.Pointer(pConvInfo)))
	return uint32(ret)
}

func DdeSetUserHandle(hConv HCONV, id uint32, hUser uintptr) BOOL {
	addr := lazyAddr(&pDdeSetUserHandle, libUser32, "DdeSetUserHandle")
	ret, _, _ := syscall.SyscallN(addr, hConv, uintptr(id), hUser)
	return BOOL(ret)
}

func DdeAbandonTransaction(idInst uint32, hConv HCONV, idTransaction uint32) BOOL {
	addr := lazyAddr(&pDdeAbandonTransaction, libUser32, "DdeAbandonTransaction")
	ret, _, _ := syscall.SyscallN(addr, uintptr(idInst), hConv, uintptr(idTransaction))
	return BOOL(ret)
}

func DdePostAdvise(idInst uint32, hszTopic HSZ, hszItem HSZ) BOOL {
	addr := lazyAddr(&pDdePostAdvise, libUser32, "DdePostAdvise")
	ret, _, _ := syscall.SyscallN(addr, uintptr(idInst), hszTopic, hszItem)
	return BOOL(ret)
}

func DdeEnableCallback(idInst uint32, hConv HCONV, wCmd DDE_ENABLE_CALLBACK_CMD) BOOL {
	addr := lazyAddr(&pDdeEnableCallback, libUser32, "DdeEnableCallback")
	ret, _, _ := syscall.SyscallN(addr, uintptr(idInst), hConv, uintptr(wCmd))
	return BOOL(ret)
}

func DdeImpersonateClient(hConv HCONV) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pDdeImpersonateClient, libUser32, "DdeImpersonateClient")
	ret, _, err := syscall.SyscallN(addr, hConv)
	return BOOL(ret), WIN32_ERROR(err)
}

func DdeNameService(idInst uint32, hsz1 HSZ, hsz2 HSZ, afCmd DDE_NAME_SERVICE_CMD) HDDEDATA {
	addr := lazyAddr(&pDdeNameService, libUser32, "DdeNameService")
	ret, _, _ := syscall.SyscallN(addr, uintptr(idInst), hsz1, hsz2, uintptr(afCmd))
	return ret
}

func DdeClientTransaction(pData *byte, cbData uint32, hConv HCONV, hszItem HSZ, wFmt uint32, wType DDE_CLIENT_TRANSACTION_TYPE, dwTimeout uint32, pdwResult *uint32) HDDEDATA {
	addr := lazyAddr(&pDdeClientTransaction, libUser32, "DdeClientTransaction")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(pData)), uintptr(cbData), hConv, hszItem, uintptr(wFmt), uintptr(wType), uintptr(dwTimeout), uintptr(unsafe.Pointer(pdwResult)))
	return ret
}

func DdeCreateDataHandle(idInst uint32, pSrc *byte, cb uint32, cbOff uint32, hszItem HSZ, wFmt uint32, afCmd uint32) HDDEDATA {
	addr := lazyAddr(&pDdeCreateDataHandle, libUser32, "DdeCreateDataHandle")
	ret, _, _ := syscall.SyscallN(addr, uintptr(idInst), uintptr(unsafe.Pointer(pSrc)), uintptr(cb), uintptr(cbOff), hszItem, uintptr(wFmt), uintptr(afCmd))
	return ret
}

func DdeAddData(hData HDDEDATA, pSrc *byte, cb uint32, cbOff uint32) HDDEDATA {
	addr := lazyAddr(&pDdeAddData, libUser32, "DdeAddData")
	ret, _, _ := syscall.SyscallN(addr, hData, uintptr(unsafe.Pointer(pSrc)), uintptr(cb), uintptr(cbOff))
	return ret
}

func DdeGetData(hData HDDEDATA, pDst *byte, cbMax uint32, cbOff uint32) uint32 {
	addr := lazyAddr(&pDdeGetData, libUser32, "DdeGetData")
	ret, _, _ := syscall.SyscallN(addr, hData, uintptr(unsafe.Pointer(pDst)), uintptr(cbMax), uintptr(cbOff))
	return uint32(ret)
}

func DdeAccessData(hData HDDEDATA, pcbDataSize *uint32) *byte {
	addr := lazyAddr(&pDdeAccessData, libUser32, "DdeAccessData")
	ret, _, _ := syscall.SyscallN(addr, hData, uintptr(unsafe.Pointer(pcbDataSize)))
	return (*byte)(unsafe.Pointer(ret))
}

func DdeUnaccessData(hData HDDEDATA) BOOL {
	addr := lazyAddr(&pDdeUnaccessData, libUser32, "DdeUnaccessData")
	ret, _, _ := syscall.SyscallN(addr, hData)
	return BOOL(ret)
}

func DdeFreeDataHandle(hData HDDEDATA) BOOL {
	addr := lazyAddr(&pDdeFreeDataHandle, libUser32, "DdeFreeDataHandle")
	ret, _, _ := syscall.SyscallN(addr, hData)
	return BOOL(ret)
}

func DdeGetLastError(idInst uint32) uint32 {
	addr := lazyAddr(&pDdeGetLastError, libUser32, "DdeGetLastError")
	ret, _, _ := syscall.SyscallN(addr, uintptr(idInst))
	return uint32(ret)
}

func DdeCreateStringHandleA(idInst uint32, psz PSTR, iCodePage int32) HSZ {
	addr := lazyAddr(&pDdeCreateStringHandleA, libUser32, "DdeCreateStringHandleA")
	ret, _, _ := syscall.SyscallN(addr, uintptr(idInst), uintptr(unsafe.Pointer(psz)), uintptr(iCodePage))
	return ret
}

var DdeCreateStringHandle = DdeCreateStringHandleW

func DdeCreateStringHandleW(idInst uint32, psz PWSTR, iCodePage int32) HSZ {
	addr := lazyAddr(&pDdeCreateStringHandleW, libUser32, "DdeCreateStringHandleW")
	ret, _, _ := syscall.SyscallN(addr, uintptr(idInst), uintptr(unsafe.Pointer(psz)), uintptr(iCodePage))
	return ret
}

func DdeQueryStringA(idInst uint32, hsz HSZ, psz PSTR, cchMax uint32, iCodePage int32) uint32 {
	addr := lazyAddr(&pDdeQueryStringA, libUser32, "DdeQueryStringA")
	ret, _, _ := syscall.SyscallN(addr, uintptr(idInst), hsz, uintptr(unsafe.Pointer(psz)), uintptr(cchMax), uintptr(iCodePage))
	return uint32(ret)
}

var DdeQueryString = DdeQueryStringW

func DdeQueryStringW(idInst uint32, hsz HSZ, psz PWSTR, cchMax uint32, iCodePage int32) uint32 {
	addr := lazyAddr(&pDdeQueryStringW, libUser32, "DdeQueryStringW")
	ret, _, _ := syscall.SyscallN(addr, uintptr(idInst), hsz, uintptr(unsafe.Pointer(psz)), uintptr(cchMax), uintptr(iCodePage))
	return uint32(ret)
}

func DdeFreeStringHandle(idInst uint32, hsz HSZ) BOOL {
	addr := lazyAddr(&pDdeFreeStringHandle, libUser32, "DdeFreeStringHandle")
	ret, _, _ := syscall.SyscallN(addr, uintptr(idInst), hsz)
	return BOOL(ret)
}

func DdeKeepStringHandle(idInst uint32, hsz HSZ) BOOL {
	addr := lazyAddr(&pDdeKeepStringHandle, libUser32, "DdeKeepStringHandle")
	ret, _, _ := syscall.SyscallN(addr, uintptr(idInst), hsz)
	return BOOL(ret)
}

func DdeCmpStringHandles(hsz1 HSZ, hsz2 HSZ) int32 {
	addr := lazyAddr(&pDdeCmpStringHandles, libUser32, "DdeCmpStringHandles")
	ret, _, _ := syscall.SyscallN(addr, hsz1, hsz2)
	return int32(ret)
}

func SetWinMetaFileBits(nSize uint32, lpMeta16Data *byte, hdcRef HDC, lpMFP *METAFILEPICT) HENHMETAFILE {
	addr := lazyAddr(&pSetWinMetaFileBits, libGdi32, "SetWinMetaFileBits")
	ret, _, _ := syscall.SyscallN(addr, uintptr(nSize), uintptr(unsafe.Pointer(lpMeta16Data)), hdcRef, uintptr(unsafe.Pointer(lpMFP)))
	return ret
}

func OpenClipboard(hWndNewOwner HWND) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pOpenClipboard, libUser32, "OpenClipboard")
	ret, _, err := syscall.SyscallN(addr, hWndNewOwner)
	return BOOL(ret), WIN32_ERROR(err)
}

func CloseClipboard() (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pCloseClipboard, libUser32, "CloseClipboard")
	ret, _, err := syscall.SyscallN(addr)
	return BOOL(ret), WIN32_ERROR(err)
}

func GetClipboardSequenceNumber() uint32 {
	addr := lazyAddr(&pGetClipboardSequenceNumber, libUser32, "GetClipboardSequenceNumber")
	ret, _, _ := syscall.SyscallN(addr)
	return uint32(ret)
}

func GetClipboardOwner() (HWND, WIN32_ERROR) {
	addr := lazyAddr(&pGetClipboardOwner, libUser32, "GetClipboardOwner")
	ret, _, err := syscall.SyscallN(addr)
	return ret, WIN32_ERROR(err)
}

func SetClipboardViewer(hWndNewViewer HWND) (HWND, WIN32_ERROR) {
	addr := lazyAddr(&pSetClipboardViewer, libUser32, "SetClipboardViewer")
	ret, _, err := syscall.SyscallN(addr, hWndNewViewer)
	return ret, WIN32_ERROR(err)
}

func GetClipboardViewer() (HWND, WIN32_ERROR) {
	addr := lazyAddr(&pGetClipboardViewer, libUser32, "GetClipboardViewer")
	ret, _, err := syscall.SyscallN(addr)
	return ret, WIN32_ERROR(err)
}

func ChangeClipboardChain(hWndRemove HWND, hWndNewNext HWND) BOOL {
	addr := lazyAddr(&pChangeClipboardChain, libUser32, "ChangeClipboardChain")
	ret, _, _ := syscall.SyscallN(addr, hWndRemove, hWndNewNext)
	return BOOL(ret)
}

func SetClipboardData(uFormat uint32, hMem HANDLE) (HANDLE, WIN32_ERROR) {
	addr := lazyAddr(&pSetClipboardData, libUser32, "SetClipboardData")
	ret, _, err := syscall.SyscallN(addr, uintptr(uFormat), hMem)
	return ret, WIN32_ERROR(err)
}

func GetClipboardData(uFormat uint32) (HANDLE, WIN32_ERROR) {
	addr := lazyAddr(&pGetClipboardData, libUser32, "GetClipboardData")
	ret, _, err := syscall.SyscallN(addr, uintptr(uFormat))
	return ret, WIN32_ERROR(err)
}

func RegisterClipboardFormatA(lpszFormat PSTR) (uint32, WIN32_ERROR) {
	addr := lazyAddr(&pRegisterClipboardFormatA, libUser32, "RegisterClipboardFormatA")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpszFormat)))
	return uint32(ret), WIN32_ERROR(err)
}

var RegisterClipboardFormat = RegisterClipboardFormatW

func RegisterClipboardFormatW(lpszFormat PWSTR) (uint32, WIN32_ERROR) {
	addr := lazyAddr(&pRegisterClipboardFormatW, libUser32, "RegisterClipboardFormatW")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpszFormat)))
	return uint32(ret), WIN32_ERROR(err)
}

func CountClipboardFormats() (int32, WIN32_ERROR) {
	addr := lazyAddr(&pCountClipboardFormats, libUser32, "CountClipboardFormats")
	ret, _, err := syscall.SyscallN(addr)
	return int32(ret), WIN32_ERROR(err)
}

func EnumClipboardFormats(format uint32) (uint32, WIN32_ERROR) {
	addr := lazyAddr(&pEnumClipboardFormats, libUser32, "EnumClipboardFormats")
	ret, _, err := syscall.SyscallN(addr, uintptr(format))
	return uint32(ret), WIN32_ERROR(err)
}

func GetClipboardFormatNameA(format uint32, lpszFormatName PSTR, cchMaxCount int32) (int32, WIN32_ERROR) {
	addr := lazyAddr(&pGetClipboardFormatNameA, libUser32, "GetClipboardFormatNameA")
	ret, _, err := syscall.SyscallN(addr, uintptr(format), uintptr(unsafe.Pointer(lpszFormatName)), uintptr(cchMaxCount))
	return int32(ret), WIN32_ERROR(err)
}

var GetClipboardFormatName = GetClipboardFormatNameW

func GetClipboardFormatNameW(format uint32, lpszFormatName PWSTR, cchMaxCount int32) (int32, WIN32_ERROR) {
	addr := lazyAddr(&pGetClipboardFormatNameW, libUser32, "GetClipboardFormatNameW")
	ret, _, err := syscall.SyscallN(addr, uintptr(format), uintptr(unsafe.Pointer(lpszFormatName)), uintptr(cchMaxCount))
	return int32(ret), WIN32_ERROR(err)
}

func EmptyClipboard() (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pEmptyClipboard, libUser32, "EmptyClipboard")
	ret, _, err := syscall.SyscallN(addr)
	return BOOL(ret), WIN32_ERROR(err)
}

func IsClipboardFormatAvailable(format uint32) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pIsClipboardFormatAvailable, libUser32, "IsClipboardFormatAvailable")
	ret, _, err := syscall.SyscallN(addr, uintptr(format))
	return BOOL(ret), WIN32_ERROR(err)
}

func GetPriorityClipboardFormat(paFormatPriorityList *uint32, cFormats int32) (int32, WIN32_ERROR) {
	addr := lazyAddr(&pGetPriorityClipboardFormat, libUser32, "GetPriorityClipboardFormat")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(paFormatPriorityList)), uintptr(cFormats))
	return int32(ret), WIN32_ERROR(err)
}

func GetOpenClipboardWindow() (HWND, WIN32_ERROR) {
	addr := lazyAddr(&pGetOpenClipboardWindow, libUser32, "GetOpenClipboardWindow")
	ret, _, err := syscall.SyscallN(addr)
	return ret, WIN32_ERROR(err)
}

func AddClipboardFormatListener(hwnd HWND) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pAddClipboardFormatListener, libUser32, "AddClipboardFormatListener")
	ret, _, err := syscall.SyscallN(addr, hwnd)
	return BOOL(ret), WIN32_ERROR(err)
}

func RemoveClipboardFormatListener(hwnd HWND) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pRemoveClipboardFormatListener, libUser32, "RemoveClipboardFormatListener")
	ret, _, err := syscall.SyscallN(addr, hwnd)
	return BOOL(ret), WIN32_ERROR(err)
}

func GetUpdatedClipboardFormats(lpuiFormats *uint32, cFormats uint32, pcFormatsOut *uint32) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pGetUpdatedClipboardFormats, libUser32, "GetUpdatedClipboardFormats")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpuiFormats)), uintptr(cFormats), uintptr(unsafe.Pointer(pcFormatsOut)))
	return BOOL(ret), WIN32_ERROR(err)
}

func GlobalDeleteAtom(nAtom uint16) (uint16, WIN32_ERROR) {
	addr := lazyAddr(&pGlobalDeleteAtom, libKernel32, "GlobalDeleteAtom")
	ret, _, err := syscall.SyscallN(addr, uintptr(nAtom))
	return uint16(ret), WIN32_ERROR(err)
}

func InitAtomTable(nSize uint32) BOOL {
	addr := lazyAddr(&pInitAtomTable, libKernel32, "InitAtomTable")
	ret, _, _ := syscall.SyscallN(addr, uintptr(nSize))
	return BOOL(ret)
}

func DeleteAtom(nAtom uint16) (uint16, WIN32_ERROR) {
	addr := lazyAddr(&pDeleteAtom, libKernel32, "DeleteAtom")
	ret, _, err := syscall.SyscallN(addr, uintptr(nAtom))
	return uint16(ret), WIN32_ERROR(err)
}

func GlobalAddAtomA(lpString PSTR) (uint16, WIN32_ERROR) {
	addr := lazyAddr(&pGlobalAddAtomA, libKernel32, "GlobalAddAtomA")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpString)))
	return uint16(ret), WIN32_ERROR(err)
}

var GlobalAddAtom = GlobalAddAtomW

func GlobalAddAtomW(lpString PWSTR) (uint16, WIN32_ERROR) {
	addr := lazyAddr(&pGlobalAddAtomW, libKernel32, "GlobalAddAtomW")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpString)))
	return uint16(ret), WIN32_ERROR(err)
}

func GlobalAddAtomExA(lpString PSTR, Flags uint32) (uint16, WIN32_ERROR) {
	addr := lazyAddr(&pGlobalAddAtomExA, libKernel32, "GlobalAddAtomExA")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpString)), uintptr(Flags))
	return uint16(ret), WIN32_ERROR(err)
}

var GlobalAddAtomEx = GlobalAddAtomExW

func GlobalAddAtomExW(lpString PWSTR, Flags uint32) (uint16, WIN32_ERROR) {
	addr := lazyAddr(&pGlobalAddAtomExW, libKernel32, "GlobalAddAtomExW")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpString)), uintptr(Flags))
	return uint16(ret), WIN32_ERROR(err)
}

func GlobalFindAtomA(lpString PSTR) (uint16, WIN32_ERROR) {
	addr := lazyAddr(&pGlobalFindAtomA, libKernel32, "GlobalFindAtomA")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpString)))
	return uint16(ret), WIN32_ERROR(err)
}

var GlobalFindAtom = GlobalFindAtomW

func GlobalFindAtomW(lpString PWSTR) (uint16, WIN32_ERROR) {
	addr := lazyAddr(&pGlobalFindAtomW, libKernel32, "GlobalFindAtomW")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpString)))
	return uint16(ret), WIN32_ERROR(err)
}

func GlobalGetAtomNameA(nAtom uint16, lpBuffer PSTR, nSize int32) (uint32, WIN32_ERROR) {
	addr := lazyAddr(&pGlobalGetAtomNameA, libKernel32, "GlobalGetAtomNameA")
	ret, _, err := syscall.SyscallN(addr, uintptr(nAtom), uintptr(unsafe.Pointer(lpBuffer)), uintptr(nSize))
	return uint32(ret), WIN32_ERROR(err)
}

var GlobalGetAtomName = GlobalGetAtomNameW

func GlobalGetAtomNameW(nAtom uint16, lpBuffer PWSTR, nSize int32) (uint32, WIN32_ERROR) {
	addr := lazyAddr(&pGlobalGetAtomNameW, libKernel32, "GlobalGetAtomNameW")
	ret, _, err := syscall.SyscallN(addr, uintptr(nAtom), uintptr(unsafe.Pointer(lpBuffer)), uintptr(nSize))
	return uint32(ret), WIN32_ERROR(err)
}

func AddAtomA(lpString PSTR) (uint16, WIN32_ERROR) {
	addr := lazyAddr(&pAddAtomA, libKernel32, "AddAtomA")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpString)))
	return uint16(ret), WIN32_ERROR(err)
}

var AddAtom = AddAtomW

func AddAtomW(lpString PWSTR) (uint16, WIN32_ERROR) {
	addr := lazyAddr(&pAddAtomW, libKernel32, "AddAtomW")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpString)))
	return uint16(ret), WIN32_ERROR(err)
}

func FindAtomA(lpString PSTR) (uint16, WIN32_ERROR) {
	addr := lazyAddr(&pFindAtomA, libKernel32, "FindAtomA")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpString)))
	return uint16(ret), WIN32_ERROR(err)
}

var FindAtom = FindAtomW

func FindAtomW(lpString PWSTR) (uint16, WIN32_ERROR) {
	addr := lazyAddr(&pFindAtomW, libKernel32, "FindAtomW")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpString)))
	return uint16(ret), WIN32_ERROR(err)
}

func GetAtomNameA(nAtom uint16, lpBuffer PSTR, nSize int32) (uint32, WIN32_ERROR) {
	addr := lazyAddr(&pGetAtomNameA, libKernel32, "GetAtomNameA")
	ret, _, err := syscall.SyscallN(addr, uintptr(nAtom), uintptr(unsafe.Pointer(lpBuffer)), uintptr(nSize))
	return uint32(ret), WIN32_ERROR(err)
}

var GetAtomName = GetAtomNameW

func GetAtomNameW(nAtom uint16, lpBuffer PWSTR, nSize int32) (uint32, WIN32_ERROR) {
	addr := lazyAddr(&pGetAtomNameW, libKernel32, "GetAtomNameW")
	ret, _, err := syscall.SyscallN(addr, uintptr(nAtom), uintptr(unsafe.Pointer(lpBuffer)), uintptr(nSize))
	return uint32(ret), WIN32_ERROR(err)
}
