# Directions

```bash
src="/usr/local/src"
mkdir -p gen
protoc --go_out=$src ./queue.proto
protoc --go-grpc_out=$src ./queue.proto
```
