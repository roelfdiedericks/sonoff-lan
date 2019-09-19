package main

import (
	"fmt"
	"log"
	"net"
	"time"
)

const (
	PING_SEND_PORT_NUMBER   = 58100
	PING_LISTEN_PORT_NUMBER = 58000
	PING_MSG_SIZE           = 80
	PING_INTERVAL           = 1000 * time.Millisecond //  Once per second
)

func main() {

	log.SetFlags(log.Lshortfile)

	//  Create UDP socket

	laddr := net.UDPAddr{IP: net.IPv4(192, 168, 1, 249), Port: PING_LISTEN_PORT_NUMBER}
	BROADCAST_IPv4 := net.IPv4(255, 255, 255, 255)
	socket, err := net.DialUDP("udp4", &laddr, &net.UDPAddr{IP: BROADCAST_IPv4, Port: PING_SEND_PORT_NUMBER})

	if err != nil {
		log.Fatalln(err)
	}

	buffer := make([]byte, PING_MSG_SIZE)

	//  We send a beacon once a second, and we collect and report
	//  beacons that come in from other nodes:

	//  Send first ping right away
	ping_at := time.Now()

	for {

		if time.Now().After(ping_at) {
			//  Broadcast our beacon
			fmt.Println("Pinging peers...")
			buffer[0] = '!'
			if _, err := socket.Write(buffer); err != nil {
				log.Fatalln(err)
			}
			ping_at = time.Now().Add(PING_INTERVAL)
		}
	}
}
