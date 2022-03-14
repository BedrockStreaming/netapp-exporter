package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/BedrockStreaming/netapp-exporter/collector"

	//"github.com/prometheus/exporter-toolkit/web"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"gopkg.in/yaml.v2"
)

type Config struct {
	Host           string `yaml:"host"`
	Port           int    `yaml:"port"`
	Username       string `yaml:"username"`
	Password       string `yaml:"password"`
	Interval       int    `yaml:"interval"`
	KnownHostsFile string `yaml:"known_hosts_file"`
}

var config []*Config

func handler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	target := query.Get("target")
	if len(query["target"]) != 1 || target == "" {
		http.Error(w, "'target' parameter must be specified once", 400)
		return
	}

	hostExists := false
	for _, config := range config {
		if target == config.Host {
			//hostExists = true
			break
		}
	}
	if !hostExists {
		http.Error(w, "this target does not exist in config", 400)
		return
	}



	// logger = log.With(logger, "target", target)
	// level.Debug(logger).Log("msg", "Starting scrape")

	// start := time.Now()
	registry := prometheus.NewRegistry()
	//c := collector.New(r.Context(), target, logger)
	//registry.MustRegister(c)
	// Delegate http serving to Prometheus client library, which will call collector.Collect.
	h := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
	h.ServeHTTP(w, r)
	// duration := time.Since(start).Seconds()
	// level.Debug(logger).Log("msg", "Finished scrape", "duration_seconds", duration)
}

func main() {
	configPath := flag.String("config", "config.yaml", "Config path")
	yamlFile, err := ioutil.ReadFile(*configPath)
	if err != nil {
		log.Printf("yamlFile.Get err   #%v ", err)
	}
	err = yaml.Unmarshal(yamlFile, &config)
	if err != nil {
		log.Fatalf("Unmarshal: %v", err)
	}

	listenAddress := flag.String("listenAddress", "0.0.0.0", "Expose prometheus metrics")
	listenPort := flag.Int("listenPort", 9148, "Expose prometheus metrics")

	flag.Parse()

	for _, config := range config {
		collector.RecordMetrics(config.Username, config.Password, config.Host, config.Port, config.Interval, config.KnownHostsFile)
	}

	log.Println("Listening on " + *listenAddress + ":" + fmt.Sprint(*listenPort) + "/metrics")
	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/netapp", func(w http.ResponseWriter, r *http.Request) {
		log.Println("Listening on " + *listenAddress + ":" + fmt.Sprint(*listenPort) + "/netapp")
		handler(w, r)
	})
	//err := http.ListenAndServe(*listenAddress+":"+fmt.Sprint(*listenPort), nil)
	// if err != nil {
	// 	log.Fatalln(err)
	// }
}
