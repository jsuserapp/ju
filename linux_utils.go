//go:build !windows

package ju

import (
	"fmt"
	"os"
	"syscall"
)

var appLockFile *os.File

func AppIsRunning(appName string) bool {
	var err error
	appLockFile, err = os.OpenFile(fmt.Sprintf("./%s.lock", appName), os.O_RDWR|os.O_CREATE, 0666)
	if LogFail(err) {
		return false
	}
	// 尝试获取非阻塞的排他锁
	err = syscall.Flock(int(appLockFile.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
	if err != nil {
		_ = appLockFile.Close()
		appLockFile = nil
		LogRed("应用正在运行")
		return true
	}
	return false
}
func ReleaseAppLock() {
	if appLockFile != nil {
		_ = appLockFile.Close()
	}
}
