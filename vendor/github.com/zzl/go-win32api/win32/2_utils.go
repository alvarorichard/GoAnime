package win32

import (
	"errors"
	"log"
	"strconv"
	"strings"
	"syscall"
	"unsafe"
)

func HRESULT_ToString(hr HRESULT) string {
	const bufSize = 255
	buf := make([]uint16, bufSize)
	flags := FORMAT_MESSAGE_FROM_SYSTEM | FORMAT_MESSAGE_IGNORE_INSERTS
	FormatMessage(flags, nil, uint32(hr), 0, &buf[0], bufSize, nil)
	hrHex := strconv.FormatUint(uint64(uint32(hr)), 16)
	s := syscall.UTF16ToString(buf)
	s = strings.TrimSpace(s)
	s = "HRESULT " + hrHex + " - " + s
	return s
}

func ASSERT_SUCCEEDED(hr HRESULT) {
	if FAILED(hr) {
		log.Fatal(HRESULT_ToString(hr))
	}
}

func StrToPwstr(str string) PWSTR {
	pwsz, _ := syscall.UTF16PtrFromString(str)
	return pwsz
}

func StrToPstr(str string) PSTR {
	bts := []byte(str)
	return (PSTR)(&bts[0])
}

func StrToPointer(str string) unsafe.Pointer {
	pwsz, _ := syscall.UTF16PtrFromString(str)
	return unsafe.Pointer(pwsz)
}

func StrToPointerOrNil(str string) unsafe.Pointer {
	if str == "" {
		return nil
	}
	pwsz, _ := syscall.UTF16PtrFromString(str)
	return unsafe.Pointer(pwsz)
}

func StrToBstr(str string) BSTR {
	pwsz, _ := syscall.UTF16PtrFromString(str)
	return SysAllocString(pwsz)
}

func PwstrToStr(pwsz PWSTR) string {
	wsz := unsafe.Slice(pwsz, 1024*64)
	return WstrToStr(wsz)
}

func WstrToStr(wsz []uint16) string {
	str := syscall.UTF16ToString(wsz)
	return str
}

func BstrToStr(bs BSTR) string {
	len := SysStringLen(bs)
	if len == 0 {
		return ""
	}
	ws := unsafe.Slice(bs, len)
	s := syscall.UTF16ToString(ws)
	return s
}

func BstrToStrAndFree(bs BSTR) string {
	str := BstrToStr(bs)
	SysFreeString(bs)
	return str
}

func GuidToStr(guid *syscall.GUID) (string, error) {
	var wszGuid [40]uint16
	hr := StringFromGUID2(guid, &wszGuid[0], 40)
	var err error
	if FAILED(hr) {
		err = errors.New(HRESULT_ToString(hr))
	}
	return WstrToStr(wszGuid[1:37]), err
}

func StrToGuid(str string) (syscall.GUID, error) {
	if str == "" {
		return syscall.GUID{}, errors.New("empty guid string")
	}
	if str[0] != '{' {
		str = "{" + str + "}"
	}
	var guid syscall.GUID
	hr := CLSIDFromString(StrToPwstr(str), &guid)
	var err error
	if FAILED(hr) {
		err = errors.New(HRESULT_ToString(hr))
	}
	return guid, err
}

func (me WIN32_ERROR) Error() string {
	return syscall.Errno(me).Error()
}

func (me WIN32_ERROR) NilOrError() error {
	if me == NO_ERROR {
		return nil
	}
	return me
}

func BoolToBOOL(b bool) BOOL {
	if b {
		return TRUE
	} else {
		return FALSE
	}
}

func BoolFromBOOL(b BOOL) bool {
	return b != 0
}

func GetCh() byte {
	hStdIn, _ := GetStdHandle(STD_INPUT_HANDLE)
	if hStdIn == 0 {
		println("??")
		return 0
	}
	var cc DWORD
	var ir INPUT_RECORD
	for {
		ReadConsoleInput(hStdIn, &ir, 1, &cc)
		if ir.EventType == uint16(KEY_EVENT) &&
			ir.Event.KeyEvent().BKeyDown == TRUE {
			b := ir.Event.KeyEvent().UChar.AsciiCharVal()
			return b
		}
	}
}
