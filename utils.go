package ju

import (
	"os"
	"path/filepath"
)

func GetExeDirectory() string {
	// 1. 获取当前正在执行的 exe 的完整路径
	exePath, err := os.Executable()
	if LogFail(err) {
		return ""
	}

	// 2. 获取该路径所属的文件夹
	return filepath.Dir(exePath)
}

// ChangeWorkingDirectory 将当前进程的工作目录切换到该文件夹
func ChangeWorkingDirectory(workDir string) {
	err := os.Chdir(workDir)
	LogError(err)
}
