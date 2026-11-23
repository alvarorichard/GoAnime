package win32

import (
	"syscall"
	"unsafe"
)

const (
	PIPE_UNLIMITED_INSTANCES uint32 = 0xff
	NMPWAIT_WAIT_FOREVER     uint32 = 0xffffffff
	NMPWAIT_NOWAIT           uint32 = 0x1
	NMPWAIT_USE_DEFAULT_WAIT uint32 = 0x0
)

// enums

// enum
// flags
type NAMED_PIPE_MODE uint32

const (
	PIPE_WAIT                  NAMED_PIPE_MODE = 0
	PIPE_NOWAIT                NAMED_PIPE_MODE = 1
	PIPE_READMODE_BYTE         NAMED_PIPE_MODE = 0
	PIPE_READMODE_MESSAGE      NAMED_PIPE_MODE = 2
	PIPE_CLIENT_END            NAMED_PIPE_MODE = 0
	PIPE_SERVER_END            NAMED_PIPE_MODE = 1
	PIPE_TYPE_BYTE             NAMED_PIPE_MODE = 0
	PIPE_TYPE_MESSAGE          NAMED_PIPE_MODE = 4
	PIPE_ACCEPT_REMOTE_CLIENTS NAMED_PIPE_MODE = 0
	PIPE_REJECT_REMOTE_CLIENTS NAMED_PIPE_MODE = 8
)

var (
	pCreatePipe                      uintptr
	pConnectNamedPipe                uintptr
	pDisconnectNamedPipe             uintptr
	pSetNamedPipeHandleState         uintptr
	pPeekNamedPipe                   uintptr
	pTransactNamedPipe               uintptr
	pCreateNamedPipeW                uintptr
	pWaitNamedPipeW                  uintptr
	pGetNamedPipeClientComputerNameW uintptr
	pImpersonateNamedPipeClient      uintptr
	pGetNamedPipeInfo                uintptr
	pGetNamedPipeHandleStateW        uintptr
	pCallNamedPipeW                  uintptr
	pCreateNamedPipeA                uintptr
	pGetNamedPipeHandleStateA        uintptr
	pCallNamedPipeA                  uintptr
	pWaitNamedPipeA                  uintptr
	pGetNamedPipeClientComputerNameA uintptr
	pGetNamedPipeClientProcessId     uintptr
	pGetNamedPipeClientSessionId     uintptr
	pGetNamedPipeServerProcessId     uintptr
	pGetNamedPipeServerSessionId     uintptr
)

func CreatePipe(hReadPipe *HANDLE, hWritePipe *HANDLE, lpPipeAttributes *SECURITY_ATTRIBUTES, nSize uint32) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pCreatePipe, libKernel32, "CreatePipe")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(hReadPipe)), uintptr(unsafe.Pointer(hWritePipe)), uintptr(unsafe.Pointer(lpPipeAttributes)), uintptr(nSize))
	return BOOL(ret), WIN32_ERROR(err)
}

func ConnectNamedPipe(hNamedPipe HANDLE, lpOverlapped *OVERLAPPED) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pConnectNamedPipe, libKernel32, "ConnectNamedPipe")
	ret, _, err := syscall.SyscallN(addr, hNamedPipe, uintptr(unsafe.Pointer(lpOverlapped)))
	return BOOL(ret), WIN32_ERROR(err)
}

func DisconnectNamedPipe(hNamedPipe HANDLE) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pDisconnectNamedPipe, libKernel32, "DisconnectNamedPipe")
	ret, _, err := syscall.SyscallN(addr, hNamedPipe)
	return BOOL(ret), WIN32_ERROR(err)
}

func SetNamedPipeHandleState(hNamedPipe HANDLE, lpMode *NAMED_PIPE_MODE, lpMaxCollectionCount *uint32, lpCollectDataTimeout *uint32) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pSetNamedPipeHandleState, libKernel32, "SetNamedPipeHandleState")
	ret, _, err := syscall.SyscallN(addr, hNamedPipe, uintptr(unsafe.Pointer(lpMode)), uintptr(unsafe.Pointer(lpMaxCollectionCount)), uintptr(unsafe.Pointer(lpCollectDataTimeout)))
	return BOOL(ret), WIN32_ERROR(err)
}

func PeekNamedPipe(hNamedPipe HANDLE, lpBuffer unsafe.Pointer, nBufferSize uint32, lpBytesRead *uint32, lpTotalBytesAvail *uint32, lpBytesLeftThisMessage *uint32) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pPeekNamedPipe, libKernel32, "PeekNamedPipe")
	ret, _, err := syscall.SyscallN(addr, hNamedPipe, uintptr(lpBuffer), uintptr(nBufferSize), uintptr(unsafe.Pointer(lpBytesRead)), uintptr(unsafe.Pointer(lpTotalBytesAvail)), uintptr(unsafe.Pointer(lpBytesLeftThisMessage)))
	return BOOL(ret), WIN32_ERROR(err)
}

func TransactNamedPipe(hNamedPipe HANDLE, lpInBuffer unsafe.Pointer, nInBufferSize uint32, lpOutBuffer unsafe.Pointer, nOutBufferSize uint32, lpBytesRead *uint32, lpOverlapped *OVERLAPPED) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pTransactNamedPipe, libKernel32, "TransactNamedPipe")
	ret, _, err := syscall.SyscallN(addr, hNamedPipe, uintptr(lpInBuffer), uintptr(nInBufferSize), uintptr(lpOutBuffer), uintptr(nOutBufferSize), uintptr(unsafe.Pointer(lpBytesRead)), uintptr(unsafe.Pointer(lpOverlapped)))
	return BOOL(ret), WIN32_ERROR(err)
}

var CreateNamedPipe = CreateNamedPipeW

func CreateNamedPipeW(lpName PWSTR, dwOpenMode FILE_FLAGS_AND_ATTRIBUTES, dwPipeMode NAMED_PIPE_MODE, nMaxInstances uint32, nOutBufferSize uint32, nInBufferSize uint32, nDefaultTimeOut uint32, lpSecurityAttributes *SECURITY_ATTRIBUTES) HANDLE {
	addr := lazyAddr(&pCreateNamedPipeW, libKernel32, "CreateNamedPipeW")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpName)), uintptr(dwOpenMode), uintptr(dwPipeMode), uintptr(nMaxInstances), uintptr(nOutBufferSize), uintptr(nInBufferSize), uintptr(nDefaultTimeOut), uintptr(unsafe.Pointer(lpSecurityAttributes)))
	return ret
}

var WaitNamedPipe = WaitNamedPipeW

func WaitNamedPipeW(lpNamedPipeName PWSTR, nTimeOut uint32) BOOL {
	addr := lazyAddr(&pWaitNamedPipeW, libKernel32, "WaitNamedPipeW")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpNamedPipeName)), uintptr(nTimeOut))
	return BOOL(ret)
}

var GetNamedPipeClientComputerName = GetNamedPipeClientComputerNameW

func GetNamedPipeClientComputerNameW(Pipe HANDLE, ClientComputerName PWSTR, ClientComputerNameLength uint32) BOOL {
	addr := lazyAddr(&pGetNamedPipeClientComputerNameW, libKernel32, "GetNamedPipeClientComputerNameW")
	ret, _, _ := syscall.SyscallN(addr, Pipe, uintptr(unsafe.Pointer(ClientComputerName)), uintptr(ClientComputerNameLength))
	return BOOL(ret)
}

func ImpersonateNamedPipeClient(hNamedPipe HANDLE) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pImpersonateNamedPipeClient, libAdvapi32, "ImpersonateNamedPipeClient")
	ret, _, err := syscall.SyscallN(addr, hNamedPipe)
	return BOOL(ret), WIN32_ERROR(err)
}

func GetNamedPipeInfo(hNamedPipe HANDLE, lpFlags *NAMED_PIPE_MODE, lpOutBufferSize *uint32, lpInBufferSize *uint32, lpMaxInstances *uint32) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pGetNamedPipeInfo, libKernel32, "GetNamedPipeInfo")
	ret, _, err := syscall.SyscallN(addr, hNamedPipe, uintptr(unsafe.Pointer(lpFlags)), uintptr(unsafe.Pointer(lpOutBufferSize)), uintptr(unsafe.Pointer(lpInBufferSize)), uintptr(unsafe.Pointer(lpMaxInstances)))
	return BOOL(ret), WIN32_ERROR(err)
}

var GetNamedPipeHandleState = GetNamedPipeHandleStateW

func GetNamedPipeHandleStateW(hNamedPipe HANDLE, lpState *NAMED_PIPE_MODE, lpCurInstances *uint32, lpMaxCollectionCount *uint32, lpCollectDataTimeout *uint32, lpUserName PWSTR, nMaxUserNameSize uint32) BOOL {
	addr := lazyAddr(&pGetNamedPipeHandleStateW, libKernel32, "GetNamedPipeHandleStateW")
	ret, _, _ := syscall.SyscallN(addr, hNamedPipe, uintptr(unsafe.Pointer(lpState)), uintptr(unsafe.Pointer(lpCurInstances)), uintptr(unsafe.Pointer(lpMaxCollectionCount)), uintptr(unsafe.Pointer(lpCollectDataTimeout)), uintptr(unsafe.Pointer(lpUserName)), uintptr(nMaxUserNameSize))
	return BOOL(ret)
}

var CallNamedPipe = CallNamedPipeW

func CallNamedPipeW(lpNamedPipeName PWSTR, lpInBuffer unsafe.Pointer, nInBufferSize uint32, lpOutBuffer unsafe.Pointer, nOutBufferSize uint32, lpBytesRead *uint32, nTimeOut uint32) BOOL {
	addr := lazyAddr(&pCallNamedPipeW, libKernel32, "CallNamedPipeW")
	ret, _, _ := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpNamedPipeName)), uintptr(lpInBuffer), uintptr(nInBufferSize), uintptr(lpOutBuffer), uintptr(nOutBufferSize), uintptr(unsafe.Pointer(lpBytesRead)), uintptr(nTimeOut))
	return BOOL(ret)
}

func CreateNamedPipeA(lpName PSTR, dwOpenMode FILE_FLAGS_AND_ATTRIBUTES, dwPipeMode NAMED_PIPE_MODE, nMaxInstances uint32, nOutBufferSize uint32, nInBufferSize uint32, nDefaultTimeOut uint32, lpSecurityAttributes *SECURITY_ATTRIBUTES) (HANDLE, WIN32_ERROR) {
	addr := lazyAddr(&pCreateNamedPipeA, libKernel32, "CreateNamedPipeA")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpName)), uintptr(dwOpenMode), uintptr(dwPipeMode), uintptr(nMaxInstances), uintptr(nOutBufferSize), uintptr(nInBufferSize), uintptr(nDefaultTimeOut), uintptr(unsafe.Pointer(lpSecurityAttributes)))
	return ret, WIN32_ERROR(err)
}

func GetNamedPipeHandleStateA(hNamedPipe HANDLE, lpState *NAMED_PIPE_MODE, lpCurInstances *uint32, lpMaxCollectionCount *uint32, lpCollectDataTimeout *uint32, lpUserName PSTR, nMaxUserNameSize uint32) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pGetNamedPipeHandleStateA, libKernel32, "GetNamedPipeHandleStateA")
	ret, _, err := syscall.SyscallN(addr, hNamedPipe, uintptr(unsafe.Pointer(lpState)), uintptr(unsafe.Pointer(lpCurInstances)), uintptr(unsafe.Pointer(lpMaxCollectionCount)), uintptr(unsafe.Pointer(lpCollectDataTimeout)), uintptr(unsafe.Pointer(lpUserName)), uintptr(nMaxUserNameSize))
	return BOOL(ret), WIN32_ERROR(err)
}

func CallNamedPipeA(lpNamedPipeName PSTR, lpInBuffer unsafe.Pointer, nInBufferSize uint32, lpOutBuffer unsafe.Pointer, nOutBufferSize uint32, lpBytesRead *uint32, nTimeOut uint32) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pCallNamedPipeA, libKernel32, "CallNamedPipeA")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpNamedPipeName)), uintptr(lpInBuffer), uintptr(nInBufferSize), uintptr(lpOutBuffer), uintptr(nOutBufferSize), uintptr(unsafe.Pointer(lpBytesRead)), uintptr(nTimeOut))
	return BOOL(ret), WIN32_ERROR(err)
}

func WaitNamedPipeA(lpNamedPipeName PSTR, nTimeOut uint32) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pWaitNamedPipeA, libKernel32, "WaitNamedPipeA")
	ret, _, err := syscall.SyscallN(addr, uintptr(unsafe.Pointer(lpNamedPipeName)), uintptr(nTimeOut))
	return BOOL(ret), WIN32_ERROR(err)
}

func GetNamedPipeClientComputerNameA(Pipe HANDLE, ClientComputerName PSTR, ClientComputerNameLength uint32) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pGetNamedPipeClientComputerNameA, libKernel32, "GetNamedPipeClientComputerNameA")
	ret, _, err := syscall.SyscallN(addr, Pipe, uintptr(unsafe.Pointer(ClientComputerName)), uintptr(ClientComputerNameLength))
	return BOOL(ret), WIN32_ERROR(err)
}

func GetNamedPipeClientProcessId(Pipe HANDLE, ClientProcessId *uint32) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pGetNamedPipeClientProcessId, libKernel32, "GetNamedPipeClientProcessId")
	ret, _, err := syscall.SyscallN(addr, Pipe, uintptr(unsafe.Pointer(ClientProcessId)))
	return BOOL(ret), WIN32_ERROR(err)
}

func GetNamedPipeClientSessionId(Pipe HANDLE, ClientSessionId *uint32) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pGetNamedPipeClientSessionId, libKernel32, "GetNamedPipeClientSessionId")
	ret, _, err := syscall.SyscallN(addr, Pipe, uintptr(unsafe.Pointer(ClientSessionId)))
	return BOOL(ret), WIN32_ERROR(err)
}

func GetNamedPipeServerProcessId(Pipe HANDLE, ServerProcessId *uint32) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pGetNamedPipeServerProcessId, libKernel32, "GetNamedPipeServerProcessId")
	ret, _, err := syscall.SyscallN(addr, Pipe, uintptr(unsafe.Pointer(ServerProcessId)))
	return BOOL(ret), WIN32_ERROR(err)
}

func GetNamedPipeServerSessionId(Pipe HANDLE, ServerSessionId *uint32) (BOOL, WIN32_ERROR) {
	addr := lazyAddr(&pGetNamedPipeServerSessionId, libKernel32, "GetNamedPipeServerSessionId")
	ret, _, err := syscall.SyscallN(addr, Pipe, uintptr(unsafe.Pointer(ServerSessionId)))
	return BOOL(ret), WIN32_ERROR(err)
}
