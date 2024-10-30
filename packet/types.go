package packet

import (
	"net"
)

type Packet interface {
	ToJson() ([]byte, error)
	FromJson() (*TCPPacket, error)
	SaveFile(path string) error
	SendOverTCP(conn net.Conn) error
}

func PacketInit(path, compType string) Packet {
	t, err := NewTCPPacket(path, compType)
	if err != nil {
		panic(err)
	}
	return t
}
