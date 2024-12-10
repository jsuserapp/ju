package ju

import (
	"fmt"
	"github.com/gookit/color"
	"strings"
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

// LogParam
// LogDb: A valid Db value, it's db value can't be nil
// SaveToLog: log whether save to database, default is false
type LogParam struct {
	logToDatabase atomic.Bool
	LogDb         *Db
}

func init() {
	logParam = &LogParam{}
}

var logParam *LogParam

type LogParamProc func(param *LogParam)

// SetLogParam Set log parameter
// noinspection GoUnusedExportedFunction
func SetLogParam(proc LogParamProc) {
	proc(logParam)
	if logParam.LogDb != nil {
		createLogTable("")
	}
}

// noinspection GoUnusedExportedFunction
func (lp *LogParam) SetLogToDb(log bool) {
	lp.logToDatabase.Store(log)
}
func createLogTable(tab string) {
	sqlCase := "CREATE TABLE IF NOT EXISTS `log` (id INTEGER PRIMARY KEY AUTO_INCREMENT, trace VARCHAR(255) NOT NULL DEFAULT '',color VARCHAR(32) DEFAULT '',log TEXT,created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP);"
	if tab != "" {
		sqlCase = strings.Replace(sqlCase, "log", "log_"+tab, 1)
	}
	logParam.LogDb.Exec(sqlCase)
}
func saveLogTo(tab, trace, color, log string) {
	sqlCase := "INSERT INTO `log` (`trace`,`color`,`log`, `created_at`) VALUES (?,?,?,?)"
	if tab != "" {
		sqlCase = strings.Replace(sqlCase, "log", "log_"+tab, 1)
	}
	sr := logParam.LogDb.Exec(sqlCase, trace, color, log, GetNowDateTime())
	if strings.Index(sr.Error, "no such table:") == 0 {
		createLogTable(tab)
		logParam.LogDb.Exec(sqlCase, trace, color, log, GetNowDateTime())
	}
}

// DeleteLog 删除默认的日志记录，idStart 到 idStop 的log都会被删除，包含这两个 id，要删除一个id 设置 idStart = id = idStop
// noinspection GoUnusedExportedFunction
func DeleteLog(idStart, idStop int64) int64 {
	return DeleteLogTo("", idStart, idStop)
}

// DeleteLogTo 删除指定表的日志记录，idStart 到 idStop 的log都会被删除，包含这两个 id，要删除一个id 设置 idStart = id = idStop
func DeleteLogTo(table string, idStart, idStop int64) int64 {
	sqlCase := "DELETE FROM log WHERE id>=? AND id<=?"
	if table != "" {
		sqlCase = strings.Replace(sqlCase, "log", "log_"+table, 1)
	}
	rst := logParam.LogDb.Exec(sqlCase, idStart, idStop)
	if rst.Error != "" {
		OutputColor(1, "red", rst.Error)
		return 0
	}
	count, _ := rst.Result.RowsAffected()
	return count
}

// ClearLog 清空数据库中的所有记录
// noinspection GoUnusedExportedFunction
func ClearLog() {
	ClearLogTo("")
}

// ClearLogTo 清空指定日志数据库中的所有记录
func ClearLogTo(table string) {
	if logParam.LogDb.db == nil {
		return
	}
	sqlCase := "TRUNCATE log"
	if table != "" {
		sqlCase = strings.Replace(sqlCase, "log", "log_"+table, 1)
	}
	logParam.LogDb.Exec(sqlCase)
}

// noinspection GoUnusedExportedFunction
func Log(format string, v ...interface{}) {
	trace := GetTrace(2)
	fmt.Print(trace, " ")
	fmt.Printf(format+"\n", v...)
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
	fmt.Print(GetNowTimeMs(), " ", trace, " ")
	cp := getColorPrint(c)
	cp("%s\n", str)
	if logParam.logToDatabase.Load() {
		saveLogTo(tab, trace, c, str)
	}
}

func logToColorF(skip int, c, tab, format string, v ...interface{}) {
	trace := GetTrace(skip)
	if logParam.logToDatabase.Load() {
		saveLogTo(tab, trace, c, fmt.Sprintf(format, v...))
	}
	fmt.Print(GetNowTimeMs(), " ", trace, " ")
	cp := getColorPrint(c)
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
