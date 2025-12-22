//go:build windows

package ju

import (
	"errors"
	"syscall"
	"unsafe"
)

const (
	ErrorAlreadyExists = 183
)

var (
	kernel32     = syscall.NewLazyDLL("kernel32.dll")
	createMutex  = kernel32.NewProc("CreateMutexW")
	releaseMutex = kernel32.NewProc("ReleaseMutex")
	closeHandle  = kernel32.NewProc("CloseHandle")
)

var appLockHandle uintptr // 全局保存句柄，便于程序退出时释放

// AppIsRunning 返回 true 加锁失败，说明已经有进程在运行，程序退出，
// false 加锁成功 或者 其它错误引起的加锁失败，程序继续运行
func AppIsRunning(appName string) bool {
	appName = "global//" + appName
	// 将字符串转为 UTF16 指针
	appNamePtr, err := syscall.UTF16PtrFromString(appName)
	if LogFail(err) {
		return false
	}

	// CreateMutexW 返回句柄
	r, _, errCode := createMutex.Call(
		0,                                   // lpMutexAttributes (默认安全属性)
		0,                                   // bInitialOwner: false
		uintptr(unsafe.Pointer(appNamePtr)), // lpName: 互斥量名字（全局唯一）
	)

	if r == 0 {
		LogRed("创建互斥量失败")
		return false
	}

	appLockHandle = r
	if errors.Is(errCode, syscall.Errno(ErrorAlreadyExists)) {
		// 已经有一个实例在运行
		_, _, _ = closeHandle.Call(r) // 释放句柄
		return true
	}

	return false
}

// ReleaseAppLock 释放应用锁，程序退出时应该调用此函数，但是不调用，比如异常退出，它也会被自动清理
func ReleaseAppLock() {
	if appLockHandle != 0 {
		_, _, _ = releaseMutex.Call(appLockHandle)
		_, _, _ = closeHandle.Call(appLockHandle)
		appLockHandle = 0
	}
}
