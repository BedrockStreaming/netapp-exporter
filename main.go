package main

import (
	"bedrock/netapp-exporter/collector"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"gopkg.in/yaml.v2"
)

type Config struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Interval int    `yaml:"interval"`
}

func redirect(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/metrics", 301)
}

func getConfigs(path string) []*Config {
	var configs []*Config
	yamlFile, err := ioutil.ReadFile(path)
	if err != nil {
		log.Printf("yamlFile.Get err #%v ", err)
		os.Exit(1)
	}
	err = yaml.Unmarshal(yamlFile, &configs)
	if err != nil {
		log.Fatalf("Unmarshal: %v", err)
		os.Exit(1)
	}
	return configs
}

func main() {
	configFile := flag.String("config", "config.yaml", "Config file path")
	listenAddress := flag.String("listenAddress", "0.0.0.0", "Expose prometheus metrics")
	listenPort := flag.Int("listenPort", 9148, "Expose prometheus metrics")

	flag.Parse()

	configs := getConfigs(*configFile)

	for _, config := range configs {
		collector.RecordMetrics(config.Username, config.Password, config.Host, config.Port, config.Interval)
	}

	fmt.Println("Listening on " + *listenAddress + ":" + fmt.Sprint(*listenPort) + "/metrics")
	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/", redirect)
	http.ListenAndServe(*listenAddress+":"+fmt.Sprint(*listenPort), nil)
}
