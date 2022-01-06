package collector

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

var netappSecureSSH = promauto.NewGaugeVec(
	prometheus.GaugeOpts{
		Name: "netapp_secure_ssh",
		Help: "Is secure ssh enabled ?",
	},
	[]string{
		"host",
		"file",
	},
)

var netappHost string
var netappPort int

func RecordMetrics(username string, password string, netappH string, netappP int, interval int, knownHostsFile string) {
	netappHost = netappH
	netappPort = netappP

	go func() {
		for {
			client, err := connectToHost(username, password, knownHostsFile)
			if err != nil {
				log.Println(err)
				netappNetworkInterfaceStatus.Reset()
				netappNetworkInterfaceIsHome.Reset()
				netappNetworkPortLinkStatus.Reset()
				netappNetworkPortHealth.Reset()
				netappStorageDiskError.Reset()
				netappStorageAggregateStatus.Reset()
				netappStorageAggregateStatus.Reset()
				netappStorageAggregateUsableSize.Reset()
				netappStorageAggregatePhysicalSize.Reset()
				netappSystemHealth.Reset()
				netappVolumeStatus.Reset()
				netappVolumeUsedSize.Reset()
				netappVolumeAvailableSize.Reset()
				netappVolumeTotalSize.Reset()
				netappVServerStatus.Reset()
				time.Sleep(10 * time.Second)
				continue
			}
			getNetworkInterfaceStatus(client)
			getNetworkPortStatus(client)
			getStorageDiskErrors(client)
			getStorageAggregateStatus(client)
			getSystemHealth(client)
			getVolumeStatus(client)
			getVServerStatus(client)
			client.Close()
			time.Sleep(time.Duration(interval) * time.Second)
		}
	}()
}

func connectToHost(username string, password string, knownHostsFile string) (*ssh.Client, error) {
	sshConfig := &ssh.ClientConfig{
		User:    username,
		Auth:    []ssh.AuthMethod{ssh.Password(password)},
		Timeout: time.Second * 30,
	}
	sshConfig.HostKeyCallback = ssh.InsecureIgnoreHostKey()
	netappSecureSSH.WithLabelValues(netappHost, knownHostsFile).Set(0)
	if knownHostsFile != "" {
		callback, err := knownhosts.New(knownHostsFile)
		if err != nil {
			log.Fatalln(err)
		}
		sshConfig.HostKeyCallback = callback
		netappSecureSSH.WithLabelValues(netappHost, knownHostsFile).Set(1)
	}

	client, err := ssh.Dial("tcp", netappHost+":"+fmt.Sprint(netappPort), sshConfig)
	if err != nil {
		return nil, err
	}

	return client, nil
}

func parseSize(size string, value float64) float64 {
	if strings.Contains(size, "PB") {
		value *= 1024.0 * 1024.0
	} else if strings.Contains(size, "TB") {
		value *= 1024.0
	} else if strings.Contains(size, "MB") {
		value /= 1024.0
	}
	return value
}
