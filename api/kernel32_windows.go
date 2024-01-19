package api

import (
	"syscall"

	"github.com/mjwhitta/win/types"
)

var kernel32 *syscall.LazyDLL = syscall.NewLazyDLL("kernel32")

// OutputDebugStringW will print a string that Dbgview.exe and
// dbgview64.exe will display. Useful for debugging DLLs.
func OutputDebugStringW(out string) {
	var proc string = "OutputDebugStringW"

	kernel32.NewProc(proc).Call(types.LpCwstr(out))
}
