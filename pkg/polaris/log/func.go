package log

import (
	"runtime"
	"strings"
)

var basePath = "github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/"

// pkgFunc returns the package and function name recorded in the callers frame
// skip number steps up the stack. Note that the function name might be absent
// in which case the empty string is returned.
func pkgFunc(skip int) string {
	var name string
	if pc, _, _, ok := runtime.Caller(skip); ok {
		frame, _ := runtime.CallersFrames([]uintptr{pc}).Next()
		name = frame.Function

		if strings.HasPrefix(name, basePath) {
			name = name[len(basePath):]
		}
	}

	return name
}
