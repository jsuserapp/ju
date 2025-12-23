package sv

const (
	CmdInstall   = "install"
	CmdUninstall = "uninstall"
	CmdStart     = "start"
	CmdStop      = "stop"
	CmdHelp      = "help"

	//CmdRestart Linux 专用，Win服务这个功能由系统自动完成
	CmdRestart = "restart"
	//CmdStatus Linux 专用
	CmdStatus = "status"

	cmdDisable      = "disable"
	cmdEnable       = "enable"
	cmdDaemonReload = "daemon-reload"

	cmdPause    = "pause"
	cmdContinue = "continue"
)
