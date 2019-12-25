package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"
)

var systemStop = false

func startExternalServices(ctx context.Context, wg *sync.WaitGroup) error {
	restartTime := 5 * time.Second

	//hostapd
	wg.Add(1)
	go func() {
		defer wg.Done()
		log := logrus.WithFields(logrus.Fields{"origin": "hostapd"})
		defer log.Info("hostapd exited")

		var err error

		for ctx.Err() == nil && !systemStop {
			log.Info("hostapd started")

			cmd := exec.CommandContext(ctx, "/etc/hostapd", filepath.Join("/etc", "hostapd", "hostapd.conf"))

			err = cmd.Start()
			if err != nil {
				log.Fatal("failed to start hostapd: " + err.Error())
			}
			cmd.Wait()
			log.Info("terminated")
			select {
			case <-ctx.Done():
			case <-time.After(restartTime):
			}
			if ctx.Err() == nil && !systemStop {
				log.Error("restart")
			}
		}
	}()

	//dhcpd
	wg.Add(1)
	go func() {
		defer wg.Done()
		log := logrus.WithFields(logrus.Fields{"origin": "dhcpd"})
		defer log.Info("dhcpd exited")

		var err error

		for ctx.Err() == nil && !systemStop {
			log.Info("dhcpd started")

			cmd := exec.CommandContext(ctx, "/etc/dhcpd", filepath.Join("/etc", "dhcpd", "dhcpd.conf"))

			err = cmd.Start()
			if err != nil {
				log.Fatal("failed to start dhcpd: " + err.Error())
			}
			cmd.Wait()
			log.Info("terminated")
			select {
			case <-ctx.Done():
			case <-time.After(restartTime):
			}
			if ctx.Err() == nil && !systemStop {
				log.Error("restart")
			}
		}
	}()

	return nil
}

func main() {
	var wg1 sync.WaitGroup
	ctxBackground, cancelBackground := context.WithCancel(context.Background())
	defer cancelBackground()

	if err := startExternalServices(ctxBackground, &wg1); err != nil {
		log.Fatal(err)
		return
	}
	startExternalServices(ctxBackground, &wg1)
	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)
	signal.Notify(quit, syscall.SIGTERM)
	srv := &http.Server{
		Addr: ":80",
	}

	go func() {
		srv.ListenAndServe()
	}()

	<-quit
	cancelBackground()
	systemStop = true
	ctxShutdown, cancelShutdown := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelShutdown()
	srv.Shutdown(ctxShutdown)
}
