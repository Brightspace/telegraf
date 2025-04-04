package all

import (
	//Blank imports for plugins to register themselves
	_ "github.com/influxdata/telegraf/plugins/inputs/activemq"
	_ "github.com/influxdata/telegraf/plugins/inputs/aerospike"
	_ "github.com/influxdata/telegraf/plugins/inputs/aliyuncms"
	_ "github.com/influxdata/telegraf/plugins/inputs/amd_rocm_smi"
	_ "github.com/influxdata/telegraf/plugins/inputs/amqp_consumer"
	_ "github.com/influxdata/telegraf/plugins/inputs/apache"
	_ "github.com/influxdata/telegraf/plugins/inputs/apcupsd"
	_ "github.com/influxdata/telegraf/plugins/inputs/aurora"
	_ "github.com/influxdata/telegraf/plugins/inputs/azure_storage_queue"
	_ "github.com/influxdata/telegraf/plugins/inputs/bcache"
	_ "github.com/influxdata/telegraf/plugins/inputs/beanstalkd"
	_ "github.com/influxdata/telegraf/plugins/inputs/beat"
	_ "github.com/influxdata/telegraf/plugins/inputs/bind"
	_ "github.com/influxdata/telegraf/plugins/inputs/bond"
	_ "github.com/influxdata/telegraf/plugins/inputs/burrow"
	_ "github.com/influxdata/telegraf/plugins/inputs/cassandra"
	_ "github.com/influxdata/telegraf/plugins/inputs/ceph"
	_ "github.com/influxdata/telegraf/plugins/inputs/cgroup"
	_ "github.com/influxdata/telegraf/plugins/inputs/chrony"
	_ "github.com/influxdata/telegraf/plugins/inputs/cisco_telemetry_mdt"
	_ "github.com/influxdata/telegraf/plugins/inputs/clickhouse"
	_ "github.com/influxdata/telegraf/plugins/inputs/cloud_pubsub"
	_ "github.com/influxdata/telegraf/plugins/inputs/cloud_pubsub_push"
	_ "github.com/influxdata/telegraf/plugins/inputs/cloudwatch"
	_ "github.com/influxdata/telegraf/plugins/inputs/conntrack"
	_ "github.com/influxdata/telegraf/plugins/inputs/consul"
	_ "github.com/influxdata/telegraf/plugins/inputs/consul_agent"
	_ "github.com/influxdata/telegraf/plugins/inputs/couchbase"
	_ "github.com/influxdata/telegraf/plugins/inputs/couchdb"
	_ "github.com/influxdata/telegraf/plugins/inputs/cpu"
	_ "github.com/influxdata/telegraf/plugins/inputs/csgo"
	_ "github.com/influxdata/telegraf/plugins/inputs/dcos"
	_ "github.com/influxdata/telegraf/plugins/inputs/directory_monitor"
	_ "github.com/influxdata/telegraf/plugins/inputs/disk"
	_ "github.com/influxdata/telegraf/plugins/inputs/diskio"
	_ "github.com/influxdata/telegraf/plugins/inputs/disque"
	_ "github.com/influxdata/telegraf/plugins/inputs/dmcache"
	_ "github.com/influxdata/telegraf/plugins/inputs/dns_query"
	_ "github.com/influxdata/telegraf/plugins/inputs/docker"
	_ "github.com/influxdata/telegraf/plugins/inputs/docker_log"
	_ "github.com/influxdata/telegraf/plugins/inputs/dovecot"
	_ "github.com/influxdata/telegraf/plugins/inputs/dpdk"
	_ "github.com/influxdata/telegraf/plugins/inputs/ecs"
	_ "github.com/influxdata/telegraf/plugins/inputs/elasticsearch"
	_ "github.com/influxdata/telegraf/plugins/inputs/elasticsearch_query"
	_ "github.com/influxdata/telegraf/plugins/inputs/ethtool"
	_ "github.com/influxdata/telegraf/plugins/inputs/eventhub_consumer"
	_ "github.com/influxdata/telegraf/plugins/inputs/exec"
	_ "github.com/influxdata/telegraf/plugins/inputs/execd"
	_ "github.com/influxdata/telegraf/plugins/inputs/fail2ban"
	_ "github.com/influxdata/telegraf/plugins/inputs/fibaro"
	_ "github.com/influxdata/telegraf/plugins/inputs/file"
	_ "github.com/influxdata/telegraf/plugins/inputs/filecount"
	_ "github.com/influxdata/telegraf/plugins/inputs/filestat"
	_ "github.com/influxdata/telegraf/plugins/inputs/fireboard"
	_ "github.com/influxdata/telegraf/plugins/inputs/fluentd"
	_ "github.com/influxdata/telegraf/plugins/inputs/github"
	_ "github.com/influxdata/telegraf/plugins/inputs/gnmi"
	_ "github.com/influxdata/telegraf/plugins/inputs/graylog"
	_ "github.com/influxdata/telegraf/plugins/inputs/haproxy"
	_ "github.com/influxdata/telegraf/plugins/inputs/hddtemp"
	_ "github.com/influxdata/telegraf/plugins/inputs/http"
	_ "github.com/influxdata/telegraf/plugins/inputs/http_listener_v2"
	_ "github.com/influxdata/telegraf/plugins/inputs/http_response"
	_ "github.com/influxdata/telegraf/plugins/inputs/httpjson"
	_ "github.com/influxdata/telegraf/plugins/inputs/hugepages"
	_ "github.com/influxdata/telegraf/plugins/inputs/icinga2"
	_ "github.com/influxdata/telegraf/plugins/inputs/infiniband"
	_ "github.com/influxdata/telegraf/plugins/inputs/influxdb"
	_ "github.com/influxdata/telegraf/plugins/inputs/influxdb_listener"
	_ "github.com/influxdata/telegraf/plugins/inputs/influxdb_v2_listener"
	_ "github.com/influxdata/telegraf/plugins/inputs/intel_pmu"
	_ "github.com/influxdata/telegraf/plugins/inputs/intel_powerstat"
	_ "github.com/influxdata/telegraf/plugins/inputs/intel_rdt"
	_ "github.com/influxdata/telegraf/plugins/inputs/internal"
	_ "github.com/influxdata/telegraf/plugins/inputs/internet_speed"
	_ "github.com/influxdata/telegraf/plugins/inputs/interrupts"
	_ "github.com/influxdata/telegraf/plugins/inputs/ipmi_sensor"
	_ "github.com/influxdata/telegraf/plugins/inputs/ipset"
	_ "github.com/influxdata/telegraf/plugins/inputs/iptables"
	_ "github.com/influxdata/telegraf/plugins/inputs/ipvs"
	_ "github.com/influxdata/telegraf/plugins/inputs/jenkins"
	_ "github.com/influxdata/telegraf/plugins/inputs/jolokia"
	_ "github.com/influxdata/telegraf/plugins/inputs/jolokia2"
	_ "github.com/influxdata/telegraf/plugins/inputs/jti_openconfig_telemetry"
	_ "github.com/influxdata/telegraf/plugins/inputs/kafka_consumer"
	_ "github.com/influxdata/telegraf/plugins/inputs/kafka_consumer_legacy"
	_ "github.com/influxdata/telegraf/plugins/inputs/kapacitor"
	_ "github.com/influxdata/telegraf/plugins/inputs/kernel"
	_ "github.com/influxdata/telegraf/plugins/inputs/kernel_vmstat"
	_ "github.com/influxdata/telegraf/plugins/inputs/kibana"
	_ "github.com/influxdata/telegraf/plugins/inputs/kinesis_consumer"
	_ "github.com/influxdata/telegraf/plugins/inputs/knx_listener"
	_ "github.com/influxdata/telegraf/plugins/inputs/kube_inventory"
	_ "github.com/influxdata/telegraf/plugins/inputs/kubernetes"
	_ "github.com/influxdata/telegraf/plugins/inputs/lanz"
	_ "github.com/influxdata/telegraf/plugins/inputs/leofs"
	_ "github.com/influxdata/telegraf/plugins/inputs/linux_sysctl_fs"
	_ "github.com/influxdata/telegraf/plugins/inputs/logparser"
	_ "github.com/influxdata/telegraf/plugins/inputs/logstash"
	_ "github.com/influxdata/telegraf/plugins/inputs/lustre2"
	_ "github.com/influxdata/telegraf/plugins/inputs/lvm"
	_ "github.com/influxdata/telegraf/plugins/inputs/mailchimp"
	_ "github.com/influxdata/telegraf/plugins/inputs/marklogic"
	_ "github.com/influxdata/telegraf/plugins/inputs/mcrouter"
	_ "github.com/influxdata/telegraf/plugins/inputs/mdstat"
	_ "github.com/influxdata/telegraf/plugins/inputs/mem"
	_ "github.com/influxdata/telegraf/plugins/inputs/memcached"
	_ "github.com/influxdata/telegraf/plugins/inputs/mesos"
	_ "github.com/influxdata/telegraf/plugins/inputs/minecraft"
	_ "github.com/influxdata/telegraf/plugins/inputs/mock"
	_ "github.com/influxdata/telegraf/plugins/inputs/modbus"
	_ "github.com/influxdata/telegraf/plugins/inputs/mongodb"
	_ "github.com/influxdata/telegraf/plugins/inputs/monit"
	_ "github.com/influxdata/telegraf/plugins/inputs/mqtt_consumer"
	_ "github.com/influxdata/telegraf/plugins/inputs/multifile"
	_ "github.com/influxdata/telegraf/plugins/inputs/mysql"
	_ "github.com/influxdata/telegraf/plugins/inputs/nats"
	_ "github.com/influxdata/telegraf/plugins/inputs/nats_consumer"
	_ "github.com/influxdata/telegraf/plugins/inputs/neptune_apex"
	_ "github.com/influxdata/telegraf/plugins/inputs/net"
	_ "github.com/influxdata/telegraf/plugins/inputs/net_response"
	_ "github.com/influxdata/telegraf/plugins/inputs/netstat"
	_ "github.com/influxdata/telegraf/plugins/inputs/nfsclient"
	_ "github.com/influxdata/telegraf/plugins/inputs/nginx"
	_ "github.com/influxdata/telegraf/plugins/inputs/nginx_plus"
	_ "github.com/influxdata/telegraf/plugins/inputs/nginx_plus_api"
	_ "github.com/influxdata/telegraf/plugins/inputs/nginx_sts"
	_ "github.com/influxdata/telegraf/plugins/inputs/nginx_upstream_check"
	_ "github.com/influxdata/telegraf/plugins/inputs/nginx_vts"
	_ "github.com/influxdata/telegraf/plugins/inputs/nomad"
	_ "github.com/influxdata/telegraf/plugins/inputs/nsd"
	_ "github.com/influxdata/telegraf/plugins/inputs/nsq"
	_ "github.com/influxdata/telegraf/plugins/inputs/nsq_consumer"
	_ "github.com/influxdata/telegraf/plugins/inputs/nstat"
	_ "github.com/influxdata/telegraf/plugins/inputs/ntpq"
	_ "github.com/influxdata/telegraf/plugins/inputs/nvidia_smi"
	_ "github.com/influxdata/telegraf/plugins/inputs/opcua"
	_ "github.com/influxdata/telegraf/plugins/inputs/openldap"
	_ "github.com/influxdata/telegraf/plugins/inputs/openntpd"
	_ "github.com/influxdata/telegraf/plugins/inputs/opensmtpd"
	_ "github.com/influxdata/telegraf/plugins/inputs/openstack"
	_ "github.com/influxdata/telegraf/plugins/inputs/opentelemetry"
	_ "github.com/influxdata/telegraf/plugins/inputs/openweathermap"
	_ "github.com/influxdata/telegraf/plugins/inputs/passenger"
	_ "github.com/influxdata/telegraf/plugins/inputs/pf"
	_ "github.com/influxdata/telegraf/plugins/inputs/pgbouncer"
	_ "github.com/influxdata/telegraf/plugins/inputs/phpfpm"
	_ "github.com/influxdata/telegraf/plugins/inputs/ping"
	_ "github.com/influxdata/telegraf/plugins/inputs/postfix"
	_ "github.com/influxdata/telegraf/plugins/inputs/postgresql"
	_ "github.com/influxdata/telegraf/plugins/inputs/postgresql_extensible"
	_ "github.com/influxdata/telegraf/plugins/inputs/powerdns"
	_ "github.com/influxdata/telegraf/plugins/inputs/powerdns_recursor"
	_ "github.com/influxdata/telegraf/plugins/inputs/processes"
	_ "github.com/influxdata/telegraf/plugins/inputs/procstat"
	_ "github.com/influxdata/telegraf/plugins/inputs/prometheus"
	_ "github.com/influxdata/telegraf/plugins/inputs/proxmox"
	_ "github.com/influxdata/telegraf/plugins/inputs/puppetagent"
	_ "github.com/influxdata/telegraf/plugins/inputs/rabbitmq"
	_ "github.com/influxdata/telegraf/plugins/inputs/raindrops"
	_ "github.com/influxdata/telegraf/plugins/inputs/ras"
	_ "github.com/influxdata/telegraf/plugins/inputs/ravendb"
	_ "github.com/influxdata/telegraf/plugins/inputs/redfish"
	_ "github.com/influxdata/telegraf/plugins/inputs/redis"
	_ "github.com/influxdata/telegraf/plugins/inputs/redis_sentinel"
	_ "github.com/influxdata/telegraf/plugins/inputs/rethinkdb"
	_ "github.com/influxdata/telegraf/plugins/inputs/riak"
	_ "github.com/influxdata/telegraf/plugins/inputs/riemann_listener"
	_ "github.com/influxdata/telegraf/plugins/inputs/salesforce"
	_ "github.com/influxdata/telegraf/plugins/inputs/sensors"
	_ "github.com/influxdata/telegraf/plugins/inputs/sflow"
	_ "github.com/influxdata/telegraf/plugins/inputs/slab"
	_ "github.com/influxdata/telegraf/plugins/inputs/smart"
	_ "github.com/influxdata/telegraf/plugins/inputs/snmp"
	_ "github.com/influxdata/telegraf/plugins/inputs/snmp_legacy"
	_ "github.com/influxdata/telegraf/plugins/inputs/snmp_trap"
	_ "github.com/influxdata/telegraf/plugins/inputs/socket_listener"
	_ "github.com/influxdata/telegraf/plugins/inputs/socketstat"
	_ "github.com/influxdata/telegraf/plugins/inputs/solr"
	_ "github.com/influxdata/telegraf/plugins/inputs/sql"
	_ "github.com/influxdata/telegraf/plugins/inputs/sqlserver"
	_ "github.com/influxdata/telegraf/plugins/inputs/stackdriver"
	_ "github.com/influxdata/telegraf/plugins/inputs/statsd"
	_ "github.com/influxdata/telegraf/plugins/inputs/suricata"
	_ "github.com/influxdata/telegraf/plugins/inputs/swap"
	_ "github.com/influxdata/telegraf/plugins/inputs/synproxy"
	_ "github.com/influxdata/telegraf/plugins/inputs/syslog"
	_ "github.com/influxdata/telegraf/plugins/inputs/sysstat"
	_ "github.com/influxdata/telegraf/plugins/inputs/system"
	_ "github.com/influxdata/telegraf/plugins/inputs/systemd_units"
	_ "github.com/influxdata/telegraf/plugins/inputs/tail"
	_ "github.com/influxdata/telegraf/plugins/inputs/tcp_listener"
	_ "github.com/influxdata/telegraf/plugins/inputs/teamspeak"
	_ "github.com/influxdata/telegraf/plugins/inputs/temp"
	_ "github.com/influxdata/telegraf/plugins/inputs/tengine"
	_ "github.com/influxdata/telegraf/plugins/inputs/tomcat"
	_ "github.com/influxdata/telegraf/plugins/inputs/trig"
	_ "github.com/influxdata/telegraf/plugins/inputs/twemproxy"
	_ "github.com/influxdata/telegraf/plugins/inputs/udp_listener"
	_ "github.com/influxdata/telegraf/plugins/inputs/unbound"
	_ "github.com/influxdata/telegraf/plugins/inputs/uwsgi"
	_ "github.com/influxdata/telegraf/plugins/inputs/varnish"
	_ "github.com/influxdata/telegraf/plugins/inputs/vault"
	_ "github.com/influxdata/telegraf/plugins/inputs/vsphere"
	_ "github.com/influxdata/telegraf/plugins/inputs/webhooks"
	_ "github.com/influxdata/telegraf/plugins/inputs/win_eventlog"
	_ "github.com/influxdata/telegraf/plugins/inputs/win_perf_counters"
	_ "github.com/influxdata/telegraf/plugins/inputs/win_services"
	_ "github.com/influxdata/telegraf/plugins/inputs/wireguard"
	_ "github.com/influxdata/telegraf/plugins/inputs/wireless"
	_ "github.com/influxdata/telegraf/plugins/inputs/x509_cert"
	_ "github.com/influxdata/telegraf/plugins/inputs/xtremio"
	_ "github.com/influxdata/telegraf/plugins/inputs/zfs"
	_ "github.com/influxdata/telegraf/plugins/inputs/zipkin"
	_ "github.com/influxdata/telegraf/plugins/inputs/zookeeper"
)
