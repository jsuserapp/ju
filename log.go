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
var logParam = struct {
	db              LogDb
	output          bool
	maxLogCount     int64
	maxMainLogCount int64
}{output: true, maxLogCount: 1000, maxMainLogCount: 10000}

// SetLogParam 设置 Log 类函数的参数和表现，如果想让日志存储到数据，则设置一个有效的 db 对象
// output 指示 Log 函数是否输出到控制台，默认这个值是 true
// 注意：这个函数没有同步控制，也就是在并发调用 Log 类函数的时候，调用这个函数可能引发问题，所以一般在应用初始化时设置
// Log 类函数会同步输出到控制台，或者存储到数据库（没有同步），在高性能和高并发场合这个可能成为主要的性能瓶颈。
// maxLogCount,maxMainLogCount 分别是数据库存储日志的最大条数，默认分别是 1000 和 10000，如果日志数量达到这个数值，
// 后续的日志会替换掉同 tag 最早的日志，maxLogCount 是 tag 不为空串的日志上限，maxMainLogCount 是 tag 为空串的
// 日志的上限，tag 为空串的日志成为缺省日志。
// 如果这两个值设置 <= 0 则日志无上限。
func SetLogParam(db LogDb, output bool, maxLogCount, maxMainLogCount int64) {
	logParam.db = db
	logParam.output = output
	logParam.maxLogCount = maxLogCount
	logParam.maxMainLogCount = maxMainLogCount
}

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
	cp := GetColorPrint(color)
	_logMutex.Lock()
	defer _logMutex.Unlock()
	fmt.Print(GetNowTimeMs(), " ", trace, " ")
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
	return color.Gray.Printf
}
func logColor(skip int, color, tag string, v ...interface{}) {
	trace := GetTrace(skip)
	var builder strings.Builder
	for i, value := range v {
		if i > 0 {
			builder.WriteString(" ")
		}
		builder.WriteString(fmt.Sprint(value))
	}
	str := builder.String()

	if logParam.db != nil {
		logParam.db.saveLog(tag, color, trace, str)
	}
	if logParam.output {
		cp := GetColorPrint(color)

		_logMutex.Lock()
		defer _logMutex.Unlock()
		fmt.Print(GetNowTimeMs(), " ", trace, " ")
		cp("%s\n", str)
	}
}

// LogColor 以指定颜色输出，skip = 0 标记当前位置，skip = 1 标记上级函数调用位置，以此类推
// 这个函数和其它 Log 颜色函数功能相同，但是多了一个 color 参数，并且可以设置记录位置。
func LogColor(skip int, color string, v ...interface{}) {
	logColor(3+skip, color, "", v...)
}
func LogColorTo(skip int, color, tag string, v ...interface{}) {
	logColor(3+skip, color, tag, v...)
}

// noinspection GoUnusedExportedFunction
func LogBlack(a ...interface{}) { logColor(3, "black", "", a...) }

// noinspection GoUnusedExportedFunction
func LogRed(a ...interface{}) { logColor(3, "red", "", a...) }

// noinspection GoUnusedExportedFunction
func LogGreen(a ...interface{}) { logColor(3, "green", "", a...) }

// noinspection GoUnusedExportedFunction
func LogYellow(a ...interface{}) { logColor(3, "yellow", "", a...) }

// noinspection GoUnusedExportedFunction
func LogBlue(a ...interface{}) { logColor(3, "blue", "", a...) }

// noinspection GoUnusedExportedFunction
func LogMagenta(a ...interface{}) { logColor(3, "magenta", "", a...) }

// noinspection GoUnusedExportedFunction
func LogCyan(a ...interface{}) { logColor(3, "cyan", "", a...) }

// noinspection GoUnusedExportedFunction
func LogWhite(a ...interface{}) { logColor(3, "white", "", a...) }

// noinspection GoUnusedExportedFunction
func LogMagentaTo(tag string, a ...interface{}) { logColor(3, "magenta", tag, a...) }

// noinspection GoUnusedExportedFunction
func LogCyanTo(tag string, a ...interface{}) { logColor(3, "cyan", tag, a...) }

// noinspection GoUnusedExportedFunction
func LogWhiteTo(tag string, a ...interface{}) { logColor(3, "white", tag, a...) }

// noinspection GoUnusedExportedFunction
func LogBlackTo(tag string, a ...interface{}) { logColor(3, "black", tag, a...) }

// noinspection GoUnusedExportedFunction
func LogRedTo(tag string, a ...interface{}) { logColor(3, "red", tag, a...) }

// noinspection GoUnusedExportedFunction
func LogGreenTo(tag string, a ...interface{}) { logColor(3, "green", tag, a...) }

// noinspection GoUnusedExportedFunction
func LogYellowTo(tag string, a ...interface{}) { logColor(3, "yellow", tag, a...) }

// noinspection GoUnusedExportedFunction
func LogBlueTo(tag string, a ...interface{}) { logColor(3, "blue", tag, a...) }

func outputColorF(skip int, color, format string, v ...interface{}) {
	trace := GetTrace(skip)
	cp := GetColorPrint(color)

	_logMutex.Lock()
	defer _logMutex.Unlock()
	fmt.Print(GetNowTimeMs(), " ", trace, " ")
	cp(format, v...)
}

// noinspection GoUnusedExportedFunction
func OutputBlackF(format string, a ...interface{}) { outputColorF(3, "black", format, a...) }

// noinspection GoUnusedExportedFunction
func OutputRedF(format string, a ...interface{}) { outputColorF(3, "red", format, a...) }

// noinspection GoUnusedExportedFunction
func OutputGreenF(format string, a ...interface{}) { outputColorF(3, "green", format, a...) }

// noinspection GoUnusedExportedFunction
func OutputYellowF(format string, a ...interface{}) { outputColorF(3, "yellow", format, a...) }

// noinspection GoUnusedExportedFunction
func OutputBlueF(format string, a ...interface{}) { outputColorF(3, "blue", format, a...) }

// noinspection GoUnusedExportedFunction
func OutputMagentaF(format string, a ...interface{}) { outputColorF(3, "magenta", format, a...) }

// noinspection GoUnusedExportedFunction
func OutputCyanF(format string, a ...interface{}) { outputColorF(3, "cyan", format, a...) }

// noinspection GoUnusedExportedFunction
func OutputWhiteF(format string, a ...interface{}) { outputColorF(3, "white", format, a...) }
