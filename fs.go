package ju

import (
	"os"
	"path/filepath"
)

// CreateFolder 生成路径指定的文件夹，包括所有上级文件夹
// noinspection GoUnusedExportedFunction
func CreateFolder(folder string) bool {
	err := os.MkdirAll(folder, 0777)
	if err != nil {
		LogRed(err.Error())
		return false
	}
	return true
}

// DeleteFolder 删除文件夹和它的内容, path 也可以是一个文件
// noinspection GoUnusedExportedFunction
func DeleteFolder(path string) {
	_ = os.RemoveAll(path)
}
func ReadFile(fn string) []byte {
	file, err := os.Open(fn)
	if err != nil {
		//LogRed(err.Error())
		return nil
	}
	defer func(file *os.File) {
		_ = file.Close()
	}(file)
	var data []byte
	fi, err := os.Stat(fn)
	data = make([]byte, fi.Size())
	_, err = file.Read(data)
	if err != nil {
		LogRed(err.Error())
		return nil
	}
	return data
}

// noinspection GoUnusedExportedFunction
func FileExist(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// SaveFileAndFolder 函数会自动创建目标文件所在的目录文件夹，如果目标文件已经存在，数据会被覆盖，
// 此函数和 SaveFileToFolder 的区别是无需分开传输目录和文件参数
// noinspection GoUnusedExportedFunction
func SaveFileAndFolder(fn string, data []byte) bool {
	folder := filepath.Dir(fn)
	err := os.MkdirAll(folder, 0777)
	if LogFail(err) {
		return false
	}
	return SaveFile(fn, data)
}

// SaveFileToFolder 函数会自动创建目标文件所在的目录文件夹，如果目标文件已经存在，数据会被覆盖
// noinspection GoUnusedExportedFunction
func SaveFileToFolder(folder, fn string, data []byte) bool {
	err := os.MkdirAll(folder, 0777)
	if err != nil {
		LogRed(err.Error())
		return false
	}
	folder += "/" + fn
	return SaveFile(folder, data)
}

// SaveFileToFolderExistFail 保存到文件，并且生成所有上级文件夹（如果不存在），但是如果文件存在，函数失败，
// 用于防止已经存在的文件被覆盖
// noinspection GoUnusedExportedFunction
func SaveFileToFolderExistFail(folder, fn string, data []byte) bool {
	err := os.MkdirAll(folder, 0777)
	if err != nil {
		LogRed(err.Error())
		return false
	}
	folder += "/" + fn
	return SaveFileExistFail(folder, data)
}

// SaveFileExistFail 防止同名文件被覆盖，只有目标文件不存在才会成功
func SaveFileExistFail(fn string, data []byte) bool {
	file, err := os.OpenFile(fn, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0777)
	if err != nil {
		LogRed(err.Error())
		return false
	}
	defer func(file *os.File) {
		_ = file.Close()
	}(file)
	_, err = file.Write(data)
	LogError(err)
	return err == nil
}

// SaveFile 函数保存数据到指定文件，要求目标目录必须存在，否则失败。目标文件如果已经存在，会被新数据覆盖
func SaveFile(fn string, data []byte) bool {
	file, err := os.OpenFile(fn, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0777)
	if LogFail(err) {
		return false
	}
	defer func(file *os.File) {
		_ = file.Close()
	}(file)
	_, err = file.Write(data)
	LogError(err)
	return err == nil
}
