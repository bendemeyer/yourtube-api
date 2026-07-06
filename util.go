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
		// Cleanup any old socket file before opening a new one
		os.Remove(socket)
		l, err := net.Listen("unix", socket)
		if err == nil {
			os.Chmod(socket, 0666)
		}
		return l, err
	}
}
