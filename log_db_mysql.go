package ju

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"time"
)

type MysqlLogDb struct {
	db *sql.DB
}

// CreateMysqlLogDb 返回一个 Mysql 的 LogDb 对象，db 参数必须是一个有效的 MySQL 数据库对象
func CreateMysqlLogDb(db *sql.DB) LogDb {
	if db == nil {
		OutputColor(1, "red", "传入的数据库对象不能是 nil")
		return nil
	}
	ldb := &MysqlLogDb{
		db: db,
	}
	ldb.createLogTable()
	return ldb
}
func CloseMysqlLogDb(db LogDb) {
	mdb := db.(*MysqlLogDb)
	if mdb != nil && mdb.db != nil {
		err := mdb.db.Close()
		OutputErrorTrace(err, 1)
	}
}

// DeleteLog 删除指定 id 的日志
func (mdb *MysqlLogDb) DeleteLog(tag string, id int64) {
	_, err := mdb.db.Exec("DELETE FROM log WHERE tag=? AND id=?", tag, id)
	OutputErrorTrace(err, 0)
}

// DeleteTagLogs 删除特定 tag before 日期之前的所有日志，但是 created_at 恰好等于 before 的日志不会删除
func (mdb *MysqlLogDb) DeleteTagLogs(tag, before string) int64 {
	ret, err := mdb.db.Exec("DELETE FROM log WHERE tag=? AND created_at<=?", tag, before)
	if OutputErrorTrace(err, 0) {
		return 0
	}
	count, err := ret.RowsAffected()
	OutputErrorTrace(err, 0)
	return count
}

// DeleteLogs 删除 before 日期之前的所有日志，但是 created_at 恰好等于 before 的日志不会删除
func (mdb *MysqlLogDb) DeleteLogs(before string) int64 {
	ret, err := mdb.db.Exec("DELETE FROM log WHERE created_at<=?", before)
	if err != nil {
		OutputColor(1, "red", err.Error())
		return 0
	}
	count, err := ret.RowsAffected()
	return count
}

// GetLogs 获取 log，返回最多 count 条数据，page 是分页，从 0 开始，顺序返回第 page*count+1 到 (page+1)*count+1 条日志.
// 如果出现错误 logs 会是 nil
// total 是对应 tag 的日志总数
func (mdb *MysqlLogDb) GetLogs(tag string, page, count int) (logs []*LogInfo, total int64) {
	total = mdb.getTotalCount(tag)
	sqlCase := "SELECT id,log,trace,color,created_at FROM log WHERE tag=? ORDER BY created_at DESC LIMIT ?,?"
	start := count * page
	rows, err := mdb.db.Query(sqlCase, tag, start, count)
	if OutputErrorTrace(err, 0) {
		return
	}
	defer func() {
		_ = rows.Close()
	}()
	logs = make([]*LogInfo, 0, count)
	for rows.Next() {
		var li LogInfo
		err = rows.Scan(&li.Id, &li.Log, &li.Trace, &li.Color, &li.CreatedAt)
		if !OutputErrorTrace(err, 0) {
			logs = append(logs, &li)
		}
	}
	return
}
func (mdb *MysqlLogDb) getTotalCount(tag string) int64 {
	sqlCase := "SELECT count(*) FROM log WHERE tag=?"
	rows, err := mdb.db.Query(sqlCase, tag)
	if OutputErrorTrace(err, 0) {
		return 0
	}
	defer func() {
		_ = rows.Close()
	}()
	var total int64
	if rows.Next() {
		err = rows.Scan(&total)
		if OutputErrorTrace(err, 0) {
			return 0
		}
	}
	return total
}

// ClearTagLogs 清空指定 tag 的日志
func (mdb *MysqlLogDb) ClearTagLogs(tag string) int64 {
	rst, err := mdb.db.Exec("DELETE FROM log WHERE tag=?", tag)
	if OutputErrorTrace(err, 0) {
		return 0
	}
	count, _ := rst.RowsAffected()
	return count
}

// ClearLogs 清空全部日志，重置表的 id 索引
func (mdb *MysqlLogDb) ClearLogs() {
	_, err := mdb.db.Exec("TRUNCATE log")
	OutputErrorTrace(err, 0)
}
func (mdb *MysqlLogDb) saveLog(tag, color, trace, log string) bool {
	count := mdb.getTotalCount(tag)
	insert := false
	var delCount int64
	if tag == "" {
		insert = logParam.maxMainLogCount <= 0 || count < logParam.maxMainLogCount
		if !insert {
			delCount = count - logParam.maxMainLogCount
		}
	} else {
		insert = logParam.maxLogCount <= 0 || count < logParam.maxLogCount
		if !insert {
			delCount = count - logParam.maxLogCount
		}
	}
	if insert {
		sqlCase := "INSERT INTO log (tag,color,trace,log) VALUES (?,?,?,?)"
		_, err := mdb.db.Exec(sqlCase, tag, color, trace, log)
		return !OutputErrorTrace(err, 0)
	} else {
		if delCount > 0 {
			_, err := mdb.db.Exec("DELETE FROM log WHERE tag=? ORDER BY created_at LIMIT ?", tag, delCount)
			if OutputErrorTrace(err, 0) {
				return false
			}
		}
		sqlCase := "UPDATE log SET color=?,trace=?,log=?,created_at=? WHERE tag=? ORDER BY created_at LIMIT 1"
		createdAt := time.Now().Format("2006-01-02 15:04:05.000")
		_, err := mdb.db.Exec(sqlCase, color, trace, log, createdAt, tag)
		return !OutputErrorTrace(err, 0)
	}
}
func (mdb *MysqlLogDb) createLogTable() bool {
	query :=
		`CREATE TABLE IF NOT EXISTS log(
		id INTEGER PRIMARY KEY AUTO_INCREMENT,
		tag VARCHAR(255) NOT NULL DEFAULT '',
		log TEXT NOT NULL,
		trace VARCHAR(255) NOT NULL,
		color VARCHAR(16),
		created_at DATETIME(3) DEFAULT CURRENT_TIMESTAMP(3),
		INDEX idx_tag_created_at (tag, created_at)
	);`
	_, err := mdb.db.Exec(query)
	OutputErrorTrace(err, 0)
	return err == nil
}
