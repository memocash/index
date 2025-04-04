# Directions

## Install the Protocol Buffers Compiler

```bash
brew install protobuf
export GOPATH=~/go
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

## Generate the Go Code

```bash
src="/usr/local/src"
cd $src/github.com/memocash/index/db/proto
mkdir -p queue_pb
protoc --go_out=$src ./queue.proto
protoc --go-grpc_out=$src ./queue.proto
```
