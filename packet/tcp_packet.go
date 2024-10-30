package packet

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
)

// check if struct == interface
var _ Packet = &TCPPacket{}

type TCPPacketMetaData struct {
	FileName       string      `json:"file_name"`
	FileType       string      `json:"file_type"`
	FileHash       string      `json:"file_hash"`
	FileMode       os.FileMode `json:"file_mode"`
	CompressedSize int64       `json:"compressed_size"`
	Size           int64       `json:"size"`
	CompressType   string      `json:"compress_type"`
}

type TCPPacket struct {
	MetaData *TCPPacketMetaData `json:"meta_data"`
	Bytes    []byte             `json:"bytes"`
}

func NewTCPPacket(path, compressType string) (*TCPPacket, error) {
	if path == "" || compressType == "" {
		return nil, fmt.Errorf("path or compress type is empty")
	}

	switch compressType {
	case "gzip":
		return NewTCPPacketGZIP(path)
	case "snappy":
		return NewTCPPacketSNAPPY(path)
	case "zlib":
		return NewTCPPacketZLIB(path)
	default:
		return NewTCPPacketSNAPPY(path)
	}
}

func (tp *TCPPacket) SendOverTCP(conn net.Conn) error {
	// Серіалізація метаданих
	meta, err := json.Marshal(tp.MetaData)
	if err != nil {
		return fmt.Errorf("error marshaling metadata: %v", err)
	}

	// Логування початку передачі
	fmt.Printf("Starting to send packet: %s, size: %d bytes\n", tp.MetaData.FileName, len(tp.Bytes))

	// Надсилаємо довжину метаданих
	if err := binary.Write(conn, binary.LittleEndian, uint32(len(meta))); err != nil {
		return fmt.Errorf("error writing metadata length: %v", err)
	}
	fmt.Printf("Sent metadata length: %d bytes\n", len(meta))

	// Надсилаємо метадані
	if n, err := conn.Write(meta); err != nil || n != len(meta) {
		return fmt.Errorf("error sending metadata: wrote %d bytes, expected %d bytes, error: %v", n, len(meta), err)
	}
	fmt.Println("Metadata sent successfully.")

	// Ініціалізуємо константу для розміру блоку
	const chunkSize = 32 * 1024 // 32 KB для ефективної передачі великих файлів
	reader := bytes.NewReader(tp.Bytes)

	// Передаємо дані великими блоками
	for {
		chunk := make([]byte, chunkSize)
		bytesRead, err := reader.Read(chunk)
		if bytesRead > 0 {
			// Спершу надсилаємо розмір блоку
			if err := binary.Write(conn, binary.LittleEndian, int32(bytesRead)); err != nil {
				return fmt.Errorf("error writing chunk size: %v", err)
			}
			// Надсилаємо самі дані
			if _, err := conn.Write(chunk[:bytesRead]); err != nil {
				return fmt.Errorf("error sending chunk: %v", err)
			}
			fmt.Printf("Sent chunk: %d bytes\n", bytesRead)
		}
		// Перевірка на закінчення даних
		if err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("error reading chunk: %v", err)
		}
	}
	fmt.Println("Data sent successfully.")
	return nil
}

func ReceiveOverTCP(conn net.Conn, path string) (*TCPPacket, error) {
	var metaLength uint32

	// get meta data length
	if err := binary.Read(conn, binary.LittleEndian, &metaLength); err != nil {
		return nil, fmt.Errorf("error reading metadata length: %v", err)
	}
	fmt.Printf("Metadata length received: %d bytes\n", metaLength)

	// get meta data
	meta := make([]byte, metaLength)
	if n, err := io.ReadFull(conn, meta); err != nil || uint32(n) != metaLength {
		return nil, fmt.Errorf("error reading metadata: read %d bytes, expected %d, error: %v", n, metaLength, err)
	}

	//write meta data to struct
	var metaData *TCPPacketMetaData
	if err := json.Unmarshal(meta, &metaData); err != nil {
		return nil, fmt.Errorf("error unmarshalling metadata: %v", err)
	}
	fmt.Printf("Received metadata: %v\n", metaData)

	// data buffer
	var packetBuffer bytes.Buffer

	for {
		var chunkSize int32
		// get len of chunk
		if err := binary.Read(conn, binary.LittleEndian, &chunkSize); err != nil {
			if err == io.EOF {
				break
			}
			return nil, fmt.Errorf("error reading chunk size: %v", err)
		}

		// get chunk
		chunk := make([]byte, chunkSize)
		if n, err := io.ReadFull(conn, chunk); err != nil || int32(n) != chunkSize {
			return nil, fmt.Errorf("error reading chunk: read %d bytes, expected %d, error: %v", n, chunkSize, err)
		}
		packetBuffer.Write(chunk)
		fmt.Printf("Received chunk: %d bytes\n", chunkSize)
	}

	fmt.Println("Data received successfully.")
	tp := &TCPPacket{
		MetaData: metaData,
		Bytes:    packetBuffer.Bytes(),
	}
	tp.print()

	err := tp.decompressToFile(path)
	if err != nil {
		return nil, fmt.Errorf("error decompressing file: %v", err)
	}

	return tp, nil
}

func (tp *TCPPacket) SaveFile(path string) error {
	//create file
	if path == "" {
		return fmt.Errorf("file path is empty")
	}
	filePath := filepath.Join(path, tp.MetaData.FileName)
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	//buffer for file
	var buff = bytes.NewBuffer(make([]byte, tp.MetaData.Size))
	b, e := io.Copy(file, buff)
	if e != nil {
		return e
	}

	//compare bytes
	if b != tp.MetaData.Size {
		return fmt.Errorf("file size mismatch: %d vs %d", b, tp.MetaData.Size)
	}

	//get check sum
	sum, er := hashSum(file)
	if er != nil {
		return er
	}

	// compare check sums
	err = tp.compareHashSUm(sum)
	if err != nil {
		return err
	}

	return nil
}

func (tp *TCPPacket) compareHashSUm(newHash string) error {
	if tp.MetaData.FileHash == newHash {
		return nil
	}
	return fmt.Errorf("file hash mismatch: %s vs %s", tp.MetaData.FileHash, newHash)
}

func (tp *TCPPacket) ToJson() ([]byte, error) {
	return json.Marshal(tp)
}

func (tp *TCPPacket) FromJson() (*TCPPacket, error) {
	var packet = &TCPPacket{}

	err := json.Unmarshal(tp.Bytes, packet)
	if err != nil {
		return nil, err
	}
	return packet, nil
}

func (tp *TCPPacket) print() {
	fmt.Println("TCPPacket: ")
	fmt.Printf("FileName: %s\n", tp.MetaData.FileName)
	fmt.Printf("FileType: %s\n", tp.MetaData.FileType)
	fmt.Printf("Compress type: %s\n", tp.MetaData.CompressType)
	fmt.Printf("Size: %d\n", tp.MetaData.Size)
	fmt.Printf("Compress size: %d\n", tp.MetaData.CompressedSize)
	fmt.Printf("HashSum: %s\n", tp.MetaData.FileHash)
	fmt.Printf("FileMode: %s\n", tp.MetaData.FileMode.String())
}
