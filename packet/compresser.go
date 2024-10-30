package packet

import (
	"bytes"
	"compress/gzip"
	"compress/zlib"
	"fmt"
	"github.com/golang/snappy"
	"io"
	"os"
	"path/filepath"
)

func (tp *TCPPacket) decompressToFile(dstFile string) error {
	var (
		err    error
		reader io.Reader
	)
	outFile, er := os.OpenFile(dstFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, tp.MetaData.FileMode)
	if er != nil {
		return er
	}
	defer outFile.Close()

	switch tp.MetaData.CompressType {
	case "gzip":
		reader, err = gzip.NewReader(bytes.NewReader(tp.Bytes))
		if err != nil {
			return err
		}
	case "zlib":
		reader, err = zlib.NewReader(bytes.NewReader(tp.Bytes))
		if err != nil {
			return err
		}
	case "snappy":
		reader = snappy.NewReader(bytes.NewReader(tp.Bytes))
	}

	n, e := io.Copy(outFile, reader)
	if e != nil {
		return e
	}

	if n != tp.MetaData.Size {
		return fmt.Errorf("file size mismatch: %d vs %d", n, tp.MetaData.Size)
	}

	fmt.Printf("Successfully decompessed to file: %s - %d bytes: ", outFile.Name(), n)
	return nil
}

func NewTCPPacketSNAPPY(path string) (*TCPPacket, error) {
	var (
		file *os.File
		info os.FileInfo
		err  error
		buff bytes.Buffer
	)

	file, err = os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	info, err = file.Stat()
	if err != nil {
		return nil, err
	}

	sum, er := hashSum(file)
	if er != nil {
		return nil, er
	}

	snppy := snappy.NewBufferedWriter(&buff)

	n, e := io.Copy(snppy, file)
	if e != nil {
		return nil, e
	}

	if n != info.Size() {
		return nil, fmt.Errorf("file size mismatch: %d vs %d", n, info.Size())
	}
	if err := snppy.Close(); err != nil {
		return nil, err
	}

	tp := &TCPPacket{
		MetaData: &TCPPacketMetaData{
			FileName:       info.Name(),
			FileType:       filepath.Ext(info.Name()),
			FileHash:       sum,
			FileMode:       info.Mode(),
			Size:           info.Size(),
			CompressedSize: int64(buff.Len()),
			CompressType:   "snappy",
		},
		Bytes: buff.Bytes(),
	}
	tp.print()

	return tp, nil
}

func NewTCPPacketZLIB(path string) (*TCPPacket, error) {
	var (
		file *os.File
		info os.FileInfo
		err  error
		buff bytes.Buffer
	)

	file, err = os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	info, err = file.Stat()
	if err != nil {
		return nil, err
	}

	sum, er := hashSum(file)
	if er != nil {
		return nil, er
	}

	zl := zlib.NewWriter(&buff)

	n, err := io.Copy(zl, file)
	if err != nil {
		return nil, err
	}

	if n != info.Size() {
		return nil, fmt.Errorf("file size mismatch: %d vs %d", n, info.Size())
	}
	if err := zl.Close(); err != nil {
		return nil, err
	}

	tp := &TCPPacket{
		MetaData: &TCPPacketMetaData{
			FileName:       info.Name(),
			FileType:       filepath.Ext(info.Name()),
			FileHash:       sum,
			FileMode:       info.Mode(),
			Size:           info.Size(),
			CompressedSize: int64(buff.Len()),
			CompressType:   "zlib",
		},
		Bytes: buff.Bytes(),
	}
	tp.print()

	return tp, nil
}

func NewTCPPacketGZIP(path string) (*TCPPacket, error) {
	var (
		file *os.File
		info os.FileInfo
		err  error
		buff bytes.Buffer
	)

	file, err = os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	info, err = file.Stat()
	if err != nil {
		return nil, err
	}

	sum, er := hashSum(file)
	if er != nil {
		return nil, er
	}

	gz := gzip.NewWriter(&buff)

	n, e := io.Copy(gz, file)
	if e != nil {
		return nil, e
	}
	if n != info.Size() {
		return nil, fmt.Errorf("file size mismatch %d vs %d", n, info.Size())
	}

	if errr := gz.Close(); errr != nil {
		return nil, errr
	}

	tp := &TCPPacket{
		MetaData: &TCPPacketMetaData{
			FileName:       info.Name(),
			FileType:       filepath.Ext(info.Name()),
			FileHash:       sum,
			FileMode:       info.Mode(),
			Size:           info.Size(),
			CompressedSize: int64(buff.Len()),
			CompressType:   "gzip",
		},
		Bytes: buff.Bytes(),
	}
	tp.print()

	return tp, nil
}
