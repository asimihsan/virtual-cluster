# virtual-cluster

## Usage

```shell
go run cmd/virtual-cluster/main.go -- \
    substrate start \
        --config-dir test_services/http_service_with_kafka \
        --db-path /tmp/vcluster.sqlite3 \
        --working-dir http_service_with_kafka=test_services/http_service_with_kafka \
        --verbose | tee /tmp/output.log
```
