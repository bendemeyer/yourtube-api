package main

import (
	"fmt"
	"net"
	"os"
)


func GetListener(port string, socket string) (net.Listener, error) {
	if port != "" {
		return net.Listen("tcp", fmt.Sprintf("127.0.0.1:%s", port))
	} else {
		l, err := net.Listen("unix", socket)
		if err == nil {
			os.Chmod(socket, 0777)
		}
		return l, err
	}
}