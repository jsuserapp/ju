//go:build !windows

package sv

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"text/template"

	"github.com/gookit/color"
	"github.com/jsuserapp/ju"
)

// linuxService Linux 服务器结构
//
// 应用运行参数
//
// ./appname (无参数) 执行服务的主功能
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
//
// ./appname help 或者 ? 打印命令帮助信息
type linuxService struct {
	name string
	desc string
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
ExecStart={{.Path}}
Restart=always
RestartSec=5s

[Install]
WantedBy=multi-user.target
`

// StartService 在Main函数中调用，Start 会获取应用的启动参数，如果没有启动参数，这个函数会打印命令的帮助信息
//
// @name 服务的名称，为了避免混淆，通常设置为和可执行文件同名，name 不能为空，否则函数会返回 false。
//
// @desc 服务的功能描述，这个值可以为空，会默认设置为服务名称
//
// @run 服务的业务函数，当启动参数设置为 run 时，run才会被调用。
//
// @return 名称为空时，函数会返回 false，其它情况都会返回true
//
// 在windows下，仍然可以使用 run 参数来运行程序，这通常方便调试，但是不能执行其它 Linux 对服务的操作
// noinspection GoUnusedExportedFunction
func StartService(name, desc string, run func()) bool {
	if name == "" {
		ju.LogRed("服务名称不能为空")
		return false
	}
	if desc == "" {
		desc = name + " service"
	}
	// 获取命令行参数
	args := os.Args
	if len(args) < 2 {
		if run != nil {
			run()
		} else {
			ju.LogRed("服务的业务函数为空")
		}
		return true
	}

	ls := &linuxService{name: name, desc: desc}
	action := args[1]

	switch action {
	case CmdHelp, "?":
		ls.printHelp()
	case CmdInstall:
		ls.installService()
	case CmdUninstall:
		ls.uninstallService()
	case CmdStart:
		ls.systemctl(CmdStart)
	case CmdStop:
		ls.systemctl(CmdStop)
	case CmdRestart:
		ls.systemctl(CmdRestart)
	case CmdStatus:
		ls.systemctl(CmdStatus)
	default:
		ls.printHelp()
	}
	return true
}

func (ls *linuxService) installService() {
	// 1. 获取当前程序的绝对路径
	exePath, err := os.Executable()
	if err != nil {
		log.Fatalf("获取程序路径失败: %v", err)
	}
	exePath, _ = filepath.Abs(exePath)
	workDir := filepath.Dir(exePath)

	// 2. 准备 Service 文件内容
	serviceContent := struct {
		Desc    string
		Path    string
		WorkDir string
	}{
		Desc:    ls.desc,
		Path:    exePath,
		WorkDir: workDir,
	}

	// 3. 定义 Service 文件路径 (/etc/systemd/system/juserv.service)
	serviceFile := fmt.Sprintf("/etc/systemd/system/%s.service", ls.name)

	// 4. 写入文件
	f, err := os.Create(serviceFile)
	if err != nil {
		ju.LogRed(fmt.Sprintf("无法创建服务文件 (请确保使用 sudo 运行): %v", err))
	}
	defer func(f *os.File) {
		_ = f.Close()
	}(f)

	tmpl := template.Must(template.New("service").Parse(serviceTemplate))
	err = tmpl.Execute(f, serviceContent)
	if err != nil {
		ju.LogRed(fmt.Sprintf("写入服务文件失败: %v", err))
		return
	}

	ju.LogGreen(fmt.Sprintf("服务文件已创建: %s", serviceFile))

	// 5. 重载 systemd 并启用服务
	ls.runCmd("systemctl", cmdDaemonReload)
	ls.runCmd("systemctl", cmdEnable, ls.name)
	ju.LogGreen("服务安装成功！可以使用 'juserv start' 启动。")
}

func (ls *linuxService) uninstallService() {
	// 1. 停止服务
	ls.runCmd("systemctl", CmdStop, ls.name)
	// 2. 禁用服务
	ls.runCmd("systemctl", cmdDisable, ls.name)

	// 3. 删除 Service 文件
	serviceFile := fmt.Sprintf("/etc/systemd/system/%s.service", ls.name)
	err := os.Remove(serviceFile)
	if err != nil {
		ju.LogRed(fmt.Sprintf("警告: 删除服务文件失败 (可能不存在): %v", err))
	} else {
		ju.LogGreen(fmt.Sprintf("已删除服务文件: %s", serviceFile))
	}

	// 4. 重载
	ls.runCmd("systemctl", cmdDaemonReload)
	ju.LogGreen("服务卸载完成。")
}

// systemctl 封装了启动、停止、重启命令
func (ls *linuxService) systemctl(action string) {
	ju.LogGreen(fmt.Sprintf("正在执行: systemctl %s %s ...", action, ls.name))
	if ls.runCmd("systemctl", action, ls.name) {
		ju.LogGreen("执行成功")
	}
}

// 辅助函数：运行 shell 命令
func (ls *linuxService) runCmd(name string, args ...string) bool {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	return ju.LogSucceed(err)
}

func (ls *linuxService) printHelp() {
	fmt.Printf("用法: %s [command]\n", ls.name)
	fmt.Println("命令:")
	fmt.Printf("  无参数\t直接运行，业务函数会被调用，执行服务的主功能\n")
	fmt.Printf("  %s\t安装系统服务 (需要 root 权限)\n", color.Blue.Sprintf("%s", CmdInstall))
	fmt.Printf("  %s\t卸载系统服务 (需要 root 权限)\n", color.Blue.Sprintf("%s", CmdUninstall))
	fmt.Printf("  %s\t\t启动服务\n", color.Blue.Sprintf("%s", CmdStart))
	fmt.Printf("  %s\t\t停止服务\n", color.Blue.Sprintf("%s", CmdStop))
	fmt.Printf("  %s\t重启服务\n", color.Blue.Sprintf("%s", CmdRestart))
	fmt.Printf("  %s\t查看状态\n", color.Blue.Sprintf("%s", CmdStatus))
	fmt.Printf("  %s 或 %s\t\t打印调用说明\n", color.Blue.Sprintf("%s", CmdHelp), color.Blue.Sprintf("%s", "?"))
}
