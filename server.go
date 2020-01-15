package main

import (
	"context"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type Configuration struct {
	Mac        string `json:"mac,omitempty"`
	Ip         string `json:"ip"`
	HostName   string `json:"hostname"`
	ExpireTime string `json:"expiretime"`
}

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

	// iptables -t nat -A POSTROUTING -o eth2 -j MASQUERADE
	// delete
	cmd = exec.CommandContext(ctx, "iptables", "-t", "nat", "-D", "POSTROUTING", "-s", "10.0.0.1/24", "-o", "eth2", "-j", "MASQUERADE")
	cmd.Run()
	// append
	cmd = exec.CommandContext(ctx, "iptables", "-t", "nat", "-A", "POSTROUTING", "-s", "10.0.0.1/24", "-o", "eth2", "-j", "MASQUERADE")
	cmd.Run()
	// iptables -A FORWARD -i eth2 -o wlan0 -m state --state RELATED,ESTABLISHED -j ACCEPT
	// delete
	cmd = exec.CommandContext(ctx, "iptables", "-D", "FORWARD", "-i", "eth2", "-o", "wlan0", "-m", "state", "--state", "RELATED,ESTABLISHED", "-j", "ACCEPT")
	cmd.Run()
	// append
	cmd = exec.CommandContext(ctx, "iptables", "-A", "FORWARD", "-i", "eth2", "-o", "wlan0", "-m", "state", "--state", "RELATED,ESTABLISHED", "-j", "ACCEPT")
	cmd.Run()
	// iptables -A FORWARD -i wlan0 -o eth2 -j ACCEPT
	// delete
	cmd = exec.CommandContext(ctx, "iptables", "-D", "FORWARD", "-i", "wlan0", "-o", "eth2", "-j", "ACCEPT")
	cmd.Run()
	// append
	cmd = exec.CommandContext(ctx, "iptables", "-A", "FORWARD", "-i", "wlan0", "-o", "eth2", "-j", "ACCEPT")
	cmd.Run()

	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)
	signal.Notify(quit, syscall.SIGTERM)

	router := gin.Default()
	router.GET("/api/v1/dhcp-leases", func(c *gin.Context) {

		initcmd := exec.CommandContext(ctx, "touch", filepath.Join("/var", "/lib", "/misc", "udhcpd.leases"))
		initcmd.Run()

		pid, err := ioutil.ReadFile("/var/run/udhcpd.pid")
		if err != nil {
			logrus.Errorf("error1: %v", err.Error())
		}
		logrus.Errorf("pid: %v", pid)
		reflashcmd := exec.CommandContext(ctx, "/bin/busybox", "kill", "-SIGUSR1", strings.TrimSpace(string(pid)))
		b, err := reflashcmd.CombinedOutput()

		if err != nil {
			logrus.Errorf("error2: %v", err.Error())
		}
		logrus.Errorf("%v", string(b))

		var str string
		var dump string
		var clientData []Configuration = []Configuration{}
		re := regexp.MustCompile(`(?m)Station ([\S]+)`)

		cmd = exec.CommandContext(ctx, "iw", "dev", "wlan0", "station", "dump")
		m, err := cmd.CombinedOutput()
		if err == nil {
			str = string(m)
		} else {
			str = err.Error()
		}

		cmd = exec.CommandContext(ctx, "dumpleases")
		d, err := cmd.CombinedOutput()
		if err == nil {
			dump = string(d)
		} else {
			dump = err.Error()
		}

		for _, match := range re.FindAllString(str, -1) {
			mac := strings.Fields(match)
			remac := regexp.MustCompile(mac[1] + ".*")

			for _, matches := range remac.FindAllString(dump, -1) {
				data := strings.Fields(matches)
				logrus.Errorf("%v", data)
				logrus.Errorf("%v", len(data))
				writedata := make([]string, 4)
				if len(data) < 4 {
					writedata[0] = data[0]
					writedata[1] = data[1]
					writedata[2] = ""
					writedata[3] = data[2]
				} else {
					writedata[0] = data[0]
					writedata[1] = data[1]
					writedata[2] = data[2]
					writedata[3] = data[3]
				}
				logrus.Errorf("%v", writedata)
				config := Configuration{
					Mac:        writedata[0],
					Ip:         writedata[1],
					HostName:   writedata[2],
					ExpireTime: writedata[3]}

				clientData = append(clientData, config)
			}
		}

		c.JSON(http.StatusOK, gin.H{"data": clientData})
	})
	srv := &http.Server{
		Addr:    ":12345",
		Handler: router,
	}

	go func() {
		srv.ListenAndServe()
		// TODO error message
	}()

	<-quit
	cancelBackground()
	systemStop = true
	ctxShutdown, cancelShutdown := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelShutdown()
	srv.Shutdown(ctxShutdown)
}
