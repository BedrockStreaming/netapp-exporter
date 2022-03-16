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

var netappStorageDiskError = promauto.NewGaugeVec(
	prometheus.GaugeOpts{
		Name: "netapp_storage_disk_error",
		Help: "Presence of disk errors",
	},
	[]string{
		"host",
	},
)

var netappStorageAggregateStatus = promauto.NewGaugeVec(
	prometheus.GaugeOpts{
		Name: "netapp_storage_aggregate_status",
		Help: "Current storage aggregate status",
	},
	[]string{
		"host",
		"owner_node",
		"raid_group",
		"position",
		"disk",
		"pool",
		"type",
		"rpm",
	},
)

var netappStorageAggregatePhysicalSize = promauto.NewGaugeVec(
	prometheus.GaugeOpts{
		Name: "netapp_storage_aggregate_physical_size",
		Help: "Current storage aggregate physical size",
	},
	[]string{
		"host",
		"owner_node",
		"raid_group",
		"position",
		"disk",
		"pool",
		"type",
		"rpm",
	},
)

var netappStorageAggregateUsableSize = promauto.NewGaugeVec(
	prometheus.GaugeOpts{
		Name: "netapp_storage_aggregate_usable_size",
		Help: "Current storage aggregate usable size",
	},
	[]string{
		"host",
		"owner_node",
		"raid_group",
		"position",
		"disk",
		"pool",
		"type",
		"rpm",
	},
)

func getStorageDiskErrors(client *ssh.Client, netappHost string) {
	session, err := client.NewSession()
	if err != nil {
		log.Println(err)
		netappStorageDiskError.Reset()
		return
	}
	defer session.Close()
	// return an error if no disk error (empty message == error lol)
	// so we don't handle error here
	out, _ := session.CombinedOutput("storage disk error show")
	if len(string(out)) > 34 {
		netappStorageDiskError.WithLabelValues(netappHost).Set(1)
	} else {
		netappStorageDiskError.WithLabelValues(netappHost).Set(0)
	}
}

func getStorageAggregateStatus(client *ssh.Client, netappHost string) {
	session, err := client.NewSession()
	if err != nil {
		log.Println(err)
		netappStorageAggregateStatus.Reset()
		netappStorageAggregateUsableSize.Reset()
		netappStorageAggregatePhysicalSize.Reset()
		return
	}
	defer session.Close()
	out, err := session.CombinedOutput("storage aggregate show-status")
	if err != nil {
		log.Println(err)
		netappStorageAggregateStatus.Reset()
		netappStorageAggregateStatus.Reset()
		netappStorageAggregateUsableSize.Reset()
		netappStorageAggregatePhysicalSize.Reset()
		return
	}

	outStr := string(out)

	lines := strings.Split(outStr, "\r\n")
	oNode := ""
	raidGrp := ""
	for i := 0; i < len(lines); i++ {
		if strings.Contains(lines[i], "Owner Node: ") {
			oNode = strings.ReplaceAll(lines[i], "Owner Node: ", "")
			continue
		}
		if strings.Contains(lines[i], "   RAID Group ") {
			raidGrp = strings.Split(strings.ReplaceAll(lines[i], "   RAID Group ", ""), " ")[0]
			i += 3
			continue
		}

		column := regexp.MustCompile(`\s+`).Split(lines[i], -1)
		if len(column) != 9 {
			continue
		}
		if column[8] == "(normal)" {
			netappStorageAggregateStatus.WithLabelValues(netappHost, oNode, raidGrp, column[1], column[2], column[3], column[4], column[5]).Set(1)
		} else {
			netappStorageAggregateStatus.WithLabelValues(netappHost, oNode, raidGrp, column[1], column[2], column[3], column[4], column[5]).Set(0)
		}

		if column[7] != "" && column[7] != "-" {
			physical, err := strconv.ParseFloat(column[7][0:len(column[7])-2], 64)
			if err != nil {
				continue
			}
			netappStorageAggregatePhysicalSize.WithLabelValues(netappHost, oNode, raidGrp, column[1], column[2], column[3], column[4], column[5]).Set(parseSize(column[7], physical))
		}
		if column[6] != "" && column[6] != "-" {
			usable, err := strconv.ParseFloat(column[6][0:len(column[6])-2], 64)
			if err != nil {
				continue
			}
			netappStorageAggregateUsableSize.WithLabelValues(netappHost, oNode, raidGrp, column[1], column[2], column[3], column[4], column[5]).Set(parseSize(column[6], usable))
		}
	}
}
