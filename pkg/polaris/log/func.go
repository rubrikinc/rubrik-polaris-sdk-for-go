package log

import (
	"runtime"
	"strings"
)

var basePath = "github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/"

// PkgFuncName returns the package and function name recorded in the callers
// frame skip number steps up the stack. Note that the function name might be
// absent in which case the empty string is returned.
func PkgFuncName(skip int) string {
	var name string
	if pc, _, _, ok := runtime.Caller(skip); ok {
		frame, _ := runtime.CallersFrames([]uintptr{pc}).Next()
		name = strings.TrimPrefix(frame.Function, basePath)
	}

	return name
}
