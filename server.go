package main

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"
	"os/exec"
	"path/filepath"

	"github.com/sirupsen/logrus"
)

func startExternalServices(ctx context.Context, wg *sync.WaitGroup) error {
	restartTime := 5 * time.Second

	//hostapd
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer log.Info("hostapd exited")
		log := logrus.WithFields(logrus.Fields{"origin": "hostapd"})

		var err error

		for ctx.Err() == nil && !systemStop {
			log.Info("hostapd started")

			cmd := exec.CommandContext(ctx, "/etc/hostapd", filepath.Join(configs.DataDir, "hostapd", "hostapd.conf"))
			cmd.Dir = filepath.Join(configs.DataDir, "hostapd")

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
		defer log.Info("dhcpd exited")
		log := logrus.WithFields(logrus.Fields{"origin": "dhcpd"})

		var err error

		for ctx.Err() == nil && !systemStop {
			log.Info("dhcpd started")

			cmd := exec.CommandContext(ctx, "/etc/dhcpd", filepath.Join(configs.DataDir, "dhcpd", "dhcpd.conf"))
			cmd.Dir = filepath.Join(configs.DataDir, "dhcpd")

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

func main() {
	http.HandleFunc("/hello", hello)
	http.HandleFunc("/headers", headers)
	http.ListenAndServe(":80", nil)
}
