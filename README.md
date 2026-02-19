# Memo Index

### Install Go

https://go.dev/doc/install

### Checkout repo and run server
```sh
git clone git@github.com:memocash/index.git
cd index
go build
./index serve live
```

## Architecture

```mermaid
graph TD
    BCH[BCH Node] <-->|P2P| Lead[Lead Processor]
    Lead -->|Blocks/Txs| CS[Cluster Shards]

    subgraph Shard["Each Shard (0..N)"]
        CS --> Queue[Queue Server]
        Queue --> DB[(LevelDB)]
    end

    GraphQL[GraphQL Server] -->|gRPC| Queue
    Admin[Admin Server] -->|gRPC| Queue
    Network[Network Server] -->|gRPC| Queue

    Client([Client]) -->|Query| GraphQL
    Client -->|Manage| Admin
    Client -->|Submit Tx| Broadcast[Broadcast Server]
    Broadcast -->|Raw Tx| Lead
```

## Configuration

Two options for setting config values.

1. Use environment variables, e.g.
    ```sh
    NODE_HOST=example.com:8333 ./index serve live
    ```
2. Use a config file, e.g. `config.yaml`:
    ```yaml
    NODE_HOST: example.com:8333
    GRAPHQL_PORT: 8080
    ```
