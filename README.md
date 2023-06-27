# virtual-cluster

## Usage

```shell
go run cmd/virtual-cluster/main.go -- \
    substrate start \
        --config-dir ~/workplace/virtual-cluster-internal/config/ \
        --db-path /tmp/vcluster.sqlite3 \
        --working-dir auth=/Users/asimi/workplace/mdu-auth \
        --verbose | tee /tmp/output.log
```
