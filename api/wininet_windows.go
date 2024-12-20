package api

import (
	"strings"
	"syscall"
	"unsafe"

	"github.com/mjwhitta/win/errors"
	"github.com/mjwhitta/win/types"
)

var wininet *syscall.LazyDLL = syscall.NewLazyDLL("Wininet")

// HTTPAddRequestHeadersW is from wininet.h
func HTTPAddRequestHeadersW(
	reqHndl uintptr,
	header string,
	addMethod uintptr,
) error {
	var e error
	var ok uintptr
	var proc string = "HttpAddRequestHeadersW"

	if header == "" {
		// Weird, just do nothing
		return nil
	}

	header = strings.TrimSpace(header) + "\r\n"

	ok, _, e = wininet.NewProc(proc).Call(
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

// HTTPOpenRequestW is from wininet.h
func HTTPOpenRequestW(
	connHndl uintptr,
	verb string,
	objectName string,
	version string,
	referrer string,
	acceptTypes []string,
	flags uintptr,
	context uintptr,
) (uintptr, error) {
	var e error
	var lplpcwstrAcceptTypes []*uint16
	var proc string = "HttpOpenRequestW"
	var reqHndl uintptr

	// Convert to Windows types
	lplpcwstrAcceptTypes = make([]*uint16, 1)
	for _, theType := range acceptTypes {
		if theType == "" {
			continue
		}

		lplpcwstrAcceptTypes = append(
			lplpcwstrAcceptTypes,
			types.Cwstr(theType),
		)
	}

	reqHndl, _, e = wininet.NewProc(proc).Call(
		connHndl,
		types.LpCwstr(verb),
		types.LpCwstr(objectName),
		types.LpCwstr(version),
		types.LpCwstr(referrer),
		uintptr(unsafe.Pointer(&lplpcwstrAcceptTypes[0])),
		flags,
		context,
	)
	if reqHndl == 0 {
		return 0, errors.Newf("%s: %w", proc, e)
	}

	return reqHndl, nil
}

// HTTPQueryInfoW is from wininet.h
func HTTPQueryInfoW(
	reqHndl uintptr,
	info uintptr,
	buffer *[]byte,
	bufferLen *int,
	index *int,
) error {
	var b []uint16
	var e error
	var proc string = "HttpQueryInfoW"
	var success uintptr
	var tmp string

	if *bufferLen > 0 {
		b = make([]uint16, *bufferLen)
	} else {
		b = make([]uint16, 1)
	}

	success, _, e = wininet.NewProc(proc).Call(
		reqHndl,
		info,
		uintptr(unsafe.Pointer(&b[0])),
		uintptr(unsafe.Pointer(bufferLen)),
		uintptr(unsafe.Pointer(index)),
	)
	if success == 0 {
		return errors.Newf("%s: %w", proc, e)
	}

	tmp = syscall.UTF16ToString(b)
	*buffer = []byte(tmp)

	return nil
}

// HTTPSendRequestW is from wininet.h
func HTTPSendRequestW(
	reqHndl uintptr,
	headers string,
	headersLen int,
	data []byte,
	dataLen int,
) error {
	var body uintptr
	var e error
	var proc string = "HttpSendRequestW"
	var success uintptr

	// Pointer to data if provided
	if (data != nil) && (len(data) > 0) {
		body = uintptr(unsafe.Pointer(&data[0]))
	}

	success, _, e = wininet.NewProc(proc).Call(
		reqHndl,
		types.LpCwstr(headers),
		uintptr(headersLen),
		body,
		uintptr(dataLen),
	)
	if success == 0 {
		return errors.Newf("%s: %w", proc, e)
	}

	return nil
}

// InternetConnectW is from wininet.h
func InternetConnectW(
	sessionHndl uintptr,
	serverName string,
	serverPort int,
	username string,
	password string,
	service uintptr,
	flags uintptr,
	context uintptr,
) (uintptr, error) {
	var connHndl uintptr
	var e error
	var proc string = "InternetConnectW"

	connHndl, _, e = wininet.NewProc(proc).Call(
		sessionHndl,
		types.LpCwstr(serverName),
		uintptr(serverPort),
		types.LpCwstr(username),
		types.LpCwstr(password),
		service,
		flags,
		context,
	)
	if connHndl == 0 {
		return 0, errors.Newf("%s: %w", proc, e)
	}

	return connHndl, nil
}

// InternetOpenW is from wininet.h
func InternetOpenW(
	userAgent string,
	accessType uintptr,
	proxy string,
	proxyBypass string,
	flags uintptr,
) (uintptr, error) {
	var e error
	var proc string = "InternetOpenW"
	var sessionHndl uintptr

	sessionHndl, _, e = wininet.NewProc(proc).Call(
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

// InternetQueryDataAvailable is from wininet.h
func InternetQueryDataAvailable(
	reqHndl uintptr,
	bytesAvailable *int64,
) error {
	var e error
	var proc string = "InternetQueryDataAvailable"
	var success uintptr

	success, _, e = wininet.NewProc(proc).Call(
		reqHndl,
		uintptr(unsafe.Pointer(bytesAvailable)),
		0,
		0,
	)
	if success == 0 {
		return errors.Newf("%s: %w", proc, e)
	}

	return nil
}

// InternetReadFile is from wininet.h
func InternetReadFile(
	reqHndl uintptr,
	buffer *[]byte,
	bytesToRead int64,
	bytesRead *int64,
) error {
	var b []byte
	var e error
	var proc string = "InternetReadFile"
	var success uintptr

	if bytesToRead > 0 {
		b = make([]byte, bytesToRead)
	} else {
		b = make([]byte, 1)
	}

	success, _, e = wininet.NewProc(proc).Call(
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

// InternetSetOptionW is from wininet.h
func InternetSetOptionW(
	hndl uintptr,
	opt uintptr,
	val []byte,
	valLen int,
) error {
	var e error
	var proc string = "InternetSetOptionW"
	var success uintptr

	// Pointer to data if provided
	if valLen == 0 {
		val = make([]byte, 1)
	}

	success, _, e = wininet.NewProc(proc).Call(
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
