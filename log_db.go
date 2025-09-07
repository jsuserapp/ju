//把 log 存到数据库的接口

package ju

import "fmt"

type LogInfo struct {
	Id        int    `json:"id"`
	Color     string `json:"color"`
	Log       string `json:"log"`
	Trace     string `json:"trace"`
	CreatedAt string `json:"created_at"`
}

func (li *LogInfo) String() string {
	return fmt.Sprintf("%s [%s %s] %s", li.CreatedAt, li.Color, li.Trace, li.Log)
}

type LogDb interface {
	//saveLog 保存日志，如果指定 tag 的日志达到设置上限，则替换掉最早的一条数据
	saveLog(tag, color, trace, log string) bool
}
