# netapp-exporter

Prometheus exporter for ONTAP 9.1 using SSH parsing.

## Build

```
GOOS=linux GOARCH=amd64 go build -o netapp-exporter-linux-amd64
```

## Usage

```
Usage of netapp-exporter:
  -config string
        Config file path (default "config.yaml")
  -listenAddress string
        Expose prometheus metrics (default "0.0.0.0")
  -listenPort int
        Expose prometheus metrics (default 9148)
```

### Configuration

```
- host: "10.0.0.1"                              # Netapp Host
  port: 22                                      # Netapp SSH Port
  interval: 30                                  # Collect Interval
  username: "XXX"                               # Netapp SSH Username
  password: "XXX"                               # Netapp SSH Password
  known_hosts_file: "/root/.ssh/known_hosts"        # For SSH Secure Connection
```

## Metrics

Ontap | Metrics
--- | ---
`network interface show` | `netapp_network_interface_status` && `netapp_network_interface_is_home`
`network port show` | `netapp_network_port_link_status` && `netapp_network_port_health`
`storage aggregate show-status` | `netapp_storage_aggregate_status` && `netapp_storage_aggregate_physical_size` && `netapp_storage_aggregate_usable_size`
`storage disk error show` | `netapp_storage_disk_error`
`system health status show` | `netapp_system_health`
`volume show` | `netapp_volume_status` && `netapp_volume_used_size` && `netapp_volume_available_size` && `netapp_volume_total_size`
`vserver show` | `netapp_vserver_status`

[More about ONTAP commands](https://docs.netapp.com/ontap-9/topic/com.netapp.doc.dot-cm-cmpr-910/home.html)