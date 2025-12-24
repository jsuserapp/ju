package ju

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

type bufWriter struct {
	mu     sync.Mutex
	path   string
	writer *bufio.Writer
}

func (bw *bufWriter) Write(p []byte) (n int, err error) {
	bw.mu.Lock()
	defer bw.mu.Unlock()

	// 关键：使用 O_APPEND 和 O_CREATE 模式打开文件。
	// O_CREATE: 如果文件不存在，就自动创建它。这处理了文件被删除的情况。
	// O_APPEND: 保证每次写入都在文件的当前末尾。这处理了文件被截断的情况。
	file, err := os.OpenFile(bw.path, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return 0, err
	}
	defer func() {
		err = file.Close() // 立即关闭文件句柄，释放文件锁
		OutputErrorTrace(err, 0)
	}()

	// 执行写入操作
	return file.Write(p)
}

// fileLogWriter 是一个健壮的日志写入器。
// 它结合了内存缓冲以提高性能，并通过定期重新打开文件来对外部的文件删除和修改操作免疫。
type fileLogWriter struct {
	mu         sync.Mutex
	once       sync.Once
	done       chan struct{}
	name       string
	folder     string
	bufSize    int
	writerList map[string]*bufWriter
}

// newFileLogWriter 创建一个新的 fileLogWriter 实例。
func newFileLogWriter(folder string, writeInterval time.Duration, bufSize int) *fileLogWriter {
	if writeInterval <= 0 {
		writeInterval = 5 * time.Second
	}
	if bufSize <= 0 {
		bufSize = 4096 // 4KB 默认缓冲区大小
	}

	exePath, exeName := getExeDirectoryAndName()
	if folder == "" {
		folder = exePath
	}
	CreateFolder(folder)

	flw := &fileLogWriter{
		writerList: map[string]*bufWriter{},
		done:       make(chan struct{}),
		folder:     folder,
		name:       exeName,
		bufSize:    bufSize,
	}

	// 启动后台的定时刷新任务
	flw.start(writeInterval)

	return flw
}

// getExeDirectoryAndName 获取当前可执行文件所在的文件夹和名称，windows下这个名称是去掉尾部的 .exe 的
// 因为win下的路径分隔符(\)和Linux(/)下是不一样的，这里还处理成Linux的格式(/)
func getExeDirectoryAndName() (path, name string) {
	exePath, err := os.Executable()
	if OutputErrorTrace(err, 0) {
		return
	}
	path = filepath.Dir(exePath)
	path = filepath.ToSlash(path)
	name = filepath.Base(exePath)
	switch runtime.GOOS {
	case "windows":
		nameLen := len(name)
		if nameLen > 4 {
			name = name[:nameLen-4]
		}
	}
	return
}

// GetServicePath 获取服务的可执行文件路径
func GetServicePath(name string) string {
	switch runtime.GOOS {
	case "windows":
		return "."
	}
	fn := "/etc/systemd/system/" + name + ".service"
	data := ReadFile(fn)
	if data == nil {
		return ""
	}
	str := string(data)
	flag := "ExecStart="
	pos := strings.Index(str, flag)
	if pos == -1 {
		return ""
	}
	str = str[pos+len(flag):]
	pos = strings.Index(str, "\n")
	if pos != -1 {
		str = str[:pos]
	}
	pos = strings.LastIndex(str, "/")
	if pos != -1 {
		str = str[:pos]
	}
	return str
}

// Write 实现了 io.Writer 接口。这是该写入器的核心逻辑，只在缓冲区 Flush 时被调用。
func (w *fileLogWriter) getWriter(folder, tag string, bufSize int) *bufWriter {
	w.mu.Lock()
	defer w.mu.Unlock()
	bw := w.writerList[tag]
	if bw == nil {
		bw = &bufWriter{
			path: filepath.Join(folder, tag),
		}
		bw.writer = bufio.NewWriterSize(bw, bufSize)
		w.writerList[tag] = bw
	}
	return bw
}
func (w *fileLogWriter) makeLine(color, trace, log string) []byte {
	createdAt := time.Now().Format("2006-01-02 15:04:05.000")
	line := fmt.Sprintf("%s\t%s\t%s\t%s\n", createdAt, color, trace, log)
	return []byte(line)
}
func (w *fileLogWriter) saveLog(tag, color, trace, log string) bool {
	tag = w.getTagName(tag)
	data := w.makeLine(color, trace, log)
	wr := w.getWriter(w.folder, tag, w.bufSize)
	_, err := wr.Write(data)
	return OutputErrorTrace(err, 0)
}
func (w *fileLogWriter) getTagName(tag string) string {
	if tag == "" {
		tag = w.name + ".log"
	} else {
		tag = w.name + "_" + tag + ".log"
	}
	return tag
}
func (w *fileLogWriter) clear(tag string) {
	tag = w.getTagName(tag)
	path := filepath.Join(w.folder, tag)
	if !FileExist(path) {
		//如果文件不存在，不会创建它
		return
	}
	file, err := os.Create(path)
	if OutputErrorTrace(err, 0) {
		return
	}
	_ = file.Close()
}

// ReadLastLog 从日志文件末尾读取大约指定字节数的内容，并按行返回。
// 这种方法性能很高，因为它只读取文件的尾部一小部分。
func (w *fileLogWriter) ReadLastLog(tag string, bytesToRead int64) []string {
	if bytesToRead <= 0 {
		bytesToRead = 4096
	}

	tag = w.getTagName(tag)
	path := filepath.Join(w.folder, tag)
	file, err := os.Open(path)
	if OutputErrorTrace(err, 0) {
		return nil
	}
	defer func() {
		_ = file.Close()
	}()

	// 获取文件大小，以计算读取的起始位置
	stat, err := file.Stat()
	if OutputErrorTrace(err, 0) {
		return nil
	}
	fileSize := stat.Size()

	if bytesToRead > fileSize {
		bytesToRead = fileSize
	}
	offset := fileSize - bytesToRead

	var lines []string
	//var buffer bytes.Buffer // 用于拼接从文件末尾读出的字节块

	// 移动文件指针到指定位置
	_, err = file.Seek(offset, io.SeekStart)
	if OutputErrorTrace(err, 0) {
		return nil
	}

	scanner := bufio.NewScanner(file)
	// 如果我们不是从文件开头读取 (offset > 0)，那么我们读到的第一行
	// 极有可能是被截断的、不完整的。我们需要先调用一次 Scan() 来读取并丢弃它。
	if offset > 0 {
		scanner.Scan() // 读取并丢弃第一个潜在的部分行
	}
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	err = scanner.Err()
	if OutputErrorTrace(err, 0) {
		return nil
	}

	return lines
}

// Flush 将缓冲区的内容写入磁盘。这是一个线程安全的操作。
func (w *fileLogWriter) Flush() {
	w.mu.Lock()
	defer w.mu.Unlock()
	for _, wr := range w.writerList {
		err := wr.writer.Flush()
		OutputErrorTrace(err, 1)
	}
}

// Close 优雅地关闭日志写入器。
// 它会停止后台的刷新协程，并确保所有缓冲的日志都被写入文件。
// 这个方法是幂等的，可以安全地多次调用。
func (w *fileLogWriter) Close() {
	w.once.Do(func() {
		// 发送停止信号给后台协程
		close(w.done)
		// 执行最后一次刷新，确保所有剩余的日志都被写入磁盘
		w.Flush()
	})
}

// start 启动一个定时任务，定期将缓冲区的内容刷入磁盘。
func (w *fileLogWriter) start(interval time.Duration) {
	ticker := time.NewTicker(interval)

	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				w.Flush()
			case <-w.done:
				// 收到关闭信号，退出协程
				return
			}
		}
	}()
}

type FileLogDb struct {
	db *fileLogWriter
}

// CreateFileLogDb 生成一个日志文件对象，日志文件只有在写入日志时才会打开日志文件，所以可以对日志文件使用其它应用进行操作，不会互相干扰。
//
// folder: 日志保存目录，传空串则日志在当前可执行文件的目录（非工作目录）
//
// writeInterval: 日志保存到文件的间隔，出于性能考虑，不是每条日志都即时的写入文件，而是间隔保存（如果有日志的话）。缺省值是 5 秒（传0）
//
// bufSize: 日志缓存缓存，默认值 4096 字节（传0），当缓存满了之后会执行一次写入文件操作。
func CreateFileLogDb(folder string, writeInterval time.Duration, bufSize int) *FileLogDb {
	ldb := &FileLogDb{
		db: newFileLogWriter(folder, writeInterval, bufSize),
	}
	return ldb
}

// CloseFileLogDb 关闭日志对象，这个操作是必要的，但是通常可能不执行这个关闭动作，日志也能成功写入日志文件，
// 这是因为操作系统优雅的处理了应用进程退出后资源关闭，但是依靠系统关闭不一定 100% 可靠，通常建议在应用退出时执行关闭操作。
func CloseFileLogDb(db *FileLogDb) {
	if db != nil && db.db != nil {
		db.db.Close()
	}
}

// DeleteLog 对于文件日志来说，这个函数什么都不做
func (mdb *FileLogDb) DeleteLog(tag string, id int64) {
}

// DeleteTagLogs 对于文件日志来说，这个函数什么都不做
func (mdb *FileLogDb) DeleteTagLogs(tag, before string) int64 {
	return 0
}

// DeleteLogs 对于文件日志来说，这个函数什么都不做
func (mdb *FileLogDb) DeleteLogs(before string) int64 {
	return 0
}

// GetLastLogs 获取最新的 log，bytes 是读取的字节数.
// 这个字节数只是参考，因为它不一定是完整的日志行，所以对于不完整的第一行会抛弃（实际上，即使是完整的行，它也会抛弃第一行）。
func (mdb *FileLogDb) GetLastLogs(tag string, bytes int) (logs []*LogInfo) {
	lines := mdb.db.ReadLastLog(tag, int64(bytes))
	for _, line := range lines {
		params := strings.Split(line, "\t")
		if len(params) < 4 {
			//不合法的日志行
			continue
		}
		li := LogInfo{
			Id:        0,
			CreatedAt: params[0],
			Color:     params[1],
			Trace:     params[2],
			Log:       params[3],
		}
		logs = append(logs, &li)
	}
	return
}

// ClearTagLogs 清空指定 tag 的日志，默认日志对应的 tag 是空字符串
func (mdb *FileLogDb) ClearTagLogs(tag string) int64 {
	mdb.db.clear(tag)
	return 0
}
func (mdb *FileLogDb) saveLog(tag, color, trace, log string) bool {
	return mdb.db.saveLog(tag, color, trace, log)
}
