package ju

import (
	"database/sql"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

const (
	createSqliteLogTab = `CREATE TABLE IF NOT EXISTS log (
    id INTEGER PRIMARY KEY, -- 在 SQLite 中, INTEGER PRIMARY KEY 默认就是自增的
    tag TEXT NOT NULL DEFAULT '',
    log TEXT NOT NULL,
    trace TEXT NOT NULL,
    color TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP -- SQLite 不支持 DATETIME 的精度定义
);`
	createSqliteLogTabIdx = `CREATE INDEX IF NOT EXISTS idx_tag_created_at ON log (tag, created_at);`
)

type SqliteLogDb struct {
	db *sql.DB
}

// CreateSqliteLogDb 返回一个 Sqlite3 的 LogDb 对象，db 参数必须是一个有效的 SqLite 数据库对象
func CreateSqliteLogDb(db *sql.DB) *SqliteLogDb {
	if db == nil {
		OutputColor(1, "red", "传入的数据库对象不能是 nil")
		return nil
	}
	ldb := &SqliteLogDb{
		db: db,
	}
	ldb.createLogTable()
	return ldb
}
func CloseSqliteLogDb(db *SqliteLogDb) {
	if db != nil && db.db != nil {
		err := db.db.Close()
		OutputErrorTrace(err, 1)
	}
}

// DeleteLog 删除指定 id 的日志
func (sdb *SqliteLogDb) DeleteLog(tag string, id int64) {
	_, err := sdb.db.Exec("DELETE FROM log WHERE tag=? AND id=?", tag, id)
	OutputErrorTrace(err, 0)
}

// DeleteTagLogs 删除特定 tag before 日期之前的所有日志，但是 created_at 恰好等于 before 的日志不会删除
func (sdb *SqliteLogDb) DeleteTagLogs(tag, before string) int64 {
	ret, err := sdb.db.Exec("DELETE FROM log WHERE tag=? AND created_at<=?", tag, before)
	if OutputErrorTrace(err, 0) {
		return 0
	}
	count, err := ret.RowsAffected()
	OutputErrorTrace(err, 0)
	return count
}

// DeleteLogs 删除 before 日期之前的所有日志，但是 created_at 恰好等于 before 的日志不会删除
func (sdb *SqliteLogDb) DeleteLogs(before string) int64 {
	ret, err := sdb.db.Exec("DELETE FROM log WHERE created_at<=?", before)
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
func (sdb *SqliteLogDb) GetLogs(tag string, page, count int) (logs []*LogInfo, total int64) {
	total = sdb.getTotalCount(tag)
	sqlCase := "SELECT id,log,trace,color,created_at FROM log WHERE tag=? ORDER BY created_at DESC LIMIT ?,?"
	start := count * page
	rows, err := sdb.db.Query(sqlCase, tag, start, count)
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
func (sdb *SqliteLogDb) getTotalCount(tag string) int64 {
	sqlCase := "SELECT count(*) FROM log WHERE tag=?"
	rows, err := sdb.db.Query(sqlCase, tag)
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
func (sdb *SqliteLogDb) ClearTagLogs(tag string) int64 {
	rst, err := sdb.db.Exec("DELETE FROM log WHERE tag=?", tag)
	if OutputErrorTrace(err, 0) {
		return 0
	}
	count, _ := rst.RowsAffected()
	return count
}

// ClearLogs 清空全部日志，并重置表的自增 ID 计数器。
// 这个版本是为 SQLite 定制的，并采用了最高效的 DROP + CREATE 方式。
func (sdb *SqliteLogDb) ClearLogs() {
	// 开启一个事务，以确保 DROP 和 CREATE 操作的原子性。
	tx, err := sdb.db.Begin()
	if OutputErrorTrace(err, 0) {
		return
	}
	// 使用 defer tx.Rollback() 保证在函数出错退出时，事务会自动回滚。
	defer func() {
		_ = tx.Rollback()
	}()

	// 1. 删除整个表。这是在 SQLite 中清空大表最快的方式。
	_, err = tx.Exec("DROP TABLE IF EXISTS log")
	if OutputErrorTrace(err, 0) {
		return
	}

	// 2. 立即用相同的结构重新创建表。
	// 这会自动重置所有状态，包括自增 ID。
	_, err = tx.Exec(createSqliteLogTab)
	if OutputErrorTrace(err, 0) {
		return
	}

	// 3. 重新创建与该表关联的所有索引。
	_, err = tx.Exec(createSqliteLogTabIdx)
	if OutputErrorTrace(err, 0) {
		return
	}

	// 4. 如果所有操作都成功，提交事务。
	err = tx.Commit()
	OutputErrorTrace(err, 0)
}
func (sdb *SqliteLogDb) saveLog(tag, color, trace, log string) bool {
	count := sdb.getTotalCount(tag)
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
	sqlCase := ""
	if insert {
		sqlCase = "INSERT INTO log (color,trace,log,created_at,tag) VALUES (?,?,?,?,?)"
	} else {
		if delCount > 0 {
			_, err := sdb.db.Exec("DELETE FROM log WHERE id IN (SELECT id FROM log WHERE tag = ? ORDER BY created_at LIMIT ?)", tag, delCount)
			if OutputErrorTrace(err, 0) {
				return false
			}
		}
		sqlCase = `UPDATE log
SET color = ?, trace = ?, log = ?, created_at = ?
WHERE id = (
    SELECT id
    FROM log
    WHERE tag = ?
    ORDER BY created_at
    LIMIT 1
);`
	}
	createdAt := time.Now().Format("2006-01-02 15:04:05.000")
	_, err := sdb.db.Exec(sqlCase, color, trace, log, createdAt, tag)
	return !OutputErrorTrace(err, 0)
}

func (sdb *SqliteLogDb) createLogTable() bool {
	_, err := sdb.db.Exec(createSqliteLogTab)
	if OutputErrorTrace(err, 0) {
		return false
	}
	_, err = sdb.db.Exec(createSqliteLogTabIdx)
	if OutputErrorTrace(err, 0) {
		return false
	}
	return err == nil
}
