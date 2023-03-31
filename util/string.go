package util

import (
	"runtime"
	"strings"
)

func GetFuncName() (output string) {
	pc := make([]uintptr, 15)

	n := runtime.Callers(2, pc)

	frames := runtime.CallersFrames(pc[:n])
	_, ok := frames.Next()
	if !ok {
		return output
	}

	f := runtime.FuncForPC(pc[0])
	fName := f.Name()

	var lastSlash, lastDot int

	if strings.Contains(fName, string('/')) {
		lastSlash = strings.LastIndexByte(fName, '/')
		lastDot = strings.LastIndexByte(fName[lastSlash:], '.') + lastSlash
		return fName[lastDot+1:]
	}
	return fName
}
