package ju

import (
	"database/sql"
	"fmt"
	"github.com/gookit/color"
	"strings"
	"sync"
	"sync/atomic"
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

// _LogParam
// LogDb: A valid Db value
// SaveToLog: log whether save to database, default is false
// Log: interface of log method
type _LogParam struct {
	Mutex   sync.Mutex
	Save    atomic.Bool
	DbType  string
	LogDb   *sql.DB
	LogPath string
}

func init() {
	logParam = &_LogParam{}
}

var logParam *_LogParam

// SetLogDb 设置 Log 保存到数据库的方式
//
// dbType string: 数据库类型, 目前支持 sqlite 和 mysql 两个值
//
// db *DB: 数据库, 支持 MySql 和 SQLite3, 如果这个值是 nil, 则相当于取消日志系统绑定的数据库
//
// save bool: 是否保存到数据库, 即使设置了日志系统绑定的数据库, 仍然可以设置不保存到数据库, 只打印到输出窗口
// noinspection GoUnusedExportedFunction
func SetLogDb(dbType string, db *sql.DB, save bool) {
	if dbType != LogDbTypeSqlite && dbType != LogDbTypeMysql && dbType != LogDbTypePostgre {
		OutputColor(0, ColorRed, "不支持的数据库类型, 必须是 sqlite,mysql,postgre 之一,", dbType)
		return
	}
	logParam.DbType = dbType
	logParam.LogDb = db
	logParam.Save.Store(save)
	logInfo.Load()
}
func SetLogPath(path string, save bool) {
	if path == "" {
		path = "./data/log"
	}
	if !CreateFolder(path) {
		OutputColor(0, ColorRed, "指定的路径无法打开:", path)
		return
	}
	logParam.LogPath = path
	logParam.Save.Store(save)
}

// SetLogLimit 设置日志表的最多条数, 防止日志无限增长, 但是旧的日志会被覆盖. 对于文件日志类型, 这个函数是无效的.
//
// name string: 需要设置的日志名称, 空串对应的是默认日志.
//
// limit int64: 设置日志的最大条数, 此值为 0 或负数, 则取消最大条数.
func SetLogLimit(name string, limit int64) {
	logInfo.Set(name, limit)
}

// ClearLog 清空指定日志数据库中的所有记录
func ClearLog(tab string) {
	clearLog(logParam.DbType, tab)
}

// DeleteLog 删除默认的日志记录，idStart 到 idStop 的log都会被删除，包含这两个 id，要删除一个id 设置 idStart = id = idStop
func DeleteLog(tab string, idStart, idStop int64) {
	deleteLog(logParam.DbType, tab, idStart, idStop)
}

type colorPrint func(format string, a ...interface{})

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
	logParam.Mutex.Lock()
	defer logParam.Mutex.Unlock()
	fmt.Print(GetNowTimeMs(), " ", trace, " ")
	cp := getColorPrint(color)
	cp("%s\n", str)
}
func logToColor(skip int, c, tab string, v ...interface{}) {
	trace := GetTrace(skip)

	var builder strings.Builder
	for i, value := range v {
		if i > 0 {
			builder.WriteString(" ")
		}
		builder.WriteString(fmt.Sprint(value))
	}
	str := builder.String()
	saveLog(logParam.DbType, tab, trace, c, str)
	cp := getColorPrint(c)

	logParam.Mutex.Lock()
	defer logParam.Mutex.Unlock()
	fmt.Print(GetNowTimeMs(), " ", trace, " ")
	cp("%s\n", str)
}

func logToColorF(skip int, c, tab, format string, v ...interface{}) {
	trace := GetTrace(skip)
	saveLog(logParam.DbType, tab, trace, c, fmt.Sprintf(format, v...))
	cp := getColorPrint(c)

	logParam.Mutex.Lock()
	defer logParam.Mutex.Unlock()
	fmt.Print(GetNowTimeMs(), " ", trace, " ")
	cp(format, v...)
}
func getColorPrint(c string) (cp colorPrint) {
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

// noinspection GoUnusedExportedFunction
func LogBlackF(format string, a ...interface{}) { logToColorF(3, "black", "", format, a...) }

// noinspection GoUnusedExportedFunction
func LogRedF(format string, a ...interface{}) { logToColorF(3, "red", "", format, a...) }

// noinspection GoUnusedExportedFunction
func LogGreenF(format string, a ...interface{}) { logToColorF(3, "green", "", format, a...) }

// noinspection GoUnusedExportedFunction
func LogYellowF(format string, a ...interface{}) { logToColorF(3, "yellow", "", format, a...) }

// noinspection GoUnusedExportedFunction
func LogBlueF(format string, a ...interface{}) { logToColorF(3, "blue", "", format, a...) }

// noinspection GoUnusedExportedFunction
func LogMagentaF(format string, a ...interface{}) { logToColorF(3, "magenta", "", format, a...) }

// noinspection GoUnusedExportedFunction
func LogCyanF(format string, a ...interface{}) { logToColorF(3, "cyan", "", format, a...) }

// noinspection GoUnusedExportedFunction
func LogWhiteF(format string, a ...interface{}) { logToColorF(3, "white", "", format, a...) }

// noinspection GoUnusedExportedFunction
func LogBlack(a ...interface{}) { logToColor(3, "black", "", a...) }

// noinspection GoUnusedExportedFunction
func LogRed(a ...interface{}) { logToColor(3, "red", "", a...) }

// noinspection GoUnusedExportedFunction
func LogGreen(a ...interface{}) { logToColor(3, "green", "", a...) }

// noinspection GoUnusedExportedFunction
func LogYellow(a ...interface{}) { logToColor(3, "yellow", "", a...) }

// noinspection GoUnusedExportedFunction
func LogBlue(a ...interface{}) { logToColor(3, "blue", "", a...) }

// noinspection GoUnusedExportedFunction
func LogMagenta(a ...interface{}) { logToColor(3, "magenta", "", a...) }

// noinspection GoUnusedExportedFunction
func LogCyan(a ...interface{}) { logToColor(3, "cyan", "", a...) }

// noinspection GoUnusedExportedFunction
func LogWhite(a ...interface{}) { logToColor(3, "white", "", a...) }

// noinspection GoUnusedExportedFunction
func LogToBlackF(tab, format string, a ...interface{}) { logToColorF(3, "black", tab, format, a...) }

// noinspection GoUnusedExportedFunction
func LogToRedF(tab, format string, a ...interface{}) { logToColorF(3, "red", tab, format, a...) }

// noinspection GoUnusedExportedFunction
func LogToGreenF(tab, format string, a ...interface{}) { logToColorF(3, "green", tab, format, a...) }

// noinspection GoUnusedExportedFunction
func LogToYellowF(tab, format string, a ...interface{}) { logToColorF(3, "yellow", tab, format, a...) }

// noinspection GoUnusedExportedFunction
func LogToBlueF(tab, format string, a ...interface{}) { logToColorF(3, "blue", tab, format, a...) }

// noinspection GoUnusedExportedFunction
func LogToMagentaF(tab, format string, a ...interface{}) {
	logToColorF(3, "magenta", tab, format, a...)
}

// noinspection GoUnusedExportedFunction
func LogToCyanF(tab, format string, a ...interface{}) { logToColorF(3, "cyan", tab, format, a...) }

// noinspection GoUnusedExportedFunction
func LogToWhiteF(tab, format string, a ...interface{}) { logToColorF(3, "white", tab, format, a...) }

// noinspection GoUnusedExportedFunction
func LogToBlack(tab string, a ...interface{}) { logToColor(3, "black", tab, a...) }

// noinspection GoUnusedExportedFunction
func LogToRed(tab string, a ...interface{}) { logToColor(3, "red", tab, a...) }

// noinspection GoUnusedExportedFunction
func LogToGreen(tab string, a ...interface{}) { logToColor(3, "green", tab, a...) }

// noinspection GoUnusedExportedFunction
func LogToYellow(tab string, a ...interface{}) { logToColor(3, "yellow", tab, a...) }

// noinspection GoUnusedExportedFunction
func LogToBlue(tab string, a ...interface{}) { logToColor(3, "blue", tab, a...) }

// noinspection GoUnusedExportedFunction
func LogToMagenta(tab string, a ...interface{}) { logToColor(3, "magenta", tab, a...) }

// noinspection GoUnusedExportedFunction
func LogToCyan(tab string, a ...interface{}) { logToColor(3, "cyan", tab, a...) }

// noinspection GoUnusedExportedFunction
func LogToToWhite(tab string, a ...interface{}) { logToColor(3, "white", tab, a...) }
