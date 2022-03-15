package collector

import (
	"log"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"golang.org/x/crypto/ssh"
)

var netappSystemHealth = promauto.NewGaugeVec(
	prometheus.GaugeOpts{
		Name: "netapp_system_health",
		Help: "Current system health",
	},
	[]string{
		"host",
	},
)

func getSystemHealth(client *ssh.Client, netappHost string) {
	session, err := client.NewSession()
	if err != nil {
		log.Println(err)
		netappSystemHealth.Reset()
		return
	}
	defer session.Close()
	out, err := session.CombinedOutput("system health status show")
	if err != nil {
		log.Println(err)
		netappSystemHealth.Reset()
		return
	}
	outStr := string(out)

	lines := strings.Split(outStr, "\r\n")

	if lines[2] == "ok" {
		netappSystemHealth.WithLabelValues(netappHost).Set(1)
	} else {
		netappSystemHealth.WithLabelValues(netappHost).Set(0)
	}
}
