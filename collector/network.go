package collector

import (
	"log"
	"regexp"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"golang.org/x/crypto/ssh"
)

var netappNetworkInterfaceStatus = promauto.NewGaugeVec(
	prometheus.GaugeOpts{
		Name: "netapp_network_interface_status",
		Help: "Current status of interface",
	},
	[]string{
		"host",
		"vserver",
		"interface",
		"ip",
		"node",
		"port",
	},
)
var netappNetworkInterfaceIsHome = promauto.NewGaugeVec(
	prometheus.GaugeOpts{
		Name: "netapp_network_interface_is_home",
		Help: "Is interface home ?",
	},
	[]string{
		"host",
		"vserver",
		"interface",
		"ip",
		"node",
		"port",
	},
)

var netappNetworkPortLinkStatus = promauto.NewGaugeVec(
	prometheus.GaugeOpts{
		Name: "netapp_network_port_link_status",
		Help: "Current link status of port",
	},
	[]string{
		"host",
		"node",
		"port",
		"ipspace",
		"domain",
		"mtu",
		"speed",
	},
)
var netappNetworkPortHealth = promauto.NewGaugeVec(
	prometheus.GaugeOpts{
		Name: "netapp_network_port_health",
		Help: "Current health of port",
	},
	[]string{
		"host",
		"node",
		"port",
		"ipspace",
		"domain",
		"mtu",
		"speed",
	},
)

func getNetworkInterfaceStatus(client *ssh.Client, netappHost string) {
	session, err := client.NewSession()
	if err != nil {
		log.Println(err)
		netappNetworkInterfaceStatus.Reset()
		netappNetworkInterfaceIsHome.Reset()
		return
	}
	defer session.Close()
	out, err := session.CombinedOutput("network interface show")
	if err != nil {
		log.Println(err)
		netappNetworkInterfaceStatus.Reset()
		netappNetworkInterfaceIsHome.Reset()
		return
	}
	outStr := string(out)

	lines := strings.Split(outStr, "\r\n")
	vserver := ""
	for _, split := range lines[3:] {
		var reg = regexp.MustCompile(`^(\S+)$`)
		line := reg.FindAllString(split, -1)

		// Vserver
		if len(line) > 0 {
			vserver = line[0]
		}

		reg = regexp.MustCompile(`^\s+(\S+)\s+(\S+)\s+(\S+)\s+(\S+)\s+(\S+)\s+(\S+)$`)
		line = reg.FindAllString(split, -1)

		// Interface
		if len(line) > 0 {
			column := regexp.MustCompile(`\s+`).Split(line[0], -1)
			if column[2] == "up/up" {
				netappNetworkInterfaceStatus.WithLabelValues(netappHost, vserver, column[1], column[3], column[4], column[5]).Set(1)
			} else {
				netappNetworkInterfaceStatus.WithLabelValues(netappHost, vserver, column[1], column[3], column[4], column[5]).Set(0)
			}
			if column[6] == "true" {
				netappNetworkInterfaceIsHome.WithLabelValues(netappHost, vserver, column[1], column[3], column[4], column[5]).Set(1)
			} else {
				netappNetworkInterfaceIsHome.WithLabelValues(netappHost, vserver, column[1], column[3], column[4], column[5]).Set(0)
			}
		}
	}
}

func getNetworkPortStatus(client *ssh.Client, netappHost string) {
	session, err := client.NewSession()
	if err != nil {
		log.Println(err)
		netappNetworkPortLinkStatus.Reset()
		netappNetworkPortHealth.Reset()
		return
	}
	defer session.Close()
	out, err := session.CombinedOutput("network port show")
	if err != nil {
		log.Println(err)
		netappNetworkPortLinkStatus.Reset()
		netappNetworkPortHealth.Reset()
		return
	}
	outStr := string(out)

	nodes := strings.Split(outStr, "Node:")
	for _, node := range nodes[1:3] {
		lines := strings.Split(node, "\r\n")
		nodeName := lines[0][1:len(lines[0])]
		for _, line := range lines[4:] {
			column := regexp.MustCompile(`\s+`).Split(line, -1)
			if len(column) < 7 {
				continue
			}

			if column[3] == "up" {
				netappNetworkPortLinkStatus.WithLabelValues(netappHost, nodeName, column[0], column[1], column[2], column[4], column[5]).Set(1)
			} else {
				netappNetworkPortLinkStatus.WithLabelValues(netappHost, nodeName, column[0], column[1], column[2], column[4], column[5]).Set(0)
			}
			if column[6] == "healthy" {
				netappNetworkPortHealth.WithLabelValues(netappHost, nodeName, column[0], column[1], column[2], column[4], column[5]).Set(1)
			} else {
				netappNetworkPortHealth.WithLabelValues(netappHost, nodeName, column[0], column[1], column[2], column[4], column[5]).Set(0)
			}
		}
	}
}
