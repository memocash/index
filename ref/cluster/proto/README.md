# Run

```bash
cd ref/cluster/proto
src="/usr/local/src"
mkdir -p cluster_pb
protoc --go_out=$src ./*.proto
protoc --go-grpc_out=$src ./*.proto
```
