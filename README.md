# Memo Index

### Install deps
```bash
dnf install -y golang,bzr
```

### Checkout repo
```bash
git clone git@github.com:memocash/index.git
cd index
export GOVCS='*:bzr|git'; build
```

### Generate some test data
```bash
./index test double_spend
```

### Open ports
Ensure your firewall allows incoming TCPv4 traffic on port 26770. This is the port used by the app.

Other ports of note which do not require firewall ingress:

* 8333 outbound to BCH node
* 19021 optional RPC port
* 26780/26781 are internal shards ports

### Run server
```bash
./index serve all 
```
