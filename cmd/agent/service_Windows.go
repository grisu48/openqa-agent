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

	for {
		select {
		case c := <-r:
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
	}

	status <- svc.Status{State: svc.StopPending}
	return false, 1
}

func RunService() error {
	return svc.Run("openqa-agent", &service{})
}
