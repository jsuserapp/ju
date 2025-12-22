package ju

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"text/template"

	"github.com/gookit/color"
)

// LinuxService Linux 服务器结构
//
// 应用运行参数
//
// ./appname (无参数) 打印命令帮助信息
//
// ./appname install 安装服务
//
// ./appname uninstall 卸载服务
//
// ./appname start 或者 systemctl start appname 启动服务
//
// ./appname restart 或者 systemctl restart appname 重启服务
//
// ./appname stop 或者 systemctl stop appname 停止服务
//
// ./appname status 或者 systemctl status appname 查看服务状态
type LinuxService struct {
	name    string
	desc    string
	workDir string
}

// NewLinuxService 创建一个 Linux 的服务对象，然后运行 Start 函数执行服务的相关功能
//
// @name 服务的名称，为了避免混淆，通常设置为和可执行文件同名，name 不能为空，否则函数会返回 nil。
//
// @desc 服务的功能描述，这个值可以为空，会默认设置为服务名称
//
// @workDir 这个参数默认为空，除非你需要设置一个特殊的工作目录，设置为空时，工作目录是可执行文件所在的目录
//
// @return 只有名称为空时，函数会返回 nil，其它情况都会返回有效的 LinuxService 对象
// noinspection GoUnusedExportedFunction
func NewLinuxService(name, desc, workDir string) *LinuxService {
	if name == "" {
		LogRed("服务名称不能为空")
		return nil
	}
	if desc == "" {
		desc = name + " service"
	}
	return &LinuxService{name: name, desc: desc, workDir: workDir}
}

// serviceTemplate 是 systemd 配置文件的模板
// {{.Path}} 是可执行文件的绝对路径
// {{.WorkDir}} 是工作目录
const serviceTemplate = `[Unit]
Description={{.Desc}}
After=network.target

[Service]
Type=simple
WorkingDirectory={{.WorkDir}}
ExecStart={{.Path}} run
Restart=always
RestartSec=5s

[Install]
WantedBy=multi-user.target
`

// Start 在Main函数中调用，Start 会获取应用的启动参数，如果没有启动参数，这个函数会打印命令的帮助信息
//
// @run 服务的业务函数，当启动参数设置为 run 时，run才会被调用。
func (ls *LinuxService) Start(run func()) {
	// 获取命令行参数
	args := os.Args
	if len(args) < 2 {
		ls.printHelp()
		return
	}

	action := args[1]

	switch action {
	case "run":
		if run != nil {
			run()
		} else {
			LogRed("服务的业务函数为空")
		}
	case "install":
		ls.installService()
	case "uninstall":
		ls.uninstallService()
	case "start":
		ls.systemctl("start")
	case "stop":
		ls.systemctl("stop")
	case "restart":
		ls.systemctl("restart")
	case "status":
		ls.systemctl("status")
	default:
		ls.printHelp()
	}
}

func (ls *LinuxService) installService() {
	// 1. 获取当前程序的绝对路径
	exePath, err := os.Executable()
	if ls.workDir == "" {
		if err != nil {
			log.Fatalf("获取程序路径失败: %v", err)
		}
		exePath, _ = filepath.Abs(exePath)
		ls.workDir = filepath.Dir(exePath)
	}

	// 2. 准备 Service 文件内容
	serviceContent := struct {
		Desc string
		Path string
		Dir  string
	}{
		Desc: ls.desc,
		Path: exePath,
		Dir:  ls.workDir,
	}

	// 3. 定义 Service 文件路径 (/etc/systemd/system/juserv.service)
	serviceFile := fmt.Sprintf("/etc/systemd/system/%s.service", ls.name)

	// 4. 写入文件
	f, err := os.Create(serviceFile)
	if err != nil {
		LogRed(fmt.Sprintf("无法创建服务文件 (请确保使用 sudo 运行): %v", err))
	}
	defer func(f *os.File) {
		_ = f.Close()
	}(f)

	tmpl := template.Must(template.New("service").Parse(serviceTemplate))
	err = tmpl.Execute(f, serviceContent)
	if err != nil {
		LogRed(fmt.Sprintf("写入服务文件失败: %v", err))
		return
	}

	LogGreen(fmt.Sprintf("服务文件已创建: %s\n", serviceFile))

	// 5. 重载 systemd 并启用服务
	ls.runCmd("systemctl", "daemon-reload")
	ls.runCmd("systemctl", "enable", ls.name)
	LogGreen("服务安装成功！可以使用 'juserv start' 启动。")
}

func (ls *LinuxService) uninstallService() {
	// 1. 停止服务
	ls.runCmd("systemctl", "stop", ls.name)
	// 2. 禁用服务
	ls.runCmd("systemctl", "disable", ls.name)

	// 3. 删除 Service 文件
	serviceFile := fmt.Sprintf("/etc/systemd/system/%s.service", ls.name)
	err := os.Remove(serviceFile)
	if err != nil {
		LogRed(fmt.Sprintf("警告: 删除服务文件失败 (可能不存在): %v", err))
	} else {
		LogGreen(fmt.Sprintf("已删除服务文件: %s\n", serviceFile))
	}

	// 4. 重载
	ls.runCmd("systemctl", "daemon-reload")
	LogGreen("服务卸载完成。")
}

// systemctl 封装了启动、停止、重启命令
func (ls *LinuxService) systemctl(action string) {
	LogGreen(fmt.Sprintf("正在执行: systemctl %s %s ...", action, ls.name))
	if ls.runCmd("systemctl", action, ls.name) {
		LogGreen("执行成功")
	}
}

// 辅助函数：运行 shell 命令
func (ls *LinuxService) runCmd(name string, args ...string) bool {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	return LogSucceed(err)
}

func (ls *LinuxService) printHelp() {
	fmt.Printf("用法: %s [command]\n", ls.name)
	fmt.Println("命令:")
	fmt.Printf("  %s\t安装系统服务 (需要 root 权限)\n", color.Blue.Sprintf("%s", "install"))
	fmt.Printf("  %s\t卸载系统服务 (需要 root 权限)\n", color.Blue.Sprintf("%s", "uninstall"))
	fmt.Printf("  %s\t\t启动服务\n", color.Blue.Sprintf("%s", "start"))
	fmt.Printf("  %s\t\t停止服务\n", color.Blue.Sprintf("%s", "stop"))
	fmt.Printf("  %s\t重启服务\n", color.Blue.Sprintf("%s", "restart"))
	fmt.Printf("  %s\t查看状态\n", color.Blue.Sprintf("%s", "status"))
	fmt.Printf("  %s\t\t前台运行主程序 (通常由 systemd 调用)\n", color.Blue.Sprintf("%s", "run"))
}
