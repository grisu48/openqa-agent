//go:build windows
// +build windows

package main

import (
	"log"
	"os"

	"golang.org/x/sys/windows/svc"
)

// Windows Service
type service struct{}

func (srv *service) Execute(args []string, r <-chan svc.ChangeRequest, status chan<- svc.Status) (bool, uint32) {
	const cmdsAccepted = svc.AcceptStop | svc.AcceptShutdown

	status <- svc.Status{State: svc.StartPending}
	status <- svc.Status{State: svc.Running, Accepts: cmdsAccepted}

	for c := range r {
		switch c.Cmd {
		case svc.Interrogate:
			status <- c.CurrentStatus
		case svc.Stop, svc.Shutdown:
			log.Print("Service shutdown")
			os.Exit(1)
		default:
			log.Printf("Unsupported service control request #%d", c)
		}
	}

	status <- svc.Status{State: svc.StopPending}
	status <- svc.Status{State: svc.Stopped}
	return false, 0
}

func RunService() error {
	if isService, err := svc.IsWindowsService(); err != nil {
		return err
	} else if isService {
		return svc.Run("openqa-agent", &service{})
	}
	return nil
}
