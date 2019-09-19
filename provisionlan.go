package main

import (
	"encoding/hex"
	_ "encoding/json"
	_ "flag"
	"fmt"
	_ "github.com/gorilla/websocket"
	"golang.org/x/net/ipv4"
	"golang.org/x/sys/windows"
	"log"
	"net"
	_ "net/http"
	_ "net/url"
	"strings"
	_ "syscall"
	"time"
)

func handleSonoffUDPConnection(conn *net.UDPConn) {

	buffer := make([]byte, 1024)

	n, addr, err := conn.ReadFromUDP(buffer)

	fmt.Println("UDP received from: ", addr)
	fmt.Println("Received UDP message  :  ", string(buffer[:n]))

	if err != nil {
		log.Fatal(err)
	}

}

const (
	srvAddr         = "224.0.0.2:58000"
	maxDatagramSize = 1000
)

func ping(a string) {
	addr, err := net.ResolveUDPAddr("udp", a)
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
		c.Write([]byte("hello, world\n"))
		time.Sleep(1 * time.Second)
	}
}

func msgHandler(src *net.UDPAddr, n int, b []byte) {
	log.Println(n, "bytes read from", src)
	log.Println(hex.Dump(b[:n]))
}

func serveMcast(ifi *net.Interface) {
	log.Println("enabling mcast on interface", ifi.Name)
	group := net.IPv4(224, 0, 0, 1)
	c, err := net.ListenPacket("udp4", "0.0.0.0:58000")
	if err != nil {
		log.Fatal("listen 0.0.0.0:58000 failed:", err)
	}
	defer c.Close()

	p := ipv4.NewPacketConn(c)

	if err := p.JoinGroup(ifi, &net.UDPAddr{IP: group}); err != nil {
		log.Println("join mcast group failed:", err)
	}

	ssmgroup := net.UDPAddr{IP: net.IPv4(232, 7, 8, 9)}
	ssmsource := net.UDPAddr{IP: net.IPv4(192, 168, 0, 1)}
	if err := p.JoinSourceSpecificGroup(ifi, &ssmgroup, &ssmsource); err != nil {
		// error handling
	}

	if err := p.SetControlMessage(ipv4.FlagDst, true); err != nil {
		log.Println("enable control message flag failed:", err)
	}

	b := make([]byte, 1500)
	data := []byte("123")
	for {

		log.Println("writing....")

		dst := &net.UDPAddr{IP: group, Port: 58000}
		for _, ifi := range []*net.Interface{ifi} {
			if err := p.SetMulticastInterface(ifi); err != nil {
				// error handling
			}
			//p.SetMulticastTTL(2)
			if _, err := p.WriteTo(data, nil, dst); err != nil {
				// error handling
			}
		}

		log.Println("reading....")
		n, cm, src, err := p.ReadFrom(b)
		log.Printf("read %d bytes ", n)
		log.Println(hex.Dump(b[:n]))
		if err != nil {
			// error handling
		}
		_ = cm
		_ = src
		/*if cm.Dst.IsMulticast() {
			if cm.Dst.Equal(group) {
				// joined group, do something
			} else {
				// unknown group, discard
				continue
			}
		}*/
		time.Sleep(1 * time.Second)

		/*
			log.Println("writing....")
			p.SetTOS(0x0)
			p.SetTTL(16)
			if _, err := p.WriteTo(data, nil, src); err != nil {
				// error handling
			}
			//dst := &net.UDPAddr{IP: group, Port: 1024}
			for _, ifi := range []*net.Interface{ifi} {
				if err := p.SetMulticastInterface(ifi); err != nil {
					// error handling
				}
				p.SetMulticastTTL(2)
				if _, err := p.WriteTo(data, nil, dst); err != nil {
					// error handling
				}
			}
		*/
	}
}

func serveMulticastUDP(a string, ifi *net.Interface, h func(*net.UDPAddr, int, []byte)) {
	addr, err := net.ResolveUDPAddr("udp4", a)
	if err != nil {
		log.Fatal(err)
	}
	l, err := net.ListenMulticastUDP("udp4", ifi, addr)
	if err != nil {
		log.Fatal(err)
	}

	/*log.Println("gett file handle")
	f, err := l.File()
	if err != nil {
		log.Fatal(err)
	}
	log.Println("setting SO_REUSEADDR")
	fd := int(f.Fd())
	*/

	raw, err := l.SyscallConn()
	raw.Control((func(fd uintptr) {
		err = windows.SetsockoptInt(windows.Handle(fd), windows.SOL_SOCKET, windows.SO_REUSEADDR, 1)
	}))

	if err != nil {
		log.Fatal(err)
	}

	/* l.setDef

	if err := syscall.SetsockoptInt(fd, syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 1); err != nil {
		log.Fatal(err)
	}
	*/

	l.SetReadBuffer(maxDatagramSize)
	fmt.Println("mcast server listening on", srvAddr)
	for {
		b := make([]byte, maxDatagramSize)
		n, src, err := l.ReadFromUDP(b)
		if err != nil {
			log.Fatal("ReadFromUDP failed:", err)
		}
		h(src, n, b)
	}
}

func findMulticastInterfaces() {
	ifaces, err := net.Interfaces()
	if err != nil {
		fmt.Print(fmt.Errorf("localAddresses: %+v\n", err.Error()))
		return
	}
	for _, i := range ifaces {

		if i.Flags&net.FlagMulticast != 0 {
			if i.Flags&net.FlagUp != 0 {
				log.Println("active mcast Interface: ", i)
				if strings.Contains(i.Name, "WiFi") {

					addrs, err := i.Addrs()
					log.Println("addresses", addrs)
					if err != nil {
						fmt.Print(fmt.Errorf("localAddresses: %+v\n", err.Error()))
					}
					for _, address := range addrs {
						if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
							if ipnet.IP.To4() != nil {
								log.Println("IPv4: ", ipnet.IP.String())
								//serveMulticastUDP(srvAddr, &i, msgHandler)
								serveMcast(&i)
							}
						}

					}

				}
			}
		}

	}
}

func main() {

	//ping(srvAddr)

	findMulticastInterfaces()

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

	for {
		// wait for UDP client to connect
		handleSonoffUDPConnection(ln)
	}

}
