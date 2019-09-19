package main

import (
	"encoding/hex"
	_ "encoding/json"
	_ "flag"
	"fmt"
	_ "github.com/gorilla/websocket"
	_ "golang.org/x/net/ipv4"
	"golang.org/x/sys/windows"
	"log"
	"net"
	_ "net/http"
	_ "net/url"
	_ "syscall"
	"time"
)

func handleSonoffUDPConnection(conn *net.UDPConn) {

	buffer := make([]byte, 1024)

	n, addr, err := conn.ReadFromUDP(buffer)

	fmt.Println("UDP received from: ", addr)
	fmt.Println("Received UDP message  :  ", string(buffer[:n]))
	log.Println(hex.Dump(buffer[:n]))

	if err != nil {
		log.Fatal(err)
	}

}

func bcastSend(local *net.UDPAddr) {
	bcastAddr := "255.255.255.255:58100"
	addr, err := net.ResolveUDPAddr("udp", bcastAddr)
	if err != nil {
		log.Fatal(err)
	}
	c, err := net.DialUDP("udp4", nil, addr)

	raw, err := c.SyscallConn()
	raw.Control((func(fd uintptr) {
		err = windows.SetsockoptInt(windows.Handle(fd), windows.SOL_SOCKET, windows.SO_REUSEADDR, 1)
	}))

	if err != nil {
		log.Fatal(err)
	}
	for {
		log.Println("sending broadcast")
		_, err := c.Write([]byte("hello, world\n"))
		if err != nil {
			log.Fatal(err)
		}

		time.Sleep(1 * time.Second)
	}
}

func bcastListen() {
	hostName := "0.0.0.0"
	portNum := "58000"
	service := hostName + ":" + portNum

	udpAddr, err := net.ResolveUDPAddr("udp4", service)

	if err != nil {
		log.Fatal(err)
	}

	// setup listener for incoming UDP connection
	ln, err := net.ListenUDP("udp", udpAddr)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("UDP server up and listening on port 58000")

	defer ln.Close()

	go bcastSend(udpAddr)

	for {
		// wait for UDP client to connect
		handleSonoffUDPConnection(ln)
	}
}

func main() {

	bcastListen()

}
