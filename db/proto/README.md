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
cd db/proto
protoc --go_out=../.. --go-grpc_out=../.. --go_opt=module=github.com/memocash/index --go-grpc_opt=module=github.com/memocash/index ./queue.proto
```
