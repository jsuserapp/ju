//把 log 存到数据库的接口

package ju

type LogInfo struct {
	Id        int    `json:"id"`
	Color     int    `json:"color"`
	Log       string `json:"log"`
	Trace     string `json:"trace"`
	CreatedAt string `json:"created_at"`
}
type LogDb interface {
	// DeleteLog 删除指定 id 的日志
	DeleteLog(tag string, id int64)
	// DeleteTagLogs 删除特定 tag before 日期之前的所有日志，但是 created_at 恰好等于 before 的日志不会删除，返回删除的日志条数
	DeleteTagLogs(tag, before string) int64
	// DeleteLogs 删除 before 日期之前的所有日志，但是 created_at 恰好等于 before 的日志不会删除，返回删除的日志条数
	DeleteLogs(before string) int64
	// ClearTagLogs 清空指定 tag 的日志，返回删除的日志条数
	ClearTagLogs(tag string) int64
	// ClearLogs 清空全部日志，重置表的 id 索引
	ClearLogs()
	// GetLogs 获取 log，返回最多 count 条数据，page 是分页，从 0 开始，顺序返回第 page*count+1 到 (page+1)*count+1 条日志.
	// 如果出现错误 logs 会是 nil
	// total 是对应 tag 的日志总数
	GetLogs(tag string, page, count int) (logs []*LogInfo, total int64)
	//saveLog 保存日志，如果指定 tag 的日志达到设置上限，则替换掉最早的一条数据
	saveLog(tag, color, trace, log string) bool
}
