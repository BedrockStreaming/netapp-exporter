package collector

import (
	"fmt"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
)

var netappHost string
var netappPort int

func RecordMetrics(username string, password string, netappH string, netappP int, interval int) {
	netappHost = netappH
	netappPort = netappP

	go func() {
		for {
			client, err := connectToHost(username, password)
			if err != nil {
				fmt.Println(err)
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

func connectToHost(username string, password string) (*ssh.Client, error) {
	sshConfig := &ssh.ClientConfig{
		User:    username,
		Auth:    []ssh.AuthMethod{ssh.Password(password)},
		Timeout: time.Second * 30,
	}
	sshConfig.HostKeyCallback = ssh.InsecureIgnoreHostKey()

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
