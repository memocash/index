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
Ensure your firewall allows incoming TCPv4 traffic on ports: 8333 19021 26780 26781

### Run server
```bash
./index serve all 
```
