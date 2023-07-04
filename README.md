# virtual-cluster

## Usage

Build the binary:

```shell
make build
````

First tab:

```shell
build/virtual-cluster substrate start \
  --db-path /tmp/vcluster.sqlite3 \
  --config-file test_services/http_service_with_kafka/http_service_with_kafka.vcluster \
  --working-dir 'http_service_with_kafka=./test_services/http_service_with_kafka' \
  --verbose | tee /tmp/output.log
```

Second tab
