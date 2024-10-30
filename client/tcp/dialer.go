package tcp

import (
	packet "EternalPacket"
	"bufio"
	"eternalStorageClient/logger"
	"fmt"
	"net"
	"os"
)

type DialerTCP struct {
	RemoteAddr string
	conn       net.Conn
	logger     *logger.EtrnlLogger

	inMsgChan  chan string
	outMsgChan chan string
	errChan    chan error
}

func NewDialerTCP(remoteAddr string) (*DialerTCP, error) {

	conn, err := net.Dial("tcp", remoteAddr)
	if err != nil {
		return nil, err
	}

	return &DialerTCP{
		RemoteAddr: remoteAddr,
		logger:     logger.NewEtrnlLogger(),
		conn:       conn,
		inMsgChan:  make(chan string),
		outMsgChan: make(chan string),
		errChan:    make(chan error),
	}, nil
}

func (d *DialerTCP) SendFile(pack *packet.TCPPacket) error {
	defer d.conn.Close()

	err := pack.SendOverTCP(d.conn)
	if err != nil {
		return err
	}
	fmt.Println("send successfully")
	return nil
}

func (d *DialerTCP) ReceiveFile(path string) (*packet.TCPPacket, error) {
	defer d.conn.Close()
	return packet.ReceiveOverTCP(d.conn, path)
}

func (d *DialerTCP) Dial() error {
	var err error
	d.conn, err = net.Dial("tcp", d.RemoteAddr)
	if err != nil {
		return d.logger.Err(err, "connection error")
	}
	d.logger.Info("connected to: " + d.conn.RemoteAddr().String())

	go d.handleIncoming()
	go d.handleOutgoing()

	for {
		select {
		case msg := <-d.inMsgChan:
			fmt.Printf("$-> %s", string(msg))
		case <-d.outMsgChan:
		case e := <-d.errChan:
			return d.logger.Err(e, "error in communication")
		}
	}
}

func (d *DialerTCP) handleIncoming() {
	for {
		reader, err := bufio.NewReader(d.conn).ReadString('\n')
		if err != nil {
			d.errChan <- err
		}
		d.inMsgChan <- reader
	}
}

func (d *DialerTCP) handleOutgoing() {
	for {
		reader, err := bufio.NewReader(os.Stdin).ReadString('\n')
		if err != nil {
			d.errChan <- err
		}

		_, err = d.conn.Write([]byte(reader + "\n"))
		if err != nil {
			d.errChan <- err
		}

		d.outMsgChan <- reader
	}
}
