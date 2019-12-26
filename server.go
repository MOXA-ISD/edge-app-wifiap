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
			var cmd *exec.Cmd

			// set ip address
			// link up
			cmd = exec.CommandContext(ctx, "ip", "addr", "replace", "10.0.0.1/24", "dev", "wlan0")
			cmd.Run()
			cmd = exec.CommandContext(ctx, "ip", "link", "set", "wlan0", "up")
			cmd.Run()

			cmd = exec.CommandContext(ctx, "/usr/sbin/hostapd", filepath.Join("/etc", "hostapd", "hostapd.conf"))

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

			cmd := exec.CommandContext(ctx, "/usr/sbin/udhcpd", "-f", filepath.Join("/etc", "dhcpd", "dhcpd.conf"))

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
	var (
		wg1 sync.WaitGroup
		cmd *exec.Cmd
	)
	ctx, cancelBackground := context.WithCancel(context.Background())
	defer cancelBackground()

	if err := startExternalServices(ctx, &wg1); err != nil {
		log.Fatal(err)
		return
	}

	// iptables -t filter -A INPUT -p udp -m udp --dport 67 -j ACCEPT
	// delete
	cmd = exec.CommandContext(ctx, "iptables", "-t", "filter", "-D", "INPUT", "-p", "udp", "-m", "udp", "--dport", "67", "-j", "ACCEPT")
	cmd.Run()
	// append
	cmd = exec.CommandContext(ctx, "iptables", "-t", "filter", "-A", "INPUT", "-p", "udp", "-m", "udp", "--dport", "67", "-j", "ACCEPT")
	cmd.Run()
	// iptables -t nat -A POSTROUTING -o eth0 -j MASQUERADE
	// delete
	cmd = exec.CommandContext(ctx, "iptables", "-t", "nat", "-D", "POSTROUTING", "-s", "10.0.0.1/24", "-o", "eth0", "-j", "MASQUERADE")
	cmd.Run()
	// append
	cmd = exec.CommandContext(ctx, "iptables", "-t", "nat", "-A", "POSTROUTING", "-s", "10.0.0.1/24", "-o", "eth0", "-j", "MASQUERADE")
	cmd.Run()
	// iptables -A FORWARD -i eth0 -o wlan0 -m state --state RELATED,ESTABLISHED -j ACCEPT
	// delete
	cmd = exec.CommandContext(ctx, "iptables", "-D", "FORWARD", "-i", "eth0", "-o", "wlan0", "-m", "state", "--state", "RELATED,ESTABLISHED", "-j", "ACCEPT")
	cmd.Run()
	// append
	cmd = exec.CommandContext(ctx, "iptables", "-A", "FORWARD", "-i", "eth0", "-o", "wlan0", "-m", "state", "--state", "RELATED,ESTABLISHED", "-j", "ACCEPT")
	cmd.Run()
	// iptables -A FORWARD -i wlan0 -o eth0 -j ACCEPT
	// delete
	cmd = exec.CommandContext(ctx, "iptables", "-D", "FORWARD", "-i", "wlan0", "-o", "eth0", "-j", "ACCEPT")
	cmd.Run()
	// append
	cmd = exec.CommandContext(ctx, "iptables", "-A", "FORWARD", "-i", "wlan0", "-o", "eth0", "-j", "ACCEPT")
	cmd.Run()

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
