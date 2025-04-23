package ju

import (
	"fmt"
	"github.com/gookit/color"
	"strings"
	"sync"
)

const (
	ColorRed     = "red"
	ColorGreen   = "green"
	ColorYellow  = "yellow"
	ColorBlack   = "black"
	ColorWhite   = "white"
	ColorMagenta = "magenta"
	ColorCyan    = "cyan"
	ColorBlue    = "blue"
)

var _logMutex sync.Mutex

type ColorPrint func(format string, a ...interface{})

// OutputColor 这个函数输出效果和logColor相同，但是只输出到控制台，任何时候都不会保存到数据库
// skip: 0 是OutputColor的调用位置, 1 是上一级函数的调用位置
func OutputColor(skip int, color string, v ...interface{}) {
	trace := GetTrace(skip + 2)
	var builder strings.Builder
	for i, value := range v {
		if i > 0 {
			builder.WriteString(" ")
		}
		builder.WriteString(fmt.Sprint(value))
	}
	str := builder.String()
	_logMutex.Lock()
	defer _logMutex.Unlock()
	fmt.Print(GetNowTimeMs(), " ", trace, " ")
	cp := GetColorPrint(color)
	cp("%s\n", str)
}
func GetColorPrint(c string) (cp ColorPrint) {
	switch c {
	case ColorBlack:
		return color.Black.Printf
	case ColorWhite:
		return color.White.Printf
	case ColorGreen:
		return color.Green.Printf
	case ColorRed:
		return color.Red.Printf
	case ColorBlue:
		return color.Blue.Printf
	case ColorMagenta:
		return color.Magenta.Printf
	case ColorYellow:
		return color.Yellow.Printf
	case ColorCyan:
		return color.Cyan.Printf
	}
	return color.Black.Printf
}
func logColor(skip int, c string, v ...interface{}) {
	trace := GetTrace(skip)

	var builder strings.Builder
	for i, value := range v {
		if i > 0 {
			builder.WriteString(" ")
		}
		builder.WriteString(fmt.Sprint(value))
	}
	str := builder.String()
	//saveLog(logParam.DbType, tab, trace, c, str)
	cp := GetColorPrint(c)

	_logMutex.Lock()
	defer _logMutex.Unlock()
	fmt.Print(GetNowTimeMs(), " ", trace, " ")
	cp("%s\n", str)
}

// noinspection GoUnusedExportedFunction
func LogBlack(a ...interface{}) { logColor(3, "black", a...) }

// noinspection GoUnusedExportedFunction
func LogRed(a ...interface{}) { logColor(3, "red", a...) }

// noinspection GoUnusedExportedFunction
func LogGreen(a ...interface{}) { logColor(3, "green", a...) }

// noinspection GoUnusedExportedFunction
func LogYellow(a ...interface{}) { logColor(3, "yellow", a...) }

// noinspection GoUnusedExportedFunction
func LogBlue(a ...interface{}) { logColor(3, "blue", a...) }

// noinspection GoUnusedExportedFunction
func LogMagenta(a ...interface{}) { logColor(3, "magenta", a...) }

// noinspection GoUnusedExportedFunction
func LogCyan(a ...interface{}) { logColor(3, "cyan", a...) }

// noinspection GoUnusedExportedFunction
func LogWhite(a ...interface{}) { logColor(3, "white", a...) }
