package packet

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
)

func hashSum(file *os.File) (string, error) {

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	sum := fmt.Sprintf("%x", hash.Sum(nil))

	//back file pointer to the start
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return "", err
	}

	return sum, nil
}
