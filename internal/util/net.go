package util

import (
	"fmt"
	"log"
	"net"
	"os"
)

func GetListener(port string, socket string) (net.Listener, error) {
	if port != "" {
		log.Printf("Listening on port %s", port)
		return net.Listen("tcp", fmt.Sprintf("0.0.0.0:%s", port))
	} else {
		// Cleanup any old socket file before opening a new one
		os.Remove(socket)
		log.Printf("Listening on unix socket %s", socket)
		l, err := net.Listen("unix", socket)
		if err == nil {
			os.Chmod(socket, 0666)
		}
		return l, err
	}
}
