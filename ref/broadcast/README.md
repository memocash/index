# Run

```bash
# Run from this directory

src="/usr/local/src"
mkdir -p gen/broadcast_pb
protoc --go_out=$src ./*.proto
protoc --go-grpc_out=$src ./*.proto
```
