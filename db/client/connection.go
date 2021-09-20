package client

import "google.golang.org/grpc"

type HostConn struct {
	Host string
	Conn *grpc.ClientConn
}

type ConnHandler struct {
	NewConn     chan HostConn
	Connections map[string]*grpc.ClientConn
}

func (h *ConnHandler) Start() {
	h.NewConn = make(chan HostConn)
	h.Connections = make(map[string]*grpc.ClientConn)
	go func() {
		for {
			select {
			case newConn := <-h.NewConn:
				h.Connections[newConn.Host] = newConn.Conn
			}
		}
	}()
}

func (h *ConnHandler) Get(host string) *grpc.ClientConn {
	return h.Connections[host]
}

func (h *ConnHandler) Add(host string, conn *grpc.ClientConn) {
	h.NewConn <- HostConn{
		Host: host,
		Conn: conn,
	}
}

