package main

import (
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"golang.org/x/sys/windows/svc"
	"gopkg.in/natefinch/lumberjack.v2"
)

const (
	LOG_DIR = `C:\Logs`
)

type serviceWrapper struct {
	executable string
	logger     *log.Logger
}

func (s *serviceWrapper) Execute(args []string, r <-chan svc.ChangeRequest, changes chan<- svc.Status) (ssec bool, errno uint32) {
	const cmdsAccepted = svc.AcceptStop | svc.AcceptShutdown
	changes <- svc.Status{State: svc.StartPending}

	//Start the wrapped executable
	cmd := exec.Command(s.executable)
	cmd.Stdout = s.logger.Writer()
	cmd.Stderr = s.logger.Writer()

	if err := cmd.Start(); err != nil {
		s.logger.Printf("Failed to start executable: %v", err)
		return false, 1
	}

	changes <- svc.Status{State: svc.Running, Accepts: cmdsAccepted}

	//TODO: Here maybe we could writhe to the eventlog that we started?

	go func() {
		if err := cmd.Wait(); err != nil {
			s.logger.Printf("Executable exited with error: %v", err)
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
					s.logger.Printf("Failed to kill Process: %v", err)
				}
				return false, 0
			default:
				s.logger.Printf("Unexpected control request: %v", c)
			}
		}
	}
}

func main() {
	if len(os.Args) < 2 {
		log.Fatalf("Usage: %s <executable>", os.Args[0])
	}
	//TODO: we should definetly check that we are running as a service

	executable := os.Args[1]
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
		logger:     logger,
	}

	svcName := strings.TrimSuffix(exeName, filepath.Ext(exeName))
	if err := svc.Run(svcName, svcConfig); err != nil {
		logger.Fatalf("Service failed: %v", err)
	}
}
