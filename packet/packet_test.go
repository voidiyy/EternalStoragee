package packet

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net"
	"os"
	"path/filepath"
	"testing"
)

// Модульний тест для функції створення TCP-пакету з файлу
func TestNewTCPPacket(t *testing.T) {
	// Використовуємо тимчасовий файл для тестування
	tmpFile, err := os.Create("test.txt")
	require.NoError(t, err)

	// Запис даних у файл
	_, err = tmpFile.WriteString("Sample data for TCP packet")
	require.NoError(t, err)

	// Тест для різних типів компресії
	for _, compressType := range []string{"gzip", "snappy", "zlib"} {
		packet, err := NewTCPPacket(tmpFile.Name(), compressType)
		assert.NoError(t, err)
		assert.NotNil(t, packet)
		assert.Equal(t, filepath.Base(tmpFile.Name()), packet.MetaData.FileName)
	}
}

// Інтеграційний тест для функцій SendOverTCP та ReceiveOverTCP
func TestSendAndReceiveOverTCP(t *testing.T) {
	// Створення тимчасового серверу для тестування
	listener, err := net.Listen("tcp", "localhost:8080")
	require.NoError(t, err)
	defer listener.Close()

	// Використовуємо канал для синхронізації клієнт-сервер взаємодії
	done := make(chan *TCPPacket)

	// Створення сервера
	go func() {
		conn, err := listener.Accept()
		require.NoError(t, err)
		defer conn.Close()

		receivedPacket, err := ReceiveOverTCP(conn)
		require.NoError(t, err)
		done <- receivedPacket
	}()

	// Підключення клієнта та відправка даних
	conn, err := net.Dial("tcp", listener.Addr().String())
	require.NoError(t, err)
	defer conn.Close()

	// Створення TCP пакету
	packet := &TCPPacket{
		MetaData: &TCPPacketMetaData{
			FileName: "testfile.txt",
			FileType: "text",
			FileHash: "dummyhash",
			Size:     int64(len("Hello World")),
		},
		Bytes: []byte("Hello World"),
	}

	err = packet.SendOverTCP(conn)
	require.NoError(t, err)

	// Отримання та перевірка пакету
	receivedPacket := <-done
	assert.Equal(t, packet.MetaData.FileName, receivedPacket.MetaData.FileName)
	assert.Equal(t, packet.Bytes, receivedPacket.Bytes)
}

// Тест функції збереження файлу
func TestSaveFile(t *testing.T) {
	packet := &TCPPacket{
		MetaData: &TCPPacketMetaData{
			FileName: "saved_file.txt",
			FileHash: "expectedhash",
			Size:     int64(len("Sample content")),
		},
		Bytes: []byte("Sample content"),
	}

	// Встановлюємо тимчасовий шлях для збереження
	tempDir := t.TempDir()
	err := packet.SaveFile(tempDir)
	assert.NoError(t, err)

	// Перевірка наявності та відповідності файлу
	savedFilePath := filepath.Join(tempDir, packet.MetaData.FileName)
	info, err := os.Stat(savedFilePath)
	assert.NoError(t, err)
	assert.Equal(t, packet.MetaData.Size, info.Size())
}

// Тест для порівняння хешів файлів
func TestCompareHashSum(t *testing.T) {
	packet := &TCPPacket{
		MetaData: &TCPPacketMetaData{
			FileHash: "hash1",
		},
	}

	// Порівняння хешів
	assert.NoError(t, packet.compareHashSUm("hash1"))
	assert.Error(t, packet.compareHashSUm("hash2"))
}

// Тест для JSON-серіалізації та десеріалізації
func TestToJsonFromJson(t *testing.T) {
	packet := &TCPPacket{
		MetaData: &TCPPacketMetaData{
			FileName: "example.txt",
			FileType: "text",
		},
		Bytes: []byte("This is a test content"),
	}

	// Серіалізація
	jsonData, err := packet.ToJson()
	require.NoError(t, err)

	// Десеріалізація
	newPacket := &TCPPacket{}
	err = json.Unmarshal(jsonData, newPacket)
	require.NoError(t, err)
	assert.Equal(t, packet.MetaData.FileName, newPacket.MetaData.FileName)
	assert.Equal(t, packet.Bytes, newPacket.Bytes)
}
