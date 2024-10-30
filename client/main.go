package main

import (
	packet "EternalPacket"
	"flag"
	"fmt"
	"log"
	"net"
)

// dial conn
// go run main.go -mode dial -addr localhost:8080

func main() {

	addr := flag.String("addr", "localhost:8080", "service address")

	conn, err := net.Dial("tcp", *addr)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Client connected")

	err = packet.ReceiveCert(conn)
	if err != nil {
		log.Fatal(err)
	}

	/*mode := flag.String("mode", "dial", "client|server mode, conn/dial")
	path := flag.String("path", "", "file path")
	compType := flag.String("compType", "gzip", "gzip/zlib/snappy")
	flag.Parse()

	if *mode != "dial" && *mode != "send" {
		fmt.Println("mode: dial/conn")
		os.Exit(1)
	}

	dialer, e := tcp.NewDialerTCP(*addr)
	if e != nil {
		fmt.Println(e)
	}

	switch *mode {
	case "dial":
		err := dialer.Dial()
		if err != nil {
			log.Fatal(err)
		}
	case "send":
		pack, err := packet.NewTCPPacket(*path, *compType)
		if err != nil {
			fmt.Println(err)
		}

		e = dialer.SendFile(pack)
		if e != nil {
			fmt.Println(e)
		}
	}*/
}
