package server

import "fmt"

func GetHost(port uint) string {
	return fmt.Sprintf("127.0.0.1:%d", port)
}
