package api

import (
	"strings"
	"syscall"
	"unsafe"

	"github.com/mjwhitta/win/errors"
	"github.com/mjwhitta/win/types"
)

var winhttp *syscall.LazyDLL = syscall.NewLazyDLL("Winhttp")

// WinHTTPAddRequestHeaders is WinHttpAddRequestHeaders from winhttp.h
func WinHTTPAddRequestHeaders(
	reqHndl uintptr,
	header string,
	addMethod uintptr,
) error {
	var e error
	var ok uintptr
	var proc string = "WinHttpAddRequestHeaders"

	if header == "" {
		// Weird, just do nothing
		return nil
	}

	header = strings.TrimSpace(header) + "\r\n"

	ok, _, e = winhttp.NewProc(proc).Call(
		reqHndl,
		types.LpCwstr(header),
		uintptr(len(header)),
		addMethod,
	)
	if ok == 0 {
		return errors.Newf("%s: %w", proc, e)
	}

	return nil
}

// WinHTTPConnect is WinHttpConnect from winhttp.h
func WinHTTPConnect(
	sessionHndl uintptr,
	serverName string,
	serverPort int,
) (uintptr, error) {
	var connHndl uintptr
	var e error
	var proc string = "WinHttpConnect"

	connHndl, _, e = winhttp.NewProc(proc).Call(
		sessionHndl,
		types.LpCwstr(serverName),
		uintptr(serverPort),
		0,
	)
	if connHndl == 0 {
		return 0, errors.Newf("%s: %w", proc, e)
	}

	return connHndl, nil
}

// WinHTTPOpen is WinHttpOpen from winhttp.h
func WinHTTPOpen(
	userAgent string,
	accessType uintptr,
	proxy string,
	proxyBypass string,
	flags uintptr,
) (uintptr, error) {
	var e error
	var proc string = "WinHttpOpen"
	var sessionHndl uintptr

	sessionHndl, _, e = winhttp.NewProc(proc).Call(
		types.LpCwstr(userAgent),
		accessType,
		types.LpCwstr(proxy),
		types.LpCwstr(proxyBypass),
		flags,
	)
	if sessionHndl == 0 {
		return 0, errors.Newf("%s: %w", proc, e)
	}

	return sessionHndl, nil
}

// WinHTTPOpenRequest is WinHttpOpenRequest from winhttp.h
func WinHTTPOpenRequest(
	connHndl uintptr,
	verb string,
	objectName string,
	version string,
	referrer string,
	acceptTypes []string,
	flags uintptr,
) (uintptr, error) {
	var e error
	var ppwszAcceptTypes []*uint16
	var proc string = "WinHttpOpenRequest"
	var reqHndl uintptr

	// Convert to Windows types
	ppwszAcceptTypes = make([]*uint16, 1)
	for _, theType := range acceptTypes {
		if theType == "" {
			continue
		}

		ppwszAcceptTypes = append(
			ppwszAcceptTypes,
			types.Cwstr(theType),
		)
	}

	reqHndl, _, e = winhttp.NewProc(proc).Call(
		connHndl,
		types.LpCwstr(verb),
		types.LpCwstr(objectName),
		types.LpCwstr(version),
		types.LpCwstr(referrer),
		uintptr(unsafe.Pointer(&ppwszAcceptTypes[0])),
		flags,
	)
	if reqHndl == 0 {
		return 0, errors.Newf("%s: %w", proc, e)
	}

	return reqHndl, nil
}

// WinHTTPQueryDataAvailable is WinHttpQueryDataAvailable from winhttp.h
func WinHTTPQueryDataAvailable(reqHndl uintptr, bytesToRead *int64) error {
	var e error
	var proc string = "WinHttpQueryDataAvailable"
	var success uintptr

	success, _, e = winhttp.NewProc(proc).Call(
		reqHndl,
		uintptr(unsafe.Pointer(bytesToRead)),
	)
	if success == 0 {
		return errors.Newf("%s: %w", proc, e)
	}

	return nil
}

// WinHTTPQueryHeaders is WinHttpQueryHeaders from winhttp.h
func WinHTTPQueryHeaders(
	reqHndl uintptr,
	info uintptr,
	name string,
	buffer *[]byte,
	bufferLen *int,
	index *int,
) error {
	var b []uint16
	var e error
	var proc string = "WinHttpQueryHeaders"
	var pwszName uintptr
	var success uintptr

	// Convert to Windows types
	if *bufferLen > 0 {
		b = make([]uint16, *bufferLen)
	} else {
		b = make([]uint16, 1)
	}

	if (name != "") && (info == Winhttp.WinhttpQueryCustom) {
		pwszName = types.LpCwstr(name)
	} else {
		pwszName = Winhttp.WinhttpHeaderNameByIndex
	}

	success, _, e = winhttp.NewProc(proc).Call(
		reqHndl,
		info,
		pwszName,
		uintptr(unsafe.Pointer(&b[0])),
		uintptr(unsafe.Pointer(bufferLen)),
		uintptr(unsafe.Pointer(index)),
	)
	if success == 0 {
		return errors.Newf("%s: %w", proc, e)
	}

	*buffer = []byte(syscall.UTF16ToString(b))

	return nil
}

// WinHTTPReadData is WinHttpReadData from winhttp.h
func WinHTTPReadData(
	reqHndl uintptr,
	buffer *[]byte,
	bytesToRead int64,
	bytesRead *int64,
) error {
	var b []byte
	var e error
	var proc string = "WinHttpReadData"
	var success uintptr

	if bytesToRead > 0 {
		b = make([]byte, bytesToRead)
	} else {
		b = make([]byte, 1)
	}

	success, _, e = winhttp.NewProc(proc).Call(
		reqHndl,
		uintptr(unsafe.Pointer(&b[0])),
		uintptr(bytesToRead),
		uintptr(unsafe.Pointer(bytesRead)),
	)
	if success == 0 {
		return errors.Newf("%s: %w", proc, e)
	}

	*buffer = b

	return nil
}

// WinHTTPReceiveResponse is WinHttpReceiveResponse from winhttp.h
func WinHTTPReceiveResponse(reqHndl uintptr) error {
	var e error
	var proc string = "WinHttpReceiveResponse"
	var success uintptr

	success, _, e = winhttp.NewProc(proc).Call(reqHndl, 0)
	if success == 0 {
		return errors.Newf("%s: %w", proc, e)
	}

	return nil
}

// WinHTTPSendRequest is WinHttpSendRequest from winhttp.h
func WinHTTPSendRequest(
	reqHndl uintptr,
	headers string,
	headersLen int,
	data []byte,
	dataLen int,
) error {
	var body uintptr
	var e error
	var proc string = "WinHttpSendRequest"
	var success uintptr

	// Pointer to data if provided
	if (data != nil) && (len(data) > 0) {
		body = uintptr(unsafe.Pointer(&data[0]))
	}

	success, _, e = winhttp.NewProc(proc).Call(
		reqHndl,
		types.LpCwstr(headers),
		uintptr(headersLen),
		body,
		uintptr(dataLen),
		uintptr(dataLen),
	)
	if success == 0 {
		return errors.Newf("%s: %w", proc, e)
	}

	return nil
}

// WinHTTPSetOption is WinHttpSetOption from winhttp.h
func WinHTTPSetOption(hndl, opt uintptr, val []byte, valLen int) error {
	var e error
	var proc string = "WinHttpSetOption"
	var success uintptr

	// Pointer to data if provided
	if valLen == 0 {
		val = make([]byte, 1)
	}

	success, _, e = winhttp.NewProc(proc).Call(
		hndl,
		opt,
		uintptr(unsafe.Pointer(&val[0])),
		uintptr(valLen),
	)
	if success == 0 {
		return errors.Newf("%s: %w", proc, e)
	}

	return nil
}
