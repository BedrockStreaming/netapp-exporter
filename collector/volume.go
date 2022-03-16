package collector

import (
	"log"
	"regexp"
	"strconv"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"golang.org/x/crypto/ssh"
)

var netappVolumeStatus = promauto.NewGaugeVec(
	prometheus.GaugeOpts{
		Name: "netapp_volume_status",
		Help: "Current volume status",
	},
	[]string{
		"host",
		"vserver",
		"volume",
		"aggregate",
		"type",
	},
)

var netappVolumeUsedSize = promauto.NewGaugeVec(
	prometheus.GaugeOpts{
		Name: "netapp_volume_used_size",
		Help: "Current volume used size (%)",
	},
	[]string{
		"host",
		"vserver",
		"volume",
		"aggregate",
		"type",
	},
)

var netappVolumeAvailableSize = promauto.NewGaugeVec(
	prometheus.GaugeOpts{
		Name: "netapp_volume_available_size",
		Help: "Current volume available size (GB)",
	},
	[]string{
		"host",
		"vserver",
		"volume",
		"aggregate",
		"type",
	},
)

var netappVolumeTotalSize = promauto.NewGaugeVec(
	prometheus.GaugeOpts{
		Name: "netapp_volume_total_size",
		Help: "Current volume total size (GB)",
	},
	[]string{
		"host",
		"vserver",
		"volume",
		"aggregate",
		"type",
	},
)

func getVolumeStatus(client *ssh.Client, netappHost string) {
	session, err := client.NewSession()
	if err != nil {
		log.Println(err)
		netappVolumeStatus.Reset()
		netappVolumeUsedSize.Reset()
		netappVolumeAvailableSize.Reset()
		netappVolumeTotalSize.Reset()
		return
	}
	defer session.Close()
	out, err := session.CombinedOutput("volume show")
	if err != nil {
		log.Println(err)
		netappVolumeStatus.Reset()
		netappVolumeUsedSize.Reset()
		netappVolumeAvailableSize.Reset()
		netappVolumeTotalSize.Reset()
		return
	}
	outStr := string(out)

	lines := strings.Split(outStr, "\r\n")

	for _, line := range lines[2:] {
		column := regexp.MustCompile(`\s+`).Split(line, -1)
		if len(column) != 8 {
			continue
		}
		if column[3] == "online" {
			netappVolumeStatus.WithLabelValues(netappHost, column[0], column[1], column[2], column[4]).Set(1)
		} else {
			netappVolumeStatus.WithLabelValues(netappHost, column[0], column[1], column[2], column[4]).Set(0)
		}

		if column[7] != "" && column[7] != "-" {
			used, err := strconv.ParseFloat(column[7][0:len(column[7])-1], 64)
			if err != nil {
				continue
			}
			netappVolumeUsedSize.WithLabelValues(netappHost, column[0], column[1], column[2], column[4]).Set(used)
		}
		if column[6] != "" && column[6] != "-" {
			available, err := strconv.ParseFloat(column[6][0:len(column[6])-2], 64)
			if err != nil {
				continue
			}
			netappVolumeAvailableSize.WithLabelValues(netappHost, column[0], column[1], column[2], column[4]).Set(parseSize(column[6], available))
		}
		if column[5] != "" && column[5] != "-" {
			size, err := strconv.ParseFloat(column[5][0:len(column[5])-2], 64)
			if err != nil {
				continue
			}
			netappVolumeTotalSize.WithLabelValues(netappHost, column[0], column[1], column[2], column[4]).Set(parseSize(column[5], size))
		}
	}
}
