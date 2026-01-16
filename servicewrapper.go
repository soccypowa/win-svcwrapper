package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/eventlog"
	"gopkg.in/natefinch/lumberjack.v2"
)

const (
	LOG_DIR = `C:\Logs`
)

type serviceWrapper struct {
	executable string
	execArgs   []string
	svcName    string
	logger     *log.Logger
}

func (s *serviceWrapper) Execute(args []string, r <-chan svc.ChangeRequest, changes chan<- svc.Status) (ssec bool, errno uint32) {
	const cmdsAccepted = svc.AcceptStop | svc.AcceptShutdown

	elog, err := eventlog.Open(s.svcName)
	if err != nil {
		log.Printf("Failed to open eventlog: %v", err)
	}

	defer elog.Close()

	//Start the wrapped executable with arguments
	cmd := exec.Command(s.executable, s.execArgs...)
	cmd.Stdout = s.logger.Writer()
	cmd.Stderr = s.logger.Writer()

	if err := cmd.Start(); err != nil {
		elog.Error(1000, fmt.Sprintf("Failed to start service %s: %v", s.svcName, err))
		return false, 1
	}

	changes <- svc.Status{State: svc.Running, Accepts: cmdsAccepted}

	elog.Info(1000, fmt.Sprintf("%v started successfully.", s.svcName))

	go func() {
		if err := cmd.Wait(); err != nil {
			elog.Error(1000, fmt.Sprintf("Service: %s exited with error: %v", s.svcName, err))
		}
	}()

	for {
		select {
		case c := <-r:
			switch c.Cmd {
			case svc.Interrogate:
				changes <- c.CurrentStatus
			case svc.Stop, svc.Shutdown:
				changes <- svc.Status{State: svc.StopPending}
				if err := cmd.Process.Kill(); err != nil {
					//TODO: Could we do this a little nicer?
					elog.Error(1000, fmt.Sprintf("Failed to kill service: %s with error: %v", s.svcName, err))
				}
				elog.Info(1000, fmt.Sprintf("%v stoped successfully.", s.svcName))
				return false, 0
			default:
				elog.Info(1000, fmt.Sprintf("Unexpected control request to service: %s request: %v", s.svcName, c))
			}
		}
	}
}

func main() {
	if len(os.Args) < 2 {
		log.Fatalf("Usage: %s <executable> [args...]", os.Args[0])
	}
	//TODO: we should definetly check that we are running as a service
	isService, err := svc.IsWindowsService()
	if err != nil {
		log.Fatalf("Failed to check if running as a service: %v", err)
	}

	if isService {
		executable := os.Args[1]
		execArgs := []string{}
		if len(os.Args) > 2 {
			execArgs = os.Args[2:]
		}
		exeName := filepath.Base(executable)
		logFile := filepath.Join(LOG_DIR, strings.TrimSuffix(exeName, filepath.Ext(exeName))+".log")

		//Ensure that the directory exists
		if err := os.MkdirAll(LOG_DIR, 0755); err != nil {
			log.Fatalf("Failed to create log directory %s: %v", LOG_DIR, err)
		}

		//Setup the log rotation
		logWriter := &lumberjack.Logger{
			Filename:   logFile,
			MaxSize:    50, //MB
			MaxBackups: 10, //Files retained on disk
			MaxAge:     0,  //No age based deletion
			Compress:   false,
		}

		logger := log.New(logWriter, "", log.LstdFlags)

		svcConfig := &serviceWrapper{
			executable: executable,
			execArgs:   execArgs,
			svcName:    strings.TrimSuffix(exeName, filepath.Ext(exeName)),
			logger:     logger,
		}

		if err := svc.Run(svcConfig.svcName, svcConfig); err != nil {
			logger.Fatalf("Service failed: %v", err)
		}
	} else {
		log.Println("This program is designed to run as a Windows Service.")
	}
}
