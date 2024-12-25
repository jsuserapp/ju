package ju

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

//日志存储的数据库相关操作

const (
	LogDbTypeFile    = "file"
	LogDbTypeMysql   = "mysql"
	LogDbTypeSqlite  = "sqlite"
	LogDbTypePostgre = "postgre"
)

type LogInfo struct {
	sync.Mutex
	info map[string]int64
}

func (li *LogInfo) Get(name string) int64 {
	if name == "" {
		name = "log_"
	}
	li.Lock()
	count := li.info[name]
	li.Unlock()
	if count <= 0 {
		count = 1000
	}
	return count
}
func (li *LogInfo) Set(name string, count int64) {
	if name == "" {
		name = "log_"
	}
	if count <= 0 {
		count = 0
	}
	li.Lock()
	li.info[name] = count
	li.Unlock()
	if logParam.LogDb == nil {
		return
	}
	var sqlCase string
	if logParam.DbType == LogDbTypeSqlite {
		sqlCase = "INSERT INTO log_info (name, max_count) VALUES (?, ?) ON CONFLICT(name) DO UPDATE SET max_count = excluded.max_count"
	} else if logParam.DbType == LogDbTypeMysql {
		sqlCase = "INSERT INTO log_info (name, max_count) VALUES (?, ?) ON DUPLICATE KEY UPDATE max_count = VALUES(max_count)"
	} else if logParam.DbType == LogDbTypePostgre {
		sqlCase = "INSERT INTO log_info (name, max_count) VALUES ($1, $2) ON CONFLICT(name) DO UPDATE SET max_count = EXCLUDED.max_count"
	}
	_, err := logParam.LogDb.Exec(sqlCase, name, count)
	if err != nil {
		OutputColor(0, ColorRed, err.Error())
	}
}
func (li *LogInfo) Load() {
	if logParam.LogDb == nil {
		return
	}
	var sqlCase string
	if logParam.DbType == LogDbTypeSqlite {
		sqlCase = "CREATE TABLE IF NOT EXISTS `log_info` (`name` TEXT NOT NULL, `max_count` BIGINT DEFAULT 1000,PRIMARY KEY (`name`))"
	} else if logParam.DbType == LogDbTypeMysql {
		sqlCase = "CREATE TABLE IF NOT EXISTS `log_info` (`name` varchar(255) NOT NULL, `max_count` BIGINT DEFAULT 1000,PRIMARY KEY (`name`))"
	} else if logParam.DbType == LogDbTypePostgre {
		sqlCase = "CREATE TABLE IF NOT EXISTS log_info (name varchar(255) NOT NULL, max_count BIGINT DEFAULT 1000,PRIMARY KEY (name))"
	}
	_, err := logParam.LogDb.Exec(sqlCase)
	if err != nil {
		OutputColor(0, ColorRed, err.Error())
		return
	}

	sqlCase = "SELECT name,max_count FROM log_info"
	rows, err := logParam.LogDb.Query(sqlCase)
	if err != nil {
		OutputColor(0, ColorRed, err.Error())
		return
	}
	defer func() {
		_ = rows.Close()
	}()
	for rows.Next() {
		var name string
		var count int64
		err = rows.Scan(&name, &count)
		if err == nil {
			li.Set(name, count)
		} else {
			OutputColor(0, ColorRed, err.Error())
			break
		}
	}
}

var logInfo = LogInfo{info: map[string]int64{}}

func saveLog(dbType, tab, trace, color, log string) {
	if !logParam.Save.Load() {
		return
	}
	if dbType == LogDbTypeFile {
		if logParam.LogPath == "" {
			logParam.LogPath = "./data/log"
		}
		tab = "log_" + tab
		logFile := filepath.Join(logParam.LogPath, tab+".log")
		file, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			OutputColor(0, ColorRed, err.Error())
			return
		}
		defer func() {
			_ = file.Close()
		}()
		str := fmt.Sprintf("%s [%s] [%s] %s\r\n", GetNowDateTimeMs(), trace, color, log)
		_, err = file.WriteString(str)
		if err != nil {
			OutputColor(0, ColorRed, err.Error())
		}
		return
	}
	if logParam.LogDb == nil {
		return
	}
	var sqlCase string
	limit := logInfo.Get(tab)
	if limit > 0 {
		sqlCase = "SELECT COUNT(*) FROM log_"
		if tab != "" {
			sqlCase = strings.Replace(sqlCase, "log_", "log_"+tab, 1)
		}
		rows, err := logParam.LogDb.Query(sqlCase)
		if err != nil {
			var e SqlResult
			e.SetError(err)
			if (dbType == LogDbTypeMysql && e.Code == "1146") || (dbType == LogDbTypePostgre && e.Code == "42P01") || (dbType == LogDbTypeSqlite && e.Code == "") {
				createLogTable(dbType, tab)
				rows, err = logParam.LogDb.Query(sqlCase)
			}
			if err != nil {
				OutputColor(0, ColorRed, err.Error())
				return
			}
		}
		defer func() {
			_ = rows.Close()
		}()
		var count int64
		if rows.Next() {
			err = rows.Scan(&count)
			if err != nil {
				OutputColor(0, ColorRed, err.Error())
				return
			}
		}
		if limit < count {
			if logParam.DbType == LogDbTypePostgre {
				sqlCase = `
					WITH rows_to_delete AS (
						SELECT id
						FROM log_
						ORDER BY created_at
						LIMIT $1
					)
					DELETE FROM log_
					WHERE id IN (SELECT id FROM rows_to_delete);`
			} else if logParam.DbType == LogDbTypeMysql {
				sqlCase = "DELETE FROM log_ ORDER BY created_at LIMIT ?"
			} else if logParam.DbType == LogDbTypeSqlite {
				sqlCase = `
					WITH rows_to_delete AS (
						SELECT ROWID
						FROM log_
						ORDER BY created_at
						LIMIT ?
					)
					DELETE FROM log_
					WHERE ROWID IN (SELECT ROWID FROM rows_to_delete)
					RETURNING *;`
			}
			if tab != "" {
				sqlCase = strings.Replace(sqlCase, "log_", "log_"+tab, -1)
			}
			delCount := count - limit
			_, err = logParam.LogDb.Exec(sqlCase, delCount)
			if err != nil {
				OutputColor(0, ColorRed, err.Error())
				return
			}
			count = limit
		}
		if limit == count {
			if dbType == LogDbTypeMysql {
				sqlCase = "UPDATE log_ SET trace=?,color=?,log=?,created_at=? ORDER BY created_at LIMIT 1"
			} else if dbType == LogDbTypeSqlite {
				sqlCase = `
					WITH row_to_update AS (
						SELECT ROWID
						FROM log_
						ORDER BY created_at
						LIMIT 1
					)
					UPDATE log_
					SET trace = ?, 
						color = ?, 
						log = ?, 
						created_at = ?
					WHERE ROWID IN (SELECT ROWID FROM row_to_update)
					RETURNING *;`
			} else if dbType == LogDbTypePostgre {
				sqlCase = `
					WITH row_to_update AS (
						SELECT id
						FROM log_
						ORDER BY created_at
						LIMIT 1
					)
					UPDATE log_
					SET trace = $1, 
						color = $2, 
						log = $3, 
						created_at = $4
					WHERE id IN (SELECT id FROM row_to_update);
					`
			}
			if tab != "" {
				sqlCase = strings.Replace(sqlCase, "log_", "log_"+tab, -1)
			}
			_, err = logParam.LogDb.Exec(sqlCase, trace, color, log, GetNowDateTimeMs())
			if err != nil {
				OutputColor(0, ColorRed, err.Error())
			}
			return
		}
	}
	if dbType == LogDbTypeMysql || dbType == LogDbTypeSqlite {
		sqlCase = "INSERT INTO `log_` (`trace`,`color`,`log`, `created_at`) VALUES (?,?,?,?)"
	} else if dbType == LogDbTypePostgre {
		sqlCase = "INSERT INTO log_ (trace,color,log, created_at) VALUES ($1,$2,$3,$4)"
	}
	if tab != "" {
		sqlCase = strings.Replace(sqlCase, "log_", "log_"+tab, 1)
	}
	_, err := logParam.LogDb.Exec(sqlCase, trace, color, log, GetNowDateTimeMs())
	var e SqlResult
	e.SetError(err)
	if (dbType == LogDbTypeMysql && e.Code == "1146") || (dbType == LogDbTypePostgre && e.Code == "42P01") || (dbType == LogDbTypeSqlite && e.Code == "") {
		createLogTable(dbType, tab)
		_, err = logParam.LogDb.Exec(sqlCase, trace, color, log, GetNowDateTimeMs())
	}
	if err != nil {
		OutputColor(0, ColorRed, err.Error())
		return
	}
}
func clearLog(dbType, tab string) {
	if dbType == LogDbTypeFile {
		if logParam.LogPath == "" {
			logParam.LogPath = "./data/log"
		}
		tab = "log_" + tab
		logFile := filepath.Join(logParam.LogPath, tab+".log")
		file, err := os.OpenFile(logFile, os.O_WRONLY|os.O_TRUNC, 0644)
		if err != nil {
			OutputColor(0, ColorRed, err.Error())
			return
		}
		defer func() {
			_ = file.Close()
		}()
	} else if dbType == LogDbTypeMysql || dbType == LogDbTypeSqlite || dbType == LogDbTypePostgre {
		if logParam.LogDb == nil {
			return
		}
		sqlCase := "TRUNCATE log_"
		if tab != "" {
			sqlCase = strings.Replace(sqlCase, "log_", "log_"+tab, 1)
		}
		_, err := logParam.LogDb.Exec(sqlCase)
		if err != nil {
			OutputColor(0, ColorRed, err.Error())
		}
	}
}
func createLogTable(dbType, tab string) {
	if logParam.LogDb == nil || !logParam.Save.Load() {
		return
	}
	if dbType == LogDbTypeMysql {
		sqlCase := "CREATE TABLE IF NOT EXISTS `log_` (`id` BIGINT AUTO_INCREMENT,`trace` varchar(255) DEFAULT '',`log` text,`color` VARCHAR(32) DEFAULT '',`created_at` datetime(3) DEFAULT CURRENT_TIMESTAMP(3),PRIMARY KEY (`id`) USING BTREE,INDEX `idx_created_at`(`created_at`) USING BTREE);"
		if tab != "" {
			sqlCase = strings.Replace(sqlCase, "log_", "log_"+tab, 1)
		}
		_, err := logParam.LogDb.Exec(sqlCase)
		if err != nil {
			OutputColor(0, ColorRed, err.Error())
		}
	} else if dbType == LogDbTypeSqlite {
		sqlCase := `CREATE TABLE IF NOT EXISTS log_ (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    trace TEXT DEFAULT '',
    log TEXT DEFAULT '',
    color TEXT DEFAULT '',
    created_at DATETIME DEFAULT (DATETIME('now', 'localtime'))
);`
		if tab != "" {
			sqlCase = strings.Replace(sqlCase, "log_", "log_"+tab, 1)
		}
		_, err := logParam.LogDb.Exec(sqlCase)
		if err != nil {
			OutputColor(0, ColorRed, err.Error())
			return
		}
		sqlCase = `CREATE INDEX IF NOT EXISTS idx_created_at ON log_ (created_at);`
		if tab != "" {
			sqlCase = strings.Replace(sqlCase, "log_", "log_"+tab, 1)
		}
		_, err = logParam.LogDb.Exec(sqlCase)
		if err != nil {
			OutputColor(0, ColorRed, err.Error())
			return
		}
	} else if dbType == LogDbTypePostgre {
		sqlCase := `CREATE TABLE IF NOT EXISTS "log_" (
		"id" BIGSERIAL PRIMARY KEY,
		"trace" VARCHAR(255) DEFAULT '',
		"log" TEXT DEFAULT '',
		"color" VARCHAR(32) DEFAULT '',
		"created_at" TIMESTAMP(3) DEFAULT CURRENT_TIMESTAMP
	);`

		if tab != "" {
			sqlCase = strings.Replace(sqlCase, "log_", "log_"+tab, 1)
		}

		_, err := logParam.LogDb.Exec(sqlCase)
		if err != nil {
			OutputColor(0, ColorRed, err.Error())
		}
		sqlCase = `CREATE INDEX "idx_created_at" ON "log_" ("created_at");`
		if tab != "" {
			sqlCase = strings.Replace(sqlCase, "log_", "log_"+tab, 1)
		}
		_, err = logParam.LogDb.Exec(sqlCase)
		if err != nil {
			OutputColor(0, ColorRed, err.Error())
		}
	}
}
func deleteLog(dbType, tab string, idStart, idStop int64) int64 {
	if logParam.LogDb == nil {
		return 0
	}
	var sqlCase string
	if dbType == LogDbTypePostgre {
		sqlCase = `DELETE FROM "log_" WHERE "id" >= $1 AND "id" <= $2`
	} else if dbType == LogDbTypeMysql || dbType == LogDbTypeSqlite {
		sqlCase = "DELETE FROM log_ WHERE id>=? AND id<=?"
	} else {
		return 0
	}
	if tab != "" {
		sqlCase = strings.Replace(sqlCase, "log_", "log_"+tab, 1)
	}
	rst, err := logParam.LogDb.Exec(sqlCase, idStart, idStop)
	if err != nil {
		OutputColor(0, ColorRed, err.Error())
		return 0
	}
	count, _ := rst.RowsAffected()
	return count
}
