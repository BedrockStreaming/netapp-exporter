package collector

import (
	"log"
	"regexp"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"golang.org/x/crypto/ssh"
)

var netappVServerStatus = promauto.NewGaugeVec(
	prometheus.GaugeOpts{
		Name: "netapp_vserver_status",
		Help: "Current vserver status",
	},
	[]string{
		"host",
		"vserver",
		"type",
		"subtype",
	},
)

func getVServerStatus(client *ssh.Client, netappHost string) {
	session, err := client.NewSession()
	if err != nil {
		log.Println(err)
		netappVServerStatus.Reset()
		return
	}
	out, err := session.CombinedOutput("vserver show")
	if err != nil {
		log.Println(err)
		netappVServerStatus.Reset()
		return
	}

	outStr := string(out)

	lines := strings.Split(outStr, "\r\n")

	for _, line := range lines[3:len(lines)] {
		column := regexp.MustCompile(`\s+`).Split(line, -1)
		if len(column) != 7 {
			continue
		}
		if column[3] == "running" && column[4] == "running" {
			netappVServerStatus.WithLabelValues(netappHost, column[0], column[1], column[2]).Set(1)
		} else if column[3] != "-" || column[4] != "-" {
			netappVServerStatus.WithLabelValues(netappHost, column[0], column[1], column[2]).Set(0)
		}
	}
}
