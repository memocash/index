# Run

```bash
# Set to your go src path
# Run from this directory

src="/usr/local/src"
mkdir -p gen/network_pb
protoc --go_out=$src ./*.proto
protoc --go-grpc_out=$src ./*.proto
```
