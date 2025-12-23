//go:build windows

package sv

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gookit/color"
	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/eventlog"
	"golang.org/x/sys/windows/svc/mgr"
)

func printHelp(name string) {
	fmt.Printf("用法: %s [command]\n", name)
	fmt.Println("命令:") //
	fmt.Printf("  无参数\t直接运行，业务函数会被调用，执行服务的主功能\n")
	fmt.Printf("  %s\t安装系统服务 (需要 管理员 权限)\n", color.Blue.Sprintf("%s", CmdInstall))
	fmt.Printf("  %s\t卸载系统服务 (需要 管理员 权限)\n", color.Blue.Sprintf("%s", CmdUninstall))
	fmt.Printf("  %s\t\t启动服务\n", color.Blue.Sprintf("%s", CmdStart))
	fmt.Printf("  %s\t\t停止服务\n", color.Blue.Sprintf("%s", CmdStop))
	fmt.Printf("  %s 或 %s\t\t打印调用说明\n", color.Blue.Sprintf("%s", CmdHelp), color.Blue.Sprintf("%s", "?"))
}

func StartService(name, desc string, run func()) {
	inService, err := svc.IsWindowsService()
	if err != nil {
		log.Fatalf("failed to determine if we are running in service: %v", err)
	}
	if inService {
		_ = svc.Run(name, &winService{onRun: run})
		return
	}

	if len(os.Args) < 2 {
		if run != nil {
			run()
		}
		return
	}

	cmd := strings.ToLower(os.Args[1])
	switch cmd {
	case CmdHelp, "?":
		printHelp(name)
	case CmdInstall:
		err = installService(name, desc)
	case CmdUninstall:
		err = removeService(name)
	case CmdStart:
		err = startService(name)
	case CmdStop:
		err = controlService(name, svc.Stop, svc.Stopped)
	case cmdPause:
		err = controlService(name, svc.Pause, svc.Paused)
	case cmdContinue:
		err = controlService(name, svc.Continue, svc.Running)
	default:
		printHelp(name)
	}
	if err != nil {
		log.Fatalf("failed to %s %s: %v", cmd, name, err)
	}
}

type winService struct {
	onRun func()
}

func (m *winService) Execute(args []string, r <-chan svc.ChangeRequest, changes chan<- svc.Status) (ssec bool, errno uint32) {
	const cmdsAccepted = svc.AcceptStop | svc.AcceptShutdown | svc.AcceptPauseAndContinue
	changes <- svc.Status{State: svc.StartPending}
	go func() {
		if m.onRun != nil {
			m.onRun()
		}
	}()
	changes <- svc.Status{State: svc.Running, Accepts: cmdsAccepted}
loop:
	for {
		select {
		case c := <-r:
			switch c.Cmd {
			case svc.Interrogate:
				changes <- c.CurrentStatus
			case svc.Stop, svc.Shutdown:
				break loop
			case svc.Pause:
				changes <- svc.Status{State: svc.Paused, Accepts: cmdsAccepted}
			case svc.Continue:
				changes <- svc.Status{State: svc.Running, Accepts: cmdsAccepted}
			default:
			}
		}
	}
	changes <- svc.Status{State: svc.StopPending}
	return
}

func startService(name string) error {
	m, err := mgr.Connect()
	if err != nil {
		return err
	}
	defer func(m *mgr.Mgr) {
		_ = m.Disconnect()
	}(m)
	s, err := m.OpenService(name)
	if err != nil {
		return fmt.Errorf("could not access service: %v", err)
	}
	defer func(s *mgr.Service) {
		_ = s.Close()
	}(s)
	err = s.Start()
	if err != nil {
		return fmt.Errorf("could not start service: %v", err)
	}
	return nil
}

func controlService(name string, c svc.Cmd, to svc.State) error {
	m, err := mgr.Connect()
	if err != nil {
		return err
	}
	defer func(m *mgr.Mgr) {
		_ = m.Disconnect()
	}(m)
	s, err := m.OpenService(name)
	if err != nil {
		return fmt.Errorf("could not access service: %v", err)
	}
	defer func(s *mgr.Service) {
		_ = s.Close()
	}(s)
	status, err := s.Control(c)
	if err != nil {
		return fmt.Errorf("could not send control=%d: %v", c, err)
	}
	timeout := time.Now().Add(10 * time.Second)
	for status.State != to {
		if timeout.Before(time.Now()) {
			return fmt.Errorf("timeout waiting for service to go to state=%d", to)
		}
		time.Sleep(300 * time.Millisecond)
		status, err = s.Query()
		if err != nil {
			return fmt.Errorf("could not retrieve service status: %v", err)
		}
	}
	return nil
}
func exePath() (string, error) {
	prog := os.Args[0]
	p, err := filepath.Abs(prog)
	if err != nil {
		return "", err
	}
	fi, err := os.Stat(p)
	if err == nil {
		if !fi.Mode().IsDir() {
			return p, nil
		}
		err = fmt.Errorf("%s is directory", p)
	}
	if filepath.Ext(p) == "" {
		p += ".exe"
		fi, err := os.Stat(p)
		if err == nil {
			if !fi.Mode().IsDir() {
				return p, nil
			}
			err = fmt.Errorf("%s is directory", p)
		}
	}
	return "", err
}

func installService(name, desc string) error {
	exepath, err := exePath()
	if err != nil {
		return err
	}
	m, err := mgr.Connect()
	if err != nil {
		return err
	}
	defer func(m *mgr.Mgr) {
		_ = m.Disconnect()
	}(m)
	s, err := m.OpenService(name)
	if err == nil {
		_ = s.Close()
		return fmt.Errorf("service %s already exists", name)
	}
	s, err = m.CreateService(name, exepath, mgr.Config{DisplayName: desc}, "is", "auto-started")
	if err != nil {
		return err
	}
	defer func(s *mgr.Service) {
		_ = s.Close()
	}(s)
	err = eventlog.InstallAsEventCreate(name, eventlog.Error|eventlog.Warning|eventlog.Info)
	if err != nil {
		_ = s.Delete()
		return fmt.Errorf("SetupEventLogSource() failed: %s", err)
	}
	return nil
}

func removeService(name string) error {
	m, err := mgr.Connect()
	if err != nil {
		return err
	}
	defer func(m *mgr.Mgr) {
		_ = m.Disconnect()
	}(m)
	s, err := m.OpenService(name)
	if err != nil {
		return fmt.Errorf("service %s is not installed", name)
	}
	defer func(s *mgr.Service) {
		_ = s.Close()
	}(s)
	err = s.Delete()
	if err != nil {
		return err
	}
	err = eventlog.Remove(name)
	if err != nil {
		return fmt.Errorf("RemoveEventLogSource() failed: %s", err)
	}
	return nil
}
