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

Another tab, example:

```shell
curl -X POST -H "Content-Type: application/json" -d '{
  "title": "My Community",
  "street_address": "123 Main St",
  "city": "Anytown",
  "country_area": "USA",
  "zipcode": "12345",
  "phone_number": "555-555-5555"
}' http://localhost:5003/mdu-auth/v1/community | jq -r '.uuid'
```
